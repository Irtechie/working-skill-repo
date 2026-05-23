# KB Work Slice Execution Prompt

Use this as the per-slice execution checklist.

```text
You are executing a single vertical slice. Complete it fully.

KB: <kb_id>
Slice: <slice_id> - <title>
Verification mode: <tdd|integration|functional|verification-only>

Plan contents:
<full slice plan content>

Instructions:
1. Read the plan completely.
2. Set up on the current branch.
3. For files marked `op: edit` in expected_files:
   - Read the current file content first.
   - Make only the change described in the `scope` field.
   - Preserve all existing behavior not mentioned in scope.
   - Current disk content is authoritative over stale plan text.
4. For files marked `op: create`, create the planned file.
5. Apply the verification mode:
   - tdd: failing test -> implementation -> passing test -> refactor.
   - integration: integration test proves the wired path.
   - functional: workflow/API/CLI/UI path is proven from public surface.
   - verification-only: build/check proves no regression.
6. Run relevant deterministic checks first, then broader checks when practical.
7. Stage only files changed for this slice.
8. Commit only if the user asked for commits.

Do not modify other slices' files unless required for this slice.
Do not add scope beyond what the plan specifies.
Do not stage unrelated dirty or untracked files.
```
