package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestAutomaticDeliveryChainContract(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	required := map[string][]string{
		".github/skills/kb-work/SKILL.md": {
			"Invoke `kb-finalize <manifest>` automatically.",
			"Successful `kb-finalize` continues automatically to `kb-complete`; `kb-work` must not stop after slice completion or finalization.",
			"`kb-work` never merges or pushes a resolved default branch.",
		},
		".github/skills/kb-finalize/SKILL.md": {
			"Invoke `kb-complete <manifest>` automatically.",
			"`kb-finalize` does not commit, push, open a PR, merge, or integrate the resolved remote default branch.",
			"Register completion state and evidence",
			"never deletes the current worktree",
		},
		".github/skills/kb-complete/SKILL.md": {
			"Successful delegated phases return to this state-driven loop automatically.",
			"Invoke `kb-ship <manifest>`",
			"invoke `kb-land <manifest>`",
			"configured delivery policy or explicit run-scoped user authorization",
			"`local-durable`",
			"`awaiting-review`",
			"`delivery-integrated`",
			"delivery, physical cleanup, ref retirement, and host session retirement",
			"Release or suspend ownership only after lifecycle registration succeeds",
			"versioned JSON outside the worktree being retired",
			"SHA-256 digest into the cleanup receipt",
			"--resume-packet <path-for-pr>",
		},
		".github/skills/kb-ship/SKILL.md": {
			"`kb-ship` never merges it",
			"Return the shipped PR evidence to `kb-complete`",
		},
		".github/skills/kb-land/SKILL.md": {
			"only KB skill authorized to integrate the resolved remote default branch",
			"Prove fetched remote-default containment before post-integration sync.",
		},
		"docs/context/architecture/kb-workflow.md": {
			"kb-work -> kb-finalize -> kb-complete -> kb-ship -> authorized kb-land",
			"`kb-work` and `kb-ship` return durable evidence to `kb-complete` instead of ending the run",
			"Only `kb-land` may integrate the resolved remote default branch.",
		},
		"evals/route-complexity/finish-plan-flow.json": {
			`"remote default containment"`,
			`"installed skill sync"`,
		},
		"evals/skill-eval/selftest/pass-finish-plan-flow.json": {
			`"kb-land"`,
			`"remote default containment"`,
			`"installed skill sync"`,
		},
	}

	for relative, phrases := range required {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatalf("read %s: %v", relative, err)
		}
		text := string(content)
		normalizedText := strings.Join(strings.Fields(text), " ")
		for _, phrase := range phrases {
			normalizedPhrase := strings.Join(strings.Fields(phrase), " ")
			if !strings.Contains(normalizedText, normalizedPhrase) {
				t.Errorf("%s missing automatic delivery contract %q", relative, phrase)
			}
		}
	}
}

func TestDeliveryAuthorityLedger(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path string
		want map[string]bool
	}{
		{
			path: ".github/skills/kb-work/SKILL.md",
			want: map[string]bool{
				"push_topic": false, "open_pr": false,
				"merge_remote_default": false, "push_remote_default": false,
				"integrate_remote_default": false,
			},
		},
		{
			path: ".github/skills/kb-ship/SKILL.md",
			want: map[string]bool{
				"push_topic": true, "open_pr": true,
				"merge_remote_default": false, "push_remote_default": false,
				"integrate_remote_default": false,
			},
		},
		{
			path: ".github/skills/kb-land/SKILL.md",
			want: map[string]bool{
				"push_topic": false, "open_pr": false,
				"merge_remote_default": true, "push_remote_default": true,
				"integrate_remote_default": true,
			},
		},
	}

	for _, tt := range tests {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(tt.path)))
		if err != nil {
			t.Fatalf("read %s: %v", tt.path, err)
		}
		assertDeliveryAuthority(t, tt.path, string(content), tt.want)
	}
}

func TestDeliveryAuthorityLedgerRejectsContradictoryGrants(t *testing.T) {
	workWant := map[string]bool{
		"push_topic": false, "open_pr": false,
		"merge_remote_default": false, "push_remote_default": false,
		"integrate_remote_default": false,
	}
	shipWant := map[string]bool{
		"push_topic": true, "open_pr": true,
		"merge_remote_default": false, "push_remote_default": false,
		"integrate_remote_default": false,
	}
	if !deliveryAuthorityMatches(deliveryAuthorityBlock(workWant), workWant) {
		t.Fatal("canonical generated kb-work authority block did not parse")
	}
	if !deliveryAuthorityMatches(deliveryAuthorityBlock(shipWant), shipWant) {
		t.Fatal("canonical generated kb-ship authority block did not parse")
	}

	workGrant := deliveryAuthorityBlock(map[string]bool{
		"push_topic": false, "open_pr": false,
		"merge_remote_default": false, "push_remote_default": true,
		"integrate_remote_default": false,
	})
	shipGrant := deliveryAuthorityBlock(map[string]bool{
		"push_topic": true, "open_pr": true,
		"merge_remote_default": true, "push_remote_default": false,
		"integrate_remote_default": false,
	})
	if deliveryAuthorityMatches(workGrant, workWant) {
		t.Fatal("kb-work authority contract accepted a resolved-default push grant")
	}
	if deliveryAuthorityMatches(shipGrant, shipWant) {
		t.Fatal("kb-ship authority contract accepted a remote-default merge grant")
	}

	missing := strings.Replace(deliveryAuthorityBlock(workWant), "  push_remote_default: false\n", "", 1)
	if deliveryAuthorityMatches(missing, workWant) {
		t.Fatal("authority contract accepted a missing expected key")
	}
	unknown := strings.Replace(deliveryAuthorityBlock(workWant), "  push_remote_default: false\n", "  bypass_checks: false\n", 1)
	if deliveryAuthorityMatches(unknown, workWant) {
		t.Fatal("authority contract accepted an unknown key")
	}
	duplicate := strings.Replace(
		deliveryAuthorityBlock(workWant),
		"  push_remote_default: false\n",
		"  push_remote_default: true\n  push_remote_default: false\n",
		1,
	)
	if deliveryAuthorityMatches(duplicate, workWant) {
		t.Fatal("authority contract accepted a contradictory duplicate key")
	}
}

func assertDeliveryAuthority(t *testing.T, path, content string, want map[string]bool) {
	t.Helper()
	if !deliveryAuthorityMatches(content, want) {
		t.Errorf("%s has an invalid or contradictory delivery_authority ledger", path)
	}
}

func deliveryAuthorityMatches(content string, want map[string]bool) bool {
	blockPattern := regexp.MustCompile("(?ms)```yaml\\s*\\ndelivery_authority:\\s*\\n(.*?)```")
	matches := blockPattern.FindAllStringSubmatch(content, -1)
	if len(matches) != 1 {
		return false
	}

	fields := map[string]bool{}
	for _, line := range strings.Split(matches[0][1], "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			return false
		}
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) != 2 {
			return false
		}
		key, raw := parts[0], strings.TrimSpace(parts[1])
		expected, known := want[key]
		if !known {
			return false
		}
		if _, duplicate := fields[key]; duplicate {
			return false
		}
		if raw != "true" && raw != "false" {
			return false
		}
		actual := raw == "true"
		if actual != expected {
			return false
		}
		fields[key] = actual
	}
	if len(fields) != len(want) {
		return false
	}
	return true
}

func deliveryAuthorityBlock(values map[string]bool) string {
	keys := []string{
		"push_topic", "open_pr", "merge_remote_default",
		"push_remote_default", "integrate_remote_default",
	}
	var builder strings.Builder
	builder.WriteString("```yaml\ndelivery_authority:\n")
	for _, key := range keys {
		builder.WriteString("  " + key + ": ")
		if values[key] {
			builder.WriteString("true\n")
		} else {
			builder.WriteString("false\n")
		}
	}
	builder.WriteString("```\n")
	return builder.String()
}
