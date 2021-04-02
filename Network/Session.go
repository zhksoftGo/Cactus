package Network

import (
	"net"

	"github.com/zhksoftGo/Packet"
)

type INetworkSession interface {
	///  Gets the ServiceInfoBase.
	GetServiceInfo() *ServiceInfo

	/// 发送Packet二进制格式消息(里面实际内容也可以是XML或JSON字符串), 原始数据含6字节网络头.
	SendPacket(pak *Packet.Packet) bool

	/// 关闭会话.
	ShutDown(notify bool)

	/// 获取对方IP地址和端口号.
	GetRemoteAddress() net.Addr

	/// 获取本地IP地址和端口号.
	GetLocalAddress() net.Addr

	Wake()
}

type tcpSession struct {
	svcInfo      *ServiceInfo
	addrIndex    int
	localAddr    net.Addr
	remoteAddr   net.Addr
	conn         net.Conn // original connection
	eventHandler *ILinkEventHandler
	loop         *stdloop // owner loop
	lnidx        int      // index of listener
	donein       []byte   // extra data for done connection
	done         int32    // 0: attached, 1: closed, 2: detached
}

type wakeReq struct {
	c *tcpSession
}

func (s *tcpSession) GetServiceInfo() *ServiceInfo       { return s.svcInfo }
func (s *tcpSession) SendPacket(pak *Packet.Packet) bool { return true }
func (s *tcpSession) ShutDown(notify bool)               { s.conn.Close() }
func (s *tcpSession) GetRemoteAddress() net.Addr         { return s.localAddr }
func (s *tcpSession) GetLocalAddress() net.Addr          { return s.remoteAddr }
func (s *tcpSession) Wake()                              { s.loop.ch <- wakeReq{s} }

type stdin struct {
	c  *tcpSession
	in []byte
}

type stderr struct {
	c   *tcpSession
	err error
}

type udpSession struct {
	svcInfo      *ServiceInfo
	addrIndex    int
	localAddr    net.Addr
	remoteAddr   net.Addr
	in           []byte
	eventHandler *ILinkEventHandler
}

func (s *udpSession) GetServiceInfo() *ServiceInfo       { return s.svcInfo }
func (s *udpSession) SendPacket(pak *Packet.Packet) bool { return true }
func (s *udpSession) ShutDown(notify bool)               {}
func (s *udpSession) GetRemoteAddress() net.Addr         { return s.localAddr }
func (s *udpSession) GetLocalAddress() net.Addr          { return s.remoteAddr }
func (s *udpSession) Wake()                              {}
