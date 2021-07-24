package main

import (
	"io"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Common"
	"github.com/zhksoftGo/Cactus/Network"
	"github.com/zhksoftGo/Packet"
)

type SessionCenterServer struct {
	Common.SessionPlayerBase
}

func (ev *SessionCenterServer) OnOpened() (opts Network.Options, action Network.Action) {
	slog.Debug("OnOpened:", ev.GetID())

	opts = Network.Options{TCPKeepAlive: time.Minute, ReuseInputBuffer: true}
	action = Network.None
	return
}

func (ev *SessionCenterServer) OnClosed(err error) (action Network.Action) {
	slog.Debug("OnClosed:", ev.GetID())

	SessionMgr.RemoveSessionPlayer(ev)

	action = Network.None
	return
}

func (ev *SessionCenterServer) OnDetached(rwc io.ReadWriteCloser) (action Network.Action) {
	slog.Debug("OnDetached:", ev.GetID())

	rwc.Close()
	ev.OnClosed(nil)

	return
}

func (ev *SessionCenterServer) HandleInComingMsg(pak *Packet.Packet) {

	defer func() {
		if err := recover(); err != nil {
			slog.Error(err)
		}
	}()

}

func (ev *SessionCenterServer) OnUpdate(dt time.Duration) {
	ev.SessionPlayerBase.OnUpdate(dt)
}
