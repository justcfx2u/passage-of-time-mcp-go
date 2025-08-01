// Passage of Time MCP Server - Go implementation
// 
// Based on the original Python implementation by Jérémie Lumbroso and Claude Opus 4.0
// Original repository: https://github.com/jlumbroso/passage-of-time-mcp
//
// Go port by JustCFX2u
// Repository: https://github.com/justcfx2u/passage-of-time-mcp-go

package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "passage-of-time"
	serverVersion = "1.0.0"
)

func main() {
	// Create a new MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	// Register tools
	registerTools(server)

	// Create stdio transport
	t := mcp.NewStdioTransport()

	// Run the server
	if err := server.Run(context.Background(), t); err != nil {
		log.Printf("Server failed: %v", err)
		os.Exit(1)
	}
}