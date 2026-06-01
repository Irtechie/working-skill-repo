---
applyTo: "evals/**/*.json,evals/**/*.md"
---

Eval fixtures must contain fixed inputs and deterministic expectations. Do not
label a fixture as quality, proof, or regression coverage unless a script reads
the expected fields and can fail mechanically.

Live-model fixtures must stay explicit: local validation may check schema,
route, trace, claims, and protected hashes, but must not call Codex, GHCP, or
other model CLIs unless the runner is explicitly invoked in live mode.

For trace rules, required commands are intent checks, while forbidden commands
and no-write safety require observed evidence when a wrapper captured it.
