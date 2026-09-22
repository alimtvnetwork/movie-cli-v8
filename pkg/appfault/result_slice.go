// result_slice.go — generic slice Result container following 02-spec/03-error-manage.
package appfault

// ResultSlice wraps a slice of items with structured AppError diagnostics.
type ResultSlice[T any] struct {
	err   *AppError
	items []T
}

// OkSlice returns a successful ResultSlice wrapping a slice of items.
func OkSlice[T any](items []T) ResultSlice[T] {
	return ResultSlice[T]{
		items: items,
	}
}

// FailSlice returns a failed ResultSlice wrapping an AppError.
func FailSlice[T any](err *AppError) ResultSlice[T] {
	return ResultSlice[T]{
		err: err,
	}
}

// FailSliceWrap returns a failed ResultSlice wrapping an existing error.
func FailSliceWrap[T any](err error, msg string) ResultSlice[T] {
	return ResultSlice[T]{
		err: Wrap(msg, err),
	}
}

// FailSliceNew returns a failed ResultSlice wrapping a new formatted message.
func FailSliceNew[T any](format string, args ...any) ResultSlice[T] {
	return ResultSlice[T]{
		err: New(format, args...),
	}
}

func (rs *ResultSlice[T]) IsSuccess() bool {
	if rs == nil {
		return false
	}

	return rs.err == nil
}

func (rs *ResultSlice[T]) IsFailure() bool {
	if rs == nil {
		return true
	}

	return rs.err != nil
}

func (rs *ResultSlice[T]) IsDefined() bool {
	if rs == nil {
		return false
	}

	if rs.err != nil {
		return false
	}

	return len(rs.items) > 0
}

func (rs *ResultSlice[T]) HasRecord() bool {
	return rs.IsDefined()
}

func (rs *ResultSlice[T]) HasRecords() bool {
	return rs.IsDefined()
}

func (rs *ResultSlice[T]) HasItems() bool {
	return rs.IsDefined()
}

func (rs *ResultSlice[T]) IsEmpty() bool {
	if rs == nil {
		return true
	}

	return len(rs.items) == 0
}

func (rs *ResultSlice[T]) Count() int {
	if rs == nil {
		return 0
	}

	if rs.err != nil {
		return 0
	}

	return len(rs.items)
}

func (rs *ResultSlice[T]) IsCountOtherThan(n int) bool {
	if rs.IsFailure() {
		return true
	}

	return rs.Count() != n
}

func (rs *ResultSlice[T]) Items() []T {
	if rs == nil {
		return nil
	}

	return rs.items
}

func (rs *ResultSlice[T]) Data() []T {
	return rs.Items()
}

func (rs *ResultSlice[T]) First() Result[T] {
	if !rs.IsDefined() {
		return Fail[T](rs.AppError())
	}

	return Ok[T](rs.items[0])
}

func (rs *ResultSlice[T]) Last() Result[T] {
	if !rs.IsDefined() {
		return Fail[T](rs.AppError())
	}

	return Ok[T](rs.items[len(rs.items)-1])
}

func (rs *ResultSlice[T]) AppError() *AppError {
	if rs == nil {
		return nil
	}

	return rs.err
}

func (rs *ResultSlice[T]) Fault() *AppError {
	return rs.AppError()
}

func (rs *ResultSlice[T]) Unwrap() ([]T, *AppError) {
	if rs == nil {
		return nil, nil
	}

	return rs.items, rs.err
}
