package gohex

import (
	"bufio"
	"encoding/hex"
	"strings"
)

type DataSegment struct {
	data    []byte
	address int
}

type Memory struct {
	dataSegments   []DataSegment
	startAddress   int
	currentAddress int
	eofFlag		   bool
	startFlag	   bool
	lineNum		   int
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

func (m *Memory) Clear() {
	m.startAddress = 0
	m.currentAddress = 0
	m.lineNum = 0
	m.dataSegments = []DataSegment{}
	m.startFlag = false
	m.eofFlag = false
}

func (m *Memory) parseIntelHexRecord(bytes []byte) error {
	if len(bytes) < 5 {
		return newParseError(DATA_ERROR, "not enought data bytes", m.lineNum)
	}
	err := checkSum(bytes)
	if err != nil {
		return newParseError(CHECKSUM_ERROR, err.Error(), m.lineNum)
	}
	err = checkRecordSize(bytes)
	if err != nil {
		return newParseError(DATA_ERROR, err.Error(), m.lineNum)
	}
	switch record_type := bytes[3]; record_type {
	case 0:
		//data
		// jesli nie ma segmentu z aktualnym ciągłym adresem to:
		// utworz segment
		// wpisuj dane
		// zwieksz aktualny adres
	case 1:
		//eof
		err = checkEOF(bytes)
		if err != nil {
			return newParseError(RECORD_ERROR, err.Error(), m.lineNum)
		}
		m.eofFlag = true
		break
	case 4:
		//extended address
		m.currentAddress, err = getExtendedAddress(bytes)
		if err != nil {
			return newParseError(RECORD_ERROR, err.Error(), m.lineNum)
		}
	case 5:
		//run address
		if m.startFlag == true {
			return newParseError(DATA_ERROR, "multiple start address lines", m.lineNum)
		}
		m.startAddress, err = getStartAddress(bytes)
		if err != nil {
			return newParseError(RECORD_ERROR, err.Error(), m.lineNum)
		}
		m.startFlag = true
	}
	return nil
}

func (m *Memory) parseIntelHexLine(line string) error {
	if len(line) == 0 {
		return nil
	}
	if line[0] != ':' {
		return newParseError(SYNTAX_ERROR, "no colon char on the first line character", m.lineNum)
	}
	bytes, err := hex.DecodeString(line[1:])
	if err != nil {
		return newParseError(SYNTAX_ERROR, err.Error(), m.lineNum)
	}
	return m.parseIntelHexRecord(bytes)
}

func (m *Memory) ParseIntelHex(str string) error {
	scanner := bufio.NewScanner(strings.NewReader(str))
	m.Clear()
	for scanner.Scan() {
		m.lineNum++
		line := scanner.Text()
		err := m.parseIntelHexLine(line)
		if err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return newParseError(SYNTAX_ERROR, err.Error(), m.lineNum)
	}
	if m.eofFlag == false {
		return newParseError(DATA_ERROR, "no end of file line", m.lineNum)
	}

	return nil
}

func (m *Memory) DumpIntelHex() error {
	return nil
}
