// envelope.go — Universal Response Envelope conforming to 02-spec/03-error-manage.
package appfault

import "time"

// ResponseStatus holds the HTTP and execution outcome metadata.
type ResponseStatus struct {
	Timestamp string `json:"Timestamp"`
	Message   string `json:"Message"`
	Code      int    `json:"Code"`
	IsSuccess bool   `json:"IsSuccess"`
	IsFailed  bool   `json:"IsFailed"`
}

// ResponseAttributes holds descriptors about the payload shape and errors.
type ResponseAttributes struct {
	RequestedAt  string `json:"RequestedAt,omitempty"`
	TotalRecords int    `json:"TotalRecords"`
	HasAnyErrors bool   `json:"HasAnyErrors"`
	IsSingle     bool   `json:"IsSingle"`
	IsMultiple   bool   `json:"IsMultiple"`
	IsEmpty      bool   `json:"IsEmpty"`
}

// ErrorItem holds structured failure detail.
type ErrorItem struct {
	Code     string `json:"Code,omitempty"`
	Message  string `json:"Message"`
	Severity string `json:"Severity,omitempty"`
	Details  string `json:"Details,omitempty"`
}

// ErrorBlock groups error details by tier.
type ErrorBlock struct {
	Backend []ErrorItem `json:"Backend,omitempty"`
}

// ResponseEnvelope represents the standardized top-level API envelope.
type ResponseEnvelope[T any] struct {
	Errors     *ErrorBlock        `json:"Errors,omitempty"`
	Results    []T                `json:"Results"`
	Status     ResponseStatus     `json:"Status"`
	Attributes ResponseAttributes `json:"Attributes"`
}

// NewSuccessEnvelope creates a successful response envelope for a slice of results.
func NewSuccessEnvelope[T any](results []T, reqURL string) ResponseEnvelope[T] {
	count := len(results)

	return ResponseEnvelope[T]{
		Status: ResponseStatus{
			IsSuccess: true,
			IsFailed:  false,
			Code:      200,
			Message:   "OK",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
		Attributes: ResponseAttributes{
			RequestedAt:  reqURL,
			HasAnyErrors: false,
			IsSingle:     count == 1,
			IsMultiple:   count > 1,
			IsEmpty:      count == 0,
			TotalRecords: count,
		},
		Results: results,
	}
}

// NewSingleSuccessEnvelope creates a successful response envelope for a single item.
func NewSingleSuccessEnvelope[T any](item T, reqURL string) ResponseEnvelope[T] {
	return NewSuccessEnvelope([]T{item}, reqURL)
}

// NewErrorEnvelope creates a failed response envelope for error scenarios.
func NewErrorEnvelope(statusCode int, code, msg, reqURL string) ResponseEnvelope[any] {
	return ResponseEnvelope[any]{
		Status: ResponseStatus{
			IsSuccess: false,
			IsFailed:  true,
			Code:      statusCode,
			Message:   msg,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
		Attributes: ResponseAttributes{
			RequestedAt:  reqURL,
			HasAnyErrors: true,
			IsSingle:     false,
			IsMultiple:   false,
			IsEmpty:      true,
			TotalRecords: 0,
		},
		Results: []any{},
		Errors: &ErrorBlock{
			Backend: []ErrorItem{
				{
					Code:     code,
					Message:  msg,
					Severity: "Error",
				},
			},
		},
	}
}

// NewFaultEnvelope creates a failed response envelope from an AppError.
func NewFaultEnvelope(statusCode int, fault *AppError, reqURL string) ResponseEnvelope[any] {
	if fault == nil {
		return NewErrorEnvelope(statusCode, "UNKNOWN_ERROR", "An unknown error occurred", reqURL)
	}

	code := fault.Code
	if code == "" {
		code = "APP_FAULT"
	}

	msg := fault.Message
	if fault.DisplayError != "" {
		msg = fault.DisplayError
	}

	env := NewErrorEnvelope(statusCode, code, msg, reqURL)
	env.Errors.Backend[0].Details = fault.Details

	return env
}
