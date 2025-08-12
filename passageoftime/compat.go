package passageoftime

import (
	"fmt"
	"time"

	"github.com/justcfx2u/passage-of-time-mcp-go/passageoftime/internal"
)

// GetSystemTimezone returns the system's timezone identifier
func GetSystemTimezone() string {
	return internal.GetSystemTimezone()
}

// GetAllTimezoneIDs returns all available timezone identifiers
func GetAllTimezoneIDs() []string {
	return internal.GetAllTimezoneIDs()
}

// GetPopularTimezoneIDs returns popular timezone identifiers as strings
// This is a compatibility wrapper for the getPopularTimezones function
func GetPopularTimezoneIDs() []string {
	timezones := GetPopularTimezones()
	var ids []string
	for _, tz := range timezones {
		ids = append(ids, tz.ID)
	}
	return ids
}

// GetTimeContext generates contextual description for time difference
// This is a compatibility wrapper for the getTimeContext function
func GetTimeContext(t, now time.Time, seconds float64) string {
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

// GetTimeDescription generates natural language description for a time
// This is a compatibility wrapper for the getTimeDescription function  
func GetTimeDescription(t, now time.Time, isDateOnly bool) string {
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

// FormatDuration formats a duration in seconds into human-readable text
// This is a compatibility wrapper for the formatDuration function
func FormatDuration(seconds float64, style string, isNegative bool) string {
	return formatDuration(seconds, style, isNegative)
}

// FormatWithPreciseTimestamp formats fuzzy output with precise timestamp in parentheses
// This is a compatibility wrapper for the formatWithPreciseTimestamp function
func FormatWithPreciseTimestamp(fuzzyText string, preciseTime time.Time, timezone string) string {
	return formatWithPreciseTimestamp(fuzzyText, preciseTime, timezone)
}

// FormatOffset formats a timezone offset in seconds to a readable string
// This is a compatibility wrapper for the formatOffset function
func FormatOffset(offsetSeconds int) string {
	return formatOffset(offsetSeconds)
}

// GetStringParam extracts a string parameter from a map with a default value
func GetStringParam(params map[string]interface{}, key, defaultValue string) string {
	if value, exists := params[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetFloatParam extracts a float parameter from a map with a default value
func GetFloatParam(params map[string]interface{}, key string, defaultValue float64) float64 {
	if value, exists := params[key]; exists {
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return defaultValue
}