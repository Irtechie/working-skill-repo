package graphrouting

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGraphifyNormalizesTypedEdgesAndInferredUncertainty(t *testing.T) {
	lookup, err := resolveGraphifyStructuralEdgesFromGraph(graphifyFixture(false, "abc123"), "inline", graphifyRecipe(), ProviderRequest{
		PacketID:   "graphify-structural",
		Repository: fixtureRepository(),
		Seed:       Symbol{ID: "api.createPayment", Name: "createPayment", Kind: "function"},
	})
	if err != nil {
		t.Fatalf("resolve graphify: %v", err)
	}
	if lookup.Authoritative {
		t.Fatal("structural Graphify provider claimed authority")
	}
	result := Validate(lookup.Packet)
	if !result.OK {
		t.Fatalf("packet failed validation: %v", result.Issues)
	}
	if issues := ValidateTraversalAnnotations(lookup.Packet); len(issues) > 0 {
		t.Fatalf("valid graphify annotations failed: %v", issues)
	}
	seen := map[string]Edge{}
	for _, edge := range lookup.Packet.Edges {
		seen[edge.Type] = edge
	}
	if seen["CALLS_STATIC"].Evidence != "structural" || seen["READS_CONFIG"].Evidence != "structural" {
		t.Fatalf("typed structural edges not preserved: %#v", seen)
	}
	if inferred := seen["INFERRED_CALLS_STATIC"]; inferred.Evidence != "heuristic" || inferred.Confidence == "exact" {
		t.Fatalf("inferred edge claimed too much authority: %#v", inferred)
	}
}

func TestGraphifyStaleRevisionFallsBack(t *testing.T) {
	lookup, err := resolveGraphifyStructuralEdgesFromGraph(graphifyFixture(false, "old"), "inline", graphifyRecipe(), ProviderRequest{
		PacketID:   "graphify-stale",
		Repository: fixtureRepository(),
		Seed:       Symbol{ID: "api.createPayment", Name: "createPayment", Kind: "function"},
	})
	if err != nil {
		t.Fatalf("stale graph should fall back: %v", err)
	}
	if lookup.Available || lookup.Packet.Fallback.Mode != "file-native" {
		t.Fatalf("stale graphify output claimed availability: %#v", lookup)
	}
	if !strings.Contains(lookup.Packet.Fallback.Reason, "stale") {
		t.Fatalf("missing stale fallback reason: %q", lookup.Packet.Fallback.Reason)
	}
}

func TestGraphifyMissingProviderFallsBack(t *testing.T) {
	lookup, err := ResolveGraphifyStructuralEdges(filepath.Join(t.TempDir(), "graph.json"), graphifyRecipe(), ProviderRequest{
		PacketID:   "graphify-missing",
		Repository: fixtureRepository(),
		Seed:       Symbol{ID: "api.createPayment", Name: "createPayment", Kind: "function"},
	})
	if err != nil {
		t.Fatalf("missing graphify output should fall back: %v", err)
	}
	if lookup.Available || lookup.Authoritative {
		t.Fatalf("missing graphify output claimed authority: %#v", lookup)
	}
	if result := Validate(lookup.Packet); !result.OK {
		t.Fatalf("fallback packet failed validation: %v", result.Issues)
	}
}

func TestGraphifyDiagnosesMultigraphCollapse(t *testing.T) {
	lookup, err := resolveGraphifyStructuralEdgesFromGraph(graphifyFixture(true, "abc123"), "inline", graphifyRecipe(), ProviderRequest{
		PacketID:   "graphify-collapse",
		Repository: fixtureRepository(),
		Seed:       Symbol{ID: "api.createPayment", Name: "createPayment", Kind: "function"},
	})
	if err != nil {
		t.Fatalf("resolve graphify collapse fixture: %v", err)
	}
	for _, edge := range lookup.Packet.Edges {
		for _, limitation := range edge.Limitations {
			if strings.Contains(limitation, "multigraph collapse") {
				return
			}
		}
	}
	t.Fatalf("multigraph collapse risk not preserved: %#v", lookup.Packet.Edges)
}

func graphifyRecipe() TraversalRecipe {
	return TraversalRecipe{
		ID:          "api-change",
		Intent:      "api-change",
		MaxDepth:    3,
		MaxEdges:    10,
		MaxBytes:    4000,
		Limitations: []string{"static graph cannot prove dynamic dispatch"},
	}
}

func graphifyFixture(collapsed bool, revision string) GraphifyGraph {
	return GraphifyGraph{
		SchemaVersion: 1,
		Provider:      GraphifyProvider,
		Repository: GraphifyRepository{
			Identity:            "git:file:///fixture",
			Revision:            revision,
			WorktreeFingerprint: "main:abc123",
		},
		MultigraphCollapsed: collapsed,
		Nodes: []GraphifyNode{
			{ID: "api.createPayment", Path: "src/api/payments.go", Symbol: "createPayment", Span: SourceSpan{Path: "src/api/payments.go", StartLine: 44, EndLine: 45}},
			{ID: "payments.Charge", Path: "src/payments/charge.go", Symbol: "Charge", Span: SourceSpan{Path: "src/payments/charge.go", StartLine: 12, EndLine: 27}},
			{ID: "config.paymentProvider", Path: "config/payments.yaml", Symbol: "paymentProvider", Span: SourceSpan{Path: "config/payments.yaml", StartLine: 3, EndLine: 3}},
			{ID: "runtime.dynamicHook", Path: "src/runtime/hooks.go", Symbol: "dynamicHook", Span: SourceSpan{Path: "src/runtime/hooks.go", StartLine: 7, EndLine: 8}},
		},
		Edges: []GraphifyGraphEdge{
			{From: "api.createPayment", To: "payments.Charge", Type: "calls", Span: SourceSpan{Path: "src/api/payments.go", StartLine: 44, EndLine: 45}},
			{From: "api.createPayment", To: "config.paymentProvider", Type: "reads_config", Span: SourceSpan{Path: "config/payments.yaml", StartLine: 3, EndLine: 3}},
			{From: "api.createPayment", To: "runtime.dynamicHook", Type: "calls", Inferred: true, Span: SourceSpan{Path: "src/runtime/hooks.go", StartLine: 7, EndLine: 8}},
		},
	}
}
