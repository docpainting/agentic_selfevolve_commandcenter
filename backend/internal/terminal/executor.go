package terminal

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Executor executes commands with history tracking
type Executor struct {
	manager *Manager
	history *CommandHistory
}

// CommandHistory tracks command execution history
type CommandHistory struct {
	entries []CommandEntry
	mu      sync.RWMutex
}

// CommandEntry represents a command execution
type CommandEntry struct {
	ID          string
	Command     string
	Output      string
	ExitCode    int
	Source      string // "ai" or "user"
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	SessionID   string
	Error       string
}

// NewExecutor creates a new command executor
func NewExecutor(manager *Manager) *Executor {
	return &Executor{
		manager: manager,
		history: &CommandHistory{
			entries: make([]CommandEntry, 0),
		},
	}
}

// Execute executes a command and records it in history
func (e *Executor) Execute(command, source string) (*CommandEntry, error) {
	return e.ExecuteInSession("default", command, source)
}

// ExecuteInSession executes a command in a specific session
func (e *Executor) ExecuteInSession(sessionID, command, source string) (*CommandEntry, error) {
	return e.ExecuteInSessionWithContext(context.Background(), sessionID, command, source)
}

// ExecuteInSessionWithContext executes a command with context
func (e *Executor) ExecuteInSessionWithContext(ctx context.Context, sessionID, command, source string) (*CommandEntry, error) {
	entry := CommandEntry{
		ID:        generateID(),
		Command:   command,
		Source:    source,
		StartTime: time.Now(),
		SessionID: sessionID,
	}

	// Execute command
	output, err := e.manager.ExecuteInSessionWithContext(ctx, sessionID, command)
	entry.EndTime = time.Now()
	entry.Duration = entry.EndTime.Sub(entry.StartTime)
	entry.Output = output

	if err != nil {
		entry.Error = err.Error()
		entry.ExitCode = 1
	} else {
		entry.ExitCode = 0
	}

	// Add to history
	e.history.Add(entry)

	return &entry, err
}

// ExecuteBatch executes multiple commands sequentially
func (e *Executor) ExecuteBatch(commands []string, source string) ([]*CommandEntry, error) {
	entries := make([]*CommandEntry, 0, len(commands))

	for _, cmd := range commands {
		entry, err := e.Execute(cmd, source)
		entries = append(entries, entry)

		if err != nil {
			// Stop on first error
			return entries, fmt.Errorf("batch execution failed at command %d: %w", len(entries), err)
		}
	}

	return entries, nil
}

// ExecuteScript executes a multi-line script
func (e *Executor) ExecuteScript(script, source string) (*CommandEntry, error) {
	// Create a temporary script file
	scriptFile := fmt.Sprintf("/tmp/script_%s.sh", generateID())
	
	// Write script to session
	session, err := e.manager.GetOrCreateSession("default")
	if err != nil {
		return nil, err
	}

	commands := []string{
		fmt.Sprintf("cat > %s << 'SCRIPT_EOF'", scriptFile),
		script,
		"SCRIPT_EOF",
		fmt.Sprintf("chmod +x %s", scriptFile),
		scriptFile,
		fmt.Sprintf("rm -f %s", scriptFile),
	}

	fullCommand := strings.Join(commands, "\n")
	return e.Execute(fullCommand, source)
}

// GetHistory returns command history
func (e *Executor) GetHistory() []CommandEntry {
	return e.history.GetAll()
}

// GetHistoryBySource returns commands filtered by source
func (e *Executor) GetHistoryBySource(source string) []CommandEntry {
	return e.history.GetBySource(source)
}

// GetHistoryBySession returns commands filtered by session
func (e *Executor) GetHistoryBySession(sessionID string) []CommandEntry {
	return e.history.GetBySession(sessionID)
}

// ClearHistory clears command history
func (e *Executor) ClearHistory() {
	e.history.Clear()
}

// CommandHistory methods

// Add adds a command entry to history
func (h *CommandHistory) Add(entry CommandEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = append(h.entries, entry)

	// Keep only last 1000 entries
	if len(h.entries) > 1000 {
		h.entries = h.entries[len(h.entries)-1000:]
	}
}

// GetAll returns all entries
func (h *CommandHistory) GetAll() []CommandEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	entries := make([]CommandEntry, len(h.entries))
	copy(entries, h.entries)
	return entries
}

// GetRecent returns the last N entries
func (h *CommandHistory) GetRecent(n int) []CommandEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n > len(h.entries) {
		n = len(h.entries)
	}

	entries := make([]CommandEntry, n)
	copy(entries, h.entries[len(h.entries)-n:])
	return entries
}

// GetBySource returns entries filtered by source
func (h *CommandHistory) GetBySource(source string) []CommandEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	filtered := make([]CommandEntry, 0)
	for _, entry := range h.entries {
		if entry.Source == source {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// GetBySession returns entries filtered by session
func (h *CommandHistory) GetBySession(sessionID string) []CommandEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	filtered := make([]CommandEntry, 0)
	for _, entry := range h.entries {
		if entry.SessionID == sessionID {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// GetByTimeRange returns entries within a time range
func (h *CommandHistory) GetByTimeRange(start, end time.Time) []CommandEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	filtered := make([]CommandEntry, 0)
	for _, entry := range h.entries {
		if entry.StartTime.After(start) && entry.StartTime.Before(end) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// Clear clears all history
func (h *CommandHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = make([]CommandEntry, 0)
}

// GetStats returns execution statistics
func (h *CommandHistory) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalCommands := len(h.entries)
	aiCommands := 0
	userCommands := 0
	successfulCommands := 0
	failedCommands := 0
	var totalDuration time.Duration

	for _, entry := range h.entries {
		if entry.Source == "ai" {
			aiCommands++
		} else {
			userCommands++
		}

		if entry.ExitCode == 0 {
			successfulCommands++
		} else {
			failedCommands++
		}

		totalDuration += entry.Duration
	}

	avgDuration := time.Duration(0)
	if totalCommands > 0 {
		avgDuration = totalDuration / time.Duration(totalCommands)
	}

	return map[string]interface{}{
		"total_commands":      totalCommands,
		"ai_commands":         aiCommands,
		"user_commands":       userCommands,
		"successful_commands": successfulCommands,
		"failed_commands":     failedCommands,
		"total_duration_ms":   totalDuration.Milliseconds(),
		"avg_duration_ms":     avgDuration.Milliseconds(),
	}
}

// Helper functions

var idCounter int64
var idMutex sync.Mutex

func generateID() string {
	idMutex.Lock()
	defer idMutex.Unlock()

	idCounter++
	return fmt.Sprintf("cmd_%d_%d", time.Now().Unix(), idCounter)
}

