package main

import (
	"testing"
	"time"
)

func TestParseTimestampISO8601(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		timezone  string
		wantErr   bool
		desc      string
	}{
		{
			name:      "ISO 8601 with Z timezone",
			timestamp: "2025-07-19T08:45:40.501Z",
			timezone:  "America/New_York",
			wantErr:   false,
			desc:      "Should parse RFC3339 format with Z",
		},
		{
			name:      "ISO 8601 with timezone offset",
			timestamp: "2025-07-19T08:45:40-05:00",
			timezone:  "America/New_York",
			wantErr:   false,
			desc:      "Should parse RFC3339 with timezone offset",
		},
		{
			name:      "ISO 8601 without timezone",
			timestamp: "2025-07-19T08:45:40",
			timezone:  "America/New_York",
			wantErr:   false,
			desc:      "Should parse ISO format without timezone",
		},
		{
			name:      "ISO 8601 with milliseconds",
			timestamp: "2025-07-19T08:45:40.501",
			timezone:  "America/New_York",
			wantErr:   false,
			desc:      "Should parse ISO format with milliseconds",
		},
		{
			name:      "ISO 8601 date only",
			timestamp: "2025-07-19",
			timezone:  "America/New_York",
			wantErr:   false,
			desc:      "Should parse date-only format",
		},
		{
			name:      "Legacy format space separated",
			timestamp: "2025-07-19 08:45:40",
			timezone:  "America/New_York",
			wantErr:   false,
			desc:      "Should still support legacy format",
		},
		{
			name:      "Invalid format",
			timestamp: "not-a-date",
			timezone:  "America/New_York",
			wantErr:   true,
			desc:      "Should error on invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimestamp(tt.timestamp, tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: parseTimestamp() error = %v, wantErr %v", tt.desc, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.IsZero() {
				t.Errorf("%s: parseTimestamp() returned zero time", tt.desc)
			}
		})
	}
}

func TestParseTimestampTimezoneConversion(t *testing.T) {
	// Test that timezone conversion works correctly
	timestamp := "2025-07-19T12:00:00Z"  // Noon UTC
	
	// Parse to New York time
	nyTime, err := parseTimestamp(timestamp, "America/New_York")
	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}
	
	// In July, New York is UTC-4 (EDT)
	// So noon UTC should be 8 AM in New York
	if nyTime.Hour() != 8 {
		t.Errorf("Expected hour to be 8 (EDT), got %d", nyTime.Hour())
	}
	
	// Parse to Tokyo time
	tokyoTime, err := parseTimestamp(timestamp, "Asia/Tokyo")
	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}
	
	// Tokyo is UTC+9
	// So noon UTC should be 9 PM in Tokyo
	if tokyoTime.Hour() != 21 {
		t.Errorf("Expected hour to be 21 (JST), got %d", tokyoTime.Hour())
	}
}

func TestTimeSinceWithISO8601(t *testing.T) {
	// Create a timestamp 1 hour ago in ISO 8601 format
	oneHourAgo := time.Now().Add(-time.Hour).Format(time.RFC3339)
	
	// This would be called by the MCP handler
	parsedTime, err := parseTimestamp(oneHourAgo, "UTC")
	if err != nil {
		t.Fatalf("Failed to parse ISO timestamp: %v", err)
	}
	
	// Calculate time since
	since := time.Since(parsedTime)
	
	// Should be approximately 1 hour
	if since < 59*time.Minute || since > 61*time.Minute {
		t.Errorf("Expected time since to be ~1 hour, got %v", since)
	}
}