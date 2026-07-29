export { skillCatalog, skillCatalogMetadata } from "./catalog.generated.js";
export {
  CONTRIBUTION_SCHEMA_VERSION,
  HOST_CONTRACT_VERSION,
  RELEASE_LOCK_SCHEMA_VERSION,
  SHELL_CONTEXT_SCHEMA_VERSION,
  validateContribution,
  validateReleaseLock
} from "./contract.js";
export { contributionDefinition, expectedRouteIds } from "./manifest.js";

export const routeLoaders = Object.freeze({
  skills: () => import("./SkillsRoute.js")
});
