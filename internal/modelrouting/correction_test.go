package modelrouting

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCorrectionPathsRejectWindowsADSDevicesTrailingNamesAndHardlinks(t *testing.T) {
	for _, name := range []string{"file.txt:secret", "CON", "aux.txt", "COM1.log", "name.", "name "} {
		if validDirectChild(name) {
			t.Errorf("unsafe Windows name accepted: %q", name)
		}
	}
}

func TestCorrectionPacketAcceptsOnlyDriverBoundLocalPilot(t *testing.T) {
	packet, attempt := validCorrectionFixture(t)
	assessment, err := assessCorrectionPilotEnvelope(packet, attempt)
	if err != nil || !assessment.Eligible || assessment.Reason != "" {
		t.Fatalf("assessment=%#v err=%v", assessment, err)
	}
	if err := validateCorrectionPacketEnvelope(packet, attempt); err != nil {
		t.Fatalf("validate: %v", err)
	}

	tests := map[string]func(*CorrectionPacket){
		"worker authority":   func(p *CorrectionPacket) { p.Authority.Owner = AuthorityWorker },
		"below planned tier": func(p *CorrectionPacket) { p.CorrectionTier = TierSmall },
		"unredacted diff":    func(p *CorrectionPacket) { p.CurrentDiff.Redacted = false },
		"oversized log":      func(p *CorrectionPacket) { p.WorkerLog.Bytes = MaxCorrectionArtifactBytes + 1 },
		"worker oracle":      func(p *CorrectionPacket) { p.AcceptedHunks[0].Oracle.Owner = AuthorityWorker },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			candidate := packet
			candidate.AcceptedHunks = append([]AcceptedHunk(nil), packet.AcceptedHunks...)
			mutate(&candidate)
			if err := validateCorrectionPacketEnvelope(candidate, attempt); !errors.Is(err, ErrInvalidCorrectionPacket) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestCorrectionPilotIneligibleRoutesToSeparateOrdinaryExecution(t *testing.T) {
	packet, attempt := validCorrectionFixture(t)

	unlocalizable := packet
	unlocalizable.Failure.Localizable = false
	unlocalizable.Failure.Location = nil
	assessment, err := assessCorrectionPilotEnvelope(unlocalizable, attempt)
	if err != nil || assessment.Eligible || assessment.Reason != IneligibleUnlocalizable {
		t.Fatalf("unlocalizable assessment=%#v err=%v", assessment, err)
	}
	if err := validateCorrectionPacketEnvelope(unlocalizable, attempt); !errors.Is(err, ErrCorrectionPilotIneligible) {
		t.Fatalf("unlocalizable validate=%v", err)
	}

	withoutOracle := packet
	withoutOracle.AcceptedHunks = nil
	assessment, err = assessCorrectionPilotEnvelope(withoutOracle, attempt)
	if err != nil || assessment.Eligible || assessment.Reason != IneligibleNoIndependentOracle {
		t.Fatalf("no-oracle assessment=%#v err=%v", assessment, err)
	}

	attemptHash, err := HashRoutingReceipt(attempt)
	if err != nil {
		t.Fatal(err)
	}
	record := ordinaryPlannedExecutionRecord{
		SchemaVersion: CorrectionSchemaVersion, RecordID: "ordinary-1", SourcePacketID: attempt.RouteEvidence.ContextPacketID,
		RunID: packet.RunID, SliceID: packet.SliceID, AuthorityOwner: AuthorityDriver, Mode: ExecutionModeOrdinaryPlanned,
		Reason: assessment.Reason, AttemptReceiptHash: attemptHash, PlannedTier: packet.PlannedTier, ExecutionTier: packet.PlannedTier,
		OriginalScopeHash: packet.Authority.OriginalScopeHash, GeneralizedBroadening: false,
	}
	if err := validateOrdinaryPlannedExecution(record, attempt); err != nil {
		t.Fatalf("ordinary validation: %v", err)
	}
	record.GeneralizedBroadening = true
	if err := validateOrdinaryPlannedExecution(record, attempt); !errors.Is(err, ErrInvalidCorrectionPacket) {
		t.Fatalf("broadening accepted: %v", err)
	}
}

func TestVerifyBoundedArtifactBindsDirectChildPathHashAndSize(t *testing.T) {
	root := t.TempDir()
	content := []byte("redacted evidence")
	writeCorrectionArtifact(t, root, "evidence.txt", content)
	artifact := BoundedArtifact{Path: "evidence.txt", Hash: SHA256Bytes(content), Bytes: int64(len(content)), Redacted: true}
	if err := verifyBoundedArtifact(root, artifact); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if err := os.WriteFile(root+string(os.PathSeparator)+"evidence.txt", []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyBoundedArtifact(root, artifact); !errors.Is(err, ErrInvalidCorrectionPacket) {
		t.Fatalf("tampered artifact accepted: %v", err)
	}
	hardlinkRoot := t.TempDir()
	writeCorrectionArtifact(t, hardlinkRoot, "evidence.txt", content)
	if err := os.Link(filepath.Join(hardlinkRoot, "evidence.txt"), filepath.Join(hardlinkRoot, "alias.txt")); err == nil {
		if err := verifyBoundedArtifact(hardlinkRoot, artifact); !errors.Is(err, ErrUnsafePath) {
			t.Fatalf("hardlinked artifact accepted: %v", err)
		}
	}
}

func validCorrectionFixture(t *testing.T) (CorrectionPacket, RoutingReceipt) {
	t.Helper()
	attempt := RoutingReceipt{
		RouteEvidence: RouteDispatchEvidence{
			RunID: "run-1", SliceID: "slice-1", ProjectID: "project-1", RouteAlias: "small", RouteFingerprint: testHash("route"),
			Adapter: "codex", AdapterRevision: "v1", DispatchMethod: "named-agent", RequestedModelID: "small-model",
			ProviderReportedModel: "small-model", SessionID: "session-1", TaskFamily: "code", ContextPacketID: "attempt-packet",
			ContextPacketHash: testHash("attempt-packet"), CapabilityEnvelopeHash: "capability-" + testHash("capability"), Attempt: 1,
		},
		WorkProof: WorkProof{Command: "go test ./...", ArtifactHash: testHash("attempt-proof"), Result: ProofFail},
	}
	attemptHash, err := HashRoutingReceipt(attempt)
	if err != nil {
		t.Fatal(err)
	}
	fix := HunkBoundary{ID: "fix", File: "internal/example.go", StartLine: 20, EndLine: 24}
	keep := HunkBoundary{ID: "keep", File: "internal/example.go", StartLine: 5, EndLine: 10}
	packet := CorrectionPacket{
		SchemaVersion: CorrectionSchemaVersion, PacketID: "correction-1", RunID: "run-1", SliceID: "slice-1",
		AttemptReceipt: AttemptReceiptReference{Path: "attempt-receipt.json", Hash: attemptHash, RouteAlias: "small", Attempt: 1},
		PlannedTier:    TierMedium, CorrectionTier: TierMedium,
		Authority: CorrectionAuthority{
			Owner: AuthorityDriver, OriginalScopeHash: testHash("original-scope"), PreCorrectionBaselineHash: testHash("pre-correction-baseline"), AllowedChanges: []HunkBoundary{fix},
			Invariants: []string{"public API unchanged"}, RelevantInterfaces: []string{"internal/example.go"}, ExactProof: []string{"go test ./..."},
		},
		Failure:       FailureEvidence{CriterionID: "proof-1", Localizable: true, Location: &fix},
		AttemptLedger: []CorrectionAttempt{{Attempt: 1, RouteAlias: "small", ReceiptHash: attemptHash, OutcomeCode: AttemptOutcomeProofFailed}},
		CurrentDiff:   BoundedArtifact{Path: "current.diff", Hash: testHash("diff"), Bytes: int64(len("diff")), Redacted: true},
		WorkerResult:  BoundedArtifact{Path: "worker-result.json", Hash: testHash("worker-result"), Bytes: int64(len("worker-result")), Redacted: true},
		WorkerLog:     BoundedArtifact{Path: "worker.log", Hash: testHash("worker-log"), Bytes: int64(len("worker-log")), Redacted: true},
		AcceptedHunks: []AcceptedHunk{{
			Boundary: keep, ContentHash: testHash("accepted-content"),
			Oracle: HunkOracle{Owner: AuthorityIndependentOracle, Scope: OracleScopeHunkLocal, Command: "go test ./... -run Keep", ArtifactHash: testHash("hunk-proof"), Result: ProofPass},
		}},
	}
	return packet, attempt
}

func testHash(value string) string { return SHA256Bytes([]byte(value)) }

func writeCorrectionArtifact(t *testing.T, root, name string, content []byte) {
	t.Helper()
	if err := os.WriteFile(root+string(os.PathSeparator)+name, content, 0o600); err != nil {
		t.Fatal(err)
	}
}
