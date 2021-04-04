package main

import (
	"io"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Network"
	"github.com/zhksoftGo/Packet"
)

type EVHandlerGMServer struct {
	Network.EventHandler
}

func (ev EVHandlerGMServer) OnOpened() (opts Network.Options, action Network.Action) {
	slog.Info("EVHandlerGMServer.OnOpened")

	opts = Network.Options{TCPKeepAlive: time.Minute, ReuseInputBuffer: true}
	action = Network.None
	return
}

func (ev EVHandlerGMServer) OnRecvPacket(pak *Packet.Packet) Network.Action {
	slog.Info("EVHandlerGMServer.OnRecvPacket")

	return Network.None
}

func (ev EVHandlerGMServer) OnClosed(err error) (action Network.Action) {
	slog.Info("EVHandlerGMServer.OnClosed")

	action = Network.None
	return
}

func (ev EVHandlerGMServer) OnDetached(rwc io.ReadWriteCloser) (action Network.Action) {
	slog.Info("EVHandlerGMServer.OnDetached")

	rwc.Close()
	action = Network.None
	return
}
