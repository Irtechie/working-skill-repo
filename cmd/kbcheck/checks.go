package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Check struct {
	Name       string
	Args       []string
	Reason     string
	Required   bool
	Confidence string
	Run        func(root string) CheckResult
	Available  func(root string) bool
	SkipReason string
}

type CheckResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func (c Check) CommandString() string {
	return quoteArgs(c.Args)
}

func DiscoverChecks(root string) ([]Check, error) {
	checks := make([]Check, 0)

	if exists(root, "package.json") {
		pkgChecks, err := packageChecks(root)
		if err != nil {
			return nil, err
		}
		checks = append(checks, pkgChecks...)
	}
	if exists(root, "pyproject.toml") || exists(root, "pytest.ini") {
		checks = append(checks, Check{Name: "pytest", Args: []string{"python", "-m", "pytest"}, Reason: "Python test config detected", Required: true, Confidence: "deterministic-local"})
	}
	if exists(root, "go.mod") {
		checks = append(checks, Check{Name: "go-test", Args: []string{"go", "test", "./..."}, Reason: "Go module detected", Required: true, Confidence: "deterministic-local"})
	}
	checks = append(checks, dotnetChecks(root)...)
	if exists(root, "Makefile") {
		checks = append(checks, Check{Name: "make-test", Args: []string{"make", "test"}, Reason: "Makefile detected", Required: true, Confidence: "deterministic-local"})
	}
	if exists(root, ".github/skills") && exists(root, "config/skill-quality.json") {
		skillChecks, err := skillRepoChecks(root)
		if err != nil {
			return nil, err
		}
		checks = append(checks, skillChecks...)
	}
	return checks, nil
}

func packageChecks(root string) ([]Check, error) {
	type packageJSON struct {
		Scripts map[string]string `json:"scripts"`
	}
	content, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return nil, err
	}
	var pkg packageJSON
	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}
	runner := "npm"
	runPrefix := []string{"npm", "run"}
	if exists(root, "pnpm-lock.yaml") {
		runner = "pnpm"
		runPrefix = []string{"pnpm"}
	} else if exists(root, "yarn.lock") {
		runner = "yarn"
		runPrefix = []string{"yarn"}
	}

	names := []string{"lint", "typecheck", "test", "test:unit", "test:integration", "test:e2e", "build"}
	checks := make([]Check, 0, len(names))
	for _, name := range names {
		if _, ok := pkg.Scripts[name]; !ok {
			continue
		}
		args := append(append([]string{}, runPrefix...), name)
		checks = append(checks, Check{Name: name, Args: args, Reason: "package.json script via " + runner, Required: true, Confidence: "deterministic-local"})
	}
	return checks, nil
}

func dotnetChecks(root string) []Check {
	sln, _ := filepath.Glob(filepath.Join(root, "*.sln"))
	if len(sln) > 0 {
		name := filepath.Base(sln[0])
		return []Check{
			{Name: "dotnet-test", Args: []string{"dotnet", "test", name}, Reason: ".NET solution detected", Required: true, Confidence: "deterministic-local"},
			{Name: "dotnet-build", Args: []string{"dotnet", "build", name, "--no-restore"}, Reason: ".NET solution detected", Required: true, Confidence: "deterministic-local"},
		}
	}
	csproj := firstRecursiveMatch(root, ".csproj")
	if csproj != "" {
		return []Check{{Name: "dotnet-test", Args: []string{"dotnet", "test"}, Reason: ".NET project detected", Required: true, Confidence: "deterministic-local"}}
	}
	return nil
}

func skillRepoChecks(root string) ([]Check, error) {
	type psCheck struct {
		Name   string
		Script string
		Extra  []string
		Reason string
	}
	psChecks := []psCheck{
		{"skill-lint", "scripts/skill-lint.ps1", nil, "skill quality config detected"},
		{"route-complexity-eval", "scripts/route-complexity-eval.ps1", nil, "route complexity eval fixtures detected"},
		{"skill-eval", "scripts/skill-eval.ps1", nil, "skill eval selftest fixtures detected"},
		{"skill-eval-manifest-selftest", "scripts/skill-eval-manifest-selftest.ps1", nil, "skill eval protected-file hash selftest detected"},
		{"skill-eval-baseline-selftest", "scripts/skill-eval-baseline-selftest.ps1", nil, "skill eval baseline regression selftest detected"},
		{"skill-eval-codex-dry-run", "scripts/skill-eval-run-codex.ps1", []string{"-FixtureId", "tiny-typo-fix", "-DryRun"}, "Codex skill eval adapter detected"},
		{"skill-eval-ghcp-dry-run", "scripts/skill-eval-run-ghcp.ps1", []string{"-FixtureId", "tiny-typo-fix", "-DryRun"}, "GHCP skill eval adapter detected"},
		{"skill-eval-quality", "scripts/skill-eval-quality.ps1", nil, "skill output quality rubric fixtures detected"},
		{"kb-pipeline-selftest", "scripts/kb-pipeline-selftest.ps1", nil, "KB coded pipeline spike selftest detected"},
		{"skill-surface-report", "scripts/skill-surface-report.ps1", nil, "skill loaded-surface report detected"},
		{"skill-marketplace-firebreak", "scripts/skill-marketplace-firebreak.ps1", nil, "private marketplace quarantine firebreak detected"},
		{"skill-marketplace-firebreak-selftest", "scripts/skill-marketplace-firebreak-selftest.ps1", nil, "private marketplace quarantine firebreak negative selftest detected"},
		{"marketplace-promotion-selftest", "scripts/promote-marketplace-skill-selftest.ps1", nil, "private marketplace safe promotion selftest detected"},
		{"kb-release-gate-selftest", "scripts/kb-release-gate-selftest.ps1", nil, "release gate profile selftest detected"},
		{"skill-surface-minimality-selftest", "scripts/skill-surface-minimality-selftest.ps1", nil, "skill/agent minimality classification selftest detected"},
		{"skill-surface-minimality", "scripts/skill-surface-minimality.ps1", nil, "static skill/agent minimality report detected"},
		{"cross-model-benchmark-validate", "scripts/cross-model-benchmark-validate.ps1", nil, "cross-model benchmark prompt fixtures detected"},
		{"atv-upstream-delta-selftest", "scripts/atv-upstream-delta-selftest.ps1", nil, "read-only ATV upstream delta selftest detected"},
		{"atv-upstream-delta", "scripts/atv-upstream-delta.ps1", nil, "read-only ATV upstream delta report detected"},
		{"skill-sync-report", "scripts/skill-sync-report.ps1", nil, "skill sync target config detected"},
	}

	checks := make([]Check, 0, len(psChecks)+1)
	for _, pc := range psChecks {
		if !exists(root, pc.Script) {
			continue
		}
		args, err := powerShellArgs(pc.Script, pc.Extra...)
		if err != nil {
			return nil, err
		}
		checks = append(checks, Check{Name: pc.Name, Args: args, Reason: pc.Reason, Required: true, Confidence: "deterministic-local"})
	}
	if exists(root, "scripts/skill-eval-wrap.ps1") && exists(root, "scripts/skill-eval-run-ghcp.ps1") {
		args, err := powerShellArgs("scripts/skill-eval-wrap.ps1", "-Runner", "scripts/skill-eval-run-ghcp.ps1", "-FixtureId", "tiny-typo-fix", "-DryRun", "-Sealed")
		if err != nil {
			return nil, err
		}
		checks = append(checks, Check{Name: "skill-eval-observed-trace-dry-run", Args: args, Reason: "observed trace eval wrapper detected", Required: true, Confidence: "deterministic-local"})
	}
	sort.SliceStable(checks, func(i, j int) bool {
		return checks[i].Name < checks[j].Name
	})
	return checks, nil
}

func exists(root, path string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(path)))
	return err == nil
}

func firstGlob(root, pattern string) string {
	matches, _ := filepath.Glob(filepath.Join(root, filepath.FromSlash(pattern)))
	if len(matches) == 0 {
		return ""
	}
	return matches[0]
}

func firstRecursiveMatch(root, suffix string) string {
	var found string
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || found != "" {
			return nil
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(strings.ToLower(entry.Name()), strings.ToLower(suffix)) {
			found = path
		}
		return nil
	})
	return found
}

func quoteArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "" || strings.ContainsAny(arg, " \t\"") {
			quoted = append(quoted, `"`+strings.ReplaceAll(arg, `"`, `\"`)+`"`)
			continue
		}
		quoted = append(quoted, arg)
	}
	return strings.Join(quoted, " ")
}
