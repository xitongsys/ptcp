package netinfo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var ARPPATH = "/proc/net/arp"

type ArpItem struct {
	Ip     uint32
	Device string
	HwAddr []byte
}

func (ai *ArpItem) String() string {
	return fmt.Sprintf("{ip:%v, dev:%v, hw:%v}", ip2s(ai.Ip), ai.Device, ai.HwAddr)
}

type Arp struct {
	arps map[uint32]*ArpItem
}

func NewArp() (*Arp, error) {
	r := &Arp{}
	err := r.Load(ARPPATH)
	return r, err
}

func (r *Arp) String() string {
	res := "{"
	for _, item := range r.arps {
		res += item.String()
	}
	res += "}"
	return res
}

func (r *Arp) Load(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}

	defer f.Close()
	reader := bufio.NewReader(f)
	_, _, err = reader.ReadLine()
	if err != nil {
		return err
	}

	r.arps = map[uint32]*ArpItem{}

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		ss := strings.Fields(string(line))
		if len(ss) < 3 {
			continue
		}

		ip, err := s2ip(ss[0])
		if err != nil {
			continue
		}

		hw, err := hws2bs(ss[3])
		if err != nil {
			continue
		}

		dev := ss[5]

		r.arps[ip] = &ArpItem{
			Ip:     ip,
			Device: dev,
			HwAddr: hw,
		}

	}
	return nil
}

func (r *Arp) GetHwAddr(ip uint32) ([]byte, error) {
	if v, ok := r.arps[ip]; ok {
		return v.HwAddr, nil
	}
	return nil, fmt.Errorf("hw of ip not found")
}
