# EvoAgentX Integration Guide

## 🧬 Overview

This project integrates **EvoAgentX**, a self-evolving agent framework, with **Gemma 3** (via Ollama) to create an intelligent system that continuously improves its workflows, strategies, and decision-making processes.

---

## 🏗️ Architecture

### High-Level Integration

```
User Request
    ↓
Agent Controller (Go)
    ↓
EvoX Adapter ←→ Gemma 3 (Ollama)
    ↓
EvoAgentX Framework (Python)
    ↓
Workflow Evolution & Optimization
    ↓
Action Execution (Browser, Terminal, MCP)
    ↓
Results → Neo4j (Learning)
```

### Component Interaction

```
┌─────────────────────────────────────────────────────────┐
│                   Frontend (React)                       │
│  User sends command via WebSocket                        │
└───────────────────────┬─────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│              Go Fiber Backend                            │
│  ┌──────────────────────────────────────────────┐       │
│  │  Agent Controller                            │       │
│  │  - Receives user command                     │       │
│  │  - Calls EvoX Adapter                        │       │
│  └──────────────────┬───────────────────────────┘       │
│                     ↓                                    │
│  ┌──────────────────────────────────────────────┐       │
│  │  EvoX Adapter                                │       │
│  │  - Wraps Gemma 3 for EvoAgentX compatibility│       │
│  │  - Implements EvoX protocol                  │       │
│  └──────────────────┬───────────────────────────┘       │
└────────────────────┼────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│              Gemma 3 (Ollama)                            │
│  - Reasoning and planning                                │
│  - Workflow generation                                   │
│  - Action parsing                                        │
└───────────────────────┬─────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│         EvoAgentX Framework (Python - Optional)          │
│  - Advanced workflow evolution                           │
│  - Multi-agent coordination                              │
│  - Strategy optimization                                 │
└───────────────────────┬─────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│              Execution & Learning                        │
│  Browser ←→ Terminal ←→ MCP Tools ←→ Neo4j              │
└─────────────────────────────────────────────────────────┘
```

---

## 🔧 How It Works

### 1. **EvoX Adapter Layer**

The `EvoXAdapter` (in `backend/internal/agent/evox_adapter.go`) provides a bridge between our Go backend and EvoAgentX's Python framework.

**Key Methods:**

```go
// Generate workflow from goal
workflow, err := adapter.GenerateWorkflow(ctx, EvoXWorkflowRequest{
    Goal: "Build a login page with JWT authentication",
    Tools: []string{"browser", "terminal", "mcp"},
    Context: map[string]interface{}{
        "project_type": "web",
        "framework": "React",
    },
})

// Parse action from observation
action, err := adapter.ParseAction(ctx, EvoXActionRequest{
    Observation: "Browser shows GitHub homepage",
    Goal: "Find go-light-rag repository",
    AvailableTools: []string{"browser", "search"},
})

// Evaluate execution results
evaluation, err := adapter.Evaluate(ctx, EvoXEvaluationRequest{
    Task: "Research LightRAG implementation",
    Actions: []string{"search", "navigate", "analyze"},
    Result: "Found repository and documented API",
    Success: true,
})
```

### 2. **Workflow Evolution**

EvoAgentX continuously evolves workflows based on success/failure patterns:

**Initial Workflow (Generated by Gemma 3):**
```json
{
  "goal": "Create authentication system",
  "steps": [
    {"action": "search", "params": {"query": "JWT best practices"}},
    {"action": "code", "params": {"file": "auth.go", "content": "..."}},
    {"action": "test", "params": {"command": "go test ./..."}}
  ]
}
```

**After Learning (Evolved by EvoAgentX):**
```json
{
  "goal": "Create authentication system",
  "steps": [
    {"action": "check_existing", "params": {"pattern": "auth"}},  // NEW: Check first
    {"action": "search", "params": {"query": "JWT best practices Go"}},  // IMPROVED: More specific
    {"action": "code", "params": {"file": "auth.go", "content": "...", "validate": true}},  // NEW: Validation
    {"action": "security_scan", "params": {"tool": "gosec"}},  // NEW: Security check
    {"action": "test", "params": {"command": "go test ./...", "coverage": true}}  // IMPROVED: Coverage
  ],
  "confidence": 0.92,  // Higher confidence after learning
  "success_rate": 0.87  // Tracked from history
}
```

### 3. **Self-Evolution Cycle**

```
┌─────────────────────────────────────────────────────────┐
│  1. EXECUTE                                              │
│     User gives task → Agent executes workflow            │
└───────────────────┬─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│  2. OBSERVE                                              │
│     Collect results, errors, performance metrics         │
└───────────────────┬─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│  3. EVALUATE                                             │
│     EvoX analyzes: What worked? What failed? Why?        │
└───────────────────┬─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│  4. EVOLVE                                               │
│     Update workflow, adjust strategies, improve prompts  │
└───────────────────┬─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│  5. STORE                                                │
│     Save to Neo4j: Patterns, strategies, learnings       │
└───────────────────┬─────────────────────────────────────┘
                    ↓
                 (Repeat)
```

---

## 🧠 Gemma 3 Integration

### Prompt Engineering for Evolution

EvoAgentX uses carefully crafted prompts with Gemma 3:

**Workflow Generation Prompt:**
```
You are an expert AI agent that generates optimal workflows.

GOAL: {user_goal}
AVAILABLE TOOLS: {tools}
PAST SUCCESSES: {successful_patterns}
PAST FAILURES: {failed_patterns}

Generate a step-by-step workflow in JSON format that:
1. Uses available tools effectively
2. Learns from past successes
3. Avoids past failures
4. Includes validation and error handling
5. Optimizes for efficiency

Output ONLY valid JSON with this structure:
{
  "steps": [...],
  "reasoning": "...",
  "confidence": 0.0-1.0
}
```

**Action Parsing Prompt:**
```
You are observing an agent's environment and deciding the next action.

CURRENT OBSERVATION: {observation}
GOAL: {goal}
AVAILABLE TOOLS: {tools}
PREVIOUS ACTIONS: {history}

What should the agent do next? Output ONLY valid JSON:
{
  "action": "tool_name",
  "params": {...},
  "reasoning": "..."
}
```

**Evaluation Prompt:**
```
You are evaluating the agent's performance on a task.

TASK: {task}
ACTIONS TAKEN: {actions}
RESULT: {result}
SUCCESS: {success}

Analyze the execution and provide insights in JSON:
{
  "success": true/false,
  "efficiency": 0.0-1.0,
  "improvements": [...],
  "patterns_learned": [...],
  "mistakes_to_avoid": [...]
}
```

---

## 📊 Learning & Storage

### Neo4j Knowledge Graph

All EvoAgentX learnings are stored in Neo4j for future reference:

**Nodes:**
- `(:Workflow)` - Successful workflow patterns
- `(:Strategy)` - Evolved strategies
- `(:Pattern)` - Recognized patterns
- `(:Failure)` - Failed attempts (to avoid)
- `(:Optimization)` - Performance improvements

**Relationships:**
- `(:Workflow)-[:EVOLVED_TO]->(:Workflow)` - Workflow evolution
- `(:Strategy)-[:IMPROVES]->(:Strategy)` - Strategy refinement
- `(:Pattern)-[:LEADS_TO]->(:Success)` - Pattern → outcome
- `(:Failure)-[:AVOIDED_BY]->(:Strategy)` - Failure avoidance

**Example Query:**
```cypher
// Find most successful workflow for authentication tasks
MATCH (w:Workflow {category: 'authentication'})
WHERE w.success_rate > 0.8
RETURN w.steps, w.success_rate, w.avg_duration
ORDER BY w.success_rate DESC, w.avg_duration ASC
LIMIT 1
```

---

## 🔄 Evolution Examples

### Example 1: Browser Automation Evolution

**Initial Workflow (Day 1):**
```json
{
  "task": "Find repository on GitHub",
  "steps": [
    {"action": "browser_navigate", "url": "https://github.com"},
    {"action": "browser_type", "selector": "input[name='q']", "text": "go-light-rag"},
    {"action": "browser_click", "selector": "button[type='submit']"}
  ],
  "success_rate": 0.6,
  "avg_duration": 15.2
}
```

**Evolved Workflow (After 50 executions):**
```json
{
  "task": "Find repository on GitHub",
  "steps": [
    {"action": "check_cache", "key": "github:go-light-rag"},  // NEW: Cache check
    {"action": "browser_navigate", "url": "https://github.com/search?q=go-light-rag&type=repositories"},  // OPTIMIZED: Direct search URL
    {"action": "wait_for_element", "selector": ".repo-list-item", "timeout": 5000},  // NEW: Wait for results
    {"action": "browser_click", "selector": ".repo-list-item:first-child a"}  // IMPROVED: First result
  ],
  "success_rate": 0.95,  // Improved!
  "avg_duration": 8.3,   // Faster!
  "optimizations": [
    "Added cache layer",
    "Use direct search URL instead of homepage",
    "Wait for elements before clicking",
    "Always click first result (most relevant)"
  ]
}
```

### Example 2: Code Generation Evolution

**Initial Approach:**
```
1. Generate entire file at once
2. No validation
3. No testing
```

**Evolved Approach:**
```
1. Check existing code for patterns
2. Generate in small, testable chunks
3. Validate syntax after each chunk
4. Run tests incrementally
5. Use linter for quality checks
6. Store successful patterns for reuse
```

---

## 🚀 API Endpoints

### EvoX-Specific Endpoints

**POST /api/evox/workflow**
Generate evolved workflow for a goal
```json
{
  "goal": "Create REST API with authentication",
  "tools": ["terminal", "mcp", "browser"],
  "context": {
    "language": "Go",
    "framework": "Fiber"
  }
}
```

**POST /api/evox/action**
Parse next action from observation
```json
{
  "observation": "Server started on port 8080",
  "goal": "Test API endpoints",
  "available_tools": ["browser", "terminal"]
}
```

**POST /api/evox/evaluate**
Evaluate task execution
```json
{
  "task": "Deploy application",
  "actions": ["build", "test", "deploy"],
  "result": "Successfully deployed",
  "success": true,
  "duration": 45.2
}
```

**GET /api/evox/patterns**
Retrieve learned patterns
```json
{
  "category": "authentication",
  "min_success_rate": 0.8
}
```

---

## 🔧 Configuration

### Environment Variables

```bash
# EvoAgentX Settings
EVOX_ENABLED=true
EVOX_LEARNING_RATE=0.1
EVOX_MIN_CONFIDENCE=0.7
EVOX_MAX_RETRIES=3

# Gemma 3 Settings
OLLAMA_MODEL=gemma3:27b
OLLAMA_TEMPERATURE=0.7
OLLAMA_MAX_TOKENS=4096

# Evolution Settings
EVOLUTION_ENABLED=true
EVOLUTION_THRESHOLD=10  # Min executions before evolution
EVOLUTION_INTERVAL=3600  # Evolve every hour
```

### Backend Configuration

```go
// backend/cmd/server/main.go
evoxAdapter := agent.NewEvoXAdapter(
    ollamaClient,
    &agent.EvoXConfig{
        Enabled:        true,
        LearningRate:   0.1,
        MinConfidence:  0.7,
        MaxRetries:     3,
        EvolutionThreshold: 10,
    },
)
```

---

## 📈 Monitoring Evolution

### Metrics Tracked

1. **Workflow Success Rate**
   - Percentage of successful executions
   - Tracked per workflow type

2. **Efficiency Improvements**
   - Average duration reduction
   - Resource usage optimization

3. **Pattern Recognition**
   - Number of patterns learned
   - Pattern reuse frequency

4. **Error Reduction**
   - Decrease in failure rate over time
   - Common errors avoided

### Dashboard Queries

```cypher
// Evolution progress over time
MATCH (w:Workflow)
WHERE w.created_at > datetime() - duration('P30D')
RETURN 
  date(w.created_at) as day,
  avg(w.success_rate) as avg_success,
  avg(w.duration) as avg_duration
ORDER BY day

// Most improved workflows
MATCH (w1:Workflow)-[:EVOLVED_TO]->(w2:Workflow)
WHERE w2.success_rate > w1.success_rate
RETURN 
  w1.task,
  w1.success_rate as before,
  w2.success_rate as after,
  (w2.success_rate - w1.success_rate) as improvement
ORDER BY improvement DESC
LIMIT 10
```

---

## 🎯 Best Practices

### 1. **Start Simple**
- Begin with basic workflows
- Let EvoAgentX learn gradually
- Don't over-engineer initial prompts

### 2. **Provide Feedback**
- Mark successful executions
- Document failures with reasons
- Add context to evaluations

### 3. **Monitor Learning**
- Check evolution metrics weekly
- Review learned patterns
- Prune ineffective strategies

### 4. **Balance Exploration vs Exploitation**
- Allow some randomness for discovery
- But favor proven patterns for production

### 5. **Version Control Workflows**
- Store workflow versions in Neo4j
- Track evolution history
- Rollback if needed

---

## 🔬 Advanced Features

### Multi-Agent Coordination

EvoAgentX can coordinate multiple agents:

```go
// Spawn specialized agents
searchAgent := evox.SpawnAgent("search", []string{"browser", "mcp"})
codeAgent := evox.SpawnAgent("code", []string{"terminal", "mcp"})
testAgent := evox.SpawnAgent("test", []string{"terminal"})

// Coordinate workflow
workflow := evox.CoordinateAgents([]Agent{
    searchAgent,  // Research phase
    codeAgent,    // Implementation phase
    testAgent,    // Validation phase
})
```

### Prompt Evolution

EvoAgentX can evolve its own prompts:

```go
// Initial prompt
prompt := "Generate a login page"

// After learning
evolvedPrompt := evox.EvolvePrompt(prompt, feedback)
// Result: "Generate a secure login page with JWT authentication, 
//          input validation, CSRF protection, and rate limiting"
```

### Strategy Trees

Build decision trees for complex tasks:

```
Task: Deploy Application
├─ Check Tests
│  ├─ All Pass → Continue
│  └─ Some Fail → Fix & Retry
├─ Build
│  ├─ Success → Continue
│  └─ Fail → Check Dependencies
└─ Deploy
   ├─ Staging First
   ├─ Run Smoke Tests
   └─ Production (if staging OK)
```

---

## 📚 Resources

- **EvoAgentX GitHub**: https://github.com/EvoAgentX/EvoAgentX
- **Gemma 3 Documentation**: https://ollama.ai/library/gemma3
- **Neo4j Graph Patterns**: https://neo4j.com/docs/
- **Our Implementation**: `backend/internal/agent/evox_adapter.go`

---

## 🤝 Contributing

To improve EvoAgentX integration:

1. Add new evolution strategies in `evox_adapter.go`
2. Create better prompts in `gemma.go`
3. Add metrics in `planner.go`
4. Document patterns in Neo4j

---

## 📝 Summary

EvoAgentX integration enables:
- ✅ Self-improving workflows
- ✅ Pattern recognition and reuse
- ✅ Continuous optimization
- ✅ Failure avoidance
- ✅ Knowledge accumulation

The system gets **smarter with every task**, building a knowledge base that makes future executions faster, more reliable, and more efficient.

**Result:** An agent that doesn't just execute commands—it **learns, adapts, and evolves**. 🧬🚀

