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
	centerClient         *SessionCenterClient
	reconnectCenterTimer Common.SimpleTimer
	centerClientSvcKey   string
}

var SessionMgr *EVHandlerManager

func CreateSessionManager() *EVHandlerManager {
	m := &EVHandlerManager{}
	m.SessionPlayers = make(map[uint64]Common.ISessionPlayer)
	return m
}

func (evMgr *EVHandlerManager) CreateEventHandler(session Network.INetworkSession) Network.IEventHandler {

	switch session.GetServiceKey() {
	case "GameServer":
		slog.Info("CreateEventHandler: GameServer")
		ev := new(SessionGameServer)
		ev.Initialize(session, ev)
		if !evMgr.AddSessionPlayer(ev) {
			return nil
		}
		return ev

	case "CenterGameClient":
		evMgr.centerClient = new(SessionCenterClient)
		evMgr.centerClient.Initialize(session, evMgr.centerClient)
		slog.Infof("CreateEventHandler: CenterGameClient %v", evMgr.centerClient)
		if !evMgr.AddSessionPlayer(evMgr.centerClient) {
			return nil
		}
		return evMgr.centerClient
	}

	return nil
}

func (evMgr *EVHandlerManager) OnConnectFailed(svcKey string) {
	slog.Info("OnConnectFailed:", svcKey)

	if SessionMgr.Running {
		slog.Info("Reconnecting:", svcKey)
		evMgr.centerClientSvcKey = svcKey
		evMgr.reconnectCenterTimer.SetTimer(3*1000, false, false)
	}
}

func (evMgr *EVHandlerManager) OnShutdown() {
	slog.Info("OnShutdown")
}

var once sync.Once

func (evMgr *EVHandlerManager) OnUpdate(dt time.Duration) {
	once.Do(func() {
		slog.Info("OnUpdate")
	})

	if evMgr.reconnectCenterTimer.OnTimer(dt.Milliseconds()) {
		slog.Info("Reconnecting:", evMgr.centerClientSvcKey)

		err := NetworkModule.ConnectSvc(evMgr.centerClientSvcKey, 5*time.Second)
		if err != nil {
			slog.Error(err)
		}
	}
}

func (evMgr *EVHandlerManager) GetCenterClient() *SessionCenterClient {
	return evMgr.centerClient
}

func (evMgr *EVHandlerManager) OnCenterClientShutdown(svcKey string) {
	slog.Info("OnCenterClientShutdown:", svcKey)

	evMgr.centerClient = nil

	if SessionMgr.Running {
		evMgr.centerClientSvcKey = svcKey
		evMgr.reconnectCenterTimer.SetTimer(3*1000, false, false)
	}
}
