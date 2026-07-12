const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const vm = require("node:vm");

function withoutComments(value) {
  return value
    .replace(/<!--[\s\S]*?-->/g, "")
    .replace(/\/\*[\s\S]*?\*\//g, "")
    .replace(/(^|\s)\/\/.*$/gm, "");
}

test("accessible structure", () => {
  const html = withoutComments(fs.readFileSync("index.html", "utf8"));
  assert.match(html, /<article[^>]*class=["'][^"']*\bprofile-card\b[^"']*["'][^>]*aria-labelledby=["']profile-name["']/i);
  assert.match(html, /<h2[^>]*id=["']profile-name["']/i);
  assert.match(html, /<button[^>]*aria-expanded=["']false["'][^>]*aria-controls=["']profile-details["']/i);
  assert.match(html, /id=["']profile-details["'][^>]*\bhidden\b/i);
  assert.doesNotMatch(html, /\sonclick\s*=/i);
  const css = withoutComments(fs.readFileSync("styles.css", "utf8"));
  assert.match(css, /:focus-visible\s*\{[^}]+(?:outline|box-shadow)\s*:/i);
});

test("interactive behavior", () => {
  const handlers = {};
  const button = {
    attributes: {"aria-expanded": "false"},
    focused: false,
    setAttribute(name, value) { this.attributes[name] = String(value); },
    getAttribute(name) { return this.attributes[name]; },
    addEventListener(type, handler) { handlers[`button:${type}`] = handler; },
    focus() { this.focused = true; }
  };
  const details = {hidden: true};
  const document = {
    getElementById(id) {
      if (id === "profile-details") return details;
      if (id === "profile-toggle" || id === "details-toggle") return button;
      return null;
    },
    querySelector(selector) {
      if (selector.includes("aria-controls")) return button;
      if (selector.includes("profile-details")) return details;
      return null;
    },
    addEventListener(type, handler) {
      if (type === "DOMContentLoaded") handler();
      else handlers[`document:${type}`] = handler;
    }
  };
  vm.runInNewContext(fs.readFileSync("app.js", "utf8"), {document, console});
  assert.ok(handlers["button:click"], "button click handler was not registered");
  handlers["button:click"]();
  assert.equal(details.hidden, false);
  assert.equal(button.getAttribute("aria-expanded"), "true");
  const escape = handlers["document:keydown"] || handlers["button:keydown"];
  assert.ok(escape, "Escape key handler was not registered");
  escape({key: "Escape", preventDefault() {}});
  assert.equal(details.hidden, true);
  assert.equal(button.getAttribute("aria-expanded"), "false");
  assert.equal(button.focused, true);
});
