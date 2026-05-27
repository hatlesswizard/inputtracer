package symbolic

import (
	"fmt"
	"strings"

	pkgSources "github.com/hatlesswizard/inputtracer/pkg/sources"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"
)

func (e *ExecutionEngine) extractSources(steps []FlowStep) []UltimateSource {
	var result []UltimateSource

	for _, step := range steps {
		for sg, sourceType := range pkgSources.SuperglobalToSourceType {
			if strings.Contains(step.Code, sg) || strings.Contains(step.Description, sg) {
				result = append(result, UltimateSource{
					Type:       string(sourceType),
					Expression: sg,
					FilePath:   step.FilePath,
					Line:       step.Line,
				})
			}
		}
	}

	return result
}

// GenerateFlowReport generates a human-readable flow report
func (flow *PropertyFlow) GenerateFlowReport() string {
	var sb strings.Builder

	sb.WriteString("=== Input Flow Trace ===\n\n")
	sb.WriteString(fmt.Sprintf("Expression: %s\n", flow.Expression))
	sb.WriteString(fmt.Sprintf("Class: %s\n", flow.ClassName))
	if flow.MethodName != "" {
		sb.WriteString(fmt.Sprintf("Method: %s()\n", flow.MethodName))
	}
	if flow.PropertyName != "" {
		sb.WriteString(fmt.Sprintf("Property: %s\n", flow.PropertyName))
	}
	if flow.AccessKey != "" {
		sb.WriteString(fmt.Sprintf("Access Key: '%s'\n", flow.AccessKey))
	}
	sb.WriteString("\n--- Flow Steps ---\n\n")

	for _, step := range flow.Steps {
		sb.WriteString(fmt.Sprintf("Step %d: %s\n", step.StepNumber, step.Description))
		sb.WriteString(fmt.Sprintf("   Code: %s\n", step.Code))
		if step.FilePath != "" && step.Line > 0 {
			sb.WriteString(fmt.Sprintf("   Location: %s:%d\n", step.FilePath, step.Line))
		}
		sb.WriteString("\n")
	}

	if len(flow.Sources) > 0 {
		sb.WriteString("--- Ultimate Sources ---\n\n")
		seen := make(map[string]bool)
		for _, src := range flow.Sources {
			key := fmt.Sprintf("%s:%d", src.Expression, src.Line)
			if seen[key] {
				continue
			}
			seen[key] = true
			sb.WriteString(fmt.Sprintf("  %s (%s)\n", src.Expression, src.Type))
			if src.FilePath != "" && src.Line > 0 {
				sb.WriteString(fmt.Sprintf("    at %s:%d\n", src.FilePath, src.Line))
			}
		}
	}

	return sb.String()
}

// GenerateMermaidDiagram generates a Mermaid flowchart for the flow
func (flow *PropertyFlow) GenerateMermaidDiagram() string {
	var sb strings.Builder

	sb.WriteString("flowchart TD\n")
	sb.WriteString("    classDef source fill:#ff6b6b,stroke:#c0392b,color:white\n")
	sb.WriteString("    classDef property fill:#4ecdc4,stroke:#16a085\n")
	sb.WriteString("    classDef method fill:#45b7d1,stroke:#2980b9\n")
	sb.WriteString("    classDef result fill:#95e1a3,stroke:#27ae60\n\n")

	// Add nodes for each step
	prevNode := ""
	for i, step := range flow.Steps {
		nodeID := fmt.Sprintf("step%d", i)
		label := strings.ReplaceAll(step.Description, "\"", "'")
		label = strings.ReplaceAll(label, "$", "\\$")

		nodeClass := ""
		switch step.Type {
		case "property_init":
			nodeClass = ":::property"
		case "method_call", "constructor_call", "method_def":
			nodeClass = ":::method"
		case "assignment":
			nodeClass = ":::source"
		case "result", "resolution":
			nodeClass = ":::result"
		}

		sb.WriteString(fmt.Sprintf("    %s[\"%s\"]%s\n", nodeID, label, nodeClass))

		if prevNode != "" {
			sb.WriteString(fmt.Sprintf("    %s --> %s\n", prevNode, nodeID))
		}
		prevNode = nodeID
	}

	// Add source nodes
	seen := make(map[string]bool)
	for i, src := range flow.Sources {
		if seen[src.Expression] {
			continue
		}
		seen[src.Expression] = true
		srcID := fmt.Sprintf("source%d", i)
		sb.WriteString(fmt.Sprintf("    %s((%s)):::source\n", srcID, src.Expression))
		sb.WriteString(fmt.Sprintf("    %s -.->|\"originates from\"| %s\n", srcID, prevNode))
	}

	return sb.String()
}

// Helper functions

func findNodesOfType(root *sitter.Node, nodeType string) []*sitter.Node {
	var nodes []*sitter.Node
	traverseTree(root, func(node *sitter.Node) bool {
		if node.Type() == nodeType {
			nodes = append(nodes, node)
		}
		return true
	})
	return nodes
}

func traverseTree(node *sitter.Node, callback func(*sitter.Node) bool) {
	if node == nil {
		return
	}
	if !callback(node) {
		return
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		traverseTree(node.Child(i), callback)
	}
}

func findChildByType(node *sitter.Node, nodeType string) *sitter.Node {
	if node == nil {
		return nil
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == nodeType {
			return child
		}
	}
	return nil
}

func getNodeText(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if start >= uint32(len(source)) || end > uint32(len(source)) {
		return ""
	}
	return string(source[start:end])
}

// CreateParser creates a new PHP parser
func CreateParser() *sitter.Parser {
	parser := sitter.NewParser()
	parser.SetLanguage(php.GetLanguage())
	return parser
}
