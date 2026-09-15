package diagnostics

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// AgentRemediation models structured corrective instructions for an autonomous AI agent.
type AgentRemediation struct {
	Action string `json:"action"`
	Prompt string `json:"prompt"`
}

// AgentDiagnostic represents an actionable architectural violation tailored for AI agents.
type AgentDiagnostic struct {
	Code        string           `json:"code"`
	Severity    string           `json:"severity"`
	File        string           `json:"file"`
	Line        int              `json:"line"`
	Violation   string           `json:"violation"`
	Rationale   string           `json:"rationale"`
	Remediation AgentRemediation `json:"remediation"`
}

// AgentPromptFormatter formats diagnostics as an agent self-healing JSON array.
type AgentPromptFormatter struct{}

// Format renders the diagnostics into formatted JSON for AI agents.
func (f *AgentPromptFormatter) Format(w io.Writer, diags []Diagnostic) error {
	agentDiags := make([]AgentDiagnostic, 0, len(diags))
	for _, d := range diags {
		agentDiags = append(agentDiags, ToAgentDiagnostic(d))
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(agentDiags)
}

// ToAgentDiagnostic converts a standard diagnostic into a rich AgentDiagnostic.
func ToAgentDiagnostic(d Diagnostic) AgentDiagnostic {
	ruleID := extractRuleID(d.Code)
	severity := string(d.Severity)
	if severity == "" {
		severity = "error"
	}

	violation := d.Message
	if d.Detail != "" && !strings.Contains(violation, d.Detail) {
		violation = fmt.Sprintf("%s (%s)", d.Message, d.Detail)
	}

	rationale, action, prompt := resolveRuleRemediation(ruleID, d)

	return AgentDiagnostic{
		Code:      d.Code,
		Severity:  severity,
		File:      d.File,
		Line:      d.Line,
		Violation: violation,
		Rationale: rationale,
		Remediation: AgentRemediation{
			Action: action,
			Prompt: prompt,
		},
	}
}

func extractRuleID(code string) string {
	if strings.HasPrefix(code, "LOY-ARCH-") {
		return strings.TrimPrefix(code, "LOY-")
	}
	if strings.HasPrefix(code, "ARCH-") {
		return code
	}
	return ""
}

func resolveRuleRemediation(ruleID string, d Diagnostic) (rationale, action, prompt string) {
	file := d.File
	if file == "" {
		file = "target file"
	}
	entityName := strings.TrimSuffix(filepath.Base(file), ".go")
	if entityName == "" || entityName == "." || entityName == "/" {
		entityName = "Entity"
	}
	entityPascal := capitalize(entityName)

	switch ruleID {
	case "ARCH-001":
		rationale = "Package dependency cycles prevent compilation and violate the Directed Acyclic Graph (DAG) package architecture."
		action = "break_cycle"
		prompt = fmt.Sprintf("Break the cyclic dependency involving %s. Extract shared types and interfaces into a lower-level leaf package or invert dependencies using interfaces.", d.File)

	case "ARCH-002":
		rationale = "Domain entities must remain pure Go structs independent of persistence drivers."
		action = "invert_dependency"
		imported := extractImportFromMsg(d.Message)
		if imported != "" {
			prompt = fmt.Sprintf("Remove import of %q from %s. Define a repository interface in the domain package: \n\ntype %sRepository interface {\n    FindByID(ctx context.Context, id int64) (*%s, error)\n}\n\nThen implement this interface in internal/.../repository/.", imported, file, entityPascal, entityPascal)
		} else {
			prompt = fmt.Sprintf("Remove infrastructure imports from %s. Define a repository interface in the domain package and implement it in the infrastructure layer.", file)
		}

	case "ARCH-003":
		rationale = "Domain models must remain completely decoupled from HTTP, gRPC, WebSocket or other transport protocols."
		action = "remove_import"
		prompt = fmt.Sprintf("Remove transport imports from %s. HTTP/gRPC handlers should receive transport DTOs, convert them to domain entities, and pass them into application services.", file)

	case "ARCH-004":
		rationale = "Application layer orchestrates business logic and use cases; it must never depend on incoming transport protocols."
		action = "invert_call_direction"
		prompt = fmt.Sprintf("Remove transport import from %s. Transport handlers should invoke application services, never vice versa.", file)

	case "ARCH-005":
		rationale = "Application use cases should depend only on abstract domain interfaces, never on concrete SQL or adapter implementations."
		action = "invert_dependency"
		prompt = fmt.Sprintf("Remove concrete infrastructure adapter imports from %s. Declare consumer interfaces in the application layer and inject concrete implementations in internal/app/wiring.go.", file)

	case "ARCH-006":
		rationale = "Infrastructure adapters (persistence, queues, cache) must not depend on web or transport layers."
		action = "remove_import"
		prompt = fmt.Sprintf("Remove transport dependencies from %s. Infrastructure adapters should operate on domain models or storage entities without web types.", file)

	case "ARCH-007":
		rationale = "Transport handlers must only parse requests, call application services, and serialize responses. Raw database queries in handlers violate separation of concerns."
		action = "delegate_to_service"
		prompt = fmt.Sprintf("Remove raw database/SQL imports from transport file %s. Move persistence logic into an application service and repository adapter.", file)

	case "ARCH-008":
		rationale = "Domain layer must remain pure Go standard library without web frameworks, ORMs, or cache SDK dependencies."
		action = "isolate_domain"
		prompt = fmt.Sprintf("Remove external framework or SDK imports from domain file %s. Move framework-specific code to transport or infrastructure.", file)

	case "ARCH-009":
		rationale = "Application services must be transport-agnostic and free of HTTP framework types."
		action = "decouple_transport"
		prompt = fmt.Sprintf("Remove web framework imports from application service %s. Use standard Go structs, primitives, and context.Context for method signatures.", file)

	case "ARCH-010":
		rationale = "Loy mandates strict Clean Architecture layer flow: Transport -> Application -> Domain <- Infrastructure."
		action = "realign_layer"
		prompt = fmt.Sprintf("Realign package imports in %s to adhere to the 4-layer dependency matrix. Avoid backwards layer imports.", file)

	case "ARCH-011":
		rationale = "Dynamic DI reflection containers (dig, wire, di) obscure call graphs and fail at runtime. Loy mandates explicit constructor injection."
		action = "explicit_wiring"
		prompt = fmt.Sprintf("Remove DI container imports from %s. Wire components explicitly using standard Go constructors (New...(...)) in internal/app/wiring.go per ADR-003.", file)

	case "ARCH-012":
		rationale = "Package-level mutable variables create race conditions, prevent concurrent testing, and violate zero-global-state invariants."
		action = "encapsulate_state"
		prompt = fmt.Sprintf("Remove package-level mutable variable from %s. Pass state explicitly as struct fields initialized via constructor injection or via context.Context.", file)

	case "ARCH-013":
		rationale = "Monorepo applications must remain independently deployable and decoupled from peer applications."
		action = "extract_shared_package"
		prompt = fmt.Sprintf("Remove cross-app import from %s. Move shared models or domain logic into a shared package under internal/ or packages/.", file)

	case "ARCH-014":
		rationale = "Loy preserves developer edits by isolating generated code inside // loy:region:<name> and // loy:endregion markers. Corrupting or omitting markers breaks atomic codegen."
		action = "restore_comment_region"
		prompt = fmt.Sprintf("Restore missing or corrupted '// loy:region:...' and '// loy:endregion' comment markers in %s, or run 'loy upgrade' to repair scaffolding.", file)

	case "ARCH-015":
		rationale = "Platform packages must remain pure standard utilities (math, hashing, strings, errors). Direct IO, network, or database operations belong in infrastructure."
		action = "relocate_impure_platform"
		prompt = fmt.Sprintf("Move impure platform code performing network, OS, or DB operations in %s into infrastructure, or define an abstraction interface in domain.", file)

	default:
		rationale = "Architectural invariant or code quality rule failure."
		action = "fix_violation"
		if d.Hint != "" {
			prompt = fmt.Sprintf("%s. Hint: %s", d.Message, d.Hint)
		} else {
			prompt = d.Message
		}
	}

	return rationale, action, prompt
}

func extractImportFromMsg(msg string) string {
	// Look for package in message: "... imports infrastructure package <pkg>"
	parts := strings.Split(msg, "imports infrastructure package ")
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	parts = strings.Split(msg, "imports transport package ")
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	parts = strings.Split(msg, "imports ")
	if len(parts) == 2 {
		f := strings.Fields(parts[1])
		if len(f) > 0 {
			return f[len(f)-1]
		}
	}
	return ""
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
