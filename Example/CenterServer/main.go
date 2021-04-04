package main

import (
	"sync"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Network"
)

type EVHandlerManager struct {
	Network.EventHandlerManager
}

func (evMngr EVHandlerManager) CreateEventHandler(session Network.INetworkSession) Network.IEventHandler {

	switch session.GetServiceKey() {
	case "GMServer":
		slog.Info("CreateEventHandler: GMServer")
		ev := EVHandlerGMServer{}
		ev.Session = session
		return ev

	case "CenterGameServer":
		slog.Info("CreateEventHandler: CenterGameServer")
		ev := EVHandlerCenterGameServer{}
		ev.Session = session
		return ev
	}

	return nil
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

		module.Listen("GMServer", "tcp://:9081")
		module.Listen("CenterGameServer", "tcp://:9082")
		module.Run(manager, 0)
	}()

	wg.Wait()
	slog.Info("end...")
}
