package gohex

import (
	"reflect"
	"strings"
	"testing"
)

func TestConstructor(t *testing.T) {
	m := NewMemory()
	if a, ok := m.GetStartAddress(); ok != false || a != 0 {
		t.Error("incorrect initial start Address")
	}
	if len(m.GetDataSegments()) != 0 {
		t.Error("incorrect initial data segments")
	}
	if m.extendedAddress != 0 {
		t.Error("incorrect initial data segments")
	}
}

func parseIntelHex(m *Memory, str string) error {
	return m.ParseIntelHex(strings.NewReader(str))
}

func assertParseError(t *testing.T, m *Memory, input string, et parseErrorType, err string) {
	if e := parseIntelHex(m, input); e != nil {
		perr, ok := e.(*parseError)
		if ok == true {
			if perr.errorType != et {
				t.Error(perr.Error())
				t.Error(err)
			}
		} else {
			t.Error(err)
		}
	} else {
		t.Error(err)
	}
}

func TestSyntaxError(t *testing.T) {
	m := NewMemory()
	assertParseError(t, m, "00000001FF\n", _SYNTAX_ERROR, "no colon error")
	assertParseError(t, m, ":qw00000001FF\n", _SYNTAX_ERROR, "no ascii hex error")
	assertParseError(t, m, ":0000001FF\n", _SYNTAX_ERROR, "no odd/even hex error")
}

func TestDataError(t *testing.T) {
	m := NewMemory()
	assertParseError(t, m, ":000000FF\n", _DATA_ERROR, "no line length error")
	assertParseError(t, m, ":02000000FE\n", _DATA_ERROR, "no data length error")
	assertParseError(t, m, "\n", _DATA_ERROR, "no end of file line error")
	assertParseError(t, m, ":000000FF01\n", _DATA_ERROR, "no end of file line error")
	assertParseError(t, m, ":0400000501000000F6\n", _DATA_ERROR, "no end of file line error")
	assertParseError(t, m, ":0400000501000000F6\n:0400000502000000F5\n:00000001FF\n", _DATA_ERROR, "no multiple start Address lines error")
	assertParseError(t, m, ":048000000102030472\n:04800300050607085F\n:00000001FF\n", _DATA_ERROR, "no segments overlap error")
	assertParseError(t, m, ":048000000102030472\n:047FFD000506070866\n:00000001FF\n", _DATA_ERROR, "no segments overlap error")
}

func TestChecksumError(t *testing.T) {
	m := NewMemory()
	assertParseError(t, m, ":00000101FF\n", _CHECKSUM_ERROR, "no checksum error")
	assertParseError(t, m, ":00000001FE\n", _CHECKSUM_ERROR, "no checksum error")
	assertParseError(t, m, ":0000000001\n", _CHECKSUM_ERROR, "no checksum error")
	assertParseError(t, m, ":000000FF02\n", _CHECKSUM_ERROR, "no checksum error")
}

func TestRecordsError(t *testing.T) {
	m := NewMemory()
	assertParseError(t, m, ":00000101FE\n", _RECORD_ERROR, "no eof record error")
	assertParseError(t, m, ":00010001FE\n", _RECORD_ERROR, "no eof record error")
	assertParseError(t, m, ":0100000100FE\n", _RECORD_ERROR, "no eof record error")
	assertParseError(t, m, ":020001040101F7\n", _RECORD_ERROR, "no extended Address record error")
	assertParseError(t, m, ":020100040101F7\n", _RECORD_ERROR, "no extended Address record error")
	assertParseError(t, m, ":03000004010100F7\n", _RECORD_ERROR, "no extended Address record error")
	assertParseError(t, m, ":0400010501010101F2\n", _RECORD_ERROR, "no start Address record error")
	assertParseError(t, m, ":0401000501010101F2\n", _RECORD_ERROR, "no start Address record error")
	assertParseError(t, m, ":050000050101010100F2\n", _RECORD_ERROR, "no start Address record error")
}

func TestAddress(t *testing.T) {
	m := NewMemory()
	err := parseIntelHex(m, ":020000041234B4\n:0400000501020304ED\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if m.lineNum != 3 {
		t.Error("incorrect lines number")
	}
	if m.extendedAddress != 0x12340000 {
		t.Errorf("incorrect extended Address: %08X", m.extendedAddress)
	}
	if a, ok := m.GetStartAddress(); a != 0x01020304 && ok == true {
		t.Errorf("incorrect start Address: %08X", m.startAddress)
	}
	if len(m.GetDataSegments()) != 0 {
		t.Error("incorrect data segments")
	}
	if m.eofFlag != true {
		t.Error("incorrect eof flag state")
	}
	if m.startFlag != true {
		t.Error("incorrect start flag state")
	}
	err = parseIntelHex(m, ":020000049ABCA4\n:0400000591929394AD\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if m.extendedAddress != 0x9ABC0000 {
		t.Errorf("incorrect extended Address: %08X", m.extendedAddress)
	}
	if a, ok := m.GetStartAddress(); a != 0x91929394 && ok == true {
		t.Errorf("incorrect start Address: %08X", m.startAddress)
	}

	err = parseIntelHex(m, ":020000041234B4\n:02000004234592\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if m.extendedAddress != 0x23450000 {
		t.Errorf("incorrect extended Address: %08X", m.extendedAddress)
	}
}

func TestDataSegments(t *testing.T) {
	m := NewMemory()
	err := parseIntelHex(m, ":048000000102030472\n:04800400050607085E\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 1 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg := m.GetDataSegments()[0]
	p := DataSegment{Address: 0x8000, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}

	err = parseIntelHex(m, ":048000000102030472\n:047FFC000506070867\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 1 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{Address: 0x7FFC, Data: []byte{5, 6, 7, 8, 1, 2, 3, 4}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}

	err = parseIntelHex(m, ":048000000102030472\n:04800800050607085A\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 2 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{Address: 0x8000, Data: []byte{1, 2, 3, 4}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}
	seg = m.GetDataSegments()[1]
	p = DataSegment{Address: 0x8008, Data: []byte{5, 6, 7, 8}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}

	err = parseIntelHex(m, ":04800800050607085A\n:048000000102030472\n\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 2 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{Address: 0x8000, Data: []byte{1, 2, 3, 4}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}
	seg = m.GetDataSegments()[1]
	p = DataSegment{Address: 0x8008, Data: []byte{5, 6, 7, 8}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}

	err = parseIntelHex(m, ":020000041000EA\n:048000000102030472\n:04800800050607085A\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 2 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{Address: 0x10008000, Data: []byte{1, 2, 3, 4}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}
	seg = m.GetDataSegments()[1]
	p = DataSegment{Address: 0x10008008, Data: []byte{5, 6, 7, 8}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}

	err = parseIntelHex(m, ":020000042000DA\n:048000000506070862\n:020000041000EA\n:048000000102030472\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 2 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{Address: 0x10008000, Data: []byte{1, 2, 3, 4}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}
	seg = m.GetDataSegments()[1]
	p = DataSegment{Address: 0x20008000, Data: []byte{5, 6, 7, 8}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}

}

func TestClear(t *testing.T) {
	m := NewMemory()
	err := parseIntelHex(m, ":020000049ABCA4\n:0400000591929394AD\n:048000000102030472\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	m.Clear()
	if m.lineNum != 0 {
		t.Error("incorrect lines number")
	}
	if len(m.GetDataSegments()) != 0 {
		t.Error("incorrect data segments")
	}
	if m.extendedAddress != 0 {
		t.Errorf("incorrect extended Address: %08X", m.extendedAddress)
	}
	if a, _ := m.GetStartAddress(); a != 0 {
		t.Errorf("incorrect start Address: %08X", m.extendedAddress)
	}
	if m.eofFlag != false {
		t.Error("incorrect eof flag state")
	}
	if m.startFlag != false {
		t.Error("incorrect start flag state")
	}
}

func TestAddBinary(t *testing.T) {
	m := NewMemory()
	err := m.AddBinary(0x20000, []byte{1, 2, 3, 4})
	err = m.AddBinary(0x20004, []byte{5, 6, 7, 8})
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 1 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg := m.GetDataSegments()[0]
	p := DataSegment{Address: 0x20000, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}

	err = m.AddBinary(0x10000, []byte{1, 2, 3, 4})
	err = m.AddBinary(0xFFFC, []byte{5, 6, 7, 8})
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 2 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{Address: 0xFFFC, Data: []byte{5, 6, 7, 8, 1, 2, 3, 4}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}
	seg = m.GetDataSegments()[1]
	p = DataSegment{Address: 0x20000, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}

	err = m.AddBinary(0x15000, []byte{1, 2, 3, 4})
	err = m.AddBinary(0x14FF8, []byte{5, 6, 7, 8, 9, 10, 11, 12})
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 3 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{Address: 0xFFFC, Data: []byte{5, 6, 7, 8, 1, 2, 3, 4}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}
	seg = m.GetDataSegments()[1]
	p = DataSegment{Address: 0x14FF8, Data: []byte{5, 6, 7, 8, 9, 10, 11, 12, 1, 2, 3, 4}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}
	seg = m.GetDataSegments()[2]
	p = DataSegment{Address: 0x20000, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8}}
	if reflect.DeepEqual(seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", seg, p)
	}
}
