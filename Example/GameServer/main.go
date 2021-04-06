package main

import (
	"sync"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Network"
)

type EVHandlerManager struct {
	Network.EventHandlerManager
}

func (evMngr EVHandlerManager) CreateEventHandler(session Network.INetworkSession) Network.IEventHandler {

	switch session.GetServiceKey() {
	case "GameServer":
		slog.Info("CreateEventHandler: GameServer")
		ev := EVHandlerGameServer{}
		ev.Session = session
		return ev

	case "CenterGameClient":
		slog.Info("CreateEventHandler: CenterGameClient")
		ev := EVHandlerCenterGameClient{}
		ev.Session = session
		return ev
	}

	return nil
}

func (evMngr EVHandlerManager) OnConnectFailed(svcKey string) {
	slog.Info("OnConnectFailed:", svcKey)
}

func (evMngr EVHandlerManager) OnShutdown() {
	slog.Info("OnShutdown")
}

func main() {

	slog.Configure(func(logger *slog.SugaredLogger) {
		f := logger.Formatter.(*slog.TextFormatter)
		f.EnableColor = true
	})

	wg := sync.WaitGroup{}
	wg.Add(1)

	var module Network.NetworkModule
	var manager EVHandlerManager

	go func() {
		defer wg.Done()

		module.Listen("GameServer", "tcp://:9091")
		module.Connect("CenterGameClient", "tcp://:9082", 10*time.Second)
		module.Run(manager, 0)
	}()

	wg.Wait()
	slog.Info("end...")
}
