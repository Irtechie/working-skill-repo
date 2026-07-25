package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

type contextContract struct {
	SchemaVersion     int               `json:"schema_version"`
	Variant           string            `json:"variant"`
	DevelopmentCorpus []string          `json:"development_corpus"`
	RoutingCorpus     []string          `json:"routing_corpus"`
	BasePacketHash    string            `json:"base_packet_hash"`
	BasePacketPath    string            `json:"base_packet_path"`
	RoleOverlayHashes map[string]string `json:"role_overlay_hashes"`
	RoleOverlayPaths  map[string]string `json:"role_overlay_paths"`
	Tools             []string          `json:"tools"`
	ProofCommand      string            `json:"proof_command"`
	TimeoutSeconds    int               `json:"timeout_seconds"`
	StoppingRule      string            `json:"stopping_rule"`
	CrossoverOrders   [][]string        `json:"crossover_orders"`
	WinnerRule        contextWinnerRule `json:"winner_rule"`
	sourceDir         string
}

type contextWinnerRule struct {
	MinimumMedianReduction float64 `json:"minimum_median_reduction"`
	RequireNoRegression    bool    `json:"require_no_regression"`
	RequireLowerAggregate  bool    `json:"require_lower_aggregate"`
	ConfidenceLevel        float64 `json:"confidence_level"`
}

type contextTrial struct {
	TaskID          string
	Seed            int
	BaselineCorrect bool
	MinimalCorrect  bool
	BaselineAIUNano int64
	MinimalAIUNano  int64
}

type contextDecision struct {
	Winner                string
	Reason                string
	BaselineCorrect       int
	MinimalCorrect        int
	BaselineAggregateNano int64
	MinimalAggregateNano  int64
	PairedMedianReduction float64
	ConfidenceLower       float64
	ConfidenceUpper       float64
}

type contextPayload struct {
	Base     string
	Worker   string
	Reviewer string
}

func loadContextContract(path string) (contextContract, error) {
	var contract contextContract
	content, err := os.ReadFile(path)
	if err != nil {
		return contract, err
	}

	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		return contract, err
	}
	contract.sourceDir = filepath.Dir(path)
	return contract, nil
}

func loadContextPayload(contractPath string) (contextPayload, error) {
	contract, err := loadContextContract(contractPath)
	if err != nil {
		return contextPayload{}, err
	}
	if err := validateContextContract(contract, contract.Variant); err != nil {
		return contextPayload{}, err
	}
	read := func(relative string) (string, error) {
		path := filepath.Join(contract.sourceDir, filepath.FromSlash(relative))
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		if len(content) > 64*1024 {
			return "", fmt.Errorf("context artifact exceeds 64 KiB")
		}
		return string(content), nil
	}
	base, err := read(contract.BasePacketPath)
	if err != nil {
		return contextPayload{}, err
	}
	worker, err := read(contract.RoleOverlayPaths["worker"])
	if err != nil {
		return contextPayload{}, err
	}
	reviewer, err := read(contract.RoleOverlayPaths["reviewer"])
	if err != nil {
		return contextPayload{}, err
	}
	return contextPayload{Base: base, Worker: worker, Reviewer: reviewer}, nil
}

func validateContextPair(baseline, minimal contextContract) error {
	if err := validateContextContract(baseline, "baseline"); err != nil {
		return fmt.Errorf("baseline: %w", err)
	}
	if err := validateContextContract(minimal, "minimal"); err != nil {
		return fmt.Errorf("minimal: %w", err)
	}
	if baseline.BasePacketHash == minimal.BasePacketHash {
		return fmt.Errorf("ambient context variants must use different base packet hashes")
	}
	if !reflect.DeepEqual(baseline.DevelopmentCorpus, minimal.DevelopmentCorpus) ||
		!reflect.DeepEqual(baseline.RoutingCorpus, minimal.RoutingCorpus) ||
		!reflect.DeepEqual(baseline.Tools, minimal.Tools) ||
		baseline.ProofCommand != minimal.ProofCommand ||
		baseline.TimeoutSeconds != minimal.TimeoutSeconds ||
		baseline.StoppingRule != minimal.StoppingRule ||
		!reflect.DeepEqual(baseline.CrossoverOrders, minimal.CrossoverOrders) ||
		!reflect.DeepEqual(baseline.WinnerRule, minimal.WinnerRule) {
		return fmt.Errorf("comparison parity controls differ")
	}
	if !reflect.DeepEqual(baseline.RoleOverlayHashes, minimal.RoleOverlayHashes) {
		return fmt.Errorf("role overlay hashes differ")
	}
	return nil
}

func validateContextContract(contract contextContract, variant string) error {
	if contract.SchemaVersion != 1 || contract.Variant != variant {
		return fmt.Errorf("invalid schema or variant")
	}
	if len(contract.DevelopmentCorpus) == 0 || len(contract.RoutingCorpus) == 0 {
		return fmt.Errorf("development and routing corpora are required")
	}
	seen := map[string]string{}
	for _, value := range contract.DevelopmentCorpus {
		if strings.TrimSpace(value) == "" || seen[value] != "" {
			return fmt.Errorf("development corpus contains an empty or duplicate item")
		}
		seen[value] = "development"
	}
	for _, value := range contract.RoutingCorpus {
		if prior := seen[value]; prior != "" {
			return fmt.Errorf("corpus overlap: %q appears in %s and routing", value, prior)
		}
		seen[value] = "routing"
	}
	hashValid := validSHA256(contract.BasePacketHash)
	if contract.sourceDir == "" {
		hashValid = strings.HasPrefix(contract.BasePacketHash, "sha256:")
	}
	if !hashValid || len(contract.RoleOverlayHashes) == 0 || len(contract.Tools) == 0 ||
		contract.ProofCommand == "" || contract.TimeoutSeconds <= 0 || contract.StoppingRule == "" {
		return fmt.Errorf("contract controls are incomplete")
	}
	for role, hash := range contract.RoleOverlayHashes {
		valid := validSHA256(hash)
		if contract.sourceDir == "" {
			valid = strings.HasPrefix(hash, "sha256:")
		}
		if strings.TrimSpace(role) == "" || !valid {
			return fmt.Errorf("role overlay is invalid")
		}
	}
	if contract.sourceDir != "" {
		if err := verifyContextArtifact(contract.sourceDir, contract.BasePacketPath, contract.BasePacketHash); err != nil {
			return fmt.Errorf("base packet: %w", err)
		}
		if len(contract.RoleOverlayPaths) != len(contract.RoleOverlayHashes) {
			return fmt.Errorf("role overlay paths do not match hashes")
		}
		for role, hash := range contract.RoleOverlayHashes {
			if err := verifyContextArtifact(contract.sourceDir, contract.RoleOverlayPaths[role], hash); err != nil {
				return fmt.Errorf("role overlay %q: %w", role, err)
			}
		}
	}
	if !hasCrossoverOrders(contract.CrossoverOrders) {
		return fmt.Errorf("both crossover orders are required")
	}
	rule := contract.WinnerRule
	if rule.MinimumMedianReduction != 0.10 || !rule.RequireNoRegression ||
		!rule.RequireLowerAggregate || rule.ConfidenceLevel != 0.95 {
		return fmt.Errorf("winner rule is weaker than the preregistered contract")
	}
	return nil
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}

func verifyContextArtifact(root, relative, expected string) error {
	clean := filepath.Clean(filepath.FromSlash(relative))
	if clean == "." || clean == ".." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("artifact path escapes contract root")
	}
	path := filepath.Join(root, clean)
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("artifact must be a regular non-symlink file")
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	relativeResolved, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || relativeResolved == ".." || strings.HasPrefix(relativeResolved, ".."+string(filepath.Separator)) {
		return fmt.Errorf("artifact escapes contract root through a linked parent")
	}
	actual, err := fileSHA256(resolvedPath)
	if err != nil {
		return err
	}
	if expected != actual {
		return fmt.Errorf("artifact hash mismatch")
	}
	return nil
}

func hasCrossoverOrders(orders [][]string) bool {
	found := map[string]bool{}
	for _, order := range orders {
		if len(order) == 2 {
			found[strings.Join(order, ",")] = true
		}
	}
	return found["baseline,minimal"] && found["minimal,baseline"] && len(found) == 2
}

func chooseContextWinner(trials []contextTrial) contextDecision {
	decision := contextDecision{Winner: "baseline", Reason: "insufficient-evidence"}
	if len(trials) < 5 {
		return decision
	}
	seen := map[string]bool{}
	reductions := make([]float64, 0, len(trials))
	for _, trial := range trials {
		key := fmt.Sprintf("%s/%d", trial.TaskID, trial.Seed)
		if trial.TaskID == "" || seen[key] || trial.BaselineAIUNano <= 0 || trial.MinimalAIUNano < 0 {
			decision.Reason = "invalid-pair"
			return decision
		}
		seen[key] = true
		if trial.BaselineCorrect {
			decision.BaselineCorrect++
		}
		if trial.MinimalCorrect {
			decision.MinimalCorrect++
		}
		if trial.BaselineCorrect && !trial.MinimalCorrect {
			decision.Reason = "right-to-wrong"
			return decision
		}
		var err error
		decision.BaselineAggregateNano, err = addContextCost(decision.BaselineAggregateNano, trial.BaselineAIUNano)
		if err != nil {
			decision.Reason = "invalid-cost"
			return decision
		}
		decision.MinimalAggregateNano, err = addContextCost(decision.MinimalAggregateNano, trial.MinimalAIUNano)
		if err != nil {
			decision.Reason = "invalid-cost"
			return decision
		}
		reductions = append(reductions, float64(trial.BaselineAIUNano-trial.MinimalAIUNano)/float64(trial.BaselineAIUNano))
	}
	if decision.MinimalCorrect < decision.BaselineCorrect {
		decision.Reason = "correctness-regression"
		return decision
	}
	decision.PairedMedianReduction = contextMedian(reductions)
	decision.ConfidenceLower, decision.ConfidenceUpper = bootstrapMedianInterval(reductions, 0.95)
	if decision.MinimalAggregateNano >= decision.BaselineAggregateNano {
		decision.Reason = "aggregate-not-lower"
		return decision
	}
	if decision.PairedMedianReduction < 0.10 || decision.ConfidenceLower <= 0 {
		decision.Reason = "savings-not-proven"
		return decision
	}
	decision.Winner = "minimal"
	decision.Reason = "preregistered-rule-passed"
	return decision
}

func bootstrapMedianInterval(values []float64, confidence float64) (float64, float64) {
	const samples = 5000
	values = append([]float64(nil), values...)
	sort.Float64s(values)
	random := rand.New(rand.NewSource(1))
	medians := make([]float64, samples)
	resample := make([]float64, len(values))
	for index := range medians {
		for valueIndex := range resample {
			resample[valueIndex] = values[random.Intn(len(values))]
		}
		medians[index] = contextMedian(resample)
	}
	sort.Float64s(medians)
	tail := (1 - confidence) / 2
	lower := int(math.Floor(tail * samples))
	upper := int(math.Ceil((1-tail)*samples)) - 1
	return medians[lower], medians[upper]
}

func contextMedian(values []float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}

func addContextCost(left, right int64) (int64, error) {
	if right > math.MaxInt64-left {
		return 0, fmt.Errorf("context cost overflow")
	}
	return left + right, nil
}
