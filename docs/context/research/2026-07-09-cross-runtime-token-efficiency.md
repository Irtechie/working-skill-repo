# Cross-Runtime Token Efficiency

Checked: 2026-07-09
Budget mode: standard

## Question

Which high-impact token-saving strategies transfer across Codex and GitHub
Copilot CLI without adding a required provider, daemon, MCP server, or
runtime-specific hook?

## Findings

### Adopt now

1. **Deduplicate ambient instructions.** GitHub Copilot CLI loads both
   `AGENTS.md` and `.github/copilot-instructions.md`. Repeating the repo contract
   in both files pays on every turn and destabilizes the cached prefix. Keep the
   full contract in `AGENTS.md`; keep the Copilot file as a thin pointer.
2. **Keep volatile state out of the prompt prefix.** Stable policy belongs in
   ambient instructions. Current status belongs in `todo.md`, manifests, and
   handoffs loaded only when relevant. OpenAI prompt caching requires exact
   prefix matches, so frequently edited task state should not sit in always-on
   instructions.
3. **Move deterministic reads outside the reasoning loop.** Use `git`, `gh`,
   `rg`, and repo-native commands to prefetch metadata, diffs, file lists, and
   test summaries. Pass compact results or artifact paths to the agent instead
   of spending a model turn deciding how to fetch known inputs.
4. **Minimize tool registration.** Unused MCP tools add schemas to every model
   request. Do not ship repo-local MCP configuration; prefer built-in
   file/search/CLI tools and enable optional tools only for the task that needs
   them. Host-native deferred tool search is useful when available, but the
   portable bundle must not depend on it.
5. **Use isolated narrow agents for noisy work.** GHCP's task/explore/general
   agents explicitly keep command output or exploration out of the main
   context. The portable contract is one bounded task, one compact context
   packet, and one result; host-specific delegation remains an adapter.

### Already present

- Fresh-session recovery through `kb-map`, handoffs, and repo-local memory.
- Lazy skill references and loaded-surface/token-estimate reports.
- Compact `kbcheck` success output with detailed failures.
- Model-tier routing with a constant proof bar.
- Session-hygiene thresholds and `kb-compact`.
- A planned context-packet/task-state absorption spike.

### Highest-value remaining gap

The live Codex/GHCP eval adapters measure duration and artifact bytes, not real
token use. Add a normalized optional usage record when the host exposes it:

```text
runtime, model, turns, input_tokens, output_tokens,
cache_read_tokens, cache_write_tokens, task_or_fixture, proof_result
```

Keep raw fields as the source of truth. A weighted "effective tokens" score may
be useful for trend reports, but its provider/model weights must be versioned;
it must never replace correctness and proof outcomes.

The next implementation should also record:

- tokens per passing task or fixture, not tokens per run alone;
- turns per task and tokens per turn;
- cache-read ratio and cache-write cost where available;
- right-to-wrong regressions and proof failures beside savings;
- run frequency, because a small saving on a frequent workflow can dominate.

## Sources

- GitHub Copilot CLI docs: custom instructions load `AGENTS.md` and
  `.github/copilot-instructions.md`; built-in task/explore/general agents use
  separate contexts and compact results:
  https://docs.github.com/en/copilot/how-tos/use-copilot-agents/use-copilot-cli
- GitHub, "Improving token efficiency in GitHub Agentic Workflows": unused MCP
  schemas, deterministic CLI prefetch, normalized usage artifacts, and
  quality-aware cost metrics:
  https://github.blog/ai-and-ml/github-copilot/improving-token-efficiency-in-github-agentic-workflows/
- VS Code, "Improving token efficiency for GitHub Copilot in VS Code": prompt
  prefix caching, deferred tool definitions, and measured tool-search savings:
  https://code.visualstudio.com/blogs/2026/06/17/improving-token-efficiency-in-github-copilot
- OpenAI Prompt Caching: exact stable prefixes, cache reads/writes, keys, and
  breakpoints:
  https://developers.openai.com/api/docs/guides/prompt-caching

## Applies When

- Editing ambient instructions, skill hot paths, adapters, or live evals.
- Adding MCP servers or large tool catalogs.
- Designing context packets, subagent boundaries, or recurring workflows.

## Stale When

- Codex or GHCP exposes a stable cross-runtime usage export that changes the
  normalized field set.
- Host-native tool search or prompt-cache behavior materially changes.

## Rejected Approaches

- Make CCE or another context-engine/MCP server required: optional lookup
  acceleration can help, but mandatory startup adds non-portable schema and
  failure overhead.
- Optimize raw token count without correctness/proof outcomes: can reward doing
  less work or using a weaker model incorrectly.
- Put live task state in `AGENTS.md`: increases every turn and breaks stable
  prompt prefixes.

## Impact On Current Project

- `.github/copilot-instructions.md` is now a thin pointer to `AGENTS.md`.
- Repo-local provider auto-start files were removed; CCE remains an owned,
  supported opt-in adapter.
- Phoenix remains credited prior art, but no Phoenix runtime/MCP connection is
  part of the bundle.
- The planner-economy telemetry slice should capture normalized real usage when
  available, while remaining valid when a host exposes no usage fields.
