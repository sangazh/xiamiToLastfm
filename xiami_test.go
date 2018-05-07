package main

import (
	"testing"
	"xiamiToLastfm/xiami"
	"xiamiToLastfm/app"
)

func TestGetTracks(t *testing.T) {
	app.InitConfig()
	xiami.Init()

	nowPlayingChan := make(chan xiami.Track)
	playedChan := make(chan xiami.Track, 5)
	defer func() {
		close(nowPlayingChan)
		close(playedChan)
	}()

	xiami.Tracks(nowPlayingChan, playedChan)
}
