package renderer

import (
	"context"
)

// Renderer defines the contract for rendering code and configuration templates.
type Renderer interface {
	Render(ctx context.Context, templateName string, tmplContent string, data any) ([]byte, error)
	RenderGo(ctx context.Context, templateName string, tmplContent string, data any) ([]byte, error)
}
