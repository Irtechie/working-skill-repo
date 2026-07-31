export const CONTRIBUTION_SCHEMA_VERSION = "universal_ui.contribution.v1";
export const RELEASE_LOCK_SCHEMA_VERSION = "universal_ui.release-lock.v1";
export const SHELL_CONTEXT_SCHEMA_VERSION = "universal_ui.shell-context.v1";
export const HOST_CONTRACT_VERSION = "0.1.0";

function requireRecord(value, label) {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new TypeError(`${label} must be an object.`);
  }
  return value;
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new TypeError(`${label} must be a non-empty string.`);
  }
  return value;
}

function requireLiteral(value, allowed, label) {
  if (!allowed.includes(value)) {
    throw new TypeError(`${label} must be one of: ${allowed.join(", ")}.`);
  }
  return value;
}

function requireInteger(value, label) {
  if (!Number.isSafeInteger(value)) {
    throw new TypeError(`${label} must be an integer.`);
  }
  return value;
}

function requireSafeLocalRoute(value, label, requiredPrefix) {
  const route = requireString(value, label);
  if (!route.startsWith(requiredPrefix) || route.startsWith("//") || route.includes("\\")) {
    throw new TypeError(`${label} must be an absolute same-origin path.`);
  }

  let decodedPath = route.split(/[?#]/, 1)[0];
  for (;;) {
    let next;
    try {
      next = decodeURIComponent(decodedPath);
    } catch {
      throw new TypeError(`${label} contains malformed percent encoding.`);
    }
    if (next === decodedPath) {
      break;
    }
    decodedPath = next;
  }
  if (
    decodedPath.startsWith("//") ||
    decodedPath.includes("\\") ||
    decodedPath.split("/").includes("..")
  ) {
    throw new TypeError(`${label} must be traversal-free.`);
  }

  const sentinel = new URL("https://local-route.invalid/");
  const resolved = new URL(route, sentinel);
  if (resolved.origin !== sentinel.origin || !resolved.pathname.startsWith(requiredPrefix)) {
    throw new TypeError(`${label} must remain on the expected local path.`);
  }
  return route;
}

export function validateContribution(definition) {
  const value = requireRecord(definition, "Contribution definition");
  if (value.schemaVersion !== CONTRIBUTION_SCHEMA_VERSION) {
    throw new TypeError("Unsupported contribution schema version.");
  }

  const contribution = requireRecord(value.contribution, "Contribution");
  requireString(contribution.id, "Contribution id");
  requireString(contribution.version, "Contribution version");
  if (contribution.supportedHostRange !== `^${HOST_CONTRACT_VERSION}`) {
    throw new TypeError("Unsupported host range.");
  }

  const owner = requireRecord(contribution.owner, "Contribution owner");
  requireString(owner.id, "Contribution owner id");
  if (!/^https:\/\/github\.com\//.test(requireString(owner.repository, "Owner repository"))) {
    throw new TypeError("Contribution owner repository must be a GitHub HTTPS URL.");
  }

  const dependencies = requireRecord(contribution.sharedDependencies, "Shared dependencies");
  const react = requireRecord(dependencies.react, "React shared dependency");
  requireString(react.version, "React version");
  if (react.owner !== "host") {
    throw new TypeError("React must be host-owned.");
  }
  if (!Array.isArray(contribution.browserApiDependencies)) {
    throw new TypeError("Browser API dependencies must be an array.");
  }
  const browserDependencyIds = new Set();
  for (const dependencyValue of contribution.browserApiDependencies) {
    const dependency = requireRecord(dependencyValue, "Browser API dependency");
    const id = requireString(dependency.id, "Browser API dependency id");
    requireString(dependency.version, "Browser API dependency version");
    if (browserDependencyIds.has(id)) {
      throw new TypeError("Duplicate browser API dependency id.");
    }
    browserDependencyIds.add(id);
  }

  const contributionHealth = requireRecord(contribution.health, "Contribution health");
  requireLiteral(contributionHealth.kind, ["static"], "Contribution health kind");
  requireLiteral(
    contributionHealth.status,
    ["healthy", "unknown", "unhealthy"],
    "Contribution health status"
  );

  if (!Array.isArray(value.routes) || value.routes.length === 0) {
    throw new TypeError("At least one route is required.");
  }
  const ids = new Set();
  const paths = new Set();
  for (const routeValue of value.routes) {
    const route = requireRecord(routeValue, "Route");
    const id = requireString(route.id, "Route id");
    const routePath = requireSafeLocalRoute(route.path, "Route path", "/apps/");
    if (
      routePath.includes("//") ||
      /[\s?#]/.test(routePath)
    ) {
      throw new TypeError("Route path is unsafe.");
    }
    if (ids.has(id) || paths.has(routePath)) {
      throw new TypeError("Duplicate route id or path.");
    }
    ids.add(id);
    paths.add(routePath);
    requireSafeLocalRoute(route.legacyRoute, "Legacy route", "/");
    requireString(route.displayName, "Display name");
    requireString(route.purpose, "Purpose");
    requireString(route.directorySummary, "Directory summary");
    requireString(route.domainGroup, "Domain group");
    requireInteger(route.order, "Route order");
    const visibility = requireRecord(route.visibility, "Route visibility");
    requireLiteral(visibility.audience, ["all"], "Route visibility audience");
    const health = requireRecord(route.health, "Route health");
    requireLiteral(
      health.projection,
      ["healthy", "unknown", "unhealthy"],
      "Route health projection"
    );
    requireLiteral(route.canvasMode, ["standard", "immersive"], "Route canvas mode");
    if (route.lifecycle !== undefined) {
      const lifecycle = requireRecord(route.lifecycle, "Route lifecycle");
      requireLiteral(lifecycle.kind, ["route"], "Route lifecycle kind");
      if (typeof lifecycle.disposeOnExit !== "boolean") {
        throw new TypeError("Route lifecycle disposeOnExit must be a boolean.");
      }
    }
  }

  return definition;
}

export function validateReleaseLock(releaseLock, artifactDigest) {
  const value = requireRecord(releaseLock, "Release lock");
  if (value.schemaVersion !== RELEASE_LOCK_SCHEMA_VERSION) {
    throw new TypeError("Unsupported release-lock schema version.");
  }
  if (!/^[a-f0-9]{64}$/i.test(artifactDigest)) {
    throw new TypeError("Artifact digest must be SHA-256.");
  }
  if (!Array.isArray(value.contributions) || value.contributions.length !== 1) {
    throw new TypeError("Release lock must contain this one contribution.");
  }
  const entry = requireRecord(value.contributions[0], "Release entry");
  validateContribution(entry.contribution);
  const artifact = requireRecord(entry.artifact, "Release artifact");
  const artifactFilename = requireString(artifact.filename, "Artifact filename");
  if (artifactFilename !== pathBasename(artifactFilename) || !artifactFilename.endsWith(".tgz")) {
    throw new TypeError("Artifact filename must be a local .tgz basename.");
  }
  if (!/^[a-f0-9]{64}$/i.test(artifact.sha256) || artifact.sha256 !== artifactDigest) {
    throw new TypeError("Artifact digest does not match.");
  }
  return releaseLock;
}

function pathBasename(value) {
  return value.split(/[\\/]/).at(-1);
}
