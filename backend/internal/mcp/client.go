package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"agent-workspace/backend/pkg/models"
)

// Client manages MCP server connections
type Client struct {
	servers map[string]*Server
	mu      sync.RWMutex
}

// Server represents an MCP server connection
type Server struct {
	Name    string
	Command string
	Args    []string
	Env     map[string]string
	Process *exec.Cmd
	Stdin   io.WriteCloser
	Stdout  io.ReadCloser
	Tools   []Tool
	Status  string
	mu      sync.Mutex
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPRequest represents an MCP request
type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   *MCPError              `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewClient creates a new MCP client
func NewClient() *Client {
	return &Client{
		servers: make(map[string]*Server),
	}
}

// ConnectServer connects to an MCP server
func (c *Client) ConnectServer(name, command string, args []string, env map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.servers[name]; exists {
		return fmt.Errorf("server %s already connected", name)
	}

	// Create process
	cmd := exec.Command(command, args...)
	if env != nil {
		for k, v := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Get stdin/stdout pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	server := &Server{
		Name:    name,
		Command: command,
		Args:    args,
		Env:     env,
		Process: cmd,
		Stdin:   stdin,
		Stdout:  stdout,
		Tools:   make([]Tool, 0),
		Status:  "connected",
	}

	// Initialize server
	if err := server.initialize(); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	// List tools
	tools, err := server.listTools()
	if err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to list tools: %w", err)
	}

	server.Tools = tools
	c.servers[name] = server

	return nil
}

// DisconnectServer disconnects from an MCP server
func (c *Client) DisconnectServer(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	server, exists := c.servers[name]
	if !exists {
		return fmt.Errorf("server %s not found", name)
	}

	server.close()
	delete(c.servers, name)

	return nil
}

// ListServers returns all connected servers
func (c *Client) ListServers() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.servers))
	for name := range c.servers {
		names = append(names, name)
	}

	return names
}

// ListTools lists tools for a server
func (c *Client) ListTools(serverName string) ([]Tool, error) {
	c.mu.RLock()
	server, exists := c.servers[serverName]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server %s not found", serverName)
	}

	return server.Tools, nil
}

// CallTool calls an MCP tool
func (c *Client) CallTool(serverName, toolName string, args map[string]interface{}) (*models.MCPToolResult, error) {
	c.mu.RLock()
	server, exists := c.servers[serverName]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server %s not found", serverName)
	}

	result, err := server.callTool(toolName, args)
	if err != nil {
		return &models.MCPToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &models.MCPToolResult{
		Success: true,
		Result:  result,
	}, nil
}

// GetServerStatus returns server status
func (c *Client) GetServerStatus(name string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	server, exists := c.servers[name]
	if !exists {
		return "", fmt.Errorf("server %s not found", name)
	}

	return server.Status, nil
}

// Cleanup disconnects all servers
func (c *Client) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for name, server := range c.servers {
		server.close()
		delete(c.servers, name)
	}
}

// Server methods

// initialize initializes the MCP server
func (s *Server) initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "agent-workspace",
				"version": "1.0.0",
			},
		},
	}

	_, err := s.sendRequest(req)
	return err
}

// listTools lists available tools
func (s *Server) listTools() ([]Tool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	resp, err := s.sendRequest(req)
	if err != nil {
		return nil, err
	}

	// Parse tools from response
	toolsData, ok := resp.Result["tools"].([]interface{})
	if !ok {
		return []Tool{}, nil
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

	return tools, nil
}

// callTool calls a tool
func (s *Server) callTool(name string, args map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      int(time.Now().Unix()),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": args,
		},
	}

	resp, err := s.sendRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	return resp.Result, nil
}

// sendRequest sends a request and waits for response
func (s *Server) sendRequest(req MCPRequest) (*MCPResponse, error) {
	// Marshal request
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request
	if _, err := s.Stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Read response
	reader := bufio.NewReader(s.Stdout)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var resp MCPResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// close closes the server connection
func (s *Server) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Stdin != nil {
		s.Stdin.Close()
	}

	if s.Stdout != nil {
		s.Stdout.Close()
	}

	if s.Process != nil && s.Process.Process != nil {
		s.Process.Process.Kill()
	}

	s.Status = "disconnected"
}

// Helper functions

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

