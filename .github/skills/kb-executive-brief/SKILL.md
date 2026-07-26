---
name: kb-executive-brief
description: Generate a low-cognitive-burden executive brief, with an optional deterministic Mermaid flow, from source-owned structured data. Use when the user asks for an executive summary, decision brief, visual summary, low-burden status artifact, or a PR/companion first screen that should make required human attention obvious.
argument-hint: "[source JSON path and optional output Markdown path]"
---

# KB Executive Brief

Create the smallest first screen that lets a busy person understand the outcome,
see whether they must respond, and inspect proof or risk without reconstructing
the work.

## Non-Goals

- Do not summarize from memory when a source-owned artifact exists.
- Do not create decorative charts, invented scores, fake precision, or a visual
  that merely repeats a short list.
- Do not hide blockers, safety context, proof, or risk to make the brief shorter.
- Do not turn an agent-owned choice into reviewer work.

## Responsibility First

Classify the apparent question before writing:

| Class | Brief behavior |
|---|---|
| `hard_response_required` | Name the exact ask, why the human owns it, what blocks, and the recommendation |
| `soft_preference` | State the safe reversible default and continue unless overridden |
| `no_response_needed` | State that no response is required; do not manufacture a question |

## Visual Gate

Use a generated flow only when relationships materially reduce reading effort:

- at least three meaningful nodes and two edges;
- dependency, branching, ownership, or sequence is easier to see than to read;
- every node and edge comes from the source JSON.

Set `visual.mode` to `auto` for the normal gate, `always` only when the user
explicitly requested the visual and the input has at least two nodes plus one
edge, or `none` when prose is clearer.

For a pull request that needs file drill-down, evidence tabs, source anchors, or
an interactive walkthrough, use `pr-review-workbench` after the PR exists.
Mermaid remains the small static relationship view; it is not the PR review UI.

## Workflow

1. Read the authoritative status, plan, result, PR, or companion artifact.
2. Create schema-version-1 JSON using
   `cmd/kbbrief/testdata/executive-brief.json` as the compact example.
3. Keep `key_points` to five or fewer. Put supporting work under
   `handled_by_agent`, `verification`, and `risks_or_later`.
4. Generate the brief:

   ```powershell
   go run ./cmd/kbbrief -input <brief.json> -output <brief.md>
   ```

   Omit `-output` to print the Markdown to stdout.
5. Inspect the generated responsibility line, decision/default section, proof,
   risks, and Mermaid relationships. Fix the source JSON rather than hand-editing
   generated output.
6. Link the brief from the PR or companion document when it is the best first
   screen; keep detailed reasoning in its source-owned artifact.

## Output Contract

- Lead with `Response required`, then outcome and key points.
- Render hard decisions, soft defaults, and no-response status differently.
- Render Mermaid only when the visual gate passes.
- Preserve exact commands and evidence strings supplied by the source.
- Report the source JSON and generated Markdown paths.
