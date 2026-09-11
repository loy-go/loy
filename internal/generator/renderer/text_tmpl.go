package renderer

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"github.com/uloydev/loy/internal/diagnostics"
	"github.com/uloydev/loy/internal/generator/naming"
)

// TextRenderer implements Renderer using Go standard text/template.
type TextRenderer struct {
	customFuncs template.FuncMap
}

// NewTextRenderer constructs a TextRenderer with standard Loy template helpers.
func NewTextRenderer(extraFuncs template.FuncMap) *TextRenderer {
	funcs := template.FuncMap{
		"pascal":     naming.ToPascalCase,
		"camel":      naming.ToCamelCase,
		"snake":      naming.ToSnakeCase,
		"kebab":      naming.ToKebabCase,
		"pkg":        naming.ToPackageName,
		"plural":     naming.Pluralize,
		"singular":   naming.Singularize,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"title":      strings.Title, // nolint:staticcheck
		"trim":       strings.TrimSpace,
	}

	for k, v := range extraFuncs {
		funcs[k] = v
	}

	return &TextRenderer{
		customFuncs: funcs,
	}
}

// Render processes standard template text without go/format.
func (r *TextRenderer) Render(ctx context.Context, name string, tmplContent string, data any) ([]byte, error) {
	tmpl, err := template.New(name).Funcs(r.customFuncs).Parse(tmplContent)
	if err != nil {
		diag := diagnostics.NewError(
			diagnostics.CodeGenTemplateError,
			fmt.Sprintf("parsing template %q: %v", name, err),
		)
		diag.Detail = err.Error()
		diag.Hint = "Check template syntax and helper function names"
		return nil, &diag
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		diag := diagnostics.NewError(
			diagnostics.CodeGenTemplateError,
			fmt.Sprintf("executing template %q: %v", name, err),
		)
		diag.Detail = err.Error()
		diag.Hint = "Verify template data model fields"
		return nil, &diag
	}

	return buf.Bytes(), nil
}

// RenderGo renders template and runs go/format (gofmt) on the resulting Go source code.
func (r *TextRenderer) RenderGo(ctx context.Context, name string, tmplContent string, data any) ([]byte, error) {
	rendered, err := r.Render(ctx, name, tmplContent, data)
	if err != nil {
		return nil, err
	}

	formatted, err := format.Source(rendered)
	if err != nil {
		// Annotate syntax error with line numbers for diagnostics
		lines := strings.Split(string(rendered), "\n")
		var annotated strings.Builder
		for i, line := range lines {
			annotated.WriteString(fmt.Sprintf("%4d | %s\n", i+1, line))
		}

		diag := diagnostics.NewError(
			diagnostics.CodeGenTemplateError,
			fmt.Sprintf("gofmt failed on template %q output: %v", name, err),
		)
		diag.Detail = fmt.Sprintf("Rendered source:\n%s", annotated.String())
		diag.Hint = "Ensure generated Go source satisfies valid Go grammar"
		return nil, &diag
	}

	return formatted, nil
}
