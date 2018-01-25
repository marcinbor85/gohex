package gohex

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"strings"
)

type DataSegment struct {
	data    []byte
	address int
}

type Memory struct {
	dataSegments []DataSegment
	startAddress int
}

func NewMemory() *Memory {
	m := new(Memory)
	return m
}

func (m *Memory) GetStartAddress() int {
	return m.startAddress
}

func (m *Memory) GetDataSegments() []DataSegment {
	return m.dataSegments
}

func (m *Memory) ParseIntelHex(str string) error {
	scanner := bufio.NewScanner(strings.NewReader(str))
	currentAddress := 0
	lineNum := 0
	eof := false
	start := false
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if line[0] != ':' {
			return newParseError(SYNTAX_ERROR, "no colon char on the first line character", lineNum)
		}
		bytes, err := hex.DecodeString(line[1:])
		if err != nil {
			return newParseError(SYNTAX_ERROR, err.Error(), lineNum)
		}
		if len(bytes) < 5 {
			return newParseError(DATA_ERROR, "not enought data bytes", lineNum)
		}

		err = checkSum(bytes)
		if err != nil {
			return newParseError(CHECKSUM_ERROR, err.Error(), lineNum)
		}

		err = checkRecordSize(bytes)
		if err != nil {
			return newParseError(DATA_ERROR, err.Error(), lineNum)
		}

		switch record_type := bytes[3]; record_type {
		case 0:
			//data
			fmt.Println(currentAddress)
			// jesli nie ma segmentu z aktualnym ciągłym adresem to:
			// utworz segment
			// wpisuj dane
			// zwieksz aktualny adres
		case 1:
			//eof
			err = checkEOF(bytes)
			if err != nil {
				return newParseError(RECORD_ERROR, err.Error(), lineNum)
			}
			eof = true
			break
		case 4:
			//extended address
			currentAddress, err = getExtendedAddress(bytes)
			if err != nil {
				return newParseError(RECORD_ERROR, err.Error(), lineNum)
			}
		case 5:
			//run address
			if start == true {
				return newParseError(DATA_ERROR, "multiple start address lines", lineNum)
			}
			m.startAddress, err = getStartAddress(bytes)
			if err != nil {
				return newParseError(RECORD_ERROR, err.Error(), lineNum)
			}
			start = true
		}
	}
	if err := scanner.Err(); err != nil {
		return newParseError(SYNTAX_ERROR, err.Error(), lineNum)
	}
	if eof == false {
		return newParseError(DATA_ERROR, "no end of file line", lineNum)
	}

	return nil
}

func (m *Memory) DumpIntelHex() error {
	return nil
}
