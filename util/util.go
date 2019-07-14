package util

import (
	"encoding/binary"
)

func Htons(h uint32) uint32 {
	b := []byte{0,0,0,0}
	binary.BigEndian.PutUint32(b, h)
	res := uint32(0)
	res = uint32(b[0]) + (uint32(b[1])>>8) + (uint32(b[1])>>16) + (uint32(b[1])>>24)
	return res
}