package file

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/filesystem"
)

const fileExists string = "file already exists"

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
	if filesystem.CheckExistence(filePath) {
		return fmt.Errorf(fileExists)
	}
	myFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	myFile.Close()
	return
}

func Delete(filePath string) error {
	return os.Remove(filePath)
}

func Write(filePath string, data []byte) error {
	if filesystem.CheckExistence(filePath) {
		return fmt.Errorf(fileExists)
	}
	return ioutil.WriteFile(filePath, data, 0644)
}

func Read(filePath string) ([]byte, error) {
	if !filesystem.CheckExistence(filePath) {
		return nil, fmt.Errorf("file does not exist")
	}
	return ioutil.ReadFile(filePath)
}
