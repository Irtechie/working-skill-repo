---
applyTo: "scripts/**/*.ps1,cmd/**/*.go"
---

Gate scripts must print concise pass/fail output, exit nonzero on blocking
failures, and have a focused selftest when they encode policy.

Prefer native Go for top-level cross-platform orchestration. PowerShell helper
scripts may remain as individual validators until their behavior has Go parity
coverage.

Do not treat read-only reports as proof of mutation safety unless the script
also checks the external state it claims to protect.
