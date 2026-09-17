package typescript

import (
	"go/ast"
	"go/parser"
	"go/token"
	iofs "io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator/naming"
)

// FieldDefinition represents an AST-parsed Go struct field mapped to TypeScript.
type FieldDefinition struct {
	Name     string
	Type     string
	Optional bool
	Nullable bool
}

// TypeDefinition represents a Go struct mapped to a TypeScript interface.
type TypeDefinition struct {
	Name   string
	Fields []FieldDefinition
}

// ParseDTOs scans the given directory for Go structs (requests, resources, DTOs).
func ParseDTOs(fs filesystem.FileSystem, rootDir string) ([]TypeDefinition, error) {
	var typeDefs []TypeDefinition
	seen := make(map[string]bool)
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

		node, parseErr := parser.ParseFile(fset, path, data, parser.ParseComments)
		if parseErr != nil {
			return nil
		}

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok || structType.Fields == nil {
					continue
				}

				typeName := typeSpec.Name.Name
				if seen[typeName] {
					continue
				}

				var fields []FieldDefinition
				hasJSONTag := false

				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 {
						continue
					}

					fieldName := field.Names[0].Name
					if !unicode.IsUpper(rune(fieldName[0])) {
						continue
					}

					jsonName, optional, omitted := parseJSONTag(field.Tag)
					if omitted {
						continue
					}
					if field.Tag != nil && strings.Contains(field.Tag.Value, "json:") {
						hasJSONTag = true
					}

					if jsonName == "" {
						jsonName = naming.ToCamelCase(fieldName)
					}

					tsType, isNullable := mapASTTypeToTS(field.Type)

					fields = append(fields, FieldDefinition{
						Name:     jsonName,
						Type:     tsType,
						Optional: optional,
						Nullable: isNullable,
					})
				}

				isDTO := strings.HasSuffix(typeName, "Request") ||
					strings.HasSuffix(typeName, "Resource") ||
					strings.HasSuffix(typeName, "DTO") ||
					strings.HasSuffix(typeName, "Response") ||
					hasJSONTag

				if isDTO && len(fields) > 0 {
					seen[typeName] = true
					typeDefs = append(typeDefs, TypeDefinition{
						Name:   typeName,
						Fields: fields,
					})
				}
			}
		}

		return nil
	})

	return typeDefs, err
}

func parseJSONTag(tag *ast.BasicLit) (name string, optional bool, omitted bool) {
	if tag == nil {
		return "", false, false
	}
	raw := strings.Trim(tag.Value, "`")
	st := reflect.StructTag(raw)
	val := st.Get("json")
	if val == "-" {
		return "", false, true
	}
	if val == "" {
		return "", false, false
	}

	parts := strings.Split(val, ",")
	name = parts[0]
	for _, p := range parts[1:] {
		if p == "omitempty" {
			optional = true
		}
	}
	return name, optional, false
}

func mapASTTypeToTS(expr ast.Expr) (string, bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "string", false
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64", "byte", "rune":
			return "number", false
		case "bool":
			return "boolean", false
		case "any":
			return "any", false
		default:
			return t.Name, false
		}
	case *ast.SelectorExpr:
		pkgIdent, ok := t.X.(*ast.Ident)
		if ok {
			switch pkgIdent.Name {
			case "uuid":
				return "string", false
			case "time":
				return "string", false
			case "json":
				if t.Sel.Name == "RawMessage" {
					return "any", false
				}
			}
		}
		return t.Sel.Name, false
	case *ast.StarExpr:
		inner, _ := mapASTTypeToTS(t.X)
		return inner, true
	case *ast.ArrayType:
		inner, _ := mapASTTypeToTS(t.Elt)
		return inner + "[]", false
	case *ast.MapType:
		key, _ := mapASTTypeToTS(t.Key)
		val, _ := mapASTTypeToTS(t.Value)
		if key != "string" && key != "number" {
			key = "string"
		}
		return "Record<" + key + ", " + val + ">", false
	case *ast.InterfaceType:
		return "any", false
	default:
		return "any", false
	}
}
