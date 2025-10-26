# OpenEvolve Integration Guide

## 🔍 Overview

**OpenEvolve** is the intelligent watchdog and self-evolution system integrated into the Agentic Command Center. It continuously monitors code patterns, detects security issues, tracks component evolution, and provides real-time alerts for code quality and architectural decisions.

Unlike traditional static analysis tools, OpenEvolve **learns from your codebase** and **evolves its detection patterns** based on your project's specific needs and conventions.

---

## 🎯 Core Capabilities

### 1. **Pattern Detection** 🕵️
Automatically identifies:
- Authentication patterns (JWT, OAuth, session-based)
- Database access patterns (SQL, NoSQL, ORM)
- API design patterns (REST, GraphQL, gRPC)
- Error handling strategies
- Caching mechanisms
- Security implementations

### 2. **Concept Wiring** 🧠
Understands relationships between:
- Functions and their purposes
- Modules and their dependencies
- Patterns and their implementations
- Concepts and their evolution over time

### 3. **Security Watchdog** 🛡️
Detects potential vulnerabilities:
- SQL injection risks
- XSS vulnerabilities
- Hardcoded secrets/credentials
- Insecure dependencies
- Missing input validation
- Unencrypted sensitive data

### 4. **Code Quality Monitoring** ✨
Tracks:
- Complexity metrics (cyclomatic complexity)
- Code duplication
- Missing error handling
- Inconsistent patterns
- Technical debt accumulation

### 5. **Component Evolution** 📈
Monitors component lifecycle:
- Pending → Review → Approved → Deployed
- Progress tracking
- Quality gates
- Approval workflows

---

## 🏗️ Architecture

### System Integration

```
┌─────────────────────────────────────────────────────────┐
│                  Code Changes                            │
│  (File created, modified, deleted)                       │
└───────────────────┬─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│              File Watcher                                │
│  Monitors filesystem for changes                         │
└───────────────────┬─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│           OpenEvolve Watchdog                            │
│  ┌──────────────────────────────────────────┐           │
│  │  Pattern Detector                        │           │
│  │  - Analyzes code structure               │           │
│  │  - Identifies patterns                   │           │
│  └──────────────────┬───────────────────────┘           │
│                     ↓                                    │
│  ┌──────────────────────────────────────────┐           │
│  │  Security Scanner                        │           │
│  │  - Checks for vulnerabilities            │           │
│  │  - Validates security practices          │           │
│  └──────────────────┬───────────────────────┘           │
│                     ↓                                    │
│  ┌──────────────────────────────────────────┐           │
│  │  Concept Analyzer                        │           │
│  │  - Understands code semantics            │           │
│  │  - Maps relationships                    │           │
│  └──────────────────┬───────────────────────┘           │
└────────────────────┼────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│              Alert Generation                            │
│  - Creates alerts with severity levels                   │
│  - Suggests fixes and improvements                       │
└───────────────────┬─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│         Neo4j Knowledge Graph                            │
│  - Stores patterns and concepts                          │
│  - Tracks evolution over time                            │
│  - Enables semantic queries                              │
└───────────────────┬─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│          Frontend UI (OpenEvolve Panel)                  │
│  - Displays alerts in real-time                          │
│  - Shows component progress                              │
│  - Provides approval interface                           │
└─────────────────────────────────────────────────────────┘
```

---

## 🔧 Implementation

### Backend Integration

**Location:** `backend/internal/watchdog/`

**Key Files:**
- `watchdog.go` - Main watchdog engine
- `alerts.go` - Alert generation and management
- `patterns.go` - Pattern detection logic
- `security.go` - Security scanning

**Initialization:**

```go
// backend/cmd/server/main.go
watchdog := watchdog.NewWatchdog(&watchdog.Config{
    Enabled:           true,
    ScanInterval:      time.Second * 30,
    MinConfidence:     0.7,
    AlertThreshold:    watchdog.SeverityWarning,
    Neo4jClient:       neo4jClient,
    LongTermMemory:    memorySystem,
})

// Start monitoring
watchdog.Start()

// Watch specific directories
watchdog.WatchDirectory("/home/ubuntu/agent-workspace/backend")
watchdog.WatchDirectory("/home/ubuntu/agent-workspace/frontend")
```

### Pattern Detection

**Example: Authentication Pattern Detection**

```go
// Detects JWT authentication implementation
func (w *Watchdog) DetectAuthPattern(file *ast.File) *Pattern {
    hasJWT := false
    hasMiddleware := false
    hasValidation := false
    
    // Analyze AST
    ast.Inspect(file, func(n ast.Node) bool {
        switch x := n.(type) {
        case *ast.ImportSpec:
            if strings.Contains(x.Path.Value, "jwt") {
                hasJWT = true
            }
        case *ast.FuncDecl:
            if strings.Contains(x.Name.Name, "Auth") {
                hasMiddleware = true
            }
            if strings.Contains(x.Name.Name, "Validate") {
                hasValidation = true
            }
        }
        return true
    })
    
    if hasJWT && hasMiddleware {
        return &Pattern{
            Type:       "Authentication",
            Subtype:    "JWT",
            Confidence: calculateConfidence(hasJWT, hasMiddleware, hasValidation),
            Components: []string{"JWT", "Middleware", "Validation"},
            File:       file.Name.Name,
        }
    }
    
    return nil
}
```

### Security Scanning

**Example: SQL Injection Detection**

```go
func (w *Watchdog) CheckSQLInjection(code string) *Alert {
    // Dangerous patterns
    dangerous := []string{
        `db.Query\(".*\+.*"\)`,           // String concatenation
        `db.Exec\(".*\+.*"\)`,
        `fmt.Sprintf\("SELECT.*%s`, // Format strings in SQL
    }
    
    for _, pattern := range dangerous {
        if matched, _ := regexp.MatchString(pattern, code); matched {
            return &Alert{
                Type:     "Security",
                Severity: SeverityHigh,
                Title:    "Potential SQL Injection",
                Message:  "SQL query uses string concatenation. Use parameterized queries instead.",
                File:     getCurrentFile(),
                Line:     getMatchLine(code, pattern),
                Suggestion: "Use: db.Query(\"SELECT * FROM users WHERE id = ?\", userID)",
            }
        }
    }
    
    return nil
}
```

### Concept Wiring

**Example: Mapping Function Relationships**

```go
func (w *Watchdog) WireConcepts(file *ast.File) {
    concepts := make(map[string]*Concept)
    
    // First pass: Identify concepts
    ast.Inspect(file, func(n ast.Node) bool {
        if fn, ok := n.(*ast.FuncDecl); ok {
            concept := w.IdentifyConcept(fn)
            concepts[fn.Name.Name] = concept
        }
        return true
    })
    
    // Second pass: Wire relationships
    ast.Inspect(file, func(n ast.Node) bool {
        if call, ok := n.(*ast.CallExpr); ok {
            if ident, ok := call.Fun.(*ast.Ident); ok {
                if concept, exists := concepts[ident.Name]; exists {
                    w.CreateRelationship(getCurrentFunction(), concept, "CALLS")
                }
            }
        }
        return true
    })
    
    // Store in Neo4j
    w.StoreConceptGraph(concepts)
}
```

---

## 📊 Alert System

### Alert Severity Levels

| Level | Color | Description | Example |
|-------|-------|-------------|---------|
| **Info** | 🔵 Blue | Informational | "New pattern detected: REST API" |
| **Warning** | 🟡 Yellow | Potential issue | "Function complexity: 15 (threshold: 10)" |
| **Error** | 🟠 Orange | Needs attention | "Missing error handling in HTTP handler" |
| **Critical** | 🔴 Red | Security/breaking | "SQL injection vulnerability detected" |

### Alert Structure

```go
type Alert struct {
    ID          string    `json:"id"`
    Type        string    `json:"type"`        // "Security", "Pattern", "Quality"
    Severity    Severity  `json:"severity"`
    Title       string    `json:"title"`
    Message     string    `json:"message"`
    File        string    `json:"file"`
    Line        int       `json:"line"`
    Column      int       `json:"column"`
    Suggestion  string    `json:"suggestion"`
    CodeSnippet string    `json:"code_snippet"`
    CreatedAt   time.Time `json:"created_at"`
    Acknowledged bool     `json:"acknowledged"`
}
```

### Real-time Alert Delivery

```go
// WebSocket broadcast to frontend
func (w *Watchdog) SendAlert(alert *Alert) {
    message := models.Message{
        ID:        uuid.New().String(),
        Type:      "watchdog_alert",
        Timestamp: time.Now(),
        Source:    "watchdog",
        Payload: map[string]interface{}{
            "alert": alert,
        },
    }
    
    // Broadcast to all connected clients
    websocket.BroadcastToAll(message)
    
    // Store in Neo4j for history
    w.StoreAlert(alert)
}
```

---

## 🎨 Frontend Integration

### OpenEvolve Panel

**Location:** `frontend/src/components/OpenEvolve/OpenEvolve.jsx`

**Features:**

1. **Components Tab**
   - Lists all components being developed
   - Shows approval status (✅ Approved, ⏳ Pending, ⚠️ Review)
   - Displays progress bars
   - Allows approval/rejection

2. **Watchdog Tab**
   - Real-time alert feed
   - Filterable by severity
   - Clickable to jump to code location
   - Acknowledgment system

**Example UI State:**

```jsx
const [components, setComponents] = useState([
  {
    id: '1',
    name: 'Authentication Module',
    status: 'approved',
    progress: 100,
    quality: 95,
    lastUpdated: 'TIMESTAMP'
  },
  {
    id: '2',
    name: 'Database Layer',
    status: 'pending',
    progress: 75,
    quality: 82,
    lastUpdated: 'TIMESTAMP'
  },
  {
    id: '3',
    name: 'API Endpoints',
    status: 'review',
    progress: 60,
    quality: 78,
    alerts: 3,
    lastUpdated: 'TIMESTAMP'
  }
]);

const [alerts, setAlerts] = useState([
  {
    id: 'a1',
    severity: 'critical',
    title: 'SQL Injection Risk',
    message: 'String concatenation in SQL query',
    file: 'backend/handlers/user.go',
    line: 45,
    timestamp: 'TIMESTAMP'
  },
  {
    id: 'a2',
    severity: 'warning',
    title: 'High Complexity',
    message: 'Function complexity: 18 (threshold: 10)',
    file: 'backend/agent/controller.go',
    line: 120,
    timestamp: 'TIMESTAMP'
  }
]);
```

---

## 📈 Component Lifecycle

### States

```
┌──────────┐
│  Pending │  ← Component created, awaiting review
└────┬─────┘
     ↓
┌──────────┐
│  Review  │  ← Under review, may have alerts
└────┬─────┘
     ↓
┌──────────┐
│ Approved │  ← Passed quality gates
└────┬─────┘
     ↓
┌──────────┐
│ Deployed │  ← In production
└──────────┘
```

### Approval Workflow

```go
func (w *Watchdog) ApproveComponent(componentID string, approver string) error {
    component, err := w.GetComponent(componentID)
    if err != nil {
        return err
    }
    
    // Check quality gates
    if component.Quality < w.config.MinQuality {
        return errors.New("component does not meet quality threshold")
    }
    
    // Check for critical alerts
    alerts := w.GetAlerts(componentID, SeverityCritical)
    if len(alerts) > 0 {
        return errors.New("component has unresolved critical alerts")
    }
    
    // Approve
    component.Status = "approved"
    component.ApprovedBy = approver
    component.ApprovedAt = time.Now()
    
    // Store in Neo4j
    w.UpdateComponent(component)
    
    // Notify frontend
    w.BroadcastComponentUpdate(component)
    
    return nil
}
```

---

## 🧠 Pattern Learning

### How OpenEvolve Learns

1. **Initial Patterns** - Starts with common patterns (JWT, REST, etc.)
2. **Observation** - Watches how you code
3. **Pattern Extraction** - Identifies your specific patterns
4. **Confidence Building** - Increases confidence as patterns repeat
5. **Evolution** - Adapts detection to your codebase

**Example Evolution:**

**Early Stage:**
```
Pattern: "Authentication"
Confidence: 0.6
Indicators: ["jwt", "token", "auth"]
```

**Week 4:**
```
Pattern: "Authentication"
Confidence: 0.95
Indicators: [
  "jwt",
  "token",
  "auth",
  "middleware.Auth",      // Learned from your code
  "ValidateJWT",          // Your specific function
  "claims.UserID",        // Your JWT structure
  "fiber.Ctx"             // Your framework choice
]
```

### Neo4j Pattern Storage

```cypher
// Store pattern
CREATE (p:Pattern {
  id: 'auth-jwt-001',
  type: 'Authentication',
  subtype: 'JWT',
  confidence: 0.95,
  indicators: ['jwt', 'token', 'middleware.Auth'],
  first_seen: datetime(),
  last_seen: datetime(),
  occurrence_count: 42
})

// Link to files
MATCH (p:Pattern {id: 'auth-jwt-001'})
MATCH (f:File {path: 'backend/middleware/auth.go'})
CREATE (p)-[:FOUND_IN {confidence: 0.98, line: 25}]->(f)

// Link to concept
MATCH (p:Pattern {id: 'auth-jwt-001'})
MATCH (c:Concept {name: 'Authentication'})
CREATE (p)-[:IMPLEMENTS]->(c)
```

---

## 🔍 Querying Patterns

### API Endpoints

**GET /api/watchdog/alerts**
Get all active alerts
```bash
curl http://localhost:8080/api/watchdog/alerts?severity=critical&acknowledged=false
```

**POST /api/watchdog/alerts/:id/acknowledge**
Acknowledge an alert
```bash
curl -X POST http://localhost:8080/api/watchdog/alerts/a1/acknowledge
```

**GET /api/watchdog/patterns**
Get detected patterns
```bash
curl http://localhost:8080/api/watchdog/patterns?type=Authentication&min_confidence=0.8
```

**GET /api/watchdog/components**
Get component status
```bash
curl http://localhost:8080/api/watchdog/components
```

**POST /api/watchdog/components/:id/approve**
Approve a component
```bash
curl -X POST http://localhost:8080/api/watchdog/components/comp-123/approve \
  -H "Content-Type: application/json" \
  -d '{"approver": "user@example.com"}'
```

### Neo4j Queries

**Find all security alerts:**
```cypher
MATCH (a:Alert {type: 'Security'})
WHERE a.severity IN ['high', 'critical']
  AND a.acknowledged = false
RETURN a.title, a.file, a.line, a.message
ORDER BY a.created_at DESC
```

**Find most common patterns:**
```cypher
MATCH (p:Pattern)
RETURN p.type, p.subtype, p.occurrence_count, p.confidence
ORDER BY p.occurrence_count DESC
LIMIT 10
```

**Find components with alerts:**
```cypher
MATCH (c:Component)-[:HAS_ALERT]->(a:Alert)
WHERE a.acknowledged = false
RETURN c.name, count(a) as alert_count, collect(a.severity) as severities
ORDER BY alert_count DESC
```

**Track pattern evolution:**
```cypher
MATCH (p:Pattern {type: 'Authentication'})
RETURN p.confidence, p.occurrence_count, p.last_seen
ORDER BY p.last_seen
```

---

## ⚙️ Configuration

### Environment Variables

```bash
# OpenEvolve Settings
WATCHDOG_ENABLED=true
WATCHDOG_SCAN_INTERVAL=30          # Seconds
WATCHDOG_MIN_CONFIDENCE=0.7        # Pattern confidence threshold
WATCHDOG_ALERT_THRESHOLD=warning   # Minimum severity to alert

# Quality Gates
QUALITY_MIN_SCORE=80               # Minimum quality score for approval
QUALITY_MAX_COMPLEXITY=10          # Maximum cyclomatic complexity
QUALITY_MIN_COVERAGE=75            # Minimum test coverage %

# Pattern Learning
PATTERN_LEARNING_ENABLED=true
PATTERN_MIN_OCCURRENCES=3          # Min occurrences to establish pattern
PATTERN_CONFIDENCE_THRESHOLD=0.8   # Min confidence to use pattern
```

### Backend Configuration

```go
watchdog := watchdog.NewWatchdog(&watchdog.Config{
    Enabled:              true,
    ScanInterval:         time.Second * 30,
    MinConfidence:        0.7,
    AlertThreshold:       watchdog.SeverityWarning,
    QualityMinScore:      80,
    QualityMaxComplexity: 10,
    PatternLearning:      true,
    Neo4jClient:          neo4jClient,
})
```

---

## 📊 Metrics & Analytics

### Dashboard Queries

**Alert trends:**
```cypher
MATCH (a:Alert)
WHERE a.created_at > datetime() - duration('P7D')
RETURN 
  date(a.created_at) as day,
  a.severity,
  count(*) as count
ORDER BY day, a.severity
```

**Component quality over time:**
```cypher
MATCH (c:Component)
RETURN 
  c.name,
  c.quality,
  c.progress,
  c.status,
  size((c)-[:HAS_ALERT]->(:Alert {acknowledged: false})) as active_alerts
ORDER BY c.quality DESC
```

**Pattern adoption rate:**
```cypher
MATCH (p:Pattern)
RETURN 
  p.type,
  p.occurrence_count,
  duration.between(p.first_seen, p.last_seen).days as days_active,
  p.occurrence_count / duration.between(p.first_seen, p.last_seen).days as adoption_rate
ORDER BY adoption_rate DESC
```

---

## 🎯 Best Practices

### 1. **Acknowledge Alerts Promptly**
- Review alerts daily
- Acknowledge or fix within 24 hours
- Don't let alerts pile up

### 2. **Set Appropriate Thresholds**
- Start with lower thresholds
- Adjust based on your codebase
- Balance strictness vs noise

### 3. **Trust the Learning**
- Let OpenEvolve learn your patterns
- Don't override too frequently
- Provide feedback on false positives

### 4. **Use Quality Gates**
- Require approval for critical components
- Set minimum quality scores
- Enforce test coverage

### 5. **Monitor Evolution**
- Check pattern confidence weekly
- Review learned patterns monthly
- Prune outdated patterns

---

## 🚀 Advanced Features

### Custom Pattern Definitions

```go
// Define custom pattern
customPattern := &watchdog.PatternDefinition{
    Name: "Custom API Pattern",
    Type: "API",
    Indicators: []watchdog.Indicator{
        {Type: "import", Value: "github.com/gofiber/fiber/v3"},
        {Type: "function", Pattern: `app\.(Get|Post|Put|Delete)`},
        {Type: "struct", Pattern: `type.*Handler struct`},
    },
    MinConfidence: 0.8,
}

watchdog.RegisterPattern(customPattern)
```

### Webhook Notifications

```go
// Send alerts to Slack/Discord
watchdog.OnAlert(func(alert *Alert) {
    if alert.Severity >= SeverityHigh {
        webhook.Send(alert.ToSlackMessage())
    }
})
```

### Integration with CI/CD

```bash
# In CI pipeline
curl -X POST http://localhost:8080/api/watchdog/scan \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/path/to/code",
    "fail_on_critical": true,
    "min_quality": 85
  }'
```

---

## 📚 Resources

- **OpenEvolve Concepts**: https://github.com/openevolve
- **Pattern Detection**: `backend/internal/watchdog/patterns.go`
- **Security Scanning**: `backend/internal/watchdog/security.go`
- **Frontend Panel**: `frontend/src/components/OpenEvolve/`

---

## 🤝 Contributing

To improve OpenEvolve:

1. Add new pattern detectors in `patterns.go`
2. Enhance security checks in `security.go`
3. Improve UI in `OpenEvolve.jsx`
4. Add metrics in Neo4j queries

---

## 📝 Summary

OpenEvolve provides:
- ✅ **Real-time monitoring** of code quality and patterns
- ✅ **Security scanning** for vulnerabilities
- ✅ **Pattern learning** that adapts to your codebase
- ✅ **Component tracking** with approval workflows
- ✅ **Concept wiring** for semantic understanding
- ✅ **Alert system** with actionable suggestions

**Result:** A self-improving watchdog that **learns your coding style**, **catches issues early**, and **guides evolution** toward better code quality. 🔍✨

