package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProviderHygieneAllowsOptionalCCE(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	path := filepath.Join(root, ".mcp.json")
	if err := os.WriteFile(path, []byte(`{"mcpServers":{"context-engine":{"command":"cce"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	result := computeProviderHygiene(root, home, false)
	if !result.OK || len(result.OptionalConfigs) != 1 {
		t.Fatalf("result=%#v", result)
	}
}

func TestProviderHygieneRejectsPhoenix(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	path := filepath.Join(home, ".copilot", "mcp-config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"mcpServers":{"phoenix":{"command":"phoenix-mcp"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	result := computeProviderHygiene(root, home, true)
	if result.OK {
		t.Fatal("Phoenix config passed")
	}
}

func TestProviderHygieneIgnoresDisabledPhoenix(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	path := filepath.Join(root, ".mcp.json")
	if err := os.WriteFile(path, []byte(`{"mcpServers":{"phoenix":{"enabled":false,"command":"phoenix-mcp"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if result := computeProviderHygiene(root, home, false); !result.OK {
		t.Fatalf("disabled provider rejected: %#v", result)
	}
}

func TestProviderHygieneDecodesEscapedPhoenix(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	path := filepath.Join(root, ".mcp.json")
	if err := os.WriteFile(path, []byte(`{"mcpServers":{"ph\u006fenix":{"command":"ph\u006fenix-mcp"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if result := computeProviderHygiene(root, home, false); result.OK {
		t.Fatal("escaped Phoenix provider passed")
	}
}

func TestProviderHygieneIgnoresDisabledTomlPhoenix(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	path := filepath.Join(home, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[mcp_servers.phoenix]\nenabled = false\ncommand = \"phoenix-mcp\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if result := computeProviderHygiene(root, home, true); !result.OK {
		t.Fatalf("disabled TOML provider rejected: %#v", result)
	}
}
