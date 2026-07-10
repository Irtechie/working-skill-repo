package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

const usage = `kbrouter routes KB model work without weakening proof gates.

Usage:
  kbrouter dispatch --run-root <path> --run-id <id> --slice-id <id> --packet <path> --route-alias <alias> --model <id> [--json]
  kbrouter models show [--user-root <path>] [--project-root <path>] [--json]
  kbrouter models discover --run-root <path> [--user-root <path>] [--current-model <id>] [--probe-openai-compatible] [--json]
  kbrouter models select --run-root <path> --run-id <id> --tier <small|medium|large> --task-family <id> --tool <id> --context-size <n> --risk <normal|broad> [--override use|require|ignore --alias <alias>] [--json]
  kbrouter models add --scope user|project [options]
  kbrouter models remove --scope user|project --alias <alias>
  kbrouter models prefer --scope user|project --alias <alias>
  kbrouter models clear --scope user|project [--alias <alias>]
  kbrouter models reset --scope user|project [--alias <alias>]
  kbrouter models approve --alias <alias> [--project-root <path>] [--expires-in <duration>]
  kbrouter models revoke --alias <alias> [--project-root <path>]
  kbrouter models deny --alias <alias> [--project-root <path>]
  kbrouter models ignore-routing --scope user|project
  kbrouter models doctor [--user-root <path>] [--project-root <path>] [--probe] [--json]
  kbrouter models calibrate --alias <alias> [--user-root <path>] [--json]
`

type approvalPrompt struct {
	ProjectPath      string
	ProjectID        string
	RouteAlias       string
	RouteFingerprint string
	Origin           string
	AuthEnv          string
	ExpiresAt        time.Time
}

var approvalConfirmer = confirmApprovalOnConsole

func confirmApprovalOnConsole(prompt approvalPrompt, output io.Writer) error {
	info, err := os.Stdin.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return fmt.Errorf("approval requires an attended interactive console; redirected approval is refused")
	}
	payload := strings.Join([]string{prompt.ProjectPath, prompt.ProjectID, prompt.RouteAlias, prompt.RouteFingerprint, prompt.Origin, prompt.AuthEnv, prompt.ExpiresAt.UTC().Format(time.RFC3339)}, "\x00")
	challenge := fmt.Sprintf("%x", sha256.Sum256([]byte(payload)))[:12]
	fmt.Fprintf(output, "ATTENDED MODEL ROUTE APPROVAL\nproject_path: %s\nproject_id: %s\nroute: %s\nroute_fingerprint: %s\nendpoint_origin: %s\nauth_environment: %s\nexpires_at: %s\nType APPROVE %s to continue: ",
		prompt.ProjectPath, prompt.ProjectID, prompt.RouteAlias, prompt.RouteFingerprint, prompt.Origin, prompt.AuthEnv, prompt.ExpiresAt.UTC().Format(time.RFC3339), challenge)
	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && len(answer) == 0 {
		return fmt.Errorf("approval confirmation failed: %w", err)
	}
	if strings.TrimSpace(answer) != "APPROVE "+challenge {
		return fmt.Errorf("approval was not confirmed")
	}
	return nil
}

func buildApprovalPrompt(projectRoot string, route modelrouting.Route, projectID, fingerprint string, expires time.Time) (approvalPrompt, error) {
	projectPath, err := filepath.Abs(filepath.Clean(projectRoot))
	if err != nil {
		return approvalPrompt{}, err
	}
	if canonical, canonicalErr := filepath.EvalSymlinks(projectPath); canonicalErr == nil {
		projectPath = canonical
	}
	return approvalPrompt{
		ProjectPath: projectPath, ProjectID: projectID, RouteAlias: route.Alias,
		RouteFingerprint: fingerprint, Origin: originForEndpoint(route.Endpoint),
		AuthEnv: route.AuthEnv, ExpiresAt: expires.UTC(),
	}, nil
}

func main() {
	code := run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(stdout, usage)
		return 0
	}
	if args[0] == "dispatch" {
		return runDispatch(args[1:], stdout, stderr)
	}
	if args[0] != "models" {
		fmt.Fprintf(stderr, "unsupported command %q\n", args[0])
		return 2
	}
	if len(args) < 2 {
		fmt.Fprint(stderr, "models requires a subcommand\n")
		return 2
	}
	switch args[1] {
	case "show":
		return runModelsShow(args[2:], stdout, stderr)
	case "discover":
		return runModelsDiscover(args[2:], stdout, stderr)
	case "select":
		return runModelsSelect(args[2:], stdout, stderr)
	case "add":
		return runModelsAdd(args[2:], stdout, stderr)
	case "remove":
		return runModelsRemove(args[2:], stdout, stderr)
	case "prefer":
		return runModelsPrefer(args[2:], stdout, stderr)
	case "clear":
		return runModelsClear(args[2:], stdout, stderr)
	case "reset":
		return runModelsClear(args[2:], stdout, stderr)
	case "approve":
		return runModelsApprove(args[2:], stdout, stderr)
	case "revoke":
		return runModelsRevoke(args[2:], stdout, stderr)
	case "deny":
		return runModelsDeny(args[2:], stdout, stderr)
	case "ignore-routing":
		return runModelsIgnoreRouting(args[2:], stdout, stderr)
	case "doctor":
		return runModelsDoctor(args[2:], stdout, stderr)
	case "calibrate":
		return runModelsCalibrate(args[2:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unsupported models subcommand %q\n", args[1])
		return 2
	}
}

func runModelsShow(args []string, stdout, stderr io.Writer) int {
	fs := flagSet("models show")
	opts := commonOptions{}
	opts.bind(fs)
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	userCatalog, userErr := loadUserCatalog(opts.userRoot)
	projectPolicy, projectErr := loadProjectPolicy(opts.projectRoot)
	report := map[string]any{
		"user_catalog":   redactCatalog(userCatalog),
		"project_policy": projectPolicy,
		"errors":         compactErrors(userErr, projectErr),
	}
	return printResult(stdout, stderr, report, opts.json, userErr, projectErr)
}

func runModelsDiscover(args []string, stdout, stderr io.Writer) int {
	fs := flagSet("models discover")
	opts := discoverOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.runRoot, "run-root", "", "explicit run root for the redacted catalog")
	fs.StringVar(&opts.currentModel, "current-model", "", "current orchestrator model id")
	fs.DurationVar(&opts.adapterTimeout, "adapter-timeout", 2*time.Second, "per-adapter deadline")
	fs.DurationVar(&opts.sessionTimeout, "session-timeout", 5*time.Second, "whole discovery deadline")
	fs.BoolVar(&opts.probeOpenAICompatible, "probe-openai-compatible", false, "probe trusted user-local OpenAI-compatible routes with GET /v1/models")
	if allowCustomUserRootForTests {
		fs.BoolVar(&opts.includeSlowFixture, "include-slow-fixture", false, "include a bounded slow adapter fixture")
		fs.StringVar(&opts.codexModelsFixture, "codex-models-fixture", "", "explicit Codex debug models JSON fixture")
	}
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if opts.probeOpenAICompatible && customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "credential-consuming discovery uses the fixed user-local trust root; custom --user-root is test-only")
		return 2
	}
	if opts.runRoot == "" {
		fmt.Fprintln(stderr, "discover requires --run-root; discovery never creates user or project catalog files by default")
		return 2
	}
	preparedRoot, err := prepareRunRoot(opts.projectRoot, opts.runRoot)
	if err != nil {
		fmt.Fprintf(stderr, "prepare run root: %v\n", err)
		return 2
	}
	report, err := discoverCatalog(opts)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := preparedRoot.revalidate(); err != nil {
		fmt.Fprintf(stderr, "revalidate run root: %v\n", err)
		return 1
	}
	if err := saveRunCatalog(opts.runRoot, report.Catalog); err != nil {
		fmt.Fprintf(stderr, "save run catalog: %v\n", err)
		return 1
	}
	if err := saveDispatchTrustedState(opts.userRoot, preparedRoot, report.Catalog); err != nil {
		fmt.Fprintf(stderr, "save dispatch state: %v\n", err)
		return 1
	}
	return printResult(stdout, stderr, report, opts.json, nil)
}

func runModelsAdd(args []string, stdout, stderr io.Writer) int {
	fs := flagSet("models add")
	opts := addOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.scope, "scope", "", "mutation scope: user or project")
	fs.StringVar(&opts.alias, "alias", "", "route alias")
	fs.StringVar(&opts.model, "model", "", "display model id")
	fs.StringVar(&opts.adapter, "adapter", "openai-compatible", "built-in adapter")
	fs.StringVar(&opts.dispatchMethod, "dispatch-method", "chat-completions", "built-in dispatch method")
	fs.StringVar(&opts.profile, "profile", "", "trusted user-local Codex profile name")
	fs.StringVar(&opts.destination, "destination", "", "destination label")
	fs.StringVar(&opts.endpoint, "endpoint", "", "provider endpoint")
	fs.StringVar(&opts.authEnv, "auth-env", "", "auth environment variable name")
	fs.StringVar(&opts.boundary, "boundary", "hosted", "hosted or private")
	fs.StringVar(&opts.retention, "retention", "unknown", "none, session, limited, or unknown")
	fs.StringVar(&opts.trainingUse, "training-use", "unknown", "no, yes, or unknown")
	fs.StringVar(&opts.residency, "residency", "unknown", "residency claim")
	fs.StringVar(&opts.trustProvenance, "trust-provenance", "", "trust metadata provenance")
	fs.StringVar(&opts.class, "class", "large", "declared capability class")
	fs.BoolVar(&opts.approveEndpoint, "approve-endpoint", false, "record local approval for this endpoint during validation")
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if opts.approveEndpoint && customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "endpoint/auth approvals use the fixed user-local trust root; custom --user-root is test-only")
		return 2
	}
	if opts.scope == "user" && customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "user-scoped model configuration uses the fixed user-local root; custom --user-root is test-only")
		return 2
	}
	switch opts.scope {
	case "user":
		route, err := routeFromAddOptions(opts)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		var approvalExpires time.Time
		var projectID, fingerprint string
		if opts.approveEndpoint {
			projectID, err = modelrouting.CanonicalProjectIdentity(opts.projectRoot)
			if err != nil {
				fmt.Fprintf(stderr, "canonical project identity: %v\n", err)
				return 1
			}
			fingerprint, err = modelrouting.ComputeRouteFingerprint(route)
			if err != nil {
				fmt.Fprintf(stderr, "route fingerprint: %v\n", err)
				return 1
			}
			approvalExpires = time.Now().Add(30 * 24 * time.Hour).UTC()
			prompt, promptErr := buildApprovalPrompt(opts.projectRoot, route, projectID, fingerprint, approvalExpires)
			if promptErr != nil {
				fmt.Fprintf(stderr, "approval prompt: %v\n", promptErr)
				return 1
			}
			if err := approvalConfirmer(prompt, stderr); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
		}
		if err := modelrouting.WithPrivateStateLock(opts.userRoot, func() error {
			catalog, loadErr := loadUserCatalog(opts.userRoot)
			if loadErr != nil {
				return fmt.Errorf("load user catalog: %w", loadErr)
			}
			catalog.Routes = upsertRoute(catalog.Routes, route)
			if saveErr := saveUserCatalog(opts.userRoot, catalog); saveErr != nil {
				return fmt.Errorf("save user catalog: %w", saveErr)
			}
			if opts.approveEndpoint {
				trust, trustErr := loadTrustFile(opts.userRoot)
				if trustErr != nil {
					return fmt.Errorf("load trust: %w", trustErr)
				}
				trust = approveRouteTrust(trust, projectID, route, fingerprint, approvalExpires)
				if saveErr := saveTrustFile(opts.userRoot, trust); saveErr != nil {
					return fmt.Errorf("save trust: %w", saveErr)
				}
			}
			return nil
		}); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return printResult(stdout, stderr, map[string]any{"scope": "user", "alias": route.Alias, "path": filepath.Join(opts.userRoot, "models.json")}, opts.json, nil)
	case "project":
		if opts.endpoint != "" || opts.authEnv != "" || opts.profile != "" || opts.dispatchMethod != "" && opts.dispatchMethod != "chat-completions" {
			fmt.Fprintln(stderr, "project scope cannot store connection details or profiles")
			return 2
		}
		if opts.alias == "" {
			fmt.Fprintln(stderr, "project add requires --alias")
			return 2
		}
		policy, err := loadProjectPolicy(opts.projectRoot)
		if err != nil {
			fmt.Fprintf(stderr, "load project policy: %v\n", err)
			return 1
		}
		policy.AllowedAliases = addUnique(policy.AllowedAliases, opts.alias)
		if err := saveProjectPolicy(opts.projectRoot, policy); err != nil {
			fmt.Fprintf(stderr, "save project policy: %v\n", err)
			return 1
		}
		return printResult(stdout, stderr, map[string]any{"scope": "project", "alias": opts.alias, "path": filepath.Join(opts.projectRoot, "kb-models.json")}, opts.json, nil)
	default:
		fmt.Fprintln(stderr, "add requires --scope user or --scope project")
		return 2
	}
}

func runModelsRemove(args []string, stdout, stderr io.Writer) int {
	fs := flagSet("models remove")
	opts := scopedAliasOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.scope, "scope", "", "mutation scope: user or project")
	fs.StringVar(&opts.alias, "alias", "", "route alias")
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if opts.alias == "" {
		fmt.Fprintln(stderr, "remove requires --alias")
		return 2
	}
	if opts.scope == "user" && customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "user-scoped model configuration uses the fixed user-local root; custom --user-root is test-only")
		return 2
	}
	switch opts.scope {
	case "user":
		if err := mutateUserCatalog(opts.userRoot, func(catalog *modelrouting.Catalog) {
			catalog.Routes = removeRoute(catalog.Routes, opts.alias)
		}); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	case "project":
		policy, err := loadProjectPolicy(opts.projectRoot)
		if err != nil {
			fmt.Fprintf(stderr, "load project policy: %v\n", err)
			return 1
		}
		policy.AllowedAliases = removeString(policy.AllowedAliases, opts.alias)
		policy.PreferredAliases = removeString(policy.PreferredAliases, opts.alias)
		if err := saveProjectPolicy(opts.projectRoot, policy); err != nil {
			fmt.Fprintf(stderr, "save project policy: %v\n", err)
			return 1
		}
	default:
		fmt.Fprintln(stderr, "remove requires --scope user or --scope project")
		return 2
	}
	return printResult(stdout, stderr, map[string]any{"scope": opts.scope, "removed": opts.alias}, opts.json, nil)
}

func runModelsClear(args []string, stdout, stderr io.Writer) int {
	fs := flagSet("models clear")
	opts := scopedAliasOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.scope, "scope", "", "mutation scope: user or project")
	fs.StringVar(&opts.alias, "alias", "", "optional alias to clear from preferences")
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if opts.scope == "user" && customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "user-scoped model configuration uses the fixed user-local root; custom --user-root is test-only")
		return 2
	}
	path, err := mutatePreferencePolicy(opts.scope, opts.userRoot, opts.projectRoot, func(policy *projectModelsPolicy) {
		if opts.alias != "" {
			policy.AllowedAliases = removeString(policy.AllowedAliases, opts.alias)
			policy.PreferredAliases = removeString(policy.PreferredAliases, opts.alias)
		} else {
			policy.AllowedAliases = nil
			policy.PreferredAliases = nil
			policy.IgnoreRouting = false
		}
	})
	if err != nil {
		fmt.Fprintf(stderr, "save clear: %v\n", err)
		return 1
	}
	return printResult(stdout, stderr, map[string]any{"scope": opts.scope, "cleared": opts.alias, "path": path}, opts.json, nil)
}

func runModelsApprove(args []string, stdout, stderr io.Writer) int {
	fs := flagSet("models approve")
	opts := approvalOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.alias, "alias", "", "route alias")
	fs.DurationVar(&opts.expiresIn, "expires-in", 24*time.Hour, "approval expiry duration")
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "trust approvals use the fixed user-local trust root; custom --user-root is test-only")
		return 2
	}
	if opts.alias == "" || opts.expiresIn <= 0 {
		fmt.Fprintln(stderr, "approve requires --alias and a positive --expires-in")
		return 2
	}
	route, projectID, fingerprint, err := routeApprovalInputs(opts.userRoot, opts.projectRoot, opts.alias)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	expires := time.Now().Add(opts.expiresIn).UTC()
	prompt, err := buildApprovalPrompt(opts.projectRoot, route, projectID, fingerprint, expires)
	if err != nil {
		fmt.Fprintf(stderr, "approval prompt: %v\n", err)
		return 1
	}
	if err := approvalConfirmer(prompt, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := modelrouting.WithPrivateStateLock(opts.userRoot, func() error {
		currentRoute, currentProjectID, currentFingerprint, currentErr := routeApprovalInputs(opts.userRoot, opts.projectRoot, opts.alias)
		if currentErr != nil {
			return currentErr
		}
		if currentProjectID != projectID || currentFingerprint != fingerprint {
			return fmt.Errorf("route changed while approval was pending; review and approve again")
		}
		trust, loadErr := loadTrustFile(opts.userRoot)
		if loadErr != nil {
			return fmt.Errorf("load trust: %w", loadErr)
		}
		trust = approveRouteTrust(trust, projectID, currentRoute, fingerprint, expires)
		return saveTrustFile(opts.userRoot, trust)
	}); err != nil {
		fmt.Fprintf(stderr, "save trust: %v\n", err)
		return 1
	}
	return printResult(stdout, stderr, map[string]any{"alias": opts.alias, "project_id": projectID, "route_fingerprint": fingerprint, "expires_at": expires}, opts.json, nil)
}

func runModelsRevoke(args []string, stdout, stderr io.Writer) int {
	return runModelsTrustRemoval(args, stdout, stderr, false)
}

func runModelsDeny(args []string, stdout, stderr io.Writer) int {
	return runModelsTrustRemoval(args, stdout, stderr, true)
}

func runModelsTrustRemoval(args []string, stdout, stderr io.Writer, deny bool) int {
	fs := flagSet("models trust")
	opts := scopedAliasOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.alias, "alias", "", "route alias")
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "trust approvals use the fixed user-local trust root; custom --user-root is test-only")
		return 2
	}
	if opts.alias == "" {
		fmt.Fprintln(stderr, "revoke/deny requires --alias")
		return 2
	}
	var projectID, fingerprint string
	if err := modelrouting.WithPrivateStateLock(opts.userRoot, func() error {
		route, resolvedProjectID, resolvedFingerprint, resolveErr := routeApprovalInputs(opts.userRoot, opts.projectRoot, opts.alias)
		if resolveErr != nil {
			return resolveErr
		}
		projectID, fingerprint = resolvedProjectID, resolvedFingerprint
		trust, loadErr := loadTrustFile(opts.userRoot)
		if loadErr != nil {
			return fmt.Errorf("load trust: %w", loadErr)
		}
		trust = removeRouteTrust(trust, projectID, route, fingerprint, deny)
		return saveTrustFile(opts.userRoot, trust)
	}); err != nil {
		fmt.Fprintf(stderr, "save trust: %v\n", err)
		return 1
	}
	key := "revoked"
	if deny {
		key = "denied"
	}
	return printResult(stdout, stderr, map[string]any{"alias": opts.alias, "project_id": projectID, "route_fingerprint": fingerprint, key: true}, opts.json, nil)
}

func runModelsPrefer(args []string, stdout, stderr io.Writer) int {
	fs := flagSet("models prefer")
	opts := scopedAliasOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.scope, "scope", "", "mutation scope: user or project")
	fs.StringVar(&opts.alias, "alias", "", "route alias")
	fs.StringVar(&opts.projectID, "project-id", "", "canonical project id")
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if opts.scope == "user" && customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "user-scoped model configuration uses the fixed user-local root; custom --user-root is test-only")
		return 2
	}
	if opts.alias == "" {
		fmt.Fprintln(stderr, "prefer requires --alias")
		return 2
	}
	path, err := mutatePreferencePolicy(opts.scope, opts.userRoot, opts.projectRoot, func(policy *projectModelsPolicy) {
		if opts.projectID != "" {
			policy.ProjectID = opts.projectID
		}
		policy.PreferredAliases = addUnique(policy.PreferredAliases, opts.alias)
	})
	if err != nil {
		fmt.Fprintf(stderr, "save preference: %v\n", err)
		return 1
	}
	return printResult(stdout, stderr, map[string]any{"scope": opts.scope, "preferred": opts.alias, "path": path}, opts.json, nil)
}

func runModelsIgnoreRouting(args []string, stdout, stderr io.Writer) int {
	fs := flagSet("models ignore-routing")
	opts := scopedAliasOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.scope, "scope", "", "mutation scope: user or project")
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if opts.scope == "user" && customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "user-scoped model configuration uses the fixed user-local root; custom --user-root is test-only")
		return 2
	}
	path, err := mutatePreferencePolicy(opts.scope, opts.userRoot, opts.projectRoot, func(policy *projectModelsPolicy) {
		policy.IgnoreRouting = true
	})
	if err != nil {
		fmt.Fprintf(stderr, "save ignore-routing: %v\n", err)
		return 1
	}
	return printResult(stdout, stderr, map[string]any{"scope": opts.scope, "ignore_routing": true, "path": path}, opts.json, nil)
}

func runModelsDoctor(args []string, stdout, stderr io.Writer) int {
	fs := flagSet("models doctor")
	opts := commonOptions{}
	opts.bind(fs)
	probe := fs.Bool("probe", false, "perform bounded endpoint and model-list probes")
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if *probe && customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "credential-consuming doctor probes use the fixed user-local trust root; custom --user-root is test-only")
		return 2
	}
	report, err := doctorReport(opts.userRoot, opts.projectRoot, *probe)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return printResult(stdout, stderr, report, opts.json, nil)
}

func runModelsCalibrate(args []string, stdout, stderr io.Writer) int {
	fs := flagSet("models calibrate")
	opts := scopedAliasOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.alias, "alias", "", "route alias")
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if opts.alias == "" {
		fmt.Fprintln(stderr, "calibrate requires --alias")
		return 2
	}
	report := map[string]any{
		"alias":   opts.alias,
		"mode":    "attended",
		"status":  "prepared",
		"message": "attended check prepared; no inference dispatched and no capability credit awarded",
	}
	return printResult(stdout, stderr, report, opts.json, nil)
}

type commonOptions struct {
	userRoot    string
	projectRoot string
	json        bool
}

func (o *commonOptions) bind(fs *flag.FlagSet) {
	home, err := operatingSystemUserHome()
	if err != nil {
		// An invalid path makes accidental fallback to the working directory fail
		// closed. Test-only custom roots are still available through the package
		// seam in catalog_test.go.
		home = string([]byte{0})
	}
	fs.StringVar(&o.userRoot, "user-root", filepath.Join(home, ".kb"), "user-local KB model root")
	fs.StringVar(&o.projectRoot, "project-root", ".", "project root")
	fs.BoolVar(&o.json, "json", false, "emit JSON")
}

type discoverOptions struct {
	commonOptions
	runRoot               string
	currentModel          string
	adapterTimeout        time.Duration
	sessionTimeout        time.Duration
	probeOpenAICompatible bool
	includeSlowFixture    bool
	codexModelsFixture    string
}

type addOptions struct {
	commonOptions
	scope           string
	alias           string
	model           string
	adapter         string
	dispatchMethod  string
	profile         string
	destination     string
	endpoint        string
	authEnv         string
	boundary        string
	retention       string
	trainingUse     string
	residency       string
	trustProvenance string
	class           string
	approveEndpoint bool
}

type scopedAliasOptions struct {
	commonOptions
	scope     string
	alias     string
	projectID string
}

type approvalOptions struct {
	commonOptions
	alias     string
	expiresIn time.Duration
}

func flagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func flagError(stderr io.Writer, err error) int {
	fmt.Fprintln(stderr, err)
	return 2
}

func printResult(stdout, stderr io.Writer, value any, jsonOut bool, errs ...error) int {
	for _, err := range errs {
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}
	if jsonOut {
		data, err := json.Marshal(value)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}
	fmt.Fprintf(stdout, "%v\n", value)
	return 0
}

func customUserRootRejected(fs *flag.FlagSet) bool {
	if allowCustomUserRootForTests {
		return false
	}
	changed := false
	fs.Visit(func(flag *flag.Flag) {
		if flag.Name == "user-root" {
			changed = true
		}
	})
	return changed
}

func compactErrors(errs ...error) []string {
	out := []string{}
	for _, err := range errs {
		if err != nil && !os.IsNotExist(err) {
			out = append(out, err.Error())
		}
	}
	return out
}
