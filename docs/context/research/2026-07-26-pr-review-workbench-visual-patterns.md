# PR Review Workbench Visual Patterns

Checked: 2026-07-26
Budget mode: standard

## Question

What do strong code-review and workflow products do that the current
`pr-review-workbench` does not, and which patterns should be adopted without
turning the artifact into a noisy dashboard or unsafe GitHub replacement?

## Findings

The current renderer is not actually workflow-like. It leads with a large
status heading and five fact cards, then offers `Impact`, `Files`, and `Gaps`
tabs. The older `interactive-workflow-workbench` example had the missing
structure: a topology, ordered steps, a decision gate, explicit terminal
states, and a coordinated evidence inspector.

Five patterns are worth adopting:

1. **A dominant dependency path.** GitHub Actions, Argo, and Prefect use graphs
   because order, parallel work, gates, and failure state are easier to
   understand spatially than as prose. The PR view should show:
   `review request → changed areas → behavioral impact → evidence gate →
   human decision`, with the failed gate ending at a visible repair/block state.
2. **Relationship context without losing the selected unit.** Graphite keeps
   one PR understandable on its own while showing where it sits in a stack.
   The workbench should keep one PR as the unit of review while connecting
   changed areas to behavioral claims and proof.
3. **Progress by meaningful review unit.** GitHub tracks viewed files;
   Reviewable goes further with a file/revision matrix and explicit reviewed
   state. This offline artifact cannot truthfully track reviewer progress, but
   it can order source-backed application effects before the files and proof
   that support them. Path-derived source, verification, docs/design, and
   delivery/config groups remain an explicitly labeled fallback.
4. **One merge/readiness surface.** GitLab widgets and GitHub status checks
   group pipeline, deployment, security, and external-check state near the
   decision. The workbench should use one evidence gate that distinguishes
   complete, missing, failed, and unsupported proof. It must not turn “checks
   passed” into “the PR is correct.”
5. **Click-to-inspect artifacts and evidence.** Argo puts artifacts directly on
   the workflow DAG and opens a detail panel when selected. Every workbench
   node should open a coordinated inspector containing its meaning, proof,
   source anchors, and caveat.

### Order By App Impact, Not The Diff

A flat file list and path-derived buckets are useful fallbacks, but they are not
impact analysis. Nx computes affected projects from both the Git change set and
the project dependency graph. Sourcegraph distinguishes compiler-derived
definitions/references from heuristic text search and uses reverse references
to determine downstream impact. GitHub's built-in code navigation likewise
links definitions and references. CodeScene adds a different signal: historical
change coupling can reveal modules that behave as logical dependencies even
when the static architecture does not make that relationship obvious.

The primary review order should therefore be:

1. user-facing, API, data/state, auth/security, or external-mutation boundary;
2. runtime entry point and central changed behavior;
3. directly affected callers, consumers, workflows, or projects;
4. transitive downstream impact and compatibility risk;
5. executable proof covering those paths;
6. delivery/config, docs, generated files, and mechanical churn.

This order must come from a commit-pinned impact packet with source spans and
typed relationship evidence. Acceptable evidence includes repo-native project
graphs, precise symbol references, import/call edges, route-to-handler maps,
tests linked to changed behavior, and bounded file-native graph packets.
Heuristic search or LLM inference must be labeled as fallback. Path grouping
must never be presented as actual application impact.

The visual hierarchy should be:

1. identity and decision state;
2. the workflow topology;
3. one recommended next action;
4. guided review steps;
5. grouped files and detailed evidence.

Polish should come from alignment, spacing, typography, semantic color,
responsive behavior, and useful motion. It should not come from gradients on
every card, fake confidence scores, decorative charts, or a wall of badges.

## Sources

- [GitHub Actions workflow visualization](https://docs.github.com/en/actions/how-tos/monitor-workflows?tool=webui)
- [GitHub pull-request review progress](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/reviewing-proposed-changes-in-a-pull-request)
- [GitHub status checks](https://docs.github.com/en/pull-requests/reference/status-checks)
- [Graphite PR page anatomy](https://graphite.com/docs/update-pull-requests)
- [Graphite stacked-review guidance](https://graphite.com/docs/best-practices-for-reviewing-stacks)
- [Reviewable file matrix and reviewed state](https://docs.reviewable.io/files.html)
- [Reviewable review-completion surface](https://docs.reviewable.io/reviews)
- [GitLab merge-request widgets](https://docs.gitlab.com/user/project/merge_requests/widgets/)
- [Argo DAG workflow model](https://argo-workflows.readthedocs.io/en/latest/walk-through/dag/)
- [Argo artifact visualization](https://argo-workflows.readthedocs.io/en/latest/artifact-visualization/)
- [Prefect task dependency graph](https://docs.prefect.io/v3/get-started/quickstart)
- [Nx affected projects](https://nx.dev/docs/features/ci-features/affected)
- [Nx project and task graphs](https://nx.dev/docs/features/explore-graph)
- [Sourcegraph precise code navigation](https://sourcegraph.com/docs/code-navigation)
- [GitHub code navigation](https://docs.github.com/en/repositories/working-with-files/using-files/navigating-code-on-github)
- [CodeScene change coupling](https://codescene.digitgaming.com/docs/guides/technical/change-coupling.html)

## Applies When

- Rendering one complex PR or an inbox of PRs for human review.
- A reviewer needs change relationships and proof gates, not another prose
  summary.
- The artifact must remain offline, inert, commit-pinned, and downloadable.

## Stale When

- The renderer gains a different authoritative interaction model.
- GitHub adds repository-native execution of arbitrary review HTML.
- The evidence packet starts carrying explicit code-ownership or reviewer
  progress state.

## Rejected Approaches

- A prettier version of the existing fact-card page.
- One enormous graph containing every changed file.
- A full topology dimmed behind every review step.
- Fake completion percentages or AI confidence scores.
- Reimplementing diffs, comments, approvals, or merge controls.
- External JavaScript, fonts, images, network calls, or browser storage.

## Impact On Current Project

- Replace the current hero-and-tabs renderer with a real decision topology.
- Add a guided five-step review path and coordinated evidence inspector.
- Require a commit-pinned impact order for the rich workbench. Grouping by file
  role remains an explicitly labeled fallback, never an impact claim.
- Keep small, low-risk, single-area PRs on the compact PR first screen. A rich
  workbench is automatic only when source-backed impact or a material boundary,
  downstream path, reconstruction cost, or repair path justifies it.
- Keep the existing fail-closed packet, SHA pin, CSP, and explicit review
  mutation boundary.
- Expand deterministic tests around topology, terminals, application-impact
  ordering, fallback labeling, inspector controls, and forbidden resources.
