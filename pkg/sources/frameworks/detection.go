// Package frameworks - detection.go provides framework detection utilities
// This centralizes file path indicators used to detect frameworks in codebases
package frameworks

import (
	"os"
	"path/filepath"

	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// FrameworkIndicator defines file paths that indicate a specific framework
type FrameworkIndicator struct {
	Framework   string   // Framework identifier
	Language    string   // Programming language
	Indicators  []string // File paths relative to project root
	Description string   // Human-readable description
}

// PHPFrameworkIndicators contains file path indicators for PHP frameworks
var PHPFrameworkIndicators = []FrameworkIndicator{
	{
		Framework:   "mybb",
		Language:    "php",
		Indicators:  []string{"inc/class_core.php", "inc/init.php"},
		Description: "MyBB forum software",
	},
	{
		Framework:   "wordpress",
		Language:    "php",
		Indicators:  []string{"wp-config.php", "wp-includes/version.php"},
		Description: "WordPress CMS",
	},
	{
		Framework:   "laravel",
		Language:    "php",
		Indicators:  []string{"artisan", "bootstrap/app.php"},
		Description: "Laravel framework",
	},
	{
		Framework:   "symfony",
		Language:    "php",
		Indicators:  []string{"symfony.lock", "config/bundles.php"},
		Description: "Symfony framework",
	},
	{
		Framework:   "codeigniter",
		Language:    "php",
		Indicators:  []string{"system/core/CodeIgniter.php"},
		Description: "CodeIgniter framework",
	},
	{
		Framework:   "drupal",
		Language:    "php",
		Indicators:  []string{"core/includes/bootstrap.inc"},
		Description: "Drupal CMS",
	},
	{
		Framework:   "yii2",
		Language:    "php",
		Indicators:  []string{"yii"},
		Description: "Yii 2 framework",
	},
	{
		Framework:   "cakephp",
		Language:    "php",
		Indicators:  []string{"config/app.php", "src/Application.php"},
		Description: "CakePHP framework",
	},
	{
		Framework:   "phpbb",
		Language:    "php",
		Indicators:  []string{"phpbb/request/request.php"},
		Description: "phpBB forum software",
	},
	{
		Framework:   "prestashop",
		Language:    "php",
		Indicators:  []string{"config/defines.inc.php"},
		Description: "PrestaShop e-commerce",
	},
}

// RubyFrameworkIndicators contains file path indicators for Ruby frameworks
var RubyFrameworkIndicators = []FrameworkIndicator{
	{
		Framework:   "rails",
		Language:    "ruby",
		Indicators:  []string{"config/application.rb", "config/routes.rb", "app/controllers"},
		Description: "Ruby on Rails",
	},
	{
		Framework:   "sinatra",
		Language:    "ruby",
		Indicators:  []string{"config.ru"},
		Description: "Sinatra web framework",
	},
	{
		Framework:   "hanami",
		Language:    "ruby",
		Indicators:  []string{"config/environment.rb", "lib/"},
		Description: "Hanami framework",
	},
	{
		Framework:   "padrino",
		Language:    "ruby",
		Indicators:  []string{"config/boot.rb", "config/apps.rb"},
		Description: "Padrino framework",
	},
}

// JavaScriptFrameworkIndicators contains file path indicators for JavaScript frameworks
var JavaScriptFrameworkIndicators = []FrameworkIndicator{
	{
		Framework:   "express",
		Language:    "javascript",
		Indicators:  []string{"node_modules/express", "package.json"},
		Description: "Express.js web framework",
	},
	{
		Framework:   "nextjs",
		Language:    "javascript",
		Indicators:  []string{"next.config.js", "next.config.mjs", "pages/", "app/"},
		Description: "Next.js React framework",
	},
	{
		Framework:   "nuxt",
		Language:    "javascript",
		Indicators:  []string{"nuxt.config.js", "nuxt.config.ts"},
		Description: "Nuxt.js Vue framework",
	},
	{
		Framework:   "nestjs",
		Language:    "typescript",
		Indicators:  []string{"nest-cli.json", "src/main.ts"},
		Description: "NestJS framework",
	},
	{
		Framework:   "koa",
		Language:    "javascript",
		Indicators:  []string{"node_modules/koa"},
		Description: "Koa web framework",
	},
	{
		Framework:   "fastify",
		Language:    "javascript",
		Indicators:  []string{"node_modules/fastify"},
		Description: "Fastify web framework",
	},
}

// PythonFrameworkIndicators contains file path indicators for Python frameworks
var PythonFrameworkIndicators = []FrameworkIndicator{
	{
		Framework:   "django",
		Language:    "python",
		Indicators:  []string{"manage.py", "settings.py", "urls.py"},
		Description: "Django web framework",
	},
	{
		Framework:   "flask",
		Language:    "python",
		Indicators:  []string{"app.py", "wsgi.py"},
		Description: "Flask web framework",
	},
	{
		Framework:   "fastapi",
		Language:    "python",
		Indicators:  []string{"main.py"},
		Description: "FastAPI web framework",
	},
}

// JavaFrameworkIndicators contains file path indicators for Java frameworks
var JavaFrameworkIndicators = []FrameworkIndicator{
	{
		Framework:   "spring",
		Language:    "java",
		Indicators:  []string{"pom.xml", "application.properties", "application.yml"},
		Description: "Spring Framework",
	},
	{
		Framework:   "springboot",
		Language:    "java",
		Indicators:  []string{"src/main/java", "Application.java"},
		Description: "Spring Boot",
	},
}

// GoFrameworkIndicators contains file path indicators for Go frameworks
var GoFrameworkIndicators = []FrameworkIndicator{
	{
		Framework:   "gin",
		Language:    "go",
		Indicators:  []string{"go.mod", "main.go"},
		Description: "Gin web framework",
	},
	{
		Framework:   "echo",
		Language:    "go",
		Indicators:  []string{"go.mod"},
		Description: "Echo web framework",
	},
}

// CSharpFrameworkIndicators contains file path indicators for C# frameworks
var CSharpFrameworkIndicators = []FrameworkIndicator{
	{
		Framework:   "aspnetcore",
		Language:    "csharp",
		Indicators:  []string{"Program.cs", "Startup.cs", ".csproj"},
		Description: "ASP.NET Core",
	},
	{
		Framework:   "aspnetmvc",
		Language:    "csharp",
		Indicators:  []string{"Global.asax.cs", "Web.config"},
		Description: "ASP.NET MVC",
	},
}

// RustFrameworkIndicators contains file path indicators for Rust frameworks
var RustFrameworkIndicators = []FrameworkIndicator{
	{
		Framework:   "actix-web",
		Language:    "rust",
		Indicators:  []string{"Cargo.toml"},
		Description: "Actix-web framework",
	},
	{
		Framework:   "rocket",
		Language:    "rust",
		Indicators:  []string{"Cargo.toml", "Rocket.toml"},
		Description: "Rocket framework",
	},
	{
		Framework:   "axum",
		Language:    "rust",
		Indicators:  []string{"Cargo.toml"},
		Description: "Axum framework",
	},
}

// CppFrameworkIndicators contains file path indicators for C++ frameworks
var CppFrameworkIndicators = []FrameworkIndicator{
	{
		Framework:   "qt",
		Language:    "cpp",
		Indicators:  []string{".pro", "CMakeLists.txt"},
		Description: "Qt framework",
	},
	{
		Framework:   "poco",
		Language:    "cpp",
		Indicators:  []string{"CMakeLists.txt"},
		Description: "POCO C++ Libraries",
	},
}

// AllFrameworkIndicators contains all framework indicators across all languages
var AllFrameworkIndicators = func() []FrameworkIndicator {
	all := make([]FrameworkIndicator, 0)
	all = append(all, PHPFrameworkIndicators...)
	all = append(all, RubyFrameworkIndicators...)
	all = append(all, JavaScriptFrameworkIndicators...)
	all = append(all, PythonFrameworkIndicators...)
	all = append(all, JavaFrameworkIndicators...)
	all = append(all, GoFrameworkIndicators...)
	all = append(all, CSharpFrameworkIndicators...)
	all = append(all, RustFrameworkIndicators...)
	all = append(all, CppFrameworkIndicators...)
	return all
}()

// DetectFramework detects which framework is being used in a codebase
// Returns the framework name or "unknown" if not detected
func DetectFramework(codebasePath string) string {
	for _, indicator := range AllFrameworkIndicators {
		for _, path := range indicator.Indicators {
			checkPath := filepath.Join(codebasePath, path)
			if _, err := os.Stat(checkPath); err == nil {
				return indicator.Framework
			}
		}
	}
	return "unknown"
}

// DetectFrameworkByLanguage detects which framework is being used for a specific language
func DetectFrameworkByLanguage(codebasePath string, language string) string {
	var indicators []FrameworkIndicator

	switch language {
	case "php":
		indicators = PHPFrameworkIndicators
	case "ruby":
		indicators = RubyFrameworkIndicators
	case "javascript", "typescript":
		indicators = JavaScriptFrameworkIndicators
	case "python":
		indicators = PythonFrameworkIndicators
	case "java":
		indicators = JavaFrameworkIndicators
	case "go":
		indicators = GoFrameworkIndicators
	case "c_sharp", "csharp":
		indicators = CSharpFrameworkIndicators
	case "rust":
		indicators = RustFrameworkIndicators
	case "cpp", "c++":
		indicators = CppFrameworkIndicators
	default:
		indicators = AllFrameworkIndicators
	}

	for _, indicator := range indicators {
		for _, path := range indicator.Indicators {
			checkPath := filepath.Join(codebasePath, path)
			if _, err := os.Stat(checkPath); err == nil {
				return indicator.Framework
			}
		}
	}
	return "unknown"
}

// GetFrameworkIndicators returns all indicators for a specific framework
func GetFrameworkIndicators(framework string) []string {
	for _, indicator := range AllFrameworkIndicators {
		if indicator.Framework == framework {
			return indicator.Indicators
		}
	}
	return nil
}

// GetFrameworksForLanguage returns all known frameworks for a language
func GetFrameworksForLanguage(language string) []string {
	var indicators []FrameworkIndicator

	switch language {
	case "php":
		indicators = PHPFrameworkIndicators
	case "ruby":
		indicators = RubyFrameworkIndicators
	case "javascript", "typescript":
		indicators = JavaScriptFrameworkIndicators
	case "python":
		indicators = PythonFrameworkIndicators
	case "java":
		indicators = JavaFrameworkIndicators
	case "go":
		indicators = GoFrameworkIndicators
	case "c_sharp", "csharp":
		indicators = CSharpFrameworkIndicators
	case "rust":
		indicators = RustFrameworkIndicators
	case "cpp", "c++":
		indicators = CppFrameworkIndicators
	default:
		return nil
	}

	frameworks := make([]string, len(indicators))
	for i, ind := range indicators {
		frameworks[i] = ind.Framework
	}
	return frameworks
}

// RegisterAllDetectors registers all framework detectors with the common package
func RegisterAllDetectors() {
	for _, indicator := range AllFrameworkIndicators {
		common.RegisterFrameworkDetector(&common.FrameworkDetector{
			Framework:  indicator.Framework,
			Indicators: indicator.Indicators,
		})
	}
}

func init() {
	// Register all detectors on package load
	RegisterAllDetectors()
}
