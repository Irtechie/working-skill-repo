package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type ablationFixture struct {
	t          *testing.T
	root       string
	id         ablationIdentity
	inv        ablationInventory
	transcript ablationTranscript
	proof      ablationProof
	contract   ablationContract
}

func (f *ablationFixture) bytes(name string, raw []byte) ablationRef {
	f.t.Helper()
	p := filepath.Join(f.root, name)
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		f.t.Fatal(e)
	}
	if e := os.WriteFile(p, raw, 0644); e != nil {
		f.t.Fatal(e)
	}
	return ablationRef{name, hashString(string(raw))}
}
func (f *ablationFixture) json(name string, v any) ablationRef {
	f.t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		f.t.Fatal(e)
	}
	return f.bytes(name, b)
}
func newAblationFixture(t *testing.T, root, condition string) *ablationFixture {
	f := &ablationFixture{t: t, root: root}
	f.id = ablationIdentity{RunID: "run-" + condition, Case: "case-1", Repetition: 1, Host: "codex", HostVersion: "1.2", Model: "model", Config: hashString("config"), Project: hashString("project"), Prompt: hashString("prompt"), Condition: condition}
	yes, no := true, false
	f.inv = ablationInventory{Schema: 1, Source: "external-capture", Coverage: map[string]bool{"global": true, "project": true, "skills": true, "native": true}, Instructions: []ablationInstruction{{Name: "host-native", Scope: "native", KB: &no, Content: f.bytes("artifacts/native.txt", []byte("native instructions"))}}}
	if condition != "none" {
		f.inv.Skills = []ablationInstruction{{Name: "kb-work", Scope: "skills", KB: &yes, Content: f.bytes("artifacts/"+condition+"-skill.txt", []byte(condition+" instructions"))}}
	}
	f.transcript = ablationTranscript{Schema: 1, Source: "external-capture", Content: "Synthetic test fixture standing in for externally captured transcript."}
	zero := 0
	cost := 2.0
	if condition == "reduced" {
		cost = 1
	}
	f.contract = ablationContract{Schema: 1, Case: "case-1", Command: []string{"never-execute-this-imported-command", "--check"}, Oracle: f.bytes("artifacts/oracle.txt", []byte("required behavior"))}
	f.proof = ablationProof{Schema: 1, Source: "external-capture", Command: f.contract.Command, Exit: &zero, Started: "2026-09-06T00:00:00Z", Ended: "2026-09-06T00:00:01Z", Outputs: []ablationRef{f.bytes("artifacts/output.txt", []byte("captured output"))}, Metrics: ablationMetrics{Cost: ablationMetric{Value: &cost, Source: "external-capture"}}}
	return f
}
func (f *ablationFixture) save() ablationRecord {
	f.t.Helper()
	prefix := "artifacts/" + f.id.RunID
	f.inv.Identity = f.id
	inventory := f.json(prefix+"-inventory.json", f.inv)
	f.transcript.Identity = f.id
	f.transcript.Profile = inventory.SHA256
	transcript := f.json(prefix+"-transcript.json", f.transcript)
	f.proof.Identity = f.id
	f.proof.Profile = inventory.SHA256
	f.proof.Transcript = transcript.SHA256
	f.proof.Contract = f.json(prefix+"-contract.json", f.contract)
	row := ablationRecord{Schema: 1, Kind: "live", Identity: f.id, Inventory: inventory, Transcript: transcript, Proof: f.json(prefix+"-proof.json", f.proof)}
	f.json("records/"+f.id.RunID+".json", row)
	return row
}
func ablationRun(t *testing.T, root string) ablationReport {
	t.Helper()
	var out, err strings.Builder
	if code := runSkillAblationCommand(root, options{resultRoot: "records", output: "report.json", json: true}, &out, &err); code != 0 {
		t.Fatalf("code=%d %s", code, err.String())
	}
	var report ablationReport
	if e := json.Unmarshal([]byte(out.String()), &report); e != nil {
		t.Fatal(e)
	}
	return report
}
func TestSkillAblationCapturedFailureAndUnknownMetrics(t *testing.T) {
	root := t.TempDir()
	for _, c := range []string{"full", "reduced", "none"} {
		f := newAblationFixture(t, root, c)
		if c == "reduced" {
			exit := 1
			f.proof.Exit = &exit
		}
		f.save()
	}
	r := ablationRun(t, root)
	if r.Eligible != 3 || r.Excluded != 0 || len(r.Matched) != 1 {
		t.Fatalf("%+v", r)
	}
	g := r.Matched[0]
	if g.Arms["reduced"].Success || g.Arms["reduced"].Exit != 1 || len(g.Regressions) != 1 || g.Regressions[0] != "reduced" {
		t.Fatalf("failure lost: %+v", g)
	}
	if *g.Differences["reduced_minus_full"]["cost_usd"] != -1 || g.Differences["reduced_minus_full"]["tokens"] != nil || g.Arms["none"].Metrics["elapsed_ms"] != nil {
		t.Fatalf("metrics fabricated: %+v", g)
	}
	if !strings.Contains(r.Limitation, "not capture authorship") {
		t.Fatal("missing trust limitation")
	}
}
func TestSkillAblationRejectsUnprovenCaptures(t *testing.T) {
	cases := map[string]func(*ablationFixture){
		"self-reported":        func(f *ablationFixture) { f.proof.Source = "model" },
		"missing-exit":         func(f *ablationFixture) { f.proof.Exit = nil },
		"incomplete-discovery": func(f *ablationFixture) { f.inv.Coverage["global"] = false },
		"none-contamination": func(f *ablationFixture) {
			yes := true
			f.inv.Skills = []ablationInstruction{{Name: "kb-work", Scope: "skills", KB: &yes, Content: f.bytes("artifacts/contaminant.txt", []byte("KB"))}}
		},
		"changed-command":  func(f *ablationFixture) { f.proof.Command = []string{"different"} },
		"backwards-time":   func(f *ablationFixture) { f.proof.Ended = "2020-01-01T00:00:00Z" },
		"escape":           func(f *ablationFixture) { f.contract.Oracle.Path = "../outside" },
		"missing-artifact": func(f *ablationFixture) { f.contract.Oracle.Path = "missing" },
		"hash-mismatch":    func(f *ablationFixture) { f.contract.Oracle.SHA256 = hashString("wrong") },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			f := newAblationFixture(t, root, "none")
			mutate(f)
			f.save()
			r := ablationRun(t, root)
			if r.Excluded != 1 || r.Eligible != 0 {
				t.Fatalf("%+v", r)
			}
		})
	}
}
func TestSkillAblationInlineProofNeverQualifies(t *testing.T) {
	root := t.TempDir()
	f := newAblationFixture(t, root, "full")
	f.json("records/inline.json", map[string]any{"evidence_kind": "live", "condition": "full", "task_success": "pass", "independent_proof": map[string]any{"command": "go test ./...", "exit_code": 0}})
	r := ablationRun(t, root)
	if r.Excluded != 1 || r.Eligible != 0 {
		t.Fatalf("%+v", r)
	}
}
func TestSkillAblationDuplicateArmInvalidatesBoth(t *testing.T) {
	root := t.TempDir()
	for _, c := range []string{"full", "reduced", "none"} {
		newAblationFixture(t, root, c).save()
	}
	f := newAblationFixture(t, root, "full")
	f.id.RunID = "another-full"
	f.save()
	r := ablationRun(t, root)
	if r.Excluded != 2 || r.Eligible != 2 || len(r.Matched) != 0 {
		t.Fatalf("%+v", r)
	}
}
func TestSkillAblationIdentityAndOracleCannotCrossMatch(t *testing.T) {
	cases := map[string]func(*ablationFixture){"host": func(f *ablationFixture) { f.id.Host = "other" }, "version": func(f *ablationFixture) { f.id.HostVersion = "2" }, "config": func(f *ablationFixture) { f.id.Config = hashString("other") }, "project": func(f *ablationFixture) { f.id.Project = hashString("other") }, "prompt": func(f *ablationFixture) { f.id.Prompt = hashString("other") }, "oracle": func(f *ablationFixture) { f.contract.Oracle = f.bytes("artifacts/other-oracle", []byte("other")) }, "native": func(f *ablationFixture) {
		f.inv.Instructions[0].Content = f.bytes("artifacts/other-native", []byte("other"))
	}}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			for _, c := range []string{"full", "reduced", "none"} {
				f := newAblationFixture(t, root, c)
				if c == "none" {
					mutate(f)
				}
				f.save()
			}
			r := ablationRun(t, root)
			if r.Eligible != 3 || len(r.Matched) != 0 || len(r.Incomplete) == 0 {
				t.Fatalf("%+v", r)
			}
		})
	}
}
func TestSkillAblationStrictCaptureIdentity(t *testing.T) {
	for _, mode := range []string{"identity", "profile", "duplicate-key", "exit-string", "large"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			f := newAblationFixture(t, root, "full")
			row := f.save()
			raw, e := os.ReadFile(filepath.Join(root, row.Proof.Path))
			if e != nil {
				t.Fatal(e)
			}
			switch mode {
			case "identity":
				raw = []byte(strings.Replace(string(raw), `"host":"codex"`, `"host":"other"`, 1))
			case "profile":
				raw = []byte(strings.Replace(string(raw), row.Inventory.SHA256, hashString("other"), 1))
			case "duplicate-key":
				raw = []byte(strings.Replace(string(raw), `"exit_code":0`, `"exit_code":1,"exit_code":0`, 1))
			case "exit-string":
				raw = []byte(strings.Replace(string(raw), `"exit_code":0`, `"exit_code":"0"`, 1))
			case "large":
				raw = []byte(strings.Repeat("x", (1<<20)+1))
			}
			row.Proof = f.bytes(row.Proof.Path, raw)
			f.json("records/"+f.id.RunID+".json", row)
			r := ablationRun(t, root)
			if r.Excluded != 1 {
				t.Fatalf("%+v", r)
			}
		})
	}
}
func TestSkillAblationOutputCannotOverwriteEvidence(t *testing.T) {
	root := t.TempDir()
	f := newAblationFixture(t, root, "full")
	row := f.save()
	for _, output := range []string{row.Proof.Path, "records/run-full.json", "../escape.json"} {
		var out, err strings.Builder
		if runSkillAblationCommand(root, options{resultRoot: "records", output: output}, &out, &err) == 0 {
			t.Fatalf("accepted %s", output)
		}
	}
}
func TestSkillAblationChildProcess(t *testing.T) {
	if os.Getenv("KB_ABLATION_CHILD") != "" {
		os.Exit(runSkillAblationCommand(os.Getenv("KB_ABLATION_ROOT"), options{resultRoot: "records", output: "child-report.json", json: true}, os.Stdout, os.Stderr))
	}
	root := t.TempDir()
	for _, c := range []string{"full", "reduced", "none"} {
		newAblationFixture(t, root, c).save()
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestSkillAblationChildProcess$")
	cmd.Env = append(os.Environ(), "KB_ABLATION_CHILD=1", "KB_ABLATION_ROOT="+root)
	raw, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("%v %s", e, raw)
	}
	var r ablationReport
	if e = json.Unmarshal(raw, &r); e != nil || len(r.Matched) != 1 {
		t.Fatalf("%v %s", e, raw)
	}
}

func TestSkillAblationAdditionalCaptureBoundaries(t *testing.T) {
	for _, mode := range []string{"no-output", "no-native", "synthetic", "route-only", "duplicate-case-key", "unvisited-output", "unknown-metric"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			f := newAblationFixture(t, root, "full")
			switch mode {
			case "no-output":
				f.proof.Outputs = nil
			case "no-native":
				f.inv.Instructions[0].Scope = "project"
			case "unknown-metric":
				v := 12.0
				f.proof.Metrics.Tokens = ablationMetric{Value: &v, Source: "model"}
			}
			row := f.save()
			switch mode {
			case "synthetic", "route-only":
				row.Kind = mode
				f.json("records/run-full.json", row)
			case "duplicate-case-key":
				raw, _ := os.ReadFile(filepath.Join(root, row.Proof.Path))
				raw = []byte(strings.Replace(string(raw), `"exit_code":0`, `"exit_code":1,"EXIT_CODE":0`, 1))
				row.Proof = f.bytes(row.Proof.Path, raw)
				f.json("records/run-full.json", row)
			case "unvisited-output":
				row.Kind = "synthetic"
				f.json("records/run-full.json", row)
				before, _ := os.ReadFile(filepath.Join(root, row.Proof.Path))
				var out, err strings.Builder
				if runSkillAblationCommand(root, options{resultRoot: "records", output: row.Proof.Path}, &out, &err) == 0 {
					t.Fatal("overwrote excluded record evidence")
				}
				after, _ := os.ReadFile(filepath.Join(root, row.Proof.Path))
				if string(before) != string(after) {
					t.Fatal("evidence changed")
				}
				return
			}
			r := ablationRun(t, root)
			if mode == "unknown-metric" {
				if r.Eligible != 1 {
					t.Fatalf("%+v", r)
				}
				out, _, e := validateAblation(&ablationReader{root: root, used: map[string]bool{}}, row)
				if e != nil || out.Metrics["tokens"] != nil {
					t.Fatalf("metric fabricated %v %+v", e, out)
				}
				return
			}
			if r.Excluded != 1 {
				t.Fatalf("%+v", r)
			}
		})
	}
}
func TestSkillAblationSymlinkEvidence(t *testing.T) {
	root := t.TempDir()
	f := newAblationFixture(t, root, "full")
	row := f.save()
	target := filepath.Join(root, "artifacts", "oracle-link")
	if e := os.Symlink(filepath.Join(root, "artifacts", "oracle.txt"), target); e != nil {
		t.Skipf("symlink unavailable: %v", e)
	}
	f.contract.Oracle.Path = "artifacts/oracle-link"
	row = f.save()
	r := ablationRun(t, root)
	if r.Excluded != 1 {
		t.Fatalf("%+v %v", r, row)
	}
}
