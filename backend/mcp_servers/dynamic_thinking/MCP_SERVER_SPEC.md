# Dynamic Thinking MCP Server - Complete Specification

## Overview

This is THE critical MCP server for agent self-awareness and deep reasoning. It implements the complete PRAR loop with enhanced multi-modal perception, graph-aware branching, confidence-based retrieval, and multiple reasoning modes.

## MCP Tools

### 1. `perceive` - Enhanced Multi-Modal Perception

**Description:** Deep analytical understanding using systems thinking, contextual reasoning, meta-perception, and three reasoning modes (deductive, inductive, abductive).

**Input:**
```json
{
  "task_id": "string",
  "task": "string",
  "goal": "string",
  "entity": {
    "name": "string",
    "type": "string",
    "code": "string (optional)"
  },
  "context": {
    "user_request": "string",
    "current_state": "object"
  }
}
```

**Output:**
```json
{
  "perception_id": "uuid",
  "task": "string",
  "goal": "string",
  
  "systems_view": {
    "components": [
      {
        "name": "string",
        "type": "string",
        "role": "string"
      }
    ],
    "relationships": [
      {
        "from": "string",
        "to": "string",
        "type": "string"
      }
    ],
    "bottlenecks": [
      {
        "location": "string",
        "description": "string",
        "impact": "float"
      }
    ],
    "emergent_behavior": "string"
  },
  
  "contextual_view": {
    "constraints": ["string"],
    "opportunities": ["string"],
    "risks": ["string"],
    "stakeholders": ["string"],
    "relevance": {
      "component_name": "float"
    }
  },
  
  "meta_view": {
    "recommended_framing": {
      "name": "string",
      "description": "string",
      "score": "float"
    },
    "alternative_framings": [
      {
        "name": "string",
        "score": "float"
      }
    ],
    "better_questions": ["string"],
    "blind_spots": ["string"],
    "assumptions": ["string"]
  },
  
  "reasoning": {
    "deductive": {
      "principles": [
        {
          "statement": "string",
          "confidence": "float"
        }
      ],
      "conclusions": [
        {
          "statement": "string",
          "confidence": "float"
        }
      ],
      "overall_confidence": "float"
    },
    "inductive": {
      "past_observations": [
        {
          "case": "string",
          "outcome": "string",
          "metrics": {}
        }
      ],
      "generalizations": [
        {
          "statement": "string",
          "sample_size": "int",
          "confidence": "float"
        }
      ],
      "overall_confidence": "float"
    },
    "abductive": {
      "observation": "string",
      "hypotheses": [
        {
          "explanation": "string",
          "likelihood": "float",
          "testable": "bool"
        }
      ],
      "best_explanation": {
        "explanation": "string",
        "likelihood": "float"
      },
      "overall_confidence": "float"
    },
    "synthesis": "string"
  },
  
  "similar_entities": [
    {
      "id": "string",
      "name": "string",
      "similarity": "float",
      "relationship": "string"
    }
  ],
  
  "processing_mode": "single_entity | cross_entity",
  "confidence": "float",
  "confidence_factors": {
    "deductive": "float",
    "inductive": "float",
    "abductive": "float"
  }
}
```

**Implementation:**
1. Query LightRAG for system context
2. Perform systems thinking analysis
3. Calculate contextual relevance
4. Generate alternative framings (meta-perception)
5. Apply deductive reasoning (find principles)
6. Apply inductive reasoning (find patterns)
7. Apply abductive reasoning (best explanation)
8. Synthesize all reasoning modes
9. Find similar entities via LightRAG
10. Calculate overall confidence
11. Store perception in LightRAG with UUID

---

### 2. `reason` - Multi-Branch Reasoning with Confidence Check

**Description:** Generate reasoning branches using past knowledge, evaluate with multiple criteria, detect low confidence, trigger online retrieval if needed, and select best strategy.

**Input:**
```json
{
  "task_id": "string",
  "perception_id": "uuid",
  "perception": "object (from perceive tool)",
  "max_branches": "int (default: 3)"
}
```

**Output:**
```json
{
  "reasoning_id": "uuid",
  "branches": [
    {
      "id": "int",
      "strategy": "string",
      "description": "string",
      "confidence": "float",
      "confidence_factors": {
        "past_experience": "float",
        "pattern_availability": "float",
        "online_evidence": "float",
        "strategy_clarity": "float",
        "risk_assessment": "float"
      },
      "feasibility": "float",
      "alignment": "float",
      "risk": "float",
      "evidence": ["string"]
    }
  ],
  "selected_branch": {
    "id": "int",
    "strategy": "string",
    "confidence": "float"
  },
  "similar_cases": [
    {
      "id": "string",
      "description": "string",
      "outcome": "string",
      "relevance": "float"
    }
  ],
  "retrieved_info": {
    "triggered": "bool",
    "reason": "string",
    "sources": [
      {
        "type": "web | github | stackoverflow | youtube",
        "title": "string",
        "url": "string",
        "relevance": "float",
        "key_insights": ["string"]
      }
    ],
    "confidence_boost": "float"
  },
  "processing_mode": "single_entity | cross_entity",
  "confidence": "float"
}
```

**Implementation:**
1. Retrieve perception from LightRAG using perception_id
2. Query LightRAG for similar past reasoning
3. Generate initial branches (single-entity or cross-entity based on perception.processing_mode)
4. Calculate confidence for each branch
5. Check if ALL branches have low confidence (< threshold)
6. If low confidence:
   - Generate retrieval query based on processing mode
   - Search online (web, GitHub, Stack Overflow, YouTube)
   - Extract insights, best practices, code examples
   - Regenerate branches WITH online knowledge
   - Recalculate confidence (boosted by evidence)
7. Evaluate branches: feasibility, alignment, risk
8. Select best branch (highest confidence)
9. Store reasoning in LightRAG with UUID
10. Link to perception_id

---

### 3. `act` - Dynamic Execution with Watchdog

**Description:** Execute action plan with dynamic adjustment, watchdog monitoring, and parallel execution for cross-entity tasks.

**Input:**
```json
{
  "task_id": "string",
  "reasoning_id": "uuid",
  "selected_branch": "object (from reason tool)",
  "perception": "object",
  "dry_run": "bool (default: false)"
}
```

**Output:**
```json
{
  "execution_id": "uuid",
  "status": "success | partial_success | failure",
  "results": "string",
  "performance_gain": "float",
  "confidence": "float",
  
  "execution_trace": [
    {
      "step": "int",
      "action": "string",
      "status": "success | failure",
      "duration_ms": "int",
      "result": "string"
    }
  ],
  
  "parallel_executions": [
    {
      "entity": "string",
      "status": "success | failure",
      "performance_gain": "float"
    }
  ],
  
  "watchdog_events": [
    {
      "timestamp": "string",
      "severity": "info | warning | error",
      "message": "string",
      "action_taken": "string"
    }
  ],
  
  "adjustments": [
    {
      "reason": "string",
      "original_plan": "string",
      "adjusted_plan": "string"
    }
  ],
  
  "context_used": {
    "perception_id": "uuid",
    "reasoning_id": "uuid",
    "similar_executions": ["uuid"]
  }
}
```

**Implementation:**
1. Retrieve complete context from LightRAG (perception + reasoning)
2. Query LightRAG for similar past executions
3. Create execution plan from selected branch
4. If cross-entity mode:
   - Execute in parallel for all similar entities
   - Monitor each execution
   - Aggregate results
5. For each step:
   - Execute with watchdog monitoring
   - Check for security issues, quality gates
   - Dynamically adjust plan if needed
   - Store intermediate results
6. Calculate performance gain
7. Store execution in LightRAG with UUID
8. Link to perception_id and reasoning_id

---

### 4. `reflect` - Self-Critique and Pattern Creation

**Description:** Reflect on complete process, extract learned patterns, create recommendations for similar entities, and store training data.

**Input:**
```json
{
  "task_id": "string",
  "execution_id": "uuid",
  "perception": "object",
  "reasoning": "object",
  "execution": "object"
}
```

**Output:**
```json
{
  "reflection_id": "uuid",
  "success": "bool",
  "performance_gain": "float",
  "confidence": "float",
  
  "learnings": [
    {
      "category": "string",
      "insight": "string",
      "confidence": "float"
    }
  ],
  
  "patterns_discovered": [
    {
      "id": "uuid",
      "name": "string",
      "description": "string",
      "applicability": "string",
      "success_rate": "float",
      "performance_gain": "float",
      "code_template": "string (optional)",
      "linked_sources": [
        {
          "type": "web | github | youtube",
          "url": "string"
        }
      ]
    }
  ],
  
  "recommendations": [
    {
      "entity": "string",
      "similarity": "float",
      "recommended_pattern": "uuid",
      "expected_gain": "float",
      "confidence": "float"
    }
  ],
  
  "similar_processes": [
    {
      "id": "uuid",
      "description": "string",
      "outcome": "string",
      "relevance": "float"
    }
  ],
  
  "critique": {
    "what_went_well": ["string"],
    "what_could_improve": ["string"],
    "alternative_approaches": ["string"]
  },
  
  "training_data": {
    "perception_vector_id": "uuid",
    "reasoning_vector_id": "uuid",
    "execution_vector_id": "uuid",
    "holistic_vector_id": "uuid",
    "stored_in_lightrag": "bool"
  },
  
  "prompt_evolution": {
    "original_prompt": "string",
    "evolved_prompt": "string",
    "improvement_reason": "string"
  }
}
```

**Implementation:**
1. Query LightRAG for ENTIRE process (perception + reasoning + execution)
2. Query for similar complete processes
3. Analyze success/failure
4. Extract learnings from process
5. Identify patterns that emerged
6. For each pattern:
   - Generate UUID
   - Create pattern node in Neo4j (via LightRAG)
   - Store pattern vector
   - Link to online sources if used
7. Find similar entities in graph
8. Generate recommendations for each similar entity
9. Create self-critique
10. Create holistic embedding combining all phases
11. Store reflection in LightRAG with UUID
12. Evolve prompts based on learnings
13. Generate training data for next generation

---

### 5. `query_memory` - Semantic Memory Search

**Description:** Query agent's memory (LightRAG) for relevant past experiences, patterns, or knowledge.

**Input:**
```json
{
  "query": "string",
  "context": "string (optional)",
  "filters": {
    "type": "perception | reasoning | execution | reflection | pattern",
    "success": "bool (optional)",
    "min_confidence": "float (optional)",
    "date_range": {
      "start": "string (optional)",
      "end": "string (optional)"
    }
  },
  "max_results": "int (default: 10)"
}
```

**Output:**
```json
{
  "results": [
    {
      "id": "uuid",
      "type": "string",
      "content": "string",
      "relevance": "float",
      "metadata": {
        "confidence": "float",
        "success": "bool",
        "performance_gain": "float",
        "timestamp": "string"
      },
      "related_entities": [
        {
          "id": "uuid",
          "name": "string",
          "relationship": "string"
        }
      ]
    }
  ],
  "total_found": "int",
  "query_confidence": "float"
}
```

**Implementation:**
1. Create embedding for query using nomic-embed-text
2. Search vector DB (ChromeM) for similar vectors
3. Apply filters (type, success, confidence, date)
4. Retrieve full nodes from Neo4j using UUIDs
5. Get related entities via graph traversal
6. Sort by relevance
7. Return top-K results

---

### 6. `evolve_prompt` - Prompt Evolution

**Description:** Evolve prompts based on past performance and learnings.

**Input:**
```json
{
  "current_prompt": "string",
  "context": "string",
  "past_performance": {
    "success_rate": "float",
    "avg_confidence": "float",
    "common_failures": ["string"]
  },
  "learnings": ["string"]
}
```

**Output:**
```json
{
  "evolved_prompt": "string",
  "changes": [
    {
      "type": "addition | removal | modification",
      "description": "string",
      "reason": "string"
    }
  ],
  "expected_improvement": "float",
  "confidence": "float"
}
```

**Implementation:**
1. Analyze current prompt effectiveness
2. Identify weaknesses from past_performance
3. Apply learnings to improve prompt
4. Generate evolved prompt
5. Calculate expected improvement
6. Store prompt evolution in LightRAG

---

## MCP Server Configuration

```json
{
  "name": "dynamic_thinking",
  "version": "1.0.0",
  "description": "Enhanced multi-modal perception and reasoning for agent self-awareness",
  
  "tools": [
    {
      "name": "perceive",
      "description": "Deep analytical understanding using systems thinking, contextual reasoning, meta-perception, and three reasoning modes"
    },
    {
      "name": "reason",
      "description": "Multi-branch reasoning with confidence check and online retrieval"
    },
    {
      "name": "act",
      "description": "Dynamic execution with watchdog monitoring and parallel processing"
    },
    {
      "name": "reflect",
      "description": "Self-critique, pattern creation, and training data generation"
    },
    {
      "name": "query_memory",
      "description": "Semantic search across agent's memory"
    },
    {
      "name": "evolve_prompt",
      "description": "Evolve prompts based on past performance"
    }
  ],
  
  "dependencies": {
    "llm": {
      "provider": "ollama",
      "model": "gemma3:27b",
      "base_url": "http://localhost:11434"
    },
    "embedding": {
      "provider": "ollama",
      "model": "nomic-embed-text:v1.5",
      "base_url": "http://localhost:11434"
    },
    "storage": {
      "lightrag": {
        "graph": "neo4j://localhost:7687",
        "vector": "/var/lib/agent/vectors.db",
        "kv": "/var/lib/agent/kv.db"
      }
    },
    "retrieval": {
      "sources": ["web", "github", "stackoverflow", "youtube"],
      "max_results": 10
    }
  },
  
  "config": {
    "perception": {
      "systems_thinking": true,
      "contextual_reasoning": true,
      "meta_perception": true,
      "reasoning_modes": {
        "deductive": {"enabled": true, "weight": 0.4},
        "inductive": {"enabled": true, "weight": 0.35},
        "abductive": {"enabled": true, "weight": 0.25}
      }
    },
    "reasoning": {
      "max_branches": 3,
      "cross_entity_branching": true,
      "confidence_threshold": {
        "proceed": 0.75,
        "retrieve": 0.6
      },
      "online_retrieval": {
        "enabled": true,
        "trigger_on_low_confidence": true
      }
    },
    "action": {
      "watchdog": true,
      "parallel_execution": true,
      "max_parallel": 5,
      "confidence_threshold": {
        "proceed": 0.8
      }
    },
    "reflection": {
      "pattern_creation": true,
      "auto_recommend": true,
      "prompt_evolution": true,
      "training_data_generation": true
    }
  }
}
```

## Implementation Structure

```
backend/mcp_servers/dynamic_thinking/
├── server.py                 # Main MCP server (FastMCP)
├── tools/
│   ├── perceive.py          # Enhanced perception tool
│   ├── reason.py            # Multi-branch reasoning tool
│   ├── act.py               # Dynamic execution tool
│   ├── reflect.py           # Reflection and pattern creation tool
│   ├── query_memory.py      # Memory search tool
│   └── evolve_prompt.py     # Prompt evolution tool
├── reasoning/
│   ├── deductive.py         # Deductive reasoning
│   ├── inductive.py         # Inductive reasoning
│   └── abductive.py         # Abductive reasoning
├── perception/
│   ├── systems_thinking.py  # Systems analysis
│   ├── contextual.py        # Contextual reasoning
│   └── meta_perception.py   # Meta-perception
├── retrieval/
│   ├── web_search.py        # Web search
│   ├── github_search.py     # GitHub search
│   ├── stackoverflow.py     # Stack Overflow search
│   └── youtube.py           # YouTube + transcript extraction
├── storage/
│   ├── lightrag_client.py   # LightRAG integration
│   └── session_manager.py   # Session tracking
├── config.py                # Configuration management
├── requirements.txt         # Dependencies
└── README.md               # Documentation
```

## Usage Example (from Go Agent)

```go
// 1. PERCEIVE
perceptionResult, err := mcpClient.CallTool("dynamic_thinking", "perceive", map[string]interface{}{
    "task_id": taskID,
    "task":    "Optimize HandleTask function",
    "goal":    "Improve performance",
    "entity": map[string]interface{}{
        "name": "HandleTask",
        "type": "function",
    },
})

// 2. REASON
reasoningResult, err := mcpClient.CallTool("dynamic_thinking", "reason", map[string]interface{}{
    "task_id":      taskID,
    "perception_id": perceptionResult["perception_id"],
    "perception":    perceptionResult,
})

// 3. ACT
executionResult, err := mcpClient.CallTool("dynamic_thinking", "act", map[string]interface{}{
    "task_id":         taskID,
    "reasoning_id":    reasoningResult["reasoning_id"],
    "selected_branch": reasoningResult["selected_branch"],
    "perception":      perceptionResult,
})

// 4. REFLECT
reflectionResult, err := mcpClient.CallTool("dynamic_thinking", "reflect", map[string]interface{}{
    "task_id":   taskID,
    "execution_id": executionResult["execution_id"],
    "perception":   perceptionResult,
    "reasoning":    reasoningResult,
    "execution":    executionResult,
})

// Patterns discovered are now in Neo4j and can be queried!
```

## Summary

This MCP server is the **core of agent self-awareness**:

✅ **Enhanced Perception** - Systems thinking, contextual reasoning, meta-perception, 3 reasoning modes  
✅ **Graph-Aware Branching** - Finds similar entities, considers cross-entity optimization  
✅ **Confidence-Based Retrieval** - Searches online when confidence is low  
✅ **Dynamic Execution** - Watchdog monitoring, parallel processing  
✅ **Pattern Creation** - Learns and stores reusable patterns  
✅ **Training Data Generation** - Creates data for next generation  

**This is THE MCP server that makes the agent truly intelligent!** 🧠

