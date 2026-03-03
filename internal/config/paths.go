// Package config handles application configuration via Lua
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetDefaultDataDir returns the default directory for storing application data (keys.json, .lock)
// - Linux: ~/.envy/
// - macOS: ~/.envy/
// - Windows: %APPDATA%\envy\
func GetDefaultDataDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "envy"
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "envy")
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return ".envy"
		}
		return filepath.Join(home, ".envy")
	}
}

// GetDefaultConfigDir returns the default directory for storing configuration files (config.lua)
// - Linux: ~/.config/envy/
// - macOS: ~/Library/Application Support/envy/
// - Windows: %APPDATA%\envy\
func GetDefaultConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "envy"
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "envy")
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "envy"
		}
		return filepath.Join(home, "Library", "Application Support", "envy")
	default:
		// Respect XDG_CONFIG_HOME if set
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig != "" {
			return filepath.Join(xdgConfig, "envy")
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".config", "envy")
		}
		return filepath.Join(home, ".config", "envy")
	}
}

// GetDefaultKeysPath returns the default path for the encrypted keys file.
func GetDefaultKeysPath() string {
	return filepath.Join(GetDefaultDataDir(), "keys.json")
}

// GetDefaultLockPath returns the default path for the lock file.
func GetDefaultLockPath() string {
	return filepath.Join(GetDefaultDataDir(), ".lock")
}

// GetDefaultConfigPath returns the default path for the Lua config file.
func GetDefaultConfigPath() string {
	return filepath.Join(GetDefaultConfigDir(), "config.lua")
}

// EnsureDirectories creates the data and config directories if they don't exist.
func EnsureDirectories() error {
	dataDir := GetDefaultDataDir()
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return fmt.Errorf("failed to create data directory %s: %w", dataDir, err)
	}

	configDir := GetDefaultConfigDir()
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	return nil
}
