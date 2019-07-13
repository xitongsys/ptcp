package ptcp

import (
	"net"
)

func Dial(proto string, remoteAddr string) (net.Conn, error) {
	localAddr, err := GetLocalAddr(remoteAddr)
	if err != nil {
		return nil, err
	}

	conn := NewConn(localAddr.String(), remoteAddr)
	ptcpServer.CreateConn(localAddr.String(), remoteAddr, conn)
	return conn, nil

}