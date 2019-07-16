package header

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net"
)

const (
	// VLANNone is a special VLAN ID which indicates that no VLAN is being
	// used in a Frame.  In this case, the VLAN's other fields may be used
	// to indicate a Frame's priority.
	VLANNone = 0x000

	// VLANMax is a reserved VLAN ID which may indicate a wildcard in some
	// management systems, but may not be configured or transmitted in a
	// VLAN tag.
	VLANMax = 0xfff
)

var (
	// ErrInvalidVLAN is returned when a VLAN tag is invalid due to one of the
	// following reasons:
	//   - Priority of greater than 7 is detected
	//   - ID of greater than 4094 (0xffe) is detected
	//   - A customer VLAN does not follow a service VLAN (when using Q-in-Q)
	ErrInvalidVLAN = errors.New("invalid VLAN")
)

// Priority is an IEEE P802.1p priority level.  Priority can be any value from
// 0 to 7.
//
// It is important to note that priority 1 (PriorityBackground) actually has
// a lower priority than 0 (PriorityBestEffort).  All other Priority constants
// indicate higher priority as the integer values increase.
type Priority uint8

// IEEE P802.1p recommended priority levels.  Note that PriorityBackground has
// a lower priority than PriorityBestEffort.
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

// A VLAN is an IEEE 802.1Q Virtual LAN (VLAN) tag.  A VLAN contains
// information regarding traffic priority and a VLAN identifier for
// a given Frame.
type VLAN struct {
	// Priority specifies a IEEE P802.1p priority level.  Priority can be any
	// value from 0 to 7.
	Priority Priority

	// DropEligible indicates if a Frame is eligible to be dropped in the
	// presence of network congestion.
	DropEligible bool

	// ID specifies the VLAN ID for a Frame.  ID can be any value from 0 to
	// 4094 (0x000 to 0xffe), allowing up to 4094 VLANs.
	//
	// If ID is 0 (0x000, VLANNone), no VLAN is specified, and the other fields
	// simply indicate a Frame's priority.
	ID uint16
}

// MarshalBinary allocates a byte slice and marshals a VLAN into binary form.
func (v *VLAN) MarshalBinary() ([]byte, error) {
	b := make([]byte, 2)
	_, err := v.read(b)
	return b, err
}

// read reads data from a VLAN into b.  read is used to marshal a VLAN into
// binary form, but does not allocate on its own.
func (v *VLAN) read(b []byte) (int, error) {
	// Check for VLAN priority in valid range
	if v.Priority > PriorityNetworkControl {
		return 0, ErrInvalidVLAN
	}

	// Check for VLAN ID in valid range
	if v.ID >= VLANMax {
		return 0, ErrInvalidVLAN
	}

	// 3 bits: priority
	ub := uint16(v.Priority) << 13

	// 1 bit: drop eligible
	var drop uint16
	if v.DropEligible {
		drop = 1
	}
	ub |= drop << 12

	// 12 bits: VLAN ID
	ub |= v.ID

	binary.BigEndian.PutUint16(b, ub)
	return 2, nil
}

// UnmarshalBinary unmarshals a byte slice into a VLAN.
func (v *VLAN) UnmarshalBinary(b []byte) error {
	// VLAN tag is always 2 bytes
	if len(b) != 2 {
		return io.ErrUnexpectedEOF
	}

	//  3 bits: priority
	//  1 bit : drop eligible
	// 12 bits: VLAN ID
	ub := binary.BigEndian.Uint16(b[0:2])
	v.Priority = Priority(uint8(ub >> 13))
	v.DropEligible = ub&0x1000 != 0
	v.ID = ub & 0x0fff

	// Check for VLAN ID in valid range
	if v.ID >= VLANMax {
		return ErrInvalidVLAN
	}

	return nil
}

//go:generate stringer -output=string.go -type=EtherType

const (
	// minPayload is the minimum payload size for an Ethernet frame, assuming
	// that no 802.1Q VLAN tags are present.
	minPayload = 46
)

var (
	// Broadcast is a special hardware address which indicates a Frame should
	// be sent to every device on a given LAN segment.
	Broadcast = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

var (
	// ErrInvalidFCS is returned when Frame.UnmarshalFCS detects an incorrect
	// Ethernet frame check sequence in a byte slice for a Frame.
	ErrInvalidFCS = errors.New("invalid frame check sequence")
)

// An EtherType is a value used to identify an upper layer protocol
// encapsulated in a Frame.
//
// A list of IANA-assigned EtherType values may be found here:
// http://www.iana.org/assignments/ieee-802-numbers/ieee-802-numbers.xhtml.
type EtherType uint16

// Common EtherType values frequently used in a Frame.
const (
	EtherTypeIPv4 EtherType = 0x0800
	EtherTypeARP  EtherType = 0x0806
	EtherTypeIPv6 EtherType = 0x86DD

	// EtherTypeVLAN and EtherTypeServiceVLAN are used as 802.1Q Tag Protocol
	// Identifiers (TPIDs).
	EtherTypeVLAN        EtherType = 0x8100
	EtherTypeServiceVLAN EtherType = 0x88a8
)

// A Frame is an IEEE 802.3 Ethernet II frame.  A Frame contains information
// such as source and destination hardware addresses, zero or more optional
// 802.1Q VLAN tags, an EtherType, and payload data.
type Frame struct {
	// Destination specifies the destination hardware address for this Frame.
	//
	// If this address is set to Broadcast, the Frame will be sent to every
	// device on a given LAN segment.
	Destination net.HardwareAddr

	// Source specifies the source hardware address for this Frame.
	//
	// Typically, this is the hardware address of the network interface used to
	// send this Frame.
	Source net.HardwareAddr

	// ServiceVLAN specifies an optional 802.1Q service VLAN tag, for use with
	// 802.1ad double tagging, or "Q-in-Q". If ServiceVLAN is not nil, VLAN must
	// not be nil as well.
	//
	// Most users should leave this field set to nil and use VLAN instead.
	ServiceVLAN *VLAN

	// VLAN specifies an optional 802.1Q customer VLAN tag, which may or may
	// not be present in a Frame.  It is important to note that the operating
	// system may automatically strip VLAN tags before they can be parsed.
	VLAN *VLAN

	// EtherType is a value used to identify an upper layer protocol
	// encapsulated in this Frame.
	EtherType EtherType

	// Payload is a variable length data payload encapsulated by this Frame.
	Payload []byte
}

// MarshalBinary allocates a byte slice and marshals a Frame into binary form.
func (f *Frame) MarshalBinary() ([]byte, error) {
	b := make([]byte, f.length())
	_, err := f.read(b)
	return b, err
}

// MarshalFCS allocates a byte slice, marshals a Frame into binary form, and
// finally calculates and places a 4-byte IEEE CRC32 frame check sequence at
// the end of the slice.
//
// Most users should use MarshalBinary instead.  MarshalFCS is provided as a
// convenience for rare occasions when the operating system cannot
// automatically generate a frame check sequence for an Ethernet frame.
func (f *Frame) MarshalFCS() ([]byte, error) {
	// Frame length with 4 extra bytes for frame check sequence
	b := make([]byte, f.length()+4)
	if _, err := f.read(b); err != nil {
		return nil, err
	}

	// Compute IEEE CRC32 checksum of frame bytes and place it directly
	// in the last four bytes of the slice
	binary.BigEndian.PutUint32(b[len(b)-4:], crc32.ChecksumIEEE(b[0:len(b)-4]))
	return b, nil
}

// read reads data from a Frame into b.  read is used to marshal a Frame
// into binary form, but does not allocate on its own.
func (f *Frame) read(b []byte) (int, error) {
	// S-VLAN must also have accompanying C-VLAN.
	if f.ServiceVLAN != nil && f.VLAN == nil {
		return 0, ErrInvalidVLAN
	}

	copy(b[0:6], f.Destination)
	copy(b[6:12], f.Source)

	// Marshal each non-nil VLAN tag into bytes, inserting the appropriate
	// EtherType/TPID before each, so devices know that one or more VLANs
	// are present.
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

		// Add VLAN EtherType and VLAN bytes.
		binary.BigEndian.PutUint16(b[n:n+2], uint16(vt.tpid))
		if _, err := vt.vlan.read(b[n+2 : n+4]); err != nil {
			return 0, err
		}
		n += 4
	}

	// Marshal actual EtherType after any VLANs, copy payload into
	// output bytes.
	binary.BigEndian.PutUint16(b[n:n+2], uint16(f.EtherType))
	copy(b[n+2:], f.Payload)

	return len(b), nil
}

// UnmarshalBinary unmarshals a byte slice into a Frame.
func (f *Frame) UnmarshalBinary(b []byte) error {
	// Verify that both hardware addresses and a single EtherType are present
	if len(b) < 14 {
		return io.ErrUnexpectedEOF
	}

	// Track offset in packet for reading data
	n := 14

	// Continue looping and parsing VLAN tags until no more VLAN EtherType
	// values are detected
	et := EtherType(binary.BigEndian.Uint16(b[n-2 : n]))
	switch et {
	case EtherTypeServiceVLAN, EtherTypeVLAN:
		// VLAN type is hinted for further parsing.  An index is returned which
		// indicates how many bytes were consumed by VLAN tags.
		nn, err := f.unmarshalVLANs(et, b[n:])
		if err != nil {
			return err
		}

		n += nn
	default:
		// No VLANs detected.
		f.EtherType = et
	}

	// Allocate single byte slice to store destination and source hardware
	// addresses, and payload
	bb := make([]byte, 6+6+len(b[n:]))
	copy(bb[0:6], b[0:6])
	f.Destination = bb[0:6]
	copy(bb[6:12], b[6:12])
	f.Source = bb[6:12]

	// There used to be a minimum payload length restriction here, but as
	// long as two hardware addresses and an EtherType are present, it
	// doesn't really matter what is contained in the payload.  We will
	// follow the "robustness principle".
	copy(bb[12:], b[n:])
	f.Payload = bb[12:]

	return nil
}

// UnmarshalFCS computes the IEEE CRC32 frame check sequence of a Frame,
// verifies it against the checksum present in the byte slice, and finally,
// unmarshals a byte slice into a Frame.
//
// Most users should use UnmarshalBinary instead.  UnmarshalFCS is provided as
// a convenience for rare occasions when the operating system cannot
// automatically verify a frame check sequence for an Ethernet frame.
func (f *Frame) UnmarshalFCS(b []byte) error {
	// Must contain enough data for FCS, to avoid panics
	if len(b) < 4 {
		return io.ErrUnexpectedEOF
	}

	// Verify checksum in slice versus newly computed checksum
	want := binary.BigEndian.Uint32(b[len(b)-4:])
	got := crc32.ChecksumIEEE(b[0 : len(b)-4])
	if want != got {
		return ErrInvalidFCS
	}

	return f.UnmarshalBinary(b[0 : len(b)-4])
}

// length calculates the number of bytes required to store a Frame.
func (f *Frame) length() int {
	// If payload is less than the required minimum length, we zero-pad up to
	// the required minimum length
	pl := len(f.Payload)
	if pl < minPayload {
		pl = minPayload
	}

	// Add additional length if VLAN tags are needed.
	var vlanLen int
	switch {
	case f.ServiceVLAN != nil && f.VLAN != nil:
		vlanLen = 8
	case f.VLAN != nil:
		vlanLen = 4
	}

	// 6 bytes: destination hardware address
	// 6 bytes: source hardware address
	// N bytes: VLAN tags (if present)
	// 2 bytes: EtherType
	// N bytes: payload length (may be padded)
	return 6 + 6 + vlanLen + 2 + pl
}

// unmarshalVLANs unmarshals S/C-VLAN tags.  It is assumed that tpid
// is a valid S/C-VLAN TPID.
func (f *Frame) unmarshalVLANs(tpid EtherType, b []byte) (int, error) {
	// 4 or more bytes must remain for valid S/C-VLAN tag and EtherType.
	if len(b) < 4 {
		return 0, io.ErrUnexpectedEOF
	}

	// Track how many bytes are consumed by VLAN tags.
	var n int

	switch tpid {
	case EtherTypeServiceVLAN:
		vlan := new(VLAN)
		if err := vlan.UnmarshalBinary(b[n : n+2]); err != nil {
			return 0, err
		}
		f.ServiceVLAN = vlan

		// Assume that a C-VLAN immediately trails an S-VLAN.
		if EtherType(binary.BigEndian.Uint16(b[n+2:n+4])) != EtherTypeVLAN {
			return 0, ErrInvalidVLAN
		}

		// 4 or more bytes must remain for valid C-VLAN tag and EtherType.
		n += 4
		if len(b[n:]) < 4 {
			return 0, io.ErrUnexpectedEOF
		}

		// Continue to parse the C-VLAN.
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
		panic(fmt.Sprintf("unknown VLAN TPID: %04x", tpid))
	}

	return n, nil
}
