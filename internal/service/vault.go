// Package service provides the business logic layer between TUI and storage.
package service

import (
	"fmt"
	"strings"
	"time"

	"envy/internal/domain"
	"envy/internal/storage"
)

// VaultService defines the interface for vault operations.
// This interface enables mocking for tests.
type VaultService interface {
	GetProjects() []domain.Project
	GetProject(name, env string) (*domain.Project, error)
	FindProject(name, env string) (*domain.Project, error) // case-insensitive lookup
	CreateProject(project domain.Project) error
	UpdateProject(oldName, oldEnv string, project domain.Project) error
	DeleteProject(name, env string) error

	AddKey(projectName, projectEnv string, key domain.APIKey) error
	UpdateKey(projectName, projectEnv, keyName, newValue, createdBy string) error
	DeleteKey(projectName, projectEnv, keyName string) error

	Save() error

	GetEncryptionKey() []byte
}

// vaultService implements VaultService.
type vaultService struct {
	projects      []domain.Project
	encryptionKey []byte
}

// NewVaultService creates a new VaultService with the given projects and encryption key.
func NewVaultService(projects []domain.Project, encryptionKey []byte) VaultService {
	return &vaultService{
		projects:      projects,
		encryptionKey: encryptionKey,
	}
}

// GetProjects returns all projects in the vault.
func (v *vaultService) GetProjects() []domain.Project {
	return v.projects
}

// GetProject returns a project by exact name and environment match.
func (v *vaultService) GetProject(name, env string) (*domain.Project, error) {
	for i := range v.projects {
		if v.projects[i].Name == name && v.projects[i].Environment == env {
			return &v.projects[i], nil
		}
	}
	return nil, fmt.Errorf("project '%s' (%s) not found", name, env)
}

// FindProject returns a project by case-insensitive name and environment match.
func (v *vaultService) FindProject(name, env string) (*domain.Project, error) {
	for i := range v.projects {
		if strings.EqualFold(v.projects[i].Name, name) && strings.EqualFold(v.projects[i].Environment, env) {
			return &v.projects[i], nil
		}
	}
	return nil, fmt.Errorf("project '%s' (%s) not found", name, env)
}

// CreateProject adds a new project to the vault after validation.
func (v *vaultService) CreateProject(project domain.Project) error {
	if err := domain.ValidateProjectName(project.Name); err != nil {
		return err
	}
	if err := domain.ValidateEnvironment(project.Environment); err != nil {
		return err
	}

	for _, p := range v.projects {
		if p.Name == project.Name && p.Environment == project.Environment {
			return fmt.Errorf("project '%s' (%s) already exists", project.Name, project.Environment)
		}
	}

	for _, key := range project.Keys {
		if err := domain.ValidateKeyName(key.Key); err != nil {
			return fmt.Errorf("invalid key '%s': %w", key.Key, err)
		}
	}

	v.projects = append(v.projects, project)
	return nil
}

// UpdateProject updates a project identified by oldName and oldEnv.
// This correctly handles renames by searching for the old identity.
func (v *vaultService) UpdateProject(oldName, oldEnv string, project domain.Project) error {
	for i, p := range v.projects {
		if p.Name == oldName && p.Environment == oldEnv {
			v.projects[i] = project
			return nil
		}
	}
	return fmt.Errorf("project '%s' (%s) not found", oldName, oldEnv)
}

// DeleteProject removes a project by name and environment, preserving order.
func (v *vaultService) DeleteProject(name, env string) error {
	for i, p := range v.projects {
		if p.Name == name && p.Environment == env {
			v.projects = append(v.projects[:i], v.projects[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("project '%s' (%s) not found", name, env)
}

// AddKey adds a new key to the specified project.
func (v *vaultService) AddKey(projectName, projectEnv string, key domain.APIKey) error {
	if err := domain.ValidateKeyName(key.Key); err != nil {
		return err
	}

	project, err := v.GetProject(projectName, projectEnv)
	if err != nil {
		return err
	}

	for _, k := range project.Keys {
		if k.Key == key.Key {
			return fmt.Errorf("key '%s' already exists in project", key.Key)
		}
	}

	project.Keys = append(project.Keys, key)
	return nil
}

// UpdateKey updates the value of an existing key, saving the old value to history.
// The createdBy parameter records the source of the change (e.g. "tui-edit", "cli-set").
func (v *vaultService) UpdateKey(projectName, projectEnv, keyName, newValue, createdBy string) error {
	project, err := v.GetProject(projectName, projectEnv)
	if err != nil {
		return err
	}

	for i, key := range project.Keys {
		if key.Key == keyName {
			project.Keys[i].History = append(project.Keys[i].History, key.Current)

			project.Keys[i].Current = domain.SecretVersion{
				Value:     newValue,
				CreatedAt: time.Now(),
				CreatedBy: createdBy,
			}
			return nil
		}
	}

	return fmt.Errorf("key '%s' not found in project '%s' (%s)", keyName, projectName, projectEnv)
}

// DeleteKey removes a key from the specified project, preserving order.
func (v *vaultService) DeleteKey(projectName, projectEnv, keyName string) error {
	project, err := v.GetProject(projectName, projectEnv)
	if err != nil {
		return err
	}

	for i, key := range project.Keys {
		if key.Key == keyName {
			project.Keys = append(project.Keys[:i], project.Keys[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("key '%s' not found in project '%s' (%s)", keyName, projectName, projectEnv)
}

// Save persists all projects to encrypted storage.
func (v *vaultService) Save() error {
	return storage.Save(v.projects, v.encryptionKey)
}

// GetEncryptionKey returns the vault's encryption key.
func (v *vaultService) GetEncryptionKey() []byte {
	return v.encryptionKey
}
