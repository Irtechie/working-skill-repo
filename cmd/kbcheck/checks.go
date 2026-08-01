package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
	Timeout    time.Duration
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
		checks = append(checks, Check{
			Name:       "go-test",
			Args:       []string{"go", "test", "-buildvcs=false", "./..."},
			Reason:     "Go module detected",
			Required:   true,
			Confidence: "deterministic-local",
			Run:        runGoTestsWithRouterIsolation,
		})
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

func runGoTestsWithRouterIsolation(root string) CheckResult {
	list := runProcessCheck(root, Check{Args: []string{"go", "list", "./..."}})
	if list.ExitCode != 0 {
		return list
	}
	var regular []string
	router := ""
	for _, pkg := range strings.Fields(list.Stdout) {
		if strings.HasSuffix(pkg, "/cmd/kbrouter") {
			router = pkg
		} else {
			regular = append(regular, pkg)
		}
	}
	var stdout, stderr strings.Builder
	if len(regular) > 0 {
		args := append([]string{"go", "test", "-buildvcs=false"}, regular...)
		result := runProcessCheck(root, Check{Args: args})
		stdout.WriteString(result.Stdout)
		stderr.WriteString(result.Stderr)
		if result.ExitCode != 0 {
			return CheckResult{ExitCode: result.ExitCode, Stdout: stdout.String(), Stderr: stderr.String()}
		}
	}
	if router != "" {
		// kbrouter tests validate their own Windows job-object containment, so
		// nesting only this package inside the generic check job blocks its
		// command fixtures from starting.
		result := runGoCommandWithoutOuterContainment(root, "test", "-buildvcs=false", router)
		stdout.WriteString(result.Stdout)
		stderr.WriteString(result.Stderr)
		if result.ExitCode != 0 {
			return CheckResult{ExitCode: result.ExitCode, Stdout: stdout.String(), Stderr: stderr.String()}
		}
	}
	return CheckResult{Stdout: stdout.String(), Stderr: stderr.String()}
}

func runGoCommandWithoutOuterContainment(root string, args ...string) CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), defaultProcessCheckTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = root
	overflow := make(chan struct{}, 1)
	stdout := newCappedCheckBuffer(overflow)
	stderr := newCappedCheckBuffer(overflow)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return CheckResult{ExitCode: 1, Stderr: err.Error()}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var err error
	select {
	case err = <-done:
	case <-ctx.Done():
		err = <-done
	case <-overflow:
		_ = cmd.Process.Kill()
		err = <-done
		return CheckResult{ExitCode: 1, Stdout: stdout.String(), Stderr: stderr.String() + fmt.Sprintf("go command output exceeded %d bytes", maxProcessCheckOutputBytes)}
	}
	if ctx.Err() == context.DeadlineExceeded {
		return CheckResult{ExitCode: 1, Stdout: stdout.String(), Stderr: stderr.String() + "go command timed out"}
	}
	if err == nil {
		return CheckResult{Stdout: stdout.String(), Stderr: stderr.String()}
	}
	exitCode := 1
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}
	return CheckResult{ExitCode: exitCode, Stdout: stdout.String(), Stderr: stderr.String()}
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
	type nativeCheck struct {
		Name   string
		Reason string
	}
	nativeChecks := []nativeCheck{
		{"skill-lint", "skill quality config detected"},
		{"skill-guidance", "current agent skill guidance contract detected"},
		{"kb-doctor-selftest", "KB doctor install drift repair selftest detected"},
		{"route-complexity-eval", "route complexity eval fixtures detected"},
		{"review-reference-guard", "review skill shared-reference drift guard detected"},
		{"skill-eval", "skill eval selftest fixtures detected"},
		{"skill-eval-manifest-selftest", "skill eval protected-file hash selftest detected"},
		{"skill-eval-baseline-selftest", "skill eval baseline regression selftest detected"},
		{"skill-eval-codex-dry-run", "Codex skill eval adapter detected"},
		{"skill-eval-ghcp-dry-run", "GHCP skill eval adapter detected"},
		{"skill-eval-quality", "skill output quality rubric fixtures detected"},
		{"manifest-contract-selftest", "KB manifest phase/gate proof contract selftest detected"},
		{"proof-governor-selftest", "change-aware proof selection and replay budget selftest detected"},
		{"kb-run-state-selftest", "KB run-state route-history guard selftest detected"},
		{"kb-work-ready-set-selftest", "KB work ready-set dispatch selftest detected"},
		{"kb-work-slice-lease-selftest", "KB work atomic slice lease selftest detected"},
		{"kb-work-scope-lease-selftest", "KB work scope lease overlap selftest detected"},
		{"kbrouter-catalog-tests", "KB model route catalog CLI conformance tests detected"},
		{"plan-worktree-lifecycle-selftest", "manifest-owned plan worktree lifecycle selftest detected"},
		{"kb-pipeline-selftest", "KB coded pipeline spike selftest detected"},
		{"skill-surface-report", "skill loaded-surface report detected"},
		{"skill-marketplace-firebreak", "private marketplace quarantine firebreak detected"},
		{"skill-marketplace-firebreak-selftest", "private marketplace quarantine firebreak negative selftest detected"},
		{"marketplace-promotion-selftest", "private marketplace safe promotion selftest detected"},
		{"kb-release-gate-selftest", "release gate profile selftest detected"},
		{"skill-surface-minimality-selftest", "skill/agent minimality classification selftest detected"},
		{"skill-surface-minimality", "static skill/agent minimality report detected"},
		{"cross-model-benchmark-validate", "cross-model benchmark prompt fixtures detected"},
		{"dishonest-completion-selftest", "dishonest completion rejection fixtures detected"},
		{"workflow-governor-selftest", "KB workflow governor question/phase gate contract detected"},
		{"context-packet-selftest", "context packet and usage telemetry contract detected"},
		{"graph-routing-lifecycle-selftest", "graph routing map-plan-work-review lifecycle contract detected"},
		{"graph-routing-eval", "graph routing correctness and local concurrency readiness fixtures detected"},
		{"execution-telemetry-selftest", "execution telemetry contract detected"},
		{"provider-hygiene", "optional provider and Phoenix activation hygiene detected"},
		{"provider-hygiene-selftest", "provider hygiene negative selftest detected"},
	}

	checks := make([]Check, 0, len(nativeChecks)+1)
	for _, pc := range nativeChecks {
		if pc.Name == "kb-work-ready-set-selftest" {
			checks = append(checks, Check{
				Name: "kb-work-ready-set-selftest", Args: []string{"kbcheck", "ready-set-selftest"},
				Reason: pc.Reason, Required: true, Confidence: "deterministic-local",
				Run: func(root string) CheckResult { return runNativeSelftest(runReadySetSelftest) },
			})
			continue
		}
		if pc.Name == "kb-work-scope-lease-selftest" {
			checks = append(checks, Check{
				Name: "kb-work-scope-lease-selftest", Args: []string{"kbcheck", "scope-lease-selftest"},
				Reason: pc.Reason, Required: true, Confidence: "deterministic-local",
				Run: func(root string) CheckResult { return runNativeSelftest(runScopeLeaseSelftest) },
			})
			continue
		}
		if pc.Name == "kb-work-slice-lease-selftest" {
			checks = append(checks, Check{
				Name: "kb-work-slice-lease-selftest", Args: []string{"kbcheck", "slice-lease-selftest"},
				Reason: pc.Reason, Required: true, Confidence: "deterministic-local",
				Run: func(root string) CheckResult { return runNativeSelftest(runSliceLeaseSelftest) },
			})
			continue
		}
		if pc.Name == "skill-lint" {
			checks = append(checks, Check{
				Name: "skill-lint", Args: []string{"kbcheck", "skill-lint"},
				Reason: pc.Reason, Required: true, Confidence: "deterministic-local",
				Run: func(root string) CheckResult { return runNativeCommand(root, []string{"skill-lint"}) },
			})
			continue
		}
		if pc.Name == "skill-guidance" {
			checks = append(checks, Check{
				Name: "skill-guidance", Args: []string{"kbcheck", "skill-guidance"},
				Reason: pc.Reason, Required: true, Confidence: "deterministic-local",
				Run: func(root string) CheckResult { return runNativeCommand(root, []string{"skill-guidance"}) },
			})
			continue
		}
		if pc.Name == "skill-marketplace-firebreak" {
			checks = append(checks, Check{
				Name: "skill-marketplace-firebreak", Args: []string{"kbcheck", "marketplace-firebreak"},
				Reason: pc.Reason, Required: true, Confidence: "deterministic-local",
				Run: func(root string) CheckResult { return runNativeCommand(root, []string{"marketplace-firebreak"}) },
			})
			continue
		}
		if pc.Name == "skill-marketplace-firebreak-selftest" {
			checks = append(checks, Check{
				Name: "skill-marketplace-firebreak-selftest", Args: []string{"kbcheck", "marketplace-firebreak-selftest"},
				Reason: pc.Reason, Required: true, Confidence: "deterministic-local",
				Run: func(root string) CheckResult {
					return runNativeCommand(root, []string{"marketplace-firebreak-selftest"})
				},
			})
			continue
		}
		nativeCommandByCheck := map[string][]string{
			"cross-model-benchmark-validate":    {"benchmark-validate"},
			"dishonest-completion-selftest":     {"dishonest-completion-selftest"},
			"kb-doctor-selftest":                {"doctor-selftest"},
			"route-complexity-eval":             {"route-eval"},
			"review-reference-guard":            {"review-reference-guard"},
			"skill-eval":                        {"skill-eval"},
			"skill-eval-quality":                {"skill-eval-quality"},
			"skill-eval-manifest-selftest":      {"skill-eval-manifest-selftest"},
			"skill-eval-baseline-selftest":      {"skill-eval-baseline-selftest"},
			"skill-eval-claims":                 {"skill-eval-claims"},
			"skill-eval-regression":             {"skill-eval-regression"},
			"skill-eval-codex-dry-run":          {"eval-run-codex", "--fixture-id", "tiny-typo-fix", "--dry-run"},
			"skill-eval-ghcp-dry-run":           {"eval-run-ghcp", "--fixture-id", "tiny-typo-fix", "--dry-run"},
			"manifest-contract-selftest":        {"manifest-contract-selftest"},
			"proof-governor-selftest":           {"proof-governor-selftest"},
			"kb-run-state-selftest":             {"run-state-selftest"},
			"kb-release-gate-selftest":          {"release-selftest"},
			"skill-surface-report":              {"surface-report"},
			"skill-surface-minimality":          {"minimality"},
			"skill-surface-minimality-selftest": {"minimality-selftest"},
			"kb-pipeline-selftest":              {"pipeline-selftest"},
			"marketplace-promotion-selftest":    {"marketplace-promote-selftest"},
			"workflow-governor-selftest":        {"workflow-governor-selftest"},
			"context-packet-selftest":           {"context-packet-selftest"},
			"graph-routing-lifecycle-selftest":  {"graph-routing-lifecycle-selftest"},
			"graph-routing-eval":                {"graph-routing-eval", "--require-ready"},
			"execution-telemetry-selftest":      {"execution-telemetry-selftest"},
			"provider-hygiene":                  {"provider-hygiene"},
			"provider-hygiene-selftest":         {"provider-hygiene-selftest"},
			"plan-worktree-lifecycle-selftest":  {"plan-worktree-selftest"},
		}
		if command, ok := nativeCommandByCheck[pc.Name]; ok {
			checks = append(checks, Check{
				Name: pc.Name, Args: append([]string{"kbcheck"}, command...),
				Reason: pc.Reason, Required: true, Confidence: "deterministic-local",
				Run: func(root string) CheckResult { return runNativeCommand(root, command) },
			})
			continue
		}
		if pc.Name == "kbrouter-catalog-tests" {
			checks = append(checks, Check{
				Name: pc.Name, Args: []string{"go", "test", "./cmd/kbrouter", "-run", "Catalog|Doctor|Policy"},
				Reason: pc.Reason, Required: true, Confidence: "deterministic-local",
			})
			continue
		}
	}
	checks = append(checks, Check{
		Name: "skill-eval-observed-trace-dry-run", Args: []string{"kbcheck", "skill-eval-wrap", "--fixture-id", "tiny-typo-fix", "--dry-run", "--sealed"},
		Reason: "observed trace eval wrapper detected", Required: true, Confidence: "deterministic-local",
		Run: func(root string) CheckResult {
			return runNativeCommand(root, []string{"skill-eval-wrap", "--fixture-id", "tiny-typo-fix", "--dry-run", "--sealed"})
		},
	})
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
