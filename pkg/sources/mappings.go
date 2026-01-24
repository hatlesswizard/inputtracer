// Package sources - mappings.go provides centralized language mappings
// Consolidated from pkg/semantic/mappings/ - that package should be deleted
package sources

import (
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
)

// FrameworkTypeInfo holds information about framework types that carry user input
type FrameworkTypeInfo struct {
	Framework  string
	SourceType SourceType
}

// LanguageMappings holds all input source mappings for a single language
type LanguageMappings struct {
	Language string

	// InputFunctions maps function/method names to source types
	InputFunctions map[string]SourceType

	// InputSources maps property/variable access patterns to source types
	InputSources map[string]SourceType

	// Superglobals maps superglobal variables to source types (PHP specific)
	Superglobals map[string]SourceType

	// DBFetchFunctions maps database fetch function names (PHP specific)
	DBFetchFunctions map[string]bool

	// GlobalSources maps browser global sources (JS specific)
	GlobalSources map[string]SourceType

	// DOMSources maps DOM property sources (JS specific)
	DOMSources map[string]SourceType

	// NodeSources maps Node.js-specific sources (JS specific)
	NodeSources map[string]SourceType

	// CGIEnvVars maps CGI environment variables to source types (C/C++ specific)
	CGIEnvVars map[string]SourceType

	// QtInputMethods maps Qt widget methods to source types (C++ specific)
	QtInputMethods map[string]SourceType

	// FrameworkTypes maps framework type names to their info (C++ specific)
	FrameworkTypes map[string]FrameworkTypeInfo

	// MethodInputs maps method names that return user input (C++ specific)
	MethodInputs map[string]SourceType

	// Annotations maps annotation/decorator names to source types (Java/C# specific)
	Annotations map[string]SourceType

	// InputMethods maps input method names to source types (Java specific)
	InputMethods map[string]SourceType
}

// mappingsRegistry holds all language mappings
var mappingsRegistry = make(map[string]*LanguageMappings)

// CGIEnvVars contains CGI environment variable mappings (shared C/C++)
var CGIEnvVars = map[string]SourceType{
	"QUERY_STRING":    SourceHTTPGet,
	"REQUEST_METHOD":  SourceHTTPHeader,
	"CONTENT_TYPE":    SourceHTTPHeader,
	"CONTENT_LENGTH":  SourceHTTPBody,
	"HTTP_COOKIE":     SourceHTTPCookie,
	"HTTP_HOST":       SourceHTTPHeader,
	"HTTP_USER_AGENT": SourceHTTPHeader,
	"HTTP_REFERER":    SourceHTTPHeader,
	"HTTP_ACCEPT":     SourceHTTPHeader,
	"PATH_INFO":       SourceHTTPPath,
	"PATH_TRANSLATED": SourceHTTPPath,
	"SCRIPT_NAME":     SourceHTTPPath,
	"REQUEST_URI":     SourceHTTPPath,
	"REMOTE_ADDR":     SourceNetwork,
	"REMOTE_HOST":     SourceNetwork,
	"SERVER_NAME":     SourceHTTPHeader,
	"SERVER_PORT":     SourceHTTPHeader,
	"HTTPS":           SourceHTTPHeader,
}

// StandardCInputFunctions contains C standard input functions
var StandardCInputFunctions = map[string]SourceType{
	"gets":          SourceStdin,
	"fgets":         SourceFile,
	"scanf":         SourceStdin,
	"fscanf":        SourceFile,
	"sscanf":        SourceUserInput,
	"getchar":       SourceStdin,
	"getc":          SourceFile,
	"fgetc":         SourceFile,
	"getline":       SourceStdin,
	"getdelim":      SourceFile,
	"read":          SourceFile,
	"pread":         SourceFile,
	"readv":         SourceFile,
	"preadv":        SourceFile,
	"fread":         SourceFile,
	"recv":          SourceNetwork,
	"recvfrom":      SourceNetwork,
	"recvmsg":       SourceNetwork,
	"recvmmsg":      SourceNetwork,
	"getenv":        SourceEnvVar,
	"secure_getenv": SourceEnvVar,
	"mmap":          SourceFile,
	"fopen":         SourceFile,
	"open":          SourceFile,
	"fdopen":        SourceFile,
}

func init() {
	registerGoMappings()
	registerPHPMappings()
	registerPythonMappings()
	registerJavaScriptMappings()
	registerTypeScriptMappings()
	registerCMappings()
	registerCPPMappings()
	registerJavaMappings()
	registerCSharpMappings()
	registerRubyMappings()
	registerRustMappings()
}

func registerGoMappings() {
	mappingsRegistry["go"] = &LanguageMappings{
		Language: "go",
		InputSources: map[string]SourceType{
			"r.Form":      SourceHTTPPost,
			"r.PostForm":  SourceHTTPPost,
			"r.URL.Query": SourceHTTPGet,
			"os.Args":     SourceCLIArg,
		},
		InputFunctions: map[string]SourceType{
			"r.FormValue": SourceHTTPPost, "r.PostFormValue": SourceHTTPPost,
			"r.URL.Query": SourceHTTPGet, "r.Body": SourceHTTPBody,
			"http.Request": SourceUserInput,
			"c.Query": SourceHTTPGet, "c.DefaultQuery": SourceHTTPGet,
			"c.Param": SourceHTTPPath, "c.PostForm": SourceHTTPPost,
			"c.DefaultPostForm": SourceHTTPPost, "c.GetHeader": SourceHTTPHeader,
			"c.Bind": SourceHTTPBody, "c.BindJSON": SourceHTTPBody,
			"c.ShouldBind": SourceHTTPBody, "c.ShouldBindJSON": SourceHTTPBody,
			"c.FormFile": SourceHTTPBody, "c.Cookie": SourceHTTPCookie,
			"c.QueryParam": SourceHTTPGet, "c.QueryParams": SourceHTTPGet,
			"c.FormValue": SourceHTTPPost, "c.FormParams": SourceHTTPPost,
			"Bind": SourceHTTPBody, "c.Request": SourceUserInput,
			"c.Params": SourceHTTPPath, "c.Queries": SourceHTTPGet,
			"c.BodyParser": SourceHTTPBody, "c.Body": SourceHTTPBody,
			"c.Get": SourceHTTPHeader, "c.Cookies": SourceHTTPCookie,
			"chi.URLParam": SourceHTTPPath, "URLParam": SourceHTTPPath,
			"mux.Vars": SourceHTTPPath, "Vars": SourceHTTPPath,
			"os.Getenv": SourceEnvVar, "flag.String": SourceCLIArg,
			"flag.Int": SourceCLIArg, "flag.Bool": SourceCLIArg,
			"flag.Parse": SourceCLIArg, "bufio.NewReader": SourceStdin,
			"bufio.NewScanner": SourceStdin, "ioutil.ReadFile": SourceFile,
			"os.ReadFile": SourceFile, "os.Open": SourceFile,
			"io.ReadAll": SourceUserInput,
		},
	}
}

func registerPHPMappings() {
	mappingsRegistry["php"] = &LanguageMappings{
		Language:     "php",
		Superglobals: SuperglobalToSourceType, // Use centralized superglobal mappings
		InputFunctions: map[string]SourceType{
			"file_get_contents":      SourceFile,
			"fgets":                  SourceFile,
			"fread":                  SourceFile,
			"fgetc":                  SourceFile,
			"fgetcsv":                SourceFile,
			"file":                   SourceFile,
			"readfile":               SourceFile,
			"getenv":                 SourceEnvVar,
			"apache_getenv":          SourceEnvVar,
			"getallheaders":          SourceHTTPHeader,
			"apache_request_headers": SourceHTTPHeader,
		},
		DBFetchFunctions: map[string]bool{
			"mysqli_fetch_array": true, "mysqli_fetch_assoc": true,
			"mysqli_fetch_row": true, "mysqli_fetch_object": true,
			"mysqli_fetch_all": true, "mysql_fetch_array": true,
			"mysql_fetch_assoc": true, "mysql_fetch_row": true,
			"mysql_fetch_object": true, "pg_fetch_array": true,
			"pg_fetch_assoc": true, "pg_fetch_row": true,
			"pg_fetch_object": true, "pg_fetch_all": true,
			"sqlite_fetch_array": true, "oci_fetch_array": true,
			"oci_fetch_assoc": true, "oci_fetch_row": true,
			"db_fetch_array": true,
		},
	}
}

func registerPythonMappings() {
	mappingsRegistry["python"] = &LanguageMappings{
		Language: "python",
		InputSources: map[string]SourceType{
			"request.args": SourceHTTPGet, "request.form": SourceHTTPPost,
			"request.data": SourceHTTPBody, "request.json": SourceHTTPJSON,
			"request.files": SourceHTTPBody, "request.cookies": SourceHTTPCookie,
			"request.headers": SourceHTTPHeader, "request.values": SourceHTTPGet,
			"request.GET": SourceHTTPGet, "request.POST": SourceHTTPPost,
			"request.FILES": SourceHTTPBody, "request.COOKIES": SourceHTTPCookie,
			"request.META": SourceHTTPHeader, "request.body": SourceHTTPBody,
			"request.query": SourceHTTPGet, "request.match_info": SourceHTTPPath,
			"request.rel_url": SourceHTTPGet, "request.ctx": SourceUserInput,
			"request.path_params": SourceHTTPPath, "request.query_params": SourceHTTPGet,
			"sys.argv": SourceCLIArg, "os.environ": SourceEnvVar,
		},
		InputFunctions: map[string]SourceType{
			"input": SourceStdin, "raw_input": SourceStdin,
			"getenv": SourceEnvVar, "os.getenv": SourceEnvVar,
			"environ.get": SourceEnvVar, "open": SourceFile,
			"read": SourceFile, "readline": SourceFile, "readlines": SourceFile,
			"get_argument": SourceHTTPGet, "get_query_argument": SourceHTTPGet,
			"get_body_argument": SourceHTTPPost, "request.post": SourceHTTPPost,
			"parse_args": SourceCLIArg, "add_argument": SourceCLIArg,
			"requests.get": SourceNetwork, "requests.post": SourceNetwork,
			"response.json": SourceNetwork, "response.text": SourceNetwork,
		},
	}
}

func registerJavaScriptMappings() {
	mappingsRegistry["javascript"] = &LanguageMappings{
		Language: "javascript",
		GlobalSources: map[string]SourceType{
			"location.href": SourceHTTPGet, "location.search": SourceHTTPGet,
			"location.hash": SourceHTTPGet, "location.pathname": SourceHTTPPath,
			"document.URL": SourceHTTPGet, "document.referrer": SourceHTTPHeader,
			"document.cookie": SourceHTTPCookie, "window.location": SourceHTTPGet,
			"window.name": SourceUserInput,
		},
		DOMSources: map[string]SourceType{
			"value": SourceUserInput, "innerHTML": SourceUserInput,
			"innerText": SourceUserInput, "textContent": SourceUserInput,
			"data": SourceUserInput,
		},
		NodeSources: map[string]SourceType{
			"process.argv": SourceCLIArg, "process.env": SourceEnvVar,
			"process.stdin": SourceStdin,
		},
	}
}

func registerTypeScriptMappings() {
	mappingsRegistry["typescript"] = &LanguageMappings{
		Language: "typescript",
		InputSources: map[string]SourceType{
			"req.body": SourceHTTPBody, "req.query": SourceHTTPGet,
			"req.params": SourceHTTPPath, "req.headers": SourceHTTPHeader,
			"req.cookies": SourceHTTPCookie, "request.body": SourceHTTPBody,
			"request.query": SourceHTTPGet, "request.params": SourceHTTPPath,
			"request.headers": SourceHTTPHeader, "request.cookies": SourceHTTPCookie,
			"process.argv": SourceCLIArg, "process.env": SourceEnvVar,
		},
		InputFunctions: map[string]SourceType{
			"prompt": SourceStdin, "readline": SourceStdin,
			"readFileSync": SourceFile, "readFile": SourceFile,
			"fetch": SourceNetwork,
		},
	}
}

func registerCMappings() {
	mappingsRegistry["c"] = &LanguageMappings{
		Language:       "c",
		InputFunctions: StandardCInputFunctions,
		CGIEnvVars:     CGIEnvVars,
	}
}

func registerCPPMappings() {
	cppInputFunctions := make(map[string]SourceType)
	for k, v := range StandardCInputFunctions {
		cppInputFunctions[k] = v
	}
	cppInputFunctions["cin"] = SourceStdin
	cppInputFunctions["getline"] = SourceStdin
	cppInputFunctions["async_read"] = SourceNetwork
	cppInputFunctions["async_read_some"] = SourceNetwork
	cppInputFunctions["async_read_until"] = SourceNetwork

	mappingsRegistry["cpp"] = &LanguageMappings{
		Language:       "cpp",
		InputFunctions: cppInputFunctions,
		CGIEnvVars:     CGIEnvVars,
		QtInputMethods: map[string]SourceType{
			"text": SourceUserInput, "displayText": SourceUserInput,
			"selectedText": SourceUserInput, "toPlainText": SourceUserInput,
			"toHtml": SourceUserInput, "currentText": SourceUserInput,
			"currentIndex": SourceUserInput, "itemText": SourceUserInput,
			"value": SourceUserInput, "cleanText": SourceUserInput,
			"sliderPosition": SourceUserInput, "date": SourceUserInput,
			"time": SourceUserInput, "dateTime": SourceUserInput,
			"isChecked": SourceUserInput, "checkState": SourceUserInput,
			"currentItem": SourceUserInput, "selectedItems": SourceUserInput,
			"readAll": SourceFile, "readLine": SourceFile,
			"readAllStandardOutput": SourceUserInput,
			"readAllStandardError":  SourceUserInput,
			"qgetenv":               SourceEnvVar,
		},
		FrameworkTypes: map[string]FrameworkTypeInfo{
			"QNetworkReply":            {Framework: "qt", SourceType: SourceNetwork},
			"QNetworkRequest":          {Framework: "qt", SourceType: SourceNetwork},
			"QHttpPart":                {Framework: "qt", SourceType: SourceHTTPBody},
			"request":                  {Framework: "boost.beast", SourceType: SourceHTTPBody},
			"http::request":            {Framework: "boost.beast", SourceType: SourceHTTPBody},
			"beast::http::request":     {Framework: "boost.beast", SourceType: SourceHTTPBody},
			"websocket::stream":        {Framework: "boost.beast", SourceType: SourceNetwork},
			"crow::request":            {Framework: "crow", SourceType: SourceHTTPBody},
			"crow::response":           {Framework: "crow", SourceType: SourceHTTPBody},
			"HttpRequestPtr":           {Framework: "drogon", SourceType: SourceHTTPBody},
			"HttpRequest":              {Framework: "drogon", SourceType: SourceHTTPBody},
			"HttpResponsePtr":          {Framework: "drogon", SourceType: SourceHTTPBody},
			"WebSocketConnectionPtr":   {Framework: "drogon", SourceType: SourceNetwork},
			"http_request":             {Framework: "cpprestsdk", SourceType: SourceHTTPBody},
			"http_response":            {Framework: "cpprestsdk", SourceType: SourceHTTPBody},
			"web::http::http_request":  {Framework: "cpprestsdk", SourceType: SourceHTTPBody},
			"HTTPServerRequest":        {Framework: "poco", SourceType: SourceHTTPBody},
			"HTTPRequest":              {Framework: "poco", SourceType: SourceHTTPBody},
			"HTMLForm":                 {Framework: "poco", SourceType: SourceHTTPPost},
			"ifstream":                 {Framework: "std", SourceType: SourceFile},
			"fstream":                  {Framework: "std", SourceType: SourceFile},
			"istringstream":            {Framework: "std", SourceType: SourceUserInput},
			"stringstream":             {Framework: "std", SourceType: SourceUserInput},
		},
		MethodInputs: map[string]SourceType{
			"body": SourceHTTPBody, "url_params": SourceHTTPGet,
			"get_header_value": SourceHTTPHeader, "getBody": SourceHTTPBody,
			"getParameter": SourceHTTPGet, "getQuery": SourceHTTPGet,
			"getCookie": SourceHTTPCookie, "getHeader": SourceHTTPHeader,
			"getPath": SourceHTTPPath, "getJsonObject": SourceHTTPJSON,
			"target": SourceHTTPPath, "at": SourceHTTPHeader,
			"find": SourceHTTPHeader, "extract_json": SourceHTTPJSON,
			"extract_string": SourceHTTPBody, "request_uri": SourceHTTPPath,
			"headers": SourceHTTPHeader, "getURI": SourceHTTPPath,
			"getMethod": SourceHTTPHeader, "getContentType": SourceHTTPHeader,
			"stream": SourceHTTPBody, "get": SourceUserInput,
			"getline": SourceUserInput, "read": SourceUserInput,
			"peek": SourceUserInput, "read_some": SourceNetwork,
			"receive": SourceNetwork, "async_receive": SourceNetwork,
		},
	}
}

func registerJavaMappings() {
	mappingsRegistry["java"] = &LanguageMappings{
		Language: "java",
		InputMethods: map[string]SourceType{
			"getParameter": SourceHTTPGet, "getParameterValues": SourceHTTPGet,
			"getParameterMap": SourceHTTPGet, "getParameterNames": SourceHTTPGet,
			"getHeader": SourceHTTPHeader, "getHeaders": SourceHTTPHeader,
			"getHeaderNames": SourceHTTPHeader, "getCookies": SourceHTTPCookie,
			"getInputStream": SourceHTTPBody, "getReader": SourceHTTPBody,
			"getQueryString": SourceHTTPGet, "getRequestURI": SourceHTTPPath,
			"getRequestURL": SourceHTTPPath, "getPathInfo": SourceHTTPPath,
			"getServletPath": SourceHTTPPath, "getRemoteAddr": SourceHTTPHeader,
			"getBody": SourceHTTPBody, "getParam": SourceHTTPGet,
			"getParams": SourceHTTPGet, "bodyAsString": SourceHTTPBody,
			"bodyAsJson": SourceHTTPBody, "getenv": SourceEnvVar,
			"getProperty": SourceEnvVar, "nextLine": SourceStdin,
			"next": SourceStdin, "nextInt": SourceStdin,
			"nextDouble": SourceStdin, "nextBoolean": SourceStdin,
			"readLine": SourceFile, "readAllBytes": SourceFile,
			"readAllLines": SourceFile,
		},
	}
}

func registerCSharpMappings() {
	mappingsRegistry["c_sharp"] = &LanguageMappings{
		Language: "c_sharp",
		InputSources: map[string]SourceType{
			"Console.ReadLine":                   SourceStdin,
			"Console.Read":                       SourceStdin,
			"Console.ReadKey":                    SourceStdin,
			"Environment.GetEnvironmentVariable": SourceEnvVar,
			"Environment.GetCommandLineArgs":     SourceCLIArg,
			"Request.Query":                      SourceHTTPGet,
			"Request.Form":                       SourceHTTPPost,
			"Request.Body":                       SourceHTTPBody,
			"Request.Headers":                    SourceHTTPHeader,
			"Request.Cookies":                    SourceHTTPCookie,
			"HttpContext.Request":                SourceUserInput,
			"File.ReadAllText":                   SourceFile,
			"File.ReadAllLines":                  SourceFile,
			"File.ReadAllBytes":                  SourceFile,
			"StreamReader.ReadLine":              SourceFile,
			"StreamReader.ReadToEnd":             SourceFile,
		},
	}
}

func registerRubyMappings() {
	mappingsRegistry["ruby"] = &LanguageMappings{
		Language: "ruby",
		InputSources: map[string]SourceType{
			"gets": SourceStdin, "readline": SourceStdin,
			"readlines": SourceStdin, "STDIN": SourceStdin,
			"ARGF": SourceStdin, "ARGV": SourceCLIArg,
			"ENV": SourceEnvVar, "params": SourceHTTPGet,
			"request": SourceUserInput, "cookies": SourceHTTPCookie,
			"session": SourceUserInput, "File.read": SourceFile,
			"IO.read": SourceFile,
		},
	}
}

func registerRustMappings() {
	mappingsRegistry["rust"] = &LanguageMappings{
		Language: "rust",
		InputSources: map[string]SourceType{
			"env::args": SourceCLIArg, "env::args_os": SourceCLIArg,
			"env::var": SourceEnvVar, "env::var_os": SourceEnvVar,
			"stdin": SourceStdin, "read_line": SourceStdin,
			"BufRead": SourceStdin, "fs::read": SourceFile,
			"read_to_string": SourceFile, "File::open": SourceFile,
			"web::Query": SourceHTTPGet, "web::Form": SourceHTTPPost,
			"web::Json": SourceHTTPBody, "web::Path": SourceHTTPGet,
			"Query<": SourceHTTPGet, "Form<": SourceHTTPPost,
			"Json<": SourceHTTPBody, "Path<": SourceHTTPGet,
		},
	}
}

// GetMappings returns the mappings for a specific language
func GetMappings(language string) *LanguageMappings {
	return mappingsRegistry[language]
}

// GetInputFunctions returns InputFunctions for a language, never nil
func GetInputFunctions(language string) map[string]SourceType {
	if lm := mappingsRegistry[language]; lm != nil && lm.InputFunctions != nil {
		return lm.InputFunctions
	}
	return make(map[string]SourceType)
}

// GetInputSources returns InputSources for a language, never nil
func GetInputSources(language string) map[string]SourceType {
	if lm := mappingsRegistry[language]; lm != nil && lm.InputSources != nil {
		return lm.InputSources
	}
	return make(map[string]SourceType)
}

// GetDBFetchFunctions returns DBFetchFunctions for PHP, never nil
func GetDBFetchFunctions() map[string]bool {
	if lm := mappingsRegistry["php"]; lm != nil && lm.DBFetchFunctions != nil {
		return lm.DBFetchFunctions
	}
	return make(map[string]bool)
}

// MergeMaps combines multiple source type maps into one
func MergeMaps(maps ...map[string]SourceType) map[string]SourceType {
	result := make(map[string]SourceType)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// convertToTypesSourceType converts sources.SourceType map to types.SourceType map
func convertToTypesSourceType(m map[string]SourceType) map[string]types.SourceType {
	if m == nil {
		return make(map[string]types.SourceType)
	}
	result := make(map[string]types.SourceType, len(m))
	for k, v := range m {
		result[k] = types.SourceType(v)
	}
	return result
}

// GetSuperglobalsMap returns Superglobals as types.SourceType map
func (lm *LanguageMappings) GetSuperglobalsMap() map[string]types.SourceType {
	return convertToTypesSourceType(lm.Superglobals)
}

// GetInputFunctionsMap returns InputFunctions as types.SourceType map
func (lm *LanguageMappings) GetInputFunctionsMap() map[string]types.SourceType {
	return convertToTypesSourceType(lm.InputFunctions)
}

// GetInputSourcesMap returns InputSources as types.SourceType map
func (lm *LanguageMappings) GetInputSourcesMap() map[string]types.SourceType {
	return convertToTypesSourceType(lm.InputSources)
}

// GetDBFetchFunctionsMap returns DBFetchFunctions map
func (lm *LanguageMappings) GetDBFetchFunctionsMap() map[string]bool {
	if lm.DBFetchFunctions == nil {
		return make(map[string]bool)
	}
	return lm.DBFetchFunctions
}

// GetGlobalSourcesMap returns GlobalSources as types.SourceType map
func (lm *LanguageMappings) GetGlobalSourcesMap() map[string]types.SourceType {
	return convertToTypesSourceType(lm.GlobalSources)
}

// GetDOMSourcesMap returns DOMSources as types.SourceType map
func (lm *LanguageMappings) GetDOMSourcesMap() map[string]types.SourceType {
	return convertToTypesSourceType(lm.DOMSources)
}

// GetNodeSourcesMap returns NodeSources as types.SourceType map
func (lm *LanguageMappings) GetNodeSourcesMap() map[string]types.SourceType {
	return convertToTypesSourceType(lm.NodeSources)
}

// GetCGIEnvVarsMap returns CGIEnvVars as types.SourceType map
func (lm *LanguageMappings) GetCGIEnvVarsMap() map[string]types.SourceType {
	return convertToTypesSourceType(lm.CGIEnvVars)
}

// GetQtInputMethodsMap returns QtInputMethods as types.SourceType map
func (lm *LanguageMappings) GetQtInputMethodsMap() map[string]types.SourceType {
	return convertToTypesSourceType(lm.QtInputMethods)
}

// GetFrameworkTypesMap returns FrameworkTypes map
func (lm *LanguageMappings) GetFrameworkTypesMap() map[string]FrameworkTypeInfo {
	if lm.FrameworkTypes == nil {
		return make(map[string]FrameworkTypeInfo)
	}
	return lm.FrameworkTypes
}

// GetMethodInputsMap returns MethodInputs as types.SourceType map
func (lm *LanguageMappings) GetMethodInputsMap() map[string]types.SourceType {
	return convertToTypesSourceType(lm.MethodInputs)
}

// GetInputMethodsMap returns InputMethods as types.SourceType map
func (lm *LanguageMappings) GetInputMethodsMap() map[string]types.SourceType {
	return convertToTypesSourceType(lm.InputMethods)
}
