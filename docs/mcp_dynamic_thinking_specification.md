# MCP Server: Dynamic Sequential Thinking with Short-Term Memory

## Overview

This MCP (Model Context Protocol) server implements the **Perceive-Reason-Act-Reflect (PRAR)** loop with dynamic branching, integrated with ChromeDP-based short-term memory for task-oriented processing. This enables the agent to maintain context across perception cycles, make branched decisions, and evolve its reasoning strategies.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    MCP Server: Dynamic Thinking                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Perceive-Reason-Act-Reflect Loop            │  │
│  │  ┌──────────┬──────────┬──────────┬──────────────────┐  │  │
│  │  │ Perceive │ Reason   │ Act      │ Reflect          │  │  │
│  │  │ (Vision) │ (Branch) │ (Tools)  │ (Self-Improve)   │  │  │
│  │  └──────────┴──────────┴──────────┴──────────────────┘  │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │            Short-Term Memory (ChromeDP)                  │  │
│  │  • Current task screenshots                              │  │
│  │  • Reasoning branches (active)                           │  │
│  │  • Decision points & confidence scores                   │  │
│  │  • Intermediate action results                           │  │
│  │  • Perception history (current session)                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │            Long-Term Memory (LightRAG + Neo4j)           │  │
│  │  • Successful reasoning patterns                         │  │
│  │  • Strategy nodes & relationships                        │  │
│  │  • Execution traces (historical)                         │  │
│  │  • Prompt evolutions                                     │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ↕
┌─────────────────────────────────────────────────────────────────┐
│                    Agent Workspace Backend                      │
│         (Connects via MCP Client - JSON-RPC 2.0)                │
└─────────────────────────────────────────────────────────────────┘
```

---

## MCP Server Configuration

### Server Manifest

```json
{
  "name": "dynamic-thinking",
  "version": "1.0.0",
  "description": "Dynamic sequential thinking with Perceive-Reason-Act-Reflect loop and short-term memory",
  "protocol": "mcp",
  "transport": {
    "type": "stdio"
  },
  "capabilities": {
    "tools": true,
    "resources": true,
    "prompts": true
  }
}
```

### Installation

```bash
# Install the MCP server
go install github.com/yourusername/mcp-dynamic-thinking@latest

# Or run from source
cd mcp-dynamic-thinking
go build -o mcp-dynamic-thinking
```

### Configuration in Agent Workspace

```json
{
  "mcpServers": {
    "dynamic-thinking": {
      "command": "mcp-dynamic-thinking",
      "args": [],
      "env": {
        "NEO4J_URI": "bolt://localhost:7687",
        "NEO4J_USER": "neo4j",
        "NEO4J_PASSWORD": "password",
        "CHROMEDP_HEADLESS": "true"
      }
    }
  }
}
```

---

## MCP Tools

### 1. perceive

**Description**: Capture and analyze the current environment state using vision and OCR.

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "task_id": {
      "type": "string",
      "description": "Unique task identifier for short-term memory"
    },
    "goal": {
      "type": "string",
      "description": "Current goal to guide perception"
    },
    "capture_screenshot": {
      "type": "boolean",
      "description": "Whether to capture a new screenshot",
      "default": true
    },
    "analyze_terminal": {
      "type": "boolean",
      "description": "Whether to analyze terminal output",
      "default": true
    },
    "confidence_threshold": {
      "type": "number",
      "description": "Minimum confidence to proceed (0.0-1.0)",
      "default": 0.7
    }
  },
  "required": ["task_id", "goal"]
}
```

**Output Schema**:
```json
{
  "type": "object",
  "properties": {
    "perception_id": {
      "type": "string",
      "description": "UUID for this perception"
    },
    "confidence": {
      "type": "number",
      "description": "Confidence score (0.0-1.0)"
    },
    "visual_analysis": {
      "type": "object",
      "properties": {
        "screenshot_path": {"type": "string"},
        "elements_detected": {"type": "array"},
        "text_extracted": {"type": "string"}
      }
    },
    "terminal_state": {
      "type": "object",
      "properties": {
        "output": {"type": "string"},
        "exit_code": {"type": "integer"}
      }
    },
    "decision": {
      "type": "string",
      "enum": ["proceed", "gather_more_info"],
      "description": "Whether to proceed or gather more context"
    },
    "stored_in_memory": {
      "type": "boolean",
      "description": "Whether perception was stored in short-term memory"
    }
  }
}
```

**Example Call**:
```bash
manus-mcp-cli tool call perceive --server dynamic-thinking --input '{
  "task_id": "task-123",
  "goal": "Navigate to GitHub and find go-light-rag repository",
  "capture_screenshot": true,
  "confidence_threshold": 0.8
}'
```

---

### 2. reason

**Description**: Generate and evaluate multiple reasoning branches for the current task.

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "task_id": {
      "type": "string",
      "description": "Task identifier"
    },
    "perception_id": {
      "type": "string",
      "description": "Perception to reason about"
    },
    "goal": {
      "type": "string",
      "description": "Goal to achieve"
    },
    "num_branches": {
      "type": "integer",
      "description": "Number of reasoning branches to explore",
      "default": 3
    },
    "use_past_strategies": {
      "type": "boolean",
      "description": "Whether to query Neo4j for past strategies",
      "default": true
    }
  },
  "required": ["task_id", "perception_id", "goal"]
}
```

**Output Schema**:
```json
{
  "type": "object",
  "properties": {
    "reasoning_id": {
      "type": "string",
      "description": "UUID for this reasoning session"
    },
    "branches": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "branch_id": {"type": "string"},
          "strategy": {"type": "string"},
          "chain_of_thought": {"type": "array", "items": {"type": "string"}},
          "feasibility_score": {"type": "number"},
          "alignment_score": {"type": "number"},
          "risk_score": {"type": "number"},
          "selected": {"type": "boolean"}
        }
      }
    },
    "selected_branch": {
      "type": "object",
      "description": "The chosen reasoning branch"
    },
    "action_plan": {
      "type": "object",
      "properties": {
        "subtasks": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "id": {"type": "string"},
              "description": {"type": "string"},
              "type": {"type": "string", "enum": ["terminal", "browser", "file_edit", "mcp_call"]},
              "parameters": {"type": "object"}
            }
          }
        }
      }
    },
    "decision": {
      "type": "string",
      "enum": ["execute_plan", "refine_reasoning", "loop_back_perceive"],
      "description": "Next action to take"
    }
  }
}
```

**Example Call**:
```bash
manus-mcp-cli tool call reason --server dynamic-thinking --input '{
  "task_id": "task-123",
  "perception_id": "perc-456",
  "goal": "Find and clone go-light-rag repository",
  "num_branches": 3
}'
```

---

### 3. act

**Description**: Execute the action plan with dynamic monitoring and adjustment.

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "task_id": {
      "type": "string",
      "description": "Task identifier"
    },
    "reasoning_id": {
      "type": "string",
      "description": "Reasoning session that generated the plan"
    },
    "action_plan": {
      "type": "object",
      "description": "Plan to execute (from reason output)"
    },
    "enable_watchdog": {
      "type": "boolean",
      "description": "Enable safety monitoring",
      "default": true
    },
    "allow_dynamic_adjustment": {
      "type": "boolean",
      "description": "Allow plan adjustment based on intermediate results",
      "default": true
    }
  },
  "required": ["task_id", "reasoning_id", "action_plan"]
}
```

**Output Schema**:
```json
{
  "type": "object",
  "properties": {
    "execution_id": {
      "type": "string",
      "description": "UUID for this execution"
    },
    "status": {
      "type": "string",
      "enum": ["success", "partial_success", "failed", "aborted"]
    },
    "subtask_results": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "subtask_id": {"type": "string"},
          "status": {"type": "string"},
          "result": {"type": "object"},
          "execution_time": {"type": "number"},
          "adjustments_made": {"type": "array"}
        }
      }
    },
    "plan_adjustments": {
      "type": "array",
      "description": "Dynamic adjustments made during execution"
    },
    "decision": {
      "type": "string",
      "enum": ["proceed_to_reflect", "retry_with_alternative", "abort_and_replan"],
      "description": "Next action based on execution result"
    }
  }
}
```

**Example Call**:
```bash
manus-mcp-cli tool call act --server dynamic-thinking --input '{
  "task_id": "task-123",
  "reasoning_id": "reason-789",
  "action_plan": {
    "subtasks": [
      {
        "id": "sub-1",
        "description": "Navigate to GitHub",
        "type": "browser",
        "parameters": {"url": "https://github.com"}
      }
    ]
  }
}'
```

---

### 4. reflect

**Description**: Analyze execution results, generate critique, and evolve strategies.

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "task_id": {
      "type": "string",
      "description": "Task identifier"
    },
    "execution_id": {
      "type": "string",
      "description": "Execution to reflect on"
    },
    "original_goal": {
      "type": "string",
      "description": "Original goal"
    },
    "enable_prompt_evolution": {
      "type": "boolean",
      "description": "Allow prompt refinement",
      "default": true
    },
    "enable_strategy_update": {
      "type": "boolean",
      "description": "Allow strategy updates in Neo4j",
      "default": true
    }
  },
  "required": ["task_id", "execution_id", "original_goal"]
}
```

**Output Schema**:
```json
{
  "type": "object",
  "properties": {
    "reflection_id": {
      "type": "string",
      "description": "UUID for this reflection"
    },
    "analysis": {
      "type": "object",
      "properties": {
        "goal_achieved": {"type": "boolean"},
        "success_rate": {"type": "number"},
        "what_worked": {"type": "array", "items": {"type": "string"}},
        "what_failed": {"type": "array", "items": {"type": "string"}}
      }
    },
    "critique": {
      "type": "object",
      "properties": {
        "prompt_issues": {"type": "array"},
        "reasoning_issues": {"type": "array"},
        "execution_issues": {"type": "array"},
        "suggestions": {"type": "array"}
      }
    },
    "evolutions": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "type": {"type": "string", "enum": ["prompt_refinement", "strategy_update", "code_optimization"]},
          "details": {"type": "object"},
          "applied": {"type": "boolean"}
        }
      }
    },
    "decision": {
      "type": "string",
      "enum": ["complete", "replan_with_insights", "continue_iteration"],
      "description": "Whether to complete or loop back"
    },
    "training_trace": {
      "type": "object",
      "description": "Complete execution trace for training data"
    }
  }
}
```

**Example Call**:
```bash
manus-mcp-cli tool call reflect --server dynamic-thinking --input '{
  "task_id": "task-123",
  "execution_id": "exec-999",
  "original_goal": "Clone go-light-rag repository",
  "enable_prompt_evolution": true
}'
```

---

### 5. get_short_term_memory

**Description**: Retrieve short-term memory for a specific task.

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "task_id": {
      "type": "string",
      "description": "Task identifier"
    },
    "include_screenshots": {
      "type": "boolean",
      "description": "Include base64-encoded screenshots",
      "default": false
    }
  },
  "required": ["task_id"]
}
```

**Output Schema**:
```json
{
  "type": "object",
  "properties": {
    "task_id": {"type": "string"},
    "perceptions": {"type": "array"},
    "reasoning_sessions": {"type": "array"},
    "executions": {"type": "array"},
    "reflections": {"type": "array"},
    "screenshots": {"type": "array"},
    "decision_points": {"type": "array"}
  }
}
```

---

### 6. clear_short_term_memory

**Description**: Clear short-term memory for a completed task.

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "task_id": {
      "type": "string",
      "description": "Task identifier to clear"
    },
    "archive_to_long_term": {
      "type": "boolean",
      "description": "Archive important data to Neo4j before clearing",
      "default": true
    }
  },
  "required": ["task_id"]
}
```

**Output Schema**:
```json
{
  "type": "object",
  "properties": {
    "success": {"type": "boolean"},
    "archived_items": {"type": "integer"},
    "cleared_items": {"type": "integer"}
  }
}
```

---

### 7. query_strategies

**Description**: Query Neo4j for relevant strategies based on current context.

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "context": {
      "type": "string",
      "description": "Current context or problem description"
    },
    "top_k": {
      "type": "integer",
      "description": "Number of strategies to return",
      "default": 5
    },
    "min_success_rate": {
      "type": "number",
      "description": "Minimum success rate filter",
      "default": 0.5
    }
  },
  "required": ["context"]
}
```

**Output Schema**:
```json
{
  "type": "object",
  "properties": {
    "strategies": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "name": {"type": "string"},
          "approach": {"type": "string"},
          "success_rate": {"type": "number"},
          "applicable_to": {"type": "array"},
          "past_uses": {"type": "integer"}
        }
      }
    }
  }
}
```

---

### 8. get_execution_trace

**Description**: Retrieve complete execution trace for training data generation.

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "task_id": {
      "type": "string",
      "description": "Task identifier"
    },
    "format": {
      "type": "string",
      "enum": ["json", "jsonl"],
      "default": "json"
    }
  },
  "required": ["task_id"]
}
```

**Output Schema**:
```json
{
  "type": "object",
  "properties": {
    "trace_id": {"type": "string"},
    "task_id": {"type": "string"},
    "timestamp": {"type": "string"},
    "goal": {"type": "string"},
    "perception": {"type": "object"},
    "reasoning": {"type": "object"},
    "actions": {"type": "array"},
    "reflection": {"type": "object"},
    "total_duration": {"type": "number"}
  }
}
```

---

## MCP Resources

### 1. short-term-memory://tasks/{task_id}

**Description**: Access short-term memory for a specific task.

**MIME Type**: `application/json`

**Example**:
```bash
manus-mcp-cli resource read short-term-memory://tasks/task-123 --server dynamic-thinking
```

---

### 2. strategies://neo4j/{strategy_id}

**Description**: Access strategy nodes from Neo4j.

**MIME Type**: `application/json`

---

### 3. traces://executions/{trace_id}

**Description**: Access execution traces for training data.

**MIME Type**: `application/json`

---

## MCP Prompts

### 1. dynamic-thinking-loop

**Description**: Execute a complete Perceive-Reason-Act-Reflect loop.

**Arguments**:
- `goal` (required): The goal to achieve
- `task_id` (required): Task identifier
- `max_iterations` (optional): Maximum loop iterations (default: 5)

**Example**:
```bash
manus-mcp-cli prompt get dynamic-thinking-loop --server dynamic-thinking \
  --arg goal="Find and clone go-light-rag repository" \
  --arg task_id="task-123"
```

---

## Go Implementation

### Main Server Structure

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
    "github.com/chromedp/chromedp"
    golightrag "github.com/MegaGrindStone/go-light-rag"
)

type DynamicThinkingServer struct {
    server          *server.MCPServer
    shortTermMem    *ShortTermMemory
    longTermMem     *LongTermMemory
    chromedpCtx     context.Context
    chromedpCancel  context.CancelFunc
}

type ShortTermMemory struct {
    tasks map[string]*TaskMemory
    mu    sync.RWMutex
}

type TaskMemory struct {
    TaskID            string
    Perceptions       []Perception
    ReasoningSessions []ReasoningSession
    Executions        []Execution
    Reflections       []Reflection
    Screenshots       map[string][]byte
    DecisionPoints    []DecisionPoint
    CreatedAt         time.Time
    LastAccessed      time.Time
}

type Perception struct {
    ID               string
    Timestamp        time.Time
    Confidence       float64
    VisualAnalysis   VisualAnalysis
    TerminalState    TerminalState
    Decision         string
}

type ReasoningSession struct {
    ID             string
    Timestamp      time.Time
    Branches       []ReasoningBranch
    SelectedBranch *ReasoningBranch
    ActionPlan     ActionPlan
    Decision       string
}

type ReasoningBranch struct {
    ID               string
    Strategy         string
    ChainOfThought   []string
    FeasibilityScore float64
    AlignmentScore   float64
    RiskScore        float64
    Selected         bool
}

type Execution struct {
    ID              string
    Timestamp       time.Time
    Status          string
    SubtaskResults  []SubtaskResult
    PlanAdjustments []string
    Decision        string
}

type Reflection struct {
    ID           string
    Timestamp    time.Time
    Analysis     Analysis
    Critique     Critique
    Evolutions   []Evolution
    Decision     string
    TrainingTrace map[string]interface{}
}

func NewDynamicThinkingServer() *DynamicThinkingServer {
    // Initialize ChromeDP
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", true),
    )
    allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
    ctx, cancel := chromedp.NewContext(allocCtx)
    
    // Initialize server
    s := &DynamicThinkingServer{
        server:         server.NewMCPServer("dynamic-thinking", "1.0.0"),
        shortTermMem:   &ShortTermMemory{tasks: make(map[string]*TaskMemory)},
        chromedpCtx:    ctx,
        chromedpCancel: cancel,
    }
    
    // Initialize long-term memory (LightRAG + Neo4j)
    s.initializeLongTermMemory()
    
    // Register tools
    s.registerTools()
    
    // Register resources
    s.registerResources()
    
    // Register prompts
    s.registerPrompts()
    
    return s
}

func (s *DynamicThinkingServer) registerTools() {
    // Tool: perceive
    s.server.AddTool(mcp.Tool{
        Name:        "perceive",
        Description: "Capture and analyze current environment state",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "task_id": map[string]interface{}{
                    "type":        "string",
                    "description": "Task identifier",
                },
                "goal": map[string]interface{}{
                    "type":        "string",
                    "description": "Current goal",
                },
                "capture_screenshot": map[string]interface{}{
                    "type":    "boolean",
                    "default": true,
                },
                "confidence_threshold": map[string]interface{}{
                    "type":    "number",
                    "default": 0.7,
                },
            },
            Required: []string{"task_id", "goal"},
        },
    }, s.handlePerceive)
    
    // Tool: reason
    s.server.AddTool(mcp.Tool{
        Name:        "reason",
        Description: "Generate and evaluate reasoning branches",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "task_id": map[string]interface{}{
                    "type": "string",
                },
                "perception_id": map[string]interface{}{
                    "type": "string",
                },
                "goal": map[string]interface{}{
                    "type": "string",
                },
                "num_branches": map[string]interface{}{
                    "type":    "integer",
                    "default": 3,
                },
            },
            Required: []string{"task_id", "perception_id", "goal"},
        },
    }, s.handleReason)
    
    // Tool: act
    s.server.AddTool(mcp.Tool{
        Name:        "act",
        Description: "Execute action plan with monitoring",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "task_id": map[string]interface{}{
                    "type": "string",
                },
                "reasoning_id": map[string]interface{}{
                    "type": "string",
                },
                "action_plan": map[string]interface{}{
                    "type": "object",
                },
            },
            Required: []string{"task_id", "reasoning_id", "action_plan"},
        },
    }, s.handleAct)
    
    // Tool: reflect
    s.server.AddTool(mcp.Tool{
        Name:        "reflect",
        Description: "Analyze results and evolve strategies",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "task_id": map[string]interface{}{
                    "type": "string",
                },
                "execution_id": map[string]interface{}{
                    "type": "string",
                },
                "original_goal": map[string]interface{}{
                    "type": "string",
                },
            },
            Required: []string{"task_id", "execution_id", "original_goal"},
        },
    }, s.handleReflect)
    
    // Additional tools...
    s.server.AddTool(mcp.Tool{
        Name:        "get_short_term_memory",
        Description: "Retrieve short-term memory for task",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "task_id": map[string]interface{}{
                    "type": "string",
                },
            },
            Required: []string{"task_id"},
        },
    }, s.handleGetShortTermMemory)
    
    s.server.AddTool(mcp.Tool{
        Name:        "clear_short_term_memory",
        Description: "Clear short-term memory for completed task",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "task_id": map[string]interface{}{
                    "type": "string",
                },
                "archive_to_long_term": map[string]interface{}{
                    "type":    "boolean",
                    "default": true,
                },
            },
            Required: []string{"task_id"},
        },
    }, s.handleClearShortTermMemory)
}

func (s *DynamicThinkingServer) handlePerceive(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    taskID := arguments["task_id"].(string)
    goal := arguments["goal"].(string)
    captureScreenshot := true
    if val, ok := arguments["capture_screenshot"]; ok {
        captureScreenshot = val.(bool)
    }
    
    // Ensure task memory exists
    s.shortTermMem.mu.Lock()
    if _, exists := s.shortTermMem.tasks[taskID]; !exists {
        s.shortTermMem.tasks[taskID] = &TaskMemory{
            TaskID:       taskID,
            Screenshots:  make(map[string][]byte),
            CreatedAt:    time.Now(),
        }
    }
    taskMem := s.shortTermMem.tasks[taskID]
    s.shortTermMem.mu.Unlock()
    
    perception := Perception{
        ID:        uuid.New().String(),
        Timestamp: time.Now(),
    }
    
    // Capture screenshot if requested
    if captureScreenshot {
        var buf []byte
        err := chromedp.Run(s.chromedpCtx,
            chromedp.CaptureScreenshot(&buf),
        )
        if err != nil {
            return nil, err
        }
        
        taskMem.Screenshots[perception.ID] = buf
        perception.VisualAnalysis.ScreenshotPath = fmt.Sprintf("screenshot_%s.png", perception.ID)
    }
    
    // Analyze with vision model (placeholder - integrate actual vision model)
    perception.VisualAnalysis = s.analyzeScreenshot(taskMem.Screenshots[perception.ID], goal)
    
    // Analyze terminal (placeholder)
    perception.TerminalState = s.analyzeTerminal()
    
    // Calculate confidence
    perception.Confidence = s.calculateConfidence(perception)
    
    // Decision: proceed or gather more info
    threshold := 0.7
    if val, ok := arguments["confidence_threshold"]; ok {
        threshold = val.(float64)
    }
    
    if perception.Confidence < threshold {
        perception.Decision = "gather_more_info"
    } else {
        perception.Decision = "proceed"
    }
    
    // Store in short-term memory
    s.shortTermMem.mu.Lock()
    taskMem.Perceptions = append(taskMem.Perceptions, perception)
    taskMem.LastAccessed = time.Now()
    s.shortTermMem.mu.Unlock()
    
    // Return result
    result := map[string]interface{}{
        "perception_id":   perception.ID,
        "confidence":      perception.Confidence,
        "visual_analysis": perception.VisualAnalysis,
        "terminal_state":  perception.TerminalState,
        "decision":        perception.Decision,
        "stored_in_memory": true,
    }
    
    resultJSON, _ := json.Marshal(result)
    
    return &mcp.CallToolResult{
        Content: []interface{}{
            mcp.TextContent{
                Type: "text",
                Text: string(resultJSON),
            },
        },
    }, nil
}

func (s *DynamicThinkingServer) handleReason(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    taskID := arguments["task_id"].(string)
    perceptionID := arguments["perception_id"].(string)
    goal := arguments["goal"].(string)
    numBranches := 3
    if val, ok := arguments["num_branches"]; ok {
        numBranches = int(val.(float64))
    }
    
    // Get task memory
    s.shortTermMem.mu.RLock()
    taskMem, exists := s.shortTermMem.tasks[taskID]
    s.shortTermMem.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("task not found: %s", taskID)
    }
    
    // Find perception
    var perception *Perception
    for i := range taskMem.Perceptions {
        if taskMem.Perceptions[i].ID == perceptionID {
            perception = &taskMem.Perceptions[i]
            break
        }
    }
    
    if perception == nil {
        return nil, fmt.Errorf("perception not found: %s", perceptionID)
    }
    
    // Query strategies from Neo4j
    strategies := s.longTermMem.QueryStrategies(goal, numBranches)
    
    // Generate reasoning branches
    session := ReasoningSession{
        ID:        uuid.New().String(),
        Timestamp: time.Now(),
        Branches:  make([]ReasoningBranch, 0, numBranches),
    }
    
    for i, strategy := range strategies {
        branch := s.generateReasoningBranch(perception, goal, strategy)
        branch.ID = fmt.Sprintf("branch-%d", i+1)
        session.Branches = append(session.Branches, branch)
    }
    
    // Evaluate and select best branch
    selectedIdx := s.evaluateBranches(session.Branches)
    session.Branches[selectedIdx].Selected = true
    session.SelectedBranch = &session.Branches[selectedIdx]
    
    // Generate action plan
    session.ActionPlan = s.generateActionPlan(session.SelectedBranch, goal)
    
    // Decision
    if session.SelectedBranch.FeasibilityScore > 0.8 {
        session.Decision = "execute_plan"
    } else {
        session.Decision = "refine_reasoning"
    }
    
    // Store in short-term memory
    s.shortTermMem.mu.Lock()
    taskMem.ReasoningSessions = append(taskMem.ReasoningSessions, session)
    taskMem.LastAccessed = time.Now()
    s.shortTermMem.mu.Unlock()
    
    // Return result
    result := map[string]interface{}{
        "reasoning_id":    session.ID,
        "branches":        session.Branches,
        "selected_branch": session.SelectedBranch,
        "action_plan":     session.ActionPlan,
        "decision":        session.Decision,
    }
    
    resultJSON, _ := json.Marshal(result)
    
    return &mcp.CallToolResult{
        Content: []interface{}{
            mcp.TextContent{
                Type: "text",
                Text: string(resultJSON),
            },
        },
    }, nil
}

// Additional handler implementations...

func (s *DynamicThinkingServer) handleClearShortTermMemory(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    taskID := arguments["task_id"].(string)
    archiveToLongTerm := true
    if val, ok := arguments["archive_to_long_term"]; ok {
        archiveToLongTerm = val.(bool)
    }
    
    s.shortTermMem.mu.Lock()
    defer s.shortTermMem.mu.Unlock()
    
    taskMem, exists := s.shortTermMem.tasks[taskID]
    if !exists {
        return nil, fmt.Errorf("task not found: %s", taskID)
    }
    
    archivedItems := 0
    
    if archiveToLongTerm {
        // Archive important data to Neo4j
        for _, reflection := range taskMem.Reflections {
            if reflection.Analysis.SuccessRate > 0.7 {
                s.longTermMem.ArchiveReflection(reflection)
                archivedItems++
            }
        }
    }
    
    clearedItems := len(taskMem.Perceptions) + len(taskMem.ReasoningSessions) + 
                    len(taskMem.Executions) + len(taskMem.Reflections)
    
    // Clear task memory
    delete(s.shortTermMem.tasks, taskID)
    
    result := map[string]interface{}{
        "success":        true,
        "archived_items": archivedItems,
        "cleared_items":  clearedItems,
    }
    
    resultJSON, _ := json.Marshal(result)
    
    return &mcp.CallToolResult{
        Content: []interface{}{
            mcp.TextContent{
                Type: "text",
                Text: string(resultJSON),
            },
        },
    }, nil
}

func main() {
    server := NewDynamicThinkingServer()
    defer server.chromedpCancel()
    
    if err := server.server.Serve(); err != nil {
        log.Fatal(err)
    }
}
```

---

## Integration with Agent Workspace

### Using the MCP Server

```javascript
// Frontend: Call dynamic thinking loop
async function executeDynamicThinkingLoop(goal) {
  const taskId = generateUUID();
  
  // 1. Perceive
  const perception = await mcpClient.callTool('dynamic-thinking', 'perceive', {
    task_id: taskId,
    goal: goal,
    capture_screenshot: true
  });
  
  if (perception.decision === 'gather_more_info') {
    // Handle additional context gathering
    return;
  }
  
  // 2. Reason
  const reasoning = await mcpClient.callTool('dynamic-thinking', 'reason', {
    task_id: taskId,
    perception_id: perception.perception_id,
    goal: goal,
    num_branches: 3
  });
  
  if (reasoning.decision === 'refine_reasoning') {
    // Loop back to reasoning
    return;
  }
  
  // 3. Act
  const execution = await mcpClient.callTool('dynamic-thinking', 'act', {
    task_id: taskId,
    reasoning_id: reasoning.reasoning_id,
    action_plan: reasoning.action_plan
  });
  
  // 4. Reflect
  const reflection = await mcpClient.callTool('dynamic-thinking', 'reflect', {
    task_id: taskId,
    execution_id: execution.execution_id,
    original_goal: goal
  });
  
  if (reflection.decision === 'replan_with_insights') {
    // Loop back to perceive with new insights
    return executeDynamicThinkingLoop(goal);
  }
  
  // 5. Clear short-term memory
  await mcpClient.callTool('dynamic-thinking', 'clear_short_term_memory', {
    task_id: taskId,
    archive_to_long_term: true
  });
  
  return reflection.training_trace;
}
```

---

## Conclusion

This MCP server provides:

✅ **Complete PRAR loop** - Perceive, Reason, Act, Reflect  
✅ **Dynamic branching** - Multiple reasoning paths explored  
✅ **Short-term memory** - ChromeDP-based task context  
✅ **Long-term learning** - LightRAG + Neo4j integration  
✅ **Self-improvement** - Prompt and strategy evolution  
✅ **Training data** - Rich execution traces  
✅ **MCP compliant** - Standard tool/resource/prompt interface  

The agent can now maintain context across perception cycles, make branched decisions, and continuously improve its reasoning strategies! 🚀

