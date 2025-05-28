package notify

import (
	"fmt"
	"os/exec"
)

// Toast displays a system notification on macOS using AppleScript.
func Toast(message string) {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`display notification "%s"`, message))
	cmd.Run()
}
