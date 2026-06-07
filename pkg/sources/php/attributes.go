package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// phpAttributeSourceDefinitions returns definitions for PHP 8 attributes that
// mark controller parameters as input sources. When a parameter has one of
// these attributes, the parameter itself IS the user input.
//
// Symfony 6.3+ introduced attribute-based request mapping:
//
//	#[MapRequestPayload] CreateUserDto $dto    → $dto IS the POST body
//	#[MapQueryString] SearchFilter $filter     → $filter IS the query params
//	#[MapQueryParameter] int $page             → $page IS a single query param
//	#[MapUploadedFile] UploadedFile $file      → $file IS the uploaded file
//	#[MapRequestHeader] string $accept         → $accept IS an HTTP header
func phpAttributeSourceDefinitions() []common.Definition {
	return []common.Definition{
		{
			Name:        "#[MapRequestPayload]",
			Pattern:     `#\[\s*MapRequestPayload\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Symfony MapRequestPayload — parameter is deserialized request body",
			NodeTypes:   []string{"attribute_list", "attribute"},
		},
		{
			Name:        "#[MapQueryString]",
			Pattern:     `#\[\s*MapQueryString\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "Symfony MapQueryString — parameter is mapped from query string",
			NodeTypes:   []string{"attribute_list", "attribute"},
		},
		{
			Name:        "#[MapQueryParameter]",
			Pattern:     `#\[\s*MapQueryParameter\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "Symfony MapQueryParameter — single query parameter mapped to argument",
			NodeTypes:   []string{"attribute_list", "attribute"},
		},
		{
			Name:        "#[MapUploadedFile]",
			Pattern:     `#\[\s*MapUploadedFile\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "Symfony MapUploadedFile — uploaded file mapped to parameter",
			NodeTypes:   []string{"attribute_list", "attribute"},
		},
		{
			Name:        "#[MapRequestHeader]",
			Pattern:     `#\[\s*MapRequestHeader\b`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Symfony MapRequestHeader — HTTP header mapped to parameter",
			NodeTypes:   []string{"attribute_list", "attribute"},
		},
	}
}
