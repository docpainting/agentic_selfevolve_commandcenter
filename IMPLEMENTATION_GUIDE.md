# Agent Workspace - Implementation Guide

This guide provides the complete implementation details for all remaining backend components.

## Status

✅ **Completed:**
- Project structure
- Frontend (complete midnight glassmorphism UI)
- Backend entry point (cmd/server/main.go)
- Shared models (pkg/models/)
- JSON-RPC 2.0 implementation (pkg/jsonrpc/)
- Environment configuration
- Documentation (README, LICENSE, .gitignore)

🚧 **To Implement:**
The following components need to be implemented based on the specifications provided:

### 1. Memory Systems (`internal/memory/`)

**Files to create:**
- `long_term.go` - LightRAG + Neo4j integration
- `short_term.go` - ChromeDP-based task memory
- `embeddings.go` - Ollama nomic-embed-text integration

**Key Functions:**
```go
// long_term.go
func NewLongTermMemory() (*LongTermMemory, error)
func (m *LongTermMemory) Store(req models.MemoryStoreRequest) (string, error)
func (m *LongTermMemory) Query(query string, mode string) (interface{}, error)
func (m *LongTermMemory) VectorSearch(query string, topK int) ([]interface{}, error)
func (m *LongTermMemory) InitializeSchema() error
func (m *LongTermMemory) Close() error

// short_term.go
func NewShortTermMemory() *ShortTermMemory
func (m *ShortTermMemory) StorePerception(taskID string, perception models.Perception) error
func (m *ShortTermMemory) GetTaskMemory(taskID string) (*TaskMemory, error)
func (m *ShortTermMemory) ClearTask(taskID string, archiveToLongTerm bool) error
```

**Dependencies:**
```go
import (
    "github.com/MegaGrindStone/go-light-rag"
    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)
```

**Configuration:**
- Neo4j URI from env: `NEO4J_URI`
- ChromeM DB path: `CHROMEM_DB_PATH`
- BoltDB path: `BOLT_DB_PATH`

### 2. Browser Manager (`internal/browser/`)

**Files to create:**
- `manager.go` - ChromeDP lifecycle management
- `vision.go` - Screenshot analysis with numbered overlays
- `actions.go` - Browser action execution

**Key Functions:**
```go
func NewManager(shortTermMem *memory.ShortTermMemory) *Manager
func (m *Manager) CaptureScreenshot(taskID string) ([]byte, error)
func (m *Manager) AnalyzeScreenshot(req models.VisionAnalyzeRequest) (interface{}, error)
func (m *Manager) DrawNumberedOverlays(screenshot []byte, elements []models.BrowserElement) ([]byte, error)
func (m *Manager) Navigate(url string) error
func (m *Manager) Click(elementID int) error
func (m *Manager) Type(elementID int, text string) error
func (m *Manager) Cleanup() error
```

**Dependencies:**
```go
import (
    "github.com/chromedp/chromedp"
    "image"
    "image/draw"
    "image/png"
)
```

**Features:**
- Headless Chrome with visible option
- Screenshot capture with ImageMagick-style overlays
- Element detection with bounding boxes
- AI command parser (CLICK 5, TYPE 3 "text")

### 3. Terminal Manager (`internal/terminal/`)

**Files to create:**
- `manager.go` - PTY management
- `executor.go` - Command execution

**Key Functions:**
```go
func NewManager() *Terminal
func (t *Terminal) Execute(command string) (string, error)
func (t *Terminal) ExecuteWithContext(ctx context.Context, command string) (string, error)
func (t *Terminal) GetOutput() string
func (t *Terminal) SendInput(input string) error
```

**Features:**
- PTY for interactive commands
- Command history tracking
- AI vs User command attribution
- Output streaming

### 4. MCP Client (`internal/mcp/`)

**Files to create:**
- `client.go` - MCP protocol client
- `stdio.go` - stdio transport for MCP servers

**Key Functions:**
```go
func NewClient() *Client
func (c *Client) ConnectServer(name string, command string, args []string) error
func (c *Client) ListTools(server string) ([]Tool, error)
func (c *Client) CallTool(server string, tool string, args map[string]interface{}) (*models.MCPToolResult, error)
func (c *Client) Disconnect(server string) error
```

**Configuration:**
Load from `config/mcp.json`:
```json
{
  "servers": {
    "dynamic-thinking": {
      "command": "../mcp-dynamic-thinking/mcp-dynamic-thinking",
      "args": [],
      "env": {
        "NEO4J_URI": "bolt://localhost:7687"
      }
    }
  }
}
```

### 5. Watchdog (`internal/watchdog/`)

**Files to create:**
- `watchdog.go` - Pattern detection and monitoring
- `alerts.go` - Alert generation

**Key Functions:**
```go
func NewWatchdog(longTermMem *memory.LongTermMemory) *Watchdog
func (w *Watchdog) Start() error
func (w *Watchdog) Stop() error
func (w *Watchdog) DetectPattern(code string) ([]Alert, error)
func (w *Watchdog) SubmitProposal(req models.ProposalRequest) (string, error)
func (w *Watchdog) GetStatus() interface{}
```

**Features:**
- Concept wiring detection
- Pattern recognition in code
- Dependency tracking
- Security issue detection
- Proposal management

### 6. Agent Controller (`internal/agent/`)

**Files to create:**
- `controller.go` - Main orchestration
- `gemma.go` - Ollama Gemma 3 integration
- `planner.go` - Task planning
- `executor.go` - Task execution

**Key Functions:**
```go
func NewController(
    longTermMem *memory.LongTermMemory,
    shortTermMem *memory.ShortTermMemory,
    browserMgr *browser.Manager,
    terminalMgr *terminal.Manager,
    mcpClient *mcp.Client,
    watchdog *watchdog.Watchdog,
) *Controller

func (c *Controller) Initialize(req models.InitializeRequest) (string, error)
func (c *Controller) ExecuteCommand(req models.CommandRequest) (string, error)
func (c *Controller) GetStatus() models.AgentStatus
func (c *Controller) Pause() error
func (c *Controller) Resume() error
```

**Ollama Integration:**
```go
// Use Ollama API
url := os.Getenv("OLLAMA_HOST") + "/api/generate"
model := os.Getenv("OLLAMA_MODEL") // gemma3:27b

payload := map[string]interface{}{
    "model": model,
    "prompt": prompt,
    "stream": false,
}
```

### 7. WebSocket Handlers (`internal/websocket/`)

**Files to create:**
- `chat.go` - Chat WebSocket handler
- `a2a.go` - A2A JSON-RPC WebSocket handler

**Key Functions:**
```go
// chat.go
func NewHandler(agentController *agent.Controller) *Handler
func (h *Handler) HandleWebSocket(c fiber.Ctx) error
func (h *Handler) BroadcastMessage(msg models.Message) error
func (h *Handler) BroadcastTerminalOutput(output string) error

// a2a.go
func NewA2AHandler(agentController *agent.Controller) *A2AHandler
func (h *A2AHandler) HandleWebSocket(c fiber.Ctx) error
func (h *A2AHandler) HandleJSONRPC(req *jsonrpc.Request) *jsonrpc.Response
```

**WebSocket Message Flow:**
```
Client → WS → Handler → Agent Controller → (Browser/Terminal/MCP) → Response → WS → Client
```

### 8. MCP Dynamic Thinking Server (`mcp-dynamic-thinking/`)

**Directory Structure:**
```
mcp-dynamic-thinking/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── perceive/
│   │   └── perceive.go
│   ├── reason/
│   │   └── reason.go
│   ├── act/
│   │   └── act.go
│   └── reflect/
│       └── reflect.go
├── go.mod
└── go.sum
```

**Key Implementation:**
Based on the MCP specification document (`mcp_dynamic_thinking_specification.md`), implement:
- Perceive phase with vision analysis
- Reason phase with branching
- Act phase with tool execution
- Reflect phase with self-improvement

**MCP Server Setup:**
```go
import "github.com/mark3labs/mcp-go/server"

s := server.NewMCPServer("dynamic-thinking", "1.0.0")

// Register tools
s.AddTool(mcp.Tool{
    Name: "perceive",
    Description: "Capture and analyze environment state",
    InputSchema: ...,
}, handlePerceive)

// Start server (stdio transport)
s.Serve()
```

## Implementation Order

1. **Memory Systems** - Foundation for everything
2. **Terminal Manager** - Needed for initialization
3. **Browser Manager** - Core automation
4. **MCP Client** - Tool integration
5. **Watchdog** - Monitoring
6. **Agent Controller** - Orchestration
7. **WebSocket Handlers** - Communication
8. **MCP Dynamic Thinking Server** - Advanced reasoning

## Testing

### Unit Tests
```bash
cd backend
go test ./...
```

### Integration Test
```bash
# Start Neo4j
neo4j start

# Start Ollama
ollama serve

# Start backend
cd backend
go run cmd/server/main.go

# Start frontend
cd frontend
npm run dev

# Access http://localhost:3000
```

## Environment Setup

```bash
# Copy and configure
cp .env.example .env
nano .env

# Required:
NEO4J_URI=bolt://localhost:7687
NEO4J_PASSWORD=your_password
OLLAMA_HOST=http://localhost:11434
OLLAMA_MODEL=gemma3:27b
```

## Dependencies to Install

```bash
cd backend
go get github.com/MegaGrindStone/go-light-rag
go get github.com/chromedp/chromedp
go get github.com/gofiber/fiber/v3
go get github.com/gofiber/websocket/v3
go get github.com/neo4j/neo4j-go-driver/v5
go get github.com/mark3labs/mcp-go
go mod tidy
```

## Additional Resources

- **LightRAG Go**: https://github.com/MegaGrindStone/go-light-rag
- **ChromeDP**: https://github.com/chromedp/chromedp
- **Fiber v3**: https://docs.gofiber.io/
- **Neo4j Go Driver**: https://neo4j.com/docs/go-manual/current/
- **MCP Go**: https://github.com/mark3labs/mcp-go

## Notes

- All specifications are in the `docs/` directory
- Frontend is 100% complete and functional
- Backend structure is complete, implementation needed
- All models and types are defined in `pkg/models/`
- JSON-RPC 2.0 is fully implemented in `pkg/jsonrpc/`

## Support

For questions or issues during implementation, refer to:
- `complete_design_with_lightrag.md`
- `final_complete_design_lightrag_chromedp.md`
- `complete_api_protocol_specification.md`
- `mcp_dynamic_thinking_specification.md`

