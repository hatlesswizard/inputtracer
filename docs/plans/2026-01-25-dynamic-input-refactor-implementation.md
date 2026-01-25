# Dynamic Input Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor InputTracer to use a fully dynamic, centralized input detection system with zero duplication and correct categorization.

**Architecture:** Create a core registry in `pkg/sources/core/` that holds all input definitions. Language-specific files register patterns with the core. Analyzers query the registry instead of using hardcoded patterns. WordPress support added dynamically.

**Tech Stack:** Go 1.23, Tree-sitter, regexp

---

## Phase 1: Create Core Infrastructure

### Task 1.1: Create Core Types

**Files:**
- Create: `pkg/sources/core/types.go`

**Step 1: Create the core directory**

```bash
mkdir -p pkg/sources/core
```

**Step 2: Write the consolidated types file**

Create `pkg/sources/core/types.go` with ALL type definitions consolidated from multiple files:

```go
// Package core provides the centralized type definitions and registry for input detection.
// This is the SINGLE SOURCE OF TRUTH for all input-related types.
package core

// SourceType categorizes the origin of input data
type SourceType string

const (
    SourceHTTPGet     SourceType = "http_get"
    SourceHTTPPost    SourceType = "http_post"
    SourceHTTPBody    SourceType = "http_body"
    SourceHTTPJSON    SourceType = "http_json"
    SourceHTTPHeader  SourceType = "http_header"
    SourceHTTPCookie  SourceType = "http_cookie"
    SourceHTTPPath    SourceType = "http_path"
    SourceHTTPFile    SourceType = "http_file"
    SourceHTTPRequest SourceType = "http_request"
    SourceSession     SourceType = "session"
    SourceCLIArg      SourceType = "cli_arg"
    SourceEnvVar      SourceType = "env_var"
    SourceStdin       SourceType = "stdin"
    SourceFile        SourceType = "file"
    SourceDatabase    SourceType = "database"
    SourceNetwork     SourceType = "network"
    SourceUserInput   SourceType = "user_input"
    SourceUnknown     SourceType = "unknown"
)

// InputLabel provides additional categorization for input sources
type InputLabel string

const (
    LabelHTTPGet     InputLabel = "http_get"
    LabelHTTPPost    InputLabel = "http_post"
    LabelHTTPCookie  InputLabel = "http_cookie"
    LabelHTTPHeader  InputLabel = "http_header"
    LabelHTTPBody    InputLabel = "http_body"
    LabelCLI         InputLabel = "cli"
    LabelEnvironment InputLabel = "environment"
    LabelFile        InputLabel = "file"
    LabelDatabase    InputLabel = "database"
    LabelNetwork     InputLabel = "network"
    LabelUserInput   InputLabel = "user_input"
)

// IsUserInput returns true if this source type represents direct user input
func (s SourceType) IsUserInput() bool {
    switch s {
    case SourceHTTPGet, SourceHTTPPost, SourceHTTPBody, SourceHTTPJSON,
         SourceHTTPCookie, SourceHTTPPath, SourceHTTPFile, SourceHTTPRequest,
         SourceCLIArg, SourceStdin, SourceUserInput:
        return true
    default:
        return false
    }
}

// IsServerSideData returns true if this source type is server-controlled
func (s SourceType) IsServerSideData() bool {
    switch s {
    case SourceSession, SourceDatabase, SourceFile, SourceEnvVar:
        return true
    default:
        return false
    }
}

// LabelToSourceType maps InputLabel to SourceType
var LabelToSourceType = map[InputLabel]SourceType{
    LabelHTTPGet:     SourceHTTPGet,
    LabelHTTPPost:    SourceHTTPPost,
    LabelHTTPCookie:  SourceHTTPCookie,
    LabelHTTPHeader:  SourceHTTPHeader,
    LabelHTTPBody:    SourceHTTPBody,
    LabelCLI:         SourceCLIArg,
    LabelEnvironment: SourceEnvVar,
    LabelFile:        SourceFile,
    LabelDatabase:    SourceDatabase,
    LabelNetwork:     SourceNetwork,
    LabelUserInput:   SourceUserInput,
}

// SourceTypeToLabel maps SourceType to InputLabel
var SourceTypeToLabel = map[SourceType]InputLabel{
    SourceHTTPGet:    LabelHTTPGet,
    SourceHTTPPost:   LabelHTTPPost,
    SourceHTTPCookie: LabelHTTPCookie,
    SourceHTTPHeader: LabelHTTPHeader,
    SourceHTTPBody:   LabelHTTPBody,
    SourceCLIArg:     LabelCLI,
    SourceEnvVar:     LabelEnvironment,
    SourceFile:       LabelFile,
    SourceDatabase:   LabelDatabase,
    SourceNetwork:    LabelNetwork,
    SourceUserInput:  LabelUserInput,
}
```

**Step 3: Run tests to verify build**

```bash
go build ./pkg/sources/core/...
```

**Step 4: Commit**

```bash
git add pkg/sources/core/
git commit -m "feat(core): create centralized type definitions

- Add SourceType enum with 18 input categories
- Add InputLabel enum for additional categorization
- Add IsUserInput() and IsServerSideData() helper methods
- Add bidirectional label/type mappings
- This is the single source of truth for all input types"
```

---

### Task 1.2: Create Core Patterns

**Files:**
- Create: `pkg/sources/core/patterns.go`

**Step 1: Write universal patterns file**

Create `pkg/sources/core/patterns.go`:

```go
package core

import (
    "regexp"
    "sync"
)

// UniversalPatterns holds pre-compiled regex patterns used across all languages
type UniversalPatterns struct {
    // Input method detection
    InputMethod     *regexp.Regexp
    InputProperty   *regexp.Regexp
    InputObject     *regexp.Regexp
    ExcludeMethod   *regexp.Regexp

    // Key/property access
    ArrayKeyAccess  *regexp.Regexp
    PropertyAccess  *regexp.Regexp
    MethodCall      *regexp.Regexp

    // Assignment patterns
    SimpleAssign    *regexp.Regexp
    PropertyAssign  *regexp.Regexp
}

var (
    universalPatterns *UniversalPatterns
    patternsOnce      sync.Once
)

// GetUniversalPatterns returns the singleton universal patterns instance
func GetUniversalPatterns() *UniversalPatterns {
    patternsOnce.Do(func() {
        universalPatterns = &UniversalPatterns{
            // Methods that indicate user input retrieval
            InputMethod: regexp.MustCompile(`(?i)^(get_?)?(input|var|variable|query_?params?|parsed_?body|cookie_?params?|server_?params?|uploaded_?files?|headers?|all|body|content)$|^(get_?)?(post|cookie|param|query|header)s?$`),

            // Properties that hold user input
            InputProperty: regexp.MustCompile(`(?i)^(input|request|params?|query|cookies?|headers?|body|data|args?|post|get|files?|server|attributes?|payload)s?$`),

            // Objects that carry user input
            InputObject: regexp.MustCompile(`(?i)(request|input|req|params?|http|ctx|context)`),

            // Methods to exclude (false positives)
            ExcludeMethod: regexp.MustCompile(`(?i)^(getData|getBody|getContent|fetch|find|load|read|save|store|cache|log|debug|info|warn|error)$`),

            // Key access patterns: ['key'] or ["key"]
            ArrayKeyAccess: regexp.MustCompile(`\[['"](\w+)['"]\]`),

            // Property access: .property or ->property
            PropertyAccess: regexp.MustCompile(`(?:->|\.)(\w+)`),

            // Method call: .method( or ->method(
            MethodCall: regexp.MustCompile(`(?:->|\.)(\w+)\s*\(`),

            // Simple assignment: var =
            SimpleAssign: regexp.MustCompile(`(\$?\w+)\s*=\s*`),

            // Property assignment: obj.prop = or obj->prop =
            PropertyAssign: regexp.MustCompile(`(\$?\w+)(?:->|\.)(\w+)\s*=\s*`),
        }
    })
    return universalPatterns
}

// IsInputMethod checks if a method name indicates input retrieval
func IsInputMethod(methodName string) bool {
    return GetUniversalPatterns().InputMethod.MatchString(methodName)
}

// IsInputProperty checks if a property name holds input data
func IsInputProperty(propName string) bool {
    return GetUniversalPatterns().InputProperty.MatchString(propName)
}

// IsInputObject checks if an object name carries input
func IsInputObject(objName string) bool {
    return GetUniversalPatterns().InputObject.MatchString(objName)
}

// IsExcludedMethod checks if a method should be excluded from input detection
func IsExcludedMethod(methodName string) bool {
    return GetUniversalPatterns().ExcludeMethod.MatchString(methodName)
}

// ExtractKey extracts the key from array/property access expressions
func ExtractKey(expr string) string {
    if match := GetUniversalPatterns().ArrayKeyAccess.FindStringSubmatch(expr); len(match) > 1 {
        return match[1]
    }
    return ""
}
```

**Step 2: Build and verify**

```bash
go build ./pkg/sources/core/...
```

**Step 3: Commit**

```bash
git add pkg/sources/core/patterns.go
git commit -m "feat(core): add universal regex patterns

- Add UniversalPatterns struct with compiled regexes
- Add helper functions: IsInputMethod, IsInputProperty, IsInputObject
- Use sync.Once for thread-safe singleton initialization
- Patterns are language-agnostic and reusable"
```

---

### Task 1.3: Create Core Registry

**Files:**
- Create: `pkg/sources/core/registry.go`

**Step 1: Write the registry**

Create `pkg/sources/core/registry.go`:

```go
package core

import (
    "regexp"
    "strings"
    "sync"
)

// InputPattern defines a pattern for detecting input sources
type InputPattern struct {
    Name        string       // Unique identifier
    Description string       // Human-readable description
    Category    SourceType   // Primary category
    Labels      []InputLabel // Additional labels
    Language    string       // Target language (empty = all)
    Framework   string       // Target framework (empty = all)

    // Pattern matching (use one or more)
    ExactMatch    string         // Exact string match
    Regex         *regexp.Regexp // Compiled regex
    MethodName    string         // Method name to match
    PropertyName  string         // Property name to match
    ObjectPattern string         // Object name pattern

    // Context requirements
    RequireObject bool   // Must be called on an object
    ObjectType    string // Required object type (if known)
    ParamIndex    int    // Which parameter receives input (-1 = return value)
}

// Registry holds all registered input patterns
type Registry struct {
    mu sync.RWMutex

    // Fast exact match lookup
    exactPatterns map[string]*InputPattern

    // Regex patterns (checked in order)
    regexPatterns []*InputPattern

    // Language-specific patterns
    languagePatterns map[string][]*InputPattern

    // Framework-specific patterns
    frameworkPatterns map[string][]*InputPattern

    // Non-input patterns (explicitly excluded)
    nonInputPatterns map[string]bool
}

var (
    globalRegistry *Registry
    registryOnce   sync.Once
)

// GetRegistry returns the global registry singleton
func GetRegistry() *Registry {
    registryOnce.Do(func() {
        globalRegistry = &Registry{
            exactPatterns:     make(map[string]*InputPattern),
            regexPatterns:     make([]*InputPattern, 0, 100),
            languagePatterns:  make(map[string][]*InputPattern),
            frameworkPatterns: make(map[string][]*InputPattern),
            nonInputPatterns:  make(map[string]bool),
        }
    })
    return globalRegistry
}

// Register adds a pattern to the registry
func (r *Registry) Register(pattern *InputPattern) {
    r.mu.Lock()
    defer r.mu.Unlock()

    // Index by exact match
    if pattern.ExactMatch != "" {
        r.exactPatterns[pattern.ExactMatch] = pattern
    }

    // Add to regex list
    if pattern.Regex != nil {
        r.regexPatterns = append(r.regexPatterns, pattern)
    }

    // Index by language
    if pattern.Language != "" {
        r.languagePatterns[pattern.Language] = append(
            r.languagePatterns[pattern.Language], pattern)
    }

    // Index by framework
    if pattern.Framework != "" {
        r.frameworkPatterns[pattern.Framework] = append(
            r.frameworkPatterns[pattern.Framework], pattern)
    }
}

// RegisterNonInput marks a pattern as explicitly NOT user input
func (r *Registry) RegisterNonInput(pattern string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.nonInputPatterns[strings.ToLower(pattern)] = true
}

// IsNonInput checks if a pattern is explicitly marked as non-input
func (r *Registry) IsNonInput(expr string) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.nonInputPatterns[strings.ToLower(expr)]
}

// MatchResult contains the result of a pattern match
type MatchResult struct {
    Pattern  *InputPattern
    Category SourceType
    Labels   []InputLabel
    Key      string // Extracted key if applicable
}

// Match attempts to match an expression against registered patterns
func (r *Registry) Match(expr string, language string, framework string) *MatchResult {
    r.mu.RLock()
    defer r.mu.RUnlock()

    // Check non-input first
    if r.nonInputPatterns[strings.ToLower(expr)] {
        return nil
    }

    // Try exact match (fastest)
    if pattern, ok := r.exactPatterns[expr]; ok {
        if r.patternApplies(pattern, language, framework) {
            return &MatchResult{
                Pattern:  pattern,
                Category: pattern.Category,
                Labels:   pattern.Labels,
            }
        }
    }

    // Try framework-specific patterns
    if framework != "" {
        if patterns, ok := r.frameworkPatterns[framework]; ok {
            if result := r.matchPatterns(expr, patterns, language, framework); result != nil {
                return result
            }
        }
    }

    // Try language-specific patterns
    if language != "" {
        if patterns, ok := r.languagePatterns[language]; ok {
            if result := r.matchPatterns(expr, patterns, language, framework); result != nil {
                return result
            }
        }
    }

    // Try regex patterns
    return r.matchPatterns(expr, r.regexPatterns, language, framework)
}

func (r *Registry) matchPatterns(expr string, patterns []*InputPattern, language string, framework string) *MatchResult {
    for _, pattern := range patterns {
        if !r.patternApplies(pattern, language, framework) {
            continue
        }

        if pattern.Regex != nil && pattern.Regex.MatchString(expr) {
            result := &MatchResult{
                Pattern:  pattern,
                Category: pattern.Category,
                Labels:   pattern.Labels,
            }
            // Try to extract key
            if matches := pattern.Regex.FindStringSubmatch(expr); len(matches) > 1 {
                result.Key = matches[1]
            }
            return result
        }
    }
    return nil
}

func (r *Registry) patternApplies(pattern *InputPattern, language string, framework string) bool {
    if pattern.Language != "" && pattern.Language != language {
        return false
    }
    if pattern.Framework != "" && pattern.Framework != framework {
        return false
    }
    return true
}

// MatchMethod checks if a method call is an input source
func (r *Registry) MatchMethod(objName string, methodName string, language string, framework string) *MatchResult {
    // Build expression to match
    expr := methodName
    if objName != "" {
        expr = objName + "." + methodName
    }

    // Check registry first
    if result := r.Match(expr, language, framework); result != nil {
        return result
    }

    // Fall back to universal patterns
    if IsInputMethod(methodName) && !IsExcludedMethod(methodName) {
        if objName == "" || IsInputObject(objName) {
            return &MatchResult{
                Category: SourceUserInput,
                Labels:   []InputLabel{LabelUserInput},
            }
        }
    }

    return nil
}

// MatchProperty checks if a property access is an input source
func (r *Registry) MatchProperty(objName string, propName string, language string, framework string) *MatchResult {
    expr := objName + "." + propName

    if result := r.Match(expr, language, framework); result != nil {
        return result
    }

    // Fall back to universal patterns
    if IsInputProperty(propName) && IsInputObject(objName) {
        return &MatchResult{
            Category: SourceUserInput,
            Labels:   []InputLabel{LabelUserInput},
        }
    }

    return nil
}
```

**Step 2: Build and verify**

```bash
go build ./pkg/sources/core/...
```

**Step 3: Commit**

```bash
git add pkg/sources/core/registry.go
git commit -m "feat(core): add centralized input pattern registry

- Add InputPattern struct for defining input sources
- Add Registry with exact, regex, language, framework indexing
- Add Match, MatchMethod, MatchProperty for detection
- Add non-input registration for explicit exclusions
- Thread-safe with RWMutex"
```

---

### Task 1.4: Create Input Characteristics

**Files:**
- Create: `pkg/sources/core/characteristics.go`

**Step 1: Write characteristics definitions**

Create `pkg/sources/core/characteristics.go`:

```go
package core

// Characteristic defines what makes something an input source
type Characteristic struct {
    Description string       // Human-readable description
    Indicators  []string     // Method/property names that indicate this type
    Category    SourceType   // Primary category
    Labels      []InputLabel // Labels to apply
    IsUserInput bool         // True if this is user-controlled input
    Reason      string       // Why this categorization
}

// UserInputCharacteristics defines patterns that ARE user input
var UserInputCharacteristics = []Characteristic{
    {
        Description: "HTTP GET query parameters",
        Indicators: []string{
            "query", "queryParams", "query_params", "searchParams",
            "getQueryParam", "getQuery", "query_string", "qs",
            "GET", "$_GET", "request.query", "req.query",
        },
        Category:    SourceHTTPGet,
        Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
        IsUserInput: true,
        Reason:      "Query parameters are directly controlled by the client URL",
    },
    {
        Description: "HTTP POST body data",
        Indicators: []string{
            "body", "postData", "formData", "requestBody", "parsedBody",
            "getBody", "getParsedBody", "input", "post",
            "POST", "$_POST", "request.body", "req.body",
        },
        Category:    SourceHTTPPost,
        Labels:      []InputLabel{LabelHTTPPost, LabelUserInput},
        IsUserInput: true,
        Reason:      "POST body is directly sent by the client",
    },
    {
        Description: "HTTP cookies",
        Indicators: []string{
            "cookies", "cookie", "getCookie", "cookieParams",
            "COOKIE", "$_COOKIE", "request.cookies", "req.cookies",
        },
        Category:    SourceHTTPCookie,
        Labels:      []InputLabel{LabelHTTPCookie, LabelUserInput},
        IsUserInput: true,
        Reason:      "Cookies are sent by the client in the request",
    },
    {
        Description: "HTTP headers",
        Indicators: []string{
            "headers", "header", "getHeader", "getHeaders",
            "SERVER", "$_SERVER", "request.headers", "req.headers",
        },
        Category:    SourceHTTPHeader,
        Labels:      []InputLabel{LabelHTTPHeader, LabelUserInput},
        IsUserInput: true,
        Reason:      "HTTP headers are sent by the client",
    },
    {
        Description: "HTTP file uploads",
        Indicators: []string{
            "files", "uploadedFiles", "getUploadedFiles", "file",
            "FILES", "$_FILES", "request.files", "req.files",
        },
        Category:    SourceHTTPFile,
        Labels:      []InputLabel{LabelFile, LabelUserInput},
        IsUserInput: true,
        Reason:      "Uploaded files are sent by the client",
    },
    {
        Description: "HTTP request (generic)",
        Indicators: []string{
            "REQUEST", "$_REQUEST", "all", "input", "getInput",
        },
        Category:    SourceHTTPRequest,
        Labels:      []InputLabel{LabelUserInput},
        IsUserInput: true,
        Reason:      "Generic request data is client-controlled",
    },
    {
        Description: "Command line arguments",
        Indicators: []string{
            "argv", "args", "arguments", "sys.argv", "os.Args",
            "process.argv", "ARGV", "$argv",
        },
        Category:    SourceCLIArg,
        Labels:      []InputLabel{LabelCLI, LabelUserInput},
        IsUserInput: true,
        Reason:      "CLI arguments are provided by the user",
    },
    {
        Description: "Standard input",
        Indicators: []string{
            "stdin", "STDIN", "sys.stdin", "os.Stdin",
            "process.stdin", "php://input", "readline",
        },
        Category:    SourceStdin,
        Labels:      []InputLabel{LabelUserInput},
        IsUserInput: true,
        Reason:      "Standard input is provided by the user",
    },
}

// NonUserInputCharacteristics defines patterns that are NOT user input
var NonUserInputCharacteristics = []Characteristic{
    {
        Description: "Session data (server-side storage)",
        Indicators: []string{
            "session", "getSession", "sessionData", "SESSION", "$_SESSION",
            "request.session", "req.session",
        },
        Category:    SourceSession,
        Labels:      []InputLabel{}, // NO LabelUserInput!
        IsUserInput: false,
        Reason:      "Session data is stored server-side, NOT sent in the request",
    },
    {
        Description: "Database query results",
        Indicators: []string{
            "fetch", "fetchAll", "fetchOne", "fetchRow", "fetchColumn",
            "query", "execute", "findOne", "findAll", "findBy",
            "get", "first", "all", "find",
        },
        Category:    SourceDatabase,
        Labels:      []InputLabel{LabelDatabase},
        IsUserInput: false,
        Reason:      "Database results may contain user data but are not direct input",
    },
    {
        Description: "File system reads",
        Indicators: []string{
            "readFile", "file_get_contents", "fread", "fgets", "file",
            "readFileSync", "fs.read", "ioutil.ReadFile", "os.ReadFile",
        },
        Category:    SourceFile,
        Labels:      []InputLabel{LabelFile},
        IsUserInput: false,
        Reason:      "File contents are server-side data, not user input",
    },
    {
        Description: "Environment variables",
        Indicators: []string{
            "env", "getenv", "environ", "ENV", "$_ENV",
            "process.env", "os.Getenv", "os.environ",
        },
        Category:    SourceEnvVar,
        Labels:      []InputLabel{LabelEnvironment},
        IsUserInput: false,
        Reason:      "Environment variables are server configuration",
    },
    {
        Description: "Cache/transient data",
        Indicators: []string{
            "cache", "getCache", "cacheGet", "transient", "getTransient",
            "memcache", "redis", "wp_cache_get",
        },
        Category:    SourceDatabase,
        Labels:      []InputLabel{},
        IsUserInput: false,
        Reason:      "Cache data is server-side storage",
    },
}

// GetCharacteristicForIndicator finds the characteristic for an indicator
func GetCharacteristicForIndicator(indicator string) *Characteristic {
    // Check user input characteristics first
    for i := range UserInputCharacteristics {
        for _, ind := range UserInputCharacteristics[i].Indicators {
            if ind == indicator {
                return &UserInputCharacteristics[i]
            }
        }
    }

    // Check non-user input characteristics
    for i := range NonUserInputCharacteristics {
        for _, ind := range NonUserInputCharacteristics[i].Indicators {
            if ind == indicator {
                return &NonUserInputCharacteristics[i]
            }
        }
    }

    return nil
}

// IsUserInputIndicator checks if an indicator represents user input
func IsUserInputIndicator(indicator string) bool {
    char := GetCharacteristicForIndicator(indicator)
    return char != nil && char.IsUserInput
}

// GetCategoryForIndicator returns the SourceType for an indicator
func GetCategoryForIndicator(indicator string) SourceType {
    char := GetCharacteristicForIndicator(indicator)
    if char != nil {
        return char.Category
    }
    return SourceUnknown
}

// GetLabelsForIndicator returns the labels for an indicator
func GetLabelsForIndicator(indicator string) []InputLabel {
    char := GetCharacteristicForIndicator(indicator)
    if char != nil {
        return char.Labels
    }
    return nil
}
```

**Step 2: Build and verify**

```bash
go build ./pkg/sources/core/...
```

**Step 3: Commit**

```bash
git add pkg/sources/core/characteristics.go
git commit -m "feat(core): add dynamic input characteristics

- Define UserInputCharacteristics for user-controlled input
- Define NonUserInputCharacteristics for server-side data
- Session explicitly marked as NOT user input
- Database, file, env, cache marked as non-user input
- Add helper functions for indicator lookup"
```

---

## Phase 2: Update Exports and Fix Bugs

### Task 2.1: Update Sources Package Exports

**Files:**
- Modify: `pkg/sources/types.go`
- Modify: `pkg/sources/labels.go`

**Step 1: Update types.go to re-export from core**

Replace `pkg/sources/types.go` content:

```go
// Package sources provides input source detection for multiple languages.
package sources

import (
    "github.com/hatlesswizard/inputtracer/pkg/sources/core"
)

// Re-export types from core for backward compatibility
type (
    SourceType   = core.SourceType
    InputLabel   = core.InputLabel
    InputPattern = core.InputPattern
    MatchResult  = core.MatchResult
)

// Re-export constants
const (
    SourceHTTPGet     = core.SourceHTTPGet
    SourceHTTPPost    = core.SourceHTTPPost
    SourceHTTPBody    = core.SourceHTTPBody
    SourceHTTPJSON    = core.SourceHTTPJSON
    SourceHTTPHeader  = core.SourceHTTPHeader
    SourceHTTPCookie  = core.SourceHTTPCookie
    SourceHTTPPath    = core.SourceHTTPPath
    SourceHTTPFile    = core.SourceHTTPFile
    SourceHTTPRequest = core.SourceHTTPRequest
    SourceSession     = core.SourceSession
    SourceCLIArg      = core.SourceCLIArg
    SourceEnvVar      = core.SourceEnvVar
    SourceStdin       = core.SourceStdin
    SourceFile        = core.SourceFile
    SourceDatabase    = core.SourceDatabase
    SourceNetwork     = core.SourceNetwork
    SourceUserInput   = core.SourceUserInput
    SourceUnknown     = core.SourceUnknown
)

const (
    LabelHTTPGet     = core.LabelHTTPGet
    LabelHTTPPost    = core.LabelHTTPPost
    LabelHTTPCookie  = core.LabelHTTPCookie
    LabelHTTPHeader  = core.LabelHTTPHeader
    LabelHTTPBody    = core.LabelHTTPBody
    LabelCLI         = core.LabelCLI
    LabelEnvironment = core.LabelEnvironment
    LabelFile        = core.LabelFile
    LabelDatabase    = core.LabelDatabase
    LabelNetwork     = core.LabelNetwork
    LabelUserInput   = core.LabelUserInput
)

// Re-export functions
var (
    GetRegistry       = core.GetRegistry
    IsInputMethod     = core.IsInputMethod
    IsInputProperty   = core.IsInputProperty
    IsInputObject     = core.IsInputObject
    IsExcludedMethod  = core.IsExcludedMethod
)
```

**Step 2: Build and verify**

```bash
go build ./pkg/sources/...
```

**Step 3: Commit**

```bash
git add pkg/sources/types.go
git commit -m "refactor(sources): re-export types from core

- All type definitions now come from core package
- Backward compatibility maintained via type aliases
- Functions re-exported for convenience"
```

---

### Task 2.2: Fix Ruby Session Bug

**Files:**
- Modify: `pkg/sources/mappings.go`

**Step 1: Find and fix the Ruby session bug**

In `pkg/sources/mappings.go`, find the Ruby mappings section and change:

```go
// BEFORE (line ~406):
"session": SourceUserInput,  // WRONG

// AFTER:
"session": SourceSession,    // CORRECT - session is server-side
```

Also fix the generic "request" mapping to be more specific:

```go
// BEFORE:
"request": SourceUserInput,  // Too broad

// AFTER:
// Remove generic "request" - let framework patterns handle it
```

**Step 2: Build and test**

```bash
go build ./pkg/sources/...
go test ./pkg/sources/...
```

**Step 3: Commit**

```bash
git add pkg/sources/mappings.go
git commit -m "fix(sources): correct Ruby session categorization

- Session data is server-side, NOT user input
- Remove overly broad 'request' mapping
- Aligns with all framework-specific definitions

BREAKING CHANGE: Ruby 'session' no longer labeled as user input"
```

---

## Phase 3: Create WordPress Support

### Task 3.1: Create WordPress Detection

**Files:**
- Create: `pkg/sources/php/wordpress.go`

**Step 1: Write WordPress patterns**

Create `pkg/sources/php/wordpress.go`:

```go
package php

import (
    "regexp"

    "github.com/hatlesswizard/inputtracer/pkg/sources/core"
)

func init() {
    registerWordPressPatterns()
}

func registerWordPressPatterns() {
    registry := core.GetRegistry()

    // WordPress REST API request methods
    wpRestMethods := []struct {
        method   string
        category core.SourceType
        labels   []core.InputLabel
        desc     string
    }{
        {"get_param", core.SourceHTTPRequest, []core.InputLabel{core.LabelUserInput}, "REST API single parameter"},
        {"get_params", core.SourceHTTPRequest, []core.InputLabel{core.LabelUserInput}, "REST API all parameters"},
        {"get_query_params", core.SourceHTTPGet, []core.InputLabel{core.LabelHTTPGet, core.LabelUserInput}, "REST API query params"},
        {"get_body_params", core.SourceHTTPPost, []core.InputLabel{core.LabelHTTPPost, core.LabelUserInput}, "REST API body params"},
        {"get_json_params", core.SourceHTTPBody, []core.InputLabel{core.LabelHTTPBody, core.LabelUserInput}, "REST API JSON body"},
        {"get_body", core.SourceHTTPBody, []core.InputLabel{core.LabelHTTPBody, core.LabelUserInput}, "REST API raw body"},
        {"get_file_params", core.SourceHTTPFile, []core.InputLabel{core.LabelFile, core.LabelUserInput}, "REST API file uploads"},
        {"get_header", core.SourceHTTPHeader, []core.InputLabel{core.LabelHTTPHeader, core.LabelUserInput}, "REST API single header"},
        {"get_headers", core.SourceHTTPHeader, []core.InputLabel{core.LabelHTTPHeader, core.LabelUserInput}, "REST API all headers"},
    }

    for _, m := range wpRestMethods {
        registry.Register(&core.InputPattern{
            Name:        "wordpress_rest_" + m.method,
            Description: "WordPress " + m.desc,
            Category:    m.category,
            Labels:      m.labels,
            Language:    "php",
            Framework:   "wordpress",
            MethodName:  m.method,
            Regex:       regexp.MustCompile(`->` + m.method + `\s*\(`),
        })
    }

    // WordPress functions that receive user input
    wpInputFunctions := []struct {
        function string
        category core.SourceType
        labels   []core.InputLabel
        desc     string
    }{
        // Query variables
        {"get_query_var", core.SourceHTTPGet, []core.InputLabel{core.LabelHTTPGet, core.LabelUserInput}, "URL query variable"},
        {"get_search_query", core.SourceHTTPGet, []core.InputLabel{core.LabelHTTPGet, core.LabelUserInput}, "Search query from URL"},

        // Sanitization functions (first param is user input)
        {"sanitize_text_field", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user text input"},
        {"sanitize_textarea_field", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user textarea"},
        {"sanitize_email", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user email"},
        {"sanitize_file_name", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user filename"},
        {"sanitize_html_class", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user HTML class"},
        {"sanitize_key", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user key"},
        {"sanitize_title", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user title"},
        {"sanitize_user", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes username"},
        {"sanitize_url", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user URL"},

        // Escaping functions (first param is user input)
        {"esc_html", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes HTML from user"},
        {"esc_attr", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes attribute from user"},
        {"esc_url", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes URL from user"},
        {"esc_js", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes JS from user"},
        {"esc_textarea", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes textarea from user"},
        {"esc_sql", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes SQL from user"},

        // Type conversion (implies user input)
        {"absint", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Converts user input to absolute int"},
        {"intval", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Converts user input to int"},
        {"wp_unslash", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Unslashes user input"},

        // File uploads
        {"wp_handle_upload", core.SourceHTTPFile, []core.InputLabel{core.LabelFile, core.LabelUserInput}, "Handles file upload"},
        {"media_handle_upload", core.SourceHTTPFile, []core.InputLabel{core.LabelFile, core.LabelUserInput}, "Handles media upload"},
        {"wp_handle_sideload", core.SourceHTTPFile, []core.InputLabel{core.LabelFile, core.LabelUserInput}, "Handles sideload upload"},
    }

    for _, f := range wpInputFunctions {
        registry.Register(&core.InputPattern{
            Name:        "wordpress_" + f.function,
            Description: "WordPress " + f.desc,
            Category:    f.category,
            Labels:      f.labels,
            Language:    "php",
            Framework:   "wordpress",
            ExactMatch:  f.function,
            Regex:       regexp.MustCompile(`\b` + f.function + `\s*\(`),
            ParamIndex:  0, // First parameter receives input
        })
    }

    // WordPress non-input functions (explicitly excluded)
    wpNonInput := []string{
        "get_option",
        "get_site_option",
        "get_transient",
        "get_site_transient",
        "get_post_meta",
        "get_user_meta",
        "get_term_meta",
        "get_comment_meta",
        "get_metadata",
        "wp_cache_get",
        "get_post",
        "get_page",
        "get_user_by",
        "get_userdata",
        "get_term",
        "get_term_by",
        "get_category",
        "get_tag",
        "get_the_ID",
        "get_the_title",
        "get_the_content",
        "get_the_excerpt",
        "get_permalink",
        "get_bloginfo",
        "get_template_directory",
        "get_stylesheet_directory",
        "home_url",
        "admin_url",
        "site_url",
        "content_url",
        "plugins_url",
    }

    for _, f := range wpNonInput {
        registry.RegisterNonInput(f)
    }

    // Register AJAX action context patterns
    registry.Register(&core.InputPattern{
        Name:        "wordpress_ajax_action",
        Description: "WordPress AJAX action handler receives POST data",
        Category:    core.SourceHTTPPost,
        Labels:      []core.InputLabel{core.LabelHTTPPost, core.LabelUserInput},
        Language:    "php",
        Framework:   "wordpress",
        Regex:       regexp.MustCompile(`add_action\s*\(\s*['"]wp_ajax_(nopriv_)?`),
    })

    // Register admin POST action context
    registry.Register(&core.InputPattern{
        Name:        "wordpress_admin_post",
        Description: "WordPress admin POST action handler",
        Category:    core.SourceHTTPPost,
        Labels:      []core.InputLabel{core.LabelHTTPPost, core.LabelUserInput},
        Language:    "php",
        Framework:   "wordpress",
        Regex:       regexp.MustCompile(`add_action\s*\(\s*['"]admin_post_(nopriv_)?`),
    })
}

// WordPressIndicators returns file patterns that indicate WordPress
var WordPressIndicators = []string{
    "wp-config.php",
    "wp-content/",
    "wp-includes/",
    "wp-admin/",
    "wp-load.php",
    "wp-settings.php",
}

// WordPressFunctionIndicators returns functions that indicate WordPress
var WordPressFunctionIndicators = []string{
    "add_action",
    "add_filter",
    "do_action",
    "apply_filters",
    "register_activation_hook",
    "register_deactivation_hook",
    "wp_enqueue_script",
    "wp_enqueue_style",
}

// WordPressConstantIndicators returns constants that indicate WordPress
var WordPressConstantIndicators = []string{
    "ABSPATH",
    "WPINC",
    "WP_CONTENT_DIR",
    "WP_PLUGIN_DIR",
    "TEMPLATEPATH",
    "STYLESHEETPATH",
}

// IsWordPress checks if the given indicators suggest WordPress
func IsWordPress(files []string, functions []string, constants []string) bool {
    // Check for WordPress files
    for _, file := range files {
        for _, indicator := range WordPressIndicators {
            if file == indicator || containsPath(file, indicator) {
                return true
            }
        }
    }

    // Check for WordPress functions
    wpFuncCount := 0
    for _, fn := range functions {
        for _, indicator := range WordPressFunctionIndicators {
            if fn == indicator {
                wpFuncCount++
                if wpFuncCount >= 2 {
                    return true
                }
            }
        }
    }

    // Check for WordPress constants
    for _, c := range constants {
        for _, indicator := range WordPressConstantIndicators {
            if c == indicator {
                return true
            }
        }
    }

    return false
}

func containsPath(path, indicator string) bool {
    return len(path) >= len(indicator) &&
        (path == indicator ||
            (len(path) > len(indicator) &&
                (path[len(path)-len(indicator)-1] == '/' || path[len(path)-len(indicator)-1] == '\\')))
}
```

**Step 2: Build and verify**

```bash
go build ./pkg/sources/php/...
```

**Step 3: Commit**

```bash
git add pkg/sources/php/wordpress.go
git commit -m "feat(wordpress): add comprehensive WordPress input detection

- Add WP_REST_Request method patterns (get_param, get_body, etc.)
- Add sanitization functions as input sources
- Add escaping functions as input sources
- Add file upload handlers
- Add AJAX and admin POST action detection
- Explicitly exclude database/cache functions from input
- Add framework detection indicators

All patterns registered dynamically with core registry"
```

---

## Phase 4: Run Full Test Suite and Verify

### Task 4.1: Run Tests

**Step 1: Run all tests**

```bash
go test ./... -v
```

**Step 2: Run with race detector**

```bash
go test ./... -race
```

**Step 3: Build everything**

```bash
go build ./...
```

**Step 4: Commit any fixes needed**

```bash
git add -A
git commit -m "fix: resolve test failures from refactor"
```

---

## Summary

This plan creates:

1. **Core Infrastructure** (`pkg/sources/core/`)
   - `types.go` - Single source of truth for all types
   - `patterns.go` - Universal regex patterns
   - `registry.go` - Central pattern registry
   - `characteristics.go` - Dynamic input detection

2. **Bug Fixes**
   - Ruby session correctly categorized as non-user-input
   - Generic "request" mapping removed

3. **WordPress Support** (`pkg/sources/php/wordpress.go`)
   - WP_REST_Request methods
   - Sanitization/escaping functions
   - File upload handlers
   - AJAX/admin POST handlers
   - Non-input exclusions

---

**Plan complete and saved to `docs/plans/2026-01-25-dynamic-input-refactor-implementation.md`.**

Given the user's request to use the perfectionist-loop agent and their situation requiring autonomous execution, I will proceed with implementation using subagent-driven development.
