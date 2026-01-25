# InputTracer Dynamic Refactoring Design

**Date:** 2026-01-25
**Status:** Design Complete
**Scope:** Major refactoring to make InputTracer fully dynamic

---

## 1. Executive Summary

This design addresses three critical problems in InputTracer:

1. **Static Definitions Scattered**: 14+ files contain hardcoded input patterns outside `pkg/sources/`
2. **Massive Duplication**: Same patterns defined 3+ times (InputMethodPattern, superglobals, etc.)
3. **Incorrect Categorization**: Ruby session data mislabeled as user input; overly broad mappings

**Goal**: Create a fully dynamic, centralized input detection system where:
- ALL input definitions live in `pkg/sources/` only
- ZERO duplication - single source of truth for each pattern
- Dynamic detection based on characteristics, not static lists
- WordPress support integrated dynamically

---

## 2. Current Problems (Detailed)

### 2.1 Static Definitions Outside Sources

| File | Problem |
|------|---------|
| `pkg/semantic/tracer/vartracer.go:579` | Hardcoded "input", "cookies", "query", "request" |
| `pkg/semantic/analyzer/java/analyzer.go:603-612` | Hardcoded object name patterns |
| `pkg/semantic/analyzer/python/analyzer.go:794-817` | Hardcoded framework names |
| `pkg/semantic/analyzer/javascript/analyzer.go:960-994` | Hardcoded framework detection |
| `pkg/semantic/analyzer/typescript/analyzer.go:579-598` | Duplicate framework detection |
| `pkg/semantic/analyzer/golang/analyzer.go:580-599` | Hardcoded Go framework detection |
| `pkg/semantic/analyzer/csharp/analyzer.go:607-614` | Hardcoded C# framework detection |
| `pkg/semantic/analyzer/c/analyzer.go:301-308` | Hardcoded argv, envp, stdin |
| `pkg/semantic/analyzer/cpp/analyzer.go:490-497` | Duplicate C patterns |

### 2.2 Duplications Within Sources

| Pattern | Files |
|---------|-------|
| `InputLabel` type | `common/types.go`, `constants/input_labels.go` |
| `InputMethodPattern` | `common/regex_patterns.go`, `php/patterns.go`, `javascript/frameworks.go` |
| `InputPropertyPattern` | Same 3 files |
| `InputObjectPattern` | Same 3 files |
| `SuperglobalPattern` | `common/regex_patterns.go`, `superglobals.go`, `php/patterns.go` |
| `ArrayKeyAccessPattern` | `php/patterns.go`, `javascript/patterns.go`, `python/patterns.go` |
| `ThisPropertyAssignPattern` | PHP, JS, Python versions |

### 2.3 Categorization Bugs

1. **Ruby session = UserInput** (CRITICAL): `pkg/sources/mappings.go:406`
2. **Generic "request" too broad**: `pkg/sources/mappings.go:405`

---

## 3. Architecture Design

### 3.1 New Directory Structure

```
pkg/sources/
├── core/                          # NEW: Single source of truth
│   ├── types.go                   # ALL type definitions (InputLabel, SourceType, etc.)
│   ├── registry.go                # Dynamic pattern registry
│   ├── patterns.go                # Universal regex patterns (compiled once)
│   ├── input_characteristics.go   # Dynamic input detection characteristics
│   └── categories.go              # Input category definitions & rules
│
├── detection/                     # NEW: Dynamic detection engine
│   ├── detector.go                # Main detection interface
│   ├── heuristics.go              # Characteristic-based detection
│   ├── framework_detector.go      # Dynamic framework detection
│   └── property_analyzer.go       # Property/method analysis
│
├── php/
│   ├── registrations.go           # Register PHP patterns with core registry
│   ├── superglobals.go            # PHP superglobal definitions (consolidated)
│   ├── laravel.go                 # Laravel patterns
│   ├── symfony.go                 # Symfony patterns
│   └── wordpress.go               # NEW: WordPress patterns (dynamic)
│
├── javascript/
│   ├── registrations.go           # Register JS patterns with core registry
│   ├── express.go
│   ├── koa.go
│   ├── fastify.go
│   └── nestjs.go
│
├── python/
│   ├── registrations.go
│   ├── django.go
│   ├── flask.go
│   └── fastapi.go
│
├── [other languages...]
│
└── DELETED FILES:
    - common/types.go              # MERGED into core/types.go
    - common/source_types.go       # MERGED into core/types.go
    - common/regex_patterns.go     # MERGED into core/patterns.go
    - constants/                   # ENTIRE DIRECTORY DELETED
    - defaults.go                  # MERGED into core/
    - labels.go                    # MERGED into core/types.go
    - types.go                     # MERGED into core/types.go
    - mappings.go                  # REPLACED by core/registry.go
    - superglobals.go              # MOVED to php/superglobals.go
    - input_methods.go             # MERGED into core/
    - ast_patterns.go              # MERGED into core/
    - graph_styles.go              # MERGED into core/
    - special_files.go             # MERGED into detection/
```

### 3.2 Core Registry Design

```go
// pkg/sources/core/registry.go

// InputCharacteristic defines what makes something an input source
type InputCharacteristic struct {
    // Method/property name patterns (dynamic matching)
    NamePatterns     []string          // e.g., "get*", "input", "*param*"

    // Object context patterns
    ObjectPatterns   []string          // e.g., "request", "req", "*context*"

    // Return type indicators
    ReturnsUserData  bool              // Does this return user-controlled data?

    // Category assignment
    Category         SourceType        // HTTP_GET, HTTP_POST, FILE, etc.
    Labels           []InputLabel      // Additional labels

    // Language specificity (empty = all languages)
    Languages        []string

    // Framework specificity (empty = all frameworks)
    Frameworks       []string

    // Confidence (for heuristic matching)
    BaseConfidence   float64           // 0.0 - 1.0
}

// Registry holds all input detection patterns
type Registry struct {
    // Exact matches (fast lookup)
    exactPatterns    map[string]*InputCharacteristic

    // Glob/regex patterns (slower but flexible)
    dynamicPatterns  []*CompiledCharacteristic

    // Language-specific registrations
    languagePatterns map[string][]*InputCharacteristic

    // Framework-specific registrations
    frameworkPatterns map[string][]*InputCharacteristic

    // Compiled universal patterns (for sharing)
    universalPatterns *UniversalPatterns
}

// Register adds a pattern to the registry
func (r *Registry) Register(characteristic *InputCharacteristic)

// Match checks if an expression matches any registered pattern
func (r *Registry) Match(expr Expression, lang string, framework string) *MatchResult

// MatchDynamic uses heuristics for unknown patterns
func (r *Registry) MatchDynamic(expr Expression, context *AnalysisContext) *MatchResult
```

### 3.3 Dynamic Input Detection

Instead of static lists, we detect input based on characteristics:

```go
// pkg/sources/core/input_characteristics.go

// UserInputCharacteristics - What makes something user input?
var UserInputCharacteristics = []Characteristic{
    // HTTP Request Data
    {
        Description: "HTTP query parameters",
        Indicators: []string{
            "query", "params", "queryParams", "searchParams",
            "getQueryParam", "getQuery", "query_string",
        },
        Category: SourceHTTPGet,
        Labels: []InputLabel{LabelHTTPGet, LabelUserInput},
    },
    {
        Description: "HTTP body/POST data",
        Indicators: []string{
            "body", "postData", "formData", "requestBody",
            "getBody", "getParsedBody", "input",
        },
        Category: SourceHTTPPost,
        Labels: []InputLabel{LabelHTTPPost, LabelUserInput},
    },
    {
        Description: "HTTP cookies",
        Indicators: []string{
            "cookies", "cookie", "getCookie", "cookieParams",
        },
        Category: SourceHTTPCookie,
        Labels: []InputLabel{LabelHTTPCookie, LabelUserInput},
    },
    {
        Description: "HTTP headers",
        Indicators: []string{
            "headers", "header", "getHeader", "getHeaders",
        },
        Category: SourceHTTPHeader,
        Labels: []InputLabel{LabelHTTPHeader, LabelUserInput},
    },
    // ... more characteristics
}

// NON-UserInputCharacteristics - What is NOT user input?
var NonUserInputCharacteristics = []Characteristic{
    {
        Description: "Session data (server-side)",
        Indicators: []string{
            "session", "getSession", "sessionData",
        },
        Category: SourceSession,
        Labels: []InputLabel{}, // NO LabelUserInput!
        Reason: "Session data is stored server-side, not sent by client",
    },
    {
        Description: "Database results",
        Indicators: []string{
            "fetch", "fetchAll", "query", "findOne", "findAll",
        },
        Category: SourceDatabase,
        Labels: []InputLabel{LabelDatabase}, // NOT user input
        Reason: "Database results may contain user data but are not direct user input",
    },
    {
        Description: "File contents",
        Indicators: []string{
            "readFile", "file_get_contents", "fread", "fgets",
        },
        Category: SourceFile,
        Labels: []InputLabel{LabelFile}, // NOT user input
        Reason: "File data is server-side storage, not user-controlled",
    },
    {
        Description: "Environment variables",
        Indicators: []string{
            "env", "getenv", "environ", "process.env",
        },
        Category: SourceEnvVar,
        Labels: []InputLabel{LabelEnvironment}, // NOT user input
        Reason: "Environment variables are server configuration",
    },
}
```

### 3.4 WordPress Dynamic Detection

```go
// pkg/sources/php/wordpress.go

func init() {
    // Register WordPress patterns dynamically
    core.RegisterFramework("wordpress", &FrameworkDefinition{
        Language: "php",

        // Detection indicators
        Indicators: []FrameworkIndicator{
            {Type: "file", Pattern: "wp-config.php"},
            {Type: "file", Pattern: "wp-content/"},
            {Type: "file", Pattern: "wp-includes/"},
            {Type: "function", Pattern: "add_action"},
            {Type: "function", Pattern: "add_filter"},
            {Type: "constant", Pattern: "ABSPATH"},
            {Type: "constant", Pattern: "WPINC"},
        },

        // Input sources - registered dynamically
        InputSources: []InputSourceDef{
            // Direct superglobal wrappers
            {
                Pattern:     `\$_GET\s*\[`,
                Category:    SourceHTTPGet,
                Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
                Description: "WordPress uses raw superglobals",
            },
            {
                Pattern:     `\$_POST\s*\[`,
                Category:    SourceHTTPPost,
                Labels:      []InputLabel{LabelHTTPPost, LabelUserInput},
            },
            {
                Pattern:     `\$_REQUEST\s*\[`,
                Category:    SourceHTTPRequest,
                Labels:      []InputLabel{LabelUserInput},
            },

            // WordPress-specific input functions
            {
                // Sanitization functions that RECEIVE user input
                FunctionPattern: `sanitize_text_field\s*\(`,
                ParamIndex:      0, // First parameter is the input
                Category:        SourceUserInput,
                Labels:          []InputLabel{LabelUserInput},
                Description:     "Sanitization functions receive user input",
            },
            {
                FunctionPattern: `wp_unslash\s*\(`,
                ParamIndex:      0,
                Category:        SourceUserInput,
                Labels:          []InputLabel{LabelUserInput},
            },
            {
                FunctionPattern: `absint\s*\(`,
                ParamIndex:      0,
                Category:        SourceUserInput,
                Labels:          []InputLabel{LabelUserInput},
            },

            // AJAX handlers receive user input
            {
                ContextPattern:  `add_action\s*\(\s*['"]wp_ajax_`,
                Category:        SourceHTTPPost,
                Labels:          []InputLabel{LabelHTTPPost, LabelUserInput},
                Description:     "AJAX handlers receive POST data",
            },
            {
                ContextPattern:  `add_action\s*\(\s*['"]wp_ajax_nopriv_`,
                Category:        SourceHTTPPost,
                Labels:          []InputLabel{LabelHTTPPost, LabelUserInput},
            },

            // REST API handlers
            {
                MethodPattern:   `->get_param\s*\(`,
                ObjectPattern:   `WP_REST_Request`,
                Category:        SourceHTTPRequest,
                Labels:          []InputLabel{LabelUserInput},
                Description:     "REST API request parameters",
            },
            {
                MethodPattern:   `->get_params\s*\(`,
                ObjectPattern:   `WP_REST_Request`,
                Category:        SourceHTTPRequest,
                Labels:          []InputLabel{LabelUserInput},
            },
            {
                MethodPattern:   `->get_body\s*\(`,
                ObjectPattern:   `WP_REST_Request`,
                Category:        SourceHTTPBody,
                Labels:          []InputLabel{LabelHTTPBody, LabelUserInput},
            },
            {
                MethodPattern:   `->get_json_params\s*\(`,
                ObjectPattern:   `WP_REST_Request`,
                Category:        SourceHTTPBody,
                Labels:          []InputLabel{LabelHTTPBody, LabelUserInput},
            },
            {
                MethodPattern:   `->get_query_params\s*\(`,
                ObjectPattern:   `WP_REST_Request`,
                Category:        SourceHTTPGet,
                Labels:          []InputLabel{LabelHTTPGet, LabelUserInput},
            },
            {
                MethodPattern:   `->get_body_params\s*\(`,
                ObjectPattern:   `WP_REST_Request`,
                Category:        SourceHTTPPost,
                Labels:          []InputLabel{LabelHTTPPost, LabelUserInput},
            },
            {
                MethodPattern:   `->get_file_params\s*\(`,
                ObjectPattern:   `WP_REST_Request`,
                Category:        SourceHTTPFile,
                Labels:          []InputLabel{LabelFile, LabelUserInput},
            },
            {
                MethodPattern:   `->get_header\s*\(`,
                ObjectPattern:   `WP_REST_Request`,
                Category:        SourceHTTPHeader,
                Labels:          []InputLabel{LabelHTTPHeader, LabelUserInput},
            },
            {
                MethodPattern:   `->get_headers\s*\(`,
                ObjectPattern:   `WP_REST_Request`,
                Category:        SourceHTTPHeader,
                Labels:          []InputLabel{LabelHTTPHeader, LabelUserInput},
            },

            // Form handling
            {
                FunctionPattern: `check_admin_referer\s*\(`,
                Category:        SourceHTTPPost,
                Labels:          []InputLabel{LabelHTTPPost},
                Description:     "Nonce verification implies form submission",
            },
            {
                FunctionPattern: `wp_verify_nonce\s*\(`,
                Category:        SourceHTTPPost,
                Labels:          []InputLabel{LabelHTTPPost},
            },

            // URL parameters
            {
                FunctionPattern: `get_query_var\s*\(`,
                Category:        SourceHTTPGet,
                Labels:          []InputLabel{LabelHTTPGet, LabelUserInput},
                Description:     "WordPress query variables from URL",
            },
            {
                FunctionPattern: `get_search_query\s*\(`,
                Category:        SourceHTTPGet,
                Labels:          []InputLabel{LabelHTTPGet, LabelUserInput},
            },

            // File uploads
            {
                FunctionPattern: `wp_handle_upload\s*\(`,
                ParamIndex:      0,
                Category:        SourceHTTPFile,
                Labels:          []InputLabel{LabelFile, LabelUserInput},
            },
            {
                FunctionPattern: `media_handle_upload\s*\(`,
                Category:        SourceHTTPFile,
                Labels:          []InputLabel{LabelFile, LabelUserInput},
            },
        },

        // NON-input sources (explicitly excluded)
        NonInputSources: []string{
            "get_option",        // Database storage, not user input
            "get_transient",     // Cache, not user input
            "get_post_meta",     // Database, not user input
            "get_user_meta",     // Database, not user input
            "get_term_meta",     // Database, not user input
            "get_site_option",   // Database, not user input
            "wp_cache_get",      // Cache, not user input
        },
    })
}
```

---

## 4. Implementation Plan

### Phase 1: Create Core Infrastructure (Clean Foundation)

1. Create `pkg/sources/core/` directory
2. Create `core/types.go` - consolidate ALL type definitions
3. Create `core/registry.go` - central pattern registry
4. Create `core/patterns.go` - compiled universal patterns
5. Create `core/input_characteristics.go` - dynamic detection rules
6. Create `core/categories.go` - category definitions with clear rules

### Phase 2: Migrate and Deduplicate

1. Move all patterns from `common/` to `core/`
2. Delete `pkg/sources/constants/` entirely
3. Delete duplicate definitions in `common/`
4. Update `pkg/sources/labels.go` to re-export from `core/`
5. Update `pkg/sources/types.go` to re-export from `core/`
6. Consolidate `superglobals.go` into `php/superglobals.go`

### Phase 3: Update Language Registrations

1. Create `{language}/registrations.go` for each language
2. Move framework patterns to use core registry
3. Remove hardcoded patterns from analyzers
4. Update analyzers to query core registry

### Phase 4: Fix Categorization Bugs

1. Fix Ruby session mislabeling in mappings
2. Fix generic "request" mapping
3. Audit all SourceType assignments for correctness
4. Add tests for correct categorization

### Phase 5: Implement WordPress Support

1. Create `pkg/sources/php/wordpress.go`
2. Register all WordPress input patterns dynamically
3. Add WordPress framework detection
4. Add comprehensive tests

### Phase 6: Update Analyzers

1. Remove all hardcoded patterns from `pkg/semantic/analyzer/`
2. Update analyzers to use core registry
3. Remove framework detection from analyzers (use core)
4. Test all languages

### Phase 7: Cleanup and Verification

1. Delete all deprecated files
2. Run full test suite
3. Verify no duplications remain
4. Verify all inputs correctly categorized
5. Performance testing

---

## 5. Files to Delete

```
pkg/sources/common/types.go           # Merged to core/types.go
pkg/sources/common/source_types.go    # Merged to core/types.go
pkg/sources/common/regex_patterns.go  # Merged to core/patterns.go
pkg/sources/constants/                # Entire directory
pkg/sources/defaults.go               # Merged to core/
pkg/sources/mappings.go               # Replaced by core/registry.go
pkg/sources/superglobals.go           # Moved to php/superglobals.go
```

---

## 6. Success Criteria

1. **Zero static patterns outside `pkg/sources/`**
   - All analyzers query the registry
   - No hardcoded strings in semantic package

2. **Zero duplication**
   - Each pattern defined exactly once
   - Single source of truth for types

3. **Correct categorization**
   - Session data NEVER labeled as user input
   - Database results labeled as database, not user input
   - File reads labeled as file, not user input
   - Only HTTP request data labeled as user input

4. **Dynamic detection**
   - Can detect input sources by characteristics
   - Works for unknown frameworks via heuristics
   - WordPress fully supported

5. **All tests pass**
   - Existing tests continue to pass
   - New tests for categorization
   - New tests for WordPress

---

## 7. Risk Mitigation

1. **Backward Compatibility**: Breaking changes acceptable per user
2. **Performance**: Registry uses maps for O(1) lookup
3. **Completeness**: Extensive testing after each phase
4. **Rollback**: Git tags before each phase

---

## 8. Estimated Scope

- Files to modify: ~30
- Files to delete: ~10
- Files to create: ~15
- Lines of code: ~3000 new, ~2000 deleted

---

*Design complete. Ready for implementation with perfectionist-loop agent.*
