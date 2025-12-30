package sil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Generator handles generation of new migration files.
type Generator struct {
	migrationsDir string
	logger        Logger
}

// NewGenerator creates a new migration generator.
func NewGenerator(migrationsDir string, logger Logger) *Generator {
	if logger == nil {
		logger = NewNoopLogger()
	}

	return &Generator{
		migrationsDir: migrationsDir,
		logger:        logger,
	}
}

// migrationTemplateData holds data for migration templates.
type migrationTemplateData struct {
	Version     string
	Description string
	StructName  string
	TableName   string
}

// Generate creates a new migration file.
func (g *Generator) Generate(description string) (string, error) {
	if description == "" {
		return "", fmt.Errorf("migration description is required")
	}

	// Ensure migrations directory exists
	if err := os.MkdirAll(g.migrationsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Generate version and filename
	version := GenerateVersion()
	filename := GenerateMigrationFileName(description)
	fullPath := filepath.Join(g.migrationsDir, filename)

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return "", fmt.Errorf("migration file already exists: %s", fullPath)
	}

	// Prepare template data
	structName := generateStructName(version, description)
	data := migrationTemplateData{
		Version:     version,
		Description: description,
		StructName:  structName,
	}

	// Generate migration file
	if err := g.generateFromTemplate(fullPath, basicMigrationTemplate, data); err != nil {
		return "", err
	}

	g.logger.Info("Created migration: %s", filename)
	return fullPath, nil
}

// GenerateCreateTable creates a migration for creating a table.
func (g *Generator) GenerateCreateTable(tableName string) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("table name is required")
	}

	description := fmt.Sprintf("create %s table", tableName)
	version := GenerateVersion()
	filename := GenerateMigrationFileName(description)
	fullPath := filepath.Join(g.migrationsDir, filename)

	// Ensure migrations directory exists
	if err := os.MkdirAll(g.migrationsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create migrations directory: %w", err)
	}

	structName := generateStructName(version, description)
	data := migrationTemplateData{
		Version:     version,
		Description: description,
		StructName:  structName,
		TableName:   tableName,
	}

	if err := g.generateFromTemplate(fullPath, createTableMigrationTemplate, data); err != nil {
		return "", err
	}

	g.logger.Info("Created create_table migration: %s", filename)
	return fullPath, nil
}

// GenerateAddColumn creates a migration for adding a column.
func (g *Generator) GenerateAddColumn(tableName, columnName string) (string, error) {
	if tableName == "" || columnName == "" {
		return "", fmt.Errorf("table name and column name are required")
	}

	description := fmt.Sprintf("add %s to %s", columnName, tableName)
	version := GenerateVersion()
	filename := GenerateMigrationFileName(description)
	fullPath := filepath.Join(g.migrationsDir, filename)

	// Ensure migrations directory exists
	if err := os.MkdirAll(g.migrationsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create migrations directory: %w", err)
	}

	structName := generateStructName(version, description)
	data := migrationTemplateData{
		Version:     version,
		Description: description,
		StructName:  structName,
		TableName:   tableName,
	}

	if err := g.generateFromTemplate(fullPath, addColumnMigrationTemplate, data); err != nil {
		return "", err
	}

	g.logger.Info("Created add_column migration: %s", filename)
	return fullPath, nil
}

// generateFromTemplate generates a file from a template.
func (g *Generator) generateFromTemplate(path string, tmpl string, data interface{}) error {
	t, err := template.New("migration").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := t.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// generateStructName generates a Go struct name from version and description.
func generateStructName(version, description string) string {
	// Clean description
	description = strings.ToLower(description)
	description = strings.ReplaceAll(description, " ", "_")

	// Title case each word
	parts := strings.Split(description, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	description = strings.Join(parts, "")
	return fmt.Sprintf("Migration_%s_%s", version, description)
}

// Migration templates

const basicMigrationTemplate = `package migrations

import (
	"context"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

func init() {
	sil.RegisterMigration(&{{.StructName}}{})
}

// {{.StructName}} {{.Description}}
type {{.StructName}} struct {
	sil.BaseMigration
}

// Version returns the migration version.
func (m *{{.StructName}}) Version() string {
	return "{{.Version}}"
}

// Description returns the migration description.
func (m *{{.StructName}}) Description() string {
	return "{{.Description}}"
}

// Up applies the migration.
func (m *{{.StructName}}) Up(adapter sil.DatabaseAdapter) error {
	ctx := context.Background()
	
	// TODO: Implement migration
	return adapter.Exec(ctx, ` + "`" + `
		-- Your SQL here
	` + "`" + `)
}

// Down reverts the migration.
func (m *{{.StructName}}) Down(adapter sil.DatabaseAdapter) error {
	ctx := context.Background()
	
	// TODO: Implement rollback
	return adapter.Exec(ctx, ` + "`" + `
		-- Your rollback SQL here
	` + "`" + `)
}
`

const createTableMigrationTemplate = `package migrations

import (
	"context"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

func init() {
	sil.RegisterMigration(&{{.StructName}}{})
}

// {{.StructName}} creates the {{.TableName}} table
type {{.StructName}} struct {
	sil.BaseMigration
}

// Version returns the migration version.
func (m *{{.StructName}}) Version() string {
	return "{{.Version}}"
}

// Description returns the migration description.
func (m *{{.StructName}}) Description() string {
	return "{{.Description}}"
}

// Up applies the migration.
func (m *{{.StructName}}) Up(adapter sil.DatabaseAdapter) error {
	ctx := context.Background()
	
	return adapter.Exec(ctx, ` + "`" + `
		CREATE TABLE {{.TableName}} (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	` + "`" + `)
}

// Down reverts the migration.
func (m *{{.StructName}}) Down(adapter sil.DatabaseAdapter) error {
	ctx := context.Background()
	
	return adapter.Exec(ctx, ` + "`" + `DROP TABLE IF EXISTS {{.TableName}}` + "`" + `)
}
`

const addColumnMigrationTemplate = `package migrations

import (
	"context"
	"github.com/toutaio/toutago-sil-migrator/pkg/sil"
)

func init() {
	sil.RegisterMigration(&{{.StructName}}{})
}

// {{.StructName}} {{.Description}}
type {{.StructName}} struct {
	sil.BaseMigration
}

// Version returns the migration version.
func (m *{{.StructName}}) Version() string {
	return "{{.Version}}"
}

// Description returns the migration description.
func (m *{{.StructName}}) Description() string {
	return "{{.Description}}"
}

// Up applies the migration.
func (m *{{.StructName}}) Up(adapter sil.DatabaseAdapter) error {
	ctx := context.Background()
	
	return adapter.Exec(ctx, ` + "`" + `
		ALTER TABLE {{.TableName}}
		ADD COLUMN column_name VARCHAR(255) NOT NULL DEFAULT ''
	` + "`" + `)
}

// Down reverts the migration.
func (m *{{.StructName}}) Down(adapter sil.DatabaseAdapter) error {
	ctx := context.Background()
	
	return adapter.Exec(ctx, ` + "`" + `
		ALTER TABLE {{.TableName}}
		DROP COLUMN column_name
	` + "`" + `)
}
`
