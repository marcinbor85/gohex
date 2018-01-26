package gohex

import (
	"bufio"
	"encoding/hex"
	"strings"
	"sort"
)

const (
	DATA_RECORD    byte = 0
	EOF_RECORD     byte = 1
	ADDRESS_RECORD byte = 4
	START_RECORD   byte = 5
)

type DataSegment struct {
	data    []byte
	address int
}

type sortByAddress []*DataSegment

func (segs sortByAddress) Len() int           { return len(segs) }
func (segs sortByAddress) Swap(i, j int)      { segs[i], segs[j] = segs[j], segs[i] }
func (segs sortByAddress) Less(i, j int) bool { return segs[i].address < segs[j].address }

type Memory struct {
	dataSegments   []*DataSegment
	startAddress   int
	extendedAddress int
	eofFlag        bool
	startFlag      bool
	lineNum        int
}

func NewMemory() *Memory {
	m := new(Memory)
	m.Clear()
	return m
}

func (m *Memory) GetStartAddress() (int, bool) {
	if m.startFlag {
		return m.startAddress, true
	}
	return 0, false
}

func (m *Memory) GetDataSegments() []*DataSegment {
	segs := m.dataSegments
	sort.Sort(sortByAddress(segs))
	return segs
}

func (m *Memory) Clear() {
	m.startAddress = 0
	m.extendedAddress = 0
	m.lineNum = 0
	m.dataSegments = []*DataSegment{}
	m.startFlag = false
	m.eofFlag = false
}

func (m *Memory) AddBinary(adr int, bytes []byte) error {
	for _, s := range m.dataSegments {
		if ((adr >= s.address) && (adr < s.address+len(s.data))) ||
			((adr < s.address) && (adr+len(bytes) > s.address)) {
			return newParseError(DATA_ERROR, "data segments overlap", m.lineNum)
		}
		
		if adr == s.address+len(s.data) {
			s.data = append(s.data, bytes...)
			return nil
		}
		if adr+len(bytes) == s.address {
			s.address = adr
			s.data = append(bytes, s.data...)
			return nil
		}
	}
	m.dataSegments = append(m.dataSegments, &DataSegment{address: adr, data: bytes})
	return nil
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
	case DATA_RECORD:
		adr, data := getDataLine(bytes)
		adr += m.extendedAddress
		err = m.AddBinary(adr, data)
		if err != nil {
			return err
		}
	case EOF_RECORD:
		err = checkEOF(bytes)
		if err != nil {
			return newParseError(RECORD_ERROR, err.Error(), m.lineNum)
		}
		m.eofFlag = true
	case ADDRESS_RECORD:
		m.extendedAddress, err = getExtendedAddress(bytes)
		if err != nil {
			return newParseError(RECORD_ERROR, err.Error(), m.lineNum)
		}
	case START_RECORD:
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
