package builtin_test

import (
	"testing"

	"github.com/uloydev/loy/internal/generator/builtin"
)

func TestParseFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		raw       []string
		wantCount int
		assertFn  func(t *testing.T, fields []builtin.Field)
		wantErr   bool
	}{
		{
			name:      "empty list",
			raw:       []string{},
			wantCount: 0,
		},
		{
			name: "basic types",
			raw: []string{
				"name:string",
				"age:int",
				"active:bool",
				"created_at:time",
			},
			wantCount: 4,
			assertFn: func(t *testing.T, fields []builtin.Field) {
				if fields[0].Name != "name" || fields[0].PascalName != "Name" || fields[0].Type != "string" || fields[0].SQLType != "TEXT" || !fields[0].IsRequired {
					t.Fatalf("unexpected field 0: %+v", fields[0])
				}
				if fields[1].Type != "int" || fields[1].SQLType != "INTEGER" {
					t.Fatalf("unexpected field 1: %+v", fields[1])
				}
				if fields[2].Type != "bool" || fields[2].SQLType != "BOOLEAN" {
					t.Fatalf("unexpected field 2: %+v", fields[2])
				}
				if fields[3].Type != "time.Time" || fields[3].SQLType != "TIMESTAMPTZ" {
					t.Fatalf("unexpected field 3: %+v", fields[3])
				}
			},
		},
		{
			name: "modifiers unique optional index",
			raw: []string{
				"email:string:unique:required",
				"bio:string:optional",
				"category_id:int64:index",
			},
			wantCount: 3,
			assertFn: func(t *testing.T, fields []builtin.Field) {
				if !fields[0].IsUnique || !fields[0].IsRequired || fields[0].IsPointer {
					t.Fatalf("unexpected field 0: %+v", fields[0])
				}
				if fields[1].IsRequired || !fields[1].IsPointer {
					t.Fatalf("unexpected field 1: %+v", fields[1])
				}
				if !fields[2].IsIndexed || fields[2].Type != "int64" || fields[2].SQLType != "BIGINT" {
					t.Fatalf("unexpected field 2: %+v", fields[2])
				}
			},
		},
		{
			name: "enum type",
			raw: []string{
				"status:enum(draft,published,archived)",
			},
			wantCount: 1,
			assertFn: func(t *testing.T, fields []builtin.Field) {
				if !fields[0].IsEnum || fields[0].Type != "string" {
					t.Fatalf("unexpected enum field: %+v", fields[0])
				}
				if len(fields[0].EnumValues) != 3 || fields[0].EnumValues[0] != "draft" {
					t.Fatalf("unexpected enum values: %+v", fields[0].EnumValues)
				}
			},
		},
		{
			name:    "invalid format error",
			raw:     []string{"single_token"},
			wantErr: true,
		},
		{
			name:    "sql injection in field name rejected",
			raw:     []string{"col; DROP TABLE users;--:string"},
			wantErr: true,
		},
		{
			name:    "sql injection in enum value rejected",
			raw:     []string{"status:enum(active', 'injected)"},
			wantErr: true,
		},
		{
			name:    "field starting with number rejected",
			raw:     []string{"123num:int"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fields, err := builtin.ParseFields(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(fields) != tc.wantCount {
				t.Fatalf("got %d fields, want %d", len(fields), tc.wantCount)
			}
			if tc.assertFn != nil {
				tc.assertFn(t, fields)
			}
		})
	}
}
