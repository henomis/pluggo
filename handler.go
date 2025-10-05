package pluggo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/invopop/jsonschema"
)

// FunctionHandler wraps a user-provided function with HTTP handling capabilities.
// It provides automatic JSON serialization/deserialization, input validation,
// and schema generation for plugin functions.
type FunctionHandler[T, R any] struct {
	handler   *Handler
	validator *Validator[T]
}

// Handler contains the HTTP handler and schema information for a plugin function.
type Handler struct {
	HTTPHandler http.Handler
	Schema      Schema
}

// NewFunctionHandler creates a new function handler that wraps a user function
// with HTTP request/response handling, JSON processing, and optional input validation.
// The handler automatically generates JSON schemas for input and output types.
func NewFunctionHandler[T, R any](fn func(context.Context, *T) (*R, error), validator *Validator[T]) *FunctionHandler[T, R] {
	inputSchema, err := structAsJSONSchema(new(T))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating input schema: %v\n", err)
	}

	outputSchema, err := structAsJSONSchema(new(R))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating output schema: %v\n", err)
	}

	schema := Schema{
		Input:  inputSchema,
		Output: outputSchema,
	}

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			fmt.Fprintf(os.Stderr, "method not allowed: %s\n", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte("method not allowed"))
			return
		}

		req, err := decodeInput(r, validator)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading request body: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		resp, err := fn(r.Context(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error executing function: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		err = encodeOutput(w, http.StatusOK, resp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error encoding response: %v\n", err)
			return
		}
	}

	return &FunctionHandler[T, R]{
		handler: &Handler{
			HTTPHandler: http.HandlerFunc(httpHandler),
			Schema:      schema,
		},
		validator: validator,
	}
}

// Handler returns the underlying HTTP handler and schema information.
// This is used internally by the plugin framework to register the function.
func (m *FunctionHandler[T, R]) Handler() *Handler {
	return m.handler
}

// encodeOutput serializes the response value to JSON and writes it to the HTTP response.
// It sets the appropriate content type and status code.
func encodeOutput(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// decodeInput reads and validates the JSON input from an HTTP request.
// It performs validation if a validator is provided, then deserializes
// the JSON into the expected input type T.
func decodeInput[T any](r *http.Request, validator *Validator[T]) (*T, error) {
	defer func() {
		_ = r.Body.Close()
	}()

	// Read body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Validate
	if validator != nil {
		result := validator.Validate(data)
		if !result.IsValid() {
			errors := make([]string, 0, len(result.Errors))
			for field, err := range result.Errors {
				errors = append(errors, fmt.Sprintf("%s: %s", field, err))
			}
			return nil, fmt.Errorf("invalid input: %s", strings.Join(errors, ", "))
		}
	}

	// Unmarshal after validation
	var req T
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

// structAsJSONSchema generates a JSON schema from a Go struct type.
// This is used for automatic schema generation and introspection capabilities.
func structAsJSONSchema(v any) (map[string]any, error) {
	r := new(jsonschema.Reflector)
	r.DoNotReference = true
	schema := r.Reflect(v)

	b, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	var jsonSchema map[string]any
	err = json.Unmarshal(b, &jsonSchema)
	if err != nil {
		return nil, err
	}

	delete(jsonSchema, "$schema")

	return jsonSchema, nil
}
