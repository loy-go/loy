package builtin

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// AuthGenerator scaffolds baseline authentication & session kit.
type AuthGenerator struct {
	modulePath string
}

// NewAuthGenerator constructs AuthGenerator.
func NewAuthGenerator(modulePath string) *AuthGenerator {
	return &AuthGenerator{modulePath: modulePath}
}

func (g *AuthGenerator) Name() string {
	return "auth"
}

func (g *AuthGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	data := NewBaseData("auth", g.modulePath, nil)
	renderer := GetRenderer()

	// 1. Password Hasher
	pwdTmpl, err := ReadTemplateContext(ctx, "auth_password.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading auth_password template: %w", err)
	}
	pwdRendered, err := renderer.RenderGo(ctx, "auth_password.go.tmpl", pwdTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering auth_password: %w", err)
	}

	// 2. Claims
	claimsTmpl, err := ReadTemplateContext(ctx, "auth_claims.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading auth_claims template: %w", err)
	}
	claimsRendered, err := renderer.RenderGo(ctx, "auth_claims.go.tmpl", claimsTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering auth_claims: %w", err)
	}

	// 3. JWT Service
	jwtTmpl, err := ReadTemplateContext(ctx, "auth_jwt.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading auth_jwt template: %w", err)
	}
	jwtRendered, err := renderer.RenderGo(ctx, "auth_jwt.go.tmpl", jwtTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering auth_jwt: %w", err)
	}

	// 4. Middleware
	mwTmpl, err := ReadTemplateContext(ctx, "auth_middleware.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading auth_middleware template: %w", err)
	}
	mwRendered, err := renderer.RenderGo(ctx, "auth_middleware.go.tmpl", mwTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering auth_middleware: %w", err)
	}

	return []model.Artifact{
		{
			Path:        "internal/platform/auth/password.go",
			Content:     pwdRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        "internal/platform/auth/claims.go",
			Content:     claimsRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        "internal/platform/auth/jwt.go",
			Content:     jwtRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
		{
			Path:        "internal/platform/auth/middleware.go",
			Content:     mwRendered,
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
		},
	}, nil
}
