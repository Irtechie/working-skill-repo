# Current Agent Skill Guidance

Checked: 2026-07-30
Budget mode: deep

## Question

What current Anthropic and OpenAI guidance should govern a portable repository
of coding-agent skills, and which older patterns should be retired?

## Findings

1. **Start with one focused agent.** Both vendors now recommend the smallest
   agent that can own a clear task. Add another agent only for independent,
   disposable context, materially different tools or permissions, or a distinct
   ownership boundary.
2. **Multi-agent coding is an escalation.** Anthropic reports roughly 15x chat
   token use for its multi-agent research system and explicitly says most coding
   tasks are less parallelizable than research.
3. **Context engineering supersedes prompt accumulation.** Load the smallest
   high-signal context just in time. Skill metadata is always visible, the skill
   body loads on activation, and references load only when needed.
4. **Skill bodies should remain compact.** Anthropic recommends keeping
   `SKILL.md` below roughly 500 lines, keeping references one level deep, and
   adding a table of contents to references over 100 lines.
5. **Deterministic proof precedes semantic judgment.** Tests, builds, lints,
   contract checks, and trace grading should establish observable behavior.
   Semantic review should target uncertainty those checks cannot settle.
6. **A separate evaluator is valuable when stakes justify it.** Evaluator loops
   need explicit criteria and demonstrable improvement; they are not generic
   ceremony. Anthropic found one judge call most consistent with human judgment
   in its research system.
7. **Long-running work needs objective stop conditions.** Skills must define a
   verifiable success condition, bounded retries, action-sensitive approval
   thresholds, and explicit failure escalation.
8. **Delegation needs a contract.** Every delegated task needs an objective,
   output shape, tools or sources, boundaries, and clear ownership. Split only
   when the branch truly needs different instructions, tools, or policy.
9. **Durable memory must be curated.** Persistent repository memory and
   disposable working notes are different systems. Automatic learning should
   not continuously inflate always-loaded context.
10. **Trace first, then formalize evals.** Use observed traces to understand
    routing and tool failures, then turn stable expectations into repeatable
    datasets and deterministic gates.

## Sources

- Anthropic, "Building effective agents":
  https://www.anthropic.com/engineering/building-effective-agents
- Anthropic, "How we built our multi-agent research system":
  https://www.anthropic.com/engineering/multi-agent-research-system
- Anthropic, "Effective context engineering for AI agents":
  https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents
- Anthropic, "Writing tools for agents":
  https://www.anthropic.com/engineering/writing-tools-for-agents
- Anthropic, Agent Skills overview and best practices:
  https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview
  and
  https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
- Anthropic, Claude Code best practices and subagents:
  https://code.claude.com/docs/en/best-practices and
  https://code.claude.com/docs/en/sub-agents
- OpenAI, Agents SDK agent definitions and orchestration:
  https://developers.openai.com/api/docs/guides/agents/define-agents and
  https://developers.openai.com/api/docs/guides/agents/orchestration
- OpenAI, Agents SDK running agents and guardrails:
  https://developers.openai.com/api/docs/guides/agents/running-agents and
  https://developers.openai.com/api/docs/guides/agents/guardrails-approvals
- OpenAI, agent evals:
  https://developers.openai.com/api/docs/guides/agent-evals
- OpenAI, GPT-5 prompting guide:
  https://developers.openai.com/cookbook/examples/gpt-5/gpt-5_prompting_guide
- OpenAI, Agents SDK session memory:
  https://developers.openai.com/cookbook/examples/agents_sdk/session_memory

## Applies When

- Authoring or trimming `SKILL.md` files.
- Deciding whether to split work across agents.
- Designing review, finalization, memory, retry, and completion loops.
- Removing compatibility aliases or dead workflow lanes.
- Adding deterministic skill-quality checks.

## Stale When

- Anthropic changes Agent Skills structure or line/reference guidance.
- OpenAI replaces the current agent orchestration, eval, or approval guidance.
- Repo evals demonstrate that a more complex workflow has better measured
  outcomes at acceptable cost.

## Rejected Approaches

- Treating every persona as cheap because it runs in parallel.
- Preserving aliases indefinitely without a caller or distinct contract.
- Encoding every edge case in the hot-path skill body.
- Letting the implementing agent self-certify completion without executable
  proof.
- Running compound, learning, evolution, and memory refresh after every task.

## Current Adoption

The bundle now uses progressive disclosure, single-agent-first execution,
zero-or-one proportional semantic review, bounded retries, and deterministic
structural checks. Removed aliases are recorded as history rather than exposed
as routes. `kbcheck` owns objective contracts while bounded reviewers retain
semantic judgment. Current inventory and workflow behavior live in the
architecture docs.
