package memory

import (
	"fmt"
	"sync"
	"time"
)

// ShortTermMemory manages task-specific short-term memory
type ShortTermMemory struct {
	tasks map[string]*TaskMemory
	mu    sync.RWMutex
}

// TaskMemory stores memory for a specific task
type TaskMemory struct {
	TaskID       string
	Perceptions  []Perception
	Reasoning    []ReasoningBranch
	Actions      []Action
	Reflections  []Reflection
	Screenshots  []Screenshot
	Context      map[string]interface{}
	CreatedAt    time.Time
	LastAccessed time.Time
	mu           sync.RWMutex
}

// Perception represents a perception event
type Perception struct {
	ID          string
	Timestamp   time.Time
	Type        string // "screenshot", "text", "state"
	Content     interface{}
	Analysis    map[string]interface{}
}

// ReasoningBranch represents a reasoning path
type ReasoningBranch struct {
	ID          string
	Timestamp   time.Time
	Prompt      string
	Response    string
	Confidence  float64
	Selected    bool
	Reasoning   string
}

// Action represents an executed action
type Action struct {
	ID          string
	Timestamp   time.Time
	Type        string // "browser", "terminal", "mcp"
	Command     string
	Parameters  map[string]interface{}
	Result      interface{}
	Success     bool
	Error       string
}

// Reflection represents a reflection on results
type Reflection struct {
	ID          string
	Timestamp   time.Time
	ActionID    string
	Critique    string
	Lessons     []string
	NextSteps   []string
}

// Screenshot represents a captured screenshot
type Screenshot struct {
	ID          string
	Timestamp   time.Time
	Data        []byte
	Elements    []interface{}
	Analysis    map[string]interface{}
}

// NewShortTermMemory creates a new short-term memory system
func NewShortTermMemory() *ShortTermMemory {
	return &ShortTermMemory{
		tasks: make(map[string]*TaskMemory),
	}
}

// CreateTask creates a new task memory
func (m *ShortTermMemory) CreateTask(taskID string) *TaskMemory {
	m.mu.Lock()
	defer m.mu.Unlock()

	task := &TaskMemory{
		TaskID:       taskID,
		Perceptions:  make([]Perception, 0),
		Reasoning:    make([]ReasoningBranch, 0),
		Actions:      make([]Action, 0),
		Reflections:  make([]Reflection, 0),
		Screenshots:  make([]Screenshot, 0),
		Context:      make(map[string]interface{}),
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}

	m.tasks[taskID] = task
	return task
}

// GetTask retrieves a task memory
func (m *ShortTermMemory) GetTask(taskID string) (*TaskMemory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	task.LastAccessed = time.Now()
	return task, nil
}

// GetOrCreateTask gets or creates a task memory
func (m *ShortTermMemory) GetOrCreateTask(taskID string) *TaskMemory {
	task, err := m.GetTask(taskID)
	if err == nil {
		return task
	}

	return m.CreateTask(taskID)
}

// DeleteTask deletes a task memory
func (m *ShortTermMemory) DeleteTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[taskID]; !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	delete(m.tasks, taskID)
	return nil
}

// ClearTask clears a task's memory but keeps the task
func (m *ShortTermMemory) ClearTask(taskID string) error {
	task, err := m.GetTask(taskID)
	if err != nil {
		return err
	}

	task.mu.Lock()
	defer task.mu.Unlock()

	task.Perceptions = make([]Perception, 0)
	task.Reasoning = make([]ReasoningBranch, 0)
	task.Actions = make([]Action, 0)
	task.Reflections = make([]Reflection, 0)
	task.Screenshots = make([]Screenshot, 0)

	return nil
}

// ListTasks returns all task IDs
func (m *ShortTermMemory) ListTasks() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.tasks))
	for id := range m.tasks {
		ids = append(ids, id)
	}

	return ids
}

// CleanupOldTasks removes tasks older than duration
func (m *ShortTermMemory) CleanupOldTasks(maxAge time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	cutoff := time.Now().Add(-maxAge)

	for id, task := range m.tasks {
		if task.LastAccessed.Before(cutoff) {
			delete(m.tasks, id)
			count++
		}
	}

	return count
}

// TaskMemory methods

// AddPerception adds a perception to task memory
func (t *TaskMemory) AddPerception(perceptionType string, content interface{}, analysis map[string]interface{}) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := fmt.Sprintf("perception_%d", time.Now().UnixNano())
	
	perception := Perception{
		ID:        id,
		Timestamp: time.Now(),
		Type:      perceptionType,
		Content:   content,
		Analysis:  analysis,
	}

	t.Perceptions = append(t.Perceptions, perception)
	return id
}

// AddReasoning adds a reasoning branch
func (t *TaskMemory) AddReasoning(prompt, response string, confidence float64, selected bool) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := fmt.Sprintf("reasoning_%d", time.Now().UnixNano())
	
	reasoning := ReasoningBranch{
		ID:         id,
		Timestamp:  time.Now(),
		Prompt:     prompt,
		Response:   response,
		Confidence: confidence,
		Selected:   selected,
		Reasoning:  response,
	}

	t.Reasoning = append(t.Reasoning, reasoning)
	return id
}

// AddAction adds an action to task memory
func (t *TaskMemory) AddAction(actionType, command string, params map[string]interface{}, result interface{}, success bool, err string) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := fmt.Sprintf("action_%d", time.Now().UnixNano())
	
	action := Action{
		ID:         id,
		Timestamp:  time.Now(),
		Type:       actionType,
		Command:    command,
		Parameters: params,
		Result:     result,
		Success:    success,
		Error:      err,
	}

	t.Actions = append(t.Actions, action)
	return id
}

// AddReflection adds a reflection
func (t *TaskMemory) AddReflection(actionID, critique string, lessons, nextSteps []string) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := fmt.Sprintf("reflection_%d", time.Now().UnixNano())
	
	reflection := Reflection{
		ID:        id,
		Timestamp: time.Now(),
		ActionID:  actionID,
		Critique:  critique,
		Lessons:   lessons,
		NextSteps: nextSteps,
	}

	t.Reflections = append(t.Reflections, reflection)
	return id
}

// AddScreenshot adds a screenshot
func (t *TaskMemory) AddScreenshot(data []byte, elements []interface{}, analysis map[string]interface{}) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := fmt.Sprintf("screenshot_%d", time.Now().UnixNano())
	
	screenshot := Screenshot{
		ID:        id,
		Timestamp: time.Now(),
		Data:      data,
		Elements:  elements,
		Analysis:  analysis,
	}

	t.Screenshots = append(t.Screenshots, screenshot)
	return id
}

// GetPerceptions returns all perceptions
func (t *TaskMemory) GetPerceptions() []Perception {
	t.mu.RLock()
	defer t.mu.RUnlock()

	perceptions := make([]Perception, len(t.Perceptions))
	copy(perceptions, t.Perceptions)
	return perceptions
}

// GetReasoning returns all reasoning branches
func (t *TaskMemory) GetReasoning() []ReasoningBranch {
	t.mu.RLock()
	defer t.mu.RUnlock()

	reasoning := make([]ReasoningBranch, len(t.Reasoning))
	copy(reasoning, t.Reasoning)
	return reasoning
}

// GetActions returns all actions
func (t *TaskMemory) GetActions() []Action {
	t.mu.RLock()
	defer t.mu.RUnlock()

	actions := make([]Action, len(t.Actions))
	copy(actions, t.Actions)
	return actions
}

// GetReflections returns all reflections
func (t *TaskMemory) GetReflections() []Reflection {
	t.mu.RLock()
	defer t.mu.RUnlock()

	reflections := make([]Reflection, len(t.Reflections))
	copy(reflections, t.Reflections)
	return reflections
}

// GetScreenshots returns all screenshots
func (t *TaskMemory) GetScreenshots() []Screenshot {
	t.mu.RLock()
	defer t.mu.RUnlock()

	screenshots := make([]Screenshot, len(t.Screenshots))
	copy(screenshots, t.Screenshots)
	return screenshots
}

// GetLatestScreenshot returns the most recent screenshot
func (t *TaskMemory) GetLatestScreenshot() (*Screenshot, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.Screenshots) == 0 {
		return nil, fmt.Errorf("no screenshots available")
	}

	screenshot := t.Screenshots[len(t.Screenshots)-1]
	return &screenshot, nil
}

// SetContext sets a context value
func (t *TaskMemory) SetContext(key string, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Context[key] = value
}

// GetContext gets a context value
func (t *TaskMemory) GetContext(key string) (interface{}, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	value, exists := t.Context[key]
	return value, exists
}

// GetSummary returns a summary of the task memory
func (t *TaskMemory) GetSummary() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return map[string]interface{}{
		"task_id":          t.TaskID,
		"perceptions":      len(t.Perceptions),
		"reasoning":        len(t.Reasoning),
		"actions":          len(t.Actions),
		"reflections":      len(t.Reflections),
		"screenshots":      len(t.Screenshots),
		"created_at":       t.CreatedAt.Format(time.RFC3339),
		"last_accessed":    t.LastAccessed.Format(time.RFC3339),
		"age_seconds":      time.Since(t.CreatedAt).Seconds(),
	}
}

// ExportTrace exports the complete execution trace
func (t *TaskMemory) ExportTrace() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return map[string]interface{}{
		"task_id":     t.TaskID,
		"perceptions": t.Perceptions,
		"reasoning":   t.Reasoning,
		"actions":     t.Actions,
		"reflections": t.Reflections,
		"context":     t.Context,
		"created_at":  t.CreatedAt.Format(time.RFC3339),
		"duration":    time.Since(t.CreatedAt).Seconds(),
	}
}

