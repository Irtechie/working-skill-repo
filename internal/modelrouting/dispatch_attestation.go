package modelrouting

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const dispatchAttestationMaxAge = time.Hour

type dispatchAttestationKey struct {
	SchemaVersion int    `json:"schema_version"`
	KeyHex        string `json:"key_hex"`
}

type dispatchReceiptAttestation struct {
	SchemaVersion     int       `json:"schema_version"`
	Issuer            string    `json:"issuer"`
	ProjectID         string    `json:"project_id"`
	RunID             string    `json:"run_id"`
	SliceID           string    `json:"slice_id"`
	PacketID          string    `json:"packet_id"`
	ContextPacketHash string    `json:"context_packet_hash"`
	Attempt           int       `json:"attempt"`
	ReceiptHash       string    `json:"receipt_hash"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	MAC               string    `json:"mac"`
}

// RecordDispatchReceiptAttestation is called only by the trusted dispatcher
// after it has launched a worker and read back the exact receipt bytes. It does
// not accept a caller-authored proof result: dispatch receipts remain
// observation-only until ordinary proof runs independently.
func RecordDispatchReceiptAttestation(root string, receipt RoutingReceipt, receiptHash string, now time.Time) error {
	if err := validateDispatchAttestationReceipt(receipt, receiptHash); err != nil {
		return err
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	attestation := dispatchReceiptAttestation{
		SchemaVersion:     1,
		Issuer:            "kbrouter-dispatch-v1",
		ProjectID:         receipt.RouteEvidence.ProjectID,
		RunID:             receipt.RouteEvidence.RunID,
		SliceID:           receipt.RouteEvidence.SliceID,
		PacketID:          receipt.RouteEvidence.ContextPacketID,
		ContextPacketHash: receipt.RouteEvidence.ContextPacketHash,
		Attempt:           receipt.RouteEvidence.Attempt,
		ReceiptHash:       receiptHash,
		CreatedAt:         now,
		ExpiresAt:         now.Add(dispatchAttestationMaxAge),
	}
	dir, name := dispatchReceiptAttestationLocation(root, receipt)
	return WithPrivateStateLock(root, func() error {
		key, err := loadOrCreateDispatchAttestationKey(root)
		if err != nil {
			return err
		}
		attestation.MAC, err = computeDispatchAttestationMAC(attestation, key)
		if err != nil {
			return err
		}
		return SaveAtomicJSON(dir, name, attestation, defaultMaxStorageBytes)
	})
}

// VerifyDispatchReceiptAttestation verifies that kbrouter, rather than the
// telemetry claimant, observed and recorded the exact receipt bytes.
func VerifyDispatchReceiptAttestation(root string, receipt RoutingReceipt, receiptHash string, now time.Time) (bool, error) {
	if err := validateDispatchAttestationReceipt(receipt, receiptHash); err != nil {
		return false, nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	dir, name := dispatchReceiptAttestationLocation(root, receipt)
	verified := false
	err := WithPrivateStateLock(root, func() error {
		key, err := loadDispatchAttestationKey(root)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		var stored dispatchReceiptAttestation
		if err := LoadStrictJSON(dir, name, &stored, defaultMaxStorageBytes); err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		expected := dispatchReceiptAttestation{
			SchemaVersion:     1,
			Issuer:            "kbrouter-dispatch-v1",
			ProjectID:         receipt.RouteEvidence.ProjectID,
			RunID:             receipt.RouteEvidence.RunID,
			SliceID:           receipt.RouteEvidence.SliceID,
			PacketID:          receipt.RouteEvidence.ContextPacketID,
			ContextPacketHash: receipt.RouteEvidence.ContextPacketHash,
			Attempt:           receipt.RouteEvidence.Attempt,
			ReceiptHash:       receiptHash,
		}
		expectedMAC, err := computeDispatchAttestationMAC(stored, key)
		if err != nil {
			return err
		}
		verified = hmac.Equal([]byte(stored.MAC), []byte(expectedMAC)) &&
			stored.SchemaVersion == expected.SchemaVersion && stored.Issuer == expected.Issuer &&
			stored.ProjectID == expected.ProjectID && stored.RunID == expected.RunID &&
			stored.SliceID == expected.SliceID && stored.PacketID == expected.PacketID &&
			stored.ContextPacketHash == expected.ContextPacketHash && stored.Attempt == expected.Attempt &&
			stored.ReceiptHash == expected.ReceiptHash && !stored.CreatedAt.IsZero() && !stored.ExpiresAt.IsZero() &&
			!stored.CreatedAt.After(now) && now.Before(stored.ExpiresAt) &&
			stored.ExpiresAt.Sub(stored.CreatedAt) <= dispatchAttestationMaxAge
		return nil
	})
	return verified, err
}

func validateDispatchAttestationReceipt(receipt RoutingReceipt, receiptHash string) error {
	evidence := receipt.RouteEvidence
	if evidence.ProjectID == "" || evidence.RunID == "" || evidence.SliceID == "" || evidence.ContextPacketID == "" ||
		evidence.ContextPacketHash == "" || evidence.RequestedModelID == "" || evidence.ProviderReportedModel == "" ||
		evidence.RequestedModelID != evidence.ProviderReportedModel || evidence.SessionID == "" || evidence.Attempt <= 0 ||
		receipt.WorkProof.Command != "kbrouter dispatch" || receipt.WorkProof.Result != ProofUnknown || receipt.WorkProof.ArtifactHash == "" ||
		!validSHA256Digest(receiptHash) {
		return fmt.Errorf("dispatch receipt is not eligible for host attestation")
	}
	return nil
}

func dispatchReceiptAttestationLocation(root string, receipt RoutingReceipt) (string, string) {
	project := sha256.Sum256([]byte(receipt.RouteEvidence.ProjectID))
	identity := strings.Join([]string{
		receipt.RouteEvidence.RunID,
		receipt.RouteEvidence.SliceID,
		strconv.Itoa(receipt.RouteEvidence.Attempt),
	}, "\x00")
	file := sha256.Sum256([]byte(identity))
	return filepath.Join(root, "execution-attestations", hex.EncodeToString(project[:16])), hex.EncodeToString(file[:16]) + ".json"
}

func loadOrCreateDispatchAttestationKey(root string) ([]byte, error) {
	key, err := loadDispatchAttestationKey(root)
	if err == nil {
		return key, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	key = make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	value := dispatchAttestationKey{SchemaVersion: 1, KeyHex: hex.EncodeToString(key)}
	if err := SaveAtomicJSON(root, "execution-attestation-key.json", value, defaultMaxStorageBytes); err != nil {
		return nil, err
	}
	return key, nil
}

func loadDispatchAttestationKey(root string) ([]byte, error) {
	var value dispatchAttestationKey
	if err := LoadStrictJSON(root, "execution-attestation-key.json", &value, defaultMaxStorageBytes); err != nil {
		return nil, err
	}
	if value.SchemaVersion != 1 {
		return nil, fmt.Errorf("unsupported dispatch attestation key schema")
	}
	key, err := hex.DecodeString(value.KeyHex)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("invalid dispatch attestation key")
	}
	return key, nil
}

func computeDispatchAttestationMAC(value dispatchReceiptAttestation, key []byte) (string, error) {
	value.MAC = ""
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return "hmac-sha256:" + hex.EncodeToString(mac.Sum(nil)), nil
}

func validSHA256Digest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}
