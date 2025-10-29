package websocket

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"agent-workspace/backend/internal/browser"
	"agent-workspace/backend/internal/mcp"
	"agent-workspace/backend/internal/terminal"
	"agent-workspace/backend/pkg/jsonrpc"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

// A2AHandler handles A2A protocol WebSocket connections with JSON-RPC 2.0
type A2AHandler struct {
	clients      map[*websocket.Conn]bool
	broadcast    chan *jsonrpc.Response
	register     chan *websocket.Conn
	unregister   chan *websocket.Conn
	mu           sync.RWMutex
	router       *jsonrpc.Router
	mcpClient    *mcp.Client
	browserMgr   *browser.Manager
	terminalMgr  *terminal.Manager
}

// NewA2AHandler creates a new A2A WebSocket handler
func NewA2AHandler(mcpClient *mcp.Client, browserMgr *browser.Manager, terminalMgr *terminal.Manager) *A2AHandler {
	h := &A2AHandler{
		clients:      make(map[*websocket.Conn]bool),
		broadcast:    make(chan *jsonrpc.Response, 256),
		register:     make(chan *websocket.Conn),
		unregister:   make(chan *websocket.Conn),
		router:       jsonrpc.NewRouter(),
		mcpClient:    mcpClient,
		browserMgr:   browserMgr,
		terminalMgr:  terminalMgr,
	}

	// Register JSON-RPC methods
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

// registerMethods registers JSON-RPC methods for browser automation
func (h *A2AHandler) registerMethods() {
	// Browser navigation - frontend calls "browser/navigate"
	h.router.Register("browser/navigate", func(params map[string]interface{}) (interface{}, error) {
		url, ok := params["url"].(string)
		if !ok {
			return nil, fmt.Errorf("url parameter required")
		}
		if err := h.browserMgr.Navigate(url); err != nil {
			return nil, fmt.Errorf("navigation failed: %w", err)
		}
		return map[string]interface{}{"success": true, "url": url}, nil
	})

	// Get DOM - frontend calls "browser/getDOM"
	h.router.Register("browser/getDOM", func(params map[string]interface{}) (interface{}, error) {
		// Get page info
		title, _ := h.browserMgr.GetPageTitle()
		html, _ := h.browserMgr.GetPageHTML()
		if len(html) > 5000 {
			html = html[:5000]
		}
		elements := h.browserMgr.GetElements()
		
		// Convert elements to interface format
		interfaceElements := make([]map[string]interface{}, 0, len(elements))
		for _, elem := range elements {
			interfaceElements = append(interfaceElements, map[string]interface{}{
				"id":   elem.ID,
				"tag":  elem.Tag,
				"text": elem.Text,
			})
		}
		
		// Capture screenshot with numbered overlays
		screenshot, err := h.browserMgr.GetScreenshotWithOverlays("")
		var screenshotDataURL string
		if err != nil {
			log.Printf("[Browser] Screenshot error: %v", err)
		} else if len(screenshot) > 0 {
			// Format as data URL for img src
			screenshotDataURL = "data:image/png;base64," + base64.StdEncoding.EncodeToString(screenshot)
			log.Printf("[Browser] Screenshot captured: %d bytes", len(screenshot))
		} else {
			log.Printf("[Browser] Screenshot empty")
		}
		
		return map[string]interface{}{
			"title":               title,
			"current_url":         h.browserMgr.GetCurrentURL(),
			"html":                html,
			"interactive_elements": interfaceElements,
			"element_count":       len(interfaceElements),
			"screenshot":          screenshotDataURL,
		}, nil
	})

	// Click element - frontend calls "browser/click"
	h.router.Register("browser/click", func(params map[string]interface{}) (interface{}, error) {
		selector, ok := params["selector"].(string)
		if !ok {
			return nil, fmt.Errorf("selector parameter required")
		}
		if err := h.browserMgr.ClickBySelector(selector); err != nil {
			return nil, fmt.Errorf("click failed: %w", err)
		}
		return map[string]interface{}{"success": true, "selector": selector}, nil
	})

	// Type text - frontend calls "browser/type"
	h.router.Register("browser/type", func(params map[string]interface{}) (interface{}, error) {
		selector, ok := params["selector"].(string)
		if !ok {
			return nil, fmt.Errorf("selector parameter required")
		}
		text, ok := params["text"].(string)
		if !ok {
			return nil, fmt.Errorf("text parameter required")
		}
		if err := h.browserMgr.TypeBySelector(selector, text); err != nil {
			return nil, fmt.Errorf("type failed: %w", err)
		}
		return map[string]interface{}{"success": true}, nil
	})

	// Execute script - frontend calls "browser/executeScript"
	h.router.Register("browser/executeScript", func(params map[string]interface{}) (interface{}, error) {
		script, ok := params["script"].(string)
		if !ok {
			return nil, fmt.Errorf("script parameter required")
		}
		result, err := h.browserMgr.ExecuteScript(script)
		if err != nil {
			return nil, fmt.Errorf("script execution failed: %w", err)
		}
		return map[string]interface{}{"success": true, "result": result}, nil
	})

	// Take screenshot - frontend calls "browser/screenshot"
	h.router.Register("browser/screenshot", func(params map[string]interface{}) (interface{}, error) {
		screenshot, err := h.browserMgr.CaptureScreenshot("")
		if err != nil {
			return nil, fmt.Errorf("screenshot failed: %w", err)
		}
		return map[string]interface{}{"success": true, "screenshot": screenshot}, nil
	})

	// Get accessibility tree - frontend calls "browser/getAccessibilityTree"
	h.router.Register("browser/getAccessibilityTree", func(params map[string]interface{}) (interface{}, error) {
		elements := h.browserMgr.GetElements()
		interfaceElements := make([]map[string]interface{}, 0, len(elements))
		for _, elem := range elements {
			interfaceElements = append(interfaceElements, map[string]interface{}{
				"id":   elem.ID,
				"tag":  elem.Tag,
				"text": elem.Text,
			})
		}
		return map[string]interface{}{"elements": interfaceElements}, nil
	})

	// Terminal methods - agent calls via A2A
	
	// Execute command - agent calls "terminal/execute"
	h.router.Register("terminal/execute", func(params map[string]interface{}) (interface{}, error) {
		command, ok := params["command"].(string)
		if !ok {
			return nil, fmt.Errorf("command parameter required")
		}
		
		// Execute command directly with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		cmd := exec.CommandContext(ctx, "bash", "-c", command)
		output, err := cmd.CombinedOutput()
		
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				return nil, fmt.Errorf("command execution failed: %w", err)
			}
		}
		
		return map[string]interface{}{
			"success":   exitCode == 0,
			"output":    string(output),
			"exit_code": exitCode,
			"command":   command,
		}, nil
	})
}

var a2aUpgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true // Allow all origins in development
	},
}

// HandleWebSocket handles A2A WebSocket upgrade and messages
func (h *A2AHandler) HandleWebSocket(c fiber.Ctx) error {
	return a2aUpgrader.Upgrade(c.RequestCtx(), func(conn *websocket.Conn) {
		defer func() {
			h.unregister <- conn
			conn.Close()
		}()

		// Register client
		h.register <- conn

		// Handle messages
		for {
			var req jsonrpc.Request
			if err := conn.ReadJSON(&req); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("A2A WebSocket error: %v", err)
				}
				break
			}

			// Handle JSON-RPC request using router
			response := h.router.Handle(&req)
			
			// Don't send response for notifications
			if response == nil {
				continue
			}

			// Send response
			if err := conn.WriteJSON(response); err != nil {
				log.Printf("Error sending A2A response: %v", err)
				break
			}
		}
	})
}

// HandleA2AWebSocket creates and returns an A2A WebSocket handler
func HandleA2AWebSocket(mcpClient *mcp.Client, browserMgr *browser.Manager, terminalMgr *terminal.Manager) fiber.Handler {
	handler := NewA2AHandler(mcpClient, browserMgr, terminalMgr)
	return handler.HandleWebSocket
}
