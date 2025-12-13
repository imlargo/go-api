package files

import (
	"errors"
	"fmt"
)

var (
	ErrFileEmpty    = errors.New("file path cannot be empty")
	ErrFileNotExist = errors.New("file does not exist")
)

func ErrFileIsDirectory(path string) error {
	return fmt.Errorf("'%s' is a directory, a file was expected", path)
}

func ErrFileAccess(path string, err error) error {
	return fmt.Errorf("error accessing file '%s': %v", path, err)
}
