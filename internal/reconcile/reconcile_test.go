package reconcile

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestInventoryNoKBRepo(t *testing.T) {
	root := initRepo(t)
	writeFile(t, filepath.Join(root, ".gitignore"), "runtime.db\n")
	runGit(t, root, "add", ".gitignore")
	runGit(t, root, "commit", "-m", "ignore runtime data")

	linked := filepath.Join(t.TempDir(), "linked")
	runGit(t, root, "worktree", "add", "-b", "feature", linked)
	writeFile(t, filepath.Join(linked, "notes.txt"), "unique\n")
	writeFile(t, filepath.Join(linked, "runtime.db"), "metadata only\n")

	cutoff := time.Date(2026, 8, 1, 16, 30, 0, 0, time.UTC)
	ledger, err := Inventory(InventoryOptions{
		Roots:            []string{root},
		Cutoff:           cutoff,
		CurrentWorktree:  root,
		CurrentSessionID: "session-current",
		Now:              func() time.Time { return cutoff },
	})
	if err != nil {
		t.Fatal(err)
	}
	if ledger.SchemaVersion != LedgerSchemaVersion || len(ledger.Repositories) != 1 {
		t.Fatalf("unexpected ledger: %#v", ledger)
	}
	repo := ledger.Repositories[0]
	if len(repo.Worktrees) != 2 {
		t.Fatalf("worktrees=%d, want 2", len(repo.Worktrees))
	}
	var linkedInventory Worktree
	for _, worktree := range repo.Worktrees {
		if samePath(worktree.Path, linked) {
			linkedInventory = worktree
		}
	}
	if len(linkedInventory.Dirt.Untracked) != 1 || len(linkedInventory.Dirt.Ignored) != 1 {
		t.Fatalf("distinct dirt not inventoried: %#v", linkedInventory.Dirt)
	}
	if !contains(linkedInventory.ProtectionReasons, "ignored-data") ||
		!contains(linkedInventory.ProtectionReasons, "model-learning-credential-live") {
		t.Fatalf("protected ignored data was not classified: %#v", linkedInventory.ProtectionReasons)
	}
	for _, name := range []string{"host-sessions", "pull-requests", "remote-authority"} {
		evidence := findEvidence(repo.Evidence, name)
		if evidence.State != EvidenceUnavailable {
			t.Fatalf("%s state=%s, want unavailable", name, evidence.State)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "cmd", "kbcheck")); !os.IsNotExist(err) {
		t.Fatal("fixture unexpectedly contains KB files")
	}
	var currentSession Artifact
	for _, artifact := range repo.Artifacts {
		if artifact.ID == "session:session-current" {
			currentSession = artifact
		}
	}
	if !contains(currentSession.ProtectionReasons, "current") {
		t.Fatalf("current session identity was not protected: %#v", currentSession)
	}
}

func TestPlanMixedPortfolioDecisionPacket(t *testing.T) {
	cutoff := time.Date(2026, 8, 1, 16, 30, 0, 0, time.UTC)
	artifacts := make([]Artifact, 0, 20)
	for index := 0; index < 8; index++ {
		artifacts = append(artifacts, routineArtifact(index, cutoff))
	}
	for index, reason := range []string{
		"current", "primary", "default", "active-claim", "post-cutoff",
		"tracked-dirt", "ignored-data", "model-learning-credential-live",
	} {
		artifacts = append(artifacts, Artifact{
			ID: "protected-" + reason, Kind: ArtifactWorktree, RepositoryID: "repo-1",
			ObservedAt: cutoff, ProtectionReasons: []string{reason},
			Predicates: passingPredicates(DefaultPolicy(), ActionWorktreeRetire, cutoff),
			UniqueWork: index == 7,
		})
	}
	artifacts = append(artifacts,
		Artifact{
			ID: "unique-safe", Kind: ArtifactBranch, RepositoryID: "repo-1", ObservedAt: cutoff,
			UniqueWork: true, SalvageSafe: true,
			Predicates: passingPredicates(DefaultPolicy(), ActionSalvage, cutoff),
		},
		Artifact{
			ID: "unique-credential", Kind: ArtifactBranch, RepositoryID: "repo-1", ObservedAt: cutoff,
			UniqueWork: true, SalvageSafe: false,
			ProtectionReasons: []string{"model-learning-credential-live"},
		},
		Artifact{
			ID: "ambiguous-a", Kind: ArtifactBranch, RepositoryID: "repo-1", ObservedAt: cutoff,
			Ambiguity: "provider-merge-identity-unavailable",
		},
		Artifact{
			ID: "ambiguous-b", Kind: ArtifactBranch, RepositoryID: "repo-1", ObservedAt: cutoff,
			Ambiguity: "provider-merge-identity-unavailable",
		},
	)
	if len(artifacts) != 20 {
		t.Fatalf("fixture has %d artifacts, want 20", len(artifacts))
	}
	ledger := Ledger{
		SchemaVersion: LedgerSchemaVersion,
		Cutoff:        cutoff,
		GeneratedAt:   cutoff,
		Repositories: []Repository{{
			ID: "repo-1", Root: `C:\repo`, Artifacts: artifacts,
		}},
	}
	policy := DefaultPolicy()
	policy.RiskBudget.PerRun[ActionWorktreeRetire] = 3
	policy.RiskBudget.PerRepository[ActionWorktreeRetire] = 2

	plan, err := BuildPlan(ledger, policy)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.DecisionPacket.Items) != 1 {
		t.Fatalf("decision groups=%d, want one compact group: %#v", len(plan.DecisionPacket.Items), plan.DecisionPacket)
	}
	if len(plan.DecisionPacket.Items[0].ArtifactIDs) != 2 {
		t.Fatalf("ambiguity was not grouped: %#v", plan.DecisionPacket.Items[0])
	}
	routinePrompts := 0
	for _, outcome := range plan.Outcomes {
		if strings.HasPrefix(outcome.ArtifactID, "routine-") && outcome.DecisionID != "" {
			routinePrompts++
		}
		if (outcome.ArtifactID == "unique-safe" || outcome.ArtifactID == "unique-credential") &&
			(outcome.Classification == ClassificationRoutineRetire || outcome.Classification == ClassificationSafeSupersede) {
			t.Fatalf("unique work became destructive: %#v", outcome)
		}
	}
	if routinePrompts != 0 {
		t.Fatalf("routine artifacts produced %d prompts", routinePrompts)
	}
	if got := deferredCount(plan.Health, ActionWorktreeRetire); got != 6 {
		t.Fatalf("budget deferred=%d, want 6", got)
	}
	if plan.Health.RunsToConvergence != 4 {
		t.Fatalf("runs-to-convergence=%d, want 4", plan.Health.RunsToConvergence)
	}
	if len(plan.Actions) != 3 { // two retire actions plus one additive salvage action
		t.Fatalf("actions=%d, want 3: %#v", len(plan.Actions), plan.Actions)
	}
}

func TestPlanMissingMandatoryEvidenceFailsClosed(t *testing.T) {
	cutoff := time.Date(2026, 8, 1, 16, 30, 0, 0, time.UTC)
	artifact := routineArtifact(0, cutoff)
	artifact.Predicates = artifact.Predicates[1:]
	ledger := Ledger{
		SchemaVersion: LedgerSchemaVersion, Cutoff: cutoff, GeneratedAt: cutoff,
		Repositories: []Repository{{ID: "repo-1", Root: "repo", Artifacts: []Artifact{artifact}}},
	}
	plan, err := BuildPlan(ledger, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	outcome := plan.Outcomes[0]
	if outcome.Classification != ClassificationQuarantine || outcome.Confidence == 1 {
		t.Fatalf("missing mandatory evidence did not fail closed: %#v", outcome)
	}
	if len(plan.Actions) != 0 {
		t.Fatalf("missing evidence planned mutation: %#v", plan.Actions)
	}
}

func TestPlanStableJSONAndDedupVocabulary(t *testing.T) {
	cutoff := time.Date(2026, 8, 1, 16, 30, 0, 0, time.UTC)
	ledger := Ledger{
		SchemaVersion: LedgerSchemaVersion, Cutoff: cutoff, GeneratedAt: cutoff,
		Repositories: []Repository{{ID: "repo-1", Root: "repo", Artifacts: []Artifact{routineArtifact(0, cutoff)}}},
	}

	first, err := BuildPlan(ledger, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildPlan(ledger, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, _ := MarshalStable(first)
	secondJSON, _ := MarshalStable(second)
	if string(firstJSON) != string(secondJSON) {
		t.Fatal("stable inputs produced different JSON")
	}
	for _, algorithm := range []string{
		DedupIdenticalBlob, DedupIdenticalTree, DedupSameBasePatch,
		DedupCommitAncestry, DedupRemoteTopicContainment, DedupRemoteDefaultAncestry,
		DedupProviderMergeTree, DedupProviderMergePatch,
	} {
		if !contains(DefaultPolicy().DedupAlgorithms, algorithm) {
			t.Fatalf("missing exact dedup algorithm %q", algorithm)
		}
	}
}

func TestPolicyManifestMatchesDefault(t *testing.T) {
	loaded, err := LoadPolicy(filepath.Join("..", "..", "config", "reconcile-predicates.json"))
	if err != nil {
		t.Fatal(err)
	}
	want, _ := MarshalStable(DefaultPolicy())
	got, _ := MarshalStable(loaded)
	if string(got) != string(want) {
		t.Fatalf("default policy drifted from versioned manifest: %s", firstJSONDiff(want, got, "$"))
	}
}

func firstJSONDiff(wantJSON, gotJSON []byte, path string) string {
	var want, got any
	_ = json.Unmarshal(wantJSON, &want)
	_ = json.Unmarshal(gotJSON, &got)
	return firstValueDiff(want, got, path)
}

func firstValueDiff(want, got any, path string) string {
	switch wantValue := want.(type) {
	case map[string]any:
		gotValue, ok := got.(map[string]any)
		if !ok {
			return path + " type mismatch"
		}
		keys := make([]string, 0, len(wantValue))
		for key := range wantValue {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if diff := firstValueDiff(wantValue[key], gotValue[key], path+"."+key); diff != "" {
				return diff
			}
		}
	case []any:
		gotValue, ok := got.([]any)
		if !ok || len(wantValue) != len(gotValue) {
			return fmt.Sprintf("%s length/type mismatch: want=%d", path, len(wantValue))
		}
		for index := range wantValue {
			if diff := firstValueDiff(wantValue[index], gotValue[index], fmt.Sprintf("%s[%d]", path, index)); diff != "" {
				return diff
			}
		}
	default:
		if fmt.Sprint(want) != fmt.Sprint(got) {
			return fmt.Sprintf("%s want=%v got=%v", path, want, got)
		}
	}
	return ""
}

func routineArtifact(index int, cutoff time.Time) Artifact {
	return Artifact{
		ID: "routine-" + string(rune('a'+index)), Kind: ArtifactWorktree, RepositoryID: "repo-1",
		ObservedAt: cutoff.Add(-24 * time.Hour),
		Predicates: passingPredicates(DefaultPolicy(), ActionWorktreeRetire, cutoff),
		DedupProofs: []DedupProof{{
			Algorithm: DedupCommitAncestry, Identity: "ancestor:abc:def",
			Authoritative: true, ObservedAt: cutoff,
		}},
	}
}

func passingPredicates(policy Policy, actionClass string, observed time.Time) []PredicateEvidence {
	var predicates []PredicateEvidence
	for _, action := range policy.Actions {
		if action.Class != actionClass {
			continue
		}
		for _, predicate := range action.Mandatory {
			predicates = append(predicates, PredicateEvidence{
				Name: predicate.Name, State: PredicatePass, Source: predicate.Adapters[0],
				Authoritative: true, ObservedAt: observed,
			})
		}
	}
	return predicates
}

func deferredCount(health ConvergenceHealth, actionClass string) int {
	for _, item := range health.DeferredByAction {
		if item.ActionClass == actionClass {
			return item.Count
		}
	}
	return 0
}

func findEvidence(items []Evidence, name string) Evidence {
	for _, item := range items {
		if item.Name == name {
			return item
		}
	}
	return Evidence{}
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func samePath(left, right string) bool {
	leftPath, _ := filepath.Abs(left)
	rightPath, _ := filepath.Abs(right)
	return strings.EqualFold(filepath.Clean(leftPath), filepath.Clean(rightPath))
}

func initRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test")
	writeFile(t, filepath.Join(root, "README.txt"), "plain repository\n")
	runGit(t, root, "add", "README.txt")
	runGit(t, root, "commit", "-m", "initial")
	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
