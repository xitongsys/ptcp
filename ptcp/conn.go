package ptcp

import (
	"fmt"
	"time"
	"net"
)

var CONNCHANBUFSIZE = 1024

type Conn struct {
	localAddress *Addr
	remoteAddress *Addr
	InputChan chan string
	OutputChan chan string
}

func NewConn(localAddr string, remoteAddr string) *Conn {
	return &Conn{
		localAddress: NewAddr(localAddr),
		remoteAddress: NewAddr(remoteAddr),
		InputChan: make(chan string, CONNCHANBUFSIZE),
		OutputChan: make(chan string, CONNCHANBUFSIZE),
	}
}

func (conn *Conn) Read(b []byte) (n int, err error) {
	defer func(){
		recover()
		n, err = -1, fmt.Errorf("closed")
	}()

	s := <- conn.InputChan
	ls, ln := len(s), len(b)
	l := ls
	if ln < ls {
		l = ln
	}
	sb := []byte(s)
	for i := 0; i < l; i++ {
		b[i] = sb[i]
	}
	return ls, nil	
}

func (conn *Conn) Write(b []byte) (n int, err error) {
	defer func(){
		recover()
		n, err = -1, fmt.Errorf("closed")
	}()

	conn.OutputChan <- string(b)
	return len(b), nil
}

func (conn *Conn) Close() error { 
	go func(){
		defer func(){
			recover()
		}()
		close(conn.InputChan)
	}()
	go func(){
		defer func(){
			recover()
		}()
		close(conn.OutputChan)
	}()
	key := conn.LocalAddr().String() + ":" + conn.RemoteAddr().String()
	ptcpServer.CloseConn(key)
	return nil
}

func (conn *Conn) LocalAddr() net.Addr {
	return conn.localAddress
}

func (conn *Conn) RemoteAddr() net.Addr {
	return conn.remoteAddress
}

func (conn *Conn) SetDeadline(t time.Time) error {
	return nil
}

func (conn *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (conn *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}