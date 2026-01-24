// Package csharp implements the C# language analyzer for semantic input tracing
package csharp

import (
	"fmt"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/mappings"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// CSharpAnalyzer implements the LanguageAnalyzer interface for C#
type CSharpAnalyzer struct {
	*analyzer.BaseAnalyzer
	inputSources map[string]types.SourceType
}

// NewCSharpAnalyzer creates a new C# analyzer
func NewCSharpAnalyzer() *CSharpAnalyzer {
	m := mappings.GetMappings("c_sharp")
	a := &CSharpAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer("c_sharp", languages.GetExtensionsForLanguage("c_sharp")),
		inputSources: m.GetInputSourcesMap(),
	}

	return a
}

func (a *CSharpAnalyzer) BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error) {
	st := types.NewSymbolTable(filePath, "c_sharp")
	st.Imports = a.extractUsings(root, source)

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

func (a *CSharpAnalyzer) extractUsings(root *sitter.Node, source []byte) []types.ImportInfo {
	var imports []types.ImportInfo

	usingNodes := analyzer.FindNodesOfType(root, "using_directive")
	for _, node := range usingNodes {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "qualified_name" || child.Type() == "identifier" {
				nsName := analyzer.GetNodeText(child, source)
				imports = append(imports, types.ImportInfo{
					Path: nsName,
					Line: int(node.StartPoint().Row) + 1,
					Type: "using",
				})
			}
		}
	}

	return imports
}

func (a *CSharpAnalyzer) ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error) {
	return nil, nil
}

func (a *CSharpAnalyzer) ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error) {
	var classes []*types.ClassDef

	classNodes := analyzer.FindNodesOfType(root, "class_declaration")
	for _, classNode := range classNodes {
		nameNode := classNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		name := analyzer.GetNodeText(nameNode, source)
		class := types.NewClassDef(name, "", int(classNode.StartPoint().Row)+1)
		class.EndLine = int(classNode.EndPoint().Row) + 1

		// Extract base class
		basesNode := classNode.ChildByFieldName("bases")
		if basesNode != nil {
			for i := 0; i < int(basesNode.ChildCount()); i++ {
				child := basesNode.Child(i)
				if child.Type() == "identifier" || child.Type() == "generic_name" {
					class.Extends = analyzer.GetNodeText(child, source)
					break
				}
			}
		}

		// Extract methods
		bodyNode := classNode.ChildByFieldName("body")
		if bodyNode != nil {
			methods := a.extractMethodsFromClass(bodyNode, source, name)
			for _, method := range methods {
				class.Methods[method.Name] = method
			}
		}

		classes = append(classes, class)
	}

	// Also extract interfaces and structs as classes
	interfaceNodes := analyzer.FindNodesOfType(root, "interface_declaration")
	for _, ifNode := range interfaceNodes {
		nameNode := ifNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}
		name := analyzer.GetNodeText(nameNode, source)
		class := types.NewClassDef(name, "", int(ifNode.StartPoint().Row)+1)
		class.EndLine = int(ifNode.EndPoint().Row) + 1
		classes = append(classes, class)
	}

	structNodes := analyzer.FindNodesOfType(root, "struct_declaration")
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

func (a *CSharpAnalyzer) extractMethodsFromClass(bodyNode *sitter.Node, source []byte, className string) []*types.MethodDef {
	var methods []*types.MethodDef

	methodNodes := analyzer.FindNodesOfType(bodyNode, "method_declaration")
	for _, methodNode := range methodNodes {
		nameNode := methodNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		method := &types.MethodDef{
			Name:    analyzer.GetNodeText(nameNode, source),
			Line:    int(methodNode.StartPoint().Row) + 1,
			EndLine: int(methodNode.EndPoint().Row) + 1,
		}

		typeNode := methodNode.ChildByFieldName("type")
		if typeNode != nil {
			method.ReturnType = analyzer.GetNodeText(typeNode, source)
		}

		paramsNode := methodNode.ChildByFieldName("parameters")
		if paramsNode != nil {
			method.Parameters = a.parseParameters(paramsNode, source)
		}

		bodyNode := methodNode.ChildByFieldName("body")
		if bodyNode != nil {
			method.BodyStart = int(bodyNode.StartPoint().Row) + 1
			method.BodyEnd = int(bodyNode.EndPoint().Row) + 1
			method.BodySource = analyzer.GetNodeText(bodyNode, source)
		}

		methods = append(methods, method)
	}

	// Also get constructors
	ctorNodes := analyzer.FindNodesOfType(bodyNode, "constructor_declaration")
	for _, ctorNode := range ctorNodes {
		nameNode := ctorNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		method := &types.MethodDef{
			Name:    analyzer.GetNodeText(nameNode, source),
			Line:    int(ctorNode.StartPoint().Row) + 1,
			EndLine: int(ctorNode.EndPoint().Row) + 1,
		}

		paramsNode := ctorNode.ChildByFieldName("parameters")
		if paramsNode != nil {
			method.Parameters = a.parseParameters(paramsNode, source)
		}

		bodyNode := ctorNode.ChildByFieldName("body")
		if bodyNode != nil {
			method.BodyStart = int(bodyNode.StartPoint().Row) + 1
			method.BodyEnd = int(bodyNode.EndPoint().Row) + 1
			method.BodySource = analyzer.GetNodeText(bodyNode, source)
		}

		methods = append(methods, method)
	}

	return methods
}

func (a *CSharpAnalyzer) ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error) {
	var functions []*types.FunctionDef

	// In C#, top-level functions are rare but can exist in newer versions
	// Most methods are inside classes, handled by ExtractClasses
	localFuncNodes := analyzer.FindNodesOfType(root, "local_function_statement")
	for _, funcNode := range localFuncNodes {
		nameNode := funcNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		fn := &types.FunctionDef{
			Name:    analyzer.GetNodeText(nameNode, source),
			Line:    int(funcNode.StartPoint().Row) + 1,
			EndLine: int(funcNode.EndPoint().Row) + 1,
		}

		typeNode := funcNode.ChildByFieldName("type")
		if typeNode != nil {
			fn.ReturnType = analyzer.GetNodeText(typeNode, source)
		}

		paramsNode := funcNode.ChildByFieldName("parameters")
		if paramsNode != nil {
			fn.Parameters = a.parseParameters(paramsNode, source)
		}

		bodyNode := funcNode.ChildByFieldName("body")
		if bodyNode != nil {
			fn.BodyStart = int(bodyNode.StartPoint().Row) + 1
			fn.BodyEnd = int(bodyNode.EndPoint().Row) + 1
			fn.BodySource = analyzer.GetNodeText(bodyNode, source)
		}

		functions = append(functions, fn)
	}

	return functions, nil
}

func (a *CSharpAnalyzer) parseParameters(node *sitter.Node, source []byte) []types.ParameterDef {
	var params []types.ParameterDef
	index := 0

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameter" {
			param := types.ParameterDef{Index: index}

			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				param.Name = analyzer.GetNodeText(nameNode, source)
			}

			typeNode := child.ChildByFieldName("type")
			if typeNode != nil {
				param.Type = analyzer.GetNodeText(typeNode, source)
			}

			defaultNode := child.ChildByFieldName("default")
			if defaultNode != nil {
				param.DefaultValue = analyzer.GetNodeText(defaultNode, source)
			}

			// Check for params keyword (variadic)
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				if c.Type() == "modifier" && analyzer.GetNodeText(c, source) == "params" {
					param.IsVariadic = true
				}
			}

			if param.Name != "" {
				params = append(params, param)
				index++
			}
		}
	}

	return params
}

func (a *CSharpAnalyzer) ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error) {
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

	// Handle variable declarators
	declNodes := analyzer.FindNodesOfType(root, "variable_declarator")
	for _, node := range declNodes {
		nameNode := node.ChildByFieldName("name")
		valueNode := node.ChildByFieldName("value")
		if nameNode != nil && valueNode != nil {
			assignment := &types.Assignment{
				Target:   analyzer.GetNodeText(nameNode, source),
				Source:   analyzer.GetNodeText(valueNode, source),
				Line:     int(node.StartPoint().Row) + 1,
				Column:   int(node.StartPoint().Column),
				Scope:    scope,
			}
			assignment.IsTainted, assignment.TaintSource = a.isExpressionTainted(valueNode, source)
			assignments = append(assignments, assignment)
		}
	}

	return assignments, nil
}

func (a *CSharpAnalyzer) isExpressionTainted(node *sitter.Node, source []byte) (bool, string) {
	if node == nil {
		return false, ""
	}

	text := analyzer.GetNodeText(node, source)

	for fn := range a.inputSources {
		if strings.Contains(text, fn) {
			return true, fn
		}
	}

	// Check for args[] access
	if strings.Contains(text, "args[") {
		return true, "args[]"
	}

	// Check for attribute patterns
	if strings.Contains(text, "[FromQuery]") || strings.Contains(text, "[FromBody]") ||
		strings.Contains(text, "[FromForm]") || strings.Contains(text, "[FromHeader]") {
		return true, "attribute binding"
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		if tainted, src := a.isExpressionTainted(node.Child(i), source); tainted {
			return true, src
		}
	}

	return false, ""
}

func (a *CSharpAnalyzer) ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error) {
	var calls []*types.CallSite

	callNodes := analyzer.FindNodesOfType(root, "invocation_expression")
	for _, node := range callNodes {
		exprNode := node.ChildByFieldName("expression")
		if exprNode == nil {
			continue
		}

		call := &types.CallSite{
			FunctionName: analyzer.GetNodeText(exprNode, source),
			Line:         int(node.StartPoint().Row) + 1,
			Column:       int(node.StartPoint().Column),
			Scope:        scope,
			Arguments:    make([]types.CallArg, 0),
		}

		// Handle method calls
		if exprNode.Type() == "member_access_expression" {
			objNode := exprNode.ChildByFieldName("expression")
			nameNode := exprNode.ChildByFieldName("name")
			if objNode != nil && nameNode != nil {
				call.ClassName = analyzer.GetNodeText(objNode, source)
				call.MethodName = analyzer.GetNodeText(nameNode, source)
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

func (a *CSharpAnalyzer) parseCallArguments(node *sitter.Node, source []byte) []types.CallArg {
	var args []types.CallArg
	index := 0

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "argument" {
			arg := types.CallArg{
				Index: index,
				Value: analyzer.GetNodeText(child, source),
			}
			arg.IsTainted, arg.TaintSource = a.isExpressionTainted(child, source)
			args = append(args, arg)
			index++
		}
	}

	return args
}

func (a *CSharpAnalyzer) FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error) {
	var sources []*types.FlowNode

	// Check invocation expressions
	callNodes := analyzer.FindNodesOfType(root, "invocation_expression")
	for _, node := range callNodes {
		exprNode := node.ChildByFieldName("expression")
		if exprNode == nil {
			continue
		}

		exprText := analyzer.GetNodeText(exprNode, source)

		for fn, sourceType := range a.inputSources {
			if strings.Contains(exprText, fn) {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "c_sharp",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       fn,
					Snippet:    analyzer.GetNodeText(node, source),
					SourceType: sourceType,
				}
				sources = append(sources, flowNode)
			}
		}
	}

	// Check element access (args[], Request.Query[], etc.)
	elementNodes := analyzer.FindNodesOfType(root, "element_access_expression")
	for _, node := range elementNodes {
		exprNode := node.ChildByFieldName("expression")
		if exprNode != nil {
			exprName := analyzer.GetNodeText(exprNode, source)
			var sourceType types.SourceType
			var name string

			switch {
			case exprName == "args":
				sourceType = types.SourceCLIArg
				name = "args[]"
			case strings.Contains(exprName, "Request.Query"):
				sourceType = types.SourceHTTPGet
				name = "Request.Query[]"
			case strings.Contains(exprName, "Request.Form"):
				sourceType = types.SourceHTTPPost
				name = "Request.Form[]"
			case strings.Contains(exprName, "Request.Headers"):
				sourceType = types.SourceHTTPHeader
				name = "Request.Headers[]"
			case strings.Contains(exprName, "Request.Cookies"):
				sourceType = types.SourceHTTPCookie
				name = "Request.Cookies[]"
			default:
				continue
			}

			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "c_sharp",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       name,
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: sourceType,
			}
			sources = append(sources, flowNode)
		}
	}

	// Check attributes ([FromQuery], [FromBody], etc.)
	sources = append(sources, a.findASPNetAttributes(root, source)...)

	return sources, nil
}

// aspNetAttributes maps ASP.NET attribute names to their source types
var aspNetAttributes = map[string]types.SourceType{
	"FromQuery":   types.SourceHTTPGet,
	"FromBody":    types.SourceHTTPBody,
	"FromForm":    types.SourceHTTPPost,
	"FromHeader":  types.SourceHTTPHeader,
	"FromRoute":   types.SourceHTTPGet,
	"FromServices": types.SourceUserInput,
}

// findASPNetAttributes finds ASP.NET parameter attributes like [FromQuery], [FromBody], etc.
func (a *CSharpAnalyzer) findASPNetAttributes(root *sitter.Node, source []byte) []*types.FlowNode {
	var sources []*types.FlowNode

	attrNodes := analyzer.FindNodesOfType(root, "attribute")
	for _, node := range attrNodes {
		// Get the attribute name (identifier child)
		var attrName string
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "identifier" {
				attrName = analyzer.GetNodeText(child, source)
				break
			}
		}

		if attrName == "" {
			continue
		}

		// Check if this is an ASP.NET input attribute
		sourceType, isInputAttribute := aspNetAttributes[attrName]
		if !isInputAttribute {
			continue
		}

		// Only detect attributes on parameters, not on methods/classes
		paramName, isOnParameter := a.getAttributedParameterInfo(node, source)
		if !isOnParameter {
			continue // Skip attributes that are not on parameters
		}

		flowNode := &types.FlowNode{
			ID:         analyzer.GenerateNodeID("", node),
			Type:       types.NodeSource,
			Language:   "c_sharp",
			Line:       int(node.StartPoint().Row) + 1,
			Column:     int(node.StartPoint().Column),
			Name:       "[" + attrName + "] " + paramName,
			Snippet:    analyzer.GetNodeText(node, source),
			SourceType: sourceType,
		}
		sources = append(sources, flowNode)
	}

	return sources
}

// getAttributedParameterInfo extracts the parameter name from an attribute's parent parameter
// Returns the parameter name and true if the attribute is on a parameter, empty string and false otherwise
func (a *CSharpAnalyzer) getAttributedParameterInfo(attributeNode *sitter.Node, source []byte) (string, bool) {
	// Walk up to find the parameter node
	// Attribute is inside: parameter -> attribute_list -> attribute
	parent := attributeNode.Parent()
	for parent != nil {
		if parent.Type() == "parameter" {
			// Get the parameter name (the last identifier child that's not the type)
			// Parameter structure: attribute_list, type (predefined_type or identifier), identifier (name)
			var lastIdentifier string
			for i := 0; i < int(parent.ChildCount()); i++ {
				child := parent.Child(i)
				if child.Type() == "identifier" {
					lastIdentifier = analyzer.GetNodeText(child, source)
				}
			}
			if lastIdentifier != "" {
				return lastIdentifier, true
			}
			return "", true // On a parameter but couldn't extract name
		}
		// Stop if we hit a method or class declaration (attribute is not on a parameter)
		if parent.Type() == "method_declaration" || parent.Type() == "class_declaration" ||
			parent.Type() == "constructor_declaration" {
			return "", false
		}
		parent = parent.Parent()
	}
	return "", false
}

func (a *CSharpAnalyzer) DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error) {
	var frameworks []string

	for _, imp := range symbolTable.Imports {
		switch {
		case strings.Contains(imp.Path, "Microsoft.AspNetCore"):
			frameworks = append(frameworks, "ASP.NET Core")
		case strings.Contains(imp.Path, "System.Web.Mvc"):
			frameworks = append(frameworks, "ASP.NET MVC")
		case strings.Contains(imp.Path, "Nancy"):
			frameworks = append(frameworks, "Nancy")
		case strings.Contains(imp.Path, "ServiceStack"):
			frameworks = append(frameworks, "ServiceStack")
		}
	}

	return frameworks, nil
}

func (a *CSharpAnalyzer) AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*analyzer.MethodFlowAnalysis, error) {
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

func (a *CSharpAnalyzer) TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error) {
	flowMap := types.NewFlowMap()
	flowMap.Target = target

	expr := target.Expression

	for fn, sourceType := range a.inputSources {
		if strings.Contains(expr, fn) {
			sourceNode := types.FlowNode{
				ID:         fmt.Sprintf("source-%s", fn),
				Type:       types.NodeSource,
				Language:   "c_sharp",
				Name:       fn,
				Snippet:    expr,
				SourceType: sourceType,
			}
			flowMap.AddSource(sourceNode)
		}
	}

	if strings.Contains(expr, "args[") {
		sourceNode := types.FlowNode{
			ID:         "source-args",
			Type:       types.NodeSource,
			Language:   "c_sharp",
			Name:       "args",
			Snippet:    expr,
			SourceType: types.SourceCLIArg,
		}
		flowMap.AddSource(sourceNode)
	}

	return flowMap, nil
}

func init() {
	analyzer.DefaultRegistry.Register(NewCSharpAnalyzer())
}
