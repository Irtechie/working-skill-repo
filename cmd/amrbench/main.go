package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/ghcpotel"
)

const defaultConfig = "evals/amr-model-benchmark/config.json"

type config struct {
	SchemaVersion    int                  `json:"schema_version"`
	Models           []modelSpec          `json:"models"`
	Cohorts          []cohortSpec         `json:"cohorts,omitempty"`
	Tasks            []taskSpec           `json:"tasks"`
	Qualification    qualification        `json:"qualification"`
	Sandbox          sandboxSpec          `json:"sandbox"`
	ContextContracts contextContractPaths `json:"context_contracts"`
	Budget           budgetSpec           `json:"budget"`
}

type budgetSpec struct {
	PerCallAIU    int `json:"per_call_ai_credits"`
	PerArmAIU     int `json:"per_arm_ai_credits"`
	ExperimentAIU int `json:"experiment_ai_credits"`
}

type contextContractPaths struct {
	Baseline string `json:"baseline"`
	Minimal  string `json:"minimal"`
}

type sandboxSpec struct {
	GoImage   string `json:"go_image,omitempty"`
	NodeImage string `json:"node_image,omitempty"`
}

type cohortSpec struct {
	ID           string   `json:"id"`
	PlannedTier  string   `json:"planned_tier"`
	AttemptTier  string   `json:"attempt_tier"`
	TaskFamilies []string `json:"task_families"`
}

type modelSpec struct {
	Alias   string `json:"alias"`
	Model   string `json:"model"`
	Runner  string `json:"runner"`
	Tier    string `json:"tier"`
	Profile string `json:"profile,omitempty"`
}

type taskSpec struct {
	ID              string            `json:"id"`
	Family          string            `json:"family"`
	PlannedTier     string            `json:"planned_tier"`
	AttemptTier     string            `json:"attempt_tier,omitempty"`
	Fixture         string            `json:"fixture"`
	Prompt          string            `json:"prompt"`
	Verify          []string          `json:"verify"`
	RequiredTests   []string          `json:"required_tests"`
	AMREligible     bool              `json:"amr_eligible"`
	IneligibleWhy   string            `json:"ineligible_reason,omitempty"`
	MutablePaths    []string          `json:"mutable_paths,omitempty"`
	ProofFiles      []string          `json:"proof_files,omitempty"`
	ProofHashes     map[string]string `json:"proof_hashes,omitempty"`
	SolutionFixture string            `json:"solution_fixture,omitempty"`
}

type qualification struct {
	MinimumSamples int     `json:"minimum_samples"`
	QualifyRate    float64 `json:"qualify_pass_rate"`
	SuspendRate    float64 `json:"suspend_pass_rate"`
}

type localProfiles struct {
	Profiles []localProfile `json:"profiles"`
}

type localProfile struct {
	Alias           string `json:"alias"`
	BaseURL         string `json:"base_url"`
	ProviderType    string `json:"provider_type"`
	ModelID         string `json:"model_id"`
	WireModel       string `json:"wire_model,omitempty"`
	WireAPI         string `json:"wire_api,omitempty"`
	APIKeyEnv       string `json:"api_key_env,omitempty"`
	BearerTokenEnv  string `json:"bearer_token_env,omitempty"`
	MaxPromptTokens int    `json:"max_prompt_tokens,omitempty"`
	MaxOutputTokens int    `json:"max_output_tokens,omitempty"`
}

type runResult struct {
	SchemaVersion       int           `json:"schema_version"`
	RunID               string        `json:"run_id"`
	Mode                string        `json:"mode"`
	TaskID              string        `json:"task_id"`
	TaskFamily          string        `json:"task_family"`
	Seed                int           `json:"seed"`
	ContextContractHash string        `json:"context_contract_hash"`
	ProofClosureHash    string        `json:"proof_closure_hash"`
	ExperimentID        string        `json:"experiment_id"`
	ApprovalHash        string        `json:"approval_hash"`
	RouteCatalogHash    string        `json:"route_catalog_hash"`
	PlannedTier         string        `json:"planned_tier"`
	AttemptTier         string        `json:"attempt_tier,omitempty"`
	AttemptModel        string        `json:"attempt_model,omitempty"`
	DriverModel         string        `json:"driver_model,omitempty"`
	StartedAt           string        `json:"started_at"`
	DurationMS          int64         `json:"duration_ms"`
	FinalProof          proofResult   `json:"final_proof"`
	Phases              []phaseResult `json:"phases"`
	ChangedFiles        []string      `json:"changed_files"`
	LinesAdded          int           `json:"lines_added"`
	LinesDeleted        int           `json:"lines_deleted"`
	Outcome             string        `json:"outcome"`
	Workspace           string        `json:"workspace"`
}

type phaseResult struct {
	Phase            string      `json:"phase"`
	Model            string      `json:"model"`
	RequestedModel   string      `json:"requested_model"`
	ActualModels     []string    `json:"actual_models"`
	ModelMatch       bool        `json:"model_match"`
	Runner           string      `json:"runner"`
	ExitCode         int         `json:"exit_code"`
	Valid            bool        `json:"valid"`
	InvalidReason    string      `json:"invalid_reason,omitempty"`
	DurationMS       int64       `json:"duration_ms"`
	AIC              float64     `json:"aic,omitempty"`
	AIUNano          int64       `json:"aiu_nano"`
	AICAvailable     bool        `json:"aic_available"`
	InputTokens      int64       `json:"input_tokens"`
	CacheReadTokens  int64       `json:"cache_read_tokens"`
	CacheWriteTokens int64       `json:"cache_write_tokens"`
	OutputTokens     int64       `json:"output_tokens"`
	Calls            int         `json:"calls"`
	ApplyError       string      `json:"apply_error,omitempty"`
	Proof            proofResult `json:"proof"`
	StdoutPath       string      `json:"stdout_path"`
	StderrPath       string      `json:"stderr_path"`
	OTelPath         string      `json:"otel_path"`
}

type proofResult struct {
	Passed       bool   `json:"passed"`
	ExitCode     int    `json:"exit_code"`
	Output       string `json:"output"`
	SandboxImage string `json:"sandbox_image,omitempty"`
}

type phaseState struct {
	SchemaVersion int    `json:"schema_version"`
	Phase         string `json:"phase"`
	Status        string `json:"status"`
	ExitCode      int    `json:"exit_code,omitempty"`
	Reason        string `json:"reason,omitempty"`
	UpdatedAt     string `json:"updated_at"`
}

type gradeRow struct {
	Mode       string  `json:"mode"`
	Route      string  `json:"route"`
	TaskFamily string  `json:"task_family"`
	Samples    int     `json:"samples"`
	Passes     int     `json:"passes"`
	PassRate   float64 `json:"pass_rate"`
	Mismatches int     `json:"model_mismatches"`
	TotalAIC   float64 `json:"total_aic"`
	MedianAIC  float64 `json:"median_aic"`
	Status     string  `json:"status"`
}

type otelUsage struct {
	AIC              float64
	AIUNano          int64
	AICAvailable     bool
	InputTokens      int64
	OutputTokens     int64
	CacheReadTokens  int64
	CacheWriteTokens int64
	Calls            int
	ActualModels     []string
}

type draftResponse struct {
	Files []draftFile `json:"files"`
}

type draftFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type conformanceReport struct {
	SchemaVersion   int                             `json:"schema_version"`
	Ready           bool                            `json:"ready"`
	NoPaid          bool                            `json:"no_paid"`
	Runner          string                          `json:"runner"`
	PaidCalls       int                             `json:"paid_calls"`
	ReleaseDecision string                          `json:"release_decision"`
	Containment     containmentReport               `json:"containment"`
	Fixtures        map[string]fixtureQualification `json:"fixtures,omitempty"`
	Issues          []string                        `json:"issues,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "list":
		err = runList(os.Args[2:], os.Stdout)
	case "run":
		err = runBenchmark(os.Args[2:], os.Stdout)
	case "grade":
		err = errors.New("legacy unpaired grade is disabled; use grade-paired")
	case "grade-paired":
		err = runPairedGrade(os.Args[2:], os.Stdout)
	case "conformance":
		err = runConformance(os.Args[2:], os.Stdout)
	default:
		usage(os.Stderr)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "amrbench list [--config path]")
	fmt.Fprintln(w, "amrbench run --dry-run --experiment-id id --routes path --context baseline|minimal --mode direct|amr --task id [--model id | --attempt id --driver id] [--repeat N] [--config path]")
	fmt.Fprintln(w, "amrbench run without --dry-run is disabled until a trusted human-approval verifier exists")
	fmt.Fprintln(w, "amrbench grade is disabled; use grade-paired")
	fmt.Fprintln(w, "amrbench grade-paired --results path")
	fmt.Fprintln(w, "amrbench conformance --config path --no-paid [--require-ready] [--json]")
}

func runConformance(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("conformance", flag.ContinueOnError)
	configPath := fs.String("config", defaultConfig, "benchmark config")
	noPaid := fs.Bool("no-paid", false, "disable all model runners")
	requireReady := fs.Bool("require-ready", false, "fail unless every deterministic gate is ready")
	asJSON := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("conformance accepts no positional arguments")
	}
	if !*noPaid {
		return errors.New("conformance requires --no-paid; live execution needs separate attended approval")
	}
	var cfg config
	if err := readJSON(*configPath, &cfg); err != nil {
		return err
	}
	if cfg.SchemaVersion != 1 {
		return errors.New("invalid benchmark config schema")
	}
	configureSandboxImages(cfg.Sandbox)
	runner := newDisabledRunner(nil, nil)
	report := conformanceReport{
		SchemaVersion:   1,
		NoPaid:          true,
		Runner:          runner.Name(),
		ReleaseDecision: "not-promoted",
		Containment:     inspectContainment(sandboxRuntime),
		Fixtures:        map[string]fixtureQualification{},
	}
	if cfg.Budget.PerCallAIU <= 0 || cfg.Budget.PerArmAIU < cfg.Budget.PerCallAIU ||
		cfg.Budget.ExperimentAIU < cfg.Budget.PerArmAIU {
		report.Issues = append(report.Issues, "credit budget must define increasing positive per-call, per-arm, and experiment ceilings")
	}
	if cfg.ContextContracts.Baseline == "" || cfg.ContextContracts.Minimal == "" {
		report.Issues = append(report.Issues, "context contracts are not configured")
	} else {
		baseline, baselineErr := loadContextContract(cfg.ContextContracts.Baseline)
		minimal, minimalErr := loadContextContract(cfg.ContextContracts.Minimal)
		switch {
		case baselineErr != nil:
			report.Issues = append(report.Issues, "baseline context contract: "+baselineErr.Error())
		case minimalErr != nil:
			report.Issues = append(report.Issues, "minimal context contract: "+minimalErr.Error())
		case validateContextPair(baseline, minimal) != nil:
			report.Issues = append(report.Issues, "context contract parity failed")
		}
	}
	if !report.Containment.Ready {
		report.Issues = append(report.Issues, report.Containment.Issues...)
	}
	for _, task := range cfg.Tasks {
		if !task.AMREligible {
			continue
		}
		if _, _, err := sandboxCommand(task.Verify); err != nil {
			report.Issues = append(report.Issues, task.ID+": "+err.Error())
			continue
		}
		qualified, err := qualifyFixture(fixtureSpec{
			Root: task.Fixture, MutablePaths: task.MutablePaths, ProofFiles: task.ProofFiles,
			ExpectedProofHashes: task.ProofHashes, SolutionRoot: task.SolutionFixture,
			Verify: task.Verify, RequiredTests: task.RequiredTests,
		})
		if err != nil {
			report.Issues = append(report.Issues, task.ID+": "+err.Error())
			continue
		}
		report.Fixtures[task.ID] = qualified
	}
	report.Ready = len(report.Issues) == 0 && runner.Started() == 0
	if *asJSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return err
		}
	} else {
		fmt.Fprintf(out, "conformance: ready=%t no_paid=%t runner=%s paid_calls=%d decision=%s issues=%d\n",
			report.Ready, report.NoPaid, report.Runner, report.PaidCalls, report.ReleaseDecision, len(report.Issues))
	}
	if *requireReady && !report.Ready {
		return fmt.Errorf("conformance is not ready: %s", strings.Join(report.Issues, "; "))
	}
	return nil
}

func runList(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	path := fs.String("config", defaultConfig, "benchmark config")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("list accepts no positional arguments")
	}
	cfg, err := loadConfig(*path)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(cfg)
}

func runBenchmark(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	configPath := fs.String("config", defaultConfig, "benchmark config")
	mode := fs.String("mode", "", "direct or amr")
	taskID := fs.String("task", "", "task id")
	modelAlias := fs.String("model", "", "direct model alias")
	attemptAlias := fs.String("attempt", "", "AMR attempt model alias")
	driverAlias := fs.String("driver", "", "AMR correction/driver model alias")
	modelRunner := fs.String("model-runner", "ghcp", "direct runner: ghcp or byok")
	attemptRunner := fs.String("attempt-runner", "ghcp", "attempt runner: ghcp or byok")
	driverRunner := fs.String("driver-runner", "ghcp", "driver runner: ghcp or byok")
	modelProfile := fs.String("model-profile", "", "direct user-local BYOK profile")
	attemptProfile := fs.String("attempt-profile", "", "attempt user-local BYOK profile")
	driverProfile := fs.String("driver-profile", "", "driver user-local BYOK profile")
	approvalPath := fs.String("approval", "", "attended approval receipt bound to the exact matrix")
	experimentID := fs.String("experiment-id", "", "durable experiment budget identity")
	contextVariant := fs.String("context", "baseline", "frozen context contract: baseline or minimal")
	routesPath := fs.String("routes", "", "observed runtime route availability/tier catalog")
	dryRun := fs.Bool("dry-run", false, "validate and emit the approval template without reserving credits or loading profiles")
	repeat := fs.Int("repeat", 1, "number of independent runs")
	outRoot := fs.String("out", ".kb/amr-model-benchmark", "result root")
	profilesPath := fs.String("profiles", "", "user-local BYOK profiles")
	maxCredits := fs.Int("max-ai-credits", 0, "per GHCP invocation credit ceiling (must not exceed config)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("run accepts no positional arguments")
	}
	if *mode != "direct" && *mode != "amr" {
		return errors.New("--mode must be direct or amr")
	}
	if *taskID == "" || *repeat < 1 {
		return errors.New("--task and positive --repeat are required")
	}
	if *mode == "direct" && *modelAlias == "" {
		return errors.New("direct mode requires --model")
	}
	if *mode == "amr" && (*attemptAlias == "" || *driverAlias == "") {
		return errors.New("amr mode requires --attempt and --driver")
	}
	if !*dryRun {
		return errors.New("attended execution is disabled until a trusted human-approval verifier is implemented; use --dry-run")
	}
	cfg, err := loadConfig(*configPath)
	if err != nil {
		return err
	}
	configureSandboxImages(cfg.Sandbox)
	if cfg.Budget.PerCallAIU <= 0 || cfg.Budget.PerArmAIU < cfg.Budget.PerCallAIU ||
		cfg.Budget.ExperimentAIU < cfg.Budget.PerArmAIU {
		return errors.New("benchmark config requires increasing positive per-call, per-arm, and experiment credit ceilings")
	}
	if *maxCredits == 0 {
		*maxCredits = cfg.Budget.PerCallAIU
	}
	if *maxCredits < 1 || *maxCredits > cfg.Budget.PerCallAIU {
		return fmt.Errorf("--max-ai-credits must be between 1 and configured per-call ceiling %d", cfg.Budget.PerCallAIU)
	}
	task, ok := findTask(cfg, *taskID)
	if !ok {
		return fmt.Errorf("unknown task %q", *taskID)
	}
	if !task.AMREligible {
		return fmt.Errorf("task %q is AMR-ineligible: %s", task.ID, task.IneligibleWhy)
	}
	if err := preflightProof(task.Verify); err != nil {
		return fmt.Errorf("benchmark proof unavailable before dispatch: %w", err)
	}
	if err := validateTaskAdmission(cfg, task); err != nil {
		return fmt.Errorf("benchmark admission failed before approval or dispatch: %w", err)
	}
	contextPath := cfg.ContextContracts.Baseline
	if *contextVariant == "minimal" {
		contextPath = cfg.ContextContracts.Minimal
	} else if *contextVariant != "baseline" {
		return errors.New("--context must be baseline or minimal")
	}
	contextHash, err := fileSHA256(contextPath)
	if err != nil {
		return err
	}
	selectedContext, err := loadContextPayload(contextPath)
	if err != nil {
		return err
	}
	proofClosureHash, err := hashStringMap(task.ProofHashes)
	if err != nil {
		return err
	}
	var direct, attempt, driver modelSpec
	routes, routesHash, err := loadRuntimeRouteCatalog(*routesPath)
	if err != nil {
		return err
	}
	if *mode == "direct" {
		direct, err = resolveRuntimeModel(cfg, routes, *modelAlias, *modelRunner, *modelProfile, task.PlannedTier)
		if err != nil {
			return err
		}
	} else {
		attempt, err = resolveRuntimeModel(cfg, routes, *attemptAlias, *attemptRunner, *attemptProfile, task.AttemptTier)
		if err != nil {
			return err
		}
		driver, err = resolveRuntimeModel(cfg, routes, *driverAlias, *driverRunner, *driverProfile, task.PlannedTier)
		if err != nil {
			return err
		}
		if attempt.Tier != task.AttemptTier {
			return errors.New("AMR attempt model must exactly match the task attempt tier")
		}

		if driver.Tier != task.PlannedTier {
			return errors.New("AMR driver model must exactly match the task planned tier")
		}
	}
	root, err := filepath.Abs(*outRoot)
	if err != nil {
		return err
	}
	expectedApproval := approvalReceipt{
		ExperimentID: *experimentID, Mode: *mode, TaskID: task.ID, Repeat: *repeat,
		ContextContractHash: contextHash,
		RouteCatalogHash:    routesHash,
		MaxAICreditsPerCall: *maxCredits, MaxAICreditsPerArm: cfg.Budget.PerArmAIU,
		MaxExperimentCredits: cfg.Budget.ExperimentAIU,
		DirectModel:          direct.Model, DirectRunner: direct.Runner, DirectProfile: direct.Profile,
		AttemptModel: attempt.Model, AttemptRunner: attempt.Runner, AttemptProfile: attempt.Profile,
		DriverModel: driver.Model, DriverRunner: driver.Runner, DriverProfile: driver.Profile,
	}
	if !experimentIDPattern.MatchString(*experimentID) {
		return errors.New("--experiment-id must be 1-64 safe identifier characters")
	}
	if *dryRun {
		configHash, err := fileSHA256(*configPath)
		if err != nil {
			return err
		}
		expectedApproval.SchemaVersion = 1
		expectedApproval.ConfigSHA256 = configHash
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(map[string]any{
			"ready": true, "paid_calls": 0, "profiles_loaded": false,
			"approval_template": expectedApproval,
		})
	}
	approval, approvalHash, err := validateApproval(*approvalPath, *configPath, expectedApproval, time.Now().UTC())
	if err != nil {
		return err
	}
	ledgerRoot, err := budgetLedgerRoot()
	if err != nil {
		return err
	}
	if err := reserveDurableBudget(ledgerRoot, approval, approvalHash); err != nil {
		return err
	}
	profiles := map[string]localProfile{}
	if direct.Runner == "byok" || attempt.Runner == "byok" || driver.Runner == "byok" {
		profiles, err = loadProfiles(*profilesPath)
		if err != nil {
			return err
		}
	}
	resultsPath := filepath.Join(root, "results.jsonl")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	var results []runResult
	budget := newCreditBudget(cfg.Budget.PerCallAIU, cfg.Budget.PerArmAIU, cfg.Budget.ExperimentAIU)
	for i := 0; i < *repeat; i++ {
		budget.BeginArm()
		result, err := executeRun(cfg, task, i, contextHash, proofClosureHash, *experimentID, approvalHash, routesHash, selectedContext, *mode, direct, attempt, driver, profiles, root, budget, *maxCredits)
		if err != nil {
			return err
		}
		results = append(results, result)
		if err := appendJSONLine(resultsPath, result); err != nil {
			return err
		}
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func resolveRuntimeModel(cfg config, routes runtimeRouteCatalog, modelID, runner, profile, tier string) (modelSpec, error) {
	_ = cfg
	if strings.TrimSpace(modelID) == "" {
		return modelSpec{}, errors.New("runtime model ID is required")
	}
	for _, route := range routes.Routes {
		if route.ModelID == modelID && route.Runner == runner && route.Profile == profile {
			if route.Tier != tier {
				return modelSpec{}, fmt.Errorf("runtime model %q is tier %q, expected %q", modelID, route.Tier, tier)
			}
			return modelSpec{Alias: modelID, Model: modelID, Runner: runner, Tier: tier, Profile: profile}, nil
		}
	}
	return modelSpec{}, fmt.Errorf("runtime model %q has no matching observed route evidence", modelID)
}

func executeRun(cfg config, task taskSpec, seed int, contextHash, proofClosureHash, experimentID, approvalHash, routeCatalogHash string, selectedContext contextPayload, mode string, direct, attempt, driver modelSpec, profiles map[string]localProfile, root string, budget *creditBudget, maxCredits int) (runResult, error) {
	_ = cfg
	start := time.Now()
	runID := start.UTC().Format("20060102T150405.000000000Z") + "-" + task.ID + "-" + mode
	runDir := filepath.Join(root, "runs", runID)
	workspace := filepath.Join(runDir, "workspace")
	if err := copyDir(task.Fixture, workspace); err != nil {
		return runResult{}, err
	}
	if err := initGit(workspace); err != nil {
		return runResult{}, err
	}
	result := runResult{
		SchemaVersion: 1, RunID: runID, Mode: mode, TaskID: task.ID, TaskFamily: task.Family,
		Seed: seed, ContextContractHash: contextHash, ProofClosureHash: proofClosureHash,
		ExperimentID: experimentID, ApprovalHash: approvalHash, RouteCatalogHash: routeCatalogHash,
		PlannedTier: task.PlannedTier, AttemptTier: task.AttemptTier, StartedAt: start.UTC().Format(time.RFC3339Nano),
		Workspace: workspace,
	}
	if mode == "direct" {
		result.DriverModel = direct.Alias
		if err := budget.Reserve("direct", maxCredits); err != nil {
			return result, err
		}
		phase, err := invokeModel(runDir, workspace, "direct", direct, profiles, directPromptWithContext(task, workspace, selectedContext), task, maxCredits)
		if err != nil {
			return result, err
		}
		result.Phases = append(result.Phases, phase)
		result.FinalProof = phase.Proof
		if !phase.Valid {
			result.Outcome = "invalid-direct"
		} else if phase.Proof.Passed {
			result.Outcome = "passed-direct"
		} else {
			result.Outcome = "failed-direct"
		}
	} else {
		result.AttemptModel, result.DriverModel = attempt.Alias, driver.Alias
		if err := budget.Reserve("attempt", maxCredits); err != nil {
			return result, err
		}
		first, err := invokeModel(runDir, workspace, "attempt", attempt, profiles, attemptPromptWithContext(task, workspace, selectedContext), task, maxCredits)
		if err != nil {
			return result, err
		}
		result.Phases = append(result.Phases, first)
		if !first.Valid {
			result.FinalProof, result.Outcome = first.Proof, "invalid-attempt"
		} else if first.Proof.Passed {
			result.FinalProof, result.Outcome = first.Proof, "passed-attempt"
		} else if phaseCanCorrect(first) {
			diff := gitOutput(workspace, "diff", "--no-ext-diff", "--unified=3")
			prompt := correctionPromptWithContext(task, workspace, diff, first.Proof.Output, selectedContext)
			if err := budget.Reserve("correction", maxCredits); err != nil {
				return result, err
			}
			second, err := invokeModel(runDir, workspace, "correction", driver, profiles, prompt, task, maxCredits)
			if err != nil {
				return result, err
			}

			result.Phases = append(result.Phases, second)
			result.FinalProof = second.Proof
			if !second.Valid {
				result.Outcome = "invalid-correction"
			} else if second.Proof.Passed {
				result.Outcome = "passed-correction"
			} else {
				result.Outcome = "failed-correction"
			}
		}
	}
	result.ChangedFiles = splitLines(gitOutput(workspace, "diff", "--name-only"))
	result.LinesAdded, result.LinesDeleted = diffStats(workspace)
	result.DurationMS = time.Since(start).Milliseconds()
	resultPath := filepath.Join(runDir, "result.json")
	if err := writeJSON(resultPath, result); err != nil {
		return result, err
	}
	return result, nil
}

func phaseCanCorrect(phase phaseResult) bool {
	return phase.Valid && !phase.Proof.Passed
}

func invokeModel(runDir, workspace, phase string, model modelSpec, profiles map[string]localProfile, prompt string, task taskSpec, maxCredits int) (phaseResult, error) {
	phaseDir := filepath.Join(runDir, phase)
	if err := os.MkdirAll(phaseDir, 0o755); err != nil {
		return phaseResult{}, err
	}
	stdoutPath := filepath.Join(phaseDir, "stdout.txt")
	stderrPath := filepath.Join(phaseDir, "stderr.txt")
	otelPath := filepath.Join(phaseDir, "otel.jsonl")
	statePath := filepath.Join(phaseDir, "state.json")
	if err := writeAtomicJSONFile(statePath, phaseState{
		SchemaVersion: 1, Phase: phase, Status: "starting", UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}); err != nil {
		return phaseResult{}, err
	}
	args := []string{
		"-p", prompt,
		"--model", model.Model,
		"--max-ai-credits", strconv.Itoa(maxCredits),
		"--max-autopilot-continues", "1",
		"--available-tools=",
		"--disable-builtin-mcps",
		"--no-custom-instructions",
		"--no-ask-user",
		"--disallow-temp-dir",
		"--silent",
	}
	cmd := exec.Command("copilot", args...)
	cmd.Dir = workspace
	env := cleanProviderEnv(os.Environ())
	env = append(env, "COPILOT_OTEL_FILE_EXPORTER_PATH="+otelPath)
	if model.Runner == "byok" {
		profile, ok := profiles[model.Profile]
		if !ok {
			return phaseResult{}, fmt.Errorf("local profile %q is unavailable; configure it in the user-local profiles file", model.Profile)
		}
		configuredEnv, err := applyProfile(env, profile)
		if err != nil {
			return phaseResult{}, err
		}
		env = configuredEnv
	}
	cmd.Env = env
	stdout := newCappedBuffer(1 << 20)
	stderr := newCappedBuffer(1 << 20)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	start := time.Now()
	if err := configureProcessTree(cmd); err != nil {
		return phaseResult{}, err
	}
	if err := cmd.Start(); err != nil {
		return phaseResult{}, err
	}
	tree, err := attachProcessTree(cmd)
	if err != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		return phaseResult{}, err
	}
	defer tree.Close()
	wait := make(chan error, 1)
	go func() { wait <- cmd.Wait() }()
	var runErr error
	select {
	case runErr = <-wait:
	case <-time.After(90 * time.Second):
		killErr := tree.Kill()
		select {
		case <-wait:
			runErr = errors.New("model invocation timed out")
			if killErr != nil {
				runErr = fmt.Errorf("model invocation timed out; terminate process tree: %w", killErr)
			}
		case <-time.After(10 * time.Second):
			_ = writeAtomicJSONFile(statePath, phaseState{
				SchemaVersion: 1, Phase: phase, Status: "indeterminate",
				Reason:    "model invocation timed out and process did not exit",
				UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
			})
			return phaseResult{}, fmt.Errorf("model invocation timed out and process did not exit")
		}
	}
	duration := time.Since(start)
	exitCode := exitCode(runErr)
	if err := writeAtomicJSONFile(statePath, phaseState{
		SchemaVersion: 1, Phase: phase, Status: "provider-finished", ExitCode: exitCode,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}); err != nil {
		return phaseResult{}, err
	}
	if writeErr := os.WriteFile(stdoutPath, stdout.Bytes(), 0o644); writeErr != nil {
		return phaseResult{}, writeErr
	}
	if writeErr := os.WriteFile(stderrPath, stderr.Bytes(), 0o644); writeErr != nil {
		return phaseResult{}, writeErr
	}
	usage, usageErr := parseOTel(otelPath)
	expectedModels := []string{model.Model}
	if model.Runner == "byok" {
		if profile, ok := profiles[model.Profile]; ok {
			expectedModels = append(expectedModels, profile.ModelID, profile.WireModel)
		}
	}
	modelMatch := usageErr == nil && modelsMatch(usage.ActualModels, expectedModels)
	invalidReason := ""
	switch {
	case runErr != nil:
		invalidReason = "provider process failed: " + runErr.Error()
	case stdout.Overflowed() || stderr.Overflowed():
		invalidReason = "provider output exceeded bounded capture"
	case usageErr != nil:
		invalidReason = "telemetry invalid: " + usageErr.Error()
	case usage.Calls < 1 || !usage.AICAvailable:
		invalidReason = "telemetry is incomplete or has no exact AIU"
	case !modelMatch:
		invalidReason = "observed model does not match requested route"
	}
	if invalidReason != "" {
		if err := writeAtomicJSONFile(statePath, phaseState{
			SchemaVersion: 1, Phase: phase, Status: "invalid", ExitCode: exitCode,
			Reason: invalidReason, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
		}); err != nil {
			return phaseResult{}, err
		}
		return phaseResult{
			Phase: phase, Model: model.Alias, RequestedModel: model.Model, ActualModels: usage.ActualModels,
			ModelMatch: modelMatch, Runner: model.Runner, ExitCode: exitCode, Valid: false,
			InvalidReason: invalidReason, DurationMS: duration.Milliseconds(), AIC: usage.AIC,
			AIUNano: usage.AIUNano, AICAvailable: usage.AICAvailable, InputTokens: usage.InputTokens,
			CacheReadTokens: usage.CacheReadTokens, CacheWriteTokens: usage.CacheWriteTokens,
			OutputTokens: usage.OutputTokens, Calls: usage.Calls,
			Proof:      proofResult{Passed: false, ExitCode: 2, Output: invalidReason},
			StdoutPath: stdoutPath, StderrPath: stderrPath, OTelPath: otelPath,
		}, nil
	}
	applyErr := applyDraftResponseAllowed(workspace, stdout.String(), task.MutablePaths)
	if applyErr == nil {
		applyErr = verifyProofClosure(workspace, task.ProofHashes)
	}
	proof := runProof(workspace, task.Verify, task.RequiredTests)
	if applyErr != nil {
		proof.Passed = false
		proof.Output = bounded("draft apply failed: "+applyErr.Error()+"\n"+proof.Output, 12000)
	}
	valid := applyErr == nil && proof.ExitCode != 2
	if !valid && proof.Output == "" {
		proof.Output = "phase invalid before or during proof"
	}
	result := phaseResult{
		Phase: phase, Model: model.Alias, RequestedModel: model.Model, ActualModels: usage.ActualModels,
		ModelMatch: modelMatch, Runner: model.Runner, ExitCode: exitCode, Valid: valid,
		InvalidReason: errorString(applyErr), DurationMS: duration.Milliseconds(), AIC: usage.AIC,
		AIUNano: usage.AIUNano, AICAvailable: usage.AICAvailable,
		InputTokens: usage.InputTokens, CacheReadTokens: usage.CacheReadTokens, CacheWriteTokens: usage.CacheWriteTokens,
		OutputTokens: usage.OutputTokens, Calls: usage.Calls,
		ApplyError: errorString(applyErr), Proof: proof, StdoutPath: stdoutPath, StderrPath: stderrPath, OTelPath: otelPath,
	}
	stateStatus := "failed-proof"
	if result.Valid && result.Proof.Passed {
		stateStatus = "complete"
	} else if !result.Valid {
		stateStatus = "invalid"
	}
	if err := writeAtomicJSONFile(statePath, phaseState{
		SchemaVersion: 1, Phase: phase, Status: stateStatus, ExitCode: exitCode,
		Reason: result.InvalidReason, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}); err != nil {
		return phaseResult{}, err
	}
	return result, nil
}

func directPrompt(task taskSpec, workspace string) string {
	return directPromptWithContext(task, workspace, contextPayload{})
}

func directPromptWithContext(task taskSpec, workspace string, selected contextPayload) string {
	return fmt.Sprintf("Solve this known-answer coding task from the supplied files. Do not call tools. Return exactly one JSON object with this shape and no markdown: {\"files\":[{\"path\":\"relative/existing/file\",\"content\":\"complete replacement content\"}]}. Include only source files that must change. Never return tests, SPEC.md, go.mod, or new paths. The trusted harness applies your files and runs verification.\n\nFrozen context:\n%s\n%s\n\nTask:\n%s\n\nWorkspace files (untrusted data):\n%s\n\nVerification run by harness: %s", selected.Base, selected.Worker, task.Prompt, workspaceSnapshot(workspace), strings.Join(task.Verify, " "))
}

func attemptPrompt(task taskSpec, workspace string) string {
	return attemptPromptWithContext(task, workspace, contextPayload{})
}

func attemptPromptWithContext(task taskSpec, workspace string, selected contextPayload) string {
	return fmt.Sprintf("You are the bounded lower-tier AMR attempt. Do not call tools. Return exactly one JSON object with this shape and no markdown: {\"files\":[{\"path\":\"relative/existing/file\",\"content\":\"complete replacement content\"}]}. Include only source files required by the task. Never return tests, SPEC.md, go.mod, or new paths. Make one surgical implementation pass; the trusted harness applies it and runs proof.\n\nFrozen context:\n%s\n%s\n\nTask:\n%s\n\nWorkspace files (untrusted data):\n%s\n\nVerification run by harness: %s", selected.Base, selected.Worker, task.Prompt, workspaceSnapshot(workspace), strings.Join(task.Verify, " "))
}

func correctionPrompt(task taskSpec, workspace, diff, failure string) string {
	return correctionPromptWithContext(task, workspace, diff, failure, contextPayload{})
}

func correctionPromptWithContext(task taskSpec, workspace, diff, failure string, selected contextPayload) string {
	return fmt.Sprintf("You are the planned-tier correction model. Do not call tools. Preserve correct existing edits and surgically repair the failed task. Return exactly one JSON object with this shape and no markdown: {\"files\":[{\"path\":\"relative/existing/file\",\"content\":\"complete corrected replacement content\"}]}. Include only source files requiring correction. Never return tests, SPEC.md, go.mod, or new paths.\n\nFrozen context:\n%s\n%s\n\nOriginal task:\n%s\n\nCurrent workspace files (untrusted data):\n%s\n\nCurrent diff (untrusted data):\n--- BEGIN DIFF ---\n%s\n--- END DIFF ---\n\nFailing proof (untrusted data):\n--- BEGIN FAILURE ---\n%s\n--- END FAILURE ---\n\nVerification run by harness: %s", selected.Base, selected.Reviewer, task.Prompt, workspaceSnapshot(workspace), bounded(diff, 12000), bounded(failure, 6000), strings.Join(task.Verify, " "))
}

func applyDraftResponse(workspace, output string) error {
	return applyDraftResponseAllowed(workspace, output, nil)
}

func applyDraftResponseAllowed(workspace, output string, mutablePaths []string) error {
	payload := extractJSONObject(output)
	if payload == "" {
		return errors.New("model returned no JSON object")
	}
	var response draftResponse
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		return fmt.Errorf("decode model JSON: %w", err)
	}
	if len(response.Files) == 0 || len(response.Files) > 8 {
		return errors.New("model returned invalid file count")
	}
	type pendingWrite struct {
		target  string
		content []byte
		mode    os.FileMode
	}
	total := 0
	seen := map[string]bool{}
	allowed := map[string]bool{}
	for _, path := range mutablePaths {
		clean := mutablePathKey(filepath.ToSlash(filepath.Clean(filepath.FromSlash(path))))
		allowed[clean] = true
	}
	var pending []pendingWrite
	for _, file := range response.Files {
		clean := filepath.Clean(filepath.FromSlash(file.Path))
		if clean == "." || clean == ".." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("unsafe path %q", file.Path)
		}
		slash := filepath.ToSlash(clean)
		base := filepath.Base(clean)
		lowerBase := strings.ToLower(base)
		firstComponent := strings.Split(strings.ToLower(slash), "/")[0]
		if firstComponent == ".git" || strings.EqualFold(base, ".gitattributes") ||
			strings.HasSuffix(lowerBase, "_test.go") || lowerBase == "verify.test.js" ||
			strings.EqualFold(base, "go.mod") || strings.EqualFold(base, "go.sum") || strings.EqualFold(base, "SPEC.md") {
			return fmt.Errorf("protected file %q", file.Path)
		}
		if len(allowed) > 0 && !allowed[mutablePathKey(slash)] {
			return fmt.Errorf("file is outside the mutable allowlist: %q", file.Path)
		}
		if seen[mutablePathKey(slash)] {
			return fmt.Errorf("duplicate file %q", file.Path)
		}
		seen[mutablePathKey(slash)] = true
		target := filepath.Join(workspace, clean)
		info, err := os.Lstat(target)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("file must already exist and be regular: %q", file.Path)
		}

		total += len(file.Content)
		if total > 1<<20 {
			return errors.New("model response exceeds file-content limit")
		}
		pending = append(pending, pendingWrite{target: target, content: []byte(file.Content), mode: info.Mode()})
	}
	originals := make(map[string][]byte, len(pending))
	for _, write := range pending {
		data, err := os.ReadFile(write.target)
		if err != nil {
			return err
		}
		originals[write.target] = data
	}
	var applied []pendingWrite
	for _, write := range pending {
		if err := writeAtomicFile(write.target, write.content, write.mode); err != nil {
			var rollbackErr error
			for _, prior := range applied {
				if restoreErr := writeAtomicFile(prior.target, originals[prior.target], prior.mode); restoreErr != nil && rollbackErr == nil {
					rollbackErr = restoreErr
				}
			}
			if rollbackErr != nil {
				return fmt.Errorf("apply failed: %v; rollback failed: %w", err, rollbackErr)
			}
			return err
		}
		applied = append(applied, write)
	}
	return nil
}

func mutablePathKey(path string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}

func extractJSONObject(value string) string {
	start := strings.Index(value, "{")
	if start < 0 {
		return ""
	}
	decoder := json.NewDecoder(strings.NewReader(value[start:]))
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return ""
	}
	return string(raw)
}

func preflightProof(command []string) error {
	if _, err := sandboxRuntime(); err != nil {
		return err
	}
	if _, _, err := sandboxCommand(command); err != nil {
		return err
	}
	return nil
}

func runProof(workspace string, command, requiredTests []string) proofResult {
	if len(command) == 0 {
		return proofResult{Passed: false, ExitCode: 2, Output: "missing verification command"}
	}
	runtime, err := sandboxRuntime()
	if err != nil {
		return proofResult{Passed: false, ExitCode: 2, Output: err.Error()}
	}
	image, inner, err := sandboxCommand(command)
	if err != nil {
		return proofResult{Passed: false, ExitCode: 2, Output: err.Error()}
	}
	if strings.Contains(workspace, ",") {
		return proofResult{Passed: false, ExitCode: 2, Output: "sandbox workspace path contains unsupported comma", SandboxImage: image}
	}
	containerName := fmt.Sprintf("amrbench-%d-%d", os.Getpid(), time.Now().UnixNano())
	args := proofContainerArgs(containerName, workspace, image, inner)
	cmd := exec.Command(runtime, args...)
	cmd.Env = containerRuntimeEnvironment(os.Environ())
	output := newCappedBuffer(2 << 20)
	cmd.Stdout, cmd.Stderr = output, output
	if err := configureProcessTree(cmd); err != nil {
		return proofResult{Passed: false, ExitCode: 2, Output: err.Error(), SandboxImage: image}
	}
	if err := cmd.Start(); err != nil {
		return proofResult{Passed: false, ExitCode: 2, Output: err.Error(), SandboxImage: image}
	}
	tree, err := attachProcessTree(cmd)
	if err != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		return proofResult{Passed: false, ExitCode: 2, Output: err.Error(), SandboxImage: image}
	}
	defer tree.Close()
	wait := make(chan error, 1)
	go func() { wait <- cmd.Wait() }()
	var runErr error
	select {
	case runErr = <-wait:
	case <-time.After(4 * time.Minute):
		killErr := tree.Kill()
		cleanupContext, cancelCleanup := context.WithTimeout(context.Background(), 15*time.Second)
		cleanup := exec.CommandContext(cleanupContext, runtime, "rm", "-f", containerName)
		cleanupErr := cleanup.Run()
		cancelCleanup()
		select {
		case <-wait:
			runErr = errors.New("proof timed out")
			if killErr != nil || cleanupErr != nil {
				runErr = fmt.Errorf("proof timed out; kill=%v cleanup=%v", killErr, cleanupErr)
			}
		case <-time.After(10 * time.Second):
			return proofResult{Passed: false, ExitCode: 2, Output: "proof timed out and process did not exit", SandboxImage: image}
		}
	}
	code := exitCode(runErr)
	if output.Overflowed() {
		runErr = fmt.Errorf("proof output exceeded bounded capture")
	}
	if runErr != nil && output.Len() == 0 {
		output.WriteString(runErr.Error())
	}
	passed := code == 0 && requiredTestEventsPresent(output.String(), requiredTests)
	if code == 0 && !passed {
		code = 3
		output.WriteString("\nrequired test-event evidence missing")
	}
	return proofResult{Passed: passed, ExitCode: code, Output: bounded(output.String(), 12000), SandboxImage: image}
}

func proofContainerArgs(containerName, workspace, image string, inner []string) []string {
	args := []string{
		"run", "--rm", "--pull", "never", "--network", "none", "--read-only",
		"--name", containerName,
		"--pids-limit", "256", "--memory", "1g", "--cpus", "2",
		"--mount", "type=bind,src=" + workspace + ",dst=/workspace,readonly",
		"--tmpfs", "/tmp:rw,noexec,nosuid,size=128m",
		"--tmpfs", "/cache:rw,noexec,nosuid,size=512m",
		// Go executes its generated test binary from GOTMPDIR.
		"--tmpfs", "/gotmp:rw,nosuid,size=256m",
		"-e", "HOME=/tmp", "-e", "GOCACHE=/cache", "-e", "GOTMPDIR=/gotmp",
		"-w", "/workspace", image,
	}
	return append(args, inner...)
}

func sandboxRuntime() (string, error) {
	for _, name := range []string{"docker", "podman"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	if programFiles := os.Getenv("ProgramFiles"); programFiles != "" {
		path := filepath.Join(programFiles, "RedHat", "Podman", "podman.exe")
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			return path, nil
		}
	}
	return "", errors.New("sandbox unavailable: install Docker/Podman or configure an equivalent network-disabled proof runner")
}

var configuredSandboxImages sandboxSpec

func configureSandboxImages(images sandboxSpec) {
	configuredSandboxImages = images
}

func sandboxCommand(command []string) (string, []string, error) {
	if len(command) == 0 {
		return "", nil, errors.New("missing sandbox command")
	}
	switch command[0] {
	case "go":
		if len(command) < 3 || command[1] != "test" {
			return "", nil, errors.New("only go test proofs are supported")
		}
		inner := append([]string{"go", "test", "-json", "-count=1", "-timeout=45s", "-p=1", "-vet=off"}, command[2:]...)
		image, err := pinnedSandboxImage("AMRBENCH_GO_IMAGE")
		return image, inner, err
	case "node":
		image, err := pinnedSandboxImage("AMRBENCH_NODE_IMAGE")
		return image, command, err
	default:
		return "", nil, fmt.Errorf("unsupported sandbox proof command %q", command[0])
	}
}

func pinnedSandboxImage(envName string) (string, error) {
	image := strings.TrimSpace(os.Getenv(envName))
	if image == "" {
		switch envName {
		case "AMRBENCH_GO_IMAGE":
			image = configuredSandboxImages.GoImage
		case "AMRBENCH_NODE_IMAGE":
			image = configuredSandboxImages.NodeImage
		}
	}
	if image == "" || !strings.Contains(image, "@sha256:") {
		return "", fmt.Errorf("%s must name an immutable image digest", envName)
	}
	return image, nil
}

func requiredTestEventsPresent(output string, required []string) bool {
	if len(required) == 0 {
		return true
	}
	passed := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		var event struct {
			Action string `json:"Action"`
			Test   string `json:"Test"`
		}
		if json.Unmarshal([]byte(line), &event) == nil && event.Action == "pass" && event.Test != "" {
			passed[event.Test] = true
		}
		for _, test := range required {
			if strings.HasPrefix(strings.TrimSpace(line), "ok ") && strings.Contains(line, test) {
				passed[test] = true
			}
		}
	}
	for _, test := range required {
		if !passed[test] {
			return false
		}
	}
	return true
}

func parseOTel(path string) (otelUsage, error) {
	file, err := os.Open(path)
	if err != nil {
		return otelUsage{}, err
	}
	defer file.Close()
	normalized, err := ghcpotel.Parse(file)
	if err != nil {
		return otelUsage{}, err
	}
	return otelUsage{
		AIC:              float64(normalized.AIUNano) / 1e9,
		AIUNano:          normalized.AIUNano,
		AICAvailable:     normalized.AIUAvailable,
		InputTokens:      normalized.InputTokens,
		OutputTokens:     normalized.OutputTokens,
		CacheReadTokens:  normalized.CacheReadTokens,
		CacheWriteTokens: normalized.CacheWriteTokens,
		Calls:            normalized.Calls,
		ActualModels:     normalized.ActualModels,
	}, nil
}

func runGrade(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("grade", flag.ContinueOnError)
	configPath := fs.String("config", defaultConfig, "benchmark config")
	resultsPath := fs.String("results", ".kb/amr-model-benchmark/results.jsonl", "results JSONL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("grade accepts no positional arguments")
	}
	cfg, err := loadConfig(*configPath)
	if err != nil {
		return err
	}
	file, err := os.Open(*resultsPath)
	if err != nil {
		return err
	}
	defer file.Close()
	type bucket struct {
		samples, passes int
		mismatches      int
		aic             []float64
	}
	buckets := map[string]*bucket{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		var result runResult
		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			return err
		}
		route := result.DriverModel
		if result.Mode == "amr" {
			route = result.AttemptModel + "->" + result.DriverModel
		}
		key := result.Mode + "\x00" + route + "\x00" + result.TaskFamily
		b := buckets[key]
		if b == nil {
			b = &bucket{}
			buckets[key] = b
		}
		b.samples++
		matched := true
		totalAIC := 0.0
		aicAvailable := true
		for _, phase := range result.Phases {
			matched = matched && phase.ModelMatch
			aicAvailable = aicAvailable && phase.AICAvailable
			totalAIC += phase.AIC
		}
		if !matched {
			b.mismatches++
		} else if result.FinalProof.Passed {
			b.passes++
		}
		if aicAvailable {
			b.aic = append(b.aic, totalAIC)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	rows := []gradeRow{}
	for key, b := range buckets {
		parts := strings.Split(key, "\x00")
		eligibleSamples := b.samples - b.mismatches
		rate := 0.0
		if eligibleSamples > 0 {
			rate = float64(b.passes) / float64(eligibleSamples)
		}
		status := "probation"
		if eligibleSamples >= cfg.Qualification.MinimumSamples {
			if rate >= cfg.Qualification.QualifyRate {
				status = "qualified"
			} else if rate < cfg.Qualification.SuspendRate {
				status = "suspended"
			} else {
				status = "probation"
			}
		}
		rows = append(rows, gradeRow{
			Mode: parts[0], Route: parts[1], TaskFamily: parts[2], Samples: b.samples, Passes: b.passes,
			PassRate: rate, Mismatches: b.mismatches, TotalAIC: sum(b.aic), MedianAIC: median(b.aic), Status: status,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Mode == rows[j].Mode {
			if rows[i].Route == rows[j].Route {
				return rows[i].TaskFamily < rows[j].TaskFamily
			}
			return rows[i].Route < rows[j].Route
		}
		return rows[i].Mode < rows[j].Mode
	})
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(rows)
}

func loadConfig(path string) (config, error) {
	var cfg config
	if err := readJSON(path, &cfg); err != nil {
		return cfg, err
	}
	if cfg.SchemaVersion != 1 || (len(cfg.Models) == 0 && len(cfg.Cohorts) == 0) || len(cfg.Tasks) == 0 {
		return cfg, errors.New("invalid benchmark config")
	}
	for _, task := range cfg.Tasks {
		if task.ID == "" || task.Family == "" || task.PlannedTier == "" || task.Fixture == "" || len(task.Verify) == 0 {
			return cfg, fmt.Errorf("invalid task %q", task.ID)
		}
		if task.AMREligible && (task.AttemptTier == "" || len(task.RequiredTests) == 0) {
			return cfg, fmt.Errorf("eligible task %q requires attempt_tier and required_tests", task.ID)
		}
		if !task.AMREligible && strings.TrimSpace(task.IneligibleWhy) == "" {
			return cfg, fmt.Errorf("ineligible task %q requires a reason", task.ID)
		}
	}
	return cfg, nil
}

func loadProfiles(path string) (map[string]localProfile, error) {
	if path == "" {
		account, err := user.Current()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(account.HomeDir, ".kb", "amr-bench-models.json")
	}
	result := map[string]localProfile{}
	var profiles localProfiles
	if err := readJSON(path, &profiles); err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}
	for _, profile := range profiles.Profiles {
		result[profile.Alias] = profile
	}
	return result, nil
}

func applyProfile(env []string, profile localProfile) ([]string, error) {
	if strings.TrimSpace(profile.BaseURL) == "" || strings.TrimSpace(profile.ModelID) == "" {
		return nil, fmt.Errorf("local profile %q requires base_url and model_id", profile.Alias)
	}
	env = append(env,
		"COPILOT_PROVIDER_BASE_URL="+profile.BaseURL,
		"COPILOT_PROVIDER_TYPE="+defaultString(profile.ProviderType, "openai"),
		"COPILOT_MODEL="+profile.ModelID,
	)
	if profile.WireModel != "" {
		env = append(env, "COPILOT_PROVIDER_WIRE_MODEL="+profile.WireModel)
	}
	if profile.WireAPI != "" {
		env = append(env, "COPILOT_PROVIDER_WIRE_API="+profile.WireAPI)
	}
	if profile.APIKeyEnv != "" {
		secret := os.Getenv(profile.APIKeyEnv)
		if secret == "" {
			return nil, fmt.Errorf("local profile %q requires environment variable %s", profile.Alias, profile.APIKeyEnv)
		}
		env = append(env, "COPILOT_PROVIDER_API_KEY="+secret)
	}
	if profile.BearerTokenEnv != "" {
		secret := os.Getenv(profile.BearerTokenEnv)
		if secret == "" {
			return nil, fmt.Errorf("local profile %q requires environment variable %s", profile.Alias, profile.BearerTokenEnv)
		}
		env = append(env, "COPILOT_PROVIDER_BEARER_TOKEN="+secret)
	}
	if profile.MaxPromptTokens > 0 {
		env = append(env, "COPILOT_PROVIDER_MAX_PROMPT_TOKENS="+strconv.Itoa(profile.MaxPromptTokens))
	}
	if profile.MaxOutputTokens > 0 {
		env = append(env, "COPILOT_PROVIDER_MAX_OUTPUT_TOKENS="+strconv.Itoa(profile.MaxOutputTokens))
	}
	return env, nil
}

func cleanProviderEnv(values []string) []string {
	prefixes := []string{"COPILOT_PROVIDER_", "COPILOT_MODEL="}
	out := values[:0]
	for _, value := range values {
		skip := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(value, prefix) {
				skip = true
				break
			}
		}
		if !skip {
			out = append(out, value)
		}
	}
	return out
}

func findModel(cfg config, alias string) (modelSpec, bool) {
	for _, model := range cfg.Models {
		if model.Alias == alias {
			return model, true
		}
	}
	return modelSpec{}, false
}

func findTask(cfg config, id string) (taskSpec, bool) {
	for _, task := range cfg.Tasks {
		if task.ID == id {
			return task, true
		}
	}
	return taskSpec{}, false
}

func tierRank(tier string) int {
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

func initGit(dir string) error {
	commands := [][]string{
		{"init", "--quiet"},
		{"config", "user.email", "amrbench@example.invalid"},
		{"config", "user.name", "AMR Bench"},
		{"add", "."},
		{"commit", "--quiet", "-m", "fixture baseline"},
	}
	for _, args := range commands {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, output)
		}
	}
	return nil
}

func copyDir(source, destination string) error {
	source, err := filepath.Abs(source)
	if err != nil {
		return err
	}
	return filepath.Walk(source, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		firstComponent := strings.Split(strings.ToLower(filepath.ToSlash(rel)), "/")[0]
		if firstComponent == ".git" || strings.EqualFold(filepath.Base(rel), ".gitattributes") {
			return fmt.Errorf("fixture Git control path is forbidden: %s", rel)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("fixture symlink is forbidden: %s", path)
		}
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("fixture special file is forbidden: %s", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

func diffStats(workspace string) (int, int) {
	output := gitOutput(workspace, "diff", "--numstat")
	added, deleted := 0, 0
	for _, line := range splitLines(output) {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		a, _ := strconv.Atoi(parts[0])
		d, _ := strconv.Atoi(parts[1])
		added += a
		deleted += d
	}
	return added, deleted
}

func gitOutput(workspace string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = workspace
	output, _ := cmd.CombinedOutput()
	return string(output)
}

func appendJSONLine(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(value)
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

func bounded(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "\n...[truncated]"
}

func splitLines(value string) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(value), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func workspaceSnapshot(root string) string {
	var sections []string
	total := 0
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				return nil
			}
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil || info.Size() > 64*1024 {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		section := fmt.Sprintf("--- FILE %s ---\n%s", filepath.ToSlash(rel), data)
		if total+len(section) > 40000 {
			return nil
		}
		sections = append(sections, section)
		total += len(section)
		return nil
	})
	sort.Strings(sections)
	return strings.Join(sections, "\n")
}

func stringValue(value any) string {
	if result, ok := value.(string); ok {
		return result
	}
	return ""
}

func numberValue(value any) float64 {
	if result, ok := value.(float64); ok {
		return result
	}
	return 0
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func modelsMatch(actual, expected []string) bool {
	if len(actual) == 0 {
		return false
	}
	allowed := map[string]bool{}
	for _, model := range expected {
		if model != "" {
			allowed[model] = true
		}
	}
	for _, model := range actual {
		if !allowed[model] {
			return false
		}
	}
	return true
}

func sum(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	copyValues := append([]float64(nil), values...)
	sort.Float64s(copyValues)
	middle := len(copyValues) / 2
	if len(copyValues)%2 == 1 {
		return copyValues[middle]
	}
	return (copyValues[middle-1] + copyValues[middle]) / 2
}
