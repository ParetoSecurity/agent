package tui

import (
	"bytes"
	"fmt"
	"sync"
	"time"
)

// logWriter is a custom writer that captures log messages
type logWriter struct {
	mu       sync.Mutex
	buffer   *[]string
	maxLines int
	enabled  bool
}

// newLogWriter creates a new log writer
func newLogWriter(buffer *[]string, maxLines int) *logWriter {
	return &logWriter{
		buffer:   buffer,
		maxLines: maxLines,
		enabled:  true,
	}
}

// Write implements io.Writer interface
func (w *logWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.enabled || w.buffer == nil {
		return len(p), nil
	}

	// Split the input into lines
	lines := bytes.Split(p, []byte("\n"))
	timestamp := time.Now().Format("15:04:05")

	for _, line := range lines {
		if len(line) > 0 {
			// Add timestamp and line to buffer
			logLine := fmt.Sprintf("[%s] %s", timestamp, string(line))
			*w.buffer = append(*w.buffer, logLine)

			// Keep buffer size limited
			if len(*w.buffer) > w.maxLines {
				*w.buffer = (*w.buffer)[len(*w.buffer)-w.maxLines:]
			}
		}
	}

	return len(p), nil
}

// Clear clears the log buffer
func (w *logWriter) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buffer != nil {
		*w.buffer = []string{}
	}
}

// SetEnabled enables or disables log capture
func (w *logWriter) SetEnabled(enabled bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.enabled = enabled
}
