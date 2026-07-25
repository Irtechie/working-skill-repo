package request

import (
	"fmt"
	"net"
	"net/http"
	"path"
	"strings"

	"example.invalid/cachekey/key"
)

func CacheKey(req *http.Request) (string, error) {
	if req == nil || req.URL == nil {
		return "", fmt.Errorf("request URL is required")
	}
	cloned := *req.URL
	host := strings.ToLower(cloned.Host)
	if hostname, port, err := net.SplitHostPort(host); err == nil {
		if (cloned.Scheme == "http" && port == "80") || (cloned.Scheme == "https" && port == "443") {
			host = hostname
		}
	}
	trailingSlash := strings.HasSuffix(cloned.Path, "/")
	cleanPath := path.Clean(cloned.Path)
	if cleanPath == "." {
		cleanPath = "/"
	}
	if trailingSlash && cleanPath != "/" {
		cleanPath += "/"
	}
	query, err := key.CanonicalQuery(cloned.RawQuery)
	if err != nil {
		return "", err
	}
	result := strings.ToLower(req.Method) + " " + strings.ToLower(cloned.Scheme) + "://" + host + cleanPath
	if query != "" {
		result += "?" + query
	}
	return result, nil
}
