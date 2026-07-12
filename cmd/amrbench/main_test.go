package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseOTelDeduplicatesLeafChatSpansAndConvertsAIC(t *testing.T) {
	path := filepath.Join(t.TempDir(), "otel.jsonl")
	row := `{"type":"span","spanId":"one","name":"chat gpt-test","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"gpt-test","gen_ai.usage.input_tokens":100,"gen_ai.usage.output_tokens":25,"github.copilot.nano_aiu":2500000000}}`
	aggregate := `{"type":"span","spanId":"agent","name":"invoke_agent","attributes":{"gen_ai.usage.input_tokens":100,"gen_ai.usage.output_tokens":25,"github.copilot.nano_aiu":2500000000}}`
	if err := os.WriteFile(path, []byte(row+"\n"+row+"\n"+aggregate+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	usage, err := parseOTel(path)
	if err != nil {
		t.Fatal(err)
	}
	if usage.Calls != 1 || usage.InputTokens != 100 || usage.OutputTokens != 25 || usage.AIC != 2.5 || !usage.AICAvailable {
		t.Fatalf("unexpected usage: %+v", usage)
	}
}

func TestParseOTelRejectsPartialAICAndMissingActualModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "otel.jsonl")
	rows := []string{
		`{"type":"span","spanId":"one","name":"chat requested","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual","github.copilot.nano_aiu":1000000000}}`,
		`{"type":"span","spanId":"two","name":"chat requested","attributes":{"gen_ai.provider.name":"github","gen_ai.response.model":"actual"}}`,
	}
	if err := os.WriteFile(path, []byte(strings.Join(rows, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	usage, err := parseOTel(path)
	if err != nil {
		t.Fatal(err)
	}
	if usage.AICAvailable || usage.AIC != 0 {
		t.Fatalf("partial AIC was treated as exact: %+v", usage)
	}

	missingActual := filepath.Join(t.TempDir(), "missing-actual.jsonl")
	row := `{"type":"span","spanId":"one","name":"chat requested","attributes":{"gen_ai.provider.name":"github","gen_ai.request.model":"requested","github.copilot.nano_aiu":1000000000}}`
	if err := os.WriteFile(missingActual, []byte(row), 0o644); err != nil {
		t.Fatal(err)
	}
	usage, err = parseOTel(missingActual)
	if err != nil {
		t.Fatal(err)
	}
	if len(usage.ActualModels) != 0 || modelsMatch(usage.ActualModels, []string{"requested"}) {
		t.Fatalf("missing response model was attributed: %+v", usage)
	}
}

func TestModelsMatchRejectsMixedFallback(t *testing.T) {
	if modelsMatch([]string{"claude-haiku-4.5", "claude-sonnet-5"}, []string{"claude-haiku-4.5"}) {
		t.Fatal("mixed fallback must not be credited to the requested model")
	}
	if !modelsMatch([]string{"gpt-5.6-terra"}, []string{"gpt-5.6-terra"}) {
		t.Fatal("exact model should match")
	}
}

func TestApplyDraftResponseProtectsTestsAndAppliesExistingSource(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "main.go")
	testFile := filepath.Join(root, "main_test.go")
	if err := os.WriteFile(source, []byte("before"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, []byte("oracle"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := applyDraftResponse(root, `{"files":[{"path":"main.go","content":"after"}]}`); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(source)
	if string(data) != "after" {
		t.Fatalf("source=%q", data)
	}
	if err := applyDraftResponse(root, `{"files":[{"path":"main_test.go","content":"cheat"}]}`); err == nil {
		t.Fatal("protected test edit was accepted")
	}
	if err := applyDraftResponse(root, `{"files":[{"path":"MAIN_TEST.GO","content":"cheat"}]}`); err == nil {
		t.Fatal("case-variant protected test edit was accepted")
	}
}

func TestApplyDraftResponseValidatesAllFilesBeforeWriting(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "main.go")
	testFile := filepath.Join(root, "main_test.go")
	if err := os.WriteFile(source, []byte("before"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, []byte("oracle"), 0o644); err != nil {
		t.Fatal(err)
	}
	response := `{"files":[{"path":"main.go","content":"after"},{"path":"main_test.go","content":"cheat"}]}`
	if err := applyDraftResponse(root, response); err == nil {
		t.Fatal("mixed valid/protected response was accepted")
	}
	data, _ := os.ReadFile(source)
	if string(data) != "before" {
		t.Fatalf("source was partially applied: %q", data)
	}
}

func TestExtractJSONObjectStopsAfterFirstValue(t *testing.T) {
	got := extractJSONObject(`prefix {"files":[]} trailing } noise`)
	if got != `{"files":[]}` {
		t.Fatalf("extracted %q", got)
	}
}

func TestApplyProfileRejectsMissingCredentialEnvironment(t *testing.T) {
	t.Setenv("MISSING_AMR_KEY", "")
	_, err := applyProfile(nil, localProfile{
		Alias: "local", BaseURL: "http://localhost:4000/v1",
		ModelID: "gpt-5.4", APIKeyEnv: "MISSING_AMR_KEY",
	})
	if err == nil || !strings.Contains(err.Error(), "MISSING_AMR_KEY") {
		t.Fatalf("missing credential was not reported: %v", err)
	}
}

func TestCorrectionPromptIsBoundedAndRequestsSurgicalRepair(t *testing.T) {
	task := taskSpec{Prompt: "fix it", Verify: []string{"go", "test", "./..."}}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	prompt := correctionPrompt(task, root, strings.Repeat("x", 20000), strings.Repeat("y", 10000))
	if len(prompt) > 20000 {
		t.Fatalf("prompt is unexpectedly large: %d", len(prompt))
	}
	if !strings.Contains(prompt, "Preserve correct existing edits") || !strings.Contains(prompt, "surgically repair") {
		t.Fatalf("missing surgical correction contract: %s", prompt)
	}
}

func TestPinnedSandboxImageRequiresDigest(t *testing.T) {
	t.Setenv("AMRBENCH_GO_IMAGE", "golang:1.25-alpine")
	if _, err := pinnedSandboxImage("AMRBENCH_GO_IMAGE"); err == nil {
		t.Fatal("mutable sandbox image was accepted")
	}
	t.Setenv("AMRBENCH_GO_IMAGE", "golang@sha256:"+strings.Repeat("a", 64))
	if got, err := pinnedSandboxImage("AMRBENCH_GO_IMAGE"); err != nil || got == "" {
		t.Fatalf("pinned image rejected: got=%q err=%v", got, err)
	}
}

func TestPreflightProofFailsBeforeDispatchWithoutSandbox(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("AMRBENCH_GO_IMAGE", "")
	if err := preflightProof([]string{"go", "test", "./..."}); err == nil {
		t.Fatal("missing sandbox runtime was accepted")
	}
}

func TestRequiredGoAndNodeTestEvents(t *testing.T) {
	goOutput := `{"Action":"pass","Test":"TestOne"}` + "\n" + `{"Action":"pass","Test":"TestTwo"}`
	if !requiredTestEventsPresent(goOutput, []string{"TestOne", "TestTwo"}) {
		t.Fatal("Go test events not recognized")
	}
	nodeOutput := "ok 1 - accessible structure\nok 2 - interactive behavior\n"
	if !requiredTestEventsPresent(nodeOutput, []string{"accessible structure", "interactive behavior"}) {
		t.Fatal("Node TAP events not recognized")
	}
}

func TestProofContainerArgsKeepOnlyGoTempExecutable(t *testing.T) {
	args := proofContainerArgs("run-1", `E:\fixture`, "image@sha256:"+strings.Repeat("a", 64), []string{"go", "test"})
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "/tmp:rw,noexec") || !strings.Contains(joined, "/cache:rw,noexec") {
		t.Fatalf("general temp/cache lost noexec: %s", joined)
	}
	if !strings.Contains(joined, "/gotmp:rw,nosuid,size=256m") || strings.Contains(joined, "/gotmp:rw,noexec") {
		t.Fatalf("GOTMPDIR must be bounded but executable: %s", joined)
	}
	if !strings.Contains(joined, "GOTMPDIR=/gotmp") || !strings.Contains(joined, "--network none") || !strings.Contains(joined, "dst=/workspace,readonly") {
		t.Fatalf("sandbox constraints missing: %s", joined)
	}
}

func TestGradeSeparatesDirectAndAMRCohorts(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.json")
	resultsPath := filepath.Join(root, "results.jsonl")
	cfg := `{"schema_version":1,"models":[{"alias":"small","model":"small","runner":"ghcp","tier":"small"}],"tasks":[{"id":"t","family":"logic","planned_tier":"medium","attempt_tier":"small","fixture":"x","prompt":"x","verify":["go","test","./..."],"required_tests":["TestX"],"amr_eligible":true}],"qualification":{"minimum_samples":1,"qualify_pass_rate":0.8,"suspend_pass_rate":0.6}}`
	if err := os.WriteFile(configPath, []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	direct := runResult{Mode: "direct", TaskFamily: "logic", DriverModel: "sol", FinalProof: proofResult{Passed: true}, Phases: []phaseResult{{Model: "sol", ModelMatch: true, AICAvailable: true, AIC: 10}}}
	amr := runResult{Mode: "amr", TaskFamily: "logic", AttemptModel: "haiku", DriverModel: "sonnet", FinalProof: proofResult{Passed: true}, Phases: []phaseResult{{Model: "haiku", ModelMatch: true, AICAvailable: true, AIC: 2}, {Model: "sonnet", ModelMatch: true, AICAvailable: true, AIC: 3}}}
	var data bytes.Buffer
	if err := json.NewEncoder(&data).Encode(direct); err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(&data).Encode(amr); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(resultsPath, data.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := runGrade([]string{"--config", configPath, "--results", resultsPath}, &output); err != nil {
		t.Fatal(err)
	}
	var rows []gradeRow
	if err := json.Unmarshal(output.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Mode == rows[1].Mode {
		t.Fatalf("cohorts were merged: %+v", rows)
	}
}
