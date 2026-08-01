package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/graphrouting"
)

type manifestGate struct {
	GateID             string
	OwnerSkill         string
	GateScope          string
	Status             string
	RequiredEvidence   []string
	Proof              []string
	Blockers           []string
	Attempted          []string
	Responsibility     string
	AffectedScope      string
	ResumeCondition    string
	Recheck            string
	CheckedAt          string
	Propagation        string
	QuarantinedScope   string
	QuarantineOwner    string
	QuarantineEvidence []string
	ForbiddenClaims    []string
	PassedAt           string
	AllowedNextAction  string
}

type manifestContractIssue struct {
	Code    string `json:"code"`
	SliceID string `json:"slice_id,omitempty"`
	GateID  string `json:"gate_id,omitempty"`
	Message string `json:"message"`
}

type manifestContractResult struct {
	OK     bool                    `json:"ok"`
	Issues []manifestContractIssue `json:"issues"`
}

type preSliceReviewArtifact struct {
	ReviewID          string                    `json:"review_id"`
	Source            string                    `json:"source"`
	SourceSHA256      string                    `json:"source_sha256"`
	ReviewedAt        string                    `json:"reviewed_at"`
	DocumentType      string                    `json:"document_type"`
	Mode              string                    `json:"mode"`
	SelectedPersonas  []string                  `json:"selected_personas"`
	CompletedPersonas []string                  `json:"completed_personas"`
	PersonaEvidence   map[string]string         `json:"persona_evidence"`
	FailedPersonas    []string                  `json:"failed_personas"`
	FindingsResolved  int                       `json:"findings_resolved"`
	UnresolvedP0      int                       `json:"unresolved_p0"`
	UnresolvedP1      int                       `json:"unresolved_p1"`
	ResidualFindings  int                       `json:"residual_findings"`
	ResidualItems     []preSliceResidualFinding `json:"residual_items"`
}

type preSliceResidualFinding struct {
	Severity   string `json:"severity"`
	Title      string `json:"title"`
	Constraint string `json:"constraint"`
}

type qualificationPlanRecord struct {
	SchemaVersion int                          `json:"schema_version"`
	Plan          qualificationPlanFileBinding `json:"plan"`
	Review        qualificationPlanFileBinding `json:"review"`
	TargetTier    string                       `json:"target_tier"`
	Invariants    []qualificationPlanInvariant `json:"invariants"`
}

type qualificationPlanFileBinding struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type qualificationPlanInvariant struct {
	ID          string                      `json:"id"`
	Requirement string                      `json:"requirement"`
	Source      qualificationPlanSource     `json:"source"`
	Guidance    *qualificationPlanGuidance  `json:"guidance,omitempty"`
	TierRaise   *qualificationPlanTierRaise `json:"tier_raise,omitempty"`
}

type qualificationPlanSource struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Anchor string `json:"anchor"`
}

type qualificationPlanGuidance struct {
	MechanismOrHazard string `json:"mechanism_or_hazard"`
	ExecutorAction    string `json:"executor_action"`
	ProofTarget       string `json:"proof_target"`
}

type qualificationPlanTierRaise struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

func runManifestContractCommand(root string, opts options, stdout, stderr io.Writer) int {
	path := resolveInputPath(root, opts.manifest)
	result, err := validateManifestContract(path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(result)
	} else if result.OK {
		fmt.Fprintln(stdout, "manifest contract: ok")
	} else {
		for _, issue := range result.Issues {
			fmt.Fprintf(stderr, "%s: %s\n", issue.Code, issue.Message)
		}
	}
	if !result.OK {
		return 2
	}
	return 0
}

func runGateLedgerCommand(root string, opts options, stdout, stderr io.Writer) int {
	path := resolveInputPath(root, opts.manifest)
	gate, err := findManifestGate(path, opts.gate)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	issues := validateAdvanceableGate(path, gate, opts.allowQuarantine)
	lifecycleEnabled, lifecycleValid := manifestBoolContractValue(path, "blocker_lifecycle_contract")
	if !lifecycleValid {
		fmt.Fprintln(stderr, "invalid-contract-boolean: blocker_lifecycle_contract must be true or false")
		return 2
	}
	if lifecycleEnabled {
		issues = append(issues, validateBlockerLifecycleGate(gate, time.Now().UTC())...)
	}
	if opts.allowedNext != "" && gate.AllowedNextAction != opts.allowedNext {
		issues = append(issues, manifestContractIssue{
			Code:    "allowed-next-mismatch",
			GateID:  gate.GateID,
			Message: fmt.Sprintf("allowed_next_action is %q, expected %q", gate.AllowedNextAction, opts.allowedNext),
		})
	}
	if len(issues) > 0 {
		for _, issue := range issues {
			fmt.Fprintf(stderr, "%s: %s\n", issue.Code, issue.Message)
		}
		return 2
	}
	fmt.Fprintf(stdout, "PASS: gate=%s status=%s required=%d proof=%d allowed_next=%s\n", gate.GateID, gate.Status, len(gate.RequiredEvidence), len(gate.Proof), gate.AllowedNextAction)
	return 0
}

func runManifestContractSelftest(stdout, stderr io.Writer) int {
	temp, err := os.MkdirTemp("", "kb-manifest-contract-selftest-*")
	if err != nil {
		fmt.Fprintf(stderr, "create temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(temp)

	proof := filepath.Join(temp, "proof.md")
	if err := os.WriteFile(proof, []byte("# proof\n"), 0o644); err != nil {
		fmt.Fprintf(stderr, "write proof: %v\n", err)
		return 1
	}
	write := func(name, body string) string {
		path := filepath.Join(temp, name)
		_ = os.WriteFile(path, []byte(strings.TrimLeft(body, "\n")), 0o644)
		return path
	}

	valid := write("valid.md", fmt.Sprintf(`
---
slices:
  - id: slice-001
    status: done
    blockers: []
gate_ledger:
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "proof file exists"
    proof:
      - %q
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-complete"
---
`, filepath.ToSlash(proof)))
	result, err := validateManifestContract(valid)
	if err != nil || !result.OK {
		fmt.Fprintf(stderr, "valid manifest failed: result=%#v err=%v\n", result, err)
		return 1
	}
	gate, err := findManifestGate(valid, "slice-slice-001-to-done")
	if err != nil || len(validateAdvanceableGate(valid, gate, false)) != 0 {
		fmt.Fprintf(stderr, "valid gate failed: gate=%#v err=%v\n", gate, err)
		return 1
	}

	missingGate := write("missing-gate.md", `
---
slices:
  - id: slice-001
    status: done
gate_ledger: []
---
`)
	result, err = validateManifestContract(missingGate)
	if err != nil || result.OK || !hasManifestIssue(result.Issues, "missing-terminal-gate") {
		fmt.Fprintf(stderr, "missing gate not rejected: result=%#v err=%v\n", result, err)
		return 1
	}

	badGate := write("bad-gate.md", `
---
slices:
  - id: slice-001
    status: done
gate_ledger:
  - gate_id: slice-slice-001-to-done
    status: passed
    required_evidence:
      - "needs proof"
    proof: []
    blockers:
      - "still blocked"
    passed_at: ""
---
`)
	result, err = validateManifestContract(badGate)
	if err != nil || result.OK || !hasManifestIssue(result.Issues, "insufficient-proof") || !hasManifestIssue(result.Issues, "blocked-advanceable-gate") {
		fmt.Fprintf(stderr, "bad gate not rejected: result=%#v err=%v\n", result, err)
		return 1
	}

	fmt.Fprintln(stdout, "KB manifest contract selftest: passed")
	return 0
}

func validateManifestContract(path string) (manifestContractResult, error) {
	slices, err := parseManifestSlices(path)
	if err != nil {
		return manifestContractResult{}, err
	}
	gates, err := parseManifestGates(path)
	if err != nil {
		return manifestContractResult{}, err
	}
	byID := map[string]manifestGate{}
	for _, gate := range gates {
		byID[gate.GateID] = gate
	}

	issues := []manifestContractIssue{}
	modelTierContract := manifestHasModelTierContract(path)
	modelSelectionContract := manifestHasTopLevelKey(path, "model_selection_contract")
	objectiveContract := manifestHasObjectiveContract(path)
	contextPacketContract := manifestHasTopLevelKey(path, "context_packet_contract")
	impactPacketContract := manifestHasTopLevelKey(path, "impact_packet_contract")
	workspaceIsolationContract := manifestHasTopLevelKey(path, "workspace_isolation_contract")
	proofGovernorContract := manifestHasTopLevelKey(path, "proof_governor_contract")
	preSliceReviewContract, preSliceReviewContractValid := manifestBoolContractValue(path, "pre_slice_review_contract")
	if !preSliceReviewContractValid {
		issues = append(issues, manifestContractIssue{Code: "invalid-contract-boolean", Message: "pre_slice_review_contract must be true or false"})
	}
	qualificationPlanContract, qualificationPlanContractValid := manifestBoolContractValue(path, "qualification_plan_contract")
	if !qualificationPlanContractValid {
		issues = append(issues, manifestContractIssue{Code: "invalid-contract-boolean", Message: "qualification_plan_contract must be true or false"})
	}
	manifestSchema, manifestSchemaQuoted, manifestSchemaDuplicate := manifestTopLevelScalarDetails(path, "manifest_schema")
	if manifestSchema != "" {
		version, err := strconv.Atoi(manifestSchema)
		if err != nil || version < 1 || manifestSchemaQuoted || manifestSchemaDuplicate {
			issues = append(issues, manifestContractIssue{Code: "invalid-manifest-schema", Message: "manifest_schema must be a positive integer"})
		} else if version >= 2 && !preSliceReviewContract {
			issues = append(issues, manifestContractIssue{Code: "missing-pre-slice-review-contract", Message: "manifest_schema 2 or newer requires pre_slice_review_contract: true"})
		}
	}
	blockerLifecycleContract, blockerLifecycleContractValid := manifestBoolContractValue(path, "blocker_lifecycle_contract")
	if !blockerLifecycleContractValid {
		issues = append(issues, manifestContractIssue{Code: "invalid-contract-boolean", Message: "blocker_lifecycle_contract must be true or false"})
	}
	if objectiveContract && !manifestHasTopLevelKey(path, "done_check") {
		issues = append(issues, manifestContractIssue{Code: "missing-done-check", Message: "objective_contract requires a top-level done_check"})
	}
	if modelSelectionContract {
		issues = append(issues, validateModelSelectionContract(path)...)
	}
	if workspaceIsolationContract {
		issues = append(issues, validateWorkspaceIsolationContract(path)...)
	}
	if preSliceReviewContract {
		issues = append(issues, validatePreSliceReviewContract(path)...)
	}
	if qualificationPlanContract {
		issues = append(issues, validateQualificationPlanContract(path)...)
	}
	for _, slice := range slices {
		if slice.TestLevel != "" && !validManifestTestLevel(slice.TestLevel) {
			issues = append(issues, manifestContractIssue{Code: "invalid-test-level", SliceID: slice.ID, Message: "test_level must be none, unit, integration, functional-api, functional-cli, functional-browser, functional-native-gui, or full"})
		}
		if proofGovernorContract {
			if !validProofGovernorExecutionClass(slice.ExecutionClass) {
				issues = append(issues, manifestContractIssue{Code: "invalid-execution-class", SliceID: slice.ID, Message: "proof_governor_contract requires execution_class cli, headless-browser, visible-browser, or native-gui"})
			}
			if slice.TestLevel == "functional-native-gui" && slice.ExecutionClass != "native-gui" {
				issues = append(issues, manifestContractIssue{Code: "native-gui-class-mismatch", SliceID: slice.ID, Message: "functional-native-gui requires execution_class native-gui"})
			}
		}
		if contextPacketContract {
			requiresPacket := slice.Status == "pending" || slice.Status == "in_progress"
			if requiresPacket && slice.ContextPacketPath == "" && slice.NoPacketReason == "" {
				issues = append(issues, manifestContractIssue{Code: "missing-context-packet", SliceID: slice.ID, Message: "pending/in_progress slice requires context_packet_path or no_packet_reason"})
			}
			if slice.ContextPacketPath != "" {
				packetPath := slice.ContextPacketPath
				if !filepath.IsAbs(packetPath) {
					packetPath = filepath.Join(manifestRepoRoot(path), filepath.FromSlash(packetPath))
				}
				var packet contextPacket
				if err := readJSONFile(packetPath, &packet); err != nil {
					issues = append(issues, manifestContractIssue{Code: "missing-context-packet-file", SliceID: slice.ID, Message: err.Error()})
				} else if result := validateContextPacket(packet); !result.OK {
					issues = append(issues, manifestContractIssue{Code: "invalid-context-packet", SliceID: slice.ID, Message: strings.Join(result.Issues, "; ")})
				}
			}
		}
		if impactPacketContract {
			requiresPacket := slice.Status == "pending" || slice.Status == "in_progress"
			if requiresPacket && slice.ImpactPacketPath == "" && slice.NoImpactPacketReason == "" {
				issues = append(issues, manifestContractIssue{Code: "missing-impact-packet", SliceID: slice.ID, Message: "pending/in_progress graph-aware slice requires impact_packet_path or no_impact_packet_reason"})
			}
			if slice.ImpactPacketPath != "" {
				packetPath := slice.ImpactPacketPath
				if !filepath.IsAbs(packetPath) {
					packetPath = filepath.Join(manifestRepoRoot(path), filepath.FromSlash(packetPath))
				}
				packet, err := graphrouting.Load(packetPath)
				if err != nil {
					issues = append(issues, manifestContractIssue{Code: "missing-impact-packet-file", SliceID: slice.ID, Message: err.Error()})
				} else if result := graphrouting.Validate(packet); !result.OK {
					issues = append(issues, manifestContractIssue{Code: "invalid-impact-packet", SliceID: slice.ID, Message: strings.Join(result.Issues, "; ")})
				}
			}
		}
		if modelTierContract && !validModelTier(slice.ModelTier) {
			issues = append(issues, manifestContractIssue{Code: "invalid-model-tier", SliceID: slice.ID, Message: "slice must set model_tier to tiny, small, medium, or large"})
		}
		if modelSelectionContract {
			if strings.TrimSpace(slice.ModelTierReason) == "" {
				issues = append(issues, manifestContractIssue{Code: "missing-model-tier-reason", SliceID: slice.ID, Message: "model_selection_contract requires model_tier_reason"})
			}
			if len(slice.ModelRequirements) == 0 {
				issues = append(issues, manifestContractIssue{Code: "missing-model-requirements", SliceID: slice.ID, Message: "model_selection_contract requires non-empty model_requirements"})
			}
			if len(slice.EscalationTriggers) == 0 {
				issues = append(issues, manifestContractIssue{Code: "missing-escalation-triggers", SliceID: slice.ID, Message: "model_selection_contract requires observable escalation_triggers"})
			}
		}
		if workspaceIsolationContract && requiresWorkspaceIsolationFields(slice) {
			if !validWorkspaceMode(slice.WorkspaceMode) {
				issues = append(issues, manifestContractIssue{Code: "invalid-workspace-mode", SliceID: slice.ID, Message: "workspace_isolation_contract requires workspace_mode shared-serial or worktree-required"})
			}
			if manifestHasPlanRunDefault(path) && slice.WorkspaceMode != "shared-serial" {
				issues = append(issues, manifestContractIssue{Code: "invalid-workspace-mode", SliceID: slice.ID, Message: "plan-run manifests use one worktree per manifest group and require shared-serial slices"})
			}
			if len(slice.ConflictDomains) == 0 {
				issues = append(issues, manifestContractIssue{Code: "missing-conflict-domains", SliceID: slice.ID, Message: "workspace_isolation_contract requires conflict_domains"})
			}
			if slice.WorkspaceMode == "worktree-required" && len(slice.SharedResources) == 0 {
				issues = append(issues, manifestContractIssue{Code: "missing-shared-resources", SliceID: slice.ID, Message: "worktree-required slices must declare shared_resources for serialization or isolation"})
			}
		}
		if objectiveContract && requiresProofCheck(slice) {
			if slice.NoCheckReason != "" {
				if !validNoCheckException(slice) {
					issues = append(issues, manifestContractIssue{Code: "invalid-no-check-exception", SliceID: slice.ID, Message: "no_check_reason is only valid for verification-only or none slices"})
				}
			} else if !slice.ProofCheck {
				issues = append(issues, manifestContractIssue{Code: "missing-proof-check", SliceID: slice.ID, Message: "objective_contract requires proof_check or a justified no_check_reason"})
			}
		}
		switch slice.Status {
		case "done":
			gateID := "slice-" + slice.ID + "-to-done"
			gate, ok := byID[gateID]
			if !ok {
				issues = append(issues, manifestContractIssue{Code: "missing-terminal-gate", SliceID: slice.ID, GateID: gateID, Message: "done slice has no matching to-done gate"})
				continue
			}
			issues = append(issues, validateAdvanceableGate(path, gate, true)...)
		case "parked":
			gateID := "slice-" + slice.ID + "-to-parked"
			gate, ok := byID[gateID]
			if !ok {
				issues = append(issues, manifestContractIssue{Code: "missing-terminal-gate", SliceID: slice.ID, GateID: gateID, Message: "parked slice has no matching to-parked gate"})
				continue
			}
			if gate.Status != "parked" && !isAdvanceableGate(gate.Status, true) {
				issues = append(issues, manifestContractIssue{Code: "invalid-parked-gate-status", SliceID: slice.ID, GateID: gate.GateID, Message: "parked slice gate must be parked, passed, or quarantined"})
			}
			if len(gate.Proof) == 0 {
				issues = append(issues, manifestContractIssue{Code: "missing-proof", SliceID: slice.ID, GateID: gate.GateID, Message: "parked slice gate must record proof"})
			}
		}
	}

	for _, gate := range gates {
		if isAdvanceableGate(gate.Status, true) {
			issues = append(issues, validateAdvanceableGate(path, gate, true)...)
		}
	}
	if blockerLifecycleContract {
		seen := map[string]bool{}
		for _, gate := range gates {
			if seen[gate.GateID] {
				issues = append(issues, manifestContractIssue{
					Code:    "duplicate-gate-id",
					GateID:  gate.GateID,
					Message: "blocker_lifecycle_contract requires unique gate_id values",
				})
			}
			seen[gate.GateID] = true
			issues = append(issues, validateBlockerLifecycleGate(gate, time.Now().UTC())...)
		}
	}
	return manifestContractResult{OK: len(issues) == 0, Issues: issues}, nil
}

func validateBlockerLifecycleGate(gate manifestGate, now time.Time) []manifestContractIssue {
	issues := []manifestContractIssue{}
	add := func(code, message string) {
		issues = append(issues, manifestContractIssue{Code: code, GateID: gate.GateID, Message: message})
	}

	if strings.TrimSpace(gate.GateID) == "" {
		add("missing-gate-id", "blocker_lifecycle_contract requires gate_id")
	}
	if strings.TrimSpace(gate.OwnerSkill) == "" {
		add("missing-gate-owner", "blocker_lifecycle_contract requires owner_skill")
	}
	if !validBlockerLifecycleGateScope(gate.GateScope) {
		add("invalid-gate-scope", "gate_scope must be implementation, integration, release, deployment, or optional-capability")
	}
	if strings.TrimSpace(gate.AllowedNextAction) == "" {
		add("missing-allowed-next-action", "blocker_lifecycle_contract requires allowed_next_action")
	}
	switch gate.Status {
	case "pending", "passed", "quarantined":
	case "blocked", "needs-human":
		if len(gate.Blockers) == 0 {
			add("missing-blocker-detail", "blocked and needs-human gates require at least one blocker")
		}
		if len(gate.Attempted) == 0 {
			add("missing-blocker-attempt", "blocked and needs-human gates must record what was attempted or how human-only ownership was confirmed")
		}
		if strings.TrimSpace(gate.AffectedScope) == "" {
			add("missing-affected-scope", "blocked and needs-human gates require affected_scope")
		}
		if strings.TrimSpace(gate.ResumeCondition) == "" {
			add("missing-resume-condition", "blocked and needs-human gates require resume_condition")
		}
		if strings.TrimSpace(gate.Recheck) == "" {
			add("missing-blocker-recheck", "blocked and needs-human gates require a runnable or inspectable recheck sensor")
		}
		if strings.TrimSpace(gate.CheckedAt) == "" {
			add("missing-blocker-checked-at", "blocked and needs-human gates require checked_at")
		} else if checked, ok := parseGateTimestamp(gate.CheckedAt); !ok {
			add("invalid-blocker-checked-at", "checked_at must be RFC3339 with timezone")
		} else {
			age := now.Sub(checked)
			if age > 72*time.Hour {
				add("stale-blocker", fmt.Sprintf("blocker was last checked %s ago; rerun %q before reporting it", age.Round(time.Hour), gate.Recheck))
			}
			if age < -5*time.Minute {
				add("future-blocker-check", "checked_at is in the future")
			}
		}
		if !validBlockerPropagation(gate.Propagation) {
			add("invalid-blocker-propagation", "propagation must be current-gate-only or dependent-slices-only")
		}
		if gate.Status == "needs-human" && gate.Responsibility != "human" {
			add("misowned-human-gate", "needs-human requires responsibility: human")
		}
		if gate.Status == "blocked" && !oneOf(gate.Responsibility, "agent", "external", "dependency") {
			add("misowned-blocked-gate", "blocked requires responsibility: agent, external, or dependency; use needs-human for human-owned action")
		}
		if oneOf(gate.GateScope, "release", "deployment", "optional-capability") && gate.Propagation != "current-gate-only" {
			add("overpropagated-gate", "release, deployment, and optional-capability blockers may affect only their current gate")
		}
	default:
		add("invalid-gate-status", "status must be pending, blocked, needs-human, quarantined, or passed; pause is execution control state, not a gate result")
	}

	if isAdvanceableGate(gate.Status, true) {
		if _, ok := parseGateTime(gate.PassedAt); !ok {
			add("invalid-passed-at", "passed_at must be YYYY-MM-DD or RFC3339")
		}
		if gate.Responsibility != "" || gate.AffectedScope != "" || gate.ResumeCondition != "" ||
			gate.Recheck != "" || gate.CheckedAt != "" || gate.Propagation != "" || len(gate.Attempted) > 0 {
			add("stale-blocker-metadata", "advanceable gate must clear prior blocker lifecycle metadata")
		}
		if gate.Status == "quarantined" {
			if strings.TrimSpace(gate.QuarantinedScope) == "" {
				add("missing-quarantined-scope", "quarantined gate requires quarantined_scope")
			}
			if strings.TrimSpace(gate.QuarantineOwner) == "" {
				add("missing-quarantine-owner", "quarantined gate requires quarantine_owner")
			}
			if len(gate.QuarantineEvidence) == 0 {
				add("missing-quarantine-evidence", "quarantined gate requires quarantine_evidence")
			}
			if len(gate.ForbiddenClaims) == 0 {
				add("missing-forbidden-claims", "quarantined gate requires forbidden_claims")
			}
		} else if gate.QuarantinedScope != "" || gate.QuarantineOwner != "" ||
			len(gate.QuarantineEvidence) > 0 || len(gate.ForbiddenClaims) > 0 {
			add("stale-quarantine-metadata", "passed gate must clear prior quarantine metadata")
		}
	} else if strings.TrimSpace(gate.PassedAt) != "" {
		add("stale-passed-at", "nonadvanceable gate must not retain passed_at")
	}
	return issues
}

func validBlockerLifecycleGateScope(value string) bool {
	return oneOf(value, "implementation", "integration", "release", "deployment", "optional-capability")
}

func validBlockerPropagation(value string) bool {
	return oneOf(value, "current-gate-only", "dependent-slices-only")
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func parseGateTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func parseGateTimestamp(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func manifestBoolContractValue(path, key string) (bool, bool) {
	frontmatter, err := loadManifestFrontmatter(path)
	if err != nil {
		return false, false
	}
	foundValue := false
	valueResult := false
	for _, line := range strings.Split(frontmatter, "\n") {
		if countIndent(line) != 0 {
			continue
		}
		found, value, quoted, ok := splitYAMLScalarNode(line)
		if ok && found == key {
			if foundValue || quoted {
				return false, false
			}
			foundValue = true
			switch strings.ToLower(strings.TrimSpace(value)) {
			case "true":
				valueResult = true
			case "false":
				valueResult = false
			default:
				return false, false
			}
		}
	}
	return valueResult, true
}

func validManifestTestLevel(value string) bool {
	switch value {
	case "none", "unit", "integration", "functional-api", "functional-cli", "functional-browser", "functional-native-gui", "full":
		return true
	default:
		return false
	}
}

func validateModelSelectionContract(path string) []manifestContractIssue {
	values, err := manifestNestedScalars(path, "model_selection_contract")
	if err != nil {
		return []manifestContractIssue{{Code: "invalid-model-selection-contract", Message: err.Error()}}
	}
	required := map[string]string{
		"timing":                         "work-time",
		"decision_owner":                 "orchestrator",
		"owner_choice":                   "current-or-delegated",
		"max_owner_decisions_per_slice":  "1",
		"catalog":                        "active-host-plus-user-local",
		"delegated_fallback":             "same-tier-then-higher",
		"automatic_downward_routing":     "false",
		"amr_required":                   "false",
		"automatic_cross_owner_fallback": "false",
	}
	issues := []manifestContractIssue{}
	for key, want := range required {
		got, ok := values[key]
		if !ok {
			issues = append(issues, manifestContractIssue{
				Code:    "missing-model-selection-contract-field",
				Message: fmt.Sprintf("model_selection_contract requires %s: %s", key, want),
			})
			continue
		}
		if got != want {
			issues = append(issues, manifestContractIssue{
				Code:    "invalid-model-selection-contract-field",
				Message: fmt.Sprintf("model_selection_contract %s is %q, expected %q", key, got, want),
			})
		}
	}
	return issues
}

func validateWorkspaceIsolationContract(path string) []manifestContractIssue {
	values, err := manifestNestedScalars(path, "workspace_isolation_contract")
	if err != nil {
		return []manifestContractIssue{{Code: "invalid-workspace-isolation-contract", Message: err.Error()}}
	}
	if _, planRunContract := values["plan_run_worktree_default"]; !planRunContract {
		return nil
	}
	required := map[string]string{
		"coordinator_owned_lifecycle":   "true",
		"plan_run_worktree_default":     "true",
		"internal_integration_target":   "plan-run-branch",
		"default_branch_delivery_owner": "kb-complete",
	}
	issues := []manifestContractIssue{}
	for key, want := range required {
		if got := values[key]; got != want {
			issues = append(issues, manifestContractIssue{
				Code:    "invalid-workspace-isolation-contract",
				Message: fmt.Sprintf("workspace_isolation_contract %s is %q, expected %q", key, got, want),
			})
		}
	}
	modes := strings.ReplaceAll(values["allowed_modes"], " ", "")
	if modes != "[shared-serial]" {
		issues = append(issues, manifestContractIssue{
			Code:    "invalid-workspace-isolation-contract",
			Message: "workspace_isolation_contract allowed_modes must be [shared-serial] because the manifest group owns the worktree",
		})
	}
	return issues
}

func validateQualificationPlanContract(path string) []manifestContractIssue {
	add := func(issues []manifestContractIssue, message string) []manifestContractIssue {
		return append(issues, manifestContractIssue{Code: "invalid-qualification-plan", Message: message})
	}
	nodes, err := manifestNestedScalarNodes(path, "qualification_plan")
	if err != nil {
		return []manifestContractIssue{{Code: "invalid-qualification-plan", Message: err.Error()}}
	}
	values := scalarNodeValues(nodes)
	issues := []manifestContractIssue{}
	for key := range values {
		if key != "record_path" && key != "record_sha256" {
			issues = add(issues, fmt.Sprintf("qualification_plan contains unknown field %s", key))
		}
	}
	recordPath := strings.TrimSpace(values["record_path"])
	recordHash := strings.ToLower(strings.TrimSpace(values["record_sha256"]))
	if recordPath == "" || filepath.IsAbs(recordPath) {
		return add(issues, "qualification_plan requires a repo-relative record_path")
	}
	content, err := readQualificationPlanFile(manifestRepoRoot(path), recordPath)
	if err != nil {
		return add(issues, fmt.Sprintf("qualification plan record is unreadable: %v", err))
	}
	if len(content) > 1<<20 {
		return add(issues, "qualification plan record exceeds 1 MiB")
	}
	gotHash := fmt.Sprintf("%x", sha256.Sum256(content))
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(recordHash) || recordHash != gotHash {
		issues = add(issues, "qualification_plan record_sha256 must match the current record")
	}
	if err := validateJSONShape(content, 12); err != nil {
		return add(issues, fmt.Sprintf("invalid qualification plan JSON shape: %v", err))
	}
	var record qualificationPlanRecord
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&record); err != nil {
		return add(issues, fmt.Sprintf("invalid qualification plan JSON: %v", err))
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return add(issues, "invalid qualification plan JSON: trailing content")
	}
	if record.SchemaVersion != 1 {
		issues = add(issues, "qualification plan schema_version must be 1")
	}
	if !validModelTier(record.TargetTier) || record.TargetTier == "tiny" {
		issues = add(issues, "qualification plan target_tier must be small, medium, or large")
	}
	root := manifestRepoRoot(path)
	issues = append(issues, validateQualificationFileBinding(root, "plan", record.Plan)...)
	issues = append(issues, validateQualificationFileBinding(root, "review", record.Review)...)
	issues = append(issues, validateQualificationReviewBinding(path, record.Review)...)
	expectedInvariantIDs, err := qualificationPlanExpectedInvariantIDs(root, record.Plan)
	if err != nil {
		issues = add(issues, err.Error())
	}
	if len(record.Invariants) == 0 {
		issues = add(issues, "qualification plan requires at least one invariant")
	}
	seen := map[string]bool{}
	for index, invariant := range record.Invariants {
		label := fmt.Sprintf("invariant %d", index+1)
		if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{2,127}$`).MatchString(invariant.ID) {
			issues = add(issues, label+" requires a stable kebab-case id")
		} else if seen[invariant.ID] {
			issues = add(issues, fmt.Sprintf("duplicate invariant id %s", invariant.ID))
		}
		seen[invariant.ID] = true
		if strings.TrimSpace(invariant.Requirement) == "" {
			issues = add(issues, label+" requires a requirement")
		}
		if (invariant.Guidance == nil) == (invariant.TierRaise == nil) {
			issues = add(issues, label+" requires exactly one of guidance or tier_raise")
			continue
		}
		if invariant.Guidance != nil {
			issues = append(issues, validateQualificationGuidance(root, label, invariant)...)
		}
		if invariant.TierRaise != nil {
			issues = append(issues, validateQualificationTierRaise(label, record.TargetTier, invariant)...)
		}
	}
	for id := range expectedInvariantIDs {
		if !seen[id] {
			issues = add(issues, fmt.Sprintf("qualification plan record omits reviewed invariant %s", id))
		}
	}
	for id := range seen {
		if !expectedInvariantIDs[id] {
			issues = add(issues, fmt.Sprintf("qualification plan record contains unreviewed invariant %s", id))
		}
	}
	return issues
}

func qualificationPlanExpectedInvariantIDs(root string, binding qualificationPlanFileBinding) (map[string]bool, error) {
	content, err := readQualificationPlanFile(root, binding.Path)
	if err != nil {
		return nil, fmt.Errorf("qualification plan invariant ledger is unreadable: %v", err)
	}
	ids := map[string]bool{}
	inLedger := false
	entryPattern := regexp.MustCompile(`^\s+-\s+([a-z0-9][a-z0-9-]{2,127})\s*$`)
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == "qualification_invariants:" {
			inLedger = true
			continue
		}
		if !inLedger {
			continue
		}
		if match := entryPattern.FindStringSubmatch(line); len(match) == 2 {
			if ids[match[1]] {
				return nil, fmt.Errorf("qualification plan invariant ledger contains duplicate id %s", match[1])
			}
			ids[match[1]] = true
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(line) == len(strings.TrimLeft(line, " \t")) {
			break
		}
		return nil, fmt.Errorf("qualification plan invariant ledger contains an invalid entry")
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("qualification plan must declare qualification_invariants")
	}
	return ids, nil
}

func validateQualificationFileBinding(root, label string, binding qualificationPlanFileBinding) []manifestContractIssue {
	issues := []manifestContractIssue{}
	add := func(message string) {
		issues = append(issues, manifestContractIssue{Code: "invalid-qualification-plan", Message: message})
	}
	content, err := readQualificationPlanFile(root, binding.Path)
	if err != nil {
		add(fmt.Sprintf("qualification plan %s binding is unreadable: %v", label, err))
		return issues
	}
	want := strings.ToLower(strings.TrimSpace(binding.SHA256))
	got := fmt.Sprintf("%x", sha256.Sum256(content))
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(want) || want != got {
		add(fmt.Sprintf("qualification plan %s sha256 must match the current file", label))
	}
	return issues
}

func validateQualificationReviewBinding(path string, binding qualificationPlanFileBinding) []manifestContractIssue {
	nodes, err := manifestNestedScalarNodes(path, "pre_slice_review")
	if err != nil {
		return []manifestContractIssue{{Code: "invalid-qualification-plan", Message: "qualification plans require a passed pre_slice_review receipt"}}
	}
	values := scalarNodeValues(nodes)
	if values["status"] != "passed" ||
		filepath.ToSlash(strings.TrimSpace(binding.Path)) != filepath.ToSlash(strings.TrimSpace(values["review_artifact"])) ||
		strings.ToLower(strings.TrimSpace(binding.SHA256)) != strings.ToLower(strings.TrimSpace(values["review_artifact_sha256"])) {
		return []manifestContractIssue{{Code: "invalid-qualification-plan", Message: "qualification plan review binding must exactly match the passed pre_slice_review artifact"}}
	}
	return nil
}

func validateQualificationGuidance(root, label string, invariant qualificationPlanInvariant) []manifestContractIssue {
	issues := []manifestContractIssue{}
	add := func(message string) {
		issues = append(issues, manifestContractIssue{Code: "invalid-qualification-plan", Message: message})
	}
	sourceContent, err := readQualificationPlanFile(root, invariant.Source.Path)
	if err != nil {
		add(label + " source is unreadable: " + err.Error())
		return issues
	}
	wantSourceHash := strings.ToLower(strings.TrimSpace(invariant.Source.SHA256))
	gotSourceHash := fmt.Sprintf("%x", sha256.Sum256(sourceContent))
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(wantSourceHash) || wantSourceHash != gotSourceHash {
		add(label + " source sha256 must match the current repository file")
	}
	anchor := strings.TrimSpace(invariant.Source.Anchor)
	if anchor == "" {
		add(label + " source requires a stable anchor")
	} else if !strings.Contains(string(sourceContent), anchor) {
		add(label + " source anchor must exist in the bound repository file")
	}
	guidance := invariant.Guidance
	mechanism := strings.TrimSpace(guidance.MechanismOrHazard)
	normalizedRequirement := normalizeQualificationText(invariant.Requirement)
	normalizedMechanism := normalizeQualificationText(mechanism)
	if len(mechanism) < 40 || normalizedMechanism == normalizedRequirement || strings.Contains(normalizedRequirement, normalizedMechanism) {
		add(label + " mechanism_or_hazard must add repository-specific guidance rather than restate the requirement")
	}
	sourceToken := normalizeQualificationText(filepath.Base(invariant.Source.Path))
	anchorToken := normalizeQualificationText(anchor)
	if !strings.Contains(normalizedMechanism, sourceToken) && !strings.Contains(normalizedMechanism, anchorToken) {
		add(label + " mechanism_or_hazard must name the source file or anchor")
	}
	if genericQualificationGuidance(mechanism) {
		add(label + " mechanism_or_hazard is generic")
	}
	if len(strings.TrimSpace(guidance.ExecutorAction)) < 20 {
		add(label + " guidance requires a concrete executor_action")
	}
	if len(strings.TrimSpace(guidance.ProofTarget)) < 10 {
		add(label + " guidance requires a concrete proof_target")
	}
	return issues
}

func validateQualificationTierRaise(label, targetTier string, invariant qualificationPlanInvariant) []manifestContractIssue {
	raise := invariant.TierRaise
	issues := []manifestContractIssue{}
	add := func(message string) {
		issues = append(issues, manifestContractIssue{Code: "invalid-qualification-plan", Message: message})
	}
	if raise.From != targetTier || qualificationTierRank(raise.To) <= qualificationTierRank(raise.From) {
		add(label + " tier_raise must move from target_tier to a higher supported tier")
	}
	reason := strings.TrimSpace(raise.Reason)
	normalizedReason := normalizeQualificationText(reason)
	if len(reason) < 40 || normalizedReason == normalizeQualificationText(invariant.Requirement) {
		add(label + " tier_raise requires a specific uncertainty reason")
	}
	lower := strings.ToLower(reason)
	if !strings.Contains(lower, "uncertain") && !strings.Contains(lower, "unknown") &&
		!strings.Contains(lower, "ambiguous") && !strings.Contains(lower, "unresolved") &&
		!strings.Contains(lower, "risk") {
		add(label + " tier_raise reason must identify the uncertainty or risk")
	}
	return issues
}

func readQualificationPlanFile(root, relative string) ([]byte, error) {
	if strings.TrimSpace(relative) == "" || filepath.IsAbs(relative) {
		return nil, fmt.Errorf("path must be repo-relative")
	}
	clean := filepath.Clean(filepath.Join(root, filepath.FromSlash(relative)))
	rel, err := filepath.Rel(root, clean)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("path escapes the repository")
	}
	info, err := os.Lstat(clean)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("path must not be a symlink")
	}
	return readContainedRepoFile(root, clean)
}

func qualificationTierRank(tier string) int {
	switch tier {
	case "small":
		return 1
	case "medium":
		return 2
	case "large":
		return 3
	default:
		return 0
	}
}

func normalizeQualificationText(value string) string {
	return regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(strings.ToLower(value), "")
}

func genericQualificationGuidance(value string) bool {
	normalized := normalizeQualificationText(value)
	for _, generic := range []string{"becareful", "followrequirements", "usetheplan", "handleedgecases", "testthoroughly"} {
		if strings.Contains(normalized, generic) {
			return true
		}
	}
	return false
}

func validatePreSliceReviewContract(path string) []manifestContractIssue {
	nodes, err := manifestNestedScalarNodes(path, "pre_slice_review")
	if err != nil {
		return []manifestContractIssue{{Code: "invalid-pre-slice-review", Message: err.Error()}}
	}
	values := scalarNodeValues(nodes)
	add := func(issues []manifestContractIssue, message string) []manifestContractIssue {
		return append(issues, manifestContractIssue{Code: "invalid-pre-slice-review", Message: message})
	}
	issues := []manifestContractIssue{}
	for _, key := range []string{"findings_resolved", "unresolved_p0", "unresolved_p1", "residual_findings"} {
		if node, ok := nodes[key]; ok && (node.quoted || !isUnsignedDecimal(node.value)) {
			issues = add(issues, fmt.Sprintf("pre_slice_review %s must be an unquoted non-negative integer", key))
		}
	}
	allowedFields := map[string]bool{
		"status": true, "source": true, "source_sha256": true, "mode": true,
		"review_id": true, "reviewed_at": true, "review_artifact": true,
		"review_artifact_sha256": true, "persona_evidence_json": true,
		"selected_personas_json": true, "completed_personas_json": true,
		"failed_personas_json": true, "findings_resolved": true,
		"unresolved_p0": true, "unresolved_p1": true, "residual_findings": true,
		"not_required_reason": true,
	}
	for key := range values {
		if !allowedFields[key] {
			issues = add(issues, fmt.Sprintf("pre_slice_review contains unknown field %s", key))
		}
	}
	if values["mode"] != "requirements-wide" {
		issues = add(issues, "pre_slice_review mode must be requirements-wide")
	}
	status := values["status"]
	switch status {
	case "passed":
		issues = append(issues, validatePassedPreSliceReview(path, values)...)
	case "not-required":
		if strings.TrimSpace(values["not_required_reason"]) == "" {
			issues = add(issues, "not-required pre_slice_review requires not_required_reason")
		}
	default:
		issues = add(issues, "pre_slice_review status must be passed or not-required")
	}
	return issues
}

func validatePassedPreSliceReview(path string, values map[string]string) []manifestContractIssue {
	add := func(issues []manifestContractIssue, message string) []manifestContractIssue {
		return append(issues, manifestContractIssue{Code: "invalid-pre-slice-review", Message: message})
	}
	issues := []manifestContractIssue{}
	source := strings.TrimSpace(values["source"])
	if source == "" || source == "direct-chat" || filepath.IsAbs(source) {
		issues = add(issues, "passed pre_slice_review requires a repo-relative requirements source")
	} else {
		root := manifestRepoRoot(path)
		sourcePath := filepath.Clean(filepath.Join(root, filepath.FromSlash(source)))
		relative, relErr := filepath.Rel(root, sourcePath)
		if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			issues = add(issues, "passed pre_slice_review source must stay inside the repository")
		} else if content, readErr := readContainedRepoFile(root, sourcePath); readErr != nil {
			issues = add(issues, fmt.Sprintf("passed pre_slice_review source is unreadable: %v", readErr))
		} else {
			want := strings.ToLower(strings.TrimSpace(values["source_sha256"]))
			got := fmt.Sprintf("%x", sha256.Sum256(content))
			if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(want) || want != got {
				issues = add(issues, "passed pre_slice_review source_sha256 must match the current requirements source")
			}
		}
	}
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{5,127}$`).MatchString(values["review_id"]) {
		issues = add(issues, "passed pre_slice_review requires a stable kebab-case review_id")
	}
	if _, parseErr := time.Parse(time.RFC3339, values["reviewed_at"]); parseErr != nil {
		issues = add(issues, "passed pre_slice_review reviewed_at must be RFC3339")
	}
	personas := map[string]string{}
	schemaText, _, _ := manifestTopLevelScalarDetails(path, "manifest_schema")
	schemaVersion, _ := strconv.Atoi(schemaText)
	personaErr := json.Unmarshal([]byte(values["persona_evidence_json"]), &personas)
	if personaErr != nil || len(personas) == 0 {
		issues = add(issues, "passed pre_slice_review persona_evidence_json must be a nonempty JSON object")
	} else if schemaVersion >= 3 && len(personas) != 1 {
		issues = add(issues, "manifest_schema 3 requires exactly one pre-slice reviewer")
	} else {
		allowed := map[string]bool{
			"coherence-reviewer": true, "feasibility-reviewer": true,
			"product-lens-reviewer": true, "design-lens-reviewer": true,
			"spec-flow-analyzer": true, "security-lens-reviewer": true,
			"scope-guardian-reviewer": true, "adversarial-document-reviewer": true,
		}
		for persona, reason := range personas {
			if !allowed[persona] {
				issues = add(issues, fmt.Sprintf("pre_slice_review persona %s is not allowed", persona))
			}
			if !validPersonaSelectionReason(persona, reason) {
				issues = add(issues, fmt.Sprintf("pre_slice_review persona %s requires a specific selection reason", persona))
			}
		}
	}
	var failed []string
	if err := json.Unmarshal([]byte(values["failed_personas_json"]), &failed); err != nil {
		issues = add(issues, "passed pre_slice_review failed_personas_json must be a JSON array")
	} else if len(failed) != 0 {
		issues = add(issues, "passed pre_slice_review cannot contain failed personas")
	}
	selected := decodePersonaList(values["selected_personas_json"], "selected_personas_json", &issues)
	completed := decodePersonaList(values["completed_personas_json"], "completed_personas_json", &issues)
	if !sameStringSet(completed, mapKeys(personas)) {
		issues = add(issues, "completed_personas_json must exactly match persona_evidence_json")
	}
	if !sameStringSet(selected, append(append([]string{}, completed...), failed...)) {
		issues = add(issues, "selected_personas_json must exactly equal completed plus failed personas")
	}
	for _, key := range []string{"findings_resolved", "unresolved_p0", "unresolved_p1", "residual_findings"} {
		count, parseErr := strconv.Atoi(values[key])
		if parseErr != nil || count < 0 {
			issues = add(issues, fmt.Sprintf("pre_slice_review %s must be a nonnegative integer", key))
		}
	}
	for _, key := range []string{"unresolved_p0", "unresolved_p1"} {
		if values[key] != "0" {
			issues = add(issues, fmt.Sprintf("passed pre_slice_review requires %s: 0", key))
		}
	}
	issues = append(issues, validatePreSliceReviewArtifact(path, values, personas, selected, completed, failed)...)
	return issues
}

func validatePreSliceReviewArtifact(path string, values map[string]string, personas map[string]string, selected, completed, failed []string) []manifestContractIssue {
	add := func(issues []manifestContractIssue, message string) []manifestContractIssue {
		return append(issues, manifestContractIssue{Code: "invalid-pre-slice-review", Message: message})
	}
	issues := []manifestContractIssue{}
	root := manifestRepoRoot(path)
	artifact := strings.TrimSpace(values["review_artifact"])
	if artifact == "" || filepath.IsAbs(artifact) {
		return add(issues, "passed pre_slice_review requires a repo-relative review_artifact")
	}
	artifactPath := filepath.Clean(filepath.Join(root, filepath.FromSlash(artifact)))
	relative, relErr := filepath.Rel(root, artifactPath)
	if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return add(issues, "passed pre_slice_review review_artifact must stay inside the repository")
	}
	content, readErr := readContainedRepoFile(root, artifactPath)
	if readErr != nil {
		return add(issues, fmt.Sprintf("passed pre_slice_review review_artifact is unreadable: %v", readErr))
	}
	wantHash := strings.ToLower(strings.TrimSpace(values["review_artifact_sha256"]))
	gotHash := fmt.Sprintf("%x", sha256.Sum256(content))
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(wantHash) || wantHash != gotHash {
		issues = add(issues, "passed pre_slice_review review_artifact_sha256 must match the review artifact")
	}
	var receipt preSliceReviewArtifact
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return add(issues, fmt.Sprintf("invalid pre-slice review artifact: %v", err))
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return add(issues, "invalid pre-slice review artifact: trailing JSON content")
	}
	expectedCounts := map[string]int{
		"findings_resolved": parseNonnegativeInt(values["findings_resolved"]),
		"unresolved_p0":     parseNonnegativeInt(values["unresolved_p0"]),
		"unresolved_p1":     parseNonnegativeInt(values["unresolved_p1"]),
		"residual_findings": parseNonnegativeInt(values["residual_findings"]),
	}
	if receipt.ReviewID != values["review_id"] ||
		receipt.Source != values["source"] ||
		receipt.SourceSHA256 != values["source_sha256"] ||
		receipt.ReviewedAt != values["reviewed_at"] ||
		receipt.DocumentType != "requirements" ||
		receipt.Mode != "requirements-wide" ||
		!sameStringSet(receipt.SelectedPersonas, selected) ||
		!sameStringSet(receipt.CompletedPersonas, completed) ||
		!equalStringMap(receipt.PersonaEvidence, personas) ||
		!equalStringSlice(receipt.FailedPersonas, failed) ||
		receipt.FindingsResolved != expectedCounts["findings_resolved"] ||
		receipt.UnresolvedP0 != expectedCounts["unresolved_p0"] ||
		receipt.UnresolvedP1 != expectedCounts["unresolved_p1"] ||
		receipt.ResidualFindings != expectedCounts["residual_findings"] {
		issues = add(issues, "pre_slice_review manifest fields must exactly match the review artifact")
	}
	if !sameStringSet(receipt.CompletedPersonas, mapKeys(receipt.PersonaEvidence)) ||
		!sameStringSet(receipt.SelectedPersonas, append(append([]string{}, receipt.CompletedPersonas...), receipt.FailedPersonas...)) {
		issues = add(issues, "review artifact persona lifecycle is inconsistent")
	}
	if len(receipt.ResidualItems) != receipt.ResidualFindings {
		issues = add(issues, "review artifact residual_items must match residual_findings")
	}
	for _, item := range receipt.ResidualItems {
		if !oneOf(item.Severity, "P2", "P3") || strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.Constraint) == "" {
			issues = add(issues, "review artifact residual items require P2/P3 severity, title, and actionable constraint")
		}
	}
	return issues
}

func decodePersonaList(raw, field string, issues *[]manifestContractIssue) []string {
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil || len(values) == 0 {
		*issues = append(*issues, manifestContractIssue{
			Code: "invalid-pre-slice-review", Message: fmt.Sprintf("%s must be a nonempty JSON array", field),
		})
	}
	return values
}

func mapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	counts := map[string]int{}
	for _, value := range left {
		counts[value]++
	}
	for _, value := range right {
		counts[value]--
		if counts[value] < 0 {
			return false
		}
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}

func parseNonnegativeInt(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return -1
	}
	return parsed
}

func equalStringMap(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func equalStringSlice(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func manifestHasPlanRunDefault(path string) bool {
	values, err := manifestNestedScalars(path, "workspace_isolation_contract")
	return err == nil && values["plan_run_worktree_default"] == "true"
}

func manifestNestedScalars(path, section string) (map[string]string, error) {
	nodes, err := manifestNestedScalarNodes(path, section)
	if err != nil {
		return nil, err
	}
	return scalarNodeValues(nodes), nil
}

type manifestScalarNode struct {
	value  string
	quoted bool
}

func manifestNestedScalarNodes(path, section string) (map[string]manifestScalarNode, error) {
	frontmatter, err := loadManifestFrontmatter(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(frontmatter, "\n")
	start := -1
	for index, line := range lines {
		if countIndent(line) == 0 && strings.TrimSpace(line) == section+":" {
			if start >= 0 {
				return nil, fmt.Errorf("manifest contains duplicate %s sections", section)
			}
			start = index
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("manifest has no %s section", section)
	}
	values := map[string]manifestScalarNode{}
	for _, line := range lines[start+1:] {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if countIndent(line) == 0 {
			break
		}
		if countIndent(line) != 2 {
			continue
		}
		key, value, quoted, ok := splitYAMLScalarNode(strings.TrimSpace(line))
		if ok {
			if _, exists := values[key]; exists {
				return nil, fmt.Errorf("%s contains duplicate field %s", section, key)
			}
			values[key] = manifestScalarNode{value: value, quoted: quoted}
		}
	}
	return values, nil
}

func scalarNodeValues(nodes map[string]manifestScalarNode) map[string]string {
	values := make(map[string]string, len(nodes))
	for key, node := range nodes {
		values[key] = node.value
	}
	return values
}

func manifestRepoRoot(path string) string {
	current := filepath.Dir(path)
	for {
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current
		}

		if _, err := os.Stat(filepath.Join(current, "config", "skill-quality.json")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return filepath.Dir(path)
		}
		current = parent
	}
}

func readContainedRepoFile(root, path string) ([]byte, error) {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolve repository root: %w", err)
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("resolved path escapes the repository")
	}
	return os.ReadFile(resolvedPath)
}

func manifestHasObjectiveContract(path string) bool {
	frontmatter, err := loadManifestFrontmatter(path)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(frontmatter, "\n") {
		if countIndent(line) != 0 {
			continue
		}
		key, value, ok := splitYAMLKeyValue(line)
		if ok && key == "objective_contract" {
			return parseBool(value)
		}
	}
	return false
}

func manifestHasModelTierContract(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), "model_tier_contract:")
}

func manifestHasTopLevelKey(path, want string) bool {
	frontmatter, err := loadManifestFrontmatter(path)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(frontmatter, "\n") {
		if countIndent(line) != 0 {
			continue
		}
		key, _, ok := splitYAMLKeyValue(line)
		if ok && key == want {
			return true
		}
	}
	return false
}

func manifestTopLevelScalar(path, want string) string {
	value, _, _ := manifestTopLevelScalarDetails(path, want)
	return value
}

func manifestTopLevelScalarDetails(path, want string) (value string, quoted, duplicate bool) {
	frontmatter, err := loadManifestFrontmatter(path)
	if err != nil {
		return "", false, false
	}
	for _, line := range strings.Split(frontmatter, "\n") {
		if countIndent(line) != 0 {
			continue
		}
		key, parsed, parsedQuoted, ok := splitYAMLScalarNode(line)
		if ok && key == want {
			if value != "" {
				duplicate = true
			}
			value = parsed
			quoted = parsedQuoted
		}
	}
	return value, quoted, duplicate
}

func validModelTier(value string) bool {
	switch value {
	case "tiny", "small", "medium", "large":
		return true
	default:
		return false
	}
}

func requiresProofCheck(slice manifestSlice) bool {
	switch slice.Status {
	case "skipped", "parked", "human-required":
		return false
	default:
		return true
	}
}

func validNoCheckException(slice manifestSlice) bool {
	switch slice.Verification {
	case "verification-only", "none":
		return true
	default:
		return false
	}
}

func requiresWorkspaceIsolationFields(slice manifestSlice) bool {
	switch slice.Status {
	case "skipped", "parked", "human-required":
		return false
	default:
		return true
	}
}

func validWorkspaceMode(value string) bool {
	switch value {
	case "shared-serial", "worktree-required":
		return true
	default:
		return false
	}
}

func countIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

func findManifestGate(path, gateID string) (manifestGate, error) {
	gates, err := parseManifestGates(path)
	if err != nil {
		return manifestGate{}, err
	}
	var match *manifestGate
	for i := range gates {
		gate := gates[i]
		if gate.GateID == gateID {
			if match != nil {
				return manifestGate{}, fmt.Errorf("duplicate gate_id %q", gateID)
			}
			match = &gates[i]
		}
	}
	if match != nil {
		return *match, nil
	}
	return manifestGate{}, fmt.Errorf("gate %q not found", gateID)
}

func validateAdvanceableGate(manifestPath string, gate manifestGate, allowQuarantine bool) []manifestContractIssue {
	issues := []manifestContractIssue{}
	if !isAdvanceableGate(gate.Status, allowQuarantine) {
		issues = append(issues, manifestContractIssue{Code: "gate-not-advanceable", GateID: gate.GateID, Message: fmt.Sprintf("status is %q", gate.Status)})
	}
	if len(gate.Proof) < len(gate.RequiredEvidence) {
		issues = append(issues, manifestContractIssue{Code: "insufficient-proof", GateID: gate.GateID, Message: fmt.Sprintf("required evidence=%d proof=%d", len(gate.RequiredEvidence), len(gate.Proof))})
	}
	if len(gate.Blockers) > 0 {
		issues = append(issues, manifestContractIssue{Code: "blocked-advanceable-gate", GateID: gate.GateID, Message: "advanceable gate still has blockers"})
	}
	if strings.TrimSpace(gate.PassedAt) == "" {
		issues = append(issues, manifestContractIssue{Code: "missing-passed-at", GateID: gate.GateID, Message: "advanceable gate has no passed_at"})
	}
	for _, item := range gate.Proof {
		if !looksLikeProofPath(item) {
			continue
		}
		proofPath := resolveManifestProofPath(manifestPath, item)
		if !pathExists(proofPath) {
			issues = append(issues, manifestContractIssue{Code: "missing-proof-path", GateID: gate.GateID, Message: fmt.Sprintf("proof path does not exist: %s", item)})
			continue
		}
		if strings.HasSuffix(strings.ToLower(item), ".proof.json") {
			root := proofGovernorRootFromManifest(manifestPath)
			for _, receiptIssue := range validateProofGovernorReceiptFile(root, proofPath, time.Now().UTC()) {
				issues = append(issues, manifestContractIssue{
					Code:    "invalid-proof-receipt",
					GateID:  gate.GateID,
					Message: fmt.Sprintf("%s: %s", item, receiptIssue),
				})
			}
		}
	}
	return issues
}

func isAdvanceableGate(status string, allowQuarantine bool) bool {
	status = strings.TrimSpace(status)
	if status == "passed" {
		return true
	}
	return allowQuarantine && status == "quarantined"
}

func parseManifestGates(path string) ([]manifestGate, error) {
	frontmatter, err := loadManifestFrontmatter(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(frontmatter, "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "gate_ledger:" {
			start = i
			break
		}
	}
	if start == -1 {
		return nil, nil
	}

	gates := []manifestGate{}
	var current *manifestGate
	currentList := ""
	for _, raw := range lines[start+1:] {
		if raw != "" && !strings.HasPrefix(raw, " ") && !strings.HasPrefix(raw, "\t") && strings.Contains(raw, ":") {
			break
		}
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") && currentList == "" {
			if current != nil {
				gates = append(gates, *current)
			}
			current = &manifestGate{}
			key, value, ok := splitYAMLKeyValue(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
			if ok {
				assignManifestGateValue(current, key, value)
			}
			continue
		}
		if current == nil {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") && currentList != "" {
			appendManifestGateList(current, currentList, cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))))
			continue
		}
		key, value, ok := splitYAMLKeyValue(trimmed)
		if !ok {
			continue
		}
		if value == "" {
			assignManifestGateList(current, key, []string{})
			currentList = key
			continue
		}
		if value == "[]" {
			assignManifestGateList(current, key, []string{})
			currentList = ""
			continue
		}
		if values, ok := parseSimpleInlineList(value); ok {
			assignManifestGateList(current, key, values)
			currentList = ""
			continue
		}
		assignManifestGateValue(current, key, value)
		currentList = ""
	}
	if current != nil {
		gates = append(gates, *current)
	}
	return gates, nil
}

func parseSimpleInlineList(value string) ([]string, bool) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil, false
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
	if inner == "" {
		return []string{}, true
	}
	parts := []string{}
	start := 0
	var quote rune
	escaped := false
	for index, char := range inner {
		if escaped {
			escaped = false
			continue
		}
		if char == '\\' && quote == '"' {
			escaped = true
			continue
		}
		if quote != 0 {
			if char == quote {
				quote = 0
			}
			continue
		}
		if char == '\'' || char == '"' {
			quote = char
			continue
		}
		if char == ',' {
			parts = append(parts, inner[start:index])
			start = index + 1
		}
	}
	if quote != 0 || escaped {
		return nil, false
	}
	parts = append(parts, inner[start:])
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		values = append(values, cleanYAMLScalar(part))
	}
	return values, true
}

func loadManifestFrontmatter(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read manifest: %w", err)
	}
	text := string(content)
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", fmt.Errorf("manifest has no YAML frontmatter")
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[1:i], "\n"), nil
		}
	}
	return "", fmt.Errorf("manifest frontmatter is not closed")
}

func splitYAMLKeyValue(value string) (string, string, bool) {
	key, scalar, _, ok := splitYAMLScalarNode(value)
	return key, scalar, ok
}

func splitYAMLScalarNode(value string) (string, string, bool, bool) {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return "", "", false, false
	}
	raw := strings.TrimSpace(stripYAMLInlineComment(parts[1]))
	quoted := len(raw) >= 2 &&
		((raw[0] == '"' && raw[len(raw)-1] == '"') || (raw[0] == '\'' && raw[len(raw)-1] == '\''))
	return strings.TrimSpace(parts[0]), cleanYAMLScalar(raw), quoted, true
}

func stripYAMLInlineComment(value string) string {
	var quote rune
	escaped := false
	for index, current := range value {
		if escaped {
			escaped = false
			continue
		}
		if quote == '"' && current == '\\' {
			escaped = true
			continue
		}
		if current == '\'' || current == '"' {
			if quote == 0 {
				quote = current
			} else if quote == current {
				quote = 0
			}
			continue
		}
		if current == '#' && quote == 0 && (index == 0 || value[index-1] == ' ' || value[index-1] == '\t') {
			return strings.TrimSpace(value[:index])
		}
	}
	return strings.TrimSpace(value)
}

func isUnsignedDecimal(value string) bool {
	if value == "" {
		return false
	}
	for _, current := range value {
		if current < '0' || current > '9' {
			return false
		}
	}
	return true
}

func validPersonaSelectionReason(persona, reason string) bool {
	basis := map[string]string{
		"coherence-reviewer":            "consistency-risk:",
		"feasibility-reviewer":          "delivery-risk:",
		"product-lens-reviewer":         "product-risk:",
		"design-lens-reviewer":          "interaction-risk:",
		"spec-flow-analyzer":            "flow-risk:",
		"security-lens-reviewer":        "security-risk:",
		"scope-guardian-reviewer":       "scope-risk:",
		"adversarial-document-reviewer": "adversarial-risk:",
	}[persona]
	if basis == "" || !strings.HasPrefix(strings.ToLower(strings.TrimSpace(reason)), basis) {
		return false
	}
	evidence := strings.TrimSpace(reason[len(basis):])
	return len(strings.Fields(evidence)) >= 4
}

func assignManifestGateValue(gate *manifestGate, key, value string) {
	switch key {
	case "gate_id":
		gate.GateID = value
	case "owner_skill":
		gate.OwnerSkill = value
	case "gate_scope":
		gate.GateScope = value
	case "status":
		gate.Status = value
	case "responsibility":
		gate.Responsibility = value
	case "affected_scope":
		gate.AffectedScope = value
	case "resume_condition":
		gate.ResumeCondition = value
	case "recheck":
		gate.Recheck = value
	case "checked_at":
		gate.CheckedAt = value
	case "propagation":
		gate.Propagation = value
	case "quarantined_scope":
		gate.QuarantinedScope = value
	case "quarantine_owner":
		gate.QuarantineOwner = value
	case "passed_at":
		gate.PassedAt = value
	case "allowed_next_action":
		gate.AllowedNextAction = value
	}
}

func assignManifestGateList(gate *manifestGate, key string, values []string) {
	switch key {
	case "required_evidence":
		gate.RequiredEvidence = values
	case "proof":
		gate.Proof = values
	case "blockers":
		gate.Blockers = values
	case "attempted":
		gate.Attempted = values
	case "quarantine_evidence":
		gate.QuarantineEvidence = values
	case "forbidden_claims":
		gate.ForbiddenClaims = values
	}
}

func appendManifestGateList(gate *manifestGate, key, value string) {
	switch key {
	case "required_evidence":
		gate.RequiredEvidence = append(gate.RequiredEvidence, value)
	case "proof":
		gate.Proof = append(gate.Proof, value)
	case "blockers":
		gate.Blockers = append(gate.Blockers, value)
	case "attempted":
		gate.Attempted = append(gate.Attempted, value)
	case "quarantine_evidence":
		gate.QuarantineEvidence = append(gate.QuarantineEvidence, value)
	case "forbidden_claims":
		gate.ForbiddenClaims = append(gate.ForbiddenClaims, value)
	}
}

func looksLikeProofPath(value string) bool {
	if strings.Contains(value, " ") && !strings.ContainsAny(value, `/\`) {
		return false
	}
	matched, _ := regexp.MatchString(`[\\/]|\.md$|\.json$|\.jsonl$|\.txt$|\.log$|\.png$|\.html$|\.ps1$|\.py$|\.go$`, value)
	return matched
}

func proofPathExists(manifestPath, proofItem string) bool {
	return pathExists(resolveManifestProofPath(manifestPath, proofItem))
}

func resolveManifestProofPath(manifestPath, proofItem string) string {
	if filepath.IsAbs(proofItem) {
		return proofItem
	}
	rootRelative := filepath.FromSlash(proofItem)
	if pathExists(rootRelative) {
		return rootRelative
	}
	return filepath.Join(filepath.Dir(manifestPath), filepath.FromSlash(proofItem))
}

func proofGovernorRootFromManifest(manifestPath string) string {
	current, err := filepath.Abs(filepath.Dir(manifestPath))
	if err != nil {
		return filepath.Dir(manifestPath)
	}
	for {
		if pathExists(filepath.Join(current, "go.mod")) || pathExists(filepath.Join(current, ".git")) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return filepath.Dir(manifestPath)
		}
		current = parent
	}
}

func hasManifestIssue(issues []manifestContractIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
