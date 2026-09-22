// appfault.go — constructors and wrapping functions for AppError.
package appfault

import "fmt"

// Wrap wraps an existing error with a contextual message.
// Returns an *AppError preserving the underlying cause and stack trace.
func Wrap(msg string, err error) *AppError {
	if err == nil {
		return nil
	}

	return &AppError{
		Message:    msg,
		Cause:      err,
		StackTrace: captureStackTrace(2),
	}
}

// Wrapf wraps an existing error with a formatted contextual message.
func Wrapf(err error, format string, args ...any) *AppError {
	if err == nil {
		return nil
	}

	return &AppError{
		Message:    fmt.Sprintf(format, args...),
		Cause:      err,
		StackTrace: captureStackTrace(2),
	}
}

// New creates a new AppError with a formatted message (no cause chain).
func New(format string, args ...any) *AppError {
	return &AppError{
		Message:    fmt.Sprintf(format, args...),
		StackTrace: captureStackTrace(2),
	}
}

// NewCode creates a new AppError with an explicit error code and formatted message.
func NewCode(code, format string, args ...any) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		StackTrace: captureStackTrace(2),
	}
}

// WrapCode wraps an existing error with an explicit error code and message.
func WrapCode(err error, code, format string, args ...any) *AppError {
	if err == nil {
		return nil
	}

	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		Cause:      err,
		StackTrace: captureStackTrace(2),
	}
}
