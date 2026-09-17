package ingest

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/loy-go/loy/internal/generator/naming"
)

var createTableRegex = regexp.MustCompile(`(?is)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([a-zA-Z0-9_]+)\s*\((.*?)\);`)
var columnDefRegex = regexp.MustCompile(`(?i)^\s*([a-zA-Z0-9_]+)\s+([a-zA-Z0-9_]+(?:\([0-9, ]+\))?)(.*)$`)

// IngestDDL parses SQL CREATE TABLE statements into Loy entities.
func IngestDDL(ddlContent string) ([]ParsedEntity, error) {
	matches := createTableRegex.FindAllStringSubmatch(ddlContent, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no CREATE TABLE statements matched in DDL")
	}

	var results []ParsedEntity

	for _, match := range matches {
		tableName := match[1]
		body := match[2]

		lines := splitDDLLines(body)
		var fields []string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(strings.ToUpper(line), "PRIMARY") ||
				strings.HasPrefix(strings.ToUpper(line), "CONSTRAINT") ||
				strings.HasPrefix(strings.ToUpper(line), "FOREIGN") ||
				strings.HasPrefix(strings.ToUpper(line), "UNIQUE") ||
				strings.HasPrefix(strings.ToUpper(line), "CHECK") {
				continue
			}

			colMatch := columnDefRegex.FindStringSubmatch(line)
			if len(colMatch) < 3 {
				continue
			}

			colName := colMatch[1]
			rawType := colMatch[2]
			rest := strings.ToUpper(colMatch[3])

			low := strings.ToLower(colName)
			if low == "id" || low == "created_at" || low == "updated_at" || low == "org_id" || low == "public_id" {
				continue
			}

			loyType := mapSQLType(rawType)
			fieldSpec := fmt.Sprintf("%s:%s", colName, loyType)
			if strings.Contains(rest, "NOT NULL") {
				fieldSpec += ":required"
			}
			if strings.Contains(rest, "UNIQUE") {
				fieldSpec += ":unique"
			}
			fields = append(fields, fieldSpec)
		}

		if len(fields) > 0 {
			results = append(results, ParsedEntity{
				Name:   naming.Singularize(tableName),
				Fields: fields,
			})
		}
	}

	return results, nil
}

func splitDDLLines(body string) []string {
	var lines []string
	var current strings.Builder
	depth := 0

	for _, r := range body {
		switch r {
		case '(':
			depth++
			current.WriteRune(r)
		case ')':
			if depth > 0 {
				depth--
			}
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				trimmed := strings.TrimSpace(current.String())
				if trimmed != "" {
					lines = append(lines, trimmed)
				}
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}
	trimmed := strings.TrimSpace(current.String())
	if trimmed != "" {
		lines = append(lines, trimmed)
	}
	return lines
}

// ColumnRecord represents a column entry from information_schema.
type ColumnRecord struct {
	TableName  string
	ColumnName string
	DataType   string
	IsNullable string
}

// BuildEntitiesFromColumns transforms a list of database column records into Loy entities.
func BuildEntitiesFromColumns(records []ColumnRecord) []ParsedEntity {
	tablesMap := make(map[string][]string)

	for _, rec := range records {
		low := strings.ToLower(rec.ColumnName)
		if low == "id" || low == "created_at" || low == "updated_at" || low == "org_id" || low == "public_id" {
			continue
		}

		loyType := mapSQLType(rec.DataType)
		fieldSpec := fmt.Sprintf("%s:%s", rec.ColumnName, loyType)
		if strings.ToUpper(rec.IsNullable) == "NO" {
			fieldSpec += ":required"
		}
		tablesMap[rec.TableName] = append(tablesMap[rec.TableName], fieldSpec)
	}

	var results []ParsedEntity
	for tbl, fields := range tablesMap {
		if len(fields) > 0 {
			results = append(results, ParsedEntity{
				Name:   naming.Singularize(tbl),
				Fields: fields,
			})
		}
	}

	return results
}

// IngestPostgres queries information_schema to reverse-engineer tables.
func IngestPostgres(ctx context.Context, dsn string, tableFilter []string) ([]ParsedEntity, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	query := `
		SELECT table_name, column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public'
	`
	var args []any
	if len(tableFilter) > 0 {
		placeholders := make([]string, len(tableFilter))
		for i, t := range tableFilter {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args = append(args, t)
		}
		query += fmt.Sprintf(" AND table_name IN (%s)", strings.Join(placeholders, ","))
	}
	query += " ORDER BY table_name, ordinal_position"

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying information_schema: %w", err)
	}
	defer rows.Close()

	var records []ColumnRecord
	for rows.Next() {
		var rec ColumnRecord
		if err := rows.Scan(&rec.TableName, &rec.ColumnName, &rec.DataType, &rec.IsNullable); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}

	return BuildEntitiesFromColumns(records), nil
}

func mapSQLType(raw string) string {
	raw = strings.ToLower(raw)
	switch {
	case strings.HasPrefix(raw, "int"), strings.HasPrefix(raw, "smallint"):
		return "int"
	case strings.HasPrefix(raw, "bigint"):
		return "int64"
	case strings.HasPrefix(raw, "bool"):
		return "bool"
	case strings.HasPrefix(raw, "float"), strings.HasPrefix(raw, "double"), strings.HasPrefix(raw, "numeric"), strings.HasPrefix(raw, "decimal"), strings.HasPrefix(raw, "real"):
		return "float"
	case strings.HasPrefix(raw, "time"), strings.HasPrefix(raw, "date"):
		return "time"
	case strings.HasPrefix(raw, "uuid"):
		return "string"
	default:
		return "string"
	}
}
