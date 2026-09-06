package main

import (
	"bytes"
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

type portableSurvey struct {
	Repository        string `json:"repository"`
	CommonDir         string `json:"common_dir"`
	RepositoryID      string `json:"repository_id"`
	Branch            string `json:"branch"`
	Head              string `json:"head"`
	IndexHash         string `json:"index_sha256"`
	DirtyFingerprint  string `json:"dirty_fingerprint"`
	InventoryComplete bool   `json:"inventory_complete"`
	Authority         struct {
		Status    string `json:"status"`
		Baseline  string `json:"baseline_sha"`
		Default   string `json:"default_ref"`
		Selection string `json:"selection"`
	} `json:"authority"`
	Default struct {
		Ahead  int `json:"ahead"`
		Behind int `json:"behind"`
	} `json:"default_divergence"`
	Upstream struct {
		Ahead  int    `json:"ahead"`
		Behind int    `json:"behind"`
		Status string `json:"status"`
	} `json:"upstream_divergence"`
	Dirty []struct {
		Path       string `json:"path"`
		Hash       string `json:"sha256"`
		Protection string `json:"protection"`
		Status     string `json:"status"`
	} `json:"dirty_paths"`
	Protections []struct {
		Kind   string `json:"kind"`
		Branch string `json:"branch"`
	} `json:"protections"`
	Candidates []struct {
		Status   string   `json:"status"`
		Evidence string   `json:"evidence"`
		Refs     []string `json:"declared_refs_or_hashes"`
	} `json:"candidates"`
	Eligibility struct {
		Merge  bool `json:"merge"`
		Delete bool `json:"delete"`
	} `json:"eligibility"`
	Policy struct {
		Status string `json:"status"`
	} `json:"policy"`
	Capabilities struct {
		Native  bool `json:"native_required"`
		KBCheck bool `json:"native_kbcheck"`
	} `json:"capabilities"`
}

func portableGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}
func portableWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
func portableRead(t *testing.T, path string) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func portableRunSurvey(t *testing.T, script, root string) portableSurvey {
	t.Helper()
	ps, err := exec.LookPath("powershell.exe")
	if err != nil {
		t.Fatal(err)
	}
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ps, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", script, "-Action", "survey", "-Root", root, "-Json")
	// The consumer has Git and Windows utilities, but no Go/Node/reconciler on PATH.
	for _, v := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(v), "PATH=") {
			cmd.Env = append(cmd.Env, v)
		}
	}
	cmd.Env = append(cmd.Env, "PATH="+filepath.Dir(git)+";"+filepath.Join(os.Getenv("SystemRoot"), "System32"))
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("installed survey: %v\n%s", err, b)
	}
	var result portableSurvey
	if err = json.Unmarshal(bytes.TrimPrefix(b, []byte{239, 187, 191}), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, b)
	}
	if result.Eligibility.Merge || result.Eligibility.Delete {
		t.Fatal("survey invented destructive authority")
	}
	return result
}

func TestPortableRecoverySurvey(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("installed Windows PowerShell 5.1 lane; non-Windows unverified")
	}
	temp := t.TempDir()
	remote := filepath.Join(temp, "remote.git")
	repo := filepath.Join(temp, "plain consumer")
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatal(err)
	}
	portableGit(t, temp, "init", "--bare", "--initial-branch=trunk", remote)
	portableGit(t, repo, "init", "--initial-branch=trunk")
	portableGit(t, repo, "config", "user.name", "Recovery Fixture")
	portableGit(t, repo, "config", "user.email", "fixture@example.invalid")
	portableWrite(t, filepath.Join(repo, ".gitignore"), "docs/plans/ignored*\n")
	portableWrite(t, filepath.Join(repo, "source.txt"), "baseline\n")
	portableGit(t, repo, "add", ".")
	portableGit(t, repo, "commit", "-m", "baseline")
	baseline := portableGit(t, repo, "rev-parse", "HEAD")
	portableGit(t, repo, "remote", "add", "origin", remote)
	portableGit(t, repo, "push", "-u", "origin", "trunk")
	portableGit(t, repo, "checkout", "-b", "codex/unfinished")
	for i := 0; i < 14; i++ {
		portableWrite(t, filepath.Join(repo, "source.txt"), fmt.Sprintf("prior source %d\n", i))
		portableGit(t, repo, "add", "source.txt")
		portableGit(t, repo, "commit", "-m", fmt.Sprintf("prior %d", i))
	}
	portableGit(t, repo, "push", "-u", "origin", "codex/unfinished")
	portableWrite(t, filepath.Join(repo, "docs/plans/audit.md"), "branch: codex/unfinished\nproof_sha256: deadbeef\nproof: docs/results/missing.json\ncommand: do-not-execute\n")
	portableWrite(t, filepath.Join(repo, "docs/plans/ignored empty.md"), "")
	portableWrite(t, filepath.Join(repo, "docs/plans/a plan ü.md"), "new plan\n")
	portableWrite(t, filepath.Join(repo, ".git/.copilot-kb/work-queue.json"), `[{"branch":"codex/unfinished","status":"active","updated_at":"2099-01-01T00:00:00Z"}]`)
	// Run a copied payload outside the source repo and consumer .github tree.
	script := filepath.Join(temp, "installed/kb-rehab/scripts/recovery.ps1")
	portableWrite(t, script, string(portableRead(t, filepath.Join("..", "..", ".github", "skills", "kb-rehab", "scripts", "recovery.ps1"))))
	index := portableRead(t, filepath.Join(repo, ".git/index"))
	refs := portableGit(t, repo, "show-ref")
	status := portableGit(t, repo, "status", "--porcelain=v1", "--ignored")
	fetchHead := filepath.Join(repo, ".git/FETCH_HEAD")
	portableWrite(t, fetchHead, "prior fetch marker\n")
	first := portableRunSurvey(t, script, repo)
	if first.Authority.Status != "verified" || first.Authority.Default != "refs/heads/trunk" || first.Authority.Baseline != baseline {
		t.Fatalf("wrong authority: %+v", first.Authority)
	}
	if first.Default.Ahead != 14 || first.Default.Behind != 0 || first.Upstream.Ahead != 0 || first.Upstream.Behind != 0 || first.Upstream.Status != "verified" {
		t.Fatalf("default/upstream conflated: %+v / %+v", first.Default, first.Upstream)
	}
	if first.Policy.Status != "absent-conservative-default" || first.Capabilities.Native || first.Capabilities.KBCheck {
		t.Fatal("plain consumer requires maintainer tools/policy")
	}
	if !first.InventoryComplete || len(first.Dirty) != 3 || len(first.DirtyFingerprint) != 64 || len(first.RepositoryID) != 64 {
		t.Fatalf("incomplete inventory: %+v", first)
	}
	ignored := false
	for _, p := range first.Dirty {
		if p.Path == "docs/plans/ignored empty.md" {
			ignored = p.Status == "ignored-workflow" && p.Protection == "preserve" && p.Hash != ""
		}
	}
	if !ignored {
		t.Fatal("ignored zero-byte plan was not preserved/fingerprinted")
	}
	claim := false
	for _, p := range first.Protections {
		if p.Kind == "work-claim" && p.Branch == "codex/unfinished" {
			claim = true
		}
	}
	if !claim {
		t.Fatal("live claim omitted")
	}
	if len(first.Candidates) == 0 {
		t.Fatal("declarations missing")
	}
	for _, c := range first.Candidates {
		if c.Status != "candidate-unproven" || c.Evidence != "not-validated" {
			t.Fatal("self-authored declaration became proof")
		}
	}
	if !bytes.Equal(index, portableRead(t, filepath.Join(repo, ".git/index"))) || refs != portableGit(t, repo, "show-ref") || status != portableGit(t, repo, "status", "--porcelain=v1", "--ignored") || string(portableRead(t, fetchHead)) != "prior fetch marker\n" {
		t.Fatal("survey changed refs/index/source/FETCH_HEAD")
	}
	second := portableRunSurvey(t, script, repo)
	if first.DirtyFingerprint != second.DirtyFingerprint {
		t.Fatal("unstable fingerprint")
	}
	portableWrite(t, filepath.Join(repo, "docs/plans/a plan ü.md"), "changed plan\n")
	if first.DirtyFingerprint == portableRunSurvey(t, script, repo).DirtyFingerprint {
		t.Fatal("fingerprint ignored changed bytes")
	}
	t.Run("FreshDefaultPreservesTrackingRefs", func(t *testing.T) {
		portableGit(t, remote, "update-ref", "refs/heads/trunk", first.Head)
		before := portableGit(t, repo, "show-ref")
		got := portableRunSurvey(t, script, repo)
		if got.Authority.Baseline != first.Head || got.Default.Ahead != 0 || before != portableGit(t, repo, "show-ref") {
			t.Fatal("fresh fetch used stale baseline or moved existing refs")
		}
		portableGit(t, remote, "update-ref", "refs/heads/trunk", baseline)
	})
	t.Run("UnreachableRemote", func(t *testing.T) {
		portableGit(t, repo, "remote", "set-url", "origin", filepath.Join(temp, "absent.git"))
		got := portableRunSurvey(t, script, repo)
		if got.Authority.Status != "unavailable" || got.Authority.Baseline != "" {
			t.Fatal("cached authority promoted after failed remote")
		}
		portableGit(t, repo, "remote", "set-url", "origin", remote)
	})
	t.Run("AmbiguousRemotes", func(t *testing.T) {
		portableGit(t, repo, "remote", "add", "second", remote)
		portableGit(t, repo, "config", "--unset", "branch.codex/unfinished.remote")
		got := portableRunSurvey(t, script, repo)
		if got.Authority.Status != "unavailable" {
			t.Fatal("guessed origin with ambiguous remotes")
		}
		portableGit(t, repo, "config", "branch.codex/unfinished.remote", "origin")
		if portableRunSurvey(t, script, repo).Authority.Selection != "branch-upstream" {
			t.Fatal("ignored explicit upstream")
		}
	})
	t.Run("LinkedWorktreeNoDeclarations", func(t *testing.T) {
		linked := filepath.Join(temp, "linked")
		portableGit(t, repo, "worktree", "add", "-b", "codex/independent", linked, baseline)
		got := portableRunSurvey(t, script, linked)
		if got.RepositoryID != first.RepositoryID || !strings.EqualFold(filepath.Clean(got.CommonDir), filepath.Join(repo, ".git")) || len(got.Candidates) != 0 {
			t.Fatalf("wrong common identity/missing-declaration handling: %+v", got)
		}
	})
}
