import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { pathToFileURL } from "node:url";

import {
  contributionDefinition,
  routeLoaders,
  validateContribution,
  validateReleaseLock
} from "../src/index.js";

test("exports the canonical single-route UniversalUI contribution", () => {
  assert.equal(validateContribution(contributionDefinition), contributionDefinition);
  assert.equal(contributionDefinition.schemaVersion, "universal_ui.contribution.v1");
  assert.deepEqual(contributionDefinition.contribution.owner, {
    id: "irtechie",
    repository: "https://github.com/Irtechie/working-skill-repo"
  });
  assert.equal(contributionDefinition.contribution.sharedDependencies.react.owner, "host");
  assert.deepEqual(contributionDefinition.contribution.browserApiDependencies, []);
  assert.deepEqual(contributionDefinition.routes.map(({ id, path, legacyRoute }) => ({
    id,
    path,
    legacyRoute
  })), [{
    id: "skills",
    path: "/apps/skills",
    legacyRoute: "/?tab=skills"
  }]);
  assert.deepEqual(Object.keys(routeLoaders), ["skills"]);
});

test("rejects unsafe routes and contribution-owned React", () => {
  const unsafeRoute = structuredClone(contributionDefinition);
  unsafeRoute.routes[0].path = "/apps/../private";
  assert.throws(() => validateContribution(unsafeRoute), /unsafe/i);

  const contributionOwnedReact = structuredClone(contributionDefinition);
  contributionOwnedReact.contribution.sharedDependencies.react.owner = "contribution";
  assert.throws(() => validateContribution(contributionOwnedReact), /host-owned/i);
});

test("release metadata binds the contribution to one SHA-256 artifact", () => {
  const digest = "a".repeat(64);
  const releaseLock = {
    schemaVersion: "universal_ui.release-lock.v1",
    contributions: [{
      contribution: contributionDefinition,
      artifact: {
        filename: "irtechie-universal-ui-skills-contribution-0.1.0.tgz",
        sha256: digest
      }
    }]
  };
  assert.equal(validateReleaseLock(releaseLock, digest), releaseLock);
  assert.throws(() => validateReleaseLock(releaseLock, "b".repeat(64)), /digest/i);
});

test("committed release lock matches the packed artifact SHA-256", async () => {
  const releaseRoot = path.resolve("release");
  const releaseLock = JSON.parse(
    await fs.readFile(path.join(releaseRoot, "universal-ui.release-lock.json"), "utf8")
  );
  const artifact = releaseLock.contributions[0].artifact;
  const digest = createHash("sha256")
    .update(await fs.readFile(path.join(releaseRoot, artifact.filename)))
    .digest("hex");
  assert.equal(digest, artifact.sha256);
  assert.equal(validateReleaseLock(releaseLock, digest), releaseLock);
});

test("lazy route default accepts ShellContextV1 without owning React or the shell", async (t) => {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "uui-skills-route-"));
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  await fs.cp(path.resolve("src"), path.join(root, "src"), { recursive: true });
  await fs.mkdir(path.join(root, "node_modules", "react"), { recursive: true });
  await fs.writeFile(
    path.join(root, "package.json"),
    JSON.stringify({ type: "module" })
  );
  await fs.writeFile(
    path.join(root, "node_modules", "react", "package.json"),
    JSON.stringify({
      name: "react",
      version: "19.0.0",
      type: "module",
      exports: "./index.js"
    })
  );
  await fs.writeFile(
    path.join(root, "node_modules", "react", "index.js"),
    `export const createElement = (type, props, ...children) => ({ type, props: { ...props, children } });
export const useMemo = (factory) => factory();
export const useState = (initial) => [initial, () => {}];
`
  );

  const contribution = await import(
    `${pathToFileURL(path.join(root, "src", "index.js")).href}?test=${Date.now()}`
  );
  const routeModule = await contribution.routeLoaders.skills();
  const tree = routeModule.default({
    context: {
      schemaVersion: "universal_ui.shell-context.v1",
      activeRouteId: "skills",
      navigation: "available",
      visibility: "all",
      session: { summary: "synthetic" }
    }
  });
  assert.equal(tree.type, "section");
  assert.equal(tree.props["data-testid"], "skills-contribution");
});

test("runtime source contains no shell, router, root mount, or raw HTML owner", async () => {
  const source = (
    await Promise.all(
      ["src/index.js", "src/SkillsRoute.js", "src/manifest.js"].map((file) =>
        fs.readFile(file, "utf8")
      )
    )
  ).join("\n");
  assert.doesNotMatch(
    source,
    /createRoot|react-dom|BrowserRouter|RouterProvider|UniversalFrame|shell-frame|dangerouslySetInnerHTML/
  );
});
