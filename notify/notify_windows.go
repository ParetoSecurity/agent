package notify

import (
	"unsafe"

	"github.com/caarlos0/log"
	"golang.org/x/sys/windows"
)

func Blocking(message string) {
	// Use MessageBoxW from user32.dll
	user32 := windows.NewLazySystemDLL("user32.dll")
	procMessageBoxW := user32.NewProc("MessageBoxW")

	// HWND = 0 (no owner), text, caption, MB_OK
	ret, _, err := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(message))),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Notification"))),
		0, // MB_OK
	)
	if ret == 0 {
		log.WithError(err).Error("Failed to send notification")
		return
	}
	return
}
