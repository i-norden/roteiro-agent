package main

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/i-norden/roteiro-agent/mcp"
)

type stubRunner struct {
	run func() error
}

func (r stubRunner) Run() error {
	if r.run == nil {
		return nil
	}
	return r.run()
}

func TestRunPrintsVersion(t *testing.T) {
	oldVersion := version
	oldFactory := newServerRunner
	version = "1.2.3"
	newServerRunner = func(*mcp.Client) interface{ Run() error } {
		t.Fatal("server should not start for --version")
		return stubRunner{}
	}
	t.Cleanup(func() {
		version = oldVersion
		newServerRunner = oldFactory
	})

	var out bytes.Buffer
	if err := run([]string{"--version"}, func(string) string { return "" }, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if got := out.String(); got != "roteiro-agent 1.2.3\n" {
		t.Fatalf("version output = %q, want %q", got, "roteiro-agent 1.2.3\n")
	}
}

func TestRunUsesEnvironmentFallbacks(t *testing.T) {
	oldFactory := newServerRunner
	var got struct {
		baseURL       string
		apiKey        string
		sessionCookie string
		projectID     string
	}
	newServerRunner = func(client *mcp.Client) interface{ Run() error } {
		got.baseURL = client.BaseURL
		got.apiKey = client.APIKey
		got.sessionCookie = client.SessionCookie
		got.projectID = client.ProjectID
		return stubRunner{}
	}
	t.Cleanup(func() {
		newServerRunner = oldFactory
	})

	env := map[string]string{
		"ROTEIRO_SERVER_URL":     "https://example.test/",
		"ROTEIRO_API_KEY":        "api-from-env",
		"ROTEIRO_SESSION_COOKIE": "session=from-env",
		"ROTEIRO_PROJECT_ID":     "42",
	}
	if err := run(nil, func(key string) string { return env[key] }, io.Discard); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if got.baseURL != "https://example.test" {
		t.Fatalf("BaseURL = %q, want %q", got.baseURL, "https://example.test")
	}
	if got.apiKey != "api-from-env" {
		t.Fatalf("APIKey = %q, want %q", got.apiKey, "api-from-env")
	}
	if got.sessionCookie != "session=from-env" {
		t.Fatalf("SessionCookie = %q, want %q", got.sessionCookie, "session=from-env")
	}
	if got.projectID != "42" {
		t.Fatalf("ProjectID = %q, want %q", got.projectID, "42")
	}
}

func TestRunFlagsOverrideEnvironment(t *testing.T) {
	oldFactory := newServerRunner
	var got struct {
		baseURL       string
		apiKey        string
		sessionCookie string
		projectID     string
	}
	newServerRunner = func(client *mcp.Client) interface{ Run() error } {
		got.baseURL = client.BaseURL
		got.apiKey = client.APIKey
		got.sessionCookie = client.SessionCookie
		got.projectID = client.ProjectID
		return stubRunner{}
	}
	t.Cleanup(func() {
		newServerRunner = oldFactory
	})

	env := map[string]string{
		"ROTEIRO_SERVER_URL":     "https://env.example.test",
		"ROTEIRO_API_KEY":        "api-from-env",
		"ROTEIRO_SESSION_COOKIE": "session=from-env",
		"ROTEIRO_PROJECT_ID":     "7",
	}
	args := []string{
		"--server-url", "https://flag.example.test/",
		"--api-key", "api-from-flag",
		"--session-cookie", "session=from-flag",
		"--project-id", "99",
	}
	if err := run(args, func(key string) string { return env[key] }, io.Discard); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if got.baseURL != "https://flag.example.test" {
		t.Fatalf("BaseURL = %q, want %q", got.baseURL, "https://flag.example.test")
	}
	if got.apiKey != "api-from-flag" {
		t.Fatalf("APIKey = %q, want %q", got.apiKey, "api-from-flag")
	}
	if got.sessionCookie != "session=from-flag" {
		t.Fatalf("SessionCookie = %q, want %q", got.sessionCookie, "session=from-flag")
	}
	if got.projectID != "99" {
		t.Fatalf("ProjectID = %q, want %q", got.projectID, "99")
	}
}

func TestRunRequiresServerURL(t *testing.T) {
	oldFactory := newServerRunner
	newServerRunner = func(*mcp.Client) interface{ Run() error } {
		t.Fatal("server should not start without server URL")
		return stubRunner{}
	}
	t.Cleanup(func() {
		newServerRunner = oldFactory
	})

	err := run(nil, func(string) string { return "" }, io.Discard)
	if !errors.Is(err, errServerURLRequired) {
		t.Fatalf("run error = %v, want %v", err, errServerURLRequired)
	}
}

func TestRunPropagatesServerError(t *testing.T) {
	oldFactory := newServerRunner
	wantErr := errors.New("boom")
	newServerRunner = func(*mcp.Client) interface{ Run() error } {
		return stubRunner{run: func() error { return wantErr }}
	}
	t.Cleanup(func() {
		newServerRunner = oldFactory
	})

	err := run([]string{"--server-url", "https://example.test"}, func(string) string { return "" }, io.Discard)
	if !errors.Is(err, wantErr) {
		t.Fatalf("run error = %v, want %v", err, wantErr)
	}
}
