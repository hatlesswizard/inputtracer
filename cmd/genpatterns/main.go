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

	for _, fwName := range frameworks {
		fw, ok := Frameworks[fwName]
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown framework: %s\n", fwName)
			os.Exit(1)
		}

		fmt.Printf("Fetching %s sources...\n", fwName)
		sources, err := fetcher.FetchFrameworkSources(fw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetch error: %v\n", err)
			os.Exit(1)
		}

		var content string
		switch fwName {
		case "laravel":
			content = generateLaravel(parser, generator, sources, fw)
		case "symfony":
			content = generateSymfony(parser, generator, sources, fw)
		case "wordpress":
			content = generateWordPress(parser, generator, sources, fw)
		}

		outputPath := filepath.Join(*outputDir, fwName+".go")
		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated %s\n", outputPath)
	}

	fmt.Println("Done!")
}

func generateLaravel(parser *Parser, generator *Generator, sources map[string]string, fw *FrameworkDefinition) string {
	var allMethods []ParsedMethod
	for className, src := range sources {
		allMethods = append(allMethods, parser.ParseMethods(src, className)...)
	}
	return generator.GenerateLaravel(filterExcluded(allMethods), fw)
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
			if _, ok := SymfonyPropertyMappings[p.Name]; ok {
				allProperties = append(allProperties, p)
			}
		}
	}

	return generator.GenerateSymfony(filterExcluded(allMethods), allProperties, fw)
}

func generateWordPress(parser *Parser, generator *Generator, sources map[string]string, fw *FrameworkDefinition) string {
	var allMethods []ParsedMethod

	// Parse WP_REST_Request methods
	if src, ok := sources["WP_REST_Request"]; ok {
		allMethods = append(allMethods, parser.ParseMethods(src, "WP_REST_Request")...)
	}

	return generator.GenerateWordPress(filterWordPressExcluded(allMethods), fw)
}

// filterExcluded removes methods that are in the exclusion list
func filterExcluded(methods []ParsedMethod) []ParsedMethod {
	var filtered []ParsedMethod
	for _, m := range methods {
		if !IsExcluded(m.Name) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// filterWordPressExcluded removes WordPress-specific excluded methods
func filterWordPressExcluded(methods []ParsedMethod) []ParsedMethod {
	var filtered []ParsedMethod
	for _, m := range methods {
		if !IsExcluded(m.Name) && !WordPressExcludedMethods[m.Name] {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
