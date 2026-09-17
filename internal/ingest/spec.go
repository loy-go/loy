package ingest

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParsedEntity represents an ingested schema ready for Loy code generation.
type ParsedEntity struct {
	Name   string
	Fields []string
}

// FieldsString returns space-separated field specifications for Loy generators.
func (e ParsedEntity) FieldsString() string {
	return strings.Join(e.Fields, " ")
}

type openAPISpec struct {
	Components struct {
		Schemas map[string]schemaDef `yaml:"schemas" json:"schemas"`
	} `yaml:"components" json:"components"`
	Definitions map[string]schemaDef `yaml:"definitions" json:"definitions"`
	Title       string               `yaml:"title" json:"title"`
	Properties  map[string]propDef   `yaml:"properties" json:"properties"`
	Required    []string             `yaml:"required" json:"required"`
}

type schemaDef struct {
	Type       string             `yaml:"type" json:"type"`
	Properties map[string]propDef `yaml:"properties" json:"properties"`
	Required   []string           `yaml:"required" json:"required"`
}

type propDef struct {
	Type   string   `yaml:"type" json:"type"`
	Format string   `yaml:"format" json:"format"`
	Enum   []string `yaml:"enum" json:"enum"`
}

// IngestSpec parses OpenAPI 3.0/3.1 or JSON Schema definitions into Loy entities.
func IngestSpec(content []byte) ([]ParsedEntity, error) {
	var spec openAPISpec

	// Attempt JSON unmarshal first, then fallback to YAML
	if err := json.Unmarshal(content, &spec); err != nil {
		if yamlErr := yaml.Unmarshal(content, &spec); yamlErr != nil {
			return nil, fmt.Errorf("parsing spec as JSON or YAML: %w", yamlErr)
		}
	}

	var results []ParsedEntity

	// 1. OpenAPI 3.x components/schemas
	if len(spec.Components.Schemas) > 0 {
		for name, sDef := range spec.Components.Schemas {
			entity := parseSchemaDef(name, sDef)
			if len(entity.Fields) > 0 {
				results = append(results, entity)
			}
		}
	}

	// 2. Swagger 2.0 / JSON Schema definitions
	if len(spec.Definitions) > 0 {
		for name, sDef := range spec.Definitions {
			entity := parseSchemaDef(name, sDef)
			if len(entity.Fields) > 0 {
				results = append(results, entity)
			}
		}
	}

	// 3. Top-level schema
	if len(spec.Properties) > 0 {
		name := spec.Title
		if name == "" {
			name = "entity"
		}
		entity := parseSchemaDef(name, schemaDef{
			Properties: spec.Properties,
			Required:   spec.Required,
		})
		if len(entity.Fields) > 0 {
			results = append(results, entity)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no valid schemas or properties found in spec")
	}

	return results, nil
}

func parseSchemaDef(name string, sDef schemaDef) ParsedEntity {
	reqMap := make(map[string]bool)
	for _, r := range sDef.Required {
		reqMap[r] = true
	}

	var fields []string
	for pName, pDef := range sDef.Properties {
		// Skip standard metadata managed by Loy
		low := strings.ToLower(pName)
		if low == "id" || low == "created_at" || low == "updated_at" || low == "org_id" || low == "public_id" {
			continue
		}

		loyType := mapPropType(pDef)
		if len(pDef.Enum) > 0 {
			loyType = fmt.Sprintf("enum(%s)", strings.Join(pDef.Enum, ","))
		}
		var modifiers []string
		if reqMap[pName] {
			modifiers = append(modifiers, "required")
		}

		fieldSpec := fmt.Sprintf("%s:%s", pName, loyType)
		if len(modifiers) > 0 {
			fieldSpec += ":" + strings.Join(modifiers, ":")
		}
		fields = append(fields, fieldSpec)
	}

	return ParsedEntity{
		Name:   name,
		Fields: fields,
	}
}

func mapPropType(p propDef) string {
	switch p.Type {
	case "integer":
		if p.Format == "int64" {
			return "int64"
		}
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	case "string":
		if p.Format == "date-time" || p.Format == "date" {
			return "time"
		}
		return "string"
	default:
		return "string"
	}
}
