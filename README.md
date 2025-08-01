# Passage of Time MCP Server - Go Implementation

A Go implementation of the "Passage of Time" Model Context Protocol (MCP) server, providing temporal awareness and time calculation abilities to language models.

Based on the original Python implementation by [Jérémie Lumbroso](https://github.com/jlumbroso) and Claude Opus 4.0.  
Original repository: https://github.com/jlumbroso/passage-of-time-mcp

## Features

All 7 time-related tools from the original implementation:
- `current_datetime` - Get current time in specified timezone
- `time_difference` - Calculate duration between timestamps
- `time_since` - Calculate elapsed time from timestamp
- `parse_timestamp` - Convert between timestamp formats
- `add_time` - Add/subtract time from timestamp
- `timestamp_context` - Get human context about a timestamp
- `format_duration` - Format duration in various styles

## Installation

### Option 1: Download Pre-built Binaries (Recommended)

Download the latest release for your platform from the [GitHub Releases](https://github.com/justcfx2u/passage-of-time-mcp-go/releases) page:

- **Windows**: `passage-of-time-windows-amd64.exe`
- **Linux**: `passage-of-time-linux-amd64` 
- **macOS**: `passage-of-time-darwin-amd64`
- **ARM64 variants**: Available for Linux and macOS

### Option 2: Build from Source

**Prerequisites:** Go 1.21 or higher ([installation guide](https://go.dev/doc/install))

```bash
git clone https://github.com/justcfx2u/passage-of-time-mcp-go
cd passage-of-time-mcp-go
```

#### Simple Build (Recommended)

**Windows PowerShell:**
```powershell
$env:CGO_ENABLED=0; go build -o passage-of-time.exe
```

**Windows Git Bash/MINGW:**
```bash
CGO_ENABLED=0 go build -o passage-of-time.exe
```

**Linux/macOS:**
```bash
CGO_ENABLED=0 go build -o passage-of-time
```

#### About CGO_ENABLED

**CGO_ENABLED=0** (Recommended for this project):
- Pure Go build with no C dependencies
- No additional tools required beyond Go installation
- Smaller, more portable binaries
- See: [Go cgo documentation](https://pkg.go.dev/cmd/cgo)

**CGO_ENABLED=1** (Default, but requires C compiler):
- Enables C interoperability (not needed for this MCP server)
- **Windows**: Requires MinGW-w64 or Visual Studio C++ Build Tools
- **Linux/macOS**: Requires GCC or Clang in PATH
- See: [Go Wiki: cgo](https://go.dev/wiki/cgo) and [Windows Build Requirements](https://go.dev/wiki/WindowsBuild)

If you encounter "cgo: C compiler not found" errors, use `CGO_ENABLED=0`.

## Usage

The server uses stdio transport for communication with MCP clients.

### Claude Code Installation

For complete installation and configuration details for Claude Code and Claude Code MCP configuration settings, see the [Claude Code Quickstart](https://docs.anthropic.com/en/docs/claude-code/quickstart) and [Settings documentation](https://docs.anthropic.com/en/docs/claude-code/settings#settings-files).

### Windows PowerShell (Anthropic Beta)
#### Local (default, not checked into source control)
```powershell
cd C:\path\to\your\project
claude mcp add passage-of-time -- C:\path\to\passage-of-time.exe
```

#### Project-specific (shared with team)
```powershell
cd C:\path\to\your\project
claude mcp add --scope project passage-of-time -- C:\path\to\passage-of-time.exe
```

#### User-wide (global)
```powershell
claude mcp add --scope user passage-of-time -- C:\path\to\passage-of-time.exe
```

### Git Bash/MINGW, WSL, Linux, macOS
#### Local (default, not checked into source control)
```bash
cd /path/to/your/project
claude mcp add passage-of-time -- /path/to/passage-of-time
```

#### Project-specific (shared with team)
```bash
cd /path/to/your/project
claude mcp add --scope project passage-of-time -- /path/to/passage-of-time
```

#### User-wide (global)
```bash
claude mcp add --scope user passage-of-time -- /path/to/passage-of-time
```

### Claude Desktop and Other MCP-Compatible Agents

1. **Claude Desktop**: Go to **File > Settings > Desktop app > Developer > Edit config**
2. Add to the `mcpServers` section in `claude_desktop_config.json` or equivalent MCP client configuration for your agent:

#### Windows Native
```json
{
  "mcpServers": {
    "passage-of-time": {
      "command": "C:\\Projects\\passage-of-time-mcp-go\\passage-of-time.exe"
    }
  }
}
```

#### Git Bash/MINGW64 on Windows
```json
{
  "mcpServers": {
    "passage-of-time": {
      "command": "/c/Projects/passage-of-time-mcp-go/passage-of-time.exe"
    }
  }
}
```

#### WSL (Windows Subsystem for Linux)
```json
{
  "mcpServers": {
    "passage-of-time": {
      "command": "/home/username/passage-of-time-mcp-go/passage-of-time"
    }
  }
}
```

#### Linux/macOS
```json
{
  "mcpServers": {
    "passage-of-time": {
      "command": "/path/to/passage-of-time"
    }
  }
}
```

**Important Notes:**
- **Windows Native**: Requires escaped backslashes (`\\`) in JSON
- **Git Bash/MINGW**: Uses Unix-style paths with Windows drive letters (`/c/`)
- **WSL**: Uses native Linux filesystem paths with Linux binary (no `.exe`)
- **Linux/macOS**: Standard Unix paths without `.exe` extension
- **Other MCP Clients**: Configuration format may vary by client, but command paths follow the same patterns

### Cursor Installation

Cursor uses a similar MCP configuration format. Click **Add Custom MCP** in Cursor settings to open `.cursor/mcp.json` (project-local) or configure globally in `~/.cursor/mcp.json`. The `mcpServers` section follows the same JSON structure as Claude Desktop above.

For detailed setup instructions, see: https://docs.cursor.com/en/context/mcp#installing-mcp-servers

## Why Go Implementation?

This Go port offers key advantages over the original Python implementation:

1. **Deployment**: Single binary with no runtime dependencies (no Python environment needed)
2. **Cross-platform**: Easy distribution across Windows, Linux, and macOS
3. **Modern Transport**: Uses stdio transport (current MCP standard) instead of deprecated SSE transport

## Timestamp Formats

All timestamps must use one of these formats:
- Full: `YYYY-MM-DD HH:MM:SS` (e.g., "2024-01-15 14:30:45")
- Date only: `YYYY-MM-DD` (e.g., "2024-01-15")

## Development

### Project Structure

```
passage-of-time-mcp-go/
├── main.go         # Entry point
├── server.go       # Tool registration
├── handlers.go     # Tool implementations
├── time_utils.go   # Time manipulation utilities
├── go.mod          # Go module definition
└── README.md       # This file
```

### Running tests

```bash
# Run all tests in current directory
go test

# Run all tests recursively (current directory and subdirectories)
go test ./...

# Run specific test file
go test handlers_test.go

# Run specific test function
go test -run TestHandleCurrentDateTime
```

For more testing options, see the [Go test documentation](https://pkg.go.dev/cmd/go#hdr-Test_packages).

### Building

**Windows:**
```powershell
go build -o passage-of-time.exe
```

**Linux/macOS:**
```bash
go build -o passage-of-time
```

## License

Mozilla Public License 2.0. See [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original concept and Python implementation by [Jérémie Lumbroso](https://github.com/jlumbroso) and Claude Opus 4.0
- Built using the [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk) v0.2.0
- Go port created with assistance from Claude (Anthropic)
- Automated releases via GitHub Actions