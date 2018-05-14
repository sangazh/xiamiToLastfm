package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"log"
	"flag"

	"xiamiToLastfm/app"
	"xiamiToLastfm/lastfm"
	"xiamiToLastfm/xiami"
)

var debug bool
var frequency = time.Minute
var config string

func main() {
	if f, _ := app.Logger(debug); f != nil {
		defer f.Close()
	}

	prepare()
	delayStart()
	run()
}

func run() {
	ticker := time.NewTicker(frequency)

	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "start scrobbling...")
	nowPlayingChan := make(chan xiami.Track)
	playedChan := make(chan xiami.Track, 10)
	defer func() {
		close(nowPlayingChan)
		windUp(playedChan)
		close(playedChan)
	}()

	quitChan := make(chan struct{})
	lastfm.QuitChan = quitChan
	stop(quitChan)

	go func() {
		if err := app.TempRead(playedChan); err != nil {
			log.Println(err)
		}
	}()

	go func() {
		for {
			if err := lastfm.Scrobble(playedChan); err != nil {
				fmt.Println("last.fm: scrobble sent failed. Try again in 5 seconds.")
				log.Println("last.fm: ", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()
	go func() {
		for {
			if err := lastfm.UpdateNowPlaying(nowPlayingChan); err != nil {
				fmt.Println("last.fm: updateNowPlaying sent failed.")
				log.Println("last.fm: ", err)
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			xiami.Tracks(nowPlayingChan, playedChan)
		case <-quitChan:
			ticker.Stop()
			return
		}
	}
}

func init() {
	flag.BoolVar(&debug, "d", false, "debug mode, will export logs to file")
	minute := flag.Uint64("m", 1, "how often to check the xiami page. unit: minute")
	flag.StringVar(&config, "c", "config.toml", "config name and path")
	flag.Parse()

	frequency *= time.Duration(*minute)

	app.InitConfig(config)
}

func prepare() {
	xiami.Init()
	lastfm.Auth()
}

// delayStart will delay the program a few seconds and start at exact next minute,
// to ensure time calculated from xiami page will be relatively accurate.
func delayStart() {
	now := time.Now()
	start := now.Truncate(time.Minute).Add(time.Minute)
	fmt.Println("Will start at", start.String())

	sleep := start.UnixNano() - now.UnixNano()
	time.Sleep(time.Duration(sleep))
}

// Detect Ctrl+C keyboard interruption and set quit signal to quitChan.
func stop(quit chan struct{}) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		close(quit)
		fmt.Println("Stopped.")
	}()
}

// Before the program quit, if scrobble play chan is not empty, save to a temp file.
func windUp(playedChan chan xiami.Track) {
	if len(playedChan) > 0 {
		if err := app.TempStore(playedChan); err != nil {
			log.Println(err)
		}
	}
}
