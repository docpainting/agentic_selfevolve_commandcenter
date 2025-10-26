package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"agent-workspace/backend/internal/browser"
	"agent-workspace/backend/internal/mcp"
	"agent-workspace/backend/internal/memory"
	"agent-workspace/backend/internal/terminal"
	"agent-workspace/backend/internal/watchdog"
	"agent-workspace/backend/pkg/models"
)

// Controller orchestrates the agent's operations
type Controller struct {
	longTermMem  *memory.LongTermMemory
	shortTermMem *memory.ShortTermMemory
	browserMgr   *browser.Manager
	terminalMgr  *terminal.Manager
	mcpClient    *mcp.Client
	watchdog     *watchdog.Watchdog
	gemma        *GemmaClient
	planner      *Planner
	executor     *Executor
	state        string
	currentTask  string
	mu           sync.RWMutex
}

// NewController creates a new agent controller
func NewController(
	longTermMem *memory.LongTermMemory,
	shortTermMem *memory.ShortTermMemory,
	browserMgr *browser.Manager,
	terminalMgr *terminal.Manager,
	mcpClient *mcp.Client,
	wdog *watchdog.Watchdog,
) *Controller {
	gemma := NewGemmaClient()
	
	c := &Controller{
		longTermMem:  longTermMem,
		shortTermMem: shortTermMem,
		browserMgr:   browserMgr,
		terminalMgr:  terminalMgr,
		mcpClient:    mcpClient,
		watchdog:     wdog,
		gemma:        gemma,
		state:        "idle",
	}

	c.planner = NewPlanner(c)
	c.executor = NewExecutor(c)

	return c
}

// Initialize initializes the agent
func (c *Controller) Initialize(req models.InitializeRequest) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	c.state = "initialized"

	// Initialize browser if needed
	if err := c.browserMgr.Initialize(); err != nil {
		return "", fmt.Errorf("failed to initialize browser: %w", err)
	}

	// Start watchdog
	if err := c.watchdog.Start(); err != nil {
		return "", fmt.Errorf("failed to start watchdog: %w", err)
	}

	return sessionID, nil
}

// ExecuteCommand executes a user command
func (c *Controller) ExecuteCommand(req models.CommandRequest) (string, error) {
	c.mu.Lock()
	taskID := fmt.Sprintf("task_%d", time.Now().Unix())
	c.currentTask = taskID
	c.state = "working"
	c.mu.Unlock()

	// Create task memory
	taskMem := c.shortTermMem.CreateTask(taskID)

	// Store command in long-term memory
	ctx := context.Background()
	if err := c.longTermMem.StoreConversation(ctx, req.Command, ""); err != nil {
		fmt.Printf("Warning: failed to store conversation: %v\n", err)
	}

	// Plan execution
	plan, err := c.planner.CreatePlan(ctx, req.Command, taskMem)
	if err != nil {
		c.mu.Lock()
		c.state = "idle"
		c.mu.Unlock()
		return "", fmt.Errorf("failed to create plan: %w", err)
	}

	// Execute plan asynchronously
	go func() {
		if err := c.executor.ExecutePlan(ctx, plan, taskMem); err != nil {
			fmt.Printf("Execution error: %v\n", err)
		}

		c.mu.Lock()
		c.state = "idle"
		c.currentTask = ""
		c.mu.Unlock()
	}()

	return taskID, nil
}

// GetStatus returns the agent's current status
func (c *Controller) GetStatus() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"state":        c.state,
		"current_task": c.currentTask,
		"timestamp":    time.Now().Format(time.RFC3339),
	}
}

// Pause pauses the agent
func (c *Controller) Pause() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != "working" {
		return fmt.Errorf("agent not working")
	}

	c.state = "paused"
	return nil
}

// Resume resumes the agent
func (c *Controller) Resume() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != "paused" {
		return fmt.Errorf("agent not paused")
	}

	c.state = "working"
	return nil
}

// GetFileTree returns the file tree
func (c *Controller) GetFileTree(path string) (interface{}, error) {
	// TODO: Implement file tree retrieval
	return map[string]interface{}{
		"path":  path,
		"files": []string{},
	}, nil
}

// GetFileContent returns file content
func (c *Controller) GetFileContent(path string) (interface{}, error) {
	// TODO: Implement file content retrieval
	return map[string]interface{}{
		"path":    path,
		"content": "",
	}, nil
}

// WriteFile writes content to a file
func (c *Controller) WriteFile(req models.FileWriteRequest) (interface{}, error) {
	// TODO: Implement file writing
	return map[string]interface{}{
		"success": true,
		"path":    req.Path,
	}, nil
}

// ApplyDiff applies a diff to a file
func (c *Controller) ApplyDiff(req models.FileDiffRequest) (interface{}, error) {
	// TODO: Implement diff application
	return map[string]interface{}{
		"success": true,
	}, nil
}

// SearchFiles searches for files
func (c *Controller) SearchFiles(query string) (interface{}, error) {
	// TODO: Implement file search
	return map[string]interface{}{
		"query":   query,
		"results": []string{},
	}, nil
}

// QueryMemory queries the knowledge graph
func (c *Controller) QueryMemory(req models.MemoryQueryRequest) (interface{}, error) {
	ctx := context.Background()
	result, err := c.longTermMem.Query(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"query":  req.Query,
		"result": result,
	}, nil
}

// StoreMemory stores content in memory
func (c *Controller) StoreMemory(req models.MemoryStoreRequest) (interface{}, error) {
	ctx := context.Background()
	if err := c.longTermMem.Store(ctx, req.Content, req.Metadata); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// GetMemoryContext retrieves memory context
func (c *Controller) GetMemoryContext(query string, maxTokens int) (interface{}, error) {
	ctx := context.Background()
	context, err := c.longTermMem.GetContext(ctx, query, maxTokens)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"query":   query,
		"context": context,
	}, nil
}

// VectorSearch performs vector search
func (c *Controller) VectorSearch(req models.VectorSearchRequest) (interface{}, error) {
	ctx := context.Background()
	results, err := c.longTermMem.VectorSearch(ctx, req.Query, req.TopK)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"query":   req.Query,
		"results": results,
	}, nil
}

// GetOpenEvolveStatus returns OpenEvolve status
func (c *Controller) GetOpenEvolveStatus() (interface{}, error) {
	return c.watchdog.GetStatus(), nil
}

// SubmitProposal submits an evolution proposal
func (c *Controller) SubmitProposal(req models.ProposalRequest) (interface{}, error) {
	id, err := c.watchdog.SubmitProposal(req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"proposal_id": id,
		"status":      "pending",
	}, nil
}

// SetRewards sets rewards for proposals
func (c *Controller) SetRewards(req models.RewardRequest) (interface{}, error) {
	if err := c.watchdog.SetReward(req); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// GetMetrics returns OpenEvolve metrics
func (c *Controller) GetMetrics() (interface{}, error) {
	return c.watchdog.GetMetrics(), nil
}

// AnalyzeVision analyzes a screenshot
func (c *Controller) AnalyzeVision(req models.VisionAnalyzeRequest) (interface{}, error) {
	return c.browserMgr.AnalyzeScreenshot(req)
}

// GetVisionState returns vision state
func (c *Controller) GetVisionState() (interface{}, error) {
	return map[string]interface{}{
		"current_url": c.browserMgr.GetCurrentURL(),
		"elements":    c.browserMgr.GetElements(),
	}, nil
}

// SetVisionContext sets vision context
func (c *Controller) SetVisionContext(req models.VisionContextRequest) (interface{}, error) {
	// TODO: Implement vision context setting
	return map[string]interface{}{
		"success": true,
	}, nil
}

// Cleanup cleans up resources
func (c *Controller) Cleanup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.browserMgr != nil {
		c.browserMgr.Cleanup()
	}

	if c.terminalMgr != nil {
		c.terminalMgr.Cleanup()
	}

	if c.mcpClient != nil {
		c.mcpClient.Cleanup()
	}

	if c.watchdog != nil {
		c.watchdog.Stop()
	}

	if c.longTermMem != nil {
		c.longTermMem.Cleanup()
	}

	return nil
}

