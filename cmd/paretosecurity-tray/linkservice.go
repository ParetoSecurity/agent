//go:build windows
// +build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ParetoSecurity/agent/notify"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/ParetoSecurity/agent/team"
	"github.com/caarlos0/log"
	"github.com/samber/lo"
)

// LinkService provides the Wails service for device linking
type LinkService struct{}

// LinkDevice links this device to a team using the invite ID
func (l *LinkService) LinkDevice(inviteID string, host string) error {
	if shared.IsRoot() {
		return fmt.Errorf("Cannot link device: Please run as a normal user, not as administrator")
	}

	if lo.IsEmpty(inviteID) {
		log.Warn("No invite ID provided")
		return errors.New("Invalid link: No invite ID found in the URL")
	}

	// Check if already linked
	if shared.IsLinked() {
		log.Info("Device already linked to a team, unlinking first")
		log.Infof("Previous Team ID: %s", shared.Config.TeamID)
		// Automatically unlink the device
		shared.Config.TeamID = ""
		shared.Config.AuthToken = ""
		shared.Config.TeamAPI = ""
		log.Info("Device unlinked, proceeding with new team linking")
	}

	// Convert empty string host to empty for team.EnrollDevice
	if host == "" {
		host = ""
	}

	// Enroll the device
	err := team.EnrollDevice(inviteID, host)
	if err != nil {
		log.WithError(err).Warn("failed to enroll device")
		// Provide more context based on the error
		errMsg := err.Error()

		// Check for common error scenarios and provide helpful messages
		if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "no such host") {
			return fmt.Errorf("Connection failed: Unable to reach the Pareto Security server. Please check your internet connection and try again")
		}
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "unauthorized") {
			return fmt.Errorf("Authorization failed: The invitation link may have expired or been used already. Please request a new invitation link")
		}
		if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "not found") {
			return fmt.Errorf("Invalid invitation: The invitation link is invalid or has been removed. Please request a new invitation link")
		}
		if strings.Contains(errMsg, "timeout") {
			return fmt.Errorf("Connection timeout: The server took too long to respond. Please check your internet connection and try again")
		}
		if strings.Contains(errMsg, "certificate") || strings.Contains(errMsg, "x509") {
			return fmt.Errorf("Security error: Certificate verification failed. Please check your system time and date settings")
		}

		// Default error with more context
		return fmt.Errorf("Enrollment failed: %s. Please contact your administrator if the problem persists", errMsg)
	}

	// Save config
	err = shared.SaveConfig()
	if err != nil {
		log.Errorf("Error saving config: %v", err)
		return fmt.Errorf("Configuration error: Failed to save settings. Please check disk permissions and available space")
	}

	// Report to team
	if shared.IsLinked() {
		err := team.ReportToTeam(false)
		if err != nil {
			log.WithError(err).Warn("failed to report to team")
		}
		log.Infof("Device successfully linked to team: %s", shared.Config.TeamID)
		notify.Toast("Device successfully linked to the team!")

		// Restart the tray application to reflect the new linked status
		go func() {
			// Get the path to the tray executable
			exePath, err := os.Executable()
			if err != nil {
				log.WithError(err).Error("failed to get executable path")
				return
			}

			// Get the lock file path
			lockDir, _ := shared.UserHomeDir()
			lockPath := filepath.Join(lockDir, ".paretosecurity-tray.lock")

			// Try to read PID from lock file
			if data, err := os.ReadFile(lockPath); err == nil {
				if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
					log.WithField("pid", pid).Info("Found existing tray application PID from lock file")

					// Use Windows API to terminate the process
					if err := terminateProcess(pid); err != nil {
						log.WithError(err).Debug("Failed to terminate process, may already be stopped")
					} else {
						log.WithField("pid", pid).Info("Successfully stopped existing tray application")
						// Wait a moment for the process to fully terminate
						time.Sleep(500 * time.Millisecond)
					}

					// Remove the lock file so the new instance can start
					os.Remove(lockPath)
				} else {
					log.WithError(err).Warn("Failed to parse PID from lock file")
				}
			} else {
				log.WithError(err).Debug("No lock file found, tray app probably not running")
			}

			// Now start the tray app without arguments (normal tray mode)
			log.Info("Starting tray application after successful linking")

			cmd := exec.Command(exePath)
			cmd.Dir = filepath.Dir(exePath)

			// Start the process detached from the current one
			err = cmd.Start()
			if err != nil {
				log.WithError(err).Error("failed to start tray application")
				return
			}

			log.Info("Tray application started successfully")

			// Don't wait for it, let it run independently
			cmd.Process.Release()
		}()
	}

	return nil
}

// QuitApp closes the application
func (l *LinkService) QuitApp() error {
	os.Exit(0)
	return nil
}

// terminateProcess uses Windows API to terminate a process by PID
func terminateProcess(pid int) error {
	const PROCESS_TERMINATE = 0x0001

	// Open the process with terminate permissions
	handle, err := syscall.OpenProcess(PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return fmt.Errorf("failed to open process: %w", err)
	}
	defer syscall.CloseHandle(handle)

	// Terminate the process
	err = syscall.TerminateProcess(handle, 0)
	if err != nil {
		return fmt.Errorf("failed to terminate process: %w", err)
	}

	return nil
}
