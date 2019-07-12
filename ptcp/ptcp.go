package ptcp

import (
	"fmt"
	"net"
	"sync"

	"github.com/xitongsys/ptcp/raw"
	"github.com/xitongsys/ptcp/header"
)

var ptcpServer *PTCP

func init() {
	var err error 
	if ptcpServer, err = NewPTCP(); err != nil {
		panic(err)
	}
	go func() {
		ptcpServer.Start()
	}()
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

func (p *PTCP) Start() {
	go func(){
		for {
			data, err := p.raw.Read()
			if err != nil && len(data) > 0 {
				if proto, src, dst, err := header.GetBase(data); err == nil && proto == "tcp" {
					key := dst + ":" + src
					if value, ok := p.router.Load(key); ok {
						_,_,_,_,tcpData,_ := header.Get(data)
						conn := value.(Conn)
						select {
						case conn.InputChan <- string(tcpData):
						}

					}else if value, ok := p.routerListener.Load(dst); ok {
						listener := value.(Listener)
						select {
						case listener.InputChan <- string(data):
						}
					}
				}
			}
		}

	}()
}