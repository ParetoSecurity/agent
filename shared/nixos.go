package shared

import (
	"log"
	"os"
	"sync"
	"testing"
)

var isNixOSOnce sync.Once
var isNixOS bool

// IsNixOS checks if the current system is NixOS by attempting to run the
// `nixos-version` command. It returns true if the command executes without
// error, indicating that NixOS is likely the operating system.
func IsNixOS() bool {
	if testing.Testing() {
		return false
	}
	isNixOSOnce.Do(func() {

		_, err := os.Stat("/run/current-system/sw")
		isNixOS = err == nil
		log.Println("Checking if system is NixOS", isNixOS)
	})
	return isNixOS
}
