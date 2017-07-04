package prototype

import "fmt"

var (
	// SourceVersion is set via the makefile
	SourceVersion = "DEVELOPMENT"
)

// VersionString returns the version of the software
func VersionString() string {
	return fmt.Sprintf("1.0.%s", SourceVersion)
}
