// Package rust implements the Rust language analyzer for semantic input tracing
package rust

import (
	"fmt"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// RustAnalyzer implements the LanguageAnalyzer interface for Rust
type RustAnalyzer struct {
	*analyzer.BaseAnalyzer
	inputSources map[string]types.SourceType
}

// NewRustAnalyzer creates a new Rust analyzer
func NewRustAnalyzer() *RustAnalyzer {
	a := &RustAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer("rust", []string{".rs"}),
	}

	a.inputSources = map[string]types.SourceType{
		"env::args":       types.SourceCLIArg,
		"env::args_os":    types.SourceCLIArg,
		"env::var":        types.SourceEnvVar,
		"env::var_os":     types.SourceEnvVar,
		"stdin":           types.SourceStdin,
		"read_line":       types.SourceStdin,
		"BufRead":         types.SourceStdin,
		"fs::read":        types.SourceFile,
		"read_to_string":  types.SourceFile,
		"File::open":      types.SourceFile,
		"web::Query":      types.SourceHTTPGet,
		"web::Form":       types.SourceHTTPPost,
		"web::Json":       types.SourceHTTPBody,
		"web::Path":       types.SourceHTTPGet,
		"Query<":          types.SourceHTTPGet,
		"Form<":           types.SourceHTTPPost,
		"Json<":           types.SourceHTTPBody,
		"Path<":           types.SourceHTTPGet,
	}

	return a
}

func (a *RustAnalyzer) BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error) {
	st := types.NewSymbolTable(filePath, "rust")
	st.Imports = a.extractUses(root, source)

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

func (a *RustAnalyzer) extractUses(root *sitter.Node, source []byte) []types.ImportInfo {
	var imports []types.ImportInfo

	useNodes := analyzer.FindNodesOfType(root, "use_declaration")
	for _, node := range useNodes {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "use_tree" || child.Type() == "scoped_identifier" {
				usePath := analyzer.GetNodeText(child, source)
				imports = append(imports, types.ImportInfo{
					Path: usePath,
					Line: int(node.StartPoint().Row) + 1,
					Type: "use",
				})
			}
		}
	}

	return imports
}

func (a *RustAnalyzer) ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error) {
	return nil, nil
}

func (a *RustAnalyzer) ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error) {
	var classes []*types.ClassDef

	// Extract structs
	structNodes := analyzer.FindNodesOfType(root, "struct_item")
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

	// Extract enums
	enumNodes := analyzer.FindNodesOfType(root, "enum_item")
	for _, enumNode := range enumNodes {
		nameNode := enumNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		name := analyzer.GetNodeText(nameNode, source)
		class := types.NewClassDef(name, "", int(enumNode.StartPoint().Row)+1)
		class.EndLine = int(enumNode.EndPoint().Row) + 1

		classes = append(classes, class)
	}

	// Extract traits
	traitNodes := analyzer.FindNodesOfType(root, "trait_item")
	for _, traitNode := range traitNodes {
		nameNode := traitNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		name := analyzer.GetNodeText(nameNode, source)
		class := types.NewClassDef(name, "", int(traitNode.StartPoint().Row)+1)
		class.EndLine = int(traitNode.EndPoint().Row) + 1

		classes = append(classes, class)
	}

	// Extract impl blocks and add methods to corresponding structs
	implNodes := analyzer.FindNodesOfType(root, "impl_item")
	for _, implNode := range implNodes {
		typeNode := implNode.ChildByFieldName("type")
		if typeNode == nil {
			continue
		}

		typeName := analyzer.GetNodeText(typeNode, source)

		// Find or create the class
		var targetClass *types.ClassDef
		for _, c := range classes {
			if c.Name == typeName {
				targetClass = c
				break
			}
		}

		if targetClass == nil {
			targetClass = types.NewClassDef(typeName, "", int(implNode.StartPoint().Row)+1)
			targetClass.EndLine = int(implNode.EndPoint().Row) + 1
			classes = append(classes, targetClass)
		}

		// Extract methods from impl body
		bodyNode := implNode.ChildByFieldName("body")
		if bodyNode != nil {
			methods := a.extractMethodsFromImpl(bodyNode, source, typeName)
			for _, method := range methods {
				targetClass.Methods[method.Name] = method
			}
		}
	}

	return classes, nil
}

func (a *RustAnalyzer) extractMethodsFromImpl(bodyNode *sitter.Node, source []byte, typeName string) []*types.MethodDef {
	var methods []*types.MethodDef

	funcNodes := analyzer.FindNodesOfType(bodyNode, "function_item")
	for _, funcNode := range funcNodes {
		nameNode := funcNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		method := &types.MethodDef{
			Name:    analyzer.GetNodeText(nameNode, source),
			Line:    int(funcNode.StartPoint().Row) + 1,
			EndLine: int(funcNode.EndPoint().Row) + 1,
		}

		paramsNode := funcNode.ChildByFieldName("parameters")
		if paramsNode != nil {
			method.Parameters = a.parseParameters(paramsNode, source)
			// Check if first param is self (instance method) or not (associated function)
			if len(method.Parameters) == 0 || !strings.Contains(method.Parameters[0].Name, "self") {
				method.IsStatic = true
			}
		} else {
			method.IsStatic = true
		}

		returnTypeNode := funcNode.ChildByFieldName("return_type")
		if returnTypeNode != nil {
			method.ReturnType = analyzer.GetNodeText(returnTypeNode, source)
		}

		bodyNode := funcNode.ChildByFieldName("body")
		if bodyNode != nil {
			method.BodyStart = int(bodyNode.StartPoint().Row) + 1
			method.BodyEnd = int(bodyNode.EndPoint().Row) + 1
			method.BodySource = analyzer.GetNodeText(bodyNode, source)
		}

		methods = append(methods, method)
	}

	return methods
}

func (a *RustAnalyzer) ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error) {
	var functions []*types.FunctionDef

	funcNodes := analyzer.FindNodesOfType(root, "function_item")
	for _, funcNode := range funcNodes {
		// Skip if inside an impl block
		parent := funcNode.Parent()
		insideImpl := false
		for parent != nil {
			if parent.Type() == "impl_item" {
				insideImpl = true
				break
			}
			parent = parent.Parent()
		}

		if insideImpl {
			continue
		}

		nameNode := funcNode.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}

		fn := &types.FunctionDef{
			Name:    analyzer.GetNodeText(nameNode, source),
			Line:    int(funcNode.StartPoint().Row) + 1,
			EndLine: int(funcNode.EndPoint().Row) + 1,
		}

		paramsNode := funcNode.ChildByFieldName("parameters")
		if paramsNode != nil {
			fn.Parameters = a.parseParameters(paramsNode, source)
		}

		returnTypeNode := funcNode.ChildByFieldName("return_type")
		if returnTypeNode != nil {
			fn.ReturnType = analyzer.GetNodeText(returnTypeNode, source)
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

func (a *RustAnalyzer) parseParameters(node *sitter.Node, source []byte) []types.ParameterDef {
	var params []types.ParameterDef
	index := 0

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()

		switch childType {
		case "parameter":
			param := types.ParameterDef{Index: index}

			patternNode := child.ChildByFieldName("pattern")
			if patternNode != nil {
				param.Name = analyzer.GetNodeText(patternNode, source)
			}

			typeNode := child.ChildByFieldName("type")
			if typeNode != nil {
				param.Type = analyzer.GetNodeText(typeNode, source)
			}

			if param.Name != "" {
				params = append(params, param)
				index++
			}

		case "self_parameter":
			params = append(params, types.ParameterDef{
				Name:  analyzer.GetNodeText(child, source),
				Index: index,
			})
			index++

		case "variadic_parameter":
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

func (a *RustAnalyzer) ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error) {
	var assignments []*types.Assignment

	// Assignment expressions
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

	// Let declarations
	letNodes := analyzer.FindNodesOfType(root, "let_declaration")
	for _, node := range letNodes {
		patternNode := node.ChildByFieldName("pattern")
		valueNode := node.ChildByFieldName("value")
		if patternNode != nil && valueNode != nil {
			assignment := &types.Assignment{
				Target:   analyzer.GetNodeText(patternNode, source),
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

func (a *RustAnalyzer) isExpressionTainted(node *sitter.Node, source []byte) (bool, string) {
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

func (a *RustAnalyzer) ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error) {
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

		// Check for method call
		if funcNode.Type() == "field_expression" {
			valueNode := funcNode.ChildByFieldName("value")
			fieldNode := funcNode.ChildByFieldName("field")
			if valueNode != nil && fieldNode != nil {
				call.ClassName = analyzer.GetNodeText(valueNode, source)
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

func (a *RustAnalyzer) parseCallArguments(node *sitter.Node, source []byte) []types.CallArg {
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

func (a *RustAnalyzer) FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error) {
	var sources []*types.FlowNode

	// Check call expressions
	callNodes := analyzer.FindNodesOfType(root, "call_expression")
	for _, node := range callNodes {
		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			continue
		}

		funcName := analyzer.GetNodeText(funcNode, source)

		for pattern, sourceType := range a.inputSources {
			if strings.Contains(funcName, pattern) {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "rust",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       pattern,
					Snippet:    analyzer.GetNodeText(node, source),
					SourceType: sourceType,
				}
				sources = append(sources, flowNode)
				break
			}
		}
	}

	// Check generic types for web extractors (Query<>, Form<>, etc.)
	genericNodes := analyzer.FindNodesOfType(root, "generic_type")
	for _, node := range genericNodes {
		typeName := analyzer.GetNodeText(node, source)

		for pattern, sourceType := range a.inputSources {
			if strings.HasPrefix(typeName, pattern) {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "rust",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       pattern,
					Snippet:    typeName,
					SourceType: sourceType,
				}
				sources = append(sources, flowNode)
				break
			}
		}
	}

	// Check attributes for derive macros
	attrNodes := analyzer.FindNodesOfType(root, "attribute_item")
	for _, node := range attrNodes {
		attrText := analyzer.GetNodeText(node, source)

		if strings.Contains(attrText, "FromForm") {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "rust",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "#[derive(FromForm)]",
				Snippet:    attrText,
				SourceType: types.SourceHTTPPost,
			}
			sources = append(sources, flowNode)
		} else if strings.Contains(attrText, "Parser") {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "rust",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "#[derive(Parser)]",
				Snippet:    attrText,
				SourceType: types.SourceCLIArg,
			}
			sources = append(sources, flowNode)
		}
	}

	return sources, nil
}

func (a *RustAnalyzer) DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error) {
	var frameworks []string

	for _, imp := range symbolTable.Imports {
		switch {
		case strings.Contains(imp.Path, "actix_web"):
			frameworks = append(frameworks, "Actix-web")
		case strings.Contains(imp.Path, "rocket"):
			frameworks = append(frameworks, "Rocket")
		case strings.Contains(imp.Path, "axum"):
			frameworks = append(frameworks, "Axum")
		case strings.Contains(imp.Path, "warp"):
			frameworks = append(frameworks, "Warp")
		case strings.Contains(imp.Path, "tide"):
			frameworks = append(frameworks, "Tide")
		case strings.Contains(imp.Path, "clap"):
			frameworks = append(frameworks, "Clap CLI")
		case strings.Contains(imp.Path, "structopt"):
			frameworks = append(frameworks, "StructOpt CLI")
		}
	}

	return frameworks, nil
}

func (a *RustAnalyzer) AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*analyzer.MethodFlowAnalysis, error) {
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

func (a *RustAnalyzer) TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error) {
	flowMap := types.NewFlowMap()
	flowMap.Target = target

	expr := target.Expression

	for fn, sourceType := range a.inputSources {
		if strings.Contains(expr, fn) {
			sourceNode := types.FlowNode{
				ID:         fmt.Sprintf("source-%s", fn),
				Type:       types.NodeSource,
				Language:   "rust",
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
	analyzer.DefaultRegistry.Register(NewRustAnalyzer())
}
