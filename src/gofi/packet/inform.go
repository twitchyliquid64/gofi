package packet

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"io"
	"strconv"
)

// Inform captures the information contained in an inform packet.
type Inform struct {
	Version uint32
	APMAC   [6]byte

	IV          []byte
	DataVersion uint32
	DataLength  uint32
	Data        []byte

	Encrypted, Compressed bool
}

// Payload decrypts and uncompresses using the given key, returning the raw payload.
func (i *Inform) Payload(key []byte) ([]byte, error) {
	if i.Compressed {
		err := i.decompress()
		if err != nil {
			return nil, err
		}
	}
	if i.Encrypted {
		err := i.decrypt(key)
		if err != nil {
			return nil, err
		}
	}
	return i.Data, nil
}

func (i *Inform) decompress() error {
	b := bytes.NewReader(i.Data)
	r, err := zlib.NewReader(b)
	if err != nil {
		return err
	}

	o := bytes.NewBuffer(make([]byte, i.DataLength/2))
	io.Copy(o, r)
	i.Compressed = false
	i.Data = o.Bytes()
	return nil
}

func (i *Inform) decrypt(key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	mode := cipher.NewCBCDecrypter(block, i.IV)
	mode.CryptBlocks(i.Data, i.Data)
	i.Encrypted = false
	return nil
}

// InformDecode decodes an inform packet.
func InformDecode(r io.Reader) (*Inform, error) {
	pkt := &Inform{}

	magic := make([]byte, 4)
	n, err := io.ReadFull(r, magic)
	if n != 4 || err != nil || string(magic) != "TNBU" {
		return nil, errors.New("Could not read magic header, got " + string(magic))
	}

	if pktVersionReadErr := binary.Read(r, binary.BigEndian, &pkt.Version); pktVersionReadErr != nil {
		return nil, pktVersionReadErr
	}
	if pkt.Version != 0 {
		return nil, errors.New("Unsupported protocol version: " + strconv.Itoa(int(pkt.Version)))
	}

	mac := make([]byte, 6)
	_, err = io.ReadFull(r, mac)
	if err != nil {
		return nil, err
	}
	copy(pkt.APMAC[:], mac)

	flags := make([]byte, 2)
	_, err = io.ReadFull(r, flags)
	if err != nil {
		return nil, err
	}
	//fmt.Println(flags)
	pkt.Encrypted = (flags[1] & (1 << 0)) > 0
	pkt.Compressed = (flags[1] & (1 << 1)) > 0

	pkt.IV = make([]byte, 16)
	_, err = io.ReadFull(r, pkt.IV)
	if err != nil {
		return nil, err
	}

	if pktDataVersionReadErr := binary.Read(r, binary.BigEndian, &pkt.DataVersion); pktDataVersionReadErr != nil {
		return nil, pktDataVersionReadErr
	}
	if pkt.DataVersion != 1 {
		return nil, errors.New("Unsupported data version: " + strconv.Itoa(int(pkt.DataVersion)))
	}

	if pktDataLengthReadErr := binary.Read(r, binary.BigEndian, &pkt.DataLength); pktDataLengthReadErr != nil {
		return nil, pktDataLengthReadErr
	}
	pkt.Data = make([]byte, pkt.DataLength)
	_, err = io.ReadFull(r, pkt.Data)
	if err != nil {
		return nil, err
	}

	return pkt, nil
}
