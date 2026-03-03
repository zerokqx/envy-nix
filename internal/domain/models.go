// Package domain: Stores all the structs necessary to store data and render the view.
package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Environment constants for project classification.
const (
	EnvProd  = "prod"
	EnvStage = "stage"
	EnvDev   = "dev"
)

// SecretVersion represents a single version of a secret value with metadata.
type SecretVersion struct {
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

// APIKey represents a named secret with its current value and version history.
type APIKey struct {
	Title   string          `json:"title"`
	Key     string          `json:"key"`
	Current SecretVersion   `json:"current"`
	History []SecretVersion `json:"history"`
}

// Project groups related secrets under a name and environment.
type Project struct {
	Name        string   `json:"name"`
	Environment string   `json:"environment"`
	Keys        []APIKey `json:"keys"`
}

// Store is the top-level structure persisted to disk as encrypted JSON.
type Store struct {
	Version  int       `json:"version"`
	Salt     string    `json:"salt"`
	AuthHash string    `json:"auth_hash"`
	Projects []Project `json:"projects"`
}

// ValidateProjectName checks that a project name is non-empty and within length limits.
func ValidateProjectName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("project name cannot be empty")
	}
	if len(name) > 256 {
		return errors.New("project name too long (max 256 characters)")
	}
	return nil
}

// ValidateEnvironment checks that env is one of the allowed values (prod, dev, stage).
func ValidateEnvironment(env string) error {
	env = strings.TrimSpace(env)
	if env != EnvProd && env != EnvDev && env != EnvStage {
		return fmt.Errorf("invalid environment '%s' (must be prod, dev, or stage)", env)
	}
	return nil
}

// ValidateKeyName checks that a key name is non-empty, within length limits,
// and does not contain forbidden characters (=, newline, carriage return).
func ValidateKeyName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("key name cannot be empty")
	}
	if strings.ContainsAny(name, "=\n\r") {
		return errors.New("key name cannot contain =, newline, or carriage return")
	}
	if len(name) > 256 {
		return errors.New("key name too long (max 256 characters)")
	}
	return nil
}
