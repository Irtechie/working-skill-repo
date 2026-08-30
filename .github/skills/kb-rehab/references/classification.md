# Classification Reference

## Lifecycle States

| State | Meaning | Evidence required |
|---|---|---|
| `live` | Another session holds a claim on this work | An unexpired queue claim whose status is not terminal |
| `unshipped` | The ref holds commits the authoritative default does not contain | `git cherry` reports at least one `+` commit |
| `dead` | The ref's work is fully contained and no declaration keeps it open | Containment proof against the freshly fetched default SHA |
| `superseded` | A named replacement already exists on the default branch | The replacement path resolves in the default tree **and** the ref is contained |
| `orphan-work` | A declared item names no reachable ref or artifact | The named manifest or branch does not resolve |
| `orphan-branch` | A ref exists with no declared work item | No declared item names this ref |
| `human-required` | Evidence is missing or contradictory | Recorded per pairing in `reason` |

## Evidence Rules

Containment is `git cherry` patch equivalence, not ancestry. Ancestry
misclassifies a squash-merged branch as unshipped, and re-merging already-landed
work is the expensive failure this lane exists to prevent.

Remote authority is resolved through a fresh `git ls-remote --symref` plus a
fetch-equality check. A remote-tracking ref is a cache, not an authority.

Supersession is never self-proving. The replacement path must already exist in
the authoritative default tree, so a branch can never author its own
replacement into existence.

A missing file is missing evidence, not proof of death. An unparsed row stays in
the report as `orphan-work` rather than being dropped.

## Fail-Closed Triggers

Each of these sets `status: fail-closed` and yields zero `dead` and zero
`superseded` pairings:

- no remote is configured;
- the remote is unreachable;
- the advertised default branch does not resolve;
- the advertised SHA and the fetched SHA disagree;
- two remotes disagree about the default branch; or
- `config/rehab-policy.json` is absent or unparseable.

A fail-closed report also refuses every `--action mark` and `--action remove`
write.

## Protected Paths

`config/rehab-policy.json` names the protected roots. In this bundle a merge
touching `.github/skills` propagates to `~/.agents/skills`, `~/.codex/skills`,
and `~/.copilot/skills`, so it is a write into every future agent session on the
host. The packet states that consequence verbatim; it is never summarized away.
