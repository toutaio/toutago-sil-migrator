package sil

import (
	"context"
	"fmt"
	"sync"
)

// BaseSeeder provides a base implementation of the Seeder interface.
type BaseSeeder struct {
	name         string
	dependencies []string
	environments []string
	seedFunc     func(ctx context.Context, adapter DatabaseAdapter) error
	shouldRunFunc func(ctx context.Context, adapter DatabaseAdapter) (bool, error)
}

// NewBaseSeeder creates a new base seeder.
func NewBaseSeeder(name string, seedFunc func(ctx context.Context, adapter DatabaseAdapter) error) *BaseSeeder {
	return &BaseSeeder{
		name:         name,
		dependencies: []string{},
		environments: []string{},
		seedFunc:     seedFunc,
		shouldRunFunc: func(ctx context.Context, adapter DatabaseAdapter) (bool, error) {
			return true, nil // Default: always run
		},
	}
}

// Name returns the seeder name.
func (s *BaseSeeder) Name() string {
	return s.name
}

// Dependencies returns the seeder dependencies.
func (s *BaseSeeder) Dependencies() []string {
	return s.dependencies
}

// SetDependencies sets the seeder dependencies.
func (s *BaseSeeder) SetDependencies(deps ...string) *BaseSeeder {
	s.dependencies = deps
	return s
}

// Environments returns the environments where this seeder should run.
func (s *BaseSeeder) Environments() []string {
	return s.environments
}

// SetEnvironments sets the environments where this seeder should run.
func (s *BaseSeeder) SetEnvironments(envs ...string) *BaseSeeder {
	s.environments = envs
	return s
}

// Seed executes the seeding operation.
func (s *BaseSeeder) Seed(ctx context.Context, adapter DatabaseAdapter) error {
	if s.seedFunc == nil {
		return fmt.Errorf("seed function not defined for seeder: %s", s.name)
	}
	return s.seedFunc(ctx, adapter)
}

// ShouldRun checks if the seeder should run.
func (s *BaseSeeder) ShouldRun(ctx context.Context, adapter DatabaseAdapter) (bool, error) {
	if s.shouldRunFunc == nil {
		return true, nil
	}
	return s.shouldRunFunc(ctx, adapter)
}

// SetShouldRun sets a custom function to determine if the seeder should run.
func (s *BaseSeeder) SetShouldRun(fn func(ctx context.Context, adapter DatabaseAdapter) (bool, error)) *BaseSeeder {
	s.shouldRunFunc = fn
	return s
}

// Seeder registry
var (
	registeredSeeders = make(map[string]Seeder)
	seedersMutex      sync.RWMutex
)

// RegisterSeeder registers a seeder globally.
func RegisterSeeder(seeder Seeder) {
	seedersMutex.Lock()
	defer seedersMutex.Unlock()

	name := seeder.Name()
	if _, exists := registeredSeeders[name]; exists {
		panic(fmt.Sprintf("seeder %s already registered", name))
	}

	registeredSeeders[name] = seeder
}

// GetRegisteredSeeders returns all registered seeders.
func GetRegisteredSeeders() []Seeder {
	seedersMutex.RLock()
	defer seedersMutex.RUnlock()

	seeders := make([]Seeder, 0, len(registeredSeeders))
	for _, seeder := range registeredSeeders {
		seeders = append(seeders, seeder)
	}
	return seeders
}

// GetSeederByName returns a seeder by name.
func GetSeederByName(name string) (Seeder, bool) {
	seedersMutex.RLock()
	defer seedersMutex.RUnlock()

	seeder, exists := registeredSeeders[name]
	return seeder, exists
}

// ClearRegisteredSeeders clears all registered seeders (for testing).
func ClearRegisteredSeeders() {
	seedersMutex.Lock()
	defer seedersMutex.Unlock()

	registeredSeeders = make(map[string]Seeder)
}

// SortSeedersByDependencies sorts seeders in dependency order using topological sort.
func SortSeedersByDependencies(seeders []Seeder) ([]Seeder, error) {
	// Build dependency graph
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	seederMap := make(map[string]Seeder)

	// Initialize
	for _, seeder := range seeders {
		name := seeder.Name()
		seederMap[name] = seeder
		graph[name] = []string{}
		inDegree[name] = 0
	}

	// Build graph
	for _, seeder := range seeders {
		name := seeder.Name()
		for _, dep := range seeder.Dependencies() {
			// Check if dependency exists
			if _, exists := seederMap[dep]; !exists {
				return nil, fmt.Errorf("seeder %s depends on %s which is not registered", name, dep)
			}
			graph[dep] = append(graph[dep], name)
			inDegree[name]++
		}
	}

	// Topological sort using Kahn's algorithm
	var sorted []Seeder
	var queue []string

	// Find all nodes with no incoming edges
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	for len(queue) > 0 {
		// Remove node from queue
		current := queue[0]
		queue = queue[1:]

		sorted = append(sorted, seederMap[current])

		// For each neighbor, reduce in-degree
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles
	if len(sorted) != len(seeders) {
		return nil, fmt.Errorf("circular dependency detected in seeders")
	}

	return sorted, nil
}
