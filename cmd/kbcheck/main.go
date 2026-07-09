package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const usage = `kbcheck is the native KB gate entrypoint.

Usage:
  kbcheck core [--root <path>] [--list] [--dry-run] [--verbose]
  kbcheck local-release [--root <path>] [--json] [--dry-run]
  kbcheck live-release [--root <path>] [--json] [--dry-run]
  kbcheck ready-set --manifest <path> [--json]
  kbcheck ready-set-selftest
  kbcheck manifest-contract --manifest <path> [--json]
  kbcheck manifest-contract-selftest
  kbcheck gate-ledger --manifest <path> --gate <gate-id> [--allowed-next <text>] [--allow-quarantine]
  kbcheck run-state --history <path> [--json]
  kbcheck run-state-selftest
  kbcheck sense --check <path> [--trace <path>] [--root <path>]
  kbcheck trace-verify [--trace <path>] [--root <path>]
  kbcheck accept --check <path> [--trace <path>] [--root <path>]
  kbcheck learning-adoption --result-path <path> [--root <path>]
  kbcheck scope-lease --ledger <path> [--json]
  kbcheck scope-lease-selftest
  kbcheck skill-lint [--root <path>] [--config <path>] [--json]
  kbcheck skill-sync-report [--root <path>] [--config <path>] [--json] [--verbose-optional]
  kbcheck doctor [--root <path>] [--config <path>] [--fix] [--json]
  kbcheck doctor-selftest [--root <path>]
  kbcheck marketplace-firebreak [--root <path>] [--config <path>] [--json]
  kbcheck marketplace-firebreak-selftest [--root <path>] [--config <path>]
  kbcheck marketplace-promote --source <path> --approval-reason <text> --approved [--skill-id <id>] [--install-targets codex,copilot,agents] [--json]
  kbcheck marketplace-promote-selftest [--root <path>]
  kbcheck atv-delta [--atv-repo <path>] [--base-ref <ref>] [--upstream-ref <ref>] [--config <path>] [--json]
  kbcheck atv-delta-selftest
  kbcheck benchmark-validate [--root <path>] [--fixture-root <path>] [--json]
  kbcheck route-eval [--root <path>] [--config <path>] [--json]
  kbcheck dishonest-completion-selftest [--root <path>] [--fixture-root <path>]
  kbcheck review-reference-guard [--root <path>] [--config <path>] [--json]
  kbcheck release-selftest
  kbcheck workflow-governor-selftest [--root <path>]
  kbcheck surface-report [--root <path>] [--skill-root <path>] [--route <name>] [--baseline <path>] [--output <path>] [--json]
  kbcheck minimality [--root <path>] [--skill-root <path>] [--agent-root <path>] [--trim-line-threshold <n>] [--json]
  kbcheck minimality-selftest
  kbcheck pipeline [--root <path>] [--start <pipeline-id> | --status] [--run-id <id>]
  kbcheck pipeline-selftest [--root <path>]
  kbcheck skill-eval [--root <path>] [--result-root <path>] [--result-path <path>] [--baseline <path>] [--update-baseline] [--json]
  kbcheck skill-eval-claims [--root <path>] [--claim-root <path>] [--claim-path <path>] [--json]
  kbcheck skill-eval-quality [--root <path>] [--quality-root <path>] [--quality-path <path>] [--min-score <n>] [--json]
  kbcheck skill-eval-regression [--root <path>] [--run-root <path>] [--baseline <path>] [--output <path>] [--json]
  kbcheck skill-eval-manifest-selftest [--root <path>]
  kbcheck skill-eval-baseline-selftest [--root <path>]
  kbcheck eval-run-codex [--root <path>] [--fixture-id <id> | --all] [--dry-run] [--keep-run] [--json]
  kbcheck eval-run-ghcp [--root <path>] [--fixture-id <id> | --all] [--dry-run] [--keep-run] [--json]
  kbcheck eval-run-live-corpus [--root <path>] [--runtime codex,ghcp] [--dry-run] [--json]
  kbcheck skill-eval-wrap [--root <path>] [--runner <command>] [--fixture-id <id> | --all] [--dry-run] [--sealed] [--keep-run] [--json]
  kbcheck help

Commands:
  core           Discover and run local deterministic checks.
  local-release  Run the local release gate with required and optional checks.
  live-release   Run local release checks plus explicit live-model surfaces.
  ready-set      Compute the safe KB manifest ready set.
  manifest-contract  Validate KB manifest phase/gate proof contracts.
  gate-ledger    Validate one KB manifest gate before phase advancement.
  run-state      Validate KB route-history run state for oscillation and no-progress loops.
  sense          Run an objective check and append a hash-chained trace event.
  trace-verify   Verify the KB proof trace hash chain.
  accept         Prove a check went red->green and is green now.
  learning-adoption  Score held-out learning promotion eligibility.
  dishonest-completion-selftest  Validate false-done rejection fixtures.
  scope-lease    Validate observed active slice/file write leases.
  doctor        Report or repair configured skill install drift.
`

type processRunner func(root string, check Check) CheckResult

type options struct {
	command           string
	root              string
	json              bool
	dryRun            bool
	verbose           bool
	list              bool
	manifest          string
	ledger            string
	config            string
	verboseOptional   bool
	fix               bool
	fixtureRoot       string
	route             string
	baseline          string
	output            string
	skillRoot         string
	agentRoot         string
	trimLineThreshold int
	start             string
	status            bool
	runID             string
	resultRoot        string
	resultPath        string
	requiredRunID     string
	manifestPath      string
	updateBaseline    bool
	qualityRoot       string
	qualityPath       string
	claimRoot         string
	claimPath         string
	minScore          int
	runRoot           string
	fixtureID         string
	all               bool
	keepRun           bool
	sealed            bool
	runner            string
	runtime           string
	model             string
	agentCommand      string
	source            string
	skillID           string
	approvalReason    string
	approvedBy        string
	sourceType        string
	upstreamRepo      string
	installTargets    string
	gate              string
	allowedNext       string
	history           string
	checkPath         string
	tracePath         string
	allowQuarantine   bool
	codexSkillsRoot   string
	copilotSkillsRoot string
	agentsSkillsRoot  string
	approved          bool
	atvRepo           string
	baseRef           string
	upstreamRef       string
}

func main() {
	code := run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}

func run(args []string, stdout, stderr io.Writer) int {
	opts, err := parse(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		fmt.Fprintln(stderr)
		fmt.Fprint(stderr, usage)
		return 2
	}

	if opts.command == "help" {
		fmt.Fprint(stdout, usage)
		return 0
	}

	root, err := filepath.Abs(opts.root)
	if err != nil {
		fmt.Fprintf(stderr, "resolve root: %v\n", err)
		return 1
	}

	switch opts.command {
	case "core":
		return runCore(root, opts, stdout, stderr, runProcessCheck)
	case "local-release", "live-release":
		return runRelease(root, opts, stdout, stderr, runProcessCheck)
	case "ready-set":
		return runReadySetCommand(root, opts, stdout, stderr)
	case "ready-set-selftest":
		return runReadySetSelftest(stdout, stderr)
	case "manifest-contract":
		return runManifestContractCommand(root, opts, stdout, stderr)
	case "manifest-contract-selftest":
		return runManifestContractSelftest(stdout, stderr)
	case "gate-ledger":
		return runGateLedgerCommand(root, opts, stdout, stderr)
	case "run-state":
		return runRunStateCommand(root, opts, stdout, stderr)
	case "run-state-selftest":
		return runRunStateSelftest(stdout, stderr)
	case "sense":
		return runProofSenseCommand(root, opts, stdout, stderr)
	case "trace-verify":
		return runProofTraceVerifyCommand(root, opts, stdout, stderr)
	case "accept":
		return runProofAcceptCommand(root, opts, stdout, stderr)
	case "learning-adoption":
		return runLearningAdoptionCommand(root, opts, stdout, stderr)
	case "scope-lease":
		return runScopeLeaseCommand(root, opts, stdout, stderr)
	case "scope-lease-selftest":
		return runScopeLeaseSelftest(stdout, stderr)
	case "skill-lint":
		return runSkillLintCommand(root, opts, stdout, stderr)
	case "skill-sync-report":
		return runSkillSyncReportCommand(root, opts, stdout, stderr)
	case "doctor":
		return runDoctorCommand(root, opts, stdout, stderr)
	case "doctor-selftest":
		return runDoctorSelftest(root, stdout, stderr)
	case "marketplace-firebreak":
		return runMarketplaceFirebreakCommand(root, opts, stdout, stderr)
	case "marketplace-firebreak-selftest":
		return runMarketplaceFirebreakSelftest(root, opts, stdout, stderr)
	case "marketplace-promote":
		return runMarketplacePromoteCommand(root, opts, stdout, stderr)
	case "marketplace-promote-selftest":
		return runMarketplacePromoteSelftest(root, stdout, stderr)
	case "atv-delta":
		return runAtvDeltaCommand(root, opts, stdout, stderr)
	case "atv-delta-selftest":
		return runAtvDeltaSelftest(stdout, stderr)
	case "benchmark-validate":
		return runBenchmarkValidateCommand(root, opts, stdout, stderr)
	case "route-eval":
		return runRouteEvalCommand(root, opts, stdout, stderr)
	case "dishonest-completion-selftest":
		return runDishonestCompletionSelftest(root, opts, stdout, stderr)
	case "review-reference-guard":
		return runReviewReferenceGuardCommand(root, opts, stdout, stderr)
	case "release-selftest":
		return runReleaseSelftestCommand(stdout, stderr)
	case "workflow-governor-selftest":
		return runWorkflowGovernorSelftest(root, stdout, stderr)
	case "surface-report":
		return runSurfaceReportCommand(root, opts, stdout, stderr)
	case "minimality":
		return runMinimalityCommand(root, opts, stdout, stderr)
	case "minimality-selftest":
		return runMinimalitySelftest(stdout, stderr)
	case "pipeline":
		return runPipelineCommand(root, opts, stdout, stderr)
	case "pipeline-selftest":
		return runPipelineSelftest(root, stdout, stderr)
	case "skill-eval":
		return runSkillEvalCommand(root, opts, stdout, stderr)
	case "skill-eval-claims":
		return runSkillEvalClaimsCommand(root, opts, stdout, stderr)
	case "skill-eval-quality":
		return runSkillEvalQualityCommand(root, opts, stdout, stderr)
	case "skill-eval-regression":
		return runSkillEvalRegressionCommand(root, opts, stdout, stderr)
	case "skill-eval-manifest-selftest":
		return runSkillEvalManifestSelftest(root, stdout, stderr)
	case "skill-eval-baseline-selftest":
		return runSkillEvalBaselineSelftest(root, stdout, stderr)
	case "eval-run-codex":
		return runEvalAdapterCommand(root, opts, "codex", stdout, stderr)
	case "eval-run-ghcp":
		return runEvalAdapterCommand(root, opts, "ghcp", stdout, stderr)
	case "eval-run-live-corpus":
		return runEvalLiveCorpusCommand(root, opts, stdout, stderr)
	case "skill-eval-wrap":
		return runSkillEvalWrapCommand(root, opts, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unsupported command %q\n", opts.command)
		return 2
	}
}

func parse(args []string) (options, error) {
	if len(args) == 0 {
		return options{command: "help", root: "."}, nil
	}

	cmd := args[0]
	if cmd == "-h" || cmd == "--help" || cmd == "help" {
		return options{command: "help", root: "."}, nil
	}
	knownCommands := map[string]bool{
		"core": true, "local-release": true, "live-release": true,
		"ready-set": true, "ready-set-selftest": true, "manifest-contract": true, "manifest-contract-selftest": true, "gate-ledger": true,
		"run-state": true, "run-state-selftest": true,
		"sense": true, "trace-verify": true, "accept": true, "learning-adoption": true,
		"scope-lease": true, "scope-lease-selftest": true,
		"skill-lint": true, "skill-sync-report": true, "doctor": true, "doctor-selftest": true,
		"marketplace-firebreak": true, "marketplace-firebreak-selftest": true,
		"marketplace-promote": true, "marketplace-promote-selftest": true,
		"atv-delta": true, "atv-delta-selftest": true,
		"benchmark-validate": true, "route-eval": true, "dishonest-completion-selftest": true, "review-reference-guard": true, "release-selftest": true, "workflow-governor-selftest": true,
		"surface-report": true, "minimality": true, "minimality-selftest": true,
		"pipeline": true, "pipeline-selftest": true,
		"skill-eval": true, "skill-eval-claims": true, "skill-eval-quality": true, "skill-eval-regression": true,
		"skill-eval-manifest-selftest": true, "skill-eval-baseline-selftest": true,
		"eval-run-codex": true, "eval-run-ghcp": true, "eval-run-live-corpus": true, "skill-eval-wrap": true,
	}
	if !knownCommands[cmd] {
		return options{}, fmt.Errorf("unknown command %q", cmd)
	}

	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	opts := options{command: cmd, root: "."}
	fs.StringVar(&opts.root, "root", ".", "repository root")
	fs.BoolVar(&opts.json, "json", false, "emit JSON when supported")
	fs.BoolVar(&opts.dryRun, "dry-run", false, "print commands instead of running them")
	fs.BoolVar(&opts.verbose, "verbose", false, "print passing check output")
	fs.BoolVar(&opts.list, "list", false, "list checks without running them")
	fs.StringVar(&opts.manifest, "manifest", "", "KB manifest path")
	fs.StringVar(&opts.ledger, "ledger", "", "scope lease ledger path")
	fs.StringVar(&opts.config, "config", "", "validator config path")
	fs.BoolVar(&opts.verboseOptional, "verbose-optional", false, "print optional sync/report rows")
	fs.BoolVar(&opts.fix, "fix", false, "repair safe required skill install drift")
	fs.StringVar(&opts.fixtureRoot, "fixture-root", "", "fixture root path")
	fs.StringVar(&opts.route, "route", "", "route filter")
	fs.StringVar(&opts.baseline, "baseline", "", "baseline path")
	fs.StringVar(&opts.output, "output", "", "output path")
	fs.StringVar(&opts.skillRoot, "skill-root", "", "skill root path")
	fs.StringVar(&opts.agentRoot, "agent-root", "", "agent root path")
	fs.IntVar(&opts.trimLineThreshold, "trim-line-threshold", 250, "line threshold for trim candidates")
	fs.StringVar(&opts.start, "start", "", "pipeline id to start")
	fs.BoolVar(&opts.status, "status", false, "show pipeline status")
	fs.StringVar(&opts.runID, "run-id", "", "pipeline run id")
	fs.StringVar(&opts.resultRoot, "result-root", "", "skill eval result root")
	fs.StringVar(&opts.resultPath, "result-path", "", "skill eval result path")
	fs.StringVar(&opts.requiredRunID, "required-run-id", "", "required eval run id")
	fs.StringVar(&opts.manifestPath, "manifest-path", "", "eval run manifest path")
	fs.BoolVar(&opts.updateBaseline, "update-baseline", false, "update baseline file")
	fs.StringVar(&opts.qualityRoot, "quality-root", "", "quality fixture root")
	fs.StringVar(&opts.qualityPath, "quality-path", "", "single quality fixture path")
	fs.StringVar(&opts.claimRoot, "claim-root", "", "claim fixture root")
	fs.StringVar(&opts.claimPath, "claim-path", "", "single claim fixture path")
	fs.IntVar(&opts.minScore, "min-score", 3, "minimum quality score")
	fs.StringVar(&opts.runRoot, "run-root", "", "eval run root")
	fs.StringVar(&opts.fixtureID, "fixture-id", "", "route fixture id")
	fs.BoolVar(&opts.all, "all", false, "run all fixtures")
	fs.BoolVar(&opts.keepRun, "keep-run", false, "keep generated run directory")
	fs.BoolVar(&opts.sealed, "sealed", false, "block forbidden command attempts")
	fs.StringVar(&opts.runner, "runner", "", "wrapped runner command")
	fs.StringVar(&opts.runtime, "runtime", "", "runtime list")
	fs.StringVar(&opts.model, "model", "", "model name")
	fs.StringVar(&opts.agentCommand, "command", "", "agent CLI command")
	fs.StringVar(&opts.source, "source", "", "source skill path")
	fs.StringVar(&opts.skillID, "skill-id", "", "skill id")
	fs.StringVar(&opts.approvalReason, "approval-reason", "", "human approval reason")
	fs.StringVar(&opts.approvedBy, "approved-by", os.Getenv("USERNAME"), "human approver")
	fs.StringVar(&opts.sourceType, "source-type", "local-reviewed", "marketplace source type")
	fs.StringVar(&opts.upstreamRepo, "upstream-repo", "", "upstream repository or source URL")
	fs.StringVar(&opts.installTargets, "install-targets", "", "comma-separated install targets: codex,copilot,agents,none")
	fs.StringVar(&opts.gate, "gate", "", "gate_id to validate")
	fs.StringVar(&opts.allowedNext, "allowed-next", "", "expected allowed_next_action")
	fs.StringVar(&opts.history, "history", "", "route-history JSONL path")
	fs.StringVar(&opts.checkPath, "check", "", "objective proof check JSON path")
	fs.StringVar(&opts.tracePath, "trace", defaultProofTrace, "objective proof trace JSONL path")
	fs.BoolVar(&opts.allowQuarantine, "allow-quarantine", false, "accept status=quarantined as advanceable")
	home, _ := os.UserHomeDir()
	fs.StringVar(&opts.codexSkillsRoot, "codex-skills-root", filepath.Join(home, ".codex", "skills"), "Codex skills root")
	fs.StringVar(&opts.copilotSkillsRoot, "copilot-skills-root", filepath.Join(home, ".copilot", "skills"), "Copilot skills root")
	fs.StringVar(&opts.agentsSkillsRoot, "agents-skills-root", filepath.Join(home, ".agents", "skills"), "Agents skills root")
	fs.BoolVar(&opts.approved, "approved", false, "confirm human-approved marketplace promotion")
	fs.StringVar(&opts.atvRepo, "atv-repo", "../all-the-vibes", "ATV repository path")
	fs.StringVar(&opts.baseRef, "base-ref", "HEAD", "base git ref")
	fs.StringVar(&opts.upstreamRef, "upstream-ref", "upstream/main", "upstream git ref")
	if err := fs.Parse(args[1:]); err != nil {
		return options{}, err
	}
	if fs.NArg() > 0 {
		return options{}, fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	if opts.command == "core" && opts.json {
		return options{}, fmt.Errorf("--json is only supported for release gate and utility commands")
	}
	if opts.command != "core" && opts.list {
		return options{}, fmt.Errorf("--list is only supported for core")
	}
	if opts.command != "core" && opts.verbose {
		return options{}, fmt.Errorf("--verbose is only supported for core")
	}
	dryRunAllowed := map[string]bool{"core": true, "local-release": true, "live-release": true, "eval-run-codex": true, "eval-run-ghcp": true, "eval-run-live-corpus": true, "skill-eval-wrap": true}
	if !dryRunAllowed[opts.command] && opts.dryRun {
		return options{}, fmt.Errorf("--dry-run is only supported for gate commands")
	}
	manifestCommands := map[string]bool{"ready-set": true, "manifest-contract": true, "gate-ledger": true}
	if !manifestCommands[opts.command] && opts.manifest != "" {
		return options{}, fmt.Errorf("--manifest is only supported for manifest commands")
	}
	if opts.command != "run-state" && opts.history != "" {
		return options{}, fmt.Errorf("--history is only supported for run-state")
	}
	if opts.command == "run-state" && opts.history == "" {
		return options{}, fmt.Errorf("run-state requires --history")
	}
	if opts.command != "scope-lease" && opts.ledger != "" {
		return options{}, fmt.Errorf("--ledger is only supported for scope-lease")
	}
	if opts.config != "" && opts.command != "skill-lint" && opts.command != "skill-sync-report" && opts.command != "marketplace-firebreak" && opts.command != "marketplace-firebreak-selftest" && opts.command != "marketplace-promote" && opts.command != "atv-delta" && opts.command != "review-reference-guard" {
		return options{}, fmt.Errorf("--config is only supported for native validator commands")
	}
	if opts.verboseOptional && opts.command != "skill-sync-report" {
		return options{}, fmt.Errorf("--verbose-optional is only supported for skill-sync-report")
	}
	if opts.fix && opts.command != "doctor" {
		return options{}, fmt.Errorf("--fix is only supported for doctor")
	}
	if opts.command == "ready-set" && opts.manifest == "" {
		return options{}, fmt.Errorf("ready-set requires --manifest")
	}
	if opts.command == "manifest-contract" && opts.manifest == "" {
		return options{}, fmt.Errorf("manifest-contract requires --manifest")
	}
	if opts.command == "gate-ledger" && opts.manifest == "" {
		return options{}, fmt.Errorf("gate-ledger requires --manifest")
	}
	if opts.command != "gate-ledger" && (opts.gate != "" || opts.allowedNext != "" || opts.allowQuarantine) {
		return options{}, fmt.Errorf("--gate, --allowed-next, and --allow-quarantine are only supported for gate-ledger")
	}
	if opts.command == "gate-ledger" && opts.gate == "" {
		return options{}, fmt.Errorf("gate-ledger requires --gate")
	}
	proofCommands := map[string]bool{"sense": true, "trace-verify": true, "accept": true}
	if !proofCommands[opts.command] && (opts.checkPath != "" || opts.tracePath != defaultProofTrace) {
		return options{}, fmt.Errorf("--check and --trace are only supported for proof-spine commands")
	}
	if (opts.command == "sense" || opts.command == "accept") && opts.checkPath == "" {
		return options{}, fmt.Errorf("%s requires --check", opts.command)
	}
	if opts.command == "scope-lease" && opts.ledger == "" {
		return options{}, fmt.Errorf("scope-lease requires --ledger")
	}
	return opts, nil
}

func runCore(root string, opts options, stdout, stderr io.Writer, runner processRunner) int {
	checks, err := DiscoverChecks(root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.list || opts.dryRun {
		printChecks(stdout, checks)
		return 0
	}

	passed := 0
	for _, check := range checks {
		if opts.verbose {
			fmt.Fprintf(stdout, "==> %s: %s\n", check.Name, check.CommandString())
		}
		result := runner(root, check)
		if result.ExitCode == 0 && !opts.verbose {
			passed++
			fmt.Fprintf(stdout, "ok   %s\n", check.Name)
			continue
		}
		if result.ExitCode == 0 {
			passed++
		}
		if result.ExitCode != 0 {
			fmt.Fprintf(stderr, "FAIL %s: %s\n", check.Name, check.CommandString())
		}
		if result.Stdout != "" {
			fmt.Fprint(stdout, result.Stdout)
			if !strings.HasSuffix(result.Stdout, "\n") {
				fmt.Fprintln(stdout)
			}
		}
		if result.Stderr != "" {
			fmt.Fprint(stderr, result.Stderr)
			if !strings.HasSuffix(result.Stderr, "\n") {
				fmt.Fprintln(stderr)
			}
		}
		if result.ExitCode != 0 {
			fmt.Fprintf(stderr, "check failed: %s\n", check.Name)
			return result.ExitCode
		}
	}
	if !opts.verbose {
		fmt.Fprintf(stdout, "core: ok checks=%d\n", passed)
	}
	return 0
}

func printChecks(w io.Writer, checks []Check) {
	for _, check := range checks {
		fmt.Fprintf(w, "%-40s %s\n", check.Name, check.CommandString())
	}
}

func runProcessCheck(root string, check Check) CheckResult {
	if check.Run != nil {
		return check.Run(root)
	}
	if len(check.Args) == 0 {
		return CheckResult{ExitCode: 1, Stderr: "check has no command"}
	}
	cmd := exec.Command(check.Args[0], check.Args[1:]...)
	cmd.Dir = root
	out, err := cmd.Output()
	result := CheckResult{ExitCode: 0, Stdout: string(out)}
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			result.Stderr = string(exitErr.Stderr)
			return result
		}
		result.ExitCode = 1
		result.Stderr = err.Error()
		return result
	}
	return result
}

func runNativeCommand(root string, args []string) CheckResult {
	var out, errOut strings.Builder
	fullArgs := append([]string{}, args...)
	fullArgs = append(fullArgs, "--root", root)
	code := run(fullArgs, &out, &errOut)
	return CheckResult{ExitCode: code, Stdout: out.String(), Stderr: errOut.String()}
}
