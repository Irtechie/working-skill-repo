# Accessible Profile Card

Implement `index.html`, `styles.css`, and `app.js`.

- Use `<article class="profile-card" aria-labelledby="profile-name">`.
- The name is an `h2` with `id="profile-name"`.
- Include a button with `aria-expanded="false"` and
  `aria-controls="profile-details"`.
- Include details with `id="profile-details"` and the `hidden` attribute.
- Clicking the button toggles `hidden` and updates `aria-expanded`.
- Pressing Escape while details are visible hides them, resets
  `aria-expanded`, and returns focus to the button.
- CSS must include a visible `:focus-visible` rule.
- Use no dependencies or inline event handlers.
