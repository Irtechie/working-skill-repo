package main

import (
	"fmt"
	"sort"
	"strings"
)

type containmentReport struct {
	Ready   bool     `json:"ready"`
	Runtime string   `json:"runtime,omitempty"`
	Issues  []string `json:"issues,omitempty"`
}

func inspectContainment(resolve func() (string, error)) containmentReport {
	runtime, err := resolve()
	if err != nil {
		return containmentReport{Issues: []string{err.Error()}}
	}
	return containmentReport{Ready: true, Runtime: runtime}
}

func (report containmentReport) RequireReady() error {
	if report.Ready {
		return nil
	}
	return fmt.Errorf("containment is not ready: %s", strings.Join(report.Issues, "; "))
}

func proofEnvironment(environment []string) []string {
	allowed := map[string]bool{"PATH": true, "SYSTEMROOT": true}
	return filteredEnvironment(environment, allowed)
}

func containerRuntimeEnvironment(environment []string) []string {
	allowed := map[string]bool{
		"PATH": true, "SYSTEMROOT": true, "HOME": true, "USERPROFILE": true,
		"APPDATA": true, "LOCALAPPDATA": true, "TEMP": true, "TMP": true,
	}
	return filteredEnvironment(environment, allowed)
}

func filteredEnvironment(environment []string, allowed map[string]bool) []string {
	values := map[string]string{}
	for _, entry := range environment {
		key, value, found := strings.Cut(entry, "=")
		if found && allowed[strings.ToUpper(key)] {
			values[strings.ToUpper(key)] = value
		}
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key+"="+values[key])
	}
	return result
}
