# Planner Gate Ledger

The manifest gate ledger is the phase-transition source of truth. Each record
names the gate ID, owner skill, status, required evidence, proof paths,
blockers, timestamp, and exact allowed next action.

Planning writes `plan-to-work: passed` only after source assurance, DAG,
slice-contract, context-packet, and manifest validation succeed.
