package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

const defaultDispatchOutputLimit int64 = 64 * 1024
const defaultSessionEvidenceLimit int64 = 2 * 1024 * 1024
const dispatchRunLockName = ".kb-dispatch.lock"

type dispatchOptions struct {
	commonOptions
	runRoot            string
	runID              string
	sliceID            string
	packetPath         string
	outputPath         string
	receiptPath        string
	handoffPath        string
	workerRequestPath  string
	routeAlias         string
	fallbackRouteAlias string
	timeout            time.Duration
	outputLimit        int64
	attemptLimit       int
	sandbox            string
	approvalPolicy     string
	network            string
	allowedRoots       repeatFlag
	allowedTools       repeatFlag
}

type repeatFlag []string

func (f *repeatFlag) String() string { return strings.Join(*f, ",") }
func (f *repeatFlag) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	*f = append(*f, value)
	return nil
}

type dispatchReport struct {
	Status                string            `json:"status"`
	RouteAlias            string            `json:"route_alias"`
	PlannedTier           modelrouting.Tier `json:"planned_tier,omitempty"`
	AttemptTier           modelrouting.Tier `json:"attempt_tier,omitempty"`
	ProviderReportedModel string            `json:"provider_reported_model,omitempty"`
	SessionID             string            `json:"session_id,omitempty"`
	Attempt               int               `json:"attempt"`
	Attribution           string            `json:"attribution"`
	ReceiptPath           string            `json:"receipt_path,omitempty"`
	OutputPath            string            `json:"output_path,omitempty"`
	HandoffPath           string            `json:"handoff_path,omitempty"`
}

type dispatchPacket struct {
	SchemaVersion  int                            `json:"schema_version"`
	PacketID       string                         `json:"packet_id"`
	TaskID         string                         `json:"task_id"`
	RunID          string                         `json:"run_id"`
	SliceID        string                         `json:"slice_id"`
	ModelTier      modelrouting.Tier              `json:"model_tier"`
	AttemptTier    modelrouting.Tier              `json:"attempt_tier,omitempty"`
	TaskFamily     string                         `json:"task_family"`
	ContextSize    int                            `json:"context_size"`
	Risk           modelrouting.RiskLevel         `json:"risk"`
	AllowedTools   []string                       `json:"allowed_tools"`
	ProofTargets   []string                       `json:"proof_targets"`
	Redaction      map[string]any                 `json:"redaction"`
	BoundedContext bool                           `json:"bounded_context"`
	Correction     *modelrouting.CorrectionPacket `json:"correction,omitempty"`
}

type dispatchTrustedState struct {
	SchemaVersion      int               `json:"schema_version"`
	ProjectID          string            `json:"project_id"`
	RunID              string            `json:"run_id"`
	CatalogFingerprint string            `json:"catalog_fingerprint"`
	SurfaceRevision    string            `json:"surface_revision"`
	RouteStates        map[string]string `json:"route_states"`
	CreatedAt          time.Time         `json:"created_at"`
	ExpiresAt          time.Time         `json:"expires_at"`
}

type trustedExecutable struct {
	Path string
	Hash string
	Info os.FileInfo
}

type processResult struct {
	stdout             string
	stderr             string
	exitCode           int
	timeout            bool
	overflow           bool
	notStarted         bool
	containmentFailure bool
	err                error
}

var dispatchExecutableResolver = func() (string, error) {
	return exec.LookPath("codex")
}

var dispatchCodexHome = func() (string, error) {
	home, err := operatingSystemUserHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex"), nil
}

var dispatchProcessTreeContainment = ensureProcessTreeContainment

var dispatchTrustedStateProvider = func(userRoot, runRoot string) (dispatchTrustedState, error) {
	var state dispatchTrustedState
	var marker runRootMarker
	if err := modelrouting.LoadStrictJSON(runRoot, ".kb-run-root.json", &marker, maxCatalogBytes); err != nil {
		return dispatchTrustedState{}, fmt.Errorf("load run marker for dispatch state: %w", err)
	}
	catalog, err := modelrouting.LoadCatalog(runRoot, "catalog.json", modelrouting.StorageOptions{MaxBytes: maxCatalogBytes, Source: modelrouting.CatalogSourceRun})
	if err != nil {
		return dispatchTrustedState{}, fmt.Errorf("load run catalog for dispatch state: %w", err)
	}
	relDir, file := dispatchStateLocation(marker.ProjectID, marker.RunID)
	root := filepath.Join(userRoot, relDir)
	if err := modelrouting.LoadStrictJSON(root, file, &state, maxCatalogBytes); err != nil {
		return dispatchTrustedState{}, fmt.Errorf("load trusted dispatch state: %w", err)
	}
	if state.SchemaVersion != 1 || state.ProjectID != marker.ProjectID || state.RunID != marker.RunID || state.CatalogFingerprint != catalog.Fingerprint {
		return dispatchTrustedState{}, fmt.Errorf("trusted dispatch state does not match run catalog")
	}
	if state.ExpiresAt.IsZero() || !time.Now().Before(state.ExpiresAt) {
		return dispatchTrustedState{}, fmt.Errorf("trusted dispatch state expired")
	}
	if state.RouteStates == nil {
		state.RouteStates = map[string]string{}
	}
	return state, nil
}

func runDispatch(args []string, stdout, stderr io.Writer) int {
	fs := flagSet("dispatch")
	opts := dispatchOptions{timeout: 10 * time.Minute, outputLimit: defaultDispatchOutputLimit, attemptLimit: 2, sandbox: "workspace-write", approvalPolicy: "never", network: "none"}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.runRoot, "run-root", "", "marked KB run root")
	fs.StringVar(&opts.runID, "run-id", "", "KB run id")
	fs.StringVar(&opts.sliceID, "slice-id", "", "slice id")
	fs.StringVar(&opts.packetPath, "packet", "packet.json", "bounded context packet run-root child")
	fs.StringVar(&opts.outputPath, "output", "output.json", "redacted output summary run-root child")
	fs.StringVar(&opts.receiptPath, "receipt", "receipt.json", "routing receipt run-root child")
	fs.StringVar(&opts.handoffPath, "handoff", "handoff.json", "redacted partial handoff run-root child")
	fs.StringVar(&opts.workerRequestPath, "worker-request", "", "typed worker request run-root child")
	fs.StringVar(&opts.routeAlias, "route-alias", "", "route alias")
	fs.StringVar(&opts.fallbackRouteAlias, "fallback-route-alias", "", "fallback route alias")
	fs.DurationVar(&opts.timeout, "timeout", 10*time.Minute, "worker timeout")
	fs.Int64Var(&opts.outputLimit, "output-limit", defaultDispatchOutputLimit, "stdout/stderr byte bound")
	fs.IntVar(&opts.attemptLimit, "attempt-limit", 2, "finite route attempt ledger")
	fs.StringVar(&opts.sandbox, "sandbox", "workspace-write", "Codex sandbox policy")
	fs.StringVar(&opts.approvalPolicy, "approval-policy", "never", "Codex approval policy")
	fs.StringVar(&opts.network, "network", "none", "Codex workspace network policy: none or enabled")
	fs.Var(&opts.allowedRoots, "allowed-root", "allowed root; repeatable")
	fs.Var(&opts.allowedTools, "allowed-tool", "allowed tool; repeatable")
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(stderr, "unexpected argument %q\n", fs.Arg(0))
		return 2
	}
	if customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "credential/trust dispatch uses the fixed user-local trust root; custom --user-root is test-only")
		return 2
	}
	report, err := dispatchCodexWorker(opts)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return printResult(stdout, stderr, report, opts.json, nil)
}

func dispatchCodexWorker(opts dispatchOptions) (dispatchReport, error) {
	if opts.runRoot == "" || opts.runID == "" || opts.sliceID == "" || opts.routeAlias == "" {
		return dispatchReport{}, fmt.Errorf("dispatch requires run, slice, and route bindings")
	}
	if opts.outputLimit <= 0 || opts.outputLimit > maxCatalogBytes {
		return dispatchReport{}, fmt.Errorf("output limit must be between 1 and %d", maxCatalogBytes)
	}
	if opts.timeout <= 0 {
		return dispatchReport{}, fmt.Errorf("timeout must be positive")
	}
	if err := validateCodexPolicies(opts); err != nil {
		return dispatchReport{}, err
	}
	prepared, err := prepareRunRoot(opts.projectRoot, opts.runRoot)
	if err != nil {
		return dispatchReport{}, fmt.Errorf("prepare dispatch run root: %w", err)
	}
	if opts.runID != prepared.marker.RunID {
		return dispatchReport{}, fmt.Errorf("run id does not match prepared marker")
	}
	codexHome, err := dispatchCodexHome()
	if err != nil {
		return dispatchReport{}, fmt.Errorf("resolve CODEX_HOME: %w", err)
	}
	codexHome, err = filepath.Abs(filepath.Clean(codexHome))
	if err != nil {
		return dispatchReport{}, err
	}
	packetPath, err := safeRunChild(prepared, opts.packetPath, "packet.json")
	if err != nil {
		return dispatchReport{}, fmt.Errorf("packet path: %w", err)
	}
	outputPath, err := safeRunChild(prepared, opts.outputPath, "output.json")
	if err != nil {
		return dispatchReport{}, fmt.Errorf("output path: %w", err)
	}
	receiptPath, err := safeRunChild(prepared, opts.receiptPath, "receipt.json")
	if err != nil {
		return dispatchReport{}, fmt.Errorf("receipt path: %w", err)
	}
	handoffPath, err := safeRunChild(prepared, opts.handoffPath, "handoff.json")
	if err != nil {
		return dispatchReport{}, fmt.Errorf("handoff path: %w", err)
	}
	workerRequestPath := ""
	if opts.workerRequestPath != "" {
		workerRequestPath, err = safeRunChild(prepared, opts.workerRequestPath, "worker-request.json")
		if err != nil {
			return dispatchReport{}, fmt.Errorf("worker request path: %w", err)
		}
	}
	schemaName := "worker-output-schema-" + sha256Text(opts.sliceID)[:16] + ".json"
	resultSchemaPath, err := safeRunChild(prepared, schemaName, schemaName)
	if err != nil {
		return dispatchReport{}, err
	}
	if err := rejectDispatchRunLockArtifactTarget(packetPath, outputPath, receiptPath, handoffPath, workerRequestPath, resultSchemaPath); err != nil {
		return dispatchReport{}, err
	}
	dispatchLock, err := acquireDispatchRunLock(prepared, opts.userRoot, dispatchRunLockTimeout(opts))
	if err != nil {
		return dispatchReport{}, err
	}
	defer dispatchLock.release()
	if err := validateDispatchArtifactNamespace(prepared, opts.attemptLimit, packetPath, outputPath, receiptPath, handoffPath, workerRequestPath, resultSchemaPath); err != nil {
		return dispatchReport{}, err
	}
	packetData, err := readRunChild(prepared, packetPath, maxCatalogBytes)
	if err != nil {
		return dispatchReport{}, fmt.Errorf("read packet: %w", err)
	}
	packet, err := decodeDispatchPacket(packetData, prepared.marker.RunID, opts.sliceID)
	if err != nil {
		return dispatchReport{}, err
	}
	if packet.Correction != nil {
		return dispatchReport{}, fmt.Errorf("correction dispatch is non-executable until an isolated workspace and compare-and-swap apply runner are available")
	}
	attemptTier := packet.AttemptTier
	if attemptTier == "" {
		attemptTier = packet.ModelTier
	}
	packetHash := modelrouting.SHA256Bytes(packetData)
	projectID, err := modelrouting.CanonicalProjectIdentity(opts.projectRoot)
	if err != nil {
		return dispatchReport{}, fmt.Errorf("canonical project identity: %w", err)
	}
	hostState, err := dispatchTrustedStateProvider(opts.userRoot, prepared.runPath)
	if err != nil {
		return dispatchReport{}, err
	}
	trustedExec, err := resolveTrustedCodexExecutable(opts.projectRoot, prepared.runPath)
	if err != nil {
		return dispatchReport{}, err
	}
	if err := writeRunJSON(prepared, resultSchemaPath, workerResultSchema()); err != nil {
		return dispatchReport{}, fmt.Errorf("write output schema: %w", err)
	}
	attemptAliases := []string{opts.routeAlias}
	if opts.fallbackRouteAlias != "" {
		attemptAliases = append(attemptAliases, opts.fallbackRouteAlias)
	}
	ledger := modelrouting.NewAttemptLedger(opts.attemptLimit)
	var firstRoute modelrouting.Route
	var lastErr error
	var lastReport dispatchReport
	for index, alias := range attemptAliases {
		if err := ledger.Record(alias); err != nil {
			return lastReport, err
		}
		if err := trustedExec.revalidate(); err != nil {
			return lastReport, fmt.Errorf("trusted codex executable changed: %w", err)
		}
		hostState, err = dispatchTrustedStateProvider(opts.userRoot, prepared.runPath)
		if err != nil {
			return lastReport, err
		}
		catalog, policy, err := loadDispatchCatalog(prepared, opts, hostState)
		if err != nil {
			return lastReport, err
		}
		route, err := routeFromValidatedCatalog(catalog, policy, alias, packet)
		if err != nil {
			return lastReport, err
		}
		route = rehydrateTrustedRoute(route, policy.RouteSources)
		if err := validateExternalDispatchRoute(route); err != nil {
			return lastReport, err
		}
		if err := revalidateCodexAdapterRevisionForLaunch(context.Background(), trustedExec, route.AdapterRevision); err != nil {
			return lastReport, fmt.Errorf("trusted codex executable revision changed: %w", err)
		}
		if route.Profile != "" {
			revision, err := trustedCodexProfileRevision(codexHome, route.Profile, route.Endpoint, route.AuthEnv)
			if err != nil {
				return lastReport, err
			}
			if revision != route.ProfileRevision {
				return lastReport, fmt.Errorf("Codex profile revision changed")
			}
		}
		if index == 0 {
			firstRoute = route
		} else if !fallbackAllowed(firstRoute, route) {
			return lastReport, fmt.Errorf("downward fallback from %s to %s is refused", firstRoute.Alias, route.Alias)
		} else if !fallbackTrustTransitionAllowed(firstRoute, route, policy) {
			return lastReport, fmt.Errorf("less-trusted fallback from %s to %s requires existing approval", firstRoute.Alias, route.Alias)
		}
		req, err := modelrouting.NewDispatchRequest(route, projectID, opts.projectRoot, opts.runID, opts.sliceID, route.DisplayModelID, route.Profile, packetHash, index+1)
		if err != nil {
			return lastReport, err
		}
		req.PacketID = packet.PacketID
		if fingerprint, fingerprintErr := modelrouting.ApprovalRouteFingerprint(route, policy.RouteSources); fingerprintErr == nil {
			req.RouteFingerprint = fingerprint
		}
		attemptOutputPath := attemptArtifactPath(outputPath, req.Attempt)
		attemptReceiptPath := attemptArtifactPath(receiptPath, req.Attempt)
		attemptWorkerRequestPath := ""
		if workerRequestPath != "" {
			attemptWorkerRequestPath = attemptArtifactPath(workerRequestPath, req.Attempt)
		}
		req.Timeout, req.OutputPath, req.ReceiptPath, req.HandoffPath, req.WorkerRequestPath = opts.timeout, attemptOutputPath, attemptReceiptPath, handoffPath, attemptWorkerRequestPath
		req.Sandbox, req.ApprovalPolicy, req.Network = opts.sandbox, opts.approvalPolicy, opts.network
		req.AllowedRoots, err = canonicalAllowedRoots(opts.allowedRoots, opts.projectRoot)
		if err != nil {
			return lastReport, err
		}
		req.AllowedTools, err = canonicalAllowedTools(packet.AllowedTools)
		if err != nil {
			return lastReport, err
		}
		req.OutputSchemaPath = resultSchemaPath
		if err := modelrouting.ValidateDispatchRequest(req, route, codexHome, opts.projectRoot, policy.RouteSources); err != nil {
			if errors.Is(err, modelrouting.ErrUnsupportedDispatchAdapter) {
				return lastReport, fmt.Errorf("direct provider dispatch is not supported; configure a trusted Codex profile")
			}
			return lastReport, err
		}
		if workerRequestPath != "" {
			if err := writeRunJSON(prepared, attemptWorkerRequestPath, req); err != nil {
				return lastReport, fmt.Errorf("write worker request: %w", err)
			}
		}
		attemptStart := time.Now().UTC()
		result := runWorkerProcess(trustedExec.Path, req, packetData, opts.outputLimit, codexHome, route.AuthEnv)
		if result.notStarted {
			return dispatchReport{Status: "dispatch-unavailable", RouteAlias: req.RouteAlias, PlannedTier: packet.ModelTier, AttemptTier: attemptTier, Attempt: req.Attempt}, fmt.Errorf("dispatch unavailable before worker start: %w", result.err)
		}
		sessionID := parseThreadStartedSession(result.stdout)
		evidence := modelrouting.ProviderEvidence{}
		if sessionID != "" {
			evidence = loadCodexSessionEvidence(codexHome, sessionID, req, route, attemptStart)
		}
		attribution := modelrouting.ClassifyProviderEvidence(req.Model, evidence)
		outputStatus := "process-complete"
		if result.containmentFailure {
			outputStatus = "containment-failed"
		}
		output := map[string]any{
			"status":         outputStatus,
			"route_alias":    req.RouteAlias,
			"planned_tier":   packet.ModelTier,
			"attempt_tier":   attemptTier,
			"attempt":        req.Attempt,
			"exit_code":      result.exitCode,
			"timeout":        result.timeout,
			"output_bounded": result.overflow,
			"stdout_sha256":  modelrouting.SHA256Bytes([]byte(result.stdout)),
			"stderr_sha256":  modelrouting.SHA256Bytes([]byte(result.stderr)),
			"attribution":    string(attribution),
			"session_id":     evidence.SessionID,
			"actual_model":   evidence.Model,
		}
		if err := writeRunJSON(prepared, attemptOutputPath, output); err != nil {
			return lastReport, fmt.Errorf("write output summary: %w", err)
		}
		if req.Attempt == 1 {
			if err := writeRunJSON(prepared, outputPath, output); err != nil {
				return lastReport, fmt.Errorf("write output compatibility summary: %w", err)
			}
		}
		if result.containmentFailure {
			return dispatchReport{Status: "containment-failed", RouteAlias: req.RouteAlias, PlannedTier: packet.ModelTier, AttemptTier: attemptTier, Attempt: req.Attempt, OutputPath: attemptOutputPath}, fmt.Errorf("worker containment cleanup failed: %w", result.err)
		}
		outputData, err := readRunChild(prepared, attemptOutputPath, maxCatalogBytes)
		if err != nil {
			return lastReport, err
		}
		receipt, err := modelrouting.BuildRoutingReceipt(req, route, evidence, modelrouting.WorkProof{Command: "kbrouter dispatch", ArtifactHash: modelrouting.SHA256Bytes(outputData), Result: modelrouting.ProofUnknown}, time.Now())
		if err != nil {
			return lastReport, err
		}
		if err := writeRunJSON(prepared, attemptReceiptPath, receipt); err != nil {
			return lastReport, fmt.Errorf("write receipt: %w", err)
		}
		if req.Attempt == 1 {
			if err := writeRunJSON(prepared, receiptPath, receipt); err != nil {
				return lastReport, fmt.Errorf("write receipt compatibility copy: %w", err)
			}
		}
		if attribution == modelrouting.AttributionExact {
			receiptData, err := readRunChild(prepared, attemptReceiptPath, maxCatalogBytes)
			if err != nil {
				return lastReport, fmt.Errorf("read exact receipt for dispatch attestation: %w", err)
			}
			if err := validateHostBuiltReceiptBytes(receipt, receiptData); err != nil {
				return lastReport, err
			}
			if req.Attempt == 1 {
				compatibilityData, err := readRunChild(prepared, receiptPath, maxCatalogBytes)
				if err != nil {
					return lastReport, fmt.Errorf("read receipt compatibility copy for dispatch attestation: %w", err)
				}
				if !bytes.Equal(receiptData, compatibilityData) {
					return lastReport, fmt.Errorf("receipt compatibility copy changed exact attested bytes")
				}
			}
			if err := modelrouting.RecordDispatchReceiptAttestation(opts.userRoot, receipt, modelrouting.SHA256Bytes(receiptData), time.Now()); err != nil {
				return lastReport, fmt.Errorf("record dispatcher-owned route attestation: %w", err)
			}
		}
		lastReport = dispatchReport{Status: "observation-only", RouteAlias: req.RouteAlias, PlannedTier: packet.ModelTier, AttemptTier: attemptTier, ProviderReportedModel: evidence.Model, SessionID: evidence.SessionID, Attempt: req.Attempt, Attribution: string(attribution), ReceiptPath: attemptReceiptPath, OutputPath: attemptOutputPath, HandoffPath: handoffPath}
		if result.exitCode == 0 && !result.timeout {
			return lastReport, nil
		}
		lastErr = dispatchFailure(result)
		if err := appendHandoff(prepared, handoffPath, req.Attempt, lastErr, result); err != nil {
			return lastReport, fmt.Errorf("write handoff: %w", err)
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("dispatch failed")
	}
	return lastReport, lastErr
}

func validateHostBuiltReceiptBytes(receipt modelrouting.RoutingReceipt, actual []byte) error {
	expected, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal host-built receipt for dispatch attestation: %w", err)
	}
	expected = append(expected, '\n')
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("receipt changed after host construction; dispatch attestation refused")
	}
	return nil
}

func rehydrateTrustedRoute(route modelrouting.Route, sources map[string]modelrouting.Route) modelrouting.Route {
	if route.SourceRouteID == "" {
		return route
	}
	source, ok := sources[route.SourceRouteID]
	if !ok {
		return route
	}
	source.Alias = route.Alias
	source.DisplayModelID = route.DisplayModelID
	source.Capability.RouteAlias = route.Alias
	source.Capability.ModelID = route.DisplayModelID
	source.SourceRouteID = route.SourceRouteID
	source.RouteID = ""
	source.Readiness = append([]modelrouting.Readiness(nil), route.Readiness...)
	return source
}

func validateCodexPolicies(opts dispatchOptions) error {
	if !slices.Contains([]string{"read-only", "workspace-write"}, opts.sandbox) {
		return fmt.Errorf("unsupported sandbox policy %q", opts.sandbox)
	}
	if !slices.Contains([]string{"untrusted", "on-request", "on-failure", "never"}, opts.approvalPolicy) {
		return fmt.Errorf("unsupported approval policy %q", opts.approvalPolicy)
	}
	if !slices.Contains([]string{"none", "enabled"}, opts.network) {
		return fmt.Errorf("unsupported network policy %q", opts.network)
	}
	if opts.network != "none" && opts.sandbox == "read-only" {
		return fmt.Errorf("network policy requires workspace-write sandbox")
	}
	return nil
}

func loadDispatchCatalog(prepared preparedRunRoot, opts dispatchOptions, hostState dispatchTrustedState) (modelrouting.ValidatedCatalog, modelrouting.PolicyContext, error) {
	if err := prepared.revalidate(); err != nil {
		return modelrouting.ValidatedCatalog{}, modelrouting.PolicyContext{}, err
	}
	if _, err := os.Lstat(filepath.Join(prepared.runPath, "catalog.json")); err != nil {
		if os.IsNotExist(err) {
			return modelrouting.ValidatedCatalog{}, modelrouting.PolicyContext{}, fmt.Errorf("run catalog is required for dispatch")
		}
		return modelrouting.ValidatedCatalog{}, modelrouting.PolicyContext{}, err
	}
	policy, err := policyContextForProject(opts.userRoot, opts.projectRoot)
	if err != nil {
		return modelrouting.ValidatedCatalog{}, modelrouting.PolicyContext{}, fmt.Errorf("load trusted route policy: %w", err)
	}
	policy.TrustedRouteStates = hostState.RouteStates
	catalog, err := modelrouting.LoadCatalog(prepared.runPath, "catalog.json", modelrouting.StorageOptions{MaxBytes: maxCatalogBytes, Source: modelrouting.CatalogSourceRun})
	if err != nil {
		return modelrouting.ValidatedCatalog{}, modelrouting.PolicyContext{}, fmt.Errorf("load run catalog: %w", err)
	}
	validated, _, err := modelrouting.ValidateCatalogForSelection(catalog, policy, nil, time.Now(), modelrouting.CatalogSourceRun)
	if err != nil {
		return modelrouting.ValidatedCatalog{}, modelrouting.PolicyContext{}, fmt.Errorf("validate run catalog: %w", err)
	}
	return validated, policy, nil
}

func decodeDispatchPacket(data []byte, runID, sliceID string) (dispatchPacket, error) {
	var packet dispatchPacket
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&packet); err != nil {
		return dispatchPacket{}, fmt.Errorf("decode dispatch packet: %w", err)
	}
	if packet.SchemaVersion != 1 || packet.PacketID == "" || packet.TaskID == "" || packet.RunID == "" || packet.SliceID == "" ||
		packet.ModelTier == "" || packet.TaskFamily == "" || packet.ContextSize <= 0 || packet.Risk == "" ||
		len(packet.AllowedTools) == 0 || len(packet.ProofTargets) == 0 || len(packet.Redaction) == 0 || !packet.BoundedContext {
		return dispatchPacket{}, fmt.Errorf("dispatch packet missing required bounded contract fields")
	}
	if packet.RunID != runID || packet.SliceID != sliceID {
		return dispatchPacket{}, fmt.Errorf("dispatch packet does not match run/slice")
	}
	if len(packet.AllowedTools) != 1 || packet.AllowedTools[0] != "codex-harness" {
		return dispatchPacket{}, fmt.Errorf("Codex CLI dispatch requires the single codex-harness authority")
	}
	if !slices.Contains([]modelrouting.Tier{modelrouting.TierSmall, modelrouting.TierMedium, modelrouting.TierLarge}, packet.ModelTier) {
		return dispatchPacket{}, fmt.Errorf("unsupported model tier %q", packet.ModelTier)
	}
	if packet.AttemptTier != "" {
		if !slices.Contains([]modelrouting.Tier{modelrouting.TierSmall, modelrouting.TierMedium, modelrouting.TierLarge}, packet.AttemptTier) {
			return dispatchPacket{}, fmt.Errorf("unsupported attempt tier %q", packet.AttemptTier)
		}
		if dispatchTierRank(packet.AttemptTier)+1 != dispatchTierRank(packet.ModelTier) {
			return dispatchPacket{}, fmt.Errorf("attempt tier %q is not the next tier below planned model tier %q", packet.AttemptTier, packet.ModelTier)
		}
	}
	if packet.Risk != modelrouting.RiskNormal && packet.Risk != modelrouting.RiskBroad {
		return dispatchPacket{}, fmt.Errorf("unsupported packet risk %q", packet.Risk)
	}
	return packet, nil
}

func workerResultSchema() map[string]any {
	return map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"completion_status", "summary", "proof_results", "partial_work"},
		"properties": map[string]any{
			"completion_status": map[string]any{"type": "string", "enum": []string{"complete", "partial", "failed"}},
			"summary":           map[string]any{"type": "string", "maxLength": 2048},
			"proof_results": map[string]any{
				"type": "array", "maxItems": 8,
				"items": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"command", "result"},
					"properties": map[string]any{"command": map[string]any{"type": "string", "maxLength": 512}, "result": map[string]any{"type": "string", "enum": []string{"pass", "fail", "unknown"}}, "summary": map[string]any{"type": "string", "maxLength": 1024}},
				},
			},
			"partial_work": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"status"}, "properties": map[string]any{"status": map[string]any{"type": "string", "enum": []string{"none", "present", "unknown"}}, "handoff_ref": map[string]any{"type": "string", "maxLength": 512}}},
		},
	}
}

func validateExternalDispatchRoute(route modelrouting.Route) error {
	if route.Alias == "current" || route.Destination == "current" || route.TrustProvenance == "active orchestrator" {
		return fmt.Errorf("current-model route is visible-only and cannot launch external dispatch")
	}
	if hasReadiness(route.Readiness, modelrouting.ReadinessDispatchProven) != route.Capability.DispatchProven {
		return fmt.Errorf("route dispatch proof/readiness mismatch")
	}
	return nil
}

func revalidateCodexAdapterRevisionForLaunch(ctx context.Context, executable trustedExecutable, revision string) error {
	if !strings.HasPrefix(revision, "codex-cli-v1:") {
		return nil
	}
	if err := executable.revalidate(); err != nil {
		return err
	}
	versionCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	version, err := codexVersionForExecutable(versionCtx, executable.Path)
	if err != nil {
		return err
	}
	contractCtx, contractCancel := context.WithTimeout(ctx, 5*time.Second)
	defer contractCancel()
	if err := codexAdapterContractProbe(contractCtx, executable.Path); err != nil {
		return err
	}
	current, err := codexAdapterRevision(executable.Hash, version)
	if err != nil {
		return err
	}
	if current != revision {
		return fmt.Errorf("expected %s got %s", revision, current)
	}
	return nil
}

func resolveTrustedCodexExecutable(projectRoot, runRoot string) (trustedExecutable, error) {
	resolved, err := dispatchExecutableResolver()
	if err != nil {
		return trustedExecutable{}, fmt.Errorf("resolve codex executable: %w", err)
	}
	if strings.TrimSpace(resolved) == "" || strings.ContainsAny(resolved, "\x00\r\n") {
		return trustedExecutable{}, fmt.Errorf("trusted codex executable path is invalid")
	}
	abs, err := filepath.Abs(filepath.Clean(resolved))
	if err != nil {
		return trustedExecutable{}, err
	}
	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return trustedExecutable{}, fmt.Errorf("resolve canonical codex executable: %w", err)
	}
	abs, err = filepath.Abs(filepath.Clean(abs))
	if err != nil {
		return trustedExecutable{}, err
	}
	projectAbs, err := canonicalExistingDispatchRoot(projectRoot)
	if err != nil {
		return trustedExecutable{}, fmt.Errorf("resolve canonical project root: %w", err)
	}
	runAbs, err := canonicalExistingDispatchRoot(runRoot)
	if err != nil {
		return trustedExecutable{}, fmt.Errorf("resolve canonical run root: %w", err)
	}
	if strings.TrimSpace(projectRoot) != "" && pathWithin(abs, projectAbs) {
		return trustedExecutable{}, fmt.Errorf("trusted codex executable cannot live under project or run root")
	}
	if strings.TrimSpace(runRoot) != "" && pathWithin(abs, runAbs) {
		return trustedExecutable{}, fmt.Errorf("trusted codex executable cannot live under project or run root")
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return trustedExecutable{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return trustedExecutable{}, fmt.Errorf("trusted codex executable must be a regular file")
	}
	hash, err := hashFile(abs, 16*1024*1024)
	if err != nil {
		return trustedExecutable{}, err
	}
	return trustedExecutable{Path: abs, Hash: hash, Info: info}, nil
}

func canonicalExistingDispatchRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", nil
	}
	abs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", err
	}
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Clean(canonical))
}

func (exe trustedExecutable) revalidate() error {
	info, err := os.Lstat(exe.Path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || !os.SameFile(info, exe.Info) {
		return fmt.Errorf("file identity changed")
	}
	hash, err := hashFile(exe.Path, 16*1024*1024)
	if err != nil {
		return err
	}
	if hash != exe.Hash {
		return fmt.Errorf("file hash changed")
	}
	return nil
}

func hashFile(path string, limit int64) (string, error) {
	data, err := readBoundedFile(path, limit)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func dispatchStateLocation(projectID, runID string) (string, string) {
	sum := sha256.Sum256([]byte(projectID))
	safeRun := safeAlias(runID)
	return filepath.Join("dispatch-state", hex.EncodeToString(sum[:16])), safeRun + ".json"
}

func saveDispatchTrustedState(userRoot string, prepared preparedRunRoot, catalog modelrouting.Catalog) error {
	states := map[string]string{}
	for _, route := range catalog.Routes {
		state, err := modelrouting.ComputeRouteStateFingerprint(route)
		if err != nil {
			return err
		}
		states[route.Alias] = state
	}
	relDir, file := dispatchStateLocation(prepared.marker.ProjectID, prepared.marker.RunID)
	root := filepath.Join(userRoot, relDir)
	if err := os.MkdirAll(root, 0o700); err != nil {
		return err
	}
	state := dispatchTrustedState{
		SchemaVersion:      1,
		ProjectID:          prepared.marker.ProjectID,
		RunID:              prepared.marker.RunID,
		CatalogFingerprint: catalog.Fingerprint,
		SurfaceRevision:    "slice-004-v1",
		RouteStates:        states,
		CreatedAt:          time.Now().UTC(),
		ExpiresAt:          time.Now().Add(time.Hour).UTC(),
	}
	return modelrouting.SaveAtomicJSON(root, file, state, maxCatalogBytes)
}

func trustedCodexProfileRevision(codexHome, profileName, endpoint, authEnv string) (string, error) {
	if !validProfileNameForDispatch(profileName) {
		return "", fmt.Errorf("safe profile name is required")
	}
	root, err := filepath.Abs(filepath.Clean(codexHome))
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, profileName+".config.toml")
	if !pathWithin(path, root) {
		return "", modelrouting.ErrUnsafePath
	}
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return "", modelrouting.ErrUnsafePath
	}
	data, err := readBoundedFile(path, 64*1024)
	if err != nil {
		return "", err
	}
	lower := strings.ToLower(string(data))
	for _, forbidden := range []string{"auth.command", "api_key", "token =", "authorization", "headers", "credential"} {
		if strings.Contains(lower, forbidden) {
			return "", fmt.Errorf("unsupported credential source in Codex profile")
		}
	}
	if endpoint != "" && !strings.Contains(string(data), endpoint) {
		return "", fmt.Errorf("Codex profile does not bind the approved endpoint")
	}
	if authEnv != "" && !strings.Contains(string(data), authEnv) {
		return "", fmt.Errorf("Codex profile does not bind the approved auth env")
	}
	return hashFile(path, 64*1024)
}

func validProfileNameForDispatch(value string) bool {
	if value == "" || len(value) > 64 || strings.Contains(value, "..") {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' || character == '_' || character == '.' {
			continue
		}
		return false
	}
	return true
}

func routeFromValidatedCatalog(validated modelrouting.ValidatedCatalog, policy modelrouting.PolicyContext, alias string, packet dispatchPacket) (modelrouting.Route, error) {
	request := modelrouting.WorkRequest{PlannedTier: packet.ModelTier, AttemptTier: packet.AttemptTier, TaskFamily: packet.TaskFamily, Tools: packet.AllowedTools, ContextSize: packet.ContextSize, Risk: packet.Risk, ProjectID: policy.Project.ProjectID}
	decision, err := modelrouting.SelectRoute(validated, request, policy, modelrouting.RunOverride{Mode: modelrouting.OverrideRequire, Alias: alias}, modelrouting.AttemptLedger{}, time.Now())
	if err != nil {
		if errors.Is(err, modelrouting.ErrRequiredRouteUnavailable) {
			return modelrouting.Route{}, fmt.Errorf("route %q not trusted/selectable for dispatch: %w", alias, err)
		}
		return modelrouting.Route{}, err
	}
	if len(decision.Routes) == 0 {
		return modelrouting.Route{}, fmt.Errorf("route %q not trusted/selectable for dispatch", alias)
	}
	return decision.Routes[0], nil
}

func dispatchTierRank(tier modelrouting.Tier) int {
	switch tier {
	case modelrouting.TierSmall:
		return 1
	case modelrouting.TierMedium:
		return 2
	case modelrouting.TierLarge:
		return 3
	default:
		return 0
	}
}

func fallbackAllowed(first, next modelrouting.Route) bool {
	return dispatchClassRank(next.Capability.Class) >= dispatchClassRank(first.Capability.Class)
}

func fallbackTrustTransitionAllowed(first, next modelrouting.Route, policy modelrouting.PolicyContext) bool {
	if !lessTrustedFallback(first, next) {
		return true
	}
	fingerprint, err := modelrouting.ApprovalRouteFingerprint(next, policy.RouteSources)
	if err != nil {
		return false
	}
	now := time.Now()
	for _, approval := range policy.Trusted.RouteApprovals {
		if approval.ProjectID == policy.Project.ProjectID && approval.RouteFingerprint == fingerprint && now.Before(approval.ExpiresAt) {
			return true
		}
	}
	return false
}

func lessTrustedFallback(first, next modelrouting.Route) bool {
	if first.Boundary != next.Boundary {
		return true
	}
	if strings.TrimSpace(first.Adapter) != strings.TrimSpace(next.Adapter) {
		return true
	}
	if strings.TrimSpace(first.AdapterRevision) != strings.TrimSpace(next.AdapterRevision) {
		return true
	}
	if strings.TrimSpace(first.DispatchMethod) != strings.TrimSpace(next.DispatchMethod) {
		return true
	}
	if strings.TrimSpace(first.Profile) != strings.TrimSpace(next.Profile) {
		return true
	}
	if strings.TrimSpace(first.ProfileRevision) != strings.TrimSpace(next.ProfileRevision) {
		return true
	}
	if strings.TrimSpace(first.Destination) != strings.TrimSpace(next.Destination) {
		return true
	}
	if strings.TrimSpace(first.Endpoint) != strings.TrimSpace(next.Endpoint) {
		return true
	}
	if strings.TrimSpace(first.AuthEnv) != strings.TrimSpace(next.AuthEnv) {
		return true
	}
	if strings.TrimSpace(first.SourceRouteID) != strings.TrimSpace(next.SourceRouteID) {
		return true
	}
	if strings.TrimSpace(first.TrustProvenance) != strings.TrimSpace(next.TrustProvenance) {
		return true
	}
	if dispatchRetentionRank(next.Retention) > dispatchRetentionRank(first.Retention) {
		return true
	}
	if first.TrainingUse != next.TrainingUse {
		return true
	}
	return !strings.EqualFold(strings.TrimSpace(first.Residency), strings.TrimSpace(next.Residency))
}

type dispatchRunLock struct {
	lock *modelrouting.PrivateStateLock
}

func dispatchRunLockTimeout(opts dispatchOptions) time.Duration {
	attempts := opts.attemptLimit
	if attempts <= 0 {
		attempts = 1
	}
	timeout := opts.timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return timeout*time.Duration(attempts) + 5*time.Second
}

func acquireDispatchRunLock(prepared preparedRunRoot, userRoot string, timeout time.Duration) (*dispatchRunLock, error) {
	if err := prepared.revalidate(); err != nil {
		return nil, err
	}
	relDir, stateFile := dispatchStateLocation(prepared.marker.ProjectID, prepared.marker.RunID)
	lockRoot := filepath.Join(userRoot, relDir)
	lockName := strings.TrimSuffix(stateFile, filepath.Ext(stateFile)) + ".dispatch.lock"
	lock, err := modelrouting.AcquirePrivateStateLock(lockRoot, lockName, timeout)
	if err != nil {
		return nil, err
	}
	return &dispatchRunLock{lock: lock}, nil
}

func rejectDispatchRunLockArtifactTarget(paths ...string) error {
	for _, path := range paths {
		if path == "" {
			continue
		}
		if strings.EqualFold(filepath.Base(path), dispatchRunLockName) {
			return fmt.Errorf("reserved artifact name %q", filepath.Base(path))
		}
	}
	return nil
}

func (lock *dispatchRunLock) release() {
	if lock == nil {
		return
	}
	if lock.lock != nil {
		_ = lock.lock.Close()
		lock.lock = nil
	}
}

func dispatchRetentionRank(value modelrouting.RetentionClass) int {
	switch value {
	case modelrouting.RetentionNone:
		return 0
	case modelrouting.RetentionSession:
		return 1
	case modelrouting.RetentionLimited:
		return 2
	default:
		return 3
	}
}

func dispatchClassRank(class modelrouting.CapabilityClass) int {
	switch class {
	case modelrouting.ClassSmall:
		return 1
	case modelrouting.ClassMedium:
		return 2
	case modelrouting.ClassLarge, modelrouting.ClassPlanner:
		return 3
	default:
		return 0
	}
}

func runWorkerProcess(execPath string, req modelrouting.DispatchRequest, stdin []byte, limit int64, codexHome, routeAuthEnv string) processResult {
	if err := dispatchProcessTreeContainment(); err != nil {
		return processResult{exitCode: -1, notStarted: true, err: err}
	}
	ctx, cancel := context.WithTimeout(context.Background(), req.Timeout)
	defer cancel()
	args := codexExecArgs(req)
	cmd := exec.CommandContext(ctx, execPath, args...)
	if err := configureProcessTree(cmd); err != nil {
		return processResult{exitCode: -1, notStarted: true, err: err}
	}
	cmd.Dir = req.CWD
	cmd.Stdin = bytes.NewReader(stdin)
	cmd.Env = minimalDispatchEnv(codexHome, routeAuthEnv)
	cmd.WaitDelay = 2 * time.Second
	stdout := &boundedBuffer{limit: limit}
	stderr := &boundedBuffer{limit: limit}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return processResult{exitCode: -1, notStarted: true, err: err}
	}
	tree, attachErr := attachProcessTree(cmd)
	if attachErr != nil {
		killErr := cmd.Process.Kill()
		_ = cmd.Wait()
		return processResult{exitCode: -1, notStarted: true, containmentFailure: killErr != nil, err: errors.Join(attachErr, killErr)}
	}
	cmd.Cancel = func() error { return tree.Kill() }
	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()
	var err error
	var killErr error
	select {
	case err = <-waitCh:
	case <-ctx.Done():
		killErr = tree.Kill()
		err = <-waitCh
	}
	killErr = errors.Join(killErr, tree.Kill())
	cleanupErr := tree.Close()
	result := processResult{stdout: stdout.String(), stderr: stderr.String(), overflow: stdout.exceeded || stderr.exceeded}
	if killErr != nil || cleanupErr != nil {
		result.exitCode = -1
		result.containmentFailure = true
		result.err = errors.Join(killErr, cleanupErr)
		return result
	}
	if ctx.Err() == context.DeadlineExceeded {
		result.timeout = true
		result.exitCode = -1
		result.err = ctx.Err()
		return result
	}
	if err != nil {
		result.err = err
		result.exitCode = 1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.exitCode = exitErr.ExitCode()
		}
		return result
	}
	return result
}

func codexExecArgs(req modelrouting.DispatchRequest) []string {
	args := []string{"exec", "--model", req.Model}
	if req.DispatchMethod == "exec-profile" {
		args = append(args, "--profile", req.Profile)
	}
	args = append(args, "--sandbox", req.Sandbox, "-C", req.CWD)
	for _, root := range req.AllowedRoots {
		if !sameFilesystemPath(root, req.CWD) {
			args = append(args, "--add-dir", root)
		} else {
			args = append(args, "--add-dir", root)
		}
	}
	args = append(args,
		"-c", `approval_policy="`+req.ApprovalPolicy+`"`,
		"-c", "sandbox_workspace_write.network_access="+networkConfigValue(req.Network),
		"--output-schema", req.OutputSchemaPath,
		"--json",
		"-",
	)
	return args
}

func networkConfigValue(policy string) string {
	return fmt.Sprintf("%v", policy == "enabled")
}

func minimalDispatchEnv(codexHome, routeAuthEnv string) []string {
	names := []string{"PATH", "PATHEXT", "SystemRoot", "COMSPEC", "TMP", "TEMP", "HOME", "USERPROFILE", "APPDATA", "LOCALAPPDATA", "XDG_CONFIG_HOME", "XDG_CACHE_HOME"}
	out := []string{"CODEX_HOME=" + codexHome}
	seen := map[string]struct{}{"CODEX_HOME": {}}
	for _, name := range names {
		if value, ok := os.LookupEnv(name); ok {
			out = append(out, name+"="+value)
			seen[name] = struct{}{}
		}
	}
	if routeAuthEnv != "" {
		if value, ok := os.LookupEnv(routeAuthEnv); ok {
			out = append(out, routeAuthEnv+"="+value)
		}
	}
	if runtime.GOOS == "windows" {
		if _, ok := seen["NoDefaultCurrentDirectoryInExePath"]; !ok {
			out = append(out, "NoDefaultCurrentDirectoryInExePath=1")
		}
	}
	return out
}

func parseThreadStartedSession(text string) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 0, 4096), int(defaultDispatchOutputLimit))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var payload struct {
			Type     string `json:"type"`
			ThreadID string `json:"thread_id"`
		}
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			continue
		}
		if payload.Type == "thread.started" && payload.ThreadID != "" {
			return payload.ThreadID
		}
	}
	return ""
}

func loadCodexSessionEvidence(codexHome, sessionID string, req modelrouting.DispatchRequest, route modelrouting.Route, attemptStart time.Time) modelrouting.ProviderEvidence {
	if req.Sandbox == "danger-full-access" || sessionID == "" {
		return modelrouting.ProviderEvidence{}
	}
	rel, err := findCodexSessionLog(codexHome, sessionID, attemptStart)
	if err != nil {
		return modelrouting.ProviderEvidence{}
	}
	var seenMeta, seenTurn bool
	var provider, model string
	if err := scanCodexSessionChild(codexHome, rel, defaultSessionEvidenceLimit, func(line []byte) bool {
		var entry struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal(line, &entry); err != nil {
			return false
		}
		switch entry.Type {
		case "session_meta":
			var payload struct {
				ID            string `json:"id"`
				ModelProvider string `json:"model_provider"`
				CWD           string `json:"cwd"`
			}
			if json.Unmarshal(entry.Payload, &payload) != nil || payload.ID != sessionID || !sameFilesystemPath(filepath.FromSlash(payload.CWD), req.CWD) {
				seenMeta, seenTurn = false, false
				return false
			}
			provider = payload.ModelProvider
			seenMeta = true
		case "turn_context":
			var payload struct {
				Model           string `json:"model"`
				CWD             string `json:"cwd"`
				ApprovalPolicy  string `json:"approval_policy"`
				Profile         string `json:"profile,omitempty"`
				ProfileRevision string `json:"profile_revision,omitempty"`
				RouteAlias      string `json:"route_alias,omitempty"`
				SandboxPolicy   struct {
					Type string `json:"type"`
				} `json:"sandbox_policy"`
			}
			if json.Unmarshal(entry.Payload, &payload) != nil || !sameFilesystemPath(filepath.FromSlash(payload.CWD), req.CWD) ||
				payload.ApprovalPolicy != req.ApprovalPolicy || payload.SandboxPolicy.Type != req.Sandbox {
				seenMeta, seenTurn = false, false
				return false
			}
			if payload.Profile != "" && payload.Profile != req.Profile {
				seenMeta, seenTurn = false, false
				return false
			}
			if payload.ProfileRevision != "" && payload.ProfileRevision != req.ProfileRevision {
				seenMeta, seenTurn = false, false
				return false
			}
			if payload.RouteAlias != "" && payload.RouteAlias != req.RouteAlias {
				seenMeta, seenTurn = false, false
				return false
			}
			model = payload.Model
			seenTurn = true
		}
		return seenMeta && seenTurn
	}); err != nil {
		return modelrouting.ProviderEvidence{}
	}
	if !seenMeta || !seenTurn || model != req.Model || provider == "" || (route.Destination != "" && !strings.EqualFold(provider, route.Destination) && route.Destination != "codex") {
		return modelrouting.ProviderEvidence{}
	}
	return modelrouting.ProviderEvidence{Model: model, SessionID: sessionID}
}

func findCodexSessionLog(codexHome, sessionID string, attemptStart time.Time) (string, error) {
	if !validCodexSessionID(sessionID) {
		return "", modelrouting.ErrUnsafePath
	}
	root, err := filepath.Abs(filepath.Clean(codexHome))
	if err != nil {
		return "", err
	}
	sessionsRoot := filepath.Join(root, "sessions")
	matches := []string{}
	legacy := filepath.Join("sessions", sessionID+".jsonl")
	if ok, err := codexSessionCandidate(root, legacy, attemptStart); err != nil {
		return "", err
	} else if ok {
		matches = append(matches, legacy)
	}
	info, err := os.Lstat(sessionsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return "", os.ErrNotExist
		}
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", modelrouting.ErrUnsafePath
	}
	for _, date := range sessionSearchDates(attemptStart) {
		pattern := filepath.Join(sessionsRoot, date.Format("2006"), date.Format("01"), date.Format("02"), "rollout-*-"+sessionID+".jsonl")
		paths, err := filepath.Glob(pattern)
		if err != nil {
			return "", err
		}
		for _, path := range paths {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return "", err
			}
			ok, err := codexSessionCandidate(root, rel, attemptStart)
			if err != nil {
				return "", err
			}
			if ok {
				matches = append(matches, rel)
			}
		}
	}
	if len(matches) != 1 {
		if len(matches) == 0 {
			return "", os.ErrNotExist
		}
		return "", fmt.Errorf("ambiguous Codex session evidence")
	}
	return matches[0], nil
}

func sessionSearchDates(attemptStart time.Time) []time.Time {
	if attemptStart.IsZero() {
		attemptStart = time.Now()
	}
	seen := map[string]struct{}{}
	out := []time.Time{}
	for _, base := range []time.Time{attemptStart.Local(), attemptStart.UTC()} {
		for offset := -1; offset <= 1; offset++ {
			day := base.AddDate(0, 0, offset)
			key := day.Format("2006-01-02")
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, day)
		}
	}
	return out
}

func codexSessionCandidate(root, rel string, attemptStart time.Time) (bool, error) {
	if filepath.IsAbs(rel) || strings.Contains(rel, "..") {
		return false, modelrouting.ErrUnsafePath
	}
	path := filepath.Join(root, rel)
	info, err := lstatSessionPath(root, path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return false, modelrouting.ErrUnsafePath
	}
	if !attemptStart.IsZero() && info.ModTime().Before(attemptStart.Add(-10*time.Second)) {
		return false, nil
	}
	return true, nil
}

func lstatSessionPath(root, path string) (os.FileInfo, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return nil, modelrouting.ErrUnsafePath
	}
	current := root
	parts := strings.Split(rel, string(filepath.Separator))
	var info os.FileInfo
	for index, part := range parts {
		if part == "" || part == "." || part == ".." {
			return nil, modelrouting.ErrUnsafePath
		}
		current = filepath.Join(current, part)
		info, err = os.Lstat(current)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, modelrouting.ErrUnsafePath
		}
		if index < len(parts)-1 && !info.IsDir() {
			return nil, modelrouting.ErrUnsafePath
		}
	}
	return info, nil
}

func validCodexSessionID(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' || character == '_' {
			continue
		}
		return false
	}
	return true
}

func scanCodexSessionChild(codexHome, rel string, limit int64, consume func([]byte) bool) error {
	if filepath.IsAbs(rel) || strings.Contains(rel, "..") {
		return modelrouting.ErrUnsafePath
	}
	path := filepath.Join(codexHome, rel)
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return modelrouting.ErrUnsafePath
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return modelrouting.ErrUnsafePath
	}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 4096), int(limit))
	var consumed int64
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		consumed += int64(len(line)) + 1
		if consumed > limit {
			return modelrouting.ErrStorageSizeExceeded
		}
		if consume(line) {
			return nil
		}
	}
	return scanner.Err()
}

func dispatchFailure(result processResult) error {
	if result.timeout {
		return fmt.Errorf("timeout")
	}
	if result.exitCode != 0 {
		return fmt.Errorf("worker exited nonzero: %d", result.exitCode)
	}
	return fmt.Errorf("dispatch failed")
}

func safeRunChild(prepared preparedRunRoot, value, defaultName string) (string, error) {
	if err := prepared.revalidate(); err != nil {
		return "", err
	}
	if strings.TrimSpace(value) == "" {
		value = defaultName
	}
	if filepath.IsAbs(value) {
		return "", fmt.Errorf("artifact path must be a direct child of the marked run root")
	}
	clean := filepath.Clean(value)
	if clean == "." || clean == ".." || strings.Contains(clean, string(filepath.Separator)) || strings.Contains(clean, "/") || strings.Contains(clean, `\`) {
		return "", fmt.Errorf("artifact path must be a direct child of the marked run root")
	}
	return filepath.Join(prepared.runPath, clean), nil
}

func validateDispatchArtifactNamespace(prepared preparedRunRoot, attemptLimit int, packetPath, outputPath, receiptPath, handoffPath, workerRequestPath, schemaPath string) error {
	if attemptLimit <= 0 {
		attemptLimit = 1
	}
	type artifactEntry struct {
		path        string
		noOverwrite bool
	}
	entries := []artifactEntry{
		{path: packetPath},
		{path: handoffPath},
		{path: outputPath, noOverwrite: true},
		{path: receiptPath, noOverwrite: true},
		{path: schemaPath, noOverwrite: true},
	}
	if workerRequestPath != "" {
		entries = append(entries, artifactEntry{path: workerRequestPath, noOverwrite: true})
	}
	for attempt := 1; attempt <= attemptLimit; attempt++ {
		entries = append(entries,
			artifactEntry{path: attemptArtifactPath(outputPath, attempt), noOverwrite: true},
			artifactEntry{path: attemptArtifactPath(receiptPath, attempt), noOverwrite: true},
		)
		if workerRequestPath != "" {
			entries = append(entries, artifactEntry{path: attemptArtifactPath(workerRequestPath, attempt), noOverwrite: true})
		}
	}
	reserved := map[string]struct{}{
		".kb-run-root.json": {},
		dispatchRunLockName: {},
		"catalog.json":      {},
	}
	seen := map[string]struct{}{}
	for _, entry := range entries {
		if entry.path == "" {
			continue
		}
		if !sameDirectChild(prepared.runPath, entry.path) {
			return modelrouting.ErrUnsafePath
		}
		base := filepath.Base(entry.path)
		if _, ok := reserved[base]; ok {
			return fmt.Errorf("reserved artifact name %q", base)
		}
		key := strings.ToLower(base)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("artifact paths must be pairwise distinct")
		}
		seen[key] = struct{}{}
		if entry.noOverwrite {
			if info, err := os.Lstat(entry.path); err == nil {
				if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
					return modelrouting.ErrUnsafePath
				}
				return fmt.Errorf("refusing to overwrite existing dispatch artifact %q", base)
			} else if !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

func validateDispatchArtifactNames(paths ...string) error {
	prepared := preparedRunRoot{}
	for _, path := range paths {
		if path != "" {
			prepared.runPath = filepath.Dir(path)
			break
		}
	}
	return validateDispatchArtifactNamespace(prepared, 1, paths[0], paths[1], paths[2], paths[3], "", "")
}

func attemptArtifactPath(path string, attempt int) string {
	if attempt <= 1 {
		ext := filepath.Ext(path)
		stem := strings.TrimSuffix(filepath.Base(path), ext)
		return filepath.Join(filepath.Dir(path), stem+"-attempt-1"+ext)
	}
	ext := filepath.Ext(path)
	stem := strings.TrimSuffix(filepath.Base(path), ext)
	return filepath.Join(filepath.Dir(path), fmt.Sprintf("%s-attempt-%d%s", stem, attempt, ext))
}

func readRunChild(prepared preparedRunRoot, path string, limit int64) ([]byte, error) {
	if err := prepared.revalidate(); err != nil {
		return nil, err
	}
	if !sameDirectChild(prepared.runPath, path) {
		return nil, modelrouting.ErrUnsafePath
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, modelrouting.ErrUnsafePath
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return nil, modelrouting.ErrUnsafePath
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, modelrouting.ErrStorageSizeExceeded
	}
	if err := prepared.revalidate(); err != nil {
		return nil, err
	}
	return data, nil
}

func writeRunJSON(prepared preparedRunRoot, path string, value any) error {
	if err := prepared.revalidate(); err != nil {
		return err
	}
	if !sameDirectChild(prepared.runPath, path) {
		return modelrouting.ErrUnsafePath
	}
	return modelrouting.SaveAtomicJSON(prepared.runPath, filepath.Base(path), value, maxCatalogBytes)
}

func appendHandoff(prepared preparedRunRoot, path string, attempt int, _ error, result processResult) error {
	if path == "" {
		return nil
	}
	var existing struct {
		Attempts []map[string]any `json:"attempts"`
	}
	if data, err := readRunChild(prepared, path, maxCatalogBytes); err == nil {
		_ = json.Unmarshal(data, &existing)
	}
	existing.Attempts = append(existing.Attempts, map[string]any{
		"attempt":       attempt,
		"cause_code":    dispatchFailureCode(result),
		"exit_code":     result.exitCode,
		"timeout":       result.timeout,
		"overflow":      result.overflow,
		"stdout_sha256": modelrouting.SHA256Bytes([]byte(result.stdout)),
		"stderr_sha256": modelrouting.SHA256Bytes([]byte(result.stderr)),
	})
	return writeRunJSON(prepared, path, existing)
}

func dispatchFailureCode(result processResult) string {
	switch {
	case result.timeout:
		return "timeout"
	case result.overflow:
		return "bounded-output-overflow"
	case result.exitCode != 0:
		return "nonzero-exit"
	default:
		return "execution-failure"
	}
}

func sameDirectChild(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return filepath.Dir(rel) == "."
}

func canonicalAllowedRoots(values []string, projectRoot string) ([]string, error) {
	if len(values) == 0 {
		values = []string{projectRoot}
	}
	projectAbs, err := filepath.Abs(filepath.Clean(projectRoot))
	if err != nil {
		return nil, err
	}
	if canonical, canonicalErr := filepath.EvalSymlinks(projectAbs); canonicalErr == nil {
		projectAbs = canonical
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		abs, err := filepath.Abs(filepath.Clean(value))
		if err != nil {
			return nil, err
		}
		if canonical, canonicalErr := filepath.EvalSymlinks(abs); canonicalErr == nil {
			abs = canonical
		}
		if !pathWithin(abs, projectAbs) {
			return nil, fmt.Errorf("allowed root %s is outside approved project authority", value)
		}
		out = append(out, abs)
	}
	return out, nil
}

func canonicalAllowedTools(values []string) ([]string, error) {
	if len(values) != 1 || values[0] != "codex-harness" {
		return nil, fmt.Errorf("Codex CLI dispatch supports only the codex-harness authority")
	}
	return []string{"codex-harness"}, nil
}

func pathWithin(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel))
}
