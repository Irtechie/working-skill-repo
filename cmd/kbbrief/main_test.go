package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutiveBriefGolden(t *testing.T) {
	source, err := os.Open(filepath.Join("testdata", "executive-brief.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	value, err := decodeBrief(source)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderBrief(value)
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile(filepath.Join("testdata", "executive-brief.golden.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rendered, golden) {
		t.Fatalf("generated brief differs from golden\n--- got ---\n%s\n--- want ---\n%s", rendered, golden)
	}
}

func TestResponsibilityContracts(t *testing.T) {
	base := brief{SchemaVersion: 1, Title: "Status", Outcome: "Done"}
	tests := []struct {
		name     string
		response response
		want     string
	}{
		{
			name:     "hard requires complete explanation",
			response: response{Class: "hard_response_required", Ask: "Approve it"},
			want:     "needs ask, why_you, blocked, and recommendation",
		},
		{
			name:     "soft requires default",
			response: response{Class: "soft_preference"},
			want:     "needs default",
		},
		{
			name:     "no response forbids ask",
			response: response{Class: "no_response_needed", Ask: "Pick one"},
			want:     "cannot contain an ask",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := base
			value.Response = test.response
			if err := validateBrief(value); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error=%v want substring %q", err, test.want)
			}
		})
	}
}

func TestVisualGateAndReferences(t *testing.T) {
	small := &visual{
		Mode:  "auto",
		Nodes: []visualNode{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}},
		Edges: []visualEdge{{From: "a", To: "b"}},
	}
	if shouldRenderVisual(small) {
		t.Fatal("auto visual rendered below the cognitive-burden threshold")
	}
	large := &visual{
		Mode:  "auto",
		Nodes: []visualNode{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}, {ID: "c", Label: "C"}},
		Edges: []visualEdge{{From: "a", To: "b"}, {From: "b", To: "c"}},
	}
	if !shouldRenderVisual(large) {
		t.Fatal("meaningful auto visual was suppressed")
	}
	large.Edges[1].To = "missing"
	if err := validateVisual(large); err == nil || !strings.Contains(err.Error(), "unknown node") {
		t.Fatalf("invalid visual edge accepted: %v", err)
	}
}

func TestStrictJSONAndOutput(t *testing.T) {
	input := `{"schema_version":1,"title":"T","outcome":"O","response":{"class":"no_response_needed"},"unknown":true}`
	if _, err := decodeBrief(strings.NewReader(input)); err == nil {
		t.Fatal("unknown JSON field was accepted")
	}
	path := filepath.Join(t.TempDir(), "nested", "brief.md")
	if err := writeOutput(path, []byte("proof\n")); err != nil {
		t.Fatal(err)
	}
	if err := writeOutput(path, []byte("updated\n")); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil || string(content) != "updated\n" {
		t.Fatalf("output=%q err=%v", content, err)
	}
}
