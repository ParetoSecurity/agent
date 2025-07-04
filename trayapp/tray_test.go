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
	// Mock dependencies
	mockStateManager := &MockStateManager{}
	mockMenuItem := NewMockMenuItem()
	mockBroadcaster := shared.NewBroadcaster()

	trayApp := NewTrayAppWithDependencies(
		nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
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
}

func TestTrayApp_updateClaim(t *testing.T) {
	t.Run("all checks pass", func(t *testing.T) {
		// Mock dependencies
		mockStateManager := &MockStateManager{}
		mockMenuItem := NewMockMenuItem()
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
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

	t.Run("no valid data - disabled state", func(t *testing.T) {
		// Mock dependencies
		mockStateManager := &MockStateManager{}
		mockMenuItem := NewMockMenuItem()
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
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

		// Test case: no valid data (not found)
		mockStateManager.On("GetLastState", "test-uuid").Return(shared.LastState{}, false, nil)
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
			nil, mockStateManager, nil, nil, nil, nil, nil, nil, nil, mockBroadcaster,
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

func TestTrayApp_watch_ErrorHandling(t *testing.T) {
	t.Run("watcher creation fails", func(t *testing.T) {
		// Mock dependencies
		mockFileWatcher := &MockFileWatcher{}
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, nil, nil, nil, mockFileWatcher, nil, nil, nil, nil, mockBroadcaster,
		)

		// Mock NewWatcher to return an error
		mockFileWatcher.On("NewWatcher").Return(nil, assert.AnError)

		// This should not panic even though NewWatcher returns an error
		trayApp.watch()

		// Give the goroutine a moment to execute
		// In a real scenario, we'd use more sophisticated synchronization
		// but for this test, a small sleep is sufficient
		time.Sleep(10 * time.Millisecond)

		mockFileWatcher.AssertExpectations(t)
	})

	t.Run("watcher creation succeeds", func(t *testing.T) {
		// Mock dependencies
		mockFileWatcher := &MockFileWatcher{}
		mockWatcher := &MockWatcher{}
		mockStateManager := &MockStateManager{}
		mockBroadcaster := shared.NewBroadcaster()

		trayApp := NewTrayAppWithDependencies(
			nil, mockStateManager, nil, nil, mockFileWatcher, nil, nil, nil, nil, mockBroadcaster,
		)

		// Mock successful watcher creation
		mockFileWatcher.On("NewWatcher").Return(mockWatcher, nil)
		mockStateManager.On("StatePath").Return("/test/path")
		mockWatcher.On("Add", "/test/path").Return(nil)
		mockWatcher.On("Close").Return(nil)

		// Create channels for events and errors
		eventCh := make(chan fsnotify.Event)
		errorCh := make(chan error)
		mockWatcher.events = eventCh
		mockWatcher.errors = errorCh

		// Start watching
		trayApp.watch()

		// Give the goroutine a moment to set up
		time.Sleep(10 * time.Millisecond)

		// Close the channels to simulate watcher shutdown
		close(eventCh)
		close(errorCh)

		// Give the goroutine a moment to clean up
		time.Sleep(10 * time.Millisecond)

		mockFileWatcher.AssertExpectations(t)
		mockStateManager.AssertExpectations(t)
		mockWatcher.AssertExpectations(t)
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
	assert.Equal(t, mockBroadcaster, trayApp.broadcaster)
}
