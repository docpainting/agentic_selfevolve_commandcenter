# OpenEvolve Evolutionary Coding Integration

This document explains how the **OpenEvolve evolutionary coding framework** is integrated into the agent workspace to enable true autonomous code evolution through reward-based learning.

## Overview

While `OPENEVOLVE_INTEGRATION.md` covers the watchdog and monitoring aspects, this document focuses on the **evolutionary coding** capabilities provided by the OpenEvolve Python library.

## What is OpenEvolve?

OpenEvolve is a **quality-diversity evolutionary coding framework** that uses:

- **MAP-Elites Algorithm**: Maintains diverse populations across feature dimensions
- **Island Model**: Multiple populations evolving independently with periodic migration
- **LLM Ensemble**: Uses Gemma 3 (via Ollama) for code generation and evaluation
- **Artifact Side-Channel**: Error feedback improves subsequent generations

## Key Concepts

### 1. Quality-Diversity Evolution

Unlike traditional optimization that finds a single "best" solution, OpenEvolve maintains a **diverse population** of solutions across multiple quality dimensions:

```
Complexity →
  Low    Medium    High
┌──────┬──────┬──────┐
│ ★    │ ★★   │      │  High Performance
├──────┼──────┼──────┤
│ ★★★  │ ★    │ ★★   │  Medium Performance
├──────┼──────┼──────┤
│ ★    │      │ ★    │  Low Performance
└──────┴──────┴──────┘
```

Each cell contains the **best solution** for that combination of features.

### 2. Island Model

Multiple populations evolve independently, preventing premature convergence:

```
Island 1        Island 2        Island 3
┌─────────┐    ┌─────────┐    ┌─────────┐
│ Pop: 100│    │ Pop: 100│    │ Pop: 100│
│ Gen: 50 │ ←→ │ Gen: 50 │ ←→ │ Gen: 50 │
│ Best: 85│    │ Best: 92│    │ Best: 78│
└─────────┘    └─────────┘    └─────────┘
     ↑              ↑              ↑
     └──────────────┴──────────────┘
          Migration every 50 gens
```

### 3. Reward-Based Learning

Code is evaluated and assigned scores based on:

- **Execution Success**: Does it work? (+20)
- **Pattern Detection**: Does it follow best practices? (+5 to +15 each)
- **Security**: Is it secure? (+10 to +15 each)
- **Performance**: Is it efficient? (+5 to +10 each)
- **Quality**: Is it maintainable? (+3 to +8 each)

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Agent Workspace (Go)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Watchdog   │  │   Evaluator  │  │   Evolver    │ │
│  │  (Monitor)   │  │  (Score)     │  │  (Trigger)   │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘ │
└─────────┼──────────────────┼──────────────────┼─────────┘
          │                  │                  │
          ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────┐
│              OpenEvolve (Python Library)                │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Evolution Engine                                │  │
│  │  - MAP-Elites (quality-diversity)                │  │
│  │  - Island model (parallel populations)           │  │
│  │  - LLM ensemble (Gemma 3 via Ollama)             │  │
│  │  - Cascade evaluation (early filtering)          │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
          │                  │                  │
          ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────┐
│                    Storage Layer                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Neo4j      │  │   BoltDB     │  │   ChromeM    │ │
│  │  (Graph)     │  │  (Evolution) │  │  (Vectors)   │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## Installation

OpenEvolve is installed as a Python library:

```bash
# Install OpenEvolve
pip install openevolve

# Verify installation
python -c "import openevolve; print(openevolve.__version__)"
```

## Configuration

The agent uses a custom configuration optimized for Gemma 3 via Ollama:

**Location**: `config/openevolve/agent_config.yaml`

**Key Settings**:

```yaml
# LLM configuration for Gemma 3 via Ollama
llm:
  models:
    - name: "gemma3:27b"
      weight: 1.0
  api_base: "http://localhost:11434/v1/"
  temperature: 0.7
  max_tokens: 8192

# Evolution parameters
max_iterations: 1000
diff_based_evolution: true
early_stopping_patience: 30

# Population settings
database:
  population_size: 500
  num_islands: 5
  migration_interval: 50
  migration_rate: 0.1
```

## Integration Points

### 1. Go → Python Bridge

The Go backend calls OpenEvolve via Python subprocess:

```go
// backend/internal/evolver/evolver.go
type Evolver struct {
    pythonPath string
    configPath string
}

func (e *Evolver) Evolve(ctx context.Context, request *EvolveRequest) (*EvolveResponse, error) {
    // 1. Write initial code to temp file
    initialFile := filepath.Join(e.tempDir, "initial.go")
    if err := os.WriteFile(initialFile, []byte(request.Code), 0644); err != nil {
        return nil, err
    }
    
    // 2. Write evaluator (scoring function)
    evaluatorFile := filepath.Join(e.tempDir, "evaluator.py")
    if err := e.writeEvaluator(evaluatorFile, request.Metrics); err != nil {
        return nil, err
    }
    
    // 3. Call OpenEvolve
    cmd := exec.CommandContext(ctx,
        e.pythonPath,
        "-m", "openevolve",
        initialFile,
        evaluatorFile,
        "--config", e.configPath,
        "--iterations", strconv.Itoa(request.Iterations),
        "--output", filepath.Join(e.tempDir, "evolved"),
    )
    
    // 4. Capture output
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("evolution failed: %w\n%s", err, output)
    }
    
    // 5. Read evolved code
    evolvedCode, err := os.ReadFile(filepath.Join(e.tempDir, "evolved", "best_program.go"))
    if err != nil {
        return nil, err
    }
    
    // 6. Parse results
    return e.parseResults(evolvedCode, output)
}
```

### 2. Custom Evaluator

The evaluator scores code based on agent-specific metrics:

```python
# Generated evaluator.py
import subprocess
import json
import re

def evaluate(program_path):
    """
    Evaluate a Go program based on agent-specific metrics.
    Returns a dict with scores for different dimensions.
    """
    scores = {}
    
    # 1. Execution test
    try:
        result = subprocess.run(
            ['go', 'run', program_path],
            capture_output=True,
            timeout=30,
            text=True
        )
        scores['execution'] = 20 if result.returncode == 0 else 0
    except subprocess.TimeoutExpired:
        scores['execution'] = -5
    except Exception as e:
        scores['execution'] = -10
    
    # 2. Read code
    with open(program_path, 'r') as f:
        code = f.read()
    
    # 3. Pattern detection
    scores['patterns'] = detect_patterns(code)
    
    # 4. Security checks
    scores['security'] = check_security(code)
    
    # 5. Quality metrics
    scores['quality'] = check_quality(code)
    
    # 6. Calculate combined score
    combined = sum(scores.values())
    
    return {
        'combined_score': combined,
        'execution': scores['execution'],
        'patterns': scores['patterns'],
        'security': scores['security'],
        'quality': scores['quality'],
        # Feature dimensions (for MAP-Elites diversity)
        'complexity': calculate_complexity(code),
        'diversity': calculate_diversity(code),
    }

def detect_patterns(code):
    """Detect best practice patterns"""
    score = 0
    
    # Authentication patterns
    if 'bcrypt' in code or 'argon2' in code:
        score += 10
    if 'jwt' in code.lower():
        score += 8
    
    # Error handling
    if 'if err != nil' in code:
        score += 5
    
    # Logging
    if 'log.' in code:
        score += 3
    
    return score

def check_security(code):
    """Check for security issues"""
    score = 0
    
    # SQL injection prevention
    if 'db.Query(' in code and '?' in code:
        score += 15  # Parameterized queries
    elif 'db.Query(' in code and '+' in code:
        score -= 20  # String concatenation (BAD)
    
    # Hardcoded secrets
    if re.search(r'password\s*=\s*["\']', code):
        score -= 15
    
    return score

def check_quality(code):
    """Check code quality"""
    score = 0
    
    # Documentation
    if '//' in code or '/*' in code:
        score += 4
    
    # Modularity (function count)
    func_count = code.count('func ')
    if 3 <= func_count <= 10:
        score += 6
    
    return score

def calculate_complexity(code):
    """Calculate code complexity (for MAP-Elites)"""
    return len(code)

def calculate_diversity(code):
    """Calculate structural diversity (for MAP-Elites)"""
    # Unique tokens
    tokens = set(re.findall(r'\b\w+\b', code))
    return len(tokens)
```

### 3. Evolution Workflow

```go
// backend/internal/agent/controller.go
func (a *AgentController) HandleTask(task *Task) error {
    // 1. Generate initial code
    initialCode := a.generateCode(task)
    
    // 2. Evaluate initial code
    initialScore := a.evaluator.Evaluate(initialCode, task.Metrics)
    
    // 3. If score is low, trigger evolution
    if initialScore.Total < 50 {
        log.Info("Score too low, triggering evolution")
        
        evolved, err := a.evolver.Evolve(context.Background(), &EvolveRequest{
            Code:       initialCode,
            Metrics:    task.Metrics,
            Iterations: 100,
        })
        if err != nil {
            return err
        }
        
        // 4. Use evolved code
        initialCode = evolved.BestCode
        initialScore = evolved.BestScore
    }
    
    // 5. Execute code
    result := a.executor.Execute(initialCode)
    
    // 6. Store in Neo4j
    a.neo4j.StoreEvolution(&Evolution{
        TaskID:       task.ID,
        InitialCode:  initialCode,
        FinalCode:    result.Code,
        InitialScore: initialScore,
        FinalScore:   result.Score,
        Generations:  evolved.Generations,
    })
    
    return nil
}
```

## Evolution Example

### Input: Create Login Endpoint

**Initial Code (Generated by Agent)**:

```go
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    username := r.FormValue("username")
    password := r.FormValue("password")
    
    // Query database
    var user User
    db.QueryRow("SELECT * FROM users WHERE username = '" + username + "'").Scan(&user)
    
    // Check password
    if user.Password == password {
        w.Write([]byte("Login successful"))
    } else {
        w.Write([]byte("Login failed"))
    }
}
```

**Evaluation**:
```yaml
Execution: +20 (compiles and runs)
Patterns: +0 (no good patterns detected)
Security: -35 (SQL injection + plaintext password)
Quality: +3 (has comments)
Total: -12 (VERY BAD)
```

**Evolution Triggered** (100 iterations):

**Generation 1-20**: Fix critical security issues
```go
// SQL injection fixed
db.QueryRow("SELECT * FROM users WHERE username = ?", username).Scan(&user)
// Score: +15 (SQL injection prevented)
```

**Generation 21-50**: Add password hashing
```go
// Password hashing added
import "golang.org/x/crypto/bcrypt"
err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
// Score: +10 (password hashing)
```

**Generation 51-80**: Add JWT tokens
```go
// JWT token generation
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "user_id": user.ID,
    "exp":     time.Now().Add(24 * time.Hour).Unix(),
})
// Score: +8 (JWT tokens)
```

**Generation 81-100**: Add rate limiting
```go
// Rate limiting
if !rateLimiter.Allow(r.RemoteAddr) {
    http.Error(w, "Too many requests", 429)
    return
}
// Score: +8 (rate limiting)
```

**Final Evolved Code**:

```go
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    // Rate limiting
    if !rateLimiter.Allow(r.RemoteAddr) {
        http.Error(w, "Too many requests", 429)
        return
    }
    
    username := r.FormValue("username")
    password := r.FormValue("password")
    
    // Input validation
    if username == "" || password == "" {
        http.Error(w, "Missing credentials", 400)
        return
    }
    
    // Query database (SQL injection prevented)
    var user User
    err := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = ?", username).Scan(
        &user.ID, &user.Username, &user.PasswordHash,
    )
    if err != nil {
        log.Error("Database error", "error", err)
        http.Error(w, "Login failed", 401)
        return
    }
    
    // Check password (bcrypt)
    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
    if err != nil {
        http.Error(w, "Login failed", 401)
        return
    }
    
    // Generate JWT token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    })
    
    tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
    if err != nil {
        log.Error("Token generation failed", "error", err)
        http.Error(w, "Internal error", 500)
        return
    }
    
    // Return token
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "token": tokenString,
    })
}
```

**Final Evaluation**:
```yaml
Execution: +20 (works perfectly)
Patterns: +26 (bcrypt +10, JWT +8, rate limiting +8)
Security: +30 (SQL injection prevented +15, password hashing +10, input validation +5)
Quality: +12 (error handling +5, logging +3, documentation +4)
Total: +88 (EXCELLENT - improvement of +100!)
```

## UI Integration

The right panel shows evolution progress in real-time:

```
┌─────────────────────────────────────┐
│       OpenEvolve Monitor            │
├─────────────────────────────────────┤
│                                     │
│ 🧬 Evolution in Progress            │
│                                     │
│ Generation: 47/100                  │
│ ████████████░░░░░░░░ 47%           │
│                                     │
│ Best Score: 73 (+61 from initial)   │
│                                     │
│ Population: 500 programs            │
│ Islands: 5                          │
│ Migrations: 0 (next at gen 50)      │
│                                     │
│ Recent Improvements:                │
│ • Gen 12: SQL injection fixed (+15) │
│ • Gen 23: Password hashing (+10)    │
│ • Gen 35: JWT tokens added (+8)     │
│ • Gen 47: Rate limiting (+8)        │
│                                     │
│ [Pause] [Stop] [Export Best]        │
│                                     │
└─────────────────────────────────────┘
```

## Best Practices

### 1. Start with Good Initial Code

OpenEvolve works best when starting from working code:

```go
// ✅ Good: Working but suboptimal
func Add(a, b int) int {
    return a + b
}

// ❌ Bad: Doesn't compile
func Add(a, b) {
    return a + b
}
```

### 2. Define Clear Metrics

Be specific about what you want to optimize:

```yaml
metrics:
  - name: "execution_success"
    weight: 0.4
  - name: "security_score"
    weight: 0.3
  - name: "performance"
    weight: 0.2
  - name: "code_quality"
    weight: 0.1
```

### 3. Use Cascade Evaluation

Filter bad solutions early to save LLM calls:

```yaml
evaluator:
  cascade_evaluation: true
  cascade_thresholds:
    - 0.4  # Must score >40% to continue
    - 0.7  # Must score >70% for full eval
    - 0.9  # Must score >90% for final eval
```

### 4. Monitor Evolution

Watch the UI to see if evolution is progressing:

- **Stuck at same score**: Increase population size or temperature
- **Score decreasing**: Check evaluator for bugs
- **No improvement after 50 gens**: Stop and adjust config

## Troubleshooting

### Evolution Not Improving

**Problem**: Score stuck at same value

**Solutions**:
- Increase `temperature` (more creativity)
- Increase `population_size` (more diversity)
- Add more `feature_dimensions` (explore different areas)
- Check evaluator for bugs

### Evolution Too Slow

**Problem**: Takes too long per generation

**Solutions**:
- Reduce `population_size`
- Enable `cascade_evaluation`
- Reduce `max_tokens`
- Use faster model (gemini-2.5-flash)

### Evolved Code Breaks

**Problem**: Evolved code doesn't compile or run

**Solutions**:
- Add compilation check to evaluator
- Penalize syntax errors heavily (-50)
- Use `diff_based_evolution: true` (safer)
- Start with better initial code

## Advanced: Custom Feature Dimensions

You can define custom dimensions for MAP-Elites:

```yaml
database:
  feature_dimensions:
    - "performance"      # Custom: execution time
    - "memory_usage"     # Custom: memory footprint
    - "code_length"      # Built-in: lines of code
```

Then return these in your evaluator:

```python
def evaluate(program_path):
    # ... existing code ...
    
    return {
        'combined_score': combined,
        # Feature dimensions
        'performance': measure_execution_time(program_path),
        'memory_usage': measure_memory_usage(program_path),
        'code_length': count_lines(program_path),
    }
```

## See Also

- [OPENEVOLVE_INTEGRATION.md](OPENEVOLVE_INTEGRATION.md) - Watchdog and monitoring integration
- [SELF_MODIFICATION.md](SELF_MODIFICATION.md) - How the agent rewrites itself
- [PROPAGATION.md](PROPAGATION.md) - Intelligence propagation philosophy
- [OpenEvolve GitHub](https://github.com/codelion/openevolve) - Official repository

