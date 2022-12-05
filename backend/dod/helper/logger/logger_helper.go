package logger

import (
	"log"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/filesystem/dir"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/programconfig"
)

var file programconfig.Logging

func Initialize(logFiles programconfig.Logging) (err error) {
	file = logFiles
	err = dir.Create(path.Dir(file.Access))
	if err != nil {
		return
	}
	err = dir.Create(path.Dir(file.Error))
	if err != nil {
		return
	}
	err = dir.Create(path.Dir(file.Info))
	if err != nil {
		return
	}
	return dir.Create(file.Task)
}

func Fatal(err error) {
	if err != nil {
		Error(err)
		os.Exit(1)
	}
}

var errorMutex sync.Mutex

func Error(err error) {
	errorMutex.Lock()
	initialize(file.Error, err.Error())
	errorMutex.Unlock()
}

var infoMutex sync.Mutex

func Info(text string) {
	infoMutex.Lock()
	initialize(file.Info, text)
	infoMutex.Unlock()
}

func Task(timeStamp int64, fileName string, text string) {
	initialize(file.Task+"/"+strconv.Itoa(int(timeStamp))+" "+fileName+".log", text)
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
