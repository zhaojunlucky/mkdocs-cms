package core

import (
	"fmt"
)

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTPError: StatusCode=%d, Message=%s", e.StatusCode, e.Message)
}

func NewHTTPError(code int, message string) error {
	return &HTTPError{
		StatusCode: code,
		Message:    message,
	}
}
