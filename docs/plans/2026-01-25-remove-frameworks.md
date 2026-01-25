# Remove PHP Frameworks Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove all PHP framework support except Laravel and Symfony from InputTracer

**Architecture:** Remove 12 framework entries from FrameworkDetectionPatterns in frameworks.go, remove 6 PHP indicators from PHPFrameworkIndicators in detection.go

**Tech Stack:** Go

---

### Task 1: Modify frameworks.go

**Files:**
- Modify: `pkg/sources/php/frameworks.go:148-219`

**Step 1: Edit FrameworkDetectionPatterns**

Replace the map to keep only laravel and symfony entries.

**Step 2: Verify build**

```bash
go build ./...
```

---

### Task 2: Modify detection.go

**Files:**
- Modify: `pkg/sources/frameworks/detection.go:21-70`

**Step 1: Edit PHPFrameworkIndicators**

Replace the slice to keep only laravel and symfony entries.

**Step 2: Verify tests**

```bash
go test ./...
```

---

### Task 3: Verification

```bash
go build ./...
go test ./...
```
