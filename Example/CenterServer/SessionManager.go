package main

import (
	"sync"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Common"
	"github.com/zhksoftGo/Cactus/Network"
)

type EVHandlerManager struct {
	Network.EventHandlerManager
	Common.SessionGroup
}

var SessionMgr *EVHandlerManager

func CreateSessionManager() *EVHandlerManager {
	m := &EVHandlerManager{}
	m.SessionPlayers = make(map[uint64]Common.ISessionPlayer)
	return m
}

func (evMgr *EVHandlerManager) CreateEventHandler(session Network.INetworkSession) Network.IEventHandler {

	switch session.GetServiceKey() {
	case "CenterGameServer":
		slog.Info("CreateEventHandler: CenterGameServer")
		ev := new(SessionCenterServer)
		ev.Initialize(session, ev)
		if !evMgr.AddSessionPlayer(ev) {
			return nil
		}
		return ev

	case "GMServer":
		slog.Info("CreateEventHandler: GMServer")
		ev := new(SessionGMServer)
		ev.Initialize(session, ev)
		if !evMgr.AddSessionPlayer(ev) {
			return nil
		}
		return ev
	}

	return nil
}

func (evMgr *EVHandlerManager) OnConnectFailed(svcKey string) {
	slog.Info("OnConnectFailed:", svcKey)
}

func (evMgr *EVHandlerManager) OnShutdown() {
	slog.Info("OnShutdown")
}

var once sync.Once

func (ev *EVHandlerManager) OnUpdate(dt time.Duration) {
	once.Do(func() {
		slog.Info("OnUpdate")
	})
}
