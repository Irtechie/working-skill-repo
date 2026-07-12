# Retry-After Parser

Implement `retry.ParseRetryAfter(value, now, max)`.

- Accept a non-negative integer number of delta-seconds.
- Accept an HTTP-date supported by `http.ParseTime`.
- Trim surrounding whitespace before parsing.
- Clamp valid future durations to `max`.
- Return zero for valid HTTP-dates at or before `now`.
- Preserve exact whole-second durations; do not accept fractional seconds.
- Return `ErrInvalidRetryAfter` for:
  - empty input;
  - signed delta-seconds such as `+1` or `-1`;
  - fractional values;
  - malformed values or dates;
  - numeric or duration overflow;
  - negative `max`.
- Do not change the public API or tests.
