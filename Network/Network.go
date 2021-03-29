package Network

import (
	"net"

	"github.com/zhksoftGo/Packet"
)

const (
	MsgEncoding_Binary   = 0x1  // raw binary
	MsgEncoding_Packet   = 0x2  // internal Packet binary
	MsgEncoding_XML      = 0x4  // XML
	MsgEncoding_JSON     = 0x8  // JSON
	MsgEncoding_TextHtml = 0x10 // Text/Html
	MsgEncodingMax       = MsgEncoding_Binary | MsgEncoding_Packet | MsgEncoding_XML | MsgEncoding_JSON | MsgEncoding_TextHtml
)

type ServiceInfoBase struct {
	///  服务名字</summary>
	name string

	///  地址及端口
	rawAddr string

	///  是否是服务器
	bIsServer bool

	///  协议类型
	proto string

	///  消息格式
	msgEncoding uint32

	///  接入的合法地址范围(只对服务器有效)
	clientIPAddress string

	///  尝试连接次数. 0 not try, > 0 try most _connectTryTimes times, -1 try forever.
	connectTryTimes int

	///  超出这个次数后, 断开连接. -1永不断开, 默认-1.
	heartBeatTimeout int

	///  超出这个数后, 断开连接. -1永不断开, 默认5.
	wrongPackets int

	///  超出这个数后, 断开连接. -1永不断开, 默认128.
	outMsgHighWaterMark int
}

type ENetworkError uint32

const (
	NetworkError_InitiativeClose  ENetworkError = 0 + iota //己方主动关闭
	NetworkError_PeerDisconnect                            //对方主动断开连接
	NetworkError_InitRead                                  //投递异步读操作失败
	NetworkError_InitWrite                                 //投递异步写操作失败
	NetworkError_Read                                      //读操作失败
	NetworkError_Write                                     //写操作失败
	NetworkError_InvalidData                               //接收到非法数据
	NetworkError_Network                                   //网络出错
	NetworkError_HeartBeatTimeOut                          //心跳次数超时
	NetworkError_MsgHighWaterMark                          //消息数量超出水位
	NetworkError_ServiceDown                               //网络服务关闭
	NetworkError_OperationTimeout                          //操作超时
	NetworkError_Unknown                                   //原因未知/未指定
)

type INetworkSession interface {
	/// 获取Session的protocol类型.
	GetSessionProtocolType() string

	/// 获取Session的EMsgEncoding类型.
	GetMsgEncoding() string

	///  Gets the ServiceInfoBase.
	GetServiceInfoBase() *ServiceInfoBase

	/// 发送二进制格式消息(里面实际内容也可以是XML或JSON字符串), 原始数据不含6字节网络头.
	/// 适用于TCP/TCPS/UDP/RUDP协议下的消息发送.
	/// 对于TCP/TCPS, Network底层会自动加上网络头.
	SendMsgBinaryNoHeader(pData *byte, nLen uint) bool

	/// Sends the data in binary.
	/// 通过TCP/TCPS/TCPRaw协议发送二进制格式数据, 数据不含6字节网络头.
	/// 如果在TCP/TCPS下面发送, 只适用于逻辑连接建立之前.
	SendMsgBinaryRawTCP(pData *byte, nLen uint) bool

	/// 发送Packet二进制格式消息(里面实际内容也可以是XML或JSON字符串), 原始数据含6字节网络头.
	/// 适用于TCP/TCPS/UDP/RUDP协议.
	/// 对于UDP/RUDP, Network底层将去掉６字节网络头.
	SendMsgPacket(pData *Packet.Packet) bool

	/// 发送String格式消息, 适合简单消息测试.
	/// 适用于TCP/TCPS/UDP/RDUP, strMsg将序列化到Packet；对于TCP/TCPS, 还会加上6字节的网络头.
	SendMsgString(strMsg string) bool

	/// 关闭一个连接.
	ShutDown(e ENetworkError, bNotify bool)

	/// 获取对方IP地址和端口号.
	GetRemoteAddress() net.Addr

	/// 获取本地IP地址和端口号.
	GetLocalAddress() net.Addr
}

type ILinkEventHandler interface {

	/// 获取底层的NetworkSession.
	GetSession() *INetworkSession

	/// 设置底层的NetworkSession.
	SetSession(p *INetworkSession) bool

	/// 设置标识.
	SetKey(key string)

	/// 获取标识.
	GetKey() string

	/// 服务已经准备好, 在收到此函数调用之前, 不要发消息.
	OnReadyToServe(str string)

	/// 自身关闭(bNotify = true时)、对方断开、网络出错、网络服务结束时回调.
	OnClose(e ENetworkError)

	/// 收到对方发来的消息, 包含网络头. 里面具体数据可以是Packet/XML/JSON.
	OnRecvMsgPacket(pak *Packet.Packet) bool

	/// 收到对方发来的消息, 不含网络头. 里面具体数据可以是Binary/Packet/XML/JSON.
	OnRecvMsgPacketNoHeader(pak *Packet.Packet) bool

	/// 收到对方发来的流数据, 格式依赖应用自身.
	OnRecvMsgTCPRaw(pak *Packet.Packet) bool

	/// Ping超时通知, 返回false将断开连接.
	OnHeartBeatTimeOut(iTimes int) bool

	/// 收到格式错误包的通知, 返回false将断开连接.
	OnWrongPacketFormat(iCounts int) bool

	/// 发送缓冲区消息个数超过高水位通知, 返回0: 不行动; 1: 清除以前消息; 2:断开连接.
	OnOutMsgHighWaterMark(szCount int) int
}

type Network struct {
}
