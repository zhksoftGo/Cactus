package Network

import (
	"net"
	"time"
)

type connector struct {
	conn    net.Conn
	svrAddr net.Addr
	network string
	addr    string
	svcKey  string
	timeOut time.Duration
}
