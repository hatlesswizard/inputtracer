package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// filterInputDefinitions returns definitions for PHP's filter_input() and
// filter_input_array() built-in functions. These are the canonical, security-
// recommended way to read and validate GET/POST/COOKIE/SERVER input in PHP >= 5.2.
func filterInputDefinitions() []common.Definition {
	keyEx := func(inputType string) string {
		return `\bfilter_input(?:_array)?\s*\(\s*` + inputType + `\s*,\s*['"]([^'"]+)['"]`
	}
	return []common.Definition{
		{
			Name:         "filter_input(INPUT_GET)",
			Pattern:      `\bfilter_input\s*\(\s*INPUT_GET\b`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "PHP filter_input(INPUT_GET) — validated read of a query-string parameter",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: keyEx("INPUT_GET"),
		},
		{
			Name:         "filter_input(INPUT_POST)",
			Pattern:      `\bfilter_input\s*\(\s*INPUT_POST\b`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "PHP filter_input(INPUT_POST) — validated read of a POST body parameter",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: keyEx("INPUT_POST"),
		},
		{
			Name:         "filter_input(INPUT_COOKIE)",
			Pattern:      `\bfilter_input\s*\(\s*INPUT_COOKIE\b`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:  "PHP filter_input(INPUT_COOKIE) — validated read of a cookie",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: keyEx("INPUT_COOKIE"),
		},
		{
			Name:         "filter_input(INPUT_SERVER)",
			Pattern:      `\bfilter_input\s*\(\s*INPUT_SERVER\b`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "PHP filter_input(INPUT_SERVER) — validated read of a server/header variable",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: keyEx("INPUT_SERVER"),
		},
		{
			Name:         "filter_input(INPUT_ENV)",
			Pattern:      `\bfilter_input\s*\(\s*INPUT_ENV\b`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "PHP filter_input(INPUT_ENV) — validated read of an environment variable",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: keyEx("INPUT_ENV"),
		},
		// filter_input_array variants
		{
			Name:        "filter_input_array(INPUT_GET)",
			Pattern:     `\bfilter_input_array\s*\(\s*INPUT_GET\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "PHP filter_input_array(INPUT_GET) — validated read of all GET parameters",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "filter_input_array(INPUT_POST)",
			Pattern:     `\bfilter_input_array\s*\(\s*INPUT_POST\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "PHP filter_input_array(INPUT_POST) — validated read of all POST parameters",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "filter_input_array(INPUT_COOKIE)",
			Pattern:     `\bfilter_input_array\s*\(\s*INPUT_COOKIE\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description: "PHP filter_input_array(INPUT_COOKIE) — validated read of all cookies",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "filter_input_array(INPUT_SERVER)",
			Pattern:     `\bfilter_input_array\s*\(\s*INPUT_SERVER\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "PHP filter_input_array(INPUT_SERVER) — validated read of all server vars",
			NodeTypes:   []string{"function_call_expression"},
		},
	}
}

// phpBuiltinWrapperDefinitions returns definitions for PHP built-in functions
// whose first argument is a superglobal. Applying these functions does NOT
// remove the user-controlled origin of the data — the source is still the HTTP
// request (or cookie / server variable). Groups are ordered by frequency in
// real-world PHP code.
func phpBuiltinWrapperDefinitions() []common.Definition {
	// Shared key extractor: pull the array key out of $_XYZ['key'] or $_XYZ[$var].
	kx := `\[\s*['"]?([^'"\]]+)['"]?\s*\]`

	// sg matches any superglobal as the first argument of a function call.
	sg := `\$_(GET|POST|REQUEST|COOKIE|SERVER|FILES)`

	return []common.Definition{
		// ── String whitespace / cleanup ───────────────────────────────────────
		{
			Name:         "trim/ltrim/rtrim($_*)",
			Pattern:      `\b(?:trim|ltrim|rtrim|chop)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Whitespace trimming applied to superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── HTML encoding / output escaping ───────────────────────────────────
		{
			Name:         "htmlspecialchars/htmlentities($_*)",
			Pattern:      `\b(?:htmlspecialchars|htmlentities|htmlspecialchars_decode|html_entity_decode)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "HTML encoding applied to superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		{
			Name:         "strip_tags/nl2br($_*)",
			Pattern:      `\b(?:strip_tags|nl2br|wordwrap|chunk_split)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "HTML/text processing applied to superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── Slash handling ────────────────────────────────────────────────────
		{
			Name:         "addslashes/stripslashes($_*)",
			Pattern:      `\b(?:addslashes|stripslashes|addcslashes|stripcslashes)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Slash escaping applied to superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── Case / string transformation ──────────────────────────────────────
		{
			Name:         "strtolower/strtoupper/mb_str*($_*)",
			Pattern:      `\b(?:strtolower|strtoupper|ucfirst|lcfirst|ucwords|mb_strtolower|mb_strtoupper|mb_convert_case)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Case conversion applied to superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── URL encoding / decoding ───────────────────────────────────────────
		{
			Name:         "urldecode/rawurldecode($_*)",
			Pattern:      `\b(?:urldecode|rawurldecode|urlencode|rawurlencode)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "URL encoding applied to superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── Binary / base64 encoding ─────────────────────────────────────────
		{
			Name:         "base64_decode/base64_encode($_*)",
			Pattern:      `\b(?:base64_decode|base64_encode|hex2bin|bin2hex|quoted_printable_decode|quoted_printable_encode)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Binary/base64 encoding applied to superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── JSON / serialization ─────────────────────────────────────────────
		{
			Name:         "json_decode/json_encode($_*)",
			Pattern:      `\b(?:json_decode|json_encode)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "JSON encoding applied to superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── Numeric type-casting / math ──────────────────────────────────────
		{
			Name:         "intval/floatval/abs($_*)",
			Pattern:      `\b(?:intval|floatval|doubleval|boolval|abs|ceil|floor|round|number_format|fmod)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Numeric cast applied to superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── String inspection ────────────────────────────────────────────────
		{
			Name:         "strlen/mb_strlen/substr($_*)",
			Pattern:      `\b(?:strlen|mb_strlen|substr|mb_substr|str_word_count|str_split)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "String operation on superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── Hashing ───────────────────────────────────────────────────────────
		{
			Name:         "md5/sha1/hash($_*)",
			Pattern:      `\b(?:md5|sha1|crc32|hash|password_hash)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Hash function applied to superglobal — traces origin of hashed user data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── preg / string replacement ────────────────────────────────────────
		{
			Name:    "preg_replace/str_replace on $_*",
			Pattern: `\b(?:preg_replace|preg_match|preg_split|str_replace|str_ireplace|str_pad)\s*\(\s*(?:[^,]+,\s*){0,2}\s*` + sg,
			Language: "php",
			Labels:   []common.InputLabel{common.LabelUserInput},
			Description:  "Regex/string replacement applied to superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── sprintf / printf with superglobals ───────────────────────────────
		{
			Name:    "sprintf/printf with $_*",
			Pattern: `\b(?:sprintf|printf|vsprintf|vprintf|fprintf|sscanf)\s*\(\s*[^,]+,\s*` + sg,
			Language: "php",
			Labels:   []common.InputLabel{common.LabelUserInput},
			Description:  "String formatting with superglobal argument — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── PHP 8.x string functions ─────────────────────────────────────────
		{
			Name:         "str_contains/str_starts_with/str_ends_with($_*)",
			Pattern:      `\b(?:str_contains|str_starts_with|str_ends_with)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "PHP 8 str_contains/str_starts_with/str_ends_with on superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── Encoding / character conversion ──────────────────────────────────
		{
			Name:         "mb_convert_encoding/iconv($_*)",
			Pattern:      `\b(?:mb_convert_encoding|iconv|utf8_encode|utf8_decode|mb_detect_encoding)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Character encoding conversion on superglobal — source is still user input",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── Serialization / deserialization ──────────────────────────────────
		{
			Name:         "unserialize($_*)",
			Pattern:      `\bunserialize\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "unserialize() on superglobal — deserializes user-controlled data (potential object injection)",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		{
			Name:         "serialize($_*)",
			Pattern:      `\bserialize\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "serialize() on superglobal — serialized form of user-controlled data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// ── Sorting — source remains user-controlled after sort ───────────────
		{
			Name:         "sort/usort/ksort($_*)",
			Pattern:      `\b(?:sort|usort|arsort|asort|ksort|krsort|rsort|uasort|uksort|array_multisort)\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Array sort on superglobal — sorted array still contains user-controlled data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
	}
}

// phpExistenceCheckDefinitions returns definitions for PHP existence/membership
// checks that wrap superglobals. These are extremely common — isset() and empty()
// are used on virtually every superglobal access in well-written PHP code.
func phpExistenceCheckDefinitions() []common.Definition {
	sg := `\$_(GET|POST|REQUEST|COOKIE|SERVER|FILES|ENV|SESSION)`
	kx := `\[\s*['"]?([^'"\]]+)['"]?\s*\]`

	return []common.Definition{
		// isset($_POST['key']), isset($_GET['field']), etc.
		{
			Name:         "isset($_*[...])",
			Pattern:      `\bisset\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "isset() check on superglobal — confirms user-input key exists before use",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// empty($_GET['field']), empty($_POST['name']), etc.
		{
			Name:         "empty($_*[...])",
			Pattern:      `\bempty\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "empty() check on superglobal — tests if user-input value is falsy",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_key_exists('key', $_POST)
		{
			Name:         "array_key_exists('key', $_*)",
			Pattern:      `\barray_key_exists\s*\(\s*[^,]+,\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_key_exists() on superglobal — checks if HTTP parameter is present",
			NodeTypes:    []string{"function_call_expression"},
		},
		// in_array($_POST['role'], $allowedValues)
		{
			Name:         "in_array($_*[...], ...)",
			Pattern:      `\bin_array\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "in_array() with superglobal as needle — validates user-input value membership",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_search($_POST['x'], $arr)
		{
			Name:         "array_search($_*[...], ...)",
			Pattern:      `\barray_search\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_search() with superglobal needle — looks up user input in an array",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
	}
}

// legacyInputDefinitions returns definitions for legacy PHP input mechanisms:
// the pre-PHP-5.6 $HTTP_RAW_POST_DATA global, old $HTTP_*_VARS globals, and
// parse_str() / mb_parse_str() applied to user-controlled query strings.
func legacyInputDefinitions() []common.Definition {
	return []common.Definition{
		// Pre-PHP-4.1 register_globals-era superglobal aliases.
		{
			Name:        "$HTTP_GET_VARS",
			Pattern:     `\$HTTP_GET_VARS\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "Legacy PHP global $HTTP_GET_VARS — GET params (deprecated since PHP 4.1, removed PHP 5.4+)",
			NodeTypes:   []string{"variable_name"},
		},
		{
			Name:        "$HTTP_POST_VARS",
			Pattern:     `\$HTTP_POST_VARS\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "Legacy PHP global $HTTP_POST_VARS — POST params (deprecated since PHP 4.1)",
			NodeTypes:   []string{"variable_name"},
		},
		{
			Name:        "$HTTP_COOKIE_VARS",
			Pattern:     `\$HTTP_COOKIE_VARS\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description: "Legacy PHP global $HTTP_COOKIE_VARS — cookie values (deprecated since PHP 4.1)",
			NodeTypes:   []string{"variable_name"},
		},
		{
			Name:        "$HTTP_SERVER_VARS",
			Pattern:     `\$HTTP_SERVER_VARS\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Legacy PHP global $HTTP_SERVER_VARS — server vars (deprecated since PHP 4.1)",
			NodeTypes:   []string{"variable_name"},
		},
		{
			Name:        "$HTTP_ENV_VARS",
			Pattern:     `\$HTTP_ENV_VARS\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelEnvironment},
			Description: "Legacy PHP global $HTTP_ENV_VARS — env vars (deprecated since PHP 4.1)",
			NodeTypes:   []string{"variable_name"},
		},
		// Raw POST body — pre-PHP-5.6.
		{
			Name:        "$HTTP_RAW_POST_DATA",
			Pattern:     `\$HTTP_RAW_POST_DATA\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Legacy PHP global $HTTP_RAW_POST_DATA — raw POST body (deprecated since PHP 5.6)",
			NodeTypes:   []string{"variable_name"},
		},
		// parse_str on user-controlled query string
		{
			Name:    "parse_str($_SERVER[QUERY_STRING])",
			Pattern: `\bparse_str\s*\(\s*\$_(?:SERVER|GET|REQUEST)`,
			Language: "php",
			Labels:   []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "parse_str() on user-controlled query string — populates variables from HTTP input",
			NodeTypes:    []string{"function_call_expression"},
		},
		{
			Name:    "mb_parse_str($_SERVER[QUERY_STRING])",
			Pattern: `\bmb_parse_str\s*\(\s*\$_(?:SERVER|GET|REQUEST)`,
			Language: "php",
			Labels:   []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "mb_parse_str() on user-controlled query string — multibyte variant of parse_str",
			NodeTypes:    []string{"function_call_expression"},
		},
	}
}

// phpStreamInputDefinitions returns definitions for PHP stream-based input:
// raw POST body via php://input, stdin via STDIN constant or php://stdin, and
// stream_get_contents(STDIN) and similar constructs used in CLI/CGI PHP.
func phpStreamInputDefinitions() []common.Definition {
	return []common.Definition{
		// php://input — raw POST body (does not require multipart form data).
		// Covered in matcher.go streamDefinitions() but repeated here for completeness.
		// The one in matcher.go takes precedence (earlier in the definition list).
		// STDIN constant forms.
		{
			Name:        "fgets(STDIN)",
			Pattern:     `\bfgets\s*\(\s*STDIN\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "fgets(STDIN) — read line from standard input (CLI PHP, CGI)",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "fread(STDIN, ...)",
			Pattern:     `\bfread\s*\(\s*STDIN\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "fread(STDIN, ...) — binary read from standard input (CLI PHP)",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "stream_get_contents(STDIN)",
			Pattern:     `\bstream_get_contents\s*\(\s*STDIN\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "stream_get_contents(STDIN) — read all of standard input at once",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "file_get_contents(php://stdin)",
			Pattern:     `\bfile_get_contents\s*\(\s*['"]php://stdin['"]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "file_get_contents('php://stdin') — read all of standard input as string",
			NodeTypes:   []string{"function_call_expression"},
		},
		// readline() — interactive CLI input (often in shell scripts / REPLs).
		{
			Name:        "readline()",
			Pattern:     `\breadline\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "readline() — interactive input from terminal (CLI PHP)",
			NodeTypes:   []string{"function_call_expression"},
		},
	}
}

// phpCliInputDefinitions returns definitions for CLI argument access:
// $argv (argument vector), $argc (argument count), and getopt().
func phpCliInputDefinitions() []common.Definition {
	return []common.Definition{
		// $argv[N] — individual command-line argument.
		{
			Name:         "$argv[N]",
			Pattern:      `\$argv\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelCLI},
			Description:  "CLI argument access: $argv[N] — individual command-line argument by position",
			NodeTypes:    []string{"subscript_expression", "variable_name"},
			KeyExtractor: `\$argv\s*\[\s*(\d+)\s*\]`,
		},
		// $argc — argument count (indirectly confirms number of CLI args).
		{
			Name:        "$argc",
			Pattern:     `\$argc\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "CLI argument count: $argc — number of command-line arguments",
			NodeTypes:   []string{"variable_name"},
		},
		// getopt() — POSIX-style option parsing.
		{
			Name:        "getopt()",
			Pattern:     `\bgetopt\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "getopt() — parse command-line options (POSIX-style short/long options)",
			NodeTypes:   []string{"function_call_expression"},
		},
	}
}

// phpCastInputDefinitions returns definitions for PHP type-cast expressions
// applied directly to superglobals. Type casting does NOT sanitize or remove the
// user-controlled origin — the source is still the HTTP parameter.
// Patterns like (int)$_GET['id'] or (string)$_POST['name'] are very common.
func phpCastInputDefinitions() []common.Definition {
	sg := `\$_(GET|POST|REQUEST|COOKIE|SERVER|FILES|ENV)`
	kx := `\[\s*['"]?([^'"\]]+)['"]?\s*\]`

	return []common.Definition{
		// Integer cast: (int)$_GET['id']
		{
			Name:         "(int)$_*",
			Pattern:      `\(\s*int\s*\)\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "(int) cast on superglobal — numeric coercion of user input",
			NodeTypes:    []string{"cast_expression", "parenthesized_expression"},
			KeyExtractor: kx,
		},
		// Integer cast variant: (integer)$_POST['amount']
		{
			Name:         "(integer)$_*",
			Pattern:      `\(\s*integer\s*\)\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "(integer) cast on superglobal — numeric coercion of user input",
			NodeTypes:    []string{"cast_expression", "parenthesized_expression"},
			KeyExtractor: kx,
		},
		// Float cast: (float)$_GET['price']
		{
			Name:         "(float)$_*",
			Pattern:      `\(\s*(?:float|double|real)\s*\)\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "(float) cast on superglobal — floating-point coercion of user input",
			NodeTypes:    []string{"cast_expression", "parenthesized_expression"},
			KeyExtractor: kx,
		},
		// String cast: (string)$_POST['name']
		{
			Name:         "(string)$_*",
			Pattern:      `\(\s*string\s*\)\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "(string) cast on superglobal — string coercion of user input",
			NodeTypes:    []string{"cast_expression", "parenthesized_expression"},
			KeyExtractor: kx,
		},
		// Bool cast: (bool)$_GET['flag']
		{
			Name:         "(bool)$_*",
			Pattern:      `\(\s*(?:bool|boolean)\s*\)\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "(bool) cast on superglobal — boolean coercion of user input",
			NodeTypes:    []string{"cast_expression", "parenthesized_expression"},
			KeyExtractor: kx,
		},
		// Array cast: (array)$_POST['items']
		{
			Name:         "(array)$_*",
			Pattern:      `\(\s*array\s*\)\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "(array) cast on superglobal — array coercion of user input",
			NodeTypes:    []string{"cast_expression", "parenthesized_expression"},
			KeyExtractor: kx,
		},
	}
}

// phpArrayOperationDefinitions returns definitions for PHP array and string
// manipulation functions when a superglobal (or its value) is the primary operand.
// The user-controlled origin propagates through these operations.
func phpArrayOperationDefinitions() []common.Definition {
	sg := `\$_(GET|POST|REQUEST|COOKIE|SERVER|FILES)`
	kx := `\[\s*['"]?([^'"\]]+)['"]?\s*\]`

	return []common.Definition{
		// array_map($fn, $_POST) — applies function to every element of superglobal array.
		{
			Name:         "array_map($fn, $_*)",
			Pattern:      `\barray_map\s*\(\s*[^,]+,\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_map() over superglobal array — user input propagates through mapped values",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_filter($_GET, $fn) — filters a superglobal array.
		{
			Name:         "array_filter($_*)",
			Pattern:      `\barray_filter\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_filter() on superglobal — filtered result still contains user-controlled data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_values($_POST) — reindexes superglobal array.
		{
			Name:         "array_values($_*)",
			Pattern:      `\barray_values\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_values() on superglobal — reindexed result still contains user data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_keys($_POST) — extracts keys from superglobal (keys can be user-controlled).
		{
			Name:         "array_keys($_*)",
			Pattern:      `\barray_keys\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_keys() on superglobal — array keys submitted by user are input sources",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_merge($_POST, ...) / array_merge(..., $_POST) — merges superglobal.
		{
			Name:         "array_merge($_*,...)",
			Pattern:      `\barray_merge\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_merge() starting with superglobal — merged result contains user data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// implode(',', $_GET) — joins superglobal array elements.
		// The separator argument may be a quoted string like ',', so [^,]+ won't work —
		// use a pattern that matches a quoted string or unquoted token as separator.
		{
			Name:         "implode($sep, $_*)",
			Pattern:      `\b(?:implode|join)\s*\(\s*(?:'[^']*'|"[^"]*"|[^,'"]+)\s*,\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "implode()/join() over superglobal array — joined string still user-controlled",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// explode(',', $_GET['list']) — splits superglobal string into array.
		{
			Name:         "explode($sep, $_*[...])",
			Pattern:      `\bexplode\s*\(\s*(?:'[^']*'|"[^"]*"|[^,'"]+)\s*,\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "explode() on superglobal value — resulting array elements are user-controlled",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// compact() with variables derived from superglobals — less direct but noteworthy.
		{
			Name:        "compact() with superglobal-derived vars",
			Pattern:     `\bcompact\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "compact() — may bundle superglobal-derived variables into an array",
			NodeTypes:   []string{"function_call_expression"},
		},
		// array_shift($_POST) — removes and returns first element of superglobal.
		{
			Name:         "array_shift($_*)",
			Pattern:      `\barray_shift\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_shift() on superglobal — shifted element is user-controlled",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_slice($_POST, 0, 5) — extracts a slice of superglobal.
		{
			Name:         "array_slice($_*)",
			Pattern:      `\barray_slice\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_slice() on superglobal — sliced elements are user-controlled",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_intersect_key($_POST, $allowed) — filters superglobal by allowed keys.
		{
			Name:         "array_intersect_key($_*, ...)",
			Pattern:      `\barray_intersect_key\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_intersect_key() on superglobal — filtered result still contains user data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_intersect($_POST['ids'], $valid_ids) — filters superglobal by allowed values.
		{
			Name:         "array_intersect($_*[...], ...)",
			Pattern:      `\barray_intersect\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_intersect() on superglobal — intersected result still contains user data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_diff($_POST['tags'], $banned) — diffs superglobal with another array.
		{
			Name:         "array_diff($_*, ...)",
			Pattern:      `\barray_diff\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_diff() on superglobal — differenced result still contains user data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_flip($_GET) — swaps keys and values in superglobal.
		{
			Name:         "array_flip($_*)",
			Pattern:      `\barray_flip\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_flip() on superglobal — flipped result still contains user data as keys/values",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_unique($_POST['choices']) — removes duplicates from superglobal array.
		{
			Name:         "array_unique($_*[...])",
			Pattern:      `\barray_unique\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_unique() on superglobal — deduplicated result still contains user data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
		// array_column($_POST, 'key') — extracts column from superglobal array.
		{
			Name:         "array_column($_*, ...)",
			Pattern:      `\barray_column\s*\(\s*` + sg,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "array_column() on superglobal — extracted column still contains user data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: kx,
		},
	}
}

// phpDynamicInputDefinitions returns definitions for PHP dynamic variable
// access patterns that can introduce user-controlled values into program state
// in unexpected ways.
func phpDynamicInputDefinitions() []common.Definition {
	return []common.Definition{
		// Variable variables: $$_GET['varname'] — dangerously dynamic variable lookup.
		{
			Name:        "$$_GET[...] (variable variables)",
			Pattern:     `\$\$_(GET|POST|REQUEST|COOKIE)\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "Variable variable $$_*['key'] — user input controls which variable is accessed",
			NodeTypes:   []string{"variable_name", "subscript_expression"},
		},
		// extract($_POST) / extract($_GET) — injects superglobal keys as local variables.
		// NOTE: extract() patterns are also in matcher.go superglobalDefinitions().
		// Those cover the basic forms; this definition additionally covers extract()
		// called with EXTR_PREFIX_ALL or other flags on any superglobal.
		{
			Name:        "extract($_*) with flags",
			Pattern:     `\bextract\s*\(\s*\$_(GET|POST|REQUEST|COOKIE|SERVER|FILES)\s*,`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelHTTPGet, common.LabelUserInput},
			Description: "extract() with flags on superglobal — user input injected as local variables",
			NodeTypes:   []string{"function_call_expression"},
		},
		// list() = $_POST['data'] — PHP 5 list() assignment from superglobal.
		{
			Name:        "list(...) = $_*[...]",
			Pattern:     `\blist\s*\([^)]*\)\s*=\s*\$_(GET|POST|REQUEST|COOKIE)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "list() destructuring from superglobal — user input distributed to multiple variables",
			NodeTypes:   []string{"list_expression", "assignment_expression"},
		},
		// [$a, $b] = $_GET['items'] — PHP 7.1+ short list syntax from superglobal.
		{
			Name:        "[$a, $b] = $_*[...]",
			Pattern:     `=\s*\$_(GET|POST|REQUEST|COOKIE)\s*\[`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Array destructuring from superglobal — PHP 7.1+ short list syntax",
			NodeTypes:   []string{"assignment_expression"},
		},
		// call_user_func($_GET['fn'], ...) — user controls function called.
		{
			Name:        "call_user_func($_*[...])",
			Pattern:     `\bcall_user_func(?:_array)?\s*\(\s*\$_(GET|POST|REQUEST|COOKIE)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "call_user_func() with user-controlled function name — dangerous dynamic dispatch",
			NodeTypes:   []string{"function_call_expression"},
		},
		// new $_GET['class']() — dynamic class instantiation from user input.
		{
			Name:        "new $_*[...]()",
			Pattern:     `\bnew\s+\$_(GET|POST|REQUEST|COOKIE)\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Dynamic class instantiation with user-controlled class name",
			NodeTypes:   []string{"object_creation_expression"},
		},
		// class_exists($_GET['class']) — user input used as class name lookup.
		{
			Name:        "class_exists/interface_exists($_*[...])",
			Pattern:     `\b(?:class_exists|interface_exists|trait_exists|function_exists)\s*\(\s*\$_(GET|POST|REQUEST|COOKIE)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "class_exists() / function_exists() with user-controlled name",
			NodeTypes:   []string{"function_call_expression"},
		},
	}
}

// phpSessionInputDefinitions returns definitions for PHP session data access.
// Session data stored in $_SESSION may have been set from user-controlled input
// in a previous request; tracing it catches stored-input propagation.
func phpSessionInputDefinitions() []common.Definition {
	return []common.Definition{
		// $_SESSION['key'] — reading from session.
		// NOTE: matcher.go already covers bare $_SESSION and subscript access.
		// This is intentionally a duplicate with UserInput label added so that
		// wiring sessionInputDefinitions() through the matcher adds the label.
		// Actually: the existing definition in matcher.go has empty Labels = []common.InputLabel{}.
		// We add a labeled version here for completeness — the aggregation will
		// deduplicate by name, so adding a new variant is safe.
		{
			Name:         "$_SESSION[...] (user-controlled)",
			Pattern:      `\$_SESSION\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Session read — $_SESSION values may originate from user input set in prior requests",
			NodeTypes:    []string{"subscript_expression"},
			KeyExtractor: `\$_SESSION\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
	}
}

// phpEnvironmentInputDefinitions returns definitions for environment variable
// access. Environment variables can be set by web-server configurations (CGI),
// by .env files, or by the process environment — they represent an external
// input source even if not directly from an HTTP request.
func phpEnvironmentInputDefinitions() []common.Definition {
	return []common.Definition{
		// getenv('VAR_NAME') — read a single environment variable.
		// NOTE: matcher.go already has a getenv() definition (Labels: LabelEnvironment).
		// This variant is included in the builtin group for organization; the
		// matcher.go definition will take precedence for exact deduplication.
		{
			Name:         "getenv('VAR')",
			Pattern:      `\bgetenv\s*\(\s*['"][^'"]+['"]`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "getenv('VAR') — read named environment variable",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bgetenv\s*\(\s*['"]([^'"]+)['"]`,
		},
		// apache_getenv() — Apache-specific env variable reader.
		{
			Name:         "apache_getenv()",
			Pattern:      `\bapache_getenv\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "apache_getenv() — read Apache environment variable (Apache SAPI only)",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bapache_getenv\s*\(\s*['"]([^'"]+)['"]`,
		},
		// putenv() with user-controlled value — treats the value as input.
		{
			Name:        "putenv() with user data",
			Pattern:     `\bputenv\s*\(\s*\$_(GET|POST|REQUEST|SERVER|COOKIE)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelEnvironment, common.LabelUserInput},
			Description: "putenv() called with superglobal value — environment modified by user input",
			NodeTypes:   []string{"function_call_expression"},
		},
		// $_ENV read — bare access to the environment superglobal.
		// NOTE: matcher.go already covers $_ENV['key'] subscript.
		// This version adds the UserInput label to the bare form.
		{
			Name:               "$_ENV (bare)",
			Pattern:            `^\$_ENV$`,
			Language:           "php",
			Labels:             []common.InputLabel{common.LabelEnvironment},
			Description:        "Bare $_ENV array — environment variables accessed as array",
			NodeTypes:          []string{"variable_name"},
			ExcludeParentTypes: []string{"subscript_expression"},
		},
	}
}

// phpHTTPAuthDefinitions returns definitions for PHP HTTP authentication input
// patterns. $PHP_AUTH_USER, $PHP_AUTH_PW, and $HTTP_AUTHORIZATION are set by
// the web server from the HTTP Authorization header and represent user-controlled
// credentials sent with each request. These are NOT covered by the generic
// $\_SERVER pattern's description but are semantically important as authentication
// input sources.
func phpHTTPAuthDefinitions() []common.Definition {
	return []common.Definition{
		// $_SERVER['PHP_AUTH_USER'] — username from HTTP Basic authentication.
		{
			Name:        "$_SERVER['PHP_AUTH_USER']",
			Pattern:     `\$_SERVER\s*\[\s*['"]PHP_AUTH_USER['"]\s*\]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "HTTP Basic auth username — $_SERVER['PHP_AUTH_USER'] set from Authorization header",
			NodeTypes:   []string{"subscript_expression"},
		},
		// $_SERVER['PHP_AUTH_PW'] — password from HTTP Basic authentication.
		{
			Name:        "$_SERVER['PHP_AUTH_PW']",
			Pattern:     `\$_SERVER\s*\[\s*['"]PHP_AUTH_PW['"]\s*\]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "HTTP Basic auth password — $_SERVER['PHP_AUTH_PW'] set from Authorization header",
			NodeTypes:   []string{"subscript_expression"},
		},
		// $_SERVER['PHP_AUTH_DIGEST'] — raw Digest auth header.
		{
			Name:        "$_SERVER['PHP_AUTH_DIGEST']",
			Pattern:     `\$_SERVER\s*\[\s*['"]PHP_AUTH_DIGEST['"]\s*\]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "HTTP Digest auth header — $_SERVER['PHP_AUTH_DIGEST'] set by PHP from Authorization header",
			NodeTypes:   []string{"subscript_expression"},
		},
		// $_SERVER['HTTP_AUTHORIZATION'] — raw Authorization header value.
		{
			Name:        "$_SERVER['HTTP_AUTHORIZATION']",
			Pattern:     `\$_SERVER\s*\[\s*['"]HTTP_AUTHORIZATION['"]\s*\]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Raw HTTP Authorization header — $_SERVER['HTTP_AUTHORIZATION']",
			NodeTypes:   []string{"subscript_expression"},
		},
	}
}

// phpFileUploadDefinitions returns definitions for PHP file upload processing
// functions. move_uploaded_file() and is_uploaded_file() operate on files
// uploaded via HTTP POST — the file content, name, and type are all
// user-controlled input that should be traced.
func phpFileUploadDefinitions() []common.Definition {
	return []common.Definition{
		// move_uploaded_file($_FILES[...]['tmp_name'], $dest) — processes an uploaded file.
		{
			Name:        "move_uploaded_file($_FILES[...])",
			Pattern:     `\bmove_uploaded_file\s*\(\s*\$_FILES`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "move_uploaded_file() with $_FILES — uploaded file moved from temp; content is user-controlled",
			NodeTypes:   []string{"function_call_expression"},
		},
		// is_uploaded_file($file) — validates a temp upload path; commonly used with $_FILES.
		{
			Name:        "is_uploaded_file($_FILES[...])",
			Pattern:     `\bis_uploaded_file\s*\(\s*\$_FILES`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "is_uploaded_file() with $_FILES — checks if file was uploaded via HTTP POST",
			NodeTypes:   []string{"function_call_expression"},
		},
	}
}

// phpRequestFactoryDefinitions returns definitions for static factory methods
// that create request objects from the current PHP superglobals. These are
// framework patterns (Symfony HttpFoundation, PSR-7 implementations) where
// the factory reads $_GET, $_POST, $_COOKIE, $_SERVER, $_FILES, and
// php://input to build a request object — the returned object is entirely
// user-controlled input.
func phpRequestFactoryDefinitions() []common.Definition {
	return []common.Definition{
		// Symfony HttpFoundation\Request::createFromGlobals() — extremely common (61 occurrences
		// in testapps), creates a Request from $_GET, $_POST, $_COOKIE, $_SERVER, $_FILES.
		// Tree-Sitter parses Class::method() as scoped_call_expression (not function_call_expression).
		{
			Name:        "Request::createFromGlobals()",
			Pattern:     `\bRequest::createFromGlobals\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelHTTPCookie, common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Symfony/Laravel Request::createFromGlobals() — creates request from all PHP superglobals",
			NodeTypes:   []string{"scoped_call_expression"},
		},
		// PSR-7 ServerRequestFactory::fromGlobals() — Slim/Laminas-style PSR-7 factory.
		{
			Name:        "ServerRequestFactory::fromGlobals()",
			Pattern:     `\bServerRequestFactory::fromGlobals\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelHTTPCookie, common.LabelHTTPHeader, common.LabelUserInput},
			Description: "PSR-7 ServerRequestFactory::fromGlobals() — creates server request from superglobals",
			NodeTypes:   []string{"scoped_call_expression"},
		},
		// Generic fromGlobals static factory (covers RequestFactory::fromGlobals etc.)
		{
			Name:        "::fromGlobals() (generic PSR-7 factory)",
			Pattern:     `::\s*fromGlobals\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelHTTPCookie, common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Static fromGlobals() factory — creates HTTP request from PHP superglobals (PSR-7 pattern)",
			NodeTypes:   []string{"scoped_call_expression"},
		},
		// PSR-7 instance method ->fromGlobals() — called on a factory object.
		// e.g. $factory->fromGlobals(), ServerRequestFactory::create()->fromGlobals()
		// Distinct from the static ClassName::fromGlobals() forms above.
		{
			Name:        "->fromGlobals() (PSR-7 instance factory)",
			Pattern:     `->\s*fromGlobals\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelHTTPCookie, common.LabelHTTPHeader, common.LabelUserInput},
			Description: "PSR-7 instance ->fromGlobals() — factory object creates server request from all superglobals",
			NodeTypes:   []string{"member_call_expression"},
		},
		// PSR-17 createStreamFromFile('php://input') — creates a stream from the raw POST body.
		// Used by Nyholm/PSR-7, Laminas Diactoros, and other PSR-17 implementations.
		{
			Name:        "->createStreamFromFile('php://input')",
			Pattern:     `->\s*createStreamFromFile\s*\(\s*['"]php://input['"]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "PSR-17 ->createStreamFromFile('php://input') — reads raw HTTP request body as stream",
			NodeTypes:   []string{"member_call_expression"},
		},
		// PSR-17 createStream('php://input') — alternative stream factory form.
		{
			Name:        "->createStream('php://input')",
			Pattern:     `->\s*createStream\s*\(\s*['"]php://input['"]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "PSR-17 ->createStream('php://input') — creates stream from raw HTTP request body",
			NodeTypes:   []string{"member_call_expression"},
		},
	}
}

// phpPSR7SingleParamDefinitions returns definitions for PSR-7 style single-parameter
// getters that are distinct from the plural forms already in matcher.go. These are
// commonly used in MediaWiki, Slim, and other frameworks that follow PSR-7 conventions
// with additional convenience methods.
func phpPSR7SingleParamDefinitions() []common.Definition {
	return []common.Definition{
		// ->getParsedBodyParam('key') — MediaWiki-specific (569 occurrences in testapps).
		// PSR-7 getParsedBody() returns the full parsed body; getParsedBodyParam() returns
		// a single named parameter from the parsed request body.
		{
			Name:         "->getParsedBodyParam()",
			Pattern:      `->\s*getParsedBodyParam\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelHTTPBody, common.LabelUserInput},
			Description:  "->getParsedBodyParam() — single POST body parameter (MediaWiki/Slim PSR-7 extension)",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getParsedBodyParam\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->getParsedBodyParamAsString('key') — string-typed variant (429 occurrences).
		{
			Name:         "->getParsedBodyParamAsString()",
			Pattern:      `->\s*getParsedBodyParamAsString\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelHTTPBody, common.LabelUserInput},
			Description:  "->getParsedBodyParamAsString() — typed POST body parameter getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getParsedBodyParamAsString\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->getQueryParam('key') — single query param getter (96 occurrences).
		// Distinct from ->getQueryParams() (plural, returns full array).
		{
			Name:         "->getQueryParam()",
			Pattern:      `->\s*getQueryParam\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "->getQueryParam() — single query string parameter getter (PSR-7 extension)",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getQueryParam\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->getVal('key') — MediaWiki WebRequest::getVal() (158 occurrences).
		// Returns a named GET/POST parameter with optional default value.
		{
			Name:         "->getVal()",
			Pattern:      `->\s*getVal\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description:  "->getVal() — MediaWiki WebRequest::getVal() / generic GET+POST parameter getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getVal\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->getText('key') — MediaWiki WebRequest::getText() — trimmed string variant.
		{
			Name:         "->getText()",
			Pattern:      `->\s*getText\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description:  "->getText() — MediaWiki WebRequest::getText() — trimmed text value of request parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getText\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->getCheck('key') — MediaWiki WebRequest::getCheck() — boolean checkbox parameter.
		{
			Name:         "->getCheck()",
			Pattern:      `->\s*getCheck\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "->getCheck() — MediaWiki WebRequest::getCheck() — boolean checkbox form parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getCheck\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->getParsedBodyParamAsStringOrNull('key') — MediaWiki nullable string POST parameter (168 occurrences).
		{
			Name:         "->getParsedBodyParamAsStringOrNull()",
			Pattern:      `->\s*getParsedBodyParamAsStringOrNull\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelHTTPBody, common.LabelUserInput},
			Description:  "->getParsedBodyParamAsStringOrNull() — nullable typed POST body parameter getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getParsedBodyParamAsStringOrNull\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->getRawVal('key') — MediaWiki WebRequest::getRawVal() — raw unfiltered value (47 occurrences).
		{
			Name:         "->getRawVal()",
			Pattern:      `->\s*getRawVal\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description:  "->getRawVal() — MediaWiki WebRequest::getRawVal() — raw unfiltered request parameter value",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getRawVal\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->getValues() — MediaWiki WebRequest::getValues() — returns all request params as array.
		{
			Name:        "->getValues()",
			Pattern:     `->\s*getValues\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "->getValues() — MediaWiki WebRequest::getValues() — returns all GET+POST parameters",
			NodeTypes:   []string{"member_call_expression"},
		},
		// ->wasPosted() — MediaWiki WebRequest::wasPosted() — checks if request is POST.
		// Technically a boolean check, not a value, but it confirms user input submission.
		{
			Name:        "->wasPosted()",
			Pattern:     `->\s*wasPosted\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "->wasPosted() — MediaWiki WebRequest::wasPosted() — true if HTTP POST request",
			NodeTypes:   []string{"member_call_expression"},
		},
		// ->getBody()->getContents() — PSR-7 stream body access pattern.
		// getBody() returns a StreamInterface; getContents() reads the full body as string.
		// Used in Slim, Laminas, and other PSR-7-compliant frameworks to read raw HTTP body.
		// Note: using a regex that matches both the chained call and bare getBody().
		{
			Name:        "->getBody()->getContents()",
			Pattern:     `->\s*getBody\s*\(\s*\)\s*->\s*getContents\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "->getBody()->getContents() — PSR-7 stream body read — raw HTTP request body",
			NodeTypes:   []string{"member_call_expression"},
		},
	}
}

// phpSymfonyParameterBagDefinitions returns definitions for Symfony HttpFoundation
// ParameterBag property access patterns. Symfony's Request object exposes named
// ParameterBag instances as public properties: ->query (GET), ->request (POST body),
// ->headers (HTTP headers), ->cookies, ->server, ->files, ->attributes.
// Each has ->get(), ->all(), ->has() methods.
func phpSymfonyParameterBagDefinitions() []common.Definition {
	// Shared key extractor for ->get('key') or ->all('key') patterns.
	kx := `->\s*(?:get|all|has)\s*\(\s*['"]([^'"]+)['"]`

	return []common.Definition{
		// ->query->get('key') — Symfony GET query parameters (82 occurrences in testapps).
		{
			Name:         "->query->get()",
			Pattern:      `->\s*query\s*->\s*(?:get|all|has)\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Symfony ParameterBag->query->get() — access HTTP GET parameters",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: kx,
		},
		// ->request->get('key') — Symfony POST body parameters (15 occurrences).
		{
			Name:         "->request->get()",
			Pattern:      `->\s*request\s*->\s*(?:get|all|has)\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Symfony ParameterBag->request->get() — access HTTP POST body parameters",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: kx,
		},
		// ->headers->get('key') — HTTP request headers (14 occurrences).
		{
			Name:         "->headers->get()",
			Pattern:      `->\s*headers\s*->\s*(?:get|all|has)\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "Symfony ParameterBag->headers->get() — access HTTP request headers",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: kx,
		},
		// ->cookies->get('key') / ->cookies->all() — cookie ParameterBag.
		{
			Name:         "->cookies->get()",
			Pattern:      `->\s*cookies\s*->\s*(?:get|all|has)\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:  "Symfony ParameterBag->cookies->get() — access HTTP cookies",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: kx,
		},
		// ->server->get('key') — server vars ParameterBag (29 occurrences).
		{
			Name:         "->server->get()",
			Pattern:      `->\s*server\s*->\s*(?:get|all|has|getInt)\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "Symfony ParameterBag->server->get() — access server/header variables",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: kx,
		},
		// ->files->get('key') — uploaded files ParameterBag.
		{
			Name:         "->files->get()",
			Pattern:      `->\s*files\s*->\s*(?:get|all|has)\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description:  "Symfony ParameterBag->files->get() — access uploaded files",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: kx,
		},
		// ->getHeaderLine('key') — PSR-7 single header value as string (18 occurrences).
		// Distinct from ->getHeader() which returns an array of header values.
		{
			Name:         "->getHeaderLine()",
			Pattern:      `->\s*getHeaderLine\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "->getHeaderLine() — PSR-7 single HTTP header value as concatenated string",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getHeaderLine\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}

// phpDispatcherDefinitions returns definitions for PHP dynamic dispatch patterns
// where framework routers pass user-controlled route parameters to controller
// methods via call_user_func_array or similar mechanisms.
func phpDispatcherDefinitions() []common.Definition {
	return []common.Definition{
		// call_user_func_array with route/request params as args
		{
			Name:        "call_user_func_array($fn, $routeParams)",
			Pattern:     `\bcall_user_func_array\s*\(\s*\[?\s*\$\w+\s*,?\s*(?:\$\w+|['"]?\w*['"]?)\s*\]?\s*,\s*\$(?:params|args|arguments|routeParams|parameters|matches)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Dynamic dispatch with route/request parameters — dispatched function receives tainted args",
			NodeTypes:   []string{"function_call_expression"},
		},
		// ReflectionMethod->invokeArgs with route params
		{
			Name:        "->invokeArgs($obj, $params)",
			Pattern:     `->\s*invokeArgs\s*\(\s*\$\w+\s*,\s*\$(?:params|args|arguments|routeParams|parameters)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Reflection invoke with route parameters — method receives tainted args",
			NodeTypes:   []string{"member_call_expression"},
		},
		// call_user_func with route-like variable names
		{
			Name:        "call_user_func($fn, ...$routeParams)",
			Pattern:     `\bcall_user_func\s*\(\s*\$\w+\s*,\s*\.\.\.\s*\$(?:params|args|arguments|routeParams|parameters|matches)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Dynamic dispatch with spread route parameters",
			NodeTypes:   []string{"function_call_expression"},
		},
		// Spread operator with route params in direct call
		{
			Name:        "$controller->$action(...$params)",
			Pattern:     `->\s*\$\w+\s*\(\s*\.\.\.\s*\$(?:params|args|arguments|routeParams|parameters|preparedArguments)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Dynamic method call with spread route parameters (TYPO3 Extbase pattern)",
			NodeTypes:   []string{"member_call_expression"},
		},
	}
}

// phpNonHTTPInputDefinitions returns definitions for PHP built-in functions that
// read input from sources OTHER than HTTP requests. These include email/IMAP,
// file metadata (EXIF), network protocols (SNMP, LDAP), and data import (CSV).
func phpNonHTTPInputDefinitions() []common.Definition {
	return []common.Definition{
		// ── Email/IMAP ────────────────────────────────────────────────────────
		{
			Name:        "imap_body()",
			Pattern:     `\bimap_body\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork, common.LabelUserInput},
			Description: "IMAP message body — email content from mailbox",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "imap_fetchbody()",
			Pattern:     `\bimap_fetchbody\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork, common.LabelUserInput},
			Description: "IMAP fetch body part — specific MIME part from mailbox",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "imap_headerinfo()",
			Pattern:     `\bimap_headerinfo\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork, common.LabelUserInput},
			Description: "IMAP header info — email headers from mailbox",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "imap_qprint()",
			Pattern:     `\bimap_qprint\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "IMAP quoted-printable decode",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "imap_base64()",
			Pattern:     `\bimap_base64\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "IMAP base64 decode",
			NodeTypes:   []string{"function_call_expression"},
		},
		// ── File metadata ────────────────────────────────────────────────────
		{
			Name:        "exif_read_data()",
			Pattern:     `\bexif_read_data\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "EXIF metadata from uploaded image — user-controlled EXIF tags",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "iptcparse()",
			Pattern:     `\biptcparse\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "IPTC metadata from uploaded image — user-controlled tags",
			NodeTypes:   []string{"function_call_expression"},
		},
		// ── Network/SNMP ─────────────────────────────────────────────────────
		{
			Name:        "snmpget()",
			Pattern:     `\bsnmpget\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "SNMP get — network device data",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "snmpwalk()",
			Pattern:     `\bsnmpwalk\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "SNMP walk — network device tree",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "snmp2_get()",
			Pattern:     `\bsnmp2_get\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "SNMPv2 get — network device data",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "snmp2_walk()",
			Pattern:     `\bsnmp2_walk\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "SNMPv2 walk — network device tree",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "snmp3_get()",
			Pattern:     `\bsnmp3_get\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "SNMPv3 get — network device data",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "snmp3_walk()",
			Pattern:     `\bsnmp3_walk\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "SNMPv3 walk — network device tree",
			NodeTypes:   []string{"function_call_expression"},
		},
		// ── Directory services/LDAP ──────────────────────────────────────────
		{
			Name:        "ldap_get_entries()",
			Pattern:     `\bldap_get_entries\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "LDAP entries — directory service data",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "ldap_get_values()",
			Pattern:     `\bldap_get_values\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "LDAP values — directory attribute values",
			NodeTypes:   []string{"function_call_expression"},
		},
		// ── Data import/CSV ──────────────────────────────────────────────────
		{
			Name:        "fgetcsv()",
			Pattern:     `\bfgetcsv\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "CSV line from file — user-uploaded CSV data",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "str_getcsv()",
			Pattern:     `\bstr_getcsv\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "CSV string parse — may contain user-controlled data",
			NodeTypes:   []string{"function_call_expression"},
		},
		// ── SOAP/XML-RPC ─────────────────────────────────────────────────────
		{
			Name:        "xmlrpc_server_call_method()",
			Pattern:     `\bxmlrpc_server_call_method\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork, common.LabelUserInput},
			Description: "XML-RPC method call — remote procedure data",
			NodeTypes:   []string{"function_call_expression"},
		},
		// ── CLI console framework accessors ──────────────────────────────────
		{
			Name:         "->argument()",
			Pattern:      `->\s*argument\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelCLI},
			Description:  "Console command argument (Laravel/Symfony)",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*argument\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->option()",
			Pattern:      `->\s*option\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelCLI},
			Description:  "Console command option (Laravel/Symfony)",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*option\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->ask()",
			Pattern:      `->\s*ask\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelCLI},
			Description:  "Console interactive ask (Laravel)",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*ask\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "->secret()",
			Pattern:     `->\s*secret\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Console secret input (Laravel)",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:         "->getArgument()",
			Pattern:      `->\s*getArgument\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelCLI},
			Description:  "Symfony Console getArgument()",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getArgument\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getOption()",
			Pattern:      `->\s*getOption\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelCLI},
			Description:  "Symfony Console getOption()",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getOption\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
