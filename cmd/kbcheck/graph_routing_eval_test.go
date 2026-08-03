package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGraphRoutingEvalRequireReadyPasses(t *testing.T) {
	t.Parallel()
	var out, errOut strings.Builder
	code := run([]string{"graph-routing-eval", "--root", filepath.Join("..", ".."), "--require-ready"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), "graph-routing-eval: ok") || !strings.Contains(out.String(), "ready=true") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestGraphRoutingEvalFailsMissedImpact(t *testing.T) {
	t.Parallel()
	fixture := symbolImpactFixture{
		SchemaVersion:   1,
		ExpectedImpact:  []string{"src/api/payments.go", "src/payments/charge.go"},
		Retrieved:       []string{"src/api/payments.go"},
		RequiredTests:   []string{"src/api/payments_test.go"},
		RetrievedTokens: 100,
	}
	_, _, _, _, failures := scoreSymbolImpact(fixture)
	if !containsFailure(failures, "missed impacted path") || !containsFailure(failures, "missed required test/doc") {
		t.Fatalf("missed impact was not detected: %v", failures)
	}
}

func TestGraphRoutingEvalFailsStaleAuthoritativeIndex(t *testing.T) {
	t.Parallel()
	failures, _ := scoreStaleIndex(staleIndexFixture{
		SchemaVersion:           1,
		IndexRevision:           "old",
		WorktreeRevision:        "new",
		DirtyFingerprint:        "dirty",
		AcceptedAsAuthoritative: true,
		FallbackMode:            "authoritative-provider",
	})
	if !containsFailure(failures, "accepted as authoritative") || !containsFailure(failures, "file-native") {
		t.Fatalf("stale authoritative index was not rejected: %v", failures)
	}
}

func TestGraphRoutingEvalFailsMultiWinnerRace(t *testing.T) {
	t.Parallel()
	failures, _ := scoreMultiSessionRace(multiSessionRaceFixture{
		SchemaVersion: 1,
		SameSliceClaims: []raceClaim{
			{Owner: "a", Acquired: true},
			{Owner: "b", Acquired: true},
		},
		DisjointWorktreesIntegrated: true,
		DirtyCheckoutPreserved:      true,
	})
	if !containsFailure(failures, "same-slice race had 2 winners") {
		t.Fatalf("multi-winner race passed: %v", failures)
	}
}

func containsFailure(failures []string, want string) bool {
	for _, failure := range failures {
		if strings.Contains(failure, want) {
			return true
		}
	}
	return false
}
