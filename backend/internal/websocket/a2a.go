package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"agent-workspace/backend/pkg/jsonrpc"
	"agent-workspace/backend/pkg/models"
	"agent-workspace/backend/pkg/ollama"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/websocket/v3"
	"github.com/google/uuid"
)

// A2AHandler handles A2A protocol WebSocket connections
type A2AHandler struct {
	clients         map[*websocket.Conn]bool
	broadcast       chan *jsonrpc.Response
	register        chan *websocket.Conn
	unregister      chan *websocket.Conn
	mu              sync.RWMutex
	router          *jsonrpc.Router
	ollama          *ollama.Client
	agentController interface{} // Will be *agent.Controller when implemented
}

// NewA2AHandler creates a new A2A WebSocket handler
func NewA2AHandler(agentController interface{}) *A2AHandler {
	h := &A2AHandler{
		clients:         make(map[*websocket.Conn]bool),
		broadcast:       make(chan *jsonrpc.Response, 256),
		register:        make(chan *websocket.Conn),
		unregister:      make(chan *websocket.Conn),
		router:          jsonrpc.NewRouter(),
		ollama:          ollama.NewClient(),
		agentController: agentController,
	}

	// Register A2A protocol methods
	h.registerMethods()

	// Start the hub
	go h.run()

	return h
}

// run handles the A2A WebSocket hub
func (h *A2AHandler) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("A2A client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
			log.Printf("A2A client disconnected. Total clients: %d", len(h.clients))

		case response := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if err := client.WriteJSON(response); err != nil {
					log.Printf("Error broadcasting to A2A client: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			// Send heartbeat
			h.mu.RLock()
			for client := range h.clients {
				if err := client.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					log.Printf("Error sending A2A heartbeat: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// HandleWebSocket handles A2A WebSocket upgrade and messages
func (h *A2AHandler) HandleWebSocket(c fiber.Ctx) error {
	return websocket.New(func(conn *websocket.Conn) {
		// Register client
		h.register <- conn
		defer func() {
			h.unregister <- conn
		}()

		// Handle messages
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("A2A WebSocket error: %v", err)
				}
				break
			}

			// Handle JSON-RPC request
			go h.handleJSONRPC(conn, message)
		}
	})(c)
}

// handleJSONRPC processes JSON-RPC 2.0 requests
func (h *A2AHandler) handleJSONRPC(conn *websocket.Conn, message []byte) {
	// Parse request
	req, err := jsonrpc.UnmarshalRequest(message)
	if err != nil {
		resp := jsonrpc.NewErrorResponse(nil, jsonrpc.ParseError, "parse error", nil)
		respBytes, _ := resp.Marshal()
		conn.WriteMessage(websocket.TextMessage, respBytes)
		return
	}

	// Handle request
	resp := h.router.Handle(req)
	if resp == nil {
		// Notification - no response
		return
	}

	// Send response
	respBytes, err := resp.Marshal()
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

// registerMethods registers A2A protocol methods
func (h *A2AHandler) registerMethods() {
	// agent/getAuthenticatedExtendedCard
	h.router.Register("agent/getAuthenticatedExtendedCard", func(params map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{
			"name":        "Agent Workspace",
			"description": "AI agent with browser automation, terminal control, and MCP integration",
			"version":     "1.0.0",
			"capabilities": map[string]bool{
				"streaming":          true,
				"file_operations":    true,
				"browser_automation": true,
				"memory_graph":       true,
			},
			"skills": []map[string]string{
				{"name": "web_browsing", "description": "Automated web browsing and data extraction"},
				{"name": "code_generation", "description": "Generate and edit code files"},
				{"name": "terminal_control", "description": "Execute terminal commands"},
				{"name": "memory_management", "description": "Store and retrieve from knowledge graph"},
			},
		}, nil
	})

	// message/send
	h.router.Register("message/send", func(params map[string]interface{}) (interface{}, error) {
		messageData, ok := params["message"].(map[string]interface{})
		if !ok {
			return nil, jsonrpc.NewErrorResponse(nil, jsonrpc.InvalidParams, "invalid message format", nil).Error
		}

		role, _ := messageData["role"].(string)
		parts, _ := messageData["parts"].([]interface{})

		if len(parts) == 0 {
			return nil, jsonrpc.NewErrorResponse(nil, jsonrpc.InvalidParams, "no message parts", nil).Error
		}

		// Extract text from parts
		var text string
		for _, part := range parts {
			partMap, ok := part.(map[string]interface{})
			if !ok {
				continue
			}
			if partMap["type"] == "text" {
				text, _ = partMap["text"].(string)
				break
			}
		}

		// Build conversation
		messages := []ollama.ChatMessage{
			{
				Role:    "system",
				Content: "You are an AI agent assistant. Respond concisely and helpfully.",
			},
			{
				Role:    role,
				Content: text,
			},
		}

		// Get response from Ollama
		resp, err := h.ollama.ChatCompletion(messages, 0.7)
		if err != nil {
			return nil, err
		}

		if len(resp.Choices) == 0 {
			return nil, jsonrpc.NewErrorResponse(nil, jsonrpc.InternalError, "no response from model", nil).Error
		}

		return map[string]interface{}{
			"role": "assistant",
			"parts": []map[string]interface{}{
				{
					"type": "text",
					"text": resp.Choices[0].Message.Content,
				},
			},
		}, nil
	})

	// message/stream
	h.router.Register("message/stream", func(params map[string]interface{}) (interface{}, error) {
		taskID, _ := params["task_id"].(string)
		if taskID == "" {
			taskID = uuid.New().String()
		}

		return map[string]interface{}{
			"task_id": taskID,
			"status":  "streaming",
			"message": "Use SSE endpoint /api/tasks/" + taskID + "/stream for streaming",
		}, nil
	})

	// tasks/get
	h.router.Register("tasks/get", func(params map[string]interface{}) (interface{}, error) {
		taskID, _ := params["task_id"].(string)
		if taskID == "" {
			return nil, jsonrpc.NewErrorResponse(nil, jsonrpc.InvalidParams, "task_id required", nil).Error
		}

		// TODO: Implement task retrieval
		return map[string]interface{}{
			"task_id": taskID,
			"status":  "completed",
			"result":  map[string]interface{}{},
		}, nil
	})

	// tasks/list
	h.router.Register("tasks/list", func(params map[string]interface{}) (interface{}, error) {
		// TODO: Implement task listing
		return map[string]interface{}{
			"tasks": []interface{}{},
		}, nil
	})

	// tasks/cancel
	h.router.Register("tasks/cancel", func(params map[string]interface{}) (interface{}, error) {
		taskID, _ := params["task_id"].(string)
		if taskID == "" {
			return nil, jsonrpc.NewErrorResponse(nil, jsonrpc.InvalidParams, "task_id required", nil).Error
		}

		// TODO: Implement task cancellation
		return map[string]interface{}{
			"task_id":   taskID,
			"status":    "cancelled",
			"cancelled": true,
		}, nil
	})

	// browser/navigate
	h.router.Register("browser/navigate", func(params map[string]interface{}) (interface{}, error) {
		url, _ := params["url"].(string)
		if url == "" {
			return nil, jsonrpc.NewErrorResponse(nil, jsonrpc.InvalidParams, "url required", nil).Error
		}

		// TODO: Implement browser navigation
		return map[string]interface{}{
			"success": true,
			"url":     url,
		}, nil
	})

	// browser/click
	h.router.Register("browser/click", func(params map[string]interface{}) (interface{}, error) {
		elementID, _ := params["element_id"].(float64)

		// TODO: Implement browser click
		return map[string]interface{}{
			"success":    true,
			"element_id": int(elementID),
		}, nil
	})

	// terminal/execute
	h.router.Register("terminal/execute", func(params map[string]interface{}) (interface{}, error) {
		command, _ := params["command"].(string)
		if command == "" {
			return nil, jsonrpc.NewErrorResponse(nil, jsonrpc.InvalidParams, "command required", nil).Error
		}

		// TODO: Implement terminal execution
		return map[string]interface{}{
			"success":   true,
			"command":   command,
			"output":    "",
			"exit_code": 0,
		}, nil
	})

	// memory/query
	h.router.Register("memory/query", func(params map[string]interface{}) (interface{}, error) {
		query, _ := params["query"].(string)
		if query == "" {
			return nil, jsonrpc.NewErrorResponse(nil, jsonrpc.InvalidParams, "query required", nil).Error
		}

		// TODO: Implement memory query
		return map[string]interface{}{
			"results": []interface{}{},
		}, nil
	})

	// mcp/listTools
	h.router.Register("mcp/listTools", func(params map[string]interface{}) (interface{}, error) {
		server, _ := params["server"].(string)

		// TODO: Implement MCP tool listing
		return map[string]interface{}{
			"server": server,
			"tools":  []interface{}{},
		}, nil
	})

	// mcp/callTool
	h.router.Register("mcp/callTool", func(params map[string]interface{}) (interface{}, error) {
		server, _ := params["server"].(string)
		tool, _ := params["tool"].(string)
		args, _ := params["args"].(map[string]interface{})

		if server == "" || tool == "" {
			return nil, jsonrpc.NewErrorResponse(nil, jsonrpc.InvalidParams, "server and tool required", nil).Error
		}

		// TODO: Implement MCP tool call
		return map[string]interface{}{
			"success": true,
			"server":  server,
			"tool":    tool,
			"result":  map[string]interface{}{},
		}, nil
	})
}

// BroadcastNotification broadcasts a JSON-RPC notification to all clients
func (h *A2AHandler) BroadcastNotification(method string, params map[string]interface{}) {
	notification := jsonrpc.NewNotification(method, params)
	notificationBytes, err := notification.Marshal()
	if err != nil {
		log.Printf("Error marshaling notification: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if err := client.WriteMessage(websocket.TextMessage, notificationBytes); err != nil {
			log.Printf("Error sending notification: %v", err)
		}
	}
}

// GetClientCount returns the number of connected A2A clients
func (h *A2AHandler) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

