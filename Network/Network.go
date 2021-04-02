package Network

import (
	"errors"
	"io"
	"net"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	MsgEncoding_Binary   = 0x1  // raw binary
	MsgEncoding_Packet   = 0x2  // internal Packet binary
	MsgEncoding_XML      = 0x4  // XML
	MsgEncoding_JSON     = 0x8  // JSON
	MsgEncoding_TextHtml = 0x10 // Text/Html
	MsgEncodingMax       = MsgEncoding_Binary | MsgEncoding_Packet | MsgEncoding_XML | MsgEncoding_JSON | MsgEncoding_TextHtml
)

type ServiceInfo struct {
	///  服务名字</summary>
	serviceKey string

	///  协议、地址及端口
	url string

	///  是否是服务器
	bIsServer bool

	///  接入的合法地址范围(只对服务器有效)
	clientIPAddress string
}

var errClosing = errors.New("closing")
var errCloseConns = errors.New("close conns")

// Action is an action that occurs after the completion of an event.
type Action int

const (
	// None indicates that no action should occur following an event.
	None Action = iota
	// Detach detaches a connection. Not available for UDP connections.
	Detach
	// Close closes the connection.
	Close
	// Shutdown shutdowns the server.
	Shutdown
)

type stdloop struct {
	idx   int                  // loop index
	ch    chan interface{}     // command channel
	conns map[*tcpSession]bool // track all the conns bound to this loop
}

type NetworkModule struct {
	svcInfos  []ServiceInfo
	evManager IEventHandlerManager
	loops     []*stdloop     // all the loops
	lns       []*listener    // all the listeners
	loopwg    sync.WaitGroup // loop close waitgroup
	lnwg      sync.WaitGroup // listener close waitgroup
	cond      *sync.Cond     // shutdown signaler
	serr      error          // signal error
	accepted  uintptr        // accept counter
}

func (s *NetworkModule) Run(evMngr IEventHandlerManager, numLoops int) error {

	defer func() {
		for _, ln := range s.lns {
			ln.close()
		}
	}()

	s.evManager = evMngr
	s.cond = sync.NewCond(&sync.Mutex{})

	if numLoops <= 0 {
		numLoops = runtime.NumCPU()
	}

	for i := 0; i < numLoops; i++ {
		s.loops = append(s.loops, &stdloop{
			idx:   i,
			ch:    make(chan interface{}),
			conns: make(map[*tcpSession]bool),
		})
	}

	var ferr error
	defer func() {
		// wait on a signal for shutdown
		ferr = s.waitForShutdown()

		// notify all loops to close by closing all listeners
		for _, l := range s.loops {
			l.ch <- errClosing
		}

		// wait on all loops to main loop channel events
		s.loopwg.Wait()

		// shutdown all listeners
		for i := 0; i < len(s.lns); i++ {
			s.lns[i].close()
		}

		// wait on all listeners to complete
		s.lnwg.Wait()

		// close all connections
		s.loopwg.Add(len(s.loops))
		for _, l := range s.loops {
			l.ch <- errCloseConns
		}
		s.loopwg.Wait()

	}()

	s.loopwg.Add(numLoops)
	for i := 0; i < numLoops; i++ {
		go stdloopRun(s, s.loops[i])
	}

	s.lnwg.Add(len(s.lns))
	for i := 0; i < len(s.lns); i++ {
		go stdlistenerRun(s, s.lns[i], i)
	}

	return ferr
}

func (s *NetworkModule) GetServiceInfo(key string) *ServiceInfo {
	for _, info := range s.svcInfos {
		if info.serviceKey == key {
			return &info
		}
	}

	return nil
}

// Serve starts handling events for the specified addresses.
//
// Addresses should use a scheme prefix and be formatted
// like `tcp://192.168.0.10:9851` or `unix://socket`.
// Valid network schemes:
//  tcp   - bind to both IPv4 and IPv6
//  tcp4  - IPv4
//  tcp6  - IPv6
//  udp   - bind to both IPv4 and IPv6
//  udp4  - IPv4
//  udp6  - IPv6
//  unix  - Unix Domain Socket
// The "tcp" network scheme is assumed when one is not specified.
func (s *NetworkModule) AddService(svcInfo ServiceInfo) error {

	for _, info := range s.svcInfos {
		if info.serviceKey == svcInfo.serviceKey {
			return errors.New("Service with same name already exists")
		}

		s.svcInfos = append(s.svcInfos, svcInfo)
	}

	var stdlib bool
	var ln listener
	ln.network, ln.addr, ln.opts, stdlib = parseAddr(svcInfo.url)
	if ln.network == "unix" {
		os.RemoveAll(ln.addr)
	}

	var err error
	if ln.network == "udp" {
		if ln.opts.reusePort {
			ln.pconn, err = reuseportListenPacket(ln.network, ln.addr)
		} else {
			ln.pconn, err = net.ListenPacket(ln.network, ln.addr)
		}
	} else {
		if ln.opts.reusePort {
			ln.ln, err = reuseportListen(ln.network, ln.addr)
		} else {
			ln.ln, err = net.Listen(ln.network, ln.addr)
		}
	}

	if err != nil {
		return err
	}

	if ln.pconn != nil {
		ln.lnaddr = ln.pconn.LocalAddr()
	} else {
		ln.lnaddr = ln.ln.Addr()
	}

	if !stdlib {
		if err := ln.system(); err != nil {
			return err
		}
	}

	s.lns = append(s.lns, &ln)

	return nil
}

func (s *NetworkModule) Connect(svcKey string, timeOut int) {

}

// waitForShutdown waits for a signal to shutdown
func (s *NetworkModule) waitForShutdown() error {
	s.cond.L.Lock()
	s.cond.Wait()
	err := s.serr
	s.cond.L.Unlock()
	return err
}

// signalShutdown signals a shutdown an begins server closing
func (s *NetworkModule) signalShutdown(err error) {
	s.cond.L.Lock()
	s.serr = err
	s.cond.Signal()
	s.cond.L.Unlock()
}

func stdlistenerRun(s *NetworkModule, ln *listener, lnidx int) {
	var ferr error
	defer func() {
		s.signalShutdown(ferr)
		s.lnwg.Done()
	}()

	var packet [0xFFFF]byte
	for {
		if ln.pconn != nil {
			// udp
			n, addr, err := ln.pconn.ReadFrom(packet[:])
			if err != nil {
				ferr = err
				return
			}

			l := s.loops[int(atomic.AddUintptr(&s.accepted, 1))%len(s.loops)]
			l.ch <- &udpSession{
				addrIndex:  lnidx,
				localAddr:  ln.lnaddr,
				remoteAddr: addr,
				in:         append([]byte{}, packet[:n]...),
			}

		} else {
			// tcp
			conn, err := ln.ln.Accept()
			if err != nil {
				ferr = err
				return
			}
			l := s.loops[int(atomic.AddUintptr(&s.accepted, 1))%len(s.loops)]
			c := &tcpSession{conn: conn, loop: l, lnidx: lnidx}
			l.ch <- c
			go func(c *tcpSession) {
				var packet [0xFFFF]byte
				for {
					n, err := c.conn.Read(packet[:])
					if err != nil {
						c.conn.SetReadDeadline(time.Time{})
						l.ch <- &stderr{c, err}
						return
					}
					l.ch <- &stdin{c, append([]byte{}, packet[:n]...)}
				}
			}(c)
		}
	}
}

func stdloopRun(s *NetworkModule, l *stdloop) {
	var err error
	tick := make(chan bool)
	tock := make(chan time.Duration)
	defer func() {
		//fmt.Println("-- loop stopped --", l.idx)
		if l.idx == 0 {
			close(tock)
			go func() {
				for range tick {
				}
			}()
		}
		s.signalShutdown(err)
		s.loopwg.Done()
		stdloopEgress(s, l)
		s.loopwg.Done()
	}()

	if l.idx == 0 {
		go func() {
			for {
				tick <- true
				delay, ok := <-tock
				if !ok {
					break
				}
				time.Sleep(delay)
			}
		}()
	}

	//fmt.Println("-- loop started --", l.idx)
	for {
		select {
		case <-tick:
			delay, action := s.evManager.Tick()
			switch action {
			case Shutdown:
				err = errClosing
			}
			tock <- delay
		case v := <-l.ch:
			switch v := v.(type) {
			case error:
				err = v
			case *tcpSession:
				err = stdloopAccept(s, l, v)
			case *stdin:
				err = stdloopRead(s, l, v.c, v.in)
			case *udpSession:
				err = stdloopReadUDP(s, l, v)
			case *stderr:
				err = stdloopError(s, l, v.c, v.err)
			case wakeReq:
				err = stdloopRead(s, l, v.c, nil)
			}
		}
		if err != nil {
			return
		}
	}
}

func stdloopEgress(s *NetworkModule, l *stdloop) {
	var closed bool
loop:
	for v := range l.ch {
		switch v := v.(type) {
		case error:
			if v == errCloseConns {
				closed = true
				for c := range l.conns {
					stdloopClose(s, l, c)
				}
			}
		case *stderr:
			stdloopError(s, l, v.c, v.err)
		}
		if len(l.conns) == 0 && closed {
			break loop
		}
	}
}

func stdloopError(s *NetworkModule, l *stdloop, c *tcpSession, err error) error {
	delete(l.conns, c)
	closeEvent := true
	switch atomic.LoadInt32(&c.done) {
	case 0: // read error
		c.conn.Close()
		if err == io.EOF {
			err = nil
		}
	case 1: // closed
		c.conn.Close()
		err = nil
	case 2: // detached
		err = nil
		if s.events.Detached == nil {
			c.conn.Close()
		} else {
			closeEvent = false
			switch s.events.Detached(c, &stddetachedConn{c.conn, c.donein}) {
			case Shutdown:
				return errClosing
			}
		}
	}
	if closeEvent {
		if s.events.Closed != nil {
			switch s.events.Closed(c, err) {
			case Shutdown:
				return errClosing
			}
		}
	}
	return nil
}

type stddetachedConn struct {
	conn net.Conn // original conn
	in   []byte   // extra input data
}

func (c *stddetachedConn) Read(p []byte) (n int, err error) {
	if len(c.in) > 0 {
		if len(c.in) <= len(p) {
			copy(p, c.in)
			n = len(c.in)
			c.in = nil
			return
		}
		copy(p, c.in[:len(p)])
		n = len(p)
		c.in = c.in[n:]
		return
	}
	return c.conn.Read(p)
}

func (c *stddetachedConn) Write(p []byte) (n int, err error) {
	return c.conn.Write(p)
}

func (c *stddetachedConn) Close() error {
	return c.conn.Close()
}

func (c *stddetachedConn) Wake() {}

func stdloopRead(s *NetworkModule, l *stdloop, c *tcpSession, in []byte) error {
	if atomic.LoadInt32(&c.done) == 2 {
		// should not ignore reads for detached connections
		c.donein = append(c.donein, in...)
		return nil
	}
	if s.events.Data != nil {
		out, action := s.events.Data(c, in)
		if len(out) > 0 {
			if s.events.PreWrite != nil {
				s.events.PreWrite()
			}
			c.conn.Write(out)
		}
		switch action {
		case Shutdown:
			return errClosing
		case Detach:
			return stdloopDetach(s, l, c)
		case Close:
			return stdloopClose(s, l, c)
		}
	}
	return nil
}

func stdloopReadUDP(s *NetworkModule, l *stdloop, c *udpSession) error {
	if s.events.Data != nil {
		out, action := s.events.Data(c, c.in)
		if len(out) > 0 {
			if s.events.PreWrite != nil {
				s.events.PreWrite()
			}
			s.lns[c.addrIndex].pconn.WriteTo(out, c.remoteAddr)
		}
		switch action {
		case Shutdown:
			return errClosing
		}
	}
	return nil
}

func stdloopDetach(s *NetworkModule, l *stdloop, c *tcpSession) error {
	atomic.StoreInt32(&c.done, 2)
	c.conn.SetReadDeadline(time.Now())
	return nil
}

func stdloopClose(s *NetworkModule, l *stdloop, c *tcpSession) error {
	atomic.StoreInt32(&c.done, 1)
	c.conn.SetReadDeadline(time.Now())
	return nil
}

func stdloopAccept(s *NetworkModule, l *stdloop, c *tcpSession) error {
	l.conns[c] = true
	c.addrIndex = c.lnidx
	c.localAddr = s.lns[c.lnidx].lnaddr
	c.remoteAddr = c.conn.RemoteAddr()

	if s.events.Opened != nil {
		out, opts, action := s.events.Opened(c)
		if len(out) > 0 {
			if s.events.PreWrite != nil {
				s.events.PreWrite()
			}
			c.conn.Write(out)
		}
		if opts.TCPKeepAlive > 0 {
			if c, ok := c.conn.(*net.TCPConn); ok {
				c.SetKeepAlive(true)
				c.SetKeepAlivePeriod(opts.TCPKeepAlive)
			}
		}
		switch action {
		case Shutdown:
			return errClosing
		case Detach:
			return stdloopDetach(s, l, c)
		case Close:
			return stdloopClose(s, l, c)
		}
	}
	return nil
}
