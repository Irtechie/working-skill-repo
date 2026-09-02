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
### Historical host snapshot

The following host details were observed on the checked date and are not the
current product contract:

2. Codex CLI had a real discovery surface. On `codex-cli 0.143.0-alpha.9`,
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
5. GHCP 1.0.70 exposed `--model`, `/subagents`, per-subagent model/effort/context
   configuration, and OpenAI-compatible providers. The exact supported command
   for enumerating the current entitled catalog still needs fixture proof.
6. Private controller and LiteLLM deployments informed the adapter boundary,
   but their addresses, topology, and machine-local catalogs are not public KB
   dependencies. Fleet capability/job discovery did not establish a generic
   chat-completions transport.
7. A private capability catalog demonstrated that bounded exported capability
   evidence can inform routing without copying fleet topology into KB.
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
  + canonical user-local route state
  + optional user-local project preference
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
- Private controller, Fleet MCP, and model-capability documents inspected on the
  checked date; these machine-local sources are intentionally not public links.

## Applies When

- Routing plan slices to subagents across Codex, GHCP, or local providers.
- Adding private GPU/fleet models without hardcoding them into a project plan.
- Choosing same-family implementation workers or cross-family reviewers.

## Stale When

- Codex or GHCP changes its catalog/subagent configuration surfaces.
- A versioned capability transport adds a documented general LLM invocation
  contract.
- User-local route or adapter boundaries change.
- KB replaces task tiers or bounded context packets.

## Current Adoption

Plans remain provider-neutral and concrete selection happens at execution time.
`kbrouter` keeps optional route definitions, approval state, and source
preference in canonical user-local state, separates host-native targets from
CLI/local routes, and treats model selection as routing evidence rather than
correctness proof. Current commands and safety policy live in
`docs/context/architecture/kbrouter.md`.
