package main

import (
	"io"
	"sync"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Network"
	"github.com/zhksoftGo/Packet"
)

type MyEventHandler struct {
	Network.EventHandler
}

func (ev MyEventHandler) OnOpened() (opts Network.Options, action Network.Action) {
	slog.Info("OnOpened")

	opts = Network.Options{TCPKeepAlive: time.Minute, ReuseInputBuffer: true}
	action = Network.None
	return
}

func (ev MyEventHandler) OnRecvPacket(pak *Packet.Packet) Network.Action {
	slog.Info("OnRecvPacket")

	return Network.None
}

func (ev MyEventHandler) OnClosed(err error) (action Network.Action) {
	slog.Info("OnClosed")

	action = Network.None
	return
}

func (ev MyEventHandler) OnDetached(rwc io.ReadWriteCloser) (action Network.Action) {
	slog.Info("OnDetached")

	rwc.Close()
	action = Network.None
	return
}

type MyEventHandlerManager struct {
	Network.EventHandlerManager
}

func (evMngr MyEventHandlerManager) CreateEventHandler(session Network.INetworkSession) Network.IEventHandler {
	slog.Info("CreateEventHandler")

	ev := MyEventHandler{}
	ev.Session = session
	return ev
}

func (evMngr MyEventHandlerManager) OnConnectFailed(svcKey string) {

}

func (evMngr MyEventHandlerManager) OnShutdown() {

}

func main() {

	slog.Configure(func(logger *slog.SugaredLogger) {
		f := logger.Formatter.(*slog.TextFormatter)
		f.EnableColor = true
	})

	wg := sync.WaitGroup{}
	wg.Add(1)

	var module Network.NetworkModule
	var manager MyEventHandlerManager

	go func() {
		defer wg.Done()

		module.Listen("GameServer", "tcp://:9081")
		module.Run(manager, 0)
	}()

	wg.Wait()
	slog.Info("end...")
}
