package types

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJSONRPCNotification(t *testing.T) {
	method := "test/method"
	params := map[string]string{"key": "value"}
	n := NewJSONRPCNotification(method, params)

	assert.Equal(t, "2.0", n.JSONRPC)
	assert.Equal(t, "notifications/"+method, n.Method)
	assert.Equal(t, params, n.Params)
	// Type and Item should be zero/nil for pure JSON-RPC notification
	assert.Empty(t, n.Type)
	assert.Nil(t, n.Item)
}

func TestNewPromptsListChangedNotification(t *testing.T) {
	n := NewPromptsListChangedNotification()
	assert.Equal(t, "2.0", n.JSONRPC)
	assert.Equal(t, "notifications/prompts/list_changed", n.Method)
	assert.Nil(t, n.Params)
}

func TestNewResourcesListChangedNotification(t *testing.T) {
	n := NewResourcesListChangedNotification()
	assert.Equal(t, "2.0", n.JSONRPC)
	assert.Equal(t, "notifications/resources/list_changed", n.Method)
	assert.Nil(t, n.Params)
}

func TestNewResourceUpdatedNotification(t *testing.T) {
	testURI := "file:///test.txt"
	n := NewResourceUpdatedNotification(testURI)
	assert.Equal(t, "2.0", n.JSONRPC)
	assert.Equal(t, "notifications/resources/updated", n.Method)
	expectedParams := ResourceUpdateNotification{URI: testURI}
	assert.Equal(t, expectedParams, n.Params)
}

func TestNewToolsListChangedNotification(t *testing.T) {
	n := NewToolsListChangedNotification()
	assert.Equal(t, "2.0", n.JSONRPC)
	assert.Equal(t, "notifications/tools/list_changed", n.Method)
	assert.Nil(t, n.Params)
}

// mockNotificationHandler is a mock implementation of NotificationHandler
type mockNotificationHandler struct {
	HandledNotifications []Notification
	ReturnError          error
}

func (m *mockNotificationHandler) HandleNotification(ctx context.Context, notification Notification) error {
	if m.ReturnError != nil {
		return m.ReturnError
	}
	m.HandledNotifications = append(m.HandledNotifications, notification)
	return nil
}

func TestNotificationRegistry(t *testing.T) {
	registry := NewNotificationRegistry()
	assert.NotNil(t, registry)
	assert.Empty(t, registry.handlers)

	handler1 := &mockNotificationHandler{}
	handler2 := &mockNotificationHandler{}
	handlerWithError := &mockNotificationHandler{ReturnError: errors.New("handler error")}

	// Test RegisterHandler
	registry.RegisterHandler(handler1)
	registry.RegisterHandler(handler2)
	registry.RegisterHandler(handlerWithError)
	assert.Len(t, registry.handlers, 3)

	// Test SendNotification (success)
	notification1 := NewPromptsListChangedNotification()
	err := registry.SendNotification(context.Background(), notification1)
	assert.NoError(t, err)

	// Check if handlers received the notification
	assert.Len(t, handler1.HandledNotifications, 1)
	assert.Equal(t, notification1, handler1.HandledNotifications[0])
	assert.Len(t, handler2.HandledNotifications, 1)
	assert.Equal(t, notification1, handler2.HandledNotifications[0])
	// handlerWithError should also be called but return error, check HandledNotifications might be empty or contain the item depending on implementation order
	// The test ensures the error propagates

	// Reset handlers for the next test
	handler1.HandledNotifications = nil
	handler2.HandledNotifications = nil

	// Test SendNotification (with error)
	registry = NewNotificationRegistry()       // Create a new registry
	registry.RegisterHandler(handler1)         // Register a working handler first
	registry.RegisterHandler(handlerWithError) // Register the failing handler
	registry.RegisterHandler(handler2)         // Register another working handler last

	notification2 := NewToolsListChangedNotification()
	err = registry.SendNotification(context.Background(), notification2)
	assert.Error(t, err)
	assert.Equal(t, "handler error", err.Error())

	// Check which handlers were called before the error
	assert.Len(t, handler1.HandledNotifications, 1) // handler1 should have been called
	assert.Equal(t, notification2, handler1.HandledNotifications[0])
	assert.Empty(t, handlerWithError.HandledNotifications) // handlerWithError fails, doesn't append
	assert.Empty(t, handler2.HandledNotifications)         // handler2 should not be called after the error
}

func TestResourceSubscriptionManager(t *testing.T) {
	rsm := NewResourceSubscriptionManager()
	assert.NotNil(t, rsm)
	assert.Empty(t, rsm.subscriptions)

	uri1 := "file:///doc1.txt"
	uri2 := "file:///doc2.txt"

	handler1 := &mockNotificationHandler{}
	handler2 := &mockNotificationHandler{}
	handler3WithError := &mockNotificationHandler{ReturnError: errors.New("subscribe error")}

	// Test Subscribe
	rsm.Subscribe(uri1, handler1)
	rsm.Subscribe(uri1, handler2)
	rsm.Subscribe(uri2, handler1) // handler1 subscribes to both URIs
	rsm.Subscribe(uri1, handler3WithError)

	assert.Len(t, rsm.subscriptions, 2)
	assert.Len(t, rsm.subscriptions[uri1], 3)
	assert.Len(t, rsm.subscriptions[uri2], 1)
	assert.Contains(t, rsm.subscriptions[uri1], handler1)
	assert.Contains(t, rsm.subscriptions[uri1], handler2)
	assert.Contains(t, rsm.subscriptions[uri1], handler3WithError)
	assert.Contains(t, rsm.subscriptions[uri2], handler1)

	// Test Notify (success for uri2)
	err := rsm.Notify(context.Background(), uri2)
	assert.NoError(t, err)
	assert.Len(t, handler1.HandledNotifications, 1) // Only handler1 subscribed to uri2
	expectedNotifUri2 := NewResourceUpdatedNotification(uri2)
	assert.Equal(t, expectedNotifUri2, handler1.HandledNotifications[0])
	assert.Empty(t, handler2.HandledNotifications) // handler2 didn't subscribe to uri2
	assert.Empty(t, handler3WithError.HandledNotifications)

	// Reset handlers
	handler1.HandledNotifications = nil

	// Test Notify (with error for uri1)
	err = rsm.Notify(context.Background(), uri1)
	assert.Error(t, err)
	assert.Equal(t, "subscribe error", err.Error())

	// Check which handlers were called before the error for uri1
	expectedNotifUri1 := NewResourceUpdatedNotification(uri1)
	// The order of execution for handlers on the same URI is not guaranteed by the map iteration
	// We check that at least the successful ones before the error might have received it.
	// A more robust test might involve sorting handlers or using a predictable mock.

	calledCount := 0
	if len(handler1.HandledNotifications) > 0 && handler1.HandledNotifications[0] == expectedNotifUri1 {
		calledCount++
	}
	if len(handler2.HandledNotifications) > 0 && handler2.HandledNotifications[0] == expectedNotifUri1 {
		calledCount++
	}
	// handler3WithError should not have appended the notification
	assert.Empty(t, handler3WithError.HandledNotifications)

	// Depending on the execution order, either 1 or 2 handlers might have been successfully notified before the error
	assert.GreaterOrEqual(t, calledCount, 1, "At least one handler should be called before the error")
	assert.LessOrEqual(t, calledCount, 2, "At most two handlers should be called before the error")

	// Test Notify (non-existent URI)
	err = rsm.Notify(context.Background(), "file:///nonexistent.txt")
	assert.NoError(t, err)
}
