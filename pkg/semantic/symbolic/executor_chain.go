package symbolic

import (
	"fmt"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	pkgSources "github.com/hatlesswizard/inputtracer/pkg/sources"
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
	"github.com/hatlesswizard/inputtracer/pkg/sources/patterns"
	phpPatterns "github.com/hatlesswizard/inputtracer/pkg/sources/php"
	sitter "github.com/smacker/go-tree-sitter"
)

// checkMagicPropertyPattern checks if a class uses magic __get or dynamic property assignment
func (e *ExecutionEngine) checkMagicPropertyPattern(classDef *types.ClassDef, classFile string) *MagicPropertyInfo {
	info := &MagicPropertyInfo{}

	// Check for __get magic method
	if method, ok := classDef.Methods["__get"]; ok {
		info.HasMagicGet = true
		// Look for return $this->property[$name] pattern
		if matches := patterns.BackingPropertyPattern.FindStringSubmatch(method.BodySource); len(matches) >= 2 {
			info.BackingProperty = matches[1]
		}
		return info
	}

	// Check for dynamic property assignment pattern: $this->$key = $val
	// This is used in classes like MyLanguage that load properties dynamically
	for methodName, method := range classDef.Methods {
		if method.BodySource != "" {
			if patterns.DynamicPropertyAssignPattern.MatchString(method.BodySource) {
				info.HasDynamicAssign = true
				info.AssignMethodName = methodName

				// Check if values come from require/include (file_include)
				if strings.Contains(method.BodySource, "require") || strings.Contains(method.BodySource, "include") {
					info.SourceType = "file_include"
				}

				// Check for foreach pattern: foreach($array as $key => $val) { $this->$key = $val; }
				if matches := patterns.ForeachWithKVPattern.FindStringSubmatch(method.BodySource); len(matches) >= 2 {
					info.BackingProperty = matches[1]
				}

				return info
			}
		}
	}

	return nil
}

// traceMagicProperty traces a property accessed via magic method or dynamic assignment
func (e *ExecutionEngine) traceMagicProperty(parsed *ParsedExpression, classDef *types.ClassDef, classFile string, magicInfo *MagicPropertyInfo, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.PropertyName = parsed.PropertyName
	flow.AccessKey = parsed.AccessKey

	if magicInfo.HasMagicGet {
		// Magic __get method
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  1,
			Description: fmt.Sprintf("Dynamic property $%s accessed via __get magic method", parsed.PropertyName),
			Code:        fmt.Sprintf("public function __get($name) { return $this->%s[$name]; }", magicInfo.BackingProperty),
			FilePath:    classFile,
			Line:        0,
			Type:        "magic_get",
		})

		if magicInfo.BackingProperty != "" {
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  2,
				Description: fmt.Sprintf("Returns $this->%s['%s']", magicInfo.BackingProperty, parsed.PropertyName),
				Code:        fmt.Sprintf("return $this->%s['%s'];", magicInfo.BackingProperty, parsed.PropertyName),
				FilePath:    classFile,
				Line:        0,
				Type:        "return",
			})
		}
	} else if magicInfo.HasDynamicAssign {
		// Dynamic property assignment pattern
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  1,
			Description: fmt.Sprintf("Dynamic property $%s assigned via %s() method", parsed.PropertyName, magicInfo.AssignMethodName),
			Code:        fmt.Sprintf("// In %s(): $this->$key = $val;", magicInfo.AssignMethodName),
			FilePath:    classFile,
			Line:        0,
			Type:        "dynamic_assignment",
		})

		if magicInfo.SourceType == "file_include" {
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  2,
				Description: "Property values loaded from included PHP files",
				Code:        fmt.Sprintf("// require/include loads data that populates $%s", parsed.PropertyName),
				FilePath:    classFile,
				Line:        0,
				Type:        "file_include",
			})

			// For language files, the source is file data not user input
			flow.Sources = append(flow.Sources, UltimateSource{
				Type:       "file_data",
				Expression: fmt.Sprintf("$lang->%s (loaded from language file)", parsed.PropertyName),
				FilePath:   classFile,
				Line:       0,
			})
		}

		if magicInfo.BackingProperty != "" {
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  3,
				Description: fmt.Sprintf("Data iterated from $%s array", magicInfo.BackingProperty),
				Code:        fmt.Sprintf("foreach($%s as $key => $val) { $this->$key = $val; }", magicInfo.BackingProperty),
				FilePath:    classFile,
				Line:        0,
				Type:        "loop",
			})
		}
	}

	return flow, nil
}

// traceStaticCall handles static method calls like Class::method()
func (e *ExecutionEngine) traceStaticCall(parsed *ParsedExpression, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.ClassName = parsed.ClassName
	flow.MethodName = parsed.MethodName

	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Static method call: %s::%s()", parsed.ClassName, parsed.MethodName),
		Code:        fmt.Sprintf("%s::%s(%s)", parsed.ClassName, parsed.MethodName, strings.Join(parsed.Arguments, ", ")),
		FilePath:    "",
		Line:        0,
		Type:        "static_call",
	})

	// Find the class definition
	classDef, classFile := e.findClassDefinition(parsed.ClassName)
	if classDef == nil {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Class %s not found", parsed.ClassName),
			Code:        "",
			FilePath:    "",
			Line:        0,
			Type:        "not_found",
		})
		return flow, nil
	}

	// Find the method
	methodDef, ok := classDef.Methods[parsed.MethodName]
	if !ok {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Static method %s not found in class %s", parsed.MethodName, parsed.ClassName),
			Code:        "",
			FilePath:    "",
			Line:        0,
			Type:        "not_found",
		})
		return flow, nil
	}

	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  2,
		Description: fmt.Sprintf("Method %s::%s() defined", parsed.ClassName, parsed.MethodName),
		Code:        fmt.Sprintf("public static function %s() { ... }", parsed.MethodName),
		FilePath:    classFile,
		Line:        methodDef.Line,
		Type:        "method_def",
	})

	// Analyze what the method returns
	returnInfo := e.analyzeMethodReturns(classDef, methodDef, classFile)
	for _, retStmt := range returnInfo.ReturnStatements {
		for sg, sgType := range pkgSources.SuperglobalToSourceType {
			if strings.Contains(retStmt, sg) {
				flow.Sources = append(flow.Sources, UltimateSource{
					Type:       string(sgType),
					Expression: sg,
					FilePath:   classFile,
					Line:       methodDef.Line,
				})
			}
		}
	}

	return flow, nil
}

// traceStaticProperty handles static property access like Class::$property or Class::CONSTANT
func (e *ExecutionEngine) traceStaticProperty(parsed *ParsedExpression, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.ClassName = parsed.ClassName
	flow.PropertyName = parsed.PropertyName

	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Static property/constant access: %s::%s", parsed.ClassName, parsed.PropertyName),
		Code:        fmt.Sprintf("%s::%s", parsed.ClassName, parsed.PropertyName),
		FilePath:    "",
		Line:        0,
		Type:        "static_property",
	})

	// Find the class definition
	classDef, classFile := e.findClassDefinition(parsed.ClassName)
	if classDef == nil {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Class %s not found", parsed.ClassName),
			Code:        "",
			FilePath:    "",
			Line:        0,
			Type:        "not_found",
		})
		return flow, nil
	}

	// Check for property
	if propDef, ok := classDef.Properties[parsed.PropertyName]; ok {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Static property %s::%s = %s", parsed.ClassName, parsed.PropertyName, propDef.InitialValue),
			Code:        fmt.Sprintf("public static $%s = %s;", parsed.PropertyName, propDef.InitialValue),
			FilePath:    classFile,
			Line:        propDef.Line,
			Type:        "property_def",
		})
	} else {
		// It might be a constant - check for constants in the class body
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Constant %s::%s (value lookup not implemented)", parsed.ClassName, parsed.PropertyName),
			Code:        fmt.Sprintf("const %s = ...;", parsed.PropertyName),
			FilePath:    classFile,
			Line:        0,
			Type:        "constant",
		})
	}

	return flow, nil
}

// traceChainedExpression traces expressions like $obj->method()->property
// GAP #4 FIX: Support chained method calls
func (e *ExecutionEngine) traceChainedExpression(parsed *ParsedExpression, classDef *types.ClassDef, classFile string, instFile string, instLine int, flow *PropertyFlow) (*PropertyFlow, error) {
	stepNum := 1

	// Step 1: Show instantiation
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  stepNum,
		Description: fmt.Sprintf("Variable %s instantiated as new %s()", parsed.VarName, classDef.Name),
		Code:        fmt.Sprintf("%s = new %s();", parsed.VarName, classDef.Name),
		FilePath:    instFile,
		Line:        instLine,
		Type:        "instantiation",
	})
	stepNum++

	// Track the current class as we traverse the chain
	currentClass := classDef
	currentClassFile := classFile

	// Process each step in the chain
	for i, step := range parsed.ChainSteps {
		isLastStep := i == len(parsed.ChainSteps)-1

		if step.Type == ExprTypeMethodCall {
			// Find the method
			methodDef, ok := currentClass.Methods[step.Name]
			if !ok {
				flow.Steps = append(flow.Steps, FlowStep{
					StepNumber:  stepNum,
					Description: fmt.Sprintf("Method %s() not found in class %s", step.Name, currentClass.Name),
					Code:        "",
					FilePath:    currentClassFile,
					Line:        0,
					Type:        "method_not_found",
				})
				break
			}

			// Show method call
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  stepNum,
				Description: fmt.Sprintf("Method call: ->%s(%s)", step.Name, strings.Join(step.Arguments, ", ")),
				Code:        fmt.Sprintf("function %s(%s) { ... }", step.Name, e.formatParams(methodDef.Parameters)),
				FilePath:    currentClassFile,
				Line:        methodDef.Line,
				Type:        "method_call",
			})
			stepNum++

			// Analyze method return type for next step
			returnInfo := e.analyzeMethodReturns(currentClass, methodDef, currentClassFile)

			// If method returns a property, and this is the last step, trace the property
			if isLastStep {
				if returnInfo.ReturnsProperty && returnInfo.PropertyName != "" {
					flow.PropertyName = returnInfo.PropertyName
					flow.MethodName = step.Name

					// Trace sources from that property
					propSteps := e.traceConstructor(currentClass, currentClassFile, returnInfo.PropertyName, step.AccessKey)
					for _, ps := range propSteps {
						ps.StepNumber = stepNum
						flow.Steps = append(flow.Steps, ps)
						stepNum++
					}
					// Extract sources using the existing extractSources function
					flow.Sources = append(flow.Sources, e.extractSources(propSteps)...)
				}
				break
			}

			// Check if method returns $this (fluent interface)
			if returnInfo.ReturnsSelf {
				// Continue with same class
				continue
			}

			// Try to infer return type from method
			returnType := e.inferMethodReturnType(currentClass, methodDef, currentClassFile)
			if returnType != "" {
				newClass, newClassFile := e.findClassDefinition(returnType)
				if newClass != nil {
					currentClass = newClass
					currentClassFile = newClassFile
					continue
				}
			}

			// Can't determine return type - provide partial trace
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  stepNum,
				Description: fmt.Sprintf("Cannot determine return type of %s() - chain tracing stopped", step.Name),
				Code:        "",
				FilePath:    "",
				Line:        0,
				Type:        "return_type_unknown",
			})
			break

		} else if step.Type == ExprTypePropertyAccess {
			// Property access
			propDef, ok := currentClass.Properties[step.Name]
			if !ok {
				// Check for magic methods or external assignments
				flow.Steps = append(flow.Steps, FlowStep{
					StepNumber:  stepNum,
					Description: fmt.Sprintf("Property %s not found in class %s", step.Name, currentClass.Name),
					Code:        "",
					FilePath:    currentClassFile,
					Line:        0,
					Type:        "property_not_found",
				})
				break
			}

			flow.PropertyName = step.Name
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  stepNum,
				Description: fmt.Sprintf("Property access: ->%s", step.Name),
				Code:        fmt.Sprintf("$this->%s = %s;", step.Name, propDef.InitialValue),
				FilePath:    currentClassFile,
				Line:        propDef.Line,
				Type:        "property_access",
			})
			stepNum++

			if isLastStep {
				// Trace the property sources
				propSteps := e.traceConstructor(currentClass, currentClassFile, step.Name, step.AccessKey)
				for _, ps := range propSteps {
					ps.StepNumber = stepNum
					flow.Steps = append(flow.Steps, ps)
					stepNum++
				}
				// Extract sources using the existing extractSources function
				flow.Sources = append(flow.Sources, e.extractSources(propSteps)...)
			}
		}
	}

	return flow, nil
}

// inferMethodReturnType tries to determine what class a method returns
func (e *ExecutionEngine) inferMethodReturnType(classDef *types.ClassDef, methodDef *types.MethodDef, classFile string) string {
	// Check for explicit return type annotation first
	// Pattern: function name(): ReturnType
	if methodDef.ReturnType != "" && methodDef.ReturnType != "void" && methodDef.ReturnType != "self" && methodDef.ReturnType != "static" {
		return methodDef.ReturnType
	}

	// Use the method body source
	body := methodDef.BodySource
	if body == "" {
		return ""
	}

	// Check for return new ClassName()
	if matches := patterns.ReturnNewPattern.FindStringSubmatch(body); len(matches) >= 2 {
		return matches[1]
	}

	// Check for @return PHPDoc annotation
	if matches := patterns.PHPDocReturnPattern.FindStringSubmatch(body); len(matches) >= 2 {
		returnType := matches[1]
		if returnType != "void" && returnType != "self" && returnType != "static" && returnType != "mixed" {
			return returnType
		}
	}

	return ""
}

// parseExpression parses any expression and determines its type

func (e *ExecutionEngine) traceMethodCall(parsed *ParsedExpression, classDef *types.ClassDef, classFile string, instFile string, instLine int, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.MethodName = parsed.MethodName
	flow.AccessKey = parsed.AccessKey

	// Find the method definition
	methodDef, ok := classDef.Methods[parsed.MethodName]
	if !ok {
		return nil, fmt.Errorf("method %s not found in class %s", parsed.MethodName, parsed.ClassName)
	}

	// Step 1: Show instantiation
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Variable %s instantiated as new %s()", parsed.VarName, parsed.ClassName),
		Code:        fmt.Sprintf("%s = new %s();", parsed.VarName, parsed.ClassName),
		FilePath:    instFile,
		Line:        instLine,
		Type:        "instantiation",
	})

	// Step 2: Show method call
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  2,
		Description: fmt.Sprintf("Method call: %s->%s('%s')", parsed.VarName, parsed.MethodName, parsed.AccessKey),
		Code:        fmt.Sprintf("%s->%s('%s')", parsed.VarName, parsed.MethodName, parsed.AccessKey),
		FilePath:    "",
		Line:        0,
		Type:        "method_call",
	})

	// Step 3: Analyze the method body to find what it returns
	returnInfo := e.analyzeMethodReturns(classDef, methodDef, classFile)

	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  3,
		Description: fmt.Sprintf("Method %s() defined", parsed.MethodName),
		Code:        fmt.Sprintf("function %s(%s) { ... }", parsed.MethodName, e.formatParams(methodDef.Parameters)),
		FilePath:    classFile,
		Line:        methodDef.Line,
		Type:        "method_def",
	})

	// Step 4: Show return analysis
	if returnInfo.ReturnsProperty {
		propName := returnInfo.PropertyName

		if returnInfo.UsesParamAsKey {
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber: 4,
				Description: fmt.Sprintf("Returns $this->%s[$%s] where $%s = '%s'",
					propName, methodDef.Parameters[returnInfo.ParamIndex].Name,
					methodDef.Parameters[returnInfo.ParamIndex].Name, parsed.AccessKey),
				Code:     fmt.Sprintf("return $this->%s[$%s];", propName, methodDef.Parameters[returnInfo.ParamIndex].Name),
				FilePath: classFile,
				Line:     methodDef.Line,
				Type:     "return",
			})

			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  5,
				Description: fmt.Sprintf("Resolves to: $this->%s['%s']", propName, parsed.AccessKey),
				Code:        fmt.Sprintf("// %s->%s('%s') == $this->%s['%s']", parsed.VarName, parsed.MethodName, parsed.AccessKey, propName, parsed.AccessKey),
				FilePath:    "",
				Line:        0,
				Type:        "resolution",
			})

			// Now trace the property that the method returns
			flow.PropertyName = propName

			// Find property definition
			if propDef, ok := classDef.Properties[propName]; ok {
				flow.Steps = append(flow.Steps, FlowStep{
					StepNumber:  6,
					Description: fmt.Sprintf("Property $%s starts as %s", propName, propDef.InitialValue),
					Code:        fmt.Sprintf("public $%s = %s;", propName, propDef.InitialValue),
					FilePath:    classFile,
					Line:        propDef.Line,
					Type:        "property_init",
				})
			}

			// Trace constructor to see how property is populated
			if classDef.Constructor != nil {
				e.currentDepth = 0
				constructorFlows := e.traceConstructor(classDef, classFile, propName, parsed.AccessKey)

				// Renumber steps
				for i := range constructorFlows {
					constructorFlows[i].StepNumber = len(flow.Steps) + i + 1
				}
				flow.Steps = append(flow.Steps, constructorFlows...)
			}

		} else {
			// Returns property directly without key
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  4,
				Description: fmt.Sprintf("Returns $this->%s", propName),
				Code:        fmt.Sprintf("return $this->%s;", propName),
				FilePath:    classFile,
				Line:        methodDef.Line,
				Type:        "return",
			})
		}
	}

	// Add return statements analysis
	for _, retStmt := range returnInfo.ReturnStatements {
		// Check for superglobals in return statements
		for sg, sgType := range pkgSources.SuperglobalToSourceType {
			if strings.Contains(retStmt, sg) {
				flow.Sources = append(flow.Sources, UltimateSource{
					Type:       string(sgType),
					Expression: sg,
					FilePath:   classFile,
					Line:       methodDef.Line,
				})
			}
		}
	}

	// Extract sources from steps
	flow.Sources = append(flow.Sources, e.extractSources(flow.Steps)...)

	return flow, nil
}

// tracePropertyAccessExpr traces a property access expression
func (e *ExecutionEngine) tracePropertyAccessExpr(parsed *ParsedExpression, classDef *types.ClassDef, classFile string, instFile string, instLine int, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.PropertyName = parsed.PropertyName
	flow.AccessKey = parsed.AccessKey

	// Find the property definition
	propDef, found := classDef.Properties[parsed.PropertyName]
	if !found {
		// GAP #6 FIX: Property not found in class definition
		// Check for external property assignments (dynamic properties)
		externalAssignments := e.findExternalPropertyAssignments(parsed.VarName, parsed.PropertyName)
		if len(externalAssignments) > 0 {
			return e.traceExternalPropertyAssignment(parsed, externalAssignments, flow)
		}
		// GAP #6 FIX Phase 2: Check for magic __get method or dynamic property assignment
		magicInfo := e.checkMagicPropertyPattern(classDef, classFile)
		if magicInfo != nil {
			return e.traceMagicProperty(parsed, classDef, classFile, magicInfo, flow)
		}
		return nil, fmt.Errorf("property %s not found in class %s", parsed.PropertyName, parsed.ClassName)
	}

	// Add step for property initialization
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Property $%s starts as %s", parsed.PropertyName, propDef.InitialValue),
		Code:        fmt.Sprintf("public $%s = %s;", parsed.PropertyName, propDef.InitialValue),
		FilePath:    classFile,
		Line:        propDef.Line,
		Type:        "property_init",
	})

	// If variable is instantiated, show that
	if instFile != "" {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Variable %s instantiated as new %s()", parsed.VarName, parsed.ClassName),
			Code:        fmt.Sprintf("%s = new %s();", parsed.VarName, parsed.ClassName),
			FilePath:    instFile,
			Line:        instLine,
			Type:        "instantiation",
		})
	}

	// Analyze the constructor
	if classDef.Constructor != nil {
		e.currentDepth = 0
		constructorFlows := e.traceConstructor(classDef, classFile, parsed.PropertyName, parsed.AccessKey)
		flow.Steps = append(flow.Steps, constructorFlows...)

		// Extract ultimate sources from constructor
		flow.Sources = e.extractSources(constructorFlows)
	}

	// PHASE 1.2: Trace EXTERNAL method calls made after instantiation
	// This handles cases like: $mybb->parse_cookies() called in init.php:210
	if instFile != "" {
		externalFlows := e.traceExternalCalls(parsed.VarName, instFile, instLine, parsed.PropertyName, parsed.AccessKey, classDef, classFile)
		if len(externalFlows) > 0 {
			// Renumber steps
			for i := range externalFlows {
				externalFlows[i].StepNumber = len(flow.Steps) + i + 1
			}
			flow.Steps = append(flow.Steps, externalFlows...)
			flow.Sources = append(flow.Sources, e.extractSources(externalFlows)...)
		}
	}

	return flow, nil
}

// traceExternalCalls finds and traces method calls made on a variable AFTER its instantiation
// This is critical for cases like: $mybb = new MyBB(); ... $mybb->parse_cookies();
func (e *ExecutionEngine) traceExternalCalls(varName string, instFile string, instLine int, targetProperty string, accessKey string, classDef *types.ClassDef, classFile string) []FlowStep {
	var steps []FlowStep

	// Get the instantiation file content
	content, ok := e.fileContents[instFile]
	if !ok {
		return steps
	}

	// Parse the file to find method calls on this variable
	root, ok := e.parsedFiles[instFile]
	if !ok {
		return steps
	}

	// Find all method calls on this variable after the instantiation line
	methodCalls := e.findExternalMethodCalls(root, content, varName, instLine)

	for _, mc := range methodCalls {
		// Check if this method exists in the class and might populate our target property
		methodDef, ok := classDef.Methods[mc.methodName]
		if !ok {
			continue
		}

		// Check if this method touches the target property
		if methodDef.BodySource != "" && strings.Contains(methodDef.BodySource, "$this->"+targetProperty) {
			steps = append(steps, FlowStep{
				StepNumber:  len(steps) + 20,
				Description: fmt.Sprintf("External call: %s->%s() at line %d", varName, mc.methodName, mc.line),
				Code:        fmt.Sprintf("%s->%s(%s);", varName, mc.methodName, mc.args),
				FilePath:    instFile,
				Line:        mc.line,
				Type:        "external_call",
			})

			// Trace into this method
			e.currentDepth = 0
			methodSteps := e.traceMethod(classDef, methodDef, classFile, targetProperty, accessKey, mc.args)
			steps = append(steps, methodSteps...)
		}
	}

	return steps
}

// externalMethodCall represents a method call found in a file

// findExternalMethodCalls finds all method calls on a variable after a given line
func (e *ExecutionEngine) findExternalMethodCalls(root *sitter.Node, source []byte, varName string, afterLine int) []externalMethodCall {
	var calls []externalMethodCall

	// Find all member_call_expression nodes
	memberCalls := findNodesOfType(root, "member_call_expression")

	varNameWithoutDollar := strings.TrimPrefix(varName, "$")

	for _, call := range memberCalls {
		callLine := int(call.StartPoint().Row) + 1

		// Only look at calls after instantiation
		if callLine <= afterLine {
			continue
		}

		// Check if this is a call on our variable
		// member_call_expression has: object, ->, name, arguments
		if call.ChildCount() < 3 {
			continue
		}

		// Get the object being called on
		obj := call.Child(0)
		if obj == nil {
			continue
		}

		objText := getNodeText(obj, source)

		// Check if it's our variable (handle both $var and $var forms)
		if objText != varName && objText != "$"+varNameWithoutDollar {
			continue
		}

		// Get method name - it's usually the "name" child
		var methodName string
		var argsText string

		for i := 0; i < int(call.ChildCount()); i++ {
			child := call.Child(i)
			if child == nil {
				continue
			}

			switch child.Type() {
			case "name":
				methodName = getNodeText(child, source)
			case "arguments":
				// Extract arguments without parentheses
				argsText = getNodeText(child, source)
				argsText = strings.TrimPrefix(argsText, "(")
				argsText = strings.TrimSuffix(argsText, ")")
			}
		}

		if methodName != "" {
			calls = append(calls, externalMethodCall{
				methodName: methodName,
				args:       argsText,
				line:       callLine,
			})
		}
	}

	return calls
}

// analyzeMethodReturns analyzes what a method returns
func (e *ExecutionEngine) analyzeMethodReturns(classDef *types.ClassDef, method *types.MethodDef, classFile string) *MethodReturnInfo {
	cacheKey := fmt.Sprintf("%s.%s", classDef.Name, method.Name)
	if cached, ok := e.methodReturns[cacheKey]; ok {
		return cached
	}

	info := &MethodReturnInfo{
		ReturnStatements: make([]string, 0),
	}

	if method.BodySource == "" {
		e.methodReturns[cacheKey] = info
		return info
	}

	body := method.BodySource

	// Find all return statements
	returnMatches := patterns.ReturnStatementPattern.FindAllStringSubmatch(body, -1)

	for _, match := range returnMatches {
		if len(match) >= 2 {
			returnExpr := strings.TrimSpace(match[1])
			info.ReturnStatements = append(info.ReturnStatements, returnExpr)

			// GAP #4 FIX: Check for fluent interface pattern: return $this;
			if returnExpr == "$this" {
				info.ReturnsSelf = true
			}

			// PHASE 2.1: Check if it returns TYPE-CASTED $this->property[$param]
			// Pattern: (int)$this->property[$paramName] or (float)$this->... etc.
			if propMatch := patterns.TypeCastPropertyReturnPattern.FindStringSubmatch(returnExpr); len(propMatch) >= 4 {
				info.ReturnsProperty = true
				info.PropertyName = propMatch[2] // property name
				paramName := propMatch[3]        // param used as key

				// Find which parameter index this is
				for i, p := range method.Parameters {
					if p.Name == paramName {
						info.UsesParamAsKey = true
						info.ParamIndex = i
						break
					}
				}
			}

			// Check if it returns $this->property[$param] (without type cast)
			// Pattern: $this->property[$paramName]
			if !info.ReturnsProperty {
				if propMatch := patterns.PropertyWithParamKeyPattern.FindStringSubmatch(returnExpr); len(propMatch) >= 3 {
					info.ReturnsProperty = true
					info.PropertyName = propMatch[1]
					paramName := propMatch[2]

					// Find which parameter index this is
					for i, p := range method.Parameters {
						if p.Name == paramName {
							info.UsesParamAsKey = true
							info.ParamIndex = i
							break
						}
					}
				}
			}

			// PHASE 2.2: Check for null coalescing pattern
			// Pattern: $this->property[$param] ?? $default
			if !info.ReturnsProperty {
				if propMatch := patterns.NullCoalescePropertyPattern.FindStringSubmatch(returnExpr); len(propMatch) >= 3 {
					info.ReturnsProperty = true
					info.PropertyName = propMatch[1]
					paramName := propMatch[2]

					for i, p := range method.Parameters {
						if p.Name == paramName {
							info.UsesParamAsKey = true
							info.ParamIndex = i
							break
						}
					}
				}
			}

			// PHASE 2.2: Check for ternary isset pattern
			// Pattern: isset($this->property[$param]) ? $this->property[$param] : default
			if !info.ReturnsProperty {
				if propMatch := patterns.TernaryIssetPattern.FindStringSubmatch(returnExpr); len(propMatch) >= 5 {
					// Verify both property refs match
					if propMatch[1] == propMatch[3] && propMatch[2] == propMatch[4] {
						info.ReturnsProperty = true
						info.PropertyName = propMatch[1]
						paramName := propMatch[2]

						for i, p := range method.Parameters {
							if p.Name == paramName {
								info.UsesParamAsKey = true
								info.ParamIndex = i
								break
							}
						}
					}
				}
			}

			// Check if it returns $this->property directly
			if !info.ReturnsProperty {
				if propMatch := patterns.DirectPropertyReturnPattern.FindStringSubmatch(returnExpr); len(propMatch) >= 2 {
					info.ReturnsProperty = true
					info.PropertyName = propMatch[1]
				}
			}

			// Check for superglobals
			for sg := range pkgSources.SuperglobalToSourceType {
				if strings.Contains(returnExpr, sg) {
					info.ReturnsUserInput = true
					info.UserInputExpression = returnExpr
					break
				}
			}
		}
	}

	e.methodReturns[cacheKey] = info
	return info
}

// formatParams formats method parameters for display
func (e *ExecutionEngine) formatParams(params []types.ParameterDef) string {
	var parts []string
	for _, p := range params {
		part := "$" + p.Name
		if p.Type != "" {
			part = p.Type + " " + part
		}
		if p.DefaultValue != "" {
			part += " = " + p.DefaultValue
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ", ")
}

// findInstantiation finds where a variable is instantiated by searching ALL parsed files
// This is fully universal - no framework-specific hints or assumptions
func (e *ExecutionEngine) findInstantiation(varName string, contextFile string) (className, filePath string, line int) {
	// First check the context file (most likely location)
	if root, ok := e.parsedFiles[contextFile]; ok {
		if content, ok := e.fileContents[contextFile]; ok {
			className, line = e.findInstantiationInAST(root, content, varName)
			if className != "" {
				return className, contextFile, line
			}
		}
	}

	// Search ALL files for the instantiation - no assumptions about file names
	for file, root := range e.parsedFiles {
		if file == contextFile {
			continue // Already checked
		}
		if content, ok := e.fileContents[file]; ok {
			className, line = e.findInstantiationInAST(root, content, varName)
			if className != "" {
				return className, file, line
			}
		}
	}

	return "", "", 0
}

// findInstantiationInAST searches an AST for object creation
// Supports:
// - $var = new Class()
// - $GLOBALS['var'] = new Class()
// - $var = $container->get('service') with type hint
func (e *ExecutionEngine) findInstantiationInAST(root *sitter.Node, source []byte, varName string) (className string, line int) {
	// Look for assignment expressions where LHS is varName and RHS is object_creation_expression
	assignments := findNodesOfType(root, "assignment_expression")

	// Strip $ from varName for GLOBALS matching
	varNameWithoutDollar := strings.TrimPrefix(varName, "$")

	for _, assign := range assignments {
		if assign.ChildCount() < 3 {
			continue
		}

		left := assign.Child(0)
		right := assign.Child(2)

		if left == nil || right == nil {
			continue
		}

		leftText := getNodeText(left, source)

		// Check direct assignment: $var = new Class()
		directMatch := leftText == varName

		// Check GLOBALS assignment: $GLOBALS['var'] = new Class()
		globalsMatch := false
		if !directMatch {
			// Uses centralized pattern from phpPatterns
			if matches := phpPatterns.GlobalsPattern.FindStringSubmatch(leftText); len(matches) >= 2 {
				if matches[1] == varNameWithoutDollar {
					globalsMatch = true
				}
			}
		}

		if !directMatch && !globalsMatch {
			continue
		}

		if right.Type() == "object_creation_expression" {
			// Found it! Extract class name
			nameNode := findChildByType(right, "name")
			if nameNode == nil {
				nameNode = findChildByType(right, "qualified_name")
			}
			if nameNode != nil {
				return getNodeText(nameNode, source), int(assign.StartPoint().Row) + 1
			}
		}

		// Check for DI container pattern: $var = $container->get('service')
		// Uses centralized pattern from phpPatterns
		rightText := getNodeText(right, source)
		if phpPatterns.DIContainerPattern.MatchString(rightText) {
			// Found DI container pattern - look for type hint above
			assignLine := int(assign.StartPoint().Row)
			typeHintClass := e.findTypeHintAboveLine(source, assignLine, varNameWithoutDollar)
			if typeHintClass != "" {
				return typeHintClass, assignLine + 1
			}
			// If no type hint, return the service name as a hint
			if matches := phpPatterns.DIContainerPattern.FindStringSubmatch(rightText); len(matches) >= 2 {
				serviceName := matches[1]
				return fmt.Sprintf("[DI:%s]", serviceName), assignLine + 1
			}
		}
	}

	return "", 0
}

// findTypeHintAboveLine searches for PHPDoc @var type hints above a line
// Pattern: /* @var $varname \namespace\classname */ or /** @var \class $var */
func (e *ExecutionEngine) findTypeHintAboveLine(source []byte, targetLine int, varName string) string {
	lines := strings.Split(string(source), "\n")
	if targetLine <= 0 || targetLine > len(lines) {
		return ""
	}

	// Look up to 5 lines above for a type hint
	startLine := targetLine - 5
	if startLine < 0 {
		startLine = 0
	}

	// Pattern 1: /* @var $varname \namespace\classname */
	// Pattern 2: /** @var \namespace\classname $varname */
	// Uses centralized pattern builder from phpPatterns
	typeHintPatterns := phpPatterns.GetTypeHintPatterns(varName)

	for i := targetLine - 1; i >= startLine; i-- {
		line := lines[i]
		for _, pattern := range typeHintPatterns {
			if matches := pattern.FindStringSubmatch(line); len(matches) >= 2 {
				// Extract class name from fully qualified name
				fqn := matches[1]
				parts := strings.Split(fqn, "\\")
				return parts[len(parts)-1] // Return just the class name
			}
		}
	}

	return ""
}

// findClassDefinition finds a class definition across all symbol tables
// Handles interfaces by stripping _interface suffix and looking for implementing class
func (e *ExecutionEngine) findClassDefinition(className string) (*types.ClassDef, string) {
	// First try exact match
	for filePath, st := range e.symbolTables {
		if classDef, ok := st.Classes[className]; ok {
			return classDef, filePath
		}
	}

	// Try case-insensitive match
	lowerClassName := strings.ToLower(className)
	for filePath, st := range e.symbolTables {
		for name, classDef := range st.Classes {
			if strings.ToLower(name) == lowerClassName {
				return classDef, filePath
			}
		}
	}

	// If name ends with _interface, try stripping it
	if strings.HasSuffix(lowerClassName, "_interface") {
		baseName := strings.TrimSuffix(className, "_interface")
		baseName = strings.TrimSuffix(baseName, "_Interface")
		for filePath, st := range e.symbolTables {
			for name, classDef := range st.Classes {
				if strings.EqualFold(name, baseName) {
					return classDef, filePath
				}
			}
		}
	}

	// Try to find a class that implements this interface
	for filePath, st := range e.symbolTables {
		for _, classDef := range st.Classes {
			for _, iface := range classDef.Implements {
				if strings.EqualFold(iface, className) {
					return classDef, filePath
				}
			}
		}
	}

	return nil, ""
}

// traceConstructor traces through a constructor to find property population
func (e *ExecutionEngine) traceConstructor(classDef *types.ClassDef, classFile string, targetProperty string, accessKey string) []FlowStep {
	var steps []FlowStep

	if classDef.Constructor == nil {
		return steps
	}

	constructor := classDef.Constructor

	steps = append(steps, FlowStep{
		StepNumber:  len(steps) + 2,
		Description: "Constructor runs",
		Code:        fmt.Sprintf("function __construct() { ... }"),
		FilePath:    classFile,
		Line:        constructor.Line,
		Type:        "constructor_call",
	})

	// Parse the constructor body
	if constructor.BodySource == "" {
		return steps
	}

	// Look for method calls that might populate the property
	// Parse: $this->methodName($arg) - uses centralized pattern
	methodCalls := phpPatterns.ThisMethodCallPattern.FindAllStringSubmatch(constructor.BodySource, -1)

	for _, call := range methodCalls {
		methodName := call[1]
		methodArgs := call[2]

		// Check if this method populates our target property
		if methodDef, ok := classDef.Methods[methodName]; ok {
			// Trace into the method FIRST to see if it affects target property
			methodSteps := e.traceMethod(classDef, methodDef, classFile, targetProperty, accessKey, methodArgs)

			// Only add method call step if method actually affects the target property
			if len(methodSteps) > 0 {
				steps = append(steps, FlowStep{
					StepNumber:  len(steps) + 2,
					Description: fmt.Sprintf("Calls $this->%s(%s)", methodName, methodArgs),
					Code:        fmt.Sprintf("$this->%s(%s);", methodName, methodArgs),
					FilePath:    classFile,
					Line:        e.findLineInBody(constructor.BodySource, constructor.BodyStart, methodName),
					Type:        "method_call",
				})
				steps = append(steps, methodSteps...)
			}
		}
	}

	// Also trace the constructor body directly for direct assignments
	// This handles cases like: if($_SERVER['REQUEST_METHOD'] == "POST") { $this->request_method = "post"; }
	constructorAsMethod := &types.MethodDef{
		Name:       "__construct",
		Line:       constructor.Line,
		BodySource: constructor.BodySource,
		BodyStart:  constructor.BodyStart,
	}
	// Reset depth to allow analyzing the constructor body directly
	// This is not a recursive call, just analyzing the current body
	savedDepth := e.currentDepth
	e.currentDepth = 0
	directSteps := e.traceMethod(classDef, constructorAsMethod, classFile, targetProperty, accessKey, "")
	e.currentDepth = savedDepth
	steps = append(steps, directSteps...)

	return steps
}

// traceMethod traces through a method to find property assignments
func (e *ExecutionEngine) traceMethod(classDef *types.ClassDef, method *types.MethodDef, classFile string, targetProperty string, accessKey string, callArgs string) []FlowStep {
	var steps []FlowStep

	e.currentDepth++
	if e.currentDepth > e.maxDepth {
		return steps
	}

	if method.BodySource == "" {
		return steps
	}

	body := method.BodySource

	// PHASE 1.1: Look for foreach loops iterating over SUPERGLOBALS directly
	// Pattern: foreach($_SUPERGLOBAL as $key => $val)
	// This handles methods like parse_cookies() that don't take parameters
	// Uses centralized pattern from common
	superglobalMatches := common.SuperglobalForeachPattern.FindAllStringSubmatch(body, -1)

	for _, match := range superglobalMatches {
		superglobalName := match[1] // e.g., "$_COOKIE"
		keyVar := match[2]          // e.g., "key"
		valVar := match[3]          // e.g., "val"

		// Check if this superglobal is a known user input source
		var sourceType string
		for sg, sgType := range pkgSources.SuperglobalToSourceType {
			if superglobalName == sg {
				sourceType = string(sgType)
				break
			}
		}

		if sourceType != "" {
			// Look for property assignment inside the loop FIRST
			// Pattern: $this->property[$key] = $val
			propAssignPattern := patterns.BuildPropertyAssignInLoopPattern(keyVar, valVar)
			propMatches := propAssignPattern.FindAllStringSubmatch(body, -1)

			for _, propMatch := range propMatches {
				assignedProperty := propMatch[1]
				// Only add steps if this assigns to our target property
				if assignedProperty == targetProperty {
					// Add the loop step only if we're assigning to target property
					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Inside %s() - loops through %s superglobal", method.Name, superglobalName),
						Code:        fmt.Sprintf("foreach(%s as $%s => $%s)", superglobalName, keyVar, valVar),
						FilePath:    classFile,
						Line:        e.findLineInBody(body, method.BodyStart, "foreach"),
						Type:        "loop",
					})

					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Assigns $this->%s[$%s] = $%s from %s", targetProperty, keyVar, valVar, superglobalName),
						Code:        fmt.Sprintf("$this->%s[$%s] = $%s;", targetProperty, keyVar, valVar),
						FilePath:    classFile,
						Line:        e.findLineInBody(body, method.BodyStart, "$this->"+targetProperty),
						Type:        "assignment",
					})

					// The ultimate source is the superglobal
					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Result: $%s['%s'] now contains %s['%s']", targetProperty, accessKey, superglobalName, accessKey),
						Code:        fmt.Sprintf("// $this->%s['%s'] = %s['%s']", targetProperty, accessKey, superglobalName, accessKey),
						FilePath:    classFile,
						Line:        0,
						Type:        "result",
					})
				}
			}
		}
	}

	// Look for foreach loops iterating over the method parameter
	// Pattern: foreach($array as $key => $val) - uses centralized pattern
	foreachMatches := phpPatterns.ForeachPattern.FindAllStringSubmatch(body, -1)

	for _, match := range foreachMatches {
		arrayVar := match[1]
		keyVar := match[2]
		valVar := match[3]

		// Skip if this is a superglobal (already handled above)
		if strings.HasPrefix(arrayVar, "_") {
			continue
		}

		// Check if the array variable matches the parameter
		if len(method.Parameters) > 0 && method.Parameters[0].Name == arrayVar {
			// Look for property assignment inside the loop FIRST
			// Pattern: $this->property[$key] = $val
			propAssignPattern := patterns.BuildPropertyAssignInLoopPattern(keyVar, valVar)
			propMatches := propAssignPattern.FindAllStringSubmatch(body, -1)

			for _, propMatch := range propMatches {
				assignedProperty := propMatch[1]
				// Only add steps if this assigns to our target property
				if assignedProperty == targetProperty {
					// Add the loop step only if we're assigning to target property
					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Inside %s() - loops through %s parameter", method.Name, callArgs),
						Code:        fmt.Sprintf("foreach(%s as $%s => $%s)", callArgs, keyVar, valVar),
						FilePath:    classFile,
						Line:        e.findLineInBody(body, method.BodyStart, "foreach"),
						Type:        "loop",
					})

					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Assigns $this->%s[$%s] = $%s from %s", targetProperty, keyVar, valVar, callArgs),
						Code:        fmt.Sprintf("$this->%s[$%s] = $%s;", targetProperty, keyVar, valVar),
						FilePath:    classFile,
						Line:        e.findLineInBody(body, method.BodyStart, "$this->"+targetProperty),
						Type:        "assignment",
					})

					// The ultimate source is the call argument
					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Result: $%s['%s'] now contains %s['%s']", targetProperty, accessKey, callArgs, accessKey),
						Code:        fmt.Sprintf("// $this->%s['%s'] = %s['%s']", targetProperty, accessKey, callArgs, accessKey),
						FilePath:    classFile,
						Line:        0,
						Type:        "result",
					})
				}
			}
		}
	}

	// Also look for direct assignments
	// Pattern: $this->property = $something
	// Uses centralized pattern builder from phpPatterns
	directAssignPattern := phpPatterns.BuildDirectAssignPattern(targetProperty)
	directMatches := directAssignPattern.FindAllStringSubmatch(body, -1)

	for _, match := range directMatches {
		source := strings.TrimSpace(match[1])
		steps = append(steps, FlowStep{
			StepNumber:  len(steps) + 10,
			Description: fmt.Sprintf("Assigns $this->%s = %s", targetProperty, source),
			Code:        fmt.Sprintf("$this->%s = %s;", targetProperty, source),
			FilePath:    classFile,
			Line:        e.findLineInBody(body, method.BodyStart, "$this->"+targetProperty),
			Type:        "assignment",
		})
	}

	// NEW: Look for conditional assignments based on superglobals
	// Pattern: if($_SUPERGLOBAL['key']... { $this->property = value }
	// This handles cases like: if($_SERVER['REQUEST_METHOD'] == "POST") { $this->request_method = "post"; }
	for sg := range pkgSources.SuperglobalToSourceType {
		if strings.Contains(body, sg) && strings.Contains(body, "$this->"+targetProperty) {
			// Check if superglobal is used in a condition and property is assigned nearby
			// Pattern: if($_SUPERGLOBAL[anything]) - uses centralized pattern builder
			condPattern := phpPatterns.BuildConditionalPattern(sg)
			if condMatches := condPattern.FindStringSubmatch(body); len(condMatches) >= 2 {
				superglobalKey := condMatches[1]
				steps = append(steps, FlowStep{
					StepNumber:  len(steps) + 10,
					Description: fmt.Sprintf("Conditional on %s['%s']", sg, superglobalKey),
					Code:        fmt.Sprintf("if(%s['%s'] == ...) { $this->%s = ...; }", sg, superglobalKey, targetProperty),
					FilePath:    classFile,
					Line:        e.findLineInBody(body, method.BodyStart, sg),
					Type:        "conditional",
				})
				steps = append(steps, FlowStep{
					StepNumber:  len(steps) + 10,
					Description: fmt.Sprintf("Property $%s is controlled by %s['%s']", targetProperty, sg, superglobalKey),
					Code:        fmt.Sprintf("// $this->%s value depends on %s['%s']", targetProperty, sg, superglobalKey),
					FilePath:    classFile,
					Line:        0,
					Type:        "taint",
				})
			}
		}
	}

	return steps
}

// findLineInBody finds the approximate line number for a pattern in the body
func (e *ExecutionEngine) findLineInBody(body string, startLine int, pattern string) int {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if strings.Contains(line, pattern) {
			return startLine + i
		}
	}
	return startLine
}

// extractSources extracts ultimate sources from the flow steps
// Uses centralized superglobal definitions from pkg/sources
