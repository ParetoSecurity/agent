package shared

import (
	"bytes"
	"net"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemUUID_Deterministic(t *testing.T) {
	// systemUUID should return the same value on repeated calls.
	id1, err1 := systemUUID()
	id2, err2 := systemUUID()

	// Both calls should have the same outcome.
	assert.Equal(t, err1 == nil, err2 == nil)
	if err1 == nil {
		assert.Equal(t, id1, id2, "systemUUID must be deterministic across calls")
		assert.NotEmpty(t, id1)
	}
}

func TestFilterAndSortInterfaces(t *testing.T) {
	// Verify our filtering/sorting logic in isolation.
	allZeroMAC := net.HardwareAddr{0, 0, 0, 0, 0, 0}
	validMAC1 := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0x01}
	validMAC2 := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0x02}

	interfaces := []net.Interface{
		{Index: 5, Name: "eth1", Flags: net.FlagUp, HardwareAddr: validMAC2},
		{Index: 1, Name: "lo", Flags: net.FlagLoopback | net.FlagUp, HardwareAddr: nil},
		{Index: 3, Name: "zero0", Flags: net.FlagUp, HardwareAddr: allZeroMAC},
		{Index: 2, Name: "eth0", Flags: 0, HardwareAddr: validMAC1},
	}

	// Apply same filtering as systemUUID
	var candidates []net.Interface
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(iface.HardwareAddr) >= 6 && !bytes.Equal(iface.HardwareAddr, make([]byte, len(iface.HardwareAddr))) {
			candidates = append(candidates, iface)
		}
	}

	// Should exclude loopback (lo) and all-zero MAC (zero0)
	require.Len(t, candidates, 2)

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Index < candidates[j].Index
	})

	// Sorted by index: eth0 (2) before eth1 (5)
	assert.Equal(t, "eth0", candidates[0].Name)
	assert.Equal(t, "eth1", candidates[1].Name)

	// Prefer interface that is up
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

	// eth1 is up, eth0 is not — so eth1 should be preferred
	require.NotNil(t, best)
	assert.Equal(t, "eth1", best.Name)
	assert.Equal(t, validMAC2, best.HardwareAddr)
}

func TestFilterAndSortInterfaces_NoUpInterface(t *testing.T) {
	validMAC1 := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0x01}
	validMAC2 := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0x02}

	interfaces := []net.Interface{
		{Index: 10, Name: "eth1", Flags: 0, HardwareAddr: validMAC2},
		{Index: 3, Name: "eth0", Flags: 0, HardwareAddr: validMAC1},
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

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Index < candidates[j].Index
	})

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

	// No interface is up, so fallback to lowest index
	require.NotNil(t, best)
	assert.Equal(t, "eth0", best.Name)
}

func TestFilterAndSortInterfaces_AllFiltered(t *testing.T) {
	allZeroMAC := net.HardwareAddr{0, 0, 0, 0, 0, 0}

	interfaces := []net.Interface{
		{Index: 1, Name: "lo", Flags: net.FlagLoopback | net.FlagUp, HardwareAddr: nil},
		{Index: 2, Name: "zero0", Flags: net.FlagUp, HardwareAddr: allZeroMAC},
		{Index: 3, Name: "short", Flags: net.FlagUp, HardwareAddr: net.HardwareAddr{0x01, 0x02}},
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

	assert.Empty(t, candidates, "all interfaces should be filtered out")
}
