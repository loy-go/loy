package builtin

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
	"github.com/loy-go/loy/internal/generator/naming"
)

// ViewData holds view template rendering context.
type ViewData struct {
	ModulePath  string
	Name        string
	Pascal      string
	Snake       string
	PackageName string
	Title       string
	IsPartial   bool
}

// ViewGenerator scaffolds Templ view components.
type ViewGenerator struct {
	modulePath string
}

// NewViewGenerator constructs ViewGenerator.
func NewViewGenerator(modulePath string) *ViewGenerator {
	return &ViewGenerator{modulePath: modulePath}
}

func (g *ViewGenerator) Name() string {
	return "view"
}

func (g *ViewGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	rawName := strings.TrimSpace(input.Name)
	if rawName == "" {
		return nil, fmt.Errorf("view name is required")
	}

	clean := filepath.ToSlash(filepath.Clean(rawName))
	if strings.HasPrefix(clean, "/") || strings.HasPrefix(clean, "../") || clean == ".." || strings.Contains(clean, "/../") {
		return nil, fmt.Errorf("view path %q escapes views directory", rawName)
	}

	isPartial := input.Args != nil && (input.Args["partial"] == "true" || input.Args["partial"] == "1")

	dir := filepath.Dir(clean)
	base := filepath.Base(clean)
	if base == "." || base == "/" || base == "" {
		return nil, fmt.Errorf("invalid view name %q", rawName)
	}

	// Validate base name adheres to identifier conventions
	safeBase := naming.ToSnakeCase(base)
	pascal := naming.ToPascalCase(base)
	title := strings.ReplaceAll(pascal, "_", " ")

	var baseDir string
	var pkgName string

	if isPartial {
		baseDir = "views/components"
		if dir == "." {
			pkgName = "components"
		} else {
			pkgName = naming.ToPackageName(filepath.Base(dir))
		}
	} else {
		baseDir = "views/pages"
		if dir == "." {
			pkgName = "pages"
		} else {
			pkgName = naming.ToPackageName(filepath.Base(dir))
		}
	}

	var artifactPath string
	if dir == "." {
		artifactPath = filepath.ToSlash(filepath.Join(baseDir, safeBase+".templ"))
	} else {
		artifactPath = filepath.ToSlash(filepath.Join(baseDir, dir, safeBase+".templ"))
	}

	data := ViewData{
		ModulePath:  g.modulePath,
		Name:        rawName,
		Pascal:      pascal,
		Snake:       safeBase,
		PackageName: pkgName,
		Title:       title,
		IsPartial:   isPartial,
	}

	tmplName := "view_page.templ.tmpl"
	if isPartial {
		tmplName = "view_component.templ.tmpl"
	}

	tmplContent, err := ReadTemplate(tmplName)
	if err != nil {
		return nil, fmt.Errorf("reading view template %s: %w", tmplName, err)
	}

	renderer := GetRenderer()
	rendered, err := renderer.Render(ctx, tmplName, tmplContent, data)
	if err != nil {
		return nil, fmt.Errorf("rendering view template: %w", err)
	}

	return []model.Artifact{
		{
			Path:        artifactPath,
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
