package main

import (
	"os"
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
    model_tier_reason: "bounded mechanical work"
    model_requirements: ["bounded edit", "objective proof"]
    escalation_triggers: ["scope expands"]
  - id: slice-002
    status: pending
    model_tier: giant
    model_tier_reason: "invalid tier fixture"
    model_requirements: ["bounded edit"]
    escalation_triggers: ["scope expands"]
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

func TestManifestContractRequiresModelTierJustification(t *testing.T) {
	path := writeManifest(t, `
---
model_tier_contract:
  allowed: [small, medium, large]
slices:
  - id: slice-001
    status: pending
    model_tier: medium
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"missing-model-tier-reason", "missing-model-requirements", "missing-escalation-triggers"} {
		if !hasManifestIssue(result.Issues, code) {
			t.Fatalf("missing %s issue: %#v", code, result)
		}
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

func TestManifestContractRejectsFalseProofCheck(t *testing.T) {
	path := writeManifest(t, `
---
objective_contract: true
done_check:
  type: command
slices:
  - id: slice-001
    status: pending
    verification: integration
    proof_check: false
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "missing-proof-check") {
		t.Fatalf("false proof_check passed: %#v", result)
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

func TestManifestContractModelRouteAllowsLegacyHintOutsideTierRoutes(t *testing.T) {
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
    model_tier_reason: "bounded compatibility fixture"
    model_requirements: ["objective proof"]
    escalation_triggers: ["scope expands"]
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
	if !result.OK {
		t.Fatalf("expected legacy model_route hint to remain readable, got %#v", result)
	}
}

func TestManifestContractModelRouteAllowsRouteFreeObjectiveContract(t *testing.T) {
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
    model_tier_reason: "bounded route-free fixture"
    model_requirements: ["objective proof"]
    escalation_triggers: ["scope expands"]
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

func TestManifestContractModelRouteDoesNotSubstituteForProofCheck(t *testing.T) {
	path := writeManifest(t, `
---
objective_contract: true
done_check:
  type: command
model_tier_contract:
  allowed: [tiny, small, medium, large]
  routes:
    small: ["local-5090-coder"]
routing_receipt:
  route: local-5090-coder
  provider_model: coder-fast
slices:
  - id: slice-001
    status: pending
    verification: integration
    model_tier: small
    model_route: local-5090-coder
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "missing-proof-check") {
		t.Fatalf("expected proof_check to remain required, got %#v", result)
	}
}

func TestManifestContractTreatsLegacyDDRMetadataAsInertTelemetry(t *testing.T) {
	proof := filepath.Join(t.TempDir(), "proof.md")
	writeFile(t, proof, "# proof\n")
	path := writeManifest(t, `
---
gate_ledger:
  - gate_id: slice-slice-001-to-done
    status: passed
    required_evidence:
      - "proof exists"
    proof:
      - "`+filepath.ToSlash(proof)+`"
    blockers: []
    passed_at: "2026-07-10"
slices:
  - id: slice-001
    status: done
    execution_mode: ddr
    ddr_status: legacy
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("legacy DDR metadata should not create a cosmetic proof gate: %#v", result)
	}
}

func TestManifestContractRequiresPacketForPendingSliceWhenEnabled(t *testing.T) {
	path := writeManifest(t, `
---
context_packet_contract: true
slices:
  - id: slice-001
    status: pending
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "missing-context-packet") {
		t.Fatalf("expected missing context packet issue, got %#v", result)
	}
}

func TestManifestContractValidatesPacketFileWhenEnabled(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), "{}")
	writeFile(t, filepath.Join(root, "packet.json"), `{
	  "schema_version": 1,
	  "packet_id": "p1",
	  "task_id": "t1",
	  "objective": "bounded work",
	  "source_files": ["a.go"],
	  "constraints": ["no daemon"],
	  "out_of_scope": ["unrelated"],
	  "acceptance_criteria": ["passes"],
	  "proof_targets": ["go test ./..."],
	  "model_tier": "small",
	  "model_tier_reason": "bounded",
	  "allowed_tools": ["view"],
	  "broad_search_policy": "bounded",
	  "escalation_triggers": ["scope expands"]
	}`)
	manifestDir := filepath.Join(root, "docs", "plans")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(manifestDir, "manifest.md")
	writeFile(t, path, `---
context_packet_contract: true
slices:
  - id: slice-001
    status: pending
    context_packet_path: packet.json
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil || !result.OK {
		t.Fatalf("expected valid packet-backed manifest, result=%#v err=%v", result, err)
	}
}
