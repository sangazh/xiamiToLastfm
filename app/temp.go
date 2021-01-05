package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"xiamiToLastfm/xiami"
)

var file = "temp.txt"

// If program quit before messages in channel completely processed,
// the messages will be saved to a temp file, and processed next time when the program start.
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

// While program start, if a temp file detected and successful loaded,
// the messages will be added to the channel and processed soon.
func TempRead(playedChan chan xiami.Track) error {
	f, err := os.OpenFile(file, os.O_RDONLY, 0666)
	if err != nil {
		return errors.New("temp file not found")
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
