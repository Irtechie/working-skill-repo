package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"
)

type graphRoutingEvalResult struct {
	OK                  bool     `json:"ok"`
	Ready               bool     `json:"ready"`
	ImpactRecall        float64  `json:"impact_recall"`
	FalsePositivePerTok float64  `json:"false_positive_per_token"`
	SafetyFailures      []string `json:"safety_failures,omitempty"`
	CorrectnessFailures []string `json:"correctness_failures,omitempty"`
	Skipped             []string `json:"skipped,omitempty"`
}

type graphRoutingExpectedResults struct {
	SchemaVersion int                      `json:"schema_version"`
	Promotion     graphRoutingPromotion    `json:"promotion"`
	Fixtures      []graphRoutingFixtureRef `json:"fixtures"`
}

type graphRoutingPromotion struct {
	MinImpactRecall             float64 `json:"min_impact_recall"`
	MaxFalsePositivePerToken    float64 `json:"max_false_positive_per_token"`
	RequireZeroSafetyFailures   bool    `json:"require_zero_safety_failures"`
	RequireOptionalProviderFree bool    `json:"require_optional_provider_free"`
}

type graphRoutingFixtureRef struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Path string `json:"path"`
}

type symbolImpactFixture struct {
	SchemaVersion            int      `json:"schema_version"`
	ExpectedImpact           []string `json:"expected_impact"`
	Retrieved                []string `json:"retrieved"`
	RequiredTests            []string `json:"required_tests"`
	RequiredDocs             []string `json:"required_docs"`
	RetrievedTokens          int      `json:"retrieved_tokens"`
	UncitedExactEdges        []string `json:"uncited_exact_edges"`
	HiddenDynamicLimitations bool     `json:"hidden_dynamic_limitations"`
}

type staleIndexFixture struct {
	SchemaVersion           int    `json:"schema_version"`
	IndexRevision           string `json:"index_revision"`
	WorktreeRevision        string `json:"worktree_revision"`
	DirtyFingerprint        string `json:"dirty_fingerprint"`
	AcceptedAsAuthoritative bool   `json:"accepted_as_authoritative"`
	FallbackMode            string `json:"fallback_mode"`
	OptionalProviderStatus  string `json:"optional_provider_status"`
}

type multiSessionRaceFixture struct {
	SchemaVersion               int               `json:"schema_version"`
	SameSliceClaims             []raceClaim       `json:"same_slice_claims"`
	DisjointWorktreesIntegrated bool              `json:"disjoint_worktrees_integrated_serially"`
	DirtyCheckoutPreserved      bool              `json:"dirty_checkout_preserved"`
	ForceCleanupUsed            bool              `json:"force_cleanup_used"`
	PrefixCollisions            []prefixCollision `json:"prefix_collisions"`
	OptionalProviderStatus      string            `json:"optional_provider_status"`
}

type raceClaim struct {
	Owner    string `json:"owner"`
	Acquired bool   `json:"acquired"`
}

type prefixCollision struct {
	First    string `json:"first"`
	Second   string `json:"second"`
	Rejected bool   `json:"rejected"`
}

func runGraphRoutingEvalCommand(root string, opts options, stdout, stderr io.Writer) int {
	result, err := evaluateGraphRouting(root)
	if err != nil {
		fmt.Fprintf(stderr, "graph-routing-eval: %v\n", err)
		return 1
	}
	if opts.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(result)
	} else if result.OK {
		fmt.Fprintf(stdout, "graph-routing-eval: ok ready=%t recall=%.2f fp_per_token=%.4f skipped=%d\n", result.Ready, result.ImpactRecall, result.FalsePositivePerTok, len(result.Skipped))
	} else {
		for _, failure := range append(result.CorrectnessFailures, result.SafetyFailures...) {
			fmt.Fprintln(stderr, failure)
		}
	}
	if opts.requireReady && !result.Ready {
		return 2
	}
	if !result.OK {
		return 2
	}
	return 0
}

func evaluateGraphRouting(root string) (graphRoutingEvalResult, error) {
	expectedPath := filepath.Join(root, "evals", "graph-routing", "expected-results.json")
	var expected graphRoutingExpectedResults
	if err := readJSONFile(expectedPath, &expected); err != nil {
		return graphRoutingEvalResult{}, err
	}
	if expected.SchemaVersion != 1 {
		return graphRoutingEvalResult{}, fmt.Errorf("expected-results schema_version must be 1")
	}
	result := graphRoutingEvalResult{OK: true, Ready: true}
	totalExpected := 0
	totalHit := 0
	totalFalsePositive := 0
	totalTokens := 0
	for _, fixture := range expected.Fixtures {
		path := filepath.Join(root, filepath.FromSlash(fixture.Path))
		switch fixture.Kind {
		case "symbol-impact":
			var data symbolImpactFixture
			if err := readJSONFile(path, &data); err != nil {
				return graphRoutingEvalResult{}, err
			}
			hit, total, fp, tokens, failures := scoreSymbolImpact(data)
			totalHit += hit
			totalExpected += total
			totalFalsePositive += fp
			totalTokens += tokens
			result.CorrectnessFailures = append(result.CorrectnessFailures, prefixFailures(fixture.ID, failures)...)
		case "stale-index":
			var data staleIndexFixture
			if err := readJSONFile(path, &data); err != nil {
				return graphRoutingEvalResult{}, err
			}
			failures, skipped := scoreStaleIndex(data)
			result.SafetyFailures = append(result.SafetyFailures, prefixFailures(fixture.ID, failures)...)
			result.Skipped = append(result.Skipped, skipped...)
		case "multisession-race":
			var data multiSessionRaceFixture
			if err := readJSONFile(path, &data); err != nil {
				return graphRoutingEvalResult{}, err
			}
			failures, skipped := scoreMultiSessionRace(data)
			result.SafetyFailures = append(result.SafetyFailures, prefixFailures(fixture.ID, failures)...)
			result.Skipped = append(result.Skipped, skipped...)
		default:
			return graphRoutingEvalResult{}, fmt.Errorf("unknown graph routing fixture kind %q", fixture.Kind)
		}
	}
	if totalExpected > 0 {
		result.ImpactRecall = float64(totalHit) / float64(totalExpected)
	}
	if totalTokens > 0 {
		result.FalsePositivePerTok = float64(totalFalsePositive) / float64(totalTokens)
	}
	if result.ImpactRecall+1e-9 < expected.Promotion.MinImpactRecall {
		result.CorrectnessFailures = append(result.CorrectnessFailures, fmt.Sprintf("impact recall %.2f below threshold %.2f", result.ImpactRecall, expected.Promotion.MinImpactRecall))
	}
	if result.FalsePositivePerTok-1e-9 > expected.Promotion.MaxFalsePositivePerToken {
		result.CorrectnessFailures = append(result.CorrectnessFailures, fmt.Sprintf("false positives per token %.4f above threshold %.4f", result.FalsePositivePerTok, expected.Promotion.MaxFalsePositivePerToken))
	}
	if expected.Promotion.RequireZeroSafetyFailures && len(result.SafetyFailures) > 0 {
		result.Ready = false
	}
	if len(result.CorrectnessFailures) > 0 || len(result.SafetyFailures) > 0 {
		result.OK = false
		result.Ready = false
	}
	if math.IsNaN(result.ImpactRecall) || math.IsInf(result.ImpactRecall, 0) {
		result.ImpactRecall = 0
	}
	return result, nil
}

func scoreSymbolImpact(data symbolImpactFixture) (int, int, int, int, []string) {
	failures := []string{}
	retrieved := map[string]bool{}
	for _, path := range data.Retrieved {
		retrieved[path] = true
	}
	hit := 0
	for _, path := range data.ExpectedImpact {
		if retrieved[path] {
			hit++
		} else {
			failures = append(failures, "missed impacted path: "+path)
		}
	}
	for _, path := range append(append([]string{}, data.RequiredTests...), data.RequiredDocs...) {
		if !retrieved[path] {
			failures = append(failures, "missed required test/doc: "+path)
		}
	}
	expected := map[string]bool{}
	for _, path := range data.ExpectedImpact {
		expected[path] = true
	}
	falsePositive := 0
	for _, path := range data.Retrieved {
		if !expected[path] {
			falsePositive++
		}
	}
	if len(data.UncitedExactEdges) > 0 {
		failures = append(failures, "uncited exact edges: "+strings.Join(data.UncitedExactEdges, ","))
	}
	if data.HiddenDynamicLimitations {
		failures = append(failures, "dynamic limitations were hidden")
	}
	return hit, len(data.ExpectedImpact), falsePositive, data.RetrievedTokens, failures
}

func scoreStaleIndex(data staleIndexFixture) ([]string, []string) {
	failures := []string{}
	skipped := []string{}
	if data.OptionalProviderStatus == "skipped-unavailable" {
		skipped = append(skipped, "optional provider unavailable")
	}
	if data.IndexRevision != data.WorktreeRevision || data.DirtyFingerprint != "clean" {
		if data.AcceptedAsAuthoritative {
			failures = append(failures, "stale or dirty index accepted as authoritative")
		}
		if data.FallbackMode != "file-native" {
			failures = append(failures, "stale or dirty index did not fall back to file-native")
		}
	}
	return failures, skipped
}

func scoreMultiSessionRace(data multiSessionRaceFixture) ([]string, []string) {
	failures := []string{}
	skipped := []string{}
	if data.OptionalProviderStatus == "skipped-unavailable" {
		skipped = append(skipped, "optional provider unavailable")
	}
	winners := 0
	for _, claim := range data.SameSliceClaims {
		if claim.Acquired {
			winners++
		}
	}
	if winners != 1 {
		failures = append(failures, fmt.Sprintf("same-slice race had %d winners", winners))
	}
	if !data.DisjointWorktreesIntegrated {
		failures = append(failures, "disjoint worktrees did not integrate serially")
	}
	if !data.DirtyCheckoutPreserved {
		failures = append(failures, "dirty checkout was not preserved")
	}
	if data.ForceCleanupUsed {
		failures = append(failures, "force cleanup was used")
	}
	for _, collision := range data.PrefixCollisions {
		if !collision.Rejected {
			failures = append(failures, "prefix collision was not rejected: "+collision.First+" vs "+collision.Second)
		}
	}
	return failures, skipped
}

func prefixFailures(id string, failures []string) []string {
	out := make([]string, 0, len(failures))
	for _, failure := range failures {
		out = append(out, id+": "+failure)
	}
	return out
}
