package trayapp

import (
	"fmt"
	"net/url"
	"runtime"

	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/claims"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/fsnotify/fsnotify"
)

// TrayApp represents the system tray application with testable dependencies
type TrayApp struct {
	commandRunner   CommandRunner
	stateManager    StateManager
	browserOpener   BrowserOpener
	systemTray      SystemTray
	fileWatcher     FileWatcher
	systemdManager  SystemdManager
	notifier        Notifier
	themeSubscriber ThemeSubscriber
	iconProvider    IconProvider
	broadcaster     *shared.Broadcaster
}

// NewTrayApp creates a new TrayApp with production dependencies
func NewTrayApp() *TrayApp {
	return &TrayApp{
		commandRunner:   &RealCommandRunner{},
		stateManager:    &RealStateManager{},
		browserOpener:   &RealBrowserOpener{},
		systemTray:      &RealSystemTray{},
		fileWatcher:     &RealFileWatcher{},
		systemdManager:  &RealSystemdManager{},
		notifier:        &RealNotifier{},
		themeSubscriber: &RealThemeSubscriber{},
		iconProvider:    &RealIconProvider{},
		broadcaster:     shared.NewBroadcaster(),
	}
}

// NewTrayAppWithDependencies creates a new TrayApp with custom dependencies for testing
func NewTrayAppWithDependencies(
	commandRunner CommandRunner,
	stateManager StateManager,
	browserOpener BrowserOpener,
	systemTray SystemTray,
	fileWatcher FileWatcher,
	systemdManager SystemdManager,
	notifier Notifier,
	themeSubscriber ThemeSubscriber,
	iconProvider IconProvider,
	broadcaster *shared.Broadcaster,
) *TrayApp {
	return &TrayApp{
		commandRunner:   commandRunner,
		stateManager:    stateManager,
		browserOpener:   browserOpener,
		systemTray:      systemTray,
		fileWatcher:     fileWatcher,
		systemdManager:  systemdManager,
		notifier:        notifier,
		themeSubscriber: themeSubscriber,
		iconProvider:    iconProvider,
		broadcaster:     broadcaster,
	}
}

// addQuitItem adds a "Quit" menu item to the system tray
func (t *TrayApp) addQuitItem() {
	mQuit := t.systemTray.AddMenuItem("Quit", "Quit the Pareto Security")
	mQuit.Enable()
	go func() {
		<-mQuit.ClickedCh()
		t.systemTray.Quit()
	}()
}

// checkStatusToIcon converts a boolean status to an icon string
func (t *TrayApp) checkStatusToIcon(status, withError bool) string {
	if withError {
		return "⚠️"
	}
	if status {
		return "✅"
	}
	return "❌"
}

// updateCheck updates the status of a specific check in the menu
func (t *TrayApp) updateCheck(chk check.Check, mCheck MenuItem) {
	checkStatus, found, _ := t.stateManager.GetLastState(chk.UUID())
	if !chk.IsRunnable() || !found {
		mCheck.Disable()
		mCheck.SetTitle(fmt.Sprintf("🚫 %s", chk.Name()))
		return
	}
	if found {
		mCheck.Enable()
		mCheck.SetTitle(fmt.Sprintf("%s %s", t.checkStatusToIcon(checkStatus.Passed, checkStatus.HasError), chk.Name()))
	}
}

// updateClaim updates the status of a claim in the menu
func (t *TrayApp) updateClaim(claim claims.Claim, mClaim MenuItem) {
	hasValidData := false

	for _, chk := range claim.Checks {
		checkStatus, found, _ := t.stateManager.GetLastState(chk.UUID())
		if found && chk.IsRunnable() {
			hasValidData = true
			if !checkStatus.Passed {
				mClaim.Enable()
				mClaim.SetTitle(fmt.Sprintf("❌ %s", claim.Title))
				return
			}
		}
	}

	if hasValidData {
		mClaim.Enable()
		mClaim.SetTitle(fmt.Sprintf("✅ %s", claim.Title))
	} else {
		mClaim.SetTitle(claim.Title)
	}
}

// addOptions adds various options to the system tray menu
func (t *TrayApp) addOptions() {
	mOptions := t.systemTray.AddMenuItem("Options", "Settings")
	mlink := mOptions.AddSubMenuItemCheckbox("Send reports to the dashboard", "Configure sending device reports to the team", t.stateManager.IsLinked())
	go func() {
		for range mlink.ClickedCh() {
			if !t.stateManager.IsLinked() {
				// open browser with help link
				if err := t.browserOpener.OpenURL("https://paretosecurity.com/docs/" + runtime.GOOS + "/link"); err != nil {
					log.WithError(err).Error("failed to open help URL")
				}
			} else {
				// execute the command in the system terminal
				_, err := t.commandRunner.RunCommand(t.stateManager.SelfExe(), "unlink")
				if err != nil {
					log.WithError(err).Error("failed to run unlink command")
				}
			}
			if t.stateManager.IsLinked() {
				mlink.Check()
			} else {
				mlink.Uncheck()
			}
		}
	}()
	if runtime.GOOS != "windows" {
		mrun := mOptions.AddSubMenuItemCheckbox("Run checks in the background", "Run checks periodically in the background while the user is logged in.", t.systemdManager.IsTimerEnabled())
		go func() {
			for range mrun.ClickedCh() {
				if !t.systemdManager.IsTimerEnabled() {
					if err := t.systemdManager.EnableTimer(); err != nil {
						log.WithError(err).Error("failed to enable timer")
						t.notifier.Toast("Failed to enable timer, please check the logs for more information.")
					}
				} else {
					if err := t.systemdManager.DisableTimer(); err != nil {
						log.WithError(err).Error("failed to disable timer")
						t.notifier.Toast("Failed to disable timer, please check the logs for more information.")
					}
				}
				if t.systemdManager.IsTimerEnabled() {
					mrun.Check()
				} else {
					mrun.Uncheck()
				}
			}
		}()
		mshow := mOptions.AddSubMenuItemCheckbox("Run the tray icon at startup", "Show tray icon", t.systemdManager.IsTrayIconEnabled())
		go func() {
			for range mshow.ClickedCh() {
				if !t.systemdManager.IsTrayIconEnabled() {
					if err := t.systemdManager.EnableTrayIcon(); err != nil {
						log.WithError(err).Error("failed to enable tray icon")
						t.notifier.Toast("Failed to enable tray icon, please check the logs for more information.")
					}
				} else {
					if err := t.systemdManager.DisableTrayIcon(); err != nil {
						log.WithError(err).Error("failed to disable tray icon")
						t.notifier.Toast("Failed to disable tray icon, please check the logs for more information.")
					}
				}
				if t.systemdManager.IsTrayIconEnabled() {
					mshow.Check()
				} else {
					mshow.Uncheck()
				}
			}
		}()
	}
}

// lastUpdated returns the last updated time as a formatted string
func (t *TrayApp) lastUpdated() string {
	return lastUpdated()
}

// OnReady initializes the system tray and its menu items
func (t *TrayApp) OnReady() {
	t.systemTray.SetTitle("Pareto Security")
	log.Info("Starting Pareto Security tray application")

	log.Info("Setting up system tray icon")
	t.iconProvider.SetIcon()
	if runtime.GOOS == "windows" {
		themeCh := make(chan bool)
		go t.themeSubscriber.SubscribeToThemeChanges(themeCh)
		go func() {
			for isDark := range themeCh {
				icon := t.iconProvider.IconBlack()
				if isDark {
					icon = t.iconProvider.IconWhite()
				}
				t.systemTray.SetTemplateIcon(icon, icon)
			}
		}()
	}
	log.Info("Setting up system tray")
	t.systemTray.AddMenuItem(fmt.Sprintf("Pareto Security - %s", shared.Version), "").Disable()

	t.addOptions()
	t.systemTray.AddSeparator()
	rcheck := t.systemTray.AddMenuItem("Run Checks", "")

	lCheck := t.systemTray.AddMenuItem(fmt.Sprintf("Last check: %s", t.lastUpdated()), "")
	lCheck.Disable()
	go func() {
		for range t.broadcaster.Register() {
			lCheck.SetTitle(fmt.Sprintf("Last check: %s", t.lastUpdated()))
		}
	}()
	go func() {
		for range t.systemTray.TrayOpenedCh() {
			t.iconProvider.SetIcon()
		}
	}()
	// Store claim menu items for disabling during checks
	var claimMenuItems []MenuItem

	t.systemTray.AddSeparator()
	for _, claim := range claims.All {
		mClaim := t.systemTray.AddMenuItem(claim.Title, "")
		mClaim.Disable()
		claimMenuItems = append(claimMenuItems, mClaim)
		t.updateClaim(claim, mClaim)

		go func(mClaim MenuItem) {
			for range t.broadcaster.Register() {
				log.WithField("claim", claim.Title).Info("Updating claim status")
				t.updateClaim(claim, mClaim)
			}
		}(mClaim)

		for _, chk := range claim.Checks {
			mCheck := mClaim.AddSubMenuItem(chk.Name(), "")
			t.updateCheck(chk, mCheck)
			go func(chk check.Check, mCheck MenuItem) {
				for range t.broadcaster.Register() {
					log.WithField("check", chk.Name()).Info("Updating check status")
					t.updateCheck(chk, mCheck)
				}
			}(chk, mCheck)
			go func(chk check.Check, mCheck MenuItem) {
				for range mCheck.ClickedCh() {
					log.WithField("check", chk.Name()).Info("Opening check URL")
					arch := "check-linux"
					if runtime.GOOS == "windows" {
						arch = "check-windows"
					}

					checkStatus, found, _ := t.stateManager.GetLastState(chk.UUID())

					var targetURL string
					if found && checkStatus.HasError {
						targetURL = "https://paretosecurity.com/docs/linux/check-error"
					} else {
						targetURL = fmt.Sprintf("https://paretosecurity.com/%s/%s?details=%s", arch, chk.UUID(), url.QueryEscape(chk.Status()))
					}

					if err := t.browserOpener.OpenURL(targetURL); err != nil {
						log.WithError(err).Error("failed to open check URL")
					}
				}
			}(chk, mCheck)
		}
	}

	// Set up "Run Checks" functionality after claims are created
	go func(rcheck MenuItem) {
		for range rcheck.ClickedCh() {
			rcheck.Disable()
			rcheck.SetTitle("Checking...")
			log.Info("Running checks...")
			t.iconProvider.WorkingIcon()

			// Disable all claim menu items and remove icons during check execution
			for i, claimMenuItem := range claimMenuItems {
				claimMenuItem.Disable()
				claimMenuItem.SetTitle(claims.All[i].Title) // Set plain title without icon
			}

			_, err := t.commandRunner.RunCommand(t.stateManager.SelfExe(), "check")
			if err != nil {
				log.WithError(err).Error("failed to run check command")
				t.iconProvider.SetIcon()
			}
			log.Info("Checks completed")
			rcheck.SetTitle("Run Checks")
			rcheck.Enable()
			t.iconProvider.SetIcon()
			t.broadcaster.Send()
		}
	}(rcheck)
	t.systemTray.AddSeparator()
	t.addQuitItem()
	log.Info("System tray setup complete")
	// watch for changes in the state file
	go t.watch()
}

// watch monitors the state file for changes
func (t *TrayApp) watch() {
	go func() {
		watcher, err := t.fileWatcher.NewWatcher()
		if err != nil {
			log.WithError(err).Error("Failed to create file watcher")
			return
		}
		defer watcher.Close()

		err = watcher.Add(t.stateManager.StatePath())
		if err != nil {
			log.WithError(err).WithField("path", t.stateManager.StatePath()).Error("Failed to add state file to watcher")
			return
		}

		for {
			select {
			case event, ok := <-watcher.Events():
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Info("State file modified, updating...")
					t.broadcaster.Send()
				}
			case err, ok := <-watcher.Errors():
				if !ok {
					return
				}
				log.WithError(err).Error("File watcher error")
			}
		}
	}()
}
