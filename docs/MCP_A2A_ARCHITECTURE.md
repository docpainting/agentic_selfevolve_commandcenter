# MCP, A2A, and Tool Architecture - Simplified Guide

This document clarifies the relationship between MCP (Model Context Protocol), A2A (Agent-to-Agent), tools, and how they all work together in the simplest possible way.

## The Confusion Explained

You're dealing with **three different concepts** that can overlap:

1. **MCP** - Protocol for connecting AI to external tools/data
2. **A2A** - Protocol for agents to talk to each other
3. **Tool Calling** - How the LLM invokes functions

Let me break down each and show how they fit together.

---

## Part 1: What is MCP?

### MCP in Simple Terms

**MCP is like a USB port for AI tools.**

Just like USB lets you plug any device into your computer, MCP lets you plug any tool into your AI agent.

### MCP Components

```
┌─────────────────────────────────────────────────────┐
│                  Your Agent (Go)                    │
│  ┌──────────────────────────────────────────────┐  │
│  │  MCP Client (built into your agent)          │  │
│  │  - Discovers tools from MCP servers          │  │
│  │  - Calls tools via JSON-RPC 2.0              │  │
│  │  - Handles responses                         │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
                      │
                      │ JSON-RPC 2.0
                      ↓
┌─────────────────────────────────────────────────────┐
│              MCP Servers (Python/Node/etc)          │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐   │
│  │ OpenEvolve │  │ Playwright │  │  Custom    │   │
│  │  Server    │  │  Server    │  │  Server    │   │
│  │            │  │            │  │            │   │
│  │ Tools:     │  │ Tools:     │  │ Tools:     │   │
│  │ - evolve   │  │ - navigate │  │ - whatever │   │
│  │ - evaluate │  │ - click    │  │ - you want │   │
│  └────────────┘  └────────────┘  └────────────┘   │
└─────────────────────────────────────────────────────┘
```

### Key Point: MCP is NOT for Agent-to-Agent Communication

**MCP is for**: Agent → External Tools  
**A2A is for**: Agent → Other Agents

---

## Part 2: What is A2A?

### A2A in Simple Terms

**A2A is how your agents talk to each other.**

Think of it like agents having a group chat where they can ask each other for help.

### A2A Protocol

```
┌──────────────┐                    ┌──────────────┐
│   Agent 1    │  ←─ JSON-RPC ─→   │   Agent 2    │
│  (Browser)   │                    │  (Terminal)  │
└──────────────┘                    └──────────────┘
        ↓                                  ↓
        └────────────→ Hub ←───────────────┘
                  (Coordinator)
```

### A2A Message Format

```json
{
  "jsonrpc": "2.0",
  "method": "agent.request",
  "params": {
    "from": "browser_agent",
    "to": "terminal_agent",
    "task": "run command 'npm install'",
    "context": {...}
  },
  "id": "123"
}
```

### Key Point: A2A is a Custom Protocol

**You define** what agents can ask each other. It's not MCP, it's your own protocol (JSON-RPC 2.0 format).

---

## Part 3: What are Tools?

### Tools in Simple Terms

**Tools are functions the LLM can call.**

When you give Gemma 3 a list of tools, it can decide to call them based on the conversation.

### How Tool Calling Works

```
User: "What's the weather in NYC?"
    ↓
LLM thinks: "I need to call get_weather tool"
    ↓
LLM returns: {
  "tool_call": {
    "name": "get_weather",
    "arguments": {"city": "NYC"}
  }
}
    ↓
Your code: Executes get_weather("NYC")
    ↓
Returns: "72°F, sunny"
    ↓
LLM: "The weather in NYC is 72°F and sunny!"
```

### Key Point: LLM Doesn't Use Natural Language to Call Tools

**No natural language needed!** The LLM returns a **structured JSON** with the tool name and arguments.

---

## Part 4: How They All Fit Together

### The Complete Picture

```
┌─────────────────────────────────────────────────────────────┐
│                    Your Agent Workspace                     │
│                                                             │
│  ┌────────────────────────────────────────────────────┐   │
│  │  Agent Controller (Go)                             │   │
│  │  ┌──────────────────────────────────────────────┐ │   │
│  │  │  Gemma 3 LLM                                 │ │   │
│  │  │  - Receives tool definitions                 │ │   │
│  │  │  - Returns tool calls (JSON)                 │ │   │
│  │  └──────────────────────────────────────────────┘ │   │
│  │                                                    │   │
│  │  ┌──────────────────────────────────────────────┐ │   │
│  │  │  Tool Registry                               │ │   │
│  │  │  - Built-in tools (browser, terminal)        │ │   │
│  │  │  - MCP tools (from MCP servers)              │ │   │
│  │  │  - All exposed to LLM as one list            │ │   │
│  │  └──────────────────────────────────────────────┘ │   │
│  │                                                    │   │
│  │  ┌──────────────────────────────────────────────┐ │   │
│  │  │  Tool Executor                               │ │   │
│  │  │  - Receives tool call from LLM               │ │   │
│  │  │  - Routes to correct handler:                │ │   │
│  │  │    * Built-in → Direct execution             │ │   │
│  │  │    * MCP → Call MCP server                   │ │   │
│  │  │    * A2A → Send to other agent               │ │   │
│  │  └──────────────────────────────────────────────┘ │   │
│  └────────────────────────────────────────────────────┘   │
│                                                             │
│  MCP Client ──→ MCP Servers (OpenEvolve, Playwright, etc)  │
│  A2A Client ──→ Other Agents (Browser, Terminal, etc)      │
└─────────────────────────────────────────────────────────────┘
```

---

## Part 5: Answering Your Questions

### Q1: Are tools separate from MCP?

**Answer**: Tools can come from **multiple sources**:

1. **Built-in tools** (you write in Go)
   - `browser_navigate(url)`
   - `terminal_execute(command)`
   - `file_read(path)`

2. **MCP tools** (from MCP servers)
   - `openevolve.evolve_code(code)`
   - `playwright.click(selector)`

3. **A2A tools** (delegate to other agents)
   - `ask_browser_agent(task)`
   - `ask_terminal_agent(task)`

**They all get merged into ONE tool list** that you give to the LLM.

### Q2: Can all agents use one MCP?

**Answer**: YES! **One MCP client can be shared by all agents.**

```
┌──────────────┐
│   Agent 1    │──┐
└──────────────┘  │
                  │
┌──────────────┐  │    ┌─────────────────┐
│   Agent 2    │──┼───→│  Shared MCP     │──→ MCP Servers
└──────────────┘  │    │  Client         │
                  │    └─────────────────┘
┌──────────────┐  │
│   Agent 3    │──┘
└──────────────┘
```

**OR** each agent can have its own MCP client (doesn't matter, both work).

### Q3: Does the model use natural language to call tools?

**Answer**: NO! The LLM returns **structured JSON**.

**Example**:

```json
// LLM returns this (not natural language):
{
  "tool_calls": [
    {
      "id": "call_123",
      "type": "function",
      "function": {
        "name": "evolve_code",
        "arguments": "{\"code\": \"func Add(a, b int) int { return a + b }\", \"iterations\": 100}"
      }
    }
  ]
}
```

Your code then:
1. Parses the JSON
2. Calls the function
3. Returns result to LLM

---

## Part 6: The Simplest Possible Architecture

### Recommended: Single Agent with Unified Tool Registry

```
┌─────────────────────────────────────────────────────────┐
│              Main Agent (Go)                            │
│                                                         │
│  ┌────────────────────────────────────────────────┐   │
│  │  Gemma 3 LLM                                   │   │
│  │  Gets ONE list of ALL tools                    │   │
│  └────────────────────────────────────────────────┘   │
│                        ↓                               │
│  ┌────────────────────────────────────────────────┐   │
│  │  Unified Tool Registry                         │   │
│  │                                                │   │
│  │  Built-in:                                     │   │
│  │  - browser_navigate                            │   │
│  │  - terminal_execute                            │   │
│  │  - file_read                                   │   │
│  │                                                │   │
│  │  MCP (auto-discovered):                        │   │
│  │  - openevolve.evolve_code                      │   │
│  │  - openevolve.evaluate_code                    │   │
│  │  - playwright.navigate                         │   │
│  │  - playwright.click                            │   │
│  │                                                │   │
│  │  (All exposed to LLM as one flat list)        │   │
│  └────────────────────────────────────────────────┘   │
│                        ↓                               │
│  ┌────────────────────────────────────────────────┐   │
│  │  Tool Executor                                 │   │
│  │  - If built-in: Execute directly               │   │
│  │  - If MCP: Call MCP server                     │   │
│  └────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

### Implementation (Go)

```go
// backend/internal/agent/tools.go
type ToolRegistry struct {
    builtinTools map[string]BuiltinTool
    mcpClient    *MCPClient
}

// Register built-in tools
func (r *ToolRegistry) RegisterBuiltin(name string, fn BuiltinTool) {
    r.builtinTools[name] = fn
}

// Get ALL tools (built-in + MCP) as one list for LLM
func (r *ToolRegistry) GetAllTools() []Tool {
    tools := []Tool{}
    
    // 1. Add built-in tools
    for name, fn := range r.builtinTools {
        tools = append(tools, Tool{
            Name:        name,
            Description: fn.Description,
            Parameters:  fn.Parameters,
            Source:      "builtin",
        })
    }
    
    // 2. Auto-discover and add MCP tools
    mcpServers := r.mcpClient.ListServers()
    for _, server := range mcpServers {
        mcpTools := r.mcpClient.ListTools(server)
        for _, tool := range mcpTools {
            tools = append(tools, Tool{
                Name:        fmt.Sprintf("%s.%s", server, tool.Name),
                Description: tool.Description,
                Parameters:  tool.InputSchema,
                Source:      "mcp",
                MCPServer:   server,
            })
        }
    }
    
    return tools
}

// Execute any tool (built-in or MCP)
func (r *ToolRegistry) Execute(toolCall ToolCall) (interface{}, error) {
    // Check if built-in
    if fn, ok := r.builtinTools[toolCall.Name]; ok {
        return fn.Execute(toolCall.Arguments)
    }
    
    // Check if MCP tool (format: "server.tool")
    parts := strings.Split(toolCall.Name, ".")
    if len(parts) == 2 {
        server := parts[0]
        tool := parts[1]
        return r.mcpClient.CallTool(server, tool, toolCall.Arguments)
    }
    
    return nil, fmt.Errorf("unknown tool: %s", toolCall.Name)
}
```

### Usage in Agent Controller

```go
// backend/internal/agent/controller.go
func (a *AgentController) HandleUserMessage(msg string) error {
    // 1. Get all tools (built-in + MCP)
    tools := a.toolRegistry.GetAllTools()
    
    // 2. Send to LLM with tools
    response, err := a.llm.Generate(LLMRequest{
        Messages: a.conversationHistory,
        Tools:    tools,  // ONE list with everything
    })
    if err != nil {
        return err
    }
    
    // 3. Check if LLM wants to call a tool
    if response.ToolCalls != nil {
        for _, toolCall := range response.ToolCalls {
            // Execute (automatically routes to built-in or MCP)
            result, err := a.toolRegistry.Execute(toolCall)
            if err != nil {
                return err
            }
            
            // Add result to conversation
            a.conversationHistory = append(a.conversationHistory, Message{
                Role:       "tool",
                Content:    result,
                ToolCallID: toolCall.ID,
            })
        }
        
        // Continue conversation with tool results
        return a.HandleUserMessage("")
    }
    
    // 4. Return response to user
    return a.sendToUser(response.Content)
}
```

---

## Part 7: Do You Need A2A?

### Short Answer: Maybe Not!

If you have **one main agent** that can use all tools (built-in + MCP), you might not need A2A at all.

### When You DO Need A2A:

1. **Multiple independent agents** running in parallel
2. **Specialized agents** that need to coordinate
3. **Distributed system** with agents on different machines

### When You DON'T Need A2A:

1. **Single agent** with access to all tools
2. **All functionality** available via built-in + MCP tools
3. **Simple architecture** is preferred

### Recommendation: Start Without A2A

```
Phase 1: Single agent + built-in tools + MCP
    ↓
    Works great? → Done!
    ↓
    Need multiple agents? → Add A2A later
```

---

## Part 8: Simplified Architecture Recommendation

### What You Actually Need

```
┌─────────────────────────────────────────────────────────┐
│                  Main Agent (Go)                        │
│                                                         │
│  User Message                                           │
│       ↓                                                 │
│  Gemma 3 LLM (with tool definitions)                    │
│       ↓                                                 │
│  Tool Call? (JSON)                                      │
│       ↓                                                 │
│  Tool Router:                                           │
│  - browser_* → Browser Manager (built-in)               │
│  - terminal_* → Terminal Manager (built-in)             │
│  - file_* → File Manager (built-in)                     │
│  - openevolve.* → OpenEvolve MCP Server                 │
│  - playwright.* → Playwright MCP Server                 │
│       ↓                                                 │
│  Return result to LLM                                   │
│       ↓                                                 │
│  Continue conversation                                  │
└─────────────────────────────────────────────────────────┘
```

### No A2A Needed!

Everything is just **tools** that the main agent can call. Some are built-in, some are MCP. The LLM doesn't care - it just calls tools.

---

## Part 9: Configuration Example

### MCP Server Config

```json
{
  "mcpServers": {
    "openevolve": {
      "command": "python",
      "args": ["-m", "backend.mcp_servers.openevolve.server"]
    },
    "playwright": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-playwright"]
    }
  }
}
```

### Tool Registration (Go)

```go
func (a *AgentController) Initialize() error {
    // 1. Register built-in tools
    a.toolRegistry.RegisterBuiltin("browser_navigate", BrowserNavigateTool{
        Description: "Navigate browser to URL",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "url": map[string]string{"type": "string"},
            },
            "required": []string{"url"},
        },
        Execute: func(args map[string]interface{}) (interface{}, error) {
            return a.browserManager.Navigate(args["url"].(string))
        },
    })
    
    a.toolRegistry.RegisterBuiltin("terminal_execute", TerminalExecuteTool{
        Description: "Execute terminal command",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "command": map[string]string{"type": "string"},
            },
            "required": []string{"command"},
        },
        Execute: func(args map[string]interface{}) (interface{}, error) {
            return a.terminalManager.Execute(args["command"].(string))
        },
    })
    
    // 2. Initialize MCP client (auto-discovers tools from servers)
    a.mcpClient = NewMCPClient("config/mcp_servers.json")
    a.mcpClient.ConnectAll()
    
    // 3. Tool registry now has EVERYTHING
    // LLM gets one list: [browser_navigate, terminal_execute, openevolve.evolve_code, playwright.click, ...]
    
    return nil
}
```

---

## Part 10: Summary - Keep It Simple!

### What You Need:

1. ✅ **One main agent** (Go)
2. ✅ **Built-in tools** (browser, terminal, file)
3. ✅ **MCP client** (connects to MCP servers)
4. ✅ **Tool registry** (merges built-in + MCP into one list)
5. ✅ **Gemma 3 LLM** (gets all tools, returns JSON tool calls)

### What You DON'T Need (Yet):

1. ❌ **A2A protocol** (not needed for single agent)
2. ❌ **Separate tool registration** (MCP auto-discovers)
3. ❌ **Natural language tool calling** (LLM uses JSON)
4. ❌ **Multiple agents** (start with one)

### The Flow:

```
User: "Improve this code"
  ↓
LLM: Returns tool call JSON: {"name": "openevolve.evolve_code", "args": {...}}
  ↓
Tool Router: Sees "openevolve.*" → Routes to MCP client
  ↓
MCP Client: Calls OpenEvolve MCP server
  ↓
OpenEvolve: Runs evolution, returns result
  ↓
LLM: "I've improved your code! Here's what changed..."
```

**No confusion, no complexity, just clean tool calling!**

---

## Next Steps

I'll create a simple implementation that shows exactly how to:

1. Register built-in tools
2. Auto-discover MCP tools
3. Merge them into one list
4. Let Gemma 3 call any tool seamlessly

Want me to create the actual Go code for this simplified architecture?

