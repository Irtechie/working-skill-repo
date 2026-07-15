package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type manifestGate struct {
	GateID            string
	OwnerSkill        string
	Status            string
	RequiredEvidence  []string
	Proof             []string
	Blockers          []string
	PassedAt          string
	AllowedNextAction string
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
	objectiveContract := manifestHasObjectiveContract(path)
	contextPacketContract := manifestHasTopLevelKey(path, "context_packet_contract")
	if objectiveContract && !manifestHasTopLevelKey(path, "done_check") {
		issues = append(issues, manifestContractIssue{Code: "missing-done-check", Message: "objective_contract requires a top-level done_check"})
	}
	for _, slice := range slices {
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
		if modelTierContract && !validModelTier(slice.ModelTier) {
			issues = append(issues, manifestContractIssue{Code: "invalid-model-tier", SliceID: slice.ID, Message: "slice must set model_tier to tiny, small, medium, or large"})
		}
		if modelTierContract && strings.TrimSpace(slice.ModelTierReason) == "" {
			issues = append(issues, manifestContractIssue{Code: "missing-model-tier-reason", SliceID: slice.ID, Message: "slice must justify model_tier from complexity, authority, risk, and proof"})
		}
		if modelTierContract && !nonBlankStrings(slice.ModelRequirements) {
			issues = append(issues, manifestContractIssue{Code: "missing-model-requirements", SliceID: slice.ID, Message: "slice must declare model_requirements"})
		}
		if modelTierContract && !nonBlankStrings(slice.EscalationTriggers) {
			issues = append(issues, manifestContractIssue{Code: "missing-escalation-triggers", SliceID: slice.ID, Message: "slice must declare escalation_triggers"})
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
	return manifestContractResult{OK: len(issues) == 0, Issues: issues}, nil
}

func manifestRepoRoot(path string) string {
	current := filepath.Dir(path)
	for {
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

func countIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

func findManifestGate(path, gateID string) (manifestGate, error) {
	gates, err := parseManifestGates(path)
	if err != nil {
		return manifestGate{}, err
	}
	for _, gate := range gates {
		if gate.GateID == gateID {
			return gate, nil
		}
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
		if !proofPathExists(manifestPath, item) {
			issues = append(issues, manifestContractIssue{Code: "missing-proof-path", GateID: gate.GateID, Message: fmt.Sprintf("proof path does not exist: %s", item)})
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
		assignManifestGateValue(current, key, value)
		currentList = ""
	}
	if current != nil {
		gates = append(gates, *current)
	}
	return gates, nil
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
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return strings.TrimSpace(parts[0]), cleanYAMLScalar(parts[1]), true
}

func assignManifestGateValue(gate *manifestGate, key, value string) {
	switch key {
	case "gate_id":
		gate.GateID = value
	case "owner_skill":
		gate.OwnerSkill = value
	case "status":
		gate.Status = value
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
	if filepath.IsAbs(proofItem) {
		return pathExists(proofItem)
	}
	rootRelative := filepath.FromSlash(proofItem)
	if pathExists(rootRelative) {
		return true
	}
	return pathExists(filepath.Join(filepath.Dir(manifestPath), filepath.FromSlash(proofItem)))
}

func hasManifestIssue(issues []manifestContractIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
