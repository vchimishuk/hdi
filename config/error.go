package config

import "fmt"

type parseError struct {
	line int
	msg  string
}

func newError(line int, msg string) parseError {
	return parseError{line: line, msg: msg}
}

func (e parseError) Error() string {
	return fmt.Sprintf("%d: %s", e.line, e.msg)
}
