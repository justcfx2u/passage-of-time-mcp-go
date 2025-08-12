package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/justcfx2u/passage-of-time-mcp-go/passageoftime"
)

const defaultTimezone = "UTC"

// Tool argument structs
type CurrentDateTimeArgs struct {
	Timezone                     string `json:"timezone,omitempty" mcp:"Timezone name (e.g., 'UTC', 'US/Pacific'). Defaults to 'UTC'."`
	AutodetectAndUseUserTimezone bool   `json:"autodetect_and_use_user_timezone,omitempty" mcp:"If true, automatically detect and use the system's local timezone instead of UTC default."`
}

type TimeDifferenceArgs struct {
	Timestamp1                   string `json:"timestamp1" mcp:"First timestamp: standard formats (YYYY-MM-DD HH:MM:SS), durations (-14d, 2h30m, -5s), natural language ('tomorrow at 3pm'), or dateparse formats"`
	Timestamp2                   string `json:"timestamp2" mcp:"Second timestamp: standard formats (YYYY-MM-DD HH:MM:SS), durations (1w, -2M, 3y), natural language ('next Monday'), or dateparse formats"`
	Unit                        string `json:"unit,omitempty" mcp:"Desired unit: auto, seconds, minutes, hours, days"`
	Timezone                    string `json:"timezone,omitempty" mcp:"Timezone for parsing ambiguous timestamps"`
	AutodetectAndUseUserTimezone bool   `json:"autodetect_and_use_user_timezone,omitempty" mcp:"If true, automatically detect and use the system's local timezone instead of UTC default."`
	EnableFuzzyParsing          bool   `json:"enable_fuzzy_parsing,omitempty" mcp:"If true, enables 4-layer parsing: 1) durations (-14d, 2h30m), 2) dateparse formats, 3) natural language ('tomorrow'), 4) fallback. Supports EN, RU, PT_BR, ZH, NL."`
}

type TimeSinceArgs struct {
	Timestamp                    string `json:"timestamp" mcp:"Past timestamp: standard formats, durations (-3d, -2h30m), natural language ('3 days ago'), or dateparse formats"`
	Timezone                     string `json:"timezone,omitempty" mcp:"Timezone for parsing and current time"`
	AutodetectAndUseUserTimezone bool   `json:"autodetect_and_use_user_timezone,omitempty" mcp:"If true, automatically detect and use the system's local timezone instead of UTC default."`
	EnableFuzzyParsing          bool   `json:"enable_fuzzy_parsing,omitempty" mcp:"If true, enables 4-layer parsing: 1) durations (-1w, -24h), 2) dateparse formats, 3) natural language ('yesterday'), 4) fallback. Supports EN, RU, PT_BR, ZH, NL."`
}

type ParseTimestampArgs struct {
	Timestamp                    string `json:"timestamp" mcp:"Timestamp: standard formats, durations (1d, -2h30m, 5w), natural language ('tomorrow at 3pm'), or dateparse formats"`
	SourceTimezone              string `json:"source_timezone,omitempty" mcp:"Timezone of the input (if None, uses target_timezone)"`
	TargetTimezone              string `json:"target_timezone,omitempty" mcp:"Desired output timezone"`
	AutodetectAndUseUserTimezone bool   `json:"autodetect_and_use_user_timezone,omitempty" mcp:"If true, automatically detect and use the system's local timezone instead of UTC default."`
	EnableFuzzyParsing          bool   `json:"enable_fuzzy_parsing,omitempty" mcp:"If true, enables 4-layer parsing: 1) durations (1d, -2h), 2) dateparse formats, 3) natural language ('next Monday'), 4) fallback. Supports EN, RU, PT_BR, ZH, NL."`
}

type AddTimeArgs struct {
	Timestamp                    string  `json:"timestamp" mcp:"Starting timestamp: standard formats, durations (-1w, 3d, 2h30m), natural language ('tomorrow'), or dateparse formats"`
	Duration                     float64 `json:"duration" mcp:"Amount to add (can be negative to subtract)"`
	Unit                        string  `json:"unit" mcp:"Unit: seconds, minutes, hours, days, weeks"`
	Timezone                    string  `json:"timezone,omitempty" mcp:"Timezone for calculations"`
	AutodetectAndUseUserTimezone bool    `json:"autodetect_and_use_user_timezone,omitempty" mcp:"If true, automatically detect and use the system's local timezone instead of UTC default."`
	EnableFuzzyParsing          bool    `json:"enable_fuzzy_parsing,omitempty" mcp:"If true, enables 4-layer parsing: 1) durations (7d, -1M, 2y), 2) dateparse formats, 3) natural language ('yesterday'), 4) fallback. Supports EN, RU, PT_BR, ZH, NL."`
}

type TimestampContextArgs struct {
	Timestamp                    string `json:"timestamp" mcp:"Timestamp to analyze: standard formats, durations (-6M, 1y, 30d), natural language ('end of month'), or dateparse formats"`
	Timezone                     string `json:"timezone,omitempty" mcp:"Timezone for context"`
	AutodetectAndUseUserTimezone bool   `json:"autodetect_and_use_user_timezone,omitempty" mcp:"If true, automatically detect and use the system's local timezone instead of UTC default."`
	EnableFuzzyParsing          bool   `json:"enable_fuzzy_parsing,omitempty" mcp:"If true, enables 4-layer parsing: 1) durations (-1y, 6M, 90d), 2) dateparse formats, 3) natural language ('next month'), 4) fallback. Supports EN, RU, PT_BR, ZH, NL."`
}

type FormatDurationArgs struct {
	Seconds float64 `json:"seconds" mcp:"Duration in seconds (can be negative)"`
	Style   string  `json:"style,omitempty" mcp:"Format style: full, compact, minimal"`
}

type ListTimezonesArgs struct {
	Filter  string `json:"filter,omitempty" mcp:"Optional filter to search timezones by region, city, or IANA ID (e.g., 'America', 'London', 'Pacific')"`
	Limit   int    `json:"limit,omitempty" mcp:"Maximum number of timezones to return (default: 25 popular timezones, max: 100 per page)"`
	Page    int    `json:"page,omitempty" mcp:"Page number for pagination (1-based, default: 1). Use with limit to paginate through all 597+ timezones"`
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

	// Register list_timezones tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_timezones",
		Description: "List IANA timezone identifiers with pagination. Returns 25 popular timezones by default. Use limit (max 100) and page parameters to paginate through all 597+ timezones. Supports filtering by region/city.",
	}, handleListTimezones)
}

// Tool handlers
func handleCurrentDateTime(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[CurrentDateTimeArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	timezone := params.Arguments.Timezone
	if timezone == "" {
		if params.Arguments.AutodetectAndUseUserTimezone {
			timezone = passageoftime.GetSystemTimezone()
		} else {
			timezone = defaultTimezone
		}
	}

	options := passageoftime.ParseOptions{
		Timezone: timezone,
	}

	result, err := passageoftime.CurrentDateTime(options)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Timestamp.Format(time.RFC3339)},
		},
	}, nil
}

func handleTimeDifference(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[TimeDifferenceArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	args := params.Arguments
	
	timezone := args.Timezone
	if timezone == "" {
		if args.AutodetectAndUseUserTimezone {
			timezone = passageoftime.GetSystemTimezone()
		} else {
			timezone = defaultTimezone
		}
	}
	
	unit := args.Unit
	if unit == "" {
		unit = "auto"
	}

	// Use passageoftime library for parsing and calculation
	options := passageoftime.ParseOptions{
		EnableFuzzyParsing: args.EnableFuzzyParsing,
		Timezone:           timezone,
		ReferenceTime:      time.Now(),
	}

	durationResult, err := passageoftime.TimeDifference(args.Timestamp1, args.Timestamp2, options)
	if err != nil {
		return nil, err
	}

	seconds := durationResult.Duration
	isNegative := seconds < 0
	
	result := map[string]interface{}{
		"seconds":     seconds,
		"formatted":   durationResult.PreciseDescription,
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
		if args.AutodetectAndUseUserTimezone {
			timezone = passageoftime.GetSystemTimezone()
		} else {
			timezone = defaultTimezone
		}
	}

	// Use passageoftime library for parsing and calculation
	options := passageoftime.ParseOptions{
		EnableFuzzyParsing: args.EnableFuzzyParsing,
		Timezone:           timezone,
		ReferenceTime:      time.Now(),
	}

	durationResult, err := passageoftime.TimeSince(args.Timestamp, options)
	if err != nil {
		return nil, err
	}

	seconds := durationResult.Duration

	// Generate context using library function
	context := passageoftime.GetTimeContext(durationResult.StartTime, durationResult.EndTime, seconds)

	result := map[string]interface{}{
		"seconds":   seconds,
		"formatted": durationResult.PreciseDescription,
		"context":   context,
		"timezone":  timezone,
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
		if args.AutodetectAndUseUserTimezone {
			targetTimezone = passageoftime.GetSystemTimezone()
		} else {
			targetTimezone = defaultTimezone
		}
	}

	// Use source timezone if provided, otherwise use target
	parseTz := args.SourceTimezone
	if parseTz == "" {
		parseTz = targetTimezone
	}

	// Use passageoftime library for parsing
	options := passageoftime.ParseOptions{
		EnableFuzzyParsing: args.EnableFuzzyParsing,
		Timezone:           parseTz,
		ReferenceTime:      time.Now(),
	}

	t, err := passageoftime.ParseFuzzyTimestamp(args.Timestamp, options)
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
		"iso":                t.Format(time.RFC3339),
		"unix":               fmt.Sprintf("%d", t.Unix()),
		"human":              t.Format("January 2, 2006 at 3:04 PM MST"),
		"timezone":           targetTimezone,
		"day_of_week":        t.Format("Monday"),
		"date":               t.Format("2006-01-02"),
		"time":               t.Format("15:04:05"),
		"source_timezone":    parseTz,
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
		if args.AutodetectAndUseUserTimezone {
			timezone = passageoftime.GetSystemTimezone()
		} else {
			timezone = defaultTimezone
		}
	}

	// Use passageoftime library for parsing
	options := passageoftime.ParseOptions{
		EnableFuzzyParsing: args.EnableFuzzyParsing,
		Timezone:           timezone,
		ReferenceTime:      time.Now(),
	}

	t, err := passageoftime.ParseFuzzyTimestamp(args.Timestamp, options)
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

	// Generate description using library function
	loc, _ := time.LoadLocation(timezone)
	now := time.Now().In(loc)
	description := passageoftime.GetTimeDescription(resultTime, now, isDateOnly)

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
		if args.AutodetectAndUseUserTimezone {
			timezone = passageoftime.GetSystemTimezone()
		} else {
			timezone = defaultTimezone
		}
	}

	// Use passageoftime library for parsing
	options := passageoftime.ParseOptions{
		EnableFuzzyParsing: args.EnableFuzzyParsing,
		Timezone:           timezone,
		ReferenceTime:      time.Now(),
	}

	t, err := passageoftime.ParseFuzzyTimestamp(args.Timestamp, options)
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

	// Use passageoftime library for duration formatting
	durationFormatted := passageoftime.FormatDuration(seconds, style, isNegative)
	
	// Add precise timestamp like other handlers
	now := time.Now()
	targetTime := now.Add(time.Duration(seconds) * time.Second)
	if isNegative {
		targetTime = now.Add(-time.Duration(seconds) * time.Second)
	}
	
	result := passageoftime.FormatWithPreciseTimestamp(durationFormatted, targetTime, "UTC")

	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil
}

func handleListTimezones(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListTimezonesArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	args := params.Arguments
	
	// Set defaults and validate pagination parameters
	limit := args.Limit
	page := args.Page
	usePopularDefault := limit == 0 && args.Filter == "" && page == 0
	
	if limit == 0 {
		if usePopularDefault {
			limit = 25 // Popular timezones default
		} else {
			limit = 100 // Default page size for pagination
		}
	} else if limit > 100 {
		limit = 100 // Max per page for performance
	}
	
	if page == 0 {
		page = 1 // Default to first page
	}
	
	// Get timezone identifiers - use popular list by default, all when filtering/pagination
	var sourceTimezones []string
	if usePopularDefault {
		sourceTimezones = passageoftime.GetPopularTimezoneIDs()
	} else {
		sourceTimezones = passageoftime.GetAllTimezoneIDs()
	}
	
	// Apply filter if provided
	var filteredTimezones []string
	filter := strings.ToLower(args.Filter)
	
	for _, tz := range sourceTimezones {
		if filter == "" || strings.Contains(strings.ToLower(tz), filter) {
			filteredTimezones = append(filteredTimezones, tz)
		}
	}
	
	// Apply pagination
	totalCount := len(filteredTimezones)
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	
	if startIndex >= totalCount {
		filteredTimezones = []string{} // Empty page
	} else {
		if endIndex > totalCount {
			endIndex = totalCount
		}
		filteredTimezones = filteredTimezones[startIndex:endIndex]
	}
	
	// Build result with timezone info
	timezoneInfos := make([]map[string]interface{}, len(filteredTimezones))
	now := time.Now()
	
	for i, tzID := range filteredTimezones {
		loc, err := time.LoadLocation(tzID)
		if err != nil {
			continue // Skip invalid timezones
		}
		
		nowInTz := now.In(loc)
		_, offset := nowInTz.Zone()
		offsetHours := float64(offset) / 3600
		
		timezoneInfos[i] = map[string]interface{}{
			"id":          tzID,
			"name":        nowInTz.Location().String(),
			"offset":      offsetHours,
			"offset_str":  passageoftime.FormatOffset(offset),
			"current_time": nowInTz.Format("2006-01-02 15:04:05 MST"),
		}
	}
	
	// Calculate pagination metadata
	actualReturned := len(timezoneInfos)
	totalPages := (totalCount + limit - 1) / limit // Ceiling division
	hasNextPage := page < totalPages
	hasPrevPage := page > 1
	
	// Always show total available from full list for reference
	allTimezones := passageoftime.GetAllTimezoneIDs()
	
	result := map[string]interface{}{
		"total_available": len(allTimezones),
		"total_filtered":  totalCount,
		"returned_count":  actualReturned,
		"page":           page,
		"limit":          limit,
		"total_pages":    totalPages,
		"has_next_page":  hasNextPage,
		"has_prev_page":  hasPrevPage,
		"filter":         args.Filter,
		"using_popular":  usePopularDefault,
		"timezones":      timezoneInfos,
	}
	
	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%v", result)},
		},
	}, nil
}