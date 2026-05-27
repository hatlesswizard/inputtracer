package common

import (
	"regexp"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/sources/core"
	sitter "github.com/smacker/go-tree-sitter"
)

// InputLabel represents the category of user input.
// This is a type alias — the canonical definition lives in pkg/sources/core.
type InputLabel = core.InputLabel

// Re-export InputLabel constants from core for backward compatibility.
const (
	LabelHTTPGet     = core.LabelHTTPGet
	LabelHTTPPost    = core.LabelHTTPPost
	LabelHTTPCookie  = core.LabelHTTPCookie
	LabelHTTPHeader  = core.LabelHTTPHeader
	LabelHTTPBody    = core.LabelHTTPBody
	LabelCLI         = core.LabelCLI
	LabelEnvironment = core.LabelEnvironment
	LabelFile        = core.LabelFile
	LabelDatabase    = core.LabelDatabase
	LabelNetwork     = core.LabelNetwork
	LabelUserInput   = core.LabelUserInput
)

// Definition represents a user input source definition
type Definition struct {
	Name               string       // e.g., "$_GET", "req.body"
	Pattern            string       // Regex pattern to match
	Language           string       // Target language
	Labels             []InputLabel // Categories
	Description        string       // Human-readable description
	NodeTypes          []string     // Tree-sitter node types to match
	KeyExtractor       string       // Regex to extract key (e.g., from $_GET['key'])
	ExcludeParentTypes []string     // Skip match if node's parent is one of these AST types
	ExcludePattern     string       // Regex pattern - skip match if node text matches this

	// Pre-compiled regexes — populated by NewBaseMatcher; nil means no pattern.
	compiledPattern      *regexp.Regexp
	compiledExclude      *regexp.Regexp
	compiledKeyExtractor *regexp.Regexp
}

// Match represents a matched source in code
type Match struct {
	SourceType string // e.g., "$_GET", "req.body"
	Key        string // e.g., "username" in $_GET['username']
	Variable   string // Variable name if assigned
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

// BaseMatcher provides common functionality for source matching
type BaseMatcher struct {
	lang    string
	sources []Definition
}

// NewBaseMatcher creates a new base matcher and pre-compiles all regex patterns.
// It panics if any Pattern, ExcludePattern, or KeyExtractor field contains an
// invalid regular expression, so invalid patterns are caught at startup rather
// than silently swallowed on every AST node visit.
func NewBaseMatcher(language string, sources []Definition) *BaseMatcher {
	compiled := make([]Definition, len(sources))
	for i, def := range sources {
		d := def // copy

		if d.Pattern != "" {
			re, err := regexp.Compile(d.Pattern)
			if err != nil {
				panic("inputtracer: invalid Pattern in Definition " + d.Name + ": " + err.Error())
			}
			d.compiledPattern = re
		}

		if d.ExcludePattern != "" {
			re, err := regexp.Compile(d.ExcludePattern)
			if err != nil {
				panic("inputtracer: invalid ExcludePattern in Definition " + d.Name + ": " + err.Error())
			}
			d.compiledExclude = re
		}

		if d.KeyExtractor != "" {
			re, err := regexp.Compile(d.KeyExtractor)
			if err != nil {
				panic("inputtracer: invalid KeyExtractor in Definition " + d.Name + ": " + err.Error())
			}
			d.compiledKeyExtractor = re
		}

		compiled[i] = d
	}

	return &BaseMatcher{
		lang:    language,
		sources: compiled,
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

		for i := range m.sources {
			source := &m.sources[i]

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

			// Check pattern match (using pre-compiled regex)
			if source.compiledPattern != nil {
				if !source.compiledPattern.MatchString(nodeText) {
					continue
				}
			}

			// Check exclude pattern (using pre-compiled regex)
			if source.compiledExclude != nil {
				if source.compiledExclude.MatchString(nodeText) {
					continue
				}
			}

			// Check parent type exclusion
			if len(source.ExcludeParentTypes) > 0 && node.Parent() != nil {
				parentType := node.Parent().Type()
				excluded := false
				for _, ept := range source.ExcludeParentTypes {
					if parentType == ept {
						excluded = true
						break
					}
				}
				if excluded {
					continue
				}
			}

			// Extract key if pattern provided (using pre-compiled regex)
			key := ""
			if source.compiledKeyExtractor != nil {
				if submatches := source.compiledKeyExtractor.FindStringSubmatch(nodeText); len(submatches) > 1 {
					key = submatches[1]
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

// Package-level compiled regexes for helper functions — compiled once at init time.
var (
	reJSIdentifier      = regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`)
	reGenericIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	reFirstIdentifier   = regexp.MustCompile(`^([a-zA-Z_$][a-zA-Z0-9_$]*)`)
	reWhitespace        = regexp.MustCompile(`\s+`)
)

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
		return reJSIdentifier.MatchString(s)
	case "python", "go":
		return reGenericIdentifier.MatchString(s)
	default:
		return reGenericIdentifier.MatchString(s)
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
		if match := reFirstIdentifier.FindStringSubmatch(s); len(match) > 1 {
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
	s = reWhitespace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
