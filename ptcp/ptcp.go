package ptcp

import (
	"sync"
	"time"

	"github.com/xitongsys/ptcp/header"
	"github.com/xitongsys/ptcp/netinfo"
)

var BUFFERSIZE = 65535
var CHANBUFFERSIZE = 1024

var ptcpServer *PTCP
var arp *netinfo.Arp
var route *netinfo.Route
var local *netinfo.Local

func Init(interfaceName string) {
	var err error
	if ptcpServer, err = NewPTCP(interfaceName); err != nil {
		panic(err)
	}

	if arp, err = netinfo.NewArp(); err != nil {
		panic(err)
	}

	if route, err = netinfo.NewRoute(); err != nil {
		panic(err)
	}

	if local, err = netinfo.NewLocal(); err != nil {
		panic(err)
	}

	ptcpServer.Start()
}

type PTCP struct {
	raw *Raw
	//Key: ip:port
	routerListener sync.Map
	//Key: localIp:localPort:remoteIp:remotePort
	router sync.Map
}

func NewPTCP(interfaceName string) (*PTCP, error) {
	r, err := NewRaw(interfaceName)
	if err != nil {
		return nil, err
	}
	return &PTCP{
		raw:            r,
		routerListener: sync.Map{},
		router:         sync.Map{},
	}, nil
}

func (p *PTCP) CleanTimeoutConns() {
	for {
		time.Sleep(time.Second * time.Duration(CONNTIMEOUT))
		p.router.Range(func(key interface{}, value interface{}) bool {
			conn := value.(*Conn)
			if conn.IsTimeout() {
				conn.Close()
			}
			return true
		})
	}
}

func (p *PTCP) CloseListener(key string) {
	p.routerListener.Delete(key)
}

func (p *PTCP) CreateListener(key string, listener *Listener) {
	go func() {
		for {
			s := <-listener.OutputChan
			p.raw.Write([]byte(s))
		}
	}()
	p.routerListener.Store(key, listener)
}

func (p *PTCP) CreateConn(localAddr string, remoteAddr string, conn *Conn) {
	key := localAddr + ":" + remoteAddr
	go func() {
		for {
			s := <-conn.OutputChan
			p.raw.Write([]byte(s))
		}
	}()
	p.router.Store(key, conn)
}

func (p *PTCP) CloseConn(key string) {
	p.router.Delete(key)
}

func (p *PTCP) Start() {
	go func() {
		for {
			data, err := p.raw.Read()
			if err == nil && len(data) > 0 {
				if proto, ipHeader, _, tcpHeader, _, err := header.Get(data); err == nil && proto == "tcp" {
					src, dst := header.GetTcpAddr(ipHeader, tcpHeader)
					key := dst + ":" + src
					if value, ok := p.router.Load(key); ok {
						conn := value.(*Conn)
						if tcpHeader.Flags == header.FIN {
							go conn.CloseResponse()

						} else if tcpHeader.Flags == header.ACK {
							conn.UpdateTime()
						}

						select {
						case conn.InputChan <- string(data):
						default:
						}

					} else if value, ok := p.routerListener.Load(dst); ok {
						listener := value.(*Listener)
						select {
						case listener.InputChan <- string(data):
						default:
						}
					}
				}
			}
		}
	}()

	go p.CleanTimeoutConns()
}
