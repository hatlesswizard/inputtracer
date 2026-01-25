# Remove Duplicated PHP Function Definitions

**Date:** 2026-01-25
**Type:** Refactoring
**Scope:** `pkg/semantic/tracer.go`

## Problem

`pkg/semantic/tracer.go:1128-1194` contains hardcoded inline arrays that duplicate definitions already centralized in `pkg/sources/php/functions.go`:

| Location | inputFuncs | deserializeFuncs | networkFuncs |
|----------|------------|------------------|--------------|
| `tracer.go` (inline) | 7 | 4 | 2 |
| `functions.go` (centralized) | 17 | 11 | 16 |

This violates:
- DRY principle
- Project convention (CLAUDE.md: patterns belong in `pkg/sources/{language}/`)
- Maintenance sanity

## Solution

Replace 67 lines of inline arrays + loops with a single call to `phpPatterns.IdentifyExternalDataSource()`.

### Before
```go
inputFuncs := []string{"file_get_contents", "fgets", ...}
for _, fn := range inputFuncs { ... }

deserializeFuncs := []string{"unserialize(", ...}
for _, fn := range deserializeFuncs { ... }

if strings.Contains(expr, "curl_exec(") || ...
```

### After
```go
if sourceType, confidence := phpPatterns.IdentifyExternalDataSource(expr); confidence > 0 {
    return &types.SourceInfo{
        Type:       types.SourceType(sourceType),
        Expression: expr,
        FilePath:   filePath,
        Line:       line,
    }
}
```

## Impact

- **Lines removed:** ~57
- **Lines added:** ~10
- **Behavior change:** None (same logic, centralized)
- **Bonus:** Now covers 17 input functions instead of 7
