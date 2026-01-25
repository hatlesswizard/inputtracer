// Package common - regex_patterns.go provides pre-compiled regex patterns for input detection
// These patterns are used across semantic analyzers for consistent pattern matching
package common

import (
	"regexp"
	"sync"
)

// Pre-compiled PHP superglobal patterns
var (
	// SuperglobalPattern matches PHP superglobal array access
	// e.g., $_GET['key'], $_POST["key"], $_REQUEST[$var]
	SuperglobalPattern = regexp.MustCompile(`\$_(GET|POST|COOKIE|REQUEST|SERVER|FILES|SESSION|ENV)\[['"]?([\w\-]+)['"]?\]`)

	// SuperglobalForeachPattern matches PHP foreach over superglobals
	// e.g., foreach ($_GET as $key => $value)
	SuperglobalForeachPattern = regexp.MustCompile(`foreach\s*\(\s*(\$_\w+)\s+as\s+\$(\w+)\s*=>\s*\$(\w+)\s*\)`)

	// SuperglobalSimplePattern matches just the superglobal name
	// e.g., $_GET, $_POST (without array access)
	SuperglobalSimplePattern = regexp.MustCompile(`\$_(GET|POST|COOKIE|REQUEST|SERVER|FILES|SESSION|ENV)`)
)

// Pre-compiled method/property patterns
var (
	// MethodCallPattern matches object method calls
	// e.g., $obj->method(, $request->input(
	MethodCallPattern = regexp.MustCompile(`\$(\w+)->(\w+)\s*\(`)

	// PropertyArrayPattern matches property with array access
	// e.g., $obj->data['key'], $request->query["param"]
	PropertyArrayPattern = regexp.MustCompile(`\$(\w+)->(\w+)\[['"]?([\w\-]+)['"]?\]`)

	// SimplePropertyPattern matches simple property access (no array, no method call)
	// e.g., $obj->data, $request->body
	SimplePropertyPattern = regexp.MustCompile(`\$(\w+)->(\w+)(?:[^\[\(]|$)`)
)

// Pre-compiled input method patterns
var (
	// InputMethodPattern matches universal PHP input method patterns
	// e.g., ->get_input(, ->getInput(, ->input(
	InputMethodPattern = regexp.MustCompile(`(?i)->(?:get_?)?(?:input|var|variable|query_?params?|parsed_?body|cookie_?params?|server_?params?|uploaded_?files?|headers?|all|post|cookie|param)s?\s*\(`)

	// InputPropertyPattern matches universal PHP input property patterns
	// e.g., ->input[, ->data[, ->request[
	InputPropertyPattern = regexp.MustCompile(`(?i)->(?:input|request|params?|query|cookies?|headers?|body|data|args?|post|get|files?|server|attributes?|payload)s?\[`)

	// InputObjectPattern matches objects that typically carry user input
	// e.g., $request, $input, $mybb, ctx, context
	InputObjectPattern = regexp.MustCompile(`(?i)(request|input|req|params?|http|ctx|context|mybb|getRequest\(\)|getApplication\(\))`)
)

// Pre-compiled context-dependent method patterns (may or may not indicate input)
var (
	// ContextDependentMethodPattern matches methods that may be input getters
	// but need context to determine (e.g., ->get( could be cache get or input get)
	ContextDependentMethodPattern = regexp.MustCompile(`(?i)->(?:get_?)?(?:val|text|int|bool|array|raw_?val|check)\s*\(`)

	// ExcludeMethodPattern matches methods that look like input but aren't
	// e.g., ->getData( is often a generic getter, not input
	ExcludeMethodPattern = regexp.MustCompile(`(?i)->(?:getData|getBody|getContent|fetch|find|load|read)\s*\(`)
)

// Pre-compiled JavaScript/TypeScript patterns
var (
	// JSRequestPattern matches common JS request object access
	// e.g., req.body, req.query, req.params, request.body
	JSRequestPattern = regexp.MustCompile(`(?:req|request|ctx)\.(?:body|query|params|cookies|headers|files?)`)

	// JSPropertyAccessPattern matches JS property access that may be input
	// e.g., ctx.request.body, event.body
	JSPropertyAccessPattern = regexp.MustCompile(`\b(\w+)\.(\w+)(?:\.(\w+))?`)
)

// Pre-compiled Python patterns
var (
	// PythonFlaskPattern matches Flask request access
	// e.g., request.form, request.args, request.json
	PythonFlaskPattern = regexp.MustCompile(`request\.(?:form|args|json|data|values|files|cookies|headers)`)

	// PythonDjangoPattern matches Django request access
	// e.g., request.GET, request.POST, request.COOKIES
	PythonDjangoPattern = regexp.MustCompile(`request\.(?:GET|POST|COOKIES|META|FILES|body|data)`)

	// PythonArgparsePattern matches argparse parsed args access
	// e.g., args.username, args.verbose
	PythonArgparsePattern = regexp.MustCompile(`args\.(\w+)`)
)

// Pre-compiled Go patterns
var (
	// GoRequestPattern matches Go http.Request access
	// e.g., r.URL.Query(), r.FormValue(, r.Header.Get(
	GoRequestPattern = regexp.MustCompile(`(?:r|req|request)\.(?:URL\.Query|FormValue|Header|Cookie|Body|Form|PostForm)`)

	// GoGinPattern matches Gin framework access
	// e.g., c.Query(, c.Param(, c.PostForm(
	GoGinPattern = regexp.MustCompile(`c\.(?:Query|Param|PostForm|GetHeader|Cookie|ShouldBind)`)

	// GoEchoPattern matches Echo framework access
	// e.g., c.QueryParam(, c.Param(, c.FormValue(
	GoEchoPattern = regexp.MustCompile(`c\.(?:QueryParam|Param|FormValue|Request)`)
)

// Pre-compiled Java patterns
var (
	// JavaServletPattern matches HttpServletRequest access
	// e.g., request.getParameter(, request.getHeader(
	JavaServletPattern = regexp.MustCompile(`(?:request|req|httpRequest)\.get(?:Parameter|Header|Cookie|Attribute|Session)\s*\(`)

	// JavaSpringPattern matches Spring annotation parameters
	// e.g., @RequestParam, @PathVariable, @RequestBody
	JavaSpringAnnotationPattern = regexp.MustCompile(`@(?:RequestParam|PathVariable|RequestBody|RequestHeader|CookieValue|ModelAttribute|RequestPart|MatrixVariable)`)
)

// Pre-compiled C/C++ patterns
var (
	// CArgvPattern matches C main function argv access
	// e.g., argv[1], argv[i]
	CArgvPattern = regexp.MustCompile(`argv\s*\[\s*(\d+|\w+)\s*\]`)

	// CEnvPattern matches C environment access
	// e.g., getenv("PATH"), envp[0]
	CEnvPattern = regexp.MustCompile(`(?:getenv\s*\(|envp\s*\[|environ\s*\[)`)

	// CStdinPattern matches C stdin reads
	// e.g., scanf(, gets(, fgets(stdin
	CStdinPattern = regexp.MustCompile(`(?:scanf|gets|fgets|getchar|getc|getline)\s*\(`)
)

// Regex cache for dynamic pattern compilation
var (
	regexCache     = make(map[string]*regexp.Regexp)
	regexCacheMu   sync.RWMutex
)

// GetOrCompileRegex returns a cached or newly compiled regex
func GetOrCompileRegex(pattern string) (*regexp.Regexp, error) {
	regexCacheMu.RLock()
	if re, ok := regexCache[pattern]; ok {
		regexCacheMu.RUnlock()
		return re, nil
	}
	regexCacheMu.RUnlock()

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	regexCacheMu.Lock()
	regexCache[pattern] = re
	regexCacheMu.Unlock()

	return re, nil
}

// MustGetOrCompileRegex returns a cached or newly compiled regex, panicking on error
func MustGetOrCompileRegex(pattern string) *regexp.Regexp {
	re, err := GetOrCompileRegex(pattern)
	if err != nil {
		panic("failed to compile regex: " + err.Error())
	}
	return re
}

// InputPattern represents a compiled regex pattern for input detection
type InputPattern struct {
	Regex *regexp.Regexp
	Name  string // Human-readable name for matched input source
}

// PHPInputPatterns contains pre-compiled patterns for PHP input access
// Used for detecting input sources in expressions without carrier map
var PHPInputPatterns = []InputPattern{
	{Regex: regexp.MustCompile(`\$mybb->input\[`), Name: "$mybb->input"},
	{Regex: regexp.MustCompile(`\$mybb->cookies\[`), Name: "$mybb->cookies"},
	{Regex: regexp.MustCompile(`\$mybb->get_input\(`), Name: "$mybb->get_input()"},
	{Regex: regexp.MustCompile(`\$request->input\(`), Name: "$request->input()"},
	{Regex: regexp.MustCompile(`\$request->get\(`), Name: "$request->get()"},
	{Regex: regexp.MustCompile(`\$request->post\(`), Name: "$request->post()"},
	{Regex: regexp.MustCompile(`\$request->query\[`), Name: "$request->query"},
	{Regex: regexp.MustCompile(`\$_REQUEST\[`), Name: "$_REQUEST"},
}
