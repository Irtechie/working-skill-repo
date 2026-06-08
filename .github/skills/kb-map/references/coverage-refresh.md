# Coverage Refresh Reference

Use this during `kb-map refresh` when lookup reveals project-memory coverage
gaps, shallow maps, or missing subsystem pointers.

## Coverage Gap Rule

Coverage is insufficient when a fresh session must rediscover basics by broad
search. Common signals:

- the relevant subsystem doc is missing or too generic;
- a broad parent doc exists but does not point to the child workflow;
- source-of-truth files, scripts, CI workflows, generated artifacts, or release
  assets are not named;
- current mode is unclear, such as bundled runtime vs download-on-demand;
- known failure modes or "do not assume" notes are missing;
- the user says the session has no clue about the subsystem after `kb-map`.

When coverage is insufficient:

1. Stop normal routing long enough to run a targeted refresh for that subsystem.
2. Search/read only the files needed to understand the missing workflow.
3. Update `docs/context/PROJECT.md` and/or the smallest relevant
   `docs/context/architecture/<subsystem>.md` child doc.
4. Add a `docs/context/memory-maintenance.md` signal:
   `stale-doc` or `repeated-rediscovery`, with the missing subsystem and source
   paths.
5. Re-run `kb-map lookup <same request>` and report the exact docs a fresh
   session should read next.

## Coverage Audit Mode

Use this when the user says the initial map missed a whole class of things, or
when one invisible subsystem suggests more may be missing.

1. Inventory major surfaces:
   - routes/screens/UI shells;
   - commands/CLIs/tools/actions/workflows/service capabilities;
   - jobs/workflows/integrations;
   - data/auth/session/storage;
   - build/test/package/installer/updater/release/deploy flows.
2. Compare the inventory to:
   - `docs/context/PROJECT.md`;
   - `docs/context/architecture/README.md`;
   - `docs/context/architecture/*.md`;
   - `docs/context/operations/*.md`;
   - `docs/context/research/*.md` and `docs/context/decisions/*.md` when
     relevant.
3. Produce a gap table:

   ```markdown
   | Area | Evidence Files | Indexed? | Doc Exists? | Good Enough? | Gap | Priority |
   |---|---|---:|---:|---:|---|---|
   ```

4. Fix P0/P1 coverage gaps immediately when they block fresh-session routing.
   Record lower-priority gaps in `docs/context/memory-maintenance.md`.

This is not a full re-bootstrap unless `PROJECT.md` or the architecture index
is so wrong that targeted repair would be more expensive than rebuilding.

## Shallow Map Detection

During lookup, warn and recommend `refresh` or `kb-map-bootstrap` when the KB
appears much thinner than the repo:

- architecture docs are generic top-level names while substantial child
  directories, route/page folders, commands, tools, actions, or workflows are
  missing from the subsystem index;
- a major source area has a rough source-file to architecture-doc ratio above
  about 25:1 without child pointers or explicit skip reasons;
- root `README.md`, `AGENTS.md`, `.github/copilot-instructions.md`, memories,
  or project notes mention architecture topics absent from `docs/context`;
- file prefix clusters, large files, large directories, or cross-process flows
  are visible but undocumented;
- must-cover concerns such as auth, storage, IPC, browser/HTTP, telemetry,
  settings, build/install, LLM/model, workers, or integrations exist but have no
  doc, parent pointer, or known-unknown entry.

If the requested work touches one of these gaps, stop normal routing and run a
targeted refresh. If many gaps exist, recommend re-bootstrap with coverage
discovery instead of one-off patching.
