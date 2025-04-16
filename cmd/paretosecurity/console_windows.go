//go:build windows

package main

import (
	"syscall"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole  = kernel32.NewProc("AttachConsole")
	ATTACH_PARENT_PROC = ^uintptr(0) // 0xFFFFFFFF
)

func attachIfTerminal() {
	// returns non‑zero on success
	procAttachConsole.Call(ATTACH_PARENT_PROC)
}
