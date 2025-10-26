# Dynamic Thinking MCP Server

## Overview

This MCP server implements the **Perceive-Reason-Act-Reflect (PRAR)** loop with dynamic branching, enabling the agent to think about its own code, make adaptive decisions, and continuously self-improve. This is a **critical component** for true self-awareness and autonomous evolution.

## Why This is Essential

Traditional agents follow rigid pipelines. This MCP server enables:

- **Dynamic branching**: Explore multiple reasoning paths in parallel
- **Adaptive decisions**: Loop back, refine, or abort based on confidence
- **Self-referential thinking**: Agent reasons about its own code
- **Continuous evolution**: Prompts and strategies improve over time
- **Rich memory integration**: Neo4j + LightRAG + ChromeM for learning

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│            Dynamic Thinking MCP Server (Go)                 │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  PRAR Loop                                            │ │
│  │  ┌──────────┬──────────┬──────────┬──────────────┐  │ │
│  │  │ Perceive │ Reason   │ Act      │ Reflect      │  │ │
│  │  │ (Vision) │ (Branch) │ (Execute)│ (Evolve)     │  │ │
│  │  └──────────┴──────────┴──────────┴──────────────┘  │ │
│  └───────────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  Short-Term Memory (ChromeM - in-memory)              │ │
│  │  • Current task context                               │ │
│  │  • Active reasoning branches                          │ │
│  │  • Decision points & confidence scores                │ │
│  │  • Intermediate results                               │ │
│  └───────────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  Long-Term Memory (Neo4j + LightRAG)                  │ │
│  │  • Successful reasoning patterns                      │ │
│  │  • Strategy nodes & relationships                     │ │
│  │  • Execution traces (historical)                      │ │
│  │  • Prompt evolutions                                  │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                          ↕ JSON-RPC 2.0
┌─────────────────────────────────────────────────────────────┐
│                  Main Agent (Gemma 3)                       │
│  Uses dynamic thinking for self-aware reasoning            │
└─────────────────────────────────────────────────────────────┘
```

## MCP Tools

### 1. `perceive`

**Purpose**: Capture and analyze current environment state with vision

**Input**:
```json
{
  "task_id": "task-123",
  "goal": "Understand current code structure",
  "capture_screenshot": true,
  "analyze_code_mirror": true,
  "confidence_threshold": 0.7
}
```

**Output**:
```json
{
  "perception_id": "perc-456",
  "confidence": 0.85,
  "visual_analysis": {
    "screenshot_path": "/tmp/screenshot.png",
    "elements_detected": ["browser", "terminal", "editor"],
    "text_extracted": "..."
  },
  "code_mirror_state": {
    "files_in_context": ["controller.go", "agent.go"],
    "current_function": "HandleTask",
    "recent_changes": [...]
  },
  "decision": "proceed",
  "stored_in_memory": true
}
```

**What it does**:
- Takes screenshot of current environment
- Analyzes with Gemma 3 vision
- Queries Neo4j code mirror for current code state
- Stores perception in ChromeM short-term memory
- Returns confidence score and decision

### 2. `reason`

**Purpose**: Generate and evaluate multiple reasoning branches

**Input**:
```json
{
  "task_id": "task-123",
  "perception_id": "perc-456",
  "goal": "Improve HandleTask function performance",
  "num_branches": 3,
  "use_past_strategies": true
}
```

**Output**:
```json
{
  "reasoning_id": "reason-789",
  "branches": [
    {
      "branch_id": "branch-1",
      "strategy": "Add caching layer",
      "chain_of_thought": [
        "Current function queries Neo4j on every call",
        "Results are often identical for same task type",
        "Cache with TTL would reduce database load",
        "Risk: Stale data if task changes rapidly"
      ],
      "feasibility_score": 0.9,
      "alignment_score": 0.85,
      "risk_score": 0.3,
      "selected": true
    },
    {
      "branch_id": "branch-2",
      "strategy": "Parallel execution",
      "chain_of_thought": [...],
      "feasibility_score": 0.7,
      "alignment_score": 0.9,
      "risk_score": 0.5,
      "selected": false
    },
    {
      "branch_id": "branch-3",
      "strategy": "Optimize Neo4j query",
      "chain_of_thought": [...],
      "feasibility_score": 0.8,
      "alignment_score": 0.8,
      "risk_score": 0.2,
      "selected": false
    }
  ],
  "selected_branch": {
    "branch_id": "branch-1",
    "rationale": "Highest feasibility with acceptable risk"
  },
  "action_plan": {
    "subtasks": [
      {
        "id": "subtask-1",
        "description": "Add cache struct to controller",
        "type": "code_edit",
        "parameters": {
          "file": "backend/internal/agent/controller.go",
          "function": "AgentController",
          "change_type": "add_field"
        }
      },
      {
        "id": "subtask-2",
        "description": "Implement cache lookup in HandleTask",
        "type": "code_edit",
        "parameters": {
          "file": "backend/internal/agent/controller.go",
          "function": "HandleTask",
          "change_type": "add_logic"
        }
      }
    ]
  },
  "decision": "execute_plan"
}
```

**What it does**:
- Queries Neo4j for similar past strategies
- Uses LightRAG for semantic search of successful patterns
- Generates 3 reasoning branches with Gemma 3
- Evaluates each branch (feasibility, alignment, risk)
- Selects best branch or synthesizes hybrid
- Creates concrete action plan with subtasks
- Stores reasoning in ChromeM and Neo4j

### 3. `act`

**Purpose**: Execute action plan with dynamic monitoring

**Input**:
```json
{
  "task_id": "task-123",
  "reasoning_id": "reason-789",
  "action_plan": {...},
  "enable_watchdog": true,
  "allow_dynamic_adjustment": true
}
```

**Output**:
```json
{
  "execution_id": "exec-101",
  "status": "success",
  "subtask_results": [
    {
      "subtask_id": "subtask-1",
      "status": "success",
      "result": {
        "file_modified": "backend/internal/agent/controller.go",
        "lines_added": 5,
        "compilation_success": true
      },
      "execution_time": 2.3,
      "adjustments_made": []
    },
    {
      "subtask_id": "subtask-2",
      "status": "partial_success",
      "result": {
        "file_modified": "backend/internal/agent/controller.go",
        "lines_added": 15,
        "compilation_success": true,
        "tests_passed": 8,
        "tests_failed": 2
      },
      "execution_time": 5.1,
      "adjustments_made": [
        "Added nil check for cache miss",
        "Fixed race condition with mutex"
      ]
    }
  ],
  "plan_adjustments": [
    "Added error handling not in original plan",
    "Implemented thread-safe cache access"
  ],
  "decision": "proceed_to_reflect"
}
```

**What it does**:
- Executes each subtask in sequence
- Monitors execution with watchdog
- Dynamically adjusts plan based on intermediate results
- Handles failures with retry or alternative approaches
- Stores execution trace in Neo4j
- Returns detailed results for reflection

### 4. `reflect`

**Purpose**: Analyze results, critique approach, and evolve strategies

**Input**:
```json
{
  "task_id": "task-123",
  "execution_id": "exec-101",
  "original_goal": "Improve HandleTask function performance",
  "enable_evolution": true
}
```

**Output**:
```json
{
  "reflection_id": "reflect-202",
  "analysis": {
    "goal_achieved": true,
    "performance_improvement": "35% faster",
    "success_factors": [
      "Cache reduced database calls by 80%",
      "Thread-safe implementation prevented race conditions"
    ],
    "failure_factors": [
      "Two tests failed due to cache invalidation timing",
      "Memory usage increased by 15MB"
    ]
  },
  "critique": {
    "what_worked": "Caching strategy was effective",
    "what_failed": "Cache invalidation logic needs refinement",
    "prompt_issues": "Original reasoning didn't consider memory overhead",
    "strategy_issues": "Should have included memory profiling in plan"
  },
  "evolutions": [
    {
      "type": "prompt_refinement",
      "target": "reasoning_template",
      "old_version": "v1.2",
      "new_version": "v1.3",
      "changes": "Added memory consideration to feasibility evaluation",
      "stored_in_neo4j": true
    },
    {
      "type": "strategy_update",
      "strategy_name": "performance_optimization",
      "updates": {
        "add_step": "Profile memory usage before and after",
        "add_criterion": "Memory overhead < 20MB"
      },
      "stored_in_neo4j": true
    },
    {
      "type": "code_pattern",
      "pattern_name": "cache_with_ttl",
      "success_rate": 0.9,
      "applicable_to": ["database_optimization", "api_performance"],
      "stored_in_neo4j": true
    }
  ],
  "training_trace": {
    "trace_id": "trace-303",
    "stored_in_neo4j": true,
    "includes": [
      "perception",
      "all_reasoning_branches",
      "action_plan",
      "execution_results",
      "reflection",
      "evolutions"
    ]
  },
  "decision": "complete"
}
```

**What it does**:
- Analyzes execution results vs original goal
- Generates detailed critique with Gemma 3
- Identifies what worked and what failed
- Evolves prompts, strategies, and patterns
- Stores evolutions in Neo4j for future use
- Creates complete training trace
- Decides whether to loop back or complete

### 5. `query_memory`

**Purpose**: Query past experiences for similar situations

**Input**:
```json
{
  "query_type": "similar_tasks",
  "current_context": {
    "goal": "Optimize database query",
    "code_file": "controller.go",
    "function": "HandleTask"
  },
  "top_k": 5
}
```

**Output**:
```json
{
  "results": [
    {
      "trace_id": "trace-101",
      "similarity": 0.92,
      "task": "Optimize Neo4j query performance",
      "strategy_used": "Add caching layer",
      "success": true,
      "performance_gain": "40%"
    },
    ...
  ]
}
```

**What it does**:
- Uses LightRAG for semantic search
- Queries Neo4j for graph-based relationships
- Returns similar past experiences
- Includes success/failure information

### 6. `evolve_prompt`

**Purpose**: Evolve a prompt template based on critique

**Input**:
```json
{
  "prompt_type": "reasoning_template",
  "current_version": "v1.2",
  "critique": {
    "issues": ["Doesn't consider memory overhead"],
    "suggestions": ["Add memory profiling step"]
  }
}
```

**Output**:
```json
{
  "new_version": "v1.3",
  "changes": [
    "Added memory consideration section",
    "Updated evaluation criteria"
  ],
  "prompt_text": "...",
  "stored_in_neo4j": true
}
```

**What it does**:
- Takes current prompt and critique
- Uses Gemma 3 to generate improved version
- Validates improvement
- Stores in Neo4j with version history

## Integration with Agent

### How Main Agent Uses Dynamic Thinking

```go
// backend/internal/agent/controller.go
func (a *AgentController) HandleComplexTask(task *Task) error {
    // 1. Perceive current state
    perception, err := a.mcpClient.CallTool("dynamic-thinking", "perceive", map[string]interface{}{
        "task_id":              task.ID,
        "goal":                 task.Description,
        "capture_screenshot":   true,
        "analyze_code_mirror":  true,
        "confidence_threshold": 0.7,
    })
    if err != nil {
        return err
    }
    
    // Check if we have enough information
    if perception.Confidence < 0.7 {
        // Gather more context...
    }
    
    // 2. Reason with branching
    reasoning, err := a.mcpClient.CallTool("dynamic-thinking", "reason", map[string]interface{}{
        "task_id":             task.ID,
        "perception_id":       perception.ID,
        "goal":                task.Description,
        "num_branches":        3,
        "use_past_strategies": true,
    })
    if err != nil {
        return err
    }
    
    // 3. Act on selected plan
    execution, err := a.mcpClient.CallTool("dynamic-thinking", "act", map[string]interface{}{
        "task_id":                  task.ID,
        "reasoning_id":             reasoning.ID,
        "action_plan":              reasoning.ActionPlan,
        "enable_watchdog":          true,
        "allow_dynamic_adjustment": true,
    })
    if err != nil {
        return err
    }
    
    // 4. Reflect and evolve
    reflection, err := a.mcpClient.CallTool("dynamic-thinking", "reflect", map[string]interface{}{
        "task_id":         task.ID,
        "execution_id":    execution.ID,
        "original_goal":   task.Description,
        "enable_evolution": true,
    })
    if err != nil {
        return err
    }
    
    // 5. Check if we need to loop back
    if reflection.Decision == "replan" {
        // Loop back with new insights
        return a.HandleComplexTask(task)
    }
    
    return nil
}
```

## Memory Integration

### Short-Term Memory (ChromeM)

Stores current task context:

```go
type ShortTermMemory struct {
    TaskID          string
    Perceptions     []Perception
    ReasoningBranches []ReasoningBranch
    DecisionPoints  []DecisionPoint
    IntermediateResults []Result
}
```

**Lifetime**: Duration of current task (cleared after reflection)

### Long-Term Memory (Neo4j)

Stores learned patterns:

```cypher
// Strategy node
CREATE (s:Strategy {
    id: "strat-123",
    name: "cache_optimization",
    approach: "Add caching layer with TTL",
    success_rate: 0.9,
    applicable_to: ["database", "api"],
    created_at: datetime(),
    version: "1.0"
})

// Prompt evolution
CREATE (p:PromptTemplate {
    id: "prompt-456",
    type: "reasoning_template",
    version: "v1.3",
    text: "...",
    improvements_over_previous: ["Added memory consideration"],
    created_at: datetime()
})

// Execution trace
CREATE (t:ExecutionTrace {
    id: "trace-789",
    task_goal: "Optimize HandleTask performance",
    strategy_used: "cache_optimization",
    success: true,
    performance_gain: 0.35,
    created_at: datetime()
})

// Relationships
CREATE (t)-[:USED_STRATEGY]->(s)
CREATE (t)-[:USED_PROMPT]->(p)
CREATE (s)-[:EVOLVED_FROM]->(previous_strategy)
```

### Vector Memory (LightRAG)

Semantic search for similar situations:

```go
// Find similar past tasks
similar := lightRAG.Query(
    "Optimize database query performance in Go controller",
    topK: 5,
)

// Returns semantically similar traces even if different wording
```

## Configuration

### MCP Server Config

```json
{
  "mcpServers": {
    "dynamic-thinking": {
      "command": "mcp-dynamic-thinking",
      "args": [],
      "env": {
        "GEMMA_API_BASE": "http://localhost:11434/v1/",
        "GEMMA_MODEL": "gemma3:27b",
        "EMBEDDING_MODEL": "nomic-embed-text:v1.5",
        "NEO4J_URI": "bolt://localhost:7687",
        "NEO4J_USER": "neo4j",
        "NEO4J_PASSWORD": "${NEO4J_PASSWORD}",
        "CONFIDENCE_THRESHOLD": "0.7",
        "ALIGNMENT_THRESHOLD": "0.75",
        "NUM_REASONING_BRANCHES": "3"
      }
    }
  }
}
```

### Server Settings

```yaml
# config/dynamic_thinking.yaml
perception:
  screenshot_enabled: true
  code_mirror_enabled: true
  confidence_threshold: 0.7

reasoning:
  num_branches: 3
  use_past_strategies: true
  evaluation_criteria:
    feasibility: 0.4
    alignment: 0.3
    risk: 0.3

action:
  enable_watchdog: true
  allow_dynamic_adjustment: true
  timeout_per_subtask: 60

reflection:
  enable_evolution: true
  auto_store_traces: true
  prompt_evolution_threshold: 0.8

memory:
  short_term:
    backend: chromem
    dimension: 768
  long_term:
    backend: neo4j
    uri: bolt://localhost:7687
  vector:
    backend: lightrag
    embedding_model: nomic-embed-text:v1.5
```

## Implementation

### File Structure

```
backend/mcp_servers/dynamic_thinking/
├── main.go                 # MCP server entry point
├── server.go               # MCP server implementation
├── tools/
│   ├── perceive.go        # Perception tool
│   ├── reason.go          # Reasoning tool
│   ├── act.go             # Action tool
│   ├── reflect.go         # Reflection tool
│   ├── query_memory.go    # Memory query tool
│   └── evolve_prompt.go   # Prompt evolution tool
├── memory/
│   ├── short_term.go      # ChromeM integration
│   ├── long_term.go       # Neo4j integration
│   └── vector.go          # LightRAG integration
├── reasoning/
│   ├── branching.go       # Multi-branch reasoning
│   ├── evaluation.go      # Branch evaluation
│   └── decision.go        # Decision framework
├── evolution/
│   ├── prompt.go          # Prompt evolution
│   ├── strategy.go        # Strategy evolution
│   └── pattern.go         # Pattern learning
└── config.yaml            # Server configuration
```

## Success Criteria

The dynamic thinking MCP server is successful when:

1. ✅ **Adaptive Branching**: Agent explores multiple reasoning paths when faced with ambiguity
2. ✅ **Intelligent Looping**: Agent recognizes when to loop back and replan based on new insights
3. ✅ **Continuous Evolution**: Prompts and strategies improve measurably over iterations
4. ✅ **Goal Alignment**: All dynamic decisions maintain alignment with original goal
5. ✅ **Rich Training Data**: Each execution generates comprehensive traces for training next generation
6. ✅ **Self-Awareness**: Agent can reason about its own code through Neo4j code mirror
7. ✅ **Memory Integration**: Short-term and long-term memory work seamlessly together

## Next Steps

1. Implement MCP server in Go
2. Create perception tool with vision analysis
3. Implement multi-branch reasoning
4. Add action execution with watchdog
5. Implement reflection and evolution
6. Integrate with Neo4j code mirror
7. Test with real tasks
8. Measure improvement over iterations

This MCP server is the **core of self-awareness** - it enables the agent to think dynamically about its own code and continuously evolve!

