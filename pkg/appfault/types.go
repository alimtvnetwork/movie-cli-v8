// types.go — core types for appfault structured error architecture.
package appfault

// StackFrame represents a single frame in the error stack trace.
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// StackTrace holds an array of caller stack frames.
type StackTrace []StackFrame

// AppError is the standard structured error type for movie-cli.
type AppError struct {
	Cause        error             `json:"-"`
	Values       map[string]string `json:"values,omitempty"`
	Code         string            `json:"code,omitempty"`
	Message      string            `json:"message"`
	DisplayError string            `json:"display_error,omitempty"`
	Details      string            `json:"details,omitempty"`
	Contact      string            `json:"contact,omitempty"`
	Errors       []error           `json:"errors,omitempty"`
	StackTrace   StackTrace        `json:"stack_trace,omitempty"`
}

// Fault is an alias for AppError for cross-spec compatibility.
type Fault = AppError
