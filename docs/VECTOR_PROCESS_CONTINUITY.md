# Vector-Based Process Continuity

## Overview

During long-running PRAR loops (especially with graph-aware branching and online retrieval), the agent needs to **maintain process continuity** by storing its entire process state as vectors and retrieving them before moving forward. This ensures the agent doesn't lose context and can resume intelligently after each phase.

## The Problem

**Without vector-based continuity:**
```
PERCEIVE → Result stored in memory
    ↓
REASON → Starts fresh, no context from perception
    ↓
ACT → Starts fresh, no context from reasoning
    ↓
REFLECT → Starts fresh, no context from actions
    ↓
Result: Disjointed process, lost insights, repeated work
```

**With vector-based continuity:**
```
PERCEIVE → Store as vector + structured data
    ↓
REASON → Retrieve perception vector, build on it
    ↓ Store reasoning as vector
ACT → Retrieve perception + reasoning vectors
    ↓ Store actions as vector
REFLECT → Retrieve all vectors (perception + reasoning + actions)
    ↓ Store reflection as vector
    ↓
Result: Continuous process, cumulative learning, no lost context
```

## Vector Storage Architecture

### What Gets Vectorized

```go
type ProcessState struct {
    TaskID string
    Phase string  // "perceive", "reason", "act", "reflect"
    
    // Structured data
    StructuredData interface{}
    
    // Vector embeddings
    Embeddings struct {
        Summary []float64      // Dense summary of this phase
        KeyPoints []float64    // Important insights
        Context []float64      // Full context for retrieval
    }
    
    // Metadata for retrieval
    Metadata struct {
        Timestamp time.Time
        Confidence float64
        RelatedEntities []string
        Tags []string
    }
}
```

### Storage Layers

```
┌─────────────────────────────────────────────────────────────┐
│  Short-Term Process Memory (ChromeM)                        │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  Current Task Vectors                                 │ │
│  │  • Perception embeddings                              │ │
│  │  • Reasoning branch embeddings                        │ │
│  │  • Action result embeddings                           │ │
│  │  • Intermediate state embeddings                      │ │
│  │  Lifetime: Duration of task (cleared after reflect)   │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                          ↓ After reflection
┌─────────────────────────────────────────────────────────────┐
│  Long-Term Process Memory (Neo4j + LightRAG)                │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  Historical Process Vectors                           │ │
│  │  • Successful process patterns                        │ │
│  │  • Failed attempts (for learning)                     │ │
│  │  • Cross-task insights                                │ │
│  │  • Evolution history                                  │ │
│  │  Lifetime: Permanent                                  │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Process Flow with Vector Continuity

### Phase 1: PERCEIVE

```go
func (p *PerceiveTool) Execute(params map[string]interface{}) (*PerceptionResult, error) {
    taskID := params["task_id"].(string)
    
    // 1. Perform perception
    perception := p.analyzeEnvironment(params)
    
    // 2. Create vector embeddings
    embeddings := p.createEmbeddings(perception)
    
    // 3. Store in ChromeM (short-term)
    processState := &ProcessState{
        TaskID: taskID,
        Phase: "perceive",
        StructuredData: perception,
        Embeddings: embeddings,
        Metadata: Metadata{
            Timestamp: time.Now(),
            Confidence: perception.Confidence,
            Tags: []string{"perception", "environment_analysis"},
        },
    }
    
    p.chromem.Store(taskID, "perceive", processState)
    
    // 4. Return result with vector ID for retrieval
    perception.VectorID = processState.ID
    
    return perception, nil
}

func (p *PerceiveTool) createEmbeddings(perception *Perception) Embeddings {
    // Create summary embedding
    summaryText := fmt.Sprintf(`
Perception Summary:
- Confidence: %f
- Primary Entity: %s
- Related Entities: %d
- Key Observations: %s
`, perception.Confidence, perception.PrimaryEntity.ID, 
   len(perception.RelatedEntities), perception.KeyObservations)
    
    summaryEmbedding := embeddings.Embed(summaryText)
    
    // Create key points embedding
    keyPointsText := strings.Join(perception.KeyObservations, "\n")
    keyPointsEmbedding := embeddings.Embed(keyPointsText)
    
    // Create full context embedding
    contextText := perception.FullContext()
    contextEmbedding := embeddings.Embed(contextText)
    
    return Embeddings{
        Summary: summaryEmbedding,
        KeyPoints: keyPointsEmbedding,
        Context: contextEmbedding,
    }
}
```

### Phase 2: REASON (with Vector Retrieval)

```go
func (r *ReasonTool) Execute(params map[string]interface{}) (*ReasoningResult, error) {
    taskID := params["task_id"].(string)
    
    // 1. RETRIEVE perception vectors from ChromeM
    perceptionState, err := r.chromem.Retrieve(taskID, "perceive")
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve perception: %w", err)
    }
    
    perception := perceptionState.StructuredData.(*Perception)
    
    log.Info("Retrieved perception from vectors",
             "confidence", perception.Confidence,
             "key_observations", len(perception.KeyObservations))
    
    // 2. Query LightRAG for similar past reasoning
    similarReasoning := r.lightRAG.QuerySimilar(
        perceptionState.Embeddings.Summary,
        topK: 5,
        filter: map[string]interface{}{
            "phase": "reason",
            "success": true,
        },
    )
    
    log.Info("Found similar past reasoning",
             "count", len(similarReasoning))
    
    // 3. Generate reasoning branches WITH context from perception
    branches := r.generateBranches(perception, similarReasoning)
    
    // 4. Create embeddings for each branch
    for i := range branches {
        branchEmbedding := r.createBranchEmbedding(&branches[i])
        branches[i].Embedding = branchEmbedding
    }
    
    // 5. Select best branch
    selectedBranch := r.selectBestBranch(branches)
    
    // 6. Create reasoning state with embeddings
    reasoningEmbeddings := r.createReasoningEmbeddings(branches, selectedBranch)
    
    reasoningState := &ProcessState{
        TaskID: taskID,
        Phase: "reason",
        StructuredData: &ReasoningData{
            Branches: branches,
            SelectedBranch: selectedBranch,
            PerceptionVectorID: perceptionState.ID,  // Link to perception
        },
        Embeddings: reasoningEmbeddings,
        Metadata: Metadata{
            Timestamp: time.Now(),
            Confidence: selectedBranch.Confidence,
            Tags: []string{"reasoning", "multi_branch"},
        },
    }
    
    // 7. Store in ChromeM
    r.chromem.Store(taskID, "reason", reasoningState)
    
    return &ReasoningResult{
        Branches: branches,
        SelectedBranch: selectedBranch,
        VectorID: reasoningState.ID,
    }, nil
}
```

### Phase 3: ACT (with Full Process Retrieval)

```go
func (a *ActTool) Execute(params map[string]interface{}) (*ExecutionResult, error) {
    taskID := params["task_id"].(string)
    
    // 1. RETRIEVE entire process history from ChromeM
    processHistory, err := a.chromem.RetrieveAll(taskID)
    if err != nil {
        return nil, err
    }
    
    perception := processHistory["perceive"].StructuredData.(*Perception)
    reasoning := processHistory["reason"].StructuredData.(*ReasoningData)
    
    log.Info("Retrieved full process history",
             "perception_confidence", perception.Confidence,
             "reasoning_branches", len(reasoning.Branches))
    
    // 2. Query LightRAG for similar successful executions
    // Use combined embedding from perception + reasoning
    combinedEmbedding := a.combineEmbeddings(
        processHistory["perceive"].Embeddings.Summary,
        processHistory["reason"].Embeddings.Summary,
    )
    
    similarExecutions := a.lightRAG.QuerySimilar(
        combinedEmbedding,
        topK: 5,
        filter: map[string]interface{}{
            "phase": "act",
            "status": "success",
        },
    )
    
    // 3. Execute action plan WITH full context
    execution := a.executeWithContext(
        reasoning.SelectedBranch.ActionPlan,
        perception,
        reasoning,
        similarExecutions,
    )
    
    // 4. Create execution embeddings
    executionEmbeddings := a.createExecutionEmbeddings(execution)
    
    executionState := &ProcessState{
        TaskID: taskID,
        Phase: "act",
        StructuredData: execution,
        Embeddings: executionEmbeddings,
        Metadata: Metadata{
            Timestamp: time.Now(),
            Confidence: execution.Confidence,
            Tags: []string{"execution", execution.Status},
        },
    }
    
    // 5. Store in ChromeM
    a.chromem.Store(taskID, "act", executionState)
    
    return execution, nil
}
```

### Phase 4: REFLECT (with Complete Process Vector Retrieval)

```go
func (r *ReflectTool) Execute(params map[string]interface{}) (*ReflectionResult, error) {
    taskID := params["task_id"].(string)
    
    // 1. RETRIEVE ENTIRE process from ChromeM
    processHistory, err := r.chromem.RetrieveAll(taskID)
    if err != nil {
        return nil, err
    }
    
    perception := processHistory["perceive"].StructuredData.(*Perception)
    reasoning := processHistory["reason"].StructuredData.(*ReasoningData)
    execution := processHistory["act"].StructuredData.(*Execution)
    
    log.Info("Retrieved complete process for reflection",
             "phases", len(processHistory))
    
    // 2. Create holistic process embedding
    holisticEmbedding := r.createHolisticEmbedding(
        processHistory["perceive"].Embeddings.Summary,
        processHistory["reason"].Embeddings.Summary,
        processHistory["act"].Embeddings.Summary,
    )
    
    // 3. Query LightRAG for similar complete processes
    similarProcesses := r.lightRAG.QuerySimilar(
        holisticEmbedding,
        topK: 10,
        filter: map[string]interface{}{
            "phase": "reflect",
            "complete_process": true,
        },
    )
    
    log.Info("Found similar complete processes",
             "count", len(similarProcesses))
    
    // 4. Reflect on ENTIRE process
    reflection := r.reflectOnProcess(
        perception,
        reasoning,
        execution,
        similarProcesses,
    )
    
    // 5. Create reflection embeddings
    reflectionEmbeddings := r.createReflectionEmbeddings(reflection)
    
    reflectionState := &ProcessState{
        TaskID: taskID,
        Phase: "reflect",
        StructuredData: reflection,
        Embeddings: reflectionEmbeddings,
        Metadata: Metadata{
            Timestamp: time.Now(),
            Confidence: reflection.Confidence,
            Tags: []string{"reflection", "complete_process"},
        },
    }
    
    // 6. Store in ChromeM (temporary)
    r.chromem.Store(taskID, "reflect", reflectionState)
    
    // 7. PERSIST entire process to Neo4j + LightRAG (permanent)
    r.persistProcessToLongTerm(taskID, processHistory, reflectionState)
    
    // 8. Clear ChromeM for this task (process complete)
    r.chromem.Clear(taskID)
    
    return reflection, nil
}
```

## Embedding Strategy

### Multi-Level Embeddings

```go
type Embeddings struct {
    // Level 1: Summary (768 dims)
    // Quick retrieval, high-level matching
    Summary []float64
    
    // Level 2: Key Points (768 dims)
    // Important insights, decision points
    KeyPoints []float64
    
    // Level 3: Full Context (768 dims)
    // Complete information for deep analysis
    Context []float64
}

func createEmbeddings(data interface{}) Embeddings {
    // Use nomic-embed-text:v1.5 (768 dimensions)
    
    // Summary: Concise representation
    summaryText := extractSummary(data)
    summaryEmbedding := embedModel.Embed(summaryText)
    
    // Key Points: Important details
    keyPointsText := extractKeyPoints(data)
    keyPointsEmbedding := embedModel.Embed(keyPointsText)
    
    // Full Context: Everything
    contextText := extractFullContext(data)
    contextEmbedding := embedModel.Embed(contextText)
    
    return Embeddings{
        Summary: summaryEmbedding,
        KeyPoints: keyPointsEmbedding,
        Context: contextEmbedding,
    }
}
```

### Combining Embeddings Across Phases

```go
func combineEmbeddings(embeddings ...[]float64) []float64 {
    // Method 1: Average (simple, works well)
    combined := make([]float64, len(embeddings[0]))
    for _, emb := range embeddings {
        for i, val := range emb {
            combined[i] += val
        }
    }
    for i := range combined {
        combined[i] /= float64(len(embeddings))
    }
    
    // Normalize
    return normalize(combined)
}

func createHolisticEmbedding(perceiveEmb, reasonEmb, actEmb []float64) []float64 {
    // Weighted combination based on importance
    weights := map[string]float64{
        "perceive": 0.2,  // Context
        "reason": 0.5,    // Strategy (most important)
        "act": 0.3,       // Results
    }
    
    combined := make([]float64, len(perceiveEmb))
    
    for i := range combined {
        combined[i] = 
            perceiveEmb[i] * weights["perceive"] +
            reasonEmb[i] * weights["reason"] +
            actEmb[i] * weights["act"]
    }
    
    return normalize(combined)
}
```

## ChromeM Integration (Short-Term)

### Storage

```go
type ChromeMStore struct {
    collections map[string]*chromem.Collection
}

func (c *ChromeMStore) Store(taskID, phase string, state *ProcessState) error {
    collectionName := fmt.Sprintf("task_%s", taskID)
    
    // Get or create collection
    collection, exists := c.collections[collectionName]
    if !exists {
        collection = chromem.NewCollection(collectionName, nil, nil)
        c.collections[collectionName] = collection
    }
    
    // Store with all three embeddings
    docID := fmt.Sprintf("%s_%s", taskID, phase)
    
    // Primary storage: Summary embedding
    collection.Add(chromem.Document{
        ID: docID,
        Embedding: state.Embeddings.Summary,
        Metadata: map[string]interface{}{
            "task_id": taskID,
            "phase": phase,
            "confidence": state.Metadata.Confidence,
            "timestamp": state.Metadata.Timestamp,
            "tags": state.Metadata.Tags,
        },
        Content: serializeStructuredData(state.StructuredData),
    })
    
    // Also store key points and context for different retrieval strategies
    collection.Add(chromem.Document{
        ID: docID + "_keypoints",
        Embedding: state.Embeddings.KeyPoints,
        Metadata: map[string]interface{}{
            "task_id": taskID,
            "phase": phase,
            "type": "keypoints",
        },
        Content: serializeStructuredData(state.StructuredData),
    })
    
    return nil
}
```

### Retrieval

```go
func (c *ChromeMStore) Retrieve(taskID, phase string) (*ProcessState, error) {
    collectionName := fmt.Sprintf("task_%s", taskID)
    collection := c.collections[collectionName]
    
    docID := fmt.Sprintf("%s_%s", taskID, phase)
    
    // Retrieve document
    docs, err := collection.Query(nil, 1, map[string]interface{}{
        "id": docID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    if len(docs) == 0 {
        return nil, fmt.Errorf("no process state found for %s/%s", taskID, phase)
    }
    
    // Deserialize
    state := &ProcessState{
        TaskID: taskID,
        Phase: phase,
        StructuredData: deserializeStructuredData(docs[0].Content),
        Embeddings: Embeddings{
            Summary: docs[0].Embedding,
        },
        Metadata: Metadata{
            Confidence: docs[0].Metadata["confidence"].(float64),
            Timestamp: docs[0].Metadata["timestamp"].(time.Time),
            Tags: docs[0].Metadata["tags"].([]string),
        },
    }
    
    return state, nil
}

func (c *ChromeMStore) RetrieveAll(taskID string) (map[string]*ProcessState, error) {
    phases := []string{"perceive", "reason", "act"}
    states := make(map[string]*ProcessState)
    
    for _, phase := range phases {
        state, err := c.Retrieve(taskID, phase)
        if err != nil {
            log.Warn("Phase not found", "phase", phase, "error", err)
            continue
        }
        states[phase] = state
    }
    
    return states, nil
}
```

## LightRAG Integration (Long-Term)

### Persisting Complete Process

```go
func (r *ReflectTool) persistProcessToLongTerm(
    taskID string,
    processHistory map[string]*ProcessState,
    reflection *ProcessState,
) error {
    // 1. Create holistic embedding
    holisticEmbedding := r.createHolisticEmbedding(
        processHistory["perceive"].Embeddings.Summary,
        processHistory["reason"].Embeddings.Summary,
        processHistory["act"].Embeddings.Summary,
    )
    
    // 2. Create process document
    processDoc := &ProcessDocument{
        TaskID: taskID,
        Goal: processHistory["perceive"].StructuredData.(*Perception).Goal,
        Phases: map[string]interface{}{
            "perceive": processHistory["perceive"].StructuredData,
            "reason": processHistory["reason"].StructuredData,
            "act": processHistory["act"].StructuredData,
            "reflect": reflection.StructuredData,
        },
        HolisticEmbedding: holisticEmbedding,
        Success: reflection.StructuredData.(*Reflection).Success,
        PerformanceGain: reflection.StructuredData.(*Reflection).PerformanceGain,
        Timestamp: time.Now(),
    }
    
    // 3. Store in LightRAG
    err := r.lightRAG.Insert(processDoc)
    if err != nil {
        return err
    }
    
    // 4. Store in Neo4j with relationships
    err = r.storeProcessInNeo4j(processDoc)
    if err != nil {
        return err
    }
    
    log.Info("Process persisted to long-term memory",
             "task_id", taskID,
             "success", processDoc.Success)
    
    return nil
}
```

### Querying Similar Processes

```go
func (r *ReasonTool) querySimilarProcesses(currentEmbedding []float64) ([]*ProcessDocument, error) {
    // Query LightRAG for similar complete processes
    results := r.lightRAG.QuerySimilar(
        currentEmbedding,
        topK: 10,
        filter: map[string]interface{}{
            "success": true,
            "performance_gain": map[string]interface{}{
                "$gt": 0.2,  // At least 20% improvement
            },
        },
    )
    
    processes := []*ProcessDocument{}
    for _, result := range results {
        process := result.Document.(*ProcessDocument)
        processes = append(processes, process)
    }
    
    return processes, nil
}
```

## Benefits of Vector-Based Continuity

### 1. **No Lost Context**
```
Traditional: Each phase starts fresh
Vector-based: Each phase builds on previous embeddings
```

### 2. **Semantic Retrieval**
```
Traditional: Retrieve by ID or timestamp
Vector-based: Retrieve by semantic similarity
    ↓
"Find processes similar to current perception"
→ Returns relevant past experiences, not just recent ones
```

### 3. **Cumulative Learning**
```
Perception → Embedding stored
    ↓
Reasoning → Retrieves perception embedding + similar past reasoning
    ↓
Action → Retrieves perception + reasoning embeddings + similar executions
    ↓
Reflection → Retrieves ALL embeddings + similar complete processes
    ↓
Each phase has MORE context than the last
```

### 4. **Cross-Task Learning**
```
Task A: Optimize HandleTask
    → Process stored as vectors in LightRAG
    
Task B: Optimize HandleRequest (similar!)
    → Retrieves Task A's process vectors
    → Learns from Task A's approach
    → Applies similar strategy
```

### 5. **Efficient Storage**
```
Structured data: Large (JSON, full transcripts, etc.)
Vector embeddings: Small (768 floats = 3KB)
    ↓
ChromeM: Stores vectors for fast retrieval
Neo4j: Stores structured data for analysis
LightRAG: Stores both for semantic search
```

## Configuration

```yaml
# config/vector_continuity.yaml
vector_storage:
  short_term:
    backend: chromem
    embedding_model: nomic-embed-text:v1.5
    embedding_dimension: 768
    lifetime: task_duration  # Cleared after reflection
    
  long_term:
    backend: lightrag
    embedding_model: nomic-embed-text:v1.5
    embedding_dimension: 768
    neo4j_integration: true
    lifetime: permanent

embeddings:
  levels:
    - summary      # Quick retrieval
    - key_points   # Important insights
    - full_context # Complete information
  
  combination_strategy: weighted_average
  weights:
    perceive: 0.2
    reason: 0.5
    act: 0.3

retrieval:
  similarity_threshold: 0.7
  top_k: 10
  filters:
    - success: true
    - confidence: ">= 0.7"
    - performance_gain: ">= 0.2"
```

## Summary

### Vector-Based Process Continuity

**What it is:**
- Store entire PRAR process as vector embeddings
- Retrieve process state before each phase
- Build cumulative context across phases
- Persist successful processes for future learning

**How it works:**
1. **PERCEIVE** → Create embeddings → Store in ChromeM
2. **REASON** → Retrieve perception vectors → Add reasoning vectors
3. **ACT** → Retrieve perception + reasoning vectors → Add execution vectors
4. **REFLECT** → Retrieve ALL vectors → Create holistic embedding → Persist to LightRAG + Neo4j

**Benefits:**
- ✅ No lost context between phases
- ✅ Semantic retrieval of similar processes
- ✅ Cumulative learning (each phase has more context)
- ✅ Cross-task learning (apply learnings from past tasks)
- ✅ Efficient storage (vectors are small)

**Storage:**
- **ChromeM** (short-term): Current task vectors, cleared after reflection
- **LightRAG** (long-term): Historical process vectors, permanent
- **Neo4j** (structured): Full process data with relationships

The agent now maintains **perfect continuity** across long-running processes with vector-based state management! 🧠

