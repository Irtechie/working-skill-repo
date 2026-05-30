---
date: 2026-05-30
topic: live-cross-runtime-skill-eval-harness
brainstorm_style: kb-brainstorm
---

# Live Cross-Runtime Skill Eval Harness

## Problem Frame

The skill repo now has deterministic skill lint, route-complexity fixtures,
captured-result scoring, sync drift checks, and a working Codex live adapter.
That proves the harness shape, but it still does not prove that Codex and GitHub
Copilot/GHCP consistently route real prompts, preserve proof discipline, avoid
hallucinated claims, and produce maintainable outputs across the full fixture
set.

The affected users are the repo maintainer and future agents entering consuming
projects. The core failure to prevent is false confidence: a skill change passes
static checks while live agents regress route choice, skip required workflow
reads, claim work that did not happen, or burn high cost for low-quality output.

## Research Summary

**Findings that shaped requirements:**
- GitHub Copilot CLI supports non-interactive programmatic prompts with
  `copilot -p`, silent output with `-s`, `--no-ask-user`, explicit model
  selection, allowed tool/path controls, and transcript export with `--share`.
  Affected: R1, R2, R4, R8. Source:
  https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference
- GitHub Copilot CLI documentation recommends precise prompts, minimal
  permissions, `-s` for clean captured output, `--no-ask-user` for automation,
  and explicit models for consistency. Affected: R2, R3, R4, R11. Source:
  https://docs.github.com/en/copilot/how-tos/copilot-cli/automate-copilot-cli/run-cli-programmatically
- GitHub documents that agent skills work with Copilot cloud agent, Copilot CLI,
  and agent mode in Visual Studio Code, and that project skills live under
  `.github/skills`. Affected: R1, R5, R15. Source:
  https://docs.github.com/en/copilot/how-tos/copilot-on-github/customize-copilot/customize-cloud-agent/add-skills
- GitHub Copilot CLI loads repository instructions from
  `.github/copilot-instructions.md`, `AGENTS.md`, and path-specific
  `.github/instructions/**/*.instructions.md`; conflicting instructions may be
  non-deterministic. Affected: R5, R15. Source:
  https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions
- Local inspection shows `copilot` is installed and supports `-p`,
  `--output-format json`, `--allow-tool`, `--no-ask-user`, `--share`,
  `--model`, and `--agent`, but it does not expose a Codex-style
  `--output-schema` option in `copilot help`. Affected: R2, R6. Source:
  local command `copilot help`.

**Confidence:** High for Codex/local repo facts and Copilot CLI surface; medium
for GHCP output-shape reliability because strict schema enforcement is not
available from the observed CLI surface.

## Requirements

**Eval Layer Contract**

| Layer | Judge | Default Gate | Purpose |
|---|---|---|---|
| Static structure | deterministic scripts | `kb-check -All` | Catch broken skills/config/docs before runtime |
| Adapter plumbing | deterministic dry-runs | `kb-check -All` | Prove wrappers emit scorer-compatible JSON without model calls |
| Live route corpus | model action plus deterministic scoring | explicit command | Prove Codex/GHCP route real prompts correctly |
| Trace and claims | deterministic checks after extraction | explicit command, later CI candidate | Catch skipped workflow reads and false final claims |
| Output quality | rubric, possibly LLM-assisted | explicit command | Catch ugly or over-ceremonial successful runs |
| Cost/regression | deterministic metrics | explicit report | Track spend, time, retries, and pass-rate movement |

**Runtime Adapters**
- R1. Add a GHCP live adapter that runs route fixtures through GitHub Copilot CLI
  and emits the same `evals/skill-eval` result shape used by
  `scripts/skill-eval.ps1`.
- R2. The GHCP adapter must run non-interactively with no user questions, minimal
  permissions, captured stdout/stderr, and a transcript artifact when available.
- R3. The GHCP adapter must fail hard on invalid or missing result JSON. It must
  not silently repair, infer, or downgrade malformed model output into a pass.
- R4. Runtime adapters must support `-FixtureId`, `-All`, `-DryRun`, `-KeepRun`,
  `-Model`, `-Json`, and configurable command paths where the runtime supports
  them.
- R5. Runtime adapters must run in disposable eval workspaces and must not edit
  the source worktree during route-selection evals.
- R6. Codex may use `--output-schema`; GHCP must use prompt-level JSON
  constraints plus deterministic parsing because no local GHCP schema flag is
  exposed.

**Live Corpus**
- R7. The live corpus must run all existing route-complexity fixtures for Codex
  and GHCP, not only `tiny-typo-fix`.
- R8. Each live run must record runtime, model when known, fixture id, route,
  expected route, expected proof, pass/fail, run duration, exit code, result
  path, and transcript/log paths.
- R9. The corpus runner must distinguish adapter plumbing failures, invalid JSON,
  deterministic score failures, and runtime/auth unavailable skips.
- R10. `kb-check -All` must keep live model calls out of the default deterministic
  gate, but must include dry-run checks for every live adapter.
- R10a. Runtime/auth unavailable skips must be explicit result states in corpus
  summaries, not silent passes and not generic failures.

**Trace And Claim Verification**
- R11. Extend result scoring to check forbidden shortcuts and required workflow
  reads where fixtures or skills require them.
- R12. Add transcript-derived claim extraction that can compare final claims
  against actual files, git state, commands, logs, and produced artifacts.
- R13. Claim verification must be deterministic after extraction. LLM assistance
  may propose candidate claims, but filesystem/git/log checks decide pass/fail.
- R14. The scorer must preserve structured claim checks as the stable contract
  while allowing transcript-derived checks to add failures.
- R14a. Transcript-derived claim extraction must include self-test fixtures with
  at least one true claim, one false claim, and one ambiguous claim that is
  reported without being treated as proof.

**Instruction And Skill Coverage**
- R15. Add fixtures or checks that exercise the instruction surfaces GHCP actually
  uses: `AGENTS.md`, `.github/copilot-instructions.md`,
  `.github/instructions/**/*.instructions.md`, `.github/skills/**/SKILL.md`, and
  global Copilot/shared-agent skill directories where applicable.
- R16. Add at least one fixture that detects cross-runtime instruction drift:
  Codex and GHCP should choose the same KB lane for the same task unless the
  fixture explicitly documents a supported runtime difference.
- R17. The compatibility matrix in `config/skill-quality.json` must distinguish
  supported, simulated, warning-only, and unsupported live eval modes.

**Output Quality Rubric**
- R18. Add a quality scoring layer for completeness, maintainability, relevance,
  proof quality, and unnecessary ceremony.
- R19. Quality scoring must not replace deterministic route/proof/claim checks.
  It is a separate rubric that can fail or warn after deterministic scoring.
- R20. The rubric must catch both bad passes and expensive overkill: correct route
  with poor proof, correct proof with unmaintainable output, and tiny work that
  burns large-session ceremony.
- R20a. Quality rubric output must record whether a score is deterministic,
  LLM-judged, or human-only so reports do not blur subjective judgment into
  machine proof.

**Cost And Regression Tracking**
- R21. Each live eval run must capture available cost proxies: wall-clock time,
  model name, tool call count when parseable, retry count, prompt/result sizes,
  and transcript/log sizes.
- R22. Add a local regression summary that compares current live runs against the
  last checked-in or selected baseline without requiring Langfuse, Braintrust, or
  LangSmith.
- R23. Dashboard/export integrations may be added later, but local JSON/Markdown
  artifacts remain the source of truth.

**Eval Map Scaffolding Safety**
- R24. Add scaffold negative-check validation for future consuming-repo eval maps:
  generated smoke evals must fail when their expected selector, status, output,
  schema, checksum, or command expectation is intentionally broken.
- R25. The negative-check validation must revert the intentional break and record
  both the passing command and failed-as-expected negative command in
  `docs/context/eval-map.md` or testing docs.

## Success Criteria

- `scripts/skill-eval-run-ghcp.ps1 -FixtureId tiny-typo-fix` can produce a
  scorer-compatible result or a clear runtime/auth unavailable skip.
- Dry-run checks for Codex and GHCP live adapters run under `kb-check -All`.
- A live corpus command can run all eight route fixtures for Codex and GHCP and
  produce a summarized pass/fail report.
- The scorer can fail at least one transcript-derived hallucinated claim case.
- The scorer or rubric can flag at least one output-quality failure independent
  of route correctness.
- Cost/regression output records enough data to answer whether a skill change
  improved, regressed, or merely shifted spend.
- Docs clearly state which evals are deterministic, live-model, LLM-judged,
  exporter-only, or skipped.

## Scope Boundaries

- Do not make live model calls part of the default `kb-check -All` gate.
- Do not require Langfuse, Braintrust, LangSmith, Promptfoo, or DeepEval before
  the local harness is stable.
- Do not pretend GHCP has schema enforcement unless the local CLI or official
  docs expose it.
- Do not run cloud-agent PR workflows as the first GHCP adapter; start with local
  Copilot CLI because it is installed and scriptable.
- Do not score subjective product taste as deterministic truth. Keep rubric
  output separate from machine-verifiable proof.
- Do not create consuming-project memory or handoffs in this repo.

## Key Decisions

- Build GHCP local CLI parity before cloud-agent integration. Rationale: the
  local CLI is installed, scriptable, and documented for programmatic prompts.
  Evidence: local `copilot help` and GitHub Copilot CLI docs.
- Keep live model calls explicit. Rationale: live runs cost tokens, can be flaky,
  and require runtime auth; deterministic dry-runs are enough for default repo
  hygiene. Evidence: current `kb-check -All` design and user tolerance for
  expensive bootstrap-style work only when intentional.
- Use local artifacts as the judge of record. Rationale: dashboards are useful
  exporters, but the skill repo must work for non-LangChain apps and ordinary
  websites/CLIs/APIs. Evidence: prior user direction and `kb-eval-map` pattern
  matrix.
- Treat schema absence as a GHCP constraint, not a reason to fake success.
  Rationale: local `copilot help` exposes JSONL output but no output-schema flag.
  Evidence: local command inspection.

## Dependencies / Assumptions

- The user has an authenticated Copilot CLI session when live GHCP runs are
  attempted.
- PowerShell remains acceptable for first-pass adapters because the existing
  harness is PowerShell and Windows-oriented.
- The existing route fixture corpus remains the first live corpus. New fixtures
  can be added after the full eight-fixture cross-runtime pass exists.
- Some cost fields may be proxies until runtime logs expose richer token data.

## Alternatives Considered

- Use only Braintrust/LangSmith/Langfuse: rejected because the user needs evals
  for non-Lang apps and local repo-native proof.
- Build cloud GHCP agent evals first: deferred because local Copilot CLI has a
  simpler non-interactive path and avoids GitHub PR/session orchestration.
- Force every runtime through one abstraction immediately: rejected because Codex
  and GHCP have materially different output controls.
- Make LLM judges decide route correctness: rejected because expected route,
  proof strings, trace checks, and claims can be scored deterministically.

## Slice Candidates

- GHCP live adapter - runs one route fixture through `copilot -p`, captures logs
  and transcript, parses strict JSON, and scores with `skill-eval`.
- Adapter dry-run integration - adds GHCP adapter dry-run to `kb-check -All` and
  updates docs/config runtime modes.
- Live corpus runner - runs all route fixtures across Codex and GHCP and emits a
  summarized report.
- Trace rule expansion - adds forbidden-shortcut and required-read checks to
  deterministic scoring.
- Transcript claim verifier - derives or reads final claims and checks them
  against files, git state, commands, logs, and artifacts.
- Output quality rubric - scores completeness, maintainability, relevance,
  proof quality, and unnecessary ceremony separately from deterministic proof.
- Cost and regression report - records runtime/model/time/tool/retry/prompt-size
  data and compares current runs to a selected baseline.
- Scaffold negative-validation support - makes future `kb-eval-map` scaffolds
  prove their smoke evals fail when intentionally broken.

## Outstanding Questions

### Resolve Before Planning

None.

### Deferred to Planning

- [Affects R1-R6][Technical] Decide whether GHCP adapter should call the
  PowerShell shim directly or bypass it through the underlying Node loader, as
  the Codex adapter does.
- [Affects R12-R14][Technical] Decide whether transcript-derived claim extraction
  starts with deterministic regex patterns, an LLM candidate extractor, or both.
- [Affects R18-R20][Technical] Decide the initial rubric scale and fail/warn
  thresholds.
- [Affects R21-R23][Technical] Decide baseline storage path and retention policy
  for live run summaries.

## Next Steps

-> /kb-plan
