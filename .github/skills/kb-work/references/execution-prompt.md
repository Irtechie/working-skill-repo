# KB Work Slice Execution Prompt

Use this as the per-slice execution checklist.

```text
You are executing a single vertical slice. Complete it fully.

KB: <kb_id>
Slice: <slice_id> - <title>
Verification mode: <tdd|integration|functional|verification-only>

Plan contents:
<full slice plan content>

Context packet:
<validated packet, or explicit small legacy/no-packet reason>

Route request:
<required tier and tier reason, execution_owner current|delegated, owner reason,
saved user-local project source priority, user overrides, and active host
surface>

Router commands:
<current/no-router reason; exact native host target and tool call; or exact
kbrouter discover, delegated-select, and dispatch commands with slice-unique
artifact names>

Slice lease:
<exact slice-lease acquire command, owner token source, generation, renewal/release command, or non-mutating/no-lease reason>

Execution policy:
<one owner decision, capability evidence, exact proof, and escalation triggers>

Instructions:
1. Read the plan completely.
2. Set up on the current branch.
3. If the slice runs Go inside a workspace sandbox, load
   `references/go-sandbox.md` and apply its environment inside every Go shell
   invocation. Never put its temp overrides on the agent launcher.
4. Use the packet's files and deterministic prefetch before broad search.
   Escalate when an escalation trigger fires or the packet is insufficient;
   do not silently expand authority.
5. Treat model routing as an owner-first live decision, not a plan commitment.
   The current orchestrator records exactly one owner immediately before
   execution: retain `current` when its reasoning, context, tools, trust, or
   authority are required; otherwise choose `delegated`.
6. Apply user-local project source priority (`automatic`, `self-hosted-first`, or
   `native-first`) only among eligible delegated routes. Plans never choose a
   model, alias, adapter, endpoint, or transport. Only run-scoped
   `require <model>` hard-pins.
7. Record the actual route and provenance only from dispatcher or host evidence.
   Inspect the active host's exact callable-agent schema and the live CLI/user
   catalog; never infer callable aliases from model memory or merge App-only and
   CLI-only catalogs.
   If the host cannot prove the selected model/session, report provenance as
   `unknown`/`unavailable`.
8. For `current`, validate current capability and execute without searching
   workers first. For native host delegation, call the exact native target
   directly. For CLI/user-local delegation, use `kbrouter` and select exactly
   one qualified same-tier or higher route. Do not send App targets through
   `kbrouter`, route downward, or silently fall back across owners.
9. Run the exact deterministic proof. Proof—not model self-review or a routing
   receipt—is the acceptance oracle. If proof fails, use ordinary bounded repair
   under the same owner. Re-plan, block, or record a new explicit ownership
   decision if the required authority changes.
10. For files marked `op: edit` in expected_files:
   - Read the current file content first.
   - Make only the change described in the `scope` field.
   - Preserve all existing behavior not mentioned in scope.
   - Current disk content is authoritative over stale plan text.
11. For files marked `op: create`, create the planned file.
12. Apply the verification mode:
   - tdd: failing test -> implementation -> passing test -> refactor.
   - integration: integration test proves the wired path.
   - functional: workflow/API/CLI/UI path is proven from public surface.
   - verification-only: build/check proves no regression.
13. Run relevant deterministic checks first, then broader checks when practical.
14. Release or renew the slice lease with the same owner token and current generation before handing back the slice.
15. Stage only files changed for this slice.
16. Commit only if the user asked for commits.

Do not modify other slices' files unless required for this slice.
Do not add scope beyond what the plan specifies.
Do not stage unrelated dirty or untracked files.
Do not invoke AMR or pass `attempt_tier` during normal KB work.
```
