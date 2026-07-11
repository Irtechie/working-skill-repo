#!/usr/bin/env node
import crypto from "node:crypto";
import fs from "node:fs/promises";
import path from "node:path";
import os from "node:os";
import process from "node:process";
import readline from "node:readline/promises";
import { fileURLToPath, pathToFileURL } from "node:url";

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
// Add domain-specific or optional skills here when needed.
const FULL_ONLY_SKILLS = new Set([]);

const VALID_TARGETS = new Set(["codex", "copilot", "agents", "repo", "all"]);
const VALID_PROFILES = new Set(["core", "full"]);
const VALID_ROUTER_MODES = new Set(["auto", "required", "skip", "uninstall"]);
const ROUTER_STATE_SCHEMA = 1;
const DEFAULT_RELEASE_BASE = "https://github.com/Irtechie/working-skill-repo/releases/download";
const DEFAULT_DOWNLOAD_TIMEOUT_MS = 30_000;
const DEFAULT_MAX_CHECKSUM_BYTES = 1024 * 1024;
const DEFAULT_MAX_BINARY_BYTES = 256 * 1024 * 1024;
const MAX_REDIRECTS = 5;
const STRICT_SEMVER = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9]\d*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$/;
const URL_SCHEME = /^[A-Za-z][A-Za-z0-9+.-]*:\/\//;

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
  --router <auto|required|skip|uninstall>   Optional native router lifecycle. Default: auto
  --router-version <version>                Router release version. Default: package version
  --router-release <url-or-path>            Release directory override
  --router-dir <path>                       Binary directory. Default: <install-root>/.kb/bin
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
    routerMode: "auto",
    routerVersion: "",
    routerRelease: "",
    routerDir: "",
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
    } else if (arg === "--router") {
      args.routerMode = requireValue(argv, ++i, arg);
    } else if (arg.startsWith("--router=")) {
      args.routerMode = arg.slice("--router=".length);
    } else if (arg === "--router-version") {
      args.routerVersion = requireValue(argv, ++i, arg);
    } else if (arg.startsWith("--router-version=")) {
      args.routerVersion = arg.slice("--router-version=".length);
    } else if (arg === "--router-release") {
      args.routerRelease = requireValue(argv, ++i, arg);
    } else if (arg.startsWith("--router-release=")) {
      args.routerRelease = arg.slice("--router-release=".length);
    } else if (arg === "--router-dir") {
      args.routerDir = requireValue(argv, ++i, arg);
    } else if (arg.startsWith("--router-dir=")) {
      args.routerDir = arg.slice("--router-dir=".length);
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
  if (!VALID_ROUTER_MODES.has(args.routerMode)) {
    throw new Error(`Invalid --router '${args.routerMode}'. Use auto, required, skip, or uninstall.`);
  }
  if (args.target === "repo" && !args.repo) {
    throw new Error("--target repo requires --repo <path>.");
  }
  if (args.routerVersion && !STRICT_SEMVER.test(args.routerVersion)) {
    throw new Error("--router-version must be a strict semantic version.");
  }

  args.source = path.resolve(expandHome(args.source));
  args.installRoot = path.resolve(expandHome(args.installRoot));
  if (args.repo) {
    args.repo = path.resolve(expandHome(args.repo));
  }
  if (args.routerDir) {
    args.routerDir = path.resolve(expandHome(args.routerDir));
  }
  if (args.routerRelease) {
    if (URL_SCHEME.test(args.routerRelease)) {
      assertHttpsReleaseRoot(args.routerRelease);
    } else {
      args.routerRelease = path.resolve(expandHome(args.routerRelease));
    }
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

  if (args.routerMode === "uninstall") {
    const result = await uninstallRouter({
      installRoot: args.installRoot,
      routerDir: args.routerDir,
      yes: args.yes,
      dryRun: args.dryRun,
    });
    console.log(`KB router: ${result.status}${result.backupPath ? ` (backup: ${result.backupPath})` : ""}`);
    return;
  }

  await assertSource(args.source);
  let routerResult = { status: "skipped" };
  if (args.routerMode !== "skip") {
    const version = args.routerVersion || await packageVersion(args.source);
    routerResult = await installRouter({
      installRoot: args.installRoot,
      routerDir: args.routerDir,
      releaseRoot: args.routerRelease || `${DEFAULT_RELEASE_BASE}/v${version}`,
      version,
      mode: args.routerMode,
      yes: args.yes,
      dryRun: args.dryRun,
    });
    if (routerResult.status === "unavailable") {
      console.warn(`KB router unavailable; continuing with skill-only install: ${routerResult.reason}`);
    } else {
      console.log(`KB router: ${routerResult.status} (${routerResult.binaryPath})`);
    }
  }
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

async function packageVersion(root) {
  const parsed = JSON.parse(await fs.readFile(path.join(root, "package.json"), "utf8"));
  if (typeof parsed.version !== "string" || !STRICT_SEMVER.test(parsed.version)) {
    throw new Error("package.json must contain a strict semantic release version.");
  }
  return parsed.version;
}

export function routerArtifactName({ platform = process.platform, arch = process.arch } = {}) {
  const platforms = { win32: "windows", darwin: "darwin", linux: "linux" };
  const architectures = { x64: "amd64", arm64: "arm64" };
  const releasePlatform = platforms[platform];
  const releaseArchitecture = architectures[arch];
  if (!releasePlatform) {
    throw new Error(`Unsupported router platform: ${platform}`);
  }
  if (!releaseArchitecture) {
    throw new Error(`Unsupported router architecture: ${arch}`);
  }
  const extension = platform === "win32" ? ".exe" : "";
  return `kbrouter-${releasePlatform}-${releaseArchitecture}${extension}`;
}

export async function installRouter(options = {}) {
  const mode = options.mode || "required";
  if (mode !== "auto" && mode !== "required") {
    throw new Error(`Router install mode must be auto or required, got '${mode}'.`);
  }

  try {
    const installRoot = path.resolve(options.installRoot || os.homedir());
    const routerDir = path.resolve(options.routerDir || path.join(installRoot, ".kb", "bin"));
    const version = options.version;
    if (typeof version !== "string" || !STRICT_SEMVER.test(version)) {
      throw new Error("Router version must be a strict semantic version.");
    }
    const targetPlatform = options.platform || process.platform;
    const asset = routerArtifactName({ platform: targetPlatform, arch: options.arch });
    const releaseRoot = options.releaseRoot;
    if (typeof releaseRoot !== "string" || releaseRoot.trim() === "") {
      throw new Error("Router release location is required.");
    }
    const remoteRelease = URL_SCHEME.test(releaseRoot);
    if (remoteRelease) {
      assertHttpsReleaseRoot(releaseRoot);
    }
    const downloadOptions = {
      fetchImpl: options.fetchImpl || globalThis.fetch,
      timeoutMs: positiveIntegerOption(options.downloadTimeoutMs, DEFAULT_DOWNLOAD_TIMEOUT_MS, "download timeout"),
      maxChecksumBytes: positiveIntegerOption(options.maxChecksumBytes, DEFAULT_MAX_CHECKSUM_BYTES, "checksum response limit"),
      maxBinaryBytes: positiveIntegerOption(options.maxBinaryBytes, DEFAULT_MAX_BINARY_BYTES, "binary response limit"),
    };

    const checksums = parseChecksums((await readReleaseFile(releaseRoot, "SHA256SUMS", {
      fetchImpl: downloadOptions.fetchImpl,
      timeoutMs: downloadOptions.timeoutMs,
      maxBytes: downloadOptions.maxChecksumBytes,
    })).toString("utf8"));
    const expected = checksums.get(asset);
    if (!expected) {
      throw new Error(`Release checksum not found for ${asset}.`);
    }
    const bytes = await readReleaseFile(releaseRoot, asset, {
      fetchImpl: downloadOptions.fetchImpl,
      timeoutMs: downloadOptions.timeoutMs,
      maxBytes: downloadOptions.maxBinaryBytes,
    });
    const actual = sha256(bytes);
    if (actual !== expected) {
      throw new Error(`Checksum mismatch for ${asset}: expected ${expected}, got ${actual}.`);
    }

    const binaryName = targetPlatform === "win32" ? "kbrouter.exe" : "kbrouter";
    const binaryPath = path.join(routerDir, binaryName);
    const statePath = path.join(routerDir, ".kbrouter-install.json");
    const existingHash = await fileHashIfExists(binaryPath);
    const stateRecord = await loadRouterState(statePath);
    const existingState = stateRecord.state;
    const unchangedManaged = Boolean(
      existingHash && stateRecord.valid && existingState.sha256 === existingHash,
    );
    if (existingHash && !unchangedManaged && !options.yes) {
      if (stateRecord.missing) {
        throw new Error("Existing kbrouter binary is untracked; rerun with --yes to back it up before replacement.");
      }
      if (!stateRecord.valid) {
        throw new Error(`Existing kbrouter has invalid KB install state; rerun with --yes to back it up before replacement: ${stateRecord.reason}`);
      }
      throw new Error("Existing kbrouter binary changed since KB installed it; rerun with --yes to back it up before replacement.");
    }
    if (!existingHash && !stateRecord.missing && !options.yes) {
      throw new Error(`Existing KB install state is inconsistent or invalid; rerun with --yes to back it up before replacement: ${stateRecord.reason || "managed binary is missing"}`);
    }
    if (unchangedManaged && existingState.version === version && existingState.sha256 === expected) {
      return { status: "current", binaryPath, version, sha256: expected };
    }
    if (options.dryRun) {
      return { status: existingHash ? "would-upgrade" : "would-install", binaryPath, version, sha256: expected };
    }

    await fs.mkdir(routerDir, { recursive: true, mode: 0o700 });
    const tempPath = path.join(routerDir, `.${binaryName}.${process.pid}.tmp`);
    await fs.writeFile(tempPath, bytes, { flag: "wx", mode: 0o700 });
    let backupPath = "";
    let stateBackupPath = "";
    try {
      if (existingHash) {
        backupPath = await managedBackupPath(installRoot, binaryName);
        await fs.mkdir(path.dirname(backupPath), { recursive: true, mode: 0o700 });
        await fs.rename(binaryPath, backupPath);
      }
      if (!stateRecord.missing && !unchangedManaged) {
        stateBackupPath = await managedBackupPath(installRoot, ".kbrouter-install.json");
        await fs.mkdir(path.dirname(stateBackupPath), { recursive: true, mode: 0o700 });
        await fs.rename(statePath, stateBackupPath);
      }
      await fs.rename(tempPath, binaryPath);
      if (process.platform !== "win32") {
        await fs.chmod(binaryPath, 0o700);
      }
      await writeJsonAtomic(statePath, {
        schema_version: ROUTER_STATE_SCHEMA,
        version,
        asset,
        binary_name: binaryName,
        sha256: expected,
      });
    } catch (error) {
      await fs.rm(tempPath, { force: true });
      await fs.rm(binaryPath, { force: true });
      if (backupPath && await exists(backupPath)) {
        await fs.rename(backupPath, binaryPath);
      }
      if (stateBackupPath && await exists(stateBackupPath)) {
        await fs.rm(statePath, { force: true });
        await fs.rename(stateBackupPath, statePath);
      }
      throw error;
    }

    return {
      status: existingHash ? "upgraded" : "installed",
      binaryPath,
      backupPath,
      version,
      sha256: expected,
    };
  } catch (error) {
    if (mode === "auto") {
      return { status: "unavailable", reason: error.message };
    }
    throw error;
  }
}

export async function uninstallRouter(options = {}) {
  const installRoot = path.resolve(options.installRoot || os.homedir());
  const routerDir = path.resolve(options.routerDir || path.join(installRoot, ".kb", "bin"));
  const statePath = path.join(routerDir, ".kbrouter-install.json");
  const stateRecord = await loadRouterState(statePath);
  if (stateRecord.missing) {
    return { status: "absent" };
  }
  if (!stateRecord.valid) {
    throw new Error(`Router install state is invalid; refusing uninstall: ${stateRecord.reason}`);
  }
  const state = stateRecord.state;
  const binaryPath = path.join(routerDir, state.binary_name);
  const currentHash = await fileHashIfExists(binaryPath);
  if (!currentHash) {
    if (!options.dryRun) {
      await fs.rm(statePath, { force: true });
    }
    return { status: "absent" };
  }
  let backupPath = "";
  if (currentHash !== state.sha256) {
    if (!options.yes) {
      throw new Error("Router binary changed since KB installed it; rerun with --yes to back it up before uninstall.");
    }
    backupPath = await managedBackupPath(installRoot, state.binary_name);
  }
  if (options.dryRun) {
    return { status: "would-uninstall", binaryPath, backupPath };
  }
  if (backupPath) {
    await fs.mkdir(path.dirname(backupPath), { recursive: true, mode: 0o700 });
    await fs.rename(binaryPath, backupPath);
  } else {
    await fs.rm(binaryPath);
  }
  await fs.rm(statePath, { force: true });
  return { status: "uninstalled", binaryPath, backupPath };
}

function parseChecksums(text) {
  const checksums = new Map();
  for (const rawLine of text.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line) continue;
    const match = /^([a-fA-F0-9]{64})\s+\*?([^\s]+)$/.exec(line);
    if (!match || path.basename(match[2]) !== match[2]) {
      throw new Error("Invalid SHA256SUMS entry.");
    }
    if (checksums.has(match[2])) {
      throw new Error(`Duplicate SHA256SUMS entry for ${match[2]}.`);
    }
    checksums.set(match[2], match[1].toLowerCase());
  }
  return checksums;
}

function positiveIntegerOption(value, fallback, label) {
  const resolved = value === undefined ? fallback : value;
  if (!Number.isSafeInteger(resolved) || resolved <= 0) {
    throw new Error(`${label} must be a positive finite integer.`);
  }
  return resolved;
}

function assertHttpsReleaseRoot(root) {
  let parsed;
  try {
    parsed = new URL(root);
  } catch (error) {
    throw new Error(`Router release URL is invalid: ${error.message}`);
  }
  if (parsed.protocol !== "https:") {
    throw new Error("Remote router release roots must use HTTPS.");
  }
  if (parsed.username || parsed.password) {
    throw new Error("Remote router release roots must not contain credentials.");
  }
  return parsed;
}

async function readReleaseFile(root, name, options) {
  try {
    if (URL_SCHEME.test(root)) {
      const releaseRoot = assertHttpsReleaseRoot(root);
      const artifactUrl = new URL(name, `${releaseRoot.href.replace(/\/$/, "")}/`);
      return await readHttpsResponse(artifactUrl, options);
    }
    const artifactPath = path.join(root, name);
    const stat = await fs.stat(artifactPath);
    if (stat.size > options.maxBytes) {
      throw new Error(`response exceeds ${options.maxBytes} byte limit`);
    }
    const bytes = await fs.readFile(artifactPath);
    if (bytes.length > options.maxBytes) {
      throw new Error(`response exceeds ${options.maxBytes} byte limit`);
    }
    return bytes;
  } catch (error) {
    throw new Error(`Release artifact not found (${name}): ${error.message}`);
  }
}

async function readHttpsResponse(initialUrl, options) {
  if (typeof options.fetchImpl !== "function") {
    throw new Error("HTTPS release downloads require a fetch implementation.");
  }
  const controller = new AbortController();
  let timeout;
  const download = (async () => {
    let currentUrl = initialUrl;
    for (let redirectCount = 0; redirectCount <= MAX_REDIRECTS; redirectCount += 1) {
      const response = await options.fetchImpl(currentUrl, {
        redirect: "manual",
        signal: controller.signal,
      });
      const responseUrl = response.url ? new URL(response.url) : currentUrl;
      if (responseUrl.protocol !== "https:") {
        throw new Error("Release response resolved to a non-HTTPS URL.");
      }
      if ([301, 302, 303, 307, 308].includes(response.status)) {
        if (redirectCount === MAX_REDIRECTS) {
          throw new Error(`Release download exceeded ${MAX_REDIRECTS} redirects.`);
        }
        const location = response.headers?.get("location");
        if (!location) {
          throw new Error(`HTTP ${response.status} redirect is missing a Location header.`);
        }
        const redirectUrl = new URL(location, responseUrl);
        if (redirectUrl.protocol !== "https:") {
          throw new Error("HTTPS release redirect to a non-HTTPS URL is not allowed.");
        }
        currentUrl = redirectUrl;
        continue;
      }
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      return await readBoundedResponseBody(response, options.maxBytes);
    }
    throw new Error(`Release download exceeded ${MAX_REDIRECTS} redirects.`);
  })();
  const deadline = new Promise((_, reject) => {
    timeout = setTimeout(() => {
      reject(new Error(`Release download timed out after ${options.timeoutMs} ms.`));
      controller.abort();
    }, options.timeoutMs);
  });
  try {
    return await Promise.race([download, deadline]);
  } finally {
    clearTimeout(timeout);
  }
}

async function readBoundedResponseBody(response, maxBytes) {
  const declaredLength = response.headers?.get("content-length");
  if (declaredLength !== null && declaredLength !== undefined) {
    if (!/^\d+$/.test(declaredLength)) {
      throw new Error("Release response has an invalid Content-Length header.");
    }
    if (Number(declaredLength) > maxBytes) {
      throw new Error(`response exceeds ${maxBytes} byte limit`);
    }
  }

  if (response.body && typeof response.body.getReader === "function") {
    const reader = response.body.getReader();
    const chunks = [];
    let total = 0;
    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        total += value.byteLength;
        if (total > maxBytes) {
          await reader.cancel("response byte limit exceeded").catch(() => {});
          throw new Error(`response exceeds ${maxBytes} byte limit`);
        }
        chunks.push(Buffer.from(value));
      }
    } finally {
      reader.releaseLock();
    }
    return Buffer.concat(chunks, total);
  }

  const bytes = Buffer.from(await response.arrayBuffer());
  if (bytes.length > maxBytes) {
    throw new Error(`response exceeds ${maxBytes} byte limit`);
  }
  return bytes;
}

async function loadRouterState(statePath) {
  try {
    const parsed = JSON.parse(await fs.readFile(statePath, "utf8"));
    const reason = validateRouterState(parsed);
    if (reason) {
      return { state: parsed, valid: false, missing: false, reason };
    }
    return {
      state: { ...parsed, sha256: parsed.sha256.toLowerCase() },
      valid: true,
      missing: false,
      reason: "",
    };
  } catch (error) {
    if (error.code === "ENOENT") {
      return { state: null, valid: false, missing: true, reason: "state is absent" };
    }
    return { state: null, valid: false, missing: false, reason: `state is unreadable: ${error.message}` };
  }
}

function validateRouterState(state) {
  if (!state || typeof state !== "object" || Array.isArray(state)) {
    return "state must be an object";
  }
  if (state.schema_version !== ROUTER_STATE_SCHEMA) {
    return `unsupported schema_version '${state.schema_version}'`;
  }
  if (state.binary_name !== "kbrouter" && state.binary_name !== "kbrouter.exe") {
    return "binary_name must be exactly kbrouter or kbrouter.exe";
  }
  if (typeof state.sha256 !== "string" || !/^[a-fA-F0-9]{64}$/.test(state.sha256)) {
    return "sha256 must contain exactly 64 hexadecimal characters";
  }
  if (typeof state.version !== "string" || !STRICT_SEMVER.test(state.version)) {
    return "version must be a semantic release version";
  }
  if (typeof state.asset !== "string" || !/^kbrouter-(?:linux|darwin)-(?:amd64|arm64)$|^kbrouter-windows-(?:amd64|arm64)\.exe$/.test(state.asset)) {
    return "asset is not a supported kbrouter release filename";
  }
  if ((state.asset.endsWith(".exe")) !== (state.binary_name.endsWith(".exe"))) {
    return "binary_name does not match the release asset platform";
  }
  return "";
}

async function writeJsonAtomic(destination, value) {
  const temp = `${destination}.${process.pid}.tmp`;
  const previous = `${destination}.${process.pid}.previous`;
  await fs.writeFile(temp, `${JSON.stringify(value, null, 2)}\n`, { flag: "wx", mode: 0o600 });
  let movedPrevious = false;
  try {
    if (await exists(destination)) {
      await fs.rename(destination, previous);
      movedPrevious = true;
    }
    await fs.rename(temp, destination);
    if (movedPrevious) {
      await fs.rm(previous, { force: true }).catch(() => {});
    }
  } catch (error) {
    await fs.rm(temp, { force: true });
    if (movedPrevious && !(await exists(destination)) && await exists(previous)) {
      await fs.rename(previous, destination);
    }
    throw error;
  }
}

async function managedBackupPath(installRoot, filename) {
  const stamp = new Date().toISOString().replace(/[:.]/g, "-");
  const base = path.join(installRoot, ".kb", "install-backups", "router", stamp, filename);
  let candidate = base;
  let suffix = 1;
  while (await exists(candidate)) {
    candidate = `${base}-${suffix}`;
    suffix += 1;
  }
  return candidate;
}

async function fileHashIfExists(target) {
  try {
    return sha256(await fs.readFile(target));
  } catch (error) {
    if (error.code === "ENOENT") return "";
    throw error;
  }
}

function sha256(bytes) {
  return crypto.createHash("sha256").update(bytes).digest("hex");
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

if (process.argv[1] && import.meta.url === pathToFileURL(path.resolve(process.argv[1])).href) {
  main().catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
  });
}
