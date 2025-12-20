// Package discovery - carrier map builder and serialization
package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CarrierMap stores discovered input carriers for a codebase
type CarrierMap struct {
	CodebasePath string       `json:"codebase_path"`
	DiscoveredAt time.Time    `json:"discovered_at"`
	PHPVersion   string       `json:"php_version,omitempty"`
	Framework    string       `json:"framework,omitempty"`
	Carriers     []InputCarrier `json:"carriers"`
	Statistics   CarrierStats `json:"statistics"`
}

// CarrierStats holds statistics about the discovery
type CarrierStats struct {
	TotalSuperglobalUsages int            `json:"total_superglobal_usages"`
	UniqueCarriers         int            `json:"unique_carriers"`
	TotalTaintFlows        int            `json:"total_taint_flows"`
	ClassesAnalyzed        int            `json:"classes_analyzed"`
	BySourceType           map[string]int `json:"by_source_type"`
	ByClassName            map[string]int `json:"by_class_name"`
}

// BuildCarrierMap analyzes a codebase and builds a carrier map
func BuildCarrierMap(codebasePath string) (*CarrierMap, error) {
	// Validate path exists
	absPath, err := filepath.Abs(codebasePath)
	if err != nil {
		return nil, fmt.Errorf("invalid codebase path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access codebase: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("codebase path is not a directory: %s", absPath)
	}

	// Run taint propagation analysis
	propagator := NewTaintPropagator()
	if err := propagator.AnalyzeCodebase(absPath); err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	// Build the carrier map
	carrierMap := &CarrierMap{
		CodebasePath: absPath,
		DiscoveredAt: time.Now(),
		Carriers:     propagator.GetCarriers(),
		Statistics:   calculateStats(propagator),
	}

	// Try to detect framework
	carrierMap.Framework = detectFramework(absPath)

	return carrierMap, nil
}

// calculateStats computes statistics from the propagator results
func calculateStats(propagator *TaintPropagator) CarrierStats {
	stats := CarrierStats{
		TotalTaintFlows: len(propagator.GetFlows()),
		UniqueCarriers:  len(propagator.GetCarriers()),
		ClassesAnalyzed: len(propagator.GetClassInfo()),
		BySourceType:    make(map[string]int),
		ByClassName:     make(map[string]int),
	}

	// Count by source type and class
	for _, carrier := range propagator.GetCarriers() {
		for _, source := range carrier.SourceTypes {
			stats.BySourceType[source]++
		}
		stats.ByClassName[carrier.ClassName]++
	}

	return stats
}

// detectFramework attempts to detect the PHP framework being used
func detectFramework(codebasePath string) string {
	// Check for common framework indicators
	frameworkIndicators := map[string][]string{
		"mybb":       {"inc/class_core.php", "inc/init.php"},
		"wordpress":  {"wp-config.php", "wp-includes/version.php"},
		"laravel":    {"artisan", "bootstrap/app.php"},
		"symfony":    {"symfony.lock", "config/bundles.php"},
		"codeigniter": {"system/core/CodeIgniter.php"},
		"drupal":     {"core/includes/bootstrap.inc"},
		"yii2":       {"yii"},
		"cakephp":    {"config/app.php", "src/Application.php"},
		"phpbb":      {"phpbb/request/request.php"},
		"prestashop": {"config/defines.inc.php"},
	}

	for framework, indicators := range frameworkIndicators {
		for _, indicator := range indicators {
			checkPath := filepath.Join(codebasePath, indicator)
			if _, err := os.Stat(checkPath); err == nil {
				return framework
			}
		}
	}

	return "unknown"
}

// SaveToFile saves the carrier map to a JSON file
func (m *CarrierMap) SaveToFile(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal carrier map: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write carrier map: %w", err)
	}

	return nil
}

// LoadCarrierMap loads a carrier map from a JSON file
func LoadCarrierMap(path string) (*CarrierMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read carrier map: %w", err)
	}

	var carrierMap CarrierMap
	if err := json.Unmarshal(data, &carrierMap); err != nil {
		return nil, fmt.Errorf("failed to parse carrier map: %w", err)
	}

	return &carrierMap, nil
}

// FindCarrier looks up a carrier by class and property/method name
func (m *CarrierMap) FindCarrier(className, name string) *InputCarrier {
	for i := range m.Carriers {
		c := &m.Carriers[i]
		if c.ClassName == className {
			if c.PropertyName == name || c.MethodName == name {
				return c
			}
		}
	}
	return nil
}

// FindCarriersByClass returns all carriers for a given class
func (m *CarrierMap) FindCarriersByClass(className string) []InputCarrier {
	var carriers []InputCarrier
	for _, c := range m.Carriers {
		if c.ClassName == className {
			carriers = append(carriers, c)
		}
	}
	return carriers
}

// FindCarriersBySourceType returns all carriers that have a specific source type
func (m *CarrierMap) FindCarriersBySourceType(sourceType string) []InputCarrier {
	var carriers []InputCarrier
	for _, c := range m.Carriers {
		for _, st := range c.SourceTypes {
			if st == sourceType {
				carriers = append(carriers, c)
				break
			}
		}
	}
	return carriers
}

// GetAllClassNames returns unique class names that have carriers
func (m *CarrierMap) GetAllClassNames() []string {
	classSet := make(map[string]bool)
	for _, c := range m.Carriers {
		classSet[c.ClassName] = true
	}

	var classes []string
	for class := range classSet {
		classes = append(classes, class)
	}
	return classes
}

// GetAllSourceTypes returns unique source types found
func (m *CarrierMap) GetAllSourceTypes() []string {
	sourceSet := make(map[string]bool)
	for _, c := range m.Carriers {
		for _, st := range c.SourceTypes {
			sourceSet[st] = true
		}
	}

	var sources []string
	for source := range sourceSet {
		sources = append(sources, source)
	}
	return sources
}

// Summary returns a human-readable summary of the carrier map
func (m *CarrierMap) Summary() string {
	summary := fmt.Sprintf("Carrier Map Summary\n")
	summary += fmt.Sprintf("==================\n")
	summary += fmt.Sprintf("Codebase: %s\n", m.CodebasePath)
	summary += fmt.Sprintf("Framework: %s\n", m.Framework)
	summary += fmt.Sprintf("Discovered: %s\n", m.DiscoveredAt.Format(time.RFC3339))
	summary += fmt.Sprintf("\nStatistics:\n")
	summary += fmt.Sprintf("  - Classes Analyzed: %d\n", m.Statistics.ClassesAnalyzed)
	summary += fmt.Sprintf("  - Taint Flows Found: %d\n", m.Statistics.TotalTaintFlows)
	summary += fmt.Sprintf("  - Unique Carriers: %d\n", m.Statistics.UniqueCarriers)

	summary += fmt.Sprintf("\nCarriers by Source Type:\n")
	for source, count := range m.Statistics.BySourceType {
		summary += fmt.Sprintf("  - %s: %d carriers\n", source, count)
	}

	summary += fmt.Sprintf("\nCarriers by Class:\n")
	for class, count := range m.Statistics.ByClassName {
		summary += fmt.Sprintf("  - %s: %d carriers\n", class, count)
	}

	summary += fmt.Sprintf("\nDetailed Carriers:\n")
	for _, c := range m.Carriers {
		if c.PropertyName != "" {
			summary += fmt.Sprintf("  - %s->%s [%s] <- %v\n",
				c.ClassName, c.PropertyName, c.AccessPattern, c.SourceTypes)
		} else if c.MethodName != "" {
			summary += fmt.Sprintf("  - %s->%s() <- %v\n",
				c.ClassName, c.MethodName, c.SourceTypes)
		}
	}

	return summary
}

// Merge merges another carrier map into this one
func (m *CarrierMap) Merge(other *CarrierMap) {
	// Add carriers that don't already exist
	existingKeys := make(map[string]bool)
	for _, c := range m.Carriers {
		key := c.ClassName + "." + c.PropertyName + "." + c.MethodName
		existingKeys[key] = true
	}

	for _, c := range other.Carriers {
		key := c.ClassName + "." + c.PropertyName + "." + c.MethodName
		if !existingKeys[key] {
			m.Carriers = append(m.Carriers, c)
		}
	}

	// Update statistics
	m.Statistics.UniqueCarriers = len(m.Carriers)
	m.Statistics.BySourceType = make(map[string]int)
	m.Statistics.ByClassName = make(map[string]int)

	for _, c := range m.Carriers {
		for _, source := range c.SourceTypes {
			m.Statistics.BySourceType[source]++
		}
		m.Statistics.ByClassName[c.ClassName]++
	}
}
