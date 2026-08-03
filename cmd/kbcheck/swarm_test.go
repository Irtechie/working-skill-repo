package main

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestReadySetParallelSlices(t *testing.T) {
	t.Parallel()
	path := writeManifest(t, `
---
type: kb-manifest
slices:
  - id: slice-001
    blockers: []
    status: pending
    hitl: false
    can_continue_other_slices: true
  - id: slice-002
    blockers: []
    status: pending
    hitl: false
    can_continue_other_slices: true
  - id: slice-003
    blockers: [slice-001]
    status: pending
    hitl: false
    can_continue_other_slices: true
---
`)
	result, err := computeReadySet(path)
	if err != nil {
		t.Fatalf("computeReadySet returned error: %v", err)
	}
	ready := result.(readySetResult)
	if !reflect.DeepEqual(ready.Ready, []string{"slice-001", "slice-002"}) {
		t.Fatalf("ready=%v", ready.Ready)
	}
}

func TestCrossManifestSchedulerReadySetRequiresLiveManifestAuthority(t *testing.T) {
	t.Parallel()
	manifest := writeManifest(t, `
---
slices:
  - id: slice-001
    blockers: []
    status: pending
    hitl: false
    can_continue_other_slices: true
---
`)
	stateRoot := t.TempDir()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	blocked, err := computeCrossManifestReadySet(manifest, stateRoot, "run-a", "owner-a", now)
	if err != nil || blocked.OK || blocked.Reason != "manifest-mutation-authority-required" {
		t.Fatalf("ready set bypassed manifest authority: result=%#v err=%v", blocked, err)
	}
	acquired, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-a", ManifestPath: manifest,
		OwnerToken: "owner-a", Files: []string{"src/a.go"}, Now: now,
	})
	if err != nil || !acquired.OK {
		t.Fatalf("plan-run acquire failed: result=%#v err=%v", acquired, err)
	}
	ready, err := computeCrossManifestReadySet(manifest, stateRoot, "run-a", "owner-a", now)
	if err != nil || !ready.OK || !reflect.DeepEqual(ready.Ready, []string{"slice-001"}) {
		t.Fatalf("authorized ready set failed: result=%#v err=%v", ready, err)
	}
}

func TestCrossManifestSchedulerSerializesOneManifestOwnedWorktree(t *testing.T) {
	t.Parallel()
	legacyManifest := `
---
slices:
  - id: slice-001
    blockers: []
    status: pending
    hitl: false
    can_continue_other_slices: true
  - id: slice-002
    blockers: []
    status: pending
    hitl: false
    can_continue_other_slices: true
---
`
	manifest := writeManifest(t, legacyManifest)
	stateRoot := t.TempDir()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	acquired, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-serial", ManifestPath: manifest,
		OwnerToken: "owner-serial", Files: []string{"src/a.go", "src/b.go"}, Now: now,
	})
	if err != nil || !acquired.OK {
		t.Fatalf("plan lease acquire failed: result=%#v err=%v", acquired, err)
	}
	writeFile(t, manifest, strings.TrimLeft(strings.Replace(legacyManifest, "slices:", "workspace_isolation_contract:\n  plan_run_worktree_default: true\nslices:", 1), "\n"))

	ready, err := computeCrossManifestReadySet(manifest, stateRoot, "run-serial", "owner-serial", now)
	if err != nil || !ready.OK || !reflect.DeepEqual(ready.Ready, []string{"slice-001"}) {
		t.Fatalf("shared-serial manifest dispatched more than one slice: result=%#v err=%v", ready, err)
	}
	slice, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, SliceID: "slice-001", RunID: "run-serial",
		OwnerToken: "owner-serial", Files: []string{"src/a.go"}, Now: now,
	})
	if err != nil || !slice.OK {
		t.Fatalf("slice lease acquire failed: result=%#v err=%v", slice, err)
	}
	blocked, err := computeCrossManifestReadySet(manifest, stateRoot, "run-serial", "owner-serial", now)
	if err != nil || blocked.OK || blocked.Reason != "manifest-shared-worktree-slice-in-flight" {
		t.Fatalf("second in-flight slice was admitted: result=%#v err=%v", blocked, err)
	}
}

func TestPlanRunLeaseCommandReportsClaimOwner(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifest := filepath.Join(root, "manifest.md")
	writeFile(t, manifest, "---\ntype: kb-manifest\nslices: []\n---\n")
	stateRoot := filepath.Join(root, "state")
	var out, errOut strings.Builder
	code := run([]string{
		"plan-run-lease", "--action", "acquire", "--root", root, "--state-root", stateRoot,
		"--run-id", "run-a", "--manifest", manifest, "--owner-token", "owner-a",
		"--file", "src/a.go", "--domain", "skill:kb-work", "--json",
	}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), `"run_id": "run-a"`) ||
		!strings.Contains(out.String(), `"coordination_scope": "git-common-dir"`) ||
		strings.Contains(out.String(), "owner-a") {
		t.Fatalf("plan-run acquire command failed: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()
	code = run([]string{
		"plan-run-lease", "--action", "acquire", "--root", root, "--state-root", stateRoot,
		"--run-id", "run-b", "--manifest", manifest, "--owner-token", "owner-b",
		"--file", "src/a.go", "--json",
	}, &out, &errOut)
	if code != 2 || !strings.Contains(out.String(), `"run_id": "run-a"`) ||
		strings.Contains(out.String(), "owner-a") || strings.Contains(out.String(), "owner-b") {
		t.Fatalf("collision owner missing: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}

	out.Reset()
	errOut.Reset()
	code = run([]string{
		"plan-run-lease", "--action", "status", "--root", root, "--state-root", stateRoot,
		"--run-id", "run-a", "--owner-token", "wrong-owner", "--json",
	}, &out, &errOut)
	if code != 2 || strings.Contains(out.String(), "owner-a") || !strings.Contains(out.String(), `"owner_fingerprint"`) {
		t.Fatalf("status leaked bearer token: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestReadySetSerialExclusionAndSingleSerial(t *testing.T) {
	t.Parallel()
	serial := writeManifest(t, `
---
slices:
  - id: slice-001
    blockers: []
    status: pending
    hitl: false
    can_continue_other_slices: false
  - id: slice-002
    blockers: []
    status: pending
    hitl: false
    can_continue_other_slices: true
---
`)
	result, err := computeReadySet(serial)
	if err != nil {
		t.Fatalf("computeReadySet returned error: %v", err)
	}
	ready := result.(readySetResult)
	if !reflect.DeepEqual(ready.Ready, []string{"slice-002"}) || !reflect.DeepEqual(ready.ExcludedSerial, []string{"slice-001"}) {
		t.Fatalf("ready=%v excluded=%v", ready.Ready, ready.ExcludedSerial)
	}

	single := writeManifest(t, `
---
slices:
  - id: slice-001
    blockers: []
    status: pending
    hitl: false
    can_continue_other_slices: false
---
`)
	result, err = computeReadySet(single)
	if err != nil {
		t.Fatalf("computeReadySet returned error: %v", err)
	}
	ready = result.(readySetResult)
	if !reflect.DeepEqual(ready.Ready, []string{"slice-001"}) || len(ready.ExcludedSerial) != 0 {
		t.Fatalf("ready=%v excluded=%v", ready.Ready, ready.ExcludedSerial)
	}
}

func TestReadySetFiltersStatusesAndDetectsCycles(t *testing.T) {
	t.Parallel()
	states := writeManifest(t, `
---
slices:
  - id: slice-001
    blockers: []
    status: done
    hitl: false
    can_continue_other_slices: true
  - id: slice-002
    blockers: []
    status: skipped
    hitl: false
    can_continue_other_slices: true
  - id: slice-003
    blockers: []
    status: blocked
    hitl: false
    can_continue_other_slices: true
  - id: slice-004
    blockers: [slice-001, slice-002]
    status: pending
    hitl: false
    can_continue_other_slices: true
---
`)
	result, err := computeReadySet(states)
	if err != nil {
		t.Fatalf("computeReadySet returned error: %v", err)
	}
	ready := result.(readySetResult)
	if !reflect.DeepEqual(ready.Ready, []string{"slice-004"}) {
		t.Fatalf("ready=%v", ready.Ready)
	}

	cycle := writeManifest(t, `
---
slices:
  - id: slice-001
    blockers: [slice-002]
    status: pending
    hitl: false
    can_continue_other_slices: true
  - id: slice-002
    blockers: [slice-001]
    status: pending
    hitl: false
    can_continue_other_slices: true
---
`)
	result, err = computeReadySet(cycle)
	if err != nil {
		t.Fatalf("computeReadySet returned error: %v", err)
	}
	if cycleResult, ok := result.(cycleResult); !ok || cycleResult.OK {
		t.Fatalf("expected cycle result, got %#v", result)
	}
}

func TestScopeLeaseCollisionAndRelease(t *testing.T) {
	t.Parallel()
	disjoint := writeLedger(t, []scopeLeaseEntry{
		{SliceID: "slice-001", Path: "src/a.ts", Status: "active"},
		{SliceID: "slice-002", Path: "src/b.ts", Status: "active"},
	})
	result, err := computeScopeLease(disjoint)
	if err != nil {
		t.Fatalf("computeScopeLease returned error: %v", err)
	}
	if !result.OK || len(result.ActiveLeases) != 2 {
		t.Fatalf("expected disjoint pass, got %#v", result)
	}

	collision := writeLedger(t, []scopeLeaseEntry{
		{SliceID: "slice-001", Path: "src/shared.ts", Status: "active"},
		{SliceID: "slice-002", Path: "src/shared.ts", Status: "active"},
	})
	result, err = computeScopeLease(collision)
	if err != nil {
		t.Fatalf("computeScopeLease returned error: %v", err)
	}
	if result.OK || len(result.Collisions) != 1 {
		t.Fatalf("expected collision, got %#v", result)
	}

	released := writeLedger(t, []scopeLeaseEntry{
		{SliceID: "slice-001", Path: "src/shared.ts", Status: "active"},
		{SliceID: "slice-001", Path: "src/shared.ts", Status: "done"},
		{SliceID: "slice-002", Path: "src/shared.ts", Status: "writing"},
	})
	result, err = computeScopeLease(released)
	if err != nil {
		t.Fatalf("computeScopeLease returned error: %v", err)
	}
	if !result.OK || result.ActiveLeases[0].SliceID != "slice-002" {
		t.Fatalf("expected released path, got %#v", result)
	}

	sameSliceDifferentOwner := writeLedger(t, []scopeLeaseEntry{
		{SliceID: "slice-001", OwnerToken: "owner-a", Path: "src/shared.ts", Status: "active"},
		{SliceID: "slice-001", OwnerToken: "owner-b", Path: "src/shared.ts", Status: "active"},
	})
	result, err = computeScopeLease(sameSliceDifferentOwner)
	if err != nil {
		t.Fatalf("computeScopeLease returned error: %v", err)
	}
	if result.OK || len(result.Collisions) != 1 {
		t.Fatalf("expected same-slice owner collision, got %#v", result)
	}
}

func TestReadySetAndScopeLeaseCommands(t *testing.T) {
	t.Parallel()
	manifest := writeManifest(t, `
---
slices:
  - id: slice-001
    blockers: []
    status: pending
    hitl: false
    can_continue_other_slices: true
---
`)
	var out, errOut strings.Builder
	code := run([]string{"ready-set", "--manifest", manifest, "--json"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("ready-set command failed: code=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"slice-001"`) {
		t.Fatalf("missing ready slice: %s", out.String())
	}

	ledger := writeLedger(t, []scopeLeaseEntry{{SliceID: "slice-001", Path: "src/a.ts", Status: "active"}})
	out.Reset()
	errOut.Reset()
	code = run([]string{"scope-lease", "--ledger", ledger, "--json"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("scope-lease command failed: code=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"ok": true`) {
		t.Fatalf("missing ok result: %s", out.String())
	}
}

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "manifest.md")
	writeFile(t, path, strings.TrimLeft(content, "\n"))
	return path
}

func writeLedger(t *testing.T, entries []scopeLeaseEntry) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ledger.json")
	content, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		t.Fatalf("marshal ledger: %v", err)
	}
	writeFile(t, path, string(content))
	return path
}
