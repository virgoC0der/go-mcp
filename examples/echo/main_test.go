package main

import (
	"context"
	"testing"
)

func TestEchoServer_GetPrompt(t *testing.T) {
	server := NewEchoServer()

	tests := []struct {
		name       string
		promptName string
		args       map[string]interface{}
		wantErr    bool
	}{
		{
			name:       "valid echo prompt",
			promptName: "echo",
			args:       map[string]interface{}{"message": "hello"},
			wantErr:    false,
		},
		{
			name:       "unknown prompt",
			promptName: "unknown",
			args:       map[string]interface{}{"message": "hello"},
			wantErr:    true,
		},
		{
			name:       "missing message",
			promptName: "echo",
			args:       map[string]interface{}{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.GetPrompt(context.Background(), tt.promptName, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("EchoServer.GetPrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("EchoServer.GetPrompt() returned nil result when no error was expected")
			}
		})
	}
}

func TestEchoServer_CallTool(t *testing.T) {
	server := NewEchoServer()

	tests := []struct {
		name     string
		toolName string
		args     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "valid echo tool",
			toolName: "echo",
			args:     map[string]interface{}{"message": "hello"},
			wantErr:  false,
		},
		{
			name:     "unknown tool",
			toolName: "unknown",
			args:     map[string]interface{}{"message": "hello"},
			wantErr:  true,
		},
		{
			name:     "missing message",
			toolName: "echo",
			args:     map[string]interface{}{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.CallTool(context.Background(), tt.toolName, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("EchoServer.CallTool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("EchoServer.CallTool() returned nil result when no error was expected")
			}
		})
	}
}

func TestEchoServer_ReadResource(t *testing.T) {
	server := NewEchoServer()

	tests := []struct {
		name         string
		resourceName string
		wantErr      bool
	}{
		{
			name:         "valid echo resource",
			resourceName: "echo",
			wantErr:      false,
		},
		{
			name:         "unknown resource",
			resourceName: "unknown",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, mimeType, err := server.ReadResource(context.Background(), tt.resourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("EchoServer.ReadResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if content == nil {
					t.Error("EchoServer.ReadResource() returned nil content when no error was expected")
				}
				if mimeType != "text/plain" {
					t.Errorf("EchoServer.ReadResource() returned mimeType = %v, want text/plain", mimeType)
				}
			}
		})
	}
}
