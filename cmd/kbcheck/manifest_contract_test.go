package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestContractRequiresDoneGate(t *testing.T) {
	path := writeManifest(t, `
---
slices:
  - id: slice-001
    status: done
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "missing-terminal-gate") {
		t.Fatalf("expected missing terminal gate issue, got %#v", result)
	}
}

func TestManifestContractRejectsBadPassedGate(t *testing.T) {
	path := writeManifest(t, `
---
slices:
  - id: slice-001
    status: done
gate_ledger:
  - gate_id: slice-slice-001-to-done
    status: passed
    required_evidence:
      - "proof required"
    proof: []
    blockers:
      - "still blocked"
    passed_at: ""
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "insufficient-proof") || !hasManifestIssue(result.Issues, "blocked-advanceable-gate") {
		t.Fatalf("expected proof and blocker issues, got %#v", result)
	}
}

func TestManifestContractPassesValidDoneAndParkedGates(t *testing.T) {
	proof := filepath.Join(t.TempDir(), "proof.md")
	writeFile(t, proof, "# proof\n")
	path := writeManifest(t, `
---
slices:
  - id: slice-001
    status: done
  - id: slice-002
    status: parked
gate_ledger:
  - gate_id: slice-slice-001-to-done
    status: passed
    required_evidence:
      - "proof exists"
    proof:
      - "`+filepath.ToSlash(proof)+`"
    blockers: []
    passed_at: "2026-06-10"
  - gate_id: slice-slice-002-to-parked
    status: parked
    required_evidence:
      - "parked proof exists"
    proof:
      - "todo.md"
    blockers: []
    passed_at: "2026-06-10"
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected valid manifest, got %#v", result)
	}
}

func TestGateLedgerCommandValidatesAllowedNext(t *testing.T) {
	path := writeManifest(t, `
---
slices:
  - id: slice-001
    status: done
gate_ledger:
  - gate_id: slice-slice-001-to-done
    status: passed
    required_evidence: []
    proof: []
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-complete"
---
`)
	var out, errOut strings.Builder
	code := run([]string{"gate-ledger", "--manifest", path, "--gate", "slice-slice-001-to-done", "--allowed-next", "kb-complete"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("gate-ledger command failed: code=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "PASS: gate=slice-slice-001-to-done") {
		t.Fatalf("missing pass output: %s", out.String())
	}
}

func TestManifestContractValidatesOptInModelTiers(t *testing.T) {
	path := writeManifest(t, `
---
model_tier_contract:
  tiny: deterministic
  small: bounded
slices:
  - id: slice-001
    status: pending
    model_tier: small
  - id: slice-002
    status: pending
    model_tier: giant
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "invalid-model-tier") {
		t.Fatalf("expected invalid model tier issue, got %#v", result)
	}
}

func TestManifestContractRequiresDoneCheckWhenObjectiveContractEnabled(t *testing.T) {
	path := writeManifest(t, `
---
objective_contract: true
slices:
  - id: slice-001
    status: pending
    verification: integration
    proof_check:
      type: command
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "missing-done-check") {
		t.Fatalf("expected missing done check issue, got %#v", result)
	}
}

func TestManifestContractRequiresProofCheckOrExceptionWhenObjectiveContractEnabled(t *testing.T) {
	path := writeManifest(t, `
---
objective_contract: true
done_check:
  type: command
slices:
  - id: slice-001
    status: pending
    verification: integration
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "missing-proof-check") {
		t.Fatalf("expected missing proof check issue, got %#v", result)
	}
}

func TestManifestContractAllowsExplicitNoCheckException(t *testing.T) {
	path := writeManifest(t, `
---
objective_contract: true
done_check:
  type: command
slices:
  - id: slice-001
    status: pending
    verification: verification-only
    no_check_reason: "documentation-only slice with no executable oracle"
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected no-check exception to pass, got %#v", result)
	}
}

func TestManifestContractValidatesModelRouteAgainstTierRoutes(t *testing.T) {
	path := writeManifest(t, `
---
objective_contract: true
done_check:
  type: command
model_tier_contract:
  allowed: [tiny, small, medium, large]
  routes:
    small: ["local-5090-coder"]
slices:
  - id: slice-001
    status: pending
    verification: integration
    model_tier: small
    model_route: hosted-sonnet
    proof_check:
      type: command
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "invalid-model-route") {
		t.Fatalf("expected invalid model route issue, got %#v", result)
	}
}

func TestManifestContractPassesObjectiveContractWithChecksAndRoutes(t *testing.T) {
	path := writeManifest(t, `
---
objective_contract: true
done_check:
  type: command
model_tier_contract:
  allowed: [tiny, small, medium, large]
  routes:
    small: ["local-5090-coder"]
slices:
  - id: slice-001
    status: pending
    verification: integration
    model_tier: small
    model_route: local-5090-coder
    proof_check:
      type: command
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected valid objective contract, got %#v", result)
	}
}
