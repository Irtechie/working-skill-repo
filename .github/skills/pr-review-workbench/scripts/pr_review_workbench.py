#!/usr/bin/env python3
"""Build a bounded PR evidence packet and an offline review workbench."""

from __future__ import annotations

import argparse
import hashlib
import html
import json
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Callable
from urllib.parse import urlparse


DATASET_STATES = {"complete", "partial", "forbidden", "unsupported", "stale"}
REQUIRED_DATASETS = {"metadata", "files", "checks", "reviews", "source"}
REVIEW_EVENTS = {"APPROVE", "COMMENT", "REQUEST_CHANGES"}
PASSING_CHECKS = {"SUCCESS", "NEUTRAL", "SKIPPED"}


def _safe_url(value: str) -> str:
    parsed = urlparse(value or "")
    return value if parsed.scheme == "https" and parsed.netloc else "#"


def _run(command: list[str], *, timeout: int = 30) -> subprocess.CompletedProcess[str]:
    return subprocess.run(command, text=True, capture_output=True, check=False, timeout=timeout)


def _dataset_blockers(datasets: dict[str, str]) -> list[str]:
    blockers = []
    for name in sorted(REQUIRED_DATASETS - set(datasets)):
        blockers.append(f"{name}: partial")
    for name, state in sorted(datasets.items()):
        if state not in DATASET_STATES:
            raise ValueError(f"invalid dataset state for {name}: {state}")
        if state != "complete":
            blockers.append(f"{name}: {state}")
    return blockers


def _normalize_pr(raw: dict[str, Any], repository: dict[str, str]) -> dict[str, Any]:
    datasets = dict(raw.get("datasets") or {})
    blockers = _dataset_blockers(datasets)
    start_sha = str(raw.get("start_head_sha") or raw.get("head_sha") or "")
    end_sha = str(raw.get("end_head_sha") or start_sha)
    packet_state = "complete"
    if not start_sha or start_sha != end_sha:
        packet_state = "stale"
        blockers.append("head SHA changed during collection")
    for check in raw.get("checks") or []:
        state = str(check.get("state") or "UNKNOWN").upper()
        if state not in PASSING_CHECKS:
            blockers.append(f"check {check.get('name', 'unknown')}: {state.lower()}")
    files = list(raw.get("files") or [])
    additions = sum(int(item.get("additions") or 0) for item in files)
    deletions = sum(int(item.get("deletions") or 0) for item in files)
    return {
        "repository": repository["nameWithOwner"],
        "number": int(raw["number"]),
        "title": str(raw.get("title") or "Untitled pull request"),
        "url": _safe_url(str(raw.get("url") or "")),
        "author": str(raw.get("author") or "unknown"),
        "updated_at": str(raw.get("updated_at") or ""),
        "base_ref": str(raw.get("base_ref") or ""),
        "head_ref": str(raw.get("head_ref") or ""),
        "head_sha": end_sha,
        "packet_state": packet_state,
        "datasets": datasets,
        "decision_state": "ready for human decision" if not blockers else "not ready",
        "blockers": blockers,
        "files": files,
        "checks": list(raw.get("checks") or []),
        "reviews": list(raw.get("reviews") or []),
        "summary": str(raw.get("summary") or "No authored summary was provided."),
        "behavioral_changes": list(raw.get("behavioral_changes") or []),
        "impact_analysis": dict(raw.get("impact_analysis") or {}),
        "change_size": {"files": len(files), "additions": additions, "deletions": deletions},
    }


def normalize_repository(raw: dict[str, Any]) -> dict[str, Any]:
    repository = dict(raw.get("repository") or {})
    if not repository.get("nameWithOwner"):
        raise ValueError("repository.nameWithOwner is required")
    items = [_normalize_pr(item, repository) for item in raw.get("pull_requests") or []]
    items.sort(key=lambda item: (item["decision_state"] != "not ready", item["updated_at"], item["number"]))
    return {
        "schema_version": 1,
        "generated_at": datetime.now(timezone.utc).replace(microsecond=0).isoformat(),
        "repository": {"nameWithOwner": repository["nameWithOwner"], "url": _safe_url(str(repository.get("url") or ""))},
        "pull_requests": items,
    }


def collect_fixture(path: Path | str) -> dict[str, Any]:
    return normalize_repository(json.loads(Path(path).read_text(encoding="utf-8")))


def collect_github(repository: str, *, limit: int = 20, runner: Callable[..., Any] = _run) -> dict[str, Any]:
    """Collect PRs in bounded, separately observable GitHub CLI queries."""
    listed = runner([
        "gh", "pr", "list", "--repo", repository, "--state", "open", "--limit", str(limit),
        "--json", "number,title,url,author,updatedAt,baseRefName,headRefName,headRefOid",
    ], timeout=30)
    if listed.returncode != 0:
        raise RuntimeError(f"GitHub PR list failed: {listed.stderr.strip()}")
    raw_prs = json.loads(listed.stdout)
    collected = []
    for item in raw_prs:
        number = int(item["number"])
        try:
            details = runner([
                "gh", "pr", "view", str(number), "--repo", repository,
                "--json", "files,statusCheckRollup,reviews,headRefOid",
            ], timeout=30)
        except (subprocess.TimeoutExpired, json.JSONDecodeError):
            details = subprocess.CompletedProcess([], 1, "", "detail query timed out or returned invalid JSON")
        datasets = {"metadata": "complete", "files": "partial", "checks": "partial", "reviews": "partial", "source": "unsupported"}
        detail_data: dict[str, Any] = {}
        if details.returncode == 0:
            try:
                detail_data = json.loads(details.stdout)
            except json.JSONDecodeError:
                detail_data = {}
            else:
                datasets.update({"files": "complete", "checks": "complete", "reviews": "complete"})
        author = item.get("author") or {}
        collected.append({
            "number": number,
            "title": item.get("title"),
            "url": item.get("url"),
            "author": author.get("login") if isinstance(author, dict) else author,
            "updated_at": item.get("updatedAt"),
            "base_ref": item.get("baseRefName"),
            "head_ref": item.get("headRefName"),
            "start_head_sha": item.get("headRefOid"),
            "end_head_sha": detail_data.get("headRefOid", item.get("headRefOid")),
            "datasets": datasets,
            "files": detail_data.get("files", []),
            "checks": [{"name": c.get("name") or c.get("context"), "state": c.get("conclusion") or c.get("state"), "url": c.get("detailsUrl", "")} for c in detail_data.get("statusCheckRollup", [])],
            "reviews": detail_data.get("reviews", []),
            "summary": "Open the evidence drill-down for source-level review.",
            "behavioral_changes": [],
        })
    return normalize_repository({"repository": {"nameWithOwner": repository, "url": f"https://github.com/{repository}"}, "pull_requests": collected})


def materialization_preview(repository_url: str, head_sha: str, destination: str) -> dict[str, Any]:
    if len(head_sha) != 40 or any(ch not in "0123456789abcdefABCDEF" for ch in head_sha):
        raise ValueError("head SHA must be a 40-character hexadecimal commit")
    if _safe_url(repository_url) == "#" or any(char in repository_url + destination for char in ('"', "'", "\r", "\n")):
        raise ValueError("repository URL must be HTTPS and URL/destination cannot contain quotes or newlines")
    prefix = ["git", "-c", "core.hooksPath=NUL", "-c", "credential.helper="]
    return {
        "credential_policy": "none",
        "environment": {"GIT_TERMINAL_PROMPT": "0", "GIT_LFS_SKIP_SMUDGE": "1"},
        "commands": [
            prefix + ["clone", "--bare", "--filter=blob:none", "--no-tags", repository_url, destination],
            prefix + ["-C", destination, "fetch", "--depth=1", "origin", head_sha],
            prefix + ["-C", destination, "rev-parse", "FETCH_HEAD"],
            prefix + ["-C", destination, "ls-tree", "-r", "--name-only", head_sha],
        ],
        "expected_head_sha": head_sha,
        "inspection": "Read individual blobs with git show SHA:path; never checkout a worktree.",
        "forbidden": ["checkout", "hooks", "filters", "LFS smudge", "submodules", "builds", "tests", "installs"],
    }


def _e(value: Any) -> str:
    return html.escape(str(value), quote=True)


AREA_ORDER = ("source", "verification", "docs", "delivery", "other")
AREA_META = {
    "source": ("Product & source", "Runtime behavior and user-facing implementation."),
    "verification": ("Verification", "Tests, fixtures, checks, and executable proof."),
    "docs": ("Docs & design", "Intent, decisions, guidance, and review context."),
    "delivery": ("Delivery & config", "Automation, packaging, configuration, and release surfaces."),
    "other": ("Other changes", "Files that need manual classification."),
}


def _file_area(path: str) -> str:
    value = path.replace("\\", "/").lower()
    name = value.rsplit("/", 1)[-1]
    if (
        value.startswith(("tests/", "test/", "spec/", "evals/"))
        or "/tests/" in value
        or name.startswith(("test_", "test-"))
        or "_test." in name
        or ".spec." in name
        or ".test." in name
    ):
        return "verification"
    if value.startswith(("docs/", "doc/")) or name.startswith(("readme", "changelog")) or value.endswith((".md", ".mdx", ".rst")):
        return "docs"
    if value.startswith((".github/", "scripts/", "config/", "deploy/", "infra/")) or name in {
        "dockerfile", "makefile", "package.json", "package-lock.json", "go.mod", "go.sum", "cargo.toml",
    }:
        return "delivery"
    if value.endswith((
        ".py", ".go", ".js", ".mjs", ".cjs", ".ts", ".tsx", ".jsx", ".rs", ".java",
        ".cs", ".cpp", ".c", ".h", ".rb", ".php", ".swift", ".kt", ".html", ".css",
    )):
        return "source"
    return "other"


def _change_areas(item: dict[str, Any]) -> list[dict[str, Any]]:
    grouped: dict[str, dict[str, Any]] = {}
    for changed in item["files"]:
        key = _file_area(str(changed.get("path") or ""))
        label, detail = AREA_META[key]
        area = grouped.setdefault(key, {
            "key": key, "label": label, "detail": detail, "files": [], "additions": 0, "deletions": 0,
            "level": "unverified", "reason": "Path-derived grouping only.", "affected_surfaces": [],
            "source_backed": False,
        })
        area["files"].append(changed)
        area["additions"] += int(changed.get("additions") or 0)
        area["deletions"] += int(changed.get("deletions") or 0)
    result = [grouped[key] for key in AREA_ORDER if key in grouped]
    for rank, area in enumerate(result, 1):
        area["rank"] = rank
    return result


def _ordered_impact_areas(item: dict[str, Any]) -> tuple[list[dict[str, Any]], dict[str, str]]:
    analysis = dict(item.get("impact_analysis") or {})
    raw_areas = list(analysis.get("areas") or [])
    revision = str(analysis.get("revision") or "")
    state = str(analysis.get("state") or "unsupported")
    if state != "complete" or revision != item["head_sha"] or not raw_areas:
        reason = "No complete impact packet was supplied."
        if revision and revision != item["head_sha"]:
            reason = "Impact packet revision does not match the reviewed head."
        return _change_areas(item), {
            "state": "fallback",
            "method": "file-role grouping",
            "reason": reason,
        }

    changed_by_path = {str(changed.get("path") or ""): changed for changed in item["files"]}
    areas: list[dict[str, Any]] = []
    covered: set[str] = set()
    seen_ranks: set[int] = set()
    for raw in raw_areas:
        rank = int(raw.get("rank") or 0)
        if rank < 1 or rank in seen_ranks:
            raise ValueError("impact area ranks must be unique positive integers")
        seen_ranks.add(rank)
        paths = [str(path) for path in raw.get("changed_files") or []]
        matched = [changed_by_path[path] for path in paths if path in changed_by_path]
        covered.update(path for path in paths if path in changed_by_path)
        anchors = list(raw.get("anchors") or [])
        areas.append({
            "key": "impact",
            "rank": rank,
            "label": str(raw.get("label") or f"Impact area {rank}"),
            "detail": str(raw.get("app_effect") or raw.get("reason") or "No application effect was stated."),
            "reason": str(raw.get("reason") or "Source-backed impact area."),
            "level": str(raw.get("level") or "medium").lower(),
            "affected_surfaces": [str(value) for value in raw.get("affected_surfaces") or []],
            "anchors": anchors,
            "files": matched,
            "additions": sum(int(changed.get("additions") or 0) for changed in matched),
            "deletions": sum(int(changed.get("deletions") or 0) for changed in matched),
            "source_backed": True,
        })
    areas.sort(key=lambda area: area["rank"])

    remaining = [changed for path, changed in changed_by_path.items() if path not in covered]
    if remaining:
        areas.append({
            "key": "supporting",
            "rank": max(seen_ranks, default=0) + 1,
            "label": "Supporting and mechanical changes",
            "detail": "Changed files outside the source-backed application-impact path.",
            "reason": "Placed last because no direct application impact was asserted.",
            "level": "low",
            "affected_surfaces": [],
            "anchors": [],
            "files": remaining,
            "additions": sum(int(changed.get("additions") or 0) for changed in remaining),
            "deletions": sum(int(changed.get("deletions") or 0) for changed in remaining),
            "source_backed": True,
        })
    return areas, {
        "state": "complete",
        "method": str(analysis.get("method") or "source and dependency inspection"),
        "reason": str(analysis.get("reason") or "Pinned source and downstream relationships."),
    }


def _flow_node(
    position: str,
    kind: str,
    eyebrow: str,
    title: str,
    detail: str,
    evidence: str,
    caveat: str,
) -> str:
    return (
        f'<button class="flow-node {position} {kind}" type="button" '
        f'data-inspector-title="{_e(title)}" data-inspector-body="{_e(detail)}" '
        f'data-inspector-evidence="{_e(evidence)}" data-inspector-caveat="{_e(caveat)}">'
        f'<span class="node-kicker">{_e(eyebrow)}</span><strong>{_e(title)}</strong>'
        f'<span>{_e(detail)}</span></button>'
    )


def _impact_spine_node(areas: list[dict[str, Any]], impact_basis: dict[str, str]) -> str:
    visible = areas[:4]
    items = "".join(
        f'<li><span>{int(area["rank"]):02d}</span><b>{_e(area["label"])}</b>'
        f'<em>{_e(area["level"])}</em></li>'
        for area in visible
    )
    detail = " → ".join(area["label"] for area in visible) or "No impact areas"
    anchor_count = sum(len(area.get("anchors") or []) for area in visible)
    qualifier = "Source-backed order" if impact_basis["state"] == "complete" else "Labeled fallback"
    return (
        '<button class="flow-node node-areas artifact impact-spine" type="button" '
        f'data-inspector-title="Application impact order" data-inspector-body="{_e(detail)}" '
        f'data-inspector-evidence="{_e(f"{qualifier}; {anchor_count} source/dependency anchors")}" '
        f'data-inspector-caveat="{_e(impact_basis["reason"])}">'
        f'<span class="node-kicker">{_e(qualifier)}</span><strong>Application impact</strong>'
        f'<ol>{items}</ol></button>'
    )


def _render_topology(item: dict[str, Any], areas: list[dict[str, Any]], impact_basis: dict[str, str]) -> str:
    datasets_complete = sum(1 for state in item["datasets"].values() if state == "complete")
    claims = item["behavioral_changes"]
    state_ready = item["decision_state"] == "ready for human decision"
    terminal_kind = "success" if state_ready else "blocked"
    terminal_title = "Human decision" if state_ready else "Repair required"
    terminal_detail = "Evidence is bounded and ready to assess." if state_ready else "Evidence gaps block a responsible decision."
    active_path = "ready-path" if state_ready else "blocked-path"
    return f'''
<div class="topology-frame" aria-label="Pull request decision topology">
  <div class="topology-heading"><div><span class="eyebrow">Decision topology</span>
    <h2>How this change reaches a human decision</h2></div>
    <span class="topology-legend"><i></i> Click any node for evidence</span></div>
  <div class="topology-scroll"><div class="flow-canvas {active_path}">
    <svg class="flow-lines" viewBox="0 0 1200 500" role="img" aria-label="Review request flows through changed areas, behavioral impact, and an evidence gate to a human decision or repair state">
      <defs><marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z"></path></marker></defs>
      <path class="edge edge-main" d="M210 245 H280"></path><text x="245" y="226">pin SHA</text>
      <path class="edge edge-main" d="M530 235 H580"></path><text x="555" y="216">trace</text>
      <path class="edge edge-main" d="M800 240 H858"></path><text x="829" y="221">prove</text>
      <path class="edge edge-pass" d="M997 230 H1042"></path><text x="1019" y="211">Pass</text>
      <path class="edge edge-fail" d="M930 315 V407 H1042"></path><text x="950" y="389">Fail</text>
    </svg>
    {_flow_node("node-trigger", "process", "Trigger", f"PR #{item['number']} requested", f"{item['head_ref']} → {item['base_ref']}", f"Immutable head {item['head_sha'][:12]}", "A changed head invalidates this workbench.")}
    {_impact_spine_node(areas, impact_basis)}
    {_flow_node("node-impact", "process", "Understand", f"{len(claims)} behavioral claims", item["summary"], f"{sum(len(claim.get('anchors') or []) for claim in claims)} source anchors", "Claims without anchors remain visibly unsupported.")}
    {_flow_node("node-gate", "gate", "Gate", "Evidence sufficient?", f"{datasets_complete}/{len(REQUIRED_DATASETS)} required datasets complete", f"{len(item['checks'])} checks and {len(item['reviews'])} reviews captured", "Passing checks prove only their covered behavior.")}
    {_flow_node("node-ready", terminal_kind, "Terminal", terminal_title, terminal_detail, item["decision_state"], "The workbench informs authority; it never approves or merges.")}
    {_flow_node("node-repair", "blocked", "Fail branch", "Evidence gaps", f"{len(item['blockers'])} blocker(s) or unsupported inputs", "; ".join(item["blockers"]) or "No active gaps", "Repair or refresh evidence before returning to the gate.")}
  </div></div>
</div>'''


def _render_review_path(item: dict[str, Any], areas: list[dict[str, Any]], impact_basis: dict[str, str]) -> str:
    checks_passing = sum(1 for check in item["checks"] if str(check.get("state", "")).upper() in PASSING_CHECKS)
    blocker_summary = "; ".join(item["blockers"]) or "No blocking evidence gaps detected."
    steps = [
        ("Pin the review", "Start", f"PR #{item['number']}", f"Freeze head {item['head_sha'][:12]}", "A moving head makes every later claim stale."),
        ("Follow app impact", "Impact", f"{len(areas)} ranked impact areas", impact_basis["method"], impact_basis["reason"]),
        ("Trace behavior", "Impact", f"{len(item['behavioral_changes'])} source-anchored claims", "Connect intent to affected source", "Do not infer behavior from filenames alone."),
        ("Test the evidence", "Proof", f"{checks_passing}/{len(item['checks'])} checks passing", f"{item['packet_state']} packet", blocker_summary),
        ("Make the decision", "Human gate", item["decision_state"], "Open the exact GitHub PR", "Ready to assess is not automatic approval."),
    ]
    rail = "".join(
        f'<button type="button" class="path-step{" active" if index == 1 else ""}" data-step="{index}">'
        f'<span>{index:02d}</span><b>{_e(step[0])}</b><small>{_e(step[1])}</small></button>'
        for index, step in enumerate(steps, 1)
    )
    panels = []
    for index, step in enumerate(steps, 1):
        previous = steps[index - 2][0] if index > 1 else "Review request"
        following = steps[index][0] if index < len(steps) else "Human-owned outcome"
        panels.append(f'''
<article class="path-panel" data-step-panel="{index}"{" " if index == 1 else " hidden"}>
  <span class="eyebrow">Step {index} of {len(steps)} · {_e(step[1])}</span>
  <h2>{_e(step[0])}</h2><p class="path-lead">{_e(step[4])}</p>
  <div class="mini-flow" aria-label="{_e(previous)} then {_e(step[0])} then {_e(following)}">
    <div><small>Before</small><b>{_e(previous)}</b></div><span>→</span>
    <div class="current"><small>Now</small><b>{_e(step[2])}</b><em>{_e(step[3])}</em></div><span>→</span>
    <div><small>Next</small><b>{_e(following)}</b></div>
  </div>
  <div class="path-actions"><button type="button" data-step-prev>Previous</button><button type="button" class="primary" data-step-next>Next step</button></div>
</article>''')
    return f'<div class="path-layout"><nav class="path-rail" aria-label="Guided review steps">{rail}</nav><div class="path-stage">{"".join(panels)}</div></div>'


def _render_changes(areas: list[dict[str, Any]], impact_basis: dict[str, str]) -> str:
    max_churn = max((area["additions"] + area["deletions"] for area in areas), default=1)
    rendered = []
    for area in areas:
        churn = area["additions"] + area["deletions"]
        share = max(4, round(churn / max_churn * 100))
        files = "".join(
            f'<li><code>{_e(changed.get("path"))}</code><span>+{int(changed.get("additions") or 0)} / -{int(changed.get("deletions") or 0)}</span></li>'
            for changed in area["files"]
        )
        affected = ", ".join(area.get("affected_surfaces") or []) or "No downstream surface asserted."
        rendered.append(f'''
<section class="area-row">
  <div class="impact-rank">{int(area['rank']):02d}</div>
  <div class="area-copy"><span class="area-mark {area['key']}"></span><div><div class="area-title"><h3>{_e(area['label'])}</h3><b class="impact-level {area['level']}">{_e(area['level'])}</b></div>
  <p>{_e(area['detail'])}</p><small>{_e(area['reason'])}</small><small>Affects: {_e(affected)}</small></div></div>
  <div class="area-measure"><div class="area-bar"><i style="width:{share}%"></i></div><b>{len(area['files'])} files</b><span>+{area['additions']} / -{area['deletions']}</span></div>
  <details><summary>Inspect files</summary><ul class="file-list">{files}</ul></details>
</section>''')
    status = "Source-backed impact order" if impact_basis["state"] == "complete" else "Fallback order — not impact analysis"
    return (
        f'<div class="section-heading"><span class="eyebrow">{_e(status)}</span>'
        f'<h2>Review by effect on the application</h2><p>{_e(impact_basis["method"])}. '
        f'{_e(impact_basis["reason"])}</p></div>' + "".join(rendered)
    )


def _render_evidence(item: dict[str, Any]) -> str:
    dataset_rows = "".join(
        f'<li><span>{_e(name)}</span><b class="proof-state {state}">{_e(state)}</b></li>'
        for name, state in sorted(item["datasets"].items())
    )
    check_rows = "".join(
        f'<li><a href="{_e(_safe_url(str(check.get("url") or "")))}">{_e(check.get("name") or "Unnamed check")}</a>'
        f'<b class="proof-state {_e(str(check.get("state") or "unknown").lower())}">{_e(check.get("state") or "unknown")}</b></li>'
        for check in item["checks"]
    ) or '<li><span>No checks reported</span><b class="proof-state unsupported">unsupported</b></li>'
    claims = []
    for index, claim in enumerate(item["behavioral_changes"], 1):
        anchors = "".join(
            f'<a href="{_e(_safe_url(str(anchor.get("url") or "")))}"><code>{_e(anchor.get("path"))}:{_e(anchor.get("line"))}</code></a>'
            for anchor in claim.get("anchors") or []
        ) or '<span class="proof-state unsupported">No source anchor</span>'
        claims.append(f'''
<article class="evidence-claim"><span class="claim-number">{index:02d}</span><div>
  <h3>{_e(claim.get("claim"))}</h3><p>{_e(claim.get("impact"))}</p>
  <div class="anchor-row"><b class="proof-state {_e(claim.get("proof_state", "unsupported"))}">{_e(claim.get("proof_state", "unsupported"))}</b>{anchors}</div>
</div></article>''')
    claims_html = "".join(claims) or '<p class="empty-state">No behavioral claim was asserted. Source inspection is still required.</p>'
    blockers = "".join(f"<li>{_e(blocker)}</li>" for blocker in item["blockers"]) or "<li>No blocking evidence gaps detected.</li>"
    return f'''
<div class="evidence-layout">
  <div><div class="section-heading"><span class="eyebrow">Behavioral evidence</span><h2>Claims must point back to source</h2></div>{claims_html}</div>
  <aside class="evidence-side"><section><h3>Dataset readiness</h3><ul class="status-list">{dataset_rows}</ul></section>
  <section><h3>Checks</h3><ul class="status-list">{check_rows}</ul></section>
  <section><h3>Gaps and blockers</h3><ul class="blocker-list">{blockers}</ul></section></aside>
</div>'''


def _render_pr_section(item: dict[str, Any], *, active: bool) -> str:
    areas, impact_basis = _ordered_impact_areas(item)
    checks_passing = sum(1 for check in item["checks"] if str(check.get("state", "")).upper() in PASSING_CHECKS)
    role = "role" if active else "kind"
    facts = [
        f"{item['change_size']['files']} files · +{item['change_size']['additions']} / -{item['change_size']['deletions']}",
        "Source-backed impact order" if impact_basis["state"] == "complete" else "Impact order unavailable",
        f"{len(item['behavioral_changes'])} behavioral claims",
        f"{checks_passing}/{len(item['checks'])} checks passing",
    ]
    fact_html = "".join(f'<li data-{role}="primary-fact">{_e(fact)}</li>' for fact in facts)
    hidden = "" if active else " hidden"
    state_class = "ready" if item["decision_state"] == "ready for human decision" else "blocked"
    return f'''<section class="pr-section" id="pr-{item['number']}" data-pr="{item['number']}"{hidden}>
<div class="pr-head"><div class="head-copy"><p class="breadcrumb">PR #{item['number']} · {_e(item['author'])} · <code>{_e(item['head_sha'][:12])}</code></p>
<div class="title-line"><h1>{_e(item['title'])}</h1><span class="decision-state {state_class}" data-{role}="decision-state">{_e(item['decision_state'])}</span></div>
<p class="summary">{_e(item['summary'])}</p><ul class="facts">{fact_html}</ul></div>
<a class="next-action" data-{role}="next-action" href="#pr-{item['number']}-view-workflow-step-1">Walk the review path <span>→</span></a></div>
<div class="view-tabs" role="tablist" aria-label="Review workbench views">
  <button type="button" aria-selected="true" data-view-tab="overview">Decision map</button>
  <button type="button" aria-selected="false" data-view-tab="workflow">Guided review</button>
  <button type="button" aria-selected="false" data-view-tab="changes">App impact</button>
  <button type="button" aria-selected="false" data-view-tab="evidence">Evidence</button>
</div>
<div class="view-panel" data-view-panel="overview">{_render_topology(item, areas, impact_basis)}
  <aside class="inspector" aria-live="polite"><span class="eyebrow">Selected evidence</span><h3 data-inspector-heading>Start with the topology</h3>
  <p data-inspector-body>Select a node to see what it means.</p><dl><dt>Evidence</dt><dd data-inspector-evidence>Commit-pinned review packet</dd>
  <dt>Caveat</dt><dd data-inspector-caveat>Ready to assess is not automatic approval.</dd></dl></aside>
</div>
<div class="view-panel" data-view-panel="workflow" hidden>{_render_review_path(item, areas, impact_basis)}</div>
<div class="view-panel" data-view-panel="changes" hidden>{_render_changes(areas, impact_basis)}</div>
<div class="view-panel" data-view-panel="evidence" hidden>{_render_evidence(item)}</div>
<footer class="pr-footer"><a href="{_e(item['url'])}">Open the original pull request</a><span>Reviewed head <code>{_e(item['head_sha'])}</code></span></footer>
</section>'''


def render_html(packet: dict[str, Any], *, selected_pr: int | None = None) -> str:
    if not packet.get("pull_requests"):
        raise ValueError("packet has no pull requests")
    if selected_pr is None:
        selected = packet["pull_requests"][0]
    else:
        selected = next((item for item in packet["pull_requests"] if item["number"] == selected_pr), None)
        if selected is None:
            raise ValueError(f"PR #{selected_pr} not found in packet")
    inbox = "".join(
        f'<a class="inbox-card" data-pr-card="{item["number"]}" href="#pr-{item["number"]}"><b>#{item["number"]}</b> {_e(item["title"])}<span>{_e(item["decision_state"])}</span></a>'
        for item in packet["pull_requests"]
    )
    sections = "".join(_render_pr_section(item, active=item["number"] == selected["number"]) for item in packet["pull_requests"])
    csp = _e("default-src 'none'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; img-src data:; base-uri 'none'; form-action 'none'")
    head = f'''<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="referrer" content="no-referrer"><meta http-equiv="Content-Security-Policy" content="{csp}">
<title>PR review workbench · {_e(packet['repository']['nameWithOwner'])} #{selected['number']}</title>'''
    styles = '''
<style>
:root{--bg:#08101d;--bg-soft:#0d1727;--panel:#111d30;--panel-2:#16243a;--text:#f4f7fb;--muted:#93a4bd;--line:#263854;--line-strong:#3b5378;--blue:#66a7ff;--blue-soft:#17345b;--amber:#ffd166;--amber-soft:#4a3b17;--red:#ff7c88;--red-soft:#4a2029;--green:#67d9a2;--green-soft:#173f33;--shadow:0 20px 60px #02060d80}
body[data-theme="light"]{--bg:#f3f6fb;--bg-soft:#e9eef6;--panel:#ffffff;--panel-2:#f7f9fc;--text:#172235;--muted:#5d6c82;--line:#d7dfeb;--line-strong:#aebbd0;--blue:#1768c5;--blue-soft:#dcecff;--amber:#946900;--amber-soft:#fff1c6;--red:#b93045;--red-soft:#ffe1e6;--green:#147a52;--green-soft:#d9f5e8;--shadow:0 18px 45px #34506b1f}
*{box-sizing:border-box}html{scroll-behavior:smooth}body{margin:0;background:radial-gradient(circle at 75% -10%,var(--blue-soft),transparent 34rem),var(--bg);color:var(--text);font:15px/1.5 Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}button,a{font:inherit}button{color:inherit}a{color:var(--blue)}code{font-family:"SFMono-Regular",Consolas,monospace;color:var(--green);font-size:.9em}button:focus-visible,a:focus-visible{outline:3px solid var(--blue);outline-offset:3px}
.app-bar{position:sticky;top:0;z-index:20;display:flex;justify-content:space-between;align-items:center;gap:18px;padding:13px 22px;background:color-mix(in srgb,var(--bg) 88%,transparent);backdrop-filter:blur(18px);border-bottom:1px solid var(--line)}.identity{display:flex;align-items:center;gap:12px;min-width:0}.identity-mark{display:grid;place-items:center;width:34px;height:34px;border-radius:10px;background:var(--blue);color:#07101e;font-weight:900}.identity b{display:block;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.identity small{display:block;color:var(--muted)}.theme-toggle{border:1px solid var(--line);background:var(--panel);padding:8px 12px;border-radius:999px;cursor:pointer}
.app-shell{display:grid;grid-template-columns:268px minmax(0,1fr);min-height:calc(100vh - 62px)}.inbox{padding:22px 16px;border-right:1px solid var(--line);background:color-mix(in srgb,var(--bg-soft) 82%,transparent)}.inbox-label,.eyebrow,.node-kicker{display:block;color:var(--muted);font-size:.72rem;font-weight:800;letter-spacing:.12em;text-transform:uppercase}.inbox-label{padding:0 10px 12px}.inbox-card{display:grid;grid-template-columns:auto 1fr;gap:2px 10px;padding:12px 11px;margin-bottom:7px;border:1px solid transparent;border-radius:12px;color:var(--text);text-decoration:none}.inbox-card:hover,.inbox-card.active{background:var(--panel);border-color:var(--line)}.inbox-card b{grid-row:1/3;color:var(--blue)}.inbox-card span{color:var(--muted);font-size:.82rem}.inbox-card strong{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.pr-section{max-width:1540px;margin:0 auto;padding:34px clamp(20px,4vw,54px) 60px}.pr-section[hidden],.view-panel[hidden],.path-panel[hidden]{display:none}.pr-head{display:flex;justify-content:space-between;align-items:end;gap:28px;padding-bottom:26px}.head-copy{max-width:980px}.breadcrumb{margin:0 0 10px;color:var(--muted)}.title-line{display:flex;align-items:center;gap:16px;flex-wrap:wrap}.title-line h1{margin:0;font-size:clamp(1.75rem,3.4vw,3.25rem);line-height:1.05;letter-spacing:-.045em}.decision-state{display:inline-flex;align-items:center;padding:7px 11px;border:1px solid;border-radius:999px;font-size:.78rem;font-weight:850;letter-spacing:.04em;text-transform:uppercase}.decision-state.ready{color:var(--green);background:var(--green-soft);border-color:var(--green)}.decision-state.blocked{color:var(--red);background:var(--red-soft);border-color:var(--red)}.summary{max-width:800px;margin:15px 0;color:var(--muted);font-size:1.05rem}.facts{display:flex;flex-wrap:wrap;gap:8px;margin:0;padding:0;list-style:none}.facts li{padding:6px 9px;border:1px solid var(--line);border-radius:8px;background:var(--panel);color:var(--muted);font-size:.82rem}.next-action{flex:0 0 auto;display:inline-flex;align-items:center;gap:12px;padding:12px 15px;border-radius:11px;background:var(--blue);color:#07101e;font-weight:850;text-decoration:none;box-shadow:0 12px 30px color-mix(in srgb,var(--blue) 30%,transparent)}.next-action span{font-size:1.25rem}
.view-tabs{position:sticky;top:61px;z-index:12;display:flex;gap:4px;padding:7px;border:1px solid var(--line);border-radius:13px;background:color-mix(in srgb,var(--panel) 92%,transparent);backdrop-filter:blur(16px);box-shadow:0 10px 30px #00000020}.view-tabs button{border:0;background:transparent;padding:9px 13px;border-radius:8px;color:var(--muted);font-weight:750;cursor:pointer}.view-tabs button[aria-selected="true"]{background:var(--blue-soft);color:var(--blue)}
.view-panel{padding-top:24px}.topology-frame{border:1px solid var(--line);border-radius:18px;background:linear-gradient(145deg,var(--panel),var(--bg-soft));box-shadow:var(--shadow);overflow:hidden}.topology-heading{display:flex;justify-content:space-between;align-items:end;gap:18px;padding:22px 24px;border-bottom:1px solid var(--line)}.topology-heading h2,.section-heading h2{margin:3px 0 0;font-size:1.35rem}.topology-legend{color:var(--muted);font-size:.82rem}.topology-legend i{display:inline-block;width:7px;height:7px;margin-right:6px;border-radius:50%;background:var(--blue);box-shadow:0 0 0 5px var(--blue-soft)}.topology-scroll{overflow-x:auto}.flow-canvas{position:relative;width:1200px;height:500px;background-image:linear-gradient(var(--line) 1px,transparent 1px),linear-gradient(90deg,var(--line) 1px,transparent 1px);background-size:32px 32px;background-position:-1px -1px}.flow-lines{position:absolute;inset:0;width:1200px;height:500px;pointer-events:none}.flow-lines .edge{fill:none;stroke:var(--line-strong);stroke-width:2;marker-end:url(#arrow)}.flow-lines marker path{fill:var(--line-strong)}.flow-lines text{fill:var(--muted);font-size:12px;font-weight:800;letter-spacing:.05em}.ready-path .edge-pass{stroke:var(--green);stroke-width:3}.ready-path .edge-fail{stroke:var(--line)}.blocked-path .edge-pass{stroke:var(--line)}.blocked-path .edge-fail{stroke:var(--red);stroke-width:3}
.flow-node{position:absolute;display:flex;flex-direction:column;align-items:flex-start;justify-content:center;gap:5px;padding:16px;text-align:left;border:1px solid var(--line-strong);border-radius:15px;background:var(--panel);box-shadow:0 14px 28px #0000002b;cursor:pointer;transition:transform .18s ease,border-color .18s ease,box-shadow .18s ease}.flow-node:hover,.flow-node.selected{transform:translateY(-3px);border-color:var(--blue);box-shadow:0 18px 34px #00000040,0 0 0 4px var(--blue-soft)}.flow-node strong{font-size:1rem}.flow-node>span:last-child{color:var(--muted);font-size:.78rem}.flow-node.artifact{border-color:var(--green)}.flow-node.gate{border-color:var(--amber);background:var(--amber-soft);clip-path:polygon(50% 0,100% 50%,50% 100%,0 50%);align-items:center;text-align:center;padding:30px}.flow-node.success{border-color:var(--green);background:var(--green-soft)}.flow-node.blocked{border-color:var(--red);background:var(--red-soft)}.node-trigger{left:40px;top:190px;width:170px;height:110px}.node-areas{left:280px;top:65px;width:250px;height:340px}.node-impact{left:580px;top:160px;width:220px;height:160px}.node-gate{left:858px;top:175px;width:144px;height:144px}.node-ready{left:1042px;top:190px;width:138px;height:110px}.node-repair{left:1042px;top:380px;width:138px;height:92px}
.impact-spine{justify-content:flex-start}.impact-spine ol{display:grid;width:100%;gap:7px;margin:9px 0 0;padding:0;list-style:none}.impact-spine li{display:grid;grid-template-columns:27px minmax(0,1fr) auto;gap:7px;align-items:center;padding:7px 0;border-top:1px solid var(--line)}.impact-spine li span{color:var(--blue);font-size:.7rem;font-weight:900}.impact-spine li b{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:.75rem}.impact-spine li em{color:var(--muted);font-size:.62rem;font-style:normal;text-transform:uppercase}
.inspector{display:grid;grid-template-columns:minmax(170px,.7fr) minmax(260px,1.3fr);gap:16px 30px;margin-top:16px;padding:20px 22px;border:1px solid var(--line);border-radius:15px;background:var(--panel)}.inspector h3{margin:4px 0 0;font-size:1.25rem}.inspector>p{margin:21px 0 0;color:var(--muted)}.inspector dl{grid-column:1/-1;display:grid;grid-template-columns:110px 1fr;gap:8px 18px;margin:0;padding-top:14px;border-top:1px solid var(--line)}.inspector dt{color:var(--muted);font-weight:750}.inspector dd{margin:0}
.path-layout{display:grid;grid-template-columns:250px minmax(0,1fr);gap:24px}.path-rail{display:grid;align-content:start;gap:7px;padding:0}.path-step{display:grid;grid-template-columns:38px 1fr;gap:1px 10px;padding:11px;border:1px solid transparent;border-radius:11px;background:transparent;text-align:left;cursor:pointer}.path-step span{grid-row:1/3;display:grid;place-items:center;width:34px;height:34px;border:1px solid var(--line);border-radius:50%;color:var(--muted);font-size:.78rem}.path-step small{color:var(--muted)}.path-step.active{background:var(--panel);border-color:var(--line)}.path-step.active span{background:var(--blue);border-color:var(--blue);color:#07101e}.path-stage{min-width:0}.path-panel{min-height:420px;padding:34px;border:1px solid var(--line);border-radius:18px;background:var(--panel);box-shadow:var(--shadow)}.path-panel h2{margin:6px 0 8px;font-size:clamp(1.6rem,3vw,2.5rem)}.path-lead{max-width:720px;color:var(--muted)}.mini-flow{display:grid;grid-template-columns:1fr auto 1.3fr auto 1fr;align-items:stretch;gap:10px;margin:42px 0}.mini-flow>div{display:flex;flex-direction:column;gap:8px;justify-content:center;min-height:130px;padding:18px;border:1px solid var(--line);border-radius:13px;background:var(--bg-soft)}.mini-flow>div.current{border-color:var(--blue);background:var(--blue-soft)}.mini-flow>span{align-self:center;color:var(--blue);font-size:1.5rem}.mini-flow small{color:var(--muted);text-transform:uppercase;letter-spacing:.1em}.mini-flow em{color:var(--muted);font-style:normal}.path-actions{display:flex;justify-content:space-between;gap:10px}.path-actions button{padding:10px 14px;border:1px solid var(--line);border-radius:9px;background:var(--bg-soft);cursor:pointer}.path-actions button.primary{background:var(--blue);border-color:var(--blue);color:#07101e;font-weight:800}
.section-heading{max-width:760px;margin:8px 0 25px}.section-heading p{color:var(--muted)}.area-row{display:grid;grid-template-columns:52px minmax(240px,1.2fr) minmax(220px,.8fr);gap:16px 22px;padding:22px 0;border-top:1px solid var(--line)}.impact-rank{display:grid;place-items:center;width:44px;height:44px;border:1px solid var(--line-strong);border-radius:13px;color:var(--blue);font-weight:900}.area-copy{display:flex;gap:14px}.area-copy h3{margin:0}.area-copy p{margin:4px 0;color:var(--muted)}.area-copy small{display:block;margin-top:5px;color:var(--muted)}.area-title{display:flex;align-items:center;gap:9px;flex-wrap:wrap}.impact-level{padding:3px 7px;border-radius:999px;background:var(--blue-soft);color:var(--blue);font-size:.68rem;letter-spacing:.06em;text-transform:uppercase}.impact-level.critical,.impact-level.high{background:var(--red-soft);color:var(--red)}.impact-level.medium{background:var(--amber-soft);color:var(--amber)}.impact-level.low{background:var(--green-soft);color:var(--green)}.area-mark{width:9px;height:42px;border-radius:6px;background:var(--blue)}.area-mark.verification{background:var(--green)}.area-mark.docs{background:var(--amber)}.area-mark.delivery{background:#b998ff}.area-mark.other,.area-mark.supporting{background:var(--muted)}.area-measure{display:grid;grid-template-columns:1fr auto;gap:4px 12px;align-content:center}.area-measure>span{grid-column:2;color:var(--muted);font-size:.82rem}.area-bar{height:7px;border-radius:9px;background:var(--line);overflow:hidden}.area-bar i{display:block;height:100%;border-radius:inherit;background:var(--blue)}.area-row details{grid-column:2/-1}.area-row summary{color:var(--blue);cursor:pointer}.file-list{display:grid;gap:6px;padding:12px 0 0;list-style:none}.file-list li{display:flex;justify-content:space-between;gap:20px;padding:8px 10px;border-radius:8px;background:var(--panel)}.file-list span{color:var(--muted);white-space:nowrap}
.evidence-layout{display:grid;grid-template-columns:minmax(0,1fr) 340px;gap:30px}.evidence-claim{display:grid;grid-template-columns:46px 1fr;gap:14px;padding:20px 0;border-top:1px solid var(--line)}.claim-number{display:grid;place-items:center;width:38px;height:38px;border-radius:10px;background:var(--blue-soft);color:var(--blue);font-weight:850}.evidence-claim h3{margin:0}.evidence-claim p{color:var(--muted)}.anchor-row{display:flex;flex-wrap:wrap;align-items:center;gap:8px}.anchor-row a{padding:5px 7px;border:1px solid var(--line);border-radius:7px;text-decoration:none}.proof-state{display:inline-flex;padding:3px 7px;border-radius:999px;background:var(--bg-soft);color:var(--muted);font-size:.72rem;text-transform:uppercase}.proof-state.complete,.proof-state.success,.proof-state.neutral,.proof-state.skipped{background:var(--green-soft);color:var(--green)}.proof-state.partial,.proof-state.pending,.proof-state.in_progress{background:var(--amber-soft);color:var(--amber)}.proof-state.failed,.proof-state.failure,.proof-state.stale,.proof-state.forbidden{background:var(--red-soft);color:var(--red)}.evidence-side{display:grid;align-content:start;gap:12px}.evidence-side section{padding:17px;border:1px solid var(--line);border-radius:13px;background:var(--panel)}.evidence-side h3{margin:0 0 12px}.status-list,.blocker-list{display:grid;gap:8px;margin:0;padding:0;list-style:none}.status-list li{display:flex;justify-content:space-between;gap:12px}.blocker-list{padding-left:18px;list-style:disc;color:var(--muted)}.empty-state{padding:24px;border:1px dashed var(--line-strong);border-radius:14px;color:var(--muted)}
.pr-footer{display:flex;justify-content:space-between;gap:20px;margin-top:28px;padding-top:18px;border-top:1px solid var(--line);color:var(--muted);font-size:.82rem}.pr-footer span{overflow-wrap:anywhere}
@media(max-width:1050px){.app-shell{grid-template-columns:1fr}.inbox{display:flex;gap:7px;overflow-x:auto;border-right:0;border-bottom:1px solid var(--line)}.inbox-label{display:none}.inbox-card{min-width:230px}.pr-head{align-items:start;flex-direction:column}.path-layout,.evidence-layout{grid-template-columns:1fr}.path-rail{display:flex;overflow-x:auto}.path-step{min-width:190px}.evidence-side{grid-template-columns:repeat(auto-fit,minmax(220px,1fr))}}
@media(max-width:700px){.app-bar{padding:11px 14px}.identity small{display:none}.pr-section{padding:24px 14px 40px}.view-tabs{top:58px;overflow-x:auto}.view-tabs button{white-space:nowrap}.inspector{grid-template-columns:1fr}.inspector>p{margin:0}.mini-flow{grid-template-columns:1fr}.mini-flow>span{transform:rotate(90deg)}.area-row{grid-template-columns:44px 1fr}.area-measure{grid-column:2}.area-row details{grid-column:1/-1}.pr-footer{flex-direction:column}.file-list li{align-items:start;flex-direction:column;gap:3px}}
@media(prefers-reduced-motion:reduce){*{scroll-behavior:auto!important;transition:none!important}}
</style>'''
    body_start = f'''</head><body data-theme="dark"><header class="app-bar"><div class="identity"><span class="identity-mark">PR</span><div><b>{_e(packet['repository']['nameWithOwner'])}</b><small>Evidence workbench · snapshot {_e(packet['generated_at'])}</small></div></div><button type="button" class="theme-toggle" data-theme-toggle>Light mode</button></header>
<main class="app-shell"><nav class="inbox" aria-label="Pull request inbox"><span class="inbox-label">Review inbox</span>{inbox}</nav><div>{sections}</div></main>'''
    script = '''
<script>
function selectPr(number){
  document.querySelectorAll(".pr-section").forEach(function(section){section.hidden=section.dataset.pr!==number});
  document.querySelectorAll("[data-pr-card]").forEach(function(card){card.classList.toggle("active",card.dataset.prCard===number)});
}
function activateView(section,name){
  section.querySelectorAll("[data-view-tab]").forEach(function(button){button.setAttribute("aria-selected",String(button.dataset.viewTab===name))});
  section.querySelectorAll("[data-view-panel]").forEach(function(panel){panel.hidden=panel.dataset.viewPanel!==name});
}
function activateStep(section,number){
  var value=String(Math.max(1,Math.min(5,Number(number)||1)));
  section.querySelectorAll("[data-step]").forEach(function(button){button.classList.toggle("active",button.dataset.step===value)});
  section.querySelectorAll("[data-step-panel]").forEach(function(panel){panel.hidden=panel.dataset.stepPanel!==value});
}
function restoreHash(){
  var match=location.hash.match(/^#pr-(\\d+)-view-(overview|workflow|changes|evidence)(?:-step-(\\d+))?$/);
  if(!match){return}
  selectPr(match[1]);
  var section=document.querySelector('.pr-section[data-pr="'+match[1]+'"]');
  if(!section){return}
  activateView(section,match[2]);
  if(match[3]){activateStep(section,match[3])}
}
document.querySelectorAll("[data-pr-card]").forEach(function(card){
  card.addEventListener("click",function(event){event.preventDefault();location.hash="pr-"+card.dataset.prCard+"-view-overview"});
});
document.querySelectorAll("[data-view-tab]").forEach(function(button){
  button.addEventListener("click",function(){
    var section=button.closest(".pr-section");
    location.hash="pr-"+section.dataset.pr+"-view-"+button.dataset.viewTab+(button.dataset.viewTab==="workflow"?"-step-1":"");
  });
});
document.querySelectorAll(".flow-node").forEach(function(node){
  node.addEventListener("click",function(){
    var section=node.closest(".pr-section");
    section.querySelectorAll(".flow-node").forEach(function(other){other.classList.toggle("selected",other===node)});
    section.querySelector("[data-inspector-heading]").textContent=node.dataset.inspectorTitle;
    section.querySelector("[data-inspector-body]").textContent=node.dataset.inspectorBody;
    section.querySelector("[data-inspector-evidence]").textContent=node.dataset.inspectorEvidence;
    section.querySelector("[data-inspector-caveat]").textContent=node.dataset.inspectorCaveat;
  });
});
document.querySelectorAll("[data-step]").forEach(function(button){
  button.addEventListener("click",function(){var section=button.closest(".pr-section");location.hash="pr-"+section.dataset.pr+"-view-workflow-step-"+button.dataset.step});
});
document.querySelectorAll("[data-step-prev],[data-step-next]").forEach(function(button){
  button.addEventListener("click",function(){
    var section=button.closest(".pr-section");
    var active=Number(section.querySelector("[data-step].active").dataset.step);
    var next=button.hasAttribute("data-step-prev")?active-1:active+1;
    location.hash="pr-"+section.dataset.pr+"-view-workflow-step-"+Math.max(1,Math.min(5,next));
  });
});
document.querySelector("[data-theme-toggle]").addEventListener("click",function(){
  var light=document.body.dataset.theme==="light";document.body.dataset.theme=light?"dark":"light";this.textContent=light?"Light mode":"Dark mode";
});
document.addEventListener("keydown",function(event){
  if(event.key!=="ArrowLeft"&&event.key!=="ArrowRight"){return}
  var section=document.querySelector(".pr-section:not([hidden])");
  if(!section||section.querySelector('[data-view-tab="workflow"]').getAttribute("aria-selected")!=="true"){return}
  var active=Number(section.querySelector("[data-step].active").dataset.step);
  var next=active+(event.key==="ArrowRight"?1:-1);
  location.hash="pr-"+section.dataset.pr+"-view-workflow-step-"+Math.max(1,Math.min(5,next));
});
window.addEventListener("hashchange",restoreHash);
if(!location.hash){location.hash="pr-"+document.querySelector(".pr-section:not([hidden])").dataset.pr+"-view-overview"}else{restoreHash()}
</script></body></html>'''
    return head + styles + body_start + script


def prepare_review(packet: dict[str, Any], number: int, event: str, body: str) -> dict[str, Any]:
    event = event.upper()
    if event not in REVIEW_EVENTS:
        raise ValueError(f"event must be one of {sorted(REVIEW_EVENTS)}")
    pr = next((item for item in packet.get("pull_requests", []) if item["number"] == number), None)
    if not pr:
        raise ValueError(f"PR #{number} not found")
    if pr["packet_state"] != "complete" or pr["decision_state"] != "ready for human decision" or pr["blockers"] or not pr["head_sha"]:
        raise RuntimeError("cannot prepare a review until evidence is ready for human decision")
    body_digest = hashlib.sha256(body.encode("utf-8")).hexdigest()[:12]
    target = f"{pr['repository']}#{number}@{pr['head_sha']}:{event}:body-{body_digest}"
    return {"schema_version": 1, "repository": pr["repository"], "number": number, "head_sha": pr["head_sha"], "event": event, "body": body, "confirmation": target, "submitted": False}


def submit_review(draft: dict[str, Any], *, dry_run: bool, confirm: str | None, runner: Callable[..., Any] = _run) -> dict[str, Any]:
    required = {"repository", "number", "head_sha", "event", "body", "confirmation"}
    if required - set(draft) or draft.get("event") not in REVIEW_EVENTS:
        raise ValueError("review draft is missing required fields or has an invalid event")
    expected = f"{draft['repository']}#{draft['number']}@{draft['head_sha']}:{draft['event']}:body-{hashlib.sha256(str(draft['body']).encode('utf-8')).hexdigest()[:12]}"
    if draft["confirmation"] != expected:
        raise RuntimeError("review draft content changed after confirmation was prepared")
    if dry_run:
        return {"status": "dry-run", "target": draft["confirmation"], "event": draft["event"], "body": draft["body"], "head_sha": draft["head_sha"]}
    if confirm != draft["confirmation"]:
        return {"status": "cancelled", "target": draft["confirmation"]}
    current = runner(["gh", "pr", "view", str(draft["number"]), "--repo", draft["repository"], "--json", "headRefOid"], timeout=30)
    if current.returncode != 0:
        raise RuntimeError(f"SHA revalidation failed: {current.stderr.strip()}")
    if json.loads(current.stdout).get("headRefOid") != draft["head_sha"]:
        raise RuntimeError("head SHA changed; regenerate the evidence packet")
    endpoint = f"repos/{draft['repository']}/pulls/{draft['number']}/reviews"
    command = ["gh", "api", "--method", "POST", endpoint, "-f", f"commit_id={draft['head_sha']}", "-f", f"event={draft['event']}", "-f", f"body={draft['body']}"]
    try:
        result = runner(command, timeout=30)
    except subprocess.TimeoutExpired as exc:
        raise RuntimeError("unknown submission state after timeout; inspect GitHub manually and do not retry automatically") from exc
    if result.returncode != 0:
        raise RuntimeError(f"review submission failed without retry: {result.stderr.strip()}")
    try:
        response = json.loads(result.stdout)
    except json.JSONDecodeError as exc:
        raise RuntimeError("review was submitted but its response was invalid; inspect GitHub manually") from exc
    return {"status": "submitted", "url": response.get("html_url") or f"https://github.com/{draft['repository']}/pull/{draft['number']}"}


def _write_json(path: str, value: dict[str, Any]) -> None:
    Path(path).parent.mkdir(parents=True, exist_ok=True)
    Path(path).write_text(json.dumps(value, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    inbox = sub.add_parser("inbox", help="Collect a normalized repository PR inbox")
    source = inbox.add_mutually_exclusive_group(required=True)
    source.add_argument("--fixture")
    source.add_argument("--repo")
    inbox.add_argument("--limit", type=int, default=20)
    inbox.add_argument("--output", required=True)
    render = sub.add_parser("render", help="Render a self-contained HTML workbench")
    render.add_argument("--packet", required=True)
    render.add_argument("--pr", type=int)
    render.add_argument("--output", required=True)
    prepare = sub.add_parser("prepare-review", help="Create an inert, SHA-pinned review draft")
    prepare.add_argument("--packet", required=True)
    prepare.add_argument("--pr", type=int, required=True)
    prepare.add_argument("--event", choices=sorted(REVIEW_EVENTS), required=True)
    prepare.add_argument("--body", required=True)
    prepare.add_argument("--output", required=True)
    submit = sub.add_parser("submit-review", help="Preview or explicitly submit a review draft")
    submit.add_argument("--draft", required=True)
    submit.add_argument("--confirm")
    submit.add_argument("--dry-run", action="store_true")
    return parser


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    if args.command == "inbox":
        packet = collect_fixture(args.fixture) if args.fixture else collect_github(args.repo, limit=args.limit)
        _write_json(args.output, packet)
        print(f"Wrote {len(packet['pull_requests'])} PRs to {args.output}")
    elif args.command == "render":
        packet = json.loads(Path(args.packet).read_text(encoding="utf-8"))
        Path(args.output).parent.mkdir(parents=True, exist_ok=True)
        Path(args.output).write_text(render_html(packet, selected_pr=args.pr), encoding="utf-8")
        print(f"Wrote offline workbench to {args.output}")
    elif args.command == "prepare-review":
        packet = json.loads(Path(args.packet).read_text(encoding="utf-8"))
        draft = prepare_review(packet, args.pr, args.event, args.body)
        _write_json(args.output, draft)
        print(f"Review draft created. Confirmation target: {draft['confirmation']}")
    elif args.command == "submit-review":
        draft = json.loads(Path(args.draft).read_text(encoding="utf-8"))
        result = submit_review(draft, dry_run=args.dry_run, confirm=args.confirm)
        print(json.dumps(result, sort_keys=True))
    return 0


if __name__ == "__main__":
    sys.exit(main())
