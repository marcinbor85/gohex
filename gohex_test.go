package gohex

import (
	"testing"
)

func TestConstructor(t *testing.T) {
	m := NewMemory()
	if m.GetStartAddress() != 0 {
		t.Errorf("wrong initial start address")
	}
	if len(m.GetDataSegments()) != 0 {
		t.Errorf("wrong initial data segments")
	}
}

func assertParseError(t *testing.T, m *Memory, input string, et ParseErrorType, err string) {
	if e := m.ParseIntelHex(input); e != nil {
		perr, ok := e.(*ParseError)
		if ok == true {
			if perr.ErrorType != et {
				t.Errorf(perr.Error())
				t.Errorf(err)
			}
		} else {
			t.Errorf(err)
		}
	} else {
		t.Errorf(err)
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
}

func TestChecksumError(t *testing.T) {
	m := NewMemory()
	assertParseError(t, m, ":00000101FF\n", CHECKSUM_ERROR, "no checking checksum error")
	assertParseError(t, m, ":00000001FE\n", CHECKSUM_ERROR, "no checking checksum error")
	assertParseError(t, m, ":0000000001\n", CHECKSUM_ERROR, "no checking checksum error")
	assertParseError(t, m, ":000000FF02\n", CHECKSUM_ERROR, "no checking checksum error")
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
