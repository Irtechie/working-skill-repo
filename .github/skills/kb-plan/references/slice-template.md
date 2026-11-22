# Slice Template

Record ID, title, blockers, verification mode, test level, functional risk,
execution class, minimum model tier and reason, `model_requirements`,
`escalation_triggers`, `token_budget`, `cost_tier` with the cheaper option ruled
out, expected files, conflict domains, shared
resources, proof check, HITL classification, acceptance criteria, test
scenarios, scope boundary, status, and blocker lifecycle fields.

Record the work boundaries (see Work Boundaries in SKILL.md):

```yaml
touch:
  - <path or folder this slice is expected to change>
do_not_touch:
  - path: <path>
    reason: <owned by slice N | safety-critical | generated | outside component>
shared_edits:
  - path: <file more than one ready slice changes>
    rule: <additive only | one named block | after slice N>
tests:
  update: [<existing test files expected to change>]
  add: [<new test files>]
  run:
    - <exact command, with its filter>
  not_run: <what is deliberately skipped, and why>
evidence: <where screenshots or logs go; must outlive the workspace>
preserved_behavior:
  - rule: <a safety, confirmation or permission rule that must still hold>
    proof: <the test that shows it>
```

Keep one observable outcome per slice. Expected files and `touch` are a
forecast, not a write allowlist. `do_not_touch` is a hard boundary: the executor
stops and escalates instead of crossing it.

The executor sees this slice and nothing else. Write it so the declared tier can
finish unaided, at the precision that tier requires (see Execution Contract
Precision in SKILL.md): `medium` gets ordered steps with a pass criterion each;
`small` also gets exact edit sites, the expected observable output, and one
proof command.
