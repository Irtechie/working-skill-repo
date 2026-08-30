package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This file is the protected oracle for slice-001. It proves every lifecycle
// classification, every fail-closed downgrade, and that adapter absence removes
// conclusions rather than adding them.

type workRealityFixture struct {
	Root   string
	Origin string
}

func runGitForWorkReality(t *testing.T, dir string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	command.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=kb", "GIT_AUTHOR_EMAIL=kb@example.invalid",
		"GIT_COMMITTER_NAME=kb", "GIT_COMMITTER_EMAIL=kb@example.invalid",
		"GIT_TERMINAL_PROMPT=0",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s failed: %v\n%s", args, dir, err, output)
	}
	return strings.TrimSpace(string(output))
}

func writeWorkRealityFile(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func copyRehabPolicy(t *testing.T, root string) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("..", "..", "config", "rehab-policy.json"))
	if err != nil {
		t.Fatalf("read repo rehab policy: %v", err)
	}
	writeWorkRealityFile(t, root, "config/rehab-policy.json", string(source))
}

// newWorkRealityFixture builds a bare origin plus a clone so remote authority is
// genuinely resolvable through ls-remote and fetch, not through cached refs.
func newWorkRealityFixture(t *testing.T) workRealityFixture {
	t.Helper()
	base := t.TempDir()
	origin := filepath.Join(base, "origin.git")
	if err := os.MkdirAll(origin, 0o755); err != nil {
		t.Fatalf("mkdir origin: %v", err)
	}
	runGitForWorkReality(t, origin, "init", "--bare", "--quiet")
	runGitForWorkReality(t, base, "--git-dir="+origin, "symbolic-ref", "HEAD", "refs/heads/main")

	root := filepath.Join(base, "work")
	runGitForWorkReality(t, base, "clone", "--quiet", origin, root)
	runGitForWorkReality(t, root, "config", "user.name", "kb")
	runGitForWorkReality(t, root, "config", "user.email", "kb@example.invalid")

	writeWorkRealityFile(t, root, "README.md", "# fixture\n")
	copyRehabPolicy(t, root)
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "initial")
	runGitForWorkReality(t, root, "branch", "-M", "main")
	runGitForWorkReality(t, root, "push", "--quiet", "-u", "origin", "main")

	return workRealityFixture{Root: root, Origin: origin}
}

func writeWorkRealityQueue(t *testing.T, root string, claims []map[string]any) {
	t.Helper()
	common := filepath.Join(root, ".git", ".copilot-kb")
	if err := os.MkdirAll(common, 0o755); err != nil {
		t.Fatalf("mkdir queue dir: %v", err)
	}
	encoded, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal queue: %v", err)
	}
	if err := os.WriteFile(filepath.Join(common, "work-queue.json"), encoded, 0o600); err != nil {
		t.Fatalf("write queue: %v", err)
	}
}

func reportPairing(t *testing.T, report workRealityReport, id string) workRealityPairing {
	t.Helper()
	for _, pairing := range report.Pairings {
		if pairing.ID == id {
			return pairing
		}
	}
	ids := []string{}
	for _, pairing := range report.Pairings {
		ids = append(ids, pairing.ID)
	}
	t.Fatalf("pairing %q not found; have %v", id, ids)
	return workRealityPairing{}
}

func runWorkRealityFixture(t *testing.T, root string) workRealityReport {
	t.Helper()
	report, err := executeWorkReality(workRealityOptions{
		Root:      root,
		SessionID: "fixture-session",
		Cutoff:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("executeWorkReality: %v", err)
	}
	return report
}

func TestWorkRealityPreservationPredicatesSupersetOfTerminalCleanup(t *testing.T) {
	preserved := map[string]bool{}
	for _, name := range workRealityPreservationPredicates() {
		preserved[name] = true
	}
	for _, name := range terminalCleanupSafetyPredicates() {
		if !preserved[name] {
			t.Fatalf("work-reality preservation set is missing terminal-cleanup predicate %q", name)
		}
	}
	if len(preserved) <= len(terminalCleanupSafetyPredicates()) {
		t.Fatalf("work-reality preservation set must exceed the terminal-cleanup floor")
	}
}

func TestWorkRealityClassifiesSquashMergedBranchAsDead(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/squashed")
	writeWorkRealityFile(t, root, "feature.txt", "landed\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "add feature")

	runGitForWorkReality(t, root, "checkout", "--quiet", "main")
	runGitForWorkReality(t, root, "cherry-pick", "codex/squashed")
	runGitForWorkReality(t, root, "push", "--quiet", "origin", "main")

	report := runWorkRealityFixture(t, root)
	if report.Status != workRealityStatusOK {
		t.Fatalf("expected ok status, got %s (%v)", report.Status, report.Limitations)
	}
	pairing := reportPairing(t, report, "branch:codex/squashed")
	if pairing.State != workRealityStateDead {
		t.Fatalf("patch-equivalent branch must be dead, got %s: %s", pairing.State, pairing.Reason)
	}
	if pairing.Contained != "true" {
		t.Fatalf("expected containment proof, got %q", pairing.Contained)
	}
}

func TestWorkRealityClassifiesUncontainedBranches(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/declared")
	writeWorkRealityFile(t, root, "declared.txt", "declared\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "declared work")

	runGitForWorkReality(t, root, "checkout", "--quiet", "main")
	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/undeclared")
	writeWorkRealityFile(t, root, "undeclared.txt", "undeclared\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "undeclared work")

	runGitForWorkReality(t, root, "checkout", "--quiet", "main")
	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Active Work",
		"",
		"| Item | Branch |",
		"|---|---|",
		"| Declared thing | `codex/declared` |",
		"",
	}, "\n"))

	report := runWorkRealityFixture(t, root)
	declared := reportPairing(t, report, "branch:codex/declared")
	if declared.State != workRealityStateUnshipped {
		t.Fatalf("uncontained declared branch must be unshipped, got %s: %s", declared.State, declared.Reason)
	}
	undeclared := reportPairing(t, report, "branch:codex/undeclared")
	if undeclared.State != workRealityStateOrphanBranch {
		t.Fatalf("uncontained undeclared branch must be orphan-branch, got %s: %s", undeclared.State, undeclared.Reason)
	}
}

func TestWorkRealityMissingManifestIsOrphanWorkNotDead(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Active Work",
		"",
		"- Ghost work tracked in docs/plans/2099-01-01-000-absent-manifest.md",
		"",
	}, "\n"))

	report := runWorkRealityFixture(t, root)
	pairing := reportPairing(t, report, "work:todo-001")
	if pairing.State != workRealityStateOrphanWork {
		t.Fatalf("missing manifest must be orphan-work, got %s: %s", pairing.State, pairing.Reason)
	}
	for _, other := range report.Pairings {
		if other.State == workRealityStateDead || other.State == workRealityStateSuperseded {
			t.Fatalf("a missing file must never produce %s", other.State)
		}
	}
}

func TestWorkRealitySupersessionIsNeverSelfProving(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	// The replacement exists only on the candidate branch, together with the
	// claim that it supersedes the branch. That must never prove itself.
	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/self-proving")
	writeWorkRealityFile(t, root, "docs/plans/replacement.md", "# replacement\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "claim own supersession")
	runGitForWorkReality(t, root, "checkout", "--quiet", "main")

	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Active Work",
		"",
		"- Old thing on `codex/self-proving` superseded by docs/plans/replacement.md",
		"",
	}, "\n"))

	report := runWorkRealityFixture(t, root)
	pairing := reportPairing(t, report, "branch:codex/self-proving")
	if pairing.State != workRealityStateHumanRequired {
		t.Fatalf("self-proving supersession must be human-required, got %s: %s", pairing.State, pairing.Reason)
	}
}

func TestWorkRealitySupersessionWithUncontainedCommitsIsHumanRequired(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	writeWorkRealityFile(t, root, "docs/plans/replacement.md", "# replacement\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "land replacement on default")
	runGitForWorkReality(t, root, "push", "--quiet", "origin", "main")

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/superseded")
	writeWorkRealityFile(t, root, "unique.txt", "still unique\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "unique work")

	runGitForWorkReality(t, root, "checkout", "--quiet", "main")
	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Parked",
		"",
		"- Old thing on `codex/superseded` superseded by docs/plans/replacement.md",
		"",
	}, "\n"))

	report := runWorkRealityFixture(t, root)
	pairing := reportPairing(t, report, "branch:codex/superseded")
	if pairing.State != workRealityStateHumanRequired {
		t.Fatalf("uncontained supersession must be human-required, got %s: %s", pairing.State, pairing.Reason)
	}
}

func TestWorkRealitySupersessionWithContainmentAndLandedReplacement(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/replaced")
	writeWorkRealityFile(t, root, "old.txt", "old\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "old work")

	runGitForWorkReality(t, root, "checkout", "--quiet", "main")
	runGitForWorkReality(t, root, "cherry-pick", "codex/replaced")
	writeWorkRealityFile(t, root, "docs/plans/replacement.md", "# replacement\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "land replacement")
	runGitForWorkReality(t, root, "push", "--quiet", "origin", "main")

	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Parked",
		"",
		"- Old thing on `codex/replaced` superseded by docs/plans/replacement.md",
		"",
	}, "\n"))

	report := runWorkRealityFixture(t, root)
	pairing := reportPairing(t, report, "branch:codex/replaced")
	if pairing.State != workRealityStateSuperseded {
		t.Fatalf("contained supersession with a landed replacement must be superseded, got %s: %s",
			pairing.State, pairing.Reason)
	}
}

func TestWorkRealityFreshClaimIsLiveAndStaleClaimIsHumanRequired(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	for _, branch := range []string{"codex/fresh-claim", "codex/stale-claim"} {
		runGitForWorkReality(t, root, "checkout", "--quiet", "-b", branch)
		writeWorkRealityFile(t, root, strings.ReplaceAll(branch, "/", "-")+".txt", branch+"\n")
		runGitForWorkReality(t, root, "add", "-A")
		runGitForWorkReality(t, root, "commit", "--quiet", "-m", branch)
		runGitForWorkReality(t, root, "checkout", "--quiet", "main")
	}

	now := time.Now().UTC()
	writeWorkRealityQueue(t, root, []map[string]any{
		{
			"work_id": "fresh", "status": "in_progress", "session_id": "peer-a",
			"branch": "codex/fresh-claim", "updated_at": now.Add(-6 * time.Hour).Format(time.RFC3339),
		},
		{
			"work_id": "stale", "status": "in_progress", "session_id": "peer-b",
			"branch": "codex/stale-claim", "updated_at": now.Add(-96 * time.Hour).Format(time.RFC3339),
		},
	})

	report := runWorkRealityFixture(t, root)
	fresh := reportPairing(t, report, "branch:codex/fresh-claim")
	if fresh.State != workRealityStateLive {
		t.Fatalf("a six-hour-old non-terminal claim must stay live, got %s: %s", fresh.State, fresh.Reason)
	}
	stale := reportPairing(t, report, "branch:codex/stale-claim")
	if stale.State != workRealityStateHumanRequired {
		t.Fatalf("a stale claim must be human-required, never takeover authority, got %s: %s",
			stale.State, stale.Reason)
	}
}

func TestWorkRealityFailsClosedWithoutRemote(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/no-remote")
	writeWorkRealityFile(t, root, "orphan.txt", "orphan\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "work")
	runGitForWorkReality(t, root, "checkout", "--quiet", "main")
	runGitForWorkReality(t, root, "remote", "remove", "origin")

	report := runWorkRealityFixture(t, root)
	if report.Status != workRealityStatusFailClosed {
		t.Fatalf("absent remote must fail closed, got %s", report.Status)
	}
	for _, pairing := range report.Pairings {
		if pairing.State == workRealityStateDead || pairing.State == workRealityStateSuperseded {
			t.Fatalf("no remote authority must yield zero terminal states, got %s for %s",
				pairing.State, pairing.ID)
		}
	}
	pairing := reportPairing(t, report, "branch:codex/no-remote")
	if pairing.State != workRealityStateHumanRequired {
		t.Fatalf("unprovable containment must downgrade to human-required, got %s", pairing.State)
	}
}

func TestWorkRealityFailsClosedWhenRemoteIsUnreachable(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/unreachable")
	writeWorkRealityFile(t, root, "work.txt", "work\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "work")
	runGitForWorkReality(t, root, "checkout", "--quiet", "main")

	if err := os.RemoveAll(fixture.Origin); err != nil {
		t.Fatalf("remove origin: %v", err)
	}

	report := runWorkRealityFixture(t, root)
	if report.Status != workRealityStatusFailClosed {
		t.Fatalf("unreachable remote must fail closed, got %s", report.Status)
	}
	if report.RemoteAuthority.State == "authoritative" {
		t.Fatalf("an unreachable remote must never be authoritative")
	}
	for _, pairing := range report.Pairings {
		if pairing.State == workRealityStateDead || pairing.State == workRealityStateSuperseded {
			t.Fatalf("unreachable remote must yield zero terminal states, got %s", pairing.State)
		}
	}
}

func TestWorkRealityFailsClosedWhenDefaultIsRewritten(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	// Point the local remote at a bare repository whose advertised HEAD cannot be
	// fetched, standing in for a default rewritten between advertisement and fetch.
	broken := filepath.Join(t.TempDir(), "broken.git")
	if err := os.MkdirAll(broken, 0o755); err != nil {
		t.Fatalf("mkdir broken: %v", err)
	}
	runGitForWorkReality(t, broken, "init", "--bare", "--quiet")
	runGitForWorkReality(t, root, "--git-dir="+broken, "symbolic-ref", "HEAD", "refs/heads/main")
	runGitForWorkReality(t, root, "remote", "set-url", "origin", broken)

	report := runWorkRealityFixture(t, root)
	if report.Status != workRealityStatusFailClosed {
		t.Fatalf("an unresolvable advertised default must fail closed, got %s", report.Status)
	}
	if report.RemoteAuthority.Limitation == "" {
		t.Fatalf("a fail-closed remote authority must carry a limitation")
	}
}

func TestWorkRealityIsReadOnly(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/read-only")
	writeWorkRealityFile(t, root, "thing.txt", "thing\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "thing")
	runGitForWorkReality(t, root, "checkout", "--quiet", "main")

	before := map[string]string{
		"refs":      runGitForWorkReality(t, root, "for-each-ref", "--format=%(refname) %(objectname)"),
		"status":    runGitForWorkReality(t, root, "status", "--porcelain"),
		"worktrees": runGitForWorkReality(t, root, "worktree", "list", "--porcelain"),
		"head":      runGitForWorkReality(t, root, "rev-parse", "HEAD"),
	}

	runWorkRealityFixture(t, root)

	after := map[string]string{
		"refs":      runGitForWorkReality(t, root, "for-each-ref", "--format=%(refname) %(objectname)"),
		"status":    runGitForWorkReality(t, root, "status", "--porcelain"),
		"worktrees": runGitForWorkReality(t, root, "worktree", "list", "--porcelain"),
		"head":      runGitForWorkReality(t, root, "rev-parse", "HEAD"),
	}
	for key, value := range before {
		if after[key] != value {
			t.Fatalf("work-reality mutated %s\nbefore:\n%s\nafter:\n%s", key, value, after[key])
		}
	}
}

func TestWorkRealityReportIsDeterministic(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/stable")
	writeWorkRealityFile(t, root, "stable.txt", "stable\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "stable")
	runGitForWorkReality(t, root, "checkout", "--quiet", "main")

	cutoff := time.Now().UTC()
	first, err := executeWorkReality(workRealityOptions{Root: root, SessionID: "s", Cutoff: cutoff})
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	second, err := executeWorkReality(workRealityOptions{Root: root, SessionID: "s", Cutoff: cutoff})
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	firstJSON, err := marshalWorkReality(first)
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	secondJSON, err := marshalWorkReality(second)
	if err != nil {
		t.Fatalf("marshal second: %v", err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("report is not byte-identical across runs")
	}
	if first.Cutoff != cutoff.Format(time.RFC3339) {
		t.Fatalf("cutoff must be immutable, got %s", first.Cutoff)
	}
}

func TestWorkRealityRedactsCredentialLikeDetail(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Active Work",
		"",
		"- Rotate deploy_secret_token for staging",
		"",
	}, "\n"))

	report := runWorkRealityFixture(t, root)
	encoded, err := marshalWorkReality(report)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(encoded), "deploy_secret_token") {
		t.Fatalf("credential-like detail must be redacted from the report")
	}
	if !strings.Contains(string(encoded), workRealityRedacted) {
		t.Fatalf("expected a redaction marker in the report")
	}
}

func TestWorkRealityPreservesCurrentBranchUnconditionally(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/current-session")
	writeWorkRealityFile(t, root, "current.txt", "current\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "current work")
	runGitForWorkReality(t, root, "push", "--quiet", "origin", "HEAD:refs/heads/landed")
	runGitForWorkReality(t, root, "checkout", "--quiet", "main")
	runGitForWorkReality(t, root, "cherry-pick", "codex/current-session")
	runGitForWorkReality(t, root, "push", "--quiet", "origin", "main")
	runGitForWorkReality(t, root, "checkout", "--quiet", "codex/current-session")

	report := runWorkRealityFixture(t, root)
	pairing := reportPairing(t, report, "branch:codex/current-session")
	if pairing.State != workRealityStateLive {
		t.Fatalf("the current branch must be preserved even when contained, got %s: %s",
			pairing.State, pairing.Reason)
	}
}

func TestWorkRealityFlagsProtectedPathsOnUnshippedWork(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/protected")
	writeWorkRealityFile(t, root, ".github/skills/example/SKILL.md", "# example\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "touch a protected root")
	runGitForWorkReality(t, root, "checkout", "--quiet", "main")

	report := runWorkRealityFixture(t, root)
	pairing := reportPairing(t, report, "branch:codex/protected")
	if pairing.State != workRealityStateOrphanBranch {
		t.Fatalf("expected orphan-branch, got %s", pairing.State)
	}
	found := false
	for _, protected := range pairing.ProtectedPaths {
		if protected == ".github/skills" {
			found = true
		}
	}
	if !found {
		t.Fatalf("protected root .github/skills must be reported, got %v", pairing.ProtectedPaths)
	}
}

func TestWorkRealityFailsClosedWithoutPredicateManifest(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "rm", "--quiet", "config/rehab-policy.json")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "drop policy")

	report := runWorkRealityFixture(t, root)
	if report.Status != workRealityStatusFailClosed {
		t.Fatalf("a missing predicate manifest must fail closed, got %s", report.Status)
	}
	for _, pairing := range report.Pairings {
		if pairing.State == workRealityStateDead || pairing.State == workRealityStateSuperseded {
			t.Fatalf("a missing predicate manifest must yield zero terminal states")
		}
	}
}

func TestWorkRealityUnparsedRowBecomesOrphanWorkNotDropped(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	// A marker and status glyph with no title behind them. "- ???" is readable
	// work with an odd name, so it no longer belongs here.
	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Blocked",
		"",
		"- ⬜",
		"",
	}, "\n"))

	report := runWorkRealityFixture(t, root)
	if len(report.Declared) == 0 {
		t.Fatalf("an unreadable row must still appear as declared work")
	}
	if report.Declared[0].Parsed {
		t.Fatalf("a row with no title must not be reported as parsed: %#v", report.Declared[0])
	}
	pairing := reportPairing(t, report, "work:todo-001")
	if pairing.State != workRealityStateOrphanWork {
		t.Fatalf("an unparsed row must be orphan-work, got %s: %s", pairing.State, pairing.Reason)
	}
}

// TestWorkRealityReadsRowStatusSoMarkedWorkSettles pins the loop-breaker: the
// mark action writes "⊘ skipped" as its terminal marker, so a run that cannot
// read that marker back re-surfaces the identical row as outstanding work on
// every later pass and the lane can never settle anything it marked.
func TestWorkRealityReadsRowStatusSoMarkedWorkSettles(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Active Work",
		"",
		"| Workstream | Status | Priority | Link |",
		"|---|---|---|---|",
		"| Already marked terminal | ⊘ skipped | P0 | superseded earlier |",
		"| Still outstanding | 🔧 in_progress | P0 | ongoing |",
		"",
		"## Queued Improvements",
		"",
		"- ⊘ bullet that was already marked",
		"- ⬜ bullet that is still pending",
		"",
	}, "\n"))

	report := runWorkRealityFixture(t, root)

	byTitle := map[string]string{}
	for _, item := range report.Declared {
		byTitle[item.Title] = item.Status
	}
	for title, want := range map[string]string{
		"Already marked terminal":        "skipped",
		"Still outstanding":              "in_progress",
		"bullet that was already marked": "skipped",
		"bullet that is still pending":   "pending",
	} {
		if got := byTitle[title]; got != want {
			t.Fatalf("status for %q = %q, want %q (declared: %#v)", title, got, want, report.Declared)
		}
	}

	settled := map[string]bool{}
	for _, item := range report.Settled {
		settled[item.ID] = true
	}
	if len(settled) != 2 {
		t.Fatalf("both skipped rows must settle, got %d: %#v", len(settled), report.Settled)
	}
	for _, pairing := range report.Pairings {
		if settled[strings.TrimPrefix(pairing.ID, "work:")] {
			t.Fatalf("a settled row must not be paired again as outstanding work: %s", pairing.ID)
		}
	}
}

// TestWorkRealityReadsListRowsAndDropsTableHeader pins three reporting defects
// that together made every bullet in todo.md untriageable: the table header was
// counted as work, bullet titles kept their markup, and the table-shaped
// cell-count test marked every readable bullet unparsed.
func TestWorkRealityReadsListRowsAndDropsTableHeader(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Active Work",
		"",
		"| Workstream | Status | Link |",
		"|---|---|---|",
		"| Real table work | in_progress | none |",
		"",
		"## Queued Improvements",
		"",
		"- ⬜ Real bullet work that names no ref",
		"",
	}, "\n"))

	report := runWorkRealityFixture(t, root)
	titles := []string{}
	for _, item := range report.Declared {
		if strings.HasPrefix(item.ID, "todo-") {
			titles = append(titles, item.Title)
		}
	}
	if len(titles) != 2 {
		t.Fatalf("expected the header dropped and 2 work rows kept, got %d: %#v", len(titles), titles)
	}
	if titles[0] != "Real table work" {
		t.Fatalf("table header must not be counted as work, got %q", titles[0])
	}
	if titles[1] != "Real bullet work that names no ref" {
		t.Fatalf("bullet title must lose its marker and glyph, got %q", titles[1])
	}

	pairing := reportPairing(t, report, "work:todo-002")
	if pairing.State != workRealityStateOrphanWork {
		t.Fatalf("an unpaired bullet stays orphan-work, got %s", pairing.State)
	}
	if strings.Contains(pairing.Reason, "could not be parsed") {
		t.Fatalf("a readable bullet must not claim a parse failure: %s", pairing.Reason)
	}
}

func TestWorkRealitySettledManifestIsNotOrphanWork(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	writeWorkRealityFile(t, root, "docs/plans/2020-01-01-000-finished-manifest.md", strings.Join([]string{
		"---",
		"kb_id: kb-finished",
		"status: complete",
		"---",
		"",
		"# finished",
		"",
	}, "\n"))
	writeWorkRealityFile(t, root, "docs/plans/2020-01-02-000-active-manifest.md", strings.Join([]string{
		"---",
		"kb_id: kb-active",
		"status: reviewed",
		"---",
		"",
		"# active",
		"",
	}, "\n"))

	report := runWorkRealityFixture(t, root)

	settled := map[string]string{}
	for _, item := range report.Settled {
		settled[item.ID] = item.Status
	}
	finishedID := "manifest:docs/plans/2020-01-01-000-finished-manifest.md"
	if settled[finishedID] != "complete" {
		t.Fatalf("a manifest with a terminal declared status must be reported settled, got %v", report.Settled)
	}
	for _, pairing := range report.Pairings {
		if pairing.DeclaredID == finishedID {
			t.Fatalf("settled work must not be paired as outstanding, got %s", pairing.State)
		}
	}

	activeID := "manifest:docs/plans/2020-01-02-000-active-manifest.md"
	foundActive := false
	for _, pairing := range report.Pairings {
		if pairing.DeclaredID == activeID {
			foundActive = true
			if pairing.State != workRealityStateOrphanWork {
				t.Fatalf("an unpaired non-terminal manifest must be orphan-work, got %s", pairing.State)
			}
		}
	}
	if !foundActive {
		t.Fatalf("a non-terminal manifest must still be paired as outstanding work")
	}
}

// ---------------------------------------------------------------------------
// Slice-002: markers, removal gate, and the decision packet
// ---------------------------------------------------------------------------

func runWorkRealityAction(t *testing.T, root, action string) workRealityReport {
	t.Helper()
	report, err := executeWorkReality(workRealityOptions{
		Root:      root,
		SessionID: "fixture-session",
		Action:    action,
		Cutoff:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("executeWorkReality(%s): %v", action, err)
	}
	return report
}

func readTodo(t *testing.T, root string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, "todo.md"))
	if err != nil {
		t.Fatalf("read todo.md: %v", err)
	}
	return string(raw)
}

// deadDeclaredFixture lands a branch by cherry-pick and declares it in todo.md
// with a real status marker, so the pairing is dead with containment proof.
func deadDeclaredFixture(t *testing.T, marker string) string {
	t.Helper()
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/landed")
	writeWorkRealityFile(t, root, "landed.txt", "landed\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "landed work")

	runGitForWorkReality(t, root, "checkout", "--quiet", "main")
	runGitForWorkReality(t, root, "cherry-pick", "codex/landed")
	runGitForWorkReality(t, root, "push", "--quiet", "origin", "main")

	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Active Work",
		"",
		"| Item | Status | Link |",
		"|---|---|---|",
		"| Landed thing | " + marker + " | `codex/landed` |",
		"",
	}, "\n"))
	return root
}

func TestRehabMarkWritesSkippedMarkerForProvenDeadWork(t *testing.T) {
	root := deadDeclaredFixture(t, "\U0001F527 in_progress")

	report := runWorkRealityAction(t, root, workRealityActionMark)
	if report.Marks == nil || !report.Marks.Applied {
		t.Fatalf("expected applied marks, got %+v", report.Marks)
	}
	body := readTodo(t, root)
	if !strings.Contains(body, rehabMarkerSkipped) {
		t.Fatalf("expected the skipped marker in todo.md, got:\n%s", body)
	}
	if strings.Contains(body, "\U0001F527 in_progress") {
		t.Fatalf("the stale marker must be replaced, got:\n%s", body)
	}
	if !strings.Contains(body, "Landed thing") {
		t.Fatalf("marking must never remove the row, got:\n%s", body)
	}
	if !strings.Contains(body, "kb-rehab: marked dead") {
		t.Fatalf("every write must carry its evidence, got:\n%s", body)
	}
}

func TestRehabMarkPreservesAmbiguousWorkWithZeroWrites(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/open")
	writeWorkRealityFile(t, root, "open.txt", "open\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "open work")
	runGitForWorkReality(t, root, "checkout", "--quiet", "main")

	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Active Work",
		"",
		"| Item | Status | Link |",
		"|---|---|---|",
		"| Open thing | \U0001F527 in_progress | `codex/open` |",
		"",
	}, "\n"))
	before := readTodo(t, root)

	report := runWorkRealityAction(t, root, workRealityActionMark)
	if report.Marks == nil {
		t.Fatalf("mark action must report a mark result")
	}
	if len(report.Marks.Writes) != 0 {
		t.Fatalf("ambiguous work must produce zero writes, got %+v", report.Marks.Writes)
	}
	if after := readTodo(t, root); after != before {
		t.Fatalf("todo.md must be byte-identical after a preserving run:\n%s", after)
	}
}

func TestRehabMarkRefusesEntirelyOnFailClosedReport(t *testing.T) {
	root := deadDeclaredFixture(t, "\U0001F527 in_progress")
	before := readTodo(t, root)
	if err := os.Remove(filepath.Join(root, "config", "rehab-policy.json")); err != nil {
		t.Fatalf("remove policy: %v", err)
	}

	report := runWorkRealityAction(t, root, workRealityActionMark)
	if report.Status != workRealityStatusFailClosed {
		t.Fatalf("expected fail-closed, got %s", report.Status)
	}
	if report.Marks == nil || report.Marks.Applied || report.Marks.Refused == "" {
		t.Fatalf("a fail-closed report must refuse every write, got %+v", report.Marks)
	}
	if after := readTodo(t, root); after != before {
		t.Fatalf("fail-closed must leave todo.md untouched:\n%s", after)
	}
}

func TestRehabRemovalBlockedByUncontainedCommitsRemarksTheRow(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/unshipped")
	writeWorkRealityFile(t, root, "unshipped.txt", "unshipped\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "unshipped work")
	runGitForWorkReality(t, root, "checkout", "--quiet", "main")

	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Active Work",
		"",
		"| Item | Status | Link |",
		"|---|---|---|",
		"| Unshipped thing | \U0001F527 in_progress | `codex/unshipped` |",
		"",
	}, "\n"))

	report := runWorkRealityAction(t, root, workRealityActionRemove)
	pairing := reportPairing(t, report, "branch:codex/unshipped")
	if pairing.Contained == "true" {
		t.Fatalf("fixture must hold uncontained commits, got contained=%s", pairing.Contained)
	}
	body := readTodo(t, root)
	if !strings.Contains(body, "Unshipped thing") {
		t.Fatalf("an uncontained commit must block removal, got:\n%s", body)
	}
	if !strings.Contains(body, rehabMarkerBlocked) {
		t.Fatalf("a blocked removal must re-mark the row, got:\n%s", body)
	}
	if !strings.Contains(body, "removal blocked") {
		t.Fatalf("the block must name its reason, got:\n%s", body)
	}
}

func TestRehabRemovalPermittedWhenArtifactLandedAndRefContained(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/replaced")
	writeWorkRealityFile(t, root, "old.txt", "old\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "old work")

	runGitForWorkReality(t, root, "checkout", "--quiet", "main")
	runGitForWorkReality(t, root, "cherry-pick", "codex/replaced")
	writeWorkRealityFile(t, root, "docs/plans/replacement.md", "# replacement\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "land replacement")
	runGitForWorkReality(t, root, "push", "--quiet", "origin", "main")

	writeWorkRealityFile(t, root, "todo.md", strings.Join([]string{
		"# todo",
		"",
		"## Parked",
		"",
		"| Item | Status | Link |",
		"|---|---|---|",
		"| Old thing | \U0001F527 in_progress | `codex/replaced` superseded by docs/plans/replacement.md |",
		"",
	}, "\n"))

	report := runWorkRealityAction(t, root, workRealityActionRemove)
	pairing := reportPairing(t, report, "branch:codex/replaced")
	if pairing.State != workRealityStateSuperseded {
		t.Fatalf("expected superseded, got %s: %s", pairing.State, pairing.Reason)
	}
	body := readTodo(t, root)
	if strings.Contains(body, "Old thing") {
		t.Fatalf("a contained supersession with a landed artifact permits removal, got:\n%s", body)
	}
	if report.Marks == nil || len(report.Marks.Writes) != 1 || report.Marks.Writes[0].Operation != "remove" {
		t.Fatalf("expected exactly one recorded removal, got %+v", report.Marks)
	}
}

func TestRehabMarkLeavesRowsWithNoStatusMarkerUntouched(t *testing.T) {
	root := deadDeclaredFixture(t, "no-marker")
	before := readTodo(t, root)

	report := runWorkRealityAction(t, root, workRealityActionMark)
	if report.Marks == nil || len(report.Marks.Blocked) == 0 {
		t.Fatalf("an unmarked row must be recorded as blocked, got %+v", report.Marks)
	}
	if after := readTodo(t, root); after != before {
		t.Fatalf("an unproven row shape must not be rewritten:\n%s", after)
	}
}

func syntheticAmbiguousPairings() []workRealityPairing {
	return []workRealityPairing{
		{ID: "b1", State: workRealityStateOrphanBranch, Reason: "no declaration", Contained: "false",
			ProtectedPaths: []string{".github/skills"}},
		{ID: "b2", State: workRealityStateOrphanBranch, Reason: "no declaration", Contained: "false",
			ProtectedPaths: []string{"cmd"}},
		{ID: "b3", State: workRealityStateOrphanBranch, Reason: "no declaration", Contained: "false"},
		{ID: "b4", State: workRealityStateUnshipped, Reason: "uncontained", Contained: "false"},
		{ID: "b5", State: workRealityStateUnshipped, Reason: "uncontained", Contained: "false",
			ProtectedPaths: []string{"cmd"}},
		{ID: "b6", State: workRealityStateOrphanWork, Reason: "no ref", Contained: "unknown"},
		{ID: "b7", State: workRealityStateHumanRequired, Reason: "stale claim", Contained: "unknown"},
		{ID: "b8", State: workRealityStateLive, Reason: "fresh claim", Contained: "unknown"},
		{ID: "b9", State: workRealityStateDead, Reason: "contained", Contained: "true"},
	}
}

func TestRehabPacketNeverExceedsFiveGroupedItemsAndDropsNothing(t *testing.T) {
	policy := workRealityPolicy{ProtectedPaths: []string{".github/skills", "cmd"}}
	packet := buildRehabPacket(syntheticAmbiguousPairings(), policy)
	if len(packet) != rehabPacketMaxItems {
		t.Fatalf("expected exactly %d grouped items, got %d", rehabPacketMaxItems, len(packet))
	}

	covered := 0
	for _, item := range packet {
		covered += len(item.PairingIDs) + item.OmittedPairings
		if item.RecommendedChoice == "" || item.Uncertainty == "" || item.SafeDefault == "" ||
			item.IrreversibleConsequence == "" || item.RecheckSensor == "" ||
			len(item.Evidence) == 0 {
			t.Fatalf("packet item %s omits a mandated field: %+v", item.ID, item)
		}
		if item.State == workRealityStateDead {
			t.Fatalf("a proven terminal pairing is not a decision: %+v", item)
		}
	}
	if covered != 8 {
		t.Fatalf("packet must account for all 8 ambiguous pairings, covered %d", covered)
	}
}

func TestRehabPacketNamesGlobalInstallRootsForSkillPaths(t *testing.T) {
	policy := workRealityPolicy{ProtectedPaths: []string{".github/skills", "cmd"}}
	packet := buildRehabPacket(syntheticAmbiguousPairings(), policy)

	found := false
	for _, item := range packet {
		if !strings.Contains(item.IrreversibleConsequence, ".github/skills") {
			continue
		}
		found = true
		for _, target := range rehabGlobalInstallRoots() {
			if !strings.Contains(item.IrreversibleConsequence, target) {
				t.Fatalf("a skills merge must name %s as a propagation target, got %q", target, item.IrreversibleConsequence)
			}
		}
	}
	if !found {
		t.Fatalf("expected a packet item scoped to .github/skills")
	}
}

func TestRehabReportActionDefaultsToReadOnly(t *testing.T) {
	root := deadDeclaredFixture(t, "\U0001F527 in_progress")
	before := readTodo(t, root)

	report := runWorkRealityFixture(t, root)
	if report.Action != workRealityActionReport {
		t.Fatalf("default action must be report, got %q", report.Action)
	}
	if report.Marks != nil {
		t.Fatalf("a report run must never carry marks, got %+v", report.Marks)
	}
	if after := readTodo(t, root); after != before {
		t.Fatalf("the default action must not write:\n%s", after)
	}
}

// TestWorkRealityReadsGoalStatusAndIgnoresPlaceholders covers the lifecycle
// directories. parseDirDeclaredWork used to list filenames without opening
// them, so a goal declaring status: complete was reported outstanding forever
// and a .gitkeep placeholder was counted as work nobody declared.
func TestWorkRealitySurveysRemoteOnlyBranchesExactlyOnce(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	root := fixture.Root

	// A branch pushed to origin and then deleted locally exists only as
	// refs/remotes/origin/*. Surveying refs/heads alone makes it invisible.
	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/remote-only")
	writeWorkRealityFile(t, root, "remote-only.txt", "remote only\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "remote only work")
	runGitForWorkReality(t, root, "push", "--quiet", "origin", "codex/remote-only")
	runGitForWorkReality(t, root, "checkout", "--quiet", "main")
	runGitForWorkReality(t, root, "branch", "-D", "codex/remote-only")

	// A branch that exists both locally and on the remote must not be counted
	// twice, because the local ref already represents it.
	runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/mirrored")
	writeWorkRealityFile(t, root, "mirrored.txt", "mirrored\n")
	runGitForWorkReality(t, root, "add", "-A")
	runGitForWorkReality(t, root, "commit", "--quiet", "-m", "mirrored work")
	runGitForWorkReality(t, root, "push", "--quiet", "origin", "codex/mirrored")
	runGitForWorkReality(t, root, "checkout", "--quiet", "main")

	report := runWorkRealityFixture(t, root)

	remoteOnly := 0
	mirrored := 0
	for _, pairing := range report.Pairings {
		switch pairing.Ref {
		case "refs/remotes/origin/codex/remote-only":
			remoteOnly++
		}
		if strings.HasSuffix(pairing.Ref, "codex/mirrored") {
			mirrored++
		}
		if strings.HasSuffix(pairing.Ref, "/HEAD") {
			t.Fatalf("origin/HEAD is a symbolic pointer, not work: %s", pairing.Ref)
		}
	}

	if remoteOnly != 1 {
		t.Fatalf("remote-only branch must be surveyed exactly once, got %d pairings", remoteOnly)
	}
	if mirrored != 1 {
		t.Fatalf("mirrored branch must be surveyed once via its local ref, got %d pairings", mirrored)
	}
}

func TestWorkRealityReadsGoalStatusAndIgnoresPlaceholders(t *testing.T) {
	fixture := newWorkRealityFixture(t)
	// The real files in this repository use a capital-S "Status:" line under the
	// H1, not YAML frontmatter. A fixture written in the shape the code expects
	// would pass while every real goal still failed to settle.
	writeWorkRealityFile(t, fixture.Root, "docs/context/goals/shipped.md",
		"# Shipped goal\n\nStatus: complete\nCreated: 2026-07-09\n")
	writeWorkRealityFile(t, fixture.Root, "docs/context/goals/yaml-shipped.md",
		"---\nstatus: done\n---\n\n# Yaml shipped goal\n")
	writeWorkRealityFile(t, fixture.Root, "docs/context/goals/open.md",
		"# Open goal\n\nStatus: active\n")
	// A Status line buried below the header must not retire a live goal.
	writeWorkRealityFile(t, fixture.Root, "docs/context/goals/prose.md",
		"# Prose goal\n\nStatus: active\n"+strings.Repeat("\nfiller\n", 12)+"\nStatus: complete\n")
	writeWorkRealityFile(t, fixture.Root, "docs/context/goals/.gitkeep", "")
	runGitForWorkReality(t, fixture.Root, "add", "-A")
	runGitForWorkReality(t, fixture.Root, "commit", "-m", "goals")

	report := runWorkRealityFixture(t, fixture.Root)

	settled := map[string]string{}
	for _, item := range report.Settled {
		settled[item.ID] = item.Status
	}
	if settled["goal:docs/context/goals/shipped.md"] != "complete" {
		t.Fatalf("a goal declaring Status: complete did not settle: %#v", report.Settled)
	}
	if settled["goal:docs/context/goals/yaml-shipped.md"] != "done" {
		t.Fatalf("a goal declaring yaml status: done did not settle: %#v", report.Settled)
	}
	if _, wrong := settled["goal:docs/context/goals/prose.md"]; wrong {
		t.Fatal("a Status line below the header region retired a live goal")
	}
	declared := map[string]bool{}
	for _, item := range report.Declared {
		declared[item.ID] = true
	}
	if declared["goal:docs/context/goals/.gitkeep"] {
		t.Fatal("a .gitkeep placeholder was counted as declared work")
	}
	// An open goal must survive. A fix that settles everything is not a fix.
	for _, item := range report.Settled {
		if item.ID == "goal:docs/context/goals/open.md" {
			t.Fatal("an active goal was wrongly settled")
		}
	}
	if !declared["goal:docs/context/goals/open.md"] {
		t.Fatalf("active goal disappeared from declared work: %#v", report.Declared)
	}
}
