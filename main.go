// roteiro-agent is an MCP (Model Context Protocol) server that exposes
// Roteiro's spatial data platform to AI agents like Claude Desktop, VS Code,
// and Cursor.
//
// Usage:
//
//	roteiro-agent --server-url http://localhost:8080 --api-key roteiro_abc123
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/i-norden/roteiro-agent/mcp"
)

var version = "dev"

func main() {
	serverURL := flag.String("server-url", "", "Roteiro server base URL (e.g. http://localhost:8080)")
	apiKey := flag.String("api-key", "", "Roteiro API key for authentication")
	sessionCookie := flag.String("session-cookie", "", "Session cookie value for authentication (alternative to API key)")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("roteiro-agent %s\n", version)
		os.Exit(0)
	}

	if *serverURL == "" {
		// Fall back to environment variable.
		*serverURL = os.Getenv("ROTEIRO_SERVER_URL")
	}
	if *apiKey == "" {
		*apiKey = os.Getenv("ROTEIRO_API_KEY")
	}
	if *sessionCookie == "" {
		*sessionCookie = os.Getenv("ROTEIRO_SESSION_COOKIE")
	}

	if *serverURL == "" {
		log.Fatal("--server-url or ROTEIRO_SERVER_URL is required")
	}

	client := mcp.NewClient(*serverURL, *apiKey)
	if *sessionCookie != "" {
		client.SessionCookie = *sessionCookie
	}

	server := mcp.NewServer(client)
	if err := server.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
