package modelrouting

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

const defaultAttemptLimit = 8

type AttemptLedger struct {
	max      int
	attempts []string
	seen     map[string]struct{}
}

func NewAttemptLedger(max int) AttemptLedger {
	if max <= 0 {
		max = defaultAttemptLimit
	}
	return AttemptLedger{max: max, seen: make(map[string]struct{})}
}

func (l *AttemptLedger) Record(alias string) error {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return ErrInvalidAttempt
	}
	if l.max <= 0 {
		l.max = defaultAttemptLimit
	}
	if l.seen == nil {
		l.seen = make(map[string]struct{})
	}
	if _, ok := l.seen[alias]; ok {
		return ErrRouteAlreadyAttempted
	}
	if len(l.attempts) >= l.max {
		return ErrAttemptLedgerFull
	}
	l.seen[alias] = struct{}{}
	l.attempts = append(l.attempts, alias)
	return nil
}

func (l AttemptLedger) Attempted(alias string) bool {
	if _, ok := l.seen[alias]; ok {
		return true
	}
	for _, attempt := range l.attempts {
		if attempt == alias {
			return true
		}
	}
	return false
}

type RoutingReceipt struct {
	RouteEvidence RouteDispatchEvidence `json:"route_evidence"`
	WorkProof     WorkProof             `json:"work_proof"`
}

type RouteDispatchEvidence struct {
	RunID                  string    `json:"run_id"`
	SliceID                string    `json:"slice_id"`
	ProjectID              string    `json:"project_id"`
	RouteAlias             string    `json:"route_alias"`
	RouteFingerprint       string    `json:"route_fingerprint"`
	Adapter                string    `json:"adapter"`
	AdapterRevision        string    `json:"adapter_revision"`
	DispatchMethod         string    `json:"dispatch_method"`
	RequestedModelID       string    `json:"requested_model_id"`
	ProviderReportedModel  string    `json:"provider_reported_model"`
	SessionID              string    `json:"session_id"`
	TaskFamily             string    `json:"task_family"`
	ContextPacketID        string    `json:"context_packet_id,omitempty"`
	ContextPacketHash      string    `json:"context_packet_hash"`
	CapabilityEnvelopeHash string    `json:"capability_envelope_hash"`
	Attempt                int       `json:"attempt"`
	ObservedAt             time.Time `json:"observed_at"`
}

type WorkProof struct {
	Command      string      `json:"command"`
	ArtifactHash string      `json:"artifact_hash"`
	Result       ProofResult `json:"result"`
}

type ProofResult string

const (
	ProofPass    ProofResult = "pass"
	ProofFail    ProofResult = "fail"
	ProofUnknown ProofResult = "unknown"
)

type EvidenceEnvelope struct {
	RunID             string
	SliceID           string
	ProjectID         string
	RouteAlias        string
	RouteFingerprint  string
	Adapter           string
	AdapterRevision   string
	DispatchMethod    string
	ModelID           string
	TaskFamily        string
	ContextPacketID   string
	ContextPacketHash string
	ProofArtifactHash string
	Surface           string
	Provider          string
	Tools             []string
	ContextSize       int
	Risk              RiskLevel
	MaxAge            time.Duration
}

type ObservationStatus string

const (
	ObservationCredited ObservationStatus = "credited"
	ObservationOnly     ObservationStatus = "observation-only"
)

type RoutingObservation struct {
	Status        ObservationStatus
	RouteEvidence RouteDispatchEvidence
	WorkProof     WorkProof
	ObservedAt    time.Time
}

func EvaluateReceiptCredit(receipt RoutingReceipt, envelope EvidenceEnvelope, now time.Time) (bool, RoutingObservation) {
	observation := RoutingObservation{Status: ObservationOnly, RouteEvidence: receipt.RouteEvidence, WorkProof: receipt.WorkProof, ObservedAt: now}
	proof := receipt.WorkProof
	if proof.Result != ProofPass || strings.TrimSpace(proof.Command) == "" || proof.ArtifactHash == "" || proof.ArtifactHash != envelope.ProofArtifactHash {
		return false, observation
	}
	evidence := receipt.RouteEvidence
	if evidence.RunID == "" || evidence.SliceID == "" || evidence.ProjectID == "" || evidence.RouteAlias == "" || evidence.RouteFingerprint == "" ||
		evidence.Adapter == "" || evidence.AdapterRevision == "" || evidence.DispatchMethod == "" || evidence.RequestedModelID == "" ||
		evidence.ProviderReportedModel == "" || evidence.SessionID == "" || evidence.TaskFamily == "" || evidence.ContextPacketHash == "" ||
		evidence.CapabilityEnvelopeHash == "" || evidence.Attempt <= 0 || evidence.ObservedAt.IsZero() {
		return false, observation
	}
	envelopeHash, err := ComputeCapabilityEnvelopeHash(envelope)
	if err != nil || evidence.CapabilityEnvelopeHash != envelopeHash {
		return false, observation
	}
	maxAge := envelope.MaxAge
	if maxAge <= 0 {
		maxAge = time.Hour
	}
	if evidence.ObservedAt.After(now) || now.Sub(evidence.ObservedAt) > maxAge {
		return false, observation
	}
	if evidence.RunID != envelope.RunID || evidence.SliceID != envelope.SliceID || evidence.ProjectID != envelope.ProjectID ||
		evidence.RouteAlias != envelope.RouteAlias || evidence.RouteFingerprint != envelope.RouteFingerprint ||
		evidence.Adapter != envelope.Adapter || evidence.AdapterRevision != envelope.AdapterRevision ||
		evidence.DispatchMethod != envelope.DispatchMethod || evidence.RequestedModelID != envelope.ModelID ||
		evidence.ProviderReportedModel != envelope.ModelID || evidence.TaskFamily != envelope.TaskFamily ||
		evidence.ContextPacketHash != envelope.ContextPacketHash || (envelope.ContextPacketID != "" && evidence.ContextPacketID != envelope.ContextPacketID) {
		return false, observation
	}
	observation.Status = ObservationCredited
	return true, observation
}

func ComputeCapabilityEnvelopeHash(envelope EvidenceEnvelope) (string, error) {
	tools := append([]string(nil), envelope.Tools...)
	sort.Strings(tools)
	payload := struct {
		TaskFamily        string    `json:"task_family"`
		Tools             []string  `json:"tools"`
		ContextSize       int       `json:"context_size"`
		Risk              RiskLevel `json:"risk"`
		Surface           string    `json:"surface"`
		Provider          string    `json:"provider"`
		ModelID           string    `json:"model_id"`
		Adapter           string    `json:"adapter"`
		AdapterRevision   string    `json:"adapter_revision"`
		ContextPacketHash string    `json:"context_packet_hash"`
	}{envelope.TaskFamily, tools, envelope.ContextSize, envelope.Risk, envelope.Surface, envelope.Provider,
		envelope.ModelID, envelope.Adapter, envelope.AdapterRevision, envelope.ContextPacketHash}
	if payload.TaskFamily == "" || len(payload.Tools) == 0 || payload.ContextSize <= 0 || !validRisk(payload.Risk) ||
		payload.Surface == "" || payload.Provider == "" || payload.ModelID == "" || payload.Adapter == "" ||
		payload.AdapterRevision == "" || payload.ContextPacketHash == "" {
		return "", ErrInvalidCatalog
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "capability-sha256:" + hex.EncodeToString(sum[:]), nil
}
