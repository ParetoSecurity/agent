package shared

import (
	"os"
	"testing"
)

// ReadFileMocks is a map that simulates file reading operations by mapping
// file paths (as keys) to their corresponding file contents (as values).
// This can be used for testing purposes to mock the behavior of reading files
// without actually accessing the file system.
var ReadFileMock func(name string) ([]byte, error)

// ReadFile reads the content of the file specified by the given name.
// If the code is running in a testing environment, it will return the content
// from the ReadFileMocks map instead of reading from the actual file system.
// If the file name is not found in the ReadFileMocks map, it returns an error.
// Otherwise, it reads the file content from the file system.
func ReadFile(name string) ([]byte, error) {
	if testing.Testing() {
		return ReadFileMock(name)

	}
	return os.ReadFile(name)
}

var UserHomeDirMock func() (string, error)

// UserHomeDir returns the current user's home directory.
//
// On Unix, including macOS, it returns the $HOME environment variable.
// On Windows, it returns %USERPROFILE%.
// On Plan 9, it returns the $home environment variable.
//
// If the expected variable is not set in the environment, UserHomeDir
// returns either a platform-specific default value or a non-nil error.
func UserHomeDir() (string, error) {
	if testing.Testing() {
		// In tests, return a mock home directory
		return UserHomeDirMock()
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homeDir, nil
}
