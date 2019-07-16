package header

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"net"
)

const (
	VLANNone = 0x000
	VLANMax  = 0xfff
)

type Priority uint8

const (
	PriorityBackground           Priority = 1
	PriorityBestEffort           Priority = 0
	PriorityExcellentEffort      Priority = 2
	PriorityCriticalApplications Priority = 3
	PriorityVideo                Priority = 4
	PriorityVoice                Priority = 5
	PriorityInternetworkControl  Priority = 6
	PriorityNetworkControl       Priority = 7
)

type VLAN struct {
	Priority     Priority
	DropEligible bool
	ID           uint16
}

func (v *VLAN) MarshalBinary() ([]byte, error) {
	b := make([]byte, 2)
	_, err := v.read(b)
	return b, err
}

func (v *VLAN) read(b []byte) (int, error) {
	if v.Priority > PriorityNetworkControl {
		return 0, fmt.Errorf("priority error")
	}

	if v.ID >= VLANMax {
		return 0, fmt.Errorf("id error")
	}

	ub := uint16(v.Priority) << 13

	var drop uint16
	if v.DropEligible {
		drop = 1
	}
	ub |= drop << 12
	ub |= v.ID
	binary.BigEndian.PutUint16(b, ub)
	return 2, nil
}

func (v *VLAN) UnmarshalBinary(b []byte) error {
	if len(b) != 2 {
		return io.ErrUnexpectedEOF
	}
	ub := binary.BigEndian.Uint16(b[0:2])
	v.Priority = Priority(uint8(ub >> 13))
	v.DropEligible = ub&0x1000 != 0
	v.ID = ub & 0x0fff

	if v.ID >= VLANMax {
		return fmt.Errorf("id > vlanmax")
	}

	return nil
}

const (
	minPayload = 46
)

var (
	Broadcast = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

type EtherType uint16

const (
	EtherTypeIPv4        EtherType = 0x0800
	EtherTypeARP         EtherType = 0x0806
	EtherTypeIPv6        EtherType = 0x86DD
	EtherTypeVLAN        EtherType = 0x8100
	EtherTypeServiceVLAN EtherType = 0x88a8
)

type Frame struct {
	Destination net.HardwareAddr
	Source      net.HardwareAddr
	ServiceVLAN *VLAN
	VLAN        *VLAN
	EtherType   EtherType
	Payload     []byte
}

func (f *Frame) MarshalBinary() ([]byte, error) {
	b := make([]byte, f.length())
	_, err := f.read(b)
	return b, err
}

func (f *Frame) MarshalFCS() ([]byte, error) {
	b := make([]byte, f.length()+4)
	if _, err := f.read(b); err != nil {
		return nil, err
	}
	binary.BigEndian.PutUint32(b[len(b)-4:], crc32.ChecksumIEEE(b[0:len(b)-4]))
	return b, nil
}

func (f *Frame) read(b []byte) (int, error) {
	if f.ServiceVLAN != nil && f.VLAN == nil {
		return 0, fmt.Errorf("vlan error")
	}

	copy(b[0:6], f.Destination)
	copy(b[6:12], f.Source)

	vlans := []struct {
		vlan *VLAN
		tpid EtherType
	}{
		{vlan: f.ServiceVLAN, tpid: EtherTypeServiceVLAN},
		{vlan: f.VLAN, tpid: EtherTypeVLAN},
	}

	n := 12
	for _, vt := range vlans {
		if vt.vlan == nil {
			continue
		}

		binary.BigEndian.PutUint16(b[n:n+2], uint16(vt.tpid))
		if _, err := vt.vlan.read(b[n+2 : n+4]); err != nil {
			return 0, err
		}
		n += 4
	}

	binary.BigEndian.PutUint16(b[n:n+2], uint16(f.EtherType))
	copy(b[n+2:], f.Payload)
	return len(b), nil
}

func (f *Frame) UnmarshalBinary(b []byte) error {
	if len(b) < 14 {
		return io.ErrUnexpectedEOF
	}

	n := 14
	et := EtherType(binary.BigEndian.Uint16(b[n-2 : n]))
	switch et {
	case EtherTypeServiceVLAN, EtherTypeVLAN:
		nn, err := f.unmarshalVLANs(et, b[n:])
		if err != nil {
			return err
		}

		n += nn
	default:
		f.EtherType = et
	}

	bb := make([]byte, 6+6+len(b[n:]))
	copy(bb[0:6], b[0:6])
	f.Destination = bb[0:6]
	copy(bb[6:12], b[6:12])
	f.Source = bb[6:12]

	copy(bb[12:], b[n:])
	f.Payload = bb[12:]

	return nil
}

func (f *Frame) UnmarshalFCS(b []byte) error {
	if len(b) < 4 {
		return io.ErrUnexpectedEOF
	}

	want := binary.BigEndian.Uint32(b[len(b)-4:])
	got := crc32.ChecksumIEEE(b[0 : len(b)-4])
	if want != got {
		return fmt.Errorf("fcs error")
	}

	return f.UnmarshalBinary(b[0 : len(b)-4])
}

func (f *Frame) length() int {
	pl := len(f.Payload)
	if pl < minPayload {
		pl = minPayload
	}

	var vlanLen int
	switch {
	case f.ServiceVLAN != nil && f.VLAN != nil:
		vlanLen = 8
	case f.VLAN != nil:
		vlanLen = 4
	}

	return 6 + 6 + vlanLen + 2 + pl
}

func (f *Frame) unmarshalVLANs(tpid EtherType, b []byte) (int, error) {
	if len(b) < 4 {
		return 0, io.ErrUnexpectedEOF
	}

	var n int

	switch tpid {
	case EtherTypeServiceVLAN:
		vlan := new(VLAN)
		if err := vlan.UnmarshalBinary(b[n : n+2]); err != nil {
			return 0, err
		}
		f.ServiceVLAN = vlan

		if EtherType(binary.BigEndian.Uint16(b[n+2:n+4])) != EtherTypeVLAN {
			return 0, fmt.Errorf("vlan error")
		}
		n += 4
		if len(b[n:]) < 4 {
			return 0, io.ErrUnexpectedEOF
		}
		fallthrough
	case EtherTypeVLAN:
		vlan := new(VLAN)
		if err := vlan.UnmarshalBinary(b[n : n+2]); err != nil {
			return 0, err
		}

		f.VLAN = vlan
		f.EtherType = EtherType(binary.BigEndian.Uint16(b[n+2 : n+4]))
		n += 4
	default:
		return -1, fmt.Errorf("unkown VLAN TPID %04x", tpid)
	}
	return n, nil
}
