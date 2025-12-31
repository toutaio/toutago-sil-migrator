package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	rootCmd = &cobra.Command{
		Use:   "sil",
		Short: "Síl - Database migration and seeding tool",
		Long: `Síl is a powerful database migration and seeding tool for Go applications.

It provides a simple, safe, and reliable way to manage database schema changes
and seed data across different environments.

Features:
  • Type-safe migrations in Go
  • Transaction-wrapped migrations
  • Distributed locking
  • Batch tracking for precise rollbacks
  • Multiple database support (PostgreSQL, MySQL, SQLite)
  • Seeding support for test data
  • Easy integration with CI/CD

Visit https://github.com/toutaio/toutago-sil-migrator for more information.`,
		Version: version,
	}

	// Global flags
	configFile string
	verbose    bool
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default: sil.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(refreshCmd)
	rootCmd.AddCommand(seedCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
