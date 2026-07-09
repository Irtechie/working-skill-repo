package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type dishonestCompletionFixture struct {
	SchemaVersion int                       `json:"schema_version"`
	Suite         string                    `json:"suite"`
	Cases         []dishonestCompletionCase `json:"cases"`
}

type dishonestCompletionCase struct {
	ID            string           `json:"id"`
	Kind          string           `json:"kind"`
	ExpectedIssue string           `json:"expected_issue"`
	Manifest      string           `json:"manifest,omitempty"`
	History       []map[string]any `json:"history,omitempty"`
}

func runDishonestCompletionSelftest(root string, opts options, stdout, stderr io.Writer) int {
	fixturePath := opts.fixtureRoot
	if fixturePath == "" {
		fixturePath = "evals/dishonest-completion/fixtures.json"
	}
	fixture, err := loadDishonestCompletionFixture(resolveRepoPath(root, fixturePath))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if fixture.SchemaVersion != 1 || fixture.Suite != "dishonest-completion" || len(fixture.Cases) == 0 {
		fmt.Fprintln(stderr, "invalid dishonest-completion fixture suite")
		return 1
	}
	temp, err := os.MkdirTemp("", "kb-dishonest-completion-*")
	if err != nil {
		fmt.Fprintf(stderr, "create temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(temp)
	for _, tc := range fixture.Cases {
		if err := runDishonestCompletionCase(root, temp, tc); err != nil {
			fmt.Fprintf(stderr, "%s: %v\n", tc.ID, err)
			return 1
		}
	}
	fmt.Fprintf(stdout, "Dishonest completion selftest: %d rejection fixtures passed.\n", len(fixture.Cases))
	return 0
}

func loadDishonestCompletionFixture(path string) (dishonestCompletionFixture, error) {
	var fixture dishonestCompletionFixture
	if err := readJSONFile(path, &fixture); err != nil {
		return fixture, fmt.Errorf("read dishonest-completion fixture: %w", err)
	}
	return fixture, nil
}

func runDishonestCompletionCase(root, temp string, tc dishonestCompletionCase) error {
	if tc.ID == "" || tc.Kind == "" || tc.ExpectedIssue == "" {
		return fmt.Errorf("case must set id, kind, and expected_issue")
	}
	switch tc.Kind {
	case "manifest":
		path := filepath.Join(temp, tc.ID+".md")
		if err := os.WriteFile(path, []byte(tc.Manifest), 0o644); err != nil {
			return err
		}
		result, err := validateManifestContract(path)
		if err != nil {
			return err
		}
		if result.OK || !hasManifestIssue(result.Issues, tc.ExpectedIssue) {
			return fmt.Errorf("expected manifest issue %s, got %#v", tc.ExpectedIssue, result)
		}
	case "route-history":
		lines := []string{}
		for _, row := range tc.History {
			content, _ := json.Marshal(row)
			lines = append(lines, string(content))
		}
		path := filepath.Join(temp, tc.ID+".jsonl")
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			return err
		}
		result, err := validateRouteHistory(path)
		if err != nil {
			return err
		}
		if result.OK || !hasRunStateIssue(result.Issues, tc.ExpectedIssue) {
			return fmt.Errorf("expected route-history issue %s, got %#v", tc.ExpectedIssue, result)
		}
	case "proof-trace":
		tracePath := filepath.Join(temp, tc.ID+".jsonl")
		check := ProofCheck{Kind: "command_exit", Target: []string{"go", "version"}, Expect: 0}
		digest, err := proofCheckDigest(root, check)
		if err != nil {
			return err
		}
		green := proofSense(root, check)
		if !green.OK {
			return fmt.Errorf("vacuous fixture check did not start green: %+v", green)
		}
		if _, err := appendProofTrace(tracePath, "sense", digest, green.OK, green.Signal, green.Evidence); err != nil {
			return err
		}
		gate, err := proofAccept(root, tracePath, check)
		if err != nil {
			return err
		}
		if tc.ExpectedIssue != "vacuous-green" {
			return fmt.Errorf("unknown proof expected_issue %q", tc.ExpectedIssue)
		}
		if gate.OK || gate.SawRed {
			return fmt.Errorf("expected vacuous green rejection, got %+v", gate)
		}
	default:
		return fmt.Errorf("unknown dishonest-completion case kind %q", tc.Kind)
	}
	return nil
}
