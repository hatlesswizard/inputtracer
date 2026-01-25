// Package golang implements the Go language analyzer for semantic input tracing
package golang

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	"github.com/hatlesswizard/inputtracer/pkg/sources"
	goPatterns "github.com/hatlesswizard/inputtracer/pkg/sources/golang"
	sitter "github.com/smacker/go-tree-sitter"
)

// GoAnalyzer implements the LanguageAnalyzer interface for Go
type GoAnalyzer struct {
	*analyzer.BaseAnalyzer
	inputSources   map[string]types.SourceType
	inputFunctions map[string]types.SourceType
}

// NewGoAnalyzer creates a new Go analyzer
func NewGoAnalyzer() *GoAnalyzer {
	m := sources.GetMappings("go")
	a := &GoAnalyzer{
		BaseAnalyzer:   analyzer.NewBaseAnalyzer("go", languages.GetExtensionsForLanguage("go")),
		inputSources:   m.GetInputSourcesMap(),
		inputFunctions: m.GetInputFunctionsMap(),
	}

	a.registerFrameworkPatterns()
	return a
}

func (a *GoAnalyzer) registerFrameworkPatterns() {
	for _, p := range goPatterns.GetAllPatterns() {
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
		}
		a.AddFrameworkPattern(fp)
	}
}

func (a *GoAnalyzer) BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error) {
	st := types.NewSymbolTable(filePath, "go")
	st.Imports = a.extractImports(root, source)

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

	frameworks, _ := a.DetectFrameworks(st, source)
	if len(frameworks) > 0 {
		st.Framework = frameworks[0]
	}

	return st, nil
}

func (a *GoAnalyzer) extractImports(root *sitter.Node, source []byte) []types.ImportInfo {
	var imports []types.ImportInfo

	importNodes := analyzer.FindNodesOfType(root, "import_declaration")
	for _, node := range importNodes {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "import_spec_list" {
				for j := 0; j < int(child.ChildCount()); j++ {
					spec := child.Child(j)
					if spec.Type() == "import_spec" {
						imp := a.parseImportSpec(spec, source)
						if imp.Path != "" {
							imports = append(imports, imp)
						}
					}
				}
			} else if child.Type() == "import_spec" {
				imp := a.parseImportSpec(child, source)
				if imp.Path != "" {
					imports = append(imports, imp)
				}
			}
		}
	}

	return imports
}

func (a *GoAnalyzer) parseImportSpec(node *sitter.Node, source []byte) types.ImportInfo {
	imp := types.ImportInfo{
		Line: int(node.StartPoint().Row) + 1,
		Type: "import",
	}

	pathNode := node.ChildByFieldName("path")
	nameNode := node.ChildByFieldName("name")

	if pathNode != nil {
		imp.Path = strings.Trim(analyzer.GetNodeText(pathNode, source), "\"")
	}

	if nameNode != nil {
		imp.Alias = analyzer.GetNodeText(nameNode, source)
	} else if imp.Path != "" {
		parts := strings.Split(imp.Path, "/")
		imp.Alias = parts[len(parts)-1]
	}

	return imp
}

func (a *GoAnalyzer) ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error) {
	return nil, nil
}

func (a *GoAnalyzer) ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error) {
	var classes []*types.ClassDef

	typeNodes := analyzer.FindNodesOfType(root, "type_declaration")
	for _, typeNode := range typeNodes {
		for i := 0; i < int(typeNode.ChildCount()); i++ {
			child := typeNode.Child(i)
			if child.Type() == "type_spec" {
				class := a.parseTypeSpec(child, source)
				if class != nil {
					classes = append(classes, class)
				}
			}
		}
	}

	return classes, nil
}

func (a *GoAnalyzer) parseTypeSpec(node *sitter.Node, source []byte) *types.ClassDef {
	nameNode := node.ChildByFieldName("name")
	typeNode := node.ChildByFieldName("type")

	if nameNode == nil || typeNode == nil {
		return nil
	}

	typeKind := typeNode.Type()
	if typeKind != "struct_type" && typeKind != "interface_type" {
		return nil
	}

	name := analyzer.GetNodeText(nameNode, source)
	class := types.NewClassDef(name, "", int(node.StartPoint().Row)+1)
	class.EndLine = int(node.EndPoint().Row) + 1

	if typeKind == "struct_type" {
		a.parseStructFields(class, typeNode, source)
	}

	return class
}

func (a *GoAnalyzer) parseStructFields(class *types.ClassDef, structNode *sitter.Node, source []byte) {
	fieldListNode := structNode.ChildByFieldName("body")
	if fieldListNode == nil {
		return
	}

	for i := 0; i < int(fieldListNode.ChildCount()); i++ {
		child := fieldListNode.Child(i)
		if child.Type() == "field_declaration" {
			typeNode := child.ChildByFieldName("type")
			typeName := ""
			if typeNode != nil {
				typeName = analyzer.GetNodeText(typeNode, source)
			}

			for j := 0; j < int(child.ChildCount()); j++ {
				nameNode := child.Child(j)
				if nameNode.Type() == "field_identifier" {
					class.Properties[analyzer.GetNodeText(nameNode, source)] = &types.PropertyDef{
						Name: analyzer.GetNodeText(nameNode, source),
						Type: typeName,
						Line: int(child.StartPoint().Row) + 1,
					}
				}
			}
		}
	}
}

func (a *GoAnalyzer) ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error) {
	var functions []*types.FunctionDef

	// Regular functions
	funcNodes := analyzer.FindNodesOfType(root, "function_declaration")
	for _, funcNode := range funcNodes {
		fn := a.parseFunctionDeclaration(funcNode, source)
		if fn != nil {
			functions = append(functions, fn)
		}
	}

	// Methods - also add as functions
	methodNodes := analyzer.FindNodesOfType(root, "method_declaration")
	for _, methodNode := range methodNodes {
		fn := a.parseMethodDeclaration(methodNode, source)
		if fn != nil {
			functions = append(functions, fn)
		}
	}

	return functions, nil
}

func (a *GoAnalyzer) parseFunctionDeclaration(node *sitter.Node, source []byte) *types.FunctionDef {
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

	paramsNode := node.ChildByFieldName("parameters")
	if paramsNode != nil {
		fn.Parameters = a.parseParameters(paramsNode, source)
	}

	resultNode := node.ChildByFieldName("result")
	if resultNode != nil {
		fn.ReturnType = analyzer.GetNodeText(resultNode, source)
	}

	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		fn.BodyStart = int(bodyNode.StartPoint().Row) + 1
		fn.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		fn.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	// Check if exported (starts with uppercase)
	if len(fn.Name) > 0 && fn.Name[0] >= 'A' && fn.Name[0] <= 'Z' {
		fn.IsExported = true
	}

	return fn
}

func (a *GoAnalyzer) parseMethodDeclaration(node *sitter.Node, source []byte) *types.FunctionDef {
	nameNode := node.ChildByFieldName("name")
	receiverNode := node.ChildByFieldName("receiver")

	if nameNode == nil {
		return nil
	}

	fn := &types.FunctionDef{
		Name:       analyzer.GetNodeText(nameNode, source),
		Line:       int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Parameters: make([]types.ParameterDef, 0),
	}

	// Extract receiver type
	if receiverNode != nil {
		for i := 0; i < int(receiverNode.ChildCount()); i++ {
			child := receiverNode.Child(i)
			if child.Type() == "parameter_declaration" {
				typeNode := child.ChildByFieldName("type")
				if typeNode != nil {
					receiverType := analyzer.GetNodeText(typeNode, source)
					receiverType = strings.TrimPrefix(receiverType, "*")
					fn.Name = receiverType + "." + fn.Name
				}
			}
		}
	}

	paramsNode := node.ChildByFieldName("parameters")
	if paramsNode != nil {
		fn.Parameters = a.parseParameters(paramsNode, source)
	}

	resultNode := node.ChildByFieldName("result")
	if resultNode != nil {
		fn.ReturnType = analyzer.GetNodeText(resultNode, source)
	}

	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		fn.BodyStart = int(bodyNode.StartPoint().Row) + 1
		fn.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		fn.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	return fn
}

func (a *GoAnalyzer) parseParameters(node *sitter.Node, source []byte) []types.ParameterDef {
	var params []types.ParameterDef
	index := 0

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameter_declaration" {
			typeNode := child.ChildByFieldName("type")
			typeName := ""
			if typeNode != nil {
				typeName = analyzer.GetNodeText(typeNode, source)
			}

			for j := 0; j < int(child.ChildCount()); j++ {
				nameNode := child.Child(j)
				if nameNode.Type() == "identifier" {
					params = append(params, types.ParameterDef{
						Name:  analyzer.GetNodeText(nameNode, source),
						Type:  typeName,
						Index: index,
					})
					index++
				}
			}
		} else if child.Type() == "variadic_parameter_declaration" {
			typeNode := child.ChildByFieldName("type")
			typeName := ""
			if typeNode != nil {
				typeName = "..." + analyzer.GetNodeText(typeNode, source)
			}

			for j := 0; j < int(child.ChildCount()); j++ {
				nameNode := child.Child(j)
				if nameNode.Type() == "identifier" {
					params = append(params, types.ParameterDef{
						Name:       analyzer.GetNodeText(nameNode, source),
						Type:       typeName,
						Index:      index,
						IsVariadic: true,
					})
					index++
				}
			}
		}
	}

	return params
}

func (a *GoAnalyzer) ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error) {
	var assignments []*types.Assignment

	assignNodes := analyzer.FindNodesOfType(root, "assignment_statement")
	assignNodes = append(assignNodes, analyzer.FindNodesOfType(root, "short_var_declaration")...)

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

func (a *GoAnalyzer) isExpressionTainted(node *sitter.Node, source []byte) (bool, string) {
	if node == nil {
		return false, ""
	}

	text := analyzer.GetNodeText(node, source)

	for src := range a.inputSources {
		if strings.Contains(text, src) {
			return true, src
		}
	}

	for fn := range a.inputFunctions {
		if strings.Contains(text, fn) {
			return true, fn
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		if tainted, src := a.isExpressionTainted(node.Child(i), source); tainted {
			return true, src
		}
	}

	return false, ""
}

func (a *GoAnalyzer) ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error) {
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

		if funcNode.Type() == "selector_expression" {
			operandNode := funcNode.ChildByFieldName("operand")
			fieldNode := funcNode.ChildByFieldName("field")
			if operandNode != nil {
				call.ClassName = analyzer.GetNodeText(operandNode, source)
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

func (a *GoAnalyzer) parseCallArguments(node *sitter.Node, source []byte) []types.CallArg {
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

func (a *GoAnalyzer) FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error) {
	var sources []*types.FlowNode

	// Selector expressions (r.Form, os.Args, etc.)
	selectorNodes := analyzer.FindNodesOfType(root, "selector_expression")
	for _, node := range selectorNodes {
		text := analyzer.GetNodeText(node, source)
		for src, sourceType := range a.inputSources {
			if strings.HasPrefix(text, src) {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "go",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       src,
					Snippet:    text,
					SourceType: sourceType,
				}
				sources = append(sources, flowNode)
			}
		}
	}

	// Call expressions (os.Getenv, c.Query, etc.)
	callNodes := analyzer.FindNodesOfType(root, "call_expression")
	for _, node := range callNodes {
		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			continue
		}
		funcName := analyzer.GetNodeText(funcNode, source)

		for fn, sourceType := range a.inputFunctions {
			if strings.Contains(funcName, fn) || strings.HasSuffix(funcName, "."+strings.Split(fn, ".")[len(strings.Split(fn, "."))-1]) {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "go",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       fn,
					Snippet:    analyzer.GetNodeText(node, source),
					SourceType: sourceType,
				}
				sources = append(sources, flowNode)
				break
			}
		}
	}

	// Index expressions (os.Args[i])
	indexNodes := analyzer.FindNodesOfType(root, "index_expression")
	for _, node := range indexNodes {
		operandNode := node.ChildByFieldName("operand")
		if operandNode != nil {
			operandName := analyzer.GetNodeText(operandNode, source)
			if strings.Contains(operandName, "os.Args") || operandName == "Args" {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "go",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       "os.Args",
					Snippet:    analyzer.GetNodeText(node, source),
					SourceType: types.SourceCLIArg,
				}
				sources = append(sources, flowNode)
			}
		}
	}

	return sources, nil
}

func (a *GoAnalyzer) DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error) {
	var frameworks []string

	for _, imp := range symbolTable.Imports {
		path := strings.ToLower(imp.Path)

		if strings.Contains(path, "gin-gonic/gin") {
			if !contains(frameworks, "gin") {
				frameworks = append(frameworks, "gin")
			}
		}
		if strings.Contains(path, "labstack/echo") {
			if !contains(frameworks, "echo") {
				frameworks = append(frameworks, "echo")
			}
		}
		if strings.Contains(path, "gofiber/fiber") {
			if !contains(frameworks, "fiber") {
				frameworks = append(frameworks, "fiber")
			}
		}
		if strings.Contains(path, "gorilla/mux") {
			if !contains(frameworks, "gorilla") {
				frameworks = append(frameworks, "gorilla")
			}
		}
	}

	return frameworks, nil
}

func (a *GoAnalyzer) AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*analyzer.MethodFlowAnalysis, error) {
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

	for i, param := range method.Parameters {
		if strings.Contains(body, "return") && strings.Contains(body, param.Name) {
			analysis.ParamsToReturn = append(analysis.ParamsToReturn, i)
		}
	}

	for src := range a.inputSources {
		if strings.Contains(body, "return") && strings.Contains(body, src) {
			analysis.ReturnsInput = true
			break
		}
	}

	return analysis, nil
}

func (a *GoAnalyzer) TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error) {
	flowMap := types.NewFlowMap()
	flowMap.Target = target

	expr := target.Expression

	for _, pattern := range a.GetFrameworkPatterns() {
		if a.matchesFrameworkPattern(expr, pattern) {
			flowMap.CarrierChain = &types.CarrierChain{
				ClassName:        pattern.CarrierClass,
				PropertyName:     pattern.CarrierProperty,
				PopulationMethod: pattern.PopulatedBy,
				PopulationCalls:  pattern.PopulatedFrom,
				Framework:        pattern.Framework,
			}

			for _, src := range pattern.PopulatedFrom {
				sourceNode := types.FlowNode{
					ID:         fmt.Sprintf("source-%s", src),
					Type:       types.NodeSource,
					Language:   "go",
					Name:       src,
					Snippet:    src,
					SourceType: pattern.SourceType,
				}
				flowMap.AddSource(sourceNode)
			}

			break
		}
	}

	if len(flowMap.Sources) == 0 {
		for src, sourceType := range a.inputSources {
			if strings.Contains(expr, src) {
				sourceNode := types.FlowNode{
					ID:         fmt.Sprintf("source-%s", src),
					Type:       types.NodeSource,
					Language:   "go",
					Name:       src,
					Snippet:    expr,
					SourceType: sourceType,
				}
				flowMap.AddSource(sourceNode)
			}
		}
	}

	return flowMap, nil
}

func (a *GoAnalyzer) matchesFrameworkPattern(expr string, pattern *types.FrameworkPattern) bool {
	if pattern.MethodPattern != "" {
		regex := regexp.MustCompile(`\.` + pattern.MethodPattern + `\(`)
		return regex.MatchString(expr)
	}
	return false
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
	analyzer.DefaultRegistry.Register(NewGoAnalyzer())
}
