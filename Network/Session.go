package Network

import (
	"net"

	"github.com/zhksoftGo/Packet"
)

type INetworkSession interface {
	GetServiceKey() string

	/// 发送Packet二进制格式消息(里面实际内容也可以是XML或JSON字符串), 原始数据含6字节网络头.
	SendPacket(pak *Packet.Packet) error

	/// 关闭会话.
	ShutDown(notify bool)

	/// 获取对方IP地址和端口号.
	GetRemoteAddress() net.Addr

	/// 获取本地IP地址和端口号.
	GetLocalAddress() net.Addr

	Wake()
}

type tcpSession struct {
	svcKey       string
	eventHandler IEventHandler
	localAddr    net.Addr
	remoteAddr   net.Addr
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
func (s *tcpSession) SendPacket(pak *Packet.Packet) error {
	_, err := s.conn.Write(pak.GetUsedBuffer())
	return err
}
func (s *tcpSession) ShutDown(notify bool)       { s.conn.Close() }
func (s *tcpSession) GetRemoteAddress() net.Addr { return s.localAddr }
func (s *tcpSession) GetLocalAddress() net.Addr  { return s.remoteAddr }
func (s *tcpSession) Wake()                      { s.loop.ch <- wakeReq{s} }

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
	lnidx        int // index of listener
	localAddr    net.Addr
	remoteAddr   net.Addr
	in           []byte
}

func (s *udpSession) GetServiceKey() string { return s.svcKey }
func (s *udpSession) SendPacket(pak *Packet.Packet) error {
	_, err := s.pconn.WriteTo(pak.GetUsedBuffer(), s.remoteAddr)
	return err
}
func (s *udpSession) ShutDown(notify bool)       {}
func (s *udpSession) GetRemoteAddress() net.Addr { return s.localAddr }
func (s *udpSession) GetLocalAddress() net.Addr  { return s.remoteAddr }
func (s *udpSession) Wake()                      {}

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
