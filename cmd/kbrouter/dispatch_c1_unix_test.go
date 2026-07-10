//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestC1UnixDispatchFailsClosedWithoutProvenContainment(t *testing.T) {
	previous := dispatchProcessTreeContainment
	dispatchProcessTreeContainment = ensureProcessTreeContainment
	t.Cleanup(func() { dispatchProcessTreeContainment = previous })
	root := t.TempDir()
	exe := filepath.Join(root, "fake-codex.sh")
	sentinel := filepath.Join(root, "worker-started.txt")
	script := strings.Join([]string{
		"#!/bin/sh",
		"echo started > " + quoteShell(sentinel),
		"exit 0",
	}, "\n")
	if err := os.WriteFile(exe, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	req := modelrouting.DispatchRequest{
		Timeout: time.Second, CWD: root, Model: "large-model", Sandbox: "workspace-write",
		ApprovalPolicy: "never", Network: "none", OutputSchemaPath: filepath.Join(root, "schema.json"),
	}
	result := runWorkerProcess(exe, req, []byte(`{"schema_version":1}`), 1024, root, "")
	if !result.notStarted || result.err == nil || !strings.Contains(result.err.Error(), "no proven process-tree containment") {
		t.Fatalf("Unix dispatch did not fail closed: %#v", result)
	}
	time.Sleep(100 * time.Millisecond)
	if _, err := os.Stat(sentinel); err == nil {
		t.Fatalf("worker started despite unavailable containment: %s", sentinel)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat sentinel: %v", err)
	}
}
