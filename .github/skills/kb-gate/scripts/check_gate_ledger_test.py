#!/usr/bin/env python3
"""Parity tests for the portable gate-ledger validator."""

from __future__ import annotations

import contextlib
import io
from pathlib import Path
import sys
import tempfile
import unittest

import check_gate_ledger as gate


class GateLedgerParserTests(unittest.TestCase):
    def test_list_entry_with_colon_is_not_a_phantom_gate(self) -> None:
        ledger = gate.parse_gate_ledger(
            """gate_ledger:
  - gate_id: dependency
    attempted:
      - command: go test ./cmd/kbcheck
    status: blocked
slices: []
"""
        )
        self.assertEqual(1, len(ledger))
        self.assertEqual(
            ["command: go test ./cmd/kbcheck"],
            ledger[0]["attempted"],
        )

    def test_inline_list_preserves_quoted_comma(self) -> None:
        self.assertEqual(
            ["Windows, Linux receipts missing"],
            gate.parse_inline_list('["Windows, Linux receipts missing"]'),
        )

    def test_lifecycle_boolean_is_tri_state(self) -> None:
        self.assertEqual(
            (True, True),
            gate.blocker_lifecycle_contract_value(
                "blocker_lifecycle_contract: true"
            ),
        )
        self.assertEqual(
            (False, False),
            gate.blocker_lifecycle_contract_value(
                "blocker_lifecycle_contract: ture"
            ),
        )

    def test_checked_at_requires_rfc3339_timezone(self) -> None:
        self.assertIsNone(gate.parse_gate_timestamp("2026-07-26"))
        self.assertIsNone(gate.parse_gate_timestamp("2026-07-26 10:00:00+00:00"))
        self.assertIsNotNone(gate.parse_gate_timestamp("2026-07-26T10:00:00Z"))

    def test_duplicate_gate_ids_cannot_advance(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            manifest = Path(temp) / "manifest.md"
            manifest.write_text(
                """---
blocker_lifecycle_contract: true
gate_ledger:
  - gate_id: duplicate
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence: []
    proof: []
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-finalize"
  - gate_id: duplicate
    owner_skill: kb-work
    gate_scope: implementation
    status: blocked
    allowed_next_action: "kb-repair"
---
""",
                encoding="utf-8",
            )
            original = sys.argv
            stderr = io.StringIO()
            try:
                sys.argv = [
                    "check_gate_ledger.py",
                    str(manifest),
                    "--gate",
                    "duplicate",
                ]
                with contextlib.redirect_stderr(stderr):
                    code = gate.main()
            finally:
                sys.argv = original
            self.assertEqual(2, code)
            self.assertIn("duplicate gate_id", stderr.getvalue())


if __name__ == "__main__":
    unittest.main()
