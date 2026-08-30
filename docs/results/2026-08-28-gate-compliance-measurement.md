# Gate Compliance Measurement

Recorded: 2026-08-28

## Question

`todo.md` carried a queued item to build Codex/Claude runtime hook files that
would block stop and phase advancement when a gate says blocked. That item
assumed a compliance failure: that sessions skip the deterministic gate unless a
runtime forces it.

Before building the hook layer, measure whether the assumed failure occurs.

## Method

Source: Copilot session store (`sessions`, `tool_requests`), repository
`Irtechie/working-skill-repo`.

Population: all sessions in the trailing 30-day window, cross-checked against a
trailing 60-day window.

Gate-run signal: a tool call whose arguments contain `run ./cmd/kbcheck`.

An earlier pass matched the looser pattern `%kbcheck%` and produced a badly
inflated count. That pattern also matches every file read, grep, and edit of
`cmd/kbcheck/*.go`, which this repository contains. The loose pattern reported
5,734 hits across 24,410 tool calls (23.5%); the invocation-only pattern reports
803 across 14,396 (5.6%). Only the invocation-only figure is meaningful.

## Result

| Measure | Value |
|---|---|
| Sessions in 30-day window | 23 |
| Sessions that ran the gate at least once | 22 |
| Sessions with zero gate runs | 1 |
| Skip rate among sessions performing work | 0% |
| Total gate runs | 803 |
| Gate runs as share of all tool calls | 5.6% |
| Median gate runs per session | ~15 |
| Minimum gate runs in a working session | 2 |

The single zero-gate session (`dd4da98a`, 12 tool calls) was
`Code review skill research` and performed no work.

The 60-day window agrees. It adds exactly one further zero-gate session
(`06e4ae88`), a model-routing benchmark task executing an external `PLAN.md`
outside KB workflow entirely.

## Conclusion

The `never ran the gate` failure mode did not occur in any session that
performed work. A runtime hook that compels `kbcheck` execution would fire on a
behavior with zero observed instances, while introducing per-provider hook files
that conflict with the portability contract in `AGENTS.md`.

Hook enforcement is not justified by evidence. The queued item is replaced by a
narrower one.

## Not Measured

The more dangerous failure mode is distinct and remains unquantified: a session
runs the gate, receives a failure, and advances anyway.

Telemetry could not answer it. Every query joining `tool_requests` to
`events.tool_complete_success`, and every scan of
`events.tool_complete_result_content`, exceeded the 60-second query timeout.

The repo-local proxy is suggestive but weak. `todo-done.md` carries 34 explicit
`Proof:` citations and one override-shaped line (`todo-done.md:262`), which
scopes an optional ATV drift warning explicitly rather than bypassing it. That
is the intended pattern, not a silent bypass. The evidence is circular: a
self-reported completion record is poor evidence about self-reporting failures.

Closing this gap needs a gate-ledger assertion inside `kbcheck` that compares
recorded gate results against advancement, not a runtime hook.
