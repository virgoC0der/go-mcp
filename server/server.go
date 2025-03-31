package server

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/virgoC0der/go-mcp/types"
)

// Server defines the interface for an MCP server
type Server interface {
	// Initialize initializes the server with the given options
	Initialize(ctx context.Context, options interface{}) error

	// GetPrompt retrieves a prompt with the given name and arguments
	GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*types.GetPromptResult, error)

	// ListPrompts lists all available prompts
	ListPrompts(ctx context.Context) ([]types.Prompt, error)

	// ListPromptsPaginated lists prompts with pagination
	ListPromptsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error)

	// CallTool calls a tool with the given name and arguments
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*types.CallToolResult, error)

	// ListTools lists all available tools
	ListTools(ctx context.Context) ([]types.Tool, error)

	// ListToolsPaginated lists tools with pagination
	ListToolsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error)

	// ReadResource reads a resource with the given name
	ReadResource(ctx context.Context, name string) ([]byte, string, error)

	// ListResources lists all available resources
	ListResources(ctx context.Context) ([]types.Resource, error)

	// ListResourcesPaginated lists resources with pagination
	ListResourcesPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error)

	// RegisterPrompt registers a prompt with the server
	RegisterPrompt(prompt types.Prompt)

	// RegisterPromptTyped registers a prompt with the server using a typed handler
	RegisterPromptTyped(name string, description string, handler interface{}) error

	// RegisterTool registers a tool with the server
	RegisterTool(tool types.Tool)

	// RegisterToolTyped registers a tool with the server using a typed handler
	RegisterToolTyped(name string, description string, handler interface{}) error

	// RegisterResource registers a resource with the server
	RegisterResource(resource types.Resource)

	// RegisterNotificationHandler registers a notification handler
	RegisterNotificationHandler(handler types.NotificationHandler)
}

// BaseServer is a base implementation of the Server interface
type BaseServer struct {
	name          string
	version       string
	prompts       map[string]types.Prompt
	tools         map[string]types.Tool
	resources     map[string]types.Resource
	notifications *types.NotificationRegistry
	schemaGen     *types.SchemaGenerator

	// Handlers for typed operations
	promptHandlers map[string]interface{}
	toolHandlers   map[string]interface{}
}

// NewBaseServer creates a new base server
func NewBaseServer(name, version string) *BaseServer {
	return &BaseServer{
		name:           name,
		version:        version,
		prompts:        make(map[string]types.Prompt),
		tools:          make(map[string]types.Tool),
		resources:      make(map[string]types.Resource),
		notifications:  types.NewNotificationRegistry(),
		schemaGen:      types.NewSchemaGenerator(),
		promptHandlers: make(map[string]interface{}),
		toolHandlers:   make(map[string]interface{}),
	}
}

// Initialize implements Server.Initialize
func (s *BaseServer) Initialize(ctx context.Context, options interface{}) error {
	// Check if options is of the right type
	opts, ok := options.(types.InitializationOptions)
	if !ok {
		// Try to convert from map[string]interface{}
		if optsMap, ok := options.(map[string]interface{}); ok {
			if serverName, ok := optsMap["serverName"].(string); ok {
				s.name = serverName
			}
			if serverVersion, ok := optsMap["serverVersion"].(string); ok {
				s.version = serverVersion
			}
			// Capabilities could be parsed here if needed
		} else {
			return fmt.Errorf("invalid initialization options type")
		}
	} else {
		s.name = opts.ServerName
		s.version = opts.ServerVersion
	}

	return nil
}

// GetPrompt implements Server.GetPrompt
func (s *BaseServer) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*types.GetPromptResult, error) {
	prompt, ok := s.prompts[name]
	if !ok {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}

	// If there's a typed handler, use it
	if _, ok := s.promptHandlers[name]; ok {
		// Here we would deserialize arguments and call the handler
		// This is a placeholder for the actual implementation
		return &types.GetPromptResult{
			Description: prompt.Description,
			Message:     fmt.Sprintf("Prompt %s called with typed handler", name),
		}, nil
	}

	// Otherwise, return a default response
	return &types.GetPromptResult{
		Description: prompt.Description,
		Message:     fmt.Sprintf("Prompt %s called with arguments: %v", name, arguments),
	}, nil
}

// ListPrompts implements Server.ListPrompts
func (s *BaseServer) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	prompts := make([]types.Prompt, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		prompts = append(prompts, prompt)
	}
	return prompts, nil
}

// ListPromptsPaginated implements Server.ListPromptsPaginated
func (s *BaseServer) ListPromptsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	prompts := make([]types.Prompt, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		prompts = append(prompts, prompt)
	}

	// Apply pagination
	totalItems := len(prompts)
	startIdx := (options.Page - 1) * options.PageSize
	endIdx := startIdx + options.PageSize

	if startIdx > totalItems {
		startIdx = totalItems
	}
	if endIdx > totalItems {
		endIdx = totalItems
	}

	pageItems := prompts[startIdx:endIdx]

	return &types.PaginatedResult{
		Items:       pageItems,
		TotalItems:  totalItems,
		CurrentPage: options.Page,
		PageSize:    options.PageSize,
		TotalPages:  (totalItems + options.PageSize - 1) / options.PageSize,
	}, nil
}

// CallTool implements Server.CallTool
func (s *BaseServer) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*types.CallToolResult, error) {
	tool, ok := s.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	// Use tool.Name to avoid the unused variable warning
	toolName := tool.Name

	// If there's a typed handler, use it
	if _, ok := s.toolHandlers[name]; ok {
		// Here we would deserialize arguments and call the handler
		// This is a placeholder for the actual implementation
		return &types.CallToolResult{
			Content: fmt.Sprintf("Tool %s called with typed handler", toolName),
		}, nil
	}

	// Otherwise, return a default response
	return &types.CallToolResult{
		Content: fmt.Sprintf("Tool %s called with arguments: %v", toolName, arguments),
	}, nil
}

// ListTools implements Server.ListTools
func (s *BaseServer) ListTools(ctx context.Context) ([]types.Tool, error) {
	tools := make([]types.Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}
	return tools, nil
}

// ListToolsPaginated implements Server.ListToolsPaginated
func (s *BaseServer) ListToolsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	tools := make([]types.Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}

	// Apply pagination
	totalItems := len(tools)
	startIdx := (options.Page - 1) * options.PageSize
	endIdx := startIdx + options.PageSize

	if startIdx > totalItems {
		startIdx = totalItems
	}
	if endIdx > totalItems {
		endIdx = totalItems
	}

	pageItems := tools[startIdx:endIdx]

	return &types.PaginatedResult{
		Items:       pageItems,
		TotalItems:  totalItems,
		CurrentPage: options.Page,
		PageSize:    options.PageSize,
		TotalPages:  (totalItems + options.PageSize - 1) / options.PageSize,
	}, nil
}

// ReadResource implements Server.ReadResource
func (s *BaseServer) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	resource, ok := s.resources[name]
	if !ok {
		return nil, "", fmt.Errorf("resource not found: %s", name)
	}

	// In a real implementation, we would fetch the actual resource content
	content := []byte(fmt.Sprintf("Content of resource: %s", name))
	return content, resource.MimeType, nil
}

// ListResources implements Server.ListResources
func (s *BaseServer) ListResources(ctx context.Context) ([]types.Resource, error) {
	resources := make([]types.Resource, 0, len(s.resources))
	for _, resource := range s.resources {
		resources = append(resources, resource)
	}
	return resources, nil
}

// ListResourcesPaginated implements Server.ListResourcesPaginated
func (s *BaseServer) ListResourcesPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	resources := make([]types.Resource, 0, len(s.resources))
	for _, resource := range s.resources {
		resources = append(resources, resource)
	}

	// Apply pagination
	totalItems := len(resources)
	startIdx := (options.Page - 1) * options.PageSize
	endIdx := startIdx + options.PageSize

	if startIdx > totalItems {
		startIdx = totalItems
	}
	if endIdx > totalItems {
		endIdx = totalItems
	}

	pageItems := resources[startIdx:endIdx]

	return &types.PaginatedResult{
		Items:       pageItems,
		TotalItems:  totalItems,
		CurrentPage: options.Page,
		PageSize:    options.PageSize,
		TotalPages:  (totalItems + options.PageSize - 1) / options.PageSize,
	}, nil
}

// RegisterPrompt implements Server.RegisterPrompt
func (s *BaseServer) RegisterPrompt(prompt types.Prompt) {
	s.prompts[prompt.Name] = prompt

	// Notify handlers of the new prompt
	if s.notifications != nil {
		if err := s.notifications.SendNotification(context.Background(), types.Notification{
			Type: types.NotificationPromptAdded,
			Item: prompt,
		}); err != nil {
			log.Printf("Failed to send notification: %v", err)
		}
	}
}

// RegisterPromptTyped implements Server.RegisterPromptTyped
func (s *BaseServer) RegisterPromptTyped(name, description string, handler interface{}) error {
	// Validate handler is a function
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("handler must be a function")
	}

	// Check handler signature
	if handlerType.NumOut() != 2 {
		return fmt.Errorf("handler must return two values (result and error)")
	}
	if !handlerType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return fmt.Errorf("second return value must be an error")
	}

	// Register the handler
	s.promptHandlers[name] = handler

	// Generate argument schema from the first parameter type
	var args []types.PromptArgument

	if handlerType.NumIn() > 0 {
		paramType := handlerType.In(0)
		if paramType.Kind() == reflect.Struct {
			// Create an instance of the parameter type
			paramInstance := reflect.New(paramType).Elem().Interface()

			// Extract argument names and descriptions
			args = s.schemaGen.GetArgumentNames(paramInstance)
		}
	}

	// Register the prompt with the extracted arguments
	prompt := types.Prompt{
		Name:        name,
		Description: description,
		Arguments:   args,
	}

	s.prompts[name] = prompt

	// Notify handlers of the new prompt
	if s.notifications != nil {
		if err := s.notifications.SendNotification(context.Background(), types.Notification{
			Type: types.NotificationPromptAdded,
			Item: prompt,
		}); err != nil {
			log.Printf("Failed to send notification: %v", err)
		}
	}

	return nil
}

// RegisterTool implements Server.RegisterTool
func (s *BaseServer) RegisterTool(tool types.Tool) {
	s.tools[tool.Name] = tool

	// Notify handlers of the new tool
	if s.notifications != nil {
		if err := s.notifications.SendNotification(context.Background(), types.Notification{
			Type: types.NotificationToolAdded,
			Item: tool,
		}); err != nil {
			log.Printf("Failed to send notification: %v", err)
		}
	}
}

// RegisterToolTyped implements Server.RegisterToolTyped
func (s *BaseServer) RegisterToolTyped(name, description string, handler interface{}) error {
	// Validate handler is a function
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("handler must be a function")
	}

	// Check handler signature
	if handlerType.NumOut() != 2 {
		return fmt.Errorf("handler must return two values (result and error)")
	}
	if !handlerType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return fmt.Errorf("second return value must be an error")
	}

	// Register the handler
	s.toolHandlers[name] = handler

	// Generate argument schema from the first parameter type
	var args []types.PromptArgument
	var schema interface{}

	if handlerType.NumIn() > 0 {
		paramType := handlerType.In(0)
		if paramType.Kind() == reflect.Struct {
			// Create an instance of the parameter type
			paramInstance := reflect.New(paramType).Elem().Interface()

			// Generate schema for the parameter
			var err error
			schema, err = s.schemaGen.GenerateSchema(paramInstance)
			if err != nil {
				return fmt.Errorf("failed to generate schema: %w", err)
			}

			// Extract argument names and descriptions
			args = s.schemaGen.GetArgumentNames(paramInstance)
		}
	}

	// Register the tool with the extracted arguments and schema
	tool := types.Tool{
		Name:        name,
		Description: description,
		Arguments:   args,
		Schema:      schema,
	}

	s.tools[name] = tool

	// Notify handlers of the new tool
	if s.notifications != nil {
		if err := s.notifications.SendNotification(context.Background(), types.Notification{
			Type: types.NotificationToolAdded,
			Item: tool,
		}); err != nil {
			log.Printf("Failed to send notification: %v", err)
		}
	}

	return nil
}

// RegisterResource implements Server.RegisterResource
func (s *BaseServer) RegisterResource(resource types.Resource) {
	s.resources[resource.Name] = resource

	// Notify handlers of the new resource
	if s.notifications != nil {
		if err := s.notifications.SendNotification(context.Background(), types.Notification{
			Type: types.NotificationResourceAdded,
			Item: resource,
		}); err != nil {
			log.Printf("Failed to send notification: %v", err)
		}
	}
}

// RegisterNotificationHandler implements Server.RegisterNotificationHandler
func (s *BaseServer) RegisterNotificationHandler(handler types.NotificationHandler) {
	if s.notifications != nil {
		s.notifications.RegisterHandler(handler)
	}
}
