package terminal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
)

// Manager manages terminal sessions
type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// Session represents a terminal session
type Session struct {
	ID        string
	PTY       *os.File
	CMD       *exec.Cmd
	Output    *OutputBuffer
	CreatedAt time.Time
	mu        sync.Mutex
}

// OutputBuffer stores terminal output
type OutputBuffer struct {
	lines []string
	mu    sync.RWMutex
}

// NewManager creates a new terminal manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new terminal session
func (m *Manager) CreateSession(id string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[id]; exists {
		return nil, fmt.Errorf("session %s already exists", id)
	}

	// Create PTY
	cmd := exec.Command("/bin/bash")
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"PS1=$ ",
	)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}

	session := &Session{
		ID:        id,
		PTY:       ptmx,
		CMD:       cmd,
		Output:    &OutputBuffer{lines: make([]string, 0)},
		CreatedAt: time.Now(),
	}

	// Start output reader
	go session.readOutput()

	m.sessions[id] = session
	return session, nil
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session %s not found", id)
	}

	return session, nil
}

// GetOrCreateSession gets an existing session or creates a new one
func (m *Manager) GetOrCreateSession(id string) (*Session, error) {
	session, err := m.GetSession(id)
	if err == nil {
		return session, nil
	}

	return m.CreateSession(id)
}

// CloseSession closes a terminal session
func (m *Manager) CloseSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session %s not found", id)
	}

	session.Close()
	delete(m.sessions, id)

	return nil
}

// ListSessions returns all active sessions
func (m *Manager) ListSessions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}

	return ids
}

// Execute executes a command in the default session
func (m *Manager) Execute(command string) (string, error) {
	return m.ExecuteInSession("default", command)
}

// ExecuteInSession executes a command in a specific session
func (m *Manager) ExecuteInSession(sessionID, command string) (string, error) {
	session, err := m.GetOrCreateSession(sessionID)
	if err != nil {
		return "", err
	}

	return session.Execute(command)
}

// ExecuteWithContext executes a command with context
func (m *Manager) ExecuteWithContext(ctx context.Context, command string) (string, error) {
	return m.ExecuteInSessionWithContext(ctx, "default", command)
}

// ExecuteInSessionWithContext executes a command in a session with context
func (m *Manager) ExecuteInSessionWithContext(ctx context.Context, sessionID, command string) (string, error) {
	session, err := m.GetOrCreateSession(sessionID)
	if err != nil {
		return "", err
	}

	return session.ExecuteWithContext(ctx, command)
}

// GetOutput returns the output buffer for a session
func (m *Manager) GetOutput(sessionID string) ([]string, error) {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return session.Output.GetLines(), nil
}

// Cleanup closes all sessions
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, session := range m.sessions {
		session.Close()
		delete(m.sessions, id)
	}
}

// Session methods

// Execute executes a command in the session
func (s *Session) Execute(command string) (string, error) {
	return s.ExecuteWithContext(context.Background(), command)
}

// ExecuteWithContext executes a command with context
func (s *Session) ExecuteWithContext(ctx context.Context, command string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear previous output
	marker := fmt.Sprintf("__CMD_START_%d__", time.Now().UnixNano())
	endMarker := fmt.Sprintf("__CMD_END_%d__", time.Now().UnixNano())

	// Write command with markers
	fullCommand := fmt.Sprintf("echo '%s' && %s && echo '%s'\n", marker, command, endMarker)
	if _, err := s.PTY.Write([]byte(fullCommand)); err != nil {
		return "", fmt.Errorf("failed to write command: %w", err)
	}

	// Wait for output with timeout
	timeout := 30 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	outputChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		output, err := s.waitForOutput(marker, endMarker, timeout)
		if err != nil {
			errorChan <- err
			return
		}
		outputChan <- output
	}()

	select {
	case output := <-outputChan:
		return output, nil
	case err := <-errorChan:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// SendInput sends input to the PTY
func (s *Session) SendInput(input string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.PTY.Write([]byte(input)); err != nil {
		return fmt.Errorf("failed to write input: %w", err)
	}

	return nil
}

// Close closes the session
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.PTY != nil {
		s.PTY.Close()
	}

	if s.CMD != nil && s.CMD.Process != nil {
		s.CMD.Process.Kill()
	}
}

// readOutput continuously reads from PTY
func (s *Session) readOutput() {
	reader := bufio.NewReader(s.PTY)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading PTY: %v\n", err)
			}
			break
		}

		s.Output.AddLine(line)
	}
}

// waitForOutput waits for command output between markers
func (s *Session) waitForOutput(startMarker, endMarker string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	output := ""
	capturing := false

	for time.Now().Before(deadline) {
		lines := s.Output.GetRecentLines(100)

		for _, line := range lines {
			if !capturing && contains(line, startMarker) {
				capturing = true
				continue
			}

			if capturing {
				if contains(line, endMarker) {
					return output, nil
				}
				output += line
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return "", fmt.Errorf("command timeout after %v", timeout)
}

// OutputBuffer methods

// AddLine adds a line to the buffer
func (b *OutputBuffer) AddLine(line string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lines = append(b.lines, line)

	// Keep only last 1000 lines
	if len(b.lines) > 1000 {
		b.lines = b.lines[len(b.lines)-1000:]
	}
}

// GetLines returns all lines
func (b *OutputBuffer) GetLines() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	lines := make([]string, len(b.lines))
	copy(lines, b.lines)
	return lines
}

// GetRecentLines returns the last N lines
func (b *OutputBuffer) GetRecentLines(n int) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if n > len(b.lines) {
		n = len(b.lines)
	}

	lines := make([]string, n)
	copy(lines, b.lines[len(b.lines)-n:])
	return lines
}

// Clear clears the buffer
func (b *OutputBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lines = make([]string, 0)
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

