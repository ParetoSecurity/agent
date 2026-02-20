//go:build darwin

package trayapp

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/caarlos0/log"
)

func (t *TrayApp) openConsole() {
	selfExe := t.stateManager.SelfExe()
	escapedSelfExe := strings.ReplaceAll(selfExe, `"`, `\"`)
	script := fmt.Sprintf(`tell application "Terminal" to do script quoted form of "%s" & " check"`, escapedSelfExe)
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Start(); err != nil {
		log.WithError(err).Error("failed to open Terminal console")
		t.notifier.Toast("Failed to open Terminal console.")
		return
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.WithError(err).Error("osascript terminated with error")
		}
	}()
}
