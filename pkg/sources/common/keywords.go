package common

// UniversalKeywords contains keywords to filter from variable extraction.
// These are common across most languages and should not be treated as variable names.
var UniversalKeywords = map[string]bool{
	// Control flow
	"if": true, "else": true, "elseif": true, "elif": true,
	"switch": true, "case": true, "default": true,
	"for": true, "while": true, "do": true, "foreach": true,
	"break": true, "continue": true, "return": true,
	"try": true, "catch": true, "finally": true, "throw": true,

	// Boolean/null literals
	"true": true, "false": true, "null": true, "nil": true, "undefined": true,
	"True": true, "False": true, "None": true, // Python

	// Logical operators (word form)
	"and": true, "or": true, "not": true, "is": true, "in": true,

	// Declaration keywords
	"function": true, "func": true, "def": true, "fn": true,
	"var": true, "let": true, "const": true, "mut": true,
	"class": true, "interface": true, "struct": true, "enum": true,
	"type": true, "impl": true, "trait": true,

	// Modifiers
	"public": true, "private": true, "protected": true,
	"static": true, "final": true, "abstract": true,
	"async": true, "await": true, "yield": true,

	// Object-oriented
	"new": true, "this": true, "self": true, "super": true,
	"instanceof": true, "typeof": true,

	// PHP specific
	"isset": true, "empty": true, "unset": true,
	"echo": true, "print": true,

	// Import/export
	"import": true, "export": true, "require": true, "include": true,
	"use": true, "namespace": true, "package": true, "module": true,
}

// LanguageKeywords provides language-specific keyword lists
var LanguageKeywords = map[string][]string{
	"php": {
		"isset", "empty", "unset", "echo", "print",
		"is_null", "is_string", "is_array", "is_int", "is_bool", "is_float",
		"array", "list", "global", "static",
	},
	"python": {
		"None", "True", "False", "and", "or", "not", "is", "in",
		"lambda", "pass", "with", "as", "global", "nonlocal",
		"assert", "raise", "except", "from",
	},
	"javascript": {
		"undefined", "null", "true", "false", "typeof", "instanceof",
		"arguments", "debugger", "delete", "void", "with",
		"NaN", "Infinity",
	},
	"typescript": {
		"undefined", "null", "true", "false", "typeof", "instanceof",
		"arguments", "debugger", "delete", "void", "with",
		"NaN", "Infinity", "keyof", "readonly", "infer", "never",
	},
	"go": {
		"nil", "true", "false", "iota",
		"make", "new", "append", "copy", "delete", "len", "cap",
		"panic", "recover", "defer", "go", "select", "chan",
		"range", "fallthrough", "goto", "map", "struct",
	},
	"java": {
		"null", "true", "false", "instanceof",
		"extends", "implements", "throws", "native", "synchronized",
		"volatile", "transient", "strictfp", "assert",
	},
	"c": {
		"NULL", "sizeof", "typedef", "extern", "register", "volatile",
		"inline", "restrict", "auto", "goto",
	},
	"cpp": {
		"NULL", "nullptr", "sizeof", "typedef", "extern", "register",
		"volatile", "inline", "restrict", "auto", "goto",
		"template", "typename", "virtual", "override", "delete",
		"constexpr", "noexcept", "decltype", "static_cast", "dynamic_cast",
	},
	"c_sharp": {
		"null", "true", "false", "typeof", "sizeof", "nameof",
		"checked", "unchecked", "lock", "fixed", "stackalloc",
		"internal", "sealed", "virtual", "override", "readonly",
		"ref", "out", "in", "params", "where", "default",
	},
	"ruby": {
		"nil", "true", "false", "self", "super",
		"defined?", "begin", "end", "rescue", "ensure", "raise",
		"alias", "undef", "redo", "retry", "yield",
	},
	"rust": {
		"true", "false", "self", "Self", "super", "crate",
		"move", "ref", "box", "dyn", "where", "unsafe",
		"extern", "mod", "pub", "priv", "loop", "match",
	},
}

// IsKeyword checks if a word is a universal keyword
func IsKeyword(word string) bool {
	return UniversalKeywords[word]
}

// IsKeywordForLanguage checks if a word is a keyword for a specific language
func IsKeywordForLanguage(word, lang string) bool {
	if UniversalKeywords[word] {
		return true
	}
	if keywords, ok := LanguageKeywords[lang]; ok {
		for _, kw := range keywords {
			if word == kw {
				return true
			}
		}
	}
	return false
}
