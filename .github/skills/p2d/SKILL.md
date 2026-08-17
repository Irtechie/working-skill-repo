---
name: p2d
description: "Plan to done. Takes a feature description, requirements source, or slice plan through kb-plan and straight into w2d, ending at an accepted PR when permissions and required checks allow. Use when the user says p2d, 'plan to done', or wants an unplanned idea carried all the way without staged confirmations."
argument-hint: "[feature description, requirements path, or plan path]"
---

# P2D

Take unplanned intent all the way to an accepted PR.

```text
idea / requirements / plan
  -> kb-plan        (manifest + plan-to-work gate)
  -> w2d            (work, finalize, ship, land)
```

`p2d` is `kb-plan` followed by `w2d`. It owns no execution or delivery logic of
its own; it exists so the user names one endpoint instead of two.

## Authorization

Explicit `p2d` invocation carries the same run-scoped merge authorization as
`w2d`: the user asked for done.

Authorization is intent, never capability. `w2d` still requires satisfied
permissions, protection, checks, and approvals before merging, and stops at an
open PR when any is missing.

## Preconditions

1. Run `kb-map lookup <request>` and resolve the active project root.
2. Claim or resume the objective in the shared `kb-start` work queue before any
   mutating phase.
3. Resolve the input before creating artifacts:

| Input | Action |
|---|---|
| validated manifest | skip planning, invoke `w2d <manifest>` |
| requirements source or slice plan | `kb-plan <path>` |
| clear feature description | `kb-plan <description>` |
| fuzzy idea or open product decisions | `kb-brainstorm` first, then resume |
| handoff or brainstorm doc | follow its source/manifest pointers first |

## Flow

1. Invoke `kb-plan <input>`. Planning owns slice design, the DAG, and the
   `plan-to-work` gate.
2. If planning blocks on a hard question, stop and ask it. Do not slice
   unresolved product intent to keep the chain moving.
3. Once `plan-to-work` passes, invoke `w2d <manifest>` and let it run to a
   single delivery state.
4. Report `w2d` output plus the manifest planning produced.

## Scope Rules

- Delegated phases keep their own gates and authority. `p2d` never grants a
  phase a permission that phase does not hold.
- Do not duplicate planning when a valid manifest already exists.
- Do not stage, commit, revert, or overwrite unrelated dirty work.
- Speed is not permission to skip planning. A fuzzy request still earns a
  manifest before any code is written.

## Stop Rules

- Do not execute a manifest that has not passed `plan-to-work`.
- Do not answer a hard product question on the user's behalf to avoid stopping.
- Do not re-plan work that `w2d` can already resume.
- Do not delete the current session worktree.

## Output

```text
P2D: <input>
Manifest: <path>
```

followed by the `w2d` delivery report.
