package trayapp

import (
	"time"

	"fyne.io/systray"
	"github.com/ParetoSecurity/agent/notify"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/ParetoSecurity/agent/systemd"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/browser"
)

// Production implementations

type RealCommandRunner struct{}

func (r *RealCommandRunner) RunCommand(cmd string, args ...string) (string, error) {
	return shared.RunCommand(cmd, args...)
}

type RealStateManager struct{}

func (r *RealStateManager) GetLastState(uuid string) (shared.LastState, bool, error) {
	return shared.GetLastState(uuid)
}

func (r *RealStateManager) IsLinked() bool {
	return shared.IsLinked()
}

func (r *RealStateManager) StatePath() string {
	return shared.StatePath
}

func (r *RealStateManager) GetModifiedTime() time.Time {
	return shared.GetModifiedTime()
}

func (r *RealStateManager) SelfExe() string {
	return shared.SelfExe()
}

type RealBrowserOpener struct{}

func (r *RealBrowserOpener) OpenURL(url string) error {
	return browser.OpenURL(url)
}

type RealSystemTray struct{}

func (r *RealSystemTray) SetTitle(title string) {
	systray.SetTitle(title)
}

func (r *RealSystemTray) SetTemplateIcon(icon, tooltip []byte) {
	systray.SetTemplateIcon(icon, tooltip)
}

func (r *RealSystemTray) AddMenuItem(title, tooltip string) MenuItem {
	return &RealMenuItem{item: systray.AddMenuItem(title, tooltip)}
}

func (r *RealSystemTray) AddSeparator() {
	systray.AddSeparator()
}

func (r *RealSystemTray) Quit() {
	systray.Quit()
}

func (r *RealSystemTray) TrayOpenedCh() <-chan struct{} {
	return systray.TrayOpenedCh
}

type RealMenuItem struct {
	item *systray.MenuItem
}

func (r *RealMenuItem) Enable() {
	r.item.Enable()
}

func (r *RealMenuItem) Disable() {
	r.item.Disable()
}

func (r *RealMenuItem) SetTitle(title string) {
	r.item.SetTitle(title)
}

func (r *RealMenuItem) AddSubMenuItem(title, tooltip string) MenuItem {
	return &RealMenuItem{item: r.item.AddSubMenuItem(title, tooltip)}
}

func (r *RealMenuItem) AddSubMenuItemCheckbox(title, tooltip string, checked bool) MenuItem {
	return &RealMenuItem{item: r.item.AddSubMenuItemCheckbox(title, tooltip, checked)}
}

func (r *RealMenuItem) Check() {
	r.item.Check()
}

func (r *RealMenuItem) Uncheck() {
	r.item.Uncheck()
}

func (r *RealMenuItem) ClickedCh() <-chan struct{} {
	return r.item.ClickedCh
}

type RealFileWatcher struct{}

func (r *RealFileWatcher) NewWatcher() (Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &RealWatcher{watcher: w}, nil
}

type RealWatcher struct {
	watcher *fsnotify.Watcher
}

func (r *RealWatcher) Add(path string) error {
	return r.watcher.Add(path)
}

func (r *RealWatcher) Close() error {
	return r.watcher.Close()
}

func (r *RealWatcher) Events() <-chan fsnotify.Event {
	return r.watcher.Events
}

func (r *RealWatcher) Errors() <-chan error {
	return r.watcher.Errors
}

type RealSystemdManager struct{}

func (r *RealSystemdManager) IsTimerEnabled() bool {
	return systemd.IsTimerEnabled()
}

func (r *RealSystemdManager) EnableTimer() error {
	return systemd.EnableTimer()
}

func (r *RealSystemdManager) DisableTimer() error {
	return systemd.DisableTimer()
}

func (r *RealSystemdManager) IsTrayIconEnabled() bool {
	return systemd.IsTrayIconEnabled()
}

func (r *RealSystemdManager) EnableTrayIcon() error {
	return systemd.EnableTrayIcon()
}

func (r *RealSystemdManager) DisableTrayIcon() error {
	return systemd.DisableTrayIcon()
}

type RealNotifier struct{}

func (r *RealNotifier) Toast(message string) {
	notify.Toast(message)
}

type RealThemeSubscriber struct{}

func (r *RealThemeSubscriber) SubscribeToThemeChanges(ch chan<- bool) {
	SubscribeToThemeChanges(ch)
}

type RealIconProvider struct{}

func (r *RealIconProvider) SetIcon() {
	setIcon()
}

func (r *RealIconProvider) WorkingIcon() {
	workingIcon()
}

func (r *RealIconProvider) IconBlack() []byte {
	return shared.IconBlack
}

func (r *RealIconProvider) IconWhite() []byte {
	return shared.IconWhite
}
