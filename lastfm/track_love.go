package lastfm

import (
	"fmt"
	"log"
	"net/url"

	"xiamiToLastfm/xiami"
)

func TrackLove(xm *xiami.Track) error {
	log.Println("lastfm.track.love track: ", xm)

	v := url.Values{}
	v.Set("artist", xm.Artist)
	v.Set("track", xm.Title)
	v.Set("method", "track.love")
	v.Set("sk", sk)
	v.Set("api_key", apiKey)
	sig := signature(&v)
	v.Set("api_sig", sig)
	v.Set("format", "json")

	if _, err := postRequest(v.Encode()); err != nil {
		//if failed, insert back to channel
		return err
	}

	fmt.Printf("loved:\t %s - %s \n", xm.Title, xm.Artist)
	return nil
}
