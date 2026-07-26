# Proof Governor Plan-to-Work Preflight

Date: 2026-07-26
Manifest: `docs/plans/2026-07-26-020-kb-change-aware-proof-governor-manifest.md`

## Gate Results

- `go run ./cmd/kbcheck manifest-contract --manifest <manifest> --json`
  passed before gate repair.
- `go run ./cmd/kbcheck core --list` passed and enumerated the contributor
  checks without executing them.
- All four context packets passed `kbcheck context-packet`.
- The four-slice DAG is linear with no missing reference or cycle.

## Dirty Baseline Classification

Current branch: `main`

Current revision: `31700f63f2ae1229337e2ac1f375d83299a5955e`

| SHA-256 | Existing dirty path | Classification |
|---|---|---|
| `267698afef9cf5439b1357246b5523b41790189c3f86adc022d111ec37c536d8` | `.github/skills/kb-work/SKILL.md` | DDR route-announcement hunks; preserve and integrate proof-governor edits around them |
| `effef5287cd275fec255400201089979bf8bb72c9bdb04f5e733d6d0863129cb` | `docs/context/architecture/kb-workflow.md` | DDR route-announcement documentation; preserve |
| `00013c3b4d15e94d2c366af06326c56289e1fd582ae2cbaa168e34a3cbc9a262` | `todo.md` | Concurrent plan and workstream rows; preserve |
| `1422b18191baebc8f77750f3a923414aba3e6391e24fd7f47a2ff506aff314e6` | `cmd/kbcheck/ddr_contract_test.go` | DDR contract work outside proof-governor implementation files |
| `4a337aa223ba594ec9ed5929f0b96cfa57ceea4d6a66d8c24ee1a38e546531c4` | `cmd/kbcheck/provider_hygiene.go` | Provider-hygiene work outside proof-governor implementation files |
| `b5f2f4ead516222dfa0d10f2dc592a81a27fc8ed666c466d1ff2a9511584cf93` | `cmd/kbcheck/provider_hygiene_test.go` | Provider-hygiene tests outside proof-governor implementation files |

## Execution Decision

- Execute serially in the current checkout.
- Preserve the classified dirty hunks; do not stage, commit, reset, stash, or
  revert them.
- Run only focused slice proof until the genuine final milestone.
- The final release slice remains blocked if full `core` or `local-release`
  cannot finish inside its bounded diagnostic window.
- No visible browser, WebView, Tauri, WDIO, or native GUI execution is approved.
