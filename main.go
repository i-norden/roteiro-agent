// roteiro-agent is an MCP (Model Context Protocol) server that exposes
// Roteiro's spatial data platform to AI agents like Claude Desktop, VS Code,
// and Cursor.
//
// Usage:
//
//	roteiro-agent --server-url http://localhost:8080 --api-key roteiro_abc123
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/i-norden/roteiro-agent/mcp"
)

var version = "dev"
var newServerRunner = func(client *mcp.Client) interface{ Run() error } {
	return mcp.NewServer(client)
}

var errServerURLRequired = errors.New("server url required")

func main() {
	if err := run(os.Args[1:], os.Getenv, os.Stdout); err != nil {
		if errors.Is(err, errServerURLRequired) {
			log.Fatal("--server-url or ROTEIRO_SERVER_URL is required")
		}
		log.Fatalf("server error: %v", err)
	}
}

func run(args []string, getenv func(string) string, stdout io.Writer) error {
	fs := flag.NewFlagSet("roteiro-agent", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	serverURL := fs.String("server-url", "", "Roteiro server base URL (e.g. http://localhost:8080)")
	apiKey := fs.String("api-key", "", "Roteiro API key for authentication")
	sessionCookie := fs.String("session-cookie", "", "Session cookie value for authentication (alternative to API key)")
	showVersion := fs.Bool("version", false, "Print version and exit")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		_, err := fmt.Fprintf(stdout, "roteiro-agent %s\n", version)
		return err
	}

	if *serverURL == "" {
		// Fall back to environment variable.
		*serverURL = getenv("ROTEIRO_SERVER_URL")
	}
	if *apiKey == "" {
		*apiKey = getenv("ROTEIRO_API_KEY")
	}
	if *sessionCookie == "" {
		*sessionCookie = getenv("ROTEIRO_SESSION_COOKIE")
	}

	if *serverURL == "" {
		return errServerURLRequired
	}

	client := mcp.NewClient(*serverURL, *apiKey)
	if *sessionCookie != "" {
		client.SessionCookie = *sessionCookie
	}

	return newServerRunner(client).Run()
}
