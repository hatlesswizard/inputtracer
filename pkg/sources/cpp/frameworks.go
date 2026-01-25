// Package cpp - frameworks.go provides C++ web framework patterns
// Includes patterns for Crow, Drogon, Boost.Beast, cpprestsdk, Poco, and Qt
package cpp

import (
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// Registry is the global C++ framework pattern registry
var Registry = common.NewFrameworkPatternRegistry("cpp")

// Crow framework patterns (lightweight C++ web framework)
var crowPatterns = []*common.FrameworkPattern{
	{
		ID:            "crow_request",
		Framework:     "crow",
		Language:      "cpp",
		Name:          "crow::request",
		Description:   "Crow HTTP request object",
		ClassPattern:  "^crow::request$",
		SourceType:    common.SourceHTTPRequest,
		Confidence:    1.0,
		Tags:          []string{"web", "framework"},
	},
	{
		ID:            "crow_response",
		Framework:     "crow",
		Language:      "cpp",
		Name:          "crow::response",
		Description:   "Crow HTTP response object",
		ClassPattern:  "^crow::response$",
		SourceType:    common.SourceHTTPRequest,
		Confidence:    1.0,
		Tags:          []string{"web", "framework"},
	},
	{
		ID:            "crow_route",
		Framework:     "crow",
		Language:      "cpp",
		Name:          "CROW_ROUTE",
		Description:   "Crow route macro",
		MethodPattern: "^CROW_ROUTE$",
		SourceType:    common.SourceHTTPPath,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "routing"},
	},
	{
		ID:            "crow_url_params",
		Framework:     "crow",
		Language:      "cpp",
		Name:          "req.url_params",
		Description:   "Crow URL query parameters",
		PropertyPattern: "^url_params$",
		SourceType:    common.SourceHTTPGet,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "query"},
	},
	{
		ID:            "crow_body",
		Framework:     "crow",
		Language:      "cpp",
		Name:          "req.body",
		Description:   "Crow request body",
		PropertyPattern: "^body$",
		CarrierClass:  "crow::request",
		SourceType:    common.SourceHTTPBody,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "body"},
	},
}

// Drogon framework patterns (high-performance C++ web framework)
var drogonPatterns = []*common.FrameworkPattern{
	{
		ID:            "drogon_http_request_ptr",
		Framework:     "drogon",
		Language:      "cpp",
		Name:          "HttpRequestPtr",
		Description:   "Drogon HTTP request pointer",
		ClassPattern:  "^HttpRequestPtr$",
		SourceType:    common.SourceHTTPRequest,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "async"},
	},
	{
		ID:            "drogon_http_response_ptr",
		Framework:     "drogon",
		Language:      "cpp",
		Name:          "HttpResponsePtr",
		Description:   "Drogon HTTP response pointer",
		ClassPattern:  "^HttpResponsePtr$",
		SourceType:    common.SourceHTTPRequest,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "async"},
	},
	{
		ID:            "drogon_app",
		Framework:     "drogon",
		Language:      "cpp",
		Name:          "drogon::app",
		Description:   "Drogon application singleton",
		MethodPattern: "^drogon::app$",
		SourceType:    common.SourceHTTPRequest,
		Confidence:    0.8,
		Tags:          []string{"web", "framework"},
	},
	{
		ID:            "drogon_get_parameter",
		Framework:     "drogon",
		Language:      "cpp",
		Name:          "getParameter()",
		Description:   "Drogon get request parameter",
		MethodPattern: "^getParameter$",
		SourceType:    common.SourceHTTPGet,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "query"},
	},
	{
		ID:            "drogon_get_body",
		Framework:     "drogon",
		Language:      "cpp",
		Name:          "getBody()",
		Description:   "Drogon get request body",
		MethodPattern: "^getBody$",
		CarrierClass:  "HttpRequestPtr",
		SourceType:    common.SourceHTTPBody,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "body"},
	},
	{
		ID:            "drogon_get_json",
		Framework:     "drogon",
		Language:      "cpp",
		Name:          "getJsonObject()",
		Description:   "Drogon get JSON from request",
		MethodPattern: "^getJsonObject$",
		CarrierClass:  "HttpRequestPtr",
		SourceType:    common.SourceHTTPJSON,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "json"},
	},
}

// Boost.Beast patterns (HTTP/WebSocket library)
var boostBeastPatterns = []*common.FrameworkPattern{
	{
		ID:            "beast_http_request",
		Framework:     "boost.beast",
		Language:      "cpp",
		Name:          "beast::http::request",
		Description:   "Boost.Beast HTTP request",
		ClassPattern:  "^beast::http::request",
		SourceType:    common.SourceHTTPRequest,
		Confidence:    1.0,
		Tags:          []string{"web", "library", "boost"},
	},
	{
		ID:            "beast_websocket_stream",
		Framework:     "boost.beast",
		Language:      "cpp",
		Name:          "websocket::stream",
		Description:   "Boost.Beast WebSocket stream",
		ClassPattern:  "^websocket::stream",
		SourceType:    common.SourceNetwork,
		Confidence:    1.0,
		Tags:          []string{"websocket", "library", "boost"},
	},
	{
		ID:            "beast_http_body",
		Framework:     "boost.beast",
		Language:      "cpp",
		Name:          "request.body()",
		Description:   "Boost.Beast request body",
		MethodPattern: "^body$",
		SourceType:    common.SourceHTTPBody,
		Confidence:    0.9,
		Tags:          []string{"web", "library", "boost"},
	},
	{
		ID:            "beast_http_target",
		Framework:     "boost.beast",
		Language:      "cpp",
		Name:          "request.target()",
		Description:   "Boost.Beast request target (path)",
		MethodPattern: "^target$",
		SourceType:    common.SourceHTTPPath,
		Confidence:    0.9,
		Tags:          []string{"web", "library", "boost"},
	},
}

// cpprestsdk (Casablanca) patterns (Microsoft's C++ REST SDK)
var cpprestsdkPatterns = []*common.FrameworkPattern{
	{
		ID:            "cpprest_http_request",
		Framework:     "cpprestsdk",
		Language:      "cpp",
		Name:          "web::http::http_request",
		Description:   "cpprestsdk HTTP request",
		ClassPattern:  "^(web::)?http::http_request$",
		SourceType:    common.SourceHTTPRequest,
		Confidence:    1.0,
		Tags:          []string{"web", "microsoft", "rest"},
	},
	{
		ID:            "cpprest_http_response",
		Framework:     "cpprestsdk",
		Language:      "cpp",
		Name:          "web::http::http_response",
		Description:   "cpprestsdk HTTP response",
		ClassPattern:  "^(web::)?http::http_response$",
		SourceType:    common.SourceHTTPRequest,
		Confidence:    1.0,
		Tags:          []string{"web", "microsoft", "rest"},
	},
	{
		ID:            "cpprest_request_uri",
		Framework:     "cpprestsdk",
		Language:      "cpp",
		Name:          "request.request_uri()",
		Description:   "cpprestsdk request URI",
		MethodPattern: "^request_uri$",
		SourceType:    common.SourceHTTPPath,
		Confidence:    1.0,
		Tags:          []string{"web", "microsoft", "rest"},
	},
	{
		ID:            "cpprest_extract_json",
		Framework:     "cpprestsdk",
		Language:      "cpp",
		Name:          "request.extract_json()",
		Description:   "cpprestsdk extract JSON from request",
		MethodPattern: "^extract_json$",
		SourceType:    common.SourceHTTPJSON,
		Confidence:    1.0,
		Tags:          []string{"web", "microsoft", "json"},
	},
	{
		ID:            "cpprest_extract_string",
		Framework:     "cpprestsdk",
		Language:      "cpp",
		Name:          "request.extract_string()",
		Description:   "cpprestsdk extract string from request body",
		MethodPattern: "^extract_string$",
		SourceType:    common.SourceHTTPBody,
		Confidence:    1.0,
		Tags:          []string{"web", "microsoft", "body"},
	},
}

// Poco framework patterns (C++ portable components)
var pocoPatterns = []*common.FrameworkPattern{
	{
		ID:            "poco_http_server_request",
		Framework:     "poco",
		Language:      "cpp",
		Name:          "HTTPServerRequest",
		Description:   "Poco HTTP server request",
		ClassPattern:  "^(Poco::Net::)?HTTPServerRequest$",
		SourceType:    common.SourceHTTPRequest,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "portable"},
	},
	{
		ID:            "poco_http_server_response",
		Framework:     "poco",
		Language:      "cpp",
		Name:          "HTTPServerResponse",
		Description:   "Poco HTTP server response",
		ClassPattern:  "^(Poco::Net::)?HTTPServerResponse$",
		SourceType:    common.SourceHTTPRequest,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "portable"},
	},
	{
		ID:            "poco_get_uri",
		Framework:     "poco",
		Language:      "cpp",
		Name:          "getURI()",
		Description:   "Poco get request URI",
		MethodPattern: "^getURI$",
		CarrierClass:  "HTTPServerRequest",
		SourceType:    common.SourceHTTPPath,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "uri"},
	},
	{
		ID:            "poco_stream",
		Framework:     "poco",
		Language:      "cpp",
		Name:          "request.stream()",
		Description:   "Poco request input stream",
		MethodPattern: "^stream$",
		CarrierClass:  "HTTPServerRequest",
		SourceType:    common.SourceHTTPBody,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "body"},
	},
	{
		ID:            "poco_get_host",
		Framework:     "poco",
		Language:      "cpp",
		Name:          "getHost()",
		Description:   "Poco get request host header",
		MethodPattern: "^getHost$",
		SourceType:    common.SourceHTTPHeader,
		Confidence:    1.0,
		Tags:          []string{"web", "framework", "header"},
	},
}

// Qt framework patterns (network and web related)
var qtPatterns = []*common.FrameworkPattern{
	{
		ID:            "qt_network_reply",
		Framework:     "qt",
		Language:      "cpp",
		Name:          "QNetworkReply",
		Description:   "Qt network reply object",
		ClassPattern:  "^QNetworkReply$",
		SourceType:    common.SourceNetwork,
		Confidence:    1.0,
		Tags:          []string{"qt", "network"},
	},
	{
		ID:            "qt_network_request",
		Framework:     "qt",
		Language:      "cpp",
		Name:          "QNetworkRequest",
		Description:   "Qt network request object",
		ClassPattern:  "^QNetworkRequest$",
		SourceType:    common.SourceNetwork,
		Confidence:    0.9,
		Tags:          []string{"qt", "network"},
	},
	{
		ID:            "qt_url_query",
		Framework:     "qt",
		Language:      "cpp",
		Name:          "QUrlQuery",
		Description:   "Qt URL query parser",
		ClassPattern:  "^QUrlQuery$",
		SourceType:    common.SourceHTTPGet,
		Confidence:    0.9,
		Tags:          []string{"qt", "url", "query"},
	},
	{
		ID:            "qt_read_all",
		Framework:     "qt",
		Language:      "cpp",
		Name:          "readAll()",
		Description:   "Qt read all data from device/reply",
		MethodPattern: "^readAll$",
		SourceType:    common.SourceNetwork,
		Confidence:    0.85,
		Tags:          []string{"qt", "network", "file"},
	},
	{
		ID:            "qt_read_line",
		Framework:     "qt",
		Language:      "cpp",
		Name:          "readLine()",
		Description:   "Qt read line from device",
		MethodPattern: "^readLine$",
		SourceType:    common.SourceUserInput,
		Confidence:    0.8,
		Tags:          []string{"qt", "file", "input"},
	},
	{
		ID:            "qt_process_stdout",
		Framework:     "qt",
		Language:      "cpp",
		Name:          "readAllStandardOutput()",
		Description:   "Qt read process stdout",
		MethodPattern: "^readAllStandardOutput$",
		SourceType:    common.SourceUserInput,
		Confidence:    0.9,
		Tags:          []string{"qt", "process"},
	},
}

func init() {
	Registry.RegisterAll(crowPatterns)
	Registry.RegisterAll(drogonPatterns)
	Registry.RegisterAll(boostBeastPatterns)
	Registry.RegisterAll(cpprestsdkPatterns)
	Registry.RegisterAll(pocoPatterns)
	Registry.RegisterAll(qtPatterns)

	// Register framework detectors
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "crow",
		Indicators: []string{"crow.h", "crow/app.h", "crow_all.h"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "drogon",
		Indicators: []string{"drogon/drogon.h", "drogon/HttpController.h"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "boost.beast",
		Indicators: []string{"boost/beast.hpp", "boost/beast/http.hpp"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "cpprestsdk",
		Indicators: []string{"cpprest/http_listener.h", "cpprest/http_client.h"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "poco",
		Indicators: []string{"Poco/Net/HTTPServer.h", "Poco/Net/HTTPRequestHandler.h"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "qt",
		Indicators: []string{"QNetworkAccessManager", "QtNetwork/QNetworkReply"},
	})
}
