package Network

import (
	"errors"
	"net"
	"os"
)

type listener struct {
	ln          net.Listener             //tcp listener
	lnaddr      net.Addr                 //address listen on
	pconn       net.PacketConn           //udp listener
	opts        addrOpts                 //reusePort?
	network     string                   //udp, tcp...
	addr        string                   //raw addr 127.0.0.1:80
	svcKey      string                   // service name
	udpSessions map[net.Addr]*udpSession //all udp sessions through this listener
}

func (ln *listener) close() {
	if ln.ln != nil {
		ln.ln.Close()
	}

	if ln.pconn != nil {
		ln.pconn.Close()
	}

	if ln.network == "unix" {
		os.RemoveAll(ln.addr)
	}
}

func reuseportListenPacket(proto, addr string) (l net.PacketConn, err error) {
	return nil, errors.New("reuseport is not available")
}

func reuseportListen(proto, addr string) (l net.Listener, err error) {
	return nil, errors.New("reuseport is not available")
}

type newListener struct {
	ln *listener
}
