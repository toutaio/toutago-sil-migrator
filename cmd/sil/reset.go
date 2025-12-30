package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Rollback all migrations",
	Long: `Rollback all applied migrations.

This command will:
  1. Acquire a migration lock
  2. Find all applied migrations
  3. Run each migration's Down() method in reverse order
  4. Remove all migration records from the database
  5. Release the lock

⚠️  WARNING: This will rollback ALL migrations. Use with caution!

Examples:
  sil reset
  sil reset --force`,
	RunE: runReset,
}

var (
	resetForce bool
)

func init() {
	resetCmd.Flags().BoolVar(&resetForce, "force", false, "skip confirmation prompt")
}

func runReset(cmd *cobra.Command, args []string) error {
	// Confirmation prompt
	if !resetForce {
		fmt.Print("⚠️  This will rollback ALL migrations. Are you sure? (yes/no): ")
		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Load config
	config, err := loadConfig()
	if err != nil {
		return err
	}

	// Override verbose if set
	if verbose {
		config.Verbose = true
	}

	// Create adapter
	adapter, err := createAdapter(config)
	if err != nil {
		return err
	}

	// Connect to database
	ctx := context.Background()
	if err := adapter.Connect(ctx, config); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer adapter.Close()

	// Create migrator
	migrator, err := sil.NewMigrator(config, adapter)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Set logger
	logger := sil.NewColorLogger(config.Verbose)
	migrator.SetLogger(logger)

	// Run reset
	start := time.Now()

	if err := migrator.Reset(ctx); err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✅ Reset completed in %v\n", elapsed)

	return nil
}
