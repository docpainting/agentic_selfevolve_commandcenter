# Confidence-Based Online Retrieval in Dynamic Thinking

## Overview

The graph-aware PRAR loop can be long and complex, especially when processing multiple related entities. The agent needs **confidence tracking** at each phase and the ability to **retrieve additional information online** when confidence drops below thresholds.

## The Problem

**Current approach:**
```
Agent attempts complex optimization
    ↓
Confidence: 0.4 (low!)
    ↓
Proceeds anyway with uncertain approach
    ↓
Result: Suboptimal or incorrect solution
```

**Better approach:**
```
Agent attempts complex optimization
    ↓
Confidence: 0.4 (low!)
    ↓
Searches online: "Go caching best practices"
    ↓
Retrieves: Articles, documentation, examples
    ↓
Confidence: 0.85 (high!)
    ↓
Proceeds with informed approach
    ↓
Result: Optimal solution based on industry best practices
```

## Confidence Tracking

### Confidence Scores at Each Phase

```
┌─────────────────────────────────────────────────────────────┐
│  PERCEIVE                                                   │
│  Confidence: 0.85                                           │
│  ✓ High confidence → Proceed                               │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  REASON                                                     │
│  Branch 1 Confidence: 0.45 (low!)                          │
│  Branch 2 Confidence: 0.50 (low!)                          │
│  Branch 3 Confidence: 0.40 (low!)                          │
│  ⚠ All branches low confidence → RETRIEVE INFORMATION      │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  ONLINE RETRIEVAL                                           │
│  Search: "Go request handler optimization patterns"        │
│  Found: 5 articles, 3 GitHub repos, 2 Stack Overflow       │
│  Synthesize: Best practices for handler caching            │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  REASON (Retry with new information)                        │
│  Branch 1 Confidence: 0.88 (high!)                         │
│  Branch 2 Confidence: 0.92 (high!)                         │
│  Branch 3 Confidence: 0.75 (acceptable)                    │
│  ✓ High confidence → Proceed with Branch 2                 │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  ACT                                                        │
│  Confidence: 0.90                                           │
│  ✓ High confidence → Execute                               │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  REFLECT                                                    │
│  Confidence: 0.95                                           │
│  ✓ High confidence → Store learnings                       │
└─────────────────────────────────────────────────────────────┘
```

## Confidence Calculation

### Factors Contributing to Confidence

```go
type ConfidenceScore struct {
    Overall float64
    Factors map[string]float64
}

func calculateConfidence(context *Context) ConfidenceScore {
    factors := make(map[string]float64)
    
    // 1. Past experience (Neo4j)
    similarTasks := queryNeo4j("MATCH (t:Task) WHERE t.similar_to = $current")
    if len(similarTasks) > 0 {
        factors["past_experience"] = 0.3
    } else {
        factors["past_experience"] = 0.0  // Never done this before!
    }
    
    // 2. Pattern availability
    applicablePatterns := queryNeo4j("MATCH (p:Pattern)-[:APPLICABLE_TO]->($entity_type)")
    if len(applicablePatterns) > 0 {
        factors["pattern_availability"] = 0.25
    } else {
        factors["pattern_availability"] = 0.0  // No known patterns!
    }
    
    // 3. Code understanding
    codeAnalysis := analyzeCode(context.TargetCode)
    if codeAnalysis.Complexity < 0.5 {
        factors["code_understanding"] = 0.2
    } else {
        factors["code_understanding"] = 0.1  // Complex code, less confident
    }
    
    // 4. Strategy clarity
    if context.Strategy.StepsWellDefined {
        factors["strategy_clarity"] = 0.15
    } else {
        factors["strategy_clarity"] = 0.0  // Vague strategy!
    }
    
    // 5. Risk assessment
    if context.Risk.Score < 0.3 {
        factors["risk_assessment"] = 0.1
    } else {
        factors["risk_assessment"] = 0.0  // High risk!
    }
    
    // Calculate overall confidence
    overall := 0.0
    for _, score := range factors {
        overall += score
    }
    
    return ConfidenceScore{
        Overall: overall,
        Factors: factors,
    }
}
```

### Confidence Thresholds

```yaml
# config/confidence.yaml
thresholds:
  perceive:
    proceed: 0.7
    retrieve: 0.5
    abort: 0.3
  
  reason:
    proceed: 0.75
    retrieve: 0.6
    abort: 0.4
  
  act:
    proceed: 0.8
    retrieve: 0.65
    abort: 0.5
  
  reflect:
    proceed: 0.7
    retrieve: 0.55
    abort: 0.4

actions:
  high_confidence: "proceed"           # >= proceed threshold
  medium_confidence: "retrieve_info"   # >= retrieve, < proceed
  low_confidence: "abort_and_ask"      # < retrieve threshold
```

## Online Retrieval Strategy

### When to Retrieve

```go
func (p *Phase) shouldRetrieveOnline(confidence ConfidenceScore) bool {
    thresholds := p.config.Thresholds
    
    // Check overall confidence
    if confidence.Overall < thresholds.Proceed && 
       confidence.Overall >= thresholds.Retrieve {
        return true
    }
    
    // Check specific factors
    if confidence.Factors["past_experience"] == 0.0 {
        // Never done this before - definitely retrieve!
        return true
    }
    
    if confidence.Factors["pattern_availability"] == 0.0 &&
       confidence.Factors["strategy_clarity"] < 0.1 {
        // No patterns and unclear strategy - retrieve!
        return true
    }
    
    return false
}
```

### What to Search For

```go
type RetrievalQuery struct {
    Query string
    Context string
    Sources []string
    MaxResults int
}

func generateRetrievalQuery(phase string, context *Context, confidence ConfidenceScore) RetrievalQuery {
    // Identify knowledge gaps from confidence factors
    gaps := []string{}
    
    if confidence.Factors["past_experience"] == 0.0 {
        gaps = append(gaps, "no_past_experience")
    }
    if confidence.Factors["pattern_availability"] == 0.0 {
        gaps = append(gaps, "no_known_patterns")
    }
    if confidence.Factors["strategy_clarity"] < 0.1 {
        gaps = append(gaps, "unclear_strategy")
    }
    
    // Generate search query based on phase and gaps
    query := ""
    
    switch phase {
    case "perceive":
        if contains(gaps, "no_past_experience") {
            query = fmt.Sprintf("How to analyze %s in %s", 
                context.TargetEntity, context.Language)
        }
    
    case "reason":
        if contains(gaps, "no_known_patterns") {
            query = fmt.Sprintf("%s optimization best practices %s",
                context.TaskType, context.Language)
        }
        if contains(gaps, "unclear_strategy") {
            query = fmt.Sprintf("How to %s in %s step by step",
                context.Goal, context.Language)
        }
    
    case "act":
        if contains(gaps, "no_past_experience") {
            query = fmt.Sprintf("%s implementation examples %s",
                context.Strategy, context.Language)
        }
    
    case "reflect":
        query = fmt.Sprintf("Common mistakes when %s",
            context.TaskType)
    }
    
    return RetrievalQuery{
        Query: query,
        Context: context.Description,
        Sources: []string{"web", "github", "stackoverflow", "docs"},
        MaxResults: 10,
    }
}
```

### Retrieval Sources

```go
type RetrievalSource interface {
    Search(query string, maxResults int) ([]Document, error)
}

type WebSearch struct {
    // General web search (Google, Bing, etc.)
}

type GitHubSearch struct {
    // Search GitHub repos and code
}

type StackOverflowSearch struct {
    // Search Stack Overflow Q&A
}

type DocumentationSearch struct {
    // Search official docs (Go, Python, etc.)
}

type ArxivSearch struct {
    // Search academic papers for advanced topics
}

type YouTubeSearch struct {
    // Search YouTube with transcript extraction
    // CRITICAL for bleeding-edge content and expert explanations
}

func retrieveInformation(query RetrievalQuery) (*RetrievalResult, error) {
    results := &RetrievalResult{
        Query: query.Query,
        Documents: []Document{},
    }
    
    // Search multiple sources in parallel
    var wg sync.WaitGroup
    resultChan := make(chan []Document, len(query.Sources))
    
    for _, source := range query.Sources {
        wg.Add(1)
        go func(s string) {
            defer wg.Done()
            
            var docs []Document
            var err error
            
            switch s {
            case "web":
                docs, err = webSearch.Search(query.Query, query.MaxResults)
            case "github":
                docs, err = githubSearch.Search(query.Query, query.MaxResults)
            case "stackoverflow":
                docs, err = stackOverflowSearch.Search(query.Query, query.MaxResults)
            case "docs":
                docs, err = docsSearch.Search(query.Query, query.MaxResults)
            case "youtube":
                docs, err = youtubeSearch.SearchWithTranscripts(query.Query, query.MaxResults)
            }
            
            if err == nil {
                resultChan <- docs
            }
        }(source)
    }
    
    wg.Wait()
    close(resultChan)
    
    // Collect all results
    for docs := range resultChan {
        results.Documents = append(results.Documents, docs...)
    }
    
    // Rank by relevance
    results.Documents = rankByRelevance(results.Documents, query)
    
    // Synthesize key insights
    results.Synthesis = synthesizeInsights(results.Documents)
    
    return results, nil
}
```

### Synthesis of Retrieved Information

```go
type Synthesis struct {
    KeyInsights []string
    BestPractices []string
    CommonMistakes []string
    CodeExamples []CodeExample
    RelevantPatterns []Pattern
}

func synthesizeInsights(documents []Document) *Synthesis {
    // Use LLM to extract key information
    prompt := fmt.Sprintf(`
Analyze the following documents and extract:
1. Key insights relevant to the task
2. Best practices mentioned
3. Common mistakes to avoid
4. Relevant code examples
5. Patterns that could be applied

Documents:
%s

Provide a structured summary.
`, documentsToText(documents))
    
    response := llm.Generate(prompt)
    
    return parseSynthesis(response)
}
```

## Enhanced MCP Tools

### New Tool: `retrieve_information`

```json
{
  "name": "retrieve_information",
  "description": "Search online for additional information when confidence is low",
  "input_schema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "What to search for"
      },
      "context": {
        "type": "string",
        "description": "Current task context"
      },
      "sources": {
        "type": "array",
        "items": {"type": "string"},
        "description": "Sources to search (web, github, stackoverflow, docs, arxiv)"
      },
      "max_results": {
        "type": "integer",
        "default": 10
      }
    },
    "required": ["query", "context"]
  }
}
```

**Output:**
```json
{
  "retrieval_id": "retr-123",
  "query": "Go request handler optimization best practices",
  "documents": [
    {
      "title": "Optimizing HTTP Handlers in Go",
      "url": "https://example.com/article",
      "source": "web",
      "relevance": 0.95,
      "summary": "Discusses caching strategies for Go handlers..."
    },
    {
      "title": "go-cache library",
      "url": "https://github.com/patrickmn/go-cache",
      "source": "github",
      "relevance": 0.88,
      "summary": "In-memory caching library with TTL support..."
    }
  ],
  "synthesis": {
    "key_insights": [
      "Caching is most effective for read-heavy handlers",
      "Use TTL to prevent stale data",
      "Consider cache invalidation strategy"
    ],
    "best_practices": [
      "Use sync.Map or third-party cache library",
      "Set appropriate TTL based on data volatility",
      "Implement cache warming for critical data"
    ],
    "common_mistakes": [
      "Not handling cache misses gracefully",
      "Forgetting to invalidate on updates",
      "Caching too much data (memory issues)"
    ],
    "code_examples": [
      {
        "description": "Basic handler caching with go-cache",
        "code": "...",
        "source": "github"
      }
    ]
  },
  "confidence_boost": 0.35,
  "new_confidence": 0.85
}
```

### Enhanced `perceive` with Confidence and Retrieval

```go
func (t *PerceiveTool) Execute(params map[string]interface{}) (*PerceptionResult, error) {
    // ... existing perception logic ...
    
    // Calculate confidence
    confidence := t.calculateConfidence(perception)
    result.Confidence = confidence.Overall
    result.ConfidenceFactors = confidence.Factors
    
    // Check if retrieval needed
    if t.shouldRetrieveOnline(confidence) {
        log.Info("Low confidence in perception, retrieving additional information",
                 "confidence", confidence.Overall)
        
        // Generate retrieval query
        query := t.generateRetrievalQuery("perceive", perception, confidence)
        
        // Retrieve information
        retrieval, err := t.retrieveInformation(query)
        if err != nil {
            log.Warn("Retrieval failed", "error", err)
        } else {
            // Incorporate retrieved information
            result.RetrievedInformation = retrieval
            
            // Recalculate confidence with new information
            newConfidence := t.recalculateConfidence(perception, retrieval)
            result.Confidence = newConfidence.Overall
            result.ConfidenceBoost = newConfidence.Overall - confidence.Overall
            
            log.Info("Confidence boosted by retrieval",
                     "old", confidence.Overall,
                     "new", newConfidence.Overall,
                     "boost", result.ConfidenceBoost)
        }
    }
    
    // Decide next action based on final confidence
    if result.Confidence >= t.config.Thresholds.Proceed {
        result.Decision = "proceed"
    } else if result.Confidence >= t.config.Thresholds.Retrieve {
        result.Decision = "gather_more_info"
    } else {
        result.Decision = "abort_and_ask_user"
    }
    
    return result, nil
}
```

### Enhanced `reason` with Confidence-Based Retrieval

```go
func (t *ReasonTool) Execute(params map[string]interface{}) (*ReasoningResult, error) {
    perception := params["perception"].(*PerceptionResult)
    
    // Generate initial branches
    branches := t.generateBranches(perception)
    
    // Calculate confidence for each branch
    for i := range branches {
        branches[i].Confidence = t.calculateBranchConfidence(&branches[i])
    }
    
    // Check if all branches have low confidence
    maxConfidence := 0.0
    for _, branch := range branches {
        if branch.Confidence > maxConfidence {
            maxConfidence = branch.Confidence
        }
    }
    
    // If all branches low confidence, retrieve information
    if maxConfidence < t.config.Thresholds.Proceed {
        log.Info("All reasoning branches have low confidence, retrieving information",
                 "max_confidence", maxConfidence)
        
        // Generate retrieval query
        query := t.generateRetrievalQuery("reason", perception, branches)
        
        // Retrieve information
        retrieval, err := t.retrieveInformation(query)
        if err != nil {
            log.Warn("Retrieval failed", "error", err)
        } else {
            // Regenerate branches with retrieved information
            enhancedBranches := t.regenerateBranches(perception, retrieval)
            
            // Recalculate confidence
            for i := range enhancedBranches {
                enhancedBranches[i].Confidence = t.calculateBranchConfidence(&enhancedBranches[i])
            }
            
            branches = enhancedBranches
            
            log.Info("Branches regenerated with retrieved information",
                     "new_max_confidence", maxBranchConfidence(branches))
        }
    }
    
    // Select best branch
    selectedBranch := t.selectBestBranch(branches)
    
    return &ReasoningResult{
        Branches: branches,
        SelectedBranch: selectedBranch,
        RetrievedInformation: retrieval,
    }, nil
}
```

## Complete Flow with Confidence and Retrieval

### Example: Complex Optimization Task

```
User: "Optimize the HandleTask function using industry best practices"
```

### Step 1: PERCEIVE

```
Capture screenshot and analyze code
    ↓
Calculate confidence:
  - Past experience: 0.0 (never optimized this before)
  - Pattern availability: 0.0 (no patterns in Neo4j)
  - Code understanding: 0.2 (complex code)
  - Strategy clarity: 0.0 (no clear strategy yet)
  - Risk assessment: 0.0 (unknown risk)
  Overall: 0.2 (LOW!)
    ↓
Decision: Retrieve information
    ↓
Search: "Go request handler optimization best practices"
    ↓
Found:
  - 5 web articles on handler optimization
  - 3 GitHub repos (go-cache, fasthttp, etc.)
  - 2 Stack Overflow answers
    ↓
Synthesize:
  - Key insight: "Caching is most effective"
  - Best practice: "Use TTL-based cache"
  - Common mistake: "Not invalidating on updates"
    ↓
Recalculate confidence:
  - Past experience: 0.15 (learned from articles)
  - Pattern availability: 0.20 (found patterns online)
  - Code understanding: 0.2 (unchanged)
  - Strategy clarity: 0.15 (clearer from examples)
  - Risk assessment: 0.1 (learned about risks)
  Overall: 0.8 (HIGH!)
    ↓
Decision: Proceed to REASON
```

### Step 2: REASON

```
Generate 3 branches using retrieved information:
    ↓
Branch 1: "Add caching with go-cache library"
  - Based on GitHub example
  - Confidence: 0.88
    ↓
Branch 2: "Optimize with sync.Map"
  - Based on Stack Overflow answer
  - Confidence: 0.75
    ↓
Branch 3: "Use fasthttp instead of net/http"
  - Based on web article
  - Confidence: 0.65
    ↓
All branches have acceptable confidence
No additional retrieval needed
    ↓
Select Branch 1 (highest confidence)
```

### Step 3: ACT

```
Execute caching implementation
    ↓
Confidence: 0.90 (high, based on clear example)
    ↓
No retrieval needed
    ↓
Apply changes to code
    ↓
Result: 35% performance improvement
```

### Step 4: REFLECT

```
Analyze results
    ↓
Confidence: 0.95 (high, clear success)
    ↓
Create pattern: "request_handler_caching"
    ↓
Store in Neo4j with:
  - Source: Retrieved from online best practices
  - Success rate: 1.0 (first use, successful)
  - Performance gain: 35%
    ↓
Future tasks can use this pattern (no retrieval needed!)
```

## Progress Indicators

### UI Shows Confidence and Retrieval Status

```
┌─────────────────────────────────────────────────────────────┐
│  Dynamic Thinking Progress                                  │
├─────────────────────────────────────────────────────────────┤
│  Phase: REASON                                              │
│  Confidence: 0.45 → 0.88 (boosted by retrieval)            │
│                                                             │
│  ⚠ Low confidence detected                                 │
│  🔍 Retrieving information...                              │
│                                                             │
│  Sources searched:                                          │
│  ✓ Web (5 articles)                                        │
│  ✓ GitHub (3 repos)                                        │
│  ✓ Stack Overflow (2 answers)                              │
│                                                             │
│  Key insights:                                              │
│  • Caching is most effective for read-heavy handlers       │
│  • Use TTL to prevent stale data                           │
│  • go-cache library recommended                            │
│                                                             │
│  Confidence boosted: +0.43                                  │
│  ✓ Proceeding with high confidence                         │
└─────────────────────────────────────────────────────────────┘
```

## Configuration

```yaml
# config/confidence_retrieval.yaml
confidence:
  tracking_enabled: true
  
  thresholds:
    perceive:
      proceed: 0.7
      retrieve: 0.5
      abort: 0.3
    reason:
      proceed: 0.75
      retrieve: 0.6
      abort: 0.4
    act:
      proceed: 0.8
      retrieve: 0.65
      abort: 0.5
    reflect:
      proceed: 0.7
      retrieve: 0.55
      abort: 0.4

retrieval:
  enabled: true
  
  sources:
    - web
    - github
    - stackoverflow
    - docs
    - youtube  # NEW! Critical for bleeding-edge content
  
  max_results_per_source: 5
  max_total_results: 20
  
  synthesis:
    enabled: true
    extract_code_examples: true
    extract_best_practices: true
    extract_common_mistakes: true
  
  caching:
    enabled: true
    ttl: 3600  # Cache retrieval results for 1 hour
    
  rate_limiting:
    max_requests_per_minute: 10
    max_requests_per_hour: 100
```

## Benefits

### 1. **Informed Decisions**
Agent doesn't guess - it retrieves industry best practices when uncertain

### 2. **Continuous Learning**
Retrieved information is synthesized and stored in Neo4j for future use

### 3. **Confidence Transparency**
User sees confidence scores and knows when agent is uncertain

### 4. **Adaptive Retrieval**
Only retrieves when needed, doesn't waste time on high-confidence tasks

### 5. **Multi-Source Knowledge**
Combines web articles, GitHub code, Stack Overflow answers, and official docs

## Summary

### Confidence Tracking
- ✅ Calculate confidence at each PRAR phase
- ✅ Consider multiple factors (experience, patterns, clarity, risk)
- ✅ Set thresholds for proceed/retrieve/abort decisions

### Online Retrieval
- ✅ Trigger when confidence below threshold
- ✅ Search multiple sources (web, GitHub, Stack Overflow, docs)
- ✅ Synthesize key insights, best practices, examples
- ✅ Boost confidence with retrieved information

### Integration
- ✅ New MCP tool: `retrieve_information`
- ✅ Enhanced `perceive`, `reason`, `act`, `reflect` with confidence
- ✅ Progress indicators show retrieval status
- ✅ Cached results for efficiency

### Result
The agent now **knows when it doesn't know** and **retrieves information** to fill knowledge gaps, resulting in **informed, high-confidence decisions**! 🎯

