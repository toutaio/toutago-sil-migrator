package sil

import (
	"context"
	"testing"
)

func TestBaseSeeder(t *testing.T) {
	ctx := context.Background()
	adapter := newMockAdapter()

	seeder := NewBaseSeeder("test_seeder", func(ctx context.Context, adapter DatabaseAdapter) error {
		return nil
	})

	if seeder.Name() != "test_seeder" {
		t.Errorf("Expected name 'test_seeder', got '%s'", seeder.Name())
	}

	if len(seeder.Dependencies()) != 0 {
		t.Errorf("Expected no dependencies, got %d", len(seeder.Dependencies()))
	}

	if len(seeder.Environments()) != 0 {
		t.Errorf("Expected no environments, got %d", len(seeder.Environments()))
	}

	// Test seed execution
	if err := seeder.Seed(ctx, adapter); err != nil {
		t.Errorf("Seed() failed: %v", err)
	}

	// Test should run
	shouldRun, err := seeder.ShouldRun(ctx, adapter)
	if err != nil {
		t.Errorf("ShouldRun() error: %v", err)
	}
	if !shouldRun {
		t.Error("ShouldRun() should return true by default")
	}
}

func TestBaseSeeder_WithDependencies(t *testing.T) {
	seeder := NewBaseSeeder("test", nil).
		SetDependencies("dep1", "dep2")

	deps := seeder.Dependencies()
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}

	if deps[0] != "dep1" || deps[1] != "dep2" {
		t.Errorf("Dependencies not set correctly: %v", deps)
	}
}

func TestBaseSeeder_WithEnvironments(t *testing.T) {
	seeder := NewBaseSeeder("test", nil).
		SetEnvironments("development", "test")

	envs := seeder.Environments()
	if len(envs) != 2 {
		t.Errorf("Expected 2 environments, got %d", len(envs))
	}

	if envs[0] != "development" || envs[1] != "test" {
		t.Errorf("Environments not set correctly: %v", envs)
	}
}

func TestBaseSeeder_CustomShouldRun(t *testing.T) {
	ctx := context.Background()
	adapter := newMockAdapter()

	called := false
	seeder := NewBaseSeeder("test", nil).
		SetShouldRun(func(ctx context.Context, adapter DatabaseAdapter) (bool, error) {
			called = true
			return false, nil
		})

	shouldRun, err := seeder.ShouldRun(ctx, adapter)
	if err != nil {
		t.Errorf("ShouldRun() error: %v", err)
	}

	if !called {
		t.Error("Custom ShouldRun function not called")
	}

	if shouldRun {
		t.Error("Expected ShouldRun to return false")
	}
}

func TestRegisterSeeder(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	seeder := NewBaseSeeder("test_register", nil)
	RegisterSeeder(seeder)

	retrieved, exists := GetSeederByName("test_register")
	if !exists {
		t.Error("Seeder not found after registration")
	}

	if retrieved.Name() != "test_register" {
		t.Errorf("Retrieved wrong seeder: %s", retrieved.Name())
	}
}

func TestRegisterSeeder_Duplicate(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	seeder := NewBaseSeeder("duplicate", nil)
	RegisterSeeder(seeder)

	// Should panic on duplicate
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on duplicate registration")
		}
	}()

	RegisterSeeder(seeder)
}

func TestGetRegisteredSeeders(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	seeder1 := NewBaseSeeder("seeder1", nil)
	seeder2 := NewBaseSeeder("seeder2", nil)

	RegisterSeeder(seeder1)
	RegisterSeeder(seeder2)

	seeders := GetRegisteredSeeders()
	if len(seeders) != 2 {
		t.Errorf("Expected 2 seeders, got %d", len(seeders))
	}
}

func TestSortSeedersByDependencies(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	// Create seeders with dependencies
	seeder1 := NewBaseSeeder("base", nil)
	seeder2 := NewBaseSeeder("depends_on_base", nil).SetDependencies("base")
	seeder3 := NewBaseSeeder("depends_on_depends", nil).SetDependencies("depends_on_base")

	seeders := []Seeder{seeder3, seeder1, seeder2} // Unsorted

	sorted, err := SortSeedersByDependencies(seeders)
	if err != nil {
		t.Fatalf("SortSeedersByDependencies() error: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("Expected 3 sorted seeders, got %d", len(sorted))
	}

	// Verify order: base -> depends_on_base -> depends_on_depends
	if sorted[0].Name() != "base" {
		t.Errorf("Expected 'base' first, got '%s'", sorted[0].Name())
	}

	if sorted[1].Name() != "depends_on_base" {
		t.Errorf("Expected 'depends_on_base' second, got '%s'", sorted[1].Name())
	}

	if sorted[2].Name() != "depends_on_depends" {
		t.Errorf("Expected 'depends_on_depends' third, got '%s'", sorted[2].Name())
	}
}

func TestSortSeedersByDependencies_CircularDependency(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	// Create circular dependency
	seeder1 := NewBaseSeeder("a", nil).SetDependencies("b")
	seeder2 := NewBaseSeeder("b", nil).SetDependencies("a")

	seeders := []Seeder{seeder1, seeder2}

	_, err := SortSeedersByDependencies(seeders)
	if err == nil {
		t.Error("Expected error for circular dependency")
	}
}

func TestSortSeedersByDependencies_MissingDependency(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	seeder := NewBaseSeeder("test", nil).SetDependencies("nonexistent")

	seeders := []Seeder{seeder}

	_, err := SortSeedersByDependencies(seeders)
	if err == nil {
		t.Error("Expected error for missing dependency")
	}
}

func TestSortSeedersByDependencies_NoDependencies(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	seeder1 := NewBaseSeeder("a", nil)
	seeder2 := NewBaseSeeder("b", nil)
	seeder3 := NewBaseSeeder("c", nil)

	seeders := []Seeder{seeder3, seeder1, seeder2}

	sorted, err := SortSeedersByDependencies(seeders)
	if err != nil {
		t.Fatalf("SortSeedersByDependencies() error: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("Expected 3 sorted seeders, got %d", len(sorted))
	}
}

func TestSortSeedersByDependencies_ComplexGraph(t *testing.T) {
	ClearRegisteredSeeders()
	defer ClearRegisteredSeeders()

	/*
		Graph:
		a -> b -> d
		  -> c -> d
	*/
	seederA := NewBaseSeeder("a", nil)
	seederB := NewBaseSeeder("b", nil).SetDependencies("a")
	seederC := NewBaseSeeder("c", nil).SetDependencies("a")
	seederD := NewBaseSeeder("d", nil).SetDependencies("b", "c")

	seeders := []Seeder{seederD, seederC, seederB, seederA}

	sorted, err := SortSeedersByDependencies(seeders)
	if err != nil {
		t.Fatalf("SortSeedersByDependencies() error: %v", err)
	}

	if len(sorted) != 4 {
		t.Errorf("Expected 4 sorted seeders, got %d", len(sorted))
	}

	// A must be first
	if sorted[0].Name() != "a" {
		t.Errorf("Expected 'a' first, got '%s'", sorted[0].Name())
	}

	// D must be last
	if sorted[3].Name() != "d" {
		t.Errorf("Expected 'd' last, got '%s'", sorted[3].Name())
	}

	// B and C can be in any order, but both after A and before D
	names := []string{sorted[1].Name(), sorted[2].Name()}
	hasB := names[0] == "b" || names[1] == "b"
	hasC := names[0] == "c" || names[1] == "c"

	if !hasB || !hasC {
		t.Errorf("Expected b and c in middle positions, got %v", names)
	}
}
