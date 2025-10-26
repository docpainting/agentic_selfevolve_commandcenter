package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agent-workspace/backend/internal/agent"
	"agent-workspace/backend/internal/browser"
	"agent-workspace/backend/internal/memory"
	"agent-workspace/backend/internal/mcp"
	"agent-workspace/backend/internal/terminal"
	"agent-workspace/backend/internal/watchdog"
	"agent-workspace/backend/internal/websocket"
	"agent-workspace/backend/pkg/models"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

type Server struct {
	app             *fiber.App
	wsHandler       *websocket.Handler
	a2aHandler      *websocket.A2AHandler
	agentController *agent.Controller
	browserMgr      *browser.Manager
	terminalMgr     *terminal.Manager
	mcpClient       *mcp.Client
	longTermMem     *memory.LongTermMemory
	shortTermMem    *memory.ShortTermMemory
	watchdog        *watchdog.Watchdog
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Create server
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer server.Cleanup()

	// Setup routes
	server.SetupRoutes()

	// Start server in goroutine
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		log.Printf("🚀 Starting Agent Workspace on :%s", port)
		if err := server.app.Listen(fmt.Sprintf(":%s", port)); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Initialize terminal (starts Neo4j, LightRAG, watchdog)
	go func() {
		time.Sleep(2 * time.Second) // Wait for server to start
		if err := server.InitializeTerminal(); err != nil {
			log.Printf("Warning: Failed to initialize terminal: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func NewServer() (*Server, error) {
	// Create Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Initialize memory systems
	longTermMem, err := memory.NewLongTermMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize long-term memory: %w", err)
	}

	shortTermMem := memory.NewShortTermMemory()

	// Initialize managers
	browserMgr := browser.NewManager(shortTermMem)
	terminalMgr := terminal.NewManager()
	mcpClient := mcp.NewClient()

	// Initialize watchdog
	wdog := watchdog.NewWatchdog(longTermMem)

	// Initialize agent controller
	agentController := agent.NewController(
		longTermMem,
		shortTermMem,
		browserMgr,
		terminalMgr,
		mcpClient,
		wdog,
	)

	// Initialize WebSocket handlers
	wsHandler := websocket.NewHandler(agentController)
	a2aHandler := websocket.NewA2AHandler(agentController)

	return &Server{
		app:             app,
		wsHandler:       wsHandler,
		a2aHandler:      a2aHandler,
		agentController: agentController,
		browserMgr:      browserMgr,
		terminalMgr:     terminalMgr,
		mcpClient:       mcpClient,
		longTermMem:     longTermMem,
		shortTermMem:    shortTermMem,
		watchdog:        wdog,
	}, nil
}

func (s *Server) SetupRoutes() {
	// CORS middleware
	s.app.Use(func(c fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Set("Access-Control-Allow-Credentials", "true")
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusOK)
		}
		return c.Next()
	})

	// Health check
	s.app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// WebSocket endpoints
	s.app.Get("/ws/chat", s.wsHandler.HandleWebSocket)
	s.app.Get("/ws/a2a", s.a2aHandler.HandleWebSocket)

	// Agent Card (A2A protocol)
	s.app.Get("/.well-known/agent.json", s.handleAgentCard)

	// API routes
	api := s.app.Group("/api")

	// Agent control
	api.Post("/agent/initialize", s.handleAgentInitialize)
	api.Post("/agent/command", s.handleAgentCommand)
	api.Get("/agent/status", s.handleAgentStatus)
	api.Post("/agent/pause", s.handleAgentPause)
	api.Post("/agent/resume", s.handleAgentResume)

	// File operations
	api.Get("/files/tree", s.handleFileTree)
	api.Get("/files/content", s.handleFileContent)
	api.Put("/files/content", s.handleFileWrite)
	api.Post("/files/diff", s.handleFileDiff)
	api.Get("/files/search", s.handleFileSearch)

	// Memory system
	api.Post("/memory/query", s.handleMemoryQuery)
	api.Post("/memory/store", s.handleMemoryStore)
	api.Get("/memory/context", s.handleMemoryContext)
	api.Post("/memory/vector-search", s.handleVectorSearch)

	// OpenEvolve
	api.Get("/openevolve/status", s.handleOpenEvolveStatus)
	api.Post("/openevolve/proposal", s.handleOpenEvolveProposal)
	api.Put("/openevolve/rewards", s.handleOpenEvolveRewards)
	api.Get("/openevolve/metrics", s.handleOpenEvolveMetrics)

	// Vision system
	api.Post("/vision/analyze", s.handleVisionAnalyze)
	api.Get("/vision/state", s.handleVisionState)
	api.Post("/vision/context", s.handleVisionContext)

	// Static files (for frontend build)
	s.app.Static("/", "./public")
}

func (s *Server) handleAgentCard(c fiber.Ctx) error {
	card := models.AgentCard{
		Name:        "Agent Workspace",
		Description: "AI agent with browser automation, terminal control, and MCP integration",
		Version:     "1.0.0",
		Capabilities: map[string]bool{
			"streaming":          true,
			"file_operations":    true,
			"browser_automation": true,
			"memory_graph":       true,
		},
		Skills: []models.Skill{
			{
				Name:        "web_browsing",
				Description: "Automated web browsing and data extraction",
			},
			{
				Name:        "code_generation",
				Description: "Generate and edit code files",
			},
			{
				Name:        "terminal_control",
				Description: "Execute terminal commands",
			},
			{
				Name:        "memory_management",
				Description: "Store and retrieve from knowledge graph",
			},
		},
		URL:       fmt.Sprintf("ws://%s/ws/a2a", c.Hostname()),
		Transport: "JSONRPC",
	}

	return c.JSON(card)
}

// Placeholder handlers - implement in separate files
func (s *Server) handleAgentInitialize(c fiber.Ctx) error {
	var req models.InitializeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	sessionID, err := s.agentController.Initialize(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"session_id":   sessionID,
		"agent_state":  "idle",
		"capabilities": []string{"browser", "terminal", "mcp", "memory"},
	})
}

func (s *Server) handleAgentCommand(c fiber.Ctx) error {
	var req models.CommandRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	taskID, err := s.agentController.ExecuteCommand(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"task_id": taskID,
		"status":  "queued",
	})
}

func (s *Server) handleAgentStatus(c fiber.Ctx) error {
	status := s.agentController.GetStatus()
	return c.JSON(status)
}

func (s *Server) handleAgentPause(c fiber.Ctx) error {
	if err := s.agentController.Pause(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (s *Server) handleAgentResume(c fiber.Ctx) error {
	if err := s.agentController.Resume(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

// File operation handlers
func (s *Server) handleFileTree(c fiber.Ctx) error {
	path := c.Query("path", "/")
	tree, err := s.agentController.GetFileTree(path)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tree)
}

func (s *Server) handleFileContent(c fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return c.Status(400).JSON(fiber.Map{"error": "path required"})
	}
	content, err := s.agentController.GetFileContent(path)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(content)
}

func (s *Server) handleFileWrite(c fiber.Ctx) error {
	var req models.FileWriteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	result, err := s.agentController.WriteFile(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (s *Server) handleFileDiff(c fiber.Ctx) error {
	var req models.FileDiffRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	result, err := s.agentController.ApplyDiff(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (s *Server) handleFileSearch(c fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "query required"})
	}
	results, err := s.agentController.SearchFiles(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(results)
}

// Memory handlers
func (s *Server) handleMemoryQuery(c fiber.Ctx) error {
	var req models.MemoryQueryRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	result, err := s.longTermMem.Query(req.Query, req.Mode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (s *Server) handleMemoryStore(c fiber.Ctx) error {
	var req models.MemoryStoreRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	nodeID, err := s.longTermMem.Store(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "node_id": nodeID})
}

func (s *Server) handleMemoryContext(c fiber.Ctx) error {
	taskID := c.Query("task_id")
	if taskID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "task_id required"})
	}
	context, err := s.longTermMem.GetContext(taskID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(context)
}

func (s *Server) handleVectorSearch(c fiber.Ctx) error {
	var req models.VectorSearchRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	results, err := s.longTermMem.VectorSearch(req.Query, req.TopK)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(results)
}

// OpenEvolve handlers
func (s *Server) handleOpenEvolveStatus(c fiber.Ctx) error {
	status := s.watchdog.GetStatus()
	return c.JSON(status)
}

func (s *Server) handleOpenEvolveProposal(c fiber.Ctx) error {
	var req models.ProposalRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	proposalID, err := s.watchdog.SubmitProposal(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"proposal_id": proposalID, "status": "pending_review"})
}

func (s *Server) handleOpenEvolveRewards(c fiber.Ctx) error {
	var req models.RewardRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := s.watchdog.UpdateRewards(req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (s *Server) handleOpenEvolveMetrics(c fiber.Ctx) error {
	metrics := s.watchdog.GetMetrics()
	return c.JSON(metrics)
}

// Vision handlers
func (s *Server) handleVisionAnalyze(c fiber.Ctx) error {
	var req models.VisionAnalyzeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	result, err := s.browserMgr.AnalyzeScreenshot(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (s *Server) handleVisionState(c fiber.Ctx) error {
	taskID := c.Query("task_id")
	if taskID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "task_id required"})
	}
	state, err := s.shortTermMem.GetPageState(taskID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(state)
}

func (s *Server) handleVisionContext(c fiber.Ctx) error {
	var req models.VisionContextRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	result, err := s.browserMgr.UpdateContext(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (s *Server) InitializeTerminal() error {
	// Start Neo4j
	output, err := s.terminalMgr.Execute("neo4j status")
	if err != nil {
		log.Printf("Neo4j not running, attempting to start...")
		output, err = s.terminalMgr.Execute("neo4j start")
		if err != nil {
			return fmt.Errorf("failed to start neo4j: %w", err)
		}
	}

	s.wsHandler.BroadcastTerminalOutput("✓ Neo4j started\n" + output)

	// Wait for Neo4j to be ready
	time.Sleep(5 * time.Second)

	// Initialize LightRAG schema
	if err := s.longTermMem.InitializeSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	s.wsHandler.BroadcastTerminalOutput("✓ LightRAG initialized")

	// Start watchdog
	s.watchdog.Start()
	s.wsHandler.BroadcastTerminalOutput("✓ Watchdog monitoring active")

	s.wsHandler.BroadcastTerminalOutput("🚀 Agent workspace ready!")

	return nil
}

func (s *Server) Cleanup() {
	log.Println("Cleaning up resources...")

	if s.browserMgr != nil {
		s.browserMgr.Cleanup()
	}

	if s.shortTermMem != nil {
		s.shortTermMem.ClearAll()
	}

	if s.longTermMem != nil {
		s.longTermMem.Close()
	}

	if s.watchdog != nil {
		s.watchdog.Stop()
	}

	log.Println("Cleanup complete")
}

