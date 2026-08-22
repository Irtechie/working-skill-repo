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
	"sync"
	"time"
)

const (
	defaultProcessCheckTimeout  = 20 * time.Minute
	processCheckTerminationWait = 5 * time.Second
	maxProcessCheckOutputBytes  = 1 << 20
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
  kbcheck delivery-state --delivery-receipt <path> [--json]
  kbcheck sense --check <path> [--trace <path>] [--root <path>]
  kbcheck trace-verify [--trace <path>] [--root <path>]
  kbcheck accept --check <path> [--trace <path>] [--root <path>]
  kbcheck proof-receipt-validate --receipt <path> [--root <path>] [--json]
  kbcheck proof-plan --registry <path> --receipt-dir <path> --request <check-id,...> [--root <path>] [--json]
  kbcheck proof-run --registry <path> --receipt-dir <path> --request <check-id,...> [--root <path>] [--json]
  kbcheck proof-governor-selftest [--root <path>]
  kbcheck learning-adoption --result-path <path> [--root <path>]
  kbcheck context-packet --packet <path> [--root <path>] [--json]
  kbcheck context-packet-selftest
  kbcheck graph-route --packet <path> [--root <path>] [--json]
  kbcheck graph-routing-lifecycle-selftest [--root <path>]
  kbcheck graph-routing-eval [--root <path>] [--require-ready] [--json]
  kbcheck execution-telemetry --telemetry <path> [--receipt <path> --evidence-envelope <path>] [--root <path>] [--json]
  kbcheck execution-telemetry-selftest
  kbcheck model-tier-eval --evidence <path> [--root <path>] [--json]
  kbcheck model-routing-release --cohort <name> --evidence <path> [--root <path>]
  kbcheck provider-hygiene [--root <path>] [--include-user] [--json]
  kbcheck provider-hygiene-selftest
  kbcheck slice-lease --action acquire|status|renew|release|recover [--root <path>] [--state-root <path>] [--json]
  kbcheck slice-lease-selftest
  kbcheck plan-run-lease --action acquire|status|renew|expand|release|recover [--run-id <id>] [--manifest <path>] [--root <path>] [--state-root <path>] [--json]
  kbcheck plan-run-lease-selftest
  kbcheck plan-worktree --action prepare|status|advance|complete|release --manifest <path> --owner-token <token> [--commit-authorized --commit-authorized-by <actor> --commit-approval-ref <reference>] [--run-id <id>] [--worktree <path>] [--branch <integration-ref>] [--base-sha <sha>] [--root <path>] [--json]
  kbcheck plan-worktree-selftest [--root <path>]
  kbcheck worktree --legacy-slice-worktree --action prepare|status|integrate|release --slice-id <id> --run-id <id> --owner-token <token> [--worktree <path>] [--branch <name>] [--base-sha <sha>] [--root <path>] [--json]
  kbcheck terminal-cleanup --action register|sweep --session-id <current-project-session-id> [--work-id <id> --worktree <path> --branch <name> --commit-sha <sha> --delivery-mode local|pr|direct --remote <name> --claim-id <id> --provider <name> --pr-id <id> --pr-url <url> --resume-packet <path>] [--root <path>] [--json]
  kbcheck session-preserve --action plan|apply --session-id <current-project-session-id> [--worktree <path>] [--branch <expected-branch>] [--root <path>] [--json]
  kbcheck cargo-storage --action resolve|register-temp|finalize|validate-ready|not-applicable|validate --run-id <id> [--cache-root <path>] [--target <path> --temp-root <path> --reason <text>] [--root <path>] [--json]
  kbcheck scope-lease --ledger <path> [--json]
  kbcheck scope-lease-selftest
  kbcheck skill-lint [--root <path>] [--config <path>] [--json]
  kbcheck skill-guidance [--root <path>] [--config <path>]
  kbcheck skill-sync-report [--root <path>] [--config <path>] [--json] [--verbose-optional]
  kbcheck doctor [--root <path>] [--config <path>] [--fix] [--json]
  kbcheck doctor-selftest [--root <path>]
  kbcheck marketplace-firebreak [--root <path>] [--config <path>] [--json]
  kbcheck marketplace-firebreak-selftest [--root <path>] [--config <path>]
  kbcheck marketplace-promote --source <path> --approval-reason <text> --approved [--skill-id <id>] [--install-targets codex,copilot,agents] [--json]
  kbcheck marketplace-promote-selftest [--root <path>]
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
	slice-lease    Atomically acquire and release local slice ownership.
	plan-run-lease Atomically claim manifest paths, domains, and shared resources.
	plan-worktree  Prepare and inspect a manifest-owned plan-run workspace.
	plan-worktree-selftest  Exercise two isolated plan runs in a disposable repository.
	worktree       Deprecated compatibility command for legacy isolated slice worktrees.
	terminal-cleanup  Register and safely reap durably delivered terminal worktrees.
	graph-route    Validate provider-neutral graph/evidence impact packets.
	graph-routing-lifecycle-selftest  Validate graph routing lifecycle invariants.
	graph-routing-eval  Score graph routing correctness and local concurrency safety fixtures.
	doctor        Report or repair configured skill install drift.
`

type processRunner func(root string, check Check) CheckResult

type options struct {
	command                 string
	root                    string
	json                    bool
	dryRun                  bool
	verbose                 bool
	list                    bool
	manifest                string
	ledger                  string
	config                  string
	verboseOptional         bool
	fix                     bool
	fixtureRoot             string
	route                   string
	baseline                string
	output                  string
	skillRoot               string
	agentRoot               string
	trimLineThreshold       int
	start                   string
	status                  bool
	runID                   string
	resultRoot              string
	resultPath              string
	requiredRunID           string
	manifestPath            string
	updateBaseline          bool
	qualityRoot             string
	qualityPath             string
	claimRoot               string
	claimPath               string
	minScore                int
	runRoot                 string
	fixtureID               string
	all                     bool
	keepRun                 bool
	sealed                  bool
	runner                  string
	runtime                 string
	model                   string
	agentCommand            string
	source                  string
	skillID                 string
	approvalReason          string
	approvedBy              string
	sourceType              string
	upstreamRepo            string
	installTargets          string
	gate                    string
	allowedNext             string
	history                 string
	checkPath               string
	tracePath               string
	packetPath              string
	telemetryPath           string
	receiptPath             string
	deliveryReceiptPath     string
	receiptDir              string
	proofRegistryPath       string
	proofRequest            string
	evidenceEnvelopePath    string
	cohort                  string
	evidencePath            string
	allowQuarantine         bool
	sliceLeaseAction        string
	sliceLeaseStateRoot     string
	sliceID                 string
	ownerToken              string
	leaseGeneration         int64
	leaseTTL                time.Duration
	leaseFiles              []string
	leasePrefixes           []string
	leaseDomains            []string
	leaseResources          []string
	baseSHA                 string
	worktreePath            string
	branchName              string
	repoIdentity            string
	expectedIntegrationHead string
	commitSHA               string
	proofReceipt            string
	codexSkillsRoot         string
	copilotSkillsRoot       string
	agentsSkillsRoot        string
	approved                bool
	includeUser             bool
	requireReady            bool
	commitAuthorized        bool
	commitAuthorizedBy      string
	commitApprovalRef       string
	legacySliceWorktree     bool
	workID                  string
	claimID                 string
	sessionID               string
	deliveryMode            string
	remote                  string
	resumePacket            string
	provider                string
	pullRequestID           string
	pullRequestURL          string
	cargoCacheRoot          string
	cargoTarget             string
	cargoTempRoot           string
	cargoReason             string
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
	case "delivery-state":
		return runDeliveryStateCommand(root, opts, stdout, stderr)
	case "sense":
		return runProofSenseCommand(root, opts, stdout, stderr)
	case "trace-verify":
		return runProofTraceVerifyCommand(root, opts, stdout, stderr)
	case "accept":
		return runProofAcceptCommand(root, opts, stdout, stderr)
	case "proof-receipt-validate":
		return runProofGovernorReceiptValidateCommand(root, opts, stdout, stderr)
	case "proof-plan":
		return runProofGovernorPlanCommand(root, opts, stdout, stderr)
	case "proof-run":
		return runProofGovernorExecuteCommand(root, opts, stdout, stderr)
	case "proof-governor-selftest":
		return runProofGovernorSelftest(root, stdout, stderr)
	case "learning-adoption":
		return runLearningAdoptionCommand(root, opts, stdout, stderr)
	case "context-packet":
		return runContextPacketCommand(root, opts, stdout, stderr)
	case "context-packet-selftest":
		return runContextPacketSelftest(stdout, stderr)
	case "graph-route":
		return runGraphRouteCommand(root, opts, stdout, stderr)
	case "graph-routing-lifecycle-selftest":
		return runGraphRoutingLifecycleSelftest(root, stdout, stderr)
	case "graph-routing-eval":
		return runGraphRoutingEvalCommand(root, opts, stdout, stderr)
	case "execution-telemetry":
		return runExecutionTelemetryCommand(root, opts, stdout, stderr)
	case "execution-telemetry-selftest":
		return runExecutionTelemetrySelftest(stdout, stderr)
	case "model-tier-eval":
		return runModelTierEvalCommand(root, opts, stdout, stderr)
	case "model-routing-release":
		return runModelRoutingReleaseCommand(root, opts, stdout, stderr)
	case "provider-hygiene":
		return runProviderHygieneCommand(root, opts, stdout, stderr)
	case "provider-hygiene-selftest":
		return runProviderHygieneSelftest(stdout, stderr)
	case "slice-lease":
		return runSliceLeaseCommand(root, opts, stdout, stderr)
	case "slice-lease-selftest":
		return runSliceLeaseSelftest(stdout, stderr)
	case "plan-run-lease":
		return runPlanRunLeaseCommand(root, opts, stdout, stderr)
	case "plan-run-lease-selftest":
		return runPlanRunLeaseSelftest(stdout, stderr)
	case "plan-worktree":
		return runPlanRunWorkspaceCommand(root, opts, stdout, stderr)
	case "plan-worktree-selftest":
		return runPlanWorktreeSelftest(root, stdout, stderr)
	case "worktree":
		return runWorktreeCommand(root, opts, stdout, stderr)
	case "terminal-cleanup":
		return runTerminalCleanupCommand(root, opts, stdout, stderr)
	case "session-preserve":
		return runSessionPreserveCommand(root, opts, stdout, stderr)
	case "cargo-storage":
		return runCargoStorageCommand(root, opts, stdout, stderr)
	case "scope-lease":
		return runScopeLeaseCommand(root, opts, stdout, stderr)
	case "scope-lease-selftest":
		return runScopeLeaseSelftest(stdout, stderr)
	case "skill-lint":
		return runSkillLintCommand(root, opts, stdout, stderr)
	case "skill-guidance":
		return runSkillGuidanceCommand(root, opts, stdout, stderr)
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
		"run-state": true, "run-state-selftest": true, "delivery-state": true,
		"sense": true, "trace-verify": true, "accept": true, "proof-receipt-validate": true, "proof-plan": true, "proof-run": true, "proof-governor-selftest": true, "learning-adoption": true,
		"context-packet": true, "context-packet-selftest": true, "graph-route": true, "graph-routing-lifecycle-selftest": true, "graph-routing-eval": true, "provider-hygiene": true, "provider-hygiene-selftest": true,
		"execution-telemetry": true, "execution-telemetry-selftest": true,
		"model-tier-eval": true, "model-routing-release": true,
		"slice-lease": true, "slice-lease-selftest": true, "plan-run-lease": true, "plan-run-lease-selftest": true, "plan-worktree": true, "plan-worktree-selftest": true, "worktree": true, "terminal-cleanup": true, "session-preserve": true, "cargo-storage": true,
		"scope-lease": true, "scope-lease-selftest": true,
		"skill-lint": true, "skill-guidance": true, "skill-sync-report": true, "doctor": true, "doctor-selftest": true,
		"marketplace-firebreak": true, "marketplace-firebreak-selftest": true,
		"marketplace-promote": true, "marketplace-promote-selftest": true,
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
	fs.StringVar(&opts.expectedIntegrationHead, "expected-integration-head", "", "expected plan-run integration head for compare-and-swap")
	fs.StringVar(&opts.commitSHA, "commit-sha", "", "slice commit to accept on the plan-run branch")
	fs.StringVar(&opts.proofReceipt, "proof-receipt", "", "coordinator-readable slice and aggregate proof receipt")
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
	fs.StringVar(&opts.packetPath, "packet", "", "context packet JSON path")
	fs.StringVar(&opts.telemetryPath, "telemetry", "", "execution telemetry JSON path")
	fs.StringVar(&opts.receiptPath, "receipt", "", "routing receipt JSON path")
	fs.StringVar(&opts.deliveryReceiptPath, "delivery-receipt", "", "delivery state receipt JSON path")
	fs.StringVar(&opts.receiptDir, "receipt-dir", "", "proof receipt directory")
	fs.StringVar(&opts.proofRegistryPath, "registry", "", "proof check registry JSON path")
	fs.StringVar(&opts.proofRequest, "request", "", "comma-separated proof check IDs")
	fs.StringVar(&opts.evidenceEnvelopePath, "evidence-envelope", "", "routing evidence envelope JSON path")
	fs.StringVar(&opts.cohort, "cohort", "", "model-routing release cohort")
	fs.StringVar(&opts.evidencePath, "evidence", "", "model evidence path")
	fs.BoolVar(&opts.allowQuarantine, "allow-quarantine", false, "accept status=quarantined as advanceable")
	registerSliceLeaseFlags(fs, &opts)
	home, _ := os.UserHomeDir()
	fs.StringVar(&opts.codexSkillsRoot, "codex-skills-root", filepath.Join(home, ".codex", "skills"), "Codex skills root")
	fs.StringVar(&opts.copilotSkillsRoot, "copilot-skills-root", filepath.Join(home, ".copilot", "skills"), "Copilot skills root")
	fs.StringVar(&opts.agentsSkillsRoot, "agents-skills-root", filepath.Join(home, ".agents", "skills"), "Agents skills root")
	fs.BoolVar(&opts.approved, "approved", false, "confirm human-approved marketplace promotion")
	fs.BoolVar(&opts.includeUser, "include-user", false, "include standard user-global provider configs")
	fs.BoolVar(&opts.requireReady, "require-ready", false, "fail when readiness thresholds are not met")
	fs.BoolVar(&opts.commitAuthorized, "commit-authorized", false, "record explicit authorization for local commits on the manifest-owned plan-run branch")
	fs.StringVar(&opts.commitAuthorizedBy, "commit-authorized-by", "", "actor that explicitly authorized local plan-run commits")
	fs.StringVar(&opts.commitApprovalRef, "commit-approval-ref", "", "durable reference to the local commit authorization")
	fs.BoolVar(&opts.legacySliceWorktree, "legacy-slice-worktree", false, "explicitly enable the deprecated per-slice worktree compatibility command")
	fs.StringVar(&opts.workID, "work-id", "", "stable KB work queue objective id")
	fs.StringVar(&opts.claimID, "claim-id", "", "stable awaiting-review lifecycle claim id")
	fs.StringVar(&opts.sessionID, "session-id", "", "Copilot project session id")
	fs.StringVar(&opts.deliveryMode, "delivery-mode", "", "delivery mode: local, pr, or direct")
	fs.StringVar(&opts.remote, "remote", "", "delivery remote name")
	fs.StringVar(&opts.resumePacket, "resume-packet", "", "versioned awaiting-review resume packet path")
	fs.StringVar(&opts.provider, "provider", "", "awaiting-review provider identity")
	fs.StringVar(&opts.pullRequestID, "pr-id", "", "awaiting-review pull request identity")
	fs.StringVar(&opts.pullRequestURL, "pr-url", "", "awaiting-review pull request URL")
	fs.StringVar(&opts.cargoCacheRoot, "cache-root", "", "stable Cargo cache root")
	fs.StringVar(&opts.cargoTarget, "target", "", "run-owned temporary Cargo target")
	fs.StringVar(&opts.cargoTempRoot, "temp-root", "", "approved parent for a run-owned temporary Cargo target")
	fs.StringVar(&opts.cargoReason, "reason", "", "technical isolation reason for a temporary Cargo target")
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
	manifestCommands := map[string]bool{"ready-set": true, "manifest-contract": true, "gate-ledger": true, "plan-worktree": true, "plan-run-lease": true}
	if !manifestCommands[opts.command] && opts.manifest != "" {
		return options{}, fmt.Errorf("--manifest is only supported for manifest commands")
	}
	if opts.command != "run-state" && opts.history != "" {
		return options{}, fmt.Errorf("--history is only supported for run-state")
	}
	if opts.command == "run-state" && opts.history == "" {
		return options{}, fmt.Errorf("run-state requires --history")
	}
	if opts.command != "delivery-state" && opts.deliveryReceiptPath != "" {
		return options{}, fmt.Errorf("--delivery-receipt is only supported for delivery-state")
	}
	if opts.command == "delivery-state" && opts.deliveryReceiptPath == "" {
		return options{}, fmt.Errorf("delivery-state requires --delivery-receipt")
	}
	if opts.command != "scope-lease" && opts.ledger != "" {
		return options{}, fmt.Errorf("--ledger is only supported for scope-lease")
	}
	if opts.config != "" && opts.command != "skill-lint" && opts.command != "skill-guidance" && opts.command != "skill-sync-report" && opts.command != "marketplace-firebreak" && opts.command != "marketplace-firebreak-selftest" && opts.command != "marketplace-promote" && opts.command != "review-reference-guard" {
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
	if opts.command == "plan-worktree" && opts.manifest == "" {
		return options{}, fmt.Errorf("plan-worktree requires --manifest")
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
	leaseFlagCommand := opts.command == "slice-lease" || opts.command == "plan-run-lease" || opts.command == "worktree" || opts.command == "plan-worktree" || opts.command == "terminal-cleanup" || opts.command == "session-preserve" || opts.command == "cargo-storage"
	if !leaseFlagCommand && (opts.sliceLeaseAction != "" || opts.sliceLeaseStateRoot != "" || opts.sliceID != "" || opts.ownerToken != "" || opts.leaseGeneration != 0 || opts.leaseTTL != defaultSliceLeaseTTL || len(opts.leaseFiles) > 0 || len(opts.leasePrefixes) > 0 || len(opts.leaseDomains) > 0 || len(opts.leaseResources) > 0 || opts.baseSHA != "" || opts.worktreePath != "" || opts.branchName != "" || opts.repoIdentity != "") {
		return options{}, fmt.Errorf("slice/worktree flags are only supported for slice-lease and worktree")
	}
	if leaseFlagCommand && opts.sliceLeaseAction == "" {
		return options{}, fmt.Errorf("%s requires --action", opts.command)
	}
	if opts.command == "worktree" && (opts.sliceLeaseStateRoot != "" || opts.leaseGeneration != 0 || opts.leaseTTL != defaultSliceLeaseTTL || len(opts.leaseFiles) > 0 || len(opts.leasePrefixes) > 0 || len(opts.leaseDomains) > 0 || len(opts.leaseResources) > 0 || opts.repoIdentity != "") {
		return options{}, fmt.Errorf("slice lease state and claim flags are only supported for slice-lease")
	}
	if opts.command == "plan-worktree" && (opts.sliceLeaseStateRoot != "" || opts.leaseGeneration != 0 || opts.leaseTTL != defaultSliceLeaseTTL || len(opts.leaseFiles) > 0 || len(opts.leasePrefixes) > 0 || len(opts.leaseDomains) > 0 || len(opts.leaseResources) > 0 || opts.repoIdentity != "") {
		return options{}, fmt.Errorf("slice lease state, identity, and claim flags are not supported for plan-worktree")
	}
	if opts.command != "plan-worktree" && (opts.expectedIntegrationHead != "" || opts.proofReceipt != "") {
		return options{}, fmt.Errorf("--expected-integration-head and --proof-receipt are only supported for plan-worktree")
	}
	if opts.command != "plan-worktree" && opts.command != "terminal-cleanup" && opts.commitSHA != "" {
		return options{}, fmt.Errorf("--commit-sha is only supported for plan-worktree and terminal-cleanup")
	}
	if opts.command != "plan-worktree" && (opts.commitAuthorized || opts.commitAuthorizedBy != "" || opts.commitApprovalRef != "") {
		return options{}, fmt.Errorf("commit authorization flags are only supported for plan-worktree")
	}
	if opts.command != "worktree" && opts.legacySliceWorktree {
		return options{}, fmt.Errorf("--legacy-slice-worktree is only supported for worktree")
	}
	if opts.command == "worktree" && !opts.legacySliceWorktree {
		return options{}, fmt.Errorf("worktree is deprecated and requires --legacy-slice-worktree; plan runs use plan-worktree")
	}
	if opts.command == "plan-worktree" {
		if (opts.commitAuthorized || opts.commitAuthorizedBy != "" || opts.commitApprovalRef != "") && opts.sliceLeaseAction != "prepare" {
			return options{}, fmt.Errorf("commit authorization flags are only supported for plan-worktree prepare")
		}
		if opts.sliceLeaseAction == "prepare" && opts.commitAuthorized &&
			(opts.commitAuthorizedBy == "" || opts.commitApprovalRef == "") {
			return options{}, fmt.Errorf("plan-worktree prepare with --commit-authorized requires --commit-authorized-by and --commit-approval-ref")
		}
		if opts.sliceLeaseAction == "advance" {
			if opts.runID == "" || opts.sliceID == "" || opts.expectedIntegrationHead == "" || opts.commitSHA == "" || opts.proofReceipt == "" || opts.worktreePath == "" || opts.branchName == "" {
				return options{}, fmt.Errorf("plan-worktree advance requires --run-id, --slice-id, --expected-integration-head, --commit-sha, --proof-receipt, --worktree, and --branch")
			}
		} else if opts.sliceID != "" || opts.expectedIntegrationHead != "" || opts.commitSHA != "" || opts.proofReceipt != "" {
			return options{}, fmt.Errorf("slice advance flags are only supported for plan-worktree advance")
		}
	}
	if opts.command == "terminal-cleanup" {
		if opts.sliceLeaseStateRoot != "" || opts.sliceID != "" || opts.ownerToken != "" ||
			opts.leaseGeneration != 0 || opts.leaseTTL != defaultSliceLeaseTTL ||
			len(opts.leaseFiles) > 0 || len(opts.leasePrefixes) > 0 ||
			len(opts.leaseDomains) > 0 || len(opts.leaseResources) > 0 ||
			opts.baseSHA != "" || opts.repoIdentity != "" {
			return options{}, fmt.Errorf("lease identity and claim flags are not supported for terminal-cleanup")
		}
		switch opts.sliceLeaseAction {
		case "register":
			if opts.workID == "" || opts.sessionID == "" || opts.worktreePath == "" ||
				opts.branchName == "" || opts.commitSHA == "" || opts.deliveryMode == "" {
				return options{}, fmt.Errorf("terminal-cleanup register requires --work-id, --session-id, --worktree, --branch, --commit-sha, and --delivery-mode")
			}
		case "sweep":
			if opts.sessionID == "" {
				return options{}, fmt.Errorf("terminal-cleanup sweep requires --session-id for current-session exclusion")
			}
			if opts.workID != "" || opts.worktreePath != "" ||
				opts.branchName != "" || opts.commitSHA != "" || opts.deliveryMode != "" ||
				opts.remote != "" || opts.resumePacket != "" || opts.claimID != "" ||
				opts.provider != "" || opts.pullRequestID != "" || opts.pullRequestURL != "" {
				return options{}, fmt.Errorf("terminal-cleanup sweep reads registered receipts and accepts only --session-id as cleanup identity")
			}
		default:
			return options{}, fmt.Errorf("terminal-cleanup action must be register or sweep")
		}
	} else if opts.command == "session-preserve" {
		if opts.sliceLeaseStateRoot != "" || opts.sliceID != "" || opts.ownerToken != "" ||
			opts.leaseGeneration != 0 || opts.leaseTTL != defaultSliceLeaseTTL ||
			len(opts.leaseFiles) > 0 || len(opts.leasePrefixes) > 0 ||
			len(opts.leaseDomains) > 0 || len(opts.leaseResources) > 0 ||
			opts.baseSHA != "" || opts.repoIdentity != "" {
			return options{}, fmt.Errorf("lease identity and claim flags are not supported for session-preserve")
		}
		switch opts.sliceLeaseAction {
		case "plan", "apply":
			if opts.sessionID == "" {
				return options{}, fmt.Errorf("session-preserve requires --session-id to attribute the preserved commit")
			}
		default:
			return options{}, fmt.Errorf("session-preserve action must be plan or apply")
		}
		if opts.workID != "" || opts.claimID != "" || opts.deliveryMode != "" ||
			opts.remote != "" || opts.resumePacket != "" || opts.provider != "" ||
			opts.pullRequestID != "" || opts.pullRequestURL != "" || opts.commitSHA != "" {
			return options{}, fmt.Errorf("session-preserve is a durability gate and accepts only --session-id, --worktree, and --branch as identity")
		}
	} else if opts.workID != "" || opts.claimID != "" || opts.sessionID != "" || opts.deliveryMode != "" ||
		opts.remote != "" || opts.resumePacket != "" || opts.provider != "" ||
		opts.pullRequestID != "" || opts.pullRequestURL != "" {
		return options{}, fmt.Errorf("work, session, delivery, PR identity, and resume-packet flags are only supported for terminal-cleanup and session-preserve")
	}
	if opts.command == "cargo-storage" {
		switch opts.sliceLeaseAction {
		case "resolve":
			if opts.runID == "" {
				return options{}, fmt.Errorf("cargo-storage resolve requires --run-id")
			}
			if opts.cargoTarget != "" || opts.cargoTempRoot != "" || opts.cargoReason != "" {
				return options{}, fmt.Errorf("cargo-storage resolve does not accept temporary-target flags")
			}
		case "register-temp":
			if opts.runID == "" || opts.cargoTarget == "" || opts.cargoTempRoot == "" || opts.cargoReason == "" {
				return options{}, fmt.Errorf("cargo-storage register-temp requires --run-id, --target, --temp-root, and --reason")
			}
		case "not-applicable":
			if opts.runID == "" || opts.cargoReason == "" {
				return options{}, fmt.Errorf("cargo-storage not-applicable requires --run-id and --reason")
			}
			if opts.cargoCacheRoot != "" || opts.cargoTarget != "" || opts.cargoTempRoot != "" {
				return options{}, fmt.Errorf("cargo-storage not-applicable accepts only --reason")
			}
		case "finalize", "validate-ready", "validate":
			if opts.runID == "" {
				return options{}, fmt.Errorf("cargo-storage %s requires --run-id", opts.sliceLeaseAction)
			}
			if opts.cargoCacheRoot != "" || opts.cargoTarget != "" || opts.cargoTempRoot != "" || opts.cargoReason != "" {
				return options{}, fmt.Errorf("cargo-storage %s reads the run receipt and accepts no path flags", opts.sliceLeaseAction)
			}
		default:
			return options{}, fmt.Errorf("cargo-storage action must be resolve, register-temp, finalize, validate-ready, not-applicable, or validate")
		}
	} else if opts.cargoCacheRoot != "" || opts.cargoTarget != "" || opts.cargoTempRoot != "" || opts.cargoReason != "" {
		return options{}, fmt.Errorf("--cache-root, --target, --temp-root, and --reason are only supported for cargo-storage")
	}
	if opts.command == "slice-lease" && len(opts.leaseDomains) > 0 {
		return options{}, fmt.Errorf("--domain is only supported for plan-run-lease")
	}
	if opts.command == "plan-run-lease" {
		if opts.sliceID != "" || opts.baseSHA != "" || opts.worktreePath != "" || opts.branchName != "" {
			return options{}, fmt.Errorf("slice and worktree identity flags are not supported for plan-run-lease")
		}
		if opts.runID == "" && opts.sliceLeaseAction != "status" {
			return options{}, fmt.Errorf("plan-run-lease requires --run-id")
		}
		if opts.sliceLeaseAction == "acquire" && opts.manifest == "" {
			return options{}, fmt.Errorf("plan-run-lease acquire requires --manifest")
		}
	}
	packetCommands := map[string]bool{"context-packet": true, "graph-route": true}
	if !packetCommands[opts.command] && opts.packetPath != "" {
		return options{}, fmt.Errorf("--packet is only supported for context-packet and graph-route")
	}
	if packetCommands[opts.command] && opts.packetPath == "" {
		return options{}, fmt.Errorf("%s requires --packet", opts.command)
	}
	receiptCommand := opts.command == "execution-telemetry" || opts.command == "proof-receipt-validate"
	if !receiptCommand && (opts.telemetryPath != "" || opts.receiptPath != "" || opts.evidenceEnvelopePath != "") {
		return options{}, fmt.Errorf("--telemetry, --receipt, and --evidence-envelope are only supported for receipt commands")
	}
	if opts.command == "execution-telemetry" && opts.telemetryPath == "" {
		return options{}, fmt.Errorf("execution-telemetry requires --telemetry")
	}
	if opts.command == "execution-telemetry" && ((opts.receiptPath == "") != (opts.evidenceEnvelopePath == "")) {
		return options{}, fmt.Errorf("execution-telemetry requires --receipt and --evidence-envelope together")
	}
	if opts.command == "proof-receipt-validate" && opts.receiptPath == "" {
		return options{}, fmt.Errorf("proof-receipt-validate requires --receipt")
	}
	if opts.command == "proof-receipt-validate" && (opts.telemetryPath != "" || opts.evidenceEnvelopePath != "") {
		return options{}, fmt.Errorf("proof-receipt-validate only accepts --receipt")
	}
	if opts.command != "proof-plan" && opts.command != "proof-run" && (opts.receiptDir != "" || opts.proofRegistryPath != "" || opts.proofRequest != "") {
		return options{}, fmt.Errorf("--receipt-dir, --registry, and --request are only supported for proof-plan or proof-run")
	}
	if (opts.command == "proof-plan" || opts.command == "proof-run") && (opts.receiptDir == "" || opts.proofRegistryPath == "" || opts.proofRequest == "") {
		return options{}, fmt.Errorf("%s requires --registry, --receipt-dir, and --request", opts.command)
	}
	if opts.command != "model-routing-release" && opts.cohort != "" {
		return options{}, fmt.Errorf("--cohort is only supported for model-routing-release")
	}
	if opts.command != "model-routing-release" && opts.command != "model-tier-eval" && opts.evidencePath != "" {
		return options{}, fmt.Errorf("--evidence is only supported for model-routing-release and model-tier-eval")
	}
	if opts.command == "model-routing-release" && (opts.cohort == "" || opts.evidencePath == "") {
		return options{}, fmt.Errorf("model-routing-release requires --cohort and --evidence")
	}
	if opts.command == "model-tier-eval" && opts.evidencePath == "" {
		return options{}, fmt.Errorf("model-tier-eval requires --evidence")
	}
	if opts.command != "provider-hygiene" && opts.includeUser {
		return options{}, fmt.Errorf("--include-user is only supported for provider-hygiene")
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
	timeout := check.Timeout
	if timeout <= 0 {
		timeout = defaultProcessCheckTimeout
	}
	cmd := exec.Command(check.Args[0], check.Args[1:]...)
	cmd.Dir = root
	if err := configureCheckProcessTree(cmd); err != nil {
		return CheckResult{ExitCode: 1, Stderr: fmt.Sprintf("configure process containment: %v", err)}
	}
	overflow := make(chan struct{}, 1)
	stdout := newCappedCheckBuffer(overflow)
	stderr := newCappedCheckBuffer(overflow)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return CheckResult{ExitCode: 1, Stderr: err.Error()}
	}
	tree, err := attachCheckProcessTree(cmd)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return CheckResult{ExitCode: 1, Stderr: fmt.Sprintf("attach process containment: %v", err)}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	exitReason := ""
	select {
	case err = <-done:
	case <-timer.C:
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
			if exitReason == "timeout" {
				return CheckResult{ExitCode: 124, Stderr: fmt.Sprintf("check timed out after %s and process tree did not exit within %s", timeout, processCheckTerminationWait)}
			}
			return CheckResult{ExitCode: 125, Stderr: fmt.Sprintf("check output exceeded %d bytes and process tree did not exit within %s", maxProcessCheckOutputBytes, processCheckTerminationWait)}
		}
	}
	_ = tree.Close()
	result := CheckResult{ExitCode: 0, Stdout: stdout.String(), Stderr: stderr.String()}
	if exitReason == "timeout" {
		result.ExitCode = 124
		result.Stderr = appendCheckDiagnostic(result.Stderr, fmt.Sprintf("check timed out after %s", timeout))
		return result
	}
	if exitReason == "overflow" || stdout.Truncated() || stderr.Truncated() {
		result.ExitCode = 125
		result.Stderr = appendCheckDiagnostic(result.Stderr, fmt.Sprintf("check output exceeded %d bytes", maxProcessCheckOutputBytes))
		return result
	}
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			return result
		}
		result.ExitCode = 1
		result.Stderr = err.Error()
		return result
	}
	return result
}

type cappedCheckBuffer struct {
	mu        sync.Mutex
	data      []byte
	truncated bool
	overflow  chan<- struct{}
}

func newCappedCheckBuffer(overflow chan<- struct{}) cappedCheckBuffer {
	return cappedCheckBuffer{overflow: overflow}
}

func (buffer *cappedCheckBuffer) Write(content []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	remaining := maxProcessCheckOutputBytes - len(buffer.data)
	if remaining > 0 {
		copyBytes := len(content)
		if copyBytes > remaining {
			copyBytes = remaining
		}
		buffer.data = append(buffer.data, content[:copyBytes]...)
	}
	if len(content) > remaining {
		buffer.truncated = true
		select {
		case buffer.overflow <- struct{}{}:
		default:
		}
	}
	return len(content), nil
}

func (buffer *cappedCheckBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return string(buffer.data)
}

func (buffer *cappedCheckBuffer) Truncated() bool {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.truncated
}

func appendCheckDiagnostic(existing, diagnostic string) string {
	if strings.TrimSpace(existing) == "" {
		return diagnostic
	}
	return strings.TrimRight(existing, "\r\n") + "\n" + diagnostic
}

func runNativeCommand(root string, args []string) CheckResult {
	var out, errOut strings.Builder
	fullArgs := append([]string{}, args...)
	fullArgs = append(fullArgs, "--root", root)
	code := run(fullArgs, &out, &errOut)
	return CheckResult{ExitCode: code, Stdout: out.String(), Stderr: errOut.String()}
}
