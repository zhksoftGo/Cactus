package main

import (
	"io"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Common"
	"github.com/zhksoftGo/Cactus/Network"
	"github.com/zhksoftGo/Packet"
)

type SessionCenterClient struct {
	Common.SessionPlayerBase
}

func (ev *SessionCenterClient) OnOpened() (opts Network.Options, action Network.Action) {
	slog.Debug("OnOpened:", ev.GetID())

	opts = Network.Options{TCPKeepAlive: time.Minute, ReuseInputBuffer: true}
	action = Network.None
	return
}

func (ev *SessionCenterClient) OnClosed(err error) (action Network.Action) {
	slog.Debug("OnClosed:", ev.GetID())

	action = Network.None
	SessionMgr.RemoveSessionPlayer(ev)

	SessionMgr.OnCenterClientShutdown(ev.Session.GetServiceKey())
	return
}

func (ev *SessionCenterClient) OnDetached(rwc io.ReadWriteCloser) (action Network.Action) {
	slog.Debug("OnDetached:", ev.GetID())

	rwc.Close()
	ev.OnClosed(nil)

	return
}

func (ev *SessionCenterClient) HandleInComingMsg(pak *Packet.Packet) {

	defer func() {
		if err := recover(); err != nil {
			slog.Error(err)
			slog.Error("\n" + pak.ToHexViewString())
		}
	}()

}

func (ev *SessionCenterClient) OnUpdate(dt time.Duration) {
	ev.SessionPlayerBase.OnUpdate(dt)
}
