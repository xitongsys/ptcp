package ptcp

import (
	"net"
	"time"

	"github.com/xitongsys/ptcp/header"
	"github.com/patrickmn/go-cache"
)

var LISTENERBUFSIZE = 1024

const (
	SYN = 0x02
	ACK = 0x10
	SYNACK = 0x12
)

func Listen(proto, addr string) (net.Listener, error) {
	if listener, err := NewListener(addr); err == nil {
		ptcpServer.CreateListener(addr, listener)
		return listener, err

	}else{
		return nil, err
	}
}

type Listener struct {
	Address string
	InputChan chan string
	OutputChan chan string

	requestCache *cache.Cache
}

func NewListener(addr string) (*Listener, error) {
	return &Listener{
		Address: addr,
		InputChan: make(chan string, LISTENERBUFSIZE),
		OutputChan: make(chan string, LISTENERBUFSIZE),

		requestCache: cache.New(30*time.Second, 10*time.Minute),

	}, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	for {
		packet := <- l.InputChan
		_, _, _, tcpHeader, data, _ := header.Get([]byte(packet))
		_, src, dst, _ := header.GetBase([]byte(packet))
		if tcpHeader.Flags == SYN && len(data) == 0 {
			seq, ack := 0, tcpHeader.Seq + 1
			ipHeaderTo, tcpHeaderTo := header.BuildTcpHeader(dst, src)
			tcpHeaderTo.Seq, tcpHeaderTo.Ack = uint32(seq), uint32(ack)
			tcpHeaderTo.Flags = SYNACK
			l.requestCache.Set(src, ack, cache.DefaultExpiration)
			l.OutputChan <- string(header.BuildTcpPacket(ipHeaderTo, tcpHeaderTo, []byte{}))

		}else if tcpHeader.Flags == ACK && len(data) == 0 {
			if acki, ok := l.requestCache.Get(src); ok && acki.(uint32) == tcpHeader.Seq {
				l.requestCache.Delete(src)
				conn := NewConn(dst, src)
				ptcpServer.CreateConn(dst, src, conn)
				return conn, nil
			}
		}
	}
}

func (l *Listener) Close() (error) {
	go func(){
		defer func(){
			recover()
		}()
		close(l.InputChan)
	}()

	go func(){
		defer func(){
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