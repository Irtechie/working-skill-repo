# Session Model Discovery and Routing Surfaces

Checked: 2026-07-09
Budget mode: standard

## Question

How can KB discover models the current host can actually use, add private local
or custom routes without a setup questionnaire, and choose conservative
model-backed subagents when `kb-work` dispatches a plan slice?

## Findings

1. Plans should remain provider-neutral. KB already assigns task tiers and
   bounded context packets; live model availability belongs to work-time
   dispatch, not the manifest.
2. Codex CLI has a real discovery surface. On `codex-cli 0.143.0-alpha.9`,
   `codex debug models` prints the refreshed catalog Codex sees. The visible
   catalog inspected here contains GPT-5.5, GPT-5.4, GPT-5.4-Mini, and
   GPT-5.3-Codex-Spark; the separately inspected bundled catalog differs.
3. Codex catalogs are surface-specific. This Codex App surface also exposes
   Sol, Terra, Luna, and GPT-5.5. Do not treat the CLI catalog as the whole Codex
   inventory or hardcode any one surface's names globally; merge the active
   surface's catalog and callable-agent schema.
4. Codex project custom agents under `.codex/agents/` may override model,
   reasoning effort, provider, tools, sandbox, MCP, and skills. However, the
   generic `spawn_agent` surface exposed in this task has no per-call model or
   agent-type field. An adapter must use a proven named-agent or explicit model
   selector rather than infer that a generic spawn used the requested model.
5. GHCP 1.0.70 exposes `--model`, `/subagents`, per-subagent model/effort/context
   configuration, and OpenAI-compatible providers. The exact supported command
   for enumerating the current entitled catalog still needs fixture proof.
6. LLMCommune's supported client paths are TinyBoss controller
   `http://192.168.1.205:4100` and LiteLLM
   `http://192.168.1.205:4000/v1`. Fleet MCP V1 exposes discovery and execution
   for fleet capabilities/jobs; the current docs do not establish it as a
   general LLM chat/completions transport. Local model inference should use the
   supported LiteLLM route unless a later versioned MCP capability proves direct
   LLM dispatch.
7. LLMCommune already has useful capability evidence: its controller catalog
   records exact model names, engines, context results, and recommendations for
   Qwen, Gemma, and other local routes. KB can consume a bounded exported view
   through a versioned adapter instead of duplicating fleet topology.
8. Native availability should be ephemeral and rediscovered once per work
   session. `~/.kb/models.json` can persist private endpoints, control/inference
   routes, and declared capability hints. Optional project policy may reference
   aliases or preferences but should not copy machine-specific details.
9. Conservative selection should use the plan tier as a floor, allow stronger
   models to do simpler work, bias upward under uncertainty, and preserve the
   context packet plus failing proof during escalation. Model choice is economy
   policy; work proof remains the correctness gate.
10. Provider/family metadata is useful beyond size. An orchestrator may prefer a
    same-family worker for continuity or a different provider/family for an
    independent review, while remaining inside project trust policy.

## Recommended Boundary

```text
host-native surface catalog
  + ~/.kb/models.json global extras/connections
  + optional project kb-models.json alias policy/preferences
  + one-run user override
                |
                v
      ephemeral session catalog
                |
                v
       kb-work live selection
                |
                v
       model-backed subagent
```

- `kb-goal`: discover and report; do not force a model questionnaire.
- `kb-plan`: record task tier/risk/context/proof; do not name a model.
- `kb-work`: choose, preview, call, fall back within class, and escalate upward.
- `kb-models`: add or calibrate non-native global routes and optional project
  preferences.
- `kb-complete`: verify and land proven work; routing status is evidence, not a
  correctness oracle.

## Sources

- OpenAI Codex manual and model/subagent references:
  https://developers.openai.com/codex/codex-manual.md
- OpenAI Codex subagents:
  https://developers.openai.com/codex/subagents
- GitHub Copilot CLI documentation:
  https://docs.github.com/copilot/how-tos/copilot-cli
- Local commands: `codex --version`, `codex debug models --help`,
  `codex debug models --bundled`, `copilot --version`, `copilot help config`, and
  `copilot help providers`.
- LLMCommune fleet runbook:
  `<llmcommune-repo>/docs/context/operations/fleet-connection-runbook.md`.
- LLMCommune Fleet MCP:
  `<llmcommune-repo>/bootstrap/fleet/apps/fleet-mcp/README.md`.
- LLMCommune model capability catalog:
  `<llmcommune-repo>/bootstrap/serving/controller/model-capabilities.json`.

## Applies When

- Routing plan slices to subagents across Codex, GHCP, or local providers.
- Adding private GPU/fleet models without hardcoding them into a project plan.
- Choosing same-family implementation workers or cross-family reviewers.

## Stale When

- Codex or GHCP changes its catalog/subagent configuration surfaces.
- Fleet MCP adds a documented general LLM invocation capability.
- TinyBoss controller/LiteLLM client boundaries change.
- KB replaces task tiers or bounded context packets.

## Impact On Current Project

Replace the per-project model questionnaire with session discovery. Keep
machine-specific extra routes global/user-local, let projects reference aliases
without copying endpoints, use one Go core for normalization and dispatch, and
move concrete model choice entirely into `kb-work`.
