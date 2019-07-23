package netinfo

import (
	"fmt"
	"strconv"
	"strings"
)

//aa:bb:cc:dd:ee:ff -> [6]byte{}
func hws2bs(s string) []byte {
	ss := strings.Split(s, ":")
	res := make([]byte, len(ss))
	for i := 0; i < len(ss); i++ {
		b, _ := strconv.ParseUint(ss[i], 16, 8)
		res[i] = byte(b)
	}
	return res
}

//bs -> uint32
func b2ip(bs []byte) uint32 {
	if len(bs) < 4 {
		return 0
	}
	return (uint32(bs[0]) << 24) + (uint32(bs[1]) << 16) + (uint32(bs[2]) << 8) + uint32(bs[3])
}

//127.0.0.1 -> uint32
func s2ip(s string) uint32 {
	ss := strings.Split(s, ".")
	res := uint32(0)
	for i := 0; i < len(ss); i++ {
		n, _ := strconv.ParseUint(ss[3-i], 10, 8)
		res += (uint32(n) << uint32(i*8))
	}
	return res
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
