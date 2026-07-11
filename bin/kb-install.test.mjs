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
