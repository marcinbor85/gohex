package gohex

import (
	"fmt"
)

type ParseErrorType int

const (
	SYNTAX_ERROR   ParseErrorType = 1
	RECORD_ERROR   ParseErrorType = 2
	DATA_ERROR     ParseErrorType = 3
	CHECKSUM_ERROR ParseErrorType = 4
)

type ParseError struct {
	ErrorType ParseErrorType
	Message   string
	LineNum   int
}

func (e *ParseError) Error() string {
	var str string = "error"
	switch e.ErrorType {
	case SYNTAX_ERROR:
		str = "syntax error"
	case RECORD_ERROR:
		str = "record error"
	case DATA_ERROR:
		str = "data error"
	case CHECKSUM_ERROR:
		str = "checksum error"
	}
	return fmt.Sprintf("%s: %s at line %d", str, e.Message, e.LineNum)
}

func newParseError(et ParseErrorType, msg string, line int) error {
	return &ParseError{ErrorType: et, Message: msg, LineNum: line}
}
