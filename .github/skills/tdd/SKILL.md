---
name: tdd
description: "Test-driven development with red-green-refactor loop using vertical slices. Use when building features test-first, fixing bugs with regression tests, or when the user says 'TDD', 'test first', 'red-green-refactor', or wants integration-style tests that verify behavior through public interfaces."
argument-hint: "[feature or behavior to implement test-first]"
---

# Test-Driven Development

Compatibility lane for explicit "TDD" or "test first" requests.

Normal KB work gets the anti-cheat behavior from `kb-plan`, `kb-work`, and
`kb-check`: define the oracle before implementation when practical, prove RED
when possible, protect the oracle with SHA/manifest evidence, implement, then
rerun the unchanged oracle.

## Keep

- Test observable behavior through the public interface.
- Write one vertical slice at a time: one failing check, one implementation,
  one passing check.
- Do not bulk-write imagined tests for future behavior.
- Do not edit the protected oracle after implementation unless the manifest
  records the reason and verification reruns from the changed oracle.

## Route

Use this skill only when the user explicitly asks for TDD/test-first flow or a
slice plan has `verification: tdd`.

Otherwise:

- `kb-plan` records the protected oracle fields.
- `kb-work` preserves the oracle and runs the red/green proof.
- `kb-check` validates protected file hashes and regression proof.

## Output

Report the test file or command, RED result, protected-oracle hash/manifest
path, GREEN result, and final unchanged-oracle verification.
