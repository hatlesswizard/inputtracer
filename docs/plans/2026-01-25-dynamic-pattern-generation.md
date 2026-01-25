# Dynamic Pattern Generation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace static method mappings in genpatterns with dynamic source type inference from method names.

**Architecture:** Remove `LaravelMethodMappings`/`SymfonyMethodMappings` static maps. Add `InferSourceType()` that uses context-aware heuristics (e.g., `query`→GET, `post`→POST, `get`→UserInput). Use exclusion list (blacklist) instead of inclusion list (whitelist).

**Tech Stack:** Go, regex for PHP parsing

---

## Task 1: Create inference.go

**Files:**
- Create: `cmd/genpatterns/inference.go`

**Step 1: Create inference.go with InferSourceType**

```go
// Package main - inference.go provides dynamic source type inference from method names
package main

import "strings"

// InferSourceType determines SourceType from method/property name.
// Uses context-aware approach: 'get' is generic, only specific names map to specific types.
func InferSourceType(name string) string {
	lower := strings.ToLower(name)

	// Exact matches for specific types (context-aware)
	switch lower {
	case "query":
		return "SourceHTTPGet"
	case "post":
		return "SourceHTTPPost"
	case "cookie", "cookies", "hascookie":
		return "SourceHTTPCookie"
	case "header", "headers", "hasheader", "bearertoken":
		return "SourceHTTPHeader"
	case "file", "files", "allfiles", "hasfile":
		return "SourceHTTPFile"
	case "json", "getpayload", "getcontent", "toarray":
		return "SourceHTTPBody"
	case "server":
		return "SourceEnvVar"
	case "session", "oldinput", "flash", "old", "getoldcollection":
		return "SourceSession"
	}

	// Partial matches for fallback
	switch {
	case strings.Contains(lower, "cookie"):
		return "SourceHTTPCookie"
	case strings.Contains(lower, "header"):
		return "SourceHTTPHeader"
	case strings.Contains(lower, "file"):
		return "SourceHTTPFile"
	case strings.Contains(lower, "flash"):
		return "SourceSession"
	default:
		return "SourceUserInput"
	}
}

// InferPopulatedFrom determines the PHP superglobals that populate this source type.
func InferPopulatedFrom(sourceType string) []string {
	switch sourceType {
	case "SourceHTTPGet":
		return []string{"$_GET"}
	case "SourceHTTPPost":
		return []string{"$_POST"}
	case "SourceHTTPCookie":
		return []string{"$_COOKIE"}
	case "SourceHTTPHeader", "SourceEnvVar":
		return []string{"$_SERVER"}
	case "SourceHTTPFile":
		return []string{"$_FILES"}
	case "SourceHTTPBody":
		return []string{}
	case "SourceSession":
		return []string{"$_SESSION"}
	default:
		return []string{"$_GET", "$_POST"}
	}
}

// InferDescription generates a description for the method based on framework and type.
func InferDescription(framework, methodName string, isProperty bool, sourceType string) string {
	var typeDesc string
	switch sourceType {
	case "SourceHTTPGet":
		typeDesc = "query string parameters"
	case "SourceHTTPPost":
		typeDesc = "POST data"
	case "SourceHTTPCookie":
		typeDesc = "cookie values"
	case "SourceHTTPHeader":
		typeDesc = "HTTP headers"
	case "SourceHTTPFile":
		typeDesc = "uploaded files"
	case "SourceHTTPBody":
		typeDesc = "request body"
	case "SourceEnvVar":
		typeDesc = "server variables"
	case "SourceSession":
		typeDesc = "session/flash data"
	default:
		typeDesc = "user input"
	}

	frameworkTitle := strings.Title(framework)
	if isProperty {
		return frameworkTitle + " $request->" + methodName + " contains " + typeDesc
	}
	return frameworkTitle + " $request->" + methodName + "() returns " + typeDesc
}
```

**Step 2: Verify it compiles**

Run: `go build ./cmd/genpatterns/inference.go`
Expected: No errors

**Step 3: Commit**

```bash
git add cmd/genpatterns/inference.go
git commit -m "feat(genpatterns): add source type inference from method names"
```

---

## Task 2: Create exclusions.go

**Files:**
- Create: `cmd/genpatterns/exclusions.go`

**Step 1: Create exclusions.go with method blacklist**

```go
// Package main - exclusions.go defines methods to exclude from pattern generation
package main

// ExcludedMethods contains method names that don't return user input.
// These are utility methods, setters, or boolean checks without data access.
var ExcludedMethods = map[string]bool{
	// Boolean checks (don't return data)
	"has":           true,
	"hasAny":        true,
	"filled":        true,
	"isNotFilled":   true,
	"anyFilled":     true,
	"missing":       true,
	"isEmptyString": true,
	"exists":        true,
	"hasSession":    true,
	"isSecure":      true,
	"ajax":          true,
	"pjax":          true,
	"prefetch":      true,
	"wantsJson":     true,
	"acceptsAnyContentType": true,
	"acceptsJson":   true,
	"acceptsHtml":   true,
	"prefers":       true,

	// Request metadata (not user-controlled input)
	"method":       true,
	"path":         true,
	"decodedPath":  true,
	"url":          true,
	"fullUrl":      true,
	"fullUrlWithQuery": true,
	"fullUrlWithoutQuery": true,
	"root":         true,
	"route":        true,
	"routeIs":      true,
	"is":           true,
	"segment":      true,
	"segments":     true,
	"ip":           true,
	"ips":          true,
	"userAgent":    true,
	"fingerprint":  true,
	"host":         true,
	"httpHost":     true,
	"schemeAndHttpHost": true,

	// Setters/mutators
	"merge":           true,
	"mergeIfMissing":  true,
	"replace":         true,
	"set":             true,
	"add":             true,
	"remove":          true,
	"offsetSet":       true,
	"offsetUnset":     true,
	"offsetExists":    true,

	// Internal/magic methods
	"count":      true,
	"getIterator": true,
	"keys":       true, // Returns keys, not values

	// Symfony-specific non-input
	"getSession":       true,
	"hasPreviousSession": true,
	"isMethodSafe":     true,
	"isMethodIdempotent": true,
	"isMethodCacheable": true,
	"getProtocolVersion": true,
	"getContentType": true,
	"getContentTypeFormat": true,
	"getDefaultLocale": true,
	"getLocale":        true,
	"setLocale":        true,
	"getFormat":        true,
	"setFormat":        true,
	"getMimeType":      true,
	"getMimeTypes":     true,
	"isXmlHttpRequest": true,
	"preferSafeContent": true,
	"isFromTrustedProxy": true,
}

// IsExcluded returns true if the method should not generate a pattern.
func IsExcluded(methodName string) bool {
	return ExcludedMethods[methodName]
}
```

**Step 2: Verify it compiles**

Run: `go build ./cmd/genpatterns/exclusions.go`
Expected: No errors

**Step 3: Commit**

```bash
git add cmd/genpatterns/exclusions.go
git commit -m "feat(genpatterns): add method exclusion list for non-input methods"
```

---

## Task 3: Update frameworks.go

**Files:**
- Modify: `cmd/genpatterns/frameworks.go`

**Step 1: Add more Laravel source URLs**

Add to Laravel Sources array:
```go
{URL: "https://raw.githubusercontent.com/illuminate/http/master/Concerns/InteractsWithFlashData.php", ClassName: "InteractsWithFlashData"},
```

**Step 2: Remove static mapping variables**

Delete these entire variables:
- `LaravelMethodMappings`
- `SymfonyMethodMappings`
- `SymfonyPropertyMappings`

Keep only:
- `FrameworkSource` struct
- `FrameworkDefinition` struct
- `MethodMapping` struct (still useful for Symfony properties)
- `Frameworks` map
- `SymfonyPropertyMappings` (keep this one - properties need explicit mapping)

**Step 3: Commit**

```bash
git add cmd/genpatterns/frameworks.go
git commit -m "refactor(genpatterns): add Laravel sources, remove method mappings"
```

---

## Task 4: Update parser.go

**Files:**
- Modify: `cmd/genpatterns/parser.go`

**Step 1: Remove FilterByMappings function**

Delete lines 84-93 (the `FilterByMappings` function).

**Step 2: Verify it compiles**

Run: `go build ./cmd/genpatterns/...`
Expected: Compile errors (main.go still references it) - this is expected

**Step 3: Commit**

```bash
git add cmd/genpatterns/parser.go
git commit -m "refactor(genpatterns): remove FilterByMappings function"
```

---

## Task 5: Update main.go

**Files:**
- Modify: `cmd/genpatterns/main.go`

**Step 1: Update generateLaravel to not filter and parse more sources**

Replace generateLaravel function:
```go
func generateLaravel(parser *Parser, generator *Generator, sources map[string]string, fw *FrameworkDefinition) string {
	var allMethods []ParsedMethod

	// Parse all Laravel input-related sources
	for className, src := range sources {
		methods := parser.ParseMethods(src, className)
		allMethods = append(allMethods, methods...)
	}

	// Filter out excluded methods (blacklist, not whitelist)
	var filtered []ParsedMethod
	for _, m := range allMethods {
		if !IsExcluded(m.Name) {
			filtered = append(filtered, m)
		}
	}

	return generator.GenerateLaravel(filtered, fw)
}
```

**Step 2: Update generateSymfony similarly**

Replace generateSymfony function:
```go
func generateSymfony(parser *Parser, generator *Generator, sources map[string]string, fw *FrameworkDefinition) string {
	var allMethods []ParsedMethod
	var allProperties []ParsedMethod

	// Parse ParameterBag and InputBag methods
	if src, ok := sources["ParameterBag"]; ok {
		methods := parser.ParseMethods(src, "ParameterBag")
		allMethods = append(allMethods, methods...)
	}
	if src, ok := sources["InputBag"]; ok {
		methods := parser.ParseMethods(src, "InputBag")
		allMethods = append(allMethods, methods...)
	}

	// Parse Request public properties (keep property mappings for now)
	if src, ok := sources["Request"]; ok {
		props := parser.ParseProperties(src, "Request")
		// Filter properties by mapping (properties still need explicit mapping)
		for _, p := range props {
			if _, ok := SymfonyPropertyMappings[p.Name]; ok {
				allProperties = append(allProperties, p)
			}
		}
	}

	// Filter out excluded methods
	var filtered []ParsedMethod
	for _, m := range allMethods {
		if !IsExcluded(m.Name) {
			filtered = append(filtered, m)
		}
	}

	return generator.GenerateSymfony(filtered, allProperties, fw)
}
```

**Step 3: Verify it compiles**

Run: `go build ./cmd/genpatterns/...`
Expected: Compile errors in generator.go (still uses old mappings)

**Step 4: Commit**

```bash
git add cmd/genpatterns/main.go
git commit -m "refactor(genpatterns): use exclusion list instead of inclusion mappings"
```

---

## Task 6: Update generator.go

**Files:**
- Modify: `cmd/genpatterns/generator.go`

**Step 1: Update GenerateLaravel to use inference**

Replace the method loop in GenerateLaravel:
```go
func (g *Generator) GenerateLaravel(methods []ParsedMethod, fw *FrameworkDefinition) string {
	var b strings.Builder

	g.writeHeader(&b, fw, "https://github.com/illuminate/http")
	g.writePackage(&b)
	g.writeImport(&b)

	b.WriteString("var laravelPatterns = []*common.FrameworkPattern{\n")

	for _, m := range methods {
		sourceType := InferSourceType(m.Name)
		populatedFrom := InferPopulatedFrom(sourceType)
		description := InferDescription(fw.Name, m.Name, false, sourceType)
		g.writePatternInferred(&b, fw, m, sourceType, populatedFrom, description)
	}

	b.WriteString("}\n\n")
	g.writeInit(&b, "laravel", fw)

	return b.String()
}
```

**Step 2: Update GenerateSymfony similarly**

```go
func (g *Generator) GenerateSymfony(methods []ParsedMethod, properties []ParsedMethod, fw *FrameworkDefinition) string {
	var b strings.Builder

	g.writeHeader(&b, fw, "https://github.com/symfony/http-foundation")
	g.writePackage(&b)
	g.writeImport(&b)

	b.WriteString("var symfonyPatterns = []*common.FrameworkPattern{\n")

	// Properties still use explicit mapping
	for _, p := range properties {
		mapping := SymfonyPropertyMappings[p.Name]
		if mapping == nil {
			continue
		}
		g.writePropertyPattern(&b, fw, p, mapping)
	}

	// Methods use inference
	for _, m := range methods {
		sourceType := InferSourceType(m.Name)
		populatedFrom := InferPopulatedFrom(sourceType)
		description := InferDescription(fw.Name, m.Name, false, sourceType)
		g.writeSymfonyMethodPatternInferred(&b, fw, m, sourceType, populatedFrom, description)
	}

	b.WriteString("}\n\n")
	g.writeInit(&b, "symfony", fw)

	return b.String()
}
```

**Step 3: Add writePatternInferred helper**

```go
func (g *Generator) writePatternInferred(b *strings.Builder, fw *FrameworkDefinition, m ParsedMethod, sourceType string, populatedFrom []string, description string) {
	id := fmt.Sprintf("%s_request_%s", fw.Name, m.Name)
	name := fmt.Sprintf("%s $request->%s()", strings.Title(fw.Name), m.Name)

	b.WriteString("\t{\n")
	b.WriteString(fmt.Sprintf("\t\tID:            %q,\n", id))
	b.WriteString(fmt.Sprintf("\t\tFramework:     %q,\n", fw.Name))
	b.WriteString(fmt.Sprintf("\t\tLanguage:      %q,\n", fw.Language))
	b.WriteString(fmt.Sprintf("\t\tName:          %q,\n", name))
	b.WriteString(fmt.Sprintf("\t\tDescription:   %q,\n", description))
	b.WriteString(fmt.Sprintf("\t\tClassPattern:  %q,\n", fw.ClassPattern))
	b.WriteString(fmt.Sprintf("\t\tMethodPattern: \"^%s$\",\n", m.Name))
	b.WriteString(fmt.Sprintf("\t\tSourceType:    common.%s,\n", sourceType))
	b.WriteString(fmt.Sprintf("\t\tCarrierClass:  %q,\n", fw.CarrierClass))
	if len(populatedFrom) > 0 {
		b.WriteString(fmt.Sprintf("\t\tPopulatedFrom: []string{%s},\n", g.formatStringSlice(populatedFrom)))
	}
	b.WriteString(fmt.Sprintf("\t\tTags:          []string{%s},\n", g.formatStringSlice(fw.Tags)))
	b.WriteString("\t},\n")
}

func (g *Generator) writeSymfonyMethodPatternInferred(b *strings.Builder, fw *FrameworkDefinition, m ParsedMethod, sourceType string, populatedFrom []string, description string) {
	classPattern := "^(Symfony\\\\\\\\Component\\\\\\\\HttpFoundation\\\\\\\\)?ParameterBag$"
	id := fmt.Sprintf("%s_parameterbag_%s", fw.Name, m.Name)
	name := fmt.Sprintf("%s ParameterBag->%s()", strings.Title(fw.Name), m.Name)

	b.WriteString("\t{\n")
	b.WriteString(fmt.Sprintf("\t\tID:            %q,\n", id))
	b.WriteString(fmt.Sprintf("\t\tFramework:     %q,\n", fw.Name))
	b.WriteString(fmt.Sprintf("\t\tLanguage:      %q,\n", fw.Language))
	b.WriteString(fmt.Sprintf("\t\tName:          %q,\n", name))
	b.WriteString(fmt.Sprintf("\t\tDescription:   %q,\n", description))
	b.WriteString(fmt.Sprintf("\t\tClassPattern:  %q,\n", classPattern))
	b.WriteString(fmt.Sprintf("\t\tMethodPattern: \"^%s$\",\n", m.Name))
	b.WriteString(fmt.Sprintf("\t\tSourceType:    common.%s,\n", sourceType))
	if len(populatedFrom) > 0 {
		b.WriteString(fmt.Sprintf("\t\tPopulatedFrom: []string{%s},\n", g.formatStringSlice(populatedFrom)))
	}
	b.WriteString(fmt.Sprintf("\t\tTags:          []string{%s},\n", g.formatStringSlice(fw.Tags)))
	b.WriteString("\t},\n")
}
```

**Step 4: Remove old writePattern that uses mappings**

Delete the old `writePattern` and `writeSymfonyMethodPattern` functions.

**Step 5: Verify full build**

Run: `go build ./cmd/genpatterns`
Expected: Success

**Step 6: Commit**

```bash
git add cmd/genpatterns/generator.go
git commit -m "refactor(genpatterns): use inference for source types in generator"
```

---

## Task 7: Build and Generate

**Step 1: Build genpatterns**

Run: `go build -o genpatterns ./cmd/genpatterns`
Expected: Success, creates `genpatterns` binary

**Step 2: Generate new pattern files**

Run: `./genpatterns -o pkg/sources/php/`
Expected:
- Outputs "Fetching laravel sources..."
- Outputs "Generated pkg/sources/php/laravel.go"
- Outputs "Fetching symfony sources..."
- Outputs "Generated pkg/sources/php/symfony.go"
- Outputs "Done!"

**Step 3: Verify generated files have MORE methods**

Run: `grep -c "ID:" pkg/sources/php/laravel.go`
Expected: More patterns than before (was ~25, should be 30+)

**Step 4: Run all tests**

Run: `go test ./...`
Expected: All tests pass

**Step 5: Commit generated files**

```bash
git add pkg/sources/php/laravel.go pkg/sources/php/symfony.go
git commit -m "chore: regenerate framework patterns with dynamic inference"
```

---

## Verification Checklist

- [ ] `go build ./cmd/genpatterns` succeeds
- [ ] `./genpatterns -o pkg/sources/php/` generates both files
- [ ] Generated laravel.go has more methods than before
- [ ] Source types are reasonable (query→GET, post→POST, get→UserInput)
- [ ] `go test ./...` passes
- [ ] No phantom methods (all methods exist in fetched sources)
