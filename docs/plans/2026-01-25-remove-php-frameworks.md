# Remove PHP Frameworks Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove all PHP framework support except Laravel and Symfony from InputTracer

**Architecture:** Delete wordpress.go file (already done), remove 12 framework entries from FrameworkDetectionPatterns map, remove 6 PHP indicators from PHPFrameworkIndicators array

**Tech Stack:** Go

---

### Task 1: Delete wordpress.go (COMPLETED)

**Files:**
- Delete: `pkg/sources/php/wordpress.go`

**Status:** Already deleted (visible in git status)

---

### Task 2: Modify frameworks.go

**Files:**
- Modify: `pkg/sources/php/frameworks.go:148-219`

**Step 1: Remove 12 framework entries from FrameworkDetectionPatterns**

Keep only `laravel` and `symfony`. Remove: wordpress, drupal, yii, cakephp, zend, slim, lumen, phpbb, mediawiki, joomla, magento, prestashop

**Step 2: Run build**

```bash
go build ./...
```

Expected: Build succeeds

---

### Task 3: Modify detection.go

**Files:**
- Modify: `pkg/sources/frameworks/detection.go:21-70`

**Step 1: Remove 6 PHP framework indicators**

Keep only `laravel` and `symfony`. Remove: wordpress, drupal, yii2, cakephp, phpbb, prestashop

**Step 2: Run tests**

```bash
go test ./...
```

Expected: All tests pass

---

### Task 4: Final Verification

**Step 1: Full verification**

```bash
go build ./...
go test ./...
```

**Step 2: Grep for removed frameworks**

```bash
grep -r "wordpress\|drupal\|yii\|cakephp\|zend\|slim\|lumen\|phpbb\|mediawiki\|joomla\|magento\|prestashop" pkg/
```

Expected: No matches in source code

---

## Impact

- ~900 lines removed
- 1 file deleted, 2 files modified
