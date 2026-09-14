package builtin_test

import (
	"context"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/builtin"
)

// TestCasingRegression_OSS_GEN_01 ensures uppercase and mixed-case feature names
// do not generate capitalized package import paths that cause case-collision compile errors.
func TestCasingRegression_OSS_GEN_01(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		inputName   string
		expectedPkg string
	}{
		{inputName: "User", expectedPkg: "user"},
		{inputName: "UserProfile", expectedPkg: "userprofile"},
		{inputName: "BillingAccount", expectedPkg: "billingaccount"},
		{inputName: "order_item", expectedPkg: "orderitem"},
	}

	modulePath := "testapp"
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.inputName, func(t *testing.T) {
			crudGen := builtin.NewCRUDGenerator(modulePath)
			artifacts, err := crudGen.Generate(ctx, generator.Input{
				Name: tc.inputName,
				Args: map[string]string{
					"fields": "title:string:required amount:int",
				},
			})
			if err != nil {
				t.Fatalf("CRUD generation failed for %q: %v", tc.inputName, err)
			}

			fset := token.NewFileSet()

			for _, art := range artifacts {
				// Splicing region snippets are fragments, not full files
				if art.Region != "" {
					contentStr := string(art.Content)
					if art.Region == "imports" {
						if strings.Contains(contentStr, "/internal/"+tc.inputName+"/") && tc.inputName != tc.expectedPkg {
							t.Errorf("wiring import snippet contains capitalized path: %s", contentStr)
						}
					}
					continue
				}

				// Only inspect Go files
				if !strings.HasSuffix(art.Path, ".go") {
					continue
				}

				// Verify artifact path is strictly lowercase package name
				expectedPrefix := "internal/" + tc.expectedPkg + "/"
				if !strings.HasPrefix(art.Path, expectedPrefix) {
					t.Errorf("artifact path %q does not start with expected lowercase prefix %q", art.Path, expectedPrefix)
				}

				// Parse Go file to verify syntax validity and inspect imports
				parsedFile, err := parser.ParseFile(fset, art.Path, art.Content, parser.ImportsOnly)
				if err != nil {
					t.Fatalf("failed to parse generated file %s: %v\nContent:\n%s", art.Path, err, string(art.Content))
				}

				for _, imp := range parsedFile.Imports {
					importPath := strings.Trim(imp.Path.Value, `"`)
					if strings.HasPrefix(importPath, modulePath+"/internal/") {
						// Ensure the import path uses tc.expectedPkg and does NOT use capitalized tc.inputName
						if strings.Contains(importPath, "/internal/"+tc.inputName+"/") && tc.inputName != tc.expectedPkg {
							t.Errorf("file %s contains capitalized import %s; want lowercase %s",
								art.Path, importPath, tc.expectedPkg)
						}
						if !strings.Contains(importPath, "/internal/"+tc.expectedPkg+"/") {
							t.Errorf("file %s expected import with lowercase package %s; got %s",
								art.Path, tc.expectedPkg, importPath)
						}
					}
				}
			}
		})
	}
}
