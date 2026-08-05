# Slice Template

Record ID, title, blockers, verification mode, test level, functional risk,
execution class, minimum model tier and reason, `model_requirements`,
`escalation_triggers`, `token_budget`, expected files, conflict domains, shared
resources, proof check, HITL classification, acceptance criteria, test
scenarios, scope boundary, status, and blocker lifecycle fields.

Keep one observable outcome per slice. Expected files are a forecast, not a
write allowlist.

The executor sees this slice and nothing else. Write it so the declared tier can
finish unaided, at the precision that tier requires (see Execution Contract
Precision in SKILL.md): `medium` gets ordered steps with a pass criterion each;
`small` also gets exact edit sites, the expected observable output, and one
proof command.
