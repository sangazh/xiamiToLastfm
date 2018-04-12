package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xiami2LastFM/lastfm"
	"xiami2LastFM/util"
	"xiami2LastFM/xiami"
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
	run()
}

func run() {
	delayStart()
	fmt.Println("start scrobbling...")

	nowPlayingChan := make(chan interface{})
	playedChan := make(chan interface{}, 10)

	tickerXM := time.NewTicker(2 * time.Minute)
	quit := make(chan struct{})
	stop(quit)

	defer func() {
		close(nowPlayingChan)
		close(playedChan)
	}()

	go func() {
		for {
			select {
			case <-tickerXM.C:
				xiami.GetTracks(nowPlayingChan, playedChan)
			case <-quit:
				tickerXM.Stop()
				return
			}
		}
	}()
	go func() {
		for {
			lastfm.StartScrobble(playedChan)
		}
	}()
	go func() {
		for {
			lastfm.UpdateNowPlaying(nowPlayingChan)
		}
	}()

	time.Sleep(time.Hour)
	close(quit)
	fmt.Println("Stopped.")
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
