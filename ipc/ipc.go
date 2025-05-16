package shared

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// StateMessage represents a state update message
type StateMessage struct {
	Running time.Time `json:"running"`
}

// IPCServer represents an IPC server
type IPCServer struct {
	listener    net.Listener
	path        string
	subscribers []chan StateMessage
	state       StateMessage
	mu          sync.RWMutex
}

// NewIPCServer creates a new IPC server
func NewIPCServer() (*IPCServer, error) {
	socketPath := getSocketPath()

	// Remove existing socket if it exists
	os.Remove(socketPath)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(socketPath), 0755); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	return &IPCServer{
		listener:    listener,
		path:        socketPath,
		subscribers: make([]chan StateMessage, 0),
	}, nil
}

// Subscribe returns a channel that receives state updates
func (s *IPCServer) Subscribe() chan StateMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan StateMessage, 5) // Buffer to prevent blocking
	s.subscribers = append(s.subscribers, ch)

	// Send current state immediately to new subscriber
	ch <- s.state

	return ch
}

// Unsubscribe removes a subscription channel
func (s *IPCServer) Unsubscribe(ch chan StateMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, subCh := range s.subscribers {
		if subCh == ch {
			// Remove subscriber by replacing with last element and truncating
			s.subscribers[i] = s.subscribers[len(s.subscribers)-1]
			s.subscribers = s.subscribers[:len(s.subscribers)-1]
			close(ch)
			break
		}
	}
}

// updateState updates the internal state and notifies all subscribers
func (s *IPCServer) updateState(msg StateMessage) {
	s.mu.Lock()
	s.state = msg
	subscribers := make([]chan StateMessage, len(s.subscribers))
	copy(subscribers, s.subscribers)
	s.mu.Unlock()

	// Notify subscribers without holding the lock
	for _, ch := range subscribers {
		select {
		case ch <- msg:
			// Message sent successfully
		default:
			// Channel is full, skip this update for this subscriber
		}
	}
}

// Start begins listening for connections
func (s *IPCServer) Start() {
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				continue
			}

			// Handle each connection in a goroutine
			go s.handleConnection(conn)
		}
	}()
}

// handleConnection processes incoming connections
func (s *IPCServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	var msg StateMessage
	if err := decoder.Decode(&msg); err != nil {
		return
	}

	// Update state and notify subscribers
	s.updateState(msg)
}

// Close shuts down the server
func (s *IPCServer) Close() error {
	err := s.listener.Close()
	os.Remove(s.path)
	return err
}

// UpdateRunningState sends a state update to the IPC server
func UpdateRunningState(t time.Time) error {
	socketPath := getSocketPath()
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	msg := StateMessage{
		Running: t,
	}

	return json.NewEncoder(conn).Encode(msg)
}

// GetRunningState requests the current running state from the server
func GetRunningState() (time.Time, error) {
	// This would be implemented based on how you store the state
	// For a complete implementation, the server would need to
	// maintain state and respond to query requests
	return time.Time{}, nil
}

// GetCurrentState returns the current state
func (s *IPCServer) GetCurrentState() StateMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// getSocketPath returns the appropriate socket path for the current platform
func getSocketPath() string {
	var socketPath string

	switch runtime.GOOS {
	case "windows":
		// Windows requires a different approach for Unix sockets
		// Use the Windows pipe naming convention
		socketPath = filepath.Join("\\\\.\\pipe\\", "pareto-agent-ipc")
	case "darwin":
		// macOS typically uses /var/run or /tmp
		socketPath = filepath.Join("/tmp", "pareto-agent-ipc.sock")
	default:
		// Linux and other Unix-like systems
		if xdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR"); xdgRuntimeDir != "" {
			socketPath = filepath.Join(xdgRuntimeDir, "pareto-agent-ipc.sock")
		} else {
			socketPath = filepath.Join("/tmp", "pareto-agent-ipc.sock")
		}
	}

	return socketPath
}
