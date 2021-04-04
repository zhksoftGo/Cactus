package main

import (
	"io"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Network"
)

type EVHandlerCenterGameClient struct {
	Network.EventHandler
}

func (ev EVHandlerCenterGameClient) OnOpened() (opts Network.Options, action Network.Action) {
	slog.Info("EVHandlerCenterGameClient.OnOpened")

	opts = Network.Options{TCPKeepAlive: time.Minute, ReuseInputBuffer: true}
	action = Network.None
	return
}

func (ev EVHandlerCenterGameClient) OnRecvMsg(b []byte) Network.Action {
	slog.Info("EVHandlerCenterGameClient.OnRecvPacket")

	return Network.None
}

func (ev EVHandlerCenterGameClient) OnClosed(err error) (action Network.Action) {
	slog.Info("EVHandlerCenterGameClient.OnClosed")

	action = Network.None
	return
}

func (ev EVHandlerCenterGameClient) OnDetached(rwc io.ReadWriteCloser) (action Network.Action) {
	slog.Info("EVHandlerCenterGameClient.OnDetached")

	rwc.Close()
	action = Network.None
	return
}
