package modelrouting

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrInvalidDispatchRequest     = errors.New("invalid dispatch request")
	ErrUnsupportedDispatchAdapter = errors.New("unsupported dispatch adapter")
	ErrProfileNotUserLocal        = errors.New("profile must be user-local")
	ErrProjectProviderConfig      = errors.New("project provider/profile configuration is not trusted for dispatch")
)

type DispatchRequest struct {
	RunID              string        `json:"run_id"`
	SliceID            string        `json:"slice_id"`
	PacketID           string        `json:"packet_id"`
	ProjectID          string        `json:"project_id"`
	ProjectRoot        string        `json:"project_root"`
	CanonicalProject   string        `json:"canonical_project"`
	CWD                string        `json:"cwd"`
	Worktree           string        `json:"worktree"`
	RouteAlias         string        `json:"route_alias"`
	RouteFingerprint   string        `json:"route_fingerprint"`
	RouteState         string        `json:"route_state"`
	Adapter            string        `json:"adapter"`
	AdapterRevision    string        `json:"adapter_revision"`
	DispatchMethod     string        `json:"dispatch_method"`
	Model              string        `json:"model"`
	Profile            string        `json:"profile,omitempty"`
	ProfileRevision    string        `json:"profile_revision,omitempty"`
	PacketHash         string        `json:"packet_hash"`
	Sandbox            string        `json:"sandbox"`
	ApprovalPolicy     string        `json:"approval_policy"`
	AllowedRoots       []string      `json:"allowed_roots"`
	AllowedTools       []string      `json:"allowed_tools"`
	Network            string        `json:"network"`
	Timeout            time.Duration `json:"timeout"`
	Attempt            int           `json:"attempt"`
	OutputPath         string        `json:"output_path"`
	ReceiptPath        string        `json:"receipt_path"`
	HandoffPath        string        `json:"handoff_path"`
	WorkerRequestPath  string        `json:"worker_request_path,omitempty"`
	OutputSchemaPath   string        `json:"output_schema_path,omitempty"`
	Authority          string        `json:"authority"`
	EndpointOrigin     string        `json:"endpoint_origin,omitempty"`
	PinnedIPs          []string      `json:"pinned_ips,omitempty"`
	TLSServerName      string        `json:"tls_server_name,omitempty"`
	RedirectPolicy     string        `json:"redirect_policy,omitempty"`
	LeastPrivilegeNote string        `json:"least_privilege_note,omitempty"`
}

type ProviderEvidence struct {
	Model     string
	SessionID string
}

type AttributionStatus string

const (
	AttributionExact    AttributionStatus = "exact"
	AttributionMissing  AttributionStatus = "missing route evidence"
	AttributionMismatch AttributionStatus = "model evidence mismatch"
)

func NewDispatchRequest(route Route, projectID, projectRoot, runID, sliceID, model, profile, packetHash string, attempt int) (DispatchRequest, error) {
	fingerprint, err := ApprovalRouteFingerprint(route, nil)
	if err != nil {
		return DispatchRequest{}, err
	}
	state, err := ComputeRouteStateFingerprint(route)
	if err != nil {
		return DispatchRequest{}, err
	}
	canonicalProject := projectID
	if canonicalProject == "" {
		canonicalProject, _ = CanonicalProjectIdentity(projectRoot)
	}
	return DispatchRequest{
		RunID: runID, SliceID: sliceID, ProjectID: canonicalProject, ProjectRoot: projectRoot,
		CanonicalProject: canonicalProject, CWD: projectRoot, Worktree: projectRoot,
		RouteAlias: route.Alias, RouteFingerprint: fingerprint, RouteState: state,
		Adapter: route.Adapter, AdapterRevision: route.AdapterRevision, DispatchMethod: route.DispatchMethod,
		Model: model, Profile: profile, ProfileRevision: route.ProfileRevision, PacketHash: packetHash, Attempt: attempt,
		Sandbox: "workspace-write", ApprovalPolicy: "never", Network: "none",
		Authority: "least-privilege", RedirectPolicy: "same-origin-only",
	}, nil
}

func ValidateDispatchRequest(req DispatchRequest, route Route, userRoot, projectRoot string, sources ...map[string]Route) error {
	if strings.TrimSpace(req.RunID) == "" || strings.TrimSpace(req.SliceID) == "" || strings.TrimSpace(req.ProjectID) == "" ||
		strings.TrimSpace(req.RouteAlias) == "" || strings.TrimSpace(req.Model) == "" || strings.TrimSpace(req.PacketHash) == "" ||
		strings.TrimSpace(req.PacketID) == "" || req.Attempt <= 0 || req.Timeout <= 0 || strings.TrimSpace(req.Authority) != "least-privilege" {
		return fmt.Errorf("%w: missing required binding", ErrInvalidDispatchRequest)
	}
	if req.Adapter != route.Adapter || req.AdapterRevision != route.AdapterRevision || req.DispatchMethod != route.DispatchMethod ||
		req.RouteAlias != route.Alias || req.Model != route.DisplayModelID || req.ProfileRevision != route.ProfileRevision {
		return fmt.Errorf("%w: route/request mismatch", ErrInvalidDispatchRequest)
	}
	if route.Adapter != "codex" {
		return fmt.Errorf("%w: direct provider dispatch is not supported; use a Codex profile", ErrUnsupportedDispatchAdapter)
	}
	switch route.DispatchMethod {
	case "exec-model":
		if req.Profile != "" {
			return fmt.Errorf("%w: exec-model cannot use a profile", ErrInvalidDispatchRequest)
		}
	case "exec-profile":
		if route.Profile == "" || req.Profile != route.Profile || route.ProfileRevision == "" {
			return fmt.Errorf("%w: profile must come from trusted route", ErrInvalidDispatchRequest)
		}
		if err := RequireUserLocalProfile(req.Profile, userRoot, projectRoot); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedDispatchAdapter, route.DispatchMethod)
	}
	var sourceMap map[string]Route
	if len(sources) > 0 {
		sourceMap = sources[0]
	}
	expectedFingerprint, err := ApprovalRouteFingerprint(route, sourceMap)
	if err != nil {
		return err
	}
	if req.RouteFingerprint != expectedFingerprint {
		return fmt.Errorf("%w: route fingerprint changed", ErrInvalidDispatchRequest)
	}
	expectedState, err := ComputeRouteStateFingerprint(route)
	if err != nil {
		return err
	}
	if req.RouteState != expectedState {
		return fmt.Errorf("%w: route state changed", ErrInvalidDispatchRequest)
	}
	if len(req.AllowedRoots) == 0 || len(req.AllowedTools) == 0 {
		return fmt.Errorf("%w: least privilege roots/tools are required", ErrInvalidDispatchRequest)
	}
	return nil
}

func RequireUserLocalProfile(profileName, userRoot, projectRoot string) error {
	if !validDispatchProfileName(profileName) {
		return fmt.Errorf("%w: safe profile name is required", ErrProfileNotUserLocal)
	}
	userAbs, err := filepath.Abs(filepath.Clean(userRoot))
	if err != nil {
		return err
	}
	projectAbs, err := filepath.Abs(filepath.Clean(projectRoot))
	if err != nil {
		return err
	}
	profileAbs := filepath.Join(userAbs, profileName+".config.toml")
	if pathWithin(profileAbs, projectAbs) {
		return ErrProjectProviderConfig
	}
	if !pathWithin(profileAbs, userAbs) {
		return ErrProfileNotUserLocal
	}
	info, err := os.Lstat(profileAbs)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return ErrProfileNotUserLocal
	}
	return nil
}

func validDispatchProfileName(value string) bool {
	if len(value) == 0 || len(value) > 64 || strings.Contains(value, "..") {
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

func ClassifyProviderEvidence(requestedModel string, evidence ProviderEvidence) AttributionStatus {
	if strings.TrimSpace(evidence.Model) == "" || strings.TrimSpace(evidence.SessionID) == "" {
		return AttributionMissing
	}
	if evidence.Model != requestedModel {
		return AttributionMismatch
	}
	return AttributionExact
}

func BuildRoutingReceipt(req DispatchRequest, route Route, evidence ProviderEvidence, proof WorkProof, now time.Time) (RoutingReceipt, error) {
	envelope := EvidenceEnvelope{
		RunID: req.RunID, SliceID: req.SliceID, ProjectID: req.ProjectID, RouteAlias: req.RouteAlias,
		RouteFingerprint: req.RouteFingerprint, Adapter: req.Adapter, AdapterRevision: req.AdapterRevision,
		DispatchMethod: req.DispatchMethod, ModelID: req.Model, TaskFamily: route.Capability.TaskFamily,
		ContextPacketID: req.PacketID, ContextPacketHash: req.PacketHash, ProofArtifactHash: proof.ArtifactHash, Surface: "codex-cli",
		Provider: route.Destination, Tools: route.Capability.Tools, ContextSize: route.Capability.ContextSize,
		Risk: route.Capability.Risk, MaxAge: time.Hour,
	}
	capabilityHash, err := ComputeCapabilityEnvelopeHash(envelope)
	if err != nil {
		return RoutingReceipt{}, err
	}
	return RoutingReceipt{
		RouteEvidence: RouteDispatchEvidence{
			RunID: req.RunID, SliceID: req.SliceID, ProjectID: req.ProjectID, RouteAlias: req.RouteAlias,
			RouteFingerprint: req.RouteFingerprint, Adapter: req.Adapter, AdapterRevision: req.AdapterRevision,
			DispatchMethod: req.DispatchMethod, RequestedModelID: req.Model, ProviderReportedModel: evidence.Model,
			SessionID: evidence.SessionID, TaskFamily: route.Capability.TaskFamily, ContextPacketID: req.PacketID, ContextPacketHash: req.PacketHash,
			CapabilityEnvelopeHash: capabilityHash, Attempt: req.Attempt, ObservedAt: now.UTC(),
		},
		WorkProof: proof,
	}, nil
}

func SHA256Bytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func pathWithin(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel))
}
