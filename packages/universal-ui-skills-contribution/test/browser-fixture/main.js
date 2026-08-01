import { createElement, StrictMode } from "react";
import { createRoot } from "react-dom/client";

import { contributionDefinition, routeLoaders } from "../../src/index.js";
import "../../src/skills-route.css";
import "./smoke.css";

const route = contributionDefinition.routes[0];
const { default: SkillsRoute } = await routeLoaders[route.id]();

createRoot(document.getElementById("root")).render(
  createElement(
    StrictMode,
    null,
    createElement(SkillsRoute, {
      context: {
        schemaVersion: "universal_ui.shell-context.v1",
        activeRouteId: route.id,
        navigation: "available",
        visibility: "all",
        session: {
          summary: "synthetic"
        }
      }
    })
  )
);
