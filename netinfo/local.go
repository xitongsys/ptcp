package netinfo

import (
	"fmt"
	"net"
)

type LocalInterface struct {
	Ip     uint32
	Device string
	Mask   uint32
}

func (li *LocalInterface) String() string {
	return fmt.Sprintf("{Ip:%v, Device:%v, Mask:%v}", ip2s(li.Ip), li.Device, ip2s(li.Mask))
}

type Local struct {
	localInterfaces map[uint32]*LocalInterface
}

func NewLocal() (*Local, error) {
	ifacesMap := map[uint32]*LocalInterface{}

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
				ip, err = s2ip(v.IP.String())
				if err != nil {
					continue
				}

				mask, err = b2ip(v.IP.DefaultMask())
				if err != nil {
					continue
				}

				dev = iface.Name

			case *net.IPNet:
				ip, err = s2ip(v.IP.String())
				if err != nil {
					continue
				}

				mask, err = b2ip(v.Mask)
				if err != nil {
					continue
				}

				dev = iface.Name

			default:
				continue
			}

			ifacesMap[ip] = &LocalInterface{
				Ip:     ip,
				Device: dev,
				Mask:   mask,
			}
		}
	}

	return &Local{
		localInterfaces: ifacesMap,
	}, nil
}

func (l *Local) String() string {
	res := "{"
	for _, li := range l.localInterfaces {
		res += li.String()
	}
	res += "}"
	return res
}

func (l *Local) GetInterfaceByIp(ip uint32) (*LocalInterface, error) {
	if v, ok := l.localInterfaces[ip]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("ip %v not found", ip2s(ip))
}

func (l *Local) GetInterfaceByName(name string) (*LocalInterface, error) {
	for _, iface := range l.localInterfaces {
		if iface.Device == name {
			return iface, nil
		}
	}
	return nil, fmt.Errorf("interface %v not found", name)
}
