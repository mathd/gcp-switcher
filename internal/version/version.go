package version

// Version information
var (
	// Version indicates the current version.
	Version = "dev"
	// Commit indicates the git commit.
	Commit = "none"
	// Date indicates the build date.
	Date = "unknown"
)

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	return "GCP Switcher " + Version + "\ncommit: " + Commit + "\nbuilt at: " + Date
}
