import assert from "node:assert/strict";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import {
  buildWiki,
  collectLearning,
  parseInstincts,
  renderWiki,
} from "./kb-learning-wiki.mjs";

test("parses scoped instincts and evidence", () => {
  const instincts = parseInstincts(`instincts:
  - id: verify-before-promoting
    scope: model-routing/evaluation
    trigger: "when promotion is considered"
    behavior: "require protected proof"
    confidence: 0.6
    domain: testing
    observations: 2
    first_seen: 2026-07-01
    last_seen: 2026-07-22
    evidence:
      - "test evidence"
`, "docs/context/kb/instincts/scoped/model-routing/evaluation.yaml");

  assert.deepEqual(instincts, [{
    id: "verify-before-promoting",
    scope: "model-routing/evaluation",
    trigger: "when promotion is considered",
    behavior: "require protected proof",
    confidence: 0.6,
    domain: "testing",
    observations: 2,
    first_seen: "2026-07-01",
    last_seen: "2026-07-22",
    evidence: ["test evidence"],
    source: "docs/context/kb/instincts/scoped/model-routing/evaluation.yaml",
  }]);
});

test("renders escaped, searchable, source-linked HTML", () => {
  const html = renderWiki({
    generatedAt: "2026-07-22T00:00:00.000Z",
    instincts: [{
      id: "<unsafe>",
      scope: "model-routing",
      trigger: "when testing",
      behavior: "escape <script>",
      confidence: 0.5,
      observations: 1,
      last_seen: "2026-07-22",
      reviewState: "current",
      evidence: ["proof"],
      source: "docs/context/kb/instincts/scoped/model-routing.yaml",
    }],
    solutions: [],
    goals: [],
    research: [],
    landmines: { active: 0, resolved: 0, reviewed: "" },
  });

  assert.match(html, /KB Learning Wiki/);
  assert.match(html, /&lt;unsafe&gt;/);
  assert.doesNotMatch(html, /<h3><unsafe><\/h3>/);
  assert.match(html, /href="\.\.\/context\/kb\/instincts\/scoped\/model-routing\.yaml"/);
  assert.match(html, /id="search"/);
  assert.doesNotMatch(html, /<script src=/);
  assert.doesNotMatch(html, /fetch\(/);
});

test("collects canonical sources and builds the wiki", async (t) => {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "kb-learning-wiki-"));
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  await fs.mkdir(path.join(root, "docs/context/kb/instincts/scoped"), { recursive: true });
  await fs.mkdir(path.join(root, "docs/solutions/workflow"), { recursive: true });
  await fs.mkdir(path.join(root, "docs/context/goals"), { recursive: true });
  await fs.mkdir(path.join(root, "docs/context/research"), { recursive: true });
  await fs.writeFile(path.join(root, "docs/context/kb/instincts/scoped/demo.yaml"), `instincts:
  - id: demo
    scope: demo
    trigger: "when demoing"
    behavior: "show proof"
    confidence: 0.5
    domain: workflow
    observations: 1
    first_seen: 2026-07-01
    last_seen: 2026-07-22
    evidence:
      - "fixture"
`);
  await fs.writeFile(path.join(root, "docs/solutions/workflow/example.md"), `---
title: Example Solution
date: 2026-07-20
---
# Example Solution

Reusable guidance lives here.
`);
  await fs.writeFile(path.join(root, "docs/context/goals/demo.md"), "# Demo Goal\n\nStatus: active\n\nReach the target.\n");
  await fs.writeFile(path.join(root, "docs/context/research/2026-07-21-demo.md"), "# Demo Research\n\nMeasured evidence.\n");
  await fs.writeFile(path.join(root, "docs/context/landmines.md"), "# Landmines\n\nLast reviewed: 2026-07-22\n\n## Active Landmines\n\nNone.\n\n## Resolved Landmines\n\nNone.\n");

  const data = await collectLearning(root, new Date("2026-07-22T00:00:00Z"));
  assert.equal(data.instincts.length, 1);
  assert.equal(data.solutions.length, 1);
  assert.equal(data.goals.length, 1);
  assert.equal(data.research.length, 1);
  assert.equal(data.instincts[0].reviewState, "current");

  const output = path.join(root, "site", "index.html");
  await buildWiki(root, output);
  const html = await fs.readFile(output, "utf8");
  assert.match(html, /Example Solution/);
  assert.match(html, /Demo Research/);
});
