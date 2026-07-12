package request

import (
	"net/http"
	"testing"
)

func TestCacheKeyCanonicalizesWithoutMutation(t *testing.T) {
	req, err := http.NewRequest("GET", "HTTPS://Example.COM:443/a/../b/?z=2&z=1&utm_SOURCE=x&a=&q=hello+world#ignored", nil)
	if err != nil {
		t.Fatal(err)
	}
	before := req.URL.String()
	got, err := CacheKey(req)
	if err != nil {
		t.Fatal(err)
	}
	want := "get https://example.com/b/?a=&q=hello+world&z=1&z=2"
	if got != want {
		t.Fatalf("key=%q, want %q", got, want)
	}
	if req.URL.String() != before {
		t.Fatalf("request URL mutated: before=%q after=%q", before, req.URL.String())
	}
}

func TestCacheKeyPreservesNonDefaultPortAndTrailingSlash(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://EXAMPLE.com:8080/api/?b=2&a=1", nil)
	got, err := CacheKey(req)
	if err != nil {
		t.Fatal(err)
	}
	if want := "post http://example.com:8080/api/?a=1&b=2"; got != want {
		t.Fatalf("key=%q, want %q", got, want)
	}
}

func TestCacheKeyRejectsMalformedQuery(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/path", nil)
	req.URL.RawQuery = "bad=%zz"
	if _, err := CacheKey(req); err == nil {
		t.Fatal("expected malformed query error")
	}
}
