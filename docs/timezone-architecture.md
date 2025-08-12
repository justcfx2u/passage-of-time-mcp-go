# Timezone System Architecture - Sequence Diagram

## How Timezone Automation Works in Passage of Time MCP Server

```mermaid
sequenceDiagram
    participant User as AI Assistant (Claude)
    participant MCP as MCP Server
    participant Handler as Tool Handler
    participant Parser as Fuzzy Parser
    participant TZ as Timezone Utils
    participant Build as Build Tool (go:generate)
    participant Go as Go Runtime

    Note over User,Go: Current Timezone System Flow

    %% Build Time (go:generate)
    Note over Build,Go: Build-Time Extraction (go:generate)
    Build->>Go: Extract from GOROOT/lib/time/zoneinfo.zip
    Go-->>Build: 597 timezone identifiers
    Build->>TZ: Generate getAllTimezoneIDs() function
    Build->>TZ: Generate getPopularTimezones() function
    Note over Build: Execution: go generate ./cmd/list-timezones

    %% Runtime Flow
    Note over User,Go: Runtime Usage Flow

    User->>MCP: MCP tool call (e.g., parse_timestamp)
    MCP->>Handler: Route to appropriate handler
    
    alt Timezone Detection Needed
        Handler->>TZ: getSystemTimezone() if autodetect=true
        TZ->>Go: Detect system timezone
        Go-->>TZ: Local timezone (e.g., "America/New_York")
        TZ-->>Handler: System timezone
    else Manual Timezone
        Handler->>Handler: Use provided timezone parameter
    end

    Handler->>Parser: parseFuzzyTimestamp(input, ref, fuzzy, timezone)
    
    Note over Parser: 4-Layer Parsing Chain
    Parser->>Parser: Layer 1: Duration parsing (-14d, 2h30m)
    alt Duration format detected
        Parser-->>Handler: Parsed time
    else Standard format
        Parser->>Go: dateparse.ParseIn(input, location)
        Go-->>Parser: Parsed time or error
        alt Success
            Parser-->>Handler: Parsed time
        else NLP needed
            Parser->>Parser: Layer 3: NLP parsing (when library)
            alt NLP success
                Parser-->>Handler: Parsed time
            else Fallback needed
                Parser->>TZ: parseTimestamp(input, timezone)
                TZ->>Go: time.LoadLocation(timezone)
                Go-->>TZ: Location object
                TZ->>Go: time.ParseInLocation(format, input, loc)
                Go-->>TZ: Parsed time
                TZ-->>Parser: Parsed time
                Parser-->>Handler: Parsed time
            end
        end
    end

    Handler->>TZ: Format result for MCP response
    TZ->>Go: time operations (format, convert timezone)
    Go-->>TZ: Formatted result
    TZ-->>Handler: MCP response data
    Handler-->>MCP: Tool response
    MCP-->>User: JSON response with timezone info

    %% Timezone List Operations
    Note over User,Go: Timezone Listing Flow
    User->>MCP: list_timezones tool call
    MCP->>Handler: handleListTimezones
    
    alt Popular timezones (default)
        Handler->>TZ: getPopularTimezones()
        TZ-->>Handler: 25 popular timezone IDs
    else Full list or filtered
        Handler->>TZ: getAllTimezoneIDs()
        TZ-->>Handler: 597 timezone IDs
        Handler->>Handler: Apply filter if provided
    end

    Handler->>Go: time.LoadLocation() for each timezone
    Go-->>Handler: Current time + offset for each
    Handler-->>MCP: Timezone list with current times
    MCP-->>User: JSON response with timezone data
```

## Key Components Explained

### Build-Time Generation (go:generate Integration)
- **Tool**: `cmd/generate-timezone-list/main.go`
- **Source**: Go's embedded `zoneinfo.zip` 
- **Output**: Functions in `cmd/list-timezones/timezones.go`
- **Status**: ✅ Complete go:generate implementation
- **Command**: `go generate ./cmd/list-timezones`

### Runtime Integration
- **Entry Point**: MCP tool handlers in `server.go`
- **Parsing**: 4-layer fuzzy parsing chain
- **Timezone Resolution**: Cross-platform system detection
- **Output**: Formatted JSON responses

### Architecture Benefits
- **Build-time Validation**: Catches timezone data mismatches before deployment
- **Platform Independence**: Generation process works across Windows/Unix
- **Future-proof**: Automatically adapts to Go version timezone updates
- **No Circular Imports**: Standalone function generation avoids package dependencies

## Implementation Status

| Component | Status | Details |
|-----------|--------|---------|
| Extraction Tool | ✅ Complete | Works (597 timezones) |
| go:generate Integration | ✅ Complete | `//go:generate go run ../generate-timezone-list -output timezones.go` |
| Runtime Functions | ✅ Working | All tests passing |
| Cross-platform Support | ✅ Complete | Windows/Unix timezone detection |
| Generated Code | ✅ Gitignored | `cmd/list-timezones/timezones.go` excluded from version control |

*Last Updated: 2025-08-17 10:29 EST*