package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"mcp-dynamic-thinking/internal/act"
	"mcp-dynamic-thinking/internal/memory"
	"mcp-dynamic-thinking/internal/perceive"
	"mcp-dynamic-thinking/internal/reason"
	"mcp-dynamic-thinking/internal/reflect"
)

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

// Server represents the MCP server
type Server struct {
	perceiver *perceive.Perceiver
	reasoner  *reason.Reasoner
	actor     *act.Actor
	reflector *reflect.Reflector
	memory    *memory.MemoryManager
}

func main() {
	// Initialize server
	server := &Server{
		perceiver: perceive.NewPerceiver(),
		reasoner:  reason.NewReasoner(),
		actor:     act.NewActor(),
		reflector: reflect.NewReflector(),
		memory:    memory.NewMemoryManager(),
	}

	// Send initialization response
	initResp := MCPResponse{
		JSONRPC: "2.0",
		ID:      0,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "mcp-dynamic-thinking",
				"version": "1.0.0",
			},
		},
	}

	if err := sendResponse(initResp); err != nil {
		log.Fatalf("Failed to send init response: %v", err)
	}

	// Start request loop
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("Failed to parse request: %v", err)
			continue
		}

		// Handle request
		resp := server.handleRequest(req)

		// Send response
		if err := sendResponse(resp); err != nil {
			log.Printf("Failed to send response: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
	}
}

func (s *Server) handleRequest(req MCPRequest) MCPResponse {
	ctx := context.Background()

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

func (s *Server) handleInitialize(req MCPRequest) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "mcp-dynamic-thinking",
				"version": "1.0.0",
			},
		},
	}
}

func (s *Server) handleToolsList(req MCPRequest) MCPResponse {
	tools := []map[string]interface{}{
		{
			"name":        "perceive",
			"description": "Capture and analyze environment state",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]string{"type": "string"},
					"goal":    map[string]string{"type": "string"},
				},
				"required": []string{"task_id", "goal"},
			},
		},
		{
			"name":        "reason",
			"description": "Generate and evaluate reasoning branches",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id":       map[string]string{"type": "string"},
					"perception_id": map[string]string{"type": "string"},
					"num_branches":  map[string]string{"type": "number"},
				},
				"required": []string{"task_id", "perception_id"},
			},
		},
		{
			"name":        "act",
			"description": "Execute action plan with monitoring",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id":     map[string]string{"type": "string"},
					"action_plan": map[string]string{"type": "object"},
				},
				"required": []string{"task_id", "action_plan"},
			},
		},
		{
			"name":        "reflect",
			"description": "Analyze results and evolve strategies",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id":      map[string]string{"type": "string"},
					"execution_id": map[string]string{"type": "string"},
				},
				"required": []string{"task_id", "execution_id"},
			},
		},
		{
			"name":        "get_short_term_memory",
			"description": "Retrieve task context from short-term memory",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]string{"type": "string"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			"name":        "clear_short_term_memory",
			"description": "Clear task memory after completion",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id":              map[string]string{"type": "string"},
					"archive_to_long_term": map[string]string{"type": "boolean"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			"name":        "query_strategies",
			"description": "Find relevant strategies from Neo4j",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]string{"type": "string"},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "get_execution_trace",
			"description": "Export training data for completed task",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]string{"type": "string"},
				},
				"required": []string{"task_id"},
			},
		},
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *Server) handleToolsCall(ctx context.Context, req MCPRequest) MCPResponse {
	params := req.Params
	toolName, _ := params["name"].(string)
	args, _ := params["arguments"].(map[string]interface{})

	var result map[string]interface{}
	var err error

	switch toolName {
	case "perceive":
		result, err = s.perceiver.Perceive(ctx, args)
	case "reason":
		result, err = s.reasoner.Reason(ctx, args)
	case "act":
		result, err = s.actor.Act(ctx, args)
	case "reflect":
		result, err = s.reflector.Reflect(ctx, args)
	case "get_short_term_memory":
		result, err = s.memory.GetShortTermMemory(ctx, args)
	case "clear_short_term_memory":
		result, err = s.memory.ClearShortTermMemory(ctx, args)
	case "query_strategies":
		result, err = s.memory.QueryStrategies(ctx, args)
	case "get_execution_trace":
		result, err = s.memory.GetExecutionTrace(ctx, args)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: fmt.Sprintf("Tool not found: %s", toolName),
			},
		}
	}

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: err.Error(),
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func sendResponse(resp MCPResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

