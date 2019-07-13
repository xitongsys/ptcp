package ptcp

import (
	"net"
	"fmt"

	"github.com/xitongsys/ptcp/header"
)

func Dial(proto string, remoteAddr string) (net.Conn, error) {
	localAddr, err := GetLocalAddr(remoteAddr)
	if err != nil {
		return nil, err
	}

	conn := NewConn(localAddr.String(), remoteAddr)
	ptcpServer.CreateConn(localAddr.String(), remoteAddr, conn)

	ipHeader, tcpHeader := header.BuildTcpHeader(localAddr.String(), remoteAddr)
	tcpHeader.Seq = 0
	tcpHeader.Flags = 0x02;
	packet := header.BuildTcpPacket(ipHeader, tcpHeader, []byte{})

	n, err := conn.WriteWithHeader(packet)
	if err != nil || n != len(packet) {
		return nil, fmt.Errorf("packet loss (expect=%v, real=%v) or %v", len(packet), n, err)
	}

	buf := make([]byte, BUFFERSIZE)
	n, err = conn.Read(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[:n]
	_,_,_,tcpHeader,_,_ = header.Get(buf)
	seq, ack := 1, tcpHeader.Seq + 1
	ipHeader, tcpHeader = header.BuildTcpHeader(localAddr.String(), remoteAddr)
	tcpHeader.Seq = uint32(seq)
	tcpHeader.Ack = ack
	tcpHeader.Flags = 0x10
	packet = header.BuildTcpPacket(ipHeader, tcpHeader, []byte{})
	n, err = conn.WriteWithHeader(packet)
	if err != nil || n != len(packet) {
		return nil, fmt.Errorf("packet loss (expect=%v, real=%v) or %v", len(packet), n, err)
	}

	return conn, nil
}