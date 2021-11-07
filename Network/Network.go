package Network

import (
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

var errShutdown = errors.New("network module is closing")
var errCloseConns = errors.New("close conns")

// Action is an action that occurs after the completion of an event.
type Action int

const (
	// None indicates that no action should occur following an event.
	None Action = iota

	// Detach detaches a connection. Not available for UDP connections.
	Detach

	// Close closes the connection.
	Close
)

type addrOpts struct {
	reusePort bool
}

// Network://Address
// like `tcp://192.168.0.10:9851` or `unix://socket`.
//		`tcp://localhost:5000?reuseport=1`
// Valid network schemes:
//  tcp   - bind to both IPv4 and IPv6
//  tcp4  - IPv4
//  tcp6  - IPv6
//  udp   - bind to both IPv4 and IPv6
//  udp4  - IPv4
//  udp6  - IPv6
//  unix  - Unix Domain Socket
type ServerInfo struct {
	Key       string
	Network   string
	Address   string
	IsServer  bool
	ReusePort bool
	//Valid client IP range, for a server. example: "192.168.1.0/24"
	IPRange string
}

type INetworkModule interface {
	AddServerInfo(info *ServerInfo) error
	GetServerInfo(svcKey string) *ServerInfo
	IsClientIPInRange(svcKey, clientip string) bool
	Run(evMngr IEventHandlerManager, numLoops int) error
	Shutdown() error
	Listen(svcKey string, url string) error
	ListenSvc(svcKey string) error
	Connect(svcKey, url string, timeOut time.Duration) error
	ConnectSvc(svcKey string, timeOut time.Duration) error
}

type NetworkModuleBase struct {
	evManager       IEventHandlerManager
	severInfoes     map[string]*ServerInfo
	serverInfoMutex sync.Mutex
}

func (m *NetworkModuleBase) AddServerInfo(info *ServerInfo) error {
	m.serverInfoMutex.Lock()
	defer m.serverInfoMutex.Unlock()

	_, ok := m.severInfoes[info.Key]
	if !ok {
		m.severInfoes[info.Key] = info
		return nil
	}

	return errors.New("ServerInfo already exist")
}

func (m *NetworkModuleBase) GetServerInfo(svcKey string) *ServerInfo {
	m.serverInfoMutex.Lock()
	defer m.serverInfoMutex.Unlock()

	info, ok := m.severInfoes[svcKey]
	if !ok {
		return nil
	}

	return info
}

func (m *NetworkModuleBase) IsClientIPInRange(svcKey, clientip string) bool {
	svcInfo := m.GetServerInfo(svcKey)
	if svcInfo == nil {
		return false
	}

	if len(svcInfo.IPRange) == 0 {
		return true
	}

	if strings.Contains(svcInfo.IPRange, ";") {
		ipRanges := strings.Split(svcInfo.IPRange, ";")
		for _, item := range ipRanges {
			_, subnet, err := net.ParseCIDR(item)
			if err != nil {
				return false
			}
			ip := net.ParseIP(clientip)
			if subnet.Contains(ip) {
				return true
			}
		}
		return false

	} else {
		_, subnet, err := net.ParseCIDR(svcInfo.IPRange)
		if err != nil {
			return false
		}
		ip := net.ParseIP(clientip)
		return subnet.Contains(ip)
	}
}

func (m *NetworkModuleBase) Run(evMngr IEventHandlerManager, numLoops int) error {
	panic("Run: You must implement this function")
	return nil
}

func (m *NetworkModuleBase) Shutdown() error {
	panic("Run: You must implement this function")
	return nil
}

func (m *NetworkModuleBase) Listen(svcKey string, url string) error {

	network, addr, opts := parseAddr(url)

	svcInfo := &ServerInfo{
		Key:       svcKey,
		Network:   network,
		Address:   addr,
		IsServer:  false,
		ReusePort: opts.reusePort}

	err := m.AddServerInfo(svcInfo)
	if err != nil {
		return err
	}

	return m.ListenSvc(svcKey)
}

func (m *NetworkModuleBase) ListenSvc(svcKey string) error {
	panic("ListenSvc: You must implement this function")
	return nil
}

func (m *NetworkModuleBase) Connect(svcKey, url string, timeOut time.Duration) error {

	network, addr, opts := parseAddr(url)

	svcInfo := &ServerInfo{
		Key:       svcKey,
		Network:   network,
		Address:   addr,
		IsServer:  false,
		ReusePort: opts.reusePort}

	err := m.AddServerInfo(svcInfo)
	if err != nil {
		return err
	}

	return m.ConnectSvc(svcKey, timeOut)
}

func (m *NetworkModuleBase) ConnectSvc(svcKey string, timeOut time.Duration) error {
	panic("ConnectSvc: You must implement this function")
	return nil
}

//"tcp://localhost:5000?reuseport=1" -> tcp, localhost:5000, true
func parseAddr(addr string) (network, address string, opts addrOpts) {
	network = "tcp"
	address = addr
	opts.reusePort = false

	if strings.Contains(address, "://") {
		network = strings.Split(address, "://")[0]
		address = strings.Split(address, "://")[1]
	}

	q := strings.Index(address, "?")
	if q != -1 {
		for _, part := range strings.Split(address[q+1:], "&") {
			kv := strings.Split(part, "=")

			if len(kv) == 2 {
				switch kv[0] {
				case "reuseport":
					if len(kv[1]) != 0 {
						switch kv[1][0] {
						default:
							opts.reusePort = kv[1][0] >= '1' && kv[1][0] <= '9'
						case 'T', 't', 'Y', 'y':
							opts.reusePort = true
						}
					}
				}
			}
		}
		address = address[:q]
	}
	return
}
