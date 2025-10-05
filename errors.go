package pluggo

import (
	"fmt"
)

// PluginNotFoundError is returned when the specified plugin file cannot be found or accessed.
type PluginNotFoundError struct {
	Err error
}

// Error implements the error interface for PluginNotFoundError.
func (e *PluginNotFoundError) Error() string {
	return fmt.Sprintf("plugin not found: %v", e.Err)
}

// PluginExecutionError is returned when there's an error starting, running, or communicating with a plugin.
type PluginExecutionError struct {
	Err error
}

// Error implements the error interface for PluginExecutionError.
func (e *PluginExecutionError) Error() string {
	return fmt.Sprintf("plugin execution error: %v", e.Err)
}

// FunctionNotFoundError is returned when attempting to call a function that doesn't exist in the plugin.
type FunctionNotFoundError struct {
	Function string
}

// Error implements the error interface for FunctionNotFoundError.
func (e *FunctionNotFoundError) Error() string {
	return fmt.Sprintf("function %q not found in plugin", e.Function)
}

// FunctionListError is returned when there's an error retrieving the list of available functions from a plugin.
type FunctionListError struct {
	Err error
}

// Error implements the error interface for FunctionListError.
func (e *FunctionListError) Error() string {
	return fmt.Sprintf("error listing functions: %v", e.Err)
}

// FunctionLookupError is returned when there's an error looking up or accessing a specific function.
type FunctionLookupError struct {
	Function string
	Err      error
}

// Error implements the error interface for FunctionLookupError.
func (e *FunctionLookupError) Error() string {
	return fmt.Sprintf("error looking up function %q: %v", e.Function, e.Err)
}

// FunctionExecutionError is returned when there's an error executing a function within a plugin.
type FunctionExecutionError struct {
	Function string
	Err      error
}

// Error implements the error interface for FunctionExecutionError.
func (e *FunctionExecutionError) Error() string {
	return fmt.Sprintf("error executing function %q: %v", e.Function, e.Err)
}
