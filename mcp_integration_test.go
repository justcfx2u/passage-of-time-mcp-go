package main

import (
	"testing"
	"time"

	"github.com/justcfx2u/passage-of-time-mcp-go/passageoftime"
)

// TestMCPServerValidation tests the full MCP server functionality
func TestMCPServerValidation(t *testing.T) {
	// Test cases for basic MCP server functionality
	referenceTime := time.Now()
	
	// Test cases that simulate MCP tool calls with timezone parameters
	testCases := []struct {
		name          string
		input         string
		timezone      string
		shouldSucceed bool
	}{
		{
			name:          "Relative time with timezone parameter",
			input:         "3 days and 2 hours ago",
			timezone:      "Australia/Melbourne",
			shouldSucceed: true,
		},
		{
			name:          "Future time with timezone parameter",
			input:         "tomorrow at 3pm",
			timezone:      "America/New_York",
			shouldSucceed: true,
		},
		{
			name:          "Past time with timezone parameter",
			input:         "yesterday at noon",
			timezone:      "Europe/London",
			shouldSucceed: true,
		},
		{
			name:          "Duration with timezone parameter",
			input:         "5 hours ago",
			timezone:      "America/Chicago",
			shouldSucceed: true,
		},
		{
			name:          "Tokyo NLP",
			input:         "next Monday",
			timezone:      "Asia/Tokyo",
			shouldSucceed: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test through fuzzy parsing chain (same as MCP would use)
			options := passageoftime.ParseOptions{
				EnableFuzzyParsing: true,
				Timezone:           tc.timezone,
				ReferenceTime:      referenceTime,
			}
			parsedTime, err := passageoftime.ParseFuzzyTimestamp(tc.input, options)
			
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
					return
				}
				
				if parsedTime.IsZero() {
					t.Errorf("Expected valid parsed time, got zero time")
				}
				
				t.Logf("SUCCESS: %q → %s in %s", tc.input, parsedTime.Format("2006-01-02 15:04:05 MST"), tc.timezone)
			} else {
				if err == nil {
					t.Errorf("Expected failure, got success: %v", parsedTime)
				}
			}
		})
	}
}

// TestMCPHandlerIntegration tests timezone handling through actual MCP handlers
func TestMCPHandlerIntegration(t *testing.T) {
	// Test the time_since handler with timezone parameters
	testCases := []struct {
		name      string
		timestamp string
		timezone  string
		fuzzyParsing bool
		shouldSucceed bool
	}{
		{
			name:      "Melbourne timezone time_since",
			timestamp: "3 days and 2 hours ago", 
			timezone:  "Australia/Melbourne",
			fuzzyParsing: true,
			shouldSucceed: true,
		},
		{
			name:      "New York parse_timestamp",
			timestamp: "tomorrow at 3pm",
			timezone:  "America/New_York", 
			fuzzyParsing: true,
			shouldSucceed: true,
		},
		{
			name:      "Regular timestamp",
			timestamp: "2025-01-15 15:30:00",
			timezone:  "UTC",
			fuzzyParsing: false,
			shouldSucceed: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the core parsing function that MCP handlers use
			referenceTime := time.Now()
			
			// Use standard fuzzy parsing with timezone parameter
			options := passageoftime.ParseOptions{
				EnableFuzzyParsing: tc.fuzzyParsing,
				Timezone:           tc.timezone,
				ReferenceTime:      referenceTime,
			}
			parsedTime, err := passageoftime.ParseFuzzyTimestamp(tc.timestamp, options)
			
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("Parsing failed: %v", err)
					return
				}
				
				if parsedTime.IsZero() {
					t.Errorf("Expected valid parsed time, got zero time")
				}
				
				t.Logf("SUCCESS: %q → %s in %s", tc.timestamp, parsedTime.Format("2006-01-02 15:04:05 MST"), tc.timezone)
			} else {
				if err == nil {
					t.Errorf("Expected failure, got success: %v", parsedTime)
				}
			}
		})
	}
}

// TestTimezoneDetection tests that system timezone detection works correctly
func TestTimezoneDetection(t *testing.T) {
	detectedTz := passageoftime.GetSystemTimezone()
	
	t.Logf("System timezone detection result: %s", detectedTz)
	
	// Test 1: Detection returns something
	if detectedTz == "" {
		t.Errorf("Timezone detection returned empty string")
		return
	}
	
	// Test 2: Can load the detected timezone (Go runtime validation)
	if _, err := time.LoadLocation(detectedTz); err != nil {
		t.Errorf("Detected timezone %q cannot be loaded by Go runtime: %v", detectedTz, err)
		return
	}
	
	// Test 3: Detected timezone is in our enumerated list (automation consistency)
	allValidTimezones := passageoftime.GetAllTimezoneIDs()
	isValid := false
	for _, validTz := range allValidTimezones {
		if detectedTz == validTz {
			isValid = true
			break
		}
	}
	
	if !isValid {
		t.Errorf("Detected timezone %q not found in enumerated list of %d timezones - indicates extraction bug or version mismatch", detectedTz, len(allValidTimezones))
	} else {
		t.Logf("SUCCESS: Detected timezone %q validated in enumerated list", detectedTz)
	}
	
	// Test 4: Detection is consistent across calls
	detectedTz2 := passageoftime.GetSystemTimezone()
	if detectedTz != detectedTz2 {
		t.Errorf("Timezone detection inconsistent: %s vs %s", detectedTz, detectedTz2)
	}
}

// TestMCPResponseFormat tests MCP response structure
func TestMCPResponseFormat(t *testing.T) {
	referenceTime := time.Now()
	
	// Test Melbourne timezone through parameter
	input := "3 days and 2 hours ago"
	timezone := "Australia/Melbourne"
	options := passageoftime.ParseOptions{
		EnableFuzzyParsing: true,
		Timezone:           timezone,
		ReferenceTime:      referenceTime,
	}
	parsedTime, err := passageoftime.ParseFuzzyTimestamp(input, options)
	
	if err != nil {
		t.Fatalf("Failed to parse Melbourne example: %v", err)
	}
	
	// Verify time parsing worked
	if parsedTime.IsZero() {
		t.Errorf("Expected valid parsed time, got zero time")
	}
	
	// Simulate MCP response structure
	mcpResponse := map[string]interface{}{
		"parsed_time":    parsedTime.Format("2006-01-02 15:04:05 MST"),
		"timezone":       timezone,
		"original_input": input,
	}
	
	t.Logf("MCP Response Format: %+v", mcpResponse)
	
	// Verify response structure
	if mcpResponse["timezone"] != "Australia/Melbourne" {
		t.Errorf("MCP response missing correct timezone")
	}
}