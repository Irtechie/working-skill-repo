package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestModelsImportCanonicalizesRouteWithoutGrantingTrustOrPersistingSecrets(t *testing.T) {
	root := t.TempDir()
	userRoot := filepath.Join(root, "user")
	projectRoot := filepath.Join(root, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(userRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := saveTrustFile(userRoot, userTrustFileData{SchemaVersion: 1}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LOCAL_MODEL_API_KEY", "dummy-secret-value")
	trustBefore := readFileForTest(t, filepath.Join(userRoot, userTrustFile))
	importPath := filepath.Join(root, "kbrouter-routes.local.json")
	writeJSONForTest(t, importPath, map[string]any{
		"schema_version": 1,
		"routes": []map[string]any{{
			"alias": "local.coder", "model_id": "operator-model",
			"endpoint": "HTTP://127.0.0.1:12345/v1/",
			"auth_env": "LOCAL_MODEL_API_KEY", "boundary": "private", "hosting": "self-hosted",
			"retention": "none", "training_use": "no", "residency": "local",
			"trust_provenance": "operator-managed endpoint",
			"capability": map[string]any{
				"class": "small", "task_family": "code", "tools": []string{"text"},
				"context_size": 4096, "risk": "normal",
			},
		}},
	})

	code, stdout, stderr := runForTest(
		"models", "import", "--user-root", userRoot, "--project-root", projectRoot,
		"--file", importPath, "--json",
	)
	if code != 0 {
		t.Fatalf("import exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	catalog := loadUserCatalogForTest(t, userRoot)
	if len(catalog.Routes) != 1 {
		t.Fatalf("routes=%#v", catalog.Routes)
	}
	route := catalog.Routes[0]
	if route.Alias != "local.coder" || route.DisplayModelID != "operator-model" ||
		route.Endpoint != "http://127.0.0.1:12345/v1" || route.AuthEnv != "LOCAL_MODEL_API_KEY" ||
		route.Adapter != "openai-compatible" || route.DispatchMethod != "chat-completions" {
		t.Fatalf("route not canonicalized: %#v", route)
	}
	if route.Capability.Source != modelrouting.EvidenceDeclared || route.Capability.DispatchQualified || route.Capability.DispatchProven {
		t.Fatalf("import granted capability credit: %#v", route.Capability)
	}
	if trustAfter := readFileForTest(t, filepath.Join(userRoot, userTrustFile)); trustAfter != trustBefore {
		t.Fatalf("import mutated trust state: before=%s after=%s", trustBefore, trustAfter)
	}
	modelsBytes := readFileForTest(t, filepath.Join(userRoot, userCatalogFile))
	if strings.Contains(modelsBytes, "dummy-secret-value") || !strings.Contains(modelsBytes, "LOCAL_MODEL_API_KEY") {
		t.Fatalf("catalog persisted a secret or lost the env name: %s", modelsBytes)
	}
	routeID := route.RouteID
	code, _, stderr = runForTest(
		"models", "import", "--user-root", userRoot, "--project-root", projectRoot,
		"--file", importPath, "--json",
	)
	if code != 0 {
		t.Fatalf("repeat import exit=%d stderr=%s", code, stderr)
	}
	reimported := loadUserCatalogForTest(t, userRoot).Routes[0]
	if reimported.RouteID != routeID {
		t.Fatalf("identical import rotated route identity: before=%s after=%s", routeID, reimported.RouteID)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, projectPolicyFile)); !os.IsNotExist(err) {
		t.Fatalf("import wrote project runtime state: %v", err)
	}
}

func TestModelsImportRejectsSecretValuesUnknownFieldsAndSymlinks(t *testing.T) {
	root := t.TempDir()
	userRoot := filepath.Join(root, "user")
	projectRoot := filepath.Join(root, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	importPath := filepath.Join(root, "routes.json")
	route := map[string]any{
		"alias": "local.bad", "model_id": "model", "endpoint": "http://127.0.0.1:1234/v1",
		"auth_env": "LOCAL_API_KEY", "auth_value": "dummy-secret-value",
		"boundary": "private", "hosting": "self-hosted", "retention": "none", "training_use": "no",
		"residency": "local", "trust_provenance": "operator",
		"capability": map[string]any{
			"class": "small", "task_family": "code", "tools": []string{"text"},
			"context_size": 4096, "risk": "normal",
		},
	}
	writeJSONForTest(t, importPath, map[string]any{"schema_version": 1, "routes": []map[string]any{route}})
	code, stdout, stderr := runForTest("models", "import", "--user-root", userRoot, "--project-root", projectRoot, "--file", importPath, "--json")
	if code == 0 || !strings.Contains(stdout+stderr, "unknown field") {
		t.Fatalf("secret-shaped unknown field accepted: code=%d output=%s%s", code, stdout, stderr)
	}
	if _, err := os.Stat(filepath.Join(userRoot, userCatalogFile)); !os.IsNotExist(err) {
		t.Fatalf("failed import wrote canonical state: %v", err)
	}

	delete(route, "auth_value")
	route["auth_env"] = "actual-secret-value"
	writeJSONForTest(t, importPath, map[string]any{"schema_version": 1, "routes": []map[string]any{route}})
	code, stdout, stderr = runForTest("models", "import", "--user-root", userRoot, "--project-root", projectRoot, "--file", importPath, "--json")
	if code == 0 || !strings.Contains(stdout+stderr, "environment variable name") {
		t.Fatalf("auth value accepted as env name: code=%d output=%s%s", code, stdout, stderr)
	}

	route["auth_env"] = ""
	route["endpoint"] = "http://operator:actual-secret-value@127.0.0.1:1234/v1"
	writeJSONForTest(t, importPath, map[string]any{"schema_version": 1, "routes": []map[string]any{route}})
	code, stdout, stderr = runForTest("models", "import", "--user-root", userRoot, "--project-root", projectRoot, "--file", importPath, "--json")
	if code == 0 || !strings.Contains(strings.ToLower(stdout+stderr), "unsafe") {
		t.Fatalf("credential-bearing endpoint accepted: code=%d output=%s%s", code, stdout, stderr)
	}

	validPath := filepath.Join(root, "valid.json")
	writeJSONForTest(t, validPath, map[string]any{"schema_version": 1, "routes": []any{}})
	linkPath := filepath.Join(root, "routes-link.json")
	if err := os.Symlink(validPath, linkPath); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	code, stdout, stderr = runForTest("models", "import", "--user-root", userRoot, "--project-root", projectRoot, "--file", linkPath, "--json")
	if code == 0 || !strings.Contains(strings.ToLower(stdout+stderr), "unsafe") {
		t.Fatalf("symlink import accepted: code=%d output=%s%s", code, stdout, stderr)
	}
}

func TestModelsImportRejectsPlaceholdersUnsupportedCapabilityAndDuplicateTools(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "placeholder alias", mutate: func(route map[string]any) { route["alias"] = "<fill-route-alias>" }},
		{name: "placeholder model", mutate: func(route map[string]any) { route["model_id"] = "<fill-model-id>" }},
		{name: "placeholder endpoint", mutate: func(route map[string]any) { route["endpoint"] = "<fill-endpoint>" }},
		{name: "broad risk", mutate: func(route map[string]any) { route["capability"].(map[string]any)["risk"] = "broad" }},
		{name: "unsupported class", mutate: func(route map[string]any) { route["capability"].(map[string]any)["class"] = "planner" }},
		{name: "duplicate tool", mutate: func(route map[string]any) { route["capability"].(map[string]any)["tools"] = []string{"text", "text"} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			userRoot := filepath.Join(root, "user")
			projectRoot := filepath.Join(root, "project")
			if err := os.MkdirAll(projectRoot, 0o755); err != nil {
				t.Fatal(err)
			}
			route := validImportRoute()
			tc.mutate(route)
			importPath := filepath.Join(root, "routes.json")
			writeJSONForTest(t, importPath, map[string]any{"schema_version": 1, "routes": []map[string]any{route}})
			code, stdout, stderr := runForTest(
				"models", "import", "--user-root", userRoot, "--project-root", projectRoot,
				"--file", importPath, "--json",
			)
			if code == 0 {
				t.Fatalf("invalid route accepted: %s%s", stdout, stderr)
			}
			if _, err := os.Stat(filepath.Join(userRoot, userCatalogFile)); !os.IsNotExist(err) {
				t.Fatalf("invalid import wrote canonical state: %v", err)
			}
		})
	}
}

func TestModelsImportIsAtomicAcrossMultipleRoutes(t *testing.T) {
	root := t.TempDir()
	userRoot := filepath.Join(root, "user")
	projectRoot := filepath.Join(root, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	importPath := filepath.Join(root, "routes.json")
	writeJSONForTest(t, importPath, map[string]any{"schema_version": 1, "routes": []map[string]any{validImportRoute()}})
	code, stdout, stderr := runForTest(
		"models", "import", "--user-root", userRoot, "--project-root", projectRoot,
		"--file", importPath, "--json",
	)
	if code != 0 {
		t.Fatalf("seed import exit=%d output=%s%s", code, stdout, stderr)
	}
	before := readFileForTest(t, filepath.Join(userRoot, userCatalogFile))
	first := validImportRoute()
	first["model_id"] = "replacement-model"
	second := validImportRoute()
	second["alias"] = "local.second"
	second["endpoint"] = "<fill-endpoint>"
	writeJSONForTest(t, importPath, map[string]any{"schema_version": 1, "routes": []map[string]any{first, second}})
	code, stdout, stderr = runForTest(
		"models", "import", "--user-root", userRoot, "--project-root", projectRoot,
		"--file", importPath, "--json",
	)
	if code == 0 {
		t.Fatalf("partially invalid import succeeded: %s%s", stdout, stderr)
	}
	if after := readFileForTest(t, filepath.Join(userRoot, userCatalogFile)); after != before {
		t.Fatalf("partially invalid import mutated catalog:\nbefore=%s\nafter=%s", before, after)
	}
}

func TestCheckedInRouteExampleIsARejectedPlaceholder(t *testing.T) {
	root := t.TempDir()
	userRoot := filepath.Join(root, "user")
	projectRoot := filepath.Join(root, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	examplePath := filepath.Join("..", "..", "config", "kbrouter-routes.example.json")
	code, stdout, stderr := runForTest(
		"models", "import", "--user-root", userRoot, "--project-root", projectRoot,
		"--file", examplePath, "--json",
	)
	if code == 0 || !strings.Contains(stdout+stderr, "placeholder") {
		t.Fatalf("checked-in placeholder import exit=%d output=%s%s", code, stdout, stderr)
	}
	if _, err := os.Stat(filepath.Join(userRoot, userCatalogFile)); !os.IsNotExist(err) {
		t.Fatalf("placeholder example wrote canonical state: %v", err)
	}
}

func validImportRoute() map[string]any {
	return map[string]any{
		"alias": "local.coder", "model_id": "operator-model", "endpoint": "http://127.0.0.1:1234/v1",
		"auth_env": "", "boundary": "private", "hosting": "self-hosted", "retention": "none",
		"training_use": "no", "residency": "local", "trust_provenance": "operator",
		"capability": map[string]any{
			"class": "small", "task_family": "code", "tools": []string{"text"},
			"context_size": 4096, "risk": "normal",
		},
	}
}

func writeJSONForTest(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
