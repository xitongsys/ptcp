package ptcp

import (
	"net"

	"github.com/xitongsys/ptcp/header"
)

var LISTENERBUFSIZE = 1024

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
}

func NewListener(addr string) (*Listener, error) {
	return &Listener{
		Address: addr,
		InputChan: make(chan string, LISTENERBUFSIZE),
		OutputChan: make(chan string, LISTENERBUFSIZE),
	}, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	for {
		packet := <- l.InputChan
		_, _, _, tcpHeader, data, _ := header.Get([]byte(packet))
		if tcpHeader.Flags != 0x02 || len(data) != 0 {
			continue
		}
		_, src, dst, _ := header.GetBase([]byte(packet))
		seq, ack := 0, tcpHeader.Seq + 1
		ipHeaderTo, tcpHeaderTo := header.BuildTcpHeader(dst, src)
		tcpHeaderTo.Seq, tcpHeaderTo.Ack = uint32(seq), uint32(ack)
		tcpHeaderTo.Flags = 0x12
		l.OutputChan <- string(header.BuildTcpPacket(ipHeaderTo, tcpHeaderTo, []byte{}))

		packet = <- l.InputChan
		_, _, _, tcpHeader, data, _ = header.Get([]byte(packet))
		if tcpHeader.Flags != 0x10 || len(data) != 0 || tcpHeader.Seq != ack{
			continue
		}

		conn := NewConn(dst, src)
		ptcpServer.CreateConn(dst, src, conn)
		return conn, nil
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