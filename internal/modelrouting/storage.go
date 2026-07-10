package modelrouting

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const defaultMaxStorageBytes int64 = 1 << 20

type StorageOptions struct {
	MaxBytes int64
	Resolver Resolver
	Now      time.Time
	Policy   PolicyContext
	Source   CatalogSource
}

func SaveCatalog(root, rel string, catalog Catalog, opts StorageOptions) error {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	if err := validateCatalog(catalog, opts.Policy, opts.Resolver, now, storageSource(opts.Source)); err != nil {
		return err
	}
	return SaveAtomicJSON(root, rel, catalog, opts.MaxBytes)
}

// SaveAtomicJSON writes a bounded JSON object without following symlinks and
// enforces a private user-local/run-state security profile. Callers remain
// responsible for validating the object schema before this storage boundary.
func SaveAtomicJSON(root, rel string, value any, maxBytes int64) error {
	return saveAtomicJSON(root, rel, value, maxBytes, true)
}

// SaveAtomicProjectJSON preserves normal repository sharing permissions while
// retaining the same strict path, symlink, size, and atomic-replacement checks.
// It must never be used for credentials, auth names, approvals, or trust state.
func SaveAtomicProjectJSON(root, rel string, value any, maxBytes int64) error {
	return saveAtomicJSON(root, rel, value, maxBytes, false)
}

// WithPrivateStateLock serializes user-local read/modify/write transactions
// across kbrouter processes. Callers must not wait for human input while held.
func WithPrivateStateLock(root string, mutate func() error) error {
	if mutate == nil {
		return ErrUnsafePath
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return err
	}
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return err
	}
	if err := secureStorageDirectoryChain(absRoot, absRoot); err != nil {
		return err
	}
	lockPath, err := safeStoragePath(absRoot, ".state.lock")
	if err != nil {
		return err
	}
	if info, statErr := os.Lstat(lockPath); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return ErrUnsafePath
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	}
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := file.Chmod(0o600); err != nil {
		return err
	}
	if err := secureStorageFileSecurity(lockPath); err != nil {
		return err
	}
	opened, err := file.Stat()
	if err != nil {
		return err
	}
	current, err := os.Lstat(lockPath)
	if err != nil || !os.SameFile(opened, current) {
		return ErrUnsafePath
	}
	if err := lockStorageFile(file); err != nil {
		return err
	}
	defer unlockStorageFile(file)
	return mutate()
}

type PrivateStateLock struct {
	file *os.File
}

func AcquirePrivateStateLock(root, name string, timeout time.Duration) (*PrivateStateLock, error) {
	if strings.TrimSpace(name) == "" || filepath.Base(name) != name || strings.ContainsAny(name, `/\`) {
		return nil, ErrUnsafePath
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return nil, err
	}
	if err := secureStorageDirectoryChain(absRoot, absRoot); err != nil {
		return nil, err
	}
	lockPath, err := safeStoragePath(absRoot, name)
	if err != nil {
		return nil, err
	}
	if info, statErr := os.Lstat(lockPath); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return nil, ErrUnsafePath
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return nil, statErr
	}
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			_ = file.Close()
		}
	}()
	if err := file.Chmod(0o600); err != nil {
		return nil, err
	}
	if err := secureStorageFileSecurity(lockPath); err != nil {
		return nil, err
	}
	opened, err := file.Stat()
	if err != nil {
		return nil, err
	}
	current, err := os.Lstat(lockPath)
	if err != nil || !os.SameFile(opened, current) {
		return nil, ErrUnsafePath
	}
	deadline := time.Now().Add(timeout)
	for {
		locked, lockErr := tryLockStorageFile(file)
		if lockErr != nil {
			return nil, lockErr
		}
		if locked {
			ok = true
			return &PrivateStateLock{file: file}, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for private state lock")
		}
		sleep := 25 * time.Millisecond
		if remaining := time.Until(deadline); remaining < sleep {
			sleep = remaining
		}
		if sleep > 0 {
			time.Sleep(sleep)
		}
	}
}

func (lock *PrivateStateLock) Close() error {
	if lock == nil || lock.file == nil {
		return nil
	}
	file := lock.file
	lock.file = nil
	unlockStorageFile(file)
	return file.Close()
}

func saveAtomicJSON(root, rel string, value any, maxBytes int64, private bool) error {
	directoryMode := os.FileMode(0o755)
	if private {
		directoryMode = 0o700
	}
	if err := os.MkdirAll(root, directoryMode); err != nil {
		return err
	}
	path, err := safeStoragePath(root, rel)
	if err != nil {
		return err
	}
	if private {
		if err := secureStorageDirectoryChain(root, filepath.Dir(path)); err != nil {
			return err
		}
	}
	mode := os.FileMode(0o644)
	if private {
		mode = 0o600
	}
	if info, statErr := os.Lstat(path); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return ErrUnsafePath
		}
		if private {
			if err := validateStorageFileSecurity(path); err != nil {
				return err
			}
			if info.Mode().Perm()&0o077 == 0 {
				mode = info.Mode().Perm()
			}
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	}
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	if int64(len(content)) > storageLimit(maxBytes) {
		return ErrStorageSizeExceeded
	}
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, directoryMode); err != nil {
		return err
	}
	if private {
		if err := secureStorageDirectoryChain(root, parent); err != nil {
			return err
		}
	}
	verifiedPath, err := safeStoragePath(root, rel)
	if err != nil || !samePath(path, verifiedPath) {
		return ErrUnsafePath
	}
	parentBefore, err := os.Stat(parent)
	if err != nil || !parentBefore.IsDir() {
		return ErrUnsafePath
	}
	temp, err := os.CreateTemp(parent, ".catalog-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return err
	}
	if private {
		if err := secureStorageFileSecurity(tempName); err != nil {
			_ = temp.Close()
			return err
		}
	}
	if _, err := temp.Write(content); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Chmod(mode); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	verifiedPath, err = safeStoragePath(root, rel)
	if err != nil || !samePath(path, verifiedPath) {
		return ErrUnsafePath
	}
	parentAfter, err := os.Stat(parent)
	if err != nil || !os.SameFile(parentBefore, parentAfter) {
		return ErrUnsafePath
	}
	if info, statErr := os.Lstat(path); statErr == nil && (info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular()) {
		return ErrUnsafePath
	} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	}
	if err := os.Rename(tempName, path); err != nil {
		return err
	}
	if err := os.Chmod(path, mode); err != nil {
		return err
	}
	if private {
		if err := secureStorageFileSecurity(path); err != nil {
			return err
		}
	}
	return syncDirectory(parent)
}

func LoadCatalog(root, rel string, opts StorageOptions) (Catalog, error) {
	var catalog Catalog
	if err := LoadStrictJSON(root, rel, &catalog, opts.MaxBytes); err != nil {
		return Catalog{}, err
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	if err := validateCatalog(catalog, opts.Policy, opts.Resolver, now, storageSource(opts.Source)); err != nil {
		return Catalog{}, err
	}
	return catalog, nil
}

// LoadStrictJSON reads a bounded JSON object from a replacement-resistant path.
// Duplicate keys, unknown fields, trailing values, symlinks, and non-regular
// files are rejected before the caller receives the decoded value.
func LoadStrictJSON(root, rel string, target any, maxBytes int64) error {
	return loadStrictJSON(root, rel, target, maxBytes, true)
}

// LoadStrictJSONBytes returns the exact bytes decoded through the same secure
// file handle so callers can bind a digest without a second-open race.
func LoadStrictJSONBytes(root, rel string, target any, maxBytes int64) ([]byte, error) {
	return loadStrictJSONBytes(root, rel, target, maxBytes, true)
}

// LoadStrictProjectJSON accepts normal tracked repository permissions while
// preserving strict JSON, bounds, regular-file, traversal, and symlink checks.
func LoadStrictProjectJSON(root, rel string, target any, maxBytes int64) error {
	return loadStrictJSON(root, rel, target, maxBytes, false)
}

func loadStrictJSON(root, rel string, target any, maxBytes int64, private bool) error {
	_, err := loadStrictJSONBytes(root, rel, target, maxBytes, private)
	return err
}

func loadStrictJSONBytes(root, rel string, target any, maxBytes int64, private bool) ([]byte, error) {
	path, err := safeStoragePath(root, rel)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, ErrUnsafePath
	}
	if private {
		if err := validateStorageDirectoryChain(root, filepath.Dir(path)); err != nil {
			return nil, err
		}
		if err := validateStorageFileSecurity(path); err != nil {
			return nil, err
		}
	}
	limit := storageLimit(maxBytes)
	if info.Size() > limit {
		return nil, ErrStorageSizeExceeded
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil || !os.SameFile(info, openedInfo) {
		return nil, ErrUnsafePath
	}
	verifiedPath, err := safeStoragePath(root, rel)
	if err != nil || !samePath(path, verifiedPath) {
		return nil, ErrUnsafePath
	}
	var buf bytes.Buffer
	if _, err := io.CopyN(&buf, file, limit+1); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if int64(buf.Len()) > limit {
		return nil, ErrStorageSizeExceeded
	}
	if err := rejectDuplicateJSONKeys(buf.Bytes()); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return nil, err
	}
	if err := requireJSONEOF(decoder); err != nil {
		return nil, err
	}
	return append([]byte(nil), buf.Bytes()...), nil
}

func decodeCatalogStrict(data []byte) (Catalog, error) {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return Catalog{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var catalog Catalog
	if err := decoder.Decode(&catalog); err != nil {
		return Catalog{}, err
	}
	if err := requireJSONEOF(decoder); err != nil {
		return Catalog{}, err
	}
	return catalog, nil
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("%w: trailing JSON value", ErrInvalidCatalog)
		}
		return err
	}
	return nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	return requireJSONEOF(decoder)
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("%w: object key is not a string", ErrInvalidCatalog)
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("%w: duplicate JSON key %q", ErrInvalidCatalog, key)
			}
			seen[key] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return fmt.Errorf("%w: unexpected JSON delimiter", ErrInvalidCatalog)
	}
}

func validateCatalog(catalog Catalog, policy PolicyContext, resolver Resolver, now time.Time, source CatalogSource) error {
	return ValidateCatalogStatic(catalog, source)
}

// ValidateCatalogStatic validates durable catalog structure and provenance
// without performing DNS or any other network activity. Endpoint resolution and
// trust checks must run immediately before a probe or dispatch.
func ValidateCatalogStatic(catalog Catalog, source CatalogSource) error {
	if catalog.SchemaVersion != CatalogSchemaVersion {
		return fmt.Errorf("%w: unsupported schema version", ErrInvalidCatalog)
	}
	if source == CatalogSourceRun && catalog.Fingerprint == "" {
		return fmt.Errorf("%w: run catalog fingerprint is required", ErrInvalidCatalog)
	}
	if catalog.Cohort != "" && catalog.Cohort != CohortUnspecified && catalog.Cohort != CohortInitialPilot {
		return fmt.Errorf("%w: unknown support cohort", ErrInvalidCatalog)
	}
	for _, surface := range catalog.Surfaces {
		if surface.Surface == "" || surface.Provider == "" || surface.Revision == "" || surface.ConfigHash == "" {
			return fmt.Errorf("%w: incomplete surface fingerprint", ErrInvalidCatalog)
		}
	}
	if catalog.Current.Route != nil {
		if catalog.Current.ModelID == "" || catalog.Current.Route.DisplayModelID != catalog.Current.ModelID {
			return fmt.Errorf("%w: current route identity mismatch", ErrInvalidCatalog)
		}
		if err := validateRouteProvenance(*catalog.Current.Route, source); err != nil {
			return err
		}
		if err := validateRouteStatic(*catalog.Current.Route); err != nil {
			return err
		}
	}
	aliases := make(map[string]struct{}, len(catalog.Routes))
	routeIDs := make(map[string]struct{}, len(catalog.Routes))
	for _, route := range catalog.Routes {
		if _, exists := aliases[route.Alias]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateAlias, route.Alias)
		}
		aliases[route.Alias] = struct{}{}
		if source == CatalogSourceUser {
			if _, exists := routeIDs[route.RouteID]; exists {
				return fmt.Errorf("%w: duplicate route id", ErrInvalidCatalog)
			}
			routeIDs[route.RouteID] = struct{}{}
		}
		if err := validateRouteProvenance(route, source); err != nil {
			return err
		}
		if err := validateRouteStatic(route); err != nil {
			return err
		}
	}
	if catalog.Fingerprint != "" {
		expected, err := ComputeCatalogFingerprint(catalog)
		if err != nil {
			return err
		}
		if expected != catalog.Fingerprint {
			return fmt.Errorf("%w: catalog fingerprint mismatch", ErrInvalidCatalog)
		}
	}
	return nil
}

func storageSource(source CatalogSource) CatalogSource {
	if source == "" {
		return CatalogSourceUser
	}
	return source
}

func safeStoragePath(root, rel string) (string, error) {
	if strings.TrimSpace(rel) == "" || filepath.IsAbs(rel) {
		return "", ErrUnsafePath
	}
	cleanRel := filepath.Clean(rel)
	if cleanRel == "." || cleanRel == ".." || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) || strings.HasPrefix(cleanRel, "../") || strings.HasPrefix(cleanRel, `..\`) {
		return "", ErrUnsafePath
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(filepath.Join(absRoot, cleanRel))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(absRoot, absPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", ErrUnsafePath
	}
	if err := rejectSymlinkAncestors(absRoot, absPath); err != nil {
		return "", err
	}
	return absPath, nil
}

func rejectSymlinkAncestors(root, path string) error {
	root = filepath.Clean(root)
	dir := filepath.Dir(filepath.Clean(path))
	for {
		info, err := os.Lstat(dir)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return ErrUnsafePath
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if samePath(dir, root) {
			return nil
		}
		next := filepath.Dir(dir)
		if next == dir {
			return ErrUnsafePath
		}
		dir = next
	}
}

func syncDirectory(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func samePath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func storageLimit(value int64) int64 {
	if value <= 0 {
		return defaultMaxStorageBytes
	}
	return value
}

func secureStorageDirectoryChain(root, parent string) error {
	return walkStorageDirectoryChain(root, parent, secureStorageDirectorySecurity)
}

func validateStorageDirectoryChain(root, parent string) error {
	return walkStorageDirectoryChain(root, parent, validateStorageDirectorySecurity)
}

func walkStorageDirectoryChain(root, parent string, check func(string) error) error {
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return err
	}
	absParent, err := filepath.Abs(filepath.Clean(parent))
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(absRoot, absParent)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return ErrUnsafePath
	}
	current := absRoot
	if err := check(current); err != nil {
		return err
	}
	if relative == "." {
		return nil
	}
	for _, element := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, element)
		if err := check(current); err != nil {
			return err
		}
	}
	return nil
}
