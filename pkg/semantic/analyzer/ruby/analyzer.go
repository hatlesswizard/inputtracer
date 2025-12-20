// Package ruby implements the Ruby language analyzer for semantic input tracing
package ruby

import (
	"fmt"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// RubyAnalyzer implements the LanguageAnalyzer interface for Ruby
type RubyAnalyzer struct {
	*analyzer.BaseAnalyzer
	inputSources map[string]types.SourceType
}

// NewRubyAnalyzer creates a new Ruby analyzer
func NewRubyAnalyzer() *RubyAnalyzer {
	a := &RubyAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer("ruby", []string{".rb", ".rake", ".gemspec"}),
	}

	a.inputSources = map[string]types.SourceType{
		"gets":      types.SourceStdin,
		"readline":  types.SourceStdin,
		"readlines": types.SourceStdin,
		"STDIN":     types.SourceStdin,
		"ARGF":      types.SourceStdin,
		"ARGV":      types.SourceCLIArg,
		"ENV":       types.SourceEnvVar,
		"params":    types.SourceHTTPGet,
		"request":   types.SourceUserInput,
		"cookies":   types.SourceHTTPCookie,
		"session":   types.SourceUserInput,
		"File.read": types.SourceFile,
		"IO.read":   types.SourceFile,
	}

	return a
}

func (a *RubyAnalyzer) BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error) {
	st := types.NewSymbolTable(filePath, "ruby")
	st.Imports = a.extractRequires(root, source)

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

func (a *RubyAnalyzer) extractRequires(root *sitter.Node, source []byte) []types.ImportInfo {
	var imports []types.ImportInfo

	callNodes := analyzer.FindNodesOfType(root, "call")
	for _, node := range callNodes {
		methodNode := node.ChildByFieldName("method")
		if methodNode == nil {
			continue
		}

		methodName := analyzer.GetNodeText(methodNode, source)
		if methodName != "require" && methodName != "require_relative" {
			continue
		}

		argsNode := node.ChildByFieldName("arguments")
		if argsNode == nil {
			continue
		}

		for i := 0; i < int(argsNode.ChildCount()); i++ {
			child := argsNode.Child(i)
			if child.Type() == "string" || child.Type() == "simple_string" {
				path := strings.Trim(analyzer.GetNodeText(child, source), `"'`)
				imports = append(imports, types.ImportInfo{
					Path: path,
					Line: int(node.StartPoint().Row) + 1,
					Type: methodName,
				})
			}
		}
	}

	return imports
}

func (a *RubyAnalyzer) ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error) {
	return nil, nil
}

func (a *RubyAnalyzer) ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error) {
	var classes []*types.ClassDef

	classNodes := analyzer.FindNodesOfType(root, "class")
	for _, classNode := range classNodes {
		nameNode := classNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		name := analyzer.GetNodeText(nameNode, source)
		class := types.NewClassDef(name, "", int(classNode.StartPoint().Row)+1)
		class.EndLine = int(classNode.EndPoint().Row) + 1

		// Extract superclass
		superNode := classNode.ChildByFieldName("superclass")
		if superNode != nil {
			class.Extends = analyzer.GetNodeText(superNode, source)
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

	// Also extract modules
	moduleNodes := analyzer.FindNodesOfType(root, "module")
	for _, moduleNode := range moduleNodes {
		nameNode := moduleNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		name := analyzer.GetNodeText(nameNode, source)
		class := types.NewClassDef(name, "", int(moduleNode.StartPoint().Row)+1)
		class.EndLine = int(moduleNode.EndPoint().Row) + 1

		bodyNode := moduleNode.ChildByFieldName("body")
		if bodyNode != nil {
			methods := a.extractMethodsFromClass(bodyNode, source, name)
			for _, method := range methods {
				class.Methods[method.Name] = method
			}
		}

		classes = append(classes, class)
	}

	return classes, nil
}

func (a *RubyAnalyzer) extractMethodsFromClass(bodyNode *sitter.Node, source []byte, className string) []*types.MethodDef {
	var methods []*types.MethodDef

	methodNodes := analyzer.FindNodesOfType(bodyNode, "method")
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

	// Also get singleton methods (class methods)
	singletonNodes := analyzer.FindNodesOfType(bodyNode, "singleton_method")
	for _, methodNode := range singletonNodes {
		nameNode := methodNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		method := &types.MethodDef{
			Name:     "self." + analyzer.GetNodeText(nameNode, source),
			Line:     int(methodNode.StartPoint().Row) + 1,
			EndLine:  int(methodNode.EndPoint().Row) + 1,
			IsStatic: true,
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

	return methods
}

func (a *RubyAnalyzer) ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error) {
	var functions []*types.FunctionDef

	// Top-level methods (not inside classes)
	methodNodes := analyzer.FindNodesOfType(root, "method")
	for _, methodNode := range methodNodes {
		// Check if this method is inside a class
		parent := methodNode.Parent()
		insideClass := false
		for parent != nil {
			if parent.Type() == "class" || parent.Type() == "module" {
				insideClass = true
				break
			}
			parent = parent.Parent()
		}

		if insideClass {
			continue
		}

		nameNode := methodNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		fn := &types.FunctionDef{
			Name:    analyzer.GetNodeText(nameNode, source),
			Line:    int(methodNode.StartPoint().Row) + 1,
			EndLine: int(methodNode.EndPoint().Row) + 1,
		}

		paramsNode := methodNode.ChildByFieldName("parameters")
		if paramsNode != nil {
			fn.Parameters = a.parseParameters(paramsNode, source)
		}

		bodyNode := methodNode.ChildByFieldName("body")
		if bodyNode != nil {
			fn.BodyStart = int(bodyNode.StartPoint().Row) + 1
			fn.BodyEnd = int(bodyNode.EndPoint().Row) + 1
			fn.BodySource = analyzer.GetNodeText(bodyNode, source)
		}

		functions = append(functions, fn)
	}

	return functions, nil
}

func (a *RubyAnalyzer) parseParameters(node *sitter.Node, source []byte) []types.ParameterDef {
	var params []types.ParameterDef
	index := 0

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()

		var param types.ParameterDef
		param.Index = index

		switch childType {
		case "identifier":
			param.Name = analyzer.GetNodeText(child, source)
		case "optional_parameter":
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				param.Name = analyzer.GetNodeText(nameNode, source)
				valueNode := child.ChildByFieldName("value")
				if valueNode != nil {
					param.DefaultValue = analyzer.GetNodeText(valueNode, source)
				}
			}
		case "splat_parameter", "hash_splat_parameter":
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				param.Name = analyzer.GetNodeText(nameNode, source)
				param.IsVariadic = true
			}
		case "block_parameter":
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				param.Name = "&" + analyzer.GetNodeText(nameNode, source)
			}
		case "keyword_parameter":
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				param.Name = analyzer.GetNodeText(nameNode, source) + ":"
				valueNode := child.ChildByFieldName("value")
				if valueNode != nil {
					param.DefaultValue = analyzer.GetNodeText(valueNode, source)
				}
			}
		default:
			continue
		}

		if param.Name != "" {
			params = append(params, param)
			index++
		}
	}

	return params
}

func (a *RubyAnalyzer) ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error) {
	var assignments []*types.Assignment

	assignNodes := analyzer.FindNodesOfType(root, "assignment")
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

func (a *RubyAnalyzer) isExpressionTainted(node *sitter.Node, source []byte) (bool, string) {
	if node == nil {
		return false, ""
	}

	text := analyzer.GetNodeText(node, source)

	for fn := range a.inputSources {
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

func (a *RubyAnalyzer) ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error) {
	var calls []*types.CallSite

	callNodes := analyzer.FindNodesOfType(root, "call")
	for _, node := range callNodes {
		methodNode := node.ChildByFieldName("method")
		if methodNode == nil {
			continue
		}

		call := &types.CallSite{
			FunctionName: analyzer.GetNodeText(methodNode, source),
			Line:         int(node.StartPoint().Row) + 1,
			Column:       int(node.StartPoint().Column),
			Scope:        scope,
			Arguments:    make([]types.CallArg, 0),
		}

		receiverNode := node.ChildByFieldName("receiver")
		if receiverNode != nil {
			call.ClassName = analyzer.GetNodeText(receiverNode, source)
		}
		call.MethodName = analyzer.GetNodeText(methodNode, source)

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

func (a *RubyAnalyzer) parseCallArguments(node *sitter.Node, source []byte) []types.CallArg {
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

func (a *RubyAnalyzer) FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error) {
	var sources []*types.FlowNode

	// Check for method calls
	callNodes := analyzer.FindNodesOfType(root, "call")
	for _, node := range callNodes {
		methodNode := node.ChildByFieldName("method")
		receiverNode := node.ChildByFieldName("receiver")

		if methodNode != nil {
			methodName := analyzer.GetNodeText(methodNode, source)
			receiverName := ""
			if receiverNode != nil {
				receiverName = analyzer.GetNodeText(receiverNode, source)
			}

			// Check for known input sources
			if sourceType, ok := a.inputSources[methodName]; ok {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "ruby",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       methodName,
					Snippet:    analyzer.GetNodeText(node, source),
					SourceType: sourceType,
				}
				sources = append(sources, flowNode)
			}

			// Check receiver for params, request, etc.
			if sourceType, ok := a.inputSources[receiverName]; ok {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "ruby",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       receiverName + "." + methodName,
					Snippet:    analyzer.GetNodeText(node, source),
					SourceType: sourceType,
				}
				sources = append(sources, flowNode)
			}

			// File/IO reads
			if receiverName == "File" || receiverName == "IO" {
				if methodName == "read" || methodName == "readlines" || methodName == "open" {
					flowNode := &types.FlowNode{
						ID:         analyzer.GenerateNodeID("", node),
						Type:       types.NodeSource,
						Language:   "ruby",
						Line:       int(node.StartPoint().Row) + 1,
						Column:     int(node.StartPoint().Column),
						Name:       receiverName + "." + methodName,
						Snippet:    analyzer.GetNodeText(node, source),
						SourceType: types.SourceFile,
					}
					sources = append(sources, flowNode)
				}
			}
		}
	}

	// Check for element references (ARGV[], ENV[], params[], etc.)
	elemNodes := analyzer.FindNodesOfType(root, "element_reference")
	for _, node := range elemNodes {
		objNode := node.ChildByFieldName("object")
		if objNode != nil {
			objName := analyzer.GetNodeText(objNode, source)

			if sourceType, ok := a.inputSources[objName]; ok {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "ruby",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       objName + "[]",
					Snippet:    analyzer.GetNodeText(node, source),
					SourceType: sourceType,
				}
				sources = append(sources, flowNode)
			}
		}
	}

	// Check for constants (ARGV, STDIN, etc.)
	constNodes := analyzer.FindNodesOfType(root, "constant")
	for _, node := range constNodes {
		constName := analyzer.GetNodeText(node, source)

		if sourceType, ok := a.inputSources[constName]; ok {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "ruby",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       constName,
				Snippet:    constName,
				SourceType: sourceType,
			}
			sources = append(sources, flowNode)
		}
	}

	return sources, nil
}

func (a *RubyAnalyzer) DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error) {
	var frameworks []string

	for _, imp := range symbolTable.Imports {
		switch {
		case strings.Contains(imp.Path, "rails"):
			frameworks = append(frameworks, "Ruby on Rails")
		case strings.Contains(imp.Path, "sinatra"):
			frameworks = append(frameworks, "Sinatra")
		case strings.Contains(imp.Path, "hanami"):
			frameworks = append(frameworks, "Hanami")
		case strings.Contains(imp.Path, "grape"):
			frameworks = append(frameworks, "Grape")
		case strings.Contains(imp.Path, "padrino"):
			frameworks = append(frameworks, "Padrino")
		}
	}

	return frameworks, nil
}

func (a *RubyAnalyzer) AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*analyzer.MethodFlowAnalysis, error) {
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

func (a *RubyAnalyzer) TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error) {
	flowMap := types.NewFlowMap()
	flowMap.Target = target

	expr := target.Expression

	for fn, sourceType := range a.inputSources {
		if strings.Contains(expr, fn) {
			sourceNode := types.FlowNode{
				ID:         fmt.Sprintf("source-%s", fn),
				Type:       types.NodeSource,
				Language:   "ruby",
				Name:       fn,
				Snippet:    expr,
				SourceType: sourceType,
			}
			flowMap.AddSource(sourceNode)
		}
	}

	return flowMap, nil
}

func init() {
	analyzer.DefaultRegistry.Register(NewRubyAnalyzer())
}
