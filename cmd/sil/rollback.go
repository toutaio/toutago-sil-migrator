package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback the last batch of migrations",
	Long: `Rollback the last batch of migrations that were applied.

This command will:
  1. Acquire a migration lock
  2. Find migrations from the last batch
  3. Run each migration's Down() method in reverse order
  4. Remove migration records from the database
  5. Release the lock

Examples:
  sil rollback                # Rollback last batch
  sil rollback --steps 3      # Rollback last 3 migrations
  sil rollback --batch 5      # Rollback specific batch`,
	RunE: runRollback,
}

var (
	rollbackSteps int
	rollbackBatch int
)

func init() {
	rollbackCmd.Flags().IntVar(&rollbackSteps, "steps", 0, "number of migrations to rollback")
	rollbackCmd.Flags().IntVar(&rollbackBatch, "batch", 0, "specific batch to rollback")
}

func runRollback(cmd *cobra.Command, args []string) error {
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

	// Run rollback
	start := time.Now()

	if rollbackSteps > 0 {
		err = migrator.MigrateDown(ctx, rollbackSteps)
	} else {
		err = migrator.Rollback(ctx)
	}

	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✅ Rollback completed in %v\n", elapsed)

	return nil
}
