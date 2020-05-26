package ring

import (
	"fmt"
	"regexp"
)

const (
	CodeNotFound string = "301001"
	CodeNotReady string = "301002"
	CodeWrongID  string = "301003"
)

var codeReg = regexp.MustCompile("^\\[\\d{6}\\]")

// ChordError represents the internal functional error of Chord
type ChordError struct {
	Code    string
	Msg     string
}

// Error wraps error interface
func(e ChordError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Msg)
}

// Equals determines whether the given error is this error type
func(e ChordError) Equals(err error) bool {
	if codeReg.FindString(err.Error()) == fmt.Sprintf("[%s]", e.Code) {
		return true
	}
	return false
}

func NewChordError(code string, msg string) ChordError {
	return ChordError{
		Code: code,
		Msg:  msg,
	}
}

var (
	// ErrNotFound is the shared object of NotFoundError
	ErrNotFound = NewChordError(CodeNotFound, "node not found")
	// ErrNotReady is the shared object of NotReadyError
	ErrNotReady = NewChordError(CodeNotReady, "node not ready")
	// ErrWrongID is the shared object of WrongIDError
	ErrWrongID = NewChordError(CodeWrongID, "wrong/invalid id")
)
