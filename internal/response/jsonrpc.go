package response

import (
	"encoding/json"
	"fmt"

	"github.com/virgoC0der/go-mcp/internal/types"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewJSONRPCRequest creates a new JSON-RPC 2.0 request
func NewJSONRPCRequest(id interface{}, method string, params interface{}) JSONRPCRequest {
	var paramsJSON json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err == nil {
			paramsJSON = data
		}
	}

	return JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  paramsJSON,
	}
}

// NewJSONRPCResponse creates a new JSON-RPC 2.0 success response
func NewJSONRPCResponse(id interface{}, result interface{}) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// NewJSONRPCErrorResponse creates a new JSON-RPC 2.0 error response
func NewJSONRPCErrorResponse(id interface{}, code int, message string, data interface{}) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// MapMCPErrorToJSONRPCCode maps MCP error codes to JSON-RPC error codes
func MapMCPErrorToJSONRPCCode(mcpErrorCode string) int {
	switch mcpErrorCode {
	case "not_found":
		return -32002 // Resource not found
	case "invalid_request":
		return -32600 // Invalid request
	case "invalid_params":
		return -32602 // Invalid params
	default:
		return -32603 // Internal error
	}
}

// HandleJSONRPCError creates a JSON-RPC error response from an error
func HandleJSONRPCError(id interface{}, err error) JSONRPCResponse {
	if mcpErr, ok := err.(*types.Error); ok {
		// Map MCP error code to JSON-RPC error code
		code := MapMCPErrorToJSONRPCCode(mcpErr.Code)
		return NewJSONRPCErrorResponse(id, code, mcpErr.Message, mcpErr)
	}
	// Internal server error
	return NewJSONRPCErrorResponse(id, -32603, err.Error(), nil)
}

// MarshalResponse marshals a JSON-RPC response to JSON
func MarshalResponse(resp JSONRPCResponse) ([]byte, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		// If we can't marshal the response, create a simplified error response
		simpleResp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      resp.ID,
			"error": map[string]interface{}{
				"code":    -32603,
				"message": fmt.Sprintf("Failed to marshal response: %v", err),
			},
		}
		return json.Marshal(simpleResp)
	}
	return data, nil
}
