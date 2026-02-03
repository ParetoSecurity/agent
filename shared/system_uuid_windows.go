//go:build windows
// +build windows

package shared

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// systemUUID generates a unique system identifier by combining the BIOS serial number
// with the first available MAC address. Using both provides a more stable and unique
// identifier than either alone.
func systemUUID() (string, error) {
	var seed []byte

	// Add serial number to seed
	serial, err := SystemSerial()
	if err == nil {
		serial = strings.TrimSpace(serial)
		if serial != "" && serial != "Unknown" {
			seed = append(seed, []byte(serial)...)
		}
	}

	// Add MAC address to seed, selected deterministically:
	// prefer interfaces that are up with a valid MAC, sorted by index.
	interfaces, err := net.Interfaces()
	if err == nil {
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
		if best != nil {
			seed = append(seed, best.HardwareAddr...)
		}
	}

	if len(seed) == 0 {
		return "", fmt.Errorf("no serial number or network interface found")
	}

	nsUUID := uuid.NewSHA1(uuid.NameSpaceOID, seed)
	return nsUUID.String(), nil
}
