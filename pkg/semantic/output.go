package semantic

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
)

// ToJSON converts the trace result to JSON format
func ToJSON(r *TraceResult) (string, error) {
	output := struct {
		Stats struct {
			FilesScanned   int     `json:"files_scanned"`
			FilesParsed    int     `json:"files_parsed"`
			ParseErrors    int     `json:"parse_errors"`
			SourcesFound   int     `json:"sources_found"`
			FlowsTraced    int     `json:"flows_traced"`
			CrossFileFlows int     `json:"cross_file_flows"`
			DurationMs     float64 `json:"duration_ms"`
		} `json:"stats"`
		Sources []struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Name       string `json:"name"`
			File       string `json:"file"`
			Line       int    `json:"line"`
			Column     int    `json:"column"`
			SourceType string `json:"source_type"`
			SourceKey  string `json:"source_key,omitempty"`
			Snippet    string `json:"snippet"`
		} `json:"sources"`
		Nodes []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Name     string `json:"name"`
			File     string `json:"file"`
			Line     int    `json:"line"`
			Snippet  string `json:"snippet"`
			Language string `json:"language"`
		} `json:"nodes"`
		Edges []struct {
			From  string `json:"from"`
			To    string `json:"to"`
			Type  string `json:"type"`
			Label string `json:"label"`
		} `json:"edges"`
		ByLanguage map[string]struct {
			Files   int `json:"files"`
			Sources int `json:"sources"`
		} `json:"by_language"`
	}{}

	// Stats
	output.Stats.FilesScanned = r.Stats.FilesScanned
	output.Stats.FilesParsed = r.Stats.FilesParsed
	output.Stats.ParseErrors = r.Stats.ParseErrors
	output.Stats.SourcesFound = r.Stats.SourcesFound
	output.Stats.FlowsTraced = r.Stats.FlowsTraced
	output.Stats.CrossFileFlows = r.Stats.CrossFileFlows
	output.Stats.DurationMs = r.Stats.TotalDuration.Seconds() * 1000

	// Sources
	for _, src := range r.Sources {
		output.Sources = append(output.Sources, struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Name       string `json:"name"`
			File       string `json:"file"`
			Line       int    `json:"line"`
			Column     int    `json:"column"`
			SourceType string `json:"source_type"`
			SourceKey  string `json:"source_key,omitempty"`
			Snippet    string `json:"snippet"`
		}{
			ID:         src.ID,
			Type:       string(src.Type),
			Name:       src.Name,
			File:       src.FilePath,
			Line:       src.Line,
			Column:     src.Column,
			SourceType: string(src.SourceType),
			SourceKey:  src.SourceKey,
			Snippet:    src.Snippet,
		})
	}

	// Nodes
	for _, node := range r.FlowMap.AllNodes {
		output.Nodes = append(output.Nodes, struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Name     string `json:"name"`
			File     string `json:"file"`
			Line     int    `json:"line"`
			Snippet  string `json:"snippet"`
			Language string `json:"language"`
		}{
			ID:       node.ID,
			Type:     string(node.Type),
			Name:     node.Name,
			File:     node.FilePath,
			Line:     node.Line,
			Snippet:  node.Snippet,
			Language: node.Language,
		})
	}

	// Edges
	for _, edge := range r.FlowMap.AllEdges {
		output.Edges = append(output.Edges, struct {
			From  string `json:"from"`
			To    string `json:"to"`
			Type  string `json:"type"`
			Label string `json:"label"`
		}{
			From:  edge.From,
			To:    edge.To,
			Type:  string(edge.Type),
			Label: edge.Description,
		})
	}

	// By language
	output.ByLanguage = make(map[string]struct {
		Files   int `json:"files"`
		Sources int `json:"sources"`
	})
	for lang, stats := range r.Stats.ByLanguage {
		output.ByLanguage[lang] = struct {
			Files   int `json:"files"`
			Sources int `json:"sources"`
		}{
			Files:   stats.Files,
			Sources: stats.Sources,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToDOT converts the trace result to GraphViz DOT format
func ToDOT(r *TraceResult) string {
	var sb strings.Builder

	sb.WriteString("digraph InputFlows {\n")
	sb.WriteString("  rankdir=TB;\n")
	sb.WriteString("  node [shape=box, fontname=\"Courier\"];\n")
	sb.WriteString("  edge [fontname=\"Arial\", fontsize=10];\n\n")

	// Define node styles
	sb.WriteString("  // Node styles\n")
	sb.WriteString("  subgraph {\n")
	sb.WriteString("    node [style=filled];\n")
	sb.WriteString("  }\n\n")

	// Create nodes
	nodeStyles := map[types.FlowNodeType]string{
		types.NodeSource:    "[shape=ellipse, style=filled, fillcolor=\"#ff6b6b\", fontcolor=white]",
		types.NodeVariable:  "[shape=box, style=filled, fillcolor=\"#4ecdc4\"]",
		types.NodeFunction:  "[shape=box, style=filled, fillcolor=\"#45b7d1\"]",
		types.NodeCarrier:   "[shape=box, style=filled, fillcolor=\"#9b59b6\", fontcolor=white]",
		types.NodeParam: "[shape=box, style=filled, fillcolor=\"#2ecc71\"]",
	}

	sb.WriteString("  // Nodes\n")
	for _, node := range r.FlowMap.AllNodes {
		id := sanitizeDotID(node.ID)
		label := sanitizeDotLabel(fmt.Sprintf("%s\\n%s:%d", node.Name, shortPath(node.FilePath), node.Line))
		style := nodeStyles[node.Type]
		if style == "" {
			style = "[shape=box]"
		}
		sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\"] %s;\n", id, label, style))
	}

	sb.WriteString("\n  // Edges\n")
	for _, edge := range r.FlowMap.AllEdges {
		from := sanitizeDotID(edge.From)
		to := sanitizeDotID(edge.To)
		label := sanitizeDotLabel(edge.Description)

		edgeStyle := ""
		switch edge.Type {
		case types.EdgeAssignment:
			edgeStyle = "[color=\"#3498db\"]"
		case types.EdgeCall:
			edgeStyle = "[color=\"#e74c3c\", style=dashed]"
		case types.EdgeDataFlow:
			edgeStyle = "[color=\"#2ecc71\"]"
		}

		sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"] %s;\n", from, to, label, edgeStyle))
	}

	sb.WriteString("\n  // Legend\n")
	sb.WriteString("  subgraph cluster_legend {\n")
	sb.WriteString("    label=\"Legend\";\n")
	sb.WriteString("    style=dashed;\n")
	sb.WriteString("    legend_source [label=\"Input Source\", shape=ellipse, style=filled, fillcolor=\"#ff6b6b\", fontcolor=white];\n")
	sb.WriteString("    legend_variable [label=\"Variable\", shape=box, style=filled, fillcolor=\"#4ecdc4\"];\n")
	sb.WriteString("    legend_function [label=\"Function\", shape=box, style=filled, fillcolor=\"#45b7d1\"];\n")
	sb.WriteString("  }\n")

	sb.WriteString("}\n")

	return sb.String()
}

// ToMermaid converts the trace result to Mermaid diagram format
func ToMermaid(r *TraceResult) string {
	var sb strings.Builder

	sb.WriteString("flowchart TB\n")

	// Define styles
	sb.WriteString("    %% Style definitions\n")
	sb.WriteString("    classDef source fill:#ff6b6b,stroke:#c0392b,color:white\n")
	sb.WriteString("    classDef variable fill:#4ecdc4,stroke:#16a085\n")
	sb.WriteString("    classDef function fill:#45b7d1,stroke:#2980b9\n")
	sb.WriteString("    classDef carrier fill:#9b59b6,stroke:#8e44ad,color:white\n\n")

	// Create nodes
	nodeClasses := make(map[string]string)
	for _, node := range r.FlowMap.AllNodes {
		id := sanitizeMermaidID(node.ID)
		label := fmt.Sprintf("%s<br/>%s:%d", node.Name, shortPath(node.FilePath), node.Line)

		shape := "[%s]"
		class := "variable"
		switch node.Type {
		case types.NodeSource:
			shape = "((%s))"
			class = "source"
		case types.NodeFunction:
			shape = "{{%s}}"
			class = "function"
		case types.NodeCarrier:
			shape = "[/%s/]"
			class = "carrier"
		}

		nodeClasses[id] = class
		sb.WriteString(fmt.Sprintf("    %s%s\n", id, fmt.Sprintf(shape, sanitizeMermaidLabel(label))))
	}

	sb.WriteString("\n")

	// Create edges
	for _, edge := range r.FlowMap.AllEdges {
		from := sanitizeMermaidID(edge.From)
		to := sanitizeMermaidID(edge.To)
		label := sanitizeMermaidLabel(edge.Description)

		arrow := "-->"
		switch edge.Type {
		case types.EdgeCall:
			arrow = "-.->|" + label + "|"
		case types.EdgeAssignment:
			arrow = "-->|" + label + "|"
		case types.EdgeDataFlow:
			arrow = "==>|" + label + "|"
		}

		if edge.Type == types.EdgeCall {
			sb.WriteString(fmt.Sprintf("    %s %s %s\n", from, arrow, to))
		} else {
			sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n", from, label, to))
		}
	}

	sb.WriteString("\n")

	// Apply classes
	for id, class := range nodeClasses {
		sb.WriteString(fmt.Sprintf("    class %s %s\n", id, class))
	}

	return sb.String()
}

// ToHTML converts the trace result to interactive HTML format
func ToHTML(r *TraceResult) string {
	mermaidDiagram := ToMermaid(r)

	jsonData, _ := ToJSON(r)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>InputTracer - Flow Analysis</title>
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #1a1a2e; color: #eee; }
        .container { max-width: 1400px; margin: 0 auto; padding: 20px; }
        h1 { text-align: center; margin-bottom: 20px; color: #4ecdc4; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin-bottom: 30px; }
        .stat-card { background: #16213e; padding: 20px; border-radius: 10px; text-align: center; }
        .stat-value { font-size: 2.5em; font-weight: bold; color: #4ecdc4; }
        .stat-label { color: #888; margin-top: 5px; }
        .tabs { display: flex; gap: 10px; margin-bottom: 20px; }
        .tab { padding: 10px 20px; background: #16213e; border: none; color: #eee; cursor: pointer; border-radius: 5px; }
        .tab.active { background: #4ecdc4; color: #1a1a2e; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        .mermaid { background: white; padding: 20px; border-radius: 10px; overflow-x: auto; }
        .sources-list { background: #16213e; border-radius: 10px; padding: 20px; }
        .source-item { padding: 15px; border-bottom: 1px solid #2a2a4a; }
        .source-item:last-child { border-bottom: none; }
        .source-name { font-weight: bold; color: #ff6b6b; font-family: monospace; }
        .source-location { color: #888; font-size: 0.9em; margin-top: 5px; }
        .source-type { display: inline-block; padding: 2px 8px; background: #4ecdc4; color: #1a1a2e; border-radius: 3px; font-size: 0.8em; margin-left: 10px; }
        .json-view { background: #16213e; padding: 20px; border-radius: 10px; }
        pre { overflow-x: auto; font-family: 'Fira Code', monospace; font-size: 0.9em; }
        .language-stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 10px; margin-top: 20px; }
        .lang-card { background: #2a2a4a; padding: 15px; border-radius: 8px; text-align: center; }
        .lang-name { text-transform: uppercase; font-weight: bold; color: #45b7d1; }
        .search-box { width: 100%%; padding: 10px; margin-bottom: 15px; background: #2a2a4a; border: none; color: #eee; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>InputTracer Flow Analysis</h1>

        <div class="stats">
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Files Analyzed</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Input Sources</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Flows Traced</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Cross-File Flows</div>
            </div>
        </div>

        <div class="tabs">
            <button class="tab active" onclick="showTab('diagram')">Flow Diagram</button>
            <button class="tab" onclick="showTab('sources')">Sources (%d)</button>
            <button class="tab" onclick="showTab('json')">JSON Data</button>
        </div>

        <div id="diagram" class="tab-content active">
            <div class="mermaid">
%s
            </div>
        </div>

        <div id="sources" class="tab-content">
            <input type="text" class="search-box" placeholder="Search sources..." onkeyup="filterSources(this.value)">
            <div class="sources-list">
                %s
            </div>
        </div>

        <div id="json" class="tab-content">
            <div class="json-view">
                <pre>%s</pre>
            </div>
        </div>

        <div class="language-stats">
            %s
        </div>
    </div>

    <script>
        mermaid.initialize({ startOnLoad: true, theme: 'neutral' });

        function showTab(tabId) {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            document.querySelector('.tab-content#' + tabId).classList.add('active');
            event.target.classList.add('active');
        }

        function filterSources(query) {
            const items = document.querySelectorAll('.source-item');
            query = query.toLowerCase();
            items.forEach(item => {
                const text = item.textContent.toLowerCase();
                item.style.display = text.includes(query) ? 'block' : 'none';
            });
        }
    </script>
</body>
</html>`,
		r.Stats.FilesParsed,
		r.Stats.SourcesFound,
		r.Stats.FlowsTraced,
		r.Stats.CrossFileFlows,
		len(r.Sources),
		mermaidDiagram,
		generateSourcesHTML(r.Sources),
		html.EscapeString(jsonData),
		generateLanguageStatsHTML(r.Stats.ByLanguage),
	)
}

func generateSourcesHTML(sources []*types.FlowNode) string {
	var sb strings.Builder
	for _, src := range sources {
		sb.WriteString(fmt.Sprintf(`
            <div class="source-item">
                <span class="source-name">%s</span>
                <span class="source-type">%s</span>
                <div class="source-location">%s:%d:%d</div>
                <div style="margin-top:5px;color:#666;font-family:monospace;font-size:0.85em;">%s</div>
            </div>`,
			html.EscapeString(src.Name),
			string(src.SourceType),
			html.EscapeString(shortPath(src.FilePath)),
			src.Line,
			src.Column,
			html.EscapeString(src.Snippet),
		))
	}
	return sb.String()
}

func generateLanguageStatsHTML(stats map[string]*LanguageStats) string {
	var sb strings.Builder
	for lang, stat := range stats {
		sb.WriteString(fmt.Sprintf(`
            <div class="lang-card">
                <div class="lang-name">%s</div>
                <div>%d files</div>
                <div>%d sources</div>
            </div>`,
			lang,
			stat.Files,
			stat.Sources,
		))
	}
	return sb.String()
}

// Helper functions

func sanitizeDotID(id string) string {
	// Replace special characters that might break DOT syntax
	replacer := strings.NewReplacer(
		":", "_",
		"/", "_",
		"\\", "_",
		".", "_",
		"-", "_",
		" ", "_",
		"[", "_",
		"]", "_",
		"'", "",
		"\"", "",
	)
	return replacer.Replace(id)
}

func sanitizeDotLabel(label string) string {
	// Escape special characters for DOT labels
	replacer := strings.NewReplacer(
		"\"", "\\\"",
		"\n", "\\n",
		"\t", "\\t",
	)
	return replacer.Replace(label)
}

func sanitizeMermaidID(id string) string {
	// Replace special characters that might break Mermaid syntax
	replacer := strings.NewReplacer(
		":", "_",
		"/", "_",
		"\\", "_",
		".", "_",
		"-", "_",
		" ", "_",
		"[", "_",
		"]", "_",
		"'", "",
		"\"", "",
		"(", "_",
		")", "_",
		"{", "_",
		"}", "_",
		"<", "_",
		">", "_",
	)
	result := replacer.Replace(id)
	// Ensure it starts with a letter
	if len(result) > 0 && (result[0] >= '0' && result[0] <= '9') {
		result = "n" + result
	}
	return result
}

func sanitizeMermaidLabel(label string) string {
	// Escape special characters for Mermaid labels
	replacer := strings.NewReplacer(
		"\"", "'",
		"\n", " ",
		"\t", " ",
		"|", "\\|",
		"[", "(",
		"]", ")",
		"{", "(",
		"}", ")",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(label)
}

func shortPath(path string) string {
	// Get the last 2 components of the path
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return ".../" + strings.Join(parts[len(parts)-2:], "/")
	}
	return path
}
