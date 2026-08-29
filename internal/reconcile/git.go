package reconcile

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

type freshWorktree struct {
	Path   string
	GitDir string
	Head   string
	Branch string
	Locked bool
	Dirt   Dirt
}

type freshRemote struct {
	Name          string
	DefaultBranch string
	DefaultSHA    string
	TopicSHA      string
}

func acquireRepositoryLock(repository Repository, timeout time.Duration) (*modelrouting.PrivateStateLock, error) {
	common := repository.CommonDir
	if common == "" {
		var err error
		common, err = gitOutput(repository.Root, "rev-parse", "--git-common-dir")
		if err != nil {
			return nil, err
		}
		if !filepath.IsAbs(common) {
			common = filepath.Join(repository.Root, common)
		}
	}
	lockDir := filepath.Join(canonicalPath(common), ".copilot-kb")
	return modelrouting.AcquireSharedProjectLock(lockDir, "work-queue.lock", timeout)
}

func listFreshWorktrees(root string) ([]freshWorktree, error) {
	output, err := gitOutput(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var result []freshWorktree
	var current freshWorktree
	flush := func() error {
		if current.Path == "" {
			return nil
		}
		current.Path = canonicalPath(current.Path)
		current.Dirt = inventoryDirt(current.Path)
		gitDir, gitErr := gitOutput(current.Path, "rev-parse", "--absolute-git-dir")
		if gitErr == nil {
			current.GitDir = canonicalPath(gitDir)
		}
		result = append(result, current)
		current = freshWorktree{}
		return nil
	}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if err := flush(); err != nil {
				return nil, err
			}
			continue
		}
		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch refs/heads/"):
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case strings.HasPrefix(line, "locked"):
			current.Locked = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return result, nil
}

func refreshRemotes(root, topicBranch string) ([]freshRemote, error) {
	names, err := gitOutput(root, "remote")
	if err != nil {
		return nil, err
	}
	var remotes []freshRemote
	for _, name := range strings.Fields(names) {
		head, err := gitBytes(root, "ls-remote", "--symref", name, "HEAD")
		if err != nil {
			return nil, fmt.Errorf("refresh remote %s HEAD: %w", name, err)
		}
		item := freshRemote{Name: name}
		item.DefaultBranch, item.DefaultSHA = ParseSymrefAdvertisement(string(head))
		if item.DefaultBranch == "" || item.DefaultSHA == "" {
			return nil, fmt.Errorf("authoritative remote default is unresolved for %s", name)
		}
		if topicBranch != "" {
			topic, topicErr := gitOutput(root, "ls-remote", name, "refs/heads/"+topicBranch)
			if topicErr != nil {
				return nil, fmt.Errorf("refresh remote %s topic: %w", name, topicErr)
			}
			fields := strings.Fields(topic)
			if len(fields) > 0 {
				item.TopicSHA = fields[0]
			}
		}
		if _, err := gitOutput(root, "cat-file", "-e", item.DefaultSHA+"^{commit}"); err != nil {
			command := exec.Command("git", "-C", root, "fetch", "--no-write-fetch-head", name, item.DefaultSHA)
			if output, fetchErr := command.CombinedOutput(); fetchErr != nil {
				return nil, fmt.Errorf("fetch remote default object for %s: %v: %s", name, fetchErr, strings.TrimSpace(string(output)))
			}
		}
		remotes = append(remotes, item)
	}
	sort.Slice(remotes, func(i, j int) bool { return remotes[i].Name < remotes[j].Name })
	return remotes, nil
}

func remoteDefaultsMonotonic(root string, repository Repository, fresh []freshRemote) error {
	oldByRemote := map[string]string{}
	for _, branch := range repository.Branches {
		if !branch.IsRemote {
			continue
		}
		for _, remote := range fresh {
			want := "refs/remotes/" + remote.Name + "/" + remote.DefaultBranch
			if branch.Ref == want {
				oldByRemote[remote.Name] = branch.SHA
			}
		}
	}
	for _, remote := range fresh {
		old := oldByRemote[remote.Name]
		if old == "" {
			return fmt.Errorf("planned evidence lacks exact remote default for %s", remote.Name)
		}
		if old == remote.DefaultSHA {
			continue
		}
		if _, err := gitOutput(root, "cat-file", "-e", old+"^{commit}"); err != nil {
			return fmt.Errorf("planned remote default object is unavailable for %s", remote.Name)
		}
		command := exec.Command("git", "-C", root, "merge-base", "--is-ancestor", old, remote.DefaultSHA)
		if err := command.Run(); err != nil {
			return fmt.Errorf("authoritative remote default was rewritten for %s", remote.Name)
		}
	}
	return nil
}

func freshDeliveryState(root string, repository Repository, branch, sha string, requireIntegrated bool) (string, error) {
	remotes, err := refreshRemotes(root, branch)
	if err != nil {
		return StateUnavailable, err
	}
	if len(remotes) == 0 {
		if requireIntegrated {
			return StateUnavailable, fmt.Errorf("integrated remote default evidence is unavailable")
		}
		if ref, refErr := gitOutput(root, "rev-parse", "refs/heads/"+branch+"^{commit}"); refErr == nil && ref == sha {
			return StateLocalDurable, nil
		}
		return StateUnavailable, fmt.Errorf("exact local durable endpoint is unavailable")
	}
	if err := remoteDefaultsMonotonic(root, repository, remotes); err != nil {
		return StateUnavailable, err
	}
	for _, remote := range remotes {
		if remote.DefaultBranch == branch {
			return StateUnavailable, fmt.Errorf("authoritative remote default branch is the target")
		}
		command := exec.Command("git", "-C", root, "merge-base", "--is-ancestor", sha, remote.DefaultSHA)
		if command.Run() == nil {
			return StateIntegratedDefault, nil
		}
	}
	if requireIntegrated {
		return StateUnavailable, fmt.Errorf("authoritative remote default does not contain the target commit")
	}
	for _, remote := range remotes {
		if remote.TopicSHA == sha {
			return StateAwaitingReview, nil
		}
	}
	if ref, refErr := gitOutput(root, "rev-parse", "refs/heads/"+branch+"^{commit}"); refErr == nil && ref == sha {
		return StateLocalDurable, nil
	}
	return StateUnavailable, fmt.Errorf("no durable endpoint contains the exact target")
}

func freshQueueClaims(common string) ([]QueueClaim, error) {
	path := filepath.Join(common, ".copilot-kb", "work-queue.json")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content = bytes.TrimPrefix(content, []byte{0xef, 0xbb, 0xbf})
	var claims []QueueClaim
	if err := json.Unmarshal(content, &claims); err == nil {
		return claims, nil
	}
	var wrapper struct {
		Work []QueueClaim `json:"work"`
	}
	if err := json.Unmarshal(content, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Work, nil
}

func exactTerminalClaim(repository Repository, artifact Artifact, cutoff time.Time, currentSession string) error {
	claims, err := freshQueueClaims(repository.CommonDir)
	if err != nil {
		return fmt.Errorf("fresh terminal queue claim is unavailable: %w", err)
	}
	for _, claim := range claims {
		if (claim.Worktree != "" && sameCanonicalPath(claim.Worktree, artifact.Path)) ||
			(claim.Branch != "" && claim.Branch == strings.TrimPrefix(artifact.Ref, "refs/heads/")) {
			if claim.SessionID != "" && claim.SessionID == currentSession {
				return fmt.Errorf("current executing session owns the target")
			}
			if claim.UpdatedAt.After(cutoff) {
				return fmt.Errorf("queue claim changed after the plan cutoff")
			}
			switch claim.Status {
			case "done", "terminal", "suspended", "awaiting-review", "retired":
				return nil
			default:
				return fmt.Errorf("queue claim is not terminal or suspended")
			}
		}
	}
	return fmt.Errorf("exact terminal or suspended queue claim is unavailable")
}

func removeNonForceWorktree(root, path string) error {
	command := exec.Command("git", "-C", root, "worktree", "remove", path)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("non-force worktree removal failed: %v: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func deleteExactLocalRef(root, ref, sha string) error {
	command := exec.Command("git", "-C", root, "update-ref", "-d", ref, sha)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("exact-SHA local ref compare-and-swap failed: %v: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func exactEmptyResidual(path, plannedRealPath string) (bool, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("residual target is not an exact directory")
	}
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false, err
	}
	if !sameCanonicalPath(real, plannedRealPath) {
		return false, fmt.Errorf("residual real path does not match planned identity")
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	if len(entries) != 0 {
		return false, fmt.Errorf("residual directory is not empty; local data is preserved")
	}
	return true, nil
}
