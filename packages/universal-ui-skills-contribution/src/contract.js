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

  if (!Array.isArray(value.routes) || value.routes.length === 0) {
    throw new TypeError("At least one route is required.");
  }
  const ids = new Set();
  const paths = new Set();
  for (const routeValue of value.routes) {
    const route = requireRecord(routeValue, "Route");
    const id = requireString(route.id, "Route id");
    const routePath = requireString(route.path, "Route path");
    if (
      !routePath.startsWith("/apps/") ||
      routePath.includes("//") ||
      routePath.includes("..") ||
      /[\s?#]/.test(routePath)
    ) {
      throw new TypeError("Route path is unsafe.");
    }
    if (ids.has(id) || paths.has(routePath)) {
      throw new TypeError("Duplicate route id or path.");
    }
    ids.add(id);
    paths.add(routePath);
    requireString(route.legacyRoute, "Legacy route");
    requireString(route.displayName, "Display name");
    requireString(route.purpose, "Purpose");
    requireString(route.directorySummary, "Directory summary");
    requireString(route.domainGroup, "Domain group");
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
  requireString(artifact.filename, "Artifact filename");
  if (artifact.sha256 !== artifactDigest) {
    throw new TypeError("Artifact digest does not match.");
  }
  return releaseLock;
}
