// Package base provides shared helpers for language analyzers.
//
// These helpers eliminate copy-paste boilerplate that was duplicated across
// all 11 language analyzers (php, javascript, typescript, python, golang,
// java, ruby, c, cpp, csharp, rust).
package base

import (
	"regexp"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// RegisterFrameworkPatterns loads a slice of common.FrameworkPattern into a
// BaseAnalyzer. This is the body that was copy-pasted verbatim across six
// language analyzers (php, javascript, typescript, python, golang, java) —
// the only difference between them was which GetAllPatterns() call was made
// and (for typescript) an optional language override.
//
// Call it from your language analyzer's init/constructor like:
//
//	base.RegisterFrameworkPatterns(a.BaseAnalyzer, jsPatterns.GetAllPatterns(), "")
//	base.RegisterFrameworkPatterns(a.BaseAnalyzer, jsPatterns.GetAllPatterns(), "typescript") // override
func RegisterFrameworkPatterns(ba *analyzer.BaseAnalyzer, patterns []*common.FrameworkPattern, languageOverride string) {
	for _, p := range patterns {
		lang := p.Language
		if languageOverride != "" {
			lang = languageOverride
		}
		fp := &types.FrameworkPattern{
			ID:              p.ID,
			Framework:       p.Framework,
			Language:        lang,
			Name:            p.Name,
			Description:     p.Description,
			ClassPattern:    p.ClassPattern,
			MethodPattern:   p.MethodPattern,
			PropertyPattern: p.PropertyPattern,
			AccessPattern:   p.AccessPattern,
			SourceType:      types.SourceType(p.SourceType),
			SourceKey:       p.SourceKey,
			CarrierClass:    p.CarrierClass,
			CarrierProperty: p.CarrierProperty,
			PopulatedBy:     p.PopulatedBy,
			PopulatedFrom:   p.PopulatedFrom,
		}
		ba.AddFrameworkPattern(fp)
	}
}

// PropertyAssignPatternFunc is a function that, given a parameter name, returns
// a compiled *regexp.Regexp that matches assignments of that parameter to an
// object property (e.g. $this->prop = $param in PHP, self.prop = param in Python,
// this.prop = param in JavaScript/TypeScript/Go).
//
// If the language has no such property-assignment convention, pass nil.
type PropertyAssignPatternFunc func(paramName string) *regexp.Regexp

// AnalyzeMethodBody is a shared implementation of the LanguageAnalyzer
// AnalyzeMethodBody method. It covers the logic that is structurally identical
// across all language analyzers:
//
//   - Initialises MethodFlowAnalysis with empty slices/maps.
//   - Skips analysis when BodySource is empty.
//   - Checks whether each parameter flows to a return statement.
//   - Optionally checks whether each parameter flows to an object property via
//     the language-specific propertyPattern callback.
//   - Checks whether the method returns a known input source directly.
//
// Parameters:
//   - method: the method definition to analyse.
//   - source: raw source bytes (unused here, kept for interface alignment).
//   - state: analysis state (unused here, kept for interface alignment).
//   - paramPrefix: the sigil prepended to parameters in the body (e.g. "$" for
//     PHP, "" for Python/Go/JS/Ruby).
//   - inputSources: the language's input-source map. If the body contains
//     "return" and any key from this map, ReturnsInput is set to true.
//   - propertyPattern: optional; when non-nil it is called with each
//     "paramPrefix+param.Name" value and the returned regexp is applied to the
//     body to detect property assignments.
func AnalyzeMethodBody(
	method *types.MethodDef,
	state *types.AnalysisState,
	paramPrefix string,
	inputSources map[string]types.SourceType,
	propertyPattern PropertyAssignPatternFunc,
) *analyzer.MethodFlowAnalysis {
	analysis := &analyzer.MethodFlowAnalysis{
		ParamsToReturn:     make([]int, 0),
		ParamsToProperties: make(map[int][]string),
		ParamsToCallArgs:   make(map[int][]*types.CallSite),
		TaintedVariables:   make(map[string]*types.TaintInfo),
		Assignments:        make([]*types.Assignment, 0),
		Calls:              make([]*types.CallSite, 0),
		Returns:            make([]analyzer.ReturnInfo, 0),
	}

	if method.BodySource == "" {
		return analysis
	}

	body := method.BodySource

	for i, param := range method.Parameters {
		paramRef := paramPrefix + param.Name

		// Check if param flows to return
		if strings.Contains(body, "return") && strings.Contains(body, paramRef) {
			analysis.ParamsToReturn = append(analysis.ParamsToReturn, i)
		}

		// Check if param flows to a property (e.g. $this->prop = $param)
		if propertyPattern != nil {
			re := propertyPattern(paramRef)
			if re != nil {
				matches := re.FindAllStringSubmatch(body, -1)
				for _, match := range matches {
					if len(match) > 1 {
						analysis.ParamsToProperties[i] = append(analysis.ParamsToProperties[i], match[1])
						analysis.ModifiesProperties = true
					}
				}
			}
		}
	}

	// Check if method returns input directly
	for src := range inputSources {
		if strings.Contains(body, "return") && strings.Contains(body, src) {
			analysis.ReturnsInput = true
			break
		}
	}

	return analysis
}
