export declare const CONTRIBUTION_SCHEMA_VERSION: "universal_ui.contribution.v1";
export declare const RELEASE_LOCK_SCHEMA_VERSION: "universal_ui.release-lock.v1";
export declare const SHELL_CONTEXT_SCHEMA_VERSION: "universal_ui.shell-context.v1";
export declare const HOST_CONTRACT_VERSION: "0.1.0";

export interface ShellContextV1 {
  schemaVersion: typeof SHELL_CONTEXT_SCHEMA_VERSION;
  activeRouteId: string;
  navigation: "available";
  visibility: "all" | "demo" | "personal" | "work";
  session: {
    summary: "synthetic";
  };
}

export interface RouteDescriptorV1 {
  id: string;
  path: string;
  legacyRoute: string;
  displayName: string;
  purpose: string;
  directorySummary: string;
  domainGroup: string;
  order: number;
  visibility: { audience: "all" };
  health: { projection: "healthy" | "unknown" | "unhealthy" };
  canvasMode: "standard" | "immersive";
  lifecycle?: {
    kind: "route";
    disposeOnExit: boolean;
  };
}

export interface ContributionDefinitionV1 {
  schemaVersion: typeof CONTRIBUTION_SCHEMA_VERSION;
  contribution: {
    id: string;
    owner: {
      id: string;
      repository: string;
    };
    version: string;
    supportedHostRange: string;
    sharedDependencies: {
      react: {
        version: string;
        owner: "host";
      };
    };
    browserApiDependencies: Array<{
      id: string;
      version: string;
    }>;
    health: {
      kind: "static";
      status: "healthy" | "unknown" | "unhealthy";
    };
  };
  routes: RouteDescriptorV1[];
}

export interface ReleaseLockV1 {
  schemaVersion: typeof RELEASE_LOCK_SCHEMA_VERSION;
  contributions: Array<{
    contribution: ContributionDefinitionV1;
    artifact: {
      filename: string;
      sha256: string;
    };
  }>;
}

export declare function validateContribution(
  definition: unknown
): ContributionDefinitionV1;
export declare function validateReleaseLock(
  releaseLock: unknown,
  artifactDigest: string
): ReleaseLockV1;
