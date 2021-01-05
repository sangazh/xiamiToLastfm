package main

import (
	"testing"

	"xiamiToLastfm/app"
	"xiamiToLastfm/xiami"
)

func TestGetTracks(t *testing.T) {
	app.InitConfig("config.toml")
	xiami.Init()

	nowPlayingChan := make(chan xiami.Track)
	playedChan := make(chan xiami.Track, 5)
	defer func() {
		close(nowPlayingChan)
		close(playedChan)
	}()

	xiami.Tracks(nowPlayingChan, playedChan)
}
