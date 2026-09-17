package rules

import (
	"context"
	"fmt"

	"github.com/loy-go/loy/internal/architecture"
	"github.com/loy-go/loy/internal/diagnostics"
)

// RuleArch002 flags domain importing infrastructure.
type RuleArch002 struct{}

func (r *RuleArch002) ID() string          { return "ARCH-002" }
func (r *RuleArch002) Description() string { return "Domain must not import infrastructure" }

func (r *RuleArch002) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	if a.Topology != nil && a.Topology.IsCustom() {
		return nil
	}
	var violations []architecture.Violation

	for _, edge := range a.Graph.AllEdges() {
		fromNode := a.Graph.Node(edge.From)
		toNode := a.Graph.Node(edge.To)
		if fromNode == nil || toNode == nil {
			continue
		}

		if fromNode.Layer == string(architecture.LayerDomain) && toNode.Layer == string(architecture.LayerInfrastructure) {
			violations = append(violations, architecture.Violation{
				RuleID:       r.ID(),
				Code:         diagnostics.CodeArchDomainInfra,
				Message:      fmt.Sprintf("domain package %s imports infrastructure package %s", edge.From, edge.To),
				Detail:       "domain must remain clean of infrastructure implementations per ADR-001 and Spec 05",
				Hint:         "define repository interface in domain or application layer and implement in infrastructure",
				File:         edge.File,
				Line:         edge.Line,
				Suppressible: true,
			})
		}
	}

	return violations
}

// RuleArch003 flags domain importing transport.
type RuleArch003 struct{}

func (r *RuleArch003) ID() string          { return "ARCH-003" }
func (r *RuleArch003) Description() string { return "Domain must not import transport" }

func (r *RuleArch003) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	if a.Topology != nil && a.Topology.IsCustom() {
		return nil
	}
	var violations []architecture.Violation

	for _, edge := range a.Graph.AllEdges() {
		fromNode := a.Graph.Node(edge.From)
		toNode := a.Graph.Node(edge.To)
		if fromNode == nil || toNode == nil {
			continue
		}

		if fromNode.Layer == string(architecture.LayerDomain) && toNode.Layer == string(architecture.LayerTransport) {
			violations = append(violations, architecture.Violation{
				RuleID:       r.ID(),
				Code:         diagnostics.CodeArchDomainTransport,
				Message:      fmt.Sprintf("domain package %s imports transport package %s", edge.From, edge.To),
				Detail:       "domain must be independent of HTTP, gRPC or transport protocols",
				Hint:         "pass plain Go domain entities from transport into application services",
				File:         edge.File,
				Line:         edge.Line,
				Suppressible: true,
			})
		}
	}

	return violations
}

// RuleArch004 flags application importing transport.
type RuleArch004 struct{}

func (r *RuleArch004) ID() string          { return "ARCH-004" }
func (r *RuleArch004) Description() string { return "Application must not import transport" }

func (r *RuleArch004) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	if a.Topology != nil && a.Topology.IsCustom() {
		return nil
	}
	var violations []architecture.Violation

	for _, edge := range a.Graph.AllEdges() {
		fromNode := a.Graph.Node(edge.From)
		toNode := a.Graph.Node(edge.To)
		if fromNode == nil || toNode == nil {
			continue
		}

		if fromNode.Layer == string(architecture.LayerApplication) && toNode.Layer == string(architecture.LayerTransport) {
			violations = append(violations, architecture.Violation{
				RuleID:       r.ID(),
				Code:         diagnostics.CodeArchAppTransport,
				Message:      fmt.Sprintf("application package %s imports transport package %s", edge.From, edge.To),
				Detail:       "application layer orchestrates business logic and must not depend on transport",
				Hint:         "transport should call application service methods, not vice versa",
				File:         edge.File,
				Line:         edge.Line,
				Suppressible: true,
			})
		}
	}

	return violations
}

// RuleArch005 flags application importing concrete infrastructure.
type RuleArch005 struct{}

func (r *RuleArch005) ID() string          { return "ARCH-005" }
func (r *RuleArch005) Description() string { return "Application must not import concrete infrastructure" }

func (r *RuleArch005) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	if a.Topology != nil && a.Topology.IsCustom() {
		return nil
	}
	var violations []architecture.Violation

	for _, edge := range a.Graph.AllEdges() {
		fromNode := a.Graph.Node(edge.From)
		toNode := a.Graph.Node(edge.To)
		if fromNode == nil || toNode == nil {
			continue
		}

		if fromNode.Layer == string(architecture.LayerApplication) && toNode.Layer == string(architecture.LayerInfrastructure) {
			violations = append(violations, architecture.Violation{
				RuleID:       r.ID(),
				Code:         diagnostics.CodeArchAppConcreteInfra,
				Message:      fmt.Sprintf("application package %s directly imports infrastructure package %s", edge.From, edge.To),
				Detail:       "application should consume interfaces; concrete infrastructure should be wired via main/wiring",
				Hint:         "declare consumer interfaces in application and inject infrastructure in internal/app/wiring.go",
				File:         edge.File,
				Line:         edge.Line,
				Suppressible: true,
			})
		}
	}

	return violations
}

// RuleArch006 flags infrastructure importing transport.
type RuleArch006 struct{}

func (r *RuleArch006) ID() string          { return "ARCH-006" }
func (r *RuleArch006) Description() string { return "Infrastructure must not import transport" }

func (r *RuleArch006) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
	if a.Topology != nil && a.Topology.IsCustom() {
		return nil
	}
	var violations []architecture.Violation

	for _, edge := range a.Graph.AllEdges() {
		fromNode := a.Graph.Node(edge.From)
		toNode := a.Graph.Node(edge.To)
		if fromNode == nil || toNode == nil {
			continue
		}

		if fromNode.Layer == string(architecture.LayerInfrastructure) && toNode.Layer == string(architecture.LayerTransport) {
			violations = append(violations, architecture.Violation{
				RuleID:       r.ID(),
				Code:         diagnostics.CodeArchInfraTransport,
				Message:      fmt.Sprintf("infrastructure package %s imports transport package %s", edge.From, edge.To),
				Detail:       "infrastructure adapters (persistence, queues, cache) must not depend on transport",
				Hint:         "remove transport dependency from infrastructure adapter",
				File:         edge.File,
				Line:         edge.Line,
				Suppressible: true,
			})
		}
	}

	return violations
}
