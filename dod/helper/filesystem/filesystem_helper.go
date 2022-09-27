package filesystem

import (
	"errors"
	"os"
)

func CheckExistance(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}
