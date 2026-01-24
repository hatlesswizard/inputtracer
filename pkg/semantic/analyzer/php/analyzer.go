// Package php implements the PHP language analyzer for semantic input tracing
package php

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// PHPAnalyzer implements the LanguageAnalyzer interface for PHP
type PHPAnalyzer struct {
	*analyzer.BaseAnalyzer
	superglobals    map[string]types.SourceType
	inputFunctions  map[string]types.SourceType
	dbFetchFunctions map[string]bool

	// Universal pattern-based detection (compiled regexes for performance)
	inputMethodPattern    *regexp.Regexp // Methods that return user input
	inputPropertyPattern  *regexp.Regexp // Properties that hold user input
	inputObjectPattern    *regexp.Regexp // Object/variable names that suggest input carrier
	excludeMethodPattern  *regexp.Regexp // Methods to exclude (database queries, etc.)
}

// NewPHPAnalyzer creates a new PHP analyzer
func NewPHPAnalyzer() *PHPAnalyzer {
	a := &PHPAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer("php", languages.GetExtensionsForLanguage("php")),
	}

	// Initialize PHP superglobals
	a.superglobals = map[string]types.SourceType{
		"$_GET":     types.SourceHTTPGet,
		"$_POST":    types.SourceHTTPPost,
		"$_REQUEST": types.SourceHTTPGet, // Can be GET, POST, or COOKIE
		"$_COOKIE":  types.SourceHTTPCookie,
		"$_SERVER":  types.SourceHTTPHeader,
		"$_FILES":   types.SourceHTTPBody,
		"$_ENV":     types.SourceEnvVar,
		"$_SESSION": types.SourceUserInput,
	}

	// Initialize input functions
	a.inputFunctions = map[string]types.SourceType{
		"file_get_contents": types.SourceFile,
		"fgets":             types.SourceFile,
		"fread":             types.SourceFile,
		"fgetc":             types.SourceFile,
		"fgetcsv":           types.SourceFile,
		"file":              types.SourceFile,
		"readfile":          types.SourceFile,
		"getenv":            types.SourceEnvVar,
		"apache_getenv":     types.SourceEnvVar,
		"getallheaders":     types.SourceHTTPHeader,
		"apache_request_headers": types.SourceHTTPHeader,
	}

	// Database fetch functions that return user-controlled data
	a.dbFetchFunctions = map[string]bool{
		"mysqli_fetch_array":  true,
		"mysqli_fetch_assoc":  true,
		"mysqli_fetch_row":    true,
		"mysqli_fetch_object": true,
		"mysqli_fetch_all":    true,
		"mysql_fetch_array":   true,
		"mysql_fetch_assoc":   true,
		"mysql_fetch_row":     true,
		"mysql_fetch_object":  true,
		"pg_fetch_array":      true,
		"pg_fetch_assoc":      true,
		"pg_fetch_row":        true,
		"pg_fetch_object":     true,
		"pg_fetch_all":        true,
		"sqlite_fetch_array":  true,
		"oci_fetch_array":     true,
		"oci_fetch_assoc":     true,
		"oci_fetch_row":       true,
		"db_fetch_array":      true,
	}

	// Initialize universal pattern-based detection
	// These patterns detect input sources across ALL PHP frameworks generically
	a.initUniversalPatterns()

	// Register framework patterns (kept for backward compatibility)
	a.registerFrameworkPatterns()

	return a
}

// initUniversalPatterns initializes regex patterns for universal input detection
// This approach detects user input sources across ANY PHP framework without
// requiring framework-specific hardcoding
func (a *PHPAnalyzer) initUniversalPatterns() {
	// High-confidence method names that ALWAYS indicate user input
	// These are specific enough to detect without checking the object name
	// Pattern matches:
	// - Explicit input getters: input, get_input, getInput, get_var, variable
	// - HTTP method getters: getPost, getQuery, getCookie, getHeader, etc.
	// - PSR-7 methods: getQueryParams, getParsedBody, getCookieParams, etc.
	// - All input: all()
	a.inputMethodPattern = regexp.MustCompile(`(?i)^(get_?)?(input|var|variable|query_?params?|parsed_?body|cookie_?params?|server_?params?|uploaded_?files?|headers?|all)$|^(get_?)?(post|cookie|param)s?$`)

	// Property names that typically hold user input (for array access patterns like ->input['key'])
	// Matches: input, request, params, query, cookies, headers, body, data, args, post, get, files, server
	// These are properties that typically store user data in framework objects
	a.inputPropertyPattern = regexp.MustCompile(`(?i)^(input|request|params?|query|cookies?|headers?|body|data|args?|post|get|files?|server|attributes?|payload)s?$`)

	// Object/variable names that suggest the object is an input carrier
	// These are common names for request objects across frameworks
	// Also matches chain calls like "->getRequest()" or "Factory::getApplication()->getInput()"
	a.inputObjectPattern = regexp.MustCompile(`(?i)(request|input|req|params?|http|ctx|context|mybb|getRequest\(\)|getApplication\(\))`)

	// Method names to EXCLUDE from input detection (false positive prevention)
	// These are methods that might match patterns but aren't typically user input
	a.excludeMethodPattern = regexp.MustCompile(`(?i)^(getData|getBody|getContent|fetch|find|load|read)$`)
}

// isContextDependentInputMethod returns true if the method name is a generic getter
// that should only be detected as user input when called on a request-like object
func (a *PHPAnalyzer) isContextDependentInputMethod(methodName string) bool {
	// Methods like getVal, getText, getInt, getBool are used in MediaWiki on request objects
	// but are also used on many other objects (Title, Message, etc.)
	// Only detect these when the object looks like a request
	contextDependent := regexp.MustCompile(`(?i)^(get_?)?(val|text|int|bool|array|raw_?val|check)$`)
	return contextDependent.MatchString(methodName)
}

// registerFrameworkPatterns registers known PHP framework patterns
func (a *PHPAnalyzer) registerFrameworkPatterns() {
	// MyBB pattern
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "mybb_input",
		Framework:       "mybb",
		Language:        "php",
		Name:            "MyBB $mybb->input",
		Description:     "MyBB input array populated from $_GET and $_POST via parse_incoming()",
		ClassPattern:    "^MyBB$",
		PropertyPattern: "^input$",
		AccessPattern:   "array",
		SourceType:      types.SourceHTTPGet, // Actually GET+POST
		CarrierClass:    "MyBB",
		CarrierProperty: "input",
		PopulatedBy:     "__construct",
		PopulatedFrom:   []string{"$_GET", "$_POST"},
		Confidence:      1.0,
	})

	// MyBB cookies pattern
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "mybb_cookies",
		Framework:       "mybb",
		Language:        "php",
		Name:            "MyBB $mybb->cookies",
		Description:     "MyBB cookies array populated from $_COOKIE via parse_cookies()",
		ClassPattern:    "^MyBB$",
		PropertyPattern: "^cookies$",
		AccessPattern:   "array",
		SourceType:      types.SourceHTTPCookie,
		CarrierClass:    "MyBB",
		CarrierProperty: "cookies",
		PopulatedBy:     "parse_cookies",
		PopulatedFrom:   []string{"$_COOKIE"},
		Confidence:      1.0,
	})

	// WordPress $_REQUEST
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "wordpress_request",
		Framework:       "wordpress",
		Language:        "php",
		Name:            "WordPress $_REQUEST",
		Description:     "WordPress uses $_REQUEST which combines $_GET, $_POST, and $_COOKIE",
		AccessPattern:   "superglobal",
		SourceType:      types.SourceHTTPGet,
		PopulatedFrom:   []string{"$_GET", "$_POST", "$_COOKIE"},
		Confidence:      1.0,
	})

	// Laravel Request
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "laravel_request_input",
		Framework:       "laravel",
		Language:        "php",
		Name:            "Laravel $request->input()",
		Description:     "Laravel request input method returns GET and POST data",
		ClassPattern:    "^(Illuminate\\\\Http\\\\)?Request$",
		MethodPattern:   "^input$",
		SourceType:      types.SourceHTTPGet,
		CarrierClass:    "Illuminate\\Http\\Request",
		PopulatedFrom:   []string{"$_GET", "$_POST"},
		Confidence:      0.95,
	})

	// Laravel Request all()
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "laravel_request_all",
		Framework:       "laravel",
		Language:        "php",
		Name:            "Laravel $request->all()",
		Description:     "Laravel request all() returns all input data",
		ClassPattern:    "^(Illuminate\\\\Http\\\\)?Request$",
		MethodPattern:   "^all$",
		SourceType:      types.SourceHTTPGet,
		CarrierClass:    "Illuminate\\Http\\Request",
		PopulatedFrom:   []string{"$_GET", "$_POST"},
		Confidence:      0.95,
	})

	// Symfony Request
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "symfony_request_query",
		Framework:       "symfony",
		Language:        "php",
		Name:            "Symfony $request->query",
		Description:     "Symfony request query bag contains GET parameters",
		ClassPattern:    "^(Symfony\\\\Component\\\\HttpFoundation\\\\)?Request$",
		PropertyPattern: "^query$",
		SourceType:      types.SourceHTTPGet,
		CarrierClass:    "Symfony\\Component\\HttpFoundation\\Request",
		CarrierProperty: "query",
		PopulatedFrom:   []string{"$_GET"},
		Confidence:      0.95,
	})

	// Symfony Request POST
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "symfony_request_request",
		Framework:       "symfony",
		Language:        "php",
		Name:            "Symfony $request->request",
		Description:     "Symfony request bag contains POST parameters",
		ClassPattern:    "^(Symfony\\\\Component\\\\HttpFoundation\\\\)?Request$",
		PropertyPattern: "^request$",
		SourceType:      types.SourceHTTPPost,
		CarrierClass:    "Symfony\\Component\\HttpFoundation\\Request",
		CarrierProperty: "request",
		PopulatedFrom:   []string{"$_POST"},
		Confidence:      0.95,
	})

	// CodeIgniter Input
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "codeigniter_input_get",
		Framework:       "codeigniter",
		Language:        "php",
		Name:            "CodeIgniter $this->input->get()",
		Description:     "CodeIgniter input class get() method",
		ClassPattern:    "^CI_Input$",
		MethodPattern:   "^get$",
		SourceType:      types.SourceHTTPGet,
		CarrierClass:    "CI_Input",
		PopulatedFrom:   []string{"$_GET"},
		Confidence:      0.9,
	})

	// CodeIgniter Input POST
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "codeigniter_input_post",
		Framework:       "codeigniter",
		Language:        "php",
		Name:            "CodeIgniter $this->input->post()",
		Description:     "CodeIgniter input class post() method",
		ClassPattern:    "^CI_Input$",
		MethodPattern:   "^post$",
		SourceType:      types.SourceHTTPPost,
		CarrierClass:    "CI_Input",
		PopulatedFrom:   []string{"$_POST"},
		Confidence:      0.9,
	})

	// PDO fetch methods
	a.AddFrameworkPattern(&types.FrameworkPattern{
		ID:              "pdo_fetch",
		Framework:       "pdo",
		Language:        "php",
		Name:            "PDOStatement->fetch()",
		Description:     "PDO fetch returns database data (potentially user-originated)",
		ClassPattern:    "^PDOStatement$",
		MethodPattern:   "^fetch(All)?$",
		SourceType:      types.SourceDatabase,
		Confidence:      0.7,
	})
}

// BuildSymbolTable builds the symbol table for a PHP file
func (a *PHPAnalyzer) BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error) {
	st := types.NewSymbolTable(filePath, "php")

	// Extract namespace
	st.Namespace = a.extractNamespace(root, source)

	// Extract imports (use statements)
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

// extractNamespace extracts the namespace from a PHP file
func (a *PHPAnalyzer) extractNamespace(root *sitter.Node, source []byte) string {
	nodes := analyzer.FindNodesOfType(root, "namespace_definition")
	if len(nodes) > 0 {
		nameNode := analyzer.FindChildByType(nodes[0], "namespace_name")
		if nameNode != nil {
			return analyzer.GetNodeText(nameNode, source)
		}
	}
	return ""
}

// extractImports extracts use statements from a PHP file
func (a *PHPAnalyzer) extractImports(root *sitter.Node, source []byte) []types.ImportInfo {
	var imports []types.ImportInfo

	// Include/require statements
	includeTypes := []string{"include_expression", "include_once_expression", "require_expression", "require_once_expression"}
	for _, incType := range includeTypes {
		nodes := analyzer.FindNodesOfType(root, incType)
		for _, node := range nodes {
			// Get the path being included
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child.Type() == "string" || child.Type() == "encapsed_string" {
					path := strings.Trim(analyzer.GetNodeText(child, source), "\"'")
					imports = append(imports, types.ImportInfo{
						Path:       path,
						Line:       int(node.StartPoint().Row) + 1,
						Type:       strings.ReplaceAll(incType, "_expression", ""),
						IsRelative: !strings.HasPrefix(path, "/") && !strings.Contains(path, "://"),
					})
				}
			}
		}
	}

	// Use statements
	useNodes := analyzer.FindNodesOfType(root, "namespace_use_declaration")
	for _, node := range useNodes {
		clauseNodes := analyzer.FindNodesOfType(node, "namespace_use_clause")
		for _, clause := range clauseNodes {
			nameNode := analyzer.FindChildByType(clause, "qualified_name")
			if nameNode == nil {
				nameNode = analyzer.FindChildByType(clause, "namespace_name")
			}
			if nameNode != nil {
				path := analyzer.GetNodeText(nameNode, source)
				alias := ""
				aliasNode := analyzer.FindChildByType(clause, "namespace_aliasing_clause")
				if aliasNode != nil {
					nameNode := analyzer.FindChildByType(aliasNode, "name")
					if nameNode != nil {
						alias = analyzer.GetNodeText(nameNode, source)
					}
				}
				imports = append(imports, types.ImportInfo{
					Path:  path,
					Alias: alias,
					Line:  int(node.StartPoint().Row) + 1,
					Type:  "use",
				})
			}
		}
	}

	return imports
}

// ResolveImports resolves import paths to actual file paths
func (a *PHPAnalyzer) ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error) {
	var resolvedPaths []string
	dir := filepath.Dir(symbolTable.FilePath)

	for _, imp := range symbolTable.Imports {
		if imp.Type == "use" {
			// Namespace use - would need autoloader knowledge
			continue
		}

		// Include/require - resolve relative path
		var resolvedPath string
		if imp.IsRelative {
			// Try relative to current file
			resolvedPath = filepath.Join(dir, imp.Path)
		} else {
			resolvedPath = imp.Path
		}

		// Clean the path
		resolvedPath = filepath.Clean(resolvedPath)

		// Handle dirname(__FILE__) and similar constructs
		resolvedPath = strings.ReplaceAll(resolvedPath, "dirname(__FILE__)", dir)
		resolvedPath = strings.ReplaceAll(resolvedPath, "__DIR__", dir)

		// Remove variable interpolation for now
		if !strings.Contains(resolvedPath, "$") {
			resolvedPaths = append(resolvedPaths, resolvedPath)
		}
	}

	return resolvedPaths, nil
}

// ExtractClasses extracts class definitions from PHP AST
func (a *PHPAnalyzer) ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error) {
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

// parseClassDeclaration parses a class declaration node
func (a *PHPAnalyzer) parseClassDeclaration(node *sitter.Node, source []byte) *types.ClassDef {
	// Get class name
	nameNode := analyzer.FindChildByType(node, "name")
	if nameNode == nil {
		return nil
	}
	name := analyzer.GetNodeText(nameNode, source)

	class := types.NewClassDef(name, "", int(node.StartPoint().Row)+1)
	class.EndLine = int(node.EndPoint().Row) + 1

	// Check for abstract/final modifiers
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "abstract_modifier":
			class.IsAbstract = true
		case "final_modifier":
			class.IsFinal = true
		}
	}

	// Get extends
	extendsNode := analyzer.FindChildByType(node, "base_clause")
	if extendsNode != nil {
		nameNode := analyzer.FindChildByType(extendsNode, "name")
		if nameNode == nil {
			nameNode = analyzer.FindChildByType(extendsNode, "qualified_name")
		}
		if nameNode != nil {
			class.Extends = analyzer.GetNodeText(nameNode, source)
		}
	}

	// Get implements
	implementsNode := analyzer.FindChildByType(node, "class_interface_clause")
	if implementsNode != nil {
		nameNodes := analyzer.FindNodesOfType(implementsNode, "name")
		for _, n := range nameNodes {
			class.Implements = append(class.Implements, analyzer.GetNodeText(n, source))
		}
	}

	// Get body
	bodyNode := analyzer.FindChildByType(node, "declaration_list")
	if bodyNode != nil {
		a.parseClassBody(class, bodyNode, source)
	}

	// Check if this is a carrier class (based on framework patterns)
	a.checkIfCarrier(class)

	return class
}

// parseClassBody parses the body of a class
func (a *PHPAnalyzer) parseClassBody(class *types.ClassDef, bodyNode *sitter.Node, source []byte) {
	for i := 0; i < int(bodyNode.ChildCount()); i++ {
		child := bodyNode.Child(i)
		switch child.Type() {
		case "property_declaration":
			props := a.parsePropertyDeclaration(child, source)
			for _, prop := range props {
				class.Properties[prop.Name] = prop
			}
		case "method_declaration":
			method := a.parseMethodDeclaration(child, source)
			if method != nil {
				if method.Name == "__construct" {
					class.Constructor = method
				}
				class.Methods[method.Name] = method
			}
		case "use_declaration":
			// Traits
			names := analyzer.FindNodesOfType(child, "name")
			for _, n := range names {
				class.Traits = append(class.Traits, analyzer.GetNodeText(n, source))
			}
		}
	}
}

// parsePropertyDeclaration parses a property declaration
func (a *PHPAnalyzer) parsePropertyDeclaration(node *sitter.Node, source []byte) []*types.PropertyDef {
	var props []*types.PropertyDef

	// Get visibility
	visibility := "public"
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "visibility_modifier":
			visibility = analyzer.GetNodeText(child, source)
		case "static_modifier":
			// Will handle below
		}
	}

	// Get property elements
	propElements := analyzer.FindNodesOfType(node, "property_element")
	for _, elem := range propElements {
		nameNode := analyzer.FindChildByType(elem, "variable_name")
		if nameNode == nil {
			continue
		}

		name := analyzer.GetNodeText(nameNode, source)
		name = strings.TrimPrefix(name, "$")

		prop := &types.PropertyDef{
			Name:       name,
			Visibility: visibility,
			Line:       int(elem.StartPoint().Row) + 1,
		}

		// Check for static
		for i := 0; i < int(node.ChildCount()); i++ {
			if node.Child(i).Type() == "static_modifier" {
				prop.IsStatic = true
				break
			}
		}

		// Get initial value
		initNode := analyzer.FindChildByType(elem, "property_initializer")
		if initNode != nil {
			for i := 0; i < int(initNode.ChildCount()); i++ {
				child := initNode.Child(i)
				if child.Type() != "=" {
					prop.InitialValue = analyzer.GetNodeText(child, source)
					break
				}
			}
		}

		// Get type if specified
		typeNode := analyzer.FindChildByType(node, "type")
		if typeNode == nil {
			typeNode = analyzer.FindChildByType(node, "union_type")
		}
		if typeNode != nil {
			prop.Type = analyzer.GetNodeText(typeNode, source)
		}

		props = append(props, prop)
	}

	return props
}

// parseMethodDeclaration parses a method declaration
func (a *PHPAnalyzer) parseMethodDeclaration(node *sitter.Node, source []byte) *types.MethodDef {
	// Get method name
	nameNode := analyzer.FindChildByType(node, "name")
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

	// Get visibility and modifiers
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "visibility_modifier":
			method.Visibility = analyzer.GetNodeText(child, source)
		case "static_modifier":
			method.IsStatic = true
		case "abstract_modifier":
			method.IsAbstract = true
		}
	}

	// Get parameters
	paramsNode := analyzer.FindChildByType(node, "formal_parameters")
	if paramsNode != nil {
		method.Parameters = a.parseParameters(paramsNode, source)
	}

	// Get return type
	returnNode := analyzer.FindChildByType(node, "return_type")
	if returnNode != nil {
		typeNode := analyzer.FindChildByType(returnNode, "type")
		if typeNode == nil {
			typeNode = analyzer.FindChildByType(returnNode, "union_type")
		}
		if typeNode != nil {
			method.ReturnType = analyzer.GetNodeText(typeNode, source)
		}
	}

	// Get body
	bodyNode := analyzer.FindChildByType(node, "compound_statement")
	if bodyNode != nil {
		method.BodyStart = int(bodyNode.StartPoint().Row) + 1
		method.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		method.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	return method
}

// parseParameters parses function/method parameters
func (a *PHPAnalyzer) parseParameters(node *sitter.Node, source []byte) []types.ParameterDef {
	var params []types.ParameterDef

	paramNodes := analyzer.FindNodesOfType(node, "simple_parameter")
	for i, paramNode := range paramNodes {
		param := types.ParameterDef{
			Index: i,
		}

		// Get name
		nameNode := analyzer.FindChildByType(paramNode, "variable_name")
		if nameNode != nil {
			param.Name = strings.TrimPrefix(analyzer.GetNodeText(nameNode, source), "$")
		}

		// Get type
		typeNode := analyzer.FindChildByType(paramNode, "type")
		if typeNode == nil {
			typeNode = analyzer.FindChildByType(paramNode, "union_type")
		}
		if typeNode != nil {
			param.Type = analyzer.GetNodeText(typeNode, source)
		}

		// Check for reference
		for j := 0; j < int(paramNode.ChildCount()); j++ {
			child := paramNode.Child(j)
			if child.Type() == "reference_modifier" {
				param.IsReference = true
			}
		}

		// Get default value
		defaultNode := analyzer.FindChildByType(paramNode, "default_value")
		if defaultNode != nil {
			for j := 0; j < int(defaultNode.ChildCount()); j++ {
				child := defaultNode.Child(j)
				if child.Type() != "=" {
					param.DefaultValue = analyzer.GetNodeText(child, source)
					break
				}
			}
		}

		params = append(params, param)
	}

	// Check for variadic parameters
	variadicNodes := analyzer.FindNodesOfType(node, "variadic_parameter")
	for _, paramNode := range variadicNodes {
		param := types.ParameterDef{
			Index:      len(params),
			IsVariadic: true,
		}

		nameNode := analyzer.FindChildByType(paramNode, "variable_name")
		if nameNode != nil {
			param.Name = strings.TrimPrefix(analyzer.GetNodeText(nameNode, source), "$")
		}

		params = append(params, param)
	}

	return params
}

// checkIfCarrier checks if a class is a known input carrier
func (a *PHPAnalyzer) checkIfCarrier(class *types.ClassDef) {
	for _, pattern := range a.GetFrameworkPatterns() {
		if pattern.ClassPattern != "" {
			matched, _ := regexp.MatchString(pattern.ClassPattern, class.Name)
			if matched {
				class.IsCarrier = true
				class.CarrierInfo = &types.CarrierInfo{
					PropertyName:      pattern.CarrierProperty,
					SourceTypes:       pattern.PopulatedFrom,
					PopulationMethod:  pattern.PopulatedBy,
					PopulationPattern: pattern.AccessPattern,
				}
				return
			}
		}
	}
}

// ExtractFunctions extracts standalone function definitions
func (a *PHPAnalyzer) ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error) {
	var functions []*types.FunctionDef

	funcNodes := analyzer.FindNodesOfType(root, "function_definition")
	for _, funcNode := range funcNodes {
		// Skip if inside a class
		if analyzer.GetEnclosingClass(funcNode, []string{"class_declaration"}) != nil {
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
func (a *PHPAnalyzer) parseFunctionDefinition(node *sitter.Node, source []byte) *types.FunctionDef {
	nameNode := analyzer.FindChildByType(node, "name")
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

	// Get return type
	returnNode := analyzer.FindChildByType(node, "return_type")
	if returnNode != nil {
		typeNode := analyzer.FindChildByType(returnNode, "type")
		if typeNode != nil {
			fn.ReturnType = analyzer.GetNodeText(typeNode, source)
		}
	}

	// Get body
	bodyNode := analyzer.FindChildByType(node, "compound_statement")
	if bodyNode != nil {
		fn.BodyStart = int(bodyNode.StartPoint().Row) + 1
		fn.BodyEnd = int(bodyNode.EndPoint().Row) + 1
		fn.BodySource = analyzer.GetNodeText(bodyNode, source)
	}

	return fn
}

// ExtractAssignments extracts all assignments from the AST
func (a *PHPAnalyzer) ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error) {
	var assignments []*types.Assignment

	// Find assignment expressions
	assignNodes := analyzer.FindNodesOfType(root, "assignment_expression")
	for _, node := range assignNodes {
		assignment := a.parseAssignment(node, source, scope)
		if assignment != nil {
			assignments = append(assignments, assignment)
		}
	}

	// Find augmented assignments (+=, .=, etc.)
	augmentedNodes := analyzer.FindNodesOfType(root, "augmented_assignment_expression")
	for _, node := range augmentedNodes {
		assignment := a.parseAssignment(node, source, scope)
		if assignment != nil {
			assignments = append(assignments, assignment)
		}
	}

	return assignments, nil
}

// parseAssignment parses an assignment expression
func (a *PHPAnalyzer) parseAssignment(node *sitter.Node, source []byte, scope string) *types.Assignment {
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
	case "variable_name":
		assignment.TargetType = "variable"
	case "member_access_expression":
		assignment.TargetType = "property"
	case "subscript_expression":
		assignment.TargetType = "array_element"
		assignment.Keys = a.extractArrayKeys(leftNode, source)
	}

	// Check if source is tainted
	assignment.IsTainted, assignment.TaintSource = a.isExpressionTainted(rightNode, source)

	return assignment
}

// extractArrayKeys extracts the access keys from a subscript expression
func (a *PHPAnalyzer) extractArrayKeys(node *sitter.Node, source []byte) []string {
	var keys []string

	// Walk up the subscript chain
	current := node
	for current != nil && current.Type() == "subscript_expression" {
		// Get the index
		for i := 0; i < int(current.ChildCount()); i++ {
			child := current.Child(i)
			if child.Type() == "string" || child.Type() == "encapsed_string" ||
				child.Type() == "integer" || child.Type() == "name" {
				key := analyzer.GetNodeText(child, source)
				key = strings.Trim(key, "\"'")
				keys = append([]string{key}, keys...) // Prepend
			}
		}

		// Get the base
		baseNode := current.Child(0)
		if baseNode != nil && baseNode.Type() == "subscript_expression" {
			current = baseNode
		} else {
			// Add the base variable name
			if baseNode != nil {
				baseName := analyzer.GetNodeText(baseNode, source)
				keys = append([]string{baseName}, keys...)
			}
			break
		}
	}

	return keys
}

// isExpressionTainted checks if an expression contains tainted data
func (a *PHPAnalyzer) isExpressionTainted(node *sitter.Node, source []byte) (bool, string) {
	if node == nil {
		return false, ""
	}

	text := analyzer.GetNodeText(node, source)

	// Check for superglobals
	for sg, _ := range a.superglobals {
		if strings.Contains(text, sg) {
			return true, sg
		}
	}

	// Check for input functions
	for fn, _ := range a.inputFunctions {
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
func (a *PHPAnalyzer) ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error) {
	var calls []*types.CallSite

	// Function calls
	funcCallNodes := analyzer.FindNodesOfType(root, "function_call_expression")
	for _, node := range funcCallNodes {
		call := a.parseFunctionCall(node, source, scope)
		if call != nil {
			calls = append(calls, call)
		}
	}

	// Method calls
	methodCallNodes := analyzer.FindNodesOfType(root, "member_call_expression")
	for _, node := range methodCallNodes {
		call := a.parseMethodCall(node, source, scope)
		if call != nil {
			calls = append(calls, call)
		}
	}

	// Static method calls
	staticCallNodes := analyzer.FindNodesOfType(root, "scoped_call_expression")
	for _, node := range staticCallNodes {
		call := a.parseStaticCall(node, source, scope)
		if call != nil {
			calls = append(calls, call)
		}
	}

	// Object creation
	newNodes := analyzer.FindNodesOfType(root, "object_creation_expression")
	for _, node := range newNodes {
		call := a.parseObjectCreation(node, source, scope)
		if call != nil {
			calls = append(calls, call)
		}
	}

	return calls, nil
}

// parseFunctionCall parses a function call expression
func (a *PHPAnalyzer) parseFunctionCall(node *sitter.Node, source []byte, scope string) *types.CallSite {
	// Get function name
	nameNode := node.Child(0)
	if nameNode == nil {
		return nil
	}

	call := &types.CallSite{
		FunctionName: analyzer.GetNodeText(nameNode, source),
		Line:         int(node.StartPoint().Row) + 1,
		Column:       int(node.StartPoint().Column),
		Scope:        scope,
		Arguments:    make([]types.CallArg, 0),
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

// parseMethodCall parses a method call expression
func (a *PHPAnalyzer) parseMethodCall(node *sitter.Node, source []byte, scope string) *types.CallSite {
	call := &types.CallSite{
		Line:      int(node.StartPoint().Row) + 1,
		Column:    int(node.StartPoint().Column),
		Scope:     scope,
		Arguments: make([]types.CallArg, 0),
	}

	// Get object and method
	objNode := node.Child(0)
	if objNode != nil {
		call.ClassName = analyzer.GetNodeText(objNode, source)
	}

	nameNode := analyzer.FindChildByType(node, "name")
	if nameNode != nil {
		call.MethodName = analyzer.GetNodeText(nameNode, source)
		call.FunctionName = call.ClassName + "->" + call.MethodName
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

// parseStaticCall parses a static method call
func (a *PHPAnalyzer) parseStaticCall(node *sitter.Node, source []byte, scope string) *types.CallSite {
	call := &types.CallSite{
		Line:      int(node.StartPoint().Row) + 1,
		Column:    int(node.StartPoint().Column),
		Scope:     scope,
		IsStatic:  true,
		Arguments: make([]types.CallArg, 0),
	}

	// Get class name
	scopeNode := analyzer.FindChildByType(node, "scope_resolution_qualifier")
	if scopeNode == nil {
		scopeNode = node.Child(0)
	}
	if scopeNode != nil {
		call.ClassName = analyzer.GetNodeText(scopeNode, source)
	}

	// Get method name
	nameNode := analyzer.FindChildByType(node, "name")
	if nameNode != nil {
		call.MethodName = analyzer.GetNodeText(nameNode, source)
		call.FunctionName = call.ClassName + "::" + call.MethodName
	}

	// Get arguments
	argsNode := analyzer.FindChildByType(node, "arguments")
	if argsNode != nil {
		call.Arguments = a.parseCallArguments(argsNode, source)
	}

	return call
}

// parseObjectCreation parses a new expression
func (a *PHPAnalyzer) parseObjectCreation(node *sitter.Node, source []byte, scope string) *types.CallSite {
	call := &types.CallSite{
		Line:          int(node.StartPoint().Row) + 1,
		Column:        int(node.StartPoint().Column),
		Scope:         scope,
		IsConstructor: true,
		Arguments:     make([]types.CallArg, 0),
	}

	// Get class name
	nameNode := analyzer.FindChildByType(node, "name")
	if nameNode == nil {
		nameNode = analyzer.FindChildByType(node, "qualified_name")
	}
	if nameNode != nil {
		call.ClassName = analyzer.GetNodeText(nameNode, source)
		call.FunctionName = "new " + call.ClassName
		call.MethodName = "__construct"
	}

	// Get arguments
	argsNode := analyzer.FindChildByType(node, "arguments")
	if argsNode != nil {
		call.Arguments = a.parseCallArguments(argsNode, source)
	}

	return call
}

// parseCallArguments parses function call arguments
func (a *PHPAnalyzer) parseCallArguments(node *sitter.Node, source []byte) []types.CallArg {
	var args []types.CallArg

	argNodes := analyzer.FindNodesOfType(node, "argument")
	for i, argNode := range argNodes {
		arg := types.CallArg{
			Index: i,
			Value: analyzer.GetNodeText(argNode, source),
		}

		// Check if tainted
		arg.IsTainted, arg.TaintSource = a.isExpressionTainted(argNode, source)

		args = append(args, arg)
	}

	return args
}

// FindInputSources finds all user input sources in the AST
func (a *PHPAnalyzer) FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error) {
	var sources []*types.FlowNode

	// Find superglobal accesses
	varNodes := analyzer.FindNodesOfType(root, "variable_name")
	for _, node := range varNodes {
		text := analyzer.GetNodeText(node, source)
		if sourceType, ok := a.superglobals[text]; ok {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "php",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       text,
				Snippet:    text,
				SourceType: sourceType,
			}

			// Check if it's a subscript access to get the key
			parent := node.Parent()
			if parent != nil && parent.Type() == "subscript_expression" {
				// Get the full expression
				flowNode.Snippet = analyzer.GetNodeText(parent, source)

				// Extract the key
				for i := 0; i < int(parent.ChildCount()); i++ {
					child := parent.Child(i)
					if child.Type() == "string" || child.Type() == "encapsed_string" {
						key := analyzer.GetNodeText(child, source)
						flowNode.SourceKey = strings.Trim(key, "\"'")
						break
					}
				}
			}

			sources = append(sources, flowNode)
		}
	}

	// Find input function calls
	funcCalls := analyzer.FindNodesOfType(root, "function_call_expression")
	for _, node := range funcCalls {
		nameNode := node.Child(0)
		if nameNode == nil {
			continue
		}
		funcName := analyzer.GetNodeText(nameNode, source)

		if sourceType, ok := a.inputFunctions[funcName]; ok {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "php",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       funcName,
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: sourceType,
			}

			// For getenv, extract the variable name
			if funcName == "getenv" {
				argsNode := analyzer.FindChildByType(node, "arguments")
				if argsNode != nil {
					for i := 0; i < int(argsNode.ChildCount()); i++ {
						child := argsNode.Child(i)
						if child.Type() == "argument" {
							arg := analyzer.GetNodeText(child, source)
							flowNode.SourceKey = strings.Trim(arg, "\"'")
							break
						}
					}
				}
			}

			sources = append(sources, flowNode)
		}

		// Check for database fetch functions
		if a.dbFetchFunctions[funcName] {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "php",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       funcName,
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: types.SourceDatabase,
			}
			sources = append(sources, flowNode)
		}
	}

	// Find method calls that are sources (PDO fetch, etc.)
	methodCalls := analyzer.FindNodesOfType(root, "member_call_expression")
	for _, node := range methodCalls {
		nameNode := analyzer.FindChildByType(node, "name")
		if nameNode == nil {
			continue
		}
		methodName := analyzer.GetNodeText(nameNode, source)

		// PDO fetch methods
		if methodName == "fetch" || methodName == "fetchAll" || methodName == "fetchColumn" {
			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "php",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "PDOStatement->" + methodName,
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: types.SourceDatabase,
			}
			sources = append(sources, flowNode)
		}

		// Universal pattern-based method detection
		// This detects ANY method that looks like an input getter, not just specific frameworks
		isInputMethod := a.inputMethodPattern.MatchString(methodName)
		isContextDependent := a.isContextDependentInputMethod(methodName)

		// For context-dependent methods (like getVal, getText), check if the object
		// looks like a request carrier before flagging as user input
		if isContextDependent && !isInputMethod {
			// Get the object being called on
			objNode := node.Child(0)
			if objNode != nil {
				objText := analyzer.GetNodeText(objNode, source)
				// Only detect if the object looks like a request carrier
				if a.inputObjectPattern.MatchString(objText) {
					isInputMethod = true
				}
			}
		}

		if isInputMethod {
			// Determine source type based on method name hints
			sourceType := a.inferSourceTypeFromMethodName(methodName)

			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "php",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "->" + methodName + "()",
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: sourceType,
			}

			// Extract the key from first argument if present
			argsNode := analyzer.FindChildByType(node, "arguments")
			if argsNode != nil {
				for i := 0; i < int(argsNode.ChildCount()); i++ {
					child := argsNode.Child(i)
					if child.Type() == "argument" {
						arg := analyzer.GetNodeText(child, source)
						flowNode.SourceKey = strings.Trim(arg, "\"'")
						break
					}
				}
			}

			sources = append(sources, flowNode)
		}
	}

	// Find carrier property array access ($obj->input['key'], $obj->data['key'], etc.)
	// Universal pattern-based property detection
	subscriptNodes := analyzer.FindNodesOfType(root, "subscript_expression")
	for _, node := range subscriptNodes {
		// Check if the base is a member_access_expression
		baseNode := node.Child(0)
		if baseNode == nil || baseNode.Type() != "member_access_expression" {
			continue
		}

		// Get the property name being accessed
		propNameNode := analyzer.FindChildByType(baseNode, "name")
		if propNameNode == nil {
			continue
		}
		propName := analyzer.GetNodeText(propNameNode, source)

		// Universal pattern-based property detection
		// Matches any property name that looks like it holds user input
		if a.inputPropertyPattern.MatchString(propName) {
			// Determine source type based on property name hints
			sourceType := a.inferSourceTypeFromPropertyName(propName)

			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "php",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       "->" + propName + "[]",
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: sourceType,
			}

			// Extract the key from the subscript
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child.Type() == "string" || child.Type() == "encapsed_string" {
					key := analyzer.GetNodeText(child, source)
					flowNode.SourceKey = strings.Trim(key, "\"'")
					break
				}
			}

			sources = append(sources, flowNode)
		}
	}

	// Additional: Detect direct property access on request-like objects
	// e.g., $request->query, $input->data (without array subscript)
	memberAccessNodes := analyzer.FindNodesOfType(root, "member_access_expression")
	for _, node := range memberAccessNodes {
		// Skip if this is a method call (has arguments)
		parent := node.Parent()
		if parent != nil && parent.Type() == "member_call_expression" {
			continue
		}
		// Skip if this is already part of a subscript expression (handled above)
		if parent != nil && parent.Type() == "subscript_expression" {
			continue
		}

		// Get the object being accessed
		objNode := node.Child(0)
		if objNode == nil {
			continue
		}
		objName := analyzer.GetNodeText(objNode, source)

		// Get the property name
		propNameNode := analyzer.FindChildByType(node, "name")
		if propNameNode == nil {
			continue
		}
		propName := analyzer.GetNodeText(propNameNode, source)

		// Check if the object name suggests it's an input carrier
		// AND the property looks like an input property
		if a.inputObjectPattern.MatchString(objName) && a.inputPropertyPattern.MatchString(propName) {
			sourceType := a.inferSourceTypeFromPropertyName(propName)

			flowNode := &types.FlowNode{
				ID:         analyzer.GenerateNodeID("", node),
				Type:       types.NodeSource,
				Language:   "php",
				Line:       int(node.StartPoint().Row) + 1,
				Column:     int(node.StartPoint().Column),
				Name:       objName + "->" + propName,
				Snippet:    analyzer.GetNodeText(node, source),
				SourceType: sourceType,
			}

			sources = append(sources, flowNode)
		}
	}

	return sources, nil
}

// inferSourceTypeFromMethodName determines the source type based on method name patterns
func (a *PHPAnalyzer) inferSourceTypeFromMethodName(methodName string) types.SourceType {
	lowerName := strings.ToLower(methodName)

	// Check for specific type hints in the method name
	switch {
	case strings.Contains(lowerName, "cookie"):
		return types.SourceHTTPCookie
	case strings.Contains(lowerName, "header"):
		return types.SourceHTTPHeader
	case strings.Contains(lowerName, "server"):
		return types.SourceHTTPHeader
	case strings.Contains(lowerName, "post") || strings.Contains(lowerName, "body") || strings.Contains(lowerName, "parsed"):
		return types.SourceHTTPPost
	case strings.Contains(lowerName, "query") || strings.Contains(lowerName, "get"):
		return types.SourceHTTPGet
	case strings.Contains(lowerName, "file") || strings.Contains(lowerName, "upload"):
		return types.SourceHTTPBody
	default:
		return types.SourceUserInput
	}
}

// inferSourceTypeFromPropertyName determines the source type based on property name patterns
func (a *PHPAnalyzer) inferSourceTypeFromPropertyName(propName string) types.SourceType {
	lowerName := strings.ToLower(propName)

	// Check for specific type hints in the property name
	switch {
	case strings.Contains(lowerName, "cookie"):
		return types.SourceHTTPCookie
	case strings.Contains(lowerName, "header"):
		return types.SourceHTTPHeader
	case strings.Contains(lowerName, "server"):
		return types.SourceHTTPHeader
	case strings.Contains(lowerName, "post") || strings.Contains(lowerName, "body"):
		return types.SourceHTTPPost
	case strings.Contains(lowerName, "query") || lowerName == "get":
		return types.SourceHTTPGet
	case strings.Contains(lowerName, "file"):
		return types.SourceHTTPBody
	default:
		return types.SourceUserInput
	}
}

// DetectFrameworks detects which PHP frameworks are being used
func (a *PHPAnalyzer) DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error) {
	var frameworks []string

	// Check imports for framework hints
	for _, imp := range symbolTable.Imports {
		path := strings.ToLower(imp.Path)

		if strings.Contains(path, "illuminate") || strings.Contains(path, "laravel") {
			if !contains(frameworks, "laravel") {
				frameworks = append(frameworks, "laravel")
			}
		}
		if strings.Contains(path, "symfony") {
			if !contains(frameworks, "symfony") {
				frameworks = append(frameworks, "symfony")
			}
		}
		if strings.Contains(path, "codeigniter") || strings.Contains(path, "ci_") {
			if !contains(frameworks, "codeigniter") {
				frameworks = append(frameworks, "codeigniter")
			}
		}
	}

	// Check class names
	for className := range symbolTable.Classes {
		lowerName := strings.ToLower(className)
		if lowerName == "mybb" {
			if !contains(frameworks, "mybb") {
				frameworks = append(frameworks, "mybb")
			}
		}
	}

	// Check for WordPress patterns
	sourceStr := string(source)
	if strings.Contains(sourceStr, "wp_") || strings.Contains(sourceStr, "WP_") ||
		strings.Contains(sourceStr, "WordPress") || strings.Contains(sourceStr, "get_option(") {
		if !contains(frameworks, "wordpress") {
			frameworks = append(frameworks, "wordpress")
		}
	}

	return frameworks, nil
}

// AnalyzeMethodBody analyzes a method body for data flow
func (a *PHPAnalyzer) AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*analyzer.MethodFlowAnalysis, error) {
	analysis := &analyzer.MethodFlowAnalysis{
		ParamsToReturn:     make([]int, 0),
		ParamsToProperties: make(map[int][]string),
		ParamsToCallArgs:   make(map[int][]*types.CallSite),
		TaintedVariables:   make(map[string]*types.TaintInfo),
		Assignments:        make([]*types.Assignment, 0),
		Calls:              make([]*types.CallSite, 0),
		Returns:            make([]analyzer.ReturnInfo, 0),
	}

	// This would need the actual AST of the method body
	// For now, we'll do text-based analysis
	if method.BodySource == "" {
		return analysis, nil
	}

	body := method.BodySource

	// Track which parameters flow where
	for i, param := range method.Parameters {
		paramName := "$" + param.Name

		// Check if param flows to return
		if strings.Contains(body, "return") && strings.Contains(body, paramName) {
			analysis.ParamsToReturn = append(analysis.ParamsToReturn, i)
		}

		// Check if param flows to $this->property
		thisAssignRegex := regexp.MustCompile(`\$this->(\w+)\s*=.*` + regexp.QuoteMeta(paramName))
		matches := thisAssignRegex.FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			if len(match) > 1 {
				analysis.ParamsToProperties[i] = append(analysis.ParamsToProperties[i], match[1])
				analysis.ModifiesProperties = true
			}
		}

		// Check if param flows to $this->property[$key] = ...
		thisArrayAssignRegex := regexp.MustCompile(`\$this->(\w+)\[.*\]\s*=.*` + regexp.QuoteMeta(paramName))
		arrayMatches := thisArrayAssignRegex.FindAllStringSubmatch(body, -1)
		for _, match := range arrayMatches {
			if len(match) > 1 {
				analysis.ParamsToProperties[i] = append(analysis.ParamsToProperties[i], match[1])
				analysis.ModifiesProperties = true
			}
		}
	}

	// Check if method returns input directly
	for sg := range a.superglobals {
		if strings.Contains(body, "return") && strings.Contains(body, sg) {
			analysis.ReturnsInput = true
			break
		}
	}

	return analysis, nil
}

// TraceExpression traces a specific expression back to its sources
func (a *PHPAnalyzer) TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error) {
	flowMap := &types.FlowMap{
		Target:   target,
		Sources:  make([]types.FlowNode, 0),
		Paths:    make([]types.FlowPath, 0),
		Carriers: make([]types.FlowNode, 0),
		AllNodes: make([]types.FlowNode, 0),
		AllEdges: make([]types.FlowEdge, 0),
		Usages:   make([]types.FlowNode, 0),
	}

	// Parse the target expression
	expr := target.Expression

	// Check for framework patterns first
	for _, pattern := range a.GetFrameworkPatterns() {
		if a.matchesFrameworkPattern(expr, pattern) {
			// This is a known framework carrier
			flowMap.CarrierChain = &types.CarrierChain{
				ClassName:        pattern.CarrierClass,
				PropertyName:     pattern.CarrierProperty,
				PopulationMethod: pattern.PopulatedBy,
				PopulationCalls:  pattern.PopulatedFrom,
				Framework:        pattern.Framework,
			}

			// Add sources based on framework knowledge
			for _, src := range pattern.PopulatedFrom {
				sourceType := types.SourceHTTPGet
				if strings.Contains(src, "POST") {
					sourceType = types.SourceHTTPPost
				} else if strings.Contains(src, "COOKIE") {
					sourceType = types.SourceHTTPCookie
				}

				// Extract key from expression if present
				key := a.extractKeyFromExpression(expr)

				sourceNode := types.FlowNode{
					ID:         fmt.Sprintf("source-%s-%s", src, key),
					Type:       types.NodeSource,
					Language:   "php",
					Name:       src,
					Snippet:    fmt.Sprintf("%s['%s']", src, key),
					SourceType: sourceType,
					SourceKey:  key,
				}
				flowMap.Sources = append(flowMap.Sources, sourceNode)
				flowMap.AllNodes = append(flowMap.AllNodes, sourceNode)
			}

			break
		}
	}

	// If no framework pattern matched, do generic tracing
	if len(flowMap.Sources) == 0 {
		// Check for direct superglobal access
		for sg, sourceType := range a.superglobals {
			if strings.Contains(expr, sg) {
				key := a.extractKeyFromExpression(expr)
				sourceNode := types.FlowNode{
					ID:         fmt.Sprintf("source-%s-%s", sg, key),
					Type:       types.NodeSource,
					Language:   "php",
					Name:       sg,
					Snippet:    expr,
					SourceType: sourceType,
					SourceKey:  key,
				}
				flowMap.Sources = append(flowMap.Sources, sourceNode)
				flowMap.AllNodes = append(flowMap.AllNodes, sourceNode)
			}
		}
	}

	return flowMap, nil
}

// matchesFrameworkPattern checks if an expression matches a framework pattern
func (a *PHPAnalyzer) matchesFrameworkPattern(expr string, pattern *types.FrameworkPattern) bool {
	// Check for class->property pattern
	if pattern.CarrierClass != "" && pattern.CarrierProperty != "" {
		// Match $varname->property or $varname->property[...]
		regex := regexp.MustCompile(`\$\w+->(` + regexp.QuoteMeta(pattern.CarrierProperty) + `)(\[|$)`)
		return regex.MatchString(expr)
	}

	// Check for method call pattern
	if pattern.MethodPattern != "" {
		regex := regexp.MustCompile(`->` + pattern.MethodPattern + `\(`)
		return regex.MatchString(expr)
	}

	return false
}

// extractKeyFromExpression extracts the array key from an expression like $mybb->input['thumbnail']
func (a *PHPAnalyzer) extractKeyFromExpression(expr string) string {
	// Match ['key'] or ["key"]
	regex := regexp.MustCompile(`\[['"](\w+)['"]\]`)
	matches := regex.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}

	// Match [$variable]
	regex2 := regexp.MustCompile(`\[(\$\w+)\]`)
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

// Register the PHP analyzer
func init() {
	analyzer.DefaultRegistry.Register(NewPHPAnalyzer())
}
