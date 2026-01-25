// Package javascript - nestjs.go provides NestJS framework input patterns
package javascript

import (
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

var nestjsPatterns = []*common.FrameworkPattern{
	{
		ID:            "nestjs_body",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @Body()",
		Description:   "NestJS body decorator for request body",
		MethodPattern: "^Body$",
		SourceType:    common.SourceHTTPBody,
		PopulatedFrom: []string{"HTTP body"},
		Tags:          []string{"framework", "nestjs", "decorator"},
	},
	{
		ID:            "nestjs_query",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @Query()",
		Description:   "NestJS query decorator for query parameters",
		MethodPattern: "^Query$",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"query string"},
		Tags:          []string{"framework", "nestjs", "decorator"},
	},
	{
		ID:            "nestjs_param",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @Param()",
		Description:   "NestJS param decorator for URL path parameters",
		MethodPattern: "^Param$",
		SourceType:    common.SourceHTTPPath,
		PopulatedFrom: []string{"URL path"},
		Tags:          []string{"framework", "nestjs", "decorator"},
	},
	{
		ID:            "nestjs_headers",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @Headers()",
		Description:   "NestJS headers decorator for HTTP headers",
		MethodPattern: "^Headers$",
		SourceType:    common.SourceHTTPHeader,
		PopulatedFrom: []string{"HTTP headers"},
		Tags:          []string{"framework", "nestjs", "decorator"},
	},
	{
		ID:            "nestjs_ip",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @Ip()",
		Description:   "NestJS IP decorator for client IP address",
		MethodPattern: "^Ip$",
		SourceType:    common.SourceNetwork,
		PopulatedFrom: []string{"TCP connection", "X-Forwarded-For header"},
		Tags:          []string{"framework", "nestjs", "decorator"},
	},
	{
		ID:            "nestjs_host_param",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @HostParam()",
		Description:   "NestJS host param decorator for host parameters",
		MethodPattern: "^HostParam$",
		SourceType:    common.SourceHTTPHeader,
		PopulatedFrom: []string{"HTTP Host header"},
		Tags:          []string{"framework", "nestjs", "decorator"},
	},
	{
		ID:            "nestjs_session",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @Session()",
		Description:   "NestJS session decorator for session data",
		MethodPattern: "^Session$",
		SourceType:    common.SourceSession,
		PopulatedFrom: []string{"session storage"},
		Tags:          []string{"framework", "nestjs", "decorator"},
	},
	{
		ID:            "nestjs_uploaded_file",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @UploadedFile()",
		Description:   "NestJS uploaded file decorator for single file upload",
		MethodPattern: "^UploadedFile$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"HTTP multipart form data"},
		Tags:          []string{"framework", "nestjs", "decorator", "file"},
	},
	{
		ID:            "nestjs_uploaded_files",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @UploadedFiles()",
		Description:   "NestJS uploaded files decorator for multiple file uploads",
		MethodPattern: "^UploadedFiles$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"HTTP multipart form data"},
		Tags:          []string{"framework", "nestjs", "decorator", "file"},
	},
	{
		ID:            "nestjs_req",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @Req()",
		Description:   "NestJS request decorator for full request object",
		MethodPattern: "^Req$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"HTTP request"},
		Tags:          []string{"framework", "nestjs", "decorator"},
	},
	{
		ID:            "nestjs_request",
		Framework:     "nestjs",
		Language:      "typescript",
		Name:          "NestJS @Request()",
		Description:   "NestJS request decorator (alias for @Req)",
		MethodPattern: "^Request$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"HTTP request"},
		Tags:          []string{"framework", "nestjs", "decorator"},
	},
}

func init() {
	Registry.RegisterAll(nestjsPatterns)

	// Register NestJS framework detector
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "nestjs",
		Indicators: []string{"nest-cli.json", "@nestjs/core", "main.ts", "app.module.ts"},
	})
}
