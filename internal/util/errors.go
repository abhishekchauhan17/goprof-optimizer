package util

import "fmt"

// BadRequestError represents a client-side error (HTTP 400).
type BadRequestError struct {
	Msg string
}

func (e BadRequestError) Error() string {
	return e.Msg
}

// InternalError represents a server-side error (HTTP 500).
type InternalError struct {
	Msg string
}

func (e InternalError) Error() string {
	return e.Msg
}

// WrapInternal wraps a lower-level error into an InternalError with context.
func WrapInternal(msg string, err error) error {
	if err == nil {
		return nil
	}
	return InternalError{Msg: fmt.Sprintf("%s: %v", msg, err)}
}
