package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Mock time for testing
var mockNow = time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)

// TestCurrentDateTime tests the current_datetime handler
func TestHandleCurrentDateTime(t *testing.T) {
	tests := []struct {
		name     string
		args     CurrentDateTimeArgs
		wantText string
		wantErr  bool
	}{
		{
			name:     "default timezone",
			args:     CurrentDateTimeArgs{Timezone: ""},
			wantText: "Z", // RFC3339 UTC uses Z
			wantErr:  false,
		},
		{
			name:     "UTC timezone",
			args:     CurrentDateTimeArgs{Timezone: "UTC"},
			wantText: "Z", // RFC3339 UTC uses Z
			wantErr:  false,
		},
		{
			name:     "New York timezone",
			args:     CurrentDateTimeArgs{Timezone: "America/New_York"},
			wantText: "", // Special handling in test for EST/EDT
			wantErr:  false,
		},
		{
			name:     "invalid timezone",
			args:     CurrentDateTimeArgs{Timezone: "Invalid/Timezone"},
			wantText: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[CurrentDateTimeArgs]{
				Arguments: tt.args,
			}

			got, err := handleCurrentDateTime(ctx, nil, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleCurrentDateTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got.Content) > 0 {
				text := got.Content[0].(*mcp.TextContent).Text
				
				// For RFC3339 format validation
				if tt.name != "invalid timezone" {
					// Parse the returned time to validate RFC3339 format
					parsedTime, parseErr := time.Parse(time.RFC3339, text)
					if parseErr != nil {
						t.Errorf("handleCurrentDateTime() returned invalid RFC3339 format: %v, error: %v", text, parseErr)
					}
					
					// Special handling for New York timezone test
					if tt.args.Timezone == "America/New_York" {
						// Check that offset is either -05:00 (EST) or -04:00 (EDT)
						if !strings.Contains(text, "-05:00") && !strings.Contains(text, "-04:00") {
							t.Errorf("handleCurrentDateTime() for New York = %v, expected EST (-05:00) or EDT (-04:00) offset", text)
						}
					} else if tt.wantText != "" && !strings.Contains(text, tt.wantText) {
						t.Errorf("handleCurrentDateTime() = %v, want to contain %v", text, tt.wantText)
					}
					
					// Verify the time is recent (within 1 second of now)
					if time.Since(parsedTime).Abs() > time.Second {
						t.Errorf("handleCurrentDateTime() returned time too far from now: %v", text)
					}
				}
			}
		})
	}
}

// TestTimeDifference tests the time_difference handler
func TestHandleTimeDifference(t *testing.T) {
	tests := []struct {
		name       string
		args       TimeDifferenceArgs
		wantResult map[string]interface{}
		wantErr    bool
	}{
		{
			name: "simple difference",
			args: TimeDifferenceArgs{
				Timestamp1: "2024-01-01 10:00:00",
				Timestamp2: "2024-01-01 13:30:00",
				Unit:       "auto",
				Timezone:   "UTC",
			},
			wantResult: map[string]interface{}{
				"seconds":     12600.0,
				"formatted":   "3 hours, 30 minutes",
				"is_negative": false,
			},
			wantErr: false,
		},
		{
			name: "negative difference",
			args: TimeDifferenceArgs{
				Timestamp1: "2024-01-01 13:30:00",
				Timestamp2: "2024-01-01 10:00:00",
				Unit:       "auto",
				Timezone:   "UTC",
			},
			wantResult: map[string]interface{}{
				"seconds":     -12600.0,
				"formatted":   "-3 hours, 30 minutes",
				"is_negative": true,
			},
			wantErr: false,
		},
		{
			name: "with specific unit",
			args: TimeDifferenceArgs{
				Timestamp1: "2024-01-01 10:00:00",
				Timestamp2: "2024-01-01 13:30:00",
				Unit:       "hours",
				Timezone:   "UTC",
			},
			wantResult: map[string]interface{}{
				"requested_unit": 3.5,
			},
			wantErr: false,
		},
		{
			name: "days difference",
			args: TimeDifferenceArgs{
				Timestamp1: "2024-01-01",
				Timestamp2: "2024-01-05",
				Unit:       "auto",
				Timezone:   "UTC",
			},
			wantResult: map[string]interface{}{
				"seconds":   345600.0,
				"formatted": "4 days",
			},
			wantErr: false,
		},
		{
			name: "invalid timestamp1",
			args: TimeDifferenceArgs{
				Timestamp1: "invalid date",
				Timestamp2: "2024-01-01 10:00:00",
				Unit:       "auto",
				Timezone:   "UTC",
			},
			wantErr: true,
		},
		{
			name: "invalid timestamp2",
			args: TimeDifferenceArgs{
				Timestamp1: "2024-01-01 10:00:00",
				Timestamp2: "invalid date",
				Unit:       "auto",
				Timezone:   "UTC",
			},
			wantErr: true,
		},
		{
			name: "invalid unit",
			args: TimeDifferenceArgs{
				Timestamp1: "2024-01-01 10:00:00",
				Timestamp2: "2024-01-01 13:30:00",
				Unit:       "invalid",
				Timezone:   "UTC",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[TimeDifferenceArgs]{
				Arguments: tt.args,
			}

			got, err := handleTimeDifference(ctx, nil, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleTimeDifference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got.Content) > 0 {
				text := got.Content[0].(*mcp.TextContent).Text
				// Verify result contains expected values
				for key, expectedValue := range tt.wantResult {
					if !strings.Contains(text, fmt.Sprintf("%v", expectedValue)) {
						t.Errorf("handleTimeDifference() result missing %s:%v in %s", key, expectedValue, text)
					}
				}
			}
		})
	}
}

// TestParseTimestamp tests the parse_timestamp handler
func TestHandleParseTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		args    ParseTimestampArgs
		wantErr bool
		check   func(string) bool
	}{
		{
			name: "basic parsing",
			args: ParseTimestampArgs{
				Timestamp:      "2024-01-15 14:30:00",
				TargetTimezone: "UTC",
			},
			wantErr: false,
			check: func(result string) bool {
				return strings.Contains(result, "2024-01-15") &&
					strings.Contains(result, "14:30:00") &&
					strings.Contains(result, "Monday")
			},
		},
		{
			name: "timezone conversion",
			args: ParseTimestampArgs{
				Timestamp:      "2024-01-15 14:30:00",
				SourceTimezone: "UTC",
				TargetTimezone: "America/New_York",
			},
			wantErr: false,
			check: func(result string) bool {
				return strings.Contains(result, "09:30") // UTC 14:30 is EST 09:30
			},
		},
		{
			name: "invalid format",
			args: ParseTimestampArgs{
				Timestamp:      "Jan 15, 2024 at 2:30 PM",
				TargetTimezone: "UTC",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[ParseTimestampArgs]{
				Arguments: tt.args,
			}

			got, err := handleParseTimestamp(ctx, nil, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleParseTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got.Content) > 0 {
				text := got.Content[0].(*mcp.TextContent).Text
				if tt.check != nil && !tt.check(text) {
					t.Errorf("handleParseTimestamp() check failed for result: %s", text)
				}
			}
		})
	}
}

// TestAddTime tests the add_time handler
func TestHandleAddTime(t *testing.T) {
	tests := []struct {
		name    string
		args    AddTimeArgs
		wantErr bool
		check   func(string) bool
	}{
		{
			name: "add hours",
			args: AddTimeArgs{
				Timestamp: "2024-01-15 10:00:00",
				Duration:  3,
				Unit:      "hours",
				Timezone:  "UTC",
			},
			wantErr: false,
			check: func(result string) bool {
				return strings.Contains(result, "13:00:00")
			},
		},
		{
			name: "add days",
			args: AddTimeArgs{
				Timestamp: "2024-01-15",
				Duration:  5,
				Unit:      "days",
				Timezone:  "UTC",
			},
			wantErr: false,
			check: func(result string) bool {
				return strings.Contains(result, "2024-01-20")
			},
		},
		{
			name: "subtract time",
			args: AddTimeArgs{
				Timestamp: "2024-01-15 10:00:00",
				Duration:  -2,
				Unit:      "hours",
				Timezone:  "UTC",
			},
			wantErr: false,
			check: func(result string) bool {
				return strings.Contains(result, "08:00:00")
			},
		},
		{
			name: "add weeks",
			args: AddTimeArgs{
				Timestamp: "2024-01-01",
				Duration:  2,
				Unit:      "weeks",
				Timezone:  "UTC",
			},
			wantErr: false,
			check: func(result string) bool {
				return strings.Contains(result, "2024-01-15")
			},
		},
		{
			name: "invalid timestamp",
			args: AddTimeArgs{
				Timestamp: "invalid date",
				Duration:  1,
				Unit:      "days",
				Timezone:  "UTC",
			},
			wantErr: true,
		},
		{
			name: "invalid unit",
			args: AddTimeArgs{
				Timestamp: "2024-01-15 10:00:00",
				Duration:  1,
				Unit:      "invalid",
				Timezone:  "UTC",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[AddTimeArgs]{
				Arguments: tt.args,
			}

			got, err := handleAddTime(ctx, nil, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleAddTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got.Content) > 0 {
				text := got.Content[0].(*mcp.TextContent).Text
				if tt.check != nil && !tt.check(text) {
					t.Errorf("handleAddTime() check failed for result: %s", text)
				}
			}
		})
	}
}

// TestTimestampContext tests the timestamp_context handler
func TestHandleTimestampContext(t *testing.T) {
	tests := []struct {
		name       string
		args       TimestampContextArgs
		wantValues map[string]interface{}
		wantErr    bool
	}{
		{
			name: "morning context",
			args: TimestampContextArgs{
				Timestamp: "2024-01-15 08:30:00",
				Timezone:  "UTC",
			},
			wantValues: map[string]interface{}{
				"time_of_day":      "early_morning",
				"typical_activity": "commute_time",
				"hour_24":          8,
			},
			wantErr: false,
		},
		{
			name: "lunch time",
			args: TimestampContextArgs{
				Timestamp: "2024-01-15 12:30:00",
				Timezone:  "UTC",
			},
			wantValues: map[string]interface{}{
				"time_of_day":      "afternoon",
				"typical_activity": "lunch_time",
			},
			wantErr: false,
		},
		{
			name: "weekend detection",
			args: TimestampContextArgs{
				Timestamp: "2024-01-13 10:00:00", // Saturday
				Timezone:  "UTC",
			},
			wantValues: map[string]interface{}{
				"is_weekend":        true,
				"is_business_hours": false,
				"day_of_week":       "Saturday",
			},
			wantErr: false,
		},
		{
			name: "business hours",
			args: TimestampContextArgs{
				Timestamp: "2024-01-15 14:00:00", // Monday
				Timezone:  "UTC",
			},
			wantValues: map[string]interface{}{
				"is_weekend":        false,
				"is_business_hours": true,
			},
			wantErr: false,
		},
		{
			name: "late night",
			args: TimestampContextArgs{
				Timestamp: "2024-01-15 23:30:00",
				Timezone:  "UTC",
			},
			wantValues: map[string]interface{}{
				"time_of_day":      "late_night",
				"typical_activity": "sleeping_time",
			},
			wantErr: false,
		},
		{
			name: "invalid timestamp",
			args: TimestampContextArgs{
				Timestamp: "invalid date",
				Timezone:  "UTC",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[TimestampContextArgs]{
				Arguments: tt.args,
			}

			got, err := handleTimestampContext(ctx, nil, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleTimestampContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got.Content) > 0 {
				text := got.Content[0].(*mcp.TextContent).Text
				for key, value := range tt.wantValues {
					if !strings.Contains(text, fmt.Sprintf("%v", value)) {
						t.Errorf("handleTimestampContext() missing %s:%v in %s", key, value, text)
					}
				}
			}
		})
	}
}

// TestFormatDuration tests the format_duration handler
func TestHandleFormatDuration(t *testing.T) {
	tests := []struct {
		name    string
		args    FormatDurationArgs
		want    string
		wantErr bool
	}{
		{
			name: "full format",
			args: FormatDurationArgs{
				Seconds: 93784, // 1 day, 2 hours, 3 minutes, 4 seconds
				Style:   "full",
			},
			want:    "1 day, 2 hours, 3 minutes, 4 seconds",
			wantErr: false,
		},
		{
			name: "compact format",
			args: FormatDurationArgs{
				Seconds: 93784,
				Style:   "compact",
			},
			want:    "1d 2h 3m 4s",
			wantErr: false,
		},
		{
			name: "minimal format",
			args: FormatDurationArgs{
				Seconds: 3665, // 1 hour, 1 minute, 5 seconds
				Style:   "minimal",
			},
			want:    "1:01:05",
			wantErr: false,
		},
		{
			name: "minimal format no hours",
			args: FormatDurationArgs{
				Seconds: 125, // 2 minutes, 5 seconds
				Style:   "minimal",
			},
			want:    "2:05",
			wantErr: false,
		},
		{
			name: "negative duration",
			args: FormatDurationArgs{
				Seconds: -3600,
				Style:   "full",
			},
			want:    "-1 hour",
			wantErr: false,
		},
		{
			name: "zero duration",
			args: FormatDurationArgs{
				Seconds: 0,
				Style:   "full",
			},
			want:    "0 seconds",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[FormatDurationArgs]{
				Arguments: tt.args,
			}

			got, err := handleFormatDuration(ctx, nil, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleFormatDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got.Content) > 0 {
				text := got.Content[0].(*mcp.TextContent).Text
				if text != tt.want {
					t.Errorf("handleFormatDuration() = %v, want %v", text, tt.want)
				}
			}
		})
	}
}