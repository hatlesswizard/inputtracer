// Package javascript implements the JavaScript language analyzer for semantic input tracing
package javascript

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	"github.com/hatlesswizard/inputtracer/pkg/sources"
	jsPatterns "github.com/hatlesswizard/inputtracer/pkg/sources/javascript"
	sitter "github.com/smacker/go-tree-sitter"
)

// JSAnalyzer implements the LanguageAnalyzer interface for JavaScript
type JSAnalyzer struct {
	*analyzer.BaseAnalyzer
	globalSources   map[string]types.SourceType
	domSources      map[string]types.SourceType
	nodeSources     map[string]types.SourceType
}

// NewJSAnalyzer creates a new JavaScript analyzer
func NewJSAnalyzer() *JSAnalyzer {
	m := sources.GetMappings("javascript")
	a := &JSAnalyzer{
		BaseAnalyzer:  analyzer.NewBaseAnalyzer("javascript", languages.GetExtensionsForLanguage("javascript")),
		globalSources: m.GetGlobalSourcesMap(),
		domSources:    m.GetDOMSourcesMap(),
		nodeSources:   m.GetNodeSourcesMap(),
	}

	// Register framework patterns
	a.registerFrameworkPatterns()

	return a
}

// registerFrameworkPatterns loads JavaScript framework patterns from pkg/sources/javascript
// This centralizes all framework patterns in one place
func (a *JSAnalyzer) registerFrameworkPatterns() {
	// Load all patterns from pkg/sources/javascript registry
	for _, p := range jsPatterns.GetAllPatterns() {
		// Convert common.FrameworkPattern to types.FrameworkPattern
		fp := &types.FrameworkPattern{
			ID:              p.ID,
			Framework:       p.Framework,
			Language:        p.Language,
			Name:            p.Name,
			Description:     p.Description,
			ClassPattern:    p.ClassPattern,
			MethodPattern:   p.MethodPattern,
			PropertyPattern: p.PropertyPattern,
			AccessPattern:   p.AccessPattern,
			SourceType:      types.SourceType(p.SourceType),
			SourceKey:       p.SourceKey,
			CarrierClass:    p.CarrierClass,
			CarrierProperty: p.CarrierProperty,
			PopulatedBy:     p.PopulatedBy,
			PopulatedFrom:   p.PopulatedFrom,
			Confidence:      p.Confidence,
		}
		a.AddFrameworkPattern(fp)
	}
}

// BuildSymbolTable builds the symbol table for a JavaScript file
func (a *JSAnalyzer) BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error) {
	st := types.NewSymbolTable(filePath, "javascript")

	// Extract imports
	st.Imports = a.extractImports(root, source)

	// Extract classes
	classes, err := a.ExtractClasses(root, source)
	if err != nil {
		return nil, err
	}
	for _, class := range classes {
		st.Classes[class.Name] = class
	}

	// Extract functions
	functions, err := a.ExtractFunctions(root, source)
	if err != nil {
		return nil, err
	}
	for _, fn := range functions {
		st.Functions[fn.Name] = fn
	}

	// Detect frameworks
	frameworks, _ := a.DetectFrameworks(st, source)
	if len(frameworks) > 0 {
		st.Framework = frameworks[0]
	}

	return st, nil
}

// extractImports extracts import/require statements
func (a *JSAnalyzer) extractImports(root *sitter.Node, source []byte) []types.ImportInfo {
	var imports []types.ImportInfo

	// ES6 imports
	importNodes := analyzer.FindNodesOfType(root, "import_statement")
	for _, node := range importNodes {
		imp := types.ImportInfo{
			Line: int(node.StartPoint().Row) + 1,
			Type: "import",
		}

		// Get source path
		sourceNode := analyzer.FindChildByType(node, "string")
		if sourceNode != nil {
			imp.Path = strings.Trim(analyzer.GetNodeText(sourceNode, source), "\"'`")
		}

		// Get imported names
		importClause := analyzer.FindChildByType(node, "import_clause")
		if importClause != nil {
			// Default import
			identNode := analyzer.FindChildByType(importClause, "identifier")
			if identNode != nil {
				imp.Names = append(imp.Names, analyzer.GetNodeText(identNode, source))
			}

			// Named imports
			namedImports := analyzer.FindChildByType(importClause, "named_imports")
			if namedImports != nil {
				specifiers := analyzer.FindNodesOfType(namedImports, "import_specifier")
				for _, spec := range specifiers {
					nameNode := analyzer.FindChildByType(spec, "identifier")
					if nameNode != nil {
						imp.Names = append(imp.Names, analyzer.GetNodeText(nameNode, source))
					}
				}
			}
		}

		imp.IsRelative = strings.HasPrefix(imp.Path, ".") || strings.HasPrefix(imp.Path, "/")
		imports = append(imports, imp)
	}

	// CommonJS require
	callNodes := analyzer.FindNodesOfType(root, "call_expression")
	for _, node := range callNodes {
		funcNode := node.Child(0)
		if funcNode != nil && analyzer.GetNodeText(funcNode, source) == "require" {
			argsNode := analyzer.FindChildByType(node, "arguments")
			if argsNode != nil {
				stringNode := analyzer.FindChildByType(argsNode, "string")
				if stringNode != nil {
					path := strings.Trim(analyzer.GetNodeText(stringNode, source), "\"'`")
					imports = append(imports, types.ImportInfo{
						Path:       path,
						Line:       int(node.StartPoint().Row) + 1,
						Type:       "require",
						IsRelative: strings.HasPrefix(path, ".") || strings.HasPrefix(path, "/"),
					})
				}
			}
		}
	}

	return imports
}

// ResolveImports resolves import paths to actual file paths
func (a *JSAnalyzer) ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error) {
	var resolvedPaths []string
	dir := filepath.Dir(symbolTable.FilePath)

	for _, imp := range symbolTable.Imports {
		if !imp.IsRelative {
			// Node module - skip for now
			continue
		}

		// Try different extensions
		extensions := []string{"", ".js", ".mjs", ".jsx", "/index.js"}
		for _, ext := range extensions {
			resolvedPath := filepath.Join(dir, imp.Path+ext)
			resolvedPath = filepath.Clean(resolvedPath)
			resolvedPaths = append(resolvedPaths, resolvedPath)
		}
	}

	return resolvedPaths, nil
}

// ExtractClasses extracts class definitions from JavaScript AST
func (a *JSAnalyzer) ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error) {
	var classes []*types.ClassDef

	classNodes := analyzer.FindNodesOfType(root, "class_declaration")
	for _, classNode := range classNodes {
		class := a.parseClassDeclaration(classNode, source)
		if class != nil {
			classes = append(classes, class)
		}
	}

	// Also handle class expressions assigned to variables
	classExprs := analyzer.FindNodesOfType(root, "class")
	for _, classNode := range classExprs {
		class := a.parseClassDeclaration(classNode, source)
		if class != nil {
			classes = append(classes, class)
		}
	}

	return classes, nil
}

// parseClassDeclaration parses a class declaration
func (a *JSAnalyzer) parseClassDeclaration(node *sitter.Node, source []byte) *types.ClassDef {
	nameNode := analyzer.FindChildByType(node, "identifier")
	if nameNode == nil {
		return nil
	}

	class := types.NewClassDef(
		analyzer.GetNodeText(nameNode, source),
		"",
		int(node.StartPoint().Row)+1,
	)
	class.EndLine = int(node.EndPoint().Row) + 1

	// Get extends clause
	heritage := analyzer.FindChildByType(node, "class_heritage")
	if heritage != nil {
		extNode := analyzer.FindChildByType(heritage, "identifier")
		if extNode != nil {
			class.Extends = analyzer.GetNodeText(extNode, source)
		}
	}

	// Get body
	bodyNode := analyzer.FindChildByType(node, "class_body")
	if bodyNode != nil {
		a.parseClassBody(class, bodyNode, source)
	}

	return class
}

// parseClassBody parses the body of a class
func (a *JSAnalyzer) parseClassBody(class *types.ClassDef, bodyNode *sitter.Node, source []byte) {
	for i := 0; i < int(bodyNode.ChildCount()); i++ {
		child := bodyNode.Child(i)
		switch child.Type() {
		case "method_definition":
			method := a.parseMethodDefinition(child, source)
			if method != nil {
				if method.Name == "constructor" {
					class.Constructor = method
				}
				class.Methods[method.Name] = method
			}
		case "field_definition", "public_field_definition":
			prop := a.parseFieldDefinition(child, source)
			if prop != nil {
				class.Properties[prop.Name] = prop
			}
		}
	}
}

// parseMethodDefinition parses a method definition
func (a *JSAnalyzer) parseMethodDefinition(node *sitter.Node, source []byte) *types.MethodDef {
	method := &types.MethodDef{
		Line:       int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Visibility: "public",
		Parameters: make([]types.ParameterDef, 0),
	}

	// Get name
	nameNode := analyzer.FindChildByType(node, "property_identifier")
	if nameNode == nil {
		nameNode = analyzer.FindChildByType(node, "identifier")
	}
	if nameNode != nil {
		method.Name = analyzer.GetNodeText(nameNode, source)
	}

	// Check for static
	for j := 0; j < int(node.ChildCount()); j++ {
		if node.Child(j).Type() == "static" {
			method.IsStatic = true
		}
	}

	// Get parameters
	paramsNode := analyzer.FindChildByType(node, "formal_parameters")
	if paramsNode != nil {
		method.Parameters = a.parseParameters(paramsNode, source)
	}

	// Check for async
	for j := 0; j < int(node.ChildCount()); j++ {
		if node.Child(j).Type() == "async" {
			method.IsAsync = true
		}
	}

	// Get body
	bodyNode := analyzer.FindChildByType(node, "statement_block")
	if bodyNode != nil {
		method.BodyStart = int(bodyNode.StartPoint().Row) + 1
		method.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		method.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	return method
}

// parseFieldDefinition parses a class field definition
func (a *JSAnalyzer) parseFieldDefinition(node *sitter.Node, source []byte) *types.PropertyDef {
	prop := &types.PropertyDef{
		Line:       int(node.StartPoint().Row) + 1,
		Visibility: "public",
	}

	// Get name
	nameNode := analyzer.FindChildByType(node, "property_identifier")
	if nameNode == nil {
		nameNode = analyzer.FindChildByType(node, "identifier")
	}
	if nameNode != nil {
		prop.Name = analyzer.GetNodeText(nameNode, source)
	}

	// Get initial value
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() != "property_identifier" && child.Type() != "identifier" && child.Type() != "=" {
			prop.InitialValue = analyzer.GetNodeText(child, source)
			break
		}
	}

	return prop
}

// parseParameters parses function parameters
func (a *JSAnalyzer) parseParameters(node *sitter.Node, source []byte) []types.ParameterDef {
	var params []types.ParameterDef

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			params = append(params, types.ParameterDef{
				Index: len(params),
				Name:  analyzer.GetNodeText(child, source),
			})
		case "assignment_pattern":
			// Default parameter
			nameNode := child.Child(0)
			if nameNode != nil {
				param := types.ParameterDef{
					Index: len(params),
					Name:  analyzer.GetNodeText(nameNode, source),
				}
				if child.ChildCount() > 2 {
					param.DefaultValue = analyzer.GetNodeText(child.Child(2), source)
				}
				params = append(params, param)
			}
		case "rest_pattern":
			// Rest parameter
			nameNode := analyzer.FindChildByType(child, "identifier")
			if nameNode != nil {
				params = append(params, types.ParameterDef{
					Index:      len(params),
					Name:       analyzer.GetNodeText(nameNode, source),
					IsVariadic: true,
				})
			}
		case "object_pattern", "array_pattern":
			// Destructuring parameter
			params = append(params, types.ParameterDef{
				Index: len(params),
				Name:  analyzer.GetNodeText(child, source),
			})
		}
	}

	return params
}

// ExtractFunctions extracts standalone function definitions
func (a *JSAnalyzer) ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error) {
	var functions []*types.FunctionDef

	// Function declarations
	funcNodes := analyzer.FindNodesOfType(root, "function_declaration")
	for _, funcNode := range funcNodes {
		fn := a.parseFunctionDeclaration(funcNode, source)
		if fn != nil {
			functions = append(functions, fn)
		}
	}

	// Arrow functions assigned to variables
	varDecls := analyzer.FindNodesOfType(root, "variable_declarator")
	for _, varDecl := range varDecls {
		nameNode := analyzer.FindChildByType(varDecl, "identifier")
		if nameNode == nil {
			continue
		}

		// Check for arrow function or function expression
		for i := 0; i < int(varDecl.ChildCount()); i++ {
			child := varDecl.Child(i)
			if child.Type() == "arrow_function" || child.Type() == "function" {
				fn := &types.FunctionDef{
					Name:       analyzer.GetNodeText(nameNode, source),
					Line:       int(varDecl.StartPoint().Row) + 1,
					EndLine:    int(child.EndPoint().Row) + 1,
					Parameters: make([]types.ParameterDef, 0),
				}

				paramsNode := analyzer.FindChildByType(child, "formal_parameters")
				if paramsNode != nil {
					fn.Parameters = a.parseParameters(paramsNode, source)
				} else {
					// Single param arrow function
					paramNode := analyzer.FindChildByType(child, "identifier")
					if paramNode != nil {
						fn.Parameters = append(fn.Parameters, types.ParameterDef{
							Index: 0,
							Name:  analyzer.GetNodeText(paramNode, source),
						})
					}
				}

				// Check for async
				for j := 0; j < int(child.ChildCount()); j++ {
					if child.Child(j).Type() == "async" {
						fn.IsAsync = true
					}
				}

				functions = append(functions, fn)
				break
			}
		}
	}

	return functions, nil
}

// parseFunctionDeclaration parses a function declaration
func (a *JSAnalyzer) parseFunctionDeclaration(node *sitter.Node, source []byte) *types.FunctionDef {
	nameNode := analyzer.FindChildByType(node, "identifier")
	if nameNode == nil {
		return nil
	}

	fn := &types.FunctionDef{
		Name:       analyzer.GetNodeText(nameNode, source),
		Line:       int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Parameters: make([]types.ParameterDef, 0),
	}

	// Get parameters
	paramsNode := analyzer.FindChildByType(node, "formal_parameters")
	if paramsNode != nil {
		fn.Parameters = a.parseParameters(paramsNode, source)
	}

	// Check for async
	for i := 0; i < int(node.ChildCount()); i++ {
		if node.Child(i).Type() == "async" {
			fn.IsAsync = true
		}
	}

	// Get body
	bodyNode := analyzer.FindChildByType(node, "statement_block")
	if bodyNode != nil {
		fn.BodyStart = int(bodyNode.StartPoint().Row) + 1
		fn.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		fn.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	return fn
}

// ExtractAssignments extracts all assignments from the AST
func (a *JSAnalyzer) ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error) {
	var assignments []*types.Assignment

	// Assignment expressions
	assignNodes := analyzer.FindNodesOfType(root, "assignment_expression")
	for _, node := range assignNodes {
		assignment := a.parseAssignment(node, source, scope)
		if assignment != nil {
			assignments = append(assignments, assignment)
		}
	}

	// Variable declarations with initializers
	varDecls := analyzer.FindNodesOfType(root, "variable_declarator")
	for _, node := range varDecls {
		nameNode := node.Child(0)
		if nameNode == nil || node.ChildCount() < 2 {
			continue
		}

		// Find the value (skip the name and =)
		var valueNode *sitter.Node
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() != "identifier" && child.Type() != "=" &&
				child.Type() != "object_pattern" && child.Type() != "array_pattern" {
				valueNode = child
				break
			}
		}

		if valueNode != nil {
			assignment := &types.Assignment{
				Target:     analyzer.GetNodeText(nameNode, source),
				Source:     analyzer.GetNodeText(valueNode, source),
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Scope:      scope,
				TargetType: "variable",
				Operator:   "=",
			}

			// Check if source is tainted
			assignment.IsTainted, assignment.TaintSource = a.isExpressionTainted(valueNode, source)

			assignments = append(assignments, assignment)
		}
	}

	return assignments, nil
}

// parseAssignment parses an assignment expression
func (a *JSAnalyzer) parseAssignment(node *sitter.Node, source []byte, scope string) *types.Assignment {
	if node.ChildCount() < 3 {
		return nil
	}

	leftNode := node.Child(0)
	opNode := node.Child(1)
	rightNode := node.Child(2)

	if leftNode == nil || rightNode == nil {
		return nil
	}

	assignment := &types.Assignment{
		Target:   analyzer.GetNodeText(leftNode, source),
		Source:   analyzer.GetNodeText(rightNode, source),
		Line:     int(node.StartPoint().Row) + 1,
		Column:   int(node.StartPoint().Column),
		Scope:    scope,
	}

	if opNode != nil {
		assignment.Operator = analyzer.GetNodeText(opNode, source)
	}

	// Determine target type
	switch leftNode.Type() {
	case "identifier":
		assignment.TargetType = "variable"
	case "member_expression":
		assignment.TargetType = "property"
	case "subscript_expression":
		assignment.TargetType = "array_element"
	}

	// Check if source is tainted
	assignment.IsTainted, assignment.TaintSource = a.isExpressionTainted(rightNode, source)

	return assignment
}

// isExpressionTainted checks if an expression contains tainted data
func (a *JSAnalyzer) isExpressionTainted(node *sitter.Node, source []byte) (bool, string) {
	if node == nil {
		return false, ""
	}

	text := analyzer.GetNodeText(node, source)

	// Check for global sources
	for src, _ := range a.globalSources {
		if strings.Contains(text, src) {
			return true, src
		}
	}

	// Check for Node.js sources
	for src, _ := range a.nodeSources {
		if strings.Contains(text, src) {
			return true, src
		}
	}

	// Check for Express patterns
	expressPatterns := []string{"req.body", "req.query", "req.params", "req.headers", "req.cookies"}
	for _, pattern := range expressPatterns {
		if strings.Contains(text, pattern) {
			return true, pattern
		}
	}

	// Check for DOM element.value
	if strings.Contains(text, ".value") {
		return true, "element.value"
	}

	// Recursively check children
	for i := 0; i < int(node.ChildCount()); i++ {
		if tainted, src := a.isExpressionTainted(node.Child(i), source); tainted {
			return true, src
		}
	}

	return false, ""
}

// ExtractCalls extracts all function/method calls from the AST
func (a *JSAnalyzer) ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error) {
	var calls []*types.CallSite

	callNodes := analyzer.FindNodesOfType(root, "call_expression")
	for _, node := range callNodes {
		call := a.parseCallExpression(node, source, scope)
		if call != nil {
			calls = append(calls, call)
		}
	}

	// New expressions
	newNodes := analyzer.FindNodesOfType(root, "new_expression")
	for _, node := range newNodes {
		call := a.parseNewExpression(node, source, scope)
		if call != nil {
			calls = append(calls, call)
		}
	}

	return calls, nil
}

// parseCallExpression parses a call expression
func (a *JSAnalyzer) parseCallExpression(node *sitter.Node, source []byte, scope string) *types.CallSite {
	call := &types.CallSite{
		Line:      int(node.StartPoint().Row) + 1,
		Column:    int(node.StartPoint().Column),
		Scope:     scope,
		Arguments: make([]types.CallArg, 0),
	}

	// Get function/method being called
	funcNode := node.Child(0)
	if funcNode == nil {
		return nil
	}

	switch funcNode.Type() {
	case "identifier":
		call.FunctionName = analyzer.GetNodeText(funcNode, source)
	case "member_expression":
		// obj.method() call
		call.FunctionName = analyzer.GetNodeText(funcNode, source)
		objNode := funcNode.Child(0)
		if objNode != nil {
			call.ClassName = analyzer.GetNodeText(objNode, source)
		}
		propNode := analyzer.FindChildByType(funcNode, "property_identifier")
		if propNode != nil {
			call.MethodName = analyzer.GetNodeText(propNode, source)
		}
	}

	// Get arguments
	argsNode := analyzer.FindChildByType(node, "arguments")
	if argsNode != nil {
		call.Arguments = a.parseCallArguments(argsNode, source)
	}

	// Check for tainted args
	for i, arg := range call.Arguments {
		if arg.IsTainted {
			call.HasTaintedArgs = true
			call.TaintedArgIndices = append(call.TaintedArgIndices, i)
		}
	}

	return call
}

// parseNewExpression parses a new expression
func (a *JSAnalyzer) parseNewExpression(node *sitter.Node, source []byte, scope string) *types.CallSite {
	call := &types.CallSite{
		Line:          int(node.StartPoint().Row) + 1,
		Column:        int(node.StartPoint().Column),
		Scope:         scope,
		IsConstructor: true,
		Arguments:     make([]types.CallArg, 0),
	}

	// Get class name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" || child.Type() == "member_expression" {
			call.ClassName = analyzer.GetNodeText(child, source)
			call.FunctionName = "new " + call.ClassName
			break
		}
	}

	// Get arguments
	argsNode := analyzer.FindChildByType(node, "arguments")
	if argsNode != nil {
		call.Arguments = a.parseCallArguments(argsNode, source)
	}

	return call
}

// parseCallArguments parses function call arguments
func (a *JSAnalyzer) parseCallArguments(node *sitter.Node, source []byte) []types.CallArg {
	var args []types.CallArg

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "(" || child.Type() == ")" || child.Type() == "," {
			continue
		}

		arg := types.CallArg{
			Index: len(args),
			Value: analyzer.GetNodeText(child, source),
		}

		// Check if tainted
		arg.IsTainted, arg.TaintSource = a.isExpressionTainted(child, source)

		args = append(args, arg)
	}

	return args
}

// FindInputSources finds all user input sources in the AST
func (a *JSAnalyzer) FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error) {
	var sources []*types.FlowNode

	// Find member expressions that could be sources
	memberExprs := analyzer.FindNodesOfType(root, "member_expression")
	for _, node := range memberExprs {
		text := analyzer.GetNodeText(node, source)

		// Check global sources
		for src, sourceType := range a.globalSources {
			if strings.Contains(text, src) {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "javascript",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       src,
					Snippet:    text,
					SourceType: sourceType,
				}
				sources = append(sources, flowNode)
			}
		}

		// Check Express patterns
		expressPatterns := map[string]types.SourceType{
			"req.body":    types.SourceHTTPBody,
			"req.query":   types.SourceHTTPGet,
			"req.params":  types.SourceHTTPPath,
			"req.headers": types.SourceHTTPHeader,
			"req.cookies": types.SourceHTTPCookie,
			"request.body":    types.SourceHTTPBody,
			"request.query":   types.SourceHTTPGet,
			"request.params":  types.SourceHTTPPath,
			"request.headers": types.SourceHTTPHeader,
			"request.payload": types.SourceHTTPBody,  // Hapi
		}
		for pattern, sourceType := range expressPatterns {
			if strings.HasPrefix(text, pattern) || strings.Contains(text, "."+pattern) {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "javascript",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       pattern,
					Snippet:    text,
					SourceType: sourceType,
				}

				// Try to extract key
				if strings.Contains(text, "[") || strings.Contains(text, ".") {
					parts := strings.SplitN(text, pattern, 2)
					if len(parts) > 1 {
						key := extractJSKey(parts[1])
						if key != "" {
							flowNode.SourceKey = key
						}
					}
				}

				sources = append(sources, flowNode)
			}
		}

		// Check Koa patterns (ctx.request.body, ctx.query, ctx.params)
		koaPatterns := map[string]types.SourceType{
			"ctx.request.body": types.SourceHTTPBody,
			"ctx.query":        types.SourceHTTPGet,
			"ctx.request.query": types.SourceHTTPGet,
			"ctx.params":       types.SourceHTTPPath,
			"ctx.headers":      types.SourceHTTPHeader,
			"ctx.cookies":      types.SourceHTTPCookie,
		}
		for pattern, sourceType := range koaPatterns {
			if strings.HasPrefix(text, pattern) {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "javascript",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       pattern,
					Snippet:    text,
					SourceType: sourceType,
				}
				sources = append(sources, flowNode)
			}
		}

		// Check for .value (form input)
		if strings.HasSuffix(text, ".value") {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "javascript",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "element.value",
				Snippet:    text,
				SourceType: types.SourceUserInput,
			}
			sources = append(sources, flowNode)
		}
	}

	// Find URLSearchParams.get() calls
	callExprs := analyzer.FindNodesOfType(root, "call_expression")
	for _, node := range callExprs {
		text := analyzer.GetNodeText(node, source)

		// URLSearchParams.get()
		if strings.Contains(text, ".get(") && (strings.Contains(text, "URLSearchParams") ||
			strings.Contains(text, "searchParams") || strings.Contains(text, "params")) {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "javascript",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "URLSearchParams.get",
				Snippet:    text,
				SourceType: types.SourceHTTPGet,
			}

			// Try to extract key from get('key')
			keyRegex := regexp.MustCompile(`\.get\(['"](\w+)['"]\)`)
			if matches := keyRegex.FindStringSubmatch(text); len(matches) > 1 {
				flowNode.SourceKey = matches[1]
			}

			sources = append(sources, flowNode)
		}

		// fetch().json() or fetch().text()
		if strings.Contains(text, "fetch(") && (strings.Contains(text, ".json()") || strings.Contains(text, ".text()")) {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "javascript",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "fetch().text()",
				Snippet:    text,
				SourceType: types.SourceNetwork,
			}
			sources = append(sources, flowNode)
		}
	}

	// Find XMLHttpRequest response access
	for _, node := range memberExprs {
		text := analyzer.GetNodeText(node, source)
		if strings.Contains(text, "responseText") || strings.Contains(text, "responseXML") ||
			(strings.Contains(text, ".response") && !strings.Contains(text, "responseType")) {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "javascript",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "XMLHttpRequest.response",
				Snippet:    text,
				SourceType: types.SourceNetwork,
			}
			sources = append(sources, flowNode)
		}
	}

	return sources, nil
}

// extractJSKey extracts a key from property access like .foo or ['foo'] or ["foo"]
func extractJSKey(s string) string {
	s = strings.TrimSpace(s)

	// Check for bracket notation ['key'] or ["key"]
	bracketRegex := regexp.MustCompile(`^\[['"](\w+)['"]\]`)
	if matches := bracketRegex.FindStringSubmatch(s); len(matches) > 1 {
		return matches[1]
	}

	// Check for dot notation .key
	dotRegex := regexp.MustCompile(`^\.(\w+)`)
	if matches := dotRegex.FindStringSubmatch(s); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// DetectFrameworks detects which JavaScript frameworks are being used
func (a *JSAnalyzer) DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error) {
	var frameworks []string

	for _, imp := range symbolTable.Imports {
		path := strings.ToLower(imp.Path)

		if path == "express" || strings.HasPrefix(path, "express/") {
			if !contains(frameworks, "express") {
				frameworks = append(frameworks, "express")
			}
		}
		if path == "koa" || strings.HasPrefix(path, "koa/") || strings.HasPrefix(path, "@koa/") {
			if !contains(frameworks, "koa") {
				frameworks = append(frameworks, "koa")
			}
		}
		if path == "fastify" || strings.HasPrefix(path, "fastify/") {
			if !contains(frameworks, "fastify") {
				frameworks = append(frameworks, "fastify")
			}
		}
		if path == "hapi" || strings.HasPrefix(path, "@hapi/") {
			if !contains(frameworks, "hapi") {
				frameworks = append(frameworks, "hapi")
			}
		}
		if strings.Contains(path, "react") {
			if !contains(frameworks, "react") {
				frameworks = append(frameworks, "react")
			}
		}
		if strings.Contains(path, "vue") {
			if !contains(frameworks, "vue") {
				frameworks = append(frameworks, "vue")
			}
		}
		if strings.Contains(path, "angular") {
			if !contains(frameworks, "angular") {
				frameworks = append(frameworks, "angular")
			}
		}
	}

	return frameworks, nil
}

// AnalyzeMethodBody analyzes a method body for data flow
func (a *JSAnalyzer) AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*analyzer.MethodFlowAnalysis, error) {
	analysis := &analyzer.MethodFlowAnalysis{
		ParamsToReturn:     make([]int, 0),
		ParamsToProperties: make(map[int][]string),
		ParamsToCallArgs:   make(map[int][]*types.CallSite),
		TaintedVariables:   make(map[string]*types.TaintInfo),
		Assignments:        make([]*types.Assignment, 0),
		Calls:              make([]*types.CallSite, 0),
		Returns:            make([]analyzer.ReturnInfo, 0),
	}

	if method.BodySource == "" {
		return analysis, nil
	}

	body := method.BodySource

	// Track which parameters flow where
	for i, param := range method.Parameters {
		paramName := param.Name

		// Check if param flows to return
		if strings.Contains(body, "return") && strings.Contains(body, paramName) {
			analysis.ParamsToReturn = append(analysis.ParamsToReturn, i)
		}

		// Check if param flows to this.property
		thisAssignRegex := regexp.MustCompile(`this\.(\w+)\s*=.*` + regexp.QuoteMeta(paramName))
		matches := thisAssignRegex.FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			if len(match) > 1 {
				analysis.ParamsToProperties[i] = append(analysis.ParamsToProperties[i], match[1])
				analysis.ModifiesProperties = true
			}
		}
	}

	return analysis, nil
}

// TraceExpression traces a specific expression back to its sources
func (a *JSAnalyzer) TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error) {
	flowMap := &types.FlowMap{
		Target:   target,
		Sources:  make([]types.FlowNode, 0),
		Paths:    make([]types.FlowPath, 0),
		Carriers: make([]types.FlowNode, 0),
		AllNodes: make([]types.FlowNode, 0),
		AllEdges: make([]types.FlowEdge, 0),
		Usages:   make([]types.FlowNode, 0),
	}

	expr := target.Expression

	// Check for Express patterns
	expressPatterns := map[string]struct {
		sourceType types.SourceType
		desc       string
	}{
		"req.body":    {types.SourceHTTPBody, "HTTP POST body"},
		"req.query":   {types.SourceHTTPGet, "HTTP GET query string"},
		"req.params":  {types.SourceHTTPPath, "HTTP URL path parameters"},
		"req.headers": {types.SourceHTTPHeader, "HTTP headers"},
		"req.cookies": {types.SourceHTTPCookie, "HTTP cookies"},
	}

	for pattern, info := range expressPatterns {
		if strings.Contains(expr, pattern) {
			key := ""
			if strings.Contains(expr, "[") || strings.Count(expr, ".") > 1 {
				key = extractJSKey(strings.TrimPrefix(expr, pattern))
			}

			sourceNode := types.FlowNode{
				ID:         fmt.Sprintf("source-%s-%s", pattern, key),
				Type:       types.NodeSource,
				Language:   "javascript",
				Name:       pattern,
				Snippet:    expr,
				SourceType: info.sourceType,
				SourceKey:  key,
			}
			flowMap.Sources = append(flowMap.Sources, sourceNode)
			flowMap.AllNodes = append(flowMap.AllNodes, sourceNode)

			flowMap.CarrierChain = &types.CarrierChain{
				PropertyName:     strings.TrimPrefix(pattern, "req."),
				PopulationCalls:  []string{info.desc},
				Framework:        "express",
			}
		}
	}

	return flowMap, nil
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Register the JavaScript analyzer
func init() {
	analyzer.DefaultRegistry.Register(NewJSAnalyzer())
}
