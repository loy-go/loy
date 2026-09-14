package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// WebSocketGenerator creates WebSocket hub, client session pumps, and protocol frames.
type WebSocketGenerator struct {
	modulePath string
}

// NewWebSocketGenerator constructs WebSocketGenerator.
func NewWebSocketGenerator(modulePath string) *WebSocketGenerator {
	return &WebSocketGenerator{modulePath: modulePath}
}

func (g *WebSocketGenerator) Name() string {
	return "ws"
}

func (g *WebSocketGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("websocket endpoint name is required")
	}

	data := NewBaseData(input.Name, g.modulePath, nil)
	renderer := GetRenderer()

	// 1. Protocol Frame
	protoTmpl, err := ReadTemplate("ws_protocol.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading ws_protocol template: %w", err)
	}
	protoRendered, err := renderer.RenderGo(ctx, "ws_protocol.go.tmpl", protoTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering ws_protocol: %w", err)
	}

	// 2. Hub
	hubTmpl, err := ReadTemplate("ws_hub.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading ws_hub template: %w", err)
	}
	hubRendered, err := renderer.RenderGo(ctx, "ws_hub.go.tmpl", hubTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering ws_hub: %w", err)
	}

	// 3. Client Pumps
	clientTmpl, err := ReadTemplate("ws_client.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading ws_client template: %w", err)
	}
	clientRendered, err := renderer.RenderGo(ctx, "ws_client.go.tmpl", clientTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering ws_client: %w", err)
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/%s/transport/ws/protocol.go", data.FeaturePkg),
			Content:     protoRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        fmt.Sprintf("internal/%s/transport/ws/hub.go", data.FeaturePkg),
			Content:     hubRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        fmt.Sprintf("internal/%s/transport/ws/client.go", data.FeaturePkg),
			Content:     clientRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
