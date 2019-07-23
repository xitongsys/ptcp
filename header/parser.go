package header

import (
	"encoding/binary"
	"fmt"
)

func GetIp(data []byte) (src uint32, dst uint32, err error) {
	if len(data) < 20 {
		return 0, 0, fmt.Errorf("packet too short")
	}

	src, dst, err = binary.BigEndian.Uint32(data[12:16]), binary.BigEndian.Uint32(data[16:20]), nil
	return
}

func GetBase(data []byte) (proto string, src string, dst string, err error) {
	iph, udph, tcph := IPv4{}, UDP{}, TCP{}
	if len(data) < 20 {
		err = fmt.Errorf("Packet too short")
		return
	}

	if err = iph.Unmarshal(data); err != nil {
		return
	}

	if iph.Protocol == uint8(UDPID) {
		proto = "udp"
		if err = udph.Unmarshal(GetSubSlice(data, int(iph.HeaderLen()), len(data))); err != nil {
			return
		}
		src = fmt.Sprintf("%s:%d", IP2Str(iph.Src), udph.SrcPort)
		dst = fmt.Sprintf("%s:%d", IP2Str(iph.Dst), udph.DstPort)

	} else if iph.Protocol == uint8(TCPID) {
		proto = "tcp"
		if err = tcph.Unmarshal(GetSubSlice(data, int(iph.HeaderLen()), len(data))); err != nil {
			return
		}
		src = fmt.Sprintf("%s:%d", IP2Str(iph.Src), tcph.SrcPort)
		dst = fmt.Sprintf("%s:%d", IP2Str(iph.Dst), tcph.DstPort)

	} else {
		err = fmt.Errorf("Protocol Unsupported: id=%d", iph.Protocol)
	}
	return

}

func Get(data []byte) (proto string, iph *IPv4, udph *UDP, tcph *TCP, packetData []byte, err error) {
	proto = ""
	iph, udph, tcph = &IPv4{}, &UDP{}, &TCP{}
	if len(data) < 20 {
		err = fmt.Errorf("Packet too short")
		return
	}

	if err = iph.Unmarshal(data); err != nil {
		return
	}

	if iph.Protocol == uint8(UDPID) {
		proto = "udp"
		if err = udph.Unmarshal(GetSubSlice(data, int(iph.HeaderLen()), len(data))); err != nil {
			return
		}
		packetData = GetSubSlice(data, int(iph.HeaderLen()+8), int(iph.HeaderLen()+udph.LenBytes()))

	} else if iph.Protocol == uint8(TCPID) {
		proto = "tcp"
		if err = tcph.Unmarshal(GetSubSlice(data, int(iph.HeaderLen()), len(data))); err != nil {
			return
		}
		packetData = GetSubSlice(data, int(iph.HeaderLen()+tcph.HeaderLen()), int(iph.LenBytes()))

	} else {
		err = fmt.Errorf("Protocol Unsupported: id=%d", iph.Protocol)
	}
	return
}
