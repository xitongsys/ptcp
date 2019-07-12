package ptcp

import (
	"net"

	"github.com/xitongsys/ptcp/header"
)

var LISTENERBUFSIZE = 1024

type Listener struct {
	InputChan chan string
	OutputChan chan string
}

func NewListener() (*Listener, error) {
	return &Listener{
		InputChan: make(chan string, LISTENERBUFSIZE),
		OutputChan: make(chan string, LISTENERBUFSIZE),
	}, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	for {
		packet := <- l.InputChan
		_, ipHeader, _, tcpHeader, data, _ := header.Get([]byte(packet))
		if tcpHeader.Flags != 0x02 || len(data) != 0 {
			continue
		}
		

	}
}

func (l *Listener) Close() (error) {

}

func (l *Listener) Addr() net.Addr {

}