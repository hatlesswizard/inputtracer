// Package main - genpatterns fetches framework sources and generates Go patterns
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	outputDir := flag.String("o", ".", "Output directory for generated files")
	framework := flag.String("framework", "", "Generate for specific framework (laravel, symfony). Empty = all")
	flag.Parse()

	fetcher := NewFetcher(30 * time.Second)
	parser := NewParser()
	generator := NewGenerator()

	frameworks := []string{"laravel", "symfony", "wordpress"}
	if *framework != "" {
		frameworks = []string{*framework}
	}

	var errs []string
	for _, fwName := range frameworks {
		fw, ok := Frameworks[fwName]
		if !ok {
			errs = append(errs, fmt.Sprintf("unknown framework: %s", fwName))
			continue
		}

		fmt.Printf("Fetching %s sources...\n", fwName)
		sources, err := fetcher.FetchFrameworkSources(fw)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: fetch error: %v", fwName, err))
			continue
		}

		content, err := generateFramework(fwName, parser, generator, sources, fw)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: generate error: %v", fwName, err))
			continue
		}

		outputPath := filepath.Join(*outputDir, fwName+".go")
		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			errs = append(errs, fmt.Sprintf("%s: write error: %v", fwName, err))
			continue
		}
		fmt.Printf("Generated %s\n", outputPath)
	}

	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "\nErrors encountered:\n")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		os.Exit(1)
	}

	fmt.Println("Done!")
}

// generateFramework dispatches to the appropriate framework-specific generation
// logic and returns the generated Go source content.
func generateFramework(fwName string, parser *Parser, generator *Generator, sources map[string]string, fw *FrameworkDefinition) (string, error) {
	switch fwName {
	case "laravel":
		return generateLaravel(parser, generator, sources, fw), nil
	case "symfony":
		return generateSymfony(parser, generator, sources, fw), nil
	case "wordpress":
		return generateWordPress(parser, generator, sources, fw), nil
	default:
		return "", fmt.Errorf("unknown framework: %s", fwName)
	}
}

func generateLaravel(parser *Parser, generator *Generator, sources map[string]string, fw *FrameworkDefinition) string {
	var allMethods []ParsedMethod
	for className, src := range sources {
		allMethods = append(allMethods, parser.ParseMethods(src, className)...)
	}
	return generator.GenerateLaravel(filterMethods(allMethods, nil), fw)
}

func generateSymfony(parser *Parser, generator *Generator, sources map[string]string, fw *FrameworkDefinition) string {
	var allMethods []ParsedMethod
	var allProperties []ParsedMethod

	// Parse ParameterBag and InputBag methods
	if src, ok := sources["ParameterBag"]; ok {
		allMethods = append(allMethods, parser.ParseMethods(src, "ParameterBag")...)
	}
	if src, ok := sources["InputBag"]; ok {
		allMethods = append(allMethods, parser.ParseMethods(src, "InputBag")...)
	}

	// Parse Request public properties (properties still need explicit mapping)
	if src, ok := sources["Request"]; ok {
		for _, p := range parser.ParseProperties(src, "Request") {
			if _, ok := symfonyPropertyMappings()[p.Name]; ok {
				allProperties = append(allProperties, p)
			}
		}
	}

	return generator.GenerateSymfony(filterMethods(allMethods, nil), allProperties, fw)
}

func generateWordPress(parser *Parser, generator *Generator, sources map[string]string, fw *FrameworkDefinition) string {
	var allMethods []ParsedMethod

	// Parse WP_REST_Request methods
	if src, ok := sources["WP_REST_Request"]; ok {
		allMethods = append(allMethods, parser.ParseMethods(src, "WP_REST_Request")...)
	}

	return generator.GenerateWordPress(filterMethods(allMethods, wordPressExcludedMethods()), fw)
}

// filterMethods removes methods that are in the shared exclusion list or in
// the optional extraExcluded map. Pass nil for extraExcluded when no extra
// exclusions are needed.
func filterMethods(methods []ParsedMethod, extraExcluded map[string]bool) []ParsedMethod {
	var filtered []ParsedMethod
	for _, m := range methods {
		if IsExcluded(m.Name) {
			continue
		}
		if extraExcluded[m.Name] {
			continue
		}
		filtered = append(filtered, m)
	}
	return filtered
}
