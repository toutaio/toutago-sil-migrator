package migrations

import (
	"context"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

func init() {
	sil.RegisterMigration(&Migration_20241230100000_CreateUsersTable{})
}

// Migration_20241230100000_CreateUsersTable creates the users table
type Migration_20241230100000_CreateUsersTable struct {
	sil.BaseMigration
}

// Version returns the migration version.
func (m *Migration_20241230100000_CreateUsersTable) Version() string {
	return "20241230100000"
}

// Description returns the migration description.
func (m *Migration_20241230100000_CreateUsersTable) Description() string {
	return "create users table"
}

// Up applies the migration.
func (m *Migration_20241230100000_CreateUsersTable) Up(adapter sil.DatabaseAdapter) error {
	ctx := context.Background()
	
	return adapter.Exec(ctx, `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
}

// Down reverts the migration.
func (m *Migration_20241230100000_CreateUsersTable) Down(adapter sil.DatabaseAdapter) error {
	ctx := context.Background()
	
	return adapter.Exec(ctx, `DROP TABLE IF EXISTS users`)
}
