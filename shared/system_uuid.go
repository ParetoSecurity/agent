//go:build !windows
// +build !windows

package shared

import (
	"bytes"
	"fmt"
	"net"
	"sort"

	"github.com/google/uuid"
)

// systemUUID generates a unique system identifier based on a deterministically
// selected network interface's hardware address (MAC address). Interfaces are
// filtered (no loopback, no all-zero MAC), sorted by index, and preference is
// given to interfaces that are up.
func systemUUID() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	var candidates []net.Interface
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(iface.HardwareAddr) >= 6 && !bytes.Equal(iface.HardwareAddr, make([]byte, len(iface.HardwareAddr))) {
			candidates = append(candidates, iface)
		}
	}

	// Sort by index for deterministic ordering
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Index < candidates[j].Index
	})

	// Prefer an interface that is up
	var best *net.Interface
	for i := range candidates {
		if candidates[i].Flags&net.FlagUp != 0 {
			best = &candidates[i]
			break
		}
	}
	if best == nil && len(candidates) > 0 {
		best = &candidates[0]
	}

	if best == nil {
		return "", fmt.Errorf("no network interface found")
	}

	nsUUID := uuid.NewSHA1(uuid.NameSpaceOID, best.HardwareAddr)
	return nsUUID.String(), nil
}
