package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Irtechie/working-skill-repo/internal/graphrouting"
)

func TestGraphRouteCommandValidatesImpactPacket(t *testing.T) {
	t.Parallel()
	var out, errOut strings.Builder
	code := run([]string{"graph-route", "--root", filepath.Join("..", ".."), "--packet", filepath.Join("evals", "graph-routing", "impact-packet-valid.json")}, &out, &errOut)
	if code != 0 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), "graph-route: ok") || !strings.Contains(out.String(), "fallback=file-native") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestGraphRouteCommandRejectsStalePacket(t *testing.T) {
	t.Parallel()
	var out, errOut strings.Builder
	code := run([]string{"graph-route", "--root", filepath.Join("..", ".."), "--packet", filepath.Join("evals", "graph-routing", "impact-packet-stale.json")}, &out, &errOut)
	if code != 2 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	if !strings.Contains(errOut.String(), "stale packets") {
		t.Fatalf("missing stale issue: %s", errOut.String())
	}
}

func TestGraphRouteParseRequiresPacket(t *testing.T) {
	t.Parallel()
	if _, err := parse([]string{"graph-route"}); err == nil {
		t.Fatal("graph-route without --packet passed")
	}
}

func TestGraphifyAnnotationRejectsExactStructuralProvider(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	packetPath := filepath.Join(tempDir, "packet.json")
	packet := graphrouting.Packet{
		SchemaVersion: 1,
		PacketID:      "graphify-invalid",
		Repository: graphrouting.Repository{
			Identity:            "git:file:///fixture",
			Root:                ".",
			VCS:                 "git",
			Revision:            "abc123",
			DirtyFingerprint:    "clean",
			WorktreeFingerprint: "main:abc123",
			Freshness:           "fresh",
		},
		Seeds: graphrouting.Seeds{Files: []string{"src/api/payments.go"}},
		Edges: []graphrouting.Edge{{
			From:       "api.createPayment",
			To:         "payments.Charge",
			Type:       "CALLS_STATIC",
			Evidence:   "exact",
			Confidence: "exact",
			Provenance: "graphify-output:inline",
			Span:       graphrouting.SourceSpan{Path: "src/api/payments.go", StartLine: 44, EndLine: 45},
			Metadata:   graphrouting.EdgeMetadata{Provider: graphrouting.GraphifyProvider},
		}},
		Fallback:    graphrouting.Fallback{Mode: "file-native", Reason: "verify structural provider edges against source"},
		Budget:      graphrouting.Budget{MaxEdges: 10, MaxBytes: 1000},
		Limitations: []string{"fixture"},
		GeneratedBy: "test",
	}
	encoded, err := json.Marshal(packet)
	if err != nil {
		t.Fatalf("marshal packet: %v", err)
	}
	if err := os.WriteFile(packetPath, encoded, 0o600); err != nil {
		t.Fatalf("write packet: %v", err)
	}

	var out, errOut strings.Builder
	code := run([]string{"graph-route", "--packet", packetPath}, &out, &errOut)
	if code != 2 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	if !strings.Contains(errOut.String(), "graphify structural provider cannot claim exact evidence") {
		t.Fatalf("missing graphify annotation issue: %s", errOut.String())
	}
}
