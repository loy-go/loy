package builtin

import (
	"fmt"
	"strings"

	"github.com/uloydev/loy/internal/generator/naming"
)

// Field represents a parsed field specification for models and DTOs.
type Field struct {
	Name         string   // original name, e.g. "email"
	PascalName   string   // PascalCase name, e.g. "Email"
	CamelName    string   // camelCase name, e.g. "email"
	SnakeName    string   // snake_case name, e.g. "email"
	Type         string   // Go type, e.g. "string", "int", "time.Time"
	SQLType      string   // SQL type, e.g. "TEXT", "BIGINT", "TIMESTAMPTZ"
	IsPointer    bool     // true if pointer or optional
	IsUnique     bool     // true if unique modifier specified
	IsIndexed    bool     // true if index modifier specified
	IsRequired   bool     // true if required modifier specified (default true unless optional)
	IsEnum       bool     // true if enum(...) type
	EnumValues   []string // values if enum
	JSONTag      string   // JSON struct tag
	DBTag        string   // DB struct tag
}

// ZeroValueCondition returns Go syntax for checking if field is at its zero-value.
func (f Field) ZeroValueCondition() string {
	if f.IsPointer {
		return fmt.Sprintf("r.%s == nil", f.PascalName)
	}
	switch f.Type {
	case "string":
		return fmt.Sprintf("r.%s == \"\"", f.PascalName)
	case "int", "int64", "float64":
		return fmt.Sprintf("r.%s == 0", f.PascalName)
	case "bool":
		return fmt.Sprintf("!r.%s", f.PascalName)
	case "time.Time":
		return fmt.Sprintf("r.%s.IsZero()", f.PascalName)
	case "[]byte":
		return fmt.Sprintf("len(r.%s) == 0", f.PascalName)
	default:
		return fmt.Sprintf("r.%s == \"\"", f.PascalName)
	}
}

// ParseFields parses CLI arguments in format: name:type[:modifier1:modifier2...]
// Example: ["name:string", "email:string:unique", "age:int:optional", "status:enum(draft,published)"]
func ParseFields(raw []string) ([]Field, error) {
	var fields []Field
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		parts := strings.Split(item, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid field specification %q: expected name:type[:modifiers]", item)
		}

		fieldName := parts[0]
		rawType := parts[1]
		modifiers := parts[2:]

		f := Field{
			Name:       fieldName,
			PascalName: naming.ToPascalCase(fieldName),
			CamelName:  naming.ToCamelCase(fieldName),
			SnakeName:  naming.ToSnakeCase(fieldName),
			IsRequired: true,
		}

		// Check for enum(val1,val2)
		if strings.HasPrefix(rawType, "enum(") && strings.HasSuffix(rawType, ")") {
			f.IsEnum = true
			f.Type = "string"
			f.SQLType = "VARCHAR(255)"
			inside := strings.TrimSuffix(strings.TrimPrefix(rawType, "enum("), ")")
			vals := strings.Split(inside, ",")
			for _, v := range vals {
				v = strings.TrimSpace(v)
				if v != "" {
					f.EnumValues = append(f.EnumValues, v)
				}
			}
		} else {
			goType, sqlType := mapType(rawType)
			f.Type = goType
			f.SQLType = sqlType
		}

		for _, mod := range modifiers {
			mod = strings.ToLower(strings.TrimSpace(mod))
			switch mod {
			case "unique":
				f.IsUnique = true
			case "index", "indexed":
				f.IsIndexed = true
			case "optional", "nullable":
				f.IsRequired = false
				f.IsPointer = true
			case "required":
				f.IsRequired = true
				f.IsPointer = false
			}
		}

		// Tags
		jsonSuffix := ""
		if !f.IsRequired {
			jsonSuffix = ",omitempty"
		}
		f.JSONTag = fmt.Sprintf("json:\"%s%s\"", f.SnakeName, jsonSuffix)
		f.DBTag = fmt.Sprintf("db:\"%s\"", f.SnakeName)

		fields = append(fields, f)
	}

	return fields, nil
}

func mapType(raw string) (goType string, sqlType string) {
	switch strings.ToLower(raw) {
	case "string", "text":
		return "string", "TEXT"
	case "int", "integer":
		return "int", "INTEGER"
	case "int64", "bigint":
		return "int64", "BIGINT"
	case "bool", "boolean":
		return "bool", "BOOLEAN"
	case "float", "float64", "double":
		return "float64", "DOUBLE PRECISION"
	case "time", "timestamp", "datetime":
		return "time.Time", "TIMESTAMPTZ"
	case "uuid":
		return "string", "UUID"
	case "json", "jsonb":
		return "[]byte", "JSONB"
	case "bytes", "[]byte":
		return "[]byte", "BYTEA"
	default:
		// Return raw type directly as Go type
		return raw, "TEXT"
	}
}
