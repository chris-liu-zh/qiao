package tools

func NewPtr[T any](v T) *T {
	return &v
}
