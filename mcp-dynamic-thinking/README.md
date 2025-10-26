# MCP Dynamic Thinking Server

Model Context Protocol (MCP) server implementing the **Perceive-Reason-Act-Reflect (PRAR)** loop for dynamic sequential thinking with branching reasoning capabilities.

## Overview

This MCP server provides advanced reasoning capabilities through a structured PRAR loop:

1. **Perceive** - Capture and analyze environment state
2. **Reason** - Generate and evaluate multiple reasoning branches
3. **Act** - Execute selected action plan with monitoring
4. **Reflect** - Analyze results and evolve strategies

## Features

- ✅ **Multi-branch Reasoning** - Generate and evaluate multiple reasoning paths
- ✅ **Short-term Memory** - Task-specific memory management
- ✅ **Long-term Learning** - Strategy evolution and pattern recognition
- ✅ **Execution Monitoring** - Real-time action tracking
- ✅ **Reflection System** - Automated learning from results
- ✅ **Strategy Evolution** - Dynamic confidence adjustment
- ✅ **Training Data Export** - Complete execution traces

## MCP Tools

### 1. `perceive`
Capture and analyze environment state.

**Input:**
```json
{
  "task_id": "task_123",
  "goal": "Find repository on GitHub",
  "screenshot": "<base64>",
  "elements": [...]
}
```

**Output:**
```json
{
  "perception_id": "perception_456",
  "analysis": {
    "interactive_elements": 15,
    "has_inputs": true,
    "ready_for_action": true
  }
}
```

### 2. `reason`
Generate and evaluate reasoning branches.

**Input:**
```json
{
  "task_id": "task_123",
  "perception_id": "perception_456",
  "num_branches": 3
}
```

**Output:**
```json
{
  "branches": [
    {
      "id": "branch_1",
      "strategy": "direct_approach",
      "confidence": 0.8,
      "selected": true
    }
  ],
  "action_plan": {
    "strategy": "direct_approach",
    "steps": [...]
  }
}
```

### 3. `act`
Execute action plan with monitoring.

**Input:**
```json
{
  "task_id": "task_123",
  "action_plan": {
    "strategy": "direct_approach",
    "steps": ["Navigate", "Click", "Verify"]
  }
}
```

**Output:**
```json
{
  "execution_id": "execution_789",
  "status": "completed",
  "steps": [
    {
      "id": "step_1",
      "status": "completed",
      "duration_ms": 150
    }
  ]
}
```

### 4. `reflect`
Analyze results and evolve strategies.

**Input:**
```json
{
  "task_id": "task_123",
  "execution_id": "execution_789"
}
```

**Output:**
```json
{
  "critique": "Execution completed successfully...",
  "lessons": ["Direct approach works well..."],
  "improvements": ["Add retry logic..."],
  "strategy_evolution": {
    "confidence_delta": 0.05,
    "new_confidence": 0.85
  }
}
```

### 5. `get_short_term_memory`
Retrieve task context.

**Input:**
```json
{
  "task_id": "task_123"
}
```

**Output:**
```json
{
  "perceptions": [...],
  "reasoning": [...],
  "actions": [...],
  "reflections": [...]
}
```

### 6. `clear_short_term_memory`
Clear task memory after completion.

**Input:**
```json
{
  "task_id": "task_123",
  "archive_to_long_term": true
}
```

### 7. `query_strategies`
Find relevant strategies from knowledge graph.

**Input:**
```json
{
  "query": "strategies for web navigation"
}
```

### 8. `get_execution_trace`
Export complete execution trace for training.

**Input:**
```json
{
  "task_id": "task_123"
}
```

## Installation

```bash
cd mcp-dynamic-thinking
go mod download
go build -o mcp-dynamic-thinking cmd/server/main.go
```

## Usage

### Standalone
```bash
./mcp-dynamic-thinking
```

### With MCP Client
```javascript
const client = new MCPClient();
await client.connect('stdio', './mcp-dynamic-thinking');

// Perceive
const perception = await client.callTool('perceive', {
  task_id: 'task_123',
  goal: 'Find repository'
});

// Reason
const reasoning = await client.callTool('reason', {
  task_id: 'task_123',
  perception_id: perception.perception_id,
  num_branches: 3
});

// Act
const execution = await client.callTool('act', {
  task_id: 'task_123',
  action_plan: reasoning.action_plan
});

// Reflect
const reflection = await client.callTool('reflect', {
  task_id: 'task_123',
  execution_id: execution.execution_id
});
```

## Architecture

```
mcp-dynamic-thinking/
├── cmd/
│   └── server/
│       └── main.go          # MCP server entry point
├── internal/
│   ├── perceive/
│   │   └── perceive.go      # Perception phase
│   ├── reason/
│   │   └── reason.go        # Reasoning phase
│   ├── act/
│   │   └── act.go           # Action phase
│   ├── reflect/
│   │   └── reflect.go       # Reflection phase
│   └── memory/
│       └── memory.go        # Memory management
└── go.mod
```

## Integration with Main Backend

The main backend can call this MCP server through the MCP client:

```go
// In backend
mcpClient := mcp.NewClient()
mcpClient.ConnectServer("dynamic-thinking", "./mcp-dynamic-thinking/mcp-dynamic-thinking")

// Call PRAR loop
result, err := mcpClient.CallTool("dynamic-thinking", "perceive", args)
```

## Memory Management

- **Short-term Memory**: Task-specific, cleared after completion
- **Long-term Memory**: Archived successful patterns and strategies
- **Automatic Cleanup**: Old task memories are automatically archived

## Strategy Evolution

Strategies evolve based on execution results:
- Successful executions → Confidence +0.05
- Failed executions → Confidence -0.10
- Patterns detected → New strategies created

## License

MIT

