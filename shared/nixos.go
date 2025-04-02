package shared

import "sync"

var isNixOSOnce sync.Once
var isNixOS bool

// IsNixOS checks if the current system is NixOS by attempting to run the
// `nixos-version` command. It returns true if the command executes without
// error, indicating that NixOS is likely the operating system.
func IsNixOS() bool {
	isNixOSOnce.Do(func() {
		_, err := RunCommand("nixos-version")
		isNixOS = err == nil
	})
	return isNixOS
}
