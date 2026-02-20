//go:build linux

package trayapp

import (
	"os/exec"

	"github.com/caarlos0/log"
)

func (t *TrayApp) openConsole() {
	selfExe := t.stateManager.SelfExe()

	type terminal struct {
		name string
		args []string
	}

	terminals := []terminal{
		{"x-terminal-emulator", []string{"-e", selfExe, "check"}},
		{"gnome-terminal", []string{"--", selfExe, "check"}},
		{"konsole", []string{"-e", selfExe, "check"}},
		{"xfce4-terminal", []string{"--execute", selfExe, "check"}},
		{"xterm", []string{"-e", selfExe, "check"}},
	}

	for _, term := range terminals {
		if _, err := exec.LookPath(term.name); err == nil {
			cmd := exec.Command(term.name, term.args...)
			if err := cmd.Start(); err != nil {
				log.WithError(err).Error("failed to open terminal console")
				continue
			}
			go func() {
				if err := cmd.Wait(); err != nil {
					log.WithError(err).Error("terminal console exited with error")
				}
			}()
			return
		}
	}

	log.Warn("no supported terminal emulator found")
	t.notifier.Toast("Failed to open console: no supported terminal emulator found.")
}
