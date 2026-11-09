//go:build linux

package runner

import (
	"net"
	"syscall"
)

// peerUID returns the uid of the process at the other end of a Unix socket.
// It returns -1 when the kernel does not report one.
func peerUID(conn net.Conn) int {
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return -1
	}
	raw, err := unixConn.SyscallConn()
	if err != nil {
		return -1
	}
	uid := -1
	_ = raw.Control(func(fd uintptr) {
		cred, err := syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
		if err == nil {
			uid = int(cred.Uid)
		}
	})
	return uid
}
