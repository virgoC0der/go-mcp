package types

import (
	"context"
	"testing"
)

// TestNotificationHandler is a mock implementation of NotificationHandler for testing
type TestNotificationHandler struct {
	Notifications []Notification
}

// HandleNotification implements NotificationHandler.HandleNotification
func (h *TestNotificationHandler) HandleNotification(ctx context.Context, notification Notification) error {
	h.Notifications = append(h.Notifications, notification)
	return nil
}

func TestNotificationRegistry(t *testing.T) {
	// Create registry
	registry := NewNotificationRegistry()

	// Create handlers
	handler1 := &TestNotificationHandler{}
	handler2 := &TestNotificationHandler{}

	// Register handlers
	registry.RegisterHandler(handler1)
	registry.RegisterHandler(handler2)

	// Create a notification
	notification := Notification{
		Type: NotificationPromptAdded,
		Item: Prompt{
			Name:        "test_prompt",
			Description: "Test prompt",
		},
	}

	// Send notification
	err := registry.SendNotification(context.Background(), notification)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	// Check that both handlers received the notification
	if len(handler1.Notifications) != 1 {
		t.Errorf("Expected handler1 to receive 1 notification, got %d", len(handler1.Notifications))
	}

	if len(handler2.Notifications) != 1 {
		t.Errorf("Expected handler2 to receive 1 notification, got %d", len(handler2.Notifications))
	}

	// Check notification content in handler1
	if handler1.Notifications[0].Type != NotificationPromptAdded {
		t.Errorf("Notification type mismatch: expected %s, got %s",
			NotificationPromptAdded, handler1.Notifications[0].Type)
	}

	// Check notification content in handler2
	if handler2.Notifications[0].Type != NotificationPromptAdded {
		t.Errorf("Notification type mismatch: expected %s, got %s",
			NotificationPromptAdded, handler2.Notifications[0].Type)
	}
}

func TestNotificationConstants(t *testing.T) {
	// Test constant values
	if NotificationPromptAdded != "promptAdded" {
		t.Errorf("Expected NotificationPromptAdded to be 'promptAdded', got '%s'", NotificationPromptAdded)
	}

	if NotificationToolAdded != "toolAdded" {
		t.Errorf("Expected NotificationToolAdded to be 'toolAdded', got '%s'", NotificationToolAdded)
	}

	if NotificationResourceAdded != "resourceAdded" {
		t.Errorf("Expected NotificationResourceAdded to be 'resourceAdded', got '%s'", NotificationResourceAdded)
	}

	if NotificationPromptRemoved != "promptRemoved" {
		t.Errorf("Expected NotificationPromptRemoved to be 'promptRemoved', got '%s'", NotificationPromptRemoved)
	}

	if NotificationToolRemoved != "toolRemoved" {
		t.Errorf("Expected NotificationToolRemoved to be 'toolRemoved', got '%s'", NotificationToolRemoved)
	}

	if NotificationResourceRemoved != "resourceRemoved" {
		t.Errorf("Expected NotificationResourceRemoved to be 'resourceRemoved', got '%s'", NotificationResourceRemoved)
	}

	if NotificationPromptChanged != "promptChanged" {
		t.Errorf("Expected NotificationPromptChanged to be 'promptChanged', got '%s'", NotificationPromptChanged)
	}

	if NotificationToolChanged != "toolChanged" {
		t.Errorf("Expected NotificationToolChanged to be 'toolChanged', got '%s'", NotificationToolChanged)
	}

	if NotificationResourceChanged != "resourceChanged" {
		t.Errorf("Expected NotificationResourceChanged to be 'resourceChanged', got '%s'", NotificationResourceChanged)
	}
}
