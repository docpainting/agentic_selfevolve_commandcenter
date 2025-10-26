# Graph-Aware Branching in Dynamic Thinking

## Overview

This enhancement enables the agent to **discover and process related entities** from the Neo4j knowledge graph during the PRAR loop. When working on one entity (function, file, strategy), the agent can identify similar or related nodes and branch to reflect on them in parallel.

## The Problem

**Current approach:**
```
User: "Optimize HandleTask function"
    ↓
Agent reasons about HandleTask in isolation
    ↓
Misses that HandleRequest has similar pattern
    ↓
Doesn't apply learning to related functions
```

**Better approach:**
```
User: "Optimize HandleTask function"
    ↓
Agent queries Neo4j: "What's related to HandleTask?"
    ↓
Finds: HandleRequest, HandleEvent (similar patterns)
    ↓
Branches to reflect on all three in parallel
    ↓
Applies optimization to entire family of functions
```

## Architecture

### Enhanced PRAR Loop with Graph Discovery

```
┌─────────────────────────────────────────────────────────────┐
│  PERCEIVE                                                   │
│  ┌───────────────────────────────────────────────────────┐ │
│  │ 1. Analyze current entity (HandleTask)                │ │
│  │ 2. Query Neo4j for related entities                   │ │
│  │ 3. Evaluate similarity scores                         │ │
│  │ 4. Decide: Process alone or with related entities?    │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  REASON                                                     │
│  ┌───────────────────────────────────────────────────────┐ │
│  │ If related entities found:                            │ │
│  │                                                        │ │
│  │ Branch 1: Optimize HandleTask alone                   │ │
│  │ Branch 2: Optimize HandleTask + related entities      │ │
│  │ Branch 3: Create shared abstraction for all           │ │
│  │                                                        │ │
│  │ Evaluate which branch is best                         │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  ACT                                                        │
│  ┌───────────────────────────────────────────────────────┐ │
│  │ If "optimize with related entities" selected:         │ │
│  │                                                        │ │
│  │ Parallel execution:                                   │ │
│  │ ├─ Thread 1: Process HandleTask                       │ │
│  │ ├─ Thread 2: Process HandleRequest                    │ │
│  │ └─ Thread 3: Process HandleEvent                      │ │
│  │                                                        │ │
│  │ Aggregate results                                     │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  REFLECT                                                    │
│  ┌───────────────────────────────────────────────────────┐ │
│  │ Cross-entity reflection:                              │ │
│  │                                                        │ │
│  │ - Did all entities benefit from same optimization?    │ │
│  │ - Should we create a shared pattern?                  │ │
│  │ - Update Neo4j relationships                          │ │
│  │ - Store cross-entity learning                         │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Enhanced MCP Tools

### 1. Enhanced `perceive` with Graph Discovery

**New Parameters:**
```json
{
  "task_id": "task-123",
  "goal": "Optimize HandleTask function",
  "capture_screenshot": true,
  "analyze_code_mirror": true,
  "discover_related_entities": true,  // NEW
  "similarity_threshold": 0.7,         // NEW
  "max_related_entities": 5            // NEW
}
```

**Enhanced Output:**
```json
{
  "perception_id": "perc-456",
  "confidence": 0.85,
  "primary_entity": {
    "id": "func-HandleTask",
    "type": "function",
    "file": "controller.go",
    "code": "..."
  },
  "related_entities": [
    {
      "id": "func-HandleRequest",
      "type": "function",
      "similarity": 0.92,
      "relationship": "SIMILAR_PATTERN",
      "reason": "Both handle incoming requests with validation"
    },
    {
      "id": "func-HandleEvent",
      "type": "function",
      "similarity": 0.85,
      "relationship": "SIMILAR_PATTERN",
      "reason": "Similar error handling and logging"
    },
    {
      "id": "func-ProcessTask",
      "type": "function",
      "similarity": 0.78,
      "relationship": "CALLS",
      "reason": "Called by HandleTask, may need coordinated changes"
    }
  ],
  "graph_context": {
    "total_similar_entities": 3,
    "relationship_types": ["SIMILAR_PATTERN", "CALLS"],
    "suggestion": "consider_batch_optimization"
  },
  "decision": "proceed_with_related_entities"
}
```

### 2. Enhanced `reason` with Cross-Entity Branching

**New Branching Strategy:**
```json
{
  "reasoning_id": "reason-789",
  "branches": [
    {
      "branch_id": "branch-1",
      "strategy": "Optimize HandleTask alone",
      "scope": "single_entity",
      "chain_of_thought": [
        "Add caching to HandleTask",
        "Risk: Other similar functions won't benefit",
        "Pro: Faster to implement"
      ],
      "feasibility_score": 0.9,
      "alignment_score": 0.7,
      "risk_score": 0.3,
      "selected": false
    },
    {
      "branch_id": "branch-2",
      "strategy": "Optimize all similar handlers",
      "scope": "multi_entity",
      "entities": ["HandleTask", "HandleRequest", "HandleEvent"],
      "chain_of_thought": [
        "All three have similar structure",
        "Apply caching pattern to all",
        "Consistent optimization across codebase",
        "Risk: More complex, but better long-term"
      ],
      "feasibility_score": 0.8,
      "alignment_score": 0.95,
      "risk_score": 0.4,
      "selected": true  // SELECTED due to high alignment
    },
    {
      "branch_id": "branch-3",
      "strategy": "Create shared abstraction",
      "scope": "refactoring",
      "entities": ["HandleTask", "HandleRequest", "HandleEvent"],
      "chain_of_thought": [
        "Extract common pattern into BaseHandler",
        "All three inherit from BaseHandler",
        "Optimization applied once in base",
        "Risk: Major refactoring, potential bugs"
      ],
      "feasibility_score": 0.6,
      "alignment_score": 0.9,
      "risk_score": 0.7,
      "selected": false
    }
  ],
  "selected_branch": {
    "branch_id": "branch-2",
    "rationale": "High alignment score, benefits multiple entities, acceptable risk"
  },
  "parallel_execution_plan": {
    "enabled": true,
    "entities": [
      {
        "id": "func-HandleTask",
        "subtasks": [
          {"id": "st-1", "description": "Add cache struct"},
          {"id": "st-2", "description": "Implement cache lookup"}
        ]
      },
      {
        "id": "func-HandleRequest",
        "subtasks": [
          {"id": "st-3", "description": "Add cache struct"},
          {"id": "st-4", "description": "Implement cache lookup"}
        ]
      },
      {
        "id": "func-HandleEvent",
        "subtasks": [
          {"id": "st-5", "description": "Add cache struct"},
          {"id": "st-6", "description": "Implement cache lookup"}
        ]
      }
    ]
  }
}
```

### 3. Enhanced `act` with Parallel Entity Processing

**New Execution Mode:**
```json
{
  "execution_id": "exec-101",
  "execution_mode": "parallel_entities",
  "entity_results": [
    {
      "entity_id": "func-HandleTask",
      "status": "success",
      "performance_improvement": "35%",
      "execution_time": 3.2
    },
    {
      "entity_id": "func-HandleRequest",
      "status": "success",
      "performance_improvement": "40%",
      "execution_time": 3.1
    },
    {
      "entity_id": "func-HandleEvent",
      "status": "partial_success",
      "performance_improvement": "25%",
      "execution_time": 3.5,
      "issues": ["Cache invalidation timing"]
    }
  ],
  "aggregate_results": {
    "total_entities_processed": 3,
    "successful": 2,
    "partial_success": 1,
    "failed": 0,
    "average_improvement": "33%",
    "pattern_consistency": 0.85
  },
  "cross_entity_insights": [
    "All three benefited from caching",
    "HandleEvent needs different cache TTL",
    "Pattern is reusable for future handlers"
  ],
  "decision": "proceed_to_cross_entity_reflection"
}
```

### 4. Enhanced `reflect` with Cross-Entity Learning

**New Reflection Type:**
```json
{
  "reflection_id": "reflect-202",
  "reflection_type": "cross_entity",
  "entities_analyzed": ["HandleTask", "HandleRequest", "HandleEvent"],
  
  "cross_entity_analysis": {
    "pattern_identified": "request_handler_caching",
    "success_rate": 0.85,
    "applicable_to": [
      "func-HandleTask",
      "func-HandleRequest", 
      "func-HandleEvent",
      "func-HandleCommand"  // Not yet optimized, but similar
    ],
    "variations_needed": {
      "cache_ttl": "Adjust based on request type",
      "invalidation_strategy": "Event handlers need immediate invalidation"
    }
  },
  
  "neo4j_updates": [
    {
      "type": "create_pattern_node",
      "pattern": {
        "id": "pattern-request-handler-cache",
        "name": "Request Handler Caching",
        "success_rate": 0.85,
        "performance_gain": "33%",
        "applicable_to": ["request_handlers"],
        "variations": {...}
      }
    },
    {
      "type": "create_relationships",
      "relationships": [
        "(HandleTask)-[:USES_PATTERN]->(pattern-request-handler-cache)",
        "(HandleRequest)-[:USES_PATTERN]->(pattern-request-handler-cache)",
        "(HandleEvent)-[:USES_PATTERN]->(pattern-request-handler-cache)",
        "(pattern-request-handler-cache)-[:APPLICABLE_TO]->(HandleCommand)"
      ]
    },
    {
      "type": "update_similarity_scores",
      "updates": [
        {
          "from": "HandleTask",
          "to": "HandleRequest",
          "old_similarity": 0.92,
          "new_similarity": 0.95,
          "reason": "Now both use same caching pattern"
        }
      ]
    }
  ],
  
  "evolutions": [
    {
      "type": "pattern_creation",
      "pattern_name": "request_handler_caching",
      "stored_in_neo4j": true,
      "reusable": true
    },
    {
      "type": "strategy_update",
      "strategy_name": "performance_optimization",
      "updates": {
        "add_step": "Check for similar entities before optimizing",
        "add_criterion": "Apply pattern to all similar entities"
      }
    }
  ],
  
  "future_recommendations": [
    {
      "entity_id": "func-HandleCommand",
      "recommendation": "Apply request_handler_caching pattern",
      "confidence": 0.88,
      "reason": "Similar structure to optimized handlers"
    }
  ],
  
  "decision": "complete_with_cross_entity_learning"
}
```

## Neo4j Graph Queries

### Finding Related Entities

```cypher
// Find similar functions by pattern
MATCH (f:Function {id: $function_id})
MATCH (similar:Function)
WHERE similar.id <> f.id
  AND (
    // Similar structure
    (f)-[:SIMILAR_PATTERN]-(similar)
    OR
    // Similar calls
    EXISTS {
      MATCH (f)-[:CALLS]->(common)<-[:CALLS]-(similar)
    }
    OR
    // Similar complexity
    (abs(f.complexity - similar.complexity) < 0.2)
  )
WITH similar, 
     // Calculate similarity score
     CASE
       WHEN (f)-[:SIMILAR_PATTERN]-(similar) THEN 0.4
       ELSE 0.0
     END +
     CASE
       WHEN EXISTS {MATCH (f)-[:CALLS]->(common)<-[:CALLS]-(similar)} THEN 0.3
       ELSE 0.0
     END +
     CASE
       WHEN abs(f.complexity - similar.complexity) < 0.2 THEN 0.3
       ELSE 0.0
     END AS similarity_score
WHERE similarity_score >= $threshold
RETURN similar, similarity_score
ORDER BY similarity_score DESC
LIMIT $max_entities
```

### Finding Entities That Would Benefit

```cypher
// Find entities that could use the same pattern
MATCH (pattern:Pattern {id: $pattern_id})
MATCH (pattern)-[:APPLICABLE_TO]->(entity_type:EntityType)
MATCH (entity:Function)
WHERE entity.type = entity_type.name
  AND NOT (entity)-[:USES_PATTERN]->(pattern)
  AND entity.complexity > 0.5  // Only suggest for complex functions
RETURN entity, 
       entity.complexity AS priority
ORDER BY priority DESC
LIMIT 10
```

### Storing Cross-Entity Patterns

```cypher
// Create pattern node with relationships
CREATE (p:Pattern {
  id: $pattern_id,
  name: $pattern_name,
  success_rate: $success_rate,
  performance_gain: $performance_gain,
  created_at: datetime()
})

// Link to entities that use it
WITH p
UNWIND $entity_ids AS entity_id
MATCH (e:Function {id: entity_id})
CREATE (e)-[:USES_PATTERN {
  applied_at: datetime(),
  improvement: $improvements[entity_id]
}]->(p)

// Link to applicable entity types
WITH p
UNWIND $applicable_types AS type_name
MERGE (et:EntityType {name: type_name})
CREATE (p)-[:APPLICABLE_TO]->(et)

RETURN p
```

## Implementation

### Enhanced Perceive Tool

```go
// backend/mcp_servers/dynamic_thinking/tools/perceive.go

func (t *PerceiveTool) Execute(params map[string]interface{}) (*PerceptionResult, error) {
    // ... existing perception logic ...
    
    // NEW: Discover related entities
    if params["discover_related_entities"].(bool) {
        relatedEntities, err := t.discoverRelatedEntities(
            primaryEntity,
            params["similarity_threshold"].(float64),
            params["max_related_entities"].(int),
        )
        if err != nil {
            return nil, err
        }
        
        result.RelatedEntities = relatedEntities
        
        // Decide if we should process them together
        if len(relatedEntities) > 0 && t.shouldProcessTogether(relatedEntities) {
            result.Decision = "proceed_with_related_entities"
            result.GraphContext = &GraphContext{
                TotalSimilarEntities: len(relatedEntities),
                Suggestion: "consider_batch_optimization",
            }
        }
    }
    
    return result, nil
}

func (t *PerceiveTool) discoverRelatedEntities(
    primary *Entity,
    threshold float64,
    maxEntities int,
) ([]*RelatedEntity, error) {
    query := `
        MATCH (f:Function {id: $function_id})
        MATCH (similar:Function)
        WHERE similar.id <> f.id
          AND (
            (f)-[:SIMILAR_PATTERN]-(similar)
            OR EXISTS {MATCH (f)-[:CALLS]->(common)<-[:CALLS]-(similar)}
          )
        WITH similar, 
             // Calculate similarity
             ...
        WHERE similarity_score >= $threshold
        RETURN similar, similarity_score
        ORDER BY similarity_score DESC
        LIMIT $max_entities
    `
    
    result, err := t.neo4j.Run(query, map[string]interface{}{
        "function_id": primary.ID,
        "threshold": threshold,
        "max_entities": maxEntities,
    })
    
    // ... parse and return ...
}
```

### Enhanced Reason Tool with Cross-Entity Branching

```go
// backend/mcp_servers/dynamic_thinking/tools/reason.go

func (t *ReasonTool) Execute(params map[string]interface{}) (*ReasoningResult, error) {
    perception := params["perception"].(*PerceptionResult)
    
    // Generate branches
    branches := []ReasoningBranch{}
    
    // Branch 1: Single entity
    branch1 := t.generateSingleEntityBranch(perception.PrimaryEntity)
    branches = append(branches, branch1)
    
    // Branch 2: Multi-entity (if related entities found)
    if len(perception.RelatedEntities) > 0 {
        branch2 := t.generateMultiEntityBranch(
            perception.PrimaryEntity,
            perception.RelatedEntities,
        )
        branches = append(branches, branch2)
        
        // Branch 3: Refactoring with abstraction
        branch3 := t.generateAbstractionBranch(
            perception.PrimaryEntity,
            perception.RelatedEntities,
        )
        branches = append(branches, branch3)
    }
    
    // Evaluate branches
    evaluations := t.evaluateBranches(branches)
    
    // Select best branch
    selectedBranch := branches[argmax(evaluations)]
    
    // If multi-entity branch selected, create parallel execution plan
    var parallelPlan *ParallelExecutionPlan
    if selectedBranch.Scope == "multi_entity" {
        parallelPlan = t.createParallelPlan(selectedBranch)
    }
    
    return &ReasoningResult{
        Branches: branches,
        SelectedBranch: selectedBranch,
        ParallelExecutionPlan: parallelPlan,
    }, nil
}
```

### Enhanced Act Tool with Parallel Execution

```go
// backend/mcp_servers/dynamic_thinking/tools/act.go

func (t *ActTool) Execute(params map[string]interface{}) (*ExecutionResult, error) {
    plan := params["action_plan"].(*ActionPlan)
    
    if plan.ParallelExecutionPlan != nil {
        // Execute entities in parallel
        return t.executeParallel(plan.ParallelExecutionPlan)
    }
    
    // Standard sequential execution
    return t.executeSequential(plan)
}

func (t *ActTool) executeParallel(plan *ParallelExecutionPlan) (*ExecutionResult, error) {
    results := make(chan *EntityResult, len(plan.Entities))
    errors := make(chan error, len(plan.Entities))
    
    // Spawn goroutine for each entity
    for _, entity := range plan.Entities {
        go func(e *EntityPlan) {
            result, err := t.executeEntity(e)
            if err != nil {
                errors <- err
                return
            }
            results <- result
        }(entity)
    }
    
    // Collect results
    entityResults := []*EntityResult{}
    for i := 0; i < len(plan.Entities); i++ {
        select {
        case result := <-results:
            entityResults = append(entityResults, result)
        case err := <-errors:
            return nil, err
        }
    }
    
    // Aggregate results
    aggregate := t.aggregateResults(entityResults)
    
    return &ExecutionResult{
        ExecutionMode: "parallel_entities",
        EntityResults: entityResults,
        AggregateResults: aggregate,
        CrossEntityInsights: t.extractCrossEntityInsights(entityResults),
        Decision: "proceed_to_cross_entity_reflection",
    }, nil
}
```

### Enhanced Reflect Tool with Cross-Entity Learning

```go
// backend/mcp_servers/dynamic_thinking/tools/reflect.go

func (t *ReflectTool) Execute(params map[string]interface{}) (*ReflectionResult, error) {
    execution := params["execution"].(*ExecutionResult)
    
    if execution.ExecutionMode == "parallel_entities" {
        return t.reflectCrossEntity(execution)
    }
    
    return t.reflectSingleEntity(execution)
}

func (t *ReflectTool) reflectCrossEntity(execution *ExecutionResult) (*ReflectionResult, error) {
    // Analyze pattern across entities
    pattern := t.identifyPattern(execution.EntityResults)
    
    // Create pattern node in Neo4j
    patternNode, err := t.createPatternNode(pattern)
    if err != nil {
        return nil, err
    }
    
    // Link entities to pattern
    err = t.linkEntitiesToPattern(execution.EntityResults, patternNode)
    if err != nil {
        return nil, err
    }
    
    // Find other entities that could benefit
    recommendations := t.findApplicableEntities(patternNode)
    
    return &ReflectionResult{
        ReflectionType: "cross_entity",
        CrossEntityAnalysis: &CrossEntityAnalysis{
            PatternIdentified: pattern.Name,
            SuccessRate: pattern.SuccessRate,
            ApplicableTo: pattern.ApplicableEntities,
        },
        Neo4jUpdates: t.generateNeo4jUpdates(pattern, execution),
        FutureRecommendations: recommendations,
        Decision: "complete_with_cross_entity_learning",
    }, nil
}
```

## Configuration

```yaml
# config/dynamic_thinking.yaml
perception:
  discover_related_entities: true
  similarity_threshold: 0.7
  max_related_entities: 5
  relationship_types:
    - SIMILAR_PATTERN
    - CALLS
    - CALLED_BY
    - SIMILAR_COMPLEXITY

reasoning:
  enable_cross_entity_branching: true
  cross_entity_branch_weight: 1.2  # Prefer cross-entity solutions
  min_entities_for_abstraction: 3

action:
  enable_parallel_execution: true
  max_parallel_entities: 5
  timeout_per_entity: 60

reflection:
  enable_pattern_creation: true
  min_success_rate_for_pattern: 0.7
  auto_recommend_similar_entities: true
  max_recommendations: 10
```

## Example Flow

### User Request
```
User: "Optimize the HandleTask function"
```

### 1. Perceive (with Graph Discovery)
```
Primary Entity: HandleTask
    ↓
Query Neo4j for similar entities
    ↓
Found:
- HandleRequest (similarity: 0.92)
- HandleEvent (similarity: 0.85)
- ProcessTask (similarity: 0.78, relationship: CALLS)
    ↓
Decision: proceed_with_related_entities
```

### 2. Reason (with Cross-Entity Branching)
```
Branch 1: Optimize HandleTask alone
  Feasibility: 0.9, Alignment: 0.7
  
Branch 2: Optimize all three handlers together ✓ SELECTED
  Feasibility: 0.8, Alignment: 0.95
  
Branch 3: Create BaseHandler abstraction
  Feasibility: 0.6, Alignment: 0.9
```

### 3. Act (Parallel Execution)
```
Parallel threads:
├─ HandleTask: Add caching → 35% improvement
├─ HandleRequest: Add caching → 40% improvement
└─ HandleEvent: Add caching → 25% improvement

Aggregate: 33% average improvement
```

### 4. Reflect (Cross-Entity Learning)
```
Pattern identified: "request_handler_caching"
Success rate: 85%

Neo4j updates:
- Create pattern node
- Link all three handlers to pattern
- Find HandleCommand (similar, not yet optimized)
- Recommend applying pattern to HandleCommand

Future: When optimizing any handler, pattern is available
```

## Benefits

### 1. Holistic Optimization
Instead of optimizing one function, optimizes entire families of related code

### 2. Pattern Discovery
Automatically identifies reusable patterns across codebase

### 3. Knowledge Propagation
Learning from one entity benefits all similar entities

### 4. Proactive Recommendations
Suggests improvements to related code before being asked

### 5. Graph-Driven Intelligence
Leverages Neo4j relationships for smarter decisions

## Summary

This enhancement transforms dynamic thinking from **entity-centric** to **graph-aware**:

- ✅ Discovers related entities during perception
- ✅ Branches to consider cross-entity optimizations
- ✅ Executes changes in parallel across entities
- ✅ Reflects on patterns across multiple entities
- ✅ Stores learnings as reusable patterns in Neo4j
- ✅ Recommends applying patterns to similar code

The agent now thinks in terms of **graph structures**, not isolated entities! 🕸️

