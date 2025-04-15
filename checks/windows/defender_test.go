//go:build windows
// +build windows

package checks

import (
	"testing"
)

func TestDefenderEnabledCheck(t *testing.T) {

	cases := []struct {
		name string
		mock map[string]map[string]uint64
		want bool
	}{
		{"Defender enabled (no disables)", map[string]map[string]uint64{}, true},
		{"Defender disabled", map[string]map[string]uint64{"SOFTWARE\\Policies\\Microsoft\\Windows Defender": {"DisableAntiSpyware": 1}}, false},
		{"Realtime off", map[string]map[string]uint64{"SOFTWARE\\Policies\\Microsoft\\Windows Defender\\Real-TimeProtection": {"DisableRealtimeMonitoring": 1}}, false},
		{"Both disables", map[string]map[string]uint64{"SOFTWARE\\Policies\\Microsoft\\Windows Defender": {"DisableAntiSpyware": 1}, "SOFTWARE\\Policies\\Microsoft\\Windows Defender\\Real-TimeProtection": {"DisableRealtimeMonitoring": 1}}, false},
		{"Explicitly enabled", map[string]map[string]uint64{"SOFTWARE\\Policies\\Microsoft\\Windows Defender": {"DisableAntiSpyware": 0}, "SOFTWARE\\Policies\\Microsoft\\Windows Defender\\Real-TimeProtection": {"DisableRealtimeMonitoring": 0}}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			chk := &DefenderEnabledCheck{}
			chk.Run()
			if chk.Passed() != c.want {
				t.Errorf("got %v, want %v", chk.Passed(), c.want)
			}
		})
	}
}
