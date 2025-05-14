package trayapp

import (
	"fmt"
	"net/url"
	"runtime"
	"time"

	"fyne.io/systray"
	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/claims"
	"github.com/ParetoSecurity/agent/notify"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/ParetoSecurity/agent/systemd"
	"github.com/caarlos0/log"
	"github.com/pkg/browser"
)

// addQuitItem adds a "Quit" menu item to the system tray.
func addQuitItem() {
	mQuit := systray.AddMenuItem("Quit", "Quit the Pareto Security")
	mQuit.Enable()
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

// checkStatusToIcon converts a boolean status to an icon string.
func checkStatusToIcon(status bool) string {
	if status {
		return "✅"
	}
	return "❌"
}

// updateCheck updates the status of a specific check in the menu.
func updateCheck(chk check.Check, mCheck *systray.MenuItem) {
	if !chk.IsRunnable() {
		mCheck.Disable()
		mCheck.SetTitle(fmt.Sprintf("🚫 %s", chk.Name()))
		return
	}
	mCheck.Enable()
	checkStatus, found, _ := shared.GetLastState(chk.UUID())
	state := chk.Passed()
	if found {
		state = checkStatus.State
	}
	mCheck.SetTitle(fmt.Sprintf("%s %s", checkStatusToIcon(state), chk.Name()))
}

// updateClaim updates the status of a claim in the menu.
func updateClaim(claim claims.Claim, mClaim *systray.MenuItem) {
	mClaim.SetTitle(fmt.Sprintf("❌ %s", claim.Title))
	for _, chk := range claim.Checks {
		checkStatus, found, _ := shared.GetLastState(chk.UUID())
		if found && !checkStatus.State && chk.IsRunnable() {
			return
		}
	}
	mClaim.SetTitle(fmt.Sprintf("✅ %s", claim.Title))
}

// addOptions adds various options to the system tray menu.
func addOptions() {
	mOptions := systray.AddMenuItem("Options", "Settings")
	mlink := mOptions.AddSubMenuItemCheckbox("Send reports to the dashboard", "Configure sending device reports to the team", shared.IsLinked())
	go func() {
		for range mlink.ClickedCh {
			if !shared.IsLinked() {
				//open browser with help link
				if err := browser.OpenURL("https://paretosecurity.com/docs/" + runtime.GOOS + "/link"); err != nil {
					log.WithError(err).Error("failed to open help URL")
				}
			} else {
				// execute the command in the system terminal
				_, err := shared.RunCommand(shared.SelfExe(), "unlink")
				if err != nil {
					log.WithError(err).Error("failed to run unlink command")
				}
			}
			if shared.IsLinked() {
				mlink.Check()
			} else {
				mlink.Uncheck()
			}
		}
	}()
	if runtime.GOOS != "windows" {
		mrun := mOptions.AddSubMenuItemCheckbox("Run checks in the background", "Run checks periodically in the background while the user is logged in.", systemd.IsTimerEnabled())
		go func() {
			for range mrun.ClickedCh {
				if !systemd.IsTimerEnabled() {
					if err := systemd.EnableTimer(); err != nil {
						log.WithError(err).Error("failed to enable timer")
						notify.Blocking("Failed to enable timer, please check the logs for more information.")
					}

				} else {
					if err := systemd.DisableTimer(); err != nil {
						log.WithError(err).Error("failed to enable timer")
						notify.Blocking("Failed to enable timer, please check the logs for more information.")
					}
				}
				if systemd.IsTimerEnabled() {
					mrun.Check()
				} else {
					mrun.Uncheck()
				}
			}
		}()
		mshow := mOptions.AddSubMenuItemCheckbox("Run the tray icon at startup", "Show tray icon", systemd.IsTrayIconEnabled())
		go func() {
			for range mshow.ClickedCh {
				if !systemd.IsTrayIconEnabled() {
					if err := systemd.EnableTrayIcon(); err != nil {
						log.WithError(err).Error("failed to enable tray icon")
						notify.Blocking("Failed to enable tray icon, please check the logs for more information.")
					}

				} else {
					if err := systemd.DisableTrayIcon(); err != nil {
						log.WithError(err).Error("failed to disable tray icon")
						notify.Blocking("Failed to disable tray icon, please check the logs for more information.")
					}
				}
				if systemd.IsTrayIconEnabled() {
					mshow.Check()
				} else {
					mshow.Uncheck()
				}
			}
		}()
	}
}

// OnReady initializes the system tray and its menu items.
func OnReady() {
	systray.SetTitle("Pareto Security")
	log.Info("Starting Pareto Security tray application")

	// Clear any running state that might be left from a previous crash
	shared.StopAllRunningChecks()
	if err := shared.CommitLastState(); err != nil {
		log.WithError(err).Warn("failed to commit running state")
	}

	broadcaster := shared.NewBroadcaster()
	log.Info("Setting up system tray icon")
	setIcon()
	if runtime.GOOS == "windows" {
		themeCh := make(chan bool)
		go SubscribeToThemeChanges(themeCh)
		go func() {
			for isDark := range themeCh {
				icon := shared.IconBlack
				if isDark {
					icon = shared.IconWhite
				}
				systray.SetTemplateIcon(icon, icon)
			}
		}()
	}
	log.Info("Setting up system tray")
	systray.AddMenuItem(fmt.Sprintf("Pareto Security - %s", shared.Version), "").Disable()

	addOptions()
	systray.AddSeparator()
	rcheck := systray.AddMenuItem("Run Checks", "")
	go func(rcheck *systray.MenuItem) {
		for range rcheck.ClickedCh {
			if shared.AreChecksRunning() {
				log.Info("Checks are already running...")
				continue
			}
			rcheck.Disable()
			rcheck.SetTitle("Checking...")
			log.Info("Running checks...")
			_, err := shared.RunCommand(shared.SelfExe(), "check")
			if err != nil {
				log.WithError(err).Error("failed to run check command")
			}
			log.Info("Checks completed")
			rcheck.SetTitle("Run Checks")
			rcheck.Enable()
			broadcaster.Send()
		}
	}(rcheck)

	// Update run checks menu item if checks are running
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if shared.AreChecksRunning() {
				rcheck.Disable()
				rcheck.SetTitle("Checking...")
			} else {
				rcheck.SetTitle("Run Checks")
				rcheck.Enable()
			}
		}
	}()

	lCheck := systray.AddMenuItem(fmt.Sprintf("Last check: %s", lastUpdated()), "")
	lCheck.Disable()
	go func() {
		for range broadcaster.Register() {

			lCheck.SetTitle(fmt.Sprintf("Last check: %s", lastUpdated()))
		}
	}()
	go func() {
		for range systray.TrayOpenedCh {
			setIcon()
		}
	}()
	systray.AddSeparator()
	for _, claim := range claims.All {
		mClaim := systray.AddMenuItem(claim.Title, "")
		updateClaim(claim, mClaim)

		go func(mClaim *systray.MenuItem) {
			for range broadcaster.Register() {
				log.WithField("claim", claim.Title).Info("Updating claim status")
				updateClaim(claim, mClaim)
			}
		}(mClaim)

		for _, chk := range claim.Checks {
			mCheck := mClaim.AddSubMenuItem(chk.Name(), "")
			updateCheck(chk, mCheck)
			go func(chk check.Check, mCheck *systray.MenuItem) {
				for range broadcaster.Register() {
					log.WithField("check", chk.Name()).Info("Updating check status")
					updateCheck(chk, mCheck)
				}
			}(chk, mCheck)
			go func(chk check.Check, mCheck *systray.MenuItem) {
				for range mCheck.ClickedCh {
					log.WithField("check", chk.Name()).Info("Opening check URL")
					arch := "check-linux"
					if runtime.GOOS == "windows" {
						arch = "check-windows"
					}

					url := fmt.Sprintf("https://paretosecurity.com/%s/%s?details=%s", arch, chk.UUID(), url.QueryEscape(chk.Status()))

					if err := browser.OpenURL(url); err != nil {
						log.WithError(err).Error("failed to open check URL")
					}
				}
			}(chk, mCheck)
		}
	}
	systray.AddSeparator()
	addQuitItem()
	log.Info("System tray setup complete")
	// watch for changes in the state file
	go watch(broadcaster)
}
