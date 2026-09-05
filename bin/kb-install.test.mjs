import assert from "node:assert/strict";
import crypto from "node:crypto";
import childProcess from "node:child_process";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";
import { promisify } from "node:util";

import {
  installReconciler,
  reconcilerArtifactName,
  uninstallReconciler,
  installRouter,
  routerArtifactName,
  uninstallRouter,
} from "./kb-install.mjs";

const execFile = promisify(childProcess.execFile);
const testDir = path.dirname(fileURLToPath(import.meta.url));

function managedBinaryPath(installRoot, platform = process.platform) {
  return path.join(installRoot, ".kb", "bin", platform === "win32" ? "kbrouter.exe" : "kbrouter");
}

async function fixture(t, { platform = process.platform, arch = process.arch, version = "1.2.3", bytes = "router-v1" } = {}) {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "kb-install-test-"));
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  const releaseRoot = path.join(root, "release");
  const installRoot = path.join(root, "home");
  await fs.mkdir(releaseRoot, { recursive: true });
  const asset = routerArtifactName({ platform, arch });
  const digest = crypto.createHash("sha256").update(bytes).digest("hex");
  await fs.writeFile(path.join(releaseRoot, asset), bytes);
  await fs.writeFile(path.join(releaseRoot, "SHA256SUMS"), `${digest}  ${asset}\n`);
  return { root, releaseRoot, installRoot, platform, arch, version, asset, digest, bytes };
}

test("maps supported operating systems and architectures to release assets", () => {
  assert.equal(routerArtifactName({ platform: "win32", arch: "x64" }), "kbrouter-windows-amd64.exe");
  assert.equal(routerArtifactName({ platform: "darwin", arch: "arm64" }), "kbrouter-darwin-arm64");
  assert.equal(routerArtifactName({ platform: "linux", arch: "x64" }), "kbrouter-linux-amd64");
  assert.throws(() => routerArtifactName({ platform: "freebsd", arch: "x64" }), /unsupported router platform/i);
  assert.throws(() => routerArtifactName({ platform: "linux", arch: "ia32" }), /unsupported router architecture/i);
});

test("minimal experimental profile requires a disposable root and records its omissions", async (t) => {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "kb-minimal-profile-"));
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  const installer = path.join(testDir, "kb-install.mjs");
  await assert.rejects(
    execFile(process.execPath, [installer, "--target", "codex", "--profile", "minimal-experimental", "--router", "skip", "--reconciler", "skip", "--yes"]),
    /requires an explicit --install-root/i,
  );
  await execFile(process.execPath, [installer, "--source", path.resolve(testDir, ".."), "--target", "codex", "--profile", "minimal-experimental", "--install-root", root, "--router", "skip", "--reconciler", "skip", "--yes"]);
  const inventory = JSON.parse(await fs.readFile(path.join(root, "kb-minimal-experimental-inventory.json"), "utf8"));
  assert.deepEqual(inventory.omissions, ["gh-copilot-cost-ops", "kb-simplify"]);
  await assert.rejects(fs.stat(path.join(root, ".codex", "skills", "kb-simplify")), /ENOENT/);
  await fs.stat(path.join(root, ".codex", "skills", "kb-work", "SKILL.md"));
});

test("maps reconciler assets and requires explicit provenance capability", () => {
  assert.equal(reconcilerArtifactName({ platform: "win32", arch: "x64" }), "kbreconcile-windows-amd64.exe");
  assert.equal(reconcilerArtifactName({ platform: "linux", arch: "arm64" }), "kbreconcile-linux-arm64");
  assert.throws(() => reconcilerArtifactName({ platform: "freebsd", arch: "x64" }), /unsupported reconciler platform/i);
});

test("optional reconciler install is checksum managed and reports no privileged provenance", async (t) => {
  const f = await fixture(t, { bytes: "reconciler-v1" });
  const asset = reconcilerArtifactName({ platform: f.platform, arch: f.arch });
  const digest = crypto.createHash("sha256").update(f.bytes).digest("hex");
  await fs.rm(path.join(f.releaseRoot, f.asset));
  await fs.writeFile(path.join(f.releaseRoot, asset), f.bytes);
  await fs.writeFile(path.join(f.releaseRoot, "SHA256SUMS"), `${digest}  ${asset}\n`);
  const result = await installReconciler(f);
  assert.equal(result.status, "installed");
  assert.equal(result.provenance, "checksum-only");
  assert.equal(result.protectedWriterCapable, false);
  assert.equal(await fs.readFile(result.binaryPath, "utf8"), f.bytes);
  const current = await installReconciler(f);
  assert.equal(current.status, "current");
});

test("automatic reconciler install preserves skill-only operation when asset is absent", async (t) => {
  const f = await fixture(t);
  const result = await installReconciler({ ...f, mode: "auto" });
  assert.equal(result.status, "unavailable");
  assert.match(result.reason, /checksum not found|not found/i);
});

test("CLI installs the optional reconciler when router installation is skipped", async (t) => {
  const f = await fixture(t, { bytes: "reconciler-cli" });
  const asset = reconcilerArtifactName({ platform: f.platform, arch: f.arch });
  const digest = crypto.createHash("sha256").update(f.bytes).digest("hex");
  await fs.rm(path.join(f.releaseRoot, f.asset));
  await fs.writeFile(path.join(f.releaseRoot, asset), f.bytes);
  await fs.writeFile(path.join(f.releaseRoot, "SHA256SUMS"), `${digest}  ${asset}\n`);

  const { stdout } = await execFile(process.execPath, [
    path.join(testDir, "kb-install.mjs"),
    "--source", path.resolve(testDir, ".."),
    "--install-root", f.installRoot,
    "--target", "codex",
    "--router", "skip",
    "--reconciler", "required",
    "--reconciler-version", f.version,
    "--reconciler-release", f.releaseRoot,
    "--yes",
  ]);
  assert.match(stdout, /KB reconciler: installed/);
  assert.equal(
    await fs.readFile(path.join(f.installRoot, ".kb", "bin", process.platform === "win32" ? "kbreconcile.exe" : "kbreconcile"), "utf8"),
    f.bytes,
  );
});

test("reconciler upgrade, downgrade refusal, and drift-safe uninstall", async (t) => {
  const first = await fixture(t, { bytes: "reconciler-v1" });
  let asset = reconcilerArtifactName({ platform: first.platform, arch: first.arch });
  let digest = crypto.createHash("sha256").update(first.bytes).digest("hex");
  await fs.rm(path.join(first.releaseRoot, first.asset));
  await fs.writeFile(path.join(first.releaseRoot, asset), first.bytes);
  await fs.writeFile(path.join(first.releaseRoot, "SHA256SUMS"), `${digest}  ${asset}\n`);
  const installed = await installReconciler(first);

  const next = await fixture(t, { version: "1.3.0", bytes: "reconciler-v2" });
  next.installRoot = first.installRoot;
  asset = reconcilerArtifactName({ platform: next.platform, arch: next.arch });
  digest = crypto.createHash("sha256").update(next.bytes).digest("hex");
  await fs.rm(path.join(next.releaseRoot, next.asset));
  await fs.writeFile(path.join(next.releaseRoot, asset), next.bytes);
  await fs.writeFile(path.join(next.releaseRoot, "SHA256SUMS"), `${digest}  ${asset}\n`);
  const upgraded = await installReconciler(next);
  assert.equal(upgraded.status, "upgraded");
  assert.equal(await fs.readFile(upgraded.backupPath, "utf8"), "reconciler-v1");

  await assert.rejects(installReconciler(first), /downgrade/i);
  await assert.rejects(installReconciler({ ...next, version: "1.3.0-beta.1" }), /downgrade/i);
  await fs.writeFile(installed.binaryPath, "operator-change");
  await assert.rejects(uninstallReconciler({ installRoot: first.installRoot }), /changed since/i);
  const removed = await uninstallReconciler({ installRoot: first.installRoot, yes: true });
  assert.equal(await fs.readFile(removed.backupPath, "utf8"), "operator-change");
});

test("native lifecycle fixtures install the native executable basename", async (t) => {
  const f = await fixture(t);
  const installed = await installRouter(f);
  assert.equal(path.basename(installed.binaryPath), process.platform === "win32" ? "kbrouter.exe" : "kbrouter");
  assert.equal(
    f.asset,
    routerArtifactName({ platform: process.platform, arch: process.arch }),
  );
});

test("rejects malformed router versions before reading a release or writing install state", async (t) => {
  const f = await fixture(t, { version: "01.2.3" });
  let fetched = false;
  await assert.rejects(
    installRouter({
      ...f,
      releaseRoot: "https://example.test/releases/v01.2.3",
      fetchImpl: async () => {
        fetched = true;
        throw new Error("must not fetch");
      },
    }),
    /semantic version/i,
  );
  assert.equal(fetched, false);
  await assert.rejects(fs.access(f.installRoot));
});

test("CLI rejects an explicit malformed router version before touching the install root", async (t) => {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "kb-install-version-test-"));
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  const installRoot = path.join(root, "home");

  await assert.rejects(
    execFile(process.execPath, [
      path.join(testDir, "kb-install.mjs"),
      "--source", path.join(root, "missing-source"),
      "--install-root", installRoot,
      "--router-version", "1.2.03",
    ]),
    (error) => {
      assert.match(error.stderr, /strict semantic version/i);
      return true;
    },
  );
  await assert.rejects(fs.access(installRoot));
});

test("remote release roots require HTTPS while filesystem release roots remain supported", async (t) => {
  const f = await fixture(t);
  await assert.rejects(
    installRouter({ ...f, releaseRoot: "http://example.test/release" }),
    /https/i,
  );
  await assert.rejects(
    installRouter({ ...f, releaseRoot: "ftp://example.test/release" }),
    /https/i,
  );

  const installed = await installRouter(f);
  assert.equal(installed.status, "installed");
});

test("rejects an HTTPS redirect that downgrades to a non-HTTPS location", async (t) => {
  const f = await fixture(t);
  const fetchImpl = async () => new Response(null, {
    status: 302,
    headers: { location: "http://example.test/insecure/SHA256SUMS" },
  });

  await assert.rejects(
    installRouter({
      ...f,
      releaseRoot: "https://example.test/release",
      fetchImpl,
    }),
    /redirect.*https|https.*redirect/i,
  );
  await assert.rejects(fs.access(f.installRoot));
});

test("times out a stalled remote release without writing install state", async (t) => {
  const f = await fixture(t);
  const fetchImpl = async (_url, { signal }) => new Promise((resolve, reject) => {
    signal.addEventListener("abort", () => reject(new Error("aborted")), { once: true });
  });

  await assert.rejects(
    installRouter({
      ...f,
      releaseRoot: "https://example.test/release",
      fetchImpl,
      downloadTimeoutMs: 10,
    }),
    /timed out/i,
  );
  await assert.rejects(fs.access(f.installRoot));
});

test("rejects oversized checksum and binary response bodies", async (t) => {
  const f = await fixture(t);
  const checksumBytes = `${f.digest}  ${f.asset}\n`;

  await assert.rejects(
    installRouter({
      ...f,
      releaseRoot: "https://example.test/release",
      maxChecksumBytes: 8,
      fetchImpl: async () => new Response(checksumBytes, {
        status: 200,
        headers: { "content-length": String(Buffer.byteLength(checksumBytes)) },
      }),
    }),
    /exceeds.*byte limit/i,
  );

  await assert.rejects(
    installRouter({
      ...f,
      releaseRoot: "https://example.test/release",
      maxBinaryBytes: 4,
      fetchImpl: async (url) => url.toString().endsWith("SHA256SUMS")
        ? new Response(checksumBytes, { status: 200 })
        : new Response(f.bytes, { status: 200 }),
    }),
    /exceeds.*byte limit/i,
  );
  await assert.rejects(fs.access(f.installRoot));
});

test("installs a verified router and skips the exact same version", async (t) => {
  const f = await fixture(t);
  const first = await installRouter(f);
  assert.equal(first.status, "installed");
  assert.equal(await fs.readFile(first.binaryPath, "utf8"), f.bytes);

  const second = await installRouter(f);
  assert.equal(second.status, "current");
  assert.equal(second.binaryPath, first.binaryPath);
});

test("upgrades with a backup and records the verified replacement", async (t) => {
  const f = await fixture(t);
  const first = await installRouter(f);

  const upgraded = await fixture(t, {
    platform: f.platform,
    arch: f.arch,
    version: "1.3.0",
    bytes: "router-v2",
  });
  upgraded.installRoot = f.installRoot;
  const result = await installRouter(upgraded);

  assert.equal(result.status, "upgraded");
  assert.equal(await fs.readFile(result.binaryPath, "utf8"), "router-v2");
  assert.equal(await fs.readFile(result.backupPath, "utf8"), f.bytes);
  assert.notEqual(result.binaryPath, first.backupPath);
});

test("auto mode degrades safely when a release is missing", async (t) => {
  const f = await fixture(t);
  await fs.rm(f.releaseRoot, { recursive: true });
  const result = await installRouter({ ...f, mode: "auto" });
  assert.equal(result.status, "unavailable");
  assert.match(result.reason, /not found/i);
});

test("CLI preserves the skill-only install when the optional router is unavailable", async (t) => {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "kb-install-cli-test-"));
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  const source = path.join(root, "source");
  const installRoot = path.join(root, "home");
  await fs.mkdir(path.join(source, ".github", "skills", "fixture-skill"), { recursive: true });
  await fs.mkdir(path.join(source, ".github", "agents"), { recursive: true });
  await fs.writeFile(path.join(source, ".github", "skills", "fixture-skill", "SKILL.md"), "fixture\n");

  const { stderr } = await execFile(process.execPath, [
    path.join(testDir, "kb-install.mjs"),
    "--target", "agents",
    "--source", source,
    "--install-root", installRoot,
    "--router-version", "1.2.3",
    "--router-release", path.join(root, "missing-release"),
    "--yes",
  ]);

  assert.match(stderr, /continuing with skill-only install/i);
  assert.equal(
    await fs.readFile(path.join(installRoot, ".agents", "skills", "fixture-skill", "SKILL.md"), "utf8"),
    "fixture\n",
  );
});

test("checksum mismatch never installs and required mode fails", async (t) => {
  const f = await fixture(t);
  await fs.writeFile(path.join(f.releaseRoot, f.asset), "tampered");

  const optional = await installRouter({ ...f, mode: "auto" });
  assert.equal(optional.status, "unavailable");
  assert.match(optional.reason, /checksum mismatch/i);

  await assert.rejects(
    installRouter({ ...f, mode: "required" }),
    /checksum mismatch/i,
  );
});

test("install never replaces an untracked binary without explicit authority", async (t) => {
  const f = await fixture(t);
  const binaryPath = managedBinaryPath(f.installRoot, f.platform);
  await fs.mkdir(path.dirname(binaryPath), { recursive: true });
  await fs.writeFile(binaryPath, "user-router");

  const optional = await installRouter({ ...f, mode: "auto" });
  assert.equal(optional.status, "unavailable");
  assert.match(optional.reason, /untracked/i);
  assert.equal(await fs.readFile(binaryPath, "utf8"), "user-router");

  await assert.rejects(installRouter(f), /untracked/i);

  const authorized = await installRouter({ ...f, yes: true });
  assert.equal(authorized.status, "upgraded");
  assert.equal(await fs.readFile(authorized.backupPath, "utf8"), "user-router");
  assert.equal(await fs.readFile(binaryPath, "utf8"), f.bytes);
});

test("install never replaces a drifted managed binary without explicit authority", async (t) => {
  const f = await fixture(t);
  const installed = await installRouter(f);
  await fs.writeFile(installed.binaryPath, "managed-but-drifted");

  const upgraded = await fixture(t, { version: "1.3.0", bytes: "router-v2" });
  upgraded.installRoot = f.installRoot;
  const optional = await installRouter({ ...upgraded, mode: "auto" });
  assert.equal(optional.status, "unavailable");
  assert.match(optional.reason, /changed since KB installed it/i);
  assert.equal(await fs.readFile(installed.binaryPath, "utf8"), "managed-but-drifted");

  const authorized = await installRouter({ ...upgraded, yes: true });
  assert.equal(authorized.status, "upgraded");
  assert.equal(await fs.readFile(authorized.backupPath, "utf8"), "managed-but-drifted");
});

test("invalid install state requires explicit authority before replacement", async (t) => {
  const f = await fixture(t);
  const routerDir = path.join(f.installRoot, ".kb", "bin");
  await fs.mkdir(routerDir, { recursive: true });
  await fs.writeFile(managedBinaryPath(f.installRoot, f.platform), "existing");
  await fs.writeFile(path.join(routerDir, ".kbrouter-install.json"), JSON.stringify({
    schema_version: 1,
    version: "",
    binary_name: "other",
    sha256: "bad",
  }));

  await assert.rejects(installRouter(f), /invalid KB install state/i);
  const authorized = await installRouter({ ...f, yes: true });
  assert.equal(authorized.status, "upgraded");
  assert.equal(await fs.readFile(authorized.backupPath, "utf8"), "existing");
});

test("uninstall removes only an unchanged managed binary", async (t) => {
  const f = await fixture(t);
  const installed = await installRouter(f);
  const result = await uninstallRouter({ installRoot: f.installRoot });
  assert.equal(result.status, "uninstalled");
  await assert.rejects(fs.access(installed.binaryPath));
});

test("uninstall preserves drift unless explicit backup authority is given", async (t) => {
  const f = await fixture(t);
  const installed = await installRouter(f);
  await fs.writeFile(installed.binaryPath, "user-change");

  await assert.rejects(
    uninstallRouter({ installRoot: f.installRoot }),
    /changed since KB installed it/i,
  );
  assert.equal(await fs.readFile(installed.binaryPath, "utf8"), "user-change");

  const result = await uninstallRouter({ installRoot: f.installRoot, yes: true });
  assert.equal(result.status, "uninstalled");
  assert.equal(await fs.readFile(result.backupPath, "utf8"), "user-change");
});

test("uninstall refuses malformed or misdirected managed state", async (t) => {
  const f = await fixture(t);
  await installRouter(f);
  const statePath = path.join(f.installRoot, ".kb", "bin", ".kbrouter-install.json");

  for (const invalid of [
    { schema_version: 1, version: "1.2.3", binary_name: "other", sha256: f.digest },
    { schema_version: 1, version: "1.2.3", binary_name: "kbrouter", sha256: "bad" },
    { schema_version: 1, version: "", binary_name: "kbrouter", sha256: f.digest },
  ]) {
    await fs.writeFile(statePath, JSON.stringify(invalid));
    await assert.rejects(uninstallRouter({ installRoot: f.installRoot, yes: true }), /install state is invalid/i);
  }
});
