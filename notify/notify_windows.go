package notify

import (
	"unsafe"

	"github.com/caarlos0/log"
	"golang.org/x/sys/windows"
)

// Toast displays a modal message box on Windows using the Win32 API.
// If the message box fails to display, an error is logged.
func Toast(message string) {

	// Use modern Windows notification system via WinRT
	shell32 := windows.NewLazySystemDLL("shell32.dll")
	procShell_NotifyIconW := shell32.NewProc("Shell_NotifyIconW")

	// For Windows 10/11, use toast notifications
	// This is a simplified approach - for full toast notifications,
	// you'd typically use the WinRT APIs or a library like go-toast

	// Fallback to system tray notification
	const NIM_ADD = 0x00000000
	const NIF_MESSAGE = 0x00000001
	const NIF_ICON = 0x00000002
	const NIF_TIP = 0x00000004
	const NIF_INFO = 0x00000010

	// Create NOTIFYICONDATA structure (simplified)
	type NOTIFYICONDATA struct {
		CbSize           uint32
		HWnd             uintptr
		UID              uint32
		UFlags           uint32
		UCallbackMessage uint32
		HIcon            uintptr
		SzTip            [128]uint16
		DwState          uint32
		DwStateMask      uint32
		SzInfo           [256]uint16
		SzInfoTitle      [64]uint16
		DwInfoFlags      uint32
	}

	nid := NOTIFYICONDATA{
		CbSize:      uint32(unsafe.Sizeof(NOTIFYICONDATA{})),
		UFlags:      NIF_INFO,
		DwInfoFlags: 0x1, // NIIF_INFO
	}

	// Convert strings to UTF16
	copy(nid.SzInfo[:], windows.StringToUTF16("Notification")[0:min(len(windows.StringToUTF16("Notification")), 256)])
	copy(nid.SzInfoTitle[:], windows.StringToUTF16(message)[0:min(len(windows.StringToUTF16(message)), 64)])

	ret, _, err := procShell_NotifyIconW.Call(
		NIM_ADD,
		uintptr(unsafe.Pointer(&nid)),
	)

	if ret == 0 {
		log.WithError(err).Error("Failed to send notification")
	}
}
