package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/justcfx2u/passage-of-time-mcp-go/passageoftime"
)

// Test parseTimestamp function
func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		timezone  string
		wantErr   bool
		validate  func(time.Time) bool
	}{
		{
			name:      "full timestamp format",
			timestamp: "2024-01-15 14:30:00",
			timezone:  "UTC",
			wantErr:   false,
			validate: func(tm time.Time) bool {
				return tm.Year() == 2024 && tm.Month() == 1 && tm.Day() == 15 &&
					tm.Hour() == 14 && tm.Minute() == 30 && tm.Second() == 0
			},
		},
		{
			name:      "date only format",
			timestamp: "2024-01-15",
			timezone:  "UTC",
			wantErr:   false,
			validate: func(tm time.Time) bool {
				return tm.Year() == 2024 && tm.Month() == 1 && tm.Day() == 15 &&
					tm.Hour() == 0 && tm.Minute() == 0 && tm.Second() == 0
			},
		},
		{
			name:      "timestamp with timezone suffix",
			timestamp: "2024-01-15 14:30:00 EST",
			timezone:  "America/New_York",
			wantErr:   false,
			validate: func(tm time.Time) bool {
				return tm.Year() == 2024 && tm.Month() == 1 && tm.Day() == 15
			},
		},
		{
			name:      "invalid format",
			timestamp: "Jan 15, 2024 2:30 PM",
			timezone:  "UTC",
			wantErr:   true,
		},
		{
			name:      "empty timestamp",
			timestamp: "",
			timezone:  "UTC",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := passageoftime.ParseOptions{
				Timezone: tt.timezone,
			}
			result, err := passageoftime.ParseTimestamp(tt.timestamp, options)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil && !tt.validate(result) {
				t.Errorf("ParseTimestamp() validation failed for %v", result)
			}
		})
	}
}

// Test formatDuration function
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name       string
		seconds    float64
		style      string
		isNegative bool
		want       string
	}{
		{
			name:       "full format with all units",
			seconds:    93784, // 1 day, 2 hours, 3 minutes, 4 seconds
			style:      "full",
			isNegative: false,
			want:       "1 day, 2 hours, 3 minutes, 4 seconds",
		},
		{
			name:       "compact format",
			seconds:    93784,
			style:      "compact",
			isNegative: false,
			want:       "1d 2h 3m 4s",
		},
		{
			name:       "minimal format with hours",
			seconds:    3665, // 1 hour, 1 minute, 5 seconds
			style:      "minimal",
			isNegative: false,
			want:       "1:01:05",
		},
		{
			name:       "minimal format no hours",
			seconds:    125, // 2 minutes, 5 seconds
			style:      "minimal",
			isNegative: false,
			want:       "2:05",
		},
		{
			name:       "negative duration",
			seconds:    3600,
			style:      "full",
			isNegative: true,
			want:       "-1 hour",
		},
		{
			name:       "zero duration",
			seconds:    0,
			style:      "full",
			isNegative: false,
			want:       "0 seconds",
		},
		{
			name:       "singular second",
			seconds:    1,
			style:      "full",
			isNegative: false,
			want:       "1 second",
		},
		{
			name:       "plural seconds",
			seconds:    2,
			style:      "full",
			isNegative: false,
			want:       "2 seconds",
		},
		{
			name:       "singular minute",
			seconds:    60,
			style:      "full",
			isNegative: false,
			want:       "1 minute",
		},
		{
			name:       "plural minutes",
			seconds:    120,
			style:      "full",
			isNegative: false,
			want:       "2 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := passageoftime.FormatDuration(tt.seconds, tt.style, tt.isNegative)
			if got != tt.want {
				t.Errorf("FormatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test getTimeContext function
func TestGetTimeContext(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")
	now := time.Date(2024, 1, 15, 15, 0, 0, 0, loc)

	tests := []struct {
		name    string
		t       time.Time
		seconds float64
		want    string
	}{
		{
			name:    "future time",
			t:       now.Add(time.Hour),
			seconds: -3600,
			want:    "in the future",
		},
		{
			name:    "just now",
			t:       now.Add(-30 * time.Second),
			seconds: 30,
			want:    "just now",
		},
		{
			name:    "earlier",
			t:       now.Add(-30 * time.Minute),
			seconds: 1800,
			want:    "earlier",
		},
		{
			name:    "earlier today",
			t:       now.Add(-2 * time.Hour),
			seconds: 7200,
			want:    "earlier today",
		},
		{
			name:    "yesterday",
			t:       now.Add(-25 * time.Hour),
			seconds: 90000,
			want:    "yesterday",
		},
		{
			name:    "this week",
			t:       now.Add(-3 * 24 * time.Hour),
			seconds: 259200,
			want:    "this week",
		},
		{
			name:    "this month",
			t:       now.Add(-10 * 24 * time.Hour),
			seconds: 864000,
			want:    "this month",
		},
		{
			name:    "a while ago",
			t:       now.Add(-40 * 24 * time.Hour),
			seconds: 3456000,
			want:    "a while ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := passageoftime.GetTimeContext(tt.t, now, tt.seconds)
			if got != tt.want {
				t.Errorf("GetTimeContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test getTimeDescription function
func TestGetTimeDescription(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")
	now := time.Date(2024, 1, 15, 15, 0, 0, 0, loc)

	tests := []struct {
		name       string
		t          time.Time
		isDateOnly bool
		want       string
	}{
		{
			name:       "today with time",
			t:          now,
			isDateOnly: false,
			want:       "today at 3:00 PM",
		},
		{
			name:       "today date only",
			t:          now,
			isDateOnly: true,
			want:       "today",
		},
		{
			name:       "tomorrow with time",
			t:          now.Add(24 * time.Hour),
			isDateOnly: false,
			want:       "tomorrow at 3:00 PM",
		},
		{
			name:       "yesterday with time",
			t:          now.Add(-24 * time.Hour),
			isDateOnly: false,
			want:       "yesterday at 3:00 PM",
		},
		{
			name:       "next week",
			t:          now.Add(3 * 24 * time.Hour), // Thursday
			isDateOnly: false,
			want:       "next Thursday at 3:00 PM",
		},
		{
			name:       "last week",
			t:          now.Add(-3 * 24 * time.Hour), // Friday
			isDateOnly: false,
			want:       "last Friday at 3:00 PM",
		},
		{
			name:       "far future",
			t:          now.Add(30 * 24 * time.Hour),
			isDateOnly: false,
			want:       "February 14, 2024 at 3:00 PM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := passageoftime.GetTimeDescription(tt.t, now, tt.isDateOnly)
			if got != tt.want {
				t.Errorf("GetTimeDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test parameter extraction helpers
func TestGetStringParam(t *testing.T) {
	params := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": nil,
	}

	tests := []struct {
		name         string
		key          string
		defaultValue string
		want         string
	}{
		{"existing string", "key1", "default", "value1"},
		{"missing key", "missing", "default", "default"},
		{"non-string value", "key2", "default", "default"},
		{"nil value", "key3", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := passageoftime.GetStringParam(params, tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetStringParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFloatParam(t *testing.T) {
	params := map[string]interface{}{
		"float":  3.14,
		"int":    42,
		"int64":  int64(100),
		"string": "not a number",
		"nil":    nil,
	}

	tests := []struct {
		name         string
		key          string
		defaultValue float64
		want         float64
	}{
		{"float value", "float", 0, 3.14},
		{"int value", "int", 0, 42},
		{"int64 value", "int64", 0, 100},
		{"missing key", "missing", 1.5, 1.5},
		{"string value", "string", 2.5, 2.5},
		{"nil value", "nil", 3.5, 3.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := passageoftime.GetFloatParam(params, tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetFloatParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to parse JSON result strings
func parseJSONResult(result string) (map[string]interface{}, error) {
	// Remove the "map[" prefix and trailing "]" if present
	if strings.HasPrefix(result, "map[") {
		result = strings.TrimPrefix(result, "map[")
		result = strings.TrimSuffix(result, "]")
	}
	
	// Try to parse as JSON
	var data map[string]interface{}
	err := json.Unmarshal([]byte(result), &data)
	if err != nil {
		// If not valid JSON, try to parse the Go map format
		// This is a simplified parser for testing
		data = make(map[string]interface{})
		// Basic parsing logic would go here
		// For now, return empty map
	}
	return data, nil
}