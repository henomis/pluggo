package pluggo

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	basePath = "/"
)

// Schema represents the input and output JSON schemas for a plugin function.
// This provides introspection capabilities for clients to understand
// the expected data structures.
type Schema struct {
	Input  map[string]any `json:"input"`
	Output map[string]any `json:"output"`
}

// Schemas is a map of function names to their corresponding schemas.
type Schemas map[string]Schema

// Plugin represents a plugin server that can host multiple functions.
// It manages the HTTP server, function registration, and provides
// health check and schema introspection endpoints.
type Plugin struct {
	logger     *slog.Logger
	functions  Schemas
	httpServer *http.Server
	mux        *http.ServeMux
}

// NewPlugin creates a new plugin instance with default configuration.
// It sets up the HTTP server, logging, health check endpoint, and schema endpoint.
func NewPlugin() *Plugin {
	mux := http.NewServeMux()

	l := &Plugin{
		mux: mux,
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
		httpServer: &http.Server{
			Handler:     mux,
			ReadTimeout: 5 * time.Second,
		},
		functions: make(map[string]Schema),
	}

	// Liveness/Readiness probe
	mux.HandleFunc(healthPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// List functions
	mux.HandleFunc(schemasPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(l.functions)
		if err != nil {
			l.logger.Error("failed to encode functions list", "error", err)
		}
	})

	return l
}

// AddFunction registers a new function with the plugin server.
// The function becomes available at the endpoint /{functionName} and
// its schema at /{functionName}/_schemas. Function names are validated
// to ensure they contain only safe characters.
func (l *Plugin) AddFunction(functionName string, handler *Handler) {
	if err := validateFunctionName(functionName); err != nil {
		l.logger.Error("invalid function name", "function", functionName, "error", err)
		return
	}

	l.functions[functionName] = handler.Schema
	l.mux.Handle(basePath+functionName, handler.HTTPHandler)
	l.mux.HandleFunc(fmt.Sprintf("%s%s%s", basePath, functionName, schemasPath), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(l.functions[functionName])
		if err != nil {
			l.logger.Error("failed to encode functions list", "error", err)
		}
	})
}

// Start begins serving the plugin on an ephemeral port.
// The port number is printed to stdout as the first line, which allows
// the client to discover how to connect to the plugin.
// This method blocks until the server stops or encounters an error.
func (l *Plugin) Start() error {
	// Bind to an ephemeral port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		l.logger.Error("failed to bind to port", "error", err)
		return err
	}

	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		l.logger.Error("failed to parse port", "error", err)
		return err
	}

	// First line to stdout MUST be the port so the launcher can parse it
	fmt.Println(port)
	_ = os.Stdout.Sync()

	l.httpServer.Addr = ln.Addr().String()
	if err := l.httpServer.Serve(ln); err != nil {
		l.logger.Error("failed to serve HTTP", "error", err)
		return err
	}

	return nil
}

// Stop gracefully shuts down the plugin server and cleans up resources.
// This method is safe to call multiple times.
func (l *Plugin) Stop() {
	defer func() {
		l.httpServer = nil
		l.mux = nil
	}()

	if l.httpServer != nil {
		_ = l.httpServer.Close()
	}
}

// validateFunctionName ensures that function names contain only safe characters
// and meet length requirements. Function names must be URL-safe since they
// become HTTP endpoints.
func validateFunctionName(function string) error {
	if function == "" {
		return errors.New("function name cannot be empty")
	}
	if function[0] == '/' {
		return errors.New("function name cannot start with '/'")
	}
	if len(function) > 128 {
		return errors.New("function name cannot be longer than 128 characters")
	}
	for _, r := range function {
		// nolint:staticcheck // allowed characters
		if !(r == '-' || r == '_' || r == '.' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return fmt.Errorf("function name contains invalid character: %q", r)
		}
	}
	return nil
}
