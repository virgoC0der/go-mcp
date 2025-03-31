package types

import (
	"fmt"
)

// Error represents an MCP error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewError creates a new error with the given code and message
func NewError(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}