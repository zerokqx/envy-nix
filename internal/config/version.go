package config

// Version is the current application version string.
const (
	Version = "1.2.1"
)

// GetVersion returns the raw version string.
func GetVersion() string {
	return Version
}

// GetFullVersion returns the version string prefixed with the application name.
func GetFullVersion() string {
	return "Envy v" + Version
}
