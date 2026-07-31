#!/usr/bin/env node

import { createHash } from "node:crypto";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { buildCatalog } from "./build-catalog.mjs";
import { runNpm } from "./npm-cli.mjs";
import { contributionDefinition } from "../src/manifest.js";
import { RELEASE_LOCK_SCHEMA_VERSION, validateReleaseLock } from "../src/contract.js";

const packageRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const releaseRoot = path.join(packageRoot, "release");

await buildCatalog({ check: true });
const packRoot = fs.mkdtempSync(path.join(os.tmpdir(), "uui-skills-pack-"));
try {
  const [packed] = JSON.parse(
    runNpm(
      ["pack", "--ignore-scripts", "--json", "--pack-destination", packRoot],
      packageRoot
    )
  );
  const packedPaths = new Set(packed.files.map((entry) => entry.path));
  for (const requiredPath of [
    "package.json",
    "src/index.js",
    "src/index.d.ts",
    "src/manifest.js",
    "src/catalog.generated.js",
    "src/SkillsRoute.js",
    "src/skills-route.css"
  ]) {
    if (!packedPaths.has(requiredPath)) {
      throw new Error(`Packed artifact is missing ${requiredPath}.`);
    }
  }
  if ([...packedPaths].some((entry) => entry.startsWith("test/") || entry.startsWith("scripts/"))) {
    throw new Error("Packed artifact must not expose tests or maintainer scripts.");
  }

  fs.mkdirSync(releaseRoot, { recursive: true });
  const sourceArtifact = path.join(packRoot, packed.filename);
  const targetArtifact = path.join(releaseRoot, packed.filename);
  fs.copyFileSync(sourceArtifact, targetArtifact);
  const digest = createHash("sha256")
    .update(fs.readFileSync(targetArtifact))
    .digest("hex");
  const releaseLock = {
    schemaVersion: RELEASE_LOCK_SCHEMA_VERSION,
    contributions: [
      {
        contribution: contributionDefinition,
        artifact: {
          filename: packed.filename,
          sha256: digest
        }
      }
    ]
  };
  validateReleaseLock(releaseLock, digest);
  fs.writeFileSync(
    path.join(releaseRoot, "universal-ui.release-lock.json"),
    `${JSON.stringify(releaseLock, null, 2)}\n`
  );
  process.stdout.write(
    `${JSON.stringify({
      artifact: path.relative(packageRoot, targetArtifact).replaceAll("\\", "/"),
      sha256: digest,
      releaseLock: "release/universal-ui.release-lock.json"
    })}\n`
  );
} finally {
  fs.rmSync(packRoot, { recursive: true, force: true });
}
