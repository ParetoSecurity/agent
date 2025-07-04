package trayapp

import (
	"time"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/fsnotify/fsnotify"
)

// CommandRunner interface for running external commands
type CommandRunner interface {
	RunCommand(cmd string, args ...string) (string, error)
}

// StateManager interface for managing application state
type StateManager interface {
	GetLastState(uuid string) (shared.LastState, bool, error)
	IsLinked() bool
	StatePath() string
	GetModifiedTime() time.Time
	SelfExe() string
}

// BrowserOpener interface for opening URLs in browser
type BrowserOpener interface {
	OpenURL(url string) error
}

// SystemTray interface for system tray operations
type SystemTray interface {
	SetTitle(title string)
	SetTemplateIcon(icon, tooltip []byte)
	AddMenuItem(title, tooltip string) MenuItem
	AddSeparator()
	Quit()
	TrayOpenedCh() <-chan struct{}
}

// MenuItem interface for system tray menu items
type MenuItem interface {
	Enable()
	Disable()
	SetTitle(title string)
	AddSubMenuItem(title, tooltip string) MenuItem
	AddSubMenuItemCheckbox(title, tooltip string, checked bool) MenuItem
	Check()
	Uncheck()
	ClickedCh() <-chan struct{}
}

// FileWatcher interface for file system watching
type FileWatcher interface {
	NewWatcher() (Watcher, error)
}

// Watcher interface for file system watcher
type Watcher interface {
	Add(path string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

// SystemdManager interface for systemd operations
type SystemdManager interface {
	IsTimerEnabled() bool
	EnableTimer() error
	DisableTimer() error
	IsTrayIconEnabled() bool
	EnableTrayIcon() error
	DisableTrayIcon() error
}

// Notifier interface for notifications
type Notifier interface {
	Toast(message string)
}

// ThemeSubscriber interface for theme changes
type ThemeSubscriber interface {
	SubscribeToThemeChanges(ch chan<- bool)
}

// IconProvider interface for icon management
type IconProvider interface {
	SetIcon()
	WorkingIcon()
	IconBlack() []byte
	IconWhite() []byte
}
