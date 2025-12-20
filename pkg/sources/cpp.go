package sources

// CPPMatcher matches C++ user input sources
type CPPMatcher struct {
	*BaseMatcher
}

// NewCPPMatcher creates a new C++ source matcher
func NewCPPMatcher() *CPPMatcher {
	// Start with all C sources
	cMatcher := NewCMatcher()
	sources := make([]Definition, len(cMatcher.sources))
	for i, src := range cMatcher.sources {
		src.Language = "cpp"
		sources[i] = src
	}

	// Add C++ specific sources
	cppSources := []Definition{
		// iostream
		{
			Name:        "std::cin",
			Pattern:     `std::cin\s*>>|cin\s*>>`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Standard input stream",
			NodeTypes:   []string{"binary_expression", "identifier"},
		},
		{
			Name:        "std::getline()",
			Pattern:     `std::getline\s*\(|getline\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Get line from stream",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "istream::get()",
			Pattern:     `\.get\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Get character(s) from stream",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "istream::getline()",
			Pattern:     `\.getline\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Get line from stream",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "istream::read()",
			Pattern:     `\.read\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Read from stream",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "istream::operator>>",
			Pattern:     `>>\s*\w+`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Stream extraction operator",
			NodeTypes:   []string{"binary_expression"},
		},

		// fstream
		{
			Name:        "std::ifstream",
			Pattern:     `std::ifstream|ifstream`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelFile},
			Description: "File input stream",
			NodeTypes:   []string{"type_identifier", "qualified_identifier"},
		},
		{
			Name:        "std::fstream",
			Pattern:     `std::fstream|fstream`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelFile},
			Description: "File stream",
			NodeTypes:   []string{"type_identifier", "qualified_identifier"},
		},

		// stringstream
		{
			Name:        "std::istringstream",
			Pattern:     `std::istringstream|istringstream`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelUserInput},
			Description: "String input stream",
			NodeTypes:   []string{"type_identifier", "qualified_identifier"},
		},
		{
			Name:        "std::stringstream",
			Pattern:     `std::stringstream|stringstream`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelUserInput},
			Description: "String stream",
			NodeTypes:   []string{"type_identifier", "qualified_identifier"},
		},

		// Boost.Asio (network)
		{
			Name:        "async_read()",
			Pattern:     `async_read\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelNetwork},
			Description: "Boost async read",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "read_some()",
			Pattern:     `\.read_some\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelNetwork},
			Description: "Boost read some",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "receive()",
			Pattern:     `\.receive\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelNetwork},
			Description: "Socket receive",
			NodeTypes:   []string{"call_expression"},
		},

		// Qt framework
		{
			Name:        "QFile::readAll()",
			Pattern:     `\.readAll\s*\(\s*\)`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelFile},
			Description: "Qt read all file contents",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "QFile::readLine()",
			Pattern:     `\.readLine\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelFile},
			Description: "Qt read line from file",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "QTextStream::readLine()",
			Pattern:     `\.readLine\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelFile, LabelUserInput},
			Description: "Qt text stream read line",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "QNetworkReply",
			Pattern:     `QNetworkReply`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelNetwork},
			Description: "Qt network reply",
			NodeTypes:   []string{"type_identifier"},
		},
		{
			Name:        "QProcess::readAllStandardOutput()",
			Pattern:     `\.readAllStandardOutput\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Qt process stdout",
			NodeTypes:   []string{"call_expression"},
		},

		// CLI argument parsing
		{
			Name:        "QCoreApplication::arguments()",
			Pattern:     `QCoreApplication::arguments\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelCLI},
			Description: "Qt command line arguments",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "boost::program_options",
			Pattern:     `program_options`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelCLI},
			Description: "Boost program options",
			NodeTypes:   []string{"namespace_identifier"},
		},

		// Environment
		{
			Name:        "std::getenv()",
			Pattern:     `std::getenv\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelEnvironment},
			Description: "Get environment variable",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "qgetenv()",
			Pattern:     `qgetenv\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelEnvironment},
			Description: "Qt get environment variable",
			NodeTypes:   []string{"call_expression"},
		},

		// JSON parsing (nlohmann/json, rapidjson)
		{
			Name:        "nlohmann::json::parse()",
			Pattern:     `json::parse\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelUserInput},
			Description: "nlohmann JSON parse",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "rapidjson::Document::Parse()",
			Pattern:     `\.Parse\s*\(`,
			Language:    "cpp",
			Labels:      []InputLabel{LabelUserInput},
			Description: "RapidJSON parse",
			NodeTypes:   []string{"call_expression"},
		},
	}

	sources = append(sources, cppSources...)

	return &CPPMatcher{
		BaseMatcher: NewBaseMatcher("cpp", sources),
	}
}
