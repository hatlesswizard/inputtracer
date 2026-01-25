# Remove Unsupported Framework Detection Patterns

> **For Claude:** Simple edit task - no sub-skill needed.

**Goal:** Remove unsupported PHP frameworks from FrameworkDetectionPatterns map, keeping only Laravel and Symfony.

**Architecture:** Direct map modification in frameworks.go

**Tech Stack:** Go

---

### Task 1: Edit FrameworkDetectionPatterns Map

**Files:**
- Modify: `pkg/sources/php/frameworks.go:148-219`

**Step 1: Replace the map**

Remove entries for: wordpress, drupal, yii, cakephp, zend, slim, lumen, phpbb, mediawiki, joomla, magento, prestashop

Keep only: laravel, symfony

**Step 2: Verify build**

Run: `go build ./...`
Expected: Build succeeds with no errors

---

**Status:** Ready for execution
