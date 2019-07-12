package ptcp

import (
	"net"
)

func GetLocalAddr(remoteAddr string) (net.Addr, error) {
	conn, err := net.Dial("udp", remoteAddr)
	if err != nil {
		return nil, err
	}
	return conn.LocalAddr(), nil
}