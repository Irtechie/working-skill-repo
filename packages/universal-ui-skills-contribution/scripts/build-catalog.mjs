#!/usr/bin/env node

import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const packageRoot = path.resolve(scriptDir, "..");
const defaultRepoRoot = path.resolve(packageRoot, "..", "..");
const defaultOutput = path.join(packageRoot, "src", "catalog.generated.js");

const categoriesPath = ["config", "skill-categories.json"];

// Categories are data, not code, so `kbcheck new-skill` can register a skill by
// editing JSON instead of rewriting this module. A projection over a root that
// ships no category file still succeeds and reports "Other"; requiring every
// skill to be mapped is enforced by `kbcheck new-skill --action check`.
async function loadCategories(repoRoot) {
  let raw;
  try {
    raw = await fs.readFile(path.join(repoRoot, ...categoriesPath), "utf8");
  } catch {
    return new Map();
  }
  const parsed = JSON.parse(raw);
  if (!Array.isArray(parsed.categories)) {
    throw new Error(`${categoriesPath.join("/")} must define a categories array.`);
  }
  const categories = new Map();
  for (const entry of parsed.categories) {
    if (typeof entry?.name !== "string" || !Array.isArray(entry.skills)) {
      throw new Error(`${categoriesPath.join("/")} entries need a name and a skills array.`);
    }
    categories.set(entry.name, new Set(entry.skills));
  }
  return categories;
}

function stripYamlQuotes(value) {
  const trimmed = value.trim();
  if (trimmed.startsWith('"') && trimmed.endsWith('"')) {
    try {
      return JSON.parse(trimmed);
    } catch {
      return trimmed.slice(1, -1);
    }
  }
  if (trimmed.startsWith("'") && trimmed.endsWith("'")) {
    return trimmed.slice(1, -1).replaceAll("''", "'");
  }
  return trimmed;
}

function parseFrontmatter(content, sourcePath) {
  const match = content.match(/^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/);
  if (!match) {
    throw new Error(`${sourcePath} must begin with YAML frontmatter.`);
  }
  const fields = {};
  for (const line of match[1].split(/\r?\n/)) {
    const field = line.match(/^([A-Za-z][A-Za-z0-9_-]*):\s*(.*)$/);
    if (field) {
      fields[field[1]] = stripYamlQuotes(field[2]);
    }
  }
  return fields;
}

function assertSafeText(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string.`);
  }
  const executableMarkup =
    /<\s*\/?\s*(?:script|iframe|object|embed|svg|img|link|style)\b|on[a-z]+\s*=|javascript\s*:|data\s*:\s*text\/html|file\s*:/i;
  const privatePath =
    /\b[A-Za-z]:[\\/](?:Users|Documents and Settings)[\\/][^\\/\s]+|\/(?:home|Users)\/[^/\s]+|%USERPROFILE%|\$HOME|~[\\/]\.|\\\\[^\\\s]+\\[^\\\s]+/i;
  if (/[\u0000-\u0008\u000b\u000c\u000e-\u001f]/.test(value)) {
    throw new Error(`${label} contains control characters.`);
  }
  if (executableMarkup.test(value)) {
    throw new Error(`${label} contains executable markup or a dangerous URL.`);
  }
  if (privatePath.test(value)) {
    throw new Error(`${label} contains a machine-private path.`);
  }
}

function categoryForSkill(categories, skillId) {
  for (const [category, skills] of categories) {
    if (skills.has(skillId)) {
      return category;
    }
  }
  return "Other";
}

export function assertSafeProjection(catalog) {
  for (const skill of catalog) {
    for (const [field, value] of Object.entries(skill)) {
      if (value !== null) {
        assertSafeText(String(value), `${skill.id}.${field}`);
      }
    }
    if (!skill.sourcePath.startsWith(".github/skills/") || skill.sourcePath.includes("\\")) {
      throw new Error(`${skill.id}.sourcePath must be repository-relative POSIX form.`);
    }
    if (!skill.sourceUrl.startsWith("https://github.com/Irtechie/working-skill-repo/blob/main/")) {
      throw new Error(`${skill.id}.sourceUrl must use the canonical public repository.`);
    }
  }
  return catalog;
}

export async function collectSkillCatalog(repoRoot = defaultRepoRoot) {
  const skillsRoot = path.join(repoRoot, ".github", "skills");
  const entries = await fs.readdir(skillsRoot, { withFileTypes: true });
  const categories = await loadCategories(repoRoot);
  const catalog = [];

  for (const entry of entries.filter((item) => item.isDirectory()).sort((a, b) => a.name.localeCompare(b.name))) {
    const sourcePath = `.github/skills/${entry.name}/SKILL.md`;
    const content = await fs.readFile(path.join(repoRoot, ...sourcePath.split("/")), "utf8");
    const frontmatter = parseFrontmatter(content, sourcePath);
    if (frontmatter.name !== entry.name) {
      throw new Error(`${sourcePath} name must match its directory identity.`);
    }
    assertSafeText(frontmatter.name, `${entry.name}.name`);
    assertSafeText(frontmatter.description, `${entry.name}.description`);
    if (frontmatter["argument-hint"]) {
      assertSafeText(frontmatter["argument-hint"], `${entry.name}.argument-hint`);
    }
    catalog.push({
      id: entry.name,
      name: frontmatter.name,
      description: frontmatter.description,
      argumentHint: frontmatter["argument-hint"] || null,
      category: categoryForSkill(categories, entry.name),
      sourcePath,
      sourceUrl: `https://github.com/Irtechie/working-skill-repo/blob/main/${sourcePath}`
    });
  }

  return assertSafeProjection(catalog);
}

export function buildCatalogModule(catalog) {
  return `// Generated from canonical SKILL.md frontmatter. Do not edit directly.
export const skillCatalog = Object.freeze(${JSON.stringify(catalog, null, 2)}.map((skill) => Object.freeze(skill)));

export const skillCatalogMetadata = Object.freeze({
  schemaVersion: "working_skill_repo.catalog.v1",
  owner: "Irtechie/working-skill-repo",
  sourceRoot: ".github/skills",
  count: skillCatalog.length
});
`;
}

export async function buildCatalog({ check = false, repoRoot = defaultRepoRoot, output = defaultOutput } = {}) {
  const generated = buildCatalogModule(await collectSkillCatalog(repoRoot));
  if (check) {
    const current = await fs.readFile(output, "utf8");
    if (current !== generated) {
      throw new Error("Generated Skills catalog is stale. Run npm run catalog:build.");
    }
    return;
  }
  await fs.writeFile(output, generated, "utf8");
}

const isDirectRun =
  process.argv[1] && pathToFileURL(path.resolve(process.argv[1])).href === import.meta.url;
if (isDirectRun) {
  await buildCatalog({ check: process.argv.includes("--check") });
  process.stdout.write(`skills-catalog: ${process.argv.includes("--check") ? "current" : "generated"}\n`);
}
