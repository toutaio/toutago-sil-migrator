package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Síl project",
	Long: `Initialize a new Síl project by creating the necessary directories
and configuration files.

This command will create:
  • migrations/ directory for migration files
  • seeders/ directory for seeder files  
  • sil.yaml configuration file

Example:
  sil init
  sil init --config custom.yaml`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().Bool("force", false, "overwrite existing files")
}

func runInit(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	// Determine config file path
	configPath := configFile
	if configPath == "" {
		configPath = "sil.yaml"
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configPath)
	}

	// Create config with defaults
	config := sil.DefaultConfig()
	
	// Update database URL if provided via environment
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.DatabaseURL = dbURL
	}

	// Create migrations directory
	if err := os.MkdirAll(config.MigrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}
	fmt.Printf("✓ Created directory: %s\n", config.MigrationsDir)

	// Create seeders directory
	if err := os.MkdirAll(config.SeedersDir, 0755); err != nil {
		return fmt.Errorf("failed to create seeders directory: %w", err)
	}
	fmt.Printf("✓ Created directory: %s\n", config.SeedersDir)

	// Save config file
	if err := sil.SaveConfig(config, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("✓ Created config file: %s\n", configPath)

	// Create a sample migration
	sampleMigration := filepath.Join(config.MigrationsDir, ".gitkeep")
	if err := os.WriteFile(sampleMigration, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create .gitkeep: %w", err)
	}

	fmt.Println("\n✅ Síl project initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Update sil.yaml with your database connection")
	fmt.Println("  2. Create your first migration:")
	fmt.Println("     sil create create_users_table")
	fmt.Println("  3. Run migrations:")
	fmt.Println("     sil migrate")

	return nil
}
