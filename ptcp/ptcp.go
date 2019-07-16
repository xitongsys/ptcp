package ptcp

import (
	"log"
	"sync"
	"time"

	"github.com/xitongsys/ptcp/header"
	"github.com/xitongsys/ptcp/raw"
)

var BUFFERSIZE = 65535

var ptcpServer *PTCP

func Init(interfaceName string) {
	var err error
	if ptcpServer, err = NewPTCP(interfaceName); err != nil {
		panic(err)
	}
	ptcpServer.Start()
}

type PTCP struct {
	raw *raw.Raw
	//Key: ip:port
	routerListener sync.Map
	//Key: localIp:localPort:remoteIp:remotePort
	router sync.Map
}

func NewPTCP(interfaceName string) (*PTCP, error) {
	r, err := raw.NewRaw(interfaceName)
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
				log.Println("close timeout conn: ", key.(string))
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
			p.raw.Write([]byte(s), key)
		}
	}()
	p.routerListener.Store(key, listener)
}

func (p *PTCP) CreateConn(localAddr string, remoteAddr string, conn *Conn) {
	key := localAddr + ":" + remoteAddr
	go func() {
		for {
			s := <-conn.OutputChan
			p.raw.Write([]byte(s), remoteAddr)
		}
	}()
	p.router.Store(key, conn)
}

func (p *PTCP) CloseConn(key string) {
	log.Println("====close===", key)
	p.router.Delete(key)
}

func (p *PTCP) Start() {
	go func() {
		for {
			data, err := p.raw.Read()
			if err == nil && len(data) > 0 {
				if proto, src, dst, err := header.GetBase(data); err == nil && proto == "tcp" {
					key := dst + ":" + src
					_, _, _, tcpHeader, _, _ := header.Get(data)
					if value, ok := p.router.Load(key); ok {
						conn := value.(*Conn)
						if tcpHeader.Flags == header.FIN {
							go conn.CloseResponse()

						} else if tcpHeader.Flags == header.ACK {
							log.Println("update time: ", key)
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
