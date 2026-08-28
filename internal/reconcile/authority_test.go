package reconcile

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRefreshRemoteAuthorityUnblocksContainedRefRetirement is the regression
// that motivated this code path.
//
// Inventory already computed containment against the cached remote-tracking
// ref, but stamped every proof non-authoritative. branchAmbiguity then refused
// the decision, plan deferred it to a human, and the adapter that could have
// supplied real authority only ran at apply time on planned actions. There were
// no planned actions, so it never ran. The refusal was unreachable by
// construction rather than a judgment about the branch.
//
// The default must stay cached and offline. Opting in must produce a decision.
func TestRefreshRemoteAuthorityUnblocksContainedRefRetirement(t *testing.T) {
	upstream := t.TempDir()
	runGit(t, upstream, "init", "--bare", "--initial-branch=main")

	root := t.TempDir()
	runGit(t, root, "init", "--initial-branch=main")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test")
	writeFile(t, filepath.Join(root, "README.txt"), "contained\n")
	runGit(t, root, "add", "README.txt")
	runGit(t, root, "commit", "-m", "initial")
	runGit(t, root, "remote", "add", "origin", filepath.ToSlash(upstream))
	runGit(t, root, "push", "-u", "origin", "main")
	// A local ref pointing at a commit the default branch already contains.
	runGit(t, root, "branch", "contained-feature")
	runGit(t, root, "fetch", "origin")

	cutoff := time.Date(2026, 8, 1, 16, 30, 0, 0, time.UTC)
	inventory := func(refresh bool) Repository {
		t.Helper()
		ledger, err := Inventory(InventoryOptions{
			Roots: []string{root}, Cutoff: cutoff, CurrentWorktree: root,
			Now: func() time.Time { return cutoff }, RefreshRemoteAuthority: refresh,
		})
		if err != nil {
			t.Fatal(err)
		}
		return ledger.Repositories[0]
	}

	cached := inventory(false)
	if cached.RemoteAuthority.State != "" {
		t.Fatalf("default inventory probed the network: %#v", cached.RemoteAuthority)
	}
	cachedArtifact := artifactForRef(t, cached, "refs/heads/contained-feature")
	if cachedArtifact.Ambiguity != "authoritative-containment-unavailable" {
		t.Fatalf("cached inventory must not claim authority: ambiguity=%q", cachedArtifact.Ambiguity)
	}

	fresh := inventory(true)
	if !fresh.RemoteAuthority.Authoritative() {
		t.Fatalf("opt-in probe failed to establish authority: %#v", fresh.RemoteAuthority)
	}
	freshArtifact := artifactForRef(t, fresh, "refs/heads/contained-feature")
	if freshArtifact.Ambiguity != "" {
		t.Fatalf("refreshed authority still ambiguous: %q", freshArtifact.Ambiguity)
	}

	// Authority must prove containment, not assume it. A ref holding a commit
	// the default branch does not contain stays ambiguous even when the probe
	// succeeded, because deleting it would lose work.
	writeFile(t, filepath.Join(root, "unmerged.txt"), "only here\n")
	runGit(t, root, "checkout", "-b", "unmerged-feature")
	runGit(t, root, "add", "unmerged.txt")
	runGit(t, root, "commit", "-m", "unmerged work")
	runGit(t, root, "checkout", "main")
	withUnmerged := inventory(true)
	unmergedArtifact := artifactForRef(t, withUnmerged, "refs/heads/unmerged-feature")
	if unmergedArtifact.Ambiguity == "" {
		t.Fatal("a ref with unmerged commits must never lose its ambiguity")
	}
}

// TestResolveRemoteAuthorityRefusesWithoutARemote proves the probe fails closed
// rather than defaulting to a probable answer.
func TestResolveRemoteAuthorityRefusesWithoutARemote(t *testing.T) {
	t.Parallel()
	root := initRepo(t)
	authority := ResolveRemoteAuthority(root, nil)
	if authority.Authoritative() || !strings.Contains(authority.Limitation, "no configured remote") {
		t.Fatalf("missing remote was not refused: %#v", authority)
	}
	unreachable := ResolveRemoteAuthority(root, []Remote{{Name: "origin"}})
	if unreachable.Authoritative() || !strings.Contains(unreachable.Limitation, "unreachable") {
		t.Fatalf("unreachable remote was not refused: %#v", unreachable)
	}
}

func artifactForRef(t *testing.T, repository Repository, ref string) Artifact {
	t.Helper()
	for _, artifact := range repository.Artifacts {
		if artifact.Kind == ArtifactBranch && artifact.Ref == ref {
			return artifact
		}
	}
	t.Fatalf("repository has no local ref artifact for %s", ref)
	return Artifact{}
}
