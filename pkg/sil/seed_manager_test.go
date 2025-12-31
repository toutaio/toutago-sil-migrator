package sil

import (
	"context"
	"testing"
)

func TestNewSeedManager(t *testing.T) {
	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	manager, err := NewSeedManager(config, adapter)
	if err != nil {
		t.Fatalf("NewSeedManager() error: %v", err)
	}

	if manager == nil {
		t.Error("NewSeedManager() returned nil")
	}
}

func TestNewSeedManager_NilConfig(t *testing.T) {
	adapter := newMockAdapter()

	_, err := NewSeedManager(nil, adapter)
	if err == nil {
		t.Error("Expected error for nil config")
	}
}

func TestNewSeedManager_NilAdapter(t *testing.T) {
	config := DefaultConfig()

	_, err := NewSeedManager(config, nil)
	if err == nil {
		t.Error("Expected error for nil adapter")
	}
}

func TestSeedManager_SeedAll_NoSeeders(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	manager, _ := NewSeedManager(config, adapter)
	ctx := context.Background()

	// Should not error with no seeders
	err := manager.SeedAll(ctx)
	if err != nil {
		t.Errorf("SeedAll() with no seeders error: %v", err)
	}
}

func TestSeedManager_SeedAll_WithSeeders(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	config.Environment = "test"
	adapter := newMockAdapter()

	// Register test seeders
	executed := make(map[string]bool)

	seeder1 := NewBaseSeeder("users", func(ctx context.Context, adapter DatabaseAdapter) error {
		executed["users"] = true
		return nil
	})

	seeder2 := NewBaseSeeder("posts", func(ctx context.Context, adapter DatabaseAdapter) error {
		executed["posts"] = true
		return nil
	}).SetDependencies("users")

	RegisterSeeder(seeder1)
	RegisterSeeder(seeder2)

	manager, _ := NewSeedManager(config, adapter)
	ctx := context.Background()

	err := manager.SeedAll(ctx)
	if err != nil {
		t.Fatalf("SeedAll() error: %v", err)
	}

	// Verify both seeders ran
	if !executed["users"] {
		t.Error("users seeder did not run")
	}

	if !executed["posts"] {
		t.Error("posts seeder did not run")
	}
}

func TestSeedManager_Seed_SpecificSeeders(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	executed := make(map[string]bool)

	seeder1 := NewBaseSeeder("seeder1", func(ctx context.Context, adapter DatabaseAdapter) error {
		executed["seeder1"] = true
		return nil
	})

	seeder2 := NewBaseSeeder("seeder2", func(ctx context.Context, adapter DatabaseAdapter) error {
		executed["seeder2"] = true
		return nil
	})

	RegisterSeeder(seeder1)
	RegisterSeeder(seeder2)

	manager, _ := NewSeedManager(config, adapter)
	ctx := context.Background()

	// Seed only seeder1
	err := manager.Seed(ctx, "seeder1")
	if err != nil {
		t.Fatalf("Seed() error: %v", err)
	}

	if !executed["seeder1"] {
		t.Error("seeder1 did not run")
	}

	if executed["seeder2"] {
		t.Error("seeder2 should not have run")
	}
}

func TestSeedManager_Seed_NonexistentSeeder(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	manager, _ := NewSeedManager(config, adapter)
	ctx := context.Background()

	err := manager.Seed(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent seeder")
	}
}

func TestSeedManager_Status(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	config.Environment = "test"
	adapter := newMockAdapter()

	seeder1 := NewBaseSeeder("seeder1", nil)
	seeder2 := NewBaseSeeder("seeder2", nil).SetDependencies("seeder1")

	RegisterSeeder(seeder1)
	RegisterSeeder(seeder2)

	manager, _ := NewSeedManager(config, adapter)
	ctx := context.Background()

	statuses, err := manager.Status(ctx)
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}

	if len(statuses) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statuses))
	}

	// Verify order (seeder1 should be before seeder2 due to dependency)
	if statuses[0].Name != "seeder1" {
		t.Errorf("Expected seeder1 first, got %s", statuses[0].Name)
	}

	if statuses[1].Name != "seeder2" {
		t.Errorf("Expected seeder2 second, got %s", statuses[1].Name)
	}
}

func TestSeedManager_EnvironmentFiltering(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	config.Environment = "production"
	adapter := newMockAdapter()

	executed := make(map[string]bool)

	devSeeder := NewBaseSeeder("dev_only", func(ctx context.Context, adapter DatabaseAdapter) error {
		executed["dev_only"] = true
		return nil
	}).SetEnvironments("development")

	prodSeeder := NewBaseSeeder("prod_only", func(ctx context.Context, adapter DatabaseAdapter) error {
		executed["prod_only"] = true
		return nil
	}).SetEnvironments("production")

	allSeeder := NewBaseSeeder("all_envs", func(ctx context.Context, adapter DatabaseAdapter) error {
		executed["all_envs"] = true
		return nil
	}) // No environments = runs in all

	RegisterSeeder(devSeeder)
	RegisterSeeder(prodSeeder)
	RegisterSeeder(allSeeder)

	manager, _ := NewSeedManager(config, adapter)
	ctx := context.Background()

	err := manager.SeedAll(ctx)
	if err != nil {
		t.Fatalf("SeedAll() error: %v", err)
	}

	if executed["dev_only"] {
		t.Error("dev_only should not run in production")
	}

	if !executed["prod_only"] {
		t.Error("prod_only should run in production")
	}

	if !executed["all_envs"] {
		t.Error("all_envs should run in all environments")
	}
}

func TestSeedManager_ShouldRunCheck(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	executed := make(map[string]bool)

	shouldNotRun := NewBaseSeeder("skip_me", func(ctx context.Context, adapter DatabaseAdapter) error {
		executed["skip_me"] = true
		return nil
	}).SetShouldRun(func(ctx context.Context, adapter DatabaseAdapter) (bool, error) {
		return false, nil
	})

	shouldRun := NewBaseSeeder("run_me", func(ctx context.Context, adapter DatabaseAdapter) error {
		executed["run_me"] = true
		return nil
	}).SetShouldRun(func(ctx context.Context, adapter DatabaseAdapter) (bool, error) {
		return true, nil
	})

	RegisterSeeder(shouldNotRun)
	RegisterSeeder(shouldRun)

	manager, _ := NewSeedManager(config, adapter)
	ctx := context.Background()

	err := manager.SeedAll(ctx)
	if err != nil {
		t.Fatalf("SeedAll() error: %v", err)
	}

	if executed["skip_me"] {
		t.Error("skip_me should not have run")
	}

	if !executed["run_me"] {
		t.Error("run_me should have run")
	}
}

func TestSeedManager_DependencyOrder(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	config := DefaultConfig()
	config.DatabaseURL = "postgres://localhost/test"
	adapter := newMockAdapter()

	executionOrder := []string{}

	seeder1 := NewBaseSeeder("first", func(ctx context.Context, adapter DatabaseAdapter) error {
		executionOrder = append(executionOrder, "first")
		return nil
	})

	seeder2 := NewBaseSeeder("second", func(ctx context.Context, adapter DatabaseAdapter) error {
		executionOrder = append(executionOrder, "second")
		return nil
	}).SetDependencies("first")

	seeder3 := NewBaseSeeder("third", func(ctx context.Context, adapter DatabaseAdapter) error {
		executionOrder = append(executionOrder, "third")
		return nil
	}).SetDependencies("second")

	RegisterSeeder(seeder3)
	RegisterSeeder(seeder1)
	RegisterSeeder(seeder2)

	manager, _ := NewSeedManager(config, adapter)
	ctx := context.Background()

	err := manager.SeedAll(ctx)
	if err != nil {
		t.Fatalf("SeedAll() error: %v", err)
	}

	if len(executionOrder) != 3 {
		t.Errorf("Expected 3 seeders to run, got %d", len(executionOrder))
	}

	if executionOrder[0] != "first" || executionOrder[1] != "second" || executionOrder[2] != "third" {
		t.Errorf("Wrong execution order: %v", executionOrder)
	}
}
