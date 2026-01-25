# Centralize Semantic Patterns Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Move scattered regex patterns from `pkg/semantic/` to language-specific files in `pkg/sources/`.

**Architecture:** Create pattern files for each language (`patterns.go`) that export pre-compiled regexes and builder functions. Semantic analyzers import and use these centralized patterns instead of inline compilation.

**Tech Stack:** Go, regexp package

---

## Summary of Patterns to Move

| Source File | Pattern | Destination |
|-------------|---------|-------------|
| `semantic/analyzer/php/analyzer.go` | `\$this->(\w+)\s*=` | `sources/php/patterns.go` |
| `semantic/analyzer/php/analyzer.go` | `\$this->(\w+)\[.*\]\s*=` | `sources/php/patterns.go` |
| `semantic/analyzer/php/analyzer.go` | `\[['"](\w+)['"]\]` | `sources/php/patterns.go` |
| `semantic/analyzer/php/analyzer.go` | `\[(\$\w+)\]` | `sources/php/patterns.go` |
| `semantic/analyzer/php/analyzer.go` | `\$\w+->(prop)(\[\|$)` | `sources/php/patterns.go` |
| `semantic/analyzer/php/analyzer.go` | `->(method)\(` | `sources/php/patterns.go` |
| `semantic/discovery/taint.go` | `return\s+\$this->` | `sources/php/patterns.go` |
| `semantic/analyzer/javascript/analyzer.go` | `\.get\(['"](\w+)['"]\)` | `sources/javascript/patterns.go` |
| `semantic/analyzer/javascript/analyzer.go` | `^\[['"](\w+)['"]\]` | `sources/javascript/patterns.go` |
| `semantic/analyzer/javascript/analyzer.go` | `^\.(\w+)` | `sources/javascript/patterns.go` |
| `semantic/analyzer/javascript/analyzer.go` | `this\.(\w+)\s*=` | `sources/javascript/patterns.go` |
| `semantic/analyzer/python/analyzer.go` | `self\.(\w+)\s*=` | `sources/python/patterns.go` |
| `semantic/analyzer/python/analyzer.go` | `\[['"](\w+)['"]\]` | `sources/python/patterns.go` |
| `semantic/analyzer/python/analyzer.go` | `\.get\(['"](\w+)['"]\)` | `sources/python/patterns.go` |
| `semantic/analyzer/typescript/analyzer.go` | `this\.(\w+)\s*=` | `sources/typescript/patterns.go` |
| `semantic/analyzer/typescript/analyzer.go` | `\[['"\x60](\w+)['"\x60]\]` | `sources/typescript/patterns.go` |
| `semantic/analyzer/typescript/analyzer.go` | `\.(body\|query\|params\|headers\|cookies)\.(\w+)` | `sources/typescript/patterns.go` |
| `semantic/analyzer/typescript/analyzer.go` | `@(method)\(` | `sources/typescript/patterns.go` |

---

### Task 1: Add PHP Semantic Patterns

**Files:**
- Modify: `pkg/sources/php/patterns.go`

**Step 1: Add the new patterns to pkg/sources/php/patterns.go**

Add after the existing SYMBOLIC EXECUTION PATTERNS section:

```go
// =============================================================================
// SEMANTIC ANALYSIS PATTERNS
// Used by semantic analyzers for flow analysis
// =============================================================================

var (
	// ThisPropertyAssignPattern matches $this->property = ...
	// Used to detect constructor/method parameter flow to properties
	ThisPropertyAssignPattern = regexp.MustCompile(`\$this->(\w+)\s*=`)

	// ThisArrayPropertyAssignPattern matches $this->property[...] = ...
	// Used to detect array property assignments
	ThisArrayPropertyAssignPattern = regexp.MustCompile(`\$this->(\w+)\[.*\]\s*=`)

	// ArrayKeyAccessPattern matches ['key'] or ["key"]
	// Used to extract array keys from expressions
	ArrayKeyAccessPattern = regexp.MustCompile(`\[['"](\w+)['"]\]`)

	// VariableKeyAccessPattern matches [$variable]
	// Used to extract variable-based array access
	VariableKeyAccessPattern = regexp.MustCompile(`\[(\$\w+)\]`)

	// ReturnThisPropertyPrefix is the static prefix for return $this->property patterns
	// Use BuildReturnPropertyPattern for dynamic patterns with specific property names
	ReturnThisPropertyPrefix = `return\s+\$this->`

	// MethodCallSuffix is the suffix pattern for method calls
	MethodCallSuffix = `\(`
)

// BuildThisPropertyAssignPattern creates a pattern for $this->property = ... paramName
func BuildThisPropertyAssignPattern(paramName string) *regexp.Regexp {
	return regexp.MustCompile(`\$this->(\w+)\s*=.*` + regexp.QuoteMeta(paramName))
}

// BuildThisArrayPropertyAssignPattern creates a pattern for $this->property[...] = ... paramName
func BuildThisArrayPropertyAssignPattern(paramName string) *regexp.Regexp {
	return regexp.MustCompile(`\$this->(\w+)\[.*\]\s*=.*` + regexp.QuoteMeta(paramName))
}

// BuildReturnPropertyPattern creates a pattern for return $this->propertyName
func BuildReturnPropertyPattern(propertyName string) *regexp.Regexp {
	return regexp.MustCompile(ReturnThisPropertyPrefix + regexp.QuoteMeta(propertyName))
}

// BuildReturnPropertyArrayPattern creates a pattern for return $this->propertyName[
func BuildReturnPropertyArrayPattern(propertyName string) *regexp.Regexp {
	return regexp.MustCompile(ReturnThisPropertyPrefix + regexp.QuoteMeta(propertyName) + `\[`)
}

// BuildPropertyAccessPattern creates a pattern for $var->property or $var->property[
func BuildPropertyAccessPattern(propertyName string) *regexp.Regexp {
	return regexp.MustCompile(`\$\w+->(` + regexp.QuoteMeta(propertyName) + `)(\[|$)`)
}

// BuildMethodCallPattern creates a pattern for ->methodName(
func BuildMethodCallPattern(methodName string) *regexp.Regexp {
	return regexp.MustCompile(`->` + methodName + MethodCallSuffix)
}
```

**Step 2: Build and verify compilation**

Run: `go build ./pkg/sources/php/...`
Expected: Build succeeds with no errors

**Step 3: Commit**

```bash
git add pkg/sources/php/patterns.go
git commit -m "feat(sources/php): add semantic analysis patterns"
```

---

### Task 2: Create JavaScript Patterns File

**Files:**
- Create: `pkg/sources/javascript/patterns.go`

**Step 1: Create the new patterns file**

```go
// Package javascript provides centralized JavaScript patterns for semantic analysis
package javascript

import "regexp"

// =============================================================================
// SEMANTIC ANALYSIS PATTERNS
// Used by semantic analyzers for flow analysis
// =============================================================================

var (
	// MapGetPattern matches .get('key') or .get("key")
	// Used to extract keys from Map/object .get() calls
	MapGetPattern = regexp.MustCompile(`\.get\(['"](\w+)['"]\)`)

	// BracketPropertyPattern matches ['key'] or ["key"] at start of string
	// Used to extract property names from bracket notation
	BracketPropertyPattern = regexp.MustCompile(`^\[['"](\w+)['"]\]`)

	// DotPropertyPattern matches .property at start of string
	// Used to extract property names from dot notation
	DotPropertyPattern = regexp.MustCompile(`^\.(\w+)`)

	// ThisPropertyAssignPattern matches this.property = ...
	// Used to detect constructor parameter flow to properties
	ThisPropertyAssignPattern = regexp.MustCompile(`this\.(\w+)\s*=`)
)

// BuildThisPropertyAssignPattern creates a pattern for this.property = ... paramName
func BuildThisPropertyAssignPattern(paramName string) *regexp.Regexp {
	return regexp.MustCompile(`this\.(\w+)\s*=.*` + regexp.QuoteMeta(paramName))
}

// ExtractMapKey extracts the key from a .get('key') expression
func ExtractMapKey(expr string) string {
	matches := MapGetPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ExtractBracketKey extracts the key from bracket notation ['key']
func ExtractBracketKey(expr string) string {
	matches := BracketPropertyPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ExtractDotProperty extracts property name from .property notation
func ExtractDotProperty(expr string) string {
	matches := DotPropertyPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
```

**Step 2: Build and verify compilation**

Run: `go build ./pkg/sources/javascript/...`
Expected: Build succeeds with no errors

**Step 3: Commit**

```bash
git add pkg/sources/javascript/patterns.go
git commit -m "feat(sources/javascript): add semantic analysis patterns"
```

---

### Task 3: Create Python Patterns File

**Files:**
- Create: `pkg/sources/python/patterns.go`

**Step 1: Create the new patterns file**

```go
// Package python provides centralized Python patterns for semantic analysis
package python

import "regexp"

// =============================================================================
// SEMANTIC ANALYSIS PATTERNS
// Used by semantic analyzers for flow analysis
// =============================================================================

var (
	// SelfPropertyAssignPattern matches self.property = ...
	// Used to detect __init__ parameter flow to properties
	SelfPropertyAssignPattern = regexp.MustCompile(`self\.(\w+)\s*=`)

	// DictKeyAccessPattern matches ['key'] or ["key"]
	// Used to extract dictionary keys from expressions
	DictKeyAccessPattern = regexp.MustCompile(`\[['"](\w+)['"]\]`)

	// DictGetPattern matches .get('key') or .get("key")
	// Used to extract keys from dict.get() calls
	DictGetPattern = regexp.MustCompile(`\.get\(['"](\w+)['"]\)`)
)

// BuildSelfPropertyAssignPattern creates a pattern for self.property = ... paramName
func BuildSelfPropertyAssignPattern(paramName string) *regexp.Regexp {
	return regexp.MustCompile(`self\.(\w+)\s*=.*\b` + regexp.QuoteMeta(paramName) + `\b`)
}

// BuildPropertyPattern creates a pattern to match a property pattern with word boundary
func BuildPropertyPattern(pattern string) *regexp.Regexp {
	return regexp.MustCompile(`\b` + pattern)
}

// BuildMethodCallPattern creates a pattern for .methodName(
func BuildMethodCallPattern(methodPattern string) *regexp.Regexp {
	return regexp.MustCompile(`\.` + methodPattern + `\(`)
}

// ExtractDictKey extracts the key from dict['key'] or dict["key"] expression
func ExtractDictKey(expr string) string {
	matches := DictKeyAccessPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ExtractDictGetKey extracts the key from dict.get('key') expression
func ExtractDictGetKey(expr string) string {
	matches := DictGetPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
```

**Step 2: Build and verify compilation**

Run: `go build ./pkg/sources/python/...`
Expected: Build succeeds with no errors

**Step 3: Commit**

```bash
git add pkg/sources/python/patterns.go
git commit -m "feat(sources/python): add semantic analysis patterns"
```

---

### Task 4: Create TypeScript Patterns File

**Files:**
- Create: `pkg/sources/typescript/patterns.go`

**Step 1: Create the directory and patterns file**

```go
// Package typescript provides centralized TypeScript patterns for semantic analysis
package typescript

import "regexp"

// =============================================================================
// SEMANTIC ANALYSIS PATTERNS
// Used by semantic analyzers for flow analysis
// =============================================================================

var (
	// ThisPropertyAssignPattern matches this.property = ...
	// Used to detect constructor parameter flow to properties
	ThisPropertyAssignPattern = regexp.MustCompile(`this\.(\w+)\s*=`)

	// BracketKeyAccessPattern matches ['key'], ["key"], or [`key`] (template literal)
	// Used to extract keys from bracket notation including template literals
	BracketKeyAccessPattern = regexp.MustCompile(`\[['"\x60](\w+)['"\x60]\]`)

	// RequestPropertyChainPattern matches .body.prop, .query.prop, .params.prop, etc.
	// Used to extract nested property access on request objects
	RequestPropertyChainPattern = regexp.MustCompile(`\.(body|query|params|headers|cookies)\.(\w+)`)

	// DecoratorPattern matches @Decorator(
	// Used to detect TypeScript/NestJS decorators
	DecoratorPatternPrefix = `@`
)

// BuildThisPropertyAssignPattern creates a pattern for this.property = ... paramName
func BuildThisPropertyAssignPattern(paramName string) *regexp.Regexp {
	return regexp.MustCompile(`this\.(\w+)\s*=.*\b` + regexp.QuoteMeta(paramName) + `\b`)
}

// BuildPropertyPattern creates a pattern for .propertyName with word boundary
func BuildPropertyPattern(pattern string) *regexp.Regexp {
	return regexp.MustCompile(`\.` + pattern + `\b`)
}

// BuildDecoratorPattern creates a pattern for @decoratorName(
func BuildDecoratorPattern(decoratorPattern string) *regexp.Regexp {
	return regexp.MustCompile(DecoratorPatternPrefix + decoratorPattern + `\(`)
}

// ExtractBracketKey extracts the key from bracket notation ['key'], ["key"], or [`key`]
func ExtractBracketKey(expr string) string {
	matches := BracketKeyAccessPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ExtractRequestChainProperty extracts the property from request.body.prop style access
// Returns (category, property) e.g., ("body", "userId")
func ExtractRequestChainProperty(expr string) (string, string) {
	matches := RequestPropertyChainPattern.FindStringSubmatch(expr)
	if len(matches) > 2 {
		return matches[1], matches[2]
	}
	return "", ""
}
```

**Step 2: Build and verify compilation**

Run: `go build ./pkg/sources/typescript/...`
Expected: Build succeeds with no errors

**Step 3: Commit**

```bash
git add pkg/sources/typescript/
git commit -m "feat(sources/typescript): create package with semantic analysis patterns"
```

---

### Task 5: Update PHP Analyzer to Use Centralized Patterns

**Files:**
- Modify: `pkg/semantic/analyzer/php/analyzer.go`

**Step 1: Add import for php patterns**

Add to imports:
```go
phppatterns "github.com/hatlesswizard/inputtracer/pkg/sources/php"
```

**Step 2: Replace inline patterns in AnalyzeMethod (around line 1246)**

Replace:
```go
thisAssignRegex := regexp.MustCompile(`\$this->(\w+)\s*=.*` + regexp.QuoteMeta(paramName))
```
With:
```go
thisAssignRegex := phppatterns.BuildThisPropertyAssignPattern(paramName)
```

Replace:
```go
thisArrayAssignRegex := regexp.MustCompile(`\$this->(\w+)\[.*\]\s*=.*` + regexp.QuoteMeta(paramName))
```
With:
```go
thisArrayAssignRegex := phppatterns.BuildThisArrayPropertyAssignPattern(paramName)
```

**Step 3: Replace inline patterns in matchesFrameworkPattern (around line 1362)**

Replace:
```go
regex := regexp.MustCompile(`\$\w+->(` + regexp.QuoteMeta(pattern.CarrierProperty) + `)(\[|$)`)
```
With:
```go
regex := phppatterns.BuildPropertyAccessPattern(pattern.CarrierProperty)
```

Replace:
```go
regex := regexp.MustCompile(`->` + pattern.MethodPattern + `\(`)
```
With:
```go
regex := phppatterns.BuildMethodCallPattern(pattern.MethodPattern)
```

**Step 4: Replace inline patterns in extractKeyFromExpression (around line 1378)**

Replace:
```go
regex := regexp.MustCompile(`\[['"](\w+)['"]\]`)
```
With:
```go
regex := phppatterns.ArrayKeyAccessPattern
```

Replace:
```go
regex2 := regexp.MustCompile(`\[(\$\w+)\]`)
```
With:
```go
regex2 := phppatterns.VariableKeyAccessPattern
```

**Step 5: Build and test**

Run: `go build ./pkg/semantic/analyzer/php/... && go test ./pkg/semantic/analyzer/php/...`
Expected: Build and tests pass

**Step 6: Commit**

```bash
git add pkg/semantic/analyzer/php/analyzer.go
git commit -m "refactor(semantic/php): use centralized patterns from sources/php"
```

---

### Task 6: Update Discovery Taint to Use Centralized Patterns

**Files:**
- Modify: `pkg/semantic/discovery/taint.go`

**Step 1: Verify import exists (should already have it)**

Ensure this import exists:
```go
phppatterns "github.com/hatlesswizard/inputtracer/pkg/sources/php"
```

**Step 2: Replace inline patterns (around line 771 and 807)**

Replace:
```go
returnPattern := regexp.MustCompile(`return\s+\$this->` + regexp.QuoteMeta(propName))
```
With:
```go
returnPattern := phppatterns.BuildReturnPropertyPattern(propName)
```

Replace:
```go
paramReturnPattern := regexp.MustCompile(`return\s+\$this->` + regexp.QuoteMeta(propName) + `\[`)
```
With:
```go
paramReturnPattern := phppatterns.BuildReturnPropertyArrayPattern(propName)
```

**Step 3: Build and test**

Run: `go build ./pkg/semantic/discovery/... && go test ./pkg/semantic/discovery/...`
Expected: Build and tests pass

**Step 4: Commit**

```bash
git add pkg/semantic/discovery/taint.go
git commit -m "refactor(semantic/discovery): use centralized patterns from sources/php"
```

---

### Task 7: Update JavaScript Analyzer to Use Centralized Patterns

**Files:**
- Modify: `pkg/semantic/analyzer/javascript/analyzer.go`

**Step 1: Add import for javascript patterns**

Add to imports:
```go
jspatterns "github.com/hatlesswizard/inputtracer/pkg/sources/javascript"
```

**Step 2: Replace inline patterns (around line 890)**

Replace:
```go
keyRegex := regexp.MustCompile(`\.get\(['"](\w+)['"]\)`)
```
With:
```go
keyRegex := jspatterns.MapGetPattern
```

**Step 3: Replace inline patterns in extractJSKey (around line 941-947)**

Replace:
```go
bracketRegex := regexp.MustCompile(`^\[['"](\w+)['"]\]`)
```
With:
```go
bracketRegex := jspatterns.BracketPropertyPattern
```

Replace:
```go
dotRegex := regexp.MustCompile(`^\.(\w+)`)
```
With:
```go
dotRegex := jspatterns.DotPropertyPattern
```

**Step 4: Replace inline patterns (around line 1030)**

Replace:
```go
thisAssignRegex := regexp.MustCompile(`this\.(\w+)\s*=.*` + regexp.QuoteMeta(paramName))
```
With:
```go
thisAssignRegex := jspatterns.BuildThisPropertyAssignPattern(paramName)
```

**Step 5: Build and test**

Run: `go build ./pkg/semantic/analyzer/javascript/... && go test ./pkg/semantic/analyzer/javascript/...`
Expected: Build and tests pass

**Step 6: Commit**

```bash
git add pkg/semantic/analyzer/javascript/analyzer.go
git commit -m "refactor(semantic/javascript): use centralized patterns from sources/javascript"
```

---

### Task 8: Update Python Analyzer to Use Centralized Patterns

**Files:**
- Modify: `pkg/semantic/analyzer/python/analyzer.go`

**Step 1: Add import for python patterns**

Add to imports:
```go
pypatterns "github.com/hatlesswizard/inputtracer/pkg/sources/python"
```

**Step 2: Replace inline patterns (around line 859)**

Replace:
```go
selfAssignRegex := regexp.MustCompile(`self\.(\w+)\s*=.*\b` + regexp.QuoteMeta(paramName) + `\b`)
```
With:
```go
selfAssignRegex := pypatterns.BuildSelfPropertyAssignPattern(paramName)
```

**Step 3: Replace inline patterns in matchesFrameworkPattern (around line 940-945)**

Replace:
```go
regex := regexp.MustCompile(`\b` + pattern.PropertyPattern)
```
With:
```go
regex := pypatterns.BuildPropertyPattern(pattern.PropertyPattern)
```

Replace:
```go
regex := regexp.MustCompile(`\.` + pattern.MethodPattern + `\(`)
```
With:
```go
regex := pypatterns.BuildMethodCallPattern(pattern.MethodPattern)
```

**Step 4: Replace inline patterns in extractKeyFromExpression (around line 955-962)**

Replace:
```go
regex := regexp.MustCompile(`\[['"](\w+)['"]\]`)
```
With:
```go
regex := pypatterns.DictKeyAccessPattern
```

Replace:
```go
regex2 := regexp.MustCompile(`\.get\(['"](\w+)['"]\)`)
```
With:
```go
regex2 := pypatterns.DictGetPattern
```

**Step 5: Build and test**

Run: `go build ./pkg/semantic/analyzer/python/... && go test ./pkg/semantic/analyzer/python/...`
Expected: Build and tests pass

**Step 6: Commit**

```bash
git add pkg/semantic/analyzer/python/analyzer.go
git commit -m "refactor(semantic/python): use centralized patterns from sources/python"
```

---

### Task 9: Update TypeScript Analyzer to Use Centralized Patterns

**Files:**
- Modify: `pkg/semantic/analyzer/typescript/analyzer.go`

**Step 1: Add import for typescript patterns**

Add to imports:
```go
tspatterns "github.com/hatlesswizard/inputtracer/pkg/sources/typescript"
```

**Step 2: Replace inline patterns (around line 627)**

Replace:
```go
thisAssignRegex := regexp.MustCompile(`this\.(\w+)\s*=.*\b` + regexp.QuoteMeta(param.Name) + `\b`)
```
With:
```go
thisAssignRegex := tspatterns.BuildThisPropertyAssignPattern(param.Name)
```

**Step 3: Replace inline patterns in matchesFrameworkPattern (around line 702-707)**

Replace:
```go
regex := regexp.MustCompile(`\.` + pattern.PropertyPattern + `\b`)
```
With:
```go
regex := tspatterns.BuildPropertyPattern(pattern.PropertyPattern)
```

Replace:
```go
regex := regexp.MustCompile(`@` + pattern.MethodPattern + `\(`)
```
With:
```go
regex := tspatterns.BuildDecoratorPattern(pattern.MethodPattern)
```

**Step 4: Replace inline patterns in extractKeyFromExpression (around line 715-721)**

Replace:
```go
regex := regexp.MustCompile(`\[['"\x60](\w+)['"\x60]\]`)
```
With:
```go
regex := tspatterns.BracketKeyAccessPattern
```

Replace:
```go
regex2 := regexp.MustCompile(`\.(body|query|params|headers|cookies)\.(\w+)`)
```
With:
```go
regex2 := tspatterns.RequestPropertyChainPattern
```

**Step 5: Build and test**

Run: `go build ./pkg/semantic/analyzer/typescript/... && go test ./pkg/semantic/analyzer/typescript/...`
Expected: Build and tests pass

**Step 6: Commit**

```bash
git add pkg/semantic/analyzer/typescript/analyzer.go
git commit -m "refactor(semantic/typescript): use centralized patterns from sources/typescript"
```

---

### Task 10: Final Verification

**Step 1: Build entire project**

Run: `go build ./...`
Expected: Build succeeds

**Step 2: Run all tests**

Run: `go test ./...`
Expected: All tests pass

**Step 3: Final commit if any cleanup needed**

```bash
git status
# If clean, done. Otherwise commit any remaining changes.
```
