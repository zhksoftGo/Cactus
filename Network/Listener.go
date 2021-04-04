package Network

import (
	"errors"
	"net"
	"os"
	"strings"
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

func (ln *listener) system() error {
	return nil
}

func reuseportListenPacket(proto, addr string) (l net.PacketConn, err error) {
	return nil, errors.New("reuseport is not available")
}

func reuseportListen(proto, addr string) (l net.Listener, err error) {
	return nil, errors.New("reuseport is not available")
}

type addrOpts struct {
	reusePort bool
}

//"tcp-net://localhost:5000?reuseport=1"
func parseAddr(addr string) (network, address string, opts addrOpts, stdlib bool) {
	network = "tcp"
	address = addr
	opts.reusePort = false

	if strings.Contains(address, "://") {
		network = strings.Split(address, "://")[0]
		address = strings.Split(address, "://")[1]
	}

	if strings.HasSuffix(network, "-net") {
		stdlib = true
		network = network[:len(network)-4]
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

type newListener struct {
	ln *listener
}
