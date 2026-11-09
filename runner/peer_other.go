//go:build !linux

package runner

import "net"

// peerUID returns -1: the root helper only runs on Linux.
func peerUID(_ net.Conn) int {
	return -1
}
