package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/virgoC0der/go-mcp/internal/types"
)

func TestWithServerAddress(t *testing.T) {
	options := &types.ClientOptions{}
	testAddress := "remotehost:1234"
	option := WithServerAddress(testAddress)
	option(options)
	assert.Equal(t, testAddress, options.ServerAddress)
}

func TestNewClient(t *testing.T) {
	options := &types.ClientOptions{
		ServerAddress: "localhost:8080",
		Type:          "http", // Specify a type, e.g., http
	}

	client, err := NewClient(options)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	// Verify the returned object implements the types.Client interface
	assert.Implements(t, (*types.Client)(nil), client)

	// Test with nil options (should use defaults from factory)
	clientDefault, errDefault := NewClient(nil)
	assert.NoError(t, errDefault)
	assert.NotNil(t, clientDefault)
	assert.Implements(t, (*types.Client)(nil), clientDefault)
}
