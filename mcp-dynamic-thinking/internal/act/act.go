package act

import (
	"context"
	"fmt"
	"time"
)

// Actor handles the Act phase of PRAR
type Actor struct {
	executions map[string]*Execution
}

// Execution represents an action execution
type Execution struct {
	ID          string
	TaskID      string
	ActionPlan  map[string]interface{}
	Steps       []ExecutionStep
	Status      string
	StartTime   time.Time
	EndTime     time.Time
	Result      map[string]interface{}
}

// ExecutionStep represents a single step execution
type ExecutionStep struct {
	ID          string
	Description string
	Status      string
	Result      interface{}
	Error       string
	StartTime   time.Time
	EndTime     time.Time
}

// NewActor creates a new actor
func NewActor() *Actor {
	return &Actor{
		executions: make(map[string]*Execution),
	}
}

// Act executes an action plan with monitoring
func (a *Actor) Act(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	taskID, ok := args["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("task_id is required")
	}

	actionPlan, ok := args["action_plan"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("action_plan is required")
	}

	// Create execution
	executionID := fmt.Sprintf("execution_%d", time.Now().UnixNano())
	
	execution := &Execution{
		ID:         executionID,
		TaskID:     taskID,
		ActionPlan: actionPlan,
		Steps:      make([]ExecutionStep, 0),
		Status:     "running",
		StartTime:  time.Now(),
	}

	// Execute plan
	if err := a.executePlan(ctx, execution); err != nil {
		execution.Status = "failed"
		execution.EndTime = time.Now()
		a.executions[executionID] = execution
		
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	execution.Status = "completed"
	execution.EndTime = time.Now()

	// Store execution
	a.executions[executionID] = execution

	return map[string]interface{}{
		"execution_id": executionID,
		"task_id":      taskID,
		"status":       execution.Status,
		"steps":        a.stepsToMap(execution.Steps),
		"duration_ms":  execution.EndTime.Sub(execution.StartTime).Milliseconds(),
		"result":       execution.Result,
	}, nil
}

// executePlan executes the action plan
func (a *Actor) executePlan(ctx context.Context, execution *Execution) error {
	// Extract steps from action plan
	steps, ok := execution.ActionPlan["steps"].([]interface{})
	if !ok {
		// If no steps, create a single step
		steps = []interface{}{"Execute action"}
	}

	// Execute each step
	for i, stepDesc := range steps {
		stepID := fmt.Sprintf("step_%d", i+1)
		
		step := ExecutionStep{
			ID:          stepID,
			Description: fmt.Sprintf("%v", stepDesc),
			Status:      "running",
			StartTime:   time.Now(),
		}

		// Execute step
		result, err := a.executeStep(ctx, step.Description)
		step.EndTime = time.Now()

		if err != nil {
			step.Status = "failed"
			step.Error = err.Error()
			execution.Steps = append(execution.Steps, step)
			return fmt.Errorf("step %s failed: %w", stepID, err)
		}

		step.Status = "completed"
		step.Result = result
		execution.Steps = append(execution.Steps, step)
	}

	// Set execution result
	execution.Result = map[string]interface{}{
		"success":      true,
		"steps_completed": len(execution.Steps),
	}

	return nil
}

// executeStep executes a single step
func (a *Actor) executeStep(ctx context.Context, description string) (interface{}, error) {
	// Simulate step execution
	// In a real implementation, this would call browser, terminal, or MCP tools
	
	time.Sleep(100 * time.Millisecond) // Simulate work

	return map[string]interface{}{
		"description": description,
		"executed":    true,
		"timestamp":   time.Now().Format(time.RFC3339),
	}, nil
}

// stepsToMap converts steps to map format
func (a *Actor) stepsToMap(steps []ExecutionStep) []map[string]interface{} {
	result := make([]map[string]interface{}, len(steps))

	for i, step := range steps {
		result[i] = map[string]interface{}{
			"id":          step.ID,
			"description": step.Description,
			"status":      step.Status,
			"result":      step.Result,
			"error":       step.Error,
			"duration_ms": step.EndTime.Sub(step.StartTime).Milliseconds(),
		}
	}

	return result
}

// GetExecution retrieves an execution by ID
func (a *Actor) GetExecution(executionID string) (*Execution, error) {
	execution, exists := a.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution %s not found", executionID)
	}

	return execution, nil
}

// MonitorExecution monitors an ongoing execution
func (a *Actor) MonitorExecution(executionID string) (map[string]interface{}, error) {
	execution, err := a.GetExecution(executionID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"execution_id":    execution.ID,
		"status":          execution.Status,
		"steps_completed": len(execution.Steps),
		"current_step":    a.getCurrentStep(execution),
	}, nil
}

// getCurrentStep gets the current step being executed
func (a *Actor) getCurrentStep(execution *Execution) map[string]interface{} {
	if len(execution.Steps) == 0 {
		return nil
	}

	lastStep := execution.Steps[len(execution.Steps)-1]
	return map[string]interface{}{
		"id":          lastStep.ID,
		"description": lastStep.Description,
		"status":      lastStep.Status,
	}
}

