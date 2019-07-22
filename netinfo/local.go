package netinfo

import (
	"net"
)

type LocalInterface struct {
	Ip     uint32
	Device string
	Mask   uint32
}

type Local struct {
	localInterfaces map[uint32]LocalInterface
}

func NewLocal() (*Local, error) {
	ifacesMap := map[uint32]LocalInterface{}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			dev, ip, mask := "", uint32(0), uint32(0)
			switch v := addr.(type) {
			case *net.IPAddr:
				dev, ip, mask = iface.Name, s2ip(v.IP.String()), b2ip(v.IP.DefaultMask())
			case *net.IPNet:
				dev, ip, mask = iface.Name, s2ip(v.IP.String()), b2ip(v.Mask)
			default:
				continue
			}

			ifacesMap[ip] = LocalInterface{
				Ip:     ip,
				Device: dev,
				Mask:   mask,
			}
		}
	}

	return &Local{
		localInterfaces: ifacesMap,
	}
}

func (l *Local) GetInterface(ip uint32) LocalInterface {
	if v, ok := l.localInterfaces[ip]; ok {
		return v
	}
	return nil
}
