package encoding

import "errors"

var (
	ErrIterStop               = errors.New("iter stop")
	ErrDynamicIndexedArgument = errors.New("dynamic indexed argument")
)
