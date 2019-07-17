package util

func Htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}
