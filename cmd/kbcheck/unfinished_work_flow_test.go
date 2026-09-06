package main

import (
	"context"
	"encoding/json"
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
	ps, e := exec.LookPath("powershell.exe")
	if e != nil {
		t.Fatal(e)
	}
	git, e := exec.LookPath("git")
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ps, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", script, "-Root", repo, "-Action", "continue", "-Request", request, "-Json")
	for _, v := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(v), "PATH=") {
			cmd.Env = append(cmd.Env, v)
		}
	}
	cmd.Env = append(cmd.Env, "PATH="+filepath.Dir(git)+";"+filepath.Join(os.Getenv("SystemRoot"), "System32"))
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
	script := filepath.Join(temp, "installed/recovery.ps1")
	for _, name := range []string{"recovery.ps1", "recovery_prepare.ps1", "recovery_continue.ps1"} {
		portableWrite(t, filepath.Join(filepath.Dir(script), name), string(portableRead(t, filepath.Join("..", "..", ".github/skills/kb-rehab/scripts", name))))
	}
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
