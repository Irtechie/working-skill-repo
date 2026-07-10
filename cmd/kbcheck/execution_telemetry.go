package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

const maxExecutionTelemetryBytes int64 = 1 << 20

var executionAttestationRoot = func() (string, error) {
	account, err := user.Current()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(account.HomeDir) == "" {
		return "", fmt.Errorf("current user home is unavailable")
	}
	return filepath.Join(account.HomeDir, ".kb"), nil
}

var executionTelemetryNow = func() time.Time { return time.Now().UTC() }

type executionTelemetry struct {
	PacketID           string `json:"packet_id,omitempty"`
	RunID              string `json:"run_id,omitempty"`
	ProjectID          string `json:"project_id,omitempty"`
	SliceID            string `json:"slice_id,omitempty"`
	ContextPacketHash  string `json:"context_packet_hash,omitempty"`
	SessionID          string `json:"session_id,omitempty"`
	Runtime            string `json:"runtime,omitempty"`
	RequestedRoute     string `json:"requested_route,omitempty"`
	ActualRoute        string `json:"actual_route,omitempty"`
	RequestedModel     string `json:"requested_model,omitempty"`
	ActualModel        string `json:"actual_model,omitempty"`
	Model              string `json:"model,omitempty"`
	ReceiptStatus      string `json:"receipt_status,omitempty"`
	PredictedTier      string `json:"predicted_tier,omitempty"`
	ActualTier         string `json:"actual_tier,omitempty"`
	Turns              int64  `json:"turns,omitempty"`
	InputTokens        int64  `json:"input_tokens,omitempty"`
	OutputTokens       int64  `json:"output_tokens,omitempty"`
	CacheReadTokens    int64  `json:"cache_read_tokens,omitempty"`
	CacheWriteTokens   int64  `json:"cache_write_tokens,omitempty"`
	ReworkCount        int64  `json:"rework_count,omitempty"`
	EscalationReason   string `json:"escalation_reason,omitempty"`
	ProofResult        string `json:"proof_result,omitempty"`
	PacketSufficiency  string `json:"packet_sufficiency,omitempty"`
	EffectiveTokenMode string `json:"effective_token_model,omitempty"`
}

type executionTelemetryEnvelopeFile struct {
	RunID             string   `json:"run_id"`
	SliceID           string   `json:"slice_id"`
	ProjectID         string   `json:"project_id"`
	RouteAlias        string   `json:"route_alias"`
	RouteFingerprint  string   `json:"route_fingerprint"`
	Adapter           string   `json:"adapter"`
	AdapterRevision   string   `json:"adapter_revision"`
	DispatchMethod    string   `json:"dispatch_method"`
	ModelID           string   `json:"model_id"`
	TaskFamily        string   `json:"task_family"`
	ContextPacketID   string   `json:"context_packet_id,omitempty"`
	ContextPacketHash string   `json:"context_packet_hash"`
	ProofArtifactHash string   `json:"proof_artifact_hash"`
	ProofArtifactPath string   `json:"proof_artifact_path"`
	Surface           string   `json:"surface"`
	Provider          string   `json:"provider"`
	Tools             []string `json:"tools"`
	ContextSize       int      `json:"context_size"`
	Risk              string   `json:"risk"`
}

type executionTelemetryEvidence struct {
	ReceiptStatus     string
	PacketID          string
	RunID             string
	ProjectID         string
	SliceID           string
	ContextPacketHash string
	ActualRoute       string
	ActualModel       string
	SessionID         string
	ProofResult       string
}

func normalizeExecutionTelemetry(raw any, evidence *executionTelemetryEvidence) (*executionTelemetry, error) {
	if raw == nil {
		return nil, nil
	}
	content, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("encode usage telemetry: %w", err)
	}
	var usage executionTelemetry
	if err := json.Unmarshal(content, &usage); err != nil {
		return nil, fmt.Errorf("decode usage telemetry: %w", err)
	}
	if usage.Turns < 0 || usage.InputTokens < 0 || usage.OutputTokens < 0 || usage.CacheReadTokens < 0 || usage.CacheWriteTokens < 0 || usage.ReworkCount < 0 {
		return nil, fmt.Errorf("usage telemetry counters must be non-negative")
	}
	if strings.TrimSpace(usage.PacketID) == "" {
		return nil, fmt.Errorf("usage packet_id is required")
	}
	if usage.PredictedTier != "" && !validModelTier(usage.PredictedTier) {
		return nil, fmt.Errorf("usage predicted_tier is invalid")
	}
	if usage.ActualTier != "" && !validModelTier(usage.ActualTier) {
		return nil, fmt.Errorf("usage actual_tier is invalid")
	}
	if usage.Model != "" && usage.RequestedModel == "" {
		usage.RequestedModel = usage.Model
	}
	if usage.ReceiptStatus != "" && !contains([]string{"credited", "observation-only", "missing", "mismatch", "unavailable"}, usage.ReceiptStatus) {
		return nil, fmt.Errorf("usage receipt_status is invalid")
	}
	if usage.ProofResult != "" && !contains([]string{"pass", "fail", "skipped", "unknown"}, usage.ProofResult) {
		return nil, fmt.Errorf("usage proof_result is invalid")
	}
	if usage.PacketSufficiency != "" && !contains([]string{"sufficient", "insufficient", "unknown"}, usage.PacketSufficiency) {
		return nil, fmt.Errorf("usage packet_sufficiency is invalid")
	}
	if usage.EffectiveTokenMode != "" && usage.EffectiveTokenMode != "raw-v1" {
		return nil, fmt.Errorf("usage effective_token_model must be raw-v1")
	}
	applyExecutionTelemetryEvidence(&usage, evidence)
	return &usage, nil
}

func runExecutionTelemetryCommand(root string, opts options, stdout, stderr io.Writer) int {
	raw, err := loadExecutionTelemetryPayload(root, opts.telemetryPath)
	if err != nil {
		if opts.json {
			writeJSON(stdout, map[string]any{"ok": false, "issues": []string{err.Error()}})
		} else {
			fmt.Fprintln(stderr, err)
		}
		return 1
	}
	evidence, err := loadExecutionTelemetryEvidence(root, opts)
	if err != nil {
		if opts.json {
			writeJSON(stdout, map[string]any{"ok": false, "issues": []string{err.Error()}})
		} else {
			fmt.Fprintln(stderr, err)
		}
		return 2
	}
	usage, err := normalizeExecutionTelemetry(raw, evidence)
	if err != nil {
		if opts.json {
			writeJSON(stdout, map[string]any{"ok": false, "issues": []string{err.Error()}})
		} else {
			fmt.Fprintln(stderr, err)
		}
		return 2
	}
	if opts.json {
		writeJSON(stdout, map[string]any{"ok": true, "telemetry": usage})
	} else {
		fmt.Fprintf(stdout, "execution telemetry: ok packet=%s runtime=%s\n", usage.PacketID, usage.Runtime)
	}
	return 0
}

func runExecutionTelemetrySelftest(stdout, stderr io.Writer) int {
	valid := map[string]any{
		"packet_id": "p1", "run_id": "run-1", "project_id": "project-a", "slice_id": "slice-1", "context_packet_hash": "sha256:packet", "session_id": "forged-session",
		"runtime": "ghcp", "requested_route": "medium-a", "actual_route": "forged-route",
		"requested_model": "model-a", "actual_model": "forged-model", "receipt_status": "credited",
		"predicted_tier": "small", "actual_tier": "small", "turns": 2, "input_tokens": 100,
		"proof_result": "pass", "packet_sufficiency": "sufficient",
		"effective_token_model": "raw-v1",
	}
	usage, err := normalizeExecutionTelemetry(valid, nil)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if usage.ReceiptStatus != "missing" || usage.ActualRoute != "" || usage.ActualModel != "" || usage.SessionID != "" || usage.ProofResult != "unknown" {
		fmt.Fprintf(stderr, "plain telemetry was not downgraded: %#v\n", usage)
		return 1
	}
	evidence, err := evaluateExecutionTelemetryEvidence(exactExecutionTelemetryReceipt(time.Now().UTC()))
	if err != nil || evidence == nil || evidence.ReceiptStatus != "credited" {
		fmt.Fprintf(stderr, "valid receipt evidence failed: evidence=%#v err=%v\n", evidence, err)
		return 1
	}
	usage, err = normalizeExecutionTelemetry(valid, evidence)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if usage.ReceiptStatus != "credited" || usage.ActualRoute != "medium-a" || usage.ActualModel != "model-a" || usage.SessionID != "session-1" || usage.ProofResult != "pass" {
		fmt.Fprintf(stderr, "receipt evidence was not applied: %#v\n", usage)
		return 1
	}
	valid["input_tokens"] = -1
	if _, err := normalizeExecutionTelemetry(valid, evidence); err == nil {
		fmt.Fprintln(stderr, "negative telemetry accepted")
		return 1
	}
	fmt.Fprintln(stdout, "execution telemetry selftest passed")
	return 0
}

func applyExecutionTelemetryEvidence(usage *executionTelemetry, evidence *executionTelemetryEvidence) {
	if evidence == nil {
		usage.RunID = ""
		usage.ProjectID = ""
		usage.SliceID = ""
		usage.ContextPacketHash = ""
		usage.SessionID = ""
		usage.ActualRoute = ""
		usage.ActualModel = ""
		usage.ReceiptStatus = "missing"
		usage.ProofResult = "unknown"
		return
	}
	if evidence.ReceiptStatus == "mismatch" || evidence.PacketID == "" ||
		usage.PacketID != evidence.PacketID ||
		usage.RunID == "" || usage.RunID != evidence.RunID ||
		usage.ProjectID == "" || usage.ProjectID != evidence.ProjectID ||
		usage.SliceID == "" || usage.SliceID != evidence.SliceID ||
		usage.ContextPacketHash == "" || usage.ContextPacketHash != evidence.ContextPacketHash {
		usage.RunID = ""
		usage.ProjectID = ""
		usage.SliceID = ""
		usage.ContextPacketHash = ""
		usage.SessionID = ""
		usage.ActualRoute = ""
		usage.ActualModel = ""
		usage.ReceiptStatus = "mismatch"
		usage.ProofResult = "unknown"
		return
	}
	usage.RunID = evidence.RunID
	usage.ProjectID = evidence.ProjectID
	usage.SliceID = evidence.SliceID
	usage.ContextPacketHash = evidence.ContextPacketHash
	usage.SessionID = evidence.SessionID
	usage.ActualRoute = evidence.ActualRoute
	usage.ActualModel = evidence.ActualModel
	usage.ReceiptStatus = evidence.ReceiptStatus
	usage.ProofResult = evidence.ProofResult
}

func loadExecutionTelemetryPayload(root, inputPath string) (any, error) {
	path, err := resolveExecutionTelemetryPath(root, inputPath)
	if err != nil {
		return nil, err
	}
	var raw executionTelemetry
	if err := readBoundedJSONFile(path, maxExecutionTelemetryBytes, &raw, true); err != nil {
		return nil, err
	}
	return &raw, nil
}

func loadExecutionTelemetryEvidence(root string, opts options) (*executionTelemetryEvidence, error) {
	if opts.receiptPath == "" && opts.evidenceEnvelopePath == "" {
		return nil, nil
	}
	if opts.receiptPath == "" || opts.evidenceEnvelopePath == "" {
		return nil, fmt.Errorf("execution-telemetry requires --receipt and --evidence-envelope together")
	}
	runRoot, receiptName, envelopeName, err := resolveExecutionEvidenceRun(root, opts.receiptPath, opts.evidenceEnvelopePath)
	if err != nil {
		return nil, err
	}
	var receipt modelrouting.RoutingReceipt
	receiptBytes, err := modelrouting.LoadStrictJSONBytes(runRoot, receiptName, &receipt, maxExecutionTelemetryBytes)
	if err != nil {
		return nil, fmt.Errorf("load receipt: %w", err)
	}
	var stored executionTelemetryEnvelopeFile
	_, err = modelrouting.LoadStrictJSONBytes(runRoot, envelopeName, &stored, maxExecutionTelemetryBytes)
	if err != nil {
		return nil, fmt.Errorf("load evidence envelope: %w", err)
	}
	if err := validateExecutionEvidenceRunMarker(root, runRoot, stored); err != nil {
		return &executionTelemetryEvidence{ReceiptStatus: "mismatch", ProofResult: "unknown"}, nil
	}
	proofHash, err := hashExecutionProofArtifact(runRoot, stored.ProofArtifactPath)
	if err != nil {
		return nil, err
	}
	if proofHash != stored.ProofArtifactHash || proofHash != receipt.WorkProof.ArtifactHash {
		return &executionTelemetryEvidence{ReceiptStatus: "mismatch", ProofResult: "unknown"}, nil
	}
	convertedReceipt, envelope, err := convertExecutionTelemetryEvidence(receipt, stored)
	if err != nil {
		return nil, err
	}
	evidence, err := evaluateExecutionTelemetryEvidence(convertedReceipt, envelope)
	if err != nil || evidence == nil || evidence.ReceiptStatus == "mismatch" {
		return evidence, err
	}
	attested, err := verifyDispatchReceiptAttestation(root, receipt, modelrouting.SHA256Bytes(receiptBytes))
	if err != nil {
		return nil, err
	}
	if !attested {
		return &executionTelemetryEvidence{ReceiptStatus: "mismatch", ProofResult: "unknown"}, nil
	}
	return evidence, nil
}

func verifyDispatchReceiptAttestation(projectRoot string, receipt modelrouting.RoutingReceipt, receiptHash string) (bool, error) {
	root, err := executionAttestationRoot()
	if err != nil {
		return false, err
	}
	root, err = canonicalPathForContainment(root)
	if err != nil {
		return false, err
	}
	projectPath, err := canonicalPathForContainment(projectRoot)
	if err != nil {
		return false, err
	}
	relative, relErr := filepath.Rel(projectPath, root)
	if relErr == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false, fmt.Errorf("host execution attestation root must remain outside the project")
	}
	return modelrouting.VerifyDispatchReceiptAttestation(root, receipt, receiptHash, executionTelemetryNow())
}

func canonicalPathForContainment(path string) (string, error) {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	probe := abs
	tail := []string{}
	for {
		_, statErr := os.Lstat(probe)
		if statErr == nil {
			break
		}
		if !os.IsNotExist(statErr) {
			return "", statErr
		}
		parent := filepath.Dir(probe)
		if parent == probe {
			return "", statErr
		}
		tail = append(tail, filepath.Base(probe))
		probe = parent
	}
	canonical, err := filepath.EvalSymlinks(probe)
	if err != nil {
		return "", err
	}
	for index := len(tail) - 1; index >= 0; index-- {
		canonical = filepath.Join(canonical, tail[index])
	}
	return filepath.Abs(filepath.Clean(canonical))
}

func resolveExecutionEvidenceRun(root, receiptPath, envelopePath string) (string, string, string, error) {
	rootPath, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", "", "", err
	}
	if canonical, canonicalErr := filepath.EvalSymlinks(rootPath); canonicalErr == nil {
		rootPath = canonical
	}
	receiptAbs, receiptRel, err := resolveExecutionEvidenceChild(rootPath, receiptPath)
	if err != nil {
		return "", "", "", fmt.Errorf("receipt path: %w", err)
	}
	envelopeAbs, envelopeRel, err := resolveExecutionEvidenceChild(rootPath, envelopePath)
	if err != nil {
		return "", "", "", fmt.Errorf("evidence envelope path: %w", err)
	}
	receiptDir := filepath.Dir(receiptRel)
	envelopeDir := filepath.Dir(envelopeRel)
	if receiptDir != envelopeDir {
		return "", "", "", fmt.Errorf("receipt and evidence envelope must be direct children of the same marked run root")
	}
	parts := strings.Split(receiptDir, string(filepath.Separator))
	if len(parts) != 3 || parts[0] != ".kb" || parts[1] != "runs" || parts[2] == "" || strings.Contains(parts[2], "..") {
		return "", "", "", fmt.Errorf("receipt and evidence envelope must be direct children of .kb/runs/<run-id>")
	}
	if filepath.Dir(receiptAbs) != filepath.Dir(envelopeAbs) {
		return "", "", "", fmt.Errorf("receipt and evidence envelope must be direct children of the same marked run root")
	}
	return filepath.Dir(receiptAbs), filepath.Base(receiptAbs), filepath.Base(envelopeAbs), nil
}

func resolveExecutionEvidenceChild(rootPath, input string) (string, string, error) {
	if filepath.IsAbs(input) {
		return "", "", fmt.Errorf("evidence path must be repository-relative")
	}
	resolved, err := filepath.Abs(filepath.Join(rootPath, filepath.FromSlash(input)))
	if err != nil {
		return "", "", err
	}
	rel, err := filepath.Rel(rootPath, resolved)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", "", fmt.Errorf("execution-telemetry evidence must resolve under %s", rootPath)
	}
	if filepath.Base(rel) == rel {
		return "", "", fmt.Errorf("execution-telemetry evidence must live under .kb/runs/<run-id>")
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return "", "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return "", "", fmt.Errorf("execution-telemetry evidence must be a regular file")
	}
	return resolved, rel, nil
}

func validateExecutionEvidenceRunMarker(projectRoot, runRoot string, stored executionTelemetryEnvelopeFile) error {
	var marker struct {
		SchemaVersion int    `json:"schema_version"`
		ProjectID     string `json:"project_id"`
		RunID         string `json:"run_id"`
	}
	if err := modelrouting.LoadStrictJSON(runRoot, ".kb-run-root.json", &marker, maxExecutionTelemetryBytes); err != nil {
		return err
	}
	projectID, err := modelrouting.CanonicalProjectIdentity(projectRoot)
	if err != nil {
		return err
	}
	runID := filepath.Base(runRoot)
	if marker.SchemaVersion != 1 || marker.ProjectID != projectID || marker.RunID != runID ||
		stored.ProjectID != projectID || stored.RunID != runID || stored.SliceID == "" ||
		stored.ContextPacketID == "" || stored.ContextPacketHash == "" {
		return fmt.Errorf("evidence bindings do not match marked run root")
	}
	return nil
}

func hashExecutionProofArtifact(runRoot, rel string) (string, error) {
	if strings.TrimSpace(rel) == "" || filepath.IsAbs(rel) {
		return "", fmt.Errorf("proof artifact path must be a direct child of the marked run root")
	}
	clean := filepath.Clean(filepath.FromSlash(rel))
	if clean == "." || clean == ".." || strings.Contains(clean, string(filepath.Separator)) {
		return "", fmt.Errorf("proof artifact path must be a direct child of the marked run root")
	}
	path := filepath.Join(runRoot, clean)
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return "", modelrouting.ErrUnsafePath
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return "", modelrouting.ErrUnsafePath
	}
	data, err := io.ReadAll(io.LimitReader(file, maxExecutionTelemetryBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxExecutionTelemetryBytes {
		return "", modelrouting.ErrStorageSizeExceeded
	}
	return modelrouting.SHA256Bytes(data), nil
}

func resolveExecutionTelemetryPath(root, path string) (string, error) {
	rootPath, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", err
	}
	if canonical, canonicalErr := filepath.EvalSymlinks(rootPath); canonicalErr == nil {
		rootPath = canonical
	}
	resolved, err := filepath.Abs(filepath.Join(rootPath, filepath.FromSlash(path)))
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(path) {
		resolved, err = filepath.Abs(filepath.Clean(path))
		if err != nil {
			return "", err
		}
	}
	resolved, err = filepath.EvalSymlinks(resolved)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(rootPath, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("execution-telemetry inputs must resolve under %s", rootPath)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("execution-telemetry input must be a regular file: %s", resolved)
	}
	return resolved, nil
}

func readBoundedJSONFile(path string, maxBytes int64, value any, disallowUnknown bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() > maxBytes {
		return fmt.Errorf("%s exceeded %d bytes", filepath.Base(path), maxBytes)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	if disallowUnknown {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if decoder.More() {
		return fmt.Errorf("%s contained trailing JSON content", filepath.Base(path))
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("%s contained trailing JSON content", filepath.Base(path))
	}
	return nil
}

func convertExecutionTelemetryEvidence(receipt modelrouting.RoutingReceipt, stored executionTelemetryEnvelopeFile) (modelrouting.RoutingReceipt, modelrouting.EvidenceEnvelope, error) {
	envelope := modelrouting.EvidenceEnvelope{
		RunID:             stored.RunID,
		SliceID:           stored.SliceID,
		ProjectID:         stored.ProjectID,
		RouteAlias:        stored.RouteAlias,
		RouteFingerprint:  stored.RouteFingerprint,
		Adapter:           stored.Adapter,
		AdapterRevision:   stored.AdapterRevision,
		DispatchMethod:    stored.DispatchMethod,
		ModelID:           stored.ModelID,
		TaskFamily:        stored.TaskFamily,
		ContextPacketID:   stored.ContextPacketID,
		ContextPacketHash: stored.ContextPacketHash,
		ProofArtifactHash: stored.ProofArtifactHash,
		Surface:           stored.Surface,
		Provider:          stored.Provider,
		Tools:             append([]string(nil), stored.Tools...),
		ContextSize:       stored.ContextSize,
		Risk:              modelrouting.RiskLevel(stored.Risk),
		MaxAge:            time.Hour,
	}
	return receipt, envelope, nil
}

func evaluateExecutionTelemetryEvidence(receipt modelrouting.RoutingReceipt, envelope modelrouting.EvidenceEnvelope) (*executionTelemetryEvidence, error) {
	now := executionTelemetryNow()
	if credit, observation := modelrouting.EvaluateReceiptCredit(receipt, envelope, now); credit {
		return telemetryEvidenceFromObservation(observation), nil
	}
	probed := receipt
	if probed.WorkProof.Result != modelrouting.ProofPass {
		probed.WorkProof.Result = modelrouting.ProofPass
		if credit, _ := modelrouting.EvaluateReceiptCredit(probed, envelope, now); credit {
			return telemetryEvidenceFromObservation(modelrouting.RoutingObservation{
				Status:        modelrouting.ObservationOnly,
				RouteEvidence: receipt.RouteEvidence,
				WorkProof:     receipt.WorkProof,
				ObservedAt:    now,
			}), nil
		}
	}
	return &executionTelemetryEvidence{ReceiptStatus: "mismatch", ProofResult: "unknown"}, nil
}

func telemetryEvidenceFromObservation(observation modelrouting.RoutingObservation) *executionTelemetryEvidence {
	status := "missing"
	switch observation.Status {
	case modelrouting.ObservationCredited:
		status = "credited"
	case modelrouting.ObservationOnly:
		status = "observation-only"
	}
	return &executionTelemetryEvidence{
		ReceiptStatus:     status,
		PacketID:          observation.RouteEvidence.ContextPacketID,
		RunID:             observation.RouteEvidence.RunID,
		ProjectID:         observation.RouteEvidence.ProjectID,
		SliceID:           observation.RouteEvidence.SliceID,
		ContextPacketHash: observation.RouteEvidence.ContextPacketHash,
		ActualRoute:       observation.RouteEvidence.RouteAlias,
		ActualModel:       observation.RouteEvidence.ProviderReportedModel,
		SessionID:         observation.RouteEvidence.SessionID,
		ProofResult:       string(observation.WorkProof.Result),
	}
}

func exactExecutionTelemetryReceipt(now time.Time) (modelrouting.RoutingReceipt, modelrouting.EvidenceEnvelope) {
	envelope := modelrouting.EvidenceEnvelope{
		RunID: "run-1", SliceID: "slice-1", ProjectID: "project-a", RouteAlias: "medium-a",
		RouteFingerprint: "route-sha256:abc", Adapter: "codex", AdapterRevision: "v1",
		DispatchMethod: "named-agent", ModelID: "model-a", TaskFamily: "code",
		ContextPacketID: "p1", ContextPacketHash: "sha256:packet", ProofArtifactHash: "sha256:proof",
		Surface: "codex-cli", Provider: "openai", Tools: []string{"codex-harness"}, ContextSize: 8192, Risk: modelrouting.RiskBroad, MaxAge: time.Hour,
	}
	capabilityHash, _ := modelrouting.ComputeCapabilityEnvelopeHash(envelope)
	receipt := modelrouting.RoutingReceipt{
		RouteEvidence: modelrouting.RouteDispatchEvidence{
			RunID: envelope.RunID, SliceID: envelope.SliceID, ProjectID: envelope.ProjectID,
			RouteAlias: envelope.RouteAlias, RouteFingerprint: envelope.RouteFingerprint,
			Adapter: envelope.Adapter, AdapterRevision: envelope.AdapterRevision, DispatchMethod: envelope.DispatchMethod,
			RequestedModelID: envelope.ModelID, ProviderReportedModel: envelope.ModelID, SessionID: "session-1",
			TaskFamily: envelope.TaskFamily, ContextPacketID: envelope.ContextPacketID, ContextPacketHash: envelope.ContextPacketHash,
			CapabilityEnvelopeHash: capabilityHash, Attempt: 1, ObservedAt: now,
		},
		WorkProof: modelrouting.WorkProof{Command: "go test ./cmd/kbcheck", ArtifactHash: envelope.ProofArtifactHash, Result: modelrouting.ProofPass},
	}
	return receipt, envelope
}
