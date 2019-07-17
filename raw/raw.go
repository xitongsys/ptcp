package raw

import (
	"net"
	"syscall"

	"github.com/xitongsys/ptcp/header"
	"github.com/xitongsys/ptcp/util"
)

var RAWBUFSIZE = 65535

type Raw struct {
	ifName string
	iface  *net.Interface
	fd     int
	buf    []byte
}

func NewRaw(interfaceName string) (*Raw, error) {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(util.Htons(syscall.ETH_P_ALL)))
	if err != nil {
		return nil, err
	}

	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, err
	}

	if err = syscall.BindToDevice(fd, interfaceName); err != nil {
		return nil, err
	}
	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return nil, err
	}

	return &Raw{
		ifName: interfaceName,
		iface:  iface,
		fd:     fd,
		buf:    make([]byte, RAWBUFSIZE),
	}, nil
}

func (r *Raw) Read() ([]byte, error) {
	n, _, err := syscall.Recvfrom(r.fd, r.buf, 0)

	if err == nil {
		eth := &header.Frame{}
		err = eth.UnmarshalBinary(r.buf[:n])
		return eth.Payload, err
	}
	return nil, err
}

func (r *Raw) Write(data []byte) error {
	eth := &header.Frame{}
	eth.EtherType = header.EtherTypeIPv4
	eth.Destination = header.Broadcast
	eth.Source = r.iface.HardwareAddr
	eth.Payload = data
	data, err := eth.MarshalBinary()
	if err != nil {
		return err
	}

	addr := syscall.SockaddrLinklayer{
		Halen:   6,
		Addr:    [8]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		Ifindex: r.iface.Index,
	}

	return syscall.Sendto(r.fd, data, 0, &addr)
}
