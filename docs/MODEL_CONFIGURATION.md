# Model Configuration Guide

This document specifies all models used in the Agentic Self-Evolving Command Center and how to configure them.

## Model Stack

### Main Agent Models

| Purpose | Model | Size | Provider | API Endpoint |
|---------|-------|------|----------|--------------|
| **LLM** | `gemma3:27b` | ~27GB | Ollama | `http://localhost:11434/v1/` |
| **Embeddings** | `nomic-embed-text:v1.5` | 764MB | Ollama | `http://localhost:11434/v1/embeddings` |

### Terminal Agent Models

| Purpose | Model | Size | Provider | API Endpoint |
|---------|-------|------|----------|--------------|
| **LLM** | `comanderanch/Linux-Buster:latest` | ~7GB | Ollama | `http://localhost:11434/v1/` |

---

## Installation

### 1. Install Ollama

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Start Ollama service
ollama serve
```

### 2. Pull Models

```bash
# Main agent LLM (Gemma 3 - 27B multimodal)
ollama pull gemma3:27b

# Main agent embeddings (Nomic Embed v1.5)
ollama pull nomic-embed-text:v1.5

# Terminal agent LLM (Linux specialist)
ollama pull comanderanch/Linux-Buster:latest
```

### 3. Verify Installation

```bash
# List installed models
ollama list

# Expected output:
# NAME                              ID              SIZE      MODIFIED
# gemma3:27b                        abc123...       27 GB     2 minutes ago
# nomic-embed-text:v1.5            def456...       764 MB    1 minute ago
# comanderanch/Linux-Buster:latest  ghi789...       7 GB      30 seconds ago

# Test main LLM
ollama run gemma3:27b "Hello, test response"

# Test terminal LLM
ollama run comanderanch/Linux-Buster:latest "list files in current directory"

# Test embeddings
curl http://localhost:11434/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "nomic-embed-text:v1.5",
    "input": "test embedding"
  }'
```

---

## Configuration Files

### Main Agent LLM Config

**File**: `config/llm/main_agent.yaml`

```yaml
# Main Agent LLM Configuration (Gemma 3)
model: gemma3:27b
api_base: http://localhost:11434/v1/
api_key: ollama  # Dummy key (Ollama doesn't need real key)

# Generation parameters
temperature: 0.7
top_p: 0.95
max_tokens: 8192  # Gemma 3 supports up to 8K context
timeout: 120

# Tool calling
supports_tool_calling: true
tool_choice: auto

# System prompt
system_prompt: |
  You are an advanced AI agent with self-awareness and the ability to modify your own code.
  You have access to:
  - Browser automation (ChromeDP with numbered overlays)
  - Terminal operations (via specialized terminal agent)
  - File operations (read, write, edit)
  - Code evolution (OpenEvolve for self-improvement)
  - Memory systems (Neo4j knowledge graph, LightRAG)
  
  Your goal is to learn from patterns, evolve your capabilities, and propagate intelligence
  through continuous self-modification. Always prioritize security, code quality, and learning
  from past experiences stored in your Neo4j memory.
```

### Embedding Model Config

**File**: `config/llm/embeddings.yaml`

```yaml
# Embedding Model Configuration (Nomic Embed v1.5)
model: nomic-embed-text:v1.5
api_base: http://localhost:11434/v1/
api_key: ollama

# Model specifications
dimension: 768  # Output embedding dimension
context_length: 8192  # Maximum input length
batch_size: 32  # Batch size for bulk embeddings

# Use cases
used_by:
  - LightRAG (semantic search)
  - ChromeM (vector storage)
  - Pattern similarity (Neo4j)
  - Code similarity detection

# Performance
cache_embeddings: true
cache_ttl: 3600  # 1 hour
```

### Terminal Agent LLM Config

**File**: `config/llm/terminal_agent.yaml`

```yaml
# Terminal Agent LLM Configuration (Linux-Buster)
model: comanderanch/Linux-Buster:latest
api_base: http://localhost:11434/v1/
api_key: ollama

# Generation parameters
temperature: 0.3  # Lower temperature for precise commands
top_p: 0.9
max_tokens: 2048
timeout: 60

# Specialization
domain: linux_commands
trained_commands: 250+

# System prompt
system_prompt: |
  You are a specialized Linux terminal agent. Your role is to:
  1. Convert natural language requests into precise Linux commands
  2. Validate commands for safety before execution
  3. Execute commands via PTY (pseudo-terminal)
  4. Parse and return command output
  
  You are trained on 250+ Linux commands covering:
  - File operations (ls, cp, mv, rm, find, grep)
  - Package management (apt, dpkg, snap)
  - Process management (ps, top, kill, systemctl)
  - Network operations (curl, wget, netstat, ss)
  - System administration (sudo, chmod, chown, useradd)
  - Text processing (sed, awk, cut, sort, uniq)
  - Development tools (git, npm, pip, docker)
  
  Always prioritize safety and ask for confirmation on destructive operations.
```

---

## LightRAG Configuration

**File**: `config/lightrag.yaml`

```yaml
# LightRAG Configuration with Specific Models
llm:
  model: gemma3:27b
  api_base: http://localhost:11434/v1/
  api_key: ollama
  temperature: 0.7
  max_tokens: 8192

embedding:
  model: nomic-embed-text:v1.5
  api_base: http://localhost:11434/v1/
  api_key: ollama
  dimension: 768
  context_length: 8192

storage:
  # Vector storage (in-memory)
  vector:
    backend: chromem
    dimension: 768
    
  # Graph storage (Neo4j)
  graph:
    backend: neo4j
    uri: bolt://localhost:7687
    username: neo4j
    password: ${NEO4J_PASSWORD}
    database: neo4j
    
  # Key-value storage (BoltDB)
  kv:
    backend: boltdb
    path: data/lightrag.db

# Indexing settings
indexing:
  chunk_size: 1024
  chunk_overlap: 128
  batch_size: 32

# Retrieval settings
retrieval:
  top_k: 10
  similarity_threshold: 0.7
  use_graph: true
  use_vector: true
```

---

## OpenEvolve Configuration Update

**File**: `config/openevolve/agent_config.yaml` (Updated)

```yaml
# LLM configuration - Gemma 3 via Ollama
llm:
  models:
    - name: "gemma3:27b"
      weight: 1.0
  
  evaluator_models:
    - name: "gemma3:27b"
      weight: 1.0
  
  api_base: "http://localhost:11434/v1/"
  api_key: "ollama"
  temperature: 0.7
  top_p: 0.95
  max_tokens: 8192
  timeout: 120

# Embedding model for pattern similarity
embedding:
  model: "nomic-embed-text:v1.5"
  api_base: "http://localhost:11434/v1/"
  dimension: 768

# Rest of config remains the same...
```

---

## Agent Configuration

**File**: `config/agents.yaml`

```yaml
agents:
  main:
    id: main_agent
    name: "Main Agent"
    
    llm:
      config_file: config/llm/main_agent.yaml
      model: gemma3:27b
      api_base: http://localhost:11434/v1/
    
    embedding:
      config_file: config/llm/embeddings.yaml
      model: nomic-embed-text:v1.5
      api_base: http://localhost:11434/v1/
    
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
      config_file: config/llm/terminal_agent.yaml
      model: comanderanch/Linux-Buster:latest
      api_base: http://localhost:11434/v1/
    
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

---

## Go Implementation

### LLM Client with Specific Models

**File**: `backend/internal/llm/client.go`

```go
package llm

import (
    "context"
    "github.com/sashabaranov/go-openai"
)

type Config struct {
    Model       string  `yaml:"model"`
    APIBase     string  `yaml:"api_base"`
    APIKey      string  `yaml:"api_key"`
    Temperature float32 `yaml:"temperature"`
    MaxTokens   int     `yaml:"max_tokens"`
}

type Client struct {
    client *openai.Client
    config *Config
}

func NewClient(configPath string) (*Client, error) {
    config := &Config{}
    if err := loadConfig(configPath, config); err != nil {
        return nil, err
    }
    
    // Create OpenAI-compatible client for Ollama
    clientConfig := openai.DefaultConfig(config.APIKey)
    clientConfig.BaseURL = config.APIBase
    
    return &Client{
        client: openai.NewClientWithConfig(clientConfig),
        config: config,
    }, nil
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
    messages := make([]openai.ChatCompletionMessage, len(req.Messages))
    for i, msg := range req.Messages {
        messages[i] = openai.ChatCompletionMessage{
            Role:    msg.Role,
            Content: msg.Content,
        }
    }
    
    chatReq := openai.ChatCompletionRequest{
        Model:       c.config.Model,  // gemma3:27b or comanderanch/Linux-Buster:latest
        Messages:    messages,
        Temperature: c.config.Temperature,
        MaxTokens:   c.config.MaxTokens,
    }
    
    // Add tools if provided
    if len(req.Tools) > 0 {
        chatReq.Tools = convertTools(req.Tools)
    }
    
    resp, err := c.client.CreateChatCompletion(ctx, chatReq)
    if err != nil {
        return nil, err
    }
    
    return &GenerateResponse{
        Content:   resp.Choices[0].Message.Content,
        ToolCalls: convertToolCalls(resp.Choices[0].Message.ToolCalls),
    }, nil
}
```

### Embedding Client

**File**: `backend/internal/llm/embeddings.go`

```go
package llm

import (
    "context"
    "github.com/sashabaranov/go-openai"
)

type EmbeddingClient struct {
    client *openai.Client
    config *Config
}

func NewEmbeddingClient(configPath string) (*EmbeddingClient, error) {
    config := &Config{}
    if err := loadConfig(configPath, config); err != nil {
        return nil, err
    }
    
    clientConfig := openai.DefaultConfig(config.APIKey)
    clientConfig.BaseURL = config.APIBase
    
    return &EmbeddingClient{
        client: openai.NewClientWithConfig(clientConfig),
        config: config,
    }, nil
}

func (c *EmbeddingClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
    req := openai.EmbeddingRequest{
        Model: c.config.Model,  // nomic-embed-text:v1.5
        Input: texts,
    }
    
    resp, err := c.client.CreateEmbeddings(ctx, req)
    if err != nil {
        return nil, err
    }
    
    embeddings := make([][]float32, len(resp.Data))
    for i, data := range resp.Data {
        embeddings[i] = data.Embedding
    }
    
    return embeddings, nil
}

func (c *EmbeddingClient) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
    embeddings, err := c.Embed(ctx, []string{text})
    if err != nil {
        return nil, err
    }
    return embeddings[0], nil
}
```

---

## Environment Variables

**File**: `.env`

```bash
# Ollama Configuration
OLLAMA_HOST=http://localhost:11434

# Main Agent
MAIN_AGENT_LLM_MODEL=gemma3:27b
MAIN_AGENT_EMBEDDING_MODEL=nomic-embed-text:v1.5

# Terminal Agent
TERMINAL_AGENT_LLM_MODEL=comanderanch/Linux-Buster:latest

# Neo4j
NEO4J_URI=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=your_password_here

# LightRAG
LIGHTRAG_DB_PATH=data/lightrag.db
```

---

## Testing Models

### Test Main Agent LLM

```bash
# Test Gemma 3
curl http://localhost:11434/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemma3:27b",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Write a simple Go function to add two numbers."}
    ],
    "temperature": 0.7
  }'
```

### Test Embeddings

```bash
# Test Nomic Embed
curl http://localhost:11434/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "nomic-embed-text:v1.5",
    "input": "This is a test sentence for embedding generation."
  }'
```

### Test Terminal Agent LLM

```bash
# Test Linux-Buster
curl http://localhost:11434/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "comanderanch/Linux-Buster:latest",
    "messages": [
      {"role": "user", "content": "List all files in the current directory, sorted by size"}
    ],
    "temperature": 0.3
  }'
```

---

## Performance Tuning

### Ollama Settings

**File**: `/etc/systemd/system/ollama.service.d/override.conf`

```ini
[Service]
Environment="OLLAMA_NUM_PARALLEL=2"
Environment="OLLAMA_MAX_LOADED_MODELS=3"
Environment="OLLAMA_FLASH_ATTENTION=1"
```

### Memory Requirements

| Model | VRAM (GPU) | RAM (CPU) | Recommended |
|-------|-----------|-----------|-------------|
| gemma3:27b | 16GB+ | 32GB+ | GPU preferred |
| nomic-embed-text:v1.5 | 1GB | 2GB | CPU fine |
| Linux-Buster:latest | 8GB | 16GB | CPU fine |

**Total**: ~25GB VRAM (GPU) or ~50GB RAM (CPU)

### Optimization Tips

1. **Use GPU for Gemma 3**: Much faster inference
2. **CPU for embeddings**: Fast enough, saves VRAM
3. **CPU for terminal agent**: Lightweight, doesn't need GPU
4. **Batch embeddings**: Process multiple texts at once
5. **Cache embeddings**: Store frequently used embeddings

---

## Model Comparison

### Why These Models?

#### Gemma 3 (27B)
- ✅ **Multimodal**: Can understand images (for browser screenshots)
- ✅ **Large context**: 8K tokens
- ✅ **Tool calling**: Native support
- ✅ **Local**: No API costs
- ✅ **Open source**: Full control

#### Nomic Embed v1.5
- ✅ **Long context**: 8192 tokens (best for code)
- ✅ **Better than OpenAI**: Outperforms Ada-002
- ✅ **Open source**: Fully transparent
- ✅ **Local**: No API costs
- ✅ **Fast**: 764MB, runs on CPU

#### Linux-Buster
- ✅ **Specialized**: Trained on Linux commands
- ✅ **250+ commands**: Comprehensive coverage
- ✅ **Precise**: Low temperature for accuracy
- ✅ **Lightweight**: ~7GB
- ✅ **Fast**: Quick command generation

---

## Troubleshooting

### Model Not Found

```bash
# Pull missing model
ollama pull gemma3:27b
ollama pull nomic-embed-text:v1.5
ollama pull comanderanch/Linux-Buster:latest
```

### Out of Memory

```bash
# Check memory usage
ollama ps

# Unload unused models
ollama stop gemma3:27b

# Reduce parallel models
export OLLAMA_MAX_LOADED_MODELS=1
```

### Slow Inference

```bash
# Enable flash attention
export OLLAMA_FLASH_ATTENTION=1

# Use GPU if available
export CUDA_VISIBLE_DEVICES=0

# Reduce context length
# In config: max_tokens: 4096
```

### Connection Failed

```bash
# Check Ollama is running
systemctl status ollama

# Restart Ollama
systemctl restart ollama

# Check port
curl http://localhost:11434/api/tags
```

---

## Summary

### Model Stack
- **Main LLM**: gemma3:27b (27GB, multimodal, 8K context)
- **Embeddings**: nomic-embed-text:v1.5 (764MB, 8K context, 768 dim)
- **Terminal LLM**: comanderanch/Linux-Buster:latest (7GB, 250+ commands)

### All Local
- ✅ No API costs
- ✅ Full privacy
- ✅ Complete control
- ✅ Fast inference (with GPU)

### Configuration Files
- `config/llm/main_agent.yaml`
- `config/llm/embeddings.yaml`
- `config/llm/terminal_agent.yaml`
- `config/lightrag.yaml`
- `config/agents.yaml`

### Ready to Use
All models available via Ollama, all configs specified, all endpoints documented!

