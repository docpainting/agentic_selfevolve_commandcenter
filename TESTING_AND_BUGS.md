# Testing & Bug Fixes

## 🐛 Known Issues & Fixes

This document tracks bugs found during integration testing and their fixes.

---

## ⚠️ Critical Issues

### 1. **Missing Agent Controller Initialization in main.go**

**Issue:** The agent controller is created but not fully initialized with all dependencies.

**Location:** `backend/cmd/server/main.go`

**Fix Needed:**
```go
// Current (incomplete)
agentController := agent.NewController(ollamaClient, memorySystem)

// Should be:
agentController := agent.NewController(&agent.Config{
    OllamaClient:    ollamaClient,
    MemorySystem:    memorySystem,
    BrowserManager:  browserMgr,
    TerminalManager: terminalMgr,
    MCPClient:       mcpClient,
    Watchdog:        watchdog,
})
```

### 2. **WebSocket Handler Missing Agent Controller Parameter**

**Issue:** WebSocket handlers expect agent controller but it's not passed.

**Location:** `backend/internal/websocket/chat.go`, `backend/internal/websocket/a2a.go`

**Current:**
```go
func HandleChatWebSocket(c *fiber.Ctx) error {
    // Missing agent controller
}
```

**Fixed Version Needed:**
```go
func HandleChatWebSocket(controller *agent.Controller) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Now has access to controller
    }
}
```

### 3. **Missing Manager Initializations**

**Issue:** Browser, Terminal, MCP managers not created in main.go

**Fix:**
```go
// Add before agent controller
browserMgr := browser.NewManager(&browser.Config{
    Headless: true,
    Timeout:  time.Second * 30,
})

terminalMgr := terminal.NewManager(&terminal.Config{
    DefaultShell: "/bin/bash",
    MaxSessions:  10,
})

mcpClient := mcp.NewClient(&mcp.Config{
    ConfigPath: "./mcp-config.json",
})
```

---

## 🔧 Medium Priority Issues

### 4. **Memory System Not Fully Initialized**

**Issue:** LongTermMemory and ShortTermMemory created but not connected properly.

**Location:** `backend/cmd/server/main.go`

**Fix:**
```go
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
    log.Fatal(err)
}

// Initialize short-term memory
shortTerm := memory.NewShortTermMemory()

// Combine into memory system
memorySystem := &memory.System{
    LongTerm:  longTerm,
    ShortTerm: shortTerm,
}
```

### 5. **Missing Environment Variable Validation**

**Issue:** No validation that required env vars are set.

**Fix:** Add at start of main():
```go
func validateEnv() error {
    required := []string{
        "NEO4J_URI",
        "NEO4J_USER",
        "NEO4J_PASSWORD",
        "OLLAMA_HOST",
    }
    
    for _, key := range required {
        if os.Getenv(key) == "" {
            return fmt.Errorf("required environment variable %s not set", key)
        }
    }
    return nil
}

// In main()
if err := validateEnv(); err != nil {
    log.Fatal(err)
}
```

### 6. **CORS Configuration Missing**

**Issue:** Frontend won't be able to connect from different port.

**Fix:**
```go
import "github.com/gofiber/fiber/v3/middleware/cors"

app.Use(cors.New(cors.Config{
    AllowOrigins: "http://localhost:3000",
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
    AllowCredentials: true,
}))
```

### 7. **Missing Graceful Shutdown**

**Issue:** No cleanup on shutdown.

**Fix:**
```go
// Add signal handling
c := make(chan os.Signal, 1)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)

go func() {
    <-c
    log.Println("Shutting down gracefully...")
    
    // Cleanup
    browserMgr.Close()
    terminalMgr.CloseAll()
    mcpClient.DisconnectAll()
    longTerm.Close()
    
    app.Shutdown()
}()
```

---

## 🐞 Minor Issues

### 8. **Hardcoded Ports**

**Issue:** Port 8080 is hardcoded.

**Fix:**
```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}

log.Fatal(app.Listen(":" + port))
```

### 9. **Missing Logging Middleware**

**Fix:**
```go
import "github.com/gofiber/fiber/v3/middleware/logger"

app.Use(logger.New(logger.Config{
    Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
}))
```

### 10. **No Health Check Endpoint**

**Fix:**
```go
app.Get("/health", func(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "status": "ok",
        "services": fiber.Map{
            "neo4j":  longTerm.HealthCheck(),
            "ollama": ollamaClient.HealthCheck(),
            "browser": browserMgr.IsHealthy(),
        },
    })
})
```

---

## 📝 Complete Fixed main.go

Here's the corrected version:

```go
package main

import (
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
        AppName: "Agentic Command Center",
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
    
    // Initialize short-term memory
    shortTerm := memory.NewShortTermMemory()
    
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
    
    // Initialize terminal manager
    terminalMgr := terminal.NewManager(&terminal.Config{
        DefaultShell: "/bin/bash",
        MaxSessions:  10,
    })
    
    // Initialize MCP client
    mcpClient := mcp.NewClient(&mcp.Config{
        ConfigPath: "./mcp-config.json",
    })
    
    // Initialize watchdog
    watchdogSvc := watchdog.NewWatchdog(&watchdog.Config{
        Enabled:        true,
        ScanInterval:   time.Second * 30,
        MinConfidence:  0.7,
        AlertThreshold: watchdog.SeverityWarning,
        Memory:         memorySystem,
    })
    watchdogSvc.Start()
    
    // Initialize agent controller
    agentController := agent.NewController(&agent.Config{
        OllamaClient:    ollamaClient,
        MemorySystem:    memorySystem,
        BrowserManager:  browserMgr,
        TerminalManager: terminalMgr,
        MCPClient:       mcpClient,
        Watchdog:        watchdogSvc,
    })
    
    // Initialize EvoX adapter
    evoxAdapter := agent.NewEvoXAdapter(ollamaClient, &agent.EvoXConfig{
        Enabled:            true,
        LearningRate:       0.1,
        MinConfidence:      0.7,
        EvolutionThreshold: 10,
    })
    
    // Routes
    api := app.Group("/api")
    
    // Health check
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status": "ok",
            "services": fiber.Map{
                "neo4j":  longTerm.HealthCheck(),
                "ollama": ollamaClient.HealthCheck(),
                "browser": browserMgr.IsHealthy(),
            },
        })
    })
    
    // Agent routes
    api.Post("/agent/initialize", agentController.Initialize)
    api.Post("/agent/command", agentController.ExecuteCommand)
    api.Get("/agent/status", agentController.GetStatus)
    
    // EvoX routes
    api.Post("/evox/workflow", evoxAdapter.GenerateWorkflowHandler)
    api.Post("/evox/action", evoxAdapter.ParseActionHandler)
    api.Post("/evox/evaluate", evoxAdapter.EvaluateHandler)
    
    // Watchdog routes
    api.Get("/watchdog/alerts", watchdogSvc.GetAlertsHandler)
    api.Post("/watchdog/alerts/:id/acknowledge", watchdogSvc.AcknowledgeAlertHandler)
    api.Get("/watchdog/patterns", watchdogSvc.GetPatternsHandler)
    api.Get("/watchdog/components", watchdogSvc.GetComponentsHandler)
    
    // WebSocket routes
    app.Get("/ws/chat", websocket.HandleChatWebSocket(agentController))
    app.Get("/ws/a2a", websocket.HandleA2AWebSocket(agentController))
    
    // Graceful shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        log.Println("Shutting down gracefully...")
        
        // Cleanup
        watchdogSvc.Stop()
        browserMgr.Close()
        terminalMgr.CloseAll()
        mcpClient.DisconnectAll()
        longTerm.Close()
        
        app.Shutdown()
    }()
    
    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server starting on port %s...\n", port)
    log.Fatal(app.Listen(":" + port))
}
```

---

## 🧪 Testing Checklist

### Backend Tests

- [ ] Server starts without errors
- [ ] Health check returns 200 OK
- [ ] Neo4j connection works
- [ ] Ollama connection works
- [ ] WebSocket chat connects
- [ ] WebSocket A2A connects
- [ ] Agent executes simple command
- [ ] Browser manager works
- [ ] Terminal manager works
- [ ] MCP client connects
- [ ] Watchdog detects patterns
- [ ] Memory stores and retrieves

### Frontend Tests

- [ ] UI loads without errors
- [ ] WebSocket connects (green indicator)
- [ ] Can send messages
- [ ] Receives agent responses
- [ ] File tree works
- [ ] OpenEvolve panel shows data
- [ ] Terminal tab works
- [ ] Browser tab works
- [ ] MCP tools tab works
- [ ] Takeover mode toggles

### Integration Tests

- [ ] Send command → Agent responds
- [ ] Agent uses browser
- [ ] Agent uses terminal
- [ ] Agent uses MCP tools
- [ ] Watchdog generates alerts
- [ ] Alerts appear in UI
- [ ] Memory persists across restarts
- [ ] Patterns are learned
- [ ] EvoX evolves workflows

---

## 🚀 Quick Test Commands

### Start Backend
```bash
cd backend
go run cmd/server/main.go
```

### Start Frontend
```bash
cd frontend
npm run dev
```

### Test Health
```bash
curl http://localhost:8080/health
```

### Test WebSocket (wscat)
```bash
npm install -g wscat
wscat -c ws://localhost:8080/ws/chat
```

### Test Agent Command
```bash
curl -X POST http://localhost:8080/api/agent/command \
  -H "Content-Type: application/json" \
  -d '{"command": "Hello, what can you do?"}'
```

---

## 📊 Expected Behavior

### On Startup

```
Server starting on port 8080...
✓ Neo4j connected
✓ Ollama connected (gemma3:27b)
✓ Browser manager initialized
✓ Terminal manager initialized
✓ MCP client initialized
✓ Watchdog started
✓ Agent controller ready
[2024-10-26 12:00:00] Server listening on :8080
```

### On First Command

```
[Chat] User: Hello, what can you do?
[Agent] Thinking...
[Agent] I can help you with:
- Code generation and editing
- Web browsing and research
- Terminal command execution
- File management
- Pattern detection
- And much more!
```

### Watchdog Alert Example

```json
{
  "id": "alert-001",
  "type": "Pattern",
  "severity": "info",
  "title": "New Pattern Detected",
  "message": "Authentication pattern detected in auth.go",
  "file": "backend/middleware/auth.go",
  "line": 25,
  "timestamp": "2024-10-26T12:05:00Z"
}
```

---

## 🐛 Common Errors & Solutions

### Error: "dial tcp: connection refused" (Neo4j)

**Solution:** Start Neo4j
```bash
sudo systemctl start neo4j
# or
neo4j start
```

### Error: "connect: connection refused" (Ollama)

**Solution:** Start Ollama
```bash
ollama serve
```

### Error: "module not found"

**Solution:** Run go mod tidy
```bash
cd backend
go mod tidy
```

### Error: "CORS policy" in browser console

**Solution:** Check CORS middleware is configured correctly

### Error: WebSocket connection fails

**Solution:** Check both servers are running and ports match

---

## 📝 Next Steps

1. **Apply all fixes** to main.go
2. **Test locally** with Go 1.25+
3. **Fix any runtime errors**
4. **Document new bugs** found
5. **Iterate until stable**

---

## 🎯 Success Criteria

System is working when:
- ✅ Backend starts without errors
- ✅ Frontend connects via WebSocket
- ✅ Can send/receive messages
- ✅ Agent executes commands
- ✅ Browser automation works
- ✅ Terminal execution works
- ✅ Watchdog generates alerts
- ✅ Memory persists data
- ✅ No console errors

**Then we're ready to demo!** 🚀

