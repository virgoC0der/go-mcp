package types

import "context"

// NotificationType defines the type of notification
type NotificationType string

const (
	// Resource and prompt change notifications
	NotificationPromptsListChanged   NotificationType = "prompts/list_changed"
	NotificationResourcesListChanged NotificationType = "resources/list_changed"
	NotificationResourceUpdated      NotificationType = "resources/updated"
	NotificationToolsListChanged     NotificationType = "tools/list_changed"

	// Legacy notification types (backward compatibility)
	NotificationPromptAdded     NotificationType = "promptAdded"
	NotificationPromptRemoved   NotificationType = "promptRemoved"
	NotificationPromptChanged   NotificationType = "promptChanged"
	NotificationToolAdded       NotificationType = "toolAdded"
	NotificationToolRemoved     NotificationType = "toolRemoved"
	NotificationToolChanged     NotificationType = "toolChanged"
	NotificationResourceAdded   NotificationType = "resourceAdded"
	NotificationResourceRemoved NotificationType = "resourceRemoved"
	NotificationResourceChanged NotificationType = "resourceChanged"
)

// ResourceUpdateNotification represents a resource update notification
type ResourceUpdateNotification struct {
	URI string `json:"uri"`
}

// Notification represents a notification about a change in the server
type Notification struct {
	Type NotificationType `json:"type"`
	Item any              `json:"item,omitempty"`

	// JSON-RPC notification specific fields
	JSONRPC string `json:"jsonrpc,omitempty"`
	Method  string `json:"method,omitempty"`
	Params  any    `json:"params,omitempty"`
}

// NewJSONRPCNotification creates a notification that complies with the JSON-RPC standard
func NewJSONRPCNotification(method string, params any) Notification {
	return Notification{
		JSONRPC: "2.0",
		Method:  "notifications/" + method,
		Params:  params,
	}
}

// NewPromptsListChangedNotification creates a prompt list change notification
func NewPromptsListChangedNotification() Notification {
	return NewJSONRPCNotification("prompts/list_changed", nil)
}

// NewResourcesListChangedNotification creates a resource list change notification
func NewResourcesListChangedNotification() Notification {
	return NewJSONRPCNotification("resources/list_changed", nil)
}

// NewResourceUpdatedNotification creates a resource update notification
func NewResourceUpdatedNotification(uri string) Notification {
	return NewJSONRPCNotification("resources/updated", ResourceUpdateNotification{
		URI: uri,
	})
}

// NewToolsListChangedNotification creates a tool list change notification
func NewToolsListChangedNotification() Notification {
	return NewJSONRPCNotification("tools/list_changed", nil)
}

// NotificationHandler is the interface for components that can receive notifications
type NotificationHandler interface {
	// HandleNotification processes a notification
	HandleNotification(ctx context.Context, notification Notification) error
}

// NotificationSender is the interface for components that can send notifications
type NotificationSender interface {
	// SendNotification sends a notification
	SendNotification(ctx context.Context, notification Notification) error
}

// NotificationRegistry is a registry for notification handlers
type NotificationRegistry struct {
	handlers []NotificationHandler
}

// NewNotificationRegistry creates a new notification registry
func NewNotificationRegistry() *NotificationRegistry {
	return &NotificationRegistry{
		handlers: make([]NotificationHandler, 0),
	}
}

// RegisterHandler registers a notification handler
func (nr *NotificationRegistry) RegisterHandler(handler NotificationHandler) {
	nr.handlers = append(nr.handlers, handler)
}

// SendNotification sends a notification to all registered handlers
func (nr *NotificationRegistry) SendNotification(ctx context.Context, notification Notification) error {
	for _, handler := range nr.handlers {
		if err := handler.HandleNotification(ctx, notification); err != nil {
			return err
		}
	}
	return nil
}

// ResourceSubscriptionManager 管理资源订阅
type ResourceSubscriptionManager struct {
	subscriptions map[string][]NotificationHandler
}

// NewResourceSubscriptionManager 创建一个新的资源订阅管理器
func NewResourceSubscriptionManager() *ResourceSubscriptionManager {
	return &ResourceSubscriptionManager{
		subscriptions: make(map[string][]NotificationHandler),
	}
}

// Subscribe 订阅资源变更
func (rsm *ResourceSubscriptionManager) Subscribe(uri string, handler NotificationHandler) {
	if _, exists := rsm.subscriptions[uri]; !exists {
		rsm.subscriptions[uri] = make([]NotificationHandler, 0)
	}
	rsm.subscriptions[uri] = append(rsm.subscriptions[uri], handler)
}

// Notify 通知所有订阅了特定资源的处理器
func (rsm *ResourceSubscriptionManager) Notify(ctx context.Context, uri string) error {
	handlers, exists := rsm.subscriptions[uri]
	if !exists {
		return nil
	}

	notification := NewResourceUpdatedNotification(uri)
	for _, handler := range handlers {
		if err := handler.HandleNotification(ctx, notification); err != nil {
			return err
		}
	}

	return nil
}
