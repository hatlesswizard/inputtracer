// Package main - frameworks.go defines framework source URLs and method mappings
package main

// FrameworkSource defines a source file to fetch from GitHub
type FrameworkSource struct {
	URL       string
	ClassName string
}

// FrameworkDefinition defines a framework's sources and mapping config
type FrameworkDefinition struct {
	Name            string
	Language        string
	Sources         []FrameworkSource
	CarrierClass    string
	ClassPattern    string
	Tags            []string
	FrameworkDetect []string
}

// MethodMapping maps a method name to its source type
type MethodMapping struct {
	SourceType    string
	Description   string
	PopulatedFrom []string
	IsProperty    bool
}

// Frameworks defines all supported frameworks
var Frameworks = map[string]*FrameworkDefinition{
	"laravel": {
		Name:            "laravel",
		Language:        "php",
		CarrierClass:    "Illuminate\\Http\\Request",
		ClassPattern:    "^(Illuminate\\\\\\\\Http\\\\\\\\)?Request$",
		Tags:            []string{"framework", "modern", "generated"},
		FrameworkDetect: []string{"artisan", "bootstrap/app.php", "config/app.php"},
		Sources: []FrameworkSource{
			{URL: "https://raw.githubusercontent.com/illuminate/http/master/Concerns/InteractsWithInput.php", ClassName: "InteractsWithInput"},
			{URL: "https://raw.githubusercontent.com/illuminate/http/master/Request.php", ClassName: "Request"},
			{URL: "https://raw.githubusercontent.com/illuminate/http/master/Concerns/InteractsWithFlashData.php", ClassName: "InteractsWithFlashData"},
		},
	},
	"symfony": {
		Name:            "symfony",
		Language:        "php",
		CarrierClass:    "Symfony\\Component\\HttpFoundation\\Request",
		ClassPattern:    "^(Symfony\\\\\\\\Component\\\\\\\\HttpFoundation\\\\\\\\)?Request$",
		Tags:            []string{"framework", "enterprise", "generated"},
		FrameworkDetect: []string{"symfony.lock", "config/bundles.php", "src/Kernel.php"},
		Sources: []FrameworkSource{
			{URL: "https://raw.githubusercontent.com/symfony/http-foundation/7.3/ParameterBag.php", ClassName: "ParameterBag"},
			{URL: "https://raw.githubusercontent.com/symfony/http-foundation/7.3/InputBag.php", ClassName: "InputBag"},
			{URL: "https://raw.githubusercontent.com/symfony/http-foundation/7.3/Request.php", ClassName: "Request"},
		},
	},
}

// SymfonyPropertyMappings maps Symfony Request public properties
var SymfonyPropertyMappings = map[string]*MethodMapping{
	"query":      {SourceType: "SourceHTTPGet", Description: "Symfony request query bag contains GET parameters", PopulatedFrom: []string{"$_GET"}, IsProperty: true},
	"request":    {SourceType: "SourceHTTPPost", Description: "Symfony request bag contains POST parameters", PopulatedFrom: []string{"$_POST"}, IsProperty: true},
	"cookies":    {SourceType: "SourceHTTPCookie", Description: "Symfony cookies bag contains cookie values", PopulatedFrom: []string{"$_COOKIE"}, IsProperty: true},
	"headers":    {SourceType: "SourceHTTPHeader", Description: "Symfony headers bag contains HTTP headers", PopulatedFrom: []string{"$_SERVER"}, IsProperty: true},
	"files":      {SourceType: "SourceHTTPFile", Description: "Symfony files bag contains uploaded files", PopulatedFrom: []string{"$_FILES"}, IsProperty: true},
	"server":     {SourceType: "SourceHTTPHeader", Description: "Symfony server bag contains server parameters", PopulatedFrom: []string{"$_SERVER"}, IsProperty: true},
	"attributes": {SourceType: "SourceHTTPPath", Description: "Symfony attributes bag (route parameters, etc.)", PopulatedFrom: []string{}, IsProperty: true},
}
