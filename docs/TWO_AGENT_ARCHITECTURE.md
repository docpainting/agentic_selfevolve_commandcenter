# Two-Agent Architecture: Main + Terminal

## Overview

The workspace uses a **two-agent architecture** designed for efficiency and specialization:

1. **Main Agent** (Gemma 3 27B) - Orchestrator with self-awareness, evolution, and memory
2. **Terminal Agent** (Linux-Buster) - Natural language → Linux commands (250+ commands)

Later, the main agent can **train new agents from scratch** using its accumulated memory and knowledge.

---

## Model Stack

### Main Agent
- **LLM**: `gemma3:27b` (27GB, multimodal, 8K context)
- **Embeddings**: `nomic-embed-text:v1.5` (764MB, 8K context, 768 dimensions)
- **Memory**: Neo4j + LightRAG + ChromeM (in-memory vector DB)

### Terminal Agent
- **LLM**: `comanderanch/Linux-Buster:latest` (~7GB, 250+ commands)
- **Tools**: PTY manager, command validator, output parser

**All models run locally via Ollama** - no API costs, full privacy, complete control.

---

## Why Two Agents?

### Main Agent (Gemma 3)
- **Role**: Orchestrator, planner, self-aware entity
- **Strengths**: Reasoning, planning, self-modification, learning
- **Tools**: Browser, file operations, MCP servers (OpenEvolve, etc.)
- **Memory**: Neo4j knowledge graph, LightRAG semantic search, ChromeM vectors

### Terminal Agent (Linux-Buster)
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
│                Main Agent (Gemma 3 27B)                     │
│  ┌────────────────────────────────────────────────────┐    │
│  │  LLM: gemma3:27b (via Ollama)                      │    │
│  │  Embeddings: nomic-embed-text:v1.5 (via Ollama)    │    │
│  └────────────────────────────────────────────────────┘    │
│                                                             │
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
│          Terminal Agent (comanderanch/Linux-Buster)         │
│  ┌────────────────────────────────────────────────────┐    │
│  │  LLM: comanderanch/Linux-Buster:latest             │    │
│  │  (via Ollama)                                      │    │
│  └────────────────────────────────────────────────────┘    │
│                                                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Core Capability:                                  │    │
│  │  Natural Language → Linux Commands                 │    │
│  │                                                     │    │
│  │  Training:                                         │    │
│  │  - 250+ Linux commands                             │    │
│  │  - Shell scripting                                 │    │
│  │  - System administration                           │    │
│  │  - Package management (apt, dpkg, snap)            │    │
│  │  - Process management (ps, systemctl, kill)        │    │
│  │  - File operations (ls, cp, mv, find, grep)        │    │
│  │  - Network tools (curl, wget, netstat)             │    │
│  │  - Development tools (git, npm, pip, docker)       │    │
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
Terminal Agent (Linux-Buster):
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
    Uses Gemma 3 + nomic-embed-text for evolution
    ↓
    Returns: {best_code: "...", score: 88}
    ↓
Main Agent:
    Stores in Neo4j
    Creates embedding with nomic-embed-text
    Stores in ChromeM for similarity search
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
Terminal Agent (Linux-Buster):
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
    Step 4: Store deployment in Neo4j with embedding
    ↓
    Responds to user: "Deployed successfully! Coverage: 85%"
```

---

## Configuration

### Agent Configuration

**File**: `config/agents.yaml`

```yaml
agents:
  main:
    id: main_agent
    name: "Main Agent"
    
    llm:
      model: gemma3:27b
      api_base: http://localhost:11434/v1/
      api_key: ollama
      temperature: 0.7
      max_tokens: 8192
    
    embedding:
      model: nomic-embed-text:v1.5
      api_base: http://localhost:11434/v1/
      dimension: 768
    
    memory:
      lightrag:
        enabled: true
        config_file: config/lightrag.yaml
      neo4j:
        enabled: true
        uri: bolt://localhost:7687
      chromem:
        enabled: true
        dimension: 768
    
    tools:
      builtin:
        - browser_navigate
        - browser_click
        - file_read
        - file_write
      mcp:
        - openevolve.*
        - playwright.*
      agents:
        - terminal_agent.execute
    
    capabilities:
      self_awareness: true
      self_modification: true
      pattern_learning: true
      evolution: true
  
  terminal:
    id: terminal_agent
    name: "Terminal Agent"
    
    llm:
      model: comanderanch/Linux-Buster:latest
      api_base: http://localhost:11434/v1/
      api_key: ollama
      temperature: 0.3  # Lower for precise commands
      max_tokens: 2048
    
    tools:
      builtin:
        - pty_execute
        - validate_command
        - parse_output
    
    a2a:
      enabled: true
      listen_port: 8081
      methods:
        - terminal.execute
    
    capabilities:
      command_execution: true
      safety_validation: true
```

### LightRAG Configuration

**File**: `config/lightrag.yaml`

```yaml
llm:
  model: gemma3:27b
  api_base: http://localhost:11434/v1/
  api_key: ollama
  temperature: 0.7
  max_tokens: 8192

embedding:
  model: nomic-embed-text:v1.5
  api_base: http://localhost:11434/v1/
  dimension: 768
  context_length: 8192

storage:
  vector:
    backend: chromem
    dimension: 768
  graph:
    backend: neo4j
    uri: bolt://localhost:7687
    username: neo4j
    password: ${NEO4J_PASSWORD}
  kv:
    backend: boltdb
    path: data/lightrag.db
```

---

## Future: Training New Agents

### Main Agent Trains New Agent

The main agent accumulates knowledge through:
- **Neo4j**: Stores code patterns, successful solutions, learned best practices
- **LightRAG**: Semantic search over past experiences
- **ChromeM**: Vector similarity for finding related solutions
- **Embeddings**: nomic-embed-text creates dense representations

When training a new agent:

```
Main Agent (after months of learning):
    ↓
    Has accumulated:
    - 10,000+ patterns in Neo4j
    - Successful code evolution history
    - Learned best practices (via OpenEvolve)
    - Security knowledge
    - Performance optimizations
    ↓
    User: "Train a new API agent"
    ↓
Main Agent:
    1. Query Neo4j for API-related patterns
       (authentication, rate limiting, error handling)
    
    2. Use LightRAG to find similar successful implementations
    
    3. Generate embeddings of best practices
    
    4. Create specialized training dataset
    
    5. Fine-tune new model or create specialized prompt
    
    6. New "API Agent" born with:
       - Dense API-specific knowledge
       - Patterns learned from main agent
       - Embeddings of successful code
       - Specialized for API tasks
    ↓
New API Agent:
    - Handles all API-related tasks
    - Communicates via A2A
    - Has its own tools
    - Reduces load on main agent
```

---

## Summary

### Current Architecture (Two Agents)

1. **Main Agent** (Gemma 3 27B + nomic-embed-text v1.5)
   - Orchestrator with self-awareness
   - Uses browser, files, MCP tools
   - Delegates terminal work to terminal agent
   - Stores knowledge in Neo4j + LightRAG + ChromeM
   - Creates embeddings for semantic search

2. **Terminal Agent** (comanderanch/Linux-Buster)
   - Natural language → Linux commands
   - 250+ commands trained
   - Reduces cognitive load on main agent
   - Communicates via A2A

### Future Architecture (Multi-Agent)

1. **Main Agent** - Orchestrator (trains others)
2. **Terminal Agent** - Linux specialist
3. **API Agent** - API design/implementation (trained by main agent)
4. **Database Agent** - Database operations (trained by main agent)
5. **Security Agent** - Security analysis (trained by main agent)
6. **...more as needed**

### Key Points

- ✅ **All models via Ollama**: No API costs, full privacy
- ✅ **Specific models**: gemma3:27b, nomic-embed-text:v1.5, Linux-Buster
- ✅ **Two agents now**: Main + Terminal
- ✅ **A2A for communication**: JSON-RPC 2.0
- ✅ **Each has own tools**: No overlap
- ✅ **Terminal agent offloads work**: Reduces main agent cognitive load
- ✅ **Embeddings for memory**: Semantic search via nomic-embed-text
- ✅ **Future: Main agent trains new agents**: From accumulated knowledge
- ✅ **Clean, scalable architecture**: Easy to add new agents

This is a **production-ready architecture** with all models specified and configured!

