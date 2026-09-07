package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type dispositionResult struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
	Items  []struct {
		Disposition string `json:"disposition"`
		Reason      string `json:"reason"`
		Completed   bool   `json:"completed"`
		Archive     string `json:"archive"`
		Receipt     string `json:"receipt"`
		Owner       string `json:"owner"`
		ForgeChecks bool   `json:"forge_checks_required"`
	} `json:"items"`
}

func runDisposition(t *testing.T, script, repo, request string) dispositionResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := portableConsumerCommand(t, ctx, script, repo, "dispose", request, os.Getenv("KB_TEST_NATIVE_PATH"))
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("dispose: %v\n%s", e, b)
	}
	var result dispositionResult
	if e = json.Unmarshal(b, &result); e != nil {
		t.Fatalf("JSON: %v\n%s", e, b)
	}
	return result
}
func TestPortableRecoveryDisposition(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows PowerShell installed runtime")
	}
	temp, tempErr := os.MkdirTemp("", "kb-disp-")
	if tempErr != nil {
		t.Fatal(tempErr)
	}
	t.Cleanup(func() {
		if strings.HasPrefix(filepath.Clean(temp), filepath.Clean(os.TempDir())+string(os.PathSeparator)) {
			_ = os.RemoveAll(temp)
		}
	})
	repo := filepath.Join(temp, "source")
	remote := filepath.Join(temp, "remote.git")
	os.MkdirAll(repo, 0755)
	portableGit(t, temp, "init", "--bare", "--initial-branch=trunk", remote)
	portableGit(t, repo, "init", "--initial-branch=trunk")
	portableGit(t, repo, "config", "user.name", "Fixture")
	portableGit(t, repo, "config", "user.email", "fixture@example.invalid")
	portableWrite(t, filepath.Join(repo, "source.txt"), "baseline\n")
	portableGit(t, repo, "add", ".")
	portableGit(t, repo, "commit", "-m", "baseline")
	baseline := portableGit(t, repo, "rev-parse", "HEAD")
	portableGit(t, repo, "remote", "add", "origin", remote)
	portableGit(t, repo, "push", "-u", "origin", "trunk")
	portableGit(t, repo, "checkout", "-b", "codex/prior")
	portableWrite(t, filepath.Join(repo, "unique.txt"), "unique rejected content\n")
	portableGit(t, repo, "add", ".")
	portableGit(t, repo, "commit", "-m", "unique")
	tip := portableGit(t, repo, "rev-parse", "HEAD")
	portableGit(t, repo, "push", "-u", "origin", "codex/prior")
	portableGit(t, repo, "checkout", "trunk")
	portableWrite(t, filepath.Join(repo, "docs/plans/notes.md"), "uncommitted notes\r\n")
	portableWrite(t, filepath.Join(repo, ".env"), "fixture credential must stay\n")
	manifestText := "ref: refs/heads/codex/prior\ntip: " + tip + "\n"
	portableWrite(t, filepath.Join(repo, "docs/plans/owner.md"), manifestText)
	script := installedRecoveryScript(t, "agents")
	survey := portableRunSurvey(t, script, repo)
	notesHash := fmt.Sprintf("%x", sha256.Sum256(portableRead(t, filepath.Join(repo, "docs/plans/notes.md"))))
	branch := map[string]any{"kind": "branch", "ref": "refs/heads/codex/prior", "tip": tip, "decision": "reject"}
	scope := []map[string]any{{"kind": "branch", "ref": "refs/heads/codex/prior", "tip": tip}, {"kind": "artifact", "path": "docs/plans/notes.md", "sha256": notesHash}}
	accept := map[string]any{"source": "current-user-reply", "accepted": true, "items": scope}
	request := map[string]any{"schema_version": 1, "run_id": "dispose-test", "objective": "accepted backlog", "repository_id": survey.RepositoryID, "acceptance": accept, "items": []any{branch}}
	requestPath := filepath.Join(temp, "dispose.json")
	writePreparationJSON(t, requestPath, request)
	first := runDisposition(t, script, repo, requestPath)
	if len(first.Items) != 1 || first.Items[0].Disposition != "review-needed" {
		t.Fatalf("missing rejection became junk: %+v", first)
	}
	branch["decision"] = "deliver"
	branch["manifest"] = "docs/plans/owner.md"
	branch["manifest_sha256"] = fmt.Sprintf("%x", sha256.Sum256([]byte(manifestText)))
	liveOld := filepath.Join(temp, "live-old")
	portableGit(t, repo, "worktree", "add", liveOld, "codex/prior")
	writePreparationJSON(t, requestPath, request)
	if r := runDisposition(t, script, repo, requestPath); r.Items[0].Disposition != "deliver-pr" || r.Items[0].Completed {
		t.Fatalf("delivery claim: %+v", r)
	}
	branch["decision"] = "reject"
	branch["rejected"] = true
	writePreparationJSON(t, requestPath, request)
	if r := runDisposition(t, script, repo, requestPath); r.Items[0].Disposition != "preserve-live" {
		t.Fatalf("occupied branch not protected: %+v", r)
	}
	portableGit(t, repo, "worktree", "remove", liveOld)
	branch["decision"] = "merge"
	request["forge_receipt"] = map[string]any{"merged": true, "checks_passed": true}
	writePreparationJSON(t, requestPath, request)
	if r := runDisposition(t, script, repo, requestPath); r.Items[0].Disposition != "deliver-pr" || r.Items[0].Completed {
		t.Fatalf("forged receipt granted merge: %+v", r)
	}
	accept["merge"] = true
	writePreparationJSON(t, requestPath, request)
	if r := runDisposition(t, script, repo, requestPath); r.Items[0].Disposition != "merge-eligible" || !r.Items[0].ForgeChecks || r.Items[0].Completed {
		t.Fatalf("authorized attempt did not route fresh forge checks: %+v", r)
	}
	policyPath := filepath.Join(repo, "docs/context/operations/kb-routing.yaml")
	portableWrite(t, policyPath, "delivery:\n  mode: local\n  merge: manual\n")
	if r := runDisposition(t, script, repo, requestPath); r.Items[0].Disposition != "retained-blocked" {
		t.Fatalf("ignored local policy: %+v", r)
	}
	for _, ambiguous := range []string{"delivery:\n  mode: pr\n  mode: local\n", "delivery: {mode: local}\n", "  delivery:\n    mode: local\n", "delivery:\n  mode: pr\ndelivery:\n  mode: local\n"} {
		portableWrite(t, policyPath, ambiguous)
		if r := runDisposition(t, script, repo, requestPath); r.Items[0].Disposition != "retained-blocked" {
			t.Fatalf("ambiguous delivery policy authorized publishing: %+v", r)
		}
	}
	portableWrite(t, policyPath, "delivery:\n  mode: pr\n  merge: auto-after-checks\n")
	delete(accept, "merge")
	writePreparationJSON(t, requestPath, request)
	if r := runDisposition(t, script, repo, requestPath); r.Items[0].Disposition != "merge-eligible" || r.Items[0].Completed {
		t.Fatalf("explicit auto policy: %+v", r)
	}
	branch["decision"] = "reject"
	branch["rejected"] = true
	branch["artifacts"] = []map[string]string{{"path": "docs/plans/notes.md", "sha256": notesHash}, {"path": ".env", "sha256": "credential-excluded"}}
	delete(request, "forge_receipt")
	writePreparationJSON(t, requestPath, request)
	key := fmt.Sprintf("%x", sha256.Sum256([]byte("refs/heads/codex/prior|"+tip)))[:32]
	archive := filepath.Join(repo, ".git/.copilot-kb/recovery/dispositions/dispose-test", key+".archive")
	restore := filepath.Join(archive, "restore.git")
	portableGit(t, temp, "init", "--bare", restore)
	portableGit(t, restore, "fetch", repo, "refs/heads/trunk:refs/heads/codex/prior")
	blocked := runDisposition(t, script, repo, requestPath)
	if blocked.Items[0].Completed || portableGit(t, repo, "rev-parse", "codex/prior") != tip {
		t.Fatalf("failed restore allowed retirement: %+v", blocked)
	}
	portableGit(t, restore, "update-ref", "-d", "refs/heads/codex/prior", baseline)
	t.Run("InstalledNativeOwner", func(t *testing.T) {
		nativeDir := filepath.Join(temp, "native")
		portableWrite(t, filepath.Join(nativeDir, "kbreconcile.cmd"), "@exit /b 91\r\n")
		t.Setenv("KB_TEST_NATIVE_PATH", nativeDir)
		if r := runDisposition(t, script, repo, requestPath); r.Items[0].Reason != "native-owner-required-no-weaker-retry" {
			t.Fatalf("installed owner bypassed: %+v", r)
		}
	})
	sourceIndex := string(portableRead(t, filepath.Join(repo, ".git/index")))
	done := runDisposition(t, script, repo, requestPath)
	if !done.Items[0].Completed || done.Items[0].Disposition != "discard-confirmed-junk" {
		t.Fatalf("retirement: %+v", done)
	}
	if portableGit(t, restore, "rev-parse", "refs/heads/codex/prior") != tip || string(portableRead(t, filepath.Join(archive, "restored-artifacts/docs/plans/notes.md"))) != "uncommitted notes\r\n" {
		t.Fatal("archive not actually restored")
	}
	if string(portableRead(t, filepath.Join(repo, ".git/index"))) != sourceIndex || string(portableRead(t, filepath.Join(repo, "docs/plans/notes.md"))) != "uncommitted notes\r\n" || string(portableRead(t, filepath.Join(repo, ".env"))) != "fixture credential must stay\n" {
		t.Fatal("source/index/credential changed")
	}
	if _, e := os.Stat(filepath.Join(archive, "artifacts/.env")); !os.IsNotExist(e) {
		t.Fatal("credential archived")
	}
	if r := runDisposition(t, script, repo, requestPath); !r.Items[0].Completed || r.Items[0].Archive != done.Items[0].Archive {
		t.Fatalf("duplicate retirement: %+v", r)
	}
	receiptBytes := portableRead(t, done.Items[0].Receipt)
	var forged map[string]any
	json.Unmarshal(receiptBytes, &forged)
	outOfScope := filepath.Join(temp, "must-not-create")
	forged["archive"] = outOfScope
	writePreparationJSON(t, done.Items[0].Receipt, forged)
	if r := runDisposition(t, script, repo, requestPath); r.Items[0].Reason != "archive-receipt-identity-mismatch" {
		t.Fatalf("forged archive path accepted: %+v", r)
	}
	if _, err := os.Stat(outOfScope); !os.IsNotExist(err) {
		t.Fatal("forged receipt wrote outside archive")
	}
	portableWrite(t, done.Items[0].Receipt, string(receiptBytes))
	// Simulate interruption after CAS but before final receipt promotion.
	var receipt map[string]any
	json.Unmarshal(portableRead(t, done.Items[0].Receipt), &receipt)
	receipt["state"] = "archive-verified"
	writePreparationJSON(t, done.Items[0].Receipt, receipt)
	if r := runDisposition(t, script, repo, requestPath); !r.Items[0].Completed {
		t.Fatalf("post-CAS resume: %+v", r)
	}
	portableGit(t, repo, "update-ref", "refs/heads/codex/prior", baseline)
	if r := runDisposition(t, script, repo, requestPath); r.Items[0].Completed || portableGit(t, repo, "rev-parse", "codex/prior") != baseline {
		t.Fatalf("moved ref retired: %+v", r)
	}
	accept["revoked"] = true
	writePreparationJSON(t, requestPath, request)
	if r := runDisposition(t, script, repo, requestPath); r.Reason != "current-acceptance-required" {
		t.Fatalf("revoked replay: %+v", r)
	}
	delete(accept, "revoked")
	t.Run("GeneratedArtifact", func(t *testing.T) {
		artifactPath := ".kb/generated/output.tmp"
		portableWrite(t, filepath.Join(repo, artifactPath), "generated fixture\n")
		digest := fmt.Sprintf("%x", sha256.Sum256([]byte("generated fixture\n")))
		provenancePath := ".kb/generated/provenance.json"
		writePreparationJSON(t, filepath.Join(repo, provenancePath), map[string]any{"kind": "generated-output", "repository_id": survey.RepositoryID, "path": artifactPath, "sha256": digest, "producer": "fixture-generator"})
		provenanceHash := fmt.Sprintf("%x", sha256.Sum256(portableRead(t, filepath.Join(repo, provenancePath))))
		item := map[string]any{"kind": "artifact", "path": artifactPath, "sha256": digest, "decision": "discard-generated", "rejected": true, "provenance": map[string]string{"path": provenancePath, "sha256": provenanceHash}}
		request["run_id"] = "generated-test"
		request["items"] = []any{item}
		accept["items"] = []any{map[string]string{"kind": "artifact", "path": artifactPath, "sha256": digest}}
		writePreparationJSON(t, requestPath, request)
		queue := filepath.Join(repo, ".git/.copilot-kb/work-queue.json")
		portableWrite(t, queue, `[{"branch":"trunk","status":"active"}]`)
		if r := runDisposition(t, script, repo, requestPath); r.Items[0].Completed {
			t.Fatal("live generated artifact removed")
		}
		os.Remove(queue)
		r := runDisposition(t, script, repo, requestPath)
		if !r.Items[0].Completed {
			t.Fatalf("generated cleanup: %+v", r)
		}
		if _, e := os.Stat(filepath.Join(repo, artifactPath)); !os.IsNotExist(e) {
			t.Fatal("accepted generated artifact still exists")
		}
		if string(portableRead(t, filepath.Join(r.Items[0].Archive, "restored-content"))) != "generated fixture\n" {
			t.Fatal("generated output not recoverable")
		}
		if r := runDisposition(t, script, repo, requestPath); !r.Items[0].Completed {
			t.Fatalf("generated resume: %+v", r)
		}
	})
}

func TestPortableRecoveryDispositionPolicySections(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows PowerShell installed runtime")
	}
	root := t.TempDir()
	repo := filepath.Join(root, "consumer")
	remote := filepath.Join(root, "remote.git")
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatal(err)
	}
	portableGit(t, root, "init", "--bare", "--initial-branch=trunk", remote)
	portableGit(t, repo, "init", "--initial-branch=trunk")
	portableGit(t, repo, "config", "user.name", "Fixture")
	portableGit(t, repo, "config", "user.email", "fixture@example.invalid")
	portableWrite(t, filepath.Join(repo, "source.txt"), "baseline\n")
	portableGit(t, repo, "add", ".")
	portableGit(t, repo, "commit", "-m", "baseline")
	portableGit(t, repo, "remote", "add", "origin", remote)
	portableGit(t, repo, "push", "-u", "origin", "trunk")
	portableGit(t, repo, "checkout", "-b", "codex/backlog")
	portableWrite(t, filepath.Join(repo, "change.txt"), "backlog\n")
	portableGit(t, repo, "add", ".")
	portableGit(t, repo, "commit", "-m", "backlog")
	tip := portableGit(t, repo, "rev-parse", "HEAD")
	portableGit(t, repo, "checkout", "trunk")
	manifest := "ref: refs/heads/codex/backlog\ntip: " + tip + "\n"
	portableWrite(t, filepath.Join(repo, "docs/plans/owner.md"), manifest)
	script := installedRecoveryScript(t, "agents")
	survey := portableRunSurvey(t, script, repo)
	branch := map[string]any{"kind": "branch", "ref": "refs/heads/codex/backlog", "tip": tip, "decision": "deliver", "manifest": "docs/plans/owner.md", "manifest_sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(manifest)))}
	request := map[string]any{"schema_version": 1, "run_id": "policy-section", "objective": "deliver accepted backlog", "repository_id": survey.RepositoryID, "acceptance": map[string]any{"source": "current-user-reply", "accepted": true, "items": []any{map[string]any{"kind": "branch", "ref": "refs/heads/codex/backlog", "tip": tip}}}, "items": []any{branch}}
	path := filepath.Join(root, "request.json")
	writePreparationJSON(t, path, request)
	cases := []struct{ name, policy, want string }{
		{"blank", "delivery:\n  merge: manual\n\n  mode: local\n", "retained-blocked"},
		{"comment", "delivery:\n  merge: manual\n# delivery note\n  mode: local\n", "retained-blocked"},
		{"next-key", "delivery:\n  merge: manual\n\n# next section\nother:\n  mode: local\n", "deliver-pr"},
		{"duplicate-after-blank", "delivery:\n  mode: pr\n\n  mode: local\n", "retained-blocked"},
		{"canonical-local", "delivery:\n  mode: local\n  merge: manual\n", "retained-blocked"},
		{"flow-local", "delivery: {mode: local}\n", "retained-blocked"},
		{"indented-local", "  delivery:\n    mode: local\n", "retained-blocked"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			portableWrite(t, filepath.Join(repo, "docs/context/operations/kb-routing.yaml"), tc.policy)
			r := runDisposition(t, script, repo, path)
			if len(r.Items) != 1 || r.Items[0].Disposition != tc.want || r.Items[0].Completed {
				t.Fatalf("policy %q: want %s, got %+v", tc.policy, tc.want, r)
			}
		})
	}
}
