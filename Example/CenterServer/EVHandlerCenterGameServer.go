package main

import (
	"io"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Network"
	"github.com/zhksoftGo/Packet"
)

type EVHandlerCenterGameServer struct {
	Network.EventHandler
}

func (ev EVHandlerCenterGameServer) OnOpened() (opts Network.Options, action Network.Action) {
	slog.Info("EVHandlerCenterGameServer.OnOpened")

	opts = Network.Options{TCPKeepAlive: time.Minute, ReuseInputBuffer: true}
	action = Network.None
	return
}

func (ev EVHandlerCenterGameServer) OnRecvPacket(pak *Packet.Packet) Network.Action {
	slog.Info("EVHandlerCenterGameServer.OnRecvPacket")

	return Network.None
}

func (ev EVHandlerCenterGameServer) OnClosed(err error) (action Network.Action) {
	slog.Info("EVHandlerCenterGameServer.OnClosed")

	action = Network.None
	return
}

func (ev EVHandlerCenterGameServer) OnDetached(rwc io.ReadWriteCloser) (action Network.Action) {
	slog.Info("EVHandlerCenterGameServer.OnDetached")

	rwc.Close()
	action = Network.None
	return
}
