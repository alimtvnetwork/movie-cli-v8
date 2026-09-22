// result.go — generic single-value Result container following 02-spec/03-error-manage.
package appfault

// Result wraps a single value with structured AppError diagnostics.
type Result[T any] struct {
	err       *AppError
	value     T
	isDefined bool
}

// Ok returns a successful Result wrapping a defined value.
func Ok[T any](value T) Result[T] {
	return Result[T]{
		value:     value,
		isDefined: true,
	}
}

// Fail returns a failed Result wrapping an AppError.
func Fail[T any](err *AppError) Result[T] {
	return Result[T]{
		err: err,
	}
}

// FailWrap returns a failed Result wrapping an existing error.
func FailWrap[T any](err error, msg string) Result[T] {
	return Result[T]{
		err: Wrap(msg, err),
	}
}

// FailNew returns a failed Result wrapping a new formatted message.
func FailNew[T any](format string, args ...any) Result[T] {
	return Result[T]{
		err: New(format, args...),
	}
}

func (r *Result[T]) IsSuccess() bool {
	if r == nil {
		return false
	}

	return r.err == nil
}

func (r *Result[T]) IsFailure() bool {
	if r == nil {
		return true
	}

	return r.err != nil
}

func (r *Result[T]) IsDefined() bool {
	if r == nil {
		return false
	}

	if r.err != nil {
		return false
	}

	return r.isDefined
}

func (r *Result[T]) HasRecord() bool {
	return r.IsDefined()
}

func (r *Result[T]) HasRecords() bool {
	return r.IsDefined()
}

func (r *Result[T]) IsEmpty() bool {
	if r == nil {
		return true
	}

	return !r.isDefined
}

func (r *Result[T]) Count() int {
	if r.IsDefined() {
		return 1
	}

	return 0
}

func (r *Result[T]) IsCountOtherThan(n int) bool {
	if r.IsFailure() {
		return true
	}

	return r.Count() != n
}

func (r *Result[T]) Value() T {
	if r == nil {
		var zero T
		return zero
	}

	return r.value
}

func (r *Result[T]) Data() T {
	return r.Value()
}

func (r *Result[T]) ValueOr(fallback T) T {
	if r.IsDefined() {
		return r.value
	}

	return fallback
}

func (r *Result[T]) AppError() *AppError {
	if r == nil {
		return nil
	}

	return r.err
}

func (r *Result[T]) Fault() *AppError {
	return r.AppError()
}

func (r *Result[T]) Unwrap() (T, *AppError) {
	if r == nil {
		var zero T
		return zero, nil
	}

	return r.value, r.err
}
