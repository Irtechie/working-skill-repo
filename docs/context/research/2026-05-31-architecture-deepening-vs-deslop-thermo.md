# Architecture Deepening vs Deslop vs Thermo-Nuclear

Checked: 2026-05-31
Budget mode: lean

## Question

How should Matt Pocock's `improve-codebase-architecture` skill compare with
deslop-style cleanup and this repo's thermo-nuclear code-quality reviewer?

## Findings

- `improve-codebase-architecture` is a strategic architecture exploration lane.
  It looks for deepening opportunities: shallow modules, weak seams, tightly
  coupled modules, poor test surfaces, and missing locality.
- Its useful vocabulary is module, interface, implementation, depth, seam,
  adapter, leverage, and locality. Its useful tests are the deletion test and
  "the interface is the test surface."
- Deslop is a hygiene cleanup lane. It targets AI artifacts such as excessive
  comments, debug statements, placeholders, ghost code, and mechanical residue.
- Thermo-nuclear review is a diff review persona. It catches structural debt
  introduced or preserved by a change: file sprawl, spaghetti branching, thin
  wrappers, boundary/type mud, wrong-layer logic, and missed code-judo deletion.
- These should not be merged. They operate at different times and scopes.

## Sources

- Matt Pocock skill page:
  `https://www.skills.sh/mattpocock/skills/improve-codebase-architecture`
- Upstream `SKILL.md`:
  `https://raw.githubusercontent.com/mattpocock/skills/main/improve-codebase-architecture/SKILL.md`
- Upstream `LANGUAGE.md`:
  `https://raw.githubusercontent.com/mattpocock/skills/main/improve-codebase-architecture/LANGUAGE.md`
- Deslop listing:
  `https://mcpmarket.com/tools/skills/deslop-ai-code-cleanup`
- Local thermo-nuclear reviewer:
  `.github/agents/thermo-nuclear-code-quality-reviewer.agent.md`

## Applies When

- Use architecture deepening when the question is where the codebase should get
  deeper, more testable, or more AI-navigable.
- Use deslop when the question is whether AI residue/noise should be removed.
- Use thermo-nuclear review when reviewing whether a specific diff adds or misses
  structural simplification.

## Stale When

- Upstream `improve-codebase-architecture` changes its core process or bundled
  vocabulary.
- This repo replaces thermo-nuclear review or adds a dedicated architecture
  audit skill.
- A first-party deslop implementation is added to this bundle.

## Rejected Approaches

- Do not merge architecture deepening into thermo-nuclear review. The former is
  exploratory and candidate-driven; the latter is a review gate.
- Do not merge architecture deepening into deslop. Hygiene cleanup is not the
  same as codebase architecture.
- Do not make architecture deepening a base-layer skill; it is specialized and
  should be loaded lazily.

## Impact On Current Project

- Add an architecture-deepening lane to the skill-minimalism epic.
- Borrow the deletion test and interface-as-test-surface language for
  thermo-nuclear or a future architecture-deepening skill.
- Keep the lane optional until loaded-surface measurement proves its cost and
  evals cover its routing.
