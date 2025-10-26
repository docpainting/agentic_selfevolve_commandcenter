package agent

import (
	"context"
	"fmt"

	"agent-workspace/backend/internal/memory"
)

// Executor executes plans
type Executor struct {
	controller *Controller
}

// NewExecutor creates a new executor
func NewExecutor(controller *Controller) *Executor {
	return &Executor{
		controller: controller,
	}
}

// ExecutePlan executes a plan
func (e *Executor) ExecutePlan(ctx context.Context, plan *Plan, taskMem *memory.TaskMemory) error {
	// Execute each step
	for _, step := range plan.Steps {
		if err := e.ExecuteStep(ctx, step, taskMem); err != nil {
			// Store failure
			taskMem.AddAction(step.Tool, step.Action, step.Parameters, nil, false, err.Error())

			// Generate reflection on failure
			reflection, _ := e.controller.gemma.GenerateReflection(ctx, step.Description, err.Error(), false)
			taskMem.AddReflection(step.Action, reflection, []string{}, []string{})

			return fmt.Errorf("step %d failed: %w", step.ID, err)
		}

		// Store success
		taskMem.AddAction(step.Tool, step.Action, step.Parameters, "success", true, "")
	}

	// Generate final reflection
	reflection, _ := e.controller.gemma.GenerateReflection(ctx, plan.Goal, "Plan completed successfully", true)
	taskMem.AddReflection(plan.ID, reflection, []string{}, []string{})

	// Store in long-term memory
	e.controller.longTermMem.StoreAction(ctx, plan.Goal, "Completed successfully", true)

	return nil
}

// ExecuteStep executes a single step
func (e *Executor) ExecuteStep(ctx context.Context, step Step, taskMem *memory.TaskMemory) error {
	switch step.Tool {
	case "browser":
		return e.executeBrowserStep(ctx, step, taskMem)
	case "terminal":
		return e.executeTerminalStep(ctx, step, taskMem)
	case "mcp":
		return e.executeMCPStep(ctx, step, taskMem)
	default:
		return fmt.Errorf("unknown tool: %s", step.Tool)
	}
}

// executeBrowserStep executes a browser step
func (e *Executor) executeBrowserStep(ctx context.Context, step Step, taskMem *memory.TaskMemory) error {
	action := step.Action

	// Capture screenshot
	screenshot, err := e.controller.browserMgr.CaptureScreenshot(taskMem.TaskID)
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Detect elements
	elements, err := e.controller.browserMgr.AnalyzeScreenshot(models.VisionAnalyzeRequest{
		TaskID:     taskMem.TaskID,
		Goal:       step.Description,
		Screenshot: screenshot,
	})
	if err != nil {
		return fmt.Errorf("failed to analyze screenshot: %w", err)
	}

	// Store screenshot
	taskMem.AddScreenshot(screenshot, []interface{}{elements}, map[string]interface{}{
		"step": step.ID,
	})

	// Determine action using Gemma
	elementStrs := make([]string, 0)
	if elemMap, ok := elements.(map[string]interface{}); ok {
		if elems, ok := elemMap["elements"].([]interface{}); ok {
			for _, elem := range elems {
				if elemStr, ok := elem.(string); ok {
					elementStrs = append(elementStrs, elemStr)
				}
			}
		}
	}

	actionPlan, err := e.controller.gemma.AnalyzeScreenshot(ctx, elementStrs, step.Description)
	if err != nil {
		return fmt.Errorf("failed to analyze screenshot: %w", err)
	}

	// Parse and execute action
	// TODO: Parse actionPlan and execute specific browser actions
	// For now, just navigate if URL is in action
	if contains(action, "http") {
		url := extractURL(action)
		if err := e.controller.browserMgr.Navigate(url); err != nil {
			return fmt.Errorf("failed to navigate: %w", err)
		}
	}

	return nil
}

// executeTerminalStep executes a terminal step
func (e *Executor) executeTerminalStep(ctx context.Context, step Step, taskMem *memory.TaskMemory) error {
	// Execute command
	output, err := e.controller.terminalMgr.ExecuteWithContext(ctx, step.Action)
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	// Store output
	taskMem.AddAction("terminal", step.Action, step.Parameters, output, true, "")

	// Store in long-term memory
	e.controller.longTermMem.StoreAction(ctx, step.Action, output, true)

	return nil
}

// executeMCPStep executes an MCP step
func (e *Executor) executeMCPStep(ctx context.Context, step Step, taskMem *memory.TaskMemory) error {
	// Parse MCP tool call
	// TODO: Parse step.Action to extract server, tool, and args
	server := "dynamic-thinking"
	tool := "perceive"
	args := step.Parameters

	// Call MCP tool
	result, err := e.controller.mcpClient.CallTool(server, tool, args)
	if err != nil {
		return fmt.Errorf("MCP tool call failed: %w", err)
	}

	// Store result
	taskMem.AddAction("mcp", step.Action, args, result, true, "")

	return nil
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func extractURL(s string) string {
	// Simple URL extraction
	words := strings.Fields(s)
	for _, word := range words {
		if strings.HasPrefix(word, "http://") || strings.HasPrefix(word, "https://") {
			return word
		}
	}
	return ""
}

