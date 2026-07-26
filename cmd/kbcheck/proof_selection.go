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
	"time"
)

type proofGovernorRegistry struct {
	SchemaVersion int                      `json:"schema_version"`
	Checks        []proofGovernorCheckSpec `json:"checks"`
}

func runProofGovernorPlanCommand(root string, opts options, stdout, stderr io.Writer) int {
	registryPath := resolveInputPath(root, opts.proofRegistryPath)
	registry, err := loadProofGovernorRegistry(registryPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	receipts, receiptIssues := loadProofGovernorReceipts(resolveInputPath(root, opts.receiptDir))
	plan := selectProofGovernorPlan(root, registry, strings.Split(opts.proofRequest, ","), receipts, time.Now().UTC())
	for _, issue := range receiptIssues {
		plan.Reasons = append(plan.Reasons, "receipt-load:"+issue)
		if plan.Decision == proofGovernorReuse {
			plan.Decision = proofGovernorRun
			plan.RunChecks = normalizeProofGovernorIDs(strings.Split(opts.proofRequest, ","))
			plan.Reused = nil
		}
	}
	if opts.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(plan)
	} else {
		fmt.Fprintf(stdout, "proof-plan: %s\n", strings.ToUpper(plan.Decision))
		if len(plan.RunChecks) > 0 {
			fmt.Fprintf(stdout, "run: %s\n", strings.Join(plan.RunChecks, ","))
		}
		if len(plan.Reused) > 0 {
			fmt.Fprintf(stdout, "reuse: %s\n", strings.Join(plan.Reused, ","))
		}
		for _, reason := range plan.Reasons {
			fmt.Fprintf(stdout, "reason: %s\n", reason)
		}
	}
	if plan.Decision == proofGovernorBlock {
		return 2
	}
	return 0
}

func selectProofGovernorPlan(root string, registry proofGovernorRegistry, requested []string, receipts []proofGovernorReceipt, now time.Time) proofGovernorDecision {
	requested = normalizeProofGovernorIDs(requested)
	result := proofGovernorDecision{Decision: proofGovernorReuse}
	specs, registryIssues := indexProofGovernorRegistry(registry)
	if len(registryIssues) > 0 {
		result.Decision = proofGovernorBlock
		result.Reasons = append(result.Reasons, registryIssues...)
		return result
	}
	for _, id := range requested {
		if _, ok := specs[id]; !ok {
			result.Decision = proofGovernorBlock
			result.Reasons = append(result.Reasons, "unknown-check:"+id)
		}
	}
	if result.Decision == proofGovernorBlock {
		return result
	}

	unresolved := map[string]bool{}
	for _, id := range requested {
		unresolved[id] = true
		bestReasons := []string{}
		for _, receipt := range receipts {
			currentSpec, ok := specs[receipt.Spec.ID]
			if !ok || !stringSet(effectiveProofGovernorCoverage(currentSpec))[id] {
				continue
			}
			registryPath := receipt.RegistryPath
			if registryPath == "" {
				registryPath = "registry.json"
			}
			decision := evaluateProofGovernorReceipt(root, currentSpec, []string{id}, registryPath, receipt, now)
			if decision.Decision == proofGovernorReuse {
				delete(unresolved, id)
				result.Reused = append(result.Reused, id)
				result.ReceiptID = receipt.ReceiptID
				bestReasons = nil
				break
			}
			for _, reason := range decision.Reasons {
				bestReasons = append(bestReasons, id+":"+reason)
			}
		}
		if unresolved[id] {
			if len(bestReasons) == 0 {
				bestReasons = []string{id + ":no-passing-receipt"}
			}
			result.Reasons = append(result.Reasons, bestReasons...)
		}
	}

	if len(unresolved) == 0 {
		result.Decision = proofGovernorReuse
		result.Reused = normalizeProofGovernorIDs(result.Reused)
		result.Reasons = []string{"all-requested-checks-covered-by-fresh-passing-receipts"}
		return result
	}

	candidates := make([]proofGovernorCheckSpec, 0, len(unresolved))
	for id := range unresolved {
		candidates = append(candidates, specs[id])
	}
	sort.Slice(candidates, func(i, j int) bool {
		left := countProofGovernorUnresolvedCoverage(candidates[i], unresolved)
		right := countProofGovernorUnresolvedCoverage(candidates[j], unresolved)
		if left == right {
			return candidates[i].ID < candidates[j].ID
		}
		return left > right
	})
	remaining := map[string]bool{}
	for id := range unresolved {
		remaining[id] = true
	}
	for _, candidate := range candidates {
		if !remaining[candidate.ID] {
			continue
		}
		result.RunChecks = append(result.RunChecks, candidate.ID)
		for _, covered := range effectiveProofGovernorCoverage(candidate) {
			delete(remaining, covered)
		}
	}
	for id := range remaining {
		result.RunChecks = append(result.RunChecks, id)
	}
	sort.Strings(result.RunChecks)
	result.Reused = normalizeProofGovernorIDs(result.Reused)
	result.Reasons = dedupeProofGovernorStringsSorted(result.Reasons)
	result.Decision = proofGovernorRun
	return result
}

func indexProofGovernorRegistry(registry proofGovernorRegistry) (map[string]proofGovernorCheckSpec, []string) {
	issues := []string{}
	if registry.SchemaVersion != proofGovernorSchemaVersion {
		issues = append(issues, "unsupported-registry-schema")
	}
	specs := map[string]proofGovernorCheckSpec{}
	for _, raw := range registry.Checks {
		spec, specIssues := normalizeProofGovernorSpec(raw)
		if len(specIssues) > 0 {
			for _, issue := range specIssues {
				issues = append(issues, "invalid-check:"+raw.ID+":"+issue)
			}
			continue
		}
		if _, exists := specs[spec.ID]; exists {
			issues = append(issues, "duplicate-check:"+spec.ID)
			continue
		}
		specs[spec.ID] = spec
	}
	if len(specs) == 0 {
		issues = append(issues, "empty-check-registry")
	}
	sort.Strings(issues)
	return specs, issues
}

func loadProofGovernorRegistry(path string) (proofGovernorRegistry, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return proofGovernorRegistry{}, fmt.Errorf("read proof registry: %w", err)
	}
	if len(content) > maxProofGovernorBytes {
		return proofGovernorRegistry{}, fmt.Errorf("proof registry exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var registry proofGovernorRegistry
	if err := decoder.Decode(&registry); err != nil {
		return proofGovernorRegistry{}, fmt.Errorf("parse proof registry: %w", err)
	}
	if _, issues := indexProofGovernorRegistry(registry); len(issues) > 0 {
		return proofGovernorRegistry{}, fmt.Errorf("invalid proof registry: %s", strings.Join(issues, "; "))
	}
	return registry, nil
}

func loadProofGovernorReceipts(dir string) ([]proofGovernorReceipt, []string) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, []string{err.Error()}
	}
	receipts := []proofGovernorReceipt{}
	issues := []string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".proof.json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		content, readErr := os.ReadFile(path)
		if readErr != nil || len(content) > maxProofGovernorBytes {
			issues = append(issues, entry.Name()+":unreadable")
			continue
		}
		decoder := json.NewDecoder(bytes.NewReader(content))
		decoder.DisallowUnknownFields()
		var receipt proofGovernorReceipt
		if decodeErr := decoder.Decode(&receipt); decodeErr != nil {
			issues = append(issues, entry.Name()+":invalid")
			continue
		}
		receipts = append(receipts, receipt)
	}
	sort.Slice(receipts, func(i, j int) bool {
		return receipts[i].ReceiptID < receipts[j].ReceiptID
	})
	sort.Strings(issues)
	return receipts, issues
}

func countProofGovernorUnresolvedCoverage(spec proofGovernorCheckSpec, unresolved map[string]bool) int {
	count := 0
	for _, id := range effectiveProofGovernorCoverage(spec) {
		if unresolved[id] {
			count++
		}
	}
	return count
}

func dedupeProofGovernorStringsSorted(values []string) []string {
	sort.Strings(values)
	return dedupeProofGovernorStrings(values)
}
