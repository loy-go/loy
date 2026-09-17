package builtin

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
	"github.com/loy-go/loy/internal/generator/naming"
)

var validSQLTypeRegex = regexp.MustCompile(`^[a-zA-Z0-9_(), ]+$`)

// MigrationData holds template variables for migration generation.
type MigrationData struct {
	Name       string
	Snake      string
	Table      string
	Column     string
	ColumnType string
}

// MigrationGenerator scaffolds safe database migrations.
type MigrationGenerator struct {
	modulePath string
}

// NewMigrationGenerator constructs MigrationGenerator.
func NewMigrationGenerator(modulePath string) *MigrationGenerator {
	return &MigrationGenerator{modulePath: modulePath}
}

func (g *MigrationGenerator) Name() string {
	return "migration"
}

func (g *MigrationGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("migration name is required")
	}

	recipe := "raw"
	if input.Args != nil && input.Args["recipe"] != "" {
		recipe = input.Args["recipe"]
	}
	table := "items"
	if input.Args != nil && input.Args["table"] != "" {
		table = input.Args["table"]
	}
	if !ValidIdentifierRegex.MatchString(table) {
		return nil, fmt.Errorf("invalid table identifier %q", table)
	}

	column := ""
	if input.Args != nil && input.Args["column"] != "" {
		column = input.Args["column"]
	} else if input.Args != nil && input.Args["col"] != "" {
		column = input.Args["col"]
	}
	if column != "" && !ValidIdentifierRegex.MatchString(column) {
		return nil, fmt.Errorf("invalid column identifier %q", column)
	}
	if column == "" {
		column = naming.ToSnakeCase(input.Name)
	}

	colType := "TEXT"
	if input.Args != nil && input.Args["type"] != "" {
		colType = strings.TrimSpace(input.Args["type"])
	}
	if !validSQLTypeRegex.MatchString(colType) || strings.Contains(colType, ";") || strings.Contains(colType, "--") || strings.Contains(colType, "/*") {
		return nil, fmt.Errorf("invalid SQL type %q", colType)
	}

	var tmplName string
	switch recipe {
	case "index-concurrent", "index":
		tmplName = "migration_index_concurrent.sql.tmpl"
	case "shadow-column", "shadow":
		tmplName = "migration_shadow_column.sql.tmpl"
	default:
		tmplName = "migration_raw.sql.tmpl"
	}

	tmpl, err := ReadTemplateContext(ctx, tmplName)
	if err != nil {
		return nil, fmt.Errorf("reading migration template: %w", err)
	}

	data := MigrationData{
		Name:       input.Name,
		Snake:      naming.ToSnakeCase(input.Name),
		Table:      naming.ToSnakeCase(table),
		Column:     naming.ToSnakeCase(column),
		ColumnType: colType,
	}

	renderer := GetRenderer()
	rendered, err := renderer.Render(ctx, tmplName, tmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering migration: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102150405")
	migPath := fmt.Sprintf("migrations/%s_%s.sql", timestamp, data.Snake)

	return []model.Artifact{
		{
			Path:        migPath,
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
