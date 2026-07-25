package graphrouting

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSCIPExactSymbolAdapterResolvesDefinitionReferenceImplementation(t *testing.T) {
	lookup, err := ResolveSCIPExactSymbols(fixtureIndexPath(), ProviderRequest{
		PacketID:   "scip-exact-symbol",
		Repository: fixtureRepository(),
		Seed:       Symbol{ID: "go github.com/example/payments Charge", Name: "Charge", Kind: "function"},
	})
	if err != nil {
		t.Fatalf("resolve exact symbols: %v", err)
	}
	if !lookup.Available || !lookup.Authoritative {
		t.Fatalf("expected authoritative lookup: %#v", lookup)
	}
	result := Validate(lookup.Packet)
	if !result.OK {
		t.Fatalf("packet did not validate: %v", result.Issues)
	}
	if got := result.Summary.EvidenceCounts["exact"]; got != 3 {
		t.Fatalf("expected three exact edges, got %d in %#v", got, lookup.Packet.Edges)
	}
	wantTypes := map[string]bool{"defines": false, "references": false, "implements": false}
	for _, edge := range lookup.Packet.Edges {
		wantTypes[edge.Type] = true
		if edge.Evidence != "exact" || edge.Confidence != "exact" || edge.Metadata.Provider != SCIPFixtureProvider {
			t.Fatalf("edge lost exact provenance: %#v", edge)
		}
	}
	for relationType, seen := range wantTypes {
		if !seen {
			t.Fatalf("missing %s relation in %#v", relationType, lookup.Packet.Edges)
		}
	}
}

func TestExactSymbolAdapterDistinguishesSameSpellingScopes(t *testing.T) {
	lookup, err := ResolveSCIPExactSymbols(fixtureIndexPath(), ProviderRequest{
		PacketID:   "scip-exact-symbol",
		Repository: fixtureRepository(),
		Seed:       Symbol{ID: "go github.com/example/refunds Charge", Name: "Charge", Kind: "function"},
	})
	if err != nil {
		t.Fatalf("resolve exact symbols: %v", err)
	}
	for _, edge := range lookup.Packet.Edges {
		joined := edge.From + " " + edge.To
		if strings.Contains(joined, "payments Charge") {
			t.Fatalf("same-spelling payment symbol leaked into refunds query: %#v", edge)
		}
	}
	if len(lookup.Packet.Edges) != 1 {
		t.Fatalf("expected only refund-scope relation, got %#v", lookup.Packet.Edges)
	}
}

func TestSCIPStaleRevisionFallsBack(t *testing.T) {
	repo := fixtureRepository()
	repo.Revision = "newer-revision"
	lookup, err := ResolveSCIPExactSymbols(fixtureIndexPath(), ProviderRequest{
		PacketID:   "scip-stale",
		Repository: repo,
		Seed:       Symbol{ID: "go github.com/example/payments Charge", Name: "Charge", Kind: "function"},
	})
	if err != nil {
		t.Fatalf("stale index should fall back, not fail: %v", err)
	}
	if lookup.Authoritative || lookup.Packet.Fallback.Mode != "file-native" {
		t.Fatalf("stale index claimed authority: %#v", lookup)
	}
	if !strings.Contains(lookup.Packet.Fallback.Reason, "stale") {
		t.Fatalf("fallback reason did not explain staleness: %q", lookup.Packet.Fallback.Reason)
	}
}

func TestSCIPRejectsOutOfRootSourceSpan(t *testing.T) {
	index := strings.NewReader(`{
	  "schema_version": 1,
	  "provider": "scip-fixture",
	  "repository": {
	    "identity": "git:file:///fixture",
	    "root": ".",
	    "revision": "abc123",
	    "worktree_fingerprint": "main:abc123"
	  },
	  "symbols": [
	    {"id": "seed", "name": "Charge", "kind": "function", "span": {"path": "src/payments.go", "start_line": 10, "end_line": 12}},
	    {"id": "escape", "name": "Charge", "kind": "reference", "span": {"path": "../outside.go", "start_line": 1, "end_line": 1}}
	  ],
	  "relations": [
	    {"from": "seed", "to": "escape", "type": "references", "span": {"path": "../outside.go", "start_line": 1, "end_line": 1}}
	  ]
	}`)
	parsed, err := DecodeSCIPExactSymbolIndex(index)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if _, err := resolveSCIPExactSymbolsFromIndex(parsed, "inline", ProviderRequest{
		PacketID:   "scip-invalid-span",
		Repository: fixtureRepository(),
		Seed:       Symbol{ID: "seed", Name: "Charge", Kind: "function"},
	}); err == nil {
		t.Fatal("out-of-root source span was accepted")
	}
}

func TestExactSymbolFallbackWhenMissingIndex(t *testing.T) {
	lookup, err := ResolveSCIPExactSymbols(filepath.Join(t.TempDir(), "missing.json"), ProviderRequest{
		PacketID:   "scip-missing",
		Repository: fixtureRepository(),
		Seed:       Symbol{ID: "go github.com/example/payments Charge", Name: "Charge", Kind: "function"},
	})
	if err != nil {
		t.Fatalf("missing index should fall back, not fail: %v", err)
	}
	if lookup.Available || lookup.Authoritative {
		t.Fatalf("missing index claimed provider authority: %#v", lookup)
	}
	if result := Validate(lookup.Packet); !result.OK {
		t.Fatalf("fallback packet should validate: %v", result.Issues)
	}
	if lookup.Packet.Fallback.Mode != "file-native" {
		t.Fatalf("missing index did not record file-native fallback: %#v", lookup.Packet.Fallback)
	}
}

func fixtureIndexPath() string {
	return filepath.Join("..", "..", "evals", "graph-routing", "exact-symbol-index.json")
}

func fixtureRepository() Repository {
	return Repository{
		Identity:            "git:file:///fixture",
		Root:                ".",
		VCS:                 "git",
		Revision:            "abc123",
		DirtyFingerprint:    "clean",
		WorktreeFingerprint: "main:abc123",
		Freshness:           "fresh",
	}
}
