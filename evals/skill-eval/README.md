# Skill Eval Results

This directory contains deterministic scoring fixtures for skill routing and
proof behavior.

The runner does not call a model. It scores captured agent results against the
route-complexity dataset:

```powershell
go run ./cmd/kbcheck skill-eval
```

Default mode is a self-test under `evals/skill-eval/selftest/`. The self-test
contains one valid result and several intentionally bad results. The runner must
pass the good result and fail the bad ones; otherwise the scorer is too weak.

For real captured runs, write a result JSON and pass it explicitly:

```powershell
go run ./cmd/kbcheck skill-eval --result-path path\to\captured-result.json
```

Adapter prompts contain only the fixture id, user prompt, and repository state.
Expected routes, guards, scoring metadata, and other oracle fields stay with
the local scorer. The model never declares pass/fail; the scorer computes it.
Adapter manifests mark dry-runs as `synthetic` and actual captures as `live`.
Synthetic records exercise plumbing only and are not runtime-qualification or
task-improvement evidence.

## Result Shape

```json
{
  "id": "run-id",
  "fixture_id": "tiny-typo-fix",
  "expected_result": "pass",
  "actual": {
    "route": "kb-fix",
    "user_questions": 0,
    "artifacts": ["changed file", "verification note"],
    "proof": ["git diff --check", "targeted text/render check if UI-visible"]
  },
  "trace": {
    "files_read": ["todo.md"],
    "commands": ["git diff --check"],
    "tools": ["shell"]
  },
  "observed_trace": {
    "captured": true,
    "method": "path-shim+git-diff",
    "commands": [],
    "writes": [],
    "deletes": []
  },
  "claim_checks": [
    {
      "type": "command_ran",
      "path": "",
      "contains": "git diff --check",
      "expected": true,
      "claim": "Agent claimed git diff --check was run"
    }
  ]
}
```

Supported claim checks:

- `file_exists`
- `command_ran`
- `file_read`

`trace` is model-reported intent evidence. It is useful for checking the route's
planned files, commands, and tools, but it is not observed behavior.

Optional `observed_trace` is the externally captured safety layer. It is added
by `go run ./cmd/kbcheck skill-eval-wrap` and currently records PATH-shim command hits
plus git-status write/delete changes. Existing results may omit it; omitted
observation is reported as lower confidence instead of silently treated as
proof.

Optional `trace_rules` fields make trace discipline deterministic when a fixture
or captured result needs stronger proof than route/proof strings alone:

```json
{
  "trace_rules": {
    "required_files_read": ["docs/context/eval-map.md"],
    "required_commands": ["git diff --check"],
    "required_tools": ["shell"],
    "forbidden_files_read": ["secrets.env"],
    "forbidden_commands": ["git reset --hard"],
    "forbidden_tools": ["browser"]
  }
}
```

Required rules are intent checks: they pass when any model-reported trace item
contains the expected text after case-insensitive whitespace normalization.

Forbidden command/tool rules are safety checks. When `observed_trace.captured`
is true, forbidden commands are enforced against `observed_trace.commands`.
When observation is missing, the scorer falls back to model-reported `trace` and
emits a self-reported confidence warning. `observed_trace` has no tools field in
v1, so forbidden tools remain self-reported unless a future wrapper captures
real tool calls.

For routing evals, observed writes/deletes must be empty. If observation is
missing, the no-write invariant is recorded as unverified.

File reads are intentionally not observed in v1. `required_files_read` and
`forbidden_files_read` use model-reported trace only.

These are scorer-only rules; live adapter JSON schemas do not need to include
them unless a future adapter wants the model to emit them directly.

## Observed Trace Wrapper

Wrap a dry-run adapter with external command/write capture:

```powershell
go run ./cmd/kbcheck skill-eval-wrap --fixture-id tiny-typo-fix --dry-run --sealed
```

The wrapper prepends temporary PATH shims for selected commands, runs the
existing adapter with preserved result JSON, adds `observed_trace`, then reruns
`skill-eval` against the augmented result. Without `--keep-run`, the
temporary run directory is removed after scoring.

`-Sealed` logs dangerous attempts and blocks known destructive patterns such as
`git reset --hard`, `git clean -fd`, `git checkout --`, and `rm`.

Limits are explicit:

- PATH shims catch invocation by command name through inherited PATH. Absolute
  paths, internal tool APIs, containers, or subprocesses with a different
  environment can bypass them.
- Git-status write/delete comparison is the backstop for file mutations in the
  repo, excluding `.atv/eval-runs/` and `.atv/tmp/`.
- Reads are out of scope for observed capture.

## Claim Artifact Verifier

Transcript-derived claim checks live outside the model transcript as JSON claim
artifacts:

```powershell
go run ./cmd/kbcheck skill-eval-claims
```

Supported deterministic claim types match structured result claim checks:
`file_exists`, `command_ran`, and `file_read`. A claim with
`"type": "ambiguous"` is reported with an ambiguous count and does not become
proof. False deterministic claims fail; ambiguous claims are visible but do not
fail by themselves.

`skill-eval` also checks any result-level `claim_artifacts` array by running
each artifact through `skill-eval-claims`.

## Output Quality Rubric

Run the quality rubric self-test:

```powershell
go run ./cmd/kbcheck skill-eval-quality
```

The rubric is separate from deterministic route/proof/claim pass/fail. It
computes deterministic quality scores from raw captured result JSON on five 0-5
dimensions:

- `completeness`
- `maintainability`
- `relevance`
- `proof_quality`
- `right_sized_ceremony`

Each dimension records `score`, `judge`, and `reason`. The current computed
scorer uses `deterministic` judgments only: route correctness, required output
fields, proof coverage, bounded ceremony, and vague-output heuristics. It is an
independent scorer for code-checkable behavior, not a subjective LLM judge.

Live adapters produce this result shape from transcripts/traces, then let this
deterministic scorer decide pass/fail. Codex and GHCP adapters exist; live model
runs are explicit because they require runtime auth and spend.

## Codex Adapter

Run the Codex adapter in safe dry-run mode:

```powershell
go run ./cmd/kbcheck eval-run-codex --fixture-id tiny-typo-fix --dry-run
```

Run one live Codex eval:

```powershell
go run ./cmd/kbcheck eval-run-codex --fixture-id tiny-typo-fix --keep-run
```

The live adapter creates a disposable git worktree under `.atv/eval-runs/`, runs
`codex exec` in read-only mode with a JSON output schema, writes `result.json`,
then calls `skill-eval --result-path <result.json>`.

Dry-run mode is part of `core`; live mode is explicit because it calls a
model. Dry-run artifacts are cleaned unless `--keep-run` is set.

## GHCP Adapter

Run the GHCP adapter in safe dry-run mode:

```powershell
go run ./cmd/kbcheck eval-run-ghcp --fixture-id tiny-typo-fix --dry-run
```

Run one live GHCP eval:

```powershell
go run ./cmd/kbcheck eval-run-ghcp --fixture-id tiny-typo-fix --keep-run
```

The GHCP adapter creates a disposable git worktree under `.atv/eval-runs/`, runs
GitHub Copilot CLI non-interactively, captures stdout/stderr plus a transcript
artifact when available, parses strict JSON from the final response, writes
`result.json`, then calls `skill-eval --result-path <result.json>`.

GHCP does not expose a Codex-style `--output-schema` flag in the currently
observed local CLI help, so this adapter uses prompt-level JSON constraints and
deterministic parsing. Invalid or missing JSON is a hard adapter failure, not a
pass.

Dry-run artifacts are cleaned unless `--keep-run` is set.

## Live Corpus Runner

### OpenCode adapter

`eval-run-opencode` supports static, fixture, and synthetic dry-run checks.
Live model behavior is **unverified**. The native-process tests use a fake CLI
and the [OpenCode 1.18.23 event contract](https://github.com/anomalyco/opencode/blob/v1.18.23/packages/opencode/src/cli/cmd/run.ts);
they do not establish agent routing quality or a successful authenticated run.

```powershell
go run ./cmd/kbcheck eval-run-opencode --fixture-id tiny-typo-fix --dry-run --keep-run --json
```

After explicit authorization for a model call, the optional live smoke command is:

```powershell
go run ./cmd/kbcheck eval-run-opencode --fixture-id tiny-typo-fix --keep-run --json
```

The adapter executes `opencode run --format json <prompt>` as native arguments
in the selected repository. On Windows it resolves the native executable behind
the supported npm shim layout (`node_modules/opencode-ai/bin/opencode.exe`);
unsupported or missing native launchers fail without evaluating shell text.
Execution uses the existing process-tree containment, output cap, and a five-minute
timeout. No server, sharing, attachment, or automatic permission flag is added.

Each run retains `stdout.txt`, `stderr.txt`, `result.json`, `manifest.json`, and
`score.json` under `.kb/eval-runs/<run-id>/`. Raw process output survives parser
and child failures. Text events must belong to one session, contain a JSON
assistant result, and end with a normal `step_finish`; the returned fixture/run
IDs must match the invocation. A missing child, error, malformed stream, timeout,
or wrong identity fails the adapter, corpus, and wrapper. Failed live results
contain diagnostics rather than synthetic expected answers.

OpenCode is opt-in with `--runtime opencode` in the corpus or
`--runner eval-run-opencode` in the wrapper. Before promoting live capability,
retain the authorized run's CLI version, exact command and exit, raw events,
result/manifest, and independent score. Help output and fixture tests alone
are insufficient.

The [OpenCode skills documentation](https://opencode.ai/docs/skills/) documents
`.agents/skills` and `~/.agents/skills`, as well as its own `.opencode/skills`
and `~/.config/opencode/skills` locations. Use the bundle's existing
`--target agents` installation for the shared global skill surface.
`.github/skills` is warning-only in the matrix because it is not a documented
default discovery location. Project `AGENTS.md` is supported by the
[OpenCode rules documentation](https://opencode.ai/docs/rules/); Copilot-specific
instruction files are not assumed to load automatically.

Run both adapters in dry-run mode:

```powershell
go run ./cmd/kbcheck eval-run-live-corpus --runtime codex,ghcp --dry-run
```

Run one explicit live cross-runtime fixture:

```powershell
go run ./cmd/kbcheck eval-run-live-corpus --runtime codex,ghcp
```

The corpus runner writes `summary.json` and `summary.md` under
`.atv/eval-runs/<timestamp>-live-corpus*/`. Result statuses distinguish pass,
adapter-missing, adapter-failed, invalid-json, score-failed, and
runtime-unavailable. Live corpus runs are never part of the default `core` gate.

## Cost And Regression Report

Summarize local run artifacts:

```powershell
go run ./cmd/kbcheck skill-eval-regression --run-root .atv/eval-runs
```

Compare against a selected baseline:

```powershell
go run ./cmd/kbcheck skill-eval-regression --run-root .atv/eval-runs --baseline path\to\baseline.json
```

The report uses local cost proxies: runtime, fixture, mode, status, duration
when available, exit code, result/log sizes, pass count, and non-pass count.
Missing token or billing data is represented by absent fields, not zero.
