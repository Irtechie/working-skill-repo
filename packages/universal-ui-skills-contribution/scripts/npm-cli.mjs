import { execFileSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";

export function runNpm(args, cwd) {
  const colocatedCli = path.join(
    path.dirname(process.execPath),
    "node_modules",
    "npm",
    "bin",
    "npm-cli.js"
  );
  const options = {
    cwd,
    encoding: "utf8",
    maxBuffer: 10 * 1024 * 1024,
    timeout: 120_000
  };
  if (fs.existsSync(colocatedCli)) {
    return execFileSync(process.execPath, [colocatedCli, ...args], options);
  }
  return execFileSync(process.platform === "win32" ? "npm.cmd" : "npm", args, {
    ...options,
    shell: process.platform === "win32"
  });
}
