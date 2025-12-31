package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Rollback all migrations and re-run them",
	Long: `Rollback all migrations and then re-run them.

This command is useful for:
  • Resetting your development database
  • Testing migration rollback functionality
  • Ensuring migrations are idempotent

⚠️  WARNING: This will rollback and re-run ALL migrations. Use with caution!

Examples:
  sil refresh
  sil refresh --force`,
	RunE: runRefresh,
}

var (
	refreshForce bool
)

func init() {
	refreshCmd.Flags().BoolVar(&refreshForce, "force", false, "skip confirmation prompt")
}

func runRefresh(cmd *cobra.Command, args []string) error {
	// Confirmation prompt
	if !refreshForce {
		fmt.Print("⚠️  This will rollback and re-run ALL migrations. Are you sure? (yes/no): ")
		var response string
		_, _ = fmt.Scanln(&response)
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
	defer func() { _ = adapter.Close() }()

	// Create migrator
	migrator, err := sil.NewMigrator(config, adapter)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Set logger
	logger := sil.NewColorLogger(config.Verbose)
	migrator.SetLogger(logger)

	// Run refresh
	start := time.Now()

	if err := migrator.Refresh(ctx); err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✅ Refresh completed in %v\n", elapsed)

	return nil
}
