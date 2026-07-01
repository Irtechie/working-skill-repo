package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type minimalityReport struct {
	GeneratedAt           string          `json:"generated_at"`
	Root                  string          `json:"root"`
	SkillRoot             string          `json:"skill_root"`
	AgentRoot             string          `json:"agent_root"`
	StaticOnly            bool            `json:"static_only"`
	Labels                []string        `json:"labels"`
	EvidenceClasses       []string        `json:"evidence_classes"`
	SkillClassifications  []minimalityRow `json:"skill_classifications"`
	AgentClassifications  []minimalityRow `json:"agent_classifications"`
	ColdStorageCandidates []minimalityRow `json:"cold_storage_candidates"`
}

type minimalityRow struct {
	Kind            string               `json:"kind"`
	Name            string               `json:"name"`
	Classification  string               `json:"classification"`
	Reason          string               `json:"reason"`
	Lines           int                  `json:"lines"`
	TokenEstimate   int                  `json:"token_estimate"`
	ReferencedBy    []string             `json:"referenced_by"`
	References      []string             `json:"references"`
	EvidenceClass   string               `json:"evidence_class"`
	EvidenceSources []minimalityEvidence `json:"evidence_sources"`
}

type minimalityEvidence struct {
	Class string `json:"class"`
	Path  string `json:"path"`
}

type loadedDoc struct {
	Class   string
	Path    string
	Content string
}

func runMinimalityCommand(root string, opts options, stdout, stderr io.Writer) int {
	skillRoot := opts.skillRoot
	if skillRoot == "" {
		skillRoot = ".github/skills"
	}
	agentRoot := opts.agentRoot
	if agentRoot == "" {
		agentRoot = ".github/agents"
	}
	report, err := computeMinimality(root, skillRoot, agentRoot, opts.trimLineThreshold)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, report)
	} else {
		fmt.Fprintln(stdout, "Skill surface minimality report: static-only=true")
		fmt.Fprintf(stdout, "Skills: %d; agents: %d; cold-storage candidates: %d\n", len(report.SkillClassifications), len(report.AgentClassifications), len(report.ColdStorageCandidates))
		counts := map[string]int{}
		for _, row := range append(report.SkillClassifications, report.AgentClassifications...) {
			counts[row.Classification]++
		}
		for _, key := range sortedMapKeys(counts) {
			fmt.Fprintf(stdout, "%s: %d\n", key, counts[key])
		}
		if len(report.ColdStorageCandidates) > 0 {
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "Cold-storage candidates, not deletion approvals:")
			for _, row := range report.ColdStorageCandidates {
				fmt.Fprintf(stdout, "%s [%s] %s evidence=%s: %s\n", row.Classification, row.Kind, row.Name, row.EvidenceClass, row.Reason)
			}
		}
	}
	return 0
}

func computeMinimality(root, skillRoot, agentRoot string, trimThreshold int) (minimalityReport, error) {
	skillRootFull := resolveRepoPath(root, skillRoot)
	agentRootFull := resolveRepoPath(root, agentRoot)
	if info, err := os.Stat(skillRootFull); err != nil || !info.IsDir() {
		return minimalityReport{}, fmt.Errorf("skill root not found: %s", skillRoot)
	}
	skills := loadSkillRows(skillRootFull)
	agents := loadAgentRows(agentRootFull)
	skillNames := keysForRows(skills)
	agentNames := keysForRows(agents)

	skillRefs := map[string][]string{}
	for _, skill := range skills {
		for _, candidate := range skillNames {
			if candidate != skill.Name && tokenReference(skill.Content, candidate) {
				skillRefs[skill.Name] = append(skillRefs[skill.Name], candidate)
			}
		}
		sort.Strings(skillRefs[skill.Name])
	}

	agentRefs := map[string][]string{}
	for _, agent := range agents {
		for _, skill := range skills {
			if tokenReference(skill.Content, agent.Name) {
				agentRefs[agent.Name] = append(agentRefs[agent.Name], skill.Name)
			}
		}
		sort.Strings(agentRefs[agent.Name])
	}

	evidenceByName := map[string][]minimalityEvidence{}
	addEvidence := func(name, class, path string) {
		evidenceByName[name] = append(evidenceByName[name], minimalityEvidence{Class: class, Path: relativePath(root, path)})
	}
	for skill, refs := range skillRefs {
		skillPath := filepath.Join(skillRootFull, skill, "SKILL.md")
		for _, ref := range refs {
			addEvidence(ref, "dispatch-static", skillPath)
		}
	}
	for agent, refs := range agentRefs {
		for _, skill := range refs {
			addEvidence(agent, "dispatch-static", filepath.Join(skillRootFull, skill, "SKILL.md"))
		}
	}
	allNames := append(append([]string{}, skillNames...), agentNames...)
	for _, doc := range minimalityEvidenceDocs(root) {
		for _, name := range allNames {
			if tokenReference(doc.Content, name) {
				addEvidence(name, doc.Class, doc.Path)
			}
		}
	}

	hotPath := setOf("kb-start", "kb-map", "kb-brainstorm", "kb-plan", "kb-work", "kb-complete", "kb-review", "kb-check")
	protectedSkills := setOf("kb-review", "ce-review", "ce-compound", "ce-compound-refresh", "document-review")
	unusedPatterns := []string{"ce-ideate", "ce-plan", "ce-work", "lfg", "slfg", "workflows-*"}
	report := minimalityReport{
		GeneratedAt: time.Now().Format(time.RFC3339Nano), Root: root, SkillRoot: skillRootFull, AgentRoot: agentRootFull, StaticOnly: true,
		Labels:          []string{"protected", "required", "conditional", "unproven", "unused-candidate", "trim-candidate"},
		EvidenceClasses: []string{"runtime", "dispatch-static", "example-only", "docs-only", "none"},
	}
	for _, skill := range skills {
		referencedBy := []string{}
		for referrer, refs := range skillRefs {
			if contains(refs, skill.Name) {
				referencedBy = append(referencedBy, referrer)
			}
		}
		sort.Strings(referencedBy)
		classification := "conditional"
		reason := "referenced by workflow skills or available as an explicit lane"
		if protectedSkills[skill.Name] {
			classification = "protected"
			reason = "protected by repo policy; do not delete unless callers and docs are rewritten"
		} else if hotPath[skill.Name] {
			classification = "required"
			reason = "hot-path KB workflow skill"
		} else if len(referencedBy) == 0 && namePattern(skill.Name, unusedPatterns) {
			classification = "unused-candidate"
			reason = "matches superseded workflow pattern and has no static inbound skill reference; cold-storage review only"
		} else if skill.Lines > trimThreshold {
			classification = "trim-candidate"
			reason = "over trim threshold; review for lazy-loading or line reduction before deletion"
		} else if len(referencedBy) == 0 {
			classification = "unproven"
			reason = "no static inbound skill reference found; runtime usage may still exist"
		}
		row := minimalityRow{Kind: "skill", Name: skill.Name, Classification: classification, Reason: reason, Lines: skill.Lines, TokenEstimate: skill.TokenEstimate, ReferencedBy: referencedBy, References: skillRefs[skill.Name], EvidenceClass: evidenceClass(evidenceByName[skill.Name]), EvidenceSources: sortedEvidence(evidenceByName[skill.Name])}
		report.SkillClassifications = append(report.SkillClassifications, row)
		if contains([]string{"unproven", "unused-candidate", "trim-candidate"}, row.Classification) {
			report.ColdStorageCandidates = append(report.ColdStorageCandidates, row)
		}
	}
	for _, agent := range agents {
		referencedBy := agentRefs[agent.Name]
		hotRefs := []string{}
		for _, ref := range referencedBy {
			if hotPath[ref] {
				hotRefs = append(hotRefs, ref)
			}
		}
		classification := "unproven"
		reason := "no static skill reference found; do not delete without runtime proof"
		if len(hotRefs) > 0 {
			classification = "required"
			reason = "referenced by hot-path skill(s): " + strings.Join(hotRefs, ", ")
		} else if len(referencedBy) > 0 {
			classification = "conditional"
			reason = "referenced by non-hot-path skill(s): " + strings.Join(referencedBy, ", ")
		} else if agent.Lines > trimThreshold {
			classification = "trim-candidate"
			reason = "unreferenced and over trim threshold; cold-storage review candidate"
		}
		row := minimalityRow{Kind: "agent", Name: agent.Name, Classification: classification, Reason: reason, Lines: agent.Lines, TokenEstimate: agent.TokenEstimate, ReferencedBy: referencedBy, References: []string{}, EvidenceClass: evidenceClass(evidenceByName[agent.Name]), EvidenceSources: sortedEvidence(evidenceByName[agent.Name])}
		report.AgentClassifications = append(report.AgentClassifications, row)
		if contains([]string{"unproven", "unused-candidate", "trim-candidate"}, row.Classification) {
			report.ColdStorageCandidates = append(report.ColdStorageCandidates, row)
		}
	}
	sort.Slice(report.ColdStorageCandidates, func(i, j int) bool {
		if report.ColdStorageCandidates[i].Kind == report.ColdStorageCandidates[j].Kind {
			return report.ColdStorageCandidates[i].Name < report.ColdStorageCandidates[j].Name
		}
		return report.ColdStorageCandidates[i].Kind < report.ColdStorageCandidates[j].Kind
	})
	return report, nil
}

func runMinimalitySelftest(stdout, stderr io.Writer) int {
	root, err := os.MkdirTemp("", "skill-minimality-*")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer os.RemoveAll(root)
	write := func(path, text string) {
		full := filepath.Join(root, filepath.FromSlash(path))
		_ = os.MkdirAll(filepath.Dir(full), 0o755)
		_ = os.WriteFile(full, []byte(text), 0o644)
	}
	write(".github/skills/kb-start/SKILL.md", "---\nname: kb-start\ndescription: route requests\n---\nUse correctness-reviewer and kb-work.\n")
	write(".github/skills/kb-work/SKILL.md", "---\nname: kb-work\ndescription: run work\n---\nRun plans and call conditional-reviewer for special checks.\n")
	write(".github/skills/feature-lane/SKILL.md", "---\nname: feature-lane\ndescription: optional lane\n---\nUse conditional-reviewer when needed.\n")
	write(".github/skills/ce-review/SKILL.md", "---\nname: ce-review\ndescription: protected generalized review skill\n---\nProtected even when static inbound references are absent.\n")
	write(".github/skills/giant-skill/SKILL.md", "---\nname: giant-skill\ndescription: large skill\n---\none\ntwo\nthree\nfour\nfive\nsix\nseven\n")
	write(".github/skills/workflows-old/SKILL.md", "---\nname: workflows-old\ndescription: superseded workflow\n---\nold alias\n")
	write(".github/skills/docs-mentioned/SKILL.md", "---\nname: docs-mentioned\ndescription: mentioned only in docs\n---\nstandalone\n")
	write(".github/skills/example-mentioned/SKILL.md", "---\nname: example-mentioned\ndescription: mentioned only in eval fixtures\n---\nstandalone\n")
	write(".github/skills/runtime-mentioned/SKILL.md", "---\nname: runtime-mentioned\ndescription: mentioned in durable runtime log\n---\nstandalone\n")
	write(".github/agents/correctness-reviewer.agent.md", "required")
	write(".github/agents/conditional-reviewer.agent.md", "conditional")
	write(".github/agents/unreferenced-reviewer.agent.md", "unproven")
	write("docs/context/research.md", "docs-mentioned is documented but not invoked.")
	write("evals/route/example.json", `{"prompt":"try example-mentioned"}`)
	write(".kb/observations.jsonl", `{"tool":"runtime-mentioned","result":"used"}`)
	report, err := computeMinimality(root, ".github/skills", ".github/agents", 6)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	assert := func(kind, name, classification, evidence string) bool {
		rows := report.SkillClassifications
		if kind == "agent" {
			rows = report.AgentClassifications
		}
		for _, row := range rows {
			if row.Name == name {
				if row.Classification != classification || (evidence != "" && row.EvidenceClass != evidence) {
					fmt.Fprintf(stderr, "%s: got class=%s evidence=%s want class=%s evidence=%s\n", name, row.Classification, row.EvidenceClass, classification, evidence)
					return false
				}
				return true
			}
		}
		fmt.Fprintf(stderr, "missing row %s\n", name)
		return false
	}
	if !assert("agent", "correctness-reviewer", "required", "dispatch-static") ||
		!assert("agent", "conditional-reviewer", "required", "") ||
		!assert("agent", "unreferenced-reviewer", "unproven", "none") ||
		!assert("skill", "giant-skill", "trim-candidate", "") ||
		!assert("skill", "workflows-old", "unused-candidate", "") ||
		!assert("skill", "ce-review", "protected", "") ||
		!assert("skill", "docs-mentioned", "unproven", "docs-only") ||
		!assert("skill", "example-mentioned", "unproven", "example-only") ||
		!assert("skill", "runtime-mentioned", "unproven", "runtime") {
		return 1
	}
	fmt.Fprintln(stdout, "skill-surface-minimality selftest passed")
	return 0
}

type textRow struct {
	Name          string
	Path          string
	Content       string
	Lines         int
	TokenEstimate int
}

func loadSkillRows(root string) []textRow {
	entries, _ := os.ReadDir(root)
	rows := []textRow{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name(), "SKILL.md")
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(content)
		rows = append(rows, textRow{Name: entry.Name(), Path: path, Content: text, Lines: countLines(text), TokenEstimate: tokenEstimate(text)})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	return rows
}

func loadAgentRows(root string) []textRow {
	files, _ := filepath.Glob(filepath.Join(root, "*.agent.md"))
	sort.Strings(files)
	rows := []textRow{}
	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(content)
		name := strings.TrimSuffix(filepath.Base(path), ".agent.md")
		rows = append(rows, textRow{Name: name, Path: path, Content: text, Lines: countLines(text), TokenEstimate: tokenEstimate(text)})
	}
	return rows
}

func keysForRows(rows []textRow) []string {
	keys := make([]string, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, row.Name)
	}
	return keys
}

func tokenReference(text, name string) bool {
	pattern := regexp.MustCompile(`(?i)(^|[^A-Za-z0-9_-])` + regexp.QuoteMeta(name) + `([^A-Za-z0-9_-]|$)`)
	return pattern.MatchString(text)
}

func namePattern(name string, patterns []string) bool {
	for _, pattern := range patterns {
		regex := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), `\*`, ".*") + "$"
		if regexp.MustCompile(regex).MatchString(name) {
			return true
		}
	}
	return false
}

func minimalityEvidenceDocs(root string) []loadedDoc {
	docs := []loadedDoc{}
	addFiles := func(path, class string) {
		full := resolveRepoPath(root, path)
		info, err := os.Stat(full)
		if err != nil {
			return
		}
		if !info.IsDir() {
			if content, err := os.ReadFile(full); err == nil {
				docs = append(docs, loadedDoc{Class: class, Path: full, Content: string(content)})
			}
			return
		}
		_ = filepath.WalkDir(full, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if !setOf(".md", ".json", ".jsonl", ".txt", ".yaml", ".yml")[ext] {
				return nil
			}
			if content, err := os.ReadFile(path); err == nil {
				docs = append(docs, loadedDoc{Class: class, Path: path, Content: string(content)})
			}
			return nil
		})
	}
	addFiles(".kb/observations.jsonl", "runtime")
	addFiles("evals", "example-only")
	for _, path := range []string{"docs", "README.md", "AGENTS.md", "todo.md", "todo-done.md"} {
		addFiles(path, "docs-only")
	}
	return docs
}

func evidenceClass(sources []minimalityEvidence) string {
	classes := map[string]bool{}
	for _, source := range sources {
		classes[source.Class] = true
	}
	for _, class := range []string{"runtime", "dispatch-static", "example-only", "docs-only"} {
		if classes[class] {
			return class
		}
	}
	return "none"
}

func sortedEvidence(sources []minimalityEvidence) []minimalityEvidence {
	sort.Slice(sources, func(i, j int) bool {
		if sources[i].Class == sources[j].Class {
			return sources[i].Path < sources[j].Path
		}
		return sources[i].Class < sources[j].Class
	})
	out := []minimalityEvidence{}
	seen := map[string]bool{}
	for _, source := range sources {
		key := source.Class + "\x00" + source.Path
		if !seen[key] {
			seen[key] = true
			out = append(out, source)
		}
	}
	return out
}

func setOf(values ...string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}
