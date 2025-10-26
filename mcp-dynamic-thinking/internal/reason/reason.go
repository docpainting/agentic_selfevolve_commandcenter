package reason

import (
	"context"
	"fmt"
	"time"
)

// Reasoner handles the Reason phase of PRAR
type Reasoner struct {
	branches map[string][]*ReasoningBranch
}

// ReasoningBranch represents a reasoning path
type ReasoningBranch struct {
	ID           string
	TaskID       string
	PerceptionID string
	Strategy     string
	Steps        []string
	Confidence   float64
	Reasoning    string
	Selected     bool
	Timestamp    time.Time
}

// NewReasoner creates a new reasoner
func NewReasoner() *Reasoner {
	return &Reasoner{
		branches: make(map[string][]*ReasoningBranch),
	}
}

// Reason generates and evaluates reasoning branches
func (r *Reasoner) Reason(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	taskID, ok := args["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("task_id is required")
	}

	perceptionID, ok := args["perception_id"].(string)
	if !ok {
		return nil, fmt.Errorf("perception_id is required")
	}

	numBranches := 3
	if n, ok := args["num_branches"].(float64); ok {
		numBranches = int(n)
	}

	// Generate reasoning branches
	branches := r.generateBranches(taskID, perceptionID, numBranches)

	// Evaluate branches
	bestBranch := r.evaluateBranches(branches)
	bestBranch.Selected = true

	// Store branches
	r.branches[taskID] = branches

	return map[string]interface{}{
		"task_id":       taskID,
		"perception_id": perceptionID,
		"branches":      r.branchesToMap(branches),
		"selected_branch": map[string]interface{}{
			"id":         bestBranch.ID,
			"strategy":   bestBranch.Strategy,
			"confidence": bestBranch.Confidence,
		},
		"action_plan": r.createActionPlan(bestBranch),
	}, nil
}

// generateBranches generates multiple reasoning branches
func (r *Reasoner) generateBranches(taskID, perceptionID string, count int) []*ReasoningBranch {
	branches := make([]*ReasoningBranch, 0, count)

	strategies := []string{
		"direct_approach",
		"exploratory_approach",
		"cautious_approach",
	}

	for i := 0; i < count && i < len(strategies); i++ {
		branchID := fmt.Sprintf("branch_%d_%d", time.Now().UnixNano(), i)
		
		branch := &ReasoningBranch{
			ID:           branchID,
			TaskID:       taskID,
			PerceptionID: perceptionID,
			Strategy:     strategies[i],
			Steps:        r.generateSteps(strategies[i]),
			Confidence:   r.calculateConfidence(strategies[i]),
			Reasoning:    r.generateReasoning(strategies[i]),
			Selected:     false,
			Timestamp:    time.Now(),
		}

		branches = append(branches, branch)
	}

	return branches
}

// generateSteps generates steps for a strategy
func (r *Reasoner) generateSteps(strategy string) []string {
	switch strategy {
	case "direct_approach":
		return []string{
			"Identify target element",
			"Execute action immediately",
			"Verify result",
		}
	case "exploratory_approach":
		return []string{
			"Explore available options",
			"Analyze each option",
			"Choose best option",
			"Execute action",
			"Verify result",
		}
	case "cautious_approach":
		return []string{
			"Analyze current state",
			"Identify potential risks",
			"Plan safe action",
			"Execute with validation",
			"Monitor for errors",
			"Verify result",
		}
	default:
		return []string{"Execute action"}
	}
}

// calculateConfidence calculates confidence for a strategy
func (r *Reasoner) calculateConfidence(strategy string) float64 {
	switch strategy {
	case "direct_approach":
		return 0.8
	case "exploratory_approach":
		return 0.7
	case "cautious_approach":
		return 0.9
	default:
		return 0.5
	}
}

// generateReasoning generates reasoning explanation
func (r *Reasoner) generateReasoning(strategy string) string {
	switch strategy {
	case "direct_approach":
		return "Direct approach is fastest when the goal is clear and the action is straightforward."
	case "exploratory_approach":
		return "Exploratory approach is best when we need to understand the environment before acting."
	case "cautious_approach":
		return "Cautious approach minimizes risk by validating each step before proceeding."
	default:
		return "Standard approach for unknown situations."
	}
}

// evaluateBranches selects the best branch
func (r *Reasoner) evaluateBranches(branches []*ReasoningBranch) *ReasoningBranch {
	if len(branches) == 0 {
		return nil
	}

	bestBranch := branches[0]
	for _, branch := range branches[1:] {
		if branch.Confidence > bestBranch.Confidence {
			bestBranch = branch
		}
	}

	return bestBranch
}

// createActionPlan creates an action plan from a branch
func (r *Reasoner) createActionPlan(branch *ReasoningBranch) map[string]interface{} {
	return map[string]interface{}{
		"branch_id": branch.ID,
		"strategy":  branch.Strategy,
		"steps":     branch.Steps,
		"reasoning": branch.Reasoning,
	}
}

// branchesToMap converts branches to map format
func (r *Reasoner) branchesToMap(branches []*ReasoningBranch) []map[string]interface{} {
	result := make([]map[string]interface{}, len(branches))

	for i, branch := range branches {
		result[i] = map[string]interface{}{
			"id":         branch.ID,
			"strategy":   branch.Strategy,
			"steps":      branch.Steps,
			"confidence": branch.Confidence,
			"reasoning":  branch.Reasoning,
			"selected":   branch.Selected,
		}
	}

	return result
}

// GetBranches retrieves branches for a task
func (r *Reasoner) GetBranches(taskID string) ([]*ReasoningBranch, error) {
	branches, exists := r.branches[taskID]
	if !exists {
		return nil, fmt.Errorf("no branches found for task %s", taskID)
	}

	return branches, nil
}

