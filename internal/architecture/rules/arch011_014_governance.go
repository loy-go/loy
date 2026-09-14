package rules

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/uloydev/loy/internal/architecture"
	"github.com/uloydev/loy/internal/diagnostics"
)

// RuleArch011 flags service locator / DI container usage.
// e.g. uber/dig, google/wire, sarulabs/di, or locator functions.
type RuleArch011 struct{}

func (r *RuleArch011) ID() string          { return "ARCH-011" }
func (r *RuleArch011) Description() string { return "Service locators and dynamic DI containers are forbidden" }

var forbiddenDIMarkers = []string{
	"go.uber.org/dig",
	"github.com/sarulabs/di",
	"github.com/golobby/container",
}

func (r *RuleArch011) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	var violations []architecture.Violation

	for _, file := range a.Files {
		if file.AST == nil {
			continue
		}
		for _, imp := range file.AST.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			for _, forbidden := range forbiddenDIMarkers {
				if strings.HasPrefix(path, forbidden) {
					line := 1
					if file.FileSet != nil {
						line = file.FileSet.Position(imp.Pos()).Line
					}
					violations = append(violations, architecture.Violation{
						RuleID:       r.ID(),
						Code:         diagnostics.CodeArchServiceLocator,
						Message:      fmt.Sprintf("forbidden service locator/container package %s", path),
						Detail:       "Loy mandates explicit constructor injection in internal/app/wiring.go per ADR-003",
						Hint:         "wire components manually with New... constructors",
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

// RuleArch012 flags package-level mutable state.
// Uses AST to detect exported package-level var declarations that are not sync.Mutex or errors.
type RuleArch012 struct{}

func (r *RuleArch012) ID() string          { return "ARCH-012" }
func (r *RuleArch012) Description() string { return "Package-level mutable state is prohibited" }

func (r *RuleArch012) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	var violations []architecture.Violation

	for _, file := range a.Files {
		if file.AST == nil {
			continue
		}
		// Skip test files for package-level mock vars
		if strings.HasSuffix(file.Path, "_test.go") {
			continue
		}

		for _, decl := range file.AST.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				continue
			}

			for _, spec := range genDecl.Specs {
				valSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for _, name := range valSpec.Names {
					// Disallow package-level mutable vars (e.g. VarDB, VarCache, or unexported state vars)
					// Exceptions:
					// - Sentinel errors (Err... or err...)
					// - Injected build metadata (Version, Commit, Date, BuildDate in version package)
					// - Package-level read-only lookup definitions and regexes
					isSentinelError := strings.HasPrefix(name.Name, "Err") || strings.HasPrefix(name.Name, "err")
					isBuildMeta := name.Name == "Version" || name.Name == "Commit" || name.Name == "Date" || name.Name == "BuildDate"
					isRegex := strings.HasSuffix(name.Name, "Regex") || strings.HasSuffix(name.Name, "Re")
					isLookupTable := strings.HasSuffix(name.Name, "Matrix") || strings.HasSuffix(name.Name, "Map") || strings.HasSuffix(name.Name, "Rules") || strings.HasSuffix(name.Name, "Plurals") || strings.HasSuffix(name.Name, "Singulars") || strings.HasSuffix(name.Name, "Drivers") || strings.HasSuffix(name.Name, "Imports") || strings.HasSuffix(name.Name, "Markers") || strings.HasSuffix(name.Name, "Packages") || strings.HasSuffix(name.Name, "Colors") || name.Name == "colors"
					isEmbedFS := false
					if ident, ok := valSpec.Type.(*ast.Ident); ok && ident.Name == "FS" {
						isEmbedFS = true
					} else if sel, ok := valSpec.Type.(*ast.SelectorExpr); ok && sel.Sel.Name == "FS" {
						isEmbedFS = true
					}

					if isSentinelError || isBuildMeta || isRegex || isLookupTable || isEmbedFS {
						continue
					}

					lower := strings.ToLower(name.Name)
					isStateVar := strings.Contains(lower, "db") ||
						strings.Contains(lower, "client") ||
						strings.Contains(lower, "pool") ||
						strings.Contains(lower, "conn") ||
						strings.Contains(lower, "repo") ||
						strings.Contains(lower, "service") ||
						strings.Contains(lower, "cache") ||
						strings.Contains(lower, "instance")

					isStar := false
					if _, ok := valSpec.Type.(*ast.StarExpr); ok {
						isStar = true
					}

					if token.IsExported(name.Name) || isStateVar || isStar {
						line := 1
						if file.FileSet != nil {
							line = file.FileSet.Position(name.Pos()).Line
						}
						violations = append(violations, architecture.Violation{
							RuleID:       r.ID(),
							Code:         diagnostics.CodeArchGlobalMutableState,
							Message:      fmt.Sprintf("package-level mutable variable %s declared", name.Name),
							Detail:       "no global mutable state per ADR-003 and AGENTS.md invariant",
							Hint:         "pass state explicitly via struct fields or context",
							File:         file.Path,
							Line:         line,
							Suppressible: true,
						})
					}
				}
			}
		}
	}

	return violations
}

// RuleArch013 validates workspace boundaries: App -> App imports prohibited. Non-suppressible.
type RuleArch013 struct{}

func (r *RuleArch013) ID() string          { return "ARCH-013" }
func (r *RuleArch013) Description() string { return "Workspace app-to-app dependencies are prohibited" }

func (r *RuleArch013) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	var violations []architecture.Violation

	for _, edge := range a.Graph.AllEdges() {
		// App -> App cross import
		fromIsApp := strings.Contains(edge.From, "/apps/") || strings.HasPrefix(edge.From, "apps/")
		toIsApp := strings.Contains(edge.To, "/apps/") || strings.HasPrefix(edge.To, "apps/")

		if fromIsApp && toIsApp {
			// Extract app names
			fromApp := extractApp(edge.From)
			toApp := extractApp(edge.To)
			if fromApp != "" && toApp != "" && fromApp != toApp {
				violations = append(violations, architecture.Violation{
					RuleID:       r.ID(),
					Code:         diagnostics.CodeArchWorkspaceBoundary,
					Message:      fmt.Sprintf("workspace app %s imports another app %s", fromApp, toApp),
					Detail:       "monorepo applications must be decoupled; share code via internal packages instead",
					Hint:         "move shared code to packages/ or a shared module",
					File:         edge.File,
					Line:         edge.Line,
					Suppressible: false,
				})
			}
		}
	}

	return violations
}

func extractApp(pkg string) string {
	parts := strings.Split(pkg, "/")
	for i, p := range parts {
		if p == "apps" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// RuleArch014 verifies generated artifact location/ownership.
// Checks if generated loy files are located outside allowed generated regions/paths.
type RuleArch014 struct{}

func (r *RuleArch014) ID() string          { return "ARCH-014" }
func (r *RuleArch014) Description() string { return "Generated code location and artifact ownership" }

func (r *RuleArch014) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	var violations []architecture.Violation

	for _, file := range a.Files {
		if file.AST == nil {
			continue
		}
		// Inspect comment header for generated marker
		isGenerated := false
		for _, cg := range file.Comments {
			text := cg.Text()
			if strings.Contains(text, "Code generated by loy") || strings.Contains(text, "DO NOT EDIT") {
				isGenerated = true
				break
			}
		}

		if isGenerated {
			clean := filepath.ToSlash(file.Path)
			rel := clean
			if a.ModuleName != "" && strings.Contains(clean, a.ModuleName) {
				idx := strings.Index(clean, a.ModuleName) + len(a.ModuleName)
				rel = clean[idx:]
			}
			rel = strings.TrimPrefix(rel, "/")

			// Generated files should be inside internal/, pkg/, or cmd/, not directly in root domain/ or repository/
			if (strings.HasPrefix(rel, "domain/") || strings.HasPrefix(rel, "repository/") || strings.HasPrefix(rel, "service/")) && !strings.HasPrefix(rel, "internal/") {
				violations = append(violations, architecture.Violation{
					RuleID:       r.ID(),
					Code:         diagnostics.CodeArchArtifactOwnership,
					Message:      fmt.Sprintf("generated file %s placed in unmanaged root domain path", file.Path),
					Detail:       "generated files must follow canonical internal/... project layout",
					Hint:         "generate code into internal/domain, internal/transport, or internal/infrastructure",
					File:         file.Path,
					Line:         1,
					Suppressible: true,
				})
			}
		}
	}

	return violations
}
