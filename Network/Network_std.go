package Network

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type stdloop struct {
	idx   int                  // loop index
	ch    chan interface{}     // command channel
	conns map[*tcpSession]bool // track all the conns bound to this loop
}

type NetworkModuleStd struct {
	NetworkModuleBase
	loops          []*stdloop       // all the loops
	lns            []*listener      // all the listeners
	connects       []*connector     // all the connectors
	clientSessions []*clientSession // all the clients
	clientMutex    sync.Mutex
	loopwg         sync.WaitGroup // loop close waitgroup
	lnwg           sync.WaitGroup // listener close waitgroup
	connectwg      sync.WaitGroup // connector close waitgroup
	cond           *sync.Cond     // shutdown signaler
	serr           error          // signal error
	accepted       uintptr        // accept counter
	status         int32          //0: init; 1: running; 2:shutting down; 3:shutdown
}

func (m *NetworkModuleStd) Run(evMngr IEventHandlerManager, numLoops int) error {

	m.evManager = evMngr
	m.cond = sync.NewCond(&sync.Mutex{})

	if numLoops <= 0 {
		numLoops = runtime.NumCPU()
	}

	for i := 0; i < numLoops; i++ {
		m.loops = append(m.loops, &stdloop{
			idx:   i,
			ch:    make(chan interface{}),
			conns: make(map[*tcpSession]bool),
		})
	}

	var ferr error
	defer func() {
		// wait on a signal for shutdown
		ferr = m.waitForShutdown()

		atomic.StoreInt32(&m.status, 2)

		// notify all loops to close by closing all listeners
		for _, l := range m.loops {
			l.ch <- errShutdown
		}
		m.loopwg.Wait()

		// shutdown all listeners
		for i := 0; i < len(m.lns); i++ {
			m.lns[i].close()
		}
		m.lnwg.Wait()

		// close all connections
		m.loopwg.Add(len(m.loops))
		for _, l := range m.loops {
			l.ch <- errCloseConns
		}
		m.loopwg.Wait()

		// close all connectors
		m.clientMutex.Lock()
		for i := 0; i < len(m.clientSessions); i++ {
			m.clientSessions[i].conn.Close()
		}
		m.clientMutex.Unlock()
		m.connectwg.Wait()

		atomic.StoreInt32(&m.status, 3)
	}()

	m.loopwg.Add(numLoops)
	for i := 0; i < numLoops; i++ {
		go stdLoopRun(m, m.loops[i])
	}

	m.lnwg.Add(len(m.lns))
	for i := 0; i < len(m.lns); i++ {
		go stdListenerRun(m, m.lns[i], i)
	}

	for i := 0; i < len(m.connects); i++ {
		connecting(m, m.connects[i])
	}

	atomic.StoreInt32(&m.status, 1)

	return ferr
}

func (m *NetworkModuleStd) Shutdown() error {
	m.signalShutdown(errShutdown)
	return nil
}

func (m *NetworkModuleStd) Listen(svcKey string, url string) error {

	network, addr, opts := parseAddr(url)

	svcInfo := &ServerInfo{
		Key:       svcKey,
		Network:   network,
		Address:   addr,
		IsServer:  false,
		ReusePort: opts.reusePort}

	err := m.AddServerInfo(svcInfo)
	if err != nil {
		return err
	}

	return m.ListenSvc(svcKey)
}

func (m *NetworkModuleStd) ListenSvc(svcKey string) error {

	svcInfo := m.GetServerInfo(svcKey)
	if svcInfo == nil {
		return errors.New("service not exist")
	}

	var ln listener
	ln.svcKey = svcKey
	ln.network = svcInfo.Network
	ln.addr = svcInfo.Address
	ln.opts.reusePort = svcInfo.ReusePort

	if ln.network == "unix" {
		os.RemoveAll(ln.addr)
	}

	var err error
	if ln.network == "udp" {
		ln.udpSessions = make(map[net.Addr]*udpSession)
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

	if atomic.LoadInt32(&m.status) == 1 {
		l := m.loops[0]
		ln := &newListener{ln: &ln}
		l.ch <- ln
	} else {
		m.lns = append(m.lns, &ln)
	}

	return nil
}

func connecting(m *NetworkModuleStd, c *connector) {

	m.connectwg.Add(1)

	go func() {
		defer m.connectwg.Done()

		conn, err := net.DialTimeout(c.network, c.addr, c.timeOut)
		if err != nil {
			m.evManager.OnConnectFailed(c.svcKey)
			return
		}

		session := &clientSession{
			conn:      conn,
			svcKey:    c.svcKey,
			sessionID: atomic.AddUint64(&allSessionID, 1),
		}
		session.eventHandler = m.evManager.CreateEventHandler(session)
		opts, _ := session.eventHandler.OnOpened()
		if opts.TCPKeepAlive > 0 {
			if conn, ok := session.conn.(*net.TCPConn); ok {
				conn.SetKeepAlive(true)
				conn.SetKeepAlivePeriod(opts.TCPKeepAlive)
			}
		}

		m.clientMutex.Lock()
		m.clientSessions = append(m.clientSessions, session)
		m.clientMutex.Unlock()

		var packet [0xFFFF]byte
		for {
			n, err := conn.Read(packet[:])
			if err != nil {
				conn.SetReadDeadline(time.Time{})
				session.eventHandler.OnClosed(err)
				return
			}

			session.eventHandler.OnRecvMsg(packet[:n])
		}
	}()
}

func (m *NetworkModuleStd) Connect(svcKey, url string, timeOut time.Duration) error {

	network, addr, opts := parseAddr(url)

	svcInfo := &ServerInfo{
		Key:       svcKey,
		Network:   network,
		Address:   addr,
		IsServer:  false,
		ReusePort: opts.reusePort}

	err := m.AddServerInfo(svcInfo)
	if err != nil {
		return err
	}

	return m.ConnectSvc(svcKey, timeOut)
}

func (m *NetworkModuleStd) ConnectSvc(svcKey string, timeOut time.Duration) error {

	svcInfo := m.GetServerInfo(svcKey)
	if svcInfo == nil {
		return errors.New("service not exist")
	}

	var c connector
	c.svcKey = svcInfo.Key
	c.timeOut = timeOut
	c.network = svcInfo.Network
	c.addr = svcInfo.Address

	if atomic.LoadInt32(&m.status) == 1 {
		connecting(m, &c)
	} else {
		m.connects = append(m.connects, &c)
	}

	return nil
}

// waitForShutdown waits for a signal to shutdown
func (m *NetworkModuleStd) waitForShutdown() error {
	m.cond.L.Lock()

	m.cond.Wait()
	err := m.serr

	m.cond.L.Unlock()
	return err
}

// signalShutdown signals a shutdown an begins server closing
func (m *NetworkModuleStd) signalShutdown(err error) {
	m.cond.L.Lock()

	m.serr = err
	m.cond.Signal()

	m.cond.L.Unlock()
}

func stdListenerRun(m *NetworkModuleStd, ln *listener, lnidx int) {
	var ferr error

	defer func() {

		if ferr != nil {
			m.signalShutdown(ferr)
		}

		m.lnwg.Done()
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

			ip := addr.String()
			if strings.Contains(ip, ":") {
				ip = strings.Split(ip, ":")[0]
			}
			if !m.IsClientIPInRange(ln.svcKey, ip) {
				fmt.Println("client ip is not in range:", ln.svcKey, ip)
				continue
			}

			l := m.loops[int(atomic.AddUintptr(&m.accepted, 1))%len(m.loops)]

			s, ok := ln.udpSessions[addr]
			if !ok {
				s = &udpSession{
					pconn:      ln.pconn,
					svcKey:     ln.svcKey,
					sessionID:  atomic.AddUint64(&allSessionID, 1),
					lnidx:      lnidx,
					remoteAddr: addr,
					in:         append([]byte{}, packet[:n]...),
				}
				s.eventHandler = m.evManager.CreateEventHandler(s)
			} else {
				s.in = append([]byte{}, packet[:n]...)
			}
			l.ch <- s

		} else {
			// tcp
			conn, err := ln.ln.Accept()
			if err != nil {
				ferr = err
				return
			}

			ip := conn.RemoteAddr().String()
			if strings.Contains(ip, ":") {
				ip = strings.Split(ip, ":")[0]
			}
			if !m.IsClientIPInRange(ln.svcKey, ip) {
				fmt.Println("client ip is not in range:", ln.svcKey, ip)
				conn.Close()
				continue
			}

			l := m.loops[int(atomic.AddUintptr(&m.accepted, 1))%len(m.loops)]
			s := &tcpSession{
				svcKey:    ln.svcKey,
				sessionID: atomic.AddUint64(&allSessionID, 1),
				conn:      conn,
				loop:      l,
				lnidx:     lnidx,
			}
			s.eventHandler = m.evManager.CreateEventHandler(s)
			l.ch <- s

			go func(session *tcpSession) {
				var packet [0xFFFF]byte
				for {
					n, err := session.conn.Read(packet[:])
					if err != nil {
						session.conn.SetReadDeadline(time.Time{})
						l.ch <- &stderr{session, err}
						return
					}

					l.ch <- &stdin{session, append([]byte{}, packet[:n]...)}
				}
			}(s)
		}
	}
}

func stdLoopRun(m *NetworkModuleStd, l *stdloop) {
	var err error

	defer func() {
		//fmt.Println("-- loop stopped --", l.idx)

		m.loopwg.Done()
		stdloopEgress(m, l)
		m.loopwg.Done()
	}()

	//fmt.Println("-- loop started --", l.idx)
	for {
		select {
		default:
			time.Sleep(time.Millisecond)
		case v := <-l.ch:
			switch v := v.(type) {
			case error:
				err = v

			case *tcpSession:
				err = stdloopAccept(m, l, v)

			case *stdin:
				err = stdloopRead(m, l, v.c, v.in)

			case *udpSession:
				err = stdloopReadUDP(m, l, v)

			case *stderr:
				err = stdloopError(m, l, v.c, v.err)

			case wakeReq:
				err = stdloopRead(m, l, v.c, nil)

			case *newListener:
				err = stdloopNewListener(m, l, v.ln)
			}
		}

		if err != nil {
			return
		}
	}
}

func stdloopEgress(m *NetworkModuleStd, l *stdloop) {
	var closed bool

loop:
	for v := range l.ch {

		switch v := v.(type) {
		case error:
			if v == errCloseConns {
				closed = true
				for c := range l.conns {
					stdloopClose(m, l, c)
				}
			}

		case *stderr:
			stdloopError(m, l, v.c, v.err)
		}

		if len(l.conns) == 0 && closed {
			break loop
		}
	}
}

func stdloopError(m *NetworkModuleStd, l *stdloop, session *tcpSession, err error) error {
	delete(l.conns, session)
	closeEvent := true

	switch atomic.LoadInt32(&session.done) {
	case 0: // read error
		session.conn.Close()
		if err == io.EOF {
			err = nil
		}

	case 1: // closed
		session.conn.Close()
		err = nil

	case 2: // detached
		err = nil
		closeEvent = false
		session.eventHandler.OnDetached(&stddetachedConn{session.conn, session.donein})
	}

	if closeEvent {
		session.eventHandler.OnClosed(err)
	}

	return nil
}

func stdloopRead(m *NetworkModuleStd, l *stdloop, session *tcpSession, in []byte) error {
	if atomic.LoadInt32(&session.done) == 2 {
		// should not ignore reads for detached connections
		session.donein = append(session.donein, in...)
		return nil
	}

	action := session.eventHandler.OnRecvMsg(in)

	switch action {
	case Detach:
		return stdloopDetach(m, l, session)
	case Close:
		return stdloopClose(m, l, session)
	}

	return nil
}

func stdloopReadUDP(m *NetworkModuleStd, l *stdloop, session *udpSession) error {
	session.eventHandler.OnRecvMsg(session.in)

	return nil
}

func stdloopDetach(m *NetworkModuleStd, l *stdloop, session *tcpSession) error {
	atomic.StoreInt32(&session.done, 2)
	session.conn.SetReadDeadline(time.Now())
	return nil
}

func stdloopClose(m *NetworkModuleStd, l *stdloop, session *tcpSession) error {
	atomic.StoreInt32(&session.done, 1)
	session.conn.SetReadDeadline(time.Now())
	return nil
}

func stdloopAccept(m *NetworkModuleStd, l *stdloop, session *tcpSession) error {
	l.conns[session] = true

	opts, action := session.eventHandler.OnOpened()
	if opts.TCPKeepAlive > 0 {
		if conn, ok := session.conn.(*net.TCPConn); ok {
			conn.SetKeepAlive(true)
			conn.SetKeepAlivePeriod(opts.TCPKeepAlive)
		}
	}

	switch action {
	case Detach:
		return stdloopDetach(m, l, session)
	case Close:
		return stdloopClose(m, l, session)
	}

	return nil
}

func stdloopNewListener(m *NetworkModuleStd, l *stdloop, ln *listener) error {

	idx := len(m.lns)
	m.lns = append(m.lns, ln)

	m.lnwg.Add(1)
	go stdListenerRun(m, ln, idx)

	return nil
}

func NewNetworkModule() INetworkModule {
	var module NetworkModuleStd
	module.severInfoes = make(map[string]*ServerInfo)
	return &module
}
