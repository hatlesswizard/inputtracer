// Package discovery - taint propagation engine
package discovery

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"

	"github.com/hatlesswizard/inputtracer/pkg/sources"
	phppatterns "github.com/hatlesswizard/inputtracer/pkg/sources/php"
)

// Regex cache for avoiding repeated compilation of the same patterns
var taintRegexCache sync.Map // pattern string -> *regexp.Regexp

// getOrCompileTaintRegex returns a cached compiled regex, compiling it if not already cached
func getOrCompileTaintRegex(pattern string) *regexp.Regexp {
	if cached, ok := taintRegexCache.Load(pattern); ok {
		return cached.(*regexp.Regexp)
	}
	compiled := regexp.MustCompile(pattern)
	taintRegexCache.Store(pattern, compiled)
	return compiled
}

// TaintSource represents the origin of tainted data
type TaintSource struct {
	Type      string   `json:"type"`       // "superglobal", "function_return", "property"
	Name      string   `json:"name"`       // "$_GET", "$_POST", etc.
	Keys      []string `json:"keys"`       // Keys accessed, or ["*"] for all keys
	FilePath  string   `json:"file_path"`
	Line      int      `json:"line"`
}

// TaintSink represents where tainted data ends up
type TaintSink struct {
	Type         string `json:"type"`          // "property", "method", "variable", "function", "array_element"
	ClassName    string `json:"class_name"`    // For properties/methods
	Name         string `json:"name"`          // Property name or method name
	IsPublic     bool   `json:"is_public"`     // Can be accessed externally
	AccessMethod string `json:"access_method"` // "array" for $obj->prop['key'], "method" for $obj->method('key'), "direct" for $obj->prop
}

// TaintFlow represents one propagation path from source to sink
type TaintFlow struct {
	Source     TaintSource `json:"source"`
	Sink       TaintSink   `json:"sink"`
	FilePath   string      `json:"file_path"`
	Line       int         `json:"line"`
	Confidence float64     `json:"confidence"`
	FlowType   string      `json:"flow_type"` // "direct_assignment", "foreach_population", "method_return", "parameter_pass"
}

// InputCarrier represents a discovered class property or method that carries user input
type InputCarrier struct {
	ClassName     string   `json:"class_name"`
	PropertyName  string   `json:"property_name,omitempty"`  // Empty if method-based
	MethodName    string   `json:"method_name,omitempty"`    // Empty if property-based
	SourceTypes   []string `json:"source_types"`             // ["$_GET", "$_POST"]
	AccessPattern string   `json:"access_pattern"`           // "array", "method", "direct"
	PopulatedIn   string   `json:"populated_in"`             // Constructor or method name that populates it
	FilePath      string   `json:"file_path"`
	Line          int      `json:"line"`
	Confidence    float64  `json:"confidence"`
}

// TaintPropagator traces data flow from superglobals to class properties/methods
type TaintPropagator struct {
	parser       *sitter.Parser
	flows        []TaintFlow
	carriers     []InputCarrier
	classInfo    map[string]*ClassInfo
	codebasePath string
}

// ClassInfo stores information about a class
type ClassInfo struct {
	Name           string
	FilePath       string
	Properties     map[string]*PropertyInfo
	Methods        map[string]*MethodInfo
	Constructor    *MethodInfo
	ParentClass    string
	Implements     []string
}

// PropertyInfo stores information about a class property
type PropertyInfo struct {
	Name         string
	IsPublic     bool
	IsProtected  bool
	IsPrivate    bool
	IsStatic     bool
	DefaultValue string
	Line         int
	TaintSources []string // Superglobals that flow into this property
}

// MethodInfo stores information about a class method
// Optimized to avoid storing full body source when not needed
type MethodInfo struct {
	Name           string
	IsPublic       bool
	IsProtected    bool
	IsPrivate      bool
	IsStatic       bool
	Line           int
	Parameters     []string
	ReturnsTainted bool
	TaintSources   []string // Superglobals that can flow through this method
	BodySource     string   // Only populated for methods that need deep analysis

	// Pre-analyzed patterns (computed once, stored as booleans)
	hasThisArrayAssign    bool // Has $this->prop[$key] = pattern
	hasDynamicPropAssign  bool // Has $this->$key = $val pattern
	hasReturnThisProp     bool // Has return $this->prop pattern
	analyzedPatterns      bool // Whether patterns have been analyzed
}

// NewTaintPropagator creates a new taint propagation engine
func NewTaintPropagator() *TaintPropagator {
	parser := sitter.NewParser()
	parser.SetLanguage(php.GetLanguage())

	return &TaintPropagator{
		parser:    parser,
		flows:     make([]TaintFlow, 0, 256),    // Pre-allocate reasonable capacity
		carriers:  make([]InputCarrier, 0, 64),  // Pre-allocate reasonable capacity
		classInfo: make(map[string]*ClassInfo),
	}
}

// analyzeMethodPatterns analyzes and caches method patterns
// This avoids repeated regex compilations
func (m *MethodInfo) analyzeMethodPatterns() {
	if m.analyzedPatterns || m.BodySource == "" {
		return
	}
	m.analyzedPatterns = true

	// Use centralized patterns from pkg/sources/php
	thisArrayPattern := phppatterns.TaintPatterns.ThisArrayPattern
	dynamicPropPattern := phppatterns.TaintPatterns.DynamicPropPattern
	returnThisPattern := phppatterns.TaintPatterns.ReturnThisPattern

	m.hasThisArrayAssign = thisArrayPattern.MatchString(m.BodySource)
	m.hasDynamicPropAssign = dynamicPropPattern.MatchString(m.BodySource)
	m.hasReturnThisProp = returnThisPattern.MatchString(m.BodySource)
}

// ReleaseBodySource releases the body source to free memory
// Call this after pattern analysis is complete
func (m *MethodInfo) ReleaseBodySource() {
	if m.analyzedPatterns {
		m.BodySource = ""
	}
}

// AnalyzeCodebase analyzes a codebase to discover input carriers
// Memory-optimized to release ASTs and method bodies after extraction
func (p *TaintPropagator) AnalyzeCodebase(codebasePath string) error {
	p.codebasePath = codebasePath

	// Phase 1: Parse all PHP files and extract class information
	if err := p.parseAllClasses(codebasePath); err != nil {
		return err
	}

	// Phase 2: Find all superglobal usages
	finder := NewSuperglobalFinder()
	usages, err := finder.FindAll(codebasePath)
	if err != nil {
		return err
	}

	// Phase 3: Trace flows from superglobals to class properties
	p.traceFlows(usages)

	// Phase 4: Discover carriers from flows
	p.discoverCarriers()

	// Phase 5: MEMORY OPTIMIZATION - Release method body sources
	// Patterns have been analyzed, body sources no longer needed
	p.releaseMethodBodies()

	return nil
}

// releaseMethodBodies releases all method body sources to free memory
func (p *TaintPropagator) releaseMethodBodies() {
	for _, classInfo := range p.classInfo {
		for _, method := range classInfo.Methods {
			method.analyzeMethodPatterns()
			method.ReleaseBodySource()
		}
		if classInfo.Constructor != nil {
			classInfo.Constructor.analyzeMethodPatterns()
			classInfo.Constructor.ReleaseBodySource()
		}
	}
}

// Clear releases all memory held by the propagator
func (p *TaintPropagator) Clear() {
	p.flows = nil
	p.carriers = nil
	p.classInfo = nil
}

// parseAllClasses parses all PHP files and extracts class information
func (p *TaintPropagator) parseAllClasses(codebasePath string) error {
	return filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".php") {
			return nil
		}

		// Skip vendor, cache directories - use centralized skip list
		if sources.ShouldSkipPHPPath(path) {
			return nil
		}

		p.parseClassesInFile(path)
		return nil
	})
}

// parseClassesInFile extracts class information from a single file
func (p *TaintPropagator) parseClassesInFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	tree, err := p.parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return err
	}
	defer tree.Close()

	p.extractClasses(tree.RootNode(), content, filePath)
	return nil
}

// extractClasses walks the AST to find class definitions
func (p *TaintPropagator) extractClasses(node *sitter.Node, content []byte, filePath string) {
	if node == nil {
		return
	}

	if node.Type() == "class_declaration" {
		classInfo := p.parseClassDeclaration(node, content, filePath)
		if classInfo != nil {
			p.classInfo[classInfo.Name] = classInfo
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		p.extractClasses(node.Child(i), content, filePath)
	}
}

// parseClassDeclaration parses a class declaration node
func (p *TaintPropagator) parseClassDeclaration(node *sitter.Node, content []byte, filePath string) *ClassInfo {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	info := &ClassInfo{
		Name:       nameNode.Content(content),
		FilePath:   filePath,
		Properties: make(map[string]*PropertyInfo),
		Methods:    make(map[string]*MethodInfo),
	}

	// Find parent class
	baseClause := node.ChildByFieldName("base_clause")
	if baseClause != nil {
		for i := 0; i < int(baseClause.ChildCount()); i++ {
			child := baseClause.Child(i)
			if child.Type() == "name" || child.Type() == "qualified_name" {
				info.ParentClass = child.Content(content)
				break
			}
		}
	}

	// Parse class body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		for i := 0; i < int(bodyNode.ChildCount()); i++ {
			child := bodyNode.Child(i)
			switch child.Type() {
			case "property_declaration":
				prop := p.parsePropertyDeclaration(child, content)
				if prop != nil {
					info.Properties[prop.Name] = prop
				}
			case "method_declaration":
				method := p.parseMethodDeclaration(child, content)
				if method != nil {
					info.Methods[method.Name] = method
					if method.Name == "__construct" {
						info.Constructor = method
					}
				}
			}
		}
	}

	return info
}

// parsePropertyDeclaration parses a property declaration
func (p *TaintPropagator) parsePropertyDeclaration(node *sitter.Node, content []byte) *PropertyInfo {
	prop := &PropertyInfo{
		Line: int(node.StartPoint().Row) + 1,
	}

	// Check visibility modifiers
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "visibility_modifier":
			modifier := child.Content(content)
			switch modifier {
			case "public":
				prop.IsPublic = true
			case "protected":
				prop.IsProtected = true
			case "private":
				prop.IsPrivate = true
			}
		case "static_modifier":
			prop.IsStatic = true
		case "property_element":
			// Get property name
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				prop.Name = strings.TrimPrefix(nameNode.Content(content), "$")
			}
			// Get default value
			valueNode := child.ChildByFieldName("value")
			if valueNode != nil {
				prop.DefaultValue = valueNode.Content(content)
			}
		}
	}

	if prop.Name == "" {
		return nil
	}

	return prop
}

// parseMethodDeclaration parses a method declaration
func (p *TaintPropagator) parseMethodDeclaration(node *sitter.Node, content []byte) *MethodInfo {
	method := &MethodInfo{
		Line:       int(node.StartPoint().Row) + 1,
		Parameters: make([]string, 0),
	}

	// Check visibility modifiers
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "visibility_modifier":
			modifier := child.Content(content)
			switch modifier {
			case "public":
				method.IsPublic = true
			case "protected":
				method.IsProtected = true
			case "private":
				method.IsPrivate = true
			}
		case "static_modifier":
			method.IsStatic = true
		}
	}

	// Get method name
	nameNode := node.ChildByFieldName("name")
	if nameNode != nil {
		method.Name = nameNode.Content(content)
	}

	// Get parameters
	paramsNode := node.ChildByFieldName("parameters")
	if paramsNode != nil {
		for i := 0; i < int(paramsNode.ChildCount()); i++ {
			child := paramsNode.Child(i)
			if child.Type() == "simple_parameter" || child.Type() == "variadic_parameter" {
				paramName := child.ChildByFieldName("name")
				if paramName != nil {
					method.Parameters = append(method.Parameters, paramName.Content(content))
				}
			}
		}
	}

	// Get method body source
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		method.BodySource = bodyNode.Content(content)
	}

	if method.Name == "" {
		return nil
	}

	return method
}

// traceFlows traces data flow from superglobal usages to class properties
func (p *TaintPropagator) traceFlows(usages []SuperglobalUsage) {
	for _, usage := range usages {
		p.traceUsage(usage)
	}
}

// traceUsage traces a single superglobal usage
func (p *TaintPropagator) traceUsage(usage SuperglobalUsage) {
	source := TaintSource{
		Type:     "superglobal",
		Name:     usage.Superglobal,
		Keys:     []string{usage.Key},
		FilePath: usage.FilePath,
		Line:     usage.Line,
	}

	switch usage.Context {
	case "assignment":
		p.traceAssignment(usage, source)
	case "foreach":
		p.traceForeach(usage, source)
	case "return":
		p.traceReturn(usage, source)
	case "function_arg":
		p.traceMethodCall(usage, source)
	}
}

// traceAssignment traces an assignment of a superglobal
func (p *TaintPropagator) traceAssignment(usage SuperglobalUsage, source TaintSource) {
	// Check if assigned to $this->property - use centralized pattern
	thisPattern := phppatterns.TaintPatterns.ThisPropertyOptionalArrayPattern
	if matches := thisPattern.FindStringSubmatch(usage.AssignedTo); len(matches) >= 2 {
		propName := matches[1]

		// Find the class this assignment is in
		if classInfo, ok := p.classInfo[usage.ClassName]; ok {
			sink := TaintSink{
				Type:         "property",
				ClassName:    usage.ClassName,
				Name:         propName,
				AccessMethod: "array",
			}

			// Check if property exists and is public
			if prop, ok := classInfo.Properties[propName]; ok {
				sink.IsPublic = prop.IsPublic
			} else {
				sink.IsPublic = true // Dynamic property, assume public
			}

			flow := TaintFlow{
				Source:     source,
				Sink:       sink,
				FilePath:   usage.FilePath,
				Line:       usage.Line,
				Confidence: 1.0,
				FlowType:   "direct_assignment",
			}
			p.flows = append(p.flows, flow)
		}
	}
}

// traceForeach traces a foreach loop over a superglobal
func (p *TaintPropagator) traceForeach(usage SuperglobalUsage, source TaintSource) {
	if usage.ClassName == "" {
		return
	}

	classInfo, ok := p.classInfo[usage.ClassName]
	if !ok {
		return
	}

	// Check the method body for patterns like:
	// $this->property[$key] = $value
	// $this->$key = $value
	methodInfo, ok := classInfo.Methods[usage.MethodName]
	if !ok {
		return
	}

	// Look for $this->property[$key] = $value pattern
	propArrayPattern := getOrCompileTaintRegex(`\$this->(\w+)\[\$` + regexp.QuoteMeta(strings.TrimPrefix(usage.LoopKeyVar, "$")) + `\]\s*=`)
	if matches := propArrayPattern.FindStringSubmatch(methodInfo.BodySource); len(matches) >= 2 {
		propName := matches[1]

		sink := TaintSink{
			Type:         "property",
			ClassName:    usage.ClassName,
			Name:         propName,
			IsPublic:     true,
			AccessMethod: "array",
		}

		if prop, ok := classInfo.Properties[propName]; ok {
			sink.IsPublic = prop.IsPublic
		}

		flow := TaintFlow{
			Source:     source,
			Sink:       sink,
			FilePath:   usage.FilePath,
			Line:       usage.Line,
			Confidence: 1.0,
			FlowType:   "foreach_population",
		}
		p.flows = append(p.flows, flow)
	}

	// Look for $this->$key = $value pattern (dynamic property)
	dynamicPropPattern := getOrCompileTaintRegex(`\$this->\$` + regexp.QuoteMeta(strings.TrimPrefix(usage.LoopKeyVar, "$")) + `\s*=`)
	if dynamicPropPattern.MatchString(methodInfo.BodySource) {
		// This means ALL superglobal keys become properties
		sink := TaintSink{
			Type:         "property",
			ClassName:    usage.ClassName,
			Name:         "*", // All keys become properties
			IsPublic:     true,
			AccessMethod: "direct",
		}

		flow := TaintFlow{
			Source:     source,
			Sink:       sink,
			FilePath:   usage.FilePath,
			Line:       usage.Line,
			Confidence: 1.0,
			FlowType:   "foreach_population",
		}
		p.flows = append(p.flows, flow)
	}
}

// traceReturn traces a return statement with a superglobal
func (p *TaintPropagator) traceReturn(usage SuperglobalUsage, source TaintSource) {
	if usage.ClassName == "" || usage.MethodName == "" {
		return
	}

	sink := TaintSink{
		Type:         "method",
		ClassName:    usage.ClassName,
		Name:         usage.MethodName,
		AccessMethod: "method",
	}

	// Check if method is public
	if classInfo, ok := p.classInfo[usage.ClassName]; ok {
		if methodInfo, ok := classInfo.Methods[usage.MethodName]; ok {
			sink.IsPublic = methodInfo.IsPublic
		}
	}

	flow := TaintFlow{
		Source:     source,
		Sink:       sink,
		FilePath:   usage.FilePath,
		Line:       usage.Line,
		Confidence: 1.0,
		FlowType:   "method_return",
	}
	p.flows = append(p.flows, flow)
}

// traceMethodCall traces a superglobal passed as an argument to a method
// e.g., $this->parse_incoming($_GET) where parse_incoming assigns to $this->input
func (p *TaintPropagator) traceMethodCall(usage SuperglobalUsage, source TaintSource) {
	if usage.ClassName == "" || usage.CalledMethod == "" {
		return
	}

	classInfo, ok := p.classInfo[usage.ClassName]
	if !ok {
		return
	}

	// Find the method that's being called
	methodInfo, ok := classInfo.Methods[usage.CalledMethod]
	if !ok {
		return
	}

	// Analyze the method body to see what property the parameter flows into
	// Pattern 1: $this->property[$key] = $val (foreach population)
	propArrayPattern := getOrCompileTaintRegex(`\$this->(\w+)\[\$\w+\]\s*=\s*\$`)
	if matches := propArrayPattern.FindStringSubmatch(methodInfo.BodySource); len(matches) >= 2 {
		propName := matches[1]

		sink := TaintSink{
			Type:         "property",
			ClassName:    usage.ClassName,
			Name:         propName,
			IsPublic:     true,
			AccessMethod: "array",
		}

		if prop, ok := classInfo.Properties[propName]; ok {
			sink.IsPublic = prop.IsPublic
		}

		flow := TaintFlow{
			Source:     source,
			Sink:       sink,
			FilePath:   usage.FilePath,
			Line:       usage.Line,
			Confidence: 1.0,
			FlowType:   "method_call_propagation",
		}
		p.flows = append(p.flows, flow)
		return
	}

	// Pattern 2: Direct assignment in method: $this->property = $param
	// Look for the first parameter name
	paramName := ""
	if len(methodInfo.Parameters) > 0 {
		paramName = strings.TrimPrefix(methodInfo.Parameters[0], "$")
	}
	if paramName != "" {
		directAssignPattern := getOrCompileTaintRegex(`\$this->(\w+)\s*=\s*\$` + regexp.QuoteMeta(paramName))
		if matches := directAssignPattern.FindStringSubmatch(methodInfo.BodySource); len(matches) >= 2 {
			propName := matches[1]

			sink := TaintSink{
				Type:         "property",
				ClassName:    usage.ClassName,
				Name:         propName,
				IsPublic:     true,
				AccessMethod: "direct",
			}

			if prop, ok := classInfo.Properties[propName]; ok {
				sink.IsPublic = prop.IsPublic
			}

			flow := TaintFlow{
				Source:     source,
				Sink:       sink,
				FilePath:   usage.FilePath,
				Line:       usage.Line,
				Confidence: 1.0,
				FlowType:   "method_call_propagation",
			}
			p.flows = append(p.flows, flow)
		}
	}
}

// discoverCarriers converts flows into input carriers
func (p *TaintPropagator) discoverCarriers() {
	// Group flows by class+property/method
	carrierMap := make(map[string]*InputCarrier)

	for _, flow := range p.flows {
		if flow.Sink.ClassName == "" {
			continue
		}

		key := flow.Sink.ClassName + "." + flow.Sink.Name
		if flow.Sink.Type == "method" {
			key = flow.Sink.ClassName + ":" + flow.Sink.Name + "()"
		}

		if carrier, ok := carrierMap[key]; ok {
			// Add source type if not already present
			if !contains(carrier.SourceTypes, flow.Source.Name) {
				carrier.SourceTypes = append(carrier.SourceTypes, flow.Source.Name)
			}
		} else {
			carrier := &InputCarrier{
				ClassName:     flow.Sink.ClassName,
				SourceTypes:   []string{flow.Source.Name},
				AccessPattern: flow.Sink.AccessMethod,
				FilePath:      flow.FilePath,
				Line:          flow.Line,
				Confidence:    flow.Confidence,
			}

			if flow.Sink.Type == "property" {
				carrier.PropertyName = flow.Sink.Name
			} else if flow.Sink.Type == "method" {
				carrier.MethodName = flow.Sink.Name
			}

			// Find where it's populated
			carrier.PopulatedIn = p.findPopulationMethod(flow.Sink.ClassName, flow.Sink.Name)

			carrierMap[key] = carrier
		}
	}

	// Convert map to slice
	for _, carrier := range carrierMap {
		p.carriers = append(p.carriers, *carrier)
	}

	// Also check for methods that return from tainted properties
	p.discoverMethodCarriersFromProperties()
}

// findPopulationMethod finds the method that populates a property
func (p *TaintPropagator) findPopulationMethod(className, propName string) string {
	for _, flow := range p.flows {
		if flow.Sink.ClassName == className && flow.Sink.Name == propName {
			// Find the method from the file path and line
			if classInfo, ok := p.classInfo[className]; ok {
				for methodName, methodInfo := range classInfo.Methods {
					if flow.Line >= methodInfo.Line {
						return methodName
					}
				}
			}
		}
	}
	return "__construct" // Default assumption
}

// discoverMethodCarriersFromProperties discovers methods that return tainted properties
func (p *TaintPropagator) discoverMethodCarriersFromProperties() {
	// For each class with tainted properties, check if any methods return them
	taintedProps := make(map[string][]string) // className -> []propertyName

	for _, carrier := range p.carriers {
		if carrier.PropertyName != "" {
			taintedProps[carrier.ClassName] = append(taintedProps[carrier.ClassName], carrier.PropertyName)
		}
	}

	for className, props := range taintedProps {
		classInfo, ok := p.classInfo[className]
		if !ok {
			continue
		}

		for methodName, methodInfo := range classInfo.Methods {
			if !methodInfo.IsPublic {
				continue
			}

			// Check if method returns a tainted property
			for _, propName := range props {
				// Build pattern for return $this->propName
				returnPattern := phppatterns.BuildReturnPropertyPattern(propName)
				if returnPattern.MatchString(methodInfo.BodySource) {
					// Find source types for this property
					var sourceTypes []string
					for _, carrier := range p.carriers {
						if carrier.ClassName == className && carrier.PropertyName == propName {
							sourceTypes = carrier.SourceTypes
							break
						}
					}

					// Check if this method carrier already exists
					exists := false
					for _, c := range p.carriers {
						if c.ClassName == className && c.MethodName == methodName {
							exists = true
							break
						}
					}

					if !exists && len(sourceTypes) > 0 {
						methodCarrier := InputCarrier{
							ClassName:     className,
							MethodName:    methodName,
							SourceTypes:   sourceTypes,
							AccessPattern: "method",
							PopulatedIn:   methodName,
							FilePath:      classInfo.FilePath,
							Line:          methodInfo.Line,
							Confidence:    0.9, // Slightly lower confidence for indirect
						}
						p.carriers = append(p.carriers, methodCarrier)
					}
				}

				// Build pattern for return $this->propName[
				paramReturnPattern := phppatterns.BuildReturnPropertyArrayPattern(propName)
				if paramReturnPattern.MatchString(methodInfo.BodySource) {
					var sourceTypes []string
					for _, carrier := range p.carriers {
						if carrier.ClassName == className && carrier.PropertyName == propName {
							sourceTypes = carrier.SourceTypes
							break
						}
					}

					exists := false
					for _, c := range p.carriers {
						if c.ClassName == className && c.MethodName == methodName {
							exists = true
							break
						}
					}

					if !exists && len(sourceTypes) > 0 {
						methodCarrier := InputCarrier{
							ClassName:     className,
							MethodName:    methodName,
							SourceTypes:   sourceTypes,
							AccessPattern: "method",
							PopulatedIn:   methodName,
							FilePath:      classInfo.FilePath,
							Line:          methodInfo.Line,
							Confidence:    0.9,
						}
						p.carriers = append(p.carriers, methodCarrier)
					}
				}
			}
		}
	}
}

// GetFlows returns all discovered taint flows
func (p *TaintPropagator) GetFlows() []TaintFlow {
	return p.flows
}

// GetCarriers returns all discovered input carriers
func (p *TaintPropagator) GetCarriers() []InputCarrier {
	return p.carriers
}

// GetClassInfo returns class information
func (p *TaintPropagator) GetClassInfo() map[string]*ClassInfo {
	return p.classInfo
}

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
