# Complete MCP Server Architecture

All MCP servers for the self-evolving agent system.

## Overview

The agent uses **5 MCP servers** to achieve full self-awareness and capabilities:

1. **Dynamic Thinking** - Core reasoning and self-awareness
2. **OpenEvolve** - Code evolution via rewards
3. **Terminal Agent** - Natural language to Linux commands
4. **Chrome DevTools** - Browser automation and debugging
5. **Playwright** - Advanced browser automation

---

## 1. Dynamic Thinking MCP Server

**Location**: `backend/mcp_servers/dynamic_thinking/`

### Purpose
Core self-awareness implementing PRAR loop (Perceive-Reason-Act-Reflect).

### Tools (6)

#### `perceive`
Deep analytical understanding using:
- Systems thinking (how things work as a whole)
- Contextual reasoning (how it relates to situation)
- Meta-perception (better ways to perceive)
- Deductive reasoning (principles → conclusions)
- Inductive reasoning (observations → patterns)
- Abductive reasoning (observation → best explanation)

**Input**:
```json
{
  "task_id": "optimize-handler",
  "task": "Optimize HandleTask function",
  "goal": "Improve performance",
  "entity": {"name": "HandleTask", "type": "function"}
}
```

**Output**: perception_id, systems_view, reasoning, confidence

#### `reason`
Multi-branch reasoning with confidence-based online retrieval.

**Features**:
- Generates 3 reasoning branches
- Calculates confidence for each
- Triggers online search if confidence < 0.6
- Searches: Web, GitHub, Stack Overflow, YouTube
- Selects best branch

**Output**: reasoning_id, branches, selected_branch, retrieved_info

#### `act`
Dynamic execution with watchdog monitoring.

**Features**:
- Executes action plan
- Monitors with watchdog
- Parallel execution for cross-entity tasks
- Dynamic plan adjustment

**Output**: execution_id, results, performance_gain, watchdog_events

#### `reflect`
Self-critique and pattern creation.

**Features**:
- Extracts learnings
- Creates reusable patterns
- Generates recommendations for similar entities
- Produces training data for next generation

**Output**: reflection_id, learnings, patterns, recommendations

#### `query_memory`
Semantic search across agent's memory.

**Input**:
```json
{
  "query": "How to optimize request handlers?",
  "filters": {"type": "pattern", "success": true}
}
```

**Output**: Relevant past experiences with metadata

#### `evolve_prompt`
Improve prompts based on past performance.

**Output**: evolved_prompt, changes, expected_improvement

### Dependencies
- LightRAG (Neo4j + ChromaDB + Ollama)
- Gemma 3 (27B) for reasoning
- nomic-embed-text (v1.5) for embeddings

---

## 2. OpenEvolve MCP Server

**Location**: `backend/mcp_servers/openevolve/`

### Purpose
Code evolution via reward-based learning.

### Tools (3)

#### `evolve_code`
Evolve code through multiple generations.

**Input**:
```json
{
  "code": "func HandleTask() { ... }",
  "language": "go",
  "goal": "Improve performance and security",
  "iterations": 100,
  "population_size": 20
}
```

**Process**:
1. Detect patterns in code
2. Calculate reward score
3. Generate variations
4. Evaluate each variation
5. Select best performers
6. Repeat for N generations

**Output**:
```json
{
  "session_id": "abc-123",
  "original_score": -12,
  "final_score": 88,
  "improvement": 100,
  "best_code": "...",
  "patterns_applied": ["error_handling", "input_validation"]
}
```

#### `evaluate_code`
Score code without evolution.

**Output**: score, patterns_detected, rewards

#### `get_evolution_status`
Check ongoing evolution progress.

### Configuration
Uses reward structure from `config/openevolve/rewards.yaml`:
- Positive rewards: +3 to +40
- Negative penalties: -3 to -20

---

## 3. Terminal Agent MCP Server

**Location**: `backend/mcp_servers/terminal/`

### Purpose
Natural language to Linux commands using specialized model.

### Model
`comanderanch/Linux-Buster:latest` - Trained on 250+ commands

### Tools (3)

#### `natural_to_command`
Convert natural language to Linux command.

**Input**:
```json
{
  "instruction": "List all files larger than 100MB",
  "context": "In the /var/log directory"
}
```

**Output**:
```json
{
  "command": "find /var/log -type f -size +100M -ls",
  "safe": true,
  "explanation": "..."
}
```

#### `execute_command`
Execute command with safety validation.

**Safety Checks**:
- Blocks `rm -rf /`
- Blocks fork bombs
- Blocks `mkfs`, `dd if=/dev/zero`
- Blocks piping to shell

**Input**:
```json
{
  "command": "ls -la",
  "dry_run": false
}
```

**Output**: exit_code, stdout, stderr, success

#### `explain_command`
Explain what a command does.

---

## 4. Chrome DevTools MCP Server

**Location**: Official MCP package

### Installation
```bash
npm install -g @modelcontextprotocol/server-chrome-devtools
```

### Tools

#### Browser Control
- `navigate` - Go to URL
- `click` - Click element
- `type` - Type text
- `screenshot` - Capture screen

#### DOM Inspection
- `query_selector` - Find elements
- `get_html` - Get page HTML
- `get_text` - Extract text

#### Console
- `execute_script` - Run JavaScript
- `get_console_logs` - View console

#### Network
- `get_network_logs` - Network requests
- `intercept_request` - Modify requests

#### Performance
- `get_performance_metrics` - Performance data
- `start_profiling` - CPU/memory profiling

---

## 5. Playwright MCP Server

**Location**: Official MCP package

### Installation
```bash
npm install -g @modelcontextprotocol/server-playwright
```

### Purpose
Advanced browser automation with multi-browser support.

**Features**:
- Chromium, Firefox, WebKit support
- Headless and headed modes
- Network interception
- Mobile emulation
- Video recording

---

## Master Configuration

**File**: `config/mcp_config.json`

```json
{
  "mcpServers": {
    "dynamic-thinking": { ... },
    "openevolve": { ... },
    "terminal-agent": { ... },
    "chrome-devtools": { ... },
    "playwright": { ... }
  }
}
```

---

## How Gemma 3 Uses These Tools

### Scenario: Optimize Code

```
1. User: "Optimize the HandleTask function"

2. Gemma 3 calls: dynamic-thinking.perceive
   → Gets deep understanding of code and system

3. Gemma 3 calls: dynamic-thinking.reason
   → Generates optimization strategies
   → If confidence low, searches online automatically

4. Gemma 3 calls: openevolve.evolve_code
   → Evolves code through 100 generations
   → Applies patterns and rewards

5. Gemma 3 calls: dynamic-thinking.act
   → Executes the optimized code
   → Monitors with watchdog

6. Gemma 3 calls: dynamic-thinking.reflect
   → Learns from the experience
   → Creates reusable pattern
   → Stores in LightRAG
```

### Scenario: Terminal Operation

```
1. User: "Find large log files"

2. Gemma 3 calls: terminal-agent.natural_to_command
   → Converts to: find /var/log -type f -size +100M

3. Gemma 3 calls: terminal-agent.execute_command
   → Executes safely
   → Returns results
```

### Scenario: Web Automation

```
1. User: "Check if the website is responsive"

2. Gemma 3 calls: chrome-devtools.navigate
   → Opens website

3. Gemma 3 calls: chrome-devtools.get_performance_metrics
   → Gets load time, FCP, LCP

4. Gemma 3 calls: chrome-devtools.screenshot
   → Captures visual proof

5. Gemma 3 calls: dynamic-thinking.reflect
   → Stores findings in memory
```

---

## Installation

### 1. Install Python MCP Servers
```bash
cd backend/mcp_servers/dynamic_thinking && pip install -r requirements.txt
cd ../openevolve && pip install -r requirements.txt
cd ../terminal && pip install -r requirements.txt
```

### 2. Install Node MCP Servers
```bash
npm install -g @modelcontextprotocol/server-chrome-devtools
npm install -g @modelcontextprotocol/server-playwright
```

### 3. Configure Environment
```bash
export NEO4J_URI="bolt://localhost:7687"
export NEO4J_PASSWORD="your_password"
export OLLAMA_BASE_URL="http://localhost:11434"
```

### 4. Pull Ollama Models
```bash
ollama pull gemma3:27b
ollama pull nomic-embed-text:v1.5
ollama pull comanderanch/Linux-Buster:latest
```

---

## Architecture Diagram

```
Gemma 3 (Main Agent)
    ↓ (function calling)
    ├─→ Dynamic Thinking MCP
    │   ├─ perceive (systems thinking + 3 reasoning modes)
    │   ├─ reason (multi-branch + online retrieval)
    │   ├─ act (execution + watchdog)
    │   ├─ reflect (learning + patterns)
    │   ├─ query_memory (semantic search)
    │   └─ evolve_prompt (prompt improvement)
    │       ↓
    │   LightRAG (Neo4j + ChromaDB + Ollama)
    │
    ├─→ OpenEvolve MCP
    │   ├─ evolve_code (reward-based evolution)
    │   ├─ evaluate_code (pattern detection)
    │   └─ get_evolution_status (progress tracking)
    │
    ├─→ Terminal Agent MCP
    │   ├─ natural_to_command (NL → command)
    │   ├─ execute_command (safe execution)
    │   └─ explain_command (command explanation)
    │       ↓
    │   Linux-Buster Model (Ollama)
    │
    ├─→ Chrome DevTools MCP
    │   ├─ Browser control (navigate, click, type)
    │   ├─ DOM inspection (query, get HTML)
    │   ├─ Console (execute script, logs)
    │   ├─ Network (logs, intercept)
    │   └─ Performance (metrics, profiling)
    │
    └─→ Playwright MCP
        └─ Advanced browser automation
```

---

## Summary

### Total Tools Available to Gemma 3

- **Dynamic Thinking**: 6 tools (PRAR loop)
- **OpenEvolve**: 3 tools (code evolution)
- **Terminal Agent**: 3 tools (Linux commands)
- **Chrome DevTools**: 10+ tools (browser automation)
- **Playwright**: 15+ tools (advanced automation)

**Total: 35+ tools for complete self-awareness and capability!**

### Key Features

✅ Deep reasoning with systems thinking  
✅ Code evolution via rewards  
✅ Natural language terminal control  
✅ Browser automation and debugging  
✅ Graph-aware decision making  
✅ Confidence-based online retrieval  
✅ Pattern creation and learning  
✅ Training data generation  

**The agent can now think, evolve, execute, and learn!** 🧠🚀

