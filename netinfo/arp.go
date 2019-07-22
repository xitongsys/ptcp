package netinfo

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

var ARPPATH = "/proc/net/arp"

type ArpItem struct {
	Ip     uint32
	Device string
	HwAddr []byte
}

type Arp struct {
	arps map[uint32]ArpItem
}

func NewArp() (*Arp, error) {
	r := &Arp{}
	err := r.Load(ARPPATH)
	return r, err
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

	r.arps = map[uint32]ArpItem{}

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		ss := strings.Fields(string(line))
		ip, dev, hw := s2ip(ss[0]), ss[5], hws2bs(ss[3])
		r.arps[ip] = ArpItem{
			Ip:     ip,
			Device: dev,
			HwAddr: hw,
		}

	}
	return nil
}

func (r *Arp) GetHwAddr(ip uint32) []byte {
	var res []byte = nil
	if v, ok := r.arps[ip]; ok {
		res = v.HwAddr
	}
	return res
}
