package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

const routeImportSchemaVersion = 1

type routeImportFile struct {
	SchemaVersion int               `json:"schema_version"`
	Routes        []routeImportItem `json:"routes"`
}

type routeImportItem struct {
	Alias           string                `json:"alias"`
	ModelID         string                `json:"model_id"`
	Endpoint        string                `json:"endpoint"`
	AuthEnv         string                `json:"auth_env,omitempty"`
	Boundary        string                `json:"boundary"`
	Hosting         string                `json:"hosting"`
	Retention       string                `json:"retention"`
	TrainingUse     string                `json:"training_use"`
	Residency       string                `json:"residency"`
	TrustProvenance string                `json:"trust_provenance"`
	Capability      routeImportCapability `json:"capability"`
}

type routeImportCapability struct {
	Class       string   `json:"class"`
	TaskFamily  string   `json:"task_family"`
	Tools       []string `json:"tools"`
	ContextSize int      `json:"context_size"`
	Risk        string   `json:"risk"`
}

func runModelsImport(args []string, stdout, stderr io.Writer) int {
	if hasHelpArg(args) {
		fmt.Fprint(stdout, modelsImportUsage)
		return 0
	}
	fs := flagSet("models import")
	opts := commonOptions{}
	opts.bind(fs)
	var sourcePath string
	fs.StringVar(&sourcePath, "file", "", "operator-owned route import JSON")
	if err := fs.Parse(args); err != nil {
		return commandError(stdout, stderr, hasExactArg(args, "--json"), 2, "invalid-arguments", err.Error())
	}
	if fs.NArg() != 0 {
		return commandError(stdout, stderr, opts.json, 2, "invalid-arguments", fmt.Sprintf("unexpected argument %q", fs.Arg(0)))
	}
	if customUserRootRejected(fs) {
		return commandError(stdout, stderr, opts.json, 2, "invalid-arguments", "route import writes the fixed user-local root; custom --user-root is test-only")
	}
	if strings.TrimSpace(sourcePath) == "" {
		return commandError(stdout, stderr, opts.json, 2, "invalid-arguments", "models import requires --file")
	}
	imported, err := loadRouteImport(sourcePath)
	if err != nil {
		return commandError(stdout, stderr, opts.json, 1, "import-read-error", fmt.Sprintf("load route import: %v", err))
	}
	routes, err := canonicalImportRoutes(imported)
	if err != nil {
		return commandError(stdout, stderr, opts.json, 1, "import-validation-error", fmt.Sprintf("validate route import: %v", err))
	}
	lock, err := modelrouting.AcquirePrivateStateLock(opts.userRoot, ".state.lock", 5*time.Second)
	if err != nil {
		return commandError(stdout, stderr, opts.json, 1, "import-lock-error", fmt.Sprintf("acquire route import lock: %v", err))
	}
	defer lock.Close()
	catalog, err := loadUserCatalog(opts.userRoot)
	if err != nil {
		return commandError(stdout, stderr, opts.json, 1, "import-state-error", fmt.Sprintf("load canonical routes: %v", err))
	}
	for _, route := range routes {
		if existing, found := findUserRoute(catalog.Routes, route.Alias); found {
			route.RouteID = existing.RouteID
		}
		catalog.Routes = upsertRoute(catalog.Routes, route)
	}
	if err := saveUserCatalog(opts.userRoot, catalog); err != nil {
		return commandError(stdout, stderr, opts.json, 1, "import-state-error", fmt.Sprintf("save imported routes: %v", err))
	}
	aliases := make([]string, 0, len(routes))
	for _, route := range routes {
		aliases = append(aliases, route.Alias)
	}
	sort.Strings(aliases)
	return printResult(stdout, stderr, map[string]any{
		"schema_version": 1, "status": "imported", "aliases": aliases,
		"path": filepath.Join(opts.userRoot, userCatalogFile), "trust_changed": false,
	}, opts.json, nil)
}

func loadRouteImport(path string) (routeImportFile, error) {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return routeImportFile{}, err
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return routeImportFile{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return routeImportFile{}, modelrouting.ErrUnsafePath
	}
	var imported routeImportFile
	if err := modelrouting.LoadStrictProjectJSON(filepath.Dir(absolute), filepath.Base(absolute), &imported, maxCatalogBytes); err != nil {
		return routeImportFile{}, err
	}
	return imported, nil
}

func canonicalImportRoutes(imported routeImportFile) ([]modelrouting.Route, error) {
	if imported.SchemaVersion != routeImportSchemaVersion || len(imported.Routes) == 0 || len(imported.Routes) > 32 {
		return nil, fmt.Errorf("route import requires schema_version 1 and 1-32 routes")
	}
	seen := make(map[string]struct{}, len(imported.Routes))
	routes := make([]modelrouting.Route, 0, len(imported.Routes))
	for _, item := range imported.Routes {
		semanticFields := append([]string{
			item.Alias, item.ModelID, item.Endpoint, item.AuthEnv, item.Boundary, item.Hosting,
			item.Retention, item.TrainingUse, item.Residency, item.TrustProvenance,
			item.Capability.Class, item.Capability.TaskFamily, item.Capability.Risk,
		}, item.Capability.Tools...)
		for _, value := range semanticFields {
			if strings.ContainsAny(value, "<>") {
				return nil, fmt.Errorf("all route placeholders must be filled before import")
			}
		}
		endpoint, err := canonicalImportEndpoint(item.Endpoint)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[item.Alias]; exists {
			return nil, fmt.Errorf("duplicate route alias %q", item.Alias)
		}
		seen[item.Alias] = struct{}{}
		if item.Boundary != string(modelrouting.BoundaryPrivate) && item.Boundary != string(modelrouting.BoundaryHosted) {
			return nil, fmt.Errorf("route %q boundary must be private or hosted", item.Alias)
		}
		if item.Hosting == "" || item.Retention == "" || item.TrainingUse == "" ||
			strings.TrimSpace(item.Residency) == "" || strings.TrimSpace(item.TrustProvenance) == "" ||
			item.Capability.TaskFamily == "" || len(item.Capability.Tools) == 0 || item.Capability.ContextSize <= 0 {
			return nil, fmt.Errorf("route %q needs complete hosting, trust, and capability metadata", item.Alias)
		}
		if item.Capability.Class != string(modelrouting.ClassSmall) &&
			item.Capability.Class != string(modelrouting.ClassMedium) &&
			item.Capability.Class != string(modelrouting.ClassLarge) {
			return nil, fmt.Errorf("route %q capability class must be small, medium, or large", item.Alias)
		}
		if item.Capability.Risk != string(modelrouting.RiskNormal) {
			return nil, fmt.Errorf("route %q capability risk must be normal", item.Alias)
		}
		toolSet := make(map[string]struct{}, len(item.Capability.Tools))
		for _, tool := range item.Capability.Tools {
			if _, duplicate := toolSet[tool]; duplicate {
				return nil, fmt.Errorf("route %q capability tools must be unique", item.Alias)
			}
			toolSet[tool] = struct{}{}
		}
		opts := addOptions{
			alias: item.Alias, model: item.ModelID, endpoint: endpoint, authEnv: item.AuthEnv,
			adapter: "openai-compatible", dispatchMethod: "chat-completions",
			boundary: item.Boundary, hosting: item.Hosting, retention: item.Retention, trainingUse: item.TrainingUse,
			residency: item.Residency, trustProvenance: item.TrustProvenance, class: item.Capability.Class,
		}
		route, err := routeFromAddOptions(opts)
		if err != nil {
			return nil, fmt.Errorf("route %q: %w", item.Alias, err)
		}
		route.Capability.TaskFamily = item.Capability.TaskFamily
		route.Capability.Tools = append([]string(nil), item.Capability.Tools...)
		route.Capability.ContextSize = item.Capability.ContextSize
		route.Capability.Risk = modelrouting.RiskLevel(item.Capability.Risk)
		route.Capability.Source = modelrouting.EvidenceDeclared
		route.Capability.DispatchQualified = false
		route.Capability.DispatchProven = false
		if err := modelrouting.ValidateCatalogStatic(modelrouting.Catalog{
			SchemaVersion: modelrouting.CatalogSchemaVersion, Routes: []modelrouting.Route{route},
		}, modelrouting.CatalogSourceUser); err != nil {
			return nil, fmt.Errorf("route %q: %w", item.Alias, err)
		}
		routes = append(routes, route)
	}
	return routes, nil
}

func canonicalImportEndpoint(value string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", modelrouting.ErrUnsafeEndpoint
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	if parsed.Path == "" {
		parsed.Path = "/v1"
	}
	canonical := parsed.String()
	if _, _, err := conservativeEndpointDefaults(canonical); err != nil {
		return "", err
	}
	return canonical, nil
}
