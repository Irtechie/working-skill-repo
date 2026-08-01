import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import { runNpm } from "../scripts/npm-cli.mjs";

const packageRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const verifierPath = path.join(packageRoot, "scripts", "verify-contract-tarball.mjs");

test("offline verifier checks immutable contract identity and public validation", (t) => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "uui-contract-fixture-"));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const source = path.join(root, "source");
  const packed = path.join(root, "packed");
  fs.mkdirSync(source);
  fs.mkdirSync(packed);
  fs.writeFileSync(
    path.join(source, "package.json"),
    JSON.stringify({
      name: "@irtechie/universal-ui-contract",
      version: "0.1.0",
      type: "module",
      main: "./index.js",
      exports: "./index.js"
    })
  );
  fs.writeFileSync(
    path.join(source, "index.js"),
    `export const CONTRIBUTION_SCHEMA_VERSION = "universal_ui.contribution.v1";
export const HOST_CONTRACT_VERSION = "0.1.0";
export function validateContribution(value) {
  if (value?.schemaVersion !== CONTRIBUTION_SCHEMA_VERSION) {
    throw new TypeError("Invalid contribution.");
  }
  return value;
}
`
  );
  const [packResult] = JSON.parse(
    runNpm(
      ["pack", source, "--ignore-scripts", "--json", "--pack-destination", packed],
      root
    )
  );
  const tarball = path.join(packed, packResult.filename);
  const digest = createHash("sha256")
    .update(fs.readFileSync(tarball))
    .digest("hex");
  const output = execFileSync(process.execPath, [verifierPath], {
    cwd: packageRoot,
    encoding: "utf8",
    env: {
      ...process.env,
      UNIVERSAL_UI_CONTRACT_TGZ: tarball,
      UNIVERSAL_UI_CONTRACT_SHA256: digest
    }
  });
  assert.match(
    output,
    /verified @irtechie\/universal-ui-contract@0\.1\.0 sha256=[a-f0-9]{64}/
  );

  assert.throws(
    () =>
      execFileSync(process.execPath, [verifierPath], {
        cwd: packageRoot,
        encoding: "utf8",
        env: {
          ...process.env,
          UNIVERSAL_UI_CONTRACT_TGZ: tarball,
          UNIVERSAL_UI_CONTRACT_SHA256: "0".repeat(64)
        },
        stdio: "pipe"
      }),
    /Command failed/
  );
});
