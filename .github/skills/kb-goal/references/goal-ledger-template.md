# Goal ledger template

Fill this shape at `docs/context/goals/<goal-slug>.md`. `kb-goal` owns the
rules; this file is the layout only.

# <Goal Name>

Status: active|paused|blocked|complete|parked
Created: YYYY-MM-DD
Last updated: YYYY-MM-DD

## Objective

One sentence.

## Done Criteria

- [user] <observable condition the user actually asked for>
- [derived] <condition the agent added, naming which [user] item it serves>

## Terminal Proof

- <command, gate, artifact, or review condition required before completion>

## Done Check

- Type: command_exit|artifact_exists|gate|human_exception
- Check: <exact command, artifact path, gate id, or exception summary>
- Expected result: <exit code, path condition, gate status, or approval source>
- Why sufficient: <which done criterion this proves>

## Current State

- Current artifact: <manifest/epic/handoff/path or none>
- Next allowed action: <exact KB command>
- Last proof: <command/artifact/status or none>

## Live Steering (optional)

Use this block only for recurring, scheduled, or trend-improvement goals where
future runs should be steered by measurements and durable feedback. Omit it for
ordinary one-shot goals.

- Set point: <desired invariant, threshold, or direction>
- Sensor: <command, query, test, or review signal that measures the gap>
- Controller: <how the next reviewable increment is selected>
- Actuator: <KB lane, coding agent, or workflow that applies the increment>
- Disturbances: <outside changes the loop must tolerate>
- Dampener: <optional check that prevents the measured issue getting worse>
- Scope gate: <paths or systems the loop may change/read>
- Batch size: <maximum targets per run>
- WIP bound: <maximum active manifests/PRs/work items for this loop>
- Steering memory: <goal-ledger section or docs/context/operations/steering/<slug>.md>

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|

## Notes
