# OpenEvolve MCP Integration

## Overview

Instead of using a subprocess bridge between Go and Python, we can create a **dedicated MCP server** for OpenEvolve. This is cleaner, more maintainable, and aligns perfectly with the agent workspace architecture.

## Why MCP is Better Than Subprocess Bridge

### Subprocess Approach (Original Plan)
```
Go Backend
  ↓ exec.Command()
Python subprocess
  ↓ stdout/stderr parsing
Go Backend (parse results)
```

**Problems**:
- Complex error handling
- Difficult debugging
- No streaming updates
- Process management overhead
- Parsing stdout/stderr is fragile

### MCP Approach (Better!)
```
Go Backend (MCP Client)
  ↓ JSON-RPC 2.0
OpenEvolve MCP Server (Python)
  ↓ Structured responses
Go Backend (typed responses)
```

**Benefits**:
- ✅ Clean JSON-RPC 2.0 protocol
- ✅ Typed request/response
- ✅ Streaming updates via SSE
- ✅ Better error handling
- ✅ Easy debugging
- ✅ Reusable by other clients
- ✅ Already have MCP client in your architecture!

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Go Backend (Fiber v3)                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Agent      │  │  MCP Client  │  │  Watchdog    │ │
│  │  Controller  │──│  (existing)  │  │              │ │
│  └──────────────┘  └──────┬───────┘  └──────────────┘ │
└─────────────────────────────┼──────────────────────────┘
                              │ JSON-RPC 2.0
                              ↓
┌─────────────────────────────────────────────────────────┐
│          OpenEvolve MCP Server (Python)                 │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Tools:                                          │  │
│  │  - evolve_code(code, config, iterations)         │  │
│  │  - evaluate_code(code, metrics)                  │  │
│  │  - get_evolution_status(session_id)              │  │
│  │  - stop_evolution(session_id)                    │  │
│  │  - get_best_solution(session_id)                 │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Resources:                                      │  │
│  │  - evolution://sessions/{id}                     │  │
│  │  - evolution://history/{id}                      │  │
│  │  - evolution://patterns                          │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  OpenEvolve Library                              │  │
│  │  - MAP-Elites                                    │  │
│  │  - Island Model                                  │  │
│  │  - LLM Ensemble (Gemma 3)                        │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## MCP Server Implementation

### File Structure

```
backend/mcp_servers/openevolve/
├── server.py              # Main MCP server
├── tools.py               # MCP tool implementations
├── resources.py           # MCP resource providers
├── evaluator.py           # Custom evaluator generator
├── session_manager.py     # Evolution session management
├── config.py              # Load YAML configs
└── requirements.txt       # Python dependencies
```

### MCP Tools

#### 1. `evolve_code`

Starts an evolution session and returns session ID.

**Request**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "evolve_code",
    "arguments": {
      "code": "func LoginHandler(w http.ResponseWriter, r *http.Request) { ... }",
      "language": "go",
      "config_path": "config/openevolve/agent_config.yaml",
      "iterations": 100,
      "metrics": {
        "execution": 0.4,
        "patterns": 0.3,
        "security": 0.2,
        "quality": 0.1
      }
    }
  }
}
```

**Response**:
```json
{
  "content": [
    {
      "type": "text",
      "text": {
        "session_id": "evo-123456",
        "status": "started",
        "initial_score": 12.5,
        "target_iterations": 100
      }
    }
  ]
}
```

#### 2. `evaluate_code`

Evaluates code without evolution (just scoring).

**Request**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "evaluate_code",
    "arguments": {
      "code": "...",
      "language": "go",
      "metrics": { ... }
    }
  }
}
```

**Response**:
```json
{
  "content": [
    {
      "type": "text",
      "text": {
        "combined_score": 73.5,
        "breakdown": {
          "execution": 20,
          "patterns": 26,
          "security": 15,
          "quality": 12.5
        },
        "patterns_detected": [
          {"name": "password_hashing", "score": 10},
          {"name": "jwt_tokens", "score": 8},
          {"name": "rate_limiting", "score": 8}
        ]
      }
    }
  ]
}
```

#### 3. `get_evolution_status`

Gets current status of an evolution session (streaming updates).

**Request**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_evolution_status",
    "arguments": {
      "session_id": "evo-123456"
    }
  }
}
```

**Response**:
```json
{
  "content": [
    {
      "type": "text",
      "text": {
        "session_id": "evo-123456",
        "status": "running",
        "current_generation": 47,
        "total_generations": 100,
        "best_score": 73.5,
        "improvement": 61.0,
        "recent_improvements": [
          {"generation": 12, "improvement": "SQL injection fixed", "score_delta": 15},
          {"generation": 23, "improvement": "Password hashing added", "score_delta": 10},
          {"generation": 35, "improvement": "JWT tokens added", "score_delta": 8},
          {"generation": 47, "improvement": "Rate limiting added", "score_delta": 8}
        ]
      }
    }
  ]
}
```

#### 4. `stop_evolution`

Stops an evolution session early.

#### 5. `get_best_solution`

Gets the best solution from a completed evolution session.

**Response**:
```json
{
  "content": [
    {
      "type": "text",
      "text": {
        "session_id": "evo-123456",
        "best_code": "func LoginHandler(w http.ResponseWriter, r *http.Request) { ... }",
        "best_score": 88.5,
        "generation": 92,
        "total_generations": 100,
        "patterns_learned": [
          "password_hashing",
          "jwt_tokens",
          "rate_limiting",
          "session_management"
        ]
      }
    }
  ]
}
```

### MCP Resources

#### 1. `evolution://sessions/{id}`

Get full evolution session data.

#### 2. `evolution://history/{id}`

Get generation-by-generation history.

#### 3. `evolution://patterns`

Get all detected patterns across all sessions.

## Implementation

### 1. MCP Server (Python)

```python
# backend/mcp_servers/openevolve/server.py
from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import Tool, Resource, TextContent
import asyncio
from openevolve import run_evolution
from .session_manager import SessionManager
from .evaluator import EvaluatorGenerator
from .config import load_config

app = Server("openevolve")
sessions = SessionManager()
evaluator_gen = EvaluatorGenerator()

@app.list_tools()
async def list_tools() -> list[Tool]:
    return [
        Tool(
            name="evolve_code",
            description="Start evolution session for code improvement",
            inputSchema={
                "type": "object",
                "properties": {
                    "code": {"type": "string", "description": "Code to evolve"},
                    "language": {"type": "string", "description": "Programming language"},
                    "config_path": {"type": "string", "description": "Path to config YAML"},
                    "iterations": {"type": "integer", "description": "Number of iterations"},
                    "metrics": {"type": "object", "description": "Metric weights"}
                },
                "required": ["code", "language", "iterations"]
            }
        ),
        Tool(
            name="evaluate_code",
            description="Evaluate code without evolution",
            inputSchema={
                "type": "object",
                "properties": {
                    "code": {"type": "string"},
                    "language": {"type": "string"},
                    "metrics": {"type": "object"}
                },
                "required": ["code", "language"]
            }
        ),
        Tool(
            name="get_evolution_status",
            description="Get current evolution status",
            inputSchema={
                "type": "object",
                "properties": {
                    "session_id": {"type": "string"}
                },
                "required": ["session_id"]
            }
        ),
        Tool(
            name="stop_evolution",
            description="Stop evolution session",
            inputSchema={
                "type": "object",
                "properties": {
                    "session_id": {"type": "string"}
                },
                "required": ["session_id"]
            }
        ),
        Tool(
            name="get_best_solution",
            description="Get best solution from evolution",
            inputSchema={
                "type": "object",
                "properties": {
                    "session_id": {"type": "string"}
                },
                "required": ["session_id"]
            }
        )
    ]

@app.call_tool()
async def call_tool(name: str, arguments: dict) -> list[TextContent]:
    if name == "evolve_code":
        # Start evolution in background
        session_id = await sessions.start_evolution(
            code=arguments["code"],
            language=arguments["language"],
            config_path=arguments.get("config_path"),
            iterations=arguments["iterations"],
            metrics=arguments.get("metrics", {})
        )
        
        return [TextContent(
            type="text",
            text=json.dumps({
                "session_id": session_id,
                "status": "started"
            })
        )]
    
    elif name == "evaluate_code":
        # Generate evaluator
        evaluator = evaluator_gen.generate(
            language=arguments["language"],
            metrics=arguments.get("metrics", {})
        )
        
        # Evaluate
        score = evaluator.evaluate(arguments["code"])
        
        return [TextContent(
            type="text",
            text=json.dumps(score)
        )]
    
    elif name == "get_evolution_status":
        status = sessions.get_status(arguments["session_id"])
        return [TextContent(
            type="text",
            text=json.dumps(status)
        )]
    
    elif name == "stop_evolution":
        sessions.stop(arguments["session_id"])
        return [TextContent(
            type="text",
            text=json.dumps({"status": "stopped"})
        )]
    
    elif name == "get_best_solution":
        solution = sessions.get_best(arguments["session_id"])
        return [TextContent(
            type="text",
            text=json.dumps(solution)
        )]

async def main():
    async with stdio_server() as (read_stream, write_stream):
        await app.run(read_stream, write_stream, app.create_initialization_options())

if __name__ == "__main__":
    asyncio.run(main())
```

### 2. Go MCP Client Usage

```go
// backend/internal/agent/controller.go
func (a *AgentController) HandleTask(task *Task) error {
    // 1. Generate initial code
    initialCode := a.generateCode(task)
    
    // 2. Evaluate via MCP
    evalResult, err := a.mcpClient.CallTool("openevolve", "evaluate_code", map[string]interface{}{
        "code":     initialCode,
        "language": "go",
        "metrics": map[string]float64{
            "execution": 0.4,
            "patterns":  0.3,
            "security":  0.2,
            "quality":   0.1,
        },
    })
    if err != nil {
        return err
    }
    
    score := parseScore(evalResult)
    
    // 3. If score low, trigger evolution
    if score.Combined < 50 {
        log.Info("Score too low, triggering evolution", "score", score.Combined)
        
        evolveResult, err := a.mcpClient.CallTool("openevolve", "evolve_code", map[string]interface{}{
            "code":        initialCode,
            "language":    "go",
            "config_path": "config/openevolve/agent_config.yaml",
            "iterations":  100,
            "metrics": map[string]float64{
                "execution": 0.4,
                "patterns":  0.3,
                "security":  0.2,
                "quality":   0.1,
            },
        })
        if err != nil {
            return err
        }
        
        sessionID := parseSessionID(evolveResult)
        
        // 4. Poll for status (or use WebSocket for real-time updates)
        for {
            status, err := a.mcpClient.CallTool("openevolve", "get_evolution_status", map[string]interface{}{
                "session_id": sessionID,
            })
            if err != nil {
                return err
            }
            
            statusData := parseStatus(status)
            
            // Broadcast to UI
            a.broadcastEvolutionUpdate(statusData)
            
            if statusData.Status == "completed" {
                break
            }
            
            time.Sleep(2 * time.Second)
        }
        
        // 5. Get best solution
        bestResult, err := a.mcpClient.CallTool("openevolve", "get_best_solution", map[string]interface{}{
            "session_id": sessionID,
        })
        if err != nil {
            return err
        }
        
        initialCode = parseBestCode(bestResult)
    }
    
    // 6. Execute code
    return a.executor.Execute(initialCode)
}
```

## Configuration

### MCP Server Config

Add to your MCP config (already have this structure):

```json
{
  "mcpServers": {
    "openevolve": {
      "command": "python",
      "args": ["-m", "backend.mcp_servers.openevolve.server"],
      "env": {
        "OPENAI_API_KEY": "ollama",
        "OPENAI_API_BASE": "http://localhost:11434/v1/"
      }
    }
  }
}
```

## Advantages Over Subprocess Bridge

### 1. Clean Protocol
- JSON-RPC 2.0 (standard)
- Typed requests/responses
- Built-in error handling

### 2. Streaming Updates
- Real-time generation updates
- No polling needed (can use SSE)
- Better UX

### 3. Reusability
- Other clients can use the same MCP server
- CLI tools can connect
- External services can integrate

### 4. Debugging
- Easy to test with MCP inspector
- Clear request/response logs
- Standard debugging tools

### 5. Maintainability
- Separation of concerns
- Python code stays in Python
- Go code stays in Go
- Clean interface between them

### 6. Already Have MCP Client!
- You already have MCP client in your architecture
- Just add another server
- No new infrastructure needed

## Migration Path

### Current Plan (Subprocess)
```go
evolver.Evolve(code) → exec.Command("python", ...) → parse stdout
```

### New Plan (MCP)
```go
mcpClient.CallTool("openevolve", "evolve_code", args) → JSON response
```

**Migration**: Just implement the MCP server, no changes to Go architecture needed (already have MCP client)!

## Implementation Checklist

### Phase 1: MCP Server
- [ ] Create `backend/mcp_servers/openevolve/` directory
- [ ] Implement `server.py` with MCP tools
- [ ] Implement `session_manager.py` for evolution sessions
- [ ] Implement `evaluator.py` for custom evaluators
- [ ] Test with MCP inspector

### Phase 2: Integration
- [ ] Add openevolve to MCP server config
- [ ] Update Go MCP client to call openevolve tools
- [ ] Test evaluate_code tool
- [ ] Test evolve_code tool
- [ ] Test status polling

### Phase 3: UI
- [ ] WebSocket updates for evolution progress
- [ ] Real-time generation display
- [ ] Pattern match visualization

## Conclusion

**Using MCP is significantly better** than subprocess bridge because:

1. ✅ You already have MCP client infrastructure
2. ✅ Clean, standard protocol (JSON-RPC 2.0)
3. ✅ Better error handling and debugging
4. ✅ Streaming updates for real-time UI
5. ✅ Reusable by other clients
6. ✅ Easier to maintain and test

**Recommendation**: Go with MCP approach, skip subprocess bridge entirely!

