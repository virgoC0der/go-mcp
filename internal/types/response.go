package types

// Response represents the standard response structure
type Response struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents the error information structure
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(result interface{}) Response {
	return Response{
		Success: true,
		Result:  result,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message string) Response {
	return Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
}

// NewMCPErrorResponse creates a new error response from an MCP error
func NewMCPErrorResponse(err *Error) Response {
	return Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    err.Code,
			Message: err.Message,
		},
	}
}
