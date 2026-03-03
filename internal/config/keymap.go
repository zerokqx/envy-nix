// Package config handles application configuration including keymaps for future Lua integration
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// KeyMap defines all keybindings used in the application
// This will be configurable via Lua in the future
type KeyMap struct {
	// Navigation
	Up    string `json:"up"`
	Down  string `json:"down"`
	Left  string `json:"left"`
	Right string `json:"right"`

	// Vim-style navigation (alternative)
	VimUp    string `json:"vim_up"`
	VimDown  string `json:"vim_down"`
	VimLeft  string `json:"vim_left"`
	VimRight string `json:"vim_right"`

	// Actions
	Enter       string `json:"enter"`
	Back        string `json:"back"`
	Quit        string `json:"quit"`
	Search      string `json:"search"`
	Yank        string `json:"yank"`
	Create      string `json:"create"`
	Edit        string `json:"edit"`
	EditProject string `json:"edit_project"`
	Delete      string `json:"delete"`
	Save        string `json:"save"`
	Add         string `json:"add"`
	History     string `json:"history"`

	// Form navigation
	Tab      string `json:"tab"`
	ShiftTab string `json:"shift_tab"`
	Space    string `json:"space"`

	// Special
	ForceQuit string `json:"force_quit"`
}

// DefaultKeyMap returns the default keybindings
// Case sensitive
// Example: "ctrl+c" is different from "Ctrl+C"
// Example: A : shift+a ,  unlike a which is just the key a
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation (arrows)
		Up:    "up",
		Down:  "down",
		Left:  "left",
		Right: "right",

		// Vim navigation
		VimUp:    "k",
		VimDown:  "j",
		VimLeft:  "h",
		VimRight: "l",

		// Actions
		Enter:       "enter",
		Back:        "esc",
		Quit:        "q",
		Search:      "i",
		Yank:        "y",
		Create:      "N",
		Edit:        "e",
		EditProject: "E",
		Delete:      "D",
		Save:        "S",
		Add:         "A",
		History:     "H",

		// Form navigation
		Tab:      "tab",
		ShiftTab: "shift+tab",
		Space:    " ",

		// Special
		ForceQuit: "ctrl+c",
	}
}

// LoadKeyMap reads keybindings from the JSON config file, falling back to defaults.
func LoadKeyMap() KeyMap {
	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultKeyMap()
	}

	var km KeyMap
	if err := json.Unmarshal(data, &km); err != nil {
		return DefaultKeyMap()
	}

	return km
}

// SaveKeyMap writes keybindings to the JSON config file.
func SaveKeyMap(km KeyMap) error {
	configPath := getConfigPath()

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(km, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0o600)
}

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "envy", "keymap.json")
	}
	return filepath.Join(home, ".config", "envy", "keymap.json")
}

// MatchesKey returns true if the pressed key matches any of the given bindings.
func (km KeyMap) MatchesKey(key string, bindings ...string) bool {
	for _, binding := range bindings {
		if key == binding {
			return true
		}
	}
	return false
}

// IsNavigationUp returns true if the key matches an upward navigation binding.
func (km KeyMap) IsNavigationUp(key string) bool {
	return km.MatchesKey(key, km.Up, km.VimUp)
}

// IsNavigationDown returns true if the key matches a downward navigation binding.
func (km KeyMap) IsNavigationDown(key string) bool {
	return km.MatchesKey(key, km.Down, km.VimDown)
}

// IsNavigationLeft returns true if the key matches a leftward navigation binding.
func (km KeyMap) IsNavigationLeft(key string) bool {
	return km.MatchesKey(key, km.Left, km.VimLeft)
}

// IsNavigationRight returns true if the key matches a rightward navigation binding.
func (km KeyMap) IsNavigationRight(key string) bool {
	return km.MatchesKey(key, km.Right, km.VimRight)
}
