package preset

import (
	"fmt"
	"strings"

	"github.com/loy-go/loy/internal/manifest"
)

// Preset represents a concrete configuration template for a new or initialized project.
type Preset struct {
	Name         string
	Description  string
	Defaults     manifest.DefaultsConfig
	MultiTenancy *manifest.MultiTenancyConfig
	Workspace    *manifest.WorkspaceConfig
}

// Manifest builds the typed manifest for a project.
func (p Preset) Manifest(projectName string) *manifest.Manifest {
	m := &manifest.Manifest{
		Version:  1,
		Project:  manifest.ProjectConfig{Name: projectName},
		Defaults: p.Defaults,
	}
	if p.MultiTenancy != nil {
		m.MultiTenancy = *p.MultiTenancy
	}
	if p.Workspace != nil {
		m.Workspace = *p.Workspace
	}
	return m
}

// MaterializeYAML produces clean, commented YAML for disk generation.
func (p Preset) MaterializeYAML(projectName string) string {
	var b strings.Builder
	b.WriteString("# Loy project configuration\n")
	b.WriteString("# See https://loy.dev/docs for configuration details.\n")
	b.WriteString("version: 1\n\n")
	fmt.Fprintf(&b, "project:\n  name: %s\n\n", projectName)

	if p.Workspace != nil {
		b.WriteString("workspace:\n")
		if p.Workspace.DefaultTarget != "" {
			fmt.Fprintf(&b, "  default_target: %s\n", p.Workspace.DefaultTarget)
		}
		if len(p.Workspace.Apps) > 0 {
			b.WriteString("  apps:\n")
			for _, app := range p.Workspace.Apps {
				fmt.Fprintf(&b, "    - %s\n", app)
			}
		}
		if len(p.Workspace.Packages) > 0 {
			b.WriteString("  packages:\n")
			for _, pkg := range p.Workspace.Packages {
				fmt.Fprintf(&b, "    - %s\n", pkg)
			}
		}
		return b.String()
	}

	b.WriteString("defaults:\n")
	if p.Defaults.HTTP != "" {
		fmt.Fprintf(&b, "  http: %s\n", p.Defaults.HTTP)
	}
	if p.Defaults.Database != "" {
		fmt.Fprintf(&b, "  database: %s\n", p.Defaults.Database)
	}
	if p.Defaults.Cache != "" {
		fmt.Fprintf(&b, "  cache: %s\n", p.Defaults.Cache)
	}
	if p.Defaults.Queue != "" {
		fmt.Fprintf(&b, "  queue: %s\n", p.Defaults.Queue)
	}
	if p.Defaults.Template != "" {
		fmt.Fprintf(&b, "  template: %s\n", p.Defaults.Template)
	}
	if p.Defaults.Assets != "" {
		fmt.Fprintf(&b, "  assets: %s\n", p.Defaults.Assets)
	}

	if p.MultiTenancy != nil && p.MultiTenancy.Enabled {
		b.WriteString("\nmulti_tenancy:\n")
		b.WriteString("  enabled: true\n")
		strategy := p.MultiTenancy.Strategy
		if strategy == "" {
			strategy = "rls"
		}
		fmt.Fprintf(&b, "  strategy: %s\n", strategy)
		if p.MultiTenancy.TenantKey != "" {
			fmt.Fprintf(&b, "  tenant_key: %s\n", p.MultiTenancy.TenantKey)
		}
	}

	return b.String()
}

// Registry manages available Loy presets.
type Registry struct {
	presets map[string]Preset
}

// NewRegistry initializes default built-in presets.
func NewRegistry() *Registry {
	r := &Registry{presets: make(map[string]Preset)}
	r.Register(Preset{
		Name:        "minimal",
		Description: "Minimal Go service with standard net/http",
		Defaults: manifest.DefaultsConfig{
			HTTP:     "nethttp",
			Database: "none",
			Cache:    "none",
			Queue:    "none",
		},
	})
	r.Register(Preset{
		Name:        "api",
		Description: "Standard REST API (Fiber, PostgreSQL/sqlc, Valkey, Asynq)",
		Defaults: manifest.DefaultsConfig{
			HTTP:     "fiber",
			Database: "postgres",
			Cache:    "valkey",
			Queue:    "asynq",
		},
	})
	r.Register(Preset{
		Name:        "saas",
		Description: "Multi-tenant SaaS platform (PostgreSQL RLS, Auth/RBAC, Valkey, Asynq)",
		Defaults: manifest.DefaultsConfig{
			HTTP:     "fiber",
			Database: "postgres",
			Cache:    "valkey",
			Queue:    "asynq",
		},
		MultiTenancy: &manifest.MultiTenancyConfig{
			Enabled:   true,
			Strategy:  "rls",
			TenantKey: "org_id",
		},
	})
	r.Register(Preset{
		Name:        "web",
		Description: "Web application with SSR templates (Fiber, Templ)",
		Defaults: manifest.DefaultsConfig{
			HTTP:     "fiber",
			Database: "postgres",
			Template: "templ",
		},
	})
	r.Register(Preset{
		Name:        "fullstack",
		Description: "Fullstack application (Fiber, Templ, Vite, PostgreSQL, Valkey)",
		Defaults: manifest.DefaultsConfig{
			HTTP:     "fiber",
			Database: "postgres",
			Cache:    "valkey",
			Queue:    "asynq",
			Template: "templ",
			Assets:   "vite",
		},
	})
	r.Register(Preset{
		Name:        "monorepo",
		Description: "Multi-application Go workspace with shared packages",
		Workspace: &manifest.WorkspaceConfig{
			DefaultTarget: "api",
			Apps:          []string{"apps/api"},
			Packages:      []string{"packages/*"},
		},
	})
	return r
}

// Register registers a concrete Preset in the registry.
func (r *Registry) Register(p Preset) {
	r.presets[strings.ToLower(p.Name)] = p
}

// Get retrieves a Preset by name.
func (r *Registry) Get(name string) (Preset, bool) {
	p, ok := r.presets[strings.ToLower(name)]
	return p, ok
}

// Names returns sorted names of available presets.
func (r *Registry) Names() []string {
	return []string{"api", "fullstack", "minimal", "monorepo", "saas", "web"}
}
