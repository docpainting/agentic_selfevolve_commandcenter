# Go-LightRAG Integration Guide

## Overview

Based on the **actual go-light-rag library**, this guide shows how to integrate LightRAG into the agent architecture. The library provides two core functions (`Insert` and `Query`) with flexible interfaces for storage, LLM, and document handling.

## Key Insight from Library Documentation

> "go-light-rag is a Go library that implements the core components of the LightRAG architecture rather than providing an end-to-end system."

This means:
- ✅ We have **full control** over document insertion workflows
- ✅ We get **direct access** to retrieved context data
- ✅ We can **craft custom prompts** for our use case
- ✅ We can **integrate with existing Go applications**

## Architecture Components

### 1. Language Models (LLM Interface)

```go
type LLM interface {
    Chat(messages []Message) (string, error)
}
```

**Implementations available:**
- OpenAI (GPT models)
- Anthropic (Claude models)
- **Ollama** (self-hosted, open-source) ← **We use this!**
- OpenRouter (unified access)

**Our implementation:**
```go
// Use Ollama with Gemma 3
llm := llm.NewOllama(
    "http://localhost:11434",
    "gemma3:27b",
    params,
    logger,
)
```

### 2. Storage Interfaces

#### GraphStorage (Neo4j)
```go
type GraphStorage interface {
    // Manages entity and relationship data
}

// Initialize Neo4j
graphDB, err := storage.NewNeo4J(
    "bolt://localhost:7687",
    "neo4j",
    "password",
)
```

#### VectorStorage (ChromeM or Milvus)
```go
type VectorStorage interface {
    // Provides semantic search capabilities
}

// Option 1: ChromeM (in-memory, persisted)
embeddingFunc := storage.EmbeddingFunc(
    chromem.NewEmbeddingFuncOllama(
        "http://localhost:11434",
        "nomic-embed-text:v1.5",
    ),
)

vecDB, err := storage.NewChromem(
    "vec.db",
    5,  // Top-K results
    embeddingFunc,
)

// Option 2: Milvus (scalable vector DB)
// vectorDim := 768  // nomic-embed-text dimension
// milvusCfg := &milvusclient.ClientConfig{
//     Address: "localhost:19530",
//     DBName:  "agent_knowledge",
// }
// vecDB, err := storage.NewMilvus(milvusCfg, 5, vectorDim, embeddingFunc)
```

#### KeyValueStorage (BoltDB or Redis)
```go
type KeyValueStorage interface {
    // Stores original document chunks
}

// Option 1: BoltDB (embedded, file-based)
kvDB, err := storage.NewBolt("kv.db")

// Option 2: Redis (networked, scalable)
// kvDB, err := storage.NewRedis("localhost:6379", "", 0)
```

### 3. Handlers

**Available handlers:**
- **Default**: General-purpose text (matches Python implementation)
- **Semantic**: Advanced chunking using LLM for natural boundaries
- **Go**: Specialized for Go source code (AST parsing)

**Our use case:**
```go
// For Go source code (agent's own code)
goHandler := handler.Go{
    ChunkSize:    1500,
    ChunkOverlap: 100,
    EntityTypes:  []string{"function", "type", "struct", "interface", "method"},
}

// For general documents (retrieved info, transcripts, etc.)
semanticHandler := handler.Semantic{
    TokenThreshold: 4096,
    MaxChunkSize:   2000,
    LLM:            llm,  // Uses Gemma 3 for semantic chunking
}

// For simple text (default)
defaultHandler := handler.Default{}
```

## Integration with Agent Architecture

### Storage Wrapper

```go
// Create unified storage wrapper
type AgentStorage struct {
    Graph  storage.GraphStorage
    Vector storage.VectorStorage
    KV     storage.KeyValueStorage
}

func NewAgentStorage() (*AgentStorage, error) {
    // 1. Initialize Neo4j
    graphDB, err := storage.NewNeo4J(
        os.Getenv("NEO4J_URI"),
        os.Getenv("NEO4J_USER"),
        os.Getenv("NEO4J_PASSWORD"),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to init Neo4j: %w", err)
    }
    
    // 2. Initialize ChromeM with Ollama embeddings
    embeddingFunc := storage.EmbeddingFunc(
        chromem.NewEmbeddingFuncOllama(
            "http://localhost:11434",
            "nomic-embed-text:v1.5",
        ),
    )
    
    vecDB, err := storage.NewChromem(
        "/var/lib/agent/vectors.db",
        10,  // Top-K
        embeddingFunc,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to init ChromeM: %w", err)
    }
    
    // 3. Initialize BoltDB
    kvDB, err := storage.NewBolt("/var/lib/agent/kv.db")
    if err != nil {
        return nil, fmt.Errorf("failed to init BoltDB: %w", err)
    }
    
    return &AgentStorage{
        Graph:  graphDB,
        Vector: vecDB,
        KV:     kvDB,
    }, nil
}
```

### LLM Initialization

```go
func NewAgentLLM() (golightrag.LLM, error) {
    // Use Ollama with Gemma 3
    return llm.NewOllama(
        "http://localhost:11434",
        "gemma3:27b",
        map[string]interface{}{
            "temperature": 0.7,
            "top_p":       0.9,
            "num_ctx":     8192,  // Context window
        },
        logger,
    )
}
```

## PRAR Loop Integration

### PERCEIVE Phase

```go
func (p *PerceiveTool) Execute(params map[string]interface{}) (*PerceptionResult, error) {
    taskID := params["task_id"].(string)
    
    // 1. Analyze environment (existing logic)
    perception := p.analyzeEnvironment(params)
    
    // 2. Create document from perception
    doc := golightrag.Document{
        ID: fmt.Sprintf("perception_%s", taskID),
        Content: fmt.Sprintf(`
Task: %s
Goal: %s
Observations: %s
Related Entities: %s
Confidence: %f
`, perception.Task, perception.Goal, 
   strings.Join(perception.Observations, "\n"),
   formatEntities(perception.RelatedEntities),
   perception.Confidence),
    }
    
    // 3. Insert into LightRAG
    err := golightrag.Insert(
        doc,
        p.semanticHandler,  // Use semantic handler for better chunking
        p.storage,
        p.llm,
        p.logger,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to insert perception: %w", err)
    }
    
    // LightRAG automatically:
    // - Extracts entities (functions, patterns, variables)
    // - Generates UUIDs for each entity
    // - Creates embeddings using nomic-embed-text
    // - Stores vectors in ChromeM
    // - Stores graph in Neo4j
    // - Creates relationships
    
    perception.DocumentID = doc.ID
    
    return perception, nil
}
```

### REASON Phase

```go
func (r *ReasonTool) Execute(params map[string]interface{}) (*ReasoningResult, error) {
    taskID := params["task_id"].(string)
    perceptionDocID := params["perception_doc_id"].(string)
    
    // 1. Query LightRAG for similar past reasoning
    conversation := []golightrag.QueryConversation{
        {
            Role: golightrag.RoleUser,
            Message: fmt.Sprintf(`
Find similar past reasoning for:
Task: %s
Context: %s
`, params["task"], params["context"]),
        },
    }
    
    result, err := golightrag.Query(
        conversation,
        r.semanticHandler,
        r.storage,
        r.llm,
        r.logger,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to query similar reasoning: %w", err)
    }
    
    // 2. Extract insights from query result
    log.Info("Found similar reasoning",
             "local_entities", len(result.LocalEntities),
             "global_entities", len(result.GlobalEntities),
             "sources", len(result.LocalSources))
    
    // 3. Generate reasoning branches WITH context from LightRAG
    branches := r.generateBranches(result)
    
    // 4. Create document for reasoning
    selectedBranch := r.selectBestBranch(branches)
    
    reasoningDoc := golightrag.Document{
        ID: fmt.Sprintf("reasoning_%s", taskID),
        Content: fmt.Sprintf(`
Reasoning for: %s
Selected Strategy: %s
Confidence: %f
Feasibility: %f
Alignment: %f
Risk: %f

Branches Considered:
%s

Past Similar Cases:
%s
`, params["task"],
   selectedBranch.Strategy,
   selectedBranch.Confidence,
   selectedBranch.Feasibility,
   selectedBranch.Alignment,
   selectedBranch.Risk,
   formatBranches(branches),
   formatSources(result.LocalSources)),
    }
    
    // 5. Insert reasoning into LightRAG
    err = golightrag.Insert(
        reasoningDoc,
        r.semanticHandler,
        r.storage,
        r.llm,
        r.logger,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to insert reasoning: %w", err)
    }
    
    // LightRAG automatically:
    // - Extracts strategy entities
    // - Creates UUIDs for theories
    // - Creates embeddings for each branch
    // - Links to perception entities
    // - Stores in Neo4j + ChromeM
    
    return &ReasoningResult{
        Branches:       branches,
        SelectedBranch: selectedBranch,
        DocumentID:     reasoningDoc.ID,
        SimilarCases:   result.LocalSources,
    }, nil
}
```

### ACT Phase

```go
func (a *ActTool) Execute(params map[string]interface{}) (*ExecutionResult, error) {
    taskID := params["task_id"].(string)
    
    // 1. Query LightRAG for complete process context
    conversation := []golightrag.QueryConversation{
        {
            Role: golightrag.RoleUser,
            Message: fmt.Sprintf(`
Retrieve complete context for task: %s
Include perception, reasoning, and similar past executions.
`, taskID),
        },
    }
    
    result, err := golightrag.Query(
        conversation,
        a.semanticHandler,
        a.storage,
        a.llm,
        a.logger,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to query context: %w", err)
    }
    
    // 2. Execute with full context
    execution := a.executeWithContext(result)
    
    // 3. Create execution document
    executionDoc := golightrag.Document{
        ID: fmt.Sprintf("execution_%s", taskID),
        Content: fmt.Sprintf(`
Execution for: %s
Status: %s
Results: %s
Performance Gain: %f
Confidence: %f

Context Used:
%s
`, taskID,
   execution.Status,
   execution.Results,
   execution.PerformanceGain,
   execution.Confidence,
   formatContext(result)),
    }
    
    // 4. Insert execution into LightRAG
    err = golightrag.Insert(
        executionDoc,
        a.semanticHandler,
        a.storage,
        a.llm,
        a.logger,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to insert execution: %w", err)
    }
    
    return execution, nil
}
```

### REFLECT Phase

```go
func (r *ReflectTool) Execute(params map[string]interface{}) (*ReflectionResult, error) {
    taskID := params["task_id"].(string)
    
    // 1. Query LightRAG for ENTIRE process
    conversation := []golightrag.QueryConversation{
        {
            Role: golightrag.RoleUser,
            Message: fmt.Sprintf(`
Retrieve complete process for task: %s
Include all phases: perception, reasoning, execution.
Find similar complete processes for comparison.
`, taskID),
        },
    }
    
    result, err := golightrag.Query(
        conversation,
        r.semanticHandler,
        r.storage,
        r.llm,
        r.logger,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to query complete process: %w", err)
    }
    
    log.Info("Retrieved complete process",
             "local_entities", len(result.LocalEntities),
             "global_entities", len(result.GlobalEntities),
             "similar_processes", len(result.LocalSources))
    
    // 2. Reflect on process
    reflection := r.reflectOnProcess(result)
    
    // 3. Extract learned patterns
    patterns := r.extractPatterns(reflection, result)
    
    // 4. Create reflection document
    reflectionDoc := golightrag.Document{
        ID: fmt.Sprintf("reflection_%s", taskID),
        Content: fmt.Sprintf(`
Reflection for: %s
Success: %t
Performance Gain: %f
Confidence: %f

Learnings:
%s

Patterns Discovered:
%s

Similar Processes:
%s

Recommendations:
%s
`, taskID,
   reflection.Success,
   reflection.PerformanceGain,
   reflection.Confidence,
   strings.Join(reflection.Learnings, "\n"),
   formatPatterns(patterns),
   formatSources(result.LocalSources),
   strings.Join(reflection.Recommendations, "\n")),
    }
    
    // 5. Insert reflection into LightRAG
    err = golightrag.Insert(
        reflectionDoc,
        r.semanticHandler,
        r.storage,
        r.llm,
        r.logger,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to insert reflection: %w", err)
    }
    
    // LightRAG automatically:
    // - Creates holistic process embedding
    // - Extracts pattern entities
    // - Creates UUID for each pattern
    // - Links patterns to all phases
    // - Stores in Neo4j + ChromeM
    
    return &ReflectionResult{
        Success:         reflection.Success,
        Learnings:       reflection.Learnings,
        Patterns:        patterns,
        DocumentID:      reflectionDoc.ID,
        SimilarProcesses: result.LocalSources,
    }, nil
}
```

## Custom Handler for Agent Code

```go
// Custom handler for agent's own Go code
type AgentCodeHandler struct {
    handler.Go  // Embed Go handler
}

func (h *AgentCodeHandler) ExtractEntities(chunk string) []string {
    // Use Go handler's AST parsing
    entities := h.Go.ExtractEntities(chunk)
    
    // Add agent-specific entity types
    agentEntities := []string{
        "mcp_tool",
        "prar_phase",
        "pattern",
        "strategy",
        "evaluation",
    }
    
    return append(entities, agentEntities...)
}

// Use custom handler for code insertion
codeHandler := &AgentCodeHandler{
    Go: handler.Go{
        ChunkSize:    1500,
        ChunkOverlap: 100,
        EntityTypes:  []string{"function", "type", "struct", "interface", "method"},
    },
}

// Insert agent's own code
codeDoc := golightrag.Document{
    ID:      "agent_code_main.go",
    Content: agentSourceCode,
}

err := golightrag.Insert(codeDoc, codeHandler, storage, llm, logger)
```

## Query Result Structure

```go
type QueryResult struct {
    LocalEntities  []Entity
    GlobalEntities []Entity
    LocalSources   []Source
    GlobalSources  []Source
}

type Source struct {
    ID        string
    Relevance float64
    Content   string
}

// Example usage
result, _ := golightrag.Query(conversation, handler, storage, llm, logger)

// Access retrieved context
for _, entity := range result.LocalEntities {
    fmt.Printf("Entity: %s (relevance: %.2f)\n", entity.Name, entity.Relevance)
}

for _, source := range result.LocalSources {
    fmt.Printf("Source: %s\nRelevance: %.2f\nContent: %s\n\n",
        source.ID, source.Relevance, source.Content)
}

// Or use formatted output
fmt.Println(result.String())
```

## Configuration

```yaml
# config/lightrag.yaml
lightrag:
  llm:
    provider: ollama
    base_url: http://localhost:11434
    model: gemma3:27b
    params:
      temperature: 0.7
      top_p: 0.9
      num_ctx: 8192
  
  embedding:
    provider: ollama
    base_url: http://localhost:11434
    model: nomic-embed-text:v1.5
    dimension: 768
  
  storage:
    graph:
      type: neo4j
      uri: bolt://localhost:7687
      user: neo4j
      password: ${NEO4J_PASSWORD}
    
    vector:
      type: chromem  # or milvus
      path: /var/lib/agent/vectors.db
      top_k: 10
    
    kv:
      type: boltdb  # or redis
      path: /var/lib/agent/kv.db
  
  handlers:
    default:
      chunk_size: 1200
      chunk_overlap: 100
    
    semantic:
      token_threshold: 4096
      max_chunk_size: 2000
    
    go_code:
      chunk_size: 1500
      chunk_overlap: 100
      entity_types:
        - function
        - type
        - struct
        - interface
        - method
        - mcp_tool
        - prar_phase
```

## Benefits of go-light-rag

### 1. **Native Go Integration**
- No Python subprocess needed
- Direct integration with agent code
- Type-safe interfaces

### 2. **Flexible Storage**
- Choose ChromeM (embedded) or Milvus (scalable)
- Choose BoltDB (embedded) or Redis (networked)
- Neo4j for graph (standard)

### 3. **Custom Handlers**
- Default for general text
- Semantic for better chunking
- Go for source code
- Custom for agent-specific content

### 4. **Direct Access to Results**
- Raw entities and sources
- No predefined prompts
- Full control over prompt engineering

### 5. **Ollama Support**
- Self-hosted LLM (Gemma 3)
- Self-hosted embeddings (nomic-embed-text)
- No external API dependencies

## Implementation Checklist

### Phase 1: Setup
- [ ] Install go-light-rag: `go get github.com/your-org/go-light-rag`
- [ ] Initialize Neo4j connection
- [ ] Initialize ChromeM with Ollama embeddings
- [ ] Initialize BoltDB for key-value storage
- [ ] Create storage wrapper

### Phase 2: LLM Integration
- [ ] Initialize Ollama LLM client
- [ ] Configure Gemma 3 parameters
- [ ] Test LLM connectivity

### Phase 3: Handler Configuration
- [ ] Configure semantic handler for general documents
- [ ] Configure Go handler for source code
- [ ] Create custom handler for agent-specific content

### Phase 4: PRAR Integration
- [ ] Update PERCEIVE to insert into LightRAG
- [ ] Update REASON to query + insert
- [ ] Update ACT to query + insert
- [ ] Update REFLECT to query + insert patterns

### Phase 5: Testing
- [ ] Test document insertion
- [ ] Test query retrieval
- [ ] Test entity extraction
- [ ] Test relationship discovery
- [ ] Test cross-task learning

## Summary

### Key Takeaways

1. **go-light-rag provides two core functions:**
   - `Insert`: Add documents with entity extraction
   - `Query`: Retrieve contextually relevant information

2. **Storage is modular:**
   - Graph: Neo4j
   - Vector: ChromeM or Milvus
   - KV: BoltDB or Redis

3. **Handlers are customizable:**
   - Default, Semantic, Go
   - Custom handlers for agent-specific content

4. **LLM integration is flexible:**
   - Ollama for self-hosted (Gemma 3)
   - OpenAI, Anthropic, OpenRouter also supported

5. **Perfect for agent architecture:**
   - Native Go (no Python bridge)
   - Full control over prompts
   - Direct access to results
   - Integrates with existing code

This is the **practical implementation** of LightRAG in the agent! 🎯

