package shared

import (
	"fmt"
	"testing"
)

func TestReadFile(t *testing.T) {
	// Mock data for testing

	ReadFileMock = func(name string) ([]byte, error) {
		if name == "testfile1.txt" {
			return []byte("This is a test file content"), nil
		}
		return nil, fmt.Errorf("ReadFile fixture not found: %s", name)
	}

	t.Run("ReadFile from mock", func(t *testing.T) {
		content, err := ReadFile("testfile1.txt")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := "This is a test file content"
		if string(content) != expected {
			t.Fatalf("expected %s, got %s", expected, string(content))
		}
	})

	t.Run("ReadFile mock not found", func(t *testing.T) {
		_, err := ReadFile("nonexistent.txt")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		expectedErr := "ReadFile fixture not found: nonexistent.txt"
		if err.Error() != expectedErr {
			t.Fatalf("expected %s, got %s", expectedErr, err.Error())
		}
	})

}
func TestUserHomeDir(t *testing.T) {
	t.Run("UserHomeDir returns mock path in tests", func(t *testing.T) {

		UserHomeDirMock = func() (string, error) {
			return "/home/user", nil
		}

		homeDir, err := UserHomeDir()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := "/home/user"
		if homeDir != expected {
			t.Fatalf("expected %s, got %s", expected, homeDir)
		}
	})
}
