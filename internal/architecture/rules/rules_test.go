package rules

import (
	"context"
	"go/parser"
	"go/token"
	"testing"

	"github.com/loy-go/loy/internal/architecture"
	"github.com/loy-go/loy/internal/graph"
)

func TestRulesSuite(t *testing.T) {
	ctx := context.Background()

	t.Run("ARCH-001 cycle rule", func(t *testing.T) {
		g := graph.New()
		g.AddNode("pkg/a", "", nil)
		g.AddNode("pkg/b", "", nil)
		g.AddEdge("pkg/a", "pkg/b", "a.go", 10)
		g.AddEdge("pkg/b", "pkg/a", "b.go", 20)

		analysis := &architecture.Analysis{Graph: g}
		rule := &RuleArch001{}
		v := rule.Check(ctx, analysis)
		if len(v) != 1 {
			t.Fatalf("expected 1 violation, got %d", len(v))
		}
		if v[0].RuleID != "ARCH-001" {
			t.Errorf("expected ARCH-001, got %s", v[0].RuleID)
		}
	})

	t.Run("ARCH-002 domain to infra rule", func(t *testing.T) {
		g := graph.New()
		g.AddNode("app/internal/domain/user", "Domain", nil)
		g.AddNode("app/internal/repository/pg", "Infrastructure", nil)
		g.AddEdge("app/internal/domain/user", "app/internal/repository/pg", "user.go", 15)

		analysis := &architecture.Analysis{Graph: g}
		rule := &RuleArch002{}
		v := rule.Check(ctx, analysis)
		if len(v) != 1 {
			t.Fatalf("expected 1 violation, got %d", len(v))
		}
		if v[0].RuleID != "ARCH-002" {
			t.Errorf("expected ARCH-002, got %s", v[0].RuleID)
		}
	})

	t.Run("ARCH-008 forbidden domain import", func(t *testing.T) {
		src := `package user
import "database/sql"
type User struct { DB *sql.DB }
`
		fset := token.NewFileSet()
		fileAst, err := parser.ParseFile(fset, "user.go", src, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}

		analysis := &architecture.Analysis{
			Files: []*architecture.FileAST{
				{
					Path:    "internal/domain/user/user.go",
					Layer:   architecture.LayerDomain,
					AST:     fileAst,
					FileSet: fset,
				},
			},
		}

		rule := &RuleArch008{}
		v := rule.Check(ctx, analysis)
		if len(v) != 1 {
			t.Fatalf("expected 1 violation, got %d", len(v))
		}
		if v[0].RuleID != "ARCH-008" {
			t.Errorf("expected ARCH-008, got %s", v[0].RuleID)
		}
	})

	t.Run("ARCH-012 package mutable state", func(t *testing.T) {
		src := `package service
var GlobalDBInstance any
`
		fset := token.NewFileSet()
		fileAst, err := parser.ParseFile(fset, "service.go", src, 0)
		if err != nil {
			t.Fatal(err)
		}

		analysis := &architecture.Analysis{
			Files: []*architecture.FileAST{
				{
					Path:    "internal/service/service.go",
					AST:     fileAst,
					FileSet: fset,
				},
			},
		}

		rule := &RuleArch012{}
		v := rule.Check(ctx, analysis)
		if len(v) != 1 {
			t.Fatalf("expected 1 violation, got %d", len(v))
		}
		if v[0].RuleID != "ARCH-012" {
			t.Errorf("expected ARCH-012, got %s", v[0].RuleID)
		}
	})

	t.Run("ARCH-013 app to app workspace violation", func(t *testing.T) {
		g := graph.New()
		g.AddNode("monorepo/apps/billing", "", nil)
		g.AddNode("monorepo/apps/identity", "", nil)
		g.AddEdge("monorepo/apps/billing", "monorepo/apps/identity", "billing.go", 5)

		analysis := &architecture.Analysis{Graph: g}
		rule := &RuleArch013{}
		v := rule.Check(ctx, analysis)
		if len(v) != 1 {
			t.Fatalf("expected 1 violation, got %d", len(v))
		}
		if v[0].RuleID != "ARCH-013" {
			t.Errorf("expected ARCH-013, got %s", v[0].RuleID)
		}
	})

	t.Run("ARCH-015 platform purity", func(t *testing.T) {
		fset := token.NewFileSet()

		// 1. Impure platform package that imports net/http
		mailerSrc := `package mailer
import "net/http"
type Client struct { HTTP *http.Client }
`
		mailerAst, err := parser.ParseFile(fset, "mailer.go", mailerSrc, 0)
		if err != nil {
			t.Fatal(err)
		}

		// 2. Pure platform package that imports strings
		mathUtilSrc := `package mathutil
import "strings"
func Clean(s string) string { return strings.TrimSpace(s) }
`
		mathAst, err := parser.ParseFile(fset, "math.go", mathUtilSrc, 0)
		if err != nil {
			t.Fatal(err)
		}

		// 3. Domain file importing impure mailer and pure mathutil
		domainSrc := `package user
import (
	"app/pkg/mailer"
	"app/pkg/mathutil"
)
type User struct { ID int }
`
		domainAst, err := parser.ParseFile(fset, "user.go", domainSrc, 0)
		if err != nil {
			t.Fatal(err)
		}

		analysis := &architecture.Analysis{
			Files: []*architecture.FileAST{
				{
					Path:    "pkg/mailer/mailer.go",
					PkgPath: "app/pkg/mailer",
					Layer:   architecture.LayerPlatform,
					AST:     mailerAst,
					FileSet: fset,
				},
				{
					Path:    "pkg/mathutil/math.go",
					PkgPath: "app/pkg/mathutil",
					Layer:   architecture.LayerPlatform,
					AST:     mathAst,
					FileSet: fset,
				},
				{
					Path:    "internal/domain/user/user.go",
					PkgPath: "app/internal/domain/user",
					Layer:   architecture.LayerDomain,
					AST:     domainAst,
					FileSet: fset,
				},
			},
		}

		rule := &RuleArch015{}
		violations := rule.Check(ctx, analysis)
		if len(violations) != 1 {
			t.Fatalf("expected exactly 1 violation for impure mailer import, got %d", len(violations))
		}
		if violations[0].RuleID != "ARCH-015" {
			t.Errorf("expected ARCH-015, got %s", violations[0].RuleID)
		}
	})
}
