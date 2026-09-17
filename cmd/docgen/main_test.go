package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/cli"
	"github.com/spf13/cobra"
)

func TestGenerateDocs(t *testing.T) {
	tempDir := t.TempDir()

	rootCmd := &cobra.Command{
		Use:   "mytool",
		Short: "A root command",
	}
	subCmd := &cobra.Command{
		Use:   "sub",
		Short: "A subcommand",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	rootCmd.AddCommand(subCmd)

	err := GenerateDocsWithRoot(rootCmd, tempDir, tempDir)
	if err != nil {
		t.Fatalf("GenerateDocs error: %v", err)
	}

	rootFile := filepath.Join(tempDir, "mytool.md")
	subFile := filepath.Join(tempDir, "mytool_sub.md")

	for _, file := range []string{rootFile, subFile} {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("expected file %s to exist: %v", file, err)
		}
		str := string(content)

		if !strings.HasPrefix(str, "---\n") {
			t.Errorf("file %s missing frontmatter prefix", file)
		}
		if !strings.Contains(str, "slug: reference/cli/") {
			t.Errorf("file %s missing slug in frontmatter", file)
		}
		if !strings.Contains(str, "sidebar:") {
			t.Errorf("file %s missing sidebar config in frontmatter", file)
		}
	}
}

func TestGenerateDocs_PathTraversal(t *testing.T) {
	tempDir := t.TempDir()
	rootCmd := &cobra.Command{Use: "mytool"}

	err := GenerateDocsWithRoot(rootCmd, tempDir, "../../escaped")
	if err == nil {
		t.Fatal("expected path traversal error, got nil")
	}
}

func TestCleanCobraMarkdown(t *testing.T) {
	input := `---
title: "loy new"
slug: reference/cli/loy-new
---

## loy new

Create a new project
`
	expected := `---
title: "loy new"
slug: reference/cli/loy-new
---


Create a new project
`
	got := string(cleanCobraMarkdown([]byte(input)))
	if got != expected {
		t.Errorf("cleanCobraMarkdown mismatch:\nGot:\n%s\nWant:\n%s", got, expected)
	}
}

func TestDocumentationSyncGate(t *testing.T) {
	// Locate project root by finding go.mod
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	rootPath := pwd
	for {
		if _, err := os.Stat(filepath.Join(rootPath, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(rootPath)
		if parent == rootPath {
			t.Skip("skipping doc sync gate: could not locate project root containing go.mod")
			return
		}
		rootPath = parent
	}

	targetDir := filepath.Join(rootPath, "website/src/content/docs/reference/cli")
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		t.Skip("skipping doc sync gate: website reference directory does not exist")
		return
	}

	tempDir := t.TempDir()
	rootCmd := cli.NewRootCmd()

	if err := GenerateDocsWithRoot(rootCmd, tempDir, tempDir); err != nil {
		t.Fatalf("GenerateDocs failed: %v", err)
	}

	genEntries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("reading temp dir: %v", err)
	}

	commEntries, err := os.ReadDir(targetDir)
	if err != nil {
		t.Fatalf("reading target dir: %v", err)
	}

	genFiles := make(map[string]bool)
	for _, e := range genEntries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			genFiles[e.Name()] = true
		}
	}

	commFiles := make(map[string]bool)
	for _, e := range commEntries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			commFiles[e.Name()] = true
		}
	}

	// 1. Verify all generated files exist in committed website directory with identical content
	for fName := range genFiles {
		if !commFiles[fName] {
			t.Errorf("Documentation missing: %s exists in live CLI command tree but is missing from %s. Run 'make docgen' to synchronize.", fName, targetDir)
			continue
		}

		genBytes, err := os.ReadFile(filepath.Join(tempDir, fName))
		if err != nil {
			t.Fatalf("reading generated %s: %v", fName, err)
		}

		commBytes, err := os.ReadFile(filepath.Join(targetDir, fName))
		if err != nil {
			t.Fatalf("reading committed %s: %v", fName, err)
		}

		genStr := strings.TrimSpace(string(genBytes))
		commStr := strings.TrimSpace(string(commBytes))

		if genStr != commStr {
			t.Errorf("Documentation drift detected in %s. Live CLI command tree output does not match committed reference. Run 'make docgen' to synchronize.", fName)
		}
	}

	// 2. Verify no obsolete files exist
	for fName := range commFiles {
		if !genFiles[fName] {
			t.Errorf("Obsolete documentation file: %s exists in %s but does not correspond to any live Cobra command. Run 'make docgen' to clean up.", fName, targetDir)
		}
	}
}
