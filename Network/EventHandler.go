package Network

import (
	"io"
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

// Options are set when the client opens.
type Options struct {
	// TCPKeepAlive (SO_KEEPALIVE) socket option.
	TCPKeepAlive time.Duration
	// ReuseInputBuffer will forces the connection to share and reuse the
	// same input packet buffer with all other connections that also use
	// this option.
	// Default value is false, which means that all input data which is
	// passed to the Data event will be a uniquely copied []byte slice.
	ReuseInputBuffer bool
}

type IEventHandlerManager interface {
	CreateEventHandler(svcInfo string, s *INetworkSession)

	OnConnectFailed(serviceInfo string)

	OnExit()

	// Opened fires when a new connection has opened.
	// The info parameter has information about the connection such as
	// it's local and remote address.
	// Use the out return value to write data to the connection.
	// The opts return value is used to set connection options.
	Opened(c INetworkSession) (out []byte, opts Options, action Action)

	// Closed fires when a connection has closed.
	// The err parameter is the last known connection error.
	Closed(c INetworkSession, err error) (action Action)

	// Detached fires when a connection has been previously detached.
	// Once detached it's up to the receiver of this event to manage the
	// state of the connection. The Closed event will not be called for
	// this connection.
	// The conn parameter is a ReadWriteCloser that represents the
	// underlying socket connection. It can be freely used in goroutines
	// and should be closed when it's no longer needed.
	Detached(c INetworkSession, rwc io.ReadWriteCloser) (action Action)

	// PreWrite fires just before any data is written to any client socket.
	PreWrite()

	// Data fires when a connection sends the server data.
	// The in parameter is the incoming data.
	// Use the out return value to write data to the connection.
	Data(c INetworkSession, in []byte) (out []byte, action Action)

	// Tick fires immediately after the server starts and will fire again
	// following the duration specified by the delay return value.
	Tick() (delay time.Duration, action Action)
}
