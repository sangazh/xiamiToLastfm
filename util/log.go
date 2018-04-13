package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

func Logger(debug bool) (f *os.File) {
	if !debug {
		log.SetOutput(ioutil.Discard)
		return nil
	}

	date := time.Now().Format("2006-01-02")
	logFile := fmt.Sprintf("log/%s.log", date)

	logPath := filepath.Join(".", "log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		os.Mkdir(logPath, os.ModePerm)
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		fmt.Println(err)
	}

	log.SetOutput(f)
	return f
}
