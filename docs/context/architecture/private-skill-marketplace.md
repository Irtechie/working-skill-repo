# Private Skill Marketplace

Checked: 2026-05-31

## Purpose

`E:/agent-marketplace` is the private trusted catalog for reusable agent skills
and pipelines. It is not a global install directory. It is an approval boundary
between project-local experiments, public marketplace imports, and global agent
skill directories.

The working skill repo knows about the marketplace through
`config/skill-marketplace.json`.

## Directory Contract

```text
E:/agent-marketplace/
  skills/       Approved reusable skill directories.
  pipelines/    Approved reusable pipeline manifests.
  harnesses/    Approved proof/check harness definitions.
  catalog/      Approval and quarantine indexes.
  quarantine/   Untrusted public or experimental imports.
  scripts/      Pull, ingest, review, and sync helpers.
```

Consuming projects keep local drift and run evidence in their own repos:

```text
<project>/
  .github/skills/learned-*          Project-local learned skills.
  config/pipelines/*.json           Project-local or pinned pipeline variants.
  .atv/pipeline-runs/               Run evidence.
  .agent-marketplace/skill-lock.json Pinned marketplace imports and valid drift.
  docs/context/marketplace-promotion.md Promotion notes when needed.
```

## Lifecycle

1. A project-specific behavior is learned in the consuming project.
2. `learn` records observations and `evolve` may create
   `.github/skills/learned-*` in that project after maturity and approval.
3. The learned skill stays project-local until it proves reuse value.
4. After repeated successful reuse, review, and human approval, the skill can be
   copied into `E:/agent-marketplace/skills/<skill>`.
5. Global installs pull only explicitly selected approved skills or loader
   skills. Quarantine never installs globally.

Pipeline promotion follows the same shape:

1. Start as a project-local `config/pipelines/<pipeline>.json`.
2. Accumulate successful run artifacts under `.atv/pipeline-runs/`.
3. Reference proof harnesses by ID rather than hard-coding every check into the
   skill.
4. Promote to `E:/agent-marketplace/pipelines/<pipeline>.json` only after proof
   artifacts, review, and approval.

## Promotion Signals

Default minimums live in `config/skill-marketplace.json`:

- confidence greater than or equal to `0.85`;
- at least `5` observations;
- last seen within `90` days;
- at least `2` successful reuses outside the original moment;
- explicit human approval before marketplace promotion.

These are minimums, not automatic permission. Security-sensitive, destructive,
or broad workflow skills still need manual review even when the numbers pass.

## Trust Rules

- Public marketplaces are discovery sources only.
- Public imports land in `E:/agent-marketplace/quarantine/`.
- Runtime loaders must never read from quarantine. The blocking firebreak is
  `scripts/skill-marketplace-firebreak.ps1`, wired into `kb-check -All`.
- Approved marketplace skills are pinned by hash before pull/install.
- Global directories stay small. They are runtime install targets, not the
  private catalog.
- Project-local skills remain in the project until there is evidence they are
  reusable outside that project.
- Valid drift is recorded in `.agent-marketplace/skill-lock.json` with a reason
  instead of being treated as a sync error.

## Approved Security Skill

`atv-security` is approved as a trusted-source exception because it comes from
the local ATV security plugin maintained by trusted people. It is installed as a
single global security capability, not as a bulk import of ATV skills.

The approval is pinned in `E:/agent-marketplace/catalog/approved-skills.json`.
The approved marketplace copy and global Codex, Copilot, and shared agents
copies must hash-match the trusted ATV source before they are treated as
current.

The paired `dependency-vulnerability-osv` harness records the OSV Scanner proof
command for A06 dependency vulnerability checks. Missing `osv-scanner` is a
tooling skip, not permission to invent vulnerability findings from package age.

## Skill, Pipeline, Harness Boundary

- Skills perform work.
- Pipelines compose skills for a domain workflow.
- Harnesses prove pipeline outputs.

For example, an audiobook pipeline may compose `audiobook-book-writer`,
`audiobook-image-writer`, and `audiobook-audio-writer`, then prove outputs with
manuscript, image-artifact, and audio-file harnesses. A PowerPoint image review
pipeline can reuse the same image core ideas but use a presentation-specific
adapter and harness.

When a project changes a marketplace skill:

- general bug fix or invariant improvement: propose upstreaming to the base
  marketplace skill;
- domain-specific behavior: keep or promote as a named adapter skill;
- one-project taste or app constraint: keep project-local and record valid
  drift in the lock file.

## Implementation Notes

The first useful scripts should be boring and read-only:

- list marketplace contents;
- validate catalog JSON;
- compare approved skill hashes;
- report project-local promotion candidates.

Mutation scripts should come later:

- ingest public skill into quarantine;
- promote reviewed skill into `skills/`;
- pull selected approved skills into a project or global target.
