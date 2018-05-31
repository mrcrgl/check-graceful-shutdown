package version

import "fmt"

var (
	// Version of the current built
	Version = "n/a"
	// GitCommit of the current built
	GitCommit = "dirty"
)

// GetInfo provides a ready to use version string
func GetInfo() string {
	return fmt.Sprintf("version: %s, commit: %s", Version, GitCommit)
}
