package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long: `Show the status of all migrations.

This command will display a table showing:
  • Migration version
  • Description
  • Status (Applied or Pending)
  • Batch number (if applied)
  • Execution timestamp (if applied)

Examples:
  sil status
  sil status --config custom.yaml`,
	RunE: runStatus,
}

var (
	statusJSON bool
)

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "output in JSON format")
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	// Get status
	statuses, err := migrator.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if len(statuses) == 0 {
		fmt.Println("No migrations found")
		return nil
	}

	// Display status
	if statusJSON {
		// TODO: Implement JSON output
		return fmt.Errorf("JSON output not yet implemented")
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_ = fmt.Fprintln(w, "VERSION\tDESCRIPTION\tSTATUS\tBATCH\tEXECUTED AT")
	_ = fmt.Fprintln(w, "-------\t-----------\t------\t-----\t-----------")

	appliedCount := 0
	pendingCount := 0

	for _, status := range statuses {
		statusStr := "❌ Pending"
		batchStr := "-"
		executedStr := "-"

		if status.Applied {
			statusStr = "✅ Applied"
			batchStr = fmt.Sprintf("%d", status.Batch)
			if status.ExecutedAt != nil {
				executedStr = status.ExecutedAt.Format("2006-01-02 15:04:05")
			}
			appliedCount++
		} else {
			pendingCount++
		}

		_ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			status.Version,
			truncate(status.Description, 40),
			statusStr,
			batchStr,
			executedStr,
		)
	}

	w.Flush()

	// Summary
	fmt.Printf("\nTotal: %d migrations (%d applied, %d pending)\n",
		len(statuses), appliedCount, pendingCount)

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
