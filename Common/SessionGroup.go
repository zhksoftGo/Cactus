package Common

import (
	"context"
	"time"

	"github.com/gookit/slog"
	"github.com/zhksoftGo/Packet"
)

type SessionGroup struct {
	SessionPlayers map[uint64]ISessionPlayer
	sMutex         RecursiveMutex
	Running        bool
}

func (evMgr *SessionGroup) Update(ctx context.Context, dtInMS int, globalUpdateFun func(dt time.Duration)) {
	slog.Info("SessionGroup.Update() begin")

	defer slog.Info("SessionGroup.Update() end")

	evMgr.Running = true

	duration := time.Duration(dtInMS) * time.Millisecond
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info(ctx.Err())
			return
		case <-ticker.C:
			globalUpdateFun(duration)

			evMgr.sMutex.Lock()
			for _, v := range evMgr.SessionPlayers {
				v.OnUpdate(duration)
			}
			evMgr.sMutex.Unlock()
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

func (evMgr *SessionGroup) AddSessionPlayer(s ISessionPlayer) bool {

	if s == nil {
		slog.Error("AddSessionPlayer failed! nil")
		return false
	}

	evMgr.sMutex.Lock()
	defer evMgr.sMutex.Unlock()

	id := s.GetID()
	_, ok := evMgr.SessionPlayers[id]
	if !ok {
		slog.Info("AddSessionPlayer, ID: ", id)
		evMgr.SessionPlayers[id] = s
		return true
	}

	slog.Error("AddSessionPlayer failed, ID already exit: ", s)
	return false
}

func (evMgr *SessionGroup) GetSessionPlayer(id uint64) ISessionPlayer {
	evMgr.sMutex.Lock()
	defer evMgr.sMutex.Unlock()

	s, ok := evMgr.SessionPlayers[id]
	if !ok {
		return nil
	}

	return s
}

func (evMgr *SessionGroup) RemoveSessionPlayer(s ISessionPlayer) {
	evMgr.sMutex.Lock()
	defer evMgr.sMutex.Unlock()

	id := s.GetID()
	delete(evMgr.SessionPlayers, id)
}

func (evMgr *SessionGroup) GetSessionPlayerByName(k string) ISessionPlayer {
	evMgr.sMutex.Lock()
	defer evMgr.sMutex.Unlock()

	for _, v := range evMgr.SessionPlayers {
		if v.GetName() == k {
			return v
		}
	}

	return nil
}

func (evMgr *SessionGroup) BroadcastPacket(pak Packet.Packet) {
	evMgr.sMutex.Lock()
	defer evMgr.sMutex.Unlock()

	for _, v := range evMgr.SessionPlayers {
		v.SendPacket(pak)
	}
}
