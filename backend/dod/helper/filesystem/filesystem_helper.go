package filesystem

import (
	"errors"
	"os"
)

func CheckExistence(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}
