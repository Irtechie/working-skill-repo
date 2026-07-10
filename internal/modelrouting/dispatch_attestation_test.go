package modelrouting

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDispatchReceiptAttestationBindsExactReceiptHash(t *testing.T) {
	now := time.Date(2026, 7, 10, 16, 0, 0, 0, time.UTC)
	receipt, _ := exactReceipt(now)
	receipt.RouteEvidence.ContextPacketID = "packet-1"
	receipt.WorkProof = WorkProof{Command: "kbrouter dispatch", ArtifactHash: "sha256:proof", Result: ProofUnknown}
	data, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	hash := SHA256Bytes(data)
	root := t.TempDir()
	if err := RecordDispatchReceiptAttestation(root, receipt, hash, now); err != nil {
		t.Fatal(err)
	}
	verified, err := VerifyDispatchReceiptAttestation(root, receipt, hash, now.Add(time.Minute))
	if err != nil || !verified {
		t.Fatalf("verified=%v err=%v", verified, err)
	}
	receipt.RouteEvidence.SessionID = "tampered-session"
	tampered, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	verified, err = VerifyDispatchReceiptAttestation(root, receipt, SHA256Bytes(tampered), now.Add(time.Minute))
	if err != nil || verified {
		t.Fatalf("tampered verified=%v err=%v", verified, err)
	}
}

func TestDispatchReceiptAttestationRefusesCallerProofClaim(t *testing.T) {
	now := time.Date(2026, 7, 10, 16, 0, 0, 0, time.UTC)
	receipt, _ := exactReceipt(now)
	receipt.RouteEvidence.ContextPacketID = "packet-1"
	data, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := RecordDispatchReceiptAttestation(t.TempDir(), receipt, SHA256Bytes(data), now); err == nil {
		t.Fatal("dispatcher attested a caller-authored proof result")
	}
}

func TestLoadStrictJSONBytesReturnsTheDecodedBytes(t *testing.T) {
	root := t.TempDir()
	value := struct {
		SchemaVersion int    `json:"schema_version"`
		Name          string `json:"name"`
	}{SchemaVersion: 1, Name: "exact"}
	if err := SaveAtomicJSON(root, "exact.json", value, 1024); err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(root, "exact.json"))
	if err != nil {
		t.Fatal(err)
	}
	var loaded struct {
		SchemaVersion int    `json:"schema_version"`
		Name          string `json:"name"`
	}
	got, err := LoadStrictJSONBytes(root, "exact.json", &loaded, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) || loaded != value {
		t.Fatalf("exact bytes or decoded value changed: got=%q want=%q loaded=%#v", got, want, loaded)
	}
}
