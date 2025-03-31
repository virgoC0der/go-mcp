package types

import "context"

// NotificationType defines the type of notification
type NotificationType string

const (
	// NotificationPromptAdded is sent when a new prompt is added
	NotificationPromptAdded NotificationType = "promptAdded"
	// NotificationPromptRemoved is sent when a prompt is removed
	NotificationPromptRemoved NotificationType = "promptRemoved"
	// NotificationPromptChanged is sent when a prompt is changed
	NotificationPromptChanged NotificationType = "promptChanged"
	
	// NotificationToolAdded is sent when a new tool is added
	NotificationToolAdded NotificationType = "toolAdded"
	// NotificationToolRemoved is sent when a tool is removed
	NotificationToolRemoved NotificationType = "toolRemoved"
	// NotificationToolChanged is sent when a tool is changed
	NotificationToolChanged NotificationType = "toolChanged"
	
	// NotificationResourceAdded is sent when a new resource is added
	NotificationResourceAdded NotificationType = "resourceAdded"
	// NotificationResourceRemoved is sent when a resource is removed
	NotificationResourceRemoved NotificationType = "resourceRemoved"
	// NotificationResourceChanged is sent when a resource is changed
	NotificationResourceChanged NotificationType = "resourceChanged"
)

// Notification represents a notification about a change in the server
type Notification struct {
	Type NotificationType `json:"type"`
	Item interface{}      `json:"item"`
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