//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestC1WindowsJobObjectKillsGrandchild(t *testing.T) {
	root := t.TempDir()
	for attempt := 1; attempt <= 6; attempt++ {
		exe := filepath.Join(root, "fake-codex-"+string(rune('0'+attempt))+".cmd")
		sentinel := filepath.Join(root, "grandchild-survived-"+string(rune('0'+attempt))+".txt")
		script := strings.Join([]string{
			"@echo off",
			"powershell -NoProfile -WindowStyle Hidden -Command \"Start-Process powershell -WindowStyle Hidden -ArgumentList '-NoProfile','-Command','Start-Sleep -Milliseconds 700; Set-Content -LiteralPath ''" + strings.ReplaceAll(sentinel, "'", "''") + "'' survived'\"",
			"powershell -NoProfile -Command Start-Sleep -Seconds 30",
			"exit /b 0",
		}, "\r\n")
		if err := os.WriteFile(exe, []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
		req := modelrouting.DispatchRequest{
			Timeout: time.Millisecond * 120, CWD: root, Model: "large-model", Sandbox: "workspace-write",
			ApprovalPolicy: "never", Network: "none", OutputSchemaPath: filepath.Join(root, "schema.json"),
		}
		result := runWorkerProcess(exe, req, []byte(`{"schema_version":1}`), 1024, root, "")
		if !result.timeout {
			t.Fatalf("attempt %d worker did not time out: err=%v result=%#v", attempt, result.err, result)
		}
		time.Sleep(900 * time.Millisecond)
		if _, err := os.Stat(sentinel); err == nil {
			t.Fatalf("attempt %d grandchild survived process containment and wrote %s", attempt, sentinel)
		} else if !os.IsNotExist(err) {
			t.Fatalf("attempt %d stat sentinel: %v", attempt, err)
		}
	}
}
