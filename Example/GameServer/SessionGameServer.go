package main

import (
	"io"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Common"
	"github.com/zhksoftGo/Cactus/Network"
	"github.com/zhksoftGo/Packet"
)

type SessionGameServer struct {
	Common.SessionPlayerBase
}

func (ev *SessionGameServer) OnOpened() (opts Network.Options, action Network.Action) {
	slog.Debug("OnOpened:", ev.GetID())

	opts = Network.Options{TCPKeepAlive: time.Minute, ReuseInputBuffer: true}
	action = Network.None
	return
}

func (ev *SessionGameServer) OnRecvMsg(b []byte) Network.Action {
	slog.Info("SessionGameServer.OnRecvMsg")

	return Network.None
}

func (ev *SessionGameServer) OnClosed(err error) (action Network.Action) {
	slog.Debug("OnClosed:", ev.GetID())

	SessionMgr.RemoveSessionPlayer(ev)

	action = Network.None
	return
}

func (ev *SessionGameServer) OnDetached(rwc io.ReadWriteCloser) (action Network.Action) {
	slog.Debug("OnDetached:", ev.GetID())

	rwc.Close()
	ev.OnClosed(nil)

	return
}

func (ev *SessionGameServer) HandleInComingMsg(pak *Packet.Packet) {

	defer func() {
		if err := recover(); err != nil {
			slog.Error(err)
			slog.Debug("\n" + pak.ToHexViewString())
		}
	}()

}

func (ev *SessionGameServer) OnUpdate(dt time.Duration) {

	ev.SessionPlayerBase.OnUpdate(dt)
}
