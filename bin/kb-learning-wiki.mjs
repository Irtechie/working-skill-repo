#!/usr/bin/env node

import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRoot = path.resolve(scriptDir, "..");

function stripQuotes(value) {
  const trimmed = value.trim();
  if (
    trimmed.length >= 2 &&
    ((trimmed.startsWith('"') && trimmed.endsWith('"')) ||
      (trimmed.startsWith("'") && trimmed.endsWith("'")))
  ) {
    return trimmed.slice(1, -1);
  }
  return trimmed;
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function posixPath(root, file) {
  return path.relative(root, file).split(path.sep).join("/");
}

async function filesUnder(root, relativeDir, extension) {
  const base = path.join(root, relativeDir);
  const found = [];

  async function walk(directory) {
    let entries;
    try {
      entries = await fs.readdir(directory, { withFileTypes: true });
    } catch (error) {
      if (error.code === "ENOENT") {
        return;
      }
      throw error;
    }
    for (const entry of entries) {
      const fullPath = path.join(directory, entry.name);
      if (entry.isDirectory()) {
        await walk(fullPath);
      } else if (!extension || entry.name.endsWith(extension)) {
        found.push(fullPath);
      }
    }
  }

  await walk(base);
  return found.sort();
}

export function parseInstincts(content, source) {
  const instincts = [];
  let current = null;
  let inEvidence = false;

  for (const line of content.split(/\r?\n/)) {
    const idMatch = line.match(/^\s{2}- id:\s*(.+)$/);
    if (idMatch) {
      if (current) {
        instincts.push(current);
      }
      current = { id: stripQuotes(idMatch[1]), evidence: [], source };
      inEvidence = false;
      continue;
    }
    if (!current) {
      continue;
    }
    const fieldMatch = line.match(/^\s{4}([a-z_]+):\s*(.*)$/);
    if (fieldMatch) {
      const [, key, rawValue] = fieldMatch;
      if (key === "evidence") {
        inEvidence = true;
        continue;
      }
      inEvidence = false;
      const value = stripQuotes(rawValue);
      current[key] =
        key === "confidence" ? Number(value) :
        key === "observations" ? Number.parseInt(value, 10) :
        value;
      continue;
    }
    const evidenceMatch = line.match(/^\s{6}-\s*(.+)$/);
    if (inEvidence && evidenceMatch) {
      current.evidence.push(stripQuotes(evidenceMatch[1]));
    }
  }
  if (current) {
    instincts.push(current);
  }
  return instincts;
}

function parseFrontmatter(content) {
  if (!content.startsWith("---")) {
    return {};
  }
  const end = content.indexOf("\n---", 3);
  if (end === -1) {
    return {};
  }
  const fields = {};
  for (const line of content.slice(3, end).split(/\r?\n/)) {
    const match = line.match(/^([a-z_]+):\s*(.*)$/i);
    if (match) {
      fields[match[1]] = stripQuotes(match[2]);
    }
  }
  return fields;
}

function markdownTitle(content, fallback) {
  return content.match(/^#\s+(.+)$/m)?.[1]?.trim() || fallback;
}

function markdownExcerpt(content) {
  const body = content.replace(/^---[\s\S]*?\n---\s*/, "");
  const lines = body.split(/\r?\n/);
  const paragraph = [];
  let inCode = false;
  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (line.startsWith("```")) {
      inCode = !inCode;
      continue;
    }
    if (inCode || !line || line.startsWith("#") || line.startsWith("|")) {
      if (paragraph.length > 0) {
        break;
      }
      continue;
    }
    paragraph.push(line.replace(/^[-*]\s+/, ""));
    if (paragraph.join(" ").length >= 220) {
      break;
    }
  }
  const excerpt = paragraph.join(" ");
  return excerpt.length > 260 ? `${excerpt.slice(0, 257)}...` : excerpt;
}

function statusFromMarkdown(content) {
  return content.match(/^Status:\s*(.+)$/mi)?.[1]?.trim() || "reference";
}

function dateFromFilename(source) {
  return source.match(/(\d{4}-\d{2}-\d{2})/)?.[1] || "";
}

async function collectMarkdown(root, directory, type) {
  const files = await filesUnder(root, directory, ".md");
  return Promise.all(files.map(async (file) => {
    const content = await fs.readFile(file, "utf8");
    const source = posixPath(root, file);
    const frontmatter = parseFrontmatter(content);
    return {
      type,
      title: frontmatter.title || markdownTitle(content, path.basename(file, ".md")),
      date: frontmatter.date || dateFromFilename(source),
      category: frontmatter.category || path.basename(path.dirname(file)),
      status: statusFromMarkdown(content),
      excerpt: markdownExcerpt(content),
      source,
    };
  }));
}

function reviewState(instinct, now) {
  const lastSeen = Date.parse(`${instinct.last_seen}T00:00:00Z`);
  if (!Number.isFinite(lastSeen)) {
    return "unknown";
  }
  const ageDays = Math.floor((now.getTime() - lastSeen) / 86_400_000);
  if (ageDays > 180) {
    return "archive";
  }
  if (ageDays > 90) {
    return "review";
  }
  return "current";
}

export async function collectLearning(root = defaultRoot, now = new Date()) {
  const instinctFiles = await filesUnder(root, "docs/context/kb/instincts", ".yaml");
  const instincts = (
    await Promise.all(instinctFiles.map(async (file) => {
      const source = posixPath(root, file);
      return parseInstincts(await fs.readFile(file, "utf8"), source);
    }))
  ).flat().map((instinct) => ({
    ...instinct,
    reviewState: reviewState(instinct, now),
  }));

  const [solutions, goals, research] = await Promise.all([
    collectMarkdown(root, "docs/solutions", "solution"),
    collectMarkdown(root, "docs/context/goals", "goal"),
    collectMarkdown(root, "docs/context/research", "research"),
  ]);

  let landmines = { active: 0, resolved: 0, reviewed: "" };
  try {
    const content = await fs.readFile(path.join(root, "docs/context/landmines.md"), "utf8");
    landmines = {
      active: /## Active Landmines\s+None\./s.test(content) ? 0 : (content.match(/status:\s*active/g) || []).length,
      resolved: /## Resolved Landmines\s+None\./s.test(content) ? 0 : (content.match(/status:\s*resolved/g) || []).length,
      reviewed: content.match(/Last reviewed:\s*(\d{4}-\d{2}-\d{2})/)?.[1] || "",
    };
  } catch (error) {
    if (error.code !== "ENOENT") {
      throw error;
    }
  }

  return {
    generatedAt: now.toISOString(),
    instincts: instincts.sort((a, b) =>
      String(b.last_seen).localeCompare(String(a.last_seen)) || a.id.localeCompare(b.id)
    ),
    solutions: solutions.sort((a, b) => b.date.localeCompare(a.date)),
    goals: goals.sort((a, b) => a.title.localeCompare(b.title)),
    research: research.sort((a, b) => b.date.localeCompare(a.date)),
    landmines,
  };
}

function sourceHref(source) {
  return `../${source.replace(/^docs\//, "").split("/").map(encodeURIComponent).join("/")}`;
}

function confidenceLabel(value) {
  if (value >= 0.8) return "promotion candidate";
  if (value >= 0.6) return "supported";
  return "emerging";
}

function instinctCard(instinct) {
  const evidence = instinct.evidence
    .map((item) => `<li>${escapeHtml(item)}</li>`)
    .join("");
  return `
    <article class="knowledge-card" data-kind="instinct" data-scope="${escapeHtml(instinct.scope)}" data-state="${escapeHtml(instinct.reviewState)}">
      <div class="card-topline">
        <span class="eyebrow">${escapeHtml(instinct.scope)}</span>
        <span class="status status-${escapeHtml(instinct.reviewState)}">${escapeHtml(instinct.reviewState)}</span>
      </div>
      <h3>${escapeHtml(instinct.id)}</h3>
      <p class="trigger">${escapeHtml(instinct.trigger)}</p>
      <p>${escapeHtml(instinct.behavior)}</p>
      <div class="meter-row">
        <span>${Math.round(instinct.confidence * 100)}% confidence</span>
        <span>${escapeHtml(confidenceLabel(instinct.confidence))}</span>
        <span>${escapeHtml(instinct.observations)} observation${instinct.observations === 1 ? "" : "s"}</span>
      </div>
      <div class="meter" aria-label="${Math.round(instinct.confidence * 100)} percent confidence">
        <span style="width: ${Math.round(instinct.confidence * 100)}%"></span>
      </div>
      <details>
        <summary>Evidence (${instinct.evidence.length})</summary>
        <ul>${evidence}</ul>
      </details>
      <footer>
        <span>Seen ${escapeHtml(instinct.last_seen)}</span>
        <a href="${sourceHref(instinct.source)}">Source</a>
      </footer>
    </article>`;
}

function documentCard(document) {
  return `
    <article class="knowledge-card" data-kind="${escapeHtml(document.type)}" data-scope="${escapeHtml(document.category)}" data-state="${escapeHtml(document.status)}">
      <div class="card-topline">
        <span class="eyebrow">${escapeHtml(document.type)}</span>
        <span class="status">${escapeHtml(document.status)}</span>
      </div>
      <h3>${escapeHtml(document.title)}</h3>
      <p>${escapeHtml(document.excerpt || "Open the source document for details.")}</p>
      <footer>
        <span>${escapeHtml(document.date || document.category)}</span>
        <a href="${sourceHref(document.source)}">Source</a>
      </footer>
    </article>`;
}

function section(id, title, description, cards) {
  return `
    <section id="${id}" class="wiki-section">
      <div class="section-heading">
        <div>
          <p class="eyebrow">${escapeHtml(id)}</p>
          <h2>${escapeHtml(title)}</h2>
          <p>${escapeHtml(description)}</p>
        </div>
        <span class="count">${cards.length}</span>
      </div>
      <div class="card-grid">${cards.join("") || '<p class="empty">No entries found.</p>'}</div>
    </section>`;
}

export function renderWiki(data) {
  const scopes = [...new Set(data.instincts.map((item) => item.scope))].sort();
  const staleCount = data.instincts.filter((item) => item.reviewState !== "current").length;
  const latest = [...data.instincts]
    .sort((a, b) => String(b.last_seen).localeCompare(String(a.last_seen)))
    .slice(0, 4);
  const scopeOptions = scopes
    .map((scope) => `<option value="${escapeHtml(scope)}">${escapeHtml(scope)}</option>`)
    .join("");

  return `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>KB Learning Wiki</title>
  <script>
    (() => {
      const param = new URLSearchParams(window.location.search).get("scoutTheme");
      const theme =
        param || (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
      document.documentElement.setAttribute("data-theme", theme);
    })();
  </script>
  <style>
    :root {
      color-scheme: light;
      --cp-bg: #f7f4ef;
      --cp-bg-elevated: #fcfbf8;
      --cp-surface: #ffffff;
      --cp-surface-soft: #f5f5f5;
      --cp-border: #dedede;
      --cp-border-strong: #919191;
      --cp-text: #242424;
      --cp-text-muted: #5c5c5c;
      --cp-text-soft: #6f6f6f;
      --cp-accent: #b11f4b;
      --cp-accent-hover: #9a1a41;
      --cp-accent-soft: rgba(177, 31, 75, 0.08);
      --cp-accent-fg: #ffffff;
      --cp-success: #16a34a;
      --cp-danger: #dc2626;
      --cp-warning: #f59e0b;
      --cp-link: #0078d4;
      --cp-shadow: 0 18px 48px rgba(0, 0, 0, 0.12);
      --cp-overlay: rgba(255, 255, 255, 0.8);
      --cp-panel: rgba(255, 255, 255, 0.86);
      --cp-panel-strong: rgba(255, 255, 255, 0.96);
      --cp-sheen: rgba(255, 255, 255, 0.55);
      --cp-highlight: rgba(177, 31, 75, 0.12);
    }
    html[data-theme="dark"] {
      color-scheme: dark;
      --cp-bg: #3d3b3a;
      --cp-bg-elevated: #343231;
      --cp-surface: #292929;
      --cp-surface-soft: #2e2e2e;
      --cp-border: #474747;
      --cp-border-strong: #5f5f5f;
      --cp-text: #dedede;
      --cp-text-muted: #919191;
      --cp-text-soft: #b0b0b0;
      --cp-accent: #fd8ea1;
      --cp-accent-hover: #fb7b91;
      --cp-accent-soft: rgba(253, 142, 161, 0.14);
      --cp-accent-fg: #1a1a1a;
      --cp-success: #4ade80;
      --cp-danger: #f87171;
      --cp-warning: #fbbf24;
      --cp-link: #4da6ff;
      --cp-shadow: 0 18px 48px rgba(0, 0, 0, 0.32);
      --cp-overlay: rgba(41, 41, 41, 0.88);
      --cp-panel: rgba(41, 41, 41, 0.72);
      --cp-panel-strong: rgba(41, 41, 41, 0.96);
      --cp-sheen: rgba(255, 255, 255, 0.04);
      --cp-highlight: rgba(253, 142, 161, 0.12);
    }
    * { box-sizing: border-box; }
    html { scroll-behavior: smooth; }
    body {
      margin: 0;
      background: var(--cp-bg);
      color: var(--cp-text);
      font-family: "Segoe UI", Aptos, Calibri, -apple-system, BlinkMacSystemFont, sans-serif;
    }
    a { color: var(--cp-link); }
    button, input, select { font: inherit; }
    .shell { display: grid; grid-template-columns: 248px minmax(0, 1fr); min-height: 100vh; }
    .sidebar {
      position: sticky; top: 0; height: 100vh; padding: 24px 20px;
      background: var(--cp-bg-elevated); border-right: 1px solid var(--cp-border);
    }
    .brand { display: flex; gap: 12px; align-items: center; margin-bottom: 28px; }
    .brand-mark {
      display: grid; place-items: center; width: 36px; height: 36px;
      border-radius: 10px; background: var(--cp-accent); color: var(--cp-accent-fg); font-weight: 700;
    }
    .brand strong, .brand span { display: block; }
    .brand span { color: var(--cp-text-muted); font-size: 12px; margin-top: 2px; }
    nav { display: grid; gap: 4px; }
    nav a {
      padding: 10px 12px; border-radius: 10px; color: var(--cp-text-muted);
      text-decoration: none; font-weight: 600;
    }
    nav a:hover { background: var(--cp-accent-soft); color: var(--cp-accent); }
    .sidebar-note {
      position: absolute; left: 20px; right: 20px; bottom: 20px;
      color: var(--cp-text-muted); font-size: 12px; line-height: 1.5;
    }
    main { min-width: 0; }
    .hero {
      padding: 48px clamp(24px, 5vw, 72px) 36px;
      background: var(--cp-bg-elevated); border-bottom: 1px solid var(--cp-border);
    }
    .hero-row { display: flex; justify-content: space-between; gap: 24px; align-items: end; }
    h1 { margin: 6px 0 12px; font-size: clamp(36px, 5vw, 64px); letter-spacing: -0.04em; }
    h2 { margin: 4px 0 8px; font-size: 28px; }
    h3 { margin: 10px 0 8px; font-size: 18px; }
    p { line-height: 1.55; }
    .hero p { max-width: 760px; color: var(--cp-text-muted); font-size: 18px; }
    .eyebrow {
      margin: 0; color: var(--cp-accent); font-size: 12px;
      font-weight: 700; letter-spacing: 0.1em; text-transform: uppercase;
    }
    .controls {
      display: grid; grid-template-columns: minmax(220px, 1fr) 220px; gap: 12px;
      margin-top: 28px; max-width: 760px;
    }
    input, select {
      width: 100%; padding: 11px 12px; color: var(--cp-text);
      background: var(--cp-surface); border: 1px solid var(--cp-border); border-radius: 10px;
    }
    .content { padding: 32px clamp(24px, 5vw, 72px) 72px; }
    .stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 28px; }
    .stat, .pipeline, .knowledge-card {
      background: var(--cp-surface); border: 1px solid var(--cp-border);
      border-radius: 16px; box-shadow: var(--cp-shadow);
    }
    .stat { padding: 20px; }
    .stat strong { display: block; font-size: 32px; }
    .stat span { color: var(--cp-text-muted); }
    .pipeline { padding: 24px; margin-bottom: 36px; overflow-x: auto; }
    .pipeline-track { display: flex; min-width: 780px; align-items: center; gap: 10px; margin-top: 20px; }
    .pipeline-step {
      flex: 1; min-height: 92px; padding: 14px; border: 1px solid var(--cp-border);
      border-radius: 10px; background: var(--cp-surface-soft);
    }
    .pipeline-step strong, .pipeline-step span { display: block; }
    .pipeline-step span { color: var(--cp-text-muted); font-size: 12px; margin-top: 6px; }
    .arrow { color: var(--cp-accent); font-weight: 700; }
    .wiki-section { scroll-margin-top: 20px; margin-top: 52px; }
    .section-heading { display: flex; justify-content: space-between; gap: 24px; align-items: end; margin-bottom: 18px; }
    .section-heading p { margin: 0; color: var(--cp-text-muted); max-width: 760px; }
    .count {
      min-width: 42px; padding: 8px 10px; text-align: center;
      border-radius: 10px; background: var(--cp-accent-soft); color: var(--cp-accent); font-weight: 700;
    }
    .card-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 16px; }
    .knowledge-card { padding: 20px; min-width: 0; }
    .knowledge-card[hidden] { display: none; }
    .card-topline, .meter-row, .knowledge-card footer {
      display: flex; justify-content: space-between; gap: 12px; align-items: center;
    }
    .status {
      padding: 4px 8px; border-radius: 10px; background: var(--cp-surface-soft);
      color: var(--cp-text-muted); font-size: 12px;
    }
    .status-current { color: var(--cp-success); }
    .status-review { color: var(--cp-warning); }
    .status-archive { color: var(--cp-danger); }
    .trigger { color: var(--cp-text-muted); font-style: italic; }
    .meter-row { flex-wrap: wrap; color: var(--cp-text-muted); font-size: 12px; margin-top: 18px; }
    .meter { height: 5px; margin: 8px 0 18px; overflow: hidden; border-radius: 10px; background: var(--cp-surface-soft); }
    .meter span { display: block; height: 100%; background: var(--cp-accent); }
    details { border-top: 1px solid var(--cp-border); padding-top: 12px; }
    summary { cursor: pointer; color: var(--cp-link); }
    li { margin: 8px 0; color: var(--cp-text-muted); line-height: 1.45; }
    .knowledge-card footer {
      margin-top: 18px; padding-top: 14px; border-top: 1px solid var(--cp-border);
      color: var(--cp-text-muted); font-size: 12px;
    }
    .empty { color: var(--cp-text-muted); }
    .no-results {
      display: none; padding: 24px; margin-top: 20px; text-align: center;
      border: 1px dashed var(--cp-border-strong); border-radius: 16px; color: var(--cp-text-muted);
    }
    @media (max-width: 900px) {
      .shell { grid-template-columns: 1fr; }
      .sidebar { position: static; height: auto; border-right: 0; border-bottom: 1px solid var(--cp-border); }
      nav { grid-template-columns: repeat(3, 1fr); }
      .sidebar-note { display: none; }
      .stats { grid-template-columns: repeat(2, 1fr); }
    }
    @media (max-width: 600px) {
      nav { grid-template-columns: 1fr 1fr; }
      .hero-row { display: block; }
      .controls, .stats { grid-template-columns: 1fr; }
      .section-heading { align-items: start; }
    }
  </style>
</head>
<body>
  <div class="shell">
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-mark">KB</div>
        <div><strong>Learning Wiki</strong><span>Generated knowledge map</span></div>
      </div>
      <nav aria-label="Wiki sections">
        <a href="#overview">Overview</a>
        <a href="#instincts">Instincts</a>
        <a href="#solutions">Solutions</a>
        <a href="#goals">Goals</a>
        <a href="#research">Research</a>
        <a href="#governance">Governance</a>
      </nav>
      <p class="sidebar-note">Read-only projection. Edit the linked YAML or Markdown source, then run <code>npm run wiki:build</code>.</p>
    </aside>
    <main>
      <header class="hero">
        <div class="hero-row">
          <div>
            <p class="eyebrow">Repository intelligence</p>
            <h1>What this project has learned.</h1>
            <p>Browse evidence-backed instincts, durable solutions, active goals, and research without flattening the scope hierarchy that keeps agent memory safe.</p>
          </div>
        </div>
        <div class="controls">
          <input id="search" type="search" placeholder="Search behavior, evidence, title, or source..." aria-label="Search wiki">
          <select id="scope" aria-label="Filter by scope">
            <option value="">All scopes</option>
            ${scopeOptions}
          </select>
        </div>
      </header>
      <div class="content">
        <section id="overview" aria-labelledby="overview-title">
          <div class="stats">
            <div class="stat"><strong>${data.instincts.length}</strong><span>active instincts</span></div>
            <div class="stat"><strong>${data.solutions.length}</strong><span>compounded solutions</span></div>
            <div class="stat"><strong>${data.goals.length}</strong><span>goal ledgers</span></div>
            <div class="stat"><strong>${staleCount}</strong><span>need review</span></div>
          </div>
          <div class="pipeline" id="governance">
            <p class="eyebrow">Learning lifecycle</p>
            <h2 id="overview-title">Evidence becomes authority slowly.</h2>
            <p>Ordinary lessons stay local. They move upward only through recurrence and measured adoption.</p>
            <div class="pipeline-track">
              <div class="pipeline-step"><strong>Evidence</strong><span>Git, proof, research, curated feedback</span></div>
              <span class="arrow">-&gt;</span>
              <div class="pipeline-step"><strong>Steering</strong><span>Changes the next run without claiming permanence</span></div>
              <span class="arrow">-&gt;</span>
              <div class="pipeline-step"><strong>Scoped instinct</strong><span>Atomic trigger and behavior at the owning scope</span></div>
              <span class="arrow">-&gt;</span>
              <div class="pipeline-step"><strong>Promotion</strong><span>Sibling recurrence plus adoption proof</span></div>
              <span class="arrow">-&gt;</span>
              <div class="pipeline-step"><strong>Evolved skill</strong><span>Reusable behavior with explicit authority</span></div>
            </div>
          </div>
        </section>
        ${section("latest", "Latest learning", "The most recently observed instincts across all scopes.", latest.map(instinctCard))}
        ${section("instincts", "Instinct catalog", "Atomic learned behaviors. Scope controls who inherits them; confidence is evidence strength, not permission to promote.", data.instincts.map(instinctCard))}
        ${section("solutions", "Compounded solutions", "Human-readable explanations of solved problems and reusable guidance.", data.solutions.map(documentCard))}
        ${section("goals", "Goal ledgers", "Long-lived objectives, current state, blockers, and terminal proof.", data.goals.map(documentCard))}
        ${section("research", "Research library", "External evidence and comparisons that can inform work without automatically becoming authority.", data.research.map(documentCard))}
        <div id="no-results" class="no-results">No knowledge entries match the current filters.</div>
      </div>
    </main>
  </div>
  <script>
    const search = document.getElementById("search");
    const scope = document.getElementById("scope");
    const cards = [...document.querySelectorAll(".knowledge-card")];
    const noResults = document.getElementById("no-results");

    function applyFilters() {
      const query = search.value.trim().toLowerCase();
      const selectedScope = scope.value;
      let visible = 0;
      for (const card of cards) {
        const matchesQuery = !query || card.textContent.toLowerCase().includes(query);
        const matchesScope = !selectedScope || card.dataset.scope === selectedScope;
        card.hidden = !(matchesQuery && matchesScope);
        if (!card.hidden) visible += 1;
      }
      noResults.style.display = visible === 0 ? "block" : "none";
    }

    search.addEventListener("input", applyFilters);
    scope.addEventListener("change", applyFilters);
  </script>
</body>
</html>`;
}

export async function buildWiki(root = defaultRoot, outputPath) {
  const data = await collectLearning(root);
  const target = outputPath || path.join(root, "docs", "learning-wiki", "index.html");
  await fs.mkdir(path.dirname(target), { recursive: true });
  const rendered = renderWiki(data).replace(/[ \t]+\r?$/gm, "");
  await fs.writeFile(target, `${rendered}\n`, "utf8");
  return { target, data };
}

async function main() {
  const rootIndex = process.argv.indexOf("--root");
  const outputIndex = process.argv.indexOf("--output");
  const root = rootIndex >= 0 ? path.resolve(process.argv[rootIndex + 1]) : defaultRoot;
  const output = outputIndex >= 0 ? path.resolve(process.argv[outputIndex + 1]) : undefined;
  const result = await buildWiki(root, output);
  process.stdout.write(
    `learning wiki: ${result.data.instincts.length} instincts, ${result.data.solutions.length} solutions, ${result.data.goals.length} goals, ${result.data.research.length} research documents -> ${result.target}\n`,
  );
}

async function invokedAsMain() {
  if (!process.argv[1]) {
    return false;
  }
  const modulePath = await fs.realpath(fileURLToPath(import.meta.url));
  const invokedPath = await fs.realpath(path.resolve(process.argv[1]));
  return modulePath.toLowerCase() === invokedPath.toLowerCase();
}

if (await invokedAsMain()) {
  main().catch((error) => {
    process.stderr.write(`${error.stack || error}\n`);
    process.exitCode = 1;
  });
}
