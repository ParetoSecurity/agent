//go:build windows

package trayapp

import (
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/caarlos0/log"
)

const createNewConsole = 0x00000010

func (t *TrayApp) openConsole() {
	selfExe := t.stateManager.SelfExe()
	dir := filepath.Dir(selfExe)
	exe := filepath.Base(selfExe)
	psCommand := "& ./" + exe + " check"

	var cmd *exec.Cmd
	if wtPath, err := exec.LookPath("wt.exe"); err == nil {
		// Call Windows Terminal directly so arguments are not discarded by
		// the default-terminal interception mechanism.
		cmd = exec.Command(wtPath, "--startingDirectory", dir,
			"powershell.exe", "-NoExit", "-Command", psCommand)
	} else {
		cmd = exec.Command("powershell.exe", "-NoExit", "-Command", psCommand)
		cmd.Dir = dir
		cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNewConsole}
	}

	if err := cmd.Start(); err != nil {
		log.WithError(err).Error("failed to open PowerShell console")
		t.notifier.Toast("Failed to open PowerShell console.")
		return
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.WithError(err).Debug("PowerShell console exited with error")
		}
	}()
}
