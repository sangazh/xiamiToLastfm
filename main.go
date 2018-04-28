package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xiamiToLastfm/lastfm"
	"xiamiToLastfm/util"
	"xiamiToLastfm/xiami"
	"log"
)

var debug bool

func init() {
	flag.BoolVar(&debug, "d", false, "debug mode, will export logs to file")
	flag.Parse()
	util.InitConfig()
}

func main() {
	if f, _ := util.Logger(debug); f != nil {
		defer f.Close()
	}

	prepare()
	delayStart()
	run()
}

func run() {
	fmt.Println("start scrobbling...")
	nowPlayingChan := make(chan xiami.Track)
	playedChan := make(chan xiami.Track, 10)
	defer func() {
		close(nowPlayingChan)
		close(playedChan)
	}()

	go func() {
		if err := util.TempRead(playedChan); err != nil {
			log.Println(err)
		}
	}()

	tickerXM := time.NewTicker(time.Minute)
	quitChan := make(chan struct{})
	lastfm.QuitChan = quitChan
	stop(quitChan)

	go func() {
		for {
			if err := lastfm.StartScrobble(playedChan); err != nil {
				fmt.Println("last.fm: scrobble sent failed. Try later.")
				log.Println("last.fm: ", err)
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
		case <-tickerXM.C:
			xiami.GetTracks(nowPlayingChan, playedChan)
		case <-quitChan:
			tickerXM.Stop()
			windUp(playedChan)
			os.Exit(1)
		}
	}
}

func prepare() {
	xiami.Init()
	lastfm.Auth()
}

func delayStart() {
	now := time.Now()
	start := now.Truncate(time.Minute).Add(time.Minute)
	fmt.Println("Will start at", start.String())

	sleep := start.UnixNano() - now.UnixNano()
	time.Sleep(time.Duration(sleep))
}

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

func windUp(playedChan chan xiami.Track) {
	if len(playedChan) > 0 {
		if err := util.TempStore(playedChan); err != nil {
			log.Println(err)
		}
	}
}
