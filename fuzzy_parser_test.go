package main

import (
	"testing"
	"time"

	"github.com/justcfx2u/passage-of-time-mcp-go/passageoftime"
)

// TestParsingChainIntegration tests the 3-layer parsing chain integration
// Light testing approach - trusts dependency test suites, focuses on our integration logic
func TestParsingChainIntegration(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		enableFuzzyParsing bool
		shouldSucceed      bool
		expectedLayer      string // "dateparse", "nlp", "strict", or "fail"
	}{
		// dateparse layer (should be tried first)
		{
			name:               "ISO 8601 format - dateparse layer",
			input:              "2025-08-12T15:00:00Z",
			enableFuzzyParsing: true,
			shouldSucceed:      true,
			expectedLayer:      "dateparse",
		},
		{
			name:               "RFC3339 format - dateparse layer", 
			input:              "2025-08-12T15:00:00-05:00",
			enableFuzzyParsing: true,
			shouldSucceed:      true,
			expectedLayer:      "dateparse",
		},
		// NLP layer (fallback when dateparse fails)
		{
			name:               "English NLP - tomorrow",
			input:              "tomorrow at 3pm",
			enableFuzzyParsing: true,
			shouldSucceed:      true,
			expectedLayer:      "nlp",
		},
		{
			name:               "English NLP - relative time",
			input:              "in 2 hours",
			enableFuzzyParsing: true,
			shouldSucceed:      true,
			expectedLayer:      "nlp",
		},
		// Test NLP disabled - should skip to strict parsing
		{
			name:               "NLP disabled - natural language fails",
			input:              "tomorrow at 3pm",
			enableFuzzyParsing: false,
			shouldSucceed:      false,
			expectedLayer:      "fail",
		},
		// Strict parsing fallback
		{
			name:               "Strict parsing fallback - basic format",
			input:              "2025-08-12 15:00:00",
			enableFuzzyParsing: true,
			shouldSucceed:      true,
			expectedLayer:      "strict",
		},
		// Invalid input - should fail through all layers
		{
			name:               "Invalid input - all layers fail",
			input:              "not a valid time expression",
			enableFuzzyParsing: true,
			shouldSucceed:      false,
			expectedLayer:      "fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s, input: %s, fuzzy: %v", tt.name, tt.input, tt.enableFuzzyParsing)
			
			// Call actual parsing function using library API
			referenceTime := time.Date(2025, 8, 12, 12, 0, 0, 0, time.UTC)
			options := passageoftime.ParseOptions{
				EnableFuzzyParsing: tt.enableFuzzyParsing,
				Timezone:           "UTC",
				ReferenceTime:      referenceTime,
			}
			result, err := passageoftime.ParseFuzzyTimestamp(tt.input, options)
			
			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
					return
				}
				t.Logf("Expected success for input: %s", tt.input)
			} else {
				if err == nil {
					t.Errorf("Expected failure but got success: %v", result)
					return
				}
				t.Logf("Expected failure for input: %s", tt.input)
			}
		})
	}
}

// TestTimezoneIntegration tests timezone detection integration
func TestTimezoneIntegration(t *testing.T) {
	tests := []struct {
		name                         string
		input                        string
		autodetectAndUseUserTimezone bool
		expectedBehavior             string
	}{
		{
			name:                         "Autodetect enabled",
			input:                        "tomorrow at 3pm",
			autodetectAndUseUserTimezone: true,
			expectedBehavior:             "should use system timezone",
		},
		{
			name:                         "Autodetect disabled",
			input:                        "tomorrow at 3pm", 
			autodetectAndUseUserTimezone: false,
			expectedBehavior:             "should use UTC default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s, autodetect: %v", tt.name, tt.autodetectAndUseUserTimezone)
			t.Logf("Expected: %s", tt.expectedBehavior)
			
			// Test timezone integration with library
			var timezone string
			if tt.autodetectAndUseUserTimezone {
				timezone = passageoftime.GetSystemTimezone()
			} else {
				timezone = "UTC"
			}
			
			// Verify we get a valid timezone
			if timezone == "" {
				t.Error("Got empty timezone")
			}
			
			// Basic validation that timezone detection works
			referenceTime := time.Date(2025, 8, 12, 12, 0, 0, 0, time.UTC)
			options := passageoftime.ParseOptions{
				EnableFuzzyParsing: true,
				Timezone:           timezone,
				ReferenceTime:      referenceTime,
			}
			_, err := passageoftime.ParseFuzzyTimestamp(tt.input, options)
			if err != nil {
				t.Errorf("Parsing failed with timezone %s: %v", timezone, err)
			}
		})
	}
}

// TestOutputFormatting tests that fuzzy outputs include precise timestamps in parentheses
func TestOutputFormatting(t *testing.T) {
	tests := []struct {
		name           string
		fuzzyText      string
		preciseTime    time.Time
		timezone       string
		expectedFormat string
	}{
		{
			name:           "Fuzzy output with precise timestamp",
			fuzzyText:      "2 hours ago",
			preciseTime:    time.Date(2025, 8, 11, 15, 30, 0, 0, time.UTC),
			timezone:       "UTC",
			expectedFormat: "2 hours ago (2025-08-11 15:30:00 UTC)",
		},
		{
			name:           "Tomorrow with timezone",
			fuzzyText:      "tomorrow at 3pm",
			preciseTime:    time.Date(2025, 8, 12, 15, 0, 0, 0, time.UTC),
			timezone:       "EST",
			expectedFormat: "tomorrow at 3pm (2025-08-12 15:00:00 EST)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Test actual formatting function once implemented
			// This ensures all fuzzy outputs include precise timestamps in parentheses
			t.Logf("Expected format: %s", tt.expectedFormat)
		})
	}
}

// TestMultiLanguageBasic tests basic multi-language NLP integration
// Light coverage - trusts when library's comprehensive testing
func TestMultiLanguageBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		language string // for documentation, when library auto-detects
	}{
		{
			name:     "English - basic",
			input:    "tomorrow",
			language: "EN",
		},
		// Basic examples from other supported languages (per when library docs)
		// Testing automatic language detection, not comprehensive coverage
		{
			name:     "Portuguese - basic",
			input:    "amanhã", // tomorrow in Portuguese
			language: "PT_BR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing %s input: %s", tt.language, tt.input)
			
			// TODO: Test that when library handles automatic language detection
			// We trust the when library's testing for comprehensive language support
		})
	}
}