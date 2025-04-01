package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/virgoC0der/go-mcp/server"
)

// WSServer wraps the WebSocket transport layer for an MCP server
type WSServer struct {
	server   server.Server
	addr     string
	upgrader websocket.Upgrader
	// clients field is deprecated and no longer used
	handler *WebSocketHandler
	srv     *http.Server
}

// Message represents the structure of a WebSocket message
type Message struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Method    string          `json:"method"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// Response represents the structure of a WebSocket response
type Response struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// NewWSServer creates a new WebSocket server instance
func NewWSServer(mcpServer server.Server, addr string) *WSServer {
	return &WSServer{
		server: mcpServer,
		addr:   addr,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Should be more strict in production
			},
		},
		handler: NewWebSocketHandler(mcpServer),
	}
}

// Start starts the WebSocket server
func (s *WSServer) Start() error {
	mux := http.NewServeMux()
	mux.Handle("/ws", s.handler)
	s.srv = &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the WebSocket server
func (s *WSServer) Shutdown(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}

// Deprecated: Legacy handler kept for backwards compatibility
func (s *WSServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Create a context for each connection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle messages
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		go s.handleMessage(ctx, conn, msg)
	}
}

func (s *WSServer) handleMessage(ctx context.Context, conn *websocket.Conn, msg Message) {
	var response Response
	response.Type = "response"
	response.ID = msg.ID

	switch msg.Method {
	case "listPrompts":
		prompts, err := s.server.ListPrompts(ctx)
		if err != nil {
			response.Error = err.Error()
		} else {
			response.Result = prompts
		}

	case "getPrompt":
		var args struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(msg.Arguments, &args); err != nil {
			response.Error = err.Error()
			break
		}

		result, err := s.server.GetPrompt(ctx, args.Name, args.Arguments)
		if err != nil {
			response.Error = err.Error()
		} else {
			response.Result = result
		}

	case "listTools":
		tools, err := s.server.ListTools(ctx)
		if err != nil {
			response.Error = err.Error()
		} else {
			response.Result = tools
		}

	case "callTool":
		var args struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(msg.Arguments, &args); err != nil {
			response.Error = err.Error()
			break
		}

		result, err := s.server.CallTool(ctx, args.Name, args.Arguments)
		if err != nil {
			response.Error = err.Error()
		} else {
			response.Result = result
		}

	case "listResources":
		resources, err := s.server.ListResources(ctx)
		if err != nil {
			response.Error = err.Error()
		} else {
			response.Result = resources
		}

	case "readResource":
		var args struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(msg.Arguments, &args); err != nil {
			response.Error = err.Error()
			break
		}

		content, mimeType, err := s.server.ReadResource(ctx, args.Name)
		if err != nil {
			response.Error = err.Error()
		} else {
			response.Result = map[string]any{
				"content":  content,
				"mimeType": mimeType,
			}
		}

	default:
		response.Error = fmt.Sprintf("unknown method: %s", msg.Method)
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}
}
