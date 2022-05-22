package file

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

func ReadDir(root string) ([]string, error) {
	var files []string
	f, err := os.Open(root)
	if err != nil {
		return files, err
	}
	fileInfo, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return files, err
	}
	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
}

func Creat(filePath string) (err error) {
	if CheckExistance(filePath) {
		return fmt.Errorf("file already exists")
	}
	myfile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	myfile.Close()
	return
}

func CheckExistance(filePath string) bool {
	_, error := os.Stat(filePath)
	//return !os.IsNotExist(err)
	return !errors.Is(error, os.ErrNotExist)
}

func Delete(filePath string) error {
	return os.Remove(filePath)
}

func Write(filePath string, data []byte) error {
	if CheckExistance(filePath) {
		return fmt.Errorf("file already exists")
	}
	return ioutil.WriteFile(filePath, data, 0644)
}
