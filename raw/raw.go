package raw

import (
	"syscall"

	"github.com/xitongsys/ptcp/header"
	"github.com/xitongsys/ptcp/util"
	"github.com/mdlayher/ethernet"
)

var RAWBUFSIZE = 65535

type Raw struct {
	ifName string
	fdWrite int
	fdRead int
	buf []byte
}

func NewRaw(interfaceName string) (*Raw, error){
	fdW, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		return nil, err
	}

	fdR, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(util.Htons(syscall.ETH_P_ALL)))
	if err != nil {
		return nil, err
	}

	/*
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, err
	}
	*/

	if err = syscall.BindToDevice(fdR, interfaceName); err!=nil{
		return nil, err
	}
	if err = syscall.SetsockoptInt(fdR, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err!=nil{
		return nil, err
	}

	return &Raw{
		ifName: interfaceName,
		fdWrite: fdW, 
		fdRead: fdR,
		buf: make([]byte, RAWBUFSIZE),
	}, nil
}

func (r *Raw) Read() ([]byte, error) {
	n, _, err := syscall.Recvfrom(r.fdRead, r.buf, 0)
	if err == nil {
		eth := &ethernet.Frame{}
		eth.UnmarshalBinary(r.buf[:n])
		return eth.Payload, err
	}
	return nil, err
}

func (r *Raw) Write(data []byte, addrs string) error {
	ip, port := header.ParseAddr(addrs)

	addr := syscall.SockaddrInet4 {
		Port: port,
		Addr: header.IpStr2Bytes(ip),
	}
	return syscall.Sendto(r.fdWrite, data, 0, &addr)
}