# Reconcile OpenCode with a noisy LiteLLM plugin

Use this when interactive OpenCode works, but `opencode run --format json`
fails in a consumer because stdout contains `[opencode-litellm]` log messages
alongside JSON events. This is a plugin output problem, not evidence that the
LiteLLM endpoint or model is broken.

## Diagnose the installed copy

Run `opencode --version` and `opencode debug paths`. Find the configured
`opencode-plugin-litellm` entry in your OpenCode configuration. Preserve the
rest of that configuration, including provider settings and secret references.
Do not paste credentials into a bug report.

For version 0.10.0, inspect the installed package's `src/plugin/index.ts`.
Its three `console.log` calls print model discovery/cache status to stdout.
These messages break consumers expecting JSONL, one JSON object per line.
Existing `console.warn` messages already use stderr.

The package may be under the cache reported by `opencode debug paths`, for
example `packages/opencode-plugin-litellm@0.10.0/node_modules/opencode-plugin-litellm`.
Use your observed path; do not assume a particular operating system or cache layout.

## Apply a reversible local repair

1. Back up your OpenCode configuration in a private local location.
2. Copy the installed plugin's `src/`, `package.json`, and `LICENSE` to a
   persistent override directory outside the package cache. For example,
   `<OpenCode config directory>/overrides/opencode-plugin-litellm-0.10.0-stderr/`.
3. In the copied `src/plugin/index.ts`, replace the **three diagnostic calls**
   `console.log(` with `console.error(`. Confirm the diff changes only those
   calls. Keep warnings, discovery logic, model metadata, and API calls intact.
   If the version or number of calls differs, inspect that version first;
   do not apply this patch blindly.
4. Replace only the existing plugin entry with a file URL pointing to the
   copied `src/index.ts`. Preserve every other plugin and provider setting.

Example configuration fragment:

```json
{
  "plugin": ["file:///absolute/path/to/opencode-plugin-litellm-0.10.0-stderr/src/index.ts"]
}
```

On Windows, a file URL looks like `file:///C:/path/to/override/src/index.ts`.
Use a properly encoded file URL when the path contains spaces or special characters.
Keep only one active entry for this plugin; loading the original and override
together can duplicate hooks. Start a new OpenCode process after changing it.

This pins a locally repaired copy. Updating the registry package does not
automatically update that override. Keep its version and patch documented.

## Verify the real transport

Run one small prompt with the existing configured provider/model, keeping stdout
and stderr separate:

```powershell
opencode run --format json 'Reply with OK. Do not use tools.' 1> opencode-events.jsonl 2> opencode-diagnostics.txt
$runExit = $LASTEXITCODE
if ($runExit -ne 0) { throw "OpenCode exited with code $runExit; inspect diagnostics" }
$events = @(Get-Content -LiteralPath opencode-events.jsonl | Where-Object { $_.Trim() } | ForEach-Object { ConvertFrom-Json -InputObject $_ -ErrorAction Stop })
if ($events.Count -eq 0) { throw 'No JSON events received' }
if (@($events | Where-Object type -EQ error).Count) { throw 'OpenCode returned an error event' }
$finish = @($events | Where-Object { $_.type -eq 'step_finish' -and $_.part.reason -eq 'stop' })
if ($finish.Count -eq 0) { throw 'No normal completion event received' }
```

Inspect the assistant text event for the requested response and the diagnostics
for normal model discovery. A successful CLI exit alone is insufficient.
Then rerun the consumer that originally failed. Transport success does not
automatically establish routing quality, tool execution, or delivery success.

Do not silence the problem by stripping arbitrary non-JSON lines in the consumer
or merging stderr into stdout. Preserve malformed events and real model/auth
errors as failures. The repair moves diagnostics to their correct stream.

## Roll back or return to upstream

Restore the original plugin entry from your private backup, preserving any
unrelated configuration changes made since that backup. Start a new process and
repeat the probe. Leave the override available until the replacement is verified.

Before returning to an upstream release, inspect its logging changes and run the
same transport probe. Do not assume a newer version contains this fix. Report
the plugin/OpenCode versions, offending log prefix, and a sanitized minimal
reproduction to the upstream project if needed.

Upstream source: <https://github.com/yuseferi/opencode-litellm>.
