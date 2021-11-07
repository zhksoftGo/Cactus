package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Cactus/Network"
)

var NetworkModule Network.INetworkModule

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	slog.Configure(func(logger *slog.SugaredLogger) {
		f := logger.Formatter.(*slog.TextFormatter)
		f.EnableColor = true
	})

	wg := sync.WaitGroup{}
	wg.Add(1)

	NetworkModule = Network.NewNetworkModule()

	SessionMgr = CreateSessionManager()
	go SessionMgr.Update(ctx, 33, SessionMgr.OnUpdate)

	go func() {
		slog.Info("Network starting")
		defer func() {
			slog.Info("Network end")
			wg.Done()
		}()

		NetworkModule.Listen("GameServer", "tcp://:9091")
		NetworkModule.Connect("CenterGameClient", "tcp://:9082", 10*time.Second)
		NetworkModule.Run(SessionMgr, 0)
	}()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for sig := range c {
			switch sig {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				slog.Info("Exit with:", sig)

				cancel()

				slog.Info("SessionMgr shutdown")
				SessionMgr.Running = false
				NetworkModule.Shutdown()
				return
			}
		}
	}()

	wg.Wait()
	slog.Info("GameServer end")
}
