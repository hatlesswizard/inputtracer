# Remove Duplicated PHP Function Definitions - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace hardcoded inline PHP function arrays in `pkg/semantic/tracer.go` with centralized definitions from `pkg/sources/php/functions.go`.

**Architecture:** Single file edit - replace ~57 lines of inline arrays with a ~10 line call to `phpPatterns.IdentifyExternalDataSource()`. No new files, no new dependencies (import already exists).

**Tech Stack:** Go, existing `pkg/sources/php` package

---

## Task 1: Replace Inline Arrays with Centralized Function

**Files:**
- Modify: `pkg/semantic/tracer.go:1127-1194`

**Step 1: Run tests to establish baseline**

```bash
go test ./pkg/semantic/... -v -count=1
```

Expected: All tests pass (establishes baseline before refactoring)

**Step 2: Replace the duplicated code**

In `pkg/semantic/tracer.go`, find the `identifySource` function and replace lines 1127-1194:

**REMOVE this code (lines 1127-1194):**
```go
	// Check for input functions
	inputFuncs := []string{
		"file_get_contents", "fgets", "fread", "fgetc",
		"getenv", "getallheaders", "apache_request_headers",
	}
	for _, fn := range inputFuncs {
		if strings.Contains(expr, fn+"(") {
			return &types.SourceInfo{
				Type:       types.SourceFile,
				Expression: expr,
				FilePath:   filePath,
				Line:       line,
			}
		}
	}

	// =====================================================
	// UNIVERSAL PHP FRAMEWORK PATTERNS (from pkg/sources/php)
	// These detect user input across ALL PHP frameworks
	// =====================================================

	// Check property array access patterns using centralized patterns
	if phpPatterns.IsInputPropertyAccess(expr) {
		return &types.SourceInfo{
			Type:       types.SourceUserInput,
			Expression: expr,
			FilePath:   filePath,
			Line:       line,
		}
	}

	// Check method call patterns using centralized patterns
	if phpPatterns.IsInputMethodCall(expr) {
		return &types.SourceInfo{
			Type:       types.SourceUserInput,
			Expression: expr,
			FilePath:   filePath,
			Line:       line,
		}
	}

	// Deserialization functions (receive potentially tainted data)
	deserializeFuncs := []string{
		"unserialize(",
		"json_decode(",
		"simplexml_load_string(",
		"yaml_parse(",
	}
	for _, fn := range deserializeFuncs {
		if strings.Contains(expr, fn) {
			return &types.SourceInfo{
				Type:       types.SourceUserInput,
				Expression: expr,
				FilePath:   filePath,
				Line:       line,
			}
		}
	}

	// cURL responses (external data)
	if strings.Contains(expr, "curl_exec(") || strings.Contains(expr, "curl_multi_getcontent(") {
		return &types.SourceInfo{
			Type:       types.SourceType("network"),
			Expression: expr,
			FilePath:   filePath,
			Line:       line,
		}
	}
```

**REPLACE with:**
```go
	// Check for input/deserialization/network functions using centralized definitions
	// from pkg/sources/php/functions.go (replaces hardcoded inline arrays)
	if sourceType, confidence := phpPatterns.IdentifyExternalDataSource(expr); confidence > 0 {
		return &types.SourceInfo{
			Type:       types.SourceType(sourceType),
			Expression: expr,
			FilePath:   filePath,
			Line:       line,
		}
	}

	// Check property array access patterns using centralized patterns
	if phpPatterns.IsInputPropertyAccess(expr) {
		return &types.SourceInfo{
			Type:       types.SourceUserInput,
			Expression: expr,
			FilePath:   filePath,
			Line:       line,
		}
	}

	// Check method call patterns using centralized patterns
	if phpPatterns.IsInputMethodCall(expr) {
		return &types.SourceInfo{
			Type:       types.SourceUserInput,
			Expression: expr,
			FilePath:   filePath,
			Line:       line,
		}
	}
```

**Step 3: Verify it compiles**

```bash
go build ./pkg/semantic/...
```

Expected: No errors

**Step 4: Run tests to verify behavior unchanged**

```bash
go test ./pkg/semantic/... -v -count=1
```

Expected: All tests pass (same as baseline)

**Step 5: Run full test suite**

```bash
go test ./... -count=1
```

Expected: All tests pass

**Step 6: Commit**

```bash
git add pkg/semantic/tracer.go
git commit -m "refactor(semantic): use centralized PHP function definitions

Replace hardcoded inline arrays in identifySource() with
phpPatterns.IdentifyExternalDataSource() from pkg/sources/php/functions.go.

- Removes ~57 lines of duplicated code
- Now covers 17 input functions (was 7)
- Now covers 11 deserialization functions (was 4)
- Now covers 16 network functions (was 2)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Summary

| Metric | Before | After |
|--------|--------|-------|
| Lines in identifySource | ~67 | ~30 |
| Input functions covered | 7 | 17 |
| Deserialization functions | 4 | 11 |
| Network functions | 2 | 16 |
| Single source of truth | No | Yes |
