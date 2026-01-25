// Package php provides PHP database-related patterns
package php

import "strings"

// =============================================================================
// DATABASE FETCH METHODS
// Methods that fetch data from database results
// =============================================================================

// PDOFetchMethods contains PDO statement fetch methods
// These are object methods called on PDOStatement objects
var PDOFetchMethods = map[string]bool{
	"fetch":       true,
	"fetchAll":    true,
	"fetchColumn": true,
	"fetchObject": true,
}

// MySQLiFetchMethods contains MySQLi result fetch methods
var MySQLiFetchMethods = map[string]bool{
	"fetch_array":  true,
	"fetch_assoc":  true,
	"fetch_row":    true,
	"fetch_object": true,
	"fetch_all":    true,
}

// AllDatabaseFetchMethods combines all database fetch method names
var AllDatabaseFetchMethods = map[string]bool{
	// PDO
	"fetch":       true,
	"fetchAll":    true,
	"fetchColumn": true,
	"fetchObject": true,
	// MySQLi (object methods)
	"fetch_array":  true,
	"fetch_assoc":  true,
	"fetch_row":    true,
	"fetch_object": true,
	"fetch_all":    true,
}

// IsDatabaseFetchMethod returns true if the method name is a database fetch method
// This checks object-oriented fetch methods (PDO, MySQLi object style)
func IsDatabaseFetchMethod(methodName string) bool {
	return AllDatabaseFetchMethods[methodName]
}

// IsPDOFetchMethod returns true if the method is a PDO fetch method
func IsPDOFetchMethod(methodName string) bool {
	return PDOFetchMethods[methodName]
}

// IsMySQLiFetchMethod returns true if the method is a MySQLi fetch method
func IsMySQLiFetchMethod(methodName string) bool {
	return MySQLiFetchMethods[methodName]
}

// =============================================================================
// DATABASE QUERY METHODS (SINKS - NOT SOURCES)
// These are methods that execute queries - they're sinks, not sources
// Included for completeness but NOT used for source detection
// =============================================================================

// DatabaseQueryMethods are methods that execute database queries (sinks)
var DatabaseQueryMethods = map[string]bool{
	// PDO
	"query":   true,
	"exec":    true,
	"prepare": true,
	"execute": true,
	// MySQLi
	"real_query":       true,
	"multi_query":      true,
	"send_query":       true,
	"real_escape_string": true,
}

// IsDatabaseQueryMethod returns true if method is a database query method (sink)
func IsDatabaseQueryMethod(methodName string) bool {
	return DatabaseQueryMethods[methodName]
}

// =============================================================================
// DATABASE RESULT OBJECT PATTERNS
// =============================================================================

// DatabaseResultObjectPatterns matches object names that are likely database results
var DatabaseResultObjectPatterns = []string{
	"result",
	"stmt",
	"statement",
	"query",
	"res",
	"row",
	"rows",
}

// IsDatabaseResultObject checks if an object name looks like a database result
func IsDatabaseResultObject(objName string) bool {
	lower := strings.ToLower(objName)
	for _, pattern := range DatabaseResultObjectPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}
