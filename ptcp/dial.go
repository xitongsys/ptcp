package ptcp

import (
	"net"
	"fmt"
	"time"

	"github.com/xitongsys/ptcp/header"
)

const (
	RETRYTIME = 5
	RETRYINTERVAL = 500
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
	tcpHeader.Flags = header.SYN;
	packet := header.BuildTcpPacket(ipHeader, tcpHeader, []byte{})

	done := make(chan int)
	go func() {
		for i:=0; i<RETRYTIME; i++ {
			select {
			case <- done:
				return
			default:
			}

			conn.WriteWithHeader(packet)
			time.Sleep(time.Millisecond * RETRYINTERVAL)
		}
	}()

	timeOut := time.After(time.Millisecond * RETRYINTERVAL * RETRYTIME)
	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := conn.ReadWithHeader(buf)
		if err != nil {
			return nil, err
		}
		_,_,_,tcpHeader,_,_ = header.Get(buf[:n])
		if tcpHeader.Flags == header.SYNACK && tcpHeader.Ack == 1 {
			close(done)
			break
		}

		select {
		case <- timeOut:
			err = fmt.Errorf("Timeout")
			break
		default:
		}
	}

	if err != nil {
		return nil, err
	}

	seq, ack := 1, tcpHeader.Seq + 1
	ipHeader, tcpHeader = header.BuildTcpHeader(localAddr.String(), remoteAddr)
	tcpHeader.Seq = uint32(seq)
	tcpHeader.Ack = ack
	tcpHeader.Flags = header.ACK
	packet = header.BuildTcpPacket(ipHeader, tcpHeader, []byte{})

	n, err := conn.WriteWithHeader(packet)
	if err != nil || n != len(packet) {
		return nil, fmt.Errorf("packet loss (expect=%v, real=%v) or %v", len(packet), n, err)
	}
	return conn, nil
}