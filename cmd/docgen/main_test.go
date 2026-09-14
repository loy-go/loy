package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
