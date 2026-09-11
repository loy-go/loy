package builtin

import (
	"context"
	"fmt"
	"time"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/model"
)

// CRUDGenerator creates full vertical stack: migration, query, domain, repo, svc, handler, wiring.
type CRUDGenerator struct {
	modulePath string
}

// NewCRUDGenerator constructs CRUDGenerator.
func NewCRUDGenerator(modulePath string) *CRUDGenerator {
	return &CRUDGenerator{modulePath: modulePath}
}

func (g *CRUDGenerator) Name() string {
	return "crud"
}

func (g *CRUDGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("crud entity name is required")
	}

	rawFields := input.Args["fields"]
	var rawList []string
	if rawFields != "" {
		rawList = splitRawArgs(rawFields)
	}

	fields, err := ParseFields(rawList)
	if err != nil {
		return nil, err
	}

	data := NewBaseData(input.Name, g.modulePath, fields)
	renderer := GetRenderer()

	// 1. Migration artifact
	migTmpl, err := ReadTemplate("migration.sql.tmpl")
	if err != nil {
		return nil, err
	}
	migRendered, err := renderer.Render(ctx, "migration.sql.tmpl", migTmpl, data)
	if err != nil {
		return nil, err
	}
	timestamp := time.Now().UTC().Format("20060102150405")
	migPath := fmt.Sprintf("migrations/%s_create_%s_table.sql", timestamp, data.PluralLower)

	// 2. sqlc Query artifact
	queryTmpl, err := ReadTemplate("query.sql.tmpl")
	if err != nil {
		return nil, err
	}
	queryRendered, err := renderer.Render(ctx, "query.sql.tmpl", queryTmpl, data)
	if err != nil {
		return nil, err
	}
	queryPath := fmt.Sprintf("queries/%s.sql", data.PluralLower)

	// 3. Feature bundle (model, repo, svc, handler, test, wiring)
	featGen := NewFeatureGenerator(g.modulePath)
	featArts, err := featGen.Generate(ctx, input)
	if err != nil {
		return nil, err
	}

	artifacts := []model.Artifact{
		{
			Path:        migPath,
			Content:     migRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        queryPath,
			Content:     queryRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}
	artifacts = append(artifacts, featArts...)

	return artifacts, nil
}
