import { createElement, useMemo, useState } from "react";
import { skillCatalog } from "./catalog.generated.js";

const h = createElement;

export default function SkillsRoute({ context }) {
  const [query, setQuery] = useState("");
  const normalizedQuery = query.trim().toLowerCase();
  const visibleSkills = useMemo(
    () =>
      skillCatalog.filter((skill) =>
        [skill.name, skill.description, skill.category]
          .join(" ")
          .toLowerCase()
          .includes(normalizedQuery)
      ),
    [normalizedQuery]
  );
  const groups = useMemo(() => {
    const result = new Map();
    for (const skill of visibleSkills) {
      const entries = result.get(skill.category) ?? [];
      entries.push(skill);
      result.set(skill.category, entries);
    }
    return result;
  }, [visibleSkills]);

  return h(
    "section",
    {
      "aria-labelledby": "skills-contribution-title",
      className: "uui-skills",
      "data-testid": "skills-contribution"
    },
    h(
      "header",
      { className: "uui-skills__hero" },
      h("p", { className: "uui-skills__eyebrow" }, "Canonical capability catalog"),
      h("h2", { id: "skills-contribution-title", tabIndex: -1 }, "Skills"),
      h(
        "p",
        null,
        `${skillCatalog.length} portable skills owned by Irtechie/working-skill-repo.`
      ),
      h(
        "output",
        { className: "uui-skills__context", "data-testid": "skills-shell-context" },
        `${context.activeRouteId} · ${context.visibility}`
      )
    ),
    h(
      "label",
      { className: "uui-skills__search" },
      h("span", null, "Filter skills"),
      h("input", {
        onChange: (event) => setQuery(event.currentTarget.value),
        placeholder: "Search by name, purpose, or category",
        type: "search",
        value: query
      })
    ),
    h(
      "p",
      { "aria-live": "polite", className: "uui-skills__result-count" },
      `${visibleSkills.length} ${visibleSkills.length === 1 ? "skill" : "skills"} shown`
    ),
    ...[...groups].map(([category, skills]) =>
      h(
        "section",
        {
          "aria-labelledby": `skills-category-${category.toLowerCase().replaceAll(/[^a-z0-9]+/g, "-")}`,
          className: "uui-skills__group",
          key: category
        },
        h(
          "h3",
          { id: `skills-category-${category.toLowerCase().replaceAll(/[^a-z0-9]+/g, "-")}` },
          category
        ),
        h(
          "ul",
          { className: "uui-skills__grid" },
          ...skills.map((skill) =>
            h(
              "li",
              { className: "uui-skills__card", key: skill.id },
              h("h4", null, skill.name),
              h("p", null, skill.description),
              skill.argumentHint
                ? h(
                    "p",
                    { className: "uui-skills__hint" },
                    h("span", null, "Input "),
                    h("code", null, skill.argumentHint)
                  )
                : null,
              h(
                "a",
                {
                  href: skill.sourceUrl,
                  rel: "noreferrer",
                  target: "_blank"
                },
                "View canonical source"
              )
            )
          )
        )
      )
    )
  );
}
