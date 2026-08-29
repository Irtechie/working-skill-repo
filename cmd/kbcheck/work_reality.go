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

	workRealityActionReport = "report"
	workRealityActionMark   = "mark"
	workRealityActionRemove = "remove"

	rehabPacketMaxItems = 5
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
	NativeCheckGate          string   `json:"native_check_gate"`
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
	Action     string
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
	Line     int    `json:"line,omitempty"`
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

// workRealityRemoteAuthority is an alias, not a copy. A second struct with the
// same fields would let this survey and kbreconcile drift apart silently, which
// is exactly how they came to disagree about which branches were dead.
type workRealityRemoteAuthority = reconcile.RemoteAuthority

type workRealitySettled struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Status string `json:"status"`
}

type workRealityReport struct {
	SchemaVersion            int                        `json:"schema_version"`
	PredicateManifestVersion string                     `json:"predicate_manifest_version"`
	Status                   string                     `json:"status"`
	Action                   string                     `json:"action"`
	Cutoff                   string                     `json:"cutoff"`
	Root                     string                     `json:"root"`
	SessionID                string                     `json:"session_id,omitempty"`
	CurrentBranch            string                     `json:"current_branch,omitempty"`
	RemoteAuthority          workRealityRemoteAuthority `json:"remote_authority"`
	PreservationPredicates   []string                   `json:"preservation_predicates"`
	Declared                 []workRealityDeclared      `json:"declared"`
	Settled                  []workRealitySettled       `json:"settled"`
	Pairings                 []workRealityPairing       `json:"pairings"`
	Packet                   []rehabPacketItem          `json:"packet"`
	Marks                    *workRealityMarkResult     `json:"marks,omitempty"`
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
		// "skipped" is the marker applyWorkRealityMarks writes for a pairing it
		// already proved terminal. Omitting it made the mark action unable to
		// settle anything: the next run re-read its own marker as outstanding.
		"skipped": true,
	}
}

func runWorkRealityCommand(root string, opts options, stdout, stderr io.Writer) int {
	report, err := executeWorkReality(workRealityOptions{
		Root:      root,
		SessionID: opts.sessionID,
		Action:    opts.sliceLeaseAction,
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

	action := strings.TrimSpace(options.Action)
	if action == "" {
		action = workRealityActionReport
	}

	report := workRealityReport{
		SchemaVersion:          workRealitySchemaVersion,
		Status:                 workRealityStatusOK,
		Action:                 action,
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
	report.Packet = buildRehabPacket(report.Pairings, policy)

	if action == workRealityActionMark || action == workRealityActionRemove {
		marks := applyWorkRealityMarks(root, report, policy, action == workRealityActionRemove)
		report.Marks = &marks
	}

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

// resolveWorkRealityRemoteAuthority delegates to the shared probe so this
// survey and kbreconcile cannot reach different conclusions about the same
// remote from two copies of the same logic.
func resolveWorkRealityRemoteAuthority(root string, repository reconcile.Repository) workRealityRemoteAuthority {
	return reconcile.ResolveRemoteAuthority(root, repository.Remotes)
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
	type todoRow struct {
		line    int
		section string
		text    string
	}
	rows := []todoRow{}
	for lineIndex, line := range strings.Split(string(raw), "\n") {
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
		if isTableDelimiterRow(trimmed) {
			// A delimiter proves the row above it was the table header, which is
			// column labels rather than declared work.
			if last := len(rows) - 1; last >= 0 && rows[last].line == lineIndex {
				rows = rows[:last]
			}
			continue
		}
		if !isDeclaredWorkRow(trimmed) {
			continue
		}
		rows = append(rows, todoRow{line: lineIndex + 1, section: section, text: trimmed})
	}
	for index, row := range rows {
		item := newDeclaredItem(fmt.Sprintf("todo-%03d", index+1), "todo.md", row.section, row.text)
		item.Line = row.line
		declared = append(declared, item)
	}
	return declared
}

func isTableDelimiterRow(line string) bool {
	return strings.HasPrefix(line, "|") && strings.Trim(line, "|-: ") == ""
}

func isDeclaredWorkRow(line string) bool {
	if line == "" {
		return false
	}
	if strings.HasPrefix(line, "|") {
		// Skip table header separators such as |---|---|.
		if isTableDelimiterRow(line) {
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
	isTableRow := strings.HasPrefix(strings.TrimSpace(text), "|")
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
	if !isTableRow {
		// A list row carries its marker and status glyph in the same cell as its
		// title, so strip them and report work rather than markup.
		item.Status = declaredRowStatus(cells[0])
		cells[0] = trimListDecoration(cells[0])
	} else if len(cells) > 1 {
		item.Status = declaredRowStatus(cells[1])
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
	if isTableRow {
		item.Parsed = item.Title != "" && (item.Branch != "" || item.Manifest != "" || len(cells) > 1)
	} else {
		// A list row always yields exactly one cell, so the table-shaped
		// cell-count test would mark every readable bullet unparsed and bury it
		// under a missing-evidence reason it does not deserve.
		item.Parsed = item.Title != ""
	}
	return item
}

// trimListDecoration removes the list marker and any leading status glyph from
// a bullet row so the recorded title is the work itself.
func trimListDecoration(text string) string {
	trimmed := strings.TrimSpace(strings.TrimLeft(text, "-* "))
	return strings.TrimSpace(strings.TrimLeft(trimmed, workRealityStatusGlyphs))
}

// workRealityStatusGlyphs are the status markers a declared row may carry ahead
// of its text, including the markers this lane writes when it marks work.
const workRealityStatusGlyphs = "⬜🔧🔒⊘✅❌🟡🟢⏳🚧🛑 "

// declaredRowStatus reads the status a todo row already states. Without this a
// row carries no status at all, so no todo row can ever be settled: the mark
// action writes "⊘ skipped" as its terminal marker, and a later run that cannot
// read that marker re-surfaces the identical row as outstanding work forever.
func declaredRowStatus(cell string) string {
	text := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(cell), "-* "))
	glyph := ""
	for _, r := range text {
		if !strings.ContainsRune(workRealityStatusGlyphs, r) || r == ' ' {
			break
		}
		glyph = string(r)
		break
	}
	rest := strings.TrimSpace(strings.TrimLeft(text, workRealityStatusGlyphs))
	// A status word the row spells out wins over the glyph, because the word is
	// what a human wrote and the glyph is only its shorthand.
	if word := firstStatusWord(rest); word != "" {
		return word
	}
	switch glyph {
	case "⊘":
		return "skipped"
	case "✅", "🟢":
		return "completed"
	case "🔒", "🛑":
		return "blocked"
	case "🔧", "🚧":
		return "in_progress"
	case "⬜", "⏳", "🟡":
		return "pending"
	}
	return ""
}

// firstStatusWord returns the leading token of a status cell when it names a
// status, so a table row's dedicated status column is read literally.
func firstStatusWord(text string) string {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return ""
	}
	word := strings.ToLower(strings.Trim(fields[0], "`*_.,;:"))
	switch word {
	case "pending", "in_progress", "blocked", "skipped", "human-required",
		"abandoned", "cancelled", "canceled", "complete", "completed",
		"delivered", "done", "integrated", "retired", "shipped", "superseded":
		return word
	}
	return ""
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

// parseDirDeclaredWork reads declared work out of a lifecycle directory.
//
// It opens each file rather than listing names. A directory listing cannot see
// that a goal already declares status: complete, so finished work is reported
// outstanding on every pass and can never settle. That is the same defect that
// made todo rows resurface, in a different source.
// remoteOnlyShortName returns the branch name behind a refs/remotes ref when no
// local ref already represents it. A remote whose work is mirrored locally is
// surveyed through the local ref, and origin/HEAD is a symbolic pointer rather
// than work.
func remoteOnlyShortName(ref string, localShorts map[string]bool) (string, bool) {
	trimmed := strings.TrimPrefix(ref, "refs/remotes/")
	if trimmed == ref {
		return "", false
	}
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[1] == "" || parts[1] == "HEAD" {
		return "", false
	}
	if localShorts[parts[1]] {
		return "", false
	}
	return parts[1], true
}

func parseDirDeclaredWork(root, dir, kind string) []workRealityDeclared {
	if strings.TrimSpace(dir) == "" {
		return nil
	}
	base := filepath.Join(root, filepath.FromSlash(dir))
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	declared := []workRealityDeclared{}
	for _, entry := range entries {
		if entry.IsDir() || !isDeclarationFile(entry.Name()) {
			continue
		}
		relative := dir + "/" + entry.Name()
		item := workRealityDeclared{
			ID:     kind + ":" + relative,
			Source: relative,
			Title:  redactCredentialLike(entry.Name()),
			Parsed: true,
		}
		if body, readErr := os.ReadFile(filepath.Join(base, entry.Name())); readErr == nil {
			item.Status = declaredHeaderStatus(string(body))
		}
		declared = append(declared, item)
	}
	return declared
}

// isDeclarationFile reports whether a file can carry a work declaration at all.
// A .gitkeep placeholder exists to keep an empty directory in git; counting it
// as outstanding work invents an item no human ever declared.
func isDeclarationFile(name string) bool {
	return !strings.HasPrefix(name, ".") && strings.EqualFold(filepath.Ext(name), ".md")
}

// declaredHeaderStatus reads a lifecycle status from a document header.
//
// Goal and handoff files in this repository do not use YAML frontmatter. They
// write a plain "Status: complete" line under the H1, with a capital S. A
// case-sensitive lowercase-only match therefore reads every one of them as
// having no status, which reports finished goals as outstanding forever.
//
// The scan is bounded to the header region so a "Status:" line quoted in prose
// or inside a later section cannot silently retire a live goal.
func declaredHeaderStatus(body string) string {
	const headerLines = 20
	for index, line := range strings.Split(body, "\n") {
		if index >= headerLines {
			break
		}
		trimmed := strings.TrimSpace(line)
		if len(trimmed) < len("status:") || !strings.EqualFold(trimmed[:len("status:")], "status:") {
			continue
		}
		value := strings.TrimSpace(trimmed[len("status:"):])
		return strings.ToLower(strings.Trim(value, `"'`+"` "))
	}
	return ""
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

	localShorts := map[string]bool{}
	for _, branch := range repository.Branches {
		if !branch.IsRemote {
			localShorts[strings.TrimPrefix(branch.Ref, "refs/heads/")] = true
		}
	}

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

	// A branch that exists only on a remote is real outstanding work that no
	// local ref represents. Surveying local refs alone silently hides it.
	for _, branch := range repository.Branches {
		if !branch.IsRemote || branch.IsDefault {
			continue
		}
		short, ok := remoteOnlyShortName(branch.Ref, localShorts)
		if !ok {
			continue
		}
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

// ---------------------------------------------------------------------------
// Decision packet
// ---------------------------------------------------------------------------

// rehabPacketItem is one grouped, answerable decision. Every field is mandatory
// because an operator who cannot see the irreversible consequence and the safe
// default is being asked to guess, not to decide.
type rehabPacketItem struct {
	ID                      string   `json:"id"`
	State                   string   `json:"state"`
	RecommendedChoice       string   `json:"recommended_choice"`
	AffectedArtifacts       []string `json:"affected_artifacts"`
	Evidence                []string `json:"evidence"`
	Uncertainty             string   `json:"uncertainty"`
	IrreversibleConsequence string   `json:"irreversible_consequence"`
	SafeDefault             string   `json:"safe_default"`
	RecheckSensor           string   `json:"recheck_sensor"`
	PairingIDs              []string `json:"pairing_ids"`
	OmittedPairings         int      `json:"omitted_pairings,omitempty"`
}

// rehabPacketAmbiguousStates are the states that need a human decision. A
// terminal state is proven and does not belong in the packet; a preserved
// current branch is not a decision either.
func rehabPacketAmbiguousStates() map[string]int {
	return map[string]int{
		workRealityStateHumanRequired: 0,
		workRealityStateOrphanBranch:  1,
		workRealityStateUnshipped:     2,
		workRealityStateOrphanWork:    3,
		workRealityStateLive:          4,
	}
}

// rehabGlobalInstallRoots are the propagation targets named in AGENTS.md. A
// merge that touches .github/skills in this bundle is a supply-chain write into
// every future agent session on this host, so the packet must say so.
func rehabGlobalInstallRoots() []string {
	return []string{"~/.agents/skills", "~/.codex/skills", "~/.copilot/skills"}
}

// buildRehabPacket groups ambiguous pairings and returns at most five items.
// Grouping is by state and protection scope so an operator answers one question
// per kind of risk instead of one per branch. Nothing is dropped silently: a
// truncated group records how many pairings it still covers.
func buildRehabPacket(pairings []workRealityPairing, policy workRealityPolicy) []rehabPacketItem {
	priority := rehabPacketAmbiguousStates()
	grouped := map[string][]workRealityPairing{}
	for _, pairing := range pairings {
		if _, ambiguous := priority[pairing.State]; !ambiguous {
			continue
		}
		grouped[pairing.State+"/"+rehabProtectionScope(pairing)] = append(grouped[pairing.State+"/"+rehabProtectionScope(pairing)], pairing)
	}

	keys := make([]string, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := strings.SplitN(keys[i], "/", 2)
		right := strings.SplitN(keys[j], "/", 2)
		if priority[left[0]] != priority[right[0]] {
			return priority[left[0]] < priority[right[0]]
		}
		return keys[i] < keys[j]
	})

	items := []rehabPacketItem{}
	for index, key := range keys {
		group := grouped[key]
		sort.Slice(group, func(i, j int) bool { return group[i].ID < group[j].ID })
		if index >= rehabPacketMaxItems {
			// Fold every remaining group into the last item rather than
			// dropping it. A dropped decision is an invisible decision.
			last := &items[len(items)-1]
			last.OmittedPairings += len(group)
			continue
		}
		items = append(items, newRehabPacketItem(key, group, policy))
	}
	return items
}

func rehabProtectionScope(pairing workRealityPairing) string {
	if len(pairing.ProtectedPaths) == 0 {
		return "unprotected"
	}
	scoped := append([]string{}, pairing.ProtectedPaths...)
	sort.Strings(scoped)
	return scoped[0]
}

func newRehabPacketItem(key string, group []workRealityPairing, policy workRealityPolicy) rehabPacketItem {
	parts := strings.SplitN(key, "/", 2)
	state := parts[0]
	scope := ""
	if len(parts) > 1 {
		scope = parts[1]
	}

	item := rehabPacketItem{
		ID:            "packet-" + strings.ReplaceAll(key, "/", "-"),
		State:         state,
		SafeDefault:   "preserve: leave the work item and the ref untouched",
		RecheckSensor: "go run ./cmd/kbcheck work-reality --root . --json",
	}
	for _, pairing := range group {
		item.PairingIDs = append(item.PairingIDs, pairing.ID)
		if pairing.Ref != "" {
			item.AffectedArtifacts = append(item.AffectedArtifacts, pairing.Ref)
		}
		if pairing.DeclaredSource != "" && pairing.DeclaredID != "" {
			item.AffectedArtifacts = append(item.AffectedArtifacts, pairing.DeclaredSource+"#"+pairing.DeclaredID)
		}
		item.Evidence = append(item.Evidence, pairing.ID+": "+pairing.Reason+" (contained="+pairing.Contained+")")
	}
	item.AffectedArtifacts = dedupeStrings(item.AffectedArtifacts)
	item.Evidence = dedupeStrings(item.Evidence)

	switch state {
	case workRealityStateHumanRequired:
		item.RecommendedChoice = "hold: resolve the missing evidence before any delivery or cleanup decision"
		item.Uncertainty = "the report could not prove a lifecycle state for these pairings"
	case workRealityStateOrphanBranch:
		item.RecommendedChoice = "review each ref, then either open delivery for it or authorize reaping"
		item.Uncertainty = "the ref exists with no declared work item, so intent is unrecorded"
	case workRealityStateUnshipped:
		item.RecommendedChoice = "deliver through kb-complete, which owns the delivery policy"
		item.Uncertainty = "the ref holds commits the authoritative default does not contain"
	case workRealityStateOrphanWork:
		item.RecommendedChoice = "pair the work item with a ref or retire the declaration"
		item.Uncertainty = "the work item names no reachable ref or artifact"
	case workRealityStateLive:
		item.RecommendedChoice = "leave it alone; another session holds this work"
		item.Uncertainty = "a claim is held, so takeover would race a live owner"
	default:
		item.RecommendedChoice = "hold"
		item.Uncertainty = "unclassified state"
	}

	item.IrreversibleConsequence = rehabIrreversibleConsequence(state, scope, policy)
	return item
}

func rehabIrreversibleConsequence(state, scope string, policy workRealityPolicy) string {
	if scope != "" && scope != "unprotected" {
		for _, protected := range policy.ProtectedPaths {
			if protected != scope || !strings.Contains(protected, "skills") {
				continue
			}
			return "merging this group writes " + scope + " into the resolved default branch, which propagates to " +
				strings.Join(rehabGlobalInstallRoots(), ", ") +
				" and therefore into every future agent session on this host"
		}
		return "merging this group changes the protected path " + scope + " on the resolved default branch"
	}
	if state == workRealityStateOrphanBranch {
		return "reaping a ref discards its unique commits permanently"
	}
	return "editing or retiring the declaration loses the recorded intent for this work"
}

// ---------------------------------------------------------------------------
// Marker writes
// ---------------------------------------------------------------------------

type workRealityMarkWrite struct {
	PairingID  string `json:"pairing_id"`
	DeclaredID string `json:"declared_id"`
	Line       int    `json:"line"`
	Operation  string `json:"operation"`
	Marker     string `json:"marker,omitempty"`
	Reason     string `json:"reason"`
}

type workRealityMarkResult struct {
	Applied   bool                   `json:"applied"`
	Refused   string                 `json:"refused,omitempty"`
	Writes    []workRealityMarkWrite `json:"writes"`
	Blocked   []workRealityMarkWrite `json:"blocked"`
	Preserved int                    `json:"preserved"`
}

const (
	rehabMarkerSkipped       = "\u2298 skipped"
	rehabMarkerBlocked       = "\U0001F512 blocked"
	rehabMarkerHumanRequired = "\U0001F6D1 human-required"
)

// rehabKnownMarkers is todo.md's own status vocabulary. A row without one of
// these is a row whose shape we cannot prove, so it is never rewritten.
func rehabKnownMarkers() []string {
	return []string{
		"\u2B1C pending",
		"\U0001F527 in_progress",
		"\u2705 done",
		rehabMarkerBlocked,
		rehabMarkerSkipped,
		rehabMarkerHumanRequired,
	}
}

// applyWorkRealityMarks writes todo.md markers for pairings the report already
// proved terminal. It refuses entirely on a fail-closed report: a report that
// withheld conclusions cannot authorize a write.
func applyWorkRealityMarks(root string, report workRealityReport, policy workRealityPolicy, removalRequested bool) workRealityMarkResult {
	result := workRealityMarkResult{Writes: []workRealityMarkWrite{}, Blocked: []workRealityMarkWrite{}}
	if report.Status != workRealityStatusOK {
		result.Refused = "report is " + report.Status + "; no write may be authorized by a report that withheld conclusions"
		return result
	}

	declaredByID := map[string]workRealityDeclared{}
	for _, item := range report.Declared {
		declaredByID[item.ID] = item
	}

	path := filepath.Join(root, "todo.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		result.Refused = "read todo.md: " + err.Error()
		return result
	}
	lines := strings.Split(string(raw), "\n")

	planned := []workRealityMarkWrite{}
	for _, pairing := range report.Pairings {
		declared, ok := declaredByID[pairing.DeclaredID]
		if !ok || declared.Source != "todo.md" || declared.Line <= 0 || declared.Line > len(lines) {
			result.Preserved++
			continue
		}
		terminal := pairing.State == workRealityStateDead || pairing.State == workRealityStateSuperseded
		if !terminal || pairing.Contained != "true" {
			if removalRequested {
				// R8: an uncontained commit blocks removal. The row stays and
				// is re-marked so the blocked state is visible in todo.md.
				planned = append(planned, workRealityMarkWrite{
					PairingID:  pairing.ID,
					DeclaredID: declared.ID,
					Line:       declared.Line,
					Operation:  "mark",
					Marker:     rehabMarkerBlocked,
					Reason:     "removal blocked: state=" + pairing.State + " contained=" + pairing.Contained,
				})
				continue
			}
			result.Preserved++
			continue
		}

		artifact := rehabCompletingArtifact(root, report.RemoteAuthority.SHA, declared)
		if removalRequested && artifact != "" {
			planned = append(planned, workRealityMarkWrite{
				PairingID:  pairing.ID,
				DeclaredID: declared.ID,
				Line:       declared.Line,
				Operation:  "remove",
				Reason:     "removed: " + pairing.State + " and superseding artifact " + artifact + " exists in the authoritative default tree",
			})
			continue
		}
		reason := "marked " + pairing.State + ": " + pairing.Reason
		if removalRequested {
			reason = "removal blocked: no superseding or completing artifact is named in the authoritative default tree"
		}
		planned = append(planned, workRealityMarkWrite{
			PairingID:  pairing.ID,
			DeclaredID: declared.ID,
			Line:       declared.Line,
			Operation:  "mark",
			Marker:     rehabMarkerSkipped,
			Reason:     reason,
		})
	}

	if len(planned) == 0 {
		result.Applied = false
		return result
	}

	sort.Slice(planned, func(i, j int) bool { return planned[i].Line < planned[j].Line })
	removeLines := map[int]bool{}
	for _, write := range planned {
		index := write.Line - 1
		if write.Operation == "remove" {
			removeLines[index] = true
			result.Writes = append(result.Writes, write)
			continue
		}
		rewritten, ok := rehabRewriteMarker(lines[index], write.Marker, write.Reason)
		if !ok {
			result.Blocked = append(result.Blocked, workRealityMarkWrite{
				PairingID:  write.PairingID,
				DeclaredID: write.DeclaredID,
				Line:       write.Line,
				Operation:  "mark",
				Reason:     "row carries no todo.md status marker; shape unproven, so it was left untouched",
			})
			continue
		}
		lines[index] = rewritten
		result.Writes = append(result.Writes, write)
	}

	if len(result.Writes) == 0 {
		return result
	}

	kept := make([]string, 0, len(lines))
	for index, line := range lines {
		if removeLines[index] {
			continue
		}
		kept = append(kept, line)
	}
	if err := os.WriteFile(path, []byte(strings.Join(kept, "\n")), 0o644); err != nil {
		result.Refused = "write todo.md: " + err.Error()
		result.Writes = nil
		return result
	}
	result.Applied = true
	return result
}

// rehabCompletingArtifact returns the superseding or completing artifact only
// when it already exists in the authoritative default tree. A branch can never
// name its own replacement into existence.
func rehabCompletingArtifact(root, defaultSHA string, declared workRealityDeclared) string {
	for _, candidate := range []string{declared.SupersededBy, declared.Manifest} {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if pathExistsInTree(root, defaultSHA, candidate) {
			return candidate
		}
	}
	return ""
}

// rehabRewriteMarker replaces the row's existing status marker and appends the
// evidence. It preserves the line ending, because this repository is CRLF and a
// silent ending flip would rewrite the whole file in the diff.
func rehabRewriteMarker(line, marker, reason string) (string, bool) {
	body := line
	suffix := ""
	if strings.HasSuffix(body, "\r") {
		body = strings.TrimSuffix(body, "\r")
		suffix = "\r"
	}
	for _, known := range rehabKnownMarkers() {
		if !strings.Contains(body, known) {
			continue
		}
		rewritten := strings.Replace(body, known, marker, 1)
		note := "kb-rehab: " + reason
		if strings.HasSuffix(strings.TrimSpace(rewritten), "|") {
			rewritten = strings.TrimRight(rewritten, " ")
			rewritten = strings.TrimSuffix(rewritten, "|") + "; " + note + " |"
		} else {
			rewritten = strings.TrimRight(rewritten, " ") + "; " + note
		}
		return rewritten + suffix, true
	}
	return line, false
}
