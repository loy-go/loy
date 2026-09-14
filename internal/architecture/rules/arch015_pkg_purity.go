package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/loy-go/loy/internal/architecture"
	"github.com/loy-go/loy/internal/diagnostics"
)

// RuleArch015 flags domain or application layers importing impure platform/pkg utilities.
// Pure platform packages (math, string, time, hashing, errors) are allowed.
// Impure platform packages importing IO/network/database drivers are forbidden in domain & application.
type RuleArch015 struct{}

func (r *RuleArch015) ID() string          { return "ARCH-015" }
func (r *RuleArch015) Description() string { return "Domain and Application must not import impure platform packages" }

var impurePackageMarkers = []string{
	"net",
	"net/http",
	"net/rpc",
	"net/smtp",
	"database/sql",
	"os",
	"syscall",
	"github.com/gofiber/fiber",
	"github.com/gin-gonic/gin",
	"github.com/labstack/echo",
	"google.golang.org/grpc",
	"gorm.io/gorm",
	"github.com/redis/go-redis",
	"github.com/valkey-io/valkey-go",
	"github.com/jackc/pgx",
}

func (r *RuleArch015) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	var violations []architecture.Violation

	// Step 1: Identify impure platform packages and what made them impure
	impurePlatformPkgs := make(map[string]string) // pkgPath -> offending import
	for _, file := range a.Files {
		if file.Layer != architecture.LayerPlatform || file.AST == nil {
			continue
		}
		for _, imp := range file.AST.Imports {
			impPath := strings.Trim(imp.Path.Value, `"`)
			for _, marker := range impurePackageMarkers {
				if impPath == marker || strings.HasPrefix(impPath, marker+"/") {
					impurePlatformPkgs[file.PkgPath] = impPath
					break
				}
			}
		}
	}

	if len(impurePlatformPkgs) == 0 {
		return nil
	}

	// Step 2: Flag any Domain or Application file importing an impure platform package
	for _, file := range a.Files {
		if (file.Layer != architecture.LayerDomain && file.Layer != architecture.LayerApplication) || file.AST == nil {
			continue
		}

		for _, imp := range file.AST.Imports {
			impPath := strings.Trim(imp.Path.Value, `"`)
			if offendingMarker, impure := impurePlatformPkgs[impPath]; impure {
				line := 1
				if file.FileSet != nil {
					line = file.FileSet.Position(imp.Pos()).Line
				}

				violations = append(violations, architecture.Violation{
					RuleID:       r.ID(),
					Code:         diagnostics.CodeArchPlatformImpure,
					Message:      fmt.Sprintf("%s package %s imports impure platform package %s (uses %s)", strings.ToLower(string(file.Layer)), file.PkgPath, impPath, offendingMarker),
					Detail:       "domain and application layers must remain clean of platform packages performing direct IO, network or database operations",
					Hint:         "move impure platform components to infrastructure or define an abstraction interface",
					File:         file.Path,
					Line:         line,
					Suppressible: true,
				})
			}
		}
	}

	return violations
}
