package ptcp

import (
	"net"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/xitongsys/ethernet-go/header"
)

var LISTENERBUFSIZE = 1024

func Listen(proto, addr string) (net.Listener, error) {
	if listener, err := NewListener(addr); err == nil {
		ptcpServer.CreateListener(addr, listener)
		return listener, err

	} else {
		return nil, err
	}
}

type Listener struct {
	Address    string
	InputChan  chan string
	OutputChan chan string

	requestCache *cache.Cache
}

func NewListener(addr string) (*Listener, error) {
	listener := &Listener{
		Address:    addr,
		InputChan:  make(chan string, LISTENERBUFSIZE),
		OutputChan: make(chan string, LISTENERBUFSIZE),

		requestCache: cache.New(10*time.Second, 1*time.Minute),
	}
	listener.sendResponse()
	return listener, nil
}

func (l *Listener) sendResponse() {
	go func() {
		for {
			items := l.requestCache.Items()
			for src := range items {
				if respi, ok := l.requestCache.Get(src); ok {
					resp := respi.(string)
					l.OutputChan <- resp
				}
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()
}

func (l *Listener) Accept() (net.Conn, error) {
	for {
		packet := <-l.InputChan
		_, ipHeader, _, tcpHeader, data, _ := header.Get([]byte(packet))
		src, dst := header.GetTcpAddr(ipHeader, tcpHeader)
		if tcpHeader.Flags == header.SYN && len(data) == 0 {
			seq, ack := 0, tcpHeader.Seq+1
			ipHeaderTo, tcpHeaderTo := header.BuildTcpHeader(dst, src)
			tcpHeaderTo.Seq, tcpHeaderTo.Ack = uint32(seq), uint32(ack)
			tcpHeaderTo.Flags = (header.SYN | header.ACK)
			response := string(header.BuildTcpPacket(ipHeaderTo, tcpHeaderTo, []byte{}))
			l.requestCache.Set(src, response, cache.DefaultExpiration)
			l.OutputChan <- response

		} else if tcpHeader.Flags == header.ACK && len(data) == 0 {
			if _, ok := l.requestCache.Get(src); ok {
				l.requestCache.Delete(src)
				conn := NewConn(dst, src, CONNECTED)
				ptcpServer.CreateConn(dst, src, conn)
				return conn, nil
			}
		}
	}
}

func (l *Listener) Close() error {
	go func() {
		defer func() {
			recover()
		}()
		close(l.InputChan)
	}()

	go func() {
		defer func() {
			recover()
		}()
		close(l.OutputChan)
	}()
	ptcpServer.CloseListener(l.Address)
	return nil
}

func (l *Listener) Addr() net.Addr {
	return nil
}
