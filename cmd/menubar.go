package cmd

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"time"

	"os/exec"

	"fyne.io/systray"
	"github.com/ParetoSecurity/pareto-linux/check"
	"github.com/ParetoSecurity/pareto-linux/claims"
	"github.com/ParetoSecurity/pareto-linux/shared"
	"github.com/caarlos0/log"
	"github.com/pkg/browser"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
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

	// isDark, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output()
	// if err == nil && strings.Contains(string(isDark), "prefer-dark") {
	// 	return shared.IconWhite
	// }
	// isKDE, err := exec.Command("kreadconfig5", "--group", "General", "--key", "ColorScheme").Output()
	// if err == nil && strings.Contains(string(isKDE), "Dark") {
	// 	return shared.IconWhite
	// }

	return shared.IconWhite
}

func addOptions() {
	mOptions := systray.AddMenuItem("Options", "Settings")
	mauto := mOptions.AddSubMenuItemCheckbox("Run checks every hour", "Toggle running checks every hour", isUserTimerInstalled())
	mlink := mOptions.AddSubMenuItemCheckbox("Send reports to the dashboard", "Configure sending device reports to the team", shared.IsLinked())
	go func() {
		for range mauto.ClickedCh {
			if isUserTimerInstalled() {
				// execute the command to toggle auto start
				err := exec.Command(shared.SelfExe(), "check", "--install").Run()
				if err != nil {
					log.WithError(err).Error("failed to run toggle-autostart command")
				}
			} else {
				// execute the command to toggle auto start
				err := exec.Command(shared.SelfExe(), "check", "--uninstall").Run()
				if err != nil {
					log.WithError(err).Error("failed to run toggle-autostart command")
				}
			}
			if isUserTimerInstalled() {
				mauto.Check()
			} else {
				mauto.Uncheck()
			}
		}
	}()
	go func() {
		for range mlink.ClickedCh {
			if shared.IsLinked() {
				// execute the command in the system terminal
				err := exec.Command(shared.SelfExe(), "link").Run()
				if err != nil {
					log.WithError(err).Error("failed to run link command")
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
}

func onReady() {
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
	systray.AddMenuItem("Pareto Security", "").Disable()

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
	allStatus := lo.Reduce(claim.Checks, func(acc bool, item check.Check, index int) bool {
		checkStatus, found, _ := shared.GetLastState(item.UUID())
		if !item.IsRunnable() {
			return acc && true
		}
		return acc && checkStatus.State && found
	}, true)

	mClaim.SetTitle(fmt.Sprintf("%s %s", checkStatusToIcon(allStatus), claim.Title))
}

var menubarCmd = &cobra.Command{
	Use:   "menubar",
	Short: "Show the checks in the menubar",
	Run: func(cc *cobra.Command, args []string) {
		onExit := func() {
			log.Info("Exiting...")
		}

		systray.Run(onReady, onExit)
	},
}

func init() {
	rootCmd.AddCommand(menubarCmd)
}
