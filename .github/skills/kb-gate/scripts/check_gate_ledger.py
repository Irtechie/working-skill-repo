#!/usr/bin/env python3
"""Validate a KB manifest gate ledger before phase advancement."""

from __future__ import annotations

import argparse
from datetime import datetime, timezone
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

        if current is None:
            item_match = re.match(r"^-\s+([A-Za-z0-9_-]+):\s*(.*)$", stripped)
            if item_match:
                current = {item_match.group(1): unquote(item_match.group(2))}
                ledger.append(current)
                current_list = None
            continue

        list_item = re.match(r"^-\s+(.*)$", stripped)
        if list_item and current_list:
            current.setdefault(current_list, []).append(unquote(list_item.group(1)))
            continue

        item_match = re.match(r"^-\s+([A-Za-z0-9_-]+):\s*(.*)$", stripped)
        if item_match:
            current = {item_match.group(1): unquote(item_match.group(2))}
            ledger.append(current)
            current_list = None
            continue

        key_value = re.match(r"^([A-Za-z0-9_-]+):\s*(.*)$", stripped)
        if key_value:
            key = key_value.group(1)
            value = key_value.group(2)
            if value == "":
                current[key] = []
                current_list = key
            elif (inline := parse_inline_list(value)) is not None:
                current[key] = inline
                current_list = None
            else:
                current[key] = unquote(value)
                current_list = None

    return ledger


def parse_inline_list(value: str) -> list[str] | None:
    value = value.strip()
    if not (value.startswith("[") and value.endswith("]")):
        return None
    inner = value[1:-1].strip()
    if not inner:
        return []
    parts: list[str] = []
    start = 0
    quote = ""
    escaped = False
    for index, char in enumerate(inner):
        if escaped:
            escaped = False
            continue
        if char == "\\" and quote == '"':
            escaped = True
            continue
        if quote:
            if char == quote:
                quote = ""
            continue
        if char in {"'", '"'}:
            quote = char
            continue
        if char == ",":
            parts.append(inner[start:index])
            start = index + 1
    if quote or escaped:
        return None
    parts.append(inner[start:])
    return [unquote(item.strip()) for item in parts]


def blocker_lifecycle_contract_value(frontmatter: str) -> tuple[bool, bool]:
    for raw in frontmatter.splitlines():
        if raw.startswith((" ", "\t")):
            continue
        match = re.match(r"^blocker_lifecycle_contract:\s*(.*)$", raw.strip())
        if match:
            value = unquote(match.group(1)).lower()
            if value == "true":
                return True, True
            if value == "false":
                return False, True
            return False, False
    return False, True


def parse_gate_time(value: str) -> datetime | None:
    value = value.strip()
    try:
        if re.fullmatch(r"\d{4}-\d{2}-\d{2}", value):
            return datetime.strptime(value, "%Y-%m-%d").replace(tzinfo=timezone.utc)
        if not re.fullmatch(
            r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?(?:Z|[+-]\d{2}:\d{2})",
            value,
        ):
            return None
        normalized = value.replace("Z", "+00:00")
        fraction = re.search(r"\.(\d{7,9})(?=Z|[+-]\d{2}:\d{2}$)", value)
        if fraction:
            normalized = normalized.replace(f".{fraction.group(1)}", f".{fraction.group(1)[:6]}", 1)
        parsed = datetime.fromisoformat(normalized)
        if parsed.tzinfo is None:
            return None
        return parsed.astimezone(timezone.utc)
    except ValueError:
        return None


def parse_gate_timestamp(value: str) -> datetime | None:
    value = value.strip()
    if re.fullmatch(r"\d{4}-\d{2}-\d{2}", value):
        return None
    return parse_gate_time(value)


def validate_blocker_lifecycle_gate(gate: dict, now: datetime) -> list[str]:
    issues: list[str] = []
    gate_id = str(gate.get("gate_id") or "")
    status = str(gate.get("status") or "")
    scope = str(gate.get("gate_scope") or "")
    responsibility = str(gate.get("responsibility") or "")
    propagation = str(gate.get("propagation") or "")

    def require(key: str) -> None:
        if not gate.get(key):
            issues.append(f"{gate_id or '<missing>'}: missing {key}")

    require("gate_id")
    require("owner_skill")
    require("gate_scope")
    require("allowed_next_action")
    if scope not in {"implementation", "integration", "release", "deployment", "optional-capability"}:
        issues.append(f"{gate_id}: invalid gate_scope {scope!r}")
    if status not in {"pending", "blocked", "needs-human", "quarantined", "passed"}:
        issues.append(
            f"{gate_id}: invalid status {status!r}; pause is execution control state, not a gate result"
        )

    if status in {"blocked", "needs-human"}:
        for key in (
            "blockers",
            "attempted",
            "affected_scope",
            "resume_condition",
            "recheck",
            "checked_at",
            "propagation",
        ):
            require(key)
        if status == "needs-human" and responsibility != "human":
            issues.append(f"{gate_id}: needs-human requires responsibility 'human'")
        if status == "blocked" and responsibility not in {"agent", "external", "dependency"}:
            issues.append(
                f"{gate_id}: blocked requires responsibility agent, external, or dependency"
            )
        if propagation not in {"current-gate-only", "dependent-slices-only"}:
            issues.append(f"{gate_id}: invalid propagation {propagation!r}")
        if scope in {"release", "deployment", "optional-capability"} and propagation != "current-gate-only":
            issues.append(f"{gate_id}: release/deployment/optional blocker is over-propagated")
        checked = parse_gate_timestamp(str(gate.get("checked_at") or ""))
        if checked is None:
            issues.append(f"{gate_id}: checked_at must be RFC3339 with timezone")
        else:
            age = now - checked
            if age.total_seconds() > 72 * 3600:
                issues.append(f"{gate_id}: blocker is stale; rerun {gate.get('recheck')!r}")
            if age.total_seconds() < -300:
                issues.append(f"{gate_id}: checked_at is in the future")

    if status in {"passed", "quarantined"}:
        if parse_gate_time(str(gate.get("passed_at") or "")) is None:
            issues.append(f"{gate_id}: passed_at must be YYYY-MM-DD or RFC3339")
        stale_keys = (
            "responsibility",
            "affected_scope",
            "resume_condition",
            "recheck",
            "checked_at",
            "propagation",
            "attempted",
        )
        if any(gate.get(key) for key in stale_keys):
            issues.append(f"{gate_id}: advanceable gate retains stale blocker metadata")
        if status == "quarantined":
            for key in (
                "quarantined_scope",
                "quarantine_owner",
                "quarantine_evidence",
                "forbidden_claims",
            ):
                require(key)
        elif any(
            gate.get(key)
            for key in (
                "quarantined_scope",
                "quarantine_owner",
                "quarantine_evidence",
                "forbidden_claims",
            )
        ):
            issues.append(f"{gate_id}: passed gate retains stale quarantine metadata")
    elif gate.get("passed_at"):
        issues.append(f"{gate_id}: nonadvanceable gate must not retain passed_at")
    return issues


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
    frontmatter = load_frontmatter(manifest)
    ledger = parse_gate_ledger(frontmatter)
    if not ledger:
        print(f"FAIL: {manifest} has no gate_ledger list", file=sys.stderr)
        return 2

    matches = [
        item
        for item in ledger
        if isinstance(item, dict) and item.get("gate_id") == args.gate
    ]
    if len(matches) > 1:
        print(f"FAIL: duplicate gate_id {args.gate!r}", file=sys.stderr)
        return 2
    gate = matches[0] if matches else None
    if not gate:
        print(f"FAIL: gate {args.gate!r} not found", file=sys.stderr)
        return 2

    lifecycle_enabled, lifecycle_valid = blocker_lifecycle_contract_value(frontmatter)
    if not lifecycle_valid:
        print(
            "FAIL: blocker_lifecycle_contract must be true or false",
            file=sys.stderr,
        )
        return 10
    if lifecycle_enabled:
        lifecycle_issues = validate_blocker_lifecycle_gate(gate, datetime.now(timezone.utc))
        if lifecycle_issues:
            for issue in lifecycle_issues:
                print(f"FAIL: {issue}", file=sys.stderr)
            return 10

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
