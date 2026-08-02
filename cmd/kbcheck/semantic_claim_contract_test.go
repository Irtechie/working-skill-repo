package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSemanticClaimRepoContract(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	required := map[string][]string{
		".github/skills/kb-start/scripts/work_queue.ps1": {
			"semantic_resources", "active_owner", "global_authority", "resource_conflict",
		},
		".github/skills/kb-start/SKILL.md": {
			"kbreconcile", "declared semantic resources", "authoritative adapter capability",
		},
		".github/skills/kb-finalize/SKILL.md": {
			"register completion state and evidence", "never deletes the current worktree",
		},
		".github/skills/kb-complete/SKILL.md": {
			"`local-durable`", "`awaiting-review`", "`delivery-integrated`",
			"delivery, physical cleanup, ref retirement, and host session retirement",
		},
	}
	for relative, phrases := range required {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		normalized := strings.ToLower(strings.Join(strings.Fields(string(content)), " "))
		for _, phrase := range phrases {
			if !strings.Contains(normalized, strings.ToLower(strings.Join(strings.Fields(phrase), " "))) {
				t.Errorf("%s weakens semantic claim contract: missing %q", relative, phrase)
			}
		}
	}
}

func TestDeliveryChainLifecycleStatesRemainSeparate(t *testing.T) {
	for _, test := range []struct {
		status, mode string
		want         bool
	}{
		{"local-durable", "local", true},
		{"awaiting-review", "pr", true},
		{"delivery-integrated", "direct", true},
		{"awaiting-review", "direct", false},
		{"delivery-integrated", "pr", false},
	} {
		if got := terminalClaimMatchesDelivery(test.status, test.mode); got != test.want {
			t.Errorf("status=%s mode=%s got=%t want=%t", test.status, test.mode, got, test.want)
		}
	}
}

func TestWorkQueueUpdateMigratesOwnedWorktreeIdentity(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell work queue helper is Windows-only")
	}

	sourceRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(sourceRoot, ".github", "skills", "kb-start", "scripts", "work_queue.ps1")
	temp := t.TempDir()
	repo := filepath.Join(temp, "repo")
	worktree := filepath.Join(temp, "worktree")

	runQueueContractCommand(t, temp, "git", "init", "--initial-branch=main", repo)
	if err := os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runQueueContractCommand(t, repo, "git", "add", "seed.txt")
	runQueueContractCommand(
		t, repo, "git", "-c", "user.name=KB Test", "-c", "user.email=kb@example.invalid",
		"commit", "-m", "seed",
	)
	runQueueContractCommand(
		t, repo, "powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
		"-File", script, "-Action", "claim", "-WorkId", "queue-migration-test",
		"-SessionId", "session-1", "-Branch", "main", "-Summary", "test", "-Scope", "test",
	)
	runQueueContractCommand(t, repo, "git", "worktree", "add", "-b", "feature", worktree)
	t.Cleanup(func() {
		command := exec.Command("git", "worktree", "remove", worktree)
		command.Dir = repo
		_ = command.Run()
	})

	runQueueContractCommand(
		t, worktree, "powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
		"-File", script, "-Action", "update", "-WorkId", "queue-migration-test",
		"-SessionId", "session-1", "-Branch", "feature", "-Summary", "test", "-Scope", "test",
	)
	output := runQueueContractCommand(
		t, worktree, "powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
		"-File", script, "-Action", "list",
	)

	type queueEntry struct {
		WorkID   string `json:"work_id"`
		Worktree string `json:"worktree"`
	}
	var entries []queueEntry
	if err := json.Unmarshal([]byte(output), &entries); err != nil {
		var entry queueEntry
		if singleErr := json.Unmarshal([]byte(output), &entry); singleErr != nil {
			t.Fatalf("decode queue list: array=%v single=%v\n%s", err, singleErr, output)
		}
		entries = []queueEntry{entry}
	}
	if len(entries) != 1 || entries[0].WorkID != "queue-migration-test" {
		t.Fatalf("unexpected queue entries: %+v", entries)
	}
	if !strings.EqualFold(filepath.Clean(entries[0].Worktree), filepath.Clean(worktree)) {
		t.Fatalf("queue worktree = %q, want %q", entries[0].Worktree, worktree)
	}
}

func runQueueContractCommand(t *testing.T, dir, name string, args ...string) string {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, output)
	}
	return strings.TrimSpace(string(output))
}
