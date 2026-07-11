package modelrouting

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	CorrectionSchemaVersion      = 1
	MaxCorrectionArtifactBytes   = 1 << 20
	maxCorrectionItems           = 32
	maxCorrectionString          = 512
	AuthorityDriver              = "driver"
	AuthorityIndependentOracle   = "independent-oracle"
	AuthorityWorker              = "worker"
	OracleScopeHunkLocal         = "hunk-local"
	AttemptOutcomeProofFailed    = "proof-failed"
	AttemptOutcomeDispatchFailed = "dispatch-failed"
	ExecutionModeOrdinaryPlanned = "ordinary-planned"
)

var (
	ErrInvalidCorrectionPacket   = errors.New("invalid correction packet")
	ErrCorrectionPilotIneligible = errors.New("correction pilot ineligible")
)

type CorrectionIneligibleReason string

const (
	IneligibleUnlocalizable       CorrectionIneligibleReason = "unlocalizable"
	IneligibleNoIndependentOracle CorrectionIneligibleReason = "no-independent-hunk-oracle"
)

type correctionAssessment struct {
	Eligible bool                       `json:"eligible"`
	Reason   CorrectionIneligibleReason `json:"reason,omitempty"`
}

type CorrectionPacket struct {
	SchemaVersion  int                     `json:"schema_version"`
	PacketID       string                  `json:"packet_id"`
	RunID          string                  `json:"run_id"`
	SliceID        string                  `json:"slice_id"`
	AttemptReceipt AttemptReceiptReference `json:"attempt_receipt"`
	PlannedTier    Tier                    `json:"planned_tier"`
	CorrectionTier Tier                    `json:"correction_tier"`
	Authority      CorrectionAuthority     `json:"authority"`
	Failure        FailureEvidence         `json:"failure"`
	AttemptLedger  []CorrectionAttempt     `json:"attempt_ledger"`
	CurrentDiff    BoundedArtifact         `json:"current_diff"`
	WorkerResult   BoundedArtifact         `json:"worker_result"`
	WorkerLog      BoundedArtifact         `json:"worker_log"`
	AcceptedHunks  []AcceptedHunk          `json:"accepted_hunks"`
}

type AttemptReceiptReference struct {
	Path       string `json:"path"`
	Hash       string `json:"hash"`
	RouteAlias string `json:"route_alias"`
	Attempt    int    `json:"attempt"`
}

type CorrectionAuthority struct {
	Owner                     string         `json:"owner"`
	OriginalScopeHash         string         `json:"original_scope_hash"`
	PreCorrectionBaselineHash string         `json:"pre_correction_baseline_hash"`
	AllowedChanges            []HunkBoundary `json:"allowed_changes"`
	Invariants                []string       `json:"invariants"`
	RelevantInterfaces        []string       `json:"relevant_interfaces"`
	ExactProof                []string       `json:"exact_proof"`
}

type HunkBoundary struct {
	ID        string `json:"id"`
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type FailureEvidence struct {
	CriterionID string        `json:"criterion_id"`
	Localizable bool          `json:"localizable"`
	Location    *HunkBoundary `json:"location,omitempty"`
}

type CorrectionAttempt struct {
	Attempt     int    `json:"attempt"`
	RouteAlias  string `json:"route_alias"`
	ReceiptHash string `json:"receipt_hash"`
	OutcomeCode string `json:"outcome_code"`
}

type BoundedArtifact struct {
	Path     string `json:"path"`
	Hash     string `json:"hash"`
	Bytes    int64  `json:"bytes"`
	Redacted bool   `json:"redacted"`
}

type AcceptedHunk struct {
	Boundary    HunkBoundary `json:"boundary"`
	ContentHash string       `json:"content_hash"`
	Oracle      HunkOracle   `json:"oracle"`
}

type HunkOracle struct {
	Owner        string      `json:"owner"`
	Scope        string      `json:"scope"`
	Command      string      `json:"command"`
	ArtifactHash string      `json:"artifact_hash"`
	Result       ProofResult `json:"result"`
}

type ordinaryPlannedExecutionRecord struct {
	SchemaVersion         int                        `json:"schema_version"`
	RecordID              string                     `json:"record_id"`
	SourcePacketID        string                     `json:"source_packet_id"`
	RunID                 string                     `json:"run_id"`
	SliceID               string                     `json:"slice_id"`
	AuthorityOwner        string                     `json:"authority_owner"`
	Mode                  string                     `json:"mode"`
	Reason                CorrectionIneligibleReason `json:"reason"`
	AttemptReceiptHash    string                     `json:"attempt_receipt_hash"`
	PlannedTier           Tier                       `json:"planned_tier"`
	ExecutionTier         Tier                       `json:"execution_tier"`
	OriginalScopeHash     string                     `json:"original_scope_hash"`
	GeneralizedBroadening bool                       `json:"generalized_broadening"`
}

func assessCorrectionPilotEnvelope(packet CorrectionPacket, attempt RoutingReceipt) (correctionAssessment, error) {
	if err := validateCorrectionBase(packet, attempt); err != nil {
		return correctionAssessment{}, err
	}
	if !packet.Failure.Localizable {
		if packet.Failure.Location != nil {
			return correctionAssessment{}, invalidCorrection("unlocalizable failure has a location")
		}
		return correctionAssessment{Reason: IneligibleUnlocalizable}, nil
	}
	if packet.Failure.Location == nil || !containsBoundary(packet.Authority.AllowedChanges, *packet.Failure.Location) {
		return correctionAssessment{}, invalidCorrection("localized failure is outside the driver-owned change boundary")
	}
	if len(packet.AcceptedHunks) == 0 {
		return correctionAssessment{Reason: IneligibleNoIndependentOracle}, nil
	}
	for _, accepted := range packet.AcceptedHunks {
		if accepted.Oracle.Owner == AuthorityWorker {
			return correctionAssessment{}, invalidCorrection("worker cannot own hunk acceptance")
		}
		if accepted.Oracle.Owner != AuthorityIndependentOracle || accepted.Oracle.Scope != OracleScopeHunkLocal ||
			accepted.Oracle.Result != ProofPass || !validBoundedString(accepted.Oracle.Command) || !validSHA256(accepted.Oracle.ArtifactHash) {
			return correctionAssessment{Reason: IneligibleNoIndependentOracle}, nil
		}
	}
	return correctionAssessment{Eligible: true}, nil
}

func validateCorrectionPacketEnvelope(packet CorrectionPacket, attempt RoutingReceipt) error {
	assessment, err := assessCorrectionPilotEnvelope(packet, attempt)
	if err != nil {
		return err
	}
	if !assessment.Eligible {
		return fmt.Errorf("%w: %s", ErrCorrectionPilotIneligible, assessment.Reason)
	}
	return nil
}

func validateCorrectionBase(packet CorrectionPacket, attempt RoutingReceipt) error {
	if packet.SchemaVersion != CorrectionSchemaVersion || !validBoundedString(packet.PacketID) || !validBoundedString(packet.RunID) || !validBoundedString(packet.SliceID) {
		return invalidCorrection("invalid identity or schema")
	}
	if packet.RunID != attempt.RouteEvidence.RunID || packet.SliceID != attempt.RouteEvidence.SliceID || attempt.RouteEvidence.ProjectID == "" {
		return invalidCorrection("attempt receipt identity mismatch")
	}
	if attempt.WorkProof.Result != ProofFail || !validBoundedString(attempt.WorkProof.Command) || !validSHA256(attempt.WorkProof.ArtifactHash) {
		return invalidCorrection("attempt receipt lacks failed ordinary proof")
	}
	attemptHash, err := HashRoutingReceipt(attempt)
	if err != nil {
		return invalidCorrection(err.Error())
	}
	ref := packet.AttemptReceipt
	if !validDirectChild(ref.Path) || ref.Hash != attemptHash || ref.RouteAlias != attempt.RouteEvidence.RouteAlias || ref.Attempt != attempt.RouteEvidence.Attempt || ref.Attempt <= 0 {
		return invalidCorrection("attempt receipt reference mismatch")
	}
	if !validCorrectionTier(packet.PlannedTier) || !validCorrectionTier(packet.CorrectionTier) || correctionTierRank(packet.CorrectionTier) < correctionTierRank(packet.PlannedTier) {
		return invalidCorrection("correction tier is below planned authority")
	}
	if packet.Authority.Owner != AuthorityDriver || !validSHA256(packet.Authority.OriginalScopeHash) || !validSHA256(packet.Authority.PreCorrectionBaselineHash) {
		return invalidCorrection("authority is not driver owned")
	}
	if err := validateStringList(packet.Authority.Invariants); err != nil {
		return err
	}
	if err := validateStringList(packet.Authority.RelevantInterfaces); err != nil {
		return err
	}
	if err := validateStringList(packet.Authority.ExactProof); err != nil {
		return err
	}
	if len(packet.Authority.AllowedChanges) == 0 || len(packet.Authority.AllowedChanges) > maxCorrectionItems {
		return invalidCorrection("invalid allowed change count")
	}
	allowed := make(map[string]HunkBoundary, len(packet.Authority.AllowedChanges))
	for _, boundary := range packet.Authority.AllowedChanges {
		if !validHunkBoundary(boundary) {
			return invalidCorrection("invalid allowed hunk")
		}
		key := hunkKey(boundary)
		if _, exists := allowed[key]; exists {
			return invalidCorrection("duplicate allowed hunk")
		}
		allowed[key] = boundary
	}
	if !validBoundedString(packet.Failure.CriterionID) {
		return invalidCorrection("missing failed criterion")
	}
	for _, artifact := range []BoundedArtifact{packet.CurrentDiff, packet.WorkerResult, packet.WorkerLog} {
		if !validBoundedArtifact(artifact) {
			return invalidCorrection("worker, diff, and log evidence must be bounded and redacted")
		}
	}
	artifactPaths := []string{packet.AttemptReceipt.Path, packet.CurrentDiff.Path, packet.WorkerResult.Path, packet.WorkerLog.Path}
	seenArtifactPaths := make(map[string]struct{}, len(artifactPaths))
	for _, path := range artifactPaths {
		key := strings.ToLower(path)
		if _, exists := seenArtifactPaths[key]; exists {
			return invalidCorrection("correction artifact paths must be pairwise distinct")
		}
		seenArtifactPaths[key] = struct{}{}
	}
	if len(packet.AttemptLedger) == 0 || len(packet.AttemptLedger) > maxCorrectionItems {
		return invalidCorrection("invalid attempt ledger")
	}
	linked := false
	seenAttempts := make(map[int]struct{}, len(packet.AttemptLedger))
	for _, item := range packet.AttemptLedger {
		if item.Attempt <= 0 || !validBoundedString(item.RouteAlias) || !validSHA256(item.ReceiptHash) ||
			(item.OutcomeCode != AttemptOutcomeProofFailed && item.OutcomeCode != AttemptOutcomeDispatchFailed) {
			return invalidCorrection("invalid attempt ledger entry")
		}
		if _, exists := seenAttempts[item.Attempt]; exists {
			return invalidCorrection("duplicate attempt ledger entry")
		}
		seenAttempts[item.Attempt] = struct{}{}
		if item.Attempt == ref.Attempt && item.RouteAlias == ref.RouteAlias && item.ReceiptHash == ref.Hash {
			linked = true
		}
	}
	if !linked {
		return invalidCorrection("attempt ledger does not contain linked receipt")
	}
	if len(packet.AcceptedHunks) > maxCorrectionItems {
		return invalidCorrection("too many accepted hunks")
	}
	seenAccepted := make(map[string]struct{}, len(packet.AcceptedHunks))
	for _, accepted := range packet.AcceptedHunks {
		if !validHunkBoundary(accepted.Boundary) || !validSHA256(accepted.ContentHash) {
			return invalidCorrection("invalid accepted hunk")
		}
		key := hunkKey(accepted.Boundary)
		if _, exists := seenAccepted[key]; exists {
			return invalidCorrection("duplicate accepted hunk")
		}
		seenAccepted[key] = struct{}{}
		for _, change := range packet.Authority.AllowedChanges {
			if hunksOverlap(change, accepted.Boundary) {
				return invalidCorrection("accepted hunk overlaps allowed correction")
			}
		}
	}
	return nil
}

func validateOrdinaryPlannedExecution(record ordinaryPlannedExecutionRecord, attempt RoutingReceipt) error {
	attemptHash, err := HashRoutingReceipt(attempt)
	if err != nil {
		return err
	}
	if record.SchemaVersion != CorrectionSchemaVersion || !validBoundedString(record.RecordID) || record.RecordID == record.SourcePacketID ||
		record.SourcePacketID != attempt.RouteEvidence.ContextPacketID || record.RunID != attempt.RouteEvidence.RunID || record.SliceID != attempt.RouteEvidence.SliceID ||
		record.AuthorityOwner != AuthorityDriver || record.Mode != ExecutionModeOrdinaryPlanned ||
		(record.Reason != IneligibleUnlocalizable && record.Reason != IneligibleNoIndependentOracle) || record.AttemptReceiptHash != attemptHash ||
		!validCorrectionTier(record.PlannedTier) || !validCorrectionTier(record.ExecutionTier) || correctionTierRank(record.ExecutionTier) < correctionTierRank(record.PlannedTier) ||
		!validSHA256(record.OriginalScopeHash) || record.GeneralizedBroadening {
		return invalidCorrection("invalid ordinary planned execution record")
	}
	return nil
}

func HashRoutingReceipt(receipt RoutingReceipt) (string, error) {
	data, err := json.Marshal(receipt)
	if err != nil {
		return "", err
	}
	return SHA256Bytes(data), nil
}

func HashCorrectionPacket(packet CorrectionPacket) (string, error) {
	data, err := json.Marshal(packet)
	if err != nil {
		return "", err
	}
	return SHA256Bytes(data), nil
}

func verifyBoundedArtifact(root string, artifact BoundedArtifact) error {
	_, err := readBoundedArtifact(root, artifact)
	return err
}

func readBoundedArtifact(root string, artifact BoundedArtifact) ([]byte, error) {
	if !validBoundedArtifact(artifact) {
		return nil, invalidCorrection("invalid bounded artifact reference")
	}
	path, err := safeStoragePath(root, artifact.Path)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Size() != artifact.Bytes || info.Size() > MaxCorrectionArtifactBytes {
		return nil, invalidCorrection("artifact path, type, or size mismatch")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return nil, ErrUnsafePath
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), artifact.Path) {
			continue
		}
		other, infoErr := entry.Info()
		if infoErr != nil {
			return nil, infoErr
		}
		if other.Mode().IsRegular() && os.SameFile(opened, other) {
			return nil, ErrUnsafePath
		}
	}
	data, err := io.ReadAll(io.LimitReader(file, MaxCorrectionArtifactBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) != artifact.Bytes || SHA256Bytes(data) != artifact.Hash {
		return nil, invalidCorrection("artifact hash or size mismatch")
	}
	current, err := os.Lstat(path)
	if err != nil || !os.SameFile(opened, current) {
		return nil, ErrUnsafePath
	}
	return data, nil
}

func CapabilityTier(class CapabilityClass) (Tier, bool) {
	switch class {
	case ClassSmall:
		return TierSmall, true
	case ClassMedium:
		return TierMedium, true
	case ClassLarge, ClassPlanner:
		return TierLarge, true
	default:
		return "", false
	}
}

func CorrectionRouteAllowed(packet CorrectionPacket, class CapabilityClass) bool {
	tier, ok := CapabilityTier(class)
	return ok && correctionTierRank(tier) >= correctionTierRank(packet.CorrectionTier)
}

func invalidCorrection(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidCorrectionPacket, reason)
}

func validCorrectionTier(tier Tier) bool {
	return tier == TierSmall || tier == TierMedium || tier == TierLarge
}

func correctionTierRank(tier Tier) int {
	switch tier {
	case TierSmall:
		return 1
	case TierMedium:
		return 2
	case TierLarge:
		return 3
	default:
		return 0
	}
}

func validBoundedString(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && len(value) <= maxCorrectionString && !strings.ContainsRune(value, '\x00')
}

func validateStringList(values []string) error {
	if len(values) == 0 || len(values) > maxCorrectionItems {
		return invalidCorrection("invalid driver authority list")
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validBoundedString(value) {
			return invalidCorrection("invalid driver authority value")
		}
		if _, exists := seen[value]; exists {
			return invalidCorrection("duplicate driver authority value")
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validSHA256(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func validBoundedArtifact(artifact BoundedArtifact) bool {
	return artifact.Redacted && validDirectChild(artifact.Path) && artifact.Bytes >= 0 && artifact.Bytes <= MaxCorrectionArtifactBytes && validSHA256(artifact.Hash)
}

func validDirectChild(value string) bool {
	return validBoundedString(value) && !filepath.IsAbs(value) && filepath.Base(value) == value && !strings.ContainsAny(value, `/\`) && validWindowsPathComponent(value)
}

func validPortableRelativePath(value string) bool {
	if value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, `\`) {
		return false
	}
	for _, part := range strings.Split(value, "/") {
		if !validWindowsPathComponent(part) {
			return false
		}
	}
	return true
}

func validWindowsPathComponent(value string) bool {
	if value == "" || value == "." || value == ".." || strings.Contains(value, ":") || strings.HasSuffix(value, ".") || strings.HasSuffix(value, " ") {
		return false
	}
	base := strings.ToUpper(strings.SplitN(value, ".", 2)[0])
	if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || base == "CLOCK$" {
		return false
	}
	if len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9' {
		return false
	}
	for _, r := range value {
		if r < 32 || strings.ContainsRune(`<>"|?*`, r) {
			return false
		}
	}
	return true
}

func validHunkBoundary(boundary HunkBoundary) bool {
	if !validBoundedString(boundary.ID) || !validBoundedString(boundary.File) || boundary.StartLine <= 0 || boundary.EndLine < boundary.StartLine {
		return false
	}
	clean := filepath.Clean(boundary.File)
	return !filepath.IsAbs(boundary.File) && clean != "." && clean != ".." && !strings.HasPrefix(clean, ".."+string(filepath.Separator)) && validPortableRelativePath(filepath.ToSlash(boundary.File))
}

func hunkKey(boundary HunkBoundary) string {
	return fmt.Sprintf("%s\x00%s\x00%d\x00%d", boundary.ID, boundary.File, boundary.StartLine, boundary.EndLine)
}

func containsBoundary(values []HunkBoundary, target HunkBoundary) bool {
	for _, value := range values {
		if sameBoundary(value, target) {
			return true
		}
	}
	return false
}

func sameBoundary(left, right HunkBoundary) bool {
	return left.ID == right.ID && left.File == right.File && left.StartLine == right.StartLine && left.EndLine == right.EndLine
}

func hunksOverlap(left, right HunkBoundary) bool {
	return left.File == right.File && left.StartLine <= right.EndLine && right.StartLine <= left.EndLine
}
