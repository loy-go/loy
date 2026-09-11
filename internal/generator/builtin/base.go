package builtin

import (
	"embed"
	"strings"

	"github.com/uloydev/loy/internal/generator/naming"
	"github.com/uloydev/loy/internal/generator/renderer"
)

//go:embed templates/*
var templateFS embed.FS

// BaseData holds common template variables for code generation.
type BaseData struct {
	ModulePath  string   // e.g. "github.com/example/app"
	Feature     string   // feature identifier, e.g. "user"
	FeaturePkg  string   // package safe name, e.g. "user"
	Pascal      string   // e.g. "User"
	Camel       string   // e.g. "user"
	Plural      string   // e.g. "Users"
	PluralLower string   // e.g. "users"
	Snake       string   // e.g. "user"
	Kebab       string   // e.g. "user"
	Fields      []Field  // parsed fields
	Imports     []string // extra imports
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

// ReadTemplate loads a template string from embedded FS.
func ReadTemplate(name string) (string, error) {
	data, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
