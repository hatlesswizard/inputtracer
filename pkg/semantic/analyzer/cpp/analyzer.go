// Package cpp implements the C++ language analyzer for semantic input tracing
package cpp

import (
	"fmt"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/mappings"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// CPPAnalyzer implements the LanguageAnalyzer interface for C++
type CPPAnalyzer struct {
	*analyzer.BaseAnalyzer
	inputFunctions  map[string]types.SourceType
	cgiEnvVars      map[string]types.SourceType
	qtInputMethods  map[string]types.SourceType
	frameworkTypes  map[string]mappings.FrameworkTypeInfo
	methodInputs    map[string]types.SourceType
}

// NewCPPAnalyzer creates a new C++ analyzer
func NewCPPAnalyzer() *CPPAnalyzer {
	m := mappings.GetMappings("cpp")
	a := &CPPAnalyzer{
		BaseAnalyzer:   analyzer.NewBaseAnalyzer("cpp", languages.GetExtensionsForLanguage("cpp")),
		inputFunctions: m.GetInputFunctionsMap(),
		cgiEnvVars:     m.GetCGIEnvVarsMap(),
		qtInputMethods: m.GetQtInputMethodsMap(),
		frameworkTypes: m.GetFrameworkTypesMap(),
		methodInputs:   m.GetMethodInputsMap(),
	}

	return a
}

func (a *CPPAnalyzer) BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error) {
	st := types.NewSymbolTable(filePath, "cpp")
	st.Imports = a.extractIncludes(root, source)

	classes, _ := a.ExtractClasses(root, source)
	for _, class := range classes {
		class.FilePath = filePath
		st.Classes[class.Name] = class
	}

	functions, _ := a.ExtractFunctions(root, source)
	for _, fn := range functions {
		fn.FilePath = filePath
		st.Functions[fn.Name] = fn
	}

	return st, nil
}

func (a *CPPAnalyzer) extractIncludes(root *sitter.Node, source []byte) []types.ImportInfo {
	var imports []types.ImportInfo

	includeNodes := analyzer.FindNodesOfType(root, "preproc_include")
	for _, node := range includeNodes {
		pathNode := node.ChildByFieldName("path")
		if pathNode != nil {
			path := strings.Trim(analyzer.GetNodeText(pathNode, source), `<>"`)
			imports = append(imports, types.ImportInfo{
				Path: path,
				Line: int(node.StartPoint().Row) + 1,
				Type: "include",
			})
		}
	}

	// using declarations
	usingNodes := analyzer.FindNodesOfType(root, "using_declaration")
	for _, node := range usingNodes {
		text := analyzer.GetNodeText(node, source)
		if strings.Contains(text, "namespace") {
			ns := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(text, "using namespace"), ";"))
			imports = append(imports, types.ImportInfo{
				Path: ns,
				Line: int(node.StartPoint().Row) + 1,
				Type: "using",
			})
		}
	}

	return imports
}

func (a *CPPAnalyzer) ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error) {
	return nil, nil
}

func (a *CPPAnalyzer) ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error) {
	var classes []*types.ClassDef

	classNodes := analyzer.FindNodesOfType(root, "class_specifier")
	for _, classNode := range classNodes {
		class := a.parseClassDefinition(classNode, source)
		if class != nil {
			classes = append(classes, class)
		}
	}

	structNodes := analyzer.FindNodesOfType(root, "struct_specifier")
	for _, structNode := range structNodes {
		class := a.parseClassDefinition(structNode, source)
		if class != nil {
			classes = append(classes, class)
		}
	}

	return classes, nil
}

func (a *CPPAnalyzer) parseClassDefinition(node *sitter.Node, source []byte) *types.ClassDef {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := analyzer.GetNodeText(nameNode, source)
	class := types.NewClassDef(name, "", int(node.StartPoint().Row)+1)
	class.EndLine = int(node.EndPoint().Row) + 1

	// Extract base classes
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "base_class_clause" {
			for j := 0; j < int(child.ChildCount()); j++ {
				baseSpec := child.Child(j)
				if baseSpec.Type() == "base_class_specifier" {
					baseTypeNode := baseSpec.ChildByFieldName("type")
					if baseTypeNode != nil {
						baseName := analyzer.GetNodeText(baseTypeNode, source)
						if class.Extends == "" {
							class.Extends = baseName
						} else {
							class.Implements = append(class.Implements, baseName)
						}
					}
				}
			}
		}
	}

	// Parse class body for methods
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		a.parseClassBody(class, bodyNode, source)
	}

	return class
}

func (a *CPPAnalyzer) parseClassBody(class *types.ClassDef, bodyNode *sitter.Node, source []byte) {
	funcNodes := analyzer.FindNodesOfType(bodyNode, "function_definition")
	for _, funcNode := range funcNodes {
		method := a.parseMethodDefinition(funcNode, source)
		if method != nil {
			class.Methods[method.Name] = method
		}
	}

	// Declaration specifiers for methods declared in class
	declNodes := analyzer.FindNodesOfType(bodyNode, "declaration")
	for _, declNode := range declNodes {
		// Check if this is a function declaration
		for i := 0; i < int(declNode.ChildCount()); i++ {
			child := declNode.Child(i)
			if child.Type() == "function_declarator" {
				nameNode := child.ChildByFieldName("declarator")
				if nameNode != nil {
					name := a.findFunctionName(nameNode, source)
					if name != "" {
						method := &types.MethodDef{
							Name: name,
							Line: int(declNode.StartPoint().Row) + 1,
						}
						class.Methods[method.Name] = method
					}
				}
			}
		}
	}
}

func (a *CPPAnalyzer) parseMethodDefinition(node *sitter.Node, source []byte) *types.MethodDef {
	declaratorNode := node.ChildByFieldName("declarator")
	if declaratorNode == nil {
		return nil
	}

	funcName := a.findFunctionName(declaratorNode, source)
	if funcName == "" {
		return nil
	}

	method := &types.MethodDef{
		Name:       funcName,
		Line:       int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Visibility: "public",
		Parameters: make([]types.ParameterDef, 0),
	}

	typeNode := node.ChildByFieldName("type")
	if typeNode != nil {
		method.ReturnType = analyzer.GetNodeText(typeNode, source)
	}

	paramsNode := a.findParameterList(declaratorNode)
	if paramsNode != nil {
		method.Parameters = a.parseParameters(paramsNode, source)
	}

	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		method.BodyStart = int(bodyNode.StartPoint().Row) + 1
		method.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		method.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	// Check for static
	nodeContent := analyzer.GetNodeText(node, source)
	if strings.Contains(nodeContent, "static ") {
		method.IsStatic = true
	}

	return method
}

func (a *CPPAnalyzer) ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error) {
	var functions []*types.FunctionDef

	funcNodes := analyzer.FindNodesOfType(root, "function_definition")
	for _, funcNode := range funcNodes {
		// Skip if inside a class
		if analyzer.GetEnclosingClass(funcNode, []string{"class_specifier", "struct_specifier"}) != nil {
			continue
		}

		fn := a.parseFunctionDefinition(funcNode, source)
		if fn != nil {
			functions = append(functions, fn)
		}
	}

	return functions, nil
}

func (a *CPPAnalyzer) parseFunctionDefinition(node *sitter.Node, source []byte) *types.FunctionDef {
	declaratorNode := node.ChildByFieldName("declarator")
	if declaratorNode == nil {
		return nil
	}

	funcName := a.findFunctionName(declaratorNode, source)
	if funcName == "" {
		return nil
	}

	fn := &types.FunctionDef{
		Name:       funcName,
		Line:       int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Parameters: make([]types.ParameterDef, 0),
	}

	typeNode := node.ChildByFieldName("type")
	if typeNode != nil {
		fn.ReturnType = analyzer.GetNodeText(typeNode, source)
	}

	paramsNode := a.findParameterList(declaratorNode)
	if paramsNode != nil {
		fn.Parameters = a.parseParameters(paramsNode, source)
	}

	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		fn.BodyStart = int(bodyNode.StartPoint().Row) + 1
		fn.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		fn.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	return fn
}

func (a *CPPAnalyzer) findFunctionName(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}

	nodeType := node.Type()

	if nodeType == "identifier" {
		return analyzer.GetNodeText(node, source)
	}

	if nodeType == "qualified_identifier" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			return analyzer.GetNodeText(nameNode, source)
		}
	}

	if nodeType == "function_declarator" {
		declarator := node.ChildByFieldName("declarator")
		if declarator != nil {
			return a.findFunctionName(declarator, source)
		}
	}

	if nodeType == "pointer_declarator" || nodeType == "reference_declarator" {
		declarator := node.ChildByFieldName("declarator")
		if declarator != nil {
			return a.findFunctionName(declarator, source)
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if name := a.findFunctionName(child, source); name != "" {
			return name
		}
	}

	return ""
}

func (a *CPPAnalyzer) findParameterList(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}

	if node.Type() == "function_declarator" {
		return node.ChildByFieldName("parameters")
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if params := a.findParameterList(child); params != nil {
			return params
		}
	}

	return nil
}

func (a *CPPAnalyzer) parseParameters(node *sitter.Node, source []byte) []types.ParameterDef {
	var params []types.ParameterDef
	index := 0

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()
		if childType == "parameter_declaration" || childType == "optional_parameter_declaration" {
			param := types.ParameterDef{Index: index}

			typeNode := child.ChildByFieldName("type")
			if typeNode != nil {
				param.Type = analyzer.GetNodeText(typeNode, source)
			}

			declaratorNode := child.ChildByFieldName("declarator")
			if declaratorNode != nil {
				param.Name = a.findParameterName(declaratorNode, source)
			}

			defaultNode := child.ChildByFieldName("default_value")
			if defaultNode != nil {
				param.DefaultValue = analyzer.GetNodeText(defaultNode, source)
			}

			if param.Name != "" {
				params = append(params, param)
				index++
			}
		} else if childType == "variadic_parameter_declaration" {
			params = append(params, types.ParameterDef{
				Name:       "...",
				Index:      index,
				IsVariadic: true,
			})
			index++
		}
	}

	return params
}

func (a *CPPAnalyzer) findParameterName(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}

	if node.Type() == "identifier" {
		return analyzer.GetNodeText(node, source)
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if name := a.findParameterName(child, source); name != "" {
			return name
		}
	}

	return ""
}

func (a *CPPAnalyzer) ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error) {
	var assignments []*types.Assignment

	assignNodes := analyzer.FindNodesOfType(root, "assignment_expression")
	for _, node := range assignNodes {
		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")
		if leftNode != nil && rightNode != nil {
			assignment := &types.Assignment{
				Target:   analyzer.GetNodeText(leftNode, source),
				Source:   analyzer.GetNodeText(rightNode, source),
				Line:     int(node.StartPoint().Row) + 1,
				Column:   int(node.StartPoint().Column),
				Scope:    scope,
			}
			assignment.IsTainted, assignment.TaintSource = a.isExpressionTainted(rightNode, source)
			assignments = append(assignments, assignment)
		}
	}

	return assignments, nil
}

func (a *CPPAnalyzer) isExpressionTainted(node *sitter.Node, source []byte) (bool, string) {
	if node == nil {
		return false, ""
	}

	text := analyzer.GetNodeText(node, source)

	// Check for cin stream extraction
	if strings.Contains(text, "cin") && strings.Contains(text, ">>") {
		return true, "std::cin"
	}

	// Check for ifstream/fstream extraction
	if strings.Contains(text, ">>") && (strings.Contains(text, "ifstream") || strings.Contains(text, "fstream")) {
		return true, "ifstream>>"
	}

	// Check for input functions
	for fn := range a.inputFunctions {
		if strings.Contains(text, fn+"(") || strings.Contains(text, fn+" ") {
			return true, fn + "()"
		}
	}

	// Check for Qt input methods
	for method := range a.qtInputMethods {
		if strings.Contains(text, "."+method+"(") || strings.Contains(text, "->"+method+"(") {
			return true, "Qt:" + method + "()"
		}
	}

	// Check for framework method inputs
	for method := range a.methodInputs {
		if strings.Contains(text, "."+method) || strings.Contains(text, "->"+method) {
			return true, "framework:" + method
		}
	}

	// Check for framework types
	for typeName, info := range a.frameworkTypes {
		if strings.Contains(text, typeName) {
			return true, info.Framework + ":" + typeName
		}
	}

	// Check for CGI environment variables
	for envVar := range a.cgiEnvVars {
		if strings.Contains(text, `"`+envVar+`"`) || strings.Contains(text, `'`+envVar+`'`) {
			return true, "CGI:" + envVar
		}
	}

	// Check for argv, stdin, envp
	if strings.Contains(text, "argv[") {
		return true, "CLI:argv"
	}
	if strings.Contains(text, "envp[") {
		return true, "ENV:envp"
	}
	if strings.Contains(text, "stdin") {
		return true, "stdin"
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		if tainted, src := a.isExpressionTainted(node.Child(i), source); tainted {
			return true, src
		}
	}

	return false, ""
}

func (a *CPPAnalyzer) ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error) {
	var calls []*types.CallSite

	callNodes := analyzer.FindNodesOfType(root, "call_expression")
	for _, node := range callNodes {
		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			continue
		}

		call := &types.CallSite{
			FunctionName: analyzer.GetNodeText(funcNode, source),
			Line:         int(node.StartPoint().Row) + 1,
			Column:       int(node.StartPoint().Column),
			Scope:        scope,
			Arguments:    make([]types.CallArg, 0),
		}

		if funcNode.Type() == "field_expression" {
			argNode := funcNode.ChildByFieldName("argument")
			fieldNode := funcNode.ChildByFieldName("field")
			if argNode != nil {
				call.ClassName = analyzer.GetNodeText(argNode, source)
			}
			if fieldNode != nil {
				call.MethodName = analyzer.GetNodeText(fieldNode, source)
			}
		}

		argsNode := node.ChildByFieldName("arguments")
		if argsNode != nil {
			call.Arguments = a.parseCallArguments(argsNode, source)
		}

		for i, arg := range call.Arguments {
			if arg.IsTainted {
				call.HasTaintedArgs = true
				call.TaintedArgIndices = append(call.TaintedArgIndices, i)
			}
		}

		calls = append(calls, call)
	}

	return calls, nil
}

func (a *CPPAnalyzer) parseCallArguments(node *sitter.Node, source []byte) []types.CallArg {
	var args []types.CallArg
	index := 0

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()

		if childType == "(" || childType == ")" || childType == "," {
			continue
		}

		arg := types.CallArg{
			Index: index,
			Value: analyzer.GetNodeText(child, source),
		}
		arg.IsTainted, arg.TaintSource = a.isExpressionTainted(child, source)

		args = append(args, arg)
		index++
	}

	return args
}

func (a *CPPAnalyzer) FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error) {
	var sources []*types.FlowNode

	// Call expressions (getline, scanf, etc.)
	callNodes := analyzer.FindNodesOfType(root, "call_expression")
	for _, node := range callNodes {
		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			continue
		}
		funcName := analyzer.GetNodeText(funcNode, source)
		snippet := analyzer.GetNodeText(node, source)

		// Check for field_expression (object.method() or object->method())
		if funcNode.Type() == "field_expression" {
			fieldNode := funcNode.ChildByFieldName("field")
			if fieldNode != nil {
				methodName := analyzer.GetNodeText(fieldNode, source)

				// Check Qt input methods
				if sourceType, ok := a.qtInputMethods[methodName]; ok {
					flowNode := &types.FlowNode{
						ID:         analyzer.GenerateNodeID("", node),
						Type:       types.NodeSource,
						Language:   "cpp",
						Line:       int(node.StartPoint().Row) + 1,
						Column:     int(node.StartPoint().Column),
						Name:       methodName + "()",
						Snippet:    snippet,
						SourceType: sourceType,
						Metadata:   map[string]interface{}{"framework": "qt"},
					}
					sources = append(sources, flowNode)
					continue
				}

				// Check framework method inputs (Crow, Drogon, Beast, etc.)
				if sourceType, ok := a.methodInputs[methodName]; ok {
					flowNode := &types.FlowNode{
						ID:         analyzer.GenerateNodeID("", node),
						Type:       types.NodeSource,
						Language:   "cpp",
						Line:       int(node.StartPoint().Row) + 1,
						Column:     int(node.StartPoint().Column),
						Name:       methodName + "()",
						Snippet:    snippet,
						SourceType: sourceType,
					}
					sources = append(sources, flowNode)
					continue
				}
			}
		}

		// Check for getenv/secure_getenv with CGI variables
		cleanName := strings.TrimPrefix(funcName, "std::")
		if cleanName == "getenv" || cleanName == "secure_getenv" || cleanName == "qgetenv" {
			argsNode := node.ChildByFieldName("arguments")
			if argsNode != nil {
				envVar := a.extractFirstStringArg(argsNode, source)
				if cgiType, ok := a.cgiEnvVars[envVar]; ok {
					flowNode := &types.FlowNode{
						ID:         analyzer.GenerateNodeID("", node),
						Type:       types.NodeSource,
						Language:   "cpp",
						Line:       int(node.StartPoint().Row) + 1,
						Column:     int(node.StartPoint().Column),
						Name:       cleanName + "(\"" + envVar + "\")",
						Snippet:    snippet,
						SourceType: cgiType,
						SourceKey:  envVar,
					}
					sources = append(sources, flowNode)
					continue
				}
			}
		}

		// Check standard input functions
		if sourceType, ok := a.inputFunctions[cleanName]; ok {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "cpp",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       funcName,
				Snippet:    snippet,
				SourceType: sourceType,
			}
			sources = append(sources, flowNode)
		}
	}

	// Binary expressions (cin >> x, ifstream >> x)
	binaryNodes := analyzer.FindNodesOfType(root, "binary_expression")
	for _, node := range binaryNodes {
		text := analyzer.GetNodeText(node, source)
		if strings.Contains(text, ">>") {
			// std::cin
			if strings.Contains(text, "cin") {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "cpp",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       "std::cin",
					Snippet:    text,
					SourceType: types.SourceStdin,
				}
				sources = append(sources, flowNode)
			}
			// Check for file stream extraction
			leftNode := node.ChildByFieldName("left")
			if leftNode != nil {
				leftText := analyzer.GetNodeText(leftNode, source)
				// Detect ifstream/fstream extraction
				if strings.Contains(leftText, "ifstream") || strings.Contains(leftText, "fstream") {
					flowNode := &types.FlowNode{
						ID:         analyzer.GenerateNodeID("", node),
						Type:       types.NodeSource,
						Language:   "cpp",
						Line:       int(node.StartPoint().Row) + 1,
						Column:     int(node.StartPoint().Column),
						Name:       "ifstream>>",
						Snippet:    text,
						SourceType: types.SourceFile,
					}
					sources = append(sources, flowNode)
				}
			}
		}
	}

	// Subscript expressions (argv[i], envp[i])
	subscriptNodes := analyzer.FindNodesOfType(root, "subscript_expression")
	for _, node := range subscriptNodes {
		argNode := node.ChildByFieldName("argument")
		if argNode != nil {
			argName := analyzer.GetNodeText(argNode, source)
			switch argName {
			case "argv":
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "cpp",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       "argv[]",
					Snippet:    analyzer.GetNodeText(node, source),
					SourceType: types.SourceCLIArg,
				}
				sources = append(sources, flowNode)
			case "envp":
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "cpp",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       "envp[]",
					Snippet:    analyzer.GetNodeText(node, source),
					SourceType: types.SourceEnvVar,
				}
				sources = append(sources, flowNode)
			}
		}
	}

	// Declaration statements - detect framework types
	declNodes := analyzer.FindNodesOfType(root, "declaration")
	for _, node := range declNodes {
		text := analyzer.GetNodeText(node, source)
		for typeName, info := range a.frameworkTypes {
			if strings.Contains(text, typeName) {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "cpp",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       typeName,
					Snippet:    text,
					SourceType: info.SourceType,
					Metadata:   map[string]interface{}{"framework": info.Framework},
				}
				sources = append(sources, flowNode)
				break
			}
		}
	}

	// Parameter declarations - detect framework types in function parameters
	paramNodes := analyzer.FindNodesOfType(root, "parameter_declaration")
	for _, node := range paramNodes {
		typeNode := node.ChildByFieldName("type")
		if typeNode != nil {
			typeText := analyzer.GetNodeText(typeNode, source)
			for typeName, info := range a.frameworkTypes {
				if strings.Contains(typeText, typeName) {
					flowNode := &types.FlowNode{
						ID:         analyzer.GenerateNodeID("", node),
						Type:       types.NodeSource,
						Language:   "cpp",
						Line:       int(node.StartPoint().Row) + 1,
						Column:     int(node.StartPoint().Column),
						Name:       typeName + " (param)",
						Snippet:    analyzer.GetNodeText(node, source),
						SourceType: info.SourceType,
						Metadata:   map[string]interface{}{"framework": info.Framework},
					}
					sources = append(sources, flowNode)
					break
				}
			}
		}
	}

	return sources, nil
}

// extractFirstStringArg extracts the first string literal argument from an argument list
func (a *CPPAnalyzer) extractFirstStringArg(argsNode *sitter.Node, source []byte) string {
	for i := 0; i < int(argsNode.ChildCount()); i++ {
		child := argsNode.Child(i)
		if child.Type() == "string_literal" {
			text := analyzer.GetNodeText(child, source)
			// Remove quotes
			return strings.Trim(text, `"'`)
		}
	}
	return ""
}

func (a *CPPAnalyzer) DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error) {
	var frameworks []string

	for _, imp := range symbolTable.Imports {
		path := strings.ToLower(imp.Path)

		// Boost libraries
		if strings.Contains(path, "boost") {
			if !contains(frameworks, "boost") {
				frameworks = append(frameworks, "boost")
			}
			if strings.Contains(path, "beast") && !contains(frameworks, "boost.beast") {
				frameworks = append(frameworks, "boost.beast")
			}
			if strings.Contains(path, "asio") && !contains(frameworks, "boost.asio") {
				frameworks = append(frameworks, "boost.asio")
			}
		}

		// Qt framework
		if strings.Contains(path, "qt") || strings.HasPrefix(path, "q") {
			if !contains(frameworks, "qt") {
				frameworks = append(frameworks, "qt")
			}
		}

		// Crow framework
		if strings.Contains(path, "crow") {
			if !contains(frameworks, "crow") {
				frameworks = append(frameworks, "crow")
			}
		}

		// Drogon framework
		if strings.Contains(path, "drogon") {
			if !contains(frameworks, "drogon") {
				frameworks = append(frameworks, "drogon")
			}
		}

		// cpprestsdk (Casablanca)
		if strings.Contains(path, "cpprest") || strings.Contains(path, "pplx") {
			if !contains(frameworks, "cpprestsdk") {
				frameworks = append(frameworks, "cpprestsdk")
			}
		}

		// Poco framework
		if strings.Contains(path, "poco") {
			if !contains(frameworks, "poco") {
				frameworks = append(frameworks, "poco")
			}
		}

		// nlohmann/json
		if strings.Contains(path, "nlohmann") || strings.Contains(path, "json.hpp") {
			if !contains(frameworks, "nlohmann-json") {
				frameworks = append(frameworks, "nlohmann-json")
			}
		}

		// RapidJSON
		if strings.Contains(path, "rapidjson") {
			if !contains(frameworks, "rapidjson") {
				frameworks = append(frameworks, "rapidjson")
			}
		}
	}

	// Also check source for framework type usage
	sourceText := string(source)
	frameworkPatterns := map[string][]string{
		"crow":       {"crow::request", "crow::response", "CROW_ROUTE"},
		"drogon":     {"HttpRequestPtr", "HttpResponsePtr", "drogon::app"},
		"boost.beast": {"beast::http::", "websocket::stream"},
		"cpprestsdk": {"web::http::", "http_request", "http_response"},
		"poco":       {"HTTPServerRequest", "HTTPServerResponse", "Poco::Net::"},
	}

	for framework, patterns := range frameworkPatterns {
		for _, pattern := range patterns {
			if strings.Contains(sourceText, pattern) {
				if !contains(frameworks, framework) {
					frameworks = append(frameworks, framework)
				}
				break
			}
		}
	}

	return frameworks, nil
}

func (a *CPPAnalyzer) AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*analyzer.MethodFlowAnalysis, error) {
	return &analyzer.MethodFlowAnalysis{
		ParamsToReturn:     make([]int, 0),
		ParamsToProperties: make(map[int][]string),
		ParamsToCallArgs:   make(map[int][]*types.CallSite),
		TaintedVariables:   make(map[string]*types.TaintInfo),
		Assignments:        make([]*types.Assignment, 0),
		Calls:              make([]*types.CallSite, 0),
		Returns:            make([]analyzer.ReturnInfo, 0),
	}, nil
}

func (a *CPPAnalyzer) TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error) {
	flowMap := types.NewFlowMap()
	flowMap.Target = target

	expr := target.Expression

	// Check for cin stream
	if strings.Contains(expr, "cin") {
		sourceNode := types.FlowNode{
			ID:         "source-cin",
			Type:       types.NodeSource,
			Language:   "cpp",
			Name:       "std::cin",
			Snippet:    expr,
			SourceType: types.SourceStdin,
		}
		flowMap.AddSource(sourceNode)
	}

	// Check for ifstream/fstream
	if strings.Contains(expr, "ifstream") || strings.Contains(expr, "fstream") {
		sourceNode := types.FlowNode{
			ID:         "source-fstream",
			Type:       types.NodeSource,
			Language:   "cpp",
			Name:       "ifstream",
			Snippet:    expr,
			SourceType: types.SourceFile,
		}
		flowMap.AddSource(sourceNode)
	}

	// Check for input functions
	for fn, sourceType := range a.inputFunctions {
		if strings.Contains(expr, fn+"(") {
			sourceNode := types.FlowNode{
				ID:         fmt.Sprintf("source-%s", fn),
				Type:       types.NodeSource,
				Language:   "cpp",
				Name:       fn,
				Snippet:    expr,
				SourceType: sourceType,
			}
			flowMap.AddSource(sourceNode)
		}
	}

	// Check for Qt input methods
	for method, sourceType := range a.qtInputMethods {
		if strings.Contains(expr, "."+method+"(") || strings.Contains(expr, "->"+method+"(") {
			sourceNode := types.FlowNode{
				ID:         fmt.Sprintf("source-qt-%s", method),
				Type:       types.NodeSource,
				Language:   "cpp",
				Name:       "Qt:" + method,
				Snippet:    expr,
				SourceType: sourceType,
			}
			flowMap.AddSource(sourceNode)
		}
	}

	// Check for framework method inputs
	for method, sourceType := range a.methodInputs {
		if strings.Contains(expr, "."+method) || strings.Contains(expr, "->"+method) {
			sourceNode := types.FlowNode{
				ID:         fmt.Sprintf("source-method-%s", method),
				Type:       types.NodeSource,
				Language:   "cpp",
				Name:       method,
				Snippet:    expr,
				SourceType: sourceType,
			}
			flowMap.AddSource(sourceNode)
		}
	}

	// Check for framework types
	for typeName, info := range a.frameworkTypes {
		if strings.Contains(expr, typeName) {
			sourceNode := types.FlowNode{
				ID:         fmt.Sprintf("source-type-%s", typeName),
				Type:       types.NodeSource,
				Language:   "cpp",
				Name:       typeName,
				Snippet:    expr,
				SourceType: info.SourceType,
				Metadata:   map[string]interface{}{"framework": info.Framework},
			}
			flowMap.AddSource(sourceNode)
		}
	}

	// Check for CGI environment variables
	for envVar, sourceType := range a.cgiEnvVars {
		if strings.Contains(expr, `"`+envVar+`"`) || strings.Contains(expr, `'`+envVar+`'`) {
			sourceNode := types.FlowNode{
				ID:         fmt.Sprintf("source-cgi-%s", envVar),
				Type:       types.NodeSource,
				Language:   "cpp",
				Name:       "CGI:" + envVar,
				Snippet:    expr,
				SourceType: sourceType,
				SourceKey:  envVar,
			}
			flowMap.AddSource(sourceNode)
		}
	}

	// Check for argv
	if strings.Contains(expr, "argv[") {
		sourceNode := types.FlowNode{
			ID:         "source-argv",
			Type:       types.NodeSource,
			Language:   "cpp",
			Name:       "argv",
			Snippet:    expr,
			SourceType: types.SourceCLIArg,
		}
		flowMap.AddSource(sourceNode)
	}

	// Check for envp
	if strings.Contains(expr, "envp[") {
		sourceNode := types.FlowNode{
			ID:         "source-envp",
			Type:       types.NodeSource,
			Language:   "cpp",
			Name:       "envp",
			Snippet:    expr,
			SourceType: types.SourceEnvVar,
		}
		flowMap.AddSource(sourceNode)
	}

	// Check for stdin
	if strings.Contains(expr, "stdin") {
		sourceNode := types.FlowNode{
			ID:         "source-stdin",
			Type:       types.NodeSource,
			Language:   "cpp",
			Name:       "stdin",
			Snippet:    expr,
			SourceType: types.SourceStdin,
		}
		flowMap.AddSource(sourceNode)
	}

	return flowMap, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	analyzer.DefaultRegistry.Register(NewCPPAnalyzer())
}
