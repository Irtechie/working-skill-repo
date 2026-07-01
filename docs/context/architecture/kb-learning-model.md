# KB Learning Model (kb-native, scoped hierarchy)

Status: contract (slice-011 of kb-2026-07-01-native-scoped-learning)
Created: 2026-07-01

This is the authoritative contract for how the KB skill bundle stores and retrieves
learning. Slices 012-016 implement it. The two problems it fixes:

1. The bundle stored learning/run-state under an `.atv/` root and read an optional
   ATV hook feed, coupling it to the ATV install.
2. Learning had only a coarse `domain` enum and one global `project.yaml`, biased
   toward global-by-default, so a lesson could not attach to the exact scope that
   owns it (the fleet-eval face-ID wrong-reference incident: an image lesson could
   not be stored as an image lesson).

## 1. Canonical roots (replacing `.atv/`)

**Durable, git-TRACKED** — `docs/context/kb/`:

| Path | Holds |
|---|---|
| `docs/context/kb/instincts/project.yaml` | project-tier + global-tier instincts (tagged by `scope`) |
| `docs/context/kb/instincts/scoped/<scope-path>.yaml` | workflow/domain and sub-component instincts |
| `docs/context/kb/instincts/archive/` | decayed/evolved instincts |
| `docs/context/kb/kb-completions.txt` | kb-complete counter |

**Ephemeral, git-IGNORED** — `.kb/`:

| Path | Holds |
|---|---|
| `.kb/snapshots/<slice-id>.json` | regression snapshots |
| `.kb/qa-screenshots/` | QA browser captures |
| `.kb/observations.jsonl` | optional tool-use feed |
| `.kb/eval-runs/`, `.kb/pipeline-runs/` | run artifacts |

The ATV observer hook (`.github/hooks/copilot-hooks.json`) that writes
`observations.jsonl` is OPTIONAL. Learning works without it (git history +
docs/solutions + steering + explicitly recorded evidence). Passive capture is a
nice-to-have, never a requirement.

## 2. Scope hierarchy (the core model)

Learning lives at the **narrowest scope that owns it**, and climbs only on
recurrence:

```
global            (rare; domain-neutral universal process lessons)
  └─ project      (spans multiple workflows in one project)
       └─ workflow/domain      (audio, image, video, motion) ← DEFAULT HOME
            └─ component/surface (audio/voice-eval, image/comparer)
```

- **Default = narrowest owning scope.** "If audio fails, it's an audio issue."
  Most lessons stop at their workflow/domain (or a sub-component of it).
- **project** tier: only genuinely project-wide conventions spanning workflows.
- **global** tier: the RARE exception — universal, domain-neutral process lessons
  (e.g. "read the manifest, don't hand-glob"). Never a default.
- A component is a sub-scope of its workflow and inherits the workflow above it.

### Scope schema

Add `scope:` to the instinct format. It is a hierarchical path.

```yaml
instincts:
  - id: kebab-case-unique-id
    scope: audio/voice-eval        # narrowest owning scope; default is NOT project
    trigger: "when [specific situation]"
    behavior: "do [specific action]"
    confidence: 0.5
    domain: code-style|testing|architecture|error-handling|workflow|tooling
    observations: 1
    first_seen: YYYY-MM-DD
    last_seen: YYYY-MM-DD
    evidence:
      - "<commit / observation / user correction>"
```

- Workflow/component instincts live in `scoped/<scope-path>.yaml` (e.g.
  `scoped/audio.yaml`, `scoped/audio/voice-eval.yaml`).
- `project` and `global` tier instincts live in `project.yaml`, distinguished by
  `scope: project` or `scope: global`.

### Pull rule (retrieval)

When working in scope `S`, load `S` + all its ANCESTORS, and nothing from siblings:

```
working in audio/voice-eval  ->  load: audio/voice-eval, audio, project, global
                                  never:  image, video, motion, image/comparer
```

X pipeline never sees Y pipeline's lessons unless a lesson was promoted to a
common ancestor.

### Promotion-on-recurrence (the only climb path)

A lesson climbs ONLY when the same `trigger`+`behavior` pattern independently
recurs across **sibling scopes**, and it climbs to their **nearest common
ancestor**, not straight to global:

| Recurs in | Promotes to |
|---|---|
| `audio/tts` and `audio/sfx` | `audio` |
| `audio` and `image` | `project` |
| pattern is domain-neutral AND recurs across projects | `global` |

The promoted instinct cites the originating scopes as evidence. No lesson reaches
global by any other path.

### Landmine fast-path

A verified high-severity trap (per the existing landmine rule) is an **instant
one-shot learn**: one observation is enough, recorded immediately at its owning
scope (workflow/component), and additionally to `docs/context/landmines.md` when it
is a concrete repo trap. Ordinary small lessons do NOT get the fast-path and do NOT
default upward.

## 3. Decay + caps (per scope)

Apply the existing time-decay (`0.5^(days/90)`) and the 50-instinct cap PER scope
file. The global/project bucket and each scoped file are capped independently, so a
busy global tier cannot crowd out scoped learning and vice versa.

## 4. Feedback-routing classification

| Route | Use When | Durable Output |
|---|---|---|
| `current-only` | changes only the active PR/session | manifest/PR note |
| `steering-memory` | should steer future target selection, not yet an instinct | goal ledger / `docs/context/operations/steering/<slug>.md` |
| `observation` | evidence point for later extraction | `.kb/observations.jsonl` |
| `landmine-candidate` | verified repo-specific trap (instant, scoped) | `docs/context/landmines.md` + owning scope |
| `scoped-instinct` | ordinary lesson owned by one scope (DEFAULT) | `scoped/<scope-path>.yaml` |
| `instinct-evidence` | pattern proven across sibling scopes (via promotion) | ancestor scope / `project.yaml` |

Ordinary lessons default to `scoped-instinct` at the narrowest scope.
`instinct-evidence` (a higher tier) is reached only via promotion-on-recurrence.

## 5. How a skill declares its active scope

A skill/agent working a task determines its scope from, in priority order:
1. an explicit `scope:` argument passed to `/learn`;
2. the workflow/domain of the touched surface (e.g. edits under an `audio`
   pipeline, or a skill named `*-audio-*`, imply `audio`);
3. fallback `project` only when no narrower scope is identifiable.

Never fall back to `global`. Global is reached only by promotion.

## Non-goals

- This contract governs the skill bundle. Applying it to a specific downstream
  component (e.g. an image-comparer's `calibration.yaml` + fixtures) is a separate
  consumer task in that repo.
