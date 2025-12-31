package sil

import (
	"context"
	"testing"
	"time"
)

// BenchmarkMigrationRegistration benchmarks registering migrations
func BenchmarkMigrationRegistration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ClearRegisteredMigrations()
		b.StartTimer()

		version := GenerateVersion()
		m := NewBaseMigration(version, "benchmark", nil, nil)
		RegisterMigration(m)
	}
}

// BenchmarkGetRegisteredMigrations benchmarks retrieving registered migrations
func BenchmarkGetRegisteredMigrations(b *testing.B) {
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	// Setup: Register 100 migrations
	for i := 0; i < 100; i++ {
		version := GenerateVersion()
		m := NewBaseMigration(version, "benchmark", nil, nil)
		RegisterMigration(m)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetRegisteredMigrations()
	}
}

// BenchmarkSortMigrations benchmarks sorting migrations
func BenchmarkSortMigrations(b *testing.B) {
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	// Create unsorted migrations
	migrations := make([]Migration, 100)
	for i := 0; i < 100; i++ {
		version := GenerateVersion()
		migrations[i] = NewBaseMigration(version, "benchmark", nil, nil)
		time.Sleep(time.Microsecond) // Ensure different timestamps
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SortMigrations(migrations)
	}
}

// BenchmarkFilterPendingMigrations benchmarks filtering pending migrations
func BenchmarkFilterPendingMigrations(b *testing.B) {
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	// Create 100 migrations
	migrations := make([]Migration, 100)
	for i := 0; i < 100; i++ {
		version := GenerateVersion()
		migrations[i] = NewBaseMigration(version, "benchmark", nil, nil)
		time.Sleep(time.Microsecond)
	}

	// Mark first 50 as applied
	applied := make([]MigrationRecord, 50)
	for i := 0; i < 50; i++ {
		applied[i] = MigrationRecord{
			Version: migrations[i].Version(),
			Batch:   1,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FilterPendingMigrations(migrations, applied)
	}
}

// BenchmarkValidateVersion benchmarks version validation
func BenchmarkValidateVersion(b *testing.B) {
	validVersion := "20240101123456"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateVersion(validVersion)
	}
}

// BenchmarkGenerateVersion benchmarks version generation
func BenchmarkGenerateVersion(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateVersion()
	}
}

// BenchmarkParseMigrationFileName benchmarks filename parsing
func BenchmarkParseMigrationFileName(b *testing.B) {
	filename := "20240101123456_create_users_table.go"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseMigrationFileName(filename)
	}
}

// BenchmarkConfigValidation benchmarks configuration validation
func BenchmarkConfigValidation(b *testing.B) {
	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

// BenchmarkConfigMerge benchmarks configuration merging
func BenchmarkConfigMerge(b *testing.B) {
	base := DefaultConfig()
	base.DatabaseURL = "postgres://localhost/base"

	override := &Config{
		DatabaseURL: "postgres://localhost/override",
		TableName:   "custom_migrations",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base.Merge(override)
	}
}

// BenchmarkMigrator_Status benchmarks migration status retrieval
func BenchmarkMigrator_Status(b *testing.B) {
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	// Register test migrations
	for i := 0; i < 10; i++ {
		version := GenerateVersion()
		m := NewBaseMigration(version, "benchmark", nil, nil)
		RegisterMigration(m)
		time.Sleep(time.Microsecond)
	}

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	migrator, _ := NewMigrator(config, adapter)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = migrator.Status(ctx)
	}
}

// BenchmarkGetMigrationStatus benchmarks status calculation
func BenchmarkGetMigrationStatus(b *testing.B) {
	ClearRegisteredMigrations()
	defer ClearRegisteredMigrations()

	// Create migrations
	migrations := make([]Migration, 50)
	for i := 0; i < 50; i++ {
		version := GenerateVersion()
		migrations[i] = NewBaseMigration(version, "benchmark", nil, nil)
		time.Sleep(time.Microsecond)
	}

	// Mark half as applied
	applied := make([]MigrationRecord, 25)
	for i := 0; i < 25; i++ {
		applied[i] = MigrationRecord{
			Version: migrations[i].Version(),
			Batch:   1,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetMigrationStatus(migrations, applied)
	}
}
