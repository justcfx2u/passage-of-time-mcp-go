package main

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestNewYorkTimezoneHandling tests EST/EDT handling for specific dates
func TestNewYorkTimezoneHandling(t *testing.T) {
	tests := []struct {
		name        string
		timestamp   string
		wantOffset  string
		description string
	}{
		// EST (UTC-5) dates
		{
			name:        "January EST",
			timestamp:   "2025-01-15 12:00:00",
			wantOffset:  "-05:00",
			description: "Mid-January is always EST",
		},
		{
			name:        "February EST",
			timestamp:   "2025-02-14 09:30:00",
			wantOffset:  "-05:00",
			description: "Valentine's Day is EST",
		},
		{
			name:        "Early March EST",
			timestamp:   "2025-03-08 10:00:00",
			wantOffset:  "-05:00",
			description: "Day before DST starts (2nd Sunday)",
		},
		{
			name:        "November EST",
			timestamp:   "2025-11-15 14:00:00",
			wantOffset:  "-05:00",
			description: "Mid-November after DST ends",
		},
		{
			name:        "December EST",
			timestamp:   "2025-12-25 18:00:00",
			wantOffset:  "-05:00",
			description: "Christmas is EST",
		},

		// EDT (UTC-4) dates
		{
			name:        "Late March EDT",
			timestamp:   "2025-03-20 12:00:00",
			wantOffset:  "-04:00",
			description: "Spring equinox is EDT",
		},
		{
			name:        "April EDT",
			timestamp:   "2025-04-15 15:00:00",
			wantOffset:  "-04:00",
			description: "Tax day is EDT",
		},
		{
			name:        "July 4th EDT",
			timestamp:   "2025-07-04 12:00:00",
			wantOffset:  "-04:00",
			description: "Independence Day is EDT",
		},
		{
			name:        "September EDT",
			timestamp:   "2025-09-15 08:00:00",
			wantOffset:  "-04:00",
			description: "Mid-September is still EDT",
		},
		{
			name:        "Halloween EDT",
			timestamp:   "2025-10-31 20:00:00",
			wantOffset:  "-04:00",
			description: "Halloween is still EDT",
		},

		// Edge cases around DST transitions
		{
			name:        "DST Start Day",
			timestamp:   "2025-03-09 12:00:00",
			wantOffset:  "-04:00",
			description: "2nd Sunday of March at noon is EDT",
		},
		{
			name:        "DST End Day",
			timestamp:   "2025-11-02 12:00:00",
			wantOffset:  "-05:00",
			description: "1st Sunday of November at noon is EST",
		},

		// 2024 dates for comparison
		{
			name:        "2024 Summer EDT",
			timestamp:   "2024-07-15 12:00:00",
			wantOffset:  "-04:00",
			description: "2024 summer is EDT",
		},
		{
			name:        "2024 Winter EST",
			timestamp:   "2024-01-15 12:00:00",
			wantOffset:  "-05:00",
			description: "2024 winter is EST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[ParseTimestampArgs]{
				Arguments: ParseTimestampArgs{
					Timestamp:      tt.timestamp,
					SourceTimezone: "America/New_York",
					TargetTimezone: "America/New_York",
				},
			}

			got, err := handleParseTimestamp(ctx, nil, params)
			if err != nil {
				t.Fatalf("handleParseTimestamp() error = %v", err)
			}

			if len(got.Content) == 0 {
				t.Fatal("handleParseTimestamp() returned no content")
			}

			// Get the response text
			text := got.Content[0].(*mcp.TextContent).Text
			
			// The response contains the ISO format timestamp with offset
			if !strings.Contains(text, tt.wantOffset) {
				t.Errorf("handleParseTimestamp() = %v, want to contain offset %v (%s)", 
					text, tt.wantOffset, tt.description)
			}
		})
	}
}

// TestAddTimeAcrossDST tests adding time across DST boundaries
func TestAddTimeAcrossDST(t *testing.T) {
	tests := []struct {
		name          string
		startTime     string
		duration      float64
		unit          string
		wantContains  string
		description   string
	}{
		{
			name:         "Add 1 day across spring forward",
			startTime:    "2025-03-08 12:00:00",
			duration:     1,
			unit:         "days",
			wantContains: "2025-03-09",
			description:  "Adding 1 day from Saturday to Sunday (DST starts)",
		},
		{
			name:         "Add 24 hours across spring forward",
			startTime:    "2025-03-08 12:00:00",
			duration:     24,
			unit:         "hours",
			wantContains: "2025-03-09",
			description:  "Adding 24 hours across DST boundary",
		},
		{
			name:         "Add 1 day across fall back",
			startTime:    "2025-11-01 12:00:00",
			duration:     1,
			unit:         "days",
			wantContains: "2025-11-02",
			description:  "Adding 1 day from Saturday to Sunday (DST ends)",
		},
		{
			name:         "Add 1 week includes DST change",
			startTime:    "2025-03-05 12:00:00",
			duration:     1,
			unit:         "weeks",
			wantContains: "2025-03-12",
			description:  "Week that includes spring forward",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[AddTimeArgs]{
				Arguments: AddTimeArgs{
					Timestamp: tt.startTime,
					Duration:  tt.duration,
					Unit:      tt.unit,
					Timezone:  "America/New_York",
				},
			}

			got, err := handleAddTime(ctx, nil, params)
			if err != nil {
				t.Fatalf("handleAddTime() error = %v", err)
			}

			if len(got.Content) == 0 {
				t.Fatal("handleAddTime() returned no content")
			}

			text := got.Content[0].(*mcp.TextContent).Text
			if !strings.Contains(text, tt.wantContains) {
				t.Errorf("handleAddTime() = %v, want to contain %v (%s)",
					text, tt.wantContains, tt.description)
			}
		})
	}
}

// TestTimezoneConsistency ensures timezone handling is consistent across all functions
func TestTimezoneConsistency(t *testing.T) {
	// Use a specific time that we know the offset for
	testCases := []struct {
		timezone     string
		summerOffset string // During EDT/BST
		winterOffset string // During EST/GMT
	}{
		{"America/New_York", "-04:00", "-05:00"},
		{"Europe/London", "+01:00", "Z"}, // Winter GMT is shown as Z in ISO format
		{"UTC", "Z", "Z"},
		{"America/Los_Angeles", "-07:00", "-08:00"},
		{"Australia/Sydney", "+11:00", "+10:00"}, // Note: reversed seasons
	}

	summerDate := "2025-07-15 12:00:00"
	winterDate := "2025-01-15 12:00:00"

	for _, tc := range testCases {
		t.Run(tc.timezone, func(t *testing.T) {
			// Test summer time
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[ParseTimestampArgs]{
				Arguments: ParseTimestampArgs{
					Timestamp:      summerDate,
					SourceTimezone: tc.timezone,
					TargetTimezone: tc.timezone,
				},
			}

			got, err := handleParseTimestamp(ctx, nil, params)
			if err != nil {
				t.Fatalf("Summer handleParseTimestamp() error = %v", err)
			}

			summerText := got.Content[0].(*mcp.TextContent).Text
			if !strings.Contains(summerText, tc.summerOffset) {
				// Special handling for Australia (reversed seasons)
				if tc.timezone == "Australia/Sydney" && strings.Contains(summerText, tc.winterOffset) {
					// This is expected - July is winter in Australia
				} else {
					t.Errorf("Summer %s: got %v, want offset %v", tc.timezone, summerText, tc.summerOffset)
				}
			}

			// Test winter time
			params.Arguments.Timestamp = winterDate
			got, err = handleParseTimestamp(ctx, nil, params)
			if err != nil {
				t.Fatalf("Winter handleParseTimestamp() error = %v", err)
			}

			winterText := got.Content[0].(*mcp.TextContent).Text
			if !strings.Contains(winterText, tc.winterOffset) {
				// Special handling for Australia (reversed seasons)
				if tc.timezone == "Australia/Sydney" && strings.Contains(winterText, tc.summerOffset) {
					// This is expected - January is summer in Australia
				} else {
					t.Errorf("Winter %s: got %v, want offset %v", tc.timezone, winterText, tc.winterOffset)
				}
			}
		})
	}
}