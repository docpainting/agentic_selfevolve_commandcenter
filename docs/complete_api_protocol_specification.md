# Complete Agent Workspace - API & Protocol Specification

## Overview

This document defines the complete communication protocol and API specification for the agent workspace, integrating WebSocket real-time communication, JSON-RPC 2.0 A2A protocol, and RESTful HTTP endpoints.

---

## Communication Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend                                │
│  ┌──────────────────┬──────────────────┬──────────────────┐    │
│  │  WebSocket Chat  │  JSON-RPC 2.0    │  REST API        │    │
│  │  (Real-time)     │  (A2A Protocol)  │  (HTTP)          │    │
│  └──────────────────┴──────────────────┴──────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                              ↕
┌─────────────────────────────────────────────────────────────────┐
│                    Go Fiber v3 Backend                          │
│  ┌──────────────────┬──────────────────┬──────────────────┐    │
│  │  WS Handler      │  JSON-RPC        │  REST Routes     │    │
│  │  /ws/chat        │  /ws/a2a         │  /api/*          │    │
│  └──────────────────┴──────────────────┴──────────────────┘    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Agent Core + Tools + Memory                 │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 1. Core Message Protocol

### Universal Message Structure

All messages across WebSocket and JSON-RPC follow this base structure:

```typescript
interface BaseMessage {
  id: string;                    // UUID v4
  type: MessageType;
  timestamp: string;             // ISO-8601
  source: "agent" | "user" | "system" | "watchdog";
  payload: any;
  metadata: MessageMetadata;
}

interface MessageMetadata {
  session_id: string;
  agent_state: "idle" | "thinking" | "executing" | "awaiting_approval" | "error";
  priority: "low" | "normal" | "high" | "critical";
  task_id?: string;
}

type MessageType = 
  | "agent_action"
  | "user_command"
  | "system_event"
  | "watchdog_alert"
  | "approval_request"
  | "tool_result";
```

---

## 2. WebSocket Communication (Real-time Chat)

### Connection

```
Endpoint: wss://localhost:8080/ws/chat
Protocol: WebSocket
Authentication: JWT token in query parameter or header
```

### Message Types

#### 2.1 User Command

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "user_command",
  "timestamp": "2025-10-26T03:45:00.000Z",
  "source": "user",
  "payload": {
    "command": "Create a login page with JWT authentication",
    "context": "project",
    "urgency": "immediate",
    "attachments": []
  },
  "metadata": {
    "session_id": "sess_abc123",
    "agent_state": "idle",
    "priority": "normal"
  }
}
```

#### 2.2 Agent Action (Reasoning)

```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "type": "agent_action",
  "timestamp": "2025-10-26T03:45:01.000Z",
  "source": "agent",
  "payload": {
    "action_type": "reasoning",
    "content": "I'll create a login page with JWT authentication. First, I need to check if we have an authentication module...",
    "reasoning_steps": [
      "Check existing project structure",
      "Create auth.go with JWT functions",
      "Create login.html page",
      "Set up middleware"
    ]
  },
  "metadata": {
    "session_id": "sess_abc123",
    "agent_state": "thinking",
    "priority": "normal",
    "task_id": "task_xyz789"
  }
}
```

#### 2.3 Agent Action (Tool Call)

```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "type": "agent_action",
  "timestamp": "2025-10-26T03:45:02.000Z",
  "source": "agent",
  "payload": {
    "action_type": "tool_call",
    "content": "Creating authentication module",
    "tool_calls": [
      {
        "tool": "file_edit",
        "parameters": {
          "path": "/project/auth/auth.go",
          "operation": "create",
          "content": "package auth\n\nimport (...)"
        },
        "requires_approval": true
      }
    ]
  },
  "metadata": {
    "session_id": "sess_abc123",
    "agent_state": "awaiting_approval",
    "priority": "normal",
    "task_id": "task_xyz789"
  }
}
```

#### 2.4 System Event

```json
{
  "id": "880e8400-e29b-41d4-a716-446655440003",
  "type": "system_event",
  "timestamp": "2025-10-26T03:45:03.000Z",
  "source": "system",
  "payload": {
    "event": "file_modified",
    "data": {
      "path": "/project/auth/auth.go",
      "operation": "created",
      "size": 1024
    }
  },
  "metadata": {
    "session_id": "sess_abc123",
    "agent_state": "executing",
    "priority": "low"
  }
}
```

#### 2.5 Watchdog Alert

```json
{
  "id": "990e8400-e29b-41d4-a716-446655440004",
  "type": "watchdog_alert",
  "timestamp": "2025-10-26T03:45:04.000Z",
  "source": "watchdog",
  "payload": {
    "alert_type": "pattern_detected",
    "severity": "info",
    "title": "Authentication Pattern Detected",
    "description": "JWT authentication pattern is being implemented",
    "related_nodes": ["concept:jwt", "file:auth.go", "pattern:authentication"],
    "action_required": false
  },
  "metadata": {
    "session_id": "sess_abc123",
    "agent_state": "executing",
    "priority": "low"
  }
}
```

#### 2.6 Approval Request

```json
{
  "id": "aa0e8400-e29b-41d4-a716-446655440005",
  "type": "approval_request",
  "timestamp": "2025-10-26T03:45:05.000Z",
  "source": "agent",
  "payload": {
    "action": "file_edit",
    "description": "Create auth.go with JWT authentication functions",
    "details": {
      "path": "/project/auth/auth.go",
      "operation": "create",
      "preview": "package auth\n\nimport (\n\t\"github.com/golang-jwt/jwt\"..."
    },
    "options": ["approve", "reject", "modify"]
  },
  "metadata": {
    "session_id": "sess_abc123",
    "agent_state": "awaiting_approval",
    "priority": "high",
    "task_id": "task_xyz789"
  }
}
```

#### 2.7 Tool Result

```json
{
  "id": "bb0e8400-e29b-41d4-a716-446655440006",
  "type": "tool_result",
  "timestamp": "2025-10-26T03:45:06.000Z",
  "source": "system",
  "payload": {
    "tool": "terminal_execute",
    "command": "go test ./...",
    "output": "PASS\nok  \tproject/auth\t0.123s",
    "exit_code": 0,
    "success": true
  },
  "metadata": {
    "session_id": "sess_abc123",
    "agent_state": "executing",
    "priority": "normal",
    "task_id": "task_xyz789"
  }
}
```

---

## 3. JSON-RPC 2.0 Protocol (A2A Communication)

### Connection

```
Endpoint: wss://localhost:8080/ws/a2a
Protocol: WebSocket with JSON-RPC 2.0
Authentication: Agent Card + OAuth
```

### Core A2A Methods

#### 3.1 agent/getAuthenticatedExtendedCard

```json
// Request
{
  "jsonrpc": "2.0",
  "id": "req-001",
  "method": "agent/getAuthenticatedExtendedCard",
  "params": {}
}

// Response
{
  "jsonrpc": "2.0",
  "id": "req-001",
  "result": {
    "name": "Gemma3 Agent Workspace",
    "description": "AI agent with browser automation, terminal control, and MCP integration",
    "version": "1.0.0",
    "capabilities": {
      "streaming": true,
      "file_operations": true,
      "browser_automation": true,
      "memory_graph": true
    },
    "skills": [
      {
        "name": "web_browsing",
        "description": "Automated web browsing and data extraction"
      },
      {
        "name": "code_generation",
        "description": "Generate and edit code files"
      },
      {
        "name": "terminal_control",
        "description": "Execute terminal commands"
      },
      {
        "name": "memory_management",
        "description": "Store and retrieve from knowledge graph"
      }
    ],
    "url": "wss://localhost:8080/ws/a2a",
    "transport": "JSONRPC"
  }
}
```

#### 3.2 message/send

```json
// Request
{
  "jsonrpc": "2.0",
  "id": "req-002",
  "method": "message/send",
  "params": {
    "message": {
      "role": "user",
      "parts": [
        {
          "type": "text",
          "text": "Create a REST API endpoint for user authentication"
        }
      ]
    },
    "configuration": {
      "streaming": true
    }
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": "req-002",
  "result": {
    "taskId": "task-12345",
    "status": {
      "state": "working",
      "message": "Processing your request..."
    }
  }
}
```

#### 3.3 tasks/get

```json
// Request
{
  "jsonrpc": "2.0",
  "id": "req-003",
  "method": "tasks/get",
  "params": {
    "taskId": "task-12345"
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": "req-003",
  "result": {
    "id": "task-12345",
    "status": {
      "state": "completed",
      "message": "REST API endpoint created successfully"
    },
    "messages": [
      {
        "role": "agent",
        "parts": [
          {
            "type": "text",
            "text": "Created /api/auth/login endpoint with JWT token generation"
          }
        ]
      }
    ],
    "artifacts": [
      {
        "name": "auth_handler.go",
        "mimeType": "text/x-go",
        "parts": [
          {
            "type": "text",
            "text": "package handlers\n\n..."
          }
        ]
      }
    ]
  }
}
```

#### 3.4 Custom Methods (Browser Control)

```json
// browser/navigate
{
  "jsonrpc": "2.0",
  "id": "req-004",
  "method": "browser/navigate",
  "params": {
    "url": "https://github.com",
    "waitUntil": "networkidle"
  }
}

// browser/click
{
  "jsonrpc": "2.0",
  "id": "req-005",
  "method": "browser/click",
  "params": {
    "element": 5,
    "waitForNavigation": true
  }
}

// browser/screenshot
{
  "jsonrpc": "2.0",
  "id": "req-006",
  "method": "browser/screenshot",
  "params": {
    "fullPage": false,
    "format": "png"
  }
}
```

---

## 4. REST API Endpoints

### 4.1 Agent Control

```
POST /api/agent/initialize
Body: {
  "session_id": "string",
  "config": {
    "model": "gemma-3",
    "temperature": 0.7
  }
}
Response: {
  "session_id": "string",
  "agent_state": "idle",
  "capabilities": []
}

POST /api/agent/command
Body: {
  "command": "string",
  "context": "project|file|system"
}
Response: {
  "task_id": "string",
  "status": "queued|processing"
}

GET /api/agent/status
Response: {
  "state": "idle|thinking|executing|awaiting_approval",
  "current_task": "string",
  "queue_length": 0
}

POST /api/agent/pause
Response: {
  "success": true,
  "previous_state": "executing"
}

POST /api/agent/resume
Response: {
  "success": true,
  "current_state": "executing"
}
```

### 4.2 File System Operations

```
GET /api/files/tree?path=/project
Response: {
  "path": "/project",
  "type": "directory",
  "children": [
    {
      "name": "auth",
      "type": "directory",
      "children": [...]
    }
  ]
}

GET /api/files/content?path=/project/auth/auth.go
Response: {
  "path": "/project/auth/auth.go",
  "content": "package auth...",
  "language": "go",
  "size": 1024,
  "modified": "2025-10-26T03:45:00.000Z"
}

PUT /api/files/content
Body: {
  "path": "/project/auth/auth.go",
  "content": "package auth...",
  "requires_approval": true
}
Response: {
  "success": true,
  "approval_id": "approval-123",
  "status": "pending_approval"
}

POST /api/files/diff
Body: {
  "path": "/project/auth/auth.go",
  "diff": "--- a/auth.go\n+++ b/auth.go\n@@ -10,3 +10,5 @@..."
}
Response: {
  "success": true,
  "lines_added": 5,
  "lines_removed": 2
}

GET /api/files/search?query=jwt&type=code
Response: {
  "results": [
    {
      "path": "/project/auth/auth.go",
      "line": 15,
      "content": "func GenerateJWT(user User) string {",
      "relevance": 0.95
    }
  ]
}
```

### 4.3 Memory System

```
POST /api/memory/query
Body: {
  "query": "What authentication patterns have we used?",
  "mode": "hybrid|local|global"
}
Response: {
  "entities": ["jwt", "oauth", "session"],
  "relationships": [
    {
      "from": "jwt",
      "to": "authentication",
      "type": "IMPLEMENTS"
    }
  ],
  "sources": [
    {
      "id": "doc-123",
      "content": "Previous conversation about JWT...",
      "relevance": 0.89
    }
  ]
}

POST /api/memory/store
Body: {
  "type": "conversation|code|concept",
  "content": "string",
  "metadata": {}
}
Response: {
  "success": true,
  "node_id": "node-456"
}

GET /api/memory/context?task_id=task-12345
Response: {
  "task_id": "task-12345",
  "related_concepts": ["authentication", "jwt", "security"],
  "previous_conversations": [...],
  "code_references": [...]
}

POST /api/memory/vector-search
Body: {
  "query": "authentication implementation",
  "top_k": 10
}
Response: {
  "results": [
    {
      "id": "doc-789",
      "content": "JWT authentication with refresh tokens...",
      "similarity": 0.92
    }
  ]
}
```

### 4.4 OpenEvolve Integration

```
GET /api/openevolve/status
Response: {
  "components": [
    {
      "name": "Component A",
      "status": "approved",
      "progress": 100
    }
  ],
  "overall_progress": 65,
  "active_watchdogs": 3
}

POST /api/openevolve/proposal
Body: {
  "component": "Authentication Module",
  "optimization": "Use bcrypt instead of plain SHA256",
  "rationale": "Better security against rainbow tables"
}
Response: {
  "proposal_id": "prop-123",
  "status": "pending_review"
}

PUT /api/openevolve/rewards
Body: {
  "component": "Component A",
  "reward": 0.95,
  "metrics": {
    "performance": 0.9,
    "security": 1.0,
    "maintainability": 0.9
  }
}
Response: {
  "success": true,
  "updated_score": 0.95
}

GET /api/openevolve/metrics
Response: {
  "overall_health": 0.87,
  "code_quality": 0.92,
  "test_coverage": 0.85,
  "security_score": 0.90
}
```

### 4.5 Vision System (Browser Screenshots)

```
POST /api/vision/analyze
Body: {
  "image": "base64_encoded_screenshot",
  "task": "identify_clickable_elements"
}
Response: {
  "elements": [
    {
      "index": 1,
      "type": "button",
      "text": "Sign In",
      "position": {"x": 100, "y": 50, "width": 80, "height": 30}
    }
  ]
}

GET /api/vision/state?task_id=task-12345
Response: {
  "task_id": "task-12345",
  "current_url": "https://github.com",
  "screenshot": "base64...",
  "elements": [...],
  "timestamp": "2025-10-26T03:45:00.000Z"
}

POST /api/vision/context
Body: {
  "task_id": "task-12345",
  "context": "Looking for repository search box"
}
Response: {
  "success": true,
  "suggested_element": 3
}
```

---

## 5. Real-time Communication Flow

### Sequence Diagram

```
User → Frontend: Send command "Create login page"
Frontend → Backend (WS): user_command message
Backend → Agent Core: Process command
Agent Core → Agent Core: Internal reasoning
Agent Core → Backend: Stream reasoning updates
Backend → Frontend (WS): agent_action (reasoning)
Frontend → User: Display reasoning in chat

Agent Core → Tools: Execute file_edit
Tools → LightRAG: Store code in memory
LightRAG → Neo4j: Create nodes/relationships
Tools → Backend: Tool execution result
Backend → Frontend (WS): system_event (file_modified)
Frontend → User: Update file tree

Agent Core → Backend: Action requiring approval
Backend → Frontend (WS): approval_request
Frontend → User: Show approval dialog
User → Frontend: Approve action
Frontend → Backend (WS): approval_response
Backend → Agent Core: Continue execution

Watchdog → Neo4j: Detect pattern
Watchdog → Backend: Pattern detected
Backend → Frontend (WS): watchdog_alert
Frontend → User: Show alert in OpenEvolve panel
```

---

## 6. Performance Optimization

### Message Compression

```go
// Use MessagePack for binary serialization
import "github.com/vmihailenco/msgpack/v5"

type CompressedMessage struct {
    Data []byte
}

func CompressMessage(msg Message) ([]byte, error) {
    return msgpack.Marshal(msg)
}

func DecompressMessage(data []byte) (*Message, error) {
    var msg Message
    err := msgpack.Unmarshal(data, &msg)
    return &msg, err
}
```

### Delta Updates for File Changes

```go
type FileDelta struct {
    Path      string
    Operation string // "insert", "delete", "replace"
    LineStart int
    LineEnd   int
    Content   string
}

// Send only changed lines instead of full file
func SendFileDelta(ws *websocket.Conn, delta FileDelta) {
    ws.WriteJSON(delta)
}
```

### Connection Management

```javascript
class ResilientWebSocket {
  constructor(url) {
    this.url = url;
    this.reconnectDelay = 1000;
    this.maxReconnectDelay = 30000;
    this.heartbeatInterval = 30000;
    this.connect();
  }
  
  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onopen = () => {
      console.log('Connected');
      this.reconnectDelay = 1000;
      this.startHeartbeat();
      this.restoreSession();
    };
    
    this.ws.onclose = () => {
      console.log('Disconnected, reconnecting...');
      this.stopHeartbeat();
      setTimeout(() => this.connect(), this.reconnectDelay);
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
    };
  }
  
  startHeartbeat() {
    this.heartbeat = setInterval(() => {
      this.ws.send(JSON.stringify({ type: 'ping' }));
    }, this.heartbeatInterval);
  }
  
  stopHeartbeat() {
    if (this.heartbeat) {
      clearInterval(this.heartbeat);
    }
  }
  
  restoreSession() {
    const sessionId = localStorage.getItem('session_id');
    if (sessionId) {
      this.ws.send(JSON.stringify({
        type: 'restore_session',
        payload: { session_id: sessionId }
      }));
    }
  }
}
```

### Caching Strategy

```go
type CacheManager struct {
    fileCache    *cache.Cache // TTL cache for file content
    memoryCache  *cache.Cache // Query result cache
    contextCache *cache.Cache // Agent context cache
}

func (cm *CacheManager) GetFile(path string) (string, bool) {
    if content, found := cm.fileCache.Get(path); found {
        return content.(string), true
    }
    return "", false
}

func (cm *CacheManager) SetFile(path, content string, version int) {
    cm.fileCache.Set(path, content, cache.DefaultExpiration)
}

func (cm *CacheManager) InvalidateFile(path string) {
    cm.fileCache.Delete(path)
}
```

---

## 7. Security Considerations

### Authentication & Authorization

```go
// JWT token generation
func GenerateJWT(userID string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
        "iat":     time.Now().Unix(),
    })
    
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// Middleware for WebSocket authentication
func AuthenticateWebSocket(c *fiber.Ctx) error {
    token := c.Query("token")
    if token == "" {
        return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
    }
    
    claims, err := ValidateJWT(token)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
    }
    
    c.Locals("user_id", claims["user_id"])
    return c.Next()
}

// Role-based permissions
type Permission string

const (
    PermissionReadFiles  Permission = "read_files"
    PermissionWriteFiles Permission = "write_files"
    PermissionExecute    Permission = "execute_commands"
)

func CheckPermission(userID string, perm Permission) bool {
    // Check user permissions from database
    return true
}
```

### Input Validation

```go
// Schema validation for messages
func ValidateMessage(msg BaseMessage) error {
    if msg.ID == "" {
        return errors.New("message ID is required")
    }
    
    if _, err := uuid.Parse(msg.ID); err != nil {
        return errors.New("invalid message ID format")
    }
    
    if msg.Type == "" {
        return errors.New("message type is required")
    }
    
    return nil
}

// File path sanitization
func SanitizePath(path string) (string, error) {
    // Prevent directory traversal
    if strings.Contains(path, "..") {
        return "", errors.New("invalid path: directory traversal detected")
    }
    
    // Ensure path is within project directory
    absPath, err := filepath.Abs(path)
    if err != nil {
        return "", err
    }
    
    projectRoot := "/project"
    if !strings.HasPrefix(absPath, projectRoot) {
        return "", errors.New("path outside project directory")
    }
    
    return absPath, nil
}
```

### Rate Limiting

```go
type RateLimiter struct {
    limits map[string]*rate.Limiter
    mu     sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
    return &RateLimiter{
        limits: make(map[string]*rate.Limiter),
    }
}

func (rl *RateLimiter) GetLimiter(sessionID string) *rate.Limiter {
    rl.mu.RLock()
    limiter, exists := rl.limits[sessionID]
    rl.mu.RUnlock()
    
    if !exists {
        rl.mu.Lock()
        limiter = rate.NewLimiter(rate.Limit(10), 20) // 10 req/s, burst 20
        rl.limits[sessionID] = limiter
        rl.mu.Unlock()
    }
    
    return limiter
}

func (rl *RateLimiter) Allow(sessionID string) bool {
    return rl.GetLimiter(sessionID).Allow()
}
```

---

## 8. Complete Backend Implementation

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/websocket/v3"
)

type Server struct {
    app          *fiber.App
    wsChat       *WebSocketHandler
    wsA2A        *A2AHandler
    rateLimiter  *RateLimiter
    cacheManager *CacheManager
}

func NewServer() *Server {
    app := fiber.New(fiber.Config{
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    })
    
    return &Server{
        app:          app,
        wsChat:       NewWebSocketHandler(),
        wsA2A:        NewA2AHandler(),
        rateLimiter:  NewRateLimiter(),
        cacheManager: NewCacheManager(),
    }
}

func (s *Server) SetupRoutes() {
    // WebSocket endpoints
    s.app.Get("/ws/chat", AuthenticateWebSocket, websocket.New(s.wsChat.Handle))
    s.app.Get("/ws/a2a", AuthenticateWebSocket, websocket.New(s.wsA2A.Handle))
    
    // Agent control
    api := s.app.Group("/api", s.rateLimitMiddleware)
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
    
    // Agent Card
    s.app.Get("/.well-known/agent.json", s.handleAgentCard)
    
    // Static files
    s.app.Static("/", "./public")
}

func (s *Server) rateLimitMiddleware(c *fiber.Ctx) error {
    sessionID := c.Locals("user_id").(string)
    
    if !s.rateLimiter.Allow(sessionID) {
        return c.Status(429).JSON(fiber.Map{
            "error": "rate limit exceeded",
        })
    }
    
    return c.Next()
}

func main() {
    server := NewServer()
    server.SetupRoutes()
    
    log.Fatal(server.app.Listen(":8080"))
}
```

---

## Conclusion

This complete API and protocol specification provides:

✅ **Unified message protocol** across all communication channels  
✅ **WebSocket real-time** chat with structured messages  
✅ **JSON-RPC 2.0 A2A** for agent-to-agent communication  
✅ **RESTful HTTP API** for stateless operations  
✅ **Performance optimizations** (compression, caching, deltas)  
✅ **Security measures** (auth, validation, rate limiting)  
✅ **Complete implementation** examples in Go  

This architecture enables seamless communication between the frontend, backend, agent core, and all integrated tools! 🚀

