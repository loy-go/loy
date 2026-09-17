package astmod

import (
	"bytes"
	"fmt"
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/loy-go/loy/internal/filesystem"
)

// FixARCH005 auto-remediates an application-to-concrete-infrastructure violation.
// It removes the infrastructure import from the application file and creates
// a consumer interface in the application package.
func FixARCH005(fs filesystem.FileSystem, filePath, infraPkgPath string) error {
	data, err := fs.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	file, err := decorator.Parse(data)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", filePath, err)
	}

	// 1. Remove the infra import
	removedPkgIdent := ""
	var newDecls []dst.Decl
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*dst.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			newDecls = append(newDecls, decl)
			continue
		}

		var newSpecs []dst.Spec
		for _, spec := range genDecl.Specs {
			impSpec, ok := spec.(*dst.ImportSpec)
			if !ok {
				newSpecs = append(newSpecs, spec)
				continue
			}
			impPath := strings.Trim(impSpec.Path.Value, `"`)
			if impPath == infraPkgPath || strings.HasSuffix(impPath, "/"+infraPkgPath) {
				if impSpec.Name != nil {
					removedPkgIdent = impSpec.Name.Name
				} else {
					parts := strings.Split(impPath, "/")
					removedPkgIdent = parts[len(parts)-1]
				}
				// Remove this import
				continue
			}
			newSpecs = append(newSpecs, spec)
		}

		if len(newSpecs) > 0 {
			genDecl.Specs = newSpecs
			newDecls = append(newDecls, genDecl)
		}
	}
	file.Decls = newDecls

	// 2. Scan struct fields and constructor parameters to replace `*removedPkgIdent.Type` with consumer interface
	interfaceName := "Repository"

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *dst.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					ts, ok := spec.(*dst.TypeSpec)
					if ok {
						st, ok := ts.Type.(*dst.StructType)
						if ok && st.Fields != nil {
							for _, f := range st.Fields.List {
								if isPkgSelector(f.Type, removedPkgIdent) {
									f.Type = dst.NewIdent(interfaceName)
								}
							}
						}
					}
				}
			}
		case *dst.FuncDecl:
			if d.Type != nil && d.Type.Params != nil {
				for _, p := range d.Type.Params.List {
					if isPkgSelector(p.Type, removedPkgIdent) {
						p.Type = dst.NewIdent(interfaceName)
					}
				}
			}
		}
	}

	// 3. Ensure consumer interface is declared in the file
	hasInterface := false
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*dst.GenDecl)
		if ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				ts, ok := spec.(*dst.TypeSpec)
				if ok && ts.Name.Name == interfaceName {
					hasInterface = true
					break
				}
			}
		}
	}

	if !hasInterface {
		dummyCode := fmt.Sprintf("package dummy\n// %s is the consumer interface defined by the application layer.\ntype %s interface {}\n", interfaceName, interfaceName)
		dummyFile, err := decorator.Parse([]byte(dummyCode))
		if err == nil {
			dummyDecl := dummyFile.Decls[0]
			dummyDecl.Decorations().Before = dst.NewLine
			file.Decls = append(file.Decls, dummyDecl)
		}
	}

	var buf bytes.Buffer
	if err := decorator.Fprint(&buf, file); err != nil {
		return fmt.Errorf("formatting modified ast: %w", err)
	}

	return fs.WriteFile(filePath, buf.Bytes(), 0644)
}

func isPkgSelector(expr dst.Expr, pkgName string) bool {
	if pkgName == "" {
		return false
	}
	switch t := expr.(type) {
	case *dst.StarExpr:
		return isPkgSelector(t.X, pkgName)
	case *dst.SelectorExpr:
		ident, ok := t.X.(*dst.Ident)
		return ok && ident.Name == pkgName
	default:
		return false
	}
}
