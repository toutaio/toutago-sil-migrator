package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil/adapters"

	// Import migrations to register them
	_ "github.com/toutaio/toutago-sil-migrator/examples/basic/migrations"
)

func main() {
	// Configure database connection
	config := sil.DefaultConfig()
	config.DatabaseURL = getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/sil_example?sslmode=disable")
	config.MigrationsDir = "./examples/basic/migrations"
	config.Verbose = true

	// Create PostgreSQL adapter
	adapter, err := adapters.NewPostgresAdapter(config)
	if err != nil {
		log.Fatalf("Failed to create adapter: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	if err := adapter.Connect(ctx, config); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer adapter.Close()

	// Create migrator
	migrator, err := sil.NewMigrator(config, adapter)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}

	// Run migrations
	fmt.Println("Running migrations...")
	if err := migrator.Migrate(ctx); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Show status
	fmt.Println("\nMigration status:")
	statuses, err := migrator.Status(ctx)
	if err != nil {
		log.Fatalf("Failed to get status: %v", err)
	}

	for _, status := range statuses {
		appliedStatus := "❌ Pending"
		if status.Applied {
			appliedStatus = fmt.Sprintf("✅ Applied (batch %d)", status.Batch)
		}
		fmt.Printf("  %s - %s - %s\n", status.Version, status.Description, appliedStatus)
	}

	fmt.Println("\n✅ Migration complete!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
