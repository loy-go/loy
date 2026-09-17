package architecture

import (
	"context"
	"go/ast"
	"go/token"

	"github.com/loy-go/loy/internal/graph"
	"golang.org/x/tools/go/packages"
)

// FileAST represents parsed AST info for a single Go file.
type FileAST struct {
	Path     string
	PkgPath  string
	Layer    Layer
	AST      *ast.File
	FileSet  *token.FileSet
	Comments []*ast.CommentGroup
	Content  []byte
}

// Analysis provides all extracted data to Architecture Rules.
type Analysis struct {
	ModuleName    string
	Graph         *graph.Graph
	Files         []*FileAST
	Packages      []*packages.Package // Populated in Deep mode
	Classifier    *Classifier
	Suppressions  []Suppression
	IsDeep        bool
	Topology      *Topology
}

// Rule defines the contract for an architecture boundary check.
type Rule interface {
	ID() string
	Description() string
	Check(ctx context.Context, a *Analysis) []Violation
}
