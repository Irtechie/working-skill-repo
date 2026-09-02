package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	modelTierEvalSchemaVersion = 1
	modelTierEvalThreshold     = "medium-v1"
	maxModelTierEvidenceBytes  = 1 << 20
)

var modelTierEvalIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{2,127}$`)

type modelTierEvidence struct {
	SchemaVersion                     int                  `json:"schema_version"`
	TargetTier                        string               `json:"target_tier"`
	ThresholdRevision                 string               `json:"threshold_revision"`
	EvidenceDate                      string               `json:"evidence_date"`
	Scope                             modelTierScope       `json:"scope"`
	Cohort                            modelTierCohort      `json:"cohort"`
	ExecutionFingerprint              modelTierFingerprint `json:"execution_fingerprint"`
	CurrentExecutionFingerprintSHA256 string               `json:"current_execution_fingerprint_sha256"`
	Attempts                          []modelTierAttempt   `json:"attempts"`
}

type modelTierScope struct {
	TaskFamilies        []string `json:"task_families"`
	Tools               []string `json:"tools"`
	ContextRiskEnvelope string   `json:"context_risk_envelope"`
	RouteRevision       string   `json:"route_revision"`
}

type modelTierCohort struct {
	ID                 string            `json:"id"`
	ManifestPath       string            `json:"manifest_path"`
	ManifestSHA256     string            `json:"manifest_sha256"`
	ManifestSignerID   string            `json:"manifest_signer_id"`
	ManifestSignature  string            `json:"manifest_signature_ed25519"`
	PreregisteredAt    string            `json:"preregistered_at"`
	FirstAttemptAt     string            `json:"first_attempt_at"`
	ProvenanceKind     string            `json:"provenance_kind"`
	ExpectedAttemptIDs []string          `json:"expected_attempt_ids"`
	Families           []modelTierFamily `json:"families"`
}

type modelTierFamily struct {
	ID         string                    `json:"id"`
	Holdout    bool                      `json:"holdout"`
	Dimensions modelTierFamilyDimensions `json:"dimensions"`
}

type modelTierCohortManifest struct {
	SchemaVersion              int               `json:"schema_version"`
	ID                         string            `json:"id"`
	TargetTier                 string            `json:"target_tier"`
	ThresholdRevision          string            `json:"threshold_revision"`
	Scope                      modelTierScope    `json:"scope"`
	PreregisteredAt            string            `json:"preregistered_at"`
	FirstAttemptAt             string            `json:"first_attempt_at"`
	ProvenanceKind             string            `json:"provenance_kind"`
	ExpectedAttemptIDs         []string          `json:"expected_attempt_ids"`
	Families                   []modelTierFamily `json:"families"`
	ExecutionFingerprintSHA256 string            `json:"execution_fingerprint_sha256"`
}

type modelTierTrustPolicy struct {
	SchemaVersion int                      `json:"schema_version"`
	Signers       []modelTierTrustedSigner `json:"signers"`
}

type modelTierTrustedSigner struct {
	ID               string `json:"id"`
	PublicKeyEd25519 string `json:"public_key_ed25519"`
}

type modelTierFamilyDimensions struct {
	TaskStructure       string `json:"task_structure"`
	RepositoryMechanism string `json:"repository_mechanism"`
	OracleType          string `json:"oracle_type"`
	PrimaryFailureMode  string `json:"primary_failure_mode"`
}

type modelTierFingerprint struct {
	RouteFingerprint       string `json:"route_fingerprint"`
	ProviderModelRevision  string `json:"provider_model_revision"`
	RouteConfigRevision    string `json:"route_config_revision"`
	SystemInstructionsHash string `json:"system_instructions_hash"`
	ToolsHash              string `json:"tools_hash"`
	ContextRiskPolicyHash  string `json:"context_risk_policy_hash"`
	PlannerRevision        string `json:"planner_revision"`
	ScorerRevision         string `json:"scorer_revision"`
	OracleRevision         string `json:"oracle_revision"`
}

type modelTierAttempt struct {
	ID                         string                  `json:"id"`
	FixtureID                  string                  `json:"fixture_id"`
	FixtureRevision            string                  `json:"fixture_revision"`
	FamilyID                   string                  `json:"family_id"`
	TargetTier                 string                  `json:"target_tier"`
	ExecutionFingerprintSHA256 string                  `json:"execution_fingerprint_sha256"`
	PlanSHA256                 string                  `json:"plan_sha256"`
	RequestSHA256              string                  `json:"request_sha256"`
	ResponseSHA256             string                  `json:"response_sha256"`
	ProofSHA256                string                  `json:"proof_sha256"`
	AttemptedAt                string                  `json:"attempted_at"`
	Outcome                    string                  `json:"outcome"`
	FailureKind                string                  `json:"failure_kind"`
	ResponseProduced           bool                    `json:"response_produced"`
	OneAttemptIdentityValid    bool                    `json:"one_attempt_identity_valid"`
	ProofBound                 bool                    `json:"proof_bound"`
	Trust                      modelTierAttemptTrust   `json:"trust"`
	ArtifactPaths              modelTierArtifactPaths  `json:"artifact_paths"`
	ArtifactHashes             modelTierArtifactHashes `json:"artifact_hashes"`
}

type modelTierAttemptTrust struct {
	Plan      string `json:"plan"`
	Route     string `json:"route"`
	Execution string `json:"execution"`
	Oracle    string `json:"oracle"`
	Proof     string `json:"proof"`
}

type modelTierArtifactHashes struct {
	CohortManifest      string `json:"cohort_manifest"`
	Fixture             string `json:"fixture"`
	FrozenPlan          string `json:"frozen_plan"`
	PlanSufficiency     string `json:"plan_sufficiency"`
	ExecutorPackage     string `json:"executor_package"`
	RouteExecution      string `json:"route_execution"`
	OracleQualification string `json:"oracle_qualification"`
	ProofResult         string `json:"proof_result"`
}

type modelTierArtifactPaths struct {
	CohortManifest      string `json:"cohort_manifest"`
	Fixture             string `json:"fixture"`
	FrozenPlan          string `json:"frozen_plan"`
	Request             string `json:"request"`
	Response            string `json:"response"`
	PlanSufficiency     string `json:"plan_sufficiency"`
	ExecutorPackage     string `json:"executor_package"`
	RouteExecution      string `json:"route_execution"`
	OracleQualification string `json:"oracle_qualification"`
	ProofResult         string `json:"proof_result"`
}

type modelTierEvalResult struct {
	SchemaVersion              int            `json:"schema_version"`
	Experimental               bool           `json:"experimental"`
	RoutingPromotion           bool           `json:"routing_promotion"`
	Decision                   string         `json:"decision"`
	TargetTier                 string         `json:"target_tier"`
	Admitted                   int            `json:"admitted"`
	Passes                     int            `json:"passes"`
	ModelFailures              int            `json:"model_failures"`
	Exclusions                 map[string]int `json:"exclusions"`
	FixtureFamilies            int            `json:"fixture_families"`
	IndependentFixtureFamilies int            `json:"independent_fixture_families"`
	HoldoutCovered             bool           `json:"holdout_covered"`
	MaxFamilyShare             float64        `json:"max_family_share"`
	FailureRateUpperBound95    float64        `json:"failure_rate_upper_bound_95,omitempty"`
	Trust                      map[string]int `json:"trust"`
	Reasons                    []string       `json:"reasons"`
	Scope                      modelTierScope `json:"scope"`
	EvidenceDate               string         `json:"evidence_date"`
	ExecutionFingerprintSHA256 string         `json:"execution_fingerprint_sha256"`
}

func runModelTierEvalCommand(root string, opts options, stdout, stderr io.Writer) int {
	path, err := resolveModelTierFile(root, opts.evidencePath)
	if err != nil {
		fmt.Fprintln(stderr, "model-tier-eval: invalid evidence path")
		return 2
	}
	content, err := readSafeBoundedModelTierFile(path)
	if err != nil {
		fmt.Fprintln(stderr, "model-tier-eval: evidence input is unavailable or unsafe")
		return 2
	}
	if err := rejectSensitiveModelTierEvidence(content); err != nil {
		fmt.Fprintln(stderr, "model-tier-eval: evidence contains a forbidden sensitive value")
		return 2
	}
	if err := validateJSONShape(content, 12); err != nil {
		fmt.Fprintln(stderr, "model-tier-eval: invalid strict JSON")
		return 2
	}
	var evidence modelTierEvidence
	if err := decodeStrictModelTierJSON(path, content, &evidence); err != nil {
		fmt.Fprintln(stderr, "model-tier-eval: invalid strict evidence document")
		return 2
	}
	result, err := evaluateModelTierEvidence(root, evidence, time.Now().UTC())
	if err != nil {
		fmt.Fprintln(stderr, "model-tier-eval:", err)
		return 2
	}
	if opts.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(result)
		return 0
	}
	fmt.Fprintf(stdout, "%s: %s tier, %d admitted (%d pass, %d model failure); experimental evidence only; does not promote routing\n",
		result.Decision, result.TargetTier, result.Admitted, result.Passes, result.ModelFailures)
	if len(result.Reasons) > 0 {
		fmt.Fprintln(stdout, strings.Join(result.Reasons, "; "))
	}
	return 0
}

func evaluateModelTierEvidence(root string, evidence modelTierEvidence, now time.Time) (modelTierEvalResult, error) {
	if evidence.SchemaVersion != modelTierEvalSchemaVersion {
		return modelTierEvalResult{}, fmt.Errorf("unsupported schema_version")
	}
	if evidence.TargetTier != "medium" {
		return modelTierEvalResult{}, fmt.Errorf("only the frozen Medium policy is supported")
	}
	if evidence.ThresholdRevision != modelTierEvalThreshold {
		return modelTierEvalResult{}, fmt.Errorf("unsupported threshold_revision")
	}
	evidenceDate, err := time.Parse(time.RFC3339, evidence.EvidenceDate)
	if err != nil {
		return modelTierEvalResult{}, fmt.Errorf("evidence_date must be RFC3339")
	}
	if err := validateModelTierScope(evidence.Scope); err != nil {
		return modelTierEvalResult{}, err
	}
	fingerprintHash, err := hashModelTierFingerprint(evidence.ExecutionFingerprint)
	if err != nil {
		return modelTierEvalResult{}, err
	}
	if err := validateModelTierFingerprint(evidence.ExecutionFingerprint); err != nil {
		return modelTierEvalResult{}, err
	}
	if err := validateModelTierCohort(evidence.Cohort); err != nil {
		return modelTierEvalResult{}, err
	}
	cohortManifest, authorityVerified, err := loadModelTierCohortManifest(root, evidence.Cohort)
	if err != nil {
		return modelTierEvalResult{}, err
	}
	if cohortManifest.ID != evidence.Cohort.ID ||
		cohortManifest.TargetTier != evidence.TargetTier ||
		cohortManifest.ThresholdRevision != evidence.ThresholdRevision ||
		cohortManifest.PreregisteredAt != evidence.Cohort.PreregisteredAt ||
		cohortManifest.FirstAttemptAt != evidence.Cohort.FirstAttemptAt ||
		cohortManifest.ProvenanceKind != evidence.Cohort.ProvenanceKind ||
		cohortManifest.ExecutionFingerprintSHA256 != fingerprintHash ||
		!reflect.DeepEqual(cohortManifest.Scope, evidence.Scope) ||
		!reflect.DeepEqual(cohortManifest.ExpectedAttemptIDs, evidence.Cohort.ExpectedAttemptIDs) ||
		!reflect.DeepEqual(cohortManifest.Families, evidence.Cohort.Families) {
		return modelTierEvalResult{}, fmt.Errorf("evidence does not match the verified cohort manifest")
	}
	if len(evidence.Attempts) > 500 {
		return modelTierEvalResult{}, fmt.Errorf("attempt count exceeds 500")
	}

	expected := modelTierStringSet(evidence.Cohort.ExpectedAttemptIDs)
	if len(expected) != len(evidence.Cohort.ExpectedAttemptIDs) {
		return modelTierEvalResult{}, fmt.Errorf("duplicate expected attempt id")
	}
	families := map[string]modelTierFamily{}
	unsupportedAuthority := !authorityVerified
	for _, family := range evidence.Cohort.Families {
		if !modelTierEvalIDPattern.MatchString(family.ID) || families[family.ID].ID != "" {
			return modelTierEvalResult{}, fmt.Errorf("invalid or duplicate family id")
		}
		if !validFamilyDimensions(family.Dimensions) {
			return modelTierEvalResult{}, fmt.Errorf("family dimensions must be complete")
		}
		families[family.ID] = family
	}

	exclusions := map[string]int{}
	trustCounts := map[string]int{"reproduced": 0, "unsupported": 0}
	attemptIDs := map[string]bool{}
	fixtureIDs := map[string]bool{}
	familyCounts := map[string]int{}
	passes := 0
	failures := 0
	for _, attempt := range evidence.Attempts {
		if !modelTierEvalIDPattern.MatchString(attempt.ID) || attemptIDs[attempt.ID] {
			return modelTierEvalResult{}, fmt.Errorf("invalid, replayed, or conflicting attempt id")
		}
		attemptIDs[attempt.ID] = true
		if !modelTierEvalIDPattern.MatchString(attempt.FixtureID) || fixtureIDs[attempt.FixtureID] {
			return modelTierEvalResult{}, fmt.Errorf("invalid or overlapping fixture identity")
		}
		fixtureIDs[attempt.FixtureID] = true
		if !expected[attempt.ID] {
			return modelTierEvalResult{}, fmt.Errorf("undeclared attempt id")
		}
		if attempt.TargetTier != evidence.TargetTier || attempt.ExecutionFingerprintSHA256 != fingerprintHash {
			return modelTierEvalResult{}, fmt.Errorf("mixed target tier or execution fingerprint")
		}
		if families[attempt.FamilyID].ID == "" {
			return modelTierEvalResult{}, fmt.Errorf("attempt references unknown family")
		}
		if _, err := time.Parse(time.RFC3339, attempt.AttemptedAt); err != nil {
			return modelTierEvalResult{}, fmt.Errorf("attempted_at must be RFC3339")
		}
		if err := validateModelTierAttemptHashes(attempt); err != nil {
			return modelTierEvalResult{}, err
		}
		if err := verifyModelTierAttemptArtifacts(root, evidence.Cohort, attempt); err != nil {
			return modelTierEvalResult{}, err
		}
		for _, status := range []string{attempt.Trust.Plan, attempt.Trust.Route, attempt.Trust.Execution, attempt.Trust.Oracle, attempt.Trust.Proof} {
			if status != "reproduced" && status != "unsupported" {
				return modelTierEvalResult{}, fmt.Errorf("invalid receipt trust status")
			}
			trustCounts[status]++
			if status == "unsupported" {
				unsupportedAuthority = true
			}
		}
		switch attempt.Outcome {
		case "model-pass":
			if attempt.FailureKind != "" || !admissibleModelAttempt(attempt) {
				return modelTierEvalResult{}, fmt.Errorf("model-pass attempt is not admissible")
			}
			passes++
			familyCounts[attempt.FamilyID]++
		case "model-failure":
			if !oneOf(attempt.FailureKind, "proof", "output-schema") || !admissibleModelAttempt(attempt) {
				return modelTierEvalResult{}, fmt.Errorf("model-failure requires proof or output-schema failure after valid execution")
			}
			failures++
			familyCounts[attempt.FamilyID]++
		case "plan-insufficient", "oracle-invalid", "preflight-schema", "route-infrastructure-not-run",
			"execution-failed-no-response", "execution-indeterminate", "proof-infrastructure":
			exclusions[attempt.Outcome]++
		default:
			return modelTierEvalResult{}, fmt.Errorf("unsupported attempt outcome")
		}
	}
	if len(attemptIDs) != len(expected) {
		return modelTierEvalResult{}, fmt.Errorf("cohort ledger is incomplete")
	}

	admitted := passes + failures
	independentFamilies := countIndependentFamilies(families, familyCounts)
	holdoutCovered := false
	maxFamilyCount := 0
	for id, count := range familyCounts {
		if count > maxFamilyCount {
			maxFamilyCount = count
		}
		if count > 0 && families[id].Holdout {
			holdoutCovered = true
		}
	}
	maxFamilyShare := 0.0
	if admitted > 0 {
		maxFamilyShare = float64(maxFamilyCount) / float64(admitted)
	}
	stale := evidence.CurrentExecutionFingerprintSHA256 != fingerprintHash
	if evidence.ExecutionFingerprint.ProviderModelRevision == "" && now.Sub(evidenceDate) > 30*24*time.Hour {
		stale = true
	}
	result := modelTierEvalResult{
		SchemaVersion:              1,
		Experimental:               true,
		RoutingPromotion:           false,
		TargetTier:                 evidence.TargetTier,
		Admitted:                   admitted,
		Passes:                     passes,
		ModelFailures:              failures,
		Exclusions:                 exclusions,
		FixtureFamilies:            len(familyCounts),
		IndependentFixtureFamilies: independentFamilies,
		HoldoutCovered:             holdoutCovered,
		MaxFamilyShare:             maxFamilyShare,
		Trust:                      trustCounts,
		Scope:                      evidence.Scope,
		EvidenceDate:               evidence.EvidenceDate,
		ExecutionFingerprintSHA256: fingerprintHash,
	}
	if failures > 0 {
		result.Decision = "not-qualified"
		result.Reasons = []string{"at least one admitted model failure"}
		return result, nil
	}
	if admitted > 0 {
		result.FailureRateUpperBound95 = 1 - math.Pow(0.05, 1/float64(admitted))
	}
	reasons := []string{}
	if admitted < 30 {
		reasons = append(reasons, "fewer than 30 admitted unique fixtures")
	}
	if independentFamilies < 5 {
		reasons = append(reasons, "fewer than five materially independent fixture families")
	}
	if maxFamilyShare > 0.20 {
		reasons = append(reasons, "one fixture family exceeds 20% of admitted evidence")
	}
	if !holdoutCovered {
		reasons = append(reasons, "no admitted holdout family")
	}
	if unsupportedAuthority {
		reasons = append(reasons, "one or more evidence authorities are unsupported")
	}
	if stale {
		reasons = append(reasons, "evidence is stale or does not match the current execution fingerprint")
	}
	if len(exclusions) > 0 {
		reasons = append(reasons, "excluded non-model outcomes leave the cohort underpowered")
	}
	if len(reasons) > 0 {
		result.Decision = "inconclusive"
		result.Reasons = reasons
		return result, nil
	}
	result.Decision = "qualified"
	result.Reasons = []string{"frozen Medium threshold satisfied with zero admitted model failures"}
	return result, nil
}

func validateModelTierScope(scope modelTierScope) error {
	if len(scope.TaskFamilies) == 0 || len(scope.TaskFamilies) > 50 || len(scope.Tools) > 50 ||
		strings.TrimSpace(scope.ContextRiskEnvelope) == "" || strings.TrimSpace(scope.RouteRevision) == "" {
		return fmt.Errorf("tested scope is incomplete or unbounded")
	}
	for _, value := range append(append([]string{}, scope.TaskFamilies...), scope.Tools...) {
		if !boundedModelTierString(value) {
			return fmt.Errorf("tested scope contains an invalid value")
		}
	}
	return nil
}

func validateModelTierCohort(cohort modelTierCohort) error {
	if !modelTierEvalIDPattern.MatchString(cohort.ID) || strings.TrimSpace(cohort.ManifestPath) == "" ||
		!validModelTierHash(cohort.ManifestSHA256) || !modelTierEvalIDPattern.MatchString(cohort.ManifestSignerID) ||
		len(cohort.ManifestSignature) != ed25519.SignatureSize*2 ||
		len(cohort.ExpectedAttemptIDs) == 0 || len(cohort.Families) == 0 {
		return fmt.Errorf("cohort manifest is incomplete")
	}
	preregistered, err := time.Parse(time.RFC3339, cohort.PreregisteredAt)
	if err != nil {
		return fmt.Errorf("preregistered_at must be RFC3339")
	}
	firstAttempt, err := time.Parse(time.RFC3339, cohort.FirstAttemptAt)
	if err != nil || !preregistered.Before(firstAttempt) {
		return fmt.Errorf("cohort must be preregistered before the first attempt")
	}
	if cohort.ProvenanceKind != "ed25519-signature" {
		return fmt.Errorf("unsupported preregistration provenance")
	}
	return nil
}

func loadModelTierCohortManifest(root string, cohort modelTierCohort) (modelTierCohortManifest, bool, error) {
	path, err := resolveModelTierFile(root, cohort.ManifestPath)
	if err != nil {
		return modelTierCohortManifest{}, false, fmt.Errorf("cohort manifest path is unsafe")
	}
	content, err := readSafeBoundedModelTierFile(path)
	if err != nil {
		return modelTierCohortManifest{}, false, fmt.Errorf("cohort manifest is unavailable or unsafe")
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(content)); got != strings.ToLower(cohort.ManifestSHA256) {
		return modelTierCohortManifest{}, false, fmt.Errorf("cohort manifest sha256 mismatch")
	}
	if err := validateJSONShape(content, 8); err != nil {
		return modelTierCohortManifest{}, false, fmt.Errorf("cohort manifest is not strict JSON")
	}
	var manifest modelTierCohortManifest
	if err := decodeStrictModelTierJSON(path, content, &manifest); err != nil {
		return modelTierCohortManifest{}, false, fmt.Errorf("cohort manifest is invalid")
	}
	if manifest.SchemaVersion != 1 {
		return modelTierCohortManifest{}, false, fmt.Errorf("unsupported cohort manifest schema")
	}
	return manifest, verifyModelTierManifestAuthority(root, cohort, content), nil
}

func verifyModelTierManifestAuthority(root string, cohort modelTierCohort, manifest []byte) bool {
	path, err := resolveModelTierFile(root, filepath.Join("config", "model-tier-trust.json"))
	if err != nil {
		return false
	}
	content, err := readSafeBoundedModelTierFile(path)
	if err != nil || validateJSONShape(content, 6) != nil {
		return false
	}
	var policy modelTierTrustPolicy
	if decodeStrictModelTierJSON(path, content, &policy) != nil || policy.SchemaVersion != 1 {
		return false
	}
	signature, err := hex.DecodeString(cohort.ManifestSignature)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return false
	}
	for _, signer := range policy.Signers {
		if signer.ID != cohort.ManifestSignerID {
			continue
		}
		publicKey, err := hex.DecodeString(signer.PublicKeyEd25519)
		if err != nil || len(publicKey) != ed25519.PublicKeySize {
			return false
		}
		return ed25519.Verify(ed25519.PublicKey(publicKey), manifest, signature)
	}
	return false
}

func validateModelTierFingerprint(fingerprint modelTierFingerprint) error {
	for _, hash := range []string{
		fingerprint.RouteFingerprint, fingerprint.SystemInstructionsHash, fingerprint.ToolsHash, fingerprint.ContextRiskPolicyHash,
	} {
		if !validModelTierHash(hash) {
			return fmt.Errorf("execution fingerprint contains an invalid hash")
		}
	}
	for _, value := range []string{
		fingerprint.RouteConfigRevision, fingerprint.PlannerRevision, fingerprint.ScorerRevision, fingerprint.OracleRevision,
	} {
		if !boundedModelTierString(value) {
			return fmt.Errorf("execution fingerprint contains an invalid revision")
		}
	}
	return nil
}

func validateModelTierAttemptHashes(attempt modelTierAttempt) error {
	hashes := []string{
		attempt.ExecutionFingerprintSHA256, attempt.PlanSHA256, attempt.RequestSHA256,
		attempt.ArtifactHashes.CohortManifest, attempt.ArtifactHashes.Fixture, attempt.ArtifactHashes.FrozenPlan,
		attempt.ArtifactHashes.PlanSufficiency, attempt.ArtifactHashes.ExecutorPackage, attempt.ArtifactHashes.RouteExecution,
		attempt.ArtifactHashes.OracleQualification, attempt.ArtifactHashes.ProofResult,
	}

	if attempt.ResponseProduced {
		hashes = append(hashes, attempt.ResponseSHA256)
	}
	if attempt.ProofBound {
		hashes = append(hashes, attempt.ProofSHA256)
	}
	for _, hash := range hashes {
		if !validModelTierHash(hash) {
			return fmt.Errorf("attempt artifact hash is missing or invalid")
		}
	}
	return nil
}

func verifyModelTierAttemptArtifacts(root string, cohort modelTierCohort, attempt modelTierAttempt) error {
	if filepath.ToSlash(attempt.ArtifactPaths.CohortManifest) != filepath.ToSlash(cohort.ManifestPath) ||
		attempt.ArtifactHashes.CohortManifest != cohort.ManifestSHA256 {
		return fmt.Errorf("attempt cohort manifest binding mismatch")
	}
	bindings := []struct {
		label string
		path  string
		hash  string
	}{
		{"cohort manifest", attempt.ArtifactPaths.CohortManifest, attempt.ArtifactHashes.CohortManifest},
		{"fixture", attempt.ArtifactPaths.Fixture, attempt.ArtifactHashes.Fixture},
		{"frozen plan", attempt.ArtifactPaths.FrozenPlan, attempt.ArtifactHashes.FrozenPlan},
		{"request", attempt.ArtifactPaths.Request, attempt.RequestSHA256},
		{"plan sufficiency", attempt.ArtifactPaths.PlanSufficiency, attempt.ArtifactHashes.PlanSufficiency},
		{"executor package", attempt.ArtifactPaths.ExecutorPackage, attempt.ArtifactHashes.ExecutorPackage},
		{"route execution", attempt.ArtifactPaths.RouteExecution, attempt.ArtifactHashes.RouteExecution},
		{"oracle qualification", attempt.ArtifactPaths.OracleQualification, attempt.ArtifactHashes.OracleQualification},
		{"proof result", attempt.ArtifactPaths.ProofResult, attempt.ArtifactHashes.ProofResult},
	}
	if attempt.ResponseProduced {
		bindings = append(bindings, struct {
			label string
			path  string
			hash  string
		}{"response", attempt.ArtifactPaths.Response, attempt.ResponseSHA256})
	}
	for _, binding := range bindings {
		if err := verifyModelTierArtifact(root, binding.label, binding.path, binding.hash); err != nil {
			return err
		}
	}
	if attempt.PlanSHA256 != attempt.ArtifactHashes.FrozenPlan ||
		(attempt.ProofBound && attempt.ProofSHA256 != attempt.ArtifactHashes.ProofResult) {
		return fmt.Errorf("attempt top-level artifact binding mismatch")
	}
	return nil
}

func verifyModelTierArtifact(root, label, relative, want string) error {
	path, err := resolveModelTierFile(root, relative)
	if err != nil {
		return fmt.Errorf("%s path is unsafe", label)
	}
	content, err := readSafeBoundedModelTierFile(path)
	if err != nil {
		return fmt.Errorf("%s artifact is unavailable or unsafe", label)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(content)); got != strings.ToLower(want) {
		return fmt.Errorf("%s artifact sha256 mismatch", label)
	}
	return nil
}

func resolveModelTierFile(root, input string) (string, error) {
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
	if info.Size() > maxModelTierEvidenceBytes {
		return "", fmt.Errorf("evidence input exceeded %d bytes", maxModelTierEvidenceBytes)
	}
	return resolvedCanonical, nil
}

func readSafeBoundedModelTierFile(path string) ([]byte, error) {
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
	content, err := io.ReadAll(io.LimitReader(file, maxModelTierEvidenceBytes+1))
	if err != nil {
		return nil, err
	}
	after, err := os.Lstat(path)
	if err != nil || after.Mode()&os.ModeSymlink != 0 || !os.SameFile(before, opened) || !os.SameFile(opened, after) {
		return nil, fmt.Errorf("input changed or became a symlink while being read")
	}
	if int64(len(content)) > maxModelTierEvidenceBytes || opened.Size() > maxModelTierEvidenceBytes {
		return nil, fmt.Errorf("%s exceeded %d bytes", filepath.Base(path), maxModelTierEvidenceBytes)
	}
	return content, nil
}

func decodeStrictModelTierJSON(path string, content []byte, value any) error {
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

func validModelTierHash(value string) bool {
	if len(value) != 64 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func admissibleModelAttempt(attempt modelTierAttempt) bool {
	return attempt.ResponseProduced && attempt.OneAttemptIdentityValid && attempt.ProofBound
}

func hashModelTierFingerprint(fingerprint modelTierFingerprint) (string, error) {
	content, err := json.Marshal(fingerprint)
	if err != nil {
		return "", err
	}
	var canonical map[string]any
	if err := json.Unmarshal(content, &canonical); err != nil {
		return "", err
	}
	content, err = json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func countIndependentFamilies(families map[string]modelTierFamily, admitted map[string]int) int {
	ids := make([]string, 0, len(admitted))
	for id, count := range admitted {
		if count > 0 {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	independent := []modelTierFamily{}
	for _, id := range ids {
		candidate := families[id]
		valid := true
		for _, accepted := range independent {
			if familyDimensionDifferences(candidate.Dimensions, accepted.Dimensions) < 3 {
				valid = false
				break
			}
		}
		if valid {
			independent = append(independent, candidate)
		}
	}
	return len(independent)
}

func familyDimensionDifferences(left, right modelTierFamilyDimensions) int {
	differences := 0
	if left.TaskStructure != right.TaskStructure {
		differences++
	}
	if left.RepositoryMechanism != right.RepositoryMechanism {
		differences++
	}
	if left.OracleType != right.OracleType {
		differences++
	}
	if left.PrimaryFailureMode != right.PrimaryFailureMode {
		differences++
	}
	return differences
}

func validFamilyDimensions(dimensions modelTierFamilyDimensions) bool {
	return boundedModelTierString(dimensions.TaskStructure) &&
		boundedModelTierString(dimensions.RepositoryMechanism) &&
		boundedModelTierString(dimensions.OracleType) &&
		boundedModelTierString(dimensions.PrimaryFailureMode)
}

func boundedModelTierString(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && len(value) <= 256 && !strings.ContainsAny(value, "\r\n")
}

func modelTierStringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func rejectSensitiveModelTierEvidence(content []byte) error {
	lower := strings.ToLower(string(content))
	for _, forbidden := range []string{
		"http://", "https://", `"api_key"`, `"password"`, `"secret"`, `"endpoint"`,
		`"prompt"`, `"response_body"`, `"protected_test"`, `"solution"`, "authorization", "bearer ",
	} {
		if strings.Contains(lower, forbidden) {
			return fmt.Errorf("forbidden sensitive field or value")
		}
	}
	return nil
}

func validateJSONShape(content []byte, maxDepth int) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if err := validateJSONToken(decoder, token, 1, maxDepth); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("trailing JSON content")
	}
	return nil
}

func validateJSONToken(decoder *json.Decoder, token json.Token, depth, maxDepth int) error {
	if depth > maxDepth {
		return fmt.Errorf("JSON nesting exceeds %d", maxDepth)
	}
	delim, ok := token.(json.Delim)
	if !ok {
		if value, ok := token.(string); ok && len(value) > 1024 {
			return fmt.Errorf("JSON string exceeds 1024 bytes")
		}
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok || seen[key] {
				return fmt.Errorf("duplicate or invalid JSON key")
			}
			seen[key] = true
			valueToken, err := decoder.Token()
			if err != nil {
				return err
			}
			if err := validateJSONToken(decoder, valueToken, depth+1, maxDepth); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	case '[':
		for decoder.More() {
			valueToken, err := decoder.Token()
			if err != nil {
				return err
			}
			if err := validateJSONToken(decoder, valueToken, depth+1, maxDepth); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	default:
		return fmt.Errorf("invalid JSON delimiter")
	}
}
