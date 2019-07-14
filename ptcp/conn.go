package ptcp

import (
	"fmt"
	"time"
	"net"

	"github.com/xitongsys/ptcp/header"
)

var CONNCHANBUFSIZE = 1024

type Conn struct {
	localAddress *Addr
	remoteAddress *Addr
	seq uint32
	ack uint32
	InputChan chan string
	OutputChan chan string
}

func NewConn(localAddr string, remoteAddr string) *Conn {
	return &Conn{
		localAddress: NewAddr(localAddr),
		remoteAddress: NewAddr(remoteAddr),
		seq: 0,
		ack: 0,
		InputChan: make(chan string, CONNCHANBUFSIZE),
		OutputChan: make(chan string, CONNCHANBUFSIZE),
	}
}

func (conn *Conn) Read(b []byte) (n int, err error) {
	defer func(){
		if r := recover(); r != nil {
			n, err = -1, fmt.Errorf("closed: %v", r)
		}
	}()
	s := <- conn.InputChan
	_,_,_,tcpHeader,data,_ := header.Get([]byte(s))
	seq := tcpHeader.Seq
	if seq + 1 >= conn.ack {
		conn.ack = seq + 1
	}

	ls, ln := len(data), len(b)
	l := ls
	if ln < ls {
		l = ln
	}
	for i := 0; i < l; i++ {
		b[i] = data[i]
	}
	return ls, nil	
}

func (conn *Conn) Write(b []byte) (n int, err error) {
	defer func(){
		if r := recover(); r != nil {
			n, err = -1, fmt.Errorf("closed: %v", r)
		}
	}()
	ipHeader, tcpHeader := header.BuildTcpHeader(conn.LocalAddr().String(), conn.RemoteAddr().String())
	tcpHeader.Seq = conn.seq + 1
	conn.seq += 1
	tcpHeader.Ack = conn.ack
	tcpHeader.Flags = 0x18

	packet := header.BuildTcpPacket(ipHeader, tcpHeader, b)
	conn.OutputChan <- string(packet)
	return len(b), nil
}

func (conn *Conn) ReadWithHeader(b []byte) (n int, err error) {
	defer func(){
		if r := recover(); r != nil {
			n, err = -1, fmt.Errorf("closed: %v", r)
		}
	}()
	s := <- conn.InputChan
	data := []byte(s)
	_,_,_,tcpHeader,_,_ := header.Get(data)
	seq := tcpHeader.Seq 
	if seq + 1 > conn.ack {
		conn.ack = seq
	}

	ls, ln := len(data), len(b)
	l := ls
	if ln < ls {
		l = ln
	}
	for i := 0; i < l; i++ {
		b[i] = data[i]
	}
	return ls, nil	
}

func (conn *Conn) WriteWithHeader(b []byte) (n int, err error) {
	defer func(){
		if r := recover(); r != nil {
			n, err = -1, fmt.Errorf("closed: %v", r)
		}
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