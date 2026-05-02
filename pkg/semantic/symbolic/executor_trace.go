package symbolic

import (
	"fmt"
	"regexp"
	"strings"

	pkgSources "github.com/hatlesswizard/inputtracer/pkg/sources"
	"github.com/hatlesswizard/inputtracer/pkg/sources/patterns"
)

func (e *ExecutionEngine) TracePropertyAccess(expression string, contextFile string) (*PropertyFlow, error) {
	// Parse the expression to determine its type
	parsed := e.parseExpression(expression)
	if parsed.Type == ExprTypeUnknown {
		return nil, fmt.Errorf("could not parse expression: %s", expression)
	}

	flow := &PropertyFlow{
		Expression: expression,
		Steps:      make([]FlowStep, 0),
		Sources:    make([]UltimateSource, 0),
	}

	// GAP #1 FIX: Handle direct superglobal access
	if parsed.Type == ExprTypeSuperglobal {
		return e.traceSuperglobal(parsed, flow)
	}

	// GAP #2 FIX: Handle local variable tracing
	if parsed.Type == ExprTypeLocalVariable {
		return e.traceLocalVariable(parsed, contextFile, flow)
	}

	// GAP #3 FIX: Handle static method/property calls
	if parsed.Type == ExprTypeStaticCall {
		return e.traceStaticCall(parsed, flow)
	}
	if parsed.Type == ExprTypeStaticProperty {
		return e.traceStaticProperty(parsed, flow)
	}

	// For object-based expressions, find instantiation
	className, instantiationFile, instantiationLine := e.findInstantiation(parsed.VarName, contextFile)
	if className == "" {
		return nil, fmt.Errorf("could not find instantiation of variable %s (searched %d files)", parsed.VarName, len(e.parsedFiles))
	}

	parsed.ClassName = className
	flow.ClassName = className

	// Find the class definition
	classDef, classFile := e.findClassDefinition(className)
	if classDef == nil {
		return nil, fmt.Errorf("could not find class definition for %s", className)
	}

	// GAP #4 FIX: Handle chained expressions like $obj->method()->property
	if parsed.IsChained {
		return e.traceChainedExpression(parsed, classDef, classFile, instantiationFile, instantiationLine, flow)
	}

	switch parsed.Type {
	case ExprTypeMethodCall:
		return e.traceMethodCall(parsed, classDef, classFile, instantiationFile, instantiationLine, flow)
	case ExprTypePropertyAccess:
		return e.tracePropertyAccessExpr(parsed, classDef, classFile, instantiationFile, instantiationLine, flow)
	default:
		return nil, fmt.Errorf("unsupported expression type: %v", parsed.Type)
	}
}

// traceSuperglobal handles direct superglobal access like $_GET['id']
func (e *ExecutionEngine) traceSuperglobal(parsed *ParsedExpression, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.PropertyName = parsed.AccessKey

	// Determine the source type using centralized mappings
	sourceType := pkgSources.GetSuperglobalSourceType(parsed.SuperglobalName)

	// Add step showing direct superglobal access
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Direct superglobal access: %s['%s']", parsed.SuperglobalName, parsed.AccessKey),
		Code:        fmt.Sprintf("%s['%s']", parsed.SuperglobalName, parsed.AccessKey),
		FilePath:    "",
		Line:        0,
		Type:        "direct_input",
	})

	// Add the source
	flow.Sources = append(flow.Sources, UltimateSource{
		Type:       string(sourceType),
		Expression: fmt.Sprintf("%s['%s']", parsed.SuperglobalName, parsed.AccessKey),
		FilePath:   "",
		Line:       0,
	})

	return flow, nil
}

// traceLocalVariable traces a local variable to find its source
func (e *ExecutionEngine) traceLocalVariable(parsed *ParsedExpression, contextFile string, flow *PropertyFlow) (*PropertyFlow, error) {
	varName := parsed.VarName

	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Tracing local variable: %s", varName),
		Code:        varName,
		FilePath:    "",
		Line:        0,
		Type:        "local_var",
	})

	// Search all files for assignments to this variable from superglobals
	foundAssignments := e.findVariableAssignments(varName)

	for i, assignment := range foundAssignments {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  i + 2,
			Description: fmt.Sprintf("Assignment found: %s = %s", varName, assignment.source),
			Code:        fmt.Sprintf("%s = %s;", varName, assignment.source),
			FilePath:    assignment.file,
			Line:        assignment.line,
			Type:        "assignment",
		})

		// Check if source is a superglobal
		for sg, sgType := range pkgSources.SuperglobalToSourceType {
			if strings.Contains(assignment.source, sg) {
				flow.Sources = append(flow.Sources, UltimateSource{
					Type:       string(sgType),
					Expression: assignment.source,
					FilePath:   assignment.file,
					Line:       assignment.line,
				})
			}
		}
	}

	if len(foundAssignments) == 0 {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("No assignments found for %s in parsed files", varName),
			Code:        "",
			FilePath:    "",
			Line:        0,
			Type:        "not_found",
		})
	}

	return flow, nil
}

// variableAssignment represents an assignment to a variable

// findVariableAssignments searches for assignments to a variable
func (e *ExecutionEngine) findVariableAssignments(varName string) []variableAssignment {
	var assignments []variableAssignment

	// Remove $ prefix for matching
	varNameClean := strings.TrimPrefix(varName, "$")

	// Search all file contents for assignments
	for file, content := range e.fileContents {
		// Pattern: $varname = something
		assignPattern := patterns.BuildVariableAssignPattern(varNameClean)
		lines := strings.Split(string(content), "\n")

		for lineNum, line := range lines {
			if matches := assignPattern.FindStringSubmatch(line); len(matches) >= 2 {
				assignments = append(assignments, variableAssignment{
					source: strings.TrimSpace(matches[1]),
					file:   file,
					line:   lineNum + 1,
				})
			}
		}
	}

	return assignments
}

// findExternalPropertyAssignments searches all parsed files for external assignments
// to an object's property, like: $varName->propertyName = value;
// This handles dynamic properties assigned outside the class definition
func (e *ExecutionEngine) findExternalPropertyAssignments(varName string, propertyName string) []ExternalAssignment {
	var assignments []ExternalAssignment

	varNameWithoutDollar := strings.TrimPrefix(varName, "$")

	// Search all file contents for external property assignments
	for file, content := range e.fileContents {
		// Pattern: $varname->property = something
		// or $varname->property['key'] = something
		assignPatterns := []*regexp.Regexp{
			// $var->property = value;
			patterns.BuildPropertyExternalAssignPattern(varNameWithoutDollar, propertyName),
			// $var->property['key'] = value; (array element assignment)
			patterns.BuildPropertyArrayExternalAssignPattern(varNameWithoutDollar, propertyName),
		}

		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			for _, pattern := range assignPatterns {
				if matches := pattern.FindStringSubmatch(line); len(matches) >= 2 {
					assignments = append(assignments, ExternalAssignment{
						PropertyName: propertyName,
						Source:       strings.TrimSpace(matches[1]),
						FilePath:     file,
						Line:         lineNum + 1,
					})
				}
			}
		}
	}

	return assignments
}

// traceExternalPropertyAssignment traces an externally assigned property
// This is for properties like $mybb->post_code that are assigned outside the class
func (e *ExecutionEngine) traceExternalPropertyAssignment(parsed *ParsedExpression, externalAssignments []ExternalAssignment, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.PropertyName = parsed.PropertyName
	flow.AccessKey = parsed.AccessKey

	// Show that this is a dynamic property (not defined in class)
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Dynamic property $%s (not defined in class)", parsed.PropertyName),
		Code:        fmt.Sprintf("// $%s is assigned externally, not in class definition", parsed.PropertyName),
		FilePath:    "",
		Line:        0,
		Type:        "dynamic_property",
	})

	// Show all external assignments
	for i, assign := range externalAssignments {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  i + 2,
			Description: fmt.Sprintf("External assignment: %s->%s = %s", parsed.VarName, parsed.PropertyName, assign.Source),
			Code:        fmt.Sprintf("%s->%s = %s;", parsed.VarName, parsed.PropertyName, assign.Source),
			FilePath:    assign.FilePath,
			Line:        assign.Line,
			Type:        "external_assignment",
		})

		// Check if source contains superglobals
		for sg, sgType := range pkgSources.SuperglobalToSourceType {
			if strings.Contains(assign.Source, sg) {
				flow.Sources = append(flow.Sources, UltimateSource{
					Type:       string(sgType),
					Expression: sg,
					FilePath:   assign.FilePath,
					Line:       assign.Line,
				})
			}
		}

		// If source is a function call, trace into that function
		if matches := patterns.FunctionCallPattern.FindStringSubmatch(assign.Source); len(matches) >= 2 {
			funcName := matches[1]
			funcArgs := ""
			if len(matches) >= 3 {
				funcArgs = matches[2]
			}

			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  len(flow.Steps) + 1,
				Description: fmt.Sprintf("Calls function: %s(%s)", funcName, funcArgs),
				Code:        fmt.Sprintf("%s(%s)", funcName, funcArgs),
				FilePath:    assign.FilePath,
				Line:        assign.Line,
				Type:        "function_call",
			})

			// Try to find and trace the function
			funcSources := e.traceFunctionForSources(funcName)
			flow.Sources = append(flow.Sources, funcSources...)
		}

		// Check for variable assignments (e.g., $mybb->request_method = $_SERVER['REQUEST_METHOD'])
		if strings.HasPrefix(assign.Source, "$_") {
			// Direct superglobal assignment
			for sg, sgType := range pkgSources.SuperglobalToSourceType {
				if strings.Contains(assign.Source, sg) {
					flow.Sources = append(flow.Sources, UltimateSource{
						Type:       string(sgType),
						Expression: assign.Source,
						FilePath:   assign.FilePath,
						Line:       assign.Line,
					})
				}
			}
		}
	}

	return flow, nil
}

// traceFunctionForSources traces a function to find what superglobals it uses
func (e *ExecutionEngine) traceFunctionForSources(funcName string) []UltimateSource {
	var sources []UltimateSource

	// Search all symbol tables for the function
	for filePath, st := range e.symbolTables {
		if funcDef, ok := st.Functions[funcName]; ok {
			// Check function body for superglobal usage
			if funcDef.BodySource != "" {
				for sg, sgType := range pkgSources.SuperglobalToSourceType {
					if strings.Contains(funcDef.BodySource, sg) {
						sources = append(sources, UltimateSource{
							Type:       string(sgType),
							Expression: sg,
							FilePath:   filePath,
							Line:       funcDef.Line,
						})
					}
				}
			}
		}
	}

	return sources
}

// MagicPropertyInfo holds information about magic property access patterns
