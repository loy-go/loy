package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() {
	outDir := flag.String("out", "website/src/content/docs/reference/cli", "Directory to output markdown files")
	flag.Parse()

	rootCmd := cli.NewRootCmd()

	if err := GenerateDocs(rootCmd, *outDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating documentation: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated CLI reference in %s\n", *outDir)
}

// GenerateDocs extracts Markdown documentation from a Cobra command tree for Starlight.
func GenerateDocs(rootCmd *cobra.Command, outDir string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}
	return GenerateDocsWithRoot(rootCmd, pwd, outDir)
}

// GenerateDocsWithRoot extracts Markdown documentation, ensuring outDir remains confined within rootDir.
func GenerateDocsWithRoot(rootCmd *cobra.Command, rootDir, outDir string) error {
	cleanOut, err := filesystem.CleanAndValidatePath(rootDir, outDir)
	if err != nil {
		return fmt.Errorf("validating output directory path %q: %w", outDir, err)
	}

	if err := os.MkdirAll(cleanOut, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Disable auto-generated timestamps across the entire tree
	setDisableAutoGen(rootCmd)

	// Index commands by their Cobra markdown filename
	cmdMap := make(map[string]*cobra.Command)
	orderMap := make(map[string]int)
	indexCommands(rootCmd, cmdMap, orderMap, 0)

	filePrepender := func(filename string) string {
		base := filepath.Base(filename)
		cmd, ok := cmdMap[base]
		if !ok {
			return ""
		}

		cmdPath := cmd.CommandPath()
		slug := toKebabSlug(cmdPath)
		order := orderMap[base]
		shortDesc := strings.ReplaceAll(cmd.Short, `"`, `\"`)

		return fmt.Sprintf("---\ntitle: \"%s\"\ndescription: \"%s\"\nslug: reference/cli/%s\nsidebar:\n  order: %d\n---\n\n",
			cmdPath, shortDesc, slug, order)
	}

	basePath := os.Getenv("DOCS_BASE")
	if basePath == "" {
		basePath = "/loy"
	}
	basePath = strings.TrimSuffix(basePath, "/")

	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, ".md")
		kebab := strings.ReplaceAll(base, "_", "-")
		if basePath == "" {
			return fmt.Sprintf("/reference/cli/%s/", kebab)
		}
		return fmt.Sprintf("%s/reference/cli/%s/", basePath, kebab)
	}

	if err := doc.GenMarkdownTreeCustom(rootCmd, cleanOut, filePrepender, linkHandler); err != nil {
		return fmt.Errorf("generating markdown tree: %w", err)
	}

	// Post-process files to remove redundant leading markdown headers
	return cleanGeneratedFiles(cleanOut)
}

func setDisableAutoGen(cmd *cobra.Command) {
	cmd.DisableAutoGenTag = true
	for _, c := range cmd.Commands() {
		setDisableAutoGen(c)
	}
}

func indexCommands(cmd *cobra.Command, cmdMap map[string]*cobra.Command, orderMap map[string]int, order int) int {
	filename := cmdFilename(cmd)
	cmdMap[filename] = cmd
	orderMap[filename] = order
	order++

	for _, sub := range cmd.Commands() {
		if !sub.IsAvailableCommand() || sub.IsAdditionalHelpTopicCommand() {
			continue
		}
		order = indexCommands(sub, cmdMap, orderMap, order)
	}
	return order
}

func cmdFilename(cmd *cobra.Command) string {
	name := cmd.CommandPath()
	return strings.ReplaceAll(name, " ", "_") + ".md"
}

func toKebabSlug(cmdPath string) string {
	return strings.ReplaceAll(strings.ToLower(cmdPath), " ", "-")
}

func cleanGeneratedFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		cleaned := cleanCobraMarkdown(data)
		if !bytes.Equal(data, cleaned) {
			if err := os.WriteFile(path, cleaned, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

// cleanCobraMarkdown removes the redundant Cobra leading `# <cmd>` or `## <cmd>` heading
// that appears right after the YAML frontmatter.
func cleanCobraMarkdown(content []byte) []byte {
	str := string(content)
	if !strings.HasPrefix(str, "---\n") {
		return content
	}

	endIdx := strings.Index(str[4:], "\n---\n")
	if endIdx == -1 {
		return content
	}

	fmEnd := 4 + endIdx + 5
	frontmatter := str[:fmEnd]
	body := str[fmEnd:]

	lines := strings.Split(body, "\n")
	var newLines []string
	headerRemoved := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !headerRemoved && (strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "# ")) {
			headerRemoved = true
			continue
		}
		newLines = append(newLines, line)
	}

	return []byte(frontmatter + strings.Join(newLines, "\n"))
}
