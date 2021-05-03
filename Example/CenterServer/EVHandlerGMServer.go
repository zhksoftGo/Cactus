package main

import (
	"io"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Network"
)

type EVHandlerGMServer struct {
	Network.EventHandler
}

func (ev *EVHandlerGMServer) OnOpened() (opts Network.Options, action Network.Action) {
	slog.Info("EVHandlerGMServer.OnOpened")

	opts = Network.Options{TCPKeepAlive: time.Minute, ReuseInputBuffer: true}
	action = Network.None
	return
}

func (ev *EVHandlerGMServer) OnRecvMsg(b []byte) Network.Action {
	slog.Info("EVHandlerGMServer.OnRecvMsg")

	return Network.None
}

func (ev *EVHandlerGMServer) OnClosed(err error) (action Network.Action) {
	slog.Info("EVHandlerGMServer.OnClosed")

	action = Network.None
	return
}

func (ev *EVHandlerGMServer) OnDetached(rwc io.ReadWriteCloser) (action Network.Action) {
	slog.Info("EVHandlerGMServer.OnDetached")

	rwc.Close()
	action = Network.None
	return
}
