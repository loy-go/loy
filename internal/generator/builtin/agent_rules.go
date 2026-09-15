package builtin

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/loy-go/loy/internal/generator"
	"github.com/loy-go/loy/internal/generator/model"
)

// AgentRulesData holds parameters for agent rule generation.
type AgentRulesData struct {
	ModulePath  string
	ProjectName string
}

// AgentRulesGenerator scaffolds authoritative AI agent instruction documents.
type AgentRulesGenerator struct {
	modulePath string
	target     string
}

// NewAgentRulesGenerator constructs an AgentRulesGenerator.
func NewAgentRulesGenerator(modulePath string) *AgentRulesGenerator {
	return &AgentRulesGenerator{
		modulePath: modulePath,
		target:     "all",
	}
}

// WithTarget sets the target assistant ("all", "cursor", "claude", "copilot", "windsurf").
func (g *AgentRulesGenerator) WithTarget(target string) *AgentRulesGenerator {
	if target != "" {
		g.target = strings.ToLower(target)
	}
	return g
}

func (g *AgentRulesGenerator) Name() string {
	return "agent-rules"
}

func (g *AgentRulesGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
	projectName := filepath.Base(g.modulePath)
	if projectName == "." || projectName == "/" || projectName == "" {
		projectName = "app"
	}
	if input.Name != "" && input.Name != "agent-rules" {
		projectName = input.Name
	}

	target := g.target
	if input.Args != nil {
		if t, ok := input.Args["target"]; ok && t != "" {
			target = strings.ToLower(t)
		} else if f, ok := input.Args["for"]; ok && f != "" {
			target = strings.ToLower(f)
		} else if fields, ok := input.Args["fields"]; ok && fields != "" {
			target = strings.ToLower(strings.TrimSpace(fields))
		}
	}

	data := AgentRulesData{
		ModulePath:  g.modulePath,
		ProjectName: projectName,
	}

	renderer := GetRenderer()

	agentsMdTmpl, err := ReadTemplate("agent_rules_agents_md.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading agent_rules_agents_md.tmpl: %w", err)
	}
	agentsMdBytes, err := renderer.Render(ctx, "agent_rules_agents_md.tmpl", agentsMdTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering AGENTS.md: %w", err)
	}

	cursorMdcTmpl, err := ReadTemplate("agent_rules_cursor_mdc.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading agent_rules_cursor_mdc.tmpl: %w", err)
	}
	cursorMdcBytes, err := renderer.Render(ctx, "agent_rules_cursor_mdc.tmpl", cursorMdcTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering cursor rules: %w", err)
	}

	copilotTmpl, err := ReadTemplate("agent_rules_copilot.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading agent_rules_copilot.tmpl: %w", err)
	}
	copilotBytes, err := renderer.Render(ctx, "agent_rules_copilot.tmpl", copilotTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering copilot instructions: %w", err)
	}

	windsurfTmpl, err := ReadTemplate("agent_rules_windsurf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading agent_rules_windsurf.tmpl: %w", err)
	}
	windsurfBytes, err := renderer.Render(ctx, "agent_rules_windsurf.tmpl", windsurfTmpl, data)
	if err != nil {
		return nil, fmt.Errorf("rendering windsurf rules: %w", err)
	}

	var artifacts []model.Artifact

	addClaude := func() {
		artifacts = append(artifacts,
			model.Artifact{
				Path:        "AGENTS.md",
				Ownership:   model.DeveloperOwned,
				Permissions: 0644,
				Content:     agentsMdBytes,
			},
			model.Artifact{
				Path:        "CLAUDE.md",
				Ownership:   model.DeveloperOwned,
				Permissions: 0644,
				Content:     agentsMdBytes,
			},
		)
	}

	addCursor := func() {
		artifacts = append(artifacts,
			model.Artifact{
				Path:        ".cursor/rules/loy.mdc",
				Ownership:   model.DeveloperOwned,
				Permissions: 0644,
				Content:     cursorMdcBytes,
			},
			model.Artifact{
				Path:        ".cursorrules",
				Ownership:   model.DeveloperOwned,
				Permissions: 0644,
				Content:     cursorMdcBytes,
			},
		)
	}

	addCopilot := func() {
		artifacts = append(artifacts,
			model.Artifact{
				Path:        ".github/copilot-instructions.md",
				Ownership:   model.DeveloperOwned,
				Permissions: 0644,
				Content:     copilotBytes,
			},
		)
	}

	addWindsurf := func() {
		artifacts = append(artifacts,
			model.Artifact{
				Path:        ".windsurfrules",
				Ownership:   model.DeveloperOwned,
				Permissions: 0644,
				Content:     windsurfBytes,
			},
		)
	}

	switch target {
	case "claude":
		addClaude()
	case "cursor":
		addCursor()
	case "copilot":
		addCopilot()
	case "windsurf":
		addWindsurf()
	case "all", "":
		addClaude()
		addCursor()
		addCopilot()
		addWindsurf()
	default:
		return nil, fmt.Errorf("unknown target %q: expected all, cursor, claude, copilot, or windsurf", target)
	}

	return artifacts, nil
}
