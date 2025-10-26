# Complete Agent Workspace - LightRAG + ChromeDP Integration

## System Architecture Overview

This system combines **long-term knowledge** (LightRAG + Neo4j) with **short-term memory** (ChromeDP) to create an intelligent agent workspace with the midnight glassmorphism aesthetic.

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Frontend (React/Vue)                            │
│              Midnight Glassmorphism Interface                       │
│  ┌──────────────┬────────────────────────────────┬─────────────────┐ │
│  │ Left Panel   │        Center Chat Area        │    Right Panel  │ │
│  │ File Tree    │    (WebSocket Communication)   │   OpenEvolve    │ │
│  │ (VS Code)    │                                │   Watchdog      │ │
│  └──────────────┴────────────────────────────────┴─────────────────┘ │
│  ┌──────────────────────────────────────────────────────────────────┐ │
│  │ Bottom Panel: Terminal | Browser | MCP | Logs | Knowledge Graph │ │
│  └──────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
                    ↕ WebSocket (Chat) + JSON-RPC 2.0 (A2A)
┌─────────────────────────────────────────────────────────────────────┐
│                   Go Fiber v3 Backend                               │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                    Communication Layer                        │  │
│  │  • WebSocket Handler (Real-time chat)                        │  │
│  │  • JSON-RPC 2.0 Handler (A2A protocol)                       │  │
│  │  • LightRAG Wrapper (Conversation capture)                   │  │
│  └──────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                    Execution Layer                            │  │
│  │  • ChromeDP Manager (Browser automation + short-term memory) │  │
│  │  • Terminal PTY Manager (Command execution)                  │  │
│  │  • MCP Client (Tool integration)                             │  │
│  └──────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │           Gemma 3 Agent Controller                           │  │
│  │         (Task Orchestration + Reasoning)                     │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                              ↕
┌─────────────────────────────────────────────────────────────────────┐
│                    Memory Architecture                              │
│  ┌────────────────────────────┬────────────────────────────────┐   │
│  │   LONG-TERM MEMORY         │   SHORT-TERM MEMORY            │   │
│  │   (LightRAG + Neo4j)       │   (ChromeDP Context)           │   │
│  ├────────────────────────────┼────────────────────────────────┤   │
│  │ • All conversations        │ • Current task screenshots     │   │
│  │ • Code history             │ • Browser state/cookies        │   │
│  │ • Concept relationships    │ • Dynamic page elements        │   │
│  │ • Patterns & learnings     │ • Temporary variables          │   │
│  │ • Entity extraction        │ • Session-specific data        │   │
│  │                            │ • Active automation context    │   │
│  │ Persistent across sessions │ Cleared after task completion  │   │
│  └────────────────────────────┴────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Memory System Design

### Long-Term Memory: LightRAG + Neo4j

**Purpose**: Persistent knowledge that grows over time and provides context for future tasks.

**What gets stored:**
- Every conversation message
- All code files and their structure
- Terminal commands and outputs
- Concepts, entities, and relationships
- Patterns detected by watchdog
- Task outcomes and learnings

**Storage Components:**
```go
type LongTermMemory struct {
    lightRAG    *golightrag.LightRAG
    neo4j       *storage.Neo4J
    chromem     *storage.Chromem  // Vector embeddings
    bolt        *storage.Bolt     // Key-value store
}
```

### Short-Term Memory: ChromeDP Context

**Purpose**: Temporary memory for active tasks, especially browser automation and dynamic data.

**What gets stored:**
- Screenshots from browser automation
- DOM state and element positions
- Cookies and session data
- Temporary task variables
- Active page context
- Dynamic form data

**Storage Components:**
```go
type ShortTermMemory struct {
    chromedpCtx     context.Context
    screenshots     map[string][]byte  // taskID -> screenshot
    pageStates      map[string]PageState
    taskVariables   map[string]interface{}
    sessionData     map[string]string
    expiryTime      time.Time
}
```

---

## Complete Go Backend Implementation

### Main Application Structure

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/websocket/v3"
    "github.com/chromedp/chromedp"
    golightrag "github.com/MegaGrindStone/go-light-rag"
    "github.com/MegaGrindStone/go-light-rag/storage"
    "github.com/MegaGrindStone/go-light-rag/handler"
    "github.com/MegaGrindStone/go-light-rag/llm"
)

type AgentWorkspace struct {
    // Communication
    app             *fiber.App
    wsClients       map[*websocket.Conn]bool
    clientsMux      sync.RWMutex
    
    // Memory Systems
    longTermMemory  *LongTermMemory
    shortTermMemory *ShortTermMemory
    
    // Execution
    chromedpMgr     *ChromeDPManager
    terminalMgr     *TerminalManager
    mcpClient       *MCPClient
    
    // Intelligence
    agentController *AgentController
    watchdog        *Watchdog
    
    // Handlers
    methodHandlers  map[string]JSONRPCHandler
}

func NewAgentWorkspace() *AgentWorkspace {
    ws := &AgentWorkspace{
        app:       fiber.New(),
        wsClients: make(map[*websocket.Conn]bool),
        methodHandlers: make(map[string]JSONRPCHandler),
    }
    
    // Initialize memory systems
    ws.initializeLongTermMemory()
    ws.initializeShortTermMemory()
    
    // Initialize execution layer
    ws.chromedpMgr = NewChromeDPManager(ws.shortTermMemory)
    ws.terminalMgr = NewTerminalManager()
    ws.mcpClient = NewMCPClient()
    
    // Initialize intelligence layer
    ws.agentController = NewAgentController(ws)
    ws.watchdog = NewWatchdog(ws.longTermMemory)
    
    // Register handlers
    ws.registerHandlers()
    
    return ws
}

func (ws *AgentWorkspace) initializeLongTermMemory() {
    // Initialize Neo4j
    neo4j, err := storage.NewNeo4J(
        "bolt://localhost:7687",
        "neo4j",
        "password",
    )
    if err != nil {
        log.Fatalf("Failed to connect to Neo4j: %v", err)
    }
    
    // Initialize ChromeM for vector storage
    embeddingFunc := storage.EmbeddingFunc(
        chromem.NewEmbeddingFuncOpenAI(
            os.Getenv("OPENAI_API_KEY"),
            chromem.EmbeddingModelOpenAI3Large,
        ),
    )
    
    chromem, err := storage.NewChromem("vec.db", 5, embeddingFunc)
    if err != nil {
        log.Fatalf("Failed to initialize ChromeM: %v", err)
    }
    
    // Initialize BoltDB for key-value storage
    bolt, err := storage.NewBolt("kv.db")
    if err != nil {
        log.Fatalf("Failed to initialize BoltDB: %v", err)
    }
    
    // Initialize LLM
    llmClient := llm.NewOpenAI(
        os.Getenv("OPENAI_API_KEY"),
        "gpt-4",
        nil,
        log.Default(),
    )
    
    ws.longTermMemory = &LongTermMemory{
        neo4j:    neo4j,
        chromem:  chromem,
        bolt:     bolt,
        llm:      llmClient,
        handler:  handler.Default{},
    }
    
    log.Println("✓ Long-term memory (LightRAG + Neo4j) initialized")
}

func (ws *AgentWorkspace) initializeShortTermMemory() {
    // Create ChromeDP context
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", true),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("no-sandbox", true),
    )
    
    allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
    ctx, cancel := chromedp.NewContext(allocCtx)
    
    ws.shortTermMemory = &ShortTermMemory{
        chromedpCtx:   ctx,
        cancel:        cancel,
        screenshots:   make(map[string][]byte),
        pageStates:    make(map[string]PageState),
        taskVariables: make(map[string]interface{}),
        sessionData:   make(map[string]string),
    }
    
    log.Println("✓ Short-term memory (ChromeDP) initialized")
}

// ===== Long-Term Memory Implementation =====

type LongTermMemory struct {
    neo4j   *storage.Neo4J
    chromem *storage.Chromem
    bolt    *storage.Bolt
    llm     llm.LLM
    handler handler.DocumentHandler
    mu      sync.RWMutex
}

func (ltm *LongTermMemory) CaptureConversation(msg Message) error {
    ltm.mu.Lock()
    defer ltm.mu.Unlock()
    
    // Create document for LightRAG
    doc := golightrag.Document{
        ID:      msg.ID,
        Content: fmt.Sprintf("[%s] %s: %s", msg.Timestamp, msg.Role, msg.Content),
    }
    
    // Insert into LightRAG (stores in Neo4j + ChromeM + Bolt)
    err := golightrag.Insert(doc, ltm.handler, ltm, ltm.llm, log.Default())
    if err != nil {
        return fmt.Errorf("failed to capture conversation: %w", err)
    }
    
    // Create additional relationships in Neo4j
    if msg.TaskID != "" {
        query := `
            MATCH (m:Message {id: $msgId})
            MATCH (t:Task {id: $taskId})
            MERGE (t)-[:HAS_MESSAGE]->(m)
        `
        err = ltm.neo4j.Execute(query, map[string]interface{}{
            "msgId":  msg.ID,
            "taskId": msg.TaskID,
        })
    }
    
    return err
}

func (ltm *LongTermMemory) CaptureCode(file FileChange) error {
    ltm.mu.Lock()
    defer ltm.mu.Unlock()
    
    // Create document for code
    doc := golightrag.Document{
        ID:      file.Path,
        Content: fmt.Sprintf("File: %s\nLanguage: %s\n\n%s", file.Path, file.Language, file.Content),
    }
    
    // Use Go handler for Go files, default for others
    var h handler.DocumentHandler
    if file.Language == "go" {
        h = handler.Go{}
    } else {
        h = handler.Default{}
    }
    
    // Insert into LightRAG
    err := golightrag.Insert(doc, h, ltm, ltm.llm, log.Default())
    if err != nil {
        return fmt.Errorf("failed to capture code: %w", err)
    }
    
    // Store file metadata in Neo4j
    query := `
        MERGE (f:File {path: $path})
        SET f.language = $language,
            f.content = $content,
            f.timestamp = $timestamp
    `
    
    return ltm.neo4j.Execute(query, map[string]interface{}{
        "path":      file.Path,
        "language":  file.Language,
        "content":   file.Content,
        "timestamp": time.Now().Unix(),
    })
}

func (ltm *LongTermMemory) QueryContext(query string) (*golightrag.QueryResult, error) {
    ltm.mu.RLock()
    defer ltm.mu.RUnlock()
    
    // Create conversation for query
    conversation := []golightrag.QueryConversation{
        {
            Role:    golightrag.RoleUser,
            Message: query,
        },
    }
    
    // Query LightRAG
    result, err := golightrag.Query(conversation, ltm.handler, ltm, ltm.llm, log.Default())
    if err != nil {
        return nil, fmt.Errorf("failed to query context: %w", err)
    }
    
    return &result, nil
}

// Implement storage interfaces for LightRAG
func (ltm *LongTermMemory) GraphStorage() storage.GraphStorage {
    return ltm.neo4j
}

func (ltm *LongTermMemory) VectorStorage() storage.VectorStorage {
    return ltm.chromem
}

func (ltm *LongTermMemory) KeyValueStorage() storage.KeyValueStorage {
    return ltm.bolt
}

// ===== Short-Term Memory Implementation =====

type ShortTermMemory struct {
    chromedpCtx   context.Context
    cancel        context.CancelFunc
    screenshots   map[string][]byte
    pageStates    map[string]PageState
    taskVariables map[string]interface{}
    sessionData   map[string]string
    mu            sync.RWMutex
}

type PageState struct {
    URL         string
    Title       string
    Elements    []ElementInfo
    Cookies     []*network.Cookie
    Screenshot  []byte
    Timestamp   time.Time
}

type ElementInfo struct {
    Index    int
    Tag      string
    Text     string
    X        float64
    Y        float64
    Width    float64
    Height   float64
}

func (stm *ShortTermMemory) CaptureScreenshot(taskID string) ([]byte, error) {
    stm.mu.Lock()
    defer stm.mu.Unlock()
    
    var buf []byte
    err := chromedp.Run(stm.chromedpCtx,
        chromedp.CaptureScreenshot(&buf),
    )
    
    if err != nil {
        return nil, err
    }
    
    stm.screenshots[taskID] = buf
    return buf, nil
}

func (stm *ShortTermMemory) GetScreenshot(taskID string) ([]byte, bool) {
    stm.mu.RLock()
    defer stm.mu.RUnlock()
    
    screenshot, exists := stm.screenshots[taskID]
    return screenshot, exists
}

func (stm *ShortTermMemory) CapturePageState(taskID string) (*PageState, error) {
    stm.mu.Lock()
    defer stm.mu.Unlock()
    
    state := &PageState{
        Timestamp: time.Now(),
    }
    
    // Capture URL, title, screenshot
    err := chromedp.Run(stm.chromedpCtx,
        chromedp.Location(&state.URL),
        chromedp.Title(&state.Title),
        chromedp.CaptureScreenshot(&state.Screenshot),
    )
    
    if err != nil {
        return nil, err
    }
    
    // Capture interactive elements
    var nodes []*cdp.Node
    err = chromedp.Run(stm.chromedpCtx,
        chromedp.Nodes("a, button, input, select, textarea", &nodes, chromedp.ByQueryAll),
    )
    
    if err == nil {
        for i, node := range nodes {
            var box *dom.BoxModel
            chromedp.Run(stm.chromedpCtx,
                chromedp.ActionFunc(func(ctx context.Context) error {
                    var err error
                    box, err = dom.GetBoxModel().WithNodeID(node.NodeID).Do(ctx)
                    return err
                }),
            )
            
            if box != nil && len(box.Content) >= 4 {
                state.Elements = append(state.Elements, ElementInfo{
                    Index:  i + 1,
                    Tag:    node.NodeName,
                    Text:   node.NodeValue,
                    X:      box.Content[0],
                    Y:      box.Content[1],
                    Width:  box.Content[2] - box.Content[0],
                    Height: box.Content[5] - box.Content[1],
                })
            }
        }
    }
    
    stm.pageStates[taskID] = *state
    return state, nil
}

func (stm *ShortTermMemory) SetVariable(key string, value interface{}) {
    stm.mu.Lock()
    defer stm.mu.Unlock()
    stm.taskVariables[key] = value
}

func (stm *ShortTermMemory) GetVariable(key string) (interface{}, bool) {
    stm.mu.RLock()
    defer stm.mu.RUnlock()
    val, exists := stm.taskVariables[key]
    return val, exists
}

func (stm *ShortTermMemory) ClearTask(taskID string) {
    stm.mu.Lock()
    defer stm.mu.Unlock()
    
    delete(stm.screenshots, taskID)
    delete(stm.pageStates, taskID)
    
    log.Printf("Cleared short-term memory for task: %s", taskID)
}

func (stm *ShortTermMemory) ClearAll() {
    stm.mu.Lock()
    defer stm.mu.Unlock()
    
    stm.screenshots = make(map[string][]byte)
    stm.pageStates = make(map[string]PageState)
    stm.taskVariables = make(map[string]interface{})
    
    log.Println("Cleared all short-term memory")
}

// ===== ChromeDP Manager =====

type ChromeDPManager struct {
    shortTermMem *ShortTermMemory
    activeTasks  map[string]context.Context
    mu           sync.RWMutex
}

func NewChromeDPManager(stm *ShortTermMemory) *ChromeDPManager {
    return &ChromeDPManager{
        shortTermMem: stm,
        activeTasks:  make(map[string]context.Context),
    }
}

func (cdm *ChromeDPManager) Navigate(taskID, url string) error {
    cdm.mu.Lock()
    defer cdm.mu.Unlock()
    
    err := chromedp.Run(cdm.shortTermMem.chromedpCtx,
        chromedp.Navigate(url),
        chromedp.WaitReady("body"),
    )
    
    if err != nil {
        return err
    }
    
    // Capture page state in short-term memory
    _, err = cdm.shortTermMem.CapturePageState(taskID)
    return err
}

func (cdm *ChromeDPManager) Click(taskID string, elementIndex int) error {
    cdm.mu.RLock()
    state, exists := cdm.shortTermMem.pageStates[taskID]
    cdm.mu.RUnlock()
    
    if !exists || elementIndex > len(state.Elements) {
        return fmt.Errorf("invalid element index")
    }
    
    element := state.Elements[elementIndex-1]
    
    // Click at element coordinates
    err := chromedp.Run(cdm.shortTermMem.chromedpCtx,
        chromedp.MouseClickXY(element.X+element.Width/2, element.Y+element.Height/2),
    )
    
    if err != nil {
        return err
    }
    
    // Update page state after click
    time.Sleep(500 * time.Millisecond)
    _, err = cdm.shortTermMem.CapturePageState(taskID)
    return err
}

func (cdm *ChromeDPManager) Type(taskID string, elementIndex int, text string) error {
    cdm.mu.RLock()
    state, exists := cdm.shortTermMem.pageStates[taskID]
    cdm.mu.RUnlock()
    
    if !exists || elementIndex > len(state.Elements) {
        return fmt.Errorf("invalid element index")
    }
    
    element := state.Elements[elementIndex-1]
    
    // Click element first, then type
    err := chromedp.Run(cdm.shortTermMem.chromedpCtx,
        chromedp.MouseClickXY(element.X+element.Width/2, element.Y+element.Height/2),
        chromedp.Sleep(100*time.Millisecond),
        chromedp.SendKeys("input, textarea", text, chromedp.ByQuery),
    )
    
    return err
}

func (cdm *ChromeDPManager) GetScreenshotWithOverlays(taskID string) ([]byte, error) {
    cdm.mu.RLock()
    state, exists := cdm.shortTermMem.pageStates[taskID]
    cdm.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("no page state for task")
    }
    
    // Capture fresh screenshot
    screenshot, err := cdm.shortTermMem.CaptureScreenshot(taskID)
    if err != nil {
        return nil, err
    }
    
    // TODO: Add numbered overlays to screenshot using image processing
    // For now, return raw screenshot
    return screenshot, nil
}

// ===== Agent Controller with Memory Integration =====

type AgentController struct {
    workspace      *AgentWorkspace
    gemma3Client   *Gemma3Client
}

func NewAgentController(ws *AgentWorkspace) *AgentController {
    return &AgentController{
        workspace:    ws,
        gemma3Client: NewGemma3Client(),
    }
}

func (ac *AgentController) ExecuteTask(taskID, userMessage string) error {
    // 1. Query long-term memory for context
    contextResult, err := ac.workspace.longTermMemory.QueryContext(userMessage)
    if err != nil {
        log.Printf("Warning: Failed to query context: %v", err)
    }
    
    // 2. Build prompt with context
    prompt := ac.buildPromptWithContext(userMessage, contextResult)
    
    // 3. Get agent decision from Gemma 3
    actions, err := ac.gemma3Client.GetActions(prompt)
    if err != nil {
        return err
    }
    
    // 4. Execute actions
    for _, action := range actions {
        switch action.Type {
        case "browser_navigate":
            url := action.Params["url"].(string)
            err = ac.workspace.chromedpMgr.Navigate(taskID, url)
            
            // Capture in short-term memory
            screenshot, _ := ac.workspace.shortTermMemory.CaptureScreenshot(taskID)
            
            // Store action in long-term memory
            ac.workspace.longTermMemory.CaptureConversation(Message{
                ID:      uuid.New().String(),
                Role:    "agent",
                Content: fmt.Sprintf("Navigated to %s", url),
                TaskID:  taskID,
            })
            
        case "browser_click":
            elementIndex := int(action.Params["element"].(float64))
            err = ac.workspace.chromedpMgr.Click(taskID, elementIndex)
            
        case "terminal_execute":
            command := action.Params["command"].(string)
            output, err := ac.workspace.terminalMgr.Execute(command)
            
            // Store in long-term memory
            ac.workspace.longTermMemory.CaptureConversation(Message{
                ID:      uuid.New().String(),
                Role:    "agent",
                Content: fmt.Sprintf("Executed: %s\nOutput: %s", command, output),
                TaskID:  taskID,
            })
        }
        
        if err != nil {
            return err
        }
    }
    
    // 5. Clear short-term memory after task completion
    defer ac.workspace.shortTermMemory.ClearTask(taskID)
    
    return nil
}

func (ac *AgentController) buildPromptWithContext(query string, context *golightrag.QueryResult) string {
    prompt := fmt.Sprintf("User Query: %s\n\n", query)
    
    if context != nil {
        prompt += "Relevant Context from Knowledge Graph:\n\n"
        
        // Add local entities
        if len(context.LocalEntities) > 0 {
            prompt += "Related Concepts:\n"
            for _, entity := range context.LocalEntities {
                prompt += fmt.Sprintf("- %s\n", entity)
            }
            prompt += "\n"
        }
        
        // Add source documents
        if len(context.LocalSources) > 0 {
            prompt += "Previous Conversations:\n"
            for _, source := range context.LocalSources {
                prompt += fmt.Sprintf("- %s\n", source.Content)
            }
            prompt += "\n"
        }
    }
    
    prompt += `
Based on the context above, determine the actions needed to complete the user's request.
Return actions in JSON format:
[
  {"type": "browser_navigate", "params": {"url": "https://..."}},
  {"type": "browser_click", "params": {"element": 5}},
  {"type": "terminal_execute", "params": {"command": "ls -la"}}
]
`
    
    return prompt
}

// ===== Terminal Initialization =====

func (ws *AgentWorkspace) InitializeTerminal() error {
    // Start Neo4j
    output, err := ws.terminalMgr.Execute("neo4j start")
    if err != nil {
        return fmt.Errorf("failed to start neo4j: %w", err)
    }
    
    ws.broadcastTerminalOutput("✓ Neo4j started\n" + output)
    
    // Wait for Neo4j to be ready
    time.Sleep(5 * time.Second)
    
    // Initialize LightRAG schema
    err = ws.longTermMemory.InitializeSchema()
    if err != nil {
        return fmt.Errorf("failed to initialize schema: %w", err)
    }
    
    ws.broadcastTerminalOutput("✓ LightRAG schema initialized")
    
    // Start watchdog
    ws.watchdog.Start()
    ws.broadcastTerminalOutput("✓ Watchdog monitoring active")
    
    ws.broadcastTerminalOutput("🚀 Agent workspace ready!")
    
    return nil
}

func (ltm *LongTermMemory) InitializeSchema() error {
    queries := []string{
        "CREATE CONSTRAINT message_id IF NOT EXISTS FOR (m:Message) REQUIRE m.id IS UNIQUE",
        "CREATE CONSTRAINT file_path IF NOT EXISTS FOR (f:File) REQUIRE f.path IS UNIQUE",
        "CREATE CONSTRAINT task_id IF NOT EXISTS FOR (t:Task) REQUIRE t.id IS UNIQUE",
        "CREATE INDEX message_timestamp IF NOT EXISTS FOR (m:Message) ON (m.timestamp)",
    }
    
    for _, query := range queries {
        if err := ltm.neo4j.Execute(query, nil); err != nil {
            return err
        }
    }
    
    return nil
}

// ===== Main Entry Point =====

func main() {
    // Create agent workspace
    workspace := NewAgentWorkspace()
    
    // Setup routes
    workspace.app.Get("/ws/chat", websocket.New(workspace.handleChatWebSocket))
    workspace.app.Get("/ws/a2a", websocket.New(workspace.handleA2AWebSocket))
    workspace.app.Get("/.well-known/agent.json", workspace.handleAgentCard)
    workspace.app.Static("/", "./public")
    
    // Initialize terminal (starts Neo4j, LightRAG, watchdog)
    go func() {
        time.Sleep(2 * time.Second) // Wait for server to start
        if err := workspace.InitializeTerminal(); err != nil {
            log.Fatalf("Failed to initialize terminal: %v", err)
        }
    }()
    
    // Start server
    log.Println("🚀 Starting Agent Workspace on :8080")
    log.Fatal(workspace.app.Listen(":8080"))
}
```

---

## Memory Management Strategy

### When to Use Long-Term Memory (LightRAG + Neo4j)

✅ **Store:**
- All conversation messages
- Code files and structure
- Terminal commands and outputs
- Concepts and relationships
- Patterns and learnings
- Task outcomes

✅ **Query:**
- "What did we discuss about authentication?"
- "Show me all files related to JWT"
- "What patterns have we used for database connections?"

### When to Use Short-Term Memory (ChromeDP)

✅ **Store:**
- Current task screenshots
- Browser page state (DOM, elements, cookies)
- Temporary task variables
- Active automation context

✅ **Clear:**
- After task completion
- On task cancellation
- On error/timeout

---

## Terminal Startup Sequence

```bash
$ neo4j start
Starting Neo4j...
✓ Neo4j started (bolt://localhost:7687)

$ Initializing LightRAG...
✓ Connected to Neo4j
✓ Vector storage (ChromeM) ready
✓ Key-value storage (BoltDB) ready
✓ LightRAG schema initialized

$ Starting watchdog monitoring...
✓ Watchdog active (monitoring patterns, concepts, dependencies)

🚀 Agent workspace ready!
```

---

## Example Task Flow

### User: "Navigate to GitHub and find the go-light-rag repository"

**1. Long-Term Memory Query:**
```
Query: "github go-light-rag repository"
Result: Previous conversations about LightRAG, repository URLs
```

**2. Agent Decision (Gemma 3):**
```json
[
  {"type": "browser_navigate", "params": {"url": "https://github.com"}},
  {"type": "browser_type", "params": {"element": 1, "text": "go-light-rag"}},
  {"type": "browser_click", "params": {"element": 3}}
]
```

**3. Execution with Short-Term Memory:**
- Navigate → Capture screenshot + page state
- Type → Update page state
- Click → Capture new screenshot + page state

**4. Store in Long-Term Memory:**
```
Message: "User requested to find go-light-rag on GitHub"
Action: "Navigated to https://github.com/MegaGrindStone/go-light-rag"
Outcome: "Successfully found repository"
```

**5. Clear Short-Term Memory:**
- Delete task screenshots
- Clear page states
- Remove temporary variables

---

## Conclusion

This complete system provides:

✅ **Long-term learning** via LightRAG + Neo4j knowledge graph  
✅ **Short-term efficiency** via ChromeDP context management  
✅ **Intelligent context** for Gemma 3 agent decisions  
✅ **Persistent knowledge** across sessions  
✅ **Efficient memory** for dynamic tasks  
✅ **Watchdog monitoring** for patterns and insights  
✅ **Midnight glassmorphism** aesthetic throughout  

The dual memory architecture ensures the agent learns from every interaction while maintaining efficient, task-focused short-term memory! 🚀

