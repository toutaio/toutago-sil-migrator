package sil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Loader handles loading migrations from disk.
type Loader struct {
	migrationsDir string
	logger        Logger
}

// NewLoader creates a new migration loader.
func NewLoader(migrationsDir string, logger Logger) *Loader {
	if logger == nil {
		logger = NewNoopLogger()
	}

	return &Loader{
		migrationsDir: migrationsDir,
		logger:        logger,
	}
}

// Load loads all migration files from the migrations directory.
// This implementation loads from the migration registry since Go migrations
// are compiled into the binary.
func (l *Loader) Load() ([]Migration, error) {
	l.logger.Debug("Loading migrations from registry")

	// Get migrations from registry
	migrations := GetRegisteredMigrations()

	if len(migrations) == 0 {
		l.logger.Warn("No migrations found in registry")
		return nil, ErrNoMigrationsFound
	}

	// Validate all migration versions
	for _, migration := range migrations {
		if err := ValidateVersion(migration.Version()); err != nil {
			return nil, fmt.Errorf("invalid migration %s: %w", migration.Version(), err)
		}
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, migration := range migrations {
		version := migration.Version()
		if seen[version] {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateMigrationVersion, version)
		}
		seen[version] = true
	}

	l.logger.Info("Loaded %d migrations", len(migrations))
	return migrations, nil
}

// LoadPending loads all pending migrations (not yet applied).
func (l *Loader) LoadPending(applied []MigrationRecord) ([]Migration, error) {
	all, err := l.Load()
	if err != nil {
		return nil, err
	}

	pending := FilterPendingMigrations(all, applied)
	l.logger.Debug("Found %d pending migrations out of %d total", len(pending), len(all))

	return pending, nil
}

// LoadApplied loads all applied migrations.
func (l *Loader) LoadApplied(applied []MigrationRecord) ([]Migration, error) {
	all, err := l.Load()
	if err != nil {
		return nil, err
	}

	appliedMigrations := FilterAppliedMigrations(all, applied)
	l.logger.Debug("Found %d applied migrations out of %d total", len(appliedMigrations), len(all))

	return appliedMigrations, nil
}

// DiscoverMigrationFiles discovers migration files in the migrations directory.
// This is useful for generating new migrations or checking for orphaned files.
func (l *Loader) DiscoverMigrationFiles() ([]string, error) {
	if l.migrationsDir == "" {
		return nil, ErrInvalidConfiguration("migrations directory not set")
	}

	// Check if directory exists
	info, err := os.Stat(l.migrationsDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s", l.migrationsDir)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat migrations directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("migrations path is not a directory: %s", l.migrationsDir)
	}

	var files []string
	err = filepath.WalkDir(l.migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only include .go files
		if filepath.Ext(path) == ".go" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk migrations directory: %w", err)
	}

	sort.Strings(files)
	return files, nil
}

// ParseMigrationFileName parses a migration file name and extracts version and description.
// Expected format: YYYYMMDDHHMMSS_description.go
func ParseMigrationFileName(filename string) (version, description string, err error) {
	// Remove directory path
	base := filepath.Base(filename)

	// Check if it's a .go file
	if filepath.Ext(base) != ".go" {
		return "", "", fmt.Errorf("migration file must have .go extension: %s", filename)
	}

	// Remove .go extension
	base = strings.TrimSuffix(base, ".go")

	// Pattern: YYYYMMDDHHMMSS_description
	re := regexp.MustCompile(`^(\d{14})_(.+)$`)
	matches := re.FindStringSubmatch(base)

	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid migration filename format: %s (expected: YYYYMMDDHHMMSS_description.go)", filename)
	}

	version = matches[1]
	description = matches[2]

	// Validate version
	if err := ValidateVersion(version); err != nil {
		return "", "", err
	}

	// Clean up description (replace underscores with spaces)
	description = strings.ReplaceAll(description, "_", " ")

	return version, description, nil
}

// GenerateMigrationFileName generates a migration file name.
func GenerateMigrationFileName(description string) string {
	version := GenerateVersion()

	// Clean description: lowercase, replace spaces with underscores
	description = strings.ToLower(description)
	description = strings.ReplaceAll(description, " ", "_")

	// Remove any characters that aren't alphanumeric or underscore
	re := regexp.MustCompile(`[^a-z0-9_]`)
	description = re.ReplaceAllString(description, "")

	return fmt.Sprintf("%s_%s.go", version, description)
}

// ValidateMigrationFiles validates all migration files in the directory.
func (l *Loader) ValidateMigrationFiles() error {
	files, err := l.DiscoverMigrationFiles()
	if err != nil {
		return err
	}

	versions := make(map[string]string)

	for _, file := range files {
		version, _, err := ParseMigrationFileName(file)
		if err != nil {
			l.logger.Error("Invalid migration file: %s: %v", file, err)
			return err
		}

		// Check for duplicate versions
		if existingFile, exists := versions[version]; exists {
			return fmt.Errorf("%w: version %s found in both %s and %s",
				ErrDuplicateMigrationVersion, version, existingFile, file)
		}

		versions[version] = file
	}

	l.logger.Info("Validated %d migration files", len(files))
	return nil
}

// GetMigrationPath returns the full path to a migration file.
func (l *Loader) GetMigrationPath(version string) (string, error) {
	files, err := l.DiscoverMigrationFiles()
	if err != nil {
		return "", err
	}

	for _, file := range files {
		v, _, err := ParseMigrationFileName(file)
		if err != nil {
			continue
		}
		if v == version {
			return file, nil
		}
	}

	return "", fmt.Errorf("%w: %s", ErrMigrationNotFound, version)
}

// CheckForOrphanedFiles checks if there are migration files without corresponding
// registered migrations.
func (l *Loader) CheckForOrphanedFiles() ([]string, error) {
	files, err := l.DiscoverMigrationFiles()
	if err != nil {
		return nil, err
	}

	registered := GetRegisteredMigrations()
	registeredVersions := make(map[string]bool)
	for _, m := range registered {
		registeredVersions[m.Version()] = true
	}

	var orphaned []string
	for _, file := range files {
		version, _, err := ParseMigrationFileName(file)
		if err != nil {
			continue
		}

		if !registeredVersions[version] {
			orphaned = append(orphaned, file)
		}
	}

	return orphaned, nil
}
