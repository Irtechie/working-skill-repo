package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRoutingAgentArgsAreNoninteractiveAndLiteral(t *testing.T) {
	// The test binary is a native executable on every supported host.
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	prompt := "quotes \" ' & | > $() ` newline\nUnicode café"
	for _, host := range []string{"codex", "ghcp"} {
		args, err := routingAgentArgs(binary, host, prompt)
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, arg := range args {
			if arg == prompt {
				count++
			}
		}
		if count != 1 || args[0] != binary {
			t.Fatalf("prompt or executable changed: %q", args)
		}
		joined := strings.Join(args, " ")
		if host == "codex" && (args[1] != "exec" || !strings.Contains(joined, "--sandbox read-only")) {
			t.Fatal(args)
		}
		if host == "ghcp" && (args[1] != "--prompt" || !strings.Contains(joined, "--deny-tool=write") || !strings.Contains(joined, "--deny-tool=shell")) {
			t.Fatal(args)
		}
		if strings.Contains(joined, "--allow-all") {
			t.Fatal("unbounded tool permissions")
		}
	}
	if _, err := routingAgentArgs(binary, "unknown", prompt); err == nil {
		t.Fatal("unknown runtime accepted")
	}
}

func TestRoutingWorkspaceDoesNotExposeEvaluator(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github/skills/kb-fix/SKILL.md"), "public skill")
	writeFile(t, filepath.Join(root, "evals/route-complexity/case.json"), "SECRET-ORACLE")
	writeFile(t, filepath.Join(root, "cmd/kbcheck/scorer.go"), "SECRET-SCORER")
	workspace, err := prepareRoutingWorkspace(root, filepath.Join(root, ".kb/run"))
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(workspace, ".github/skills/kb-fix/SKILL.md")); err != nil || string(data) != "public skill" {
		t.Fatal("skill payload missing", err)
	}
	for _, path := range []string{"evals", "cmd"} {
		if _, err := os.Stat(filepath.Join(workspace, path)); !os.IsNotExist(err) {
			t.Fatal("exposed evaluator", path)
		}
	}
	out, err := exec.Command("git", "-C", workspace, "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil || filepath.Clean(strings.TrimSpace(string(out))) != filepath.Clean(workspace) {
		t.Fatalf("not isolated: %s %v", out, err)
	}
}

func TestRoutingNativeInvocationRetainsEvidence(t *testing.T) {
	root := t.TempDir()
	source := `package main
import("fmt";"os";"encoding/json")
func main(){
 if len(os.Args)<3 || (os.Args[1]!="exec" && os.Args[1]!="--prompt") {os.Exit(19)}
 fmt.Fprint(os.Stderr,"captured diagnostic")
 if os.Getenv("KB_ROUTING_FAIL")=="1" {fmt.Print("raw failure");os.Exit(23)}
 response:="{\"fixture_id\":\"case\",\"eval_run_id\":\"run\"}"
 if os.Args[1]=="--prompt" {
  json.NewEncoder(os.Stdout).Encode(map[string]any{"type":"assistant.message","data":map[string]any{"content":response}})
  fmt.Println("{\"type\":\"result\",\"sessionId\":\"s\",\"exitCode\":0}")
 } else {fmt.Print(response)}
}`
	file := filepath.Join(root, "main.go")
	writeFile(t, file, source)
	binary := filepath.Join(root, "fake-agent")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", binary, file).CombinedOutput(); err != nil {
		t.Fatalf("build: %s %v", out, err)
	}
	for _, host := range []string{"codex", "ghcp"} {
		t.Run(host, func(t *testing.T) {
			_, process, err := invokeLiveAgent(root, host, map[string]any{"id": "case"}, "run", options{agentCommand: binary})
			if err != nil || process.ExitCode != 0 || process.Stderr != "captured diagnostic" {
				t.Fatalf("launch: %+v %v", process, err)
			}
			t.Setenv("KB_ROUTING_FAIL", "1")
			_, process, err = invokeLiveAgent(root, host, map[string]any{"id": "case"}, "run", options{agentCommand: binary})
			if err == nil || process.ExitCode != 23 || process.Stdout != "raw failure" || process.Stderr != "captured diagnostic" {
				t.Fatalf("failure lost: %+v %v", process, err)
			}
		})
	}
}

func TestCopilotStructuredCompletion(t *testing.T) {
	message := `{"type":"assistant.message","data":{"content":"{\"fixture_id\":\"case\",\"eval_run_id\":\"run\"}"}}` + "\n"
	terminal := `{"type":"result","sessionId":"s","exitCode":0}`
	if result, err := parseCopilotEventStream(message + terminal); err != nil || result["fixture_id"] != "case" {
		t.Fatal(result, err)
	}
	for _, bad := range []string{message, terminal, message + `{"type":"result","sessionId":"s","exitCode":1}`, message + `{"type":"result","sessionId":"s"}`, message + `{"type":"session.error"}` + "\n" + terminal, message + terminal + "\n{}", "not json\n" + message + terminal} {
		if _, err := parseCopilotEventStream(bad); err == nil {
			t.Fatalf("accepted invalid stream: %s", bad)
		}
	}
}

func TestEvalPromptWithholdsPrivateFixtureFields(t *testing.T) {
	t.Parallel()
	fixture := map[string]any{
		"id": "canary-case", "user_prompt": "route this request", "repo_state": map[string]any{"branch": "main"},
		"expected": map[string]any{"route": "SECRET-EXPECTED-ROUTE"},
		"guards":   []any{"SECRET-GUARD"}, "scoring_metadata": map[string]any{"canary": "SECRET-SCORER"},
	}
	prompt := evalPrompt(fixture, "codex", "run-1")
	for _, private := range []string{"SECRET-EXPECTED-ROUTE", "SECRET-GUARD", "SECRET-SCORER", "expected_result \"pass\""} {
		if strings.Contains(prompt, private) {
			t.Errorf("private or self-scored fixture content leaked into prompt: %q", private)
		}
	}
	for _, public := range []string{"canary-case", "route this request", "\"branch\": \"main\"", "scorer, not you, determines pass or fail"} {
		if !strings.Contains(prompt, public) {
			t.Errorf("public prompt content missing: %q", public)
		}
	}
}

func TestEvalAdapterMarksSyntheticEvidence(t *testing.T) {
	t.Parallel()
	fixture := map[string]any{"id": "fixture", "expected": map[string]any{"route": "kb-fix"}}
	if got := stringValue(dryRunResult(fixture, "codex", "run")["evidence_kind"]); got != "synthetic" {
		t.Fatalf("dry run evidence kind=%q", got)
	}
	manifest := newRunManifest(".", "run", "codex", "dry-run", fixture)
	if got := stringValue(manifest["evidence_kind"]); got != "synthetic" {
		t.Fatalf("dry-run manifest evidence kind=%q", got)
	}
	if got := stringValue(newRunManifest(".", "run", "codex", "live", fixture)["evidence_kind"]); got != "live" {
		t.Fatalf("live manifest evidence kind=%q", got)
	}
}

func TestOpenCodeEventStreamRequiresFinalResult(t *testing.T) {
	t.Parallel()
	stream := openCodeTestStream(`{"actual":{"route":"kb-fix"}}`)
	result, err := parseOpenCodeEventStream(stream)
	if err != nil || stringValue(result["actual"].(map[string]any)["route"]) != "kb-fix" {
		t.Fatalf("valid OpenCode stream rejected: result=%#v err=%v", result, err)
	}
	for name, bad := range map[string]string{
		"invented result":   `{"result":{"actual":{"route":"kb-fix"}}}`,
		"missing finish":    strings.Split(stream, "\n")[0],
		"malformed":         stream + "{",
		"error":             `{"type":"error","sessionID":"s","error":{"name":"APIError"}}`,
		"wrong session":     strings.Replace(stream, `"sessionID":"s"`, `"sessionID":"other"`, 1),
		"incomplete reason": strings.Replace(stream, `"stop"`, `"length"`, 1),
		"prose":             openCodeTestStream("Here is JSON: {}"),
	} {
		if _, err := parseOpenCodeEventStream(bad); err == nil {
			t.Errorf("%s passed", name)
		}
	}
}

func openCodeTestStream(text string) string {
	part, _ := json.Marshal(map[string]any{"type": "text", "sessionID": "s", "part": map[string]any{"type": "text", "text": text}})
	return string(part) + "\n" + `{"type":"step_finish","sessionID":"s","part":{"type":"step-finish","reason":"stop"}}` + "\n"
}

func TestEvalPromptUsesRepositoryPromptField(t *testing.T) {
	if !strings.Contains(evalPrompt(map[string]any{"id": "case", "prompt": "actual request"}, "opencode", "run"), "actual request") {
		t.Fatal("repository prompt omitted")
	}
}

// This native fixture exercises the real bounded process runner and argv. It
// emits the documented 1.18.23 envelopes; it never invokes a model or shell.
func buildOpenCodeFake(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	source := `package main
import("encoding/json"; "fmt"; "os"; "strings"; "time")
func main(){
 if p:=os.Getenv("KB_OC_MARKER");p!=""{os.WriteFile(p,[]byte("called"),0600)}
 if len(os.Args)!=5||os.Args[1]!="run"||os.Args[2]!="--format"||os.Args[3]!="json"{fmt.Fprintln(os.Stderr,"bad argv");os.Exit(19)}
 fmt.Fprint(os.Stderr,"native stderr")
 switch os.Getenv("KB_OC_MODE") {case "hang":fmt.Print("before hang");time.Sleep(10*time.Second);return;case "error":fmt.Print("raw failure");os.Exit(23);case "malformed":fmt.Print("{broken");return}
 result:=map[string]any{"echo":os.Args[4]}
 if strings.Contains(os.Args[4],"Set eval_run_id exactly to") {result=map[string]any{"fixture_id":"wrong-case","eval_run_id":"wrong-run"}}
 b,_:=json.Marshal(result)
 json.NewEncoder(os.Stdout).Encode(map[string]any{"type":"text","sessionID":"s","part":map[string]any{"type":"text","text":string(b)}})
 json.NewEncoder(os.Stdout).Encode(map[string]any{"type":"step_finish","sessionID":"s","part":map[string]any{"type":"step-finish","reason":"stop"}})
}`
	file := filepath.Join(root, "main.go")
	writeFile(t, file, source)
	binary := filepath.Join(root, "fake-opencode")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, file)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fake: %v %s", err, out)
	}
	return binary
}

func TestOpenCodeNativeProcessEvidenceAndBounds(t *testing.T) {
	binary := buildOpenCodeFake(t)
	root := t.TempDir()
	prompt := "quotes \" ' & | > $() ` newline\nUnicode café"
	result, process, err := invokeOpenCode(root, binary, prompt, 5*time.Second)
	if err != nil || process.ExitCode != 0 || result["echo"] != prompt || process.Stderr != "native stderr" || !strings.Contains(process.Stdout, `"step_finish"`) {
		t.Fatalf("native argv/evidence: result=%v process=%+v err=%v", result, process, err)
	}
	for _, mode := range []string{"error", "malformed", "hang"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("KB_OC_MODE", mode)
			bound := 5 * time.Second
			if mode == "hang" {
				bound = 250 * time.Millisecond
			}
			_, got, err := invokeOpenCode(root, binary, prompt, bound)
			if err == nil {
				t.Fatal("failure passed")
			}
			if mode == "error" && (got.ExitCode != 23 || got.Stdout != "raw failure" || got.Stderr != "native stderr") {
				t.Fatalf("lost failure evidence: %+v", got)
			}
			if mode == "hang" && got.ExitCode != 124 {
				t.Fatalf("timeout code: %+v", got)
			}
			if mode == "malformed" && got.Stdout != "{broken" {
				t.Fatalf("lost malformed stream: %+v", got)
			}
		})
	}
	_, _, err = invokeLiveAgent(root, "opencode", map[string]any{"id": "case"}, "run", options{agentCommand: binary})
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("wrong identity admitted: %v", err)
	}
	t.Run("persist-failure", func(t *testing.T) {
		t.Setenv("KB_OC_MODE", "error")
		run, err := runOneAdapterFixture(root, "runs", "opencode", "live", map[string]any{"id": "case", "expected": map[string]any{"route": "SECRET-EXPECTED"}}, options{agentCommand: binary})
		if err != nil || run.Status != "fail" || run.ExitCode != 23 {
			t.Fatalf("live failure: %+v %v", run, err)
		}
		for name, expected := range map[string]string{"stdout.txt": "raw failure", "stderr.txt": "native stderr"} {
			data, err := os.ReadFile(filepath.Join(run.RunDir, name))
			if err != nil || string(data) != expected {
				t.Fatalf("%s lost raw evidence: %q %v", name, data, err)
			}
		}
		data, err := os.ReadFile(run.ResultPath)
		if err != nil || strings.Contains(string(data), "SECRET-EXPECTED") || strings.Contains(string(data), `"actual"`) {
			t.Fatalf("live failure copied expected answers: %s %v", data, err)
		}
	})
	t.Run("dry-run", func(t *testing.T) {
		marker := filepath.Join(root, "child-called")
		t.Setenv("KB_OC_MARKER", marker)
		_, err := runOneAdapterFixture(root, "runs", "opencode", "dry-run", map[string]any{"id": "case", "expected": map[string]any{"route": "kb-fix"}}, options{agentCommand: binary})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatal("dry-run invoked child")
		}
	})
}

func TestEvalAdapterFailureNeverBecomesSyntheticSuccess(t *testing.T) {
	root := t.TempDir()
	fixture := map[string]any{"id": "case", "prompt": "request", "expected": map[string]any{"route": "PRIVATE-EXPECTED"}}
	if err := writeAdapterJSON(filepath.Join(root, "missing", "no.json"), fixture); err == nil {
		t.Fatal("write failure swallowed")
	}
	writeFile(t, filepath.Join(root, "evals", "route-complexity", "case.json"), `{"id":"case","expected":{"route":"PRIVATE-EXPECTED"}}`)
	opts := options{fixtureID: "case", agentCommand: filepath.Join(root, "missing-cli"), keepRun: true, json: true, runtime: "opencode", runner: "eval-run-opencode"}
	result, err := runEvalAdapter(root, opts, "opencode")
	if err != nil || result.OK || len(result.Runs) != 1 || result.Runs[0].ExitCode != 127 {
		t.Fatalf("missing child: %+v %v", result, err)
	}
	content, err := os.ReadFile(result.Runs[0].ResultPath)
	if err != nil || strings.Contains(string(content), "PRIVATE-EXPECTED") || strings.Contains(string(content), `"synthetic"`) || strings.Contains(string(content), `"actual"`) {
		t.Fatalf("fabricated live result: %s %v", content, err)
	}
	for _, command := range []func(string, options, *strings.Builder, *strings.Builder) int{
		func(r string, o options, out, err *strings.Builder) int {
			return runEvalAdapterCommand(r, o, "opencode", out, err)
		},
		func(r string, o options, out, err *strings.Builder) int {
			return runEvalLiveCorpusCommand(r, o, out, err)
		},
		func(r string, o options, out, err *strings.Builder) int {
			return runSkillEvalWrapCommand(r, o, out, err)
		},
	} {
		var out, stderr strings.Builder
		if code := command(root, opts, &out, &stderr); code == 0 || !strings.Contains(out.String(), `"ok": false`) {
			t.Fatalf("failed adapter propagated success: %d %s %s", code, out.String(), stderr.String())
		}
	}
}

func TestOpenCodeWindowsShimResolvesNativeSibling(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows npm layout")
	}
	root := t.TempDir()
	shim := filepath.Join(root, "opencode.cmd")
	writeFile(t, shim, "never execute this shell")
	if _, err := resolveOpenCodeExecutable(shim); err == nil {
		t.Fatal("missing native fallback passed")
	}
	native := filepath.Join(root, "node_modules", "opencode-ai", "bin", "opencode.exe")
	writeFile(t, native, "fixture")
	if got, err := resolveOpenCodeExecutable(shim); err != nil || got != native {
		t.Fatalf("native resolution: %s %v", got, err)
	}
}
