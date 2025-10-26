package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"agent-workspace/backend/pkg/models"
	"agent-workspace/backend/pkg/ollama"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/websocket/v3"
	"github.com/google/uuid"
)

// Handler handles WebSocket chat connections
type Handler struct {
	clients         map[*websocket.Conn]bool
	broadcast       chan models.Message
	register        chan *websocket.Conn
	unregister      chan *websocket.Conn
	mu              sync.RWMutex
	ollama          *ollama.Client
	agentController interface{} // Will be *agent.Controller when implemented
}

// NewHandler creates a new WebSocket handler
func NewHandler(agentController interface{}) *Handler {
	h := &Handler{
		clients:         make(map[*websocket.Conn]bool),
		broadcast:       make(chan models.Message, 256),
		register:        make(chan *websocket.Conn),
		unregister:      make(chan *websocket.Conn),
		ollama:          ollama.NewClient(),
		agentController: agentController,
	}

	// Start the hub
	go h.run()

	return h
}

// run handles the WebSocket hub
func (h *Handler) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if err := client.WriteJSON(message); err != nil {
					log.Printf("Error broadcasting to client: %v", err)
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
					log.Printf("Error sending heartbeat: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// HandleWebSocket handles WebSocket upgrade and messages
func (h *Handler) HandleWebSocket(c fiber.Ctx) error {
	return websocket.New(func(conn *websocket.Conn) {
		// Register client
		h.register <- conn
		defer func() {
			h.unregister <- conn
		}()

		// Send welcome message
		welcomeMsg := models.Message{
			ID:        uuid.New().String(),
			Type:      "system_event",
			Timestamp: time.Now().Format(time.RFC3339),
			Source:    "system",
			Payload: map[string]interface{}{
				"event":   "connected",
				"message": "Connected to Agent Workspace",
			},
		}
		conn.WriteJSON(welcomeMsg)

		// Handle messages
		for {
			var msg models.Message
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}

			// Handle different message types
			go h.handleMessage(conn, msg)
		}
	})(c)
}

// handleMessage processes incoming messages
func (h *Handler) handleMessage(conn *websocket.Conn, msg models.Message) {
	switch msg.Type {
	case "user_command":
		h.handleUserCommand(conn, msg)
	case "heartbeat":
		// Respond to heartbeat
		h.sendToClient(conn, models.Message{
			ID:        uuid.New().String(),
			Type:      "heartbeat_ack",
			Timestamp: time.Now().Format(time.RFC3339),
			Source:    "system",
		})
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleUserCommand processes user commands
func (h *Handler) handleUserCommand(conn *websocket.Conn, msg models.Message) {
	command, ok := msg.Payload["command"].(string)
	if !ok {
		h.sendError(conn, "Invalid command format")
		return
	}

	// Send thinking status
	h.sendToClient(conn, models.Message{
		ID:        uuid.New().String(),
		Type:      "agent_status",
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    "agent",
		Payload: map[string]interface{}{
			"state":   "thinking",
			"message": "Processing your request...",
		},
	})

	// Build conversation history
	messages := []ollama.ChatMessage{
		{
			Role:    "system",
			Content: "You are an AI agent assistant with access to browser automation, terminal control, and file operations. Help the user accomplish their tasks efficiently.",
		},
		{
			Role:    "user",
			Content: command,
		},
	}

	// Stream response from Ollama
	responseID := uuid.New().String()
	fullResponse := ""

	err := h.ollama.ChatCompletionStream(messages, 0.7, func(chunk string) error {
		fullResponse += chunk

		// Send chunk to client
		return h.sendToClient(conn, models.Message{
			ID:        responseID,
			Type:      "agent_response_chunk",
			Timestamp: time.Now().Format(time.RFC3339),
			Source:    "agent",
			Payload: map[string]interface{}{
				"chunk":    chunk,
				"complete": false,
			},
		})
	})

	if err != nil {
		log.Printf("Error streaming from Ollama: %v", err)
		h.sendError(conn, "Failed to generate response")
		return
	}

	// Send completion
	h.sendToClient(conn, models.Message{
		ID:        responseID,
		Type:      "agent_response_complete",
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    "agent",
		Payload: map[string]interface{}{
			"response": fullResponse,
			"complete": true,
		},
	})

	// Send idle status
	h.sendToClient(conn, models.Message{
		ID:        uuid.New().String(),
		Type:      "agent_status",
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    "agent",
		Payload: map[string]interface{}{
			"state": "idle",
		},
	})
}

// sendToClient sends a message to a specific client
func (h *Handler) sendToClient(conn *websocket.Conn, msg models.Message) error {
	return conn.WriteJSON(msg)
}

// sendError sends an error message to a client
func (h *Handler) sendError(conn *websocket.Conn, errMsg string) {
	h.sendToClient(conn, models.Message{
		ID:        uuid.New().String(),
		Type:      "error",
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    "system",
		Payload: map[string]interface{}{
			"error": errMsg,
		},
	})
}

// BroadcastMessage broadcasts a message to all connected clients
func (h *Handler) BroadcastMessage(msg models.Message) {
	h.broadcast <- msg
}

// BroadcastTerminalOutput broadcasts terminal output
func (h *Handler) BroadcastTerminalOutput(output string) {
	msg := models.Message{
		ID:        uuid.New().String(),
		Type:      "terminal_output",
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    "terminal",
		Payload: map[string]interface{}{
			"output": output,
		},
	}
	h.broadcast <- msg
}

// BroadcastBrowserUpdate broadcasts browser state update
func (h *Handler) BroadcastBrowserUpdate(screenshot string, elements []models.BrowserElement) {
	msg := models.Message{
		ID:        uuid.New().String(),
		Type:      "browser_update",
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    "browser",
		Payload: map[string]interface{}{
			"screenshot": screenshot,
			"elements":   elements,
		},
	}
	h.broadcast <- msg
}

// BroadcastWatchdogAlert broadcasts a watchdog alert
func (h *Handler) BroadcastWatchdogAlert(alertType, title, message string) {
	msg := models.Message{
		ID:        uuid.New().String(),
		Type:      "watchdog_alert",
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    "watchdog",
		Payload: map[string]interface{}{
			"alert_type": alertType,
			"title":      title,
			"message":    message,
		},
	}
	h.broadcast <- msg
}

// GetClientCount returns the number of connected clients
func (h *Handler) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

