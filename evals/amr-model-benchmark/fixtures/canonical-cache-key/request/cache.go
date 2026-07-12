package request

import "net/http"

// CacheKey returns the stable request cache key described in SPEC.md.
func CacheKey(req *http.Request) (string, error) {
	panic("TODO")
}
