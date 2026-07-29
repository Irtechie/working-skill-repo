package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

const (
	ddrAttemptSchemaVersion = 1
	ddrParentReturnExitCode = 10
	ddrBlockedExitCode      = 3
	ddrDefaultProbeTimeout  = 2 * time.Second
	ddrDefaultTimeout       = 20 * time.Second
	ddrMaxProbeTimeout      = 5 * time.Second
	ddrMaxTimeout           = 60 * time.Second
)

var proofHashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type ddrAttemptOptions struct {
	commonOptions
	runID, sliceID, alias, tier, tierReason, taskFamily, risk string
	requestPath                                               string
	tools                                                     repeatFlag
	contextSize                                               int
	sensitive, sensitiveSet, require                          bool
	probeTimeout, timeout                                     time.Duration
}

type ddrResolveOptions struct {
	commonOptions
	runID, sliceID, proofResult, proofCommand, proofArtifactHash string
}

type ddrAttemptReport struct {
	SchemaVersion     int                    `json:"schema_version"`
	Status            string                 `json:"status"`
	Action            string                 `json:"action"`
	ProjectID         string                 `json:"project_id"`
	RunID             string                 `json:"run_id"`
	SliceID           string                 `json:"slice_id"`
	RouteAlias        string                 `json:"route_alias"`
	RouteFingerprint  string                 `json:"route_fingerprint,omitempty"`
	RequestHash       string                 `json:"request_hash"`
	AttemptBinding    string                 `json:"attempt_binding"`
	Attempt           int                    `json:"attempt"`
	FailureClass      string                 `json:"failure_class,omitempty"`
	FailurePhase      string                 `json:"failure_phase,omitempty"`
	ParentSelection   string                 `json:"parent_selection"`
	RequirePin        bool                   `json:"require_pin"`
	ProbeLatencyMS    int64                  `json:"probe_latency_ms"`
	DispatchLatencyMS int64                  `json:"dispatch_latency_ms"`
	TotalLatencyMS    int64                  `json:"total_latency_ms"`
	Proof             modelrouting.WorkProof `json:"proof"`
	Response          string                 `json:"response,omitempty"`
	ResultDelivery    string                 `json:"result_delivery"`
	ObservedAt        time.Time              `json:"observed_at"`
}

type ddrRequestFile struct {
	SchemaVersion int          `json:"schema_version"`
	Messages      []ddrMessage `json:"messages"`
	MaxTokens     int          `json:"max_tokens"`
}

type ddrMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ddrAttemptBinding struct {
	ProjectID        string        `json:"project_id"`
	RunID            string        `json:"run_id"`
	SliceID          string        `json:"slice_id"`
	RouteAlias       string        `json:"route_alias"`
	RouteFingerprint string        `json:"route_fingerprint,omitempty"`
	RequestHash      string        `json:"request_hash"`
	Tier             string        `json:"tier"`
	TierReason       string        `json:"tier_reason"`
	TaskFamily       string        `json:"task_family"`
	Tools            []string      `json:"tools"`
	ContextSize      int           `json:"context_size"`
	Risk             string        `json:"risk"`
	SensitiveData    bool          `json:"sensitive_data"`
	RequirePin       bool          `json:"require_pin"`
	ProbeTimeout     time.Duration `json:"probe_timeout"`
	DispatchTimeout  time.Duration `json:"dispatch_timeout"`
}

type ddrHTTPStatusError struct {
	StatusCode int
	Phase      string
}

func (e ddrHTTPStatusError) Error() string {
	return fmt.Sprintf("%s returned HTTP %d", e.Phase, e.StatusCode)
}

func runDDRAttempt(args []string, stdout, stderr io.Writer) int {
	if hasHelpArg(args) {
		fmt.Fprint(stdout, ddrAttemptUsage)
		return 0
	}
	fs := flagSet("ddr attempt")
	opts := ddrAttemptOptions{probeTimeout: ddrDefaultProbeTimeout, timeout: ddrDefaultTimeout}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.runID, "run-id", "", "stable parent run id")
	fs.StringVar(&opts.sliceID, "slice-id", "", "stable slice id")
	fs.StringVar(&opts.alias, "alias", "", "configured route alias")
	fs.StringVar(&opts.tier, "tier", "", "small, medium, or large")
	fs.StringVar(&opts.tierReason, "tier-reason", "", "minimum capability rationale")
	fs.StringVar(&opts.taskFamily, "task-family", "", "task family")
	fs.Var(&opts.tools, "tool", "required tool; repeatable")
	fs.IntVar(&opts.contextSize, "context-size", 0, "required context size")
	fs.StringVar(&opts.risk, "risk", "", "normal only for local preference")
	fs.BoolVar(&opts.sensitive, "sensitive-data", false, "request contains sensitive data")
	fs.StringVar(&opts.requestPath, "request", "", "strict request JSON inside the project root")
	fs.DurationVar(&opts.probeTimeout, "probe-timeout", ddrDefaultProbeTimeout, "bounded model-presence probe timeout")
	fs.DurationVar(&opts.timeout, "timeout", ddrDefaultTimeout, "bounded dispatch timeout")
	fs.BoolVar(&opts.require, "require", false, "hard pin this alias instead of returning to the parent")
	if err := fs.Parse(args); err != nil {
		return commandError(stdout, stderr, hasExactArg(args, "--json"), 2, "invalid-arguments", err.Error())
	}
	fs.Visit(func(flag *flag.Flag) {
		if flag.Name == "sensitive-data" {
			opts.sensitiveSet = true
		}
	})
	if fs.NArg() != 0 {
		return commandError(stdout, stderr, opts.json, 2, "invalid-arguments", fmt.Sprintf("unexpected argument %q", fs.Arg(0)))
	}
	if customUserRootRejected(fs) {
		return commandError(stdout, stderr, opts.json, 2, "invalid-arguments", "DDR attempts use the fixed canonical user-local root; custom --user-root is test-only")
	}
	if err := validateDDRAttemptOptions(opts); err != nil {
		return commandError(stdout, stderr, opts.json, 2, "invalid-arguments", err.Error())
	}
	report, code, err := executeDDRAttempt(opts)
	if err != nil {
		return commandError(stdout, stderr, opts.json, 1, "attempt-error", err.Error())
	}
	if printResult(stdout, stderr, report, opts.json, nil) != 0 {
		return 1
	}
	return code
}

func validateDDRAttemptOptions(opts ddrAttemptOptions) error {
	if opts.runID == "" || opts.sliceID == "" || opts.alias == "" || opts.tier == "" ||
		strings.TrimSpace(opts.tierReason) == "" || opts.taskFamily == "" || len(opts.tools) == 0 ||
		opts.contextSize <= 0 || opts.requestPath == "" {
		return fmt.Errorf("ddr attempt requires complete run, route, capability, and request bindings")
	}
	if opts.tier != string(modelrouting.TierSmall) && opts.tier != string(modelrouting.TierMedium) && opts.tier != string(modelrouting.TierLarge) {
		return fmt.Errorf("tier must be small, medium, or large")
	}
	if !opts.sensitiveSet {
		return fmt.Errorf("sensitive-data must be explicitly set to true or false")
	}
	if opts.risk != string(modelrouting.RiskNormal) {
		return fmt.Errorf("local DDR attempt requires normal risk; broad work stays with parent selection")
	}
	if opts.probeTimeout <= 0 || opts.probeTimeout > ddrMaxProbeTimeout || opts.timeout <= 0 || opts.timeout > ddrMaxTimeout {
		return fmt.Errorf("timeouts must be positive and at most %s probe / %s dispatch", ddrMaxProbeTimeout, ddrMaxTimeout)
	}
	return nil
}

func runDDRResolve(args []string, stdout, stderr io.Writer) int {
	if hasHelpArg(args) {
		fmt.Fprint(stdout, ddrResolveUsage)
		return 0
	}
	fs := flagSet("ddr resolve")
	opts := ddrResolveOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.runID, "run-id", "", "stable parent run id")
	fs.StringVar(&opts.sliceID, "slice-id", "", "stable slice id")
	fs.StringVar(&opts.proofResult, "proof-result", "", "deterministic parent proof result: pass or fail")
	fs.StringVar(&opts.proofCommand, "proof-command", "", "deterministic proof command")
	fs.StringVar(&opts.proofArtifactHash, "proof-artifact-hash", "", "sha256 proof artifact hash")
	if err := fs.Parse(args); err != nil {
		return commandError(stdout, stderr, hasExactArg(args, "--json"), 2, "invalid-arguments", err.Error())
	}
	if fs.NArg() != 0 {
		return commandError(stdout, stderr, opts.json, 2, "invalid-arguments", fmt.Sprintf("unexpected argument %q", fs.Arg(0)))
	}
	if customUserRootRejected(fs) {
		return commandError(stdout, stderr, opts.json, 2, "invalid-arguments", "DDR receipts use the fixed canonical user-local root; custom --user-root is test-only")
	}
	if err := validateDDRResolveOptions(opts); err != nil {
		return commandError(stdout, stderr, opts.json, 2, "invalid-arguments", err.Error())
	}
	report, code, err := executeDDRResolve(opts)
	if err != nil {
		return commandError(stdout, stderr, opts.json, 1, "resolve-error", err.Error())
	}
	if printResult(stdout, stderr, report, opts.json, nil) != 0 {
		return 1
	}
	return code
}

func validateDDRResolveOptions(opts ddrResolveOptions) error {
	if opts.runID == "" || opts.sliceID == "" || opts.proofCommand == "" {
		return fmt.Errorf("ddr resolve requires run, slice, and proof bindings")
	}
	if opts.proofResult != string(modelrouting.ProofPass) && opts.proofResult != string(modelrouting.ProofFail) {
		return fmt.Errorf("proof-result must be pass or fail")
	}
	if !proofHashPattern.MatchString(opts.proofArtifactHash) {
		return fmt.Errorf("proof-artifact-hash must be sha256:<64 lowercase hex>")
	}
	return nil
}

func executeDDRAttempt(opts ddrAttemptOptions) (ddrAttemptReport, int, error) {
	start := time.Now()
	projectID, err := modelrouting.CanonicalProjectIdentity(opts.projectRoot)
	if err != nil {
		return ddrAttemptReport{}, 1, fmt.Errorf("canonical project identity: %w", err)
	}
	request, err := loadDDRRequest(opts.projectRoot, opts.requestPath)
	if err != nil {
		return ddrAttemptReport{}, 1, err
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return ddrAttemptReport{}, 1, err
	}
	requestHash := modelrouting.SHA256Bytes(requestBytes)
	catalog, err := loadUserCatalog(opts.userRoot)
	if err != nil {
		return ddrAttemptReport{}, 1, err
	}
	route, found := findUserRoute(catalog.Routes, opts.alias)
	policy, err := policyContextForProject(opts.userRoot, opts.projectRoot)
	if err != nil {
		return ddrAttemptReport{}, 1, err
	}
	routeFingerprint := ""
	if found {
		routeFingerprint, err = modelrouting.ApprovalRouteFingerprint(route, policy.RouteSources)
		if err != nil {
			return ddrAttemptReport{}, 1, err
		}
	}
	binding, err := ddrAttemptBindingHash(opts, projectID, requestHash, routeFingerprint)
	if err != nil {
		return ddrAttemptReport{}, 1, err
	}
	receiptRoot, receiptFile := ddrAttemptReceiptLocation(opts.userRoot, projectID, opts.runID, opts.sliceID)
	lockName := "ddr-" + sha256Text(projectID + "\x00" + opts.runID + "\x00" + opts.sliceID)[:24] + ".lock"
	lock, err := modelrouting.AcquirePrivateStateLock(opts.userRoot, lockName, opts.probeTimeout+opts.timeout+time.Second)
	if err != nil {
		return ddrAttemptReport{}, 1, err
	}
	defer lock.Close()
	var existing ddrAttemptReport
	if err := modelrouting.LoadStrictJSON(receiptRoot, receiptFile, &existing, maxCatalogBytes); err == nil {
		if existing.SchemaVersion != ddrAttemptSchemaVersion || existing.ProjectID != projectID || existing.AttemptBinding != binding {
			return ddrAttemptReport{}, 1, fmt.Errorf("DDR receipt binding conflict for project/run/slice")
		}
		if existing.Status == "attempting" || (existing.Status == "awaiting-proof" && existing.Response == "") {
			failure := "attempt-state-uncertain"
			if existing.Status == "awaiting-proof" {
				failure = "result-not-retained"
			}
			existing.Status, existing.Action = "parent-return", "continue-parent"
			if existing.RequirePin {
				existing.Status, existing.Action = "blocked", "block-required-route"
			}
			existing.FailureClass, existing.FailurePhase = failure, "replay"
			existing.ResultDelivery, existing.Response = "none", ""
			if err := modelrouting.SaveAtomicJSON(receiptRoot, receiptFile, existing, maxCatalogBytes); err != nil {
				return ddrAttemptReport{}, 1, err
			}
			return existing, exitCodeForDDRReport(existing), nil
		}
		return existing, exitCodeForDDRReport(existing), nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return ddrAttemptReport{}, 1, err
	}
	report := ddrAttemptReport{
		SchemaVersion: ddrAttemptSchemaVersion, ProjectID: projectID, RunID: opts.runID, SliceID: opts.sliceID,
		RouteAlias: opts.alias, RouteFingerprint: routeFingerprint, RequestHash: requestHash, AttemptBinding: binding,
		ParentSelection: "active-parent-or-host-native", RequirePin: opts.require,
		ObservedAt: time.Now().UTC(),
	}
	finish := func(status, action, failure, phase string, code int) (ddrAttemptReport, int, error) {
		report.Status, report.Action, report.FailureClass, report.FailurePhase = status, action, failure, phase
		report.TotalLatencyMS = elapsedMilliseconds(start)
		report.ResultDelivery = "none"
		if status == "awaiting-proof" {
			report.ResultDelivery = "first-response-only"
		}
		persisted := report
		persisted.Response = ""
		if err := modelrouting.SaveAtomicJSON(receiptRoot, receiptFile, persisted, maxCatalogBytes); err != nil {
			return ddrAttemptReport{}, 1, err
		}
		return report, code, nil
	}
	blockOrReturn := func(failure, phase string) (ddrAttemptReport, int, error) {
		if opts.require {
			return finish("blocked", "block-required-route", failure, phase, ddrBlockedExitCode)
		}
		return finish("parent-return", "continue-parent", failure, phase, ddrParentReturnExitCode)
	}

	if !found {
		return blockOrReturn("route-unavailable", "eligibility")
	}
	if !routeProjectSelectable(route, policy) || !routeProjectApproved(route, policy, time.Now()) || !routeAuthApproved(route, policy, time.Now()) {
		return blockOrReturn("untrusted", "eligibility")
	}
	work := modelrouting.WorkRequest{
		PlannedTier: modelrouting.Tier(opts.tier), ExecutionOwner: modelrouting.ExecutionOwnerDelegated,
		OwnerReason: "bounded one-attempt DDR", TierReason: opts.tierReason, TaskFamily: opts.taskFamily,
		Tools: []string(opts.tools), ContextSize: opts.contextSize, Risk: modelrouting.RiskNormal,
		SensitiveData: opts.sensitive, ProjectID: projectID,
	}
	if !eligibleImportedDDRRoute(route, work, policy) {
		return blockOrReturn("route-ineligible", "eligibility")
	}
	token := ""
	if route.AuthEnv != "" {
		token = os.Getenv(route.AuthEnv)
		if token == "" {
			return blockOrReturn("auth-unavailable", "eligibility")
		}
	}
	report.Attempt = 1
	report.Status = "attempting"
	report.Action = "await-local-attempt"
	report.ResultDelivery = "none"
	if err := modelrouting.SaveAtomicJSON(receiptRoot, receiptFile, report, maxCatalogBytes); err != nil {
		return ddrAttemptReport{}, 1, err
	}
	probeStart := time.Now()
	probeContext, cancelProbe := context.WithTimeout(context.Background(), opts.probeTimeout)
	endpoint, endpointErr := modelrouting.ValidateEndpointContext(probeContext, route, policy, nil, time.Now())
	if endpointErr != nil {
		cancelProbe()
		report.ProbeLatencyMS = elapsedMilliseconds(probeStart)
		return blockOrReturn(classifyDDRFailure(endpointErr), "probe")
	}
	if token != "" && endpoint.URL.Scheme != "https" && !allLoopbackIPs(endpoint.PinnedIPs) {
		cancelProbe()
		report.ProbeLatencyMS = elapsedMilliseconds(probeStart)
		return blockOrReturn("endpoint-unavailable", "probe")
	}
	models, probeErr := fetchOpenAICompatibleModels(probeContext, endpoint, route, token, maxCatalogBytes)
	cancelProbe()
	report.ProbeLatencyMS = elapsedMilliseconds(probeStart)
	if probeErr != nil {
		return blockOrReturn(classifyDDRFailure(probeErr), "probe")
	}
	if !containsString(models, route.DisplayModelID) {
		return blockOrReturn("model-missing", "probe")
	}
	if !ddrRouteStillApproved(opts.userRoot, opts.projectRoot, route, routeFingerprint, work) {
		return blockOrReturn("untrusted", "dispatch")
	}

	dispatchStart := time.Now()
	dispatchContext, cancelDispatch := context.WithTimeout(context.Background(), opts.timeout)
	response, dispatchErr := dispatchOpenAICompatible(dispatchContext, endpoint, route, token, request)
	cancelDispatch()
	report.DispatchLatencyMS = elapsedMilliseconds(dispatchStart)
	if dispatchErr != nil {
		return blockOrReturn(classifyDDRFailure(dispatchErr), "dispatch")
	}
	report.Response = response
	return finish("awaiting-proof", "run-proof", "", "", 0)
}

func executeDDRResolve(opts ddrResolveOptions) (ddrAttemptReport, int, error) {
	projectID, err := modelrouting.CanonicalProjectIdentity(opts.projectRoot)
	if err != nil {
		return ddrAttemptReport{}, 1, fmt.Errorf("canonical project identity: %w", err)
	}
	receiptRoot, receiptFile := ddrAttemptReceiptLocation(opts.userRoot, projectID, opts.runID, opts.sliceID)
	lockName := "ddr-" + sha256Text(projectID + "\x00" + opts.runID + "\x00" + opts.sliceID)[:24] + ".lock"
	lock, err := modelrouting.AcquirePrivateStateLock(opts.userRoot, lockName, ddrMaxProbeTimeout+ddrMaxTimeout+time.Second)
	if err != nil {
		return ddrAttemptReport{}, 1, err
	}
	defer lock.Close()
	var report ddrAttemptReport
	if err := modelrouting.LoadStrictJSON(receiptRoot, receiptFile, &report, maxCatalogBytes); err != nil {
		return ddrAttemptReport{}, 1, err
	}
	if report.SchemaVersion != ddrAttemptSchemaVersion || report.ProjectID != projectID ||
		report.RunID != opts.runID || report.SliceID != opts.sliceID {
		return ddrAttemptReport{}, 1, fmt.Errorf("DDR receipt binding conflict for project/run/slice")
	}
	incomingProof := modelrouting.WorkProof{
		Command: opts.proofCommand, ArtifactHash: opts.proofArtifactHash,
		Result: modelrouting.ProofResult(opts.proofResult),
	}
	switch report.Status {
	case "completed", "parent-return", "blocked":
		if report.Proof.Command != "" && report.Proof != incomingProof {
			return ddrAttemptReport{}, 1, fmt.Errorf("DDR proof replay conflict for project/run/slice")
		}
		return report, exitCodeForDDRReport(report), nil
	case "attempting":
		report.Status = "parent-return"
		report.Action = "continue-parent"
		if report.RequirePin {
			report.Status = "blocked"
			report.Action = "block-required-route"
		}
		report.FailureClass = "attempt-state-uncertain"
		report.FailurePhase = "resolve"
	case "awaiting-proof":
		report.Proof = incomingProof
		if opts.proofResult == string(modelrouting.ProofPass) {
			report.Status = "completed"
			report.Action = "accept-result"
			report.FailureClass = ""
			report.FailurePhase = ""
		} else if report.RequirePin {
			report.Status = "blocked"
			report.Action = "block-required-route"
			report.FailureClass = "proof-failed"
			report.FailurePhase = "proof"
		} else {
			report.Status = "parent-return"
			report.Action = "continue-parent"
			report.FailureClass = "proof-failed"
			report.FailurePhase = "proof"
		}
	default:
		return ddrAttemptReport{}, 1, fmt.Errorf("invalid DDR receipt status %q", report.Status)
	}
	report.Response = ""
	report.ResultDelivery = "none"
	if err := modelrouting.SaveAtomicJSON(receiptRoot, receiptFile, report, maxCatalogBytes); err != nil {
		return ddrAttemptReport{}, 1, err
	}
	return report, exitCodeForDDRReport(report), nil
}

func loadDDRRequest(projectRoot, path string) (ddrRequestFile, error) {
	projectAbs, err := filepath.Abs(filepath.Clean(projectRoot))
	if err != nil {
		return ddrRequestFile{}, err
	}
	requestAbs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return ddrRequestFile{}, err
	}
	rel, err := filepath.Rel(projectAbs, requestAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return ddrRequestFile{}, modelrouting.ErrUnsafePath
	}
	var request ddrRequestFile
	if err := modelrouting.LoadStrictProjectJSON(projectAbs, rel, &request, maxCatalogBytes); err != nil {
		return ddrRequestFile{}, err
	}
	if request.SchemaVersion != 1 || len(request.Messages) == 0 || len(request.Messages) > 64 ||
		request.MaxTokens <= 0 || request.MaxTokens > 4096 {
		return ddrRequestFile{}, fmt.Errorf("invalid DDR request envelope")
	}
	for _, message := range request.Messages {
		if (message.Role != "system" && message.Role != "user" && message.Role != "assistant") ||
			strings.TrimSpace(message.Content) == "" || len(message.Content) > 64*1024 ||
			strings.Contains(strings.ToLower(message.Content), "<fill-") {
			return ddrRequestFile{}, fmt.Errorf("invalid DDR request message")
		}
	}
	return request, nil
}

func dispatchOpenAICompatible(ctx context.Context, endpoint modelrouting.ValidatedEndpoint, route modelrouting.Route, token string, request ddrRequestFile) (string, error) {
	target := *endpoint.URL
	target.Path = strings.TrimRight(target.Path, "/") + "/chat/completions"
	target.RawQuery = ""
	payload, err := json.Marshal(map[string]any{
		"model": route.DisplayModelID, "messages": request.Messages,
		"max_tokens": request.MaxTokens, "stream": false,
	})
	if err != nil {
		return "", err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	if token != "" {
		httpRequest.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := newOpenAICompatibleClient(endpoint).Do(httpRequest)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", ddrHTTPStatusError{StatusCode: response.StatusCode, Phase: "dispatch"}
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxCatalogBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxCatalogBytes {
		return "", modelrouting.ErrStorageSizeExceeded
	}
	var payloadResponse struct {
		Model   string `json:"model"`
		Choices []struct {
			Message ddrMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &payloadResponse); err != nil {
		return "", fmt.Errorf("dispatch response decode: %w", err)
	}
	if payloadResponse.Model != route.DisplayModelID || len(payloadResponse.Choices) != 1 ||
		strings.TrimSpace(payloadResponse.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("dispatch response lacked exact model attribution and one bounded result")
	}
	return payloadResponse.Choices[0].Message.Content, nil
}

func newOpenAICompatibleClient(endpoint modelrouting.ValidatedEndpoint) *http.Client {
	transport := &http.Transport{
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				host = address
			}
			if !strings.EqualFold(host, endpoint.URL.Hostname()) {
				return nil, fmt.Errorf("cross-origin dial refused")
			}
			if port == "" {
				if endpoint.URL.Scheme == "https" {
					port = "443"
				} else {
					port = "80"
				}
			}
			dialer := net.Dialer{}
			var lastErr error
			for _, ip := range endpoint.PinnedIPs {
				conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
				if dialErr == nil {
					return conn, nil
				}
				lastErr = dialErr
			}
			return nil, lastErr
		},
		TLSClientConfig: &tls.Config{ServerName: endpoint.TLSServerName, MinVersion: tls.VersionTLS12},
	}
	return &http.Client{
		Transport:     transport,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

func classifyDDRFailure(err error) string {
	var statusErr ddrHTTPStatusError
	switch {
	case errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusUnauthorized:
		return "unauthorized"
	case errors.As(err, &statusErr) && statusErr.StatusCode >= 500:
		return "server-error"
	case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "deadline exceeded"):
		return "timeout"
	case errors.Is(err, modelrouting.ErrPrivateEndpointRequiresApproval):
		return "untrusted"
	case errors.Is(err, modelrouting.ErrUnsafeEndpoint):
		return "endpoint-unavailable"
	default:
		if strings.Contains(err.Error(), "GET /v1/models returned 401") {
			return "unauthorized"
		}
		if strings.Contains(err.Error(), "GET /v1/models returned 5") {
			return "server-error"
		}
		if strings.Contains(err.Error(), "dispatch") {
			return "dispatch-failed"
		}
		return "probe-failed"
	}
}

func eligibleImportedDDRRoute(route modelrouting.Route, request modelrouting.WorkRequest, policy modelrouting.PolicyContext) bool {
	if route.Hosting != modelrouting.HostingSelfHosted ||
		route.Adapter != "openai-compatible" || route.DispatchMethod != "chat-completions" ||
		route.Capability.RouteAlias != route.Alias || route.Capability.ModelID != route.DisplayModelID ||
		route.Capability.TaskFamily != request.TaskFamily || route.Capability.ContextSize < request.ContextSize ||
		route.Capability.Source != modelrouting.EvidenceDeclared || route.Capability.DispatchProven ||
		dispatchClassRank(route.Capability.Class) < dispatchTierRank(request.PlannedTier) {
		return false
	}
	if route.Capability.Risk != modelrouting.RiskNormal && route.Capability.Risk != modelrouting.RiskBroad {
		return false
	}
	for _, tool := range request.Tools {
		if !containsString(route.Capability.Tools, tool) {
			return false
		}
	}
	if route.Retention == modelrouting.RetentionUnknown || route.Retention == "" ||
		(request.SensitiveData && (route.TrainingUse != modelrouting.TrainingNo ||
			strings.TrimSpace(route.Residency) == "" || strings.EqualFold(route.Residency, "unknown") ||
			strings.TrimSpace(route.TrustProvenance) == "")) {
		return false
	}
	return modelrouting.RouteAllowedByPolicy(route, request, policy, time.Now())
}

func ddrAttemptBindingHash(opts ddrAttemptOptions, projectID, requestHash, routeFingerprint string) (string, error) {
	binding := ddrAttemptBinding{
		ProjectID: projectID, RunID: opts.runID, SliceID: opts.sliceID,
		RouteAlias: opts.alias, RouteFingerprint: routeFingerprint, RequestHash: requestHash,
		Tier: opts.tier, TierReason: opts.tierReason, TaskFamily: opts.taskFamily,
		Tools: sortedUnique([]string(opts.tools)), ContextSize: opts.contextSize, Risk: opts.risk,
		SensitiveData: opts.sensitive, RequirePin: opts.require,
		ProbeTimeout: opts.probeTimeout, DispatchTimeout: opts.timeout,
	}
	data, err := json.Marshal(binding)
	if err != nil {
		return "", err
	}
	return modelrouting.SHA256Bytes(data), nil
}

func routeProjectApproved(route modelrouting.Route, policy modelrouting.PolicyContext, now time.Time) bool {
	fingerprint, err := modelrouting.ApprovalRouteFingerprint(route, policy.RouteSources)
	if err != nil {
		return false
	}
	for _, approval := range policy.Trusted.RouteApprovals {
		if approval.ProjectID == policy.Project.ProjectID && approval.RouteFingerprint == fingerprint && now.Before(approval.ExpiresAt) {
			return true
		}
	}
	return false
}

func ddrRouteStillApproved(userRoot, projectRoot string, expected modelrouting.Route, expectedFingerprint string, work modelrouting.WorkRequest) bool {
	catalog, err := loadUserCatalog(userRoot)
	if err != nil {
		return false
	}
	route, found := findUserRoute(catalog.Routes, expected.Alias)
	if !found {
		return false
	}
	policy, err := policyContextForProject(userRoot, projectRoot)
	if err != nil {
		return false
	}
	fingerprint, err := modelrouting.ApprovalRouteFingerprint(route, policy.RouteSources)
	if err != nil || fingerprint != expectedFingerprint {
		return false
	}
	now := time.Now()
	return routeProjectSelectable(route, policy) &&
		routeProjectApproved(route, policy, now) &&
		routeAuthApproved(route, policy, now) &&
		eligibleImportedDDRRoute(route, work, policy)
}

func allLoopbackIPs(values []net.IP) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if !value.IsLoopback() {
			return false
		}
	}
	return true
}

func ddrAttemptReceiptLocation(userRoot, projectID, runID, sliceID string) (string, string) {
	projectHash := sha256Text(projectID)[:24]
	attemptHash := sha256Text(runID + "\x00" + sliceID)[:32]
	return filepath.Join(userRoot, "ddr-attempts", projectHash), attemptHash + ".json"
}

func findUserRoute(routes []modelrouting.Route, alias string) (modelrouting.Route, bool) {
	for _, route := range routes {
		if route.Alias == alias {
			return route, true
		}
	}
	return modelrouting.Route{}, false
}

func elapsedMilliseconds(start time.Time) int64 {
	value := time.Since(start).Milliseconds()
	if value < 0 {
		return 0
	}
	return value
}

func exitCodeForDDRReport(report ddrAttemptReport) int {
	switch report.Status {
	case "completed":
		return 0
	case "parent-return":
		return ddrParentReturnExitCode
	case "blocked":
		return ddrBlockedExitCode
	default:
		return 1
	}
}
