package key

import "testing"

func TestCanonicalQuerySortsFiltersAndPreservesEmpty(t *testing.T) {
	got, err := CanonicalQuery("z=2&utm_medium=email&z=1&A=&utm_SOURCE=x&q=hello+world")
	if err != nil {
		t.Fatal(err)
	}
	if want := "A=&q=hello+world&z=1&z=2"; got != want {
		t.Fatalf("query=%q, want %q", got, want)
	}
}

func TestCanonicalQueryRejectsMalformedEscapes(t *testing.T) {
	if _, err := CanonicalQuery("bad=%zz"); err == nil {
		t.Fatal("expected malformed escape error")
	}
}
