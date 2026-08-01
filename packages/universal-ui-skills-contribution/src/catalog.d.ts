export interface SkillCatalogEntry {
  id: string;
  name: string;
  description: string;
  argumentHint: string | null;
  category: string;
  sourcePath: string;
  sourceUrl: string;
}

export declare const skillCatalog: readonly Readonly<SkillCatalogEntry>[];
export declare const skillCatalogMetadata: Readonly<{
  schemaVersion: "working_skill_repo.catalog.v1";
  owner: "Irtechie/working-skill-repo";
  sourceRoot: ".github/skills";
  count: number;
}>;
