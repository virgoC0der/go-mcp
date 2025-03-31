package transport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/virgoC0der/go-mcp/types"
)

// WebSocketHandler handles WebSocket connections for the MCP server
type WebSocketHandler struct {
	server   Server
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mutex    sync.Mutex
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(server Server) *WebSocketHandler {
	return &WebSocketHandler{
		server: server,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

// ServeHTTP implements the http.Handler interface
func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upgrade to WebSocket: %v", err), http.StatusInternalServerError)
		return
	}

	// Register client
	h.mutex.Lock()
	h.clients[conn] = true
	h.mutex.Unlock()

	// Handle WebSocket connection
	go h.handleConnection(conn)
}

// handleConnection handles a WebSocket connection
func (h *WebSocketHandler) handleConnection(conn *websocket.Conn) {
	defer func() {
		// Unregister client
		h.mutex.Lock()
		delete(h.clients, conn)
		h.mutex.Unlock()

		// Close connection
		conn.Close()
	}()

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(10 * time.Minute))

	// Handle messages
	for {
		// Read message
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// Reset read deadline
		conn.SetReadDeadline(time.Now().Add(10 * time.Minute))

		// Handle message
		go h.handleMessage(conn, message)
	}
}

// handleMessage handles a WebSocket message
func (h *WebSocketHandler) handleMessage(conn *websocket.Conn, message []byte) {
	// Parse message
	var request map[string]interface{}
	err := json.Unmarshal(message, &request)
	if err != nil {
		h.sendErrorResponse(conn, "", "invalid_request", fmt.Sprintf("Failed to parse message: %v", err))
		return
	}

	// Extract request type and message ID
	requestType, ok := request["type"].(string)
	if !ok {
		h.sendErrorResponse(conn, "", "invalid_request", "Missing 'type' field")
		return
	}

	messageID, _ := request["messageId"].(string)

	// Handle request
	switch requestType {
	case "initialize":
		h.handleInitializeRequest(conn, messageID, request)
	case "listPrompts":
		h.handleListPromptsRequest(conn, messageID)
	case "getPrompt":
		h.handleGetPromptRequest(conn, messageID, request)
	case "listTools":
		h.handleListToolsRequest(conn, messageID)
	case "callTool":
		h.handleCallToolRequest(conn, messageID, request)
	case "listResources":
		h.handleListResourcesRequest(conn, messageID)
	case "readResource":
		h.handleReadResourceRequest(conn, messageID, request)
	default:
		h.sendErrorResponse(conn, messageID, "invalid_request", fmt.Sprintf("Unknown request type: %s", requestType))
	}
}

// handleInitializeRequest handles an initialize request
func (h *WebSocketHandler) handleInitializeRequest(conn *websocket.Conn, messageID string, request map[string]interface{}) {
	options, ok := request["data"]
	if !ok {
		h.sendErrorResponse(conn, messageID, "invalid_request", "Missing 'data' field")
		return
	}

	err := h.server.Initialize(context.Background(), options)
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			h.sendErrorResponse(conn, messageID, mcpErr.Code, mcpErr.Message)
		} else {
			h.sendErrorResponse(conn, messageID, "unknown_error", err.Error())
		}
		return
	}

	h.sendSuccessResponse(conn, messageID, nil)
}

// handleListPromptsRequest handles a list prompts request
func (h *WebSocketHandler) handleListPromptsRequest(conn *websocket.Conn, messageID string) {
	prompts, err := h.server.ListPrompts(context.Background())
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			h.sendErrorResponse(conn, messageID, mcpErr.Code, mcpErr.Message)
		} else {
			h.sendErrorResponse(conn, messageID, "unknown_error", err.Error())
		}
		return
	}

	h.sendSuccessResponse(conn, messageID, prompts)
}

// handleGetPromptRequest handles a get prompt request
func (h *WebSocketHandler) handleGetPromptRequest(conn *websocket.Conn, messageID string, request map[string]interface{}) {
	name, ok := request["name"].(string)
	if !ok {
		h.sendErrorResponse(conn, messageID, "invalid_request", "Missing 'name' field")
		return
	}

	args, _ := request["args"].(map[string]interface{})

	result, err := h.server.GetPrompt(context.Background(), name, args)
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			h.sendErrorResponse(conn, messageID, mcpErr.Code, mcpErr.Message)
		} else {
			h.sendErrorResponse(conn, messageID, "unknown_error", err.Error())
		}
		return
	}

	h.sendSuccessResponse(conn, messageID, result)
}

// handleListToolsRequest handles a list tools request
func (h *WebSocketHandler) handleListToolsRequest(conn *websocket.Conn, messageID string) {
	tools, err := h.server.ListTools(context.Background())
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			h.sendErrorResponse(conn, messageID, mcpErr.Code, mcpErr.Message)
		} else {
			h.sendErrorResponse(conn, messageID, "unknown_error", err.Error())
		}
		return
	}

	h.sendSuccessResponse(conn, messageID, tools)
}

// handleCallToolRequest handles a call tool request
func (h *WebSocketHandler) handleCallToolRequest(conn *websocket.Conn, messageID string, request map[string]interface{}) {
	name, ok := request["name"].(string)
	if !ok {
		h.sendErrorResponse(conn, messageID, "invalid_request", "Missing 'name' field")
		return
	}

	args, _ := request["args"].(map[string]interface{})

	result, err := h.server.CallTool(context.Background(), name, args)
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			h.sendErrorResponse(conn, messageID, mcpErr.Code, mcpErr.Message)
		} else {
			h.sendErrorResponse(conn, messageID, "unknown_error", err.Error())
		}
		return
	}

	h.sendSuccessResponse(conn, messageID, result)
}

// handleListResourcesRequest handles a list resources request
func (h *WebSocketHandler) handleListResourcesRequest(conn *websocket.Conn, messageID string) {
	resources, err := h.server.ListResources(context.Background())
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			h.sendErrorResponse(conn, messageID, mcpErr.Code, mcpErr.Message)
		} else {
			h.sendErrorResponse(conn, messageID, "unknown_error", err.Error())
		}
		return
	}

	h.sendSuccessResponse(conn, messageID, resources)
}

// handleReadResourceRequest handles a read resource request
func (h *WebSocketHandler) handleReadResourceRequest(conn *websocket.Conn, messageID string, request map[string]interface{}) {
	name, ok := request["name"].(string)
	if !ok {
		h.sendErrorResponse(conn, messageID, "invalid_request", "Missing 'name' field")
		return
	}

	content, mimeType, err := h.server.ReadResource(context.Background(), name)
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			h.sendErrorResponse(conn, messageID, mcpErr.Code, mcpErr.Message)
		} else {
			h.sendErrorResponse(conn, messageID, "unknown_error", err.Error())
		}
		return
	}

	// Base64 encode the content
	encodedContent := base64.StdEncoding.EncodeToString(content)

	h.sendSuccessResponse(conn, messageID, map[string]interface{}{
		"content":  encodedContent,
		"mimeType": mimeType,
	})
}

// sendSuccessResponse sends a success response
func (h *WebSocketHandler) sendSuccessResponse(conn *websocket.Conn, messageID string, result interface{}) {
	response := map[string]interface{}{
		"type":      "response",
		"messageId": messageID,
		"success":   true,
	}

	if result != nil {
		response["result"] = result
	}

	h.sendResponse(conn, response)
}

// sendErrorResponse sends an error response
func (h *WebSocketHandler) sendErrorResponse(conn *websocket.Conn, messageID, code, message string) {
	response := map[string]interface{}{
		"type":      "response",
		"messageId": messageID,
		"success":   false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}

	h.sendResponse(conn, response)
}

// requestCounter is used to generate unique message IDs
var requestCounter int64

// sendResponse sends a response
func (h *WebSocketHandler) sendResponse(conn *websocket.Conn, response map[string]interface{}) {
	// Generate message ID if not present
	if response["messageId"] == "" {
		response["messageId"] = fmt.Sprintf("%d", atomic.AddInt64(&requestCounter, 1))
	}

	// Send response
	conn.WriteJSON(response)
}
