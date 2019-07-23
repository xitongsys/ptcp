package netinfo

import (
	"fmt"
	"strconv"
	"strings"
)

//aa:bb:cc:dd:ee:ff -> [6]byte{}
func hws2bs(s string) ([]byte, error) {
	ss := strings.Split(s, ":")
	if len(ss) != 6 {
		return nil, fmt.Errorf("hw %v error", s)
	}
	res := make([]byte, len(ss))
	for i := 0; i < len(ss); i++ {
		b, err := strconv.ParseUint(ss[i], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("hw %v error", s)
		}
		res[i] = byte(b)
	}
	return res, nil
}

//bs -> uint32
func b2ip(bs []byte) (uint32, error) {
	if len(bs) < 4 {
		return 0, fmt.Errorf("bs %v error", bs)
	}
	return (uint32(bs[0]) << 24) + (uint32(bs[1]) << 16) + (uint32(bs[2]) << 8) + uint32(bs[3]), nil
}

//127.0.0.1 -> uint32
func s2ip(s string) (uint32, error) {
	ss := strings.Split(s, ".")
	if len(ss) != 4 {
		return 0, fmt.Errorf("ip %v error", s)
	}
	res := uint32(0)
	for i := 0; i < len(ss); i++ {
		n, err := strconv.ParseUint(ss[3-i], 10, 8)
		if err != nil {
			return 0, fmt.Errorf("ip %v error", s)
		}
		res += (uint32(n) << uint32(i*8))
	}
	return res, nil
}

func ip2s(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", (ip>>24)&(0xff), (ip>>16)&(0xff), (ip>>8)&(0xff), ip&(0xff))
}

//3A010000 -> uint32
func ips2ip(s string) uint32 {
	r, _ := strconv.ParseUint(s, 16, 32)
	return uint32(r)
}

//000013A -> uint32
func iprs2ip(s string) uint32 {
	s = s[6:8] + s[4:6] + s[2:4] + s[0:2]
	r, _ := strconv.ParseUint(s, 16, 32)
	return uint32(r)
}
