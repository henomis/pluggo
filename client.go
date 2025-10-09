// Package pluggo provides a framework for creating and communicating with plugin systems
// using HTTP-based communication. It allows for dynamic loading and execution of
// plugins as separate processes that communicate over HTTP.
//
// The package consists of two main components:
// - Client: for launching and communicating with plugins
// - Plugin: for creating plugins that can be launched by the client
//
// Plugins are executable files that start an HTTP server and communicate their
// port number to the client via stdout. The client then uses HTTP requests to
// execute functions within the plugin.
package pluggo

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	defaultSchema = "http://"
	defaultHost   = "127.0.0.1"
	schemasPath   = "/_schemas"
	healthPath    = "/_healthz"

	// DefaultFunctionExecutionTimeout is the HTTP timeout for requests the launcher makes to the plugin (health + exec)
	DefaultFunctionExecutionTimeout = 2 * time.Minute
	// DefaultHealthCheckTimeout is the total time the launcher will wait for the plugin to become healthy
	DefaultHealthCheckTimeout = 5 * time.Second
	// DefaultHealthCheckInterval defines how often to retry hitting /_healthz while waiting
	DefaultHealthCheckInterval = 150 * time.Millisecond
)

// Connection represents an active HTTP connection to a plugin server.
// It contains the base URL and configuration for communication with the plugin.
type Connection struct {
	FunctionExecutionTimeout time.Duration
	BaseURL                  string
}

// Client manages the lifecycle and communication with a plugin process.
// It handles launching the plugin executable, establishing HTTP communication,
// health checking, and graceful shutdown.
type Client struct {
	path                     string
	functionExecutionTimeout time.Duration
	healthCheckTimeout       time.Duration
	healthCheckInterval      time.Duration
	heartbeatInterval        time.Duration
	heartbeatChan            chan struct{}

	httpClient     *http.Client
	connection     *Connection
	commandContext *exec.Cmd
	cancel         context.CancelFunc
}

// ClientOption is a function that configures a Client during creation.
type ClientOption func(*Client)

// WithFunctionExecutionTimeout sets the HTTP request timeout for function execution calls.
func WithFunctionExecutionTimeout(timeout time.Duration) ClientOption {
	return func(p *Client) {
		p.functionExecutionTimeout = timeout
	}
}

// WithHealthCheckTimeout sets the total timeout duration for waiting for the plugin to become healthy.
func WithHealthCheckTimeout(timeout time.Duration) ClientOption {
	return func(p *Client) {
		p.healthCheckTimeout = timeout
	}
}

// WithHealthCheckInterval sets the interval between health check attempts during plugin startup.
func WithHealthCheckInterval(interval time.Duration) ClientOption {
	return func(p *Client) {
		p.healthCheckInterval = interval
	}
}

// WithHeartbeatInterval sets the interval between heartbeat checks for the plugin.
func WithHeartbeatInterval(interval time.Duration) ClientOption {
	return func(p *Client) {
		p.heartbeatInterval = interval
	}
}

// New creates a new Client instance with the specified plugin path and optional configuration.
// The path should point to an executable file that implements the plugin protocol.
// Options can be provided to customize timeouts and other behavior.
func New(path string, opts ...ClientOption) *Client {
	p := &Client{
		path:                     path,
		functionExecutionTimeout: DefaultFunctionExecutionTimeout,
		healthCheckTimeout:       DefaultHealthCheckTimeout,
		healthCheckInterval:      DefaultHealthCheckInterval,
		heartbeatInterval:        0,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Open launches the plugin process and establishes communication.
// It performs the following steps:
// 1. Validates that the plugin file exists and is executable
// 2. Starts the plugin process
// 3. Reads the HTTP port from the plugin's stdout
// 4. Establishes HTTP connection and waits for the plugin to become healthy
//
// Returns an error if any step fails. The plugin process will be terminated
// automatically if initialization fails.
func (c *Client) Open(ctx context.Context) error {
	if c.commandContext != nil {
		return errors.New("plugin is already running")
	}

	fileInfo, err := os.Stat(c.path)
	if err != nil || fileInfo.IsDir() {
		return &PluginNotFoundError{Err: err}
	}

	fileIsExecutable := fileInfo.Mode()&0111 != 0

	if !fileIsExecutable {
		return &PluginExecutionError{Err: errors.New("plugin must be an executable")}
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	commandContext := exec.CommandContext(cancelCtx, c.path)
	stdout, _ := commandContext.StdoutPipe()
	commandContext.Stderr = os.Stderr

	if err := commandContext.Start(); err != nil {
		_ = c.Close()
		return &PluginExecutionError{Err: err}
	}
	c.commandContext = commandContext

	// Read port from plugin's stdout
	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	if err != nil {
		_ = c.Close()
		return &PluginExecutionError{Err: err}
	}

	pluginPort := strings.TrimSpace(line)
	_, err = strconv.Atoi(pluginPort)
	if err != nil {
		_ = c.Close()
		return &PluginExecutionError{Err: fmt.Errorf("invalid port received from plugin: %s", pluginPort)}
	}

	c.connection = &Connection{
		BaseURL: fmt.Sprintf("%s%s:%s", defaultSchema, defaultHost, pluginPort),
	}

	c.httpClient = &http.Client{Timeout: c.functionExecutionTimeout}
	if err := c.waitForHealth(); err != nil {
		_ = c.Close()
		return &PluginExecutionError{Err: err}
	}

	if c.heartbeatInterval > 0 {
		c.heartbeatChan = make(chan struct{})
		go func() {
			ticker := time.NewTicker(c.heartbeatInterval)
			defer ticker.Stop()

			for range ticker.C {
				err := c.waitForHealth()
				if err != nil {
					_ = c.Close()
					return
				}
			}
		}()
	}

	return nil
}

// Done returns a channel that signals the health status of the plugin.
func (c *Client) Done() <-chan struct{} {
	return c.heartbeatChan
}

// Close gracefully shuts down the plugin process and cleans up resources.
// It cancels the plugin's context and kills the process if it's still running.
// This method is safe to call multiple times.
func (c *Client) Close() error {
	defer func() {
		c.commandContext = nil
		c.cancel = nil
		if c.heartbeatChan != nil {
			close(c.heartbeatChan)
			c.heartbeatChan = nil
		}
		c.connection = nil
		c.httpClient = nil
	}()

	if c.cancel != nil {
		c.cancel()
	}

	if c.commandContext != nil && c.commandContext.Process != nil {
		return c.commandContext.Process.Kill()
	}

	return nil
}

// Connection returns the current HTTP connection details for the plugin.
// Returns nil if the plugin is not currently running or connected.
func (c *Client) Connection() *Connection {
	return c.connection
}

// Schemas retrieves the list of available functions and their input/output schemas
// from the plugin. This provides introspection capabilities to understand what
// functions are available and their expected data structures.
func (c *Client) Schemas() (Schemas, error) {
	if c.connection == nil {
		return nil, errors.New("plugin is not connected")
	}

	resp, err := c.httpClient.Get(c.connection.BaseURL + schemasPath)
	if err != nil {
		return nil, &PluginExecutionError{Err: err}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, &PluginExecutionError{Err: fmt.Errorf("plugin returned status %d", resp.StatusCode)}
	}

	var schemas Schemas
	err = json.NewDecoder(resp.Body).Decode(&schemas)
	if err != nil {
		return nil, &PluginExecutionError{Err: err}
	}

	return schemas, nil
}

// waitForHealth repeatedly checks the plugin's health endpoint until it responds
// successfully or the health check timeout is reached. This ensures the plugin
// is fully initialized before allowing function calls.
func (c *Client) waitForHealth() error {
	deadline := time.Now().Add(c.healthCheckTimeout)

	for {
		if time.Now().After(deadline) {
			return errors.New("timeout waiting for plugin to become healthy")
		}
		resp, err := c.httpClient.Get(c.connection.BaseURL + healthPath)
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			return nil
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(c.healthCheckInterval)
	}
}
