package output

import (
	"encoding/json"
	"io"
	"os"

	"github.com/hatlesswizard/inputtracer/pkg/tracer"
)

// JSONExporter exports trace results to JSON format
type JSONExporter struct {
	PrettyPrint bool
	Indent      string
}

// NewJSONExporter creates a new JSON exporter
func NewJSONExporter(prettyPrint bool) *JSONExporter {
	return &JSONExporter{
		PrettyPrint: prettyPrint,
		Indent:      "  ",
	}
}

// Export exports the trace result to JSON string
func (e *JSONExporter) Export(result *tracer.TraceResult) (string, error) {
	var data []byte
	var err error

	if e.PrettyPrint {
		data, err = json.MarshalIndent(result, "", e.Indent)
	} else {
		data, err = json.Marshal(result)
	}

	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ExportToWriter exports the trace result to an io.Writer
func (e *JSONExporter) ExportToWriter(result *tracer.TraceResult, w io.Writer) error {
	encoder := json.NewEncoder(w)
	if e.PrettyPrint {
		encoder.SetIndent("", e.Indent)
	}
	return encoder.Encode(result)
}

// ExportToFile exports the trace result to a file
func (e *JSONExporter) ExportToFile(result *tracer.TraceResult, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return e.ExportToWriter(result, f)
}

// SummaryReport generates a summary report of the trace
type SummaryReport struct {
	TotalFiles           int                       `json:"total_files"`
	TotalSources         int                       `json:"total_sources"`
	TotalTaintedVars     int                       `json:"total_tainted_variables"`
	TotalTaintedFuncs    int                       `json:"total_tainted_functions"`
	SourcesByType        map[string]int            `json:"sources_by_type"`
	SourcesByLanguage    map[string]int            `json:"sources_by_language"`
	TaintedByLanguage    map[string]int            `json:"tainted_by_language"`
	MostTaintedFiles     []FileStatistic           `json:"most_tainted_files"`
	PropagationDepthDist map[int]int               `json:"propagation_depth_distribution"`
}

// FileStatistic represents statistics for a single file
type FileStatistic struct {
	FilePath       string `json:"file_path"`
	SourceCount    int    `json:"source_count"`
	TaintedVars    int    `json:"tainted_variables"`
	TaintedFuncs   int    `json:"tainted_functions"`
}

// GenerateSummary generates a summary report from trace results
func GenerateSummary(result *tracer.TraceResult) *SummaryReport {
	report := &SummaryReport{
		TotalFiles:           result.Stats.FilesAnalyzed,
		TotalSources:         len(result.Sources),
		TotalTaintedVars:     len(result.TaintedVariables),
		TotalTaintedFuncs:    len(result.TaintedFunctions),
		SourcesByType:        make(map[string]int),
		SourcesByLanguage:    make(map[string]int),
		TaintedByLanguage:    make(map[string]int),
		PropagationDepthDist: make(map[int]int),
	}

	// Count sources by type
	for _, source := range result.Sources {
		report.SourcesByType[source.Type]++
		report.SourcesByLanguage[source.Language]++
	}

	// Count tainted vars by language and depth
	fileStats := make(map[string]*FileStatistic)
	for _, tv := range result.TaintedVariables {
		report.TaintedByLanguage[tv.Language]++
		report.PropagationDepthDist[tv.Depth]++

		// Track per-file stats
		fp := tv.Location.FilePath
		if _, exists := fileStats[fp]; !exists {
			fileStats[fp] = &FileStatistic{FilePath: fp}
		}
		fileStats[fp].TaintedVars++
	}

	// Count sources per file
	for _, source := range result.Sources {
		fp := source.Location.FilePath
		if _, exists := fileStats[fp]; !exists {
			fileStats[fp] = &FileStatistic{FilePath: fp}
		}
		fileStats[fp].SourceCount++
	}

	// Count tainted funcs per file
	for _, tf := range result.TaintedFunctions {
		fp := tf.FilePath
		if _, exists := fileStats[fp]; !exists {
			fileStats[fp] = &FileStatistic{FilePath: fp}
		}
		fileStats[fp].TaintedFuncs++
	}

	// Get top 10 most tainted files
	report.MostTaintedFiles = getTopTaintedFiles(fileStats, 10)

	return report
}

// getTopTaintedFiles returns the top N files by taint count
func getTopTaintedFiles(stats map[string]*FileStatistic, n int) []FileStatistic {
	var files []FileStatistic
	for _, stat := range stats {
		files = append(files, *stat)
	}

	// Sort by total taint score (sources + vars + funcs)
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			scoreI := files[i].SourceCount + files[i].TaintedVars + files[i].TaintedFuncs
			scoreJ := files[j].SourceCount + files[j].TaintedVars + files[j].TaintedFuncs
			if scoreJ > scoreI {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	if len(files) > n {
		files = files[:n]
	}
	return files
}

// ExportSummary exports just the summary report
func (e *JSONExporter) ExportSummary(result *tracer.TraceResult) (string, error) {
	summary := GenerateSummary(result)

	var data []byte
	var err error

	if e.PrettyPrint {
		data, err = json.MarshalIndent(summary, "", e.Indent)
	} else {
		data, err = json.Marshal(summary)
	}

	if err != nil {
		return "", err
	}
	return string(data), nil
}
