package gohex

import (
	"testing"
	"reflect"
)

func TestConstructor(t *testing.T) {
	m := NewMemory()
	if m.GetStartAddress() != 0 {
		t.Error("incorrect initial start address")
	}
	if len(m.GetDataSegments()) != 0 {
		t.Error("incorrect initial data segments")
	}
	if m.extendedAddress != 0 {
		t.Error("incorrect initial data segments")
	}
}

func assertParseError(t *testing.T, m *Memory, input string, et ParseErrorType, err string) {
	if e := m.ParseIntelHex(input); e != nil {
		perr, ok := e.(*ParseError)
		if ok == true {
			if perr.ErrorType != et {
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
	assertParseError(t, m, "00000001FF\n", SYNTAX_ERROR, "no colon error")
	assertParseError(t, m, ":qw00000001FF\n", SYNTAX_ERROR, "no ascii hex error")
	assertParseError(t, m, ":0000001FF\n", SYNTAX_ERROR, "no odd/even hex error")
}

func TestDataError(t *testing.T) {
	m := NewMemory()
	assertParseError(t, m, ":000000FF\n", DATA_ERROR, "no line length error")
	assertParseError(t, m, ":02000000FE\n", DATA_ERROR, "no data length error")
	assertParseError(t, m, "\n", DATA_ERROR, "no end of file line error")
	assertParseError(t, m, ":000000FF01\n", DATA_ERROR, "no end of file line error")
	assertParseError(t, m, ":0400000501000000F6\n", DATA_ERROR, "no end of file line error")
	assertParseError(t, m, ":0400000501000000F6\n:0400000502000000F5\n:00000001FF\n", DATA_ERROR, "no multiple start address lines error")
	assertParseError(t, m, ":048000000102030472\n:04800300050607085F\n:00000001FF\n", DATA_ERROR, "no segments overlap error")
	assertParseError(t, m, ":048000000102030472\n:047FFD000506070866\n:00000001FF\n", DATA_ERROR, "no segments overlap error")
}

func TestChecksumError(t *testing.T) {
	m := NewMemory()
	assertParseError(t, m, ":00000101FF\n", CHECKSUM_ERROR, "no checksum error")
	assertParseError(t, m, ":00000001FE\n", CHECKSUM_ERROR, "no checksum error")
	assertParseError(t, m, ":0000000001\n", CHECKSUM_ERROR, "no checksum error")
	assertParseError(t, m, ":000000FF02\n", CHECKSUM_ERROR, "no checksum error")
}

func TestRecordsError(t *testing.T) {
	m := NewMemory()
	assertParseError(t, m, ":00000101FE\n", RECORD_ERROR, "no eof record error")
	assertParseError(t, m, ":00010001FE\n", RECORD_ERROR, "no eof record error")
	assertParseError(t, m, ":0100000100FE\n", RECORD_ERROR, "no eof record error")
	assertParseError(t, m, ":020001040101F7\n", RECORD_ERROR, "no extended address record error")
	assertParseError(t, m, ":020100040101F7\n", RECORD_ERROR, "no extended address record error")
	assertParseError(t, m, ":03000004010100F7\n", RECORD_ERROR, "no extended address record error")
	assertParseError(t, m, ":0400010501010101F2\n", RECORD_ERROR, "no start address record error")
	assertParseError(t, m, ":0401000501010101F2\n", RECORD_ERROR, "no start address record error")
	assertParseError(t, m, ":050000050101010100F2\n", RECORD_ERROR, "no start address record error")
}

func TestAddress(t *testing.T) {
	m := NewMemory()
	err := m.ParseIntelHex(":020000041234B4\n:0400000501020304ED\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if m.lineNum != 3 {
		t.Error("incorrect lines number")
	}
	if m.extendedAddress != 0x12340000 {
		t.Errorf("incorrect extended address: %08X", m.extendedAddress)
	}
	if m.startAddress != 0x01020304 {
		t.Errorf("incorrect start address: %08X", m.startAddress)
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
	err = m.ParseIntelHex(":020000049ABCA4\n:0400000591929394AD\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if m.extendedAddress != 0x9ABC0000 {
		t.Errorf("incorrect extended address: %08X", m.extendedAddress)
	}
	if m.startAddress != 0x91929394 {
		t.Errorf("incorrect start address: %08X", m.startAddress)
	}
	m.Clear()
	if m.lineNum != 0 {
		t.Error("incorrect lines number")
	}
	if len(m.GetDataSegments()) != 0 {
		t.Error("incorrect data segments")
	}
	if m.extendedAddress != 0 {
		t.Errorf("incorrect extended address: %08X", m.extendedAddress)
	}
	if m.startAddress != 0 {
		t.Errorf("incorrect start address: %08X", m.extendedAddress)
	}
	if m.eofFlag != false {
		t.Error("incorrect eof flag state")
	}
	if m.startFlag != false {
		t.Error("incorrect start flag state")
	}
	err = m.ParseIntelHex(":020000041234B4\n:02000004234592\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if m.extendedAddress != 0x23450000 {
		t.Errorf("incorrect extended address: %08X", m.extendedAddress)
	}
}

func TestDataSegments(t *testing.T) {
	m := NewMemory()
	err := m.ParseIntelHex(":048000000102030472\n:04800400050607085E\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 1 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg := m.GetDataSegments()[0]
	p := DataSegment{address: 0x8000, data: []byte{1,2,3,4,5,6,7,8}}
	if reflect.DeepEqual(*seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", *seg, p)
	}
	
	err = m.ParseIntelHex(":048000000102030472\n:047FFC000506070867\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 1 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{address: 0x7FFC, data: []byte{5,6,7,8,1,2,3,4}}
	if reflect.DeepEqual(*seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", *seg, p)
	}
	
	err = m.ParseIntelHex(":048000000102030472\n:04800800050607085A\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 2 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{address: 0x8000, data: []byte{1,2,3,4}}
	if reflect.DeepEqual(*seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", *seg, p)
	}
	seg = m.GetDataSegments()[1]
	p = DataSegment{address: 0x8008, data: []byte{5,6,7,8}}
	if reflect.DeepEqual(*seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", *seg, p)
	}
	
	err = m.ParseIntelHex(":04800800050607085A\n:048000000102030472\n\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 2 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{address: 0x8008, data: []byte{5,6,7,8}}
	if reflect.DeepEqual(*seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", *seg, p)
	}
	seg = m.GetDataSegments()[1]
	p = DataSegment{address: 0x8000, data: []byte{1,2,3,4}}
	if reflect.DeepEqual(*seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", *seg, p)
	}
	
	err = m.ParseIntelHex(":020000041000EA\n:048000000102030472\n:04800800050607085A\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 2 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{address: 0x10008000, data: []byte{1,2,3,4}}
	if reflect.DeepEqual(*seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", *seg, p)
	}
	seg = m.GetDataSegments()[1]
	p = DataSegment{address: 0x10008008, data: []byte{5,6,7,8}}
	if reflect.DeepEqual(*seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", *seg, p)
	}
	
	err = m.ParseIntelHex(":020000041000EA\n:048000000102030472\n:020000042000DA\n:048000000506070862\n:00000001FF\n")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
	if len(m.GetDataSegments()) != 2 {
		t.Errorf("incorrect number of data segments: %v", len(m.GetDataSegments()))
	}
	seg = m.GetDataSegments()[0]
	p = DataSegment{address: 0x10008000, data: []byte{1,2,3,4}}
	if reflect.DeepEqual(*seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", *seg, p)
	}
	seg = m.GetDataSegments()[1]
	p = DataSegment{address: 0x20008000, data: []byte{5,6,7,8}}
	if reflect.DeepEqual(*seg, p) == false {
		t.Errorf("incorrect segment: %v != %v", *seg, p)
	}

}

