// methods.go — methods and fluent setters for AppError.
package appfault

import "fmt"

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Cause
}

func (e *AppError) WithContact(contact string) *AppError {
	if e != nil {
		e.Contact = contact
	}

	return e
}

func (e *AppError) WithErrors(errs ...error) *AppError {
	if e != nil {
		e.Errors = append(e.Errors, errs...)
	}

	return e
}

func (e *AppError) WithMsg(msg string) *AppError {
	if e != nil {
		e.Message = msg
	}

	return e
}

func (e *AppError) WithValue(key, val string) *AppError {
	if e != nil {
		if e.Values == nil {
			e.Values = make(map[string]string)
		}

		e.Values[key] = val
	}

	return e
}

func (e *AppError) WithDetails(details string) *AppError {
	if e != nil {
		e.Details = details
	}

	return e
}

func (e *AppError) WithDisplay(display string) *AppError {
	if e != nil {
		e.DisplayError = display
	}

	return e
}
