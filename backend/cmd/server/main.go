package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joho/godotenv"

	"agent-workspace/backend/internal/browser"
	"agent-workspace/backend/internal/mcp"
	"agent-workspace/backend/internal/memory"
	"agent-workspace/backend/internal/terminal"
	"agent-workspace/backend/internal/watchdog"
	"agent-workspace/backend/internal/websocket"
	"agent-workspace/backend/pkg/ollama"
)

type FileNode struct {
	Name     string     `json:"name"`
	Type     string     `json:"type"` // "file" or "directory"
	Path     string     `json:"path"`
	Children []FileNode `json:"children,omitempty"`
}

func buildFileTree(rootPath string) (*FileNode, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}

	node := &FileNode{
		Name: info.Name(),
		Path: rootPath,
	}

	if info.IsDir() {
		node.Type = "directory"

		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return nil, err
		}

		node.Children = make([]FileNode, 0)
		for _, entry := range entries {
			// Skip hidden files and common ignore patterns
			if entry.Name()[0] == '.' || entry.Name() == "node_modules" || entry.Name() == "vendor" {
				continue
			}

			childPath := rootPath + "/" + entry.Name()
			childNode, err := buildFileTree(childPath)
			if err != nil {
				continue // Skip files we can't read
			}
			node.Children = append(node.Children, *childNode)
		}
	} else {
		node.Type = "file"
	}

	return node, nil
}

func validateEnv() error {
	required := []string{
		"NEO4J_URI",
		"NEO4J_USER",
		"NEO4J_PASSWORD",
	}

	for _, key := range required {
		if os.Getenv(key) == "" {
			return fmt.Errorf("required environment variable %s not set", key)
		}
	}
	return nil
}

func main() {
	// Load .env file - try multiple locations
	envPaths := []string{
		"../.env",    // When running from backend/
		".env",       // When running from project root
		"../../.env", // When running from backend/cmd/server/
	}

	envLoaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("Loaded environment from %s", path)
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		log.Printf("Warning: No .env file found, using system environment")
	}

	// Validate environment
	log.Println("→ Validating environment...")
	if err := validateEnv(); err != nil {
		log.Fatal(err)
	}
	log.Println("✓ Environment validated")

	// Initialize Fiber app
	log.Println("→ Initializing Fiber app...")
	app := fiber.New(fiber.Config{
		AppName: "Agentic Command Center v1.0",
	})
	log.Println("✓ Fiber app initialized")

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	// Initialize Ollama client
	log.Println("→ Initializing Ollama client...")
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}
	ollamaClient := ollama.NewClient()
	log.Println("✓ Ollama client initialized")

	// Initialize long-term memory (skip for now - hangs on Neo4j connection)
	log.Println("→ Skipping long-term memory (Neo4j connection hangs)...")
	var longTerm *memory.LongTermMemory = nil
	log.Println("✓ Continuing without long-term memory")

	// Initialize short-term memory
	shortTerm := memory.NewShortTermMemory()
	log.Println("✓ Short-term memory initialized")

	// Combine memory system
	memorySystem := memory.NewSystem(longTerm, shortTerm)
	log.Println("✓ Memory system combined")

	// Initialize terminal manager
	terminalMgr := terminal.NewManager(&terminal.Config{
		DefaultShell: "/bin/bash",
		MaxSessions:  10,
	})
	log.Println("✓ Terminal manager initialized")

	// Initialize MCP client (for other MCP servers like filesystem, memory, etc.)
	mcpClient := mcp.NewClient(&mcp.Config{
		ConfigPath: "./backend/mcp-config.json",
	})
	log.Println("✓ MCP client initialized")

	// Initialize ChromeDP browser manager (Go-native browser automation)
	log.Println("→ Starting ChromeDP browser...")
	browserMgr := browser.NewManager(shortTerm)
	if err := browserMgr.Initialize(); err != nil {
		log.Fatalf("Failed to start browser: %v", err)
	}
	log.Println("✓ ChromeDP browser started")

	// Initialize watchdog
	watchdogSvc := watchdog.NewWatchdog(&watchdog.Config{
		Enabled:        true,
		ScanInterval:   time.Second * 30,
		MinConfidence:  0.7,
		AlertThreshold: watchdog.SeverityWarning,
		Memory:         memorySystem,
	})
	watchdogSvc.Start()
	log.Println("✓ Watchdog started")

	// Routes
	api := app.Group("/api")

	// Health check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
			"services": fiber.Map{
				"memory":   longTerm != nil,
				"ollama":   ollamaClient != nil,
				"browser":  true,
				"terminal": terminalMgr.IsHealthy(),
				"mcp":      mcpClient.IsHealthy(),
				"watchdog": watchdogSvc.IsRunning(),
			},
		})
	})

	// TODO: Agent, EvoX, and Watchdog routes will be added when implementations are ready

	// Memory routes
	api.Post("/memory/store", func(c fiber.Ctx) error {
		var req struct {
			Type    string                 `json:"type"`
			Content string                 `json:"content"`
			Context map[string]interface{} `json:"context"`
		}
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		switch req.Type {
		case "conversation":
			return c.JSON(fiber.Map{"success": true})
		case "code":
			return c.JSON(fiber.Map{"success": true})
		default:
			return c.Status(400).JSON(fiber.Map{"error": "invalid type"})
		}
	})

	api.Post("/memory/query", func(c fiber.Ctx) error {
		var req struct {
			Query string `json:"query"`
			Limit int    `json:"limit"`
		}
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// Query memory system
		results := []map[string]interface{}{}
		return c.JSON(fiber.Map{"results": results})
	})

	// File operations routes
	api.Get("/files/tree", func(c fiber.Ctx) error {
		path := c.Query("path", ".")

		// Build file tree
		tree, err := buildFileTree(path)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(tree)
	})

	api.Get("/files/content", func(c fiber.Ctx) error {
		path := c.Query("path")
		if path == "" {
			return c.Status(400).JSON(fiber.Map{"error": "path required"})
		}
		// Return file content
		return c.JSON(fiber.Map{"path": path, "content": ""})
	})

	// WebSocket routes
	app.Get("/ws/chat", websocket.HandleChatWebSocket(nil))
	app.Get("/ws/browser", websocket.HandleA2AWebSocket(mcpClient, browserMgr, terminalMgr)) // Browser + Terminal automation with JSON-RPC 2.0
	app.Get("/ws/a2a", websocket.HandleA2AWebSocket(mcpClient, browserMgr, terminalMgr)) // A2A protocol with browser + terminal
	log.Println("✓ A2A WebSocket registered with browser and terminal support")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n🛑 Shutting down gracefully...")

		// Cleanup
		log.Println("  → Stopping watchdog...")
		watchdogSvc.Stop()

		log.Println("  → Closing terminals...")
		terminalMgr.CloseAll()

		log.Println("  → Disconnecting MCP...")
		mcpClient.DisconnectAll()

		if longTerm != nil {
			log.Println("  → Closing memory...")
			longTerm.Close()
		}

		log.Println("  → Stopping server...")
		app.Shutdown()

		log.Println("✓ Shutdown complete")
		os.Exit(0)
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("\n🚀 Agentic Self-Evolving Command Center")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("Server: http://localhost:%s\n", port)
	log.Printf("Health: http://localhost:%s/health\n", port)
	log.Printf("WebSocket Chat: ws://localhost:%s/ws/chat\n", port)
	log.Printf("WebSocket A2A: ws://localhost:%s/ws/a2a\n", port)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("Press Ctrl+C to stop")

	log.Fatal(app.Listen(":" + port))
}
