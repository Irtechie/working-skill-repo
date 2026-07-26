#!/usr/bin/env python3
"""Validate a KB manifest gate ledger before phase advancement."""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

def unquote(value: str) -> str:
    value = value.strip()
    if len(value) >= 2 and value[0] == value[-1] and value[0] in {"'", '"'}:
        return value[1:-1]
    return value


def load_frontmatter(path: Path) -> str:
    text = path.read_text(encoding="utf-8")
    match = re.match(r"^---\s*\n(.*?)\n---\s*\n", text, re.S)
    if not match:
        raise ValueError(f"{path} has no YAML frontmatter")
    return match.group(1)


def parse_gate_ledger(frontmatter: str) -> list[dict]:
    """Parse the small YAML subset used by KB manifest gate_ledger entries."""
    lines = frontmatter.splitlines()
    try:
        start = next(i for i, line in enumerate(lines) if line.strip() == "gate_ledger:")
    except StopIteration:
        return []

    ledger: list[dict] = []
    current: dict | None = None
    current_list: str | None = None

    for raw in lines[start + 1 :]:
        if raw and not raw.startswith((" ", "\t")) and re.match(r"^[A-Za-z0-9_-]+:", raw):
            break
        stripped = raw.strip()
        if not stripped or stripped.startswith("#"):
            continue

        item_match = re.match(r"^-\s+([A-Za-z0-9_-]+):\s*(.*)$", stripped)
        if item_match:
            current = {item_match.group(1): unquote(item_match.group(2))}
            ledger.append(current)
            current_list = None
            continue

        if current is None:
            continue

        list_item = re.match(r"^-\s+(.*)$", stripped)
        if list_item and current_list:
            current.setdefault(current_list, []).append(unquote(list_item.group(1)))
            continue

        key_value = re.match(r"^([A-Za-z0-9_-]+):\s*(.*)$", stripped)
        if key_value:
            key = key_value.group(1)
            value = key_value.group(2)
            if value == "":
                current[key] = []
                current_list = key
            elif value == "[]":
                current[key] = []
                current_list = None
            else:
                current[key] = unquote(value)
                current_list = None

    return ledger


_EXT_RE = re.compile(r"\.(md|json|jsonl|txt|log|png|html|ps1|py|yaml|yml|mjs|sh)$", re.I)
_SEP_RE = re.compile(r"[\\/]")
_RUNNER_RE = re.compile(
    r"(?:^|\s)(?:python[\d.]*|pytest|node|npm|npx|go|powershell(?:\.exe)?|pwsh|bash|sh|git|"
    r"kubectl|docker|podman|buildah|ctr|curl|make|cargo|dotnet)(?:\s|$)",
    re.I,
)
_ENV_ASSIGN_RE = re.compile(r"(?:^|\s)[A-Z_][A-Z0-9_]*=\S")
_FLAG_RE = re.compile(r"(?:^|\s)-{1,2}[A-Za-z]")


def looks_like_command(value: str) -> bool:
    """True when the value records a command invocation rather than a file reference.

    Paths inside a recorded command are arguments evaluated in that command's
    own working directory, which is frequently a different repository. Treating
    them as assertions about this repo produces false failures.
    """
    return bool(
        _ENV_ASSIGN_RE.search(value)
        or _RUNNER_RE.search(value)
        or _FLAG_RE.search(value)
    )


def looks_like_path(value: str) -> bool:
    """True when the whole value is a single path-like token.

    Prose is never a path, even when it contains slashes. Values such as
    "origin/main is ancestor of codex/topic" describe a git ref, not a file,
    and must not be resolved against the filesystem.
    """
    stripped = value.strip()
    if not stripped or re.search(r"\s", stripped):
        return False
    return bool(_SEP_RE.search(stripped) or _EXT_RE.search(stripped))


def prose_path_tokens(value: str) -> list[str]:
    """Extract unambiguous repo-relative file references embedded in prose.

    Only tokens carrying both a separator and a known file extension qualify,
    and only when the surrounding text is not a recorded command. This keeps
    typo detection for real references like `docs/plans/foo.md` while leaving
    git refs, digests, and command arguments alone.
    """
    if looks_like_command(value):
        return []
    out = []
    for tok in re.findall(r"[^\s`'\"(),;]+", value):
        tok = tok.rstrip(".,;:")
        if _SEP_RE.search(tok) and _EXT_RE.search(tok):
            out.append(tok)
    return out


def proof_path_exists(manifest: Path, proof_item: str) -> bool:
    item_path = Path(proof_item)
    if item_path.is_absolute():
        return item_path.exists()
    return (Path.cwd() / item_path).exists() or (manifest.parent / item_path).exists()


def main() -> int:
    parser = argparse.ArgumentParser(description="Check a KB gate_ledger gate.")
    parser.add_argument("manifest", type=Path)
    parser.add_argument("--gate", required=True, help="gate_id to validate")
    parser.add_argument("--allowed-next", default="", help="Expected allowed_next_action")
    parser.add_argument(
        "--allow-quarantine",
        action="store_true",
        help="Accept status=quarantined as advanceable",
    )
    args = parser.parse_args()

    manifest = args.manifest.resolve()
    ledger = parse_gate_ledger(load_frontmatter(manifest))
    if not ledger:
        print(f"FAIL: {manifest} has no gate_ledger list", file=sys.stderr)
        return 2

    gate = next((item for item in ledger if isinstance(item, dict) and item.get("gate_id") == args.gate), None)
    if not gate:
        print(f"FAIL: gate {args.gate!r} not found", file=sys.stderr)
        return 2

    status = gate.get("status")
    advanceable = {"passed"}
    if args.allow_quarantine:
        advanceable.add("quarantined")
    if status not in advanceable:
        print(f"FAIL: gate {args.gate} status is {status!r}, expected {sorted(advanceable)}", file=sys.stderr)
        return 3

    required = gate.get("required_evidence") or []
    proof = gate.get("proof") or []
    blockers = gate.get("blockers") or []
    if not isinstance(required, list) or not isinstance(proof, list) or not isinstance(blockers, list):
        print("FAIL: required_evidence, proof, and blockers must be lists", file=sys.stderr)
        return 4
    if len(proof) < len(required):
        print(f"FAIL: gate {args.gate} has {len(required)} required evidence items but only {len(proof)} proof items", file=sys.stderr)
        return 5
    if blockers:
        print(f"FAIL: gate {args.gate} still has blockers: {blockers}", file=sys.stderr)
        return 6
    if not gate.get("passed_at"):
        print(f"FAIL: gate {args.gate} is advanceable but has no passed_at", file=sys.stderr)
        return 7
    if args.allowed_next and gate.get("allowed_next_action") != args.allowed_next:
        print(
            f"FAIL: gate {args.gate} allowed_next_action is {gate.get('allowed_next_action')!r}, expected {args.allowed_next!r}",
            file=sys.stderr,
        )
        return 8

    missing_paths: list[str] = []
    for item in proof:
        if not isinstance(item, str):
            continue
        candidates = [item.strip()] if looks_like_path(item) else prose_path_tokens(item)
        for candidate in candidates:
            if not proof_path_exists(manifest, candidate):
                missing_paths.append(candidate)
    if missing_paths:
        print(f"FAIL: proof paths do not exist: {missing_paths}", file=sys.stderr)
        return 9

    print(
        f"PASS: gate={args.gate} status={status} required={len(required)} proof={len(proof)} allowed_next={gate.get('allowed_next_action')}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
