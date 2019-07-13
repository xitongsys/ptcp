package ptcp

import (
	"sync"
	"log"

	"github.com/xitongsys/ptcp/raw"
	"github.com/xitongsys/ptcp/header"
)

var BUFFERSIZE = 65535

var ptcpServer *PTCP

func init() {
	var err error 
	if ptcpServer, err = NewPTCP(); err != nil {
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

func NewPTCP() (*PTCP, error) {
	r, err := raw.NewRaw()
	if err != nil {
		return nil, err
	}
	return &PTCP{
		raw: r,
		routerListener: sync.Map{},
		router: sync.Map{},
	}, nil
}

func (p *PTCP) CloseListener(key string) {
	p.routerListener.Delete(key)
}

func (p *PTCP) CreateListener(key string, listener *Listener) {
	go func(){
		for {
			s := <- listener.OutputChan
			p.raw.Write([]byte(s), key)
		}
	}()
	p.routerListener.Store(key, listener)
}

func (p *PTCP) CreateConn(localAddr string, remoteAddr string, conn *Conn) {
	key := localAddr + ":" + remoteAddr
	go func(){
		for {
			s := <- conn.OutputChan
			p.raw.Write([]byte(s), remoteAddr)
		}
	}()
	p.router.Store(key, conn)
}

func (p *PTCP) CloseConn(key string){
	p.router.Delete(key)
}

func (p *PTCP) Start() {
	go func(){
		for {
			data, err := p.raw.Read()
			if err == nil && len(data) > 0 {
				if proto, src, dst, err := header.GetBase(data); err == nil && proto == "tcp" {
					key := dst + ":" + src
					if value, ok := p.router.Load(key); ok {
						conn := value.(Conn)
						select {
						case conn.InputChan <- string(data):
						}

					}else if value, ok := p.routerListener.Load(dst); ok {
						log.Println("key======", key)
						listener := value.(*Listener)
						select {
						case listener.InputChan <- string(data):
						}
					}
				}
			}
		}
	}()
}