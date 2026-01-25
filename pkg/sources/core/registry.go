package core

import (
	"regexp"
	"strings"
	"sync"
)

// InputPattern defines a pattern for detecting input sources
type InputPattern struct {
	Name        string       // Unique identifier
	Description string       // Human-readable description
	Category    SourceType   // Primary category
	Labels      []InputLabel // Additional labels
	Language    string       // Target language (empty = all)
	Framework   string       // Target framework (empty = all)

	// Pattern matching (use one or more)
	ExactMatch    string         // Exact string match
	Regex         *regexp.Regexp // Compiled regex
	MethodName    string         // Method name to match
	PropertyName  string         // Property name to match
	ObjectPattern string         // Object name pattern

	// Context requirements
	RequireObject bool   // Must be called on an object
	ObjectType    string // Required object type (if known)
	ParamIndex    int    // Which parameter receives input (-1 = return value)
}

// Registry holds all registered input patterns
type Registry struct {
	mu sync.RWMutex

	// Fast exact match lookup
	exactPatterns map[string]*InputPattern

	// Regex patterns (checked in order)
	regexPatterns []*InputPattern

	// Language-specific patterns
	languagePatterns map[string][]*InputPattern

	// Framework-specific patterns
	frameworkPatterns map[string][]*InputPattern

	// Non-input patterns (explicitly excluded)
	nonInputPatterns map[string]bool
}

var (
	globalRegistry *Registry
	registryOnce   sync.Once
)

// GetRegistry returns the global registry singleton
func GetRegistry() *Registry {
	registryOnce.Do(func() {
		globalRegistry = &Registry{
			exactPatterns:     make(map[string]*InputPattern),
			regexPatterns:     make([]*InputPattern, 0, 100),
			languagePatterns:  make(map[string][]*InputPattern),
			frameworkPatterns: make(map[string][]*InputPattern),
			nonInputPatterns:  make(map[string]bool),
		}
	})
	return globalRegistry
}

// Register adds a pattern to the registry
func (r *Registry) Register(pattern *InputPattern) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Index by exact match
	if pattern.ExactMatch != "" {
		r.exactPatterns[pattern.ExactMatch] = pattern
	}

	// Add to regex list
	if pattern.Regex != nil {
		r.regexPatterns = append(r.regexPatterns, pattern)
	}

	// Index by language
	if pattern.Language != "" {
		r.languagePatterns[pattern.Language] = append(
			r.languagePatterns[pattern.Language], pattern)
	}

	// Index by framework
	if pattern.Framework != "" {
		r.frameworkPatterns[pattern.Framework] = append(
			r.frameworkPatterns[pattern.Framework], pattern)
	}
}

// RegisterNonInput marks a pattern as explicitly NOT user input
func (r *Registry) RegisterNonInput(pattern string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nonInputPatterns[strings.ToLower(pattern)] = true
}

// IsNonInput checks if a pattern is explicitly marked as non-input
func (r *Registry) IsNonInput(expr string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nonInputPatterns[strings.ToLower(expr)]
}

// MatchResult contains the result of a pattern match
type MatchResult struct {
	Pattern  *InputPattern
	Category SourceType
	Labels   []InputLabel
	Key      string // Extracted key if applicable
}

// Match attempts to match an expression against registered patterns
func (r *Registry) Match(expr string, language string, framework string) *MatchResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check non-input first
	if r.nonInputPatterns[strings.ToLower(expr)] {
		return nil
	}

	// Try exact match (fastest)
	if pattern, ok := r.exactPatterns[expr]; ok {
		if r.patternApplies(pattern, language, framework) {
			return &MatchResult{
				Pattern:  pattern,
				Category: pattern.Category,
				Labels:   pattern.Labels,
			}
		}
	}

	// Try framework-specific patterns
	if framework != "" {
		if patterns, ok := r.frameworkPatterns[framework]; ok {
			if result := r.matchPatterns(expr, patterns, language, framework); result != nil {
				return result
			}
		}
	}

	// Try language-specific patterns
	if language != "" {
		if patterns, ok := r.languagePatterns[language]; ok {
			if result := r.matchPatterns(expr, patterns, language, framework); result != nil {
				return result
			}
		}
	}

	// Try regex patterns
	return r.matchPatterns(expr, r.regexPatterns, language, framework)
}

func (r *Registry) matchPatterns(expr string, patterns []*InputPattern, language string, framework string) *MatchResult {
	for _, pattern := range patterns {
		if !r.patternApplies(pattern, language, framework) {
			continue
		}

		if pattern.Regex != nil && pattern.Regex.MatchString(expr) {
			result := &MatchResult{
				Pattern:  pattern,
				Category: pattern.Category,
				Labels:   pattern.Labels,
			}
			// Try to extract key
			if matches := pattern.Regex.FindStringSubmatch(expr); len(matches) > 1 {
				result.Key = matches[1]
			}
			return result
		}
	}
	return nil
}

func (r *Registry) patternApplies(pattern *InputPattern, language string, framework string) bool {
	if pattern.Language != "" && pattern.Language != language {
		return false
	}
	if pattern.Framework != "" && pattern.Framework != framework {
		return false
	}
	return true
}

// MatchMethod checks if a method call is an input source
func (r *Registry) MatchMethod(objName string, methodName string, language string, framework string) *MatchResult {
	// Build expression to match
	expr := methodName
	if objName != "" {
		expr = objName + "." + methodName
	}

	// Check registry first
	if result := r.Match(expr, language, framework); result != nil {
		return result
	}

	// Fall back to universal patterns
	if IsInputMethod(methodName) && !IsExcludedMethod(methodName) {
		if objName == "" || IsInputObject(objName) {
			return &MatchResult{
				Category: SourceUserInput,
				Labels:   []InputLabel{LabelUserInput},
			}
		}
	}

	return nil
}

// MatchProperty checks if a property access is an input source
func (r *Registry) MatchProperty(objName string, propName string, language string, framework string) *MatchResult {
	expr := objName + "." + propName

	if result := r.Match(expr, language, framework); result != nil {
		return result
	}

	// Fall back to universal patterns
	if IsInputProperty(propName) && IsInputObject(objName) {
		return &MatchResult{
			Category: SourceUserInput,
			Labels:   []InputLabel{LabelUserInput},
		}
	}

	return nil
}
