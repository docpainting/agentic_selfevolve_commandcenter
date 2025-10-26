package memory

import (
	"context"
	"fmt"
	"time"
)

// MemoryManager manages short-term and long-term memory
type MemoryManager struct {
	shortTerm map[string]*TaskMemory
	longTerm  map[string]interface{}
}

// TaskMemory represents short-term memory for a task
type TaskMemory struct {
	TaskID      string
	Perceptions []interface{}
	Reasoning   []interface{}
	Actions     []interface{}
	Reflections []interface{}
	Context     map[string]interface{}
	CreatedAt   time.Time
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		shortTerm: make(map[string]*TaskMemory),
		longTerm:  make(map[string]interface{}),
	}
}

// GetShortTermMemory retrieves task context from short-term memory
func (m *MemoryManager) GetShortTermMemory(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	taskID, ok := args["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("task_id is required")
	}

	memory, exists := m.shortTerm[taskID]
	if !exists {
		// Create new memory if it doesn't exist
		memory = &TaskMemory{
			TaskID:      taskID,
			Perceptions: make([]interface{}, 0),
			Reasoning:   make([]interface{}, 0),
			Actions:     make([]interface{}, 0),
			Reflections: make([]interface{}, 0),
			Context:     make(map[string]interface{}),
			CreatedAt:   time.Now(),
		}
		m.shortTerm[taskID] = memory
	}

	return map[string]interface{}{
		"task_id":      memory.TaskID,
		"perceptions":  memory.Perceptions,
		"reasoning":    memory.Reasoning,
		"actions":      memory.Actions,
		"reflections":  memory.Reflections,
		"context":      memory.Context,
		"created_at":   memory.CreatedAt.Format(time.RFC3339),
		"age_seconds":  time.Since(memory.CreatedAt).Seconds(),
	}, nil
}

// ClearShortTermMemory clears task memory after completion
func (m *MemoryManager) ClearShortTermMemory(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	taskID, ok := args["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("task_id is required")
	}

	archiveToLongTerm := false
	if archive, ok := args["archive_to_long_term"].(bool); ok {
		archiveToLongTerm = archive
	}

	memory, exists := m.shortTerm[taskID]
	if !exists {
		return nil, fmt.Errorf("task memory %s not found", taskID)
	}

	// Archive to long-term if requested
	if archiveToLongTerm {
		m.archiveToLongTerm(memory)
	}

	// Clear short-term memory
	delete(m.shortTerm, taskID)

	return map[string]interface{}{
		"task_id":  taskID,
		"cleared":  true,
		"archived": archiveToLongTerm,
	}, nil
}

// QueryStrategies finds relevant strategies from Neo4j
func (m *MemoryManager) QueryStrategies(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query is required")
	}

	// In a real implementation, this would query Neo4j
	// For now, return mock strategies
	strategies := []map[string]interface{}{
		{
			"name":         "direct_approach",
			"description":  "Direct execution for straightforward tasks",
			"success_rate": 0.85,
			"use_cases":    []string{"clear_goals", "simple_actions"},
		},
		{
			"name":         "exploratory_approach",
			"description":  "Explore options before deciding",
			"success_rate": 0.75,
			"use_cases":    []string{"ambiguous_goals", "multiple_options"},
		},
		{
			"name":         "cautious_approach",
			"description":  "Validate each step carefully",
			"success_rate": 0.90,
			"use_cases":    []string{"high_risk", "critical_operations"},
		},
	}

	return map[string]interface{}{
		"query":      query,
		"strategies": strategies,
		"count":      len(strategies),
	}, nil
}

// GetExecutionTrace exports training data for completed task
func (m *MemoryManager) GetExecutionTrace(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	taskID, ok := args["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("task_id is required")
	}

	memory, exists := m.shortTerm[taskID]
	if !exists {
		return nil, fmt.Errorf("task memory %s not found", taskID)
	}

	// Create execution trace
	trace := map[string]interface{}{
		"task_id":     memory.TaskID,
		"created_at":  memory.CreatedAt.Format(time.RFC3339),
		"duration":    time.Since(memory.CreatedAt).Seconds(),
		"perceptions": memory.Perceptions,
		"reasoning":   memory.Reasoning,
		"actions":     memory.Actions,
		"reflections": memory.Reflections,
		"context":     memory.Context,
		"trace_type":  "prar_loop",
		"version":     "1.0",
	}

	return trace, nil
}

// archiveToLongTerm archives task memory to long-term storage
func (m *MemoryManager) archiveToLongTerm(memory *TaskMemory) {
	archiveKey := fmt.Sprintf("archive_%s_%d", memory.TaskID, time.Now().Unix())
	
	m.longTerm[archiveKey] = map[string]interface{}{
		"task_id":     memory.TaskID,
		"perceptions": memory.Perceptions,
		"reasoning":   memory.Reasoning,
		"actions":     memory.Actions,
		"reflections": memory.Reflections,
		"context":     memory.Context,
		"created_at":  memory.CreatedAt.Format(time.RFC3339),
		"archived_at": time.Now().Format(time.RFC3339),
	}
}

// AddPerception adds a perception to task memory
func (m *MemoryManager) AddPerception(taskID string, perception interface{}) error {
	memory, exists := m.shortTerm[taskID]
	if !exists {
		memory = &TaskMemory{
			TaskID:      taskID,
			Perceptions: make([]interface{}, 0),
			Reasoning:   make([]interface{}, 0),
			Actions:     make([]interface{}, 0),
			Reflections: make([]interface{}, 0),
			Context:     make(map[string]interface{}),
			CreatedAt:   time.Now(),
		}
		m.shortTerm[taskID] = memory
	}

	memory.Perceptions = append(memory.Perceptions, perception)
	return nil
}

// AddReasoning adds reasoning to task memory
func (m *MemoryManager) AddReasoning(taskID string, reasoning interface{}) error {
	memory, exists := m.shortTerm[taskID]
	if !exists {
		return fmt.Errorf("task memory %s not found", taskID)
	}

	memory.Reasoning = append(memory.Reasoning, reasoning)
	return nil
}

// AddAction adds an action to task memory
func (m *MemoryManager) AddAction(taskID string, action interface{}) error {
	memory, exists := m.shortTerm[taskID]
	if !exists {
		return fmt.Errorf("task memory %s not found", taskID)
	}

	memory.Actions = append(memory.Actions, action)
	return nil
}

// AddReflection adds a reflection to task memory
func (m *MemoryManager) AddReflection(taskID string, reflection interface{}) error {
	memory, exists := m.shortTerm[taskID]
	if !exists {
		return fmt.Errorf("task memory %s not found", taskID)
	}

	memory.Reflections = append(memory.Reflections, reflection)
	return nil
}

