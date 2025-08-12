# Passage of Time MCP Server

A Model Context Protocol (MCP) server for comprehensive time and date operations, providing timezone-aware parsing and manipulation tools for AI assistants.

## Usage

### MCP Configuration

Add to your Claude Desktop (or compatible MCP client) configuration:
```json
{
  "mcpServers": {
    "passage-of-time": {
      "type": "stdio",
      "command": "C:\\Projects\\passage-of-time-mcp-go\\passage-of-time-mcp-windows-amd64.exe",
      "args": [],
      "env": {}
    }
  }
}
```

### Available Tools

- **`current_datetime`** - Get current time in any timezone
- **`parse_timestamp`** - Parse timestamps with 4-layer fallback chain
- **`add_time`** - Add/subtract time durations
- **`time_difference`** - Calculate time between timestamps  
- **`time_since`** - Time elapsed since timestamp
- **`format_duration`** - Human-readable duration formatting
- **`list_timezones`** - Browse timezones with pagination (597 total)

## Building

### Recommended: Use Build Tool
```bash
go run cmd/build/main.go
```

This will:
- Generate latest timezone data from Go's embedded IANA database
- Validate timezone extraction (597 timezones)
- Build cross-platform binaries for Linux, Windows, macOS (AMD64/ARM64)

### Manual Build
```bash
# Generate timezone data first
go run cmd/generate-timezone-list/main.go
cd cmd/list-timezones && go generate

# Build for your platform
go build -o passage-of-time-mcp

# Or cross-platform
GOOS=linux GOARCH=amd64 go build -o passage-of-time-mcp-linux-amd64
```

## Timezone Data Updates

### IANA Timezone Data
- **Source**: Go's embedded IANA timezone database (`$GOROOT/lib/time/zoneinfo.zip`)
- **Update frequency**: 2-3 times per year (matches IANA release schedule)
- **How to update**: Upgrade Go version and rebuild project
- **Scope**: Appropriate for development tools (not mission-critical systems)

### Windows Timezone Mapping
- **Source**: Unicode CLDR (on-demand download)
- **URL**: `https://raw.githubusercontent.com/unicode-org/cldr/refs/heads/main/common/supplemental/windowsZones.xml`
- **How to update**: 
  ```bash
  go run cmd/generate-timezone-mapping/main.go
  ```

### Complete Update Process
```bash
# 1. Update IANA data (upgrade Go, then rebuild)
go run cmd/build/main.go

# 2. Update Windows mapping (optional, as needed)
go run cmd/generate-timezone-mapping/main.go
```

## Development Tools

### Debug Timezone Data
```bash
# List all available timezones
go run cmd/list-timezones/main.go

# Test timezone automation
go run cmd/test-timezone-automation/main.go
```

### Generate Fresh Timezone Data
```bash
# Generate main timezone functions
go run cmd/generate-timezone-list/main.go

# Generate list-timezones data
cd cmd/list-timezones && go generate
```

## Running the Server

```bash
# Linux/macOS
./passage-of-time-mcp-linux-amd64

# Windows
passage-of-time-mcp-windows-amd64.exe
```

## Project Origin

Go port of the Python [passage-of-time-mcp](https://github.com/jlumbroso/passage-of-time-mcp) with enhanced timezone automation and cross-platform support.