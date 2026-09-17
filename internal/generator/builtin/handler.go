package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// HandlerGenerator creates HTTP transport handler and routes.
type HandlerGenerator struct {
	modulePath    string
	httpFramework string
}

// NewHandlerGenerator constructs HandlerGenerator.
func NewHandlerGenerator(modulePath string) *HandlerGenerator {
	return &HandlerGenerator{modulePath: modulePath, httpFramework: "fiber"}
}

// WithHTTPFramework sets the HTTP framework.
func (g *HandlerGenerator) WithHTTPFramework(framework string) *HandlerGenerator {
	if framework != "" {
		g.httpFramework = framework
	}
	return g
}

func (g *HandlerGenerator) Name() string {
	return "handler"
}

func (g *HandlerGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("handler name is required")
	}

	httpFw := g.httpFramework
	if input.Args != nil && input.Args["http"] != "" {
		httpFw = input.Args["http"]
	}
	if httpFw == "" {
		httpFw = "fiber"
	}

	data := NewBaseData(input.Name, g.modulePath, nil)
	var tmplName string
	switch httpFw {
	case "chi":
		tmplName = "handler_chi.go.tmpl"
	case "gin":
		tmplName = "handler_gin.go.tmpl"
	case "nethttp":
		tmplName = "handler_nethttp.go.tmpl"
	default:
		tmplName = "handler.go.tmpl"
	}

	tmpl, err := ReadTemplateContext(ctx, tmplName)
	if err != nil {
		return nil, err
	}

	renderer := GetRenderer()
	rendered, err := renderer.RenderGo(ctx, tmplName, tmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/transport/http/handler.go", data.FeaturePkg),
			Content:     rendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
