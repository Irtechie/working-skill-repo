# Low-Burden Review Artifacts

Use this structure for pull requests and companion design, research, or plan
documents. The goal is not fewer words. The goal is less reviewer effort with
the important decision impossible to miss.

## Review Responsibility

Classify every item before presenting it:

| Class | Put it where | Reviewer action |
|---|---|---|
| Hard response required | `Needs reviewer attention` | Decide, authorize, or supply the named input |
| Soft preference | `Handled by agent`, with the reversible default | Optional override; do not block |
| No response needed | `Handled by agent` or `Verification` | Read only if useful |

An item belongs in `Needs reviewer attention` only when the reviewer genuinely
owns it and dependent work cannot safely continue without the answer. State the
exact ask, why the reviewer owns it, what blocks, and the recommendation.

## Pull Request First Screen

Use these headings in this order:

```markdown
## What changed / Why it matters
<Outcome and user or system effect.>

## Needs reviewer attention
<Exact hard questions or "None — no reviewer decision required.">

## Handled by agent
<Safe choices, routine fixes, and soft preferences already handled.>

## Verification
<Commands, checks, and visible proof. Put compact success here; link full logs
for failures or material caveats.>

## Risks / deferred
<Remaining risk, quarantined failures, scope exclusions, and follow-up owners.>
```

The PR is an executive first screen, not a chronological work diary. Lead with
the outcome and the reviewer-owned decision. Keep exact evidence, blockers,
risks, and safety information.

### Presentation Ladder

Keep text as the default. Use a table for repeated fields, one real screenshot
for a small rendered UI change, and Mermaid for a compact sequence or
relationship. Use `interactive-workflow-workbench-light` only when one bounded
path benefits from a few selectable nodes. Reserve the full
`interactive-workflow-workbench` for epic, multi-path, deep-evidence, or
client/showcase presentation.

These are optional visual companions, not PR requirements. If an optional visual
capability is unavailable, the static PR first screen remains complete and
delivery continues.

## Companion Document

Link a companion document when the change required material research, design
tradeoffs, or a multi-slice plan. Structure it for mental alignment:

1. **Decision and outcome** — what is proposed or now true.
2. **Why** — the problem, constraints, and evidence that changed the decision.
3. **Reviewer-owned decisions** — hard responses only, with recommendations.
4. **Agent-owned decisions** — defaults and implementation choices already
   handled.
5. **Verification and risks** — proof, failure detail, boundaries, and deferred
   work.

Keep exploration chronology in linked research notes. Do not force the reviewer
to reconstruct the design from commits, logs, or chat.

## Generated Executive Brief

Use `kb-executive-brief` when the PR or companion document needs a reusable
first screen rather than a one-off handwritten summary.

```powershell
go run ./cmd/kbbrief -input <brief.json> -output <brief.md>
```

The JSON is authoritative; generated Markdown is a projection. Update the JSON
and regenerate instead of hand-editing the output. The generator keeps response
responsibility, outcome, key points, agent-handled work, proof, and risk
separate. In `auto` mode it emits Mermaid only with at least three nodes and two
edges, so a simple list does not become decorative visual noise.

## Interactive PR Workbench

Use `pr-review-workbench` only after an open PR exists and only when the user or
repo policy asks for the visual. The PR first screen above remains complete
without it.

The normal delivery is one self-contained HTML file:

1. Generate and verify it locally against the immutable PR head SHA.
2. If another reviewer needs it, store it on a separate
   `pr-review-artifacts` branch at
   `pr-<number>/<reviewed-head-sha>/index.html`.
3. Link the GitHub file page with **Download, then open locally in a browser**.

Do not commit it to the PR branch; that changes the SHA it claims to review.
GitHub renders repository HTML as source rather than as an interactive page.
Pages is an optional, explicitly authorized view over the separate artifact
branch, not a requirement for this workflow. Keep private evidence private.

## HumanLayer Ideas Adopted

This contract applies HumanLayer's high-leverage review pattern:

- concentrate human attention on research, design, and plan decisions where an
  early correction prevents downstream rework;
- make code review explain both how and why so the reviewer reaches mental
  alignment without reverse-engineering intent;
- treat the design document as the interface to the work, with decisions fed
  back into the artifact;
- return compact success evidence and full failure detail when the failure
  changes the decision.

Sources:

- <https://www.humanlayer.dev/blog/advanced-context-engineering>
- <https://www.humanlayer.dev/blog/writing-a-good-claude-md>
- <https://www.humanlayer.dev/>
- <https://www.humanlayer.dev/blog/context-efficient-backpressure>
