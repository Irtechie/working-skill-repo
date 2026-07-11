package modelrouting

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProjectPolicy is repository-controlled and may only narrow selection.
// It deliberately contains no approvals or credential bindings.
type ProjectPolicy struct {
	ProjectID           string         `json:"project_id"`
	AllowedDestinations []string       `json:"allowed_destinations,omitempty"`
	AllowedAliases      []string       `json:"allowed_aliases,omitempty"`
	MaxRetention        RetentionClass `json:"max_retention,omitempty"`
	DenyCurrentFallback bool           `json:"deny_current_fallback,omitempty"`
}

// UserTrust is user-local trusted state. Repository content must never populate it.
type UserTrust struct {
	ProjectID         string             `json:"project_id"`
	RouteApprovals    []RouteApproval    `json:"route_approvals,omitempty"`
	RouteDenials      []RouteDenial      `json:"route_denials,omitempty"`
	EndpointApprovals []EndpointApproval `json:"endpoint_approvals,omitempty"`
	AuthBindings      []AuthBinding      `json:"auth_bindings,omitempty"`
}

type PolicyContext struct {
	Project                  ProjectPolicy     `json:"project"`
	Trusted                  UserTrust         `json:"trusted"`
	RouteSources             map[string]Route  `json:"-"`
	TrustedRouteStates       map[string]string `json:"-"`
	TrustedCurrentModelID    string            `json:"-"`
	TrustedCurrentSurface    string            `json:"-"`
	TrustedCurrentRouteState string            `json:"-"`
}

type RouteApproval struct {
	ProjectID        string    `json:"project_id"`
	RouteFingerprint string    `json:"route_fingerprint"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type RouteDenial struct {
	ProjectID        string    `json:"project_id"`
	RouteFingerprint string    `json:"route_fingerprint"`
	CreatedAt        time.Time `json:"created_at"`
}

type EndpointApproval struct {
	Origin    string    `json:"origin"`
	ProjectID string    `json:"project_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type AuthBinding struct {
	Env       string    `json:"env"`
	Adapter   string    `json:"adapter"`
	Origin    string    `json:"origin"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Resolver interface {
	LookupIP(ctx context.Context, host string) ([]net.IP, error)
}

type staticResolver map[string][]net.IP

func StaticResolver(values map[string][]net.IP) Resolver {
	copied := make(staticResolver, len(values))
	for host, ips := range values {
		copied[strings.ToLower(host)] = append([]net.IP(nil), ips...)
	}
	return copied
}

func (r staticResolver) LookupIP(_ context.Context, host string) ([]net.IP, error) {
	ips := r[strings.ToLower(host)]
	if len(ips) == 0 {
		return nil, errors.New("host not found")
	}
	return append([]net.IP(nil), ips...), nil
}

// ValidatedEndpoint is a dispatch-time input. Dispatchers must dial a PinnedIP,
// retain TLSServerName, and refuse cross-origin redirects rather than resolving again.
type ValidatedEndpoint struct {
	URL           *url.URL
	Origin        string
	PinnedIPs     []net.IP
	TLSServerName string
}

func ValidateRoute(route Route, policy PolicyContext, resolver Resolver, now time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ValidateRouteContext(ctx, route, policy, resolver, now)
}

func ValidateRouteContext(ctx context.Context, route Route, policy PolicyContext, resolver Resolver, now time.Time) error {
	if err := validateRouteSchema(route); err != nil {
		return err
	}
	_, err := ValidateEndpointContext(ctx, route, policy, resolver, now)
	return err
}

func validateRouteStatic(route Route) error {
	if err := validateRouteSchema(route); err != nil {
		return err
	}
	return validateEndpointStatic(route)
}

func validateEndpointStatic(route Route) error {
	if route.Endpoint == "" {
		if route.AuthEnv != "" {
			return ErrAuthOriginMismatch
		}
		return nil
	}
	parsed, err := url.Parse(route.Endpoint)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" || parsed.RawQuery != "" {
		return ErrUnsafeEndpoint
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && scheme != "http" {
		return ErrUnsafeEndpoint
	}
	host := parsed.Hostname()
	if strings.EqualFold(host, "localhost") {
		if route.Boundary != BoundaryPrivate {
			return fmt.Errorf("%w: localhost requires private trust boundary", ErrInvalidCatalog)
		}
		return nil
	}
	if ip := net.ParseIP(host); ip != nil {
		if unsafeMetadataIP(ip) {
			return ErrUnsafeEndpoint
		}
		private := ip.IsLoopback() || ip.IsPrivate()
		if private && route.Boundary != BoundaryPrivate {
			return fmt.Errorf("%w: private endpoint requires private trust boundary", ErrInvalidCatalog)
		}
		if !private && scheme != "https" {
			return ErrUnsafeEndpoint
		}
		return nil
	}
	if scheme != "https" {
		return ErrUnsafeEndpoint
	}
	return nil
}

func validateRouteForStorage(route Route, policy PolicyContext, resolver Resolver, now time.Time) error {
	if err := validateRouteSchema(route); err != nil {
		return err
	}
	_, err := validateEndpoint(context.Background(), route, policy, resolver, now, false)
	return err
}

func validateEndpointForStorage(route Route, policy PolicyContext, resolver Resolver, now time.Time) error {
	_, err := validateEndpoint(context.Background(), route, policy, resolver, now, false)
	return err
}

func ValidateEndpoint(route Route, policy PolicyContext, resolver Resolver, now time.Time) (ValidatedEndpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ValidateEndpointContext(ctx, route, policy, resolver, now)
}

// ValidateEndpointContext binds DNS resolution to the caller's probe or
// dispatch deadline. The returned IPs must be used directly by the caller.
func ValidateEndpointContext(ctx context.Context, route Route, policy PolicyContext, resolver Resolver, now time.Time) (ValidatedEndpoint, error) {
	return validateEndpoint(ctx, route, policy, resolver, now, true)
}

func validateEndpoint(ctx context.Context, route Route, policy PolicyContext, resolver Resolver, now time.Time, projectScoped bool) (ValidatedEndpoint, error) {
	if err := validateEndpointStatic(route); err != nil {
		return ValidatedEndpoint{}, err
	}
	if route.Endpoint == "" {
		return ValidatedEndpoint{}, nil
	}
	parsed, err := url.Parse(route.Endpoint)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" || parsed.RawQuery != "" {
		return ValidatedEndpoint{}, ErrUnsafeEndpoint
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && scheme != "http" {
		return ValidatedEndpoint{}, ErrUnsafeEndpoint
	}
	host := parsed.Hostname()
	ips, err := literalOrResolvedIPs(ctx, host, resolver)
	if err != nil || len(ips) == 0 {
		return ValidatedEndpoint{}, ErrUnsafeEndpoint
	}
	private := false
	for _, ip := range ips {
		if unsafeMetadataIP(ip) {
			return ValidatedEndpoint{}, ErrUnsafeEndpoint
		}
		if ip.IsLoopback() || ip.IsPrivate() {
			private = true
		}
	}
	origin := endpointOrigin(parsed)
	if private {
		if route.Boundary != BoundaryPrivate {
			return ValidatedEndpoint{}, fmt.Errorf("%w: private endpoint requires private trust boundary", ErrInvalidCatalog)
		}
		if projectScoped {
			if !projectTrustMatches(policy, policy.Project.ProjectID) || !hasEndpointApproval(origin, policy.Trusted.ProjectID, policy.Trusted.EndpointApprovals, now) {
				return ValidatedEndpoint{}, ErrPrivateEndpointRequiresApproval
			}
		} else if !hasAnyEndpointApproval(origin, policy.Trusted.EndpointApprovals, now) {
			return ValidatedEndpoint{}, ErrPrivateEndpointRequiresApproval
		}
	} else if scheme != "https" {
		return ValidatedEndpoint{}, ErrUnsafeEndpoint
	}
	if route.AuthEnv != "" && !authBound(route.AuthEnv, route.Adapter, origin, policy.Trusted.AuthBindings, now) {
		return ValidatedEndpoint{}, ErrAuthOriginMismatch
	}
	return ValidatedEndpoint{URL: parsed, Origin: origin, PinnedIPs: cloneIPs(ips), TLSServerName: host}, nil
}

func hasAnyEndpointApproval(origin string, approvals []EndpointApproval, now time.Time) bool {
	for _, approval := range approvals {
		if strings.EqualFold(approval.Origin, origin) && now.Before(approval.ExpiresAt) {
			return true
		}
	}
	return false
}

func endpointOrigin(parsed *url.URL) string {
	return strings.ToLower(parsed.Scheme + "://" + parsed.Host)
}

func literalOrResolvedIPs(ctx context.Context, host string, resolver Resolver) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	if resolver == nil {
		return net.DefaultResolver.LookupIP(ctx, "ip", host)
	}
	return resolver.LookupIP(ctx, host)
}

func cloneIPs(values []net.IP) []net.IP {
	result := make([]net.IP, 0, len(values))
	for _, value := range values {
		result = append(result, append(net.IP(nil), value...))
	}
	return result
}

func unsafeMetadataIP(ip net.IP) bool {
	if ip == nil || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	return ip.Equal(net.ParseIP("169.254.169.254"))
}

func authBound(env, adapter, origin string, bindings []AuthBinding, now time.Time) bool {
	for _, binding := range bindings {
		if binding.Env == env && binding.Adapter == adapter && strings.EqualFold(binding.Origin, origin) && now.Before(binding.ExpiresAt) {
			return true
		}
	}
	return false
}

func hasEndpointApproval(origin, projectID string, approvals []EndpointApproval, now time.Time) bool {
	if projectID == "" {
		return false
	}
	for _, approval := range approvals {
		if approval.ProjectID == projectID && strings.EqualFold(approval.Origin, origin) && now.Before(approval.ExpiresAt) {
			return true
		}
	}
	return false
}

func routeAllowedByPolicy(route Route, req WorkRequest, policy PolicyContext, now time.Time) bool {
	if !projectTrustMatches(policy, req.ProjectID) {
		return false
	}
	project := policy.Project
	if len(project.AllowedDestinations) > 0 && !containsString(project.AllowedDestinations, route.Destination) {
		return false
	}
	if len(project.AllowedAliases) > 0 && !containsString(project.AllowedAliases, route.Alias) {
		return false
	}
	if !retentionAllowed(route.Retention, project.MaxRetention) {
		return false
	}
	if hasRouteDenial(route, req.ProjectID, policy.Trusted.RouteDenials, policy.RouteSources) {
		return false
	}
	if route.Boundary == BoundaryPrivate && !hasRouteApproval(route, req.ProjectID, policy.Trusted.RouteApprovals, policy.RouteSources, now) {
		return false
	}
	if req.SensitiveData {
		if project.MaxRetention == "" || project.MaxRetention == RetentionUnknown || !validRetention(project.MaxRetention) {
			return false
		}
		if len(project.AllowedDestinations) == 0 || !containsString(project.AllowedDestinations, route.Destination) {
			return false
		}
		if route.Retention == RetentionUnknown || route.TrainingUse != TrainingNo || strings.TrimSpace(route.Residency) == "" || strings.EqualFold(route.Residency, "unknown") || strings.TrimSpace(route.TrustProvenance) == "" {
			return false
		}
	}
	return true
}

func projectTrustMatches(policy PolicyContext, projectID string) bool {
	if projectID == "" {
		return false
	}
	if policy.Project.ProjectID != "" && policy.Project.ProjectID != projectID {
		return false
	}
	if policy.Trusted.ProjectID != "" && policy.Trusted.ProjectID != projectID {
		return false
	}
	return true
}

func retentionAllowed(routeRetention, max RetentionClass) bool {
	if !validRetention(routeRetention) || routeRetention == RetentionUnknown {
		return false
	}
	if max == "" {
		return true
	}
	if !validRetention(max) || max == RetentionUnknown {
		return false
	}
	return retentionRank(routeRetention) <= retentionRank(max)
}

func retentionRank(value RetentionClass) int {
	switch value {
	case RetentionNone:
		return 0
	case RetentionSession:
		return 1
	case RetentionLimited:
		return 2
	default:
		return 99
	}
}

func hasRouteApproval(route Route, projectID string, approvals []RouteApproval, sources map[string]Route, now time.Time) bool {
	fingerprint, err := ApprovalRouteFingerprint(route, sources)
	if err != nil {
		return false
	}
	for _, approval := range approvals {
		if approval.ProjectID == projectID && approval.RouteFingerprint == fingerprint && now.Before(approval.ExpiresAt) {
			return true
		}
	}
	return false
}

func hasRouteDenial(route Route, projectID string, denials []RouteDenial, sources map[string]Route) bool {
	fingerprint, err := ApprovalRouteFingerprint(route, sources)
	if err != nil {
		return false
	}
	for _, denial := range denials {
		if denial.ProjectID == projectID && denial.RouteFingerprint == fingerprint {
			return true
		}
	}
	return false
}

func CanonicalProjectIdentity(root string) (string, error) {
	abs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", err
	}
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("canonicalize project path: %w", err)
	}
	gitRoot, err := findGitRoot(canonical)
	if err != nil {
		return "", err
	}
	identityPath := canonical
	if gitRoot != "" {
		identityPath, err = gitCommonDirectory(gitRoot)
		if err != nil {
			return "", err
		}
	}
	identityPath = filepath.Clean(identityPath)
	objectID, err := fileObjectIdentity(identityPath)
	if err != nil {
		return "", fmt.Errorf("read project filesystem identity: %w", err)
	}
	// The filesystem object is authoritative: its identity survives path aliases
	// and moves, while a replacement clone receives a different object ID.
	// Linked worktrees share the Git common-directory object.
	sum := sha256.Sum256([]byte("filesystem-object-v1\x00" + objectID))
	return "project-sha256:" + hex.EncodeToString(sum[:]), nil
}

func findGitRoot(path string) (string, error) {
	for {
		marker := filepath.Join(path, ".git")
		if _, err := os.Lstat(marker); err == nil {
			return path, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		next := filepath.Dir(path)
		if next == path {
			return "", nil
		}
		path = next
	}
}

func gitCommonDirectory(gitRoot string) (string, error) {
	marker := filepath.Join(gitRoot, ".git")
	info, err := os.Lstat(marker)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", ErrUnsafePath
	}
	gitDir := marker
	if !info.IsDir() {
		data, err := os.ReadFile(marker)
		if err != nil {
			return "", err
		}
		line := strings.TrimSpace(string(data))
		if strings.ContainsAny(line, "\r\n\x00") || !strings.HasPrefix(line, "gitdir: ") {
			return "", fmt.Errorf("%w: malformed .git file", ErrUnsafePath)
		}
		gitDir = strings.TrimSpace(strings.TrimPrefix(line, "gitdir: "))
		if !filepath.IsAbs(gitDir) {
			gitDir = filepath.Join(gitRoot, gitDir)
		}
	}
	gitDir, err = filepath.EvalSymlinks(filepath.Clean(gitDir))
	if err != nil {
		return "", err
	}
	common := gitDir
	if data, err := os.ReadFile(filepath.Join(gitDir, "commondir")); err == nil {
		value := strings.TrimSpace(string(data))
		if value == "" || strings.ContainsAny(value, "\r\n\x00") {
			return "", fmt.Errorf("%w: malformed commondir", ErrUnsafePath)
		}
		if filepath.IsAbs(value) {
			common = value
		} else {
			common = filepath.Join(gitDir, value)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	common, err = filepath.EvalSymlinks(filepath.Clean(common))
	if err != nil {
		return "", err
	}
	return common, nil
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
