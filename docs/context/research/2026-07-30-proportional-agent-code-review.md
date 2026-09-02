# Proportional Agent Code Review

Checked: 2026-07-30
Budget mode: standard

## Question

What code-review pattern preserves meaningful defect detection without spawning
large reviewer swarms for routine changes?

## Findings

1. **Multi-agent review should be an escalation, not the baseline.** Anthropic
   reports that multi-agent systems use about 15 times the tokens of ordinary
   chat and work best for valuable, breadth-first tasks with genuinely
   independent directions. Anthropic also says most coding tasks have fewer
   truly parallelizable subtasks. A fixed six-reviewer minimum is therefore a
   poor default for routine code changes.
2. **One bounded generalist can own the routine review checklist.** Correctness,
   test quality, maintainability, repository standards, and agent parity all
   inspect the same diff and need much of the same context. Combining these
   lenses into one review pass avoids repeatedly loading the diff and project
   instructions. The prompt should be a compact checklist, not a concatenation
   of six full persona prompts.
3. **Specialists should be selected by risk boundaries.** Google recommends
   adding qualified reviewers for areas such as security, privacy, concurrency,
   and accessibility when the primary reviewer lacks that expertise. This
   supports conditional specialist escalation, not an always-on roster.
4. **Small changes reduce review burden more reliably than extra reviewers.**
   Google recommends one self-contained change and notes that smaller changes
   are reviewed faster and more thoroughly, with fewer missed issues. Review
   scope and change slicing should be the first cost control.
5. **Review effort should match criticality.** GitHub Copilot code review exposes
   low and medium effort levels: low for routine changes and higher-cost medium
   for security-sensitive, cross-service, or strict-quality work. This is a
   direct precedent for explicit review tiers.
6. **Deterministic work belongs outside the reviewer.** GitHub's token-efficiency
   guidance moves predictable data gathering and checks out of the model loop.
   Repository standards, formatting, tests, changed-file discovery, and
   learnings lookup should use deterministic commands where practical rather
   than dedicated reviewer agents.
7. **Optimize cost per unique actionable finding, not reviewer count.** Track
   tokens or AI credits, unique P0/P1 findings, false positives, and findings
   duplicated by another reviewer or deterministic check. A reviewer with no
   marginal finding yield should not remain always-on.
8. **ATV-Phoenix demonstrated a zero-persona evidence gate.** At inspected main
   revision `d473ddfbeab156b102436101dafc9f49a44aed4c`, `phoenix-review`
   re-runs the acceptance check, full regression suite, and tamper-evident trace,
   then reports only bugs, unmet criteria, regressions, and concrete
   security/correctness risks. It explicitly rejects style review and
   self-grading by diff inspection. This is a strong Tier 0 foundation, but it
   cannot discover semantic risks that were never encoded in the acceptance
   contract; one independent semantic reviewer still has value for non-trivial
   or risky changes.
9. **Working tests sharply reduce review uncertainty, but do not eliminate it.**
   Google's review guidance expects changes to work before review, then asks the
   reviewer to check whether the intended behavior is right, whether edge cases
   remain, whether tests would actually fail for broken code, and whether the
   design increases complexity. The review target is therefore the uncertainty
   left after proof, not a second execution of the proof.
10. **The durable value of review is broader than defect finding.** Microsoft's
    empirical study found that reviews produced fewer defect findings than
    developers expected and instead contributed understanding, alternative
    solutions, awareness, and knowledge transfer. For an agent workflow, the
    highest-value substitutes are intent coverage, test validity, and code
    health; duplicating deterministic checks is low-value.
11. **Simple single-call review should be the default escalation.** Anthropic
    recommends starting with the simplest solution and adding agentic
    complexity only when it demonstrably improves outcomes. Parallel
    perspectives are justified when independent attention materially increases
    confidence, not merely because multiple personas exist.

## Recommended Policy

| Boundary | Trigger | Agent review |
|---|---|---|
| Per slice | Every implementation slice | No reviewer agent; run the slice's coded tests and proof |
| Before slicing | Only when requirements contain material product, architecture, flow, scope, or trust uncertainty | One Spec/plan reviewer; select one specialist instead only when that specialist is clearly better qualified |
| After all slices | Non-trivial integrated code change | One generalist covering intent, proof validity, and code health |
| Post-integration escalation | Structural or high-risk integrated diff | Add at most one matching specialist, such as Thermonuclear for structural change |

Per-plan lifecycle cap: across one complete plan/manifest run, use zero reviewer
calls for mechanical/proof-complete work, one for ordinary work that skips
material pre-slice review, two for normal work with both boundaries, and three
for work that also justifies one post-integration specialist. Slices never
receive independent reviewer budgets. A specialist replaces a generic lens when
possible; it does not open another roster.

Executable proof is always required and does not consume a reviewer-agent call.
A docs-only, generated-only, or mechanically constrained change may stop there
when the evidence contract covers the change and records why semantic review is
not required.

Run review once per exact tree at the finalization or PR boundary. Reuse a
receipt while the reviewed tree and review policy fingerprint remain unchanged.
Do not run duplicate semantic-review entry points on the same tree.

## Sources

- Anthropic, "How we built our multi-agent research system": multi-agent systems
  use about 15 times the tokens of chat, fit breadth-first parallel work, and are
  a weaker fit for most coding tasks:
  https://www.anthropic.com/engineering/multi-agent-research-system
- Google Engineering Practices, "What to look for in a code review": use
  qualified additional reviewers for specialized areas rather than treating
  every reviewer as responsible for every lens:
  https://google.github.io/eng-practices/review/reviewer/looking-for.html
- Google Engineering Practices, "Small CLs": one self-contained change is
  faster and easier to review thoroughly:
  https://google.github.io/eng-practices/review/developer/small-cls.html
- Google Engineering Practices, "Speed of Code Reviews": large changes should
  usually be split rather than compensated for with slower review:
  https://google.github.io/eng-practices/review/reviewer/speed.html
- GitHub Docs, "About GitHub Copilot code review": low effort is the routine
  default; higher-cost medium effort targets complex and security-sensitive
  changes:
  https://docs.github.com/en/copilot/concepts/agents/code-review
- GitHub, "Improving token efficiency in GitHub Agentic Workflows": instrument
  token usage, move deterministic retrieval outside the model loop, and judge
  savings alongside output quality:
  https://github.blog/ai-and-ml/github-copilot/improving-token-efficiency-in-github-agentic-workflows/
- Google Research, "Modern Code Review: A Case Study at Google": exploratory
  evidence from 12 interviews, 44 survey respondents, and nine million reviewed
  changes:
  https://research.google/pubs/modern-code-review-a-case-study-at-google/
- Microsoft Research, "Expectations, Outcomes, and Challenges of Modern Code
  Review": review's observed value included understanding, awareness, and
  alternative solutions beyond expected defect finding:
  https://www.microsoft.com/en-us/research/publication/expectations-outcomes-and-challenges-of-modern-code-review/
- Anthropic, "Building effective agents": start with the simplest solution and
  add parallel/evaluator complexity only when it demonstrably improves outcomes:
  https://www.anthropic.com/engineering/building-effective-agents
- ATV-Phoenix, `phoenix-review`: evidence-based acceptance, regression, and
  trace review with no reviewer-persona swarm:
  https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/skills/phoenix-review/SKILL.md
- ATV-Phoenix, long-horizon design research: escalating verifier grades and
  objective environment feedback:
  https://github.com/All-The-Vibes/ATV-Phoenix/blob/main/research/long-horizon-agent-design.md

## Applies When

- Designing `kb-review` or completion review gates.
- Deciding whether a reviewer persona should be always-on or conditional.
- Evaluating review cost, duplicate findings, or reviewer marginal yield.

## Stale When

- Runtime pricing or review-effort controls materially change.
- Repository evals show that a different reviewer count has better unique
  high-severity finding yield at comparable cost.

## Rejected Approaches

- **Six or more always-on personas:** repeatedly loads overlapping context and
  pays multi-agent coordination cost before risk is classified.
- **One giant concatenated persona prompt:** removes agent fan-out but retains
  prompt bloat and competing instructions. Use one compact generalist rubric.
- **No LLM review at all:** deterministic checks cannot reliably judge design,
  intent mismatch, subtle edge cases, or inappropriate complexity.
- **Treat Phoenix proof as complete semantic review:** it proves the checks and
  trace, not that the acceptance contract covered every important risk.
- **Run every specialist on large diffs:** size alone does not make every domain
  relevant. Select specialists by changed risk boundaries.

## Current Adoption

The recommendation is implemented: deterministic proof runs first;
`kb-finalize` selects zero or one semantic review; `kb-review` chooses one broad
or replacement specialist profile; and `document-review` invokes one best-fit
specialist only when material uncertainty remains. Phoenix remains cited as
historical prior art for an evidence-first gate, not as a runtime dependency.
