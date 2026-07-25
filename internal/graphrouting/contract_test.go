package graphrouting

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestImpactPacketValidFixture(t *testing.T) {
	packet, err := Load(filepath.Join("..", "..", "evals", "graph-routing", "impact-packet-valid.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	result := Validate(packet)
	if !result.OK {
		t.Fatalf("valid fixture failed: %v", result.Issues)
	}
	if result.Summary.EvidenceCounts["exact"] != 1 || result.Summary.EvidenceCounts["observed"] != 1 {
		t.Fatalf("unexpected evidence counts: %#v", result.Summary.EvidenceCounts)
	}
}

func TestImpactPacketRejectsStaleAuthoritativeProvider(t *testing.T) {
	packet, err := Load(filepath.Join("..", "..", "evals", "graph-routing", "impact-packet-stale.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	result := Validate(packet)
	if result.OK {
		t.Fatal("stale authoritative packet passed")
	}
	assertIssue(t, result.Issues, "stale packets must record an explicit non-authoritative fallback")
}

func TestImpactPacketRejectsMissingExactSpan(t *testing.T) {
	packet := validPacketForTest()
	packet.Edges[0].Span = SourceSpan{}
	result := Validate(packet)
	if result.OK {
		t.Fatal("exact edge without source span passed")
	}
	assertIssue(t, result.Issues, "load-bearing exact/observed edge requires source span")
}

func TestImpactPacketRejectsPathEscape(t *testing.T) {
	packet := validPacketForTest()
	packet.DirectImpact[0].Path = "../outside.go"
	result := Validate(packet)
	if result.OK {
		t.Fatal("path escape passed")
	}
	assertIssue(t, result.Issues, "impact path must be repository-relative")
}

func TestImpactPacketRejectsUnknownConfidence(t *testing.T) {
	packet := validPacketForTest()
	packet.Edges[0].Confidence = "certain"
	result := Validate(packet)
	if result.OK {
		t.Fatal("unknown confidence passed")
	}
	assertIssue(t, result.Issues, "confidence is invalid")
}

func TestGraphRouteSummaryIsCompact(t *testing.T) {
	summary := FormatSummary(Summarize(validPacketForTest()))
	if !strings.Contains(summary, "packet=fixture-valid") || !strings.Contains(summary, "exact=1") {
		t.Fatalf("summary missing signal: %s", summary)
	}
	if strings.Contains(summary, "cmd/kbcheck/main.go:") {
		t.Fatalf("summary dumped source detail: %s", summary)
	}
}

func validPacketForTest() Packet {
	return Packet{
		SchemaVersion: 1,
		PacketID:      "fixture-valid",
		Repository: Repository{
			Identity:            "git:file:///repo",
			Root:                ".",
			VCS:                 "git",
			Revision:            "abc123",
			DirtyFingerprint:    "clean",
			WorktreeFingerprint: "main:abc123",
			Freshness:           "fresh",
		},
		Seeds: Seeds{Files: []string{"cmd/kbcheck/main.go"}},
		Edges: []Edge{{
			From: "cmd/kbcheck.run", To: "internal/graphrouting.Validate", Type: "calls",
			Evidence: "exact", Confidence: "exact", Provenance: "file-native",
			Span:        SourceSpan{Path: "cmd/kbcheck/main.go", StartLine: 1, EndLine: 2},
			LoadBearing: true,
		}},
		DirectImpact:  []ImpactNode{{Path: "cmd/kbcheck/main.go", Reason: "registered command", Confidence: "high", Edges: []string{"cmd/kbcheck.run->internal/graphrouting.Validate"}}},
		ReverseImpact: []ImpactNode{{Path: ".github/skills/kb-map/SKILL.md", Reason: "consumes packet", Confidence: "medium"}},
		Tests:         []string{"internal/graphrouting/contract_test.go"},
		Docs:          []string{".github/skills/kb-map/references/graph-routing.md"},
		Fallback:      Fallback{Mode: "file-native", Reason: "no optional provider required"},
		Budget:        Budget{MaxEdges: 50, MaxBytes: 12000},
		Limitations:   []string{"fixture coverage only"},
		GeneratedBy:   "test",
		GeneratedAt:   "2026-07-19T00:00:00Z",
	}
}

func assertIssue(t *testing.T, issues []string, want string) {
	t.Helper()
	for _, issue := range issues {
		if strings.Contains(issue, want) {
			return
		}
	}
	t.Fatalf("missing issue containing %q in %v", want, issues)
}
