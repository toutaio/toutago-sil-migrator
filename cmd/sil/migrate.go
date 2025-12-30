package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run all pending migrations",
	Long: `Run all pending migrations that have not been applied yet.

This command will:
  1. Acquire a migration lock to prevent concurrent runs
  2. Load all pending migrations
  3. Run each migration in a transaction
  4. Record successful migrations in the database
  5. Release the lock

Examples:
  sil migrate
  sil migrate --steps 5
  sil migrate --config custom.yaml`,
	RunE: runMigrate,
}

var (
	migrateSteps int
)

func init() {
	migrateCmd.Flags().IntVar(&migrateSteps, "steps", 0, "number of migrations to run (0 = all)")
}

func runMigrate(cmd *cobra.Command, args []string) error {
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

	// Set callbacks for better output
	migrator.SetBeforeMigrate(func(migration sil.Migration, direction string) error {
		return nil
	})

	migrator.SetAfterMigrate(func(migration sil.Migration, direction string) error {
		return nil
	})

	// Run migrations
	start := time.Now()

	if migrateSteps > 0 {
		err = migrator.MigrateUp(ctx, migrateSteps)
	} else {
		err = migrator.Migrate(ctx)
	}

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✅ Migration completed in %v\n", elapsed)

	return nil
}
