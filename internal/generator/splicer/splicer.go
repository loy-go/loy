package splicer

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"strings"

	"github.com/uloydev/loy/internal/diagnostics"
)

// RegionInfo stores parsed coordinates of a managed region.
type RegionInfo struct {
	Name       string
	StartLine  int // 1-based index
	EndLine    int // 1-based index
	StartFound bool
	EndFound   bool
}

// Splicer handles comment-based managed region code insertion.
type Splicer struct{}

// New constructs a Splicer.
func New() *Splicer {
	return &Splicer{}
}

// ParseRegions scans content for // loy:region:<name> and // loy:endregion markers.
// Validates uniqueness and non-nested constraints per ADR-014.
func (s *Splicer) ParseRegions(content []byte) (map[string]RegionInfo, error) {
	regions := make(map[string]RegionInfo)
	var activeRegion *RegionInfo

	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Check both Go comment style (// loy:region:) and hash comment style (# loy:region:)
		markerPrefix := ""
		markerEnd := ""
		if strings.HasPrefix(line, "// loy:region:") {
			markerPrefix = "// loy:region:"
		} else if strings.HasPrefix(line, "# loy:region:") {
			markerPrefix = "# loy:region:"
		}

		if strings.HasPrefix(line, "// loy:endregion") {
			markerEnd = "// loy:endregion"
		} else if strings.HasPrefix(line, "# loy:endregion") {
			markerEnd = "# loy:endregion"
		}

		if markerPrefix != "" {
			if activeRegion != nil {
				diag := diagnostics.NewError(
					diagnostics.CodeGenRegionCorrupt,
					fmt.Sprintf("nested region %q inside %q at line %d", line, activeRegion.Name, lineNum),
				)
				diag.Line = lineNum
				diag.Hint = "Nested regions are not permitted. Close the active region with // loy:endregion or # loy:endregion"
				return nil, &diag
			}

			regionName := strings.TrimSpace(strings.TrimPrefix(line, markerPrefix))
			if regionName == "" {
				diag := diagnostics.NewError(
					diagnostics.CodeGenRegionCorrupt,
					fmt.Sprintf("empty region marker at line %d", lineNum),
				)
				diag.Line = lineNum
				diag.Hint = "Provide a region name, e.g. // loy:region:services"
				return nil, &diag
			}

			if _, exists := regions[regionName]; exists {
				diag := diagnostics.NewError(
					diagnostics.CodeGenRegionCorrupt,
					fmt.Sprintf("duplicate region %q found at line %d", regionName, lineNum),
				)
				diag.Line = lineNum
				diag.Hint = "Region names must be unique within a file"
				return nil, &diag
			}

			activeRegion = &RegionInfo{
				Name:       regionName,
				StartLine:  lineNum,
				StartFound: true,
			}
		} else if markerEnd != "" {
			if activeRegion == nil {
				diag := diagnostics.NewError(
					diagnostics.CodeGenRegionCorrupt,
					fmt.Sprintf("unexpected %s without matching start marker at line %d", markerEnd, lineNum),
				)
				diag.Line = lineNum
				diag.Hint = "Ensure each end marker has a preceding start marker"
				return nil, &diag
			}

			activeRegion.EndLine = lineNum
			activeRegion.EndFound = true
			regions[activeRegion.Name] = *activeRegion
			activeRegion = nil
		}
	}

	if activeRegion != nil {
		diag := diagnostics.NewError(
			diagnostics.CodeGenRegionCorrupt,
			fmt.Sprintf("unclosed region %q starting at line %d", activeRegion.Name, activeRegion.StartLine),
		)
		diag.Line = activeRegion.StartLine
		diag.Hint = "Add // loy:endregion to close the region"
		return nil, &diag
	}

	return regions, nil
}

// SpliceRegion inserts entry into target region in existingContent.
// Idempotency: if trimmed entry line already exists inside region, returns content unchanged.
// If isGoSource is true, runs go/format on output.
func (s *Splicer) SpliceRegion(existingContent []byte, regionName string, entry string, isGoSource bool) ([]byte, error) {
	regions, err := s.ParseRegions(existingContent)
	if err != nil {
		return nil, err
	}

	reg, found := regions[regionName]
	if !found {
		diag := diagnostics.NewError(
			diagnostics.CodeGenRegionCorrupt,
			fmt.Sprintf("target region %q not found", regionName),
		)
		diag.Hint = fmt.Sprintf("Verify that // loy:region:%s exists in the file", regionName)
		return nil, &diag
	}

	// Split lines while preserving exact line terminators
	rawLines := strings.Split(string(existingContent), "\n")

	// Check idempotency: does the entry already exist in the region?
	if s.containsEntry(rawLines, reg.StartLine, reg.EndLine-1, entry) {
		return existingContent, nil
	}

	// Insert before reg.EndLine-1 (0-based index of EndLine is reg.EndLine - 1)
	insertIndex := reg.EndLine - 1

	// Determine indentation from end marker or line before
	indentation := "\t"
	if insertIndex < len(rawLines) {
		endLineRaw := rawLines[insertIndex]
		trimmed := strings.TrimLeft(endLineRaw, " \t")
		indentation = endLineRaw[:len(endLineRaw)-len(trimmed)]
	}

	var entryLines []string
	for _, l := range strings.Split(entry, "\n") {
		if strings.TrimSpace(l) == "" {
			entryLines = append(entryLines, "")
		} else {
			// If entry doesn't already have indentation, prepend matching indentation
			if !strings.HasPrefix(l, " ") && !strings.HasPrefix(l, "\t") {
				entryLines = append(entryLines, indentation+l)
			} else {
				entryLines = append(entryLines, l)
			}
		}
	}

	var resultLines []string
	resultLines = append(resultLines, rawLines[:insertIndex]...)
	resultLines = append(resultLines, entryLines...)
	resultLines = append(resultLines, rawLines[insertIndex:]...)

	joined := []byte(strings.Join(resultLines, "\n"))

	if isGoSource {
		formatted, err := format.Source(joined)
		if err != nil {
			diag := diagnostics.NewError(
				diagnostics.CodeGenTemplateError,
				fmt.Sprintf("gofmt failed after splicing region %q: %v", regionName, err),
			)
			diag.Detail = err.Error()
			diag.Hint = "Verify that spliced entry produces syntactically valid Go"
			return nil, &diag
		}
		return formatted, nil
	}

	return joined, nil
}

func (s *Splicer) containsEntry(rawLines []string, startLine, endLine int, entry string) bool {
	entryLines := strings.Split(strings.TrimSpace(entry), "\n")
	if len(entryLines) == 0 {
		return true
	}

	for i := 0; i < len(entryLines); i++ {
		entryLines[i] = strings.TrimSpace(entryLines[i])
	}

	// Filter out blank lines in entry comparison
	var cleanEntry []string
	for _, l := range entryLines {
		if l != "" {
			cleanEntry = append(cleanEntry, l)
		}
	}
	if len(cleanEntry) == 0 {
		return true
	}

	regionLines := make([]string, 0, endLine-startLine)
	for i := startLine; i < endLine; i++ {
		regionLines = append(regionLines, strings.TrimSpace(rawLines[i]))
	}

	// Subsequence search for cleanEntry in regionLines
	for i := 0; i <= len(regionLines)-len(cleanEntry); i++ {
		match := true
		for j := 0; j < len(cleanEntry); j++ {
			if regionLines[i+j] != cleanEntry[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}
