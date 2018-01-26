package gohex

import (
	"bufio"
	"encoding/hex"
	"sort"
	"strings"
)

// Constants definitions of IntelHex record types
const (
	DATA_RECORD    byte = 0 // Record with data bytes
	EOF_RECORD     byte = 1 // Record with end of file indicator
	ADDRESS_RECORD byte = 4 // Record with extended linear address
	START_RECORD   byte = 5 // Record with start linear address
)

// Structure with binary data segment fields
type DataSegment struct {
	Address int // Starting address of data segment
	Data    []byte // Data segment bytes
}

// Helper type for data segments sorting operations
type sortByAddress []*DataSegment

func (segs sortByAddress) Len() int           { return len(segs) }
func (segs sortByAddress) Swap(i, j int)      { segs[i], segs[j] = segs[j], segs[i] }
func (segs sortByAddress) Less(i, j int) bool { return segs[i].Address < segs[j].Address }

type Memory struct {
	dataSegments    []*DataSegment
	startAddress    int
	extendedAddress int
	eofFlag         bool
	startFlag       bool
	lineNum         int
}

func NewMemory() *Memory {
	m := new(Memory)
	m.Clear()
	return m
}

// Method to retrieve start address from IntelHex data
func (m *Memory) GetStartAddress() (adr int, ok bool) {
	if m.startFlag {
		return m.startAddress, true
	}
	return 0, false
}

// Method to retrieve data segments address from IntelHex data
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
		if ((adr >= s.Address) && (adr < s.Address+len(s.Data))) ||
			((adr < s.Address) && (adr+len(bytes) > s.Address)) {
			return newParseError(DATA_ERROR, "data segments overlap", m.lineNum)
		}

		if adr == s.Address+len(s.Data) {
			s.Data = append(s.Data, bytes...)
			return nil
		}
		if adr+len(bytes) == s.Address {
			s.Address = adr
			s.Data = append(bytes, s.Data...)
			return nil
		}
	}
	m.dataSegments = append(m.dataSegments, &DataSegment{Address: adr, Data: bytes})
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
