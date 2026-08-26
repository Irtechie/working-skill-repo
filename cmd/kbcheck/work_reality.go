package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/reconcile"
)

const (
	workRealitySchemaVersion = 1

	workRealityStateLive          = "live"
	workRealityStateUnshipped     = "unshipped"
	workRealityStateDead          = "dead"
	workRealityStateSuperseded    = "superseded"
	workRealityStateOrphanWork    = "orphan-work"
	workRealityStateOrphanBranch  = "orphan-branch"
	workRealityStateHumanRequired = "human-required"

	workRealityStatusOK         = "ok"
	workRealityStatusFailClosed = "fail-closed"

	workRealityRedacted = "[redacted]"
)

// workRealityPreservationPredicates is the preservation floor. It must remain a
// superset of terminalCleanupSafetyPredicates(); work-reality may protect more
// than terminal cleanup, never less.
func workRealityPreservationPredicates() []string {
	names := map[string]bool{}
	for _, name := range terminalCleanupSafetyPredicates() {
		names[name] = true
	}
	for _, name := range []string{
		"canonical-real-path",
		"containment-proven",
		"current-session-identity",
		"fresh-remote-authority",
		"peer-claim-holder",
	} {
		names[name] = true
	}
	ordered := make([]string, 0, len(names))
	for name := range names {
		ordered = append(ordered, name)
	}
	sort.Strings(ordered)
	return ordered
}

type workRealityPolicy struct {
	SchemaVersion            int      `json:"schema_version"`
	PredicateManifestVersion string   `json:"predicate_manifest_version"`
	TerminalClaimStatuses    []string `json:"terminal_claim_statuses"`
	StaleClaimReviewAfterHrs float64  `json:"stale_claim_review_after_hours"`
	ProtectedPaths           []string `json:"protected_paths"`
	CredentialPathPatterns   []string `json:"credential_path_patterns"`
	DeclaredWorkSources      struct {
		TodoSections []string `json:"todo_sections"`
		ManifestGlob string   `json:"manifest_glob"`
		GoalDir      string   `json:"goal_dir"`
		HandoffDir   string   `json:"handoff_dir"`
	} `json:"declared_work_sources"`
}

type workRealityOptions struct {
	Root       string
	SessionID  string
	Cutoff     time.Time
	PolicyPath string
}

type workRealityEvidence struct {
	Name          string `json:"name"`
	State         string `json:"state"`
	Source        string `json:"source"`
	Authoritative bool   `json:"authoritative"`
	Detail        string `json:"detail,omitempty"`
}

type workRealityDeclared struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Source   string `json:"source"`
	Section  string `json:"section,omitempty"`
	Status   string `json:"status,omitempty"`
	Branch   string `json:"branch,omitempty"`
	Manifest string `json:"manifest,omitempty"`
	// SupersededBy is the repo-relative path a declared item nominates as its
	// replacement. A nomination is never self-proving; see classify.
	SupersededBy string `json:"superseded_by,omitempty"`
	Parsed       bool   `json:"parsed"`
}

type workRealityPairing struct {
	ID                string                `json:"id"`
	DeclaredID        string                `json:"declared_id,omitempty"`
	DeclaredSource    string                `json:"declared_source,omitempty"`
	Ref               string                `json:"ref,omitempty"`
	SHA               string                `json:"sha,omitempty"`
	State             string                `json:"state"`
	Reason            string                `json:"reason"`
	Contained         string                `json:"contained"`
	ProtectedPaths    []string              `json:"protected_paths,omitempty"`
	ProtectionReasons []string              `json:"protection_reasons,omitempty"`
	Evidence          []workRealityEvidence `json:"evidence"`
}

type workRealityRemoteAuthority struct {
	State         string `json:"state"`
	Remote        string `json:"remote,omitempty"`
	DefaultBranch string `json:"default_branch,omitempty"`
	SHA           string `json:"sha,omitempty"`
	Limitation    string `json:"limitation,omitempty"`
}

type workRealitySettled struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Status string `json:"status"`
}

type workRealityReport struct {
	SchemaVersion            int                        `json:"schema_version"`
	PredicateManifestVersion string                     `json:"predicate_manifest_version"`
	Status                   string                     `json:"status"`
	Cutoff                   string                     `json:"cutoff"`
	Root                     string                     `json:"root"`
	SessionID                string                     `json:"session_id,omitempty"`
	CurrentBranch            string                     `json:"current_branch,omitempty"`
	RemoteAuthority          workRealityRemoteAuthority `json:"remote_authority"`
	PreservationPredicates   []string                   `json:"preservation_predicates"`
	Declared                 []workRealityDeclared      `json:"declared"`
	Settled                  []workRealitySettled       `json:"settled"`
	Pairings                 []workRealityPairing       `json:"pairings"`
	Limitations              []string                   `json:"limitations,omitempty"`
}

// workRealityTerminalDeclaredStatuses name declared states that carry no
// outstanding work. Such an item is reported as settled, never as orphan-work:
// finished work is not orphaned, and treating it as orphaned would push a human
// to re-review plans that already landed.
func workRealityTerminalDeclaredStatuses() map[string]bool {
	return map[string]bool{
		"abandoned": true, "cancelled": true, "canceled": true, "complete": true,
		"completed": true, "delivered": true, "done": true, "integrated": true,
		"retired": true, "shipped": true, "superseded": true,
	}
}

func runWorkRealityCommand(root string, opts options, stdout, stderr io.Writer) int {
	report, err := executeWorkReality(workRealityOptions{
		Root:      root,
		SessionID: opts.sessionID,
		Cutoff:    time.Now().UTC(),
	})
	if err != nil {
		fmt.Fprintf(stderr, "work-reality: %v\n", err)
		return 1
	}
	encoded, err := marshalWorkReality(report)
	if err != nil {
		fmt.Fprintf(stderr, "work-reality: %v\n", err)
		return 1
	}
	if strings.TrimSpace(opts.output) != "" {
		if err := os.WriteFile(opts.output, append(encoded, '\n'), 0o600); err != nil {
			fmt.Fprintf(stderr, "work-reality: %v\n", err)
			return 1
		}
	}
	if opts.json || strings.TrimSpace(opts.output) == "" {
		fmt.Fprintln(stdout, string(encoded))
	}
	if report.Status == workRealityStatusFailClosed {
		// A fail-closed report is a valid, complete answer: it withheld
		// conclusions it could not prove. It is not a command failure.
		return 0
	}
	return 0
}

func marshalWorkReality(report workRealityReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

func executeWorkReality(options workRealityOptions) (workRealityReport, error) {
	if options.Cutoff.IsZero() {
		options.Cutoff = time.Now().UTC()
	}
	cutoff := options.Cutoff.UTC()

	root, err := gitTopLevel(options.Root)
	if err != nil {
		return workRealityReport{}, err
	}

	policyPath := options.PolicyPath
	if strings.TrimSpace(policyPath) == "" {
		policyPath = filepath.Join(root, "config", "rehab-policy.json")
	}
	policy, policyErr := loadWorkRealityPolicy(policyPath)

	report := workRealityReport{
		SchemaVersion:          workRealitySchemaVersion,
		Status:                 workRealityStatusOK,
		Cutoff:                 cutoff.Format(time.RFC3339),
		Root:                   root,
		SessionID:              options.SessionID,
		PreservationPredicates: workRealityPreservationPredicates(),
	}
	report.PredicateManifestVersion = policy.PredicateManifestVersion
	if policyErr != nil {
		report.Status = workRealityStatusFailClosed
		report.Limitations = append(report.Limitations,
			"predicate manifest unavailable: "+policyErr.Error()+"; no pairing may reach a terminal state")
	}

	ledger, err := reconcile.Inventory(reconcile.InventoryOptions{
		Roots:            []string{root},
		Cutoff:           cutoff,
		CurrentWorktree:  root,
		CurrentSessionID: options.SessionID,
		Actor:            "kbcheck-work-reality",
		Now:              func() time.Time { return cutoff },
	})
	if err != nil {
		return workRealityReport{}, fmt.Errorf("inventory: %w", err)
	}
	if len(ledger.Repositories) == 0 {
		return workRealityReport{}, fmt.Errorf("inventory returned no repository for %s", root)
	}
	repository := ledger.Repositories[0]
	report.CurrentBranch = repository.CurrentBranch

	authority := resolveWorkRealityRemoteAuthority(root, repository)
	report.RemoteAuthority = authority
	if authority.State != "authoritative" {
		report.Status = workRealityStatusFailClosed
		report.Limitations = append(report.Limitations,
			"remote authority unavailable: "+authority.Limitation+"; no pairing may reach a terminal state")
	}

	declared := parseDeclaredWork(root, policy)
	report.Declared = declared

	outstanding := []workRealityDeclared{}
	settled := []workRealitySettled{}
	terminalStatus := workRealityTerminalDeclaredStatuses()
	for _, item := range declared {
		if item.Status != "" && terminalStatus[strings.ToLower(item.Status)] {
			settled = append(settled, workRealitySettled{ID: item.ID, Source: item.Source, Status: item.Status})
			continue
		}
		outstanding = append(outstanding, item)
	}
	report.Settled = settled

	report.Pairings = pairWorkAgainstReality(root, repository, outstanding, policy, authority, cutoff, report.Status)

	sort.Strings(report.Limitations)
	return report, nil
}

func loadWorkRealityPolicy(path string) (workRealityPolicy, error) {
	policy := workRealityPolicy{}
	raw, err := os.ReadFile(path)
	if err != nil {
		return policy, fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}
	if err := json.Unmarshal(raw, &policy); err != nil {
		return policy, fmt.Errorf("parse %s: %w", filepath.Base(path), err)
	}
	if strings.TrimSpace(policy.PredicateManifestVersion) == "" {
		return policy, fmt.Errorf("%s has no predicate_manifest_version", filepath.Base(path))
	}
	return policy, nil
}

// resolveWorkRealityRemoteAuthority proves the default branch from a fresh
// remote advertisement, not from a cached remote-tracking ref. Every configured
// remote is consulted and the strictest outcome wins.
func resolveWorkRealityRemoteAuthority(root string, repository reconcile.Repository) workRealityRemoteAuthority {
	if len(repository.Remotes) == 0 {
		return workRealityRemoteAuthority{
			State:      "unavailable",
			Limitation: "no configured remote; containment cannot be proven",
		}
	}
	var resolved workRealityRemoteAuthority
	for _, remote := range repository.Remotes {
		candidate := resolveRemoteAuthorityFor(root, remote.Name)
		if candidate.State != "authoritative" {
			return candidate
		}
		if resolved.State == "" {
			resolved = candidate
			continue
		}
		if resolved.SHA != candidate.SHA || resolved.DefaultBranch != candidate.DefaultBranch {
			return workRealityRemoteAuthority{
				State: "unavailable",
				Limitation: fmt.Sprintf("remotes disagree on the default branch: %s/%s vs %s/%s",
					resolved.Remote, resolved.DefaultBranch, candidate.Remote, candidate.DefaultBranch),
			}
		}
	}
	return resolved
}

func resolveRemoteAuthorityFor(root, remote string) workRealityRemoteAuthority {
	advertised, err := gitCapture(root, "ls-remote", "--symref", remote, "HEAD")
	if err != nil {
		return workRealityRemoteAuthority{
			State:      "unavailable",
			Remote:     remote,
			Limitation: fmt.Sprintf("remote %s is unreachable", remote),
		}
	}
	branch, sha := parseSymrefAdvertisement(advertised)
	if branch == "" || sha == "" {
		return workRealityRemoteAuthority{
			State:      "unavailable",
			Remote:     remote,
			Limitation: fmt.Sprintf("remote %s advertised no resolvable default branch", remote),
		}
	}
	if _, err := gitCapture(root, "fetch", "--no-tags", remote, branch); err != nil {
		return workRealityRemoteAuthority{
			State:      "unavailable",
			Remote:     remote,
			Limitation: fmt.Sprintf("fetch of %s/%s failed", remote, branch),
		}
	}
	fetched, err := gitCapture(root, "rev-parse", "FETCH_HEAD")
	if err != nil {
		return workRealityRemoteAuthority{
			State:      "unavailable",
			Remote:     remote,
			Limitation: fmt.Sprintf("fetched head of %s/%s is unresolvable", remote, branch),
		}
	}
	fetched = strings.TrimSpace(fetched)
	if fetched != sha {
		return workRealityRemoteAuthority{
			State:  "unavailable",
			Remote: remote,
			Limitation: fmt.Sprintf("remote %s default moved between advertisement %s and fetch %s",
				remote, shortSHA(sha), shortSHA(fetched)),
		}
	}
	return workRealityRemoteAuthority{
		State:         "authoritative",
		Remote:        remote,
		DefaultBranch: branch,
		SHA:           sha,
	}
}

func parseSymrefAdvertisement(output string) (branch, sha string) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "ref:") {
			fields := strings.Fields(strings.TrimPrefix(line, "ref:"))
			if len(fields) >= 1 {
				branch = strings.TrimPrefix(fields[0], "refs/heads/")
			}
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "HEAD" {
			sha = fields[0]
		}
	}
	return branch, sha
}

var (
	workRealityBranchToken = regexp.MustCompile(`(?:^|[\s\x60(\[])((?:codex|preserved|parked|deaderestpool|feature|fix)/[A-Za-z0-9._\-/]+)`)
	workRealityManifestRef = regexp.MustCompile(`docs/plans/[A-Za-z0-9._\-]+\.md`)
	workRealitySuperseded  = regexp.MustCompile(`(?i)superseded\s+by\s+[\x60"']?([A-Za-z0-9._\-/]+)`)
)

func parseDeclaredWork(root string, policy workRealityPolicy) []workRealityDeclared {
	declared := []workRealityDeclared{}
	declared = append(declared, parseTodoDeclaredWork(root, policy)...)
	declared = append(declared, parseManifestDeclaredWork(root, policy)...)
	declared = append(declared, parseDirDeclaredWork(root, policy.DeclaredWorkSources.GoalDir, "goal")...)
	declared = append(declared, parseDirDeclaredWork(root, policy.DeclaredWorkSources.HandoffDir, "handoff")...)
	sort.SliceStable(declared, func(i, j int) bool { return declared[i].ID < declared[j].ID })
	return declared
}

func parseTodoDeclaredWork(root string, policy workRealityPolicy) []workRealityDeclared {
	path := filepath.Join(root, "todo.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	sections := policy.DeclaredWorkSources.TodoSections
	if len(sections) == 0 {
		sections = []string{"Active Work", "Queued Improvements", "Blocked", "Parked"}
	}
	declared := []workRealityDeclared{}
	section := ""
	index := 0
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			heading := strings.TrimSpace(strings.TrimLeft(trimmed, "# "))
			section = ""
			for _, candidate := range sections {
				if strings.EqualFold(heading, candidate) || strings.Contains(strings.ToLower(heading), strings.ToLower(candidate)) {
					section = candidate
					break
				}
			}
			continue
		}
		if section == "" {
			continue
		}
		if !isDeclaredWorkRow(trimmed) {
			continue
		}
		index++
		item := newDeclaredItem(fmt.Sprintf("todo-%03d", index), "todo.md", section, trimmed)
		declared = append(declared, item)
	}
	return declared
}

func isDeclaredWorkRow(line string) bool {
	if line == "" {
		return false
	}
	if strings.HasPrefix(line, "|") {
		// Skip table header separators such as |---|---|.
		if strings.Trim(line, "|-: ") == "" {
			return false
		}
		return true
	}
	return strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ")
}

// newDeclaredItem never drops a row it cannot understand. An unparsed row is a
// declared item in state orphan-work, because a row we cannot read is missing
// evidence, not absent work.
func newDeclaredItem(id, source, section, text string) workRealityDeclared {
	item := workRealityDeclared{ID: id, Source: source, Section: section}
	cleaned := strings.TrimSpace(strings.Trim(text, "|"))
	cells := []string{}
	for _, cell := range strings.Split(cleaned, "|") {
		cell = strings.TrimSpace(cell)
		if cell != "" {
			cells = append(cells, cell)
		}
	}
	if len(cells) == 0 {
		cells = []string{strings.TrimSpace(strings.TrimLeft(cleaned, "-* "))}
	}
	item.Title = redactCredentialLike(cells[0])
	if match := workRealityBranchToken.FindStringSubmatch(text); len(match) > 1 {
		item.Branch = strings.Trim(match[1], "`\"' ")
	}
	if match := workRealityManifestRef.FindString(text); match != "" {
		item.Manifest = match
	}
	if match := workRealitySuperseded.FindStringSubmatch(text); len(match) > 1 {
		item.SupersededBy = strings.Trim(match[1], "`\"' ")
	}
	item.Parsed = item.Title != "" && (item.Branch != "" || item.Manifest != "" || len(cells) > 1)
	return item
}

func parseManifestDeclaredWork(root string, policy workRealityPolicy) []workRealityDeclared {
	glob := policy.DeclaredWorkSources.ManifestGlob
	if strings.TrimSpace(glob) == "" {
		glob = "docs/plans/*-manifest.md"
	}
	matches, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(glob)))
	if err != nil {
		return nil
	}
	sort.Strings(matches)
	declared := []workRealityDeclared{}
	for _, match := range matches {
		relative := toSlashRelative(root, match)
		raw, err := os.ReadFile(match)
		if err != nil {
			declared = append(declared, workRealityDeclared{
				ID: "manifest:" + relative, Source: relative, Title: relative, Parsed: false,
			})
			continue
		}
		item := workRealityDeclared{ID: "manifest:" + relative, Source: relative, Manifest: relative}
		body := string(raw)
		item.Status = scalarFrontmatterValue(body, "status")
		item.Title = firstNonEmpty(scalarFrontmatterValue(body, "kb_id"), relative)
		item.Branch = scalarFrontmatterValue(body, "branch")
		item.Parsed = item.Status != ""
		declared = append(declared, item)
	}
	return declared
}

func parseDirDeclaredWork(root, dir, kind string) []workRealityDeclared {
	if strings.TrimSpace(dir) == "" {
		return nil
	}
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
	if err != nil {
		return nil
	}
	declared := []workRealityDeclared{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		relative := dir + "/" + entry.Name()
		declared = append(declared, workRealityDeclared{
			ID:     kind + ":" + relative,
			Source: relative,
			Title:  redactCredentialLike(entry.Name()),
			Parsed: true,
		})
	}
	return declared
}

func scalarFrontmatterValue(body, key string) string {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" && len(line) == len(trimmed) {
			continue
		}
		if !strings.HasPrefix(trimmed, key+":") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, key+":"))
		return strings.Trim(value, `"' `)
	}
	return ""
}

func pairWorkAgainstReality(
	root string,
	repository reconcile.Repository,
	declared []workRealityDeclared,
	policy workRealityPolicy,
	authority workRealityRemoteAuthority,
	cutoff time.Time,
	reportStatus string,
) []workRealityPairing {
	terminalAllowed := reportStatus == workRealityStatusOK && authority.State == "authoritative"

	claimsByBranch := map[string][]reconcile.QueueClaim{}
	for _, claim := range repository.QueueClaims {
		if claim.Branch == "" {
			continue
		}
		claimsByBranch[claim.Branch] = append(claimsByBranch[claim.Branch], claim)
	}
	protectedWorktreeBranches := map[string][]string{}
	for _, worktree := range repository.Worktrees {
		if worktree.Branch == "" {
			continue
		}
		reasons := append([]string{}, worktree.ProtectionReasons...)
		if worktree.IsCurrent {
			reasons = append(reasons, "current-session-identity")
		}
		if worktree.Locked {
			reasons = append(reasons, "not-locked")
		}
		if len(worktree.Dirt.Tracked)+len(worktree.Dirt.Untracked)+len(worktree.Dirt.Ignored) > 0 {
			reasons = append(reasons, "clean-tracked")
		}
		if len(reasons) > 0 {
			protectedWorktreeBranches[worktree.Branch] = append(protectedWorktreeBranches[worktree.Branch], reasons...)
		}
	}

	declaredByBranch := map[string]workRealityDeclared{}
	for _, item := range declared {
		if item.Branch != "" {
			declaredByBranch[item.Branch] = item
		}
	}

	pairings := []workRealityPairing{}
	pairedDeclared := map[string]bool{}

	for _, branch := range repository.Branches {
		if branch.IsRemote || branch.IsDefault {
			continue
		}
		short := strings.TrimPrefix(branch.Ref, "refs/heads/")
		if short == authority.DefaultBranch || short == repository.DefaultBranch {
			continue
		}
		item, hasItem := declaredByBranch[short]
		if hasItem {
			pairedDeclared[item.ID] = true
		}
		pairing := classifyPairing(root, classifyInput{
			Branch:            branch,
			Short:             short,
			Declared:          item,
			HasDeclared:       hasItem,
			Repository:        repository,
			Policy:            policy,
			Authority:         authority,
			Claims:            claimsByBranch[short],
			ProtectionReasons: dedupeStrings(protectedWorktreeBranches[short]),
			Cutoff:            cutoff,
			TerminalAllowed:   terminalAllowed,
		})
		pairings = append(pairings, pairing)
	}

	for _, item := range declared {
		if pairedDeclared[item.ID] {
			continue
		}
		pairings = append(pairings, classifyUnpairedDeclared(root, item, authority, terminalAllowed))
	}

	sort.SliceStable(pairings, func(i, j int) bool { return pairings[i].ID < pairings[j].ID })
	return pairings
}

type classifyInput struct {
	Branch            reconcile.Branch
	Short             string
	Declared          workRealityDeclared
	HasDeclared       bool
	Repository        reconcile.Repository
	Policy            workRealityPolicy
	Authority         workRealityRemoteAuthority
	Claims            []reconcile.QueueClaim
	ProtectionReasons []string
	Cutoff            time.Time
	TerminalAllowed   bool
}

func classifyPairing(root string, in classifyInput) workRealityPairing {
	pairing := workRealityPairing{
		ID:                "branch:" + in.Short,
		Ref:               in.Branch.Ref,
		SHA:               in.Branch.SHA,
		ProtectionReasons: in.ProtectionReasons,
		Contained:         "unknown",
	}
	if in.HasDeclared {
		pairing.DeclaredID = in.Declared.ID
		pairing.DeclaredSource = in.Declared.Source
	}

	if in.Short == in.Repository.CurrentBranch {
		pairing.ProtectionReasons = dedupeStrings(append(pairing.ProtectionReasons, "not-current", "current-session-identity"))
	}

	if len(pairing.ProtectionReasons) > 0 {
		pairing.State = workRealityStateLive
		pairing.Reason = "preserved: " + strings.Join(pairing.ProtectionReasons, ", ")
		pairing.Evidence = append(pairing.Evidence, workRealityEvidence{
			Name: "preservation", State: "protected", Source: "reconcile.Inventory", Authoritative: true,
			Detail: pairing.Reason,
		})
		return pairing
	}

	if state, reason, evidence, decided := classifyByClaim(in); decided {
		pairing.State = state
		pairing.Reason = reason
		pairing.Evidence = append(pairing.Evidence, evidence)
		return pairing
	}

	if !in.TerminalAllowed {
		pairing.State = workRealityStateHumanRequired
		pairing.Reason = "containment cannot be proven: " + in.Authority.Limitation
		pairing.Evidence = append(pairing.Evidence, workRealityEvidence{
			Name: "fresh-remote-authority", State: in.Authority.State, Source: "git ls-remote --symref",
			Authoritative: false, Detail: in.Authority.Limitation,
		})
		return pairing
	}

	contained, containErr := branchContained(root, in.Authority.SHA, in.Branch.SHA)
	if containErr != nil {
		pairing.State = workRealityStateHumanRequired
		pairing.Reason = "containment probe failed: " + containErr.Error()
		pairing.Evidence = append(pairing.Evidence, workRealityEvidence{
			Name: "containment-proven", State: "unavailable", Source: "git cherry", Authoritative: false,
			Detail: containErr.Error(),
		})
		return pairing
	}
	pairing.Contained = boolLabel(contained)
	pairing.Evidence = append(pairing.Evidence, workRealityEvidence{
		Name: "containment-proven", State: boolLabel(contained), Source: "git cherry", Authoritative: true,
		Detail: fmt.Sprintf("patch equivalence against %s/%s@%s",
			in.Authority.Remote, in.Authority.DefaultBranch, shortSHA(in.Authority.SHA)),
	})

	if in.HasDeclared && in.Declared.SupersededBy != "" {
		return classifySupersession(root, in, pairing, contained)
	}

	if contained {
		pairing.State = workRealityStateDead
		pairing.Reason = "every commit is patch-equivalent to the authoritative default"
		return pairing
	}

	pairing.ProtectedPaths = protectedPathsTouched(root, in.Authority.SHA, in.Branch.SHA, in.Policy)
	if len(pairing.ProtectedPaths) > 0 {
		pairing.Evidence = append(pairing.Evidence, workRealityEvidence{
			Name: "protected-path", State: "present", Source: "git diff --name-only", Authoritative: true,
			Detail: "changes touch protected roots; auto-merge ineligible",
		})
	}
	if in.HasDeclared {
		pairing.State = workRealityStateUnshipped
		pairing.Reason = "uncontained commits with a declared work item"
		return pairing
	}
	pairing.State = workRealityStateOrphanBranch
	pairing.Reason = "uncontained commits with no declared work item; preserved for review"
	return pairing
}

func classifyByClaim(in classifyInput) (string, string, workRealityEvidence, bool) {
	terminal := map[string]bool{}
	for _, status := range in.Policy.TerminalClaimStatuses {
		terminal[strings.ToLower(status)] = true
	}
	threshold := in.Policy.StaleClaimReviewAfterHrs
	if threshold <= 0 {
		threshold = 24
	}
	for _, claim := range in.Claims {
		if terminal[strings.ToLower(claim.Status)] {
			continue
		}
		age := in.Cutoff.Sub(claim.UpdatedAt)
		if !claim.UpdatedAt.IsZero() && age > time.Duration(threshold*float64(time.Hour)) {
			return workRealityStateHumanRequired,
				fmt.Sprintf("work claim %s is %s past the %.0fh review threshold; staleness is a review signal, not takeover authority",
					claim.WorkID, age.Round(time.Hour), threshold),
				workRealityEvidence{
					Name: "peer-claim-holder", State: "stale", Source: "work_queue", Authoritative: true,
					Detail: "claim status " + claim.Status,
				}, true
		}
		return workRealityStateLive,
			fmt.Sprintf("work claim %s is %s and non-terminal", claim.WorkID, claim.Status),
			workRealityEvidence{
				Name: "peer-claim-holder", State: "active", Source: "work_queue", Authoritative: true,
				Detail: "claim status " + claim.Status,
			}, true
	}
	return "", "", workRealityEvidence{}, false
}

// classifySupersession refuses a self-proving nomination. The replacement must
// already exist in the authoritative default tree, so a candidate branch can
// never authorize its own supersession.
func classifySupersession(root string, in classifyInput, pairing workRealityPairing, contained bool) workRealityPairing {
	target := in.Declared.SupersededBy
	onDefault := pathExistsInTree(root, in.Authority.SHA, target)
	pairing.Evidence = append(pairing.Evidence, workRealityEvidence{
		Name: "supersession-target", State: boolLabel(onDefault), Source: "git cat-file", Authoritative: true,
		Detail: fmt.Sprintf("%s present in %s/%s", redactCredentialLike(target), in.Authority.Remote, in.Authority.DefaultBranch),
	})
	if !onDefault {
		pairing.State = workRealityStateHumanRequired
		pairing.Reason = "supersession names " + redactCredentialLike(target) +
			", which is not on the authoritative default; a branch may not prove its own supersession"
		return pairing
	}
	if !contained {
		pairing.State = workRealityStateHumanRequired
		pairing.Reason = "supersession is nominated but this branch still holds uncontained commits"
		return pairing
	}
	pairing.State = workRealityStateSuperseded
	pairing.Reason = "replacement is on the authoritative default and every commit is contained"
	return pairing
}

func classifyUnpairedDeclared(root string, item workRealityDeclared, authority workRealityRemoteAuthority, terminalAllowed bool) workRealityPairing {
	pairing := workRealityPairing{
		ID:             "work:" + item.ID,
		DeclaredID:     item.ID,
		DeclaredSource: item.Source,
		Contained:      "unknown",
	}
	if !item.Parsed {
		pairing.State = workRealityStateOrphanWork
		pairing.Reason = "declared row could not be parsed; an unreadable row is missing evidence, never absent work"
		pairing.Evidence = append(pairing.Evidence, workRealityEvidence{
			Name: "declared-parse", State: "unavailable", Source: item.Source, Authoritative: false,
		})
		return pairing
	}
	if item.Manifest != "" && !fileExists(filepath.Join(root, filepath.FromSlash(item.Manifest))) {
		pairing.State = workRealityStateOrphanWork
		pairing.Reason = "named manifest " + item.Manifest + " is absent from this tree; a missing file is missing evidence, never proof of death"
		pairing.Evidence = append(pairing.Evidence, workRealityEvidence{
			Name: "manifest-present", State: "unavailable", Source: item.Manifest, Authoritative: false,
		})
		return pairing
	}
	if !terminalAllowed {
		pairing.State = workRealityStateHumanRequired
		pairing.Reason = "declared work has no paired ref and containment cannot be proven: " + authority.Limitation
		return pairing
	}
	pairing.State = workRealityStateOrphanWork
	pairing.Reason = "declared work has no paired ref"
	pairing.Evidence = append(pairing.Evidence, workRealityEvidence{
		Name: "ref-pairing", State: "absent", Source: item.Source, Authoritative: true,
	})
	return pairing
}

func branchContained(root, defaultSHA, branchSHA string) (bool, error) {
	if defaultSHA == "" || branchSHA == "" {
		return false, fmt.Errorf("missing sha for containment probe")
	}
	output, err := gitCapture(root, "cherry", defaultSHA, branchSHA)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "+") {
			return false, nil
		}
	}
	return true, nil
}

func protectedPathsTouched(root, defaultSHA, branchSHA string, policy workRealityPolicy) []string {
	output, err := gitCapture(root, "diff", "--name-only", defaultSHA+"..."+branchSHA)
	if err != nil {
		return nil
	}
	hits := map[string]bool{}
	for _, line := range strings.Split(output, "\n") {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}
		for _, protected := range policy.ProtectedPaths {
			if path == protected || strings.HasPrefix(path, protected+"/") {
				hits[protected] = true
			}
		}
	}
	ordered := make([]string, 0, len(hits))
	for hit := range hits {
		ordered = append(ordered, hit)
	}
	sort.Strings(ordered)
	return ordered
}

func pathExistsInTree(root, sha, path string) bool {
	if sha == "" || strings.TrimSpace(path) == "" {
		return false
	}
	_, err := gitCapture(root, "cat-file", "-e", sha+":"+path)
	return err == nil
}

var workRealityCredentialPatterns = []string{
	".env", ".p12", ".pem", ".pfx", "apikey", "api-key", "credential",
	"id_dsa", "id_ecdsa", "id_rsa", "passwd", "password", "private-key", "secret", "token",
}

func redactCredentialLike(value string) string {
	lowered := strings.ToLower(value)
	for _, pattern := range workRealityCredentialPatterns {
		if strings.Contains(lowered, pattern) {
			return workRealityRedacted
		}
	}
	return value
}

func gitTopLevel(dir string) (string, error) {
	output, err := gitCapture(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("%s is not a Git repository", dir)
	}
	resolved, err := filepath.EvalSymlinks(strings.TrimSpace(output))
	if err != nil {
		return strings.TrimSpace(output), nil
	}
	return resolved, nil
}

func gitCapture(dir string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = dir
	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0")
	output, err := command.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output)), err
	}
	return strings.TrimSpace(string(output)), nil
}

func toSlashRelative(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(relative)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func boolLabel(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	ordered := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		ordered = append(ordered, value)
	}
	sort.Strings(ordered)
	return ordered
}
