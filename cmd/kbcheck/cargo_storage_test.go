package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCargoStorageResolveIsStableAcrossWorktrees(t *testing.T) {
	root := initWorktreeRepo(t)
	worktree, _, _ := createTerminalCleanupWorktree(t, root, "cargo-storage")
	cacheRoot := filepath.Join(t.TempDir(), "cache")
	t.Setenv("CARGO_TARGET_DIR", "target")

	first, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-a", RepoRoot: root, CacheRoot: cacheRoot, Now: time.Now().UTC(),
	})
	if err != nil || !first.OK || first.Receipt == nil {
		t.Fatalf("root resolve failed: result=%#v err=%v", first, err)
	}
	second, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-b", RepoRoot: worktree, CacheRoot: cacheRoot, Now: time.Now().UTC(),
	})
	if err != nil || !second.OK || second.Receipt == nil {
		t.Fatalf("worktree resolve failed: result=%#v err=%v", second, err)
	}
	if first.Receipt.StableTarget != second.Receipt.StableTarget || !filepath.IsAbs(first.Receipt.StableTarget) {
		t.Fatalf("worktrees resolved different targets: root=%s worktree=%s", first.Receipt.StableTarget, second.Receipt.StableTarget)
	}
	if first.Receipt.AppliedEnvironment["CARGO_TARGET_DIR"] != first.Receipt.StableTarget {
		t.Fatalf("resolved target was not enforced in environment: %#v", first.Receipt.AppliedEnvironment)
	}
}

func TestCargoStorageKeysExternalAbsoluteCacheRootByRepository(t *testing.T) {
	root := initWorktreeRepo(t)
	cacheRoot := filepath.Join(t.TempDir(), "shared-target")
	t.Setenv("CARGO_TARGET_DIR", cacheRoot)

	result, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-absolute", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !result.OK || result.Receipt == nil {
		t.Fatalf("resolve failed: result=%#v err=%v", result, err)
	}
	if filepath.Dir(result.Receipt.StableTarget) != cacheRoot ||
		result.Receipt.StableSource != "environment-cache-root" {
		t.Fatalf("absolute cache root was not project-keyed: %#v", result.Receipt)
	}
}

func TestCargoStorageSeparatesRepositoriesUnderExternalCacheRoot(t *testing.T) {
	firstRoot := initWorktreeRepo(t)
	secondRoot := initWorktreeRepo(t)
	cacheRoot := filepath.Join(t.TempDir(), "shared-target")
	t.Setenv("CARGO_TARGET_DIR", cacheRoot)

	first, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-first", RepoRoot: firstRoot, Now: time.Now().UTC(),
	})
	if err != nil || !first.OK {
		t.Fatalf("first resolve failed: result=%#v err=%v", first, err)
	}
	second, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-second", RepoRoot: secondRoot, Now: time.Now().UTC(),
	})
	if err != nil || !second.OK {
		t.Fatalf("second resolve failed: result=%#v err=%v", second, err)
	}
	if first.Receipt.StableTarget == second.Receipt.StableTarget {
		t.Fatalf("unrelated repositories shared one Cargo target: %s", first.Receipt.StableTarget)
	}
}

func TestCargoStorageResolveResumesAuthoritativeReceipt(t *testing.T) {
	root := initWorktreeRepo(t)
	firstCache := filepath.Join(t.TempDir(), "first-cache")
	secondCache := filepath.Join(t.TempDir(), "second-cache")
	first, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-resume", RepoRoot: root,
		CacheRoot: firstCache, Now: time.Now().UTC(),
	})
	if err != nil || !first.OK {
		t.Fatalf("first resolve failed: result=%#v err=%v", first, err)
	}
	writeFile(t, filepath.Join(root, ".cargo", "config.toml"), "[build]\nincremental = true\n")
	stale, err := executeCargoStorage(cargoStorageOptions{
		Action: "validate-ready", RunID: "run-resume", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || stale.OK {
		t.Fatalf("stale Cargo config fingerprint was accepted: result=%#v err=%v", stale, err)
	}
	resumed, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-resume", RepoRoot: root,
		CacheRoot: secondCache, Now: time.Now().UTC(),
	})
	if err != nil || !resumed.OK {
		t.Fatalf("resume failed: result=%#v err=%v", resumed, err)
	}
	if resumed.Receipt.StableTarget != first.Receipt.StableTarget {
		t.Fatalf("resume abandoned authoritative target: first=%s resumed=%s",
			first.Receipt.StableTarget, resumed.Receipt.StableTarget)
	}
	ready, err := executeCargoStorage(cargoStorageOptions{
		Action: "validate-ready", RunID: "run-resume", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !ready.OK {
		t.Fatalf("re-resolve did not refresh Cargo config fingerprint: result=%#v err=%v", ready, err)
	}
}

func TestCargoStorageRemapsAbsoluteTargetInsideAnyLinkedWorktree(t *testing.T) {
	root := initWorktreeRepo(t)
	worktree, _, _ := createTerminalCleanupWorktree(t, root, "cargo-local-target")
	localTarget := filepath.Join(root, "target")
	cacheRoot := filepath.Join(t.TempDir(), "cache")
	t.Setenv("CARGO_TARGET_DIR", localTarget)

	result, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-local-target", RepoRoot: worktree,
		CacheRoot: cacheRoot, Now: time.Now().UTC(),
	})
	if err != nil || !result.OK || result.Receipt == nil {
		t.Fatalf("resolve failed: result=%#v err=%v", result, err)
	}
	if result.Receipt.StableTarget == localTarget || !strings.HasPrefix(result.Receipt.StableTarget, cacheRoot) {
		t.Fatalf("worktree-local absolute target was not remapped: %#v", result.Receipt)
	}
}

func TestCargoStorageFinalizesOnlyOwnedContainedTemporaryTarget(t *testing.T) {
	root := initWorktreeRepo(t)
	cacheRoot := filepath.Join(t.TempDir(), "cache")
	tempRoot := filepath.Join(t.TempDir(), "run-temp")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	resolved, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-clean", RepoRoot: root, CacheRoot: cacheRoot, Now: time.Now().UTC(),
	})
	if err != nil || !resolved.OK {
		t.Fatalf("resolve failed: result=%#v err=%v", resolved, err)
	}
	ready, err := executeCargoStorage(cargoStorageOptions{
		Action: "validate-ready", RunID: "run-clean", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !ready.OK {
		t.Fatalf("newly resolved receipt was not execution-ready: result=%#v err=%v", ready, err)
	}
	target := filepath.Join(tempRoot, "isolated-target")
	registered, err := executeCargoStorage(cargoStorageOptions{
		Action: "register-temp", RunID: "run-clean", RepoRoot: root,
		Target: target, TempRoot: tempRoot, Reason: "incompatible compiler flags", Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK {
		t.Fatalf("register failed: result=%#v err=%v", registered, err)
	}
	writeFile(t, filepath.Join(target, "artifact.bin"), strings.Repeat("x", 64))

	finalized, err := executeCargoStorage(cargoStorageOptions{
		Action: "finalize", RunID: "run-clean", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !finalized.OK || finalized.Receipt == nil {
		t.Fatalf("finalize failed: result=%#v err=%v", finalized, err)
	}
	if pathExists(target) || finalized.Receipt.Cleanup.RemovedBytes < 64 ||
		finalized.Receipt.Cleanup.Status != "done" {
		t.Fatalf("owned target was not safely removed: %#v", finalized.Receipt.Cleanup)
	}
	validated, err := executeCargoStorage(cargoStorageOptions{
		Action: "validate", RunID: "run-clean", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !validated.OK {
		t.Fatalf("cleanup receipt did not validate: result=%#v err=%v", validated, err)
	}
	repeated, err := executeCargoStorage(cargoStorageOptions{
		Action: "finalize", RunID: "run-clean", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !repeated.OK ||
		repeated.Receipt.Cleanup.RemovedBytes != finalized.Receipt.Cleanup.RemovedBytes {
		t.Fatalf("repeated finalization changed accounting: first=%#v repeated=%#v err=%v",
			finalized.Receipt.Cleanup, repeated.Receipt.Cleanup, err)
	}
}

func TestCargoStorageBlocksForgedOwnershipMarker(t *testing.T) {
	root := initWorktreeRepo(t)
	tempRoot := filepath.Join(t.TempDir(), "run-temp")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-forged", RepoRoot: root,
		CacheRoot: filepath.Join(t.TempDir(), "cache"), Now: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(tempRoot, "isolated-target")
	registered, err := executeCargoStorage(cargoStorageOptions{
		Action: "register-temp", RunID: "run-forged", RepoRoot: root,
		Target: target, TempRoot: tempRoot, Reason: "test", Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK || registered.Receipt == nil {
		t.Fatalf("register failed: result=%#v err=%v", registered, err)
	}
	markerPath := registered.Receipt.TemporaryTargets[0].MarkerPath
	content, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	var marker cargoStorageMarker
	if err := json.Unmarshal(content, &marker); err != nil {
		t.Fatal(err)
	}
	marker.OwnerToken = "forged"
	if err := writeAtomicJSON(markerPath, marker); err != nil {
		t.Fatal(err)
	}

	finalized, err := executeCargoStorage(cargoStorageOptions{
		Action: "finalize", RunID: "run-forged", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || finalized.OK || finalized.Receipt == nil {
		t.Fatalf("forged marker was accepted: result=%#v err=%v", finalized, err)
	}
	if !pathExists(target) || finalized.Receipt.Cleanup.Status != "blocked" {
		t.Fatalf("forged target was deleted: %#v", finalized.Receipt.Cleanup)
	}
}

func TestCargoStorageRejectsTemporaryTargetOutsideApprovedRoot(t *testing.T) {
	root := initWorktreeRepo(t)
	tempRoot := filepath.Join(t.TempDir(), "run-temp")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-outside", RepoRoot: root,
		CacheRoot: filepath.Join(t.TempDir(), "cache"), Now: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(filepath.Dir(tempRoot), "outside-target")
	result, err := executeCargoStorage(cargoStorageOptions{
		Action: "register-temp", RunID: "run-outside", RepoRoot: root,
		Target: target, TempRoot: tempRoot, Reason: "test", Now: time.Now().UTC(),
	})
	if err != nil || result.OK {
		t.Fatalf("outside target was accepted: result=%#v err=%v", result, err)
	}
	if pathExists(target) {
		t.Fatalf("outside target was created: %s", target)
	}
}

func TestCargoStorageRejectsSymlinkedTemporaryRoot(t *testing.T) {
	root := initWorktreeRepo(t)
	realRoot := filepath.Join(t.TempDir(), "real-root")
	if err := os.MkdirAll(realRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	linkRoot := filepath.Join(t.TempDir(), "linked-root")
	if err := os.Symlink(realRoot, linkRoot); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	_, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-linked-root", RepoRoot: root,
		CacheRoot: filepath.Join(t.TempDir(), "cache"), Now: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(linkRoot, "isolated-target")
	result, err := executeCargoStorage(cargoStorageOptions{
		Action: "register-temp", RunID: "run-linked-root", RepoRoot: root,
		Target: target, TempRoot: linkRoot, Reason: "test", Now: time.Now().UTC(),
	})
	if err != nil || result.OK {
		t.Fatalf("symlinked root was accepted: result=%#v err=%v", result, err)
	}
	if pathExists(filepath.Join(realRoot, "isolated-target")) {
		t.Fatal("target was created through symlinked root")
	}
}

func TestCargoStorageReceiptNamesDoNotCollideAfterSanitizing(t *testing.T) {
	root := initWorktreeRepo(t)
	first, err := cargoStorageReceiptPath(root, "feature/a")
	if err != nil {
		t.Fatal(err)
	}
	second, err := cargoStorageReceiptPath(root, "feature?a")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("distinct run IDs collided at %s", first)
	}
}

func TestCargoStorageConcurrentRegistrationsAreMerged(t *testing.T) {
	root := initWorktreeRepo(t)
	tempRoot := filepath.Join(t.TempDir(), "run-temp")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "run-concurrent", RepoRoot: root,
		CacheRoot: filepath.Join(t.TempDir(), "cache"), Now: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	results := make(chan cargoStorageResult, 2)
	errs := make(chan error, 2)
	for _, name := range []string{"target-a", "target-b"} {
		name := name
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, registerErr := executeCargoStorage(cargoStorageOptions{
				Action: "register-temp", RunID: "run-concurrent", RepoRoot: root,
				Target: filepath.Join(tempRoot, name), TempRoot: tempRoot,
				Reason: "concurrency test", Now: time.Now().UTC(),
			})
			results <- result
			errs <- registerErr
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for registerErr := range errs {
		if registerErr != nil {
			t.Fatal(registerErr)
		}
	}
	for result := range results {
		if !result.OK {
			t.Fatalf("concurrent registration failed: %#v", result)
		}
	}
	receiptPath, err := cargoStorageReceiptPath(root, "run-concurrent")
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := loadCargoStorageReceipt(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.TemporaryTargets) != 2 {
		t.Fatalf("concurrent registration was lost: %#v", receipt.TemporaryTargets)
	}
}

func TestCargoStoragePublicCLIFlow(t *testing.T) {
	root := initWorktreeRepo(t)
	cacheRoot := filepath.Join(t.TempDir(), "cache")
	runID := "run-cli"
	var stdout, stderr bytes.Buffer

	code := run([]string{"cargo-storage", "--action", "resolve", "--run-id", runID,
		"--root", root, "--cache-root", cacheRoot, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("resolve CLI failed: code=%d stderr=%s", code, stderr.String())
	}

	var result cargoStorageResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || !result.OK || result.Receipt == nil {
		t.Fatalf("resolve CLI returned invalid JSON: result=%#v err=%v", result, err)
	}

	stdout.Reset()
	stderr.Reset()
	code = run([]string{"cargo-storage", "--action", "validate-ready", "--run-id", runID,
		"--root", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("validate-ready CLI failed: code=%d stderr=%s", code, stderr.String())
	}

	tempRoot := filepath.Join(t.TempDir(), "run-temp")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"cargo-storage", "--action", "register-temp", "--run-id", runID,
		"--root", root, "--target", filepath.Join(filepath.Dir(tempRoot), "outside"),
		"--temp-root", tempRoot, "--reason", "CLI test", "--json"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("unsafe CLI registration was not blocked: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = run([]string{"cargo-storage", "--action", "finalize", "--run-id", runID,
		"--root", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("finalize CLI failed: code=%d stderr=%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"cargo-storage", "--action", "validate", "--run-id", runID,
		"--root", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("validate CLI failed: code=%d stderr=%s", code, stderr.String())
	}
}

func TestCargoStorageNotApplicableCLIFlow(t *testing.T) {
	root := initWorktreeRepo(t)
	runID := "run-no-cargo"
	var stdout, stderr bytes.Buffer
	code := run([]string{"cargo-storage", "--action", "not-applicable", "--run-id", runID,
		"--root", root, "--reason", "no Cargo command selected", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("not-applicable CLI failed: code=%d stderr=%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"cargo-storage", "--action", "validate", "--run-id", runID,
		"--root", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("not-applicable validation failed: code=%d stderr=%s", code, stderr.String())
	}
}

func TestCargoStorageFinalValidationRejectsTamperedCleanup(t *testing.T) {
	root := initWorktreeRepo(t)
	runID := "run-tampered"
	resolved, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: runID, RepoRoot: root,
		CacheRoot: filepath.Join(t.TempDir(), "cache"), Now: time.Now().UTC(),
	})
	if err != nil || !resolved.OK {
		t.Fatalf("resolve failed: result=%#v err=%v", resolved, err)
	}

	finalized, err := executeCargoStorage(cargoStorageOptions{
		Action: "finalize", RunID: runID, RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !finalized.OK {
		t.Fatalf("finalize failed: result=%#v err=%v", finalized, err)
	}
	receiptPath, err := cargoStorageReceiptPath(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := loadCargoStorageReceipt(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	receipt.AppliedEnvironment["CARGO_TARGET_DIR"] = filepath.Join(t.TempDir(), "wrong")
	if err := writeCargoStorageReceipt(receiptPath, receipt); err != nil {
		t.Fatal(err)
	}
	validated, err := executeCargoStorage(cargoStorageOptions{
		Action: "validate", RunID: runID, RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || validated.OK {
		t.Fatalf("tampered cleanup receipt was accepted: result=%#v err=%v", validated, err)
	}
}

func TestCargoStorageReconcilesDeletionIntentAfterTargetDisappears(t *testing.T) {
	root := initWorktreeRepo(t)
	tempRoot := filepath.Join(t.TempDir(), "run-temp")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	runID := "run-reconcile"
	_, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: runID, RepoRoot: root,
		CacheRoot: filepath.Join(t.TempDir(), "cache"), Now: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(tempRoot, "isolated-target")
	registered, err := executeCargoStorage(cargoStorageOptions{
		Action: "register-temp", RunID: runID, RepoRoot: root,
		Target: target, TempRoot: tempRoot, Reason: "test", Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK {
		t.Fatalf("register failed: result=%#v err=%v", registered, err)
	}
	writeFile(t, filepath.Join(target, "artifact.bin"), strings.Repeat("x", 32))
	receiptPath, err := cargoStorageReceiptPath(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := loadCargoStorageReceipt(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	size, issue := validateCargoTemporaryTarget(receipt.TemporaryTargets[0], runID)
	if issue != "" {
		t.Fatal(issue)
	}
	receipt.TemporaryTargets[0].RemovalStartedAt = time.Now().UTC().Format(time.RFC3339Nano)
	receipt.TemporaryTargets[0].RemovalBytes = size
	if err := writeCargoStorageReceipt(receiptPath, receipt); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(target); err != nil {
		t.Fatal(err)
	}

	finalized, err := executeCargoStorage(cargoStorageOptions{
		Action: "finalize", RunID: runID, RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !finalized.OK || finalized.Receipt.Cleanup.RemovedBytes != size {
		t.Fatalf("deletion intent was not reconciled: result=%#v err=%v", finalized, err)
	}
}

func TestCargoStorageFinalValidationRejectsFalseAccounting(t *testing.T) {
	root := initWorktreeRepo(t)
	runID := "run-false-accounting"
	resolved, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: runID, RepoRoot: root,
		CacheRoot: filepath.Join(t.TempDir(), "cache"), Now: time.Now().UTC(),
	})
	if err != nil || !resolved.OK {
		t.Fatalf("resolve failed: result=%#v err=%v", resolved, err)
	}
	finalized, err := executeCargoStorage(cargoStorageOptions{
		Action: "finalize", RunID: runID, RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !finalized.OK {
		t.Fatalf("finalize failed: result=%#v err=%v", finalized, err)
	}
	receiptPath, err := cargoStorageReceiptPath(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	original, err := loadCargoStorageReceipt(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func(*cargoStorageReceipt){
		"extra removed path": func(receipt *cargoStorageReceipt) {
			receipt.Cleanup.RemovedPaths = append(receipt.Cleanup.RemovedPaths, filepath.Join(t.TempDir(), "unregistered"))
		},
		"negative removed bytes": func(receipt *cargoStorageReceipt) {
			receipt.Cleanup.RemovedBytes = -1
		},
		"negative retained bytes": func(receipt *cargoStorageReceipt) {
			receipt.Cleanup.RetainedBytes = -1
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			receipt := original
			receipt.Cleanup.RemovedPaths = append([]string(nil), original.Cleanup.RemovedPaths...)
			mutate(&receipt)
			if err := writeCargoStorageReceipt(receiptPath, receipt); err != nil {
				t.Fatal(err)
			}
			validated, err := executeCargoStorage(cargoStorageOptions{
				Action: "validate", RunID: runID, RepoRoot: root, Now: time.Now().UTC(),
			})
			if err != nil || validated.OK {
				t.Fatalf("false accounting was accepted: result=%#v err=%v", validated, err)
			}
		})
	}
}
