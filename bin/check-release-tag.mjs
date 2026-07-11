#!/usr/bin/env node

import { readFile } from "node:fs/promises";
import { pathToFileURL } from "node:url";

const SEMVER = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9]\d*|\d*[A-Za-z-][0-9A-Za-z-]*))*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$/;

export function assertReleaseTag(tag, version) {
  if (typeof version !== "string" || !SEMVER.test(version)) {
    throw new Error(`package.json version is not exact SemVer: ${JSON.stringify(version)}`);
  }
  const expected = `v${version}`;
  if (tag !== expected) {
    throw new Error(`release tag must be exactly ${expected}; received ${JSON.stringify(tag)}`);
  }
  return expected;
}

function parseArgs(argv) {
  const options = { packagePath: "package.json" };
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === "--tag") {
      options.tag = argv[++index];
    } else if (argument === "--package") {
      options.packagePath = argv[++index];
    } else {
      throw new Error(`unknown argument: ${argument}`);
    }
  }
  if (!options.tag) {
    throw new Error("--tag is required");
  }
  return options;
}

export async function main(argv = process.argv.slice(2)) {
  const options = parseArgs(argv);
  const packageDocument = JSON.parse(await readFile(options.packagePath, "utf8"));
  const expected = assertReleaseTag(options.tag, packageDocument.version);
  process.stdout.write(`release-tag: valid ${expected}\n`);
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  main().catch((error) => {
    process.stderr.write(`release-tag: ${error.message}\n`);
    process.exitCode = 1;
  });
}
