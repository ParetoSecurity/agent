package trayapp

import (
	"testing"
	"time"
)

func TestLastUpdated(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name             string
		modifiedTime     time.Time
		expected         string
		timeTravelOffset time.Duration
	}{
		{
			name:         "never updated",
			modifiedTime: time.Time{},
			expected:     "never",
		},
		{
			name:             "just now",
			modifiedTime:     now,
			timeTravelOffset: time.Duration(0),
			expected:         "just now",
		},
		{
			name:             "1 minute ago",
			modifiedTime:     now.Add(-time.Minute),
			timeTravelOffset: time.Duration(0),
			expected:         "1m ago",
		},
		{
			name:             "less than an hour",
			modifiedTime:     now.Add(-25 * time.Minute),
			timeTravelOffset: time.Duration(0),
			expected:         "25m ago",
		},
		{
			name:             "an hour ago",
			modifiedTime:     now.Add(-time.Hour),
			timeTravelOffset: time.Duration(0),
			expected:         "1h ago",
		},
		{
			name:             "less than a day",
			modifiedTime:     now.Add(-5 * time.Hour),
			timeTravelOffset: time.Duration(0),
			expected:         "5h ago",
		},
		{
			name:             "less than a day with minutes",
			modifiedTime:     now.Add(-5*time.Hour - 30*time.Minute),
			timeTravelOffset: time.Duration(0),
			expected:         "5h 30m ago",
		},
		{
			name:             "a day ago",
			modifiedTime:     now.Add(-24 * time.Hour),
			timeTravelOffset: time.Duration(0),
			expected:         "1d ago",
		},
		{
			name:             "less than a week",
			modifiedTime:     now.Add(-3 * 24 * time.Hour),
			timeTravelOffset: time.Duration(0),
			expected:         "3d ago",
		},
		{
			name:             "less than a week with hours",
			modifiedTime:     now.Add(-3*24*time.Hour - 12*time.Hour),
			timeTravelOffset: time.Duration(0),
			expected:         "3d 12h ago",
		},
		{
			name:             "a week ago",
			modifiedTime:     now.Add(-7 * 24 * time.Hour),
			timeTravelOffset: time.Duration(0),
			expected:         "1w ago",
		},
		{
			name:             "more than a week",
			modifiedTime:     now.Add(-9 * 24 * time.Hour),
			timeTravelOffset: time.Duration(0),
			expected:         "1w 2d ago",
		},
		{
			name:             "many weeks",
			modifiedTime:     now.Add(-30 * 24 * time.Hour),
			timeTravelOffset: time.Duration(0),
			expected:         "4w 2d ago",
		},
		{
			name:             "many weeks, no extra days",
			modifiedTime:     now.Add(-28 * 24 * time.Hour),
			timeTravelOffset: time.Duration(0),
			expected:         "4w ago",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			actual := lastUpdated()
			if actual != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}
