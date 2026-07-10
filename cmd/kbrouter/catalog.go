package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

const (
	userCatalogFile             = "models.json"
	userPreferencesFile         = "preferences.json"
	userTrustFile               = "trust.json"
	projectPolicyFile           = "kb-models.json"
	maxCatalogBytes       int64 = 1 << 20
	codexPriorWindow            = 48 * time.Hour
	codexPriorContextSize       = 8192
)

var authEnvPattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]{0,127}$`)
var allowCustomUserRootForTests bool
var catalogNow = func() time.Time { return time.Now().UTC() }

type projectModelsPolicy struct {
	SchemaVersion    int      `json:"schema_version"`
	ProjectID        string   `json:"project_id,omitempty"`
	AllowedAliases   []string `json:"allowed_aliases,omitempty"`
	PreferredAliases []string `json:"preferred_aliases,omitempty"`
	IgnoreRouting    bool     `json:"ignore_routing,omitempty"`
}

type userTrustFileData struct {
	SchemaVersion int                `json:"schema_version"`
	Projects      []userProjectTrust `json:"projects,omitempty"`
}

type userProjectTrust struct {
	ProjectID         string                          `json:"project_id"`
	RouteApprovals    []modelrouting.RouteApproval    `json:"route_approvals,omitempty"`
	RouteDenials      []modelrouting.RouteDenial      `json:"route_denials,omitempty"`
	EndpointApprovals []modelrouting.EndpointApproval `json:"endpoint_approvals,omitempty"`
	AuthBindings      []modelrouting.AuthBinding      `json:"auth_bindings,omitempty"`
}

type discoveryReport struct {
	Catalog  modelrouting.Catalog `json:"catalog"`
	Adapters []adapterReport      `json:"adapters"`
}

type adapterReport struct {
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Message string   `json:"message,omitempty"`
	Models  []string `json:"models,omitempty"`
}

type doctorDimension struct {
	Status  string `json:"status"`
	Count   int    `json:"count,omitempty"`
	Message string `json:"message,omitempty"`
}

type doctorOutput struct {
	Discovery         doctorDimension `json:"discovery"`
	Configured        doctorDimension `json:"configured"`
	ProjectSelectable doctorDimension `json:"project_selectable"`
	Selectable        doctorDimension `json:"selectable"`
	Auth              doctorDimension `json:"auth"`
	Reachability      doctorDimension `json:"reachability"`
	ModelPresence     doctorDimension `json:"model_presence"`
	Dispatch          doctorDimension `json:"dispatch"`
	DispatchProven    doctorDimension `json:"dispatch_proven"`
	Control           doctorDimension `json:"control"`
}

type runRootMarker struct {
	SchemaVersion int    `json:"schema_version"`
	ProjectID     string `json:"project_id"`
	RunID         string `json:"run_id"`
}

type preparedRunRoot struct {
	projectPath  string
	kbPath       string
	basePath     string
	runPath      string
	marker       runRootMarker
	kbIdentity   os.FileInfo
	baseIdentity os.FileInfo
	runIdentity  os.FileInfo
}

type modelListFetcher func(context.Context, modelrouting.ValidatedEndpoint, modelrouting.Route, string, int64) ([]string, error)

var fetchOpenAICompatibleModels modelListFetcher = defaultOpenAICompatibleModelsFetcher

func loadUserCatalog(root string) (modelrouting.Catalog, error) {
	var catalog modelrouting.Catalog
	err := modelrouting.LoadStrictJSON(root, userCatalogFile, &catalog, maxCatalogBytes)
	if os.IsNotExist(err) {
		return modelrouting.Catalog{SchemaVersion: modelrouting.CatalogSchemaVersion}, nil
	}
	if err != nil {
		return modelrouting.Catalog{}, err
	}
	if err := modelrouting.ValidateCatalogStatic(catalog, modelrouting.CatalogSourceUser); err != nil {
		return modelrouting.Catalog{}, err
	}
	return catalog, nil
}

func saveUserCatalog(root string, catalog modelrouting.Catalog) error {
	if catalog.SchemaVersion == 0 {
		catalog.SchemaVersion = modelrouting.CatalogSchemaVersion
	}
	catalog.Fingerprint = ""
	sort.SliceStable(catalog.Routes, func(i, j int) bool { return catalog.Routes[i].Alias < catalog.Routes[j].Alias })
	if err := modelrouting.ValidateCatalogStatic(catalog, modelrouting.CatalogSourceUser); err != nil {
		return err
	}
	return modelrouting.SaveAtomicJSON(root, userCatalogFile, catalog, maxCatalogBytes)
}

func mutateUserCatalog(root string, mutate func(*modelrouting.Catalog)) error {
	return modelrouting.WithPrivateStateLock(root, func() error {
		catalog, err := loadUserCatalog(root)
		if err != nil {
			return fmt.Errorf("load user catalog: %w", err)
		}
		mutate(&catalog)
		if err := saveUserCatalog(root, catalog); err != nil {
			return fmt.Errorf("save user catalog: %w", err)
		}
		return nil
	})
}

func loadProjectPolicy(root string) (projectModelsPolicy, error) {
	var policy projectModelsPolicy
	err := modelrouting.LoadStrictProjectJSON(root, projectPolicyFile, &policy, maxCatalogBytes)
	if os.IsNotExist(err) {
		return projectModelsPolicy{SchemaVersion: 1}, nil
	}
	if err != nil {
		return projectModelsPolicy{}, err
	}
	if policy.SchemaVersion != 1 {
		return projectModelsPolicy{}, fmt.Errorf("unsupported project policy schema version %d", policy.SchemaVersion)
	}
	return policy, nil
}

func saveProjectPolicy(root string, policy projectModelsPolicy) error {
	policy.SchemaVersion = 1
	policy.AllowedAliases = sortedUnique(policy.AllowedAliases)
	policy.PreferredAliases = sortedUnique(policy.PreferredAliases)
	return modelrouting.SaveAtomicProjectJSON(root, projectPolicyFile, policy, maxCatalogBytes)
}

func loadPreferencePolicy(scope, userRoot, projectRoot string) (projectModelsPolicy, string, error) {
	switch scope {
	case "user":
		policy, err := loadUserPreferences(userRoot)
		return policy, filepath.Join(userRoot, userPreferencesFile), err
	case "project":
		policy, err := loadProjectPolicy(projectRoot)
		return policy, filepath.Join(projectRoot, projectPolicyFile), err
	default:
		return projectModelsPolicy{}, "", fmt.Errorf("requires --scope user or --scope project")
	}
}

func savePreferencePolicy(scope, userRoot, projectRoot string, policy projectModelsPolicy) error {
	switch scope {
	case "user":
		return saveUserPreferences(userRoot, policy)
	case "project":
		return saveProjectPolicy(projectRoot, policy)
	default:
		return fmt.Errorf("requires --scope user or --scope project")
	}
}

func mutatePreferencePolicy(scope, userRoot, projectRoot string, mutate func(*projectModelsPolicy)) (string, error) {
	if scope == "user" {
		path := filepath.Join(userRoot, userPreferencesFile)
		err := modelrouting.WithPrivateStateLock(userRoot, func() error {
			policy, loadErr := loadUserPreferences(userRoot)
			if loadErr != nil {
				return loadErr
			}
			mutate(&policy)
			return saveUserPreferences(userRoot, policy)
		})
		return path, err
	}
	if scope == "project" {
		policy, err := loadProjectPolicy(projectRoot)
		if err != nil {
			return "", err
		}
		mutate(&policy)
		return filepath.Join(projectRoot, projectPolicyFile), saveProjectPolicy(projectRoot, policy)
	}
	return "", fmt.Errorf("requires --scope user or project")
}

func loadUserPreferences(root string) (projectModelsPolicy, error) {
	var policy projectModelsPolicy
	err := modelrouting.LoadStrictJSON(root, userPreferencesFile, &policy, maxCatalogBytes)
	if os.IsNotExist(err) {
		return projectModelsPolicy{SchemaVersion: 1}, nil
	}
	if err != nil {
		return projectModelsPolicy{}, err
	}
	if policy.SchemaVersion != 1 {
		return projectModelsPolicy{}, fmt.Errorf("unsupported user preferences schema version %d", policy.SchemaVersion)
	}
	return policy, nil
}

func saveUserPreferences(root string, policy projectModelsPolicy) error {
	policy.SchemaVersion = 1
	policy.AllowedAliases = sortedUnique(policy.AllowedAliases)
	policy.PreferredAliases = sortedUnique(policy.PreferredAliases)
	return atomicWriteJSON(root, userPreferencesFile, policy)
}

func routeFromAddOptions(opts addOptions) (modelrouting.Route, error) {
	if opts.alias == "" || opts.model == "" || opts.destination == "" || opts.trustProvenance == "" {
		return modelrouting.Route{}, fmt.Errorf("user add requires --alias, --model, --destination, and --trust-provenance")
	}
	if opts.authEnv != "" && !authEnvPattern.MatchString(opts.authEnv) {
		return modelrouting.Route{}, fmt.Errorf("auth-env must be an environment variable name")
	}
	routeID, err := newRouteID()
	if err != nil {
		return modelrouting.Route{}, err
	}
	route := modelrouting.Route{
		RouteID:         routeID,
		Alias:           opts.alias,
		DisplayModelID:  opts.model,
		Adapter:         opts.adapter,
		AdapterRevision: "v1",
		DispatchMethod:  opts.dispatchMethod,
		Profile:         opts.profile,
		Destination:     opts.destination,
		Endpoint:        opts.endpoint,
		AuthEnv:         opts.authEnv,
		Boundary:        modelrouting.TrustBoundary(opts.boundary),
		Retention:       modelrouting.RetentionClass(opts.retention),
		TrainingUse:     modelrouting.TrainingUse(opts.trainingUse),
		Residency:       opts.residency,
		TrustProvenance: opts.trustProvenance,
		Readiness:       []modelrouting.Readiness{modelrouting.ReadinessDiscovered, modelrouting.ReadinessConfigured, modelrouting.ReadinessSelectable},
		Capability: modelrouting.CapabilityEvidence{
			Class:          modelrouting.CapabilityClass(opts.class),
			Source:         modelrouting.EvidenceDeclared,
			RouteAlias:     opts.alias,
			ModelID:        opts.model,
			TaskFamily:     "code",
			Risk:           modelrouting.RiskNormal,
			DispatchProven: false,
		},
	}
	if route.Profile != "" {
		codexHome, err := dispatchCodexHome()
		if err != nil {
			return modelrouting.Route{}, err
		}
		revision, err := trustedCodexProfileRevision(codexHome, route.Profile, route.Endpoint, route.AuthEnv)
		if err != nil {
			return modelrouting.Route{}, err
		}
		route.ProfileRevision = revision
	}
	if err := modelrouting.ValidateCatalogStatic(modelrouting.Catalog{SchemaVersion: modelrouting.CatalogSchemaVersion, Routes: []modelrouting.Route{route}}, modelrouting.CatalogSourceUser); err != nil {
		return modelrouting.Route{}, err
	}
	return route, nil
}

func newRouteID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate route id: %w", err)
	}
	return "route-id:" + hex.EncodeToString(value), nil
}

func discoverCatalog(opts discoverOptions) (discoveryReport, error) {
	if opts.sessionTimeout <= 0 {
		opts.sessionTimeout = 5 * time.Second
	}
	if opts.adapterTimeout <= 0 {
		opts.adapterTimeout = 2 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), opts.sessionTimeout)
	defer cancel()

	userCatalog, err := loadUserCatalog(opts.userRoot)
	userCatalogErr := err
	if userCatalogErr != nil {
		userCatalog = modelrouting.Catalog{SchemaVersion: modelrouting.CatalogSchemaVersion}
	}
	projectPolicy, projectPolicyErr := policyContextForProject(opts.userRoot, opts.projectRoot)
	report := discoveryReport{
		Catalog: modelrouting.Catalog{
			SchemaVersion: modelrouting.CatalogSchemaVersion,
			Cohort:        modelrouting.CohortInitialPilot,
			Surfaces:      localSurfaceFingerprints(userCatalog, projectPolicy),
			Current:       currentModel(opts.currentModel),
			Routes:        redactedUserRoutes(userCatalog.Routes),
		},
	}
	if projectPolicyErr != nil {
		report.Adapters = append(report.Adapters, adapterReport{Name: "project-policy", Status: "unavailable", Message: projectPolicyErr.Error()})
	}
	if userCatalogErr != nil {
		report.Adapters = append(report.Adapters, adapterReport{Name: "user-catalog", Status: "unavailable", Message: userCatalogErr.Error()})
	}
	adapters := []discoveryAdapter{
		{name: "current", run: func(ctx context.Context) adapterDiscovery { return discoverCurrent(ctx, opts.currentModel) }},
		{name: "codex", run: func(ctx context.Context) adapterDiscovery {
			return discoverCodex(ctx, opts.projectRoot, opts.runRoot, opts.codexModelsFixture)
		}},
	}
	if opts.probeOpenAICompatible && projectPolicyErr == nil {
		adapters = append(adapters, discoveryAdapter{name: "openai-compatible", run: func(ctx context.Context) adapterDiscovery {
			return discoverOpenAICompatible(ctx, userCatalog, projectPolicy)
		}})
	}
	if opts.includeSlowFixture {
		adapters = append(adapters, discoveryAdapter{name: "slow-fixture", run: discoverSlowFixture})
	}

	results := make(chan namedAdapterDiscovery, len(adapters))
	pending := make(map[string]struct{}, len(adapters))
	for _, adapter := range adapters {
		adapter := adapter
		pending[adapter.name] = struct{}{}
		go func() {
			adapterCtx, adapterCancel := context.WithTimeout(ctx, opts.adapterTimeout)
			defer adapterCancel()
			results <- namedAdapterDiscovery{name: adapter.name, result: adapter.run(adapterCtx)}
		}()
	}

	for range adapters {
		select {
		case completed := <-results:
			delete(pending, completed.name)
			report.Adapters = append(report.Adapters, completed.result.Report)
			for _, route := range completed.result.Routes {
				report.Catalog.Routes = upsertRoute(report.Catalog.Routes, route)
			}
			if completed.result.Surface != nil {
				report.Catalog.Surfaces = append(report.Catalog.Surfaces, *completed.result.Surface)
			}
		case <-ctx.Done():
			names := make([]string, 0, len(pending))
			for name := range pending {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				report.Adapters = append(report.Adapters, adapterReport{Name: name, Status: "unavailable", Message: ctx.Err().Error()})
			}
			sort.SliceStable(report.Adapters, func(i, j int) bool { return report.Adapters[i].Name < report.Adapters[j].Name })
			fingerprintAndSort(&report.Catalog)
			return report, nil
		}
	}
	sort.SliceStable(report.Adapters, func(i, j int) bool { return report.Adapters[i].Name < report.Adapters[j].Name })
	fingerprintAndSort(&report.Catalog)
	return report, nil
}

type discoveryAdapter struct {
	name string
	run  func(context.Context) adapterDiscovery
}

type namedAdapterDiscovery struct {
	name   string
	result adapterDiscovery
}

type adapterDiscovery struct {
	Report  adapterReport
	Routes  []modelrouting.Route
	Surface *modelrouting.SurfaceFingerprint
}

type codexModelInfo struct {
	ID         string
	Capability *codexSurfaceCapability
}

type codexSurfaceCapability struct {
	Class       modelrouting.CapabilityClass
	TaskFamily  string
	Tools       []string
	ContextSize int
	Risk        modelrouting.RiskLevel
}

type codexExecutablePrior struct {
	Executable trustedExecutable
	Version    string
	Revision   string
}

func discoverCurrent(ctx context.Context, model string) adapterDiscovery {
	select {
	case <-ctx.Done():
		return adapterDiscovery{Report: adapterReport{Name: "current", Status: "unavailable", Message: ctx.Err().Error()}}
	default:
	}
	if model == "" {
		return adapterDiscovery{Report: adapterReport{Name: "current", Status: "unavailable", Message: "current model not provided"}}
	}
	return adapterDiscovery{
		Report:  adapterReport{Name: "current", Status: "available", Models: []string{model}},
		Routes:  []modelrouting.Route{redactRoute(currentRoute(model))},
		Surface: &modelrouting.SurfaceFingerprint{Surface: "current", Provider: "active-orchestrator", Revision: "v1", ConfigHash: "sha256:" + sha256Text(model)},
	}
}

func discoverCodex(ctx context.Context, projectRoot, runRoot, fixture string) adapterDiscovery {
	var executable trustedExecutable
	var executableOK bool
	if fixture == "" {
		var err error
		executable, err = resolveTrustedCodexExecutable(projectRoot, runRoot)
		if err != nil {
			return adapterDiscovery{Report: adapterReport{Name: "codex", Status: "unavailable", Message: err.Error()}}
		}
		executableOK = true
	}
	data, err := codexModelsJSON(ctx, fixture, executable.Path)
	if err != nil {
		return adapterDiscovery{Report: adapterReport{Name: "codex", Status: "unavailable", Message: err.Error()}}
	}
	entries, err := parseCodexModels(data)
	if err != nil {
		return adapterDiscovery{Report: adapterReport{Name: "codex", Status: "unavailable", Message: err.Error()}}
	}
	models := make([]string, 0, len(entries))
	routes := make([]modelrouting.Route, 0, len(entries))
	for _, entry := range entries {
		models = append(models, entry.ID)
		route := redactedDiscoveredRoute("codex."+safeAlias(entry.ID), entry.ID, "codex", "exec-model", "codex")
		if entry.Capability != nil {
			route.TrustProvenance = "codex surface capability metadata"
			route.Capability.Class = entry.Capability.Class
			route.Capability.TaskFamily = entry.Capability.TaskFamily
			route.Capability.Tools = append([]string(nil), entry.Capability.Tools...)
			route.Capability.ContextSize = entry.Capability.ContextSize
			route.Capability.Risk = entry.Capability.Risk
		}
		routes = append(routes, route)
	}
	revision := "v1"
	var prior *codexExecutablePrior
	if fixture == "" && executableOK {
		if candidate, priorErr := codexExecutablePriorIdentity(ctx, executable); priorErr == nil {
			prior = &candidate
			revision = candidate.Revision
		}
	}
	return adapterDiscovery{
		Report:  adapterReport{Name: "codex", Status: "available", Models: models},
		Routes:  codexDiscoveredRoutes(routes, prior, catalogNow()),
		Surface: &modelrouting.SurfaceFingerprint{Surface: "codex-cli", Provider: "openai", Revision: revision, ConfigHash: "sha256:" + sha256Text(strings.Join(models, "\x00"))},
	}
}

func codexDiscoveredRoutes(routes []modelrouting.Route, prior *codexExecutablePrior, now time.Time) []modelrouting.Route {
	out := make([]modelrouting.Route, 0, len(routes))
	for _, route := range routes {
		if prior != nil && codexRouteHasExactSurfaceCapability(route) {
			route = codexAdapterPriorRoute(route, prior, now)
		}
		out = append(out, route)
	}
	return out
}

func codexRouteHasExactSurfaceCapability(route modelrouting.Route) bool {
	return route.TrustProvenance == "codex surface capability metadata" &&
		route.Capability.TaskFamily != "" &&
		len(route.Capability.Tools) > 0 &&
		route.Capability.ContextSize > 0 &&
		route.Capability.Risk != ""
}

func codexAdapterPriorRoute(route modelrouting.Route, prior *codexExecutablePrior, now time.Time) modelrouting.Route {
	route.Readiness = []modelrouting.Readiness{
		modelrouting.ReadinessDiscovered,
		modelrouting.ReadinessConfigured,
		modelrouting.ReadinessSelectable,
		modelrouting.ReadinessDispatchProven,
	}
	route.AdapterRevision = prior.Revision
	route.TrustProvenance = "codex CLI " + prior.Version + " exact executable " + prior.Executable.Hash
	route.Capability.Source = modelrouting.EvidenceAdapterPrior
	route.Capability.DispatchProven = true
	route.Capability.ExpiresAt = codexAdapterPriorExpiry(now)
	return route
}

func codexAdapterPriorExpiry(now time.Time) time.Time {
	return now.UTC().Truncate(24 * time.Hour).Add(codexPriorWindow)
}

func codexExecutablePriorIdentity(ctx context.Context, executable trustedExecutable) (codexExecutablePrior, error) {
	if runtime.GOOS != "windows" {
		return codexExecutablePrior{}, fmt.Errorf("Codex adapter prior is unavailable without Windows job-object containment")
	}
	if err := dispatchProcessTreeContainment(); err != nil {
		return codexExecutablePrior{}, err
	}
	if err := executable.revalidate(); err != nil {
		return codexExecutablePrior{}, err
	}
	versionCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	version, err := codexVersionForExecutable(versionCtx, executable.Path)
	if err != nil {
		return codexExecutablePrior{}, err
	}
	contractCtx, contractCancel := context.WithTimeout(ctx, 5*time.Second)
	defer contractCancel()
	if err := codexAdapterContractProbe(contractCtx, executable.Path); err != nil {
		return codexExecutablePrior{}, err
	}
	revision, err := codexAdapterRevision(executable.Hash, version)
	if err != nil {
		return codexExecutablePrior{}, err
	}
	return codexExecutablePrior{Executable: executable, Version: version, Revision: revision}, nil
}

func codexAdapterRevision(executableHash, version string) (string, error) {
	hash := strings.TrimPrefix(executableHash, "sha256:")
	if len(hash) != 64 {
		return "", fmt.Errorf("invalid executable hash")
	}
	versionPart := safeAlias(version)
	if len(versionPart) > 40 {
		versionPart = versionPart[:40]
	}
	revision := "codex-cli-v1:" + hash + ":" + versionPart
	if len(revision) > 128 {
		return "", fmt.Errorf("codex adapter revision exceeded bound")
	}
	return revision, nil
}

func codexVersion(ctx context.Context) (string, error) {
	return codexVersionForExecutable(ctx, "codex")
}

func codexVersionForExecutable(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, path, "--version")
	buffer := &boundedBuffer{limit: 256}
	cmd.Stdout = buffer
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return "", err
	}
	if buffer.exceeded {
		return "", fmt.Errorf("codex version output exceeded bound")
	}
	return normalizeModelID(buffer.String())
}

func codexAdapterContractProbe(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, path, "exec", "--help")
	buffer := &boundedBuffer{limit: 4096}
	cmd.Stdout = buffer
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("codex exec contract probe failed: %w", err)
	}
	if buffer.exceeded {
		return fmt.Errorf("codex exec contract output exceeded bound")
	}
	help := buffer.String()
	for _, required := range []string{"--model", "--sandbox", "-C", "--add-dir", "-c", "--output-schema", "--json"} {
		if !strings.Contains(help, required) {
			return fmt.Errorf("codex exec contract missing %s", required)
		}
	}
	return nil
}

func codexModelsJSON(ctx context.Context, fixture, executablePath string) ([]byte, error) {
	if fixture != "" {
		return readBoundedFile(fixture, maxCatalogBytes)
	}
	if strings.TrimSpace(executablePath) == "" {
		return nil, fmt.Errorf("codex CLI unavailable")
	}
	help := exec.CommandContext(ctx, executablePath, "debug", "models", "--help")
	if err := help.Run(); err != nil {
		return nil, fmt.Errorf("codex model enumeration unavailable: %w", err)
	}
	cmd := exec.CommandContext(ctx, executablePath, "debug", "models")
	buffer := &boundedBuffer{limit: maxCatalogBytes}
	cmd.Stdout = buffer
	cmd.Stderr = io.Discard
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("codex debug models failed: %w", err)
	}
	if buffer.exceeded {
		return nil, fmt.Errorf("codex catalog response exceeded bound")
	}
	return buffer.Bytes(), nil
}

func parseCodexModels(data []byte) ([]codexModelInfo, error) {
	var payload struct {
		Models []struct {
			Slug          string                   `json:"slug"`
			DisplayName   string                   `json:"display_name"`
			Description   string                   `json:"description"`
			ContextWindow int                      `json:"context_window"`
			Class         string                   `json:"class"`
			Capability    codexCapabilityPayload   `json:"capability"`
			KBCapability  codexCapabilityPayload   `json:"kb_capability"`
			Capabilities  []codexCapabilityPayload `json:"capabilities"`
		} `json:"models"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}
	models := []codexModelInfo{}
	for _, model := range payload.Models {
		id := strings.TrimSpace(model.Slug)
		if id == "" {
			id = strings.TrimSpace(model.DisplayName)
		}
		if id != "" {
			normalized, err := normalizeModelID(id)
			if err != nil {
				return nil, err
			}
			candidates := []codexCapabilityPayload{model.Capability, model.KBCapability}
			candidates = append(candidates, model.Capabilities...)
			capability := parseCodexSurfaceCapability(model.Class, candidates...)
			if capability == nil {
				capability = providerCodexSurfaceCapability(model.Description, model.ContextWindow)
			}
			models = append(models, codexModelInfo{ID: normalized, Capability: capability})
		}
	}
	sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
	if len(models) == 0 {
		return nil, fmt.Errorf("codex catalog contained no models")
	}
	return models, nil
}

type codexCapabilityPayload struct {
	Class       string   `json:"class"`
	TaskFamily  string   `json:"task_family"`
	Tools       []string `json:"tools"`
	ContextSize int      `json:"context_size"`
	Risk        string   `json:"risk"`
}

func parseCodexSurfaceCapability(topLevelClass string, candidates ...codexCapabilityPayload) *codexSurfaceCapability {
	for _, candidate := range candidates {
		if capability := exactCodexSurfaceCapability(candidate); capability != nil {
			return capability
		}
	}
	candidate := codexCapabilityPayload{Class: topLevelClass}
	return exactCodexSurfaceCapability(candidate)
}

func exactCodexSurfaceCapability(candidate codexCapabilityPayload) *codexSurfaceCapability {
	class := modelrouting.CapabilityClass(strings.TrimSpace(candidate.Class))
	if class != modelrouting.ClassSmall && class != modelrouting.ClassMedium && class != modelrouting.ClassLarge && class != modelrouting.ClassPlanner {
		return nil
	}
	taskFamily := strings.TrimSpace(candidate.TaskFamily)
	if taskFamily == "" || len(candidate.Tools) == 0 || candidate.ContextSize <= 0 {
		return nil
	}
	risk := modelrouting.RiskLevel(strings.TrimSpace(candidate.Risk))
	if risk != modelrouting.RiskNormal && risk != modelrouting.RiskBroad {
		return nil
	}
	tools := sortedUnique(candidate.Tools)
	if len(tools) == 0 {
		return nil
	}
	return &codexSurfaceCapability{Class: class, TaskFamily: taskFamily, Tools: tools, ContextSize: candidate.ContextSize, Risk: risk}
}

func providerCodexSurfaceCapability(description string, contextWindow int) *codexSurfaceCapability {
	if contextWindow <= 0 {
		return nil
	}
	capability := &codexSurfaceCapability{
		TaskFamily:  "code",
		Tools:       []string{"codex-harness"},
		ContextSize: contextWindow,
		Risk:        modelrouting.RiskNormal,
	}
	switch strings.TrimSpace(description) {
	case "Small, fast, and cost-efficient model for simpler coding tasks.":
		capability.Class = modelrouting.ClassSmall
	case "Strong model for everyday coding.":
		capability.Class = modelrouting.ClassMedium
	case "Frontier model for complex coding, research, and real-world work.":
		capability.Class = modelrouting.ClassLarge
		capability.Risk = modelrouting.RiskBroad
	default:
		return nil
	}
	return capability
}

func discoverOpenAICompatible(ctx context.Context, catalog modelrouting.Catalog, policy modelrouting.PolicyContext) adapterDiscovery {
	routes := []modelrouting.Route{}
	models := []string{}
	for _, route := range catalog.Routes {
		if route.Adapter != "openai-compatible" || route.Endpoint == "" {
			continue
		}
		if !routeProjectSelectable(route, policy) {
			continue
		}
		validated, err := modelrouting.ValidateEndpointContext(ctx, route, policy, nil, time.Now())
		if err != nil {
			return adapterDiscovery{Report: adapterReport{Name: "openai-compatible", Status: "unavailable", Message: err.Error()}}
		}
		token := ""
		if route.AuthEnv != "" {
			token = os.Getenv(route.AuthEnv)
			if token == "" {
				return adapterDiscovery{Report: adapterReport{Name: "openai-compatible", Status: "unavailable", Message: "auth environment variable is not set"}}
			}
		}
		found, err := fetchOpenAICompatibleModels(ctx, validated, route, token, maxCatalogBytes)
		if err != nil {
			return adapterDiscovery{Report: adapterReport{Name: "openai-compatible", Status: "unavailable", Message: err.Error()}}
		}
		for _, model := range found {
			models = append(models, model)
			routes = append(routes, redactedChildRoute(route, model))
		}
	}
	sort.Strings(models)
	if len(models) == 0 {
		return adapterDiscovery{Report: adapterReport{Name: "openai-compatible", Status: "unavailable", Message: "no trusted configured route with models"}}
	}
	return adapterDiscovery{Report: adapterReport{Name: "openai-compatible", Status: "available", Models: models}, Routes: routes}
}

func defaultOpenAICompatibleModelsFetcher(ctx context.Context, endpoint modelrouting.ValidatedEndpoint, route modelrouting.Route, token string, maxBytes int64) ([]string, error) {
	target := *endpoint.URL
	target.Path = strings.TrimRight(target.Path, "/") + "/models"
	target.RawQuery = ""
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				host = address
			}
			if !strings.EqualFold(host, endpoint.URL.Hostname()) {
				return nil, fmt.Errorf("cross-origin dial refused")
			}
			if port == "" {
				if endpoint.URL.Scheme == "https" {
					port = "443"
				} else {
					port = "80"
				}
			}
			dialer := net.Dialer{}
			var lastErr error
			for _, ip := range endpoint.PinnedIPs {
				conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
				if err == nil {
					return conn, nil
				}
				lastErr = err
			}
			return nil, lastErr
		},
		TLSClientConfig: &tls.Config{ServerName: endpoint.TLSServerName, MinVersion: tls.VersionTLS12},
	}
	client := http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET /v1/models returned %s", resp.Status)
	}
	limited := io.LimitReader(resp.Body, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("model list response exceeded bound")
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	models := []string{}
	for _, model := range payload.Data {
		if strings.TrimSpace(model.ID) != "" {
			normalized, err := normalizeModelID(model.ID)
			if err != nil {
				return nil, err
			}
			models = append(models, normalized)
		}
	}
	return models, nil
}

type boundedBuffer struct {
	bytes.Buffer
	limit    int64
	exceeded bool
}

func (b *boundedBuffer) Write(value []byte) (int, error) {
	originalLength := len(value)
	remaining := b.limit - int64(b.Len())
	if remaining <= 0 {
		b.exceeded = true
		return originalLength, nil
	}
	if int64(len(value)) > remaining {
		value = value[:remaining]
		b.exceeded = true
	}
	_, _ = b.Buffer.Write(value)
	return originalLength, nil
}

func readBoundedFile(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("fixture exceeded bound")
	}
	return data, nil
}

func normalizeModelID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 256 || !utf8.ValidString(value) {
		return "", fmt.Errorf("invalid model id")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return "", fmt.Errorf("invalid model id")
		}
	}
	return value, nil
}

func discoverSlowFixture(ctx context.Context) adapterDiscovery {
	select {
	case <-time.After(10 * time.Second):
		return adapterDiscovery{Report: adapterReport{Name: "slow-fixture", Status: "available"}}
	case <-ctx.Done():
		return adapterDiscovery{Report: adapterReport{Name: "slow-fixture", Status: "unavailable", Message: ctx.Err().Error()}}
	}
}

func saveRunCatalog(root string, catalog modelrouting.Catalog) error {
	return modelrouting.SaveCatalog(root, "catalog.json", catalog, modelrouting.StorageOptions{
		MaxBytes: maxCatalogBytes,
		Source:   modelrouting.CatalogSourceRun,
	})
}

func prepareRunRoot(projectRoot, runRoot string) (preparedRunRoot, error) {
	projectPath, err := filepath.Abs(filepath.Clean(projectRoot))
	if err != nil {
		return preparedRunRoot{}, err
	}
	if canonical, canonicalErr := filepath.EvalSymlinks(projectPath); canonicalErr == nil {
		projectPath = canonical
	}
	runPath, err := filepath.Abs(filepath.Clean(runRoot))
	if err != nil {
		return preparedRunRoot{}, err
	}
	kbPath := filepath.Join(projectPath, ".kb")
	allowedBase := filepath.Join(kbPath, "runs")
	relative, err := filepath.Rel(allowedBase, runPath)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.Dir(relative) != "." {
		return preparedRunRoot{}, fmt.Errorf("run root must be one dedicated direct child of %s", allowedBase)
	}
	kbIdentity, err := ensureRealDirectory(kbPath, 0o755)
	if err != nil {
		return preparedRunRoot{}, fmt.Errorf("prepare .kb directory: %w", err)
	}
	baseIdentity, err := ensureRealDirectory(allowedBase, 0o755)
	if err != nil {
		return preparedRunRoot{}, fmt.Errorf("prepare runs directory: %w", err)
	}
	if err := requireCanonicalDirectory(allowedBase); err != nil {
		return preparedRunRoot{}, err
	}
	projectID, err := modelrouting.CanonicalProjectIdentity(projectPath)
	if err != nil {
		return preparedRunRoot{}, err
	}
	marker := runRootMarker{SchemaVersion: 1, ProjectID: projectID, RunID: relative}
	info, statErr := os.Lstat(runPath)
	switch {
	case os.IsNotExist(statErr):
		if err := modelrouting.SaveAtomicJSON(runPath, ".kb-run-root.json", marker, maxCatalogBytes); err != nil {
			return preparedRunRoot{}, err
		}
	case statErr != nil:
		return preparedRunRoot{}, statErr
	case info.Mode()&os.ModeSymlink != 0 || !info.IsDir():
		return preparedRunRoot{}, modelrouting.ErrUnsafePath
	default:
		var existing runRootMarker
		if err := modelrouting.LoadStrictJSON(runPath, ".kb-run-root.json", &existing, maxCatalogBytes); err != nil {
			return preparedRunRoot{}, fmt.Errorf("existing run root is not a marked private KB run directory: %w", err)
		}
		if existing != marker {
			return preparedRunRoot{}, fmt.Errorf("run root marker does not match this project and run id")
		}
	}
	if err := requireCanonicalDirectory(runPath); err != nil {
		return preparedRunRoot{}, err
	}
	runIdentity, err := os.Lstat(runPath)
	if err != nil || runIdentity.Mode()&os.ModeSymlink != 0 || !runIdentity.IsDir() {
		return preparedRunRoot{}, modelrouting.ErrUnsafePath
	}
	prepared := preparedRunRoot{projectPath: projectPath, kbPath: kbPath, basePath: allowedBase, runPath: runPath, marker: marker, kbIdentity: kbIdentity, baseIdentity: baseIdentity, runIdentity: runIdentity}
	if err := prepared.revalidate(); err != nil {
		return preparedRunRoot{}, err
	}
	return prepared, nil
}

func ensureRealDirectory(path string, mode os.FileMode) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		if mkdirErr := os.Mkdir(path, mode); mkdirErr != nil && !os.IsExist(mkdirErr) {
			return nil, mkdirErr
		}
		info, err = os.Lstat(path)
	}
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil, modelrouting.ErrUnsafePath
	}
	return info, nil
}

func requireCanonicalDirectory(path string) error {
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil {
		return modelrouting.ErrUnsafePath
	}
	absCanonical, err := filepath.Abs(canonical)
	if err != nil || !sameFilesystemPath(absCanonical, path) {
		return modelrouting.ErrUnsafePath
	}
	return nil
}

func sameFilesystemPath(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

func (prepared preparedRunRoot) revalidate() error {
	for _, expected := range []struct {
		path string
		info os.FileInfo
	}{{prepared.kbPath, prepared.kbIdentity}, {prepared.basePath, prepared.baseIdentity}, {prepared.runPath, prepared.runIdentity}} {
		observed, err := os.Lstat(expected.path)
		if err != nil || observed.Mode()&os.ModeSymlink != 0 || !observed.IsDir() || !os.SameFile(expected.info, observed) {
			return modelrouting.ErrUnsafePath
		}
		if err := requireCanonicalDirectory(expected.path); err != nil {
			return err
		}
	}
	var marker runRootMarker
	if err := modelrouting.LoadStrictJSON(prepared.runPath, ".kb-run-root.json", &marker, maxCatalogBytes); err != nil {
		return err
	}
	if marker != prepared.marker {
		return modelrouting.ErrUnsafePath
	}
	return nil
}

func doctorReport(userRoot, projectRoot string, probe bool) (doctorOutput, error) {
	catalog, err := loadUserCatalog(userRoot)
	if err != nil {
		return doctorOutput{}, fmt.Errorf("load user catalog: %w", err)
	}
	policy, projectErr := policyContextForProject(userRoot, projectRoot)
	selectable := 0
	projectSelectable := 0
	missingAuth := 0
	reachable := 0
	modelPresent := 0
	probed := 0
	probeMessage := "not probed"
	modelMessage := "not probed"
	var probeContext context.Context = context.Background()
	cancelProbeSession := func() {}
	if probe {
		probeContext, cancelProbeSession = context.WithTimeout(context.Background(), 5*time.Second)
	}
	defer cancelProbeSession()
	for _, route := range catalog.Routes {
		if hasReadiness(route.Readiness, modelrouting.ReadinessSelectable) {
			selectable++
		}
		if routeProjectSelectable(route, policy) {
			projectSelectable++
		}
		if route.AuthEnv != "" && (os.Getenv(route.AuthEnv) == "" || !routeAuthApproved(route, policy, time.Now())) {
			missingAuth++
		}
		if probe && route.Endpoint != "" {
			if projectErr != nil || !routeProjectSelectable(route, policy) {
				if projectErr != nil {
					probeMessage = "project policy unavailable; configured probes refused"
					modelMessage = probeMessage
				}
				continue
			}
			if probeContext.Err() != nil {
				probeMessage = probeContext.Err().Error()
				modelMessage = probeContext.Err().Error()
				break
			}
			probed++
			ctx, cancel := context.WithTimeout(probeContext, 2*time.Second)
			validated, err := modelrouting.ValidateEndpointContext(ctx, route, policy, nil, time.Now())
			if err == nil && (route.AuthEnv == "" || os.Getenv(route.AuthEnv) != "") {
				token := ""
				if route.AuthEnv != "" {
					token = os.Getenv(route.AuthEnv)
				}
				models, fetchErr := fetchOpenAICompatibleModels(ctx, validated, route, token, maxCatalogBytes)
				if fetchErr == nil {
					reachable++
					for _, model := range models {
						if model == route.DisplayModelID {
							modelPresent++
							break
						}
					}
				} else {
					probeMessage = fetchErr.Error()
					modelMessage = fetchErr.Error()
				}
			} else if err != nil {
				probeMessage = err.Error()
				modelMessage = err.Error()
			}
			cancel()
		}
	}
	configStatus := "available"
	if len(catalog.Routes) == 0 {
		configStatus = "unavailable"
	}
	authStatus := "available"
	authMessage := ""
	if missingAuth > 0 {
		authStatus = "unavailable"
		authMessage = "one or more auth environment variables are not set"
	}
	discoveryStatus := "available"
	discoveryMessage := "current-model and configured-route discovery are non-mutating"
	if projectErr != nil && !os.IsNotExist(projectErr) {
		discoveryStatus = "partial"
		discoveryMessage = projectErr.Error()
	}
	reachability := doctorDimension{Status: "unavailable", Message: probeMessage}
	presence := doctorDimension{Status: "unavailable", Message: modelMessage}
	if probe {
		reachability = doctorDimension{Status: statusForCount(reachable), Count: reachable, Message: ""}
		presence = doctorDimension{Status: statusForCount(modelPresent), Count: modelPresent, Message: ""}
		if probed > 0 && reachable == 0 && probeMessage != "not probed" {
			reachability.Message = probeMessage
		}
		if probed > 0 && modelPresent == 0 && modelMessage != "not probed" {
			presence.Message = modelMessage
		}
	}
	return doctorOutput{
		Discovery:         doctorDimension{Status: discoveryStatus, Message: discoveryMessage},
		Configured:        doctorDimension{Status: configStatus, Count: len(catalog.Routes)},
		ProjectSelectable: doctorDimension{Status: statusForCount(projectSelectable), Count: projectSelectable},
		Selectable:        doctorDimension{Status: statusForCount(selectable), Count: selectable},
		Auth:              doctorDimension{Status: authStatus, Count: len(catalog.Routes) - missingAuth, Message: authMessage},
		Reachability:      reachability,
		ModelPresence:     presence,
		Dispatch:          doctorDimension{Status: "unavailable", Message: "dispatch is not implemented in this slice"},
		DispatchProven:    doctorDimension{Status: "unavailable", Message: "listing/configuration alone does not prove dispatch"},
		Control:           doctorDimension{Status: "unavailable", Message: "control is not implemented in this slice"},
	}, nil
}

func currentModel(model string) modelrouting.CurrentModel {
	if model == "" {
		return modelrouting.CurrentModel{}
	}
	route := redactRoute(currentRoute(model))
	return modelrouting.CurrentModel{ModelID: model, Surface: "current", Route: &route}
}

func currentRoute(model string) modelrouting.Route {
	return modelrouting.Route{
		Alias:           "current",
		DisplayModelID:  model,
		Adapter:         "codex",
		AdapterRevision: "v1",
		DispatchMethod:  "exec-model",
		Destination:     "current",
		Boundary:        modelrouting.BoundaryHosted,
		Retention:       modelrouting.RetentionSession,
		TrainingUse:     modelrouting.TrainingUnknown,
		Residency:       "unknown",
		TrustProvenance: "active orchestrator",
		Readiness:       []modelrouting.Readiness{modelrouting.ReadinessDiscovered, modelrouting.ReadinessConfigured, modelrouting.ReadinessSelectable},
		Capability: modelrouting.CapabilityEvidence{
			Class:          modelrouting.ClassLarge,
			Source:         modelrouting.EvidenceDeclared,
			RouteAlias:     "current",
			ModelID:        model,
			TaskFamily:     "code",
			Risk:           modelrouting.RiskBroad,
			DispatchProven: false,
		},
	}
}

func redactedDiscoveredRoute(alias, model, adapter, method, destination string) modelrouting.Route {
	return modelrouting.Route{
		Alias:           alias,
		DisplayModelID:  model,
		Adapter:         adapter,
		AdapterRevision: "v1",
		DispatchMethod:  method,
		Destination:     destination,
		Boundary:        modelrouting.BoundaryHosted,
		Retention:       modelrouting.RetentionSession,
		TrainingUse:     modelrouting.TrainingUnknown,
		Residency:       "unknown",
		TrustProvenance: "adapter discovery",
		Readiness:       []modelrouting.Readiness{modelrouting.ReadinessDiscovered},
		Capability: modelrouting.CapabilityEvidence{
			Class:          modelrouting.ClassSmall,
			Source:         modelrouting.EvidenceDeclared,
			RouteAlias:     alias,
			ModelID:        model,
			TaskFamily:     "code",
			Risk:           modelrouting.RiskBroad,
			DispatchProven: false,
		},
	}
}

func redactedChildRoute(parent modelrouting.Route, model string) modelrouting.Route {
	route := redactRoute(parent)
	if route.SourceRouteID == "" {
		route.SourceRouteID = route.RouteID
	}
	route.RouteID = ""
	route.Alias = parent.Alias + "." + safeAlias(model)
	route.DisplayModelID = model
	route.Readiness = []modelrouting.Readiness{modelrouting.ReadinessDiscovered}
	route.Capability.RouteAlias = route.Alias
	route.Capability.ModelID = model
	route.Capability.Source = modelrouting.EvidenceDeclared
	route.Capability.DispatchProven = false
	route.Capability.ExpiresAt = time.Time{}
	return route
}

func redactRoute(route modelrouting.Route) modelrouting.Route {
	route.Endpoint = ""
	route.AuthEnv = ""
	route.Capability.DispatchProven = false
	return route
}

func redactCatalog(catalog modelrouting.Catalog) modelrouting.Catalog {
	catalog.Routes = redactedUserRoutes(catalog.Routes)
	if catalog.Current.Route != nil {
		route := redactRoute(*catalog.Current.Route)
		catalog.Current.Route = &route
	}
	return catalog
}

func redactedUserRoutes(routes []modelrouting.Route) []modelrouting.Route {
	out := make([]modelrouting.Route, 0, len(routes))
	for _, route := range routes {
		route.SourceRouteID = route.RouteID
		route.RouteID = ""
		out = append(out, redactRoute(route))
	}
	return out
}

func loadTrustFile(root string) (userTrustFileData, error) {
	var trust userTrustFileData
	err := modelrouting.LoadStrictJSON(root, userTrustFile, &trust, maxCatalogBytes)
	if os.IsNotExist(err) {
		return userTrustFileData{SchemaVersion: 1}, nil
	}
	if err != nil {
		return userTrustFileData{}, err
	}
	if trust.SchemaVersion != 1 {
		return userTrustFileData{}, fmt.Errorf("unsupported trust schema version %d", trust.SchemaVersion)
	}
	seenProjects := map[string]struct{}{}
	for _, project := range trust.Projects {
		if strings.TrimSpace(project.ProjectID) == "" {
			return userTrustFileData{}, fmt.Errorf("trust project_id is required")
		}
		if _, exists := seenProjects[project.ProjectID]; exists {
			return userTrustFileData{}, fmt.Errorf("duplicate trust project_id %q", project.ProjectID)
		}
		seenProjects[project.ProjectID] = struct{}{}
		for _, approval := range project.RouteApprovals {
			if approval.ProjectID != project.ProjectID || approval.RouteFingerprint == "" || approval.ExpiresAt.IsZero() {
				return userTrustFileData{}, fmt.Errorf("invalid route approval")
			}
		}
		for _, denial := range project.RouteDenials {
			if denial.ProjectID != project.ProjectID || denial.RouteFingerprint == "" || denial.CreatedAt.IsZero() {
				return userTrustFileData{}, fmt.Errorf("invalid route denial")
			}
		}
		for _, approval := range project.EndpointApprovals {
			if approval.ProjectID != project.ProjectID || approval.Origin == "" || approval.ExpiresAt.IsZero() {
				return userTrustFileData{}, fmt.Errorf("invalid endpoint approval")
			}
		}
		for _, binding := range project.AuthBindings {
			if !authEnvPattern.MatchString(binding.Env) || binding.Adapter == "" || binding.Origin == "" || binding.ExpiresAt.IsZero() {
				return userTrustFileData{}, fmt.Errorf("invalid auth binding")
			}
		}
	}
	return trust, nil
}

func saveTrustFile(root string, trust userTrustFileData) error {
	trust.SchemaVersion = 1
	sort.SliceStable(trust.Projects, func(i, j int) bool { return trust.Projects[i].ProjectID < trust.Projects[j].ProjectID })
	return modelrouting.SaveAtomicJSON(root, userTrustFile, trust, maxCatalogBytes)
}

func policyContextForProject(userRoot, projectRoot string) (modelrouting.PolicyContext, error) {
	projectID, err := modelrouting.CanonicalProjectIdentity(projectRoot)
	if err != nil {
		return modelrouting.PolicyContext{}, err
	}
	projectPolicy, err := loadProjectPolicy(projectRoot)
	if err != nil && !os.IsNotExist(err) {
		return modelrouting.PolicyContext{}, err
	}
	trustFile, err := loadTrustFile(userRoot)
	if err != nil {
		return modelrouting.PolicyContext{}, err
	}
	userCatalog, err := loadUserCatalog(userRoot)
	if err != nil {
		return modelrouting.PolicyContext{}, err
	}
	routeSources := make(map[string]modelrouting.Route, len(userCatalog.Routes))
	for _, route := range userCatalog.Routes {
		routeSources[route.RouteID] = route
	}
	return modelrouting.PolicyContext{
		Project: modelrouting.ProjectPolicy{
			ProjectID:      projectID,
			AllowedAliases: projectPolicy.AllowedAliases,
			MaxRetention:   modelrouting.RetentionLimited,
		},
		Trusted:      trustForProject(trustFile, projectID),
		RouteSources: routeSources,
	}, nil
}

func trustForProject(trust userTrustFileData, projectID string) modelrouting.UserTrust {
	out := modelrouting.UserTrust{ProjectID: projectID}
	for _, project := range trust.Projects {
		if project.ProjectID != projectID {
			continue
		}
		out.RouteApprovals = append([]modelrouting.RouteApproval(nil), project.RouteApprovals...)
		out.RouteDenials = append([]modelrouting.RouteDenial(nil), project.RouteDenials...)
		out.EndpointApprovals = append([]modelrouting.EndpointApproval(nil), project.EndpointApprovals...)
		out.AuthBindings = append([]modelrouting.AuthBinding(nil), project.AuthBindings...)
	}
	return out
}

func routeApprovalInputs(userRoot, projectRoot, alias string) (modelrouting.Route, string, string, error) {
	catalog, err := loadUserCatalog(userRoot)
	if err != nil {
		return modelrouting.Route{}, "", "", fmt.Errorf("load user catalog: %w", err)
	}
	var route modelrouting.Route
	found := false
	for _, candidate := range catalog.Routes {
		if candidate.Alias == alias {
			route = candidate
			found = true
			break
		}
	}
	if !found {
		return modelrouting.Route{}, "", "", fmt.Errorf("route alias %q not found", alias)
	}
	projectID, err := modelrouting.CanonicalProjectIdentity(projectRoot)
	if err != nil {
		return modelrouting.Route{}, "", "", err
	}
	fingerprint, err := modelrouting.ApprovalRouteFingerprint(route, nil)
	if err != nil {
		return modelrouting.Route{}, "", "", err
	}
	return route, projectID, fingerprint, nil
}

func approveRouteTrust(trust userTrustFileData, projectID string, route modelrouting.Route, fingerprint string, expires time.Time) userTrustFileData {
	project := findProjectTrust(trust, projectID)
	project.RouteDenials = removeRouteDenials(project.RouteDenials, projectID, fingerprint)
	project.RouteApprovals = upsertRouteApproval(project.RouteApprovals, modelrouting.RouteApproval{ProjectID: projectID, RouteFingerprint: fingerprint, ExpiresAt: expires})
	if origin := originForEndpoint(route.Endpoint); origin != "" {
		if route.Boundary == modelrouting.BoundaryPrivate {
			project.EndpointApprovals = upsertEndpointApproval(project.EndpointApprovals, modelrouting.EndpointApproval{ProjectID: projectID, Origin: origin, ExpiresAt: expires})
		}
		if route.AuthEnv != "" {
			project.AuthBindings = upsertAuthBinding(project.AuthBindings, modelrouting.AuthBinding{Env: route.AuthEnv, Adapter: route.Adapter, Origin: origin, ExpiresAt: expires})
		}
	}
	return storeProjectTrust(trust, project)
}

func removeRouteTrust(trust userTrustFileData, projectID string, route modelrouting.Route, fingerprint string, deny bool) userTrustFileData {
	project := findProjectTrust(trust, projectID)
	project.RouteApprovals = removeRouteApprovals(project.RouteApprovals, projectID, fingerprint)
	origin := originForEndpoint(route.Endpoint)
	project.EndpointApprovals = removeEndpointApprovals(project.EndpointApprovals, projectID, origin)
	project.AuthBindings = removeAuthBindings(project.AuthBindings, route.AuthEnv, route.Adapter, origin)
	if deny {
		project.RouteDenials = append(removeRouteDenials(project.RouteDenials, projectID, fingerprint), modelrouting.RouteDenial{ProjectID: projectID, RouteFingerprint: fingerprint, CreatedAt: time.Now().UTC()})
	} else {
		project.RouteDenials = removeRouteDenials(project.RouteDenials, projectID, fingerprint)
	}
	return storeProjectTrust(trust, project)
}

func findProjectTrust(trust userTrustFileData, projectID string) userProjectTrust {
	for _, project := range trust.Projects {
		if project.ProjectID == projectID {
			return project
		}
	}
	return userProjectTrust{ProjectID: projectID}
}

func storeProjectTrust(trust userTrustFileData, project userProjectTrust) userTrustFileData {
	out := userTrustFileData{SchemaVersion: 1}
	for _, existing := range trust.Projects {
		if existing.ProjectID == project.ProjectID {
			continue
		}
		out.Projects = append(out.Projects, existing)
	}
	out.Projects = append(out.Projects, project)
	return out
}

func upsertRouteApproval(values []modelrouting.RouteApproval, approval modelrouting.RouteApproval) []modelrouting.RouteApproval {
	values = removeRouteApprovals(values, approval.ProjectID, approval.RouteFingerprint)
	return append(values, approval)
}

func removeRouteApprovals(values []modelrouting.RouteApproval, projectID, fingerprint string) []modelrouting.RouteApproval {
	out := []modelrouting.RouteApproval{}
	for _, value := range values {
		if value.ProjectID == projectID && value.RouteFingerprint == fingerprint {
			continue
		}
		out = append(out, value)
	}
	return out
}

func removeRouteDenials(values []modelrouting.RouteDenial, projectID, fingerprint string) []modelrouting.RouteDenial {
	out := []modelrouting.RouteDenial{}
	for _, value := range values {
		if value.ProjectID == projectID && value.RouteFingerprint == fingerprint {
			continue
		}
		out = append(out, value)
	}
	return out
}

func upsertEndpointApproval(values []modelrouting.EndpointApproval, approval modelrouting.EndpointApproval) []modelrouting.EndpointApproval {
	out := []modelrouting.EndpointApproval{}
	for _, value := range values {
		if value.ProjectID == approval.ProjectID && strings.EqualFold(value.Origin, approval.Origin) {
			continue
		}
		out = append(out, value)
	}
	return append(out, approval)
}

func removeEndpointApprovals(values []modelrouting.EndpointApproval, projectID, origin string) []modelrouting.EndpointApproval {
	out := []modelrouting.EndpointApproval{}
	for _, value := range values {
		if value.ProjectID == projectID && origin != "" && strings.EqualFold(value.Origin, origin) {
			continue
		}
		out = append(out, value)
	}
	return out
}

func upsertAuthBinding(values []modelrouting.AuthBinding, binding modelrouting.AuthBinding) []modelrouting.AuthBinding {
	out := []modelrouting.AuthBinding{}
	for _, value := range values {
		if value.Env == binding.Env && value.Adapter == binding.Adapter && strings.EqualFold(value.Origin, binding.Origin) {
			continue
		}
		out = append(out, value)
	}
	return append(out, binding)
}

func removeAuthBindings(values []modelrouting.AuthBinding, env, adapter, origin string) []modelrouting.AuthBinding {
	out := []modelrouting.AuthBinding{}
	for _, value := range values {
		if env != "" && value.Env == env && value.Adapter == adapter && strings.EqualFold(value.Origin, origin) {
			continue
		}
		out = append(out, value)
	}
	return out
}

func routeProjectSelectable(route modelrouting.Route, policy modelrouting.PolicyContext) bool {
	if !hasReadiness(route.Readiness, modelrouting.ReadinessSelectable) {
		return false
	}
	if len(policy.Project.AllowedAliases) > 0 && !containsString(policy.Project.AllowedAliases, route.Alias) {
		return false
	}
	fingerprint, err := modelrouting.ApprovalRouteFingerprint(route, policy.RouteSources)
	if err != nil {
		return false
	}
	now := time.Now()
	for _, denial := range policy.Trusted.RouteDenials {
		if denial.ProjectID == policy.Project.ProjectID && denial.RouteFingerprint == fingerprint {
			return false
		}
	}
	for _, approval := range policy.Trusted.RouteApprovals {
		if approval.ProjectID == policy.Project.ProjectID && approval.RouteFingerprint == fingerprint && now.Before(approval.ExpiresAt) {
			return true
		}
	}
	return route.Boundary != modelrouting.BoundaryPrivate
}

func routeAuthApproved(route modelrouting.Route, policy modelrouting.PolicyContext, now time.Time) bool {
	if route.AuthEnv == "" {
		return true
	}
	origin := originForEndpoint(route.Endpoint)
	for _, binding := range policy.Trusted.AuthBindings {
		if binding.Env == route.AuthEnv && binding.Adapter == route.Adapter && strings.EqualFold(binding.Origin, origin) && now.Before(binding.ExpiresAt) {
			return true
		}
	}
	return false
}

func localSurfaceFingerprints(catalog modelrouting.Catalog, policy modelrouting.PolicyContext) []modelrouting.SurfaceFingerprint {
	routeFingerprints := make([]string, 0, len(catalog.Routes))
	for _, route := range catalog.Routes {
		if fingerprint, err := modelrouting.ComputeRouteFingerprint(route); err == nil {
			routeFingerprints = append(routeFingerprints, fingerprint)
		}
	}
	sort.Strings(routeFingerprints)
	catalogHash := sha256Text(strings.Join(routeFingerprints, "\x00"))
	policyHash := sha256Text(policy.Project.ProjectID + "\x00" + strings.Join(policy.Project.AllowedAliases, "\x00"))
	return []modelrouting.SurfaceFingerprint{{
		Surface:    "kbrouter",
		Provider:   "user-local-config",
		Revision:   "v1",
		ConfigHash: "sha256:" + catalogHash + "." + policyHash,
	}}
}

func sha256Text(value string) string {
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:])
}

func originForEndpoint(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return strings.ToLower(parsed.Scheme + "://" + parsed.Host)
}

func fingerprintAndSort(catalog *modelrouting.Catalog) {
	sort.SliceStable(catalog.Routes, func(i, j int) bool { return catalog.Routes[i].Alias < catalog.Routes[j].Alias })
	sort.SliceStable(catalog.Surfaces, func(i, j int) bool {
		left := catalog.Surfaces[i]
		right := catalog.Surfaces[j]
		return left.Surface+"\x00"+left.Provider+"\x00"+left.Revision < right.Surface+"\x00"+right.Provider+"\x00"+right.Revision
	})
	if fingerprint, err := modelrouting.ComputeCatalogFingerprint(*catalog); err == nil {
		catalog.Fingerprint = fingerprint
	}
}

func upsertRoute(routes []modelrouting.Route, route modelrouting.Route) []modelrouting.Route {
	out := make([]modelrouting.Route, 0, len(routes)+1)
	replaced := false
	for _, existing := range routes {
		if existing.Alias == route.Alias {
			out = append(out, route)
			replaced = true
			continue
		}
		out = append(out, existing)
	}
	if !replaced {
		out = append(out, route)
	}
	return out
}

func removeRoute(routes []modelrouting.Route, alias string) []modelrouting.Route {
	out := []modelrouting.Route{}
	for _, route := range routes {
		if route.Alias != alias {
			out = append(out, route)
		}
	}
	return out
}

func hasReadiness(values []modelrouting.Readiness, target modelrouting.Readiness) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func statusForCount(count int) string {
	if count > 0 {
		return "available"
	}
	return "unavailable"
}

func atomicWriteJSON(root, name string, value any) error {
	return modelrouting.SaveAtomicJSON(root, name, value, maxCatalogBytes)
}

func addUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func removeString(values []string, value string) []string {
	out := []string{}
	for _, existing := range values {
		if existing != value {
			out = append(out, existing)
		}
	}
	return out
}

func containsString(values []string, value string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}

func sortedUnique(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func safeAlias(value string) string {
	value = strings.ToLower(value)
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('-')
	}
	out := strings.Trim(b.String(), "-.")
	if out == "" {
		return "model"
	}
	if len(out) > 40 {
		return out[:40]
	}
	return out
}
