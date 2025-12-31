package main

import (
	"context"
	"fmt"
	"log"

	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil/adapters"
)

// Example seeders demonstrating the seeding system

func init() {
	// Register all seeders
	sil.RegisterSeeder(NewCountriesSeeder())
	sil.RegisterSeeder(NewRolesSeeder())
	sil.RegisterSeeder(NewUsersSeeder())
	sil.RegisterSeeder(NewPostsSeeder())
}

// Countries seeder - no dependencies
func NewCountriesSeeder() sil.Seeder {
	return sil.NewBaseSeeder("countries", func(ctx context.Context, adapter sil.DatabaseAdapter) error {
		countries := []string{"USA", "Canada", "UK", "Germany", "France"}

		for _, country := range countries {
			err := adapter.Exec(ctx, `
				INSERT INTO countries (name, code)
				VALUES ($1, $2)
				ON CONFLICT (code) DO NOTHING
			`, country, country[:2])

			if err != nil {
				return fmt.Errorf("failed to insert country %s: %w", country, err)
			}
		}

		return nil
	}).
		SetEnvironments("development", "test"). // Only in dev/test
		SetShouldRun(func(ctx context.Context, adapter sil.DatabaseAdapter) (bool, error) {
			// Only run if countries table is empty
			rows, err := adapter.Query(ctx, "SELECT COUNT(*) FROM countries")
			if err != nil {
				return false, err
			}
			defer rows.Close()

			var count int
			if rows.Next() {
				rows.Scan(&count)
			}

			return count == 0, nil
		})
}

// Roles seeder - no dependencies
func NewRolesSeeder() sil.Seeder {
	return sil.NewBaseSeeder("roles", func(ctx context.Context, adapter sil.DatabaseAdapter) error {
		roles := []struct {
			name        string
			description string
		}{
			{"admin", "Administrator with full access"},
			{"editor", "Can create and edit content"},
			{"viewer", "Read-only access"},
		}

		for _, role := range roles {
			err := adapter.Exec(ctx, `
				INSERT INTO roles (name, description)
				VALUES ($1, $2)
				ON CONFLICT (name) DO NOTHING
			`, role.name, role.description)

			if err != nil {
				return fmt.Errorf("failed to insert role %s: %w", role.name, err)
			}
		}

		return nil
	}).
		SetEnvironments("development", "test").
		SetShouldRun(func(ctx context.Context, adapter sil.DatabaseAdapter) (bool, error) {
			rows, err := adapter.Query(ctx, "SELECT COUNT(*) FROM roles")
			if err != nil {
				return false, err
			}
			defer rows.Close()

			var count int
			if rows.Next() {
				rows.Scan(&count)
			}

			return count == 0, nil
		})
}

// Users seeder - depends on countries and roles
func NewUsersSeeder() sil.Seeder {
	return sil.NewBaseSeeder("users", func(ctx context.Context, adapter sil.DatabaseAdapter) error {
		// Get admin role ID
		rows, err := adapter.Query(ctx, "SELECT id FROM roles WHERE name = 'admin' LIMIT 1")
		if err != nil {
			return err
		}
		defer rows.Close()

		var roleID int
		if rows.Next() {
			rows.Scan(&roleID)
		} else {
			return fmt.Errorf("admin role not found")
		}

		// Insert admin user
		return adapter.Exec(ctx, `
			INSERT INTO users (email, name, role_id)
			VALUES ('admin@example.com', 'Admin User', $1)
			ON CONFLICT (email) DO NOTHING
		`, roleID)
	}).
		SetDependencies("countries", "roles"). // Must run after countries and roles
		SetEnvironments("development", "test").
		SetShouldRun(func(ctx context.Context, adapter sil.DatabaseAdapter) (bool, error) {
			rows, err := adapter.Query(ctx, "SELECT COUNT(*) FROM users")
			if err != nil {
				return false, err
			}
			defer rows.Close()

			var count int
			if rows.Next() {
				rows.Scan(&count)
			}

			return count == 0, nil
		})
}

// Posts seeder - depends on users
func NewPostsSeeder() sil.Seeder {
	return sil.NewBaseSeeder("posts", func(ctx context.Context, adapter sil.DatabaseAdapter) error {
		// Get admin user ID
		rows, err := adapter.Query(ctx, "SELECT id FROM users WHERE email = 'admin@example.com' LIMIT 1")
		if err != nil {
			return err
		}
		defer rows.Close()

		var userID int
		if rows.Next() {
			rows.Scan(&userID)
		} else {
			return fmt.Errorf("admin user not found")
		}

		// Insert sample posts
		posts := []struct {
			title   string
			content string
		}{
			{"Welcome Post", "Welcome to our platform!"},
			{"Getting Started", "Here's how to get started..."},
			{"Best Practices", "Follow these best practices..."},
		}

		for _, post := range posts {
			err := adapter.Exec(ctx, `
				INSERT INTO posts (title, content, author_id)
				VALUES ($1, $2, $3)
			`, post.title, post.content, userID)

			if err != nil {
				return fmt.Errorf("failed to insert post %s: %w", post.title, err)
			}
		}

		return nil
	}).
		SetDependencies("users"). // Must run after users
		SetEnvironments("development", "test").
		SetShouldRun(func(ctx context.Context, adapter sil.DatabaseAdapter) (bool, error) {
			rows, err := adapter.Query(ctx, "SELECT COUNT(*) FROM posts")
			if err != nil {
				return false, err
			}
			defer rows.Close()

			var count int
			if rows.Next() {
				rows.Scan(&count)
			}

			return count == 0, nil
		})
}

func main() {
	// Configuration
	config := sil.DefaultConfig()
	config.DatabaseURL = "postgres://localhost/mydb"
	config.Environment = "development"
	config.Verbose = true

	// Create adapter
	adapter, err := adapters.NewAdapter(config)
	if err != nil {
		log.Fatalf("Failed to create adapter: %v", err)
	}

	// Connect
	ctx := context.Background()
	if err := adapter.Connect(ctx, config); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer adapter.Close()

	// Create seed manager
	manager, err := sil.NewSeedManager(config, adapter)
	if err != nil {
		log.Fatalf("Failed to create seed manager: %v", err)
	}

	// Seed all
	fmt.Println("Running seeders...")
	if err := manager.SeedAll(ctx); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	fmt.Println("\n✓ Seeding complete!")

	// Show status
	fmt.Println("\nSeeder Status:")
	statuses, _ := manager.Status(ctx)
	for _, status := range statuses {
		icon := "○"
		if status.Executed {
			icon = "●"
		}
		fmt.Printf("  %s %s\n", icon, status.Name)
	}
}
