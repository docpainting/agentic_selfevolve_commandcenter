package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// StdioTransport handles stdio-based MCP communication
type StdioTransport struct {
	reader      *bufio.Reader
	writer      io.Writer
	mu          sync.Mutex
	pendingReqs map[int]chan *MCPResponse
	nextID      int
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(reader io.Reader, writer io.Writer) *StdioTransport {
	return &StdioTransport{
		reader:      bufio.NewReader(reader),
		writer:      writer,
		pendingReqs: make(map[int]chan *MCPResponse),
		nextID:      1,
	}
}

// Send sends a request and returns a response channel
func (t *StdioTransport) Send(method string, params map[string]interface{}) (<-chan *MCPResponse, error) {
	t.mu.Lock()
	id := t.nextID
	t.nextID++

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.mu.Unlock()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create response channel
	respChan := make(chan *MCPResponse, 1)
	t.pendingReqs[id] = respChan

	// Send request
	if _, err := t.writer.Write(append(data, '\n')); err != nil {
		delete(t.pendingReqs, id)
		t.mu.Unlock()
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	t.mu.Unlock()
	return respChan, nil
}

// SendAndWait sends a request and waits for response
func (t *StdioTransport) SendAndWait(method string, params map[string]interface{}) (*MCPResponse, error) {
	respChan, err := t.Send(method, params)
	if err != nil {
		return nil, err
	}

	resp := <-respChan
	return resp, nil
}

// StartReading starts reading responses
func (t *StdioTransport) StartReading() {
	go t.readLoop()
}

// readLoop continuously reads responses
func (t *StdioTransport) readLoop() {
	for {
		line, err := t.reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading from MCP server: %v\n", err)
			}
			break
		}

		var resp MCPResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			fmt.Printf("Error unmarshaling MCP response: %v\n", err)
			continue
		}

		// Find pending request
		t.mu.Lock()
		if respChan, exists := t.pendingReqs[resp.ID]; exists {
			respChan <- &resp
			close(respChan)
			delete(t.pendingReqs, resp.ID)
		}
		t.mu.Unlock()
	}
}

// Close closes the transport
func (t *StdioTransport) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Close all pending channels
	for id, ch := range t.pendingReqs {
		close(ch)
		delete(t.pendingReqs, id)
	}
}

// MCPConnection wraps a transport with convenience methods
type MCPConnection struct {
	transport *StdioTransport
	tools     []Tool
	mu        sync.RWMutex
}

// NewMCPConnection creates a new MCP connection
func NewMCPConnection(reader io.Reader, writer io.Writer) *MCPConnection {
	transport := NewStdioTransport(reader, writer)
	transport.StartReading()

	return &MCPConnection{
		transport: transport,
		tools:     make([]Tool, 0),
	}
}

// Initialize initializes the MCP connection
func (c *MCPConnection) Initialize() error {
	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "agent-workspace",
			"version": "1.0.0",
		},
	}

	resp, err := c.transport.SendAndWait("initialize", params)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("initialization error: %s", resp.Error.Message)
	}

	// List tools after initialization
	return c.RefreshTools()
}

// RefreshTools refreshes the tool list
func (c *MCPConnection) RefreshTools() error {
	resp, err := c.transport.SendAndWait("tools/list", nil)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("tools/list error: %s", resp.Error.Message)
	}

	// Parse tools
	toolsData, ok := resp.Result["tools"].([]interface{})
	if !ok {
		return nil
	}

	tools := make([]Tool, 0, len(toolsData))
	for _, t := range toolsData {
		toolMap, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		tool := Tool{
			Name:        getString(toolMap, "name"),
			Description: getString(toolMap, "description"),
		}

		if schema, ok := toolMap["inputSchema"].(map[string]interface{}); ok {
			tool.InputSchema = schema
		}

		tools = append(tools, tool)
	}

	c.mu.Lock()
	c.tools = tools
	c.mu.Unlock()

	return nil
}

// GetTools returns the list of tools
func (c *MCPConnection) GetTools() []Tool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tools := make([]Tool, len(c.tools))
	copy(tools, c.tools)
	return tools
}

// CallTool calls a tool
func (c *MCPConnection) CallTool(name string, args map[string]interface{}) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": args,
	}

	resp, err := c.transport.SendAndWait("tools/call", params)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tool call error: %s", resp.Error.Message)
	}

	return resp.Result, nil
}

// Close closes the connection
func (c *MCPConnection) Close() {
	c.transport.Close()
}

// LoadMCPConfig loads MCP server configuration from file
func LoadMCPConfig(filepath string) (map[string]ServerConfig, error) {
	// TODO: Implement config loading from JSON file
	// For now, return default config
	return map[string]ServerConfig{
		"dynamic-thinking": {
			Command: "./mcp-dynamic-thinking/mcp-dynamic-thinking",
			Args:    []string{},
			Env: map[string]string{
				"NEO4J_URI": "bolt://localhost:7687",
			},
		},
	}, nil
}

// ServerConfig represents MCP server configuration
type ServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

