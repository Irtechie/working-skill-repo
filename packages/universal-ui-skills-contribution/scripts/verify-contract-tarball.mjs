#!/usr/bin/env node

import { createHash } from "node:crypto";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { pathToFileURL } from "node:url";

import { contributionDefinition } from "../src/manifest.js";
import { runNpm } from "./npm-cli.mjs";

const EXPECTED_NAME = "@irtechie/universal-ui-contract";
const EXPECTED_VERSION = "0.1.0";

const tarball = process.env.UNIVERSAL_UI_CONTRACT_TGZ;
const expectedDigest = process.env.UNIVERSAL_UI_CONTRACT_SHA256;

if (!tarball && !expectedDigest) {
  process.stdout.write("universal-ui-contract: skipped (immutable tarball not configured)\n");
  process.exit(0);
}
if (!tarball || !expectedDigest) {
  throw new Error(
    "UNIVERSAL_UI_CONTRACT_TGZ and UNIVERSAL_UI_CONTRACT_SHA256 must be set together."
  );
}
if (!/^[a-f0-9]{64}$/i.test(expectedDigest)) {
  throw new Error("UNIVERSAL_UI_CONTRACT_SHA256 must be a SHA-256 digest.");
}

const tarballPath = path.resolve(tarball);
const actualDigest = createHash("sha256")
  .update(fs.readFileSync(tarballPath))
  .digest("hex");
if (actualDigest !== expectedDigest.toLowerCase()) {
  throw new Error("UniversalUI contract tarball digest mismatch.");
}

const consumerRoot = fs.mkdtempSync(path.join(os.tmpdir(), "uui-skills-contract-"));
try {
  fs.writeFileSync(
    path.join(consumerRoot, "package.json"),
    JSON.stringify({
      name: "uui-skills-contract-conformance",
      version: "0.0.0",
      private: true,
      type: "module",
      dependencies: {
        [EXPECTED_NAME]: `file:${tarballPath.replaceAll("\\", "/")}`
      }
    })
  );
  runNpm(
    [
      "install",
      "--ignore-scripts",
      "--offline",
      "--no-audit",
      "--no-fund",
      "--package-lock=false"
    ],
    consumerRoot
  );

  const installedRoot = path.join(
    consumerRoot,
    "node_modules",
    ...EXPECTED_NAME.split("/")
  );
  const installedPackage = JSON.parse(
    fs.readFileSync(path.join(installedRoot, "package.json"), "utf8")
  );
  if (installedPackage.name !== EXPECTED_NAME || installedPackage.version !== EXPECTED_VERSION) {
    throw new Error(
      `UniversalUI contract identity mismatch: expected ${EXPECTED_NAME}@${EXPECTED_VERSION}.`
    );
  }

  const entry = path.join(consumerRoot, "entry.mjs");
  fs.writeFileSync(entry, `export * as contract from ${JSON.stringify(EXPECTED_NAME)};`);
  const { contract } = await import(pathToFileURL(entry).href);
  if (
    contract.CONTRIBUTION_SCHEMA_VERSION !== "universal_ui.contribution.v1" ||
    contract.HOST_CONTRACT_VERSION !== EXPECTED_VERSION
  ) {
    throw new Error("UniversalUI contract public version exports do not match 0.1.0.");
  }
  contract.validateContribution(contributionDefinition);
  process.stdout.write(
    `universal-ui-contract: verified ${EXPECTED_NAME}@${EXPECTED_VERSION} sha256=${actualDigest}\n`
  );
} finally {
  fs.rmSync(consumerRoot, { recursive: true, force: true });
}
