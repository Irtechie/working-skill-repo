package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContextContractRejectsCorpusOverlapAndControlDrift(t *testing.T) {
	baseline := validContextContract("baseline")
	minimal := validContextContract("minimal")
	if err := validateContextPair(baseline, minimal); err != nil {
		t.Fatal(err)
	}

	minimal.RoutingCorpus = append(minimal.RoutingCorpus, baseline.DevelopmentCorpus[0])
	if err := validateContextPair(baseline, minimal); err == nil || !strings.Contains(err.Error(), "overlap") {
		t.Fatalf("corpus overlap err=%v", err)
	}

	minimal = validContextContract("minimal")
	minimal.TimeoutSeconds++
	if err := validateContextPair(baseline, minimal); err == nil || !strings.Contains(err.Error(), "parity") {
		t.Fatalf("control drift err=%v", err)
	}
}

func TestContextContractRequiresCrossoverAndFrozenOverlays(t *testing.T) {
	baseline := validContextContract("baseline")
	minimal := validContextContract("minimal")
	minimal.CrossoverOrders = [][]string{{"baseline", "minimal"}}
	if err := validateContextPair(baseline, minimal); err == nil || !strings.Contains(err.Error(), "crossover") {
		t.Fatalf("crossover err=%v", err)
	}

	minimal = validContextContract("minimal")
	minimal.RoleOverlayHashes["worker"] = "sha256:changed"
	if err := validateContextPair(baseline, minimal); err == nil || !strings.Contains(err.Error(), "overlay") {
		t.Fatalf("overlay err=%v", err)
	}
}

func TestContextWinnerRejectsRegressionAndWeakSavings(t *testing.T) {
	wins := []contextTrial{
		{TaskID: "a", Seed: 1, BaselineCorrect: true, MinimalCorrect: true, BaselineAIUNano: 100, MinimalAIUNano: 80},
		{TaskID: "b", Seed: 2, BaselineCorrect: true, MinimalCorrect: true, BaselineAIUNano: 100, MinimalAIUNano: 75},
		{TaskID: "c", Seed: 3, BaselineCorrect: false, MinimalCorrect: true, BaselineAIUNano: 100, MinimalAIUNano: 70},
		{TaskID: "d", Seed: 4, BaselineCorrect: true, MinimalCorrect: true, BaselineAIUNano: 100, MinimalAIUNano: 80},
		{TaskID: "e", Seed: 5, BaselineCorrect: true, MinimalCorrect: true, BaselineAIUNano: 100, MinimalAIUNano: 75},
	}
	decision := chooseContextWinner(wins)
	if decision.Winner != "minimal" || decision.PairedMedianReduction < 0.10 || decision.ConfidenceLower <= 0 {
		t.Fatalf("decision=%+v", decision)
	}

	regression := append([]contextTrial(nil), wins...)
	regression[0].MinimalCorrect = false
	if decision := chooseContextWinner(regression); decision.Winner != "baseline" || decision.Reason != "right-to-wrong" {
		t.Fatalf("regression decision=%+v", decision)
	}

	weak := append([]contextTrial(nil), wins...)
	for index := range weak {
		weak[index].MinimalAIUNano = 95
	}
	if decision := chooseContextWinner(weak); decision.Winner != "baseline" {
		t.Fatalf("weak decision=%+v", decision)
	}
}

func TestContextContractsLoadAndSchemaDeclaresRequiredControls(t *testing.T) {
	root := filepath.Join("..", "..", "evals", "amr-model-benchmark")
	baseline, err := loadContextContract(filepath.Join(root, "context-contracts", "baseline.json"))
	if err != nil {
		t.Fatal(err)
	}
	minimal, err := loadContextContract(filepath.Join(root, "context-contracts", "minimal.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateContextPair(baseline, minimal); err != nil {
		t.Fatal(err)
	}
	schema, err := os.ReadFile(filepath.Join(root, "context-contract.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"development_corpus", "routing_corpus", "role_overlay_hashes", "crossover_orders", "timeout_seconds", "winner_rule"} {
		if !strings.Contains(string(schema), `"`+field+`"`) {
			t.Fatalf("schema missing %q", field)
		}
	}
}

func validContextContract(variant string) contextContract {
	return contextContract{
		SchemaVersion:     1,
		Variant:           variant,
		DevelopmentCorpus: []string{"dev-a", "dev-b"},
		RoutingCorpus:     []string{"holdout-a", "holdout-b"},
		BasePacketHash:    "sha256:" + variant,
		RoleOverlayHashes: map[string]string{"worker": "sha256:worker", "reviewer": "sha256:reviewer"},
		Tools:             []string{"apply_patch", "go test"},
		ProofCommand:      "go test ./...",
		TimeoutSeconds:    90,
		StoppingRule:      "one-pass-plus-proof",
		CrossoverOrders:   [][]string{{"baseline", "minimal"}, {"minimal", "baseline"}},
		WinnerRule: contextWinnerRule{
			MinimumMedianReduction: 0.10,
			RequireNoRegression:    true,
			RequireLowerAggregate:  true,
			ConfidenceLevel:        0.95,
		},
	}
}
