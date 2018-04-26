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
)

var debug bool

func init() {
	flag.BoolVar(&debug, "d", false, "debug mode, will export logs to file")
	flag.Parse()
	util.InitConfig()
}

func main() {
	f := util.Logger(debug)

	if f != nil {
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
		util.TempRead(playedChan)
	}()

	tickerXM := time.NewTicker(time.Minute)
	quitChan := make(chan struct{})
	stop(quitChan)

	go func() {
		for {
			lastfm.StartScrobble(playedChan, quitChan)
		}
	}()
	go func() {
		for {
			lastfm.UpdateNowPlaying(nowPlayingChan, quitChan)
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
		util.TempStore(playedChan)
	}
}
