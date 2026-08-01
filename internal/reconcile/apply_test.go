package reconcile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestApplyRetiresCleanWorktreeAndIsIdempotent(t *testing.T) {
	fixture := newApplyFixture(t, false)
	receipt, err := Apply(fixture.options())
	if err != nil {
		t.Fatal(err)
	}
	worktree := actionReceiptForClass(t, receipt, ActionWorktreeRetire)
	if worktree.Result != "applied" || worktree.PhysicalCleanupState != StateVerifiedRetired {
		t.Fatalf("unexpected worktree receipt: %#v", worktree)
	}
	if pathExistsLocal(fixture.worktree) {
		t.Fatal("clean worktree was not retired")
	}
	if got, _ := gitOutput(fixture.root, "rev-parse", "refs/heads/"+fixture.branch); got != fixture.sha {
		t.Fatalf("physical cleanup changed durable local ref: got=%s want=%s", got, fixture.sha)
	}

	options := fixture.options()
	options.ExistingReceipt = &receipt
	repeated, err := Apply(options)
	if err != nil {
		t.Fatal(err)
	}
	if got := actionReceiptForClass(t, repeated, ActionWorktreeRetire); got.Result != "applied" {
		t.Fatalf("repeated apply was not idempotent: %#v", got)
	}
}

func TestApplyFreshEvidencePreservesDriftAndDirt(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, *applyFixture)
		issue  string
	}{
		{
			name: "branch-head",
			mutate: func(t *testing.T, fixture *applyFixture) {
				writeFile(t, filepath.Join(fixture.worktree, "new.txt"), "new\n")
				runGit(t, fixture.worktree, "add", "new.txt")
				runGit(t, fixture.worktree, "commit", "-m", "move target")
			},
			issue: "identity",
		},
		{
			name: "untracked",
			mutate: func(t *testing.T, fixture *applyFixture) {
				writeFile(t, filepath.Join(fixture.worktree, "preserve.txt"), "local\n")
			},
			issue: "dirt",
		},
		{
			name: "ignored",
			mutate: func(t *testing.T, fixture *applyFixture) {
				writeFile(t, filepath.Join(fixture.common, "info", "exclude"), "private.env\n")
				writeFile(t, filepath.Join(fixture.worktree, "private.env"), "preserve\n")
			},
			issue: "dirt",
		},
		{
			name: "locked",
			mutate: func(t *testing.T, fixture *applyFixture) {
				runGit(t, fixture.root, "worktree", "lock", fixture.worktree)
			},
			issue: "locked",
		},
		{
			name: "moved",
			mutate: func(t *testing.T, fixture *applyFixture) {
				moved := filepath.Join(filepath.Dir(fixture.worktree), "moved")
				runGit(t, fixture.root, "worktree", "move", fixture.worktree, moved)
				fixture.moved = moved
			},
			issue: "moved",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newApplyFixture(t, false)
			test.mutate(t, fixture)
			receipt, err := Apply(fixture.options())
			if err != nil {
				t.Fatal(err)
			}
			action := actionReceiptForClass(t, receipt, ActionWorktreeRetire)
			if action.Result != "blocked" || !strings.Contains(strings.ToLower(action.Issue), test.issue) {
				t.Fatalf("fresh evidence did not preserve target: %#v", action)
			}
			path := fixture.worktree
			if fixture.moved != "" {
				path = fixture.moved
			}
			if !pathExistsLocal(path) {
				t.Fatal("drifted target was removed")
			}
		})
	}
}

func TestApplySerializesOnCompatibleRepositoryLock(t *testing.T) {
	fixture := newApplyFixture(t, false)
	lock, err := modelrouting.AcquireSharedProjectLock(
		filepath.Join(fixture.common, ".copilot-kb"), "work-queue.lock", time.Second,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	options := fixture.options()
	options.LockTimeout = 50 * time.Millisecond
	receipt, err := Apply(options)
	if err != nil {
		t.Fatal(err)
	}
	action := actionReceiptForClass(t, receipt, ActionWorktreeRetire)
	if action.Result != "contended" || !pathExistsLocal(fixture.worktree) {
		t.Fatalf("compatible lock did not serialize apply: %#v", action)
	}
}

func TestApplyRepairsOnlyExactEmptyResidual(t *testing.T) {
	for _, nonempty := range []bool{false, true} {
		t.Run(map[bool]string{false: "empty", true: "nonempty"}[nonempty], func(t *testing.T) {
			fixture := newApplyFixture(t, false)
			runGit(t, fixture.root, "worktree", "remove", fixture.worktree)
			if err := os.Mkdir(fixture.worktree, 0o755); err != nil {
				t.Fatal(err)
			}
			if nonempty {
				writeFile(t, filepath.Join(fixture.worktree, "preserve.txt"), "local\n")
			}
			receipt, err := Apply(fixture.options())
			if err != nil {
				t.Fatal(err)
			}
			action := actionReceiptForClass(t, receipt, ActionWorktreeRetire)
			if nonempty {
				if action.Result != "blocked" || !pathExistsLocal(filepath.Join(fixture.worktree, "preserve.txt")) {
					t.Fatalf("nonempty residual was not preserved: %#v", action)
				}
			} else if action.Result != "applied" || pathExistsLocal(fixture.worktree) {
				t.Fatalf("empty residual was not repaired: %#v", action)
			}
		})
	}
}

func TestApplyIntegratedRef(t *testing.T) {
	fixture := newApplyFixture(t, true)
	receipt, err := Apply(fixture.options())
	if err != nil {
		t.Fatal(err)
	}
	worktree := actionReceiptForClass(t, receipt, ActionWorktreeRetire)
	ref := actionReceiptForClass(t, receipt, ActionLocalRefRetire)
	if worktree.PhysicalCleanupState != StateVerifiedRetired ||
		ref.DeliveryState != StateIntegratedDefault ||
		ref.RefRetirementState != StateVerifiedRetired {
		t.Fatalf("independent cleanup gates not proven: worktree=%#v ref=%#v", worktree, ref)
	}
	if got, _ := gitOutput(fixture.root, "show-ref", "--verify", "refs/heads/"+fixture.branch); got != "" {
		t.Fatalf("exact integrated local ref remains: %s", got)
	}
}

func TestApplyRefGuards(t *testing.T) {
	t.Run("cas-mismatch", func(t *testing.T) {
		fixture := newApplyFixture(t, true)
		runGit(t, fixture.root, "worktree", "remove", fixture.worktree)
		writeFile(t, filepath.Join(fixture.root, "move.txt"), "move\n")
		runGit(t, fixture.root, "add", "move.txt")
		runGit(t, fixture.root, "commit", "-m", "move default")
		moved := mustGit(t, fixture.root, "rev-parse", "HEAD")
		runGit(t, fixture.root, "update-ref", "refs/heads/"+fixture.branch, moved, fixture.sha)
		receipt, err := Apply(fixture.options())
		if err != nil {
			t.Fatal(err)
		}
		ref := actionReceiptForClass(t, receipt, ActionLocalRefRetire)
		if ref.Result != "blocked" || !strings.Contains(ref.Issue, "changed after planning") {
			t.Fatalf("CAS mismatch was accepted: %#v", ref)
		}
	})
	t.Run("remote-rewrite", func(t *testing.T) {
		fixture := newApplyFixture(t, true)
		tree := mustGit(t, fixture.root, "rev-parse", fixture.sha+"^{tree}")
		rewritten := mustGit(t, "", "--git-dir", fixture.remote, "commit-tree", tree, "-m", "rewrite")
		runGit(t, "", "--git-dir", fixture.remote, "update-ref", "refs/heads/"+fixture.defaultBranch, rewritten)
		receipt, err := Apply(fixture.options())
		if err != nil {
			t.Fatal(err)
		}
		worktree := actionReceiptForClass(t, receipt, ActionWorktreeRetire)
		if worktree.Result != "blocked" || !strings.Contains(worktree.Issue, "rewritten") {
			t.Fatalf("rewritten remote evidence was accepted: %#v", worktree)
		}
		if !pathExistsLocal(fixture.worktree) {
			t.Fatal("remote rewrite caused worktree mutation")
		}
	})
}

func TestApplyRejectsExpiredOrMismatchedPlanAndProtectedAction(t *testing.T) {
	fixture := newApplyFixture(t, false)
	expired := fixture.options()
	expired.Now = fixture.bundle.Plan.ExpiresAt
	if _, err := Apply(expired); err == nil || !strings.Contains(err.Error(), "not currently valid") {
		t.Fatalf("expired plan was accepted: %v", err)
	}
	mismatch := fixture.options()
	mismatch.Bundle.Plan.PolicyVersion = "other/v1"
	if _, err := Apply(mismatch); err == nil || !strings.Contains(err.Error(), "policy version") {
		t.Fatalf("policy mismatch was accepted: %v", err)
	}

	action := fixture.bundle.Plan.Actions[0]
	action.ID = "action:remote-ref-retire:test"
	action.ActionClass = ActionRemoteRefRetire
	action.Preconditions = passingPredicates(DefaultPolicy(), ActionRemoteRefRetire, fixture.now)
	for repositoryIndex := range fixture.bundle.Ledger.Repositories {
		for artifactIndex := range fixture.bundle.Ledger.Repositories[repositoryIndex].Artifacts {
			artifact := &fixture.bundle.Ledger.Repositories[repositoryIndex].Artifacts[artifactIndex]
			if artifact.ID == action.ArtifactIDs[0] {
				artifact.Predicates = append([]PredicateEvidence(nil), action.Preconditions...)
			}
		}
	}
	fingerprint, err := FingerprintLedger(fixture.bundle.Ledger)
	if err != nil {
		t.Fatal(err)
	}
	for _, repository := range fixture.bundle.Ledger.Repositories {
		for _, artifact := range repository.Artifacts {
			if artifact.ID == action.ArtifactIDs[0] {
				action.Preconditions = append([]PredicateEvidence(nil), artifact.Predicates...)
			}
		}
	}
	fixture.bundle.Plan.LedgerFingerprint = fingerprint
	fixture.bundle.Plan.Actions = []PlannedAction{action}
	receipt, err := Apply(fixture.options())
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Actions[0].Result != "unavailable" || !pathExistsLocal(fixture.worktree) {
		t.Fatalf("protected external action escaped quarantine: %#v", receipt.Actions[0])
	}
}

func TestVerifyKeepsOutcomeDimensionsSeparate(t *testing.T) {
	fixture := newApplyFixture(t, false)
	receipt, err := Apply(fixture.options())
	if err != nil {
		t.Fatal(err)
	}
	verification, updated, err := Verify(fixture.options(), receipt)
	if err != nil {
		t.Fatal(err)
	}
	action := actionReceiptForClass(t, updated, ActionWorktreeRetire)
	if verification.Status != "verified" ||
		action.DeliveryState != StateLocalDurable ||
		action.PhysicalCleanupState != StateVerifiedRetired ||
		action.RefRetirementState != StatePreserved ||
		action.SessionRecordState != StateUnavailable {
		t.Fatalf("verify collapsed independent states: %#v", action)
	}
}

type applyFixture struct {
	root          string
	common        string
	worktree      string
	moved         string
	branch        string
	sha           string
	remote        string
	defaultBranch string
	now           time.Time
	bundle        PlanBundle
}

func newApplyFixture(t *testing.T, includeRef bool) *applyFixture {
	t.Helper()
	root := initApplyRepo(t)
	defaultBranch := mustGit(t, root, "branch", "--show-current")
	worktree := filepath.Join(t.TempDir(), "linked")
	branch := "feature-apply"
	runGit(t, root, "worktree", "add", "-b", branch, worktree)
	sha := mustGit(t, worktree, "rev-parse", "HEAD")
	common := mustGit(t, root, "rev-parse", "--git-common-dir")
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	common = canonicalPath(common)
	now := time.Date(2026, 8, 1, 17, 30, 0, 0, time.UTC)
	writeQueueForApply(t, common, QueueClaim{
		WorkID: "apply-work", SessionID: "session-owner", Status: "done",
		Branch: branch, Worktree: worktree, UpdatedAt: now.Add(-time.Minute),
	})
	fixture := &applyFixture{
		root: root, common: common, worktree: worktree, branch: branch, sha: sha,
		defaultBranch: defaultBranch, now: now,
	}
	if includeRef {
		fixture.remote = filepath.Join(t.TempDir(), "remote.git")
		runGit(t, "", "init", "--bare", fixture.remote)
		runGit(t, "", "--git-dir", fixture.remote, "config", "core.longpaths", "true")
		runGit(t, root, "remote", "add", "origin", fixture.remote)
		runGit(t, root, "push", "origin", defaultBranch)
		runGit(t, "", "--git-dir", fixture.remote, "symbolic-ref", "HEAD", "refs/heads/"+defaultBranch)
		runGit(t, root, "fetch", "origin")
		runGit(t, root, "remote", "set-head", "origin", "-a")
	}
	fixture.bundle = buildApplyBundle(t, fixture, includeRef)
	return fixture
}

func initApplyRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "core.longpaths", "true")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test")
	writeFile(t, filepath.Join(root, "README.txt"), "plain repository\n")
	runGit(t, root, "add", "README.txt")
	runGit(t, root, "commit", "-m", "initial")
	return root
}

func buildApplyBundle(t *testing.T, fixture *applyFixture, includeRef bool) PlanBundle {
	t.Helper()
	ledger, err := Inventory(InventoryOptions{
		Roots: []string{fixture.root}, Cutoff: fixture.now,
		CurrentWorktree: fixture.root, CurrentSessionID: "session-sweeper",
		Now: func() time.Time { return fixture.now },
	})
	if err != nil {
		t.Fatal(err)
	}
	repository := &ledger.Repositories[0]
	proof := DedupProof{
		Algorithm: DedupCommitAncestry, Identity: "ancestor:" + fixture.sha,
		ObservedAt: fixture.now, Authoritative: true,
	}
	for index := range repository.Artifacts {
		artifact := &repository.Artifacts[index]
		if artifact.Kind == ArtifactWorktree && sameCanonicalPath(artifact.Path, fixture.worktree) {
			artifact.Predicates = passingPredicates(DefaultPolicy(), ActionWorktreeRetire, fixture.now)
			artifact.DedupProofs = []DedupProof{proof}
			artifact.ProtectionReasons = nil
		}
		if includeRef && artifact.Kind == ArtifactBranch && artifact.Ref == "refs/heads/"+fixture.branch {
			artifact.Predicates = passingPredicates(DefaultPolicy(), ActionLocalRefRetire, fixture.now)
			artifact.DedupProofs = []DedupProof{proof}
			artifact.ProtectionReasons = nil
			artifact.Ambiguity = ""
			artifact.UpdatedAt = time.Time{}
		}
	}
	plan, err := BuildPlan(ledger, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if want := 1; includeRef {
		want = 2
		if len(plan.Actions) != want {
			t.Fatalf("fixture planned %d actions, want %d: %#v", len(plan.Actions), want, plan.Outcomes)
		}
	} else if len(plan.Actions) != want {
		t.Fatalf("fixture planned %d actions, want %d: %#v", len(plan.Actions), want, plan.Outcomes)
	}
	return PlanBundle{Ledger: ledger, Plan: plan}
}

func (fixture *applyFixture) options() ApplyOptions {
	return ApplyOptions{
		Bundle: fixture.bundle, Policy: DefaultPolicy(),
		CurrentWorktree: fixture.root, CurrentSession: "session-sweeper",
		Now: fixture.now.Add(time.Minute), LockTimeout: time.Second,
	}
}

func writeQueueForApply(t *testing.T, common string, claim QueueClaim) {
	t.Helper()
	path := filepath.Join(common, ".copilot-kb", "work-queue.json")
	content, err := MarshalStable([]QueueClaim{claim})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func actionReceiptForClass(t *testing.T, receipt ApplyReceipt, class string) ActionReceipt {
	t.Helper()
	for _, action := range receipt.Actions {
		if action.ActionClass == class {
			return action
		}
	}
	t.Fatalf("receipt has no %s action: %#v", class, receipt)
	return ActionReceipt{}
}

func mustGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	value, err := gitOutput(root, args...)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
