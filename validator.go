package pluggo

import (
	"encoding/json"
	"fmt"

	"github.com/kaptinlin/jsonschema"
)

// Validator provides JSON schema validation for input data.
// It uses a compiled JSON schema to validate incoming data against
// the expected structure before deserialization.
type Validator[T any] struct {
	schema *jsonschema.Schema
}

// NewValidator creates a new validator for type T by generating a JSON schema
// from the provided struct. The validator can then be used to validate
// JSON input before deserialization.
func NewValidator[T any](v *T) (*Validator[T], error) {
	schema, err := structAsJSONSchema(v)
	if err != nil {
		return nil, fmt.Errorf("error generating input schema: %w", err)
	}

	compiler := jsonschema.NewCompiler()

	// Marshal schema to JSON
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	schemaValidator, err := compiler.Compile(schemaBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &Validator[T]{schema: schemaValidator}, nil
}

// Validate checks the provided data against the compiled JSON schema.
// It returns an evaluation result that contains validation status and
// any errors found during validation.
func (v *Validator[T]) Validate(data any) *jsonschema.EvaluationResult {
	return v.schema.Validate(data)
}
