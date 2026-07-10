//go:build windows

package modelrouting

import (
	"errors"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

func TestStrictStorageRejectsUnsafeWindowsDACL(t *testing.T) {
	root := t.TempDir()
	if err := SaveAtomicJSON(root, "private.json", map[string]int{"schema_version": 1}, 1024); err != nil {
		t.Fatal(err)
	}
	sid, err := currentWindowsSID()
	if err != nil {
		t.Fatal(err)
	}
	descriptor, free, err := windowsDescriptor("O:" + sid + "D:P(A;;FA;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	defer free()
	pathPointer, err := syscall.UTF16PtrFromString(filepath.Join(root, "private.json"))
	if err != nil {
		t.Fatal(err)
	}
	result, _, callErr := setFileSecurityW.Call(uintptr(unsafe.Pointer(pathPointer)), uintptr(ownerSecurityInformation|daclSecurityInformation|protectedDACLSSecurityInformation), descriptor)
	if result == 0 {
		t.Fatalf("install unsafe test DACL: %v", callErr)
	}
	var loaded map[string]int
	if err := LoadStrictJSON(root, "private.json", &loaded, 1024); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("unsafe Windows DACL error=%v", err)
	}
}

func TestProjectJSONDoesNotReplaceRepositoryDACL(t *testing.T) {
	root := t.TempDir()
	beforeDescriptor, err := getWindowsFileDescriptor(root, ownerSecurityInformation|daclSecurityInformation)
	if err != nil {
		t.Fatal(err)
	}
	before, err := windowsDescriptorString(uintptr(unsafe.Pointer(&beforeDescriptor[0])), ownerSecurityInformation|daclSecurityInformation)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveAtomicProjectJSON(root, "kb-models.json", map[string]int{"schema_version": 1}, 1024); err != nil {
		t.Fatal(err)
	}
	afterDescriptor, err := getWindowsFileDescriptor(root, ownerSecurityInformation|daclSecurityInformation)
	if err != nil {
		t.Fatal(err)
	}
	after, err := windowsDescriptorString(uintptr(unsafe.Pointer(&afterDescriptor[0])), ownerSecurityInformation|daclSecurityInformation)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(before, after) {
		t.Fatalf("project save changed repository DACL\nbefore: %s\nafter:  %s", before, after)
	}
}
