---
name: handoff
description: Compact the current conversation into a handoff document for another agent to pick up.
argument-hint: "What will the next session be used for?"
---

Write a handoff document summarising the current conversation so a fresh agent can continue the work. Save to `C:\Users\marowe\.copilot\handoffs\` with a timestamped filename (e.g., `2026-05-23-token-optimization.md`).

Credit: Based on mattpocock/skills (MIT). Adapted for ATV workflows.

## What to include

1. **Context** — What repo, branch, and working directory. What the user was doing.
2. **Decisions made** — Key architectural or design decisions from this session, with reasoning.
3. **Work completed** — Files changed, commits made, what's done.
4. **Work remaining** — What's left, any blockers, next steps.
5. **Key files** — File paths the next session will need to read first.
6. **Suggested skills** — Skills the next agent should invoke to continue.

## Rules

- Do NOT duplicate content already in other artifacts (plans, brainstorms, PRDs, committed docs). Reference by path instead.
- Redact any sensitive information (API keys, passwords, PII).
- Keep it under 2000 words — this is a handoff, not a novel.
- If the user passed arguments, treat them as a description of what the next session will focus on and tailor accordingly.
- If a `plan.md` exists in the session state folder, reference it rather than repeating its content.
