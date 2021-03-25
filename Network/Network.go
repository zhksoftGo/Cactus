package Network

import "github.com/zhksoftGo/Packet"

const (
	MsgEncoding_Binary   = 0x1  // raw binary
	MsgEncoding_Packet   = 0x2  // internal Packet binary
	MsgEncoding_XML      = 0x4  // XML
	MsgEncoding_JSON     = 0x8  // JSON
	MsgEncoding_TextHtml = 0x10 // Text/Html
	MsgEncodingMax       = MsgEncoding_Binary | MsgEncoding_Packet | MsgEncoding_XML | MsgEncoding_JSON | MsgEncoding_TextHtml
)

type ServiceInfoBase struct {
	/// <summary> 服务名字</summary>
	name string

	/// <summary> 地址及端口 </summary>
	rawAddr string

	/// <summary> 是否是服务器 </summary>
	bIsServer bool

	/// <summary> 协议类型 </summary>
	proto string

	/// <summary> 消息格式 </summary>
	msgEncoding uint32

	/// <summary> 分配给哪个Worker(只对Reactor有效) </summary>
	workerName string

	/// <summary> 接入的合法地址范围(只对服务器有效) </summary>
	clientIPAddress string

	/// <summary> 最大连接数(只对服务器有效) </summary>
	maxConnections int

	/// <summary> 尝试连接次数. 0 not try, > 0 try most _connectTryTimes times, -1 try forever. </summary>
	connectTryTimes int

	/// <summary> 是否记录收到的消息, 默认false. </summary>
	recordMsg bool

	/// <summary> 超出这个次数后, 断开连接. -1永不断开, 默认-1. </summary>
	heartBeatTimeout int

	/// <summary> 超出这个数后, 断开连接. -1永不断开, 默认5. </summary>
	wrongPackets int

	/// <summary> 超出这个数后, 断开连接. -1永不断开, 默认128. </summary>
	outMsgHighWaterMark int

	//一些统计相关变量
	connections        int
	totalConnections   int
	packetIn           int
	packetOut          int
	networkUpload      int
	networkDownload    int
	curConnectTryTimes int
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
	/// <summary> Gets the type of the protocol.
	/// 		  获取Session的protocol类型.
	/// </summary>
	GetSessionProtocolType() string

	/// <summary> Gets the type of the local MSG.
	/// 		  获取Session的EMsgEncoding类型.
	/// </summary>
	GetMsgEncoding() string

	/// <summary> Gets the ServiceInfoBase. </summary>
	GetServiceInfoBase() *ServiceInfoBase

	/// <summary>
	/// For Reactor only. Gets the name of the session worker.
	/// For Reactor only. 获取Session的Worker名字.
	/// </summary>
	GetSessionWorkerName() string

	/// <summary>
	/// Sends the MSG in binary.
	/// 发送二进制格式消息(里面实际内容也可以是XML或JSON字符串), 原始数据不含6字节网络头.
	/// 适用于TCP/TCPS/UDP/RUDP协议下的消息发送.
	/// 对于TCP/TCPS, Network底层会自动加上网络头.
	/// </summary>
	SendMsgBinaryNoHeader(pData *byte, nLen uint) bool

	/// <summary>
	/// Sends the data in binary.
	/// 通过TCP/TCPS/TCPRaw协议发送二进制格式数据, 数据不含6字节网络头.
	/// 如果在TCP/TCPS下面发送, 只适用于逻辑连接建立之前.
	/// </summary>
	SendMsgBinaryRawTCP(pData *byte, nLen uint) bool

	/// <summary>
	/// 发送Packet二进制格式消息(里面实际内容也可以是XML或JSON字符串), 原始数据含6字节网络头.
	/// 适用于TCP/TCPS/UDP/RUDP协议.
	/// 对于UDP/RUDP, Network底层将去掉６字节网络头.
	/// </summary>
	SendMsgPacket(pData *Packet.Packet) bool

	/// <summary>
	/// Sends the MSG in Packet for broadcasting. Broadcast msg will be not encrypted, even in TCPS.
	/// Others are same as SendMsgPacket().
	/// 广播消息不会被加密, 即使在TCPS协议下.此函数其它方面和SendMsgPacket()相同.
	/// </summary>
	SendMsgPacketForBroadcast(pData *Packet.Packet) bool

	/// <summary>
	/// 发送String格式消息, 适合简单消息测试.
	/// 适用于TCP/TCPS/UDP/RDUP, strMsg将序列化到Packet；对于TCP/TCPS, 还会加上6字节的网络头.
	/// </summary>
	SendMsgString(strMsg string) bool

	/// <summary>
	/// Shuts down the connection.
	/// 关闭一个连接.
	/// Reactor模式下, bNofity参数无效, bNotify总是为true. bNotify为true时，ILinkEventHandler::OnClose()将被调用.
	/// </summary>
	ShutDown(e ENetworkError, bNotify bool)

	/// <summary>
	/// Gets the remote address.
	/// 获取对方IP地址和端口号.
	/// </summary>
	GetRemoteAddress() string

	/// <summary>
	/// Gets the local address.
	/// 获取本地IP地址和端口号.
	/// </summary>
	GetLocalAddress() string
}

type ILinkEventHandler interface {

	/// <summary>
	/// Gets the session.
	/// 获取底层的NetworkSession.
	/// </summary>
	GetSession() *INetworkSession

	/// <summary>
	/// 设置底层的NetworkSession.
	/// </summary>
	SetSession(p *INetworkSession) bool

	/// <summary> 设置标识. </summary>
	SetKey(key string)

	/// <summary> 获取标识. </summary>
	GetKey() string

	/// <summary>
	/// Notify high level, the service is OK to serve.
	/// 服务已经准备好, 在收到此函数调用之前, 不要发消息.
	/// Proactor模式下是多线程调用.
	/// </summary>
	OnReadyToServe(str string)

	/// <summary>
	/// 自身关闭(bNotify = true时)、对方断开、网络出错、网络服务结束时回调.
	/// Proactor模式下是多线程调用.
	/// </summary>
	OnClose(e ENetworkError)

	/// <summary>
	/// 收到对方发来的消息, 包含网络头. 里面具体数据可以是Packet/XML/JSON.
	/// Proactor模式下是多线程调用.
	/// </summary>
	OnRecvMsgPacket(pak *Packet.Packet) bool

	/// <summary>
	/// 收到对方发来的消息, 不含网络头. 里面具体数据可以是Binary/Packet/XML/JSON.
	/// Proactor模式下是多线程调用.
	/// </summary>
	OnRecvMsgPacketNoHeader(pak *Packet.Packet) bool

	/// <summary>
	/// 收到对方发来的流数据, 格式依赖应用自身.
	/// Proactor模式下是多线程调用.
	/// </summary>
	OnRecvMsgTCPRaw(pak *Packet.Packet) bool

	/// <summary>
	/// Ping超时通知, 返回false将断开连接.
	/// Proactor模式下是多线程调用.
	/// </summary>
	OnHeartBeatTimeOut(iTimes int) bool

	/// <summary>
	/// 收到格式错误包的通知, 返回false将断开连接.
	/// Proactor模式下是多线程调用.
	/// </summary>
	OnWrongPacketFormat(iCounts int) bool

	/// <summary>
	/// 发送缓冲区消息个数超过高水位通知, 返回0: 不行动; 1: 清除以前消息; 2:断开连接.
	/// Proactor模式下是多线程调用.
	/// </summary>
	OnOutMsgHighWaterMark(szCount int) int
}

type INetwork interface {
}

type Network struct {
}
