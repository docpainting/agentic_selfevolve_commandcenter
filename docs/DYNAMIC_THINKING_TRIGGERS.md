# How Dynamic Thinking is Triggered

## Overview

The Dynamic Thinking MCP server is triggered in **two main ways**:

1. **Automatically** - When the main agent detects complex tasks or low confidence
2. **Explicitly** - When the agent's LLM (Gemma 3) chooses to use the thinking tools

Let's break down both approaches.

---

## Approach 1: Automatic Triggering (Recommended)

The main agent automatically triggers dynamic thinking when certain conditions are met.

### Trigger Conditions

```go
// backend/internal/agent/controller.go
func (a *AgentController) HandleTask(task *Task) error {
    // Analyze task complexity
    complexity := a.analyzeComplexity(task)
    
    // Trigger dynamic thinking if:
    if complexity.Score > 0.7 ||                    // High complexity
       task.RequiresSelfModification ||             // Needs to modify own code
       task.Type == "code_optimization" ||          // Optimization task
       task.Type == "self_improvement" ||           // Self-improvement task
       a.lastAttemptFailed(task) {                  // Previous attempt failed
        
        // Use dynamic thinking loop
        return a.handleWithDynamicThinking(task)
    }
    
    // Simple task - use standard approach
    return a.handleSimpleTask(task)
}
```

### Complexity Analysis

```go
func (a *AgentController) analyzeComplexity(task *Task) ComplexityScore {
    score := 0.0
    
    // Check for complexity indicators
    if strings.Contains(task.Description, "optimize") {
        score += 0.3
    }
    if strings.Contains(task.Description, "improve") {
        score += 0.3
    }
    if strings.Contains(task.Description, "refactor") {
        score += 0.4
    }
    if task.RequiresMultipleSteps {
        score += 0.2
    }
    if task.RequiresCodeAnalysis {
        score += 0.3
    }
    if task.AffectsOwnCode {
        score += 0.5  // Self-modification is complex
    }
    
    return ComplexityScore{
        Score: min(score, 1.0),
        Reason: "Task requires multi-step reasoning and code analysis",
    }
}
```

### Full Automatic Flow

```go
func (a *AgentController) handleWithDynamicThinking(task *Task) error {
    log.Info("Task complexity high, using dynamic thinking", 
             "task_id", task.ID, 
             "complexity", task.Complexity)
    
    // 1. PERCEIVE
    perception, err := a.mcpClient.CallTool("dynamic-thinking", "perceive", map[string]interface{}{
        "task_id": task.ID,
        "goal": task.Description,
        "capture_screenshot": true,
        "analyze_code_mirror": task.AffectsOwnCode,
        "confidence_threshold": 0.7,
    })
    if err != nil {
        return fmt.Errorf("perception failed: %w", err)
    }
    
    // Check confidence
    if perception.Confidence < 0.7 {
        log.Warn("Low perception confidence, gathering more context")
        // Trigger additional perception or ask user for clarification
    }
    
    // 2. REASON
    reasoning, err := a.mcpClient.CallTool("dynamic-thinking", "reason", map[string]interface{}{
        "task_id": task.ID,
        "perception_id": perception.ID,
        "goal": task.Description,
        "num_branches": 3,
        "use_past_strategies": true,
    })
    if err != nil {
        return fmt.Errorf("reasoning failed: %w", err)
    }
    
    // 3. ACT
    execution, err := a.mcpClient.CallTool("dynamic-thinking", "act", map[string]interface{}{
        "task_id": task.ID,
        "reasoning_id": reasoning.ID,
        "action_plan": reasoning.ActionPlan,
        "enable_watchdog": true,
        "allow_dynamic_adjustment": true,
    })
    if err != nil {
        return fmt.Errorf("execution failed: %w", err)
    }
    
    // 4. REFLECT
    reflection, err := a.mcpClient.CallTool("dynamic-thinking", "reflect", map[string]interface{}{
        "task_id": task.ID,
        "execution_id": execution.ID,
        "original_goal": task.Description,
        "enable_evolution": true,
    })
    if err != nil {
        return fmt.Errorf("reflection failed: %w", err)
    }
    
    // 5. Check if we need to loop back
    if reflection.Decision == "replan" {
        log.Info("Reflection suggests replanning, looping back")
        return a.handleWithDynamicThinking(task)  // Recursive call with new insights
    }
    
    log.Info("Dynamic thinking completed successfully", 
             "improvements", reflection.Evolutions)
    
    return nil
}
```

---

## Approach 2: LLM-Driven Triggering (More Flexible)

The agent's LLM (Gemma 3) has access to dynamic thinking tools and can choose when to use them.

### Tool Registration

```go
// backend/internal/agent/tools.go
func (a *AgentController) RegisterTools() {
    // ... other tools ...
    
    // Register dynamic thinking tools
    a.toolRegistry.RegisterMCP("dynamic-thinking", "perceive", Tool{
        Name: "dynamic_thinking_perceive",
        Description: "Analyze current environment and code state with vision. Use this when you need to understand the current situation before making decisions.",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "task_id": map[string]string{
                    "type": "string",
                    "description": "Unique task identifier",
                },
                "goal": map[string]string{
                    "type": "string",
                    "description": "What you're trying to understand",
                },
                "capture_screenshot": map[string]interface{}{
                    "type": "boolean",
                    "description": "Whether to capture screenshot",
                    "default": true,
                },
                "analyze_code_mirror": map[string]interface{}{
                    "type": "boolean",
                    "description": "Whether to analyze your own code structure",
                    "default": false,
                },
            },
            "required": []string{"task_id", "goal"},
        },
    })
    
    a.toolRegistry.RegisterMCP("dynamic-thinking", "reason", Tool{
        Name: "dynamic_thinking_reason",
        Description: "Generate and evaluate multiple reasoning approaches. Use this when facing a complex problem that might have multiple solutions.",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "task_id": map[string]string{"type": "string"},
                "perception_id": map[string]string{"type": "string"},
                "goal": map[string]string{"type": "string"},
                "num_branches": map[string]interface{}{
                    "type": "integer",
                    "description": "Number of reasoning approaches to explore",
                    "default": 3,
                },
            },
            "required": []string{"task_id", "perception_id", "goal"},
        },
    })
    
    a.toolRegistry.RegisterMCP("dynamic-thinking", "act", Tool{
        Name: "dynamic_thinking_act",
        Description: "Execute an action plan with monitoring and dynamic adjustment. Use this to execute complex multi-step plans.",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "task_id": map[string]string{"type": "string"},
                "reasoning_id": map[string]string{"type": "string"},
                "action_plan": map[string]interface{}{
                    "type": "object",
                    "description": "The plan to execute",
                },
            },
            "required": []string{"task_id", "reasoning_id", "action_plan"},
        },
    })
    
    a.toolRegistry.RegisterMCP("dynamic-thinking", "reflect", Tool{
        Name: "dynamic_thinking_reflect",
        Description: "Analyze execution results and evolve your strategies. Use this after completing a task to learn and improve.",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "task_id": map[string]string{"type": "string"},
                "execution_id": map[string]string{"type": "string"},
                "original_goal": map[string]string{"type": "string"},
            },
            "required": []string{"task_id", "execution_id", "original_goal"},
        },
    })
}
```

### LLM Decides When to Use

```
User: "Optimize the HandleTask function to be 30% faster"
    ↓
Gemma 3 thinks:
    "This is a complex optimization task. I should use dynamic thinking
     to explore multiple approaches and learn from the results."
    ↓
Gemma 3 returns tool call:
{
  "tool_calls": [{
    "name": "dynamic_thinking_perceive",
    "arguments": {
      "task_id": "opt-123",
      "goal": "Understand current HandleTask performance",
      "capture_screenshot": true,
      "analyze_code_mirror": true
    }
  }]
}
    ↓
Agent executes tool → Returns perception
    ↓
Gemma 3 receives perception, decides next step:
{
  "tool_calls": [{
    "name": "dynamic_thinking_reason",
    "arguments": {
      "task_id": "opt-123",
      "perception_id": "perc-456",
      "goal": "Find best optimization approach",
      "num_branches": 3
    }
  }]
}
    ↓
And so on...
```

---

## Hybrid Approach (Best of Both Worlds)

Combine automatic and LLM-driven triggering for maximum flexibility.

### Configuration

```yaml
# config/dynamic_thinking.yaml
triggering:
  mode: hybrid  # auto, llm, or hybrid
  
  auto_triggers:
    complexity_threshold: 0.7
    enable_for_task_types:
      - code_optimization
      - self_improvement
      - refactoring
      - debugging_complex
    enable_on_failure: true
    enable_on_low_confidence: true
  
  llm_access:
    expose_all_tools: true
    require_explicit_call: false  # LLM can use tools freely
    suggest_on_complexity: true   # Hint to LLM when task is complex
```

### Implementation

```go
func (a *AgentController) HandleTask(task *Task) error {
    config := a.config.DynamicThinking
    
    // Check if auto-trigger conditions met
    shouldAutoTrigger := a.shouldAutoTrigger(task, config)
    
    if config.Mode == "auto" && shouldAutoTrigger {
        // Force dynamic thinking
        return a.handleWithDynamicThinking(task)
    }
    
    if config.Mode == "llm" || config.Mode == "hybrid" {
        // Give LLM access to tools
        tools := a.toolRegistry.GetAllTools()
        
        // If hybrid mode and should auto-trigger, hint to LLM
        systemPrompt := a.buildSystemPrompt()
        if config.Mode == "hybrid" && shouldAutoTrigger {
            systemPrompt += "\n\nNOTE: This task is complex. Consider using dynamic_thinking_* tools for better results."
        }
        
        // Let LLM decide
        response, err := a.llm.Generate(LLMRequest{
            Messages: []Message{
                {Role: "system", Content: systemPrompt},
                {Role: "user", Content: task.Description},
            },
            Tools: tools,
        })
        
        // Execute whatever LLM decides
        return a.executeResponse(response)
    }
    
    // Default: simple task handling
    return a.handleSimpleTask(task)
}
```

---

## Specific Trigger Scenarios

### Scenario 1: Self-Modification Task

```
User: "Add a new feature to the agent controller"
    ↓
Agent detects: task.AffectsOwnCode = true
    ↓
Auto-trigger: handleWithDynamicThinking()
    ↓
PERCEIVE: Analyze current controller code via Neo4j mirror
REASON: Explore 3 approaches to add feature
ACT: Implement chosen approach
REFLECT: Critique implementation, evolve patterns
```

### Scenario 2: Failed Task Retry

```
First attempt: Standard approach fails
    ↓
Agent detects: a.lastAttemptFailed(task) = true
    ↓
Auto-trigger: handleWithDynamicThinking()
    ↓
PERCEIVE: Analyze why previous attempt failed
REASON: Generate alternative approaches
ACT: Execute with dynamic adjustment
REFLECT: Learn from failure, update strategies
```

### Scenario 3: User Explicitly Requests Thinking

```
User: "Think carefully about how to optimize this code"
    ↓
Gemma 3 detects: "think carefully" = use dynamic thinking
    ↓
LLM calls: dynamic_thinking_perceive
    ↓
Full PRAR loop executes
```

### Scenario 4: Low Confidence Detection

```
Agent attempts task with standard approach
    ↓
Generates code but confidence score = 0.5 (low)
    ↓
Auto-trigger: handleWithDynamicThinking()
    ↓
PERCEIVE: Gather more context
REASON: Explore multiple approaches
ACT: Execute with higher confidence
REFLECT: Learn what caused low confidence
```

---

## Monitoring and Control

### Dashboard Indicators

The UI shows when dynamic thinking is active:

```
┌─────────────────────────────────────┐
│  Dynamic Thinking Status            │
├─────────────────────────────────────┤
│  🧠 Active                          │
│  Phase: REASON (Branch 2/3)         │
│  Confidence: 0.85                   │
│  Task: Optimize HandleTask          │
│                                     │
│  Branches:                          │
│  ✓ Add caching (0.9)                │
│  ○ Parallel execution (0.7)         │
│  ○ Optimize query (0.8)             │
└─────────────────────────────────────┘
```

### Manual Override

User can force or disable dynamic thinking:

```
User: "Use dynamic thinking for this task"
    ↓
Agent: Forces PRAR loop regardless of complexity

User: "Don't use dynamic thinking, just do it"
    ↓
Agent: Skips dynamic thinking, uses standard approach
```

---

## Performance Considerations

### When to Use Dynamic Thinking

**Use it for:**
- ✅ Complex multi-step tasks
- ✅ Self-modification (editing own code)
- ✅ Optimization and refactoring
- ✅ Tasks that failed with standard approach
- ✅ Learning opportunities (new patterns)

**Don't use it for:**
- ❌ Simple file operations
- ❌ Basic CRUD operations
- ❌ Straightforward terminal commands
- ❌ Tasks with clear single solution
- ❌ Time-sensitive operations (adds overhead)

### Overhead

Dynamic thinking adds:
- **Time**: 2-5x longer (due to multi-branch reasoning)
- **LLM calls**: 4-6 calls per loop (perceive, reason, act, reflect)
- **Memory**: Stores branches and traces in ChromeM/Neo4j

**But gains:**
- **Quality**: 30-50% better solutions
- **Learning**: Every task improves future performance
- **Adaptability**: Handles complex/ambiguous tasks

---

## Configuration Examples

### Conservative (Auto-trigger only for critical tasks)

```yaml
triggering:
  mode: hybrid
  auto_triggers:
    complexity_threshold: 0.9  # Very high threshold
    enable_for_task_types:
      - self_improvement
    enable_on_failure: true
```

### Aggressive (Use dynamic thinking often)

```yaml
triggering:
  mode: hybrid
  auto_triggers:
    complexity_threshold: 0.5  # Low threshold
    enable_for_task_types:
      - code_optimization
      - self_improvement
      - refactoring
      - debugging_complex
      - api_design
    enable_on_failure: true
    enable_on_low_confidence: true
```

### LLM-Only (Let Gemma 3 decide everything)

```yaml
triggering:
  mode: llm
  llm_access:
    expose_all_tools: true
    require_explicit_call: false
    suggest_on_complexity: true
```

---

## Summary

### Triggering Methods

1. **Automatic** - Agent detects complexity/failure and triggers PRAR loop
2. **LLM-Driven** - Gemma 3 chooses when to use thinking tools
3. **Hybrid** - Automatic triggers + LLM has access to tools
4. **Manual** - User explicitly requests dynamic thinking

### Recommended Setup

**Start with Hybrid mode:**
- Auto-trigger for self-modification and failures
- Give LLM access to tools for flexibility
- Hint to LLM when task is complex
- Monitor usage and adjust thresholds

### Key Points

- ✅ Dynamic thinking is **optional** - not used for every task
- ✅ Triggered based on **complexity, task type, or failure**
- ✅ LLM can **choose** to use it via tool calling
- ✅ Adds overhead but **significantly improves quality**
- ✅ **Learns and improves** with every use
- ✅ User can **force or disable** it

This makes dynamic thinking a **powerful tool** that's used when needed, not a bottleneck on every operation!

