// Package discovery provides auto-discovery of input carriers by tracing PHP superglobals
package discovery

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/hatlesswizard/inputtracer/pkg/sources"
	phppatterns "github.com/hatlesswizard/inputtracer/pkg/sources/php"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"
)

// PHPSuperglobals references the centralized list from pkg/sources
var PHPSuperglobals = sources.SuperglobalNames()

// SuperglobalUsage tracks where a superglobal is used in the codebase
type SuperglobalUsage struct {
	Superglobal  string `json:"superglobal"`       // "$_GET", "$_POST", etc.
	Key          string `json:"key"`               // The accessed key, e.g., "id" from $_GET['id'], or "*" for all keys
	FilePath     string `json:"file_path"`
	Line         int    `json:"line"`
	Column       int    `json:"column"`
	AssignedTo   string `json:"assigned_to"`       // Variable or property it's assigned to
	Context      string `json:"context"`           // "assignment", "function_arg", "method_arg", "return", "foreach", "direct"
	ClassName    string `json:"class_name"`        // If inside a class method
	MethodName   string `json:"method_name"`       // If inside a method/function
	CalledMethod string `json:"called_method"`     // Method being called when context is "function_arg" (e.g., "parse_incoming")
	CodeSnippet  string `json:"code_snippet"`      // The actual code line
	IsLoopVar    bool   `json:"is_loop_var"`       // True if used as foreach source
	LoopKeyVar   string `json:"loop_key_var"`      // The key variable in foreach
	LoopValVar   string `json:"loop_value_var"`    // The value variable in foreach
}

// SuperglobalFinder finds all superglobal usages in a codebase
type SuperglobalFinder struct {
	parser *sitter.Parser
	// Compiled patterns for superglobal detection
	patterns []*regexp.Regexp
}

// NewSuperglobalFinder creates a new finder instance
func NewSuperglobalFinder() *SuperglobalFinder {
	parser := sitter.NewParser()
	parser.SetLanguage(php.GetLanguage())

	// Compile patterns for each superglobal
	patterns := make([]*regexp.Regexp, 0, len(PHPSuperglobals))
	for _, sg := range PHPSuperglobals {
		// Pattern matches: $_GET['key'], $_GET["key"], $_GET[$var], or just $_GET
		escaped := regexp.QuoteMeta(sg)
		pattern := regexp.MustCompile(escaped + `(?:\s*\[\s*['"]?([^'"\]]+)['"]?\s*\])?`)
		patterns = append(patterns, pattern)
	}

	return &SuperglobalFinder{
		parser:   parser,
		patterns: patterns,
	}
}

// FindAll scans a codebase and returns all superglobal usages (parallelized)
func (f *SuperglobalFinder) FindAll(codebasePath string) ([]SuperglobalUsage, error) {
	// First, collect all PHP file paths
	var filePaths []string
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip non-PHP files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".php") {
			return nil
		}

		// Skip vendor, cache, test directories
		if sources.ShouldSkipPHPPath(path) {
			return nil
		}

		filePaths = append(filePaths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// If few files, process sequentially
	if len(filePaths) <= 4 {
		var usages []SuperglobalUsage
		for _, path := range filePaths {
			fileUsages, err := f.findInFile(path)
			if err != nil {
				continue // Skip files we can't parse
			}
			usages = append(usages, fileUsages...)
		}
		return usages, nil
	}

	// Process files in parallel using worker pool
	numWorkers := runtime.NumCPU()
	if numWorkers > len(filePaths) {
		numWorkers = len(filePaths)
	}

	// Create file channel
	pathChan := make(chan string, len(filePaths))
	for _, p := range filePaths {
		pathChan <- p
	}
	close(pathChan)

	// Results channel
	results := make(chan []SuperglobalUsage, numWorkers)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each worker creates its own parser (tree-sitter parsers are not thread-safe)
			localParser := sitter.NewParser()
			localParser.SetLanguage(php.GetLanguage())

			localUsages := make([]SuperglobalUsage, 0, 32)

			for path := range pathChan {
				fileUsages, err := f.findInFileWithParser(path, localParser)
				if err != nil {
					continue // Skip files we can't parse
				}
				localUsages = append(localUsages, fileUsages...)
			}

			results <- localUsages
		}()
	}

	// Close results when workers done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Merge results
	var usages []SuperglobalUsage
	for localUsages := range results {
		usages = append(usages, localUsages...)
	}

	return usages, nil
}

// findInFileWithParser finds superglobal usages using a provided parser (for parallel processing)
// MEMORY FIX: No longer creates full lines slice - uses getLineFromContent on demand
func (f *SuperglobalFinder) findInFileWithParser(filePath string, parser *sitter.Parser) ([]SuperglobalUsage, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	var usages []SuperglobalUsage

	// Walk the AST to find superglobal usages
	// MEMORY FIX: Pass nil for lines - we'll use getLineFromContent on demand
	f.walkNode(tree.RootNode(), content, nil, filePath, "", "", &usages)

	return usages, nil
}

// findInFile finds superglobal usages in a single file
// MEMORY FIX: No longer creates full lines slice - uses getLineFromContent on demand
func (f *SuperglobalFinder) findInFile(filePath string) ([]SuperglobalUsage, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := f.parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	var usages []SuperglobalUsage

	// Walk the AST to find superglobal usages
	// MEMORY FIX: Pass nil for lines - we'll use getLineFromContent on demand
	f.walkNode(tree.RootNode(), content, nil, filePath, "", "", &usages)

	return usages, nil
}

// walkNode recursively walks the AST to find superglobal usages
func (f *SuperglobalFinder) walkNode(node *sitter.Node, content []byte, lines []string, filePath, currentClass, currentMethod string, usages *[]SuperglobalUsage) {
	if node == nil {
		return
	}

	nodeType := node.Type()

	// Track current class context
	if nodeType == "class_declaration" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			currentClass = nameNode.Content(content)
		}
	}

	// Track current method/function context
	if nodeType == "method_declaration" || nodeType == "function_definition" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			currentMethod = nameNode.Content(content)
		}
	}

	// Check for superglobal usage in subscript_expression (e.g., $_GET['key'])
	if nodeType == "subscript_expression" {
		objNode := node.ChildByFieldName("object")
		if objNode != nil {
			objContent := objNode.Content(content)
			for _, sg := range PHPSuperglobals {
				if objContent == sg {
					usage := f.extractUsage(node, content, lines, filePath, currentClass, currentMethod, sg)
					*usages = append(*usages, usage)
					break
				}
			}
		}
	}

	// Check for superglobal used directly (e.g., just $_GET without subscript)
	if nodeType == "variable_name" {
		varContent := node.Content(content)
		for _, sg := range PHPSuperglobals {
			if varContent == sg {
				// Check if this is the object of a subscript expression (already handled above)
				parent := node.Parent()
				if parent != nil && parent.Type() == "subscript_expression" {
					break
				}
				usage := f.extractUsage(node, content, lines, filePath, currentClass, currentMethod, sg)
				*usages = append(*usages, usage)
				break
			}
		}
	}

	// Check for foreach loops with superglobals
	if nodeType == "foreach_statement" {
		f.checkForeachSuperglobal(node, content, lines, filePath, currentClass, currentMethod, usages)
	}

	// Recurse into children
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		f.walkNode(child, content, lines, filePath, currentClass, currentMethod, usages)
	}
}

// extractUsage extracts details about a superglobal usage
// MEMORY FIX: Uses getLineFromContent instead of pre-allocated lines slice
func (f *SuperglobalFinder) extractUsage(node *sitter.Node, content []byte, lines []string, filePath, currentClass, currentMethod, superglobal string) SuperglobalUsage {
	line := int(node.StartPoint().Row) + 1
	column := int(node.StartPoint().Column)

	// Get the key if it's a subscript expression
	key := "*" // Default to all keys
	if node.Type() == "subscript_expression" {
		indexNode := node.ChildByFieldName("index")
		if indexNode != nil {
			keyContent := indexNode.Content(content)
			// Remove quotes if present
			keyContent = strings.Trim(keyContent, "'\"")
			if keyContent != "" && !strings.HasPrefix(keyContent, "$") {
				key = keyContent
			}
		}
	}

	// Get the code snippet - MEMORY FIX: use on-demand line extraction
	snippet := ""
	if lines != nil && line > 0 && line <= len(lines) {
		snippet = strings.TrimSpace(lines[line-1])
	} else {
		snippet = strings.TrimSpace(getLineFromContent(content, line))
	}

	// Determine context and what it's assigned to
	context, assignedTo, calledMethod := f.determineContext(node, content)

	return SuperglobalUsage{
		Superglobal:  superglobal,
		Key:          key,
		FilePath:     filePath,
		Line:         line,
		Column:       column,
		AssignedTo:   assignedTo,
		Context:      context,
		ClassName:    currentClass,
		MethodName:   currentMethod,
		CalledMethod: calledMethod,
		CodeSnippet:  snippet,
	}
}

// determineContext analyzes the parent nodes to determine how the superglobal is used
func (f *SuperglobalFinder) determineContext(node *sitter.Node, content []byte) (context string, assignedTo string, calledMethod string) {
	context = "direct"
	assignedTo = ""
	calledMethod = ""

	parent := node.Parent()
	for parent != nil {
		parentType := parent.Type()

		switch parentType {
		case "assignment_expression":
			// Check if this is the right side of an assignment
			leftNode := parent.ChildByFieldName("left")
			rightNode := parent.ChildByFieldName("right")
			if rightNode != nil && f.nodeContains(rightNode, node) {
				context = "assignment"
				if leftNode != nil {
					assignedTo = leftNode.Content(content)
				}
				return
			}

		case "argument", "arguments":
			context = "function_arg"
			// Try to find the method/function being called
			calledMethod = f.findCalledMethod(parent, content)
			return

		case "return_statement":
			context = "return"
			return

		case "foreach_statement":
			context = "foreach"
			return

		case "member_access_expression":
			// $this->prop = $_GET['x']
			continue

		case "array_element_initializer":
			// Array context
			context = "array_element"
			return
		}

		parent = parent.Parent()
	}

	return
}

// findCalledMethod finds the name of the method/function being called
func (f *SuperglobalFinder) findCalledMethod(argumentNode *sitter.Node, content []byte) string {
	// Go up to find the function_call_expression or member_call_expression
	parent := argumentNode.Parent()
	for parent != nil {
		switch parent.Type() {
		case "function_call_expression":
			// Regular function call: parse_incoming($_GET)
			funcNode := parent.ChildByFieldName("function")
			if funcNode != nil {
				return funcNode.Content(content)
			}
		case "member_call_expression":
			// Method call: $this->parse_incoming($_GET)
			nameNode := parent.ChildByFieldName("name")
			if nameNode != nil {
				return nameNode.Content(content)
			}
		case "scoped_call_expression":
			// Static method call: self::parse_incoming($_GET)
			nameNode := parent.ChildByFieldName("name")
			if nameNode != nil {
				return nameNode.Content(content)
			}
		}
		parent = parent.Parent()
	}
	return ""
}

// nodeContains checks if ancestor contains descendant
func (f *SuperglobalFinder) nodeContains(ancestor, descendant *sitter.Node) bool {
	if ancestor == nil || descendant == nil {
		return false
	}
	return ancestor.StartByte() <= descendant.StartByte() && ancestor.EndByte() >= descendant.EndByte()
}

// checkForeachSuperglobal checks if a foreach loop iterates over a superglobal
// MEMORY FIX: Uses getLineFromContent instead of pre-allocated lines slice
func (f *SuperglobalFinder) checkForeachSuperglobal(node *sitter.Node, content []byte, lines []string, filePath, currentClass, currentMethod string, usages *[]SuperglobalUsage) {
	// foreach($_GET as $key => $value) or foreach($_GET as $value)
	// The structure varies by tree-sitter version, so we'll use regex on the line

	line := int(node.StartPoint().Row) + 1

	// MEMORY FIX: Get line on-demand instead of using pre-allocated slice
	var lineContent string
	if lines != nil && line > 0 && line <= len(lines) {
		lineContent = lines[line-1]
	} else {
		lineContent = getLineFromContent(content, line)
	}

	if lineContent == "" {
		return
	}

	for i, sg := range PHPSuperglobals {
		if f.patterns[i].MatchString(lineContent) && strings.Contains(lineContent, "foreach") {
			// Extract key and value variables
			keyVar, valVar := f.extractForeachVars(lineContent)

			usage := SuperglobalUsage{
				Superglobal: sg,
				Key:         "*", // Iterating all keys
				FilePath:    filePath,
				Line:        line,
				Column:      0,
				AssignedTo:  valVar,
				Context:     "foreach",
				ClassName:   currentClass,
				MethodName:  currentMethod,
				CodeSnippet: strings.TrimSpace(lineContent),
				IsLoopVar:   true,
				LoopKeyVar:  keyVar,
				LoopValVar:  valVar,
			}
			*usages = append(*usages, usage)
			break
		}
	}
}

// extractForeachVars extracts key and value variable names from a foreach line
func (f *SuperglobalFinder) extractForeachVars(line string) (keyVar, valVar string) {
	// Pattern: foreach($x as $key => $value) or foreach($x as $value)
	// Use centralized patterns from pkg/sources/php
	asPattern := phppatterns.TaintPatterns.LoopVariablePattern
	simplePattern := phppatterns.TaintPatterns.ForeachValueOnlyPattern

	if matches := asPattern.FindStringSubmatch(line); len(matches) >= 3 {
		keyVar = "$" + matches[1]
		valVar = "$" + matches[2]
	} else if matches := simplePattern.FindStringSubmatch(line); len(matches) >= 2 {
		valVar = "$" + matches[1]
	}

	return
}

// FilterByContext filters usages by context type
func FilterByContext(usages []SuperglobalUsage, context string) []SuperglobalUsage {
	var filtered []SuperglobalUsage
	for _, u := range usages {
		if u.Context == context {
			filtered = append(filtered, u)
		}
	}
	return filtered
}

// FilterByClass filters usages by class name
func FilterByClass(usages []SuperglobalUsage, className string) []SuperglobalUsage {
	var filtered []SuperglobalUsage
	for _, u := range usages {
		if u.ClassName == className {
			filtered = append(filtered, u)
		}
	}
	return filtered
}

// GroupBySuperglobal groups usages by superglobal type
func GroupBySuperglobal(usages []SuperglobalUsage) map[string][]SuperglobalUsage {
	grouped := make(map[string][]SuperglobalUsage)
	for _, u := range usages {
		grouped[u.Superglobal] = append(grouped[u.Superglobal], u)
	}
	return grouped
}

// GroupByClass groups usages by class name
func GroupByClass(usages []SuperglobalUsage) map[string][]SuperglobalUsage {
	grouped := make(map[string][]SuperglobalUsage)
	for _, u := range usages {
		key := u.ClassName
		if key == "" {
			key = "(global)"
		}
		grouped[key] = append(grouped[key], u)
	}
	return grouped
}

// getLineFromContent extracts a single line from content without creating a full lines slice
// MEMORY FIX: This avoids allocating a full string slice for the entire file
func getLineFromContent(content []byte, lineNum int) string {
	if lineNum < 1 {
		return ""
	}

	currentLine := 1
	lineStart := 0

	for i, b := range content {
		if b == '\n' {
			if currentLine == lineNum {
				// Found the line - return it without the trailing content
				return string(content[lineStart:i])
			}
			currentLine++
			lineStart = i + 1
		}
	}

	// Handle last line (no trailing newline)
	if currentLine == lineNum && lineStart < len(content) {
		return string(content[lineStart:])
	}

	return ""
}
