//go:build windows

package checks

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	coInitApartmentThreaded = 0x2
	coInitDisableOle1DDE    = 0x4

	vtI4  = 0x3
	vtUI4 = 0x13
)

var (
	modOle32             = syscall.NewLazyDLL("Ole32.dll")
	modShell             = syscall.NewLazyDLL("Shell32.dll")
	modProps             = syscall.NewLazyDLL("Propsys.dll")
	procCoInit           = modOle32.NewProc("CoInitializeEx")
	procCoUninit         = modOle32.NewProc("CoUninitialize")
	procPropVariantClear = modOle32.NewProc("PropVariantClear")

	procSHGetPropertyStoreFromParsingName = modShell.NewProc("SHGetPropertyStoreFromParsingName")
	procPSGetPropertyKeyFromName          = modProps.NewProc("PSGetPropertyKeyFromName")

	iidPropertyStore = windows.GUID{Data1: 0x886D8EEB, Data2: 0x8CF2, Data3: 0x4446, Data4: [8]byte{0x8D, 0x02, 0xCD, 0xBA, 0x1D, 0xBD, 0xCF, 0x99}}
)

type propertyKey struct {
	Fmtid windows.GUID
	Pid   uint32
}

type propVariant struct {
	Vt         uint16
	Reserved1  uint16
	Reserved2  uint16
	Reserved3  uint16
	ValueUnion [8]byte
}

type iUnknownVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

type iPropertyStore struct {
	vtbl *iPropertyStoreVtbl
}

type iPropertyStoreVtbl struct {
	iUnknownVtbl
	GetCount uintptr
	GetAt    uintptr
	GetValue uintptr
	SetValue uintptr
	Commit   uintptr
}

func defaultSystemDrive() string {
	return strings.TrimRight(os.Getenv("SystemDrive"), "\\")
}

func defaultListFixedDrives() ([]string, error) {
	buf := make([]uint16, 254)
	n, err := windows.GetLogicalDriveStrings(uint32(len(buf)), &buf[0])
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	var drives []string
	start := 0
	for i, r := range buf {
		if r == 0 {
			if i > start {
				drive := windows.UTF16ToString(buf[start:i])
				if drive != "" {
					driveType := windows.GetDriveType(syscall.StringToUTF16Ptr(drive))
					if driveType == windows.DRIVE_FIXED {
						drives = append(drives, strings.TrimRight(drive, "\\"))
					}
				}
			}
			start = i + 1
		}
	}
	return drives, nil
}

func defaultGetBitLockerProtectionStatus(path string) (int, error) {
	// COM must be initialized on a single OS thread for this sequence.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := coInitializeEx(coInitApartmentThreaded | coInitDisableOle1DDE); err != nil {
		return 0, err
	}
	defer coUninitialize()

	store, err := shGetPropertyStoreFromParsingName(path)
	if err != nil {
		return 0, err
	}
	defer store.release()

	var key propertyKey
	if err := psGetPropertyKeyFromName("System.Volume.BitLockerProtection", &key); err != nil {
		return 0, err
	}

	var pv propVariant
	if err := store.getValue(&key, &pv); err != nil {
		return 0, err
	}
	defer propVariantClear(&pv)

	switch pv.Vt {
	case vtI4, vtUI4:
		value := *(*int32)(unsafe.Pointer(&pv.ValueUnion[0]))
		return int(value), nil
	default:
		return 0, fmt.Errorf("unexpected PROPVARIANT type: %d", pv.Vt)
	}
}

func coInitializeEx(flags uint32) error {
	hr, _, _ := procCoInit.Call(0, uintptr(flags))
	return hresultToError(hr)
}

func coUninitialize() {
	procCoUninit.Call()
}

func propVariantClear(pv *propVariant) {
	procPropVariantClear.Call(uintptr(unsafe.Pointer(pv)))
}

func shGetPropertyStoreFromParsingName(path string) (*iPropertyStore, error) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	var store *iPropertyStore
	hr, _, _ := procSHGetPropertyStoreFromParsingName.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		0,
		uintptr(unsafe.Pointer(&iidPropertyStore)),
		uintptr(unsafe.Pointer(&store)),
	)
	if err := hresultToError(hr); err != nil {
		return nil, err
	}
	if store == nil {
		return nil, fmt.Errorf("property store is nil")
	}
	return store, nil
}

func psGetPropertyKeyFromName(name string, key *propertyKey) error {
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	hr, _, _ := procPSGetPropertyKeyFromName.Call(
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(key)),
	)
	return hresultToError(hr)
}

func (store *iPropertyStore) getValue(key *propertyKey, pv *propVariant) error {
	hr, _, _ := syscall.SyscallN(
		store.vtbl.GetValue,
		uintptr(unsafe.Pointer(store)),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(pv)),
	)
	return hresultToError(hr)
}

func (store *iPropertyStore) release() {
	if store == nil || store.vtbl == nil {
		return
	}
	syscall.SyscallN(store.vtbl.Release, uintptr(unsafe.Pointer(store)))
}

func hresultToError(hr uintptr) error {
	if int32(hr) < 0 {
		return windows.Errno(hr)
	}
	return nil
}
