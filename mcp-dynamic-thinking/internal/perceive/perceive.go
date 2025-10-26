package perceive

import (
	"context"
	"fmt"
	"time"
)

// Perceiver handles the Perceive phase of PRAR
type Perceiver struct {
	perceptions map[string]*Perception
}

// Perception represents a perception event
type Perception struct {
	ID          string
	TaskID      string
	Goal        string
	Screenshot  []byte
	Elements    []interface{}
	State       map[string]interface{}
	Analysis    map[string]interface{}
	Timestamp   time.Time
}

// NewPerceiver creates a new perceiver
func NewPerceiver() *Perceiver {
	return &Perceiver{
		perceptions: make(map[string]*Perception),
	}
}

// Perceive captures and analyzes environment state
func (p *Perceiver) Perceive(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	taskID, ok := args["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("task_id is required")
	}

	goal, ok := args["goal"].(string)
	if !ok {
		return nil, fmt.Errorf("goal is required")
	}

	// Create perception
	perceptionID := fmt.Sprintf("perception_%d", time.Now().UnixNano())
	
	perception := &Perception{
		ID:        perceptionID,
		TaskID:    taskID,
		Goal:      goal,
		State:     make(map[string]interface{}),
		Analysis:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Extract screenshot if provided
	if screenshot, ok := args["screenshot"].([]byte); ok {
		perception.Screenshot = screenshot
	}

	// Extract elements if provided
	if elements, ok := args["elements"].([]interface{}); ok {
		perception.Elements = elements
	}

	// Analyze environment
	analysis := p.analyzeEnvironment(perception)
	perception.Analysis = analysis

	// Store perception
	p.perceptions[perceptionID] = perception

	return map[string]interface{}{
		"perception_id": perceptionID,
		"task_id":       taskID,
		"goal":          goal,
		"analysis":      analysis,
		"timestamp":     perception.Timestamp.Format(time.RFC3339),
		"elements_count": len(perception.Elements),
	}, nil
}

// analyzeEnvironment analyzes the perceived environment
func (p *Perceiver) analyzeEnvironment(perception *Perception) map[string]interface{} {
	analysis := make(map[string]interface{})

	// Analyze elements
	if len(perception.Elements) > 0 {
		analysis["interactive_elements"] = len(perception.Elements)
		analysis["has_inputs"] = p.hasInputElements(perception.Elements)
		analysis["has_buttons"] = p.hasButtonElements(perception.Elements)
		analysis["has_links"] = p.hasLinkElements(perception.Elements)
	}

	// Analyze state
	analysis["state_complexity"] = len(perception.State)

	// Determine readiness
	analysis["ready_for_action"] = len(perception.Elements) > 0

	return analysis
}

// hasInputElements checks if there are input elements
func (p *Perceiver) hasInputElements(elements []interface{}) bool {
	for _, elem := range elements {
		if elemMap, ok := elem.(map[string]interface{}); ok {
			if tag, ok := elemMap["tag"].(string); ok {
				if tag == "input" || tag == "textarea" {
					return true
				}
			}
		}
	}
	return false
}

// hasButtonElements checks if there are button elements
func (p *Perceiver) hasButtonElements(elements []interface{}) bool {
	for _, elem := range elements {
		if elemMap, ok := elem.(map[string]interface{}); ok {
			if tag, ok := elemMap["tag"].(string); ok {
				if tag == "button" {
					return true
				}
			}
		}
	}
	return false
}

// hasLinkElements checks if there are link elements
func (p *Perceiver) hasLinkElements(elements []interface{}) bool {
	for _, elem := range elements {
		if elemMap, ok := elem.(map[string]interface{}); ok {
			if tag, ok := elemMap["tag"].(string); ok {
				if tag == "a" {
					return true
				}
			}
		}
	}
	return false
}

// GetPerception retrieves a perception by ID
func (p *Perceiver) GetPerception(perceptionID string) (*Perception, error) {
	perception, exists := p.perceptions[perceptionID]
	if !exists {
		return nil, fmt.Errorf("perception %s not found", perceptionID)
	}

	return perception, nil
}

