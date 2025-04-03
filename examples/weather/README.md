# Weather Service Example

This example demonstrates how to create a weather service using the Go-MCP framework and OpenWeatherMap API.

## Features

- Real-time weather information retrieval
- Support for multiple cities
- Temperature in Celsius
- Weather condition descriptions
- Both HTTP and WebSocket server support

## Prerequisites

Before running this example, you need:

1. An OpenWeatherMap API key (get one at [OpenWeatherMap](https://openweathermap.org/api))
2. Go 1.20 or later

## Configuration

Set your OpenWeatherMap API key as an environment variable:

```bash
export OPENWEATHERMAP_API_KEY=your_api_key_here
```

## Running the Example

```bash
go run main.go
```

This will start:
- HTTP server on port 8080
- WebSocket server on port 8081

## API Usage

### Prompts

The weather service provides a `weather` prompt:

```json
{
  "name": "weather",
  "args": {
    "city": "London"
  }
}
```

Response:
```json
{
  "content": "The weather in London is 15.2°C with scattered clouds"
}
```

### Tools

The service provides a `getWeather` tool:

```json
{
  "name": "getWeather",
  "args": {
    "city": "Tokyo"
  }
}
```

Response:
```json
{
  "output": {
    "temperature": 25.6,
    "description": "clear sky",
    "city": "Tokyo"
  }
}
```

### Resources

The service provides a `cities` resource that lists supported cities:

```json
GET /resources/cities
```

Response:
```json
[
  "London",
  "New York",
  "Tokyo",
  "Paris",
  "Beijing"
]
```

## Error Handling

The service handles various error cases:
- Invalid city names
- API key configuration issues
- OpenWeatherMap API errors
- Network connectivity problems

## Implementation Details

The example demonstrates:
- Implementing the MCP Server interface
- Handling prompts, tools, and resources
- Error handling and validation
- Graceful shutdown
- Environment variable configuration
- HTTP and WebSocket server setup

## Testing

You can test the service using curl:

```bash
# Get weather using HTTP
curl -X POST http://localhost:8080/prompts/weather -d '{"args":{"city":"London"}}'

# Get supported cities
curl http://localhost:8080/resources/cities
```

Or using WebSocket client libraries for real-time updates.