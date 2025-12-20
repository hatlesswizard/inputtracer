package sources

import (
	"regexp"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
)

// InputLabel represents the category of user input
type InputLabel string

const (
	LabelHTTPGet     InputLabel = "http_get"
	LabelHTTPPost    InputLabel = "http_post"
	LabelHTTPCookie  InputLabel = "http_cookie"
	LabelHTTPHeader  InputLabel = "http_header"
	LabelHTTPBody    InputLabel = "http_body"
	LabelCLI         InputLabel = "cli"
	LabelEnvironment InputLabel = "environment"
	LabelFile        InputLabel = "file"
	LabelDatabase    InputLabel = "database"
	LabelNetwork     InputLabel = "network"
	LabelUserInput   InputLabel = "user_input"
)

// Definition represents a user input source definition
type Definition struct {
	Name        string       // e.g., "$_GET", "req.body"
	Pattern     string       // Regex pattern to match
	Language    string       // Target language
	Labels      []InputLabel // Categories
	Description string       // Human-readable description
	NodeTypes   []string     // Tree-sitter node types to match
	KeyExtractor string      // Regex to extract key (e.g., from $_GET['key'])
}

// Match represents a matched source in code
type Match struct {
	SourceType string       // e.g., "$_GET", "req.body"
	Key        string       // e.g., "username" in $_GET['username']
	Variable   string       // Variable name if assigned
	Line       int
	Column     int
	EndLine    int
	EndColumn  int
	Snippet    string
	Labels     []InputLabel
}

// Matcher interface for language-specific source detection
type Matcher interface {
	Language() string
	FindSources(root *sitter.Node, src []byte) []Match
}

// Registry manages all source matchers
type Registry struct {
	matchers map[string]Matcher
	sources  map[string][]Definition
	mu       sync.RWMutex
}

// NewRegistry creates a new source registry
func NewRegistry() *Registry {
	return &Registry{
		matchers: make(map[string]Matcher),
		sources:  make(map[string][]Definition),
	}
}

// RegisterMatcher registers a language-specific matcher
func (r *Registry) RegisterMatcher(matcher Matcher) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matchers[matcher.Language()] = matcher
}

// AddSource adds a source definition
func (r *Registry) AddSource(def Definition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sources[def.Language] = append(r.sources[def.Language], def)
}

// GetMatcher returns the matcher for a language
func (r *Registry) GetMatcher(language string) Matcher {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.matchers[language]
}

// GetSources returns all source definitions for a language
func (r *Registry) GetSources(language string) []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sources[language]
}

// BaseMatcher provides common functionality for source matching
type BaseMatcher struct {
	lang    string
	sources []Definition
}

// NewBaseMatcher creates a new base matcher
func NewBaseMatcher(language string, sources []Definition) *BaseMatcher {
	return &BaseMatcher{
		lang:    language,
		sources: sources,
	}
}

// Language returns the language this matcher handles
func (m *BaseMatcher) Language() string {
	return m.lang
}

// FindSources finds all input sources in the AST
func (m *BaseMatcher) FindSources(root *sitter.Node, src []byte) []Match {
	var matches []Match

	m.traverse(root, src, func(node *sitter.Node) {
		nodeType := node.Type()
		nodeText := string(src[node.StartByte():node.EndByte()])

		for _, source := range m.sources {
			// Check node type match
			nodeTypeMatch := len(source.NodeTypes) == 0
			for _, nt := range source.NodeTypes {
				if nodeType == nt {
					nodeTypeMatch = true
					break
				}
			}
			if !nodeTypeMatch {
				continue
			}

			// Check pattern match
			if source.Pattern != "" {
				re, err := regexp.Compile(source.Pattern)
				if err != nil {
					continue
				}
				if !re.MatchString(nodeText) {
					continue
				}
			}

			// Extract key if pattern provided
			key := ""
			if source.KeyExtractor != "" {
				re, err := regexp.Compile(source.KeyExtractor)
				if err == nil {
					if submatches := re.FindStringSubmatch(nodeText); len(submatches) > 1 {
						key = submatches[1]
					}
				}
			}

			// Check if this is part of an assignment
			variable := m.findAssignmentTarget(node, src)

			matches = append(matches, Match{
				SourceType: source.Name,
				Key:        key,
				Variable:   variable,
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				EndLine:    int(node.EndPoint().Row) + 1,
				EndColumn:  int(node.EndPoint().Column),
				Snippet:    truncateSnippet(nodeText, 100),
				Labels:     source.Labels,
			})
		}
	})

	return matches
}

// traverse recursively traverses the AST
func (m *BaseMatcher) traverse(node *sitter.Node, src []byte, callback func(*sitter.Node)) {
	if node == nil {
		return
	}
	callback(node)
	for i := 0; i < int(node.ChildCount()); i++ {
		m.traverse(node.Child(i), src, callback)
	}
}

// findAssignmentTarget finds the variable being assigned if this is part of an assignment
func (m *BaseMatcher) findAssignmentTarget(node *sitter.Node, src []byte) string {
	// Walk up to find assignment expression
	parent := node.Parent()
	for parent != nil {
		parentType := parent.Type()
		if strings.Contains(parentType, "assignment") {
			// Look for left-hand side
			for i := 0; i < int(parent.ChildCount()); i++ {
				child := parent.Child(i)
				if child != nil && child != node {
					childText := string(src[child.StartByte():child.EndByte()])
					// Check if this looks like a variable
					if isLikelyVariable(childText, m.lang) {
						return extractVariableName(childText, m.lang)
					}
				}
			}
		}
		parent = parent.Parent()
	}
	return ""
}

// isLikelyVariable checks if a string looks like a variable name
func isLikelyVariable(s string, lang string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	switch lang {
	case "php":
		return strings.HasPrefix(s, "$")
	case "javascript", "typescript":
		return regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`).MatchString(s)
	case "python":
		return regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(s)
	case "go":
		return regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(s)
	default:
		return regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(s)
	}
}

// extractVariableName extracts the variable name from an expression
func extractVariableName(s string, lang string) string {
	s = strings.TrimSpace(s)

	switch lang {
	case "php":
		// Remove $ prefix for consistency, or keep it
		return s
	default:
		// Extract first identifier
		re := regexp.MustCompile(`^([a-zA-Z_$][a-zA-Z0-9_$]*)`)
		if match := re.FindStringSubmatch(s); len(match) > 1 {
			return match[1]
		}
		return s
	}
}

// truncateSnippet truncates a snippet to a maximum length
func truncateSnippet(s string, maxLen int) string {
	// Remove newlines and normalize whitespace
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// RegisterAll registers all language matchers with the registry
func RegisterAll(r *Registry) {
	// Register PHP
	r.RegisterMatcher(NewPHPMatcher())

	// Register JavaScript
	r.RegisterMatcher(NewJavaScriptMatcher())

	// Register TypeScript (uses JS matcher with different language)
	r.RegisterMatcher(NewTypeScriptMatcher())

	// Register Python
	r.RegisterMatcher(NewPythonMatcher())

	// Register Go
	r.RegisterMatcher(NewGoMatcher())

	// Register Java
	r.RegisterMatcher(NewJavaMatcher())

	// Register C
	r.RegisterMatcher(NewCMatcher())

	// Register C++
	r.RegisterMatcher(NewCPPMatcher())

	// Register C#
	r.RegisterMatcher(NewCSharpMatcher())

	// Register Ruby
	r.RegisterMatcher(NewRubyMatcher())

	// Register Rust
	r.RegisterMatcher(NewRustMatcher())
}
