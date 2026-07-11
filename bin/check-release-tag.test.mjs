import assert from "node:assert/strict";
import { test } from "node:test";

import { assertReleaseTag } from "./check-release-tag.mjs";

test("accepts only the exact v-prefixed package version", () => {
  assert.equal(assertReleaseTag("v1.2.3", "1.2.3"), "v1.2.3");
  assert.equal(assertReleaseTag("v1.2.3-rc.1+build.7", "1.2.3-rc.1+build.7"), "v1.2.3-rc.1+build.7");
});

test("rejects a different, unprefixed, or decorated tag", () => {
  assert.throws(() => assertReleaseTag("v1.2.4", "1.2.3"), /must be exactly v1\.2\.3/);
  assert.throws(() => assertReleaseTag("1.2.3", "1.2.3"), /must be exactly v1\.2\.3/);
  assert.throws(() => assertReleaseTag("release-v1.2.3", "1.2.3"), /must be exactly v1\.2\.3/);
});

test("rejects non-SemVer package versions", () => {
  for (const version of ["1.2", "01.2.3", "1.02.3", "v1.2.3", "1.2.3-"]) {
    assert.throws(() => assertReleaseTag(`v${version}`, version), /not exact SemVer/);
  }
});
