package reconcile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

const (
	ApplyReceiptSchemaVersion = 1

	StateLocalDurable      = "local-durable"
	StateAwaitingReview    = "awaiting-review"
	StateIntegratedDefault = "delivery-integrated"
	StateUnavailable       = "unavailable"
	StatePreserved         = "preserved"
	StateVerifiedRetired   = "verified-retired"
	StatePartialRepairable = "partial-repairable"
	StateBlocked           = "blocked"
	StateNotEligible       = "not-eligible"
)

type PlanBundle struct {
	Ledger Ledger `json:"ledger"`
	Plan   Plan   `json:"plan"`
}

type ApplyReceipt struct {
	SchemaVersion     int             `json:"schema_version"`
	ID                string          `json:"id"`
	PlanFingerprint   string          `json:"plan_fingerprint"`
	LedgerFingerprint string          `json:"ledger_fingerprint"`
	PolicyVersion     string          `json:"policy_version"`
	Cutoff            time.Time       `json:"cutoff"`
	Actor             string          `json:"actor"`
	SessionID         string          `json:"session_id,omitempty"`
	AppliedAt         time.Time       `json:"applied_at"`
	VerifiedAt        time.Time       `json:"verified_at,omitempty"`
	Status            string          `json:"status"`
	Actions           []ActionReceipt `json:"actions"`
}

type ActionReceipt struct {
	ActionID             string    `json:"action_id"`
	ActionClass          string    `json:"action_class"`
	RepositoryID         string    `json:"repository_id"`
	ArtifactID           string    `json:"artifact_id"`
	TargetPath           string    `json:"target_path,omitempty"`
	TargetRef            string    `json:"target_ref,omitempty"`
	TargetSHA            string    `json:"target_sha,omitempty"`
	Result               string    `json:"result"`
	Issue                string    `json:"issue,omitempty"`
	ObservedBefore       string    `json:"observed_before,omitempty"`
	ObservedAfter        string    `json:"observed_after,omitempty"`
	DeliveryState        string    `json:"delivery_state"`
	PhysicalCleanupState string    `json:"physical_cleanup_state"`
	RefRetirementState   string    `json:"ref_retirement_state"`
	SessionRecordState   string    `json:"session_record_state"`
	AppliedAt            time.Time `json:"applied_at"`
	VerifiedAt           time.Time `json:"verified_at,omitempty"`
}

type Verification struct {
	SchemaVersion int             `json:"schema_version"`
	Status        string          `json:"status"`
	VerifiedAt    time.Time       `json:"verified_at"`
	Actions       []ActionReceipt `json:"actions"`
}

func FingerprintPlan(plan Plan) (string, error) {
	content, err := json.Marshal(plan)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func NewApplyReceipt(bundle PlanBundle, planFingerprint string, now time.Time) ApplyReceipt {
	sum := sha256.Sum256([]byte(planFingerprint))
	return ApplyReceipt{
		SchemaVersion:   ApplyReceiptSchemaVersion,
		ID:              "reconcile-apply:" + hex.EncodeToString(sum[:12]),
		PlanFingerprint: planFingerprint, LedgerFingerprint: bundle.Plan.LedgerFingerprint,
		PolicyVersion: bundle.Plan.PolicyVersion, Cutoff: bundle.Plan.Cutoff,
		Actor: bundle.Plan.Actor, SessionID: bundle.Plan.SessionID,
		AppliedAt: now.UTC(), Status: "pending",
	}
}

func ValidateApplyReceipt(receipt ApplyReceipt, bundle PlanBundle, planFingerprint string) error {
	if receipt.SchemaVersion != ApplyReceiptSchemaVersion {
		return fmt.Errorf("unsupported apply receipt schema %d", receipt.SchemaVersion)
	}
	if receipt.PlanFingerprint != planFingerprint ||
		receipt.LedgerFingerprint != bundle.Plan.LedgerFingerprint ||
		receipt.PolicyVersion != bundle.Plan.PolicyVersion ||
		!receipt.Cutoff.Equal(bundle.Plan.Cutoff) {
		return fmt.Errorf("apply receipt does not match the cutoff-bound plan")
	}
	seen := map[string]bool{}
	for _, action := range receipt.Actions {
		if action.ActionID == "" || seen[action.ActionID] {
			return fmt.Errorf("apply receipt has missing or duplicate action identity")
		}
		seen[action.ActionID] = true
	}
	return nil
}

func normalizeReceipt(receipt *ApplyReceipt) {
	sort.Slice(receipt.Actions, func(i, j int) bool {
		return receipt.Actions[i].ActionID < receipt.Actions[j].ActionID
	})
	switch {
	case len(receipt.Actions) == 0:
		receipt.Status = "verified"
	default:
		receipt.Status = "verified"
		for _, action := range receipt.Actions {
			if action.Result == "blocked" || action.Result == "contended" ||
				action.PhysicalCleanupState == StatePartialRepairable {
				receipt.Status = "partial"
				break
			}
			if action.Result == "preserved" || action.Result == "quarantined" ||
				action.Result == "unavailable" {
				receipt.Status = "preserved"
			}
		}
	}
}
