# KB ablation protocol

`skill-eval-ablation` is an offline reducer for imported, independently
evidenced task outcomes. It does not run a model or commands named by records.

```powershell
go run ./cmd/kbcheck skill-eval-ablation --result-root evals/skill-eval/ablation/results --output evals/skill-eval/ablation/report.json
```

Only matching full/reduced/none triples from the same case, repetition, host,
configuration fingerprint, project hash, and task-prompt hash are comparable.
Synthetic, route-only, self-reported, or unproven results are excluded.

The later live pilot requires separate explicit authorization for model use and
spend. Verify effective skills and every ambient instruction source per host;
if treatment isolation cannot be demonstrated, mark that host unsupported.
