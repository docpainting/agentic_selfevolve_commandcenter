package watchdog

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"agent-workspace/backend/internal/memory"
	"agent-workspace/backend/pkg/models"
)

// Watchdog monitors code and detects patterns
type Watchdog struct {
	longTermMem *memory.LongTermMemory
	alerts      []Alert
	proposals   map[string]*Proposal
	patterns    []Pattern
	mu          sync.RWMutex
	running     bool
}

// Alert represents a watchdog alert
type Alert struct {
	ID          string
	Type        string // "pattern", "security", "dependency", "concept_drift"
	Severity    string // "info", "warning", "error"
	Title       string
	Message     string
	Context     map[string]interface{}
	Timestamp   time.Time
	Acknowledged bool
}

// Proposal represents an evolution proposal
type Proposal struct {
	ID          string
	Component   string
	Description string
	Changes     map[string]interface{}
	Status      string // "pending", "approved", "rejected"
	Reward      float64
	Feedback    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Pattern represents a detected pattern
type Pattern struct {
	ID          string
	Name        string
	Type        string
	Occurrences int
	FirstSeen   time.Time
	LastSeen    time.Time
	Context     map[string]interface{}
}

// NewWatchdog creates a new watchdog
func NewWatchdog(longTermMem *memory.LongTermMemory) *Watchdog {
	return &Watchdog{
		longTermMem: longTermMem,
		alerts:      make([]Alert, 0),
		proposals:   make(map[string]*Proposal),
		patterns:    make([]Pattern, 0),
		running:     false,
	}
}

// Start starts the watchdog monitoring
func (w *Watchdog) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return fmt.Errorf("watchdog already running")
	}

	w.running = true

	// Start monitoring goroutine
	go w.monitorLoop()

	return nil
}

// Stop stops the watchdog monitoring
func (w *Watchdog) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return fmt.Errorf("watchdog not running")
	}

	w.running = false
	return nil
}

// monitorLoop continuously monitors for patterns
func (w *Watchdog) monitorLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		w.mu.RLock()
		running := w.running
		w.mu.RUnlock()

		if !running {
			break
		}

		// Perform monitoring checks
		w.checkPatterns()
		w.checkSecurity()
		w.checkDependencies()

		<-ticker.C
	}
}

// DetectPattern detects patterns in code
func (w *Watchdog) DetectPattern(code string) ([]Alert, error) {
	alerts := make([]Alert, 0)

	// Check for authentication patterns
	if strings.Contains(code, "JWT") || strings.Contains(code, "auth") {
		alert := w.createAlert("pattern", "info", "Authentication Pattern",
			"Authentication pattern detected in code",
			map[string]interface{}{
				"pattern": "authentication",
				"code":    code[:min(len(code), 100)],
			})
		alerts = append(alerts, alert)
	}

	// Check for database patterns
	if strings.Contains(code, "SELECT") || strings.Contains(code, "INSERT") || strings.Contains(code, "UPDATE") {
		alert := w.createAlert("pattern", "warning", "SQL Query Detected",
			"Direct SQL query detected - consider using parameterized queries",
			map[string]interface{}{
				"pattern": "sql_query",
				"code":    code[:min(len(code), 100)],
			})
		alerts = append(alerts, alert)
	}

	// Check for API calls
	if strings.Contains(code, "http.") || strings.Contains(code, "fetch") || strings.Contains(code, "axios") {
		alert := w.createAlert("pattern", "info", "API Call Pattern",
			"API call pattern detected",
			map[string]interface{}{
				"pattern": "api_call",
				"code":    code[:min(len(code), 100)],
			})
		alerts = append(alerts, alert)
	}

	// Check for error handling
	if !strings.Contains(code, "try") && !strings.Contains(code, "catch") && !strings.Contains(code, "if err") {
		if strings.Contains(code, "func") || strings.Contains(code, "function") {
			alert := w.createAlert("pattern", "warning", "Missing Error Handling",
				"Function may lack proper error handling",
				map[string]interface{}{
					"pattern": "missing_error_handling",
					"code":    code[:min(len(code), 100)],
				})
			alerts = append(alerts, alert)
		}
	}

	// Store alerts
	w.mu.Lock()
	w.alerts = append(w.alerts, alerts...)
	w.mu.Unlock()

	return alerts, nil
}

// checkPatterns checks for emerging patterns
func (w *Watchdog) checkPatterns() {
	// TODO: Query long-term memory for pattern analysis
	// For now, this is a placeholder
}

// checkSecurity checks for security issues
func (w *Watchdog) checkSecurity() {
	// TODO: Implement security checks
	// - SQL injection patterns
	// - XSS vulnerabilities
	// - Exposed secrets
	// - Insecure dependencies
}

// checkDependencies checks for dependency issues
func (w *Watchdog) checkDependencies() {
	// TODO: Implement dependency checks
	// - Circular dependencies
	// - Outdated packages
	// - Unused dependencies
}

// SubmitProposal submits an evolution proposal
func (w *Watchdog) SubmitProposal(req models.ProposalRequest) (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	id := fmt.Sprintf("proposal_%d", time.Now().Unix())

	proposal := &Proposal{
		ID:          id,
		Component:   req.Component,
		Description: req.Description,
		Changes:     req.Changes,
		Status:      "pending",
		Reward:      0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	w.proposals[id] = proposal

	// Create alert for new proposal
	alert := w.createAlert("proposal", "info", "New Proposal",
		fmt.Sprintf("New evolution proposal for %s", req.Component),
		map[string]interface{}{
			"proposal_id": id,
			"component":   req.Component,
		})

	w.alerts = append(w.alerts, alert)

	return id, nil
}

// ApproveProposal approves a proposal
func (w *Watchdog) ApproveProposal(id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	proposal, exists := w.proposals[id]
	if !exists {
		return fmt.Errorf("proposal %s not found", id)
	}

	proposal.Status = "approved"
	proposal.UpdatedAt = time.Now()

	return nil
}

// RejectProposal rejects a proposal
func (w *Watchdog) RejectProposal(id string, reason string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	proposal, exists := w.proposals[id]
	if !exists {
		return fmt.Errorf("proposal %s not found", id)
	}

	proposal.Status = "rejected"
	proposal.Feedback = reason
	proposal.UpdatedAt = time.Now()

	return nil
}

// SetReward sets reward for a proposal
func (w *Watchdog) SetReward(req models.RewardRequest) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	proposal, exists := w.proposals[req.ProposalID]
	if !exists {
		return fmt.Errorf("proposal %s not found", req.ProposalID)
	}

	proposal.Reward = req.Reward
	if req.Feedback != "" {
		proposal.Feedback = req.Feedback
	}
	proposal.UpdatedAt = time.Now()

	return nil
}

// GetStatus returns watchdog status
func (w *Watchdog) GetStatus() interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return map[string]interface{}{
		"running":         w.running,
		"alerts_count":    len(w.alerts),
		"proposals_count": len(w.proposals),
		"patterns_count":  len(w.patterns),
		"recent_alerts":   w.getRecentAlerts(5),
	}
}

// GetAlerts returns all alerts
func (w *Watchdog) GetAlerts() []Alert {
	w.mu.RLock()
	defer w.mu.RUnlock()

	alerts := make([]Alert, len(w.alerts))
	copy(alerts, w.alerts)
	return alerts
}

// GetProposals returns all proposals
func (w *Watchdog) GetProposals() []*Proposal {
	w.mu.RLock()
	defer w.mu.RUnlock()

	proposals := make([]*Proposal, 0, len(w.proposals))
	for _, p := range w.proposals {
		proposals = append(proposals, p)
	}

	return proposals
}

// GetProposal returns a specific proposal
func (w *Watchdog) GetProposal(id string) (*Proposal, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	proposal, exists := w.proposals[id]
	if !exists {
		return nil, fmt.Errorf("proposal %s not found", id)
	}

	return proposal, nil
}

// AcknowledgeAlert acknowledges an alert
func (w *Watchdog) AcknowledgeAlert(id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i := range w.alerts {
		if w.alerts[i].ID == id {
			w.alerts[i].Acknowledged = true
			return nil
		}
	}

	return fmt.Errorf("alert %s not found", id)
}

// ClearAlerts clears all alerts
func (w *Watchdog) ClearAlerts() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.alerts = make([]Alert, 0)
}

// createAlert creates a new alert
func (w *Watchdog) createAlert(alertType, severity, title, message string, context map[string]interface{}) Alert {
	return Alert{
		ID:           fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Type:         alertType,
		Severity:     severity,
		Title:        title,
		Message:      message,
		Context:      context,
		Timestamp:    time.Now(),
		Acknowledged: false,
	}
}

// getRecentAlerts returns the N most recent alerts
func (w *Watchdog) getRecentAlerts(n int) []Alert {
	if n > len(w.alerts) {
		n = len(w.alerts)
	}

	if n == 0 {
		return []Alert{}
	}

	alerts := make([]Alert, n)
	copy(alerts, w.alerts[len(w.alerts)-n:])
	return alerts
}

// GetMetrics returns watchdog metrics
func (w *Watchdog) GetMetrics() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	totalAlerts := len(w.alerts)
	acknowledgedAlerts := 0
	alertsByType := make(map[string]int)
	alertsBySeverity := make(map[string]int)

	for _, alert := range w.alerts {
		if alert.Acknowledged {
			acknowledgedAlerts++
		}
		alertsByType[alert.Type]++
		alertsBySeverity[alert.Severity]++
	}

	proposalsByStatus := make(map[string]int)
	totalReward := 0.0

	for _, proposal := range w.proposals {
		proposalsByStatus[proposal.Status]++
		totalReward += proposal.Reward
	}

	return map[string]interface{}{
		"total_alerts":        totalAlerts,
		"acknowledged_alerts": acknowledgedAlerts,
		"alerts_by_type":      alertsByType,
		"alerts_by_severity":  alertsBySeverity,
		"total_proposals":     len(w.proposals),
		"proposals_by_status": proposalsByStatus,
		"total_reward":        totalReward,
		"patterns_detected":   len(w.patterns),
	}
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

