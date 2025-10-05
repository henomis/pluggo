package pluggo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Function represents a typed function that can be called on a remote plugin.
// T is the input type and R is the output type for the function.
// It handles JSON serialization/deserialization and HTTP communication automatically.
type Function[T, R any] struct {
	name             string
	fn               func(*T) (*R, error)
	httpClient       *http.Client
	clientConnection *Connection
}

// NewFunction creates a new typed function client for calling a specific function on a plugin.
// The function will serialize input of type T to JSON, send it to the plugin,
// and deserialize the response into type R.
func NewFunction[T, R any](name string, clientConnection *Connection) *Function[T, R] {
	function := &Function[T, R]{
		name:             name,
		clientConnection: clientConnection,
		httpClient:       &http.Client{Timeout: clientConnection.FunctionExecutionTimeout},
	}

	fn := func(input *T) (*R, error) {
		b, err := json.Marshal(input)
		if err != nil {
			return nil, &FunctionExecutionError{Function: name, Err: err}
		}

		url := fmt.Sprintf("%s/%s", clientConnection.BaseURL, name)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		if err != nil {
			return nil, &FunctionExecutionError{Function: name, Err: err}
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := function.httpClient.Do(req)
		if err != nil {
			return nil, &FunctionExecutionError{Function: name, Err: err}
		}

		if resp.Body != nil {
			defer func() {
				_ = resp.Body.Close()
			}()
		}

		out, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, &FunctionExecutionError{Function: name, Err: err}
		}

		if resp.StatusCode != http.StatusOK {
			return nil, &FunctionExecutionError{Function: name, Err: fmt.Errorf("plugin returned status %d: %s", resp.StatusCode, string(out))}
		}

		var output R
		err = json.Unmarshal(out, &output)
		if err != nil {
			return nil, &FunctionExecutionError{Function: name, Err: err}
		}

		return &output, nil
	}

	function.fn = fn
	return function
}

// SetTimeout configures the HTTP timeout for this specific function.
// This overrides the default timeout set in the connection.
func (f *Function[T, R]) SetTimeout(timeout time.Duration) {
	f.httpClient.Timeout = timeout
}

// Call executes the function with the provided input and returns the result.
// The input is serialized to JSON, sent to the plugin via HTTP POST,
// and the response is deserialized back to the expected output type.
func (f *Function[T, R]) Call(input *T) (*R, error) {
	out, err := f.fn(input)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Name returns the name of this function as registered with the plugin.
func (f *Function[T, R]) Name() string {
	return f.name
}

// Schema retrieves the JSON schema definition for this function's input and output types.
// This provides introspection capabilities to understand the expected data structure.
func (f *Function[T, R]) Schema() (*Schema, error) {
	url := fmt.Sprintf("%s/%s%s", f.clientConnection.BaseURL, f.name, schemasPath)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, &FunctionExecutionError{Function: f.Name(), Err: err}
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, &FunctionExecutionError{Function: f.Name(), Err: err}
	}

	if resp.Body != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}

	if resp.StatusCode != http.StatusOK {
		out, _ := io.ReadAll(resp.Body)
		return nil, &FunctionExecutionError{Function: f.Name(), Err: fmt.Errorf("plugin returned status %d: %s", resp.StatusCode, string(out))}
	}
	var schema Schema
	err = json.NewDecoder(resp.Body).Decode(&schema)
	if err != nil {
		return nil, &FunctionExecutionError{Function: f.Name(), Err: err}
	}
	return &schema, nil
}
