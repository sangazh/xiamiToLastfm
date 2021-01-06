package lastfm

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"xiamiToLastfm/musicbrainz"
	"xiamiToLastfm/xiami"
)

// https://www.last.fm/api/show/track.scrobble
type ScrobbleParams struct {
	Artist            []string `json:"artist,omitempty"`
	Track             []string `json:"track,omitempty"`
	Timestamp         []string `json:"timestamp,omitempty"`
	Album             []string `json:"album,omitempty"`
	TrackNumber       []string `json:"trackNumber,omitempty"`
	Mbid              []string `json:"mbid,omitempty"` //The MusicBrainz Track ID
	AlbumArtist       []string `json:"albumArtist,omitempty"`
	DurationInSeconds []string `json:"duration,omitempty"`
	ApiKey            string   `json:"api_key"`
	ApiSig            string   `json:"api_sig"`
	Sk                string   `json:"sk"`
	Format            string   `json:"format"`
	Method            string   `json:"method"`
}

// Send scrobble info to last.fm server
// https://www.last.fm/api/show/track.scrobble
func Scrobble(xm xiami.Track) error {
	log.Println("lastfm.Scrobble playedChan track: ", xm)

	v := url.Values{}
	v.Set("artist[0]", xm.Artist)
	v.Set("album[0]", xm.Album)
	v.Set("track[0]", xm.Title)

	if mbid, ok := musicbrainz.MbID(xm.Title, xm.Artist, xm.Album); ok {
		v.Set("mbid[0]", string(mbid))
	}

	v.Set("timestamp[0]", fmt.Sprint(xm.Timestamp))
	v.Set("method", "track.scrobble")
	v.Set("sk", sk)
	v.Set("api_key", apiKey)
	sig := signature(&v)
	v.Set("api_sig", sig)
	v.Set("format", "json")

	respData, err := postRequest(v.Encode())
	if err != nil {
		//if failed, insert back to channel
		return err
	}

	accepted, ignored := scrobbleResponse(respData)
	log.Printf("last.fm: Scrobble succese: accepted - %d, ignored - %d\n", accepted, ignored)
	fmt.Printf("scrobbled:\t %s - %s \n", xm.Title, xm.Artist)

	return nil
}

func scrobbleResponse(data []byte) (accepted, ignored int) {
	type response struct {
		Data struct {
			Msg struct {
				Accepted int `json:"accepted"`
				Ignored  int `json:"ignored"`
			} `json:"@attr"`
		} `json:"scrobbles"`
	}

	var resp response
	json.Unmarshal(data, &resp)
	return resp.Data.Msg.Accepted, resp.Data.Msg.Ignored
}

// Update nowplaying
// https://www.last.fm/api/show/track.updateNowPlaying
func UpdateNowPlaying(xm xiami.Track) error {
	log.Println("last.fm: nowPlayingChan track: ", xm)

	v := url.Values{}
	v.Set("method", "track.updateNowPlaying")
	v.Set("sk", sk)
	v.Set("api_key", apiKey)
	v.Set("artist", xm.Artist)
	v.Set("album", xm.Album)
	v.Set("track", xm.Title)
	sig := signature(&v)
	v.Set("api_sig", sig)
	v.Set("format", "json")

	if _, err := postRequest(v.Encode()); err != nil {
		//if failed, as discard.
		return err
	}

	fmt.Printf("updated:\t %s - %s \n", xm.Title, xm.Artist)
	return nil
}
