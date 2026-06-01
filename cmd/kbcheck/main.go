package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const usage = `kbcheck is the native KB gate entrypoint.

Usage:
  kbcheck core [--root <path>] [--list] [--dry-run]
  kbcheck local-release [--root <path>] [--json] [--dry-run]
  kbcheck live-release [--root <path>] [--json] [--dry-run]
  kbcheck help

Commands:
  core           Discover and run local deterministic checks.
  local-release  Run the local release gate with required and optional checks.
  live-release   Run local release checks plus explicit live-model surfaces.
`

type processRunner func(root string, check Check) CheckResult

type options struct {
	command string
	root    string
	json    bool
	dryRun  bool
	list    bool
}

func main() {
	code := run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}

func run(args []string, stdout, stderr io.Writer) int {
	opts, err := parse(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		fmt.Fprintln(stderr)
		fmt.Fprint(stderr, usage)
		return 2
	}

	if opts.command == "help" {
		fmt.Fprint(stdout, usage)
		return 0
	}

	root, err := filepath.Abs(opts.root)
	if err != nil {
		fmt.Fprintf(stderr, "resolve root: %v\n", err)
		return 1
	}

	switch opts.command {
	case "core":
		return runCore(root, opts, stdout, stderr, runProcessCheck)
	case "local-release", "live-release":
		return runRelease(root, opts, stdout, stderr, runProcessCheck)
	default:
		fmt.Fprintf(stderr, "unsupported command %q\n", opts.command)
		return 2
	}
}

func parse(args []string) (options, error) {
	if len(args) == 0 {
		return options{command: "help", root: "."}, nil
	}

	cmd := args[0]
	if cmd == "-h" || cmd == "--help" || cmd == "help" {
		return options{command: "help", root: "."}, nil
	}
	if cmd != "core" && cmd != "local-release" && cmd != "live-release" {
		return options{}, fmt.Errorf("unknown command %q", cmd)
	}

	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	opts := options{command: cmd, root: "."}
	fs.StringVar(&opts.root, "root", ".", "repository root")
	fs.BoolVar(&opts.json, "json", false, "emit JSON when supported")
	fs.BoolVar(&opts.dryRun, "dry-run", false, "print commands instead of running them")
	fs.BoolVar(&opts.list, "list", false, "list checks without running them")
	if err := fs.Parse(args[1:]); err != nil {
		return options{}, err
	}
	if fs.NArg() > 0 {
		return options{}, fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	if opts.command == "core" && opts.json {
		return options{}, fmt.Errorf("--json is only supported for release gate commands")
	}
	if opts.command != "core" && opts.list {
		return options{}, fmt.Errorf("--list is only supported for core")
	}
	return opts, nil
}

func runCore(root string, opts options, stdout, stderr io.Writer, runner processRunner) int {
	checks, err := DiscoverChecks(root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.list || opts.dryRun {
		printChecks(stdout, checks)
		return 0
	}

	for _, check := range checks {
		fmt.Fprintf(stdout, "==> %s: %s\n", check.Name, check.CommandString())
		result := runner(root, check)
		if result.Stdout != "" {
			fmt.Fprint(stdout, result.Stdout)
			if !strings.HasSuffix(result.Stdout, "\n") {
				fmt.Fprintln(stdout)
			}
		}
		if result.Stderr != "" {
			fmt.Fprint(stderr, result.Stderr)
			if !strings.HasSuffix(result.Stderr, "\n") {
				fmt.Fprintln(stderr)
			}
		}
		if result.ExitCode != 0 {
			fmt.Fprintf(stderr, "check failed: %s\n", check.Name)
			return result.ExitCode
		}
	}
	return 0
}

func printChecks(w io.Writer, checks []Check) {
	for _, check := range checks {
		fmt.Fprintf(w, "%-40s %s\n", check.Name, check.CommandString())
	}
}

func runProcessCheck(root string, check Check) CheckResult {
	if check.Run != nil {
		return check.Run(root)
	}
	if len(check.Args) == 0 {
		return CheckResult{ExitCode: 1, Stderr: "check has no command"}
	}
	cmd := exec.Command(check.Args[0], check.Args[1:]...)
	cmd.Dir = root
	out, err := cmd.Output()
	result := CheckResult{ExitCode: 0, Stdout: string(out)}
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			result.Stderr = string(exitErr.Stderr)
			return result
		}
		result.ExitCode = 1
		result.Stderr = err.Error()
		return result
	}
	return result
}

func findPowerShell() (string, error) {
	if override := os.Getenv("KBCHECK_POWERSHELL"); override != "" {
		return override, nil
	}
	for _, candidate := range []string{"pwsh", "powershell"} {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}
	if runtime.GOOS == "windows" {
		if path, err := exec.LookPath("powershell.exe"); err == nil {
			return path, nil
		}
	}
	return "", errors.New("PowerShell not found; install PowerShell 7 (pwsh) or set KBCHECK_POWERSHELL")
}

func powerShellArgs(script string, extra ...string) ([]string, error) {
	ps, err := findPowerShell()
	if err != nil {
		return nil, err
	}
	base := strings.ToLower(filepath.Base(ps))
	args := []string{ps}
	if base == "powershell" || base == "powershell.exe" {
		args = append(args, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script)
	} else {
		args = append(args, "-NoProfile", "-File", script)
	}
	args = append(args, extra...)
	return args, nil
}
