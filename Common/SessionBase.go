package Common

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"sync"
	"time"

	"github.com/gookit/slog"
	"github.com/smallnest/ringbuffer"
	"github.com/zhksoftGo/Cactus/Network"
	"github.com/zhksoftGo/Packet"
)

// default custom protocol const
const (
	DefaultHeadLength = 6

	/// <summary> packet for game logic, without encryption. </summary>
	EPacketGameLogic = 0x0

	/// <summary> packet for broadcast message without encryption. </summary>
	EPacketBroadcast = 0x10

	/// <summary> packet for encrypted game logic packet. </summary>
	EPacketGameLogicEncrypted = 0x20

	/// <summary> packet for heart beat, encryption related packets. 心跳、加解密相关，具体内容见包内数据. </summary>
	EPacketNetworkInternal = 0x30

	/// <summary>
	/// 过大的自动分割包
	/// 4 bytes body length + 2 bytes type + body
	/// 4 bytes body length + 2 bytes type + body(4 bytes big Packet index + 2 bytes totalCount + 2 bytes index + body)
	/// </summary>
	EPacketAutoSplitLarge = 0x40
)

func isCorrectAction(actionType uint16) bool {
	switch actionType {
	case EPacketGameLogic, EPacketBroadcast, EPacketGameLogicEncrypted, EPacketNetworkInternal, EPacketAutoSplitLarge:
		return true
	default:
		return false
	}
}

type IMessageHandle interface {
	HandleInComingMsg(pak *Packet.Packet)
}

type ISessionPlayer interface {
	Initialize(session Network.INetworkSession, handler IMessageHandle)
	OnUpdate(dt time.Duration)
	GetID() uint64
	GetName() string
	SetName(n string)
	SendMsg(b []byte) error
	SendPacket(pak Packet.Packet) error
	Shutdown(notify bool)
}

type SessionPlayerBase struct {
	Network.EventHandler
	msgHandlerUpdatable IMessageHandle
	pakRecv             *list.List
	pakUpdate           *list.List
	pakQueueMutex       sync.Mutex
	Name                string
	dataRecv            *ringbuffer.RingBuffer
}

const maxPacketCountPerUpdate = 5

func (s *SessionPlayerBase) Initialize(session Network.INetworkSession, handler IMessageHandle) {
	s.Session = session
	s.msgHandlerUpdatable = handler
	s.pakRecv = list.New()
	s.pakUpdate = list.New()
	s.dataRecv = ringbuffer.New(1024)
}

func (s *SessionPlayerBase) GetID() uint64 {
	return s.Session.GetSessionID()
}

func (s *SessionPlayerBase) GetName() string {
	return s.Name
}

func (s *SessionPlayerBase) SetName(n string) {
	s.Name = n
}

func (s *SessionPlayerBase) Shutdown(notify bool) {
	s.Session.Shutdown(notify)
}

func (ev *SessionPlayerBase) OnRecvMsg(b []byte) Network.Action {

	ev.dataRecv.Write(b)

	if ev.dataRecv.Length() >= DefaultHeadLength {
		// parse header
		headerData := make([]byte, DefaultHeadLength)
		copy(headerData, ev.dataRecv.Bytes()[0:DefaultHeadLength])
		headerBuffer := bytes.NewBuffer(headerData)

		var dataLength uint32
		var actionType uint16
		_ = binary.Read(headerBuffer, binary.LittleEndian, &dataLength)
		_ = binary.Read(headerBuffer, binary.LittleEndian, &actionType)

		// check the protocol actionType,
		if !isCorrectAction(actionType) {
			slog.Warn("abnormal protocol:", actionType, dataLength)
			return Network.Close
		}

		if dataLength == 0 {
			ev.dataRecv.Read(headerData)
			return Network.None
		}

		// parse payload
		frameLen := int(DefaultHeadLength + dataLength)
		if ev.dataRecv.Length() >= frameLen {

			data := make([]byte, frameLen)
			ev.dataRecv.Read(data)

			pak := new(Packet.Packet)
			pak.FromBuff(data)
			ev.OnRecvPacket(pak)

			return Network.None
		}
	}

	return Network.None
}

func (s *SessionPlayerBase) SendMsg(b []byte) error {
	return s.Session.SendMsg(b)
}

func (s *SessionPlayerBase) SendPacket(pak Packet.Packet) error {
	return s.Session.SendMsg(pak.GetUsedBuffer())
}

func (s *SessionPlayerBase) OnRecvPacket(pak *Packet.Packet) {
	s.pakQueueMutex.Lock()
	defer s.pakQueueMutex.Unlock()

	defer func() {
		if err := recover(); err != nil {
			slog.Error(err)
		}
	}()

	//slog.Debug(pak)
	pak.OffsetReadPos(6)
	s.pakRecv.PushBack(pak)
}

func (s *SessionPlayerBase) OnUpdate(dt time.Duration) {
	// slog.Debug("SessionPlayerBase.OnUpdate")
	msgCount := 0
	for msgCount < maxPacketCountPerUpdate && s.pakUpdate.Len() != 0 {
		e := s.pakUpdate.Front()
		s.msgHandlerUpdatable.HandleInComingMsg(e.Value.(*Packet.Packet))
		s.pakUpdate.Remove(e)
		msgCount++
	}

	if s.pakUpdate.Len() == 0 {
		s.pakQueueMutex.Lock()

		t := s.pakUpdate
		s.pakUpdate = s.pakRecv
		s.pakRecv = t

		s.pakQueueMutex.Unlock()
	}

	for msgCount < maxPacketCountPerUpdate && s.pakUpdate.Len() != 0 {
		e := s.pakUpdate.Front()
		s.msgHandlerUpdatable.HandleInComingMsg(e.Value.(*Packet.Packet))
		s.pakUpdate.Remove(e)
		msgCount++
	}
}
