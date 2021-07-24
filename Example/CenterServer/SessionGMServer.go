package main

import (
	"io"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Common"
	"github.com/zhksoftGo/Cactus/Network"
	"github.com/zhksoftGo/Packet"
)

type SessionGMServer struct {
	Common.SessionPlayerBase
}

func (ev *SessionGMServer) OnOpened() (opts Network.Options, action Network.Action) {
	slog.Debug("OnOpened:", ev.GetID())

	opts = Network.Options{TCPKeepAlive: time.Minute, ReuseInputBuffer: true}
	action = Network.None
	return
}

func (ev *SessionGMServer) OnClosed(err error) (action Network.Action) {
	slog.Debug("OnClosed:", ev.GetID())

	SessionMgr.RemoveSessionPlayer(ev)

	action = Network.None
	return
}

func (ev *SessionGMServer) OnDetached(rwc io.ReadWriteCloser) (action Network.Action) {
	slog.Debug("OnDetached:", ev.GetID())

	rwc.Close()
	ev.OnClosed(nil)

	return
}

func (ev *SessionGMServer) HandleInComingMsg(pak *Packet.Packet) {

	defer func() {
		if err := recover(); err != nil {
			slog.Error(err)
		}
	}()

}

func (ev *SessionGMServer) OnUpdate(dt time.Duration) {
	ev.SessionPlayerBase.OnUpdate(dt)
}
