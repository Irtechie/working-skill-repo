import path from "node:path";
import { fileURLToPath } from "node:url";

const fixtureRoot = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(fixtureRoot, "..", "..", "..", "..");
const universalUiRoot = process.env.UNIVERSAL_UI_ROOT;

if (!universalUiRoot) {
  throw new Error("UNIVERSAL_UI_ROOT must point to a local UniversalUI checkout.");
}

const universalUiModules = path.join(universalUiRoot, "node_modules");

export default {
  resolve: {
    alias: [
      {
        find: /^react$/,
        replacement: path.join(universalUiModules, "react", "index.js")
      },
      {
        find: /^react-dom\/client$/,
        replacement: path.join(universalUiModules, "react-dom", "client.js")
      }
    ]
  },
  server: {
    fs: {
      allow: [repoRoot, universalUiModules]
    }
  }
};
