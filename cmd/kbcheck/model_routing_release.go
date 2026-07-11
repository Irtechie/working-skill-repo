package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	maxModelRoutingReleaseBytes int64 = 1 << 20
	modelRoutingProofTimeout          = 10 * time.Minute
)

type modelRoutingReleaseEvidence struct {
	SchemaVersion     int                         `json:"schema_version"`
	Cohort            string                      `json:"cohort"`
	EvidenceMode      string                      `json:"evidence_mode"`
	ModelProvenance   string                      `json:"model_provenance"`
	PaidCalls         int                         `json:"paid_calls"`
	Baseline          string                      `json:"baseline"`
	ReleaseDecision   string                      `json:"release_decision"`
	NextLowerAttempts string                      `json:"next_lower_attempts"`
	LiveSupportStatus string                      `json:"live_support_status"`
	Router            modelRoutingRouterStatus    `json:"router"`
	SupportedCohorts  []string                    `json:"supported_cohorts"`
	SupportClaims     []modelRoutingSupportClaim  `json:"support_claims"`
	ParkedSurfaces    []string                    `json:"parked_surfaces"`
	Fixtures          []modelRoutingFixtureRef    `json:"fixtures"`
	Metrics           modelRoutingReleaseMetrics  `json:"metrics"`
	Promotion         modelRoutingPromotionStatus `json:"promotion"`
}

type modelRoutingRouterStatus struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type modelRoutingFileRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type modelRoutingFixtureRef struct {
	Class              string `json:"class"`
	Path               string `json:"path"`
	SHA256             string `json:"sha256"`
	Live               bool   `json:"live"`
	EfficiencyEvidence bool   `json:"efficiency_evidence"`
	SupportEvidence    bool   `json:"support_evidence"`
}

type modelRoutingSupportClaim struct {
	Cohort          string               `json:"cohort"`
	Status          string               `json:"status"`
	Adapter         string               `json:"adapter"`
	Runtime         string               `json:"runtime"`
	SelectionPolicy string               `json:"selection_policy"`
	ProofHarness    string               `json:"proof_harness"`
	RouteRevision   string               `json:"route_revision"`
	EvidenceSource  string               `json:"evidence_source"`
	LiveReceipt     *modelRoutingFileRef `json:"live_receipt,omitempty"`
	InstallProof    *modelRoutingFileRef `json:"install_proof,omitempty"`
}

type modelRoutingMetric struct {
	Status string   `json:"status"`
	Value  *float64 `json:"value,omitempty"`
	Unit   string   `json:"unit,omitempty"`
}

type modelRoutingReleaseMetrics struct {
	TotalBilledCost modelRoutingMetric `json:"total_billed_cost"`
	TotalTokens     modelRoutingMetric `json:"total_tokens"`
	WallClockMS     modelRoutingMetric `json:"wall_clock_ms"`
}

type modelRoutingPromotionStatus struct {
	PreregisteredGatesMet               bool     `json:"preregistered_gates_met"`
	IndependentHoldout                  bool     `json:"independent_holdout"`
	PowerJustified                      bool     `json:"power_justified"`
	LiveSampleSize                      int      `json:"live_sample_size"`
	TaskFamilies                        int      `json:"task_families"`
	ZeroRightToWrong                    bool     `json:"zero_right_to_wrong"`
	ConfidenceBoundMaxRegressionPercent *float64 `json:"confidence_bound_max_regression_percent,omitempty"`
	PrimaryMetric                       string   `json:"primary_metric"`
	MedianImprovementPercent            *float64 `json:"median_improvement_percent,omitempty"`
	CorrectionSafety                    string   `json:"correction_safety"`
}

type modelRoutingPilotFixture struct {
	SchemaVersion       int                     `json:"schema_version"`
	Suite               string                  `json:"suite"`
	EvidenceClass       string                  `json:"evidence_class"`
	ExecutionCapability string                  `json:"execution_capability"`
	Live                bool                    `json:"live"`
	EfficiencyEvidence  bool                    `json:"efficiency_evidence"`
	Cases               []modelRoutingPilotCase `json:"cases"`
}

type modelRoutingProofResult struct {
	ExitCode  int
	Output    string
	Err       error
	Truncated bool
}

type modelRoutingProofRunner func(context.Context, string, []string) modelRoutingProofResult

var modelRoutingDeterministicProofCommands = [][]string{
	{"go", "test", "./internal/modelrouting", "-run", "^(TestSelectRouteRespectsPlannedAndAttemptTiersOverridesAndEvidenceStrength|TestSelectRouteAcceptsOnlyTheExactNextLowerAttemptTier|TestCorrectionPilotIneligibleRoutesToSeparateOrdinaryExecution)$", "-count=1"},
	{"go", "test", "./cmd/kbrouter", "-run", "^(TestModelsSelectReportsExplicitAttemptAndPlannedCorrectionTiers|TestDispatchCorrectionPacketRefusesBeforeWorkerLaunchOrReceipt|TestFallbackUsesFreshFallbackRouteModelAndNeverDownward)$", "-count=1"},
}

type modelRoutingPilotCase struct {
	ID         string `json:"id"`
	TaskFamily string `json:"task_family"`
	Scenario   string `json:"scenario,omitempty"`
	Expected   string `json:"expected"`
}

type modelRoutingLiveReceipt struct {
	SchemaVersion     int    `json:"schema_version"`
	EvidenceClass     string `json:"evidence_class"`
	Cohort            string `json:"cohort"`
	Adapter           string `json:"adapter"`
	Runtime           string `json:"runtime"`
	SelectionPolicy   string `json:"selection_policy"`
	ProofHarness      string `json:"proof_harness"`
	RouteRevision     string `json:"route_revision"`
	RouteFingerprint  string `json:"route_fingerprint"`
	ContextPacketHash string `json:"context_packet_hash"`
	WorkProofHash     string `json:"work_proof_hash"`
	ProofStatus       string `json:"proof_status"`
	Live              bool   `json:"live"`
}

type modelRoutingInstallProof struct {
	SchemaVersion   int      `json:"schema_version"`
	EvidenceClass   string   `json:"evidence_class"`
	Cohort          string   `json:"cohort"`
	AdapterRevision string   `json:"adapter_revision"`
	InstalledHash   string   `json:"installed_hash"`
	Platforms       []string `json:"platforms"`
}

func runModelRoutingReleaseCommand(root string, opts options, stdout, stderr io.Writer) int {
	return runModelRoutingReleaseCommandWithRunner(root, opts, stdout, stderr, runModelRoutingProductionProof)
}

func runModelRoutingReleaseCommandWithRunner(root string, opts options, stdout, stderr io.Writer, runner modelRoutingProofRunner) int {
	if strings.TrimSpace(opts.cohort) == "" || strings.TrimSpace(opts.evidencePath) == "" {
		fmt.Fprintln(stderr, "model-routing-release requires --cohort and --evidence")
		return 2
	}
	evidencePath, err := resolveModelRoutingReleaseFile(root, opts.evidencePath)
	if err != nil {
		fmt.Fprintf(stderr, "model-routing release evidence: %v\n", err)
		return 1
	}
	var evidence modelRoutingReleaseEvidence
	if err := readStrictModelRoutingJSON(evidencePath, &evidence); err != nil {
		fmt.Fprintf(stderr, "model-routing release evidence: %v\n", err)
		return 1
	}
	if err := validateModelRoutingRelease(root, opts.cohort, evidence); err != nil {
		fmt.Fprintf(stderr, "model-routing release evidence: %v\n", err)
		return 1
	}
	if err := runModelRoutingDeterministicProof(root, runner); err != nil {
		fmt.Fprintf(stderr, "model-routing release evidence: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "model-routing-release: honest cohort=%s decision=%s supported_cohorts=%d paid_calls=%d\n", evidence.Cohort, evidence.ReleaseDecision, len(evidence.SupportedCohorts), evidence.PaidCalls)
	return 0
}

func validateModelRoutingRelease(root, cohort string, evidence modelRoutingReleaseEvidence) error {
	if evidence.SchemaVersion != 1 || evidence.Cohort != cohort || cohort != "initial-pilot" {
		return fmt.Errorf("unsupported schema or cohort")
	}
	if evidence.Baseline != "planned-tier-host-native" {
		return fmt.Errorf("planned-tier host-native baseline is required")
	}
	// This release validator currently has no independently attested live-evidence
	// verifier. Keep the only accepted state deliberately narrow: deterministic,
	// no-paid evidence may prove conformance, but it may not qualify support or
	// enable AMR. A future external verifier must introduce a separate schema and
	// validation boundary before attended-live claims can become releasable.
	if evidence.EvidenceMode != "deterministic-no-paid" {
		return fmt.Errorf("live evidence is not accepted without an external verifier")
	}
	if evidence.PaidCalls < 0 {
		return fmt.Errorf("paid_calls cannot be negative")
	}
	if err := validateModelRoutingMetrics(evidence.Metrics); err != nil {
		return err
	}
	if err := validateParkedSurfaces(evidence.ParkedSurfaces); err != nil {
		return err
	}
	if err := validateModelRoutingFixtureRefs(root, evidence.Fixtures); err != nil {
		return err
	}
	if err := validateModelRoutingSupportClaims(root, evidence); err != nil {
		return err
	}
	if evidence.ModelProvenance != "unavailable" {
		return fmt.Errorf("model provenance must remain unavailable when no route-bound provenance exists")
	}
	if evidence.PaidCalls != 0 {
		return fmt.Errorf("deterministic-no-paid evidence requires paid_calls=0")
	}
	if evidence.LiveSupportStatus != "not-qualified" || len(evidence.SupportedCohorts) != 0 || len(evidence.SupportClaims) != 0 {
		return fmt.Errorf("deterministic-no-paid evidence cannot qualify live support")
	}
	if evidence.Router.Status != "unavailable" || evidence.Router.Reason != "private-acl-access-denied" {
		return fmt.Errorf("no-paid pilot must record router unavailable/private-acl-access-denied")
	}
	if evidence.ReleaseDecision != "not-promoted" || evidence.NextLowerAttempts != "disabled" {
		return fmt.Errorf("no-paid evidence must remain not-promoted with next-lower attempts disabled")
	}
	return nil
}

func validateModelRoutingMetrics(metrics modelRoutingReleaseMetrics) error {
	values := map[string]modelRoutingMetric{
		"total_billed_cost": metrics.TotalBilledCost,
		"total_tokens":      metrics.TotalTokens,
		"wall_clock_ms":     metrics.WallClockMS,
	}
	for name, metric := range values {
		switch metric.Status {
		case "unavailable":
			if metric.Value != nil || metric.Unit != "" {
				return fmt.Errorf("unavailable metric %s must omit value and unit; zero is not unavailable", name)
			}
		case "measured":
			if metric.Value == nil || *metric.Value < 0 || strings.TrimSpace(metric.Unit) == "" {
				return fmt.Errorf("measured metric %s requires a nonnegative value and unit", name)
			}
		default:
			return fmt.Errorf("metric %s must be measured or unavailable", name)
		}
	}
	return nil
}

func validateParkedSurfaces(actual []string) error {
	required := []string{"codex-app-exact-attribution", "direct-chat-completions", "ghcp", "mcp-dispatch", "tinyboss-control"}
	copyActual := append([]string(nil), actual...)
	sort.Strings(copyActual)
	if strings.Join(copyActual, "\x00") != strings.Join(required, "\x00") {
		return fmt.Errorf("parked surfaces must explicitly retain unsupported GHCP, Codex App, TinyBoss, MCP, and direct chat-completions claims")
	}
	return nil
}

func validateModelRoutingFixtureRefs(root string, refs []modelRoutingFixtureRef) error {
	if len(refs) != 2 {
		return fmt.Errorf("exactly two deterministic fixture references are required")
	}
	wantSuites := map[string]string{
		"deterministic-conformance": "model-routing-initial-pilot",
		"seeded-correction-safety":  "model-routing-correction-pilot",
	}
	wantCapabilities := map[string]string{
		"deterministic-conformance": "non-live-routing-conformance",
		"seeded-correction-safety":  "handoff-validation-and-refusal-only",
	}
	seen := map[string]bool{}
	for _, ref := range refs {
		if seen[ref.Class] || wantSuites[ref.Class] == "" {
			return fmt.Errorf("invalid or duplicate fixture class %q", ref.Class)
		}
		seen[ref.Class] = true
		if ref.Live {
			return fmt.Errorf("deterministic fixture must be non-live")
		}
		if ref.EfficiencyEvidence {
			return fmt.Errorf("deterministic or correction fixture cannot count as efficiency evidence")
		}
		if ref.SupportEvidence {
			return fmt.Errorf("deterministic or correction fixture cannot count as support evidence")
		}
		path, err := resolveModelRoutingReleaseFile(root, ref.Path)
		if err != nil {
			return fmt.Errorf("fixture %s: %w", ref.Class, err)
		}
		var fixture modelRoutingPilotFixture
		if err := readStrictHashedModelRoutingJSON(path, ref.SHA256, &fixture); err != nil {
			return fmt.Errorf("fixture %s: %w", ref.Class, err)
		}
		if fixture.SchemaVersion != 1 || fixture.Suite != wantSuites[ref.Class] || fixture.EvidenceClass != ref.Class || fixture.ExecutionCapability != wantCapabilities[ref.Class] || fixture.Live || fixture.EfficiencyEvidence {
			return fmt.Errorf("fixture must be non-live, non-efficiency, and match its declared class")
		}
		if len(fixture.Cases) == 0 {
			return fmt.Errorf("fixture %s has no cases", ref.Class)
		}
		caseIDs := map[string]bool{}
		for _, testCase := range fixture.Cases {
			if strings.TrimSpace(testCase.ID) == "" || caseIDs[testCase.ID] || strings.TrimSpace(testCase.TaskFamily) == "" || strings.TrimSpace(testCase.Expected) == "" {
				return fmt.Errorf("fixture %s contains an invalid case", ref.Class)
			}
			caseIDs[testCase.ID] = true
		}
	}
	return nil
}

func validateModelRoutingSupportClaims(root string, evidence modelRoutingReleaseEvidence) error {
	claimByCohort := map[string]modelRoutingSupportClaim{}
	for _, claim := range evidence.SupportClaims {
		lower := strings.ToLower(claim.Cohort + " " + claim.Adapter + " " + claim.Runtime)
		for _, forbidden := range []string{"ghcp", "tinyboss", "mcp", "chat-completions", "codex-app"} {
			if strings.Contains(lower, forbidden) {
				return fmt.Errorf("unsupported cohort or dispatch surface %q", claim.Cohort)
			}
		}
		if claim.Status != "supported" || claim.EvidenceSource != "live-route-bound" {
			return fmt.Errorf("supported cohort requires route-bound live evidence")
		}
		if strings.TrimSpace(claim.Cohort) == "" || strings.TrimSpace(claim.Adapter) == "" || strings.TrimSpace(claim.Runtime) == "" ||
			strings.TrimSpace(claim.SelectionPolicy) == "" || strings.TrimSpace(claim.ProofHarness) == "" || strings.TrimSpace(claim.RouteRevision) == "" {
			return fmt.Errorf("supported cohort binding is incomplete")
		}
		if claim.LiveReceipt == nil || claim.InstallProof == nil {
			return fmt.Errorf("supported cohort requires a bound live receipt and install proof")
		}
		if _, duplicate := claimByCohort[claim.Cohort]; duplicate {
			return fmt.Errorf("duplicate support claim for %s", claim.Cohort)
		}
		if err := validateModelRoutingLiveReceipt(root, claim); err != nil {
			return err
		}
		if err := validateModelRoutingInstallProof(root, claim); err != nil {
			return err
		}
		claimByCohort[claim.Cohort] = claim
	}
	if len(evidence.SupportedCohorts) != len(claimByCohort) {
		return fmt.Errorf("supported cohort list does not match qualified support claims")
	}
	seen := map[string]bool{}
	for _, cohort := range evidence.SupportedCohorts {
		if seen[cohort] || claimByCohort[cohort].Cohort == "" {
			return fmt.Errorf("supported cohort %q lacks a unique qualified claim", cohort)
		}
		seen[cohort] = true
	}
	return nil
}

func validateModelRoutingLiveReceipt(root string, claim modelRoutingSupportClaim) error {
	path, err := resolveAndHashModelRoutingRef(root, *claim.LiveReceipt)
	if err != nil {
		return fmt.Errorf("live receipt: %w", err)
	}
	var receipt modelRoutingLiveReceipt
	if err := readStrictHashedModelRoutingJSON(path, claim.LiveReceipt.SHA256, &receipt); err != nil {
		return fmt.Errorf("live receipt: %w", err)
	}
	if receipt.SchemaVersion != 1 || receipt.EvidenceClass != "route-bound-live-receipt" || !receipt.Live || receipt.ProofStatus != "pass" ||
		receipt.Cohort != claim.Cohort || receipt.Adapter != claim.Adapter || receipt.Runtime != claim.Runtime ||
		receipt.SelectionPolicy != claim.SelectionPolicy || receipt.ProofHarness != claim.ProofHarness || receipt.RouteRevision != claim.RouteRevision ||
		!validReleaseHash(receipt.RouteFingerprint) || !validReleaseHash(receipt.ContextPacketHash) || !validReleaseHash(receipt.WorkProofHash) {
		return fmt.Errorf("live receipt is not exact, route-bound, and proof-passing")
	}
	return nil
}

func validateModelRoutingInstallProof(root string, claim modelRoutingSupportClaim) error {
	path, err := resolveAndHashModelRoutingRef(root, *claim.InstallProof)
	if err != nil {
		return fmt.Errorf("install proof: %w", err)
	}
	var proof modelRoutingInstallProof
	if err := readStrictHashedModelRoutingJSON(path, claim.InstallProof.SHA256, &proof); err != nil {
		return fmt.Errorf("install proof: %w", err)
	}
	if proof.SchemaVersion != 1 || proof.EvidenceClass != "install-proof" || proof.Cohort != claim.Cohort || proof.AdapterRevision != claim.RouteRevision ||
		!validReleaseHash(proof.InstalledHash) || len(proof.Platforms) == 0 {
		return fmt.Errorf("install proof is incomplete or does not match the supported cohort")
	}
	return nil
}

func validateModelRoutingPromotion(evidence modelRoutingReleaseEvidence) error {
	promotion := evidence.Promotion
	if !promotion.PreregisteredGatesMet || !promotion.IndependentHoldout || !promotion.PowerJustified || !promotion.ZeroRightToWrong ||
		promotion.LiveSampleSize < 20 || promotion.TaskFamilies < 3 || promotion.ConfidenceBoundMaxRegressionPercent == nil ||
		*promotion.ConfidenceBoundMaxRegressionPercent > 2 || promotion.MedianImprovementPercent == nil || *promotion.MedianImprovementPercent < 20 ||
		promotion.CorrectionSafety != "qualified" || len(evidence.SupportClaims) == 0 {
		return fmt.Errorf("promotion gates are not satisfied")
	}
	metrics := map[string]modelRoutingMetric{
		"total_billed_cost": evidence.Metrics.TotalBilledCost,
		"wall_clock_ms":     evidence.Metrics.WallClockMS,
	}
	metric, allowed := metrics[promotion.PrimaryMetric]
	if !allowed {
		return fmt.Errorf("promotion primary metric is not preregistered and comparable")
	}
	if metric.Status != "measured" || metric.Value == nil {
		return fmt.Errorf("promotion primary metric must be measured")
	}
	if *metric.Value <= 0 {
		return fmt.Errorf("promotion primary metric must have a nonzero measured baseline")
	}
	return nil
}

func runModelRoutingDeterministicProof(root string, runner modelRoutingProofRunner) error {
	if runner == nil {
		return fmt.Errorf("deterministic proof runner is required")
	}
	for _, command := range modelRoutingDeterministicProofCommands {
		ctx, cancel := context.WithTimeout(context.Background(), modelRoutingProofTimeout)
		result := runner(ctx, root, append([]string(nil), command...))
		deadline := ctx.Err()
		cancel()
		if deadline != nil {
			return fmt.Errorf("deterministic proof command timed out: %s", strings.Join(command, " "))
		}
		if result.Err != nil || result.ExitCode != 0 || result.Truncated || !strings.Contains(result.Output, "ok") {
			return fmt.Errorf("deterministic proof command failed: %s exit=%d truncated=%t error=%v output=%q", strings.Join(command, " "), result.ExitCode, result.Truncated, result.Err, result.Output)
		}
	}
	return nil
}

func runModelRoutingProductionProof(ctx context.Context, root string, command []string) modelRoutingProofResult {
	if len(command) < 2 || command[0] != "go" || command[1] != "test" {
		return modelRoutingProofResult{ExitCode: 1, Err: fmt.Errorf("refused non-fixed proof command")}
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = root
	if err := configureCheckProcessTree(cmd); err != nil {
		return modelRoutingProofResult{ExitCode: 1, Err: fmt.Errorf("configure proof containment: %w", err)}
	}
	overflow := make(chan struct{}, 1)
	stdout := newCappedCheckBuffer(overflow)
	stderr := newCappedCheckBuffer(overflow)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return modelRoutingProofResult{ExitCode: 1, Err: err}
	}
	tree, err := attachCheckProcessTree(cmd)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return modelRoutingProofResult{ExitCode: 1, Err: fmt.Errorf("attach proof containment: %w", err)}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	exitReason := ""
	select {
	case err = <-done:
	case <-ctx.Done():
		exitReason = "timeout"
	case <-overflow:
		exitReason = "overflow"
	}
	if exitReason != "" {
		_ = tree.Kill()
		select {
		case err = <-done:
		case <-time.After(processCheckTerminationWait):
			_ = tree.Close()
			return modelRoutingProofResult{ExitCode: 1, Output: combineProofOutput(stdout.String(), stderr.String()), Err: fmt.Errorf("proof process tree did not exit within %s", processCheckTerminationWait), Truncated: exitReason == "overflow"}
		}
	}
	if closeErr := tree.Close(); closeErr != nil && err == nil {
		err = fmt.Errorf("close proof containment: %w", closeErr)
	}
	output := combineProofOutput(stdout.String(), stderr.String())
	if exitReason == "timeout" {
		return modelRoutingProofResult{ExitCode: 1, Output: output, Err: ctx.Err()}
	}
	if exitReason == "overflow" || stdout.truncated || stderr.truncated {
		return modelRoutingProofResult{ExitCode: 1, Output: output, Err: fmt.Errorf("proof output exceeded %d bytes", maxProcessCheckOutputBytes), Truncated: true}
	}
	exitCode := 0
	if err != nil {
		exitCode = 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	return modelRoutingProofResult{ExitCode: exitCode, Output: output, Err: err}
}

func combineProofOutput(stdout, stderr string) string {
	if stdout == "" {
		return stderr
	}
	if stderr == "" {
		return stdout
	}
	return strings.TrimRight(stdout, "\r\n") + "\n" + stderr
}

func modelRoutingDisabledProofRunner(_ context.Context, _ string, command []string) modelRoutingProofResult {
	if len(command) < 2 || command[0] != "go" || command[1] != "test" {
		return modelRoutingProofResult{ExitCode: 1, Err: fmt.Errorf("disabled proof runner received a non-test command")}
	}
	return modelRoutingProofResult{ExitCode: 0, Output: "ok disabled no-spawn deterministic proof"}
}

func resolveAndHashModelRoutingRef(root string, ref modelRoutingFileRef) (string, error) {
	path, err := resolveModelRoutingReleaseFile(root, ref.Path)
	if err != nil {
		return "", err
	}
	if !validReleaseHash(ref.SHA256) {
		return "", fmt.Errorf("invalid sha256")
	}
	return path, nil
}

func resolveModelRoutingReleaseFile(root, input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", fmt.Errorf("path is required")
	}
	rootAbs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", err
	}
	resolved := input
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(rootAbs, filepath.FromSlash(input))
	}
	resolved, err = filepath.Abs(filepath.Clean(resolved))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(rootAbs, resolved)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("path must be repository-relative and contained under the repository root")
	}
	// Inspect the caller-visible path before canonicalization so a symlink inside
	// the repository cannot disappear behind EvalSymlinks. Aliases above the
	// repository root (including Windows 8.3 temp paths) remain permissible.
	probe := rootAbs
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		probe = filepath.Join(probe, part)
		info, statErr := os.Lstat(probe)
		if statErr != nil {
			return "", statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symlink path component is forbidden: %s", part)
		}
	}
	rootCanonical, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", err
	}
	resolvedCanonical, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		return "", err
	}
	canonicalRelative, err := filepath.Rel(rootCanonical, resolvedCanonical)
	if err != nil || canonicalRelative == "." || canonicalRelative == ".." || strings.HasPrefix(canonicalRelative, ".."+string(filepath.Separator)) || filepath.IsAbs(canonicalRelative) {
		return "", fmt.Errorf("path must be repository-relative and contained under the repository root")
	}
	info, err := os.Lstat(resolvedCanonical)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("evidence input must be a regular file")
	}
	if info.Size() > maxModelRoutingReleaseBytes {
		return "", fmt.Errorf("evidence input exceeded %d bytes", maxModelRoutingReleaseBytes)
	}
	return resolvedCanonical, nil
}

func readStrictModelRoutingJSON(path string, value any) error {
	content, err := readSafeBoundedModelRoutingFile(path)
	if err != nil {
		return err
	}
	return decodeStrictModelRoutingJSON(path, content, value)
}

func readStrictHashedModelRoutingJSON(path, expected string, value any) error {
	if !validReleaseHash(expected) {
		return fmt.Errorf("invalid sha256")
	}
	content, err := readSafeBoundedModelRoutingFile(path)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(content)
	actual := hex.EncodeToString(sum[:])
	if actual != expected {
		return fmt.Errorf("hash mismatch: expected %s got %s", expected, actual)
	}
	return decodeStrictModelRoutingJSON(path, content, value)
}

func readSafeBoundedModelRoutingFile(path string) ([]byte, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() {
		return nil, fmt.Errorf("input must be a non-symlink regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return nil, err
	}
	content, err := io.ReadAll(io.LimitReader(file, maxModelRoutingReleaseBytes+1))
	if err != nil {
		return nil, err
	}
	after, err := os.Lstat(path)
	if err != nil || after.Mode()&os.ModeSymlink != 0 || !os.SameFile(before, opened) || !os.SameFile(opened, after) {
		return nil, fmt.Errorf("input changed or became a symlink while being read")
	}
	if int64(len(content)) > maxModelRoutingReleaseBytes || opened.Size() > maxModelRoutingReleaseBytes {
		return nil, fmt.Errorf("%s exceeded %d bytes", filepath.Base(path), maxModelRoutingReleaseBytes)
	}
	return content, nil
}

func decodeStrictModelRoutingJSON(path string, content []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("%s contained trailing JSON content", filepath.Base(path))
	}
	return nil
}

func validReleaseHash(value string) bool {
	if len(value) != 64 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
