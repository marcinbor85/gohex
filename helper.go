package gohex

import (
	"encoding/binary"
	"errors"
	"fmt"
)

func checkSum(bytes []byte) error {
	sum := 0
	for _, b := range bytes[:len(bytes)-1] {
		sum += int(b)
	}
	sum %= 256
	sum = 256 - sum
	last := int(bytes[len(bytes)-1])
	if sum != last {
		return errors.New("incorrect checksum (sum = " + fmt.Sprintf("%02X != %02X", sum, last) + ")")
	}
	return nil
}

func checkRecordSize(bytes []byte) error {
	if (int(bytes[0]) + 5) != len(bytes) {
		return errors.New("incorrect data length")
	}
	return nil
}

func checkEOF(bytes []byte) error {
	if bytes[0] != 0 {
		return errors.New("incorrect data length field in eof line")
	}
	if binary.BigEndian.Uint16(bytes[1:3]) != 0 {
		return errors.New("incorrect address field in eof line")
	}
	return nil
}

func getExtendedAddress(bytes []byte) (int, error) {
	if bytes[0] != 2 {
		return 0, errors.New("incorrect data length field in extended linear address line")
	}
	if binary.BigEndian.Uint16(bytes[1:3]) != 0 {
		return 0, errors.New("incorrect address field in extended linear address line")
	}
	adr := int(binary.BigEndian.Uint16(bytes[4:6])) << 16
	return adr, nil
}

func getDataLine(bytes []byte) (int, []byte) {
	size := bytes[0]
	adr := int(binary.BigEndian.Uint16(bytes[1:3]))
	data := bytes[4 : size+4]
	return adr, data
}

func getStartAddress(bytes []byte) (int, error) {
	if bytes[0] != 4 {
		return 0, errors.New("incorrect data length field in start address line")
	}
	if binary.BigEndian.Uint16(bytes[1:3]) != 0 {
		return 0, errors.New("incorrect address field in start address line")
	}
	adr := int(binary.BigEndian.Uint32(bytes[4:8]))
	return adr, nil
}
