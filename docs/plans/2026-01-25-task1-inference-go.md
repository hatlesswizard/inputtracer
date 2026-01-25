# Task 1: Create inference.go Implementation Plan

> **For Claude:** This is a subagent task from the parent's Dynamic Pattern Generation plan.

**Goal:** Create `cmd/genpatterns/inference.go` with source type inference functions

**Architecture:** Three functions to dynamically infer source types from method names instead of static mappings

**Tech Stack:** Go, strings package

---

### Task 1: Create inference.go

**Files:**
- Create: `cmd/genpatterns/inference.go`

**Step 1: Create the file with InferSourceType, InferPopulatedFrom, InferDescription**

See parent task instructions for exact code.

**Step 2: Verify it compiles**

Run: `go build ./cmd/genpatterns/inference.go`
Expected: No errors (unused warnings OK)

**Step 3: Commit**

```bash
git add cmd/genpatterns/inference.go
git commit -m "feat(genpatterns): add source type inference from method names"
```
