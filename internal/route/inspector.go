package route

import (
	"go/ast"
	"go/parser"
	"go/token"
	iofs "io/fs"
	"path/filepath"
	"strings"

	"github.com/loy-go/loy/internal/filesystem"
)

// RouteInfo models a discovered application endpoint.
type RouteInfo struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
	File    string `json:"file"`
	Line    int    `json:"line"`
}

// InspectRoutes walks the filesystem starting from rootDir and discovers registered HTTP/WS routes using AST.
func InspectRoutes(fs filesystem.FileSystem, rootDir string) ([]RouteInfo, error) {
	var routes []RouteInfo
	fset := token.NewFileSet()

	err := fs.Walk(rootDir, func(path string, d iofs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" || (strings.HasPrefix(name, ".") && name != ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		data, readErr := fs.ReadFile(path)
		if readErr != nil {
			return nil
		}

		node, parseErr := parser.ParseFile(fset, path, data, 0)
		if parseErr != nil {
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			methodName := sel.Sel.Name
			isHTTPMethod := false
			var httpMethod string

			switch methodName {
			case "Get", "GET":
				isHTTPMethod, httpMethod = true, "GET"
			case "Post", "POST":
				isHTTPMethod, httpMethod = true, "POST"
			case "Put", "PUT":
				isHTTPMethod, httpMethod = true, "PUT"
			case "Delete", "DELETE":
				isHTTPMethod, httpMethod = true, "DELETE"
			case "Patch", "PATCH":
				isHTTPMethod, httpMethod = true, "PATCH"
			case "Options", "OPTIONS":
				isHTTPMethod, httpMethod = true, "OPTIONS"
			case "Head", "HEAD":
				isHTTPMethod, httpMethod = true, "HEAD"
			case "HandleFunc":
				if len(call.Args) >= 2 {
					lit, ok := call.Args[0].(*ast.BasicLit)
					if ok && lit.Kind == token.STRING {
						pattern := strings.Trim(lit.Value, `"`)
						parts := strings.SplitN(pattern, " ", 2)
						routePath := pattern
						m := "ALL"
						if len(parts) == 2 {
							m = parts[0]
							routePath = parts[1]
						}
						handlerName := exprToString(call.Args[len(call.Args)-1])
						pos := fset.Position(call.Pos())
						routes = append(routes, RouteInfo{
							Method:  m,
							Path:    routePath,
							Handler: handlerName,
							File:    pos.Filename,
							Line:    pos.Line,
						})
					}
				}
			}

			if isHTTPMethod && len(call.Args) >= 2 {
				lit, ok := call.Args[0].(*ast.BasicLit)
				if ok && lit.Kind == token.STRING {
					routePath := strings.Trim(lit.Value, `"`)
					handlerName := exprToString(call.Args[len(call.Args)-1])
					pos := fset.Position(call.Pos())

					routes = append(routes, RouteInfo{
						Method:  httpMethod,
						Path:    routePath,
						Handler: handlerName,
						File:    pos.Filename,
						Line:    pos.Line,
					})
				}
			}

			return true
		})

		return nil
	})

	return routes, err
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	default:
		return "func"
	}
}
