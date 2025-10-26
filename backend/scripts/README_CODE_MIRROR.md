# Neo4j Code Mirroring System

## Overview

This system automatically mirrors your codebase into Neo4j, creating a knowledge graph that's optimized for LLM understanding. The graph represents code structure, relationships, concepts, and execution patterns in a way that makes it easy for Gemma 3 to query and reason about your code.

## Why This Schema?

### LLM-Optimized Design

The schema is designed specifically for LLM comprehension:

1. **Semantic Relationships** - Uses natural language relationship names (IMPLEMENTS_CONCEPT, RELATES_TO) that LLMs understand intuitively
2. **Rich Context** - Stores documentation, descriptions, and examples alongside code
3. **Concept Nodes** - Abstract concepts (patterns, algorithms) are first-class entities
4. **Execution History** - Links code to actual executions for learning
5. **Conversation Links** - Connects user conversations to relevant code

### Node Types

```
Project → File → Function/Class → Variable
                ↓
              Concept ← Pattern
                ↓
            Execution ← Conversation
```

## Schema Components

### Core Nodes

| Node Type | Purpose | Key Properties |
|-----------|---------|----------------|
| `Project` | Top-level container | name, description, language, version |
| `File` | Source code files | path, content, language, lines, hash |
| `Function` | Functions/methods | signature, parameters, return_type, documentation, complexity |
| `Class` | Classes/structs/interfaces | fully_qualified_name, type, documentation |
| `Variable` | Variables/constants/fields | name, type, scope, value |
| `Import` | Dependencies | module, alias, items |
| `Concept` | Abstract concepts | name, category, description, examples |
| `Pattern` | Design patterns | signature, pattern_type, use_cases, frequency |
| `Execution` | Runtime executions | execution_id, command, status, result |
| `Conversation` | User-agent chats | user_message, agent_response, context |

### Key Relationships

| Relationship | From → To | Meaning |
|--------------|-----------|---------|
| `CONTAINS_FILE` | Project → File | Project contains file |
| `DEFINES_FUNCTION` | File → Function | File defines function |
| `CALLS` | Function → Function | Function calls another |
| `IMPLEMENTS_CONCEPT` | Function → Concept | Function implements concept |
| `USES_PATTERN` | Function → Pattern | Function uses pattern |
| `INHERITS_FROM` | Class → Class | Inheritance |
| `DEPENDS_ON` | Function → Import | Dependency |
| `EXECUTED` | Execution → Function | Execution ran function |
| `DISCUSSES_CONCEPT` | Conversation → Concept | Conversation about concept |
| `TRIGGERED_EXECUTION` | Conversation → Execution | User command triggered execution |

## Usage

### 1. Initialize Schema

Run the Cypher schema first to create constraints and indexes:

```bash
cat backend/scripts/neo4j_code_mirror.cypher | cypher-shell -u neo4j -p your_password
```

Or in Neo4j Browser:
```cypher
// Copy and paste the contents of neo4j_code_mirror.cypher
```

### 2. Run Code Mirror

Mirror your codebase into Neo4j:

```bash
cd backend/scripts
export NEO4J_PASSWORD="your_password"
go run mirror_code_to_neo4j.go
```

This will:
- Parse all `.go` files in `backend/`
- Extract functions, classes, imports, variables
- Create nodes and relationships in Neo4j
- Use `MERGE` to prevent duplicates

### 3. Query from LLM

The agent can now query the knowledge graph:

```go
// Example: Find functions implementing authentication
query := `
  MATCH (fn:Function)-[:IMPLEMENTS_CONCEPT]->(con:Concept {name: "Authentication"})
  RETURN fn.name, fn.signature, fn.documentation
`

// Example: Find call chain
query := `
  MATCH path = (fn1:Function {name: "ExecuteCommand"})-[:CALLS*]->(fn2:Function)
  RETURN path
  LIMIT 10
`

// Example: Find patterns in project
query := `
  MATCH (p:Project {name: "agent-workspace"})-[:CONTAINS_FILE]->(f:File)
        -[:DEFINES_FUNCTION]->(fn:Function)-[:USES_PATTERN]->(pat:Pattern)
  RETURN pat.pattern_type, pat.description, count(fn) as usage_count
  ORDER BY usage_count DESC
`
```

## LLM Query Examples

### 1. "Show me all authentication code"

```cypher
MATCH (fn:Function)-[:IMPLEMENTS_CONCEPT]->(con:Concept)
WHERE con.name CONTAINS "Auth" OR con.name CONTAINS "Security"
RETURN fn.name, fn.signature, fn.documentation
```

### 2. "What functions does ExecuteCommand call?"

```cypher
MATCH (fn:Function {name: "ExecuteCommand"})-[:CALLS]->(called:Function)
RETURN called.name, called.signature
```

### 3. "Find all WebSocket handlers"

```cypher
MATCH (fn:Function)-[:IMPLEMENTS_CONCEPT]->(con:Concept {name: "WebSocket"})
RETURN fn.name, fn.file_path, fn.documentation
```

### 4. "Show me the most complex functions"

```cypher
MATCH (fn:Function)
WHERE fn.complexity IS NOT NULL
RETURN fn.name, fn.complexity, fn.lines
ORDER BY fn.complexity DESC
LIMIT 10
```

### 5. "What did we discuss about memory systems?"

```cypher
MATCH (conv:Conversation)-[:DISCUSSES_CONCEPT]->(con:Concept)
WHERE con.name CONTAINS "Memory"
RETURN conv.user_message, conv.agent_response, conv.timestamp
ORDER BY conv.timestamp DESC
```

### 6. "Find files that import LightRAG"

```cypher
MATCH (f:File)-[:HAS_IMPORT]->(i:Import)
WHERE i.module CONTAINS "light-rag"
RETURN f.path, f.name
```

### 7. "Show me design patterns used"

```cypher
MATCH (pat:Pattern)<-[:USES_PATTERN]-(fn:Function)
RETURN pat.pattern_type, pat.description, count(fn) as usage_count
ORDER BY usage_count DESC
```

## Automatic Mirroring

### On File Save (Watch Mode)

```go
// Add to your backend
watcher := fsnotify.NewWatcher()
watcher.Add("./backend")

go func() {
    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                if strings.HasSuffix(event.Name, ".go") {
                    mirror.MirrorFile(ctx, event.Name)
                }
            }
        }
    }
}()
```

### On Git Commit (Hook)

Create `.git/hooks/post-commit`:

```bash
#!/bin/bash
cd backend/scripts
export NEO4J_PASSWORD="your_password"
go run mirror_code_to_neo4j.go
```

```bash
chmod +x .git/hooks/post-commit
```

## Integration with Agent

### In Agent Controller

```go
// Query code knowledge
func (c *Controller) QueryCodeKnowledge(query string) (string, error) {
    ctx := context.Background()
    
    // Use LLM to convert natural language to Cypher
    cypherQuery, err := c.gemma.GenerateCypherQuery(ctx, query)
    if err != nil {
        return "", err
    }
    
    // Execute Cypher query
    result, err := c.longTermMem.ExecuteCypher(ctx, cypherQuery)
    if err != nil {
        return "", err
    }
    
    return result, nil
}
```

### In Gemma Client

```go
// Generate Cypher from natural language
func (g *GemmaClient) GenerateCypherQuery(ctx context.Context, question string) (string, error) {
    prompt := fmt.Sprintf(`Convert this question to a Cypher query for Neo4j:

Question: %s

Schema:
- Nodes: Project, File, Function, Class, Variable, Concept, Pattern, Execution
- Relationships: CONTAINS_FILE, DEFINES_FUNCTION, CALLS, IMPLEMENTS_CONCEPT, etc.

Return only the Cypher query, nothing else.`, question)

    return g.GenerateResponse(ctx, []models.Message{{
        Role: "user",
        Parts: []models.MessagePart{{Type: "text", Text: prompt}},
    }}, 0.3)
}
```

## Concept Detection

The system can automatically detect concepts:

```go
// Detect concepts in function
func detectConcepts(fn *ast.FuncDecl, doc string) []string {
    concepts := []string{}
    
    // From function name
    if strings.Contains(fn.Name.Name, "Auth") {
        concepts = append(concepts, "Authentication")
    }
    if strings.Contains(fn.Name.Name, "WebSocket") {
        concepts = append(concepts, "WebSocket")
    }
    
    // From documentation
    if strings.Contains(doc, "database") {
        concepts = append(concepts, "Database")
    }
    if strings.Contains(doc, "cache") {
        concepts = append(concepts, "Caching")
    }
    
    return concepts
}
```

## Benefits for LLM

1. **Contextual Understanding** - LLM can see how code relates to concepts
2. **Pattern Recognition** - Identify reusable patterns across codebase
3. **Impact Analysis** - Understand what changes when modifying code
4. **Learning from Execution** - Link successful/failed executions to code
5. **Conversation Context** - Remember what was discussed about each file
6. **Dependency Tracking** - Understand module relationships
7. **Code Generation** - Generate code following existing patterns

## Performance

- **Indexes** - Full-text and property indexes for fast queries
- **MERGE** - Idempotent operations, safe to run multiple times
- **Batch Processing** - Process files in parallel (future enhancement)
- **Incremental Updates** - Only update changed files

## Maintenance

### Clear All Data

```cypher
MATCH (n) DETACH DELETE n;
```

### Clear Project Data

```cypher
MATCH (p:Project {name: "agent-workspace"})-[r*]->(n)
DETACH DELETE p, n;
```

### Update Single File

```bash
go run mirror_code_to_neo4j.go --file backend/internal/agent/controller.go
```

## Future Enhancements

- [ ] Support for multiple languages (JavaScript, Python, etc.)
- [ ] Automatic concept extraction using LLM
- [ ] Pattern detection using graph algorithms
- [ ] Code smell detection
- [ ] Dependency vulnerability tracking
- [ ] Performance metrics integration
- [ ] Test coverage visualization

## License

MIT

