package logger

import (
	"log"
	"os"
	"path"
	"sync"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/filesystem/dir"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
)

var File programconfig.Logging

func Initialize(logFiles programconfig.Logging) (err error) {
	File = logFiles
	err = dir.Create(path.Dir(File.Access))
	if err != nil {
		return
	}
	err = dir.Create(path.Dir(File.Error))
	if err != nil {
		return
	}
	return dir.Create(path.Dir(File.Info))
}

func Fatal(err error) {
	if err != nil {
		Error(err)
		os.Exit(1)
	}
}

var ErrorMutex sync.Mutex

func Error(err error) {
	ErrorMutex.Lock()
	initialize(File.Error, err.Error())
	ErrorMutex.Unlock()
}

var InfoMutex sync.Mutex

func Info(text string) {
	InfoMutex.Lock()
	initialize(File.Info, text)
	InfoMutex.Unlock()
}

func initialize(filePath string, text string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	logger := log.New(file, "", log.LstdFlags)
	logger.Println(text)
}
