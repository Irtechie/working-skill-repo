package key

import (
	"net/url"
	"sort"
	"strings"
)

func CanonicalQuery(raw string) (string, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return "", err
	}
	for name := range values {
		switch strings.ToLower(name) {
		case "utm_source", "utm_medium", "utm_campaign":
			delete(values, name)
		default:
			sort.Strings(values[name])
		}
	}
	return values.Encode(), nil
}
