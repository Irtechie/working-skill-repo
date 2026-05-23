---
name: kb-functional-test
description: Functional-test strategy and test-quality audit for KB workflows. Use when deciding whether a slice needs functional/e2e/browser/API workflow tests, when existing tests may be mocked theater, or when user-visible behavior must be verified without manual QA.
argument-hint: "[slice plan, feature, changed files, or test file]"
---

# KB Functional Test

Functional tests prove the real workflow works. Unit tests prove parts. Both matter, but mocked tests that never exercise the behavior do not count.

## When Functional Tests Are Required

Require at least one functional or integration-style test when a change touches:

- user-visible UI flow;
- API route, command, tool/action, or workflow orchestration;
- auth, permissions, session, persistence, streaming, external integration, or background job behavior;
- wiring between two or more subsystems;
- a bug that escaped unit tests.

Skip or defer functional tests only when the change is pure copy/style, dead-code removal, local refactor with existing coverage, or a generated/config-only change covered by build/lint.

## Test Quality Audit

An existing test is meaningful only if it:

- would fail against the broken or pre-change behavior;
- exercises the public surface, not private implementation details;
- asserts observable output, persisted state, emitted event, response contract, DOM state, or side effect;
- mocks only external boundaries, not the behavior under test;
- can fail for the bug it claims to cover.

If a test mostly asserts mocks were called, snapshots noise, or duplicates implementation logic, mark it weak and add a better functional probe.

## Execution Timing

- **During a slice:** run the narrowest functional check that proves the changed path. Prefer headless browser/API/CLI checks.
- **After a manifest:** run broader workflow smoke tests across changed areas.
- **Before ship:** run full functional/e2e suite when practical, plus targeted high-risk flows.
- **During parallel work:** only one worker owns browser/e2e execution at a time. Other workers may run unit/lint/typecheck. Queue UI functional checks to avoid spawning many visible sessions.

## Browser / UI Rules

- Headless by default.
- Visible browser only when debugging visual behavior, SSO/CDP is required, or the user explicitly asks.
- Prefer programmatic probes: Playwright locators, API checks, DOM text extraction, screenshot assertions, or CLI commands.
- Save screenshots as evidence for UI functional checks:
  - one baseline/pass screenshot per tested page or major workflow state;
  - mandatory screenshot for each failure state;
  - responsive screenshots only for deep tier or layout-sensitive changes.
- Store evidence under `.atv/qa-screenshots/` or the repo's configured QA artifact path.
- Keep screenshots until `kb-complete` cleanup. Do not keep unlimited traces/videos unless they explain a failure.

## Script Rule

If a functional check will be repeated, turn it into a script or test:

- Playwright/Cypress test for UI flows.
- API smoke script for endpoints.
- CLI smoke script for commands.
- Small HTML/DOM extractor only when a full browser test is overkill.

Add the command to `docs/context/operations/testing.md` and make `kb-check` able to discover or call it.

## Output

Report:

- functional risk level: none, narrow, broad, full;
- tests audited;
- weak tests found;
- functional checks added or run;
- screenshot evidence path for UI checks;
- command/results;
- remaining manual-only verification and why.
