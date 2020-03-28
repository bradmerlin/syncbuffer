package syncbuffer

// Error allows errors to be consts.
type Error string

func (e Error) Error() string {
	return string(e)
}
