package main

import (
	"strings"
	"testing"
)

func TestContextPacketValidation(t *testing.T) {
	packet := contextPacket{
		SchemaVersion:      1,
		PacketID:           "p1",
		TaskID:             "t1",
		Objective:          "make a bounded change",
		SourceFiles:        []string{"cmd/kbcheck/main.go"},
		Constraints:        []string{"no provider dependency"},
		OutOfScope:         []string{"runtime scheduler"},
		AcceptanceCriteria: []string{"validation passes"},
		ProofTargets:       []string{"go test ./..."},
		ModelTier:          "small",
		ModelTierReason:    "mechanical change",
		AllowedTools:       []string{"view", "go test"},
		BroadSearchPolicy:  "bounded",
		EscalationTriggers: []string{"scope expands"},
	}

	if result := validateContextPacket(packet); !result.OK {
		t.Fatalf("valid packet failed: %v", result.Issues)
	}

	packet.ModelTier = "cheap"
	if result := validateContextPacket(packet); result.OK {
		t.Fatal("invalid model tier passed")
	}

}

func TestContextPacketCommandUsesFixtures(t *testing.T) {
	var out, errOut strings.Builder
	code := run([]string{"context-packet", "--root", ".", "--packet", "testdata/context-packet-valid.json"}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), "context packet: ok") {
		t.Fatalf("valid fixture failed code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()
	code = run([]string{"context-packet", "--root", ".", "--packet", "testdata/context-packet-invalid.json"}, &out, &errOut)
	if code != 2 || !strings.Contains(errOut.String(), "source_files") {
		t.Fatalf("invalid fixture result code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestContextPacketJSONReadErrorIsStructured(t *testing.T) {
	var out, errOut strings.Builder
	code := run([]string{"context-packet", "--root", ".", "--packet", "missing.json", "--json"}, &out, &errOut)
	if code != 1 || !strings.Contains(out.String(), `"ok": false`) || errOut.Len() != 0 {
		t.Fatalf("read error code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestContextPacketRequiresAuthorityBounds(t *testing.T) {
	packet := contextPacket{
		SchemaVersion:      1,
		PacketID:           "p1",
		TaskID:             "t1",
		Objective:          "measure usage",
		SourceFiles:        []string{"result.json"},
		Constraints:        []string{"raw values only"},
		AcceptanceCriteria: []string{"negative usage fails"},
		ProofTargets:       []string{"go test ./..."},
		ModelTier:          "tiny",
		ModelTierReason:    "schema validation",
		BroadSearchPolicy:  "forbidden",
		EscalationTriggers: []string{"unknown field"},
	}
	if result := validateContextPacket(packet); result.OK {
		t.Fatal("packet without out_of_scope and allowed_tools passed")
	}
}

func TestContextPacketRejectsBroadGlobWhenSearchBounded(t *testing.T) {
	packet := contextPacket{
		SchemaVersion: 1, PacketID: "p1", TaskID: "t1", Objective: "bounded",
		SourceFiles: []string{"**/*"}, Constraints: []string{"bounded"},
		OutOfScope: []string{"unrelated"}, AcceptanceCriteria: []string{"reject glob"},
		ProofTargets: []string{"go test ./..."}, ModelTier: "small",
		ModelTierReason: "bounded", AllowedTools: []string{"rg"},
		BroadSearchPolicy: "forbidden", EscalationTriggers: []string{"missing file"},
	}

	if result := validateContextPacket(packet); result.OK {
		t.Fatal("glob passed with forbidden broad search")
	}
}

func TestContextPacketRejectsRepoRootAndBlankMemoryPath(t *testing.T) {
	packet := contextPacket{
		SchemaVersion: 1, PacketID: "p1", TaskID: "t1", Objective: "bounded",
		MemoryFiles: []string{""}, SourceFiles: []string{"."},
		Constraints: []string{"bounded"}, OutOfScope: []string{"unrelated"},
		AcceptanceCriteria: []string{"reject root"}, ProofTargets: []string{"go test ./..."},
		ModelTier: "small", ModelTierReason: "bounded", AllowedTools: []string{"view"},
		BroadSearchPolicy: "forbidden", EscalationTriggers: []string{"missing file"},
	}
	if result := validateContextPacket(packet); result.OK {
		t.Fatal("repo root or blank memory path passed")
	}
}
