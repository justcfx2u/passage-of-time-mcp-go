package passageoftime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

// ParseFuzzyTimestamp implements a 4-layer parsing approach (TDD-driven):
// 1. Duration parsing (time.ParseDuration) - handles "-14d", "2h30m", etc.
// 2. Dateparse library - handles standard timestamp formats  
// 3. NLP parsing (when library) - handles natural language
// 4. Fallback - existing strict parsing
func ParseFuzzyTimestamp(input string, options ParseOptions) (time.Time, error) {
	// Load timezone for context
	loc, err := time.LoadLocation(options.Timezone)
	if err != nil {
		loc = time.UTC // fallback to UTC
	}
	
	// Ensure reference time is in correct timezone
	refInTz := options.ReferenceTime.In(loc)
	
	// Layer 1: Try duration parsing first (handles relative durations like "-14d", "2h30m")
	if options.EnableFuzzyParsing {
		if parsed, err := parseDurationRelative(input, refInTz); err == nil {
			return parsed, nil
		}
	}
	
	// Layer 2: Try dateparse (handles standard timestamp formats efficiently)
	if parsed, err := dateparse.ParseIn(input, loc); err == nil {
		return parsed, nil
	}
	
	// Layer 3: Try NLP parsing if enabled (handles natural language)
	if options.EnableFuzzyParsing {
		if parsed, err := parseWithWhenLibrary(input, refInTz, loc); err == nil {
			return parsed, nil
		}
	}
	
	// Layer 4: Final fallback to existing strict parsing
	return ParseTimestamp(input, options)
}

// parseWithWhenLibrary uses the when library for natural language parsing
func parseWithWhenLibrary(input string, referenceTime time.Time, loc *time.Location) (time.Time, error) {
	// Try compound duration parsing first (e.g. "3 days and 2 hours ago")
	if parsed, err := parseCompoundDuration(input, referenceTime, loc); err == nil {
		return parsed, nil
	}
	
	// Initialize when parser with English and common rules
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)
	
	// Parse with reference time in the specified timezone
	refInTz := referenceTime.In(loc)
	
	r, err := w.Parse(input, refInTz)
	if err != nil {
		return time.Time{}, fmt.Errorf("NLP parsing failed: %w", err)
	}
	
	if r == nil {
		return time.Time{}, fmt.Errorf("NLP parsing returned nil result for: %s", input)
	}
	
	if r.Time.IsZero() {
		return time.Time{}, fmt.Errorf("NLP parsing returned zero time for: %s", input)
	}
	
	// Ensure result is in the correct timezone
	return r.Time.In(loc), nil
}

// parseDurationRelative handles duration inputs like "-14d", "2h30m", etc.
// This is Layer 1 of the 4-layer parsing chain
func parseDurationRelative(input string, referenceTime time.Time) (time.Time, error) {
	// First try standard Go duration parsing for formats like "2h30m", "-5s", "1m"
	if duration, err := time.ParseDuration(input); err == nil {
		return referenceTime.Add(duration), nil
	}
	
	// Handle day/week/month/year durations that Go's ParseDuration doesn't support
	// Support formats like: "-14d", "2w", "-1M", "1y", etc.
	if parsed, err := parseExtendedDuration(input, referenceTime); err == nil {
		return parsed, nil
	}
	
	return time.Time{}, fmt.Errorf("not a valid duration format: %s", input)
}

// parseExtendedDuration handles durations with days/weeks/months/years
func parseExtendedDuration(input string, referenceTime time.Time) (time.Time, error) {
	// Regex to match: optional minus, number, unit (d/w/M/y)
	re := regexp.MustCompile(`^(-?)(\d+)([dwMy])$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(input))
	
	if len(matches) != 4 {
		return time.Time{}, fmt.Errorf("invalid extended duration format")
	}
	
	isNegative := matches[1] == "-"
	valueStr := matches[2]
	unit := matches[3]
	
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration value: %s", valueStr)
	}
	
	if isNegative {
		value = -value
	}
	
	switch unit {
	case "d": // days
		return referenceTime.AddDate(0, 0, value), nil
	case "w": // weeks
		return referenceTime.AddDate(0, 0, value*7), nil
	case "M": // months
		return referenceTime.AddDate(0, value, 0), nil
	case "y": // years
		return referenceTime.AddDate(value, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported duration unit: %s", unit)
	}
}

// parseCompoundDuration handles compound durations like "3 days and 2 hours ago"
func parseCompoundDuration(input string, referenceTime time.Time, loc *time.Location) (time.Time, error) {
	// Detect compound duration patterns with "and"
	compoundPattern := regexp.MustCompile(`(?i)(.+?)\s+and\s+(.+?)(\s+ago)?$`)
	matches := compoundPattern.FindStringSubmatch(strings.TrimSpace(input))
	
	if len(matches) < 3 {
		return time.Time{}, fmt.Errorf("not a compound duration")
	}
	
	part1 := strings.TrimSpace(matches[1])
	part2 := strings.TrimSpace(matches[2])
	hasAgo := len(matches) > 3 && matches[3] != ""
	
	// Add "ago" to each part if the original had it
	if hasAgo {
		part1 += " ago"
		part2 += " ago"
	}
	
	// Initialize when parser
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)
	
	refInTz := referenceTime.In(loc)
	
	// Parse first part
	r1, err := w.Parse(part1, refInTz)
	if err != nil || r1 == nil || r1.Time.IsZero() {
		return time.Time{}, fmt.Errorf("failed to parse first part: %s", part1)
	}
	
	// Parse second part  
	r2, err := w.Parse(part2, refInTz)
	if err != nil || r2 == nil || r2.Time.IsZero() {
		return time.Time{}, fmt.Errorf("failed to parse second part: %s", part2)
	}
	
	// Calculate combined duration from reference time
	duration1 := r1.Time.Sub(refInTz)
	duration2 := r2.Time.Sub(refInTz)
	combinedDuration := duration1 + duration2
	
	// Apply combined duration to reference time
	result := refInTz.Add(combinedDuration)
	
	return result.In(loc), nil
}