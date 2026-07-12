# Canonical Request Cache Key

Implement a stable cache key without changing public APIs.

- `key.CanonicalQuery` parses a raw query, removes `utm_source`, `utm_medium`,
  and `utm_campaign` case-insensitively, sorts keys and values, and encodes using
  standard URL query escaping.
- Empty values are preserved.
- Malformed percent escapes return an error.
- `request.CacheKey` lowercases the HTTP method, canonicalizes the URL host to
  lowercase, removes the default `:80`/`:443` port, cleans the path while
  preserving a trailing slash, and appends the canonical query.
- Fragments never affect the key.
- Input request and URL values must not be mutated.
