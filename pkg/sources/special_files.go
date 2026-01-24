// Package sources - special_files.go provides centralized special file handling
// All special filename patterns should be defined here
package sources

import "strings"

// UnsupportedFilenames contains filenames that should not be parsed
// Replaces hardcoded switch in parser/service.go
var UnsupportedFilenames = map[string]bool{
	"makefile":      true,
	"gnumakefile":   true,
	"dockerfile":    true,
	"vagrantfile":   true,
	".gitignore":    true,
	".dockerignore": true,
	".npmignore":    true,
	".eslintignore": true,
	"license":       true,
	"licence":       true,
	"readme":        true,
	"changelog":     true,
	"contributing":  true,
}

// SpecialFilenameLanguages maps special filenames to their language
// Some special files have a known language even without extension
var SpecialFilenameLanguages = map[string]string{
	"makefile":    "makefile",
	"gnumakefile": "makefile",
	"dockerfile":  "dockerfile",
	"vagrantfile": "ruby",
	"gemfile":     "ruby",
	"rakefile":    "ruby",
	"guardfile":   "ruby",
	"podfile":     "ruby",
	"fastfile":    "ruby",
	"appfile":     "ruby",
	"dangerfile":  "ruby",
	"brewfile":    "ruby",
	"cakefile":    "coffeescript",
	"gruntfile":   "javascript",
	"gulpfile":    "javascript",
	"jakefile":    "javascript",
	"procfile":    "yaml",
	"jenkinsfile": "groovy",
}

// BinaryExtensions contains file extensions that are binary/non-parseable
var BinaryExtensions = map[string]bool{
	".exe":   true,
	".dll":   true,
	".so":    true,
	".dylib": true,
	".a":     true,
	".lib":   true,
	".obj":   true,
	".o":     true,
	".bin":   true,
	".dat":   true,
	".db":    true,
	".sqlite": true,
	".jpg":   true,
	".jpeg":  true,
	".png":   true,
	".gif":   true,
	".bmp":   true,
	".ico":   true,
	".svg":   true,
	".webp":  true,
	".pdf":   true,
	".doc":   true,
	".docx":  true,
	".xls":   true,
	".xlsx":  true,
	".zip":   true,
	".tar":   true,
	".gz":    true,
	".rar":   true,
	".7z":    true,
	".woff":  true,
	".woff2": true,
	".ttf":   true,
	".otf":   true,
	".eot":   true,
	".mp3":   true,
	".mp4":   true,
	".wav":   true,
	".avi":   true,
	".mov":   true,
}

// SkipPathPatterns contains path patterns that should be skipped
var SkipPathPatterns = []string{
	"/vendor/",
	"/node_modules/",
	"/.git/",
	"/cache/",
	"/__pycache__/",
	"/.venv/",
	"/venv/",
	"/target/",
	"/build/",
	"/dist/",
	"/.idea/",
	"/.vscode/",
}

// IsUnsupportedFilename checks if a filename should be skipped
func IsUnsupportedFilename(basename string) bool {
	return UnsupportedFilenames[strings.ToLower(basename)]
}

// GetSpecialFilenameLanguage returns the language for a special filename, or empty string
func GetSpecialFilenameLanguage(basename string) string {
	return SpecialFilenameLanguages[strings.ToLower(basename)]
}

// IsBinaryExtension checks if a file extension indicates a binary file
func IsBinaryExtension(ext string) bool {
	return BinaryExtensions[strings.ToLower(ext)]
}

// ShouldSkipPath checks if a path matches any skip pattern
func ShouldSkipPath(path string) bool {
	lowerPath := strings.ToLower(path)
	for _, pattern := range SkipPathPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}
	return false
}

// ShouldSkipDir checks if a directory name should be skipped
func ShouldSkipDir(dirName string) bool {
	lowerName := strings.ToLower(dirName)
	for _, d := range DefaultSkipDirs {
		if lowerName == d {
			return true
		}
	}
	return false
}
