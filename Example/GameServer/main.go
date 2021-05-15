package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Network"
)

type EVHandlerManager struct {
	Network.EventHandlerManager
}

func (evMngr *EVHandlerManager) CreateEventHandler(session Network.INetworkSession) Network.IEventHandler {

	switch session.GetServiceKey() {
	case "GameServer":
		slog.Info("CreateEventHandler: GameServer")
		ev := new(EVHandlerGameServer)
		ev.Session = session
		return ev

	case "CenterGameClient":
		slog.Info("CreateEventHandler: CenterGameClient")
		ev := new(EVHandlerCenterGameClient)
		ev.Session = session
		return ev
	}

	return nil
}

func (evMngr *EVHandlerManager) OnConnectFailed(svcKey string) {
	slog.Error("OnConnectFailed:", svcKey)
}

func (evMngr *EVHandlerManager) OnShutdown() {
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
	manager := new(EVHandlerManager)

	go func() {
		slog.Info("Network starting")
		defer func() {
			slog.Info("Network end")
			wg.Done()
		}()

		module.Listen("GameServer", "tcp://:9091")
		module.Connect("CenterGameClient", "tcp://:9082", 10*time.Second)
		module.Run(manager, 0)
	}()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for sig := range c {
			switch sig {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				slog.Info("Exit with:", sig)
				module.Shutdown()
				return
			}
		}
	}()

	wg.Wait()
	slog.Info("end...")
}
