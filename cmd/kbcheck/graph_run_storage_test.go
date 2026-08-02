package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGraphRunStorageDryRunAndApplyRemoveOnlyOwnedTerminalRuns(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 7, 30, 6, 0, 0, 0, time.UTC)

	registerGraphRunForTest(t, root, "terminal", "terminal", false, now.Add(-48*time.Hour))
	registerGraphRunForTest(t, root, "active", "active", false, now.Add(-48*time.Hour))
	registerGraphRunForTest(t, root, "pinned", "terminal", true, now.Add(-48*time.Hour))
	unowned := filepath.Join(root, ".kb", "runs", "unowned")
	if err := os.MkdirAll(unowned, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(unowned, "evidence.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	dry, err := executeGraphRunStorage(graphRunStorageOptions{
		Action: "cleanup", RepoRoot: root, RetentionDays: 1, MaxBytes: 1 << 30, Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !dry.OK || dry.Cleanup == nil || dry.Cleanup.Applied {
		t.Fatalf("unexpected dry-run result: %#v", dry)
	}
	if got := cleanupRunIDs(dry.Cleanup.Selected); !equalStrings(got, []string{"terminal"}) {
		t.Fatalf("selected=%v want terminal", got)
	}
	for _, runID := range []string{"terminal", "active", "pinned", "unowned"} {
		if !pathExists(filepath.Join(root, ".kb", "runs", runID)) {
			t.Fatalf("dry-run removed %s", runID)
		}
	}

	applied, err := executeGraphRunStorage(graphRunStorageOptions{
		Action: "cleanup", RepoRoot: root, RetentionDays: 1, MaxBytes: 1 << 30, Apply: true, Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !applied.OK || applied.Cleanup == nil || !applied.Cleanup.Applied {
		t.Fatalf("unexpected apply result: %#v", applied)
	}
	if pathExists(filepath.Join(root, ".kb", "runs", "terminal")) {
		t.Fatal("owned terminal run was retained")
	}
	for _, runID := range []string{"active", "pinned", "unowned"} {
		if !pathExists(filepath.Join(root, ".kb", "runs", runID)) {
			t.Fatalf("apply removed protected run %s", runID)
		}
	}
}

func TestGraphRunStoragePreservesCorruptAndForgedMarkers(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 7, 30, 6, 0, 0, 0, time.UTC)
	registerGraphRunForTest(t, root, "forged", "terminal", false, now.Add(-48*time.Hour))
	markerPath := filepath.Join(root, ".kb", "runs", "forged", graphRunStorageMarkerName)
	var marker graphRunStorageMarker
	readJSONForTest(t, markerPath, &marker)
	marker.OwnerToken = "forged"
	writeJSONForTest(t, markerPath, marker)

	corrupt := filepath.Join(root, ".kb", "runs", "corrupt")
	if err := os.MkdirAll(corrupt, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(corrupt, graphRunStorageMarkerName), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := executeGraphRunStorage(graphRunStorageOptions{
		Action: "cleanup", RepoRoot: root, RetentionDays: 1, Apply: true, Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("cleanup should retain unsafe paths, got %#v", result)
	}
	for _, runID := range []string{"forged", "corrupt"} {
		if !pathExists(filepath.Join(root, ".kb", "runs", runID)) {
			t.Fatalf("unsafe run %s was removed", runID)
		}
	}
	if statusForRun(t, result.Runs, "forged").Reason != "ownership marker mismatch" {
		t.Fatalf("forged marker reason=%q", statusForRun(t, result.Runs, "forged").Reason)
	}
	if statusForRun(t, result.Runs, "corrupt").Reason != "ownership marker invalid" {
		t.Fatalf("corrupt marker reason=%q", statusForRun(t, result.Runs, "corrupt").Reason)
	}
}

func TestGraphRunStorageByteCapSelectsOldestTerminalRuns(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 7, 30, 6, 0, 0, 0, time.UTC)
	registerGraphRunForTest(t, root, "older", "terminal", false, now.Add(-2*time.Hour))
	registerGraphRunForTest(t, root, "newer", "terminal", false, now.Add(-time.Hour))
	for _, runID := range []string{"older", "newer"} {
		if err := os.WriteFile(filepath.Join(root, ".kb", "runs", runID, "payload"), make([]byte, 128), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	result, err := executeGraphRunStorage(graphRunStorageOptions{
		Action: "cleanup", RepoRoot: root, RetentionDays: 365, MaxBytes: 300, Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Cleanup.Selected) == 0 || result.Cleanup.Selected[0].RunID != "older" {
		t.Fatalf("oldest run was not selected first: %#v", result.Cleanup.Selected)
	}
}

func TestGraphRunStorageRejectsSymlinkedRun(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("use runtime symlink capability check below")
	}
	root := t.TempDir()
	runs := filepath.Join(root, ".kb", "runs")
	outside := filepath.Join(root, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(runs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(runs, "linked")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	result, err := executeGraphRunStorage(graphRunStorageOptions{Action: "cleanup", RepoRoot: root, Apply: true, Now: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	if !pathExists(outside) || statusForRun(t, result.Runs, "linked").Reason != "run path is not a real directory" {
		t.Fatalf("symlink was not retained: %#v", result)
	}
}

func registerGraphRunForTest(t *testing.T, root, runID, status string, pinned bool, now time.Time) {
	t.Helper()
	result, err := executeGraphRunStorage(graphRunStorageOptions{
		Action: "register", RepoRoot: root, RunID: runID, Status: status, Pinned: pinned, Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("register %s: %#v", runID, result)
	}
}

func cleanupRunIDs(items []graphRunCleanupItem) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.RunID)
	}
	return result
}

func statusForRun(t *testing.T, statuses []graphRunStorageStatus, runID string) graphRunStorageStatus {
	t.Helper()
	for _, status := range statuses {
		if status.RunID == runID {
			return status
		}
	}
	t.Fatalf("missing status for %s", runID)
	return graphRunStorageStatus{}
}

func readJSONForTest(t *testing.T, path string, target any) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(content, target); err != nil {
		t.Fatal(err)
	}
}

func writeJSONForTest(t *testing.T, path string, value any) {
	t.Helper()
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(content, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
