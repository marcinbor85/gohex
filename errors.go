package gohex

import (
	"fmt"
)

type parseErrorType int

const (
	SYNTAX_ERROR   parseErrorType = 1
	RECORD_ERROR   parseErrorType = 2
	DATA_ERROR     parseErrorType = 3
	CHECKSUM_ERROR parseErrorType = 4
)

type parseError struct {
	errorType parseErrorType
	message   string
	lineNum   int
}

func (e *parseError) Error() string {
	var str string = "error"
	switch e.errorType {
	case SYNTAX_ERROR:
		str = "syntax error"
	case RECORD_ERROR:
		str = "record error"
	case DATA_ERROR:
		str = "data error"
	case CHECKSUM_ERROR:
		str = "checksum error"
	}
	return fmt.Sprintf("%s: %s at line %d", str, e.message, e.lineNum)
}

func newParseError(et parseErrorType, msg string, line int) error {
	return &parseError{errorType: et, message: msg, lineNum: line}
}
