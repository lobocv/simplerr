package simplerr

func NewNotFoundErrorW(err error, fmt string, args ...interface{}) *SimpleError {
	return Wrap(err, fmt, args...).Code(CodeNotFound)
}

func IsNotFoundError(err error) bool {
	return HasErrorCode(err, CodeNotFound)
}
