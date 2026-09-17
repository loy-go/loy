package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// GRPCGenerator creates Proto contract and gRPC transport server adapter.
type GRPCGenerator struct {
	modulePath string
}

// NewGRPCGenerator constructs GRPCGenerator.
func NewGRPCGenerator(modulePath string) *GRPCGenerator {
	return &GRPCGenerator{modulePath: modulePath}
}

func (g *GRPCGenerator) Name() string {
	return "grpc"
}

func (g *GRPCGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("grpc service name is required")
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

	// 1. Proto contract
	protoTmpl, err := ReadTemplateContext(ctx, "grpc_proto.proto.tmpl")
	if err != nil {
		return nil, err
	}
	protoRendered, err := renderer.Render(ctx, "grpc_proto.proto.tmpl", protoTmpl, data)
	if err != nil {
		return nil, err
	}

	// 2. Server Adapter
	serverTmpl, err := ReadTemplateContext(ctx, "grpc_server.go.tmpl")
	if err != nil {
		return nil, err
	}
	serverRendered, err := renderer.RenderGo(ctx, "grpc_server.go.tmpl", serverTmpl, data)
	if err != nil {
		return nil, err
	}

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("proto/%s/v1/%s.proto", data.FeaturePkg, data.Snake),
			Content:     protoRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        fmt.Sprintf("internal/%s/transport/grpc/server.go", data.FeaturePkg),
			Content:     serverRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
