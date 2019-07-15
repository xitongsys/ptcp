package header

func BuildUdpPacket(ipHeader *IPv4, udpHeader *UDP, data []byte) []byte {
	ipHeader.Len = uint16(20 + 8 + len(data))
	udpHeader.Len = uint16(8 + len(data))
	ipHeader.ResetChecksum()
	bs := []byte{}
	bs = append(bs, ipHeader.Marshal()...)
	bs = append(bs, udpHeader.Marshal()...)
	bs = append(bs, data...)
	ReCalUdpCheckSum(bs)
	return bs
} 

func BuildTcpPacket(ipHeader *IPv4, tcpHeader *TCP, data []byte) []byte {
	ipHeader.Len = uint16(20 + 20 + len(data))
	ipHeader.ResetChecksum()
	bs := []byte{}
	bs = append(bs, ipHeader.Marshal()...)
	bs = append(bs, tcpHeader.Marshal()...)
	bs = append(bs, data...)
	ReCalTcpCheckSum(bs)
	return bs
} 

func BuildTcpHeader(src, dst string) (*IPv4, *TCP) {
	srcIp, srcPort := ParseAddr(src)
	dstIp, dstPort := ParseAddr(dst)

	ipv4Header := &IPv4{
		VerIHL: 0x45,
		Tos: 0,
		Len: 0,
		Id: 0,
		Offset: 0,
		TTL: 255,
		Protocol: 0x06,
		Checksum: 0,
		Src: Str2IP(srcIp),
		Dst: Str2IP(dstIp),
	}
	ipv4Header.ResetChecksum()

	tcpHeader := &TCP{
		SrcPort: uint16(srcPort),
		DstPort: uint16(dstPort),
		Seq: 2,
		Ack: 3,
		Offset: 0x50,
		Flags: 0x02,
		Win: 29200,
		Checksum: 0,
		UrgPointer: 0,
	}
	return ipv4Header, tcpHeader
}

func BuildUdpHeader(src, dst string) (*IPv4, *UDP) {
	srcIp, srcPort := ParseAddr(src)
	dstIp, dstPort := ParseAddr(dst)

	ipv4Header := &IPv4{
		VerIHL: 0x45,
		Tos: 0,
		Len: 0,
		Id: 0,
		Offset: 0,
		TTL: 255,
		Protocol: 0x11,
		Checksum: 0,
		Src: Str2IP(srcIp),
		Dst: Str2IP(dstIp),
	}
	ipv4Header.ResetChecksum()

	udpHeader := &UDP{
		SrcPort: uint16(srcPort),
		DstPort: uint16(dstPort),
		Len: 0,
		Checksum: 0,
	}
	return ipv4Header, udpHeader
}
