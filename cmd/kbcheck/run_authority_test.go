package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunAuthorityHistoricalPlanDoesNotDenyCurrentLocalGrant(t *testing.T) {
	root := initWorktreeRepo(t)
	manifest := writePlanRunTestManifest(t, root, "kb-current-authority")
	body := mustReadFile(t, manifest)
	body = []byte(strings.Replace(string(body), "\n---\n", "\nplanning_only: true\nplan_run_worktree:\n  commit_authorized: false\n---\n", 1))
	if err := os.WriteFile(manifest, body, 0644); err != nil {
		t.Fatal(err)
	}
	gitOK(t, root, "add", "docs/plans/kb-current-authority.md")
	gitOK(t, root, "commit", "-m", "historical planning snapshot")
	result, err := executePlanRunWorkspace(planRunWorkspaceOptions{
		Action: "prepare", ManifestPath: manifest, OwnerToken: "current-owner",
		CommitAuthorized: true, CommitAuthorizedBy: "user", CommitApprovalRef: "explicit-p2d",
		BaseSHA: gitOutput(root, "rev-parse", "HEAD"), Worktree: filepath.Join(t.TempDir(), "current-grant"),
		IntegrationRef: "codex/current-grant", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !result.OK {
		t.Fatalf("current authority rejected: %#v %v", result, err)
	}
	if result.Receipt.CommitApprovalRef != "explicit-p2d" {
		t.Fatal("lost current provenance")
	}
	if string(mustReadFile(t, manifest)) != string(body) {
		t.Fatal("rewrote historical plan")
	}
}

func TestRunAuthorityPersistentLocalPolicyIsRetained(t *testing.T) {
	root := initWorktreeRepo(t)
	manifest := writePlanRunTestManifest(t, root, "kb-local-policy")
	policy := filepath.Join(root, "docs", "context", "operations", "kb-routing.yaml")
	if err := os.MkdirAll(filepath.Dir(policy), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(policy, []byte("schema_version: 1\ndelivery:\n  mode: local\n  merge: manual\n  post_merge_sync: false\n"), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := executePlanRunWorkspace(planRunWorkspaceOptions{Action: "prepare", ManifestPath: manifest, OwnerToken: "local-owner", CommitAuthorized: true, CommitAuthorizedBy: "user", CommitApprovalRef: "work", BaseSHA: gitOutput(root, "rev-parse", "HEAD"), Worktree: filepath.Join(t.TempDir(), "local-policy"), IntegrationRef: "codex/local-policy", RepoRoot: root, Now: time.Now().UTC()})
	if err != nil || !result.OK {
		t.Fatalf("safe local preparation blocked: %#v %v", result, err)
	}
	if result.Receipt.DeliveryMode != "local" {
		t.Fatal("local policy was broadened")
	}
}
