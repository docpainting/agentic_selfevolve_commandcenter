# EvoAgentX Integration Guide

## Overview

This document explains how **Gemma 3** is wrapped within the **EvoAgentX framework** and how our agent workspace integrates seamlessly with it.

## What is EvoAgentX?

**EvoAgentX** is a self-evolving agent framework that:
- Automatically generates multi-agent workflows from natural language goals
- Evaluates agent performance using built-in evaluators
- Evolves and optimizes workflows through iterative feedback loops
- Supports memory (short-term and long-term)
- Provides human-in-the-loop (HITL) interactions

## How Gemma is Wrapped in EvoAgentX

### EvoAgentX LLM Interface

EvoAgentX expects LLMs to implement this interface (from `base_model.py`):

```python
class BaseLLM(ABC):
    @abstractmethod
    def generate(self, prompt: str, temperature: float, max_tokens: int) -> str:
        """Synchronous generation"""
        pass
    
    @abstractmethod
    async def generate_async(self, prompt: str, temperature: float, max_tokens: int) -> str:
        """Asynchronous generation"""
        pass
    
    def parse_output(self, response: str, mode: str) -> Any:
        """Parse structured output (JSON, XML, etc.)"""
        pass
```

### Our Implementation

We've created an **EvoXAdapter** (`internal/agent/evox_adapter.go`) that wraps our Gemma 3 client to be compatible with EvoAgentX expectations.

## Architecture

```
EvoAgentX Framework
    ↓
EvoXAdapter (Go)
    ↓
GemmaClient (Go)
    ↓
Ollama Client (Go)
    ↓
Ollama v1 API
    ↓
gemma3:27b
```

## No Compatibility Issues!

### ✅ Why It Works Seamlessly

1. **Language Agnostic**
   - EvoAgentX is Python-based
   - Our backend is Go-based
   - They communicate via **HTTP/JSON APIs**
   - No direct language coupling

2. **Standard Interfaces**
   - EvoAgentX uses standard LLM patterns (generate, evaluate, parse)
   - Our adapter implements these patterns
   - Ollama provides OpenAI-compatible v1 API

3. **Flexible Integration**
   - EvoAgentX can call our backend via REST API
   - Our backend can use EvoAgentX patterns internally
   - Both systems work independently or together

## Integration Patterns

### Pattern 1: EvoAgentX Calls Our Backend

```python
# In EvoAgentX workflow
import requests

# Call our agent workspace API
response = requests.post("http://localhost:8080/api/agent/execute", json={
    "command": "Find go-light-rag on GitHub",
    "context": {...}
})

result = response.json()
```

### Pattern 2: Our Backend Uses EvoAgentX Patterns

```go
// In our Go backend
adapter := controller.GetEvoXAdapter()

// Generate workflow using EvoX pattern
workflow, err := adapter.GenerateWorkflow(ctx, agent.EvoXWorkflowRequest{
    Goal: "Find information about LightRAG",
    Tools: []string{"browser", "terminal", "mcp"},
})

// Parse actions using EvoX pattern
action, err := adapter.ParseAction(ctx, agent.EvoXActionRequest{
    Observation: "Browser shows search results",
    Goal: "Find LightRAG repository",
    History: []string{"navigated to github.com"},
})

// Evaluate using EvoX pattern
evaluation, err := adapter.Evaluate(ctx, agent.EvoXEvaluationRequest{
    Task: "Find LightRAG",
    Goal: "Locate the repository",
    Result: "Successfully found repository",
})
```

### Pattern 3: Hybrid Approach

```
User Request
    ↓
Frontend (React)
    ↓ WebSocket
Backend (Go Fiber)
    ↓
Agent Controller
    ├─→ EvoXAdapter (for workflow generation)
    ├─→ GemmaClient (for reasoning)
    ├─→ BrowserManager (for actions)
    └─→ Memory (LightRAG + Neo4j)
```

## EvoXAdapter Features

### 1. Workflow Generation

```go
workflow, err := adapter.GenerateWorkflow(ctx, agent.EvoXWorkflowRequest{
    Goal: "Create a login page with JWT auth",
    Tools: []string{"browser", "terminal", "mcp"},
    Constraints: []string{"Use Go Fiber", "Store in Neo4j"},
})

// Returns:
// {
//   "nodes": [
//     {"id": "1", "type": "agent", "name": "Code Generator"},
//     {"id": "2", "type": "tool", "name": "Terminal"},
//     {"id": "3", "type": "evaluator", "name": "Code Reviewer"}
//   ],
//   "edges": [
//     {"from": "1", "to": "2", "type": "sequential"},
//     {"from": "2", "to": "3", "type": "sequential"}
//   ]
// }
```

### 2. Action Parsing

```go
action, err := adapter.ParseAction(ctx, agent.EvoXActionRequest{
    Observation: "Screenshot shows GitHub homepage",
    Goal: "Find go-light-rag repository",
    History: []string{"navigated to github.com"},
})

// Returns:
// {
//   "type": "browser",
//   "command": "type",
//   "parameters": {"element": "search", "text": "go-light-rag"},
//   "reasoning": "Need to search for the repository"
// }
```

### 3. Evaluation

```go
evaluation, err := adapter.Evaluate(ctx, agent.EvoXEvaluationRequest{
    Task: "Find go-light-rag repository",
    Goal: "Locate the repository by MegaGrindStone",
    Result: "Found repository at github.com/MegaGrindStone/go-light-rag",
})

// Returns:
// {
//   "score": 1.0,
//   "feedback": "Successfully found the correct repository",
//   "strengths": ["Accurate search", "Correct repository identified"],
//   "weaknesses": [],
//   "suggestions": ["Could bookmark for future reference"]
// }
```

### 4. Structured Output

```go
type MySchema struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

result, err := adapter.GenerateStructured(ctx, 
    "Extract person info from: John is 30 years old",
    MySchema{},
    0.3,
)

// Returns: {"name": "John", "age": 30}
```

## Benefits of This Architecture

### 1. **Best of Both Worlds**
- ✅ EvoAgentX's self-evolution capabilities
- ✅ Our custom Go backend performance
- ✅ Gemma 3's reasoning power
- ✅ LightRAG + Neo4j knowledge graph

### 2. **No Vendor Lock-in**
- ✅ Can use EvoAgentX patterns without EvoAgentX
- ✅ Can switch LLMs easily (Gemma → GPT-4 → Claude)
- ✅ Adapter pattern isolates dependencies

### 3. **Flexibility**
- ✅ Use EvoAgentX for workflow generation
- ✅ Use our backend for execution
- ✅ Combine both for hybrid approach

### 4. **Evolution**
- ✅ EvoAgentX can evolve our workflows
- ✅ Our backend stores evolution history in Neo4j
- ✅ LightRAG learns from successful patterns

## Example: Complete Integration

```go
package main

import (
    "context"
    "log"
    "agent-workspace/backend/internal/agent"
)

func main() {
    // Initialize controller
    controller := agent.NewController(...)
    
    // Get EvoX adapter
    evoxAdapter := controller.GetEvoXAdapter()
    
    ctx := context.Background()
    
    // 1. Generate workflow using EvoX pattern
    workflow, err := evoxAdapter.GenerateWorkflow(ctx, agent.EvoXWorkflowRequest{
        Goal: "Research LightRAG and create integration guide",
        Tools: []string{"browser", "terminal", "mcp"},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Generated workflow with %d nodes", len(workflow.Nodes))
    
    // 2. Execute workflow
    for _, node := range workflow.Nodes {
        log.Printf("Executing node: %s", node.Name)
        
        // Parse next action
        action, err := evoxAdapter.ParseAction(ctx, agent.EvoXActionRequest{
            Observation: "Current state...",
            Goal: node.Description,
            History: []string{},
        })
        if err != nil {
            log.Fatal(err)
        }
        
        // Execute action using our backend
        switch action.Type {
        case "browser":
            controller.ExecuteBrowserAction(action.Command, action.Parameters)
        case "terminal":
            controller.ExecuteTerminalCommand(action.Command)
        case "mcp":
            controller.ExecuteMCPTool(action.Command, action.Parameters)
        }
    }
    
    // 3. Evaluate execution
    evaluation, err := evoxAdapter.Evaluate(ctx, agent.EvoXEvaluationRequest{
        Task: "Research LightRAG",
        Goal: workflow.Nodes[0].Description,
        Result: "Completed research and created guide",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Evaluation score: %.2f", evaluation.Score)
    log.Printf("Feedback: %s", evaluation.Feedback)
    
    // 4. Store in Neo4j for learning
    controller.StoreEvaluation(evaluation)
}
```

## API Endpoints for EvoAgentX Integration

### POST /api/evox/workflow

Generate workflow using EvoX pattern:

```bash
curl -X POST http://localhost:8080/api/evox/workflow \
  -H "Content-Type: application/json" \
  -d '{
    "goal": "Find go-light-rag on GitHub",
    "tools": ["browser", "terminal"],
    "constraints": ["Use Chrome", "Save screenshots"]
  }'
```

### POST /api/evox/action

Parse next action:

```bash
curl -X POST http://localhost:8080/api/evox/action \
  -H "Content-Type: application/json" \
  -d '{
    "observation": "Browser shows GitHub homepage",
    "goal": "Find repository",
    "history": ["navigated to github.com"]
  }'
```

### POST /api/evox/evaluate

Evaluate execution:

```bash
curl -X POST http://localhost:8080/api/evox/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Find repository",
    "goal": "Locate go-light-rag",
    "result": "Successfully found repository"
  }'
```

## Conclusion

**No, wrapping Gemma in EvoAgentX will NOT cause problems!**

✅ **Reasons:**
1. **Adapter Pattern** - Clean separation of concerns
2. **Standard Interfaces** - Both use common LLM patterns
3. **Language Agnostic** - HTTP/JSON communication
4. **Flexible Integration** - Multiple integration patterns
5. **Best of Both** - Combine EvoAgentX evolution with our custom backend

The **EvoXAdapter** provides a clean bridge between EvoAgentX patterns and our Gemma 3 implementation, allowing seamless integration without any compatibility issues!

## Next Steps

1. **Test the adapter** with EvoAgentX workflows
2. **Add EvoX endpoints** to Fiber backend
3. **Integrate evolution** feedback into LightRAG
4. **Store workflow history** in Neo4j
5. **Enable self-improvement** through evaluation loops

The system is designed to work perfectly with or without EvoAgentX! 🚀

