package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/loy-go/loy/internal/architecture"
	"github.com/loy-go/loy/internal/diagnostics"
)

// RuleArch007 checks transport business logic leakage.
// Specifically flags transport directly importing database/sql, sqlc, or gorm drivers.
type RuleArch007 struct{}

func (r *RuleArch007) ID() string          { return "ARCH-007" }
func (r *RuleArch007) Description() string { return "Transport must not contain raw database/persistence operations" }

func (r *RuleArch007) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	var violations []architecture.Violation

	forbiddenPersistence := map[string]bool{
		"database/sql": true,
		"gorm.io/gorm": true,
	}

	for _, file := range a.Files {
		if file.Layer != architecture.LayerTransport {
			continue
		}
		if file.AST == nil {
			continue
		}
		for _, imp := range file.AST.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if forbiddenPersistence[path] || strings.Contains(path, "sqlc") {
				line := 1
				if file.FileSet != nil {
					line = file.FileSet.Position(imp.Pos()).Line
				}
				violations = append(violations, architecture.Violation{
					RuleID:       r.ID(),
					Code:         diagnostics.CodeArchTransportLogic,
					Message:      fmt.Sprintf("transport file imports persistence package %s", path),
					Detail:       "database and business persistence belong behind application services and infrastructure repositories",
					Hint:         "delegate data persistence calls to application services",
					File:         file.Path,
					Line:         line,
					Suppressible: true,
				})
			}
		}
	}

	return violations
}

// RuleArch008 flags forbidden raw imports across layers.
// e.g. Domain must never import HTTP frameworks, ORMs, or Valkey/Redis SDKs.
type RuleArch008 struct{}

func (r *RuleArch008) ID() string          { return "ARCH-008" }
func (r *RuleArch008) Description() string { return "Domain must not import external framework/driver packages" }

var forbiddenDomainImports = []string{
	"net/http",
	"database/sql",
	"github.com/gofiber/fiber",
	"github.com/gin-gonic/gin",
	"github.com/labstack/echo",
	"gorm.io/gorm",
	"github.com/redis/go-redis",
	"github.com/valkey-io/valkey-go",
	"google.golang.org/grpc",
}

func (r *RuleArch008) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	var violations []architecture.Violation

	for _, file := range a.Files {
		if file.Layer != architecture.LayerDomain || file.AST == nil {
			continue
		}
		for _, imp := range file.AST.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			for _, forbidden := range forbiddenDomainImports {
				if strings.HasPrefix(path, forbidden) {
					line := 1
					if file.FileSet != nil {
						line = file.FileSet.Position(imp.Pos()).Line
					}
					violations = append(violations, architecture.Violation{
						RuleID:       r.ID(),
						Code:         diagnostics.CodeArchForbiddenImport,
						Message:      fmt.Sprintf("domain file imports forbidden package %s", path),
						Detail:       "domain must remain pure Go without HTTP, SQL, or cache framework dependencies",
						Hint:         "move external framework interactions to transport or infrastructure",
						File:         file.Path,
						Line:         line,
						Suppressible: true,
					})
				}
			}
		}
	}

	return violations
}

// RuleArch009 checks forbidden dependency categories.
type RuleArch009 struct{}

func (r *RuleArch009) ID() string          { return "ARCH-009" }
func (r *RuleArch009) Description() string { return "Forbidden dependency categories per layer" }

func (r *RuleArch009) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	var violations []architecture.Violation

	// Application must not import web frameworks (e.g. fiber, gin)
	forbiddenAppPackages := []string{
		"github.com/gofiber/fiber",
		"github.com/gin-gonic/gin",
		"github.com/labstack/echo",
	}

	for _, file := range a.Files {
		if file.Layer != architecture.LayerApplication || file.AST == nil {
			continue
		}
		for _, imp := range file.AST.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			for _, forbidden := range forbiddenAppPackages {
				if strings.HasPrefix(path, forbidden) {
					line := 1
					if file.FileSet != nil {
						line = file.FileSet.Position(imp.Pos()).Line
					}
					violations = append(violations, architecture.Violation{
						RuleID:       r.ID(),
						Code:         diagnostics.CodeArchForbiddenCategory,
						Message:      fmt.Sprintf("application file imports web transport framework %s", path),
						Detail:       "application services must be transport-agnostic",
						Hint:         "wrap transport-specific request/response types in transport layer DTOs",
						File:         file.Path,
						Line:         line,
						Suppressible: true,
					})
				}
			}
		}
	}

	return violations
}

// RuleArch010 verifies global layer matrix direction. Non-suppressible.
type RuleArch010 struct{}

func (r *RuleArch010) ID() string          { return "ARCH-010" }
func (r *RuleArch010) Description() string { return "Strict 4-layer dependency matrix direction" }

func (r *RuleArch010) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	var violations []architecture.Violation

	for _, edge := range a.Graph.AllEdges() {
		fromNode := a.Graph.Node(edge.From)
		toNode := a.Graph.Node(edge.To)
		if fromNode == nil || toNode == nil {
			continue
		}

		fromLayer := architecture.Layer(fromNode.Layer)
		toLayer := architecture.Layer(toNode.Layer)

		// ARCH-002 through ARCH-006 govern specific single-layer violations and allow suppression.
		// ARCH-010 acts as the fallback matrix enforcer.
		if !architecture.IsAllowedDirection(fromLayer, toLayer) {
			violations = append(violations, architecture.Violation{
				RuleID:       r.ID(),
				Code:         diagnostics.CodeArchLayerDirection,
				Message:      fmt.Sprintf("illegal layer dependency: %s (%s) -> %s (%s)", edge.From, fromLayer, edge.To, toLayer),
				Detail:       "violates strict 4-layer matrix: Transport -> Application -> Domain <- Infrastructure",
				Hint:         "invert dependency with an interface or adjust package layer responsibilities",
				File:         edge.File,
				Line:         edge.Line,
				Suppressible: true,
			})
		}
	}

	return violations
}
