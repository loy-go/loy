package ingest

import (
	"context"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/generator/naming"
)

func TestIngestSpec_OpenAPI3_YAML(t *testing.T) {
	yamlSpec := `
openapi: 3.0.0
info:
  title: Petstore API
  version: 1.0.0
components:
  schemas:
    Pet:
      type: object
      required:
        - name
        - status
      properties:
        id:
          type: integer
          format: int64
        public_id:
          type: string
        name:
          type: string
        tag:
          type: string
        price:
          type: number
        is_available:
          type: boolean
        status:
          type: string
          enum: [available, pending, sold]
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
`
	entities, err := IngestSpec([]byte(yamlSpec))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(entities))
	}

	pet := entities[0]
	if pet.Name != "Pet" {
		t.Errorf("expected entity name 'Pet', got '%s'", pet.Name)
	}

	fieldsStr := pet.FieldsString()
	// Should skip id, public_id, created_at, updated_at
	if strings.Contains(fieldsStr, "id:") || strings.Contains(fieldsStr, "created_at:") {
		t.Errorf("fields should not contain metadata fields, got: %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "name:string:required") {
		t.Errorf("expected name:string:required in fields, got: %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "price:float") {
		t.Errorf("expected price:float in fields, got: %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "is_available:bool") {
		t.Errorf("expected is_available:bool in fields, got: %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "status:enum(available,pending,sold):required") {
		t.Errorf("expected status with enum in fields, got: %s", fieldsStr)
	}
}

func TestIngestSpec_JSON(t *testing.T) {
	jsonSpec := `{
		"definitions": {
			"Customer": {
				"type": "object",
				"required": ["email"],
				"properties": {
					"id": {"type": "integer"},
					"email": {"type": "string"},
					"age": {"type": "integer"},
					"birth_date": {"type": "string", "format": "date"}
				}
			}
		}
	}`
	entities, err := IngestSpec([]byte(jsonSpec))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(entities))
	}
	customer := entities[0]
	if customer.Name != "Customer" {
		t.Errorf("expected name Customer, got %s", customer.Name)
	}
	fieldsStr := customer.FieldsString()
	if !strings.Contains(fieldsStr, "email:string:required") {
		t.Errorf("expected email:string:required, got %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "age:int") {
		t.Errorf("expected age:int, got %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "birth_date:time") {
		t.Errorf("expected birth_date:time, got %s", fieldsStr)
	}
}

func TestIngestSpec_TopLevelProperties(t *testing.T) {
	spec := `{
		"title": "Article",
		"type": "object",
		"required": ["title", "content"],
		"properties": {
			"title": {"type": "string"},
			"content": {"type": "string"}
		}
	}`
	entities, err := IngestSpec([]byte(spec))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entities) != 1 || entities[0].Name != "Article" {
		t.Fatalf("expected Article entity, got %+v", entities)
	}
}

func TestIngestSpec_Errors(t *testing.T) {
	// Malformed
	_, err := IngestSpec([]byte(`{not valid json or yaml: `))
	if err == nil {
		t.Error("expected error for malformed spec, got nil")
	}

	// Empty
	_, err = IngestSpec([]byte(`{"openapi": "3.0.0"}`))
	if err == nil {
		t.Error("expected error for spec with no schemas, got nil")
	}
}

func TestIngestDDL(t *testing.T) {
	ddl := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		public_id UUID NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE,
		full_name TEXT NOT NULL,
		age INT,
		is_active BOOLEAN NOT NULL DEFAULT true,
		score NUMERIC(10, 2),
		created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
		CONSTRAINT uq_users_email UNIQUE (email)
	);

	CREATE TABLE categories (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		code VARCHAR(50) UNIQUE,
		created_at TIMESTAMP
	);
	`

	entities, err := IngestDDL(ddl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(entities))
	}

	// users -> user (singular)
	user := entities[0]
	if user.Name != "user" {
		t.Errorf("expected entity user, got %s", user.Name)
	}
	userFields := user.FieldsString()
	if strings.Contains(userFields, "id:") || strings.Contains(userFields, "public_id:") {
		t.Errorf("should skip id and public_id, got %s", userFields)
	}
	if !strings.Contains(userFields, "email:string:required:unique") {
		t.Errorf("expected email:string:required:unique, got %s", userFields)
	}
	if !strings.Contains(userFields, "full_name:string:required") {
		t.Errorf("expected full_name:string:required, got %s", userFields)
	}
	if !strings.Contains(userFields, "age:int") {
		t.Errorf("expected age:int, got %s", userFields)
	}
	if !strings.Contains(userFields, "is_active:bool:required") {
		t.Errorf("expected is_active:bool:required, got %s", userFields)
	}
	if !strings.Contains(userFields, "score:float") {
		t.Errorf("expected score:float, got %s", userFields)
	}

	// categories -> category
	cat := entities[1]
	if cat.Name != "category" {
		t.Errorf("expected category, got %s", cat.Name)
	}
	catFields := cat.FieldsString()
	if !strings.Contains(catFields, "name:string:required") {
		t.Errorf("expected name:string:required, got %s", catFields)
	}
	if !strings.Contains(catFields, "code:string:unique") {
		t.Errorf("expected code:string:unique, got %s", catFields)
	}
}

func TestIngestDDL_Invalid(t *testing.T) {
	_, err := IngestDDL("SELECT * FROM foo;")
	if err == nil {
		t.Error("expected error for non-CREATE TABLE SQL, got nil")
	}
}

func TestIngestPostgres_ConnectionFailure(t *testing.T) {
	ctx := context.Background()
	_, err := IngestPostgres(ctx, "postgres://invalid:invalid@127.0.0.1:54321/nonexistent?sslmode=disable", nil)
	if err == nil {
		t.Error("expected error connecting to invalid postgres, got nil")
	}
}

func TestBuildEntitiesFromColumns(t *testing.T) {
	records := []ColumnRecord{
		{TableName: "orders", ColumnName: "id", DataType: "bigint", IsNullable: "NO"},
		{TableName: "orders", ColumnName: "public_id", DataType: "uuid", IsNullable: "NO"},
		{TableName: "orders", ColumnName: "order_num", DataType: "varchar(100)", IsNullable: "NO"},
		{TableName: "orders", ColumnName: "total_amount", DataType: "numeric(12,2)", IsNullable: "NO"},
		{TableName: "orders", ColumnName: "is_shipped", DataType: "boolean", IsNullable: "YES"},
		{TableName: "orders", ColumnName: "shipped_at", DataType: "timestamptz", IsNullable: "YES"},
		{TableName: "orders", ColumnName: "item_count", DataType: "integer", IsNullable: "NO"},
		{TableName: "orders", ColumnName: "tracking_code", DataType: "uuid", IsNullable: "YES"},
		{TableName: "orders", ColumnName: "created_at", DataType: "timestamptz", IsNullable: "NO"},
	}

	entities := BuildEntitiesFromColumns(records)
	if len(entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(entities))
	}
	ord := entities[0]
	if ord.Name != "order" {
		t.Errorf("expected entity name 'order', got %s", ord.Name)
	}
	fieldsStr := ord.FieldsString()
	if strings.Contains(fieldsStr, "id:") || strings.Contains(fieldsStr, "created_at:") {
		t.Errorf("expected metadata to be skipped, got: %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "order_num:string:required") {
		t.Errorf("expected order_num:string:required, got: %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "total_amount:float:required") {
		t.Errorf("expected total_amount:float:required, got: %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "is_shipped:bool") {
		t.Errorf("expected is_shipped:bool, got: %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "shipped_at:time") {
		t.Errorf("expected shipped_at:time, got: %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "item_count:int:required") {
		t.Errorf("expected item_count:int:required, got: %s", fieldsStr)
	}
	if !strings.Contains(fieldsStr, "tracking_code:string") {
		t.Errorf("expected tracking_code:string, got: %s", fieldsStr)
	}
}

func TestSingularize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"categories", "category"},
		{"boxes", "box"},
		{"users", "user"},
		{"status", "status"},
		{"boss", "boss"},
		{"item", "item"},
	}

	for _, tc := range tests {
		got := naming.Singularize(tc.input)
		if got != tc.expected {
			t.Errorf("singularize(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}
