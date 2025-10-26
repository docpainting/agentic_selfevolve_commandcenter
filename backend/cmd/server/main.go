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

	"agent-workspace/backend/internal/agent"
	"agent-workspace/backend/internal/browser"
	"agent-workspace/backend/internal/memory"
	"agent-workspace/backend/internal/mcp"
	"agent-workspace/backend/internal/terminal"
	"agent-workspace/backend/internal/watchdog"
	"agent-workspace/backend/internal/websocket"
	"agent-workspace/backend/pkg/ollama"
)

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
	// Validate environment
	if err := validateEnv(); err != nil {
		log.Fatal(err)
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Agentic Command Center v1.0",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Initialize Ollama client
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}
	ollamaClient := ollama.NewClient()

	// Initialize embeddings
	embeddingFunc := memory.NewOllamaEmbedding(ollamaClient, "nomic-embed-text:v1.5")

	// Initialize long-term memory
	longTerm, err := memory.NewLongTermMemory(&memory.LongTermConfig{
		Neo4jURI:      os.Getenv("NEO4J_URI"),
		Neo4jUser:     os.Getenv("NEO4J_USER"),
		Neo4jPassword: os.Getenv("NEO4J_PASSWORD"),
		ChromemPath:   "./data/chromem.db",
		BoltPath:      "./data/bolt.db",
		Embedding:     embeddingFunc,
	})
	if err != nil {
		log.Fatal("Failed to initialize long-term memory:", err)
	}
	log.Println("✓ Long-term memory initialized")

	// Initialize short-term memory
	shortTerm := memory.NewShortTermMemory()
	log.Println("✓ Short-term memory initialized")

	// Combine memory system
	memorySystem := &memory.System{
		LongTerm:  longTerm,
		ShortTerm: shortTerm,
	}

	// Initialize browser manager
	browserMgr := browser.NewManager(&browser.Config{
		Headless: true,
		Timeout:  time.Second * 30,
	})
	log.Println("✓ Browser manager initialized")

	// Initialize terminal manager
	terminalMgr := terminal.NewManager(&terminal.Config{
		DefaultShell: "/bin/bash",
		MaxSessions:  10,
	})
	log.Println("✓ Terminal manager initialized")

	// Initialize MCP client
	mcpConfigPath := os.Getenv("MCP_CONFIG_PATH")
	if mcpConfigPath == "" {
		mcpConfigPath = "./mcp-config.json"
	}
	mcpClient := mcp.NewClient(&mcp.Config{
		ConfigPath: mcpConfigPath,
	})
	log.Println("✓ MCP client initialized")

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

	// Initialize agent controller
	agentController := agent.NewController(&agent.Config{
		OllamaClient:    ollamaClient,
		MemorySystem:    memorySystem,
		BrowserManager:  browserMgr,
		TerminalManager: terminalMgr,
		MCPClient:       mcpClient,
		Watchdog:        watchdogSvc,
	})
	log.Println("✓ Agent controller initialized")

	// Initialize EvoX adapter
	evoxAdapter := agent.NewEvoXAdapter(ollamaClient, &agent.EvoXConfig{
		Enabled:            true,
		LearningRate:       0.1,
		MinConfidence:      0.7,
		EvolutionThreshold: 10,
	})
	log.Println("✓ EvoX adapter initialized")

	// Routes
	api := app.Group("/api")

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"timestamp": time.Now().Format(time.RFC3339),
			"services": fiber.Map{
				"neo4j":    longTerm.HealthCheck(),
				"ollama":   ollamaClient.HealthCheck(),
				"browser":  browserMgr.IsHealthy(),
				"terminal": terminalMgr.IsHealthy(),
				"mcp":      mcpClient.IsHealthy(),
				"watchdog": watchdogSvc.IsRunning(),
			},
		})
	})

	// Agent routes
	api.Post("/agent/initialize", agentController.Initialize)
	api.Post("/agent/command", agentController.ExecuteCommand)
	api.Get("/agent/status", agentController.GetStatus)
	api.Post("/agent/pause", agentController.Pause)
	api.Post("/agent/resume", agentController.Resume)

	// EvoX routes
	api.Post("/evox/workflow", evoxAdapter.GenerateWorkflowHandler)
	api.Post("/evox/action", evoxAdapter.ParseActionHandler)
	api.Post("/evox/evaluate", evoxAdapter.EvaluateHandler)
	api.Get("/evox/patterns", evoxAdapter.GetPatternsHandler)

	// Watchdog routes
	api.Get("/watchdog/alerts", watchdogSvc.GetAlertsHandler)
	api.Post("/watchdog/alerts/:id/acknowledge", watchdogSvc.AcknowledgeAlertHandler)
	api.Get("/watchdog/patterns", watchdogSvc.GetPatternsHandler)
	api.Get("/watchdog/components", watchdogSvc.GetComponentsHandler)
	api.Post("/watchdog/components/:id/approve", watchdogSvc.ApproveComponentHandler)

	// Memory routes
	api.Post("/memory/store", func(c *fiber.Ctx) error {
		var req struct {
			Type    string                 `json:"type"`
			Content string                 `json:"content"`
			Context map[string]interface{} `json:"context"`
		}
		if err := c.BodyParser(&req); err != nil {
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

	api.Post("/memory/query", func(c *fiber.Ctx) error {
		var req struct {
			Query string `json:"query"`
			Limit int    `json:"limit"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// Query memory system
		results := []map[string]interface{}{}
		return c.JSON(fiber.Map{"results": results})
	})

	// File operations routes
	api.Get("/files/tree", func(c *fiber.Ctx) error {
		path := c.Query("path", ".")
		// Return file tree
		return c.JSON(fiber.Map{"path": path, "files": []interface{}{}})
	})

	api.Get("/files/content", func(c *fiber.Ctx) error {
		path := c.Query("path")
		if path == "" {
			return c.Status(400).JSON(fiber.Map{"error": "path required"})
		}
		// Return file content
		return c.JSON(fiber.Map{"path": path, "content": ""})
	})

	// WebSocket routes
	app.Get("/ws/chat", websocket.HandleChatWebSocket(agentController))
	app.Get("/ws/a2a", websocket.HandleA2AWebSocket(agentController))

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n🛑 Shutting down gracefully...")

		// Cleanup
		log.Println("  → Stopping watchdog...")
		watchdogSvc.Stop()

		log.Println("  → Closing browser...")
		browserMgr.Close()

		log.Println("  → Closing terminals...")
		terminalMgr.CloseAll()

		log.Println("  → Disconnecting MCP...")
		mcpClient.DisconnectAll()

		log.Println("  → Closing memory...")
		longTerm.Close()

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
	log.Println("Press Ctrl+C to stop\n")

	log.Fatal(app.Listen(":" + port))
}

