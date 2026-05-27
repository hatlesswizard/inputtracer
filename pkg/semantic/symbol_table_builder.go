package semantic

// symbol_table_builder.go — symbol-table construction responsibility extracted from tracer.go.
//
// symbolTableBuilder handles:
//   1. Merging per-file symbol tables into a single global symbol table
//      (buildGlobalSymbolTable)
//   2. Releasing body-source strings after the table is no longer needed
//      (releaseBodySources)
//   3. Releasing per-file symbol tables once the global one has been built
//      (releasePerFileSymbolTables)
//
// Like fileDiscoverer, this type reads from t.files / t.symbolTable (still
// owned by *Tracer) under t.mu, so the public API of *Tracer is undisturbed.

import "github.com/hatlesswizard/inputtracer/pkg/semantic/types"

// symbolTableBuilder merges per-file symbol tables and manages their lifecycle.
type symbolTableBuilder struct{}

// newSymbolTableBuilder constructs a symbolTableBuilder.
func newSymbolTableBuilder() *symbolTableBuilder {
	return &symbolTableBuilder{}
}

// buildGlobalSymbolTable merges all per-file symbol tables held in t.files into
// t.symbolTable.  The O(n×m) iteration runs lock-free on a snapshot; only the
// final map-swap holds the write lock.
func (b *symbolTableBuilder) buildGlobalSymbolTable(t *Tracer) {
	// Take a snapshot of the files map under a read lock.
	t.mu.RLock()
	type fileEntry struct {
		filePath string
		st       *types.SymbolTable
	}
	entries := make([]fileEntry, 0, len(t.files))
	for filePath, fileInfo := range t.files {
		if fileInfo.SymbolTable != nil {
			entries = append(entries, fileEntry{filePath, fileInfo.SymbolTable})
		}
	}
	t.mu.RUnlock()

	// Build merged tables locally without holding any lock.
	localClasses := make(map[string]*types.ClassDef)
	localFunctions := make(map[string]*types.FunctionDef)

	for _, e := range entries {
		for name, class := range e.st.Classes {
			key := e.filePath + "::" + name
			localClasses[key] = class
			if localClasses[name] == nil {
				localClasses[name] = class
			}
		}

		for name, fn := range e.st.Functions {
			key := e.filePath + "::" + name
			localFunctions[key] = fn
			if localFunctions[name] == nil {
				localFunctions[name] = fn
			}
		}
	}

	// Swap under write lock (brief critical section).
	t.mu.Lock()
	t.symbolTable.Classes = localClasses
	t.symbolTable.Functions = localFunctions
	t.mu.Unlock()
}

// releaseBodySources releases all body-source strings from per-file and global
// symbol tables to reclaim memory once they are no longer needed for analysis.
func (b *symbolTableBuilder) releaseBodySources(t *Tracer) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, fileInfo := range t.files {
		if fileInfo.SymbolTable != nil {
			fileInfo.SymbolTable.ReleaseBodySources()
		}
	}

	for _, class := range t.symbolTable.Classes {
		class.ReleaseBodySources()
	}
	for _, fn := range t.symbolTable.Functions {
		fn.BodySource = ""
	}
}

// releasePerFileSymbolTables releases per-file symbol tables after the global
// symbol table has been built, freeing memory before flow analysis begins.
func (b *symbolTableBuilder) releasePerFileSymbolTables(t *Tracer) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, fileInfo := range t.files {
		if fileInfo.SymbolTable != nil {
			fileInfo.SymbolTable.ReleaseBodySources()
			fileInfo.SymbolTable.Variables = nil
			fileInfo.SymbolTable.Constants = nil
			fileInfo.SymbolTable.Imports = nil
			fileInfo.SymbolTable.Classes = nil
			fileInfo.SymbolTable.Functions = nil
		}
	}
}
