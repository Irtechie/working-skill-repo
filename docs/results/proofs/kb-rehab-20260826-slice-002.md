# Proof Receipt - kb-rehab slice-002

- kb_id: kb-2026-08-26-kb-rehab-outstanding-work
- slice_id: slice-002
- title: The lane, markers, and the packet
- recorded_at: 2026-08-26T23:40:00Z
- tree: working tree of `codex/w2d-unplanned-entry` at slice-002 acceptance
- route: current orchestrator (DDR exception `context-required`)

## Exact Proof

| Check | Command | Result |
|---|---|---|
| Slice proof | `go test ./cmd/kbcheck -run 'WorkReality\|Rehab\|SkillRepoContract' -count=1` | PASS (97.6s) |
| Package regression | `go test ./cmd/kbcheck -count=1` | PASS (238.8s) |
| Vet | `go vet ./...` | PASS |
| Build | `go build ./...` | PASS |
| Skill lint | `go run ./cmd/kbcheck skill-lint` | 0 errors; `kb-rehab` not flagged |
| Whitespace | `git diff --check` | clean |

## Protected Oracle

| Path | Role | SHA-256 |
|---|---|---|
| cmd/kbcheck/work_reality_test.go | Slice-001 lifecycle oracle extended with the slice-002 marker, removal-gate, preservation, and packet-bound oracle | 29d5fd11517756e463750e36342b8cb49f42680f135373846fd4d36c041be914 |

The slice-001 hash `d9592f7c...` is superseded by this hash. The change is
additive: nine tests were appended and no slice-001 assertion was weakened,
relaxed, or deleted.

## Acceptance Evidence

Each test scenario named in the slice plan has a test:

| Scenario | Test | Result |
|---|---|---|
| `dead` with containment proof | `TestRehabMarkWritesSkippedMarkerForProvenDeadWork` | row marked `⊘ skipped`, row retained |
| `superseded` with containment proof | `TestRehabRemovalPermittedWhenArtifactLandedAndRefContained` | classified `superseded` before any write |
| Ambiguous pairing | `TestRehabMarkPreservesAmbiguousWorkWithZeroWrites` | zero writes, `todo.md` byte-identical |
| Removal with uncontained commits | `TestRehabRemovalBlockedByUncontainedCommitsRemarksTheRow` | removal blocked, row re-marked `🔒 blocked` |
| Removal, artifact named, ref contained | `TestRehabRemovalPermittedWhenArtifactLandedAndRefContained` | exactly one recorded removal |
| Eight ambiguous pairings | `TestRehabPacketNeverExceedsFiveGroupedItemsAndDropsNothing` | exactly 5 items covering all 8, no mandated field empty |
| Merge decision in this repository | `TestRehabPacketNamesGlobalInstallRootsForSkillPaths` | irreversible consequence names all three global install roots |
| Fail-closed report | `TestRehabMarkRefusesEntirelyOnFailClosedReport` | every write refused, `todo.md` untouched |
| Default action | `TestRehabReportActionDefaultsToReadOnly` | `report`, no marks, no write |
| Row shape unproven | `TestRehabMarkLeavesRowsWithNoStatusMarkerUntouched` | recorded blocked, file unchanged |

Run against this repository, `go run ./cmd/kbcheck work-reality --root . --json`
returned `status=ok`, `action=report`, 65 pairings and 4 packet items:

| Packet item | Pairings |
|---|---|
| `packet-orphan-branch-.github-skills` | 2 |
| `packet-orphan-branch-cmd` | 1 |
| `packet-orphan-work-unprotected` | 57 |
| `packet-live-unprotected` | 2 |

Four items is within the five-item bound without truncation, and `todo.md` was
unchanged by the read-only run.

## Scope Comparison

Forecast six files; actual six files, all forecast, none discovered, none unused:

| Path | Op | Forecast |
|---|---|---|
| .github/skills/kb-rehab/SKILL.md | create | yes |
| .github/skills/kb-rehab/references/classification.md | create | yes |
| cmd/kbcheck/work_reality.go | modify | yes |
| cmd/kbcheck/work_reality_test.go | modify | yes |
| .github/skills/kb-start/SKILL.md | modify | yes |
| README.md | modify | yes |

`cmd/kbcheck/main.go` also changed to accept `--action` for `work-reality`. It
was forecast in slice-001 and is the same three-line command-surface edit, not a
discovered path.

## Design Decisions

**Grouping bounds the packet without dropping anything.** Items group by
`state/protection-scope`. When more than five groups exist the surplus folds
into the last item's `omitted_pairings` count rather than disappearing. A
dropped decision is an invisible decision.

**Marking reuses `todo.md`'s own vocabulary.** `⊘ skipped` is already how this
repository records superseded work (see the "Plan-to-PR finish lane" row). No
second vocabulary was introduced, so the `todo-triage` contract holds.

**A row whose shape is unproven is never rewritten.** A row carrying no known
status marker is recorded as blocked, not guessed at.

**Removal is a separate action.** `--action mark` never removes. `--action
remove` removes only when a superseding or completing artifact already resolves
in the authoritative default tree *and* the ref is contained.

**Line endings are preserved per line.** This repository is CRLF; the rewriter
strips and restores the trailing `\r` so a marker edit does not rewrite the whole
file in the diff.

## Memory Impact

Durable. `kb-rehab` is a new routable lane at `kb-start` rank 3a, and
`kbcheck work-reality` now has a write mode gated on the read-only report.
`README.md` records the lane in both the route table and the verification list.

## Scope Boundary Held

No PR, no push, no merge, no ref deletion, and no proof execution occurred in
this slice. Delivery and reaping remain slice-003.
