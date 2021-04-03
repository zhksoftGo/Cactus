package Network

import (
	"net"
	"time"
)

type connector struct {
	conn    net.Conn
	lnaddr  net.Addr
	network string
	addr    string
	svcKey  string
	timeOut time.Duration
}
