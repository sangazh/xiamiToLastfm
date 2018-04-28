package util

import (
	"xiamiToLastfm/xiami"
	"os"
	"encoding/json"
	"log"
	"fmt"
	"bufio"
	"errors"
)

var file = "temp.txt"

func TempStore(playedChan chan xiami.Track) error {
	fmt.Println("unsent xiami tracks exist, saving to a temp file.")
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return errors.New("temp file create failed")
	}
	defer f.Close()

	n := len(playedChan)
	for i := 0; i < n; i++ {
		t := <-playedChan
		a, _ := json.Marshal(t)
		f.Write(a)
		f.WriteString("\n")
	}
	log.Println("Temp file created: ", file)
	return nil
}

func TempRead(playedChan chan xiami.Track) error {
	f, err := os.OpenFile(file, os.O_RDONLY, 0666)
	if err != nil {
		return errors.New("temp file read failed")
	}

	fmt.Println("previous unsent xiami tracks detected, will send to server later.")

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var t xiami.Track
		json.Unmarshal(scanner.Bytes(), &t)
		playedChan <- t
	}
	f.Close()

	os.Remove(file)
	log.Println("Temp file removed: ", file)
	return nil

}
