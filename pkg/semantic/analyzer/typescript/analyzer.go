// Package typescript implements the TypeScript language analyzer for semantic input tracing
package typescript

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/mappings"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// TypeScriptAnalyzer implements the LanguageAnalyzer interface for TypeScript
type TypeScriptAnalyzer struct {
	*analyzer.BaseAnalyzer
	inputSources   map[string]types.SourceType
	inputFunctions map[string]types.SourceType
}

// NewTypeScriptAnalyzer creates a new TypeScript analyzer
func NewTypeScriptAnalyzer() *TypeScriptAnalyzer {
	// TypeScript analyzer handles both .ts and .tsx files
	exts := languages.GetExtensionsForLanguage("typescript")
	exts = append(exts, languages.GetExtensionsForLanguage("tsx")...)
	m := mappings.GetMappings("typescript")
	a := &TypeScriptAnalyzer{
		BaseAnalyzer:   analyzer.NewBaseAnalyzer("typescript", exts),
		inputSources:   m.GetInputSourcesMap(),
		inputFunctions: m.GetInputFunctionsMap(),
	}

	a.registerFrameworkPatterns()
	return a
}

func (a *TypeScriptAnalyzer) registerFrameworkPatterns() {
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "express_req_body",
		Framework:       "express",
		Language:        "typescript",
		Name:            "Express req.body",
		Description:     "Express request body",
		PropertyPattern: "^body$",
		SourceType:      types.SourceHTTPBody,
		CarrierClass:    "Request",
		CarrierProperty: "body",
		PopulatedFrom:   []string{"HTTP body"},
		Confidence:      0.95,
	})

	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:            "nestjs_body",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @Body()",
		Description:   "NestJS body decorator",
		MethodPattern: "^Body$",
		SourceType:    types.SourceHTTPBody,
		PopulatedFrom: []string{"HTTP body"},
		Confidence:    0.95,
	})
}

func (a *TypeScriptAnalyzer) BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error) {
	st := types.NewSymbolTable(filePath, "typescript")
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

func (a *TypeScriptAnalyzer) extractImports(root *sitter.Node, source []byte) []types.ImportInfo {
	var imports []types.ImportInfo

	importNodes := analyzer.FindNodesOfType(root, "import_statement")
	for _, node := range importNodes {
		sourceNode := analyzer.FindChildByType(node, "string")
		if sourceNode == nil {
			continue
		}
		path := strings.Trim(analyzer.GetNodeText(sourceNode, source), "\"'")

		imp := types.ImportInfo{
			Path: path,
			Line: int(node.StartPoint().Row) + 1,
			Type: "import",
		}

		clauseNode := analyzer.FindChildByType(node, "import_clause")
		if clauseNode != nil {
			var names []string
			nameNode := analyzer.FindChildByType(clauseNode, "identifier")
			if nameNode != nil {
				names = append(names, analyzer.GetNodeText(nameNode, source))
			}
			namedImports := analyzer.FindChildByType(clauseNode, "named_imports")
			if namedImports != nil {
				specifiers := analyzer.FindNodesOfType(namedImports, "import_specifier")
				for _, spec := range specifiers {
					n := analyzer.FindChildByType(spec, "identifier")
					if n != nil {
						names = append(names, analyzer.GetNodeText(n, source))
					}
				}
			}
			imp.Names = names
		}

		imports = append(imports, imp)
	}

	return imports
}

func (a *TypeScriptAnalyzer) ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error) {
	return nil, nil
}

func (a *TypeScriptAnalyzer) ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error) {
	var classes []*types.ClassDef

	classNodes := analyzer.FindNodesOfType(root, "class_declaration")
	for _, classNode := range classNodes {
		class := a.parseClassDeclaration(classNode, source)
		if class != nil {
			classes = append(classes, class)
		}
	}

	return classes, nil
}

func (a *TypeScriptAnalyzer) parseClassDeclaration(node *sitter.Node, source []byte) *types.ClassDef {
	nameNode := analyzer.FindChildByType(node, "type_identifier")
	if nameNode == nil {
		nameNode = analyzer.FindChildByType(node, "identifier")
	}
	if nameNode == nil {
		return nil
	}

	name := analyzer.GetNodeText(nameNode, source)
	class := types.NewClassDef(name, "", int(node.StartPoint().Row)+1)
	class.EndLine = int(node.EndPoint().Row) + 1

	heritageNode := analyzer.FindChildByType(node, "class_heritage")
	if heritageNode != nil {
		extendsNode := analyzer.FindChildByType(heritageNode, "extends_clause")
		if extendsNode != nil {
			typeNode := analyzer.FindChildByType(extendsNode, "type_identifier")
			if typeNode == nil {
				typeNode = analyzer.FindChildByType(extendsNode, "identifier")
			}
			if typeNode != nil {
				class.Extends = analyzer.GetNodeText(typeNode, source)
			}
		}

		implementsNode := analyzer.FindChildByType(heritageNode, "implements_clause")
		if implementsNode != nil {
			typeNodes := analyzer.FindNodesOfType(implementsNode, "type_identifier")
			for _, t := range typeNodes {
				class.Implements = append(class.Implements, analyzer.GetNodeText(t, source))
			}
		}
	}

	bodyNode := analyzer.FindChildByType(node, "class_body")
	if bodyNode != nil {
		a.parseClassBody(class, bodyNode, source)
	}

	return class
}

func (a *TypeScriptAnalyzer) parseClassBody(class *types.ClassDef, bodyNode *sitter.Node, source []byte) {
	for i := 0; i < int(bodyNode.ChildCount()); i++ {
		child := bodyNode.Child(i)
		switch child.Type() {
		case "public_field_definition", "field_definition":
			prop := a.parsePropertyDefinition(child, source)
			if prop != nil {
				class.Properties[prop.Name] = prop
			}
		case "method_definition":
			method := a.parseMethodDefinition(child, source)
			if method != nil {
				if method.Name == "constructor" {
					class.Constructor = method
				}
				class.Methods[method.Name] = method
			}
		}
	}
}

func (a *TypeScriptAnalyzer) parsePropertyDefinition(node *sitter.Node, source []byte) *types.PropertyDef {
	nameNode := analyzer.FindChildByType(node, "property_identifier")
	if nameNode == nil {
		return nil
	}

	prop := &types.PropertyDef{
		Name:       analyzer.GetNodeText(nameNode, source),
		Visibility: "public",
		Line:       int(node.StartPoint().Row) + 1,
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "accessibility_modifier":
			prop.Visibility = analyzer.GetNodeText(child, source)
		case "static":
			prop.IsStatic = true
		case "readonly":
			prop.IsReadonly = true
		}
	}

	typeNode := analyzer.FindChildByType(node, "type_annotation")
	if typeNode != nil {
		prop.Type = analyzer.GetNodeText(typeNode, source)
	}

	return prop
}

func (a *TypeScriptAnalyzer) parseMethodDefinition(node *sitter.Node, source []byte) *types.MethodDef {
	nameNode := analyzer.FindChildByType(node, "property_identifier")
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

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "accessibility_modifier":
			method.Visibility = analyzer.GetNodeText(child, source)
		case "static":
			method.IsStatic = true
		case "async":
			method.IsAsync = true
		}
	}

	paramsNode := analyzer.FindChildByType(node, "formal_parameters")
	if paramsNode != nil {
		method.Parameters = a.parseParameters(paramsNode, source)
	}

	returnNode := analyzer.FindChildByType(node, "type_annotation")
	if returnNode != nil {
		method.ReturnType = strings.TrimPrefix(analyzer.GetNodeText(returnNode, source), ": ")
	}

	bodyNode := analyzer.FindChildByType(node, "statement_block")
	if bodyNode != nil {
		method.BodyStart = int(bodyNode.StartPoint().Row) + 1
		method.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		method.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	return method
}

func (a *TypeScriptAnalyzer) parseParameters(node *sitter.Node, source []byte) []types.ParameterDef {
	var params []types.ParameterDef
	index := 0

	paramNodes := analyzer.FindNodesOfType(node, "required_parameter")
	paramNodes = append(paramNodes, analyzer.FindNodesOfType(node, "optional_parameter")...)

	for _, paramNode := range paramNodes {
		nameNode := analyzer.FindChildByType(paramNode, "identifier")
		if nameNode == nil {
			continue
		}

		param := types.ParameterDef{
			Name:  analyzer.GetNodeText(nameNode, source),
			Index: index,
		}

		typeNode := analyzer.FindChildByType(paramNode, "type_annotation")
		if typeNode != nil {
			param.Type = strings.TrimPrefix(analyzer.GetNodeText(typeNode, source), ": ")
		}

		params = append(params, param)
		index++
	}

	restNodes := analyzer.FindNodesOfType(node, "rest_parameter")
	for _, restNode := range restNodes {
		nameNode := analyzer.FindChildByType(restNode, "identifier")
		if nameNode != nil {
			params = append(params, types.ParameterDef{
				Name:       analyzer.GetNodeText(nameNode, source),
				Index:      index,
				IsVariadic: true,
			})
			index++
		}
	}

	return params
}

func (a *TypeScriptAnalyzer) ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error) {
	var functions []*types.FunctionDef

	funcNodes := analyzer.FindNodesOfType(root, "function_declaration")
	for _, funcNode := range funcNodes {
		if analyzer.GetEnclosingClass(funcNode, []string{"class_declaration"}) != nil {
			continue
		}

		fn := a.parseFunctionDeclaration(funcNode, source)
		if fn != nil {
			functions = append(functions, fn)
		}
	}

	return functions, nil
}

func (a *TypeScriptAnalyzer) parseFunctionDeclaration(node *sitter.Node, source []byte) *types.FunctionDef {
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

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "async" {
			fn.IsAsync = true
		}
	}

	parent := node.Parent()
	if parent != nil && parent.Type() == "export_statement" {
		fn.IsExported = true
	}

	paramsNode := analyzer.FindChildByType(node, "formal_parameters")
	if paramsNode != nil {
		fn.Parameters = a.parseParameters(paramsNode, source)
	}

	returnNode := analyzer.FindChildByType(node, "type_annotation")
	if returnNode != nil {
		fn.ReturnType = strings.TrimPrefix(analyzer.GetNodeText(returnNode, source), ": ")
	}

	bodyNode := analyzer.FindChildByType(node, "statement_block")
	if bodyNode != nil {
		fn.BodyStart = int(bodyNode.StartPoint().Row) + 1
		fn.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		fn.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	return fn
}

func (a *TypeScriptAnalyzer) ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error) {
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

func (a *TypeScriptAnalyzer) isExpressionTainted(node *sitter.Node, source []byte) (bool, string) {
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
		if strings.Contains(text, fn+"(") {
			return true, fn + "()"
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		if tainted, src := a.isExpressionTainted(node.Child(i), source); tainted {
			return true, src
		}
	}

	return false, ""
}

func (a *TypeScriptAnalyzer) ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error) {
	var calls []*types.CallSite

	callNodes := analyzer.FindNodesOfType(root, "call_expression")
	for _, node := range callNodes {
		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			funcNode = node.Child(0)
		}
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

		if funcNode.Type() == "member_expression" {
			objNode := funcNode.ChildByFieldName("object")
			propNode := funcNode.ChildByFieldName("property")
			if objNode != nil {
				call.ClassName = analyzer.GetNodeText(objNode, source)
			}
			if propNode != nil {
				call.MethodName = analyzer.GetNodeText(propNode, source)
			}
		}

		argsNode := analyzer.FindChildByType(node, "arguments")
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

func (a *TypeScriptAnalyzer) parseCallArguments(node *sitter.Node, source []byte) []types.CallArg {
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

func (a *TypeScriptAnalyzer) FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error) {
	var sources []*types.FlowNode

	memberNodes := analyzer.FindNodesOfType(root, "member_expression")
	for _, node := range memberNodes {
		text := analyzer.GetNodeText(node, source)
		for src, sourceType := range a.inputSources {
			if strings.HasPrefix(text, src) {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "typescript",
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

	callNodes := analyzer.FindNodesOfType(root, "call_expression")
	for _, node := range callNodes {
		funcNode := node.Child(0)
		if funcNode == nil {
			continue
		}
		funcName := analyzer.GetNodeText(funcNode, source)

		if sourceType, ok := a.inputFunctions[funcName]; ok {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "typescript",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       funcName,
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: sourceType,
			}
			sources = append(sources, flowNode)
		}
	}

	return sources, nil
}

func (a *TypeScriptAnalyzer) DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error) {
	var frameworks []string

	for _, imp := range symbolTable.Imports {
		path := strings.ToLower(imp.Path)

		if strings.Contains(path, "express") {
			if !contains(frameworks, "express") {
				frameworks = append(frameworks, "express")
			}
		}
		if strings.Contains(path, "@nestjs") {
			if !contains(frameworks, "nestjs") {
				frameworks = append(frameworks, "nestjs")
			}
		}
		if strings.Contains(path, "koa") {
			if !contains(frameworks, "koa") {
				frameworks = append(frameworks, "koa")
			}
		}
		if strings.Contains(path, "fastify") {
			if !contains(frameworks, "fastify") {
				frameworks = append(frameworks, "fastify")
			}
		}
	}

	return frameworks, nil
}

func (a *TypeScriptAnalyzer) AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*analyzer.MethodFlowAnalysis, error) {
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

		thisAssignRegex := regexp.MustCompile(`this\.(\w+)\s*=.*\b` + regexp.QuoteMeta(param.Name) + `\b`)
		matches := thisAssignRegex.FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			if len(match) > 1 {
				analysis.ParamsToProperties[i] = append(analysis.ParamsToProperties[i], match[1])
				analysis.ModifiesProperties = true
			}
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

func (a *TypeScriptAnalyzer) TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error) {
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
					Language:   "typescript",
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
				key := a.extractKeyFromExpression(expr)
				sourceNode := types.FlowNode{
					ID:         fmt.Sprintf("source-%s-%s", src, key),
					Type:       types.NodeSource,
					Language:   "typescript",
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

func (a *TypeScriptAnalyzer) matchesFrameworkPattern(expr string, pattern *types.FrameworkPattern) bool {
	if pattern.PropertyPattern != "" {
		regex := regexp.MustCompile(`\.` + pattern.PropertyPattern + `\b`)
		return regex.MatchString(expr)
	}

	if pattern.MethodPattern != "" {
		regex := regexp.MustCompile(`@` + pattern.MethodPattern + `\(`)
		return regex.MatchString(expr)
	}

	return false
}

func (a *TypeScriptAnalyzer) extractKeyFromExpression(expr string) string {
	regex := regexp.MustCompile(`\[['"\x60](\w+)['"\x60]\]`)
	matches := regex.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}

	regex2 := regexp.MustCompile(`\.(body|query|params|headers|cookies)\.(\w+)`)
	matches2 := regex2.FindStringSubmatch(expr)
	if len(matches2) > 2 {
		return matches2[2]
	}

	return ""
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
	analyzer.DefaultRegistry.Register(NewTypeScriptAnalyzer())
}
