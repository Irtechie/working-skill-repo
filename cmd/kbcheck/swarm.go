package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type manifestSlice struct {
	ID                     string
	Blockers               []string
	Status                 string
	Verification           string
	ModelTier              string
	ModelRoute             string
	ProofCheck             bool
	NoCheckReason          string
	ContextPacketPath      string
	NoPacketReason         string
	CanContinueOtherSlices bool
	HITL                   bool
}

type readySetResult struct {
	OK             bool     `json:"ok"`
	Reason         string   `json:"reason"`
	Ready          []string `json:"ready"`
	Runnable       []string `json:"runnable"`
	ExcludedSerial []string `json:"excluded_serial"`
}

type missingBlockerResult struct {
	OK              bool                   `json:"ok"`
	Reason          string                 `json:"reason"`
	MissingBlockers []missingBlockerRecord `json:"missing_blockers"`
	Ready           []string               `json:"ready"`
}

type missingBlockerRecord struct {
	Slice   string `json:"slice"`
	Blocker string `json:"blocker"`
}

type cycleResult struct {
	OK     bool     `json:"ok"`
	Reason string   `json:"reason"`
	Cycle  []string `json:"cycle"`
	Ready  []string `json:"ready"`
}

func runNativeSelftest(fn func(io.Writer, io.Writer) int) CheckResult {
	var out, err bytes.Buffer
	code := fn(&out, &err)
	return CheckResult{ExitCode: code, Stdout: out.String(), Stderr: err.String()}
}

func runReadySetCommand(root string, opts options, stdout, stderr io.Writer) int {
	path := resolveInputPath(root, opts.manifest)
	result, err := computeReadySet(path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeReadySet(stdout, result, opts.json)
}

func runReadySetSelftest(stdout, stderr io.Writer) int {
	temp, err := os.MkdirTemp("", "kb-work-ready-set-selftest-*")
	if err != nil {
		fmt.Fprintf(stderr, "create temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(temp)

	write := func(name, body string) string {
		path := filepath.Join(temp, name)
		_ = os.WriteFile(path, []byte(strings.TrimLeft(body, "\n")), 0o644)
		return path
	}
	assertReady := func(name string, result any, want []string) bool {
		ready, ok := result.(readySetResult)
		if !ok {
			fmt.Fprintf(stderr, "%s: expected ready result, got %#v\n", name, result)
			return false
		}
		if !equalStrings(ready.Ready, want) {
			fmt.Fprintf(stderr, "%s: ready=%v want=%v\n", name, ready.Ready, want)
			return false
		}
		return true
	}

	parallel := write("parallel.md", `
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
  - id: slice-003
    blockers: [slice-001]
    status: pending
    hitl: false
    can_continue_other_slices: true
---
`)
	result, err := computeReadySet(parallel)
	if err != nil || !assertReady("parallel", result, []string{"slice-001", "slice-002"}) {
		if err != nil {
			fmt.Fprintf(stderr, "parallel: %v\n", err)
		}
		return 1
	}

	serial := write("serial.md", `
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
	result, err = computeReadySet(serial)
	if err != nil || !assertReady("serial", result, []string{"slice-002"}) {
		if err != nil {
			fmt.Fprintf(stderr, "serial: %v\n", err)
		}
		return 1
	}

	singleSerial := write("single-serial.md", `
---
slices:
  - id: slice-001
    blockers: []
    status: pending
    hitl: false
    can_continue_other_slices: false
---
`)
	result, err = computeReadySet(singleSerial)
	if err != nil || !assertReady("single-serial", result, []string{"slice-001"}) {
		if err != nil {
			fmt.Fprintf(stderr, "single-serial: %v\n", err)
		}
		return 1
	}

	states := write("states.md", `
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
	result, err = computeReadySet(states)
	if err != nil || !assertReady("states", result, []string{"slice-004"}) {
		if err != nil {
			fmt.Fprintf(stderr, "states: %v\n", err)
		}
		return 1
	}

	cycle := write("cycle.md", `
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
		fmt.Fprintf(stderr, "cycle: %v\n", err)
		return 1
	}
	if value, ok := result.(cycleResult); !ok || value.OK {
		fmt.Fprintf(stderr, "cycle: expected failing cycle result, got %#v\n", result)
		return 1
	}

	fmt.Fprintln(stdout, "kb-work ready-set selftest: passed")
	return 0
}

func computeReadySet(path string) (any, error) {
	slices, err := parseManifestSlices(path)
	if err != nil {
		return nil, err
	}
	byID := map[string]manifestSlice{}
	ids := map[string]bool{}
	for _, slice := range slices {
		byID[slice.ID] = slice
		ids[slice.ID] = true
	}
	missing := []missingBlockerRecord{}
	for _, slice := range slices {
		for _, blocker := range slice.Blockers {
			if !ids[blocker] {
				missing = append(missing, missingBlockerRecord{Slice: slice.ID, Blocker: blocker})
			}
		}
	}
	if len(missing) > 0 {
		return missingBlockerResult{OK: false, Reason: "missing-blocker", MissingBlockers: missing, Ready: []string{}}, nil
	}
	if cycle := findManifestCycle(slices); len(cycle) > 0 {
		return cycleResult{OK: false, Reason: "cycle", Cycle: cycle, Ready: []string{}}, nil
	}

	doneStates := map[string]bool{"done": true, "skipped": true}
	terminalOrWaiting := map[string]bool{"done": true, "skipped": true, "blocked": true, "human-required": true, "parked": true, "in_progress": true}
	runnable := []manifestSlice{}
	for _, slice := range slices {
		if terminalOrWaiting[slice.Status] || slice.Status != "pending" {
			continue
		}
		ready := true
		for _, blocker := range slice.Blockers {
			if !doneStates[byID[blocker].Status] {
				ready = false
				break
			}
		}
		if ready {
			runnable = append(runnable, slice)
		}
	}

	ready := []manifestSlice{}
	excluded := []manifestSlice{}
	if len(runnable) == 1 {
		ready = runnable
	} else {
		for _, slice := range runnable {
			if slice.CanContinueOtherSlices && !slice.HITL {
				ready = append(ready, slice)
			} else {
				excluded = append(excluded, slice)
			}
		}
	}

	return readySetResult{
		OK:             true,
		Reason:         "ready",
		Ready:          sliceIDs(ready),
		Runnable:       sliceIDs(runnable),
		ExcludedSerial: sliceIDs(excluded),
	}, nil
}

func writeReadySet(stdout io.Writer, result any, asJSON bool) int {
	if asJSON {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(result)
	} else {
		if ready, ok := result.(readySetResult); ok {
			fmt.Fprint(stdout, strings.Join(ready.Ready, "\n"))
			if len(ready.Ready) > 0 {
				fmt.Fprintln(stdout)
			}
		} else {
			fmt.Fprintln(stdout, "no ready set")
		}
	}
	switch value := result.(type) {
	case readySetResult:
		if value.OK {
			return 0
		}
	default:
		return 2
	}
	return 2
}

func parseManifestSlices(path string) ([]manifestSlice, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	slices := []manifestSlice{}
	var current *manifestSlice
	for _, line := range strings.Split(string(content), "\n") {
		trimmedLeft := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmedLeft, "- id:") {
			if current != nil {
				slices = append(slices, *current)
			}
			current = &manifestSlice{
				ID:                     cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmedLeft, "- id:"))),
				Blockers:               []string{},
				CanContinueOtherSlices: true,
			}
			continue
		}
		if current == nil {
			continue
		}
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "blockers:"):
			current.Blockers = parseInlineList(strings.TrimSpace(strings.TrimPrefix(trimmed, "blockers:")))
		case strings.HasPrefix(trimmed, "status:"):
			current.Status = cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "status:")))
		case strings.HasPrefix(trimmed, "verification:"):
			current.Verification = cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "verification:")))
		case strings.HasPrefix(trimmed, "model_tier:"):
			current.ModelTier = cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "model_tier:")))
		case strings.HasPrefix(trimmed, "model_route:"):
			current.ModelRoute = cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "model_route:")))
		case strings.HasPrefix(trimmed, "proof_check:"):
			value := cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "proof_check:")))
			current.ProofCheck = value == "" || parseBool(value)
		case strings.HasPrefix(trimmed, "no_check_reason:"):
			current.NoCheckReason = cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "no_check_reason:")))
		case strings.HasPrefix(trimmed, "context_packet_path:"):
			current.ContextPacketPath = cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "context_packet_path:")))
		case strings.HasPrefix(trimmed, "no_packet_reason:"):
			current.NoPacketReason = cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "no_packet_reason:")))
		case strings.HasPrefix(trimmed, "can_continue_other_slices:"):
			current.CanContinueOtherSlices = parseBool(strings.TrimSpace(strings.TrimPrefix(trimmed, "can_continue_other_slices:")))
		case strings.HasPrefix(trimmed, "hitl:"):
			current.HITL = parseBool(strings.TrimSpace(strings.TrimPrefix(trimmed, "hitl:")))
		}
	}
	if current != nil {
		slices = append(slices, *current)
	}
	return slices, nil
}

func parseInlineList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "[]" {
		return []string{}
	}
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		value = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
		if value == "" {
			return []string{}
		}
		parts := strings.Split(value, ",")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			cleaned := cleanYAMLScalar(part)
			if cleaned != "" {
				out = append(out, cleaned)
			}
		}
		return out
	}
	return []string{cleanYAMLScalar(value)}
}

func cleanYAMLScalar(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return value
}

func parseBool(value string) bool {
	return strings.EqualFold(cleanYAMLScalar(value), "true")
}

func findManifestCycle(slices []manifestSlice) []string {
	byID := map[string]manifestSlice{}
	for _, slice := range slices {
		byID[slice.ID] = slice
	}
	visiting := map[string]bool{}
	visited := map[string]bool{}
	cycle := []string{}

	var visit func(string) bool
	visit = func(id string) bool {
		if visited[id] {
			return false
		}
		if visiting[id] {
			cycle = append(cycle, id)
			return true
		}
		slice, ok := byID[id]
		if !ok {
			return false
		}
		visiting[id] = true
		for _, blocker := range slice.Blockers {
			if visit(blocker) {
				cycle = append(cycle, id)
				return true
			}
		}
		delete(visiting, id)
		visited[id] = true
		return false
	}

	for _, slice := range slices {
		if visit(slice.ID) {
			reverseStrings(cycle)
			return cycle
		}
	}
	return nil
}

func sliceIDs(slices []manifestSlice) []string {
	ids := make([]string, 0, len(slices))
	for _, slice := range slices {
		ids = append(ids, slice.ID)
	}
	return ids
}

func reverseStrings(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

type scopeLeaseEntry struct {
	SliceID string `json:"slice_id"`
	Path    string `json:"path"`
	Status  string `json:"status"`
}

type scopeLeaseResult struct {
	OK           bool             `json:"ok"`
	ActiveLeases []activeLease    `json:"active_leases"`
	Collisions   []scopeCollision `json:"collisions"`
}

type activeLease struct {
	Path    string `json:"path"`
	SliceID string `json:"slice_id"`
}

type scopeCollision struct {
	Path      string `json:"path"`
	Owner     string `json:"owner"`
	Contender string `json:"contender"`
}

func runScopeLeaseCommand(root string, opts options, stdout, stderr io.Writer) int {
	path := resolveInputPath(root, opts.ledger)
	result, err := computeScopeLease(path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(result)
	} else if result.OK {
		fmt.Fprintln(stdout, "scope lease: ok")
	} else {
		fmt.Fprintln(stdout, "scope lease: collision")
	}
	if !result.OK {
		return 2
	}
	return 0
}

func runScopeLeaseSelftest(stdout, stderr io.Writer) int {
	temp, err := os.MkdirTemp("", "kb-work-scope-lease-selftest-*")
	if err != nil {
		fmt.Fprintf(stderr, "create temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(temp)

	write := func(name string, entries []scopeLeaseEntry) string {
		path := filepath.Join(temp, name)
		content, _ := json.MarshalIndent(entries, "", "  ")
		_ = os.WriteFile(path, content, 0o644)
		return path
	}

	disjoint, err := computeScopeLease(write("disjoint.json", []scopeLeaseEntry{
		{SliceID: "slice-001", Path: "src/a.ts", Status: "active"},
		{SliceID: "slice-002", Path: "src/b.ts", Status: "active"},
	}))
	if err != nil || !disjoint.OK || len(disjoint.ActiveLeases) != 2 {
		fmt.Fprintf(stderr, "disjoint failed: result=%#v err=%v\n", disjoint, err)
		return 1
	}

	collision, err := computeScopeLease(write("collision.json", []scopeLeaseEntry{
		{SliceID: "slice-001", Path: "src/shared.ts", Status: "active"},
		{SliceID: "slice-002", Path: "src/shared.ts", Status: "active"},
	}))
	if err != nil || collision.OK || len(collision.Collisions) != 1 {
		fmt.Fprintf(stderr, "collision failed: result=%#v err=%v\n", collision, err)
		return 1
	}

	released, err := computeScopeLease(write("released.json", []scopeLeaseEntry{
		{SliceID: "slice-001", Path: "src/shared.ts", Status: "active"},
		{SliceID: "slice-001", Path: "src/shared.ts", Status: "done"},
		{SliceID: "slice-002", Path: "src/shared.ts", Status: "active"},
	}))
	if err != nil || !released.OK || len(released.ActiveLeases) != 1 || released.ActiveLeases[0].SliceID != "slice-002" {
		fmt.Fprintf(stderr, "released failed: result=%#v err=%v\n", released, err)
		return 1
	}

	requeued, err := computeScopeLease(write("requeued.json", []scopeLeaseEntry{
		{SliceID: "slice-001", Path: "src/shared.ts", Status: "claimed"},
		{SliceID: "slice-001", Path: "src/shared.ts", Status: "requeued"},
		{SliceID: "slice-002", Path: "src/shared.ts", Status: "writing"},
	}))
	if err != nil || !requeued.OK || len(requeued.ActiveLeases) != 1 || requeued.ActiveLeases[0].SliceID != "slice-002" {
		fmt.Fprintf(stderr, "requeued failed: result=%#v err=%v\n", requeued, err)
		return 1
	}

	fmt.Fprintln(stdout, "kb-work scope-lease selftest: passed")
	return 0
}

func computeScopeLease(path string) (scopeLeaseResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return scopeLeaseResult{}, fmt.Errorf("read ledger: %w", err)
	}
	var entries []scopeLeaseEntry
	if err := json.Unmarshal(content, &entries); err != nil {
		return scopeLeaseResult{}, fmt.Errorf("parse ledger JSON: %w", err)
	}

	activeStates := map[string]bool{"active": true, "claimed": true, "writing": true}
	releaseStates := map[string]bool{"done": true, "skipped": true, "requeued": true, "released": true}
	owners := map[string]string{}
	collisions := []scopeCollision{}

	for _, entry := range entries {
		if entry.SliceID == "" || entry.Path == "" || entry.Status == "" {
			return scopeLeaseResult{}, fmt.Errorf("ledger entry missing slice_id, path, or status")
		}
		pathKey := normalizeLeasePath(entry.Path)
		status := strings.ToLower(entry.Status)
		if releaseStates[status] {
			if owners[pathKey] == entry.SliceID {
				delete(owners, pathKey)
			}
			continue
		}
		if !activeStates[status] {
			continue
		}
		if owner, ok := owners[pathKey]; ok && owner != entry.SliceID {
			collisions = append(collisions, scopeCollision{Path: pathKey, Owner: owner, Contender: entry.SliceID})
			continue
		}
		owners[pathKey] = entry.SliceID
	}

	active := make([]activeLease, 0, len(owners))
	for path, sliceID := range owners {
		active = append(active, activeLease{Path: path, SliceID: sliceID})
	}
	sort.Slice(active, func(i, j int) bool { return active[i].Path < active[j].Path })
	return scopeLeaseResult{OK: len(collisions) == 0, ActiveLeases: active, Collisions: collisions}, nil
}

func normalizeLeasePath(path string) string {
	return strings.ToLower(strings.ReplaceAll(filepath.ToSlash(path), "\\", "/"))
}

func resolveInputPath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
