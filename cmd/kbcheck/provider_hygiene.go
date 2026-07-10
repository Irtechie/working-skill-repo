package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type providerHygieneResult struct {
	OK              bool     `json:"ok"`
	Checked         []string `json:"checked"`
	OptionalConfigs []string `json:"optional_configs"`
	Issues          []string `json:"issues"`
}

func runProviderHygieneCommand(root string, opts options, stdout, stderr io.Writer) int {
	home, _ := os.UserHomeDir()
	result := computeProviderHygiene(root, home, opts.includeUser)
	if opts.json {
		writeJSON(stdout, result)
	} else if result.OK {
		fmt.Fprintf(stdout, "provider hygiene: ok checked=%d optional=%d\n", len(result.Checked), len(result.OptionalConfigs))
		for _, path := range result.OptionalConfigs {
			fmt.Fprintf(stdout, "optional: %s\n", path)
		}
	} else {
		for _, issue := range result.Issues {
			fmt.Fprintln(stderr, issue)
		}
	}
	if !result.OK {
		return 2
	}
	return 0
}

func computeProviderHygiene(root, home string, includeUser bool) providerHygieneResult {
	ccePattern := regexp.MustCompile(`(?i)(^|[^a-z0-9_-])cce(?:\.exe)?([^a-z0-9_-]|$)|context-engine`)
	paths := []string{
		filepath.Join(root, ".mcp.json"),
		filepath.Join(root, ".claude", "settings.json"),
		filepath.Join(root, ".claude", "settings.local.json"),
	}
	if includeUser {
		paths = append(paths,
			filepath.Join(home, ".copilot", "mcp-config.json"),
			filepath.Join(home, ".codex", "config.toml"),
			filepath.Join(home, ".claude.json"),
			filepath.Join(home, ".claude", "settings.json"),
		)
	}
	result := providerHygieneResult{}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				result.Issues = append(result.Issues, "Cannot read provider config: "+path+": "+err.Error())
			}
			continue
		}
		result.Checked = append(result.Checked, path)
		activeText, err := activeProviderText(path, content)
		if err != nil {
			result.Issues = append(result.Issues, "Cannot parse provider config: "+path+": "+err.Error())
			continue
		}
		lower := strings.ToLower(activeText)
		if providerToken(lower, "phoenix") {
			result.Issues = append(result.Issues, "Phoenix provider activation found: "+path)
			continue
		}
		if ccePattern.MatchString(lower) {
			result.OptionalConfigs = append(result.OptionalConfigs, path)
		}
	}
	result.OK = len(result.Issues) == 0
	return result
}

func activeProviderText(path string, content []byte) (string, error) {
	if strings.EqualFold(filepath.Ext(path), ".toml") {
		lines := []string{}
		inProviderSection := false
		disabled := false
		section := []string{}
		flush := func() {
			if inProviderSection && !disabled {
				lines = append(lines, section...)
			}
			section = nil
			disabled = false
		}
		for _, raw := range strings.Split(string(content), "\n") {
			line := strings.TrimSpace(strings.SplitN(raw, "#", 2)[0])
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "[") {
				flush()
				lower := strings.ToLower(line)
				inProviderSection = strings.Contains(lower, "mcp") || strings.Contains(lower, "provider")
			}
			if inProviderSection {
				section = append(section, line)
				if regexp.MustCompile(`(?i)^enabled\s*=\s*false$`).MatchString(line) {
					disabled = true
				}
			}
		}
		flush()
		return strings.Join(lines, "\n"), nil
	}
	var value any
	if err := json.Unmarshal(content, &value); err != nil {
		return "", err
	}
	active := []string{}
	collectActiveProviderText(value, false, &active)
	return strings.Join(active, "\n"), nil
}

func collectActiveProviderText(value any, providerContext bool, active *[]string) {
	switch node := value.(type) {
	case map[string]any:
		if disabled, ok := node["enabled"].(bool); ok && !disabled {
			return
		}
		for key, child := range node {
			lower := strings.ToLower(key)
			nextContext := providerContext || lower == "mcpservers" || lower == "mcp_servers" || lower == "hooks"
			if nextContext && !providerNodeDisabled(child) {
				*active = append(*active, key)
			}
			collectActiveProviderText(child, nextContext, active)
		}
	case []any:
		for _, child := range node {
			collectActiveProviderText(child, providerContext, active)
		}
	case string:
		if providerContext {
			*active = append(*active, node)
		}
	}

}

func providerNodeDisabled(value any) bool {
	node, ok := value.(map[string]any)
	if !ok {
		return false
	}
	enabled, ok := node["enabled"].(bool)
	return ok && !enabled
}

func providerToken(text, token string) bool {
	pattern := regexp.MustCompile(`(?i)(^|[^a-z0-9_-])` + regexp.QuoteMeta(token) + `(?:[-_][a-z0-9_-]+)?([^a-z0-9_-]|$)`)
	return pattern.MatchString(text)
}

func runProviderHygieneSelftest(stdout, stderr io.Writer) int {
	temp, err := os.MkdirTemp("", "provider-hygiene-selftest-*")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer os.RemoveAll(temp)
	root := filepath.Join(temp, "repo")
	home := filepath.Join(temp, "home")
	_ = os.MkdirAll(filepath.Join(root, ".claude"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".claude", "settings.local.json"), []byte(`{"hooks":{"SessionStart":[{"command":"cce status"}]}}`), 0o644)
	result := computeProviderHygiene(root, home, true)
	if !result.OK || len(result.OptionalConfigs) != 1 {
		fmt.Fprintf(stderr, "optional CCE config rejected: %#v\n", result)
		return 1
	}
	_ = os.MkdirAll(filepath.Join(home, ".copilot"), 0o755)
	_ = os.WriteFile(filepath.Join(home, ".copilot", "mcp-config.json"), []byte(`{"mcpServers":{"phoenix":{"command":"phoenix-mcp"}}}`), 0o644)
	result = computeProviderHygiene(root, home, true)
	if result.OK || len(result.Issues) != 1 {
		fmt.Fprintf(stderr, "Phoenix config accepted: %#v\n", result)
		return 1
	}
	fmt.Fprintln(stdout, "provider hygiene selftest passed")
	return 0
}
