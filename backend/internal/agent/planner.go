package agent

import (
	"context"
	"fmt"
	"strings"

	"agent-workspace/backend/internal/memory"
)

// Planner creates execution plans
type Planner struct {
	controller *Controller
}

// Plan represents an execution plan
type Plan struct {
	ID          string
	Goal        string
	Steps       []Step
	Tools       []string
	Context     string
	CreatedAt   string
}

// Step represents a plan step
type Step struct {
	ID          int
	Description string
	Tool        string
	Action      string
	Parameters  map[string]interface{}
	Expected    string
}

// NewPlanner creates a new planner
func NewPlanner(controller *Controller) *Planner {
	return &Planner{
		controller: controller,
	}
}

// CreatePlan creates an execution plan
func (p *Planner) CreatePlan(ctx context.Context, command string, taskMem *memory.TaskMemory) (*Plan, error) {
	// Get relevant context from long-term memory
	contextStr, err := p.controller.longTermMem.GetContext(ctx, command, 2000)
	if err != nil {
		contextStr = "No relevant context found"
	}

	// Generate plan using Gemma
	planText, err := p.controller.gemma.GeneratePlan(ctx, command, contextStr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// Parse plan text into structured plan
	plan := p.parsePlan(planText, command)

	// Store plan in task memory
	taskMem.SetContext("plan", plan)

	return plan, nil
}

// parsePlan parses plan text into a Plan struct
func (p *Planner) parsePlan(planText, command string) *Plan {
	plan := &Plan{
		ID:      fmt.Sprintf("plan_%d", generateTimestamp()),
		Goal:    command,
		Steps:   make([]Step, 0),
		Tools:   make([]string, 0),
		Context: planText,
	}

	// Parse sections
	lines := strings.Split(planText, "\n")
	currentSection := ""

	stepNum := 1
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "GOAL:") {
			plan.Goal = strings.TrimPrefix(line, "GOAL:")
			plan.Goal = strings.TrimSpace(plan.Goal)
			currentSection = "goal"
		} else if strings.HasPrefix(line, "STEPS:") {
			currentSection = "steps"
		} else if strings.HasPrefix(line, "TOOLS:") {
			currentSection = "tools"
		} else if strings.HasPrefix(line, "EXPECTED_OUTCOME:") {
			currentSection = "expected"
		} else if line != "" {
			switch currentSection {
			case "steps":
				if strings.Contains(line, ".") {
					// Parse step
					parts := strings.SplitN(line, ".", 2)
					if len(parts) == 2 {
						description := strings.TrimSpace(parts[1])
						tool := p.detectTool(description)

						step := Step{
							ID:          stepNum,
							Description: description,
							Tool:        tool,
							Action:      p.extractAction(description),
							Parameters:  make(map[string]interface{}),
						}

						plan.Steps = append(plan.Steps, step)
						stepNum++

						// Add tool to tools list if not already there
						if !contains(plan.Tools, tool) {
							plan.Tools = append(plan.Tools, tool)
						}
					}
				}
			case "tools":
				// Parse tools
				tools := strings.Split(line, ",")
				for _, tool := range tools {
					tool = strings.TrimSpace(tool)
					if tool != "" && !contains(plan.Tools, tool) {
						plan.Tools = append(plan.Tools, tool)
					}
				}
			}
		}
	}

	// If no steps were parsed, create a default step
	if len(plan.Steps) == 0 {
		plan.Steps = append(plan.Steps, Step{
			ID:          1,
			Description: command,
			Tool:        "terminal",
			Action:      command,
			Parameters:  make(map[string]interface{}),
		})
		plan.Tools = []string{"terminal"}
	}

	return plan
}

// detectTool detects which tool to use based on description
func (p *Planner) detectTool(description string) string {
	lower := strings.ToLower(description)

	if strings.Contains(lower, "browser") || strings.Contains(lower, "navigate") ||
		strings.Contains(lower, "click") || strings.Contains(lower, "website") ||
		strings.Contains(lower, "url") {
		return "browser"
	}

	if strings.Contains(lower, "mcp") || strings.Contains(lower, "tool") ||
		strings.Contains(lower, "perceive") || strings.Contains(lower, "reason") {
		return "mcp"
	}

	return "terminal"
}

// extractAction extracts the action from description
func (p *Planner) extractAction(description string) string {
	// Simple extraction - just return the description
	// In a real implementation, this would parse more intelligently
	return description
}

// RevisePlan revises a plan based on feedback
func (p *Planner) RevisePlan(ctx context.Context, plan *Plan, feedback string) (*Plan, error) {
	// Generate revised plan
	prompt := fmt.Sprintf("Original plan:\n%s\n\nFeedback:\n%s\n\nRevise the plan.", plan.Context, feedback)

	revisedText, err := p.controller.gemma.GeneratePlan(ctx, plan.Goal, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to revise plan: %w", err)
	}

	// Parse revised plan
	revisedPlan := p.parsePlan(revisedText, plan.Goal)
	revisedPlan.ID = fmt.Sprintf("plan_%d_revised", generateTimestamp())

	return revisedPlan, nil
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateTimestamp() int64 {
	// TODO: Use time.Now().Unix()
	return 1234567890
}

