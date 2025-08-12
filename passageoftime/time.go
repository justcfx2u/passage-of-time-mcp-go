package passageoftime

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

// CurrentDateTime returns the current date and time in the specified timezone
func CurrentDateTime(options ParseOptions) (*TimeResult, error) {
	// Load timezone
	loc, err := time.LoadLocation(options.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}
	
	now := time.Now().In(loc)
	
	return &TimeResult{
		FormattedTime: now.Format("2006-01-02 15:04:05 MST"),
		Timestamp:     now,
		Timezone:      options.Timezone,
	}, nil
}

// TimeDifference calculates the time difference between two timestamps
func TimeDifference(timestamp1, timestamp2 string, options ParseOptions) (*DurationResult, error) {
	// Parse both timestamps
	t1, err := ParseTimestamp(timestamp1, options)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp1: %w", err)
	}
	
	t2, err := ParseTimestamp(timestamp2, options)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp2: %w", err)
	}
	
	// Calculate difference
	diff := t2.Sub(t1)
	seconds := diff.Seconds()
	
	// Format human-readable duration
	humanText := formatDurationWithHumanize(t1, t2, "full", options.Timezone)
	
	return &DurationResult{
		HumanReadable:      humanText,
		PreciseDescription: humanText,
		Duration:          seconds,
		StartTime:         t1,
		EndTime:           t2,
		Timezone:          options.Timezone,
	}, nil
}

// TimeSince calculates the time elapsed since a given timestamp until now
func TimeSince(timestamp string, options ParseOptions) (*DurationResult, error) {
	// Parse the timestamp
	t, err := ParseTimestamp(timestamp, options)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}
	
	// Get current time in the specified timezone
	loc, err := time.LoadLocation(options.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}
	
	now := time.Now().In(loc)
	
	// Calculate difference
	diff := now.Sub(t)
	seconds := diff.Seconds()
	
	// Format human-readable duration
	humanText := formatDurationWithHumanize(t, now, "full", options.Timezone)
	
	return &DurationResult{
		HumanReadable:      humanText,
		PreciseDescription: humanText,
		Duration:          seconds,
		StartTime:         t,
		EndTime:           now,
		Timezone:          options.Timezone,
	}, nil
}

// ParseTimestamp parses a timestamp string in standard formats
func ParseTimestamp(timestamp string, options ParseOptions) (time.Time, error) {
	loc, err := time.LoadLocation(options.Timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone: %w", err)
	}
	
	timestamp = strings.TrimSpace(timestamp)
	
	// Try ISO 8601 formats first (most standard)
	// Full ISO 8601 with timezone
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		// Convert to requested timezone
		return t.In(loc), nil
	}
	
	// ISO 8601 with nanoseconds
	if t, err := time.Parse(time.RFC3339Nano, timestamp); err == nil {
		return t.In(loc), nil
	}
	
	// ISO 8601 without timezone (assume provided timezone)
	if t, err := time.ParseInLocation("2006-01-02T15:04:05", timestamp, loc); err == nil {
		return t, nil
	}
	
	// ISO 8601 with milliseconds without timezone
	if t, err := time.ParseInLocation("2006-01-02T15:04:05.000", timestamp, loc); err == nil {
		return t, nil
	}
	
	// Try full timestamp format (backward compatibility)
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", timestamp, loc); err == nil {
		return t, nil
	}
	
	// Try date-only format
	if t, err := time.ParseInLocation("2006-01-02", timestamp, loc); err == nil {
		return t, nil
	}
	
	// Try with timezone suffix (ignore it, use provided timezone)
	parts := strings.Fields(timestamp)
	if len(parts) == 3 && len(parts[2]) <= 4 {
		// Likely has timezone suffix
		dtStr := parts[0] + " " + parts[1]
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", dtStr, loc); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("invalid timestamp format: '%s'. Expected ISO 8601 (e.g., '2025-07-19T08:45:40.501Z'), 'YYYY-MM-DD HH:MM:SS', or 'YYYY-MM-DD'", timestamp)
}

// formatDurationWithHumanize uses go-humanize for duration formatting with precise timestamps
func formatDurationWithHumanize(referenceTime time.Time, targetTime time.Time, style string, timezone string) string {
	// Get human-readable duration using go-humanize
	fuzzyText := humanize.Time(targetTime)
	
	// Apply style preferences
	switch style {
	case "compact":
		// For compact style, use a shorter format
		diff := targetTime.Sub(referenceTime)
		if diff < 0 {
			diff = -diff
			fuzzyText = "in " + humanize.Time(referenceTime.Add(-diff))[len("in "):]
		}
	case "minimal":
		// For minimal style, show just the time difference
		diff := targetTime.Sub(referenceTime)
		fuzzyText = diff.String()
	default: // "full" - use existing duration formatting 
		diff := targetTime.Sub(referenceTime)
		seconds := diff.Seconds()
		isNegative := seconds < 0
		if isNegative {
			seconds = -seconds
		}
		fuzzyText = formatDuration(seconds, "full", isNegative)
	}
	
	// Add precise timestamp in parentheses as required
	return formatWithPreciseTimestamp(fuzzyText, targetTime, timezone)
}

// formatDuration - Legacy function for backward compatibility
func formatDuration(seconds float64, style string, isNegative bool) string {
	// Convert seconds to time for humanize compatibility
	now := time.Now()
	targetTime := now.Add(time.Duration(seconds) * time.Second)
	if isNegative {
		targetTime = now.Add(-time.Duration(seconds) * time.Second)
	}
	
	// Use go-humanize for the basic formatting
	fuzzyText := humanize.Time(targetTime)
	
	// Apply style preferences (simplified for compatibility)
	switch style {
	case "compact":
		// Keep existing compact logic for now
		days := int(seconds / 86400)
		hours := int((int(seconds) % 86400) / 3600)
		minutes := int((int(seconds) % 3600) / 60)
		secs := int(seconds) % 60
		var parts []string
		if days > 0 {
			parts = append(parts, fmt.Sprintf("%dd", days))
		}
		if hours > 0 {
			parts = append(parts, fmt.Sprintf("%dh", hours))
		}
		if minutes > 0 {
			parts = append(parts, fmt.Sprintf("%dm", minutes))
		}
		if secs > 0 || len(parts) == 0 {
			parts = append(parts, fmt.Sprintf("%ds", secs))
		}
		fuzzyText = strings.Join(parts, " ")
	case "minimal":
		// Keep existing minimal logic
		days := int(seconds / 86400)
		hours := int((int(seconds) % 86400) / 3600)
		minutes := int((int(seconds) % 3600) / 60)
		secs := int(seconds) % 60
		if days > 0 {
			fuzzyText = fmt.Sprintf("%d:%02d:%02d:%02d", days, hours, minutes, secs)
		} else if hours > 0 {
			fuzzyText = fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
		} else {
			fuzzyText = fmt.Sprintf("%d:%02d", minutes, secs)
		}
	default: // "full" - maintain backward compatibility
		// Use original logic for backward compatibility with tests
		days := int(seconds / 86400)
		hours := int((int(seconds) % 86400) / 3600)
		minutes := int((int(seconds) % 3600) / 60)
		secs := int(seconds) % 60
		var parts []string
		if days > 0 {
			parts = append(parts, fmt.Sprintf("%d day%s", days, plural(days)))
		}
		if hours > 0 {
			parts = append(parts, fmt.Sprintf("%d hour%s", hours, plural(hours)))
		}
		if minutes > 0 {
			parts = append(parts, fmt.Sprintf("%d minute%s", minutes, plural(minutes)))
		}
		if secs > 0 || len(parts) == 0 {
			parts = append(parts, fmt.Sprintf("%d second%s", secs, plural(secs)))
		}
		fuzzyText = strings.Join(parts, ", ")
	}
	
	if isNegative {
		fuzzyText = "-" + fuzzyText
	}
	
	return fuzzyText
}

// formatWithPreciseTimestamp formats fuzzy output with precise timestamp in parentheses
func formatWithPreciseTimestamp(fuzzyText string, preciseTime time.Time, timezone string) string {
	// Load timezone for formatting
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	
	// Format precise timestamp
	preciseFormatted := preciseTime.In(loc).Format("2006-01-02 15:04:05 MST")
	
	// Combine fuzzy text with precise timestamp in parentheses
	return fmt.Sprintf("%s (%s)", fuzzyText, preciseFormatted)
}

// Helper functions
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// GetPopularTimezones returns the most commonly used timezones
func GetPopularTimezones() []TimezoneInfo {
	popularIds := []string{
		"UTC",
		"America/New_York",
		"America/Chicago",
		"America/Denver",
		"America/Los_Angeles",
		"America/Toronto",
		"America/Mexico_City",
		"America/Sao_Paulo",
		"America/Buenos_Aires",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"Europe/Rome",
		"Europe/Amsterdam",
		"Europe/Zurich",
		"Europe/Stockholm",
		"Europe/Moscow",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Asia/Hong_Kong",
		"Asia/Singapore",
		"Asia/Kolkata",
		"Asia/Dubai",
		"Australia/Sydney",
		"Pacific/Auckland",
	}
	
	var result []TimezoneInfo
	for _, id := range popularIds {
		if loc, err := time.LoadLocation(id); err == nil {
			now := time.Now().In(loc)
			_, offset := now.Zone()
			
			result = append(result, TimezoneInfo{
				ID:           id,
				Name:         id, // Use ID as name for now
				Offset:       offset,
				OffsetString: FormatOffset(offset),
				IsPopular:    true,
			})
		}
	}
	
	return result
}

// formatOffset formats a timezone offset in seconds to a readable string (private helper)
func formatOffset(offsetSeconds int) string {
	if offsetSeconds == 0 {
		return "+00:00"
	}
	
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}
	
	hours := offsetSeconds / 3600
	minutes := (offsetSeconds % 3600) / 60
	
	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}

