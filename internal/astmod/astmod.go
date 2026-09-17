package astmod

import (
	"bytes"
	"fmt"
	"go/token"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

// InsertField appends a typed field with tags into structName in the Go source.
func InsertField(src []byte, structName, fieldName, fieldType, tags string) ([]byte, error) {
	file, err := decorator.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parsing go source: %w", err)
	}

	var targetStruct *dst.StructType
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*dst.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*dst.TypeSpec)
			if ok && typeSpec.Name.Name == structName {
				st, ok := typeSpec.Type.(*dst.StructType)
				if ok {
					targetStruct = st
					break
				}
			}
		}
		if targetStruct != nil {
			break
		}
	}

	if targetStruct == nil {
		return nil, fmt.Errorf("struct %q not found in source", structName)
	}

	if targetStruct.Fields == nil {
		targetStruct.Fields = &dst.FieldList{}
	}

	// Check for duplicate field
	for _, f := range targetStruct.Fields.List {
		for _, n := range f.Names {
			if n.Name == fieldName {
				return nil, fmt.Errorf("field %q already exists in struct %q", fieldName, structName)
			}
		}
	}

	// Parse fieldType expression using a synthetic Go snippet
	dummyCode := fmt.Sprintf("package dummy\ntype _ %s\n", fieldType)
	dummyFile, err := decorator.Parse([]byte(dummyCode))
	if err != nil {
		return nil, fmt.Errorf("parsing field type %q: %w", fieldType, err)
	}

	dummyGenDecl, ok := dummyFile.Decls[0].(*dst.GenDecl)
	if !ok || len(dummyGenDecl.Specs) == 0 {
		return nil, fmt.Errorf("invalid type expression %q", fieldType)
	}
	dummyTypeSpec, ok := dummyGenDecl.Specs[0].(*dst.TypeSpec)
	if !ok {
		return nil, fmt.Errorf("invalid type expression %q", fieldType)
	}
	typeExpr := dst.Clone(dummyTypeSpec.Type).(dst.Expr)

	newField := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent(fieldName)},
		Type:  typeExpr,
	}

	if tags != "" {
		tagVal := tags
		if !strings.HasPrefix(tagVal, "`") {
			tagVal = "`" + tagVal + "`"
		}
		newField.Tag = &dst.BasicLit{
			Kind:  token.STRING,
			Value: tagVal,
		}
	}
	newField.Decorations().Before = dst.NewLine

	targetStruct.Fields.List = append(targetStruct.Fields.List, newField)

	var buf bytes.Buffer
	if err := decorator.Fprint(&buf, file); err != nil {
		return nil, fmt.Errorf("formatting modified ast: %w", err)
	}

	return buf.Bytes(), nil
}

// AddRoute adds an HTTP route registration to the route registration function in the Go source.
func AddRoute(src []byte, method, path, handler string) ([]byte, error) {
	file, err := decorator.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parsing go source: %w", err)
	}

	var targetFunc *dst.FuncDecl
	// Find route registration function
	for _, decl := range file.Decls {
		fn, ok := decl.(*dst.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		name := strings.ToLower(fn.Name.Name)
		if strings.Contains(name, "route") || strings.Contains(name, "register") || strings.Contains(name, "setup") || strings.Contains(name, "init") {
			targetFunc = fn
			break
		}
	}

	if targetFunc == nil {
		// Fallback to first function with a body
		for _, decl := range file.Decls {
			fn, ok := decl.(*dst.FuncDecl)
			if ok && fn.Body != nil {
				targetFunc = fn
				break
			}
		}
	}

	if targetFunc == nil {
		return nil, fmt.Errorf("no function found to insert route into")
	}

	// Detect receiver identifier (e.g. router, r, group, mux, app)
	receiver := "router"
	isMux := false

	// Check existing statements
	for _, stmt := range targetFunc.Body.List {
		exprStmt, ok := stmt.(*dst.ExprStmt)
		if !ok {
			continue
		}
		call, ok := exprStmt.X.(*dst.CallExpr)
		if !ok {
			continue
		}
		sel, ok := call.Fun.(*dst.SelectorExpr)
		if !ok {
			continue
		}
		ident, ok := sel.X.(*dst.Ident)
		if ok {
			receiver = ident.Name
			if sel.Sel.Name == "HandleFunc" || sel.Sel.Name == "Handle" {
				isMux = true
			}
			break
		}
	}

	// If receiver wasn't detected from statements, inspect func parameters
	if targetFunc.Type.Params != nil {
		for _, p := range targetFunc.Type.Params.List {
			if len(p.Names) > 0 {
				pName := p.Names[0].Name
				typeStr := fmt.Sprintf("%v", p.Type)
				if strings.Contains(typeStr, "ServeMux") || strings.Contains(typeStr, "mux") {
					receiver = pName
					isMux = true
					break
				}
				if strings.Contains(typeStr, "Router") || strings.Contains(typeStr, "fiber") || strings.Contains(typeStr, "gin") || strings.Contains(typeStr, "chi") {
					receiver = pName
					break
				}
			}
		}
	}

	method = strings.ToUpper(method)
	var routeStmtCode string
	if isMux {
		routePattern := fmt.Sprintf("%s %s", method, path)
		routeStmtCode = fmt.Sprintf("package dummy\nfunc _() {\n\t%s.HandleFunc(%s, %s)\n}\n", receiver, strconv.Quote(routePattern), handler)
	} else {
		methodTitle := strings.ToUpper(method[:1]) + strings.ToLower(method[1:])
		routeStmtCode = fmt.Sprintf("package dummy\nfunc _() {\n\t%s.%s(%s, %s)\n}\n", receiver, methodTitle, strconv.Quote(path), handler)
	}

	dummyFile, err := decorator.Parse([]byte(routeStmtCode))
	if err != nil {
		return nil, fmt.Errorf("generating route ast: %w", err)
	}

	dummyFunc := dummyFile.Decls[0].(*dst.FuncDecl)
	newStmt := dst.Clone(dummyFunc.Body.List[0]).(dst.Stmt)
	newStmt.Decorations().Before = dst.NewLine

	targetFunc.Body.List = append(targetFunc.Body.List, newStmt)

	var buf bytes.Buffer
	if err := decorator.Fprint(&buf, file); err != nil {
		return nil, fmt.Errorf("formatting modified ast: %w", err)
	}

	return buf.Bytes(), nil
}

// BindDependency inserts a constructor call (e.g. depName, _ := providerFunc) into the wiring function.
func BindDependency(src []byte, providerFunc, depName string) ([]byte, error) {
	file, err := decorator.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parsing go source: %w", err)
	}

	var targetFunc *dst.FuncDecl
	for _, decl := range file.Decls {
		fn, ok := decl.(*dst.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		name := strings.ToLower(fn.Name.Name)
		if strings.Contains(name, "wire") || strings.Contains(name, "app") || strings.Contains(name, "init") || strings.Contains(name, "new") {
			targetFunc = fn
			break
		}
	}

	if targetFunc == nil {
		for _, decl := range file.Decls {
			fn, ok := decl.(*dst.FuncDecl)
			if ok && fn.Body != nil {
				targetFunc = fn
				break
			}
		}
	}

	if targetFunc == nil {
		return nil, fmt.Errorf("no wiring function found in source")
	}

	// Check if depName is already bound in targetFunc
	for _, stmt := range targetFunc.Body.List {
		assignStmt, ok := stmt.(*dst.AssignStmt)
		if !ok {
			continue
		}
		for _, lhs := range assignStmt.Lhs {
			ident, ok := lhs.(*dst.Ident)
			if ok && ident.Name == depName {
				// Already declared
				return src, nil
			}
		}
	}

	dummyCode := fmt.Sprintf("package dummy\nfunc _() {\n\t%s, _ := %s\n\t_ = %s\n}\n", depName, providerFunc, depName)
	dummyFile, err := decorator.Parse([]byte(dummyCode))
	if err != nil {
		return nil, fmt.Errorf("generating dependency ast: %w", err)
	}

	dummyFunc := dummyFile.Decls[0].(*dst.FuncDecl)
	newStmt1 := dst.Clone(dummyFunc.Body.List[0]).(dst.Stmt)
	newStmt2 := dst.Clone(dummyFunc.Body.List[1]).(dst.Stmt)
	newStmt1.Decorations().Before = dst.NewLine

	// Insert before the last return statement if present
	inserted := false
	for i := len(targetFunc.Body.List) - 1; i >= 0; i-- {
		if _, ok := targetFunc.Body.List[i].(*dst.ReturnStmt); ok {
			targetFunc.Body.List = append(targetFunc.Body.List[:i], append([]dst.Stmt{newStmt1, newStmt2}, targetFunc.Body.List[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		targetFunc.Body.List = append(targetFunc.Body.List, newStmt1, newStmt2)
	}

	var buf bytes.Buffer
	if err := decorator.Fprint(&buf, file); err != nil {
		return nil, fmt.Errorf("formatting modified ast: %w", err)
	}

	return buf.Bytes(), nil
}
