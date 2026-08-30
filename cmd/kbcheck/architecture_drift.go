package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// architectureConfig is the closed component table for a repository.
//
// Components must be declared explicitly rather than inferred. Auto-detection
// is unreliable across repositories: components live at different depths under
// differently named parents, and some repositories carry no packaging
// manifests at all. The declaration is also the point of friction that keeps
// adding a component from being free.
type architectureConfig struct {
	Roots      []string `json:"roots"`
	DocsDir    string   `json:"docs_dir"`
	Components []string `json:"components"`
	Exempt     []string `json:"exempt"`
}

type architectureFinding struct {
	Component string `json:"component"`
	Kind      string `json:"kind"`
	Path      string `json:"path"`
	Detail    string `json:"detail"`
}

type architectureDriftReport struct {
	GeneratedAt  string                `json:"generated_at"`
	Root         string                `json:"root"`
	Status       string                `json:"status"`
	ConfigPath   string                `json:"config_path"`
	DocsDir      string                `json:"docs_dir"`
	Declared     []string              `json:"declared"`
	Observed     []string              `json:"observed"`
	Undeclared   []architectureFinding `json:"undeclared"`
	Phantom      []architectureFinding `json:"phantom"`
	Undocumented []architectureFinding `json:"undocumented"`
}

const architectureConfigRelPath = "config/architecture-components.json"

func runArchitectureDriftCommand(root string, opts options, stdout, stderr io.Writer) int {
	if opts.sliceLeaseAction == "init" {
		return runArchitectureDriftInit(root, opts, stdout, stderr)
	}
	report, err := computeArchitectureDrift(root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	findings := len(report.Undeclared) + len(report.Phantom) + len(report.Undocumented)

	if opts.json {
		writeJSON(stdout, report)
		if findings > 0 {
			return 1
		}
		return 0
	}

	switch report.Status {
	case "not-configured":
		fmt.Fprintf(stdout, "Architecture drift: not configured (%s absent); nothing to enforce\n", architectureConfigRelPath)
		return 0
	case "ok":
		fmt.Fprintf(stdout, "Architecture drift: ok; %d declared components, all present and documented\n", len(report.Declared))
		return 0
	}

	fmt.Fprintf(stdout, "Architecture drift: %d issues across %d declared components\n", findings, len(report.Declared))
	for _, f := range report.Undeclared {
		fmt.Fprintf(stdout, "ERROR undeclared: %s :: %s :: %s\n", f.Component, f.Path, f.Detail)
	}
	for _, f := range report.Phantom {
		fmt.Fprintf(stdout, "ERROR phantom: %s :: %s\n", f.Component, f.Detail)
	}
	for _, f := range report.Undocumented {
		fmt.Fprintf(stdout, "ERROR undocumented: %s :: %s\n", f.Component, f.Detail)
	}
	return 1
}

// runArchitectureDriftInit writes the initial declaration for a repository.
//
// It refuses to overwrite an existing declaration. Adding a component after
// bootstrap must stay a deliberate edit to a tracked file: that friction is the
// mechanism. A re-init that silently re-blessed whatever appeared on disk would
// convert the check into a rubber stamp.
func runArchitectureDriftInit(root string, opts options, stdout, stderr io.Writer) int {
	configPath := resolveRepoPath(root, architectureConfigRelPath)
	if _, err := os.Stat(configPath); err == nil {
		fmt.Fprintf(stderr, "%s already exists; edit it directly to declare a component\n", architectureConfigRelPath)
		return 1
	} else if !os.IsNotExist(err) {
		fmt.Fprintln(stderr, err)
		return 1
	}

	var roots []string
	for _, part := range strings.Split(opts.architectureRoots, ",") {
		part = strings.TrimSpace(strings.Trim(strings.TrimSpace(part), "/\\"))
		if part != "" {
			roots = append(roots, filepath.ToSlash(part))
		}
	}
	if len(roots) == 0 {
		fmt.Fprintln(stderr, "architecture-drift --action init requires at least one root")
		return 1
	}

	components := map[string]bool{}
	for _, rel := range roots {
		entries, err := os.ReadDir(resolveRepoPath(root, rel))
		if err != nil {
			fmt.Fprintf(stderr, "read root %q: %v\n", rel, err)
			return 1
		}
		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
				continue
			}
			components[name] = true
		}
	}

	cfg := architectureConfig{
		Roots:      roots,
		DocsDir:    "docs/context/architecture",
		Components: sortedSetKeys(components),
		Exempt:     []string{},
	}
	encoded, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := os.WriteFile(configPath, append(encoded, '\n'), 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stdout, "Architecture drift: wrote %s declaring %d components under %s\n",
		architectureConfigRelPath, len(cfg.Components), strings.Join(roots, ", "))
	fmt.Fprintln(stdout, "Run architecture-drift to see which declared components are undocumented.")
	return 0
}

func computeArchitectureDrift(root string) (architectureDriftReport, error) {
	report := architectureDriftReport{
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		Root:         root,
		ConfigPath:   architectureConfigRelPath,
		Declared:     []string{},
		Observed:     []string{},
		Undeclared:   []architectureFinding{},
		Phantom:      []architectureFinding{},
		Undocumented: []architectureFinding{},
	}

	configPath := resolveRepoPath(root, architectureConfigRelPath)
	raw, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			report.Status = "not-configured"
			return report, nil
		}
		return report, fmt.Errorf("read %s: %w", architectureConfigRelPath, err)
	}

	var cfg architectureConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return report, fmt.Errorf("parse %s: %w", architectureConfigRelPath, err)
	}
	if len(cfg.Roots) == 0 {
		return report, fmt.Errorf("%s declares no roots", architectureConfigRelPath)
	}

	docsDir := cfg.DocsDir
	if docsDir == "" {
		docsDir = "docs/context/architecture"
	}
	report.DocsDir = docsDir

	exempt := map[string]bool{}
	for _, name := range cfg.Exempt {
		exempt[name] = true
	}

	declared := map[string]bool{}
	for _, name := range cfg.Components {
		declared[name] = true
	}
	report.Declared = sortedSetKeys(declared)

	observed := map[string]string{}
	for _, rel := range cfg.Roots {
		full := resolveRepoPath(root, rel)
		entries, err := os.ReadDir(full)
		if err != nil {
			if os.IsNotExist(err) {
				return report, fmt.Errorf("declared root %q does not exist", rel)
			}
			return report, fmt.Errorf("read root %q: %w", rel, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || exempt[name] {
				continue
			}
			observed[name] = filepath.ToSlash(filepath.Join(rel, name))
		}
	}
	report.Observed = sortedMapStringKeys(observed)

	docs := loadArchitectureDocs(resolveRepoPath(root, docsDir))

	for _, name := range report.Observed {
		if !declared[name] {
			report.Undeclared = append(report.Undeclared, architectureFinding{
				Component: name,
				Kind:      "undeclared",
				Path:      observed[name],
				Detail:    fmt.Sprintf("component exists but is not declared in %s; declare it against an existing owner or remove it", architectureConfigRelPath),
			})
		}
	}

	for _, name := range report.Declared {
		if _, ok := observed[name]; !ok {
			report.Phantom = append(report.Phantom, architectureFinding{
				Component: name,
				Kind:      "phantom",
				Detail:    fmt.Sprintf("declared in %s but no matching directory under any declared root", architectureConfigRelPath),
			})
			continue
		}
		if !architectureDocumented(docs, name) {
			report.Undocumented = append(report.Undocumented, architectureFinding{
				Component: name,
				Kind:      "undocumented",
				Path:      observed[name],
				Detail:    fmt.Sprintf("no mention in %s; an undocumented component cannot be routed to by planning", docsDir),
			})
		}
	}

	if len(report.Undeclared)+len(report.Phantom)+len(report.Undocumented) == 0 {
		report.Status = "ok"
	} else {
		report.Status = "drift"
	}
	return report, nil
}

func loadArchitectureDocs(dir string) []loadedDoc {
	docs := []loadedDoc{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return docs
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		docs = append(docs, loadedDoc{Class: "architecture", Path: path, Content: string(content)})
	}
	return docs
}

// architectureDocumented matches on whole tokens so that a component named
// "memory" is not considered documented by an unrelated mention of
// "memory-eval".
func architectureDocumented(docs []loadedDoc, name string) bool {
	for _, doc := range docs {
		if tokenReference(filepath.Base(doc.Path), name) {
			return true
		}
		if tokenReference(doc.Content, name) {
			return true
		}
	}
	return false
}

func sortedSetKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedMapStringKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
