# Low-Cognitive-Burden Agent Communication

Checked: 2026-07-26
Budget mode: standard

## Question

How should an agent combine plain-language and action-first response patterns
without confusing brevity with usability or making the reader determine which
questions genuinely require a human answer?

## Findings

1. Plannotator's `bro` skill is a manual repair command: restate the last
   message plainly, without jargon, metaphors, or fluff. It improves wording
   but does not classify responsibility or decide whether the user must reply.
2. AYGHRI's `i-have-adhd` skill contributes useful action-first structure,
   numbered steps, visible state, tangent suppression, and matter-of-fact
   errors. Its mandatory next actions, time estimates, state repetition, and
   hard list cap are unsafe as universal defaults because they can manufacture
   user work or hide necessary evidence.
3. KB already classifies unknowns as `ask-now`, `research-first`,
   `safe-assumption`, `defer-to-planning`, or `parked`. The missing layer is a
   plain-language presentation contract that tells the reader whether a reply
   is required and why.
4. HumanLayer's workflow puts human attention on research, design, and plans
   because mistakes there cascade into much larger implementation mistakes.
   Its stated goal for code review is mental alignment: help reviewers
   understand how the system is changing and why, rather than forcing them to
   reconstruct intent from a large diff.
5. HumanLayer treats a reviewed design document as an interface to the work:
   comments and decisions flow back into implementation. In a portable
   file-native workflow, a companion document can provide that same stable
   review surface when the PR body links it, names unresolved decisions, and is
   updated when those decisions are resolved.

Use this responsibility test before asking:

| Class | Meaning | User-facing treatment |
|---|---|---|
| Must respond | Only the user can authorize, supply, or decide it, and the dependent path cannot continue | Say `Need from you - hard stop`, ask exactly one question, explain why the agent cannot decide, state the consequence, and recommend an option when possible |
| May respond | The user may have a preference, but the agent can make a safe reversible choice | State the recommended default and say the agent can handle it; do not block unless the user overrides |
| No response needed | Status, proof, completed work, or a decision the agent already owns | Lead with the outcome; explicitly say `No response needed` only when the message could otherwise look like an assignment |

Preferred response order:

1. Bottom line or outcome.
2. Human action, only when one exists.
3. Why it matters and why the responsibility belongs to the human or agent.
4. Recommendation or agent-owned default.
5. Optional proof and technical depth under `Details` or another descriptive
   heading.

For pull requests and review companions, use the same responsibility test:

1. `What changed` and `Why it matters` establish mental alignment.
2. `Needs reviewer attention` lists only decisions, risks, or claims a human
   should inspect; say why each item matters.
3. `Handled by the agent` names completed mechanical work and proof so the
   reviewer does not spend attention rediscovering it.
4. `Verification` gives compact results and links verbose evidence by path.
5. `Risks / deferred` distinguishes a real merge blocker from optional
   follow-up.

Keep the PR body as the low-burden first screen. Link a companion research,
design, plan, or proof document for depth instead of pasting the entire
tactical history into the PR.

## Sources

- [plannotator/dev-skills `bro`](https://github.com/plannotator/dev-skills/blob/main/skills/general/bro/SKILL.md)
- [ayghri/i-have-adhd](https://github.com/ayghri/i-have-adhd/blob/main/skills/i-have-adhd/SKILL.md)
- [HumanLayer: Advanced Context Engineering for Coding Agents](https://www.humanlayer.dev/blog/advanced-context-engineering)
- [HumanLayer: Writing a good CLAUDE.md](https://www.humanlayer.dev/blog/writing-a-good-claude-md)
- [HumanLayer: comment-driven design reviews](https://www.humanlayer.dev/)
- [HumanLayer: Context-Efficient Backpressure](https://www.humanlayer.dev/blog/context-efficient-backpressure)

## Applies When

- Writing status, completion, blocker, approval, clarification, or decision
  messages.
- Compressing a response without losing proof or safety context.
- Translating internal workflow classifications into a human-readable ask.
- Preparing PR descriptions and companion documents for high-leverage human
  review.

## Stale When

- The upstream skills materially change their stated contracts.
- KB replaces the Question Gate classes or the global response policy.
- Runtime-specific question tools gain a native hard/soft/no-response
  responsibility field.

## Rejected Approaches

- Universal terseness: word count does not measure cognitive burden.
- A hard five-item limit: proof, risk, or safety information may require more.
- Mandatory time estimates: ungrounded estimates create false certainty.
- Mandatory next actions: completed work should not manufacture user work.
- A second always-loaded style skill: the ambient rule belongs in `AGENTS.md`;
  `kb-cognitive` remains the explicit repair/organization lane.

## Current Adoption

`AGENTS.md` and `kb-cognitive` now optimize for comprehension and responsibility
rather than raw terseness. The workflow distinguishes hard human questions,
soft preferences, and agent-owned information; PR and completion surfaces lead
with the outcome and required attention. Deterministic contract coverage
protects the responsibility classes.
