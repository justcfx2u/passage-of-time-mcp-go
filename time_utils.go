package main

import (
	"fmt"
	"strings"
	"time"
)

// parseTimestamp parses a timestamp string in standard formats
func parseTimestamp(timestamp, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
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

// formatDuration formats seconds into human-readable text
func formatDuration(seconds float64, style string, isNegative bool) string {
	days := int(seconds / 86400)
	hours := int((int(seconds) % 86400) / 3600)
	minutes := int((int(seconds) % 3600) / 60)
	secs := int(seconds) % 60
	
	var result string
	
	switch style {
	case "compact":
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
		result = strings.Join(parts, " ")
		
	case "minimal":
		if days > 0 {
			result = fmt.Sprintf("%d:%02d:%02d:%02d", days, hours, minutes, secs)
		} else if hours > 0 {
			result = fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
		} else {
			result = fmt.Sprintf("%d:%02d", minutes, secs)
		}
		
	default: // "full"
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
		result = strings.Join(parts, ", ")
	}
	
	if isNegative {
		result = "-" + result
	}
	
	return result
}

// getTimeContext generates contextual description for time difference
func getTimeContext(t, now time.Time, seconds float64) string {
	absSeconds := seconds
	if seconds < 0 {
		absSeconds = -seconds
	}
	
	switch {
	case seconds < 0:
		return "in the future"
	case absSeconds < 60:
		return "just now"
	case absSeconds < 3600:
		return "earlier"
	case absSeconds < 86400:
		if t.Day() == now.Day() {
			return "earlier today"
		}
		return "yesterday"
	case absSeconds < 172800: // 2 days
		return "yesterday"
	case absSeconds < 604800: // 1 week
		return "this week"
	case absSeconds < 2592000: // 30 days
		return "this month"
	default:
		return "a while ago"
	}
}

// getTimeDescription generates natural language description for a time
func getTimeDescription(t, now time.Time, isDateOnly bool) string {
	daysDiff := int(t.Sub(now).Hours() / 24)
	
	var dayDesc string
	switch daysDiff {
	case 0:
		dayDesc = "today"
	case 1:
		dayDesc = "tomorrow"
	case -1:
		dayDesc = "yesterday"
	case 2, 3, 4, 5, 6, 7:
		dayDesc = "next " + t.Format("Monday")
	case -7, -6, -5, -4, -3, -2:
		dayDesc = "last " + t.Format("Monday")
	default:
		dayDesc = t.Format("January 2, 2006")
	}
	
	if isDateOnly {
		return dayDesc
	}
	
	timeDesc := t.Format("3:04 PM")
	return fmt.Sprintf("%s at %s", dayDesc, timeDesc)
}

// Helper functions for parameter extraction
func getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if v, ok := params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}

func getFloatParam(params map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := params[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return defaultValue
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}