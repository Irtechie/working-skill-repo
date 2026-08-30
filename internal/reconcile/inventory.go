package reconcile

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type InventoryOptions struct {
	Roots            []string
	Cutoff           time.Time
	CurrentWorktree  string
	CurrentSessionID string
	Actor            string
	Now              func() time.Time

	// RefreshRemoteAuthority opts into a network probe that establishes fresh
	// remote authority. It is off by default so inventory stays read-only and
	// offline, and because most callers only need cached observations.
	//
	// Without it no containment proof can be authoritative, so every
	// irreversible ref decision reports authoritative-containment-unavailable
	// and is deferred to a human. That is correct when the caller declined to
	// look, and wrong only if the caller wanted a real answer.
	RefreshRemoteAuthority bool
}

func Inventory(options InventoryOptions) (Ledger, error) {
	if len(options.Roots) == 0 {
		return Ledger{}, fmt.Errorf("at least one repository root is required")
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.Cutoff.IsZero() {
		options.Cutoff = options.Now().UTC()
	}
	if strings.TrimSpace(options.Actor) == "" {
		options.Actor = "kbreconcile"
	}
	ledger := Ledger{
		SchemaVersion: LedgerSchemaVersion,
		Cutoff:        options.Cutoff.UTC(), GeneratedAt: options.Now().UTC(),
		Actor: options.Actor, SessionID: options.CurrentSessionID,
		Limitations: []string{
			"missing host or provider adapters are reported unavailable and never inferred from Git",
		},
	}
	if options.RefreshRemoteAuthority {
		ledger.Limitations = append(ledger.Limitations,
			"remote authority was probed over the network; containment is proven against the advertised default")
	} else {
		ledger.Limitations = append(ledger.Limitations,
			"inventory is read-only; cached remote refs do not establish fresh remote authority")
	}
	sort.Strings(ledger.Limitations)
	seen := map[string]bool{}
	for _, candidate := range options.Roots {
		repository, err := inventoryRepository(candidate, options)
		if err != nil {
			return Ledger{}, err
		}
		if seen[repository.ID] {
			continue
		}
		seen[repository.ID] = true
		ledger.Repositories = append(ledger.Repositories, repository)
	}
	return normalizeLedger(ledger), nil
}

func inventoryRepository(candidate string, options InventoryOptions) (Repository, error) {
	root, err := gitOutput(candidate, "rev-parse", "--show-toplevel")
	if err != nil {
		return Repository{}, fmt.Errorf("%s is not a Git repository: %w", candidate, err)
	}
	root = canonicalPath(root)
	common, err := gitOutput(root, "rev-parse", "--git-common-dir")
	if err != nil {
		return Repository{}, err
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	common = canonicalPath(common)
	observed := options.Cutoff.UTC()
	repository := Repository{
		ID: repositoryID(common), Root: root, CommonDir: common, CurrentWorktree: canonicalPath(options.CurrentWorktree),
		DefaultBranchState: EvidenceUnavailable,
		Evidence: []Evidence{
			unavailableEvidence("host-sessions", "host-session-adapter", observed, "no compatible host session adapter configured"),
			unavailableEvidence("pull-requests", "provider-adapter", observed, "no compatible forge/provider adapter configured"),
			unavailableEvidence("remote-authority", "git", observed, "no network fetch or provider authority was requested; remote refs are cached observations"),
		},
	}
	repository.CurrentBranch, _ = gitOutput(root, "branch", "--show-current")
	repository.Remotes = inventoryRemotes(root)
	repository.DefaultBranch, repository.DefaultBranchState = inventoryDefaultBranch(root, repository.Remotes)
	if options.RefreshRemoteAuthority {
		authority := ResolveRemoteAuthority(root, repository.Remotes)
		repository.RemoteAuthority = &authority
		if authority.Authoritative() {
			repository.DefaultBranch = authority.DefaultBranch
			repository.DefaultBranchState = EvidenceAvailable
			repository.Evidence = replaceEvidence(repository.Evidence, Evidence{
				Name: "remote-authority", State: EvidenceAvailable, Source: RemoteAuthoritySource,
				ObservedAt: observed, Authoritative: true,
				Value: authority.Remote + "/" + authority.DefaultBranch,
			})
		} else {
			repository.Evidence = replaceEvidence(repository.Evidence, unavailableEvidence(
				"remote-authority", RemoteAuthoritySource, observed, authority.Limitation,
			))
		}
	}
	repository.Worktrees, err = inventoryWorktrees(root, common, repository, options)
	if err != nil {
		return Repository{}, err
	}
	if len(repository.Worktrees) > 0 {
		repository.PrimaryWorktree = repository.Worktrees[0].Path
		for index := range repository.Worktrees {
			repository.Worktrees[index].IsPrimary = sameCanonicalPath(repository.Worktrees[index].Path, repository.PrimaryWorktree)
			if repository.Worktrees[index].IsPrimary {
				addReason(&repository.Worktrees[index].ProtectionReasons, "primary")
			}
			repository.Worktrees[index].Predicates = inventoryWorktreePredicates(repository.Worktrees[index], options.Cutoff)
		}
	}
	repository.QueueClaims, repository.Evidence = inventoryQueue(common, observed, repository.Evidence)
	applyQueueProtections(&repository)
	repository.Receipts, repository.Evidence = inventoryReceipts(common, observed, repository.Evidence)
	repository.Branches, err = inventoryBranches(root, repository.DefaultBranch, observed, repository.RemoteAuthority)
	if err != nil {
		return Repository{}, err
	}
	repository.Artifacts = repositoryArtifacts(repository, options)
	return repository, nil
}

func inventoryWorktrees(root, common string, repository Repository, options InventoryOptions) ([]Worktree, error) {
	output, err := gitOutput(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var worktrees []Worktree
	var current Worktree
	flush := func() error {
		if current.Path == "" {
			return nil
		}
		current.Path = canonicalPath(current.Path)
		current.ID = "worktree:" + current.Path
		current.IsCurrent = sameCanonicalPath(current.Path, options.CurrentWorktree)
		current.IsDefault = repository.DefaultBranch != "" && current.Branch == repository.DefaultBranch
		current.Dirt = inventoryDirt(current.Path)
		current.AdminRoundTrip, current.GitDir = gitAdminRoundTrip(current.Path, common)
		if current.IsCurrent {
			addReason(&current.ProtectionReasons, "current")
		}
		if current.IsDefault {
			addReason(&current.ProtectionReasons, "default")
		}
		if current.Locked {
			addReason(&current.ProtectionReasons, "locked")
		}
		if !current.AdminRoundTrip {
			addReason(&current.ProtectionReasons, "ambiguous-git-admin")
		}
		if len(current.Dirt.Tracked) > 0 {
			addReason(&current.ProtectionReasons, "tracked-dirt")
		}
		if len(current.Dirt.Untracked) > 0 {
			addReason(&current.ProtectionReasons, "untracked-dirt")
		}
		if len(current.Dirt.Ignored) > 0 {
			addReason(&current.ProtectionReasons, "ignored-data")
		}
		if protectedMetadataPaths(current.Dirt) {
			addReason(&current.ProtectionReasons, "model-learning-credential-live")
		}
		current.Predicates = inventoryWorktreePredicates(current, options.Cutoff)
		sort.Strings(current.ProtectionReasons)
		worktrees = append(worktrees, current)
		current = Worktree{}
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
	return worktrees, nil
}

func inventoryDirt(worktree string) Dirt {
	output, err := gitBytes(worktree, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignored=matching")
	if err != nil {
		return Dirt{}
	}
	var dirt Dirt
	records := bytes.Split(output, []byte{0})
	for _, record := range records {
		if len(record) < 4 {
			continue
		}
		status := string(record[:2])
		path := strings.TrimSpace(string(record[3:]))
		switch status {
		case "??":
			dirt.Untracked = append(dirt.Untracked, filepath.ToSlash(path))
		case "!!":
			dirt.Ignored = append(dirt.Ignored, filepath.ToSlash(path))
		default:
			dirt.Tracked = append(dirt.Tracked, filepath.ToSlash(path))
		}
	}
	sort.Strings(dirt.Tracked)
	sort.Strings(dirt.Untracked)
	sort.Strings(dirt.Ignored)
	return dirt
}

func inventoryBranches(root, defaultBranch string, observed time.Time, authority *RemoteAuthority) ([]Branch, error) {
	output, err := gitBytes(root, "for-each-ref",
		"--format=%(refname)%00%(objectname)%00%(committerdate:unix)%00", "refs/heads", "refs/remotes")
	if err != nil {
		return nil, err
	}
	fields := bytes.Split(output, []byte{0})
	var branches []Branch
	for index := 0; index+2 < len(fields); index += 3 {
		ref := strings.TrimSpace(string(fields[index]))
		if ref == "" {
			continue
		}
		sha := strings.TrimSpace(string(fields[index+1]))
		unixValue := strings.TrimSpace(string(fields[index+2]))
		unixSeconds, _ := strconv.ParseInt(unixValue, 10, 64)
		branch := Branch{
			Ref: ref, SHA: sha, IsRemote: strings.HasPrefix(ref, "refs/remotes/"),
			IsDefault: strings.TrimPrefix(ref, "refs/heads/") == defaultBranch,
		}
		if unixSeconds > 0 {
			branch.UpdatedAt = time.Unix(unixSeconds, 0).UTC()
		}
		if branch.IsRemote {
			branch.Evidence = append(branch.Evidence, Evidence{
				Name: "remote-ref-cache", State: EvidenceAvailable, Source: "git-for-each-ref",
				ObservedAt: observed, Authoritative: false, Value: sha,
				Limitation: "cached remote-tracking ref; refresh required before any mutation",
			})
		}
		branches = append(branches, branch)
	}
	addCachedContainment(root, branches, defaultBranch, observed, authority)
	return branches, nil
}

func addCachedContainment(root string, branches []Branch, defaultBranch string, observed time.Time, authority *RemoteAuthority) {
	remoteByShort := map[string]Branch{}
	var defaultRemote Branch
	for _, branch := range branches {
		if !branch.IsRemote {
			continue
		}
		parts := strings.Split(strings.TrimPrefix(branch.Ref, "refs/remotes/"), "/")
		if len(parts) < 2 || parts[len(parts)-1] == "HEAD" {
			continue
		}
		short := strings.Join(parts[1:], "/")
		remoteByShort[short] = branch
		if short == defaultBranch {
			defaultRemote = branch
		}
	}
	// A freshly proven default outranks the cached remote-tracking ref, and is
	// the only target whose ancestry may be stamped authoritative.
	ancestryTarget := defaultRemote.SHA
	ancestrySource := "git-remote-tracking-ref"
	ancestryIdentity := "cached:"
	ancestryAuthoritative := false
	if authority != nil && authority.Authoritative() && authority.DefaultBranch == defaultBranch {
		ancestryTarget = authority.SHA
		ancestrySource = RemoteAuthoritySource
		ancestryIdentity = "fresh:"
		ancestryAuthoritative = true
	}
	for index := range branches {
		branch := &branches[index]
		if branch.IsRemote {
			continue
		}
		short := strings.TrimPrefix(branch.Ref, "refs/heads/")
		if remote, ok := remoteByShort[short]; ok && remote.SHA == branch.SHA {
			branch.DedupProofs = append(branch.DedupProofs, DedupProof{
				Algorithm: DedupRemoteTopicContainment,
				Identity:  "cached:" + branch.SHA, Source: "git-remote-tracking-ref",
				ObservedAt: observed, Authoritative: false,
			})
		}
		if ancestryTarget != "" {
			command := exec.Command("git", "-C", root, "merge-base", "--is-ancestor", branch.SHA, ancestryTarget)
			if command.Run() == nil {
				branch.DedupProofs = append(branch.DedupProofs, DedupProof{
					Algorithm: DedupRemoteDefaultAncestry,
					Identity:  ancestryIdentity + branch.SHA + ":" + ancestryTarget,
					Source:    ancestrySource, ObservedAt: observed,
					Authoritative: ancestryAuthoritative,
				})
			}
		}
	}
}

func inventoryRemotes(root string) []Remote {
	output, err := gitOutput(root, "remote", "-v")
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var remotes []Remote
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || seen[fields[0]+"\x00"+fields[1]] {
			continue
		}
		seen[fields[0]+"\x00"+fields[1]] = true
		remotes = append(remotes, Remote{Name: fields[0], URL: fields[1]})
	}
	return remotes
}

func inventoryDefaultBranch(root string, remotes []Remote) (string, string) {
	for _, remote := range remotes {
		ref, err := gitOutput(root, "symbolic-ref", "--quiet", "refs/remotes/"+remote.Name+"/HEAD")
		if err == nil && ref != "" {
			return strings.TrimPrefix(ref, "refs/remotes/"+remote.Name+"/"), EvidenceAvailable
		}
	}
	return "", EvidenceUnavailable
}

func inventoryQueue(common string, observed time.Time, evidence []Evidence) ([]QueueClaim, []Evidence) {
	path := filepath.Join(common, ".copilot-kb", "work-queue.json")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, append(evidence, unavailableEvidence(
			"queue-claims", "kb-work-queue", observed, "compatible queue file unavailable",
		))
	}
	content = bytes.TrimPrefix(content, []byte{0xef, 0xbb, 0xbf})
	var claims []QueueClaim
	if err := json.Unmarshal(content, &claims); err != nil {
		var claim QueueClaim
		if oneErr := json.Unmarshal(content, &claim); oneErr != nil {
			return nil, append(evidence, Evidence{
				Name: "queue-claims", State: EvidenceDisputed, Source: "kb-work-queue",
				ObservedAt: observed, Authoritative: false, Limitation: "queue schema could not be parsed",
			})
		}
		claims = []QueueClaim{claim}
	}
	return claims, append(evidence, Evidence{
		Name: "queue-claims", State: EvidenceAvailable, Source: "kb-work-queue",
		ObservedAt: observed, FreshUntil: observed.Add(5 * time.Minute),
		Authoritative: true, Value: strconv.Itoa(len(claims)),
	})
}

func inventoryReceipts(common string, observed time.Time, evidence []Evidence) ([]Receipt, []Evidence) {
	directory := filepath.Join(common, ".copilot-kb", "terminal-cleanup")
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, append(evidence, unavailableEvidence(
			"lifecycle-receipts", "kb-terminal-cleanup", observed, "compatible receipt directory unavailable",
		))
	}
	var receipts []Receipt
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			continue
		}
		receipts = append(receipts, Receipt{
			ID:   strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
			Path: filepath.Join(directory, entry.Name()), UpdatedAt: info.ModTime().UTC(),
		})
	}
	return receipts, append(evidence, Evidence{
		Name: "lifecycle-receipts", State: EvidenceAvailable, Source: "kb-terminal-cleanup",
		ObservedAt: observed, Authoritative: true, Value: strconv.Itoa(len(receipts)),
	})
}

func applyQueueProtections(repository *Repository) {
	for _, claim := range repository.QueueClaims {
		if claim.Status != "in_progress" && claim.Status != "active" && claim.Status != "paused" {
			continue
		}
		for index := range repository.Worktrees {
			worktree := &repository.Worktrees[index]
			if (claim.Worktree != "" && sameCanonicalPath(claim.Worktree, worktree.Path)) ||
				(claim.Branch != "" && claim.Branch == worktree.Branch) {
				addReason(&worktree.ProtectionReasons, "active-claim")
				sort.Strings(worktree.ProtectionReasons)
			}
		}
	}
}

func repositoryArtifacts(repository Repository, options InventoryOptions) []Artifact {
	var artifacts []Artifact
	if options.CurrentSessionID != "" {
		artifacts = append(artifacts, Artifact{
			ID: "session:" + options.CurrentSessionID, Kind: ArtifactSession,
			RepositoryID: repository.ID, Path: canonicalPath(options.CurrentWorktree),
			ObservedAt: options.Cutoff, ProtectionReasons: []string{"current"},
		})
	}
	for _, worktree := range repository.Worktrees {
		artifacts = append(artifacts, Artifact{
			ID: worktree.ID, Kind: ArtifactWorktree, RepositoryID: repository.ID,
			Path: worktree.Path, Ref: worktree.Branch, SHA: worktree.Head,
			ObservedAt: options.Cutoff, Dirt: worktree.Dirt,
			ProtectionReasons: append([]string(nil), worktree.ProtectionReasons...),
			Predicates:        append([]PredicateEvidence(nil), worktree.Predicates...),
			UniqueWork:        len(worktree.Dirt.Tracked)+len(worktree.Dirt.Untracked)+len(worktree.Dirt.Ignored) > 0,
		})
	}
	for _, branch := range repository.Branches {
		if branch.IsRemote {
			continue
		}
		var protections []string
		if branch.IsDefault {
			protections = append(protections, "default")
		}
		for _, worktree := range repository.Worktrees {
			if worktree.Branch == strings.TrimPrefix(branch.Ref, "refs/heads/") {
				protections = append(protections, "checked-out")
				break
			}
		}
		artifacts = append(artifacts, Artifact{
			ID: "branch:" + branch.Ref, Kind: ArtifactBranch, RepositoryID: repository.ID,
			Ref: branch.Ref, SHA: branch.SHA, ObservedAt: options.Cutoff, UpdatedAt: branch.UpdatedAt,
			ProtectionReasons: protections, DedupProofs: append([]DedupProof(nil), branch.DedupProofs...),
			Ambiguity: branchAmbiguity(branch),
		})
	}
	for _, claim := range repository.QueueClaims {
		protections := []string{}
		if claim.Status == "active" || claim.Status == "in_progress" || claim.Status == "paused" {
			protections = append(protections, "active-claim")
		}
		artifacts = append(artifacts, Artifact{
			ID: "queue:" + claim.WorkID, Kind: ArtifactQueueClaim, RepositoryID: repository.ID,
			Path: claim.Worktree, Ref: claim.Branch, ObservedAt: options.Cutoff, UpdatedAt: claim.UpdatedAt,
			ProtectionReasons: protections, Ambiguity: "host-session-terminal-proof-unavailable",
		})
	}
	for _, receipt := range repository.Receipts {
		artifacts = append(artifacts, Artifact{
			ID: "receipt:" + receipt.ID, Kind: ArtifactReceipt, RepositoryID: repository.ID,
			Path: receipt.Path, ObservedAt: options.Cutoff, UpdatedAt: receipt.UpdatedAt,
			ProtectionReasons: []string{"receipt-evidence"},
		})
	}
	return artifacts
}

func branchAmbiguity(branch Branch) string {
	for _, proof := range branch.DedupProofs {
		if proof.Authoritative {
			return ""
		}
	}
	return "authoritative-containment-unavailable"
}

func inventoryWorktreePredicates(worktree Worktree, observed time.Time) []PredicateEvidence {
	values := map[string]bool{
		"different-executor":        !worktree.IsCurrent,
		"clean-tracked":             len(worktree.Dirt.Tracked) == 0,
		"clean-untracked":           len(worktree.Dirt.Untracked) == 0,
		"clean-ignored":             len(worktree.Dirt.Ignored) == 0,
		"exact-worktree-generation": worktree.AdminRoundTrip,
		"git-admin-round-trip":      worktree.AdminRoundTrip,
		"not-current":               !worktree.IsCurrent,
		"not-primary":               !worktree.IsPrimary,
		"not-default":               !worktree.IsDefault,
		"not-locked":                !worktree.Locked,
		"not-moved":                 worktree.AdminRoundTrip,
		"not-post-cutoff":           true,
		"non-force-only":            true,
		"empty-residual-only":       true,
	}
	var predicates []PredicateEvidence
	for name, pass := range values {
		state := PredicateFail
		if pass {
			state = PredicatePass
		}
		source := "git"
		if name == "non-force-only" || name == "empty-residual-only" {
			source = "policy"
		}
		predicates = append(predicates, PredicateEvidence{
			Name: name, State: state, Source: source, ObservedAt: observed, Authoritative: true,
		})
	}
	for _, name := range []string{"terminal-or-suspended-claim", "durable-endpoint", "remote-monotonic"} {
		predicates = append(predicates, PredicateEvidence{
			Name: name, State: PredicateUnavailable, Source: "optional-adapter",
			ObservedAt: observed, Authoritative: false,
			Limitation: "required compatible adapter evidence is unavailable",
		})
	}
	sort.Slice(predicates, func(i, j int) bool { return predicates[i].Name < predicates[j].Name })
	return predicates
}

func protectedMetadataPaths(dirt Dirt) bool {
	paths := append(append(append([]string(nil), dirt.Tracked...), dirt.Untracked...), dirt.Ignored...)
	for _, path := range paths {
		lower := strings.ToLower(filepath.ToSlash(path))
		base := strings.ToLower(filepath.Base(lower))
		switch {
		case strings.HasPrefix(base, ".env"), strings.Contains(lower, "credential"),
			strings.Contains(lower, "secret"), strings.Contains(lower, "token"),
			strings.Contains(lower, "model"), strings.Contains(lower, "learning"),
			strings.Contains(lower, "memory"), strings.Contains(lower, "runtime"),
			strings.HasSuffix(base, ".db"), strings.HasSuffix(base, ".sqlite"),
			strings.HasSuffix(base, ".sock"), strings.HasSuffix(base, ".lock"),
			strings.Contains(lower, "/.kb/"):
			return true
		}
	}
	return false
}

func gitAdminRoundTrip(worktree, common string) (bool, string) {
	top, topErr := gitOutput(worktree, "rev-parse", "--show-toplevel")
	gitDir, gitErr := gitOutput(worktree, "rev-parse", "--absolute-git-dir")
	worktreeCommon, commonErr := gitOutput(worktree, "rev-parse", "--git-common-dir")
	if commonErr == nil && !filepath.IsAbs(worktreeCommon) {
		worktreeCommon = filepath.Join(worktree, worktreeCommon)
	}
	ok := topErr == nil && gitErr == nil && commonErr == nil &&
		sameCanonicalPath(top, worktree) && sameCanonicalPath(worktreeCommon, common)
	return ok, canonicalPath(gitDir)
}

func unavailableEvidence(name, source string, observed time.Time, limitation string) Evidence {
	return Evidence{
		Name: name, State: EvidenceUnavailable, Source: source,
		ObservedAt: observed, Authoritative: false, Limitation: limitation,
	}
}

// replaceEvidence overwrites an existing named observation rather than
// appending a second one, so a report never carries two answers for the same
// question and let a reader pick the convenient one.
func replaceEvidence(items []Evidence, replacement Evidence) []Evidence {
	for index := range items {
		if items[index].Name == replacement.Name {
			items[index] = replacement
			return items
		}
	}
	return append(items, replacement)
}

func addReason(reasons *[]string, reason string) {
	for _, existing := range *reasons {
		if existing == reason {
			return
		}
	}
	*reasons = append(*reasons, reason)
}

func canonicalPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return filepath.Clean(path)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(absolute); resolveErr == nil {
		return resolved
	}
	return absolute
}

func sameCanonicalPath(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	return strings.EqualFold(canonicalPath(left), canonicalPath(right))
}

func gitOutput(root string, args ...string) (string, error) {
	content, err := gitBytes(root, args...)
	return strings.TrimSpace(string(content)), err
}

func gitBytes(root string, args ...string) ([]byte, error) {
	commandArgs := append([]string{"-C", root}, args...)
	command := exec.Command("git", commandArgs...)
	content, err := command.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(exitError.Stderr)))
		}
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return content, nil
}
