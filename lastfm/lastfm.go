package lastfm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"xiami2LastFM/xiami"
)

var (
	domain, apiKey, sharedSecret, apiUrl string
)

type ScrobbleParams struct {
	Artist            []string `json:"artist,omitempty"`
	Track             []string `json:"track,omitempty"`
	Timestamp         []string `json:"timestamp,omitempty"`
	Album             []string `json:"album,omitempty"`
	TrackNumber       []string `json:"trackNumber,omitempty"`
	Mbid              []string `json:"mbid,omitempty"` //The MusicBrainz Track ID
	AlbumArtist       []string `json:"albumArtist,omitempty"`
	DurationInSeconds []string `jsonapikey:"duration,omitempty"`
	ApiKey            string   `json:"api_key"`
	ApiSig            string   `json:"api_sig"`
	Sk                string   `json:"sk"`
	Format            string   `json:"format"`
	Method            string   `json:"method"`
}

func StartScrobble(playedChan chan interface{}) bool {
	t := <-playedChan
	if t == nil {
		return false
	}

	xm := t.(xiami.Track)
	log.Println("playedChan track: ", xm)

	query := make(map[string]interface{}, 0)

	query["artist[0]"] = xm.Artist
	query["album[0]"] = xm.Album
	query["track[0]"] = xm.Title
	query["timestamp[0]"] = fmt.Sprint(xm.Timestamp)

	query["method"] = "track.scrobble"
	query["sk"] = sk
	query["api_key"] = apiKey
	query["api_sig"] = signature(query)
	query["format"] = "json"
	log.Println("StartScrobble - request query:", query)

	respData, ok := postRequest(query)
	if !ok {
		fmt.Println("scrobble sent failed.")
		return false
	}

	accepted, ignored := renderScrobbleResp(respData)
	fmt.Printf("Scrobbled succeseful - accepted: %d, ignored: %d\n", accepted, ignored)
	return true
}

type ScrobbleResponse struct {
	Data ScrobbleData `json:"scrobbles"`
}

type ScrobbleData struct {
	Msg ScrobbleMsg `json:"@attr"`
}

type ScrobbleMsg struct {
	Accepted int `json:"accepted"`
	Ignored  int `json:"ignored"`
}

func renderScrobbleResp(data []byte) (accepted, ignored int) {
	var resp ScrobbleResponse
	json.Unmarshal(data, &resp)
	return resp.Data.Msg.Accepted, resp.Data.Msg.Ignored
}

func UpdateNowPlaying(nowPlayingChan chan interface{}) bool {
	t := <-nowPlayingChan
	if t == nil {
		return false
	}

	xm := t.(xiami.Track)
	log.Println("nowPlayingChan track: ", xm)

	query := map[string]interface{}{
		"method":  "track.updateNowPlaying",
		"sk":      sk,
		"api_key": apiKey,
		"artist":  xm.Artist,
		"album":   xm.Album,
		"track":   xm.Title,
	}

	query["api_sig"] = signature(query)
	query["format"] = "json"
	log.Println("UpdateNowPlaying - request query:", query)

	_, ok := postRequest(query)
	if !ok {
		fmt.Println("UpdateNowPlaying sent failed.")
		return false
	}

	fmt.Println("nowPlaying", xm)
	return true
}

// query input as map format, expect output with format of key=value&key=value
func queryString(query map[string]interface{}) (text string) {
	for key, value := range query {
		text += fmt.Sprintf("%s=%s&", key, value)
	}

	return strings.TrimRight(text, "&")
}

func getRequest(url string) ([]byte, bool) {
	log.Println("Getting request.. url: ", url)
	res, err := http.Get(url)

	if err != nil {
		log.Println(err)
		return nil, false
	}
	defer res.Body.Close()

	resData, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s on %s ", res.StatusCode, res.Status, url)
		log.Println("err body: ", string(resData))
		return nil, false
	}
	log.Println("Get response body: ", string(resData))

	return resData, true
}

func postRequest(query map[string]interface{}) ([]byte, bool) {
	r := bytes.NewReader([]byte(queryString(query)))
	contentType := "application/x-www-form-urlencoded"

	log.Println("Posting request.. params: ", query)
	res, err := http.Post(apiUrl, contentType, r)

	if err != nil {
		log.Println(err)
		return nil, false
	}

	resData, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s on %s ", res.StatusCode, res.Status, apiUrl)
		log.Println("err body: ", string(resData))
		return nil, false
	}
	log.Println("Post response body: ", string(resData))

	return resData, true
}

func toMap(byteData []byte) (result map[string]string) {
	r := bytes.NewReader(byteData)
	json.NewDecoder(r).Decode(&result)
	return result
}
