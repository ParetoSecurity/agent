package shared

import (
	"testing"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello, 世界!", "Hello,_!"},
		{"123 ABC abc", "123 ABC abc"},
		{"Special chars: @#$%^&*()", "Special chars: _"},
		{"Mixed: 你好, 世界! 123", "Mixed: __,_! 123"},
		{"Punctuation: .,!-_'\"", "Punctuation: .,!-_'\""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := Sanitize(test.input)
			if result != test.expected {
				t.Errorf("Sanitize(%q) = %q; want %q", test.input, result, test.expected)
			}
		})
	}
}
