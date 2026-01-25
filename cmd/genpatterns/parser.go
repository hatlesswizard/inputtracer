// Package main - parser.go extracts methods and properties from PHP source
package main

import (
	"regexp"
	"strings"
)

// ParsedMethod represents an extracted PHP method
type ParsedMethod struct {
	Name       string
	ClassName  string
	IsProperty bool
}

// Parser extracts methods from PHP source code
type Parser struct {
	methodRegex   *regexp.Regexp
	propertyRegex *regexp.Regexp
}

// NewParser creates a new PHP parser
func NewParser() *Parser {
	return &Parser{
		// Match: public function methodName(
		methodRegex: regexp.MustCompile(`public\s+function\s+(\w+)\s*\(`),
		// Match: public TypeHint $propertyName or public $propertyName
		propertyRegex: regexp.MustCompile(`public\s+(?:\??\w+\s+)?\$(\w+)`),
	}
}

// ParseMethods extracts public methods from PHP source
func (p *Parser) ParseMethods(source string, className string) []ParsedMethod {
	var methods []ParsedMethod
	seen := make(map[string]bool)

	matches := p.methodRegex.FindAllStringSubmatch(source, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			name := match[1]
			// Skip magic methods and internal methods
			if strings.HasPrefix(name, "__") {
				continue
			}
			if seen[name] {
				continue
			}
			seen[name] = true
			methods = append(methods, ParsedMethod{
				Name:       name,
				ClassName:  className,
				IsProperty: false,
			})
		}
	}

	return methods
}

// ParseProperties extracts public properties from PHP source
func (p *Parser) ParseProperties(source string, className string) []ParsedMethod {
	var props []ParsedMethod
	seen := make(map[string]bool)

	matches := p.propertyRegex.FindAllStringSubmatch(source, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			name := match[1]
			if seen[name] {
				continue
			}
			seen[name] = true
			props = append(props, ParsedMethod{
				Name:       name,
				ClassName:  className,
				IsProperty: true,
			})
		}
	}

	return props
}
