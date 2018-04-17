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
var quick bool

func init() {
	flag.BoolVar(&debug, "d", false, "debug mode, will export logs to file")
	flag.BoolVar(&quick, "q", false, "quick start mode, start immediately andd only run one time")
	flag.Parse()
	util.InitConfig()
}

func main() {
	f := util.Logger(debug)

	if f != nil {
		defer f.Close()
	}

	prepare()

	if quick {
		quickStart()
	} else {
		delayStart()
		run()
	}
}

func run() {
	fmt.Println("start scrobbling...")

	nowPlayingChan := make(chan xiami.Track)
	playedChan := make(chan xiami.Track, 10)

	tickerXM := time.NewTicker(time.Minute)
	quitChan := make(chan struct{})
	stop(quitChan)

	defer func() {
		close(nowPlayingChan)
		close(playedChan)
	}()

	go func() {
		for {
			select {
			case <-tickerXM.C:
				xiami.GetTracks(nowPlayingChan, playedChan)
			case <-quitChan:
				tickerXM.Stop()
				return
			}
		}
	}()
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
		time.Sleep(time.Hour)
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
		os.Exit(1)
	}()
}

func quickStart() {
	nowPlayingChan := make(chan xiami.Track)
	playedChan := make(chan xiami.Track, 10)
	xiami.GetTracks(nowPlayingChan, playedChan)
	quitChan := make(chan struct{})
	stop(quitChan)

	go func() {
		for {
			lastfm.StartScrobble(playedChan, quitChan)
		}
	}()

	lastfm.UpdateNowPlaying(nowPlayingChan, quitChan)

	defer func() {
		close(nowPlayingChan)
		close(playedChan)
	}()

	time.Sleep(time.Minute * 5)
	close(quitChan)
	fmt.Println("Stopped.")
}
