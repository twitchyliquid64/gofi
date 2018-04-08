package packet

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/golang/snappy"
)

// Inform captures the information contained in an inform packet.
// Actual data stored in the payload is represented as JSON and can be
// decoded with other methods in this package.
type Inform struct {
	Version uint32
	APMAC   [6]byte

	IV          []byte
	DataVersion uint32
	DataLength  uint32
	Data        []byte
	RawFlags    uint16

	Encrypted, CompressedSnappy, CompressedZib bool
}

// CloneForReply duplicates the struct into a new structure to be modified and transmitted.
func (i *Inform) CloneForReply() *Inform {
	r := &Inform{
		Version:          i.Version,
		APMAC:            i.APMAC,
		DataVersion:      i.DataVersion,
		DataLength:       i.DataLength,
		Encrypted:        i.Encrypted,
		CompressedSnappy: i.CompressedSnappy,
		CompressedZib:    i.CompressedZib,
		RawFlags:         i.RawFlags,
	}
	r.IV = make([]byte, len(i.IV))
	copy(r.IV, i.IV)
	r.Data = make([]byte, len(i.Data))
	copy(r.Data, i.Data)
	return r
}

// Marshal creates a 'on-the-wire' bitstream representing the contents of the Inform packet.
// NOTE: All flags and DataLength is ignored.
func (i *Inform) Marshal(key []byte) ([]byte, error) {
	var buff bytes.Buffer
	payload, err := encrypt(i.Data, key, i.IV)
	if err != nil {
		return nil, err
	}

	buff.Grow(len(payload) + 40)
	buff.WriteString("TNBU")
	err = binary.Write(&buff, binary.BigEndian, i.Version)
	if err != nil {
		return nil, err
	}

	buff.Write(i.APMAC[:])

	err = binary.Write(&buff, binary.BigEndian, uint16(1)) //encryption only
	if err != nil {
		return nil, err
	}

	buff.Write(i.IV[:])

	err = binary.Write(&buff, binary.BigEndian, i.DataVersion)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buff, binary.BigEndian, uint32(len(payload)))
	if err != nil {
		return nil, err
	}

	buff.Write(payload)

	return buff.Bytes(), nil
}

// performs PKCS7 padding and AES encryption on the given data.
func encrypt(data, key, iv []byte) ([]byte, error) {
	d, err := pkcs7Pad(data, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(d, d)
	return d, nil
}

// Payload decrypts and uncompresses using the given key, returning the raw payload.
func (i *Inform) Payload(key []byte) ([]byte, error) {
	if i.Encrypted {
		err := i.decrypt(key)
		if err != nil {
			return nil, err
		}
	}
	if i.CompressedSnappy {
		err := i.decompressSnappy()
		if err != nil {
			return nil, err
		}
	}
	if i.CompressedZib {
		err := i.decompressZlib()
		if err != nil {
			return nil, err
		}
	}
	//fmt.Println(string(i.Data))
	return i.Data, nil
}

// Called internally to reverse Zlib compression on i.Data.
func (i *Inform) decompressZlib() error {
	b := bytes.NewReader(i.Data)

	r, err := zlib.NewReader(b)
	if err != nil {
		return err
	}

	i.Data, err = ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	i.CompressedZib = false
	return nil
}

// Called internally to reverse Snappy compression on i.Data.
func (i *Inform) decompressSnappy() error {
	var err error
	i.Data, err = snappy.Decode(i.Data, i.Data)
	i.CompressedSnappy = false
	return err
}

func (i *Inform) decrypt(key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	mode := cipher.NewCBCDecrypter(block, i.IV)
	mode.CryptBlocks(i.Data, i.Data)
	d, err := pkcs7Unpad(i.Data, aes.BlockSize)
	if err != nil {
		return err
	}
	i.Data = d
	i.Encrypted = false
	return nil
}

// InformDecode decodes an inform packet.
// use .Payload() to extract the contents of the packet.
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

	if pktFlagsReadErr := binary.Read(r, binary.BigEndian, &pkt.RawFlags); pktFlagsReadErr != nil {
		return nil, pktFlagsReadErr
	}
	pkt.Encrypted = (pkt.RawFlags & 0x01) > 0
	pkt.CompressedZib = (pkt.RawFlags & 0x02) > 0
	pkt.CompressedSnappy = (pkt.RawFlags & 0x04) > 0

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
