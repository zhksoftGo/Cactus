package Network

import (
	"net"
)

type INetworkSession interface {
	GetServiceKey() string

	/// 发送消息
	SendMsg(b []byte) error

	/// 关闭会话.
	ShutDown(notify bool)

	/// 获取对方IP地址和端口号.
	GetRemoteAddr() net.Addr

	/// 获取本地IP地址和端口号.
	GetLocalAddr() net.Addr

	Wake()
}

type tcpSession struct {
	svcKey       string
	eventHandler IEventHandler
	conn         net.Conn // original connection
	loop         *stdloop // owner loop
	lnidx        int      // index of listener
	donein       []byte   // extra data for done connection
	done         int32    // 0: attached, 1: closed, 2: detached
}

type wakeReq struct {
	c *tcpSession
}

func (s *tcpSession) GetServiceKey() string { return s.svcKey }
func (s *tcpSession) SendMsg(b []byte) error {
	_, err := s.conn.Write(b)
	return err
}
func (s *tcpSession) ShutDown(notify bool)    { s.conn.Close() }
func (s *tcpSession) GetRemoteAddr() net.Addr { return s.conn.LocalAddr() }
func (s *tcpSession) GetLocalAddr() net.Addr  { return s.conn.RemoteAddr() }
func (s *tcpSession) Wake()                   { s.loop.ch <- wakeReq{s} }

type stdin struct {
	c  *tcpSession
	in []byte
}

type stderr struct {
	c   *tcpSession
	err error
}

type udpSession struct {
	svcKey       string
	eventHandler IEventHandler
	pconn        net.PacketConn
	remoteAddr   net.Addr
	lnidx        int // index of listener
	in           []byte
}

func (s *udpSession) GetServiceKey() string { return s.svcKey }
func (s *udpSession) SendMsg(b []byte) error {
	_, err := s.pconn.WriteTo(b, s.remoteAddr)
	return err
}
func (s *udpSession) ShutDown(notify bool)    {}
func (s *udpSession) GetRemoteAddr() net.Addr { return s.pconn.LocalAddr() }
func (s *udpSession) GetLocalAddr() net.Addr  { return s.remoteAddr }
func (s *udpSession) Wake()                   {}

type clientSession struct {
	svcKey       string
	eventHandler IEventHandler
	conn         net.Conn
}

func (s *clientSession) GetServiceKey() string { return s.svcKey }
func (s *clientSession) SendMsg(b []byte) error {
	_, err := s.conn.Write(b)
	return err
}
func (s *clientSession) ShutDown(notify bool)    {}
func (s *clientSession) GetRemoteAddr() net.Addr { return s.conn.LocalAddr() }
func (s *clientSession) GetLocalAddr() net.Addr  { return s.conn.RemoteAddr() }
func (s *clientSession) Wake()                   {}

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
