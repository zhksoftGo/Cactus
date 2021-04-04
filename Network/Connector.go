package Network

import (
	"time"
)

type connector struct {
	network string
	addr    string
	svcKey  string
	timeOut time.Duration
}
