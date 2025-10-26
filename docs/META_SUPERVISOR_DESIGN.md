# Meta-Supervisor Design
## Immutable Safety Layer Based on Metamorph Principles

Inspired by the Metamorph project's proven approach to safe self-modification.

---

## Overview

The **Meta-Supervisor** is an immutable layer that:
1. Cannot be modified by the evolution process
2. Monitors all self-modifications
3. Automatically rollbacks broken changes
4. Enforces safety boundaries
5. Provides emergency recovery

---

## Architecture

### Components

```
Meta-Supervisor (Immutable - Go)
├── Supervisor Process (cannot be evolved)
│   ├── Process Monitor
│   ├── Git-Based Rollback
│   ├── Safety Validator
│   └── Emergency Recovery
├── Evolution Coordinator
│   ├── Strategy Selector
│   ├── Change Approver
│   └── Performance Tracker
└── MCP Server Manager
    ├── Server Lifecycle
    ├── Health Checks
    └── Restart Logic
```

---

## Implementation (Go)

### 1. Supervisor Process

**File**: `backend/supervisor/supervisor.go`

```go
package supervisor

import (
    "context"
    "os/exec"
    "time"
    "github.com/go-git/go-git/v5"
)

type Supervisor struct {
    repo *git.Repository
    config *SupervisorConfig
    healthChecker *HealthChecker
}

type SupervisorConfig struct {
    MinRuntime time.Duration // If process dies before this, rollback
    MaxFailures int          // Max consecutive failures before emergency stop
    ProtectedPaths []string  // Files that cannot be modified
}

func NewSupervisor(repoPath string, config *SupervisorConfig) (*Supervisor, error) {
    repo, err := git.PlainOpen(repoPath)
    if err != nil {
        return nil, err
    }
    
    return &Supervisor{
        repo: repo,
        config: config,
        healthChecker: NewHealthChecker(),
    }, nil
}

func (s *Supervisor) Run(ctx context.Context) error {
    consecutiveFailures := 0
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Commit current state before running
            commitHash, err := s.commitCurrentState()
            if err != nil {
                return err
            }
            
            // Run the evolution process
            startTime := time.Now()
            err = s.runEvolutionProcess(ctx)
            runtime := time.Since(startTime)
            
            if err != nil {
                // Process failed
                if runtime < s.config.MinRuntime {
                    // Failed too quickly - likely broken by last change
                    log.Warn("Process failed quickly, rolling back last commit")
                    if err := s.rollbackToCommit(commitHash); err != nil {
                        return err
                    }
                    consecutiveFailures++
                } else {
                    // Failed after running for a while - might be normal
                    log.Info("Process exited after normal runtime")
                    consecutiveFailures = 0
                }
                
                if consecutiveFailures >= s.config.MaxFailures {
                    return fmt.Errorf("too many consecutive failures, stopping")
                }
            } else {
                // Process succeeded
                consecutiveFailures = 0
            }
        }
    }
}

func (s *Supervisor) runEvolutionProcess(ctx context.Context) error {
    // Run the Python MCP servers
    cmd := exec.CommandContext(ctx, "python", "backend/mcp_servers/dynamic_thinking/server.py")
    
    // Monitor health
    go s.healthChecker.Monitor(cmd.Process.Pid)
    
    return cmd.Run()
}

func (s *Supervisor) commitCurrentState() (string, error) {
    w, err := s.repo.Worktree()
    if err != nil {
        return "", err
    }
    
    // Add all changes
    _, err = w.Add(".")
    if err != nil {
        return "", err
    }
    
    // Commit
    commit, err := w.Commit("Auto-commit before evolution", &git.CommitOptions{
        Author: &object.Signature{
            Name: "Meta-Supervisor",
            When: time.Now(),
        },
    })
    
    return commit.String(), err
}

func (s *Supervisor) rollbackToCommit(commitHash string) error {
    w, err := s.repo.Worktree()
    if err != nil {
        return err
    }
    
    hash := plumbing.NewHash(commitHash)
    
    return w.Reset(&git.ResetOptions{
        Commit: hash,
        Mode:   git.HardReset,
    })
}

func (s *Supervisor) validateChange(change *ProposedChange) error {
    // Check if change modifies protected files
    for _, path := range change.ModifiedFiles {
        if s.isProtected(path) {
            return fmt.Errorf("cannot modify protected file: %s", path)
        }
    }
    
    // Run safety checks
    if err := s.runSafetyChecks(change); err != nil {
        return err
    }
    
    return nil
}

func (s *Supervisor) isProtected(path string) bool {
    for _, protected := range s.config.ProtectedPaths {
        if strings.HasPrefix(path, protected) {
            return true
        }
    }
    return false
}
```

---

### 2. Health Checker

**File**: `backend/supervisor/health_checker.go`

```go
package supervisor

import (
    "time"
    "syscall"
)

type HealthChecker struct {
    checks []HealthCheck
}

type HealthCheck interface {
    Check(pid int) error
}

func NewHealthChecker() *HealthChecker {
    return &HealthChecker{
        checks: []HealthCheck{
            &ProcessAliveCheck{},
            &MemoryCheck{maxMemoryMB: 4096},
            &CPUCheck{maxCPUPercent: 90},
        },
    }
}

func (hc *HealthChecker) Monitor(pid int) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        for _, check := range hc.checks {
            if err := check.Check(pid); err != nil {
                log.Error("Health check failed", "error", err)
                // Kill process if unhealthy
                syscall.Kill(pid, syscall.SIGTERM)
                return
            }
        }
    }
}

type ProcessAliveCheck struct{}

func (c *ProcessAliveCheck) Check(pid int) error {
    process, err := os.FindProcess(pid)
    if err != nil {
        return err
    }
    
    // Send signal 0 to check if process exists
    err = process.Signal(syscall.Signal(0))
    if err != nil {
        return fmt.Errorf("process not alive: %w", err)
    }
    
    return nil
}

type MemoryCheck struct {
    maxMemoryMB int64
}

func (c *MemoryCheck) Check(pid int) error {
    // Read /proc/{pid}/status for memory usage
    // Implementation details...
    return nil
}

type CPUCheck struct {
    maxCPUPercent float64
}

func (c *CPUCheck) Check(pid int) error {
    // Read /proc/{pid}/stat for CPU usage
    // Implementation details...
    return nil
}
```

---

### 3. Evolution Coordinator

**File**: `backend/supervisor/evolution_coordinator.go`

```go
package supervisor

type EvolutionCoordinator struct {
    strategySelector *StrategySelector
    changeApprover *ChangeApprover
    performanceTracker *PerformanceTracker
}

type EvolutionStrategy string

const (
    MetamorphicImprovement EvolutionStrategy = "metamorphic"
    ExploratoryEvolution   EvolutionStrategy = "exploratory"
    HybridOptimization     EvolutionStrategy = "hybrid"
    TrainingBased          EvolutionStrategy = "training"
)

func (ec *EvolutionCoordinator) SelectStrategy() (EvolutionStrategy, error) {
    // Get current performance metrics
    metrics := ec.performanceTracker.GetCurrentMetrics()
    
    // Select strategy based on metrics
    if metrics.ConsecutiveFailures > 3 {
        // Too many failures, use safe metamorphic approach
        return MetamorphicImprovement, nil
    }
    
    if metrics.PerformanceGain < 0.05 {
        // Stagnating, try exploratory
        return ExploratoryEvolution, nil
    }
    
    // Default to hybrid
    return HybridOptimization, nil
}

func (ec *EvolutionCoordinator) ApproveChange(change *ProposedChange) (bool, error) {
    // Run safety checks
    if err := ec.changeApprover.Validate(change); err != nil {
        return false, err
    }
    
    // Check if change aligns with current strategy
    strategy, _ := ec.SelectStrategy()
    if !ec.changeApprover.AlignsWith(change, strategy) {
        return false, nil
    }
    
    return true, nil
}
```

---

## Metamorph-Inspired Three-Stage Process

### Stage 1: Proposal
```
Agent generates short, concise proposal:
"Add caching layer to reduce database queries"
```

### Stage 2: Expansion
```
Agent expands proposal into concrete steps:
1. Create cache manager in storage/cache.go
2. Modify query functions to check cache first
3. Add cache invalidation on writes
4. Update tests
```

### Stage 3: Edit
```
Agent returns JSON with file changes:
{
  "changes": [
    {
      "file": "storage/cache.go",
      "content": "package storage\n\n..."
    },
    {
      "file": "storage/query.go",
      "content": "..."
    }
  ]
}
```

---

## Safety Mechanisms

### 1. Protected Files (Cannot Be Modified)
```go
protectedPaths := []string{
    "backend/supervisor/",           // Supervisor itself
    "backend/internal/core/",        // Core business logic
    "config/safety_rules.yaml",      // Safety configuration
    ".git/",                         // Git metadata
}
```

### 2. Automatic Rollback Triggers
- Process exits in < 10 seconds
- Health check failures (memory, CPU)
- Compilation errors
- Test failures
- Safety rule violations

### 3. Emergency Recovery
```go
func (s *Supervisor) EmergencyRecovery() error {
    // 1. Stop all processes
    s.stopAllProcesses()
    
    // 2. Rollback to last known good state
    if err := s.rollbackToLastGood(); err != nil {
        // 3. If rollback fails, restore from backup
        return s.restoreFromBackup()
    }
    
    return nil
}
```

---

## Integration with MCP Servers

### MCP Server Lifecycle Management

```go
type MCPServerManager struct {
    servers map[string]*MCPServer
}

type MCPServer struct {
    Name string
    Command string
    Args []string
    Process *os.Process
}

func (m *MCPServerManager) StartServer(name string) error {
    server := m.servers[name]
    
    cmd := exec.Command(server.Command, server.Args...)
    
    if err := cmd.Start(); err != nil {
        return err
    }
    
    server.Process = cmd.Process
    
    // Monitor health
    go m.monitorServer(server)
    
    return nil
}

func (m *MCPServerManager) RestartServer(name string) error {
    if err := m.StopServer(name); err != nil {
        return err
    }
    
    time.Sleep(1 * time.Second)
    
    return m.StartServer(name)
}
```

---

## Configuration

**File**: `config/supervisor.yaml`

```yaml
supervisor:
  min_runtime: 10s
  max_failures: 5
  
  protected_paths:
    - backend/supervisor/
    - backend/internal/core/
    - config/safety_rules.yaml
    - .git/
  
  health_checks:
    interval: 5s
    max_memory_mb: 4096
    max_cpu_percent: 90
  
  rollback:
    enabled: true
    keep_commits: 10
    backup_interval: 1h
  
  mcp_servers:
    - name: dynamic-thinking
      command: python
      args: [backend/mcp_servers/dynamic_thinking/server.py]
      restart_on_failure: true
    
    - name: openevolve
      command: python
      args: [backend/mcp_servers/openevolve/server.py]
      restart_on_failure: true
    
    - name: terminal-agent
      command: python
      args: [backend/mcp_servers/terminal/server.py]
      restart_on_failure: true
```

---

## Usage

### Start Supervisor
```bash
# Start the meta-supervisor
go run backend/supervisor/main.go

# The supervisor will:
# 1. Commit current state
# 2. Start MCP servers
# 3. Monitor health
# 4. Rollback if failures occur
# 5. Repeat forever
```

### Monitor Status
```bash
# Check supervisor status
curl http://localhost:9000/supervisor/status

# Response:
{
  "status": "running",
  "current_commit": "abc123",
  "consecutive_failures": 0,
  "uptime": "2h34m",
  "mcp_servers": {
    "dynamic-thinking": "healthy",
    "openevolve": "healthy",
    "terminal-agent": "healthy"
  }
}
```

---

## Summary

### Key Features

✅ **Immutable Supervisor** - Cannot be modified by evolution  
✅ **Automatic Rollback** - Git-based recovery from failures  
✅ **Health Monitoring** - Memory, CPU, process checks  
✅ **Protected Files** - Core components cannot be changed  
✅ **Emergency Recovery** - Restore from backup if needed  
✅ **MCP Server Management** - Lifecycle control  

### Safety Guarantees

1. **Supervisor cannot modify itself** - Protected path
2. **Automatic rollback on quick failures** - < 10s runtime
3. **Health checks prevent runaway processes** - Memory/CPU limits
4. **Emergency stop after too many failures** - Max 5 consecutive
5. **Git-based version control** - Always can rollback

### Metamorph Lessons Applied

1. ✅ **Immutable supervisor** - Learned from Metamorph's supervisor.js
2. ✅ **Three-stage prompting** - Proposal → Expansion → Edit
3. ✅ **Automatic rollback** - Time-based failure detection
4. ✅ **Modular approach** - One function per file
5. ✅ **Protected components** - Supervisor cannot be evolved

---

## Next Steps

1. Implement `backend/supervisor/` in Go
2. Create health check system
3. Integrate with MCP servers
4. Test rollback mechanisms
5. Add monitoring dashboard

This provides the **critical safety layer** needed before enabling full self-evolution! 🛡️

