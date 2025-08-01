package main

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const defaultTimezone = "UTC"

// Tool argument structs
type CurrentDateTimeArgs struct {
	Timezone string `json:"timezone,omitempty" mcp:"Timezone name (e.g., 'UTC', 'US/Pacific'). Defaults to 'UTC'."`
}

type TimeDifferenceArgs struct {
	Timestamp1 string `json:"timestamp1" mcp:"First timestamp in format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD"`
	Timestamp2 string `json:"timestamp2" mcp:"Second timestamp in format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD"`
	Unit       string `json:"unit,omitempty" mcp:"Desired unit: auto, seconds, minutes, hours, days"`
	Timezone   string `json:"timezone,omitempty" mcp:"Timezone for parsing ambiguous timestamps"`
}

type TimeSinceArgs struct {
	Timestamp string `json:"timestamp" mcp:"Past timestamp in format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD"`
	Timezone  string `json:"timezone,omitempty" mcp:"Timezone for parsing and current time"`
}

type ParseTimestampArgs struct {
	Timestamp      string `json:"timestamp" mcp:"Timestamp string in format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD"`
	SourceTimezone string `json:"source_timezone,omitempty" mcp:"Timezone of the input (if None, uses target_timezone)"`
	TargetTimezone string `json:"target_timezone,omitempty" mcp:"Desired output timezone"`
}

type AddTimeArgs struct {
	Timestamp string  `json:"timestamp" mcp:"Starting timestamp in format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD"`
	Duration  float64 `json:"duration" mcp:"Amount to add (can be negative to subtract)"`
	Unit      string  `json:"unit" mcp:"Unit: seconds, minutes, hours, days, weeks"`
	Timezone  string  `json:"timezone,omitempty" mcp:"Timezone for calculations"`
}

type TimestampContextArgs struct {
	Timestamp string `json:"timestamp" mcp:"Timestamp to analyze in format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD"`
	Timezone  string `json:"timezone,omitempty" mcp:"Timezone for context"`
}

type FormatDurationArgs struct {
	Seconds float64 `json:"seconds" mcp:"Duration in seconds (can be negative)"`
	Style   string  `json:"style,omitempty" mcp:"Format style: full, compact, minimal"`
}

// registerTools registers all time-related tools with the MCP server
func registerTools(server *mcp.Server) {
	// Register current_datetime tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "current_datetime",
		Description: "Returns the current date and time as a string",
	}, handleCurrentDateTime)

	// Register time_difference tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "time_difference",
		Description: "Calculate the time difference between two timestamps",
	}, handleTimeDifference)

	// Register time_since tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "time_since",
		Description: "Calculate time elapsed since a given timestamp until now",
	}, handleTimeSince)

	// Register parse_timestamp tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "parse_timestamp",
		Description: "Parse and convert a timestamp to multiple formats",
	}, handleParseTimestamp)

	// Register add_time tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_time",
		Description: "Add a duration to a timestamp",
	}, handleAddTime)

	// Register timestamp_context tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "timestamp_context",
		Description: "Provide contextual information about a timestamp",
	}, handleTimestampContext)

	// Register format_duration tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "format_duration",
		Description: "Format a duration in seconds into human-readable text",
	}, handleFormatDuration)
}

// Tool handlers
func handleCurrentDateTime(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[CurrentDateTimeArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	timezone := params.Arguments.Timezone
	if timezone == "" {
		timezone = defaultTimezone
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("unknown timezone '%s': %w", timezone, err)
	}

	now := time.Now().In(loc)
	result := now.Format(time.RFC3339)

	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil
}

func handleTimeDifference(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[TimeDifferenceArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	args := params.Arguments
	
	timezone := args.Timezone
	if timezone == "" {
		timezone = defaultTimezone
	}
	
	unit := args.Unit
	if unit == "" {
		unit = "auto"
	}

	t1, err := parseTimestamp(args.Timestamp1, timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp1: %w", err)
	}

	t2, err := parseTimestamp(args.Timestamp2, timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp2: %w", err)
	}

	diff := t2.Sub(t1)
	seconds := diff.Seconds()
	isNegative := seconds < 0
	absSeconds := seconds
	if isNegative {
		absSeconds = -seconds
	}

	result := map[string]interface{}{
		"seconds":     seconds,
		"formatted":   formatDuration(absSeconds, "full", isNegative),
		"is_negative": isNegative,
	}

	if unit != "auto" {
		var divisor float64
		switch unit {
		case "seconds":
			divisor = 1
		case "minutes":
			divisor = 60
		case "hours":
			divisor = 3600
		case "days":
			divisor = 86400
		default:
			return nil, fmt.Errorf("invalid unit: %s", unit)
		}
		result["requested_unit"] = seconds / divisor
	}

	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%v", result)},
		},
	}, nil
}

func handleTimeSince(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[TimeSinceArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	args := params.Arguments
	
	timezone := args.Timezone
	if timezone == "" {
		timezone = defaultTimezone
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("unknown timezone '%s': %w", timezone, err)
	}

	t, err := parseTimestamp(args.Timestamp, timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	now := time.Now().In(loc)
	diff := now.Sub(t)
	seconds := diff.Seconds()
	absSeconds := seconds
	if seconds < 0 {
		absSeconds = -seconds
	}

	// Generate context
	context := getTimeContext(t, now, seconds)

	formatted := formatDuration(absSeconds, "full", false)
	if seconds >= 0 {
		formatted += " ago"
	} else {
		formatted += " from now"
	}

	result := map[string]interface{}{
		"seconds":   seconds,
		"formatted": formatted,
		"context":   context,
	}

	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%v", result)},
		},
	}, nil
}

func handleParseTimestamp(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ParseTimestampArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	args := params.Arguments
	
	targetTimezone := args.TargetTimezone
	if targetTimezone == "" {
		targetTimezone = defaultTimezone
	}

	// Use source timezone if provided, otherwise use target
	parseTz := args.SourceTimezone
	if parseTz == "" {
		parseTz = targetTimezone
	}

	t, err := parseTimestamp(args.Timestamp, parseTz)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Convert to target timezone if different
	if args.SourceTimezone != "" && args.SourceTimezone != targetTimezone {
		loc, err := time.LoadLocation(targetTimezone)
		if err != nil {
			return nil, fmt.Errorf("unknown target timezone '%s': %w", targetTimezone, err)
		}
		t = t.In(loc)
	}

	result := map[string]interface{}{
		"iso":         t.Format(time.RFC3339),
		"unix":        fmt.Sprintf("%d", t.Unix()),
		"human":       t.Format("January 2, 2006 at 3:04 PM MST"),
		"timezone":    targetTimezone,
		"day_of_week": t.Format("Monday"),
		"date":        t.Format("2006-01-02"),
		"time":        t.Format("15:04:05"),
	}

	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%v", result)},
		},
	}, nil
}

func handleAddTime(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[AddTimeArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	args := params.Arguments
	
	timezone := args.Timezone
	if timezone == "" {
		timezone = defaultTimezone
	}

	t, err := parseTimestamp(args.Timestamp, timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Remember if input was date-only
	isDateOnly := len(args.Timestamp) == 10 // YYYY-MM-DD

	// Calculate duration
	var d time.Duration
	switch args.Unit {
	case "seconds":
		d = time.Duration(args.Duration * float64(time.Second))
	case "minutes":
		d = time.Duration(args.Duration * float64(time.Minute))
	case "hours":
		d = time.Duration(args.Duration * float64(time.Hour))
	case "days":
		d = time.Duration(args.Duration * 24 * float64(time.Hour))
	case "weeks":
		d = time.Duration(args.Duration * 7 * 24 * float64(time.Hour))
	default:
		return nil, fmt.Errorf("invalid unit: %s", args.Unit)
	}

	resultTime := t.Add(d)

	// Generate description
	loc, _ := time.LoadLocation(timezone)
	now := time.Now().In(loc)
	description := getTimeDescription(resultTime, now, isDateOnly)

	// Format result to match input format
	var resultStr string
	if isDateOnly {
		resultStr = resultTime.Format("2006-01-02")
	} else {
		resultStr = resultTime.Format("2006-01-02 15:04:05")
	}

	result := map[string]interface{}{
		"result":      resultStr,
		"iso":         resultTime.Format(time.RFC3339),
		"description": description,
	}

	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%v", result)},
		},
	}, nil
}

func handleTimestampContext(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[TimestampContextArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	args := params.Arguments
	
	timezone := args.Timezone
	if timezone == "" {
		timezone = defaultTimezone
	}

	t, err := parseTimestamp(args.Timestamp, timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	loc, _ := time.LoadLocation(timezone)
	now := time.Now().In(loc)

	hour := t.Hour()
	weekday := t.Weekday()

	// Determine time of day
	var timeOfDay string
	switch {
	case hour >= 5 && hour < 9:
		timeOfDay = "early_morning"
	case hour >= 9 && hour < 12:
		timeOfDay = "morning"
	case hour >= 12 && hour < 17:
		timeOfDay = "afternoon"
	case hour >= 17 && hour < 21:
		timeOfDay = "evening"
	default:
		timeOfDay = "late_night"
	}

	// Business hours check
	isWeekend := weekday == time.Saturday || weekday == time.Sunday
	isBusinessHours := !isWeekend && hour >= 9 && hour < 17

	// Typical activity
	var typicalActivity string
	switch {
	case hour >= 6 && hour < 9:
		typicalActivity = "commute_time"
	case hour >= 12 && hour < 13:
		typicalActivity = "lunch_time"
	case hour >= 17 && hour < 19:
		typicalActivity = "commute_time"
	case hour >= 19 && hour < 21:
		typicalActivity = "dinner_time"
	case hour >= 22 || hour < 6:
		typicalActivity = "sleeping_time"
	case isBusinessHours:
		typicalActivity = "work_time"
	default:
		typicalActivity = "leisure_time"
	}

	// Relative day
	daysDiff := int(t.Sub(now).Hours() / 24)
	var relativeDay *string
	switch daysDiff {
	case 0:
		s := "today"
		relativeDay = &s
	case -1:
		s := "yesterday"
		relativeDay = &s
	case 1:
		s := "tomorrow"
		relativeDay = &s
	}

	result := map[string]interface{}{
		"time_of_day":       timeOfDay,
		"day_of_week":       t.Format("Monday"),
		"is_weekend":        isWeekend,
		"is_business_hours": isBusinessHours,
		"hour_24":           hour,
		"typical_activity":  typicalActivity,
		"relative_day":      relativeDay,
	}

	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%v", result)},
		},
	}, nil
}

func handleFormatDuration(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[FormatDurationArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	args := params.Arguments
	
	style := args.Style
	if style == "" {
		style = "full"
	}

	seconds := args.Seconds
	isNegative := seconds < 0
	if isNegative {
		seconds = -seconds
	}

	result := formatDuration(seconds, style, isNegative)

	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil
}