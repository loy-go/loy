package builtin

import (
	"context"
	"embed"
	"path/filepath"
	"strings"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator/naming"
	"github.com/loy-go/loy/internal/generator/renderer"
)

//go:embed templates/*
var templateFS embed.FS

// BaseData holds common template variables for code generation.
type BaseData struct {
	ModulePath     string   // e.g. "github.com/example/app"
	Feature        string   // feature identifier, e.g. "user"
	FeaturePkg     string   // package safe name, e.g. "user"
	Pascal         string   // e.g. "User"
	Camel          string   // e.g. "user"
	Plural         string   // e.g. "Users"
	PluralLower    string   // e.g. "users"
	Snake          string   // e.g. "user"
	Kebab          string   // e.g. "user"
	Fields         []Field  // parsed fields
	Imports        []string // extra imports
	MultiTenancy   bool     // whether multi-tenancy is active
	TenantStrategy string   // "rls" or "column"
	DualID         bool     // whether dual identifier (BIGINT identity + UUID public) is active
}

// NewBaseData initializes BaseData from feature name and module path.
func NewBaseData(featureName, modulePath string, fields []Field) BaseData {
	pkgName := naming.ToPackageName(featureName)
	pascal := naming.ToPascalCase(featureName)
	plural := naming.Pluralize(pascal)

	return BaseData{
		ModulePath:  modulePath,
		Feature:     featureName,
		FeaturePkg:  pkgName,
		Pascal:      pascal,
		Camel:       naming.ToCamelCase(featureName),
		Plural:      plural,
		PluralLower: strings.ToLower(plural),
		Snake:       naming.ToSnakeCase(featureName),
		Kebab:       naming.ToKebabCase(featureName),
		Fields:      fields,
	}
}

var defaultRenderer = renderer.NewTextRenderer(map[string]any{
	"add": func(a, b int) int {
		return a + b
	},
})

// GetRenderer returns a text renderer configured with common template helpers.
func GetRenderer() *renderer.TextRenderer {
	return defaultRenderer
}

type templateResolverKey struct{}

// TemplateResolver provides template content resolution.
type TemplateResolver interface {
	ResolveTemplate(name string) (string, bool, error)
}

// FilesystemTemplateResolver resolves templates from a project filesystem directory.
type FilesystemTemplateResolver struct {
	fs  filesystem.FileSystem
	dir string
}

// NewFilesystemTemplateResolver creates a resolver pointing to a project directory (e.g. .loy/templates).
func NewFilesystemTemplateResolver(fs filesystem.FileSystem, dir string) *FilesystemTemplateResolver {
	return &FilesystemTemplateResolver{fs: fs, dir: dir}
}

// ResolveTemplate checks if custom template exists in the target directory.
func (r *FilesystemTemplateResolver) ResolveTemplate(name string) (string, bool, error) {
	if r == nil || r.fs == nil || r.dir == "" {
		return "", false, nil
	}
	targetPath := filepath.Join(r.dir, name)
	exists, err := r.fs.Exists(targetPath)
	if err != nil || !exists {
		return "", false, nil
	}
	data, err := r.fs.ReadFile(targetPath)
	if err != nil {
		return "", false, err
	}
	return string(data), true, nil
}

// WithTemplateResolver injects a TemplateResolver into context.
func WithTemplateResolver(ctx context.Context, resolver TemplateResolver) context.Context {
	return context.WithValue(ctx, templateResolverKey{}, resolver)
}

// ReadTemplate loads a template string from embedded FS.
func ReadTemplate(name string) (string, error) {
	data, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadTemplateContext loads a template string, checking context resolver overrides first before embedded FS.
func ReadTemplateContext(ctx context.Context, name string) (string, error) {
	if ctx != nil {
		if resolver, ok := ctx.Value(templateResolverKey{}).(TemplateResolver); ok && resolver != nil {
			if content, found, err := resolver.ResolveTemplate(name); err != nil {
				return "", err
			} else if found {
				return content, nil
			}
		}
	}
	return ReadTemplate(name)
}

// ListTemplates returns all available embedded template names.
func ListTemplates() ([]string, error) {
	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}
