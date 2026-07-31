//go:build !windows

package modelrouting

import "os"

func createStorageTempFile(parent string, _ bool) (*os.File, error) {
	return os.CreateTemp(parent, ".catalog-*.tmp")
}
