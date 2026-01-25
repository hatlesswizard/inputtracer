// Package php - mybb.go provides MyBB forum software input patterns
package php

import (
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

var mybbPatterns = []*common.FrameworkPattern{
	{
		ID:              "mybb_input",
		Framework:       "mybb",
		Language:        "php",
		Name:            "MyBB $mybb->input",
		Description:     "MyBB input array populated from $_GET and $_POST via parse_incoming()",
		ClassPattern:    "^MyBB$",
		PropertyPattern: "^input$",
		AccessPattern:   "array",
		SourceType:      common.SourceHTTPGet, // Actually GET+POST
		CarrierClass:    "MyBB",
		CarrierProperty: "input",
		PopulatedBy:     "__construct",
		PopulatedFrom:   []string{"$_GET", "$_POST"},
		Confidence:      1.0,
		Tags:            []string{"forum", "bulletin-board"},
	},
	{
		ID:              "mybb_cookies",
		Framework:       "mybb",
		Language:        "php",
		Name:            "MyBB $mybb->cookies",
		Description:     "MyBB cookies array populated from $_COOKIE via parse_cookies()",
		ClassPattern:    "^MyBB$",
		PropertyPattern: "^cookies$",
		AccessPattern:   "array",
		SourceType:      common.SourceHTTPCookie,
		CarrierClass:    "MyBB",
		CarrierProperty: "cookies",
		PopulatedBy:     "parse_cookies",
		PopulatedFrom:   []string{"$_COOKIE"},
		Confidence:      1.0,
		Tags:            []string{"forum", "bulletin-board"},
	},
	{
		ID:              "mybb_get_input",
		Framework:       "mybb",
		Language:        "php",
		Name:            "MyBB $mybb->get_input()",
		Description:     "MyBB method to retrieve sanitized input from the input array",
		ClassPattern:    "^MyBB$",
		MethodPattern:   "^get_input$",
		SourceType:      common.SourceHTTPGet,
		CarrierClass:    "MyBB",
		PopulatedFrom:   []string{"$_GET", "$_POST"},
		Confidence:      1.0,
		Tags:            []string{"forum", "bulletin-board"},
	},
}

func init() {
	Registry.RegisterAll(mybbPatterns)

	// Register framework detector
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "mybb",
		Indicators: []string{"inc/class_core.php", "inc/init.php", "inc/class_parser.php"},
	})
}
