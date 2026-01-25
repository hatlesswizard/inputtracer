package sources

import (
	"sync"

	"github.com/hatlesswizard/inputtracer/pkg/sources/c"
	"github.com/hatlesswizard/inputtracer/pkg/sources/cpp"
	"github.com/hatlesswizard/inputtracer/pkg/sources/csharp"
	"github.com/hatlesswizard/inputtracer/pkg/sources/golang"
	"github.com/hatlesswizard/inputtracer/pkg/sources/java"
	"github.com/hatlesswizard/inputtracer/pkg/sources/javascript"
	"github.com/hatlesswizard/inputtracer/pkg/sources/php"
	"github.com/hatlesswizard/inputtracer/pkg/sources/python"
	"github.com/hatlesswizard/inputtracer/pkg/sources/ruby"
	"github.com/hatlesswizard/inputtracer/pkg/sources/rust"
)

// Registry manages all source matchers
type Registry struct {
	matchers map[string]Matcher
	sources  map[string][]Definition
	mu       sync.RWMutex
}

// NewRegistry creates a new source registry
func NewRegistry() *Registry {
	return &Registry{
		matchers: make(map[string]Matcher),
		sources:  make(map[string][]Definition),
	}
}

// RegisterMatcher registers a language-specific matcher
func (r *Registry) RegisterMatcher(matcher Matcher) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matchers[matcher.Language()] = matcher
}

// AddSource adds a source definition
func (r *Registry) AddSource(def Definition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sources[def.Language] = append(r.sources[def.Language], def)
}

// GetMatcher returns the matcher for a language
func (r *Registry) GetMatcher(language string) Matcher {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.matchers[language]
}

// GetSources returns all source definitions for a language
func (r *Registry) GetSources(language string) []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sources[language]
}

// RegisterAll registers all language matchers with the registry
func RegisterAll(r *Registry) {
	// Register PHP
	r.RegisterMatcher(php.NewMatcher())

	// Register JavaScript
	r.RegisterMatcher(javascript.NewMatcher())

	// Register TypeScript (uses JS patterns with different language)
	r.RegisterMatcher(javascript.NewTypeScriptMatcher())

	// Register Python
	r.RegisterMatcher(python.NewMatcher())

	// Register Go
	r.RegisterMatcher(golang.NewMatcher())

	// Register Java
	r.RegisterMatcher(java.NewMatcher())

	// Register C
	r.RegisterMatcher(c.NewMatcher())

	// Register C++
	r.RegisterMatcher(cpp.NewMatcher())

	// Register C#
	r.RegisterMatcher(csharp.NewMatcher())

	// Register Ruby
	r.RegisterMatcher(ruby.NewMatcher())

	// Register Rust
	r.RegisterMatcher(rust.NewMatcher())
}
