package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

const cargoStorageSchemaVersion = 1
const cargoStorageLockTimeout = 5 * time.Second

type cargoStorageOptions struct {
	Action    string
	RunID     string
	RepoRoot  string
	CacheRoot string
	Target    string
	TempRoot  string
	Reason    string
	Now       time.Time
}

type cargoStorageReceipt struct {
	SchemaVersion          int                  `json:"schema_version"`
	RunID                  string               `json:"run_id"`
	RepoIdentity           string               `json:"repo_identity"`
	StableTarget           string               `json:"stable_target"`
	StableSource           string               `json:"stable_source"`
	CargoConfigFingerprint string               `json:"cargo_config_fingerprint"`
	AppliedEnvironment     map[string]string    `json:"applied_environment"`
	TemporaryTargets       []cargoTemporaryPath `json:"temporary_targets"`
	Cleanup                cargoStorageCleanup  `json:"cleanup"`
	ReceiptPath            string               `json:"receipt_path"`
	NotApplicableReason    string               `json:"not_applicable_reason,omitempty"`
	CreatedAt              string               `json:"created_at"`
	UpdatedAt              string               `json:"updated_at"`
}

type cargoTemporaryPath struct {
	Path             string `json:"path"`
	TempRoot         string `json:"temp_root"`
	Reason           string `json:"reason"`
	OwnerToken       string `json:"owner_token"`
	MarkerPath       string `json:"marker_path"`
	CreatedAt        string `json:"created_at"`
	RemovalStartedAt string `json:"removal_started_at,omitempty"`
	RemovalBytes     int64  `json:"removal_bytes,omitempty"`
	RemovedAt        string `json:"removed_at,omitempty"`
}

type cargoStorageCleanup struct {
	Status        string   `json:"status"`
	StableTargets []string `json:"stable_targets"`
	RemovedPaths  []string `json:"temporary_targets_removed"`
	RetainedBytes int64    `json:"retained_bytes"`
	RemovedBytes  int64    `json:"removed_bytes"`
	Unresolved    []string `json:"unresolved"`
	CompletedAt   string   `json:"completed_at,omitempty"`
}

type cargoStorageResult struct {
	OK      bool                 `json:"ok"`
	Action  string               `json:"action"`
	Issue   string               `json:"issue,omitempty"`
	Receipt *cargoStorageReceipt `json:"receipt,omitempty"`
}

type cargoStorageMarker struct {
	SchemaVersion int    `json:"schema_version"`
	RunID         string `json:"run_id"`
	OwnerToken    string `json:"owner_token"`
	Target        string `json:"target"`
	TempRoot      string `json:"temp_root"`
}

func runCargoStorageCommand(root string, opts options, stdout, stderr io.Writer) int {
	result, err := executeCargoStorage(cargoStorageOptions{
		Action: opts.sliceLeaseAction, RunID: opts.runID, RepoRoot: root,
		CacheRoot: opts.cargoCacheRoot, Target: opts.cargoTarget,
		TempRoot: opts.cargoTempRoot, Reason: opts.cargoReason, Now: time.Now().UTC(),
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else if result.OK && result.Receipt != nil {
		fmt.Fprintf(stdout, "cargo-storage: %s run=%s target=%s receipt=%s\n",
			result.Action, result.Receipt.RunID, result.Receipt.StableTarget, result.Receipt.ReceiptPath)
	} else {
		fmt.Fprintf(stdout, "cargo-storage: %s blocked: %s\n", result.Action, result.Issue)
	}
	if !result.OK {
		return 2
	}
	return 0
}

func executeCargoStorage(opts cargoStorageOptions) (cargoStorageResult, error) {
	action := strings.ToLower(strings.TrimSpace(opts.Action))
	if action != "resolve" && action != "register-temp" && action != "finalize" &&
		action != "validate-ready" && action != "not-applicable" && action != "validate" {
		return cargoStorageResult{}, fmt.Errorf("cargo-storage action must be resolve, register-temp, finalize, validate-ready, not-applicable, or validate")
	}
	if strings.TrimSpace(opts.RunID) == "" {
		return cargoStorageResult{}, fmt.Errorf("cargo-storage requires run-id")
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}
	root, err := filepath.Abs(filepath.Clean(opts.RepoRoot))
	if err != nil {
		return cargoStorageResult{}, err
	}
	opts.RepoRoot = root
	receiptPath, err := cargoStorageReceiptPath(root, opts.RunID)
	if err != nil {
		return cargoStorageResult{}, err
	}
	lockName := fmt.Sprintf("cargo-storage-%s.lock", cargoStorageProjectKey(opts.RunID)[:16])
	lock, err := modelrouting.AcquireSharedProjectLock(filepath.Dir(receiptPath), lockName, cargoStorageLockTimeout)
	if err != nil {
		return cargoStorageResult{}, fmt.Errorf("acquire Cargo storage receipt lock: %w", err)
	}
	defer lock.Close()
	return executeCargoStorageLocked(opts, action, receiptPath)
}

func executeCargoStorageLocked(opts cargoStorageOptions, action, receiptPath string) (cargoStorageResult, error) {
	if action == "resolve" {
		receipt, err := resolveCargoStorage(opts, receiptPath)
		if err != nil {
			return cargoStorageResult{}, err
		}
		return cargoStorageResult{OK: true, Action: action, Receipt: &receipt}, nil
	}
	if action == "not-applicable" {
		receipt, err := recordCargoStorageNotApplicable(opts, receiptPath)
		if err != nil {
			return cargoStorageResult{}, err
		}
		return cargoStorageResult{OK: true, Action: action, Receipt: &receipt}, nil
	}
	receipt, err := loadCargoStorageReceipt(receiptPath)
	if err != nil {
		return cargoStorageResult{Action: action, Issue: err.Error()}, nil
	}
	identityMatches, err := cargoStorageReceiptIdentityMatches(opts.RepoRoot, receiptPath, receipt.RepoIdentity)
	if err != nil {
		return cargoStorageResult{}, err
	}
	if receipt.RunID != opts.RunID || !identityMatches {
		return cargoStorageResult{Action: action, Issue: "receipt identity does not match the active run and repository", Receipt: &receipt}, nil
	}
	switch action {
	case "register-temp":
		return registerCargoTemporaryTarget(opts, receiptPath, receipt)
	case "finalize":
		return finalizeCargoStorage(opts, receiptPath, receipt)
	case "validate-ready":
		return validateCargoStorageReady(opts, receipt)
	default:
		return validateCargoStorageFinal(opts, receiptPath, receipt)
	}
}

func validateCargoStorageReady(opts cargoStorageOptions, receipt cargoStorageReceipt) (cargoStorageResult, error) {
	result := cargoStorageResult{Action: "validate-ready", Receipt: &receipt}
	if receipt.Cleanup.Status != "pending" {
		result.Issue = "build-storage receipt is not open for Cargo execution"
		return result, nil
	}
	if !filepath.IsAbs(receipt.StableTarget) ||
		receipt.AppliedEnvironment["CARGO_TARGET_DIR"] != receipt.StableTarget {
		result.Issue = "build-storage receipt does not enforce one absolute CARGO_TARGET_DIR"
		return result, nil
	}
	info, err := os.Lstat(receipt.StableTarget)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		result.Issue = "stable Cargo target is unavailable or unsafe"
		return result, nil
	}
	if receipt.CargoConfigFingerprint == "" ||
		receipt.CargoConfigFingerprint != cargoConfigFingerprint(opts.RepoRoot) {
		result.Issue = "Cargo configuration fingerprint is missing or stale"
		return result, nil
	}
	result.OK = true
	return result, nil
}

func validateCargoStorageFinal(opts cargoStorageOptions, receiptPath string, receipt cargoStorageReceipt) (cargoStorageResult, error) {
	result := cargoStorageResult{Action: "validate", Receipt: &receipt}
	if filepath.Clean(receipt.ReceiptPath) != filepath.Clean(receiptPath) {
		result.Issue = "build-storage receipt path does not match its canonical location"
		return result, nil
	}
	if receipt.Cleanup.Status == "not-applicable" {
		if strings.TrimSpace(receipt.NotApplicableReason) == "" || receipt.StableTarget != "" ||
			len(receipt.TemporaryTargets) != 0 || len(receipt.Cleanup.StableTargets) != 0 ||
			len(receipt.Cleanup.RemovedPaths) != 0 || len(receipt.Cleanup.Unresolved) != 0 ||
			receipt.Cleanup.RetainedBytes != 0 || receipt.Cleanup.RemovedBytes != 0 {
			result.Issue = "not-applicable receipt is incomplete or contains Cargo targets"
			return result, nil
		}
		result.OK = true
		return result, nil
	}
	if receipt.Cleanup.Status != "done" || len(receipt.Cleanup.Unresolved) != 0 {
		result.Issue = "build-storage cleanup receipt is not complete"
		return result, nil
	}
	if receipt.Cleanup.RetainedBytes < 0 || receipt.Cleanup.RemovedBytes < 0 ||
		len(receipt.Cleanup.StableTargets) != 1 ||
		filepath.Clean(receipt.Cleanup.StableTargets[0]) != filepath.Clean(receipt.StableTarget) ||
		len(receipt.Cleanup.RemovedPaths) != len(receipt.TemporaryTargets) {
		result.Issue = "build-storage cleanup accounting is inconsistent"
		return result, nil
	}
	if !filepath.IsAbs(receipt.StableTarget) ||
		receipt.AppliedEnvironment["CARGO_TARGET_DIR"] != receipt.StableTarget {
		result.Issue = "build-storage receipt does not preserve its absolute stable target"
		return result, nil
	}
	info, err := os.Lstat(receipt.StableTarget)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		result.Issue = "stable Cargo target is unavailable or unsafe"
		return result, nil
	}
	if receipt.CargoConfigFingerprint == "" ||
		receipt.CargoConfigFingerprint != cargoConfigFingerprint(opts.RepoRoot) {
		result.Issue = "Cargo configuration fingerprint is missing or stale"
		return result, nil
	}
	removed := make(map[string]bool, len(receipt.Cleanup.RemovedPaths))
	for _, path := range receipt.Cleanup.RemovedPaths {
		clean := filepath.Clean(path)
		if removed[clean] {
			result.Issue = "build-storage cleanup contains duplicate removal paths"
			return result, nil
		}
		removed[clean] = true
	}
	var removedBytes int64
	for _, entry := range receipt.TemporaryTargets {
		if entry.RemovalStartedAt == "" || entry.RemovalBytes < 0 || entry.RemovedAt == "" ||
			!removed[filepath.Clean(entry.Path)] || pathExists(entry.Path) {
			result.Issue = "temporary Cargo target cleanup is incomplete"
			return result, nil
		}
		removedBytes += entry.RemovalBytes
	}
	if removedBytes != receipt.Cleanup.RemovedBytes {
		result.Issue = "temporary Cargo target byte accounting is inconsistent"
		return result, nil
	}
	result.OK = true
	return result, nil
}

func resolveCargoStorage(opts cargoStorageOptions, receiptPath string) (cargoStorageReceipt, error) {
	identity, err := cargoStorageRepositoryIdentity(opts.RepoRoot)
	if err != nil {
		return cargoStorageReceipt{}, err
	}
	if _, err := os.Stat(receiptPath); err == nil {
		existing, loadErr := loadCargoStorageReceipt(receiptPath)
		if loadErr != nil {
			return cargoStorageReceipt{}, loadErr
		}
		identityMatches, matchErr := cargoStorageReceiptIdentityMatches(opts.RepoRoot, receiptPath, existing.RepoIdentity)
		if matchErr != nil {
			return cargoStorageReceipt{}, matchErr
		}
		if existing.RunID != opts.RunID || !identityMatches {
			return cargoStorageReceipt{}, fmt.Errorf("existing Cargo storage receipt conflicts with the active run or repository")
		}
		if existing.Cleanup.Status == "pending" {
			existing.RepoIdentity = identity
			existing.CargoConfigFingerprint = cargoConfigFingerprint(opts.RepoRoot)
			existing.UpdatedAt = opts.Now.Format(time.RFC3339Nano)
			if err := writeCargoStorageReceipt(receiptPath, existing); err != nil {
				return cargoStorageReceipt{}, err
			}
		}
		return existing, nil
	} else if !os.IsNotExist(err) {
		return cargoStorageReceipt{}, fmt.Errorf("inspect Cargo storage receipt: %w", err)
	}
	target, source, err := selectStableCargoTarget(opts, identity)
	if err != nil {
		return cargoStorageReceipt{}, err
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return cargoStorageReceipt{}, fmt.Errorf("create stable Cargo target: %w", err)
	}
	now := opts.Now.Format(time.RFC3339Nano)
	receipt := cargoStorageReceipt{
		SchemaVersion: cargoStorageSchemaVersion, RunID: opts.RunID,
		RepoIdentity: identity, StableTarget: target, StableSource: source,
		CargoConfigFingerprint: cargoConfigFingerprint(opts.RepoRoot),
		AppliedEnvironment:     map[string]string{"CARGO_TARGET_DIR": target},
		Cleanup:                cargoStorageCleanup{Status: "pending", StableTargets: []string{target}},
		ReceiptPath:            receiptPath, CreatedAt: now, UpdatedAt: now,
	}
	return receipt, writeCargoStorageReceipt(receiptPath, receipt)
}

func recordCargoStorageNotApplicable(opts cargoStorageOptions, receiptPath string) (cargoStorageReceipt, error) {
	if strings.TrimSpace(opts.Reason) == "" {
		return cargoStorageReceipt{}, fmt.Errorf("not-applicable requires reason")
	}
	if _, err := os.Stat(receiptPath); err == nil {
		return cargoStorageReceipt{}, fmt.Errorf("Cargo storage receipt already exists for this run")
	} else if !os.IsNotExist(err) {
		return cargoStorageReceipt{}, err
	}
	identity, err := cargoStorageRepositoryIdentity(opts.RepoRoot)
	if err != nil {
		return cargoStorageReceipt{}, err
	}
	now := opts.Now.Format(time.RFC3339Nano)
	receipt := cargoStorageReceipt{
		SchemaVersion: cargoStorageSchemaVersion,
		RunID:         opts.RunID,
		RepoIdentity:  identity,
		StableSource:  "not-applicable",
		Cleanup: cargoStorageCleanup{
			Status: "not-applicable",
		},
		ReceiptPath:         receiptPath,
		NotApplicableReason: strings.TrimSpace(opts.Reason),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	return receipt, writeCargoStorageReceipt(receiptPath, receipt)
}

func selectStableCargoTarget(opts cargoStorageOptions, identity string) (string, string, error) {
	key := cargoStorageProjectKey(identity)
	if configured := strings.TrimSpace(os.Getenv("CARGO_TARGET_DIR")); configured != "" && filepath.IsAbs(configured) {
		root, err := canonicalPath(configured)
		if err != nil {
			return "", "", err
		}
		if !cargoTargetInWorktree(opts.RepoRoot, root) {
			if filepath.Base(root) == key {
				return root, "environment-project-target", nil
			}
			return filepath.Join(root, key), "environment-cache-root", nil
		}
	}
	cacheRoot := strings.TrimSpace(opts.CacheRoot)
	source := "cache-root"
	if cacheRoot == "" {
		cacheRoot = strings.TrimSpace(os.Getenv("KB_CARGO_TARGET_ROOT"))
		source = "KB_CARGO_TARGET_ROOT"
	}
	if cacheRoot == "" {
		userCache, err := os.UserCacheDir()
		if err != nil {
			return "", "", fmt.Errorf("resolve user cache directory: %w", err)
		}
		cacheRoot = filepath.Join(userCache, "kb", "cargo-target")
		source = "user-cache"
	}
	cacheRoot, err := canonicalPath(cacheRoot)
	if err != nil {
		return "", "", err
	}
	target := filepath.Join(cacheRoot, key)
	return target, source, nil
}

func cargoStorageProjectKey(identity string) string {
	sum := sha256.Sum256([]byte(identity))
	return hex.EncodeToString(sum[:12])
}

func registerCargoTemporaryTarget(opts cargoStorageOptions, receiptPath string, receipt cargoStorageReceipt) (cargoStorageResult, error) {
	if strings.TrimSpace(opts.Target) == "" || strings.TrimSpace(opts.TempRoot) == "" || strings.TrimSpace(opts.Reason) == "" {
		return cargoStorageResult{}, fmt.Errorf("register-temp requires target, temp-root, and reason")
	}
	ready, err := validateCargoStorageReady(opts, receipt)
	if err != nil {
		return cargoStorageResult{}, err
	}
	if !ready.OK {
		return cargoStorageResult{
			Action:  "register-temp",
			Issue:   "build-storage receipt is not execution-ready: " + ready.Issue,
			Receipt: &receipt,
		}, nil
	}
	target, err := canonicalPath(opts.Target)
	if err != nil {
		return cargoStorageResult{}, err
	}
	tempRoot, err := canonicalPath(opts.TempRoot)
	if err != nil {
		return cargoStorageResult{}, err
	}
	rootInfo, err := os.Lstat(tempRoot)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return cargoStorageResult{Action: "register-temp", Issue: "approved temporary root must be an existing real directory", Receipt: &receipt}, nil
	}
	realRoot, err := filepath.EvalSymlinks(tempRoot)
	if err != nil || filepath.Clean(realRoot) != filepath.Clean(tempRoot) {
		return cargoStorageResult{Action: "register-temp", Issue: "approved temporary root must not traverse a link", Receipt: &receipt}, nil
	}
	if filepath.Dir(target) != tempRoot || !cargoPathWithin(target, tempRoot) {
		return cargoStorageResult{Action: "register-temp", Issue: "temporary target must be a direct child of the approved temporary root", Receipt: &receipt}, nil
	}
	if cargoTemporaryTargetNameForbidden(target) {
		return cargoStorageResult{Action: "register-temp", Issue: "phase-, worker-, slice-, and run-specific Cargo targets are prohibited", Receipt: &receipt}, nil
	}
	if _, err := os.Lstat(target); err == nil {
		return cargoStorageResult{Action: "register-temp", Issue: "temporary target must not already exist", Receipt: &receipt}, nil
	} else if !os.IsNotExist(err) {
		return cargoStorageResult{}, err
	}
	if err := os.Mkdir(target, 0o755); err != nil {
		return cargoStorageResult{}, err
	}
	token, err := randomCargoStorageToken()
	if err != nil {
		return cargoStorageResult{}, err
	}
	markerPath := filepath.Join(target, ".kb-cargo-owner.json")
	marker := cargoStorageMarker{SchemaVersion: cargoStorageSchemaVersion, RunID: opts.RunID, OwnerToken: token, Target: target, TempRoot: tempRoot}
	if err := writeAtomicJSON(markerPath, marker); err != nil {
		return cargoStorageResult{}, err
	}
	receipt.TemporaryTargets = append(receipt.TemporaryTargets, cargoTemporaryPath{
		Path: target, TempRoot: tempRoot, Reason: opts.Reason, OwnerToken: token,
		MarkerPath: markerPath, CreatedAt: opts.Now.Format(time.RFC3339Nano),
	})
	receipt.Cleanup.Status = "pending"
	receipt.UpdatedAt = opts.Now.Format(time.RFC3339Nano)
	if err := writeCargoStorageReceipt(receiptPath, receipt); err != nil {
		return cargoStorageResult{}, err
	}
	return cargoStorageResult{OK: true, Action: "register-temp", Receipt: &receipt}, nil
}

func finalizeCargoStorage(opts cargoStorageOptions, receiptPath string, receipt cargoStorageReceipt) (cargoStorageResult, error) {
	if receipt.Cleanup.Status == "done" {
		result, err := validateCargoStorageFinal(opts, receiptPath, receipt)
		result.Action = "finalize"
		return result, err
	}
	cleanup := cargoStorageCleanup{Status: "done", StableTargets: []string{receipt.StableTarget}}
	cleanup.RetainedBytes, _ = pathSize(receipt.StableTarget)
	for i := range receipt.TemporaryTargets {
		entry := &receipt.TemporaryTargets[i]
		if entry.RemovedAt != "" {
			cleanup.RemovedPaths = append(cleanup.RemovedPaths, entry.Path)
			cleanup.RemovedBytes += entry.RemovalBytes
			continue
		}
		if entry.RemovalStartedAt != "" && !pathExists(entry.Path) {
			entry.RemovedAt = opts.Now.Format(time.RFC3339Nano)
			cleanup.RemovedPaths = append(cleanup.RemovedPaths, entry.Path)
			cleanup.RemovedBytes += entry.RemovalBytes
			continue
		}
		size, issue := validateCargoTemporaryTarget(*entry, receipt.RunID)
		if issue != "" {
			cleanup.Unresolved = append(cleanup.Unresolved, entry.Path+": "+issue)
			continue
		}
		if entry.RemovalStartedAt == "" {
			entry.RemovalStartedAt = opts.Now.Format(time.RFC3339Nano)
			entry.RemovalBytes = size
			receipt.UpdatedAt = entry.RemovalStartedAt
			if err := writeCargoStorageReceipt(receiptPath, receipt); err != nil {
				return cargoStorageResult{}, fmt.Errorf("persist Cargo target deletion intent: %w", err)
			}
		}
		if err := os.RemoveAll(entry.Path); err != nil {
			cleanup.Unresolved = append(cleanup.Unresolved, entry.Path+": "+err.Error())
			continue
		}
		entry.RemovedAt = opts.Now.Format(time.RFC3339Nano)
		cleanup.RemovedPaths = append(cleanup.RemovedPaths, entry.Path)
		cleanup.RemovedBytes += entry.RemovalBytes
	}
	if len(cleanup.Unresolved) > 0 {
		cleanup.Status = "blocked"
	}
	cleanup.CompletedAt = opts.Now.Format(time.RFC3339Nano)
	receipt.Cleanup = cleanup
	receipt.UpdatedAt = cleanup.CompletedAt
	if err := writeCargoStorageReceipt(receiptPath, receipt); err != nil {
		return cargoStorageResult{}, err
	}
	result := cargoStorageResult{OK: cleanup.Status == "done", Action: "finalize", Receipt: &receipt}
	if !result.OK {
		result.Issue = "one or more temporary Cargo targets failed ownership or containment validation"
	}
	return result, nil
}

func validateCargoTemporaryTarget(entry cargoTemporaryPath, runID string) (int64, string) {
	info, err := os.Lstat(entry.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, "target is missing"
		}
		return 0, err.Error()
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return 0, "target is not a real directory"
	}
	realTarget, err := filepath.EvalSymlinks(entry.Path)
	if err != nil {
		return 0, "resolve target: " + err.Error()
	}
	realRoot, err := filepath.EvalSymlinks(entry.TempRoot)
	if err != nil {
		return 0, "resolve temp root: " + err.Error()
	}
	if filepath.Clean(realTarget) == filepath.Clean(realRoot) || !cargoPathWithin(realTarget, realRoot) {
		return 0, "target escapes approved temporary root"
	}
	var marker cargoStorageMarker
	expectedMarker := filepath.Join(entry.Path, ".kb-cargo-owner.json")
	if filepath.Clean(entry.MarkerPath) != filepath.Clean(expectedMarker) {
		return 0, "ownership marker path mismatch"
	}
	content, err := os.ReadFile(entry.MarkerPath)
	if err != nil {
		return 0, "ownership marker unavailable"
	}
	if err := json.Unmarshal(content, &marker); err != nil {
		return 0, "ownership marker invalid"
	}
	if marker.RunID != runID || marker.OwnerToken != entry.OwnerToken ||
		filepath.Clean(marker.Target) != filepath.Clean(entry.Path) ||
		filepath.Clean(marker.TempRoot) != filepath.Clean(entry.TempRoot) {
		return 0, "ownership marker mismatch"
	}
	size, err := pathSize(entry.Path)
	if err != nil {
		return 0, "measure target: " + err.Error()
	}
	return size, ""
}

func cargoStorageReceiptPath(root, runID string) (string, error) {
	common, err := gitCommonDir(root)
	if err != nil {
		return "", err
	}
	safeRunID := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '-'
	}, runID)
	if safeRunID == "" || safeRunID == "." || safeRunID == ".." {
		return "", fmt.Errorf("run-id has no safe filename characters")
	}
	runHash := sha256.Sum256([]byte(runID))
	filename := fmt.Sprintf("%s-%s.json", safeRunID, hex.EncodeToString(runHash[:8]))
	return filepath.Join(common, "kb", "cargo-storage", filename), nil
}

func cargoStorageRepositoryIdentity(root string) (string, error) {
	common, err := gitCommonDir(root)
	if err != nil {
		return "", err
	}
	identity, err := canonicalPath(common)
	if err != nil {
		return "", fmt.Errorf("canonicalize Git common directory: %w", err)
	}
	return filepath.Clean(identity), nil
}

func cargoStorageReceiptIdentityMatches(root, receiptPath, receiptIdentity string) (bool, error) {
	identity, err := cargoStorageRepositoryIdentity(root)
	if err != nil {
		return false, err
	}
	if receiptIdentity == identity {
		return true, nil
	}
	legacyStateRoot := filepath.Dir(receiptPath)
	return receiptIdentity == repoIdentity(root, legacyStateRoot) ||
		receiptIdentity == "state-root:"+filepath.ToSlash(legacyStateRoot), nil
}

func cargoTemporaryTargetNameForbidden(path string) bool {
	name := strings.ToLower(filepath.Base(filepath.Clean(path)))
	tokens := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})
	hasPhase := false
	for _, token := range tokens {
		switch token {
		case "audit", "bench", "benchmark", "build", "check", "clippy",
			"compile", "complete", "coverage", "debug", "diagnostic", "e2e",
			"finalize", "fix", "integration", "lint", "package", "probe",
			"release", "repair", "repro", "reproduction", "retry", "smoke",
			"test", "troubleshoot", "unit", "verification", "verify", "work",
			"worker", "slice", "run":
			hasPhase = true
		default:
			for _, prefix := range []string{"worker", "slice", "run"} {
				suffix := strings.TrimPrefix(token, prefix)
				if suffix != token && suffix != "" && allDecimalDigits(suffix) {
					hasPhase = true
				}
			}
		}
	}
	return hasPhase
}

func allDecimalDigits(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return value != ""
}

func loadCargoStorageReceipt(path string) (cargoStorageReceipt, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return cargoStorageReceipt{}, fmt.Errorf("read Cargo storage receipt: %w", err)
	}
	var receipt cargoStorageReceipt
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return cargoStorageReceipt{}, fmt.Errorf("decode Cargo storage receipt: %w", err)
	}
	if receipt.SchemaVersion != cargoStorageSchemaVersion {
		return cargoStorageReceipt{}, fmt.Errorf("unsupported Cargo storage receipt schema %d", receipt.SchemaVersion)
	}
	return receipt, nil
}

func writeCargoStorageReceipt(path string, receipt cargoStorageReceipt) error {
	receipt.ReceiptPath = path
	return writeAtomicJSON(path, receipt)
}

func writeAtomicJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	temp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if _, err := temp.Write(content); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, path)
}

func canonicalPath(path string) (string, error) {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	if real, err := filepath.EvalSymlinks(absolute); err == nil {
		return filepath.Abs(filepath.Clean(real))
	}
	return absolute, nil
}

func cargoPathWithin(path, root string) bool {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	return err == nil && relative != "." && relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}

func cargoTargetInWorktree(root, target string) bool {
	output := gitOutput(root, "worktree", "list", "--porcelain")
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "worktree ") {
			continue
		}
		worktree := strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
		if filepath.Clean(target) == filepath.Clean(worktree) || cargoPathWithin(target, worktree) {
			return true
		}
	}
	return filepath.Clean(target) == filepath.Clean(root) || cargoPathWithin(target, root)
}

func cargoConfigFingerprint(root string) string {
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(root, ".cargo", "config"),
		filepath.Join(root, ".cargo", "config.toml"),
		filepath.Join(home, ".cargo", "config"),
		filepath.Join(home, ".cargo", "config.toml"),
	}
	hash := sha256.New()
	for _, path := range paths {
		if content, err := os.ReadFile(path); err == nil {
			_, _ = hash.Write([]byte(filepath.Clean(path)))
			_, _ = hash.Write(content)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func pathSize(root string) (int64, error) {
	var total int64
	err := filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

func randomCargoStorageToken() (string, error) {
	value := make([]byte, 24)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
