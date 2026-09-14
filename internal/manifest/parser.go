package manifest

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/loy-go/loy/internal/diagnostics"
	"gopkg.in/yaml.v3"
)

// Parser parses and validates raw loy.yaml bytes into a Manifest.
type Parser struct{}

// NewParser creates a new manifest Parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParseStrict strictly decodes YAML bytes, rejecting unknown fields and emitting diagnostics.
func (p *Parser) ParseStrict(filename string, data []byte) (*Manifest, *diagnostics.Diagnostic) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeConfigSyntaxError,
			Message:  "manifest file is empty",
			Hint:     "generate a valid manifest using 'loy init'",
			File:     filename,
		}
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var m Manifest
	if err := decoder.Decode(&m); err != nil {
		diag := parseYAMLError(filename, err)
		return nil, diag
	}

	// Extra content check
	var extra interface{}
	if err := decoder.Decode(&extra); err == nil {
		return nil, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     diagnostics.CodeConfigSyntaxError,
			Message:  "unexpected multiple documents in manifest",
			Hint:     "loy.yaml must contain exactly one YAML document",
			File:     filename,
		}
	} else if !errors.Is(err, io.EOF) {
		diag := parseYAMLError(filename, err)
		return nil, diag
	}

	return &m, nil
}

func parseYAMLError(filename string, err error) *diagnostics.Diagnostic {
	diag := &diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     diagnostics.CodeConfigSyntaxError,
		Message:  err.Error(),
		File:     filename,
		Hint:     "check loy.yaml syntax and ensure all fields are recognized",
	}

	var typeErr *yaml.TypeError
	if errors.As(err, &typeErr) {
		diag.Hint = "remove unrecognized fields from loy.yaml"
	}

	errStr := err.Error()
	re := regexp.MustCompile(`line (\d+)(?:: column (\d+))?: (.+)`)
	matches := re.FindStringSubmatch(errStr)
	if len(matches) > 3 {
		if lineNum, convErr := strconv.Atoi(matches[1]); convErr == nil {
			diag.Line = lineNum
		}
		if matches[2] != "" {
			if colNum, convErr := strconv.Atoi(matches[2]); convErr == nil {
				diag.Column = colNum
			}
		}
		diag.Message = strings.TrimSpace(matches[3])
	}

	return diag
}
