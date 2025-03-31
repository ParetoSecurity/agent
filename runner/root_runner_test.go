package runner

import (
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/claims"
	"github.com/stretchr/testify/assert"
)

type MockCheck struct {
	UUIDValue         string
	RequiresRootValue bool
	RunError          error
	PassedValue       bool
	StatusValue       string
}

func (m *MockCheck) UUID() string          { return m.UUIDValue }
func (m *MockCheck) RequiresRoot() bool    { return m.RequiresRootValue }
func (m *MockCheck) Run() error            { return m.RunError }
func (m *MockCheck) Passed() bool          { return m.PassedValue }
func (m *MockCheck) Status() string        { return m.StatusValue }
func (m *MockCheck) FailedMessage() string { return "Mock check failed" }
func (m *MockCheck) PassedMessage() string { return "Mock check passed" }
func (m *MockCheck) IsRunnable() bool      { return true }
func (m *MockCheck) Name() string          { return "MockCheck" }

func TestHandleConnection(t *testing.T) {
	// Setup
	uuid := "test-uuid"
	input := map[string]string{"uuid": uuid}
	inputJSON, _ := json.Marshal(input)

	// Mock connection
	conn := &mockConn{
		readData: string(inputJSON),
	}

	// Mock check
	mockCheck := &MockCheck{
		UUIDValue:         uuid,
		RequiresRootValue: true,
		RunError:          nil,
		PassedValue:       true,
		StatusValue:       "Check passed",
	}

	// Set claims.All to use the mock check
	claims.All = []claims.Claim{
		{
			Checks: []check.Check{mockCheck},
		},
	}

	// Execute
	HandleConnection(conn)

	// Assertions
	assert.True(t, conn.closed, "Connection should be closed")

	var response CheckStatus
	err := json.Unmarshal([]byte(conn.writtenData), &response)
	assert.NoError(t, err, "Should not error when unmarshaling response")

	assert.Equal(t, uuid, response.UUID, "UUID should match")
	assert.True(t, response.Passed, "Check should pass")
	assert.Equal(t, "Check passed", response.Details, "Details should match")
}

func TestRunCheckViaRoot(t *testing.T) {
	// Setup

	expectedStatus := &CheckStatus{
		UUID:    "",
		Passed:  false,
		Details: "",
	}

	// Execute
	status, _ := RunCheckViaRoot("test-uuid")
	assert.Equal(t, expectedStatus, status, "Status should match")

}

// mockConn is a mock implementation of the net.Conn interface.
type mockConn struct {
	readData    string
	writtenData string
	closed      bool
}

// Read mocks the Read method of the net.Conn interface.
func (m *mockConn) Read(b []byte) (n int, err error) {
	r := strings.NewReader(m.readData)
	return r.Read(b)
}

// Write mocks the Write method of the net.Conn interface.
func (m *mockConn) Write(b []byte) (n int, err error) {
	m.writtenData = string(b)
	return len(b), nil
}

// Close mocks the Close method of the net.Conn interface.
func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

// LocalAddr mocks the LocalAddr method of the net.Conn interface.
func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
}

// RemoteAddr mocks the RemoteAddr method of the net.Conn interface.
func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}
}

// SetDeadline mocks the SetDeadline method of the net.Conn interface.
func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline mocks the SetReadDeadline method of the net.Conn interface.
func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline mocks the SetWriteDeadline method of the net.Conn interface.
func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}
