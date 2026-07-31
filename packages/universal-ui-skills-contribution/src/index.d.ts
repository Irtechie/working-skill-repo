import type { ComponentType } from "react";
import type { ShellContextV1 } from "./contract.js";

export * from "./catalog.js";
export * from "./contract.js";
export * from "./manifest.js";

export type SkillsRouteProps = {
  context: ShellContextV1;
};

export type SkillsRouteLoader = () => Promise<{
  default: ComponentType<SkillsRouteProps>;
}>;

export declare const routeLoaders: Readonly<{
  skills: SkillsRouteLoader;
}>;
