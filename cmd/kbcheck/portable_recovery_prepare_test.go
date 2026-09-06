package main

import (
	"context"
	"crypto/sha256"
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

type preparationResult struct {
	Status      string `json:"status"`
	Reason      string `json:"reason"`
	Receipt     string `json:"receipt"`
	Destination string `json:"destination"`
	Next        string `json:"next_action"`
}

func TestPortableRecoveryPrepareReservedBranch(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows PowerShell installed runtime")
	}
	temp := t.TempDir()
	repo := filepath.Join(temp, "source")
	remote := filepath.Join(temp, "remote.git")
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatal(err)
	}
	portableGit(t, temp, "init", "--bare", "--initial-branch=trunk", remote)
	portableGit(t, repo, "init", "--initial-branch=trunk")
	portableGit(t, repo, "config", "user.name", "Fixture")
	portableGit(t, repo, "config", "user.email", "fixture@example.invalid")
	portableWrite(t, filepath.Join(repo, "source.txt"), "baseline\n")
	portableGit(t, repo, "add", ".")
	portableGit(t, repo, "commit", "-m", "baseline")
	portableGit(t, repo, "remote", "add", "origin", remote)
	portableGit(t, repo, "push", "-u", "origin", "trunk")
	script := filepath.Join(temp, "installed/recovery.ps1")
	for _, name := range []string{"recovery.ps1", "recovery_prepare.ps1"} {
		portableWrite(t, filepath.Join(filepath.Dir(script), name), string(portableRead(t, filepath.Join("..", "..", ".github/skills/kb-rehab/scripts", name))))
	}
	survey := portableRunSurvey(t, script, repo)
	dest := filepath.Join(temp, ".kb-recovery-worktrees", "reserved")
	request := map[string]any{"schema_version": 1, "run_id": "reserved", "objective": "new feature", "survey": survey, "branch": "codex/reserved", "destination": dest, "dependency_status": "independent", "authority": map[string]any{"source": "current-run", "prepare": true, "objective": "new feature"}, "artifacts": []any{}}
	requestPath := filepath.Join(temp, "request.json")
	writePreparationJSON(t, requestPath, request)
	requestDigest := fmt.Sprintf("%x", sha256.Sum256(portableRead(t, requestPath)))
	// Durable reservation plus branch models interruption before directory creation.
	portableGit(t, repo, "branch", "codex/reserved", survey.Head)
	receipt := map[string]any{"schema_version": 1, "run_id": "reserved", "objective": "new feature", "request_sha256": requestDigest, "repository_id": survey.RepositoryID, "source_head": survey.Head, "source_dirty_fingerprint": survey.DirtyFingerprint, "source_index_sha256": survey.IndexHash, "baseline_sha": survey.Head, "branch": "codex/reserved", "destination": dest, "state": "reserved", "artifacts": []any{}, "next_action": "continue-independent-work"}
	writePreparationJSON(t, filepath.Join(repo, ".git/.copilot-kb/recovery/reserved.json"), receipt)
	if got := runPreparation(t, script, repo, requestPath, "prepare"); got.Status != "prepared" {
		t.Fatalf("reserved branch resume: %+v", got)
	}
	if portableGit(t, dest, "rev-parse", "HEAD") != survey.Head {
		t.Fatal("reserved branch wrong baseline")
	}
}

func runPreparation(t *testing.T, script, repo, request, action string) preparationResult {
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
	cmd := exec.CommandContext(ctx, ps, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", script, "-Root", repo, "-Action", action, "-Request", request, "-Json")
	for _, v := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(v), "PATH=") {
			cmd.Env = append(cmd.Env, v)
		}
	}
	cmd.Env = append(cmd.Env, "PATH="+filepath.Dir(git)+";"+filepath.Join(os.Getenv("SystemRoot"), "System32"))
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("prepare: %v\n%s", e, b)
	}
	var result preparationResult
	if e = json.Unmarshal(b, &result); e != nil {
		t.Fatalf("JSON: %v\n%s", e, b)
	}
	return result
}
func writePreparationJSON(t *testing.T, path string, value any) {
	t.Helper()
	b, e := json.Marshal(value)
	if e != nil {
		t.Fatal(e)
	}
	portableWrite(t, path, string(b))
}
func TestPortableRecoveryPreparePreserve(t *testing.T) {
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
	portableWrite(t, filepath.Join(repo, ".gitattributes"), "*.txt filter=unsafe diff=unsafe\n")
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
	marker := filepath.Join(temp, "filter-executed")
	unsafeCommand := "cmd /c echo executed > " + filepath.ToSlash(marker)
	portableGit(t, repo, "config", "filter.unsafe.clean", unsafeCommand)
	portableGit(t, repo, "config", "filter.unsafe.smudge", unsafeCommand)
	portableGit(t, repo, "config", "diff.unsafe.textconv", unsafeCommand)
	portableGit(t, repo, "config", "diff.external", unsafeCommand)
	portableGit(t, repo, "config", "core.fsmonitor", unsafeCommand)
	portableWrite(t, filepath.Join(repo, "source.txt"), "unrelated dirty source\n")
	portableWrite(t, filepath.Join(repo, "docs/plans/a plan.md"), "selected\r\n")
	portableWrite(t, filepath.Join(repo, "docs/plans/empty.md"), "")
	script := filepath.Join(temp, "installed/recovery.ps1")
	for _, name := range []string{"recovery.ps1", "recovery_prepare.ps1"} {
		portableWrite(t, filepath.Join(filepath.Dir(script), name), string(portableRead(t, filepath.Join("..", "..", ".github/skills/kb-rehab/scripts", name))))
	}
	survey := portableRunSurvey(t, script, repo)
	artifacts := []map[string]string{}
	for _, p := range survey.Dirty {
		if strings.HasPrefix(p.Path, "docs/plans/") {
			artifacts = append(artifacts, map[string]string{"path": p.Path, "sha256": p.Hash})
		}
	}
	dest := filepath.Join(temp, ".kb-recovery-worktrees", "new-work")
	request := map[string]any{"schema_version": 1, "run_id": "prepare-test", "objective": "new feature", "survey": survey, "branch": "codex/new-work", "destination": dest, "dependency_status": "independent", "authority": map[string]any{"source": "current-run", "prepare": true, "objective": "new feature"}, "artifacts": artifacts}
	requestPath := filepath.Join(temp, "request.json")
	writePreparationJSON(t, requestPath, request)
	head := portableGit(t, repo, "rev-parse", "HEAD")
	index := string(portableRead(t, filepath.Join(repo, ".git/index")))
	source := string(portableRead(t, filepath.Join(repo, "source.txt")))
	got := runPreparation(t, script, repo, requestPath, "prepare")
	if got.Status != "prepared" || got.Next != "kb-work" {
		t.Fatalf("prepare failed: %+v", got)
	}
	if portableGit(t, dest, "rev-parse", "HEAD") != baseline || portableGit(t, repo, "rev-parse", "HEAD") != head || string(portableRead(t, filepath.Join(repo, ".git/index"))) != index || string(portableRead(t, filepath.Join(repo, "source.txt"))) != source {
		t.Fatal("incorrect baseline or source mutated")
	}
	if string(portableRead(t, filepath.Join(dest, "docs/plans/a plan.md"))) != "selected\r\n" || len(portableRead(t, filepath.Join(dest, "docs/plans/empty.md"))) != 0 {
		t.Fatal("copy changed CRLF/empty bytes")
	}
	refs := portableGit(t, repo, "show-ref")
	if again := runPreparation(t, script, repo, requestPath, "prepare"); again.Status != "prepared" || again.Receipt != got.Receipt || portableGit(t, repo, "show-ref") != refs {
		t.Fatalf("not idempotent: %+v", again)
	}
	if verify := runPreparation(t, script, repo, requestPath, "verify"); verify.Status != "prepared" {
		t.Fatalf("verify: %+v", verify)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("source-configured filter/fsmonitor executed")
	}
	// Resume the durable state left by interruption after worktree creation.
	var interrupted map[string]any
	if err := json.Unmarshal(portableRead(t, got.Receipt), &interrupted); err != nil {
		t.Fatal(err)
	}
	interrupted["state"] = "reserved"
	writePreparationJSON(t, got.Receipt, interrupted)
	if err := os.Remove(filepath.Join(dest, "source.txt")); err != nil {
		t.Fatal(err)
	}
	if resumed := runPreparation(t, script, repo, requestPath, "prepare"); resumed.Status != "prepared" {
		t.Fatalf("interrupted creation: %+v", resumed)
	}
	if strings.TrimSpace(string(portableRead(t, filepath.Join(dest, "source.txt")))) != "baseline" {
		t.Fatal("resume did not restore missing baseline file")
	}
	interrupted["state"] = "reserved"
	writePreparationJSON(t, got.Receipt, interrupted)
	stagedBlob := portableGit(t, dest, "hash-object", "-w", "docs/plans/a plan.md")
	portableGit(t, dest, "update-index", "--cacheinfo", "100644,"+stagedBlob+",source.txt")
	destinationIndex := portableGit(t, dest, "rev-parse", "--git-path", "index")
	stagedIndex := string(portableRead(t, destinationIndex))
	if r := runPreparation(t, script, repo, requestPath, "prepare"); r.Reason != "destination-staged-work-preserved" {
		t.Fatalf("resumed over staged work: %+v", r)
	}
	if string(portableRead(t, destinationIndex)) != stagedIndex {
		t.Fatal("resumed preparation changed destination index")
	}
	baselineBlob := portableGit(t, repo, "rev-parse", baseline+":source.txt")
	portableGit(t, dest, "update-index", "--cacheinfo", "100644,"+baselineBlob+",source.txt")
	interrupted["state"] = "prepared"
	writePreparationJSON(t, got.Receipt, interrupted)
	// A partially copied selected set resumes without duplicating directories.
	os.Remove(filepath.Join(dest, "docs/plans/empty.md"))
	if verify := runPreparation(t, script, repo, requestPath, "verify"); verify.Reason != "artifact-copy-incomplete" {
		t.Fatalf("partial verify: %+v", verify)
	}
	if resume := runPreparation(t, script, repo, requestPath, "prepare"); resume.Status != "prepared" {
		t.Fatalf("partial resume: %+v", resume)
	}
	portableWrite(t, filepath.Join(dest, "docs/plans/a plan.md"), "user edit\n")
	if changed := runPreparation(t, script, repo, requestPath, "prepare"); changed.Reason != "destination-artifact-changed" {
		t.Fatalf("overwrote resumed edit: %+v", changed)
	}
	portableWrite(t, filepath.Join(dest, "docs/plans/a plan.md"), "selected\r\n")
	portableWrite(t, filepath.Join(repo, "docs/plans/a plan.md"), "source changed\n")
	if changed := runPreparation(t, script, repo, requestPath, "prepare"); changed.Reason != "survey-refresh-required" {
		t.Fatalf("accepted stale source: %+v", changed)
	}
	portableWrite(t, filepath.Join(repo, "docs/plans/a plan.md"), "selected\r\n")
	t.Run("EscapeCredentialCollisionAndDependency", func(t *testing.T) {
		original := request["artifacts"]
		for _, bad := range []string{"../outside.md", "docs/plans/credentials.md"} {
			request["artifacts"] = []map[string]string{{"path": bad, "sha256": artifacts[0]["sha256"]}}
			writePreparationJSON(t, requestPath, request)
			if r := runPreparation(t, script, repo, requestPath, "prepare"); r.Status == "prepared" {
				t.Fatalf("accepted %s", bad)
			}
		}
		request["artifacts"] = []map[string]string{artifacts[0], {"path": strings.ToUpper(artifacts[0]["path"]), "sha256": artifacts[0]["sha256"]}}
		writePreparationJSON(t, requestPath, request)
		if r := runPreparation(t, script, repo, requestPath, "prepare"); r.Reason != "artifact-case-collision" {
			t.Fatalf("case collision: %+v", r)
		}
		request["artifacts"] = original
		request["dependency_status"] = "unknown"
		writePreparationJSON(t, requestPath, request)
		if r := runPreparation(t, script, repo, requestPath, "prepare"); r.Reason != "dependency-needed" {
			t.Fatalf("dependency: %+v", r)
		}
		request["dependency_status"] = "independent"
		writePreparationJSON(t, requestPath, request)
	})
	t.Run("MovedDefault", func(t *testing.T) {
		portableGit(t, remote, "update-ref", "refs/heads/trunk", head)
		if r := runPreparation(t, script, repo, requestPath, "prepare"); r.Reason != "survey-refresh-required" {
			t.Fatalf("default changed: %+v", r)
		}
		portableGit(t, remote, "update-ref", "refs/heads/trunk", baseline)
	})
	t.Run("MalformedAuthority", func(t *testing.T) {
		authority := request["authority"].(map[string]any)
		for _, value := range []any{"true", 1, false} {
			authority["prepare"] = value
			writePreparationJSON(t, requestPath, request)
			if r := runPreparation(t, script, repo, requestPath, "prepare"); r.Reason != "current-preparation-authority-required" {
				t.Fatalf("coerced authority: %+v", r)
			}
		}
		authority["prepare"] = true
		authority["paused"] = "false"
		writePreparationJSON(t, requestPath, request)
		if r := runPreparation(t, script, repo, requestPath, "prepare"); r.Reason != "current-preparation-authority-required" {
			t.Fatalf("coerced pause: %+v", r)
		}
		delete(authority, "paused")
		writePreparationJSON(t, requestPath, request)
	})
	t.Run("ForeignAmbientGit", func(t *testing.T) {
		foreign := filepath.Join(temp, "foreign")
		if err := os.MkdirAll(foreign, 0755); err != nil {
			t.Fatal(err)
		}
		portableGit(t, foreign, "init")
		foreignIndex := filepath.Join(temp, "foreign-index")
		portableWrite(t, foreignIndex, "foreign index sentinel")
		t.Setenv("GIT_DIR", filepath.Join(foreign, ".git"))
		t.Setenv("GIT_WORK_TREE", foreign)
		t.Setenv("GIT_COMMON_DIR", filepath.Join(foreign, ".git"))
		t.Setenv("GIT_INDEX_FILE", foreignIndex)
		if r := runPreparation(t, script, repo, requestPath, "verify"); r.Status != "prepared" {
			t.Fatalf("foreign ambient redirect: %+v", r)
		}
		if string(portableRead(t, foreignIndex)) != "foreign index sentinel" {
			t.Fatal("foreign index mutated")
		}
	})
	t.Run("ActiveClaim", func(t *testing.T) {
		queue := filepath.Join(repo, ".git/.copilot-kb/work-queue.json")
		portableWrite(t, queue, `[{"branch":"codex/new-work","status":"active"}]`)
		if r := runPreparation(t, script, repo, requestPath, "prepare"); r.Reason != "destination-claimed" {
			t.Fatalf("claim stolen: %+v", r)
		}
		if err := os.Remove(queue); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("OccupiedDestination", func(t *testing.T) {
		request["run_id"] = "different-run"
		writePreparationJSON(t, requestPath, request)
		if r := runPreparation(t, script, repo, requestPath, "prepare"); r.Reason != "destination-occupied" {
			t.Fatalf("occupied: %+v", r)
		}
		request["run_id"] = "prepare-test"
		writePreparationJSON(t, requestPath, request)
	})
	t.Run("JunctionEscape", func(t *testing.T) {
		outside := filepath.Join(temp, "outside")
		os.MkdirAll(outside, 0755)
		junction := filepath.Join(dest, "docs/plans/link")
		cmd := exec.Command("cmd", "/c", "mklink", "/J", junction, outside)
		if b, e := cmd.CombinedOutput(); e != nil {
			t.Fatalf("fixture junction: %v %s", e, b)
		}
		request["artifacts"] = []map[string]string{{"path": "docs/plans/link/a.md", "sha256": artifacts[0]["sha256"]}}
		portableWrite(t, filepath.Join(repo, "docs/plans/link/a.md"), "selected\r\n")
		request["survey"] = portableRunSurvey(t, script, repo)
		writePreparationJSON(t, requestPath, request)
		if r := runPreparation(t, script, repo, requestPath, "prepare"); r.Reason != "path-link-or-junction" {
			t.Fatalf("junction: %+v", r)
		}
		if _, e := os.Stat(filepath.Join(outside, "a.md")); !os.IsNotExist(e) {
			t.Fatal("wrote through junction")
		}
	})
}
