package commands

import (
	"fmt"
	"strings"
	"time"

	"envy/internal/auth"
	"envy/internal/config"
	"envy/internal/domain"
	"envy/internal/service"
	"envy/internal/storage"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [project] [KEY=VALUE]",
	Short: "Set or update a secret in a project",
	Long: `Set or update a secret (environment variable) in a project.

If the project doesn't exist, you'll be prompted to create it.
If the key already exists, it will be updated (old value saved to history).

Examples:
  envy set myproject API_KEY=sk-1234567890
  envy set "My Project" DATABASE_URL=postgresql://localhost/db
  envy set production SECRET_TOKEN=abc123 -e prod

Shorthand syntax (using -s flag):
  envy -s myproject API_KEY=sk-1234567890
  envy -s "My Project" DATABASE_URL=postgresql://localhost/db
  envy -s production SECRET_TOKEN=abc123 -e prod

The -e/--env flag specifies the environment (default: dev).`,
	Args: cobra.ExactArgs(2),
	RunE: runSetCommand,
}

func init() {
	RootCmd.AddCommand(setCmd)
	setCmd.Flags().StringP("env", "e", "dev", "Environment (dev, staging, prod)")
}

func runSetCommand(cmd *cobra.Command, args []string) error {
	projectName := args[0]
	keyValuePair := args[1]
	environment, _ := cmd.Flags().GetString("env")

	return performSet(projectName, keyValuePair, environment)
}

// performSet contains the core logic for setting a secret.
func performSet(projectName, keyValuePair, environment string) error {
	// Parse KEY=VALUE
	parts := strings.SplitN(keyValuePair, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format. Expected: KEY=VALUE, got: %s", keyValuePair)
	}

	keyName := strings.TrimSpace(parts[0])
	keyValue := parts[1] // Don't trim value, user might want leading/trailing spaces

	if keyName == "" {
		return fmt.Errorf("key name cannot be empty")
	}

	if keyValue == "" {
		return fmt.Errorf("key value cannot be empty")
	}

	if err := domain.ValidateKeyName(keyName); err != nil {
		return fmt.Errorf("invalid key name: %w", err)
	}

	if err := domain.ValidateEnvironment(environment); err != nil {
		return fmt.Errorf("invalid environment: %w", err)
	}

	if err := config.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	firstRun, err := storage.IsFirstRun()
	if err != nil {
		return fmt.Errorf("failed to check vault status: %w", err)
	}

	if firstRun {
		return fmt.Errorf("no vault found. Please run 'envy' to create a vault first")
	}

	password, err := auth.PromptPassword("Enter master password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	projects, key, err := storage.Load(password)
	if err != nil {
		return fmt.Errorf("failed to load vault: %w", err)
	}

	vault := service.NewVaultService(projects, key)

	// Use case-insensitive lookup (P2 #11)
	project, err := vault.FindProject(projectName, environment)

	// If project doesn't exist, ask to create it
	if err != nil {
		fmt.Printf("Project '%s' (%s) not found.\n", projectName, environment)
		fmt.Println("Tip: Check the environment and use -e {environment} to specify it. Default is 'dev'.")
		confirm, promptErr := auth.PromptText("Create new project? [y/N]: ")
		if promptErr != nil {
			return fmt.Errorf("failed to read input: %w", promptErr)
		}

		if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
			fmt.Println("Operation cancelled.")
			return nil
		}

		newProject := domain.Project{
			Name:        projectName,
			Environment: environment,
			Keys:        []domain.APIKey{},
		}
		if createErr := vault.CreateProject(newProject); createErr != nil {
			return fmt.Errorf("failed to create project: %w", createErr)
		}
		fmt.Printf("Created project '%s' (%s)\n", projectName, environment)

		project, err = vault.GetProject(projectName, environment)
		if err != nil {
			return fmt.Errorf("failed to retrieve created project: %w", err)
		}
	}

	// Check if key already exists
	keyExists := false
	for _, k := range project.Keys {
		if k.Key == keyName {
			keyExists = true
			break
		}
	}

	if keyExists {
		if err := vault.UpdateKey(project.Name, project.Environment, keyName, keyValue, "cli-set"); err != nil {
			return fmt.Errorf("failed to update key: %w", err)
		}
		fmt.Printf("Updated '%s' in project '%s' (%s)\n", keyName, project.Name, project.Environment)
		fmt.Printf("  Old value saved to history\n")
	} else {
		newKey := domain.APIKey{
			Title: keyName,
			Key:   keyName,
			Current: domain.SecretVersion{
				Value:     keyValue,
				CreatedAt: time.Now(),
				CreatedBy: "cli-set",
			},
			History: []domain.SecretVersion{},
		}
		if err := vault.AddKey(project.Name, project.Environment, newKey); err != nil {
			return fmt.Errorf("failed to add key: %w", err)
		}
		fmt.Printf("Added '%s' to project '%s' (%s)\n", keyName, project.Name, project.Environment)
	}

	if err := vault.Save(); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	fmt.Println("Vault saved successfully")
	return nil
}
