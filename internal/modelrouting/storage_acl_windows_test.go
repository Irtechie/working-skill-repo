//go:build windows

package modelrouting

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
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

func TestWindowsStorageOwnerMayBeSecuredIsNarrow(t *testing.T) {
	const currentSID = "S-1-5-21-100-200-300-1001"
	for _, owner := range []string{"O:" + currentSID, "o:ba", "O:" + builtinAdministratorsSID, "O:SY"} {
		if !windowsStorageOwnerMayBeSecured(owner, currentSID) {
			t.Fatalf("rejected permitted Windows storage owner %q", owner)
		}
	}
	for _, owner := range []string{"", "O:WD", "O:S-1-5-21-100-200-300-1002"} {
		if windowsStorageOwnerMayBeSecured(owner, currentSID) {
			t.Fatalf("accepted unsafe Windows storage owner %q", owner)
		}
	}
}

func TestWindowsStorageDescriptorMatchAcceptsSafeProtectedMetadataAndAceOrder(t *testing.T) {
	const sid = "S-1-5-21-100-200-300-1001"
	for _, descriptor := range []string{
		"O:" + sid + "D:P(A;;FA;;;SY)(A;;FA;;;" + sid + ")",
		"O:" + sid + "D:PAI(A;;FA;;;SY)(A;;FA;;;" + sid + ")",
		"O:" + sid + "D:PAR(A;;FA;;;SY)(A;;FA;;;" + sid + ")",
		"O:" + sid + "D:P(A;;FA;;;" + sid + ")(A;;FA;;;SY)",
		"O:" + sid + "G:SYD:P(A;;FA;;;SY)(A;;FA;;;" + sid + ")",
		"O:" + sid + "G:" + sid + "D:PAI(A;;FA;;;SY)(A;;FA;;;" + sid + ")",
	} {
		if !windowsStorageDescriptorMatches(descriptor, sid, false) {
			t.Fatalf("rejected equivalent file descriptor %q", descriptor)
		}
	}
	if !windowsStorageDescriptorMatches("O:"+sid+"D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;"+sid+")", sid, true) {
		t.Fatal("rejected equivalent directory descriptor")
	}
	if !windowsStorageDescriptorMatches(
		"O:"+sid+"G:"+sid+"D:PAI(A;OICI;FA;;;SY)(A;OICI;0x1200a9;;;S-1-5-21-100-200-300-1002)(A;OICI;FA;;;"+sid+")",
		sid,
		true,
	) {
		t.Fatal("rejected protected directory with a non-mutating traversal ACE")
	}
	for _, descriptor := range []string{
		"O:" + sid + "D:P(A;;FA;;;WD)(A;;FA;;;" + sid + ")",
		"O:" + sid + "D:PAI(A;ID;FA;;;SY)(A;;FA;;;" + sid + ")",
		"O:S-1-5-21-100-200-300-1002D:P(A;;FA;;;SY)(A;;FA;;;" + sid + ")",
		"O:" + sid + "D:P(A;;FA;;;SY)",
		"O:" + sid + "D:PAI(A;;FA;;;SY)(A;;0x1200a9;;;S-1-5-21-100-200-300-1002)(A;;FA;;;" + sid + ")",
		"O:" + sid + "D:PAI(A;OICIID;FA;;;SY)(A;OICI;0x1200a9;;;S-1-5-21-100-200-300-1002)(A;OICI;FA;;;" + sid + ")",
		"O:" + sid + "D:PAI(A;OICI;FA;;;SY)(A;OICI;0x1201bf;;;S-1-5-21-100-200-300-1002)(A;OICI;FA;;;" + sid + ")",
		"O:" + sid + "D:PAI(A;OICI;FA;;;SY)(A;OICI;0x11200a9;;;S-1-5-21-100-200-300-1002)(A;OICI;FA;;;" + sid + ")",
		"O:" + sid + "D:PAI(A;OICI;FA;;;SY)(A;OICI;0x21200a9;;;S-1-5-21-100-200-300-1002)(A;OICI;FA;;;" + sid + ")",
	} {
		directory := strings.Contains(descriptor, "OICI")
		if windowsStorageDescriptorMatches(descriptor, sid, directory) {
			t.Fatalf("accepted unsafe descriptor %q", descriptor)
		}
	}
}

func TestPrivateTempFileReceivesProtectedACLAtCreation(t *testing.T) {
	root := t.TempDir()
	sid, err := currentWindowsSID()
	if err != nil {
		t.Fatal(err)
	}
	descriptor, free, err := windowsDescriptor(
		"O:" + sid + "D:PAI(A;OICI;FA;;;SY)(A;OICI;0x1200a9;;;S-1-5-21-100-200-300-1002)(A;OICI;FA;;;" + sid + ")",
	)
	if err != nil {
		t.Fatal(err)
	}
	defer free()
	pathPointer, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		t.Fatal(err)
	}
	result, _, callErr := setFileSecurityW.Call(
		uintptr(unsafe.Pointer(pathPointer)),
		uintptr(ownerSecurityInformation|daclSecurityInformation|protectedDACLSSecurityInformation),
		descriptor,
	)
	if result == 0 {
		t.Fatalf("install host-style directory DACL: %v", callErr)
	}

	originalCreateFile := windowsCreateFile
	defer func() {
		windowsCreateFile = originalCreateFile
	}()
	protectedAtCreation := false
	windowsCreateFile = func(
		name *uint16,
		access uint32,
		mode uint32,
		security *syscall.SecurityAttributes,
		createMode uint32,
		attrs uint32,
		templateFile int32,
	) (syscall.Handle, error) {
		if security != nil && security.SecurityDescriptor != 0 {
			sddl, descriptorErr := windowsDescriptorString(
				security.SecurityDescriptor,
				ownerSecurityInformation|daclSecurityInformation,
			)
			protectedAtCreation = descriptorErr == nil && windowsStorageDescriptorMatches(sddl, sid, false)
		}
		return originalCreateFile(name, access, mode, security, createMode, attrs, templateFile)
	}
	file, err := createStorageTempFile(root, true)
	if err != nil {
		t.Fatal(err)
	}
	name := file.Name()
	defer os.Remove(name)
	defer file.Close()
	if !protectedAtCreation {
		t.Fatal("private temporary file was not created with the protected DACL")
	}
	if err := validateStorageFileSecurity(name); err != nil {
		t.Fatalf("private temporary file was not protected at creation: %v", err)
	}
}

func TestSecureWindowsStorageCommitsCurrentUserOwner(t *testing.T) {
	root := t.TempDir()
	if err := SaveAtomicJSON(root, "private.json", map[string]int{"schema_version": 1}, 1024); err != nil {
		t.Fatal(err)
	}
	sid, err := currentWindowsSID()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{root, filepath.Join(root, "private.json")} {
		owner, err := windowsStorageOwner(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.EqualFold(owner, "O:"+sid) {
			t.Fatalf("secured path %q owner=%q want current user %q", path, owner, sid)
		}
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

func TestSharedProjectLockPreservesInheritedWindowsACL(t *testing.T) {
	root := t.TempDir()
	lockPath := filepath.Join(root, "work-queue.lock")
	if err := os.WriteFile(lockPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	beforeDescriptor, err := getWindowsFileDescriptor(lockPath, ownerSecurityInformation|daclSecurityInformation)
	if err != nil {
		t.Fatal(err)
	}
	before, err := windowsDescriptorString(
		uintptr(unsafe.Pointer(&beforeDescriptor[0])),
		ownerSecurityInformation|daclSecurityInformation,
	)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := AcquireSharedProjectLock(root, "work-queue.lock", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	afterDescriptor, err := getWindowsFileDescriptor(lockPath, ownerSecurityInformation|daclSecurityInformation)
	if err != nil {
		t.Fatal(err)
	}
	after, err := windowsDescriptorString(
		uintptr(unsafe.Pointer(&afterDescriptor[0])),
		ownerSecurityInformation|daclSecurityInformation,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(before, after) {
		t.Fatalf("shared lock changed inherited Windows ACL\nbefore: %s\nafter:  %s", before, after)
	}
}
