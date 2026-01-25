// Package ruby - frameworks.go provides Ruby web framework patterns
// Includes patterns for Rails, Sinatra, Hanami, Grape, and Padrino
package ruby

import (
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// Registry is the global Ruby framework pattern registry
var Registry = common.NewFrameworkPatternRegistry("ruby")

// Ruby on Rails patterns (full-featured MVC framework)
var railsPatterns = []*common.FrameworkPattern{
	{
		ID:              "rails_params",
		Framework:       "rails",
		Language:        "ruby",
		Name:            "params[]",
		Description:     "Rails combined request parameters (GET, POST, route)",
		PropertyPattern: "^params$",
		SourceType:      common.SourceHTTPRequest,
		PopulatedFrom:   []string{"query_parameters", "request_parameters", "path_parameters"},
		Tags:            []string{"web", "mvc", "popular"},
	},
	{
		ID:            "rails_params_permit",
		Framework:     "rails",
		Language:      "ruby",
		Name:          "params.permit()",
		Description:   "Rails strong parameters permit",
		MethodPattern: "^permit$",
		CarrierClass:  "ActionController::Parameters",
		SourceType:    common.SourceHTTPRequest,
		Tags:          []string{"web", "mvc", "security"},
	},
	{
		ID:            "rails_params_require",
		Framework:     "rails",
		Language:      "ruby",
		Name:          "params.require()",
		Description:   "Rails strong parameters require",
		MethodPattern: "^require$",
		CarrierClass:  "ActionController::Parameters",
		SourceType:    common.SourceHTTPRequest,
		Tags:          []string{"web", "mvc", "security"},
	},
	{
		ID:              "rails_request_params",
		Framework:       "rails",
		Language:        "ruby",
		Name:            "request.params",
		Description:     "Rails request parameters hash",
		ClassPattern:    "^ActionDispatch::Request$",
		PropertyPattern: "^params$",
		SourceType:      common.SourceHTTPRequest,
		Tags:            []string{"web", "mvc"},
	},
	{
		ID:              "rails_request_body",
		Framework:       "rails",
		Language:        "ruby",
		Name:            "request.body",
		Description:     "Rails request body stream",
		ClassPattern:    "^ActionDispatch::Request$",
		PropertyPattern: "^body$",
		SourceType:      common.SourceHTTPBody,
		Tags:            []string{"web", "mvc"},
	},
	{
		ID:            "rails_request_raw_post",
		Framework:     "rails",
		Language:      "ruby",
		Name:          "request.raw_post",
		Description:   "Rails raw POST body",
		MethodPattern: "^raw_post$",
		SourceType:    common.SourceHTTPBody,
		Tags:          []string{"web", "mvc"},
	},
	{
		ID:              "rails_request_headers",
		Framework:       "rails",
		Language:        "ruby",
		Name:            "request.headers[]",
		Description:     "Rails request headers",
		ClassPattern:    "^ActionDispatch::Request$",
		PropertyPattern: "^headers$",
		SourceType:      common.SourceHTTPHeader,
		Tags:            []string{"web", "mvc"},
	},
	{
		ID:              "rails_cookies",
		Framework:       "rails",
		Language:        "ruby",
		Name:            "cookies[]",
		Description:     "Rails cookie jar",
		PropertyPattern: "^cookies$",
		SourceType:      common.SourceHTTPCookie,
		Tags:            []string{"web", "mvc"},
	},
	{
		ID:              "rails_session",
		Framework:       "rails",
		Language:        "ruby",
		Name:            "session[]",
		Description:     "Rails session store",
		PropertyPattern: "^session$",
		SourceType:      common.SourceSession,
		Tags:            []string{"web", "mvc"},
	},
	{
		ID:              "rails_request_env",
		Framework:       "rails",
		Language:        "ruby",
		Name:            "request.env",
		Description:     "Rails request environment (Rack env)",
		ClassPattern:    "^ActionDispatch::Request$",
		PropertyPattern: "^env$",
		SourceType:      common.SourceHTTPHeader,
		Tags:            []string{"web", "mvc", "rack"},
	},
}

// Sinatra patterns (lightweight DSL for web applications)
var sinatraPatterns = []*common.FrameworkPattern{
	{
		ID:              "sinatra_params",
		Framework:       "sinatra",
		Language:        "ruby",
		Name:            "params[]",
		Description:     "Sinatra combined request parameters",
		PropertyPattern: "^params$",
		SourceType:      common.SourceHTTPRequest,
		Tags:            []string{"web", "dsl", "lightweight"},
	},
	{
		ID:            "sinatra_request_body_read",
		Framework:     "sinatra",
		Language:      "ruby",
		Name:          "request.body.read",
		Description:   "Sinatra request body read",
		MethodPattern: "^read$",
		CarrierClass:  "Rack::Request",
		SourceType:    common.SourceHTTPBody,
		Tags:          []string{"web", "dsl", "lightweight"},
	},
	{
		ID:            "sinatra_env",
		Framework:     "sinatra",
		Language:      "ruby",
		Name:          "env[]",
		Description:   "Sinatra Rack environment",
		PropertyPattern: "^env$",
		SourceType:    common.SourceHTTPHeader,
		Tags:          []string{"web", "dsl", "rack"},
	},
	{
		ID:            "sinatra_halt",
		Framework:     "sinatra",
		Language:      "ruby",
		Name:          "halt",
		Description:   "Sinatra halt helper (takes user input for response)",
		MethodPattern: "^halt$",
		SourceType:    common.SourceUserInput,
		Tags:          []string{"web", "dsl"},
	},
}

// Hanami patterns (modern Ruby framework)
var hanamiPatterns = []*common.FrameworkPattern{
	{
		ID:              "hanami_params",
		Framework:       "hanami",
		Language:        "ruby",
		Name:            "params[]",
		Description:     "Hanami action parameters",
		PropertyPattern: "^params$",
		SourceType:      common.SourceHTTPRequest,
		Tags:            []string{"web", "modern", "clean"},
	},
	{
		ID:            "hanami_params_get",
		Framework:     "hanami",
		Language:      "ruby",
		Name:          "params.get()",
		Description:   "Hanami get parameter value",
		MethodPattern: "^get$",
		CarrierClass:  "Hanami::Action::Params",
		SourceType:    common.SourceHTTPRequest,
		Tags:          []string{"web", "modern"},
	},
	{
		ID:              "hanami_request_params",
		Framework:       "hanami",
		Language:        "ruby",
		Name:            "request.params",
		Description:     "Hanami request parameters",
		ClassPattern:    "^Hanami::Action::Request$",
		PropertyPattern: "^params$",
		SourceType:      common.SourceHTTPRequest,
		Tags:            []string{"web", "modern"},
	},
}

// Grape patterns (REST API framework)
var grapePatterns = []*common.FrameworkPattern{
	{
		ID:              "grape_params",
		Framework:       "grape",
		Language:        "ruby",
		Name:            "params[]",
		Description:     "Grape API parameters",
		PropertyPattern: "^params$",
		SourceType:      common.SourceHTTPRequest,
		Tags:            []string{"api", "rest"},
	},
	{
		ID:              "grape_declared_params",
		Framework:       "grape",
		Language:        "ruby",
		Name:            "declared(params)",
		Description:     "Grape declared parameters (filtered)",
		MethodPattern:   "^declared$",
		SourceType:      common.SourceHTTPRequest,
		Tags:            []string{"api", "rest", "filtered"},
	},
	{
		ID:            "grape_request_body",
		Framework:     "grape",
		Language:      "ruby",
		Name:          "request.body.read",
		Description:   "Grape request body",
		MethodPattern: "^read$",
		SourceType:    common.SourceHTTPBody,
		Tags:          []string{"api", "rest"},
	},
	{
		ID:              "grape_headers",
		Framework:       "grape",
		Language:        "ruby",
		Name:            "headers[]",
		Description:     "Grape request headers",
		PropertyPattern: "^headers$",
		SourceType:      common.SourceHTTPHeader,
		Tags:            []string{"api", "rest"},
	},
	{
		ID:            "grape_env",
		Framework:     "grape",
		Language:      "ruby",
		Name:          "env[]",
		Description:   "Grape Rack environment",
		PropertyPattern: "^env$",
		SourceType:    common.SourceHTTPHeader,
		Tags:          []string{"api", "rack"},
	},
}

// Padrino patterns (Rails-like modular framework built on Sinatra)
var padrinoPatterns = []*common.FrameworkPattern{
	{
		ID:              "padrino_params",
		Framework:       "padrino",
		Language:        "ruby",
		Name:            "params[]",
		Description:     "Padrino request parameters",
		PropertyPattern: "^params$",
		SourceType:      common.SourceHTTPRequest,
		Tags:            []string{"web", "modular", "sinatra"},
	},
	{
		ID:            "padrino_request_body",
		Framework:     "padrino",
		Language:      "ruby",
		Name:          "request.body.read",
		Description:   "Padrino request body",
		MethodPattern: "^read$",
		SourceType:    common.SourceHTTPBody,
		Tags:          []string{"web", "modular"},
	},
	{
		ID:              "padrino_flash",
		Framework:       "padrino",
		Language:        "ruby",
		Name:            "flash[]",
		Description:     "Padrino flash messages (user-originated data)",
		PropertyPattern: "^flash$",
		SourceType:      common.SourceSession,
		Tags:            []string{"web", "modular"},
	},
}

// Rack patterns (common interface used by most Ruby web frameworks)
var rackPatterns = []*common.FrameworkPattern{
	{
		ID:              "rack_request_params",
		Framework:       "rack",
		Language:        "ruby",
		Name:            "Rack::Request#params",
		Description:     "Rack combined query and POST parameters",
		ClassPattern:    "^Rack::Request$",
		PropertyPattern: "^params$",
		SourceType:      common.SourceHTTPRequest,
		Tags:            []string{"web", "interface", "core"},
	},
	{
		ID:            "rack_request_get",
		Framework:     "rack",
		Language:      "ruby",
		Name:          "Rack::Request#GET",
		Description:   "Rack query string parameters",
		MethodPattern: "^GET$",
		ClassPattern:  "^Rack::Request$",
		SourceType:    common.SourceHTTPGet,
		Tags:          []string{"web", "interface", "core"},
	},
	{
		ID:            "rack_request_post",
		Framework:     "rack",
		Language:      "ruby",
		Name:          "Rack::Request#POST",
		Description:   "Rack POST parameters",
		MethodPattern: "^POST$",
		ClassPattern:  "^Rack::Request$",
		SourceType:    common.SourceHTTPPost,
		Tags:          []string{"web", "interface", "core"},
	},
	{
		ID:              "rack_env",
		Framework:       "rack",
		Language:        "ruby",
		Name:            "env[]",
		Description:     "Rack environment hash (contains all request data)",
		PropertyPattern: "^env$",
		SourceType:      common.SourceHTTPRequest,
		Tags:            []string{"web", "interface", "core"},
	},
}

func init() {
	Registry.RegisterAll(railsPatterns)
	Registry.RegisterAll(sinatraPatterns)
	Registry.RegisterAll(hanamiPatterns)
	Registry.RegisterAll(grapePatterns)
	Registry.RegisterAll(padrinoPatterns)
	Registry.RegisterAll(rackPatterns)

	// Register framework detectors
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "rails",
		Indicators: []string{"config/application.rb", "config/routes.rb", "app/controllers"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "sinatra",
		Indicators: []string{"Gemfile", "config.ru"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "hanami",
		Indicators: []string{"config/environment.rb", "lib/", "spec/web"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "grape",
		Indicators: []string{"Gemfile", "api/"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "padrino",
		Indicators: []string{"config/boot.rb", "config/apps.rb"},
	})
}
