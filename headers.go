package gokeepasslib

import (
	"encoding/binary"
	"fmt"
	"io"
	"crypto/rand"
)

var AESCipherID = []byte{0x31, 0xC1, 0xF2, 0xE6, 0xBF, 0x71, 0x43, 0x50, 0xBE, 0x58, 0x05, 0x21, 0x6A, 0xFC, 0x5A, 0xFF}
var GzipCompressionFlag = uint32(1)
var SalsaInnerRandomStreamID = []byte{0x02,0x00,0x00,0x00}

// FileHeaders holds the header information of the Keepass File.
type FileHeaders struct {
	Comment             []byte // FieldID:  1
	CipherID            []byte // FieldID:  2
	CompressionFlags    uint32 // FieldID:  3
	MasterSeed          []byte // FieldID:  4
	TransformSeed       []byte // FieldID:  5
	TransformRounds     uint64 // FieldID:  6
	EncryptionIV        []byte // FieldID:  7
	ProtectedStreamKey  []byte // FieldID:  8
	StreamStartBytes    []byte // FieldID:  9
	InnerRandomStreamID []byte // FieldID: 10
}

//Creates a new FileHeaders with good defaults
func NewFileHeaders () (*FileHeaders) {
	h := new(FileHeaders)
	
	h.CipherID = []byte(AESCipherID)
	h.CompressionFlags = GzipCompressionFlag
	
	h.MasterSeed = make([]byte,32)
	rand.Read(h.MasterSeed)

	h.TransformSeed = make([]byte,32)
	rand.Read(h.TransformSeed)

	h.TransformRounds = 6000

	h.EncryptionIV = make([]byte,16)
	rand.Read(h.EncryptionIV)

	h.ProtectedStreamKey = make([]byte,32)
	rand.Read(h.ProtectedStreamKey)
	
	h.StreamStartBytes = make([]byte,32)
	rand.Read(h.StreamStartBytes)
	
	h.InnerRandomStreamID = SalsaInnerRandomStreamID
	
	return h
}

func (h FileHeaders) String() string {
	return fmt.Sprintf(
		"(1) Comment: %x\n"+
			"(2) CipherID: %x\n"+
			"(3) CompressionFlags: %x\n"+
			"(4) MasterSeed: %x\n"+
			"(5) TransformSeed: %x\n"+
			"(6) TransformRounds: %d\n"+
			"(7) EncryptionIV: %x\n"+
			"(8) ProtectedStreamKey: %x\n"+
			"(9) StreamStartBytes: %x\n"+
			"(10) InnerRandomStreamID: %x\n",
		h.Comment,
		h.CipherID,
		h.CompressionFlags,
		h.MasterSeed,
		h.TransformSeed,
		h.TransformRounds,
		h.EncryptionIV,
		h.ProtectedStreamKey,
		h.StreamStartBytes,
		h.InnerRandomStreamID,
	)
}

func ReadHeaders(r io.Reader) (*FileHeaders, error) {
	headers := new(FileHeaders)
	for {
		var fieldID byte
		if err := binary.Read(r, binary.LittleEndian, &fieldID); err != nil {
			return nil, err
		}

		var fieldLength [2]byte
		if err := binary.Read(r, binary.LittleEndian, &fieldLength); err != nil {
			return nil, err
		}

		var fieldData = make([]byte, binary.LittleEndian.Uint16(fieldLength[:]))
		if err := binary.Read(r, binary.LittleEndian, &fieldData); err != nil {
			return nil, err
		}

		switch fieldID {
		case 1:
			headers.Comment = fieldData
		case 2:
			headers.CipherID = fieldData
		case 3:
			headers.CompressionFlags = binary.LittleEndian.Uint32(fieldData)
		case 4:
			headers.MasterSeed = fieldData
		case 5:
			headers.TransformSeed = fieldData
		case 6:
			headers.TransformRounds = binary.LittleEndian.Uint64(fieldData)
		case 7:
			headers.EncryptionIV = fieldData
		case 8:
			headers.ProtectedStreamKey = fieldData
		case 9:
			headers.StreamStartBytes = fieldData
		case 10:
			headers.InnerRandomStreamID = fieldData
		}

		if fieldID == 0 {
			break
		}
	}

	return headers, nil
}

func (h *FileHeaders) WriteHeaders(w io.Writer) error {
	for i := 1; i <= 10; i++ {
		var data []byte
		switch i {
		case 1:
			data = append(data, h.Comment...)
		case 2:
			data = append(data, h.CipherID...)
		case 3:
			d := make([]byte, 4)
			binary.LittleEndian.PutUint32(d, h.CompressionFlags)
			data = append(data, d...)
		case 4:
			data = append(data, h.MasterSeed...)
		case 5:
			data = append(data, h.TransformSeed...)
		case 6:
			d := make([]byte, 8)
			binary.LittleEndian.PutUint64(d, h.TransformRounds)
			data = append(data, d...)
		case 7:
			data = append(data, h.EncryptionIV...)
		case 8:
			data = append(data, h.ProtectedStreamKey...)
		case 9:
			data = append(data, h.StreamStartBytes...)
		case 10:
			data = append(data, h.InnerRandomStreamID...)
		}

		if len(data) > 0 {
			err := binary.Write(w, binary.LittleEndian, uint8(i))
			if err != nil {
				return err
			}

			l := len(data)
			err = binary.Write(w, binary.LittleEndian, uint16(l))
			if err != nil {
				return err
			}

			err = binary.Write(w, binary.LittleEndian, data)
			if err != nil {
				return err
			}
		}
	}

	// End of header
	err := binary.Write(w, binary.LittleEndian, uint8(0))
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint16(4))
	if err != nil {
		return err
	}

	if _, err := w.Write([]byte{0x0d, 0x0a, 0x0d, 0x0a}); err != nil {
		return err
	}

	return nil
}
