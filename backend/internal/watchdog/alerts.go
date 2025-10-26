package watchdog

import (
	"fmt"
	"strings"
	"time"
)

// AlertType constants
const (
	AlertTypePattern       = "pattern"
	AlertTypeSecurity      = "security"
	AlertTypeDependency    = "dependency"
	AlertTypeConceptDrift  = "concept_drift"
	AlertTypePerformance   = "performance"
	AlertTypeProposal      = "proposal"
)

// AlertSeverity constants
const (
	AlertSeverityInfo    = "info"
	AlertSeverityWarning = "warning"
	AlertSeverityError   = "error"
)

// AlertGenerator generates alerts based on various conditions
type AlertGenerator struct {
	watchdog *Watchdog
}

// NewAlertGenerator creates a new alert generator
func NewAlertGenerator(watchdog *Watchdog) *AlertGenerator {
	return &AlertGenerator{
		watchdog: watchdog,
	}
}

// GenerateSecurityAlert generates a security alert
func (g *AlertGenerator) GenerateSecurityAlert(title, message string, context map[string]interface{}) Alert {
	return g.watchdog.createAlert(AlertTypeSecurity, AlertSeverityError, title, message, context)
}

// GeneratePerformanceAlert generates a performance alert
func (g *AlertGenerator) GeneratePerformanceAlert(title, message string, context map[string]interface{}) Alert {
	return g.watchdog.createAlert(AlertTypePerformance, AlertSeverityWarning, title, message, context)
}

// GenerateConceptDriftAlert generates a concept drift alert
func (g *AlertGenerator) GenerateConceptDriftAlert(title, message string, context map[string]interface{}) Alert {
	return g.watchdog.createAlert(AlertTypeConceptDrift, AlertSeverityInfo, title, message, context)
}

// AnalyzeCode analyzes code for various issues
func (g *AlertGenerator) AnalyzeCode(code, filename string) []Alert {
	alerts := make([]Alert, 0)

	// Security checks
	alerts = append(alerts, g.checkSQLInjection(code, filename)...)
	alerts = append(alerts, g.checkHardcodedSecrets(code, filename)...)
	alerts = append(alerts, g.checkXSS(code, filename)...)

	// Pattern checks
	alerts = append(alerts, g.checkErrorHandling(code, filename)...)
	alerts = append(alerts, g.checkComplexity(code, filename)...)

	// Dependency checks
	alerts = append(alerts, g.checkImports(code, filename)...)

	return alerts
}

// checkSQLInjection checks for SQL injection vulnerabilities
func (g *AlertGenerator) checkSQLInjection(code, filename string) []Alert {
	alerts := make([]Alert, 0)

	patterns := []string{
		"SELECT * FROM",
		"INSERT INTO",
		"UPDATE",
		"DELETE FROM",
	}

	for _, pattern := range patterns {
		if strings.Contains(code, pattern) && strings.Contains(code, "+") {
			alert := g.GenerateSecurityAlert(
				"Potential SQL Injection",
				fmt.Sprintf("SQL query with string concatenation detected in %s", filename),
				map[string]interface{}{
					"file":    filename,
					"pattern": pattern,
					"recommendation": "Use parameterized queries or prepared statements",
				},
			)
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// checkHardcodedSecrets checks for hardcoded secrets
func (g *AlertGenerator) checkHardcodedSecrets(code, filename string) []Alert {
	alerts := make([]Alert, 0)

	patterns := []string{
		"password",
		"api_key",
		"secret",
		"token",
		"private_key",
	}

	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(code), pattern) && strings.Contains(code, "=") {
			alert := g.GenerateSecurityAlert(
				"Potential Hardcoded Secret",
				fmt.Sprintf("Possible hardcoded secret detected in %s", filename),
				map[string]interface{}{
					"file":    filename,
					"pattern": pattern,
					"recommendation": "Use environment variables or secure vaults",
				},
			)
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// checkXSS checks for XSS vulnerabilities
func (g *AlertGenerator) checkXSS(code, filename string) []Alert {
	alerts := make([]Alert, 0)

	if strings.Contains(code, "innerHTML") || strings.Contains(code, "dangerouslySetInnerHTML") {
		alert := g.GenerateSecurityAlert(
			"Potential XSS Vulnerability",
			fmt.Sprintf("Direct HTML injection detected in %s", filename),
			map[string]interface{}{
				"file":    filename,
				"recommendation": "Sanitize user input and use safe DOM methods",
			},
		)
		alerts = append(alerts, alert)
	}

	return alerts
}

// checkErrorHandling checks for missing error handling
func (g *AlertGenerator) checkErrorHandling(code, filename string) []Alert {
	alerts := make([]Alert, 0)

	// Check for functions without error handling
	if (strings.Contains(code, "func") || strings.Contains(code, "function")) &&
		!strings.Contains(code, "try") &&
		!strings.Contains(code, "catch") &&
		!strings.Contains(code, "if err") {

		alert := g.watchdog.createAlert(
			AlertTypePattern,
			AlertSeverityWarning,
			"Missing Error Handling",
			fmt.Sprintf("Function in %s may lack proper error handling", filename),
			map[string]interface{}{
				"file":    filename,
				"recommendation": "Add try-catch blocks or error checking",
			},
		)
		alerts = append(alerts, alert)
	}

	return alerts
}

// checkComplexity checks for code complexity
func (g *AlertGenerator) checkComplexity(code, filename string) []Alert {
	alerts := make([]Alert, 0)

	// Simple complexity check based on nesting level
	maxNesting := 0
	currentNesting := 0

	for _, char := range code {
		if char == '{' {
			currentNesting++
			if currentNesting > maxNesting {
				maxNesting = currentNesting
			}
		} else if char == '}' {
			currentNesting--
		}
	}

	if maxNesting > 5 {
		alert := g.GeneratePerformanceAlert(
			"High Code Complexity",
			fmt.Sprintf("High nesting level (%d) detected in %s", maxNesting, filename),
			map[string]interface{}{
				"file":           filename,
				"nesting_level":  maxNesting,
				"recommendation": "Consider refactoring to reduce complexity",
			},
		)
		alerts = append(alerts, alert)
	}

	return alerts
}

// checkImports checks for dependency issues
func (g *AlertGenerator) checkImports(code, filename string) []Alert {
	alerts := make([]Alert, 0)

	// Check for unused imports (basic check)
	lines := strings.Split(code, "\n")
	imports := make([]string, 0)

	inImportBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "import") {
			inImportBlock = true
			continue
		}

		if inImportBlock {
			if trimmed == ")" {
				inImportBlock = false
				continue
			}

			if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
				// Extract package name
				parts := strings.Fields(trimmed)
				if len(parts) > 0 {
					pkg := strings.Trim(parts[len(parts)-1], `"`)
					imports = append(imports, pkg)
				}
			}
		}
	}

	// Check if imports are used (very basic check)
	for _, imp := range imports {
		pkgName := imp
		if idx := strings.LastIndex(imp, "/"); idx != -1 {
			pkgName = imp[idx+1:]
		}

		if !strings.Contains(code, pkgName+".") {
			alert := g.watchdog.createAlert(
				AlertTypeDependency,
				AlertSeverityInfo,
				"Potentially Unused Import",
				fmt.Sprintf("Import '%s' may be unused in %s", imp, filename),
				map[string]interface{}{
					"file":   filename,
					"import": imp,
					"recommendation": "Remove unused imports to reduce dependencies",
				},
			)
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// MonitorFileChanges monitors file changes and generates alerts
func (g *AlertGenerator) MonitorFileChanges(filepath string, oldContent, newContent string) []Alert {
	alerts := make([]Alert, 0)

	// Check for significant changes
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	addedLines := len(newLines) - len(oldLines)

	if addedLines > 100 {
		alert := g.watchdog.createAlert(
			AlertTypePattern,
			AlertSeverityInfo,
			"Large Code Addition",
			fmt.Sprintf("Large code addition (%d lines) in %s", addedLines, filepath),
			map[string]interface{}{
				"file":         filepath,
				"added_lines":  addedLines,
				"recommendation": "Consider breaking into smaller commits",
			},
		)
		alerts = append(alerts, alert)
	}

	// Analyze new content
	alerts = append(alerts, g.AnalyzeCode(newContent, filepath)...)

	return alerts
}

// GenerateConceptWiringAlert generates an alert for concept wiring
func (g *AlertGenerator) GenerateConceptWiringAlert(concept1, concept2, relationship string) Alert {
	return g.GenerateConceptDriftAlert(
		"Concept Wiring Detected",
		fmt.Sprintf("New relationship detected: %s %s %s", concept1, relationship, concept2),
		map[string]interface{}{
			"concept1":     concept1,
			"concept2":     concept2,
			"relationship": relationship,
			"timestamp":    time.Now().Format(time.RFC3339),
		},
	)
}

// GeneratePatternEmergenceAlert generates an alert for emerging patterns
func (g *AlertGenerator) GeneratePatternEmergenceAlert(patternName string, occurrences int) Alert {
	return g.watchdog.createAlert(
		AlertTypePattern,
		AlertSeverityInfo,
		"Pattern Emergence",
		fmt.Sprintf("Pattern '%s' detected %d times", patternName, occurrences),
		map[string]interface{}{
			"pattern":     patternName,
			"occurrences": occurrences,
			"timestamp":   time.Now().Format(time.RFC3339),
		},
	)
}

// GenerateDependencyAlert generates a dependency-related alert
func (g *AlertGenerator) GenerateDependencyAlert(title, message string, context map[string]interface{}) Alert {
	return g.watchdog.createAlert(AlertTypeDependency, AlertSeverityWarning, title, message, context)
}

