# Code Mirror Design
## Mirroring Codebase in Neo4j for Self-Analysis

## Overview

The **Code Mirror** is a complete representation of the agent's codebase stored in Neo4j as a knowledge graph, combined with vector embeddings for semantic search. This enables the agent to:

1. **Understand structure** - "How does everything work?"
2. **Analyze dependencies** - "What calls this function?"
3. **Find similar patterns** - "Show me similar code for optimization"
4. **Self-improve** - "This function is inefficient - refactor it"

---

## Architecture

```
Codebase (Go files)
    ↓
Code Parser (AST)
    ↓
LightRAG.Insert()
    ├─→ Neo4j (Code Graph)
    │   - Modules, Functions, Variables
    │   - Relationships: CALLS, DEPENDS_ON, PART_OF
    │   - UUID on every node
    ├─→ ChromaDB (Code Embeddings)
    │   - Semantic chunks
    │   - Linked by UUID
    └─→ BoltDB (Original Code)
        - Full source code
        - Linked by UUID
```

---

## Node Types in Neo4j

### 1. Module/File Node

**Label**: `Module`

**Properties**:
```cypher
{
  uuid: "mod-abc-123",           // Generated UUID
  name: "agent_controller.go",   // File name
  path: "/backend/internal/",    // File path
  language: "Go",                // Programming language
  version: 1,                    // Version for tracking evolution
  summary: "Handles agent lifecycle and task execution",  // LLM-generated
  embeddingUuid: "mod-abc-123"   // Links to vector store
}
```

**Relationships**:
- `IMPORTS` → Other modules
- `CONTAINS` → Functions/Structs

---

### 2. Struct Node

**Label**: `Struct`

**Properties**:
```cypher
{
  uuid: "struct-def-456",
  name: "AgentController",
  content: "type AgentController struct { ... }",
  fields: ["llmClient", "lightRAG", "mcpClient"],
  docstring: "Main controller for agent operations",
  embeddingUuid: "struct-def-456"
}
```

**Relationships**:
- `PART_OF` → Module
- `HAS_METHOD` → Functions
- `IMPLEMENTS` → Interface

---

### 3. Function/Method Node (Core Unit)

**Label**: `Function`

**Properties**:
```cypher
{
  uuid: "func-ghi-789",
  name: "HandleTask",
  content: "func (a *AgentController) HandleTask(...) { ... }",
  parameters: ["ctx context.Context", "task Task"],
  returnType: "error",
  docstring: "Processes incoming tasks with PRAR loop",
  receiver: "AgentController",  // For methods
  embeddingUuid: "func-ghi-789",
  complexity: 15,  // Cyclomatic complexity
  lines: 45
}
```

**Relationships**:
- `PART_OF` → Struct/Module
- `CALLS` → Other Functions
- `DEPENDS_ON` → External packages
- `SIMILAR_TO` → Other Functions (based on embeddings)
- `EVOLVED_FROM` → Previous version
- `USES` → Variables

---

### 4. Variable/Constant Node

**Label**: `Variable`

**Properties**:
```cypher
{
  uuid: "var-jkl-012",
  name: "maxRetries",
  type: "int",
  value: "3",  // If constant
  scope: "function"  // function, package, global
}
```

**Relationships**:
- `USED_IN` → Function
- `DEFINED_IN` → Module

---

### 5. CodeChunk Node (For Vector Linking)

**Label**: `CodeChunk`

**Properties**:
```cypher
{
  uuid: "chunk-mno-345",
  content: "// Small semantic chunk (100-500 tokens)",
  startLine: 45,
  endLine: 65,
  embeddingUuid: "chunk-mno-345"
}
```

**Relationships**:
- `EXTRACTED_FROM` → Function/Module

---

## Relationships

### Structural Relationships

```cypher
// Module imports
(m1:Module)-[:IMPORTS]->(m2:Module)

// Module contains
(m:Module)-[:CONTAINS]->(f:Function)
(m:Module)-[:CONTAINS]->(s:Struct)

// Struct methods
(s:Struct)-[:HAS_METHOD]->(f:Function)

// Function calls
(f1:Function)-[:CALLS]->(f2:Function)

// Dependencies
(f:Function)-[:DEPENDS_ON]->(pkg:Package)

// Variable usage
(v:Variable)-[:USED_IN]->(f:Function)
```

### Semantic Relationships

```cypher
// Similar code (auto-generated via embeddings)
(f1:Function)-[:SIMILAR_TO {similarity: 0.92}]->(f2:Function)

// Evolution history
(f_new:Function)-[:EVOLVED_FROM]->(f_old:Function)

// Pattern relationships
(f:Function)-[:IMPLEMENTS_PATTERN]->(p:Pattern {name: "Caching"})
```

---

## Vector Storage

### Chunking Strategy

**Per Function**:
```
Function: HandleTask (45 lines)
  ↓
Chunk 1: Lines 1-20 (function signature + setup)
Chunk 2: Lines 15-35 (main logic, 20% overlap)
Chunk 3: Lines 30-45 (cleanup + return)
```

**Storage in ChromaDB**:
```python
{
  "id": "func-ghi-789",  // UUID from Neo4j
  "embedding": [0.123, 0.456, ...],  // 768 dims from nomic-embed-text
  "metadata": {
    "nodeType": "Function",
    "name": "HandleTask",
    "summary": "Processes tasks with PRAR loop",
    "neo4jUuid": "func-ghi-789",
    "module": "agent_controller.go",
    "language": "Go"
  }
}
```

### Hybrid Search

**Semantic Query** (Vector):
```python
# Find similar code
query = "function that handles caching"
embedding = embed(query)
results = chromadb.query(embedding, top_k=10)
uuids = [r.metadata["neo4jUuid"] for r in results]
```

**Structural Query** (Graph):
```cypher
// Get full context from Neo4j
MATCH (f:Function {uuid: $uuid})-[:CALLS*]->(dep)
RETURN f, dep
```

**Combined**:
1. Vector search finds semantically similar functions
2. Graph traversal gets full call chain and dependencies
3. Agent has complete context for reasoning

---

## Implementation

### 1. Code Parser (Go AST)

```go
package codemirror

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type CodeParser struct {
	lightRAG *lightrag.Client
}

func (p *CodeParser) ParseFile(filePath string) error {
	// Parse Go file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Extract module info
	moduleUUID := generateUUID()
	moduleName := filepath.Base(filePath)
	
	// Insert module into LightRAG
	_, err = p.lightRAG.InsertPerception(
		ctx,
		moduleUUID,
		fmt.Sprintf("Module: %s\nPath: %s", moduleName, filePath),
		map[string]interface{}{
			"type": "module",
			"name": moduleName,
			"path": filePath,
		},
	)

	// Walk AST and extract functions
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			p.extractFunction(x, moduleUUID)
		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				p.extractStruct(x, structType, moduleUUID)
			}
		}
		return true
	})

	return nil
}

func (p *CodeParser) extractFunction(fn *ast.FuncDecl, moduleUUID string) {
	funcUUID := generateUUID()
	
	// Extract function details
	name := fn.Name.Name
	content := getFunctionSource(fn)
	params := extractParameters(fn)
	returnType := extractReturnType(fn)
	docstring := extractDocstring(fn)

	// Insert into LightRAG
	p.lightRAG.InsertPerception(
		ctx,
		funcUUID,
		fmt.Sprintf("Function: %s\n%s\n\nCode:\n%s", name, docstring, content),
		map[string]interface{}{
			"type":       "function",
			"name":       name,
			"parameters": params,
			"returnType": returnType,
			"moduleUUID": moduleUUID,
		},
	)
}
```

### 2. Cypher Queries for Analysis

**Find all functions that call a specific function**:
```cypher
MATCH (f:Function {name: 'HandleTask'})<-[:CALLS]-(caller)
RETURN caller.name, caller.uuid
```

**Find dependency chain**:
```cypher
MATCH path = (f:Function {name: 'HandleTask'})-[:CALLS*]->(dep)
RETURN path
```

**Find similar functions**:
```cypher
MATCH (f:Function {name: 'HandleTask'})-[:SIMILAR_TO]->(similar)
WHERE similar.uuid <> f.uuid
RETURN similar.name, similar.content
```

**Find evolution history**:
```cypher
MATCH path = (f:Function {name: 'HandleTask'})-[:EVOLVED_FROM*]->(old)
RETURN path
ORDER BY old.version DESC
```

### 3. Self-Improvement Flow

```
1. Agent analyzes code
   ↓
   Query: "Find inefficient functions"
   ↓
   Vector search: Functions with high complexity
   ↓
   Graph traversal: Get call chains

2. Agent reasons
   ↓
   Query: "Find similar optimized code"
   ↓
   Vector search: Similar functions with better performance
   ↓
   Extract patterns from optimized code

3. Agent proposes changes
   ↓
   Generate refactored code
   ↓
   Store as new version with EVOLVED_FROM relationship

4. Agent reflects
   ↓
   Compare performance: old vs new
   ↓
   Extract learnings
   ↓
   Create pattern: "High complexity → Refactor with pattern X"
```

---

## Integration with PRAR Loop

### PERCEIVE

```go
// Agent perceives its own code
perception := "The HandleTask function has high cyclomatic complexity (15) and makes 5 database calls without caching."

// Query code mirror
result := lightRAG.Query("HandleTask function structure")

// Get full context from Neo4j
cypher := `
  MATCH (f:Function {name: 'HandleTask'})-[:CALLS]->(dep)
  RETURN f, dep
`
```

### REASON

```go
// Query for similar optimized code
result := lightRAG.Query("optimized functions with caching")

// Generate reasoning branches
branches := []string{
  "Add caching layer (based on similar pattern in UserController)",
  "Reduce database calls (based on similar pattern in OrderProcessor)",
  "Implement connection pooling (based on similar pattern in DataManager)",
}

// Select best branch based on similarity to successful patterns
```

### ACT

```go
// Generate refactored code
newCode := generateRefactoredCode(selectedBranch)

// Store as new version
newUUID := generateUUID()
lightRAG.InsertAction(
  ctx,
  newUUID,
  "Refactor HandleTask with caching",
  newCode,
  reasoningUUID,
)

// Create evolution relationship in Neo4j
cypher := `
  MATCH (old:Function {name: 'HandleTask', version: 1})
  MATCH (new:Function {uuid: $newUUID})
  MERGE (new)-[:EVOLVED_FROM]->(old)
  SET new.version = old.version + 1
`
```

### REFLECT

```go
// Compare performance
oldPerf := measurePerformance(oldCode)
newPerf := measurePerformance(newCode)
improvement := (oldPerf - newPerf) / oldPerf * 100

// Extract pattern
pattern := "High complexity + multiple DB calls → Add caching layer"

// Store pattern in Neo4j
cypher := `
  MERGE (p:Pattern {name: $patternName})
  SET p.description = $description,
      p.successRate = $successRate,
      p.avgImprovement = $avgImprovement
  WITH p
  MATCH (f:Function {uuid: $funcUUID})
  MERGE (f)-[:IMPLEMENTS_PATTERN]->(p)
`
```

---

## Benefits

### 1. Structural Understanding

Agent can query:
- "What functions call this?"
- "What are the dependencies?"
- "Show me the call chain"

### 2. Semantic Understanding

Agent can query:
- "Find similar code patterns"
- "Show me optimized versions"
- "What code handles caching?"

### 3. Self-Improvement

Agent can:
- Identify inefficient code
- Find better patterns from its own history
- Propose refactorings based on proven solutions
- Track evolution over time

### 4. Knowledge Accumulation

Every improvement:
- Creates new version (EVOLVED_FROM)
- Extracts patterns (IMPLEMENTS_PATTERN)
- Links to performance metrics
- Enables learning from experience

---

## Tools and Optimizations

### APOC (Neo4j Plugin)

```cypher
// Generate UUIDs
CREATE (n:Function)
SET n.uuid = apoc.create.uuid()

// Similarity algorithms
MATCH (f1:Function), (f2:Function)
WHERE f1.uuid <> f2.uuid
WITH f1, f2, apoc.ml.cosine(f1.embedding, f2.embedding) AS similarity
WHERE similarity > 0.85
MERGE (f1)-[:SIMILAR_TO {similarity: similarity}]->(f2)
```

### Indexing

```cypher
// Speed up UUID lookups
CREATE INDEX function_uuid FOR (f:Function) ON (f.uuid)

// Speed up name searches
CREATE INDEX function_name FOR (f:Function) ON (f.name)

// Full-text search
CREATE FULLTEXT INDEX function_content FOR (f:Function) ON EACH [f.content, f.docstring]
```

### Versioning

```cypher
// Track versions
(:Function {name: 'HandleTask', version: 1})-[:EVOLVED_TO]->
(:Function {name: 'HandleTask', version: 2})-[:EVOLVED_TO]->
(:Function {name: 'HandleTask', version: 3})

// Get latest version
MATCH (f:Function {name: 'HandleTask'})
WHERE NOT (f)-[:EVOLVED_TO]->()
RETURN f
```

---

## Implementation Phases

### Phase 1: Basic Code Mirror (1 week)

1. Go AST parser
2. Extract modules, structs, functions
3. Insert into LightRAG
4. Basic Cypher queries

### Phase 2: Semantic Search (3 days)

1. Chunk code into semantic units
2. Generate embeddings
3. Link to Neo4j via UUID
4. Hybrid search (vector + graph)

### Phase 3: Self-Improvement (1 week)

1. Identify inefficient code
2. Query for similar optimized code
3. Generate refactorings
4. Track evolution

### Phase 4: Pattern Learning (3 days)

1. Extract patterns from improvements
2. Store in Neo4j
3. Apply patterns to new code
4. Measure effectiveness

---

## Summary

The **Code Mirror** enables true self-awareness:

✅ **Structural understanding** - Graph traversal  
✅ **Semantic understanding** - Vector search  
✅ **Self-improvement** - Evolution tracking  
✅ **Pattern learning** - Knowledge accumulation  
✅ **UUID linking** - Perfect consistency  

Combined with PRAR loop and LightRAG, the agent can:
- Analyze its own code
- Find inefficiencies
- Propose improvements based on proven patterns
- Track evolution over time
- Learn from experience

**This is true self-improving code!** 🧬

