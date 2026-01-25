package sources

import (
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
	sitter "github.com/smacker/go-tree-sitter"
)

// Re-export types from common package for backwards compatibility
type InputLabel = common.InputLabel

const (
	LabelHTTPGet     = common.LabelHTTPGet
	LabelHTTPPost    = common.LabelHTTPPost
	LabelHTTPCookie  = common.LabelHTTPCookie
	LabelHTTPHeader  = common.LabelHTTPHeader
	LabelHTTPBody    = common.LabelHTTPBody
	LabelCLI         = common.LabelCLI
	LabelEnvironment = common.LabelEnvironment
	LabelFile        = common.LabelFile
	LabelDatabase    = common.LabelDatabase
	LabelNetwork     = common.LabelNetwork
	LabelUserInput   = common.LabelUserInput
)

type Definition = common.Definition
type Match = common.Match
type BaseMatcher = common.BaseMatcher

// Matcher interface for language-specific source detection
type Matcher interface {
	Language() string
	FindSources(root *sitter.Node, src []byte) []Match
}

// NewBaseMatcher creates a new base matcher
func NewBaseMatcher(language string, sources []Definition) *BaseMatcher {
	return common.NewBaseMatcher(language, sources)
}
