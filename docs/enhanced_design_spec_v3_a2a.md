# Enhanced UI/UX Design Specification - Midnight Glassmorphism with A2A Protocol (JSON-RPC 2.0)

## Design Philosophy Enhancement

This enhanced specification builds upon the original midnight glassmorphism theme by adding **interactive browser and terminal takeover capabilities** similar to Manus computer interface, with **full Agent-to-Agent (A2A) protocol compliance using JSON-RPC 2.0** for standardized agent communication. The design maintains visual consistency while enabling seamless transitions between AI-driven automation and manual control, with proper A2A-compliant agent discovery, task management, and multi-agent collaboration.

---

## Architecture Overview - A2A Compliant

```
┌─────────────────────────────────────────────────────────────┐
│                Frontend (React/Vue) - A2A Client            │
│                  Midnight Glassmorphism UI                  │
│  ┌──────────────┬──────────────┬──────────────────────────┐ │
│  │ Left Panel   │ Center Chat  │ Right Panel              │ │
│  │ File Tree    │ Messages     │ Agent Discovery          │ │
│  │ Knowledge    │ Task Status  │ MCP Integration          │ │
│  └──────────────┴──────────────┴──────────────────────────┘ │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ Bottom Dock: Terminal | Browser | MCP Tools | Logs      │ │
│  └──────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                    ↕ WebSocket (JSON-RPC 2.0)
┌─────────────────────────────────────────────────────────────┐
│         Go Fiber v3 Backend - A2A Server/Client             │
│  ┌──────────────┬──────────────┬──────────────────────────┐ │
│  │ JSON-RPC 2.0 │ A2A Server   │ A2A Client               │ │
│  │ Handler      │ (This Agent) │ (Remote Agents)          │ │
│  └──────────────┴──────────────┴──────────────────────────┘ │
│  ┌──────────────┬──────────────┬──────────────────────────┐ │
│  │ Browser      │ Terminal     │ MCP Client               │ │
│  │ Automation   │ PTY Manager  │ Integration              │ │
│  │ (Playwright) │              │                          │ │
│  └──────────────┴──────────────┴──────────────────────────┘ │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ Gemma 3 Agent Controller (Task Orchestration)           │ │
│  └──────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                    ↕ JSON-RPC 2.0 over HTTPS
┌─────────────────────────────────────────────────────────────┐
│                   Remote A2A Agents                         │
│  • Other A2A-compliant agents                              │
│  • MCP Servers (via A2A bridge)                            │
│  • External services with Agent Cards                      │
└─────────────────────────────────────────────────────────────┘
```

---

## JSON-RPC 2.0 Message Format (A2A Compliant)

### Request Structure

```json
{
  "jsonrpc": "2.0",
  "id": "unique-request-id",
  "method": "message/send",
  "params": {
    "message": {
      "role": "user",
      "parts": [
        {
          "type": "text",
          "text": "Navigate to example.com and extract the title"
        }
      ]
    },
    "configuration": {
      "streaming": true
    }
  }
}
```

### Response Structure

```json
{
  "jsonrpc": "2.0",
  "id": "unique-request-id",
  "result": {
    "taskId": "task-12345",
    "status": {
      "state": "working",
      "message": "Processing your request..."
    }
  }
}
```

### Error Response

```json
{
  "jsonrpc": "2.0",
  "id": "unique-request-id",
  "error": {
    "code": -32001,
    "message": "Task not found",
    "data": {
      "taskId": "task-12345"
    }
  }
}
```

### Notification (No Response Expected)

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/status",
  "params": {
    "taskId": "task-12345",
    "status": {
      "state": "completed",
      "message": "Task completed successfully"
    }
  }
}
```

---

## A2A Protocol Methods (JSON-RPC 2.0)

### Core Methods

| Method | Description | Request/Response |
|--------|-------------|------------------|
| `message/send` | Send a message to start or continue a task | Request-Response |
| `message/stream` | Stream task updates via SSE | Streaming |
| `tasks/get` | Retrieve task details | Request-Response |
| `tasks/list` | List all tasks | Request-Response |
| `tasks/cancel` | Cancel a running task | Request-Response |
| `agent/getAuthenticatedExtendedCard` | Get agent capabilities | Request-Response |

---

## Frontend WebSocket Client (JSON-RPC 2.0)

```javascript
class A2AWebSocketClient {
  constructor(url = 'wss://localhost:8080/ws') {
    this.url = url;
    this.ws = null;
    this.reconnectInterval = 3000;
    this.requestId = 0;
    this.pendingRequests = new Map(); // id -> {resolve, reject, timeout}
    this.notificationHandlers = new Map(); // method -> handler
    this.connectionStatus = 'disconnected';
    this.listeners = [];
  }
  
  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onopen = () => {
      console.log('A2A WebSocket connected');
      this.connectionStatus = 'connected';
      this.notifyStatusChange('connected');
      this.updateConnectionIndicator('connected');
      
      // Fetch agent card on connect
      this.getAgentCard();
    };
    
    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleJSONRPC(message);
    };
    
    this.ws.onerror = (error) => {
      console.error('A2A WebSocket error:', error);
      this.connectionStatus = 'error';
      this.notifyStatusChange('error');
      this.updateConnectionIndicator('error');
    };
    
    this.ws.onclose = () => {
      console.log('A2A WebSocket disconnected');
      this.connectionStatus = 'disconnected';
      this.notifyStatusChange('disconnected');
      this.updateConnectionIndicator('disconnected');
      
      // Auto-reconnect
      setTimeout(() => this.connect(), this.reconnectInterval);
    };
  }
  
  // Generate unique request ID
  generateId() {
    return `req-${++this.requestId}-${Date.now()}`;
  }
  
  // Send JSON-RPC 2.0 request and return Promise
  request(method, params, timeout = 30000) {
    return new Promise((resolve, reject) => {
      if (this.ws.readyState !== WebSocket.OPEN) {
        reject(new Error('WebSocket not connected'));
        return;
      }
      
      const id = this.generateId();
      const request = {
        jsonrpc: '2.0',
        id,
        method,
        params
      };
      
      // Store pending request
      const timeoutId = setTimeout(() => {
        this.pendingRequests.delete(id);
        reject(new Error(`Request timeout: ${method}`));
      }, timeout);
      
      this.pendingRequests.set(id, { resolve, reject, timeout: timeoutId });
      
      // Send request
      this.ws.send(JSON.stringify(request));
      
      console.log('→ JSON-RPC Request:', method, params);
    });
  }
  
  // Send JSON-RPC 2.0 notification (no response expected)
  notify(method, params) {
    if (this.ws.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected');
      return;
    }
    
    const notification = {
      jsonrpc: '2.0',
      method,
      params
    };
    
    this.ws.send(JSON.stringify(notification));
    console.log('→ JSON-RPC Notification:', method, params);
  }
  
  // Handle incoming JSON-RPC 2.0 messages
  handleJSONRPC(message) {
    // Response to a request
    if (message.id && this.pendingRequests.has(message.id)) {
      const pending = this.pendingRequests.get(message.id);
      clearTimeout(pending.timeout);
      this.pendingRequests.delete(message.id);
      
      if (message.error) {
        console.error('← JSON-RPC Error:', message.error);
        pending.reject(new Error(message.error.message));
      } else {
        console.log('← JSON-RPC Response:', message.result);
        pending.resolve(message.result);
      }
    }
    // Notification from server (no id)
    else if (!message.id && message.method) {
      console.log('← JSON-RPC Notification:', message.method, message.params);
      this.handleNotification(message.method, message.params);
    }
    // Invalid message
    else {
      console.warn('Invalid JSON-RPC message:', message);
    }
  }
  
  // Handle server notifications
  handleNotification(method, params) {
    if (this.notificationHandlers.has(method)) {
      this.notificationHandlers.get(method)(params);
    }
    
    // Built-in notification handlers
    switch (method) {
      case 'tasks/status':
        this.handleTaskStatus(params);
        break;
      case 'browser/action':
        this.handleBrowserAction(params);
        break;
      case 'terminal/output':
        this.handleTerminalOutput(params);
        break;
      case 'mcp/toolResult':
        this.handleMCPToolResult(params);
        break;
      default:
        console.log('Unhandled notification:', method);
    }
  }
  
  // Register notification handler
  on(method, handler) {
    this.notificationHandlers.set(method, handler);
  }
  
  // ===== A2A Protocol Methods =====
  
  // Get agent card
  async getAgentCard() {
    try {
      const result = await this.request('agent/getAuthenticatedExtendedCard', {});
      console.log('Agent Card:', result);
      this.displayAgentCard(result);
      return result;
    } catch (error) {
      console.error('Failed to get agent card:', error);
      throw error;
    }
  }
  
  // Send message to agent
  async sendMessage(text, streaming = true) {
    try {
      const result = await this.request('message/send', {
        message: {
          role: 'user',
          parts: [
            {
              type: 'text',
              text: text
            }
          ]
        },
        configuration: {
          streaming: streaming
        }
      });
      
      console.log('Message sent, task ID:', result.taskId);
      this.displayTaskStatus(result.taskId, result.status);
      return result;
    } catch (error) {
      console.error('Failed to send message:', error);
      throw error;
    }
  }
  
  // Get task details
  async getTask(taskId) {
    try {
      const result = await this.request('tasks/get', {
        taskId: taskId
      });
      
      console.log('Task details:', result);
      return result;
    } catch (error) {
      console.error('Failed to get task:', error);
      throw error;
    }
  }
  
  // List all tasks
  async listTasks(limit = 50, offset = 0) {
    try {
      const result = await this.request('tasks/list', {
        limit: limit,
        offset: offset
      });
      
      console.log('Tasks:', result.tasks);
      this.displayTaskList(result.tasks);
      return result;
    } catch (error) {
      console.error('Failed to list tasks:', error);
      throw error;
    }
  }
  
  // Cancel task
  async cancelTask(taskId) {
    try {
      const result = await this.request('tasks/cancel', {
        taskId: taskId
      });
      
      console.log('Task cancelled:', taskId);
      return result;
    } catch (error) {
      console.error('Failed to cancel task:', error);
      throw error;
    }
  }
  
  // ===== Browser Control Methods =====
  
  async browserNavigate(url) {
    return this.request('browser/navigate', { url });
  }
  
  async browserClick(elementNumber) {
    return this.request('browser/click', { element: elementNumber });
  }
  
  async browserType(elementNumber, text) {
    return this.request('browser/type', { element: elementNumber, text });
  }
  
  async browserScreenshot() {
    return this.request('browser/screenshot', {});
  }
  
  // ===== Terminal Control Methods =====
  
  async terminalExecute(command) {
    return this.request('terminal/execute', { command });
  }
  
  terminalInput(input) {
    this.notify('terminal/input', { input });
  }
  
  // ===== MCP Control Methods =====
  
  async mcpListTools(server) {
    return this.request('mcp/listTools', { server });
  }
  
  async mcpCallTool(server, tool, args) {
    return this.request('mcp/callTool', { server, tool, args });
  }
  
  // ===== Takeover Control =====
  
  toggleBrowserTakeover() {
    this.notify('takeover/browser', {});
  }
  
  toggleTerminalTakeover() {
    this.notify('takeover/terminal', {});
  }
  
  // ===== UI Update Handlers =====
  
  handleTaskStatus(params) {
    const { taskId, status } = params;
    this.displayTaskStatus(taskId, status);
    
    // Update task status in UI
    const taskElement = document.querySelector(`[data-task-id="${taskId}"]`);
    if (taskElement) {
      taskElement.querySelector('.task-state').textContent = status.state;
      taskElement.querySelector('.task-message').textContent = status.message;
      taskElement.className = `task-item ${status.state}`;
    }
  }
  
  handleBrowserAction(params) {
    const { action, target, value } = params;
    console.log(`Browser action: ${action} on ${target}`);
    
    // Show AI command parser
    this.showAICommandParser(`${action.toUpperCase()} ${target}`, value);
  }
  
  handleTerminalOutput(params) {
    const { output, attribution } = params;
    
    // Add output to terminal
    const outputDiv = document.createElement('div');
    outputDiv.className = 'terminal-output';
    outputDiv.textContent = output;
    
    document.querySelector('.terminal-content').appendChild(outputDiv);
    
    // Auto-scroll
    const terminal = document.querySelector('.terminal-content');
    terminal.scrollTop = terminal.scrollHeight;
  }
  
  handleMCPToolResult(params) {
    const { server, tool, result, success } = params;
    
    // Add to MCP activity timeline
    const item = document.createElement('div');
    item.className = 'mcp-activity-item';
    item.textContent = `${server}.${tool} → ${success ? 'Success' : 'Failed'}`;
    item.style.borderLeftColor = success ? '#15A7FF' : '#FF2A6D';
    
    document.querySelector('.mcp-activity-timeline').prepend(item);
  }
  
  displayAgentCard(card) {
    const panel = document.querySelector('.agent-card-panel');
    if (panel) {
      panel.innerHTML = `
        <div class="agent-card">
          <h3>${card.name}</h3>
          <p>${card.description}</p>
          <div class="agent-capabilities">
            <strong>Skills:</strong>
            ${card.skills.map(s => `<span class="skill-badge">${s.name}</span>`).join('')}
          </div>
        </div>
      `;
    }
  }
  
  displayTaskStatus(taskId, status) {
    // Update or create task status display
    let taskElement = document.querySelector(`[data-task-id="${taskId}"]`);
    
    if (!taskElement) {
      taskElement = document.createElement('div');
      taskElement.className = `task-item ${status.state}`;
      taskElement.setAttribute('data-task-id', taskId);
      taskElement.innerHTML = `
        <div class="task-header">
          <span class="task-id">${taskId}</span>
          <span class="task-state">${status.state}</span>
        </div>
        <div class="task-message">${status.message || ''}</div>
      `;
      
      document.querySelector('.task-list').prepend(taskElement);
    }
  }
  
  displayTaskList(tasks) {
    const listContainer = document.querySelector('.task-list');
    if (listContainer) {
      listContainer.innerHTML = tasks.map(task => `
        <div class="task-item ${task.status.state}" data-task-id="${task.id}">
          <div class="task-header">
            <span class="task-id">${task.id}</span>
            <span class="task-state">${task.status.state}</span>
          </div>
          <div class="task-message">${task.status.message || ''}</div>
        </div>
      `).join('');
    }
  }
  
  showAICommandParser(command, description) {
    const parser = document.querySelector('.ai-command-parser');
    if (parser) {
      parser.querySelector('.command-text').textContent = command;
      parser.querySelector('.command-description').textContent = description || '';
      parser.classList.add('active');
      
      setTimeout(() => parser.classList.remove('active'), 3000);
    }
  }
  
  updateConnectionIndicator(status) {
    const indicator = document.querySelector('.ws-connection-indicator');
    if (indicator) {
      indicator.className = `ws-connection-indicator ${status}`;
      indicator.querySelector('.status-text').textContent = 
        status === 'connected' ? 'A2A Connected' : 
        status === 'error' ? 'Connection Error' : 
        'Disconnected';
    }
  }
  
  notifyStatusChange(status) {
    this.listeners.forEach(listener => listener(status));
  }
  
  addStatusListener(callback) {
    this.listeners.push(callback);
  }
}

// Initialize A2A WebSocket client
const a2aClient = new A2AWebSocketClient('wss://localhost:8080/ws');
a2aClient.connect();

// Example usage
a2aClient.on('tasks/status', (params) => {
  console.log('Task status update:', params);
});
```

---

## Go Fiber v3 Backend - A2A Server Implementation

```go
package main

import (
    "encoding/json"
    "log"
    "sync"
    "time"
    
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/websocket/v3"
    "github.com/google/uuid"
)

// ===== JSON-RPC 2.0 Structures =====

type JSONRPCRequest struct {
    JSONRPC string                 `json:"jsonrpc"`
    ID      interface{}            `json:"id,omitempty"`
    Method  string                 `json:"method"`
    Params  map[string]interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id,omitempty"`
    Result  interface{} `json:"result,omitempty"`
    Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type JSONRPCNotification struct {
    JSONRPC string                 `json:"jsonrpc"`
    Method  string                 `json:"method"`
    Params  map[string]interface{} `json:"params"`
}

// ===== A2A Protocol Structures =====

type Message struct {
    Role  string `json:"role"`
    Parts []Part `json:"parts"`
}

type Part struct {
    Type string `json:"type"`
    Text string `json:"text,omitempty"`
}

type TaskStatus struct {
    State   string `json:"state"` // "pending", "working", "completed", "failed", "cancelled"
    Message string `json:"message,omitempty"`
}

type Task struct {
    ID        string     `json:"id"`
    Status    TaskStatus `json:"status"`
    Messages  []Message  `json:"messages"`
    CreatedAt time.Time  `json:"createdAt"`
    UpdatedAt time.Time  `json:"updatedAt"`
}

type AgentCard struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Version     string   `json:"version"`
    Skills      []Skill  `json:"skills"`
    URL         string   `json:"url"`
    Transport   string   `json:"transport"`
}

type Skill struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

// ===== A2A Server =====

type A2AServer struct {
    app              *fiber.App
    browserManager   *BrowserManager
    terminalManager  *TerminalManager
    mcpClient        *MCPClient
    agentController  *AgentController
    clients          map[*websocket.Conn]bool
    clientsMux       sync.RWMutex
    tasks            map[string]*Task
    tasksMux         sync.RWMutex
    methodHandlers   map[string]JSONRPCHandler
}

type JSONRPCHandler func(params map[string]interface{}) (interface{}, *JSONRPCError)

func NewA2AServer() *A2AServer {
    server := &A2AServer{
        app:             fiber.New(),
        browserManager:  NewBrowserManager(),
        terminalManager: NewTerminalManager(),
        mcpClient:       NewMCPClient(),
        agentController: NewAgentController(),
        clients:         make(map[*websocket.Conn]bool),
        tasks:           make(map[string]*Task),
        methodHandlers:  make(map[string]JSONRPCHandler),
    }
    
    server.registerHandlers()
    return server
}

func (s *A2AServer) registerHandlers() {
    // A2A Protocol methods
    s.methodHandlers["agent/getAuthenticatedExtendedCard"] = s.handleGetAgentCard
    s.methodHandlers["message/send"] = s.handleMessageSend
    s.methodHandlers["tasks/get"] = s.handleTasksGet
    s.methodHandlers["tasks/list"] = s.handleTasksList
    s.methodHandlers["tasks/cancel"] = s.handleTasksCancel
    
    // Browser control methods
    s.methodHandlers["browser/navigate"] = s.handleBrowserNavigate
    s.methodHandlers["browser/click"] = s.handleBrowserClick
    s.methodHandlers["browser/type"] = s.handleBrowserType
    s.methodHandlers["browser/screenshot"] = s.handleBrowserScreenshot
    
    // Terminal control methods
    s.methodHandlers["terminal/execute"] = s.handleTerminalExecute
    
    // MCP control methods
    s.methodHandlers["mcp/listTools"] = s.handleMCPListTools
    s.methodHandlers["mcp/callTool"] = s.handleMCPCallTool
}

func (s *A2AServer) Setup() {
    // WebSocket endpoint for JSON-RPC 2.0
    s.app.Get("/ws", websocket.New(s.handleWebSocket))
    
    // HTTP endpoint for JSON-RPC 2.0 (optional)
    s.app.Post("/jsonrpc", s.handleHTTPJSONRPC)
    
    // Agent Card endpoint
    s.app.Get("/.well-known/agent.json", s.handleAgentCard)
    
    // Static files
    s.app.Static("/", "./public")
}

func (s *A2AServer) handleWebSocket(c *websocket.Conn) {
    // Register client
    s.clientsMux.Lock()
    s.clients[c] = true
    s.clientsMux.Unlock()
    
    defer func() {
        s.clientsMux.Lock()
        delete(s.clients, c)
        s.clientsMux.Unlock()
        c.Close()
    }()
    
    log.Println("A2A WebSocket client connected")
    
    // Message loop
    for {
        var req JSONRPCRequest
        if err := c.ReadJSON(&req); err != nil {
            log.Println("WebSocket read error:", err)
            break
        }
        
        // Handle JSON-RPC request
        response := s.handleJSONRPCRequest(req)
        
        // Send response (only if request has ID)
        if req.ID != nil {
            if err := c.WriteJSON(response); err != nil {
                log.Println("WebSocket write error:", err)
                break
            }
        }
    }
}

func (s *A2AServer) handleJSONRPCRequest(req JSONRPCRequest) JSONRPCResponse {
    log.Printf("JSON-RPC Request: %s %v", req.Method, req.Params)
    
    // Check if handler exists
    handler, exists := s.methodHandlers[req.Method]
    if !exists {
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      req.ID,
            Error: &JSONRPCError{
                Code:    -32601,
                Message: "Method not found",
                Data:    req.Method,
            },
        }
    }
    
    // Call handler
    result, err := handler(req.Params)
    
    if err != nil {
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      req.ID,
            Error:   err,
        }
    }
    
    return JSONRPCResponse{
        JSONRPC: "2.0",
        ID:      req.ID,
        Result:  result,
    }
}

// ===== A2A Protocol Method Handlers =====

func (s *A2AServer) handleGetAgentCard(params map[string]interface{}) (interface{}, *JSONRPCError) {
    card := AgentCard{
        Name:        "Gemma3 Agent",
        Description: "AI agent with browser automation, terminal control, and MCP integration",
        Version:     "1.0.0",
        Skills: []Skill{
            {Name: "web_browsing", Description: "Automated web browsing and data extraction"},
            {Name: "terminal_control", Description: "Execute terminal commands"},
            {Name: "mcp_integration", Description: "Access MCP tools and services"},
        },
        URL:       "wss://localhost:8080/ws",
        Transport: "JSONRPC",
    }
    
    return card, nil
}

func (s *A2AServer) handleMessageSend(params map[string]interface{}) (interface{}, *JSONRPCError) {
    // Parse message
    messageData, ok := params["message"].(map[string]interface{})
    if !ok {
        return nil, &JSONRPCError{
            Code:    -32602,
            Message: "Invalid params: message required",
        }
    }
    
    // Create new task
    taskId := uuid.New().String()
    task := &Task{
        ID: taskId,
        Status: TaskStatus{
            State:   "working",
            Message: "Processing your request...",
        },
        Messages:  []Message{},
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    s.tasksMux.Lock()
    s.tasks[taskId] = task
    s.tasksMux.Unlock()
    
    // Extract text from message parts
    parts, _ := messageData["parts"].([]interface{})
    var text string
    if len(parts) > 0 {
        part, _ := parts[0].(map[string]interface{})
        text, _ = part["text"].(string)
    }
    
    // Process message with agent controller
    go s.processTask(taskId, text)
    
    return map[string]interface{}{
        "taskId": taskId,
        "status": task.Status,
    }, nil
}

func (s *A2AServer) processTask(taskId string, text string) {
    // Execute task with Gemma 3 agent
    result := s.agentController.ExecuteTask(text)
    
    // Update task status
    s.tasksMux.Lock()
    if task, exists := s.tasks[taskId]; exists {
        task.Status = TaskStatus{
            State:   "completed",
            Message: result,
        }
        task.UpdatedAt = time.Now()
    }
    s.tasksMux.Unlock()
    
    // Broadcast task status update
    s.broadcastNotification("tasks/status", map[string]interface{}{
        "taskId": taskId,
        "status": TaskStatus{
            State:   "completed",
            Message: result,
        },
    })
}

func (s *A2AServer) handleTasksGet(params map[string]interface{}) (interface{}, *JSONRPCError) {
    taskId, ok := params["taskId"].(string)
    if !ok {
        return nil, &JSONRPCError{
            Code:    -32602,
            Message: "Invalid params: taskId required",
        }
    }
    
    s.tasksMux.RLock()
    task, exists := s.tasks[taskId]
    s.tasksMux.RUnlock()
    
    if !exists {
        return nil, &JSONRPCError{
            Code:    -32001,
            Message: "Task not found",
            Data:    taskId,
        }
    }
    
    return task, nil
}

func (s *A2AServer) handleTasksList(params map[string]interface{}) (interface{}, *JSONRPCError) {
    limit := 50
    if l, ok := params["limit"].(float64); ok {
        limit = int(l)
    }
    
    s.tasksMux.RLock()
    tasks := make([]*Task, 0, len(s.tasks))
    for _, task := range s.tasks {
        tasks = append(tasks, task)
        if len(tasks) >= limit {
            break
        }
    }
    s.tasksMux.RUnlock()
    
    return map[string]interface{}{
        "tasks": tasks,
        "total": len(s.tasks),
    }, nil
}

func (s *A2AServer) handleTasksCancel(params map[string]interface{}) (interface{}, *JSONRPCError) {
    taskId, ok := params["taskId"].(string)
    if !ok {
        return nil, &JSONRPCError{
            Code:    -32602,
            Message: "Invalid params: taskId required",
        }
    }
    
    s.tasksMux.Lock()
    if task, exists := s.tasks[taskId]; exists {
        task.Status = TaskStatus{
            State:   "cancelled",
            Message: "Task cancelled by user",
        }
        task.UpdatedAt = time.Now()
    }
    s.tasksMux.Unlock()
    
    return map[string]interface{}{
        "success": true,
    }, nil
}

// ===== Browser Control Handlers =====

func (s *A2AServer) handleBrowserNavigate(params map[string]interface{}) (interface{}, *JSONRPCError) {
    url, ok := params["url"].(string)
    if !ok {
        return nil, &JSONRPCError{
            Code:    -32602,
            Message: "Invalid params: url required",
        }
    }
    
    s.browserManager.Navigate(url)
    
    // Broadcast action
    s.broadcastNotification("browser/action", map[string]interface{}{
        "action": "navigate",
        "target": url,
    })
    
    return map[string]interface{}{"success": true}, nil
}

func (s *A2AServer) handleBrowserClick(params map[string]interface{}) (interface{}, *JSONRPCError) {
    element, ok := params["element"].(float64)
    if !ok {
        return nil, &JSONRPCError{
            Code:    -32602,
            Message: "Invalid params: element required",
        }
    }
    
    s.browserManager.Click(int(element))
    
    s.broadcastNotification("browser/action", map[string]interface{}{
        "action": "click",
        "target": int(element),
    })
    
    return map[string]interface{}{"success": true}, nil
}

func (s *A2AServer) handleBrowserType(params map[string]interface{}) (interface{}, *JSONRPCError) {
    element, ok := params["element"].(float64)
    if !ok {
        return nil, &JSONRPCError{
            Code:    -32602,
            Message: "Invalid params: element required",
        }
    }
    
    text, ok := params["text"].(string)
    if !ok {
        return nil, &JSONRPCError{
            Code:    -32602,
            Message: "Invalid params: text required",
        }
    }
    
    s.browserManager.Type(int(element), text)
    
    return map[string]interface{}{"success": true}, nil
}

func (s *A2AServer) handleBrowserScreenshot(params map[string]interface{}) (interface{}, *JSONRPCError) {
    screenshot := s.browserManager.CaptureScreenshot()
    
    return map[string]interface{}{
        "image": screenshot,
        "timestamp": time.Now().Unix(),
    }, nil
}

// ===== Terminal Control Handlers =====

func (s *A2AServer) handleTerminalExecute(params map[string]interface{}) (interface{}, *JSONRPCError) {
    command, ok := params["command"].(string)
    if !ok {
        return nil, &JSONRPCError{
            Code:    -32602,
            Message: "Invalid params: command required",
        }
    }
    
    output := s.terminalManager.Execute(command)
    
    // Broadcast output
    s.broadcastNotification("terminal/output", map[string]interface{}{
        "output":      output,
        "attribution": "ai",
    })
    
    return map[string]interface{}{
        "output": output,
    }, nil
}

// ===== MCP Control Handlers =====

func (s *A2AServer) handleMCPListTools(params map[string]interface{}) (interface{}, *JSONRPCError) {
    server, ok := params["server"].(string)
    if !ok {
        return nil, &JSONRPCError{
            Code:    -32602,
            Message: "Invalid params: server required",
        }
    }
    
    tools := s.mcpClient.ListTools(server)
    
    return map[string]interface{}{
        "tools": tools,
    }, nil
}

func (s *A2AServer) handleMCPCallTool(params map[string]interface{}) (interface{}, *JSONRPCError) {
    server, _ := params["server"].(string)
    tool, _ := params["tool"].(string)
    args := params["args"]
    
    result := s.mcpClient.CallTool(server, tool, args)
    
    s.broadcastNotification("mcp/toolResult", map[string]interface{}{
        "server":  server,
        "tool":    tool,
        "result":  result,
        "success": result != nil,
    })
    
    return map[string]interface{}{
        "result": result,
    }, nil
}

// ===== Broadcast Notification =====

func (s *A2AServer) broadcastNotification(method string, params map[string]interface{}) {
    notification := JSONRPCNotification{
        JSONRPC: "2.0",
        Method:  method,
        Params:  params,
    }
    
    s.clientsMux.RLock()
    defer s.clientsMux.RUnlock()
    
    for client := range s.clients {
        if err := client.WriteJSON(notification); err != nil {
            log.Println("Broadcast error:", err)
        }
    }
}

// ===== HTTP JSON-RPC Handler =====

func (s *A2AServer) handleHTTPJSONRPC(c *fiber.Ctx) error {
    var req JSONRPCRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      nil,
            Error: &JSONRPCError{
                Code:    -32700,
                Message: "Parse error",
            },
        })
    }
    
    response := s.handleJSONRPCRequest(req)
    return c.JSON(response)
}

// ===== Agent Card Handler =====

func (s *A2AServer) handleAgentCard(c *fiber.Ctx) error {
    card := AgentCard{
        Name:        "Gemma3 Agent",
        Description: "AI agent with browser automation, terminal control, and MCP integration",
        Version:     "1.0.0",
        Skills: []Skill{
            {Name: "web_browsing", Description: "Automated web browsing and data extraction"},
            {Name: "terminal_control", Description: "Execute terminal commands"},
            {Name: "mcp_integration", Description: "Access MCP tools and services"},
        },
        URL:       "https://localhost:8080/jsonrpc",
        Transport: "JSONRPC",
    }
    
    return c.JSON(card)
}

func (s *A2AServer) Start(port string) {
    log.Printf("Starting A2A Server on port %s", port)
    log.Fatal(s.app.Listen(":" + port))
}

func main() {
    server := NewA2AServer()
    server.Setup()
    server.Start("8080")
}
```

---

## Task Status Display (Glassmorphism)

```css
.task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
}

.task-item {
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(15px);
  border-left: 3px solid #15A7FF;
  border-radius: 8px;
  padding: 12px 16px;
  transition: all 0.3s ease;
  position: relative;
}

.task-item.working {
  border-left-color: #15A7FF;
  animation: taskPulse 2s ease-in-out infinite;
}

.task-item.completed {
  border-left-color: #00FF88;
}

.task-item.failed {
  border-left-color: #FF2A6D;
}

.task-item.cancelled {
  border-left-color: #8D9AA8;
  opacity: 0.6;
}

@keyframes taskPulse {
  0%, 100% {
    box-shadow: -4px 0 12px rgba(21, 167, 255, 0.3);
  }
  50% {
    box-shadow: -4px 0 20px rgba(21, 167, 255, 0.6);
  }
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.task-id {
  font-family: 'Monaco', 'Consolas', 'Courier New', monospace;
  font-size: 0.8rem;
  color: #8D9AA8;
}

.task-state {
  background: rgba(21, 167, 255, 0.3);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: var(--font-weight-medium);
  color: #FFFFFF;
  text-transform: uppercase;
}

.task-item.completed .task-state {
  background: rgba(0, 255, 136, 0.3);
}

.task-item.failed .task-state {
  background: rgba(255, 42, 109, 0.3);
}

.task-message {
  color: #FFFFFF;
  font-family: var(--font-primary);
  font-size: 0.9rem;
  line-height: 1.4;
}
```

---

## Implementation Checklist

### Frontend (A2A Client)
- [ ] JSON-RPC 2.0 WebSocket client
- [ ] Request/response promise handling
- [ ] Notification handler system
- [ ] Agent card display
- [ ] Task status tracking
- [ ] Browser takeover UI
- [ ] Terminal takeover UI
- [ ] MCP integration panel
- [ ] Connection status indicator

### Backend (A2A Server)
- [ ] JSON-RPC 2.0 request handler
- [ ] Method routing system
- [ ] A2A protocol methods (message/send, tasks/*, agent/*)
- [ ] Browser automation integration
- [ ] Terminal PTY management
- [ ] MCP client integration
- [ ] Task lifecycle management
- [ ] Notification broadcasting
- [ ] Agent card endpoint

### Agent Integration (Gemma 3)
- [ ] Task orchestration
- [ ] Action parsing
- [ ] Browser/terminal control
- [ ] MCP tool invocation
- [ ] Error handling

### Protocol Compliance
- [ ] JSON-RPC 2.0 format validation
- [ ] A2A method naming conventions
- [ ] Task state management
- [ ] Error code mapping
- [ ] Agent card structure

---

## Conclusion

This comprehensive design specification provides a **fully A2A-compliant architecture** using **JSON-RPC 2.0** for agent-to-agent communication, maintaining the midnight glassmorphism aesthetic while enabling standardized, interoperable agent interactions. The implementation supports browser/terminal takeover, MCP integration, and full Gemma 3-powered agent control through a robust, spec-compliant communication protocol.

