package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"envy/internal/auth"
	"envy/internal/domain"
	"envy/internal/service"
	"envy/internal/storage"

	"github.com/hashicorp/go-envparse"
)

// RunImport imports secrets from a .env file into the vault.
func RunImport(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	envMap, err := envparse.Parse(file)
	if err != nil {
		fmt.Printf("Error parsing .env file: %v\n", err)
		return
	}

	if len(envMap) == 0 {
		fmt.Println("File is empty or contains no valid keys.")
		return
	}

	name, err := auth.PromptText("Enter Project Name: ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	name = strings.TrimSpace(name)

	if err := domain.ValidateProjectName(name); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	env, err := auth.PromptText("Environment (prod/dev/stage) [dev]: ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	env = strings.TrimSpace(env)
	if env == "" {
		env = domain.EnvDev
	}

	if err := domain.ValidateEnvironment(env); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	firstRun, err := storage.IsFirstRun()
	if err != nil {
		fmt.Printf("Error checking vault status: %v\n", err)
		return
	}

	var password string
	var projects []domain.Project
	var key []byte

	if firstRun {
		fmt.Println("No vault found. Creating new vault...")
		password, err = auth.PromptNewPassword()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if err := storage.Initialize(password); err != nil {
			fmt.Printf("Error initializing vault: %v\n", err)
			return
		}

		projects, key, err = storage.Load(password)
		if err != nil {
			fmt.Printf("Error loading vault: %v\n", err)
			return
		}
	} else {
		password, err = auth.PromptPassword("Enter master password: ")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		projects, key, err = storage.Load(password)
		if err != nil {
			fmt.Printf("Error loading vault: %v\n", err)
			return
		}
	}

	vault := service.NewVaultService(projects, key)

	// Use case-insensitive lookup (P2 #11)
	existing, findErr := vault.FindProject(name, env)
	if findErr == nil {
		response, promptErr := auth.PromptText(fmt.Sprintf("Project '%s' (%s) already exists. Overwrite? [y/N]: ", existing.Name, existing.Environment))
		if promptErr != nil {
			fmt.Printf("Error: %v\n", promptErr)
			return
		}
		if strings.ToLower(response) != "y" {
			fmt.Println("Import cancelled.")
			return
		}
		// Remove the existing project before re-creating
		if err := vault.DeleteProject(existing.Name, existing.Environment); err != nil {
			fmt.Printf("Error removing existing project: %v\n", err)
			return
		}
	}

	var newKeys []domain.APIKey
	for k, v := range envMap {
		if err := domain.ValidateKeyName(k); err != nil {
			fmt.Printf("Warning: Skipping invalid key '%s': %v\n", k, err)
			continue
		}

		newKeys = append(newKeys, domain.APIKey{
			Title: k,
			Key:   k,
			Current: domain.SecretVersion{
				Value:     v,
				CreatedAt: time.Now(),
				CreatedBy: "cli-import",
			},
			History: make([]domain.SecretVersion, 0),
		})
	}

	newProject := domain.Project{
		Name:        name,
		Environment: env,
		Keys:        newKeys,
	}

	if err := vault.CreateProject(newProject); err != nil {
		fmt.Printf("Error creating project: %v\n", err)
		return
	}

	if err := vault.Save(); err != nil {
		fmt.Printf("Error saving to vault: %v\n", err)
		return
	}

	fmt.Printf("\nSuccess! Imported project '%s' (%s) with %d keys.\n", name, env, len(newKeys))
	fmt.Println("Run 'envy' to view your keys in the TUI.")
}
