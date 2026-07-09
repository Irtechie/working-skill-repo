#!/usr/bin/env node
import crypto from "node:crypto";
import fs from "node:fs/promises";
import path from "node:path";
import os from "node:os";
import process from "node:process";
import readline from "node:readline/promises";
import { fileURLToPath } from "node:url";

const CORE_AGENTS = [
  "adversarial-document-reviewer",
  "coherence-reviewer",
  "correctness-reviewer",
  "design-lens-reviewer",
  "feasibility-reviewer",
  "product-lens-reviewer",
  "project-standards-reviewer",
  "scope-guardian-reviewer",
  "security-lens-reviewer",
  "testing-reviewer",
  "thermo-nuclear-code-quality-reviewer",
];

// Skills excluded from the core profile — only installed with --profile full.
// Add domain-specific or optional skills here.
const FULL_ONLY_SKILLS = new Set([
  "gh-copilot-cost-ops",
]);

const VALID_TARGETS = new Set(["codex", "copilot", "agents", "repo", "all"]);
const VALID_PROFILES = new Set(["core", "full"]);

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const sourceRoot = path.resolve(__dirname, "..");

function usage() {
  return `KB skill installer

Usage:
  npx github:Irtechie/working-skill-repo --target all --profile core
  node ./bin/kb-install.mjs --target codex --profile full

Options:
  --target <codex|copilot|agents|repo|all>  Install target. Default: all
  --profile <core|full>                     Skill profile. Default: core
  --repo <path>                             Repo-local install root for --target repo
  --install-root <path>                     Home/root override for global installs
  --source <path>                           Source repo override. Default: current package
  --yes                                    Back up and overwrite changed existing files
  --dry-run                                Print actions without writing
  --help                                   Show this help

Core installs every runtime skill plus baseline review/document agents.
Full installs every runtime skill plus every reviewer/specialist agent for Codex, Copilot, and repo-local targets.
The Go gate and marketplace are maintainer tools; they are not required to use the skills.
`;
}

function parseArgs(argv) {
  const args = {
    target: "all",
    profile: "core",
    repo: "",
    installRoot: os.homedir(),
    source: sourceRoot,
    yes: false,
    dryRun: false,
  };

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--help" || arg === "-h") {
      args.help = true;
    } else if (arg === "--yes" || arg === "-y") {
      args.yes = true;
    } else if (arg === "--dry-run") {
      args.dryRun = true;
    } else if (arg === "--target") {
      args.target = requireValue(argv, ++i, arg);
    } else if (arg.startsWith("--target=")) {
      args.target = arg.slice("--target=".length);
    } else if (arg === "--profile") {
      args.profile = requireValue(argv, ++i, arg);
    } else if (arg.startsWith("--profile=")) {
      args.profile = arg.slice("--profile=".length);
    } else if (arg === "--repo") {
      args.repo = requireValue(argv, ++i, arg);
    } else if (arg.startsWith("--repo=")) {
      args.repo = arg.slice("--repo=".length);
    } else if (arg === "--install-root") {
      args.installRoot = requireValue(argv, ++i, arg);
    } else if (arg.startsWith("--install-root=")) {
      args.installRoot = arg.slice("--install-root=".length);
    } else if (arg === "--source") {
      args.source = requireValue(argv, ++i, arg);
    } else if (arg.startsWith("--source=")) {
      args.source = arg.slice("--source=".length);
    } else {
      throw new Error(`Unknown argument: ${arg}`);
    }
  }

  if (!VALID_TARGETS.has(args.target)) {
    throw new Error(`Invalid --target '${args.target}'. Use one of: ${[...VALID_TARGETS].join(", ")}`);
  }
  if (!VALID_PROFILES.has(args.profile)) {
    throw new Error(`Invalid --profile '${args.profile}'. Use core or full.`);
  }
  if (args.target === "repo" && !args.repo) {
    throw new Error("--target repo requires --repo <path>.");
  }

  args.source = path.resolve(expandHome(args.source));
  args.installRoot = path.resolve(expandHome(args.installRoot));
  if (args.repo) {
    args.repo = path.resolve(expandHome(args.repo));
  }
  return args;
}

function requireValue(argv, index, flag) {
  const value = argv[index];
  if (!value || value.startsWith("--")) {
    throw new Error(`${flag} requires a value.`);
  }
  return value;
}

function expandHome(value) {
  if (value === "~") {
    return os.homedir();
  }
  if (value.startsWith(`~${path.sep}`) || value.startsWith("~/")) {
    return path.join(os.homedir(), value.slice(2));
  }
  return value;
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help) {
    process.stdout.write(usage());
    return;
  }

  await assertSource(args.source);
  const plan = await buildInstallPlan(args);
  if (plan.length === 0) {
    throw new Error("No install actions were planned.");
  }

  const rl = process.stdin.isTTY && !args.yes
    ? readline.createInterface({ input: process.stdin, output: process.stdout })
    : null;

  try {
    const summary = { copied: 0, skipped: 0, backedUp: 0, conflicts: 0 };
    for (const item of plan) {
      const result = await installItem(item, args, rl);
      summary[result] += 1;
      if (result === "backedUp") {
        summary.copied += 1;
      }
      if (result === "conflicts") {
        summary.conflicts += 1;
      }
    }

    if (summary.conflicts > 0) {
      throw new Error(`Install stopped with ${summary.conflicts} unresolved conflict(s). Rerun with --yes to back up and overwrite.`);
    }

    console.log(`Installed KB ${args.profile} profile to '${args.target}'. copied=${summary.copied} skipped=${summary.skipped} backups=${summary.backedUp}`);
  } finally {
    rl?.close();
  }
}

async function assertSource(root) {
  await requirePath(path.join(root, ".github", "skills"));
  await requirePath(path.join(root, ".github", "agents"));
}

async function requirePath(target) {
  try {
    await fs.access(target);
  } catch {
    throw new Error(`Required path not found: ${target}`);
  }
}

async function buildInstallPlan(args) {
  const targets = args.target === "all" ? ["codex", "copilot", "agents"] : [args.target];
  const allSkills = await listDirectories(path.join(args.source, ".github", "skills"));
  const skills = args.profile === "core"
    ? allSkills.filter(s => !FULL_ONLY_SKILLS.has(s))
    : allSkills;
  const items = [];

  for (const target of targets) {
    const roots = targetRoots(target, args);
    for (const skill of skills) {
      items.push({
        kind: "skill",
        source: path.join(args.source, ".github", "skills", skill),
        destination: path.join(roots.skills, skill),
        backupRoot: roots.backups,
      });
    }

    if (roots.agents) {
      if (args.profile === "full") {
        items.push({
          kind: "agents",
          source: path.join(args.source, ".github", "agents"),
          destination: roots.agents,
          backupRoot: roots.backups,
        });
      } else {
        for (const agent of CORE_AGENTS) {
          items.push({
            kind: "agent",
            source: path.join(args.source, ".github", "agents", `${agent}.agent.md`),
            destination: path.join(roots.agents, `${agent}.agent.md`),
            backupRoot: roots.backups,
          });
        }
      }
    }

    if (target === "repo") {
      items.push({
        kind: "file",
        source: path.join(args.source, "AGENTS.md"),
        destination: path.join(args.repo, "AGENTS.md"),
        backupRoot: roots.backups,
      });
      items.push({
        kind: "file",
        source: path.join(args.source, ".github", "copilot-instructions.md"),
        destination: path.join(args.repo, ".github", "copilot-instructions.md"),
        backupRoot: roots.backups,
      });
    }
  }

  return items;
}

function targetRoots(target, args) {
  if (target === "codex") {
    return {
      skills: path.join(args.installRoot, ".codex", "skills"),
      agents: path.join(args.installRoot, ".codex", "agents"),
      backups: path.join(args.installRoot, ".codex", ".kb-install-backups"),
    };
  }
  if (target === "copilot") {
    return {
      skills: path.join(args.installRoot, ".copilot", "skills"),
      agents: path.join(args.installRoot, ".copilot", "agents"),
      backups: path.join(args.installRoot, ".copilot", ".kb-install-backups"),
    };
  }
  if (target === "agents") {
    return {
      skills: path.join(args.installRoot, ".agents", "skills"),
      backups: path.join(args.installRoot, ".agents", ".kb-install-backups"),
    };
  }
  if (target === "repo") {
    return {
      skills: path.join(args.repo, ".github", "skills"),
      agents: path.join(args.repo, ".github", "agents"),
      backups: path.join(args.repo, ".github", ".kb-install-backups"),
    };
  }
  throw new Error(`Unhandled target: ${target}`);
}

async function listDirectories(root) {
  const entries = await fs.readdir(root, { withFileTypes: true });
  return entries.filter((entry) => entry.isDirectory()).map((entry) => entry.name).sort();
}

async function installItem(item, args, rl) {
  await requirePath(item.source);
  const sourceHash = await treeHash(item.source);
  const destinationExists = await exists(item.destination);

  if (destinationExists) {
    const destinationHash = await treeHash(item.destination);
    if (sourceHash === destinationHash) {
      console.log(`skip same ${item.kind}: ${item.destination}`);
      return "skipped";
    }

    if (args.dryRun) {
      console.log(`would back up and replace ${item.kind}: ${item.destination}`);
      return "skipped";
    }

    const approved = args.yes || await confirm(rl, `Replace changed ${item.kind} at ${item.destination}? A backup will be written. [y/N] `);
    if (!approved) {
      console.log(`conflict left unchanged: ${item.destination}`);
      return "conflicts";
    }

    const backup = await backupPath(item);
    await fs.mkdir(path.dirname(backup), { recursive: true });
    await fs.rename(item.destination, backup);
    await copyPath(item.source, item.destination);
    console.log(`replaced ${item.kind}: ${item.destination} (backup: ${backup})`);
    return "backedUp";
  }

  if (args.dryRun) {
    console.log(`would copy ${item.kind}: ${item.destination}`);
    return "skipped";
  }

  await copyPath(item.source, item.destination);
  console.log(`copied ${item.kind}: ${item.destination}`);
  return "copied";
}

async function exists(target) {
  try {
    await fs.access(target);
    return true;
  } catch {
    return false;
  }
}

async function copyPath(source, destination) {
  await fs.mkdir(path.dirname(destination), { recursive: true });
  const stat = await fs.stat(source);
  if (stat.isDirectory()) {
    await fs.cp(source, destination, { recursive: true });
  } else {
    await fs.copyFile(source, destination);
  }
}

async function backupPath(item) {
  const stamp = new Date().toISOString().replace(/[:.]/g, "-");
  const base = path.join(item.backupRoot, stamp, path.basename(item.destination));
  let candidate = base;
  let suffix = 1;
  while (await exists(candidate)) {
    candidate = `${base}-${suffix}`;
    suffix += 1;
  }
  return candidate;
}

async function confirm(rl, question) {
  if (!rl) {
    return false;
  }
  const answer = await rl.question(question);
  return answer.trim().toLowerCase() === "y" || answer.trim().toLowerCase() === "yes";
}

async function treeHash(target) {
  const stat = await fs.stat(target);
  const hash = crypto.createHash("sha256");
  if (stat.isDirectory()) {
    hash.update("dir\n");
    const files = await listFiles(target);
    for (const file of files) {
      const relative = path.relative(target, file).split(path.sep).join("/");
      hash.update(relative);
      hash.update("\0");
      hash.update(await fs.readFile(file));
      hash.update("\0");
    }
  } else {
    hash.update("file\n");
    hash.update(await fs.readFile(target));
  }
  return hash.digest("hex");
}

async function listFiles(root) {
  const entries = await fs.readdir(root, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    const fullPath = path.join(root, entry.name);
    if (entry.isDirectory()) {
      files.push(...await listFiles(fullPath));
    } else if (entry.isFile()) {
      files.push(fullPath);
    }
  }
  return files.sort();
}

main().catch((error) => {
  console.error(error.message);
  process.exitCode = 1;
});
