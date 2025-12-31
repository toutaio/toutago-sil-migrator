package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil/adapters"
)

var (
	seedCmd = &cobra.Command{
		Use:   "seed",
		Short: "Run database seeders",
		Long:  `Run database seeders to populate your database with test or reference data.`,
		RunE:  runSeed,
	}

	seedCreateCmd = &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new seeder",
		Long:  `Create a new seeder file from template.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runSeedCreate,
	}

	seedStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show seeder execution status",
		Long:  `Display the status of all registered seeders.`,
		RunE:  runSeedStatus,
	}

	// Seed flags
	seedAll    bool
	seedForce  bool
	seedDryRun bool
	seedEnv    string
)

func init() {
	// seed command flags
	seedCmd.Flags().BoolVarP(&seedAll, "all", "a", false, "run all seeders")
	seedCmd.Flags().BoolVarP(&seedForce, "force", "f", false, "force re-run (ignore already executed)")
	seedCmd.Flags().BoolVarP(&seedDryRun, "dry-run", "d", false, "show what would be seeded without executing")
	seedCmd.Flags().StringVarP(&seedEnv, "env", "e", "", "environment to run seeders for")

	// Add subcommands
	seedCmd.AddCommand(seedCreateCmd)
	seedCmd.AddCommand(seedStatusCmd)
}

func runSeed(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if seedEnv != "" {
		cfg.Environment = seedEnv
	}

	cfg.Verbose = verbose

	// Connect to database
	adapter, err := adapters.NewAdapter(cfg)
	if err != nil {
		return fmt.Errorf("failed to create adapter: %w", err)
	}

	ctx := context.Background()
	if err := adapter.Connect(ctx, cfg); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer adapter.Close()

	// Create seed manager
	manager, err := sil.NewSeedManager(cfg, adapter)
	if err != nil {
		return fmt.Errorf("failed to create seed manager: %w", err)
	}

	// Load seeders from directory
	if err := loadSeeders(cfg.SeedersDir); err != nil {
		return fmt.Errorf("failed to load seeders: %w", err)
	}

	printBanner("Seeding Database")

	if seedDryRun {
		return runSeedDryRun(manager, args)
	}

	// Run seeders
	if seedAll || len(args) == 0 {
		return manager.SeedAll(ctx)
	}

	return manager.Seed(ctx, args...)
}

func runSeedDryRun(manager sil.SeedManager, seeders []string) error {
	ctx := context.Background()

	fmt.Println("\n🔍 Dry Run Mode - No changes will be made")

	statuses, err := manager.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get seeder status: %w", err)
	}

	if len(seeders) > 0 {
		// Filter to requested seeders
		filtered := make([]sil.SeedStatus, 0)
		for _, status := range statuses {
			for _, name := range seeders {
				if status.Name == name {
					filtered = append(filtered, status)
					break
				}
			}
		}
		statuses = filtered
	}

	fmt.Println("Seeders that would run:")
	for i, status := range statuses {
		icon := "✓"
		reason := ""

		if status.Executed && !seedForce {
			icon = "⊘"
			reason = " (already executed)"
		} else if status.Skipped {
			icon = "⊘"
			reason = fmt.Sprintf(" (%s)", status.Reason)
		}

		fmt.Printf("  %s [%d/%d] %s%s\n", icon, i+1, len(statuses), status.Name, reason)
	}

	fmt.Println("\nℹ️  Use --force to re-run already executed seeders")
	fmt.Println("ℹ️  Remove --dry-run to execute seeders")

	return nil
}

func runSeedCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// For create command, we only need the seeders directory
	// Try to load config, but use defaults if it fails
	cfg, err := loadConfig()
	if err != nil {
		// Use defaults if config can't be loaded
		cfg = sil.DefaultConfig()
	}

	// Create seeders directory if it doesn't exist
	if err := os.MkdirAll(cfg.SeedersDir, 0755); err != nil {
		return fmt.Errorf("failed to create seeders directory: %w", err)
	}

	// Generate filename
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.go", timestamp, toSnakeCase(name))
	filepath := filepath.Join(cfg.SeedersDir, filename)

	// Check if file exists
	if _, err := os.Stat(filepath); err == nil {
		return fmt.Errorf("seeder file already exists: %s", filepath)
	}

	// Create seeder from template
	tmpl, err := template.New("seeder").Parse(seederTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	data := map[string]string{
		"Name":      toPascalCase(name),
		"NameLower": toSnakeCase(name),
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	printSuccess(fmt.Sprintf("Created seeder: %s", filepath))
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit the seeder file to implement your seeding logic")
	fmt.Println("  2. Build your application to register the seeder")
	fmt.Println("  3. Run: sil seed")

	return nil
}

func runSeedStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.Verbose = verbose

	// Connect to database
	adapter, err := adapters.NewAdapter(cfg)
	if err != nil {
		return fmt.Errorf("failed to create adapter: %w", err)
	}

	ctx := context.Background()
	if err := adapter.Connect(ctx, cfg); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer adapter.Close()

	// Create seed manager
	manager, err := sil.NewSeedManager(cfg, adapter)
	if err != nil {
		return fmt.Errorf("failed to create seed manager: %w", err)
	}

	// Load seeders
	if err := loadSeeders(cfg.SeedersDir); err != nil {
		return fmt.Errorf("failed to load seeders: %w", err)
	}

	// Get status
	statuses, err := manager.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get seeder status: %w", err)
	}

	printBanner("Seeder Status")

	if len(statuses) == 0 {
		fmt.Println("\n⚠️  No seeders found")
		return nil
	}

	fmt.Printf("\nEnvironment: %s\n", cfg.Environment)
	fmt.Printf("Total Seeders: %d\n\n", len(statuses))

	// Count statistics
	executed := 0
	skipped := 0
	pending := 0

	for _, status := range statuses {
		if status.Executed {
			executed++
		} else if status.Skipped {
			skipped++
		} else {
			pending++
		}
	}

	fmt.Printf("Executed: %d | Pending: %d | Skipped: %d\n\n", executed, pending, skipped)

	// Display table header
	fmt.Println("Seeder                          Status      Executed At")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	for _, status := range statuses {
		statusIcon := "○"
		statusText := "Pending"
		executedAt := "-"

		if status.Executed {
			statusIcon = "●"
			statusText = "Executed"
			if status.ExecutedAt != nil {
				executedAt = status.ExecutedAt.Format("2006-01-02 15:04:05")
			}
		} else if status.Skipped {
			statusIcon = "⊘"
			statusText = "Skipped"
		}

		fmt.Printf("%-30s  %s %-10s  %s\n",
			truncate(status.Name, 30),
			statusIcon,
			statusText,
			executedAt,
		)
	}

	return nil
}

// loadSeeders loads seeders from the seeders directory
func loadSeeders(dir string) error {
	// In a real implementation, this would:
	// 1. Discover .go files in the seeders directory
	// 2. Use go/build to compile them
	// 3. Load them dynamically
	//
	// For now, seeders must be registered in the application code
	// using sil.RegisterSeeder() in an init() function

	// This is a placeholder - in practice, seeders are registered
	// when the application is built
	return nil
}

// Seeder template
const seederTemplate = `package seeders

import (
	"context"

	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

func init() {
	sil.RegisterSeeder(New{{.Name}}Seeder())
}

// New{{.Name}}Seeder creates a new {{.NameLower}} seeder
func New{{.Name}}Seeder() sil.Seeder {
	return sil.NewBaseSeeder("{{.NameLower}}", seed{{.Name}}).
		SetDependencies(). // Add dependencies here
		SetEnvironments("development", "test") // Adjust environments
}

// seed{{.Name}} implements the seeding logic
func seed{{.Name}}(ctx context.Context, adapter sil.DatabaseAdapter) error {
	// TODO: Implement your seeding logic here
	// Example:
	// return adapter.Exec(ctx, ` + "`" + `
	//     INSERT INTO {{.NameLower}} (name, description)
	//     VALUES ('Example', 'Example description')
	// ` + "`" + `)
	
	return nil
}
`
