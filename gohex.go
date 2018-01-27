package gohex

import (
	"bufio"
	"encoding/hex"
	"io"
	"sort"
)

// Constants definitions of IntelHex record types
const (
	_DATA_RECORD    byte = 0 // Record with data bytes
	_EOF_RECORD     byte = 1 // Record with end of file indicator
	_ADDRESS_RECORD byte = 4 // Record with extended linear address
	_START_RECORD   byte = 5 // Record with start linear address
)

// Structure with binary data segment fields
type DataSegment struct {
	Address uint32 // Starting address of data segment
	Data    []byte // Data segment bytes
}

// Helper type for data segments sorting operations
type sortByAddress []*DataSegment

func (segs sortByAddress) Len() int           { return len(segs) }
func (segs sortByAddress) Swap(i, j int)      { segs[i], segs[j] = segs[j], segs[i] }
func (segs sortByAddress) Less(i, j int) bool { return segs[i].Address < segs[j].Address }

// Main structure with private fields of IntelHex parser
type Memory struct {
	dataSegments    []*DataSegment // Slice with pointers to DataSegments
	startAddress    uint32         // Start linear address
	extendedAddress uint32         // Extended linear address
	eofFlag         bool           // End of file record exist flag
	startFlag       bool           // Start address record exist flag
	lineNum         uint           // Parser input line number
}

// Constructor of Memory structure
func NewMemory() *Memory {
	m := new(Memory)
	m.Clear()
	return m
}

// Method to getting start address from IntelHex data
func (m *Memory) GetStartAddress() (adr uint32, ok bool) {
	if m.startFlag {
		return m.startAddress, true
	}
	return 0, false
}

// Method to setting start address to IntelHex data
func (m *Memory) SetStartAddress(adr uint32) {
	m.startAddress = adr
	m.startFlag = true
}

// Method to getting data segments address from IntelHex data
func (m *Memory) GetDataSegments() []DataSegment {
	segs := []DataSegment{}
	for _, s := range m.dataSegments {
		segs = append(segs, *s)
	}
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

func (m *Memory) AddBinary(adr uint32, bytes []byte) error {
	var segBefore *DataSegment = nil
	var segAfter *DataSegment = nil
	var segAfterIndex int
	for i, s := range m.dataSegments {
		if (adr >= s.Address) && (adr < s.Address+uint32(len(s.Data))) {
			return newParseError(_DATA_ERROR, "data segments overlap", m.lineNum)
		}
		if (adr < s.Address) && (adr+uint32(len(bytes)) > s.Address) {
			return newParseError(_DATA_ERROR, "data segments overlap", m.lineNum)
		}
		
		if adr == s.Address+uint32(len(s.Data)) {
			segBefore = s
		}
		if adr+uint32(len(bytes)) == s.Address {
			segAfter, segAfterIndex = s, i
		}
	}
	
	if segBefore != nil && segAfter != nil {
		segBefore.Data = append(segBefore.Data, bytes...)
		segBefore.Data = append(segBefore.Data, segAfter.Data...)
		m.dataSegments = append(m.dataSegments[:segAfterIndex], m.dataSegments[segAfterIndex+1:]...)

	} else if segBefore != nil && segAfter == nil {
		segBefore.Data = append(segBefore.Data, bytes...)
	} else if segBefore == nil && segAfter != nil {
		segAfter.Address = adr
		segAfter.Data = append(bytes, segAfter.Data...)
	} else {
		m.dataSegments = append(m.dataSegments, &DataSegment{Address: adr, Data: bytes})
	}
	sort.Sort(sortByAddress(m.dataSegments))
	return nil
}

func (m *Memory) parseIntelHexRecord(bytes []byte) error {
	if len(bytes) < 5 {
		return newParseError(_DATA_ERROR, "not enought data bytes", m.lineNum)
	}
	err := checkSum(bytes)
	if err != nil {
		return newParseError(_CHECKSUM_ERROR, err.Error(), m.lineNum)
	}
	err = checkRecordSize(bytes)
	if err != nil {
		return newParseError(_DATA_ERROR, err.Error(), m.lineNum)
	}
	switch record_type := bytes[3]; record_type {
	case _DATA_RECORD:
		a, data := getDataLine(bytes)
		adr := uint32(a) + m.extendedAddress
		err = m.AddBinary(adr, data)
		if err != nil {
			return err
		}
	case _EOF_RECORD:
		err = checkEOF(bytes)
		if err != nil {
			return newParseError(_RECORD_ERROR, err.Error(), m.lineNum)
		}
		m.eofFlag = true
	case _ADDRESS_RECORD:
		m.extendedAddress, err = getExtendedAddress(bytes)
		if err != nil {
			return newParseError(_RECORD_ERROR, err.Error(), m.lineNum)
		}
	case _START_RECORD:
		if m.startFlag == true {
			return newParseError(_DATA_ERROR, "multiple start address lines", m.lineNum)
		}
		m.startAddress, err = getStartAddress(bytes)
		if err != nil {
			return newParseError(_RECORD_ERROR, err.Error(), m.lineNum)
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
		return newParseError(_SYNTAX_ERROR, "no colon char on the first line character", m.lineNum)
	}
	bytes, err := hex.DecodeString(line[1:])
	if err != nil {
		return newParseError(_SYNTAX_ERROR, err.Error(), m.lineNum)
	}
	return m.parseIntelHexRecord(bytes)
}

func (m *Memory) ParseIntelHex(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
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
		return newParseError(_SYNTAX_ERROR, err.Error(), m.lineNum)
	}
	if m.eofFlag == false {
		return newParseError(_DATA_ERROR, "no end of file line", m.lineNum)
	}

	return nil
}

func (m *Memory) dumpDataSegment(writer io.Writer, s *DataSegment, lineLength byte) {
	lineAdr := s.Address
	lineData := []byte{}
	for byteAdr := s.Address; byteAdr < s.Address+uint32(len(s.Data)); byteAdr++ {
		if (byteAdr & 0xFFFF0000) != m.extendedAddress {
			if len(lineData) != 0 {
				writeDataLine(writer, &lineAdr, byteAdr, &lineData)
			}
			m.extendedAddress = (byteAdr & 0xFFFF0000)
			writeExtendedAddressLine(writer, m.extendedAddress)
		}
		if len(lineData) >= int(lineLength) {
			writeDataLine(writer, &lineAdr, byteAdr, &lineData)
		}
		lineData = append(lineData, s.Data[byteAdr-s.Address])
	}

	if len(lineData) != 0 {
		writeDataLine(writer, &lineAdr, 0, &lineData)
	}
}

func (m *Memory) DumpIntelHex(writer io.Writer, lineLength byte) {
	if m.startFlag {
		writeStartAddressLine(writer, m.startAddress)
	}

	m.extendedAddress = 0
	for _, s := range m.dataSegments {
		m.dumpDataSegment(writer, s, lineLength)
	}

	writeEofLine(writer)
}

func (m *Memory) ToBinary(address uint32, size uint32, padding byte) []byte {
	data := make([]byte, size)
	
	i := uint32(0)
	for i < size {
		ok := false
		for _, s := range m.dataSegments {
			if (address >= s.Address) && (address < s.Address + uint32(len(s.Data))) {
				data[i] = s.Data[address - s.Address]
				i++
				address++
				ok = true
				break
			}
		}
		if ok == false {
			data[i] = padding
			i++
			address++
		}
	}
	
	return data
}