// Package condition provides key condition extraction for branch analysis.
//
// Key conditions help determine:
// - What guards exist on the path to data usage
// - What constraints exist on user input
// - Control flow conditions affecting data flow
package condition

import (
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
	"github.com/hatlesswizard/inputtracer/pkg/sources/constants"
	"github.com/hatlesswizard/inputtracer/pkg/sources/patterns"
)

// Re-export types from centralized location for backward compatibility
type ConditionType = constants.ConditionType
type ConditionEffect = constants.ConditionEffect

// Re-export constants for backward compatibility
const (
	CondTypeComparison  = constants.CondTypeComparison
	CondTypeNullCheck   = constants.CondTypeNullCheck
	CondTypeTypeCheck   = constants.CondTypeTypeCheck
	CondTypeLengthCheck = constants.CondTypeLengthCheck
	CondTypeLogical     = constants.CondTypeLogical
	CondTypeUnknown     = constants.CondTypeUnknown
)

const (
	EffectAllows  = constants.EffectAllows
	EffectBlocks  = constants.EffectBlocks
	EffectUnknown = constants.EffectUnknown
)

// KeyCondition represents a condition that guards code execution
type KeyCondition struct {
	// Location
	FilePath   string `json:"file_path"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`

	// Condition details
	Expression    string          `json:"expression"` // The condition expression
	Type          ConditionType   `json:"type"`
	Effect        ConditionEffect `json:"effect"`
	IsNegated     bool            `json:"is_negated"` // If inside else or has !

	// Variables involved
	Variables   []string `json:"variables"`    // Variables referenced
	TaintedVars []string `json:"tainted_vars"` // Which are tainted

	// Scope
	GuardsLines     []int         `json:"guards_lines"`              // Lines guarded by this condition
	NestingDepth    int           `json:"nesting_depth"`             // How nested this condition is
	ParentCondition *KeyCondition `json:"parent_condition,omitempty"`
}

// ConditionPath represents a path through conditions to reach a point
type ConditionPath struct {
	Conditions []*KeyCondition `json:"conditions"`
	TargetLine int             `json:"target_line"`
	TargetExpr string          `json:"target_expr,omitempty"`
	Feasible   bool            `json:"feasible"`      // Is this path feasible?
	Reason     string          `json:"reason,omitempty"` // Why infeasible if not
}

// Extractor extracts key conditions from code
type Extractor struct {
	// Language-specific settings
	language string
}

// NewExtractor creates a new condition extractor for a language
func NewExtractor(language string) *Extractor {
	return &Extractor{
		language: language,
	}
}

// ExtractFromCode extracts key conditions from code text
func (e *Extractor) ExtractFromCode(code string, filePath string) []*KeyCondition {
	var conditions []*KeyCondition

	lines := strings.Split(code, "\n")
	nestingStack := make([]*KeyCondition, 0)

	for lineNum, line := range lines {
		// Detect condition statements
		if cond := e.detectCondition(line, lineNum+1, filePath, nestingStack); cond != nil {
			// Calculate guarded lines (simple heuristic)
			cond.GuardsLines = e.estimateGuardedLines(lines, lineNum)
			cond.NestingDepth = len(nestingStack)
			if len(nestingStack) > 0 {
				cond.ParentCondition = nestingStack[len(nestingStack)-1]
			}
			conditions = append(conditions, cond)

			// Update nesting stack
			if strings.Contains(line, "{") {
				nestingStack = append(nestingStack, cond)
			}
		}

		// Track block endings
		if strings.Contains(line, "}") && len(nestingStack) > 0 {
			// Pop from stack (simple heuristic - count braces)
			openBraces := strings.Count(line, "{")
			closeBraces := strings.Count(line, "}")
			for i := 0; i < closeBraces-openBraces && len(nestingStack) > 0; i++ {
				nestingStack = nestingStack[:len(nestingStack)-1]
			}
		}
	}

	return conditions
}

// detectCondition checks if a line contains a condition
func (e *Extractor) detectCondition(line string, lineNum int, filePath string, stack []*KeyCondition) *KeyCondition {
	trimmed := strings.TrimSpace(line)

	// Skip non-condition lines
	if !e.isConditionLine(trimmed) {
		return nil
	}

	// Extract the condition expression
	expr := e.extractConditionExpression(trimmed)
	if expr == "" {
		return nil
	}

	cond := &KeyCondition{
		FilePath:   filePath,
		Line:       lineNum,
		Expression: expr,
		Variables:  e.extractVariables(expr),
		IsNegated:  strings.Contains(trimmed, "else") || strings.HasPrefix(strings.TrimSpace(expr), "!"),
	}

	// Classify the condition
	e.classifyCondition(cond)

	return cond
}

// isConditionLine checks if line contains a condition statement
func (e *Extractor) isConditionLine(line string) bool {
	return patterns.IsConditionLine(line)
}

// extractConditionExpression extracts the condition from a line
func (e *Extractor) extractConditionExpression(line string) string {
	return patterns.ExtractConditionExpression(line)
}

// extractVariables extracts variable names from an expression
func (e *Extractor) extractVariables(expr string) []string {
	var vars []string
	seen := make(map[string]bool)

	// Use centralized patterns for the language
	langPatterns := patterns.GetVariablePatterns(e.language)

	for _, p := range langPatterns {
		matches := p.FindAllString(expr, -1)
		for _, m := range matches {
			m = strings.TrimSpace(m)
			// Use centralized keyword check
			if !seen[m] && !common.IsKeyword(m) && !common.IsKeyword(strings.ToLower(m)) && len(m) > 1 {
				vars = append(vars, m)
				seen[m] = true
			}
		}
	}

	return vars
}

// classifyCondition classifies a condition type
func (e *Extractor) classifyCondition(cond *KeyCondition) {
	expr := cond.Expression

	// Check for null/empty checks using centralized pattern
	if patterns.NullCheckPattern.MatchString(expr) {
		cond.Type = CondTypeNullCheck
		cond.Effect = EffectUnknown
		return
	}

	// Check for type checks using centralized pattern
	if patterns.TypeCheckPattern.MatchString(expr) {
		cond.Type = CondTypeTypeCheck
		cond.Effect = EffectUnknown
		return
	}

	// Check for length/count checks using centralized pattern
	if patterns.LengthCheckPattern.MatchString(expr) {
		cond.Type = CondTypeLengthCheck
		cond.Effect = EffectUnknown
		return
	}

	// Check for comparison operators using centralized pattern
	if patterns.ComparisonPattern.MatchString(expr) {
		cond.Type = CondTypeComparison
		cond.Effect = EffectUnknown
		return
	}

	// Check for logical operators using centralized pattern
	if patterns.LogicalOperatorPattern.MatchString(expr) {
		cond.Type = CondTypeLogical
		cond.Effect = EffectUnknown
		return
	}

	// Default
	cond.Type = CondTypeUnknown
	cond.Effect = EffectUnknown
}

// estimateGuardedLines estimates which lines are guarded by a condition
func (e *Extractor) estimateGuardedLines(lines []string, condLineIdx int) []int {
	var guarded []int
	braceCount := 0
	started := false

	for i := condLineIdx; i < len(lines) && i < condLineIdx+100; i++ {
		line := lines[i]
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")

		if openBraces > 0 && !started {
			started = true
		}

		braceCount += openBraces - closeBraces

		if started && braceCount > 0 {
			guarded = append(guarded, i+1) // 1-indexed
		}

		if started && braceCount <= 0 {
			break
		}
	}

	return guarded
}

// GetConditionPathToLine finds all conditions that guard a specific line
func (e *Extractor) GetConditionPathToLine(conditions []*KeyCondition, targetLine int) *ConditionPath {
	path := &ConditionPath{
		Conditions: make([]*KeyCondition, 0),
		TargetLine: targetLine,
		Feasible:   true,
	}

	for _, cond := range conditions {
		for _, guardedLine := range cond.GuardsLines {
			if guardedLine == targetLine {
				path.Conditions = append(path.Conditions, cond)
				break
			}
		}
	}

	// Check feasibility (simple check for contradictions)
	e.checkPathFeasibility(path)

	return path
}

// checkPathFeasibility checks if a condition path is feasible
func (e *Extractor) checkPathFeasibility(path *ConditionPath) {
	// Simple contradiction detection
	varStates := make(map[string]map[string]bool) // var -> (must_be_null, must_be_not_null, etc.)

	for _, cond := range path.Conditions {
		for _, v := range cond.Variables {
			if varStates[v] == nil {
				varStates[v] = make(map[string]bool)
			}

			// Track null states
			if cond.Type == CondTypeNullCheck {
				if cond.IsNegated {
					// !isset or !empty means we need null
					if varStates[v]["not_null"] {
						path.Feasible = false
						path.Reason = "Contradictory null check on " + v
						return
					}
					varStates[v]["null"] = true
				} else {
					// isset or !empty means not null
					if varStates[v]["null"] {
						path.Feasible = false
						path.Reason = "Contradictory null check on " + v
						return
					}
					varStates[v]["not_null"] = true
				}
			}
		}
	}
}

// SummarizeConditions creates a summary of conditions
func (e *Extractor) SummarizeConditions(conditions []*KeyCondition) map[string]interface{} {
	summary := map[string]interface{}{
		"total":     len(conditions),
		"by_type":   make(map[string]int),
		"by_effect": make(map[string]int),
	}

	byType := summary["by_type"].(map[string]int)
	byEffect := summary["by_effect"].(map[string]int)

	for _, cond := range conditions {
		byType[string(cond.Type)]++
		byEffect[string(cond.Effect)]++
	}

	return summary
}
