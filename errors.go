package gohex

import (
	"fmt"
)

type parseErrorType int

const (
	_SYNTAX_ERROR   parseErrorType = 1
	_RECORD_ERROR   parseErrorType = 2
	_DATA_ERROR     parseErrorType = 3
	_CHECKSUM_ERROR parseErrorType = 4
)

type parseError struct {
	errorType parseErrorType
	message   string
	lineNum   int
}

func (e *parseError) Error() string {
	var str string = "error"
	switch e.errorType {
	case _SYNTAX_ERROR:
		str = "syntax error"
	case _RECORD_ERROR:
		str = "record error"
	case _DATA_ERROR:
		str = "data error"
	case _CHECKSUM_ERROR:
		str = "checksum error"
	}
	return fmt.Sprintf("%s: %s at line %d", str, e.message, e.lineNum)
}

func newParseError(et parseErrorType, msg string, line int) error {
	return &parseError{errorType: et, message: msg, lineNum: line}
}
