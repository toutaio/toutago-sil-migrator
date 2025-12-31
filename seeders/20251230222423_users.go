package seeders

import (
	"context"

	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

func init() {
	sil.RegisterSeeder(NewUsersSeeder())
}

// NewUsersSeeder creates a new users seeder
func NewUsersSeeder() sil.Seeder {
	return sil.NewBaseSeeder("users", seedUsers).
		SetDependencies(). // Add dependencies here
		SetEnvironments("development", "test") // Adjust environments
}

// seedUsers implements the seeding logic
func seedUsers(ctx context.Context, adapter sil.DatabaseAdapter) error {
	// TODO: Implement your seeding logic here
	// Example:
	// return adapter.Exec(ctx, `
	//     INSERT INTO users (name, description)
	//     VALUES ('Example', 'Example description')
	// `)
	
	return nil
}
