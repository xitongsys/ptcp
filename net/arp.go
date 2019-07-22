package net

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

var ARPPATH = "/proc/net/arp"

func hws2bs(s string) []byte {
	ss := strings.Split(s, ":")
	res := make([]byte, len(ss))
	for i := 0; i < len(ss); i++ {
		b, _ := strconv.ParseUint(ss[i], 16, 8)
		res[i] = byte(b)
	}
	return res
}

func s2ip(s string) uint32 {
	ss := strings.Split(s, ".")
	res := uint32(0)
	for i := 0; i < len(ss); i++ {
		n, _ := strconv.ParseUint(ss[3-i], 10, 8)
		res += (uint32(n) << uint32(i*8))
	}
	return res
}

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
