package main

import (
	"io"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Network"
	"github.com/zhksoftGo/Packet"
)

type EVHandlerGameServer struct {
	Network.EventHandler
}

func (ev EVHandlerGameServer) OnOpened() (opts Network.Options, action Network.Action) {
	slog.Info("EVHandlerGameServer.OnOpened")

	opts = Network.Options{TCPKeepAlive: time.Minute, ReuseInputBuffer: true}
	action = Network.None
	return
}

func (ev EVHandlerGameServer) OnRecvPacket(pak *Packet.Packet) Network.Action {
	slog.Info("EVHandlerGameServer.OnRecvPacket")

	return Network.None
}

func (ev EVHandlerGameServer) OnClosed(err error) (action Network.Action) {
	slog.Info("EVHandlerGameServer.OnClosed")

	action = Network.None
	return
}

func (ev EVHandlerGameServer) OnDetached(rwc io.ReadWriteCloser) (action Network.Action) {
	slog.Info("EVHandlerGameServer.OnDetached")

	rwc.Close()
	action = Network.None
	return
}
