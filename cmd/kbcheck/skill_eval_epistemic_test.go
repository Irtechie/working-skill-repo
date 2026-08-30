package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillEvalEpistemicScoresDetectionResolutionRevisionAndCostSeparately(t *testing.T) {
	root := testRepoRoot(t)
	oracle, err := loadEpistemicOracle(root, "investigate-contradicted")
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]any{
		"fixture_id":    "investigate-contradicted",
		"actual":        map[string]any{"user_questions": float64(0)},
		"trace":         map[string]any{"commands": []any{"Get-Content evidence/config.json"}},
		"response_text": "The checked-in flag contradicts the premise, so the rollout should not proceed.",
		"epistemic": map[string]any{
			"decision":           "investigate",
			"evidence_inspected": []any{"evidence/config.json"},
			"final_state":        "contradicted",
		},
	}
	metrics, issues := scoreEpistemicResult(result, oracle)
	if len(issues) != 0 {
		t.Fatalf("unexpected issues: %#v", issues)
	}
	if metrics.Detection.CorrectInvestigation != 1 || metrics.Resolution.Correct != 1 || metrics.Revision.Correct != 1 {
		t.Fatalf("separate epistemic metrics not credited: %#v", metrics)
	}
	if metrics.Cost.UserQuestions != 0 || metrics.Cost.ObservedCommands != 1 || metrics.Cost.Tokens != nil {
		t.Fatalf("cost or unknown telemetry was invented: %#v", metrics.Cost)
	}
}

func TestSkillEvalEpistemicDetectionConfusionMatrix(t *testing.T) {
	cases := []struct {
		name     string
		oracle   epistemicOracle
		decision string
		want     epistemicDetectionMetrics
	}{
		{"correct proceed", epistemicOracle{Decision: "proceed"}, "proceed", epistemicDetectionMetrics{CorrectProceed: 1}},
		{"unnecessary investigation", epistemicOracle{Decision: "proceed"}, "investigate", epistemicDetectionMetrics{UnnecessaryInvestigation: 1}},
		{"missed investigation", epistemicOracle{Decision: "investigate"}, "proceed", epistemicDetectionMetrics{MissedInvestigation: 1}},
		{"correct investigation", epistemicOracle{Decision: "investigate"}, "investigate", epistemicDetectionMetrics{CorrectInvestigation: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := scoreEpistemicResult(map[string]any{"epistemic": map[string]any{"decision": tc.decision}}, tc.oracle)
			if got.Detection != tc.want {
				t.Fatalf("got %#v want %#v", got.Detection, tc.want)
			}
		})
	}
}

func TestSkillEvalEpistemicSupportedProceedRejectsVisibleCeremony(t *testing.T) {
	oracle := epistemicOracle{FixtureID: "supported-proceed", Decision: "proceed", FinalState: "supported", Scored: true}
	result := map[string]any{
		"response_text": "Let me investigate and verify this premise before I continue. Checklist: inspect the config.",
		"epistemic":     map[string]any{"decision": "proceed"},
	}
	metrics, issues := scoreEpistemicResult(result, oracle)
	if !metrics.VisibleCeremony || len(issues) == 0 {
		t.Fatalf("visible verification ceremony was not rejected: metrics=%#v issues=%#v", metrics, issues)
	}
}

func TestSkillEvalEpistemicCorpusRejectsUnverifiableOracle(t *testing.T) {
	root := t.TempDir()
	visible := filepath.Join(root, "evals", "skill-eval", "epistemic", "visible", "bad", "evidence")
	oracles := filepath.Join(root, "evals", "skill-eval", "epistemic", "oracles")
	if err := os.MkdirAll(visible, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(oracles, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(visible, "config.json"), []byte(`{"enabled":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oracles, "bad.json"), []byte(`{"schema_version":1,"fixture_id":"bad","decision":"proceed","final_state":"supported","scored":true,"provenance":{"rule":"hand-authored","path":"evidence/config.json"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if issues := validateEpistemicCorpus(root); len(issues) == 0 {
		t.Fatal("unverifiable hand-authored oracle entered correctness corpus")
	}
}

func TestSkillEvalEpistemicActorWorkspaceExcludesOracleAndFailsClosedOnUnknownSurfaces(t *testing.T) {
	root := testRepoRoot(t)
	workspace := t.TempDir()
	surfaces := []instructionSurface{{Scope: "repo", Path: "AGENTS.md", SHA256: strings.Repeat("a", 64), LoadState: "proven"}, {Scope: "user", Path: "~/.agents/skills", LoadState: "unknown"}}
	if err := materializeEpistemicActorWorkspace(root, "supported-proceed", workspace, surfaces); err == nil {
		t.Fatal("unknown influencing instruction surface did not fail closed")
	}
	surfaces[1].LoadState, surfaces[1].SHA256 = "isolated", strings.Repeat("b", 64)
	if err := materializeEpistemicActorWorkspace(root, "supported-proceed", workspace, surfaces); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(workspace, "evals", "skill-eval", "epistemic", "oracles")); !os.IsNotExist(err) {
		t.Fatal("oracle directory is reachable from actor workspace")
	}
}

func TestSkillEvalBaselineLegacyRowsRemainComparable(t *testing.T) {
	old := []evalRow{{File: "legacy.json", FixtureID: "tiny-typo-fix", ExpectedResult: "pass", ActualResult: "pass"}}
	current := append([]evalRow(nil), old...)
	if issues := compareSkillEvalBaseline(old, current); len(issues) != 0 {
		t.Fatalf("legacy baseline changed: %#v", issues)
	}
}

func TestSkillEvalEpistemicUsesOnlyApprovedPlanningInstructionSurfaces(t *testing.T) {
	root := testRepoRoot(t)
	surfaces, err := epistemicInstructionSurfaces(root, "codex", "dry-run")
	if err != nil {
		t.Fatal(err)
	}
	repoPaths := []string{}
	for _, surface := range surfaces {
		if surface.Scope == "repo" {
			repoPaths = append(repoPaths, filepath.ToSlash(surface.Path))
		}
	}
	want := []string{".agents/skills/kb-plan/SKILL.md", ".agents/skills/kb-gate/SKILL.md"}
	if strings.Join(repoPaths, "|") != strings.Join(want, "|") {
		t.Fatalf("repo instruction surfaces = %#v, want %#v", repoPaths, want)
	}
	workspace := t.TempDir()
	if err := materializeEpistemicInstructionSurfaces(root, workspace, "codex"); err != nil {
		t.Fatal(err)
	}
	for _, rel := range want {
		if _, err := os.Stat(filepath.Join(workspace, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("approved surface %s was not materialized: %v", rel, err)
		}
	}
	for _, rel := range []string{"AGENTS.md", ".github/skills/kb-first-principles/SKILL.md", "config/skill-guidance-audit.json"} {
		if _, err := os.Stat(filepath.Join(workspace, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Fatalf("non-instruction surface %s was materialized", rel)
		}
	}
	manifest := newRunManifest(root, "run", "codex", map[string]any{"id": "supported-proceed", "_epistemic": true}, surfaces, workspace)
	foundGovernanceHash := false
	protected, ok := manifest["protected_files"].([]map[string]any)
	if !ok {
		t.Fatalf("protected_files has unexpected type %T", manifest["protected_files"])
	}
	for _, entry := range protected {
		if stringValue(entry["path"]) == "config/skill-guidance-audit.json" && len(stringValue(entry["sha256"])) == 64 {
			foundGovernanceHash = true
		}
	}
	if !foundGovernanceHash {
		t.Fatal("skill-guidance-audit governance config was not hash-bound in the parent manifest")
	}
}

func TestSkillEvalEpistemicRuntimeSpecificProjectSkillMapping(t *testing.T) {
	root := testRepoRoot(t)
	cases := []struct {
		runtime string
		prefix  string
	}{
		{"codex", ".agents/skills"},
		{"ghcp", ".github/skills"},
	}
	for _, tc := range cases {
		t.Run(tc.runtime, func(t *testing.T) {
			workspace := t.TempDir()
			if err := materializeEpistemicInstructionSurfaces(root, workspace, tc.runtime); err != nil {
				t.Fatal(err)
			}
			for _, skill := range []string{"kb-plan", "kb-gate"} {
				rel := filepath.Join(filepath.FromSlash(tc.prefix), skill, "SKILL.md")
				if _, err := os.Stat(filepath.Join(workspace, rel)); err != nil {
					t.Fatalf("%s did not receive %s at %s: %v", tc.runtime, skill, rel, err)
				}
			}
			other := ".github/skills"
			if tc.runtime == "ghcp" {
				other = ".agents/skills"
			}
			if _, err := os.Stat(filepath.Join(workspace, filepath.FromSlash(other), "kb-plan", "SKILL.md")); !os.IsNotExist(err) {
				t.Fatalf("%s received a duplicate project skill under %s", tc.runtime, other)
			}
		})
	}
}

func TestSkillEvalEpistemicManifestBindsRouterSelectionAndRuntimeIdentity(t *testing.T) {
	root := testRepoRoot(t)
	actorRoot := t.TempDir()
	surfaces, err := epistemicInstructionSurfaces(root, "codex", "dry-run")
	if err != nil {
		t.Fatal(err)
	}
	fixture := map[string]any{"id": "supported-proceed", "_epistemic": true}
	config, contractPath := writeEpistemicTestRouterContract(t, root)
	manifest, err := newRoutedEpistemicRunManifest(root, "run", "codex", fixture, surfaces, actorRoot, config, contractPath)
	if err != nil {
		t.Fatal(err)
	}
	if stringValue(manifest["model_selection_owner"]) != "kbrouter" || stringValue(manifest["selected_route_alias"]) != config.RouteAlias || len(stringValue(manifest["router_dispatch_contract_sha256"])) != 64 {
		t.Fatalf("router-selected route contract not bound exactly: %#v", manifest)
	}
	if _, exists := manifest["requested_model"]; exists {
		t.Fatalf("epistemic preview retained coordinator-selected model: %#v", manifest["requested_model"])
	}
	identity, _ := manifest["runtime_identity"].(map[string]any)
	if stringValue(identity["runtime"]) != "codex" || stringValue(identity["executable"]) == "" || len(stringValue(identity["sha256"])) != 64 {
		t.Fatalf("runtime executable identity/hash not bound: %#v", identity)
	}
}

func TestSkillEvalEpistemicCodexIsolationDisablesAndHashBindsLocalSkills(t *testing.T) {
	root := t.TempDir()
	codexHome := filepath.Join(root, "codex-home")
	userHome := filepath.Join(root, "user-home")
	pluginRoot := filepath.Join(root, "plugin", "cache")
	files := map[string]string{
		filepath.Join(codexHome, "skills", "local-a", "SKILL.md"):            "local-a",
		filepath.Join(codexHome, "skills", ".system", "bundled", "SKILL.md"): "bundled",
		filepath.Join(userHome, ".agents", "skills", "shared-a", "SKILL.md"): "shared-a",
		filepath.Join(pluginRoot, "pkg", "skills", "plugin-a", "SKILL.md"):   "plugin-a",
	}
	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	projectSkills := []string{".agents/skills/kb-plan/SKILL.md", ".agents/skills/kb-gate/SKILL.md"}
	config, err := buildCodexEpistemicSkillIsolation(codexHome, userHome, []string{pluginRoot}, projectSkills)
	if err != nil {
		t.Fatal(err)
	}
	wantDisabled := map[string]bool{}
	for path := range files {
		wantDisabled[filepath.Clean(path)] = true
	}
	if len(config.Disabled) != len(wantDisabled) {
		t.Fatalf("disabled=%#v want paths=%#v", config.Disabled, wantDisabled)
	}
	for _, entry := range config.Disabled {
		path := filepath.Clean(entry.Path)
		if !wantDisabled[path] || entry.Enabled || len(entry.SHA256) != 64 || entry.SHA256 != fileHashOrEmpty(path) {
			t.Fatalf("invalid disabled/hash-bound skill entry: %#v", entry)
		}
	}
	if strings.Join(config.EnabledProject, "|") != strings.Join(projectSkills, "|") {
		t.Fatalf("enabled project skills=%#v want %#v", config.EnabledProject, projectSkills)
	}
	for _, path := range projectSkills {
		if strings.Contains(config.CLIConfig, "path='"+path+"',enabled=false") {
			t.Fatalf("project planning skill disabled in CLI config: %s", path)
		}
	}
	for path := range wantDisabled {
		if !strings.Contains(config.CLIConfig, "path='"+filepath.ToSlash(path)+"',enabled=false") {
			t.Fatalf("CLI isolation config omitted %s: %s", path, config.CLIConfig)
		}
	}
}

func TestSkillEvalEpistemicActorWorkspaceIsExternalFreshGitRoot(t *testing.T) {
	sourceRoot := testRepoRoot(t)
	runDir := filepath.Join(sourceRoot, ".kb", "eval-runs", "nested-run")
	actorRoot, cleanup, err := createEpistemicActorWorkspace(sourceRoot, runDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	sourceAbs, err := filepath.Abs(sourceRoot)
	if err != nil {
		t.Fatal(err)
	}
	actorAbs, err := filepath.Abs(actorRoot)
	if err != nil {
		t.Fatal(err)
	}
	if strings.EqualFold(filepath.VolumeName(sourceAbs), filepath.VolumeName(actorAbs)) {
		rel, err := filepath.Rel(sourceAbs, actorAbs)
		if err != nil {
			t.Fatal(err)
		}
		if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Fatalf("actor workspace %s is nested under source repository %s", actorAbs, sourceAbs)
		}
	}
	info, err := os.Stat(filepath.Join(actorAbs, ".git"))
	if err != nil || !info.IsDir() {
		t.Fatalf("actor workspace is not a fresh git root: %v", err)
	}
	for current := filepath.Dir(actorAbs); ; current = filepath.Dir(current) {
		if samePath(current, sourceAbs) {
			t.Fatalf("source repository AGENTS.md remains in actor instruction ancestry: %s", current)
		}
		next := filepath.Dir(current)
		if next == current {
			break
		}
	}
}

func TestSkillEvalEpistemicCodexIsolationPreservesSurfaceScope(t *testing.T) {
	root := t.TempDir()
	codexHome := filepath.Join(root, "codex-home")
	userHome := filepath.Join(root, "user-home")
	pluginRoot := filepath.Join(root, "plugin", "cache")
	wantScopes := map[string]string{
		filepath.Clean(filepath.Join(codexHome, "skills", "local-a", "SKILL.md")):            "user",
		filepath.Clean(filepath.Join(codexHome, "skills", ".system", "bundled", "SKILL.md")): "global",
		filepath.Clean(filepath.Join(userHome, ".agents", "skills", "shared-a", "SKILL.md")): "user",
		filepath.Clean(filepath.Join(pluginRoot, "pkg", "skills", "plugin-a", "SKILL.md")):   "cached",
	}
	for path := range wantScopes {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(path), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	projectSkills := []string{".agents/skills/kb-plan/SKILL.md", ".agents/skills/kb-gate/SKILL.md"}
	config, err := buildCodexEpistemicSkillIsolation(codexHome, userHome, []string{pluginRoot}, projectSkills)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range config.Disabled {
		if got, want := entry.Scope, wantScopes[filepath.Clean(entry.Path)]; got != want {
			t.Fatalf("disabled skill %s scope=%q want %q", entry.Path, got, want)
		}
	}
	surfaces := codexIsolationInstructionSurfaces(config)
	if len(surfaces) != len(wantScopes) {
		t.Fatalf("surface count=%d want %d", len(surfaces), len(wantScopes))
	}
	for _, surface := range surfaces {
		if got, want := surface.Scope, wantScopes[filepath.Clean(surface.Path)]; got != want {
			t.Fatalf("surface %s scope=%q want %q", surface.Path, got, want)
		}
		if surface.LoadState != "isolated" || len(surface.SHA256) != 64 {
			t.Fatalf("surface lacks isolation/hash proof: %#v", surface)
		}
	}
}

func TestSkillEvalEpistemicRunPreservesAuditableActorWorkspace(t *testing.T) {
	root := testRepoRoot(t)
	fixture, err := loadVisibleEpistemicFixture(root, "supported-proceed")
	if err != nil {
		t.Fatal(err)
	}
	_, contractPath := writeEpistemicTestRouterContract(t, root)
	run, err := runOneAdapterFixture(root, filepath.Join(root, ".kb", "epistemic-audit-test-"+randomShortHash()), "codex", "dry-run", fixture, options{dryRun: true, keepRun: true, agentCommand: "kbrouter-contract:" + contractPath})
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(run.RunDir)
	var manifest map[string]any
	if err := readJSONFile(run.ManifestPath, &manifest); err != nil {
		t.Fatal(err)
	}
	auditRoot := stringValue(manifest["actor_workspace"])
	rel, err := filepath.Rel(run.RunDir, auditRoot)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		t.Fatalf("auditable actor workspace %s is not preserved under run directory %s", auditRoot, run.RunDir)
	}
	if info, err := os.Stat(auditRoot); err != nil || !info.IsDir() {
		t.Fatalf("auditable actor workspace missing after run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(auditRoot, "evals", "skill-eval", "epistemic", "oracles")); !os.IsNotExist(err) {
		t.Fatal("saved actor workspace contains hidden oracle material")
	}
	if !boolValue(manifest["execution_actor_workspace_external"]) || len(stringValue(manifest["execution_actor_workspace_sha256"])) != 64 {
		t.Fatalf("external execution workspace proof missing: %#v", manifest)
	}
	if stringValue(manifest["execution_actor_workspace_sha256"]) != directoryContentHash(auditRoot) {
		t.Fatal("saved audit workspace does not match the pre-run external workspace hash")
	}
}

func TestSkillEvalEpistemicCodexDryRunIsExactNoCallPreview(t *testing.T) {
	root := testRepoRoot(t)
	fixture, err := loadVisibleEpistemicFixture(root, "supported-proceed")
	if err != nil {
		t.Fatal(err)
	}
	config, contractPath := writeEpistemicTestRouterContract(t, root)
	run, err := runOneAdapterFixture(root, filepath.Join(root, ".kb", "epistemic-preview-test-"+randomShortHash()), "codex", "dry-run", fixture, options{dryRun: true, keepRun: true, agentCommand: "kbrouter-contract:" + contractPath})
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(run.RunDir)
	var manifest map[string]any
	if err := readJSONFile(run.ManifestPath, &manifest); err != nil {
		t.Fatal(err)
	}
	if stringValue(manifest["model_selection_owner"]) != "kbrouter" || stringValue(manifest["selected_route_alias"]) != config.RouteAlias {
		t.Fatalf("dry-run preview did not bind router-selected route: %#v", manifest)
	}
	if _, exists := manifest["requested_model"]; exists {
		t.Fatalf("dry-run preview accepted a coordinator-selected model: %#v", manifest)
	}
	identity, _ := manifest["runtime_identity"].(map[string]any)
	if stringValue(identity["runtime"]) != "codex" || len(stringValue(identity["sha256"])) != 64 || strings.TrimSpace(stringValue(identity["version"])) == "" {
		t.Fatalf("dry-run preview lacks exact runtime identity: %#v", identity)
	}
	var result map[string]any
	if err := readJSONFile(run.ResultPath, &result); err != nil {
		t.Fatal(err)
	}
	trace, _ := result["trace"].(map[string]any)
	commands, _ := trace["commands"].([]any)
	if len(commands) != 1 || stringValue(commands[0]) != "dry-run" {
		t.Fatalf("preview did not preserve no-call trace: %#v", trace)
	}
	surfaceValues, _ := manifest["instruction_surfaces"].([]any)
	if len(surfaceValues) < 2 {
		t.Fatalf("preview instruction inventory is incomplete: %#v", surfaceValues)
	}
	for _, value := range surfaceValues {
		surface, _ := value.(map[string]any)
		if stringValue(surface["path"]) == "<isolated>" {
			t.Fatalf("preview retained placeholder isolation instead of the exact inventory: %#v", surface)
		}
		if len(stringValue(surface["sha256"])) != 64 {
			t.Fatalf("preview surface is not hash-bound: %#v", surface)
		}
	}
}

func TestSkillEvalEpistemicPromptInvokesPlanningTreatmentSkills(t *testing.T) {
	fixture := map[string]any{"id": "supported-proceed", "task": "Plan the follow-up.", "_epistemic": true}
	prompt := evalPrompt(fixture, "codex", "run-id")
	for _, invocation := range []string{"$kb-plan", "$kb-gate"} {
		if !strings.Contains(prompt, invocation) {
			t.Fatalf("epistemic prompt does not invoke %s: %s", invocation, prompt)
		}
	}
}

func TestSkillEvalEpistemicRoutedCommandDelegatesModelChoiceAndSealsEvaluatorContract(t *testing.T) {
	config := epistemicRouterDispatchConfig{
		SchemaVersion: 1, Command: "kbrouter", ProjectRoot: `E:\source`, RunRoot: `E:\source\.kb\route-run`,
		RunID: "route-run", SliceID: "epistemic-fixture", Packet: "packet.json", RouteAlias: "selected.planner",
	}
	actor := `E:\external-actor`
	prompt := actor + `\sealed-prompt.txt`
	schema := actor + `\result.schema.json`
	isolation := `skills.config=[{path="ambient",enabled=false}]`
	command, args, err := epistemicRoutedAgentCommand(config, actor, prompt, schema, "epistemic-result.json", isolation)
	if err != nil {
		t.Fatal(err)
	}
	if command != "kbrouter" {
		t.Fatalf("command=%q", command)
	}
	joined := " " + strings.Join(args, " ") + " "
	for _, required := range []string{
		" dispatch ", " --route-alias selected.planner ", " --evaluator-actor-cwd " + actor + " ",
		" --evaluator-prompt " + prompt + " ", " --evaluator-output-schema " + schema + " ",
		" --evaluator-structured-output epistemic-result.json ", " --evaluator-instruction-config " + isolation + " ",
		" --output epistemic-result-dispatch-output.json ", " --receipt epistemic-result-dispatch-receipt.json ",
		" --handoff epistemic-result-dispatch-handoff.json ",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("routed evaluator args missing %q: %s", required, joined)
		}
	}
	if strings.Contains(joined, " --model ") || strings.Contains(joined, "reasoning") {
		t.Fatalf("evaluator statically selected a model or reasoning effort: %s", joined)
	}
	config.RouteAlias = "current"
	if _, _, err := epistemicRoutedAgentCommand(config, actor, prompt, schema, "epistemic-result.json", isolation); err == nil {
		t.Fatal("current/App-only route was accepted for external evaluator dispatch")
	}
	if _, _, err := invokeLiveAgent(actor, "codex", map[string]any{"id": "fixture", "_epistemic": true}, "run", options{model: "statically-forced"}, isolation); err == nil || !strings.Contains(err.Error(), "direct model dispatch is refused") {
		t.Fatalf("epistemic evaluator accepted direct static model dispatch: %v", err)
	}
}

func TestSkillEvalEpistemicRuntimeIdentityBindsReportedVersion(t *testing.T) {
	identity := runtimeIdentity("codex")
	version := strings.TrimSpace(stringValue(identity["version"]))
	if version == "" {
		t.Fatalf("Codex runtime identity lacks runtime-reported version: %#v", identity)
	}
	if stringValue(identity["executable"]) == "" || len(stringValue(identity["sha256"])) != 64 {
		t.Fatalf("Codex runtime identity lost executable/hash binding: %#v", identity)
	}
	root := testRepoRoot(t)
	actorRoot := t.TempDir()
	surfaces, err := epistemicInstructionSurfaces(root, "codex", "dry-run")
	if err != nil {
		t.Fatal(err)
	}
	config, contractPath := writeEpistemicTestRouterContract(t, root)
	manifest, err := newRoutedEpistemicRunManifest(root, "run", "codex", map[string]any{"id": "supported-proceed", "_epistemic": true}, surfaces, actorRoot, config, contractPath)
	if err != nil {
		t.Fatal(err)
	}
	manifestIdentity, _ := manifest["runtime_identity"].(map[string]any)
	if got := strings.TrimSpace(stringValue(manifestIdentity["version"])); got != version {
		t.Fatalf("manifest runtime version=%q want exact reported version %q", got, version)
	}
}

func writeEpistemicTestRouterContract(t *testing.T, root string) (epistemicRouterDispatchConfig, string) {
	t.Helper()
	contractRoot := t.TempDir()
	config := epistemicRouterDispatchConfig{
		SchemaVersion: 1,
		Command:       "kbrouter",
		ProjectRoot:   root,
		RunRoot:       filepath.Join(root, ".kb", "router-test-run"),
		RunID:         "router-test-run",
		SliceID:       "epistemic-fixture",
		Packet:        "packet.json",
		RouteAlias:    "selected.planner",
	}
	path := filepath.Join(contractRoot, "router-contract.json")
	writeJSONFile(path, config)
	return config, path
}

func TestSkillEvalEpistemicRegressionDecision(t *testing.T) {
	complete := func(missed, unnecessary, resolution, revision, questions, ceremony int) map[string]any {
		return map[string]any{
			"result_count": 4, "missed_investigation": missed, "unnecessary_investigation": unnecessary,
			"resolution_correct": resolution, "revision_correct": revision,
			"user_questions": questions, "visible_ceremony": ceremony,
		}
	}
	t.Run("zero missed held with no protected regression promotes", func(t *testing.T) {
		baseline := complete(0, 0, 2, 2, 0, 0)
		current := complete(0, 0, 2, 2, 0, 0)
		if got := compareEpistemicRegression(baseline, current); got != "promote" {
			t.Fatalf("got %q want promote", got)
		}
	})
	t.Run("worse protected metric rejects", func(t *testing.T) {
		baseline := complete(0, 0, 2, 2, 0, 0)
		current := complete(0, 1, 2, 2, 0, 0)
		if got := compareEpistemicRegression(baseline, current); got != "reject" {
			t.Fatalf("got %q want reject", got)
		}
	})
	t.Run("missing comparison evidence is inconclusive", func(t *testing.T) {
		baseline := complete(0, 0, 2, 2, 0, 0)
		delete(baseline, "revision_correct")
		if got := compareEpistemicRegression(baseline, complete(0, 0, 2, 2, 0, 0)); got != "inconclusive" {
			t.Fatalf("got %q want inconclusive", got)
		}
	})
}

func testRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}
