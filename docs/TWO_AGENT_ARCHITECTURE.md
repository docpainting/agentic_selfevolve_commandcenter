# Two-Agent Architecture: Main + Terminal

## Overview

The workspace uses a **two-agent architecture** designed for efficiency and specialization:

1. **Main Agent** (Gemma 3 27B) - Orchestrator with self-awareness, evolution, and memory
2. **Terminal Agent** (Specialized) - Natural language → Linux commands (250+ commands)

Later, the main agent can **train new agents from scratch** using its accumulated memory and knowledge.

---

## Why Two Agents?

### Main Agent (Gemma 3)
- **Role**: Orchestrator, planner, self-aware entity
- **Strengths**: Reasoning, planning, self-modification, learning
- **Tools**: Browser, file operations, MCP servers (OpenEvolve, etc.)
- **Memory**: Neo4j knowledge graph, LightRAG, short-term context

### Terminal Agent (Specialized)
- **Role**: Linux command execution specialist
- **Strengths**: Natural language → precise Linux commands
- **Training**: 250+ Linux commands, shell scripting, system administration
- **Purpose**: **Offload terminal work from main agent** (reduce cognitive load)

### Why Separate?

**Terminal operations are cognitively expensive for the main agent:**
- Requires precise command syntax
- Needs deep Linux knowledge
- Many edge cases and flags
- Distracts from higher-level reasoning

**Solution**: Dedicated terminal agent trained specifically on Linux commands.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Main Agent (Gemma 3)                    │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Core Capabilities:                                │    │
│  │  - Self-awareness (Neo4j code mirror)              │    │
│  │  - Self-modification (OpenEvolve)                  │    │
│  │  - Planning and reasoning                          │    │
│  │  - Memory (LightRAG + Neo4j)                       │    │
│  │  - Learning patterns                               │    │
│  └────────────────────────────────────────────────────┘    │
│                                                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Tools (via Unified Registry):                     │    │
│  │                                                     │    │
│  │  Built-in:                                         │    │
│  │  - browser_navigate                                │    │
│  │  - browser_click                                   │    │
│  │  - file_read                                       │    │
│  │  - file_write                                      │    │
│  │                                                     │    │
│  │  MCP:                                              │    │
│  │  - openevolve.evolve_code                          │    │
│  │  - openevolve.evaluate_code                        │    │
│  │  - playwright.navigate                             │    │
│  │                                                     │    │
│  │  Agent Delegation:                                 │    │
│  │  - terminal_agent.execute(natural_language)        │    │
│  │                                                     │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ A2A (JSON-RPC 2.0)
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Terminal Agent (Specialized)              │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Core Capability:                                  │    │
│  │  Natural Language → Linux Commands                 │    │
│  │                                                     │    │
│  │  Training:                                         │    │
│  │  - 250+ Linux commands                             │    │
│  │  - Shell scripting                                 │    │
│  │  - System administration                           │    │
│  │  - Package management                              │    │
│  │  - Process management                              │    │
│  │  - File operations                                 │    │
│  └────────────────────────────────────────────────────┘    │
│                                                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Tools:                                            │    │
│  │  - pty_manager (execute commands)                  │    │
│  │  - command_validator (check safety)                │    │
│  │  - output_parser (parse results)                   │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

---

## Communication Protocol (A2A)

### Main Agent → Terminal Agent

**Request Format**:
```json
{
  "jsonrpc": "2.0",
  "method": "terminal.execute",
  "params": {
    "task": "install nginx and start the service",
    "context": {
      "os": "ubuntu-22.04",
      "user": "ubuntu",
      "working_dir": "/home/ubuntu/agent-workspace"
    }
  },
  "id": "req-123"
}
```

**Response Format**:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "commands_executed": [
      "sudo apt-get update",
      "sudo apt-get install -y nginx",
      "sudo systemctl start nginx",
      "sudo systemctl enable nginx"
    ],
    "output": "nginx installed and started successfully",
    "exit_code": 0,
    "success": true
  },
  "id": "req-123"
}
```

### Terminal Agent → Main Agent (Clarification)

**Request Format**:
```json
{
  "jsonrpc": "2.0",
  "method": "agent.clarify",
  "params": {
    "question": "Should I install nginx with default config or custom config?",
    "options": ["default", "custom"],
    "context": "User requested nginx installation but didn't specify configuration"
  },
  "id": "req-456"
}
```

---

## How It Works

### Example 1: User Asks to Install Package

```
User: "Install and configure PostgreSQL"
    ↓
Main Agent (Gemma 3):
    Thinks: "This is a terminal task, delegate to terminal agent"
    ↓
    Calls tool: terminal_agent.execute({
        task: "install and configure PostgreSQL",
        context: {...}
    })
    ↓
Terminal Agent:
    Interprets: "install and configure PostgreSQL"
    ↓
    Plans commands:
    1. sudo apt-get update
    2. sudo apt-get install -y postgresql postgresql-contrib
    3. sudo systemctl start postgresql
    4. sudo systemctl enable postgresql
    5. sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'secure_password';"
    ↓
    Executes via PTY
    ↓
    Returns: {success: true, output: "..."}
    ↓
Main Agent:
    Receives result
    ↓
    Responds to user: "PostgreSQL installed and configured successfully!"
```

### Example 2: User Asks to Improve Code

```
User: "Improve this login function"
    ↓
Main Agent (Gemma 3):
    Thinks: "This is a code evolution task, use OpenEvolve"
    ↓
    Calls tool: openevolve.evolve_code({
        code: "...",
        iterations: 100
    })
    ↓
OpenEvolve MCP Server:
    Runs evolution
    ↓
    Returns: {best_code: "...", score: 88}
    ↓
Main Agent:
    Stores in Neo4j
    ↓
    Responds to user: "Code improved! Score: 88 (+61)"
```

### Example 3: Complex Task (Both Agents)

```
User: "Deploy the frontend to production"
    ↓
Main Agent (Gemma 3):
    Plans:
    1. Build frontend (terminal task)
    2. Run tests (terminal task)
    3. Deploy to server (terminal task)
    4. Update deployment record (main agent task)
    ↓
    Step 1: terminal_agent.execute("build frontend with npm")
    ↓
Terminal Agent:
    cd frontend && npm run build
    ↓
    Returns: {success: true}
    ↓
Main Agent:
    Step 2: terminal_agent.execute("run frontend tests")
    ↓
Terminal Agent:
    npm test
    ↓
    Returns: {success: true, coverage: 85%}
    ↓
Main Agent:
    Step 3: terminal_agent.execute("deploy to production server")
    ↓
Terminal Agent:
    rsync -avz dist/ user@server:/var/www/
    ssh user@server "systemctl restart nginx"
    ↓
    Returns: {success: true}
    ↓
Main Agent:
    Step 4: Store deployment in Neo4j
    ↓
    Responds to user: "Deployed successfully! Coverage: 85%"
```

---

## Tool Registration

### Main Agent Tools

```go
// backend/internal/agent/tools.go
func (a *AgentController) RegisterTools() {
    // Built-in tools
    a.toolRegistry.RegisterBuiltin("browser_navigate", BrowserNavigateTool{...})
    a.toolRegistry.RegisterBuiltin("file_read", FileReadTool{...})
    a.toolRegistry.RegisterBuiltin("file_write", FileWriteTool{...})
    
    // MCP tools (auto-discovered)
    a.mcpClient.ConnectAll()
    
    // Terminal agent delegation
    a.toolRegistry.RegisterBuiltin("terminal_agent.execute", TerminalAgentTool{
        Description: "Execute terminal commands via specialized terminal agent",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "task": map[string]string{
                    "type": "string",
                    "description": "Natural language description of terminal task",
                },
                "context": map[string]string{
                    "type": "object",
                    "description": "Additional context (os, user, working_dir)",
                },
            },
            "required": []string{"task"},
        },
        Execute: func(args map[string]interface{}) (interface{}, error) {
            // Send A2A request to terminal agent
            return a.a2aClient.Request("terminal_agent", "terminal.execute", args)
        },
    })
}
```

### Terminal Agent Tools

```go
// backend/cmd/terminal_agent/main.go
func (t *TerminalAgent) RegisterTools() {
    // Only has terminal-related tools
    t.toolRegistry.RegisterBuiltin("pty_execute", PTYExecuteTool{...})
    t.toolRegistry.RegisterBuiltin("validate_command", ValidateCommandTool{...})
    t.toolRegistry.RegisterBuiltin("parse_output", ParseOutputTool{...})
}
```

---

## A2A Implementation

### Main Agent Side

```go
// backend/internal/a2a/client.go
type A2AClient struct {
    agents map[string]*AgentConnection
}

func (c *A2AClient) Request(agentID, method string, params interface{}) (interface{}, error) {
    conn := c.agents[agentID]
    if conn == nil {
        return nil, fmt.Errorf("agent not found: %s", agentID)
    }
    
    request := JSONRPCRequest{
        JSONRPC: "2.0",
        Method:  method,
        Params:  params,
        ID:      generateID(),
    }
    
    // Send via WebSocket or HTTP
    response, err := conn.Send(request)
    if err != nil {
        return nil, err
    }
    
    return response.Result, nil
}
```

### Terminal Agent Side

```go
// backend/cmd/terminal_agent/server.go
type TerminalAgentServer struct {
    ptyManager      *PTYManager
    commandParser   *CommandParser
    llm             *LLM  // Specialized terminal LLM
}

func (s *TerminalAgentServer) HandleRequest(req JSONRPCRequest) JSONRPCResponse {
    switch req.Method {
    case "terminal.execute":
        return s.executeTerminalTask(req.Params)
    default:
        return JSONRPCResponse{
            JSONRPC: "2.0",
            Error: &JSONRPCError{
                Code:    -32601,
                Message: "Method not found",
            },
            ID: req.ID,
        }
    }
}

func (s *TerminalAgentServer) executeTerminalTask(params map[string]interface{}) JSONRPCResponse {
    task := params["task"].(string)
    context := params["context"].(map[string]interface{})
    
    // 1. Use specialized LLM to convert natural language to commands
    commands, err := s.llm.GenerateCommands(task, context)
    if err != nil {
        return errorResponse(err)
    }
    
    // 2. Validate commands for safety
    for _, cmd := range commands {
        if !s.commandParser.IsSafe(cmd) {
            return errorResponse(fmt.Errorf("unsafe command: %s", cmd))
        }
    }
    
    // 3. Execute commands via PTY
    results := []CommandResult{}
    for _, cmd := range commands {
        result, err := s.ptyManager.Execute(cmd)
        if err != nil {
            return errorResponse(err)
        }
        results = append(results, result)
    }
    
    // 4. Return results
    return JSONRPCResponse{
        JSONRPC: "2.0",
        Result: map[string]interface{}{
            "commands_executed": commands,
            "output":           aggregateOutput(results),
            "exit_code":        results[len(results)-1].ExitCode,
            "success":          true,
        },
        ID: req.ID,
    }
}
```

---

## Future: Training New Agents

### Main Agent Trains New Agent

```
Main Agent (after months of learning):
    ↓
    Has accumulated:
    - 10,000+ patterns in Neo4j
    - Successful code evolution history
    - Learned best practices
    - Security knowledge
    - Performance optimizations
    ↓
    User: "Train a new API agent"
    ↓
Main Agent:
    1. Extract relevant patterns from Neo4j
       (API design, authentication, rate limiting, etc.)
    
    2. Generate training data:
       - Successful API implementations
       - Common patterns and anti-patterns
       - Security best practices
    
    3. Create specialized LLM fine-tuning dataset
    
    4. Train new agent with focused knowledge
    
    5. New "API Agent" born with:
       - Dense API-specific knowledge
       - Patterns learned from main agent
       - Specialized for API tasks
    ↓
New API Agent:
    - Handles all API-related tasks
    - Communicates via A2A
    - Has its own tools
    - Reduces load on main agent
```

### Training Process

```go
// backend/internal/agent/trainer.go
type AgentTrainer struct {
    mainAgent *AgentController
    neo4j     *Neo4jClient
}

func (t *AgentTrainer) TrainNewAgent(spec AgentSpec) (*Agent, error) {
    // 1. Extract relevant patterns from Neo4j
    patterns := t.neo4j.QueryPatterns(spec.Domain)
    
    // 2. Generate training data
    trainingData := t.generateTrainingData(patterns, spec)
    
    // 3. Create specialized LLM config
    llmConfig := t.createLLMConfig(spec)
    
    // 4. Fine-tune or configure LLM
    llm := t.fineTuneLLM(llmConfig, trainingData)
    
    // 5. Create new agent
    newAgent := &Agent{
        ID:          spec.ID,
        Name:        spec.Name,
        LLM:         llm,
        Tools:       spec.Tools,
        Memory:      t.createMemory(spec),
        A2AClient:   t.createA2AClient(),
    }
    
    // 6. Register with main agent
    t.mainAgent.RegisterSubAgent(newAgent)
    
    return newAgent, nil
}
```

---

## Benefits of This Architecture

### 1. Cognitive Load Distribution
- **Main agent**: High-level reasoning, planning, self-awareness
- **Terminal agent**: Specialized Linux command execution
- **Future agents**: Domain-specific tasks

### 2. Specialization
- Each agent is **expert** in its domain
- Better performance than generalist
- Faster, more accurate responses

### 3. Scalability
- Add new agents as needed
- Each agent has its own tools
- Communicate via A2A
- Main agent orchestrates

### 4. Knowledge Transfer
- Main agent accumulates knowledge
- Can train new agents from scratch
- New agents inherit learned patterns
- Continuous improvement

### 5. Maintainability
- Clear separation of concerns
- Each agent is independent
- Easy to debug and update
- Modular architecture

---

## Configuration

### Agent Registry

```yaml
# config/agents.yaml
agents:
  main:
    id: main_agent
    llm:
      model: gemma3:27b
      api_base: http://localhost:11434/v1/
    tools:
      - browser_*
      - file_*
      - openevolve.*
      - terminal_agent.execute
    memory:
      neo4j: true
      lightrag: true
    
  terminal:
    id: terminal_agent
    llm:
      model: terminal_specialist  # Specialized model
      api_base: http://localhost:11434/v1/
    tools:
      - pty_execute
      - validate_command
      - parse_output
    a2a:
      listen_port: 8081
      methods:
        - terminal.execute
```

---

## Summary

### Current Architecture (Two Agents)

1. **Main Agent** (Gemma 3)
   - Orchestrator with self-awareness
   - Uses browser, files, MCP tools
   - Delegates terminal work to terminal agent
   - Stores knowledge in Neo4j

2. **Terminal Agent** (Specialized)
   - Natural language → Linux commands
   - 250+ commands trained
   - Reduces cognitive load on main agent
   - Communicates via A2A

### Future Architecture (Multi-Agent)

1. **Main Agent** - Orchestrator
2. **Terminal Agent** - Linux specialist
3. **API Agent** - API design/implementation (trained by main agent)
4. **Database Agent** - Database operations (trained by main agent)
5. **Security Agent** - Security analysis (trained by main agent)
6. **...more as needed**

### Key Points

- ✅ **Two agents now**: Main + Terminal
- ✅ **A2A for communication**: JSON-RPC 2.0
- ✅ **Each has own tools**: No overlap
- ✅ **Terminal agent offloads work**: Reduces main agent cognitive load
- ✅ **Future: Main agent trains new agents**: From accumulated knowledge
- ✅ **Clean, scalable architecture**: Easy to add new agents

This is a **much better architecture** than trying to do everything with one agent!

