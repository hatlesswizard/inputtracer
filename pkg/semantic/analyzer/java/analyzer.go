// Package java implements the Java language analyzer for semantic input tracing
package java

import (
	"fmt"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/mappings"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// JavaAnalyzer implements the LanguageAnalyzer interface for Java
type JavaAnalyzer struct {
	*analyzer.BaseAnalyzer
	inputMethods map[string]types.SourceType
}

// NewJavaAnalyzer creates a new Java analyzer
func NewJavaAnalyzer() *JavaAnalyzer {
	m := mappings.GetMappings("java")
	a := &JavaAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer("java", languages.GetExtensionsForLanguage("java")),
		inputMethods: m.GetInputMethodsMap(),
	}

	a.registerFrameworkPatterns()
	return a
}

func (a *JavaAnalyzer) registerFrameworkPatterns() {
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:            "servlet_request",
		Framework:     "servlet",
		Language:      "java",
		Name:          "HttpServletRequest",
		Description:   "Servlet request parameters",
		MethodPattern: "^getParameter",
		SourceType:    types.SourceHTTPGet,
		CarrierClass:  "HttpServletRequest",
		PopulatedFrom: []string{"query string", "form data"},
		Confidence:    0.95,
	})

	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:            "spring_requestparam",
		Framework:     "spring",
		Language:      "java",
		Name:          "Spring @RequestParam",
		Description:   "Spring request parameter annotation",
		MethodPattern: "^RequestParam$",
		SourceType:    types.SourceHTTPGet,
		PopulatedFrom: []string{"query string"},
		Confidence:    0.95,
	})

	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:            "spring_requestbody",
		Framework:     "spring",
		Language:      "java",
		Name:          "Spring @RequestBody",
		Description:   "Spring request body annotation",
		MethodPattern: "^RequestBody$",
		SourceType:    types.SourceHTTPBody,
		PopulatedFrom: []string{"HTTP body"},
		Confidence:    0.95,
	})
}

func (a *JavaAnalyzer) BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error) {
	st := types.NewSymbolTable(filePath, "java")
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

func (a *JavaAnalyzer) extractImports(root *sitter.Node, source []byte) []types.ImportInfo {
	var imports []types.ImportInfo

	importNodes := analyzer.FindNodesOfType(root, "import_declaration")
	for _, node := range importNodes {
		text := analyzer.GetNodeText(node, source)
		importPath := strings.TrimPrefix(text, "import ")
		importPath = strings.TrimPrefix(importPath, "static ")
		importPath = strings.TrimSuffix(importPath, ";")
		importPath = strings.TrimSpace(importPath)

		parts := strings.Split(importPath, ".")
		alias := parts[len(parts)-1]
		if alias == "*" && len(parts) > 1 {
			alias = parts[len(parts)-2]
		}

		imports = append(imports, types.ImportInfo{
			Path:  importPath,
			Alias: alias,
			Line:  int(node.StartPoint().Row) + 1,
			Type:  "import",
		})
	}

	return imports
}

func (a *JavaAnalyzer) ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error) {
	return nil, nil
}

func (a *JavaAnalyzer) ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error) {
	var classes []*types.ClassDef

	classNodes := analyzer.FindNodesOfType(root, "class_declaration")
	for _, classNode := range classNodes {
		class := a.parseClassDeclaration(classNode, source)
		if class != nil {
			classes = append(classes, class)
		}
	}

	interfaceNodes := analyzer.FindNodesOfType(root, "interface_declaration")
	for _, interfaceNode := range interfaceNodes {
		class := a.parseClassDeclaration(interfaceNode, source)
		if class != nil {
			classes = append(classes, class)
		}
	}

	return classes, nil
}

func (a *JavaAnalyzer) parseClassDeclaration(node *sitter.Node, source []byte) *types.ClassDef {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := analyzer.GetNodeText(nameNode, source)
	class := types.NewClassDef(name, "", int(node.StartPoint().Row)+1)
	class.EndLine = int(node.EndPoint().Row) + 1

	// Get superclass
	superclassNode := node.ChildByFieldName("superclass")
	if superclassNode != nil {
		class.Extends = analyzer.GetNodeText(superclassNode, source)
	}

	// Get interfaces
	interfacesNode := node.ChildByFieldName("interfaces")
	if interfacesNode != nil {
		for i := 0; i < int(interfacesNode.ChildCount()); i++ {
			child := interfacesNode.Child(i)
			if child.Type() == "type_identifier" || child.Type() == "generic_type" {
				class.Implements = append(class.Implements, analyzer.GetNodeText(child, source))
			}
		}
	}

	// Parse body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		a.parseClassBody(class, bodyNode, source)
	}

	return class
}

func (a *JavaAnalyzer) parseClassBody(class *types.ClassDef, bodyNode *sitter.Node, source []byte) {
	methodNodes := analyzer.FindNodesOfType(bodyNode, "method_declaration")
	for _, methodNode := range methodNodes {
		method := a.parseMethodDeclaration(methodNode, source)
		if method != nil {
			class.Methods[method.Name] = method
		}
	}

	constructorNodes := analyzer.FindNodesOfType(bodyNode, "constructor_declaration")
	for _, constructorNode := range constructorNodes {
		method := a.parseMethodDeclaration(constructorNode, source)
		if method != nil {
			class.Constructor = method
			class.Methods[method.Name] = method
		}
	}

	fieldNodes := analyzer.FindNodesOfType(bodyNode, "field_declaration")
	for _, fieldNode := range fieldNodes {
		props := a.parseFieldDeclaration(fieldNode, source)
		for _, prop := range props {
			class.Properties[prop.Name] = prop
		}
	}
}

func (a *JavaAnalyzer) parseMethodDeclaration(node *sitter.Node, source []byte) *types.MethodDef {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	method := &types.MethodDef{
		Name:       analyzer.GetNodeText(nameNode, source),
		Line:       int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Visibility: "public",
		Parameters: make([]types.ParameterDef, 0),
	}

	// Get return type
	typeNode := node.ChildByFieldName("type")
	if typeNode != nil {
		method.ReturnType = analyzer.GetNodeText(typeNode, source)
	}

	// Get parameters
	paramsNode := node.ChildByFieldName("parameters")
	if paramsNode != nil {
		method.Parameters = a.parseParameters(paramsNode, source)
	}

	// Get body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		method.BodyStart = int(bodyNode.StartPoint().Row) + 1
		method.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		method.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	// Check modifiers
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "modifiers" {
			modContent := analyzer.GetNodeText(child, source)
			if strings.Contains(modContent, "static") {
				method.IsStatic = true
			}
			if strings.Contains(modContent, "private") {
				method.Visibility = "private"
			} else if strings.Contains(modContent, "protected") {
				method.Visibility = "protected"
			}
		}
	}

	return method
}

func (a *JavaAnalyzer) parseFieldDeclaration(node *sitter.Node, source []byte) []*types.PropertyDef {
	var props []*types.PropertyDef

	typeNode := node.ChildByFieldName("type")
	typeName := ""
	if typeNode != nil {
		typeName = analyzer.GetNodeText(typeNode, source)
	}

	declarators := analyzer.FindNodesOfType(node, "variable_declarator")
	for _, declarator := range declarators {
		nameNode := declarator.ChildByFieldName("name")
		if nameNode != nil {
			prop := &types.PropertyDef{
				Name:       analyzer.GetNodeText(nameNode, source),
				Type:       typeName,
				Line:       int(node.StartPoint().Row) + 1,
				Visibility: "private",
			}

			// Check modifiers
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child.Type() == "modifiers" {
					modContent := analyzer.GetNodeText(child, source)
					if strings.Contains(modContent, "static") {
						prop.IsStatic = true
					}
					if strings.Contains(modContent, "public") {
						prop.Visibility = "public"
					} else if strings.Contains(modContent, "protected") {
						prop.Visibility = "protected"
					}
				}
			}

			props = append(props, prop)
		}
	}

	return props
}

func (a *JavaAnalyzer) parseParameters(node *sitter.Node, source []byte) []types.ParameterDef {
	var params []types.ParameterDef
	index := 0

	paramNodes := analyzer.FindNodesOfType(node, "formal_parameter")
	for _, paramNode := range paramNodes {
		param := types.ParameterDef{Index: index}

		typeNode := paramNode.ChildByFieldName("type")
		if typeNode != nil {
			param.Type = analyzer.GetNodeText(typeNode, source)
		}

		nameNode := paramNode.ChildByFieldName("name")
		if nameNode != nil {
			param.Name = analyzer.GetNodeText(nameNode, source)
		}

		if param.Name != "" {
			params = append(params, param)
			index++
		}
	}

	// Spread parameter
	spreadNodes := analyzer.FindNodesOfType(node, "spread_parameter")
	for _, spreadNode := range spreadNodes {
		param := types.ParameterDef{
			Index:      index,
			IsVariadic: true,
		}

		nameNode := spreadNode.ChildByFieldName("name")
		if nameNode != nil {
			param.Name = analyzer.GetNodeText(nameNode, source)
		}

		if param.Name != "" {
			params = append(params, param)
			index++
		}
	}

	return params
}

func (a *JavaAnalyzer) ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error) {
	// In Java, functions are always part of a class
	// We extract methods from all classes and add them as functions for reference
	var functions []*types.FunctionDef

	methodNodes := analyzer.FindNodesOfType(root, "method_declaration")
	for _, methodNode := range methodNodes {
		// Find enclosing class
		classNode := analyzer.GetEnclosingClass(methodNode, []string{"class_declaration"})
		className := ""
		if classNode != nil {
			nameNode := classNode.ChildByFieldName("name")
			if nameNode != nil {
				className = analyzer.GetNodeText(nameNode, source)
			}
		}

		fn := a.parseFunction(methodNode, source, className)
		if fn != nil {
			functions = append(functions, fn)
		}
	}

	return functions, nil
}

func (a *JavaAnalyzer) parseFunction(node *sitter.Node, source []byte, className string) *types.FunctionDef {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := analyzer.GetNodeText(nameNode, source)
	if className != "" {
		name = className + "." + name
	}

	fn := &types.FunctionDef{
		Name:       name,
		Line:       int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Parameters: make([]types.ParameterDef, 0),
	}

	typeNode := node.ChildByFieldName("type")
	if typeNode != nil {
		fn.ReturnType = analyzer.GetNodeText(typeNode, source)
	}

	paramsNode := node.ChildByFieldName("parameters")
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

func (a *JavaAnalyzer) ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error) {
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

	// Variable declarators
	varDeclarators := analyzer.FindNodesOfType(root, "variable_declarator")
	for _, node := range varDeclarators {
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

func (a *JavaAnalyzer) isExpressionTainted(node *sitter.Node, source []byte) (bool, string) {
	if node == nil {
		return false, ""
	}

	text := analyzer.GetNodeText(node, source)

	for method := range a.inputMethods {
		if strings.Contains(text, "."+method+"(") {
			return true, method
		}
	}

	if strings.Contains(text, "args[") {
		return true, "args[]"
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		if tainted, src := a.isExpressionTainted(node.Child(i), source); tainted {
			return true, src
		}
	}

	return false, ""
}

func (a *JavaAnalyzer) ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error) {
	var calls []*types.CallSite

	callNodes := analyzer.FindNodesOfType(root, "method_invocation")
	for _, node := range callNodes {
		call := a.parseMethodInvocation(node, source, scope)
		if call != nil {
			calls = append(calls, call)
		}
	}

	return calls, nil
}

func (a *JavaAnalyzer) parseMethodInvocation(node *sitter.Node, source []byte, scope string) *types.CallSite {
	nameNode := node.ChildByFieldName("name")
	objNode := node.ChildByFieldName("object")

	if nameNode == nil {
		return nil
	}

	call := &types.CallSite{
		FunctionName: analyzer.GetNodeText(nameNode, source),
		MethodName:   analyzer.GetNodeText(nameNode, source),
		Line:         int(node.StartPoint().Row) + 1,
		Column:       int(node.StartPoint().Column),
		Scope:        scope,
		Arguments:    make([]types.CallArg, 0),
	}

	if objNode != nil {
		call.ClassName = analyzer.GetNodeText(objNode, source)
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

	return call
}

func (a *JavaAnalyzer) parseCallArguments(node *sitter.Node, source []byte) []types.CallArg {
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

// inputAnnotations maps Java framework annotation names to their source types
// This includes Spring MVC, JAX-RS (Jersey, RESTEasy), and other frameworks
var inputAnnotations = map[string]types.SourceType{
	// Spring MVC annotations
	"RequestParam":   types.SourceHTTPGet,
	"PathVariable":   types.SourceHTTPGet,
	"RequestBody":    types.SourceHTTPBody,
	"RequestHeader":  types.SourceHTTPHeader,
	"CookieValue":    types.SourceHTTPCookie,
	"ModelAttribute": types.SourceHTTPPost,
	"RequestPart":    types.SourceHTTPBody,
	"MatrixVariable": types.SourceHTTPGet,

	// JAX-RS annotations (Jersey, RESTEasy, Quarkus)
	"QueryParam":  types.SourceHTTPGet,
	"PathParam":   types.SourceHTTPGet,
	"FormParam":   types.SourceHTTPPost,
	"HeaderParam": types.SourceHTTPHeader,
	"CookieParam": types.SourceHTTPCookie,
	"BeanParam":   types.SourceUserInput,
	"MatrixParam": types.SourceHTTPGet,

	// Micronaut annotations (some overlap with Spring)
	"QueryValue": types.SourceHTTPGet,
	"PathValue":  types.SourceHTTPGet,
	"Body":       types.SourceHTTPBody,
	"Header":     types.SourceHTTPHeader,

	// Vert.x Web annotations
	"Param": types.SourceHTTPGet,
}

func (a *JavaAnalyzer) FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error) {
	var sources []*types.FlowNode

	// Spring annotations (marker_annotation and annotation)
	sources = append(sources, a.findSpringAnnotations(root, source)...)

	// Method invocations
	callNodes := analyzer.FindNodesOfType(root, "method_invocation")
	for _, node := range callNodes {
		nameNode := node.ChildByFieldName("name")
		objNode := node.ChildByFieldName("object")

		if nameNode == nil {
			continue
		}

		methodName := analyzer.GetNodeText(nameNode, source)
		objName := ""
		if objNode != nil {
			objName = analyzer.GetNodeText(objNode, source)
		}

		if sourceType, ok := a.inputMethods[methodName]; ok {
			// Check if it's likely a request object or scanner/reader
			objLower := strings.ToLower(objName)
			isInputSource := strings.Contains(objLower, "request") ||
				strings.Contains(objLower, "req") ||
				strings.Contains(objName, "System") ||
				strings.Contains(objLower, "scanner") ||
				strings.Contains(objLower, "reader") ||
				strings.Contains(objLower, "input") ||
				strings.Contains(objLower, "context") ||
				strings.Contains(objLower, "params") ||
				strings.Contains(objLower, "body") ||
				strings.Contains(objLower, "files") ||
				objName == "" // Static method call

			if isInputSource {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "java",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       objName + "." + methodName,
					Snippet:    analyzer.GetNodeText(node, source),
					SourceType: sourceType,
				}
				sources = append(sources, flowNode)
			}
		}
	}

	// Array access for args[]
	arrayAccessNodes := analyzer.FindNodesOfType(root, "array_access")
	for _, node := range arrayAccessNodes {
		arrayNode := node.ChildByFieldName("array")
		if arrayNode != nil {
			arrayName := analyzer.GetNodeText(arrayNode, source)
			if arrayName == "args" {
				flowNode := &types.FlowNode{
					ID:         analyzer.GenerateNodeID("", node),
					Type:       types.NodeSource,
					Language:   "java",
					Line:       int(node.StartPoint().Row) + 1,
					Column:     int(node.StartPoint().Column),
					Name:       "args[]",
					Snippet:    analyzer.GetNodeText(node, source),
					SourceType: types.SourceCLIArg,
				}
				sources = append(sources, flowNode)
			}
		}
	}

	return sources, nil
}

// findSpringAnnotations finds Spring parameter annotations like @RequestParam, @PathVariable, etc.
func (a *JavaAnalyzer) findSpringAnnotations(root *sitter.Node, source []byte) []*types.FlowNode {
	var sources []*types.FlowNode

	// Find both marker_annotation (no args) and annotation (with args)
	annotationNodes := analyzer.FindNodesOfTypes(root, []string{"marker_annotation", "annotation"})
	for _, node := range annotationNodes {
		// Get the annotation name (identifier child)
		var annotationName string
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "identifier" {
				annotationName = analyzer.GetNodeText(child, source)
				break
			}
		}

		if annotationName == "" {
			continue
		}

		// Check if this is a known input annotation (Spring, JAX-RS, etc.)
		sourceType, isInputAnnotation := inputAnnotations[annotationName]
		if !isInputAnnotation {
			continue
		}

		// Only detect annotations on parameters (formal_parameter), not on methods
		paramName, isOnParameter := a.getAnnotatedParameterInfo(node, source)
		if !isOnParameter {
			continue // Skip annotations that are not on parameters
		}

		flowNode := &types.FlowNode{
			ID:         analyzer.GenerateNodeID("", node),
			Type:       types.NodeSource,
			Language:   "java",
			Line:       int(node.StartPoint().Row) + 1,
			Column:     int(node.StartPoint().Column),
			Name:       "@" + annotationName + " " + paramName,
			Snippet:    analyzer.GetNodeText(node, source),
			SourceType: sourceType,
		}
		sources = append(sources, flowNode)
	}

	return sources
}

// getAnnotatedParameterInfo extracts the parameter name from an annotation's parent formal_parameter
// Returns the parameter name and true if the annotation is on a parameter, empty string and false otherwise
func (a *JavaAnalyzer) getAnnotatedParameterInfo(annotationNode *sitter.Node, source []byte) (string, bool) {
	// Walk up to find the formal_parameter node
	// Annotation is inside: formal_parameter -> modifiers -> marker_annotation/annotation
	parent := annotationNode.Parent()
	for parent != nil {
		if parent.Type() == "formal_parameter" {
			// Get the parameter name (identifier child, but not type_identifier)
			for i := 0; i < int(parent.ChildCount()); i++ {
				child := parent.Child(i)
				if child.Type() == "identifier" {
					return analyzer.GetNodeText(child, source), true
				}
			}
			return "", true // On a parameter but couldn't extract name
		}
		// Stop if we hit a method or class declaration (annotation is not on a parameter)
		if parent.Type() == "method_declaration" || parent.Type() == "class_declaration" ||
			parent.Type() == "constructor_declaration" {
			return "", false
		}
		parent = parent.Parent()
	}
	return "", false
}

func (a *JavaAnalyzer) DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error) {
	var frameworks []string

	for _, imp := range symbolTable.Imports {
		path := strings.ToLower(imp.Path)

		if strings.Contains(path, "javax.servlet") || strings.Contains(path, "jakarta.servlet") {
			if !contains(frameworks, "servlet") {
				frameworks = append(frameworks, "servlet")
			}
		}
		if strings.Contains(path, "springframework") || strings.Contains(path, "org.springframework") {
			if !contains(frameworks, "spring") {
				frameworks = append(frameworks, "spring")
			}
		}
		if strings.Contains(path, "javax.ws.rs") || strings.Contains(path, "jakarta.ws.rs") {
			if !contains(frameworks, "jaxrs") {
				frameworks = append(frameworks, "jaxrs")
			}
		}
	}

	return frameworks, nil
}

func (a *JavaAnalyzer) AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*analyzer.MethodFlowAnalysis, error) {
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

func (a *JavaAnalyzer) TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error) {
	flowMap := types.NewFlowMap()
	flowMap.Target = target

	expr := target.Expression

	for method, sourceType := range a.inputMethods {
		if strings.Contains(expr, "."+method+"(") {
			sourceNode := types.FlowNode{
				ID:         fmt.Sprintf("source-%s", method),
				Type:       types.NodeSource,
				Language:   "java",
				Name:       method,
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
			Language:   "java",
			Name:       "args",
			Snippet:    expr,
			SourceType: types.SourceCLIArg,
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
	analyzer.DefaultRegistry.Register(NewJavaAnalyzer())
}
