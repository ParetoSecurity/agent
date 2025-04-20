package trayapp

import (
	"bytes"
	"fmt"
	"image/png"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"fyne.io/systray"
	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/claims"
	"github.com/ParetoSecurity/agent/notify"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/ParetoSecurity/agent/systemd"
	"github.com/caarlos0/log"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/browser"
)

func addQuitItem() {
	mQuit := systray.AddMenuItem("Quit", "Quit the Pareto Security")
	mQuit.Enable()
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
		os.Exit(0)
	}()
}

func checkStatusToIcon(status bool) string {
	if status {
		return "✅"
	}
	return "❌"
}

func getIcon() []byte {

	if runtime.GOOS == "windows" {

		var icoBuffer bytes.Buffer
		pngImage, err := png.Decode(bytes.NewReader(shared.IconBlack))
		if err != nil {
			log.WithError(err).Error("failed to decode PNG image")
			return shared.IconBlack
		}
		if err := ico.Encode(&icoBuffer, pngImage); err != nil {
			log.WithError(err).Error("failed to encode ICO image")
		}
		return icoBuffer.Bytes()
	}

	return shared.IconBlack
}

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
				err := exec.Command(shared.SelfExe(), "unlink").Run()
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

func OnReady() {
	broadcaster := shared.NewBroadcaster()
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			log.Info("Periodic update")
			broadcaster.Send()
		}
	}()
	if runtime.GOOS == "windows" {
		systray.SetTemplateIcon(shared.IconBlack, shared.IconBlack)
	}
	systray.SetTemplateIcon(getIcon(), getIcon())
	systray.SetTooltip("Pareto Security")
	systray.AddMenuItem(fmt.Sprintf("Pareto Security - %s", shared.Version), "").Disable()

	addOptions()
	systray.AddSeparator()
	rcheck := systray.AddMenuItem("Run Checks", "")
	go func(rcheck *systray.MenuItem) {
		for range rcheck.ClickedCh {
			log.Info("Running checks...")
			err := exec.Command(shared.SelfExe(), "check").Run()
			if err != nil {
				log.WithError(err).Error("failed to run check command")
			}
			log.Info("Checks completed")
			broadcaster.Send()
		}
	}(rcheck)
	lastUpdated := time.Since(shared.GetModifiedTime()).Round(time.Minute)
	lCheck := systray.AddMenuItem(fmt.Sprintf("Last check %s ago", lastUpdated), "")
	lCheck.Disable()
	go func() {
		for range broadcaster.Register() {
			lastUpdated := time.Since(shared.GetModifiedTime()).Round(time.Minute)
			lCheck.SetTitle(fmt.Sprintf("Last check %s ago", lastUpdated))
		}
	}()

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

	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.WithError(err).Error("Failed to create file watcher")
			return
		}
		defer watcher.Close()

		err = watcher.Add(shared.StatePath)
		if err != nil {
			log.WithError(err).WithField("path", shared.StatePath).Error("Failed to add state file to watcher")
			return
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Info("State file modified, updating...")
					broadcaster.Send()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.WithError(err).Error("File watcher error")
			}
		}
	}()

}

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

func updateClaim(claim claims.Claim, mClaim *systray.MenuItem) {
	for _, chk := range claim.Checks {
		checkStatus, found, _ := shared.GetLastState(chk.UUID())
		if found && !checkStatus.State && chk.IsRunnable() {
			mClaim.SetTitle(fmt.Sprintf("❌ %s", claim.Title))
			return
		}
	}

	mClaim.SetTitle(fmt.Sprintf("✅ %s", claim.Title))
}
