package trayapp

import (
	"testing"
	"time"

	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/claims"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing

type MockCommandRunner struct {
	mock.Mock
}

func (m *MockCommandRunner) RunCommand(cmd string, args ...string) (string, error) {
	// Convert variadic args to interface{} slice for mock.Called
	callArgs := make([]interface{}, len(args)+1)
	callArgs[0] = cmd
	for i, arg := range args {
		callArgs[i+1] = arg
	}
	arguments := m.Called(callArgs...)
	return arguments.String(0), arguments.Error(1)
}

type MockStateManager struct {
	mock.Mock
}

func (m *MockStateManager) GetLastState(uuid string) (shared.LastState, bool, error) {
	arguments := m.Called(uuid)
	return arguments.Get(0).(shared.LastState), arguments.Bool(1), arguments.Error(2)
}

func (m *MockStateManager) IsLinked() bool {
	arguments := m.Called()
	return arguments.Bool(0)
}

func (m *MockStateManager) StatePath() string {
	arguments := m.Called()
	return arguments.String(0)
}

func (m *MockStateManager) GetModifiedTime() time.Time {
	arguments := m.Called()
	return arguments.Get(0).(time.Time)
}

func (m *MockStateManager) SelfExe() string {
	arguments := m.Called()
	return arguments.String(0)
}

type MockBrowserOpener struct {
	mock.Mock
}

func (m *MockBrowserOpener) OpenURL(url string) error {
	arguments := m.Called(url)
	return arguments.Error(0)
}

type MockSystemTray struct {
	mock.Mock
}

func (m *MockSystemTray) SetTitle(title string) {
	m.Called(title)
}

func (m *MockSystemTray) SetTemplateIcon(icon, tooltip []byte) {
	m.Called(icon, tooltip)
}

func (m *MockSystemTray) AddMenuItem(title, tooltip string) MenuItem {
	arguments := m.Called(title, tooltip)
	return arguments.Get(0).(MenuItem)
}

func (m *MockSystemTray) AddSeparator() {
	m.Called()
}

func (m *MockSystemTray) Quit() {
	m.Called()
}

func (m *MockSystemTray) TrayOpenedCh() <-chan struct{} {
	arguments := m.Called()
	return arguments.Get(0).(<-chan struct{})
}

type MockMenuItem struct {
	mock.Mock
	clickedCh chan struct{}
}

func NewMockMenuItem() *MockMenuItem {
	return &MockMenuItem{
		clickedCh: make(chan struct{}),
	}
}

func (m *MockMenuItem) Enable() {
	m.Called()
}

func (m *MockMenuItem) Disable() {
	m.Called()
}

func (m *MockMenuItem) SetTitle(title string) {
	m.Called(title)
}

func (m *MockMenuItem) AddSubMenuItem(title, tooltip string) MenuItem {
	arguments := m.Called(title, tooltip)
	return arguments.Get(0).(MenuItem)
}

func (m *MockMenuItem) AddSubMenuItemCheckbox(title, tooltip string, checked bool) MenuItem {
	arguments := m.Called(title, tooltip, checked)
	return arguments.Get(0).(MenuItem)
}

func (m *MockMenuItem) Check() {
	m.Called()
}

func (m *MockMenuItem) Uncheck() {
	m.Called()
}

func (m *MockMenuItem) ClickedCh() <-chan struct{} {
	return m.clickedCh
}

type MockFileWatcher struct {
	mock.Mock
}

func (m *MockFileWatcher) NewWatcher() (Watcher, error) {
	arguments := m.Called()
	if arguments.Get(0) == nil {
		return nil, arguments.Error(1)
	}
	return arguments.Get(0).(Watcher), arguments.Error(1)
}

type MockWatcher struct {
	mock.Mock
	events <-chan fsnotify.Event
	errors <-chan error
}

func (m *MockWatcher) Add(path string) error {
	arguments := m.Called(path)
	return arguments.Error(0)
}

func (m *MockWatcher) Close() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *MockWatcher) Events() <-chan fsnotify.Event {
	return m.events
}

func (m *MockWatcher) Errors() <-chan error {
	return m.errors
}

type MockSystemdManager struct {
	mock.Mock
}

func (m *MockSystemdManager) IsTimerEnabled() bool {
	arguments := m.Called()
	return arguments.Bool(0)
}

func (m *MockSystemdManager) EnableTimer() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *MockSystemdManager) DisableTimer() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *MockSystemdManager) IsTrayIconEnabled() bool {
	arguments := m.Called()
	return arguments.Bool(0)
}

func (m *MockSystemdManager) EnableTrayIcon() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *MockSystemdManager) DisableTrayIcon() error {
	arguments := m.Called()
	return arguments.Error(0)
}

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) Toast(message string) {
	m.Called(message)
}

type MockThemeSubscriber struct {
	mock.Mock
}

func (m *MockThemeSubscriber) SubscribeToThemeChanges(ch chan<- bool) {
	m.Called(ch)
}

type MockIconProvider struct {
	mock.Mock
}

func (m *MockIconProvider) SetIcon() {
	m.Called()
}

func (m *MockIconProvider) WorkingIcon() {
	m.Called()
}

func (m *MockIconProvider) IconBlack() []byte {
	arguments := m.Called()
	return arguments.Get(0).([]byte)
}

func (m *MockIconProvider) IconWhite() []byte {
	arguments := m.Called()
	return arguments.Get(0).([]byte)
}

type MockStartupManager struct {
	mock.Mock
}

func (m *MockStartupManager) IsStartupEnabled() bool {
	arguments := m.Called()
	return arguments.Bool(0)
}

func (m *MockStartupManager) EnableStartup() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *MockStartupManager) DisableStartup() error {
	arguments := m.Called()
	return arguments.Error(0)
}

// Test implementations

func TestTrayApp_checkStatusToIcon(t *testing.T) {
	trayApp := NewTrayApp()

	tests := []struct {
		name      string
		status    bool
		withError bool
		expected  string
	}{
		{"Error status", false, true, "⚠️"},
		{"Passed status", true, false, "✅"},
		{"Failed status", false, false, "❌"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trayApp.checkStatusToIcon(tt.status, tt.withError)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrayApp_updateCheck(t *testing.T) {
	t.Run("check is runnable and found", func(t *testing.T) {
		// Mock dependencies
		mockStateManager := &MockStateManager{}
		mockMenuItem := NewMockMenuItem()
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
		)

		// Mock check
		mockCheck := &MockCheck{}
		mockCheck.On("UUID").Return("test-uuid")
		mockCheck.On("Name").Return("Test Check")
		mockCheck.On("IsRunnable").Return(true).Maybe()

		// Test case: check is runnable and found
		checkState := shared.LastState{Passed: true, HasError: false}
		mockStateManager.On("GetLastState", "test-uuid").Return(checkState, true, nil)
		mockMenuItem.On("Enable").Return()
		mockMenuItem.On("SetTitle", "✅ Test Check").Return()

		trayApp.updateCheck(mockCheck, mockMenuItem)

		mockStateManager.AssertExpectations(t)
		mockMenuItem.AssertExpectations(t)
		mockCheck.AssertExpectations(t)
	})

	t.Run("check is not runnable", func(t *testing.T) {
		// Mock dependencies
		mockStateManager := &MockStateManager{}
		mockMenuItem := NewMockMenuItem()
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
		)

		// Mock check
		mockCheck := &MockCheck{}
		mockCheck.On("UUID").Return("test-uuid")
		mockCheck.On("Name").Return("Test Check")
		mockCheck.On("IsRunnable").Return(false)

		// Setup expectations
		mockStateManager.On("GetLastState", "test-uuid").Return(shared.LastState{}, false, nil).Maybe()
		mockMenuItem.On("Disable").Return()
		mockMenuItem.On("SetTitle", "🚫 Test Check").Return()

		trayApp.updateCheck(mockCheck, mockMenuItem)

		mockStateManager.AssertExpectations(t)
		mockMenuItem.AssertExpectations(t)
		mockCheck.AssertExpectations(t)
	})

	t.Run("check not found in state but runnable", func(t *testing.T) {
		// Mock dependencies
		mockStateManager := &MockStateManager{}
		mockMenuItem := NewMockMenuItem()
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
		)

		// Mock check
		mockCheck := &MockCheck{}
		mockCheck.On("UUID").Return("test-uuid")
		mockCheck.On("Name").Return("Test Check")
		mockCheck.On("IsRunnable").Return(true)

		// Test case: check not found but runnable - should be enabled
		mockStateManager.On("GetLastState", "test-uuid").Return(shared.LastState{}, false, nil)
		mockMenuItem.On("Enable").Return()
		mockMenuItem.On("SetTitle", "Test Check").Return()

		trayApp.updateCheck(mockCheck, mockMenuItem)

		mockStateManager.AssertExpectations(t)
		mockMenuItem.AssertExpectations(t)
		mockCheck.AssertExpectations(t)
	})

	t.Run("check with error", func(t *testing.T) {
		// Mock dependencies
		mockStateManager := &MockStateManager{}
		mockMenuItem := NewMockMenuItem()
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
		)

		// Mock check
		mockCheck := &MockCheck{}
		mockCheck.On("UUID").Return("test-uuid")
		mockCheck.On("Name").Return("Test Check")
		mockCheck.On("IsRunnable").Return(true)

		// Test case: check has error
		checkState := shared.LastState{Passed: false, HasError: true}
		mockStateManager.On("GetLastState", "test-uuid").Return(checkState, true, nil)
		mockMenuItem.On("Enable").Return()
		mockMenuItem.On("SetTitle", "⚠️ Test Check").Return()

		trayApp.updateCheck(mockCheck, mockMenuItem)

		mockStateManager.AssertExpectations(t)
		mockMenuItem.AssertExpectations(t)
		mockCheck.AssertExpectations(t)
	})
}

func TestTrayApp_updateClaim(t *testing.T) {
	t.Run("all checks pass", func(t *testing.T) {
		// Mock dependencies
		mockStateManager := &MockStateManager{}
		mockMenuItem := NewMockMenuItem()
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
		)

		// Mock check
		mockCheck := &MockCheck{}
		mockCheck.On("UUID").Return("test-uuid")
		mockCheck.On("IsRunnable").Return(true).Maybe()

		// Mock claim
		claim := claims.Claim{
			Title:  "Test Claim",
			Checks: []check.Check{mockCheck},
		}

		// Test case: all checks pass
		checkState := shared.LastState{Passed: true, HasError: false}
		mockStateManager.On("GetLastState", "test-uuid").Return(checkState, true, nil)
		mockMenuItem.On("Enable").Return().Once()
		mockMenuItem.On("SetTitle", "✅ Test Claim").Return().Once()

		trayApp.updateClaim(claim, mockMenuItem)

		mockStateManager.AssertExpectations(t)
		mockMenuItem.AssertExpectations(t)
		mockCheck.AssertExpectations(t)
	})

	t.Run("no valid data but has runnable checks - enabled state", func(t *testing.T) {
		// Mock dependencies
		mockStateManager := &MockStateManager{}
		mockMenuItem := NewMockMenuItem()
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
		)

		// Mock check
		mockCheck := &MockCheck{}
		mockCheck.On("UUID").Return("test-uuid")
		mockCheck.On("IsRunnable").Return(true).Maybe()

		// Mock claim
		claim := claims.Claim{
			Title:  "Test Claim",
			Checks: []check.Check{mockCheck},
		}

		// Test case: no valid data (not found) but has runnable checks
		mockStateManager.On("GetLastState", "test-uuid").Return(shared.LastState{}, false, nil)
		mockMenuItem.On("Enable").Return().Once()
		mockMenuItem.On("SetTitle", "Test Claim").Return().Once()

		trayApp.updateClaim(claim, mockMenuItem)

		mockStateManager.AssertExpectations(t)
		mockMenuItem.AssertExpectations(t)
		mockCheck.AssertExpectations(t)
	})

	t.Run("no runnable checks - disabled state", func(t *testing.T) {
		// Mock dependencies
		mockStateManager := &MockStateManager{}
		mockMenuItem := NewMockMenuItem()
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
		)

		// Mock check
		mockCheck := &MockCheck{}
		mockCheck.On("UUID").Return("test-uuid")
		mockCheck.On("IsRunnable").Return(false).Maybe()

		// Mock claim
		claim := claims.Claim{
			Title:  "Test Claim",
			Checks: []check.Check{mockCheck},
		}

		// Test case: no runnable checks
		mockStateManager.On("GetLastState", "test-uuid").Return(shared.LastState{}, false, nil).Maybe()
		mockMenuItem.On("Disable").Return().Once()
		mockMenuItem.On("SetTitle", "Test Claim").Return().Once()

		trayApp.updateClaim(claim, mockMenuItem)

		mockStateManager.AssertExpectations(t)
		mockMenuItem.AssertExpectations(t)
		mockCheck.AssertExpectations(t)
	})

	t.Run("check with failure", func(t *testing.T) {
		// Mock dependencies
		mockStateManager := &MockStateManager{}
		mockMenuItem := NewMockMenuItem()
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
		)

		// Mock check
		mockCheck := &MockCheck{}
		mockCheck.On("UUID").Return("test-uuid")
		mockCheck.On("IsRunnable").Return(true).Maybe()

		// Mock claim
		claim := claims.Claim{
			Title:  "Test Claim",
			Checks: []check.Check{mockCheck},
		}

		// Test case: check fails
		checkState := shared.LastState{Passed: false, HasError: false}
		mockStateManager.On("GetLastState", "test-uuid").Return(checkState, true, nil)
		mockMenuItem.On("Enable").Return().Once()
		mockMenuItem.On("SetTitle", "❌ Test Claim").Return().Once()

		trayApp.updateClaim(claim, mockMenuItem)

		mockStateManager.AssertExpectations(t)
		mockMenuItem.AssertExpectations(t)
		mockCheck.AssertExpectations(t)
	})
}

func TestMockCommandRunner_RunCommand(t *testing.T) {
	t.Run("single argument", func(t *testing.T) {
		mockCommandRunner := &MockCommandRunner{}

		// Test with single argument (like the actual usage in tray app)
		mockCommandRunner.On("RunCommand", "/path/to/paretosecurity", "check").Return("check output", nil)

		result, err := mockCommandRunner.RunCommand("/path/to/paretosecurity", "check")

		assert.NoError(t, err)
		assert.Equal(t, "check output", result)
		mockCommandRunner.AssertExpectations(t)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		mockCommandRunner := &MockCommandRunner{}

		// Test with multiple arguments
		mockCommandRunner.On("RunCommand", "/path/to/exe", "arg1", "arg2", "arg3").Return("multi output", nil)

		result, err := mockCommandRunner.RunCommand("/path/to/exe", "arg1", "arg2", "arg3")

		assert.NoError(t, err)
		assert.Equal(t, "multi output", result)
		mockCommandRunner.AssertExpectations(t)
	})

	t.Run("no arguments", func(t *testing.T) {
		mockCommandRunner := &MockCommandRunner{}

		// Test with no arguments
		mockCommandRunner.On("RunCommand", "/path/to/exe").Return("no args output", nil)

		result, err := mockCommandRunner.RunCommand("/path/to/exe")

		assert.NoError(t, err)
		assert.Equal(t, "no args output", result)
		mockCommandRunner.AssertExpectations(t)
	})
}

// MockCheck for testing
type MockCheck struct {
	mock.Mock
}

func (m *MockCheck) Name() string {
	arguments := m.Called()
	return arguments.String(0)
}

func (m *MockCheck) PassedMessage() string {
	arguments := m.Called()
	return arguments.String(0)
}

func (m *MockCheck) FailedMessage() string {
	arguments := m.Called()
	return arguments.String(0)
}

func (m *MockCheck) Run() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *MockCheck) Passed() bool {
	arguments := m.Called()
	return arguments.Bool(0)
}

func (m *MockCheck) IsRunnable() bool {
	arguments := m.Called()
	return arguments.Bool(0)
}

func (m *MockCheck) UUID() string {
	arguments := m.Called()
	return arguments.String(0)
}

func (m *MockCheck) Status() string {
	arguments := m.Called()
	return arguments.String(0)
}

func (m *MockCheck) RequiresRoot() bool {
	arguments := m.Called()
	return arguments.Bool(0)
}

func TestNewTrayApp(t *testing.T) {
	trayApp := NewTrayApp()
	assert.NotNil(t, trayApp)
	assert.NotNil(t, trayApp.commandRunner)
	assert.NotNil(t, trayApp.stateManager)
	assert.NotNil(t, trayApp.browserOpener)
	assert.NotNil(t, trayApp.systemTray)
	assert.NotNil(t, trayApp.fileWatcher)
	assert.NotNil(t, trayApp.systemdManager)
	assert.NotNil(t, trayApp.notifier)
	assert.NotNil(t, trayApp.themeSubscriber)
	assert.NotNil(t, trayApp.iconProvider)
	assert.NotNil(t, trayApp.broadcaster)
}

func TestNewTrayAppWithDependencies(t *testing.T) {
	mockCommandRunner := &MockCommandRunner{}
	mockStateManager := &MockStateManager{}
	mockBrowserOpener := &MockBrowserOpener{}
	mockSystemTray := &MockSystemTray{}
	mockFileWatcher := &MockFileWatcher{}
	mockSystemdManager := &MockSystemdManager{}
	mockNotifier := &MockNotifier{}
	mockThemeSubscriber := &MockThemeSubscriber{}
	mockIconProvider := &MockIconProvider{}
	mockStartupManager := &MockStartupManager{}
	mockBroadcaster := shared.NewBroadcaster()

	trayApp := NewTrayAppWithDependencies(
		mockCommandRunner,
		mockStateManager,
		mockBrowserOpener,
		mockSystemTray,
		mockFileWatcher,
		mockSystemdManager,
		mockNotifier,
		mockThemeSubscriber,
		mockIconProvider,
		mockStartupManager,
		mockBroadcaster,
	)

	assert.NotNil(t, trayApp)
	assert.Equal(t, mockCommandRunner, trayApp.commandRunner)
	assert.Equal(t, mockStateManager, trayApp.stateManager)
	assert.Equal(t, mockBrowserOpener, trayApp.browserOpener)
	assert.Equal(t, mockSystemTray, trayApp.systemTray)
	assert.Equal(t, mockFileWatcher, trayApp.fileWatcher)
	assert.Equal(t, mockSystemdManager, trayApp.systemdManager)
	assert.Equal(t, mockNotifier, trayApp.notifier)
	assert.Equal(t, mockThemeSubscriber, trayApp.themeSubscriber)
	assert.Equal(t, mockIconProvider, trayApp.iconProvider)
	assert.Equal(t, mockStartupManager, trayApp.startupManager)
	assert.Equal(t, mockBroadcaster, trayApp.broadcaster)
}

func TestTrayApp_addQuitItem(t *testing.T) {
	// Mock dependencies
	mockSystemTray := &MockSystemTray{}
	mockMenuItem := NewMockMenuItem()
	mockBroadcaster := shared.NewBroadcaster()

	trayApp := NewTrayAppWithDependencies(
		nil, nil, nil, mockSystemTray, nil, nil, nil, nil, nil, nil, mockBroadcaster,
	)

	// Setup expectations
	mockSystemTray.On("AddMenuItem", "Quit", "Quit the Pareto Security").Return(mockMenuItem)
	mockMenuItem.On("Enable").Return()
	mockSystemTray.On("Quit").Return().Maybe()

	// Call the method
	trayApp.addQuitItem()

	// Verify expectations
	mockSystemTray.AssertExpectations(t)
	mockMenuItem.AssertExpectations(t)

	// Test clicking quit
	go func() {
		mockMenuItem.clickedCh <- struct{}{}
	}()

	// Give goroutine time to process
	time.Sleep(10 * time.Millisecond)
}

func TestTrayApp_lastUpdated(t *testing.T) {
	// TrayApp uses the standalone lastUpdated() function which
	// doesn't use stateManager - it calls a global function
	trayApp := NewTrayApp()

	result := trayApp.lastUpdated()

	// The result should be a formatted string
	assert.NotEmpty(t, result)
}

func TestTrayApp_IconMethods(t *testing.T) {
	// Test icon provider methods through tray app
	mockIconProvider := &MockIconProvider{}
	mockBroadcaster := shared.NewBroadcaster()

	trayApp := NewTrayAppWithDependencies(
		nil, nil, nil, nil, nil, nil, nil, nil, mockIconProvider, nil, mockBroadcaster,
	)

	// Test working icon
	mockIconProvider.On("WorkingIcon").Return().Once()
	trayApp.iconProvider.WorkingIcon()

	// Test set icon
	mockIconProvider.On("SetIcon").Return().Once()
	trayApp.iconProvider.SetIcon()

	// Test icon colors
	blackIcon := []byte{0x1, 0x2, 0x3}
	whiteIcon := []byte{0x4, 0x5, 0x6}
	mockIconProvider.On("IconBlack").Return(blackIcon).Once()
	mockIconProvider.On("IconWhite").Return(whiteIcon).Once()

	resultBlack := trayApp.iconProvider.IconBlack()
	resultWhite := trayApp.iconProvider.IconWhite()

	assert.Equal(t, blackIcon, resultBlack)
	assert.Equal(t, whiteIcon, resultWhite)

	mockIconProvider.AssertExpectations(t)
}

func TestTrayApp_ImplementationMethods(t *testing.T) {
	// Test implementation wrapper methods
	mockCommandRunner := &MockCommandRunner{}
	mockStateManager := &MockStateManager{}
	mockBrowserOpener := &MockBrowserOpener{}
	mockBroadcaster := shared.NewBroadcaster()

	trayApp := NewTrayAppWithDependencies(
		mockCommandRunner, mockStateManager, mockBrowserOpener, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
	)

	// Test command runner
	mockCommandRunner.On("RunCommand", "test-cmd", "arg1").Return("output", nil).Once()
	result, err := trayApp.commandRunner.RunCommand("test-cmd", "arg1")
	assert.NoError(t, err)
	assert.Equal(t, "output", result)

	// Test state manager methods
	testTime := time.Now()
	mockStateManager.On("IsLinked").Return(true).Once()
	mockStateManager.On("StatePath").Return("/test/path").Once()
	mockStateManager.On("GetModifiedTime").Return(testTime).Once()
	mockStateManager.On("SelfExe").Return("/test/exe").Once()

	assert.True(t, trayApp.stateManager.IsLinked())
	assert.Equal(t, "/test/path", trayApp.stateManager.StatePath())
	assert.Equal(t, testTime, trayApp.stateManager.GetModifiedTime())
	assert.Equal(t, "/test/exe", trayApp.stateManager.SelfExe())

	// Test browser opener
	mockBrowserOpener.On("OpenURL", "https://example.com").Return(nil).Once()
	err = trayApp.browserOpener.OpenURL("https://example.com")
	assert.NoError(t, err)

	mockCommandRunner.AssertExpectations(t)
	mockStateManager.AssertExpectations(t)
	mockBrowserOpener.AssertExpectations(t)
}

func TestTrayApp_StartupAndSystemdMethods(t *testing.T) {
	// Test startup and systemd manager methods
	mockStartupManager := &MockStartupManager{}
	mockSystemdManager := &MockSystemdManager{}
	mockBroadcaster := shared.NewBroadcaster()

	trayApp := NewTrayAppWithDependencies(
		nil, nil, nil, nil, nil, mockSystemdManager, nil, nil, nil, mockStartupManager, mockBroadcaster,
	)

	// Test startup manager
	mockStartupManager.On("IsStartupEnabled").Return(true).Once()
	mockStartupManager.On("EnableStartup").Return(nil).Once()
	mockStartupManager.On("DisableStartup").Return(nil).Once()

	assert.True(t, trayApp.startupManager.IsStartupEnabled())
	assert.NoError(t, trayApp.startupManager.EnableStartup())
	assert.NoError(t, trayApp.startupManager.DisableStartup())

	// Test systemd manager
	mockSystemdManager.On("IsTimerEnabled").Return(false).Once()
	mockSystemdManager.On("EnableTimer").Return(nil).Once()
	mockSystemdManager.On("DisableTimer").Return(nil).Once()

	assert.False(t, trayApp.systemdManager.IsTimerEnabled())
	assert.NoError(t, trayApp.systemdManager.EnableTimer())
	assert.NoError(t, trayApp.systemdManager.DisableTimer())

	mockStartupManager.AssertExpectations(t)
	mockSystemdManager.AssertExpectations(t)
}

func TestTrayApp_NotifierAndThemeMethods(t *testing.T) {
	// Test notifier and theme subscriber methods
	mockNotifier := &MockNotifier{}
	mockThemeSubscriber := &MockThemeSubscriber{}
	mockBroadcaster := shared.NewBroadcaster()

	trayApp := NewTrayAppWithDependencies(
		nil, nil, nil, nil, nil, nil, mockNotifier, mockThemeSubscriber, nil, nil, mockBroadcaster,
	)

	// Test notifier
	mockNotifier.On("Toast", "Test message").Return().Once()
	trayApp.notifier.Toast("Test message")

	// Test theme subscriber
	themeCh := make(chan bool, 1)
	mockThemeSubscriber.On("SubscribeToThemeChanges", (chan<- bool)(themeCh)).Return().Once()
	trayApp.themeSubscriber.SubscribeToThemeChanges(themeCh)

	mockNotifier.AssertExpectations(t)
	mockThemeSubscriber.AssertExpectations(t)
}

func TestTrayApp_FileWatcherMethods(t *testing.T) {
	// Test file watcher methods
	mockFileWatcher := &MockFileWatcher{}
	mockWatcher := &MockWatcher{}
	mockBroadcaster := shared.NewBroadcaster()

	trayApp := NewTrayAppWithDependencies(
		nil, nil, nil, nil, mockFileWatcher, nil, nil, nil, nil, nil, mockBroadcaster,
	)

	// Test successful watcher creation
	mockFileWatcher.On("NewWatcher").Return(mockWatcher, nil).Once()
	watcher, err := trayApp.fileWatcher.NewWatcher()
	assert.NoError(t, err)
	assert.Equal(t, mockWatcher, watcher)

	// Test watcher methods
	mockWatcher.On("Add", "/test/path").Return(nil).Once()
	mockWatcher.On("Close").Return(nil).Once()

	err = watcher.Add("/test/path")
	assert.NoError(t, err)

	err = watcher.Close()
	assert.NoError(t, err)

	mockFileWatcher.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
}

func TestTrayApp_WatchWithEvents(t *testing.T) {
	t.Run("watch handles non-write events", func(t *testing.T) {
		// Mock dependencies
		mockFileWatcher := &MockFileWatcher{}
		mockWatcher := &MockWatcher{}
		mockStateManager := &MockStateManager{}
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, mockFileWatcher, nil, nil, nil, nil, nil, mockBroadcaster,
		)

		// Mock successful watcher creation
		mockFileWatcher.On("NewWatcher").Return(mockWatcher, nil)
		mockStateManager.On("StatePath").Return("/test/path")
		mockWatcher.On("Add", "/test/path").Return(nil)
		mockWatcher.On("Close").Return(nil).Maybe()

		// Create channels for events and errors
		eventCh := make(chan fsnotify.Event, 2)
		errorCh := make(chan error)
		mockWatcher.events = eventCh
		mockWatcher.errors = errorCh

		// Subscribe to broadcaster
		receiveCh := mockBroadcaster.Register()

		// Start watching
		trayApp.watch()

		// Give the goroutine a moment to set up
		time.Sleep(10 * time.Millisecond)

		// Send a non-write event (should be ignored)
		eventCh <- fsnotify.Event{Op: fsnotify.Create}

		// Send a write event (should trigger broadcast)
		eventCh <- fsnotify.Event{Op: fsnotify.Write}

		// Should only receive one broadcast (from write event)
		select {
		case <-receiveCh:
			// Success - received broadcast from write event
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Did not receive expected broadcast")
		}

		// Should not receive another broadcast
		select {
		case <-receiveCh:
			t.Fatal("Received unexpected second broadcast")
		case <-time.After(50 * time.Millisecond):
			// Expected - no second broadcast
		}

		// Clean up
		close(eventCh)
		close(errorCh)
		time.Sleep(10 * time.Millisecond)

		mockFileWatcher.AssertExpectations(t)
		mockStateManager.AssertExpectations(t)
	})
}
