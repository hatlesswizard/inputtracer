package mappings

import (
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
)

func init() {
	// Register all language mappings
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
	Registry["go"] = &LanguageMappings{
		Language: "go",
		InputSources: map[string]types.SourceType{
			"r.Form":      types.SourceHTTPPost,
			"r.PostForm":  types.SourceHTTPPost,
			"r.URL.Query": types.SourceHTTPGet,
			"os.Args":     types.SourceCLIArg,
		},
		InputFunctions: map[string]types.SourceType{
			// Standard HTTP
			"r.FormValue":       types.SourceHTTPPost,
			"r.PostFormValue":   types.SourceHTTPPost,
			"r.URL.Query":       types.SourceHTTPGet,
			"r.Body":            types.SourceHTTPBody,
			"http.Request":      types.SourceUserInput,
			// Gin
			"c.Query":           types.SourceHTTPGet,
			"c.DefaultQuery":    types.SourceHTTPGet,
			"c.Param":           types.SourceHTTPPath,
			"c.PostForm":        types.SourceHTTPPost,
			"c.DefaultPostForm": types.SourceHTTPPost,
			"c.GetHeader":       types.SourceHTTPHeader,
			"c.Bind":            types.SourceHTTPBody,
			"c.BindJSON":        types.SourceHTTPBody,
			"c.ShouldBind":      types.SourceHTTPBody,
			"c.ShouldBindJSON":  types.SourceHTTPBody,
			"c.FormFile":        types.SourceHTTPBody,
			"c.Cookie":          types.SourceHTTPCookie,
			// Echo
			"c.QueryParam":  types.SourceHTTPGet,
			"c.QueryParams": types.SourceHTTPGet,
			"c.FormValue":   types.SourceHTTPPost,
			"c.FormParams":  types.SourceHTTPPost,
			"Bind":          types.SourceHTTPBody,
			"c.Request":     types.SourceUserInput,
			// Fiber
			"c.Params":     types.SourceHTTPPath,
			"c.Queries":    types.SourceHTTPGet,
			"c.BodyParser": types.SourceHTTPBody,
			"c.Body":       types.SourceHTTPBody,
			"c.Get":        types.SourceHTTPHeader,
			"c.Cookies":    types.SourceHTTPCookie,
			// Chi
			"chi.URLParam": types.SourceHTTPPath,
			"URLParam":     types.SourceHTTPPath,
			// Gorilla mux
			"mux.Vars": types.SourceHTTPPath,
			"Vars":     types.SourceHTTPPath,
			// OS/CLI
			"os.Getenv":   types.SourceEnvVar,
			"flag.String": types.SourceCLIArg,
			"flag.Int":    types.SourceCLIArg,
			"flag.Bool":   types.SourceCLIArg,
			"flag.Parse":  types.SourceCLIArg,
			// File/stdin
			"bufio.NewReader":  types.SourceStdin,
			"bufio.NewScanner": types.SourceStdin,
			"ioutil.ReadFile":  types.SourceFile,
			"os.ReadFile":      types.SourceFile,
			"os.Open":          types.SourceFile,
			"io.ReadAll":       types.SourceUserInput,
		},
	}
}

func registerPHPMappings() {
	Registry["php"] = &LanguageMappings{
		Language: "php",
		Superglobals: map[string]types.SourceType{
			"$_GET":     types.SourceHTTPGet,
			"$_POST":    types.SourceHTTPPost,
			"$_REQUEST": types.SourceHTTPGet,
			"$_COOKIE":  types.SourceHTTPCookie,
			"$_SERVER":  types.SourceHTTPHeader,
			"$_FILES":   types.SourceHTTPBody,
			"$_ENV":     types.SourceEnvVar,
			"$_SESSION": types.SourceUserInput,
		},
		InputFunctions: map[string]types.SourceType{
			"file_get_contents":      types.SourceFile,
			"fgets":                  types.SourceFile,
			"fread":                  types.SourceFile,
			"fgetc":                  types.SourceFile,
			"fgetcsv":                types.SourceFile,
			"file":                   types.SourceFile,
			"readfile":               types.SourceFile,
			"getenv":                 types.SourceEnvVar,
			"apache_getenv":          types.SourceEnvVar,
			"getallheaders":          types.SourceHTTPHeader,
			"apache_request_headers": types.SourceHTTPHeader,
		},
		DBFetchFunctions: map[string]bool{
			"mysqli_fetch_array":  true,
			"mysqli_fetch_assoc":  true,
			"mysqli_fetch_row":    true,
			"mysqli_fetch_object": true,
			"mysqli_fetch_all":    true,
			"mysql_fetch_array":   true,
			"mysql_fetch_assoc":   true,
			"mysql_fetch_row":     true,
			"mysql_fetch_object":  true,
			"pg_fetch_array":      true,
			"pg_fetch_assoc":      true,
			"pg_fetch_row":        true,
			"pg_fetch_object":     true,
			"pg_fetch_all":        true,
			"sqlite_fetch_array":  true,
			"oci_fetch_array":     true,
			"oci_fetch_assoc":     true,
			"oci_fetch_row":       true,
			"db_fetch_array":      true,
		},
	}
}

func registerPythonMappings() {
	Registry["python"] = &LanguageMappings{
		Language: "python",
		InputSources: map[string]types.SourceType{
			// Flask
			"request.args":    types.SourceHTTPGet,
			"request.form":    types.SourceHTTPPost,
			"request.data":    types.SourceHTTPBody,
			"request.json":    types.SourceHTTPJSON,
			"request.files":   types.SourceHTTPBody,
			"request.cookies": types.SourceHTTPCookie,
			"request.headers": types.SourceHTTPHeader,
			"request.values":  types.SourceHTTPGet,
			// Django
			"request.GET":     types.SourceHTTPGet,
			"request.POST":    types.SourceHTTPPost,
			"request.FILES":   types.SourceHTTPBody,
			"request.COOKIES": types.SourceHTTPCookie,
			"request.META":    types.SourceHTTPHeader,
			"request.body":    types.SourceHTTPBody,
			// aiohttp
			"request.query":      types.SourceHTTPGet,
			"request.match_info": types.SourceHTTPPath,
			"request.rel_url":    types.SourceHTTPGet,
			// Sanic
			"request.ctx": types.SourceUserInput,
			// FastAPI / Starlette
			"request.path_params":  types.SourceHTTPPath,
			"request.query_params": types.SourceHTTPGet,
			// CLI
			"sys.argv":   types.SourceCLIArg,
			"os.environ": types.SourceEnvVar,
		},
		InputFunctions: map[string]types.SourceType{
			// Built-in
			"input":     types.SourceStdin,
			"raw_input": types.SourceStdin,
			// OS
			"getenv":      types.SourceEnvVar,
			"os.getenv":   types.SourceEnvVar,
			"environ.get": types.SourceEnvVar,
			// File
			"open":      types.SourceFile,
			"read":      types.SourceFile,
			"readline":  types.SourceFile,
			"readlines": types.SourceFile,
			// Tornado
			"get_argument":       types.SourceHTTPGet,
			"get_query_argument": types.SourceHTTPGet,
			"get_body_argument":  types.SourceHTTPPost,
			// aiohttp
			"request.post": types.SourceHTTPPost,
			// argparse
			"parse_args":   types.SourceCLIArg,
			"add_argument": types.SourceCLIArg,
			// requests (HTTP client)
			"requests.get":  types.SourceNetwork,
			"requests.post": types.SourceNetwork,
			"response.json": types.SourceNetwork,
			"response.text": types.SourceNetwork,
		},
	}
}

func registerJavaScriptMappings() {
	Registry["javascript"] = &LanguageMappings{
		Language: "javascript",
		GlobalSources: map[string]types.SourceType{
			"location.href":     types.SourceHTTPGet,
			"location.search":   types.SourceHTTPGet,
			"location.hash":     types.SourceHTTPGet,
			"location.pathname": types.SourceHTTPPath,
			"document.URL":      types.SourceHTTPGet,
			"document.referrer": types.SourceHTTPHeader,
			"document.cookie":   types.SourceHTTPCookie,
			"window.location":   types.SourceHTTPGet,
			"window.name":       types.SourceUserInput,
		},
		DOMSources: map[string]types.SourceType{
			"value":       types.SourceUserInput,
			"innerHTML":   types.SourceUserInput,
			"innerText":   types.SourceUserInput,
			"textContent": types.SourceUserInput,
			"data":        types.SourceUserInput,
		},
		NodeSources: map[string]types.SourceType{
			"process.argv":  types.SourceCLIArg,
			"process.env":   types.SourceEnvVar,
			"process.stdin": types.SourceStdin,
		},
	}
}

func registerTypeScriptMappings() {
	Registry["typescript"] = &LanguageMappings{
		Language: "typescript",
		InputSources: map[string]types.SourceType{
			"req.body":        types.SourceHTTPBody,
			"req.query":       types.SourceHTTPGet,
			"req.params":      types.SourceHTTPPath,
			"req.headers":     types.SourceHTTPHeader,
			"req.cookies":     types.SourceHTTPCookie,
			"request.body":    types.SourceHTTPBody,
			"request.query":   types.SourceHTTPGet,
			"request.params":  types.SourceHTTPPath,
			"request.headers": types.SourceHTTPHeader,
			"request.cookies": types.SourceHTTPCookie,
			"process.argv":    types.SourceCLIArg,
			"process.env":     types.SourceEnvVar,
		},
		InputFunctions: map[string]types.SourceType{
			"prompt":       types.SourceStdin,
			"readline":     types.SourceStdin,
			"readFileSync": types.SourceFile,
			"readFile":     types.SourceFile,
			"fetch":        types.SourceNetwork,
		},
	}
}

func registerCMappings() {
	Registry["c"] = &LanguageMappings{
		Language:       "c",
		InputFunctions: StandardCInputFunctions,
		CGIEnvVars:     CGIEnvVars,
	}
}

func registerCPPMappings() {
	// C++ extends C standard functions with C++-specific ones
	cppInputFunctions := MergeMaps(StandardCInputFunctions, map[string]types.SourceType{
		// C++ streams
		"cin":     types.SourceStdin,
		"getline": types.SourceStdin,
		// Boost.Asio network
		"async_read":       types.SourceNetwork,
		"async_read_some":  types.SourceNetwork,
		"async_read_until": types.SourceNetwork,
	})

	Registry["cpp"] = &LanguageMappings{
		Language:       "cpp",
		InputFunctions: cppInputFunctions,
		CGIEnvVars:     CGIEnvVars,
		QtInputMethods: map[string]types.SourceType{
			// QLineEdit
			"text":         types.SourceUserInput,
			"displayText":  types.SourceUserInput,
			"selectedText": types.SourceUserInput,
			// QTextEdit / QPlainTextEdit
			"toPlainText": types.SourceUserInput,
			"toHtml":      types.SourceUserInput,
			// QComboBox
			"currentText":  types.SourceUserInput,
			"currentIndex": types.SourceUserInput,
			"itemText":     types.SourceUserInput,
			// QSpinBox / QDoubleSpinBox
			"value":     types.SourceUserInput,
			"cleanText": types.SourceUserInput,
			// QSlider / QScrollBar
			"sliderPosition": types.SourceUserInput,
			// QDateEdit / QTimeEdit / QDateTimeEdit
			"date":     types.SourceUserInput,
			"time":     types.SourceUserInput,
			"dateTime": types.SourceUserInput,
			// QCheckBox / QRadioButton
			"isChecked":  types.SourceUserInput,
			"checkState": types.SourceUserInput,
			// QListWidget / QTreeWidget / QTableWidget
			"currentItem":   types.SourceUserInput,
			"selectedItems": types.SourceUserInput,
			// QFile / QIODevice / QNetworkReply
			"readAll":  types.SourceFile,
			"readLine": types.SourceFile,
			// QProcess
			"readAllStandardOutput": types.SourceUserInput,
			"readAllStandardError":  types.SourceUserInput,
			// Qt environment
			"qgetenv": types.SourceEnvVar,
		},
		FrameworkTypes: map[string]FrameworkTypeInfo{
			// Qt types
			"QNetworkReply":   {Framework: "qt", SourceType: types.SourceNetwork},
			"QNetworkRequest": {Framework: "qt", SourceType: types.SourceNetwork},
			"QHttpPart":       {Framework: "qt", SourceType: types.SourceHTTPBody},
			// Boost.Beast types
			"request":              {Framework: "boost.beast", SourceType: types.SourceHTTPBody},
			"http::request":        {Framework: "boost.beast", SourceType: types.SourceHTTPBody},
			"beast::http::request": {Framework: "boost.beast", SourceType: types.SourceHTTPBody},
			"websocket::stream":    {Framework: "boost.beast", SourceType: types.SourceNetwork},
			// Crow framework types
			"crow::request":  {Framework: "crow", SourceType: types.SourceHTTPBody},
			"crow::response": {Framework: "crow", SourceType: types.SourceHTTPBody},
			// Drogon framework types
			"HttpRequestPtr":        {Framework: "drogon", SourceType: types.SourceHTTPBody},
			"HttpRequest":           {Framework: "drogon", SourceType: types.SourceHTTPBody},
			"HttpResponsePtr":       {Framework: "drogon", SourceType: types.SourceHTTPBody},
			"WebSocketConnectionPtr": {Framework: "drogon", SourceType: types.SourceNetwork},
			// cpprestsdk (Casablanca)
			"http_request":            {Framework: "cpprestsdk", SourceType: types.SourceHTTPBody},
			"http_response":           {Framework: "cpprestsdk", SourceType: types.SourceHTTPBody},
			"web::http::http_request": {Framework: "cpprestsdk", SourceType: types.SourceHTTPBody},
			// Poco framework types
			"HTTPServerRequest": {Framework: "poco", SourceType: types.SourceHTTPBody},
			"HTTPRequest":       {Framework: "poco", SourceType: types.SourceHTTPBody},
			"HTMLForm":          {Framework: "poco", SourceType: types.SourceHTTPPost},
			// Standard streams as types
			"ifstream":      {Framework: "std", SourceType: types.SourceFile},
			"fstream":       {Framework: "std", SourceType: types.SourceFile},
			"istringstream": {Framework: "std", SourceType: types.SourceUserInput},
			"stringstream":  {Framework: "std", SourceType: types.SourceUserInput},
		},
		MethodInputs: map[string]types.SourceType{
			// HTTP body access (Crow, Beast, cpprestsdk, etc.)
			"body": types.SourceHTTPBody,
			// Crow framework methods
			"url_params":       types.SourceHTTPGet,
			"get_header_value": types.SourceHTTPHeader,
			// Drogon framework methods
			"getBody":       types.SourceHTTPBody,
			"getParameter":  types.SourceHTTPGet,
			"getQuery":      types.SourceHTTPGet,
			"getCookie":     types.SourceHTTPCookie,
			"getHeader":     types.SourceHTTPHeader,
			"getPath":       types.SourceHTTPPath,
			"getJsonObject": types.SourceHTTPJSON,
			// Boost.Beast methods
			"target": types.SourceHTTPPath,
			"at":     types.SourceHTTPHeader,
			"find":   types.SourceHTTPHeader,
			// cpprestsdk methods
			"extract_json":   types.SourceHTTPJSON,
			"extract_string": types.SourceHTTPBody,
			"request_uri":    types.SourceHTTPPath,
			"headers":        types.SourceHTTPHeader,
			// Poco methods
			"getURI":         types.SourceHTTPPath,
			"getMethod":      types.SourceHTTPHeader,
			"getContentType": types.SourceHTTPHeader,
			"stream":         types.SourceHTTPBody,
			// Standard stream methods
			"get":     types.SourceUserInput,
			"getline": types.SourceUserInput,
			"read":    types.SourceUserInput,
			"peek":    types.SourceUserInput,
			// Boost.Asio socket methods
			"read_some":     types.SourceNetwork,
			"receive":       types.SourceNetwork,
			"async_receive": types.SourceNetwork,
		},
	}
}

func registerJavaMappings() {
	Registry["java"] = &LanguageMappings{
		Language: "java",
		InputMethods: map[string]types.SourceType{
			// Servlet API
			"getParameter":       types.SourceHTTPGet,
			"getParameterValues": types.SourceHTTPGet,
			"getParameterMap":    types.SourceHTTPGet,
			"getParameterNames":  types.SourceHTTPGet,
			"getHeader":          types.SourceHTTPHeader,
			"getHeaders":         types.SourceHTTPHeader,
			"getHeaderNames":     types.SourceHTTPHeader,
			"getCookies":         types.SourceHTTPCookie,
			"getInputStream":     types.SourceHTTPBody,
			"getReader":          types.SourceHTTPBody,
			"getQueryString":     types.SourceHTTPGet,
			"getRequestURI":      types.SourceHTTPPath,
			"getRequestURL":      types.SourceHTTPPath,
			"getPathInfo":        types.SourceHTTPPath,
			"getServletPath":     types.SourceHTTPPath,
			"getRemoteAddr":      types.SourceHTTPHeader,
			// Spring MVC
			"getBody": types.SourceHTTPBody,
			// Vert.x
			"getParam":     types.SourceHTTPGet,
			"getParams":    types.SourceHTTPGet,
			"bodyAsString": types.SourceHTTPBody,
			"bodyAsJson":   types.SourceHTTPBody,
			// Environment/CLI
			"getenv":      types.SourceEnvVar,
			"getProperty": types.SourceEnvVar,
			// Scanner/Reader
			"nextLine":     types.SourceStdin,
			"next":         types.SourceStdin,
			"nextInt":      types.SourceStdin,
			"nextDouble":   types.SourceStdin,
			"nextBoolean":  types.SourceStdin,
			"readLine":     types.SourceFile,
			"readAllBytes": types.SourceFile,
			"readAllLines": types.SourceFile,
		},
	}
}

func registerCSharpMappings() {
	Registry["c_sharp"] = &LanguageMappings{
		Language: "c_sharp",
		InputSources: map[string]types.SourceType{
			"Console.ReadLine":                    types.SourceStdin,
			"Console.Read":                        types.SourceStdin,
			"Console.ReadKey":                     types.SourceStdin,
			"Environment.GetEnvironmentVariable":  types.SourceEnvVar,
			"Environment.GetCommandLineArgs":      types.SourceCLIArg,
			"Request.Query":                       types.SourceHTTPGet,
			"Request.Form":                        types.SourceHTTPPost,
			"Request.Body":                        types.SourceHTTPBody,
			"Request.Headers":                     types.SourceHTTPHeader,
			"Request.Cookies":                     types.SourceHTTPCookie,
			"HttpContext.Request":                 types.SourceUserInput,
			"File.ReadAllText":                    types.SourceFile,
			"File.ReadAllLines":                   types.SourceFile,
			"File.ReadAllBytes":                   types.SourceFile,
			"StreamReader.ReadLine":               types.SourceFile,
			"StreamReader.ReadToEnd":              types.SourceFile,
		},
	}
}

func registerRubyMappings() {
	Registry["ruby"] = &LanguageMappings{
		Language: "ruby",
		InputSources: map[string]types.SourceType{
			"gets":      types.SourceStdin,
			"readline":  types.SourceStdin,
			"readlines": types.SourceStdin,
			"STDIN":     types.SourceStdin,
			"ARGF":      types.SourceStdin,
			"ARGV":      types.SourceCLIArg,
			"ENV":       types.SourceEnvVar,
			"params":    types.SourceHTTPGet,
			"request":   types.SourceUserInput,
			"cookies":   types.SourceHTTPCookie,
			"session":   types.SourceUserInput,
			"File.read": types.SourceFile,
			"IO.read":   types.SourceFile,
		},
	}
}

func registerRustMappings() {
	Registry["rust"] = &LanguageMappings{
		Language: "rust",
		InputSources: map[string]types.SourceType{
			"env::args":      types.SourceCLIArg,
			"env::args_os":   types.SourceCLIArg,
			"env::var":       types.SourceEnvVar,
			"env::var_os":    types.SourceEnvVar,
			"stdin":          types.SourceStdin,
			"read_line":      types.SourceStdin,
			"BufRead":        types.SourceStdin,
			"fs::read":       types.SourceFile,
			"read_to_string": types.SourceFile,
			"File::open":     types.SourceFile,
			"web::Query":     types.SourceHTTPGet,
			"web::Form":      types.SourceHTTPPost,
			"web::Json":      types.SourceHTTPBody,
			"web::Path":      types.SourceHTTPGet,
			"Query<":         types.SourceHTTPGet,
			"Form<":          types.SourceHTTPPost,
			"Json<":          types.SourceHTTPBody,
			"Path<":          types.SourceHTTPGet,
		},
	}
}
