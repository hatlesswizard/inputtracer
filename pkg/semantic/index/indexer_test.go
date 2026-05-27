package index

import (
	"regexp"
	"testing"
)

func TestNewIndexer(t *testing.T) {
	idx := NewIndexer()
	if idx == nil {
		t.Fatal("expected non-nil Indexer")
	}

	stats := idx.Stats()
	if stats["total_symbols"] != 0 {
		t.Error("new indexer should have 0 symbols")
	}
}

func TestIndexer_AddSymbol(t *testing.T) {
	idx := NewIndexer()

	sym := &Symbol{
		Name:     "processInput",
		Type:     SymbolFunction,
		FilePath: "/app/handler.go",
		Line:     10,
		Language: "go",
	}

	idx.AddSymbol(sym)

	stats := idx.Stats()
	if stats["total_symbols"] != 1 {
		t.Errorf("expected 1 symbol, got %d", stats["total_symbols"])
	}

	// Verify retrieval
	retrieved := idx.GetSymbol(sym.ID)
	if retrieved == nil {
		t.Fatal("failed to retrieve added symbol")
	}
	if retrieved.Name != "processInput" {
		t.Errorf("expected name 'processInput', got '%s'", retrieved.Name)
	}
}

func TestIndexer_GetByName(t *testing.T) {
	idx := NewIndexer()

	// Add symbols with same name in different files
	idx.AddSymbol(&Symbol{
		Name:     "validate",
		Type:     SymbolFunction,
		FilePath: "/app/a.go",
		Language: "go",
	})
	idx.AddSymbol(&Symbol{
		Name:     "validate",
		Type:     SymbolFunction,
		FilePath: "/app/b.go",
		Language: "go",
	})
	idx.AddSymbol(&Symbol{
		Name:     "other",
		Type:     SymbolFunction,
		FilePath: "/app/c.go",
		Language: "go",
	})

	results := idx.GetByName("validate")
	if len(results) != 2 {
		t.Errorf("expected 2 symbols named 'validate', got %d", len(results))
	}
}

func TestIndexer_GetFunction(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{
		Name:     "handleRequest",
		Type:     SymbolFunction,
		FilePath: "/app/handler.go",
	})
	idx.AddSymbol(&Symbol{
		Name:     "User",
		Type:     SymbolClass,
		FilePath: "/app/user.go",
	})

	funcs := idx.GetFunction("handleRequest")
	if len(funcs) != 1 {
		t.Errorf("expected 1 function, got %d", len(funcs))
	}

	// Should not return class
	funcs = idx.GetFunction("User")
	if len(funcs) != 0 {
		t.Error("GetFunction should not return classes")
	}
}

func TestIndexer_GetClass(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{
		Name:     "UserService",
		Type:     SymbolClass,
		FilePath: "/app/user.go",
	})

	class := idx.GetClass("UserService")
	if class == nil {
		t.Fatal("expected to find class")
	}
	if class.Name != "UserService" {
		t.Errorf("expected class name 'UserService', got '%s'", class.Name)
	}
}

func TestIndexer_GetMethod(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{
		Name:      "save",
		Type:      SymbolMethod,
		ClassName: "UserRepository",
		FilePath:  "/app/user_repo.go",
	})

	method := idx.GetMethod("UserRepository", "save")
	if method == nil {
		t.Fatal("expected to find method")
	}
	if method.Name != "save" {
		t.Errorf("expected method name 'save', got '%s'", method.Name)
	}

	// Should not find method on wrong class
	method = idx.GetMethod("OtherClass", "save")
	if method != nil {
		t.Error("should not find method on wrong class")
	}
}

func TestIndexer_GetSymbolsInFile(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{Name: "func1", Type: SymbolFunction, FilePath: "/app/handler.go"})
	idx.AddSymbol(&Symbol{Name: "func2", Type: SymbolFunction, FilePath: "/app/handler.go"})
	idx.AddSymbol(&Symbol{Name: "other", Type: SymbolFunction, FilePath: "/app/other.go"})

	syms := idx.GetSymbolsInFile("/app/handler.go")
	if len(syms) != 2 {
		t.Errorf("expected 2 symbols in handler.go, got %d", len(syms))
	}
}

func TestIndexer_AddReference(t *testing.T) {
	idx := NewIndexer()

	sym := &Symbol{
		Name:     "processData",
		Type:     SymbolFunction,
		FilePath: "/app/processor.go",
	}
	idx.AddSymbol(sym)

	ref := Reference{
		FilePath: "/app/main.go",
		Line:     25,
		RefType:  "call",
	}
	idx.AddReference(sym.ID, ref)

	refs := idx.GetReferences(sym.ID)
	if len(refs) != 1 {
		t.Errorf("expected 1 reference, got %d", len(refs))
	}
	if refs[0].FilePath != "/app/main.go" {
		t.Error("reference file path mismatch")
	}
}

func TestIndexer_Search_ExactName(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{Name: "processInput", Type: SymbolFunction, FilePath: "/app/a.go"})
	idx.AddSymbol(&Symbol{Name: "processOutput", Type: SymbolFunction, FilePath: "/app/b.go"})
	idx.AddSymbol(&Symbol{Name: "validate", Type: SymbolFunction, FilePath: "/app/c.go"})

	results := idx.Search(&SearchQuery{Name: "processInput"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Symbol.Name != "processInput" {
		t.Error("expected exact name match")
	}
	if results[0].MatchedBy != "exact_name" {
		t.Errorf("expected MatchedBy='exact_name', got '%s'", results[0].MatchedBy)
	}
}

func TestIndexer_Search_Pattern(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{Name: "getUserById", Type: SymbolFunction, FilePath: "/app/a.go"})
	idx.AddSymbol(&Symbol{Name: "getUserByName", Type: SymbolFunction, FilePath: "/app/b.go"})
	idx.AddSymbol(&Symbol{Name: "createUser", Type: SymbolFunction, FilePath: "/app/c.go"})

	results := idx.Search(&SearchQuery{NamePattern: "^getUser.*"})
	if len(results) != 2 {
		t.Errorf("expected 2 results for pattern, got %d", len(results))
	}
}

func TestIndexer_Search_TypeFilter(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{Name: "User", Type: SymbolClass, FilePath: "/app/user.go"})
	idx.AddSymbol(&Symbol{Name: "User", Type: SymbolFunction, FilePath: "/app/utils.go"})

	results := idx.Search(&SearchQuery{Name: "User", Type: SymbolClass})
	if len(results) != 1 {
		t.Errorf("expected 1 class result, got %d", len(results))
	}
	if results[0].Symbol.Type != SymbolClass {
		t.Error("expected class type")
	}
}

func TestIndexer_Search_LanguageFilter(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{Name: "main", Type: SymbolFunction, FilePath: "/app/main.go", Language: "go"})
	idx.AddSymbol(&Symbol{Name: "main", Type: SymbolFunction, FilePath: "/app/main.php", Language: "php"})

	results := idx.Search(&SearchQuery{Name: "main", Language: "go"})
	if len(results) != 1 {
		t.Errorf("expected 1 go result, got %d", len(results))
	}
	if results[0].Symbol.Language != "go" {
		t.Error("expected go language")
	}
}

func TestIndexer_Search_ParameterMatch(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{
		Name:       "processRequest",
		Type:       SymbolFunction,
		FilePath:   "/app/handler.go",
		Parameters: []string{"req", "res"},
	})
	idx.AddSymbol(&Symbol{
		Name:       "handleData",
		Type:       SymbolFunction,
		FilePath:   "/app/data.go",
		Parameters: []string{"data"},
	})

	results := idx.Search(&SearchQuery{HasParameter: "req"})
	if len(results) != 1 {
		t.Errorf("expected 1 result with 'req' parameter, got %d", len(results))
	}
}

func TestIndexer_Search_Limit(t *testing.T) {
	idx := NewIndexer()

	for i := 0; i < 10; i++ {
		idx.AddSymbol(&Symbol{
			Name:     "func",
			Type:     SymbolFunction,
			FilePath: "/app/f" + string(rune('0'+i)) + ".go",
		})
	}

	results := idx.Search(&SearchQuery{Name: "func", Limit: 5})
	if len(results) != 5 {
		t.Errorf("expected 5 results with limit, got %d", len(results))
	}
}

func TestIndexer_Search_Caching(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{Name: "testFunc", Type: SymbolFunction, FilePath: "/app/test.go"})

	// First search
	results1 := idx.Search(&SearchQuery{Name: "testFunc"})

	// Second search (should be cached)
	results2 := idx.Search(&SearchQuery{Name: "testFunc"})

	if len(results1) != len(results2) {
		t.Error("cached results should match")
	}

	// Add new symbol and search again (cache should be invalidated)
	idx.AddSymbol(&Symbol{Name: "testFunc", Type: SymbolFunction, FilePath: "/app/test2.go"})
	results3 := idx.Search(&SearchQuery{Name: "testFunc"})

	if len(results3) != 2 {
		t.Errorf("expected 2 results after cache invalidation, got %d", len(results3))
	}
}

func TestIndexer_FindDefinition(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{Name: "validate", Type: SymbolFunction, FilePath: "/app/validator.go", Line: 10})
	idx.AddSymbol(&Symbol{Name: "validate", Type: SymbolFunction, FilePath: "/app/handler.go", Line: 20})

	// Should prefer definition in same file
	def := idx.FindDefinition("/app/handler.go", 50, "validate")
	if def == nil {
		t.Fatal("expected to find definition")
	}
	if def.FilePath != "/app/handler.go" {
		t.Error("should prefer definition in same file")
	}

	// Should find definition in other file if not in current
	def = idx.FindDefinition("/app/other.go", 50, "validate")
	if def == nil {
		t.Fatal("expected to find definition")
	}
}

func TestIndexer_GetAllFunctions(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{Name: "func1", Type: SymbolFunction, FilePath: "/a.go"})
	idx.AddSymbol(&Symbol{Name: "func2", Type: SymbolFunction, FilePath: "/b.go"})
	idx.AddSymbol(&Symbol{Name: "Class1", Type: SymbolClass, FilePath: "/c.go"})

	funcs := idx.GetAllFunctions()
	if len(funcs) != 2 {
		t.Errorf("expected 2 functions, got %d", len(funcs))
	}
}

func TestIndexer_GetAllClasses(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{Name: "UserService", Type: SymbolClass, FilePath: "/a.go"})
	idx.AddSymbol(&Symbol{Name: "DataService", Type: SymbolClass, FilePath: "/b.go"})
	idx.AddSymbol(&Symbol{Name: "helper", Type: SymbolFunction, FilePath: "/c.go"})

	classes := idx.GetAllClasses()
	if len(classes) != 2 {
		t.Errorf("expected 2 classes, got %d", len(classes))
	}
}

func TestIndexer_Clear(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{Name: "func1", Type: SymbolFunction, FilePath: "/a.go"})
	idx.AddSymbol(&Symbol{Name: "func2", Type: SymbolFunction, FilePath: "/b.go"})

	stats := idx.Stats()
	if stats["total_symbols"] != 2 {
		t.Errorf("expected 2 symbols before clear, got %d", stats["total_symbols"])
	}

	idx.Clear()

	stats = idx.Stats()
	if stats["total_symbols"] != 0 {
		t.Errorf("expected 0 symbols after clear, got %d", stats["total_symbols"])
	}
}

func TestIndexer_Stats(t *testing.T) {
	idx := NewIndexer()

	idx.AddSymbol(&Symbol{Name: "func1", Type: SymbolFunction, FilePath: "/a.go"})
	idx.AddSymbol(&Symbol{Name: "func2", Type: SymbolFunction, FilePath: "/b.go"})
	idx.AddSymbol(&Symbol{Name: "Class1", Type: SymbolClass, FilePath: "/c.go"})

	stats := idx.Stats()

	if stats["total_symbols"] != 3 {
		t.Errorf("expected 3 total symbols, got %d", stats["total_symbols"])
	}
	if stats["functions"] < 2 {
		t.Errorf("expected at least 2 functions, got %d", stats["functions"])
	}
	if stats["classes"] != 1 {
		t.Errorf("expected 1 class, got %d", stats["classes"])
	}

	t.Logf("Stats: %+v", stats)
}

func TestSymbolTypes(t *testing.T) {
	types := []SymbolType{
		SymbolFunction,
		SymbolMethod,
		SymbolClass,
		SymbolInterface,
		SymbolVariable,
		SymbolConstant,
		SymbolProperty,
		SymbolParameter,
	}

	for _, st := range types {
		if string(st) == "" {
			t.Error("symbol type should not be empty")
		}
	}
}

func TestIndexer_GlobToRegex(t *testing.T) {
	idx := NewIndexer()

	tests := []struct {
		glob     string
		input    string
		expected bool
	}{
		{"*.go", "main.go", true},
		{"*.go", "main.php", false},
		{"**/*.go", "pkg/foo/bar.go", true},
		{"pkg/**/*.go", "pkg/a/b/c.go", true},
		{"handler?.go", "handler1.go", true},
		{"handler?.go", "handlers.go", true}, // ? matches any single char including 's'
		{"handler?.go", "handler.go", false}, // ? must match exactly one char
	}

	for _, tc := range tests {
		regex := idx.globToRegex(tc.glob)
		matched, _ := regexp.MatchString(regex, tc.input)
		if matched != tc.expected {
			t.Errorf("glob '%s' vs '%s': expected %v, got %v (regex: %s)",
				tc.glob, tc.input, tc.expected, matched, regex)
		}
	}
}
