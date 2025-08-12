package passageoftime

import "time"

// TimeResult represents the result of time-related operations
type TimeResult struct {
	// FormattedTime is the human-readable formatted time string
	FormattedTime string
	
	// Timestamp is the precise timestamp in ISO 8601 format
	Timestamp time.Time
	
	// Timezone is the timezone identifier (e.g., "America/New_York")
	Timezone string
}

// DurationResult represents the result of duration calculations
type DurationResult struct {
	// HumanReadable is the human-readable duration string (e.g., "2 hours ago")
	HumanReadable string
	
	// PreciseDescription includes precise timestamp in parentheses
	PreciseDescription string
	
	// Duration is the calculated duration in seconds
	Duration float64
	
	// StartTime is the start time for the duration calculation
	StartTime time.Time
	
	// EndTime is the end time for the duration calculation
	EndTime time.Time
	
	// Timezone is the timezone identifier used for calculations
	Timezone string
}

// TimezoneInfo represents timezone information
type TimezoneInfo struct {
	// ID is the IANA timezone identifier
	ID string
	
	// Name is the human-readable timezone name
	Name string
	
	// Offset is the current UTC offset in seconds
	Offset int
	
	// OffsetString is the formatted offset (e.g., "+05:00")
	OffsetString string
	
	// IsPopular indicates if this is a commonly used timezone
	IsPopular bool
}

// ParseOptions controls parsing behavior
type ParseOptions struct {
	// EnableFuzzyParsing enables natural language and duration parsing
	EnableFuzzyParsing bool
	
	// Timezone specifies the timezone for parsing and formatting
	Timezone string
	
	// ReferenceTime is the reference time for relative parsing
	ReferenceTime time.Time
}