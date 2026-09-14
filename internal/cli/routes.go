package cli

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
	"github.com/spf13/cobra"
)

// RouteInfo models a discovered application endpoint.
type RouteInfo struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
	File    string `json:"file"`
	Line    int    `json:"line"`
}

func newRoutesCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "routes [path]",
		Short: "Inspect and list all registered HTTP and WebSocket routes",
		Long:  `loy routes statically inspects application source code to discover all registered endpoints and route bindings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir := "."
			if len(args) > 0 {
				rootDir = args[0]
			}

			cliOpts := GetOptions(cmd.Context())
			var routes []RouteInfo

			fset := token.NewFileSet()
			_ = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					name := info.Name()
					if name == ".git" || name == "vendor" || (strings.HasPrefix(name, ".") && name != ".") {
						return filepath.SkipDir
					}
					return nil
				}
				if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}

				node, err := parser.ParseFile(fset, path, nil, 0)
				if err != nil {
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
					case "Get":
						isHTTPMethod, httpMethod = true, "GET"
					case "Post":
						isHTTPMethod, httpMethod = true, "POST"
					case "Put":
						isHTTPMethod, httpMethod = true, "PUT"
					case "Delete":
						isHTTPMethod, httpMethod = true, "DELETE"
					case "Patch":
						isHTTPMethod, httpMethod = true, "PATCH"
					case "Options":
						isHTTPMethod, httpMethod = true, "OPTIONS"
					case "Head":
						isHTTPMethod, httpMethod = true, "HEAD"
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

			if cliOpts.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(routes)
			}

			if len(routes) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No route registrations discovered.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
			_, _ = fmt.Fprintln(w, "METHOD\tPATH\tHANDLER\tSOURCE")
			for _, r := range routes {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s:%d\n", r.Method, r.Path, r.Handler, r.File, r.Line)
			}
			return w.Flush()
		},
	}
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
