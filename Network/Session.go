package Netwrok

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
	localAddr    net.Addr
	remoteAddr   net.Addr
	conn         net.Conn // original connection
	svcInfo      *ServiceInfo
	eventHandler *ILinkEventHandler
	loop         *stdloop // owner loop
}

type wakeReq struct {
	c *tcpSession
}

func (s *TcpSession) GetServiceInfo() *ServiceInfo       { return s.svcInfo }
func (s *TcpSession) SendPacket(pak *Packet.Packet) bool { return true }
func (s *TcpSession) ShutDown(notify bool)               { conn.Close() }
func (s *TcpSession) GetRemoteAddress() net.Addr         { return c.localAddr }
func (s *TcpSession) GetLocalAddress() net.Addr          { return c.remoteAddr }
func (c *TcpSession) Wake()                              { c.loop.ch <- wakeReq{c} }

type stdin struct {
	c  *stdconn
	in []byte
}

type stderr struct {
	c   *stdconn
	err error
}

type udpSession struct {
	addrIndex    int
	localAddr    net.Addr
	remoteAddr   net.Addr
	in           []byte
	eventHandler *ILinkEventHandler
}

func (s *udpSession) GetServiceInfo() *ServiceInfo       { return s.svcInfo }
func (s *udpSession) SendPacket(pak *Packet.Packet) bool { return true }
func (s *udpSession) ShutDown(notify bool)               { conn.Close() }
func (s *udpSession) GetRemoteAddress() net.Addr         { return c.localAddr }
func (s *udpSession) GetLocalAddress() net.Addr          { return c.remoteAddr }
func (c *udpSession) Wake()                              {}
