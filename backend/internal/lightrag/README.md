# LightRAG Integration

Complete LightRAG integration for the self-evolving agent using [go-light-rag](https://github.com/MegaGrindStone/go-light-rag).

## Overview

LightRAG is the **sovereign orchestrator** for all agent data. Everything flows through LightRAG before being stored in Neo4j and ChromaDB.

```
Agent Data → LightRAG → Neo4j (graph) + ChromaDB (vectors) + BoltDB (key-value)
```

## Architecture

### Storage Layers

1. **Neo4j** (Graph Database)
   - Stores entities and relationships
   - Enables graph traversal
   - Discovers connections between concepts

2. **ChromaDB** (Vector Database)
   - Stores embeddings via nomic-embed-text:v1.5
   - Enables semantic search
   - 768-dimensional vectors

3. **BoltDB** (Key-Value Store)
   - Stores original document chunks
   - Fast retrieval by UUID
   - Embedded database (no separate server)

### LLM Integration

- **Gemma 3 (27B)** via Ollama for text generation
- **nomic-embed-text:v1.5** via Ollama for embeddings

## Installation

### 1. Install Dependencies

```bash
cd backend
go get github.com/MegaGrindStone/go-light-rag
go get github.com/philippgille/chromem-go
```

### 2. Start Neo4j

```bash
# Using Docker
docker run \
    --name neo4j \
    -p 7474:7474 -p 7687:7687 \
    -e NEO4J_AUTH=neo4j/your_password \
    neo4j:latest

# Or install locally
# https://neo4j.com/download/
```

### 3. Start Ollama

```bash
# Pull models
ollama pull gemma3:27b
ollama pull nomic-embed-text:v1.5

# Ollama runs on http://localhost:11434 by default
```

### 4. Set Environment Variables

```bash
export NEO4J_PASSWORD="your_password"
export OLLAMA_BASE_URL="http://localhost:11434"
```

## Usage

### Initialize Client

```go
package main

import (
    "context"
    "log"
    "os"
    
    "github.com/docpainting/agentic_selfevolve_commandcenter/backend/internal/lightrag"
)

func main() {
    ctx := context.Background()
    logger := log.New(os.Stdout, "[LightRAG] ", log.LstdFlags)

    cfg := &lightrag.Config{
        Neo4jURI:      "bolt://localhost:7687",
        Neo4jUser:     "neo4j",
        Neo4jPassword: os.Getenv("NEO4J_PASSWORD"),
        VectorDBPath:  "./data/vector.db",
        VectorDBSize:  10,
        KVDBPath:      "./data/kv.db",
        OllamaBaseURL: "http://localhost:11434",
        LLMModel:      "gemma3:27b",
        EmbedModel:    "nomic-embed-text:v1.5",
        HandlerType:   "default",
        Logger:        logger,
    }

    client, err := lightrag.NewClient(ctx, cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Use client...
}
```

### PRAR Loop Integration

#### 1. Perceive

```go
perceptionUUID, err := client.InsertPerception(
    ctx,
    "perception-001",
    "The HandleTask function is slow due to multiple database queries without caching.",
    map[string]interface{}{
        "type":       "perception",
        "confidence": 0.85,
        "timestamp":  time.Now(),
    },
)
```

**What LightRAG Does**:
- Generates UUID: `perception-001`
- Extracts entities: `HandleTask`, `database`, `caching`
- Creates embeddings via nomic-embed-text
- Stores in Neo4j: `(:Perception {uuid: "perception-001", ...})`
- Stores vectors in ChromaDB with same UUID
- Stores original content in BoltDB

#### 2. Reason

```go
reasoningUUID, err := client.InsertReasoning(
    ctx,
    "reasoning-001",
    []string{
        "Add caching layer to reduce database calls",
        "Optimize database queries with indexes",
        "Implement connection pooling",
    },
    "Add caching layer to reduce database calls", // Selected branch
    perceptionUUID, // Links to perception
)
```

**What LightRAG Does**:
- Generates UUID: `reasoning-001`
- Extracts entities from branches
- Creates relationship: `(reasoning-001)-[:BASED_ON]->(perception-001)`
- Stores embeddings for each branch
- Enables semantic search of past reasoning

#### 3. Act

```go
actionUUID, err := client.InsertAction(
    ctx,
    "action-001",
    "Implement Redis caching for database queries",
    "Successfully added caching. Performance improved by 40%.",
    reasoningUUID,
)
```

**What LightRAG Does**:
- Generates UUID: `action-001`
- Creates relationship: `(action-001)-[:BASED_ON]->(reasoning-001)`
- Stores execution trace
- Links performance metrics to action

#### 4. Reflect

```go
reflectionUUID, err := client.InsertReflection(
    ctx,
    "reflection-001",
    []string{
        "Caching significantly improves performance for read-heavy operations",
        "Redis is effective for caching database query results",
    },
    []string{
        "Pattern: Database bottleneck → Add caching layer",
        "Pattern: Read-heavy operations → Use Redis cache",
    },
    actionUUID,
)
```

**What LightRAG Does**:
- Generates UUID: `reflection-001`
- Creates pattern entities in Neo4j
- Links patterns to action: `(pattern)-[:LEARNED_FROM]->(action-001)`
- Enables pattern retrieval for future tasks

### Query for Similar Situations

```go
result, err := client.Query(ctx, "How can I improve database performance?")

// LightRAG returns:
// - Local entities (directly related)
// - Global entities (broader context)
// - Local sources (most relevant documents)
// - Global sources (related documents)

for _, source := range result.LocalSources {
    fmt.Printf("Relevance: %.2f\nContent: %s\n", 
        source.Relevance, source.Content)
}
```

**What LightRAG Does**:
- Embeds query with nomic-embed-text
- Searches ChromaDB for similar vectors
- Traverses Neo4j graph for related entities
- Combines vector similarity + graph relationships
- Returns ranked results

### Retrieve by UUID

```go
source, err := client.QueryByUUID(ctx, "perception-001")
fmt.Println(source.Content)
```

## The Complete Flow

### Example: Optimize HandleTask Function

```
1. PERCEIVE
   ↓
   LightRAG.InsertPerception("HandleTask is slow...")
   ↓
   Neo4j: (:Perception {uuid: "p-001", content: "..."})
   ChromaDB: {uuid: "p-001", embedding: [0.123, ...]}
   BoltDB: {p-001: "HandleTask is slow..."}
   ↓
   Returns: "p-001"

2. REASON
   ↓
   LightRAG.Query("Similar performance issues?")
   ↓
   Finds: Previous caching solutions
   ↓
   LightRAG.InsertReasoning(branches, selected, "p-001")
   ↓
   Neo4j: (:Reasoning {uuid: "r-001"})-[:BASED_ON]->(:Perception {uuid: "p-001"})
   ↓
   Returns: "r-001"

3. ACT
   ↓
   Execute: Add Redis caching
   ↓
   LightRAG.InsertAction(plan, result, "r-001")
   ↓
   Neo4j: (:Action {uuid: "a-001"})-[:BASED_ON]->(:Reasoning {uuid: "r-001"})
   ↓
   Returns: "a-001"

4. REFLECT
   ↓
   Extract learnings and patterns
   ↓
   LightRAG.InsertReflection(learnings, patterns, "a-001")
   ↓
   Neo4j: 
     (:Reflection {uuid: "ref-001"})-[:BASED_ON]->(:Action {uuid: "a-001"})
     (:Pattern {name: "DB bottleneck → Caching"})-[:LEARNED_FROM]->(:Action {uuid: "a-001"})
   ↓
   Returns: "ref-001"

5. FUTURE TASK
   ↓
   LightRAG.Query("Database is slow")
   ↓
   Finds pattern: "DB bottleneck → Caching"
   ↓
   Agent applies learned solution!
```

## Benefits

### 1. Automatic UUID Generation
- No manual UUID management
- Consistent across all storage layers

### 2. Automatic Entity Extraction
- LightRAG identifies entities from text
- Creates graph relationships automatically

### 3. Automatic Embeddings
- All content embedded via nomic-embed-text
- Semantic search enabled by default

### 4. Hybrid Retrieval
- Vector similarity (ChromaDB)
- Graph traversal (Neo4j)
- Best of both worlds!

### 5. Knowledge Accumulation
- Every PRAR loop adds to knowledge graph
- Patterns emerge over time
- Agent learns from experience

## Handler Types

### Default Handler
```go
cfg.HandlerType = "default"
```
- General-purpose text processing
- Follows Python LightRAG implementation
- Good for most use cases

### Semantic Handler
```go
cfg.HandlerType = "semantic"
```
- Creates semantically meaningful chunks
- Uses LLM to identify natural boundaries
- Better quality, more LLM calls

### Go Handler
```go
cfg.HandlerType = "go"
```
- Specialized for Go source code
- AST parsing for logical sections
- Perfect for code mirror

## Configuration

### Chunking Parameters

```go
// In handler.Default{}
ChunkSize:    1200  // Tokens per chunk
ChunkOverlap: 100   // Overlap between chunks
```

### Concurrency

```go
// Number of concurrent operations
Concurrency: 4  // Balance speed vs resources
```

### Vector DB Size

```go
VectorDBSize: 10  // Return top 10 results
```

## Troubleshooting

### Neo4j Connection Failed

```bash
# Check Neo4j is running
curl http://localhost:7474

# Check credentials
export NEO4J_PASSWORD="correct_password"
```

### Ollama Not Found

```bash
# Check Ollama is running
curl http://localhost:11434/api/tags

# Pull models
ollama pull gemma3:27b
ollama pull nomic-embed-text:v1.5
```

### Embeddings Not Working

```bash
# Test embeddings
curl http://localhost:11434/api/embeddings -d '{
  "model": "nomic-embed-text:v1.5",
  "prompt": "test"
}'
```

## Integration with MCP Servers

### Dynamic Thinking MCP Server

```python
# In perceive tool
result = lightrag_client.insert_perception(
    id=generate_uuid(),
    content=perception_data,
    metadata={"confidence": 0.85}
)
```

### OpenEvolve MCP Server

```python
# Store code evolution iterations
result = lightrag_client.insert_action(
    id=f"evolution-{iteration}",
    plan=f"Evolve code iteration {iteration}",
    result=evolved_code,
    reasoning_uuid=reasoning_id
)
```

## Summary

LightRAG is the **foundation** of the self-evolving agent:

✅ **Sovereign orchestrator** - All data flows through it  
✅ **Automatic UUID generation** - No manual management  
✅ **Automatic entity extraction** - Builds knowledge graph  
✅ **Automatic embeddings** - Semantic search enabled  
✅ **Hybrid retrieval** - Vector + graph combined  
✅ **Knowledge accumulation** - Learns from experience  

**Everything depends on LightRAG - it's the foundation!** 🎯

