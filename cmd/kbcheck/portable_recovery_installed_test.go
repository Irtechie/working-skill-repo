package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// One real npm archive and installation per test process. The Windows TestMain
// removes it after every test using these installed paths has finished.
var recoveryInstall struct {
	once                                   sync.Once
	root, archive, unpacked, home, wrapper string
	err                                    error
}

var recoveryRequiredScripts = []string{"recovery.ps1", "recovery_prepare.ps1", "recovery_continue.ps1", "recovery_dispose.ps1"}

func cleanupPortableRecoveryInstall() error {
	if recoveryInstall.root == "" {
		return nil
	}
	return os.RemoveAll(recoveryInstall.root)
}

func installedRecoveryScript(t *testing.T, target string) string {
	t.Helper()
	recoveryInstall.once.Do(func() { recoveryInstall.err = createPortableRecoveryInstall() })
	if recoveryInstall.err != nil {
		t.Fatalf("packed recovery installation: %v", recoveryInstall.err)
	}
	path := filepath.Join(recoveryInstall.home, "."+target, "skills", "kb-rehab", "scripts")
	if err := validateRecoveryPayload(path); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(path, "recovery.ps1")
}

func validateRecoveryPayload(dir string) error {
	for _, name := range recoveryRequiredScripts {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
			return fmt.Errorf("required installed recovery script missing: %s", name)
		}
	}
	return nil
}

func recoveryBuildCommand(root, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%s %v: %w\n%s", name, args, err, out)
	}
	return out, nil
}

func createPortableRecoveryInstall() error {
	root, err := os.MkdirTemp("", "kb-packed-recovery-")
	if err != nil {
		return err
	}
	recoveryInstall.root = root
	source, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return err
	}
	node, err := exec.LookPath("node")
	if err != nil {
		return err
	}
	npm, err := exec.LookPath("npm")
	if err != nil {
		return err
	}
	// Invoke npm's JavaScript CLI through Node rather than shell interpolation.
	npmCLI := filepath.Join(filepath.Dir(npm), "node_modules", "npm", "bin", "npm-cli.js")
	if _, err := os.Stat(npmCLI); err != nil {
		return fmt.Errorf("npm native Node entrypoint: %w", err)
	}
	packed, err := recoveryBuildCommand(source, node, npmCLI, "pack", "--ignore-scripts", "--json", "--pack-destination", root)
	if err != nil {
		return err
	}
	var packs []struct {
		Filename string `json:"filename"`
	}
	if err := json.Unmarshal(packed, &packs); err != nil || len(packs) != 1 {
		return fmt.Errorf("npm pack result: %s", packed)
	}
	if filepath.Base(packs[0].Filename) != packs[0].Filename {
		return fmt.Errorf("unsafe package filename")
	}
	recoveryInstall.unpacked = filepath.Join(root, "unpacked")
	recoveryInstall.archive = filepath.Join(root, packs[0].Filename)
	if err := extractRecoveryPackage(recoveryInstall.archive, recoveryInstall.unpacked); err != nil {
		return err
	}
	packageRoot := filepath.Join(recoveryInstall.unpacked, "package")
	if err := validateRecoveryPayload(filepath.Join(packageRoot, ".github", "skills", "kb-rehab", "scripts")); err != nil {
		return err
	}
	recoveryInstall.home = filepath.Join(root, "home")
	if _, err := recoveryBuildCommand(packageRoot, node, filepath.Join(packageRoot, "bin", "kb-install.mjs"), "--target", "all", "--install-root", recoveryInstall.home, "--router", "skip", "--reconciler", "skip", "--yes"); err != nil {
		return err
	}
	for _, target := range []string{"codex", "copilot", "agents"} {
		dir := filepath.Join(recoveryInstall.home, "."+target, "skills", "kb-rehab", "scripts")
		if err := validateRecoveryPayload(dir); err != nil {
			return err
		}
		for _, name := range recoveryRequiredScripts {
			packedBytes, err := os.ReadFile(filepath.Join(packageRoot, ".github", "skills", "kb-rehab", "scripts", name))
			if err != nil {
				return err
			}
			installed, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				return err
			}
			if string(installed) != string(packedBytes) {
				return fmt.Errorf("installed payload differs: %s/%s", target, name)
			}
		}
	}
	recoveryInstall.wrapper = filepath.Join(root, "assert-consumer.ps1")
	return os.WriteFile(recoveryInstall.wrapper, []byte(`param([string]$Helper,[string]$Root,[string]$Action,[string]$Request,[string]$AllowedNative)
$ErrorActionPreference='Stop'
foreach($name in @('go','node','kbcheck','kbreconcile')) {
  $found=Get-Command $name -ErrorAction SilentlyContinue
  if($found) {
    if($name -ne 'kbreconcile' -or -not $AllowedNative -or $found.Source -ne $AllowedNative) { throw "contaminated consumer executable: $name" }
  } elseif($name -eq 'kbreconcile' -and $AllowedNative) { throw 'explicit native fixture unavailable' }
}
foreach($relative in @('cmd','.github/skills/kb-rehab','config/rehab-policy.json')) {
  if(Test-Path -LiteralPath (Join-Path $Root $relative)) { throw "contaminated consumer repository: $relative" }
}
$invoke=@{Root=$Root;Action=$Action;Json=$true}
if($Request){$invoke.Request=$Request}
& $Helper @invoke
`), 0600)
}

func extractRecoveryPackage(archive, destination string) error {
	file, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()
	r := tar.NewReader(gz)
	for {
		h, err := r.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := filepath.Clean(filepath.FromSlash(h.Name))
		if filepath.IsAbs(name) || filepath.VolumeName(name) != "" || name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return fmt.Errorf("unsafe packed path")
		}
		path := filepath.Join(destination, name)
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			out, err := os.Create(path)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, r)
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		default:
			return fmt.Errorf("unsupported packed entry: %s", h.Name)
		}
	}
}

func portableConsumerCommand(t *testing.T, ctx context.Context, script, repo, action, request, nativeDir string) *exec.Cmd {
	t.Helper()
	installedRecoveryScript(t, "agents")
	ps, err := exec.LookPath("powershell.exe")
	if err != nil {
		t.Fatal(err)
	}
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", recoveryInstall.wrapper, "-Helper", script, "-Root", repo, "-Action", action}
	if request != "" {
		args = append(args, "-Request", request)
	}
	if nativeDir != "" {
		args = append(args, "-AllowedNative", filepath.Join(nativeDir, "kbreconcile.cmd"))
	}
	cmd := exec.CommandContext(ctx, ps, args...)
	cmd.Dir = repo
	for _, v := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(v), "PATH=") {
			cmd.Env = append(cmd.Env, v)
		}
	}
	childPath := filepath.Dir(git) + ";" + filepath.Join(os.Getenv("SystemRoot"), "System32")
	if nativeDir != "" {
		childPath += ";" + nativeDir
	}
	cmd.Env = append(cmd.Env, "PATH="+childPath)
	return cmd
}

func TestPortableRecoveryPackedTargetsAndMissingPayload(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows installed payload proof")
	}
	repo := t.TempDir()
	portableGit(t, repo, "init", "--initial-branch=main")
	for _, target := range []string{"codex", "copilot", "agents"} {
		t.Run(target, func(t *testing.T) { portableRunSurvey(t, installedRecoveryScript(t, target), repo) })
	}
	// Tamper only a disposable copy of the extracted package, never shared install.
	bad := t.TempDir()
	if err := extractRecoveryPackage(recoveryInstall.archive, bad); err != nil {
		t.Fatal(err)
	}
	badPackage := filepath.Join(bad, "package")
	if err := os.Remove(filepath.Join(badPackage, ".github/skills/kb-rehab/scripts/recovery_prepare.ps1")); err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	badHome := filepath.Join(bad, "home")
	if _, err := recoveryBuildCommand(badPackage, node, filepath.Join(badPackage, "bin/kb-install.mjs"), "--target", "agents", "--install-root", badHome, "--router", "skip", "--reconciler", "skip", "--yes"); err != nil {
		t.Fatal(err)
	}
	if err := validateRecoveryPayload(filepath.Join(badHome, ".agents/skills/kb-rehab/scripts")); err == nil || !strings.Contains(err.Error(), "recovery_prepare.ps1") {
		t.Fatalf("missing packed helper escaped installed oracle: %v", err)
	}
	// The exact child preflight used by every scenario rejects a consumer helper.
	portableWrite(t, filepath.Join(repo, "cmd", "kbcheck", "main.go"), "contamination")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := portableConsumerCommand(t, ctx, installedRecoveryScript(t, "agents"), repo, "survey", "", "")
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "contaminated consumer repository") {
		t.Fatalf("contamination was not rejected: %v %s", err, out)
	}
	cleanConsumer := t.TempDir()
	contaminant := t.TempDir()
	portableWrite(t, filepath.Join(contaminant, "kbcheck.cmd"), "@exit /b 0")
	cmd = portableConsumerCommand(t, ctx, installedRecoveryScript(t, "agents"), cleanConsumer, "survey", "", "")
	for i, v := range cmd.Env {
		if strings.HasPrefix(strings.ToUpper(v), "PATH=") {
			cmd.Env[i] = "PATH=" + contaminant + ";" + v[len("PATH="):]
		}
	}
	out, err = cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "contaminated consumer executable") {
		t.Fatalf("accidental native tool was not rejected: %v %s", err, out)
	}
}
