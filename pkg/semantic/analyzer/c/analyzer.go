// Package c implements the C language analyzer for semantic input tracing
package c

import (
	"fmt"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	"github.com/hatlesswizard/inputtracer/pkg/sources"
	sitter "github.com/smacker/go-tree-sitter"
)

// CAnalyzer implements the LanguageAnalyzer interface for C
type CAnalyzer struct {
	*analyzer.BaseAnalyzer
	inputFunctions map[string]types.SourceType
	cgiEnvVars     map[string]types.SourceType
}

// NewCAnalyzer creates a new C analyzer
func NewCAnalyzer() *CAnalyzer {
	m := sources.GetMappings("c")
	a := &CAnalyzer{
		BaseAnalyzer:   analyzer.NewBaseAnalyzer("c", languages.GetExtensionsForLanguage("c")),
		inputFunctions: m.GetInputFunctionsMap(),
		cgiEnvVars:     m.GetCGIEnvVarsMap(),
	}

	return a
}

func (a *CAnalyzer) BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error) {
	st := types.NewSymbolTable(filePath, "c")
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

func (a *CAnalyzer) extractIncludes(root *sitter.Node, source []byte) []types.ImportInfo {
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

	return imports
}

func (a *CAnalyzer) ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error) {
	return nil, nil
}

func (a *CAnalyzer) ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error) {
	var classes []*types.ClassDef

	structNodes := analyzer.FindNodesOfType(root, "struct_specifier")
	for _, structNode := range structNodes {
		nameNode := structNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		name := analyzer.GetNodeText(nameNode, source)
		class := types.NewClassDef(name, "", int(structNode.StartPoint().Row)+1)
		class.EndLine = int(structNode.EndPoint().Row) + 1

		classes = append(classes, class)
	}

	return classes, nil
}

func (a *CAnalyzer) ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error) {
	var functions []*types.FunctionDef

	funcNodes := analyzer.FindNodesOfType(root, "function_definition")
	for _, funcNode := range funcNodes {
		fn := a.parseFunctionDefinition(funcNode, source)
		if fn != nil {
			functions = append(functions, fn)
		}
	}

	return functions, nil
}

func (a *CAnalyzer) parseFunctionDefinition(node *sitter.Node, source []byte) *types.FunctionDef {
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

func (a *CAnalyzer) findFunctionName(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}

	if node.Type() == "identifier" {
		return analyzer.GetNodeText(node, source)
	}

	if node.Type() == "function_declarator" {
		declarator := node.ChildByFieldName("declarator")
		if declarator != nil {
			return a.findFunctionName(declarator, source)
		}
	}

	if node.Type() == "pointer_declarator" {
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

func (a *CAnalyzer) findParameterList(node *sitter.Node) *sitter.Node {
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

func (a *CAnalyzer) parseParameters(node *sitter.Node, source []byte) []types.ParameterDef {
	var params []types.ParameterDef
	index := 0

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameter_declaration" {
			typeNode := child.ChildByFieldName("type")
			declaratorNode := child.ChildByFieldName("declarator")

			param := types.ParameterDef{Index: index}

			if typeNode != nil {
				param.Type = analyzer.GetNodeText(typeNode, source)
			}

			if declaratorNode != nil {
				param.Name = a.findParameterName(declaratorNode, source)
			}

			if param.Name != "" {
				params = append(params, param)
				index++
			}
		} else if child.Type() == "variadic_parameter" {
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

func (a *CAnalyzer) findParameterName(node *sitter.Node, source []byte) string {
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

func (a *CAnalyzer) ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error) {
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

func (a *CAnalyzer) isExpressionTainted(node *sitter.Node, source []byte) (bool, string) {
	if node == nil {
		return false, ""
	}

	text := analyzer.GetNodeText(node, source)

	// Check for input functions
	for fn := range a.inputFunctions {
		if strings.Contains(text, fn+"(") {
			return true, fn + "()"
		}
	}

	// Check for CGI environment variables
	for envVar := range a.cgiEnvVars {
		if strings.Contains(text, `"`+envVar+`"`) || strings.Contains(text, `'`+envVar+`'`) {
			return true, "CGI:" + envVar
		}
	}

	// Check for argv, stdin, envp, environ
	if strings.Contains(text, "argv[") {
		return true, "CLI:argv"
	}
	if strings.Contains(text, "envp[") || strings.Contains(text, "environ") {
		return true, "ENV:environ"
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

func (a *CAnalyzer) ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error) {
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

func (a *CAnalyzer) parseCallArguments(node *sitter.Node, source []byte) []types.CallArg {
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

func (a *CAnalyzer) FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error) {
	var sources []*types.FlowNode

	// Call expressions (gets, scanf, etc.)
	callNodes := analyzer.FindNodesOfType(root, "call_expression")
	for _, node := range callNodes {
		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			continue
		}
		funcName := analyzer.GetNodeText(funcNode, source)

		// Special handling for getenv/secure_getenv with CGI variables
		if funcName == "getenv" || funcName == "secure_getenv" {
			argsNode := node.ChildByFieldName("arguments")
			if argsNode != nil {
				envVar := a.extractFirstStringArg(argsNode, source)
				if cgiType, ok := a.cgiEnvVars[envVar]; ok {
					flowNode := &types.FlowNode{
						ID:         analyzer.GenerateNodeID("", node),
						Type:       types.NodeSource,
						Language:   "c",
						Line:       int(node.StartPoint().Row) + 1,
						Column:     int(node.StartPoint().Column),
						Name:       funcName + "(\"" + envVar + "\")",
						Snippet:    analyzer.GetNodeText(node, source),
						SourceType: cgiType,
						SourceKey:  envVar,
					}
					sources = append(sources, flowNode)
					continue
				}
			}
		}

		if sourceType, ok := a.inputFunctions[funcName]; ok {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "c",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       funcName,
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: sourceType,
			}
			sources = append(sources, flowNode)
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
					Language:   "c",
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
					Language:   "c",
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

	// Identifiers (stdin, argc, environ)
	identNodes := analyzer.FindNodesOfType(root, "identifier")
	for _, node := range identNodes {
		name := analyzer.GetNodeText(node, source)
		switch name {
		case "stdin":
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "c",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "stdin",
				Snippet:    name,
				SourceType: types.SourceStdin,
			}
			sources = append(sources, flowNode)
		case "argc":
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "c",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "argc",
				Snippet:    name,
				SourceType: types.SourceCLIArg,
			}
			sources = append(sources, flowNode)
		case "environ":
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "c",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "environ",
				Snippet:    name,
				SourceType: types.SourceEnvVar,
			}
			sources = append(sources, flowNode)
		}
	}

	return sources, nil
}

// extractFirstStringArg extracts the first string literal argument from an argument list
func (a *CAnalyzer) extractFirstStringArg(argsNode *sitter.Node, source []byte) string {
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

func (a *CAnalyzer) DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error) {
	return nil, nil
}

func (a *CAnalyzer) AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*analyzer.MethodFlowAnalysis, error) {
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

func (a *CAnalyzer) TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error) {
	flowMap := types.NewFlowMap()
	flowMap.Target = target

	expr := target.Expression

	// Check for input functions
	for fn, sourceType := range a.inputFunctions {
		if strings.Contains(expr, fn+"(") {
			sourceNode := types.FlowNode{
				ID:         fmt.Sprintf("source-%s", fn),
				Type:       types.NodeSource,
				Language:   "c",
				Name:       fn,
				Snippet:    expr,
				SourceType: sourceType,
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
				Language:   "c",
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
			Language:   "c",
			Name:       "argv",
			Snippet:    expr,
			SourceType: types.SourceCLIArg,
		}
		flowMap.AddSource(sourceNode)
	}

	// Check for envp/environ
	if strings.Contains(expr, "envp[") || strings.Contains(expr, "environ") {
		sourceNode := types.FlowNode{
			ID:         "source-environ",
			Type:       types.NodeSource,
			Language:   "c",
			Name:       "environ",
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
			Language:   "c",
			Name:       "stdin",
			Snippet:    expr,
			SourceType: types.SourceStdin,
		}
		flowMap.AddSource(sourceNode)
	}

	return flowMap, nil
}

func init() {
	analyzer.DefaultRegistry.Register(NewCAnalyzer())
}
