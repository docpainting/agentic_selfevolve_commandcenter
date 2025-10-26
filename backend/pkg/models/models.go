package models

import "time"

// Agent Card (A2A Protocol)
type AgentCard struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Version      string            `json:"version"`
	Capabilities map[string]bool   `json:"capabilities"`
	Skills       []Skill           `json:"skills"`
	URL          string            `json:"url"`
	Transport    string            `json:"transport"`
}

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WebSocket Messages
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Source    string                 `json:"source"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Agent Requests
type InitializeRequest struct {
	SessionID string                 `json:"session_id,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

type CommandRequest struct {
	SessionID string                 `json:"session_id"`
	Command   string                 `json:"command"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// Agent Status
type AgentStatus struct {
	State         string    `json:"state"`
	CurrentTask   string    `json:"current_task,omitempty"`
	TasksQueued   int       `json:"tasks_queued"`
	LastActivity  time.Time `json:"last_activity"`
	Capabilities  []string  `json:"capabilities"`
	MemoryUsage   int64     `json:"memory_usage"`
	BrowserActive bool      `json:"browser_active"`
	TerminalActive bool     `json:"terminal_active"`
}

// File Operations
type FileNode struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Path     string      `json:"path"`
	Size     int64       `json:"size,omitempty"`
	Modified time.Time   `json:"modified,omitempty"`
	Children []*FileNode `json:"children,omitempty"`
}

type FileContent struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Language string `json:"language,omitempty"`
	Size     int64  `json:"size"`
}

type FileWriteRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type FileDiffRequest struct {
	Path string `json:"path"`
	Diff string `json:"diff"`
}

// Memory System
type MemoryQueryRequest struct {
	Query string `json:"query"`
	Mode  string `json:"mode"` // "naive", "local", "global", "hybrid"
}

type MemoryStoreRequest struct {
	Type    string                 `json:"type"` // "message", "file", "concept", etc.
	Content string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type VectorSearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}

// Browser
type VisionAnalyzeRequest struct {
	TaskID     string `json:"task_id"`
	Goal       string `json:"goal"`
	Screenshot []byte `json:"screenshot,omitempty"`
}

type VisionContextRequest struct {
	TaskID  string                 `json:"task_id"`
	Context map[string]interface{} `json:"context"`
}

type BrowserElement struct {
	ID       int     `json:"id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Text     string  `json:"text"`
	Tag      string  `json:"tag"`
	Clickable bool   `json:"clickable"`
}

// OpenEvolve
type ProposalRequest struct {
	Component   string                 `json:"component"`
	Description string                 `json:"description"`
	Changes     map[string]interface{} `json:"changes"`
}

type RewardRequest struct {
	ProposalID string  `json:"proposal_id"`
	Reward     float64 `json:"reward"`
	Feedback   string  `json:"feedback,omitempty"`
}

// Task Management
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Goal        string                 `json:"goal"`
	Steps       []TaskStep             `json:"steps"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

type TaskStep struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Result      map[string]interface{} `json:"result,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// MCP
type MCPToolCall struct {
	Server string                 `json:"server"`
	Tool   string                 `json:"tool"`
	Args   map[string]interface{} `json:"args"`
}

type MCPToolResult struct {
	Success bool                   `json:"success"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// Memory Structures
type Perception struct {
	ID             string         `json:"id"`
	TaskID         string         `json:"task_id"`
	Timestamp      time.Time      `json:"timestamp"`
	Confidence     float64        `json:"confidence"`
	VisualAnalysis VisualAnalysis `json:"visual_analysis"`
	TerminalState  TerminalState  `json:"terminal_state"`
	Decision       string         `json:"decision"`
}

type VisualAnalysis struct {
	ScreenshotPath   string            `json:"screenshot_path"`
	ElementsDetected []BrowserElement  `json:"elements_detected"`
	TextExtracted    string            `json:"text_extracted"`
}

type TerminalState struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
}

type ReasoningSession struct {
	ID             string            `json:"id"`
	TaskID         string            `json:"task_id"`
	Timestamp      time.Time         `json:"timestamp"`
	Branches       []ReasoningBranch `json:"branches"`
	SelectedBranch *ReasoningBranch  `json:"selected_branch"`
	ActionPlan     ActionPlan        `json:"action_plan"`
	Decision       string            `json:"decision"`
}

type ReasoningBranch struct {
	ID               string   `json:"id"`
	Strategy         string   `json:"strategy"`
	ChainOfThought   []string `json:"chain_of_thought"`
	FeasibilityScore float64  `json:"feasibility_score"`
	AlignmentScore   float64  `json:"alignment_score"`
	RiskScore        float64  `json:"risk_score"`
	Selected         bool     `json:"selected"`
}

type ActionPlan struct {
	Subtasks []Subtask `json:"subtasks"`
}

type Subtask struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type Execution struct {
	ID              string          `json:"id"`
	TaskID          string          `json:"task_id"`
	Timestamp       time.Time       `json:"timestamp"`
	Status          string          `json:"status"`
	SubtaskResults  []SubtaskResult `json:"subtask_results"`
	PlanAdjustments []string        `json:"plan_adjustments"`
	Decision        string          `json:"decision"`
}

type SubtaskResult struct {
	SubtaskID       string                 `json:"subtask_id"`
	Status          string                 `json:"status"`
	Result          map[string]interface{} `json:"result"`
	ExecutionTime   float64                `json:"execution_time"`
	AdjustmentsMade []string               `json:"adjustments_made"`
}

type Reflection struct {
	ID            string                 `json:"id"`
	TaskID        string                 `json:"task_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Analysis      Analysis               `json:"analysis"`
	Critique      Critique               `json:"critique"`
	Evolutions    []Evolution            `json:"evolutions"`
	Decision      string                 `json:"decision"`
	TrainingTrace map[string]interface{} `json:"training_trace"`
}

type Analysis struct {
	GoalAchieved bool     `json:"goal_achieved"`
	SuccessRate  float64  `json:"success_rate"`
	WhatWorked   []string `json:"what_worked"`
	WhatFailed   []string `json:"what_failed"`
}

type Critique struct {
	PromptIssues    []string `json:"prompt_issues"`
	ReasoningIssues []string `json:"reasoning_issues"`
	ExecutionIssues []string `json:"execution_issues"`
	Suggestions     []string `json:"suggestions"`
}

type Evolution struct {
	Type    string                 `json:"type"`
	Details map[string]interface{} `json:"details"`
	Applied bool                   `json:"applied"`
}

