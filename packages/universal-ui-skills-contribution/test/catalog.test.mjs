import assert from "node:assert/strict";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import { skillCatalog } from "../src/catalog.generated.js";
import {
  buildCatalogModule,
  collectSkillCatalog
} from "../scripts/build-catalog.mjs";

const packageRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const repoRoot = path.resolve(packageRoot, "..", "..");
const machinePrivatePathPattern =
  /\b[A-Za-z]:[\\/]{1,2}(?:Users|Documents and Settings)[\\/]{1,2}|\/(?:home|Users)\/|%USERPROFILE%|\$HOME|file\s*:/i;

test("generated catalog exactly matches canonical SKILL.md frontmatter", async () => {
  const current = await collectSkillCatalog(repoRoot);
  assert.deepEqual(skillCatalog, current);
  const generated = await fs.readFile(
    path.join(packageRoot, "src", "catalog.generated.js"),
    "utf8"
  );
  assert.equal(generated, buildCatalogModule(current));
});

test("catalog projection excludes executable markup and machine-private paths", () => {
  const serialized = JSON.stringify(skillCatalog);
  assert.doesNotMatch(
    serialized,
    /<\s*\/?\s*(?:script|iframe|object|embed|svg|img|link|style)\b|javascript\s*:|data\s*:\s*text\/html/i
  );
  assert.doesNotMatch(
    serialized,
    machinePrivatePathPattern
  );
  assert.ok(skillCatalog.every((skill) => skill.sourcePath.startsWith(".github/skills/")));
});

test("machine-private path matcher covers serialized Windows paths", () => {
  const windowsPath = ["C:", "Users", "private-user", "secret"].join("\\");
  assert.match(windowsPath, machinePrivatePathPattern);
  assert.match(JSON.stringify(windowsPath), machinePrivatePathPattern);
});

test("Markdown bodies never cross the catalog boundary", async (t) => {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "uui-skills-catalog-"));
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  const skillRoot = path.join(root, ".github", "skills", "safe-skill");
  await fs.mkdir(skillRoot, { recursive: true });
  await fs.writeFile(
    path.join(skillRoot, "SKILL.md"),
    `---
name: safe-skill
description: Safe projected description.
argument-hint: "[request]"
---
# Safe Skill

PRIVATE-BODY-MARKER local-private-value
<script>alert("not projected")</script>
[unsafe](javascript:alert(1))
`
  );

  const projected = await collectSkillCatalog(root);
  assert.equal(projected.length, 1);
  assert.doesNotMatch(
    JSON.stringify(projected),
    /PRIVATE-BODY-MARKER|<script|local-private-value/
  );

  await fs.writeFile(
    path.join(skillRoot, "SKILL.md"),
    `---
name: safe-skill
description: <script>alert("frontmatter")</script>
---
`
  );
  await assert.rejects(() => collectSkillCatalog(root), /executable markup/i);
});
