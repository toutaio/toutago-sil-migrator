package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new migration file",
	Long: `Create a new migration file with the specified name.

The migration file will be created in the migrations directory with a
timestamp prefix (YYYYMMDDHHMMSS_name.go).

Examples:
  sil create create_users_table
  sil create add_email_to_users
  sil create --table users create_users_table
  sil create --table users --column email add_email_column`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

var (
	createTable  string
	createColumn string
)

func init() {
	createCmd.Flags().StringVar(&createTable, "table", "", "table name for create_table template")
	createCmd.Flags().StringVar(&createColumn, "column", "", "column name for add_column template")
}

func runCreate(cmd *cobra.Command, args []string) error {
	description := args[0]

	// Load config
	config, err := loadConfig()
	if err != nil {
		return err
	}

	// Create generator
	logger := sil.NewColorLogger(verbose)
	generator := sil.NewGenerator(config.MigrationsDir, logger)

	var filePath string

	// Generate migration based on flags
	switch {
	case createTable != "":
		// Create table migration
		filePath, err = generator.GenerateCreateTable(createTable)
		if err != nil {
			return fmt.Errorf("failed to generate migration: %w", err)
		}

	case createColumn != "" && createTable != "":
		// Add column migration
		filePath, err = generator.GenerateAddColumn(createTable, createColumn)
		if err != nil {
			return fmt.Errorf("failed to generate migration: %w", err)
		}

	default:
		// Basic migration
		filePath, err = generator.Generate(description)
		if err != nil {
			return fmt.Errorf("failed to generate migration: %w", err)
		}
	}

	fmt.Printf("✅ Created migration: %s\n", filePath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit the migration file to add your SQL")
	fmt.Println("  2. Run the migration:")
	fmt.Println("     sil migrate")

	return nil
}
