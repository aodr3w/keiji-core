package common

import "fmt"

type PathNotFound struct {
	Message string
}

func (e *PathNotFound) Error() string {
	return e.Message
}

func Is(target error) bool {
	_, ok := target.(*PathNotFound)
	return ok
}

func NewPathNotFound(path string) *PathNotFound {
	return &PathNotFound{
		fmt.Sprintf("path %v not found", path),
	}
}

var ErrPathNotFound = func(msg string) *PathNotFound {
	return NewPathNotFound(msg)
}
