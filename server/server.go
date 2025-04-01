package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/virgoC0der/go-mcp/types"
)

// Server defines the interface for an MCP server
type Server interface {
	// Initialize initializes the server with the given options
	Initialize(ctx context.Context, options any) error

	// GetPrompt retrieves a prompt with the given name and arguments
	GetPrompt(ctx context.Context, name string, arguments map[string]any) (*types.GetPromptResult, error)

	// ListPrompts lists all available prompts
	ListPrompts(ctx context.Context) ([]types.Prompt, error)

	// ListPromptsPaginated lists prompts with pagination
	ListPromptsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error)

	// CallTool calls a tool with the given name and arguments
	CallTool(ctx context.Context, name string, arguments map[string]any) (*types.CallToolResult, error)

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
	RegisterPromptTyped(name string, description string, handler any) error

	// RegisterTool registers a tool with the server
	RegisterTool(tool types.Tool)

	// RegisterToolTyped registers a tool with the server using a typed handler
	// handler must be a function that accepts a struct parameter and returns (*types.CallToolResult, error)
	// This approach reduces the use of runtime reflection, improving performance
	RegisterToolTyped(name string, description string, handler any) error

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
	promptHandlers map[string]TypedPromptHandler
	toolHandlers   map[string]TypedToolHandler
}

// TypedToolHandler defines the interface for tool processing functions
type TypedToolHandler interface {
	Execute(ctx context.Context, arguments map[string]any) (*types.CallToolResult, error)
}

// TypedPromptHandler defines the interface for prompt processing functions
type TypedPromptHandler interface {
	Execute(ctx context.Context, arguments map[string]any) (*types.GetPromptResult, error)
}

// typedToolHandlerImpl is a generic implementation of TypedToolHandler
type typedToolHandlerImpl[T any] struct {
	handler   ToolHandler[T]
	paramType reflect.Type
	schemaGen *types.SchemaGenerator
}

// Execute implements TypedToolHandler interface
func (h *typedToolHandlerImpl[T]) Execute(ctx context.Context, arguments map[string]any) (*types.CallToolResult, error) {
	jsonData, err := json.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	var request T
	if err := json.Unmarshal(jsonData, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	// Directly call the handler function, no reflection needed
	return h.handler(request)
}

// ToolHandler defines the type for tool processing functions
type ToolHandler[T any] func(request T) (*types.CallToolResult, error)

// PromptHandler defines the type for prompt processing functions
type PromptHandler[T any] func(request T) (*types.GetPromptResult, error)

// genericToolHandlerImpl is a generic implementation of TypedToolHandler
type genericToolHandlerImpl struct {
	handler     any
	handlerType reflect.Type
	schemaGen   *types.SchemaGenerator
}

// Execute implements TypedToolHandler interface
func (h *genericToolHandlerImpl) Execute(ctx context.Context, arguments map[string]any) (*types.CallToolResult, error) {
	handlerValue := reflect.ValueOf(h.handler)

	// If the handler function doesn't have parameters, call it directly
	if h.handlerType.NumIn() == 0 {
		results := handlerValue.Call(nil)

		// Process function results
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}

		return results[0].Interface().(*types.CallToolResult), nil
	}

	// Get parameter type
	paramType := h.handlerType.In(0)

	// Use JSON conversion to populate parameters, instead of using reflection to set fields one by one
	jsonData, err := json.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	// Create a new instance of the parameter type
	paramValue := reflect.New(paramType)

	// Use JSON parsing to populate parameters
	if err := json.Unmarshal(jsonData, paramValue.Interface()); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	// Get Elem() because paramValue is a pointer
	if paramType.Kind() != reflect.Ptr {
		paramValue = paramValue.Elem()
	}

	// Call the handler function
	results := handlerValue.Call([]reflect.Value{paramValue})

	// Process function results
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	return results[0].Interface().(*types.CallToolResult), nil
}

// genericPromptHandlerImpl is a generic implementation of TypedPromptHandler
type genericPromptHandlerImpl struct {
	handler     any
	handlerType reflect.Type
	schemaGen   *types.SchemaGenerator
}

// Execute implements TypedPromptHandler interface
func (h *genericPromptHandlerImpl) Execute(ctx context.Context, arguments map[string]any) (*types.GetPromptResult, error) {
	handlerValue := reflect.ValueOf(h.handler)

	// If the handler function doesn't have parameters, call it directly
	if h.handlerType.NumIn() == 0 {
		results := handlerValue.Call(nil)

		// Process function results
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}

		return results[0].Interface().(*types.GetPromptResult), nil
	}

	// Get parameter type
	paramType := h.handlerType.In(0)

	// Use JSON conversion to populate parameters, instead of using reflection to set fields one by one
	jsonData, err := json.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	paramValue := reflect.New(paramType)

	if err := json.Unmarshal(jsonData, paramValue.Interface()); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	if paramType.Kind() != reflect.Ptr {
		paramValue = paramValue.Elem()
	}

	results := handlerValue.Call([]reflect.Value{paramValue})

	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	return results[0].Interface().(*types.GetPromptResult), nil
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
		promptHandlers: make(map[string]TypedPromptHandler),
		toolHandlers:   make(map[string]TypedToolHandler),
	}
}

// Initialize implements Server.Initialize
func (s *BaseServer) Initialize(ctx context.Context, options any) error {
	// Check if options is of the right type
	opts, ok := options.(types.InitializationOptions)
	if !ok {
		// Try to convert from map[string]any
		if optsMap, ok := options.(map[string]any); ok {
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
func (s *BaseServer) GetPrompt(ctx context.Context, name string, arguments map[string]any) (*types.GetPromptResult, error) {
	prompt, ok := s.prompts[name]
	if !ok {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}

	if handler, ok := s.promptHandlers[name]; ok {
		return handler.Execute(ctx, arguments)
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
func (s *BaseServer) CallTool(ctx context.Context, name string, arguments map[string]any) (*types.CallToolResult, error) {
	_, ok := s.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	// Use TypedToolHandler interface to directly call the handler function, avoiding reflection
	if handler, ok := s.toolHandlers[name]; ok {
		return handler.Execute(ctx, arguments)
	}

	// Otherwise, return a default response
	return &types.CallToolResult{
		Content: fmt.Sprintf("Tool %s called with arguments: %v", name, arguments),
	}, nil
}

// fillStructFromMap fills the struct with values from the map
func (s *BaseServer) fillStructFromMap(structValue reflect.Value, m map[string]any) error {
	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// Get json tag name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = strings.ToLower(fieldType.Name)
		} else {
			// Handle json tags with options
			jsonTag = strings.Split(jsonTag, ",")[0]
		}

		// Check if there is a corresponding value in the map
		if value, ok := m[jsonTag]; ok {
			// Set field value
			if err := s.setFieldValue(field, value); err != nil {
				return fmt.Errorf("failed to set field %s: %w", fieldType.Name, err)
			}
		}
	}
	return nil
}

// setFieldValue sets the value of a struct field
func (s *BaseServer) setFieldValue(field reflect.Value, value any) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	// Get reflection values of target type and source value
	targetType := field.Type()
	sourceValue := reflect.ValueOf(value)

	// If the source value can be directly converted to the target type
	if sourceValue.Type().ConvertibleTo(targetType) {
		field.Set(sourceValue.Convert(targetType))
		return nil
	}

	// Handle some common type conversion cases
	switch targetType.Kind() {
	case reflect.String:
		// Try to convert value to string
		field.SetString(fmt.Sprintf("%v", value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Try to convert value to int
		v, ok := value.(float64)
		if ok {
			field.SetInt(int64(v))
		} else if strValue, ok := value.(string); ok {
			if intValue, err := strconv.ParseInt(strValue, 10, 64); err == nil {
				field.SetInt(intValue)
			} else {
				return fmt.Errorf("cannot convert string to int: %v", err)
			}
		} else {
			return fmt.Errorf("cannot convert %T to int", value)
		}
	case reflect.Float32, reflect.Float64:
		// Try to convert value to float
		v, ok := value.(float64)
		if ok {
			field.SetFloat(v)
		} else if strValue, ok := value.(string); ok {
			if floatValue, err := strconv.ParseFloat(strValue, 64); err == nil {
				field.SetFloat(floatValue)
			} else {
				return fmt.Errorf("cannot convert string to float: %v", err)
			}
		} else {
			return fmt.Errorf("cannot convert %T to float", value)
		}
	case reflect.Bool:
		// Try to convert value to bool
		v, ok := value.(bool)
		if ok {
			field.SetBool(v)
		} else if strValue, ok := value.(string); ok {
			if boolValue, err := strconv.ParseBool(strValue); err == nil {
				field.SetBool(boolValue)
			} else {
				return fmt.Errorf("cannot convert string to bool: %v", err)
			}
		} else {
			return fmt.Errorf("cannot convert %T to bool", value)
		}
	default:
		return fmt.Errorf("unsupported type: %s", targetType)
	}

	return nil
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
func (s *BaseServer) RegisterPromptTyped(name, description string, handler any) error {
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

	// Check if the first return value is *types.GetPromptResult
	if handlerType.Out(0).Kind() != reflect.Ptr ||
		handlerType.Out(0).Elem().Name() != "GetPromptResult" {
		return fmt.Errorf("first return value must be *types.GetPromptResult")
	}

	// Create a generic handler wrapper
	typedHandler := &genericPromptHandlerImpl{
		handler:     handler,
		handlerType: handlerType,
		schemaGen:   s.schemaGen,
	}

	// Register the handler function
	s.promptHandlers[name] = typedHandler

	// Generate parameter schema and parameter descriptions
	var args []types.PromptArgument

	if handlerType.NumIn() > 0 {
		paramType := handlerType.In(0)
		if paramType.Kind() == reflect.Struct {
			// Create an instance of the parameter type
			paramInstance := reflect.New(paramType).Elem().Interface()

			// Extract parameter names and descriptions
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
func (s *BaseServer) RegisterToolTyped(name, description string, handler any) error {
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

	// Check if the first return value is *types.CallToolResult
	if handlerType.Out(0).Kind() != reflect.Ptr ||
		handlerType.Out(0).Elem().Name() != "CallToolResult" {
		return fmt.Errorf("first return value must be *types.CallToolResult")
	}

	// Create a generic handler wrapper
	typedHandler := &genericToolHandlerImpl{
		handler:     handler,
		handlerType: handlerType,
		schemaGen:   s.schemaGen,
	}

	// Register the handler function
	s.toolHandlers[name] = typedHandler

	// Generate parameter schema and parameter descriptions
	var args []types.PromptArgument
	var schema any

	// Only generate schema when handling struct parameters
	if handlerType.NumIn() > 0 {
		paramType := handlerType.In(0)
		if paramType.Kind() == reflect.Struct {
			// Create a zero value instance of the parameter type through reflection
			paramInstance := reflect.New(paramType).Elem().Interface()

			// Generate parameter schema
			var err error
			schema, err = s.schemaGen.GenerateSchema(paramInstance)
			if err != nil {
				return fmt.Errorf("failed to generate schema: %w", err)
			}

			// 提取参数名称和描述
			args = s.schemaGen.GetArgumentNames(paramInstance)
		}
	}

	// 注册工具
	tool := types.Tool{
		Name:        name,
		Description: description,
		Arguments:   args,
		Schema:      schema,
	}

	s.tools[name] = tool

	// 发送通知
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

// CreateToolHandler 创建一个类型安全的工具处理函数
// 该函数用于创建与RegisterToolTyped兼容的处理函数
// 使用例子:
//
//	type WeatherRequest struct {
//	  Longitude float64 `json:"longitude"`
//	  Latitude  float64 `json:"latitude"`
//	}
//
// server.RegisterToolTyped("getWeather", "获取天气",
//
//	server.CreateToolHandler(func(req WeatherRequest) (*types.CallToolResult, error) {
//	  // 处理逻辑...
//	  return &types.CallToolResult{...}, nil
//	})
//
// )
func CreateToolHandler[T any](handler ToolHandler[T]) any {
	return handler
}

// CreatePromptHandler 创建一个类型安全的提示处理函数
// 该函数用于创建与RegisterPromptTyped兼容的处理函数
// 使用例子:
//
//	type GreetInput struct {
//	  Name    string `json:"name"`
//	  Formal  bool   `json:"formal"`
//	}
//
// server.RegisterPromptTyped("greet", "问候信息",
//
//	server.CreatePromptHandler(func(req GreetInput) (*types.GetPromptResult, error) {
//	  // 处理逻辑...
//	  return &types.GetPromptResult{...}, nil
//	})
//
// )
func CreatePromptHandler[T any](handler PromptHandler[T]) any {
	return handler
}
