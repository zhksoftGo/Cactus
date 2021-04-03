package Network

import (
	"io"
	"time"

	"github.com/zhksoftGo/Packet"
)

type IEventHandler interface {

	// OnOpened fires when a new connection has opened.
	// The info parameter has information about the connection such as
	// it's local and remote address.
	// Use the out return value to write data to the connection.
	// The opts return value is used to set connection options.
	OnOpened() (opts Options, action Action)

	/// 收到对方发来的消息, 包含网络头. 里面具体数据可以是Packet/XML/JSON.
	OnRecvPacket(pak *Packet.Packet) Action

	// OnClosed fires when a connection has closed.
	// The err parameter is the last known connection error.
	OnClosed(err error) (action Action)

	// OnDetached fires when a connection has been previously detached.
	// Once detached it's up to the receiver of this event to manage the
	// state of the connection. The Closed event will not be called for
	// this connection.
	// The rwc parameter is a ReadWriteCloser that represents the
	// underlying socket connection. It can be freely used in goroutines
	// and should be closed when it's no longer needed.
	OnDetached(rwc io.ReadWriteCloser) (action Action)
}

type EventHandler struct {
	session INetworkSession
	ready   bool
}

func (ev EventHandler) OnOpened() (opts Options, action Action) {
	ev.ready = true
	opts = Options{time.Minute, true}
	action = None
	return
}

func (ev EventHandler) OnRecvPacket(pak *Packet.Packet) Action {
	return None
}

func (ev EventHandler) OnClosed(err error) (action Action) {
	action = None
	return
}

func (ev EventHandler) OnDetached(rwc io.ReadWriteCloser) (action Action) {
	rwc.Close()
	action = None
	return
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
	CreateEventHandler(session INetworkSession) IEventHandler

	OnConnectFailed(svcKey string)

	OnShutdown()
}

type EventHandlerManager struct {
}

func (evMngr *EventHandlerManager) CreateEventHandler(session INetworkSession) IEventHandler {
	ev := EventHandler{session: session}
	return ev
}

func (evMngr *EventHandlerManager) OnConnectFailed(svcKey string) {

}

func (evMngr *EventHandlerManager) OnShutdown() {

}
