// Package python implements the Python language analyzer for semantic input tracing
package python

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// PythonAnalyzer implements the LanguageAnalyzer interface for Python
type PythonAnalyzer struct {
	*analyzer.BaseAnalyzer
	inputSources    map[string]types.SourceType
	inputFunctions  map[string]types.SourceType
}

// NewPythonAnalyzer creates a new Python analyzer
func NewPythonAnalyzer() *PythonAnalyzer {
	a := &PythonAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer("python", languages.GetExtensionsForLanguage("python")),
	}

	// Initialize Python input sources
	a.inputSources = map[string]types.SourceType{
		// Flask
		"request.args":      types.SourceHTTPGet,
		"request.form":      types.SourceHTTPPost,
		"request.data":      types.SourceHTTPBody,
		"request.json":      types.SourceHTTPJSON,
		"request.files":     types.SourceHTTPBody,
		"request.cookies":   types.SourceHTTPCookie,
		"request.headers":   types.SourceHTTPHeader,
		"request.values":    types.SourceHTTPGet,
		// Django
		"request.GET":       types.SourceHTTPGet,
		"request.POST":      types.SourceHTTPPost,
		"request.FILES":     types.SourceHTTPBody,
		"request.COOKIES":   types.SourceHTTPCookie,
		"request.META":      types.SourceHTTPHeader,
		"request.body":      types.SourceHTTPBody,
		// aiohttp
		"request.query":     types.SourceHTTPGet,
		"request.match_info": types.SourceHTTPPath,
		"request.rel_url":   types.SourceHTTPGet,
		// Sanic
		"request.ctx":       types.SourceUserInput,
		// FastAPI / Starlette
		"request.path_params": types.SourceHTTPPath,
		"request.query_params": types.SourceHTTPGet,
		// CLI
		"sys.argv":          types.SourceCLIArg,
		"os.environ":        types.SourceEnvVar,
	}

	// Initialize input functions
	a.inputFunctions = map[string]types.SourceType{
		// Built-in
		"input":           types.SourceStdin,
		"raw_input":       types.SourceStdin,
		// OS
		"getenv":          types.SourceEnvVar,
		"os.getenv":       types.SourceEnvVar,
		"environ.get":     types.SourceEnvVar,
		// File
		"open":            types.SourceFile,
		"read":            types.SourceFile,
		"readline":        types.SourceFile,
		"readlines":       types.SourceFile,
		// Tornado
		"get_argument":     types.SourceHTTPGet,
		"get_query_argument": types.SourceHTTPGet,
		"get_body_argument": types.SourceHTTPPost,
		// aiohttp
		"request.post":     types.SourceHTTPPost,
		"request.json":     types.SourceHTTPJSON,
		// argparse
		"parse_args":       types.SourceCLIArg,
		"add_argument":     types.SourceCLIArg,
		// requests (HTTP client)
		"requests.get":     types.SourceNetwork,
		"requests.post":    types.SourceNetwork,
		"response.json":    types.SourceNetwork,
		"response.text":    types.SourceNetwork,
	}

	// Register framework patterns
	a.registerFrameworkPatterns()

	return a
}

// registerFrameworkPatterns registers known Python framework patterns
func (a *PythonAnalyzer) registerFrameworkPatterns() {
	// Flask request
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "flask_request",
		Framework:       "flask",
		Language:        "python",
		Name:            "Flask request",
		Description:     "Flask request object containing user input",
		ClassPattern:    "^Request$",
		PropertyPattern: "^(args|form|data|json|files|cookies|headers)$",
		AccessPattern:   "dict",
		SourceType:      types.SourceHTTPGet,
		CarrierClass:    "Request",
		PopulatedFrom:   []string{"HTTP request"},
		Confidence:      0.95,
	})

	// Django request
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "django_request_get",
		Framework:       "django",
		Language:        "python",
		Name:            "Django request.GET",
		Description:     "Django GET parameters",
		ClassPattern:    "^(Http)?Request$",
		PropertyPattern: "^GET$",
		AccessPattern:   "dict",
		SourceType:      types.SourceHTTPGet,
		CarrierClass:    "HttpRequest",
		CarrierProperty: "GET",
		PopulatedFrom:   []string{"query string"},
		Confidence:      0.95,
	})

	// Django POST
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "django_request_post",
		Framework:       "django",
		Language:        "python",
		Name:            "Django request.POST",
		Description:     "Django POST parameters",
		ClassPattern:    "^(Http)?Request$",
		PropertyPattern: "^POST$",
		AccessPattern:   "dict",
		SourceType:      types.SourceHTTPPost,
		CarrierClass:    "HttpRequest",
		CarrierProperty: "POST",
		PopulatedFrom:   []string{"form data"},
		Confidence:      0.95,
	})

	// FastAPI
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "fastapi_request",
		Framework:       "fastapi",
		Language:        "python",
		Name:            "FastAPI Request",
		Description:     "FastAPI request object",
		ClassPattern:    "^Request$",
		SourceType:      types.SourceHTTPBody,
		CarrierClass:    "Request",
		PopulatedFrom:   []string{"HTTP request"},
		Confidence:      0.9,
	})

	// argparse
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "argparse",
		Framework:       "argparse",
		Language:        "python",
		Name:            "argparse arguments",
		Description:     "Command line arguments parsed by argparse",
		MethodPattern:   "^parse_args$",
		SourceType:      types.SourceCLIArg,
		PopulatedFrom:   []string{"sys.argv"},
		Confidence:      0.95,
	})
}

// BuildSymbolTable builds the symbol table for a Python file
func (a *PythonAnalyzer) BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error) {
	st := types.NewSymbolTable(filePath, "python")

	// Extract imports
	st.Imports = a.extractImports(root, source)

	// Extract classes
	classes, err := a.ExtractClasses(root, source)
	if err != nil {
		return nil, err
	}
	for _, class := range classes {
		class.FilePath = filePath
		st.Classes[class.Name] = class
	}

	// Extract functions
	functions, err := a.ExtractFunctions(root, source)
	if err != nil {
		return nil, err
	}
	for _, fn := range functions {
		fn.FilePath = filePath
		st.Functions[fn.Name] = fn
	}

	// Detect frameworks
	frameworks, _ := a.DetectFrameworks(st, source)
	if len(frameworks) > 0 {
		st.Framework = frameworks[0]
	}

	return st, nil
}

// extractImports extracts import statements from a Python file
func (a *PythonAnalyzer) extractImports(root *sitter.Node, source []byte) []types.ImportInfo {
	var imports []types.ImportInfo

	// import statements
	importNodes := analyzer.FindNodesOfType(root, "import_statement")
	for _, node := range importNodes {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "dotted_name" {
				imports = append(imports, types.ImportInfo{
					Path: analyzer.GetNodeText(child, source),
					Line: int(node.StartPoint().Row) + 1,
					Type: "import",
				})
			} else if child.Type() == "aliased_import" {
				nameNode := child.ChildByFieldName("name")
				aliasNode := child.ChildByFieldName("alias")
				path := ""
				alias := ""
				if nameNode != nil {
					path = analyzer.GetNodeText(nameNode, source)
				}
				if aliasNode != nil {
					alias = analyzer.GetNodeText(aliasNode, source)
				}
				imports = append(imports, types.ImportInfo{
					Path:  path,
					Alias: alias,
					Line:  int(node.StartPoint().Row) + 1,
					Type:  "import",
				})
			}
		}
	}

	// from ... import statements
	fromImportNodes := analyzer.FindNodesOfType(root, "import_from_statement")
	for _, node := range fromImportNodes {
		moduleNode := node.ChildByFieldName("module_name")
		modulePath := ""
		if moduleNode != nil {
			modulePath = analyzer.GetNodeText(moduleNode, source)
		}

		// Find imported names
		var names []string
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "dotted_name" && child != moduleNode {
				names = append(names, analyzer.GetNodeText(child, source))
			} else if child.Type() == "aliased_import" {
				nameNode := child.ChildByFieldName("name")
				if nameNode != nil {
					names = append(names, analyzer.GetNodeText(nameNode, source))
				}
			}
		}

		imports = append(imports, types.ImportInfo{
			Path:  modulePath,
			Names: names,
			Line:  int(node.StartPoint().Row) + 1,
			Type:  "from",
		})
	}

	return imports
}

// ResolveImports resolves import paths to actual file paths
func (a *PythonAnalyzer) ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error) {
	var resolvedPaths []string
	// Python imports are typically resolved at runtime, so this is limited
	// We could add support for PYTHONPATH-based resolution here
	return resolvedPaths, nil
}

// ExtractClasses extracts class definitions from Python AST
func (a *PythonAnalyzer) ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error) {
	var classes []*types.ClassDef

	classNodes := analyzer.FindNodesOfType(root, "class_definition")
	for _, classNode := range classNodes {
		class := a.parseClassDefinition(classNode, source)
		if class != nil {
			classes = append(classes, class)
		}
	}

	return classes, nil
}

// parseClassDefinition parses a class definition node
func (a *PythonAnalyzer) parseClassDefinition(node *sitter.Node, source []byte) *types.ClassDef {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := analyzer.GetNodeText(nameNode, source)
	class := types.NewClassDef(name, "", int(node.StartPoint().Row)+1)
	class.EndLine = int(node.EndPoint().Row) + 1

	// Get base classes (first one as Extends, rest as Implements for compatibility)
	superclassesNode := node.ChildByFieldName("superclasses")
	if superclassesNode != nil {
		first := true
		for i := 0; i < int(superclassesNode.ChildCount()); i++ {
			child := superclassesNode.Child(i)
			if child.Type() == "identifier" || child.Type() == "attribute" {
				baseName := analyzer.GetNodeText(child, source)
				if first {
					class.Extends = baseName
					first = false
				} else {
					class.Implements = append(class.Implements, baseName)
				}
			}
		}
	}

	// Parse class body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		a.parseClassBody(class, bodyNode, source)
	}

	return class
}

// parseClassBody parses the body of a class
func (a *PythonAnalyzer) parseClassBody(class *types.ClassDef, bodyNode *sitter.Node, source []byte) {
	for i := 0; i < int(bodyNode.ChildCount()); i++ {
		child := bodyNode.Child(i)
		switch child.Type() {
		case "function_definition":
			method := a.parseMethodDefinition(child, source)
			if method != nil {
				if method.Name == "__init__" {
					class.Constructor = method
				}
				class.Methods[method.Name] = method
			}
		case "expression_statement":
			// Class-level assignments could be properties
			exprNode := child.Child(0)
			if exprNode != nil && exprNode.Type() == "assignment" {
				leftNode := exprNode.ChildByFieldName("left")
				if leftNode != nil && leftNode.Type() == "identifier" {
					propName := analyzer.GetNodeText(leftNode, source)
					class.Properties[propName] = &types.PropertyDef{
						Name:       propName,
						Visibility: "public",
						Line:       int(child.StartPoint().Row) + 1,
					}
				}
			}
		}
	}
}

// parseMethodDefinition parses a method definition
func (a *PythonAnalyzer) parseMethodDefinition(node *sitter.Node, source []byte) *types.MethodDef {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := analyzer.GetNodeText(nameNode, source)
	method := &types.MethodDef{
		Name:       name,
		Line:       int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Visibility: "public",
		Parameters: make([]types.ParameterDef, 0),
	}

	// Check for decorators
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "decorator" {
			decoratorText := analyzer.GetNodeText(child, source)
			if strings.Contains(decoratorText, "staticmethod") {
				method.IsStatic = true
			} else if strings.Contains(decoratorText, "async") {
				method.IsAsync = true
			}
		}
	}

	// Check for async def
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "async" {
			method.IsAsync = true
		}
	}

	// Parse parameters
	paramsNode := node.ChildByFieldName("parameters")
	if paramsNode != nil {
		method.Parameters = a.parseParameters(paramsNode, source)
	}

	// Get return type annotation
	returnNode := node.ChildByFieldName("return_type")
	if returnNode != nil {
		method.ReturnType = analyzer.GetNodeText(returnNode, source)
	}

	// Get body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		method.BodyStart = int(bodyNode.StartPoint().Row) + 1
		method.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		method.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	// Determine visibility from name convention
	if strings.HasPrefix(name, "__") && !strings.HasSuffix(name, "__") {
		method.Visibility = "private"
	} else if strings.HasPrefix(name, "_") {
		method.Visibility = "protected"
	}

	return method
}

// parseParameters parses function/method parameters
func (a *PythonAnalyzer) parseParameters(node *sitter.Node, source []byte) []types.ParameterDef {
	var params []types.ParameterDef
	index := 0

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()

		switch childType {
		case "identifier":
			name := analyzer.GetNodeText(child, source)
			// Skip self and cls
			if name != "self" && name != "cls" {
				params = append(params, types.ParameterDef{
					Name:  name,
					Index: index,
				})
				index++
			}

		case "typed_parameter":
			nameNode := child.Child(0)
			if nameNode != nil {
				name := analyzer.GetNodeText(nameNode, source)
				if name != "self" && name != "cls" {
					param := types.ParameterDef{
						Name:  name,
						Index: index,
					}
					// Get type annotation
					typeNode := child.ChildByFieldName("type")
					if typeNode != nil {
						param.Type = analyzer.GetNodeText(typeNode, source)
					}
					params = append(params, param)
					index++
				}
			}

		case "default_parameter":
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				name := analyzer.GetNodeText(nameNode, source)
				if name != "self" && name != "cls" {
					param := types.ParameterDef{
						Name:  name,
						Index: index,
					}
					valueNode := child.ChildByFieldName("value")
					if valueNode != nil {
						param.DefaultValue = analyzer.GetNodeText(valueNode, source)
					}
					params = append(params, param)
					index++
				}
			}

		case "typed_default_parameter":
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				name := analyzer.GetNodeText(nameNode, source)
				if name != "self" && name != "cls" {
					param := types.ParameterDef{
						Name:  name,
						Index: index,
					}
					typeNode := child.ChildByFieldName("type")
					if typeNode != nil {
						param.Type = analyzer.GetNodeText(typeNode, source)
					}
					valueNode := child.ChildByFieldName("value")
					if valueNode != nil {
						param.DefaultValue = analyzer.GetNodeText(valueNode, source)
					}
					params = append(params, param)
					index++
				}
			}

		case "list_splat_pattern":
			// *args
			for j := 0; j < int(child.ChildCount()); j++ {
				if idNode := child.Child(j); idNode != nil && idNode.Type() == "identifier" {
					params = append(params, types.ParameterDef{
						Name:       analyzer.GetNodeText(idNode, source),
						Index:      index,
						IsVariadic: true,
					})
					index++
					break
				}
			}

		case "dictionary_splat_pattern":
			// **kwargs
			for j := 0; j < int(child.ChildCount()); j++ {
				if idNode := child.Child(j); idNode != nil && idNode.Type() == "identifier" {
					params = append(params, types.ParameterDef{
						Name:       analyzer.GetNodeText(idNode, source),
						Index:      index,
						IsVariadic: true,
					})
					index++
					break
				}
			}
		}
	}

	return params
}

// ExtractFunctions extracts standalone function definitions
func (a *PythonAnalyzer) ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error) {
	var functions []*types.FunctionDef

	funcNodes := analyzer.FindNodesOfType(root, "function_definition")
	for _, funcNode := range funcNodes {
		// Skip if inside a class
		if analyzer.GetEnclosingClass(funcNode, []string{"class_definition"}) != nil {
			continue
		}

		fn := a.parseFunctionDefinition(funcNode, source)
		if fn != nil {
			functions = append(functions, fn)
		}
	}

	return functions, nil
}

// parseFunctionDefinition parses a function definition node
func (a *PythonAnalyzer) parseFunctionDefinition(node *sitter.Node, source []byte) *types.FunctionDef {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	fn := &types.FunctionDef{
		Name:       analyzer.GetNodeText(nameNode, source),
		Line:       int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Parameters: make([]types.ParameterDef, 0),
	}

	// Check for async
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "async" {
			fn.IsAsync = true
		}
	}

	// Parse parameters
	paramsNode := node.ChildByFieldName("parameters")
	if paramsNode != nil {
		fn.Parameters = a.parseParameters(paramsNode, source)
	}

	// Get return type
	returnNode := node.ChildByFieldName("return_type")
	if returnNode != nil {
		fn.ReturnType = analyzer.GetNodeText(returnNode, source)
	}

	// Get body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		fn.BodyStart = int(bodyNode.StartPoint().Row) + 1
		fn.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		fn.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	return fn
}

// ExtractAssignments extracts all assignments from the AST
func (a *PythonAnalyzer) ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error) {
	var assignments []*types.Assignment

	assignNodes := analyzer.FindNodesOfType(root, "assignment")
	for _, node := range assignNodes {
		assignment := a.parseAssignment(node, source, scope)
		if assignment != nil {
			assignments = append(assignments, assignment)
		}
	}

	// Augmented assignments (+=, -=, etc.)
	augmentedNodes := analyzer.FindNodesOfType(root, "augmented_assignment")
	for _, node := range augmentedNodes {
		assignment := a.parseAssignment(node, source, scope)
		if assignment != nil {
			assignments = append(assignments, assignment)
		}
	}

	return assignments, nil
}

// parseAssignment parses an assignment expression
func (a *PythonAnalyzer) parseAssignment(node *sitter.Node, source []byte, scope string) *types.Assignment {
	leftNode := node.ChildByFieldName("left")
	rightNode := node.ChildByFieldName("right")

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

	// Determine target type
	switch leftNode.Type() {
	case "identifier":
		assignment.TargetType = "variable"
	case "attribute":
		assignment.TargetType = "property"
	case "subscript":
		assignment.TargetType = "array_element"
	}

	// Check if source is tainted
	assignment.IsTainted, assignment.TaintSource = a.isExpressionTainted(rightNode, source)

	return assignment
}

// isExpressionTainted checks if an expression contains tainted data
func (a *PythonAnalyzer) isExpressionTainted(node *sitter.Node, source []byte) (bool, string) {
	if node == nil {
		return false, ""
	}

	text := analyzer.GetNodeText(node, source)

	// Check for known input sources
	for src := range a.inputSources {
		if strings.Contains(text, src) {
			return true, src
		}
	}

	// Check for input functions
	for fn := range a.inputFunctions {
		if strings.Contains(text, fn+"(") {
			return true, fn + "()"
		}
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
func (a *PythonAnalyzer) ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error) {
	var calls []*types.CallSite

	callNodes := analyzer.FindNodesOfType(root, "call")
	for _, node := range callNodes {
		call := a.parseCall(node, source, scope)
		if call != nil {
			calls = append(calls, call)
		}
	}

	return calls, nil
}

// parseCall parses a function call expression
func (a *PythonAnalyzer) parseCall(node *sitter.Node, source []byte, scope string) *types.CallSite {
	funcNode := node.ChildByFieldName("function")
	if funcNode == nil {
		return nil
	}

	call := &types.CallSite{
		FunctionName: analyzer.GetNodeText(funcNode, source),
		Line:         int(node.StartPoint().Row) + 1,
		Column:       int(node.StartPoint().Column),
		Scope:        scope,
		Arguments:    make([]types.CallArg, 0),
	}

	// Check for method call (attribute access)
	if funcNode.Type() == "attribute" {
		objNode := funcNode.ChildByFieldName("object")
		attrNode := funcNode.ChildByFieldName("attribute")
		if objNode != nil {
			call.ClassName = analyzer.GetNodeText(objNode, source)
		}
		if attrNode != nil {
			call.MethodName = analyzer.GetNodeText(attrNode, source)
		}
	}

	// Parse arguments
	argsNode := node.ChildByFieldName("arguments")
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

// parseCallArguments parses function call arguments
func (a *PythonAnalyzer) parseCallArguments(node *sitter.Node, source []byte) []types.CallArg {
	var args []types.CallArg
	index := 0

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()

		// Skip punctuation
		if childType == "(" || childType == ")" || childType == "," {
			continue
		}

		arg := types.CallArg{
			Index: index,
			Value: analyzer.GetNodeText(child, source),
		}

		// Check if tainted
		arg.IsTainted, arg.TaintSource = a.isExpressionTainted(child, source)

		args = append(args, arg)
		index++
	}

	return args
}

// FindInputSources finds all user input sources in the AST
func (a *PythonAnalyzer) FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error) {
	var sources []*types.FlowNode

	// Find attribute accesses (request.args, request.form, etc.)
	attrNodes := analyzer.FindNodesOfType(root, "attribute")
	for _, node := range attrNodes {
		text := analyzer.GetNodeText(node, source)
		for src, sourceType := range a.inputSources {
			if strings.HasPrefix(text, src) {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "python",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       src,
					Snippet:    text,
					SourceType: sourceType,
				}

				// Try to extract key
				parent := node.Parent()
				if parent != nil && parent.Type() == "subscript" {
					flowNode.Snippet = analyzer.GetNodeText(parent, source)
					// Extract key from subscript
					for j := 0; j < int(parent.ChildCount()); j++ {
						child := parent.Child(j)
						if child.Type() == "string" {
							key := analyzer.GetNodeText(child, source)
							flowNode.SourceKey = strings.Trim(key, "\"'")
							break
						}
					}
				}

				sources = append(sources, flowNode)
			}
		}
	}

	// Find function calls (input(), getenv(), etc.)
	callNodes := analyzer.FindNodesOfType(root, "call")
	for _, node := range callNodes {
		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			continue
		}
		funcName := analyzer.GetNodeText(funcNode, source)

		if sourceType, ok := a.inputFunctions[funcName]; ok {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "python",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       funcName,
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: sourceType,
			}
			sources = append(sources, flowNode)
		}
	}

	// Find subscript accesses (sys.argv[x], os.environ[x])
	subscriptNodes := analyzer.FindNodesOfType(root, "subscript")
	for _, node := range subscriptNodes {
		valueNode := node.ChildByFieldName("value")
		if valueNode == nil {
			continue
		}
		valueName := analyzer.GetNodeText(valueNode, source)

		if strings.Contains(valueName, "sys.argv") || valueName == "argv" {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "python",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "sys.argv",
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: types.SourceCLIArg,
			}
			sources = append(sources, flowNode)
		} else if strings.Contains(valueName, "os.environ") || valueName == "environ" {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "python",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "os.environ",
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: types.SourceEnvVar,
			}
			sources = append(sources, flowNode)
		}
	}

	return sources, nil
}

// DetectFrameworks detects which Python frameworks are being used
func (a *PythonAnalyzer) DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error) {
	var frameworks []string

	// Check imports for framework hints
	for _, imp := range symbolTable.Imports {
		path := strings.ToLower(imp.Path)

		if strings.Contains(path, "flask") {
			if !contains(frameworks, "flask") {
				frameworks = append(frameworks, "flask")
			}
		}
		if strings.Contains(path, "django") {
			if !contains(frameworks, "django") {
				frameworks = append(frameworks, "django")
			}
		}
		if strings.Contains(path, "fastapi") {
			if !contains(frameworks, "fastapi") {
				frameworks = append(frameworks, "fastapi")
			}
		}
		if strings.Contains(path, "tornado") {
			if !contains(frameworks, "tornado") {
				frameworks = append(frameworks, "tornado")
			}
		}
		if strings.Contains(path, "bottle") {
			if !contains(frameworks, "bottle") {
				frameworks = append(frameworks, "bottle")
			}
		}
		if strings.Contains(path, "pyramid") {
			if !contains(frameworks, "pyramid") {
				frameworks = append(frameworks, "pyramid")
			}
		}
	}

	return frameworks, nil
}

// AnalyzeMethodBody analyzes a method body for data flow
func (a *PythonAnalyzer) AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*analyzer.MethodFlowAnalysis, error) {
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

		// Check if param flows to self.property
		selfAssignRegex := regexp.MustCompile(`self\.(\w+)\s*=.*\b` + regexp.QuoteMeta(paramName) + `\b`)
		matches := selfAssignRegex.FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			if len(match) > 1 {
				analysis.ParamsToProperties[i] = append(analysis.ParamsToProperties[i], match[1])
				analysis.ModifiesProperties = true
			}
		}
	}

	// Check if method returns input directly
	for src := range a.inputSources {
		if strings.Contains(body, "return") && strings.Contains(body, src) {
			analysis.ReturnsInput = true
			break
		}
	}

	return analysis, nil
}

// TraceExpression traces a specific expression back to its sources
func (a *PythonAnalyzer) TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error) {
	flowMap := types.NewFlowMap()
	flowMap.Target = target

	expr := target.Expression

	// Check for framework patterns first
	for _, pattern := range a.GetFrameworkPatterns() {
		if a.matchesFrameworkPattern(expr, pattern) {
			flowMap.CarrierChain = &types.CarrierChain{
				ClassName:        pattern.CarrierClass,
				PropertyName:     pattern.CarrierProperty,
				PopulationMethod: pattern.PopulatedBy,
				PopulationCalls:  pattern.PopulatedFrom,
				Framework:        pattern.Framework,
			}

			// Add sources based on framework knowledge
			for _, src := range pattern.PopulatedFrom {
				sourceNode := types.FlowNode{
					ID:         fmt.Sprintf("source-%s", src),
					Type:       types.NodeSource,
					Language:   "python",
					Name:       src,
					Snippet:    src,
					SourceType: pattern.SourceType,
				}
				flowMap.AddSource(sourceNode)
			}

			break
		}
	}

	// If no framework pattern matched, check for direct input sources
	if len(flowMap.Sources) == 0 {
		for src, sourceType := range a.inputSources {
			if strings.Contains(expr, src) {
				key := a.extractKeyFromExpression(expr)
				sourceNode := types.FlowNode{
					ID:         fmt.Sprintf("source-%s-%s", src, key),
					Type:       types.NodeSource,
					Language:   "python",
					Name:       src,
					Snippet:    expr,
					SourceType: sourceType,
					SourceKey:  key,
				}
				flowMap.AddSource(sourceNode)
			}
		}
	}

	return flowMap, nil
}

// matchesFrameworkPattern checks if an expression matches a framework pattern
func (a *PythonAnalyzer) matchesFrameworkPattern(expr string, pattern *types.FrameworkPattern) bool {
	if pattern.PropertyPattern != "" {
		regex := regexp.MustCompile(`\b` + pattern.PropertyPattern)
		return regex.MatchString(expr)
	}

	if pattern.MethodPattern != "" {
		regex := regexp.MustCompile(`\.` + pattern.MethodPattern + `\(`)
		return regex.MatchString(expr)
	}

	return false
}

// extractKeyFromExpression extracts the key from an expression like request.args['key']
func (a *PythonAnalyzer) extractKeyFromExpression(expr string) string {
	// Match ['key'] or ["key"]
	regex := regexp.MustCompile(`\[['"](\w+)['"]\]`)
	matches := regex.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}

	// Match .get('key')
	regex2 := regexp.MustCompile(`\.get\(['"](\w+)['"]\)`)
	matches2 := regex2.FindStringSubmatch(expr)
	if len(matches2) > 1 {
		return matches2[1]
	}

	return ""
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

// Register the Python analyzer
func init() {
	analyzer.DefaultRegistry.Register(NewPythonAnalyzer())
}
