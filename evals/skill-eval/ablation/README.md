# KB ablation protocol

`skill-eval-ablation` reduces operator-selected external capture imports offline.
It never runs a model or a command named by an input. Hashes verify bytes and
matching identities; they do not authenticate the capture author. A forged but
internally consistent capture can pass. The operator must select a trusted
capture export, not ask the evaluated model to write its own success receipt.
Local test fixtures are synthetic reducer tests, never live benefit evidence.

```powershell
go run ./cmd/kbcheck skill-eval-ablation --result-root evals/skill-eval/ablation/results --output evals/skill-eval/ablation/report.json
```

Store top-level run records in `results/` and referenced artifacts elsewhere
under the repository root. Paths are repository-relative; symlinks, escapes,
files over 1 MiB, unknown JSON fields, duplicate keys and mismatched SHA256s are
rejected. There is a maximum of 512 records. Keep secrets out of captures.

Each record uses this shape (substitute actual lowercase SHA256s):

```json
{
  "schema_version": 1,
  "evidence_kind": "live",
  "identity": {
    "run_id": "pilot-small-fix-codex-full-1",
    "case": "small-fix", "repetition": 1,
    "host": "codex", "host_version": "captured-version", "model": "captured-model",
    "config_sha256": "...", "project_sha256": "...", "prompt_sha256": "...",
    "condition": "full"
  },
  "inventory": {"path": "captures/run-inventory.json", "sha256": "..."},
  "transcript": {"path": "captures/run-transcript.json", "sha256": "..."},
  "proof": {"path": "captures/run-command.json", "sha256": "..."}
}
```

The three referenced envelopes repeat `schema_version: 1`, the exact `identity`
object and `source: "external-capture"`. They contain:

- **Inventory:** `discovery_coverage` with `global`, `project`, `skills`, `native`
  all true; `skills` and `instructions` arrays. Every entry has `name`, `scope`,
  an actual boolean `kb`, and `content: {path, sha256}` referencing the effective
  instruction bytes. Include the host-native instruction baseline. The inventory
  file SHA256 is the effective profile identity, bound by both other envelopes.
- **Transcript:** `profile_sha256` and nonempty `content` holding the externally
  captured session transcript. The command receipt binds this file's hash.
- **Proof:** `profile_sha256`, `transcript_sha256`, exact `command` argv array,
  required integer `exit_code`, RFC3339 `started_at` and `ended_at`,
  `check_contract: {path, sha256}`, `outputs: [{path, sha256}]`, and `metrics`.
  The frozen check contract contains `schema_version: 1`, `case`, the identical
  command array and `oracle: {path, sha256}` for the unchanged check definition.
  Output references identify captured command output and produced artifacts.

Optional metrics are `tokens`, `cost_usd`, `interventions`, `artifact_count`,
`elapsed_ms`. Each is `{ "value": 123, "source": "external-capture" }`.
Missing, negative or unproven values become null, never zero. Count metrics must
be integers. Runtime of the proof command is not assumed to be task elapsed time.
Only the captured exit code determines task success; nonzero exits remain valid
failures. Inline `task_success` or `independent_proof` self-reports do not qualify.

Reports retain admitted failures, exclusions with reasons, missing groups,
matched arm outcomes and reduced/none minus full differences. Any repeated arm
or run ID invalidates every affected record. A comparison requires identical
case, repetition, host/version, model, configuration, starting project, prompt,
frozen check/oracle and non-KB instructions. Unknown differences stay null.
Correctness regressions remain explicit even when a failing arm costs less.
There are no pooled causal claims, automatic recommendations or significance
claims. Routing accuracy remains the separate route evaluation's outcome.

The later pilot needs explicit live-run and spend authorization. It is bounded
to four cases (small fix, ambiguous debugging, feature, resumed task), three
hosts, three conditions and one repetition: at most 36 task runs, no automatic
reruns. Freeze the initial repos, prompts, model/config, checks and oracle before
execution. Counterbalance condition order and use fresh sessions. Full uses the
pinned core catalogue; reduced uses its recorded experimental inventory; none
removes every KB skill and KB ambient instruction while retaining identical
host-native safety instructions. Verify effective global, local, project and
skill discovery for every host, not just requested configuration. Mark a host
unsupported if isolation cannot be demonstrated. Retain failed runs and all
capture artifacts; a cheaper failure never counts as an improvement.
