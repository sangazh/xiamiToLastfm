package main

import (
	"flag"
	"log"

	"xiamiToLastfm/app"
	"xiamiToLastfm/lastfm"
	"xiamiToLastfm/xiami"
)

var debug bool
var config string

func main() {
	if f, _ := app.Logger(debug); f != nil {
		defer f.Close()
	}

	prepare()
	run()
}

func run() {
	data, err := xiami.ReadFile()
	if err != nil {
		panic(err)
	}

	for _, track := range data.Data {
		if err := lastfm.TrackLove(track); err != nil {
			log.Println("")
		}
	}

}

func init() {
	flag.BoolVar(&debug, "d", false, "debug mode, will export logs to file")
	flag.StringVar(&config, "c", "config.toml", "config name and path")
	flag.Parse()

	app.InitConfig(config)
}

func prepare() {
	lastfm.Auth()
}
