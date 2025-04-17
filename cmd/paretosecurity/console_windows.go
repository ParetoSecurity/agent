//go:build windows

package main

import (
	"os"
	"syscall"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole    = kernel32.NewProc("AttachConsole")
	procAllocConsole     = kernel32.NewProc("AllocConsole")
	procSetStdHandle     = kernel32.NewProc("SetStdHandle")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
)

const (
	ATTACH_PARENT_PROCESS = ^uint32(0) // (DWORD)-1
	STD_INPUT_HANDLE      = uint32(-10 & 0xFFFFFFFF)
	STD_OUTPUT_HANDLE     = uint32(-11 & 0xFFFFFFFF)
	STD_ERROR_HANDLE      = uint32(-12 & 0xFFFFFFFF)
)

func attachConsole() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		return // Already attached to a console
	}

	r, _, _ := procAttachConsole.Call(uintptr(ATTACH_PARENT_PROCESS))
	if r == 0 {
		// Attach failed, allocate a new console
		procAllocConsole.Call()
	}

	// Redirect STDOUT and STDERR to console
	stdoutHandle, err := syscall.CreateFile(
		syscall.StringToUTF16Ptr("CONOUT$"),
		syscall.GENERIC_WRITE|syscall.GENERIC_READ,
		syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_READ,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0)
	if err == nil {
		procSetStdHandle.Call(uintptr(STD_OUTPUT_HANDLE), uintptr(stdoutHandle))
		procSetStdHandle.Call(uintptr(STD_ERROR_HANDLE), uintptr(stdoutHandle))
		os.Stdout = os.NewFile(uintptr(stdoutHandle), "CONOUT$")
		os.Stderr = os.NewFile(uintptr(stdoutHandle), "CONOUT$")
	}

	// Redirect STDIN to console
	stdinHandle, err := syscall.CreateFile(
		syscall.StringToUTF16Ptr("CONIN$"),
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0)
	if err == nil {
		procSetStdHandle.Call(uintptr(STD_INPUT_HANDLE), uintptr(stdinHandle))
		os.Stdin = os.NewFile(uintptr(stdinHandle), "CONIN$")
	}
}
