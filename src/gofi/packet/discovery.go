package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

// Discovery TLV Types
const (
	MAC             uint8 = 1
	IPInfo          uint8 = 2
	FirmwareVersion uint8 = 3
	Uptime          uint8 = 0xA
	Hostname        uint8 = 0x0B
	Platform        uint8 = 0x0C
)

// Discovery represents the information in a discovery packet.
type Discovery struct {
	PktSize uint16
	RawTLVs []*TLV `json:"-"`

	MAC      [6]byte
	Hostname string
	IPInfo   net.Addr

	Platform        string
	FirmwareVersion string

	UptimeSecs uint32
}

func (d *Discovery) unpack() error {
	for _, tlv := range d.RawTLVs {
		//fmt.Printf("TLV %d is %x: [%d]%s\n", i, tlv.Kind, tlv.Length, tlv.Payload)
		switch tlv.Kind {
		case MAC:
			if tlv.Length != 6 {
				return errors.New("Invalid MAC payload length")
			}
			copy(d.MAC[:], tlv.Payload[:6])
		case Uptime:
			if err := binary.Read(bytes.NewBuffer(tlv.Payload), binary.BigEndian, &d.UptimeSecs); err != nil {
				return err
			}
		case Hostname:
			d.Hostname = string(tlv.Payload)
		case FirmwareVersion:
			d.FirmwareVersion = string(tlv.Payload)
		case Platform:
			d.Platform = string(tlv.Payload)
		}
	}
	return nil
}

// Debug prints the contents of the packet to console.
func (d *Discovery) Debug() {
	fmt.Printf("Discovery packet of length=%d\n", d.PktSize)
	fmt.Printf("\tMAC=%x\n", d.MAC)
	fmt.Printf("\tHostname=%s\n", d.Hostname)
	fmt.Printf("\tFirmware=%s\n", d.FirmwareVersion)
	fmt.Printf("\tPlatform=%s\n", d.Platform)
	fmt.Printf("\tUptime=%d\n", d.UptimeSecs)
	fmt.Printf("\tAddr=%+v\n", d.IPInfo)
}

// DiscoveryDecode decodes a discovery packet from a ubiquiti device.
func DiscoveryDecode(addr *net.UDPAddr, pkt []byte) (*Discovery, error) {
	r := bytes.NewBuffer(pkt)
	out := Discovery{IPInfo: addr}

	magic := make([]byte, 2)
	n, err := io.ReadFull(r, magic)
	if n != 2 || err != nil {
		return nil, errors.New("Could not read magic header")
	}
	if magic[0] != 2 || magic[1] != 6 {
		return nil, errors.New("Incorrect header")
	}

	if pktSizeErr := binary.Read(r, binary.BigEndian, &out.PktSize); pktSizeErr != nil {
		return nil, pktSizeErr
	}

	var tlv *TLV
	for err == nil {
		tlv, err = decodeTLV(r)
		if err == nil {
			out.RawTLVs = append(out.RawTLVs, tlv)
		} else if err != io.EOF {
			return nil, err
		}
	}

	unpackErr := out.unpack()
	if unpackErr != nil {
		return nil, unpackErr
	}
	return &out, nil
}

// TLV is a decoded type-length-value block.
type TLV struct {
	Kind    uint8
	Length  uint16
	Payload []byte
}

func decodeTLV(r io.Reader) (*TLV, error) {
	out := TLV{}
	if err := binary.Read(r, binary.BigEndian, &out.Kind); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &out.Length); err != nil {
		return nil, err
	}
	out.Payload = make([]byte, out.Length)
	n, err := io.ReadFull(r, out.Payload)
	if err != nil {
		return nil, err
	}
	out.Payload = out.Payload[:n]
	return &out, nil
}
