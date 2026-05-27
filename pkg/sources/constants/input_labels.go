// Package constants provides centralized type constants for the tracer.
package constants

import "github.com/hatlesswizard/inputtracer/pkg/sources/core"

// InputLabel categorizes the type of user input.
// This is a type alias — the canonical definition lives in pkg/sources/core.
type InputLabel = core.InputLabel

// Re-export InputLabel constants from core for backward compatibility.
const (
	LabelHTTPGet     = core.LabelHTTPGet
	LabelHTTPPost    = core.LabelHTTPPost
	LabelHTTPCookie  = core.LabelHTTPCookie
	LabelHTTPHeader  = core.LabelHTTPHeader
	LabelHTTPBody    = core.LabelHTTPBody
	LabelCLI         = core.LabelCLI
	LabelEnvironment = core.LabelEnvironment
	LabelFile        = core.LabelFile
	LabelDatabase    = core.LabelDatabase
	LabelNetwork     = core.LabelNetwork
	LabelUserInput   = core.LabelUserInput
)
