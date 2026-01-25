// Package constants provides centralized type constants for the tracer.
package constants

// InputLabel categorizes the type of user input
type InputLabel string

const (
	LabelHTTPGet     InputLabel = "http_get"
	LabelHTTPPost    InputLabel = "http_post"
	LabelHTTPCookie  InputLabel = "http_cookie"
	LabelHTTPHeader  InputLabel = "http_header"
	LabelHTTPBody    InputLabel = "http_body"
	LabelCLI         InputLabel = "cli"
	LabelEnvironment InputLabel = "environment"
	LabelFile        InputLabel = "file"
	LabelDatabase    InputLabel = "database"
	LabelNetwork     InputLabel = "network"
	LabelUserInput   InputLabel = "user_input"
)
