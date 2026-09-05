package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestManifestContractRequiresDoneGate(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

func TestManifestContractRejectsRuntimeInvalidSharedResources(t *testing.T) {
	t.Parallel()
	invalid := writeManifest(t, `
---
slices:
  - id: claim-001
    status: pending
    shared_resources: [manifest-worktree, "resource:has whitespace"]
gate_ledger: []
---
`)
	result, err := validateManifestContract(invalid)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "invalid-shared-resource-claim") {
		t.Fatalf("expected invalid shared-resource claim issue, got %#v", result)
	}

	valid := writeManifest(t, `
---
slices:
  - id: claim-001
    status: pending
    shared_resources: [resource:manifest-worktree]
gate_ledger: []
---
`)
	result, err = validateManifestContract(valid)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected valid shared-resource claim, got %#v", result)
	}
	if _, err := normalizePlanRunClaims(nil, nil, nil, []string{"resource:manifest-worktree"}); err != nil {
		t.Fatalf("runtime lease rejected valid shared-resource claim: %v", err)
	}
}

func TestPreSliceReviewContractAcceptsRequirementsWideReceipt(t *testing.T) {
	t.Parallel()
	path := writePreSliceReviewManifest(t, validPreSliceReviewReceipt())
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected valid pre-slice review receipt, got %#v", result)
	}
}

func TestPreSliceReviewContractAcceptsNotRequiredReceipt(t *testing.T) {
	t.Parallel()
	path := writePreSliceReviewManifest(t, `
pre_slice_review:
  status: not-required
  source: direct-chat
  mode: requirements-wide
  not_required_reason: "one mechanical edit with no product, flow, architecture, scope, or trust decision"
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected valid not-required receipt, got %#v", result)
	}
}

func TestPreSliceReviewContractRejectsInvalidReceipts(t *testing.T) {
	t.Parallel()
	valid := validPreSliceReviewReceipt()
	tests := map[string][2]string{
		"missing-source":            {"source: docs/brainstorms/feature.md", "source: docs/brainstorms/missing.md"},
		"direct-chat-passed-source": {"source: docs/brainstorms/feature.md", "source: direct-chat"},
		"stale-source-hash":         {"source_sha256: {{SOURCE_SHA256}}", "source_sha256: " + strings.Repeat("0", 64)},
		"missing-review-artifact":   {"review_artifact: {{REVIEW_ARTIFACT}}", "review_artifact: docs/results/document-reviews/missing.json"},
		"stale-artifact-hash":       {"review_artifact_sha256: {{REVIEW_ARTIFACT_SHA256}}", "review_artifact_sha256: " + strings.Repeat("0", 64)},
		"wrong-mode":                {"mode: requirements-wide", "mode: per-slice"},
		"invalid-status":            {"status: passed", "status: complete"},
		"invalid-review-id":         {"review_id: feature-requirements-review", "review_id: BAD"},
		"invalid-reviewed-at":       {`reviewed_at: "2026-07-29T20:00:00Z"`, `reviewed_at: yesterday`},
		"scalar-persona-evidence":   {`persona_evidence_json: '{"coherence-reviewer":"consistency-risk: source has cross-section terminology drift"}'`, `persona_evidence_json: 'coherence-reviewer'`},
		"unknown-persona":           {"coherence-reviewer", "implementation-owner"},
		"missing-selection-reason":  {"consistency-risk: source has cross-section terminology drift", ""},
		"failed-persona":            {`failed_personas_json: '[]'`, `failed_personas_json: '[\"security-lens-reviewer\"]'`},
		"negative-count":            {"findings_resolved: 2", "findings_resolved: -1"},
		"noninteger-count":          {"residual_findings: 1", "residual_findings: many"},
		"quoted-count":              {"findings_resolved: 2", `findings_resolved: "2"`},
		"placeholder-reason":        {"consistency-risk: source has cross-section terminology drift", "reviewer applies here"},
		"mismatched-reason-basis":   {"consistency-risk: source has cross-section terminology drift", "security-risk: source changes authentication boundaries"},
		"multiple-reviewers": {
			`persona_evidence_json: '{"coherence-reviewer":"consistency-risk: source has cross-section terminology drift"}'`,
			`persona_evidence_json: '{"coherence-reviewer":"consistency-risk: source has cross-section terminology drift","feasibility-reviewer":"delivery-risk: source has implementation dependency uncertainty"}'`,
		},
		"incomplete-completed-roster": {
			`completed_personas_json: '["coherence-reviewer"]'`,
			`completed_personas_json: '[]'`,
		},
		"unresolved-p0": {"unresolved_p0: 0", "unresolved_p0: 1"},
	}
	for name, replacement := range tests {
		t.Run(name, func(t *testing.T) {
			receipt := strings.Replace(valid, replacement[0], replacement[1], 1)
			path := writePreSliceReviewManifest(t, receipt)
			result, err := validateManifestContract(path)
			if err != nil {
				t.Fatalf("validateManifestContract returned error: %v", err)
			}
			if result.OK || !hasManifestIssue(result.Issues, "invalid-pre-slice-review") {
				t.Fatalf("expected invalid pre-slice review issue, got %#v", result)
			}
		})
	}
}

func TestPreSliceReviewContractRejectsDuplicateReceiptFields(t *testing.T) {
	t.Parallel()
	receipt := strings.Replace(
		validPreSliceReviewReceipt(),
		"  status: passed",
		"  status: not-required\n  status: passed",
		1,
	)
	path := writePreSliceReviewManifest(t, receipt)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "invalid-pre-slice-review") {
		t.Fatalf("expected duplicate receipt field issue, got %#v", result)
	}
}

func TestPreSliceReviewContractRejectsDuplicateSectionOrOptIn(t *testing.T) {
	t.Parallel()
	for name, mutate := range map[string]func(string) string{
		"section": func(receipt string) string {
			return receipt + "\npre_slice_review:\n  status: not-required\n  mode: requirements-wide\n  not_required_reason: duplicate\n"
		},
		"opt-in": func(receipt string) string {
			return strings.Replace(receipt, "pre_slice_review:", "pre_slice_review_contract: false\npre_slice_review:", 1)
		},
	} {
		t.Run(name, func(t *testing.T) {
			path := writePreSliceReviewManifest(t, mutate(validPreSliceReviewReceipt()))
			result, err := validateManifestContract(path)
			if err != nil {
				t.Fatalf("validateManifestContract returned error: %v", err)
			}
			if result.OK {
				t.Fatalf("expected duplicate %s to fail, got %#v", name, result)
			}
		})
	}
}

func TestPreSliceReviewContractAcceptsInlineComments(t *testing.T) {
	t.Parallel()
	receipt := strings.Replace(validPreSliceReviewReceipt(), "  status: passed", "  status: passed # reviewed", 1)
	receipt = strings.Replace(receipt, "  mode: requirements-wide", "  mode: requirements-wide # whole document", 1)
	path := writePreSliceReviewManifest(t, receipt)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected inline YAML comments to pass, got %#v", result)
	}
}

func TestPreSliceReviewContractRejectsSourceSymlinkOutsideRepository(t *testing.T) {
	t.Parallel()
	path := writePreSliceReviewManifest(t, validPreSliceReviewReceipt())
	root := filepath.Dir(filepath.Dir(filepath.Dir(path)))
	sourcePath := filepath.Join(root, "docs", "brainstorms", "feature.md")
	outside := filepath.Join(t.TempDir(), "feature.md")
	writeFile(t, outside, "# Feature requirements\n")
	if err := os.Remove(sourcePath); err != nil {
		t.Fatalf("remove source before symlink: %v", err)
	}
	if err := os.Symlink(outside, sourcePath); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "invalid-pre-slice-review") {
		t.Fatalf("expected escaping source symlink issue, got %#v", result)
	}
}

func TestManifestSchemaTwoRequiresPreSliceReviewContract(t *testing.T) {
	t.Parallel()
	path := writeManifest(t, `
---
manifest_schema: 2
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
	if result.OK || !hasManifestIssue(result.Issues, "missing-pre-slice-review-contract") {
		t.Fatalf("expected schema contract issue, got %#v", result)
	}
}

func TestManifestContractRejectsQuotedOrDuplicateSchema(t *testing.T) {
	t.Parallel()
	for name, schema := range map[string]string{
		"quoted":    `manifest_schema: "2"`,
		"duplicate": "manifest_schema: 1\nmanifest_schema: 2",
	} {
		t.Run(name, func(t *testing.T) {
			path := writeManifest(t, `
---
`+schema+`
pre_slice_review_contract: true
slices: []
gate_ledger: []
---
`)
			result, err := validateManifestContract(path)
			if err != nil {
				t.Fatalf("validateManifestContract returned error: %v", err)
			}
			if result.OK || !hasManifestIssue(result.Issues, "invalid-manifest-schema") {
				t.Fatalf("expected invalid schema issue, got %#v", result)
			}
		})
	}
}

func TestLegacyManifestWithoutSchemaDoesNotRequirePreSliceReview(t *testing.T) {
	t.Parallel()
	path := writeManifest(t, `
---
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
	if !result.OK {
		t.Fatalf("expected legacy manifest compatibility, got %#v", result)
	}
}

func TestQualificationPlanContractAcceptsSpecificGuidance(t *testing.T) {
	t.Parallel()
	path := writeQualificationPlanManifest(t, nil)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected valid qualification plan, got %#v", result)
	}
}

func TestQualificationPlanContractAcceptsExplicitTierRaise(t *testing.T) {
	t.Parallel()
	path := writeQualificationPlanManifest(t, func(record map[string]any) {
		invariant := record["invariants"].([]any)[0].(map[string]any)
		delete(invariant, "guidance")
		invariant["tier_raise"] = map[string]any{
			"from":   "medium",
			"to":     "large",
			"reason": "Uncertainty remains around manifest YAML alias behavior, so stronger reasoning is required before execution.",
		}
	})
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected explicit tier raise to pass, got %#v", result)
	}
}

func TestQualificationPlanContractRejectsWeakOrStaleRecords(t *testing.T) {
	t.Parallel()
	tests := map[string]func(map[string]any){
		"restated-requirement": func(record map[string]any) {
			invariant := record["invariants"].([]any)[0].(map[string]any)
			invariant["guidance"].(map[string]any)["mechanism_or_hazard"] = invariant["requirement"]
		},
		"generic-hazard": func(record map[string]any) {
			record["invariants"].([]any)[0].(map[string]any)["guidance"].(map[string]any)["mechanism_or_hazard"] = "Be careful with this requirement."
		},
		"empty-executor-action": func(record map[string]any) {
			record["invariants"].([]any)[0].(map[string]any)["guidance"].(map[string]any)["executor_action"] = ""
		},
		"missing-proof-target": func(record map[string]any) {
			delete(record["invariants"].([]any)[0].(map[string]any)["guidance"].(map[string]any), "proof_target")
		},
		"stale-plan-hash": func(record map[string]any) {
			record["plan"].(map[string]any)["sha256"] = strings.Repeat("0", 64)
		},
		"duplicate-invariant": func(record map[string]any) {
			invariants := record["invariants"].([]any)
			record["invariants"] = append(invariants, invariants[0])
		},
		"missing-reviewed-invariant": func(record map[string]any) {
			record["invariants"] = []any{}
		},
		"absent-source-anchor": func(record map[string]any) {
			record["invariants"].([]any)[0].(map[string]any)["source"].(map[string]any)["anchor"] = "missingAnchor"
		},
		"unknown-field": func(record map[string]any) {
			record["unexpected"] = true
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			path := writeQualificationPlanManifest(t, mutate)
			result, err := validateManifestContract(path)
			if err != nil {
				t.Fatalf("validateManifestContract returned error: %v", err)
			}
			if result.OK || !hasManifestIssue(result.Issues, "invalid-qualification-plan") {
				t.Fatalf("expected invalid qualification plan issue, got %#v", result)
			}
		})
	}
}

func TestQualificationPlanContractRejectsStaleReviewBinding(t *testing.T) {
	t.Parallel()
	path := writeQualificationPlanManifest(t, func(record map[string]any) {
		record["review"].(map[string]any)["sha256"] = strings.Repeat("0", 64)
	})
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "invalid-qualification-plan") {
		t.Fatalf("expected stale review binding issue, got %#v", result)
	}
}

func TestQualificationPlanContractRejectsEscapedOrStaleRecord(t *testing.T) {
	t.Parallel()
	for name, mutate := range map[string]func(string) string{
		"escaped-path": func(manifest string) string {
			return strings.Replace(manifest, "record_path: docs/plans/qualification-plan-record.json", "record_path: ../../qualification-plan-record.json", 1)
		},
		"stale-record-hash": func(manifest string) string {
			return regexp.MustCompile(`record_sha256: [0-9a-f]{64}`).ReplaceAllString(manifest, "record_sha256: "+strings.Repeat("0", 64))
		},
	} {
		t.Run(name, func(t *testing.T) {
			path := writeQualificationPlanManifest(t, nil)
			writeFile(t, path, mutate(string(mustReadFile(t, path))))
			result, err := validateManifestContract(path)
			if err != nil {
				t.Fatalf("validateManifestContract returned error: %v", err)
			}
			if result.OK || !hasManifestIssue(result.Issues, "invalid-qualification-plan") {
				t.Fatalf("expected invalid qualification record issue, got %#v", result)
			}
		})
	}
}

func TestQualificationPlanContractRejectsDuplicateJSONKeys(t *testing.T) {
	t.Parallel()
	mutations := map[string]func(string) string{
		"top-level": func(record string) string {
			return strings.Replace(record, `"schema_version": 1,`, `"schema_version": 1,`+"\n  "+`"schema_version": 1,`, 1)
		},
		"nested": func(record string) string {
			return strings.Replace(record, `"path": "docs/plans/qualification-evidence-plan.md",`, `"path": "docs/plans/qualification-evidence-plan.md",`+"\n    "+`"path": "docs/plans/qualification-evidence-plan.md",`, 1)
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			path := writeQualificationPlanManifest(t, nil)
			root := filepath.Dir(filepath.Dir(filepath.Dir(path)))
			recordPath := filepath.Join(root, "docs", "plans", "qualification-plan-record.json")
			record := mutate(string(mustReadFile(t, recordPath)))
			writeFile(t, recordPath, record)
			manifest := string(mustReadFile(t, path))
			manifest = regexp.MustCompile(`record_sha256: [0-9a-f]{64}`).ReplaceAllString(
				manifest,
				fmt.Sprintf("record_sha256: %x", sha256.Sum256([]byte(record))),
			)
			writeFile(t, path, manifest)

			result, err := validateManifestContract(path)
			if err != nil {
				t.Fatalf("validateManifestContract returned error: %v", err)
			}
			if result.OK || !hasManifestIssue(result.Issues, "invalid-qualification-plan") {
				t.Fatalf("duplicate JSON key was accepted: %#v", result)
			}
		})
	}
}

func writeQualificationPlanManifest(t *testing.T, mutate func(map[string]any)) string {
	t.Helper()
	path := writePreSliceReviewManifest(t, validPreSliceReviewReceipt())
	root := filepath.Dir(filepath.Dir(filepath.Dir(path)))
	planRelative := "docs/plans/qualification-evidence-plan.md"
	planContent := []byte("# Qualification evidence plan\n\nqualification_invariants:\n  - strict-manifest-input\n\nUse the manifest parser and focused contract tests.\n")
	writeFile(t, filepath.Join(root, filepath.FromSlash(planRelative)), string(planContent))
	reviewRelative := "docs/results/document-reviews/feature-requirements-review.json"
	reviewContent := mustReadFile(t, filepath.Join(root, filepath.FromSlash(reviewRelative)))
	sourceRelative := "cmd/kbcheck/manifest_contract.go"
	sourceContent := []byte("func validateManifestContract(path string) {}\n")
	writeFile(t, filepath.Join(root, filepath.FromSlash(sourceRelative)), string(sourceContent))
	record := map[string]any{
		"schema_version": 1,
		"plan": map[string]any{
			"path":   planRelative,
			"sha256": fmt.Sprintf("%x", sha256.Sum256(planContent)),
		},
		"review": map[string]any{
			"path":   reviewRelative,
			"sha256": fmt.Sprintf("%x", sha256.Sum256(reviewContent)),
		},
		"target_tier": "medium",
		"invariants": []any{map[string]any{
			"id":          "strict-manifest-input",
			"requirement": "Qualification evidence must reject malformed or stale plan records.",
			"source": map[string]any{
				"path":   sourceRelative,
				"sha256": fmt.Sprintf("%x", sha256.Sum256(sourceContent)),
				"anchor": "validateManifestContract",
			},
			"guidance": map[string]any{
				"mechanism_or_hazard": "manifest_contract.go parses opt-in records before terminal-gate validation, so stale evidence cannot be accepted later.",
				"executor_action":     "Add strict decoding and contained-path hash checks in validateManifestContract.",
				"proof_target":        "go test ./cmd/kbcheck -run QualificationPlan -count=1",
			},
		}},
	}
	if mutate != nil {
		mutate(record)
	}
	recordContent, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("marshal qualification plan: %v", err)
	}
	recordRelative := "docs/plans/qualification-plan-record.json"
	writeFile(t, filepath.Join(root, filepath.FromSlash(recordRelative)), string(recordContent))
	manifest := string(mustReadFile(t, path))
	contract := fmt.Sprintf("qualification_plan_contract: true\nqualification_plan:\n  record_path: %s\n  record_sha256: %x\n", recordRelative, sha256.Sum256(recordContent))
	manifest = strings.Replace(manifest, "slices:\n", contract+"slices:\n", 1)
	writeFile(t, path, manifest)
	return path
}

func validPreSliceReviewReceipt() string {
	return `
pre_slice_review:
  status: passed
  source: docs/brainstorms/feature.md
  source_sha256: {{SOURCE_SHA256}}
  mode: requirements-wide
  review_id: feature-requirements-review
  reviewed_at: "2026-07-29T20:00:00Z"
  review_artifact: {{REVIEW_ARTIFACT}}
  review_artifact_sha256: {{REVIEW_ARTIFACT_SHA256}}
  persona_evidence_json: '{"coherence-reviewer":"consistency-risk: source has cross-section terminology drift"}'
  selected_personas_json: '["coherence-reviewer"]'
  completed_personas_json: '["coherence-reviewer"]'
  failed_personas_json: '[]'
  findings_resolved: 2
  unresolved_p0: 0
  unresolved_p1: 0
  residual_findings: 1
  not_required_reason: ""
`
}

func writePreSliceReviewManifest(t *testing.T, receipt string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatalf("create fake git root: %v", err)
	}
	source := []byte("# Feature requirements\n")
	sourceRelative := "docs/brainstorms/feature.md"
	sourceHash := fmt.Sprintf("%x", sha256.Sum256(source))
	writeFile(t, filepath.Join(dir, filepath.FromSlash(sourceRelative)), string(source))
	personas := map[string]string{
		"coherence-reviewer": "consistency-risk: source has cross-section terminology drift",
	}
	artifact := preSliceReviewArtifact{
		ReviewID:          "feature-requirements-review",
		Source:            sourceRelative,
		SourceSHA256:      sourceHash,
		ReviewedAt:        "2026-07-29T20:00:00Z",
		DocumentType:      "requirements",
		Mode:              "requirements-wide",
		SelectedPersonas:  []string{"coherence-reviewer"},
		CompletedPersonas: []string{"coherence-reviewer"},
		PersonaEvidence:   personas,
		FailedPersonas:    []string{},
		FindingsResolved:  2,
		UnresolvedP0:      0,
		UnresolvedP1:      0,
		ResidualFindings:  1,
		ResidualItems: []preSliceResidualFinding{{
			Severity: "P2", Title: "Retain rollout observation", Constraint: "Carry the observation into the implementation context packet.",
		}},
	}
	artifactContent, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		t.Fatalf("marshal review artifact: %v", err)
	}
	artifactRelative := "docs/results/document-reviews/feature-requirements-review.json"
	writeFile(t, filepath.Join(dir, filepath.FromSlash(artifactRelative)), string(artifactContent))
	receipt = strings.ReplaceAll(receipt, "{{SOURCE_SHA256}}", sourceHash)
	receipt = strings.ReplaceAll(receipt, "{{REVIEW_ARTIFACT}}", artifactRelative)
	receipt = strings.ReplaceAll(receipt, "{{REVIEW_ARTIFACT_SHA256}}", fmt.Sprintf("%x", sha256.Sum256(artifactContent)))
	path := filepath.Join(dir, "docs", "plans", "manifest.md")
	writeFile(t, path, strings.TrimLeft(`
---
manifest_schema: 3
pre_slice_review_contract: true
`+receipt+`
slices:
  - id: slice-001
    status: pending
gate_ledger: []
---
`, "\n"))
	return path
}

func TestGateLedgerCommandValidatesAllowedNext(t *testing.T) {
	t.Parallel()
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

func TestBlockerLifecycleContractAcceptsScopedHumanAndExternalGates(t *testing.T) {
	t.Parallel()
	checkedAt := time.Now().UTC().Format(time.RFC3339)
	path := writeManifest(t, `
---
blocker_lifecycle_contract: true
slices: []
gate_ledger:
  - gate_id: provider-login
    owner_skill: kb-work
    gate_scope: optional-capability
    status: needs-human
    required_evidence: []
    proof: []
    blockers: ["provider-owned login is required"]
    attempted: ["confirmed no token-free activation path exists"]
    responsibility: human
    affected_scope: "optional live provider activation"
    resume_condition: "user completes provider-owned login"
    recheck: "inspect provider activation receipt"
    checked_at: "`+checkedAt+`"
    propagation: current-gate-only
    allowed_next_action: "await provider login while unrelated slices continue"
  - gate_id: linux-package
    owner_skill: kb-work
    gate_scope: release
    status: blocked
    required_evidence: []
    proof: []
    blockers:
      - "controller has no Linux package executor"
    attempted:
      - "inspected the live controller capability catalog"
    responsibility: external
    affected_scope: "Linux release receipt only"
    resume_condition: "controller publishes a package executor"
    recheck: "fleet snapshot package capability"
    checked_at: "`+checkedAt+`"
    propagation: current-gate-only
    allowed_next_action: "continue implementation and Windows proof"
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected valid blocker lifecycle contract, got %#v", result)
	}
}

func TestBlockerLifecycleContractRejectsMisownedStaleAndOverpropagatedGates(t *testing.T) {
	t.Parallel()
	stale := time.Now().UTC().Add(-96 * time.Hour).Format(time.RFC3339)
	path := writeManifest(t, `
---
blocker_lifecycle_contract: true
slices: []
gate_ledger:
  - gate_id: agent-repair-called-human
    owner_skill: kb-work
    gate_scope: implementation
    status: needs-human
    blockers: ["PowerShell wrapper has an exit-code bug"]
    attempted: ["reproduced the wrapper failure"]
    responsibility: agent
    affected_scope: "Windows smoke receipt"
    resume_condition: "repair the wrapper"
    recheck: "run focused wrapper test"
    checked_at: "`+stale+`"
    propagation: dependent-slices-only
    allowed_next_action: "kb-fix wrapper"
  - gate_id: release-rollup
    owner_skill: kb-complete
    gate_scope: release
    status: blocked
    blockers: ["macOS receipt is unavailable"]
    attempted: ["confirmed macOS is not a supported platform"]
    responsibility: external
    affected_scope: "macOS release only"
    resume_condition: "add macOS to the supported matrix"
    recheck: "inspect supported platform matrix"
    checked_at: "`+time.Now().UTC().Format(time.RFC3339)+`"
    propagation: dependent-slices-only
    allowed_next_action: "continue core implementation"
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	for _, code := range []string{"misowned-human-gate", "stale-blocker", "overpropagated-gate"} {
		if !hasManifestIssue(result.Issues, code) {
			t.Fatalf("expected %s, got %#v", code, result)
		}
	}
}

func TestBlockerLifecycleContractRejectsPauseAsGateStatus(t *testing.T) {
	t.Parallel()
	path := writeManifest(t, `
---
blocker_lifecycle_contract: true
slices: []
gate_ledger:
  - gate_id: user-pause
    owner_skill: kb-work
    gate_scope: implementation
    status: paused
    blockers: []
    allowed_next_action: "wait for resume"
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "invalid-gate-status") {
		t.Fatalf("pause must remain execution control state, got %#v", result)
	}
}

func TestBlockerLifecycleContractParsesInlineBlockerLists(t *testing.T) {
	t.Parallel()
	checkedAt := time.Now().UTC().Format(time.RFC3339)
	path := writeManifest(t, `
---
blocker_lifecycle_contract: true
slices: []
gate_ledger:
  - gate_id: dependency
    owner_skill: kb-work
    gate_scope: integration
    status: blocked
    blockers: ["upstream contract unavailable"]
    attempted: ["checked upstream contract"]
    responsibility: dependency
    affected_scope: "dependent integration slice"
    resume_condition: "upstream contract exists"
    recheck: "test -f contract.json"
    checked_at: "`+checkedAt+`"
    propagation: dependent-slices-only
    allowed_next_action: "continue unrelated slices"
---
`)
	gates, err := parseManifestGates(path)
	if err != nil {
		t.Fatalf("parseManifestGates returned error: %v", err)
	}
	if len(gates) != 1 || len(gates[0].Blockers) != 1 || len(gates[0].Attempted) != 1 {
		t.Fatalf("inline lists were not parsed: %#v", gates)
	}
}

func TestBlockerLifecycleContractRejectsInvalidBooleanAndDuplicateGateIDs(t *testing.T) {
	t.Parallel()
	path := writeManifest(t, `
---
blocker_lifecycle_contract: ture
slices: []
gate_ledger:
  - gate_id: duplicate
    status: passed
  - gate_id: duplicate
    status: blocked
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "invalid-contract-boolean") {
		t.Fatalf("invalid lifecycle boolean was accepted: %#v", result)
	}
	if _, err := findManifestGate(path, "duplicate"); err == nil || !strings.Contains(err.Error(), "duplicate gate_id") {
		t.Fatalf("duplicate gate lookup was accepted: %v", err)
	}
}

func TestBlockerLifecycleContractRequiresQuarantineBoundary(t *testing.T) {
	t.Parallel()
	path := writeManifest(t, `
---
blocker_lifecycle_contract: true
slices: []
gate_ledger:
  - gate_id: macos-release
    owner_skill: kb-complete
    gate_scope: optional-capability
    status: quarantined
    required_evidence: ["Windows and Linux are proven"]
    proof: ["release receipts recorded"]
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "ship supported platforms"
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	for _, code := range []string{"missing-quarantined-scope", "missing-quarantine-owner", "missing-quarantine-evidence", "missing-forbidden-claims"} {
		if !hasManifestIssue(result.Issues, code) {
			t.Fatalf("expected %s, got %#v", code, result)
		}
	}

	valid := writeManifest(t, `
---
blocker_lifecycle_contract: true
slices: []
gate_ledger:
  - gate_id: macos-release
    owner_skill: kb-complete
    gate_scope: optional-capability
    status: quarantined
    required_evidence: ["Windows and Linux are proven"]
    proof: ["release receipts recorded"]
    blockers: []
    quarantined_scope: "macOS packaging only"
    quarantine_owner: "release engineering"
    quarantine_evidence: ["supported-platform matrix"]
    forbidden_claims: ["macOS package is available"]
    passed_at: "2026-07-26"
    allowed_next_action: "ship supported platforms"
---
`)
	result, err = validateManifestContract(valid)
	if err != nil || !result.OK {
		t.Fatalf("valid quarantine boundary failed: result=%#v err=%v", result, err)
	}
}

func TestBlockerLifecycleContractRejectsStalePassAndDateOnlyCheck(t *testing.T) {
	t.Parallel()
	path := writeManifest(t, `
---
blocker_lifecycle_contract: true
slices: []
gate_ledger:
  - gate_id: stale-pending
    owner_skill: kb-work
    gate_scope: implementation
    status: pending
    passed_at: "2026-07-26"
    allowed_next_action: "run proof"
  - gate_id: date-only-blocker
    owner_skill: kb-work
    gate_scope: implementation
    status: blocked
    blockers: ["dependency unavailable"]
    attempted: ["checked dependency"]
    responsibility: dependency
    affected_scope: "integration slice"
    resume_condition: "dependency responds"
    recheck: "curl dependency health"
    checked_at: "2026-07-26"
    propagation: dependent-slices-only
    allowed_next_action: "continue unrelated slices"
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if !hasManifestIssue(result.Issues, "stale-passed-at") ||
		!hasManifestIssue(result.Issues, "invalid-blocker-checked-at") {
		t.Fatalf("expected stale pass and strict timestamp issues, got %#v", result)
	}
}

func TestBlockerLifecycleContractPreservesColonAndCommaListItems(t *testing.T) {
	t.Parallel()
	path := writeManifest(t, `
---
blocker_lifecycle_contract: true
slices: []
gate_ledger:
  - gate_id: dependency
    owner_skill: kb-work
    gate_scope: integration
    status: blocked
    blockers: ["Windows, Linux receipts missing"]
    attempted:
      - command: go test ./cmd/kbcheck
    responsibility: dependency
    affected_scope: "dependent integration slice"
    resume_condition: "receipts exist"
    recheck: "go test ./cmd/kbcheck"
    checked_at: "`+time.Now().UTC().Format(time.RFC3339)+`"
    propagation: dependent-slices-only
    allowed_next_action: "continue unrelated slices"
---
`)
	gates, err := parseManifestGates(path)
	if err != nil {
		t.Fatalf("parseManifestGates returned error: %v", err)
	}
	if len(gates) != 1 || len(gates[0].Blockers) != 1 ||
		gates[0].Blockers[0] != "Windows, Linux receipts missing" ||
		len(gates[0].Attempted) != 1 ||
		gates[0].Attempted[0] != "command: go test ./cmd/kbcheck" {
		t.Fatalf("list items were corrupted: %#v", gates)
	}
}

func TestManifestContractValidatesOptInModelTiers(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	if !result.OK {
		t.Fatalf("expected legacy model_route hint to remain readable, got %#v", result)
	}
}

func TestManifestContractModelRouteAllowsRouteFreeObjectiveContract(t *testing.T) {
	t.Parallel()
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

func TestManifestContractRequiresModelSelectionMetadataWhenEnabled(t *testing.T) {
	t.Parallel()
	path := writeManifest(t, `
---
objective_contract: true
done_check:
  type: command
model_tier_contract:
  allowed: [small, medium, large]
model_selection_contract:
  timing: work-time
slices:
  - id: slice-001
    status: pending
    verification: integration
    model_tier: medium
    proof_check:
      type: command
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "missing-model-tier-reason") ||
		!hasManifestIssue(result.Issues, "missing-model-requirements") ||
		!hasManifestIssue(result.Issues, "missing-escalation-triggers") {
		t.Fatalf("expected model selection metadata issues, got %#v", result)
	}
}

func TestManifestContractAcceptsOrchestratorDirectedModelSelectionContract(t *testing.T) {
	t.Parallel()
	path := writeManifest(t, `
---
model_tier_contract:
  allowed: [small, medium, large]
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  owner_choice: current-or-delegated
  max_owner_decisions_per_slice: 1
  catalog: active-host-plus-user-local
  delegated_fallback: same-tier-then-higher
  automatic_downward_routing: false
  automatic_cross_owner_fallback: false
slices:
  - id: slice-001
    status: pending
    model_tier: medium
    model_tier_reason: "bounded implementation needs repository reasoning"
    model_requirements: ["apply_patch", "go test"]
    escalation_triggers: ["scope expands beyond owned files"]
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("valid orchestrator-directed contract failed: %#v", result)
	}
}

func TestManifestContractRejectsDownwardOrCrossOwnerFallback(t *testing.T) {
	t.Parallel()
	path := writeManifest(t, `
---
model_selection_contract:
  timing: work-time
  decision_owner: worker
  owner_choice: automatic
  max_owner_decisions_per_slice: 3
  catalog: remembered-model-names
  delegated_fallback: lower-then-current
  automatic_downward_routing: true
  automatic_cross_owner_fallback: true
slices: []
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "invalid-model-selection-contract-field") {
		t.Fatalf("unsafe routing contract was accepted: %#v", result)
	}
}

func TestManifestContractRejectsEachUnsafeRoutingMutation(t *testing.T) {
	t.Parallel()
	validContract := `
---
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  owner_choice: current-or-delegated
  max_owner_decisions_per_slice: 1
  catalog: active-host-plus-user-local
  delegated_fallback: same-tier-then-higher
  automatic_downward_routing: false
  automatic_cross_owner_fallback: false
slices: []
gate_ledger: []
---
`
	mutations := map[string][2]string{
		"catalog":                    {"catalog: active-host-plus-user-local", "catalog: remembered-model-names"},
		"delegated-fallback":         {"delegated_fallback: same-tier-then-higher", "delegated_fallback: lower-then-current"},
		"automatic-downward-routing": {"automatic_downward_routing: false", "automatic_downward_routing: true"},
		"cross-owner-fallback":       {"automatic_cross_owner_fallback: false", "automatic_cross_owner_fallback: true"},
	}
	for name, mutation := range mutations {
		t.Run(name, func(t *testing.T) {
			path := writeManifest(t, strings.Replace(validContract, mutation[0], mutation[1], 1))
			result, err := validateManifestContract(path)
			if err != nil {
				t.Fatal(err)
			}
			if result.OK || !hasManifestIssue(result.Issues, "invalid-model-selection-contract-field") {
				t.Fatalf("unsafe %s mutation was accepted: %#v", name, result)
			}
		})
	}
}

func TestManifestContractValidatesWorkspaceIsolationIntent(t *testing.T) {
	t.Parallel()
	valid := writeManifest(t, `
---
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
slices:
  - id: slice-001
    status: pending
    workspace_mode: worktree-required
    conflict_domains: [file:src/a.go]
    shared_resources: [git:integration-owner]
gate_ledger: []
---
`)
	result, err := validateManifestContract(valid)
	if err != nil || !result.OK {
		t.Fatalf("expected valid workspace isolation fields, result=%#v err=%v", result, err)
	}

	invalid := writeManifest(t, `
---
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
slices:
  - id: slice-001
    status: pending
    workspace_mode: live-edit
    conflict_domains: []
gate_ledger: []
---
`)
	result, err = validateManifestContract(invalid)
	if err != nil {
		t.Fatalf("validateManifestContract returned error: %v", err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "invalid-workspace-mode") ||
		!hasManifestIssue(result.Issues, "missing-conflict-domains") {
		t.Fatalf("expected workspace isolation issues, got %#v", result)
	}
}

func TestPlanRunManifestContractRejectsIncompleteWorkspaceIntent(t *testing.T) {
	t.Parallel()
	valid := writeManifest(t, `
---
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
  plan_run_worktree_default: true
  internal_integration_target: plan-run-branch
  default_branch_delivery_owner: kb-complete
  allowed_modes: [shared-serial]
slices: []
gate_ledger: []
---
`)
	result, err := validateManifestContract(valid)
	if err != nil || !result.OK {
		t.Fatalf("expected valid plan-run contract, result=%#v err=%v", result, err)
	}

	for name, mutation := range map[string][2]string{
		"coordinator": {"coordinator_owned_lifecycle: true", "coordinator_owned_lifecycle: false"},
		"default":     {"plan_run_worktree_default: true", "plan_run_worktree_default: false"},
		"integration": {"internal_integration_target: plan-run-branch", "internal_integration_target: default-branch"},
		"delivery":    {"default_branch_delivery_owner: kb-complete", "default_branch_delivery_owner: kb-work"},
		"modes":       {"allowed_modes: [shared-serial]", "allowed_modes: [worktree-required]"},
	} {
		t.Run(name, func(t *testing.T) {
			path := writeManifest(t, strings.Replace(string(mustReadFile(t, valid)), mutation[0], mutation[1], 1))
			got, err := validateManifestContract(path)
			if err != nil {
				t.Fatal(err)
			}
			if got.OK || !hasManifestIssue(got.Issues, "invalid-workspace-isolation-contract") {
				t.Fatalf("unsafe plan-run contract was accepted: %#v", got)
			}
		})
	}
}

func TestManifestContractModelRouteDoesNotSubstituteForProofCheck(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

func TestManifestContractValidatesOptionalImpactPacketContract(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), "{}")
	writeFile(t, filepath.Join(root, "impact.json"), `{
	  "schema_version": 1,
	  "packet_id": "impact",
	  "repository": {
	    "identity": "git:file:///fixture",
	    "root": ".",
	    "vcs": "git",
	    "revision": "abc123",
	    "dirty_fingerprint": "clean",
	    "worktree_fingerprint": "main:abc123",
	    "freshness": "fresh"
	  },
	  "seeds": {"files": ["src/api/payments.go"]},
	  "edges": [],
	  "direct_impact": [],
	  "reverse_impact": [],
	  "tests": [],
	  "docs": [],
	  "fallback": {"mode": "file-native", "reason": "fixture fallback"},
	  "budget": {"max_edges": 10, "max_bytes": 1000, "truncated": false},
	  "limitations": ["fixture"],
	  "generated_by": "test",
	  "generated_at": "2026-07-19T00:00:00Z"
	}`)
	manifestDir := filepath.Join(root, "docs", "plans")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(manifestDir, "manifest.md")
	writeFile(t, path, `---
impact_packet_contract:
  optional: true
slices:
  - id: slice-001
    status: pending
    impact_packet_path: impact.json
  - id: slice-002
    status: pending
    no_impact_packet_reason: "file-native fallback only"
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil || !result.OK {
		t.Fatalf("expected valid impact packet manifest, result=%#v err=%v", result, err)
	}

	missing := filepath.Join(manifestDir, "missing-impact.md")
	writeFile(t, missing, `---
impact_packet_contract:
  optional: true
slices:
  - id: slice-001
    status: pending
gate_ledger: []
---
`)
	result, err = validateManifestContract(missing)
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "missing-impact-packet") {
		t.Fatalf("expected missing impact packet issue, got %#v", result)
	}
}
