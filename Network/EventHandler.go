package Network

import (
	"time"

	"github.com/zhksoftGo/Packet"
)

type ILinkEventHandler interface {

	/// 获取底层的NetworkSession.
	GetSession() *INetworkSession

	/// 设置底层的NetworkSession.
	SetSession(s *INetworkSession) bool

	/// 设置标识.
	SetKey(key string)

	/// 获取标识.
	GetKey() string

	/// 服务已经准备好, 在收到此函数调用之前, 不要发消息.
	OnReadyToServe(serviceKey string)

	/// 自身关闭(notify = true时)、对方断开、网络出错、网络服务结束时回调.
	OnClose()

	/// 收到对方发来的消息, 包含网络头. 里面具体数据可以是Packet/XML/JSON.
	OnRecvPacket(pak *Packet.Packet) bool
}

type LinkEventHandler struct {
	session *INetworkSession
	key     string
	isReady bool
}

type ILinkEventHandlerManager interface {
	CreateLinkEventHandler(svcInfo string, s *INetworkSession)
	OnConnectFailed(serviceInfo string)
	OnExit()
	Tick() (delay time.Duration, action Action)
}
