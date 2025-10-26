package agent

import (
	"context"
	"fmt"

	"agent-workspace/backend/pkg/models"
	"agent-workspace/backend/pkg/ollama"
)

// GemmaClient handles communication with Gemma 3 via Ollama
type GemmaClient struct {
	client *ollama.Client
}

// NewGemmaClient creates a new Gemma client
func NewGemmaClient() *GemmaClient {
	return &GemmaClient{
		client: ollama.NewClient(),
	}
}

// GenerateResponse generates a response from Gemma
func (g *GemmaClient) GenerateResponse(ctx context.Context, messages []models.Message, temperature float64) (string, error) {
	return g.client.ChatCompletion(messages, temperature)
}

// GenerateResponseStream generates a streaming response
func (g *GemmaClient) GenerateResponseStream(ctx context.Context, messages []models.Message, temperature float64, callback func(string) error) error {
	return g.client.ChatCompletionStream(messages, temperature, callback)
}

// GeneratePlan generates an execution plan
func (g *GemmaClient) GeneratePlan(ctx context.Context, command string, context string) (string, error) {
	prompt := fmt.Sprintf(`You are an AI agent planning how to execute a user command.

User Command: %s

Context:
%s

Generate a detailed execution plan with these sections:
1. GOAL: What we're trying to achieve
2. STEPS: Numbered list of actions to take
3. TOOLS: Which tools to use (browser, terminal, mcp)
4. EXPECTED_OUTCOME: What success looks like

Be specific and actionable.`, command, context)

	messages := []models.Message{
		{
			Role: "user",
			Parts: []models.MessagePart{
				{Type: "text", Text: prompt},
			},
		},
	}

	return g.GenerateResponse(ctx, messages, 0.7)
}

// GenerateReasoning generates reasoning for a decision
func (g *GemmaClient) GenerateReasoning(ctx context.Context, situation string, options []string) (string, error) {
	optionsText := ""
	for i, opt := range options {
		optionsText += fmt.Sprintf("%d. %s\n", i+1, opt)
	}

	prompt := fmt.Sprintf(`You are an AI agent making a decision.

Situation: %s

Options:
%s

Analyze each option and provide:
1. ANALYSIS: Pros and cons of each option
2. RECOMMENDATION: Which option to choose and why
3. CONFIDENCE: Your confidence level (0-1)
4. REASONING: Detailed explanation of your choice

Be logical and thorough.`, situation, optionsText)

	messages := []models.Message{
		{
			Role: "user",
			Parts: []models.MessagePart{
				{Type: "text", Text: prompt},
			},
		},
	}

	return g.GenerateResponse(ctx, messages, 0.7)
}

// GenerateReflection generates a reflection on results
func (g *GemmaClient) GenerateReflection(ctx context.Context, action string, result string, success bool) (string, error) {
	statusText := "succeeded"
	if !success {
		statusText = "failed"
	}

	prompt := fmt.Sprintf(`You are an AI agent reflecting on an action you took.

Action: %s
Result: %s
Status: %s

Provide a reflection with:
1. CRITIQUE: What went well and what didn't
2. LESSONS: Key learnings from this experience
3. IMPROVEMENTS: How to do better next time
4. NEXT_STEPS: What to do now

Be honest and constructive.`, action, result, statusText)

	messages := []models.Message{
		{
			Role: "user",
			Parts: []models.MessagePart{
				{Type: "text", Text: prompt},
			},
		},
	}

	return g.GenerateResponse(ctx, messages, 0.7)
}

// ParseAction parses an action from text
func (g *GemmaClient) ParseAction(ctx context.Context, text string) (*Action, error) {
	prompt := fmt.Sprintf(`Parse this text into a structured action.

Text: %s

Return JSON with:
{
  "type": "browser|terminal|mcp",
  "command": "the specific command",
  "parameters": {"key": "value"}
}

Only return the JSON, nothing else.`, text)

	messages := []models.Message{
		{
			Role: "user",
			Parts: []models.MessagePart{
				{Type: "text", Text: prompt},
			},
		},
	}

	response, err := g.GenerateResponse(ctx, messages, 0.3)
	if err != nil {
		return nil, err
	}

	// TODO: Parse JSON response into Action struct
	// For now, return a placeholder
	return &Action{
		Type:    "terminal",
		Command: text,
	}, nil
}

// GenerateCode generates code
func (g *GemmaClient) GenerateCode(ctx context.Context, description string, language string) (string, error) {
	prompt := fmt.Sprintf(`Generate %s code for the following:

%s

Only return the code, no explanations.`, language, description)

	messages := []models.Message{
		{
			Role: "user",
			Parts: []models.MessagePart{
				{Type: "text", Text: prompt},
			},
		},
	}

	return g.GenerateResponse(ctx, messages, 0.5)
}

// AnalyzeScreenshot analyzes a screenshot description
func (g *GemmaClient) AnalyzeScreenshot(ctx context.Context, elements []string, goal string) (string, error) {
	elementsText := ""
	for i, elem := range elements {
		elementsText += fmt.Sprintf("%d: %s\n", i, elem)
	}

	prompt := fmt.Sprintf(`You are analyzing a web page to achieve a goal.

Goal: %s

Visible Elements:
%s

Determine:
1. RELEVANT_ELEMENTS: Which elements are relevant to the goal
2. ACTION: What action to take (click, type, navigate)
3. TARGET: Which element to interact with (by number)
4. INPUT: What to type (if applicable)

Be specific and use element numbers.`, goal, elementsText)

	messages := []models.Message{
		{
			Role: "user",
			Parts: []models.MessagePart{
				{Type: "text", Text: prompt},
			},
		},
	}

	return g.GenerateResponse(ctx, messages, 0.7)
}

// Action represents a parsed action
type Action struct {
	Type       string
	Command    string
	Parameters map[string]interface{}
}

