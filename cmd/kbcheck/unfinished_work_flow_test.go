package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type continuationItem struct {
	ID   string `json:"id"`
	Ref  string `json:"ref"`
	Tip  string `json:"tip"`
	Kind string `json:"kind"`
}
type continuationResult struct {
	Status    string `json:"status"`
	Next      string `json:"next_action"`
	Reason    string `json:"reason"`
	Workspace string `json:"workspace"`
	Manifest  string `json:"manifest"`
	Notice    struct {
		Emit   bool               `json:"emit"`
		ID     string             `json:"offer_id"`
		Status string             `json:"status"`
		Items  []continuationItem `json:"items"`
	} `json:"notice"`
	Cleanup struct {
		Status    string             `json:"status"`
		Mode      string             `json:"mode"`
		Completed bool               `json:"completed"`
		Merge     bool               `json:"merge_authorized"`
		Items     []continuationItem `json:"items"`
		Changed   []string           `json:"changed_item_ids"`
	} `json:"cleanup"`
}

func runContinuation(t *testing.T, script, repo, request string) continuationResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := portableConsumerCommand(t, ctx, script, repo, "continue", request, "")
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("continue: %v\n%s", e, b)
	}
	var result continuationResult
	if e = json.Unmarshal(b, &result); e != nil {
		t.Fatalf("JSON: %v\n%s", e, b)
	}
	if result.Cleanup.Completed || result.Cleanup.Merge {
		t.Fatal("advisory claimed cleanup completion or merge permission")
	}
	return result
}
func TestPortableRecoveryContinue(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows PowerShell installed runtime")
	}
	temp := t.TempDir()
	repo := filepath.Join(temp, "source")
	remote := filepath.Join(temp, "remote.git")
	os.MkdirAll(repo, 0755)
	portableGit(t, temp, "init", "--bare", "--initial-branch=trunk", remote)
	portableGit(t, repo, "init", "--initial-branch=trunk")
	portableGit(t, repo, "config", "user.name", "Fixture")
	portableGit(t, repo, "config", "user.email", "fixture@example.invalid")
	portableWrite(t, filepath.Join(repo, ".gitignore"), "docs/plans/\n")
	portableWrite(t, filepath.Join(repo, "source.txt"), "baseline\n")
	portableGit(t, repo, "add", ".")
	portableGit(t, repo, "commit", "-m", "baseline")
	baseline := portableGit(t, repo, "rev-parse", "HEAD")
	portableGit(t, repo, "remote", "add", "origin", remote)
	portableGit(t, repo, "push", "-u", "origin", "trunk")
	portableGit(t, repo, "checkout", "-b", "codex/prior")
	for i := 0; i < 14; i++ {
		portableGit(t, repo, "commit", "--allow-empty", "-m", "prior")
	}
	portableGit(t, repo, "push", "-u", "origin", "codex/prior")
	portableGit(t, repo, "checkout", "trunk")
	script := installedRecoveryScript(t, "agents")
	// Clean default still inventories another unmerged branch, not just current HEAD.
	clean := portableRunSurvey(t, script, repo)
	if clean.Default.Ahead != 0 {
		t.Fatal("fixture not on clean default")
	}
	portableWrite(t, filepath.Join(repo, "docs/plans/manifest.md"), "fixture manifest; existing gate must be revalidated\n")
	survey := portableRunSurvey(t, script, repo)
	artifacts := []map[string]string{}
	for _, p := range survey.Dirty {
		if p.Path == "docs/plans/manifest.md" {
			artifacts = append(artifacts, map[string]string{"path": p.Path, "sha256": p.Hash})
		}
	}
	prepPath := filepath.Join(temp, "prepare.json")
	prep := map[string]any{"schema_version": 1, "run_id": "advisory", "objective": "new feature", "survey": survey, "branch": "codex/current", "destination": filepath.Join(temp, ".kb-recovery-worktrees/current"), "dependency_status": "independent", "authority": map[string]any{"source": "current-run", "prepare": true, "objective": "new feature"}, "artifacts": artifacts}
	writePreparationJSON(t, prepPath, prep)
	requestPath := filepath.Join(temp, "continue.json")
	request := map[string]any{"schema_version": 1, "run_id": "advisory", "objective": "new feature", "preparation_request": prepPath, "manifest": "docs/plans/manifest.md", "authority": map[string]any{"source": "current-run", "continue": true, "objective": "new feature"}}
	writePreparationJSON(t, requestPath, request)
	index := string(portableRead(t, filepath.Join(repo, ".git/index")))
	prior := portableGit(t, repo, "rev-parse", "codex/prior")
	first := runContinuation(t, script, repo, requestPath)
	if first.Status != "ready" || first.Next != "kb-work" || !first.Notice.Emit || first.Cleanup.Status != "not-authorized" {
		t.Fatalf("unanswered failed continuation: %+v", first)
	}
	branchID := ""
	for _, item := range first.Notice.Items {
		if item.Ref == "refs/heads/codex/prior" && item.Tip == prior {
			branchID = item.ID
		}
	}
	if branchID == "" {
		t.Fatal("clean default missed pushed unmerged branch")
	}
	if portableGit(t, first.Workspace, "rev-parse", "HEAD") != baseline || portableGit(t, repo, "rev-parse", "codex/prior") != prior || string(portableRead(t, filepath.Join(repo, ".git/index"))) != index {
		t.Fatal("silence damaged source or imported old commits")
	}
	if _, e := os.Stat(first.Manifest); e != nil {
		t.Fatal("no concrete manifest handoff")
	}
	again := runContinuation(t, script, repo, requestPath)
	if again.Notice.Emit || again.Notice.ID != first.Notice.ID || again.Status != "ready" {
		t.Fatalf("duplicate notice/loop: %+v", again)
	}
	request["reply"] = map[string]any{"source": "current-user-reply", "offer_id": first.Notice.ID, "response": "decline"}
	writePreparationJSON(t, requestPath, request)
	if r := runContinuation(t, script, repo, requestPath); r.Status != "ready" || r.Cleanup.Status != "not-authorized" {
		t.Fatalf("decline blocked: %+v", r)
	}
	request["reply"] = map[string]any{"source": "current-user-reply", "offer_id": first.Notice.ID, "response": "review", "item_ids": []string{branchID}}
	writePreparationJSON(t, requestPath, request)
	if r := runContinuation(t, script, repo, requestPath); r.Status != "ready" || r.Cleanup.Mode != "review" || len(r.Cleanup.Items) != 1 {
		t.Fatalf("review scope: %+v", r)
	}
	request["reply"].(map[string]any)["response"] = "cleanup"
	request["solo_auto_merge"] = true
	writePreparationJSON(t, requestPath, request)
	if r := runContinuation(t, script, repo, requestPath); r.Status != "ready" || r.Cleanup.Mode != "cleanup" || len(r.Cleanup.Items) != 1 {
		t.Fatalf("cleanup scope: %+v", r)
	}
	for _, response := range []string{"revoked", "paused", "unknown-action"} {
		request["reply"].(map[string]any)["response"] = response
		writePreparationJSON(t, requestPath, request)
		if r := runContinuation(t, script, repo, requestPath); r.Status != "ready" || len(r.Cleanup.Items) != 0 || r.Cleanup.Status == "accepted-scope" {
			t.Fatalf("invalid or withdrawn reply accepted: %+v", r)
		}
	}
	request["reply"].(map[string]any)["response"] = "cleanup"
	writePreparationJSON(t, requestPath, request)
	// Advance the offered ref without changing source/default or selecting a new scope.
	cmd := exec.Command("git", "-C", repo, "commit-tree", portableGit(t, repo, "rev-parse", prior+"^{tree}"), "-p", prior, "-m", "late commit")
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("commit-tree: %v %s", e, b)
	}
	newTip := strings.TrimSpace(string(b))
	portableGit(t, repo, "update-ref", "refs/heads/codex/prior", newTip, prior)
	late := runContinuation(t, script, repo, requestPath)
	if late.Status != "ready" || len(late.Cleanup.Items) != 0 || len(late.Cleanup.Changed) != 1 || !late.Notice.Emit {
		t.Fatalf("late acceptance enlarged scope: %+v", late)
	}
	// Advisory persistence failure cannot stop a proven independent preparation.
	portableWrite(t, filepath.Join(repo, ".git/.copilot-kb/recovery/notices/advisory.json"), "broken json")
	if r := runContinuation(t, script, repo, requestPath); r.Status != "ready" || r.Notice.Status != "advisory-unavailable" {
		t.Fatalf("advisory failure stopped work: %+v", r)
	}
	prep["dependency_status"] = "unknown"
	writePreparationJSON(t, prepPath, prep)
	if r := runContinuation(t, script, repo, requestPath); r.Status != "dependency-needed" || r.Next != "continue-other-ready-work" {
		t.Fatalf("global dependency stop: %+v", r)
	}
	prep["dependency_status"] = "independent"
	writePreparationJSON(t, prepPath, prep)
	portableWrite(t, filepath.Join(repo, ".git/.copilot-kb/work-queue.json"), `[{"branch":"codex/current","status":"active"}]`)
	if r := runContinuation(t, script, repo, requestPath); r.Status != "dependency-needed" || r.Next != "continue-other-ready-work" || r.Reason != "destination-claimed" {
		t.Fatalf("ownership conflict lost scope: %+v", r)
	}
}

func TestUnfinishedWorkFlowContract(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, path := range []string{".github/skills/kb-start/references/unfinished-work.md", ".github/skills/w2d/SKILL.md", ".github/skills/kb-complete/SKILL.md"} {
		content := string(portableRead(t, filepath.Join(root, path)))
		if !strings.Contains(content, "kb-work") {
			t.Fatalf("missing executable owner in %s", path)
		}
	}
	w2d := string(portableRead(t, filepath.Join(root, ".github/skills/w2d/SKILL.md")))
	for _, required := range []string{"packaging and browser proof pending", "ready dependent slices", "Do not close the slice", "Current explicit"} {
		if !strings.Contains(w2d, required) {
			t.Fatalf("HB unfinished-proof regression missing %q", required)
		}
	}
	// This is a routing-contract assertion, not observed live agent behavior.
	var fixture struct {
		Expected struct {
			Route        string `json:"route"`
			MaxQuestions int    `json:"max_user_questions"`
		} `json:"expected"`
	}
	if e := json.Unmarshal(portableRead(t, filepath.Join(root, "evals/route-complexity/unfinished-branch-advisory.json")), &fixture); e != nil {
		t.Fatal(e)
	}
	if fixture.Expected.Route != "w2d" || fixture.Expected.MaxQuestions != 0 {
		t.Fatal("fixture routes unfinished proof into another user confirmation")
	}
}

// Scripted state-machine proof, not a conversational agent driver or real UI
// validation. These fixture files model proof arriving through its owning lane.
func TestUnfinishedWorkFlowHBProofPrecedesDependenciesAndDelivery(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "hb-manifest.md")
	build := filepath.ToSlash(filepath.Join(root, "build.log"))
	packaging := filepath.ToSlash(filepath.Join(root, "packaging.log"))
	browser := filepath.ToSlash(filepath.Join(root, "browser.log"))
	portableWrite(t, build, "fixture: WASM build passed")
	writeHB := func(status, gateStatus, proof string) {
		t.Helper()
		portableWrite(t, path, fmt.Sprintf(`---
gate_ledger:
  - gate_id: slice-HB02-to-done
    status: %s
    required_evidence: [build, packaging, browser]
    proof: [%s]
    blockers: []
    passed_at: "2026-09-06"
    allowed_next_action: kb-work
  - gate_id: complete-to-ship
    status: pending
    required_evidence: [HB03, HB04, HB05]
    proof: []
    blockers: [HB03, HB04, HB05]
    allowed_next_action: kb-ship
slices:
  - id: HB01
    status: done
    blockers: []
  - id: HB02
    status: %s
    blockers: [HB01]
    next_agent_action: finish packaging and browser proof
  - id: HB03
    status: pending
    blockers: [HB02]
    hitl: false
    can_continue_other_slices: true
  - id: HB04
    status: pending
    blockers: [HB02]
    hitl: false
    can_continue_other_slices: true
  - id: HB05
    status: pending
    blockers: [HB02]
    hitl: false
    can_continue_other_slices: true
---
`, gateStatus, proof, status))
	}
	gate := func(id, next string) int {
		t.Helper()
		var out, stderr strings.Builder
		return run([]string{"gate-ledger", "--manifest", path, "--gate", id, "--allowed-next", next}, &out, &stderr)
	}
	writeHB("in_progress", "pending", build)
	before := string(portableRead(t, path))
	ready, err := computeReadySet(path)
	readyState, ok := ready.(readySetResult)
	if err != nil || !ok || len(readyState.Ready) != 0 {
		t.Fatalf("HB03-05 started before HB02 proof: %+v %v", ready, err)
	}
	if gate("slice-HB02-to-done", "kb-work") == 0 || gate("complete-to-ship", "kb-ship") == 0 {
		t.Fatal("in-progress update admitted completion or shipment")
	}
	if string(portableRead(t, path)) != before {
		t.Fatal("read-only progress/gate probes changed unfinished work")
	}
	// A premature status declaration cannot replace missing packaging/browser proof.
	writeHB("in_progress", "passed", build)
	if gate("slice-HB02-to-done", "kb-work") == 0 {
		t.Fatal("build-only result closed HB02")
	}
	writeHB("in_progress", "passed", strings.Join([]string{build, packaging, browser}, ", "))
	if gate("slice-HB02-to-done", "kb-work") == 0 {
		t.Fatal("nonexistent proof artifacts closed HB02")
	}
	portableWrite(t, packaging, "fixture: fresh assets packaged")
	portableWrite(t, browser, "fixture: browser assertions passed")
	if gate("slice-HB02-to-done", "kb-work") != 0 {
		t.Fatal("complete fixture proof did not unlock HB02 gate")
	}
	// The existing owner closes the slice only after that gate accepts the proof.
	writeHB("done", "passed", strings.Join([]string{build, packaging, browser}, ", "))
	ready, err = computeReadySet(path)
	readyState, ok = ready.(readySetResult)
	if err != nil || !ok || strings.Join(readyState.Ready, ",") != "HB03,HB04,HB05" {
		t.Fatalf("proven HB02 did not unlock dependents: %+v %v", ready, err)
	}
	if gate("complete-to-ship", "kb-ship") == 0 {
		t.Fatal("unfinished dependent slices admitted a premature PR")
	}
}

func TestUnfinishedWorkFlowOrderedPhaseTraceEvaluator(t *testing.T) {
	root := t.TempDir()
	edge := func(next, owner string) map[string]any { return map[string]any{"next_state": next, "owner": owner} }
	states := map[string]any{
		"HB02-built":     map[string]any{"progress": edge("HB02-built", "agent"), "package": edge("packaged", "kb-work")},
		"packaged":       map[string]any{"browser-pass": edge("browser-proven", "kb-qa"), "browser-fail": map[string]any{"next_state": "needs-repair", "owner": "kb-qa", "requires_reason": true}},
		"needs-repair":   map[string]any{"repair": edge("packaged", "kb-repair")},
		"browser-proven": map[string]any{"close-HB02": edge("HB02-done", "kb-work")},
		"HB02-done":      map[string]any{"complete-HB03": edge("HB03-done", "kb-work")},
		"HB03-done":      map[string]any{"complete-HB04": edge("HB04-done", "kb-work")},
		"HB04-done":      map[string]any{"complete-HB05": edge("implemented", "kb-work")},
		"implemented":    map[string]any{"integrated-proof": edge("proven", "kb-finalize")},
		"proven":         map[string]any{"review": edge("reviewed", "kb-review")},
		"reviewed":       map[string]any{"topic-push": edge("pushed", "kb-ship")},
		"pushed":         map[string]any{"create-pr": edge("awaiting-merge", "kb-ship")},
		"awaiting-merge": map[string]any{"merge": edge("delivered", "kb-land"), "merge-refused": map[string]any{"next_state": "awaiting-review", "owner": "kb-land", "requires_reason": true}},
		"delivered":      map[string]any{}, "awaiting-review": map[string]any{},
	}
	fixture := map[string]any{"id": "hb-phases", "expected": map[string]any{"route": "w2d", "max_user_questions": 0}, "phase_trace_contract": map[string]any{"initial_state": "HB02-built", "terminal_states": []string{"delivered", "awaiting-review"}, "states": states}}
	fixturePath := filepath.Join(root, "evals/route-complexity/hb-phases.json")
	writePreparationJSON(t, fixturePath, fixture)
	// Fake phase/forge boundaries emit records independently of the fixture graph.
	// The production public evaluator, not this recorder, decides admissibility.
	record := func(action, owner string) any { return map[string]any{"action": action, "owner": owner} }
	trace := []any{record("progress", "agent"), record("package", "kb-work"), record("browser-pass", "kb-qa"), record("close-HB02", "kb-work"), record("complete-HB03", "kb-work"), record("complete-HB04", "kb-work"), record("complete-HB05", "kb-work"), record("integrated-proof", "kb-finalize"), record("review", "kb-review"), record("topic-push", "kb-ship"), record("create-pr", "kb-ship"), record("merge", "kb-land")}
	base := map[string]any{"id": "hb-run", "fixture_id": "hb-phases", "actual": map[string]any{"route": "w2d", "user_questions": 0}, "trace": map[string]any{"files_read": []string{}, "commands": []string{}}, "phase_trace": trace, "final_state": "delivered"}
	clone := func() map[string]any {
		b, _ := json.Marshal(base)
		var c map[string]any
		if err := json.Unmarshal(b, &c); err != nil {
			t.Fatal(err)
		}
		return c
	}
	evaluate := func(result map[string]any, wantPass bool, issue string) {
		t.Helper()
		resultPath := filepath.Join(root, "result.json")
		writePreparationJSON(t, resultPath, result)
		var out, stderr strings.Builder
		code := run([]string{"skill-eval", "--root", root, "--result-path", resultPath, "--json"}, &out, &stderr)
		if (code == 0) != wantPass {
			t.Fatalf("public evaluator code=%d wantPass=%v output=%s stderr=%s", code, wantPass, out.String(), stderr.String())
		}
		if issue != "" && !strings.Contains(out.String(), issue) {
			t.Fatalf("missing rejection %q: %s", issue, out.String())
		}
		if !strings.Contains(out.String(), "supplied-trace-consistency-only") {
			t.Fatal("phase evidence presented without its confidence limit")
		}
	}
	t.Run("normal", func(t *testing.T) { evaluate(clone(), true, "") })
	t.Run("repair-retry", func(t *testing.T) {
		r := clone()
		a := r["phase_trace"].([]any)
		events := append([]any{}, a[:2]...)
		events = append(events, map[string]any{"action": "browser-fail", "owner": "kb-qa", "reason": "fake browser boundary: assertion failed"}, record("repair", "kb-repair"))
		events = append(events, a[2:]...)
		r["phase_trace"] = events
		evaluate(r, true, "")
	})
	t.Run("merge-refusal", func(t *testing.T) {
		r := clone()
		a := r["phase_trace"].([]any)
		a[len(a)-1] = map[string]any{"action": "merge-refused", "owner": "kb-land", "reason": "fake forge boundary: required review outstanding"}
		r["final_state"] = "awaiting-review"
		evaluate(r, true, "")
	})
	for _, spec := range []struct {
		name, issue string
		modify      func(map[string]any)
	}{
		{"premature-terminal", "premature terminal", func(r map[string]any) { r["phase_trace"] = r["phase_trace"].([]any)[:1] }},
		{"ask-continue", "not allowed", func(r map[string]any) { r["phase_trace"].([]any)[1] = record("ask-continue", "agent") }},
		{"out-of-order-proof", "not allowed", func(r map[string]any) { r["phase_trace"].([]any)[2] = record("close-HB02", "kb-work") }},
		{"browser-failure-no-repair", "not allowed", func(r map[string]any) {
			r["phase_trace"].([]any)[2] = map[string]any{"action": "browser-fail", "owner": "kb-qa", "reason": "assertion failed"}
		}},
		{"wrong-owner", "wrong owner", func(r map[string]any) { r["phase_trace"].([]any)[2] = record("browser-pass", "kb-work") }},
		{"missing-trace", "requires nonempty", func(r map[string]any) { delete(r, "phase_trace") }},
		{"invented-refusal", "requires refusal", func(r map[string]any) {
			a := r["phase_trace"].([]any)
			a[len(a)-1] = record("merge-refused", "kb-land")
			r["final_state"] = "awaiting-review"
		}},
		{"false-final-state", "final_state differs", func(r map[string]any) { r["final_state"] = "awaiting-review" }},
	} {
		t.Run(spec.name, func(t *testing.T) { r := clone(); spec.modify(r); evaluate(r, false, spec.issue) })
	}
	t.Run("malformed-contract", func(t *testing.T) {
		fixture["phase_trace_contract"] = "bad"
		writePreparationJSON(t, fixturePath, fixture)
		evaluate(clone(), false, "malformed fixture contract")
	})
}
