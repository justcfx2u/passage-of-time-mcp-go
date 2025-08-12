package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/araddon/dateparse"
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
			name: "natural language format",
			args: ParseTimestampArgs{
				Timestamp:      "Jan 15, 2024 at 2:30 PM",
				TargetTimezone: "UTC",
			},
			wantErr: false,
			check: func(result string) bool {
				return strings.Contains(result, "2024-01-15") &&
					strings.Contains(result, "14:30")
			},
		},
		{
			name: "invalid format",
			args: ParseTimestampArgs{
				Timestamp:      "not a date at all just random text",
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
		name        string
		args        FormatDurationArgs
		wantPattern string  // Pattern to match (duration part)
		wantErr     bool
	}{
		{
			name: "full format",
			args: FormatDurationArgs{
				Seconds: 93784, // 1 day, 2 hours, 3 minutes, 4 seconds
				Style:   "full",
			},
			wantPattern: "1 day, 2 hours, 3 minutes, 4 seconds",
			wantErr:     false,
		},
		{
			name: "compact format",
			args: FormatDurationArgs{
				Seconds: 93784,
				Style:   "compact",
			},
			wantPattern: "1d 2h 3m 4s",
			wantErr:     false,
		},
		{
			name: "minimal format",
			args: FormatDurationArgs{
				Seconds: 3665, // 1 hour, 1 minute, 5 seconds
				Style:   "minimal",
			},
			wantPattern: "1:01:05",
			wantErr:     false,
		},
		{
			name: "minimal format no hours",
			args: FormatDurationArgs{
				Seconds: 125, // 2 minutes, 5 seconds
				Style:   "minimal",
			},
			wantPattern: "2:05",
			wantErr:     false,
		},
		{
			name: "negative duration",
			args: FormatDurationArgs{
				Seconds: -3600,
				Style:   "full",
			},
			wantPattern: "-1 hour",
			wantErr:     false,
		},
		{
			name: "zero duration",
			args: FormatDurationArgs{
				Seconds: 0,
				Style:   "full",
			},
			wantPattern: "0 seconds",
			wantErr:     false,
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
				if !strings.Contains(text, tt.wantPattern) {
					t.Errorf("handleFormatDuration() = %v, want to contain %v", text, tt.wantPattern)
				}
				
				// Extract and validate the precise timestamp in parentheses
				startParen := strings.Index(text, "(")
				endParen := strings.LastIndex(text, ")")
				if startParen == -1 || endParen == -1 || startParen >= endParen {
					t.Errorf("handleFormatDuration() = %v, expected precise timestamp in parentheses", text)
				} else {
					timestampStr := text[startParen+1 : endParen]
					// Use dateparse to validate the timestamp
					if _, err := dateparse.ParseAny(timestampStr); err != nil {
						t.Errorf("handleFormatDuration() = %v, timestamp in parentheses '%s' is not a valid date/time: %v", text, timestampStr, err)
					}
				}
			}
		})
	}
}

// TestHandleListTimezones tests the list_timezones handler with pagination functionality
func TestHandleListTimezones(t *testing.T) {
	tests := []struct {
		name       string
		args       ListTimezonesArgs
		wantErr    bool
		checkFunc  func(*testing.T, string)
	}{
		{
			name: "default popular timezones",
			args: ListTimezonesArgs{},
			wantErr: false,
			checkFunc: validateDefaultPopularResponse,
		},
		{
			name: "first page with limit",
			args: ListTimezonesArgs{
				Limit: 10,
				Page:  1,
			},
			wantErr: false,
			checkFunc: validateFirstPageResponse,
		},
		{
			name: "pagination metadata validation",
			args: ListTimezonesArgs{
				Limit: 50,
				Page:  2,
			},
			wantErr: false,
			checkFunc: validatePaginationMetadata,
		},
		{
			name: "filter with pagination",
			args: ListTimezonesArgs{
				Filter: "America",
				Limit:  20,
				Page:   1,
			},
			wantErr: false,
			checkFunc: validateFilteredPaginationResponse,
		},
		{
			name: "boundary conditions - empty page",
			args: ListTimezonesArgs{
				Limit: 10,
				Page:  999, // Page beyond available data
			},
			wantErr: false,
			checkFunc: validateEmptyPageResponse,
		},
		{
			name: "max limit enforcement",
			args: ListTimezonesArgs{
				Limit: 150, // Over max of 100
				Page:  1,
			},
			wantErr: false,
			checkFunc: validateMaxLimitEnforcement,
		},
		{
			name: "filter with no results",
			args: ListTimezonesArgs{
				Filter: "NonExistentTimezone",
				Limit:  10,
				Page:   1,
			},
			wantErr: false,
			checkFunc: validateNoResultsResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := &mcp.CallToolParamsFor[ListTimezonesArgs]{
				Arguments: tt.args,
			}

			got, err := handleListTimezones(ctx, nil, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleListTimezones() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && len(got.Content) > 0 {
				text := got.Content[0].(*mcp.TextContent).Text
				if tt.checkFunc != nil {
					tt.checkFunc(t, text)
				}
			}
		})
	}
}

// Helper functions for pagination validation

// validateDefaultPopularResponse checks the default popular timezones response
func validateDefaultPopularResponse(t *testing.T, response string) {
	// Should use popular timezones (using_popular: true)
	if !strings.Contains(response, "using_popular:true") {
		t.Errorf("Expected using_popular:true in response: %s", response)
	}
	
	// Should have 25 popular timezones by default
	if !strings.Contains(response, "returned_count:25") {
		t.Errorf("Expected returned_count:25 for default popular response: %s", response)
	}
	
	// Should contain essential pagination metadata
	validateBasicPaginationFields(t, response)
}

// validateFirstPageResponse checks first page pagination response
func validateFirstPageResponse(t *testing.T, response string) {
	// Should be page 1
	if !strings.Contains(response, "page:1") {
		t.Errorf("Expected page:1 in response: %s", response)
	}
	
	// Should have has_prev_page: false for first page
	if !strings.Contains(response, "has_prev_page:false") {
		t.Errorf("Expected has_prev_page:false for first page: %s", response)
	}
	
	// Should have limit: 10
	if !strings.Contains(response, "limit:10") {
		t.Errorf("Expected limit:10 in response: %s", response)
	}
	
	validateBasicPaginationFields(t, response)
	validateTimezoneDataStructure(t, response)
}

// validatePaginationMetadata checks pagination calculation accuracy
func validatePaginationMetadata(t *testing.T, response string) {
	// Should be page 2
	if !strings.Contains(response, "page:2") {
		t.Errorf("Expected page:2 in response: %s", response)
	}
	
	// Should have has_prev_page: true for page 2
	if !strings.Contains(response, "has_prev_page:true") {
		t.Errorf("Expected has_prev_page:true for page 2: %s", response)
	}
	
	// Should have limit: 50
	if !strings.Contains(response, "limit:50") {
		t.Errorf("Expected limit:50 in response: %s", response)
	}
	
	validateBasicPaginationFields(t, response)
	validateTimezoneDataStructure(t, response)
}

// validateFilteredPaginationResponse checks filtered results with pagination
func validateFilteredPaginationResponse(t *testing.T, response string) {
	// Should have filter: America
	if !strings.Contains(response, "filter:America") {
		t.Errorf("Expected filter:America in response: %s", response)
	}
	
	// Should not use popular default when filtering
	if !strings.Contains(response, "using_popular:false") {
		t.Errorf("Expected using_popular:false when filtering: %s", response)
	}
	
	// Should have limit: 20
	if !strings.Contains(response, "limit:20") {
		t.Errorf("Expected limit:20 in response: %s", response)
	}
	
	validateBasicPaginationFields(t, response)
	validateTimezoneDataStructure(t, response)
}

// validateEmptyPageResponse checks behavior when requesting page beyond available data
func validateEmptyPageResponse(t *testing.T, response string) {
	// Should be page 999 as requested
	if !strings.Contains(response, "page:999") {
		t.Errorf("Expected page:999 in response: %s", response)
	}
	
	// Should have returned_count: 0 for empty page
	if !strings.Contains(response, "returned_count:0") {
		t.Errorf("Expected returned_count:0 for empty page: %s", response)
	}
	
	// Should have has_next_page: false
	if !strings.Contains(response, "has_next_page:false") {
		t.Errorf("Expected has_next_page:false for page beyond data: %s", response)
	}
	
	validateBasicPaginationFields(t, response)
}

// validateMaxLimitEnforcement checks that limit is capped at 100
func validateMaxLimitEnforcement(t *testing.T, response string) {
	// Should enforce max limit of 100
	if !strings.Contains(response, "limit:100") {
		t.Errorf("Expected limit to be enforced to 100: %s", response)
	}
	
	validateBasicPaginationFields(t, response)
	validateTimezoneDataStructure(t, response)
}

// validateNoResultsResponse checks behavior when filter returns no results
func validateNoResultsResponse(t *testing.T, response string) {
	// Should have filter: NonExistentTimezone
	if !strings.Contains(response, "filter:NonExistentTimezone") {
		t.Errorf("Expected filter:NonExistentTimezone in response: %s", response)
	}
	
	// Should have total_filtered: 0
	if !strings.Contains(response, "total_filtered:0") {
		t.Errorf("Expected total_filtered:0 for no results: %s", response)
	}
	
	// Should have returned_count: 0
	if !strings.Contains(response, "returned_count:0") {
		t.Errorf("Expected returned_count:0 for no results: %s", response)
	}
	
	validateBasicPaginationFields(t, response)
}

// validateBasicPaginationFields checks that all required pagination fields are present
func validateBasicPaginationFields(t *testing.T, response string) {
	requiredFields := []string{
		"total_available:",
		"total_filtered:",
		"returned_count:",
		"page:",
		"limit:",
		"total_pages:",
		"has_next_page:",
		"has_prev_page:",
		"using_popular:",
		"timezones:",
	}
	
	for _, field := range requiredFields {
		if !strings.Contains(response, field) {
			t.Errorf("Missing required pagination field '%s' in response: %s", field, response)
		}
	}
}

// validateTimezoneDataStructure checks that timezone data has proper structure
func validateTimezoneDataStructure(t *testing.T, response string) {
	// Should have timezone objects with required fields
	requiredTimezoneFields := []string{
		"id:",
		"name:",
		"offset:",
		"offset_str:",
		"current_time:",
	}
	
	// Only check if we have timezones in the response (not empty page)
	if strings.Contains(response, "returned_count:0") {
		return // Skip validation for empty results
	}
	
	for _, field := range requiredTimezoneFields {
		if !strings.Contains(response, field) {
			t.Errorf("Missing required timezone field '%s' in response: %s", field, response)
		}
	}
	
	// Should contain valid timezone format (IANA identifiers)
	if !strings.Contains(response, "/") {
		t.Errorf("Expected IANA timezone identifiers (containing '/') in response: %s", response)
	}
}